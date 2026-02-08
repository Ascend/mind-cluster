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

import re
from typing import List

from diag_tool.core.common.diag_enum import Customer
from diag_tool.core.common.json_obj import JsonObj
from diag_tool.core.context.register import register_inspection_check_item
from diag_tool.core.inspection.base import InspectionCheckItem
from diag_tool.core.model.cluster_info_cache import ClusterInfoCache
from diag_tool.core.model.inspection import InspectionErrorItem


class CrcErrInfo(JsonObj):

    def __init__(self, err_statistics="", crc_threshold="", crc_interval="", if_name=""):
        self.err_statistics = err_statistics
        self.crc_threshold = crc_threshold
        self.crc_interval = crc_interval
        self.if_name = if_name


@register_inspection_check_item
class CrcRisingCheckItem(InspectionCheckItem):
    _ERR_CODE = 0x081300bc
    _CRC_ERR_INFO_PATTERN = re.compile(r"hwIfMonitorCrcErrorStatistics=(?P<err_statistics>\d+), *"
                                       r"hwIfMonitorCrcErrorThreshold=(?P<crc_threshold>\d+), *"
                                       r"hwIfMonitorCrcErrorInterval=(?P<crc_interval>\d+), *"
                                       r"(hwIfMonitorName|EthPhysicalName|InterfaceName)=(?P<if_name>[^ ,)]+)")
    _ERR_DESC_TEMPLATE = "端口{} CRC快速增长告警统计次数{}, 阈值{}"

    def __init__(self, cluster_info: ClusterInfoCache, customer: Customer):
        super().__init__(cluster_info, customer)

    def check(self) -> List[InspectionErrorItem]:
        result = []
        for swi_info in self.cluster_info.swis_info.values():
            interface_full_infos = swi_info.interface_full_infos
            for alarm_info in swi_info.active_alarm_info:
                if alarm_info.alarm_id_int != self._ERR_CODE:
                    continue
                search = self._CRC_ERR_INFO_PATTERN.search(alarm_info.description)
                if not search:
                    continue
                crc_err_info = CrcErrInfo.from_dict(search.groupdict())
                if_info = interface_full_infos.get(crc_err_info.if_name)
                if not if_info:
                    continue
                local_if_info = if_info.get_inspection_interface_info()
                fault_desc = self._ERR_DESC_TEMPLATE.format(crc_err_info.if_name, crc_err_info.err_statistics,
                                                            crc_err_info.crc_threshold)
                err_item = InspectionErrorItem(local_interface=local_if_info, fault_desc=fault_desc)
                _, peer_interface_info = self.cluster_info.find_peer_swi_interface_info_by_if_info(if_info)
                if peer_interface_info:
                    err_item.peer_interface = peer_interface_info.get_inspection_interface_info()
                result.append(err_item)

        return result
