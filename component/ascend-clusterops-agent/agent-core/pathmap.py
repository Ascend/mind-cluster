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

"""Agent central pathmap-writer: full pod watch -> host:container mount pairs -> profile dedup -> global CM.

The node-collector watches this global CM and caches it locally in memory.
"""

# pylint: disable=duplicate-code  # watch/GC lifecycle mirrors relcache

from __future__ import annotations

import hashlib
import json
import logging


from agent_core.constants import (
    CLUSTER_SYSTEM_NS,
    PATHMAP_CM_NAME,
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


# The diagnosis system's own components (agent-core / node-collector) are not diagnosis targets
SKIP_APP_LABELS = ("agent-core", "node-collector")
# Deleted pod entries are retained for POD_TTL (default 7d, see agent_core.constants).
GC_INTERVAL = 300
# Entry key referencing the profile_id (pods of the same mount pairs share one profile).
KEY_PROFILE = "profile"
# Entry/snapshot keys (kept consistent with the relcache/k8s entry layout).
KEY_DELETED_AT = "deleted_at"
KEY_ENV = "env"
KEY_POD_IP = "pod_ip"
KEY_PAIRS = "pairs"
KEY_NAMESPACE = "namespace"
KEY_OWNER_NAME = "owner_name"
KEY_OWNER_UID = "owner_uid"
KEY_SHARD = "shard"
KEY_SNAPSHOT_PROFILES = "profiles"
KEY_SNAPSHOT_PODS = "pods"


def _all_mount_pairs(pod) -> list[str]:
    """All hostPath mount pairs of a pod (host:container), for collector-side paths/mount_keywords matching."""
    out: list[str] = []
    host_by_name = {
        v.name: v.host_path.path.rstrip("/")
        for v in pod.spec.volumes or []
        if getattr(v, "host_path", None) and getattr(v.host_path, "path", None)
    }
    for c in pod.spec.containers or []:
        for vm in c.volume_mounts or []:
            host = host_by_name.get(vm.name)
            if not host:
                continue
            pair = f"{host}:{vm.mount_path.rstrip('/')}"
            if pair not in out:
                out.append(pair)
    return out


def _pod_env(pod) -> dict[str, str]:
    """Literal container env values (name -> value); valueFrom references are skipped (not resolvable).

    Recorded into the pathmap CM so the collector can resolve env-based entities (e.g.
    ASCEND_PROCESS_LOG_PATH) even after the task pod is deleted (kept within TTL).
    """
    out: dict[str, str] = {}
    for c in pod.spec.containers or []:
        for env in getattr(c, "env", None) or []:
            if env.name and env.value is not None:
                out.setdefault(env.name, env.value)
    return out


def _pod_ip(pod) -> str:
    """Host IP the framework keys plog dirs by (e.g. XDL_IP = status.hostIP), used by the
    collector for shared-storage plog matching; falls back to the pod IP.
    """
    status = getattr(pod, "status", None)
    return getattr(status, "host_ip", None) or getattr(status, "pod_ip", None) or ""


def _profile_id(pairs: list[str]) -> str:
    """Content hash of the mount pairs -> profile_id (content-addressed; pods of the same task share one profile)."""
    raw = json.dumps(sorted(set(pairs)), sort_keys=True, ensure_ascii=False)
    return hashlib.sha256(raw.encode()).hexdigest()[:8]


class PathmapWriter(SequentialShards):
    """Full pod -> profile (host:container mount pairs) dedup table + global CM sync.

    apply_pod/apply_pod_deleted/_gc can be called directly by tests; no real cluster needed.
    """

    def __init__(self):
        super().__init__(SNAPSHOT_SHARD_FILL_LIMIT, SNAPSHOT_SHARDS)
        self._cm_name = PATHMAP_CM_NAME
        self._cm_ns = CLUSTER_SYSTEM_NS
        self._cm_key = "pathmap.json"
        self._log_name = "pathmap"
        self._gc_interval = GC_INTERVAL
        # Record only task-CR-managed pods; an empty config (read failure) disables the filter.
        crds = [e for e in (load_task_crds_config().get("task_crds") or []) if e.get("kind")]
        self._task_kinds = {e["kind"] for e in crds}
        self._profiles: dict[str, dict] = {}  # profile_id -> {"pairs": [host:container, ...]}

    # ---- Event application (pure logic, callable directly by tests) ---- #
    def apply_pod(self, pod) -> None:
        uid = pod.metadata.uid
        if not uid:
            return
        owner = controller_owner(getattr(pod.metadata, "owner_references", None) or [])
        if owner is None:
            return
        owner_uid, owner_name, owner_kind = owner
        if self._task_kinds and owner_kind not in self._task_kinds:
            return  # only task-CR-managed pods enter the pathmap
        labels = getattr(pod.metadata, "labels", None) or {}
        if labels.get("app") in SKIP_APP_LABELS:
            return  # diagnose system's own components (agent-core/node-collector), not collected or added to pathmap
        pairs = _all_mount_pairs(pod)
        if not pairs:
            return  # no hostPath mounts -> no profile
        pid = _profile_id(pairs)
        pod_name = pod.metadata.name
        ns = pod.metadata.namespace or "default"
        node = getattr(pod.spec, "node_name", "") or ""
        with self._lock:
            self._check_job_replacement(ns, owner_name, owner_uid)
            old = self._pods.get(uid)
            self._profiles.setdefault(pid, {KEY_PAIRS: pairs})
            # Store only the fields the collector needs (profile/env/deleted_at/pod_ip/owner).
            entry = {
                KEY_PROFILE: pid,
                KEY_ENV: _pod_env(pod),
                KEY_DELETED_AT: None,
                KEY_POD_IP: _pod_ip(pod),
                KEY_NAMESPACE: ns,
                KEY_OWNER_NAME: owner_name,
                KEY_OWNER_UID: owner_uid,
            }
            if old and old.get(KEY_DELETED_AT):
                entry[KEY_DELETED_AT] = None  # rescheduled back, clear the deleted marker
            # Unify add / reschedule as a newly detected task pod
            if old is None or old.get(KEY_DELETED_AT):
                logger.info(
                    "pathmap detected new task pod: pod=%s ns=%s node=%s profile=%s pairs=%d",
                    pod_name,
                    ns,
                    node,
                    pid,
                    len(pairs),
                )
            self._job_pods.setdefault((ns, owner_name), {})[uid] = entry
            self._store_pod(uid, entry, old, SNAPSHOT_MAX_BYTES)
            # if the pod's mount pairs changed, release the old profile reference;
            # runs after _store_pod so _pods no longer counts the old entry
            if old and old.get(KEY_PROFILE) and old[KEY_PROFILE] != pid:
                self._drop_profile_if_unused(old[KEY_PROFILE])
            mark_dirty(self._cond, self._dirty)  # schedule an immediate CM sync (change-driven)

    def _after_evict(self, uid: str, e: dict) -> None:
        """Drop the evicted pod's profile when no other pod references it (job index is handled by the base)."""
        self._drop_profile_if_unused(e.get(KEY_PROFILE, ""))

    def _drop_profile_if_unused(self, pid: str) -> None:
        if any(p.get(KEY_PROFILE) == pid for p in self._pods.values()):
            return
        self._profiles.pop(pid, None)

    # ---- Query/snapshot (for sync and tests) ---- #
    def pod_pairs_env(self, pod_uid: str) -> tuple[list[str], dict[str, str], str]:
        """Return the pod's mount pairs, literal env and pod IP (for TriggerCollect; empty for unknown pods)."""
        with self._lock:
            pod = self._pods.get(pod_uid)
            if not pod:
                return [], {}, ""
            profile = self._profiles.get(pod.get(KEY_PROFILE, ""), {})
            return (
                list(profile.get(KEY_PAIRS, []) or []),
                dict(pod.get(KEY_ENV) or {}),
                pod.get(KEY_POD_IP, ""),
            )

    def _snapshot(self, shard: int) -> str:
        with self._lock:
            pods = {uid: e for uid, e in self._pods.items() if e.get(KEY_SHARD) == shard}
            pids = {e.get(KEY_PROFILE, "") for e in pods.values()} & set(self._profiles)
            profiles = {pid: self._profiles[pid] for pid in pids}
            return json.dumps(
                {KEY_SNAPSHOT_PROFILES: profiles, KEY_SNAPSHOT_PODS: pods},
                ensure_ascii=False,
                sort_keys=True,
                default=str,
            )

    def _apply_snapshot(self, raw: str | None) -> None:
        if not raw:
            return
        try:
            data = json.loads(raw)
        except json.JSONDecodeError:
            return
        with self._lock:
            self._profiles.update(dict(data.get(KEY_SNAPSHOT_PROFILES) or {}))
            for uid, e in (data.get(KEY_SNAPSHOT_PODS) or {}).items():
                self._index_loaded_entry(uid, e)
        self._resume_fill_position()


# Process-level singleton (started by main.py lifespan)
_writer: PathmapWriter | None = None


def get_writer() -> PathmapWriter:
    global _writer
    if _writer is None:
        _writer = PathmapWriter()
    return _writer


def init_writer() -> PathmapWriter:
    """main.py lifespan: start and return the singleton."""
    w = get_writer()
    w.start()
    return w


def shutdown_writer() -> None:
    if _writer is not None:
        _writer.stop()
