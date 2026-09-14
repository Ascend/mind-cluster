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
import asyncio

from ascend_fd_tk.core.collect.collector.podmanager_collector import PoDManagerCollector
from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.service.base import DiagService


class CollectPoDManagerInfo(DiagService):
    """独立采集 PoDManager 设备信息：串行切换槽位采集后增量合并进 cache。"""

    async def run(self):
        if not self.diag_ctx.pod_manager_fetchers:
            return
        # PoDManager 登录即 A5 代际（superpod），写入集群级代际，后续阈值按 A5 默认值应用
        self.diag_ctx.cache.generation = NpuType.A5.value
        collect_results = await asyncio.gather(
            *(PoDManagerCollector(fetcher).collect() for fetcher in self.diag_ctx.pod_manager_fetchers.values())
        )
        for switch_info_list, bmc_info_list in collect_results:
            for switch_info in switch_info_list:
                self.diag_ctx.location_config.enrich_switch_info(switch_info)
                self.diag_ctx.cache.swis_info.update({f"{switch_info.swi_id}_{switch_info.slot_id}": switch_info})
            for bmc_info in bmc_info_list:
                self.diag_ctx.cache.bmcs_info.update({f"{bmc_info.bmc_id}_{bmc_info.slot_id}": bmc_info})
