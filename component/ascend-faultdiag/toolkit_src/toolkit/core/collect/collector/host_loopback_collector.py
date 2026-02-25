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
from typing import List

from toolkit.core.collect.base import log_collect_async_event
from toolkit.core.collect.collector.host_collector import HostCollector
from toolkit.core.collect.fetcher.host_fetcher import HostFetcher
from toolkit.core.common.diag_enum import OpticalLoopbackMode
from toolkit.core.model.host import HostInfo, NpuChipLoopBackInfo


class HostLoopbackCollector(HostCollector):

    def __init__(self, fetcher: HostFetcher):
        super().__init__(fetcher)
        self.server_superpod_id = ""
        self.server_index = ""
        self.same_sn_npu_ids_map = {}
        self.same_npu_loopback_map = {}

    @log_collect_async_event()
    async def collect(self) -> HostInfo:
        host_id = await self.fetcher.fetch_id()
        npu_mapping = await self.fetcher.fetch_npu_mapping()
        sn_num = await self.fetcher.fetch_sn_num()
        server_superpod_id = None
        server_index = None
        for npu_id, chips in npu_mapping.items():
            for chip_id, chip_phy_id in chips.items():
                hccn_optical_info = await self.collect_optical_info(chip_phy_id)
                hccn_lldp_info = await self.collect_lldp_info(chip_phy_id)
                spod_info = await self.collect_spod_info(npu_id, chip_id)
                if not self.server_superpod_id and spod_info:
                    self.server_superpod_id = spod_info.super_pod_id
                if not self.server_index and spod_info:
                    self.server_index = spod_info.server_index
                loopback_info = NpuChipLoopBackInfo(hccn_lldp_info=hccn_lldp_info,
                                                    hccn_optical_info=hccn_optical_info,
                                                    spod_info=spod_info,
                                                    npu_id=npu_id, chip_id=chip_id, chip_phy_id=chip_phy_id)
                self.same_npu_loopback_map.setdefault(npu_id, []).append(loopback_info)
                optical_sn = loopback_info.hccn_optical_info.vendor_serial_number
                if optical_sn:
                    self.same_sn_npu_ids_map.setdefault(optical_sn, []).append(loopback_info.npu_id)

        # 去重
        same_sn_npu_ids_tuple = {tuple(sorted(npu_ids)) for npu_ids in self.same_sn_npu_ids_map.values()}
        same_sn_npu_ids_list = [list(npu_ids_tuple) for npu_ids_tuple in same_sn_npu_ids_tuple]

        for same_sn_npu_ids in same_sn_npu_ids_list:
            await self.enable_optical_loopback(same_sn_npu_ids)

        # 关闭环回
        for npu_id in npu_mapping.keys():
            await self.collect_optical_loopback_enable(npu_id, OpticalLoopbackMode.NO_LOOPBACK.value)

        host_info = HostInfo(host_id, sn_num, server_superpod_id=server_superpod_id, server_index=server_index,
                             loopback_info_list=[loopback_info for sublist in self.same_npu_loopback_map.values() for
                                                 loopback_info in sublist])
        return host_info

    async def enable_optical_loopback(self, same_sn_npu_ids: List[str]):
        await self.enable_optical_loopback_by_model(same_sn_npu_ids, OpticalLoopbackMode.HOST_SIDE_INPUT.value)
        await self.enable_optical_loopback_by_model(same_sn_npu_ids, OpticalLoopbackMode.MEDIA_SIDE_OUTPUT.value)

    async def enable_optical_loopback_by_model(self, same_sn_npu_ids: List[str], model):
        if not same_sn_npu_ids:
            return
        status = await self.collect_optical_loopback_enable(same_sn_npu_ids[0], model)
        if not status:
            return
        await asyncio.sleep(10)
        for npu_id in same_sn_npu_ids:
            for loopback_info in self.same_npu_loopback_map.get(npu_id, []):
                if OpticalLoopbackMode.HOST_SIDE_INPUT.value == model:
                    loopback_info.host_input_enable = True
                    loopback_info.host_input_link_stat = await self.collect_link_stat_info(loopback_info.chip_phy_id)
                if OpticalLoopbackMode.MEDIA_SIDE_OUTPUT.value == model:
                    loopback_info.media_output_enable = True
                    loopback_info.media_output_link_stat = await self.collect_link_stat_info(loopback_info.chip_phy_id)

    async def collect_optical_loopback_enable(self, npu_id, model) -> bool:
        recv = await self.fetcher.fetch_optical_loopback_enable(npu_id, model)
        return self.parser.parse_optical_loopback_enable(recv)
