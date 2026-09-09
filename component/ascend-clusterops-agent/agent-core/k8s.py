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

import logging
import os
import threading
from collections.abc import Iterable
from pathlib import Path
from typing import Any, Callable

import yaml
from kubernetes import client, config

from agent_core.constants import TASK_CRDS_CM_KEY, TASK_CRDS_CM_NAME, TASK_CRDS_CM_NS

logger = logging.getLogger(__name__)

_SA_NAMESPACE_FILE = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
DEFAULT_NS = "mindx-dl"


def load_task_crds_config() -> dict:
    """Read the task CR config CM via the K8s API; empty config + warning on failure (no ownerRef filtering)."""
    try:
        v1 = K8s.core()
        cm = v1.read_namespaced_config_map(TASK_CRDS_CM_NAME, TASK_CRDS_CM_NS)
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
