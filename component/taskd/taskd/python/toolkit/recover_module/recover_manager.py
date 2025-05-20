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
import os
import time
import threading
from abc import ABC, abstractmethod
from typing import Optional, List

import grpc
from taskd.python.toolkit.constants import constants
from taskd.python.toolkit.validator.cert_check import CertContentsChecker
from taskd.python.toolkit.recover_module.pb import recover_pb2 as pb
from taskd.python.toolkit.recover_module.pb import recover_pb2_grpc as service
from taskd.python.utils.log import run_log
from taskd.python.toolkit.validator.file_process import safe_get_file_info
from taskd.python.toolkit.recover_module import shared_data

MAX_CONNECT_GAP = 10
BASE_CONNECT_GAP = 1

import_flag = True
try:
    from mindio_ttp.controller_ttp import (tft_init_controller, tft_start_controller,
                                           tft_notify_controller_dump, tft_notify_controller_stop_train,
                                           tft_register_mindx_callback, tft_notify_controller_on_global_rank,
                                           tft_notify_controller_change_strategy, tft_destroy_controller,
                                           tft_query_high_availability_switch)
except ImportError:
    run_log.warning("mindio not found, process-rescheduling checkpoint-saving DO NOT work!")
    import_flag = False


class ProtoBufData:
    def __init__(self, fault_ranks: dict, change_strategy: str):
        self.fault_ranks = fault_ranks
        self.change_strategy = change_strategy


class RecoverManager(ABC):
    """
    Abstract class of process recover manager.
    """

    @abstractmethod
    def register(self, request: pb.ClientInfo) -> pb.Status:
        raise NotImplementedError

    @abstractmethod
    def start_subscribe(self):
        raise NotImplementedError


class DLRecoverManager(RecoverManager):
    """
    DLRecoverManager is a realization of RecoverManager.
    """
    instance = None

    def __new__(cls, *args, **kwargs):
        if not cls.instance:
            cls.instance = super().__new__(cls)
        return cls.instance

    def __init__(self, info: pb.ClientInfo, server_addr: str, secure_conn: bool = True, cert_path: str = ""):
        """
        __init__ construct ProcessRecoverDLManager instance
        :param info: client base information
        :param server_addr: server_addr like [domain_name|ip]:port
        :param secure_conn: 使用安全连接
        :param cert_path: 证书路径
        :return: None
        """
        super().__init__()
        self.client_info = info
        self.action_func_map = {
            'save_and_exit': tft_notify_controller_dump,
            'stop_train': tft_notify_controller_stop_train,
            'on_global_rank': tft_notify_controller_on_global_rank,
            'change_strategy': tft_notify_controller_change_strategy,
        }
        self.server_addr = server_addr
        self.lock = threading.Lock()
        self.secure_conn = secure_conn
        self.cert_path = cert_path
        self.grpc_channel = None
        self.grpc_stub = None
        self._init_client()

    def register(self, request: pb.ClientInfo) -> pb.Status:
        info = f"call Register, jobId={request.jobId}"
        run_log.info(info)
        try:
            return self.grpc_stub.Register(request)
        except Exception as e:
            raise e

    def init_clusterd(self):
        while True:
            try:
                status = self.grpc_stub.Init(self.client_info)
                if status.code == 0:
                    run_log.info("init process recover succeed")
                    break
                time.sleep(constants.SLEEP_GAP)
            except Exception as e:
                run_log.warning(f"init process recover catch exception:{e}")
                time.sleep(constants.SLEEP_GAP)
                self._init_client()
                continue

    def start_subscribe(self, frame: str = "pytorch"):
        run_log.info("call start_subscribe")
        i = 1
        while True:
            if shared_data.shared_data_inst.get_exit_flag():
                run_log.info("start_subscribe stop and init controller")
                tft_destroy_controller()
                init_mindio_controller(frame)
                shared_data.shared_data_inst.set_exit_flag(False)
            try:
                run_log.info("try to init and register again")
                self.init_clusterd()
                status = self.register(self.client_info)
                if status.code == 0:
                    i = 1
                self.__listen_signal()
            except RuntimeError as e:
                continue
            except Exception as e:
                info = (f"{self.client_info.role} subscribe signal for error: {e.__str__()}, "
                        f"task id is: {self.client_info.jobId}, retry it after a few second")
                run_log.warning(info)
            time.sleep(min(MAX_CONNECT_GAP, i * BASE_CONNECT_GAP))
            i += 1

    def _init_client(self):
        options = [
            (constants.GRPC_KEEPALIVE_TIME_MS, 3 * 1000),
            (constants.GRPC_KEEPALIVE_TIMEOUT_MS, 1 * 1000),
            (constants.GRPC_KEEPALIVE_PERMIT_WITHOUT_CALLS, 1),
            (constants.GRPC_MAX_PINGS_WITHOUT_DATA, 1)
        ]
        if not self.secure_conn:
            run_log.warning("using insecure channel is not safe.")
            self.grpc_channel = grpc.insecure_channel(self.server_addr, options)
            self.grpc_stub = service.RecoverStub(self.grpc_channel)
            return
        try:
            cert_bytes = safe_get_file_info(self.cert_path).encode()
            domain_name = CertContentsChecker().check_cert_info(cert_bytes)
        except Exception as err:
            run_log.error(f"check cert failed, {err} and set kill flag to end training")
            tft_destroy_controller()
            shared_data.shared_data_inst.set_kill_flag(True)
            raise ValueError from err
        ssl_credentials = grpc.ssl_channel_credentials(root_certificates=cert_bytes)
        options.append((constants.GRPC_SSL_TARGET_NAME_OVERRIDE, domain_name))
        self.grpc_channel = grpc.secure_channel(self.server_addr, ssl_credentials, options)
        self.grpc_stub = service.RecoverStub(self.grpc_channel)

    def __listen_signal(self):
        stream = self.grpc_stub.SubscribeProcessManageSignal(self.client_info)
        for signal in stream:
            if shared_data.shared_data_inst.get_exit_flag():
                run_log.info(f"__listen_signal raise exit_flag:{shared_data.shared_data_inst.get_exit_flag()}")
                raise RuntimeError
            if shared_data.shared_data_inst.get_kill_flag():
                run_log.info("listen_signal destroy_controller")
                tft_destroy_controller()
                return
            time.sleep(0.01)
            self.__signal_pipe_line(signal)

    def __signal_pipe_line(self, signal: pb.ProcessManageSignal):
        info = (f"jobId={signal.jobId}, event_id={signal.uuid}, signal_type={signal.signalType}, "
                f"actions={signal.actions}, faultRanks={signal.faultRanks}, changeStrategy={signal.changeStrategy}")
        run_log.debug(f"__signal_pipe_line receive signal: {info}")
        if signal.signalType == "killMaster":
            run_log.info("killMaster signal received, set kill flag")
            tft_destroy_controller()
            shared_data.shared_data_inst.set_kill_flag(True)
            return
        if signal.signalType == "keep-alive":
            run_log.debug("keep-alive signal now not handle")
            return
        if len(signal.actions) == 0:
            run_log.info("signal actions length is 0")
            return
        if signal.jobId != self.client_info.jobId:
            run_log.info(
                f"discard signal cause client_jobId={self.client_info.jobId}, but signal_jobId={signal.jobId}")
            return
        fault_rank_dict = {}
        for item in signal.faultRanks:
            fault_rank_dict[int(item.rankId)] = int(item.faultType)
        actions = list(set(signal.actions))
        pb_data = ProtoBufData(fault_rank_dict, signal.changeStrategy)
        for action in actions:
            self.__do_action(action=action, arg=pb_data)

    def __do_action(self, action: str, arg: ProtoBufData):
        try:
            run_log.info(f"do action {action}, jobId={self.client_info.jobId}, arg={arg}")
            self.lock.acquire()
            func = self.action_func_map.get(action)
            if func is None:
                raise Exception(f"action {action} unregistered")

            if action == 'save_and_exit':
                func()
            elif action == 'stop_train':
                func(arg.fault_ranks)
            elif action == 'on_global_rank':
                func(arg.fault_ranks)
            elif action == 'change_strategy':
                func(arg.change_strategy)

            run_log.info(f"do action {action} finish, jobId={self.client_info.jobId}, arg={arg}")
        except Exception as e:
            run_log.info(f"do action {action} err, err={e}, jobId={self.client_info.jobId}, arg={arg}")
        finally:
            self.lock.release()


class CallBackFuncs:
    def __init__(self):
        self.callback_func_dict = {
            constants.REPORT_FAULT_RANKS_CALLBACK: report_process_fault,
            constants.STOP_COMPLETE_CALLBACK: report_stop_complete,
            constants.REPORT_STRATEGIES_CALLBACK: report_recover_strategy,
            constants.REPORT_RESULT_CALLBACK: report_recover_status
        }


def get_instance(cls):
    # Check whether the class has the instance attribute.
    if not hasattr(cls, 'instance'):
        raise TypeError(f"{cls.__name__} is not a singleton class")
    # Check whether the instance has been created.
    if cls.instance is None:
        raise RuntimeError(f"{cls.__name__} instance has not been created yet")
    return cls.instance


def register_callback_func():
    callback = CallBackFuncs()
    for key, value in callback.callback_func_dict.items():
        rsp = tft_register_mindx_callback(key, value)
        if rsp != 0:
            run_log.error(f"Callback fun register failed, action：{key}, func:{value}")


def init_mindio_controller(frame: str = "pytorch"):
    run_log.info(f"will init frame {frame} mindio controller")
    world_size = "0"
    if frame == "pytorch":
        world_size = os.getenv("WORLD_SIZE")
    if frame == "mindspore":
        world_size = os.getenv("MS_WORKER_NUM")
    process_recover = os.getenv("PROCESS_RECOVER")

    if process_recover == "on":
        process_recover = True
    else:
        process_recover = False

    if world_size is None:
        run_log.error(f"init mindio controller failed, world_size: {world_size}")
        raise ValueError
    server_addr = os.getenv("POD_IP")
    ttp_port = os.getenv("TTP_PORT")
    if server_addr is None or ttp_port is None:
        run_log.error(f"start_mindio_controller failed,"
                      f" server_addr(POD_IP): {server_addr}, ttp_port(TTP_PORT):{ttp_port}"
                      f" if POD_IP/TTP_PORT is None, please add environment variables in yaml"
                      f" by referring to the document.")
        raise ValueError

    run_log.info(f"init mindio controller info: world_size:{int(world_size)}, process_recover:{process_recover}"
                 f"start mindio controller info: server_addr:{server_addr}, ttp_port:{int(ttp_port)}")
    try:
        tft_init_controller(constants.MINDX_START_CONTROLLER_RANK, int(world_size), False, process_recover)
        tft_start_controller(server_addr, int(ttp_port), False, "")
    except Exception as e:
        run_log.error(f"init mindio/start mindio controller failed, Exception: {e}")


def init_grpc_client(frame: str = "pytorch"):
    start_process = threading.Thread(target=init_grpc_process, args=(frame,))
    start_process.setDaemon(True)
    start_process.start()


def init_grpc_process(frame: str = "pytorch"):
    # minio not found, can not start controller and grpc client
    if not import_flag:
        return

    os.environ[constants.TORCH_AGENT_START] = "0"
    register_callback_func()
    init_mindio_controller(frame)

    register_retry_times = 0
    recover_manager = init_grpc_recover_manager()
    run_log.info("init_grpc_process start to init process recover")
    recover_manager.init_clusterd()
    run_log.info("init_grpc_process start check high_availability_switch")
    time_used = 0
    while True:
        switch_status = tft_query_high_availability_switch()
        if switch_status:
            run_log.info("high_availability_switch switch is on, start grpc client")
            break
        if time_used > constants.HIGH_AVAILABILITY_SWITCH_CHECK_TIMEOUT:
            run_log.warning(f"high_availability_switch switch is off for {time_used}s, destroy controller")
            tft_destroy_controller()
            return
        time.sleep(constants.SLEEP_GAP)
        time_used += constants.SLEEP_GAP
    # grpc client register and retry
    while True:
        if register_retry_times >= constants.GRPC_REGISTER_RETRY_TIME_LIMIT:
            run_log.error("recover_manager.register failed, retried to the limit")
            break
        try:
            rsp = recover_manager.register(recover_manager.client_info)
            if rsp.code == 0:
                run_log.info("recover_manager.register succeed")
                break
            else:
                register_retry_times = register_retry_times + 1
                run_log.warning(f"recover_manager.register failed rsp.code:  {rsp.code}, "
                                f"retry times: {register_retry_times}")
                time.sleep(constants.SLEEP_GAP)
        except Exception as e:
            register_retry_times = register_retry_times + 1
            run_log.warning(f"recover_manager.register failed! Exception:{e} retry time:{register_retry_times}")
            time.sleep(constants.SLEEP_GAP)
    # start listen grpc sever
    recover_manager.start_subscribe(frame)


def init_grpc_recover_manager() -> DLRecoverManager:
    job_id = os.getenv("MINDX_TASK_ID")
    server_addr = os.getenv("MINDX_SERVER_IP")
    if job_id is None or server_addr is None:
        run_log.error(f"job_id or server_addr is wrong, job_id：{job_id}, server_addr: {server_addr}")
        raise ValueError
    server_addr = server_addr + ":" + constants.GRPC_SERVER_PORT
    secure_conn = os.getenv(constants.ELASTIC_GRPC_SECURE_CONNECT_PATH)
    if secure_conn == "on":
        secure_conn = True
    else:
        secure_conn = False
    cert_path = os.getenv(constants.ELASTIC_GRPC_SECURE_CERTIFICATES_PATH)
    if cert_path is None:
        cert_path = ""
    client_info = pb.ClientInfo(jobId=job_id, role="master agent")
    run_log.info(f"DLRecoverManager init info: job_id：{job_id}, server_addr: {server_addr}"
                 f"secure_conn：{secure_conn}, cert_path: {cert_path}")
    return DLRecoverManager(client_info, server_addr, secure_conn, cert_path)


def get_recover_manager_instance() -> DLRecoverManager:
    try:
        recover_manager = get_instance(DLRecoverManager)
    except TypeError as e:
        raise e
    except RuntimeError:
        recover_manager = init_grpc_recover_manager()
    return recover_manager


def report_stop_complete(code: int, msg: str, fault_ranks: dict) -> int:
    recover_mgr = get_recover_manager_instance()
    info = (f"call ReportStopComplete, role={recover_mgr.client_info.role}, jobId={recover_mgr.client_info.jobId}, "
            f"server={recover_mgr.server_addr}, fault_ranks={fault_ranks}")
    run_log.info(info)
    request = pb.StopCompleteRequest()
    request.jobId = recover_mgr.client_info.jobId
    request.status.code = code
    request.status.info = msg
    for key, value in fault_ranks.items():
        fault_rank = pb.FaultRank(rankId=str(key), faultType=str(value))
        request.faultRanks.append(fault_rank)
    try:
        rsp = recover_mgr.grpc_stub.ReportStopComplete(request)
        if rsp.code != 0:
            run_log.warning(f"ReportStopComplete rsp code:{rsp.code}, info:{rsp.info}")
            recover_mgr.register(request=recover_mgr.client_info)
            rsp = recover_mgr.grpc_stub.ReportStopComplete(request)
        return rsp.code
    except Exception as e:
        raise e


def report_recover_strategy(fault_ranks: dict, strategy_list: list) -> int:
    recover_mgr = get_recover_manager_instance()
    info = (f"call ReportRecoverStrategy, role={recover_mgr.client_info.role}, jobId={recover_mgr.client_info.jobId}, "
            f"server={recover_mgr.server_addr}, fault_ranks={fault_ranks}, strategy_list={strategy_list}")
    run_log.info(info)
    request = pb.RecoverStrategyRequest()
    request.jobId = recover_mgr.client_info.jobId
    for key, value in fault_ranks.items():
        fault_rank = pb.FaultRank(rankId=str(key), faultType=str(value))
        request.faultRanks.append(fault_rank)
    for strategy in strategy_list:
        request.strategies.append(strategy)
    try:
        rsp = recover_mgr.grpc_stub.ReportRecoverStrategy(request)
        if rsp.code != 0:
            run_log.warning(f"ReportRecoverStrategy rsp code:{rsp.code}, info:{rsp.info}")
            recover_mgr.register(request=recover_mgr.client_info)
            rsp = recover_mgr.grpc_stub.ReportRecoverStrategy(request)
        return rsp.code
    except Exception as e:
        raise e


def report_recover_status(code: int, msg: str, fault_ranks: dict, strategy: str) -> int:
    recover_mgr = get_recover_manager_instance()
    info = (f"call ReportRecoverStatus, role={recover_mgr.client_info.role}, jobId={recover_mgr.client_info.jobId}, "
            f"server={recover_mgr.server_addr}")
    run_log.info(info)
    request = pb.RecoverStatusRequest()
    request.jobId = recover_mgr.client_info.jobId
    request.status.code = code
    request.status.info = msg
    request.strategy = strategy
    try:
        rsp = recover_mgr.grpc_stub.ReportRecoverStatus(request)
        if rsp.code != 0:
            run_log.warning(f"ReportRecoverStatus rsp code:{rsp.code}, info:{rsp.info}")
            recover_mgr.register(request=recover_mgr.client_info)
            rsp = recover_mgr.grpc_stub.ReportRecoverStatus(request)
        return rsp.code
    except Exception as e:
        raise e


def report_process_fault(fault_ranks: dict) -> int:
    recover_mgr = get_recover_manager_instance()
    info = (f"call ReportProcessFault, role={recover_mgr.client_info.role}, jobId={recover_mgr.client_info.jobId}, "
            f"server={recover_mgr.server_addr}, fault_ranks={fault_ranks}")
    run_log.info(info)
    request = pb.ProcessFaultRequest()
    request.jobId = recover_mgr.client_info.jobId
    for key, value in fault_ranks.items():
        fault_rank = pb.FaultRank(rankId=str(key), faultType=str(value))
        request.faultRanks.append(fault_rank)
    try:
        rsp = recover_mgr.grpc_stub.ReportProcessFault(request)
        if rsp.code != 0:
            run_log.warning(f"ReportProcessFault rsp code:{rsp.code}, info:{rsp.info}")
            recover_mgr.register(request=recover_mgr.client_info)
            rsp = recover_mgr.grpc_stub.ReportProcessFault(request)
        return rsp.code
    except Exception as e:
        raise e
