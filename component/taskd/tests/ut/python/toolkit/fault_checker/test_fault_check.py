#!/usr/bin/python3
# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
import logging
import os.path
import signal

from unittest import TestCase, mock

from torch.distributed.elastic.agent.server import RunResult, WorkerState, WorkerSpec, WorkerGroup, Worker

from taskd.python.toolkit.fault_checker.fault_check import FaultProcessor, clean_before_restart, grace_exit_pids, \
    force_exit_pids, all_pid_stopped, stop_pids
from taskd.python.constants import constants
from taskd.python.toolkit.config.path import TORCH_EXTENSIONS_CACHE_DIR
from taskd.python.utils.log import run_log


def mock_get_env(key: str) -> str:
    return "test-host"


def do_nothing():
    pass


def gen_worker_group() -> WorkerGroup:
    spec = WorkerSpec(
        role="test_trainer",
        local_world_size=8,
        fn=do_nothing,
        args=(),
        rdzv_handler=None,
        max_restarts=5,
        monitor_interval=1,
    )
    worker_group = WorkerGroup(spec)
    worker_group.workers = [Worker(local_rank=i, global_rank=i) for i in range(spec.local_world_size)]
    return worker_group


def mock_run(self, role: str = "default") -> RunResult:
    if self._worker_group.spec.max_restarts == constants.MAX_INT16:
        return RunResult(WorkerState("SUCCEEDED"), {1: None}, {})
    return RunResult(WorkerState("FAILED"), {0: None}, {})


@mock.patch('os.getenv', mock_get_env)
class TestFaultProcessor(TestCase):
    # when reset.json has unrecovered rank then return false
    def test_unrecovered(self):
        fault_processor = FaultProcessor()
        fault_processor.reset_cm_path = os.path.join(os.path.dirname(__file__),
                                                     "../../reset_config_files/unrecovered/reset.json")
        fault_processor._update_reset_info()
        is_recovered = fault_processor.is_recovered()
        self.assertFalse(is_recovered)

    # when reset.json has no unrecovered rank then return true
    def test_recovered(self):
        fault_processor = FaultProcessor()
        fault_processor.reset_cm_path = os.path.join(os.path.dirname(__file__),
                                                     "../../reset_config_files/recovered/reset.json")
        # read a recovered config for the first time should return true
        fault_processor._update_reset_info()
        is_recovered = fault_processor.is_recovered()
        self.assertTrue(is_recovered)
        # only a recovered case will update the retry_time cache
        # retry_time:2, max_restarts:1, return:0
        self.assertEqual(fault_processor.get_remain_retry_time(1), 0)
        # read this same recovered config should return false

    # when reset.json is recovered and restart type is hotreset then return true
    def test_ranktable_unupdated_hotreset_recover(self):
        fault_processor = FaultProcessor()
        fault_processor.reset_cm_path = os.path.join(os.path.dirname(__file__),
                                                     "../../reset_config_files/recovered/reset.json")
        fault_processor.restart_type_path = os.path.join(os.path.dirname(__file__),
                                                     "../../reset_config_files/rank/restartTypeHotReset")
        fault_processor.rank_table_version = os.path.join(os.path.dirname(__file__),
                                                     "../../reset_config_files/rank/version1")
        # read version1 file_rank_version will be 1,self rank version is  1, 1=1, but type is hot-reset will be
        # recovered status
        fault_processor._update_reset_info()
        fault_processor.rank_table_version = 1
        is_recovered = fault_processor.is_recovered()
        self.assertTrue(is_recovered)

    # when reset.json is recovered and restart type is pod reschedule and version is bigger than current
    # then return true and update version
    def test_ranktable_updated_recover(self):
        fault_processor = FaultProcessor()
        fault_processor.reset_cm_path = os.path.join(os.path.dirname(__file__),
                                                     "../../reset_config_files/recovered/reset.json")
        fault_processor.restart_type_path = os.path.join(os.path.dirname(__file__),
                                                         "../../reset_config_files/rank/restartTypePodReschedule")
        fault_processor.rank_version_path = os.path.join(os.path.dirname(__file__),
                                                          "../../reset_config_files/rank/version1")
        fault_processor._update_reset_info()
        # read version1 file_rank_version will be 1,self rank version is  1, 1>0, will be recovered status
        fault_processor.rank_table_version = 0
        is_recovered = fault_processor.is_recovered()
        self.assertTrue(is_recovered)
        self.assertEqual(fault_processor.rank_table_version, 1)

    # when reset.json is recovered and restart type is pod reschedule and version is the same as current version
    # then return false
    def test_ranktable_unupdate_hotreset_unrecover(self):
        fault_processor = FaultProcessor()
        fault_processor.reset_cm_path = os.path.join(os.path.dirname(__file__),
                                                     "../../reset_config_files/recovered/reset.json")
        fault_processor.restart_type_path = os.path.join(os.path.dirname(__file__),
                                                         "../../reset_config_files/rank/restartTypePodReschedule")
        fault_processor.rank_version_path = os.path.join(os.path.dirname(__file__),
                                                         "../../reset_config_files/rank/version1")
        fault_processor._update_reset_info()
        # read version1 file_rank_version will be 1,self rank version is  1, 1=1, but type is podReschedule, rank table
        # is not updated yet
        # is_recovered should be false
        fault_processor.rank_table_version = 1
        is_recovered = fault_processor.is_recovered()
        self.assertFalse(is_recovered)

    # when version file is valid then return version
    def test_get_rank_version(self):
        fault_processor = FaultProcessor()
        fault_processor.rank_version_path = os.path.join(os.path.dirname(__file__),
                                                         "../../reset_config_files/rank/version1")
        version = fault_processor.read_rank_table_version()
        self.assertEqual(version, 1)

    # when version file is invalid then return -1
    def test_get_rank_version_fault(self):
        fault_processor = FaultProcessor()
        fault_processor.rank_version_path = os.path.join(os.path.dirname(__file__),
                                                         "../../reset_config_files/rank/version_fault")
        version = fault_processor.read_rank_table_version()
        self.assertEqual(version, -1)

    # when reset.json contain fault status and unrecovered status then fault_status is fault and unrecovered
    def test_fault(self):
        worker_group = gen_worker_group()
        fault_processor = FaultProcessor()
        fault_processor.reset_cm_path = os.path.join(os.path.dirname(__file__),
                                                     "../../reset_config_files/fault/reset.json")
        fault_status = fault_processor.get_fault_status(worker_group)
        self.assertTrue(fault_status.is_fault)
        self.assertTrue(fault_status.is_unrecovered)

    # when reset.json has fault then return false
    def test_wait_start_fail(self):
        worker_group = gen_worker_group()
        fault_processor = FaultProcessor()
        fault_processor.reset_cm_path = os.path.join(os.path.dirname(__file__),
                                                     "../../reset_config_files/fault/reset.json")
        is_to_start = fault_processor.wait_to_start(worker_group)
        self.assertFalse(is_to_start)

    def test_clean_before_restart_no_dir(self):
        with mock.patch('os.path.exists', return_value=False):
            clean_before_restart()
            assert not os.path.exists(TORCH_EXTENSIONS_CACHE_DIR)

    def test_clean_before_restart_file(self):
        with mock.patch('os.path.exists', return_value=True):
            with mock.patch('os.path.isfile', return_value=True):
                with mock.patch('os.remove') as mock_remove:
                    clean_before_restart()
                    mock_remove.assert_called_once_with(TORCH_EXTENSIONS_CACHE_DIR)

    @mock.patch('os.path.exists')
    def test_clean_before_restart_link(self, mock_exists):
        mock_exists.return_value = True
        with mock.patch('os.path.isfile', return_value=False):
            with mock.patch('os.path.islink', return_value=True):
                with mock.patch('os.remove') as mock_remove:
                    clean_before_restart()
                    mock_remove.assert_called_once_with(TORCH_EXTENSIONS_CACHE_DIR)

    @mock.patch('os.path.exists')
    @mock.patch('os.path.isfile')
    def test_clean_before_restart_dir(self, mock_isfile, mock_exists):
        mock_exists.return_value = True
        mock_isfile.return_value = False
        with mock.patch('os.path.islink', return_value=False):
            with mock.patch('os.path.isdir', return_value=True):
                with mock.patch('shutil.rmtree') as mock_rmtree:
                    clean_before_restart()
                    mock_rmtree.assert_called_once_with(TORCH_EXTENSIONS_CACHE_DIR, ignore_errors=True)

    def test_grace_exit_pids_invalid_input(self):
        with self.assertRaises(ValueError):
            grace_exit_pids("invalid_input")

    @mock.patch('os.path.exists')
    @mock.patch('os.kill')
    def test_grace_exit_pids_valid_input(self, mock_kill, mock_exists):
        mock_exists.return_value = True
        grace_exit_pids({"process1": 1234})
        mock_kill.assert_called_once_with(1234, signal.SIGTERM)

    @mock.patch('os.path.exists')
    @mock.patch('os.kill')
    def test_grace_exit_pids_process_not_exist(self, mock_kill, mock_exists):
        mock_exists.return_value = False
        grace_exit_pids({"process1": 1234})
        mock_kill.assert_not_called()

    @mock.patch('os.path.exists')
    @mock.patch('os.kill')
    def test_grace_exit_pids_process_lookup_error(self, mock_kill, mock_exists):
        mock_exists.return_value = True
        mock_kill.side_effect = ProcessLookupError
        with mock.patch.object(run_log, 'warning') as mock_warning:
            grace_exit_pids({"process1": 1234})
            mock_warning.assert_called_once()

    @mock.patch('os.path.exists')
    @mock.patch('os.kill')
    def test_grace_exit_pids_unexpected_error(self, mock_kill, mock_exists):
        mock_exists.return_value = True
        mock_kill.side_effect = Exception('Unexpected error')
        pids = {'process1': 1234}
        with mock.patch('taskd.python.utils.log.run_log.error') as mock_run_log_error:
            with self.assertRaises(Exception):
                grace_exit_pids(pids)
            mock_run_log_error.assert_called_once_with(
                'An unexpected error Unexpected error occur when kill the process of 1234')

    @mock.patch('os.path.exists')
    @mock.patch('os.kill')
    def test_force_exit_pids_with_valid_pid(self, mock_kill, mock_exists):
        mock_exists.return_value = True
        mock_kill.return_value = None
        pids = {'pid1': 1234}
        force_exit_pids(pids)
        mock_kill.assert_called_once_with(1234, signal.SIGKILL)

    @mock.patch('os.path.exists')
    @mock.patch('os.kill')
    def test_force_exit_pids_process_not_exist(self, mock_kill, mock_exists):
        mock_exists.return_value = False
        force_exit_pids([1234])
        mock_kill.assert_not_called()

    @mock.patch('os.path.exists')
    @mock.patch('os.kill')
    def test_force_exit_pids_with_lookup_error(self, mock_kill, mock_exists):
        mock_exists.return_value = True
        mock_kill.side_effect = ProcessLookupError
        pids = [1234]
        with mock.patch.object(run_log, 'warning') as mock_warning:
            force_exit_pids({"process1": 1234})
            mock_warning.assert_called_once()

    @mock.patch('os.path.exists')
    @mock.patch('os.kill')
    def test_force_exit_pids_with_unexpected_error(self, mock_kill, mock_exists):
        mock_exists.return_value = True
        mock_kill.side_effect = Exception("Unexpected error")
        pids = [1234]
        with mock.patch('taskd.python.utils.log.run_log.error') as mock_run_log_error:
            with self.assertRaises(Exception):
                force_exit_pids(pids)
            mock_run_log_error.assert_called_once_with(
                'An unexpected error Unexpected error occur when kill the process of 1234')

    @mock.patch('os.path.exists')
    def test_all_pid_stopped_with_list(self, mock_exists):
        mock_exists.return_value = False
        self.assertTrue(all_pid_stopped([1, 2, 3]))

    @mock.patch('os.path.exists')
    def test_all_pid_stopped_with_dict(self, mock_exists):
        mock_exists.return_value = False
        self.assertTrue(all_pid_stopped({'p1': 1, 'p2': 2, 'p3': 3}))

    @mock.patch('os.path.exists')
    def test_all_pid_stopped_with_some_running(self, mock_exists):
        mock_exists.side_effect = [False, True, False]
        self.assertFalse(all_pid_stopped([1, 2, 3]))

    @mock.patch('os.path.exists')
    def test_all_pid_stopped_with_all_running(self, mock_exists):
        mock_exists.return_value = True
        self.assertFalse(all_pid_stopped([1, 2, 3]))

    @mock.patch('taskd.python.constants.constants')
    @mock.patch('time.time')
    @mock.patch('taskd.python.toolkit.fault_checker.fault_check.all_pid_stopped')
    @mock.patch('taskd.python.toolkit.fault_checker.fault_check.force_exit_pids')
    def test_stop_pids_not_timeout(self, mock_force_exit_pids, mock_all_pid_stopped, mock_time, mock_constants):
        mock_constants.GRACE_TIME_OUT = 10
        mock_constants.SLEEP_GAP = 1
        mock_time.side_effect = [0, 5, 10, 15]
        mock_all_pid_stopped.side_effect = [False, True, True]
        stop_pids([1, 2, 3])
        mock_all_pid_stopped.assert_called_with([1, 2, 3])
        mock_force_exit_pids.assert_not_called()