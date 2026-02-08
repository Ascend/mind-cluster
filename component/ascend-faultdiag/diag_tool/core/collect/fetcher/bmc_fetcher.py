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

import abc
from typing import List

from diag_tool.core.collect.fetcher.base import Fetcher
from diag_tool.core.model.bmc import BmcSensorInfo, BmcSelInfo, BmcHealthEvents, \
    LinkDownOpticalModuleHistoryLog


class BmcFetcher(Fetcher):

    @abc.abstractmethod
    async def fetch_bmc_sn(self) -> str:
        pass

    @abc.abstractmethod
    async def fetch_bmc_sel_list(self) -> List[BmcSelInfo]:
        pass

    @abc.abstractmethod
    async def fetch_bmc_health_events(self) -> List[BmcHealthEvents]:
        pass

    @abc.abstractmethod
    async def fetch_bmc_sensor_list(self) -> List[BmcSensorInfo]:
        pass

    @abc.abstractmethod
    async def fetch_bmc_date(self) -> str:
        pass

    @abc.abstractmethod
    async def fetch_bmc_diag_info_log(self):
        pass

    @abc.abstractmethod
    async def fetch_bmc_optical_module_history_info_log(self) -> List[LinkDownOpticalModuleHistoryLog]:
        pass