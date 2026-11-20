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

"""Pure DeviceList parser shared by the ClusterD shard and the single-node DP ConfigMap."""

from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Set, Tuple

from agent_core.datasources.faults import fault_detail
from agent_core.datasources.utils import (
    RESOURCE_PREFIXES,
    chip_type,
    decode_json_string,
    first_non_empty,
    infer_time_unit,
    normalize_topology,
    positive_number,
    supported_whole_card_key,
    split_csv,
    strip_device_list_suffix,
    timestamp_value,
)
from agent_core.query_schemas import DeviceFaultRecord, FaultDetail, QueryIssue, SuperPod


def list_resource_prefixes(device_list: Dict[str, Any]) -> List[str]:
    prefixes: List[str] = []
    for key in device_list:
        base, _suffix = strip_device_list_suffix(str(key))
        if supported_whole_card_key(base) and base not in prefixes:
            prefixes.append(base)
    return prefixes


@dataclass
class _PrefixMembers:
    """One prefix's raw membership sets, decoded once and shared by records and device-name sets."""

    names: Set[str]
    unhealthy: Set[str]
    network: Set[str]
    recovering: Set[str]
    fault_items: List[Any]
    fault_present: bool


def _decode_faults_silent(raw: Any) -> Any:
    """Decode ``-Fault`` without reporting; the device-name path ignores malformed payloads."""
    if isinstance(raw, str):
        try:
            return json.loads(raw)
        except (TypeError, ValueError):
            return None
    return raw


def _members(
    device_list: Dict[str, Any],
    prefix: str,
    issues: Optional[List[QueryIssue]] = None,
    source_id: str = "",
) -> _PrefixMembers:
    """Read one prefix's base/health/fault members; malformed faults are reported only when
    ``issues`` is given (records path), matching the previous roster/record split.
    """
    raw_faults = device_list.get(prefix + "-Fault")
    fault_items: List[Any] = []
    if raw_faults is not None:
        if issues is None:
            value = _decode_faults_silent(raw_faults)
        else:
            value = decode_json_string(raw_faults, prefix + "-Fault", issues, source_id)
        if isinstance(value, list):
            fault_items = value
        elif value is not None and issues is not None:
            issues.append(
                QueryIssue(
                    code="SOURCE_INVALID",
                    source_id=source_id,
                    path=prefix + "-Fault",
                    message="Fault is not an array",
                )
            )
    return _PrefixMembers(
        names=set(split_csv(device_list.get(prefix))),
        unhealthy=set(split_csv(device_list.get(prefix + "-Unhealthy"))),
        network=set(split_csv(device_list.get(prefix + "-NetworkUnhealthy"))),
        recovering=set(split_csv(device_list.get(prefix + "-Recovering"))),
        fault_items=fault_items,
        fault_present=raw_faults is not None,
    )


def parse_device_names_by_chip(device_list: Dict[str, Any]) -> Dict[str, Set[str]]:
    """Return every known device name for each chip type on one node.

    The value of a base prefix lists only the allocatable devices; an isolated or
    unhealthy card is moved out of it into ``-Unhealthy``/``-NetworkUnhealthy``/
    ``-Recovering``/``-Fault``. Isolated-card attribution must use the union, otherwise
    a genuinely isolated card is "not in any device-name set" and its count becomes null.
    ClusterD isolation members carry no prefix, so their bucket is resolved by looking
    each name up in these sets.
    """
    names_by_chip: Dict[str, Set[str]] = {}
    if not isinstance(device_list, dict):
        return names_by_chip
    for prefix in RESOURCE_PREFIXES:
        if prefix not in device_list:
            continue
        members = _members(device_list, prefix)
        names = members.names | members.unhealthy | members.network | members.recovering
        names |= {
            str(item.get("npu_name")).strip()
            for item in members.fault_items
            if isinstance(item, dict) and str(item.get("npu_name") or "").strip()
        }
        names_by_chip[chip_type(prefix)] = names
    return names_by_chip


_FAULT_LEVEL_FIELDS = ("fault_handling", "large_model_fault_level", "fault_level")


def _device_type(fault_type: Any) -> Optional[str]:
    """Node-level device categories are normalized (``PSU`` -> ``psu``), never guessed."""
    if not isinstance(fault_type, str) or not fault_type.strip():
        return None
    return fault_type.strip().lower()


def _superpod(topology: Optional[Dict[str, Any]]) -> Optional[SuperPod]:
    if not topology:
        return None
    if all(topology.get(key) is None for key in ("super_pod_id", "rack_id", "server_index")):
        return None
    return SuperPod(
        super_pod_id=topology.get("super_pod_id"),
        rack_id=topology.get("rack_id"),
        server_index=topology.get("server_index"),
    )


def _fault_details(raw: Dict[str, Any], path: str, source_id: str, issues: List[QueryIssue]) -> List[FaultDetail]:
    """Split one source fault entry into one detail per code, keeping the source code order."""
    codes = split_csv(raw.get("fault_code"))
    time_map = raw.get("fault_time_and_level_map")
    entries: Dict[str, Any] = {}
    if isinstance(time_map, dict):
        entries = time_map
    elif time_map is not None:
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                source_id=source_id,
                path=f"{path}.fault_time_and_level_map",
                message="expected an object",
            )
        )
    fallback_handling, _field = first_non_empty(raw, _FAULT_LEVEL_FIELDS)
    unknown_time = 0
    out_of_range = 0
    details: List[FaultDetail] = []
    for code in codes or [None]:
        entry = entries.get(code) if code is not None else None
        handling: Optional[str] = None
        last_time: Optional[str] = None
        if isinstance(entry, dict):
            handling = entry.get("fault_level")
            raw_time = entry.get("fault_time")
            if raw_time is not None:
                # fault_time is milliseconds from DCMI events and seconds from the hang
                # detector, so the unit is decided per value instead of per field.
                unit = infer_time_unit(raw_time)
                if unit is None:
                    if positive_number(raw_time):
                        unknown_time += 1
                else:
                    when = timestamp_value(raw_time, unit, semantics="fault_time")
                    if when.utc is None:
                        out_of_range += 1
                    else:
                        last_time = when.utc
        details.append(fault_detail(code, handling or fallback_handling, last_time))
    if unknown_time:
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                source_id=source_id,
                path=f"{path}.fault_time_and_level_map",
                message="a fault timestamp is neither a plausible epoch second nor millisecond value",
            )
        )
    if out_of_range:
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                source_id=source_id,
                path=f"{path}.fault_time_and_level_map",
                message="a fault timestamp is outside the convertible datetime range",
            )
        )
    return details


def parse_device_list_grouped(
    device_list: Dict[str, Any],
    issues: List[QueryIssue],
    source_id: str = "",
    topology: Optional[Dict[str, Any]] = None,
) -> Tuple[Dict[str, List[DeviceFaultRecord]], Set[str]]:
    """Parse one node's DeviceList into one record group per resource prefix.

    Both card prefixes are read independently, so a node carrying two chip generations
    publishes two groups and the same device id may appear once per prefix (alias
    double-write). Returns ``(groups, structurally_valid_prefixes)`` where a structurally
    valid prefix decoded without a structural issue and keeps its records when the DP
    fallback is consulted.
    Fault entries without ``npu_name`` are node-level facts: they stay in their prefix
    group for fallback accounting but publish ``chip_type=None``.
    """
    groups: Dict[str, List[DeviceFaultRecord]] = {}
    structurally_valid_prefixes: Set[str] = set()
    if not isinstance(device_list, dict):
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                source_id=source_id,
                path="DeviceList",
                message="DeviceList is not an object",
            )
        )
        return groups, structurally_valid_prefixes
    superpod = _superpod(topology)
    for prefix in list_resource_prefixes(device_list):
        before = len(issues)
        records = groups.setdefault(prefix, [])
        chip = chip_type(prefix)
        fault_key = prefix + "-Fault"
        members = _members(device_list, prefix, issues, source_id)
        faulted: Set[str] = set()
        for index, item in enumerate(members.fault_items):
            if not isinstance(item, dict):
                issues.append(
                    QueryIssue(
                        code="SOURCE_INVALID",
                        source_id=source_id,
                        path=f"{fault_key}[{index}]",
                        message="fault item is not an object",
                    )
                )
                continue
            device_name = item.get("npu_name") or ""
            fault_type = item.get("fault_type")
            records.append(
                DeviceFaultRecord(
                    device_id=str(device_name) if device_name else "",
                    device_type="npu" if device_name else _device_type(fault_type),
                    chip_type=chip if device_name else None,
                    fault_type=fault_type,
                    faults=_fault_details(item, f"{fault_key}[{index}]", source_id, issues),
                    superpod=superpod,
                )
            )
            if device_name:
                faulted.add(str(device_name))
        # Abnormal devices reported by the health lists but absent from -Fault details.
        health_devices_without_fault_details = sorted(
            (members.unhealthy | members.network | members.recovering) - faulted
        )
        for name in health_devices_without_fault_details:
            records.append(
                DeviceFaultRecord(
                    device_id=name,
                    device_type="npu",
                    chip_type=chip,
                    fault_type=None,
                    faults=[],
                    superpod=superpod,
                )
            )
        if health_devices_without_fault_details and not members.fault_present:
            # Valid health data without fault details: keep the record and log it, but the
            # node's device facts are still usable, so this is not a decode failure.
            issues.append(
                QueryIssue(
                    code="SOURCE_INCOMPLETE",
                    source_id=source_id,
                    path=fault_key,
                    message="health list reports abnormal devices but no fault details are present",
                )
            )
        if not any(issue.code == "SOURCE_INVALID" for issue in issues[before:]):
            structurally_valid_prefixes.add(prefix)
    return groups, structurally_valid_prefixes


def parse_device_list(
    device_list: Dict[str, Any],
    issues: List[QueryIssue],
    source_id: str = "",
    topology: Optional[Dict[str, Any]] = None,
) -> List[DeviceFaultRecord]:
    """Flattened view of :func:`parse_device_list_grouped`, for verification/test callers.

    Production query paths use the grouped form so ``chip_type`` and the per-prefix clean
    state are preserved; this helper only exists for callers that need one flat list.
    """
    groups, _valid_prefixes = parse_device_list_grouped(device_list, issues, source_id, topology)
    return [record for records in groups.values() for record in records]


def decode_dp_deviceinfo(
    data: Dict[str, Any], source_id: str, issues: List[QueryIssue]
) -> Tuple[Dict[str, Any], Any, Dict[str, Any]]:
    """Decode ``data.DeviceInfoCfg`` into (DeviceList, UpdateTime, topology).

    The DP CM keeps ``ManuallySeparateNPU`` and ``UpgradeFaultReason`` at the top level;
    only the nested device document is handled here. ``SuperPodID/ServerIndex/RackID`` live
    beside ``DeviceInfo`` in the same JSON, so a DP fallback keeps the topology instead of
    dropping it.
    """
    cfg = decode_json_string(data.get("DeviceInfoCfg"), "DeviceInfoCfg", issues, source_id)
    if not isinstance(cfg, dict):
        return {}, None, {}
    info = cfg.get("DeviceInfo")
    if not isinstance(info, dict):
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                source_id=source_id,
                path="DeviceInfoCfg.DeviceInfo",
                message="expected an object",
            )
        )
        return {}, None, {}
    device_list = info.get("DeviceList")
    if not isinstance(device_list, dict):
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                source_id=source_id,
                path="DeviceInfoCfg.DeviceInfo.DeviceList",
                message="expected an object",
            )
        )
        return {}, None, {}
    topology = normalize_topology(
        {
            "super_pod_id": cfg.get("SuperPodID"),
            "server_index": cfg.get("ServerIndex"),
            "rack_id": cfg.get("RackID"),
        }
    )
    return device_list, info.get("UpdateTime"), topology
