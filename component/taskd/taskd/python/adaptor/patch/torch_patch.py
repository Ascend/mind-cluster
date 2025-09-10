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
import time
import threading
import signal
import shutil

from string import Template
from typing import Any, Dict, Optional, List, Union, Callable
import torch.distributed.elastic.agent.server.api
from torch.distributed.launcher import LaunchConfig, launch_agent
from torch.distributed.elastic.agent.server.api import DEFAULT_ROLE, RunResult, WorkerGroup, WorkerState
from torch.distributed.elastic.utils import macros
from torch.distributed.elastic.utils.logging import get_logger
from torch.distributed.elastic.multiprocessing import PContext, start_processes
from taskd.python.utils.log import run_log
from taskd.api.taskd_proxy_api import init_taskd_proxy
from taskd.api.taskd_agent_api import init_taskd_agent, start_taskd_agent, register_func
from taskd.python.toolkit.constants.constants import SLEEP_GAP, MAX_INI16
from taskd.python.framework.common.type import CONFIG_UPSTREAMIP_KEY, LOCAL_HOST, CONFIG_FRAMEWORK_KEY
from taskd.python.framework.agent.pt_agent.pt_agent import get_pids


def patch_default_signal():
    time.sleep(SLEEP_GAP)
    return signal.SIGKILL


def patch_restart_workers(self, worker_group: WorkerGroup) -> None:
    role = self._worker_group.spec.role
    run_log.info("[%s], Stopping worker group", role)
    self._stop_workers(worker_group)
    worker_group.state = WorkerState.STOPPED
    self._worker_group.state = WorkerState.STOPPED
    run_log.info("[%s] Starting worker group", role)
    worker_ids = self._start_workers(worker_group)
    for local_rank, w_id in worker_ids.items():
        worker = self._worker_group.workers[local_rank]
        worker.id = w_id
    worker_group.state = WorkerState.HEALTHY
    self._worker_group.state = WorkerState.HEALTHY


def patch_stop_workers(self, worker_group: WorkerGroup) -> None:
    worker_local_ranks = {w.local_rank for w in worker_group.workers}
    run_log.info(f"stop workers, local rank: {worker_local_ranks}")
    if self._worker_watchdog is not None:
        self._worker_watchdog.stop()
        self._worker_watchdog = None
    if self._pcontext_dict is not None:
        threads = []
        for worker in worker_group.workers:
            p_context = self._pcontext_dict.get(worker.local_rank)
            t = threading.Thread(target=stop_worker_task, args=(p_context,))
            threads.append(t)
            t.start()
        run_log.info("start wait close func done")
        for t in threads:
            t.join()
    run_log.info("stop workers end")
    
    
def stop_worker_task(p_context):
    p_context.close(signal.SIGKILL)
    
    
def patch_start_workers(self, worker_group: WorkerGroup) -> Dict[int, Any]:
    worker_local_ranks = {w.local_rank for w in worker_group.workers}
    run_log.info(f"start workers, local rank: {worker_local_ranks}")
    if not hasattr(self, '_envs'):
        self._envs: Dict[int, Dict[str, str]] = {}
    spec = worker_group.spec
    store = worker_group.store
    assert store is not None
    assert spec.entrypoint is not None
    master_addr, master_port = os.getenv("MASTER_ADDR"), os.getenv("MASTER_PORT")
    restart_count = spec.max_restarts - self._remaining_restarts
    
    use_agent_store = get_use_agent_store(spec)
    run_log.info("use_agent_store: %s", use_agent_store)
    p_context_dict: Dict[int, PContext] = {}
    for worker in worker_group.workers:
        args: Dict[int, tuple] = {}
        envs: Dict[int, Dict[str, str]] = {}
        local_rank = worker.local_rank
        worker_env = get_worker_env(local_rank, master_addr, master_port, restart_count, spec, use_agent_store, worker,
                                    worker_group)
        envs[0] = worker_env
        self._envs[local_rank] = worker_env
        worker_args = list(spec.args)
        worker_args = macros.substitute(worker_args, str(local_rank))
        args[0] = tuple(worker_args)
        attempt_log_dir = ""
        if hasattr(self, '_log_dir') and self._log_dir:
            attempt_log_dir = os.path.join(self._log_dir, f"attempt_{restart_count}")
            shutil.rmtree(attempt_log_dir, ignore_errors=True)
            os.makedirs(attempt_log_dir)
        log_line_prefixes: Optional[Dict[int, str]] = None
        if hasattr(self, '_log_line_prefix_template') and self._log_line_prefix_template:
            log_line_prefixes = {}
            log_line_prefix = (Template(self._log_line_prefix_template).
                               safe_substitute(role_name=spec.role, rank=worker.global_rank, local_rank=local_rank))
            log_line_prefixes[local_rank] = log_line_prefix
        if hasattr(self, "_logs_specs") and self._logs_specs is not None:
            p_context = start_processes_with_logs_spec(args, self._logs_specs, log_line_prefixes, envs, 
                                                        self._start_method, spec)
        else:
            p_context = start_processes_with_log_dir(args, attempt_log_dir, envs, self._start_method, spec)
        self._pcontext_dict[local_rank] = p_context
        p_context_dict[local_rank] = p_context
    self._setup_local_watchdog(envs=self._envs)
    run_log.info("start worker end")
    return get_pids(p_context_dict)


def start_processes_with_log_dir(args, attempt_log_dir, envs, start_method, spec):
    return start_processes(
        name=spec.role, entrypoint=spec.entrypoint, args=args, envs=envs, log_dir=attempt_log_dir,
        start_method=start_method, redirects=spec.redirects, tee=spec.tee)
    

def start_processes_with_logs_spec(args, logs_specs, log_line_prefixes, envs, start_method, spec):
    return start_processes(
        name=spec.role, entrypoint=spec.entrypoint, args=args, envs=envs, logs_specs=logs_specs,
        log_line_prefixes=log_line_prefixes, start_method=start_method)


def get_worker_env(local_rank, master_addr, master_port, restart_count, spec, use_agent_store, worker, worker_group):
    worker_env = {
        "LOCAL_RANK": str(local_rank),
        "RANK": str(worker.global_rank),
        "GROUP_RANK": str(worker_group.group_rank),
        "ROLE_RANK": str(worker.role_rank),
        "ROLE_NAME": spec.role,
        "LOCAL_WORLD_SIZE": str(spec.local_world_size),
        "WORLD_SIZE": str(worker.world_size),
        "GROUP_WORLD_SIZE": str(worker_group.group_world_size),
        "ROLE_WORLD_SIZE": str(worker.role_world_size),
        "MASTER_ADDR": master_addr,
        "MASTER_PORT": str(master_port),
        "TORCHELASTIC_RESTART_COUNT": str(restart_count),
        "TORCHELASTIC_MAX_RESTARTS": str(spec.max_restarts),
        "TORCHELASTIC_RUN_ID": spec.rdzv_handler.get_run_id(),
        "TORCHELASTIC_USE_AGENT_STORE": str(use_agent_store),
        "NCCL_ASYNC_ERROR_HANDLING": os.getenv(
            "NCCL_ASYNC_ERROR_HANDLING", str(1)
        ),
        "TORCH_NCCL_ASYNC_ERROR_HANDLING": os.getenv(
            "TORCH_NCCL_ASYNC_ERROR_HANDLING", str(1)
        ),
    }
    if "OMP_NUM_THREADS" in os.environ:
        worker_env["OMP_NUM_THREADS"] = os.environ["OMP_NUM_THREADS"]
    return worker_env


def patch_monitor_workers(self, worker_group: WorkerGroup) -> RunResult:
    role = worker_group.spec.role
    worker_pids = {w.id for w in worker_group.workers}
    assert self._pcontext_dict is not None
    pc_pids = set(get_pids(self._pcontext_dict).values())
    if worker_pids != pc_pids:
        run_log.error(
            "[%s] worker pids do not match process_context pids."
            " Expected: %s, actual: %s",
            role,
            worker_pids,
            pc_pids,
        )
        return RunResult(state=WorkerState.UNKNOWN)
    result_dict = {}
    lock = threading.Lock()
    threads = []
    for local_rank, p_context in self._pcontext_dict.items():
        t = threading.Thread(target=context_wait_task, args=(local_rank, p_context, result_dict, lock))
        threads.append(t)
        t.start()
    for t in threads:
        t.join()
    
    worker_failures = {}
    workers_ret_vals = {}
    for local_rank, result in result_dict.items():
        if result is None:
            continue
        if result.is_failed():
            for _, failure in result.failures.items():
                worker = worker_group.workers[local_rank]
                worker_failures[worker.global_rank] = failure
        else:
            for _, ret_val in result.return_values.items():
                worker = worker_group.workers[local_rank]
                workers_ret_vals[worker.global_rank] = ret_val
    if len(workers_ret_vals) == 0 and len(worker_failures) == 0:
        return RunResult(state=WorkerState.HEALTHY)
    elif len(worker_failures) > 0:
        return RunResult(
            state=WorkerState.FAILED,
            failures=worker_failures,
        )
    else:
        return get_success_run_result(worker_group)


def get_success_run_result(worker_group: WorkerGroup) -> RunResult:
    workers_ret_vals = {worker.global_rank: "completed" for worker in worker_group.workers}
    return RunResult(
        state=WorkerState.SUCCEEDED,
        return_values=workers_ret_vals,
    )


# p_context only contain on process,
# when a failure occur or training end, only one of multiple p_context.wait calls (device num) can get a non-None value
def context_wait_task(local_rank, p_context, result_dict, lock):
    result = p_context.wait(0)
    with lock:
        result_dict[local_rank] = result


def get_use_agent_store(spec):
    if hasattr(spec.rdzv_handler, "use_agent_store"):
        return spec.rdzv_handler.use_agent_store
    return spec.rdzv_handler.get_backend() == "static"


def patch_launch_agent(config: LaunchConfig, entrypoint: Union[Callable, str, None], args: List[Any]
                       ) -> Dict[int, Any]:
    if config.max_restarts == 0:
        config.max_restarts = MAX_INI16
    return launch_agent(config, entrypoint, args)


def patch_invoke_run(self, role: str = DEFAULT_ROLE) -> RunResult:
    if not hasattr(self, '_pcontext_dict'):
        self._pcontext_dict: Dict[int, PContext] = {}
    
    proxy = threading.Thread(target=init_taskd_proxy, args=({CONFIG_UPSTREAMIP_KEY: os.getenv("MASTER_ADDR", LOCAL_HOST)},))
    proxy.daemon = True
    proxy.start()
    init_taskd_agent({CONFIG_FRAMEWORK_KEY: 'PyTorch'}, self)
    register_func('KILL_WORKER', self._stop_workers)
    register_func('START_ALL_WORKER', self._initialize_workers)
    register_func('MONITOR', self._monitor_workers)
    register_func('RESTART', self._restart_workers)
    run_log.info("start taskd agent")
    return start_taskd_agent()


def patch_torch_method():
    torch.distributed.elastic.agent.server.api.SimpleElasticAgent._invoke_run = patch_invoke_run
    torch.distributed.elastic.multiprocessing.api._get_default_signal = patch_default_signal
    torch.distributed.launcher.api.launch_agent = patch_launch_agent
    torch.distributed.elastic.agent.server.api.SimpleElasticAgent._restart_workers = patch_restart_workers
    torch.distributed.elastic.agent.server.local_elastic_agent.LocalElasticAgent._start_workers = patch_start_workers
    torch.distributed.elastic.agent.server.local_elastic_agent.LocalElasticAgent._stop_workers = patch_stop_workers
    torch.distributed.elastic.agent.server.local_elastic_agent.LocalElasticAgent._monitor_workers = patch_monitor_workers
