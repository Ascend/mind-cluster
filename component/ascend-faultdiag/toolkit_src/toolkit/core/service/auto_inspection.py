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

import os

from toolkit.core.common.diag_enum import Customer
from toolkit.core.common.path import CommonPath
from toolkit.core.context.diag_ctx import DiagCtx
from toolkit.core.context.register import recursive_scan_and_register, INSPECTION_STORE
from toolkit.core.service.base import DiagService
from toolkit.utils import csv_tool
from toolkit.utils.logger import DIAG_LOGGER


class AutoInspection(DiagService):

    def __init__(self, diag_ctx: DiagCtx, customer: Customer):
        super().__init__(diag_ctx)
        self.customer = customer

    @staticmethod
    def _open_excel_windows(file_path):
        """在 Windows 系统上打开 Excel 文件"""
        if os.name == 'nt':  # Windows
            try:
                os.startfile(file_path)
            except Exception as e:
                DIAG_LOGGER.error(f"打开文件失败: {e}")

    async def run(self):
        recursive_scan_and_register("toolkit.core.inspection")
        for cls in INSPECTION_STORE:
            self.diag_ctx.inspection_result.extend(cls(self.diag_ctx.cache, self.customer).check())
        result = [inspection_result.to_csv_dict() for inspection_result in self.diag_ctx.inspection_result]
        if not result:
            DIAG_LOGGER.error("诊断数据为空, 请确认是否进行信息采集")
            return
        os.makedirs(CommonPath.REPORT_DIR, exist_ok=True)
        columns = ["A端设备名称", "A端IP", "A端接口", "A端SN", "A端光模块SN", "B端设备名称", "B端IP", "B端接口",
                   "B端SN", "B端光模块SN", "问题现象"]
        csv_tool.dict_list_to_csv(result, CommonPath.INSPECTION_ERRORS_REPORT_FILE, columns=columns)
        self._open_excel_windows(CommonPath.INSPECTION_ERRORS_REPORT_FILE)
