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
from collections import defaultdict
from typing import Dict, List, Set, Tuple, Union

from ascend_fd_tk.core.root_cause.constants import (
    ROOT_CAUSE_NPU_CPU,
    ROOT_CAUSE_L1_SWITCH_BOARD,
    ROOT_CAUSE_L1_OPTICAL,
    ROOT_CAUSE_L2_OPTICAL,
    LINK_STATUS_ABNORMAL,
    ROOT_CAUSE_CABLE_TRAY,
    IS_ROOT_CAUSE,
    NOT_ROOT_CAUSE,
    RULE_L1_200G_HILINK,
    RULE_L1_400G_HILINK,
    RULE_L1_OPTICAL_HOST,
    RULE_L1_OPTICAL_MEDIA,
    RULE_L2_200G_HILINK,
    RULE_L2_OPTICAL_MEDIA,
    NPU,
    CPU,
)
from ascend_fd_tk.core.root_cause.model import (
    HostToL1LinkData,
    L1ToL2LinkData,
    AnalyzedRootCausePort,
)


class FaultAnalyzer:
    """故障分析器，负责Host->L1和L1->L2的故障分析与根因端口记录

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

    def analyze_host_to_l1_fault(
        self,
        host_to_l1_links: List[HostToL1LinkData],
    ) -> List[AnalyzedRootCausePort]:
        """分析 Host→L1 链路故障，返回根因端口列表

        流程:
        1. 按 (host_id, chip_key, l1_switch_id) 分组
        2. 统计每组中异常交换板数量
        3. 构建故障分析文本
        4. 创建 NPU/CPU、L1交换板、L2光模块 三类根因端口记录
        """
        analyzed_ports: List[AnalyzedRootCausePort] = []
        group_data = self._group_host_to_l1_links(host_to_l1_links)

        for rows in group_data.values():
            abnormal_count = _count_rule(rows, RULE_L1_200G_HILINK)
            total_count = len(rows)
            analysis = self._host_to_l1_fault_component_analysis(rows, total_count, abnormal_count)
            self._set_fault_component_analysis(rows, analysis)
            analyzed_ports.extend(self._create_host_to_l1_root_cause_ports(rows, total_count, abnormal_count))

        return analyzed_ports

    def analyze_l1_to_l2_fault(
        self,
        l1_to_l2_links: List[L1ToL2LinkData],
    ) -> List[AnalyzedRootCausePort]:
        """分析 L1→L2 链路故障，返回根因端口列表

        流程:
        1. 按 (l1_switch_id, l1_swi_chip_id) 分组
        2. 构建故障分析文本
        3. 分别对 L1交换板、L1光模块、L2光模块、L2交换板 四类部件做根因判定
        """
        analyzed_ports: List[AnalyzedRootCausePort] = []
        group_data = self._group_l1_to_l2_links(l1_to_l2_links)

        for rows in group_data.values():
            analysis = self.l1_to_l2_component_analysis(rows)
            self._set_fault_component_analysis(rows, analysis)
            analyzed_ports.extend(self._create_l1_to_l2_root_cause_ports(rows))

        return analyzed_ports

    # ---- Host→L1 helpers ----

    @staticmethod
    def _group_host_to_l1_links(
        links: List[HostToL1LinkData],
    ) -> Dict[Tuple[str, str, str], List[HostToL1LinkData]]:
        """按 (host_id, chip_key, l1_switch_id) 分组

        chip_key: NPU使用chip_phy_id，CPU使用"cpu_{npu_id}"作为标识
        同一NPU/CPU芯片到同一L1交换机的所有200G端口链路归为一组
        """
        group_data: Dict[Tuple[str, str, str], List[HostToL1LinkData]] = defaultdict(list)
        for data in links:
            chip_key = data.chip_phy_id if data.chip_phy_id else f"cpu_{data.npu_id}"
            group_key = (data.host_id, chip_key, data.l1_switch_id)
            group_data[group_key].append(data)
        return group_data

    @staticmethod
    def _calc_board_stats(rows: List[HostToL1LinkData]) -> Tuple[int, int]:
        """统计一组链路中涉及的交换板总数和异常交换板数

        返回: (total_count, abnormal_count)
        用于判断NPU/CPU是否为根因: 全部异常→是根因, 部分异常→不是根因
        """
        total_boards: Set[str] = set()
        abnormal_boards: Set[str] = set()
        for row in rows:
            total_boards.add(row.l1_switch_chip_id)
            if row.link_status == LINK_STATUS_ABNORMAL:
                abnormal_boards.add(row.l1_switch_chip_id)
        return len(total_boards), len(abnormal_boards)

    @staticmethod
    def _host_to_l1_fault_component_analysis(
        rows: List[HostToL1LinkData],
        total_count: int,
        abnormal_count: int,
    ) -> str:
        """构建 Host→L1 链路的故障分析文本

        分析逻辑:
        - 全部交换板异常 → NPU/CPU故障，需排查NPU/CPU
        - 部分交换板异常 → 部分链路异常，需排查线缆桥
        """
        parts: List[str] = []

        if abnormal_count == total_count and total_count > 0:
            first = rows[0]
            xpu_label = _build_xpu_label(first)
            parts.append(
                f"与{xpu_label}连接的L1交换机所有200G端口hilink SNR异常，建议排查异常链路的{first.npu_type}板、模组或所有异常链路的{ROOT_CAUSE_CABLE_TRAY}"
            )
        elif abnormal_count > 0:
            first = rows[0]
            xpu_label = _build_xpu_label(first)
            parts.append(
                f"与{xpu_label}连接的L1交换机部分200G端口hilink SNR异常，建议排查异常链路{ROOT_CAUSE_CABLE_TRAY}"
            )
        return "；".join(parts)

    @staticmethod
    def _create_host_to_l1_root_cause_ports(
        rows: List[HostToL1LinkData],
        total_count: int,
        abnormal_count: int,
    ) -> List[AnalyzedRootCausePort]:
        """为 Host→L1 链路创建 NPU/CPU 根因端口记录

        NPU/CPU: 全部交换板异常时为根因 (规则R1)
        记录字段: host_id, npu_id, chip_phy_id
        """
        first = rows[0]
        is_root = abnormal_count > 0 and abnormal_count == total_count
        ports: List[AnalyzedRootCausePort] = [
            AnalyzedRootCausePort(
                is_root_cause=IS_ROOT_CAUSE if is_root else NOT_ROOT_CAUSE,
                component=ROOT_CAUSE_NPU_CPU,
                host_id=first.host_id,
                npu_id=first.npu_id,
                chip_phy_id=first.chip_phy_id,
            )
        ]
        return ports

    # ---- L1→L2 helpers ----

    @staticmethod
    def _group_l1_to_l2_links(
        links: List[L1ToL2LinkData],
    ) -> Dict[Tuple[str, str], List[L1ToL2LinkData]]:
        """按 (l1_switch_id, l1_swi_chip_id) 分组

        同一L1交换板上同一芯片的所有400G上行端口链路归为一组
        """
        group_data: Dict[Tuple[str, str], List[L1ToL2LinkData]] = defaultdict(list)
        for data in links:
            group_key = (data.l1_switch_id, data.l1_switch_chip_id)
            group_data[group_key].append(data)
        return group_data

    @staticmethod
    def l1_to_l2_component_analysis(rows: List[L1ToL2LinkData]) -> str:
        """构建 L1→L2 链路的故障分析文本

        分析逻辑:
        - L1 host SNR: 全部异常→交换板故障, 部分异常→交换板端口故障 (规则R3)
        - l1_400g_hilink 触发 → 需排查L1光模块 (规则R2)
        - l1_optical_media 触发 → 需排查L2光模块 (规则R4)

        - l2_optical_media 触发 → 需排查L1光模块 (规则R6)
        - l2_200g_hilink 触发 → 需排查L2光模块 (规则R7)
        - l2_optical_host 触发 → 需排查L2交换板 (规则R5)
        """
        parts: List[str] = []
        l1_host_count = _count_rule(rows, RULE_L1_OPTICAL_HOST)
        total = len(rows)

        if l1_host_count == total and total > 0:
            parts.append(
                f"与L1交换板(chip_id:{rows[0].l1_switch_chip_phy_id})连接的所有光模块端口host SNR均异常，建议更换该交换板或排查所有链路光纤"
            )
        elif l1_host_count > 0:
            parts.append(
                f"与L1交换板({rows[0].l1_switch_chip_phy_id})连接的部分光模块端口host SNR异常"
                f"({l1_host_count}/{total})，建议排查异常链路的交换板端口及光纤"
            )

        if _has_rule(rows, RULE_L1_400G_HILINK):
            parts.append("L1交换机400G端口hilink SNR异常，建议排查异常链路L1交换机的光模块（清污或更换）")
        if _has_rule(rows, RULE_L1_OPTICAL_MEDIA):
            parts.append("L1交换机的光模块media SNR异常，建议排查异常链路对端L2交换机的光模块（清污或更换）")
        if _has_rule(rows, RULE_L2_OPTICAL_MEDIA):
            parts.append("L2交换机的光模块media SNR异常，建议排查异常链路对端L1交换机的光模块（清污或更换）")
        if _has_rule(rows, RULE_L2_200G_HILINK):
            parts.append("L2交换机200G端口hilink SNR异常，建议排查异常链路L2交换机的光模块（清污或更换）")

        return "；".join(parts)

    @staticmethod
    def _create_l1_to_l2_root_cause_ports(rows: List[L1ToL2LinkData]) -> List[AnalyzedRootCausePort]:
        """编排四类部件的根因端口创建"""
        ports: List[AnalyzedRootCausePort] = []
        ports.extend(FaultAnalyzer._analyze_l1_switch_board(rows))
        ports.extend(FaultAnalyzer._analyze_l1_optical(rows))
        ports.extend(FaultAnalyzer._analyze_l2_optical(rows))
        return ports

    @staticmethod
    def _analyze_l1_switch_board(rows: List[L1ToL2LinkData]) -> List[AnalyzedRootCausePort]:
        """L1交换板根因分析 (规则R3: L1光模块host SNR异常 → L1交换板故障)

        判定逻辑:
        - 所有端口l1_optical_host均触发 → 交换板整体故障，是根因
        - 部分端口触发 → 交换板整体不是根因，但触发端口是根因
        - 均未触发 → 不是根因

        记录字段: switch_id(交换机IP), swi_chip_id(交换板ID)
        """
        ports = []
        for row in rows:
            is_root = RULE_L1_OPTICAL_HOST in row.triggered_rules
            ports.append(
                AnalyzedRootCausePort(
                    is_root_cause=IS_ROOT_CAUSE if is_root else NOT_ROOT_CAUSE,
                    component=ROOT_CAUSE_L1_SWITCH_BOARD,
                    switch_id=row.l1_switch_id,
                    switch_chip_id=row.l1_switch_chip_id,
                    switch_chip_phy_id=row.l1_switch_chip_phy_id,
                )
            )
        return ports

    @staticmethod
    def _analyze_l1_optical(rows: List[L1ToL2LinkData]) -> List[AnalyzedRootCausePort]:
        """L1光模块根因分析

        判定逻辑 (任一触发即为根因):
        - 规则R2: l1_400g_hilink触发 → L1光模块故障
        - 规则R6: l2_optical_media触发 → L1光模块故障

        记录字段: switch_id(交换机IP), port(交换机端口)
        """
        ports: List[AnalyzedRootCausePort] = []
        for row in rows:
            is_root = RULE_L1_400G_HILINK in row.triggered_rules or RULE_L2_OPTICAL_MEDIA in row.triggered_rules
            ports.append(
                AnalyzedRootCausePort(
                    is_root_cause=IS_ROOT_CAUSE if is_root else NOT_ROOT_CAUSE,
                    component=ROOT_CAUSE_L1_OPTICAL,
                    switch_id=row.l1_switch_id,
                    port=row.l1_interface,
                )
            )
        return ports

    @staticmethod
    def _analyze_l2_optical(rows: List[L1ToL2LinkData]) -> List[AnalyzedRootCausePort]:
        """L2光模块根因分析

        判定逻辑 (任一触发即为根因):
        - 规则R4: l1_optical_media触发 → L2光模块故障
        - 规则R7: l2_200g_hilink触发 → L2光模块故障

        无L2对端信息的端口跳过（l2_switch_id为空时无法定位L2光模块）

        记录字段: switch_id(交换机IP), port(交换机端口)
        """
        ports: List[AnalyzedRootCausePort] = []
        for row in rows:
            is_root = RULE_L1_OPTICAL_MEDIA in row.triggered_rules or RULE_L2_200G_HILINK in row.triggered_rules
            if not row.l2_switch_id:
                continue
            ports.append(
                AnalyzedRootCausePort(
                    is_root_cause=IS_ROOT_CAUSE if is_root else NOT_ROOT_CAUSE,
                    component=ROOT_CAUSE_L2_OPTICAL,
                    switch_id=row.l2_switch_id,
                    port=row.l2_interface,
                )
            )
        return ports

    # ---- Common helpers ----

    @staticmethod
    def _set_fault_component_analysis(rows: List[Union[HostToL1LinkData, L1ToL2LinkData]], analysis: str) -> None:
        """将故障分析文本写入每条链路数据的fault_analysis字段"""
        for row in rows:
            row.fault_analysis = analysis


def _has_rule(rows: List[Union[HostToL1LinkData, L1ToL2LinkData]], rule: str) -> bool:
    """判断行列表中是否有任一行触发了指定规则"""
    return any(rule in row.triggered_rules for row in rows)


def _count_rule(rows: List[Union[HostToL1LinkData, L1ToL2LinkData]], rule: str) -> int:
    """统计行列表中触发指定规则的行数"""
    return sum(1 for row in rows if rule in row.triggered_rules)


def _build_xpu_label(data: HostToL1LinkData) -> str:
    """构建NPU/CPU标签文本，格式: NPU(npu_id:xxx, chip_phy_id:xxx) 或 CPU(cpu_id:xxx)"""
    if data.npu_type == NPU:
        label = f"{data.npu_type}(npu_id:{data.npu_id}"
    elif data.npu_type == CPU:
        label = f"{data.npu_type}(cpu_id:{data.npu_id}"
    else:
        return ""
    chip_phy_id_str = f", chip_phy_id:{data.chip_phy_id}" if data.chip_phy_id else ""
    return f"{label}{chip_phy_id_str})"
