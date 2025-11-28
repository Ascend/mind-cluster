#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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
import queue
import time
import threading
from typing import Callable

import torch

import mindio_acp
from mindio_acp.common import mindio_logger
from mindio_acp.acc_checkpoint.utils.utils import time_used

logging = mindio_logger.LOGGER


class CheckpointAsyncSaverMixin(object):
    __cond = threading.Condition()
    __running_count = 0
    __optimizer_save_thread = None
    __model_saved_thread = None

    __model_thread_queue = queue.Queue(maxsize=0)
    __optimizer_thread_queue = queue.Queue(maxsize=0)
    __background_saving_checkpoint = False

    def __init__(self, rank):
        self._device = torch.npu.current_device()
        self.__rank = rank

    @time_used
    def async_save_model_checkpoint(self, checkpoint_name: str, model: dict):
        self.__background_saving_checkpoint = True

        logging.debug("Rank: %d start to save model checkpoint %s", self.__rank, checkpoint_name)
        if self.__model_saved_thread is None:
            self.__model_saved_thread = threading.Thread(
                target=self.__model_thread_process,
            )
            self.__model_saved_thread.start()

        with self.__cond:
            self.__running_count += 1
        self.__model_thread_queue.put({"model": model, "checkpoint_name": checkpoint_name})

    @time_used
    def async_save_optimizer_checkpoint(self, checkpoint_name: str, get_parameter_state_func: Callable,
                                        has_sig_para: bool):
        self.__background_saving_checkpoint = True

        logging.debug("Rank: %d start to save optimizer checkpoint %s", self.__rank, checkpoint_name)
        if self.__optimizer_save_thread is None:
            self.__optimizer_save_thread = threading.Thread(
                target=self.__optimizer_thread_process,
            )
            self.__optimizer_save_thread.start()

        with self.__cond:
            self.__running_count += 1
        self.__optimizer_thread_queue.put(
            {
                "checkpoint_name": checkpoint_name,
                "get_parameter_state_func": get_parameter_state_func,
                "has_sig_para": has_sig_para,
            }
        )

    @time_used
    def wait_d2h_checkpoint_finished(self):
        start_time = time.time()

        if self.__background_saving_checkpoint:
            with self.__cond:
                while self.__running_count > 0:
                    self.__cond.wait()
            self.__background_saving_checkpoint = False

        torch.distributed.barrier()
        used_time = time.time() - start_time
        logging.debug("Rank: %d wait time in next step is %f seconds", self.__rank, used_time)

    def try_notify(self):
        with self.__cond:
            self.__running_count -= 1
            if self.__running_count <= 0:
                self.__cond.notify()

    @time_used
    def __model_thread_handler(self, msg):
        self.__torch_mindio_save(msg["model"], msg["checkpoint_name"])
        self.try_notify()

    @time_used
    def __optimizer_thread_handler(self, msg):
        checkpoint_name = msg["checkpoint_name"]
        get_parameter_state_func = msg["get_parameter_state_func"]
        has_sig_para = msg["has_sig_para"]

        if has_sig_para:
            states = get_parameter_state_func(notify_callback=self.try_notify)
        else:
            states = get_parameter_state_func()
            self.try_notify()
        if states:
            logging.debug("mindio_acp save optimizer checkpoint dict")
            self.__torch_mindio_save(states, checkpoint_name)

    @time_used
    def __torch_mindio_save(self, state_dict, checkpoint_name):
        if isinstance(checkpoint_name, str):
            mindio_acp.save(state_dict, checkpoint_name)
        else:
            mindio_acp.multi_save(state_dict, checkpoint_name)

    def __model_thread_process(self):
        torch.npu.set_device(self._device)
        while True:
            try:
                msg = self.__model_thread_queue.get(timeout=5)
            except Exception:
                if not threading.main_thread().is_alive():
                    return
                continue
            self.__model_thread_handler(msg)

    def __optimizer_thread_process(self):
        torch.npu.set_device(self._device)
        while True:
            try:
                msg = self.__optimizer_thread_queue.get(timeout=5)
            except Exception:
                if not threading.main_thread().is_alive():
                    return
                continue
            self.__optimizer_thread_handler(msg)