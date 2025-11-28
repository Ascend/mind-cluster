#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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
import os
from mindio_acp.common.mindio_logger import LOGGER
from mindio_acp.launch_server_conf.launch_server_param import server_worker_dir

SOCKET_PATH = f'{server_worker_dir}/uds/mindio_memfs_123.s'


def check_process(print_log=True):
    try:
        if os.access(SOCKET_PATH, os.R_OK | os.W_OK):
            return True
        else:
            if print_log:
                LOGGER.error('[mindio_acp] ockiod service not available.')
    except Exception as e:
        if print_log:
            LOGGER.error('[mindio_acp] check ockiod process failed. Exception is:%s', str(e))
    return False
