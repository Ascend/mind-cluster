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

from ascend_fd_tk.core.collect.collector.host_collector_factory import create_host_collector
from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.service.base import DiagService


class CollectHostsInfo(DiagService):
    async def run(self):
        if not self.diag_ctx.host_fetchers:
            return
        async_tasks = []
        fetchers = list(self.diag_ctx.host_fetchers.values())
        for fetcher in fetchers:
            # 按 fetcher 携带的代际选择对应的主机采集器（A3/A5）
            collector = create_host_collector(fetcher)
            async_tasks.append(collector.collect())
        for fetcher, task in zip(fetchers, async_tasks):
            host_info = await task
            self.diag_ctx.location_config.enrich_host_info(host_info)
            # 代际写入 HostInfo，随 JSON 落盘；诊断阶段 LoadCache 读回后聚合到 cache
            generation = getattr(fetcher, 'chip_generation', NpuType.A3)
            host_info.chip_generation = generation.value if generation else NpuType.A3.value
            self.diag_ctx.cache.hosts_info.update({host_info.host_id: host_info})
