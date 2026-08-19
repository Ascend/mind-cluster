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

import re
from typing import List

from ascend_fd_tk.core.collect.fetcher.host_fetcher import HostFetcher
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.base import SshFetcher
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.cmd_provider.host_provider import HostCmdProvider
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.cmd_provider.host_base import HostBaseProvider
from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.context.register import register_host_fetcher
from ascend_fd_tk.core.log_parser.base import FindResult
from ascend_fd_tk.core.log_parser.parse_config import msnpureport_log_config
from ascend_fd_tk.core.log_parser.remote_log_parser import RemoteLogPyScriptParser
from ascend_fd_tk.core.model.cluster_mapping import DefaultNPUInfo
from ascend_fd_tk.utils.executors import CmdTask
from ascend_fd_tk.utils import logger

_CONSOLE_LOGGER = logger.CONSOLE_LOGGER


@register_host_fetcher(NpuType.A3)
class HostSshFetcher(SshFetcher, HostFetcher):
    _ANSI_ESCAPE = re.compile(r'\x1b\[[0-?]*[ -/]*[@-~]')

    def __init__(self, executor, npu_mapping_cache: dict = None):
        super().__init__(executor)
        # 默认 A3 命令提供者，保持向后兼容；上层可按代际注入
        self.cmd_provider: HostBaseProvider = HostCmdProvider()
        self.chip_generation: NpuType = NpuType.A3
        # 代际探测阶段已执行过 npu-smi info -m，缓存其 npu_mapping 避免重复采集
        self._npu_mapping_cache = npu_mapping_cache

    async def fetch_id(self):
        return self.executor.host

    async def fetch_hostname(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.hostname_cmd()))
        lines = cmd_res.stdout.strip().splitlines()
        if lines:
            return lines[-1].strip()
        return ""

    async def fetch_npu_mapping(self) -> dict:
        """
        获取NPU映射信息
        通过执行npu-smi命令获取NPU芯片的映射关系，解析命令输出并构建成字典结构
        Args:
            self: 类实例
        Returns:
            dict: NPU映射字典，格式为 {npu_id: {chip_id: chip_phy_id}}
                  其中npu_id为NPU编号，chip_id为芯片ID，chip_phy_id为芯片物理ID
        """
        # 代际探测阶段已缓存 npu_mapping 时直接复用，避免重复执行 npu-smi info -m
        if self._npu_mapping_cache is not None:
            return self._npu_mapping_cache
        command_stdout = ""
        try:
            command_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.npu_mapping_cmd()))
            if command_res.is_success():
                command_stdout = command_res.stdout
            else:
                _CONSOLE_LOGGER.info("执行失败：%s", command_res.stderr)
        except Exception as e:
            _CONSOLE_LOGGER.info(e)
        npu_mapping = {}
        lines = command_stdout.strip().split('\n')
        for line in lines[2:]:
            parts = re.split(r'\s{2,}', line.strip())
            if len(parts) >= 5:
                npu_id = parts[0]
                chip_id = parts[1]
                chip_phy_id = parts[3]
                chip_name = parts[4]
                if chip_name != 'Mcu' and npu_id != '-':
                    npu_mapping.update(
                        {chip_phy_id: DefaultNPUInfo(npu_id=npu_id, chip_id=chip_id, chip_phy_id=chip_phy_id)}
                    )
        return npu_mapping

    async def fetch_optical_info(self, chip_phy_id) -> str:
        # A3 单端口光模块采集；A5 需遍历 optical_id，由子类重写。
        command_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.optical_cmd(chip_phy_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_link_stat_info(self, chip_phy_id) -> str:
        command_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.link_stat_cmd(chip_phy_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_stat_info(self, chip_phy_id) -> str:
        command_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.stat_cmd(chip_phy_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_lldp_info(self, chip_phy_id) -> str:
        command_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.lldp_cmd(chip_phy_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_npu_type(self) -> str:
        command_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.npu_type_cmd()))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_sn_num(self) -> str:
        command_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.sn_num_cmd()))
        if not command_res.is_success():
            return ""
        # 先清理ANSI转义码
        clean_out = self._ANSI_ESCAPE.sub("", command_res.stdout.strip())
        lines = clean_out.split('\n')
        # 返回第一行非空行作为序列号
        for line in lines[1:]:  # 跳过第一行标题行
            if line.strip():
                return line.strip()
        return ""

    async def fetch_hccs_info(self, npu_id, chip_id) -> str:
        command_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.hccs_cmd(npu_id, chip_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_spod_info(self, npu_id, chip_id) -> str:
        command_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.spod_cmd(npu_id, chip_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_msnpureport_log(self) -> List[FindResult]:
        recv = await self.executor.run_cmd(CmdTask(self.cmd_provider.msnpureport_cmd()))
        output_dir = recv.stdout.splitlines()[1].split(":")[-1].strip()
        msnpureport_pattern_map = {}
        for config in msnpureport_log_config.MS_NPU_REPORT_PARSE_CONFIG:
            msnpureport_pattern_map[config.keyword_config.pattern_key] = config
        res = await RemoteLogPyScriptParser(self.executor).find(output_dir, msnpureport_pattern_map)
        return res

    async def fetch_roce_speed(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.speed_cmd(chip_phy_id)))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_roce_duplex(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.duplex_cmd(chip_phy_id)))
        parts = cmd_res.stdout.strip().splitlines()[-1].split(":")
        if len(parts) >= 2 and "Duplex" in parts[0]:
            return parts[1].strip()
        return ""

    async def fetch_hccn_tool_net_health(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.net_health_cmd(chip_phy_id)))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_hccn_tool_link_status(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.hccn_link_cmd(chip_phy_id)))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_hccn_tool_cdr_snr(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.cdr_snr_cmd(chip_phy_id)))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_hccn_dfx_cfg(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(self.cmd_provider.dfx_cfg_cmd(chip_phy_id)))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_optical_loopback_enable(self, npu_id, model) -> str:
        all_cmd_list = [self.cmd_provider.optical_loopback_enable_cmd(npu_id, model), "y"]
        all_cmd_str = "\n".join(all_cmd_list)
        cmd_res = await self.executor.run_cmd(CmdTask(all_cmd_str))
        return cmd_res.stdout
