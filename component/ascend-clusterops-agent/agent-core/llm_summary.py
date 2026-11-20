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

"""Optional LLM summary of the deterministic diagnosis report.

Kept separate from ``agent.py`` so the diagnosis tool handler and the legacy
``/diag`` presenter can share it without an import cycle through the tool
registry.
"""

from __future__ import annotations

import json
import logging
from typing import Optional

from agent_core.llm import get_llm

logger = logging.getLogger(__name__)


class LLMSummaryError(Exception):
    """LLM report summarization failed; carries a user-facing reason. The deterministic diagnosis is unaffected."""


def _llm_error_hint(exc: Exception) -> str:
    """Map an LLM call failure to an actionable reason the user can understand and fix."""
    msg = str(exc) or exc.__class__.__name__
    if "UnsupportedProtocol" in msg or "missing an 'http://' or 'https://' protocol" in msg:
        return "Diagnosis error, LLM mis-configured: base_url is missing an 'http://' or 'https://' prefix; reconfigure with: kubectl clusterops --create-llm-config --base-url http(s)://<url>"
    if "APIConnectionError" in msg or "Connection error" in msg or "timed out" in msg:
        return "Diagnosis error, LLM mis-configured: cannot connect to the LLM service or the request timed out; check network reachability to the LLM service, or disable LLM with: kubectl clusterops --clear-llm-config"
    if "authentication" in msg.lower() or "401" in msg:
        return "Diagnosis error, LLM mis-configured: invalid api-key (authentication failed 401); reconfigure the correct api-key with: kubectl clusterops --create-llm-config"
    if "permission" in msg.lower() or "403" in msg:
        return "Diagnosis error, LLM mis-configured: insufficient api-key permissions (403); check the api-key permissions, or disable LLM with: kubectl clusterops --clear-llm-config"
    return f"Diagnosis error, LLM call failed: {msg[:200]}; disable LLM with: kubectl clusterops --clear-llm-config to fall back to the deterministic report"


def summarize_report(diag_report: dict) -> Optional[str]:
    """Summarize diag_report.json into a Chinese root-cause report (used by /diag and the diagnose tool).

    Returns None when no LLM is configured (caller falls back to raw formatting).
    Raises LLMSummaryError when the LLM call fails (mis-configured base_url / bad api-key /
    network / any unexpected summarizer error), so callers only ever need to catch
    LLMSummaryError; the deterministic diagnosis is never affected.
    """
    try:
        llm_obj = get_llm()
        if llm_obj is None:
            return None
        prompt = (
            "你是 Ascend NPU 训练故障诊断助手。根据下面的 ascend-fd 诊断结果，"
            "用中文给运维人员写一份简明根因报告：故障事件、根因节点、建议处置。"
            "先给一句话结论。诊断结果 JSON:\n" + json.dumps(diag_report or {}, ensure_ascii=False)
        )
        return llm_obj.invoke(prompt).content
    except LLMSummaryError:
        raise
    except Exception as e:  # noqa: BLE001  # any LLM failure must never fail the deterministic diagnosis
        logger.warning("LLM report summarization failed (using deterministic report): %s", e)
        raise LLMSummaryError(_llm_error_hint(e)) from e
