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

"""Tool: cluster_resource.

Capacity comes from Node ``status.capacity``/``allocatable`` and occupation from Running
ordinary-container requests; every whole-card key (``Ascend910``/``npu``) is its own
bucket, a sibling of ``cpu``. Fault/isolated are independent per-bucket statistics over
the same device facts as faulty_device/isolated_device (identity sets, no capacity-difference
reconciliation, a card that is both faulty and isolated counts once in each field); a
source problem only nulls the affected field and never changes the envelope status.
"""

from __future__ import annotations

from typing import Dict, List, Set, Tuple

from agent_core.datasources.capacity import calculate_resources
from agent_core.datasources.query_context import QueryContext, get_gateway
from agent_core.datasources.node_targets import validate_target_nodes
from agent_core.query_schemas import (
    NodeResourceData,
    QueryIssue,
    ResourceData,
    ResourceQuery,
    ResourceQueryResult,
)
from agent_core.tools.envelope import (
    build_result,
    is_empty_scope,
    log_node_listing_unavailable,
    log_parse_issues,
    log_source_issue,
)
from agent_core.tools.faulty_device import FaultCollection, collect_faults
from agent_core.tools.isolated_device import IsolationCollection, collect_isolation
from agent_core.tools.data_builders import build_resource_data

TOOL = "cluster_resource"


def _fault_identities(faults: FaultCollection) -> Dict[str, Dict[str, Set[str]]]:
    """node -> chip type -> device ids of the recognizable NPU fault records.

    Only ``device_type=npu`` records with a non-empty ``device_id`` and a ``chip_type``
    count: node-level records and non-NPU categories are not physical cards.
    """
    identities: Dict[str, Dict[str, Set[str]]] = {}
    for node, entry in faults.nodes.items():
        for record in entry.devices or []:
            if record.device_type != "npu" or not record.device_id or not record.chip_type:
                continue
            identities.setdefault(node, {}).setdefault(record.chip_type, set()).add(record.device_id)
    return identities


def _isolated_identities(
    isolation: IsolationCollection,
    device_names_by_node: Dict[str, Dict[str, Set[str]]],
    issues: List[QueryIssue],
) -> Tuple[Dict[str, Dict[str, Set[str]]], Set[str]]:
    """Attribute each isolated device name to the chip buckets whose name set contains it.

    A name in exactly one device-name set belongs to that bucket; a name in both sets is an
    alias double-write counted once per bucket; a name in neither cannot be attributed
    and only that node's ``isolated`` counts become ``null`` (``fault`` is independent
    and still published from the device facts).
    """
    attributed: Dict[str, Dict[str, Set[str]]] = {}
    unattributed: Set[str] = set()
    for node, entry in isolation.nodes.items():
        node_device_names = device_names_by_node.get(node, {})
        for record in entry.devices:
            hits = {chip for chip, names in node_device_names.items() if record.device_id in names}
            if not hits:
                unattributed.add(node)
                issues.append(
                    QueryIssue(
                        code="SOURCE_INVALID",
                        path=f"node[{node}].isolated[{record.device_id}]",
                        message=(
                            "isolated device is not in any whole-card device-name set; isolated counts are null for this node"
                        ),
                    )
                )
                continue
            for chip in hits:
                attributed.setdefault(node, {}).setdefault(chip, set()).add(record.device_id)
    return attributed, unattributed


def apply_fault_and_isolation_counts(
    nodes: Dict[str, NodeResourceData],
    faults: FaultCollection,
    isolation: IsolationCollection,
    issues: List[QueryIssue],
) -> None:
    """Fill every bucket's fault/isolated with independent identity-set statistics.

    Per bucket: ``fault_P = |F_P|`` and ``isolated_P = |I_P|``; a card that is both
    faulty and isolated is counted once in each field (fault never subtracts isolation).
    The two failure domains are separate: ``fault`` is ``null`` only when the device
    facts for that prefix are unavailable, while ``isolated`` is ``null`` only when the
    isolation source is unreadable or its identities cannot be attributed to a bucket.
    ``0`` means the evidence for that prefix was read and simply has no identity of its
    own; both buckets are computed independently and never merged.
    """
    fault_ids = _fault_identities(faults)
    isolated_ids, unattributed = _isolated_identities(isolation, faults.device_names_by_node, issues)
    for node, node_data in nodes.items():
        chain = faults.prefix_state.get(node, {})
        isolation_known = isolation.cm_ok and node not in isolation.broken
        isolated_known = isolation_known and node not in unattributed
        for chip, bucket in node_data.buckets().items():
            # A chip with a known chain state may be read (True) or broken (False); a chip
            # absent from the chain is a clean zero only when the node itself was covered.
            fault_known = chain[chip] if chip in chain else node in faults.covered
            bucket.fault = len(fault_ids.get(node, {}).get(chip, set())) if fault_known else None
            bucket.isolated = len(isolated_ids.get(node, {}).get(chip, set())) if isolated_known else None


def cluster_resource(request: ResourceQuery, ctx: QueryContext) -> ResourceQueryResult:
    issues: List[QueryIssue] = []

    if is_empty_scope(ctx):
        return build_result(ResourceQueryResult, TOOL, ctx, ResourceData(), core_ok=True)

    missing_issues = validate_target_nodes(ctx)
    nodes = ctx.requested_nodes
    gateway = get_gateway(ctx)
    node_result = gateway.list_nodes(nodes)
    pod_result = gateway.list_pods(None)
    log_source_issue(TOOL, "nodes", node_result.status, node_result.error_code, node_result.message)
    log_source_issue(TOOL, "pods", pod_result.status, pod_result.error_code, pod_result.message)

    listing_complete = node_result.listing_complete and pod_result.listing_complete
    computation = calculate_resources(node_result.items, pod_result.items, issues)

    # Independent fault/isolated statistics: always attempted for the requested scope,
    # never gated by capacity and never changing the envelope status.
    faults = collect_faults(ctx)
    isolation = collect_isolation(ctx)
    apply_fault_and_isolation_counts(computation.nodes, faults, isolation, issues)

    # A stopped or partial read only reduces coverage and is recorded in the server log;
    # the public status stays ``partial`` and never turns into an execution error.
    targets_complete = (
        # "empty" is a complete answer: the cluster simply has no Node objects.
        node_result.status in ("ok", "empty")
        and not node_result.missing
        and pod_result.status in ("ok", "empty")
        and listing_complete
    )
    if nodes is not None:
        for missing in node_result.missing:
            log_source_issue(
                TOOL, "nodes", "missing", "TARGET_NOT_FOUND", f"node {missing!r} does not exist or is not readable"
            )
        # A named target present in the Node list without any recognized whole-card key is
        # a complete "no NPU resources" answer, unless it only carries unknown card keys.
        present = {(item.get("metadata") or {}).get("name") or "" for item in node_result.items}
        for node in nodes:
            if node in present and node not in computation.nodes and node in computation.unsupported:
                targets_complete = False
                log_source_issue(
                    TOOL,
                    "nodes",
                    "invalid",
                    "SOURCE_INVALID",
                    f"node {node!r} only carries unrecognized card resources: "
                    + ", ".join(computation.unsupported[node]),
                )
    elif not node_result.listing_complete:
        log_node_listing_unavailable(
            TOOL,
            f"node listing stopped early; observed={len(node_result.items)}, remaining node count unknown",
            node_result.source_id,
        )

    log_parse_issues(TOOL, issues)
    data = build_resource_data(computation.nodes)
    return build_result(
        ResourceQueryResult,
        TOOL,
        ctx,
        data,
        core_ok=True,
        complete=targets_complete and not computation.incomplete and not missing_issues,
        errors=missing_issues,
    )
