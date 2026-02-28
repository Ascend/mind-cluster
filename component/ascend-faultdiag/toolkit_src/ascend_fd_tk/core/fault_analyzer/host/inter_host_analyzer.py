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

from ascend_fd_tk.core.common import diag_enum
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import DiagResult, Domain
from ascend_fd_tk.core.fault_analyzer.optical_fault_check import OpticalFaultChecker
from ascend_fd_tk.core.model.host import NpuChipInfo
from ascend_fd_tk.core.model.optical_module import OpticalModuleInfo
from ascend_fd_tk.utils import logger

_DIAG_LOGGER = logger.DIAG_LOGGER


@register_analyzer
class InterHostFaultAnalyzer(Analyzer):

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self.switches_info_by_name = self.init_switches_info()
        self.fault_check = OpticalFaultChecker(cluster_info.get_threshold())

    def init_switches_info(self):
        # 获取交换机名称和交换机信息映射关系
        switches_info_by_name = {}
        for switch_info in self.cluster_info.swis_info.values():
            switches_info_by_name.update({switch_info.name: switch_info})
        return switches_info_by_name

    def analyse(self) -> List[DiagResult]:
        diag_results = []
        for host_info in self.cluster_info.hosts_info.values():
            if not host_info.npu_chip_info:
                continue
            for _, chip_info in host_info.npu_chip_info.items():
                if not chip_info.hccn_optical_info:
                    continue
                # 本端光模块信息
                optical_module_info = chip_info.get_optical_module_info()
                local_domain = [Domain(diag_enum.DeviceType.SERVER, host_info.host_id),
                                Domain(diag_enum.DeviceType.NPU, chip_info.npu_id),
                                Domain(diag_enum.DeviceType.CHIP, chip_info.chip_phy_id)]
                # 对端光模块信息
                remote_optical_module_info, remote_domain = self.get_remote_info(host_info.host_id, chip_info)
                domain_list = local_domain + remote_domain
                if not remote_optical_module_info:
                    # 未找到对端信息，只分析本端信息
                    diag_results.extend(self.fault_analyze_single_ended(optical_module_info, domain_list))
                    continue
                diag_results.extend(self.inter_host_analyzer(
                    optical_module_info, remote_optical_module_info, domain_list))
        return diag_results

    def get_remote_info(self, host_id: str, chip_info: NpuChipInfo):
        lldp_info = chip_info.hccn_lldp_info
        switch_name = lldp_info.system_name_tlv
        port_name = lldp_info.port_id_tlv
        if not lldp_info or not switch_name or not port_name:
            _DIAG_LOGGER.warning(
                f"未收集到{host_id}, npu_id: {chip_info.npu_id}, chip_id: {chip_info.chip_id}的对端信息")
            return None, []
        remote_domain = [Domain(diag_enum.DeviceType.ROCE_SWITCH, switch_name),
                         Domain(diag_enum.DeviceType.SWI_PORT, port_name)]
        remote_switch = self.switches_info_by_name.get(switch_name)
        if not remote_switch:
            _DIAG_LOGGER.warning(f"未收集到对端交换机[{switch_name}]信息")
            return None, remote_domain
        interface_full_info = remote_switch.interface_full_infos.get(port_name)
        if not interface_full_info or not interface_full_info.swi_optical_model:
            _DIAG_LOGGER.warning(f"未收集到对端交换机[{switch_name}]端口[{port_name}]的光模块信息")
            return None, remote_domain
        return interface_full_info.get_optical_module_info(), remote_domain

    def inter_host_analyzer(self, local_info: OpticalModuleInfo, remote_info: OpticalModuleInfo,
                            domain_list: List[Domain]) -> List[DiagResult]:
        res_list = []
        res_list.extend(self.fault_check.power_analyze(local_info, remote_info, domain_list))
        res_list.extend(self.fault_check.snr_analyze(local_info, remote_info, domain_list))
        res_list.extend(self.fault_check.bias_analyze(local_info, remote_info, domain_list))
        return res_list

    def fault_analyze_single_ended(self, optical_module_info: OpticalModuleInfo, domain_list: List[Domain]):
        res_list = []
        res_list.extend(self.fault_check.power_analyze_single_ended(optical_module_info, domain_list))
        res_list.extend(self.fault_check.snr_analyze_single_ended(optical_module_info, domain_list))
        res_list.extend(self.fault_check.bias_analyze_single_ended(optical_module_info, domain_list))
        return res_list
