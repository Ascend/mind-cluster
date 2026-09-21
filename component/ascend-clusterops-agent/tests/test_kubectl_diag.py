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

"""kubectl-ascend_diag plugin unit tests (mocks kubectl/port-forward, no cluster needed)."""

from __future__ import annotations

import importlib.machinery
import importlib.util
import os
from types import SimpleNamespace

import pytest

PLUGIN_PATH = os.path.join(
    os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
    "kubectl-plugin",
    "kubectl-ascend_diag",
)


def _load_plugin():
    """Load the extension-less plugin script as an importable module."""
    loader = importlib.machinery.SourceFileLoader("kubectl_ascend_diag", PLUGIN_PATH)
    spec = importlib.util.spec_from_loader("kubectl_ascend_diag", loader)
    mod = importlib.util.module_from_spec(spec)
    loader.exec_module(mod)
    return mod


@pytest.fixture(name="plugin_env")
def _plugin_env_factory(monkeypatch):
    """Plugin with kubectl/port-forward/HTTP stubbed; returns (module, captured_popen_cmds)."""
    mod = _load_plugin()
    captured: list[list[str]] = []

    def fake_popen(cmd, **kwargs):
        captured.append(cmd)
        return SimpleNamespace(
            stdout=None,
            stderr=None,
            poll=lambda: None,
            terminate=lambda: None,
            wait=lambda *a, **k: 0,
            kill=lambda: None,
        )

    monkeypatch.setattr(mod.shutil, "which", lambda *_a, **_k: "/usr/bin/kubectl")
    monkeypatch.setattr(mod, "_free_port", lambda: 12345)
    monkeypatch.setattr(mod, "_wait_port_forward", lambda *_a, **_k: True)
    monkeypatch.setattr(mod, "_post_diag", lambda *_a, **_k: {"final_text": "ok", "cached": False})
    monkeypatch.setattr(mod.subprocess, "Popen", fake_popen)
    return mod, captured


def _run(plugin_env, svc: str | None):
    mod, _captured = plugin_env
    args = SimpleNamespace(
        agent_core_service=svc,
        job="job1",
        namespace="ns1",
        refresh=False,
        json=False,
    )
    rc = mod._run_remote(args)
    return mod, _captured, rc


def test_default_service_uses_9700(plugin_env):
    """Default svc (no --agent-core-service) must port-forward to 9700 in mindx-dl."""
    _mod, captured, rc = _run(plugin_env, None)
    assert rc == 0
    cmd = captured[0]
    assert cmd[0] == "/usr/bin/kubectl"
    assert cmd[1] == "port-forward"
    assert "-n" in cmd and cmd[cmd.index("-n") + 1] == "mindx-dl"
    assert "svc/agent-core" in cmd
    assert cmd[-1] == "12345:9700"


def test_custom_service_port_used(plugin_env):
    """--agent-core-service svc.ns:9703 must forward to 9703, not the hardcoded 9700."""
    _mod, captured, rc = _run(plugin_env, "agent-core.mindx-dl:9703")
    assert rc == 0
    assert captured[0][-1] == "12345:9703"


def test_custom_service_without_port_defaults_9700(plugin_env):
    """--agent-core-service svc.ns (no port) must fall back to the default 9700."""
    _mod, captured, rc = _run(plugin_env, "agent-core.mindx-dl")
    assert rc == 0
    assert captured[0][-1] == "12345:9700"


def test_invalid_service_port_returns_error(plugin_env, capsys):
    """A non-numeric port must fail gracefully without starting port-forward."""
    mod, captured, rc = _run(plugin_env, "agent-core.mindx-dl:abc")
    assert rc == 1
    assert captured == []
    assert "diagnosis error, invalid agent-core service" in capsys.readouterr().err
