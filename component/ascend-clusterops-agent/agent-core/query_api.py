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

"""HTTP boundary for discovery and unified execution.

``GET /tools`` returns the catalog and ``POST /tools/{name}`` executes the
registered tool. Every tool, including diagnose, returns its typed tool envelope
(QueryResult or DiagnosisResult); the legacy /diag projection lives in main.py.
Importing this module does not build an LLM client or touch the cluster.
"""

from __future__ import annotations

import logging
import time
import uuid
from typing import Any, Dict, Optional

from agent_core.datasources.query_context import QueryContext
from agent_core.diagnosis_presentation import present_diagnosis, summarize_diagnosis
from agent_core.query_schemas import QueryFailure
from agent_core.tools import (
    QUERY_POLICY,
    TOOL_REGISTRY,
    ToolDefinition,
    ToolInvocation,
)
from agent_core.tools.errors import build_error_envelope
from agent_core.tools.schemas import (
    DiagnosisData,
    DiagnosisQuery,
    ToolCatalog,
    tool_response_payload,
)
from fastapi import APIRouter, Body, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from fastapi.routing import APIRoute
from pydantic import ValidationError

logger = logging.getLogger(__name__)


def _request_id(invocation: Any) -> str:
    return getattr(invocation, "request_id", None) or uuid.uuid4().hex


def _error_response(
    *,
    tool: str,
    code: str,
    message: str,
    http_status: int,
    path: Optional[str] = None,
    invocation: Any = None,
    elapsed_ms: int = 0,
    definition: Optional[ToolDefinition] = None,
) -> JSONResponse:
    request_id = _request_id(invocation)
    envelope = build_error_envelope(
        tool=tool,
        code=code,
        message=message,
        request_id=request_id,
        elapsed_ms=elapsed_ms,
        path=path,
    )
    if definition is not None:
        # Every registered output model extends ToolEnvelope, so one construction covers
        # both the query envelope and the diagnosis envelope.
        payload: Any = definition.output_model.model_validate(envelope.model_dump(mode="python"))
    else:
        payload = envelope
    return JSONResponse(status_code=http_status, content=payload.model_dump(mode="json"))


def _validation_error_response(request: Request, exc: RequestValidationError) -> JSONResponse:
    """Turn FastAPI's default ``{"detail": ...}`` into the same tool envelope."""
    tool_name = str(request.path_params.get("tool_name") or "tools")
    details = exc.errors()
    first = details[0] if details else {}
    error_type = first.get("type")
    if error_type == "json_invalid":
        message = "request body is not valid JSON"
        path = None
    elif error_type == "dict_type":
        message = "request body must be a JSON object"
        path = None
    else:
        message = str(first.get("msg") or "invalid request")
        path = ".".join(str(part) for part in first.get("loc", ()) if part != "body") or None
    return _error_response(
        tool=tool_name,
        code="INVALID_ARGUMENT",
        message=message,
        http_status=422,
        path=path,
        definition=_lookup(tool_name),
    )


class ToolEnvelopeRoute(APIRoute):
    """Keep ``/tools/*`` parse failures inside the tool contract, never ``detail``."""

    def get_route_handler(self):
        original = super().get_route_handler()

        async def handler(request: Request):
            try:
                return await original(request)
            except RequestValidationError as exc:
                return _validation_error_response(request, exc)

        return handler


router = APIRouter(tags=["tools"], route_class=ToolEnvelopeRoute)


def _status_for(result: Any) -> int:
    """``partial`` and ``ok`` are both 200; a tool-returned error is always 500."""
    return 500 if result.status == "error" else 200


def _lookup(name: str) -> Optional[ToolDefinition]:
    try:
        return TOOL_REGISTRY.get(name)
    except QueryFailure:
        return None


def _elapsed_ms(started: float) -> int:
    return int((time.monotonic() - started) * 1000)


def _run(tool_name: str, payload: Dict[str, Any]) -> JSONResponse:
    timer_started = time.monotonic()
    definition = _lookup(tool_name)
    if definition is None:
        return _error_response(
            tool=tool_name,
            code="TOOL_NOT_FOUND",
            message=f"unknown tool: {tool_name!r}",
            http_status=404,
            elapsed_ms=_elapsed_ms(timer_started),
        )

    invocation: Any = None
    try:
        if definition.execution_policy == QUERY_POLICY:
            invocation = QueryContext()
        else:
            invocation = ToolInvocation()
        result = TOOL_REGISTRY.execute(tool_name, payload, invocation)
        if tool_name == "diagnose" and result.data is not None:
            request = DiagnosisQuery.model_validate(payload)
            data = result.data.model_dump(mode="python")
            summarize_diagnosis(data, request.job, request.namespace)
            result.data = DiagnosisData.model_validate(present_diagnosis(data))
    except QueryFailure as exc:
        logger.info("tool %s failed: %s (%s)", tool_name, exc.code, exc.message)
        return _error_response(
            tool=tool_name,
            code=exc.code,
            message=exc.message,
            http_status=exc.http_status,
            path=exc.path,
            invocation=invocation,
            elapsed_ms=_elapsed_ms(timer_started),
            definition=definition,
        )
    except ValidationError as exc:
        details = exc.errors()
        first = details[0] if details else {}
        path = ".".join(str(part) for part in first.get("loc", ())) or None
        message = first.get("msg", "invalid request")
        return _error_response(
            tool=tool_name,
            code="INVALID_ARGUMENT",
            message=message,
            http_status=422,
            path=path,
            invocation=invocation,
            elapsed_ms=_elapsed_ms(timer_started),
            definition=definition,
        )
    except Exception:  # noqa: BLE001  # preserve the tool's JSON contract, never a bare traceback
        logger.exception("tool %s crashed", tool_name)
        return _error_response(
            tool=tool_name,
            code="INTERNAL_ERROR",
            message=f"internal error (request_id={_request_id(invocation)})",
            http_status=500,
            invocation=invocation,
            elapsed_ms=_elapsed_ms(timer_started),
            definition=definition,
        )
    response = tool_response_payload(result)
    return JSONResponse(status_code=_status_for(result), content=response)


@router.get("/tools")
def list_tools() -> JSONResponse:
    catalog = ToolCatalog(request_id=uuid.uuid4().hex, tools=TOOL_REGISTRY.describe())
    return JSONResponse(status_code=200, content=catalog.model_dump(mode="json"))


@router.post("/tools/{tool_name}")
def execute_tool(
    tool_name: str,
    payload: Any = Body(default={}),
) -> JSONResponse:
    if not isinstance(payload, dict):
        timer_started = time.monotonic()
        definition = _lookup(tool_name)
        if definition is None:
            return _error_response(
                tool=tool_name,
                code="TOOL_NOT_FOUND",
                message=f"unknown tool: {tool_name!r}",
                http_status=404,
                elapsed_ms=_elapsed_ms(timer_started),
            )
        invocation: Any = ToolInvocation() if definition.execution_policy != QUERY_POLICY else None
        return _error_response(
            tool=tool_name,
            code="INVALID_ARGUMENT",
            message="request body must be a JSON object",
            http_status=422,
            invocation=invocation,
            elapsed_ms=_elapsed_ms(timer_started),
            definition=definition,
        )
    return _run(tool_name, payload)
