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

"""测试用的工具适配器替身：清单与调用换成演示数据，不连集群（产品代码不引用它）。

`DemoAdapter` 继承 `ToolAdapter`，信封与真实返回模型同形（节点类工具记录挂在
`data.node{<节点名>}` 下）。
"""

from __future__ import annotations

from typing import Any, Dict, List, Optional, Tuple

from agent_core.adapters.tool_adapter import ToolAdapter, ToolOutcome

TOOL_DESCRIPTORS: List[Dict[str, Any]] = [
    {
        "name": "faulty_device",
        "description": "List faulty NPU devices from ClusterD summaries with per-node DP fallback, plus node health.",
        "input_schema": {
            "type": "object",
            "additionalProperties": False,
            "properties": {"node": {"type": "array", "items": {"type": "string"}, "default": None}},
        },
    },
    {
        "name": "isolated_device",
        "description": "List ClusterD-confirmed isolated devices per node with fault level, fault code and last separate time.",
        "input_schema": {
            "type": "object",
            "additionalProperties": False,
            "properties": {"node": {"type": "array", "items": {"type": "string"}, "default": None}},
        },
    },
    {
        "name": "cluster_resource",
        "description": "Report per-node CPU cores and per-chip-type card accounting (capacity, allocatable, used, idle) with independent fault/isolated counts.",
        "input_schema": {
            "type": "object",
            "additionalProperties": False,
            "properties": {"node": {"type": "array", "items": {"type": "string"}, "default": None}},
        },
    },
    {
        "name": "task_status",
        "description": "Report a job's authoritative status from its AscendJob/Volcano Job object plus its live Pods.",
        "input_schema": {
            "type": "object",
            "additionalProperties": False,
            "properties": {
                "job": {"type": "string"},
                "namespace": {"type": "string", "default": "default"},
            },
            "required": ["job"],
        },
    },
    {
        "name": "diagnose",
        "description": (
            "Collect, clean and diagnose a task's logs, returning the deterministic fault report; "
            "summarize=true also applies the legacy optional model summary."
        ),
        "input_schema": {
            "type": "object",
            "additionalProperties": False,
            "properties": {
                "job": {"type": "string"},
                "namespace": {"type": "string", "default": "default"},
                "refresh": {"type": "boolean", "default": False},
                "summarize": {"type": "boolean", "default": False},
            },
            "required": ["job"],
        },
    },
]

FAULT_DATA: Dict[str, Any] = {
    "node": {
        "node-1": {
            "devices": [
                {
                    "device_id": "3",
                    "device_type": "npu",
                    "fault_type": "CardUnhealthy",
                    "faults": [
                        {
                            "fault_code": "8087276281234567890",
                            "fault_handling": "RestartBusiness",
                            "last_time": "2026-11-01 10:00:00",
                        }
                    ],
                    "superpod": {"super_pod_id": "1", "rack_id": "rack-1", "server_index": 2},
                }
            ],
            "switch": [],
            "dpu": [],
            "public_faults": [],
            "node_health": {"node_status": "Ready"},
        },
        "node-2": {
            "devices": [
                {
                    "device_id": "5",
                    "device_type": "npu",
                    "fault_type": "CardUnhealthy",
                    "faults": [
                        {
                            "fault_code": "8087276281234567891",
                            "fault_handling": "PreSeparateNPU",
                            "last_time": "2026-11-02 11:30:00",
                        }
                    ],
                    "superpod": None,
                }
            ],
            "switch": [{"device_id": "switch-1", "fault_code": "FC-NET"}],
            "dpu": [],
            "public_faults": [],
            "node_health": {"node_status": "NotReady"},
        },
    }
}

ISOLATION_DATA: Dict[str, Any] = {
    "node": {
        "node-1": {
            "devices": [
                {
                    "device_id": "3",
                    # the real producer emits comma-joined fault levels, not a normalized type
                    "fault_level": "ManuallySeparateNPU",
                    "fault_code": "8087276281234567890",
                    "last_separate_time": "2026-11-01 10:05:00",
                }
            ]
        }
    }
}

RESOURCE_DATA: Dict[str, Any] = {
    "node": {
        "node-1": {
            "cpu": {"capacity_cores": 128.0, "allocatable_cores": 120.0, "used_cores": 3.5, "idle_cores": 116.5},
            "Ascend910": {"capacity": 8, "allocatable": 8, "used": 3, "idle": 5, "fault": 0, "isolated": 0},
        },
        "node-2": {
            "cpu": {"capacity_cores": 96.0, "allocatable_cores": 90.0, "used_cores": 2.0, "idle_cores": 88.0},
            "Ascend910": {"capacity": 8, "allocatable": 7, "used": 3, "idle": 4, "fault": 1, "isolated": 1},
        },
    }
}

JOB_DATA: Dict[str, Any] = {
    "namespace": "default",
    "job_name": "train-stub",
    "job_type": "ACJob",
    "status": "Running",
    "pods": [
        {"pod_name": "train-stub-worker-0", "status": "Running"},
        {"pod_name": "train-stub-worker-1", "status": "CrashLoopBackOff"},
    ],
}

DIAG_DATA: Dict[str, Any] = {
    "diag_report": {"card": {}, "node": {}},
    "pods": [{"name": "train-stub-worker-0"}],
    "incomplete_nodes": [],
    "empty_nodes": [],
    "cached": False,
    "final_text": "",
    "error": "",
    "llm_error": "",
}

_DEMO_DATA = {
    "faulty_device": FAULT_DATA,
    "isolated_device": ISOLATION_DATA,
    "cluster_resource": RESOURCE_DATA,
    "task_status": JOB_DATA,
    "diagnose": DIAG_DATA,
}


def demo_envelope(
    name: str,
    data: Optional[Dict[str, Any]] = None,
    status: str = "ok",
    errors: Optional[List[Dict[str, Any]]] = None,
    request_id: str = "req-demo",
    elapsed_ms: int = 7,
) -> Dict[str, Any]:
    """Build one tool envelope shaped like the real result models."""
    import copy

    payload = copy.deepcopy(_DEMO_DATA[name] if data is None else data)
    return {
        "request_id": request_id,
        "tool": name,
        "status": status,
        "elapsed_ms": elapsed_ms,
        "data": payload,
        "errors": list(errors or []),
    }


class DemoAdapter(ToolAdapter):
    """适配层替身：同一接口，演示载荷，不连集群。"""

    def __init__(self, descriptors: Optional[List[Dict[str, Any]]] = None) -> None:
        super().__init__()
        self._descriptors = descriptors if descriptors is not None else TOOL_DESCRIPTORS
        self.calls: List[Tuple[str, Dict[str, Any]]] = []

    # -- discovery ---------------------------------------------------------- #
    def list_tools(self) -> List[Dict[str, Any]]:
        return [dict(item) for item in self._descriptors]

    def names(self) -> List[str]:
        return [item["name"] for item in self._descriptors]

    # -- execution ---------------------------------------------------------- #
    def invoke(self, name: str, args: Dict[str, Any]) -> ToolOutcome:
        self.calls.append((name, dict(args or {})))
        if name not in {item["name"] for item in self._descriptors}:
            return ToolOutcome(
                ok=False, result=None, error={"code": "TOOL_NOT_FOUND", "message": f"unknown tool: {name!r}"}
            )
        return ToolOutcome(ok=True, result=demo_envelope(name), error=None)
