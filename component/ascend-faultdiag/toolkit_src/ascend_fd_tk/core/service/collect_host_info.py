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

from ascend_fd_tk.core.collect.collector.host_collector import HostCollector
from ascend_fd_tk.core.context.diag_ctx import DiagCtx
from ascend_fd_tk.core.service.base import DiagService


class CollectHostsInfo(DiagService):

    def __init__(self, diag_ctx: DiagCtx):
        super().__init__(diag_ctx)

    async def run(self):
        if not self.diag_ctx.host_fetchers:
            return
        async_tasks = []
        for fetcher in self.diag_ctx.host_fetchers.values():
            async_tasks.append(HostCollector(fetcher).collect())
        for task in async_tasks:
            host_info = await task
            self.diag_ctx.cache.hosts_info.update({host_info.host_id: host_info})
