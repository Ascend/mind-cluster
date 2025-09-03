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
import ctypes

import taskd
import threading
from taskd.python.cython_api import cython_api
from taskd.python.utils.log import run_log
from taskd.python.toolkit.constants import constants


class Manager:
    """
    Manager is a framework of task management
    """
    def __init__(self):
        self.callback = None
        self.c_callback = None
        self.config = {}

    def init_taskd_manager(self, config: dict) -> bool:
        self.config = config
        if os.getenv(constants.PROCESS_RECOVER) == "on":
            config[constants.FAULT_RECOVER] = "on"
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
            self.start_controller()
            result = start_taskd_manager_func()
            if result == 0:
                run_log.info(f"successfully start taskd manager")
                return True
            run_log.warning(f"failed to start taskd manager with ret code:f{result}")
            return False
        except Exception as e:
            run_log.error(f"failed to start manager, error:{e}")
            return False
        
    def start_controller(self):
        try:
            from taskd.python.framework.manager.controller import init_controller, backend_send_callback
            self.callback = backend_send_callback
            c_callback = ctypes.CFUNCTYPE(ctypes.c_int, ctypes.c_char_p)
            self.c_callback = c_callback(self.callback)
            start_mindio_controller = threading.Thread(target=init_controller)
            start_mindio_controller.daemon = True
            start_mindio_controller.start()
            if cython_api.lib:
                register_func = cython_api.lib.RegisterBackendCallback
                register_func(self.c_callback)
                run_log.info("Successfully register controller callback")
        except Exception as e:
            run_log.error(f"register switch controller failed, error: {e}")
            