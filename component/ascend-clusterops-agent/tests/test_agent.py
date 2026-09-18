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

"""agent.agent unit tests: optional LLM wrapper layer (summarize / build_agent / tool)."""

from __future__ import annotations

import json
from types import SimpleNamespace

import pytest

import agent_core.agent as A


def test_summarize_report_no_llm(monkeypatch):
    # no LLM_API_KEY in the test env -> get_llm() returns None -> summarize returns None
    monkeypatch.setattr(A, "get_llm", lambda: None)
    assert A.summarize_report({"root": "n"}) is None


def test_summarize_report_with_llm(monkeypatch):
    fake = SimpleNamespace(invoke=lambda prompt: SimpleNamespace(content="根因报告: n"))
    monkeypatch.setattr(A, "get_llm", lambda: fake)
    out = A.summarize_report({"root": "n"})
    assert out == "根因报告: n"


def test_summarize_report_llm_error_raises_hint(monkeypatch):
    # LLM call fails (e.g. base_url missing a protocol) -> raises LLMSummaryError with an actionable reason, no raw stack
    def _boom(_prompt):
        raise RuntimeError("Request URL is missing an 'http://' or 'https://' protocol.")

    monkeypatch.setattr(A, "get_llm", lambda: SimpleNamespace(invoke=_boom))
    with pytest.raises(A.LLMSummaryError) as ei:
        A.summarize_report({"root": "n"})
    msg = str(ei.value)
    assert "Diagnosis error" in msg
    assert "base_url" in msg
    assert "UnsupportedProtocol" not in msg


def test_summarize_report_get_llm_error_wrapped(monkeypatch):
    # get_llm() itself raises (non-LLMSummaryError) -> also wrapped as LLMSummaryError, never leaks
    def _boom():
        raise RuntimeError("llm module broken")

    monkeypatch.setattr(A, "get_llm", _boom)
    with pytest.raises(A.LLMSummaryError):
        A.summarize_report({"root": "n"})


def test_build_agent_no_llm(monkeypatch):
    monkeypatch.setattr(A, "get_llm", lambda: None)
    assert A.build_agent() is None


def test_diagnose_tool_invokes_underlying(monkeypatch):
    # the @tool-decorated diagnose must call agent.tools.diagnose
    monkeypatch.setattr(A.T, "diagnose", lambda job, namespace="default": {"diag_report": {"root": "n"}, "error": None})
    # LangChain tools are invoked via .invoke(dict)
    out = A.diagnose.invoke({"job": "job-x", "namespace": "default"})
    assert isinstance(out, str)
    obj = json.loads(out)
    assert obj["diag_report"] == {"root": "n"}
    assert obj["error"] is None
