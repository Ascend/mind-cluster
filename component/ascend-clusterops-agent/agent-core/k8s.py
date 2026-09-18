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

"""Shared Kubernetes client helpers for agent-core.

Provides in-cluster / kubeconfig loading (K8s), the agent-core namespace resolution,
and two reusable primitives: the watch-with-reconnect loop (watch_events) and the
change-driven CM sync (mark_dirty / run_change_driven_sync).
"""

from __future__ import annotations

import json
import logging
import os
import threading
import time
from collections import deque
from collections.abc import Iterable
from pathlib import Path
from typing import Any, Callable

import yaml
from kubernetes import client, config, watch

from agent_core.constants import CLUSTER_SYSTEM_NS, POD_TTL, TASK_CRDS_CM_KEY, TASK_CRDS_CM_NAME

logger = logging.getLogger(__name__)

_SA_NAMESPACE_FILE = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
DEFAULT_NS = "mindx-dl"


def load_task_crds_config() -> dict:
    """Read the task CR config CM via the K8s API; empty config + warning on failure (no ownerRef filtering)."""
    try:
        v1 = K8s.core()
        cm = v1.read_namespaced_config_map(TASK_CRDS_CM_NAME, CLUSTER_SYSTEM_NS)
        raw = (cm.data or {}).get(TASK_CRDS_CM_KEY, "")
        return yaml.safe_load(raw) or {}
    except Exception as e:  # noqa: BLE001
        logger.warning("failed to read task_crds config (no task CR configured, pod ownerRef not filtered): %s", e)
        return {}


def controller_owner(refs) -> tuple[str, str, str] | None:
    """Return the controller ownerReference -> (uid, name, kind); None if absent."""
    for r in refs or []:
        if getattr(r, "controller", None) and getattr(r, "uid", None):
            return r.uid, r.name, r.kind
    return None


def shard_cm_name(base_name: str, shard: int) -> str:
    """Per-shard ConfigMap name (relcache/pathmap snapshots); shard 0 keeps the legacy name."""
    return base_name if shard == 0 else f"{base_name}-{shard}"


def serialized_bytes(entry: dict) -> int:
    """Serialized size of one cache entry (proxy for its in-memory footprint)."""
    return len(json.dumps(entry, ensure_ascii=False, default=str, sort_keys=True))


def shard_has_pods(snap: str) -> bool:
    """Whether a shard snapshot carries pod data (empty shards must not create a CM)."""
    try:
        return bool(json.loads(snap).get("pods"))
    except json.JSONDecodeError:
        return True  # unparseable -> keep as non-empty so data is never silently dropped


class SequentialShards:
    """Shared pod-cache skeleton: sequential shard routing + byte accounting + job-instance
    replacement + sharded CM snapshot + watch/GC/sync threads (used by relcache/pathmap).

    ``_after_evict`` is the subclass hook for business indexes; the subclass provides the
    entry format via ``apply_pod`` and the snapshot format via ``_snapshot``/``_apply_snapshot``.
    The snapshot/threading parameters (``_cm_name``/``_cm_ns``/``_cm_key``/``_log_name``/
    ``_gc_interval``) are set by the subclass ``__init__``.
    """

    def __init__(self, fill_limit: int, shards: int):
        self._active = 0  # shard being filled sequentially (new pods land here)
        self._shard_bytes: dict[int, int] = {}  # per-shard accumulated entry bytes (fill-level probe)
        # Shared indexes.
        self._pods: dict[str, dict] = {}  # pod_uid -> entry
        self._order: deque[str] = deque()  # pod_uid insertion order for capacity eviction (FIFO)
        self._bytes = 0  # summed serialized size of _pods (compared against the subclass's max_bytes)
        # Job-instance indexes (memory-only, not persisted): same ns+name job CR with a new
        # owner UID replaces the cached pods of the old instance (see _purge_job).
        self._job_uid: dict[tuple[str, str], str] = {}  # (ns, owner_name) -> current owner UID
        self._job_pods: dict[tuple[str, str], dict[str, dict]] = {}  # (ns, owner_name) -> {pod_uid: entry}
        self._last: dict[int, str] = {}  # per-shard last-synced snapshot (only changed shards are written)
        self._lock = threading.RLock()
        self._cond = threading.Condition(self._lock)
        self._dirty: list[int] = [0]  # pending-change counter box (see mark_dirty / run_change_driven_sync)
        self._stop = threading.Event()
        self._threads: list[threading.Thread] = []
        self._v1 = None  # lazy k8s client (tests can avoid touching the cluster)
        self._fill_limit = fill_limit
        self._shards = shards
        # Snapshot / threading parameters (overridden by the subclass __init__).
        self._cm_name = ""
        self._cm_ns = ""
        self._cm_key = ""
        self._log_name = "cache"
        self._gc_interval = 600

    def _route_new_shard(self, size: int) -> int:
        """Advance past full shards, wrapping around to the oldest one when all are full (recycled).

        An empty shard always accepts its entry (a single oversized entry overruns the
        fill limit instead of looping forever); only occupied shards are advanced past.
        """
        while (
            self._shard_bytes.get(self._active, 0) and self._shard_bytes.get(self._active, 0) + size > self._fill_limit
        ):
            self._active = (self._active + 1) % self._shards
            if self._shard_bytes.get(self._active, 0):
                self._evict_shard(self._active)  # recycled old shard: drop all of its pods to make room
        return self._active

    def _evict_shard(self, shard: int) -> None:
        """Evict every pod of a recycled shard (oldest batch) so it can be reused."""
        for uid, e in list(self._pods.items()):
            if e.get("shard") == shard:
                self._evict_pod(uid, e)

    def _evict_index_only(self, uid: str, e: dict) -> None:
        """Subtract one entry's size from the total and its shard (used when the entry is replaced)."""
        size = serialized_bytes(e)
        self._bytes -= size
        shard = int(e.get("shard", 0))
        self._shard_bytes[shard] = max(0, self._shard_bytes.get(shard, 0) - size)

    def _evict_pod(self, uid: str, e: dict) -> None:
        """Remove one pod entry from all indexes (capacity eviction / TTL GC / shard recycle / job purge)."""
        self._pods.pop(uid, None)
        self._evict_index_only(uid, e)
        self._drop_job_index(uid, e)
        self._after_evict(uid, e)
        if uid in self._order:
            self._order.remove(uid)

    def _drop_job_index(self, uid: str, e: dict) -> None:
        """Remove the evicted pod from its job instance index (shared by relcache/pathmap)."""
        job = (e.get("namespace"), e.get("owner_name"))
        jp = self._job_pods.get(job)
        if jp:
            jp.pop(uid, None)
            if not jp:
                self._job_pods.pop(job, None)
                self._job_uid.pop(job, None)

    def _purge_job(self, ns: str, owner_name: str) -> None:
        """Drop all cached pods of a replaced job instance (same ns+name, new owner UID)."""
        pods = self._job_pods.pop((ns, owner_name), None)
        self._job_uid.pop((ns, owner_name), None)
        if not pods:
            return
        for uid in list(pods):
            e = self._pods.get(uid)
            if e is not None:
                self._evict_pod(uid, e)

    def _after_evict(self, uid: str, e: dict) -> None:
        """Subclass hook: drop the evicted pod from business indexes (profile refs / per-job entries)."""

    def _snapshot(self, shard: int) -> str:
        """Subclass hook: serialize one shard's pod data into a snapshot string."""
        raise NotImplementedError

    def _apply_snapshot(self, raw: str | None) -> None:
        """Subclass hook: parse a snapshot string back into the in-memory cache."""
        raise NotImplementedError

    def apply_pod(self, pod) -> None:
        """Subclass hook: apply one pod watch event to the cache."""
        raise NotImplementedError

    def _index_loaded_entry(self, uid: str, e: dict) -> None:
        """Reindex one snapshot entry: legacy shard refill, byte accounting and job-instance index."""
        if "shard" not in e:
            # legacy snapshot (hash-routed, no shard field): refill sequentially
            e["shard"] = self._route_new_shard(serialized_bytes(e))
        self._pods[uid] = e
        size = serialized_bytes(e)
        self._bytes += size
        shard = int(e["shard"])
        self._shard_bytes[shard] = self._shard_bytes.get(shard, 0) + size
        job = (e.get("namespace"), e.get("owner_name"))
        self._job_pods.setdefault(job, {})[uid] = e
        if e.get("owner_uid"):
            self._job_uid[job] = e["owner_uid"]

    def _resume_fill_position(self) -> None:
        """Resume the sequential fill position from the highest non-empty shard (after a snapshot load)."""
        if self._shard_bytes:
            self._active = max(self._shard_bytes)

    def _enforce_cap(self, max_bytes: int) -> None:
        """Evict oldest pod entries beyond max_bytes (kept consistent with the snapshot capacity)."""
        while self._bytes > max_bytes:
            uid = self._order.popleft()
            e = self._pods.get(uid)
            if e is None:
                continue
            self._evict_pod(uid, e)

    # ---- Shared cache-skeleton methods (pod application / TTL GC / snapshot / threads) ---- #

    def _check_job_replacement(self, ns: str, owner_name: str, owner_uid: str) -> None:
        """Purge the old instance's pods when the same ns+name job CR was recreated (new owner UID)."""
        prev_uid = self._job_uid.get((ns, owner_name))
        if prev_uid and prev_uid != owner_uid:
            logger.info(
                "%s job replaced (new owner UID): job=%s ns=%s old_uid=%s new_uid=%s",
                self._log_name,
                owner_name,
                ns,
                prev_uid,
                owner_uid,
            )
            self._purge_job(ns, owner_name)
        self._job_uid[(ns, owner_name)] = owner_uid

    def _store_pod(self, uid: str, entry: dict, old: dict | None, max_bytes: int) -> None:
        """Update FIFO order, route to the sequential shard and write the entry (shared apply_pod tail)."""
        if old is None:
            self._order.append(uid)
        elif uid in self._order:
            self._order.remove(uid)  # refresh eviction position on reschedule/update
            self._order.append(uid)
        if old is not None:
            self._evict_index_only(uid, old)
        # Sequential shard routing: new pods fill the active shard, then advance.
        target = self._active if old is not None else self._route_new_shard(serialized_bytes(entry))
        entry["shard"] = target
        self._pods[uid] = entry
        self._bytes += serialized_bytes(entry)
        self._shard_bytes[target] = self._shard_bytes.get(target, 0) + serialized_bytes(entry)
        self._enforce_cap(max_bytes)

    def apply_pod_deleted(self, pod_uid: str, ts: float | None = None) -> None:
        """Mark a pod deleted (retained within POD_TTL for diagnosis); schedules a CM sync when newly marked."""
        ts = ts if ts is not None else time.time()
        with self._lock:
            e = self._pods.get(pod_uid)
            if e and not e.get("deleted_at"):
                e["deleted_at"] = ts
                logger.info(
                    "%s task pod deleted: job=%s ns=%s pod=%s node=%s",
                    self._log_name,
                    e.get("owner_name"),
                    e.get("namespace"),
                    e.get("pod_name"),
                    e.get("node"),
                )
                mark_dirty(self._cond, self._dirty)  # schedule an immediate CM sync (change-driven)

    def _gc(self, now: float | None = None) -> None:
        """Evict entries deleted beyond POD_TTL (shared TTL GC); schedules a CM sync when anything was removed."""
        now = now if now is not None else time.time()
        changed = False
        with self._lock:
            for uid in list(self._pods):
                e = self._pods[uid]
                if e.get("deleted_at") and now - e["deleted_at"] > POD_TTL:
                    self._evict_pod(uid, e)
                    changed = True
        if changed:
            mark_dirty(self._cond, self._dirty)  # schedule an immediate CM sync (change-driven)

    def _gc_loop(self) -> None:
        while not self._stop.wait(self._gc_interval):
            try:
                self._gc()
            except Exception:  # noqa: BLE001  # GC is best-effort
                logger.exception("%s GC error", self._log_name)

    def _reorder_from_pods(self) -> None:
        """Rebuild the FIFO order after a snapshot load (pod_uid insertion order is not persisted)."""
        with self._lock:
            self._order = deque(self._pods)

    def _load_snapshot(self) -> None:
        try:
            v1 = self._v1 or K8s.core()
            for shard in range(self._shards):
                name = shard_cm_name(self._cm_name, shard)
                try:
                    cm = v1.read_namespaced_config_map(name, self._cm_ns)
                except client.ApiException as e:
                    if e.status == 404:
                        continue  # shard CM not created yet -> empty shard
                    raise
                self._apply_snapshot((cm.data or {}).get(self._cm_key))
        except Exception as e:  # noqa: BLE001
            logger.warning("failed to read %s snapshot CM (start with empty cache): %s", self._log_name, e)
            return  # snapshot missing/no permission -> start with empty cache, rebuilt by watch
        self._reorder_from_pods()

    def _save_snapshot(self) -> bool:
        """Sync each shard snapshot to its CM; writes only shards whose content changed. Best-effort, never raises."""
        v1 = self._v1 or K8s.core()
        ok = True
        for shard in range(self._shards):
            snap = self._snapshot(shard)
            if not shard_has_pods(snap):
                continue  # never create a CM for an empty shard
            if self._last.get(shard) == snap:
                continue
            try:
                name = shard_cm_name(self._cm_name, shard)
                try:
                    v1.read_namespaced_config_map(name, self._cm_ns)
                    v1.patch_namespaced_config_map(name, self._cm_ns, {"data": {self._cm_key: snap}})
                except client.ApiException as e:
                    if e.status != 404:
                        raise
                    v1.create_namespaced_config_map(
                        self._cm_ns,
                        client.V1ConfigMap(
                            api_version="v1",
                            kind="ConfigMap",
                            metadata=client.V1ObjectMeta(name=name),
                            data={self._cm_key: snap},
                        ),
                    )
                self._last[shard] = snap
            except Exception as e:  # noqa: BLE001  # snapshot failure must not block diagnosis
                logger.warning(
                    "failed to save %s snapshot shard %d (diagnosis unaffected): %s", self._log_name, shard, e
                )  # snapshot failure does not block diagnosis; in-memory cache remains usable
                ok = False
        return ok

    def _sync_loop(self) -> None:
        """Change-driven CM sync: syncs as soon as in-memory pod data changed."""
        run_change_driven_sync(self._cond, self._dirty, self._stop, self._save_snapshot)

    def _pod_watch_loop(self) -> None:
        """CoreV1 list+watch pods via the shared k8s.watch_events (automatic reconnection)."""

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

        watch_events(stream, on_event, self._stop, f"{self._log_name} pod")

    # ---- Lifecycle ---- #
    def start(self) -> None:
        self._load_snapshot()
        for target, name in (
            (self._pod_watch_loop, f"{self._log_name}-pod-watch"),
            (self._gc_loop, f"{self._log_name}-gc"),
            (self._sync_loop, f"{self._log_name}-sync"),
        ):
            t = threading.Thread(target=target, name=name, daemon=True)
            t.start()
            self._threads.append(t)
        logger.info("%s started: threads=%s", self._log_name, [t.name for t in self._threads])

    def stop(self) -> None:
        self._stop.set()
        for t in self._threads:
            t.join(timeout=5)
        self._threads.clear()
        self._stop = threading.Event()
        logger.info("%s stopped", self._log_name)


class K8s:
    """Lazy, memoized Kubernetes API access (in-cluster, falling back to kubeconfig)."""

    _core: client.CoreV1Api | None = None

    @classmethod
    def core(cls) -> client.CoreV1Api:
        """Return a memoized CoreV1Api, loading in-cluster/kubeconfig config on first use."""
        if cls._core is None:
            try:
                config.load_incluster_config()
            except Exception:
                config.load_kube_config()
            cls._core = client.CoreV1Api()
        return cls._core

    @classmethod
    def namespace(cls) -> str:
        """Agent-core runtime namespace: env ASCEND_CLUSTEROPS_NS > in-cluster SA namespace file > default mindx-dl."""
        env = os.environ.get("ASCEND_CLUSTEROPS_NS")
        if env:
            return env
        try:
            return Path(_SA_NAMESPACE_FILE).read_text(encoding="utf-8").strip()
        except Exception:
            return DEFAULT_NS


def mark_dirty(condition: threading.Condition, counter: list[int]) -> None:
    """Flag one pending change-driven CM sync and wake up the sync loop.

    ``counter`` is a 1-element list box shared with :func:`run_change_driven_sync`;
    its value is incremented in place and the condition is notified.
    """
    with condition:
        counter[0] += 1
        condition.notify_all()


def run_change_driven_sync(
    condition: threading.Condition,
    counter: list[int],
    stop: threading.Event,
    save: Callable[[], bool],
    backoff: float = 5.0,
) -> None:
    """Change-driven CM sync loop, meant to run in its own thread.

    Waits on ``condition`` until ``counter`` (a 1-element list box flagged via
    :func:`mark_dirty`) is non-zero, consumes it, and calls ``save()``; a failed save
    re-flags the counter and retries after ``backoff`` seconds. Saves once more on exit.
    """
    while not stop.is_set():
        with condition:
            while counter[0] == 0 and not stop.is_set():
                condition.wait()
            if stop.is_set():
                break
            counter[0] = 0
        if not save():
            with condition:
                counter[0] += 1
            stop.wait(backoff)
    save()


def watch_events(
    stream_func: Callable[[], Iterable[dict]],
    on_event: Callable[[str, Any], None],
    stop_event: threading.Event,
    resource_name: str,
    reconnect_delay: int = 5,
) -> None:
    """Kubernetes watch loop that reconnects automatically on failure.

    Iterates the event stream returned by ``stream_func`` and forwards each event to
    ``on_event``; on disconnect or error waits ``reconnect_delay`` seconds and
    reconnects, until ``stop_event`` is set.
    """
    while not stop_event.is_set():
        try:
            for ev in stream_func():
                if stop_event.is_set():
                    break
                on_event(ev["type"], ev["object"])
        except Exception:  # noqa: BLE001  # transient watch errors; reconnect
            if stop_event.wait(reconnect_delay):
                return
            logger.warning("%s watch disconnected, reconnecting in %ds", resource_name, reconnect_delay)
