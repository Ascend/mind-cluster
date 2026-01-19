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
import unittest
import signal
import threading
from unittest.mock import patch, MagicMock
import torch.distributed.elastic.agent.server.api
import torch.distributed.elastic.multiprocessing.api
from taskd.python.adaptor.patch.torch_patch import patch_torch_method, patch_invoke_run, patch_default_signal
from taskd.python.toolkit.constants.constants import SLEEP_GAP
from taskd.python.adaptor.patch.torch_patch import (
    patch_restart_workers, patch_stop_workers, stop_worker_task,
    patch_start_workers, start_processes_with_log_dir,
    start_processes_with_logs_spec, get_worker_env,
    patch_monitor_workers, context_wait_task, get_use_agent_store
)

from torch.distributed.elastic.agent.server.api import WorkerGroup, WorkerState, RunResult
from torch.distributed.elastic.multiprocessing import PContext


class TestTorchPatch(unittest.TestCase):
    def setUp(self):
        self.original_invoke_run = torch.distributed.elastic.agent.server.api.SimpleElasticAgent._invoke_run
        self.original_get_signal = torch.distributed.elastic.multiprocessing.api._get_default_signal

    def tearDown(self):
        torch.distributed.elastic.agent.server.api.SimpleElasticAgent._invoke_run = self.original_invoke_run
        torch.distributed.elastic.multiprocessing.api._get_default_signal = self.original_get_signal

    def test_patch_torch_method(self):
        patch_torch_method()
        self.assertEqual(torch.distributed.elastic.agent.server.api.SimpleElasticAgent._invoke_run, patch_invoke_run)
        self.assertEqual(torch.distributed.elastic.multiprocessing.api._get_default_signal, patch_default_signal)

    @patch('taskd.python.adaptor.patch.torch_patch.threading.Thread')
    @patch('taskd.python.adaptor.patch.torch_patch.init_taskd_proxy')
    @patch('taskd.python.adaptor.patch.torch_patch.init_taskd_agent')
    @patch('taskd.python.adaptor.patch.torch_patch.register_func')
    @patch('taskd.python.adaptor.patch.torch_patch.start_taskd_agent')
    def test_patch_invoke_run(self, mock_start_agent, mock_register, mock_init_agent, mock_init_proxy, mock_thread):
        mock_self = MagicMock()
        mock_thread_instance = MagicMock()
        mock_thread.return_value = mock_thread_instance
        mock_start_agent.return_value = 'test_result'

        result = patch_invoke_run(mock_self)

        mock_thread.assert_called_once()
        mock_thread_instance.start.assert_called_once()

        mock_init_agent.assert_called_once()

        self.assertEqual(mock_register.call_count, 4)
        expected_calls = [
            unittest.mock.call('KILL_WORKER', mock_self._stop_workers),
            unittest.mock.call('START_ALL_WORKER', mock_self._initialize_workers),
            unittest.mock.call('MONITOR', mock_self._monitor_workers),
            unittest.mock.call('RESTART', mock_self._restart_workers)
        ]
        mock_register.assert_has_calls(expected_calls, any_order=False)

        mock_start_agent.assert_called_once()
        self.assertEqual(result, 'test_result')

    @patch('taskd.python.adaptor.patch.torch_patch.time.sleep')
    def test_patch_default_signal(self, mock_sleep):
        result = patch_default_signal()

        mock_sleep.assert_called_once_with(SLEEP_GAP)
        self.assertEqual(result, signal.SIGKILL)


class TestTorchPatch(unittest.TestCase):

    def setUp(self):
        self.mock_self = MagicMock()
        self.mock_worker_group = MagicMock(spec=WorkerGroup)
        self.mock_spec = MagicMock()
        self.mock_spec.role = "test_role"
        self.mock_worker_group.spec = self.mock_spec
        self.mock_worker_group.state = WorkerState.HEALTHY
        self.mock_self._worker_group = self.mock_worker_group
        self.mock_self._worker_watchdog = MagicMock()
    
    def test_patch_restart_workers(self):
        mock_worker_ids = {0: 100, 1: 101}
        self.mock_self._start_workers.return_value = mock_worker_ids

        mock_worker0 = MagicMock()
        mock_worker1 = MagicMock()
        self.mock_worker_group.workers = {0: mock_worker0, 1: mock_worker1}

        patch_restart_workers(self.mock_self, self.mock_worker_group)

        self.mock_self._stop_workers.assert_called_once_with(self.mock_worker_group)
        self.mock_self._start_workers.assert_called_once_with(self.mock_worker_group)
        self.assertEqual(self.mock_worker_group.state, WorkerState.HEALTHY)
        self.assertEqual(self.mock_self._worker_group.state, WorkerState.HEALTHY)
        self.assertEqual(mock_worker0.id, 100)
        self.assertEqual(mock_worker1.id, 101)

    @patch('taskd.python.adaptor.patch.torch_patch.stop_worker_task')
    def test_patch_stop_workers_with_pcontext_dict(self, mock_stop_worker_task):
        mock_worker1 = MagicMock(local_rank=0)
        mock_worker2 = MagicMock(local_rank=1)
        self.mock_worker_group.workers = [mock_worker1, mock_worker2]

        mock_pcontext0 = MagicMock()
        mock_pcontext1 = MagicMock()
        self.mock_self._pcontext_dict = {0: mock_pcontext0, 1: mock_pcontext1}
        self.mock_self._worker_watchdog = MagicMock()
        mock_worker_watchdog = self.mock_self._worker_watchdog

        patch_stop_workers(self.mock_self, self.mock_worker_group)

        mock_worker_watchdog.stop.assert_called_once()
        self.assertIsNone(self.mock_self._worker_watchdog)
        self.assertEqual(mock_stop_worker_task.call_count, 2)

    def test_patch_stop_workers_without_pcontext_dict(self):
        self.mock_self._pcontext_dict = None
        self.mock_worker_group.workers = []
        self.mock_self._worker_watchdog = MagicMock()
        mock_worker_watchdog = self.mock_self._worker_watchdog

        patch_stop_workers(self.mock_self, self.mock_worker_group)

        mock_worker_watchdog.stop.assert_called_once()
        self.assertIsNone(self.mock_self._worker_watchdog)

    def test_stop_worker_task(self):
        mock_pcontext = MagicMock()

        stop_worker_task(mock_pcontext)

        mock_pcontext.close.assert_called_once()

    @patch('taskd.python.adaptor.patch.torch_patch.get_pids')
    @patch('taskd.python.adaptor.patch.torch_patch.get_use_agent_store')
    @patch('taskd.python.adaptor.patch.torch_patch.get_worker_env')
    @patch('taskd.python.adaptor.patch.torch_patch.start_processes_with_log_dir')
    @patch('os.getenv')
    @patch('taskd.python.adaptor.patch.torch_patch.Template')
    def test_patch_start_workers(self, mock_template, mock_getenv, mock_start_processes, 
                               mock_get_worker_env, mock_get_use_agent_store, mock_get_pids):
        def mock_getenv_side_effect(key, *args):
            env_vars = {
                "Master_ADDR": "localhost",
                "MASTER_PORT": "29500"
            }
            if key in env_vars:
                return env_vars[key]
            elif args:
                return args[0]
            else:
                return None

        mock_getenv.side_effect = mock_getenv_side_effect

        mock_get_use_agent_store.return_value = True
        mock_worker_env = {"ENV_KEY": "ENV_VALUE"}
        mock_get_worker_env.return_value = mock_worker_env

        mock_pcontext = MagicMock(spec=PContext)
        mock_start_processes.return_value = mock_pcontext

        mock_pids = {0: 100, 1: 101}
        mock_get_pids.return_value = mock_pids

        mock_template_instance = MagicMock()
        mock_template.return_value = mock_template_instance
        mock_template_instance.safe_substitute.return_value = "[worker0]"

        mock_worker1 = MagicMock(local_rank=0, global_rank=0)
        mock_worker2 = MagicMock(local_rank=1, global_rank=1)
        self.mock_worker_group.workers = [mock_worker1, mock_worker2]
        self.mock_spec.args = ["--arg1", "value1"]
        self.mock_spec.max_restarts = 3
        self.mock_self._remaining_restarts = 2
        self.mock_self._start_method = "spawn"
        self.mock_self._log_line_prefix_template = "[${role_name}:${rank}:${local_rank}]"
        self.mock_self._logs_specs = None

        result = patch_start_workers(self.mock_self, self.mock_worker_group)

        self.assertEqual(result, mock_pids)
        self.mock_self._setup_local_watchdog.assert_called_once_with(envs=self.mock_self._envs)
        mock_template.assert_called_with("[${role_name}:${rank}:${local_rank}]")
        mock_template_instance.safe_substitute.assert_called()

    @patch('taskd.python.adaptor.patch.torch_patch.start_processes')
    def test_start_processes_with_log_dir(self, mock_start_processes):
        mock_args = {0: ("python", "script.py")}
        mock_envs = {0: {"ENV": "VALUE"}}
        mock_spec = MagicMock()
        mock_spec.role = "test_role"
        mock_spec.entrypoint = "python"
        mock_spec.redirects = {}
        mock_spec.tee = False

        start_processes_with_log_dir(mock_args, "log_dir", mock_envs, "spawn", mock_spec)

        mock_start_processes.assert_called_once_with(
            name="test_role",
            entrypoint="python",
            args=mock_args,
            envs=mock_envs,
            log_dir="log_dir",
            start_method="spawn",
            redirects={},
            tee=False
        )

    @patch('taskd.python.adaptor.patch.torch_patch.start_processes')
    def test_start_processes_with_logs_specs(self, mock_start_processes):
        mock_args = {0: ("python", "script.py")}
        mock_envs = {0: {"ENV": "VALUE"}}
        mock_logs_specs = {"stdout": "log.txt"}
        mock_log_line_prefixes = {0: "[worker0]"}
        mock_spec = MagicMock()
        mock_spec.role = "test_role"
        mock_spec.entrypoint = "python"
        start_processes_with_logs_spec(
            mock_args, mock_logs_specs, mock_log_line_prefixes, mock_envs, "spawn", mock_spec
        )

        mock_start_processes.assert_called_once_with(
            name="test_role",
            entrypoint="python",
            args=mock_args,
            envs=mock_envs,
            logs_specs=mock_logs_specs,
            log_line_prefixes=mock_log_line_prefixes,
            start_method="spawn"
        )

    @patch('os.getenv')
    def test_get_worker_env(self, mock_getenv):
        def mock_getenv_side_effect(x, *args):
            env_var_dict = {
                "NCCL_ASYNC_ERROR_HANDLING": "1",
                "TORCH_NCCL_ASYNC_ERROR_HANDLING": "1"
            }
            return env_var_dict.get(x, args[0] if args else None)

        mock_getenv.side_effect = mock_getenv_side_effect

        mock_spec = MagicMock()
        mock_spec.role = "test_role"
        mock_spec.local_world_size = 2
        mock_spec.max_restarts = 3
        mock_rdzv_handler = MagicMock()
        mock_rdzv_handler.get_run_id.return_value = "run_123"
        mock_spec.rdzv_handler = mock_rdzv_handler

        mock_worker = MagicMock()
        mock_worker.global_rank = 0
        mock_worker.role_rank = 0
        mock_worker.world_size = 4
        mock_worker.role_world_size = 2

        mock_worker_group = MagicMock()
        mock_worker_group.group_rank = 0
        mock_worker_group.group_world_size = 2

        result = get_worker_env(
            local_rank=0,
            master_addr="localhost",
            master_port="29500",
            restart_count=1,
            spec=mock_spec,
            use_agent_store=True,
            worker=mock_worker,
            worker_group=mock_worker_group
        )

        self.assertEqual(result["LOCAL_RANK"], "0")
        self.assertEqual(result["RANK"], "0")
        self.assertEqual(result["ROLE_NAME"], "test_role")
        self.assertEqual(result["MASTER_ADDR"], "localhost")
        self.assertEqual(result["MASTER_PORT"], "29500")

    @patch('taskd.python.adaptor.patch.torch_patch.get_pids')
    @patch('taskd.python.adaptor.patch.torch_patch.context_wait_task')
    def test_patch_monitor_workers_healthy(self, mock_context_wait_task, mock_get_pids):
        mock_worker1 = MagicMock(id=100, global_rank=0)
        mock_worker2 = MagicMock(id=101, global_rank=1)
        self.mock_worker_group.workers = [mock_worker1, mock_worker2]

        mock_pcontext0 = MagicMock()
        mock_pcontext1 = MagicMock()
        self.mock_self._pcontext_dict = {0: mock_pcontext0, 1: mock_pcontext1}

        mock_get_pids.return_value = {0: 100, 1: 101}

        result = patch_monitor_workers(self.mock_self, self.mock_worker_group)

        self.assertEqual(result.state, WorkerState.HEALTHY)
        self.assertEqual(mock_context_wait_task.call_count, 2)

    def test_context_wait_task(self):
        mock_pcontext = MagicMock()
        mock_result = MagicMock()
        mock_pcontext.wait.return_value = mock_result
        result_dict = {}
        lock = threading.Lock()

        context_wait_task(0, mock_pcontext, result_dict, lock)

        mock_pcontext.wait.assert_called_once_with(0)
        self.assertEqual(result_dict.get(0), mock_result)
    
    def test_get_use_agent_store_with_attribute(self):
        mock_spec = MagicMock()
        mock_spec.rdzv_handler.use_agent_store = True
        result = get_use_agent_store(mock_spec)
        self.assertTrue(result)

    def test_get_use_agent_store_with_backend(self):
        mock_spec = MagicMock()
        delattr(mock_spec.rdzv_handler, 'use_agent_store')
        mock_spec.rdzv_handler.get_backend.return_value = "static"
        result = get_use_agent_store(mock_spec)
        self.assertTrue(result)

    def test_get_use_agent_store_with_backend_not_static(self):
        mock_spec = MagicMock()
        delattr(mock_spec.rdzv_handler, 'use_agent_store')
        mock_spec.rdzv_handler.get_backend.return_value = "dynamic"
        result = get_use_agent_store(mock_spec)
        self.assertFalse(result)


if __name__ == '__main__':
    unittest.main()
