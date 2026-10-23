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

"""utils.multi_level_dict_tool 多级字典工具单元测试。"""

import unittest

from ascend_fd_tk.utils.multi_level_dict_tool import MultiLevelDict


class TestMultiLevelDict(unittest.TestCase):
    def test_write_and_read(self):
        mld = MultiLevelDict()
        mld.write(["a", "b", "c"], 1)
        self.assertEqual(mld.read(["a", "b", "c"]), 1)
        self.assertEqual(mld.read(["a", "b"]), {"c": 1})

    def test_write_overwrites_value(self):
        mld = MultiLevelDict({"a": 1})
        mld.write(["a"], 2)
        self.assertEqual(mld.read(["a"]), 2)

    def test_read_missing_returns_default(self):
        mld = MultiLevelDict()
        self.assertIsNone(mld.read(["x", "y"]))
        self.assertEqual(mld.read(["x", "y"], default="d"), "d")

    def test_write_empty_keys_raises(self):
        mld = MultiLevelDict()
        with self.assertRaises(ValueError):
            mld.write([], 1)

    def test_write_mid_level_not_dict_raises(self):
        mld = MultiLevelDict({"a": 1})
        with self.assertRaises(TypeError):
            mld.write(["a", "b"], 1)

    def test_delete_existing(self):
        mld = MultiLevelDict({"a": {"b": 1}})
        self.assertTrue(mld.delete(["a", "b"]))
        self.assertEqual(mld.get_target_data(), {"a": {}})

    def test_delete_missing_returns_false(self):
        mld = MultiLevelDict()
        self.assertFalse(mld.delete(["x", "y"]))

    def test_delete_empty_keys_raises(self):
        mld = MultiLevelDict()
        with self.assertRaises(ValueError):
            mld.delete([])


if __name__ == "__main__":
    unittest.main()
