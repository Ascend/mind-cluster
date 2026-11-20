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

"""Parser for the ClusterD node summary: cluster-info-node-cm (health and node faults)."""

from __future__ import annotations

from typing import Any, Dict, List

from agent_core.datasources.cm_documents import iter_cm_documents
from agent_core.datasources.faults import build_fault_details
from agent_core.datasources.utils import node_from_cm_name, split_csv, timestamp_value
from agent_core.query_schemas import DeviceFaultRecord, QueryIssue

NODE_CM_PREFIX = "mindx-dl-nodeinfo-"


def parse_node_cm(
    items: List[Dict[str, Any]], issues: List[QueryIssue], source_id: str = ""
) -> Dict[str, Dict[str, Any]]:
    """Decode cluster-info-node-cm into node -> {fault_dev_list, node_status, update_time}."""
    out: Dict[str, Dict[str, Any]] = {}
    for name, source, decoded in iter_cm_documents(items, issues, source_id):
        if "FaultDevList" in decoded or "NodeStatus" in decoded:
            node = node_from_cm_name(str(decoded.get("CmName") or name), (NODE_CM_PREFIX,))
            out[node] = {
                "fault_dev_list": decoded.get("FaultDevList") or [],
                "node_status": decoded.get("NodeStatus"),
                "update_time": decoded.get("UpdateTime"),
            }
            continue
        for key, value in decoded.items():
            if not isinstance(value, dict):
                continue
            node = node_from_cm_name(str(key), (NODE_CM_PREFIX,))
            cm_name = value.get("CmName") or key
            if isinstance(cm_name, str) and cm_name.startswith(NODE_CM_PREFIX):
                node = node_from_cm_name(cm_name, (NODE_CM_PREFIX,))
            out[node] = {
                "fault_dev_list": value.get("FaultDevList") or [],
                "node_status": value.get("NodeStatus"),
                "update_time": value.get("UpdateTime"),
            }
    return out


def parse_node_faults(node_info: Dict[str, Any]) -> List[DeviceFaultRecord]:
    """Build public node-level device records from ``FaultDevList[]``."""
    updated = timestamp_value(node_info.get("update_time"), "ms", semantics="node_update_time")
    last_time = updated.utc if updated.utc is not None else None
    records: List[DeviceFaultRecord] = []
    for raw in node_info.get("fault_dev_list") or []:
        if not isinstance(raw, dict):
            continue
        device_type = raw.get("DeviceType")
        fault_level = raw.get("FaultLevel")
        details = build_fault_details(split_csv(raw.get("FaultCode")), fault_level, last_time)
        device_id = raw.get("DeviceId")
        records.append(
            DeviceFaultRecord(
                device_id="" if device_id in (None, "") else str(device_id),
                device_type=str(device_type).strip().lower()
                if isinstance(device_type, str) and device_type.strip()
                else None,
                chip_type=None,
                fault_type=device_type,
                faults=details,
                superpod=None,
            )
        )
    return records
