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
import json
import time
from unittest.mock import MagicMock, patch, call
import uuid

from taskd.python.framework.agent.base_agent.base_agent import BaseAgent
from taskd.python.framework.common.type import MsgBody, MessageInfo, AgentReportInfo
from taskd.python.toolkit.fault_checker.fault_check import grace_exit_pids, stop_pids


class TestBaseAgent(unittest.TestCase):
    def setUp(self):
        class ConcreteAgent(BaseAgent):
            def initialize_workers(self, msg):
                pass
            def stop_workers(self, msg):
                pass
            def exit_agent(self, msg):
                pass
            def restart_workers(self, msg):
                pass

        self.agent = ConcreteAgent()
        self.mock_msg = MagicMock()
        self.mock_msg.msg_type = "TEST_COMMAND"

    def test_initialization(self):
        self.assertEqual(self.agent.node_rank, -1)
        self.assertEqual(self.agent.local_world_size, 0)
        self.assertEqual(self.agent.local_rank, [])
        self.assertEqual(self.agent.pids, {})
        self.assertEqual(self.agent.local_fault_rank, [])
        self.assertEqual(self.agent._func_map, {})
        self.assertEqual(self.agent.command_map, {})
        self.assertIsNone(self.agent.network_config)
        self.assertIsInstance(self.agent.msg_queue, queue.Queue)

    def test_register_callbacks(self):
        mock_func = MagicMock()
        self.agent.register_callbacks("TEST_OP", mock_func)
        self.assertIn("TEST_OP", self.agent._func_map)
        self.assertEqual(self.agent._func_map["TEST_OP"], mock_func)

    def test_check_new_fault_with_new_faults(self):
        self.agent.local_fault_rank = [1, 2]
        self.assertTrue(self.agent.check_new_fault([1, 2, 3]))
        self.assertTrue(self.agent.check_new_fault([3]))
        self.assertTrue(self.agent.check_new_fault([]))

    def test_check_new_fault_without_new_faults(self):
        self.agent.local_fault_rank = [1, 2]
        self.assertFalse(self.agent.check_new_fault([1, 2]))
        self.assertFalse(self.agent.check_new_fault([2, 1]))

    def test_handle_message_with_valid_command(self):
        mock_command_func = MagicMock()
        self.agent.command_map = {"TEST_COMMAND": mock_command_func}
        test_msg = MagicMock()
        test_msg.code = "TEST_COMMAND"
        self.agent.msg_queue.put(test_msg)

        self.agent.handle_message()
        mock_command_func.assert_called_once_with(test_msg)

    def test_handle_message_with_unknown_command(self):
        self.agent.command_map = {}
        test_msg = MagicMock()
        test_msg.MsgType = "UNKNOWN_COMMAND"
        self.agent.msg_queue.put(test_msg)

        with self.assertRaises(TypeError):
            self.agent.handle_message()

    @patch('taskd.python.utils.log.run_log.info')
    @patch('taskd.python.utils.log.run_log.error')
    def test_grace_exit_success(self, mock_error_log, mock_info_log):
        self.agent.pids = {"test_pid": 12345}
        mock_grace_exit = MagicMock()
        mock_stop = MagicMock()

        with patch('taskd.python.framework.agent.base_agent.base_agent.grace_exit_pids', mock_grace_exit), \
             patch('taskd.python.framework.agent.base_agent.base_agent.stop_pids', mock_stop):
            self.agent.grace_exit(self.mock_msg)
            mock_info_log.assert_called_once()
            mock_grace_exit.assert_called_once_with({"test_pid": 12345})
            mock_stop.assert_called_once_with({"test_pid": 12345})
            mock_error_log.assert_not_called()

    @patch('taskd.python.utils.log.run_log.error')
    def test_grace_exit_with_exception(self, mock_error_log):
        self.agent.pids = {"test_pid": 12345}
        test_exception = Exception("Test exception")
        mock_grace_exit = MagicMock(side_effect=test_exception)
        mock_stop = MagicMock()

        with patch('taskd.python.framework.agent.base_agent.base_agent.grace_exit_pids', mock_grace_exit), \
             patch('taskd.python.framework.agent.base_agent.base_agent.stop_pids', mock_stop):
            self.agent.grace_exit(self.mock_msg)
            mock_error_log.assert_called_once_with('grace_exit encountered an exception: Test exception')
            mock_stop.assert_called_once_with({"test_pid": 12345})

    @patch('time.time')
    @patch('uuid.uuid4')
    @patch('taskd.python.framework.agent.base_agent.base_agent.network_send_message')
    def test_send_message_to_manager(self, mock_send, mock_uuid4, mock_time):
        mock_time.return_value = 1
        mock_uuid4.return_value = uuid.UUID('12345678-1234-5678-1234-567812345678')
        test_report_info = AgentReportInfo(restart_times=0)

        self.agent.send_message_to_manager("TEST_CMD", 200, test_report_info)

        mock_send.assert_called_once()
        sent_msg = mock_send.call_args[0][0]
        self.assertIsInstance(sent_msg, MessageInfo)
        self.assertEqual(sent_msg.uuid, '12345678-1234-5678-1234-567812345678')
        self.assertEqual(sent_msg.biz_type, 'DEFAULT')
        self.assertEqual(sent_msg.dst, {
            "role": "Mgr",
            "server_rank": "0",
            "process_rank": "-1"
        })

        body = json.loads(sent_msg.body)
        self.assertEqual(body['msg_type'], 'TEST_CMD')
        self.assertEqual(body['code'], 200)
        self.assertEqual(json.loads(body['message']), {'fault_ranks': [], 'restart_times': 0})

    @patch('taskd.python.utils.log.run_log.info')
    @patch('taskd.python.utils.log.run_log.error')
    @patch('taskd.python.framework.agent.base_agent.base_agent.get_msg_network_instance')
    def test_check_network_success(self, mock_get_instance, mock_error_log, mock_info_log):
        mock_network_instance = MagicMock()
        mock_get_instance.side_effect = [None, None, mock_network_instance]

        self.agent.check_network()
        mock_info_log.assert_any_call('waiting for message manager')
        mock_info_log.assert_called_with('message manager is ready')
        mock_error_log.assert_not_called()

    @patch('taskd.python.utils.log.run_log.error')
    @patch('taskd.python.framework.agent.base_agent.base_agent.get_msg_network_instance')
    def test_check_network_timeout(self, mock_get_instance, mock_error_log):
        mock_get_instance.return_value = None

        with patch('time.sleep') as mock_sleep, self.assertRaises(ValueError) as context:
            self.agent.check_network()
            self.assertEqual(mock_sleep.call_count, 61)
            mock_error_log.assert_called_once_with('waiting for message manager timeout')
            self.assertIn('initialization message_manager timeout', str(context.exception))


if __name__ == '__main__':
    unittest.main()
