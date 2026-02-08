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

from diag_tool.core.collect.fetcher.bmc_fetcher import BmcFetcher
from diag_tool.core.common.json_obj import JsonObj
from diag_tool.utils import logger

LOGGER = logger.DIAG_LOGGER


class BmcLogCollector:

    def __init__(self, fetcher: BmcFetcher):
        self.fetcher = fetcher

    # @log_collect_event(error_msg="收集bmc日志失败", raise_exception=False)
    async def collect(self) -> JsonObj:
        return await self.fetcher.fetch_bmc_diag_info_log()
