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

"""utils.list_tool 列表工具函数单元测试。"""

import unittest

from ascend_fd_tk.utils.list_tool import (
    find_first,
    group_by_to_dict,
    group_by_to_list,
    safe_get,
    list_of_lists_to_dict_list,
)


class TestFindFirst(unittest.TestCase):
    def test_find_first_hit(self):
        result = find_first([1, 2, 3, 4], lambda x: x % 2 == 0)
        self.assertEqual(result, 2)

    def test_find_first_miss_default(self):
        result = find_first([1, 3], lambda x: x % 2 == 0, default=-1)
        self.assertEqual(result, -1)


class TestGroupBy(unittest.TestCase):
    def test_group_by_to_dict(self):
        data = [1, 2, 3, 4, 5]
        result = group_by_to_dict(data, lambda x: x % 2)
        self.assertEqual(result[0], [2, 4])
        self.assertEqual(result[1], [1, 3, 5])

    def test_group_by_to_list(self):
        data = ["aa", "bb", "cc"]
        result = group_by_to_list(data, lambda x: x[0])
        self.assertEqual(result, [["aa"], ["bb"], ["cc"]])


class TestSafeGet(unittest.TestCase):
    def test_safe_get_positive(self):
        self.assertEqual(safe_get([1, 2, 3], 1), 2)

    def test_safe_get_negative(self):
        self.assertEqual(safe_get([1, 2, 3], -1), 3)

    def test_safe_get_out_of_range(self):
        self.assertEqual(safe_get([1, 2], 5, default="x"), 2)

    def test_safe_get_empty(self):
        self.assertEqual(safe_get([], 0, default="d"), "d")


class TestListOfListsToDictList(unittest.TestCase):
    def test_headers_from_first_row(self):
        data = [["a", "b"], [1, 2], [3, 4]]
        result = list_of_lists_to_dict_list(data)
        self.assertEqual(result, [{"a": 1, "b": 2}, {"a": 3, "b": 4}])

    def test_explicit_headers_pad_short_row(self):
        data = [[1, 2], [3]]
        result = list_of_lists_to_dict_list(data, headers=["a", "b", "c"])
        self.assertEqual(result[1], {"a": 3, "b": None, "c": None})

    def test_empty_data(self):
        self.assertEqual(list_of_lists_to_dict_list([]), [])


if __name__ == "__main__":
    unittest.main()
