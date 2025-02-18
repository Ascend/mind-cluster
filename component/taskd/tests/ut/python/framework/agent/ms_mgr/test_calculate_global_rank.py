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

import os
import unittest
from unittest.mock import patch
from taskd.python.framework.agent.ms_mgr.MsUtils import calculate_global_rank


class TestCalculateGlobalRank(unittest.TestCase):
    @patch('os.getenv')
    def test_valid_env_vars(self, mock_getenv):
        # 模拟环境变量
        mock_getenv.side_effect = ['2', '3']
        result = calculate_global_rank()
        expected = [3 * 2 + 0, 3 * 2 + 1]
        self.assertEqual(result, expected)

    @patch('os.getenv')
    def test_missing_env_vars(self, mock_getenv):
        # 模拟缺少环境变量
        mock_getenv.side_effect = [None, '3']
        result = calculate_global_rank()
        self.assertEqual(result, [])

    @patch('os.getenv')
    def test_invalid_env_vars(self, mock_getenv):
        # 模拟无效的环境变量
        mock_getenv.side_effect = ['abc', '3']
        result = calculate_global_rank()
        self.assertEqual(result, [])

if __name__ == '__main__':
    unittest.main()