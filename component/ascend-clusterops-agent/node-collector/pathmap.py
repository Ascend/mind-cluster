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

"""Collector-local pod -> host_paths table (RFC §3.3.5; data sourced from the global CM).

agent-core's pathmap-writer scans all pods, parses host:container mount pairs, de-duplicates
by fingerprint, and writes the single global CM `clusterops-pathmap` (cluster-system). The
collector watches this CM and caches it in memory. After a pod dies, the agent side keeps
deleted_at within POD_TTL (default 7d), and the collector mirrors it in memory.
"""

# pylint: disable=duplicate-code  # watch lifecycle mirrors relcache

from __future__ import annotations

import json
import logging
import os
import threading
from typing import Any

import yaml
from kubernetes import client, config, watch

logger = logging.getLogger(__name__)

MANIFEST_PATH = os.environ.get("COLLECT_MANIFEST", "/home/hwMindX/collect_manifest.yaml")
MANIFEST_CM_NAME = os.environ.get("MANIFEST_CM", "collect-manifest")
MANIFEST_CM_NS = os.environ.get("MANIFEST_CM_NAMESPACE", "cluster-system")
MANIFEST_CM_KEY = "collect_manifest.yaml"
PATHMAP_CM_NAME = os.environ.get("PATHMAP_CM_NAME", "clusterops-pathmap")
PATHMAP_CM_NS = os.environ.get("PATHMAP_CM_NS", "cluster-system")
PATHMAP_CM_KEY = "pathmap.json"


def _k8s_core():
    try:
        config.load_incluster_config()
    except Exception:
        config.load_kube_config()
    return client.CoreV1Api()


def _load_manifest_cm() -> dict[str, Any] | None:
    """Prefer reading ConfigMap <MANIFEST_CM_NAME>.collect_manifest.yaml (user site override).

    CM missing / no read permission / missing content / parse failure -> return None, and the
    caller falls back to the file built into the image.
    """
    try:
        v1 = _k8s_core()
        cm = v1.read_namespaced_config_map(MANIFEST_CM_NAME, MANIFEST_CM_NS)
        data = (cm.data or {}).get(MANIFEST_CM_KEY)
        if data:
            return yaml.safe_load(data)
    except Exception:  # noqa: BLE001  # nosec B110: fall back to the local file when the CM cannot be fetched
        pass
    return None


def load_manifest() -> dict[str, Any]:
    """最新采集契约 (CM 覆盖 -> 镜像内置文件兜底), 由 ManifestWatcher 缓存维持。

    serve() 启动时经 init_manifest_watch() 加载初始值并 watch CM; 采集线程调用此函数
    拿到的是始终最新的缓存。
    """
    return get_manifest_watcher().current()


class ManifestWatcher:
    """维护内存中的采集契约 (collect-manifest CM 覆盖 -> 镜像内置文件兜底)。

    镜像内置 /home/hwMindX/collect_manifest.yaml 为默认值; 集群 ConfigMap <collect-manifest>
    (cluster-system, 由 kubectl ascend_diag --collect-manifest 写入) 覆盖之。watcher 在 CM
    变化时更新缓存并把生效的最新数据打印到日志, 便于运维确认覆盖内容。
    """

    def __init__(self):
        self._data: dict[str, Any] = {}
        self._lock = threading.RLock()
        self._stop = threading.Event()
        self._thread: threading.Thread | None = None
        self._v1 = None  # injectable by tests

    def _read_file(self) -> dict[str, Any]:
        try:
            with open(MANIFEST_PATH, encoding="utf-8") as f:
                return yaml.safe_load(f)
        except FileNotFoundError:
            return {}

    @staticmethod
    def _summary(data: dict[str, Any]) -> str:
        return (
            f"manifest_version={data.get('manifest_version')} "
            f"entities={[e.get('name') for e in data.get('entities', [])]}"
        )

    def _apply(self, data: dict[str, Any]) -> bool:
        with self._lock:
            changed = data != self._data
            self._data = data
            return changed

    def current(self) -> dict[str, Any]:
        with self._lock:
            return self._data

    def _load(self) -> None:
        data = _load_manifest_cm()
        if data is None:
            data = self._read_file()
        self._apply(data)

    def _watch_loop(self) -> None:
        while not self._stop.is_set():
            try:
                v1 = self._v1 or _k8s_core()
                w = watch.Watch()
                for ev in w.stream(
                    v1.list_namespaced_config_map,
                    MANIFEST_CM_NS,
                    field_selector=f"metadata.name={MANIFEST_CM_NAME}",
                    timeout_seconds=300,
                    _request_timeout=320,
                ):
                    if self._stop.is_set():
                        break
                    typ = ev["type"]
                    if typ == "DELETED":
                        self._apply(self._read_file())
                        logger.warning(
                            "collect-manifest CM deleted, fall back to the built-in file: %s",
                            self._summary(self._data),
                        )
                        continue
                    raw = (ev["object"].data or {}).get(MANIFEST_CM_KEY)
                    if not raw:
                        logger.warning("collect-manifest CM event without data, skipping: type=%s", typ)
                        continue
                    try:
                        data = yaml.safe_load(raw)
                    except yaml.YAMLError:
                        logger.warning("collect-manifest CM event with unparseable data, skipping: type=%s", typ)
                        continue
                    changed = self._apply(data)
                    if typ == "MODIFIED" and changed:
                        logger.info(
                            "collect-manifest CM updated: type=MODIFIED %s changed=True",
                            self._summary(data),
                        )
            except Exception:  # noqa: BLE001
                if self._stop.wait(5):
                    return
                logger.warning("collect-manifest CM watch disconnected, reconnecting in 5s")

    def start(self) -> None:
        self._load()
        self._thread = threading.Thread(target=self._watch_loop, name="collect-manifest-watch", daemon=True)
        self._thread.start()
        logger.info(
            "collect-manifest watch started: CM %s/%s (fallback file %s)",
            MANIFEST_CM_NS,
            MANIFEST_CM_NAME,
            MANIFEST_PATH,
        )

    def stop(self) -> None:
        self._stop.set()
        if self._thread is not None:
            self._thread.join(timeout=5)
        self._stop = threading.Event()
        logger.info("collect-manifest watch stopped")


_mw: ManifestWatcher | None = None


def get_manifest_watcher() -> ManifestWatcher:
    global _mw
    if _mw is None:
        _mw = ManifestWatcher()
    return _mw


def init_manifest_watch() -> ManifestWatcher:
    watcher = get_manifest_watcher()
    watcher.start()
    return watcher


def shutdown_manifest_watch() -> None:
    if _mw is not None:
        _mw.stop()


class PathMap:
    """pod_uid -> all host:container mount pairs (data sourced from the global CM, cached locally only).

    _update_cache/_load can be called directly by tests; the collector matches manifest
    entities against all_pairs (paths prefix / mount_keywords substring).
    """

    def __init__(self):
        self._profiles: dict[str, dict] = {}  # profile_id -> {"pairs": [host:container, ...]}
        self._pods: dict[str, dict] = {}  # pod_uid -> {profile, deleted_at}
        self._lock = threading.RLock()
        self._stop = threading.Event()
        self._threads: list[threading.Thread] = []
        self._v1 = None  # injectable by tests

    # ---- CM data application (pure logic, callable directly by tests) ---- #
    def _update_cache(self, data: dict) -> None:
        with self._lock:
            self._profiles = dict(data.get("profiles") or {})
            self._pods = dict(data.get("pods") or {})

    def _load(self) -> None:
        try:
            v1 = self._v1 or _k8s_core()
            cm = v1.read_namespaced_config_map(PATHMAP_CM_NAME, PATHMAP_CM_NS)
        except Exception as e:  # noqa: BLE001
            logger.warning("failed to read pathmap CM (start with an empty table): %s", e)
            return  # CM not ready -> start with an empty table, backfill from later watch events
        raw = (cm.data or {}).get(PATHMAP_CM_KEY)
        if not raw:
            return
        try:
            data = json.loads(raw)
        except json.JSONDecodeError:
            return
        self._update_cache(data)

    # ---- Queries ---- #
    def all_pairs(self, pod_uid: str) -> list[str]:
        """pod_uid -> all host:container mount pairs (all hostPath mounts of the task pod).

        The collect contract is matched collector-side against these pairs (paths prefix /
        mount_keywords substring); agent-core stores every mount pair, no per-entity resolution.
        """
        with self._lock:
            pod = self._pods.get(pod_uid)
            if not pod:
                return []
            profile = self._profiles.get(pod.get("profile", ""), {})
            return list(profile.get("pairs", []) or [])

    def pod_env(self, pod_uid: str, env_name: str) -> str | None:
        """Return the pod's recorded env value from the pathmap CM (retained for deleted pods within TTL)."""
        with self._lock:
            pod = self._pods.get(pod_uid)
            if not pod:
                return None
            return (pod.get("env") or {}).get(env_name)

    # ---- CM watch ---- #
    def _cm_watch_loop(self) -> None:
        while not self._stop.is_set():
            try:
                v1 = self._v1 or _k8s_core()
                w = watch.Watch()
                for ev in w.stream(
                    v1.list_namespaced_config_map,
                    PATHMAP_CM_NS,
                    field_selector=f"metadata.name={PATHMAP_CM_NAME}",
                    timeout_seconds=300,
                    _request_timeout=320,
                ):
                    if self._stop.is_set():
                        break
                    typ = ev["type"]
                    if typ == "DELETED":
                        logger.warning("pathmap CM deleted, cache cleared")
                        self._update_cache({})
                        continue
                    raw = (ev["object"].data or {}).get(PATHMAP_CM_KEY)
                    if not raw:
                        logger.warning("pathmap CM event without data, skipping: type=%s", typ)
                        continue
                    try:
                        data = json.loads(raw)
                    except json.JSONDecodeError:
                        logger.warning("pathmap CM event with unparseable data, skipping: type=%s", typ)
                        continue
                    with self._lock:
                        changed = (data.get("profiles") or {}) != self._profiles or (
                            data.get("pods") or {}
                        ) != self._pods
                    # Log only real updates (MODIFIED with content change); unchanged ADDED replays are silent.
                    if typ == "MODIFIED" and changed:
                        logger.info(
                            "pathmap CM updated: type=MODIFIED profiles=%d pods=%d changed=True",
                            len(data.get("profiles") or {}),
                            len(data.get("pods") or {}),
                        )
                    self._update_cache(data)
            except Exception:  # noqa: BLE001
                if self._stop.wait(5):
                    return
                logger.warning("pathmap CM watch disconnected, reconnecting in 5s")

    def start(self) -> None:
        self._load()
        t = threading.Thread(target=self._cm_watch_loop, name="pathmap-cm-watch", daemon=True)
        t.start()
        self._threads.append(t)
        logger.info("pathmap started: CM watch %s/%s", PATHMAP_CM_NS, PATHMAP_CM_NAME)

    def stop(self) -> None:
        self._stop.set()
        for t in self._threads:
            t.join(timeout=5)
        self._threads.clear()
        self._stop = threading.Event()
        logger.info("pathmap stopped")


# Singleton (started by collector serve)
_pm: PathMap | None = None


def get_pathmap() -> PathMap:
    global _pm
    if _pm is None:
        _pm = PathMap()
    return _pm


def init_pathmap() -> PathMap:
    c = get_pathmap()
    c.start()
    return c


def shutdown_pathmap() -> None:
    if _pm is not None:
        _pm.stop()


def all_pairs(pod_uid: str) -> list[str]:
    return get_pathmap().all_pairs(pod_uid)
