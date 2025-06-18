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
import signal
from unittest.mock import patch, MagicMock
import torch.distributed.elastic.agent.server.api
import torch.distributed.elastic.multiprocessing.api
from taskd.python.adaptor.patch.torch_patch import patch_torch_method, patch_invoke_run, patch_default_signal
from taskd.python.toolkit.constants.constants import SLEEP_GAP

class TestTorchPatch(unittest.TestCase):
    def setUp(self):
        self.original_invoke_run = torch.distributed.elastic.agent.server.api.SimpleElasticAgent._invoke_run
        self.original_get_signal = torch.distributed.elastic.multiprocessing.api._get_default_signal

    def tearDown(self):
        torch.distributed.elastic.agent.server.api.SimpleElasticAgent._invoke_run = self.original_invoke_run
        torch.distributed.elastic.multiprocessing.api._get_default_signal = self.original_get_signal

    def test_patch_torch_method(self):
        patch_torch_method()
        self.assertEqual(torch.distributed.elastic.agent.server.api.SimpleElasticAgent._invoke_run, patch_invoke_run)
        self.assertEqual(torch.distributed.elastic.multiprocessing.api._get_default_signal, patch_default_signal)

    @patch('taskd.python.adaptor.patch.torch_patch.threading.Thread')
    @patch('taskd.python.adaptor.patch.torch_patch.init_taskd_proxy')
    @patch('taskd.python.adaptor.patch.torch_patch.init_taskd_agent')
    @patch('taskd.python.adaptor.patch.torch_patch.register_func')
    @patch('taskd.python.adaptor.patch.torch_patch.start_taskd_agent')
    def test_patch_invoke_run(self, mock_start_agent, mock_register, mock_init_agent, mock_init_proxy, mock_thread):
        mock_self = MagicMock()
        mock_thread_instance = MagicMock()
        mock_thread.return_value = mock_thread_instance
        mock_start_agent.return_value = 'test_result'

        result = patch_invoke_run(mock_self)

        mock_thread.assert_called_once()
        mock_thread_instance.start.assert_called_once()

        mock_init_agent.assert_called_once()

        self.assertEqual(mock_register.call_count, 4)
        expected_calls = [
            unittest.mock.call('KILL_WORKER', mock_self._stop_workers),
            unittest.mock.call('START_ALL_WORKER', mock_self._initialize_workers),
            unittest.mock.call('MONITOR', mock_self._monitor_workers),
            unittest.mock.call('RESTART', mock_self._restart_workers)
        ]
        mock_register.assert_has_calls(expected_calls, any_order=False)

        mock_start_agent.assert_called_once()
        self.assertEqual(result, 'test_result')

    @patch('taskd.python.adaptor.patch.torch_patch.time.sleep')
    def test_patch_default_signal(self, mock_sleep):
        result = patch_default_signal()

        mock_sleep.assert_called_once_with(SLEEP_GAP)
        self.assertEqual(result, signal.SIGKILL)

if __name__ == '__main__':
    unittest.main()
