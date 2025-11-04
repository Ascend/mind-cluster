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

from taskd.python.framework.worker.worker import Worker
from taskd.python.utils.log import run_log

taskd_worker = None


def init_taskd_worker(rank_id: int, upper_limit_of_disk_in_mb: int = 5000, framework: str = "pt") -> bool:
    """
    init_taskd_worker: to init taskd worker
    rank_id: the global rank id of the process, this should be called after rank is initialized
    upper_limit_of_disk_in_mb: the limit of profiling file of all jobs
    """
    global taskd_worker
    try:
        if not isinstance(rank_id, int) or not isinstance(upper_limit_of_disk_in_mb, int):
            run_log.error(f"rank_id {rank_id} and upper_limit_of_disk_in_mb {upper_limit_of_disk_in_mb} "
                          f"should be integers")
            return False
        if rank_id < 0:
            run_log.error(f"rank_id {rank_id} should not less than 0")
            return False
        if upper_limit_of_disk_in_mb < 0:
            run_log.error(f"upper_limit_of_disk_in_mb {upper_limit_of_disk_in_mb} should not less than 0")
            return False
        taskd_worker = Worker(rank_id, framework)
        return taskd_worker.init_worker(upper_limit_of_disk_in_mb)
    except Exception as e:
        run_log.error(f"Failed to initialize worker: {e}")
        return False


def start_taskd_worker() -> bool:
    """
    Starts the taskd worker
    """
    if taskd_worker is None:
        # if worker has not been initialized
        run_log.error("Worker is not initialized. Please call init_worker first.")
        return False
    try:
        return taskd_worker.start()
    except Exception as e:
        run_log.error(f"Failed to start worker: {e}")
        return False

def destroy_taskd_worker() -> bool:
    """
    Destroys the taskd worker
    """
    if taskd_worker is None:
        # if worker has not been initialized
        run_log.error("Worker is not initialized. Please call init_worker first.")
        return False
    try:
        return taskd_worker.destroy()
    except Exception as e:
        run_log.error(f"Failed to destroy worker: {e}")
        return False