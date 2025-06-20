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
import json
import ctypes
import queue
from unittest.mock import patch, MagicMock, call
from venv import logger
from taskd.python.framework.agent.base_agent.agent_network import AgentMessageManager, init_network_client, init_message_manager, get_message_manager
from taskd.python.framework.common.type import MsgBody, MessageInfo, Position, DEFAULT_BIZTYPE, NetworkConfig
from taskd.python.toolkit.constants.constants import SEND_RETRY_TIMES

class TestAgentMessageManager(unittest.TestCase):
    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    def setUp(self, mock_cython):
        self.mock_lib = MagicMock()
        mock_cython.lib = self.mock_lib
        self.mock_lib.InitNetwork.return_value = ctypes.c_void_p(1)
        self.mock_queue = MagicMock()
        self.network_config = NetworkConfig(
            pos=Position(
                role='test',
                server_rank='0',
                process_rank='-1'
            ),
            upstream_addr="127.0.0.1:8080",
            listen_addr='',
            enable_tls=False,
            tls_conf=None,
        )
        self.logger = MagicMock()
        
        AgentMessageManager.instance = None

    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    def test_singleton_pattern(self, mock_cython):
        mock_cython.lib = self.mock_lib
        
        instance1 = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        instance2 = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        self.assertIs(instance1, instance2)
        self.assertIs(AgentMessageManager.instance, instance1)

    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    @patch('taskd.python.framework.agent.base_agent.agent_network.run_log')
    def test_init_network_success(self, mock_log, mock_cython):
        mock_cython.lib = self.mock_lib
        self.mock_lib.InitNetwork.return_value = ctypes.c_void_p(1)
        
        agent = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        
        self.mock_lib.InitNetwork.assert_called_once()
        args, _ = self.mock_lib.InitNetwork.call_args
        config_json = json.loads(args[0].decode('utf-8'))
        self.assertEqual(config_json['pos']['server_rank'], '0')
        self.assertIsNotNone(agent.get_network_instance())

    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    @patch('taskd.python.framework.agent.base_agent.agent_network.run_log')
    def test_init_network_failure(self, mock_log, mock_cython):
        self.mock_lib = MagicMock()
        mock_cython.lib = self.mock_lib
        self.mock_lib.InitNetwork.return_value = None
        
        with self.assertRaises(Exception) as context:
            AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        self.assertEqual(str(context.exception), 'init_network_func failed!')

    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    def test_register_message(self, mock_cython):
        mock_cython.lib = self.mock_lib
        
        agent = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        agent.send_message = MagicMock()
        
        agent.register('0')
        
        agent.send_message.assert_called_once()
        sent_msg = agent.send_message.call_args[0][0]
        self.assertIsInstance(sent_msg, MessageInfo)
        self.assertEqual(sent_msg.dst.role, 'Mgr')
        self.assertEqual(sent_msg.dst.server_rank, '0')
        body = json.loads(sent_msg.body)
        self.assertEqual(body['msg_type'], 'REGISTER')

    @patch('taskd.python.framework.agent.base_agent.agent_network.time.sleep')
    @patch('taskd.python.framework.agent.base_agent.agent_network.run_log')
    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    def test_send_message_success(self, mock_cython, mock_log, mock_sleep):
        mock_cython.lib = self.mock_lib
        
        agent = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        test_msg = MessageInfo(uuid='test-uuid', biz_type=DEFAULT_BIZTYPE, 
                              dst=Position(role='test', server_rank='0', process_rank='-1'),
                              body='{}')
        self.mock_lib.SyncSendMessage.return_value = 0
        
        agent.send_message(test_msg)
        
        self.mock_lib.SyncSendMessage.assert_called_once()
        mock_log.info.assert_any_call(f'agent send message success, msg: {test_msg.uuid}')
        mock_sleep.assert_not_called()

    @patch('taskd.python.framework.agent.base_agent.agent_network.time.sleep')
    @patch('taskd.python.framework.agent.base_agent.agent_network.run_log')
    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    def test_send_message_retry(self, mock_cython, mock_log, mock_sleep):
        mock_cython.lib = self.mock_lib
        
        agent = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        test_msg = MessageInfo(uuid='test-uuid', biz_type=DEFAULT_BIZTYPE, 
                              dst=Position(role='test', server_rank='0', process_rank='-1'),
                              body='{}')
        self.mock_lib.SyncSendMessage.side_effect = [1, 2, 0]
        
        agent.send_message(test_msg)
        
        self.assertEqual(self.mock_lib.SyncSendMessage.call_count, 3)
        self.assertEqual(mock_sleep.call_count, 2)
        mock_log.warning.assert_any_call('agent send message failed, result: 1')
        mock_log.warning.assert_any_call('agent send message failed, result: 2')

    @patch('taskd.python.framework.agent.base_agent.agent_network.time.sleep')
    @patch('taskd.python.framework.agent.base_agent.agent_network.run_log')
    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    def test_send_message_max_retries(self, mock_cython, mock_log, mock_sleep):
        mock_cython.lib = self.mock_lib
        
        agent = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        test_msg = MessageInfo(uuid='test-uuid', biz_type=DEFAULT_BIZTYPE, 
                              dst=Position(role='test', server_rank='0', process_rank='-1'), 
                              body='{}')
        self.mock_lib.SyncSendMessage.return_value = 1
        
        agent.send_message(test_msg)
        
        self.assertEqual(self.mock_lib.SyncSendMessage.call_count, SEND_RETRY_TIMES)
        mock_log.error.assert_called_with(f'agent send message failed, msg: {test_msg.uuid}')

    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    @patch('taskd.python.framework.agent.base_agent.agent_network.run_log')
    def test_parse_valid_message(self, mock_log, mock_cython):
        mock_cython.lib = self.mock_lib
        
        agent = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        test_msg = {
            "body": json.dumps({
                "msg_type": "TEST",
                "code": 200,
                "message": "OK",
                "extension": {}
            })
        }
        
        result = agent._parse_msg(json.dumps(test_msg))
        
        self.assertIsInstance(result, MsgBody)
        self.assertEqual(result.msg_type, "TEST")
        self.assertEqual(result.code, 200)

    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    @patch('taskd.python.framework.agent.base_agent.agent_network.run_log')
    def test_parse_invalid_message(self, mock_log, mock_cython):
        mock_cython.lib = self.mock_lib
        
        agent = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        
        result = agent._parse_msg('invalid json')
        self.assertIsNone(result)
        mock_log.error.assert_called()
        
        result = agent._parse_msg(json.dumps({"InvalidKey": "value"}))
        self.assertIsNone(result)

    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    @patch('taskd.python.framework.agent.base_agent.agent_network.time.sleep')
    def test_receive_normal_message(self, mock_sleep, mock_cython):
        mock_cython.lib = self.mock_lib
        
        agent = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        test_msg = json.dumps({
            "body": json.dumps({
                "msg_type": "TEST",
                "code": 200,
                "message": "OK",
                "extension": {}
            })
        }).encode('utf-8')
        exit_msg = json.dumps({
            "body": json.dumps({
                "msg_type": "exit",
                "code": 0,
                "message": "",
                "extension": {}
            })
        }).encode('utf-8')
        
        test_buffer = ctypes.create_string_buffer(test_msg)
        exit_buffer = ctypes.create_string_buffer(exit_msg)
        
        self.mock_lib.ReceiveMessageC.side_effect = [
            test_buffer,
            exit_buffer,
            None
        ]
        agent._parse_msg = MagicMock(side_effect=[
            MsgBody(msg_type="TEST", code=200, message="", extension={}),
            MsgBody(msg_type="exit", code=0, message="", extension={})
        ])
        
        agent.receive_message()
        
        self.assertEqual(self.mock_queue.put.call_count, 2)
        self.mock_queue.put.assert_has_calls([
            call(MsgBody(msg_type="TEST", code=200, message="", extension={})),
            call(MsgBody(msg_type="exit", code=0, message="", extension={}))
        ])
        
        self.mock_lib.FreeCMemory.assert_any_call(test_buffer)
        self.mock_lib.FreeCMemory.assert_any_call(exit_buffer)
        self.mock_lib.DestroyNetwork.assert_called_once_with(agent.get_network_instance())
        
    @patch('taskd.python.framework.agent.base_agent.agent_network.cython_api')
    def test_receive_exit_message(self, mock_cython):
        mock_cython.lib = self.mock_lib
        
        agent = AgentMessageManager(self.network_config, self.mock_queue, self.logger)
        exit_msg = MsgBody(msg_type="exit", code=0, message="", extension={})
        agent._parse_msg = MagicMock(return_value=exit_msg)
        self.mock_lib.ReceiveMessageC.return_value = ctypes.create_string_buffer(b'{"Body": {}}')
        
        agent.receive_message()
        
        self.mock_lib.DestroyNetwork.assert_called_once_with(agent.get_network_instance())

class TestAgentNetworkFunctions(unittest.TestCase):
    @patch('taskd.python.framework.agent.base_agent.agent_network.threading.Thread')
    def test_init_network_client(self, mock_thread):
        mock_queue = MagicMock()
        network_config = MagicMock()
        mock_logger = MagicMock()
        
        init_network_client(network_config, mock_queue, mock_logger)
        
        mock_thread.assert_called_once()
        args, kwargs = mock_thread.call_args
        self.assertEqual(kwargs['target'], init_message_manager)
        self.assertEqual(kwargs['args'], (network_config, mock_queue, mock_logger))

    @patch('taskd.python.framework.agent.base_agent.agent_network.time.sleep')
    @patch('taskd.python.framework.agent.base_agent.agent_network.run_log')
    @patch('taskd.python.framework.agent.base_agent.agent_network.AgentMessageManager')
    def test_init_message_manager_success(self, mock_manager_cls, mock_log, mock_sleep):
        mock_queue = MagicMock()
        network_config = MagicMock()
        network_config.pos.server_rank = '0'
        mock_manager = MagicMock()
        mock_manager_cls.return_value = mock_manager
        mock_manager.get_network_instance.side_effect = [None, None, ctypes.c_void_p(1)]
        mock_logger = MagicMock()
        init_message_manager(network_config, mock_queue, mock_logger)
        
        mock_manager_cls.assert_called_once_with(network_config, mock_queue, mock_logger)
        mock_manager.register.assert_called_once_with('0')
        mock_manager.receive_message.assert_called_once()
        self.assertEqual(mock_sleep.call_count, 2)
        mock_log.info.assert_any_call('init message manager success!')

    @patch('taskd.python.framework.agent.base_agent.agent_network.time.sleep')
    @patch('taskd.python.framework.agent.base_agent.agent_network.run_log')
    @patch('taskd.python.framework.agent.base_agent.agent_network.AgentMessageManager')
    def test_init_message_manager_timeout(self, mock_manager_cls, mock_log, mock_sleep):
        mock_queue = MagicMock()
        network_config = MagicMock()
        mock_manager = MagicMock()
        mock_logger = MagicMock()
        mock_manager_cls.return_value = mock_manager
        mock_manager.get_network_instance.return_value = None
        
        init_message_manager(network_config, mock_queue, mock_logger)
        
        self.assertEqual(mock_sleep.call_count, 61)
        mock_log.error.assert_called_with('init message manager failed!')

if __name__ == '__main__':
    unittest.main()
    