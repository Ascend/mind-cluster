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

from typing import List

from diag_tool.core.collect.base import Collector, log_collect_async_event
from diag_tool.core.collect.fetcher.bmc_fetcher import BmcFetcher
from diag_tool.core.model.bmc import BmcInfo, BmcSensorInfo, BmcHealthEvents, BmcSelInfo, \
    LinkDownOpticalModuleHistoryLog


class BmcCollector(Collector):

    def __init__(self, fetcher: BmcFetcher):
        self.fetcher = fetcher

    @log_collect_async_event()
    async def collect(self) -> BmcInfo:
        bmc_id = await self.fetcher.fetch_id()
        sn_num = await self.coll_bmc_sn()
        bmc_date = await self.coll_bnc_date()
        sensor_info_list = await self.coll_bmc_sensor_list()
        sel_info_list = await self.coll_bmc_sel_list()
        health_events = await self.coll_bmc_health_events()
        link_down_optical_module_history_logs = await self.coll_optical_module_history_logs()
        bmc_info = BmcInfo(bmc_id, sn_num,
                           bmc_sel_list=sel_info_list,
                           sensor_info_list=sensor_info_list,
                           health_events=health_events,
                           link_down_optical_module_history_logs=link_down_optical_module_history_logs,
                           bmc_date=bmc_date)
        return bmc_info

    async def get_id(self) -> str:
        return await self.fetcher.fetch_id()

    async def coll_bmc_sensor_list(self) -> List[BmcSensorInfo]:
        return await self.fetcher.fetch_bmc_sensor_list()

    async def coll_bmc_sn(self) -> str:
        return await self.fetcher.fetch_bmc_sn()

    async def coll_bmc_health_events(self) -> List[BmcHealthEvents]:
        return await self.fetcher.fetch_bmc_health_events()

    async def coll_bmc_sel_list(self) -> List[BmcSelInfo]:
        return await self.fetcher.fetch_bmc_sel_list()

    async def coll_optical_module_history_logs(self) -> List[LinkDownOpticalModuleHistoryLog]:
        return await self.fetcher.fetch_bmc_optical_module_history_info_log()

    async def coll_bnc_date(self) -> str:
        return await self.fetcher.fetch_bmc_date()
