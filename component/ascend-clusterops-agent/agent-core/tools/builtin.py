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

"""Built-in tool definitions.

Each capability registers itself on the shared registry. Adding a tool means
adding its module plus one entry here (or calling ``TOOL_REGISTRY.register``
from the module), with no HTTP/CLI wiring.
"""

from __future__ import annotations

from agent_core.query_schemas import (
    FaultQuery,
    IsolationQuery,
    JobQuery,
    FaultQueryResult,
    IsolationQueryResult,
    JobQueryResult,
    ResourceQueryResult,
    ResourceQuery,
)
from agent_core.tools.cluster_resource import cluster_resource
from agent_core.tools.diagnosis_adapter import diagnose_handler
from agent_core.tools.faulty_device import faulty_device
from agent_core.tools.isolated_device import isolated_device
from agent_core.tools.registry import (
    DIAGNOSIS_POLICY,
    QUERY_POLICY,
    TOOL_REGISTRY,
    ToolDefinition,
)
from agent_core.tools.schemas import DiagnosisQuery, DiagnosisResult
from agent_core.tools.task_status import task_status

BUILTIN_TOOLS = (
    ToolDefinition(
        name="faulty_device",
        description="List faulty NPU devices from ClusterD summaries with per-node DP fallback, plus node health.",
        input_model=FaultQuery,
        output_model=FaultQueryResult,
        handler=faulty_device,
        execution_policy=QUERY_POLICY,
        read_only=True,
    ),
    ToolDefinition(
        name="isolated_device",
        description=(
            "List ClusterD-confirmed isolated devices per node with fault level, fault code and last separate time."
        ),
        input_model=IsolationQuery,
        output_model=IsolationQueryResult,
        handler=isolated_device,
        execution_policy=QUERY_POLICY,
        read_only=True,
    ),
    ToolDefinition(
        name="cluster_resource",
        description=(
            "Report per-node CPU cores and per-chip-type card accounting "
            "(capacity, allocatable, used, idle) with independent fault/isolated counts."
        ),
        input_model=ResourceQuery,
        output_model=ResourceQueryResult,
        handler=cluster_resource,
        execution_policy=QUERY_POLICY,
        read_only=True,
    ),
    ToolDefinition(
        name="task_status",
        description="Report a job's authoritative status from its AscendJob/Volcano Job object plus its live Pods.",
        input_model=JobQuery,
        output_model=JobQueryResult,
        handler=task_status,
        execution_policy=QUERY_POLICY,
        read_only=True,
    ),
    ToolDefinition(
        name="diagnose",
        description=(
            "Collect, clean and diagnose a task's logs, returning the deterministic fault report; "
            "with a configured model the report is summarized, otherwise the report text is returned."
        ),
        input_model=DiagnosisQuery,
        output_model=DiagnosisResult,
        handler=diagnose_handler,
        execution_policy=DIAGNOSIS_POLICY,
        read_only=False,
    ),
)

for _definition in BUILTIN_TOOLS:
    TOOL_REGISTRY.register(_definition)
