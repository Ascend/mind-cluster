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

from ascend_fd_tk.core.collect.collect_config import SwiCliOutputDataType
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.cli_output_parsed_data import CliOutputParsedData
from ascend_fd_tk.core.collect.fetcher.switch_fetcher import SwitchFetcher
from ascend_fd_tk.core.log_parser.base import FindResult
from ascend_fd_tk.core.model.hccs import HccsMapTable, HccsChipPortSnr, ProxyTimeoutStatis, HccsSerdesDumpInfo
from ascend_fd_tk.core.model.switch import InterfaceBrief, PortDownStatus
from ascend_fd_tk.utils import list_tool


class SwiCliOutputFetcher(SwitchFetcher):

    def __init__(self, parsed_data: CliOutputParsedData):
        self.parsed_data = parsed_data

    async def fetch_id(self) -> str:
        return (self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.SWI_IP.name) or
                self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.SWI_NAME.name))

    async def fetch_interface_lane_information(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_IF_LANE_INFO.name)

    async def init_fetcher(self):
        pass

    async def fetch_serial_num(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.LICENSE_ESN.name)

    async def fetch_interface_brief(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.IF_BRIEF.name)

    async def get_switch_name(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.SWI_NAME.name)

    async def fetch_optical_module_info(self, interface_briefs: List[InterfaceBrief]) -> str:
        return ""

    async def fetch_switch_log_info(self) -> List[FindResult]:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.DIAG_INFO_LOG.name)

    async def fetch_bit_error_rate(self, interface_briefs: List[InterfaceBrief]) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.BIT_ERR_RATE.name)

    async def fetch_lldp_nei_brief(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.LLDP_NEI_B.name)

    async def fetch_active_alarms(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.ALARM_ACTIVE.name)

    async def fetch_history_alarms(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.ALARM_HISTORY.name)

    async def fetch_active_alarms_verbose(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.ALARM_ACTIVE_VERBOSE.name)

    async def fetch_history_alarms_verbose(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.ALARM_HISTORY_VERBOSE.name)

    async def fetch_interface_info(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.IF_INFO.name)

    async def fetch_datetime(self) -> str:
        clock: str = self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.CLOCK.name)
        if not clock:
            return ""
        date_line = list_tool.find_first(clock.splitlines(), lambda line: "-" in line, "")
        return date_line

    async def fetch_hccs_proxy_response_statistics(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_PROXY_RESP_STATISTIC.name)

    async def fetch_hccs_proxy_response_detail_interfaces(
            self, proxy_response_error_records: List[ProxyTimeoutStatis]
    ) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_PROXY_RESP_DETAIL.name)

    async def fetch_hccs_route_miss(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_ROUTE_MISS.name)

    async def fetch_link_status(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_PORT_LINK_STATUS.name)

    async def fetch_port_statistic(self) -> str:
        return ""

    async def fetch_hccs_port_invalid_drop(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_PORT_INVALID_DROP.name)

    async def fetch_port_credit_back_pressure_statistics(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_PORT_CREDIT_BACK_PRESSURES_STATISTIC.name)

    async def has_hccs(self) -> bool:
        return bool(self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_IF_SNR.name))

    async def fetch_port_snr(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_PORT_SNR.name)

    async def fetch_hccs_map_table(self) -> List[HccsMapTable]:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_MAP_TABLE.name)

    async def fetch_interface_snr(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.HCCS_IF_SNR.name)

    async def fetch_transceiver_info(self):
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.IF_TRANSCEIVER_INFO.name)

    async def fetch_serdes_dump_info(self, port_snr_list: List[HccsChipPortSnr]) -> List[HccsSerdesDumpInfo]:
        return []

    async def fetch_interface_port_mapping(self) -> str:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.PORT_MAPPING.name)

    async def fetch_port_down_status(self) -> List[PortDownStatus]:
        return self.parsed_data.fetch_data_by_name(SwiCliOutputDataType.PORT_DOWN_STATUS.name, [])
