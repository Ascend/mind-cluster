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

from diag_tool.core.common import diag_enum, constants
from diag_tool.core.common.json_obj import JsonObj
from diag_tool.core.config import port_mapping_config
from diag_tool.core.log_parser.base import FindResult
from diag_tool.core.model.hccs import ProxyTimeoutStatis, InterfaceProxyResponseDetail, HccsRouteMiss, \
    PortLinkStatusRecord, HccsPortStatistic, HccsPortInvalidDrop, PortCreditBackPressure, InterfaceSnr, LaneSnr, \
    LaneInfo, HccsMapTable, HccsChipPortSnr
from diag_tool.utils import helpers
from diag_tool.utils.helpers import to_int
from diag_tool.utils.logger import DIAG_LOGGER
from diag_tool.utils.table_parser import TableParser


class PortStatisticInfo(JsonObj):

    def __init__(self, chip_id="", port_id="", module_id=""):
        self.chip_id = chip_id
        self.port_id = port_id
        self.module_id = module_id


class HccsParser:
    INTERFACE = "interface"
    SNR = "snr"
    SAVE_NUM = 4
    MODULE_ERROR_KEY_MAP = {
        diag_enum.HCCSProxyModule.RP.value: [diag_enum.HccsPackErrorCnt.RP_PACK_STUACK.value],
        diag_enum.HCCSProxyModule.LP.value: [diag_enum.HccsPackErrorCnt.LP_PACK_STUACK.value],
        diag_enum.HCCSProxyModule.VOQ.value: [diag_enum.HccsPackErrorCnt.VOQ_PACK_DROP.value],
    }

    _XPU_SNR_LIMIT_MAP = {
        diag_enum.XPU.CPU.value: constants.CHIP_CPU_PORT_SNR_LIMIT,
        diag_enum.XPU.NPU.value: constants.CHIP_NPU_PORT_SNR_LIMIT,
    }

    _PORT_STATISTIC_PATTERN = re.compile(
        r'enp s 1 c (?P<chip_id>\d{1,5}) "get port statistic count port (?P<port_id>\d{1,5}) module "(?P<module_id>\d{1,5})"')

    @staticmethod
    def _parse_hccs_proxy_response_detail_interfaces(interface: str, recv: str):
        titles_dict = {
            "proxy_type": "ProxyType",
            "response_type": "ResponseType",
            "address": "Address",
            "collect_time": "CollectTime"
        }
        proxy_timeout_detail = TableParser.parse(recv, titles_dict, {}, 1)
        res = []
        for timeout_detail in proxy_timeout_detail:
            detail = InterfaceProxyResponseDetail.from_dict(timeout_detail)
            detail.interface = interface
            res.append(detail)
        return res

    @classmethod
    def parse_hccs_proxy_response_statistics(cls, cmd_res: str) -> List[ProxyTimeoutStatis]:
        titles_dict = {
            "interface": "Interface",
            "rp_miss": "RemoteProxyMiss",
            "rp_rx": "RemoteProxyRxTimeout",
            "rp_tx": "RemoteProxyTxTimeout",
            "lp_miss": "LocalProxyMiss",
            "lp_rx": "LocalProxyRxTimeout",
            "lp_tx": "LocalProxyTxTimeout"
        }
        proxy_timeout_statistics = TableParser.parse(cmd_res, titles_dict, {}, 1)
        proxy_timeout_statistics = [ProxyTimeoutStatis.from_dict(item) for item in proxy_timeout_statistics]
        error_records = [item for item in proxy_timeout_statistics if
                         item.is_rx_timeout_happend() or item.is_rp_tx_timeout_happend()]
        return error_records

    @classmethod
    def parse_hccs_proxy_response_detail_interfaces(
            cls, all_cmd_res: str, proxy_response_error_records: List[ProxyTimeoutStatis]
    ) -> List[InterfaceProxyResponseDetail]:
        if not proxy_response_error_records:
            return []
        cmd_res_list = helpers.split_str(all_cmd_res, "display hccs proxy")
        result = []
        for cmd_res, record in zip(cmd_res_list, proxy_response_error_records):
            result.extend(cls._parse_hccs_proxy_response_detail_interfaces(record.interface, cmd_res))
        return result

    @classmethod
    def parse_hccs_route_miss(cls, cmd_res: str) -> List[HccsRouteMiss]:
        titles_dict = {
            'interface': 'Interface',
            'rp_direction': 'RpDirection',
            'lp_direction': 'LpDirection',
            'nc_direction': 'NcDirection'
        }
        route_miss_statistics = TableParser.parse(cmd_res, titles_dict, {}, 1)
        route_miss_statistics_objs = [HccsRouteMiss.from_dict(item) for item in route_miss_statistics]
        miss_route_interface = [obj for obj in route_miss_statistics_objs if obj.is_lp_direction_miss()]
        return miss_route_interface

    @classmethod
    def parse_link_status(cls, cmd_res: str) -> List[PortLinkStatusRecord]:
        if not cmd_res:
            return []
        chip_parts = cmd_res.split("display for info")
        titles_dict = {
            'index': 'index',
            'record': 'record'
        }
        link_info_list = []
        port_mapping_config_instance = port_mapping_config.get_port_mapping_config_instance()
        for chip_idx, item in enumerate(chip_parts[1:]):
            record_parts = re.split(r"[\r\n]{3,}", item.strip())
            for port_idx, record in enumerate(record_parts):
                # 数据太多了,每个先保留4行
                for table in TableParser.parse(record, titles_dict)[:cls.SAVE_NUM]:
                    record = PortLinkStatusRecord.from_dict(table)
                    record.interface = str(port_idx)
                    record.chip = str(chip_idx)
                    peer_swi_info = port_mapping_config_instance.find_swi_port(record.chip, phy_id=record.interface)
                    if peer_swi_info:
                        record.swi_port = peer_swi_info.swi_port
                    link_info_list.append(record)
        return link_info_list

    @classmethod
    def parse_port_statistics_chip_info(cls, cmd_res: str) -> List[HccsPortStatistic]:
        titles_dict = {
            'dfx_state_name': 'Dfx_StatName',
            'dfx_result': 'Dfx_Result'
        }
        res = []
        port_info_cmd_res_list = [part for part in cmd_res.strip().split("display for info") if part.strip()]
        for port_info_cmd_res in port_info_cmd_res_list:
            port_statistic_info = PortStatisticInfo()
            search = cls._PORT_STATISTIC_PATTERN.search(port_info_cmd_res)
            if search:
                port_statistic_info = PortStatisticInfo.from_dict(search.groupdict())
            for table in TableParser.parse(port_info_cmd_res, titles_dict, end_sign="diagnose]"):
                statistics = HccsPortStatistic.from_dict(table)
                if statistics.dfx_state_name not in cls.MODULE_ERROR_KEY_MAP.get(port_statistic_info.module_id, []):
                    continue
                statistics.chip = port_statistic_info.chip_id
                statistics.swi_port = port_statistic_info.port_id
                statistics.proxy_module = port_statistic_info.module_id
                res.append(statistics)
        return res

    @classmethod
    def parse_hccs_port_invalid_drop(cls, cmd_res: str) -> List[HccsPortInvalidDrop]:
        titles_dict = {
            "ub_instance": "Ub-instance",
            "link_group": "link-group",
            "rplp": "RPLP",
            "nc": "NC"
        }
        table = TableParser.parse(cmd_res, titles_dict, {}, 1)
        return [HccsPortInvalidDrop.from_dict(item) for item in table]

    @classmethod
    def parse_port_credit_back_pressure_statistics(cls, cmd_res: str) -> List[PortCreditBackPressure]:
        titles_dict = {
            cls.INTERFACE: "Interface",
            "vl": "VL",
            "back_pressure_counts": "Back-pressure Counts",
            "last_time": "           Last-time",
        }
        tables = TableParser.parse(cmd_res, titles_dict, {}, 1)
        last_port = ""
        new_tables = []
        for table in tables:
            cur_port = table.get(cls.INTERFACE, "")
            if cur_port.startswith("----") or table.get("last_time", "") == "-":
                continue
            if not cur_port:
                table[cls.INTERFACE] = last_port
            else:
                last_port = cur_port
            new_tables.append(table)
        return [PortCreditBackPressure.from_dict(table) for table in new_tables]

    @classmethod
    def parse_interface_snr(cls, cmd_res: str) -> List[InterfaceSnr]:
        titles_dict = {
            "interface_name": "interfaceName", "lane1": "lane1", "lane2": "lane2", "lane3": "lane3", "lane4": "lane4",
            "lane5": "lane5", "lane6": "lane6", "lane7": "lane7", "lane8": "lane8",
        }
        end_sign = "------------"
        parse_data_list = TableParser.parse(cmd_res, titles_dict, separate_title_content_lines_num=1, end_sign=end_sign)
        interface_snr_list = []
        for data in parse_data_list:
            interface_name = data.get("interface_name", "")
            lane_snr_list = []
            for title_name, snr_value in data.items():
                if "lane" not in title_name or snr_value == "-":
                    continue
                lane_snr_list.append(LaneSnr(title_name, snr_value))
            if lane_snr_list:
                interface_snr_list.append(InterfaceSnr(interface_name, lane_snr_list))
        return interface_snr_list

    @classmethod
    def parse_if_lane_info(cls, cmd_res: str) -> List[LaneInfo]:
        titles_dict = {
            "if_name": "interfaceName",
            "running_lane_number": "running-lane-num",
            "real_lane_number": "real-lane-num",
        }
        tables = TableParser.parse(cmd_res, titles_dict, {}, 1)
        result_list = [LaneInfo.from_dict(table) for table in tables]
        return result_list

    @classmethod
    def parse_hccs_map_table(cls, cmd_res: str) -> List[HccsMapTable]:
        titles_dict = {
            "port": "Interface", "start_addr": "StartAddr",
            "end_addr": "EndAddr", "base_eid": "BaseEid"
        }
        mapping_config_instance = port_mapping_config.get_port_mapping_config_instance()
        tables = TableParser.parse(cmd_res, titles_dict, {}, 1)
        for table in tables:
            table_obj = HccsMapTable.from_dict(table)
            if not mapping_config_instance.is_local_addr(table_obj.start_addr):
                return [table_obj]
        return []

    @classmethod
    def parse_hccs_port_snr_table(cls, cmd_res: str) -> List[HccsChipPortSnr]:
        port_mapping_config_instance = port_mapping_config.get_port_mapping_config_instance()
        _chip_port_id_regx = re.compile(r"chip id:\s{0,3}(\d)\s{0,3}port id:\s{0,3}(\d{1,3})")
        res = []
        if not cmd_res:
            return res
        titles_dict = {"lane_id": "laneId", "snr": "snr"}
        parts = [part for part in cmd_res.split("display for info enp s 1 c") if part.strip()]
        for part in parts:
            if not part.strip():
                continue
            match = _chip_port_id_regx.search(part)
            if not match:
                continue
            for row in TableParser.parse(part, titles_dict):
                port_snr = HccsChipPortSnr.from_dict(row)
                port_snr.swi_chip_id = match.group(1)
                port_snr.port_id = match.group(2)
                port_mapping_instance = port_mapping_config_instance.find_swi_port(str(port_snr.swi_chip_id),
                                                                                   phy_id=str(port_snr.port_id))
                if port_mapping_instance:
                    port_snr.swi_port = port_mapping_instance.swi_port
                    port_snr.xpu = port_mapping_instance.xpu
                else:
                    DIAG_LOGGER.warning(f"未找到chip: {port_snr.swi_chip_id} port: {port_snr.port_id}对应的端口")
                res.append(port_snr)
        return res

    @classmethod
    def parse_hccs_port_snr_line(cls, switch_log_info: List[FindResult]) -> List[HccsChipPortSnr]:
        res = []
        titles_dict = {
            "lane_id": "laneId", cls.SNR: cls.SNR, "cnt": "data-rate(MHz)", "fec": "tx-amp-ctl-en", "loss": "losStatus"
        }
        port_mapping_config_instance = port_mapping_config.get_port_mapping_config_instance()
        for log_info in switch_log_info:
            info_dict = log_info.info_dict
            if not info_dict or log_info.pattern_key != "link_snr":
                continue
            table_lane_info_list = TableParser.parse(log_info.logline, titles_dict, {},
                                                     end_sign="-----------", both_strip=False)
            for lane_info in table_lane_info_list:
                port_snr = HccsChipPortSnr()
                port_snr.swi_chip_id = to_int(info_dict.get("swi_chip_id", "")) - 1
                port_snr.port_id = to_int(info_dict.get("port_id", "")) - 1
                port_mapping_instance = port_mapping_config_instance.find_swi_port(str(port_snr.swi_chip_id),
                                                                                   phy_id=str(port_snr.port_id))
                if port_mapping_instance:
                    port_snr.swi_port = port_mapping_instance.swi_port
                    port_snr.xpu = port_mapping_instance.xpu
                else:
                    DIAG_LOGGER.warning(f"未找到chip: {port_snr.swi_chip_id} port: {port_snr.port_id}对应的端口")
                port_snr.lane_id = lane_info.get("lane_id", "")
                port_snr.snr = lane_info.get(cls.SNR, "")
                res.append(port_snr)
        return res
