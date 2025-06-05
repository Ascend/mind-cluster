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
from taskd.python.toolkit.constants.constants import RANK_PID_KEY
from taskd.python.utils.log import run_log


# check_monitor_res_valid to check whether mindspore monitor interface given a valid result
def check_monitor_res_valid(rank_status_dict: dict):
    if not isinstance(rank_status_dict, dict):
        run_log.warning("monitor result should be a dict")
        return False

    for rank, info in rank_status_dict.items():
        if not isinstance(info, dict):
            run_log.warning(f"monitor result for rank {rank} should be a dict")
            return False

        # to check every dict has key" 'pid', 'status', 'global_rank'
        required_keys = [constants.RANK_PID_KEY, constants.RANK_STATUS_KEY, constants.GLOBAL_RANK_ID_KEY]
        for key in required_keys:
            if key not in info:
                run_log.warning(f"{rank} has no key: {key}")
                return False

        if not isinstance(info['pid'], int):
            run_log.warning(f"info['pid']is not int, but {info['pid']}")
            return False

        if not isinstance(info['status'], int) and info['status'] is not None:
            run_log.warning(f"info['status']is not int,but {info['status']}")
            return False

        if not isinstance(info[constants.GLOBAL_RANK_ID_KEY], int):
            run_log.warning(f"info['global_rank']is not int, but {info[constants.GLOBAL_RANK_ID_KEY]}")
            return False
    return True


def calculate_global_rank():
    # calculate global ranks from env MS_LOCAL_WORKER and MS_NODE_RANK
    ms_local_worker = os.getenv('MS_LOCAL_WORKER')
    ms_node_rank = os.getenv('MS_NODE_RANK')
    if ms_local_worker is None or ms_node_rank is None:
        run_log.error("the env variable MS_LOCAL_WORKER or MS_NODE_RANK is not set")
        return []
    try:
        ms_local_worker = int(ms_local_worker)
        ms_node_rank = int(ms_node_rank)
    except ValueError as e:
        run_log.info(f"failed to get MS_LOCAL_WORKER and MS_NODE_RANK from env, please set it: {e}")
        return []
    global_rank = []
    for local_worker in range(ms_local_worker):
        global_rank.append(ms_node_rank * ms_local_worker + local_worker)
    return global_rank


def calculate_local_rank_by_global_rank(global_rank_list: list):
    ms_local_worker = os.getenv('MS_LOCAL_WORKER')
    if ms_local_worker is None :
        run_log.error("the env variable MS_LOCAL_WORKER is not set")
        return None
    try:
        ms_local_worker = int(ms_local_worker)
    except ValueError as e:
        run_log.info(f"failed to get MS_LOCAL_WORKER from env, please set it: {e}")
        return None
    local_rank_list = []
    for global_rank in global_rank_list:
        local_rank_list.append(global_rank % ms_local_worker)
    return local_rank_list