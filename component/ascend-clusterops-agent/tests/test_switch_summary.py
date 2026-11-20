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
"""Switch summary: the per-alarm FaultTimeAndLevelMap wins over the document aggregate.

ClusterD derives the document-level ``FaultLevel`` as the maximum over
``FaultTimeAndLevelMap`` (see clusterd ``SwitchInfo.updateFaultLevelAndNodeStatus``), so
expanding that maximum onto every alarm would escalate each record to the worst one.
"""

import json

from agent_core.datasources.clusterd.switch_summary import parse_switch_cm, parse_switch_faults

NODE = "localhost.localdomain"
CM_NAME = "cluster-info-switch-0"
ALARM_TIME = 1794835118483
DISTINCT_TIME = 1794900000000

# Real payload shape: five alarms, three NotHandle, one SubHealthFault, one Separate.
FAULTS = [
    ("[0x08520003,na,npu,na]", 0, 16, "Separate"),
    ("[0x00f10509,132332,L2,na]", 0, 7, "NotHandle"),
    ("[0x00f10509,132332,L2,na]", 2, 27, "NotHandle"),
    ("[0x00f10509,132332,L2,na]", 6, 9, "NotHandle"),
    ("[0x00f10509,132333,npu,na]", 0, 16, "SubHealthFault"),
]


def _fault(code, chip, port):
    return json.dumps(
        {
            "EventType": 15795465,
            "AssembledFaultCode": code,
            "PeerPortDevice": 1,
            "PeerPortId": 0,
            "SwitchChipId": chip,
            "SwitchPortId": port,
            "Severity": 3,
            "Assertion": 1,
            "AlarmRaisedTime": ALARM_TIME,
        }
    )


def _document(with_map=True):
    level_map = {}
    if with_map:
        for code, chip, port, level in FAULTS:
            # Give the first alarm a distinct time so the map path is observable.
            stamp = DISTINCT_TIME if level == "Separate" else ALARM_TIME
            level_map[f"{code}_{chip}_{port}"] = {"fault_time": stamp, "fault_level": level}
    return {
        "mindx-dl-switchinfo-localhost.localdomain": {
            "FaultCode": [_fault(code, chip, port) for code, chip, port, _ in FAULTS],
            "FaultLevel": "Separate",  # document-level maximum
            "UpdateTime": 1794835141,
            "NodeStatus": "UnHealthy",
            "FaultTimeAndLevelMap": level_map,
            "CmName": "mindx-dl-switchinfo-localhost.localdomain",
        }
    }


def _records(document):
    items = [{"name": CM_NAME, "data": {CM_NAME: json.dumps(document)}}]
    issues = []
    parsed = parse_switch_cm(items, issues, "s1")
    records = []
    for _, info in parsed.items():
        records.extend(parse_switch_faults(NODE, info, issues))
    return records, issues


def test_switch_level_uses_per_fault_map_not_document_maximum():
    records, issues = _records(_document())

    assert not issues
    got = {(record.device_id, record.fault_code): record.fault_handling for record in records}
    assert got == {
        ("switch:0:16", "[0x08520003,na,npu,na]"): "Separate",
        ("switch:0:7", "[0x00f10509,132332,L2,na]"): "NotHandle",
        ("switch:2:27", "[0x00f10509,132332,L2,na]"): "NotHandle",
        ("switch:6:9", "[0x00f10509,132332,L2,na]"): "NotHandle",
        ("switch:0:16", "[0x00f10509,132333,npu,na]"): "SubHealthFault",
    }


def test_switch_time_prefers_per_fault_map_over_alarm_raised_time():
    records, _ = _records(_document())

    by_code = {record.fault_code: record for record in records if record.device_id == "switch:0:16"}
    assert by_code["[0x08520003,na,npu,na]"].last_time != by_code["[0x00f10509,132333,npu,na]"].last_time
    # Every alarm carries the same AlarmRaisedTime, so a difference proves the map value won.
    others = [record.last_time for record in records if record.fault_code == "[0x00f10509,132332,L2,na]"]
    assert len(set(others)) == 1
    assert by_code["[0x08520003,na,npu,na]"].last_time not in others


def test_switch_level_falls_back_to_document_value_without_map():
    records, issues = _records(_document(with_map=False))

    assert not issues
    assert {record.fault_handling for record in records} == {"Separate"}
    assert len(records) == len(FAULTS)
