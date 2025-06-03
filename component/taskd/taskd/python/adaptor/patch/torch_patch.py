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
import threading
import signal

import torch.distributed.elastic.agent.server.api
from torch.distributed.elastic.agent.server.api import DEFAULT_ROLE, RunResult
from taskd.python.utils.log import run_log
from taskd.api.taskd_proxy_api import init_taskd_proxy
from taskd.api.taskd_agent_api import init_taskd_agent, start_taskd_agent, register_func
from taskd.python.toolkit.constants.constants import SLEEP_GAP
from taskd.python.framework.common.type import CONFIG_UPSTREAMIP_KEY, LOCAL_HOST, CONFIG_FRAMEWORK_KEY


def patch_default_signal():
    time.sleep(SLEEP_GAP)
    return signal.SIGKILL


def patch_invoke_run(self, role: str = DEFAULT_ROLE) -> RunResult:
    proxy = threading.Thread(target=init_taskd_proxy, args=({CONFIG_UPSTREAMIP_KEY : os.getenv("MASTER_ADDR", LOCAL_HOST)},))
    proxy.daemon = True
    proxy.start()
    init_taskd_agent({CONFIG_FRAMEWORK_KEY: 'PyTorch'}, self)
    register_func('KILL_WORKER', self._stop_workers)
    register_func('START_ALL_WORKER', self._initialize_workers)
    register_func('MONITOR', self._monitor_workers)
    register_func('RESTART', self._restart_workers)
    run_log.info("start taskd agent")
    return start_taskd_agent()


def patch_torch_method():
    torch.distributed.elastic.agent.server.api.SimpleElasticAgent._invoke_run = patch_invoke_run
    torch.distributed.elastic.multiprocessing.api._get_default_signal = patch_default_signal
