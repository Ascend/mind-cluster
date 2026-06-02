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
from typing import List, Dict, Tuple

from ascend_fd_tk.core.common import constants
from ascend_fd_tk.core.model.diag_result import DiagResult
from ascend_fd_tk.core.report.sheet.base import BaseSheetGenerator
from ascend_fd_tk.core.report.threshold_report import ThresholdConfig, create_threshold_report, generate_threshold_excel


@dataclass
class DiagReportData:
    """诊断报告数据类，用于存储诊断报告信息"""

    fault_domain: str = ""  # 故障域
    fault_code: str = ""  # 故障码
    fault_info: str = ""  # 故障信息
    solution: str = ""  # 处理建议


@dataclass
class HostReportData(DiagReportData):
    """host sheet 诊断报告信息"""

    host_id: str = ""  # 主机ID
    hostname: str = ""  # 主机名
    sn_num: str = ""  # 主机SN
    npu_id: str = ""  # NPU ID
    chip_phy_id: str = ""  # 物理芯片ID


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


class DiagReportSheetGenerator(BaseSheetGenerator):
    """诊断报告Sheet生成器"""

    def __init__(self, cluster_info, excel_gen=None, diag_results: List[DiagResult] = None):
        """
        初始化诊断报告Sheet生成器

        :param cluster_info: 集群信息缓存对象
        :param excel_gen: Excel生成器对象，如果不提供则创建新实例
        :param diag_results: 诊断结果列表
        """
        super().__init__(cluster_info, excel_gen)
        self.diag_results = diag_results or []

    @staticmethod
    def _create_threshold_configs() -> List[ThresholdConfig]:
        """
        创建阈值配置（诊断报告可能不需要阈值检查，返回空列表）

        :return: 阈值配置列表
        """
        return []

    @staticmethod
    def _create_header_config(sheet_type: str) -> Tuple[Dict[str, str], List[str]]:
        """
        创建header映射和顺序

        :param sheet_type: sheet类型，支持"host", "bmc", "switch"
        :return: (header_mapping, header_order)
            header_mapping: {field_name: header_name}
            header_order: [header_name]
        """
        # 基础字段（只包含故障码、故障信息、处理建议）
        base_mapping = {"fault_code": "故障码", "fault_info": "故障信息", "solution": "处理建议"}
        merge_columns = ["处理建议"]

        if sheet_type == constants.FAULT_TYPE_HOST:
            header_mapping = {
                "host_id": "主机ID",
                "hostname": "主机名",
                "sn_num": "SN",
                "npu_id": "NPU ID",
                "chip_phy_id": "物理芯片ID",
                **base_mapping,
            }
            merge_columns.extend(["主机ID", "主机名", "SN", "NPU ID", "物理芯片ID"])
        elif sheet_type == constants.FAULT_TYPE_BMC:
            header_mapping = {
                "bmc_id": "BMC ID",
                "sn_num": "SN",
                "npu_id": "NPU ID",
                "chip_phy_id": "物理芯片ID",
                **base_mapping,
            }
            merge_columns.extend(["BMC ID", "SN", "NPU ID", "物理芯片ID"])
        elif sheet_type == constants.FAULT_TYPE_SWITCH:
            header_mapping = {
                "swi_name": "交换机名称",
                "swi_id": "交换机ID",
                "sn_num": "SN",
                "interface": "端口",
                **base_mapping,
            }
            merge_columns.extend(["交换机名称", "交换机ID", "SN", "端口"])
        else:
            header_mapping = base_mapping
        return header_mapping, merge_columns

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
        for sheet_name, data_list in diag_report_category_data.items():
            if data_list:
                # 按故障域排序，确保相同故障域的行在一起
                sorted_data = sorted(data_list, key=lambda x: x.fault_domain)
                # 创建header映射和顺序
                header_mapping, merge_columns = self._create_header_config(sheet_name)
                sheet = create_threshold_report(
                    sheet_name=sheet_name,
                    data_list=sorted_data,
                    header_mapping=header_mapping,
                    merge_columns=merge_columns,
                    threshold_configs=threshold_configs,
                    na_rep="-",
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
        return data
