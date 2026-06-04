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
from typing import Dict, List, Union

from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import HostDomain, BmcDomain, SwitchDomain
from ascend_fd_tk.core.root_cause.constants import IS_ROOT_CAUSE, NOT_ROOT_CAUSE, UNKNOWN_ROOT_CAUSE
from ascend_fd_tk.core.root_cause.model import (
    AnalyzedRootCausePort,
    HostToL1LinkData,
    L1ToL2LinkData,
)
from ascend_fd_tk.core.root_cause.link_builder import LinkBuilder
from ascend_fd_tk.core.root_cause.fault_analyzer import FaultAnalyzer


class RootCauseFilter:
    """基于 SNR 信号链规则的根因筛选器（端口级别）

    信号链路模型:
        NPU/CPU -> L1交换板 -> L1光模块 -> L2光模块 -> L2交换板

    根因定位规则 (A->B, 在B上检测SNR异常, 说明A部件故障 或 A->B链路故障):
        R1: L1 200G hilink SNR异常  -> NPU/CPU故障
        R2: L1 400G hilink SNR异常  -> L1光模块故障
        R3: L1光模块 host SNR异常   -> L1交换板故障
        R4: L1光模块 media SNR异常  -> L2光模块故障
        R5: L2光模块 host SNR异常   -> L2交换板故障
        R6: L2光模块 media SNR异常  -> L1光模块故障
        R7: L2 200G hilink SNR异常  -> L2光模块故障
    """

    def __init__(self, cluster_info: ClusterInfoCache):
        self.cluster_info = cluster_info
        self._link_builder = LinkBuilder(cluster_info)
        self._fault_analyzer = FaultAnalyzer()

        self._host_to_l1_links: List[HostToL1LinkData] = []
        self._l1_to_l2_links: List[L1ToL2LinkData] = []
        self._analyzed_root_cause_ports: List[AnalyzedRootCausePort] = []
        self._port_key_map: Dict[str, int] = {}

    @property
    def host_to_l1_links(self) -> List[HostToL1LinkData]:
        """Host→L1 链路数据列表"""
        return self._host_to_l1_links

    @property
    def l1_to_l2_links(self) -> List[L1ToL2LinkData]:
        """L1→L2 链路数据列表"""
        return self._l1_to_l2_links

    @property
    def analyzed_root_cause_ports(self) -> List[AnalyzedRootCausePort]:
        """分析后的根因端口列表"""
        return self._analyzed_root_cause_ports

    @staticmethod
    def _build_port_key(port: AnalyzedRootCausePort) -> str:
        """构建根因端口的唯一键，用于去重比较"""
        return (
            f"{port.component}|{port.port}|{port.switch_id}"
            f"|{port.switch_chip_id}|{port.host_id}"
            f"|{port.npu_id}|{port.chip_phy_id}"
        )

    def _merge_analyzed_ports(self, ports: List[AnalyzedRootCausePort]) -> None:
        """合并根因端口到列表，去重并保留"是根因"优先级

        去重规则: 以 (component, port, switch_id, host_id, npu_id, chip_phy_id) 为唯一键，
        若已存在且新记录为"是根因"，则替换旧记录
        若唯一键已存在:
        - 新记录为"是根因"时替换旧记录
        - 新记录为"不是根因"时保留旧记录
        若唯一键不存在则追加
        """
        for port in ports:
            key = self._build_port_key(port)
            if key in self._port_key_map:
                if port.is_root_cause == IS_ROOT_CAUSE:
                    idx = self._port_key_map[key]
                    self._analyzed_root_cause_ports[idx] = port
                continue
            self._port_key_map[key] = len(self._analyzed_root_cause_ports)
            self._analyzed_root_cause_ports.append(port)

    def build_and_analyze(self) -> None:
        """构建链路数据并执行故障分析，合并结果到根因端口列表"""
        self._host_to_l1_links, self._l1_to_l2_links = self._link_builder.build()

        host_analyzed = self._fault_analyzer.analyze_host_to_l1_fault(self._host_to_l1_links)
        self._merge_analyzed_ports(host_analyzed)

        l1_analyzed = self._fault_analyzer.analyze_l1_to_l2_fault(self._l1_to_l2_links)
        self._merge_analyzed_ports(l1_analyzed)

    def get_root_cause_status(self, domain: Union[HostDomain, BmcDomain, SwitchDomain]) -> str:
        """查询给定故障域的根因状态

        优先查找"是根因"的匹配记录，其次查找"不是根因"的匹配记录，
        均未找到则返回"未知"

        Args:
            domain: 故障域对象，HostDomain/BmcDomain/SwitchDomain

        Returns:
            "是根因"、"不是根因" 或 "未知"
        """
        for port in self._analyzed_root_cause_ports:
            if port.is_root_cause == IS_ROOT_CAUSE and port.matches_domain(domain):
                return IS_ROOT_CAUSE
        for port in self._analyzed_root_cause_ports:
            if port.is_root_cause == NOT_ROOT_CAUSE and port.matches_domain(domain):
                return NOT_ROOT_CAUSE
        return UNKNOWN_ROOT_CAUSE
