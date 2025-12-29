#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
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
from dataclasses import dataclass
from typing import List, Dict

from ascend_fd.pkg.diag.message import NoteMsg
from ascend_fd.pkg.diag.root_cluster.fault_description import FaultDescription
from ascend_fd.model.node_info import DeviceInfo, FaultFilterTime
from ascend_fd.utils.json_dict import JsonObj


class RCDiagResult(JsonObj):
    def __init__(self, analyze_success: bool = False,
                 fault_description: FaultDescription = None,
                 root_cause_device: List[str] = None,
                 device_link: List[DeviceInfo] = None,
                 remote_link: str = "",
                 first_error_device=None,
                 last_error_device=None,
                 note_msgs: List[NoteMsg] = None,
                 fault_filter_time: FaultFilterTime = None,
                 fault_description_list: List[FaultDescription] = None,
                 mindie_error_device: List[str] = None,
                 show_device_info: Dict = None,
                 detect_workers_devices: Dict[str, List[str]] = None):
        self.analyze_success = analyze_success
        self.fault_description = fault_description
        self.root_cause_device = root_cause_device or []
        self.device_link = device_link or []
        self.remote_link = remote_link
        self.first_error_device = first_error_device
        self.last_error_device = last_error_device
        self.note_msgs = note_msgs or []
        self.fault_filter_time = fault_filter_time
        self.fault_description_list = fault_description_list or []
        self.mindie_error_device = mindie_error_device or []
        self.show_device_info = show_device_info or {}
        self.detect_workers_devices = detect_workers_devices or {}
