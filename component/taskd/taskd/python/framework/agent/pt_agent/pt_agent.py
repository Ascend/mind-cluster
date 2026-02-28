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
import json
import os
import time
import queue
from typing import Dict

from taskd.python.utils.log import run_log
from taskd.python.framework.agent.base_agent.agent_network import init_network_client
from taskd.python.framework.agent.base_agent.base_agent import BaseAgent
from taskd.python.framework.common.type import AgentReportInfo
from taskd.python.toolkit.constants import constants
try:
    from torch.distributed.elastic.agent.server.api import WorkerState, RunResult, WorkerGroup
    from torch.distributed.elastic.multiprocessing import PContext
except ImportError:
    run_log.debug("torch not installed, please install torch to use pt_agent")
    WorkerState = None
    RunResult = None
    PContext = None
    WorkerGroup = None


class PtAgent(BaseAgent):
    """
    PtAgent is for PyTorch to manage training process.
    """
    def __init__(self, cls, network_config, logger):
        super().__init__()
        self.pt_instance = cls
        self.worker_group = cls._worker_group
        self.node_rank = cls._worker_group.spec.rdzv_handler.rank
        self.local_world_size = cls._worker_group.spec.local_world_size
        self.network_config = network_config
        self.command_map = {
            constants.STARTAGENTCODE: self.initialize_workers,
            constants.STOPWORKERSCODE: self.stop_workers,
            constants.EXITAGENTCODE: self.exit_agent,
            constants.RESTARTAGENTCODE: self.restart_workers,
            constants.GRACEEXITAGENTCODE: self.grace_exit,
            constants.RESTARTWORKERCODE: self.recover_in_place,
        }
        self.logger = logger

    def invoke_run(self, role) -> RunResult:
        init_network_client(self.network_config, self.msg_queue, self.logger)
        self.check_network()
        spec = self.worker_group.spec
        role = spec.role
        run_log.info("[%s] starting workers for entrypoint: %s", role, spec.get_entrypoint_name())
        self.start_worker()
        self.update_agent_info()
        monitor_interval = constants.MONITOR_INTERVAL

        while True:
            self.send_message_to_manager('KEEP_ALIVE', 0, AgentReportInfo())
            self.handle_message()
            time.sleep(monitor_interval)
            run_result = self._func_map.get('MONITOR')(self.worker_group)
            state = run_result.state
            self.worker_group.state = state

            if state == WorkerState.SUCCEEDED:
                run_log.info("[%s] worker group successfully finished.", role)
                return run_result
            elif state in {WorkerState.UNHEALTHY, WorkerState.FAILED}:
                self.report_fault_rank(run_result)
                continue
            elif state == WorkerState.HEALTHY:
                continue
            else:
                raise Exception(f"[{role}] Worker group in {state.name} state")

    def update_agent_info(self):
        self.local_rank = [worker.global_rank for worker in self.worker_group.workers]
        for worker in self.worker_group.workers:
            self.pids[worker.global_rank] = worker.id
        self.local_fault_rank = []
        return

    def report_fault_rank(self, result: RunResult):
        fault_ranks = list(result.failures.keys())
        if not self.check_new_fault(fault_ranks):
            run_log.info(f'no additional fault process, fault_rank: {fault_ranks}')
            return
        report_info = AgentReportInfo(fault_ranks=fault_ranks)
        self.send_message_to_manager('STATUS', constants.REPORT_CODE, report_info, {"REPORT_FAULT_TIME": str(int(time.time()))})
        self.local_fault_rank = fault_ranks
        run_log.info(f'New fault process detected, fault_rank: {fault_ranks}')
        return

    def initialize_workers(self, msg):
        run_log.info(f'receive {msg.code} command, restart time is {msg.message},'
                     f' start to initialize workers')
        if int(msg.message) < 0:
            run_log.warning("initialize_workers restart times is negative, exit agent")
            exit(1)
        self.pt_instance._remaining_restarts = int(msg.message)
        self._func_map.get('START_ALL_WORKER')(self.worker_group)

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
        run_log.info(f"fault_ranks is {int_fault_ranks}")
        fault_workers = get_fault_workers(self.pt_instance._worker_group, int_fault_ranks)
        self._func_map.get('KILL_WORKER')(fault_workers)
        self.worker_group.state = WorkerState.STOPPED

    def exit_agent(self, msg):
        run_log.info(f'receive {msg.code} command, start to exit agent')
        self._func_map.get('KILL_WORKER')(self.worker_group)
        self.send_message_to_manager('STATUS', constants.REPORT_CODE, AgentReportInfo())
        exit(1)

    def restart_workers(self, msg):
        run_log.info(f'receive {msg.code} command, start to restart workers, restart time is {msg.message}')
        self.pt_instance._remaining_restarts = int(msg.message)

        self._func_map.get('KILL_WORKER')(self.worker_group)
        if int(msg.message) < 0:
            run_log.warning("restart times is negative, exit agent")
            exit(1)
        self.worker_group.state = WorkerState.STOPPED
        self._func_map.get('START_ALL_WORKER')(self.worker_group)
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
        os.environ[constants.RESTART_FAULT_PROCESS_TYPE_ENV] = "worker"
        run_log.info(f"fault_ranks is {int_fault_ranks}")
        fault_workers = get_fault_workers(self.pt_instance._worker_group, int_fault_ranks)
        self._func_map.get('RESTART')(fault_workers)
        self.local_fault_rank = []

    def start_worker(self):
        time_use = 0
        self.send_message_to_manager('STATUS', constants.RESTARTTIMESCODE,
                                     str(self.pt_instance._remaining_restarts))
        run_log.info(f"agent {self.node_rank} start worker, restart times is {self.pt_instance._remaining_restarts}")
        while True:
            try:
                item = self.msg_queue.get_nowait()
                if item.code == constants.STARTAGENTCODE:
                    self.command_map.get(item.code)(item)
                    break
            except queue.Empty:
                run_log.debug('msg_queue is empty')
            time.sleep(1)
            if time_use > constants.INIT_TIMEOUT:
                raise RuntimeError("start_worker timeout")


def get_pids(p_context_dict: Dict[int, PContext]) -> Dict[int, int]:
    worker_ids: Dict[int, int] = {}
    if len(p_context_dict) <= 0:
        return worker_ids
    for local_rank, pcontext in p_context_dict.items():
        new_dict = {local_rank: v for v in pcontext.pids().values()}
        worker_ids.update(new_dict)
    return worker_ids


def get_fault_workers(wg: WorkerGroup, global_rank: list) -> WorkerGroup:
    new_wg = WorkerGroup(wg.spec)
    new_wg.store = wg.store
    new_wg.workers = {w for w in wg.workers if w.global_rank in global_rank}
    new_wg.state = wg.state
    new_wg.group_rank = wg.group_rank
    new_wg.group_world_size = wg.group_world_size
    return new_wg
