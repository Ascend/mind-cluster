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

"""Parser for the ClusterD device summary: cluster-info-device-{m}."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional

from agent_core.datasources.cm_documents import iter_cm_documents
from agent_core.datasources.device_list import list_resource_prefixes
from agent_core.datasources.sources import DP_DEVICES
from agent_core.datasources.utils import node_from_cm_name, normalize_topology
from agent_core.query_schemas import QueryIssue, SourcePointer

DP_CM_PREFIX = DP_DEVICES.name_prefix


@dataclass
class DeviceEntry:
    node: str
    device_list: Dict[str, Any] = field(default_factory=dict)
    update_time: Any = None
    topology: Dict[str, Any] = field(default_factory=dict)
    source: Optional[SourcePointer] = None
    profile: str = "spec"


def _is_device_list(mapping: Dict[str, Any]) -> bool:
    return bool(list_resource_prefixes(mapping))


def parse_cluster_device_shards(
    items: List[Dict[str, Any]], issues: List[QueryIssue], source_id: str = ""
) -> Dict[str, DeviceEntry]:
    """Decode cluster-info-device-{m}.

    Current source shape: data[<shard CM name>] = JSON string of
    {<DP CM name>: {DeviceList, UpdateTime, CmName, SuperPodID, RackID, ServerIndex}}.
    An already-expanded node -> DeviceList shape is accepted as a fallback profile.
    """
    entries: Dict[str, DeviceEntry] = {}
    for name, source, decoded in iter_cm_documents(items, issues, source_id):
        for key, value in decoded.items():
            if not isinstance(value, dict):
                issues.append(
                    QueryIssue(
                        code="SOURCE_INVALID",
                        source_id=source.source_id,
                        path=f"{name}.{key}",
                        message="node entry is not an object",
                    )
                )
                continue
            node = node_from_cm_name(str(key), (DP_CM_PREFIX,))
            cm_name = value.get("CmName") or key
            if isinstance(cm_name, str) and cm_name.startswith(DP_CM_PREFIX):
                node = node_from_cm_name(cm_name, (DP_CM_PREFIX,))
            device_list = value.get("DeviceList")
            if not isinstance(device_list, dict):
                if "DeviceList" not in value and _is_device_list(value):
                    # Expanded node -> DeviceList profile: the entry itself is the device document.
                    device_list = value
                else:
                    # A type-unusable DeviceList is an empty roster: the node still counts as
                    # read (status stays ok) and only the log records the gap.
                    present = "DeviceList" in value
                    issues.append(
                        QueryIssue(
                            code="SOURCE_INVALID",
                            source_id=source.source_id,
                            path=f"{name}.{key}.DeviceList",
                            message="DeviceList is not an object" if present else "DeviceList is missing",
                        )
                    )
                    device_list = {}
            topology = normalize_topology(
                {
                    "super_pod_id": value.get("SuperPodID"),
                    "rack_id": value.get("RackID"),
                    "server_index": value.get("ServerIndex"),
                }
            )
            entries[node] = DeviceEntry(
                node=node,
                device_list=device_list,
                update_time=value.get("UpdateTime"),
                topology={k: v for k, v in topology.items() if v is not None},
                source=source,
            )
    return entries
