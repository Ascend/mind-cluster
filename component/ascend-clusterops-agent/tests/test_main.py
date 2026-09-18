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

from fastapi.testclient import TestClient

from agent_core import main, tools


def test_diag_endpoint_with_summary(monkeypatch):
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda job, namespace="default", refresh=False: {"diag_report": {"root": "n"}, "error": None, "pods": []},
    )
    monkeypatch.setattr(main, "summarize_report", lambda r: "中文报告")
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x", "namespace": "default"})
    assert resp.status_code == 200
    body = resp.json()
    assert body["final_text"].startswith("中文报告")
    assert "--json" in body["final_text"]  # hint to print the full JSON response
    assert body["diag_report"] == {"root": "n"}


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
    monkeypatch.setattr(main, "summarize_report", lambda r: None)
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
    monkeypatch.setattr(main, "summarize_report", lambda r: None)
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
    monkeypatch.setattr(main, "summarize_report", lambda r: None)
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x"})
    assert resp.status_code == 200
    assert resp.json()["error"] == "no pods found"
    assert resp.json()["final_text"] == "no pods found"  # diag failed -> final_text is the reason directly


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
        raise main.LLMSummaryError("LLM base_url 配置无效 (缺少 http:// 或 https:// 前缀); 请重新配置")

    monkeypatch.setattr(main, "summarize_report", _boom)
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

    monkeypatch.setattr(main, "summarize_report", _boom)
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
    monkeypatch.setattr(main, "summarize_report", lambda r: called.append(1) or "should not be used")
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
    monkeypatch.setattr(main, "summarize_report", lambda r: "fresh LLM summary")
    written = {}
    monkeypatch.setattr(tools, "cache_diag_final_text", lambda job, ns, text: written.update(job=job, ns=ns, text=text))
    client = TestClient(main.app)
    resp = client.post("/diag", json={"job": "job-x", "namespace": "train"})
    assert resp.status_code == 200
    assert resp.json()["final_text"].startswith("fresh LLM summary")
    assert "--json" in resp.json()["final_text"]  # the hint is cached together with the summary
    assert written["job"] == "job-x" and written["ns"] == "train"
    assert written["text"].startswith("fresh LLM summary")
