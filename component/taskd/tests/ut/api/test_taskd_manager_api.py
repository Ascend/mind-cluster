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
from taskd.api.taskd_manager_api import init_taskd_manager, start_taskd_manager
from taskd.python.framework.manager.manager import Manager

class TestTaskdManagerAPI(unittest.TestCase):
    @patch('taskd.api.taskd_manager_api.Manager')
    def test_init_taskd_manager_success(self, mock_manager):
        # mock the init_taskd_manager method of manager instance to return true
        mock_manager_instance = MagicMock()
        mock_manager_instance.init_taskd_manager.return_value = True
        mock_manager.return_value = mock_manager_instance

        config = {}
        result = init_taskd_manager(config)

        # verify manager class is instantiated correctly
        mock_manager.assert_called_once()
        # verify init_taskd_manager method is called
        mock_manager_instance.init_taskd_manager.assert_called_once_with(config)
        # verify return value is true
        self.assertEqual(result, True)

    @patch('taskd.api.taskd_manager_api.Manager')
    def test_init_taskd_manager_failure(self, mock_manager):
        # mock the init_taskd_manager method of manager instance to return false
        mock_manager_instance = MagicMock()
        mock_manager_instance.init_taskd_manager.return_value = False
        mock_manager.return_value = mock_manager_instance

        config = {}
        result = init_taskd_manager(config)

        # verify manager class is instantiated correctly
        mock_manager.assert_called_once()
        # verify init_taskd_manager method is called
        mock_manager_instance.init_taskd_manager.assert_called_once_with(config)
        # verify return value is false
        self.assertEqual(result, False)

    @patch('taskd.api.taskd_manager_api.taskd_manager')
    def test_start_taskd_manager_success(self, mock_taskd_manager):
        # mock taskd_manager's start_taskd_manager method to return true
        mock_taskd_manager.start_taskd_manager.return_value = True

        result = start_taskd_manager()

        # verify start_taskd_manager method is called
        mock_taskd_manager.start_taskd_manager.assert_called_once()
        # verify return value is true
        self.assertEqual(result, True)

    @patch('taskd.api.taskd_manager_api.taskd_manager', None)
    def test_start_taskd_manager_uninitialized(self):
        result = start_taskd_manager()
        # verify return value is false
        self.assertEqual(result, False)


if __name__ == '__main__':
    unittest.main()