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
import queue
from unittest.mock import MagicMock, patch, call
from taskd.python.framework.agent.pt_agent.pt_agent import PtAgent
from taskd.python.framework.common.type import AgentReportInfo
from taskd.python.framework.agent.base_agent.base_agent import REPORT_CODE
from torch.distributed.elastic.agent.server.api import WorkerState

class TestPtAgent(unittest.TestCase):
    def setUp(self):
        self.mock_worker = MagicMock()
        self.mock_worker.global_rank = 0
        self.mock_worker.id = 12345
        self.logger = MagicMock()
        self.mock_worker_group = MagicMock()
        self.mock_worker_group.workers = [self.mock_worker]
        self.mock_worker_group.spec.rdzv_handler.rank = 0
        self.mock_worker_group.spec.local_world_size = 1
        self.mock_worker_group.spec.monitor_interval = 1
        self.mock_worker_group.spec.role = 'trainer'
        self.mock_worker_group.spec.get_entrypoint_name.return_value = 'test_entrypoint'
        
        self.mock_cls = MagicMock()
        self.mock_cls._worker_group = self.mock_worker_group
        self.network_config = {'host': 'localhost', 'port': 8080}
        
        with patch('taskd.python.framework.agent.pt_agent.pt_agent.WorkerState', create=True) as mock_WorkerState:
            with patch('taskd.python.framework.agent.pt_agent.pt_agent.RunResult', create=True) as mock_RunResult:
                mock_WorkerState.SUCCEEDED = 'SUCCEEDED'
                mock_WorkerState.FAILED = 'FAILED'
                mock_WorkerState.HEALTHY = 'HEALTHY'
                mock_WorkerState.STOPPED = 'STOPPED'
                
                self.mock_run_result = MagicMock()
                self.mock_run_result.state = mock_WorkerState.HEALTHY
                self.mock_run_result.failures = {}
                mock_RunResult.return_value = self.mock_run_result
                
                self.agent = PtAgent(self.mock_cls, self.network_config, self.logger)
                self.agent._func_map = {
                    'START_ALL_WORKER': MagicMock(),
                    'KILL_WORKER': MagicMock(),
                    'MONITOR': MagicMock(return_value=self.mock_run_result)
                }
                self.agent.msg_queue = queue.Queue()

    def test_initialization(self):
        self.assertEqual(self.agent.node_rank, 0)
        self.assertEqual(self.agent.local_world_size, 1)
        self.assertEqual(self.agent.network_config, self.network_config)
        self.assertIn('START', self.agent.command_map)
        self.assertIn('STOP', self.agent.command_map)
        self.assertIn('EXIT', self.agent.command_map)
        self.assertIn('RESTART', self.agent.command_map)
        self.assertIn('GRACE_EXIT', self.agent.command_map)

    @patch('taskd.python.framework.agent.pt_agent.pt_agent.init_network_client')
    @patch('taskd.python.framework.agent.pt_agent.pt_agent.time.sleep')
    @patch.object(PtAgent, 'check_network')
    @patch.object(PtAgent, 'send_message_to_manager')
    @patch.object(PtAgent, 'handle_message')
    def test_invoke_run_success(self, mock_handle, mock_send, mock_check_net, mock_sleep, mock_init_net):
        with patch('taskd.python.framework.agent.pt_agent.pt_agent.WorkerState') as mock_WorkerState:
            mock_WorkerState.SUCCEEDED = 'SUCCEEDED'
            self.mock_run_result.state = mock_WorkerState.SUCCEEDED
            
            result = self.agent.invoke_run('trainer')
            
            mock_init_net.assert_called_once_with(self.network_config, self.agent.msg_queue, self.logger)
            mock_check_net.assert_called_once()
            self.agent._func_map['START_ALL_WORKER'].assert_called_once_with(self.mock_worker_group)
            mock_send.assert_has_calls([call('KEEP_ALIVE', 0, AgentReportInfo())])
            mock_handle.assert_called_once()
            mock_sleep.assert_called_once_with(1)
            self.assertEqual(result, self.mock_run_result)

    @patch('taskd.python.framework.agent.pt_agent.pt_agent.time.sleep')
    @patch('taskd.python.framework.agent.pt_agent.pt_agent.init_network_client')
    @patch.object(PtAgent, 'check_network')
    @patch.object(PtAgent, 'send_message_to_manager')
    @patch.object(PtAgent, 'handle_message')
    @patch.object(PtAgent, 'update_agent_info')
    @patch.object(PtAgent, 'report_fault_rank')
    def test_invoke_run_faulty_state(self, mock_report, mock_update_info, mock_handle, mock_send, \
        mock_check_net, mock_init_net, mock_sleep):
        self.mock_run_result.state = 'FAILED'
        
        self.agent._func_map['MONITOR'].side_effect = [self.mock_run_result, Exception('Break loop')]
        
        with self.assertRaises(Exception):
            self.agent.invoke_run('trainer')
            
        mock_report.assert_called_once_with(self.mock_run_result)

    def test_update_agent_info(self):
        self.agent.update_agent_info()
        
        self.assertEqual(self.agent.local_rank, [0])
        self.assertEqual(self.agent.pids, {0: 12345})
        self.assertEqual(self.agent.local_fault_rank, [])

    @patch.object(PtAgent, 'send_message_to_manager')
    @patch.object(PtAgent, 'check_new_fault')
    def test_report_fault_rank_new_fault(self, mock_check_new, mock_send):
        mock_check_new.return_value = True
        self.mock_run_result.failures = {0: 'error'}
        
        self.agent.report_fault_rank(self.mock_run_result)
        
        mock_check_new.assert_called_once_with([0])
        mock_send.assert_called_once_with('STATUS', REPORT_CODE, AgentReportInfo(fault_ranks=[0]))
        self.assertEqual(self.agent.local_fault_rank, [0])

    @patch.object(PtAgent, 'send_message_to_manager')
    @patch.object(PtAgent, 'check_new_fault')
    def test_report_fault_rank_no_new_fault(self, mock_check_new, mock_send):
        mock_check_new.return_value = False
        self.mock_run_result.failures = {0: 'error'}
        
        self.agent.report_fault_rank(self.mock_run_result)
        
        mock_check_new.assert_called_once_with([0])
        mock_send.assert_not_called()

    def test_initialize_workers(self):
        mock_msg = MagicMock()
        mock_msg.msg_type = 'START'
        mock_msg.extension = '3'
        
        self.agent.initialize_workers(mock_msg)
        
        self.assertEqual(self.mock_cls._remaining_restarts, 3)
        self.agent._func_map['START_ALL_WORKER'].assert_called_once_with(self.mock_worker_group)

    def test_stop_workers(self):
        mock_msg = MagicMock()
        mock_msg.msg_type = 'STOP'
        
        with patch('taskd.python.framework.agent.pt_agent.pt_agent.WorkerState') as mock_WorkerState:
            mock_WorkerState.STOPPED = 'STOPPED'
            self.agent.stop_workers(mock_msg)
            
            self.agent._func_map['KILL_WORKER'].assert_called_once_with(self.mock_worker_group)
            self.assertEqual(self.mock_worker_group.state, mock_WorkerState.STOPPED)

    def test_restart_workers(self):
        mock_msg = MagicMock()
        mock_msg.msg_type = 'RESTART'
        mock_msg.extension = '2'
        
        with patch('taskd.python.framework.agent.pt_agent.pt_agent.WorkerState') as mock_WorkerState:
            mock_WorkerState.STOPPED = 'STOPPED'
            self.agent.restart_workers(mock_msg)
            
            self.assertEqual(self.mock_cls._remaining_restarts, 2)
            self.agent._func_map['KILL_WORKER'].assert_called_once_with(self.mock_worker_group)
            self.assertEqual(self.mock_worker_group.state, mock_WorkerState.STOPPED)
            self.agent._func_map['START_ALL_WORKER'].assert_called_once_with(self.mock_worker_group)
            
    @patch.object(PtAgent, 'send_message_to_manager')
    @patch('taskd.python.framework.agent.pt_agent.pt_agent.exit')
    def test_exit_agent(self, mock_exit, mock_send):
        mock_msg = MagicMock()
        mock_msg.msg_type = 'EXIT'
        
        self.agent.exit_agent(mock_msg)
        
        self.agent._func_map['KILL_WORKER'].assert_called_once_with(self.mock_worker_group)
        mock_send.assert_called_once_with('STATUS', REPORT_CODE, AgentReportInfo())
        mock_exit.assert_called_once_with(1)

if __name__ == '__main__':
    unittest.main()
    