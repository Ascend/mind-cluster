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

"""utils.csv_tool CSV 读写转换工具单元测试。"""

import os
import tempfile
import unittest

from ascend_fd_tk.utils import csv_tool


class TestFlattenDict(unittest.TestCase):
    def test_flatten_nested_dict(self):
        self.assertEqual(csv_tool.flatten_dict({"a": {"b": 1}, "c": 2}), {"a_b": 1, "c": 2})

    def test_flatten_custom_sep(self):
        self.assertEqual(csv_tool.flatten_dict({"a": {"b": 1}}, sep="."), {"a.b": 1})


class TestGetAllColumns(unittest.TestCase):
    def test_flatten_columns(self):
        self.assertEqual(csv_tool.get_all_columns([{"a": {"b": 1}}, {"c": 2}]), ["a_b", "c"])

    def test_non_flatten_columns(self):
        self.assertEqual(csv_tool.get_all_columns([{"a": {"b": 1}}], flatten=False), ["a"])


class TestDictListToCsv(unittest.TestCase):
    def test_write_then_read(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = os.path.join(tmp, "out.csv")
            csv_tool.dict_list_to_csv([{"name": "a", "age": 1}, {"name": "b", "age": 2}], path)
            result = csv_tool.csv_to_dict_list(path, encoding="utf-8-sig")
            self.assertEqual(result, [{"name": "a", "age": "1"}, {"name": "b", "age": "2"}])

    def test_empty_list_raises(self):
        with self.assertRaises(ValueError):
            csv_tool.dict_list_to_csv([], "x.csv")


class TestCsvToDictListFlatten(unittest.TestCase):
    def test_restore_nested_structure(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = os.path.join(tmp, "out.csv")
            csv_tool.dict_list_to_csv([{"a": {"b": 1}}], path, flatten=True)
            result = csv_tool.csv_to_dict_list(path, flatten=True, encoding="utf-8-sig")
            self.assertEqual(result, [{"a": {"b": "1"}}])


class TestCsvToListOfLists(unittest.TestCase):
    def test_read_normal(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = os.path.join(tmp, "t.csv")
            with open(path, "w", encoding="utf-8") as f:
                f.write("a,b\n1,2\n")
            self.assertEqual(csv_tool.csv_to_list_of_lists(path), [["a", "b"], ["1", "2"]])

    def test_file_not_found_returns_empty(self):
        self.assertEqual(csv_tool.csv_to_list_of_lists(os.path.join(tempfile.gettempdir(), "no_such.csv")), [])


if __name__ == "__main__":
    unittest.main()
