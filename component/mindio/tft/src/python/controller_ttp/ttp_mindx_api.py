#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import threading
from typing import Callable, Any, Optional, List
import ttp_logger
import ttp_c2python_api

RET_OK = 0
RET_ERROR = 1


class MindxEngineHandler:
    _instance = None
    _lock = threading.Lock()
    _action_func_map = None

    def __new__(cls, *args, **kwargs):
        with cls._lock:
            if cls._instance is None:
                cls._instance = super().__new__(cls, *args, **kwargs)
                cls._action_func_map = {}
        return cls._instance

    def __init__(self):
        self.lock = threading.Lock()

    @staticmethod
    def support_actions() -> List:
        return ['report_fault_ranks', 'report_stop_complete', 'report_strategies', 'report_result']

    def check_supported_action(self, action: str) -> bool:
        return action in self.support_actions()

    def register_api_callback(self, action: str, func: Callable):
        with self.lock:
            if not self.check_supported_action(action):
                ttp_logger.LOGGER.warning(f"register action[{action}] fail, "
                                          f"only support action {MindxEngineHandler.support_actions()}")
                return RET_ERROR

            if func is None or not callable(func):
                ttp_logger.LOGGER.error(f"register action[{action}] fail, func must be a callable!")
                return RET_ERROR

            self._action_func_map[action] = func
            ttp_logger.LOGGER.info(f"register action[{action}] success")
            return RET_OK

    def execute_callback(self, action: str, args: Optional[Any]):
        with self.lock:
            try:
                ttp_logger.LOGGER.info(f"do action {action}, arg={repr(args)}")
                if action not in self._action_func_map:
                    ttp_logger.LOGGER.error(f"action {action} unregistered")
                    return RET_ERROR

                func = self._action_func_map[action]
                if args is None:
                    func()
                else:
                    func(*args)
                return RET_OK
            except Exception as e:
                ttp_logger.LOGGER.info(f"do action {action} err, err={e.__str__()}, arg={args}")
                return RET_ERROR

    def is_all_action_registered(self):
        count = len(self._action_func_map)
        all_registered = count == len(self.support_actions())
        ttp_logger.LOGGER.info(f" {count} actions api has been registered..., all registered:{all_registered}")
        return RET_OK if all_registered else RET_ERROR


mindx_handler: MindxEngineHandler = MindxEngineHandler()


def tft_register_mindx_callback(action: str, func: Callable):
    ret = mindx_handler.register_api_callback(action, func)
    return RET_OK if ret == RET_OK else RET_ERROR


def tft_notify_controller_stop_train(fault_ranks: dict, stop_type: str = "stop", timeout: int = None):
    if stop_type == "stop":
        ret = ttp_c2python_api.mindx_stop_train_callback(fault_ranks)
    elif stop_type == "pause":
        ret = ttp_c2python_api.mindx_pause_train_callback(timeout)
    else:
        ret = RET_ERROR
    return RET_OK if ret == RET_OK else RET_ERROR


def tft_notify_controller_on_global_rank(fault_ranks: dict, time: int = 1):
    ret = ttp_c2python_api.mindx_notify_fault_callback(fault_ranks, time)
    return RET_OK if ret == RET_OK else RET_ERROR


def tft_notify_controller_prepare_action(action: str, fault_ranks: dict = None):
    fault_ranks = fault_ranks or {}  # for code check
    ret = ttp_c2python_api.mindx_prepare_action_callback(action, fault_ranks)
    return RET_OK if ret == RET_OK else RET_ERROR


def tft_notify_controller_change_strategy(strategy: str, params: str = ""):
    ret = ttp_c2python_api.mindx_change_strategy_callback(strategy, params)
    return RET_OK if ret == RET_OK else RET_ERROR


def tft_notify_controller_dump():
    ret = ttp_c2python_api.mindx_notify_dump_callback()
    return RET_OK if ret == RET_OK else RET_ERROR


def report_fault_ranks_callback(error_rank_dict, error_rank_code):
    ret = mindx_handler.execute_callback('report_fault_ranks', (error_rank_dict, error_rank_code,))
    return RET_OK if ret == RET_OK else RET_ERROR


def report_stop_complete_callback(code, msg, error_rank_dict):
    ret = mindx_handler.execute_callback('report_stop_complete', (code, msg, error_rank_dict,))
    return RET_OK if ret == RET_OK else RET_ERROR


def report_strategies_callback(error_rank_dict, strategy_list):
    ret = mindx_handler.execute_callback('report_strategies', (error_rank_dict, strategy_list,))
    return RET_OK if ret == RET_OK else RET_ERROR


def report_result_callback(code, msg, error_rank_dict, curr_strategy):
    ret = mindx_handler.execute_callback('report_result', (code, msg, error_rank_dict, curr_strategy,))
    return RET_OK if ret == RET_OK else RET_ERROR


def register_check_callback():
    ret = mindx_handler.is_all_action_registered()
    return RET_OK if ret == RET_OK else RET_ERROR
