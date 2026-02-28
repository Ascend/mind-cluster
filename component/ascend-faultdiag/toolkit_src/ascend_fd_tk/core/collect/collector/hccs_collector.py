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

from ascend_fd_tk.core.collect.base import Collector, log_collect_async_event
from ascend_fd_tk.core.collect.fetcher.switch_fetcher import SwitchFetcher
from ascend_fd_tk.core.collect.parser.hccs_parser import HccsParser
from ascend_fd_tk.core.model.hccs import HccsInfo, ProxyTimeoutStatis, InterfaceProxyResponseDetail, HccsRouteMiss, \
    PortLinkStatusRecord, HccsPortStatistic, HccsPortInvalidDrop, PortCreditBackPressure, InterfaceSnr, LaneInfo, \
    HccsMapTable, HccsChipPortSnr


class HccsCollector(Collector):

    def __init__(self, fetcher: SwitchFetcher):
        self.fetcher = fetcher
        self.parser = HccsParser()

    @log_collect_async_event()
    async def collect(self) -> HccsInfo:
        if not await self.fetcher.has_hccs():
            return None
        proxy_response_error_records = await self.coll_hccs_proxy_response_statistics()
        proxy_response_detail_interfaces = await self.coll_hccs_proxy_response_detail_interfaces(
            proxy_response_error_records)
        hccs_route_miss = await self.coll_hccs_route_miss()
        link_status_records = await self.coll_link_status()
        hccs_port_statistic = await self.coll_port_statistic()
        hccs_port_invalid_drop = await self.coll_hccs_port_invalid_drop()
        port_credit_back_pressure_statistics = await self.coll_port_credit_back_pressure_statistics()
        hccs_chip_port_snr_list = await self.coll_hccs_port_snr()
        hccs_map_table_list = await self.coll_hccs_map_table()
        interface_snr_list = await self.coll_interface_snr()
        lane_info_list = await self.coll_if_lane_infos()
        serdes_dump_info_list = await self.fetcher.fetch_serdes_dump_info(hccs_chip_port_snr_list)
        return HccsInfo(
            hccs_route_miss=hccs_route_miss,
            proxy_timeout_statis=proxy_response_error_records,
            proxy_response_detail_interfaces=proxy_response_detail_interfaces,
            link_status_records=link_status_records,
            hccs_port_statistic=hccs_port_statistic,
            hccs_port_invalid_drop=hccs_port_invalid_drop,
            port_credit_back_pressure_statistics=port_credit_back_pressure_statistics,
            hccs_chip_port_snr_list=hccs_chip_port_snr_list,
            hccs_map_table_list=hccs_map_table_list,
            interface_snr_list=interface_snr_list,
            lane_info_list=lane_info_list,
            serdes_dump_info_list=serdes_dump_info_list
        )

    async def get_id(self) -> str:
        return await self.fetcher.fetch_id()

    async def coll_interface_snr(self) -> List[InterfaceSnr]:
        cmd_res = await self.fetcher.fetch_interface_snr()
        return self.parser.parse_interface_snr(cmd_res)

    async def coll_hccs_proxy_response_statistics(self) -> List[ProxyTimeoutStatis]:
        cmd_res = await self.fetcher.fetch_hccs_proxy_response_statistics()
        return self.parser.parse_hccs_proxy_response_statistics(cmd_res)

    async def coll_hccs_proxy_response_detail_interfaces(
            self, proxy_response_error_records: List[ProxyTimeoutStatis]
    ) -> List[InterfaceProxyResponseDetail]:
        cmd_res = await self.fetcher.fetch_hccs_proxy_response_detail_interfaces(proxy_response_error_records)
        return self.parser.parse_hccs_proxy_response_detail_interfaces(cmd_res, proxy_response_error_records)

    async def coll_hccs_route_miss(self) -> List[HccsRouteMiss]:
        cmd_res = await self.fetcher.fetch_hccs_route_miss()
        return self.parser.parse_hccs_route_miss(cmd_res)

    async def coll_link_status(self) -> List[PortLinkStatusRecord]:
        cmd_res = await self.fetcher.fetch_link_status()
        return self.parser.parse_link_status(cmd_res)

    async def coll_port_statistic(self) -> List[HccsPortStatistic]:
        cmd_res = await self.fetcher.fetch_port_statistic()
        return self.parser.parse_port_statistics_chip_info(cmd_res)

    async def coll_hccs_port_invalid_drop(self) -> List[HccsPortInvalidDrop]:
        cmd_res = await self.fetcher.fetch_hccs_port_invalid_drop()
        return self.parser.parse_hccs_port_invalid_drop(cmd_res)

    async def coll_port_credit_back_pressure_statistics(self) -> List[PortCreditBackPressure]:
        cmd_res = await self.fetcher.fetch_port_credit_back_pressure_statistics()
        return self.parser.parse_port_credit_back_pressure_statistics(cmd_res)

    async def coll_if_lane_infos(self) -> List[LaneInfo]:
        cmd_res = await self.fetcher.fetch_interface_lane_information()
        return self.parser.parse_if_lane_info(cmd_res)

    async def coll_hccs_map_table(self) -> List[HccsMapTable]:
        cmd_res = await self.fetcher.fetch_hccs_map_table()
        return self.parser.parse_hccs_map_table(cmd_res)

    async def coll_hccs_port_snr(self) -> List[HccsChipPortSnr]:
        cmd_res = await self.fetcher.fetch_port_snr()
        port_snr_table_list = self.parser.parse_hccs_port_snr_table(cmd_res)

        switch_log_info = await self.fetcher.fetch_port_down_status()
        port_snr_line_list = self.parser.parse_hccs_port_snr_line(switch_log_info)
        return port_snr_table_list + port_snr_line_list
