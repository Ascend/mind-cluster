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

from ascend_fd_tk.core.collect.base import log_collect_async_event, Collector
from ascend_fd_tk.core.context.register import register_host_collector
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.host_ssh_fetcher_a5 import HostSshFetcherA5
from ascend_fd_tk.core.collect.parser.host_parser_a5 import HostParserA5
from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.model.host_a5 import HostInfoA5, OpticalTopHeadline, NpuChipInfoA5, NICInfoA5


@register_host_collector(NpuType.A5)
class HostCollectorA5(Collector):
    """A5 主机采集器，独立实现，不继承 A3。"""

    def __init__(self, fetcher: HostSshFetcherA5):
        self.fetcher = fetcher
        self.parser = HostParserA5()

    @log_collect_async_event()
    async def collect(self) -> HostInfoA5:
        # pylint: disable=R0801
        host_id = await self.get_id()
        host_name = await self.fetcher.fetch_hostname()
        msnpureport_log = await self.fetcher.fetch_msnpureport_log()
        npu_mapping = await self.fetcher.fetch_npu_mapping()
        npu_type = await self.collect_npu_type()
        sn_num = await self.fetcher.fetch_sn_num()
        npu_chip_info = {}
        for npu_id in npu_mapping.keys():
            optical_top_headline = await self.collect_optical_top_headline(npu_id)
            optical_info_dict = {}
            for top_info in optical_top_headline:
                hccn_optical_info = await self.collect_optical_info(npu_id, top_info.optical_id)
                optical_info_dict.update({top_info.optical_id: hccn_optical_info})
            npu_chip_info[npu_id] = NpuChipInfoA5(
                npu_id=npu_id,
                npu_type=npu_type,
                hccn_optical_info=optical_info_dict,
            )

        nic_info_list = await self.collect_nic_info()

        host_info = HostInfoA5(
            host_id,
            sn_num,
            hostname=host_name,
            msnpureport_log=msnpureport_log,
            npu_chip_info=npu_chip_info,
            nic_info_list=nic_info_list,
        )
        return host_info

    async def get_id(self) -> str:
        return await self.fetcher.fetch_id()

    async def collect_npu_type(self) -> str:
        recv = await self.fetcher.fetch_npu_type()
        return self.parser.parse_npu_type(recv)

    async def collect_optical_top_headline(self, npu_id) -> List[OpticalTopHeadline]:
        recv = await self.fetcher.fetch_optical_top_headline(npu_id)
        return self.parser.parse_optical_top_headline(recv)

    async def collect_optical_info(self, npu_id, optical_id):
        recv = await self.fetcher.fetch_optical_info_a5(npu_id, optical_id)
        return self.parser.parse_optical_info_a5(recv, npu_id, optical_id)

    async def collect_nic_info(self) -> List[NICInfoA5]:
        """采集所有网卡信息：先查网卡列表，再查每张网卡端口数，最后遍历每个端口采集 SFP 信息"""
        nic_info_list = []
        card_names_recv = await self.fetcher.fetch_nic_info()
        card_names = self.parser.parse_nic_card_names(card_names_recv)
        if not card_names:
            return nic_info_list
        for card_name in card_names:
            port_num_recv = await self.fetcher.fetch_nic_port_num(card_name)
            port_num = self.parser.parse_nic_port_num(port_num_recv)
            nic_info = NICInfoA5(card_name=card_name, port_num=port_num)
            if port_num and port_num.isdigit():
                for port_id in range(int(port_num)):
                    port_id = str(port_id)
                    sfp_recv = await self.fetcher.fetch_nic_sfp_info(card_name, port_id)
                    port_lane_info = self.parser.parse_nic_sfp_info(sfp_recv, port_id)
                    nic_info.port_lane_info_list.append(port_lane_info)
            nic_info_list.append(nic_info)
        return nic_info_list
