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
import os
import sys
import ctypes
from unittest.mock import MagicMock, patch
from taskd.python.utils.log import run_log
from taskd.python.toolkit.constants import constants
from taskd.python.cython_api import cython_api
from taskd.python.framework.manager.controller import (
    init_controller,
    register_callback_func,
    backend_send_callback,
    send_msg_to_controller,
    restart_controller,
    report_stop_complete,
    report_recover_strategy,
    report_recover_status,
    report_process_fault,
    controller_send_to_backend,
    ControllerMessage,
    CallBackFuncs,
    action_func_map
)


TEST_ACTION = "test_action"
TEST_MESSAGE = "test message"
TEST_STRATEGY = "test_strategy"
TEST_PARAM = "test_params"


class TestController(unittest.TestCase):

    def setUp(self):
        self.original_env = dict(os.environ)
        for env_var in [
            constants.WORLD_SIZE,
            constants.MS_WORKER_NUM,
            constants.PROCESS_RECOVER,
            constants.HIGH_AVAILABILITY_STRATEGY,
            constants.POD_IP,
            constants.TTP_PORT
        ]:
            if env_var in os.environ:
                del os.environ[env_var]

        self.patcher_log_error = patch.object(run_log, 'error')
        self.patcher_log_info = patch.object(run_log, 'info')
        self.patcher_log_warning = patch.object(run_log, 'warning')

        self.mock_log_error = self.patcher_log_error.start()
        self.mock_log_info = self.patcher_log_info.start()
        self.mock_log_warning = self.patcher_log_warning.start()

        self.patcher_tft_init_controller = patch('taskd.python.framework.manager.controller.tft_init_controller')
        self.patcher_tft_start_controller = patch('taskd.python.framework.manager.controller.tft_start_controller')
        self.patcher_tft_register_mindx_callback = patch('taskd.python.framework.manager.controller.tft_register_mindx_callback')
        self.patcher_tft_destroy_controller = patch('taskd.python.framework.manager.controller.tft_destroy_controller')

        self.mock_tft_init_controller = self.patcher_tft_init_controller.start()
        self.mock_tft_start_controller = self.patcher_tft_start_controller.start()
        self.mock_tft_register_mindx_callback = self.patcher_tft_register_mindx_callback.start()
        self.mock_tft_destroy_controller = self.patcher_tft_destroy_controller.start()

        self.patcher_cython_api = patch.object(cython_api, 'lib')
        self.mock_cython_api_lib = self.patcher_cython_api.start()
        self.mock_cython_api_lib.SendMessageToBackend = MagicMock(return_value=0)

        self.patcher_tft_notify_controller_dump = patch('taskd.python.framework.manager.controller.tft_notify_controller_dump')
        self.patcher_tft_notify_controller_stop_train = patch('taskd.python.framework.manager.controller.tft_notify_controller_stop_train')
        self.patcher_tft_notify_controller_on_global_rank = patch('taskd.python.framework.manager.controller.tft_notify_controller_on_global_rank')
        self.patcher_tft_notify_controller_change_strategy = patch('taskd.python.framework.manager.controller.tft_notify_controller_change_strategy')

        self.mock_tft_notify_controller_dump = self.patcher_tft_notify_controller_dump.start()
        self.mock_tft_notify_controller_stop_train = self.patcher_tft_notify_controller_stop_train.start()
        self.mock_tft_notify_controller_on_global_rank = self.patcher_tft_notify_controller_on_global_rank.start()
        self.mock_tft_notify_controller_change_strategy = self.patcher_tft_notify_controller_change_strategy.start()

    def tearDown(self):
        os.environ.clear()
        os.environ.update(self.original_env)
        self.patcher_log_error.stop()
        self.patcher_log_info.stop()
        self.patcher_log_warning.stop()
        self.patcher_tft_init_controller.stop()
        self.patcher_tft_start_controller.stop()
        self.patcher_tft_register_mindx_callback.stop()
        self.patcher_tft_destroy_controller.stop()
        self.patcher_cython_api.stop()
        self.patcher_tft_notify_controller_dump.stop()
        self.patcher_tft_notify_controller_stop_train.stop()
        self.patcher_tft_notify_controller_on_global_rank.stop()
        self.patcher_tft_notify_controller_change_strategy.stop()

    def test_controller_message_init(self):
        msg = ControllerMessage(
            action=TEST_ACTION,
            code=200,
            msg=TEST_MESSAGE,
            timeout=0,
            strategy=TEST_STRATEGY,
            params=TEST_PARAM
        )

        self.assertEqual(msg.action, TEST_ACTION)
        self.assertEqual(msg.code, 200)
        self.assertEqual(msg.msg, TEST_MESSAGE)
        self.assertEqual(msg.strategy, TEST_STRATEGY)
        self.assertEqual(msg.params, TEST_PARAM)
        self.assertEqual(msg.actions, [])
        self.assertEqual(msg.strategy_list, [])
        self.assertEqual(msg.fault_ranks, {})

        msg = ControllerMessage(
            action=TEST_ACTION,
            code=200,
            msg=TEST_MESSAGE,
            strategy=TEST_STRATEGY,
            params=TEST_PARAM,
            timeout=0,
            actions=["action1", "action2"],
            strategy_list=["strategy1", "strategy2"],
            fault_ranks={1: 2, 3: 4}
        )

        self.assertEqual(msg.actions, ["action1", "action2"])
        self.assertEqual(msg.strategy_list, ["strategy1", "strategy2"])
        self.assertEqual(msg.fault_ranks, {1: 2, 3: 4})

    def test_callback_funcs_init(self):
        callback = CallBackFuncs()  
        self.assertIn(constants.REPORT_FAULT_RANKS_CALLBACK, callback.callback_func_dict)
        self.assertIn(constants.STOP_COMPLETE_CALLBACK, callback.callback_func_dict)
        self.assertIn(constants.REPORT_STRATEGIES_CALLBACK, callback.callback_func_dict)
        self.assertIn(constants.REPORT_RESULT_CALLBACK, callback.callback_func_dict)

        self.assertEqual(callback.callback_func_dict[constants.REPORT_FAULT_RANKS_CALLBACK], report_process_fault)
        self.assertEqual(callback.callback_func_dict[constants.STOP_COMPLETE_CALLBACK], report_stop_complete)
        self.assertEqual(callback.callback_func_dict[constants.REPORT_STRATEGIES_CALLBACK], report_recover_strategy)
        self.assertEqual(callback.callback_func_dict[constants.REPORT_RESULT_CALLBACK], report_recover_status)

    def test_register_callback_func_success(self):
        self.mock_tft_register_mindx_callback.return_value = 0
        register_callback_func()
        callback = CallBackFuncs()
        self.assertEqual(self.mock_tft_register_mindx_callback.call_count, len(callback.callback_func_dict))
        self.mock_log_error.assert_not_called()

    def test_register_callback_func_failure(self):
        self.mock_tft_register_mindx_callback.return_value = 1
        register_callback_func()
        self.mock_log_error.assert_called()

    def test_init_controller_success(self):
        os.environ[constants.WORLD_SIZE] = "4"
        os.environ[constants.PROCESS_RECOVER] = "on"
        os.environ[constants.HIGH_AVAILABILITY_STRATEGY] = "elastic-training"
        os.environ[constants.POD_IP] = "127.0.0.1"
        os.environ[constants.TTP_PORT] = "8899"

        self.mock_tft_register_mindx_callback.return_value = 0

        init_controller()

        self.mock_tft_init_controller.assert_called_once_with(
            constants.MINDX_START_CONTROLLER_RANK, 4, False, True, True
        )
        self.mock_tft_start_controller.assert_called_once_with("127.0.0.1", 8899, False, "")
        self.mock_log_info.assert_called()
        self.mock_log_error.assert_not_called()

    def test_init_controller_no_world_size(self):
        os.environ[constants.POD_IP] = "127.0.0.1"
        os.environ[constants.TTP_PORT] = "8899"

        with self.assertRaises(ValueError):
            init_controller()

        self.mock_log_error.assert_called_with("init mindio controller failed, world_size: None")

    def test_init_controller_no_pod_ip(self):
        os.environ[constants.WORLD_SIZE] = "4"
        os.environ[constants.TTP_PORT] = "8899"

        with self.assertRaises(ValueError):
            init_controller()

        self.mock_log_error.assert_called()

    def test_init_controller_exception(self):
        os.environ[constants.WORLD_SIZE] = "4"
        os.environ[constants.POD_IP] = "127.0.0.1"
        os.environ[constants.TTP_PORT] = "8899"
        self.mock_tft_init_controller.side_effect = Exception("init failed")
        init_controller()
        self.mock_log_error.assert_called_with("init mindio/start mindio controller failed, Exception: init failed")
    
    def test_backend_send_callback_success(self):
        test_data = {
            "actions": ["save_and_exit"],
            "action": "test_action",
            "code": 200,
            "msg": "test message",
            "strategy": "test_strategy",
            "strategy_list": ["strategy1"],
            "fault_ranks": {"1": "2"},
            "params": "test_params"
        }
        data_str = json.dumps(test_data).encode('utf-8')

        data_ptr = ctypes.cast(ctypes.create_string_buffer(data_str), ctypes.c_void_p)

        with patch('taskd.python.framework.manager.controller.send_msg_to_controller') as mock_send_msg:
            result = backend_send_callback(data_ptr)

        self.assertEqual(result, 0)
        mock_send_msg.assert_called_once_with("save_and_exit", unittest.mock.ANY)
        self.mock_log_info.assert_called()
    
    def test_backend_send_callback_invalid_json(self):
        invalid_data = "{invalid json}".encode('utf-8')
        data_ptr = ctypes.cast(ctypes.create_string_buffer(invalid_data), ctypes.c_void_p)

        result = backend_send_callback(data_ptr)

        self.assertEqual(result, 1)
        self.mock_log_error.assert_called_with(unittest.mock.ANY)
    
    def test_backend_send_callback_exception(self):
        with patch('ctypes.cast', side_effect=Exception("cast failed")):
            result = backend_send_callback(None)

        self.assertEqual(result, 1)
        self.mock_log_error.assert_called_with("backend_callback parse message failed, reason: cast failed")

    def test_send_msg_to_controller_restart(self):
        with patch('taskd.python.framework.manager.controller.restart_controller') as mock_restart:
            send_msg_to_controller(constants.RESTARTCONTROLLER, None)

        mock_restart.assert_called_once()

    def test_send_msg_to_controller_destroy(self):
        send_msg_to_controller(constants.DESTRYCONTROLLER, None)

        self.mock_tft_destroy_controller.assert_called_once()
        self.mock_log_info.assert_called_with("destroy mindio controller")

    def test_send_msg_to_controller_unknown_action(self):
        test_data = ControllerMessage(
            action="test_action",
            code=200,
            timeout=0,
            msg="test message",
            strategy="test_strategy",
            params="test_params"
        )

        send_msg_to_controller("unknown_action", test_data)

        self.mock_tft_notify_controller_dump.assert_not_called()
        self.mock_tft_notify_controller_stop_train.assert_not_called()
        self.mock_tft_notify_controller_on_global_rank.assert_not_called()
        self.mock_tft_notify_controller_change_strategy.assert_not_called()

        self.mock_log_info.assert_called_with("do action unknown_action err, err=action unknown_action unregistered, "
                                              "data=ControllerMessage(action='test_action', code=200, msg='test message', "
                                              "strategy='test_strategy', params='test_params', timeout=0, actions=[], strategy_list=[], fault_ranks={})")

    def test_restart_controller(self):
        with patch('taskd.python.framework.manager.controller.init_controller') as mock_init_controller, \
             patch('time.sleep') as mock_sleep:
            restart_controller()
        self.mock_tft_destroy_controller.assert_called_once()
        mock_sleep.assert_called_once_with(1)
        mock_init_controller.assert_called_once()
        self.mock_log_info.assert_any_call("restart controller")
        self.mock_log_info.assert_any_call("restart controller finish")

    def test_report_stop_complete(self):
        with patch('taskd.python.framework.manager.controller.controller_send_to_backend') as mock_send:
            report_stop_complete(200, "test message", {1: 2})

        mock_send.assert_called_once()
        self.assertIsInstance(mock_send.call_args[0][0], ControllerMessage)
        self.assertEqual(mock_send.call_args[0][0].action, "stop_complete")
        self.assertEqual(mock_send.call_args[0][0].code, 200)
        self.assertEqual(mock_send.call_args[0][0].msg, "test message")
        self.assertEqual(mock_send.call_args[0][0].fault_ranks, {1: 2})

        self.mock_log_info.assert_called_with("call ReportStopComplete, msg:test message, fault_ranks={1: 2}")

    def test_report_recover_strategy(self):
        with patch('taskd.python.framework.manager.controller.controller_send_to_backend') as mock_send:
            report_recover_strategy({1: 2}, ["strategy1", "strategy2"])

        mock_send.assert_called_once()
        self.assertIsInstance(mock_send.call_args[0][0], ControllerMessage)
        self.assertEqual(mock_send.call_args[0][0].action, "recover_strategy")
        self.assertEqual(mock_send.call_args[0][0].fault_ranks, {1: 2})
        self.assertEqual(mock_send.call_args[0][0].strategy_list, ["strategy1", "strategy2"])

        self.mock_log_info.assert_called_with("call ReportRecoverStrategy, fault_ranks:{1: 2}, strategy_list:['strategy1', 'strategy2']")

    def test_report_recover_status(self):
        with patch('taskd.python.framework.manager.controller.controller_send_to_backend') as mock_send:
            report_recover_status(200, "test message", {1: 2}, "test_strategy")

        mock_send.assert_called_once()
        self.assertIsInstance(mock_send.call_args[0][0], ControllerMessage)
        self.assertEqual(mock_send.call_args[0][0].action, "recover_status")
        self.assertEqual(mock_send.call_args[0][0].code, 200)
        self.assertEqual(mock_send.call_args[0][0].msg, "test message")
        self.assertEqual(mock_send.call_args[0][0].fault_ranks, {1: 2})
        self.assertEqual(mock_send.call_args[0][0].strategy, "test_strategy")

        self.mock_log_info.assert_called_with("call ReportRecoverStatus, strategy: test_strategy, msg: test message")

    def test_report_process_fault(self):
        with patch('taskd.python.framework.manager.controller.controller_send_to_backend') as mock_send:
            report_process_fault({1: 2})

        mock_send.assert_called_once()
        self.assertIsInstance(mock_send.call_args[0][0], ControllerMessage)
        self.assertEqual(mock_send.call_args[0][0].action, "process_fault")
        self.assertEqual(mock_send.call_args[0][0].fault_ranks, {1: 2})
        self.mock_log_info.assert_called_with("call ReportProcessFault, fault_ranks:{1: 2}")

    def test_controller_send_to_backend_success(self):
        test_message = ControllerMessage(
            action="test_action",
            code=200,
            timeout=0,
            msg="test message",
            strategy="test_strategy",
            params="test_params"
        )
        controller_send_to_backend(test_message)

        self.mock_cython_api_lib.SendMessageToBackend.assert_called_once()
        self.assertIsInstance(self.mock_cython_api_lib.SendMessageToBackend.call_args[0][0], bytes)

    def test_controller_send_to_backend_lib_none(self):
        self.mock_cython_api_lib = None

        test_message = ControllerMessage(
            action="test_action",
            code=200,
            timeout=0,
            msg="test message",
            strategy="test_strategy",
            params="test_params"
        )
        with patch('taskd.python.framework.manager.controller.cython_api.lib', None):
            controller_send_to_backend(test_message)

        self.mock_log_error.assert_called_with('controller_send_to_backend cython_api.lib is None')

    def test_controller_send_to_backend_failure(self):
        self.mock_cython_api_lib.SendMessageToBackend.return_value = 1

        test_message = ControllerMessage(
            action="test_action",
            code=200,
            timeout=0,
            msg="test message",
            strategy="test_strategy",
            params="test_params"
        )

        controller_send_to_backend(test_message)

        self.mock_cython_api_lib.SendMessageToBackend.assert_called_once()

        self.mock_log_error.assert_called_with('controller_send_to_backend send message failed, res: 1')

    def test_controller_send_to_backend_exception(self):
        self.mock_cython_api_lib.SendMessageToBackend.side_effect = Exception("send failed")

        test_message = ControllerMessage(
            action="test_action",
            code=200,
            timeout=0,
            msg="test message",
            strategy="test_strategy",
            params="test_params"
        )

        controller_send_to_backend(test_message)
        self.mock_cython_api_lib.SendMessageToBackend.assert_called_once()
        self.mock_log_error.assert_called_with('controller_send_to_backend send message failed, error: send failed')


if __name__ == '__main__':
    unittest.main()