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

import os
from typing import List

from ascend_fd_tk.core.collect.fetcher.bmc_fetcher import BmcFetcher
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.base import SshFetcher
from ascend_fd_tk.core.collect.parser.bmc_parser import BmcParser
from ascend_fd_tk.core.common import diag_enum
from ascend_fd_tk.core.common.path import CommonPath
from ascend_fd_tk.core.model.bmc import BmcSensorInfo, BmcSelInfo, BmcHealthEvents, \
    LinkDownOpticalModuleHistoryLog
from ascend_fd_tk.utils import logger, helpers
from ascend_fd_tk.utils.executors import CmdTask, AsyncSSHExecutor

LOGGER = logger.DIAG_LOGGER


class BmcSshFetcher(SshFetcher, BmcFetcher):
    _DUMP_INFO_TAR = "dump_info.tar.gz"

    def __init__(self, executor: AsyncSSHExecutor):
        super().__init__(executor)
        self.parser = BmcParser()

    async def fetch_id(self) -> str:
        return self.executor.host

    async def fetch_bmc_sn(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("ipmcget -d serialnumber", timeout=10))
        return cmd_res.stdout.replace("ipmcget -d serialnumber\r\nSystem SN is:", "").strip()

    async def fetch_bmc_date(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("ipmcget -d time", timeout=10))
        date_str = cmd_res.stdout.replace("ipmcget -d time\r\n", "").strip()
        date_parts = date_str.split()
        if len(date_parts) < 3:
            return ""
        return f"{date_parts[0]} {date_parts[2]}"

    async def fetch_bmc_sel_list(self) -> List[BmcSelInfo]:
        cmd_res = await self.executor.run_cmd(CmdTask("ipmcget -d sel -v list", timeout=10))
        await self.executor.run_cmd(CmdTask("q", timeout=10))
        return self.parser.trans_sel_results(cmd_res.stdout)

    async def fetch_bmc_sensor_list(self) -> List[BmcSensorInfo]:
        cmd_res = await self.executor.run_cmd(CmdTask("ipmcget -t sensor -d list", timeout=10))
        return self.parser.trans_sensor_results(cmd_res.stdout)

    async def fetch_bmc_health_events(self) -> List[BmcHealthEvents]:
        cmd_res = await self.executor.run_cmd(CmdTask("ipmcget -d healthevents", timeout=10))
        await self.executor.run_cmd(CmdTask("q", timeout=10))
        return self.parser.trans_health_events_results(cmd_res.stdout)

    async def fetch_bmc_diag_info_log(self):
        res = await self.executor.run_cmd(CmdTask("ipmcget -d diaginfo", timeout=900))
        remote_path = f"/tmp/{self._DUMP_INFO_TAR}"
        sn_num = await self.fetch_bmc_sn()
        date_time = await self.fetch_bmc_date()
        subfix_date_time = helpers.trans_date_fmt(date_time, diag_enum.TimeFormat.BMC_DATE_FMT.value,
                                                  diag_enum.TimeFormat.BMC_TAR_FILE.value)
        os.makedirs(CommonPath.TOOL_HOME_BMC_DUMP_CACHE_DIR, exist_ok=True)
        dump_path = os.path.join(CommonPath.TOOL_HOME_BMC_DUMP_CACHE_DIR,
                                 f"{self.executor.host}_{sn_num}_{subfix_date_time}.tar.gz")
        await self.executor.download_file(remote_path, dump_path)
        return res

    async def fetch_bmc_optical_module_history_info_log(self) -> List[LinkDownOpticalModuleHistoryLog]:
        return []
