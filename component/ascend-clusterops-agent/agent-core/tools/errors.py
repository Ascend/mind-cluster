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

"""Execution-error envelope for the tool HTTP boundary.

The HTTP handler still maps ``QueryFailure.http_status`` to the response code; every tool
returns the same envelope in the body.
"""

from __future__ import annotations

import uuid
from typing import Optional

from agent_core.query_schemas import QueryIssue, ToolEnvelope


def build_error_envelope(
    *,
    tool: str,
    code: str,
    message: str,
    request_id: str = "",
    elapsed_ms: int = 0,
    path: Optional[str] = None,
) -> ToolEnvelope:
    """Single-issue, ``data=None`` execution error envelope."""
    return ToolEnvelope(
        request_id=request_id or uuid.uuid4().hex,
        tool=tool,
        status="error",
        elapsed_ms=elapsed_ms,
        data=None,
        errors=[QueryIssue(code=code, message=message, path=path)],
    )
