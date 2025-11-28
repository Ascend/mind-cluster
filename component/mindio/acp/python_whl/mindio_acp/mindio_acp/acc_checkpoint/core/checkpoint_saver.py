#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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
import threading
from typing import Callable

import torch
import torch.distributed

from mindio_acp.acc_checkpoint.utils.utils import time_used
from mindio_acp.common import mindio_logger

logging = mindio_logger.LOGGER


def import_torch_npu():
    global torch_npu
    import torch_npu


class CheckpointSaverMixin(object):
    __cond = threading.Condition()
    __running_count = 0
    __checkpoint_thread = None
    __checkpoint_queue = queue.Queue(maxsize=0)
    __checkpoint_copy_stream = None
    __background_saving_checkpoint = False

    def __init__(self):
        import_torch_npu()
        if not self.__checkpoint_copy_stream:
            self.__checkpoint_copy_stream = torch_npu.npu.Stream(device=torch.npu.current_device())

    @time_used
    def _async_save_checkpoint(self, thread_func: Callable, kwargs):
        self.__background_saving_checkpoint = True
        if self.__checkpoint_thread is None:
            self.__checkpoint_thread = threading.Thread(
                target=CheckpointSaverMixin.__checkpoint_thread_process,
                args=(self,)
            )
            self.__checkpoint_thread.start()

        with self.__cond:
            self.__running_count += 1
        self.__checkpoint_queue.put({'func': thread_func, 'kwargs': kwargs})

    def _async_checkpoint_count(self):
        with self.__cond:
            self.__running_count -= 1
            if self.__running_count <= 0:
                self.__cond.notify()

    def _get_checkpoint_copy_stream(self):
        return self.__checkpoint_copy_stream

    def __checkpoint_thread_process(self):
        while True:
            try:
                msg = self.__checkpoint_queue.get(timeout=5)
            except Exception:
                if not threading.main_thread().is_alive():
                    return
                continue
            try:
                func = msg['func']
                func(msg['kwargs'])
            except Exception as e:
                logging.error('[mindio_acp] async run fault. Exception is: {}.'.format(e))
