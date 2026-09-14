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

"""InitFetcher._add_pod_manager_executors 单元测试。"""

import asyncio
import unittest
from unittest.mock import patch

from ascend_fd_tk.core.common import constants
from ascend_fd_tk.core.config.conn_config import Conn
from ascend_fd_tk.core.service.init_fetcher import InitFetcher


class _FakeExecutor:
    """模拟 AsyncSSHExecutor：shell_channel 为空表示建连失败。"""

    def __init__(self, host, ok):
        self.host = host
        self.shell_channel = object() if ok else None

    def ensure_shell_session(self):
        return None


class _FakeFetcher:
    """模拟 PoDManagerSshFetcher：记录分到的槽位列表。"""

    def __init__(self, executor, slot_ids):
        self.host = executor.host
        self.slot_ids = list(slot_ids)


class TestInitFetcherPodManager(unittest.TestCase):
    def setUp(self):
        self.conn = Conn(host="10.1.1.1", port=22, username="root", password="pwd", private_key=None)

    def _run(self, executor_factory):
        fetchers_map = {}
        with (
            patch("ascend_fd_tk.core.service.init_fetcher.AsyncSSHExecutor", side_effect=executor_factory),
            patch("ascend_fd_tk.core.service.init_fetcher.PoDManagerSshFetcher", side_effect=_FakeFetcher),
        ):
            asyncio.run(InitFetcher._add_pod_manager_executors(self.conn, fetchers_map))
        return fetchers_map

    def test_redistribute_slots_when_one_connection_fails(self):
        """建连失败时按成功连接数重新分片，所有槽位仍被覆盖。"""
        all_slots = constants.POD_MANAGER_SWITCH_SLOT_IDS
        executors = [_FakeExecutor("10.1.1.1", ok=(i != 2)) for i in range(len(all_slots))]  # 第 2 个连接失败
        fetchers_map = self._run(executors)
        self.assertEqual(len(fetchers_map), len(all_slots) - 1)
        covered_slots = set()
        for fetcher in fetchers_map.values():
            covered_slots.update(fetcher.slot_ids)
        self.assertEqual(covered_slots, set(all_slots))
        # 分片互不重叠
        slot_lists = [f.slot_ids for f in fetchers_map.values()]
        for i in range(len(slot_lists)):
            for j in range(i + 1, len(slot_lists)):
                self.assertEqual(set(slot_lists[i]) & set(slot_lists[j]), set())

    def test_all_connections_fail_skips_collection(self):
        """所有连接建连失败时直接返回，不注册任何 fetcher。"""
        executors = [_FakeExecutor("10.1.1.1", ok=False) for _ in range(len(constants.POD_MANAGER_SWITCH_SLOT_IDS))]
        self.assertEqual(self._run(executors), {})

    def test_info_log_initialized(self):
        """初始化成功应打印包含 IP 与连接数的 INFO 日志。"""
        all_slots = constants.POD_MANAGER_SWITCH_SLOT_IDS
        executors = [_FakeExecutor("10.1.1.1", ok=(i != 2)) for i in range(len(all_slots))]
        fetchers_map = {}
        with (
            self.assertLogs("ascend_fd_tk", level="INFO") as cm,
            patch("ascend_fd_tk.core.service.init_fetcher.AsyncSSHExecutor", side_effect=executors),
            patch("ascend_fd_tk.core.service.init_fetcher.PoDManagerSshFetcher", side_effect=_FakeFetcher),
        ):
            asyncio.run(InitFetcher._add_pod_manager_executors(self.conn, fetchers_map))
        self.assertEqual(len(fetchers_map), len(all_slots) - 1)
        self.assertTrue(any("初始化成功" in msg and "10.1.1.1" in msg for msg in cm.output))


if __name__ == "__main__":
    unittest.main()
