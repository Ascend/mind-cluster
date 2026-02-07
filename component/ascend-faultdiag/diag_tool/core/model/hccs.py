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
from collections import defaultdict
from typing import List

from diag_tool.core.common.json_obj import JsonObj
from diag_tool.utils.helpers import to_int

NORMAL_LANE_WIDTH = 4


class HccsRouteMiss(JsonObj):

    def __init__(self, interface="", rp_direction_miss="0", lp_direction_miss="0", nc_direction_miss="0"):
        self.interface = interface
        self.rp_direction_miss = to_int(rp_direction_miss)
        self.lp_direction_miss = to_int(lp_direction_miss)
        self.nc_direction_miss = to_int(nc_direction_miss)

    def is_lp_direction_miss(self) -> bool:
        return self.lp_direction_miss > 0


class ProxyTimeoutStatis(JsonObj):

    def __init__(self, interface="", rp_miss="0", rp_rx="0", rp_tx="0", lp_miss="0", lp_rx="0", lp_tx="0"):
        self.interface = interface
        self.rp_miss = to_int(rp_miss)
        self.rp_rx = to_int(rp_rx)
        self.rp_tx = to_int(rp_tx)
        self.lp_miss = to_int(lp_miss)
        self.lp_rx = to_int(lp_rx)
        self.lp_tx = to_int(lp_tx)

    # 超时代答会导致SDMA报错，触发HCCL算子重执行，建议排除代答次数大于1的端口
    # lp_tx场景不常见，暂不分析
    def is_rx_timeout_happend(self) -> bool:
        return self.rp_rx > 0 or self.lp_rx > 0

    def is_rp_tx_timeout_happend(self) -> bool:
        return self.rp_tx > 0

    def is_lp_tx_timeout_happend(self) -> bool:
        return self.lp_tx > 0


class PortLinkStatusRecord(JsonObj):
    _TIME_PATTERN = re.compile(r"\[[\d:\- ]{0,10}]")

    def __init__(self, chip="", port="", swi_port="", index="", record=""):
        self.chip = chip
        self.port = port
        self.swi_port = swi_port
        self.index = index
        self.record = record
        self.record_time = self._parse_record_time(record)

    @classmethod
    def _parse_record_time(cls, record) -> str:
        search = cls._TIME_PATTERN.search(record)
        if search:
            return search.group(0)
        return ""


class InterfaceProxyResponseDetail(JsonObj):

    def __init__(self, interface="", proxy_type="", response_type="", address="", collect_time=""):
        self.interface = interface
        self.proxy_type = proxy_type
        self.response_type = response_type
        self.address = address
        self.collect_time = collect_time


class HccsPortStatistic(JsonObj):

    def __init__(self, chip="", port="", proxy_module="", dfx_state_name="", dfx_result=""):
        self.chip = chip
        self.port = port
        self.proxy_module = proxy_module
        self.dfx_state_name = dfx_state_name
        self.dfx_result = dfx_result


class HccsPortInvalidDrop(JsonObj):

    def __init__(self, ub_instance="", link_group="", rplp="0", nc="0"):
        self.ub_instance = ub_instance
        self.link_group = link_group
        self.rplp = to_int(rplp)
        self.nc = to_int(nc)


class PortCreditBackPressure(JsonObj):
    def __init__(self, interface="", vl="", back_pressure_counts="0", last_time=""):
        self.interface = interface
        self.vl = vl
        self.back_pressure_counts = to_int(back_pressure_counts)
        self.last_time = last_time


class HccsMapTable(JsonObj):

    def __init__(self, port="", start_addr="", end_addr=""):
        self.port = port
        self.start_addr = start_addr
        self.end_addr = end_addr


class LaneSnr(JsonObj):
    def __init__(self, lane_name: str = "", snr_value: str = ""):
        self.lane_name = lane_name
        self.snr_value = snr_value

    def __str__(self) -> str:
        return f"{self.lane_name}:{self.snr_value}"


class InterfaceSnr(JsonObj):
    def __init__(self, interface_name="", abnormal_lane_snr: List[LaneSnr] = None, threshold: int = -1):
        self.interface_name = interface_name
        self.abnormal_lane_snr = abnormal_lane_snr or []

    def lane_snr_to_str(self) -> str:
        return str([str(lane_snr) for lane_snr in self.abnormal_lane_snr])


class HccsChipPortSnr(JsonObj):

    def __init__(self, swi_chip_id="", port_id="", swi_port="", lane_id="", snr="", xpu=""):
        self.swi_chip_id = swi_chip_id
        self.port_id = port_id
        self.swi_port = swi_port
        self.lane_id = lane_id
        self.snr = snr
        self.xpu = xpu


class LinkDownDetail(JsonObj):
    def __init__(self, collect_time="", down_start_time="", down_end_time=""):
        self.collect_time = collect_time
        self.down_start_time = down_start_time
        self.down_end_time = down_end_time


class LCNEInfo(JsonObj):
    def __init__(self, interface="", port_id="", npu_id="", die_id="", server_id="", chip_id="", xpu="",
                 link_status: List[PortLinkStatusRecord] = None,
                 rp_direction_miss=0, lp_direction_miss=0, nc_direction_miss=0,
                 rp_id_using_cnt=0, lp_id_using_cnt=0, cur_pkt=0):
        self.interface = interface
        self.port_id = port_id
        self.die_id = die_id
        self.npu_id = npu_id
        self.chip_id = chip_id
        self.server_id = server_id
        self.xpu = xpu
        # 端口状态信息
        self.link_status: List[PortLinkStatusRecord] = link_status
        # 路由miss
        self.rp_direction_miss = rp_direction_miss
        self.lp_direction_miss = lp_direction_miss
        self.nc_direction_miss = nc_direction_miss
        # 窝包
        self.rp_id_using_cnt = rp_id_using_cnt
        self.lp_id_using_cnt = lp_id_using_cnt
        self.cur_pkt = cur_pkt

    def __str__(self):
        return str(
            f"server-{self.server_id} {self.xpu}-{self.npu_id} die-{self.die_id} port-{self.port_id} {self.interface}"
        )

    def get_first_link_down_time(self) -> str:
        if not self.link_status:
            return ""
        record = self.link_status[0].record
        if not record or "link down" not in record:
            return ""
        start_index = record.find("[")
        end_index = record.find("]")
        if start_index == -1 or end_index == -1:
            return ""
        return record[start_index + 1:end_index]

    def is_lp_route_miss(self) -> bool:
        """
        检查端口lp方向是否路由miss
        :return: True or False
        """
        return self.lp_direction_miss > 0

    def is_lp_pack_block(self) -> bool:
        """
        检查端口lp方向是否窝包
        :return: True or False
        """
        return self.lp_id_using_cnt > 0

    def is_rp_pack_block(self) -> bool:
        """
        检查端口rp方向是否窝包
        :return: True or False
        """
        return self.rp_id_using_cnt > 0

    def is_lane_error(self) -> bool:
        """
        检查降lane告警
        :return: True or False
        """
        lane_width = self.get_lane_width()
        if lane_width and lane_width < NORMAL_LANE_WIDTH:
            return True
        return False

    def get_lane_width(self) -> int:
        if not self.link_status:
            return NORMAL_LANE_WIDTH
        for link_status in self.link_status:
            if "up" in link_status.record:
                parts = link_status.record.split("width:")
                if len(parts) == 2:
                    return int(parts[1])
        return NORMAL_LANE_WIDTH

    def is_voq_pack_block(self) -> bool:
        return self.cur_pkt > 0


class LaneInfo(JsonObj):

    def __init__(self, if_name="", running_lane_num="", real_lane_num=""):
        self.if_name = if_name
        self.running_lane_num = to_int(running_lane_num)
        self.real_lane_num = to_int(real_lane_num)


class HccsSerdesDumpInfo(JsonObj):

    def __init__(self, chip_id="", port_id="", land_id="", swi_port_id="", cdr_los=False, csr119_data=""):
        self.chip_id = chip_id
        self.port_id = port_id
        self.land_id = land_id
        self.swi_port_id = swi_port_id
        self.cdr_los = cdr_los
        self.csr119_data = csr119_data


class HccsInfo(JsonObj):

    def __init__(self, hccs_route_miss: List[HccsRouteMiss] = None,
                 proxy_timeout_statis: List[ProxyTimeoutStatis] = None,
                 proxy_response_detail_interfaces: List[InterfaceProxyResponseDetail] = None,
                 link_status_records: List[PortLinkStatusRecord] = None,
                 hccs_port_statistic: List[HccsPortStatistic] = None,
                 hccs_port_invalid_drop: List[HccsPortInvalidDrop] = None,
                 port_credit_back_pressure_statistics: List[PortCreditBackPressure] = None,
                 hccs_chip_port_snr_list: List[HccsChipPortSnr] = None,
                 hccs_map_table_list: List[HccsMapTable] = None,
                 interface_snr_list: List[InterfaceSnr] = None,
                 lane_info_list: List[LaneInfo] = None,
                 serdes_dump_info_list: List[HccsSerdesDumpInfo] = None, ):
        self.hccs_route_miss = hccs_route_miss or []
        self.proxy_timeout_statis = proxy_timeout_statis or []
        self.proxy_response_detail_interfaces = proxy_response_detail_interfaces or []
        self.link_status_records = link_status_records or []
        self.hccs_port_statistic = hccs_port_statistic or []
        self.hccs_port_invalid_drop = hccs_port_invalid_drop or []
        self.port_credit_back_pressure_statistics = port_credit_back_pressure_statistics or []
        self.hccs_chip_port_snr_list = hccs_chip_port_snr_list or []
        self.hccs_map_table_list = hccs_map_table_list or []
        self.interface_snr_list = interface_snr_list or []
        self.lane_info_list = lane_info_list or []
        self.serdes_dump_info_list = serdes_dump_info_list or []

    def get_route_miss_by_interface(self, interface: str) -> HccsRouteMiss:
        for hccs_route_miss in self.hccs_route_miss:
            if hccs_route_miss.interface == interface:
                return hccs_route_miss
        return HccsRouteMiss()

    def get_link_status_by_chip_port(self, chip: str, port: str) -> List[PortLinkStatusRecord]:
        link_status_map = {}
        for record in self.link_status_records:
            link_status_map.setdefault(f"{record.chip}_{record.port}", []).append(record)
        return link_status_map.get(f"{chip}_{port}", [])

    def get_package_block_by_condition(self, chip: str, port: str, proxy_module: str, dfx_state_name: str) -> str:
        for record in self.hccs_port_statistic:
            condition = f"{chip}_{port}_{proxy_module}_{dfx_state_name}"
            record_condition = f"{record.chip}_{record.port}_{record.proxy_module}_{record.dfx_state_name}"
            if condition == record_condition:
                return record.dfx_result
        return ""

    def get_timeout_detail_by_condition(self, interface: str, proxy_type: str, response_type: str):
        timeout_detail_map = defaultdict(list[InterfaceProxyResponseDetail])
        for item in self.proxy_response_detail_interfaces:
            timeout_detail_map[item.interface].append(item)
        timeout_details = timeout_detail_map.get(interface, [])
        addr_list = []
        if not timeout_details:
            return addr_list
        for timeout_detail in timeout_details:
            if timeout_detail.proxy_type == proxy_type and timeout_detail.response_type == response_type:
                addr_list.append(timeout_detail.address)
        return addr_list
