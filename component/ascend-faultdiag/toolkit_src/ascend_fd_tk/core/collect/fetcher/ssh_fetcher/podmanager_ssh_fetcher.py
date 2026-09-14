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

from typing import List

from ascend_fd_tk.core.collect.fetcher.podmanager_fetcher import PoDManagerFetcher
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.switch_ssh_fetcher import SwiSshFetcher
from ascend_fd_tk.utils.executors import CmdTask


class PoDManagerSshFetcher(SwiSshFetcher, PoDManagerFetcher):
    """PoDManager SSH 采集器：loginUBM -slot 进入 UBM 视图后复用 SwiSshFetcher switch 命令。"""

    def __init__(self, executor, slot_ids: List[int]):
        SwiSshFetcher.__init__(self, executor)
        PoDManagerFetcher.__init__(self, slot_ids)

    async def fetch_id(self) -> str:
        return f"{self.executor.host}_{self.current_slot}"

    async def switch_slot(self, slot_id: int):
        self.current_slot = slot_id

    async def init_fetcher(self):
        # pylint: disable=duplicate-code
        await self.executor.run_cmd(CmdTask(f"loginUBM -slot {self.current_slot}", timeout=1))
        await self.executor.run_cmd(CmdTask("n", timeout=1))
        await self.executor.run_cmd(CmdTask("sys", timeout=1))
        await self.executor.run_cmd(CmdTask("diag", timeout=1))

    async def last_quit(self):
        # pylint: disable=duplicate-code
        await self.executor.run_cmd(CmdTask("quit", timeout=1))
        await self.executor.run_cmd(CmdTask("quit", timeout=1))
        await self.executor.run_cmd(CmdTask("quit", timeout=1))
