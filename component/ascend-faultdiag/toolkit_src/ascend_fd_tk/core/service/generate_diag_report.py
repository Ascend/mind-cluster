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

import os.path
import time

from ascend_fd_tk.core.common.path import CommonPath
from ascend_fd_tk.core.service.base import DiagService
from ascend_fd_tk.utils.logger import DIAG_LOGGER
from ascend_fd_tk.utils.excel_tool import ExcelGenerator
from ascend_fd_tk.core.report.sheet.diag_report_sheet import DiagReportSheetGenerator
from ascend_fd_tk.core.report.sheet.optical_module_sheet import HostToSwitchOpticalModuleSheetGenerator
from ascend_fd_tk.core.report.sheet.signal_link_mapping_sheet import SignalLinkMappingSheetGenerator
from ascend_fd_tk.core.report.sheet.switch_optical_module_sheet import SwitchOpticalModuleSheetGenerator


class GenerateDiagReport(DiagService):
    @staticmethod
    def _open_excel_windows(file_path):
        """在 Windows 系统上打开 Excel 文件"""
        if os.name == 'nt':  # Windows
            try:
                os.startfile(file_path)  # pylint: disable=no-member
            except Exception as e:
                DIAG_LOGGER.error("打开文件失败：%s", str(e))

    async def run(self):
        # 收集诊断结果
        diag_results = self.diag_ctx.diag_result
        if not diag_results:
            DIAG_LOGGER.warning("诊断数据为空，请确认是否使用auto_collect进行信息采集")
            return

        # 创建Excel生成器实例
        excel_gen = ExcelGenerator()

        # 创建报告目录
        os.makedirs(CommonPath.REPORT_DIR, exist_ok=True)

        # 生成带时间后缀的文件名
        time_suffix = time.strftime("%Y%m%d_%H%M%S")
        excel_file_name = f"diag_report_{time_suffix}.xlsx"
        excel_file_path = os.path.join(CommonPath.REPORT_DIR, excel_file_name)

        try:
            DIAG_LOGGER.info("正在生成链路分析Sheet...")
            signal_link_mapping_sheet = SignalLinkMappingSheetGenerator(
                cluster_info=self.diag_ctx.cache, excel_gen=excel_gen, root_cause_filter=self.diag_ctx.root_cause_filter
            )
            signal_link_mapping_sheet.generate_sheet()

            DIAG_LOGGER.info("正在生成故障诊断报告Sheet...")
            diag_report_sheet = DiagReportSheetGenerator(
                cluster_info=self.diag_ctx.cache,
                excel_gen=excel_gen,
                diag_results=diag_results,
                root_cause_filter=self.diag_ctx.root_cause_filter,
            )
            diag_report_sheet.generate_sheet()

            DIAG_LOGGER.info("正在生成光模块信息Sheet...")
            optical_module_sheet = HostToSwitchOpticalModuleSheetGenerator(
                cluster_info=self.diag_ctx.cache,
                excel_gen=excel_gen,
            )
            optical_module_sheet.generate_sheet()

            DIAG_LOGGER.info("正在生成交换机间端口连接光模块信息Sheet...")
            switch_optical_module_sheet = SwitchOpticalModuleSheetGenerator(
                cluster_info=self.diag_ctx.cache,
                excel_gen=excel_gen,
            )
            switch_optical_module_sheet.generate_sheet()

            DIAG_LOGGER.info("正在保存诊断报告到：%s...", excel_file_path)
            excel_gen.generate_excel(excel_file_path)
            DIAG_LOGGER.info("诊断报告生成完成：%s", excel_file_path)

            # 尝试打开Excel文件
            self._open_excel_windows(excel_file_path)

        except Exception as e:
            DIAG_LOGGER.error("生成诊断报告失败：%s", str(e))
            raise e
