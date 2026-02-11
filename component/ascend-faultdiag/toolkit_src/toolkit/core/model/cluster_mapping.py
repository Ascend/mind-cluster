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

from typing import List

from toolkit.core.common.json_obj import JsonObj


class L1SwiServerMapping(JsonObj):

    def __init__(self, l1_swi_ip: str, l1_swi_name: str, server_super_pod_id: str, server_ip: str, bmc_ip: str):
        self.l1_swi_ip = l1_swi_ip
        self.l1_swi_name = l1_swi_name
        self.server_super_pod_id = server_super_pod_id
        self.server_ip = server_ip
        self.bmc_ip = bmc_ip

# 机框映射关系 l1->服务器->bmc
class ChassisMapping(JsonObj):

    def __init__(self, l1_swi_server_mappings: List[L1SwiServerMapping] = None):
        self.l1_swi_server_mappings = l1_swi_server_mappings or []

    def find_mapping_by_bmc_ip(self, bmc_ip: str) -> L1SwiServerMapping:
        for mapping in self.l1_swi_server_mappings:
            if mapping.bmc_ip == bmc_ip:
                return mapping
        return None

    def find_mapping_by_l1_swi_ip(self, l1_swi_ip: str) -> L1SwiServerMapping:
        for mapping in self.l1_swi_server_mappings:
            if mapping.l1_swi_ip == l1_swi_ip:
                return mapping
        return None

    def find_mapping_by_server_ip(self, server_ip: str) -> L1SwiServerMapping:
        for mapping in self.l1_swi_server_mappings:
            if mapping.server_ip == server_ip:
                return mapping
        return None
