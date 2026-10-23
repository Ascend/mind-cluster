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

"""utils.date_tool 日期对象工具单元测试。"""

import unittest

from ascend_fd_tk.utils.date_tool import DateObj


class TestDateObj(unittest.TestCase):
    def test_timestamp_cached(self):
        date_obj = DateObj("2024-01-01 00:00:00", "%Y-%m-%d %H:%M:%S")
        t1 = date_obj.timestamp
        t2 = date_obj.timestamp
        self.assertEqual(t1, t2)

    def test_diff_seconds(self):
        earlier = DateObj("2024-01-01 00:00:00", "%Y-%m-%d %H:%M:%S")
        later = DateObj("2024-01-01 00:01:00", "%Y-%m-%d %H:%M:%S")
        self.assertAlmostEqual(later.diff_seconds(earlier), 60, delta=1)
        self.assertAlmostEqual(earlier.diff_seconds(later), -60, delta=1)


if __name__ == "__main__":
    unittest.main()
