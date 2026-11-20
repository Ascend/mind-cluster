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

"""Diagnosis CLI contract: ascend_diag stays legacy; `tool diagnose` uses the tool envelope."""

# pylint: disable=redefined-outer-name  # pytest fixtures are injected by parameter name

from __future__ import annotations

import importlib.util
import io
import json
import shutil
import subprocess
import sys
import urllib.error
from importlib.machinery import SourceFileLoader
from pathlib import Path
from types import SimpleNamespace
from urllib.parse import urlparse

import pytest

PLUGIN_DIR = Path(__file__).resolve().parents[1] / "kubectl-plugin"
sys.path.insert(0, str(PLUGIN_DIR))


def _load(filename):
    loader = SourceFileLoader(filename.replace("-", "_"), str(PLUGIN_DIR / filename))
    module = importlib.util.module_from_spec(importlib.util.spec_from_loader(loader.name, loader))
    loader.exec_module(module)
    return module


legacy = _load("kubectl-ascend_diag")
clusterops = _load("kubectl-clusterops")
client = clusterops.client

# The generic CLI only needs a descriptor with the diagnosis input schema.
DIAGNOSE_CATALOG = {
    "request_id": "r1",
    "tools": [
        {
            "name": "diagnose",
            "input_schema": {
                "type": "object",
                "properties": {
                    "job": {"type": "string"},
                    "namespace": {"type": "string", "default": "default"},
                    "refresh": {"type": "boolean", "default": False},
                },
                "required": ["job"],
            },
        }
    ],
}


def _tool_envelope(legacy_result):
    if legacy_result.get("error"):
        status = "error"
    elif legacy_result.get("incomplete_nodes") or legacy_result.get("empty_nodes"):
        status = "partial"
    else:
        status = "ok"
    errors = []
    if status == "error":
        errors = [
            {
                "code": "DIAG_FAILED",
                "source_id": None,
                "path": None,
                "message": legacy_result["error"],
                "retryable": False,
            }
        ]
    return {
        "request_id": "r1",
        "tool": "diagnose",
        "status": status,
        # Every diagnosis tool entry carries the old flat /diag body in data.
        "data": legacy_result,
        "errors": errors,
    }


class _FakeForward:
    local = 12345

    def __init__(self, service):
        self.service = service

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb):
        return None


@pytest.fixture
def transport(monkeypatch):
    # Import the app before replacing subprocess.Popen: importing kubernetes probes
    # platform.system(), which itself uses subprocess on Windows.
    from agent_core import main as _agent_main  # noqa: F401

    state = {
        "requests": [],
        "tool_requests": [],
        "processes": [],
        "legacy": {"final_text": "report", "error": None},
        "tool_status": 200,
        "failure": None,
    }

    def popen(command, **kwargs):
        process = SimpleNamespace(
            command=command,
            stderr=io.StringIO("forward failed"),
            terminated=False,
            wait=lambda **kwargs: 0,
            kill=lambda: None,
        )
        process.terminate = lambda: setattr(process, "terminated", True)
        state["processes"].append(process)
        return process

    def urlopen(request, timeout):
        state["requests"].append((urlparse(request.full_url).path, json.loads(request.data), timeout))
        failure = state["failure"]
        if failure is not None:
            raise failure
        return io.BytesIO(json.dumps(state["legacy"]).encode("utf-8"))

    def request_json(local, method, path, payload=None, timeout=None):
        state["tool_requests"].append((method, path, payload, timeout))
        if state["failure"] is not None:
            raise client.TransportError(str(state["failure"]))
        if method == "GET":
            return 200, DIAGNOSE_CATALOG
        return state["tool_status"], _tool_envelope(state["legacy"])

    monkeypatch.setattr(shutil, "which", lambda *_: "kubectl")
    monkeypatch.setattr(legacy, "_free_port", lambda: 12345)
    monkeypatch.setattr(legacy, "_wait_port_forward", lambda *_: True)
    monkeypatch.setattr(subprocess, "Popen", popen)
    monkeypatch.setattr(urllib.request, "urlopen", urlopen)
    monkeypatch.setattr(client, "PortForward", _FakeForward)
    monkeypatch.setattr(client, "request_json", request_json)
    return state


def _run(monkeypatch, capsys, module, prefix, arguments):
    monkeypatch.setattr(sys, "argv", prefix + arguments)
    code = module.main()
    captured = capsys.readouterr()
    return code, captured.out, captured.err


@pytest.mark.parametrize(
    "legacy_result,expected_code,expected_out,expected_err",
    [
        ({"final_text": "LLM summary", "error": None}, 0, "LLM summary\n", ""),
        ({"final_text": "deterministic report", "error": None}, 0, "deterministic report\n", ""),
        (
            {"final_text": "cached summary", "cached": True, "cached_note": "use --refresh", "error": None},
            0,
            "cached summary\n",
            "use --refresh\n",
        ),
        ({"error": "job is being diagnosed", "pods": [], "diag_report": {}}, 1, "", "job is being diagnosed\n"),
    ],
)
def test_ascend_diag_keeps_the_legacy_business_output(
    transport, monkeypatch, capsys, legacy_result, expected_code, expected_out, expected_err
):
    transport["legacy"] = legacy_result
    code, stdout, stderr = _run(
        monkeypatch,
        capsys,
        legacy,
        ["kubectl-ascend_diag"],
        ["--job", "train-a", "-n", "training", "--refresh", "--agent-core-service", "custom.ops:9703"],
    )

    assert code == expected_code
    assert stdout == expected_out
    assert stderr == expected_err
    assert transport["requests"] == [("/diag", {"job": "train-a", "namespace": "training", "refresh": True}, 1200)]
    assert transport["tool_requests"] == []


def test_ascend_diag_default_parameters_are_unchanged(transport, monkeypatch, capsys):
    code, stdout, _ = _run(monkeypatch, capsys, legacy, ["kubectl-ascend_diag"], ["--job", "train-a"])
    assert code == 0
    assert stdout == "report\n"
    assert transport["requests"] == [("/diag", {"job": "train-a", "namespace": "default", "refresh": False}, 1200)]


@pytest.mark.parametrize(
    "failure", [urllib.error.URLError("offline"), TimeoutError(), json.JSONDecodeError("invalid", "x", 0)]
)
def test_ascend_diag_transport_failures_are_legacy(transport, monkeypatch, capsys, failure):
    transport["failure"] = failure
    code, stdout, stderr = _run(monkeypatch, capsys, legacy, ["kubectl-ascend_diag"], ["--job", "train-a", "--json"])
    assert code == 1
    assert stdout == ""
    assert "diagnosis error" in stderr


@pytest.mark.parametrize("report_text", [None, "raw report body"])
def test_print_diagnosis_result_flat_fallback(capsys, report_text):
    """The fallback printer still shows deterministic text or the original report JSON."""
    report = {"Root_Cluster": {"fault_description": {"code": 109, "string": "AICPU timeout"}}}
    clusterops._print_diagnosis_result(
        {"status": "ok", "data": {"diag_report": report, "diag_report_text": report_text}}
    )
    stdout = capsys.readouterr().out
    if report_text:
        assert report_text in stdout
    else:
        assert json.loads(stdout) == report
    assert "(no output)" not in stdout


def test_generic_diagnose_sends_only_given_params_with_the_diag_timeout(transport, monkeypatch, capsys):
    code, stdout, _ = _run(
        monkeypatch, capsys, clusterops, ["kubectl-clusterops", "tool", "diagnose"], ["--param", "job=train-a"]
    )

    assert code == 0
    assert stdout == "report\n"
    assert transport["tool_requests"] == [
        ("GET", "/tools", None, client.QUERY_TIMEOUT),
        ("POST", "/tools/diagnose", {"job": "train-a"}, client.DIAG_TIMEOUT),
    ]


def test_generic_diagnose_json_is_one_object_with_empty_stderr(transport, monkeypatch, capsys):
    transport["legacy"] = {"final_text": "report", "error": None}
    code, stdout, stderr = _run(
        monkeypatch,
        capsys,
        clusterops,
        ["kubectl-clusterops", "tool", "diagnose"],
        ["--param", "job=train-a", "--json"],
    )

    assert code == 0
    assert json.loads(stdout)["tool"] == "diagnose"
    assert stderr == ""


def test_generic_diagnose_transport_failure_is_a_tool_envelope(transport, monkeypatch, capsys):
    def request_json(local, method, path, payload=None, timeout=None):
        if method == "GET":
            return 200, DIAGNOSE_CATALOG
        raise client.TransportError("offline")

    monkeypatch.setattr(client, "request_json", request_json)
    code, stdout, stderr = _run(
        monkeypatch,
        capsys,
        clusterops,
        ["kubectl-clusterops", "tool", "diagnose"],
        ["--param", "job=train-a", "--json"],
    )

    assert code == 1
    body = json.loads(stdout)
    assert body["tool"] == "diagnose"
    assert body["status"] == "error"
    assert body["errors"][0]["code"] == "TRANSPORT_ERROR"
    assert stderr == ""


def test_generic_diagnose_missing_job_exits_two_without_post(transport, monkeypatch, capsys):
    code, stdout, stderr = _run(monkeypatch, capsys, clusterops, ["kubectl-clusterops", "tool", "diagnose"], ["--json"])

    assert code == 2
    assert stdout == ""
    assert "missing required parameter(s): job" in stderr
    assert transport["tool_requests"] == [("GET", "/tools", None, client.QUERY_TIMEOUT)]


@pytest.mark.parametrize(
    "argv",
    [
        ["diagnose", "--job", "train-a"],
        ["--collect-manifest", "collect.yaml"],
        ["faulty-device"],
        ["isolated-card"],
        ["cluster-resource"],
        ["task-status", "--job", "train-a"],
    ],
)
def test_removed_alias_subcommands_are_rejected(argv, monkeypatch):
    monkeypatch.setattr(sys, "argv", ["kubectl-clusterops"] + argv)
    with pytest.raises(SystemExit) as exc:
        clusterops.main()
    assert exc.value.code == 2


@pytest.mark.parametrize("outcome", ["ok", "missing", "empty", "render-failed", "apply-failed"])
def test_collect_manifest_remains_available_on_ascend_diag(tmp_path, transport, monkeypatch, capsys, outcome):
    manifest = tmp_path / "collect.yaml"
    if outcome != "missing":
        manifest.write_text("" if outcome == "empty" else "version: 1\n", encoding="utf-8")

    def run(command, **kwargs):
        failed = outcome == "render-failed" and "create" in command or outcome == "apply-failed" and "apply" in command
        return SimpleNamespace(returncode=1 if failed else 0, stdout="rendered manifest", stderr="mock failure")

    monkeypatch.setattr(subprocess, "run", run)
    code, stdout, stderr = _run(
        monkeypatch, capsys, legacy, ["kubectl-ascend_diag"], ["--collect-manifest", str(manifest)]
    )
    assert code == (0 if outcome == "ok" else 1)
    assert transport["requests"] == [] and transport["tool_requests"] == []
    if outcome == "ok":
        assert "subsequent diagnostics" in stdout
    else:
        assert "error:" in stderr


def test_missing_job_shows_help_without_execution(transport, monkeypatch, capsys):
    monkeypatch.setattr(sys, "argv", ["kubectl-ascend_diag"])
    assert legacy.main() == 2
    assert "--collect-manifest" in capsys.readouterr().out
    assert transport["requests"] == [] and transport["tool_requests"] == []


@pytest.mark.parametrize("mode", ["model", "no-model", "model-failed", "cached"])
def test_generic_diagnose_applies_presentation_automatically(transport, monkeypatch, capsys, mode):
    from agent_core import llm_summary, main
    from agent_core.tools import diagnosis as tools
    from fastapi.testclient import TestClient

    summaries = []
    writes = []
    raw = {"diag_report": {"root": "node-a"}, "diag_report_text": "raw report", "error": None}
    if mode == "cached":
        raw.update(final_text="cached summary", cached=True)
    monkeypatch.setattr(tools, "diagnose_cached", lambda **kwargs: dict(raw))
    monkeypatch.setattr(tools, "cache_diag_final_text", lambda *args: writes.append(args))

    def summarize(report):
        summaries.append(report)
        if mode == "model-failed":
            raise llm_summary.LLMSummaryError("model unavailable")
        return None if mode == "no-model" else "model summary"

    monkeypatch.setattr(llm_summary, "summarize_report", summarize)
    http = TestClient(main.app)

    expected = {
        "model": "model summary",
        "no-model": "raw report",
        "model-failed": "model unavailable",
        "cached": "cached summary",
    }[mode]

    tool_response = http.post("/tools/diagnose", json={"job": "train-a"})
    assert tool_response.status_code == 200
    tool_data = tool_response.json()["data"]
    assert raw["diag_report"] == tool_data["diag_report"]
    assert tool_data["diag_report_text"] == "raw report"
    assert expected in tool_data["final_text"]
    assert len(summaries) == (0 if mode == "cached" else 1)

    summaries.clear()
    writes.clear()

    def request_json(local, method, path, payload=None, timeout=None):
        response = http.get(path) if method == "GET" else http.post(path, json=payload)
        return response.status_code, response.json()

    monkeypatch.setattr(client, "request_json", request_json)
    code, stdout, _ = _run(
        monkeypatch,
        capsys,
        clusterops,
        ["kubectl-clusterops", "tool", "diagnose"],
        ["--param", "job=train-a", "--json"],
    )
    assert code == 0
    body = json.loads(stdout)
    assert expected in body["data"]["final_text"]
    assert len(summaries) == (0 if mode == "cached" else 1)
    assert len(writes) == (0 if mode == "cached" else 1)
