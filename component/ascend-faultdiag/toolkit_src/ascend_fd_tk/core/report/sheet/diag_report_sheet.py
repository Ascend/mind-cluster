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
诊断报告Sheet生成器
"""

from dataclasses import dataclass
from typing import List, Dict, Tuple, Optional

from ascend_fd_tk.core.common import constants
from ascend_fd_tk.core.model.diag_result import DiagResult
from ascend_fd_tk.core.report.sheet.base import BaseSheetGenerator
from ascend_fd_tk.core.report.threshold_report import ThresholdConfig, create_threshold_report, generate_threshold_excel
from ascend_fd_tk.core.root_cause.constants import UNKNOWN_ROOT_CAUSE
from ascend_fd_tk.core.root_cause.filter import RootCauseFilter


@dataclass
class DiagReportData:
    """诊断报告数据类，用于存储诊断报告信息"""

    fault_domain: str = ""  # 故障域
    fault_code: str = ""  # 故障码
    fault_info: str = ""  # 故障信息
    solution: str = ""  # 处理建议
    root_cause_status: str = ""  # 是否链路故障根因


@dataclass
class HostReportData(DiagReportData):
    """host sheet 诊断报告信息"""

    host_id: str = ""  # 主机ID
    hostname: str = ""  # 主机名
    sn_num: str = ""  # 主机SN
    npu_id: str = ""  # NPU ID
    chip_phy_id: str = ""  # 物理芯片ID
    room_name: str = ""  # 机房名称
    cabinet_id: str = ""  # 机柜编号


@dataclass
class BmcReportData(DiagReportData):
    """Bmc sheet 诊断报告信息"""

    bmc_id: str = ""  # BMC ID
    sn_num: str = ""  # SN
    npu_id: str = ""  # NPU ID
    chip_phy_id: str = ""  # 物理芯片ID


@dataclass
class SwitchReportData(DiagReportData):
    """Switch sheet 诊断报告信息"""

    swi_id: str = ""  # 交换机ID
    sn_num: str = ""  # 交换机SN
    swi_name: str = ""  # 交换机名
    interface: str = ""  # 交换机端口
    room_name: str = ""  # 机房名称
    cabinet_id: str = ""  # 机柜编号


class DiagReportSheetGenerator(BaseSheetGenerator):
    """诊断报告Sheet生成器"""

    _TWO_ROW = 2
    _TAB_COLOR_DIAG = "9DC3E6"

    SHEET_MAPPING = {
        constants.FAULT_TYPE_HOST: "带内故障分析(Host)",
        constants.FAULT_TYPE_BMC: "带外故障分析(BMC)",
        constants.FAULT_TYPE_SWITCH: "交换机故障分析(L1&L2&RoCE)",
    }

    def __init__(
        self,
        cluster_info,
        excel_gen=None,
        diag_results: List[DiagResult] = None,
        root_cause_filter: Optional[RootCauseFilter] = None,
    ):
        """
        初始化诊断报告Sheet生成器

        :param cluster_info: 集群信息缓存对象
        :param excel_gen: Excel生成器对象，如果不提供则创建新实例
        :param diag_results: 诊断结果列表
        """
        super().__init__(cluster_info, excel_gen)
        self.diag_results = diag_results or []
        self._root_cause_filter = root_cause_filter

    @staticmethod
    def _create_threshold_configs() -> List[ThresholdConfig]:
        """
        创建阈值配置（诊断报告可能不需要阈值检查，返回空列表）

        :return: 阈值配置列表
        """
        return []

    @staticmethod
    def _create_header_config(sheet_type: str) -> Dict[str, str]:
        """
        创建header映射

        :param sheet_type: sheet类型，支持"host", "bmc", "switch"
        :return: header_mapping
        """
        base_mapping = {
            "fault_code": "故障码",
            "fault_info": "故障信息",
            "solution": "处理建议",
            "root_cause_status": "是否链路故障根因",
        }

        if sheet_type == constants.FAULT_TYPE_HOST:
            header_mapping = {
                "host_id": "主机ID",
                "hostname": "主机名",
                "sn_num": "SN",
                "room_name": "机房名称",
                "cabinet_id": "机柜编号",
                "npu_id": "NPU ID",
                "chip_phy_id": "物理芯片ID",
                **base_mapping,
            }
        elif sheet_type == constants.FAULT_TYPE_BMC:
            header_mapping = {
                "bmc_id": "BMC ID",
                "sn_num": "SN",
                "npu_id": "NPU ID",
                "chip_phy_id": "物理芯片ID",
                **base_mapping,
            }
        elif sheet_type == constants.FAULT_TYPE_SWITCH:
            header_mapping = {
                "swi_name": "交换机名称",
                "swi_id": "交换机ID",
                "sn_num": "SN",
                "room_name": "机房名称",
                "cabinet_id": "机柜编号",
                "interface": "端口",
                **base_mapping,
            }
        else:
            header_mapping = base_mapping
        return header_mapping

    @staticmethod
    def _get_merge_col_titles(sheet_type: str) -> List[str]:
        """获取需要合并的列标题列表"""
        base_titles = ["处理建议", "是否链路故障根因"]
        if sheet_type == constants.FAULT_TYPE_HOST:
            return base_titles + ["主机ID", "主机名", "SN", "机房名称", "机柜编号", "NPU ID", "物理芯片ID"]
        elif sheet_type == constants.FAULT_TYPE_BMC:
            return base_titles + ["BMC ID", "SN", "NPU ID", "物理芯片ID"]
        elif sheet_type == constants.FAULT_TYPE_SWITCH:
            return base_titles + ["交换机名称", "交换机ID", "SN", "机房名称", "机柜编号", "端口"]
        return base_titles

    def generate_sheet(self) -> None:
        """
        生成诊断报告Excel Sheet
        """
        # 收集诊断报告数据
        diag_report_category_data = self._collect_diag_report_category_data()

        # 如果没有数据，跳过生成Sheet
        if not diag_report_category_data:
            return

        # 创建阈值配置（诊断报告可能不需要阈值检查）
        threshold_configs = self._create_threshold_configs()
        sheets = []
        for sheet_type, data_list in diag_report_category_data.items():
            if data_list:
                # 按故障域排序，确保相同故障域的行在一起
                sorted_data = sorted(data_list, key=lambda x: x.fault_domain)
                # 创建header映射
                header_mapping = self._create_header_config(sheet_type)
                # 计算合并范围
                merge_cells = self._compute_merge_ranges(sorted_data, header_mapping, sheet_type)
                sheet = create_threshold_report(
                    sheet_name=self.SHEET_MAPPING.get(sheet_type, ""),
                    data_list=sorted_data,
                    header_mapping=header_mapping,
                    threshold_configs=threshold_configs,
                    na_rep="-",
                    merge_cells=merge_cells,
                    tab_color=self._TAB_COLOR_DIAG,
                )
                sheets.append(sheet)
        # 生成Excel
        generate_threshold_excel(excel_gen=self.excel_gen, sheets=sheets)

    def _collect_diag_report_category_data(self) -> Dict[str, List[DiagReportData]]:
        """
        收集诊断报告数据

        :return: 诊断报告数据列表
        """
        sheet_category_data = {
            # 定义sheet顺序
            constants.FAULT_TYPE_HOST: [],
            constants.FAULT_TYPE_BMC: [],
            constants.FAULT_TYPE_SWITCH: [],
        }
        for diag_result in self.diag_results:
            if diag_result.fault_type == constants.FAULT_TYPE_HOST:
                data = self._get_host_sheet_data(diag_result)
            elif diag_result.fault_type == constants.FAULT_TYPE_BMC:
                data = self._get_bmc_sheet_data(diag_result)
            elif diag_result.fault_type == constants.FAULT_TYPE_SWITCH:
                data = self._get_switch_sheet_data(diag_result)
            else:
                continue
            # 公共列名的数据
            data.fault_domain = diag_result.get_domain_desc()
            data.fault_info = diag_result.fault_info
            data.fault_code = diag_result.err_code
            data.solution = diag_result.suggestion
            data.root_cause_status = UNKNOWN_ROOT_CAUSE
            if self._root_cause_filter:
                data.root_cause_status = self._root_cause_filter.get_root_cause_status(diag_result.domain)
            sheet_category_data[diag_result.fault_type].append(data)
        return sheet_category_data

    def _get_host_sheet_data(self, diag_result: DiagResult) -> HostReportData:
        data = HostReportData()
        domain = diag_result.domain
        data.host_id = domain.host_id
        data.npu_id = domain.npu_id
        data.chip_phy_id = domain.chip_phy_id
        host_info = self.cluster_info.hosts_info.get(data.host_id)
        if host_info:
            data.hostname = host_info.hostname
            data.sn_num = host_info.sn_num
            data.room_name = host_info.room_name
            data.cabinet_id = host_info.cabinet_id
        return data

    def _get_bmc_sheet_data(self, diag_result: DiagResult) -> BmcReportData:
        data = BmcReportData()
        domain = diag_result.domain
        data.bmc_id = domain.bmc_id
        data.npu_id = domain.npu_id
        data.chip_phy_id = domain.chip_phy_id
        bmc_info = self.cluster_info.bmcs_info.get(data.bmc_id)
        if bmc_info:
            data.sn_num = bmc_info.sn_num
        return data

    def _get_switch_sheet_data(self, diag_result: DiagResult) -> SwitchReportData:
        data = SwitchReportData()
        domain = diag_result.domain
        data.swi_id = domain.swi_id
        data.interface = domain.interface
        swi_info = self.cluster_info.swis_info.get(data.swi_id)
        if swi_info:
            data.sn_num = swi_info.sn
            data.swi_name = swi_info.name
            data.room_name = swi_info.room_name
            data.cabinet_id = swi_info.cabinet_id
        return data

    def _compute_merge_ranges(
        self,
        data_list: List[DiagReportData],
        header_mapping: Dict[str, str],
        sheet_type: str,
    ) -> List[Tuple[int, int, int, int]]:
        """计算诊断报告的合并区域

        按故障域(fault_domain)分组，对同一组内连续相同的行合并指定列。
        合并范围格式: (start_row, start_col, end_row, end_col)
        """
        if not data_list:
            return []

        title_to_col = {title: idx + 1 for idx, title in enumerate(header_mapping.values())}
        merge_titles = self._get_merge_col_titles(sheet_type)
        merge_col_indices = [title_to_col[t] for t in merge_titles if t in title_to_col]

        merge_ranges = []
        i = 0
        while i < len(data_list):
            fault_domain = data_list[i].fault_domain
            j = i + 1
            while j < len(data_list) and data_list[j].fault_domain == fault_domain:
                # pylint: disable=duplicate-code  # 已与同类分析器复用逻辑，忽略重复警告
                j += 1

            if j - i > 1:
                start_row = i + self._TWO_ROW
                end_row = j + 1
                for col in merge_col_indices:
                    merge_ranges.append((start_row, col, end_row, col))

            i = j

        return merge_ranges
