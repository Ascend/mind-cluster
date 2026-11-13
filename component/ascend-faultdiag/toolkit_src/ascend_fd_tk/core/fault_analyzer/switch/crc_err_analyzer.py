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
import statistics
from typing import List

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.core.config.port_mapping_config import get_port_mapping_config_instance
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.diag_result import DiagResult, SwitchDomain


class CrcErrInfo(JsonObj):
    def __init__(self, err_statistics="", crc_threshold="", crc_interval="", if_name=""):
        self.err_statistics = err_statistics
        self.crc_threshold = crc_threshold
        self.crc_interval = crc_interval
        self.if_name = if_name


def _get_peer_xpu_desc(cluster_info, swi_info, interface: str) -> str:
    """交换机端口直连CPU/NPU时，通过端口映射关系获取对端XPU的IP与编号"""
    port_mapping = get_port_mapping_config_instance().find_port_mapping_by_name(interface)
    if not port_mapping or not port_mapping.xpu or not port_mapping.xpu_id:
        return ""
    chassis_mappings = cluster_info.get_chassis_mappings()
    swi_server_mapping = chassis_mappings and chassis_mappings.find_mapping_by_l1_swi_ip(swi_info.swi_id)
    server_ip = swi_server_mapping.server_ip if swi_server_mapping else ""
    xpu = port_mapping.xpu.upper()
    peer_desc = f"。对端{xpu}信息："
    if server_ip:
        peer_desc += f"设备id：{server_ip}，"
    peer_desc += f"{xpu.lower()} id：{port_mapping.xpu_id}"
    return peer_desc


@register_analyzer(generation=[NpuType.A3, NpuType.A5])
class CrcRisingCheckItem(Analyzer):
    _ERR_CODE = 0x081300BC
    _CRC_ERR_INFO_PATTERN = re.compile(
        r"hwIfMonitorCrcErrorStatistics=(?P<err_statistics>\d+), *"
        r"hwIfMonitorCrcErrorThreshold=(?P<crc_threshold>\d+), *"
        r"hwIfMonitorCrcErrorInterval=(?P<crc_interval>\d+), *"
        r"(hwIfMonitorName|EthPhysicalName|InterfaceName)=(?P<if_name>[^ ,)]+)"
    )
    _ERR_DESC_TEMPLATE = "端口{} CRC快速增长告警统计次数{}，阈值{}"

    def analyse(self) -> List[DiagResult]:
        result = []
        # pylint: disable=duplicate-code  # 已与同类分析器复用逻辑，忽略重复警告
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
                fault_info = self._ERR_DESC_TEMPLATE.format(
                    crc_err_info.if_name, crc_err_info.err_statistics, crc_err_info.crc_threshold
                )
                _, peer_interface_info = self.cluster_info.find_peer_swi_interface_info_by_if_info(if_info)
                if peer_interface_info:
                    fault_info += f"，对端设备{peer_interface_info.device_id}，对端端口{peer_interface_info.interface}"
                else:
                    fault_info += _get_peer_xpu_desc(self.cluster_info, swi_info, if_info.interface)
                result.append(
                    DiagResult(
                        domain=SwitchDomain(
                            swi_id=local_if_info.device_id,
                            slot_id=swi_info.slot_id,
                            interface=local_if_info.interface,
                        ),
                        fault_info=fault_info,
                        suggestion="端口CRC快速增长，请检查端口",
                        err_code=alarm_info.alarm_id,
                    )
                )
        return result


@register_analyzer
class CrcCountersErrorsCheckItem(Analyzer):
    SIGMA_K = 3

    def analyse(self) -> List[DiagResult]:
        result = []
        for swi_info in self.cluster_info.swis_info.values():
            interface_full_infos = swi_info.interface_full_infos
            if not interface_full_infos:
                continue
            crc_error_list = [crc_errors_info.crc_errors for crc_errors_info in swi_info.crc_counters_errors]
            if len(crc_error_list) <= 1:
                continue
            mean_val = statistics.mean(crc_error_list)
            std_val = statistics.stdev(crc_error_list)
            for crc_errors_info in swi_info.crc_counters_errors:
                if_info = interface_full_infos.get(crc_errors_info.interface_name)
                if not if_info:
                    continue
                local_if_info = if_info.get_inspection_interface_info()
                fault_info = f"端口{crc_errors_info.interface_name} CRC快速增长，统计次数{crc_errors_info.crc_errors}"
                _, peer_interface_info = self.cluster_info.find_peer_swi_interface_info_by_if_info(if_info)
                if peer_interface_info:
                    fault_info += f"，对端设备{peer_interface_info.device_id}，对端端口{peer_interface_info.interface}"
                else:
                    fault_info += _get_peer_xpu_desc(self.cluster_info, swi_info, if_info.interface)
                if abs(crc_errors_info.crc_errors - mean_val) > self.SIGMA_K * std_val:
                    result.append(
                        DiagResult(
                            domain=SwitchDomain(
                                swi_id=local_if_info.device_id,
                                slot_id=swi_info.slot_id,
                                interface=local_if_info.interface,
                            ),
                            fault_info=fault_info,
                            suggestion="端口CRC快速增长，请排查线缆或者对端设备问题",
                        )
                    )
        return result
