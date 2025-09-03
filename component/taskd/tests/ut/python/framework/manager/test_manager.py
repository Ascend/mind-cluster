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
import sys
import json
import os
import unittest
from unittest.mock import MagicMock, patch, call
from taskd.python.cython_api import cython_api
from taskd.python.utils.log import run_log
from taskd.python.framework.manager.manager import Manager

cur_dir = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(cur_dir, '../../'))

class TestManager(unittest.TestCase):
    def setUp(self):
        self.manager = Manager()
        self.config = {"key": "value"}
        os.environ["TTP_PORT"] = "1234"
        os.environ["POD_IP"] = "127.0.0.1"
        os.environ["WORLD_SIZE"] = "1"

    @patch.object(run_log, 'error')
    def test_init_taskd_manager_lib_none(self, mock_error):
        cython_api.lib = None
        result = self.manager.init_taskd_manager(self.config)
        self.assertFalse(result)
        mock_error.assert_called_once_with("the libtaskd.so has not been loaded")

    @patch.object(run_log, 'error')
    @patch.object(run_log, 'info')
    @patch.object(cython_api, 'lib')
    def test_init_taskd_manager_success(self, mock_lib, mock_info, mock_error):
        mock_func = MagicMock(return_value=0)
        mock_lib.InitTaskdManager = mock_func
        cython_api.lib = mock_lib

        result = self.manager.init_taskd_manager(self.config)

        self.assertTrue(result)
        config_str = json.dumps(self.config).encode('utf-8')
        mock_func.assert_called_once_with(config_str)
        mock_info.assert_called_once_with("successfully init taskd manager")
        mock_error.assert_not_called()

    @patch.object(run_log, 'error')
    @patch.object(run_log, 'warning')
    @patch.object(cython_api, 'lib')
    def test_init_taskd_manager_failure(self, mock_lib, mock_warning, mock_error):
        mock_func = MagicMock(return_value=1)
        mock_lib.InitTaskdManager = mock_func
        cython_api.lib = mock_lib

        result = self.manager.init_taskd_manager(self.config)

        self.assertFalse(result)
        mock_func.assert_called_once()
        mock_warning.assert_called_once_with("failed to init taskd manager with ret code:f1")
        mock_error.assert_not_called()

    @patch.object(run_log, 'error')
    def test_start_taskd_manager_lib_none(self, mock_error):
        cython_api.lib = None
        result = self.manager.start_taskd_manager()
        self.assertFalse(result)
        mock_error.assert_called_once_with("the libtaskd.so has not been loaded")

    @patch.object(run_log, 'error')
    @patch.object(run_log, 'info')
    @patch.object(cython_api, 'lib')
    def test_start_taskd_manager_success(self, mock_lib, mock_info, mock_error):
        mock_func = MagicMock(return_value=0)
        mock_lib.StartTaskdManager = mock_func
        cython_api.lib = mock_lib

        result = self.manager.start_taskd_manager()

        self.assertTrue(result)
        mock_func.assert_called_once_with()
        mock_info.assert_called_with("successfully start taskd manager")

    @patch.object(run_log, 'error')
    @patch.object(run_log, 'warning')
    @patch.object(cython_api, 'lib')
    def test_start_taskd_manager_failure(self, mock_lib, mock_warning, mock_error):
        mock_func = MagicMock(return_value=1)
        mock_lib.StartTaskdManager = mock_func
        cython_api.lib = mock_lib

        result = self.manager.start_taskd_manager()

        self.assertFalse(result)
        mock_func.assert_called_once_with()
        mock_warning.assert_called_with("failed to start taskd manager with ret code:f1")

    @patch.object(run_log, 'error')
    @patch.object(cython_api, 'lib')
    def test_start_taskd_manager_exception(self, mock_lib, mock_error):
        mock_func = MagicMock(side_effect=Exception("test error"))
        mock_lib.StartTaskdManager = mock_func
        cython_api.lib = mock_lib

        result = self.manager.start_taskd_manager()

        self.assertFalse(result)
        mock_func.assert_called_with()
        mock_error.assert_called_with("failed to start manager, error:test error")

    @patch('taskd.python.framework.manager.manager.threading.Thread')
    @patch('taskd.python.framework.manager.controller.init_controller')
    @patch('taskd.python.framework.manager.controller.backend_send_callback')
    @patch.object(run_log, 'error')
    @patch.object(run_log, 'info')
    @patch.object(cython_api, 'lib')
    def test_start_controller_success(self, mock_lib, mock_info, mock_error, mock_backend_send_callback, \
                                    mock_init_controller, mock_thread):
        mock_register_func = MagicMock()
        mock_lib.RegisterBackendCallback = mock_register_func
        cython_api.lib = mock_lib
        self.manager.start_controller()
        self.assertEqual(self.manager.callback, mock_backend_send_callback)
        self.assertIsNotNone(self.manager.c_callback)

        mock_thread.assert_called_once_with(target=mock_init_controller)
        mock_thread_instance = mock_thread.return_value
        self.assertTrue(mock_thread_instance.daemon)
        mock_thread_instance.start.assert_called_once()

        mock_register_func.assert_called_once_with(self.manager.c_callback)
        mock_info.assert_called_once_with("Successfully register controller callback")
        mock_error.assert_not_called()

if __name__ == '__main__':
    unittest.main()