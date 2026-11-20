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

"""Parser for the ClusterD switch summary: cluster-info-switch-{x}."""

from __future__ import annotations

from typing import Any, Dict, List, Optional

from agent_core.datasources.cm_documents import iter_cm_documents
from agent_core.datasources.utils import (
    decode_json_string,
    infer_time_unit,
    join_code,
    node_from_cm_name,
    timestamp_value,
)
from agent_core.query_schemas import QueryIssue, SwitchFaultRecord

SWITCH_CM_PREFIX = "mindx-dl-switchinfo-"


def parse_switch_cm(
    items: List[Dict[str, Any]], issues: List[QueryIssue], source_id: str = ""
) -> Dict[str, Dict[str, Any]]:
    """Decode cluster-info-switch-{x}; each FaultCode element is itself a JSON string.

    ``FaultLevel`` at document level is the maximum ClusterD derives from
    ``FaultTimeAndLevelMap``, so it is only a fallback: the level and time of one alarm
    must come from the map entry keyed ``<fault_code>_<chip>_<port>``.
    """
    out: Dict[str, Dict[str, Any]] = {}
    for name, source, decoded in iter_cm_documents(items, issues, source_id):
        for key, value in decoded.items():
            if not isinstance(value, dict):
                continue
            node = node_from_cm_name(str(value.get("CmName") or key), (SWITCH_CM_PREFIX,))
            raw_codes = value.get("FaultCode") or []
            codes: List[Dict[str, Any]] = []
            if not isinstance(raw_codes, list):
                issues.append(
                    QueryIssue(
                        code="SOURCE_INVALID",
                        source_id=source.source_id,
                        path=f"{name}.{key}.FaultCode",
                        message="expected an array",
                    )
                )
                raw_codes = []
            for index, code in enumerate(raw_codes):
                parsed = decode_json_string(code, f"{name}.{key}.FaultCode[{index}]", issues, source.source_id)
                if isinstance(parsed, dict):
                    codes.append(parsed)
                elif parsed is not None:
                    issues.append(
                        QueryIssue(
                            code="SOURCE_INVALID",
                            source_id=source.source_id,
                            path=f"{name}.{key}.FaultCode[{index}]",
                            message="fault code is not an object",
                        )
                    )
            out[node] = {
                "codes": codes,
                "fault_level": value.get("FaultLevel"),
                "update_time": value.get("UpdateTime"),
                "level_map": value.get("FaultTimeAndLevelMap") or {},
            }
    return out


def _level_key(raw: Dict[str, Any]) -> str:
    """ClusterD keys FaultTimeAndLevelMap as ``<AssembledFaultCode>_<chip>_<port>``."""
    return f"{raw.get('AssembledFaultCode')}_{raw.get('SwitchChipId')}_{raw.get('SwitchPortId')}"


def _level_time(value: Any) -> Any:
    """Decode one per-alarm ``fault_time``; None when the map has no usable value."""
    if value in (None, ""):
        return None
    unit = infer_time_unit(value) or "ms"
    return timestamp_value(value, unit, semantics="switch_fault_time")


def _device_id(raw: Dict[str, Any]) -> str:
    parts = [raw.get("SwitchChipId"), raw.get("SwitchPortId")]
    materialized = [str(part) for part in parts if part not in (None, "")]
    return "switch:" + ":".join(materialized) if materialized else "switch"


def parse_switch_faults(
    node: str, info: Dict[str, Any], issues: Optional[List[QueryIssue]] = None
) -> List[SwitchFaultRecord]:
    """Emit one public record per active switch alarm.

    The per-alarm entry in ``FaultTimeAndLevelMap`` is authoritative for level and time.
    The document-level ``FaultLevel`` is ClusterD's maximum over those entries, so taking
    it per record would escalate every alarm to the worst one on the node.
    """
    updated = timestamp_value(info.get("update_time"), "s", semantics="switch_update_time")
    level_map = info.get("level_map") or {}
    records: List[SwitchFaultRecord] = []
    for raw in info.get("codes") or []:
        assertion = raw.get("Assertion")
        if assertion in (0, "0"):
            continue
        entry = level_map.get(_level_key(raw)) or {}
        when = _level_time(entry.get("fault_time"))
        if when is None:
            alarms = raw.get("AlarmRaisedTime")
            if alarms:
                unit = infer_time_unit(alarms)
                when = timestamp_value(alarms, unit, semantics="alarm_raised_time")
                if unit and unit != "ms" and issues is not None:
                    issues.append(
                        QueryIssue(
                            code="SOURCE_INVALID",
                            path="AlarmRaisedTime",
                            message=f"timestamp unit inferred as {unit}, documented contract is ms",
                        )
                    )
            else:
                # The switch UpdateTime is a refresh time, not the alarm time.
                when = updated
        records.append(
            SwitchFaultRecord(
                device_id=_device_id(raw),
                fault_handling=entry.get("fault_level") or info.get("fault_level"),
                fault_code=join_code(raw.get("AssembledFaultCode")),
                last_time=when.utc,
            )
        )
    return records
