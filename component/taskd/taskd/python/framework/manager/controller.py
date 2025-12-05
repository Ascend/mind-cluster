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
import ctypes
import json
import os
import time
from dataclasses import dataclass, field, asdict
from typing import Dict

from taskd.python.cython_api import cython_api
from taskd.python.framework.common.utils import get_hccl_switch_nic_timeout
from taskd.python.toolkit.constants import constants
from taskd.python.utils.log import run_log
from mindio_ttp.controller_ttp import (tft_init_controller, tft_start_controller,
                                       tft_notify_controller_dump, tft_notify_controller_stop_train,
                                       tft_register_mindx_callback, tft_notify_controller_on_global_rank,
                                       tft_notify_controller_change_strategy, tft_destroy_controller,
                                       tft_query_high_availability_switch
                                       )
# Compatibility handling: tft_notify_controller_prepare_action may not exist in some versions
try:
    from mindio_ttp.controller_ttp import tft_notify_controller_prepare_action
except (ImportError, AttributeError) as ex:
    run_log.warning(f"Failed to import tft_notify_controller_prepare_action, "
                    f"current version may not support this method: {ex}")
    tft_notify_controller_prepare_action = None


def _compatible_prepare_action(action, fault_ranks):
    run_log.warning(f"tft_notify_controller_prepare_action is not supported in current version, "
                    f"skipping execution. action={action}, fault_ranks={fault_ranks}")

# Select function based on whether tft_notify_controller_prepare_action was successfully imported
_prepare_action_func = tft_notify_controller_prepare_action \
    if tft_notify_controller_prepare_action is not None \
    else _compatible_prepare_action

action_func_map = {
            constants.SAVE_AND_EXIT: tft_notify_controller_dump,
            constants.STOP_TRAIN: tft_notify_controller_stop_train,
            constants.PAUSE_TRAIN: tft_notify_controller_stop_train,
            constants.ON_GLOBAL_RANK: tft_notify_controller_on_global_rank,
            constants.CHANGE_STRATEGY: tft_notify_controller_change_strategy,
            constants.HOT_SWITCH: _prepare_action_func,
            constants.STOP_SWITCH: _prepare_action_func,
            constants.NEW_POD_RUNNING: _prepare_action_func,
        }


@dataclass
class ControllerMessage:
    action: str
    code: int
    msg: str
    strategy: str
    params: str
    timeout: int
    actions: list = field(default_factory=list)
    strategy_list: list = field(default_factory=list)
    fault_ranks: Dict[int, int] = field(default_factory=dict)


class CallBackFuncs:
    def __init__(self):
        self.callback_func_dict = {
            constants.REPORT_FAULT_RANKS_CALLBACK: report_process_fault,
            constants.STOP_COMPLETE_CALLBACK: report_stop_complete,
            constants.REPORT_STRATEGIES_CALLBACK: report_recover_strategy,
            constants.REPORT_RESULT_CALLBACK: report_recover_status
        }


def init_controller():
    register_callback_func()
    run_log.info(f"will init mindio controller")
    world_size = "0"
    world_size = os.getenv(constants.WORLD_SIZE) or os.getenv(constants.MS_WORKER_NUM)
    process_recover = os.getenv(constants.PROCESS_RECOVER)
    if process_recover == "on":
        process_recover = True
    else:
        process_recover = False

    elastic_training = False
    high_availability_strategy = os.getenv(constants.HIGH_AVAILABILITY_STRATEGY)
    if high_availability_strategy:
        strategy_list = high_availability_strategy.split(',')
        if constants.ELASTIC_TRAINING in strategy_list:
            elastic_training = True

    if world_size is None:
        run_log.error(f"init mindio controller failed, world_size: {world_size}")
        raise ValueError
    server_addr = os.getenv(constants.POD_IP)
    ttp_port = os.getenv(constants.TTP_PORT)
    if server_addr is None or ttp_port is None:
        run_log.error(f"start_mindio_controller failed,"
                      f" server_addr(POD_IP): {server_addr}, ttp_port(TTP_PORT):{ttp_port}"
                      f" if POD_IP/TTP_PORT is None, please add environment variables in yaml"
                      f" by referring to the document.")
        raise ValueError

    run_log.info(f"init mindio controller info: world_size:{int(world_size)}, process_recover:{process_recover}, "
                 f"downgrade_train:{elastic_training} "
                 f"start mindio controller info: server_addr:{server_addr}, ttp_port:{int(ttp_port)}")
    try:
        tft_init_controller(constants.MINDX_START_CONTROLLER_RANK, int(world_size), False, process_recover, elastic_training)
        tft_start_controller(server_addr, int(ttp_port), False, "")
    except Exception as e:
        run_log.error(f"init mindio/start mindio controller failed, Exception: {e}")


def register_callback_func():
    callback = CallBackFuncs()
    for key, value in callback.callback_func_dict.items():
        rsp = tft_register_mindx_callback(key, value)
        if rsp != 0:
            run_log.error(f"Callback fun register failed, action:{key}, func:{value}")


def backend_send_callback(data_ptr) -> int:
    try:
        data_str = ctypes.cast(data_ptr, ctypes.c_char_p).value.decode('utf-8')
        run_log.info(f"manager recv data: {data_str}")
        data_json = json.loads(data_str)
        if data_json is None:
            run_log.error(f"manager recv data is None")
            return 1
        fault_ranks = data_json.get("fault_ranks", {})
        msg_fault_ranks = {}
        for key, value in fault_ranks.items():
            try:
                msg_fault_ranks[int(key)] = int(value)
            except (ValueError, TypeError) as e:
                run_log.error(f"Invalid fault_rank key-value pair: {key}={value}, error: {e}")
                return 1
        data = ControllerMessage(
            actions=data_json.get("actions", []),
            action=data_json.get("action", ""),
            code=data_json.get("code", 0),
            msg=data_json.get("msg", ""),
            strategy=data_json.get("strategy", ""),
            strategy_list=data_json.get("strategy_list", []),
            fault_ranks=msg_fault_ranks,
            params=data_json.get("params", ""),
            timeout=data_json.get("timeout", 0)
        )
    except Exception as e:
        run_log.error(f"backend_callback parse message failed, reason: {e}")
        return 1
    run_log.info(f"manager receive data: {data}")
    for action in data.actions:
        send_msg_to_controller(action, data)
    return 0


def send_msg_to_controller(action, data):
    if action == constants.RESTARTCONTROLLER:
        restart_controller()
        return
    if action == constants.DESTRYCONTROLLER:
        run_log.info("destroy mindio controller")
        tft_destroy_controller()
        return

    try:
        run_log.info(f"do action {action}, data={data}")
        func = action_func_map.get(action)
        if func is None:
            raise Exception(f"action {action} unregistered")
        if action == 'save_and_exit':
            func()
        elif action == "pause_train":
            timeout = data.timeout
            if timeout == 0:
                timeout = get_hccl_switch_nic_timeout()
            func(data.fault_ranks, constants.STOP_TRAIN_PAUSE, timeout)
            run_log.info(f"will pause train, timeout={timeout}")
        elif action == 'stop_train':
            func(data.fault_ranks, constants.STOP_TRAIN_ABORT)
        elif action == 'on_global_rank':
            if data.timeout == 0:
                func(data.fault_ranks)
            else:
                func(data.fault_ranks, data.timeout)
        elif action == 'change_strategy':
            func(data.strategy, data.params)
        elif action == 'hot switch':
            run_log.info(f"notify prepare not switch, fault_rank={data.fault_ranks}")
            func(action, data.fault_ranks)
        elif action == 'stop switch':
            run_log.info(f"notify stop not switch, fault_rank={data.fault_ranks}")
            func(action, data.fault_ranks)
        elif action == 'new pod running':
            # only print log info
            run_log.info("new pod running")
        run_log.info(f"do action {action} finish,  data={data}")
    except Exception as e:
        run_log.info(f"do action {action} err, err={e}, data={data}")


def restart_controller():
    run_log.info("restart controller")
    tft_destroy_controller()
    time.sleep(1)
    init_controller()
    run_log.info("restart controller finish")


def report_stop_complete(code: int, msg: str, fault_ranks: dict):
    run_log.info(f"call ReportStopComplete, msg:{msg}, fault_ranks={fault_ranks}")
    message = ControllerMessage(
        actions=[],
        action="stop_complete",
        code=code,
        msg=msg,
        fault_ranks=fault_ranks,
        strategy="",
        strategy_list=[],
        params="",
        timeout=0
    )
    controller_send_to_backend(message)


def report_recover_strategy(fault_ranks: dict, strategy_list: list):
    run_log.info(f"call ReportRecoverStrategy, fault_ranks:{fault_ranks}, strategy_list:{strategy_list}")
    message = ControllerMessage(
        actions=[],
        action="recover_strategy",
        code=0,
        msg="",
        fault_ranks=fault_ranks,
        strategy_list=strategy_list,
        strategy="",
        params="",
        timeout=0

    )
    controller_send_to_backend(message)


def report_recover_status(code: int, msg: str, fault_ranks: dict, strategy: str):
    run_log.info(f"call ReportRecoverStatus, strategy: {strategy}, msg: {msg}")
    message = ControllerMessage(
        actions=[],
        action="recover_status",
        code=code,
        msg=msg,
        fault_ranks=fault_ranks,
        strategy=strategy,
        strategy_list=[],
        params="",
        timeout=0
    )
    controller_send_to_backend(message)


def report_process_fault(fault_ranks: dict):
    run_log.info(f"call ReportProcessFault, fault_ranks:{fault_ranks}")
    message = ControllerMessage(
        actions=[],
        action="process_fault",
        code=0,
        msg="",
        fault_ranks=fault_ranks,
        strategy="",
        strategy_list=[],
        params="",
        timeout=0
    )
    controller_send_to_backend(message)


def controller_send_to_backend(message):
    msg_json = json.dumps(asdict(message)).encode('utf-8')
    run_log.info(f"controller_send_to_backend msg_json: {msg_json}")
    try:
        if cython_api.lib is None:
            run_log.error(f'controller_send_to_backend cython_api.lib is None')
            return
        send_func = cython_api.lib.SendMessageToBackend
        send_func.argtypes = [ctypes.c_char_p]
        send_func.restype = ctypes.c_int
        res = send_func(msg_json)
        if res != 0:
            run_log.error(f'controller_send_to_backend send message failed, res: {res}')
    except Exception as e:
        run_log.error(f'controller_send_to_backend send message failed, error: {e}')
    return
