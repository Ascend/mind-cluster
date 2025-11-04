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
import ctypes
import os
from taskd.python.utils.log import run_log
from taskd.python.constants.constants import LIB_SO_NAME, LIB_SO_PATH


def load_lib_taskd():
    try:
        mode = os.RTLD_LAZY | os.RTLD_LOCAL
        lib_path = os.path.join(os.path.dirname(__file__), LIB_SO_PATH, LIB_SO_NAME)
        if os.path.islink(lib_path):
            run_log.error(f"{LIB_SO_NAME} is symlink")
            return None
        taskd_lib = ctypes.CDLL(lib_path, mode=mode)
        run_log.info(f"{LIB_SO_NAME} loaded successfully")
        return taskd_lib
    except Exception as e:
        run_log.error(f"{LIB_SO_NAME} loaded failed: {e}")
        return None


lib = load_lib_taskd()