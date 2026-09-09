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
import os
import threading
import time
from collections.abc import Iterable

from kubernetes import client, watch

from agent_core.constants import POD_TTL
from agent_core.k8s import (
    K8s,
    controller_owner,
    load_task_crds_config,
    mark_dirty,
    run_change_driven_sync,
    watch_events,
)

logger = logging.getLogger(__name__)

PATHMAP_CM_NAME = os.environ.get("PATHMAP_CM_NAME", "clusterops-pathmap")
PATHMAP_CM_NS = os.environ.get("PATHMAP_CM_NS", "cluster-system")
# The diagnosis system's own components (agent-core / node-collector) are not diagnosis targets
SKIP_APP_LABELS = ("agent-core", "node-collector")
# Deleted pod entries are retained for POD_TTL (default 7d, see agent_core.constants).
GC_INTERVAL = int(os.environ.get("PATHMAP_GC_INTERVAL", "300"))


def _all_mount_pairs(pod) -> list[str]:
    """All hostPath mount pairs of a pod (host:container), for collector-side paths/mount_keywords matching."""
    out: list[str] = []
    host_by_name = {
        v.name: v.host_path.path.rstrip("/") for v in pod.spec.volumes or [] if v.host_path and v.host_path.path
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


def _profile_id(pairs: list[str]) -> str:
    """Content hash of the mount pairs -> profile_id (content-addressed; pods of the same task share one profile)."""
    raw = json.dumps(sorted(set(pairs)), sort_keys=True, ensure_ascii=False)
    return hashlib.sha256(raw.encode()).hexdigest()[:8]


class PathmapWriter:
    """Full pod -> profile (host:container mount pairs) dedup table + global CM sync.

    apply_pod/apply_pod_deleted/_gc can be called directly by tests; no real cluster needed.
    """

    def __init__(self):
        # Record only task-CR-managed pods; an empty config (read failure) disables the filter.
        crds = [e for e in (load_task_crds_config().get("task_crds") or []) if e.get("kind")]
        self._task_kinds = {e["kind"] for e in crds}
        self._profiles: dict[str, dict] = {}  # profile_id -> {"pairs": [host:container, ...]}
        self._pods: dict[str, dict] = {}  # pod_uid -> {profile, env, deleted_at}
        self._lock = threading.RLock()
        self._cond = threading.Condition(self._lock)
        self._dirty: list[int] = [0]  # pending-change counter box (see k8s.mark_dirty / run_change_driven_sync)
        self._stop = threading.Event()
        self._threads: list[threading.Thread] = []
        self._v1 = None  # injectable in tests

    # ---- Event application (pure logic, callable directly by tests) ---- #
    def apply_pod(self, pod) -> None:
        uid = pod.metadata.uid
        if not uid:
            return
        owner = controller_owner(getattr(pod.metadata, "owner_references", None) or [])
        if owner is None or (self._task_kinds and owner[2] not in self._task_kinds):
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
            old = self._pods.get(uid)
            self._profiles.setdefault(pid, {"pairs": pairs})
            # Store only the fields the collector needs (profile/env/deleted_at).
            self._pods[uid] = {
                "profile": pid,
                "env": _pod_env(pod),
                "deleted_at": None,
            }
            if old and old.get("deleted_at"):
                self._pods[uid]["deleted_at"] = None  # rescheduled back, clear the deleted marker
            # Unify add / reschedule as a newly detected task pod
            if old is None or old.get("deleted_at"):
                logger.info(
                    "pathmap detected new task pod: pod=%s ns=%s node=%s profile=%s pairs=%d",
                    pod_name,
                    ns,
                    node,
                    pid,
                    len(pairs),
                )
            # if the pod's mount pairs changed, release the old profile reference
            if old and old.get("profile") and old["profile"] != pid:
                self._drop_profile_if_unused(old["profile"])
            mark_dirty(self._cond, self._dirty)  # schedule an immediate CM sync (change-driven)

    def apply_pod_deleted(self, pod_uid: str, ts: float | None = None) -> None:
        ts = ts if ts is not None else time.time()
        with self._lock:
            e = self._pods.get(pod_uid)
            if e and not e.get("deleted_at"):
                e["deleted_at"] = ts
                logger.info(
                    "pathmap task pod deleted: pod_uid=%s profile=%s",
                    pod_uid,
                    e.get("profile"),
                )
                mark_dirty(self._cond, self._dirty)  # schedule an immediate CM sync (change-driven)

    def _drop_profile_if_unused(self, pid: str) -> None:
        if any(p.get("profile") == pid for p in self._pods.values()):
            return
        self._profiles.pop(pid, None)

    # ---- Query/snapshot (for sync and tests) ---- #
    def _snapshot(self) -> dict:
        with self._lock:
            return {"profiles": self._profiles, "pods": self._pods}

    # ---- TTL GC ---- #
    def _gc(self, now: float | None = None) -> None:
        now = now if now is not None else time.time()
        changed = False
        with self._lock:
            for uid in list(self._pods):
                e = self._pods[uid]
                if e.get("deleted_at") and now - e["deleted_at"] > POD_TTL:
                    self._pods.pop(uid, None)
                    self._drop_profile_if_unused(e["profile"])
                    changed = True
        if changed:
            mark_dirty(self._cond, self._dirty)  # schedule an immediate CM sync (change-driven)

    def _gc_loop(self) -> None:
        while not self._stop.wait(GC_INTERVAL):
            try:
                self._gc()
            except Exception:  # noqa: BLE001  # GC is best-effort
                logger.exception("pathmap GC error")

    # ---- Global CM snapshot ---- #
    def _load_snapshot(self) -> None:
        try:
            v1 = self._v1 or K8s.core()
            cm = v1.read_namespaced_config_map(PATHMAP_CM_NAME, PATHMAP_CM_NS)
        except Exception as e:  # noqa: BLE001
            logger.warning("failed to read pathmap snapshot CM (start with empty table): %s", e)
            return  # snapshot missing/no permission -> start with empty table, rebuilt by watch
        raw = (cm.data or {}).get("pathmap.json")
        if not raw:
            return
        try:
            data = json.loads(raw)
        except json.JSONDecodeError:
            return
        with self._lock:
            self._profiles = dict(data.get("profiles") or {})
            self._pods = dict(data.get("pods") or {})

    def _save_snapshot(self) -> bool:
        """Sync the in-memory snapshot to the global CM; returns True on success (best-effort, never raises)."""
        snap = json.dumps(self._snapshot(), ensure_ascii=False, sort_keys=True, default=str)
        try:
            v1 = self._v1 or K8s.core()
            try:
                v1.read_namespaced_config_map(PATHMAP_CM_NAME, PATHMAP_CM_NS)
                v1.patch_namespaced_config_map(PATHMAP_CM_NAME, PATHMAP_CM_NS, {"data": {"pathmap.json": snap}})
            except client.ApiException as e:
                if e.status != 404:
                    raise
                v1.create_namespaced_config_map(
                    PATHMAP_CM_NS,
                    client.V1ConfigMap(
                        api_version="v1",
                        kind="ConfigMap",
                        metadata=client.V1ObjectMeta(name=PATHMAP_CM_NAME),
                        data={"pathmap.json": snap},
                    ),
                )
            return True
        except Exception as e:  # noqa: BLE001  # snapshot failure must not block diagnosis
            logger.warning(
                "failed to save pathmap snapshot (diagnosis unaffected): %s", e
            )  # snapshot failure is not blocking; in-memory cache remains usable
            return False

    def _sync_loop(self) -> None:
        """Change-driven CM sync: syncs as soon as in-memory pod data changed (shared impl in k8s)."""
        run_change_driven_sync(self._cond, self._dirty, self._stop, self._save_snapshot)

    # ---- Full pod watch ---- #
    def _pod_watch_loop(self) -> None:
        """Full pod watch via the shared k8s.watch_events (automatic reconnection)."""

        def stream() -> Iterable[dict]:
            return watch.Watch().stream(
                (self._v1 or K8s.core()).list_pod_for_all_namespaces,
                timeout_seconds=300,
                _request_timeout=320,
            )

        def on_event(typ: str, obj) -> None:
            if typ == "DELETED":
                self.apply_pod_deleted(obj.metadata.uid)
            else:  # ADDED / MODIFIED
                self.apply_pod(obj)

        watch_events(stream, on_event, self._stop, "pathmap pod")

    # ---- Lifecycle ---- #
    def start(self) -> None:
        self._load_snapshot()
        for target, name in (
            (self._pod_watch_loop, "pathmap-watch"),
            (self._gc_loop, "pathmap-gc"),
            (self._sync_loop, "pathmap-sync"),
        ):
            t = threading.Thread(target=target, name=name, daemon=True)
            t.start()
            self._threads.append(t)
        logger.info("pathmap-writer started: threads=%s", [t.name for t in self._threads])

    def stop(self) -> None:
        self._stop.set()
        for t in self._threads:
            t.join(timeout=5)
        self._threads.clear()
        self._stop = threading.Event()
        logger.info("pathmap-writer stopped")


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
