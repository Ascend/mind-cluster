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

from ascend_fd_tk.core.collect.fetcher.bmc_fetcher import BmcFetcher
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.bmc.base import BmcDumpLogDataType
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.cli_output_parsed_data import CliOutputParsedData
from ascend_fd_tk.core.collect.parser.bmc_parser import BmcParser
from ascend_fd_tk.core.model.bmc import BmcSensorInfo, BmcSelInfo, BmcHealthEvents, \
    LinkDownOpticalModuleHistoryLog


class BmcDumpLogFetcher(BmcFetcher):

    def __init__(self, parse_dir: str, parsed_data: CliOutputParsedData):
        self.parse_dir = parse_dir
        self.parsed_data = parsed_data

    async def fetch_id(self) -> str:
        return self.parsed_data.fetch_data_by_name(BmcDumpLogDataType.BMC_IP.name)

    async def fetch_bmc_sn(self) -> str:
        return self.parsed_data.fetch_data_by_name(BmcDumpLogDataType.SN_NUM.name)

    async def fetch_bmc_sel_list(self) -> List[BmcSelInfo]:
        data = self.parsed_data.fetch_data_by_name(BmcDumpLogDataType.SEL_INFO.name)
        return BmcParser.trans_sel_results(data)

    async def fetch_bmc_health_events(self) -> List[BmcHealthEvents]:
        data = self.parsed_data.fetch_data_by_name(BmcDumpLogDataType.HEALTH_EVENTS.name)
        return BmcParser.trans_health_events_results(data)

    async def fetch_bmc_sensor_list(self) -> List[BmcSensorInfo]:
        data = self.parsed_data.fetch_data_by_name(BmcDumpLogDataType.SENSOR_INFO.name)
        return BmcParser.trans_sensor_results(data)

    async def fetch_bmc_date(self) -> str:
        return ""

    async def fetch_bmc_diag_info_log(self):
        pass

    async def fetch_bmc_optical_module_history_info_log(self) -> List[LinkDownOpticalModuleHistoryLog]:
        link_down_optical_module_history_log_data = self.parsed_data.fetch_data_by_name(
            BmcDumpLogDataType.OP_HISTORY_INFO_LOG.name)
        link_down_optical_module_history_logs = []
        for log_dict in link_down_optical_module_history_log_data:
            link_down_optical_module_history_logs.append(LinkDownOpticalModuleHistoryLog.from_dict(log_dict))
        return link_down_optical_module_history_logs
