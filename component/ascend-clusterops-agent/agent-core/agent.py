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

"""Optional LLM agent wrapper layer.

The main flow is the deterministic diagnose tool (see tools.py). This module wraps a
minimal react agent on top: the LLM receives user natural language -> extracts the job
name -> calls diagnose -> summarizes diag_report for the user; it also provides the
ability to summarize diag_report into a Chinese root-cause report (summarize_report).

When LLM_API_KEY is not set, build_agent() returns None and the main flow can still be
called directly via HTTP/CLI. Interactive UI extension = attach more read-only tools
(read_plog / read_diag_artifact / list_nodes) to the same agent without reworking the
main flow.
"""

from __future__ import annotations

import json
import logging
from typing import Optional

from langchain_core.tools import tool
from langgraph.prebuilt import create_react_agent

from agent_core import tools as T
from agent_core.llm import get_llm

logger = logging.getLogger(__name__)


class LLMSummaryError(Exception):
    """LLM report summarization failed; carries a user-facing reason. The deterministic diagnosis is unaffected."""


def _llm_error_hint(exc: Exception) -> str:
    """Map an LLM call failure to an actionable reason the user can understand and fix."""
    msg = str(exc) or exc.__class__.__name__
    if "UnsupportedProtocol" in msg or "missing an 'http://' or 'https://' protocol" in msg:
        return "诊断错误，LLM 配置有误：base_url 缺少 http:// 或 https:// 前缀；请用 kubectl clusterops --create-llm-config --base-url http(s)://<url> 重新配置"
    if "APIConnectionError" in msg or "Connection error" in msg or "timed out" in msg:
        return "诊断错误，LLM 配置有误：无法连接 LLM 服务或请求超时；请检查网络与 LLM 服务是否可达，或用 kubectl clusterops --clear-llm-config 关闭 LLM"
    if "authentication" in msg.lower() or "401" in msg:
        return "诊断错误，LLM 配置有误：api-key 无效（认证失败 401）；请用 kubectl clusterops --create-llm-config 重新配置正确的 api-key"
    if "permission" in msg.lower() or "403" in msg:
        return "诊断错误，LLM 配置有误：api-key 权限不足（403）；请检查 api-key 权限，或用 kubectl clusterops --clear-llm-config 关闭 LLM"
    return f"诊断错误，LLM 调用失败：{msg[:200]}；可用 kubectl clusterops --clear-llm-config 关闭 LLM，回退确定性报告"


@tool
def diagnose(job: str, namespace: str = "default") -> str:
    """Diagnose a fault of an NPU training job. Pass the job name (jobname, task CR name)
    and the namespace; it looks up the central relation table by jobname to get all pods
    and runs the diagnosis. Returns the ascend-fd diagnosis result JSON.
    """
    r = T.diagnose(job=job, namespace=namespace)
    return json.dumps(r, ensure_ascii=False, default=str)


def build_agent():
    """Return the react agent; return None when no LLM is configured."""
    llm_obj = get_llm()
    if llm_obj is None:
        return None
    # Attach the diagnose tool; add more read-only tools here for interactive UI extension.
    return create_react_agent(llm_obj, [diagnose])


def summarize_report(diag_report: dict) -> Optional[str]:
    """Summarize diag_report.json into a Chinese root-cause report (used by /diag path).

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


def format_diag_report(diag_report: dict) -> str:
    """Format diag_report.json into human-readable text (fallback when no LLM configured).

    Field organization mirrors ascend-fd's PrintWrapper: root cause (Root_Cluster)
    first, then each fault in Knowledge_Graph.fault with cause/description/suggestion
    and a key log line per fault.
    """
    lines: list[str] = []
    rc = diag_report.get("Root_Cluster") or {}
    kg = diag_report.get("Knowledge_Graph") or {}
    fd = rc.get("fault_description") or {}
    if fd.get("string"):
        lines.append(f"[Root Cause] {fd['string']}")
    devices = rc.get("root_cause_device") or []
    if devices:
        lines.append(f"Root cause device: {', '.join(str(d) for d in devices)}")
    note = rc.get("note") or kg.get("note")
    if note:
        lines.append(f"Note: {note}")
    for i, fault in enumerate(kg.get("fault") or [], 1):
        lines.append("")
        lines.append(f"[{i}] Fault: {fault.get('code')}")
        cls = " ".join(str(fault.get(k)) for k in ("class", "component", "module") if fault.get(k))
        if cls:
            lines.append(f"    Type: {cls}")
        if fault.get("cause_zh"):
            lines.append(f"    Name: {fault['cause_zh']}")
        if fault.get("description_zh"):
            lines.append(f"    Desc: {fault['description_zh']}")
        for s in fault.get("suggestion_zh") or []:
            lines.append(f"    Suggestion: {s}")
        for events in (fault.get("event_attr") or {}).values():
            for ev in events or []:
                if ev.get("key_info"):
                    lines.append(f"    Log: {ev['key_info'][:200]}")
                    break
            break  # one key log line per fault to keep output compact
    return "\n".join(lines) or "(no fault detected)"
