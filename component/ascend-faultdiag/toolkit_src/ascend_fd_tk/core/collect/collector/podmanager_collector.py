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
from typing import List, Tuple

from ascend_fd_tk.core.collect.base import log_collect_async_event
from ascend_fd_tk.core.collect.collector.switch_collector_a5 import SwitchCollectorA5
from ascend_fd_tk.core.collect.fetcher.podmanager_fetcher import PoDManagerFetcher
from ascend_fd_tk.core.common import constants
from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.model.bmc import BmcInfo
from ascend_fd_tk.core.model.switch import SwitchInfo
from ascend_fd_tk.utils.logger import DIAG_LOGGER


class PoDManagerCollector(SwitchCollectorA5):
    """PoDManager 设备采集器：连接内串行切换槽位采集。

    继承 SwitchCollectorA5 复用其采集方法（coll_*）与解析器；
    SFU 槽位（sfu_slot_ids）走父类 collect() 全量采集；
    NPU 槽位（npu_slot_ids）按需采集基础信息并批量查询 switch 芯片 QoS Credit。
    """

    def __init__(self, fetcher: PoDManagerFetcher):
        super().__init__(fetcher)
        self.host = getattr(getattr(self.fetcher, "executor", None), "host", "")

    async def get_id(self) -> str:
        slot_ids = self.fetcher.sfu_slot_ids + self.fetcher.npu_slot_ids
        return f"{self.host}_slots_{','.join(slot_ids)}"

    @log_collect_async_event()
    async def collect(self) -> Tuple[List[SwitchInfo], List[BmcInfo]]:
        switch_info_list = await self.collect_sfu()
        switch_info_list.extend(await self.collect_npu())
        return switch_info_list, []

    async def collect_sfu(self) -> List[SwitchInfo]:
        """SFU 槽位采集：串行切换槽位，复用父类 collect 全量采集交换机常规信息。"""
        switch_info_list = []
        for slot_id in self.fetcher.sfu_slot_ids:
            # 单槽位故障隔离：失败只跳过当前槽位，不影响其余槽位与该连接已采集结果
            try:
                await self.fetcher.switch_slot(slot_id)
                # 显式调用父类 collect（不能用 self.collect，会递归回本类编排入口）
                switch_info = await super().collect()
                switch_info.swi_id = self.host
                switch_info.slot_id = slot_id
                switch_info_list.append(switch_info)
                # bmc 采集后续在此接入：槽位切换后复用 BmcCollector 采集该槽位 bmc，追加进 bmc_info_list
            except Exception as e:
                DIAG_LOGGER.warning("PoDManager %s 槽位 %s 采集失败，跳过该槽位：%s", self.host, slot_id, e)
                await self.fetcher.last_quit()
        return switch_info_list

    async def collect_npu(self) -> List[SwitchInfo]:
        """NPU 槽位 信息采集：串行跳转槽位，复用 A5 采集器采集基础信息并批量查询 16 个芯片 QoS Credit。"""
        switch_info_list = []
        for slot_id in self.fetcher.npu_slot_ids:
            # 单槽位故障隔离：失败只跳过当前槽位，不影响其余槽位与该连接已采集结果
            try:
                await self.fetcher.switch_slot(slot_id)
                await self.fetcher.init_fetcher()
                # 基础信息（name/sn/datetime/interface_brief）复用父类 coll 方法
                switch_name = await self.fetcher.get_switch_name()
                sn = await self.coll_serial_num()
                date_time = await self.coll_datetime()
                interface_briefs = await self.coll_interface_brief()
                qos_credit_infos = await self.collect_qos_credit(slot_id)
                switch_info_list.append(
                    SwitchInfo(
                        name=switch_name,
                        swi_id=self.host,
                        sn=sn,
                        slot_id=slot_id,
                        interface_briefs=interface_briefs,
                        date_time=date_time,
                        generation=NpuType.A5.value,
                        qos_credit_infos=qos_credit_infos,
                    )
                )
            except Exception as e:
                DIAG_LOGGER.warning("PoDManager %s NPU 槽位 %s credit 采集失败，跳过该槽位：%s", self.host, slot_id, e)
            finally:
                await self.fetcher.last_quit()
        return switch_info_list

    async def collect_qos_credit(self, slot_id: str):
        cmd_res = await self.fetcher.fetch_qos_credit(slot_id, range(constants.POD_MANAGER_NPU_SLOT_CHIP_NUM))
        return self.parser.parse_qos_credit(cmd_res)
