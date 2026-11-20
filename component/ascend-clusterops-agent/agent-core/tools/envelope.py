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

"""Deterministic read-only query tools plus their shared result assembly.

Business-layer problems (source read failures, uncovered targets, parse issues) are
logged only and never reach the public envelope. ``build_result`` therefore derives
``status`` from one question: did this request answer every target it had to answer?
``error`` means the tool could not execute, ``ok`` means coverage is complete, and
``partial`` means execution succeeded but coverage is not.
"""

from __future__ import annotations

import logging
from typing import Any, Iterable, List, Optional, Type

from agent_core.datasources.query_context import QueryContext
from agent_core.query_schemas import QueryIssue, QueryResult

logger = logging.getLogger(__name__)

OK_SOURCE_STATUSES = ("ok", "empty", "absent_expected", "not_needed")
# Internal source status -> the log-only code; these are never returned in `errors`.
LOG_CODES = {
    "missing": "SOURCE_NOT_FOUND",
    "forbidden": "SOURCE_FORBIDDEN",
    "invalid": "SOURCE_INVALID",
    "timeout": "SOURCE_UNAVAILABLE",
    "unavailable": "SOURCE_UNAVAILABLE",
}


def source_failed(status: str) -> bool:
    return status not in OK_SOURCE_STATUSES and status != "disabled"


def log_source_issue(
    tool: str,
    family: str,
    status: str,
    error_code: Optional[str] = None,
    message: Optional[str] = None,
    source_id: Optional[str] = None,
    path: Optional[str] = None,
) -> None:
    """Record a business-layer source problem in the server log only."""
    if not source_failed(status):
        return
    logger.warning(
        "query source issue: tool=%s source=%s code=%s source_id=%s path=%s message=%s",
        tool,
        family,
        error_code or LOG_CODES.get(status, "SOURCE_UNAVAILABLE"),
        source_id,
        path,
        message or status,
    )


def log_node_listing_unavailable(tool: str, message: str, source_id: Optional[str] = None) -> None:
    """Log a Node-listing read failure; the ``nodes`` source family is log-only."""
    log_source_issue(tool, "nodes", "unavailable", "SOURCE_UNAVAILABLE", message, source_id=source_id)


def log_parse_issues(tool: str, issues: List[QueryIssue]) -> None:
    """Record parser-level issues in the server log only; the affected record is already isolated."""
    for issue in issues:
        logger.warning(
            "query parse issue: tool=%s code=%s source_id=%s path=%s message=%s",
            tool,
            issue.code,
            issue.source_id,
            issue.path,
            issue.message,
        )


def build_result(
    result_cls: Type[QueryResult],
    tool: str,
    ctx: QueryContext,
    data: Any,
    core_ok: bool = True,
    complete: bool = True,
    errors: Optional[Iterable[QueryIssue]] = None,
) -> QueryResult:
    if not core_ok:
        status = "error"
    elif complete:
        status = "ok"
    else:
        status = "partial"
    return result_cls(
        request_id=ctx.request_id,
        tool=tool,
        status=status,
        data=data if core_ok else None,
        errors=list(errors or []),
    )


def is_empty_scope(ctx: QueryContext) -> bool:
    """``node: []`` means an explicitly empty target scope, never a full-cluster read."""
    return ctx.requested_nodes is not None and len(ctx.requested_nodes) == 0
