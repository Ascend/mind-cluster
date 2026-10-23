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

"""GenerateDiagReport 报告生成服务测试"""

import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.service.generate_diag_report import GenerateDiagReport


class TestGenerateDiagReport(unittest.IsolatedAsyncioTestCase):
    async def test_run_no_results_returns(self):
        ctx = MagicMock()
        ctx.diag_result = []
        with patch("ascend_fd_tk.core.service.generate_diag_report.SignalLinkMappingSheetGenerator") as mock_sig:
            await GenerateDiagReport(ctx).run()
            mock_sig.assert_not_called()

    async def test_run_generates_excel(self):
        ctx = MagicMock()
        ctx.diag_result = [MagicMock()]
        ctx.cache = MagicMock()
        ctx.root_cause_filter = MagicMock()
        with (
            patch("ascend_fd_tk.core.service.generate_diag_report.ExcelGenerator") as mock_excel_cls,
            patch("ascend_fd_tk.core.service.generate_diag_report.SignalLinkMappingSheetGenerator") as mock_sig,
            patch("ascend_fd_tk.core.service.generate_diag_report.DiagReportSheetGenerator") as mock_diag,
            patch("ascend_fd_tk.core.service.generate_diag_report.HostToSwitchOpticalModuleSheetGenerator") as mock_opt,
            patch("ascend_fd_tk.core.service.generate_diag_report.SwitchOpticalModuleSheetGenerator") as mock_swi,
            patch("ascend_fd_tk.core.service.generate_diag_report.os.makedirs"),
            patch("ascend_fd_tk.core.service.generate_diag_report.time.strftime", return_value="20260829_120000"),
            patch.object(GenerateDiagReport, "_open_excel_windows") as mock_open,
        ):
            await GenerateDiagReport(ctx).run()
            mock_sig.return_value.generate_sheet.assert_called_once()
            mock_diag.return_value.generate_sheet.assert_called_once()
            mock_opt.return_value.generate_sheet.assert_called_once()
            mock_swi.return_value.generate_sheet.assert_called_once()
            mock_excel_cls.return_value.generate_excel.assert_called_once()
            mock_open.assert_called_once()

    async def test_run_propagates_exception(self):
        ctx = MagicMock()
        ctx.diag_result = [MagicMock()]
        ctx.cache = MagicMock()
        ctx.root_cause_filter = MagicMock()
        with (
            patch("ascend_fd_tk.core.service.generate_diag_report.ExcelGenerator"),
            patch("ascend_fd_tk.core.service.generate_diag_report.SignalLinkMappingSheetGenerator") as mock_sig,
            patch("ascend_fd_tk.core.service.generate_diag_report.os.makedirs"),
        ):
            mock_sig.return_value.generate_sheet.side_effect = RuntimeError("boom")
            with self.assertRaises(RuntimeError):
                await GenerateDiagReport(ctx).run()


if __name__ == "__main__":
    unittest.main()
