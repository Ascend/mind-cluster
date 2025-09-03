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
import time
import queue

from taskd.python.utils.log import run_log
from taskd.python.framework.agent.base_agent.agent_network import init_network_client
from taskd.python.framework.agent.base_agent.base_agent import BaseAgent
from taskd.python.framework.common.type import AgentReportInfo
from taskd.python.toolkit.constants import constants
try:
    from torch.distributed.elastic.agent.server.api import WorkerState, RunResult
except ImportError:
    run_log.debug("torch not installed, please install torch to use pt_agent")
    WorkerState = None
    RunResult = None


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
            constants.STOPAGENTCODE: self.stop_workers,
            constants.EXITAGENTCODE: self.exit_agent,
            constants.RESTARTAGENTCODE: self.restart_workers,
            constants.GRACEEXITAGENTCODE: self.grace_exit,
        }
        self.logger = logger

    def invoke_run(self, role) -> RunResult:
        init_network_client(self.network_config, self.msg_queue, self.logger)
        self.check_network()
        spec = self.worker_group.spec
        role = spec.role
        run_log.info("[%s] starting workers for entrypoint: %s", role, spec.get_entrypoint_name())
        self._func_map.get('START_ALL_WORKER')(self.worker_group)
        self.update_agent_info()
        monitor_interval = spec.monitor_interval

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
        self.send_message_to_manager('STATUS', constants.REPORT_CODE, report_info)
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
        self._func_map.get('KILL_WORKER')(self.worker_group)
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
