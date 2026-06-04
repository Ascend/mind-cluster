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

"""
信号链路映射关系报告Sheet生成器
拆分为两个Sheet：
  1. NPU-CPU到L1下行链路：L1 200G端口与NPU/CPU的对应关系，含链路分析
  2. L1到L2上行链路：L1 400G端口与L2 200G端口的对应关系，含链路分析

复用 RootCauseFilter 预构建的 HostToL1LinkData 和 L1ToL2LinkData（含 fault_analysis）
"""

from typing import List, Dict, Tuple, Optional

from ascend_fd_tk.core.root_cause.filter import (
    RootCauseFilter,
    HostToL1LinkData,
    L1ToL2LinkData,
)
from ascend_fd_tk.core.report.sheet.base import BaseSheetGenerator
from ascend_fd_tk.core.report.threshold_report import create_threshold_report, generate_threshold_excel


class SignalLinkMappingSheetGenerator(BaseSheetGenerator):
    _TWO_ROW = 2

    def __init__(self, cluster_info, excel_gen=None, root_cause_filter: Optional[RootCauseFilter] = None):
        super().__init__(cluster_info, excel_gen)
        self._root_cause_filter = root_cause_filter

    @staticmethod
    def _create_downstream_header_config() -> Dict[str, str]:
        header_mapping = {
            "host_id": "主机ID",
            "host_name": "主机名",
            "room_name": "机房名称",
            "cabinet_id": "机柜编号",
            "npu_type": "NPU/CPU类型",
            "npu_id": "NPU/CPU ID",
            "chip_phy_id": "物理芯片ID",
            "l1_switch_id": "L1交换机IP",
            "l1_switch_name": "L1交换机名称",
            "l1_switch_chip_id": "L1交换板ID",
            "l1_switch_chip_phy_id": "L1交换板物理端口号",
            "l1_interface": "L1端口",
            "l1_hilink_snr": "hilink SNR",
            "link_status": "链路状态",
            "fault_analysis": "链路分析",
        }
        return header_mapping

    @staticmethod
    def _create_upstream_header_config() -> Dict[str, str]:
        header_mapping = {
            "l1_switch_id": "L1交换机IP",
            "l1_switch_name": "L1交换机名称",
            "l1_room_name": "L1机房名称",
            "l1_cabinet_id": "L1机柜编号",
            "l1_switch_chip_id": "L1交换板ID",
            "l1_switch_chip_phy_id": "L1交换板物理端口号",
            "l1_interface": "L1端口",
            "l1_hilink_snr": "L1 hilink SNR",
            "l1_host_snr": "L1 host SNR",
            "l1_media_snr": "L1 media SNR",
            "l2_switch_id": "L2交换机IP",
            "l2_switch_name": "L2交换机名称",
            "l2_room_name": "L2机房名称",
            "l2_cabinet_id": "L2机柜编号",
            "l2_interface": "L2端口",
            "l2_hilink_snr": "L2 hilink SNR",
            "l2_host_snr": "L2 host SNR",
            "l2_media_snr": "L2 media SNR",
            "link_status": "链路状态",
            "fault_analysis": "链路分析",
        }
        return header_mapping

    _DOWNSTREAM_COARSE_MERGE_TITLES = [
        "主机ID",
        "主机名",
        "机房名称",
        "机柜编号",
        "NPU/CPU类型",
        "NPU/CPU ID",
        "物理芯片ID",
        "L1交换机IP",
        "L1交换机名称",
        "链路分析",
    ]

    _DOWNSTREAM_FINE_MERGE_TITLES = [
        "L1交换板ID",
    ]

    _UPSTREAM_MERGE_COL_TITLES = [
        "L1交换机IP",
        "L1交换机名称",
        "L1机房名称",
        "L1机柜编号",
        "L1交换板ID",
        "链路分析",
    ]

    @staticmethod
    def _downstream_coarse_key(data: HostToL1LinkData) -> Tuple[str, str, str]:
        chip_key = data.chip_phy_id if data.chip_phy_id else f"cpu_{data.npu_id}"
        return data.host_id, chip_key, data.l1_switch_id

    @staticmethod
    def _upstream_group_key(data: L1ToL2LinkData) -> Tuple[str, str]:
        return data.l1_switch_id, data.l1_switch_chip_id

    def generate_sheet(self) -> None:
        if not self._root_cause_filter:
            return

        downstream_data = list(self._root_cause_filter.host_to_l1_links)
        upstream_data = list(self._root_cause_filter.l1_to_l2_links)

        sheets = []

        # Sheet标签颜色：链路故障分析使用浅橙色
        _TAB_COLOR_LINK = "F4B183"

        if downstream_data:
            downstream_data.sort(key=lambda d: (self._downstream_coarse_key(d), d.l1_switch_chip_id))
            merge_cells = self._compute_downstream_merge_ranges(downstream_data)
            header_mapping = self._create_downstream_header_config()
            sheets.append(
                create_threshold_report(
                    sheet_name="NPU&CPU到L1链路分析",
                    data_list=downstream_data,
                    header_mapping=header_mapping,
                    threshold_configs=[],
                    na_rep="-",
                    merge_cells=merge_cells,
                    tab_color=_TAB_COLOR_LINK,
                )
            )

        if upstream_data:
            upstream_data.sort(key=lambda d: (self._upstream_group_key(d), d.l1_interface))
            merge_cells = self._compute_upstream_merge_ranges(upstream_data)
            header_mapping = self._create_upstream_header_config()
            sheets.append(
                create_threshold_report(
                    sheet_name="L1到L2链路分析",
                    data_list=upstream_data,
                    header_mapping=header_mapping,
                    threshold_configs=[],
                    na_rep="-",
                    merge_cells=merge_cells,
                    tab_color=_TAB_COLOR_LINK,
                )
            )

        if sheets:
            generate_threshold_excel(excel_gen=self.excel_gen, sheets=sheets)

    def _compute_downstream_merge_ranges(self, data_list: List[HostToL1LinkData]) -> List[Tuple[int, int, int, int]]:
        """计算下行链路的合并区域，支持两级合并

        粗粒度合并: 按 (host_id, chip_key, l1_switch_id) 分组，
                    合并 "主机名"、"NPU/CPU类型" 等列
        细粒度合并: 在粗粒度组内，按连续相同的 l1_switch_chip_id 分段，
                    合并 "L1交换板ID" 列，确保不同交换板显示正确值
        """
        if not data_list:
            return []

        header_mapping = self._create_downstream_header_config()
        title_to_col = {title: idx + 1 for idx, title in enumerate(header_mapping.values())}

        coarse_col_indices = [title_to_col[t] for t in self._DOWNSTREAM_COARSE_MERGE_TITLES if t in title_to_col]
        fine_col_indices = [title_to_col[t] for t in self._DOWNSTREAM_FINE_MERGE_TITLES if t in title_to_col]

        merge_ranges = []
        i = 0
        while i < len(data_list):
            coarse_key = self._downstream_coarse_key(data_list[i])
            j = i + 1
            while j < len(data_list) and self._downstream_coarse_key(data_list[j]) == coarse_key:
                j += 1

            if j - i > 1:
                start_row = i + self._TWO_ROW
                end_row = j + 1
                for col in coarse_col_indices:
                    merge_ranges.append((start_row, col, end_row, col))

            k = i
            while k < j:
                chip_id = data_list[k].l1_switch_chip_id
                m = k + 1
                while m < j and data_list[m].l1_switch_chip_id == chip_id:
                    m += 1
                if m - k > 1:
                    for col in fine_col_indices:
                        merge_ranges.append((k + self._TWO_ROW, col, m + 1, col))
                k = m

            i = j

        return merge_ranges

    def _compute_upstream_merge_ranges(self, data_list: List[L1ToL2LinkData]) -> List[Tuple[int, int, int, int]]:
        if not data_list:
            return []

        header_mapping = self._create_upstream_header_config()
        title_to_col = {title: idx + 1 for idx, title in enumerate(header_mapping.values())}
        merge_col_indices = [title_to_col[t] for t in self._UPSTREAM_MERGE_COL_TITLES if t in title_to_col]

        merge_ranges = []
        i = 0
        while i < len(data_list):
            group_key = self._upstream_group_key(data_list[i])
            j = i + 1
            while j < len(data_list) and self._upstream_group_key(data_list[j]) == group_key:
                j += 1
            # pylint: disable=duplicate-code  # 已与同类分析器复用逻辑，忽略重复警告
            if j - i > 1:
                start_row = i + self._TWO_ROW
                end_row = j + 1
                for col in merge_col_indices:
                    merge_ranges.append((start_row, col, end_row, col))

            i = j

        return merge_ranges
