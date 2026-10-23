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

"""threshold_report 单元测试。"""

import unittest
from dataclasses import dataclass
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.model.threshold import Threshold
from ascend_fd_tk.core.report.threshold_report import (
    ReportGenerator,
    ReportSheet,
    ThresholdConfig,
    create_threshold_report,
    generate_threshold_excel,
)


@dataclass
class _Row:
    name: str = "row"
    temp: str = "30"
    sub: object = None


class TestReportSheet(unittest.TestCase):
    def test_invalid_configs_raise(self):
        with self.assertRaises(ValueError):
            ReportSheet(
                sheet_name="s",
                data_list=[],
                header_mapping={"a": "列A"},
                header_order=["不存在的列"],
                threshold_configs=[],
            )
        config = ThresholdConfig(field_name="x", threshold=Threshold(high_value_alarm="1"))
        with self.assertRaises(ValueError):
            ReportSheet(
                sheet_name="s",
                data_list=[],
                header_mapping={"a": "列A"},
                header_order=[],
                threshold_configs=[config],
            )
        with self.assertRaises(ValueError):
            ReportSheet(
                sheet_name="s",
                data_list=[],
                header_mapping={"a": "列A"},
                header_order=[],
                threshold_configs=[ThresholdConfig(field_name="a")],
            )

    def test_header_behavior(self):
        sheet = ReportSheet(
            sheet_name="s",
            data_list=[],
            header_mapping={"a": "列A", "b": "列B"},
            header_order=[],
            threshold_configs=[],
        )
        self.assertEqual(sheet.header_order, ["列A", "列B"])
        th = Threshold(high_value_alarm="1", unit="dB")
        sheet = ReportSheet(
            sheet_name="s",
            data_list=[],
            header_mapping={"a": "SNR", "b": "其他"},
            header_order=["SNR", "其他"],
            threshold_configs=[ThresholdConfig(field_name="a", threshold=th)],
            merged_headers=[{"组": ["SNR", "其他"]}],
        )
        self.assertEqual(sheet.header_mapping["a"], "SNR (dB)")
        self.assertEqual(sheet.header_order, ["SNR (dB)", "其他"])
        self.assertEqual(sheet.merged_headers, [{"组": ["SNR (dB)", "其他"]}])


class TestReportGenerator(unittest.TestCase):
    def setUp(self):
        self.gen = ReportGenerator(excel_gen=MagicMock())

    def test_styled_cells(self):
        config = ThresholdConfig(field_name="temp", threshold=Threshold(high_value_alarm="50"), display_name="温度")
        cell = self.gen._build_styled_cell(_Row(), "80", config, na_rep="-")
        self.assertIn("80 | > 50", cell.value)
        self.assertIsNotNone(cell.style.bg_color)
        cell = self.gen._build_styled_cell(_Row(), "30", config, na_rep="-")
        self.assertEqual(cell.value, "30")
        self.assertIsNotNone(cell.style.bg_color)

    def test_field_value_and_convert(self):
        row = _Row(sub=MagicMock(level="deep"))
        self.assertEqual(ReportGenerator._get_field_value(row, "sub.level"), "deep")
        self.assertIsNone(ReportGenerator._get_field_value(row, "not_exist"))
        sheet = ReportSheet(
            sheet_name="s",
            data_list=[_Row(name="r1"), _Row(name="r2")],
            header_mapping={"name": "名称", "temp": "温度"},
            header_order=[],
            threshold_configs=[],
            na_rep="-",
        )
        styled = self.gen._convert_to_styled_dict(sheet)
        self.assertEqual(len(styled), 2)
        self.assertEqual(styled[0]["名称"], "r1")

    def test_generate_excel(self):
        with self.assertRaises(ValueError):
            self.gen.generate_excel()
        sheet = ReportSheet(
            sheet_name="s",
            data_list=[_Row(name="r1")],
            header_mapping={"name": "名称"},
            header_order=[],
            threshold_configs=[],
        )
        self.gen.add_sheet(sheet)
        self.assertIsNone(self.gen.generate_excel(output_path="/tmp/out.xlsx"))
        self.gen.excel_gen.add_sheet.assert_called_once()
        self.gen.excel_gen.generate_excel.assert_called_once_with("/tmp/out.xlsx")
        gen2 = ReportGenerator(excel_gen=MagicMock())
        gen2.add_sheet(sheet)
        self.assertIs(gen2.generate_excel(), gen2.excel_gen)


class TestConvenienceFunctions(unittest.TestCase):
    @patch("ascend_fd_tk.core.report.threshold_report.ReportGenerator")
    def test_convenience(self, MockGenerator):
        sheet = create_threshold_report(
            sheet_name="s", data_list=[_Row()], header_mapping={"name": "名称"}, threshold_configs=[]
        )
        self.assertIsInstance(sheet, ReportSheet)
        self.assertEqual(sheet.na_rep, "")
        self.assertEqual(sheet.tab_color, None)
        gen = MagicMock()
        MockGenerator.return_value = gen
        generate_threshold_excel(output_path="/tmp/out.xlsx", sheets=[MagicMock(), MagicMock()], excel_gen=MagicMock())
        self.assertEqual(gen.add_sheet.call_count, 2)
        gen.generate_excel.assert_called_once_with(output_path="/tmp/out.xlsx")


if __name__ == "__main__":
    unittest.main()
