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

"""utils.form_parser 表单解析工具单元测试。"""

import unittest

from ascend_fd_tk.utils.form_parser import FormParser


class TestFormParser(unittest.TestCase):
    def test_get_key_value(self):
        parser = FormParser()
        self.assertEqual(parser.get_key_value("name : alice"), ("name", "alice"))

    def test_cnt_space_indent(self):
        parser = FormParser()
        self.assertEqual(parser.cnt_space_indent("  abc"), 2)
        self.assertEqual(parser.cnt_space_indent("abc"), 0)

    def test_parse_basic(self):
        parser = FormParser()
        result = parser.parse("name : alice\nage : 18\n")
        self.assertEqual(result, {"name": "alice", "age": "18"})

    def test_parse_nested_section(self):
        parser = FormParser()
        result = parser.parse("config:\n  key : value\n  num : 1\n")
        self.assertEqual(result, {"config": {"key": "value", "num": "1"}})

    def test_parse_skip_sign(self):
        parser = FormParser(skip_sign="#")
        result = parser.parse("name : alice\n# comment\nage : 18\n")
        self.assertEqual(result, {"name": "alice", "age": "18"})

    def test_parse_append_multi_line(self):
        parser = FormParser(append_multi_line=True)
        result = parser.parse("desc : hello\n  world\n")
        self.assertEqual(result, {"desc": "hello\nworld"})


if __name__ == "__main__":
    unittest.main()
