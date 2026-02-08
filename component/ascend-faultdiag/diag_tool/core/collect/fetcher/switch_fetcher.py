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

import abc
from typing import List

from diag_tool.core.log_parser.base import FindResult
from diag_tool.core.model.hccs import ProxyTimeoutStatis, HccsMapTable, HccsChipPortSnr, HccsSerdesDumpInfo

from diag_tool.core.collect.fetcher.base import Fetcher
from diag_tool.core.model.switch import InterfaceBrief, PortDownStatus


class SwitchFetcher(Fetcher):

    @abc.abstractmethod
    async def init_fetcher(self):
        pass

    @abc.abstractmethod
    async def fetch_serial_num(self):
        return ""

    @abc.abstractmethod
    async def fetch_interface_brief(self) -> str:
        return ""

    @abc.abstractmethod
    async def get_switch_name(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_optical_module_info(self, interface_briefs: List[InterfaceBrief]) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_switch_log_info(self) -> List[FindResult]:
        return []

    @abc.abstractmethod
    async def fetch_bit_error_rate(self, interface_briefs: List[InterfaceBrief]) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_lldp_nei_brief(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_active_alarms(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_history_alarms(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_active_alarms_verbose(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_history_alarms_verbose(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_interface_info(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_datetime(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_hccs_proxy_response_statistics(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_hccs_proxy_response_detail_interfaces(self, proxy_response_error_records: List[
        ProxyTimeoutStatis]) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_hccs_route_miss(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_link_status(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_port_statistic(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_hccs_port_invalid_drop(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_port_credit_back_pressure_statistics(self) -> str:
        return ""

    @abc.abstractmethod
    async def has_hccs(self) -> bool:
        return ""

    @abc.abstractmethod
    async def fetch_port_snr(self) -> str:
        pass

    @abc.abstractmethod
    async def fetch_hccs_map_table(self) -> str:
        pass

    @abc.abstractmethod
    async def fetch_interface_snr(self) -> str:
        pass

    @abc.abstractmethod
    async def fetch_transceiver_info(self):
        pass

    @abc.abstractmethod
    async def fetch_interface_lane_information(self) -> str:
        pass

    @abc.abstractmethod
    async def fetch_serdes_dump_info(self, port_snr_list: List[HccsChipPortSnr]) -> List[HccsSerdesDumpInfo]:
        return []

    @abc.abstractmethod
    async def fetch_diag_info_log(self) -> List[FindResult]:
        return []

    @abc.abstractmethod
    async def fetch_interface_port_mapping(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_port_down_status(self) -> List[PortDownStatus]:
        return []