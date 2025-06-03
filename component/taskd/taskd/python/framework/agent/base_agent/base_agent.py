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
import time
import queue
import uuid

from dataclasses import asdict
from taskd.python.utils.log import run_log
from taskd.python.toolkit.fault_checker.fault_check import grace_exit_pids, stop_pids
from taskd.python.framework.agent.base_agent.agent_network import get_message_manager, network_send_message
from taskd.python.framework.common.type import MsgBody, MessageInfo


DEFAULT_DST = {
    "Role": "Mgr",
    "ServerRank": "0",
    "ProcessRank": "-1"
}


REPORT_CODE = 601
DEFAULT_MSG_TYPE = "DEFAULT"
STATUS_MSG_TYPE = "STATUS"
        
    
class BaseAgent:
    """
    BaseAgent manages the lifecycle of the training process,
    including monitoring, starting, and stopping the training process.
    At the same time, respond to the manager's message instructions to complete fault recovery.
    """
    def __init__(self):
        self.node_rank = -1
        self.local_world_size = 0
        self.local_rank = []
        self.pids = {}
        self.local_fault_rank = []
        self._func_map = {}
        self.command_map = {}
        self.network_config = None
        self.msg_queue = queue.Queue()

    def register_callbacks(self, operator, func):
        self._func_map[operator] = func

    def check_new_fault(self, fault_ranks: list) -> bool:
        return not sorted(self.local_fault_rank) == sorted(fault_ranks)


    def initialize_workers(self, msg):
        raise NotImplementedError

    def stop_workers(self, msg):
        raise NotImplementedError

    def exit_agent(self, msg):
        raise NotImplementedError

    def restart_workers(self, msg):
        raise NotImplementedError

    def handle_message(self):
        try:
            item = self.msg_queue.get_nowait()
        except queue.Empty:
            run_log.debug('msg_queue is empty')
            return
        self.command_map.get(item.MsgType)(item)

    def grace_exit(self, msg):
        run_log.info(f'receive {item.MsgType} command, start to grace exit workers')
        try:
            grace_exit_pids(self.pids)
        except Exception as e:
            run_log.error(f'grace_exit encountered an exception: {e}')
        finally:
            stop_pids(self.pids)
            
    
    def send_message_to_manager(self, command, code, report_info):
        report_json = json.dumps(asdict(report_info))
        msg_body = MsgBody(
            msg_type=command,
            code=code,
            message=report_json,
            extension={},
        )
        body_json = json.dumps(asdict(msg_body))
        msg_info = MessageInfo(
            uuid=str(uuid.uuid4()),
            biz_type=DEFAULT_MSG_TYPE,
            dst=DEFAULT_DST,
            body=body_json
        )
        network_send_message(msg_info)
            
    def check_network(self):
        time_cost = 0
        while True:
            if time_cost > 60:
                run_log.error('waiting for message manager timeout')
                raise ValueError("failed to initialized agent network, initialization message_manager timeout")
            if get_message_manager() is None:
                run_log.info('waiting for message manager')
                time.sleep(1)
                time_cost += 1
                continue
            run_log.info('message manager is ready')
            break
