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

"""Diagnosis adapter: direct reuse of the guarded chain and unchanged result contracts."""

from __future__ import annotations

import pytest
from agent_core.tools import diagnosis as tools
from agent_core.query_api import _status_for
from agent_core.tools.diagnosis_adapter import DiagnosisContext, run_diagnosis
from agent_core.tools.registry import TOOL_REGISTRY, ToolInvocation
from agent_core.tools.schemas import DiagnosisResult


def _stub(monkeypatch, result):
    monkeypatch.setattr(tools, "diagnose_cached", lambda **kwargs: dict(result))


def test_registry_diagnosis_never_summarizes_or_writes_summary(monkeypatch):
    from agent_core import llm_summary

    raw = {
        "_status": "ok",
        "error": None,
        "diag_report": {"root": "n"},
        "diag_report_text": "raw report",
    }
    _stub(monkeypatch, raw)

    def unexpected(*args):
        pytest.fail("Registry execution must not summarize or write summary cache")

    monkeypatch.setattr(llm_summary, "summarize_report", unexpected)
    monkeypatch.setattr(tools, "cache_diag_final_text", unexpected)

    run = run_diagnosis("job-x")
    assert run.legacy == {key: value for key, value in raw.items() if key != "_status"}
    assert "final_text" not in run.legacy

    result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x"})
    assert result.data.diag_report == raw["diag_report"]
    assert result.data.final_text is None


def test_adapter_classifies_internal_states(monkeypatch):
    cases = [
        ("busy", "error", "INTERNAL_ERROR", 500),
        ("not_found", "error", "INTERNAL_ERROR", 500),
        ("timeout", "error", "INTERNAL_ERROR", 500),
        ("failed", "error", "INTERNAL_ERROR", 500),
        ("partial", "partial", None, 200),
        ("ok", "ok", None, 200),
    ]
    for internal, status, code, http_status in cases:
        _stub(monkeypatch, {"error": None if status != "error" else f"{internal} reason", "_status": internal})
        result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x"})
        assert result.status == status
        assert [issue.code for issue in result.errors] == ([code] if code else [])
        assert _status_for(result) == http_status


def test_cached_result_without_internal_state_is_classified(monkeypatch):
    _stub(
        monkeypatch,
        {
            "diag_report": {"root": "n"},
            "diag_report_text": "report",
            "incomplete_nodes": ["worker-b"],
            "empty_nodes": [],
            "error": None,
        },
    )
    result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x"})
    assert result.status == "partial"


def test_internal_status_never_exposed(monkeypatch):
    _stub(monkeypatch, {"diag_report": {}, "error": None, "_status": "ok"})
    result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x"})
    dumped = result.data.model_dump(mode="json")
    assert "_status" not in str(dumped)
    assert "_status" not in result.data.diag_report


def test_unexpected_exception_becomes_diagnosis_failure(monkeypatch):
    def _boom(**_kwargs):
        raise RuntimeError("k8s unavailable")

    monkeypatch.setattr(tools, "diagnose_cached", _boom)
    result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x"})
    assert result.status == "error"
    assert result.errors[0].code == "INTERNAL_ERROR"
    assert "k8s unavailable" in result.errors[0].message
    assert _status_for(result) == 500
    assert result.data.error == result.errors[0].message
    assert result.data.pods == []
    assert result.data.diag_report == {}


def test_registry_execute_returns_diagnosis_result(monkeypatch):
    _stub(monkeypatch, {"diag_report": {"root": "n"}, "error": None, "_status": "ok"})
    result = TOOL_REGISTRY.execute(
        "diagnose", {"job": "job-x", "namespace": "default", "refresh": False}, ToolInvocation()
    )
    assert isinstance(result, DiagnosisResult)
    assert result.status == "ok"
    assert result.tool == "diagnose"
    assert result.data.diag_report.get("root") == "n"  # unknown report keys are preserved
    assert "task" not in result.data.model_dump()
    assert result.errors == []


def test_diagnosis_is_not_retried(monkeypatch):
    calls = []

    def fail(**kwargs):
        calls.append(kwargs)
        raise TimeoutError("collector did not answer")

    monkeypatch.setattr(tools, "diagnose_cached", fail)
    result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x"}, ToolInvocation())

    assert len(calls) == 1
    assert result.status == "error"
    assert result.errors[0].code == "INTERNAL_ERROR"


def test_registry_execute_maps_busy_to_error(monkeypatch):
    _stub(monkeypatch, {"error": "job=job-x is being diagnosed", "pods": [], "_status": "busy"})
    result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x"}, ToolInvocation())
    assert result.status == "error"
    assert result.errors[0].code == "INTERNAL_ERROR"
    assert result.data.error


def test_flat_result_preserves_report_and_collection(monkeypatch):
    raw = {
        "_status": "ok",
        "pods": [{"pod_name": "train-0", "node": "worker-a", "host_ip": "10.0.0.1", "phase": "Running"}],
        "collected": [{"node": "worker-a", "host_ip": "10.0.0.1", "ok": True, "error": None}],
        "incomplete_nodes": [],
        "empty_nodes": [],
        "diag_input_dir": "/in",
        "diag_output_dir": "/out",
        "diag_report_text": "report text",
        "diag_report": {
            "Version": "26.3.0",
            "Build_Time": "2026-10-29",
            "Task_id": "task-1",
            "Root_Cluster": {
                "analyze_success": True,
                "fault_description": {"code": 109, "string": "AICPU timeout"},
                "root_cause_device": ["10.0.0.1 device-11"],
                "remote_link": "device-3 -> device-11",
                "detect_workers_devices": {"10.0.0.1": ["11"]},
            },
            "Knowledge_Graph": {"analyze_success": False, "note": "no event"},
        },
    }
    _stub(monkeypatch, raw)
    result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x", "namespace": "training"})
    data = result.data.model_dump(mode="json")
    assert data["diag_report"] == raw["diag_report"]
    assert data["diag_report_text"] == "report text"
    assert "final_text" not in data


def test_diagnosis_query_rejects_extra_fields():
    from agent_core.tools.schemas import DiagnosisQuery
    from pydantic import ValidationError

    with pytest.raises(ValidationError):
        DiagnosisQuery.model_validate({"job": "job-x", "unexpected": 1})
    with pytest.raises(ValidationError):
        DiagnosisQuery.model_validate({"job": "job-x", "summarize": True})


def test_handler_returns_legacy_fields_directly(monkeypatch):
    raw = {
        "_status": "ok",
        "error": None,
        "pods": [],
        "diag_report": {"Root_Cluster": {"analyze_success": True}},
        "final_text": "legacy presentation",
    }
    _stub(monkeypatch, raw)
    context = DiagnosisContext(request_id="r1")
    result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x"}, context)
    assert result.data.final_text.startswith("legacy presentation")
    assert "--json" not in result.data.final_text
    assert "legacy_result" not in result.data.model_dump()


def test_adapter_calls_cached_chain_once_and_preserves_raw_result(monkeypatch):
    raw = {
        "_status": "ok",
        "error": None,
        "diag_report": {"root": "n"},
        "cached": True,
        "final_text": "cached summary",
        "extra": {"future": [1, None]},
    }
    calls = []

    def cached(**kwargs):
        calls.append(kwargs)
        return raw

    monkeypatch.setattr(tools, "diagnose_cached", cached)
    result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x", "namespace": "training", "refresh": True})
    assert calls == [{"job": "job-x", "namespace": "training", "refresh": True}]
    assert result.data.cached is True
    assert result.data.final_text.startswith("cached summary")
    assert "--json" not in result.data.final_text
    assert result.data.diag_report.get("root") == "n"
    assert raw["_status"] == "ok"  # removing internal metadata must not mutate the cached object
    assert raw["extra"] == {"future": [1, None]}
    assert raw["final_text"] == "cached summary"
    assert result.data.model_dump()["extra"] == raw["extra"]
    assert "final_text" in result.data.model_dump()


@pytest.mark.parametrize(
    "raw,status,code",
    [
        ({}, "ok", None),
        ({"error": None, "final_text": "old cached text"}, "ok", None),
        ({"empty_nodes": ["node-a"]}, "partial", None),
        ({"error": "diagnosis failed"}, "error", "INTERNAL_ERROR"),
        ({"_status": "unknown", "error": "unrecognized result"}, "error", "INTERNAL_ERROR"),
    ],
)
def test_legacy_cache_and_unknown_state_keep_their_contract(monkeypatch, raw, status, code):
    _stub(monkeypatch, raw)
    result = TOOL_REGISTRY.execute("diagnose", {"job": "job-x"})
    assert result.status == status
    assert [issue.code for issue in result.errors] == ([code] if code else [])
    assert result.data.error == raw.get("error")
    assert result.data.empty_nodes == raw.get("empty_nodes")
    assert result.data.final_text == raw.get("final_text")
