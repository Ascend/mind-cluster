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

from typing import Dict, List

from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.utils import helpers


# 光模块环回信息
class OpticalModuleLoopbackInfo(JsonObj):

    def __init__(self, npu_id="", chip_id="", chip_phy_id="", loopback_capabilities="0",
                 loopback_controls: List[str] = None):
        self.npu_id = npu_id
        self.chip_id = chip_id
        self.chip_phy_id = chip_phy_id
        self.loopback_capabilities = helpers.parse_hex(loopback_capabilities)
        self.loopback_controls: List[int] = self._parse_to_int_list(loopback_controls or [])

    @staticmethod
    def _parse_to_int_list(origin_loopback_controls: List[str]) -> List[int]:
        if not origin_loopback_controls:
            return []
        result = [helpers.parse_hex(item) for item in origin_loopback_controls]
        return result


class HostOpticalModuleLoopbackInfo(JsonObj):

    def __init__(self, host_ip="", optical_module_info: Dict[str, OpticalModuleLoopbackInfo] = None):
        self.host_ip = host_ip
        self.optical_module_info = optical_module_info
