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

from ascend_fd_tk.core.common.diag_enum import DeviceType
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.hccs.hccs_rp_tx_analyzer import HCCSCommonAnalyzer
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import DiagResult, Domain
from ascend_fd_tk.core.model.switch import SwitchInfo
from ascend_fd_tk.utils.logger import DIAG_LOGGER
from ascend_fd_tk.core.model.hccs import ProxyTimeoutStatis


@register_analyzer
class HCCSAnalyzer(HCCSCommonAnalyzer):

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self.data_init()

    @staticmethod
    def filter_timeout_interface(swi_info: SwitchInfo) -> List[ProxyTimeoutStatis]:
        rx_timeout_interfaces = []
        for proxy_timeout in swi_info.hccs_info.proxy_timeout_statis:
            if not proxy_timeout.is_rx_timeout_happend():
                continue
            rx_timeout_interfaces.append(proxy_timeout)
        return rx_timeout_interfaces

    def rx_timeout_diag(self, rx_timeout_interfaces: List[ProxyTimeoutStatis], swi_info: SwitchInfo):
        """
        检查rp_rx、lp_rx超时
        :param rx_timeout_interfaces: 发生rp_rx、lp_rx超时的接口
        :param swi_info: 交换机信息
        """
        diag_results = []
        swi_server_info = self.chassis_mappings.find_mapping_by_l1_swi_ip(swi_info.swi_id)
        if not swi_server_info or not swi_server_info.server_super_pod_id:
            DIAG_LOGGER.warning(f"交换机[{swi_info.swi_id}]与服务器映射关系缺失")
            return diag_results

        swi_lcne_infos = self.lcne_infos.get(swi_server_info.server_super_pod_id)
        if not swi_lcne_infos:
            DIAG_LOGGER.warning(f"交换机[{swi_info.swi_id}]端口信息采集缺失")
            return diag_results
        for interface in rx_timeout_interfaces:
            lcne_info = swi_lcne_infos.get(interface.interface)
            if not lcne_info:
                continue
            domain = [
                Domain(DeviceType.SWITCH.value, swi_info.swi_id),
                Domain(DeviceType.SWI_PORT.value, str(lcne_info))
            ]
            if self.check_long_link_down(swi_info.date_time, lcne_info):
                diag_results.append(DiagResult(domain, "交换机端口长期down", "排查交换机端口link状态信息"))
                continue
            if self.check_link_up_down():
                diag_results.append(DiagResult(domain, "交换机端口闪断", "排查交换机端口link状态信息"))
                continue
            if lcne_info.is_lane_error():
                diag_results.append(DiagResult(domain, "交换机链路降lane", "排查交换机链路降lane"))
                continue
            # 设备异常
            domain = [
                Domain(DeviceType.SERVER.value, swi_server_info.server_ip),
                Domain(DeviceType.XPU.value, str(lcne_info))
            ]
            diag_results.append(DiagResult(domain, "xpu设备异常", "排查对应端口所连CPU/NPU异常"))
        return diag_results

    def analyse(self) -> List[DiagResult]:
        if not self.swis_info:
            return []
        diag_results = []
        for swi_info in self.swis_info.values():
            rx_timeout_interfaces = self.filter_timeout_interface(swi_info)
            for rx_timeout_interface in rx_timeout_interfaces:
                domain = [
                    Domain(DeviceType.SWITCH.value, swi_info.swi_id),
                    Domain(DeviceType.SWI_PORT.value, rx_timeout_interface.interface)
                ]
                fault_info = f"HCCS RX超时，超时次数：{rx_timeout_interface.rp_rx + rx_timeout_interface.lp_tx}"
                suggestion = "交换机端口长期down、端口闪断、链路降lane或者对应端口所连XPU异常"
                diag_results.append(DiagResult(domain, fault_info, suggestion))
            diag_results.extend(self.rx_timeout_diag(rx_timeout_interfaces, swi_info))
        return diag_results
