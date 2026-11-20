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

"""HTTP diagnosis presentation and optional LLM summary with cache write-back."""

from __future__ import annotations

import logging
from typing import Any, Dict

from agent_core import llm_summary

logger = logging.getLogger(__name__)

# Hint appended after the report output: the kubectl plugin prints full JSON with --json
JSON_HINT = "\n\nTip: to output the full diagnosis JSON, add --json after the command"


def present_diagnosis(result: Dict[str, Any]) -> Dict[str, Any]:
    """Format a diagnosis result without invoking the LLM or writing the cache."""
    from agent_core.tools import diagnosis as tools

    presented = result if result is not None else {}
    if presented.get("error"):
        presented["final_text"] = presented["error"]
        return presented
    final_text = presented.get("final_text")
    if not final_text:
        final_text = presented.get("diag_report_text") or ""
        notes = tools.nodes_note(presented.get("incomplete_nodes") or [], presented.get("empty_nodes") or [])
        if notes and notes not in final_text:
            final_text = (final_text + "\n\n" + notes).strip()
    if not final_text.endswith(JSON_HINT):
        final_text += JSON_HINT
    presented["final_text"] = final_text
    return presented


def summarize_diagnosis(result: Dict[str, Any], job: str, namespace: str) -> Dict[str, Any]:
    """Summarize a raw result in place, reusing cached text and preserving failures."""
    # Load diagnosis helpers only when summary processing is requested.
    from agent_core.tools import diagnosis as tools

    presented = result if result is not None else {}
    if presented.get("error") or presented.get("final_text"):
        return presented

    report_text = presented.get("diag_report_text") or ""
    notes_text = tools.nodes_note(presented.get("incomplete_nodes") or [], presented.get("empty_nodes") or [])
    if notes_text and notes_text not in report_text:
        report_text = (report_text + "\n\n" + notes_text).strip()

    llm_error = None
    try:
        text = llm_summary.summarize_report(presented.get("diag_report") or {})
        final_text = text or report_text
    except llm_summary.LLMSummaryError as e:  # mis-configured/unreachable LLM -> keep the deterministic report
        llm_error = str(e)
        final_text = f"{e}\n\n{report_text}"
    except Exception as e:  # noqa: BLE001  # unexpected summarizer bug -> keep the deterministic report, never 500
        logger.exception("LLM report summarization crashed: %s", e)
        llm_error = (
            "Diagnosis error: an exception occurred during report summarization, please check the agent-core logs"
        )
        final_text = report_text
    if notes_text and notes_text not in final_text:
        final_text = (final_text + "\n\n" + notes_text).strip()

    # The hint is presentation-only; cache the report text without it so a later
    # cache hit appends it exactly once.
    presented["final_text"] = final_text
    if llm_error:
        presented["llm_error"] = llm_error
    tools.cache_diag_final_text(job, namespace, final_text)
    return presented
