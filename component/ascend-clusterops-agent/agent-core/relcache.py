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

"""Agent central relationship cache: task->pod->{node, poduid, rank}."""

# pylint: disable=duplicate-code  # watch/GC lifecycle mirrors pathmap

from __future__ import annotations

import json
import logging
import time


from agent_core.constants import (
    CLUSTER_SYSTEM_NS,
    POD_TTL,
    RELCACHE_CM_NAME,
    SNAPSHOT_MAX_BYTES,
    SNAPSHOT_SHARDS,
    SNAPSHOT_SHARD_FILL_LIMIT,
)
from agent_core.k8s import (
    SequentialShards,
    controller_owner,
    load_task_crds_config,
    mark_dirty,
)

logger = logging.getLogger(__name__)


def _rank_of(pod) -> str:
    ann = getattr(pod.metadata, "annotations", None) or {}
    return ann.get("ascend/job-rank", "") or ann.get("rank", "")


class RelationshipCache(SequentialShards):
    """task->pod->{node, poduid, rank} relationship cache (metadata-level).

    Threading model: start() launches background threads (pod watch + TTL GC + CM sync).
    apply_pod/apply_pod_deleted can be called directly by tests; no real cluster needed.
    """

    def __init__(self):
        super().__init__(SNAPSHOT_SHARD_FILL_LIMIT, SNAPSHOT_SHARDS)
        self._cm_name = RELCACHE_CM_NAME
        self._cm_ns = CLUSTER_SYSTEM_NS
        self._cm_key = "snapshot"
        self._log_name = "relcache"
        crds = [e for e in (load_task_crds_config().get("task_crds") or []) if e.get("kind")]
        self._task_kinds = {e["kind"] for e in crds}
        logger.info(
            "task CRs supported by agent-core: %s",
            json.dumps([f"{e.get('api_version', '')}/{e['kind']}" for e in crds], ensure_ascii=False)
            or "(none configured)",
        )
        # (ns, owner_name) -> {pod_uid: entry}; aliased to the base _job_pods so the shared
        # job-instance logic (replacement purge / evict index cleanup) operates on this index.
        self._by_job: dict[tuple[str, str], dict[str, dict]] = {}
        self._job_pods = self._by_job

    # ---- Event application (pure logic, callable directly by tests) ---- #
    def apply_pod(self, pod) -> None:
        owner = controller_owner(getattr(pod.metadata, "owner_references", None) or [])
        if owner is None:
            return
        owner_uid, owner_name, owner_kind = owner
        if self._task_kinds and owner_kind not in self._task_kinds:
            return  # ignore pods not managed by a configured task CR
        ns = pod.metadata.namespace or "default"
        uid = pod.metadata.uid
        pod_name = pod.metadata.name
        node = getattr(pod.spec, "node_name", "") or ""
        host_ip = getattr(pod.status, "host_ip", "") or ""
        rank = _rank_of(pod)
        phase = getattr(pod.status, "phase", "") or ""
        entry = {
            "pod_name": pod_name,
            "pod_uid": uid,
            "namespace": ns,
            "node": node,
            "host_ip": host_ip,
            "rank": rank,
            "owner_name": owner_name,
            "owner_uid": owner_uid,
            "deleted_at": None,
            "phase": phase,
        }
        with self._lock:
            self._check_job_replacement(ns, owner_name, owner_uid)
            old = self._pods.get(uid)
            if old and old.get("deleted_at"):
                entry["deleted_at"] = None  # rescheduled back, clear the deleted marker
            # Unify add / reschedule / node change as "detected new task pod"
            if old is None or old.get("deleted_at") or old.get("node") != node:
                logger.info(
                    "relcache detected new task pod: job=%s ns=%s pod=%s node=%s rank=%s phase=%s",
                    owner_name,
                    ns,
                    pod_name,
                    node,
                    rank,
                    phase,
                )
            else:
                old_phase = old.get("phase")
                if phase in ("Succeeded", "Failed") and old_phase not in ("Succeeded", "Failed"):
                    logger.info(
                        "relcache task pod completed: job=%s ns=%s pod=%s phase=%s",
                        owner_name,
                        ns,
                        pod_name,
                        phase,
                    )
            self._store_pod(uid, entry, old, SNAPSHOT_MAX_BYTES)
            self._by_job.setdefault((ns, owner_name), {})[uid] = entry
            mark_dirty(self._cond, self._dirty)  # schedule an immediate CM sync (change-driven)

    # ---- Query ---- #
    def lookup(self, jobname: str, namespace: str = "default") -> list[dict]:
        """Look up all pods of a job by jobname -> [{pod_name, pod_uid, node, host_ip, rank, namespace, phase, deleted_at}].

        Includes pods deleted within TTL (their logs still remain on nodes); entries
        expired beyond POD_TTL are pruned on access (lazy GC) and synced to the CM.
        ``deleted_at`` is the deletion timestamp (None for a live pod), used by the
        dispatcher to keep only the latest instance of a rescheduled pod.
        """
        now = time.time()
        with self._lock:
            m = self._by_job.get((namespace, jobname), {})
            pruned = False
            out: list[dict] = []
            for uid, e in list(m.items()):
                if e.get("deleted_at") and now - e["deleted_at"] > POD_TTL:
                    del m[uid]
                    self._evict_pod(uid, e)
                    pruned = True
                else:
                    out.append(
                        {
                            "pod_name": e["pod_name"],
                            "pod_uid": e["pod_uid"],
                            "node": e["node"],
                            "host_ip": e.get("host_ip", ""),
                            "rank": e["rank"],
                            "namespace": e["namespace"],
                            "phase": e.get("phase", ""),
                            "deleted_at": e.get("deleted_at"),
                        }
                    )
            if not m:
                self._by_job.pop((namespace, jobname), None)
            if pruned:
                mark_dirty(self._cond, self._dirty)  # sync the CM without the pruned entries
            return out

    # ---- CM snapshot ---- #
    def _snapshot(self, shard: int) -> str:
        with self._lock:
            pods = [e for uid, e in self._pods.items() if e.get("shard") == shard]
            return json.dumps({"pods": pods}, ensure_ascii=False, default=str)

    def _apply_snapshot(self, raw: str | None) -> None:
        """Parse a snapshot JSON string into the in-memory cache (pure logic, callable by tests)."""
        if not raw:
            return
        try:
            data = json.loads(raw)
        except json.JSONDecodeError:
            return
        with self._lock:
            for e in data.get("pods", []):
                self._index_loaded_entry(e["pod_uid"], e)
        self._resume_fill_position()


# Process-level singleton (started by main.py lifespan; reused by tools.py)
_cache: RelationshipCache | None = None


def get_cache() -> RelationshipCache:
    global _cache
    if _cache is None:
        _cache = RelationshipCache()
    return _cache


def init_cache() -> RelationshipCache:
    """main.py lifespan: start and return the singleton."""
    c = get_cache()
    c.start()
    return c


def shutdown_cache() -> None:
    if _cache is not None:
        _cache.stop()


def lookup(jobname: str, namespace: str = "default") -> list[dict]:
    """Entry point called by tools.diagnose; returns [] when the cache is not initialized."""
    if _cache is None:
        return []
    return _cache.lookup(jobname, namespace)
