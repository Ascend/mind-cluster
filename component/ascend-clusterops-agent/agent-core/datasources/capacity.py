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

"""Per-node CPU/NPU accounting from Node capacity/allocatable and Running requests.

Each whole-card resource key (``huawei.com/Ascend910``, ``huawei.com/npu``) becomes its
own bucket; the two buckets never merge or convert into each other. CPU is accounted in
millicores internally and published in cores. ``cpu`` and the buckets are sibling keys of
the node object (no ``resources`` wrapper). There is no cross-node total: nodes without
any whole-card key stay outside ``nodes`` entirely.
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional

from agent_core.datasources.utils import RESOURCE_PREFIXES, chip_type, parse_non_negative_number, to_int_quantity
from agent_core.query_schemas import CpuResource, NodeResourceData, QueryIssue, ResourceBucket

logger = logging.getLogger(__name__)

_CPU_KEY = "cpu"


@dataclass
class ResourceComputation:
    """Per-node accounting plus the NPU-like keys this version cannot bucket."""

    nodes: Dict[str, NodeResourceData] = field(default_factory=dict)
    # node -> unsupported NPU-like keys (e.g. 310 / vNPU derivatives), logged only
    unsupported: Dict[str, List[str]] = field(default_factory=dict)
    incomplete: bool = False


def _is_unknown_npu_resource(key: str) -> bool:
    """NPU-like keys outside the supported whole-card prefixes (virtual/shared units)."""
    if not isinstance(key, str) or key in RESOURCE_PREFIXES:
        return False
    lowered = key.lower()
    return lowered.startswith("huawei.com/") and ("ascend" in lowered or "npu" in lowered)


def _metadata(obj: Dict[str, Any]) -> Dict[str, Any]:
    return obj.get("metadata") or {}


def _status(obj: Dict[str, Any]) -> Dict[str, Any]:
    return obj.get("status") or {}


def node_whole_card_keys(node: Dict[str, Any]) -> List[str]:
    status = _status(node)
    keys = list((status.get("capacity") or {}).keys()) + list((status.get("allocatable") or {}).keys())
    return [prefix for prefix in RESOURCE_PREFIXES if prefix in keys]


def _to_millicores(value: Any) -> Optional[int]:
    """Parse a Kubernetes CPU quantity into millicores ("2" -> 2000, "500m" -> 500)."""
    if isinstance(value, str):
        text = value.strip()
        if not text:
            return None
        milli = text.endswith("m")
        number = parse_non_negative_number(text[:-1] if milli else text)
        if number is None:
            return None
        return round(number) if milli else round(number * 1000)
    number = parse_non_negative_number(value)
    return round(number * 1000) if number is not None else None


def _container_requests(pod: Dict[str, Any], resource: str, issues: List[QueryIssue], path: str) -> int:
    """Sum one resource over ordinary-container requests (millicores for cpu, cards otherwise)."""
    spec = pod.get("spec") or {}
    total = 0
    for index, container in enumerate(spec.get("containers") or []):
        if not isinstance(container, dict):
            continue
        requests = (container.get("resources") or {}).get("requests") or {}
        if resource not in requests:
            continue
        raw = requests[resource]
        parsed = _to_millicores(raw) if resource == _CPU_KEY else to_int_quantity(raw)
        if parsed is None:
            issues.append(
                QueryIssue(
                    code="SOURCE_INVALID",
                    path=f"{path}.spec.containers[{index}].resources.requests[{resource}]",
                    message=f"invalid quantity: {raw!r}",
                )
            )
            continue
        total += parsed
    return total


def _parse_node_quantity(
    node: str,
    section: str,
    resource: str,
    raw: Any,
    issues: List[QueryIssue],
    result: ResourceComputation,
    parser=to_int_quantity,
) -> Optional[int]:
    """Parse one Node quantity; a missing or invalid value is reported and stays unknown."""
    if raw is None:
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                path=f"nodes[{node}].status.{section}[{resource}]",
                message=f"{section} is missing for {resource}",
            )
        )
        result.incomplete = True
        return None
    parsed = parser(raw)
    if parsed is None:
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                path=f"nodes[{node}].status.{section}[{resource}]",
                message=f"invalid quantity: {raw!r}",
            )
        )
        result.incomplete = True
    return parsed


def _cpu_block(
    node: str,
    capacity: Dict[str, Any],
    allocatable: Dict[str, Any],
    used_millicores: int,
    issues: List[QueryIssue],
    result: ResourceComputation,
) -> CpuResource:
    cap = _parse_node_quantity(
        node, "capacity", _CPU_KEY, capacity.get(_CPU_KEY), issues, result, parser=_to_millicores
    )
    alloc = _parse_node_quantity(
        node, "allocatable", _CPU_KEY, allocatable.get(_CPU_KEY), issues, result, parser=_to_millicores
    )
    idle = (alloc - used_millicores) / 1000 if alloc is not None else None
    if idle is not None and idle < 0:
        logger.warning("negative idle cores on node %s: allocatable=%s used=%s", node, alloc, used_millicores)
    return CpuResource(
        capacity_cores=cap / 1000 if cap is not None else None,
        allocatable_cores=alloc / 1000 if alloc is not None else None,
        used_cores=used_millicores / 1000,
        idle_cores=idle,
    )


def _card_bucket(
    node: str,
    prefix: str,
    capacity: Dict[str, Any],
    allocatable: Dict[str, Any],
    used: int,
    issues: List[QueryIssue],
    result: ResourceComputation,
) -> ResourceBucket:
    cap = _parse_node_quantity(node, "capacity", prefix, capacity.get(prefix), issues, result)
    alloc = _parse_node_quantity(node, "allocatable", prefix, allocatable.get(prefix), issues, result)
    if cap is not None and alloc is not None and cap < alloc:
        # allocatable must not exceed capacity; surface the inconsistent producer data
        # instead of letting the derived difference look like a valid negative count.
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                path=f"nodes[{node}]",
                message=f"capacity ({cap}) is smaller than allocatable ({alloc}) for {prefix}",
            )
        )
        result.incomplete = True
    idle = alloc - used if alloc is not None else None
    if idle is not None and idle < 0:
        logger.warning("negative idle cards on node %s for %s: allocatable=%s used=%s", node, prefix, alloc, used)
    return ResourceBucket(capacity=cap, allocatable=alloc, used=used, idle=idle)


def calculate_resources(
    node_items: List[Dict[str, Any]],
    pod_items: List[Dict[str, Any]],
    issues: List[QueryIssue],
) -> ResourceComputation:
    result = ResourceComputation()

    # Running ordinary-container requests, attributed by spec.nodeName; unbound pods and
    # pods outside the returned node list are dropped (there is no cross-node total).
    known_nodes = {(_metadata(node).get("name") or "") for node in node_items}
    requested: Dict[str, Dict[str, int]] = {}
    for index, pod in enumerate(pod_items):
        if (pod.get("status") or {}).get("phase") != "Running":
            continue
        node_name = (pod.get("spec") or {}).get("nodeName") or ""
        if not node_name or node_name not in known_nodes:
            continue
        pod_path = f"pods[{index}]"
        for key in (_CPU_KEY, *RESOURCE_PREFIXES):
            amount = _container_requests(pod, key, issues, pod_path)
            if amount:
                bucket = requested.setdefault(node_name, {})
                bucket[key] = bucket.get(key, 0) + amount

    for node in node_items:
        name = _metadata(node).get("name") or ""
        status = _status(node)
        capacity = status.get("capacity") or {}
        allocatable = status.get("allocatable") or {}
        unsupported = sorted(
            {key for section in (capacity, allocatable) for key in section if _is_unknown_npu_resource(key)}
        )
        if unsupported:
            result.unsupported[name] = unsupported
            for key in unsupported:
                issues.append(
                    QueryIssue(
                        code="SOURCE_INVALID",
                        path=f"nodes[{name}].status.capacity[{key}]",
                        message=f"resource {key!r} is not a supported whole-card resource; no bucket is created",
                    )
                )
        card_keys = [prefix for prefix in RESOURCE_PREFIXES if prefix in capacity or prefix in allocatable]
        if not card_keys:
            # Outside the expected NPU set: no buckets, no data.node entry, no gap.
            continue
        node_requests = requested.get(name, {})
        buckets = {
            chip_type(prefix): _card_bucket(
                name, prefix, capacity, allocatable, node_requests.get(prefix, 0), issues, result
            )
            for prefix in card_keys
        }
        # ``cpu`` and the whole-card buckets are sibling keys of the node object.
        result.nodes[name] = NodeResourceData(
            cpu=_cpu_block(name, capacity, allocatable, node_requests.get(_CPU_KEY, 0), issues, result),
            **buckets,
        )
    return result
