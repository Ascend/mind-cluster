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

"""utils.table_parser 命令行表格解析工具单元测试。"""

import unittest

from ascend_fd_tk.utils.table_parser import TableParser


class TestFormatTitlesDict(unittest.TestCase):
    def test_from_sequence(self):
        self.assertEqual(TableParser.format_titles_dict(["a", "b"]), {"a": "a", "b": "b"})

    def test_from_dict_passthrough(self):
        self.assertEqual(TableParser.format_titles_dict({"x": "X"}), {"x": "X"})


class TestTableParserParse(unittest.TestCase):
    def test_parse_basic(self):
        recv = "Name  Age\nTom   18\nJerry 20\n"
        result = TableParser.parse(recv, {"name": "Name", "age": "Age"})
        self.assertEqual(result, [{"name": "Tom", "age": "18"}, {"name": "Jerry", "age": "20"}])

    def test_parse_with_end_sign_and_replace(self):
        recv = "System Name\nA B\nTotal entries displayed: 1\n"
        titles = {"sys": "System Name"}
        replace = {"sys": {"pattern": r"\s", "repl": ""}}
        result = TableParser.parse(recv, titles, regex_replace_dict=replace, end_sign="Total entries")
        self.assertEqual(result, [{"sys": "AB"}])

    def test_parse_empty_recv(self):
        self.assertEqual(TableParser.parse("   \n", {"a": "A"}), [])

    def test_get_title_start_line(self):
        lines = ["xxx", "Name  Age"]
        title, index = TableParser.get_title_start_line(2, {"name": "Name", "age": "Age"}, lines)
        self.assertEqual(index, 1)
        self.assertEqual(title, "Name  Age")

    def test_regex_replace_dict_value(self):
        result = TableParser.regex_replace_dict_value([{"id": "a 1"}], {"id": {"pattern": r"\s", "repl": ""}})
        self.assertEqual(result, [{"id": "a1"}])


if __name__ == "__main__":
    unittest.main()
