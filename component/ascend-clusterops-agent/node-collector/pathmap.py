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

"""Collector-local pod -> host_paths table (RFC §3.3.5; data shipped per collect request).

agent-core resolves task-pod host:container mount pairs and env values in its pathmap and
ships them inside the TriggerCollect request (CollectRequest.mounts). The collector caches
that per-request data locally by pod_uid; no global ConfigMap is watched, so cluster size
does not drive per-node memory or API-server broadcast.
"""

# pylint: disable=duplicate-code  # watch lifecycle mirrors relcache

from __future__ import annotations

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
    """Per-request pod -> {pairs, env, pod_ip} table (data shipped inside the TriggerCollect request).

    update_mounts merges the CollectRequest mounts (keyed by pod_uid); the collector matches
    manifest entities against all_pairs (paths prefix / mount_keywords substring), pod_env
    (env entity container->host reverse lookup) and pod_ip (shared-storage plog subpath
    matching). Keyed by pod_uid, so concurrent collects of different jobs never overwrite
    each other.
    """

    def __init__(self):
        # pod_uid -> {"pairs": [...], "env": {...}, "pod_ip": ""}
        self._pods: dict[str, dict] = {}
        self._lock = threading.RLock()

    def update_mounts(self, mounts) -> None:
        """Merge the pod mounts shipped in a CollectRequest into the local table."""
        with self._lock:
            for m in mounts:
                self._pods[m.pod_uid] = {
                    "pairs": list(m.pairs or []),
                    "env": dict(m.env or {}),
                    "pod_ip": getattr(m, "pod_ip", "") or "",
                }

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
            return list(pod.get("pairs", []) or [])

    def pod_env(self, pod_uid: str, env_name: str) -> str | None:
        """Return the pod's env value shipped in the collect request (retained for deleted pods within TTL)."""
        with self._lock:
            pod = self._pods.get(pod_uid)
            if not pod:
                return None
            return (pod.get("env") or {}).get(env_name)

    def pod_ip(self, pod_uid: str) -> str:
        """Return the pod's IP shipped in the collect request ('' for unknown pods)."""
        with self._lock:
            pod = self._pods.get(pod_uid)
            if not pod:
                return ""
            return pod.get("pod_ip", "")


# Singleton (lazily created; data injected per TriggerCollect request)
_pm: PathMap | None = None


def get_pathmap() -> PathMap:
    global _pm
    if _pm is None:
        _pm = PathMap()
    return _pm


def init_pathmap() -> PathMap:
    return get_pathmap()


def shutdown_pathmap() -> None:
    global _pm
    _pm = None


def all_pairs(pod_uid: str) -> list[str]:
    return get_pathmap().all_pairs(pod_uid)
