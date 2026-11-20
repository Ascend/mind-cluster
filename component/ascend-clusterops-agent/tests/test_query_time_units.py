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

"""Timestamp-unit, topology-sentinel and DPU node-event regressions for the query tools."""

import json

from agent_core.datasources.clusterd.dpu_summary import parse_dpu_cm
from agent_core.datasources.clusterd.switch_summary import parse_switch_faults
from agent_core.datasources.device_list import parse_device_list, decode_dp_deviceinfo
from agent_core.datasources.public_faults import parse_public_faults
from agent_core.datasources.utils import (
    infer_time_unit,
    normalize_topology,
    timestamp_value,
    timestamp_value_inferred,
)

NPU = "huawei.com/Ascend910"


def _device_list(faults, names="Ascend910-0,Ascend910-1"):
    return {
        NPU: names,
        NPU + "-Fault": json.dumps(faults),
    }


def test_infer_time_unit_uses_magnitude():
    assert infer_time_unit(1793325217) == "s"
    assert infer_time_unit("1793325217") == "s"
    assert infer_time_unit(1793325217000) == "ms"
    assert infer_time_unit(0) is None
    assert infer_time_unit(-1) is None
    assert infer_time_unit("not-a-time") is None
    assert infer_time_unit(True) is None


def test_millisecond_timestamps_render_three_digits():
    assert timestamp_value(1793325217656, "ms", "fault_time").utc == "2026-10-30T01:53:37.656Z"
    assert timestamp_value(1793325217, "s", "fault_time").utc == "2026-10-30T01:53:37Z"


def test_fault_time_seconds_and_milliseconds_are_both_readable():
    issues = []
    devices = parse_device_list(
        _device_list(
            [
                {
                    "npu_name": "Ascend910-0",
                    "fault_type": "CardUnhealthy",
                    "fault_code": "200001002",
                    "fault_handling": "NotHandleFault",
                    # hang-detection faults are written in seconds
                    "fault_time_and_level_map": {"200001002": {"fault_time": 1793325217}},
                },
                {
                    "npu_name": "Ascend910-1",
                    "fault_type": "CardUnhealthy",
                    "fault_code": "80E01801",
                    "fault_handling": "SeparateNPU",
                    # DCMI hardware faults are written in milliseconds
                    "fault_time_and_level_map": {"80E01801": {"fault_time": 1793325217000}},
                },
            ]
        ),
        issues,
        "s1",
    )

    by_name = {device.device_id: device for device in devices}
    assert by_name["Ascend910-0"].faults[0].last_time == "2026-10-30T01:53:37Z"
    assert by_name["Ascend910-1"].faults[0].last_time == "2026-10-30T01:53:37Z"
    assert not [issue for issue in issues if issue.code.startswith("TIME_")]


def test_each_fault_code_keeps_its_own_normalized_time():
    issues = []
    devices = parse_device_list(
        _device_list(
            [
                {
                    "npu_name": "Ascend910-0",
                    "fault_type": "CardUnhealthy",
                    "fault_code": "1,2",
                    "fault_handling": "SeparateNPU",
                    # code 1 is earlier in absolute time but its raw value is 1000x larger
                    "fault_time_and_level_map": {
                        "1": {"fault_time": 1793325217000},  # 2026-10-30T01:53:37Z
                        "2": {"fault_time": 1793325218},  # 2026-10-30T01:53:38Z (later)
                    },
                }
            ]
        ),
        issues,
        "s1",
    )

    details = devices[0].faults
    assert [(detail.fault_code, detail.last_time) for detail in details] == [
        ("1", "2026-10-30T01:53:37Z"),
        ("2", "2026-10-30T01:53:38Z"),
    ]


def test_public_fault_seconds_no_longer_renders_1970():
    issues = []
    records = parse_public_faults(
        {
            "PublicFaults": json.dumps(
                {"worker-a": [{"devIds": [0], "resource": "npu", "faultCode": "8", "faultTime": 1793325217}]}
            )
        },
        None,
        "s1",
        issues,
    )

    record = records["worker-a"][0]
    assert record.last_time == "2026-10-30T01:53:37Z"
    assert not [issue for issue in issues if issue.code == "SOURCE_INVALID"]


def test_public_fault_milliseconds_is_readable_and_flagged():
    issues = []
    records = parse_public_faults(
        {
            "PublicFaults": json.dumps(
                {"worker-a": [{"devIds": [0], "resource": "npu", "faultCode": "8", "faultTime": 1793325217000}]}
            )
        },
        None,
        "s1",
        issues,
    )

    assert records["worker-a"][0].last_time == "2026-10-30T01:53:37Z"
    assert any(issue.code == "SOURCE_INVALID" for issue in issues)


def test_unconvertible_timestamp_warns_and_keeps_raw_only():
    issues = []
    value = timestamp_value_inferred(10**20, "ms", "fault_time", issues, "path", "s1")

    assert value.raw == 10**20
    assert value.utc is None
    assert any(issue.code == "SOURCE_INVALID" for issue in issues)


def test_topology_sentinels_are_dropped_but_zero_is_kept():
    assert normalize_topology({"super_pod_id": -1, "server_index": 0, "rack_id": 0}) == {}
    assert normalize_topology({"super_pod_id": 0, "server_index": 0, "rack_id": 0}) == {
        "super_pod_id": 0,
        "server_index": 0,
        "rack_id": 0,
    }
    assert normalize_topology({"super_pod_id": 1, "server_index": 1023, "rack_id": -1}) == {"super_pod_id": 1}
    assert normalize_topology({"super_pod_id": 2, "server_index": "3", "rack_id": "1"}) == {
        "super_pod_id": 2,
        "server_index": 3,
        "rack_id": 1,
    }


def test_dp_fallback_keeps_topology():
    data = {
        "DeviceInfoCfg": json.dumps(
            {
                "DeviceInfo": {"DeviceList": {NPU: "Ascend910-0"}, "UpdateTime": 1793325340},
                "SuperPodID": 3,
                "ServerIndex": 0,
                "RackID": 0,
            }
        )
    }
    issues = []
    device_list, update_time, topology = decode_dp_deviceinfo(data, "s1", issues)

    assert device_list[NPU] == "Ascend910-0"
    assert update_time == 1793325340
    assert topology == {"super_pod_id": 3, "server_index": 0, "rack_id": 0}
    assert not issues


def test_dpu_node_event_is_emitted_with_node_scope():
    decoded = {
        "dpuinfo-worker-a": {
            "DPUInfo": {
                "DPUList": [],
                "NodeEvent": {
                    "NodeName": "worker-a",
                    "FaultList": [
                        {
                            "FaultCode": "card-drop",
                            "Time": 1793325217000,
                            "Description": "drop",
                            "FaultLevel": "Warning",
                        }
                    ],
                },
            },
            "UpdateTime": 1793325340000,
        }
    }
    issues = []
    faults = parse_dpu_cm(
        [{"name": "cluster-info-dpu-0", "data": {"cluster-info-dpu-0": json.dumps(decoded)}}], issues, "s1"
    )

    assert len(faults["worker-a"]) == 1
    fault = faults["worker-a"][0]
    assert (fault.device_id, fault.fault_code) == ("", "card-drop")
    assert "fault_type" not in fault.model_dump()
    assert fault.last_time == "2026-10-30T01:53:37Z"
    assert fault.affected_npu is None


def test_dpu_device_fault_time_is_milliseconds():
    decoded = {
        "dpuinfo-worker-a": {
            "DPUInfo": {
                "DPUList": [
                    {
                        "HcaName": "hca-0",
                        "FaultList": [
                            {
                                "FaultCode": "8",
                                "Time": 1793325217000,
                                "Description": "link down",
                                "FaultLevel": "Warning",
                            }
                        ],
                        "AffectedNPU": [0],
                    }
                ]
            },
            "UpdateTime": 1793325340000,
        }
    }
    issues = []
    faults = parse_dpu_cm(
        [{"name": "cluster-info-dpu-0", "data": {"cluster-info-dpu-0": json.dumps(decoded)}}], issues, "s1"
    )

    record = faults["worker-a"][0]
    assert record.device_id == "dpu:hca-0"
    assert record.affected_npu == ["0"]
    assert record.last_time == "2026-10-30T01:53:37Z"


def test_switch_alarm_time_is_milliseconds_and_falls_back_to_update_time():
    issues = []
    with_alarm = parse_switch_faults(
        "worker-a",
        {
            "codes": [{"AssembledFaultCode": "0x1", "Assertion": 1, "AlarmRaisedTime": 1793325217000}],
            "update_time": 1793325340,
        },
        issues,
    )
    assert with_alarm[0].last_time == "2026-10-30T01:53:37Z"

    without_alarm = parse_switch_faults(
        "worker-a",
        {"codes": [{"AssembledFaultCode": "0x2", "Assertion": 1}], "update_time": 1793325340},
        [],
    )
    assert without_alarm[0].last_time == "2026-10-30T01:55:40Z"
