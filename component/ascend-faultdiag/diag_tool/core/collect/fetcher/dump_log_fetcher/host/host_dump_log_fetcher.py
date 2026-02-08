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

import re
from typing import List

from diag_tool.core.collect.collect_config import ToolLogCollectionDataType
from diag_tool.core.collect.fetcher.dump_log_fetcher.cli_output_parsed_data import CliOutputParsedData
from diag_tool.core.collect.fetcher.host_fetcher import HostFetcher
from diag_tool.core.log_parser.base import FindResult


class HostDumpLogFetcher(HostFetcher):

    def __init__(self, parse_dir: str, parsed_data: CliOutputParsedData):
        self.parse_dir = parse_dir
        self.parsed_data = parsed_data

    async def fetch_hostname(self) -> str:
        return await self.fetch_id()

    async def fetch_optical_loopback_enable(self, npu_id, model) -> bool:
        return False

    async def fetch_id(self) -> str:
        return self.parsed_data.fetch_data_by_name(ToolLogCollectionDataType.HOST_ID.name)

    async def fetch_npu_mapping(self) -> dict:
        return await super().fetch_npu_mapping()

    async def fetch_optical_info(self, chip_phy_id) -> str:
        return self.parsed_data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.OPTICAL, chip_phy_id)

    async def fetch_link_stat_info(self, chip_phy_id) -> str:
        return self.parsed_data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.LINK_STAT, chip_phy_id)

    async def fetch_stat_info(self, chip_phy_id) -> str:
        return self.parsed_data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.STAT, chip_phy_id)

    async def fetch_lldp_info(self, chip_phy_id) -> str:
        return self.parsed_data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.LLDP, chip_phy_id)

    async def fetch_npu_type(self) -> str:
        return self.parsed_data.fetch_data_by_name(ToolLogCollectionDataType.NPU_TYPE.name)

    async def fetch_sn_num(self) -> str:
        data = self.parsed_data.fetch_data([ToolLogCollectionDataType.SN.name])
        if data:
            match = re.search(r'Serial Number: ([0-9a-zA-Z]{1,50})', data)
            if match:
                return match.group(1)
        return ""

    async def fetch_hccs_info(self, npu_id, chip_id) -> str:
        return self.parsed_data.fetch_data_by_device_chip_id(ToolLogCollectionDataType.HCCS, npu_id, chip_id)

    async def fetch_spod_info(self, npu_id, chip_id) -> str:
        return self.parsed_data.fetch_data_by_device_chip_id(ToolLogCollectionDataType.SPOD_INFO, npu_id, chip_id)

    async def fetch_msnpureport_log(self) -> List[FindResult]:
        return self.parsed_data.fetch_data([ToolLogCollectionDataType.MS_NPU_REPORT.name])

    async def fetch_roce_speed(self, chip_phy_id) -> str:
        return self.parsed_data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.SPEED, chip_phy_id)

    async def fetch_roce_duplex(self, chip_phy_id) -> str:
        return ""

    async def fetch_hccn_tool_net_health(self, chip_phy_id) -> str:
        return self.parsed_data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.NET_HEALTH, chip_phy_id)

    async def fetch_hccn_tool_cdr(self, chip_phy_id) -> str:
        return self.parsed_data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.CDR_SNR, chip_phy_id)

    async def fetch_hccn_dfx_cfg(self, chip_phy_id) -> str:
        return self.parsed_data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.DFX_CFG, chip_phy_id)

    async def fetch_hccn_tool_link_status(self, chip_phy_id) -> str:
        return self.parsed_data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.HCCN_LINK_STATUS, chip_phy_id)

