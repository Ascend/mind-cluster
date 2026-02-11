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

from toolkit.core.common.json_obj import JsonObj


class InspectionInterfaceInfo(JsonObj):

    def __init__(self, device_name="", device_id="", device_sn="", interface="", interface_sn=""):
        self.device_name = device_name
        self.device_id = device_id
        self.device_sn = device_sn
        self.interface = interface
        self.interface_sn = interface_sn

    def __str__(self):
        return f"设备名：{self.device_name}，设备id：{self.device_id}，设备sn：{self.device_sn}，端口：{self.interface}，端口sn：{self.interface_sn}"


class InspectionErrorItem(JsonObj):

    def __init__(self, local_interface: InspectionInterfaceInfo,
                 peer_interface: InspectionInterfaceInfo = InspectionInterfaceInfo(), fault_desc=""):
        self.local_interface = local_interface
        self.peer_interface = peer_interface
        self.fault_desc = fault_desc

    def to_csv_dict(self):
        return {
            "A端设备名称": self.local_interface.device_name,
            "A端IP": self.local_interface.device_id,
            "A端接口": self.local_interface.interface,
            "A端SN": self.local_interface.device_sn,
            "A端光模块SN": self.local_interface.interface_sn,
            "B端设备名称": self.peer_interface.device_name,
            "B端IP": self.peer_interface.device_id,
            "B端接口": self.peer_interface.interface,
            "B端SN": self.peer_interface.device_sn,
            "B端光模块SN": self.peer_interface.interface_sn,
            "问题现象": self.fault_desc,
        }
