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
import json
import queue

from taskd.python.framework.agent.ms_mgr.ms_utils import check_monitor_res_valid, calculate_global_rank, \
    calculate_local_rank_by_global_rank
from taskd.python.toolkit.constants import constants
from taskd.python.utils.log import run_log
from taskd.python.framework.agent.base_agent.agent_network import init_network_client
from taskd.python.framework.agent.base_agent.base_agent import BaseAgent
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

    def __init__(self, network_config, logger):
        super().__init__()
        self.all_rank_succeed = False
        self.monitor_interval = constants.MONITOR_INTERVAL
        self.node_rank = os.getenv("MS_NODE_RANK")
        self.rank_pids = []
        self.node_global_rank_ids = []
        self.network_config = network_config
        self.command_map = {
            constants.STARTAGENTCODE: self.initialize_workers,
            constants.STOPWORKERSCODE: self.stop_workers,
            constants.EXITAGENTCODE: self.exit_agent,
            constants.RESTARTAGENTCODE: self.restart_workers,
            'GRACE_EXIT': self.grace_exit,
            constants.RESTARTWORKERCODE: self.recover_in_place,
        }
        self.logger = logger
        self.rank_status = ''
        self.local_rank_to_pid = {}

    def start_worker(self, start_worker_func):
        time_use = 0
        run_log.info(f"agent {self.node_rank} start worker")
        self.send_message_to_manager('STATUS', constants.RESTARTTIMESCODE, "1")
        while True:
            try:
                item = self.msg_queue.get_nowait()
                if item.code == constants.STARTAGENTCODE:
                    start_worker_func()
                    break
            except queue.Empty:
                run_log.debug('msg_queue is empty')
            time.sleep(1)
            if time_use > constants.INIT_TIMEOUT:
                raise RuntimeError("start_worker timeout")

    def start(self):
        kill_worker_func = self._func_map.get(constants.KILL_ALL_WORKER_CALLBACK_NAME)
        start_worker_func = self._func_map.get(constants.START_ALL_WORKER_CALLBACK_NAME)
        monitor_func = self._func_map.get(constants.MONITOR_CALLBACK_NAME)
        start_single_worker_func = self._func_map.get(constants.START_WORKER_LIST_CALLBACK_NAME)
        if (kill_worker_func is None or start_worker_func is None or monitor_func is None or
                start_single_worker_func is None):
            raise Exception(f"{self.FRAMEWORK_MS_NAME} hasn't fully registered all callbacks")
        self.node_global_rank_ids = calculate_global_rank()
        init_network_client(self.network_config, self.msg_queue, self.logger)
        self.check_network()
        self.start_worker(start_worker_func)

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
            if self.rank_status == self.RANK_STATUS_UNHEALTHY:
                run_log.info(f"status unhealthy, {ms_proc_status}")
                self.report_fault_rank(fault_ranks)
            self.handle_all_process_succeed()

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
        local_rank_to_pid = {}
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
            local_rank_to_pid[details[constants.GLOBAL_RANK_ID_KEY]] = details[constants.RANK_PID_KEY]
        self.rank_pids = rank_pids
        self.node_global_rank_ids = local_rank_ids
        self.all_rank_succeed = all_succeed
        self.local_rank_to_pid = local_rank_to_pid
        if all_healthy:
            self.rank_status = self.RANK_STATUS_HEALTHY
        return fault_ranks

    def report_fault_rank(self, fault_ranks: list):
        if not self.check_new_fault(fault_ranks):
            run_log.info(f'no additional fault process, fault_rank: {fault_ranks}')
            return
        report_info = AgentReportInfo(fault_ranks=fault_ranks)
        self.send_message_to_manager('STATUS', constants.REPORT_CODE, report_info, {"REPORT_FAULT_TIME": str(int(time.time()))})
        self.local_fault_rank = fault_ranks
        run_log.info(f'New fault process detected, fault_rank: {fault_ranks}')
        return

    def initialize_workers(self, msg):
        run_log.info(f'receive {msg.code} command, start to initialize workers')
        self._func_map.get('START_ALL_WORKER')()

    def stop_workers(self, msg):
        run_log.info(f'receive {msg.code} command, start to stop workers')
        try:
            fault_ranks = json.loads(msg.message)
            if fault_ranks is None:
                run_log.error("fault_ranks is None")
                return
            int_fault_ranks = [int(rank) for rank in fault_ranks]
        except Exception as e:
            run_log.error(f"Convert fault_ranks to int failed: {e}")
            return
        run_log.info(f"message fault_ranks is {int_fault_ranks}")
        local_fault_ranks = self.get_fault_local_ranks(int_fault_ranks)
        run_log.info(f"local fault_ranks is {local_fault_ranks}")
        fault_pid_list = self.get_fault_pids(local_fault_ranks)
        if len(fault_pid_list) > 0:
            run_log.info(f"nodeRank:{self.node_rank} stop workers, pid:{fault_pid_list}")
            self._func_map.get(constants.KILL_ALL_WORKER_CALLBACK_NAME)(fault_pid_list)

    def exit_agent(self, msg):
        run_log.info(f'receive {msg.code} command, start to exit agent')
        self._func_map.get('KILL_WORKER')([constants.KILL_ALL_WORKERS])
        self.send_message_to_manager('STATUS', constants.REPORT_CODE, AgentReportInfo())
        exit(1)

    def restart_workers(self, msg):
        run_log.info(f'receive {msg.code} command, start to restart workers')
        self._func_map.get('KILL_WORKER')([constants.KILL_ALL_WORKERS])
        run_log.warning(f"nodeRank:{self.node_rank}"
                        f"will sleep for {constants.KILL_INTERVAL} secs, after kill workers to restart")
        time.sleep(constants.KILL_INTERVAL)
        self._func_map.get('START_ALL_WORKER')()
        self.local_fault_rank = []
        
    def recover_in_place(self, msg):
        run_log.info(f'receive {msg.code} command, start to recover in place')
        fault_ranks = json.loads(msg.message)
        if fault_ranks is None:
            run_log.error("fault_ranks is None")
            return
        try:
            int_fault_ranks = [int(rank) for rank in fault_ranks]
        except ValueError as e:
            run_log.error(f"Convert fault_ranks to int failed: {e}")
            return
        run_log.info(f"message fault_ranks is {int_fault_ranks}")
        local_fault_ranks = self.get_fault_local_ranks(int_fault_ranks)
        run_log.info(f"local fault_ranks is {local_fault_ranks}")
        fault_pid_list = self.get_fault_pids(local_fault_ranks)
        if len(fault_pid_list) > 0:
            run_log.info(f"nodeRank:{self.node_rank} restart part workers, pid:{fault_pid_list}")
            restart_local_rank = calculate_local_rank_by_global_rank(local_fault_ranks)
            self._func_map.get(constants.KILL_ALL_WORKER_CALLBACK_NAME)(fault_pid_list)
            time.sleep(constants.RELEASE_INTERVAL) # wait for device release resources
            self._func_map.get(constants.START_WORKER_LIST_CALLBACK_NAME)(restart_local_rank)
            self.local_fault_rank = []

    def get_fault_pids(self, local_ranks):
        pid_list = []
        for local_rank in local_ranks:
            if local_rank in self.local_rank_to_pid:
                pid = self.local_rank_to_pid.get(local_rank)
                pid_list.append(pid)
        return pid_list

    def get_fault_local_ranks(self, fault_ranks):
        fault_local_ranks = []
        for fault_rank in fault_ranks:
            if fault_rank in self.node_global_rank_ids:
                fault_local_ranks.append(fault_rank)
        return fault_local_ranks

    def handle_all_process_succeed(self):
        if not self.all_rank_succeed:
            return
        run_log.info(
            f"nodeRank:{self.node_rank} successfully finished."
        )
        # wait for MS release resources
        time.sleep(constants.WAITING_INTERVAL * constants.WAIT_TIMES)
        stop_res = self._func_map.get(constants.KILL_ALL_WORKER_CALLBACK_NAME)([constants.KILL_ALL_WORKERS])
        run_log.info(f"rank with pid {self.rank_pids} will be cleared")
        if stop_res != constants.RES_OK:
            run_log.error(f"nodeRank:{self.node_rank} failed to stop workers with return code:{stop_res}")
        exit(0)
