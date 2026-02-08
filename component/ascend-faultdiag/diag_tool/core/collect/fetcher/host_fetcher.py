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

from diag_tool.core.collect.fetcher.base import Fetcher
from diag_tool.core.log_parser.base import FindResult


def _build_default_npu_mapping():
    npu_size = 8
    chip_size = 2
    result = {}
    for npu_id in range(npu_size):
        tmp_dict = {}
        result[str(npu_id)] = tmp_dict
        for chip_id in range(chip_size):
            tmp_dict[str(chip_id)] = str(npu_id * 2 + chip_id)
    return result


_DEFAULT_NPU_MAPPING = _build_default_npu_mapping()


class HostFetcher(Fetcher):

    @abc.abstractmethod
    async def fetch_hostname(self) -> str:
        pass

    @abc.abstractmethod
    async def fetch_npu_mapping(self) -> dict:
        return _DEFAULT_NPU_MAPPING

    @abc.abstractmethod
    async def fetch_optical_info(self, chip_phy_id) -> str:
        pass

    @abc.abstractmethod
    async def fetch_link_stat_info(self, chip_phy_id) -> str:
        pass

    @abc.abstractmethod
    async def fetch_stat_info(self, chip_phy_id) -> str:
        pass

    @abc.abstractmethod
    async def fetch_lldp_info(self, chip_phy_id) -> str:
        pass

    @abc.abstractmethod
    async def fetch_npu_type(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_sn_num(self) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_hccs_info(self, npu_id, chip_id) -> str:
        pass

    @abc.abstractmethod
    async def fetch_spod_info(self, npu_id, chip_id) -> str:
        pass

    @abc.abstractmethod
    async def fetch_msnpureport_log(self) -> List[FindResult]:
        pass

    @abc.abstractmethod
    async def fetch_roce_speed(self, chip_phy_id) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_roce_duplex(self, chip_phy_id) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_hccn_tool_net_health(self, chip_phy_id) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_hccn_tool_link_status(self, chip_phy_id) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_hccn_tool_cdr(self, chip_phy_id) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_hccn_dfx_cfg(self, chip_phy_id) -> str:
        return ""

    @abc.abstractmethod
    async def fetch_optical_loopback_enable(self, npu_id, model) -> bool:
        return False
