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
from typing import List, Dict, Set

from ascend_fd_tk.core.common.diag_enum import DeviceType
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import DiagResult, Domain
from ascend_fd_tk.core.fault_analyzer.optical_fault_check import OpticalFaultChecker
from ascend_fd_tk.core.model.optical_module import OpticalModuleInfo
from ascend_fd_tk.core.model.switch import SwitchInfo, DeviceInterface
from ascend_fd_tk.utils import logger

_DIAG_LOGGER = logger.DIAG_LOGGER


class AbnormalInfo:
    def __init__(self, switch_name: str, interface_name: str, rx_power_alarm: List[str] = None,
                 tx_power_alarm: List[str] = None, rx_power_warn: List[str] = None,
                 tx_power_warn: List[str] = None, snr: List[str] = None, bias_warn: List[str] = None,
                 bias_alarm: List[str] = None):
        self.switch_name = switch_name
        self.interface_name = interface_name
        self.rx_power_alarm = rx_power_alarm or []
        self.tx_power_alarm = tx_power_alarm or []
        self.rx_power_warn = rx_power_warn or []
        self.tx_power_warn = tx_power_warn or []
        self.snr = snr or []
        self.bias_alarm = bias_alarm or []
        self.bias_warn = bias_warn or []


@register_analyzer
class SwitchAnalyzer(Analyzer):
    _MAC_DIR_REGEX = re.compile(r"^[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}")

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self.switches_info_by_name = self.init_switches_info()
        self.analyzed_switch_names: Set[str] = set()
        self.fault_check = OpticalFaultChecker(cluster_info.get_threshold())

    def init_switches_info(self):
        # 获取交换机名称和交换机信息映射关系
        switches_info_by_name = {}
        for switch_info in self.cluster_info.swis_info.values():
            switches_info_by_name.update({switch_info.name: switch_info})
        return switches_info_by_name

    def analyse(self) -> List[DiagResult]:
        diag_results = []
        # 交换机端口间光模块故障分析
        for switch_info in self.cluster_info.swis_info.values():
            diag_results.extend(self.inter_switch_fault_analyze(switch_info))
        return diag_results

    def inter_switch_fault_analyze(self, switch_info: SwitchInfo):
        # 获取交换机端口名称和对端交换机信息映射关系
        interface_mapping_by_name = {}
        for mapping in switch_info.interface_mapping:
            interface_mapping_by_name.update({mapping.local_interface_name: mapping.remote_device_interface})
        res_list = []
        for interface, full_info in switch_info.interface_full_infos.items():
            # 本端光模块信息
            analyzed_tag = f"{switch_info.name}:{interface}"
            if analyzed_tag in self.analyzed_switch_names:
                continue
            optical_module_info = full_info.get_optical_module_info()
            if not optical_module_info:
                continue
            self.analyzed_switch_names.add(analyzed_tag)
            local_domain = [Domain(DeviceType.SWITCH, switch_info.name), Domain(DeviceType.SWI_PORT, interface)]
            # 获取对端信息（对端交换机信息+对端交换机端口光模块信息）
            remote_optical_module_info, remote_domain = self.get_remote_info(interface_mapping_by_name, interface,
                                                                             switch_info.name)
            if not remote_optical_module_info and remote_domain is None:
                # 与host侧相连或者已经分析
                continue
            domain_list = local_domain + remote_domain
            if remote_optical_module_info and remote_domain:
                res_list.extend(
                    self.fault_analyze_double_ended(optical_module_info, remote_optical_module_info, domain_list)
                )
                continue
            res_list.extend(self.fault_analyze_single_ended(optical_module_info, domain_list))
        return res_list

    def get_remote_info(self, interface_mapping_by_name: Dict[str, DeviceInterface], local_interface_name: str,
                        switch_name: str):
        remove_device = interface_mapping_by_name.get(local_interface_name)
        if not remove_device:
            _DIAG_LOGGER.warning(f"未收集到交换机[{switch_name}]端口[{local_interface_name}]的对端信息")
            return None, []
        remote_domain = [Domain(DeviceType.SWITCH, remove_device.device_name),
                         Domain(DeviceType.SWI_PORT, remove_device.interface)]
        if self._MAC_DIR_REGEX.match(remove_device.interface):
            # 与host侧相连
            return None, None
        remote_analyzed_tag = f"{remove_device.device_name}:{remove_device.interface}"
        if remote_analyzed_tag in self.analyzed_switch_names:
            return None, None
        remote_switch = self.switches_info_by_name.get(remove_device.device_name)
        if not remote_switch:
            # 有对端信息但未找到对端交换机，日志警告
            _DIAG_LOGGER.warning(f"未收集到对端交换机[{remove_device.device_name}]信息")
            return None, remote_domain
        interface_full_info = remote_switch.interface_full_infos.get(remove_device.interface)
        remote_optical_module_info = interface_full_info.get_optical_module_info()
        if not remote_optical_module_info:
            _DIAG_LOGGER.warning(
                f"未收集到对端交换机[{remove_device.device_name}]端口[{remove_device.interface}]的光模块信息"
            )
        return remote_optical_module_info, remote_domain

    def fault_analyze_single_ended(self, optical_module_info: OpticalModuleInfo, domain_list: List[Domain]):
        res_list = []
        res_list.extend(self.fault_check.power_analyze_single_ended(optical_module_info, domain_list))
        res_list.extend(self.fault_check.snr_analyze_single_ended(optical_module_info, domain_list))
        res_list.extend(self.fault_check.bias_analyze_single_ended(optical_module_info, domain_list))
        return res_list

    def fault_analyze_double_ended(self, local_info: OpticalModuleInfo, remote_info: OpticalModuleInfo,
                                   domain_list: List[Domain]) -> List[DiagResult]:
        res_list = []
        # power分析
        res_list.extend(self.fault_check.power_analyze(local_info, remote_info, domain_list))
        # snr分析
        res_list.extend(self.fault_check.snr_analyze(local_info, remote_info, domain_list))
        # 电流分析
        res_list.extend(self.fault_check.bias_analyze(local_info, remote_info, domain_list))
        return res_list
