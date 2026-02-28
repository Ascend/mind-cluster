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

import unittest

from ascend_fd_tk.core.collect.parser.bmc_parser import BmcParser


class TestBmcParser(unittest.TestCase):

    def test_trans_sel_results(self):
        # 测试trans_sel_results方法
        cmd_res = """ID          Generation Time    Severity    Event Code    Status    Event Description
1           2026-02-01 10:00:00    Critical    1001          Active    Test Event
2           2026-02-01 11:00:00    Warning     1002          Inactive  Another Event"""
        result = BmcParser.trans_sel_results(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].sel_id, "1")
        self.assertEqual(result[0].severity, "Critical")
        self.assertEqual(result[1].event_code, "1002")

    def test_trans_sensor_results(self):
        # 测试trans_sensor_results方法
        cmd_res = """sensor id    sensor name    value    unit    status    lnr    lc    lnc    unc      uc    unr    phys    nhys
1             Temperature    35         C        OK         0       0      0       0        0      0      35       0
2             Voltage        5.0       V        OK         0       0      0       0        0      0      5.0      0"""
        result = BmcParser.trans_sensor_results(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].sensor_name, "Temperature")
        self.assertEqual(result[0].value, "35")
        self.assertEqual(result[1].sensor_id, "2")

    def test_trans_health_events_results(self):
        # 测试trans_health_events_results方法
        # 使用更简单的表格格式，确保列对齐准确
        cmd_res = """Event Num Event Time Alarm Level Event Code Event Description
1         2026-02-01     Critical    2001        Test Event 1
2         2026-02-02     Major       2002        Test Event 2"""
        result = BmcParser.trans_health_events_results(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].event_num, "1")
        self.assertEqual(result[0].alarm_level, "Critical")
        self.assertEqual(result[1].event_code, "2002")


if __name__ == '__main__':
    unittest.main()
