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
from typing import List, Tuple

from ascend_fd_tk.core.collect.base import Collector, log_collect_async_event
from ascend_fd_tk.core.collect.collector.switch_collector_a5 import SwitchCollectorA5
from ascend_fd_tk.core.collect.fetcher.podmanager_fetcher import PoDManagerFetcher
from ascend_fd_tk.core.model.bmc import BmcInfo
from ascend_fd_tk.core.model.switch import SwitchInfo
from ascend_fd_tk.utils.logger import DIAG_LOGGER


class PoDManagerCollector(Collector):
    """PoDManager 设备采集器：连接内串行切换槽位，逐槽位复用 SwitchCollectorA5 采集。"""

    def __init__(self, fetcher: PoDManagerFetcher):
        self.fetcher = fetcher
        self.host = getattr(getattr(self.fetcher, "executor", None), "host", "")

    async def get_id(self) -> str:
        return f"{self.host}_slots_{','.join(map(str, self.fetcher.slot_ids))}"

    @log_collect_async_event()
    async def collect(self) -> Tuple[List[SwitchInfo], List[BmcInfo]]:
        switch_info_list = []
        bmc_info_list = []
        for slot_id in self.fetcher.slot_ids:
            # 单槽位故障隔离：失败只跳过当前槽位，不影响其余槽位与该连接已采集结果
            try:
                await self.fetcher.switch_slot(slot_id)
                switch_info = await SwitchCollectorA5(self.fetcher).collect()
                switch_info.swi_id = self.host
                switch_info.slot_id = str(slot_id)
                switch_info_list.append(switch_info)
                # bmc 采集后续在此接入：槽位切换后复用 BmcCollector 采集该槽位 bmc，追加进 bmc_info_list
            except Exception as e:
                DIAG_LOGGER.warning("PoDManager %s 槽位 %s 采集失败，跳过该槽位：%s", self.host, slot_id, e)
        return switch_info_list, bmc_info_list
