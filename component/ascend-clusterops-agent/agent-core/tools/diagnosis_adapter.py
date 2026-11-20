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

"""Diagnosis adapter: reuse the existing chain and attach internal execution metadata."""

from __future__ import annotations

import logging
from dataclasses import dataclass
from typing import Any, Dict

from agent_core.tools import diagnosis as tools
from agent_core.query_schemas import QueryIssue
from agent_core.tools.schemas import DiagnosisData, DiagnosisQuery, DiagnosisResult

logger = logging.getLogger(__name__)

# Internal state -> (result status, error code, fallback message).
# Diagnosis timeouts, concurrent runs and internal exceptions all map to one execution
# error; the HTTP boundary turns it into 500.
_OUTCOMES = {
    "busy": ("error", "INTERNAL_ERROR", "the same task is already being diagnosed"),
    "not_found": ("error", "INTERNAL_ERROR", "training/inference task not found"),
    "timeout": ("error", "INTERNAL_ERROR", "diagnosis timed out"),
    "failed": ("error", "INTERNAL_ERROR", "diagnosis failed"),
    "partial": ("partial", None, None),
    "ok": ("ok", None, None),
}


@dataclass
class DiagnosisContext:
    request_id: str


def _classify(legacy: dict) -> str:
    """Infer the execution state of an old cache entry without internal metadata."""
    if legacy.get("error"):
        return "failed"
    if legacy.get("incomplete_nodes") or legacy.get("empty_nodes"):
        return "partial"
    return "ok"


@dataclass(frozen=True)
class DiagnosisRun:
    """One run of the diagnosis chain: the legacy flat payload plus its outcome."""

    legacy: Dict[str, Any]
    outcome: str


def diagnosis_failure(exc: Exception) -> Dict[str, Any]:
    """The readable legacy failure body shared by /diag and the diagnose tool."""
    return {
        "error": f"Diagnosis error: an exception occurred during diagnosis, please check the agent-core logs ({exc})",
        "pods": [],
        "diag_report": {},
    }


def run_diagnosis(
    job: str,
    namespace: str = "default",
    refresh: bool = False,
) -> DiagnosisRun:
    """Run the deterministic diagnosis chain and return its raw result.

    No LLM summary or presentation is applied here: the business result stays raw and
    is wrapped by the tool adapter and presented at the HTTP boundary.
    """
    try:
        raw = tools.diagnose_cached(job=job, namespace=namespace, refresh=refresh)
    except Exception as exc:  # noqa: BLE001  # preserve the readable diagnosis failure at the tool boundary
        logger.exception("diagnosis crashed: job=%s ns=%s", job, namespace)
        return DiagnosisRun(legacy=diagnosis_failure(exc), outcome="failed")

    legacy = {key: value for key, value in raw.items() if key != "_status"}
    # The business result is exactly what the cache returned; cached final_text is
    # preserved, but a new summary is only computed at the HTTP boundary.
    outcome = raw.get("_status") or _classify(legacy)
    return DiagnosisRun(legacy=legacy, outcome=outcome)


def diagnose_handler(request: DiagnosisQuery, context: DiagnosisContext) -> DiagnosisResult:
    """Wrap the raw diagnosis result without LLM summary or cache write-back."""
    run = run_diagnosis(request.job, request.namespace, request.refresh)
    status, code, fallback_message = _OUTCOMES.get(run.outcome, _OUTCOMES["failed"])
    errors = []
    if code:
        errors.append(QueryIssue(code=code, message=str(run.legacy.get("error") or fallback_message)))
    return DiagnosisResult(
        request_id=context.request_id,
        tool="diagnose",
        status=status,
        data=DiagnosisData.model_validate(run.legacy),
        errors=errors,
    )
