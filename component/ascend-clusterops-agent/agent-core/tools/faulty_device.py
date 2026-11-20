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

"""Tool: faulty_device.

Primary source is the ClusterD device summary shards; the per-node DP ConfigMap is only
read for targets the summary does not cover, and only for resource prefixes the primary
could not parse (the same node may take different prefixes from different sources). The
node summary supplies ``node_health`` and, when the primary has no node-level fact for a
covered node, the node-level ``devices[]`` records. Switch/DPU/public-fault sources are
independent supplements. All records are grouped by node and every source failure is
logged only.
"""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Set

from agent_core.datasources.capacity import node_whole_card_keys
from agent_core.datasources.clusterd.device_summary import parse_cluster_device_shards
from agent_core.datasources.clusterd.dpu_summary import parse_dpu_cm
from agent_core.datasources.clusterd.node_summary import parse_node_cm, parse_node_faults
from agent_core.datasources.clusterd.switch_summary import parse_switch_cm, parse_switch_faults
from agent_core.datasources.query_context import QueryContext, get_gateway
from agent_core.datasources.gateway import KubernetesReadGateway
from agent_core.datasources.device_list import (
    list_resource_prefixes,
    parse_device_list_grouped,
    parse_device_names_by_chip,
    decode_dp_deviceinfo,
)
from agent_core.datasources.dp_devices import DpDocument, read_dp_configmaps
from agent_core.datasources.public_faults import parse_public_faults
from agent_core.datasources.sources import CLUSTER_DEVICES, DP_DEVICES, DPU, NODE_CM, PUBLIC_FAULTS, SWITCH
from agent_core.datasources.node_targets import validate_target_nodes
from agent_core.datasources.utils import chip_type, dedupe_records
from agent_core.query_schemas import (
    DeviceFaultRecord,
    DpuFaultRecord,
    FaultData,
    FaultQuery,
    FaultQueryResult,
    NodeFaults,
    NodeHealth,
    PublicFaultRecord,
    QueryIssue,
    SwitchFaultRecord,
)
from agent_core.tools.envelope import (
    OK_SOURCE_STATUSES,
    build_result,
    is_empty_scope,
    log_node_listing_unavailable,
    log_parse_issues,
    log_source_issue,
)
from agent_core.tools.data_builders import build_fault_data

TOOL = "faulty_device"


@dataclass
class FaultCollection:
    """Processing facts shared by faulty_device and resource statistics."""

    nodes: Dict[str, NodeFaults] = field(default_factory=dict)
    device_facts_complete: bool = False
    # Health is part of the faulty_device answer only; resource evidence needs device facts.
    node_health_complete: bool = False
    # Evidence for cluster_resource's independent fault/isolated counts (internal only):
    # covered nodes, per-chip base device-name sets and the per-chip evidence-chain state. A chip
    # missing from prefix_state on a covered node means "read and no identity of that
    # prefix" (a clean zero); a False entry means the chain is broken (null).
    covered: Set[str] = field(default_factory=set)
    device_names_by_node: Dict[str, Dict[str, Set[str]]] = field(default_factory=dict)
    prefix_state: Dict[str, Dict[str, bool]] = field(default_factory=dict)


def _in_scope(node: str, nodes: Optional[Set[str]]) -> bool:
    return nodes is None or node in nodes


def _prefix_clean(issues: List[QueryIssue], before: int) -> bool:
    """A node is covered when its device facts decoded without a structural error."""
    return not any(issue.code == "SOURCE_INVALID" for issue in issues[before:])


def _record_key(record) -> str:
    """Two records are the same only when every public field matches."""
    return json.dumps(record.model_dump(mode="json"), sort_keys=True)


@dataclass
class _DeviceEvidence:
    """Mutable device evidence collected from the primary shards and DP fallback."""

    device_groups: Dict[str, Dict[str, List[DeviceFaultRecord]]] = field(default_factory=dict)
    valid_prefixes_by_node: Dict[str, Set[str]] = field(default_factory=dict)
    device_names_by_node: Dict[str, Dict[str, Set[str]]] = field(default_factory=dict)
    prefix_state: Dict[str, Dict[str, bool]] = field(default_factory=dict)
    covered: Set[str] = field(default_factory=set)
    node_level_present: Set[str] = field(default_factory=set)
    listing_complete: bool = True


def _dp_fallback_targets(
    nodes: Optional[List[str]],
    expected_nodes: Optional[List[str]],
    covered: Set[str],
) -> Optional[List[str]]:
    """Resolve the DP read scope: None means full cluster, [] means an explicit empty scope."""
    if nodes is not None:
        return [node for node in nodes if node not in covered]
    if expected_nodes is not None:
        return [node for node in expected_nodes if node not in covered]
    return None


def _should_read_dp_fallback(targets: Optional[List[str]]) -> bool:
    """None means full-cluster, [] means explicit empty scope, a list means scoped nodes."""
    return targets is None or bool(targets)


def _merge_dp_document(
    document: DpDocument,
    nodes_set: Optional[Set[str]],
    issues: List[QueryIssue],
    evidence: _DeviceEvidence,
) -> None:
    """Merge one DP document into the evidence, keeping primary-valid prefixes authoritative."""
    if not _in_scope(document.node, nodes_set):
        return
    before = len(issues)
    device_list, _update_time, topology = decode_dp_deviceinfo(document.data, document.source_id, issues)
    if not device_list:
        return
    groups, valid_prefixes = parse_device_list_grouped(device_list, issues, document.source_id, topology)
    primary_valid = evidence.valid_prefixes_by_node.get(document.node, set())
    primary_chips = {chip_type(prefix) for prefix in primary_valid}
    merged = {
        prefix: list(records)
        for prefix, records in evidence.device_groups.get(document.node, {}).items()
        if prefix in primary_valid
    }
    for prefix, records in groups.items():
        if prefix not in primary_valid:
            merged.setdefault(prefix, []).extend(records)
    evidence.device_groups[document.node] = merged
    # Node-level entries from DP suppress the node-summary fallback exactly like primary ones.
    if any(record.device_id == "" for records in merged.values() for record in records):
        evidence.node_level_present.add(document.node)
    merged_device_names = {
        chip: names
        for chip, names in evidence.device_names_by_node.get(document.node, {}).items()
        if chip in primary_chips
    }
    for chip, names in parse_device_names_by_chip(device_list).items():
        if chip not in primary_chips:
            merged_device_names[chip] = names
    evidence.device_names_by_node[document.node] = merged_device_names
    merged_state = dict(evidence.prefix_state.get(document.node, {}))
    for prefix in list_resource_prefixes(device_list):
        if prefix not in primary_valid:
            merged_state[chip_type(prefix)] = prefix in valid_prefixes
    evidence.prefix_state[document.node] = merged_state
    if _prefix_clean(issues, before):
        evidence.covered.add(document.node)


def _apply_dp_fallback(
    ctx: QueryContext,
    nodes: Optional[List[str]],
    nodes_set: Optional[Set[str]],
    expected_nodes: Optional[List[str]],
    issues: List[QueryIssue],
    evidence: _DeviceEvidence,
) -> None:
    """Read the per-node DP ConfigMaps for uncovered targets and merge them into evidence."""
    fallback_targets = _dp_fallback_targets(nodes, expected_nodes, evidence.covered)
    if not _should_read_dp_fallback(fallback_targets):
        return
    # A primary that answered the node is not re-read from DP.
    listing, documents, failed = read_dp_configmaps(
        ctx,
        DP_DEVICES.namespace,
        DP_DEVICES.name_prefix,
        nodes=fallback_targets,
        skip=evidence.covered,
    )
    if listing is not None:
        log_source_issue(TOOL, "dp_devices", listing.status, listing.error_code, listing.message)
        evidence.listing_complete = evidence.listing_complete and listing.listing_complete
    for read in failed:
        # A failed target read only stops that node from being covered; the recorded
        # error already turns the answer into a partial one.
        log_source_issue(TOOL, "dp_devices", read.status, read.error_code, read.message)
    for document in documents:
        _merge_dp_document(document, nodes_set, issues, evidence)


@dataclass
class _SupplementalFaults:
    """Node health and independent fault records supplementing device evidence."""

    node_source_complete: bool = False
    node_health: Dict[str, NodeHealth] = field(default_factory=dict)
    node_level_records: Dict[str, List[DeviceFaultRecord]] = field(default_factory=dict)
    switch_records: Dict[str, List[SwitchFaultRecord]] = field(default_factory=dict)
    dpu_records: Dict[str, List[DpuFaultRecord]] = field(default_factory=dict)
    public_records: Dict[str, List[PublicFaultRecord]] = field(default_factory=dict)


def _collect_primary_devices(
    gateway: KubernetesReadGateway,
    nodes_set: Optional[Set[str]],
    issues: List[QueryIssue],
    evidence: _DeviceEvidence,
) -> None:
    """Collect primary device facts and track which nodes and prefixes they cover."""
    result = gateway.list_cms(CLUSTER_DEVICES.namespace, CLUSTER_DEVICES.label_selector, CLUSTER_DEVICES.name_prefix)
    log_source_issue(TOOL, "cluster_devices", result.status, result.error_code, result.message)
    evidence.listing_complete = evidence.listing_complete and result.listing_complete
    entries = parse_cluster_device_shards(result.items, issues, result.source_id)
    for node, entry in sorted(entries.items()):
        if not _in_scope(node, nodes_set):
            continue
        before = len(issues)
        source_id = entry.source.source_id if entry.source is not None else result.source_id
        groups, valid_prefixes = parse_device_list_grouped(entry.device_list, issues, source_id, entry.topology)
        evidence.device_groups[node] = groups
        evidence.valid_prefixes_by_node[node] = valid_prefixes
        evidence.device_names_by_node[node] = parse_device_names_by_chip(entry.device_list)
        evidence.prefix_state[node] = {
            chip_type(prefix): prefix in valid_prefixes for prefix in list_resource_prefixes(entry.device_list)
        }
        if any(record.device_id == "" for records in groups.values() for record in records):
            evidence.node_level_present.add(node)
        if _prefix_clean(issues, before):
            evidence.covered.add(node)


def _expected_npu_nodes(gateway: KubernetesReadGateway) -> Optional[List[str]]:
    """Return full-cluster NPU targets, or None when the node listing is unavailable."""
    expected_nodes: Optional[List[str]] = None
    node_read = gateway.list_nodes(None)
    if node_read.ok:
        expected_nodes = [
            (item.get("metadata") or {}).get("name") or "" for item in node_read.items if node_whole_card_keys(item)
        ]
        expected_nodes = [name for name in expected_nodes if name]
    else:
        log_node_listing_unavailable(
            TOOL,
            f"expected NPU nodes unavailable: {node_read.message or node_read.status}",
            node_read.source_id,
        )
    return expected_nodes


def _collect_node_faults(
    gateway: KubernetesReadGateway,
    nodes_set: Optional[Set[str]],
    issues: List[QueryIssue],
    node_level_present: Set[str],
    supplement: _SupplementalFaults,
) -> None:
    """Read node health and fill node-level faults absent from device evidence."""
    result = gateway.get_cm(NODE_CM.namespace, NODE_CM.name, absent_expected=NODE_CM.absent_expected)
    log_source_issue(TOOL, "node_faults", result.status, result.error_code, result.message)
    supplement.node_source_complete = result.status in OK_SOURCE_STATUSES
    if result.ok:
        for node, info in parse_node_cm(result.items, issues, result.source_id).items():
            if not _in_scope(node, nodes_set):
                continue
            status = info.get("node_status")
            if isinstance(status, str) and status.strip():
                supplement.node_health[node] = NodeHealth(node_status=status)
            if node not in node_level_present:
                # Node-level facts the device sources did not provide. This fallback does not
                # require device coverage: an uncovered node still publishes its node-level
                # records (the envelope stays ``partial`` because its NPU facts are missing).
                records = parse_node_faults(info)
                if records:
                    supplement.node_level_records.setdefault(node, []).extend(records)


def _collect_independent_faults(
    gateway: KubernetesReadGateway,
    nodes_set: Optional[Set[str]],
    issues: List[QueryIssue],
    supplement: _SupplementalFaults,
) -> None:
    """Read switch, DPU and public faults independently of device coverage."""
    result = gateway.list_cms(SWITCH.namespace, SWITCH.label_selector, SWITCH.name_prefix)
    log_source_issue(TOOL, "switch_faults", result.status, result.error_code, result.message)
    for node, info in parse_switch_cm(result.items, issues, result.source_id).items():
        if not _in_scope(node, nodes_set):
            continue
        records = parse_switch_faults(node, info, issues)
        if records:
            supplement.switch_records[node] = records

    result = gateway.list_cms(DPU.namespace, DPU.label_selector, DPU.name_prefix)
    log_source_issue(TOOL, "dpu_faults", result.status, result.error_code, result.message)
    for node, records in parse_dpu_cm(result.items, issues, result.source_id).items():
        if _in_scope(node, nodes_set) and records:
            supplement.dpu_records[node] = records

    result = gateway.get_cm(PUBLIC_FAULTS.namespace, PUBLIC_FAULTS.name, absent_expected=PUBLIC_FAULTS.absent_expected)
    log_source_issue(TOOL, "public_faults", result.status, result.error_code, result.message)
    if result.ok:
        data = (result.items or [{}])[0].get("data") or {}
        supplement.public_records = parse_public_faults(
            data, nodes_set, result.source_id, issues, data_key=PUBLIC_FAULTS.name
        )


def _group_faults_by_node(
    evidence: _DeviceEvidence,
    supplement: _SupplementalFaults,
) -> Dict[str, NodeFaults]:
    """Group and deduplicate records, preserving unknown versus empty device facts."""
    names = (
        set(evidence.device_groups)
        | set(supplement.node_level_records)
        | set(supplement.switch_records)
        | set(supplement.dpu_records)
        | set(supplement.public_records)
        | set(supplement.node_health)
    )
    collected: Dict[str, NodeFaults] = {}
    for node in sorted(names):
        if node in evidence.device_groups or node in supplement.node_level_records:
            devices = [record for records in evidence.device_groups.get(node, {}).values() for record in records]
            devices.extend(supplement.node_level_records.get(node, []))
            devices = dedupe_records(devices, _record_key)
        else:
            devices = None
        collected[node] = NodeFaults(
            devices=devices,
            switch=dedupe_records(supplement.switch_records.get(node, []), _record_key),
            dpu=dedupe_records(supplement.dpu_records.get(node, []), _record_key),
            public_faults=dedupe_records(supplement.public_records.get(node, []), _record_key),
            node_health=supplement.node_health.get(node),
        )

    return collected


def collect_faults(ctx: QueryContext) -> FaultCollection:
    """Read and parse every faulty-device source without constructing an envelope."""
    issues: List[QueryIssue] = []
    nodes = ctx.requested_nodes
    nodes_set = set(nodes) if nodes is not None else None
    gateway = get_gateway(ctx)
    evidence = _DeviceEvidence()
    supplement = _SupplementalFaults()

    _collect_primary_devices(gateway, nodes_set, issues, evidence)
    expected_nodes = _expected_npu_nodes(gateway) if nodes is None else None
    _apply_dp_fallback(ctx, nodes, nodes_set, expected_nodes, issues, evidence)
    _collect_node_faults(gateway, nodes_set, issues, evidence.node_level_present, supplement)
    _collect_independent_faults(gateway, nodes_set, issues, supplement)
    collected = _group_faults_by_node(evidence, supplement)

    # ---- completeness: every target has device facts and a node health status ---- #
    targets = nodes if nodes is not None else (expected_nodes or [])
    if nodes is not None:
        targets_complete = all(node in evidence.covered for node in targets)
    else:
        targets_complete = (
            expected_nodes is not None
            and evidence.listing_complete
            and all(node in evidence.covered for node in expected_nodes)
        )
    node_health_complete = supplement.node_source_complete and all(node in supplement.node_health for node in targets)

    log_parse_issues(TOOL, issues)
    return FaultCollection(
        nodes=collected,
        device_facts_complete=targets_complete and evidence.listing_complete,
        node_health_complete=node_health_complete,
        covered=evidence.covered,
        device_names_by_node=evidence.device_names_by_node,
        prefix_state=evidence.prefix_state,
    )


def faulty_device(request: FaultQuery, ctx: QueryContext) -> FaultQueryResult:
    if is_empty_scope(ctx):
        return build_result(FaultQueryResult, TOOL, ctx, FaultData(), core_ok=True)
    missing = validate_target_nodes(ctx)
    collected = collect_faults(ctx)
    return build_result(
        FaultQueryResult,
        TOOL,
        ctx,
        build_fault_data(collected.nodes),
        core_ok=True,
        complete=collected.device_facts_complete and collected.node_health_complete and not missing,
        errors=missing,
    )
