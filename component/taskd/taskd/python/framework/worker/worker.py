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
import threading
import time
from ctypes import CFUNCTYPE, POINTER, c_int, c_bool, c_char_p, create_string_buffer, addressof
from typing import List

from taskd.python.constants import constants
from taskd.python.cython_api import cython_api
from taskd.python.utils.log import run_log


class StressStatus:
    def __init__(self):
        self.cur_result = ""
        self.c_buf = None
        self.code_result_map = {
            constants.StressTestInit: "",
            constants.StressTestOK: f"{constants.StressTestOK}-OK",
            constants.StressTestExecFail: f"{constants.StressTestExecFail}-exec failed",
            constants.StressTestFindFault: f"{constants.StressTestFindFault}-find fault rank",
            constants.StressTestTimeout: f"{constants.StressTestTimeout}-exec timeout",
            constants.StressTestVolRecoverFail: f"{constants.StressTestVolRecoverFail}-voltage recovery failed",
        }
        self.ms_op_map = {0: "aic", 1: "hccs"}

    def get_stress_test_result(self):
        return self.cur_result

    def set_stress_test_result(self, code):
        self.cur_result = self.code_result_map.get(code, f"{constants.StressTestExecFail}-exec failed")


class Worker:
    """
    Worker is a framework of training thread management
    """

    def __init__(self, global_rank: int, framework: str = "pt"):
        self._stress_test_callback = None
        self._switch_nic_callback = None
        self.global_rank = global_rank
        self.framework = framework
        self.stress_status = StressStatus()

    @staticmethod
    def destroy() -> bool:
        if cython_api.lib is None:
            run_log.error("destroy_taskd_worker: the libtaskd.so has not been loaded")
            return False
        try:
            destroy_proxy_func = cython_api.lib.DestroyTaskdWorker
            if destroy_proxy_func is None:
                run_log.error("destroy_taskd_worker: func DestroyTaskdWorker has not been loaded from libtaskd.so")
                return False
            destroy_proxy_func()
        except Exception as e:
            run_log.error(f"destroy_taskd_worker: encounter exception: {e}")
            return False
        run_log.info("successfully destroy taskd worker")
        return True

    def start(self) -> bool:
        return self._start_up_monitor()

    def init_worker(self, upper_limit_of_disk_in_mb: int) -> bool:
        if cython_api.lib is None:
            run_log.error("the libtaskd.so has not been loaded")
            return False
        init_taskd_func = cython_api.lib.InitWorker
        if init_taskd_func is None:
            run_log.error("init_worker: func InitWorker has not been loaded from libtaskd.so")
            return False

        node_rank = None
        try:
            if self.framework == "ms":
                node_rank = int(os.getenv("MS_NODE_RANK"))
            elif self.framework == "pt":
                # current "RANK" is global rank， not node rank
                node_rank = int(os.getenv("RANK")) // int(os.getenv("LOCAL_WORLD_SIZE"))
            if node_rank < 0 or node_rank is None:
                run_log.error('invalid node id')
                return False
            run_log.info(f"{self.framework} node_rank: {node_rank}")
        except ValueError:
            run_log.error('invalid node rank')
            return False

        result = init_taskd_func(self.global_rank, node_rank, upper_limit_of_disk_in_mb)
        if result == 0:
            self.register_callback()
            run_log.info("Successfully init taskd monitor")
            return True
        run_log.warning(f"failed to init taskd monitor with ret code:f{result}")
        return False

    def register_callback(self):
        c_switch_nic_callback = CFUNCTYPE(c_bool, POINTER(c_int), POINTER(c_bool), c_int)
        c_stress_test_callback = CFUNCTYPE(c_char_p, c_int)
        self._switch_nic_callback = c_switch_nic_callback(self._switch_nic)
        self._stress_test_callback = c_stress_test_callback(self._stress_test)
        if cython_api.lib:
            cython_api.lib.RegisterSwitchCallback(self._switch_nic_callback)
            cython_api.lib.RegisterStressTestCallback(self._stress_test_callback)
            run_log.info("Successfully register callback func")
        else:
            run_log.error("register callback func failed")

    def _start_up_monitor(self) -> bool:
        try:
            if cython_api.lib is None:
                run_log.error("the libtaskd.so has not been loaded")
                return False
            run_log.info(f"begin cython_api.lib.StartMonitorClient")
            start_monitor_client_func = cython_api.lib.StartMonitorClient
            if start_monitor_client_func is None:
                run_log.error("start_up_monitor: func StartMonitorClient has not been loaded from libtaskd.so")
                return False
            run_log.info(f"end cython_api.lib.StartMonitorClient")
            result = start_monitor_client_func()
            if result == 0:
                run_log.info(f"Successfully start monitor client for rank:{self.global_rank}")
                return True
            run_log.warning(f"failed to start up monitor client with ret code:f{result}")
            return False
        except Exception as e:
            run_log.error(f"failed to start up monitro client, e:{e}")
            return False

    def _do_switch_nic(self, ranks: List[int], ops: List[bool]) -> bool:
        try:
            if self.framework == "pt":
                from torch_npu.npu import _comm_switch_nic
                switch_func = _comm_switch_nic
            elif self.framework == "ms":
                from mindspore.communication.management import _comm_switch_nic
                switch_func = _comm_switch_nic
            run_log.info(f"rankID:{self.global_rank} ranks:{ranks}, ops:{ops}")
            ret = switch_func(ranks, ops)
            run_log.info(f"rankID:{self.global_rank} ret:{ret}")
            return ret
        except Exception as e:
            run_log.error(e)
            return False

    def _switch_nic(self, ranks_ptr, ops_ptr, length):
        try:
            ranks = [ranks_ptr[i] for i in range(length)]
            ops = [ops_ptr[i] for i in range(length)]
            return self._do_switch_nic(ranks, ops)
        except Exception as e:
            run_log.error(f"callback failed: {str(e)}")
            return False

    def _stress_test(self, op):
        run_log.info(f"rank:{self.global_rank}, ops:{op}")
        self._exec_stress_test(op)
        ret = self.stress_status.get_stress_test_result()
        run_log.info(f"{self.global_rank} current op {op} result: {ret}")
        self.stress_status.c_buf = create_string_buffer(ret.encode("utf-8"))
        return addressof(self.stress_status.c_buf)

    def _exec_stress_test(self, op):
        stress_test_func = None
        try:
            op = self.stress_status.ms_op_map[op]
            if self.framework == "pt":
                import torch
                import torch_npu
                from torch_npu.npu import stress_detect
                stress_test_func = stress_detect
                local_rank = int(os.environ["LOCAL_RANK"])
                torch.npu.set_device(torch.device('npu:{}'.format(local_rank)))
                run_log.info(f"{self.global_rank} set local rank:{local_rank}")
            elif self.framework == "ms":
                from mindspore.utils import stress_detect
                stress_test_func = stress_detect
        except Exception as e:
            self.stress_status.set_stress_test_result(constants.StressTestExecFail)
            run_log.error(f"stress detect init failed: {e}, rank: {self.global_rank}")
            return

        run_log.info(f"{self.global_rank} stress detect init finish")
        try:
            run_log.info(f"{self.global_rank} start {op} exec stress test...")
            code = stress_test_func(op)
            run_log.info(f"{self.global_rank} stress test {op} finish, code: {code}")
        except Exception as err:
            run_log.error(f"{self.global_rank} exception: {err} maybe the voltage has not recovered.")
            code = constants.StressTestVolRecoverFail

        self.stress_status.set_stress_test_result(code)