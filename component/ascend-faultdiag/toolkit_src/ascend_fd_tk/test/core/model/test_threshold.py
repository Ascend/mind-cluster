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

"""Threshold 单元测试：阈值比较、字符串告警描述、Lane 差值检查与合并。"""

import unittest

from ascend_fd_tk.core.model.threshold import Threshold, ThresholdStatus


class TestThreshold(unittest.TestCase):
    def setUp(self):
        self.threshold = Threshold(
            low_value_alarm="1",
            high_value_alarm="10",
            low_value_warn="2",
            high_value_warn="9",
            desc="测试项",
            unit="dB",
        )

    def test_check_value_high_alarm(self):
        status, th = self.threshold.check_value("11")
        self.assertEqual(status, ThresholdStatus.HIGH_THRESHOLD_ALARM)
        self.assertEqual(th, "10")

    def test_check_value_normal(self):
        status, th = self.threshold.check_value("5")
        self.assertEqual(status, ThresholdStatus.NORMAL)
        self.assertEqual(th, "")

    def test_check_value_str_low_warn(self):
        desc = self.threshold.check_value_str("1.5")
        self.assertIn("测试项", desc)
        self.assertIn("低于告警阈值", desc)
        self.assertIn("单位：dB", desc)

    def test_check_value_str_normal_empty(self):
        self.assertEqual(self.threshold.check_value_str("5"), "")

    def test_check_value_string_not_equal_alarm(self):
        th = Threshold(normal_value_alarm="enabled", desc="开关")
        status, _ = th.check_value("disabled")
        self.assertEqual(status, ThresholdStatus.NOT_EQUAL_THRESHOLD_ALARM)
        status, _ = th.check_value("enabled")
        self.assertEqual(status, ThresholdStatus.NORMAL)

    def test_check_lane_diff_desc(self):
        desc_list = self.threshold.check_lane_diff_desc([["0", "1"], ["1", "5"], ["2", "20"]], value_type="Host SNR")
        self.assertTrue(desc_list)
        self.assertTrue(any("Lane1" in desc and "Lane2" in desc for desc in desc_list))

    def test_check_lane_diff_desc_unparsable(self):
        desc_list = self.threshold.check_lane_diff_desc([["0", "abc"], ["1", "def"]])
        self.assertEqual(desc_list, [])

    def test_merged(self):
        merged = self.threshold.merged(high_value_alarm="20")
        self.assertEqual(merged.high_alarm_th, "20")
        self.assertEqual(merged.low_alarm_th, "1")
        self.assertEqual(self.threshold.high_alarm_th, "10")


if __name__ == "__main__":
    unittest.main()
