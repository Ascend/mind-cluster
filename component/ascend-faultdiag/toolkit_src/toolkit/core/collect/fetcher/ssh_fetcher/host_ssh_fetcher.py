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

from toolkit.core.collect.fetcher.host_fetcher import HostFetcher
from toolkit.core.collect.fetcher.ssh_fetcher.base import SshFetcher
from toolkit.core.log_parser.base import FindResult
from toolkit.core.log_parser.parse_config import msnpureport_log_config
from toolkit.core.log_parser.remote_log_parser import RemoteLogPyScriptParser
from toolkit.utils.executors import AsyncSSHExecutor, CmdTask
from toolkit.utils import logger

_CONSOLE_LOGGER = logger.CONSOLE_LOGGER


class HostSshFetcher(SshFetcher, HostFetcher):

    def __init__(self, executor: AsyncSSHExecutor):
        super().__init__(executor)

    async def fetch_id(self):
        return self.executor.host

    async def fetch_hostname(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("hostname"))
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
        command_stdout = ""
        try:
            command_res = await self.executor.run_cmd(CmdTask("npu-smi info -m"))
            if command_res.is_success():
                command_stdout = command_res.stdout
            else:
                _CONSOLE_LOGGER.info("执行失败：", command_res.stderr)
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
                    if npu_id not in npu_mapping:
                        npu_mapping[npu_id] = {}
                    npu_mapping[npu_id][chip_id] = chip_phy_id

        return npu_mapping

    async def fetch_optical_info(self, chip_phy_id) -> str:
        command_res = await self.executor.run_cmd(CmdTask("hccn_tool -i {} -optical -g".format(chip_phy_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_link_stat_info(self, chip_phy_id) -> str:
        command_res = await self.executor.run_cmd(CmdTask("hccn_tool -i {} -link_stat -g".format(chip_phy_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_stat_info(self, chip_phy_id) -> str:
        command_res = await self.executor.run_cmd(CmdTask("hccn_tool -i {} -stat -g".format(chip_phy_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_lldp_info(self, chip_phy_id) -> str:
        command_res = await self.executor.run_cmd(CmdTask("hccn_tool -i {} -lldp -g".format(chip_phy_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_npu_type(self) -> str:
        command_res = await self.executor.run_cmd(CmdTask("lspci |grep 'Device d80' --color=never"))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_sn_num(self) -> str:
        command_res = await self.executor.run_cmd(CmdTask("dmidecode -s system-serial-number"))
        if command_res.is_success():
            lines = command_res.stdout.strip().split('\n')
            # 返回第一行非空行作为序列号
            for line in lines[1:]:  # 跳过第一行标题行
                if line.strip():
                    return line.strip()
        return ""

    async def fetch_hccs_info(self, npu_id, chip_id) -> str:
        command_res = await self.executor.run_cmd(
            CmdTask("npu-smi info -t hccs -i {} -c {}".format(npu_id, chip_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_spod_info(self, npu_id, chip_id) -> str:
        command_res = await self.executor.run_cmd(
            CmdTask("npu-smi info -t spod-info -i {} -c {}".format(npu_id, chip_id)))
        if command_res.is_success():
            return command_res.stdout
        return ""

    async def fetch_msnpureport_log(self) -> List[FindResult]:
        recv = await self.executor.run_cmd(CmdTask("msnpureport"))
        output_dir = recv.stdout.splitlines()[1].split(":")[-1].strip()
        msnpureport_pattern_map = {}
        for config in msnpureport_log_config.MS_NPU_REPORT_PARSE_CONFIG:
            msnpureport_pattern_map[config.keyword_config.pattern_key] = config
        res = await RemoteLogPyScriptParser(self.executor).find(output_dir, msnpureport_pattern_map)
        return res

    async def fetch_roce_speed(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(f"hccn_tool -i {chip_phy_id} -speed -g"))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_roce_duplex(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(f"hccn_tool -i {chip_phy_id} -duplex -g"))
        parts = cmd_res.stdout.strip().splitlines()[-1].split(":")
        if len(parts) >= 2 and "Duplex" in parts[0]:
            return parts[1].strip()
        return ""

    async def fetch_hccn_tool_net_health(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(f"hccn_tool -i {chip_phy_id} -net_health -g"))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_hccn_tool_link_status(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(f"hccn_tool -i {chip_phy_id} -link_status -g"))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_hccn_tool_cdr(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(f"hccn_tool -i {chip_phy_id} -scdr -t 5"))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_hccn_dfx_cfg(self, chip_phy_id) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask(f"hccn_tool -i {chip_phy_id} -optical -g dfx_cfg"))
        if cmd_res.is_success():
            return cmd_res.stdout
        return ""

    async def fetch_optical_loopback_enable(self, npu_id, model) -> str:
        all_cmd_list = [f"hccn_tool -i {npu_id} -optical -t {model}", "y"]
        all_cmd_str = "\n".join(all_cmd_list)
        cmd_res = await self.executor.run_cmd(CmdTask(all_cmd_str))
        return cmd_res.stdout
