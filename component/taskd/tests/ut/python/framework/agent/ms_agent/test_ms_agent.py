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
import os
import time
import json
from unittest.mock import patch, MagicMock, call
from taskd.python.framework.agent.ms_agent.ms_agent import MsAgent
from taskd.python.toolkit.constants import constants
from taskd.python.framework.common.type import AgentReportInfo

class TestMsAgent(unittest.TestCase):
    def setUp(self):
        self.network_config = {'test': 'config'}
        self.logger = MagicMock()
        self.agent = MsAgent(self.network_config, self.logger)
        self.agent._func_map = {
            'KILL_WORKER': MagicMock(),
            'START_ALL_WORKER': MagicMock(),
            'MONITOR': MagicMock(),
            'START_WORKER_LIST': MagicMock()
        }
        self.agent.msg_queue = MagicMock()
        self.agent.local_fault_rank = []

    @patch('taskd.python.framework.agent.ms_agent.ms_agent.os.getenv')
    def test_init(self, mock_getenv):
        mock_getenv.return_value = '0'
        agent = MsAgent(self.network_config, self.logger)
        
        self.assertEqual(agent.network_config, self.network_config)
        self.assertEqual(agent.monitor_interval, 1)
        self.assertEqual(agent.node_rank, '0')
        self.assertEqual(agent.command_map.keys(), {
                         constants.STARTAGENTCODE, constants.STOPWORKERSCODE, constants.EXITAGENTCODE,
                         constants.RESTARTAGENTCODE, 'GRACE_EXIT',
                         constants.RESTARTWORKERCODE})

    @patch('taskd.python.framework.agent.ms_agent.ms_agent.calculate_global_rank')
    @patch('taskd.python.framework.agent.ms_agent.ms_agent.init_network_client')
    @patch('taskd.python.framework.agent.ms_agent.ms_agent.time.sleep')
    @patch.object(MsAgent, 'check_network')
    @patch.object(MsAgent, 'handle_message')
    @patch.object(MsAgent, 'update_rank_status')
    @patch.object(MsAgent, 'report_fault_rank')
    @patch.object(MsAgent, 'start_worker')
    def test_start(self, mock_start_worker, mock_report,mock_update, mock_handle, mock_check_net, mock_sleep,
                   mock_init_net, mock_calc_rank):
        mock_calc_rank.return_value = [0, 1]
        self.agent._func_map['MONITOR'].return_value = {
            '0': {
                constants.RANK_PID_KEY: 1,
                constants.RANK_STATUS_KEY: constants.RANK_STATUS_OK,
                constants.GLOBAL_RANK_ID_KEY: 1,
            }
        }
        mock_update.return_value = []
        
        with patch.object(MsAgent, 'send_message_to_manager', side_effect=[None, Exception('Break loop')]):
            with self.assertRaises(Exception):
                self.agent.start()

        mock_calc_rank.assert_called_once()
        mock_start_worker.assert_called_once()
        mock_init_net.assert_called_once_with(self.network_config, self.agent.msg_queue, self.logger)
        mock_check_net.assert_called_once()
        
        mock_sleep.assert_called_once_with(1)
        mock_handle.assert_called_once()
        mock_update.assert_called_once()

    def test_update_rank_status_all_healthy(self):
        rank_status = {
            '0': {constants.RANK_STATUS_KEY: constants.RANK_STATUS_OK, 
                  constants.RANK_PID_KEY: 123, 
                  constants.GLOBAL_RANK_ID_KEY: 0},
            '1': {constants.RANK_STATUS_KEY: constants.RANK_STATUS_OK, 
                  constants.RANK_PID_KEY: 456, 
                  constants.GLOBAL_RANK_ID_KEY: 1}
        }
        
        fault_ranks = self.agent.update_rank_status(rank_status)
        
        self.assertEqual(self.agent.rank_status, MsAgent.RANK_STATUS_HEALTHY)
        self.assertEqual(fault_ranks, [])
        self.assertEqual(self.agent.rank_pids, [123, 456])
        self.assertEqual(self.agent.node_global_rank_ids, [0, 1])

    def test_update_rank_status_with_fault(self):
        rank_status = {
            '0': {constants.RANK_STATUS_KEY: constants.RANK_STATUS_OK, 
                  constants.RANK_PID_KEY: 123, 
                  constants.GLOBAL_RANK_ID_KEY: 0},
            '1': {constants.RANK_STATUS_KEY: 'FAILED', 
                  constants.RANK_PID_KEY: 456, 
                  constants.GLOBAL_RANK_ID_KEY: 1}
        }
        
        fault_ranks = self.agent.update_rank_status(rank_status)
        
        self.assertEqual(self.agent.rank_status, MsAgent.RANK_STATUS_UNHEALTHY)
        self.assertEqual(fault_ranks, [1])

    @patch('time.time')
    @patch.object(MsAgent, 'check_new_fault')
    @patch.object(MsAgent, 'send_message_to_manager')
    def test_report_fault_rank_new_fault(self, mock_send, mock_check_new, mock_time):
        mock_time.return_value = 1
        mock_check_new.return_value = True
        fault_ranks = [1, 2]
        
        self.agent.report_fault_rank(fault_ranks)
        
        mock_check_new.assert_called_once_with(fault_ranks)
        mock_send.assert_called_once_with('STATUS', constants.REPORT_CODE, AgentReportInfo(fault_ranks=fault_ranks),
        {"REPORT_FAULT_TIME": "1"})
        self.assertEqual(self.agent.local_fault_rank, fault_ranks)

    @patch.object(MsAgent, 'check_new_fault')
    @patch.object(MsAgent, 'send_message_to_manager')
    def test_report_fault_rank_no_new_fault(self, mock_send, mock_check_new):
        mock_check_new.return_value = False
        
        self.agent.report_fault_rank([1, 2])
        
        mock_send.assert_not_called()

    def test_initialize_workers(self):
        mock_msg = MagicMock()
        mock_msg.msg_type = 'START'
        
        self.agent.initialize_workers(mock_msg)
        self.agent._func_map['START_ALL_WORKER'].assert_called_once()
        
    @patch('taskd.python.framework.agent.ms_agent.ms_agent.MsAgent.get_fault_pids')
    @patch('taskd.python.framework.agent.ms_agent.ms_agent.MsAgent.get_fault_local_ranks')
    def test_stop_workers(self, mock_get_fault_local_ranks, mock_get_fault_pids):
        mock_msg = MagicMock()
        mock_msg.code = 'STOP'
        mock_msg.message = json.dumps([1, 2])
        mock_get_fault_local_ranks.return_value = [1, 2]
        mock_get_fault_pids.return_value = [1, 2]
        
        self.agent.stop_workers(mock_msg)
        self.agent._func_map['KILL_WORKER'].assert_called_once_with([1, 2])

    def test_restart_workers(self):
        mock_msg = MagicMock()
        
        self.agent.restart_workers(mock_msg)
        expected_calls = [
            call([constants.KILL_ALL_WORKER_CALLBACK_NAME]),
            call()
        ]
        self.agent._func_map['KILL_WORKER'].assert_has_calls([call([constants.KILL_ALL_WORKERS])])
        self.agent._func_map['START_ALL_WORKER'].assert_called_once()

    @patch('taskd.python.framework.agent.ms_agent.ms_agent.exit')
    @patch('taskd.python.framework.agent.base_agent.base_agent.network_send_message')
    def test_exit_agent(self, mock_network_send, mock_exit):
        mock_msg = MagicMock()
        
        self.agent.exit_agent(mock_msg)
        self.agent._func_map['KILL_WORKER'].assert_called_once_with([constants.KILL_ALL_WORKERS])
        mock_network_send.assert_called_once()
        mock_exit.assert_called_once_with(1)

if __name__ == '__main__':
    unittest.main()
