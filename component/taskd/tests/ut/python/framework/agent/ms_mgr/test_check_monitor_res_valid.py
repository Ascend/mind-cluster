#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

import unittest
from taskd.python.toolkit.constants import constants
from taskd.python.framework.agent.ms_mgr.MsUtils import check_monitor_res_valid


class TestCheckMonitorResValid(unittest.TestCase):
    def setUp(self):
        # 为测试用例设置一个有效的输入示例
        self.valid_rank_status_dict = {
            0: {'pid': 101, 'status': None, 'global_rank': 16},
            1: {'pid': 110, 'status': None, 'global_rank': 17},
            2: {'pid': 119, 'status': None, 'global_rank': 18},
            3: {'pid': 129, 'status': None, 'global_rank': 19},
            4: {'pid': 143, 'status': None, 'global_rank': 20},
            5: {'pid': 155, 'status': None, 'global_rank': 21},
            6: {'pid': 167, 'status': None, 'global_rank': 22},
            7: {'pid': 176, 'status': None, 'global_rank': 23}
        }

    def test_valid_input(self):
        # 测试有效输入，期望返回 True
        result = check_monitor_res_valid(self.valid_rank_status_dict)
        self.assertEqual(result, True)

    def test_non_dict_input(self):
        # 测试输入不是字典的情况，期望返回 False
        non_dict_input = [1, 2, 3]
        result = check_monitor_res_valid(non_dict_input)
        self.assertEqual(result, False)

    def test_non_dict_info(self):
        # 测试某个 rank 的信息不是字典的情况，期望返回 False
        invalid_rank_status_dict = self.valid_rank_status_dict.copy()
        invalid_rank_status_dict[0] = [1, 2, 3]
        result = check_monitor_res_valid(invalid_rank_status_dict)
        self.assertEqual(result, False)

    def test_missing_key(self):
        # 测试某个 rank 的信息缺少必要键的情况，期望返回 False
        invalid_rank_status_dict = self.valid_rank_status_dict.copy()
        del invalid_rank_status_dict[0]['pid']
        result = check_monitor_res_valid(invalid_rank_status_dict)
        self.assertEqual(result, False)

    def test_non_int_pid(self):
        # 测试 'pid' 不是整数的情况，期望返回 False
        invalid_rank_status_dict = self.valid_rank_status_dict.copy()
        invalid_rank_status_dict[0]['pid'] = 'abc'
        result = check_monitor_res_valid(invalid_rank_status_dict)
        self.assertEqual(result, False)

    def test_non_int_status_non_none(self):
        # 测试 'status' 不是整数且不为 None 的情况，期望返回 False
        invalid_rank_status_dict = self.valid_rank_status_dict.copy()
        invalid_rank_status_dict[0]['status'] = 'xyz'
        result = check_monitor_res_valid(invalid_rank_status_dict)
        self.assertEqual(result, False)

    def test_non_int_global_rank(self):
        # 测试 'global_rank' 不是整数的情况，期望返回 False
        invalid_rank_status_dict = self.valid_rank_status_dict.copy()
        invalid_rank_status_dict[0][constants.GLOBAL_RANK_ID_KEY] = 'def'
        result = check_monitor_res_valid(invalid_rank_status_dict)
        self.assertEqual(result, False)


if __name__ == '__main__':
    unittest.main()