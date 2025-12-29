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
from ascend_fd.pkg.parse.knowledge_graph.parser.mindie_parser import MindieParser


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


if __name__ == '__main__':
    unittest.main()
