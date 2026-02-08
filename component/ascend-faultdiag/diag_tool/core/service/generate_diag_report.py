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

from diag_tool.core.common.path import CommonPath
from diag_tool.core.context.diag_ctx import DiagCtx
from diag_tool.core.service.base import DiagService
from diag_tool.utils import csv_tool
from diag_tool.utils.logger import DIAG_LOGGER


class GenerateDiagReport(DiagService):
    def __init__(self, diag_ctx: DiagCtx):
        super().__init__(diag_ctx)

    @staticmethod
    def _open_excel_windows(file_path):
        """在 Windows 系统上打开 Excel 文件"""
        if os.name == 'nt':  # Windows
            try:
                os.startfile(file_path)
            except Exception as e:
                DIAG_LOGGER.error(f"打开文件失败: {e}")

    async def run(self):
        result = [diag_result.to_dict() for diag_result in self.diag_ctx.diag_result]
        result.sort(key=lambda x: x["故障域"])
        if not result:
            DIAG_LOGGER.warn("诊断数据为空, 请确认是否使用auto_collect进行信息采集")
            return
        os.makedirs(CommonPath.REPORT_DIR, exist_ok=True)
        csv_tool.dict_list_to_csv(result, CommonPath.REPORT_FILE, columns=["故障域", "故障码", "故障信息", "处理建议"])
        self._open_excel_windows(CommonPath.REPORT_FILE)
