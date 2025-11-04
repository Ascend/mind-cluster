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
from taskd.python.cython_api import cython_api
from taskd.python.utils.log import run_log
from taskd.python.constants.constants import CHECK_STEP_PERIOD, JOB_ID_KEY, DEFAULT_GROUP_DIR, \
    PROFILING_DIR_MODE, GROUP_INFO_NAME, GROUP_INFO_KEY, GROUP_NAME_KEY, GROUP_RANK_KEY, \
        GLOBAL_RANKS_KEY, DEFAULT_GROUP, GROUP_BASE_DIR_ENV
import threading
import os
import json
import time


def get_save_path(rank) -> str:
    job_id = os.getenv(JOB_ID_KEY)
    if job_id is None or job_id == "":
        run_log.error(f"job id is invalid")
        return ""
    base_dir = os.getenv(GROUP_BASE_DIR_ENV)
    if base_dir is None:
        base_dir = ""
    if not os.path.exists(base_dir):
        run_log.warning(f"config group base dir {base_dir} not exists, use default group info dir")
        base_dir = DEFAULT_GROUP_DIR
    rank_path = os.path.join(base_dir, job_id, str(rank))
    if os.path.islink(rank_path):
        run_log.error(f"rank path {rank_path} is symlink")
        return ""
    try:
        os.makedirs(rank_path, mode=PROFILING_DIR_MODE, exist_ok=True)
    except FileExistsError:
        run_log.warning(f"filepath={rank_path} exist")
        return rank_path
    except OSError as err:
        run_log.error(f"filepath={rank_path} failed, err={err}")
        return ""
    return rank_path


def get_group_info(rank: int) -> dict:
    try:
        import torch
        from torch.distributed.distributed_c10d import _world as distributed_world
        if not torch.distributed.is_available() or not torch.distributed.is_initialized():
            run_log.error(f'distributed is not available or not initialized, rank={rank}')
            return {}
        group_info = {}
        global_rank = rank
        distributed_world.pg_names
        for group, group_config in distributed_world.pg_map.items():
            run_log.debug(f'distributed world data: {group}, {group_config}')
            if len(group_config) < 1:
                run_log.warning(f'group config is invalid, group={group}, group_config={group_config}')
                continue
            backend = str(group_config[0]).lower()
            if backend != "hccl":
                continue
            hccl_group = group._get_backend(torch.device("npu"))
            comm_name = hccl_group.get_hccl_comm_name(global_rank, init_comm=False)
            if comm_name is not None:
                group_info[comm_name] = {
                    GROUP_NAME_KEY: hccl_group.options.hccl_config.get("group_name", ""),
                    GROUP_RANK_KEY: torch.distributed.get_group_rank(group, global_rank),
                    GLOBAL_RANKS_KEY: torch.distributed.get_process_group_ranks(group)
                }
        default_group = torch.distributed.distributed_c10d._get_default_group()
        comm_name = default_group._get_backend(torch.device("npu")).get_hccl_comm_name(global_rank, init_comm=False)
        if comm_name is not None:
            group_info[comm_name] = {
                GROUP_NAME_KEY: DEFAULT_GROUP,
                GROUP_RANK_KEY: torch.distributed.get_group_rank(default_group, global_rank),
                GLOBAL_RANKS_KEY: torch.distributed.get_process_group_ranks(default_group)
            }
        return group_info
    except Exception as err:
        run_log.error(f'get group info failed, err={err}')
        return {}


def dump_group_info():
    try:
        import torch
        rank = torch.distributed.get_rank()
        run_log.info(f'start dump group info for rank={rank}')
        group_info = get_group_info(rank)
        if group_info is not None:
            run_log.info(f'get group info: {group_info}')
            save_path = get_save_path(rank)
            if save_path == "":
                run_log.error(f'get save path for group info failed')
                return
            run_log.info(f'save group info to: {save_path}')
            full_path = os.path.join(save_path, GROUP_INFO_NAME)
            if os.path.islink(full_path):
                run_log.error(f'dump path {save_path} is symlink, skip dump')
                return
            with open(full_path, "w", encoding="utf-8") as f:
                json.dump(group_info, f, ensure_ascii=False, indent=4)
    except Exception as err:
        run_log.error(f'save group info failed: {err}')