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

import taskd
from taskd.python.cython_api import cython_api
from taskd.python.utils.log import run_log


class Manager:
    """
    Manager is a framework of task management
    """

    def init_taskd_manager(self, config: dict) -> bool:
        if cython_api.lib is None:
            run_log.error("the libtaskd.so has not been loaded")
            return False
        config_str = json.dumps(config).encode('utf-8')
        init_taskd_manager_func = cython_api.lib.InitTaskdManager
        if init_taskd_manager_func is None:
            run_log.error("init_taskd_manager: func InitTaskdManager has not been loaded from libtaskd.so")
            return False
        result = init_taskd_manager_func(config_str)
        if result == 0:
            run_log.info("successfully init taskd manager")
            return True
        run_log.warning(f"failed to init taskd manager with ret code:f{result}")
        return False

    def start_taskd_manager(self) -> bool:
        try:
            if cython_api.lib is None:
                run_log.error("the libtaskd.so has not been loaded")
                return False
            start_taskd_manager_func = cython_api.lib.StartTaskdManager
            if start_taskd_manager_func is None:
                run_log.error("start_taskd_manager: func StartTaskdManager has not been loaded from libtaskd.so")
                return False
            result = start_taskd_manager_func()
            if result == 0:
                run_log.info(f"successfully start taskd manager")
                return True
            run_log.warning(f"failed to start taskd manager with ret code:f{result}")
            return False
        except Exception as e:
            run_log.error(f"failed to start manager, error:{e}")
            return False
