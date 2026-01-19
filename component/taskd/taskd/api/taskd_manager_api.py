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
from taskd.python.framework.manager.manager import Manager
from taskd.python.utils.log import run_log


taskd_manager = None


def init_taskd_manager(config: dict) -> bool:
    """
    init_taskd_manager: to init taskd manager
    """
    global taskd_manager
    try:
        taskd_manager = Manager()
        return taskd_manager.init_taskd_manager(config)
    except Exception as e:
        run_log.error(f"Failed to initialize manager: {e}")
        return False


def start_taskd_manager() -> bool:
    """
    Starts the taskd manager
    """
    if taskd_manager is None:
        # if manager has not been initialized
        run_log.error("Manager is not initialized. Please call init_manager first.")
        return False
    try:
        return taskd_manager.start_taskd_manager()
    except Exception as e:
        run_log.error(f"Failed to start manager: {e}")
        return False
