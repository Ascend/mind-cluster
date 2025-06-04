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
from taskd.python.framework.agent.pt_agent.pt_agent import PtAgent
from taskd.python.framework.agent.ms_agent.ms_agent import MsAgent
from taskd.python.framework.common.type import CONFIG_SERVERRANK_KEY, Position, NetworkConfig, LOCAL_HOST,\
     DEFAULT_AGENT_ROLE, DEFAULT_SERVERRANK, DEFAULT_PROCESSRANK, CONFIG_UPSTREAMIP_KEY,\
     CONFIG_UPSTREAMPORT_KEY, CONFIG_FRAMEWORK_KEY, DEFAULT_AGENT_UPSTREAMPORT


taskd_agent = None
framework = None


def init_taskd_agent(config : dict = {}, cls = None) -> bool:
    global taskd_agent, framework
    if cython_api.lib is None:
        run_log.error("init_taskd_agent: the libtaskd.so has not been loaded!")
        return False
    framework = config.get(CONFIG_FRAMEWORK_KEY)
    if framework == "PyTorch" and cls is not None:
        default_rank = os.getenv("RANK", DEFAULT_SERVERRANK)
    if framework == "MindSpore":
        default_rank = os.getenv("MS_NODE_RANK", DEFAULT_SERVERRANK)

    default_values = {
        CONFIG_UPSTREAMIP_KEY: LOCAL_HOST,
        CONFIG_UPSTREAMPORT_KEY: DEFAULT_AGENT_UPSTREAMPORT,
        CONFIG_SERVERRANK_KEY: default_rank
    }

    config_values = {}
    for key, default in default_values.items():
        config_values[key] = config.get(key, default)
    network_config = NetworkConfig(
            pos=Position(
                role = DEFAULT_AGENT_ROLE,
                server_rank = config_values[CONFIG_SERVERRANK_KEY],
                process_rank = DEFAULT_PROCESSRANK
            ),
            upstream_addr = config_values[CONFIG_UPSTREAMIP_KEY] + ":" + config_values[CONFIG_UPSTREAMPORT_KEY],
            listen_addr = '',
            enable_tls = False,
            tls_conf = None
        )
    log_name = "agent-" + config_values[CONFIG_SERVERRANK_KEY] + ".log"
    init_taskd_log_func = cython_api.lib.InitTaskdLog
    result = init_taskd_log_func(log_name)
    if result != 0:
        run_log.error("init_taskd_agent: init_taskd_log failed")
        return False
    run_log.info(f"init_taskd_agent: network configs is {network_config}")
    if framework == "PyTorch" and cls is not None:
        taskd_agent = PtAgent(cls, network_config)
    if framework == "MindSpore":
        taskd_agent = MsAgent(network_config)
    return True


def start_taskd_agent():
    if taskd_agent is None:
        run_log.error("taskd_agent is None")
        return None
    if framework == "PyTorch":
        return taskd_agent.invoke_run("DEFAULT_ROLE")
    if framework == "MindSpore":
        return taskd_agent.start()
    return None


def register_func(operator, func) -> bool:
    if taskd_agent is None:
        run_log.error("taskd_agent is None")
        return False
    taskd_agent.register_callbacks(operator, func)
    return True
