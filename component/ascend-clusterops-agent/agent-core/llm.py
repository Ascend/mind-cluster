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

"""LLM client manager: standard OpenAI-compatible interface.

The main flow is fixed as the deterministic diagnose tool and does not depend on the
LLM. The LLM only serves as an optional "agent wrapper" and report summarizer; without
a configured key everything degrades gracefully and the main flow still runs standalone.

LLMManager is the process-level singleton: it builds the client lazily and caches it,
rebuilding only when the LLM config changed. The config is kept in an in-memory cache
that is refreshed in REAL TIME by a background watch on the llm-secret (agent_core main
starts it via init_watch()), so `kubectl clusterops --create-llm-config` /
`--clear-llm-config` take effect on the next call without a pod restart AND without a
k8s API read per call. When the watch is not running (e.g. local CLI debugging), a
one-shot read falls back to env vars (LLM_API_KEY / LLM_BASE_URL / LLM_MODEL).
Building a new client logs base_url/model so operators see which endpoint/model is
actually used.
"""

from __future__ import annotations

import base64
import logging
import os
import threading
from collections.abc import Iterable

from kubernetes import client, watch
from langchain_openai import ChatOpenAI

from agent_core.k8s import K8s, watch_events

logger = logging.getLogger(__name__)

# llm-secret name matches the kubectl-clusterops contract (keys: api-key/base-url/model)
LLM_SECRET_NAME = "llm-secret"  # nosec B105


class LLMManager:
    """Lazy-built, config-cached OpenAI-compatible LLM client + real-time llm-secret watch.

    Process-level singleton (``_manager`` below); the module-level ``get_llm`` /
    ``init_watch`` / ``shutdown_watch`` delegate to it so callers keep the same API.
    """

    def __init__(self):
        # Client cache: rebuilt only when the config it was built from changes.
        self._llm: ChatOpenAI | None = None
        self._cache_base_url: str | None = None
        self._cache_api_key: str | None = None
        self._cache_model: str | None = None
        self._lock = threading.Lock()
        # Latest LLM config, refreshed in real time by the llm-secret watch (no k8s API read per call).
        self._config_cache: tuple[str, str, str] | None = None
        self._config_lock = threading.Lock()
        self._stop = threading.Event()
        self._watch_thread: threading.Thread | None = None

    # ---- Config helpers (pure logic, callable by tests) ---- #
    @staticmethod
    def _dec(value: str) -> str:
        """Decode a base64 secret value (k8s Secret.data is base64); empty on any failure."""
        try:
            return base64.b64decode(value).decode("utf-8", errors="replace")
        except Exception:
            return ""

    @staticmethod
    def _env_config() -> tuple[str, str, str]:
        """Fallback LLM config read from environment variables (local debugging / k8s API unavailable)."""
        return (
            os.environ.get("LLM_BASE_URL", ""),
            os.environ.get("LLM_API_KEY", ""),
            os.environ.get("LLM_MODEL", ""),
        )

    @staticmethod
    def _secret_to_config(secret) -> tuple[str, str, str]:
        """Extract (base_url, api_key, model) from a V1Secret (its data is base64)."""
        data = secret.data or {}
        return (
            LLMManager._dec(data.get("base-url", "")),
            LLMManager._dec(data.get("api-key", "")),
            LLMManager._dec(data.get("model", "")),
        )

    # ---- Config read / cache ---- #
    def _read_secret_once(self) -> tuple[str, str, str]:
        """One-shot read of llm-secret via the k8s API; 404 -> no config; env fallback on API failure."""
        try:
            v1 = K8s.core()
            secret = v1.read_namespaced_secret(LLM_SECRET_NAME, K8s.namespace())
        except client.exceptions.ApiException as e:
            if e.status == 404:  # llm-secret removed (--clear-llm-config) -> no LLM config
                return "", "", ""
            logger.warning("read llm-secret via k8s failed, fall back to env: %s", e)
            return self._env_config()
        except Exception as e:  # noqa: BLE001  # no cluster / no kubeconfig -> env fallback
            logger.warning("read llm-secret via k8s failed, fall back to env: %s", e)
            return self._env_config()
        return self._secret_to_config(secret)

    def _refresh_cache(self, cfg: tuple[str, str, str]) -> None:
        """Store the latest LLM config in the in-memory cache (called by the watch / one-shot read)."""
        with self._config_lock:
            if self._config_cache == cfg:
                return
            self._config_cache = cfg
        logger.info(
            "llm-secret config updated: base_url=%s api_key=%s model=%s",
            cfg[0],
            "set" if cfg[1] else "cleared",
            cfg[2],
        )

    def _read_llm_config(self) -> tuple[str, str, str]:
        """Return the in-memory LLM config cache (refreshed in real time by the llm-secret watch).

        If the cache is not populated yet (watch not started / initial snapshot pending), fall back
        to a one-shot read. Returns (base_url, api_key, model); an empty api_key means deterministic mode.
        """
        with self._config_lock:
            cfg = self._config_cache
        if cfg is not None:
            return cfg
        cfg = self._read_secret_once()
        self._refresh_cache(cfg)
        return cfg

    # ---- Client build ---- #
    def get_llm(self) -> ChatOpenAI | None:
        """Return the cached LLM client, rebuilding it when the LLM config changed.

        Without a complete config (api-key/base-url/model are all required, no defaults)
        returns None (the main flow degrades to deterministic mode).
        """
        base_url, api_key, model = self._read_llm_config()
        with self._lock:
            if not (base_url and api_key and model):
                if self._llm is not None:
                    logger.info("LLM config incomplete (api-key/base-url/model): falling back to deterministic mode")
                self._llm = None
                self._cache_base_url = self._cache_api_key = self._cache_model = None
                return None
            if self._llm is not None and (base_url, api_key, model) == (
                self._cache_base_url,
                self._cache_api_key,
                self._cache_model,
            ):
                return self._llm
            logger.info("building LLM client: base_url=%s model=%s", base_url, model)
            self._llm = ChatOpenAI(
                base_url=base_url,
                api_key=api_key,
                model=model,
                temperature=0.2,
            )
            self._cache_base_url, self._cache_api_key, self._cache_model = base_url, api_key, model
            return self._llm

    # ---- llm-secret watch ---- #
    def _watch_loop(self) -> None:
        """Background thread: watch llm-secret changes and refresh the cache in real time."""

        def stream() -> Iterable[dict]:
            return watch.Watch().stream(
                K8s.core().list_namespaced_secret,
                K8s.namespace(),
                field_selector=f"metadata.name={LLM_SECRET_NAME}",
                timeout_seconds=300,
                _request_timeout=320,
            )

        def on_event(typ: str, obj) -> None:
            if typ == "DELETED":
                self._refresh_cache(("", "", ""))
            else:  # ADDED / MODIFIED
                self._refresh_cache(self._secret_to_config(obj))

        watch_events(stream, on_event, self._stop, "llm-secret")

    def init_watch(self) -> None:
        """Start the llm-secret watch thread (idempotent); called by agent-core main lifespan."""
        with self._config_lock:
            if self._watch_thread is not None:
                return
            self._stop.clear()
            t = threading.Thread(target=self._watch_loop, name="llm-secret-watch", daemon=True)
            t.start()
            self._watch_thread = t
        logger.info("llm-secret watch started")

    def shutdown_watch(self) -> None:
        """Stop the llm-secret watch thread (agent-core main lifespan shutdown)."""
        with self._config_lock:
            t = self._watch_thread
            self._watch_thread = None
        if t is not None:
            self._stop.set()
            t.join(timeout=5)
        self._stop.clear()


# Process-level singleton; module-level functions delegate to it.
_manager = LLMManager()


def get_llm() -> ChatOpenAI | None:
    """Return the cached LLM client (delegates to LLMManager.get_llm)."""
    return _manager.get_llm()


def init_watch() -> None:
    """Start the llm-secret watch thread (delegates to LLMManager.init_watch)."""
    _manager.init_watch()


def shutdown_watch() -> None:
    """Stop the llm-secret watch thread (delegates to LLMManager.shutdown_watch)."""
    _manager.shutdown_watch()
