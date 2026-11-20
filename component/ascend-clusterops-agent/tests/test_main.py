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

"""agent.main HTTP entry tests (FastAPI TestClient, mocks diagnose / summarize)."""

from __future__ import annotations

import pytest
from agent_core import llm_summary, main
from agent_core.tools import TOOL_REGISTRY
from agent_core.tools import diagnosis as tools
from fastapi.testclient import TestClient


def test_diag_malformed_json_keeps_the_legacy_framework_response():
    client = TestClient(main.app)
    response = client.post("/diag", content='{"job": ', headers={"content-type": "application/json"})

    assert response.status_code == 422
    assert "detail" in response.json()
    assert "errors" not in response.json()


def test_diag_endpoint_with_summary(monkeypatch):
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda job, namespace="default", refresh=False: {"diag_report": {"root": "n"}, "error": None, "pods": []},
    )
    monkeypatch.setattr(llm_summary, "summarize_report", lambda r: "中文报告")
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x", "namespace": "default"})
    assert resp.status_code == 200
    body = resp.json()
    assert body["final_text"].startswith("中文报告")
    assert "--json" in body["final_text"]  # hint to print the full JSON response
    assert body["diag_report"] == {"root": "n"}


def test_tools_diagnose_data_matches_diag_body(monkeypatch):
    """The diagnose tool's data is the old flat /diag body."""
    raw = {
        "diag_report": {"Root_Cluster": {"analyze_success": True}},
        "diag_report_text": "raw report",
        "pods": [{"pod_name": "worker-0"}],
        "cached": False,
        "cached_at": "2026-10-31T00:00:00Z",
        "error": None,
    }
    monkeypatch.setattr(tools, "diagnose_cached", lambda **kwargs: dict(raw))
    monkeypatch.setattr(llm_summary, "summarize_report", lambda report: "summary")
    monkeypatch.setattr(tools, "cache_diag_final_text", lambda *args: None)
    client = TestClient(main.app)

    legacy_body = client.post("/diag", json={"job": "job-x"}).json()
    envelope = client.post(
        "/tools/diagnose",
        json={"job": "job-x"},
    ).json()
    assert envelope["status"] == "ok"
    assert envelope["data"] == legacy_body
    assert envelope["data"]["diag_report"] == raw["diag_report"]
    assert envelope["data"]["pods"] == raw["pods"]
    assert "task" not in envelope["data"]

    standard = client.post("/tools/diagnose", json={"job": "job-x"}).json()["data"]
    assert standard == legacy_body
    raw_data = client.post("/tools/diagnose", json={"job": "job-x"}).json()["data"]
    assert raw_data["diag_report"] == raw["diag_report"]
    assert raw_data["diag_report_text"] == raw["diag_report_text"]
    assert raw_data["final_text"]


def test_diag_endpoint_no_llm_formats_report(monkeypatch):
    report_text = "Mindcluster Fault-Diag Report\n[Root Cause] no plog found\n| F1 | Suggestion: fix it |"
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda job, namespace="default", refresh=False: {
            "diag_report": {"Root_Cluster": {}},
            "diag_report_text": report_text,
            "error": None,
            "pods": [],
        },
    )
    monkeypatch.setattr(llm_summary, "summarize_report", lambda r: None)
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x"})
    assert resp.status_code == 200
    text = resp.json()["final_text"]
    assert "no plog found" in text
    assert "F1" in text
    assert "Suggestion: fix it" in text


def test_diag_endpoint_uses_diag_report_text(monkeypatch):
    # diag stdout (pretty table) is preferred over the JSON reformat when present
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda job, namespace="default", refresh=False: {
            "diag_report": {"Root_Cluster": {}},
            "diag_report_text": "Mindcluster Fault-Diag Report\n| Fault code |",
            "error": None,
            "pods": [],
        },
    )
    monkeypatch.setattr(llm_summary, "summarize_report", lambda r: None)
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x"})
    assert resp.status_code == 200
    assert "Mindcluster Fault-Diag Report" in resp.json()["final_text"]


def test_diag_endpoint_carries_error(monkeypatch):
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda job, namespace="default", refresh=False: {"error": "no pods found", "pods": []},
    )
    monkeypatch.setattr(llm_summary, "summarize_report", lambda r: None)
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x"})
    assert resp.status_code == 200
    assert resp.json()["error"] == "no pods found"
    assert resp.json()["final_text"] == "no pods found"  # diag failed -> final_text is the reason directly


def test_diag_unexpected_failure_returns_the_original_flat_body(monkeypatch):
    # Original /diag path: an exception from diagnose_cached returns the readable body
    # directly (no presentation step), exactly as before the tooling change.
    def _boom(**_kwargs):
        raise RuntimeError("k8s api unavailable")

    monkeypatch.setattr(tools, "diagnose_cached", _boom)
    resp = TestClient(main.app).post("/diag", json={"job": "job-x"})
    assert resp.status_code == 200
    body = resp.json()
    assert "k8s api unavailable" in body["error"]
    assert body["pods"] == []
    assert body["diag_report"] == {}
    assert "final_text" not in body


def test_diag_endpoint_captures_llm_error(monkeypatch):
    # LLM mis-configured -> no 500: returns the llm_error reason + deterministic report text
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda job, namespace="default", refresh=False: {
            "diag_report": {"Root_Cluster": {}},
            "diag_report_text": "Mindcluster Fault-Diag Report\n[Root Cause] no plog found",
            "error": None,
            "pods": [],
        },
    )

    def _boom(_r):
        raise llm_summary.LLMSummaryError("LLM base_url 配置无效 (缺少 http:// 或 https:// 前缀); 请重新配置")

    monkeypatch.setattr(llm_summary, "summarize_report", _boom)
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x"})
    assert resp.status_code == 200
    body = resp.json()
    assert body["llm_error"].startswith("LLM base_url 配置无效")
    assert "LLM base_url 配置无效" in body["final_text"]
    assert "no plog found" in body["final_text"]  # the deterministic report is still kept


def test_diag_endpoint_llm_unexpected_error_falls_back(monkeypatch):
    # summarize_report raises an unexpected (non-LLMSummaryError) exception -> still 200, fall back to the deterministic report
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda job, namespace="default", refresh=False: {
            "diag_report": {"Root_Cluster": {}},
            "diag_report_text": "Mindcluster Fault-Diag Report\n[Root Cause] no plog found",
            "error": None,
            "pods": [],
        },
    )

    def _boom(_r):
        raise RuntimeError("unexpected summarizer bug")

    monkeypatch.setattr(llm_summary, "summarize_report", _boom)
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x"})
    assert resp.status_code == 200
    body = resp.json()
    assert "report summarization" in body["llm_error"]
    assert "no plog found" in body["final_text"]


def test_diag_endpoint_reuses_cached_summary(monkeypatch):
    # cache hit with final_text (LLM summary cached) -> reuse it directly, no LLM call
    report = {"Root_Cluster": {"fault_description": {"string": "no plog found"}}}
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda job, namespace="default", refresh=False: {
            "diag_report": report,
            "error": None,
            "pods": [],
            "final_text": "cached LLM summary",
            "cached": True,
        },
    )
    called = []
    monkeypatch.setattr(llm_summary, "summarize_report", lambda r: called.append(1) or "should not be used")
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x"})
    assert resp.status_code == 200
    assert resp.json()["final_text"].startswith("cached LLM summary")
    assert "--json" in resp.json()["final_text"]  # the --json hint is appended on cache hits too
    assert not called  # no LLM call


def test_diag_endpoint_writes_summary_to_cache(monkeypatch):
    # fresh diagnosis -> compute the summary and write it back to the cache for later hits
    report = {"Root_Cluster": {"fault_description": {"string": "no plog found"}}}
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda job, namespace="default", refresh=False: {"diag_report": report, "error": None, "pods": []},
    )
    monkeypatch.setattr(llm_summary, "summarize_report", lambda r: "fresh LLM summary")
    written = {}
    monkeypatch.setattr(tools, "cache_diag_final_text", lambda job, ns, text: written.update(job=job, ns=ns, text=text))
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x", "namespace": "train"})
    assert resp.status_code == 200
    assert resp.json()["final_text"].startswith("fresh LLM summary")
    assert "--json" in resp.json()["final_text"]  # the hint is cached together with the summary
    assert written["job"] == "job-x" and written["ns"] == "train"
    assert written["text"].startswith("fresh LLM summary")


def test_diag_shares_the_service_but_not_the_tool_registry(monkeypatch):
    # /diag reuses run_diagnosis directly, so the registry (and its policies or
    # strict tool argument model) must never see a legacy kubectl ascend_diag call.
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda job, namespace="default", refresh=False: {"diag_report": {"root": "n"}, "error": None, "pods": []},
    )
    monkeypatch.setattr(llm_summary, "summarize_report", lambda r: None)
    monkeypatch.setattr(
        TOOL_REGISTRY, "execute", lambda *args, **kwargs: pytest.fail("/diag must not call the tool registry")
    )
    response = TestClient(main.app).post("/diag", json={"job": "job-x"})
    assert response.status_code == 200
    assert response.json()["diag_report"] == {"root": "n"}


def test_repeated_diag_reuses_the_summary_like_before(monkeypatch):
    # The old /diag wrote the presentation back into the diagnose_cached result, so a
    # repeat call within RECENT_TTL reused it instead of calling the model again.
    calls = []
    cached_result = {"diag_report": {"root": "n"}, "error": None, "pods": [], "diag_report_text": "report"}
    monkeypatch.setattr(tools, "diagnose_cached", lambda **kwargs: cached_result)
    monkeypatch.setattr(llm_summary, "summarize_report", lambda r: calls.append(1) or "总结")
    monkeypatch.setattr(tools, "cache_diag_final_text", lambda *a: None)
    client = TestClient(main.app)

    first = client.post("/diag", json={"job": "job-x"}).json()
    second = client.post("/diag", json={"job": "job-x"}).json()
    assert first["final_text"].startswith("总结")
    assert second["final_text"].startswith("总结")
    assert len(calls) == 1  # the second call reused the cached summary


@pytest.mark.parametrize("outcome", ["ok", "partial", "busy", "not_found", "timeout", "failed"])
def test_diagnosis_data_matches_ascend_diag_across_http_and_tool(monkeypatch, outcome):
    """Compare full legacy bodies, including omitted/null/extension/nested original fields."""
    raw = {
        "_status": outcome,
        "error": None if outcome in {"ok", "partial"} else "diagnosis error",
        "pods": [{"pod_name": "worker-0", "deleted_at": 123.5, "extension": {"x": None}}],
        "collected": [{"node": "worker-a", "ok": True, "files": ["plog"]}],
        "diag_report": {"Version": "26.3.0", "Root_Cluster": {"fault_description": {"code": 102}}},
        "diag_report_text": "deterministic report",
        "incomplete_nodes": ["worker-b"] if outcome == "partial" else [],
        "cached": True,
        "cached_note": None,
        "cached_at": "2026-11-02T00:00:00Z",
        "extension": {"future": [None, 0, False]},
    }
    monkeypatch.setattr(tools, "diagnose_cached", lambda **kwargs: dict(raw))
    monkeypatch.setattr(llm_summary, "summarize_report", lambda report: None)
    monkeypatch.setattr(tools, "cache_diag_final_text", lambda *args: None)
    client = TestClient(main.app)
    legacy = client.post("/diag", json={"job": "train-a"}).json()
    body = client.post("/tools/diagnose", json={"job": "train-a"}).json()
    assert body["data"] == legacy
    assert "_status" not in legacy
    assert "empty_nodes" not in legacy  # absent fields must not be synthesized
    assert legacy["diag_report"] == raw["diag_report"]
    assert legacy["pods"] == raw["pods"]


@pytest.mark.parametrize("mode", ["model", "no-model", "failed", "crashed", "cached", "error"])
def test_tool_http_summarizes_once_and_matches_legacy(monkeypatch, mode):
    raw = {
        "diag_report": {"root": "node-a"},
        "diag_report_text": "raw report",
        "incomplete_nodes": ["node-b"],
    }
    if mode == "cached":
        raw["final_text"] = "cached summary"
    if mode == "error":
        raw["error"] = "diagnosis failed"
    calls, writes = [], []

    def summarize(report):
        calls.append(report)
        if mode == "failed":
            raise llm_summary.LLMSummaryError("model unavailable")
        if mode == "crashed":
            raise RuntimeError("unexpected model failure")
        return None if mode == "no-model" else "model summary"

    monkeypatch.setattr(tools, "diagnose_cached", lambda **kwargs: dict(raw))
    monkeypatch.setattr(llm_summary, "summarize_report", summarize)
    monkeypatch.setattr(tools, "cache_diag_final_text", lambda *args: writes.append(args))
    client = TestClient(main.app)
    response = client.post("/tools/diagnose", json={"job": "train-a", "namespace": "training"})
    assert response.status_code == (500 if mode == "error" else 200)
    data = response.json()["data"]
    expected_count = 0 if mode in {"cached", "error"} else 1
    assert len(calls) == len(writes) == expected_count
    if writes:
        assert writes[0][:2] == ("train-a", "training")
        assert "--json" not in writes[0][2]
    legacy = client.post("/diag", json={"job": "train-a", "namespace": "training"}).json()
    assert data == legacy
    assert "final_text" not in raw or mode == "cached"
