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

"""HostSshFetcherA5 单元测试。

覆盖本次改动：
- @register_host_fetcher(NpuType.A5) 代际注册
- A5 新增 fetch 方法（fetch_optical_top_headline / fetch_optical_info_a5 / fetch_nic_*）
- chip_generation = NpuType.A5
"""

import asyncio
import unittest
from unittest.mock import AsyncMock, MagicMock

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.context.register import HOST_FETCHER_REGISTRY
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.host_ssh_fetcher_a5 import HostSshFetcherA5
from ascend_fd_tk.utils.executors import CommandResult


def _make_executor(stdout: str = "", returncode: int = 0) -> MagicMock:
    """构造 mock executor，run_cmd 返回指定 stdout 的成功结果。"""
    executor = MagicMock()
    executor.host = "host1"
    executor.run_cmd = AsyncMock(return_value=CommandResult(cmd="x", returncode=returncode, stdout=stdout))
    return executor


class TestHostSshFetcherA5(unittest.TestCase):
    def test_generation_registration(self):
        """A5 fetcher 应注册到 HOST_FETCHER_REGISTRY。"""
        self.assertIs(HOST_FETCHER_REGISTRY[NpuType.A5], HostSshFetcherA5)

    def test_chip_generation_is_a5(self):
        """A5 fetcher 的 chip_generation 应为 NpuType.A5。"""
        fetcher = HostSshFetcherA5(_make_executor())
        self.assertEqual(fetcher.chip_generation, NpuType.A5)

    def test_fetch_optical_top_headline_success(self):
        """fetch_optical_top_headline 成功时返回 stdout。"""
        fetcher = HostSshFetcherA5(_make_executor(stdout="optical headline"))
        result = asyncio.run(fetcher.fetch_optical_top_headline("0"))
        self.assertEqual(result, "optical headline")

    def test_fetch_optical_info_a5_empty_input(self):
        """npu_id 或 optical_id 为空时应返回空串，不执行命令。"""
        fetcher = HostSshFetcherA5(_make_executor(stdout="should_not_reach"))
        self.assertEqual(asyncio.run(fetcher.fetch_optical_info_a5("", "0")), "")
        self.assertEqual(asyncio.run(fetcher.fetch_optical_info_a5("0", "")), "")
        fetcher.executor.run_cmd.assert_not_called()

    def test_fetch_nic_sfp_info_empty_card(self):
        """card_name 为空时应返回空串。"""
        fetcher = HostSshFetcherA5(_make_executor(stdout="should_not_reach"))
        self.assertEqual(asyncio.run(fetcher.fetch_nic_sfp_info("", "0")), "")
        fetcher.executor.run_cmd.assert_not_called()

    def test_fetch_failed_returns_empty(self):
        """命令失败时（returncode != 0）应返回空串。"""
        fetcher = HostSshFetcherA5(_make_executor(stdout="error", returncode=1))
        self.assertEqual(asyncio.run(fetcher.fetch_optical_top_headline("0")), "")
        self.assertEqual(asyncio.run(fetcher.fetch_nic_info()), "")


if __name__ == "__main__":
    unittest.main()
