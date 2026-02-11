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

from toolkit.core.collect.fetcher.ssh_fetcher.base import SshFetcher, HiLinkType
from toolkit.core.collect.fetcher.switch_fetcher import SwitchFetcher
from toolkit.core.common import constants, diag_enum
from toolkit.core.config import chip_port_range
from toolkit.core.log_parser.base import FindResult
from toolkit.core.model.hccs import ProxyTimeoutStatis, HccsChipPortSnr, HccsSerdesDumpInfo
from toolkit.core.model.switch import InterfaceBrief, PortDownStatus
from toolkit.utils.executors import AsyncSSHExecutor, CmdTask
from toolkit.utils.table_parser import TableParser


class SwiSshFetcher(SshFetcher, SwitchFetcher):

    def __init__(self, executor: AsyncSSHExecutor):
        super().__init__(executor)

    async def init_fetcher(self):
        await self.executor.run_cmd(CmdTask("n", timeout=1))
        await self.executor.run_cmd(CmdTask("sys", timeout=1))
        await self.executor.run_cmd(CmdTask("diag", timeout=1))

    async def fetch_id(self):
        return self.executor.host

    async def fetch_serial_num(self):
        cmd_res = await self.executor.run_cmd(CmdTask("display license esn"))
        return cmd_res.stdout

    async def fetch_interface_brief(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("dis int b | no-more"))
        return cmd_res.stdout

    async def get_switch_name(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("dis cu | in sysname"))
        return cmd_res.stdout.replace("dis cu | in sysname\r\nsysname", "").strip()

    async def fetch_optical_module_info(self, interface_briefs: List[InterfaceBrief]) -> str:
        # 1.拼接命令
        all_cmd_list = []
        for interface_brief in interface_briefs:
            all_cmd_list.append(f"dis optical-module interface {interface_brief.interface} | no-more")
        all_cmd_str = "\n".join(all_cmd_list)
        # 2. 执行命令
        all_cmd_res = await self.executor.run_cmd(CmdTask(all_cmd_str))
        return all_cmd_res.stdout

    async def fetch_switch_log_info(self) -> List[FindResult]:
        return []

    async def fetch_bit_error_rate(self, interface_briefs: List[InterfaceBrief]) -> str:
        # 退出到sys
        await self.executor.run_cmd(CmdTask("quit", timeout=1))
        await self.executor.run_cmd(CmdTask("quit", timeout=1))
        all_cmd_list = []
        for interface_brief in interface_briefs:
            all_cmd_list.append(f"display interface troubleshooting {interface_brief.interface}")
        all_cmd_str = "\n".join(all_cmd_list)
        all_cmd_res = await self.executor.run_cmd(CmdTask(all_cmd_str, 15))
        # 回到诊断视图
        await self.executor.run_cmd(CmdTask("sys", timeout=1))
        await self.executor.run_cmd(CmdTask("diag", timeout=1))
        return all_cmd_res.stdout

    async def fetch_lldp_nei_brief(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("dis lldp nei b | n"))
        return cmd_res.stdout

    async def fetch_active_alarms(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("display alarm active | no-more"))
        return cmd_res.stdout

    async def fetch_history_alarms(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("display alarm history | no-more"))
        return cmd_res.stdout

    async def fetch_active_alarms_verbose(self) -> str:
        return ""

    async def fetch_history_alarms_verbose(self) -> str:
        return ""

    async def fetch_interface_info(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("display interface | no-more", timeout=10, timeout_once=0.4))
        return cmd_res.stdout

    async def fetch_datetime(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("display clock | include -"))
        return cmd_res.stdout

    async def fetch_hccs_proxy_response_statistics(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("display hccs proxy response statistics | no-more"))
        return cmd_res.stdout

    async def fetch_hccs_proxy_response_detail_interfaces(
            self, proxy_response_error_records: List[ProxyTimeoutStatis]
    ) -> str:
        all_cmd_list = []
        for record in proxy_response_error_records:
            cmd = f"display hccs proxy response detail interface {record.interface} | no-more"
            all_cmd_list.append(cmd)
        all_cmd_str = "\n".join(all_cmd_list)
        all_cmd_res = await self.executor.run_cmd(CmdTask(all_cmd_str))
        return all_cmd_res.stdout

    async def fetch_hccs_route_miss(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("display hccs route miss statistics | no-more"))
        return cmd_res.stdout

    async def fetch_link_status(self) -> str:
        all_cmd_list = []
        for i in range(constants.L1_CHIP_NUM):
            all_cmd_list.append(f'display for info enp s 1 c {i} "get port link start 0 end 47" | no-more')
        cmd_task = CmdTask("\n".join(all_cmd_list))
        cmd_res = await self.executor.run_cmd(cmd_task)
        return cmd_res.stdout

    async def fetch_port_statistic(self) -> str:
        total_res = []
        for chip_id in range(constants.L1_CHIP_NUM):
            all_cmd_list = []
            for port_id in chip_port_range.tiancheng_xpu_port_list[chip_id]:
                for proxy_module in diag_enum.HCCSProxyModule:
                    cmd = (f'display for info enp s 1 c {chip_id} "get port statistic count port {port_id} module'
                           f' {proxy_module.value} type 0 path 2" | no-more')
                    all_cmd_list.append(cmd)
            task = CmdTask("\n".join([cmd for cmd in all_cmd_list]), end_sign="", timeout=8,
                           timeout_once=0.4)
            cmd_res = await self.executor.run_cmd(task)
            total_res.append(cmd_res.stdout)
        return "".join(total_res)

    async def fetch_hccs_port_invalid_drop(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("display hccs port-invalid drop statistics | no-more"))
        return cmd_res.stdout

    async def fetch_port_credit_back_pressure_statistics(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("display qos port-credit back-pressure statistics | no-more"))
        return cmd_res.stdout

    # 以下内容可能不会会被采集, 不必分离解析部分
    async def has_hccs(self) -> bool:
        cmd_res = await self.executor.run_cmd(CmdTask("display hccs eid ub-instance 0 | no-more"))
        titles_dict = {
            "ub_instance": "Ub-instance",
            "interface": "Interface",
            "eid": "EID"
        }
        table = TableParser.parse(cmd_res.stdout, titles_dict, {}, 1)
        return len(table) > 0

    async def fetch_port_snr(self) -> str:
        for swi_chip_id in range(constants.L1_CHIP_NUM):
            all_cmd_info_list = []
            for port_id in chip_port_range.tiancheng_npu_port_list[swi_chip_id]:
                cmd = f'display for info enp s 1 c {swi_chip_id} "get port snr port-id {port_id}"'
                all_cmd_info_list.append(cmd)
            for port_id in chip_port_range.tiancheng_cpu_port_list[swi_chip_id]:
                cmd = f'display for info enp s 1 c {swi_chip_id} "get port snr port-id {port_id}"'
                all_cmd_info_list.append(cmd)
            task = CmdTask("\n".join([cmd_info for cmd_info in all_cmd_info_list]), timeout=5,
                           timeout_once=0.4)
            cmd_res = await self.executor.run_cmd(task)
            return cmd_res.stdout
        return ""

    async def fetch_hccs_map_table(self) -> str:
        task = CmdTask(f"display hccs decode and map table | in MAP_TABLE | no-more")
        cmd_res = await self.executor.run_cmd(task)
        return cmd_res.stdout

    async def fetch_interface_snr(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("dis int hilink snr | n"))
        return cmd_res.stdout

    async def fetch_transceiver_info(self):
        cmd_res = await self.executor.run_cmd(CmdTask("display interface transceiver verbose | no-more", timeout=10))
        return cmd_res.stdout

    async def fetch_interface_lane_information(self) -> str:
        cmd_res = await self.executor.run_cmd(CmdTask("display interface information | no-more"))
        return cmd_res.stdout

    async def fetch_serdes_dump_info(self, port_snr_list: List[HccsChipPortSnr]) -> List[HccsSerdesDumpInfo]:
        results = []
        cdr_los_pattern = re.compile(r"CDR_LOS = (\d+)")
        rx_dig_csr119_value_pattern = re.compile(r"RX_DIG_CSR119: (\w+)")
        for port_snr in port_snr_list:
            src_cmd = (f'display for info enp s 1 c {port_snr.swi_chip_id} "get port serdes dump-info marco-id'
                       f' {port_snr.port_id} lane-id {port_snr.lane_id} hilink ')
            cdr_cmd = src_cmd + f'{HiLinkType.SERDES_INFO.value}" | no-more | in ^LOS_STATUS.*CDR_LOS'
            ds_cmd = src_cmd + f'{HiLinkType.DS.value}" | no-more | in RX_DIG_CSR119'
            all_cmd = "\n".join([cdr_cmd, ds_cmd])
            cmd_res = await self.executor.run_cmd(CmdTask(all_cmd))
            result = HccsSerdesDumpInfo(chip_id=port_snr.swi_chip_id, port_id=port_snr.port_id,
                                        land_id=port_snr.lane_id, swi_port_id=port_snr.swi_port)
            cdr_los_search = cdr_los_pattern.search(cmd_res.stdout)
            if cdr_los_search:
                result.cdr_los = cdr_los_search.group(1)
            rx_dig_search = rx_dig_csr119_value_pattern.search(cmd_res.stdout)
            if rx_dig_search:
                result.rx_dig_CSR119 = rx_dig_search.group(1)
            if cdr_los_search or rx_dig_search:
                results.append(result)
        return results

    async def fetch_interface_port_mapping(self) -> str:
        return ""

    async def fetch_port_down_status(self) -> List[PortDownStatus]:
        return []
