#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

"""agent-core relcache (central relationship cache) unit tests.

Covers: task->pod relationship writes, change-driven CM sync markers, retention of all
historic nodes after reschedule (A,B,C -> B,C,D still keeps A,B,C,D), TTL retention +
GC, and snapshot round-trip.
"""

from __future__ import annotations

import time
from types import SimpleNamespace

from agent_core import relcache as rc
from conftest import make_pod_metadata


def _make_pod(uid, name="p", ns="default", node="n1", job="job1", phase="Running"):
    """Build the pod object needed by relcache (SimpleNamespace is enough, no real cluster needed)."""
    owner_refs = [SimpleNamespace(controller=True, uid=f"job-uid-{job}", name=job, kind="AscendJob")]
    metadata = make_pod_metadata(uid, name, ns, owner_refs)
    return SimpleNamespace(
        metadata=metadata,
        spec=SimpleNamespace(node_name=node),
        status=SimpleNamespace(phase=phase),
    )


# --------------------------------------------------------------------------- #
# change-driven CM sync marker (apply/delete/GC mark it pending immediately)
# --------------------------------------------------------------------------- #
def test_apply_marks_dirty_for_change_driven_sync():
    c = rc.RelationshipCache()
    assert c._dirty[0] == 0  # pylint: disable=protected-access
    c.apply_pod(_make_pod(uid="u1"))
    assert c._dirty[0] == 1  # pylint: disable=protected-access
    c.apply_pod_deleted("u1", ts=1.0)
    assert c._dirty[0] == 2  # pylint: disable=protected-access


def test_gc_marks_dirty_only_when_removed():
    c = rc.RelationshipCache()
    now = 1_000_000.0
    c.apply_pod(_make_pod(uid="u1"))
    c.apply_pod_deleted("u1", ts=now)
    dirty_before = c._dirty[0]  # pylint: disable=protected-access
    # GC within TTL: nothing removed -> no sync triggered
    c._gc(now=now + rc.POD_TTL - 10)  # pylint: disable=protected-access
    assert c._dirty[0] == dirty_before  # pylint: disable=protected-access
    # GC beyond TTL: entries removed -> sync triggered
    c._gc(now=now + rc.POD_TTL + 10)  # pylint: disable=protected-access
    assert c._dirty[0] == dirty_before + 1  # pylint: disable=protected-access


# --------------------------------------------------------------------------- #
# reschedule retention: ran on A,B,C, rescheduled to B,C,D -> A,B,C,D all kept (within TTL)
# --------------------------------------------------------------------------- #
def test_rescheduled_job_keeps_all_historic_nodes():
    c = rc.RelationshipCache()
    for uid, node in (("ua", "node-a"), ("ub", "node-b"), ("uc", "node-c")):
        c.apply_pod(_make_pod(uid=uid, name=uid, node=node))
    c.apply_pod_deleted("ua", ts=time.time())  # the pod on the original node A died
    c.apply_pod(_make_pod(uid="ud", name="ud", node="node-d"))  # reschedule adds a pod on node D
    pods = c.lookup("job1", "default")
    nodes = {p["pod_name"] for p in pods}
    assert nodes == {"ua", "ub", "uc", "ud"}  # A,B,C,D are all kept
    deleted = next(p for p in pods if p["pod_name"] == "ua")
    assert deleted["node"] == "node-a"


# --------------------------------------------------------------------------- #
# deleted_at / TTL retention + GC (dead pods are retained for 7 days by default)
# --------------------------------------------------------------------------- #
def test_deleted_pod_retained_within_ttl_then_gc():
    c = rc.RelationshipCache()
    c.apply_pod(_make_pod(uid="u5", name="p5", node="node-x"))
    now = time.time()
    c.apply_pod_deleted("u5", ts=now)
    # within TTL: retained (deleted_at marker, data is not lost)
    assert c.lookup("job1", "default")[0]["pod_name"] == "p5"
    # beyond TTL: GC removes it, the task is no longer queryable
    c._gc(now=now + rc.POD_TTL + 10)  # pylint: disable=protected-access
    assert c.lookup("job1", "default") == []


def test_apply_pod_deleted_unknown_uid_is_noop():
    c = rc.RelationshipCache()
    c.apply_pod_deleted("no-such-uid", ts=1.0)  # must not raise


def test_lookup_prunes_expired_entries_lazily():
    # lazy pruning: lookup filters expired entries and triggers a CM sync
    c = rc.RelationshipCache()
    now = time.time()
    c.apply_pod(_make_pod(uid="u1", name="p1", node="node-a"))
    c.apply_pod_deleted("u1", ts=now - rc.POD_TTL - 10)  # already expired
    dirty_before = c._dirty[0]  # pylint: disable=protected-access
    assert c.lookup("job1", "default") == []
    assert c._dirty[0] == dirty_before + 1  # pylint: disable=protected-access


def test_lookup_keeps_within_ttl_entries():
    # within TTL (including just-deleted): lookup returns normally
    c = rc.RelationshipCache()
    now = time.time()
    c.apply_pod(_make_pod(uid="u1", name="p1", node="node-a"))
    c.apply_pod_deleted("u1", ts=now - rc.POD_TTL + 10)
    dirty_before = c._dirty[0]  # pylint: disable=protected-access
    pods = c.lookup("job1", "default")
    assert [p["pod_name"] for p in pods] == ["p1"]
    assert pods[0]["deleted_at"] is not None  # lookup exposes deleted_at for dedup of the newest instance
    assert c._dirty[0] == dirty_before  # pylint: disable=protected-access  # no expired entries, no sync


# --------------------------------------------------------------------------- #
# snapshot round-trip: after save -> load the task relationships stay queryable (restart reload)
# --------------------------------------------------------------------------- #
def test_snapshot_roundtrip_after_restart():
    c = rc.RelationshipCache()
    c.apply_pod(_make_pod(uid="u1", name="p1", node="node-a"))
    c.apply_pod_deleted("u1", ts=time.time())
    snap = c._snapshot()  # pylint: disable=protected-access

    c2 = rc.RelationshipCache()
    c2._apply_snapshot(snap)  # pylint: disable=protected-access
    pods = c2.lookup("job1", "default")
    assert len(pods) == 1
    assert pods[0]["pod_name"] == "p1"
    assert pods[0]["node"] == "node-a"
