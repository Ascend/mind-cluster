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

from taskd.python.toolkit.constants import constants
from taskd.python.utils.log import run_log


def get_hccl_switch_nic_timeout():
    timeout = os.getenv(constants.HCCL_CONNECT_TIMEOUT)
    if timeout is None:
        return constants.SWITCH_NIC_DEFAULT_TIMEOUT
    try:
        timeout = int(timeout)
        if timeout <= 0:
            run_log.warning("HCCL_CONNECT_TIMEOUT is invalid")
            return constants.SWITCH_NIC_DEFAULT_TIMEOUT
        return min(int(timeout * 2.5), constants.SWITCH_NIC_MAX_TIMEOUT)
    except Exception as err:
        run_log.warning(f"get HCCL_CONNECT_TIMEOUT failed, {err}")
        return constants.SWITCH_NIC_DEFAULT_TIMEOUT