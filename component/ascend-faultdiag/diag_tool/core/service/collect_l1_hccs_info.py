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

from diag_tool.core.collect.collector.hccs_collector import HccsCollector
from diag_tool.core.context.diag_ctx import DiagCtx
from diag_tool.core.service.base import DiagService
from diag_tool.utils import logger

_DIAG_LOGGER = logger.DIAG_LOGGER


class CollectL1HccsInfo(DiagService):

    def __init__(self, diag_ctx: DiagCtx):
        super().__init__(diag_ctx)

    async def run(self):
        if not self.diag_ctx.switch_fetchers:
            return
        async_tasks = []
        for fetcher in self.diag_ctx.switch_fetchers.values():
            async_tasks.append(HccsCollector(fetcher).collect())
        hccs_info_list = await asyncio.gather(*async_tasks)
        for switch_id, hccs_info in zip(self.diag_ctx.switch_fetchers.keys(), hccs_info_list):
            switch_info = self.diag_ctx.cache.swis_info.get(switch_id)
            if not switch_info:
                _DIAG_LOGGER.warning(f"未收集到交换机{switch_id}信息")
                continue
            switch_info.hccs_info = hccs_info
