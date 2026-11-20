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

"""Parser for the ClusterD DPU summary: cluster-info-dpu-{n}."""

from __future__ import annotations

from typing import Any, Dict, List, Optional

from agent_core.datasources.cm_documents import iter_cm_documents
from agent_core.datasources.utils import (
    infer_time_unit,
    join_code,
    node_from_cm_name,
    timestamp_value,
    timestamp_value_inferred,
)
from agent_core.query_schemas import DpuFaultRecord, QueryIssue

DPU_CM_PREFIX = "dpuinfo-"


def _device_id(dpu: Dict[str, Any]) -> str:
    value = dpu.get("DeviceID") or dpu.get("HcaName") or dpu.get("EthName")
    return f"dpu:{value}" if value not in (None, "") else "dpu"


def _affected(dpu: Dict[str, Any]) -> List[str]:
    value = dpu.get("AffectedNPU")
    if not isinstance(value, list):
        return []
    return [str(item) for item in value if item not in (None, "")]


def _fault_records(
    faults: Any,
    path_prefix: str,
    device_id: str,
    affected: Optional[List[str]],
    fallback: Any,
    issues: List[QueryIssue],
    source_id: str,
) -> List[DpuFaultRecord]:
    """One public record per DPU fault; a missing time falls back to the CM update time."""
    records: List[DpuFaultRecord] = []
    for index, fault in enumerate(faults or []):
        if not isinstance(fault, dict):
            continue
        raw_time = fault.get("Time")
        when = fallback
        if raw_time:
            when = timestamp_value_inferred(
                raw_time, "ms", "dpu_fault_time", issues, f"{path_prefix}[{index}].Time", source_id
            )
        records.append(
            DpuFaultRecord(
                device_id=device_id,
                fault_handling=fault.get("FaultLevel"),
                fault_code=join_code(fault.get("FaultCode")),
                last_time=when.utc,
                affected_npu=affected,
            )
        )
    return records


def parse_dpu_cm(
    items: List[Dict[str, Any]], issues: List[QueryIssue], source_id: str = ""
) -> Dict[str, List[DpuFaultRecord]]:
    """Decode cluster-info-dpu-{n} into node -> public DPU records.

    Device ``FaultList`` entries carry ``affected_npu``; ``NodeEvent.FaultList`` entries are
    separate node-level records without that relation.
    """
    out: Dict[str, List[DpuFaultRecord]] = {}
    for name, source, decoded in iter_cm_documents(items, issues, source_id):
        updated = timestamp_value(
            decoded.get("UpdateTime"), infer_time_unit(decoded.get("UpdateTime")), semantics="dpu_update_time"
        )
        for key, value in decoded.items():
            if not isinstance(value, dict):
                continue
            node = node_from_cm_name(str(key), (DPU_CM_PREFIX,))
            dpu_info = value.get("DPUInfo") or {}
            if not isinstance(dpu_info, dict):
                continue
            records = out.setdefault(node, [])
            for dpu_index, dpu in enumerate(dpu_info.get("DPUList") or []):
                if not isinstance(dpu, dict):
                    continue
                records.extend(
                    _fault_records(
                        dpu.get("FaultList"),
                        f"{name}.{key}.DPUInfo.DPUList[{dpu_index}].FaultList",
                        _device_id(dpu),
                        _affected(dpu),
                        updated,
                        issues,
                        source.source_id,
                    )
                )
            node_event = _node_event(dpu_info.get("NodeEvent"))
            records.extend(
                _fault_records(
                    (node_event or {}).get("FaultList"),
                    f"{name}.{key}.DPUInfo.NodeEvent.FaultList",
                    "",
                    None,
                    updated,
                    issues,
                    source.source_id,
                )
            )
    return out


def _node_event(value: Any) -> Optional[Dict[str, Any]]:
    if isinstance(value, dict):
        return value
    if isinstance(value, list):
        return {"FaultList": value}
    return None
