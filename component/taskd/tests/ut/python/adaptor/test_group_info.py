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
from unittest.mock import patch, MagicMock
import os

from taskd.python.adaptor.pytorch.group_info import get_save_path, get_group_info, dump_group_info


class TestGroupInfoFunctions(unittest.TestCase):

    @patch('os.getenv')
    @patch('os.path.exists')
    @patch('os.makedirs')
    def test_get_save_path(self, mock_makedirs, mock_exists, mock_getenv):
        mock_getenv.side_effect = [None, None]
        result = get_save_path(1)
        self.assertEqual(result, "")

        mock_getenv.side_effect = ['job_id', 'base_dir']
        mock_exists.return_value = True
        result = get_save_path(1)
        self.assertEqual(result, os.path.join('base_dir', 'job_id', '1'))

        mock_getenv.side_effect = ['job_id', 'base_dir']
        mock_exists.return_value = False
        mock_makedirs.side_effect = OSError('Test error')
        result = get_save_path(1)
        self.assertEqual(result, "")

if __name__ == '__main__':
    unittest.main()