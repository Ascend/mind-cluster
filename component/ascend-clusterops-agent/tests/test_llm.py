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

"""agent.llm: get_llm() lazy build + cache rebuild on config change; llm-secret watch real-time config.

LLM config is refreshed in real time by a watch on llm-secret (no k8s API read per call); these
tests mock K8s.core / _read_llm_config so no cluster is needed.
"""

from __future__ import annotations

import base64
import importlib
from types import SimpleNamespace

import pytest
from kubernetes import client

from agent_core import k8s, llm


@pytest.fixture(autouse=True)
def _reset_config_cache():
    """Reset the config cache / watch state before each test to avoid cross-test pollution."""
    llm._manager._config_cache = None  # pylint: disable=protected-access
    llm._manager._watch_thread = None  # pylint: disable=protected-access
    yield


# --------------------------------------------------------------------------- #
# get_llm: lazy build + rebuild on config change (mocks _read_llm_config)
# --------------------------------------------------------------------------- #
def test_get_llm_none_without_key(monkeypatch):
    importlib.reload(llm)
    monkeypatch.setattr(llm._manager, "_read_llm_config", lambda: ("", "", ""))
    try:
        assert llm.get_llm() is None
    finally:
        importlib.reload(llm)


def test_get_llm_none_without_base_url_or_model(monkeypatch):
    # no defaults: missing base-url/model (even with api-key) -> deterministic mode (None)
    importlib.reload(llm)
    monkeypatch.setattr(llm._manager, "_read_llm_config", lambda: ("", "sk-test", ""))
    try:
        assert llm.get_llm() is None
    finally:
        importlib.reload(llm)


def test_get_llm_built_with_key(monkeypatch):
    importlib.reload(llm)
    monkeypatch.setattr(
        llm._manager, "_read_llm_config", lambda: ("https://example.invalid/v1", "sk-test", "test-model")
    )
    try:
        client_obj = llm.get_llm()
        assert client_obj is not None
    finally:
        importlib.reload(llm)


def test_get_llm_cache_reuse_and_rebuild(monkeypatch):
    """Config unchanged -> reuse cache; config changed -> rebuild; key cleared -> None."""
    importlib.reload(llm)
    cfg = {"base_url": "https://u", "api_key": "sk-test", "model": "model-a"}
    monkeypatch.setattr(llm._manager, "_read_llm_config", lambda: (cfg["base_url"], cfg["api_key"], cfg["model"]))
    try:
        first = llm.get_llm()
        assert first is not None
        assert llm.get_llm() is first  # config unchanged -> reuse the cached client

        cfg["model"] = "model-b"
        second = llm.get_llm()
        assert second is not None and second is not first  # config changed -> rebuild

        cfg["api_key"] = ""
        assert llm.get_llm() is None  # key cleared (--clear-llm-config) -> deterministic mode
    finally:
        importlib.reload(llm)


# --------------------------------------------------------------------------- #
# Config cache: after the watch refreshes it, get_llm only reads memory, not the k8s API
# --------------------------------------------------------------------------- #
def test_read_llm_config_served_from_cache(monkeypatch):
    # Cache populated by the watch -> _read_llm_config returns directly, no k8s API call
    llm._manager._config_cache = ("https://cached/v1", "sk-cached", "cached-model")  # pylint: disable=protected-access

    def _boom():
        raise AssertionError("should not hit k8s API when cache is populated")

    monkeypatch.setattr(k8s.K8s, "core", classmethod(lambda cls: _boom()))
    assert llm._manager._read_llm_config() == ("https://cached/v1", "sk-cached", "cached-model")


def test_refresh_cache_updates_and_logs():
    llm._manager._refresh_cache(("https://u", "sk", "m"))  # pylint: disable=protected-access
    assert llm._manager._config_cache == ("https://u", "sk", "m")  # pylint: disable=protected-access
    llm._manager._refresh_cache(("", "", ""))  # secret deleted -> cleared
    assert llm._manager._config_cache == ("", "", "")  # pylint: disable=protected-access


def test_init_shutdown_watch_idempotent(monkeypatch):
    importlib.reload(llm)
    monkeypatch.setattr(k8s.K8s, "core", classmethod(lambda cls: (_ for _ in ()).throw(RuntimeError("no cluster"))))
    try:
        llm.init_watch()
        llm.init_watch()  # idempotent: does not start a second thread
        assert llm._manager._watch_thread is not None  # pylint: disable=protected-access
        llm.shutdown_watch()
        assert llm._manager._watch_thread is None  # pylint: disable=protected-access
    finally:
        llm.shutdown_watch()
        importlib.reload(llm)


# --------------------------------------------------------------------------- #
# _read_llm_config one-shot read: llm-secret via k8s API + env fallback (when the watch has not populated the cache)
# --------------------------------------------------------------------------- #
def _fake_v1(secret_result):
    class _FakeV1:
        def read_namespaced_secret(self, name, namespace):
            if isinstance(secret_result, Exception):
                raise secret_result
            return secret_result

    return _FakeV1()


def _mock_core(monkeypatch, secret_result):
    """Mock the shared K8s.core() to return a fake CoreV1Api (no real cluster needed)."""
    monkeypatch.setattr(k8s.K8s, "core", classmethod(lambda cls: _fake_v1(secret_result)))


def test_read_llm_config_from_secret(monkeypatch):
    secret = SimpleNamespace(
        data={
            "api-key": base64.b64encode(b"sk-test").decode(),
            "base-url": base64.b64encode(b"https://example.invalid/v1").decode(),
            "model": base64.b64encode(b"test-model").decode(),
        }
    )
    _mock_core(monkeypatch, secret)
    assert llm._manager._read_llm_config() == ("https://example.invalid/v1", "sk-test", "test-model")


def test_read_llm_config_no_defaults_when_fields_missing(monkeypatch):
    # base-url/model missing -> empty (no defaults), get_llm degrades to deterministic mode
    secret = SimpleNamespace(data={"api-key": base64.b64encode(b"sk").decode()})
    _mock_core(monkeypatch, secret)
    assert llm._manager._read_llm_config() == ("", "sk", "")


def test_read_llm_config_secret_missing_404(monkeypatch):
    # secret missing (404) after --clear-llm-config -> no LLM config
    _mock_core(monkeypatch, client.exceptions.ApiException(status=404))
    assert llm._manager._read_llm_config() == ("", "", "")


def test_read_llm_config_env_fallback(monkeypatch):
    # k8s API unavailable (no cluster) -> fall back to env vars
    _mock_core(monkeypatch, RuntimeError("no cluster"))
    monkeypatch.delenv("LLM_API_KEY", raising=False)
    monkeypatch.setenv("LLM_BASE_URL", "https://env.invalid/v1")
    monkeypatch.setenv("LLM_MODEL", "env-model")
    assert llm._manager._read_llm_config() == ("https://env.invalid/v1", "", "env-model")
