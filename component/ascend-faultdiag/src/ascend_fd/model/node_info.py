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
from typing import List

from ascend_fd.utils.json_dict import JsonObj


class DeviceInfo(JsonObj):
    def __init__(self):
        self.device_id = None
        self.phy_device_id = None
        self.logic_device_id = None
        self.device_ip = None
        self.pid = None

    def multi_init(self, device_id, logic_device_id, phy_device_id, pid):
        self.device_id = device_id
        self.logic_device_id = logic_device_id
        self.phy_device_id = phy_device_id
        self.pid = pid


class SingleServerInfo(JsonObj):
    def __init__(self, container_ip):
        self.container_ip = container_ip


class MindieClusterInfo(JsonObj):
    def __init__(self, server_count, server_list: List[SingleServerInfo]):
        self.server_count = server_count
        self.server_list = server_list

    def trans_to_map(self):
        ip_list = []
        for server_info in self.server_list:
            ip_list.append(server_info.container_ip)
        ip_list.sort()
        return {"-".join(ip_list): ip_list}


class FaultFilterTime(JsonObj):
    def __init__(self, start_train_time: str = "", end_train_time: str = ""):
        self.start_train_time = start_train_time
        self.end_train_time = end_train_time
