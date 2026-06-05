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
from typing import List, Dict, Tuple, Optional, Union

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

    _SHEET_MAPPING = {
        constants.FAULT_TYPE_HOST: "带内故障分析(Host)",
        constants.FAULT_TYPE_BMC: "带外故障分析(BMC)",
        constants.FAULT_TYPE_SWITCH: "交换机故障分析(L1&L2&RoCE)",
    }

    # 每种sheet类型的实体排序属性（同时也是实体分组键）
    _ENTITY_SORT_ATTRS: Dict[str, List[str]] = {
        constants.FAULT_TYPE_HOST: ["host_id", "npu_id", "chip_phy_id"],
        constants.FAULT_TYPE_BMC: ["bmc_id", "npu_id", "chip_phy_id"],
        constants.FAULT_TYPE_SWITCH: ["swi_id", "interface"],
    }

    # 故障列属性（不参与实体合并）
    _FAULT_ATTRS = {"fault_code", "fault_info", "solution", "root_cause_status"}
    # 第二级合并的故障列属性
    _FAULT_MERGE_ATTRS = ["solution", "root_cause_status"]

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

    @classmethod
    def _get_entity_merge_titles(cls, sheet_type: str) -> List[str]:
        """第一级合并列标题：从header_mapping中排除故障列，自动推导"""
        header_mapping = cls._create_header_config(sheet_type)
        return [title for attr, title in header_mapping.items() if attr not in cls._FAULT_ATTRS]

    @classmethod
    def _get_fault_merge_titles(cls) -> List[str]:
        """第二级合并列标题：在实体合并基础上按故障域分组"""
        header_mapping = cls._create_header_config(sheet_type="")
        return [header_mapping[attr] for attr in cls._FAULT_MERGE_ATTRS if attr in header_mapping]

    @classmethod
    def _get_sort_key(cls, data: Union[HostReportData, BmcReportData, SwitchReportData], sheet_type: str) -> Tuple:
        """获取排序键：实体属性 + fault_domain"""
        sort_attrs = cls._ENTITY_SORT_ATTRS.get(sheet_type, [])
        return tuple(getattr(data, attr, "") for attr in sort_attrs) + (data.fault_domain,)

    @classmethod
    def _get_entity_key(cls, data: Union[HostReportData, BmcReportData, SwitchReportData], sheet_type: str) -> Tuple:
        """获取实体分组键（第一级合并依据）"""
        sort_attrs = cls._ENTITY_SORT_ATTRS.get(sheet_type, [])
        return tuple(getattr(data, attr, "") for attr in sort_attrs)

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
                # 按实体+故障域排序，确保相同实体的行相邻
                sorted_data = sorted(data_list, key=lambda x, st=sheet_type: self._get_sort_key(x, st))
                # 创建header映射
                header_mapping = self._create_header_config(sheet_type)
                # 计算合并范围（两级合并）
                merge_cells = self._compute_merge_ranges(sorted_data, header_mapping, sheet_type)
                sheet = create_threshold_report(
                    sheet_name=self._SHEET_MAPPING.get(sheet_type, ""),
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
        """计算诊断报告的合并区域（两级合并）

        第一级：按实体分组，合并实体列（主机ID、主机名、SN等）
        第二级：在实体组内按 fault_domain 分组，合并故障列（处理建议、是否根因）
        """
        if not data_list:
            return []

        title_to_col = {title: idx + 1 for idx, title in enumerate(header_mapping.values())}
        entity_col_indices = [title_to_col[t] for t in self._get_entity_merge_titles(sheet_type) if t in title_to_col]
        fault_col_indices = [title_to_col[t] for t in self._get_fault_merge_titles() if t in title_to_col]

        merge_ranges = []
        group_start = 0
        while group_start < len(data_list):
            group_end = self._find_entity_group_end(data_list, group_start, sheet_type)
            # 第一级：合并实体列
            if group_end - group_start > 1:
                for col in entity_col_indices:
                    merge_ranges.append((group_start + self._TWO_ROW, col, group_end + 1, col))
            # 第二级：在实体组内按故障域分组，合并故障列
            merge_ranges.extend(
                self._compute_fault_merge_in_group(data_list, group_start, group_end, fault_col_indices)
            )
            group_start = group_end

        return merge_ranges

    def _find_entity_group_end(self, data_list: List[DiagReportData], group_start: int, sheet_type: str) -> int:
        """查找实体分组的结束位置（不含），返回第一个不属于当前实体的行索引"""
        entity_key = self._get_entity_key(data_list[group_start], sheet_type)
        group_end = group_start + 1
        while group_end < len(data_list) and self._get_entity_key(data_list[group_end], sheet_type) == entity_key:
            group_end += 1
        return group_end

    def _compute_fault_merge_in_group(
        self,
        data_list: List[DiagReportData],
        group_start: int,
        group_end: int,
        fault_col_indices: List[int],
    ) -> List[Tuple[int, int, int, int]]:
        """在实体组内按 fault_domain 分组，计算故障列的合并范围"""
        merge_ranges = []
        seg_start = group_start
        while seg_start < group_end:
            fault_domain = data_list[seg_start].fault_domain
            seg_end = seg_start + 1
            while seg_end < group_end and data_list[seg_end].fault_domain == fault_domain:
                seg_end += 1
            if seg_end - seg_start > 1:
                for col in fault_col_indices:
                    merge_ranges.append((seg_start + self._TWO_ROW, col, seg_end + 1, col))
            seg_start = seg_end
        return merge_ranges
