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

from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.cmd_provider.host_provider_a5 import HostCmdProviderA5
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.host_ssh_fetcher import HostSshFetcher
from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.context.register import register_host_fetcher
from ascend_fd_tk.utils.executors import CmdTask


@register_host_fetcher(NpuType.A5)
class HostSshFetcherA5(HostSshFetcher):
    def __init__(self, executor, npu_mapping_cache: dict = None):
        super().__init__(executor, npu_mapping_cache)
        self.cmd_provider: HostCmdProviderA5 = HostCmdProviderA5()
        self.chip_generation: NpuType = NpuType.A5

    async def fetch_optical_top_headline(self, npu_id: str) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.optical_top_headline(npu_id)))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_optical_info_a5(self, npu_id: str, optical_id: str) -> str:
        if not npu_id or not optical_id:
            return ""
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.optical_info_cmd(npu_id, optical_id)))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_nic_info(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.nic_info_cmd()))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_nic_port_num(self, card_name: str) -> str:
        if not card_name:
            return ""
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.nic_port_num_cmd(card_name)))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_nic_sfp_info(self, card_name: str, port_id: str) -> str:
        if not card_name or port_id == "":
            return ""
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.nic_sfp_cmd(card_name, port_id)))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""
