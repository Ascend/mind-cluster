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
from unittest import TestCase
from unittest.mock import patch, MagicMock

from taskd.python.framework.worker.worker import Worker

class TestWorker(TestCase):
    @patch('taskd.python.cython_api.cython_api.lib')
    def test_init_worker(self, mock_lib):
        mock_lib.InitWorker = MagicMock(return_value=0)
        os.environ["MS_NODE_RANK"] = "0"
        w = Worker(0, "ms")
        w.register_callback = MagicMock()
        res = w.init_worker(0)
        self.assertTrue(res)

    @patch('taskd.python.utils.log.run_log.info')
    @patch('taskd.python.cython_api.cython_api.lib')
    @patch('ctypes.CFUNCTYPE')
    def test_register_callback(self, mock_func, mock_lib, mock_log: MagicMock):
        mock_lib.RegisterSwitchCallback = MagicMock()
        mock_lib.RegisterStressTestCallback = MagicMock()
        w = Worker(0)
        w.register_callback()
        mock_log.assert_called_with("Successfully register callback func")

    @patch('taskd.python.cython_api.cython_api.lib')
    def test_start_up_monitor(self, mock_lib: MagicMock):
        w = Worker(0)
        mock_lib.StartMonitorClient = MagicMock(return_value=0)
        res = w._start_up_monitor()
        self.assertTrue(res)

        mock_lib.StartMonitorClient = MagicMock(return_value=1)
        res = w._start_up_monitor()
        self.assertFalse(res)