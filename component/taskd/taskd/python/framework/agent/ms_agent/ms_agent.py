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

from taskd.python.framework.agent.ms_mgr.ms_utils import check_monitor_res_valid, calculate_global_rank
from taskd.python.toolkit.constants import constants
from taskd.python.utils.log import run_log
from taskd.python.framework.agent.base_agent.agent_network import init_network_client
from taskd.python.framework.agent.base_agent.base_agent import BaseAgent, REPORT_CODE
from taskd.python.framework.common.type import AgentReportInfo


class MsAgent(BaseAgent):
    """
    MsAgent is for MindSpore to manage training process.
    """

    RANK_STATUS_UNHEALTHY = "UNHEALTHY"
    RANK_STATUS_UNKNOWN = "UNKNOWN"
    RANK_STATUS_INIT = "INIT"
    RANK_STATUS_HEALTHY = "HEALTHY"
    RANK_STATUS_STOPPED = "STOPPED"
    RANK_STATUS_SUCCEEDED = "SUCCEEDED"
    RANK_STATUS_FAILED = "FAILED"
    FRAMEWORK_MS_NAME = "mindspore"

    def __init__(self, network_config=None):
        super().__init__()
        self.all_rank_succeed = False
        self.monitor_interval = 5
        self.node_rank = os.getenv("MS_NODE_RANK")
        self.rank_pids = []
        self.node_global_rank_ids = []
        self.network_config = network_config
        self.command_map = {
            'START': self.initialize_workers,
            'STOP': self.stop_workers,
            'EXIT': self.exit_agent,
            'RESTART': self.restart_workers,
            'GRACE_EXIT': self.grace_exit,
        }

    def start(self):
        kill_worker_func = self._func_map.get('KILL_WORKER')
        start_worker_func = self._func_map.get('START_ALL_WORKER')
        monitor_func = self._func_map.get('MONITOR')
        if kill_worker_func is None or start_worker_func is None or monitor_func is None:
            raise Exception(f"{self.FRAMEWORK_MS_NAME} hasn't fully registered all callbacks")
        self.node_global_rank_ids = calculate_global_rank()
        init_network_client(self.network_config, self.msg_queue)
        self.check_network()
        start_worker_func()

        while True:
            self.send_message_to_manager('KEEP_ALIVE', 0, AgentReportInfo())
            time.sleep(self.monitor_interval)
            self.handle_message()
            # After entering the loop, first obtain the process status once.
            ms_proc_status = monitor_func([constants.MONITOR_ALL_WORKERS])
            run_log.debug(f"nodeRank:{self.node_rank} has got mindspore process status:{ms_proc_status}")
            if not check_monitor_res_valid(ms_proc_status):
                run_log.warning(f"monitor not return a valid result, but {ms_proc_status}")
                continue
            fault_ranks = self.update_rank_status(ms_proc_status)
            self.report_fault_rank(fault_ranks)



    def update_rank_status(self, rank_status_dict: dict) -> list:
        """
        update_rank_status updates the single status value of all current ranks based on
        the return value of the monitor.
        """
        all_healthy = True
        all_succeed = True
        rank_pids = []
        local_rank_ids = []
        fault_ranks = []
        for key, details in rank_status_dict.items():
            # if process is in ok, not start yet[msrun taken over by taskd, monitor maybe called before training],
            # sleeping[during process recover]
            if details[constants.RANK_STATUS_KEY] not in {constants.RANK_STATUS_OK, constants.RANK_STATUS_NOT_START,
                                                          constants.RANK_STATUS_COMPLETE}:
                self.rank_status = self.RANK_STATUS_UNHEALTHY
                fault_ranks.append(int(key))
                all_healthy = False
            if details[constants.RANK_STATUS_KEY] not in {constants.RANK_STATUS_COMPLETE}:
                all_succeed = False
            rank_pids.append(details[constants.RANK_PID_KEY])
            local_rank_ids.append(details[constants.GLOBAL_RANK_ID_KEY])
        self.rank_pids = rank_pids
        self.node_global_rank_ids = local_rank_ids
        self.all_rank_succeed = all_succeed
        if all_healthy:
            self.rank_status = self.RANK_STATUS_HEALTHY
        return fault_ranks

    def report_fault_rank(self, fault_ranks: list):
        if not self.check_new_fault(fault_ranks):
            run_log.info(f'no additional fault process, fault_rank: {fault_ranks}')
            return
        report_info = AgentReportInfo(fault_ranks=fault_ranks)
        self.send_message_to_manager('STATUS', REPORT_CODE, report_info)
        self.local_fault_rank = fault_ranks
        run_log.info(f'New fault process detected, fault_rank: {fault_ranks}')
        return


    def initialize_workers(self, msg):
        run_log.info(f'receive {msg.msg_type} command, start to initialize workers')
        self._func_map.get('START_ALL_WORKER')()

    def stop_workers(self, msg):
        run_log.info(f'receive {msg.msg_type} command, start to stop workers')
        self._func_map.get('KILL_WORKER')([constants.KILL_ALL_WORKERS])

    def exit_agent(self, msg):
        run_log.info(f'receive {msg.msg_type} command, start to exit agent')
        self._func_map.get('KILL_WORKER')([constants.KILL_ALL_WORKERS])
        self.send_message_to_manager('STATUS', REPORT_CODE, AgentReportInfo())
        exit(1)

    def restart_workers(self, msg):
        run_log.info(f'receive {msg.msg_type} command, start to restart workers')
        self._func_map.get('KILL_WORKER')([constants.KILL_ALL_WORKERS])
        self._func_map.get('START_ALL_WORKER')()
