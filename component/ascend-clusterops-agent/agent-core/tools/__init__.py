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

"""Tool layer package.

Business capabilities live here. ``registry`` holds the shared registry,
``schemas`` the public contract, ``diagnosis`` the legacy diagnosis chain and
the four query modules the read-only tools. ``agent_core.tools`` keeps the public
re-exports below for existing callers; private diagnosis helpers stay in
``tools.diagnosis``.
"""

from __future__ import annotations

# Registry first, then the built-in definitions (registration side effect).
from agent_core.tools.registry import (
    DIAGNOSIS_POLICY,
    QUERY_POLICY,
    TOOL_REGISTRY,
    DiagnosisPolicy,
    QueryPolicy,
    ToolDefinition,
    ToolInvocation,
    ToolRegistry,
)
from agent_core.tools import builtin as _builtin  # noqa: F401  # registers the five tools
from agent_core.tools.schemas import (
    DiagnosisData,
    DiagnosisQuery,
    DiagnosisResult,
    ToolCatalog,
    ToolDescriptor,
    ToolEnvelope,
    tool_response_payload,
)
from agent_core.tools.diagnosis import (
    COLLECT_TIMEOUT,
    COLLECTOR_PORT,
    RECENT_TTL,
    TRIGGER_TIMEOUT,
    WORK_ROOT,
    assemble_diag_input,
    cache_diag_final_text,
    diagnose,
    diagnose_cached,
    dispatch_collect,
    is_diag_active,
    job_exists,
    nodes_note,
    run_diag,
    tracker,
)

__all__ = [
    "COLLECT_TIMEOUT",
    "COLLECTOR_PORT",
    "DIAGNOSIS_POLICY",
    "QUERY_POLICY",
    "RECENT_TTL",
    "TOOL_REGISTRY",
    "TRIGGER_TIMEOUT",
    "WORK_ROOT",
    "DiagnosisData",
    "DiagnosisPolicy",
    "DiagnosisQuery",
    "DiagnosisResult",
    "QueryPolicy",
    "ToolCatalog",
    "ToolDefinition",
    "ToolDescriptor",
    "ToolEnvelope",
    "ToolInvocation",
    "ToolRegistry",
    "assemble_diag_input",
    "cache_diag_final_text",
    "diagnose",
    "diagnose_cached",
    "dispatch_collect",
    "is_diag_active",
    "job_exists",
    "nodes_note",
    "run_diag",
    "tool_response_payload",
    "tracker",
]
