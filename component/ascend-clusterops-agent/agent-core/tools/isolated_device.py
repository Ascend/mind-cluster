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

"""Tool: isolated_device.

``clusterd-manual-info-cm`` is the only source: ``Total`` is the roster of isolated
devices, ``Detail`` supplements roster members with level/code/time facts. The DP
ConfigMap is never read. A missing CM is a legitimate empty roster (``ok``); a failed
CM read or an unparsable node roster only downgrades the answer to ``partial``.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Dict, List, Set

from agent_core.datasources.query_context import QueryContext, get_gateway
from agent_core.datasources.isolation import parse_manual_isolation
from agent_core.datasources.sources import MANUAL_ISOLATION
from agent_core.datasources.node_targets import validate_target_nodes
from agent_core.query_schemas import (
    IsolationData,
    IsolationQuery,
    IsolationQueryResult,
    NodeIsolation,
    QueryIssue,
)
from agent_core.tools.envelope import (
    OK_SOURCE_STATUSES,
    build_result,
    is_empty_scope,
    log_parse_issues,
    log_source_issue,
)
from agent_core.tools.data_builders import build_isolation_data

TOOL = "isolated_device"


@dataclass
class IsolationCollection:
    """Processing facts shared by isolated_device and resource statistics."""

    nodes: Dict[str, NodeIsolation] = field(default_factory=dict)
    # The CM read itself succeeded (a missing CM is a legitimate empty roster).
    cm_ok: bool = False
    # Nodes whose roster could not be fully determined; their published members are partial.
    broken: Set[str] = field(default_factory=set)
    complete: bool = False


def collect_isolation(ctx: QueryContext) -> IsolationCollection:
    """Read the ClusterD isolation roster without constructing a result envelope."""
    issues: List[QueryIssue] = []
    gateway = get_gateway(ctx)
    name = MANUAL_ISOLATION.name
    result = gateway.get_cm(MANUAL_ISOLATION.namespace, name, absent_expected=MANUAL_ISOLATION.absent_expected)
    per_node: Dict[str, NodeIsolation] = {}
    broken: Set[str] = set()
    if result.ok:
        data = (result.items or [{}])[0].get("data") or {}
        per_node, broken = parse_manual_isolation(data, ctx.requested_nodes, result.source_id, issues, data_key=name)
    log_source_issue(TOOL, "manual_isolation", result.status, result.error_code, result.message)
    # A missing CM (absent_expected) or a node key absent from its data are legitimate
    # empty rosters; only a failed read or a broken node roster downgrades the answer.
    cm_ok = result.status in OK_SOURCE_STATUSES
    log_parse_issues(TOOL, issues)
    return IsolationCollection(nodes=per_node, cm_ok=cm_ok, broken=broken, complete=cm_ok and not broken)


def isolated_device(request: IsolationQuery, ctx: QueryContext) -> IsolationQueryResult:
    if is_empty_scope(ctx):
        return build_result(IsolationQueryResult, TOOL, ctx, IsolationData(), core_ok=True)
    missing = validate_target_nodes(ctx)
    collected = collect_isolation(ctx)
    return build_result(
        IsolationQueryResult,
        TOOL,
        ctx,
        build_isolation_data(collected.nodes),
        core_ok=True,
        complete=collected.complete and not missing,
        errors=missing,
    )
