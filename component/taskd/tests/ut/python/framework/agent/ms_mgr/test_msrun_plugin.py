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
from unittest.mock import MagicMock, patch

from taskd.python.toolkit.constants import constants
from taskd.python.framework.agent.ms_mgr.msrun_plugin import MSRunPlugin
from taskd.python.toolkit.constants.constants import START_ALL_WORKER_CALLBACK_NAME, START_WORKER_LIST_CALLBACK_NAME, \
    KILL_ALL_WORKER_CALLBACK_NAME, MONITOR_CALLBACK_NAME
from taskd.python.toolkit.fault_checker.fault_check import ResetCmData, FaultStatus


def getDemoMsPlugin()->MSRunPlugin:
    ms = MSRunPlugin()
    ms.start_proxy = MagicMock()
    mock_kill = MagicMock()
    mock_start_all = MagicMock()
    mock_start = MagicMock()
    mock_monitor = MagicMock()
    ms.register_callbacks(KILL_ALL_WORKER_CALLBACK_NAME, mock_kill)
    ms.register_callbacks(START_ALL_WORKER_CALLBACK_NAME, mock_start_all)
    ms.register_callbacks(START_WORKER_LIST_CALLBACK_NAME, mock_start)
    ms.register_callbacks(MONITOR_CALLBACK_NAME, mock_monitor)
    return ms


class TestMSRunPlugin(unittest.TestCase):
    @patch('time.sleep')
    def test_start_mindspore_workers_success(self, mock_sleep):
        ms = MSRunPlugin()
        mock_func = MagicMock()
        ms.register_callbacks(START_ALL_WORKER_CALLBACK_NAME, mock_func)
        ms.start_mindspore_workers()
        ms.wait_to_start = lambda: True
        mock_func.assert_called_once()

    @patch('time.sleep')
    def test_start_mindspore_workers_fail(self, mock_sleep):
        ms = MSRunPlugin()
        mock_func = MagicMock()
        ms.register_callbacks(START_ALL_WORKER_CALLBACK_NAME, mock_func)
        ms.wait_to_start = lambda: False
        with self.assertRaises(ValueError) as cm:
            ms.start_mindspore_workers()
        self.assertIsNotNone(cm.exception)

    @patch('time.sleep')
    @patch('os.getenv')
    def test_start_mindspore_worker_list_success(self, mock_env, mock_sleep):
        mock_env.return_value = 8
        ms = MSRunPlugin()
        mock_func = MagicMock(side_effect=lambda x: None)
        ms.register_callbacks(START_WORKER_LIST_CALLBACK_NAME, mock_func)
        ms.wait_to_start = lambda: True
        ms.start_mindspore_worker_list([8,9,10,11])
        mock_func.assert_called_once()

    @patch('time.sleep')
    @patch('os.getenv')
    def test_start_mindspore_worker_list_fail(self, mock_env, mock_sleep):
        mock_env.return_value = 8
        ms = MSRunPlugin()
        mock_func = MagicMock(side_effect=lambda x: None)
        ms.register_callbacks(START_WORKER_LIST_CALLBACK_NAME, mock_func)
        ms.wait_to_start = lambda: False
        with self.assertRaises(ValueError) as cm:
            ms.start_mindspore_worker_list([8, 9, 10, 11])
        mock_func.assert_not_called()
        self.assertIsNotNone(cm.exception)

    def test_all_fault_has_recovered_failure_on_not_all_recovered(self):
        ms = MSRunPlugin()
        ms.fault_ranks = [{'Status': 'recovered'},{'Status': 'unrecovered'},{'hello':'world'}]
        res = ms.all_fault_has_recovered()
        self.assertFalse(res)

    @patch('os.path.exists')
    def test_all_fault_has_recovered_failure_on_wrong_version(self, mock_exists):
        mock_exists.return_value = True
        ms = MSRunPlugin()
        ms.fault_ranks = [{'Status': 'recovered'},{'Status': 'recovered'}]
        ms.restart_type = constants.VALUE_RESTART_RESCHEDULE_TYPE
        ms.rank_table_version = 1
        ms.read_rank_table_version = lambda: 0
        res = ms.all_fault_has_recovered()
        self.assertFalse(res)

    @patch('os.path.exists')
    def test_all_fault_has_recovered_success(self, mock_exists):
        mock_exists.return_value = True
        ms = MSRunPlugin()
        ms.fault_ranks = [{'Status': 'recovered'}, {'Status': 'recovered'}]
        ms.restart_type = constants.VALUE_RESTART_RESCHEDULE_TYPE
        ms.rank_table_version = 1
        ms.read_rank_table_version = lambda: 2
        res = ms.all_fault_has_recovered()
        self.assertTrue(res)

    @patch('taskd.python.toolkit.fault_checker.fault_check.fault_processor.get_reset_info_from_cm')
    def test_get_fault_status(self, mock_cm: MagicMock):
        mock_cm.return_value = ResetCmData(
            fault_ranks=[
                {'Status': 'recovered', 'RankId': 0},
                {'Status': 'unrecovered', 'RankId': 1},
                {'Status': 'fault', 'RankId': 2},
                {'hello':'world'}
            ],
            retry_time=2,
            grace_exit=True,
            restart_type=constants.VALUE_RESTART_HOTRESET_TYPE,
            fault_flush=False,
            restart_fault_process=False)
        ms = MSRunPlugin()
        ms.node_global_rank_ids = [0,1,2,3]
        ms.pre_retry_time = 1
        res = ms.get_fault_status()
        self.assertTrue(res.is_fault)
        self.assertTrue(res.is_unrecovered)
        self.assertTrue(res.is_retried)
        self.assertFalse(res.restart_fault_process)
        self.assertEqual(res.local_ranks, [2])

    def test_read_rank_table_version_fail(self):
        ms = MSRunPlugin()
        res = ms.read_rank_table_version()
        self.assertEqual(res, -1)

    @patch('taskd.python.framework.agent.ms_mgr.msrun_plugin.safe_get_file_info')
    def test_read_rank_table_version_success(self, mock_info: MagicMock):
        mock_info.return_value = "111"
        ms = MSRunPlugin()
        res = ms.read_rank_table_version()
        self.assertEqual(res, 111)

    def test_start_fail_when_some_func_not_register(self):
        ms = getDemoMsPlugin()
        ms.register_callbacks(START_WORKER_LIST_CALLBACK_NAME, None)
        with self.assertRaises(Exception) as exp:
            ms.start()
        self.assertIsNotNone(exp)

    @patch('taskd.python.toolkit.recover_module.shared_data.shared_data_inst.get_kill_flag')
    @patch('builtins.exit')
    def test_start_kill_master(self, mock_exit: MagicMock, mock_flag: MagicMock):
        mock_flag.return_value = True
        mock_exit.side_effect = mock_exit_raise
        ms = getDemoMsPlugin()
        ms.start_mindspore_workers = lambda: None
        ms._init_grpc_client_if_needed = lambda: None
        ms.ms_node_rank = "0"
        with self.assertRaises(Exception) as exp:
            ms.start()
        mock_exit.assert_called()
        mock_flag.assert_called_once()
        self.assertIsNotNone(exp)

    @patch("time.sleep")
    @patch("taskd.python.framework.agent.ms_mgr.msrun_plugin.check_monitor_res_valid")
    def test_start(self, mock_res_valid: MagicMock, mock_sleep: MagicMock):
        ms = getDemoMsPlugin()
        mock_res_valid.side_effect = [False, True]
        ms.ms_node_rank = "1"
        mock_monitor = MagicMock()
        ms.register_callbacks(MONITOR_CALLBACK_NAME, mock_monitor)
        ms.update_rank_status = lambda _: None
        ms.update_reset_info = lambda: None
        ms._handle_grace_exit = lambda: False
        ms._handle_fault_status = lambda _: False
        ms._handle_process_retry_fault = lambda _: False
        ms._handle_hardware_fault = lambda _: False
        ms._handle_all_process_succeed = MagicMock()
        ms._handle_exist_unhealthy_process = mock_exit_raise
        with self.assertRaises(Exception) as exp:
            ms.start()
        ms._handle_all_process_succeed.assert_called_once()
        self.assertIsNotNone(exp)

    def test_update_rank_status_unhealthy(self):
        ms = getDemoMsPlugin()
        rank_status_dict = {
            0: {'pid': 101, 'status': 'UNHEALTHY', 'global_rank': 16},
            1: {'pid': 110, 'status': None, 'global_rank': 17}
        }
        ms.update_rank_status(rank_status_dict)
        self.assertEqual(ms.node_global_rank_ids, [16, 17])
        self.assertEqual(ms.all_rank_succeed, False)
        self.assertEqual(ms.local_rank_to_pid, {16:101,17:110})
        self.assertEqual(ms.rank_status, ms.RANK_STATUS_UNHEALTHY)

    def test_update_rank_status_succeed(self):
        ms = getDemoMsPlugin()
        rank_status_dict = {
            0: {'pid': 101, 'status': 0, 'global_rank': 16},
            1: {'pid': 110, 'status': 0, 'global_rank': 17}
        }
        ms.update_rank_status(rank_status_dict)
        self.assertEqual(ms.node_global_rank_ids, [16, 17])
        self.assertEqual(ms.all_rank_succeed, True)
        self.assertEqual(ms.local_rank_to_pid, {16:101,17:110})
        self.assertEqual(ms.rank_status, ms.RANK_STATUS_HEALTHY)

    def test_update_rank_status_healthy(self):
        ms = getDemoMsPlugin()
        rank_status_dict = {
            0: {'pid': 101, 'status': None, 'global_rank': 16},
            1: {'pid': 110, 'status': None, 'global_rank': 17}
        }
        ms.update_rank_status(rank_status_dict)
        self.assertEqual(ms.node_global_rank_ids, [16, 17])
        self.assertEqual(ms.all_rank_succeed, False)
        self.assertEqual(ms.local_rank_to_pid, {16:101,17:110})
        self.assertEqual(ms.rank_status, ms.RANK_STATUS_HEALTHY)

    @patch('taskd.python.toolkit.fault_checker.fault_check.fault_processor.get_reset_info_from_cm')
    def test_update_reset_info(self, mock_cm: MagicMock):
        data = ResetCmData(fault_ranks=[{'Status': 'recovered', 'RankId': 0}, {'Status': 'unrecovered', 'RankId': 1},
                                        {'Status': 'fault', 'RankId': 2}, {'hello': 'world'}], retry_time=2,
                           grace_exit=True, restart_type=constants.VALUE_RESTART_HOTRESET_TYPE, fault_flush=False,
                           restart_fault_process=False)
        mock_cm.return_value = data
        ms = MSRunPlugin()
        ms.update_reset_info()
        self.assertTrue(ms.fault_ranks, data.fault_ranks)
        self.assertTrue(ms.retry_time, data.retry_time)
        self.assertTrue(ms.grace_exit, data.grace_exit)
        self.assertEqual(ms.restart_type, data.restart_type)
        self.assertEqual(ms.restart_fault_process, data.restart_fault_process)

    def test_update_pre_fault_infos(self):
        ms = MSRunPlugin()
        ms.retry_time = 1
        ms.update_pre_fault_infos()
        self.assertTrue(ms.pre_retry_time, ms.retry_time)

def mock_exit_raise():
    raise Exception("exit")

def mock_fault_status():
    return FaultStatus([], False, False, False, False)

if __name__ == '__main__':
    unittest.main()
