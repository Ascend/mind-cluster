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
import time
from taskd.python.toolkit.fault_checker.fault_check import FaultStatus
from taskd.python.toolkit.constants import constants
from taskd.python.framework.agent.ms_mgr.MsRunPlugin import MSRunPlugin
from taskd.python.toolkit.recover_module import shared_data

class TestMSRunPlugin(unittest.TestCase):
    def setUp(self):
        self.plugin = MSRunPlugin()
        self.plugin.register_callbacks("START_ALL_WORKER", MagicMock())
        self.plugin.register_callbacks("KILL_WORKER", MagicMock())
        self.plugin.register_callbacks("MONITOR", MagicMock(return_value={}))

    def test_register_callbacks(self):
        func = MagicMock()
        self.plugin.register_callbacks("TEST_OP", func)
        self.assertEqual(self.plugin._MSRunPlugin__funcMap["TEST_OP"], func)

    @patch('time.sleep')
    def test_start_mindspore_workers(self, mock_sleep):
        self.plugin.start_mindspore_workers()
        self.plugin._MSRunPlugin__funcMap["START_ALL_WORKER"].assert_called_once()

    @patch('os.getenv', return_value="0")
    def test_init_grpc_client_if_needed(self, mock_getenv):
        with patch.object(self.plugin, '_init_grpc_client_if_needed') as mock_method:
            self.plugin._init_grpc_client_if_needed()
            mock_method.assert_called_once()

    def test_handle_grace_exit(self):
        self.plugin.grace_exit = 1
        result = self.plugin._handle_grace_exit()
        self.assertEqual(result, True)
        self.plugin._MSRunPlugin__funcMap["KILL_WORKER"].assert_called_once_with([constants.KILL_ALL_WORKERS])

    def test_handle_fault_status(self):
        fault_status = FaultStatus([], True, False, False)
        with self.assertRaises(SystemExit):
            self.plugin._handle_fault_status(fault_status)
        self.plugin._MSRunPlugin__funcMap["KILL_WORKER"].assert_called_once_with([constants.KILL_ALL_WORKERS])

    @patch('time.sleep')
    def test_handle_process_fault(self, mock_sleep):
        fault_status = FaultStatus([], False, False, True)
        with patch.object(self.plugin, 'all_fault_has_recovered', return_value=True):
            result = self.plugin._handle_process_fault(fault_status)
            self.assertEqual(result, True)
            self.plugin._MSRunPlugin__funcMap["KILL_WORKER"].assert_called_once_with([constants.KILL_ALL_WORKERS])

    @patch('time.sleep')
    def test_handle_hardware_fault(self, mock_sleep):
        fault_status = FaultStatus([], False, True, False)
        with patch.object(self.plugin, 'all_fault_has_recovered', return_value=False):
            result = self.plugin._handle_hardware_fault(fault_status)
            self.assertEqual(result, True)
            self.plugin._MSRunPlugin__funcMap["KILL_WORKER"].assert_called_once_with([constants.KILL_ALL_WORKERS])

    @patch('time.sleep')
    @patch.object(shared_data.shared_data_inst, 'set_kill_flag')
    def test_handle_all_process_succeed(self, mock_set_kill_flag, mock_sleep):
        self.plugin.all_rank_succeed = True
        with self.assertRaises(SystemExit):
            self.plugin._handle_all_process_succeed()
        mock_set_kill_flag.assert_called_once_with(True)

    def test_handle_exist_unhealthy_process(self):
        self.plugin.rank_status = self.plugin.RankStatusUNHEALTHY
        with self.assertRaises(SystemExit):
            self.plugin._handle_exist_unhealthy_process()
        self.plugin._MSRunPlugin__funcMap["KILL_WORKER"].assert_called_once_with([constants.KILL_ALL_WORKERS])

    @patch('time.sleep')
    def test_update_rank_status(self, mock_sleep):
        rank_status_dict = {
            0: {
                constants.RANK_PID_KEY: 1,
                constants.RANK_STATUS_KEY: constants.rank_status_ok,
                constants.GLOBAL_RANK_ID_KEY: 0
            }
        }
        self.plugin.update_rank_status(rank_status_dict)
        self.assertEqual(self.plugin.rank_info, rank_status_dict)
        self.assertEqual(self.plugin.rank_pids, [1])
        self.assertEqual(self.plugin.node_global_rank_ids, [0])
        self.assertEqual(self.plugin.rank_status, self.plugin.RankStatusHEALTHY)


if __name__ == '__main__':
    unittest.main()