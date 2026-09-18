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

"""pathmap-writer (agent-core) and collector-side CM cache (node-collector) unit tests.

agent-core stores all hostPath mount pairs per task pod (deduplicated by profile
fingerprint); the collector matches these pairs against the entity paths/mount_keywords.
Covers: mount-pair parsing, profile fingerprint dedup, TTL retention + GC, CM snapshot
round-trip, and all_pairs queries.
"""

from __future__ import annotations

import json
from types import SimpleNamespace

from kubernetes import client

from agent_core import pathmap as pw
from agent_core.constants import POD_TTL
from conftest import make_pod_metadata
from node_collector import pathmap as pm


def _all_snap(w):
    """Merge all sharded snapshots of a PathmapWriter into one dict (mirrors restart reload)."""
    merged = {"profiles": {}, "pods": {}}
    for shard in range(pw.SNAPSHOT_SHARDS):
        snap = json.loads(w._snapshot(shard))  # pylint: disable=protected-access
        merged["profiles"].update(snap.get("profiles") or {})
        merged["pods"].update(snap.get("pods") or {})
    return merged


def _make_pod(
    uid,
    name="p",
    ns="default",
    node="n1",
    mounts=(),
    volumes=(),
    owner_kind="AscendJob",
    owner_uid="owner-uid",
    owner_name="job1",
):
    """Build the pod object needed by pathmap-writer (SimpleNamespace is enough, no real cluster needed)."""
    container = SimpleNamespace(
        name="c0",
        volume_mounts=[SimpleNamespace(name=vn, mount_path=mp) for vn, mp in mounts],
    )
    spec = SimpleNamespace(
        node_name=node,
        containers=[container],
        volumes=[SimpleNamespace(name=vn, host_path=SimpleNamespace(path=hp)) for vn, hp in volumes],
    )
    owner_refs = (
        [SimpleNamespace(controller=True, uid=owner_uid, name=owner_name, kind=owner_kind)] if owner_kind else None
    )
    metadata = make_pod_metadata(uid, name, ns, owner_refs)
    return SimpleNamespace(metadata=metadata, spec=spec)


# --------------------------------------------------------------------------- #
# full mount-pair parsing (all host:container pairs, not per entity)
# --------------------------------------------------------------------------- #
def test_all_mount_pairs_keeps_host_container():
    pod = _make_pod(
        uid="u1",
        mounts=[("logs", "/home/hwMindX"), ("data", "/mnt")],
        volumes=[("logs", "/var/log/ascend"), ("data", "/mnt/data")],
    )
    assert pw._all_mount_pairs(pod) == ["/var/log/ascend:/home/hwMindX", "/mnt/data:/mnt"]


def test_all_mount_pairs_dedup():
    # the same volume mounted by multiple containers at the same path -> dedup to one pair
    pod = SimpleNamespace(
        spec=SimpleNamespace(
            containers=[
                SimpleNamespace(name="c0", volume_mounts=[SimpleNamespace(name="m", mount_path="/a")]),
                SimpleNamespace(name="c1", volume_mounts=[SimpleNamespace(name="m", mount_path="/a")]),
            ],
            volumes=[SimpleNamespace(name="m", host_path=SimpleNamespace(path="/h"))],
        ),
        metadata=SimpleNamespace(uid="u1", name="p", namespace="ns", annotations={}),
    )
    assert pw._all_mount_pairs(pod) == ["/h:/a"]


def test_pod_without_host_path_is_ignored():
    pod = _make_pod(uid="u6")
    assert pw._all_mount_pairs(pod) == []


def test_apply_pod_without_host_path_no_profile():
    w = pw.PathmapWriter()
    w.apply_pod(_make_pod(uid="u0"))
    assert _all_snap(w) == {"profiles": {}, "pods": {}}


# --------------------------------------------------------------------------- #
# record only task pods (ownerRef filter): non-task pods are never diagnosis targets
# --------------------------------------------------------------------------- #
def test_apply_pod_skips_non_task_owners():
    w = pw.PathmapWriter()
    w._task_kinds = {"AscendJob"}  # pylint: disable=protected-access
    mounts = [("m", "/x")]
    volumes = [("m", "/var/log/asc")]
    w.apply_pod(_make_pod(uid="u1", owner_kind="AscendJob", mounts=mounts, volumes=volumes))
    w.apply_pod(_make_pod(uid="u2", owner_kind="ReplicaSet", mounts=mounts, volumes=volumes))
    w.apply_pod(_make_pod(uid="u3", owner_kind="DaemonSet", mounts=mounts, volumes=volumes))
    snap = _all_snap(w)
    assert set(snap["pods"]) == {"u1"}


def test_apply_pod_skips_pod_without_owner():
    # no controller ownerRef (bare pod) -> skipped
    w = pw.PathmapWriter()
    w.apply_pod(_make_pod(uid="u4", owner_kind=None, mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")]))
    assert _all_snap(w) == {"profiles": {}, "pods": {}}  # pylint: disable=protected-access


def test_snapshot_pod_entries_slim():
    # slim pod entries: the collector needs {profile, env, deleted_at, shard, pod_ip};
    # owner fields are kept for same-ns+name job replacement on restart
    w = pw.PathmapWriter()
    w.apply_pod(_make_pod(uid="u5", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")]))
    snap = _all_snap(w)
    assert set(snap["pods"]["u5"]) == {
        "profile",
        "env",
        "deleted_at",
        "shard",
        "pod_ip",
        "namespace",
        "owner_name",
        "owner_uid",
    }


# --------------------------------------------------------------------------- #
# profile fingerprint dedup (500 pods of the same task share one profile)
# --------------------------------------------------------------------------- #
def test_profile_dedup_same_paths_share_profile():
    w = pw.PathmapWriter()
    pod_a = _make_pod(uid="ua", name="a", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")])
    pod_b = _make_pod(uid="ub", name="b", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")])
    w.apply_pod(pod_a)
    w.apply_pod(pod_b)
    snap = _all_snap(w)
    assert len(snap["profiles"]) == 1
    pid = snap["pods"]["ua"]["profile"]
    assert snap["pods"]["ub"]["profile"] == pid


def test_profile_dedup_different_paths_get_distinct_profiles():
    w = pw.PathmapWriter()
    pod_a = _make_pod(uid="ua", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")])
    pod_b = _make_pod(uid="ub", mounts=[("m", "/y")], volumes=[("m", "/var/log/asc")])
    w.apply_pod(pod_a)
    w.apply_pod(pod_b)
    snap = _all_snap(w)
    assert len(snap["profiles"]) == 2
    assert snap["pods"]["ua"]["profile"] != snap["pods"]["ub"]["profile"]


def test_pod_mount_change_releases_old_profile():
    # the same pod rescheduled with different mounts: the old profile is released once the
    # new entry replaces it in _pods (the release check must run after the replacement)
    w = pw.PathmapWriter()
    w.apply_pod(_make_pod(uid="u1", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")]))
    old_pid = _all_snap(w)["pods"]["u1"]["profile"]
    w.apply_pod(_make_pod(uid="u1", mounts=[("m", "/y")], volumes=[("m", "/var/log/asc")]))
    snap = _all_snap(w)
    assert snap["pods"]["u1"]["profile"] != old_pid
    assert len(snap["profiles"]) == 1  # only the new profile remains


# --------------------------------------------------------------------------- #
# deleted_at / TTL retention + GC (deleted pods stay diagnosable within POD_TTL)
# --------------------------------------------------------------------------- #
def test_deleted_pod_retained_within_ttl_then_gc():
    w = pw.PathmapWriter()
    pod = _make_pod(uid="u5", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")])
    w.apply_pod(pod)
    now = 1_000_000.0
    w.apply_pod_deleted("u5", ts=now)
    # within TTL: retained (deleted_at marker, data is not lost)
    snap = _all_snap(w)
    assert snap["pods"]["u5"]["deleted_at"] == now
    # beyond TTL: GC removes the pod and its now-unreferenced profile
    w._gc(now=now + POD_TTL + 10)  # pylint: disable=protected-access
    snap = _all_snap(w)
    assert not snap["pods"]
    assert not snap["profiles"]


def test_apply_pod_deleted_unknown_uid_is_noop():
    w = pw.PathmapWriter()
    w.apply_pod_deleted("no-such-uid", ts=1.0)  # must not raise


def test_reapplied_pod_clears_deleted_marker():
    # pod rescheduled back (re-ADDED) -> deleted_at cleared, data stays usable
    w = pw.PathmapWriter()
    pod = _make_pod(uid="u6", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")])
    w.apply_pod(pod)
    w.apply_pod_deleted("u6", ts=1.0)
    w.apply_pod(pod)
    snap = _all_snap(w)
    assert snap["pods"]["u6"]["deleted_at"] is None


# --------------------------------------------------------------------------- #
# global CM snapshot round-trip + collector-side cache/queries
# --------------------------------------------------------------------------- #
def test_snapshot_roundtrip_collector_all_pairs():
    w = pw.PathmapWriter()
    pod = _make_pod(
        uid="u7",
        name="p7",
        mounts=[("logs", "/home/hwMindX"), ("data", "/var/log")],
        volumes=[("logs", "/var/log/ascend"), ("data", "/mnt")],
    )
    w.apply_pod(pod)
    snap = _all_snap(w)
    entry = snap["pods"]["u7"]
    pairs = snap["profiles"][entry["profile"]]["pairs"]

    # collector gets the same pairs via the TriggerCollect request; all_pairs returns every host:container pair
    c = pm.PathMap()
    c.update_mounts([SimpleNamespace(pod_uid="u7", pairs=pairs, env=entry["env"])])
    assert c.all_pairs("u7") == ["/var/log/ascend:/home/hwMindX", "/mnt:/var/log"]
    # data is stored per mount pair, not per entity (no entity-name keys)
    assert "process_log" not in c.all_pairs("u7")


def test_collector_all_pairs_unknown_uid_empty():
    c = pm.PathMap()
    assert c.all_pairs("ghost") == []


def test_collector_update_mounts_merges_by_pod_uid():
    # per-request injection: same pod re-updated overwrites, other pods untouched
    c = pm.PathMap()
    c.update_mounts([SimpleNamespace(pod_uid="u8", pairs=["/a:/b"], env={})])
    assert c.all_pairs("u8") == ["/a:/b"]
    c.update_mounts([SimpleNamespace(pod_uid="u8", pairs=["/x:/y"], env={})])
    assert c.all_pairs("u8") == ["/x:/y"]


# --------------------------------------------------------------------------- #
# env recording (agent-core writes into pathmap) + collector pod_env query
# --------------------------------------------------------------------------- #
def test_apply_pod_records_literal_env():
    pod = SimpleNamespace(
        spec=SimpleNamespace(
            node_name="n1",
            containers=[
                SimpleNamespace(
                    name="c0",
                    volume_mounts=[SimpleNamespace(name="m", mount_path="/var/log")],
                    env=[
                        SimpleNamespace(name="ASCEND_PROCESS_LOG_PATH", value="/var/log/mindx-dl/plog"),
                        SimpleNamespace(name="FROM_SECRET", value=None),  # valueFrom reference -> skipped
                    ],
                )
            ],
            volumes=[SimpleNamespace(name="m", host_path=SimpleNamespace(path="/host"))],
        ),
        metadata=SimpleNamespace(
            uid="u1",
            name="p",
            namespace="ns",
            annotations={},
            owner_references=[SimpleNamespace(controller=True, uid="owner-uid", name="job1", kind="AscendJob")],
        ),
    )
    w = pw.PathmapWriter()
    w.apply_pod(pod)
    snap = _all_snap(w)
    env = snap["pods"]["u1"]["env"]
    assert env["ASCEND_PROCESS_LOG_PATH"] == "/var/log/mindx-dl/plog"
    assert "FROM_SECRET" not in env


def test_apply_pod_marks_dirty_for_change_driven_sync():
    # change-driven sync: apply_pod updates memory and immediately marks it pending sync
    w = pw.PathmapWriter()
    pod = _make_pod(uid="u1", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")])
    assert w._dirty[0] == 0  # pylint: disable=protected-access
    w.apply_pod(pod)
    assert w._dirty[0] >= 1  # pylint: disable=protected-access
    w.apply_pod_deleted("u1", ts=1.0)
    assert w._dirty[0] >= 2  # pylint: disable=protected-access


def test_save_snapshot_skips_empty_shards():
    # startup must not create CMs for empty shards: only the shard holding pod data is created
    created = []

    class FakeV1:  # noqa: N801  # minimal k8s CoreV1Api stand-in
        def read_namespaced_config_map(self, name, ns):
            raise client.ApiException(status=404)

        def create_namespaced_config_map(self, ns, cm):
            created.append(cm.metadata.name)

    w = pw.PathmapWriter()
    w._v1 = FakeV1()  # pylint: disable=protected-access
    w.apply_pod(_make_pod(uid="u1", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")]))
    assert w._save_snapshot()  # pylint: disable=protected-access
    assert len(created) == 1  # only the filled shard (0), not all SNAPSHOT_SHARDS
    assert created[0] == "clusterops-pathmap"  # shard 0 keeps the legacy CM name


def test_new_pods_fill_shards_sequentially(monkeypatch):
    # sequential fill: new pods go to shard 0, advance to shard 1 once the fill limit is reached
    monkeypatch.setattr(pw, "SNAPSHOT_SHARD_FILL_LIMIT", 1)  # every entry overflows the current shard
    w = pw.PathmapWriter()
    mounts = [("m", "/x")]
    volumes = [("m", "/var/log/asc")]
    for uid in ("u0", "u1", "u2"):
        w.apply_pod(_make_pod(uid=uid, mounts=mounts, volumes=volumes))
    shards = {int(w._pods[u]["shard"]) for u in ("u0", "u1", "u2")}  # pylint: disable=protected-access
    assert shards == {0, 1, 2}  # 0 filled -> 1 -> 2


def test_apply_pod_evicts_oldest_beyond_cap(monkeypatch):
    # in-memory byte cap: memory stays consistent with the sharded snapshot capacity
    w = pw.PathmapWriter()
    mounts = [("m", "/x")]
    volumes = [("m", "/var/log/asc")]
    for uid in ("u1", "u2", "u3"):
        w.apply_pod(_make_pod(uid=uid, mounts=mounts, volumes=volumes))
    monkeypatch.setattr(pw, "SNAPSHOT_MAX_BYTES", w._bytes - 1)  # pylint: disable=protected-access
    w.apply_pod(_make_pod(uid="u4", name="d", mounts=mounts, volumes=volumes))
    assert len(w._pods) <= 3  # pylint: disable=protected-access
    snap = _all_snap(w)
    assert "u1" not in snap["pods"]  # u1 (oldest) was evicted


def test_collector_pod_env_query():
    c = pm.PathMap()
    c.update_mounts([SimpleNamespace(pod_uid="u1", pairs=[], env={"ASCEND_PROCESS_LOG_PATH": "/var/log/plog"})])
    assert c.pod_env("u1", "ASCEND_PROCESS_LOG_PATH") == "/var/log/plog"
    assert c.pod_env("u1", "NOPE") is None
    assert c.pod_env("ghost", "ASCEND_PROCESS_LOG_PATH") is None  # pod deleted / not recorded


# --------------------------------------------------------------------------- #
# pod_ip recording (host IP keys the shared-storage plog dirs; pod IP is the fallback)
# --------------------------------------------------------------------------- #
def test_apply_pod_records_host_ip():
    w = pw.PathmapWriter()
    pod = _make_pod(uid="u1", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")])
    pod.status = SimpleNamespace(host_ip="10.0.0.5", pod_ip="192.168.1.10")
    w.apply_pod(pod)
    assert w._pods["u1"]["pod_ip"] == "10.0.0.5"  # pylint: disable=protected-access


def test_apply_pod_falls_back_to_pod_ip():
    w = pw.PathmapWriter()
    pod = _make_pod(uid="u2", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")])
    pod.status = SimpleNamespace(pod_ip="192.168.1.10")
    w.apply_pod(pod)
    assert w._pods["u2"]["pod_ip"] == "192.168.1.10"  # pylint: disable=protected-access


def test_apply_pod_missing_pod_ip_empty():
    w = pw.PathmapWriter()
    w.apply_pod(_make_pod(uid="u3", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")]))
    assert w._pods["u3"]["pod_ip"] == ""  # pylint: disable=protected-access


def test_same_job_replaced_clears_old_pods():
    # same ns+name job recreated (new owner UID): old pods are dropped, profile index stays consistent
    w = pw.PathmapWriter()
    mounts = [("m", "/x")]
    volumes = [("m", "/var/log/asc")]
    w.apply_pod(_make_pod(uid="u1", owner_uid="old-uid", mounts=mounts, volumes=volumes))
    w.apply_pod(_make_pod(uid="u2", name="p2", owner_uid="new-uid", mounts=mounts, volumes=volumes))
    snap = _all_snap(w)
    assert set(snap["pods"]) == {"u2"}  # old instance's pod dropped
    assert len(snap["profiles"]) == 1  # shared profile still referenced by u2


def test_same_job_replaced_releases_unused_profile():
    # old pod had a distinct profile; after replacement the profile is released
    w = pw.PathmapWriter()
    w.apply_pod(_make_pod(uid="u1", owner_uid="old-uid", mounts=[("m", "/x")], volumes=[("m", "/var/log/asc")]))
    w.apply_pod(
        _make_pod(uid="u2", name="p2", owner_uid="new-uid", mounts=[("m", "/y")], volumes=[("m", "/var/log/asc")])
    )
    snap = _all_snap(w)
    assert set(snap["pods"]) == {"u2"}
    assert len(snap["profiles"]) == 1  # only u2's profile remains
