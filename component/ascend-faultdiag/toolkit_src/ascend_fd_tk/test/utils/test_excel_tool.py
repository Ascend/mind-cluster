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

"""ExcelGenerator 单元测试。

覆盖本次改动：
- _get_cell_value_and_style: dict/list/set 等容器类型转字符串，空容器用 na_rep
- 修复 FormParser 空字典 {} 导致的 ValueError: Cannot convert {} to Excel
"""

import unittest

from ascend_fd_tk.utils.excel_tool import ExcelGenerator, StyledCell, CellStyle


class TestExcelToolCellParsing(unittest.TestCase):
    def test_dict_value_to_str(self):
        """非空 dict 应转为字符串，避免 openpyxl ValueError。"""
        value, style = ExcelGenerator._get_cell_value_and_style({"k": "v"}, na_rep="-")
        self.assertEqual(value, "{'k': 'v'}")
        self.assertIsInstance(style, CellStyle)

    def test_empty_dict_to_na_rep(self):
        """空 dict {} 应转为 na_rep，避免 ValueError: Cannot convert {} to Excel。"""
        value, _ = ExcelGenerator._get_cell_value_and_style({}, na_rep="-")
        self.assertEqual(value, "-")

    def test_list_value_to_str(self):
        """非空 list 应转为字符串。"""
        value, _ = ExcelGenerator._get_cell_value_and_style([1, 2, 3], na_rep="-")
        self.assertEqual(value, "[1, 2, 3]")

    def test_empty_list_to_na_rep(self):
        """空 list 应转为 na_rep。"""
        value, _ = ExcelGenerator._get_cell_value_and_style([], na_rep="-")
        self.assertEqual(value, "-")

    def test_styled_cell_preserved(self):
        """StyledCell 对象应原样返回 value 和 style，不走容器兜底。"""
        custom_style = CellStyle()
        cell = StyledCell(value={"k": "v"}, style=custom_style)
        value, style = ExcelGenerator._get_cell_value_and_style(cell, na_rep="-")
        self.assertEqual(value, {"k": "v"})
        self.assertIs(style, custom_style)


if __name__ == "__main__":
    unittest.main()
