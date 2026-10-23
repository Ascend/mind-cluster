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

"""AutoInspection 巡检服务测试"""

import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.common.diag_enum import Customer
from ascend_fd_tk.core.service.auto_inspection import AutoInspection


class _FakeInspection:
    def __init__(self, cache, customer):
        self.cache = cache
        self.customer = customer

    def check(self):
        item = MagicMock()
        item.to_csv_dict.return_value = {"问题现象": "fake"}
        return [item]


class TestAutoInspection(unittest.IsolatedAsyncioTestCase):
    async def test_run_writes_csv(self):
        ctx = MagicMock()
        ctx.inspection_result = []
        ctx.cache = MagicMock()
        with (
            patch("ascend_fd_tk.core.service.auto_inspection.recursive_scan_and_register"),
            patch("ascend_fd_tk.core.service.auto_inspection.INSPECTION_STORE", [_FakeInspection]),
            patch("ascend_fd_tk.core.service.auto_inspection.os.makedirs"),
            patch("ascend_fd_tk.core.service.auto_inspection.csv_tool.dict_list_to_csv") as mock_csv,
            patch.object(AutoInspection, "_open_excel_windows") as mock_open,
        ):
            await AutoInspection(ctx, Customer.DEFAULT).run()
            self.assertEqual(len(ctx.inspection_result), 1)
            mock_csv.assert_called_once()
            mock_open.assert_called_once()

    async def test_run_empty_result_returns(self):
        ctx = MagicMock()
        ctx.inspection_result = []
        ctx.cache = MagicMock()
        with (
            patch("ascend_fd_tk.core.service.auto_inspection.recursive_scan_and_register"),
            patch("ascend_fd_tk.core.service.auto_inspection.INSPECTION_STORE", []),
            patch("ascend_fd_tk.core.service.auto_inspection.csv_tool.dict_list_to_csv") as mock_csv,
        ):
            await AutoInspection(ctx, Customer.DEFAULT).run()
            mock_csv.assert_not_called()


if __name__ == "__main__":
    unittest.main()
