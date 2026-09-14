#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

"""PoDManagerSshFetcher 单元测试"""

import asyncio
import unittest
from unittest.mock import AsyncMock, MagicMock

from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.podmanager_ssh_fetcher import PoDManagerSshFetcher
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.switch_ssh_fetcher import SwiSshFetcher
from ascend_fd_tk.core.collect.fetcher.podmanager_fetcher import PoDManagerFetcher
from ascend_fd_tk.utils.executors import CommandResult


def _make_executor(host="10.1.1.1") -> MagicMock:
    executor = MagicMock()
    executor.host = host
    executor.run_cmd = AsyncMock(return_value=CommandResult(cmd="x", returncode=0, stdout=""))
    return executor


class TestPoDManagerSshFetcher(unittest.TestCase):
    def setUp(self):
        self.executor = _make_executor()
        self.fetcher = PoDManagerSshFetcher(self.executor, [61, 62, 63, 64])

    def test_mro_inherits_switch_and_podmanager_fetcher(self):
        """应同时具备 switch 命令能力与 PoDManager 槽位能力。"""
        self.assertIsInstance(self.fetcher, SwiSshFetcher)
        self.assertIsInstance(self.fetcher, PoDManagerFetcher)

    def test_fetch_id_contains_host_and_current_slot(self):
        """初始 current_slot 为第一个槽位，fetch_id 返回 host_slot。"""
        self.assertEqual(self.fetcher.current_slot, 61)
        self.assertEqual(asyncio.run(self.fetcher.fetch_id()), "10.1.1.1_61")

    def test_switch_slot_updates_current_slot(self):
        """switch_slot 后 current_slot 与 fetch_id 同步更新。"""
        asyncio.run(self.fetcher.switch_slot(64))
        self.assertEqual(self.fetcher.current_slot, 64)
        self.assertEqual(asyncio.run(self.fetcher.fetch_id()), "10.1.1.1_64")

    def test_init_fetcher_sends_ubm_login_sequence(self):
        """init_fetcher 应按当前槽位执行 loginUBM -slot {slot} 并进入 diag 视图。"""
        asyncio.run(self.fetcher.switch_slot(63))
        asyncio.run(self.fetcher.init_fetcher())
        tasks = [call.args[0] for call in self.executor.run_cmd.await_args_list]
        self.assertEqual([task.cmd for task in tasks], ["loginUBM -slot 63", "n", "sys", "diag"])
        self.assertEqual(tasks[0].timeout, 1)


if __name__ == "__main__":
    unittest.main()
