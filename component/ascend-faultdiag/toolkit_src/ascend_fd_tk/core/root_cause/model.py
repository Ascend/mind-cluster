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
from dataclasses import dataclass, field
from typing import List, Union

from ascend_fd_tk.core.model.diag_result import HostDomain, BmcDomain, SwitchDomain


@dataclass
class HostToL1LinkData:
    """Host→L1 链路数据

    记录从主机(NPU/CPU)到L1交换板的信号链路信息，
    包括主机标识、NPU/CPU信息、L1交换板和端口信息、SNR数据及故障分析结果
    """

    host_id: str = ""
    host_name: str = ""
    room_name: str = ""
    cabinet_id: str = ""
    npu_type: str = ""
    npu_id: str = ""
    chip_phy_id: str = ""
    l1_switch_id: str = ""
    l1_switch_name: str = ""
    l1_switch_chip_id: str = ""
    l1_interface: str = ""
    l1_port_speed: str = ""
    l1_switch_chip_phy_id: str = ""
    l1_hilink_snr: str = ""
    link_status: str = ""
    triggered_rules: List[str] = field(default_factory=list)
    fault_analysis: str = ""


@dataclass
class L1ToL2LinkData:
    """L1→L2 链路数据

    记录从L1交换板到L2交换板的信号链路信息，
    包括L1/L2交换板、端口、光模块、SNR数据及故障分析结果。
    无L2对端信息时，L2相关字段为空
    """

    l1_switch_id: str = ""
    l1_switch_name: str = ""
    l1_room_name: str = ""
    l1_cabinet_id: str = ""
    l1_switch_chip_id: str = ""
    l1_switch_chip_phy_id: str = ""
    l1_interface: str = ""
    l1_port_speed: str = ""
    l1_hilink_snr: str = ""
    l1_host_snr: str = ""
    l1_media_snr: str = ""
    l2_switch_id: str = ""
    l2_switch_name: str = ""
    l2_room_name: str = ""
    l2_cabinet_id: str = ""
    l2_interface: str = ""
    l2_port_speed: str = ""
    l2_hilink_snr: str = ""
    l2_host_snr: str = ""
    l2_media_snr: str = ""
    link_status: str = ""
    triggered_rules: List[str] = field(default_factory=list)
    fault_analysis: str = ""


@dataclass
class AnalyzedRootCausePort:
    """根因分析结果端口记录

    记录经故障分析后的根因端口信息，包括是否为根因、部件类型、
    端口标识、交换机/主机/NPU标识等

    各部件类型的记录字段:
    - NPU/CPU: host_id, npu_id, chip_phy_id
    - 交换板: switch_id(交换机IP), switch_chip_id(交换板ID)
    - 光模块: switch_id(交换机IP), port(交换机端口)
    """

    is_root_cause: str = ""
    component: str = ""
    port: str = ""
    switch_id: str = ""
    switch_chip_id: str = ""
    switch_chip_phy_id: str = ""
    host_id: str = ""
    npu_id: str = ""
    chip_phy_id: str = ""

    def matches_domain(self, domain: Union[HostDomain, BmcDomain, SwitchDomain]) -> bool:
        """判断当前根因端口是否匹配给定的故障域

        根据domain的实际类型分别匹配:
        - HostDomain: 匹配 host_id、npu_id、chip_phy_id、peer_switch_id、peer_interface
        - BmcDomain: 匹配 bmc_id、npu_id、chip_phy_id
        - SwitchDomain: 匹配 swi_id、interface、peer_switch_id、peer_switch_interface

        匹配规则: 当前记录中有值的字段必须在domain中找到对应匹配，
        至少有一个字段匹配成功才算匹配

        Returns:
            True 表示当前根因端口属于该故障域
        """
        if domain is None:
            return False

        if isinstance(domain, HostDomain):
            return self._match_host_domain(domain)
        elif isinstance(domain, BmcDomain):
            return self._match_bmc_domain(domain)
        elif isinstance(domain, SwitchDomain):
            return self._match_switch_domain(domain)

        return False

    def _match_host_domain(self, domain: HostDomain) -> bool:
        """匹配 HostDomain 故障域

        匹配逻辑:
        优先匹配主机维度: host_id、npu_id、chip_phy_id
           - 有值的字段必须全部匹配，至少有一个有值
        """
        has_host_fields = self.host_id or self.npu_id or self.chip_phy_id
        if has_host_fields:
            return self._match_host_fields(domain)

        return False

    def _match_host_fields(self, domain: HostDomain) -> bool:
        """匹配主机维度字段: host_id、npu_id、chip_phy_id"""
        if self.host_id and self.host_id != domain.host_id:
            return False
        if self.npu_id and self.npu_id != domain.npu_id:
            return False
        if self.chip_phy_id and self.chip_phy_id != domain.chip_phy_id:
            return False
        return bool(self.host_id or self.npu_id or self.chip_phy_id)

    def _match_bmc_domain(self, domain: BmcDomain) -> bool:
        """匹配 BmcDomain 故障域

        匹配 npu_id、chip_phy_id 两个维度字段。
        BmcDomain 没有 host_id 字段，因此 AnalyzedRootCausePort 的 host_id
        不参与匹配（BMC域仅做NPU/芯片级别匹配）。

        匹配规则: 有值的字段必须全部匹配，至少有一个有值字段匹配成功
        """
        if self.npu_id and self.npu_id != domain.npu_id:
            return False
        if self.chip_phy_id and self.chip_phy_id != domain.chip_phy_id:
            return False
        matched = False
        if self.npu_id and self.npu_id == domain.npu_id:
            matched = True
        if self.chip_phy_id and self.chip_phy_id == domain.chip_phy_id:
            matched = True
        return matched

    def _match_switch_domain(self, domain: SwitchDomain) -> bool:
        """匹配 SwitchDomain 故障域

        匹配逻辑:
        匹配本端交换机: switch_id ↔ swi_id
           - 交换板: swi_chip_id ↔ interface (交换板级别匹配)
           - 光模块: port ↔ interface (端口级别匹配)
        """
        if self._match_switch_fields(domain.swi_id, domain.interface):
            return True

        return False

    def _match_switch_fields(self, swi_id: str, interface: str) -> bool:
        """匹配交换机维度字段: switch_id、port/swi_chip_id

        switch_id 必须匹配。
        interface 匹配优先级: port > swi_chip_id
        - 光模块记录: port 有值，与 interface 匹配
        - 交换板记录: port 为空但 swi_chip_id 有值，与 interface 匹配
        - 无端口信息: 仅 switch_id 匹配即可
        """
        if not self.switch_id:
            return False
        if self.switch_id != swi_id:
            return False
        if self.port and self.port != interface:
            return False

        if not self.port and self.switch_chip_id and self.switch_chip_id != interface:
            return False
        if self.port or self.switch_chip_id:
            return True
        return True
