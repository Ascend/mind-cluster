#!/usr/bin/python3
# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

import os
import sys
from concurrent import futures
from unittest import TestCase, mock
import grpc
import taskd.python.toolkit.validator.cert_check
from taskd.python.constants import constants
from taskd.python.utils.log.logger import run_log

from taskd.python.toolkit.recover_module.pb import recover_pb2_grpc
from taskd.python.toolkit.recover_module.pb import recover_pb2 as pb


class TestRecoverServicer(recover_pb2_grpc.RecoverServicer):
    def Init(self, request, context):
        status = pb.Status()
        status.code = 0
        status.info = "ok"
        return status

    def Register(self, request, context):
        status = pb.Status()
        status.code = 0
        status.info = "ok"
        return status

    def SubscribeProcessManageSignal(self, request, context):
        status = pb.Status()
        status.code = 0
        status.info = "ok"
        return status

    def ReportStopComplete(self, request, context):
        status = pb.Status()
        status.code = 0
        status.info = "ok"
        return status

    def ReportRecoverStrategy(self, request, context):
        status = pb.Status()
        status.code = 0
        status.info = "ok"
        return status

    def ReportRecoverStatus(self, request, context):
        status = pb.Status()
        status.code = 0
        status.info = "ok"
        return status

    def ReportProcessFault(self, request, context):
        status = pb.Status()
        status.code = 0
        status.info = "ok"
        return status


def start_grpc_server():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    recover_pb2_grpc.add_RecoverServicer_to_server(TestRecoverServicer(), server)
    server.add_insecure_port("[::]:8899")
    server.start()
    return server


def set_env():
    os.environ["MINDX_TASK_ID"] = "123456789"
    os.environ["MINDX_SERVER_IP"] = "localhost"
    os.environ["TTP_PORT"] = "8000"
    os.environ["WORLD_SIZE"] = "16"
    os.environ['POD_IP'] = '1.2.3.4'
    os.environ['LOCAL_PROXY_ENABLE'] = "on"


def del_env():
    del os.environ["MINDX_TASK_ID"]
    del os.environ["MINDX_SERVER_IP"]
    del os.environ["TTP_PORT"]
    del os.environ["WORLD_SIZE"]
    if os.getenv('POD_IP') is not None:
        del os.environ['POD_IP']


class TestRecoverManager(TestCase):
    def setUp(self) -> None:
        set_env()
        self.server = start_grpc_server()

    def tearDown(self) -> None:
        from taskd.python.toolkit.recover_module import shared_data
        shared_data.shared_data_inst.set_kill_flag(False)
        shared_data.shared_data_inst.set_exit_flag(False)
        del_env()
        self.server.stop(0)

    def set_action_map(self, obj):
        obj.action_func_map = {
            'save_and_exit': mock.MagicMock(),
            'stop_train': mock.MagicMock(),
            'pause_train': mock.MagicMock(),
            'on_global_rank': mock.MagicMock(),
            'change_strategy': mock.MagicMock(),
        }

    def test_init_grpc_recover_manager(self):
        from taskd.python.toolkit.recover_module import recover_manager, DLRecoverManager
        manager = recover_manager.init_grpc_recover_manager()
        self.assertIsInstance(manager, DLRecoverManager)

    def test_init_mindio_controller_no_pod_ip(self):
        from taskd.python.toolkit.recover_module import recover_manager
        with self.assertRaises(ValueError):
            del os.environ['POD_IP']
            recover_manager.init_mindio_controller()

    def test_report_stop_complete(self):
        from taskd.python.toolkit.recover_module import recover_manager
        fault_ranks = {1: 0, 2: 0, 3: 0}
        ret = recover_manager.report_stop_complete(0, 'stop', fault_ranks)
        self.assertEqual(ret, 0)

    def test_report_recover_strategy(self):
        from taskd.python.toolkit.recover_module import recover_manager
        fault_ranks = {1: 0, 2: 0, 3: 0}
        strategies = ['recover', 'dump', 'exit']
        ret = recover_manager.report_recover_strategy(fault_ranks, strategies)
        self.assertEqual(ret, 0)

    def test_report_recover_status(self):
        from taskd.python.toolkit.recover_module import recover_manager
        fault_ranks = {1: 0, 2: 0, 3: 0}
        strategy = 'recover'
        ret = recover_manager.report_recover_status(0, 'recover success', fault_ranks, strategy)
        self.assertEqual(ret, 0)

    def test_report_process_fault(self):
        from taskd.python.toolkit.recover_module import recover_manager
        fault_ranks = {1: 0, 2: 0, 3: 0}
        ret = recover_manager.report_process_fault(fault_ranks)
        self.assertEqual(ret, 0)

    def test_init_init_grpc_process(self):
        from taskd.python.toolkit.recover_module import recover_manager
        with mock.patch('taskd.python.toolkit.recover_module.recover_manager.DLRecoverManager.start_subscribe')\
                as mock_subscribe:
            mock_subscribe.return_value = 0
            recover_manager.init_grpc_process()
            mock_subscribe.assert_called_once()

    def test_start_subscribe_save_exit_action(self):
        from taskd.python.toolkit.recover_module import recover_manager
        manager = recover_manager.init_grpc_recover_manager()
        self.set_action_map(manager)
        fault_ranks = {1: 0, 2: 0, 3: 0}
        pb_data = recover_manager.ProtoBufData(fault_ranks, 'dump', 0)
        manager._DLRecoverManager__do_action('save_and_exit', pb_data)
        manager.action_func_map['save_and_exit'].assert_called_once()

    def test_start_subscribe_stop_train_action(self):
        from taskd.python.toolkit.recover_module import recover_manager
        manager = recover_manager.init_grpc_recover_manager()
        self.set_action_map(manager)
        fault_ranks = {1: 0, 2: 0, 3: 0}
        pb_data = recover_manager.ProtoBufData(fault_ranks, 'stop_train', 0)
        manager._DLRecoverManager__do_action('stop_train', pb_data)
        manager.action_func_map['stop_train'].assert_called_once()

    def test_start_subscribe_pause_train_action(self):
        from taskd.python.toolkit.recover_module import recover_manager
        manager = recover_manager.init_grpc_recover_manager()
        self.set_action_map(manager)
        fault_ranks = {1: 0, 2: 0, 3: 0}
        pb_data = recover_manager.ProtoBufData(fault_ranks, 'pause_train', 0)
        manager._DLRecoverManager__do_action('pause_train', pb_data)
        manager.action_func_map['pause_train'].assert_called_once()

    def test_start_subscribe_on_global_rank_action(self):
        from taskd.python.toolkit.recover_module import recover_manager
        manager = recover_manager.init_grpc_recover_manager()
        self.set_action_map(manager)
        fault_ranks = {1: 0, 2: 0, 3: 0}
        pb_data = recover_manager.ProtoBufData(fault_ranks, 'on_global_rank', 1)
        manager._DLRecoverManager__do_action('on_global_rank', pb_data)
        manager.action_func_map['on_global_rank'].assert_called_once()

    def test_start_subscribe_change_strategy_action(self):
        from taskd.python.toolkit.recover_module import recover_manager
        manager = recover_manager.init_grpc_recover_manager()
        self.set_action_map(manager)
        fault_ranks = {1: 0, 2: 0, 3: 0}
        pb_data = recover_manager.ProtoBufData(fault_ranks, 'recover', 0)
        manager._DLRecoverManager__do_action('change_strategy', pb_data)
        manager.action_func_map['change_strategy'].assert_called_once()

    def test_signal_pipe_line_kill_master(self):
        from taskd.python.toolkit.recover_module import recover_manager
        from taskd.python.toolkit.recover_module import shared_data
        manager = recover_manager.init_grpc_recover_manager()
        signal = pb.ProcessManageSignal()
        signal.uuid = '123456'
        signal.jobId = os.environ['MINDX_TASK_ID']
        signal.signalType = 'killMaster'
        signal.actions.append('')
        signal.changeStrategy = ''
        manager._DLRecoverManager__signal_pipe_line(signal)
        self.assertTrue(shared_data.shared_data_inst.get_kill_flag())

    def test_signal_pipe_line_keep_alive(self):
        from taskd.python.toolkit.recover_module import recover_manager
        with mock.patch('taskd.python.utils.log.logger.run_log.debug') as mock_run_log_debug:
            manager = recover_manager.init_grpc_recover_manager()
            signal = pb.ProcessManageSignal()
            signal.uuid = '123456'
            signal.jobId = os.environ['MINDX_TASK_ID']
            signal.signalType = 'keep-alive'
            signal.actions.append('')
            signal.changeStrategy = ''
            manager._DLRecoverManager__signal_pipe_line(signal)
            mock_run_log_debug.assert_any_call('keep-alive signal now not handle')

    def test_signal_pipe_line_invalid_jobid(self):
        from taskd.python.toolkit.recover_module import recover_manager
        with mock.patch('taskd.python.utils.log.logger.run_log.info') as mock_run_log_info:
            manager = recover_manager.init_grpc_recover_manager()
            signal = pb.ProcessManageSignal()
            signal.uuid = '123456'
            signal.jobId = '654321'
            signal.signalType = 'test_signal'
            signal.actions.append('')
            signal.changeStrategy = ''
            manager._DLRecoverManager__signal_pipe_line(signal)
            mock_run_log_info.assert_called_with(
                f"discard signal cause client_jobId={manager.client_info.jobId}, but signal_jobId={signal.jobId}")

    def test_signal_pipe_line_do_action(self):
        from taskd.python.toolkit.recover_module import recover_manager
        with mock.patch.object(recover_manager.DLRecoverManager, '_DLRecoverManager__do_action') as mock_do_action:
            manager = recover_manager.init_grpc_recover_manager()
            signal = pb.ProcessManageSignal()
            signal.uuid = '123456'
            signal.jobId = os.environ['MINDX_TASK_ID']
            signal.signalType = 'save_and_exit'
            signal.actions.append('save_and_exit')
            signal.changeStrategy = ''
            manager._DLRecoverManager__signal_pipe_line(signal)
            mock_do_action.assert_called_once()

    @mock.patch('taskd.python.utils.log.logger.run_log.info')
    @mock.patch('time.sleep')
    def test_init_clusterd_success(self, mock_sleep, mock_run_log):
        from taskd.python.toolkit.recover_module import recover_manager
        manager = recover_manager.init_grpc_recover_manager()
        status = pb.Status()
        status.code = 0
        status.info = "ok"
        manager.grpc_stub.Init = mock.Mock(return_value=status)
        ret = manager.init_clusterd()
        self.assertTrue(ret)
        mock_run_log.assert_called_with("init process recover succeed")
        mock_sleep.assert_not_called()

    @mock.patch('taskd.python.utils.log.logger.run_log.info')
    @mock.patch('time.sleep')
    def test_init_clusterd_process_recover_enable_off(self, mock_sleep, mock_run_log):
        from taskd.python.toolkit.recover_module import recover_manager
        manager = recover_manager.init_grpc_recover_manager()
        status = pb.Status()
        status.code = 403
        status.info = "process recover enable off"
        manager.grpc_stub.Init = mock.Mock(return_value=status)
        ret = manager.init_clusterd()
        self.assertFalse(ret)
        mock_run_log.assert_called_with("ProcessRecoverEnable is off, will no longer init clusterd")
        mock_sleep.assert_not_called()

    def test_init_clusterd_use_proxy(self):
        from taskd.python.toolkit.recover_module import recover_manager
        manager = recover_manager.init_grpc_recover_manager()
        self.assertEqual(manager.server_addr, '127.0.0.1:8899')

    @mock.patch('taskd.python.utils.log.logger.run_log.warning')
    @mock.patch('time.sleep')
    def test_init_clusterd_exception(self, mock_sleep, mock_run_log):
        from taskd.python.toolkit.recover_module import recover_manager
        manager = recover_manager.init_grpc_recover_manager()
        manager.grpc_stub.Init = mock.MagicMock()
        status = pb.Status()
        status.code = 0
        status.info = "ok"
        manager.grpc_stub.Init.side_effect = [Exception("Test exception"), status]
        manager.init_clusterd()
        mock_run_log.assert_called()

    def test_register_success(self):
        from taskd.python.toolkit.recover_module import recover_manager
        status = pb.Status()
        status.code = 0
        status.info = "ok"
        manager = recover_manager.init_grpc_recover_manager()
        manager.grpc_stub.Register = mock.Mock()
        manager.grpc_stub.Register = mock.Mock(return_value=status)
        result = manager.register(manager.client_info)
        self.assertEqual(result.code, 0)

    def test_register_exception(self):
        from taskd.python.toolkit.recover_module import recover_manager
        manager = recover_manager.init_grpc_recover_manager()
        manager.grpc_stub.Register = mock.MagicMock()
        manager.grpc_stub.Register.side_effect = Exception("Test exception")

        with self.assertRaises(Exception):
            manager.register(manager.client_info)
        manager.grpc_stub.Register.assert_called_once_with(manager.client_info)


class TestDLRecoverManagerConnectionCheck(TestCase):
    def setUp(self):
        from taskd.python.toolkit.recover_module import DLRecoverManager
        self.client_info = pb.ClientInfo(jobId="test_job", role="master agent")
        self.manager = DLRecoverManager(
            info=self.client_info,
            server_addr="localhost:50051",
        )
        DLRecoverManager.instance = None
        self.manager.grpc_stub = mock.MagicMock()
        self.manager.register_event = mock.MagicMock()
        self.manager.init_clusterd = mock.MagicMock()
        self.manager.register = mock.MagicMock()

    @mock.patch('taskd.python.toolkit.recover_module.recover_manager.time.sleep')
    def test_connection_check_success(self, mock_sleep):
        self.manager.grpc_stub.HealthCheck.return_value = None
        mock_sleep.side_effect = [None, KeyboardInterrupt]

        try:
            self.manager._connection_check()
        except KeyboardInterrupt:
            pass

        self.manager.grpc_stub.HealthCheck.assert_called_with(self.client_info)
        mock_sleep.assert_any_call(1)
        self.manager.register_event.clear.assert_not_called()

    @mock.patch('taskd.python.toolkit.recover_module.recover_manager.time.sleep')
    def test_connection_check_recover_success(self, mock_sleep):
        self.manager.grpc_stub.HealthCheck.side_effect = Exception("Connection failed")
        self.manager.register = mock.MagicMock(return_value=pb.Status(code=0))
        mock_sleep.side_effect = [None, KeyboardInterrupt]
        self.manager.register_event.is_set.return_value = False

        try:
            self.manager._connection_check()
        except KeyboardInterrupt:
            pass

        self.manager.register_event.clear.assert_called()
        self.manager.init_clusterd.assert_called_once()
        self.manager.register.assert_called_with(self.client_info)
        self.manager.register_event.set.assert_called_once()
        mock_sleep.assert_any_call(10)

    @mock.patch('taskd.python.toolkit.recover_module.recover_manager.time.sleep')
    def test_connection_check_recover_failed(self, mock_sleep):
        self.manager.grpc_stub.HealthCheck.side_effect = Exception("Connection failed")
        self.manager.register = mock.MagicMock(return_value=pb.Status(code=1))
        mock_sleep.side_effect = [None, KeyboardInterrupt]

        try:
            self.manager._connection_check()
        except KeyboardInterrupt:
            pass

        self.manager.register_event.set.assert_not_called()

    @mock.patch('taskd.python.toolkit.recover_module.recover_manager.time.sleep')
    def test_connection_check_event_set_during_wait(self, mock_sleep):
        self.manager.grpc_stub.HealthCheck.side_effect = Exception("Connection failed")
        def sleep_side_effect(seconds):
            if seconds == 10:
                self.manager.register_event.is_set.return_value = True
            raise KeyboardInterrupt
        mock_sleep.side_effect = sleep_side_effect

        try:
            self.manager._connection_check()
        except KeyboardInterrupt:
            pass

        self.manager.init_clusterd.assert_not_called()
        self.manager.register.assert_not_called()
        
