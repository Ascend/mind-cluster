#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
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
import unittest
import shutil
import os
import tempfile

from ascend_fd.pkg.parse.knowledge_graph.parser.mindie_parser import MindieParser
from ascend_fd.pkg.parse.knowledge_graph.parser.bmc_parser import BMCAppDumpParser


class TestParser(unittest.TestCase):
    def setUp(self):
        self.parser = MindieParser({})

    def test_log_time_format_with_utc(self):
        line = "[2025-11-18 23:01:07.144+08:00] Some log message"
        expected_time = "2025-11-18 23:01:07.144000"
        self.assertEqual(self.parser.get_log_time(line), expected_time)

    def test_log_time_format_with_utc_two(self):
        line = "[2025-06-30 20:41:24.643119+08:00] Some log message"
        expected_time = "2025-06-30 20:41:24.643119"
        self.assertEqual(self.parser.get_log_time(line), expected_time)

    def test_log_time_format_with_utc_trans(self):
        self.parser.timezone_trans_flag = True
        line = "[2025-11-18 23:01:07.144+08:00] Some log message"
        expected_time = "2025-11-18 15:01:07.144000"
        self.assertEqual(self.parser.get_log_time(line), expected_time)

    def test_log_time_format_with_utc_two_trans(self):
        self.parser.timezone_trans_flag = True
        line = "[2025-06-30 20:41:24.643119+08:00] Some log message"
        expected_time = "2025-06-30 12:41:24.643119"
        self.assertEqual(self.parser.get_log_time(line), expected_time)

    def test_log_time_format_with_no_utc(self):
        line = "[2025-11-18 23:00:59.940] Some log message"
        expected_time = "2025-11-18 23:00:59.940000"
        self.assertEqual(self.parser.get_log_time(line), expected_time)

    def test_log_time_format_with_no_utc_two(self):
        line = "[2025-11-19 00:42:30,094] Some log message"
        expected_time = "2025-11-19 00:42:30.094000"
        self.assertEqual(self.parser.get_log_time(line), expected_time)

    def test_log_time_with_no_brackets(self):
        line = "2025-11-18 23:01:07.144+08:00 Some log message"
        self.assertEqual(self.parser.get_log_time(line), "")

    def test_log_time_with_empty_brackets(self):
        line = "[] Some log message"
        self.assertEqual(self.parser.get_log_time(line), "")

    def test_log_time_with_invalid_time_format(self):
        line = "[InvalidTime] Some log message"
        self.assertEqual(self.parser.get_log_time(line), "")

    def test_log_time_with_comma_separated_time(self):
        line = "[2025-11-19 00:42:30,094] Some log message"
        expected_time = "2025-11-19 00:42:30.094000"
        self.assertEqual(self.parser.get_log_time(line), expected_time)

    def test_log_time_with_multiple_brackets(self):
        line = "[2025-11-18 23:01:07.144+08:00] [Another log] Some log message"
        expected_time = "2025-11-18 23:01:07.144000"
        self.assertEqual(self.parser.get_log_time(line), expected_time)


class TestBMCAppDumpParser(unittest.TestCase):
    """BMCAppDumpParser（AppDump/sel.txt）的时间解析与逐行过滤"""

    BMC_APP_DUMP_CODE = "0x8F180200"
    VALID_LINE = "1164    |2026-06-04 03:44:24 |0x8F180200 Asserted"
    NO_RULE_LINE = "1165    |2026-06-04 03:45:24 |Normal"
    INVALID_TIME_LINE = "1166    |2026-06-04 03:4 |0x8F180200 Asserted"
    NO_TIME_FIELD_LINE = "0x8F180200 Asserted"

    def setUp(self) -> None:
        self.params = {
            "default_conf": {"BMCAppDumpLog": {self.BMC_APP_DUMP_CODE: {"in": [self.BMC_APP_DUMP_CODE, "Asserted"]}}},
            "user_conf": {},
            "start_time": "2026-01-01 00:00:00.000000",
            "end_time": "2026-12-31 23:59:59.000000",
        }
        self.parser = BMCAppDumpParser(self.params)
        self.parser.start_time = self.params["start_time"]
        self.parser.end_time = self.params["end_time"]

    def _write_sel_txt(self, lines: list) -> str:
        tmp_dir = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, tmp_dir)
        file_path = os.path.join(tmp_dir, "sel.txt")
        with open(file_path, "w", encoding="utf-8") as f:
            f.write("\n".join(lines))
        return file_path

    def test_filter_log_time(self):
        self.assertEqual("2026-06-04 03:44:24.000000", self.parser._filter_log_time(self.VALID_LINE))

    def test_filter_log_time_with_blank(self):
        self.assertEqual(
            "2026-06-04 03:44:24.000000", self.parser._filter_log_time("1164|  2026-06-04 03:44:24  |Normal")
        )

    def test_filter_log_time_without_separator(self):
        self.assertEqual("", self.parser._filter_log_time(self.NO_TIME_FIELD_LINE))

    def test_filter_log_time_invalid(self):
        self.assertEqual("", self.parser._filter_log_time(self.INVALID_TIME_LINE))
        self.assertEqual("", self.parser._filter_log_time("1164| |Normal"))

    def test_parse_file(self):
        file_path = self._write_sel_txt(
            [self.VALID_LINE, self.NO_RULE_LINE, self.INVALID_TIME_LINE, self.NO_TIME_FIELD_LINE]
        )
        event_list = self.parser._parse_file(file_path)
        self.assertEqual(1, len(event_list))
        self.assertEqual(self.BMC_APP_DUMP_CODE, event_list[0]["event_code"])
        self.assertEqual("2026-06-04 03:44:24.000000", event_list[0]["occur_time"])

    def test_parse_file_time_out_of_range(self):
        file_path = self._write_sel_txt([self.VALID_LINE])
        self.parser.start_time = "2027-01-01 00:00:00.000000"
        self.parser.end_time = "2027-12-31 23:59:59.000000"
        self.assertEqual([], self.parser._parse_file(file_path))

    def test_parse_file_time_before_start(self):
        file_path = self._write_sel_txt([self.VALID_LINE])
        self.parser.start_time = None
        self.parser.end_time = "2026-06-04 03:44:23.000000"
        self.assertEqual([], self.parser._parse_file(file_path))


if __name__ == '__main__':
    unittest.main()
