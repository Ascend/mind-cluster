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

from taskd.python.cython_api import cython_api
from taskd.python.utils.log import run_log
from taskd.python.adaptor.pytorch.group_info import dump_group_info

class Worker:
    """
    Worker is a framework of training thread management
    """

    def __init__(self, global_rank: int, framework: str = "pt"):
        self.global_rank = global_rank
        self.framework = framework

    def start(self) -> bool:
        return self._start_up_monitor()

    def init_worker(self, upper_limit_of_disk_in_mb: int) -> bool:
        if cython_api.lib is None:
            run_log.error("the libtaskd.so has not been loaded!")
            return False
        init_taskd_func = cython_api.lib.InitWorker
        node_rank_str = os.getenv('RANK') or os.getenv('MS_NODE_RANK') or '-1'
        try:
            node_rank = int(node_rank_str)
        except ValueError:
            run_log.error(f'invalid node_rank_str {node_rank_str}')
            return False
        result = init_taskd_func(self.global_rank, node_rank, upper_limit_of_disk_in_mb)
        if result == 0:
            dump_group_info(self.global_rank)
            run_log.info("Successfully init taskd monitor")
            return True
        run_log.warning(f"failed to init taskd monitor with ret code:f{result}")
        return False

    def _start_up_monitor(self) -> bool:
        try:
            if cython_api.lib is None:
                run_log.error("the libtaskd.so has not been loaded!")
                return False
            run_log.info(f"begin cython_api.lib.StartMonitorClient")
            start_monitor_client_func = cython_api.lib.StartMonitorClient
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
