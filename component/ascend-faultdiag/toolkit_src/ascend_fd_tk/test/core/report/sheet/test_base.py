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

"""BaseSheetGenerator 单元测试"""

import unittest
from unittest.mock import MagicMock

from ascend_fd_tk.core.report.sheet.base import BaseSheetGenerator


class _ConcreteGenerator(BaseSheetGenerator):
    def generate_sheet(self) -> None:
        self.excel_gen.add_sheet(sheet_name="concrete", data=[], columns=[])


class TestBaseSheetGenerator(unittest.TestCase):
    def test_abstract_method_raises(self):
        with self.assertRaises(TypeError):
            BaseSheetGenerator(MagicMock())  # pylint: disable=abstract-class-instantiated

    def test_default_excel_gen_created(self):
        gen = _ConcreteGenerator(MagicMock())
        self.assertIsNotNone(gen.excel_gen)
        self.assertTrue(hasattr(gen.excel_gen, "add_sheet"))

    def test_excel_gen_used(self):
        excel_gen = MagicMock()
        gen = _ConcreteGenerator(MagicMock(), excel_gen=excel_gen)
        self.assertIs(gen.excel_gen, excel_gen)

    def test_get_excel_gen_and_save_excel(self):
        excel_gen = MagicMock()
        gen = _ConcreteGenerator(MagicMock(), excel_gen=excel_gen)
        self.assertIs(gen.get_excel_gen(), excel_gen)
        gen.save_excel("/tmp/out.xlsx")
        excel_gen.generate_excel.assert_called_once_with("/tmp/out.xlsx")

    def test_generate_sheet_runs(self):
        excel_gen = MagicMock()
        gen = _ConcreteGenerator(MagicMock(), excel_gen=excel_gen)
        gen.generate_sheet()
        excel_gen.add_sheet.assert_called_once()


if __name__ == "__main__":
    unittest.main()
