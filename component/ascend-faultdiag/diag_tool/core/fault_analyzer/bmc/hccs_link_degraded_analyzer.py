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

from diag_tool.core.common import diag_enum
from diag_tool.core.common.json_obj import JsonObj
from diag_tool.core.config import port_mapping_config
from diag_tool.core.config.chip_port_range import tiancheng_cpu_port_list
from diag_tool.core.config.port_mapping_config import L1InterfacePortMapping
from diag_tool.core.context.register import register_analyzer
from diag_tool.core.fault_analyzer.base import Analyzer
from diag_tool.core.model.bmc import BmcInfo
from diag_tool.core.model.cluster_info_cache import ClusterInfoCache
from diag_tool.core.model.cluster_mapping import L1SwiServerMapping
from diag_tool.core.model.diag_result import DiagResult, Domain


class CpuBoardUbcInfo(JsonObj):

    def __init__(self, cpu_id="", ubc_id="", macro_id="", cud_board_id=""):
        self.cpu_id = cpu_id
        self.ubc_id = ubc_id
        self.macro_id = macro_id
        self.cud_board_id = cud_board_id


@register_analyzer
class HccsLinkDegradedAnalyzer(Analyzer):
    _ERROR_CODE = "0x28000049"
    _HCCS_LINK_DEGRADED_PATTERN = re.compile(
        r"CPU(?P<cpu_id>\d+) UBC(?P<ubc_id>\d+) macro(?P<macro_id>\d+) on CPU board (?P<cud_board_id>\d+)")

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)

        # 找到该cpu对端的1520端口

    @staticmethod
    def _find_cpu_peer_swi_port(cpu_board_ubc_info: CpuBoardUbcInfo) -> L1InterfacePortMapping:
        port_mapping_config_instance = port_mapping_config.get_port_mapping_config_instance()
        for mapping_info in port_mapping_config_instance.l1_interface_port_map.values():
            if mapping_info.xpu == diag_enum.XPU.CPU.value and \
                    mapping_info.xpu_id == cpu_board_ubc_info.cpu_id and \
                    mapping_info.macro == cpu_board_ubc_info.macro_id:
                return mapping_info
        return None

    def analyse(self) -> List[DiagResult]:
        result = []
        for bmc_info in self.cluster_info.bmcs_info.values():
            host_error_event_infos = self._find_hccs_link_degrade_events(bmc_info)
            chassis_mapping = self.cluster_info.get_chassis_mappings().find_mapping_by_bmc_ip(bmc_info.bmc_id)
            if not host_error_event_infos:
                continue
            result.extend(self._port_fault_analyse(chassis_mapping, host_error_event_infos))
        return result

    def _port_fault_analyse(self, chassis_mapping: L1SwiServerMapping, host_error_event_infos: list[CpuBoardUbcInfo]):
        result = []
        port_mappings = [self._find_cpu_peer_swi_port(event_info) for event_info in host_error_event_infos]
        # 端口级别的分析
        for event_info, port_mapping in zip(host_error_event_infos, port_mappings):
            if not port_mapping:
                continue
            diag_result = DiagResult(
                [
                    Domain(diag_enum.DeviceType.L1_SWITCH, chassis_mapping.l1_swi_ip),
                    Domain(diag_enum.DeviceType.SWI_PORT, port_mapping.swi_port)
                ],
                f"Cpu{event_info.cpu_id} UBC{event_info.ubc_id} macro{event_info.macro_id} "
                f"CPU board {event_info.cud_board_id}与L1端口之间发生故障",
                f"建议检查L1端口或cpu板抽屉", err_code=self._ERROR_CODE
            )
            result.append(diag_result)
        # 板级别的分析
        for swi_chip_id, cpu_port_list in enumerate(tiancheng_cpu_port_list):
            swi_chip_err_ports = set()
            cpu_port_set = set(cpu_port_list)
            for port_mapping in port_mappings:
                if port_mapping.swi_chip_id == str(swi_chip_id):
                    swi_chip_err_ports.add(port_mappings)
            if cpu_port_set == swi_chip_err_ports:
                diag_result = DiagResult(
                    [
                        Domain(diag_enum.DeviceType.L1_SWITCH, chassis_mapping.l1_swi_ip),
                        Domain(diag_enum.DeviceType.SWI_CHIP, str(swi_chip_id))
                    ],
                    f"L1交换芯片所有端口异常",
                    f"建议检查L1交换板或2个CPU抽屉", err_code=self._ERROR_CODE
                )
                result.append(diag_result)
        return result

    # 从事件列表中找到hccs_link_degrade事件
    def _find_hccs_link_degrade_events(self, bmc_info: BmcInfo) -> List[CpuBoardUbcInfo]:
        host_error_event_infos = []
        for event in bmc_info.health_events:
            if event.event_code.lower() != self._ERROR_CODE:
                continue
            search = self._HCCS_LINK_DEGRADED_PATTERN.search(event.event_description)
            if search:
                info = CpuBoardUbcInfo.from_dict(search.groupdict())
                host_error_event_infos.append(info)
        return host_error_event_infos
