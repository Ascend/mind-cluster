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
from taskd.python.constants.constants import PYTORCH, CHECK_STEP_PERIOD, JOB_ID_KEY, PROFILING_BASE_DIR, \
    PROFILING_DIR_MODE, GROUP_INFO_NAME, GROUP_INFO_KEY, GROUP_NAME_KEY, GROUP_RANK_KEY, \
        GLOBAL_RANKS_KEY, DEFAULT_GROUP
import threading, os, json, time

class WorkerConfig:
    """
    framework: AI framework. such as torch or ms
    rank_id: the global rank id of the process, this should be called after rank is initialized
    upper_limit_of_disk_in_mb: the limit of profiling file of all jobs
    """
    def __init__(self, framework: str, rank: int, disk_limit: int = 5000):
        self.framework = framework
        self.rank_id = rank
        self.upper_limit_of_disk_in_mb = disk_limit

class Worker:
    """
    Worker is a framework of training thread management
    """

    def __init__(self, rank: int):
        self.rank = rank
        self.framework = PYTORCH
        self.upper_limit_of_disk_in_mb = 5000

    def start(self) -> bool:
        return self._start_up_monitor()

    def init_monitor(self, config: WorkerConfig) -> bool:
        self.rank = config.rank_id
        self.framework = config.framework
        self.upper_limit_of_disk_in_mb = config.upper_limit_of_disk_in_mb
        if cython_api.lib is None:
            run_log.error("the libtaskd.so has not been loaded!")
            return False
        init_taskd_func = cython_api.lib.InitTaskMonitor
        result = init_taskd_func(self.rank, self.upper_limit_of_disk_in_mb)
        if result == 0:
            run_log.info("Successfully init taskd monitor")
            return True
        run_log.warning(f"failed to init taskd monitor with ret code:f{result}")
        return False

    def _start_up_monitor(self) -> bool:
        try:
            if cython_api.lib is None:
                run_log.error("the libtaskd.so has not been loaded!")
                return False
            start_monitor_client_func = cython_api.lib.StartMonitorClient
            result = start_monitor_client_func()
            if result == 0:
                run_log.info(f"Successfully start monitor client for rank:{self.rank}")
                thread = threading.Thread(target=save_group_info, args=(self.framework, self.rank))
                thread.daemon = True
                thread.start()
                return True
            run_log.warning(f"failed to start up monitor client with ret code:f{result}")
            return False
        except Exception as e:
            run_log.error(f"failed to start up monitro client, e:{e}")
            return False


def get_save_path(rank) -> str:
    job_id = os.getenv(JOB_ID_KEY)
    if not job_id:
        run_log.error(f"job id is invalid")
        return 
    rank_path = os.path.join(PROFILING_BASE_DIR, job_id, str(rank))
    try:
        os.makedirs(rank_path, mode=PROFILING_DIR_MODE, exist_ok=True)
    except FileExistsError:
        run_log.warn(f"filepath={rank_path} exist")
        return rank_path
    except OSError as err:
        run_log.error(f"filepath={rank_path} failed, err={err}")
        return ""
    return rank_path

def save_group_info(framework: str, rank: int):
    if framework != PYTORCH:
        run_log.warn(f'framework={framework} not support save group info')
        return
    check_step_out = cython_api.lib.StepOut
    try:    
        while check_step_out() != 1:
            run_log.warn(f'not ready to write group info, try it after a few seconds')
            time.sleep(CHECK_STEP_PERIOD)
        run_log.info(f'start dump group info for rank={rank}')
        import torch
        from torch.distributed.distributed_c10d import _world as distributed_world
        if not torch.distributed.is_available() or not torch.distributed.is_initialized():
            run_log.error(f'distributed is not available or not initialized, rank={rank}')
            return
        group_info = {}
        global_rank = torch.distributed.get_rank()
        for group, group_config in distributed_world.pg_map.items():
            run_log.info(f'distributed world data: {group}, {group_config}')
            if len(group_config) < 1:
                continue
            backend = str(group_config[0]).lower()
            if backend != "hccl":
                continue
            hccl_group = group._get_backend(torch.device("npu"))
            comm_name = hccl_group.get_hccl_comm_name(global_rank, init_comm=False)
            if comm_name:
                group_info[comm_name] = {
                    GROUP_NAME_KEY: hccl_group.options.hccl_config.get("group_name", ""),
                    GROUP_RANK_KEY: torch.distributed.get_group_rank(group, global_rank),
                    GLOBAL_RANKS_KEY: torch.distributed.get_process_group_ranks(group)
                }
        default_group = torch.distributed.distributed_c10d._get_default_group()
        comm_name = default_group._get_backend(torch.device("npu")).get_hccl_comm_name(global_rank, init_comm=False)
        if comm_name:
            group_info[comm_name] = {
                GROUP_NAME_KEY: DEFAULT_GROUP,
                GROUP_RANK_KEY: torch.distributed.get_group_rank(default_group, global_rank),
                GLOBAL_RANKS_KEY: torch.distributed.get_process_group_ranks(default_group)
            }
        if group_info:
            data = {GROUP_INFO_KEY: group_info}
            run_log.info(f'get group info: {data}')
            save_path = get_save_path(rank)
            if save_path == "":
                run_log.error(f'get save path for group info failed')
                return
            full_path = os.path.join(save_path, GROUP_INFO_NAME)
            with open(full_path, "w", encoding="utf-8") as f:
                json.dump(data, f, ensure_ascii=False, indent=4)
    except Exception as err:
        run_log.error(f'save group info failed: {err}')