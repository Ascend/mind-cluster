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
import os
from typing import List

from toolkit.core.collect.collect_config import ToolLogCollectionDataType, ToolLogCollectionSplitType


class ParseConfig:

    def __init__(self, data_type: ToolLogCollectionDataType, split_type: ToolLogCollectionSplitType, title: str,
                 file_path: str):
        self.data_type = data_type
        self.title = title
        self.file_path = file_path
        self.split_type = split_type


class ParseConfigCollection(abc.ABC):

    @classmethod
    @abc.abstractmethod
    def get_configs(cls) -> List[ParseConfig]:
        pass

class ParseConfigCollectionV1(ParseConfigCollection):
    _HCCN_TOOL_PATH = "hccn_tool.log"
    _NPU_CARD_INFO = "npu_card_info.log"
    _PCIE_INFO_PATH = "pcie_info.log"
    _VERSION_INFO_PATH = "version_info.log"

    @classmethod
    def get_configs(cls) -> List[ParseConfig]:
        return [
            # ====> hccn_tool -i 0 -lldp -g
            ParseConfig(ToolLogCollectionDataType.LLDP, ToolLogCollectionSplitType.DEVICE_ID, "lldp",
                        cls._HCCN_TOOL_PATH),
            # ====> hccn_tool -i 0 -speed -g
            ParseConfig(ToolLogCollectionDataType.SPEED, ToolLogCollectionSplitType.DEVICE_ID, "speed -g",
                        cls._HCCN_TOOL_PATH),
            # ====> hccn_tool -i 0 -optical -g
            ParseConfig(ToolLogCollectionDataType.OPTICAL, ToolLogCollectionSplitType.DEVICE_ID, "optical",
                        cls._HCCN_TOOL_PATH),
            # ====> hccn_tool -i 0 -link_stat -g
            ParseConfig(ToolLogCollectionDataType.LINK_STAT, ToolLogCollectionSplitType.DEVICE_ID, "link stat",
                        cls._HCCN_TOOL_PATH),
            # ====> hccn_tool -i 0 -stat -g
            ParseConfig(ToolLogCollectionDataType.STAT, ToolLogCollectionSplitType.DEVICE_ID, "stat",
                        cls._HCCN_TOOL_PATH),
            # ====> hccn_tool -i 0 -link -g
            ParseConfig(ToolLogCollectionDataType.HCCN_LINK_STATUS, ToolLogCollectionSplitType.DEVICE_ID, "link",
                        cls._HCCN_TOOL_PATH),
            # ====> hccn_tool -i 0 -scdr -t 5
            ParseConfig(ToolLogCollectionDataType.CDR_SNR, ToolLogCollectionSplitType.DEVICE_ID, "cdr5 snr 1 times",
                        cls._HCCN_TOOL_PATH),
            # npu-smi info -t spod-info -i 0 -c 0
            ParseConfig(ToolLogCollectionDataType.SPOD_INFO, ToolLogCollectionSplitType.DEVICE_CHIP_ID,
                        "Collect spod-info info for all NPUs", cls._NPU_CARD_INFO),
            # npu-smi info -t hccs -i 0 -c 0
            ParseConfig(ToolLogCollectionDataType.HCCS, ToolLogCollectionSplitType.DEVICE_CHIP_ID,
                        "Collect hccs info for all NPUs", cls._NPU_CARD_INFO),

            ParseConfig(ToolLogCollectionDataType.NPU_TYPE, ToolLogCollectionSplitType.NONE, "lspci",
                        cls._PCIE_INFO_PATH),

            ParseConfig(ToolLogCollectionDataType.SN, ToolLogCollectionSplitType.NONE, "timeout 30s dmidecode -t1",
                        cls._VERSION_INFO_PATH)
        ]

class ParseConfigCollectionV2(ParseConfigCollection):
    _NET_CONF_PATH = os.path.join("hccn_log", "net_conf.log")
    _OPTICAL_PATH = os.path.join("hccn_log", "optical.log")
    _STAT_PATH = os.path.join("hccn_log", "stat.log")

    _NPU_SMI_PATH = os.path.join("npu_smi_log", "npu_smi.log")
    _PCIE_PATH = os.path.join("pcie_log", "pcie.log")

    @classmethod
    def get_configs(cls) -> List[ParseConfig]:
        return [
            # hccn_log/net_conf.log
            ParseConfig(ToolLogCollectionDataType.LLDP, ToolLogCollectionSplitType.DEVICE_ID, "lldp",
                        cls._NET_CONF_PATH),
            ParseConfig(ToolLogCollectionDataType.SPEED, ToolLogCollectionSplitType.DEVICE_ID, "speed",
                        cls._NET_CONF_PATH),
            # hccn_log/optical.log
            ParseConfig(ToolLogCollectionDataType.OPTICAL, ToolLogCollectionSplitType.DEVICE_ID, "optical",
                        cls._OPTICAL_PATH),
            ParseConfig(ToolLogCollectionDataType.LINK_STAT, ToolLogCollectionSplitType.DEVICE_ID, "link stat",
                        cls._OPTICAL_PATH),
            ParseConfig(ToolLogCollectionDataType.NET_HEALTH, ToolLogCollectionSplitType.DEVICE_ID, "health info",
                        cls._OPTICAL_PATH),
            ParseConfig(ToolLogCollectionDataType.HCCN_LINK_STATUS, ToolLogCollectionSplitType.DEVICE_ID, "link info",
                        cls._OPTICAL_PATH),
            # hccn_log/stat.log
            ParseConfig(ToolLogCollectionDataType.STAT, ToolLogCollectionSplitType.DEVICE_ID, "stat", cls._STAT_PATH),
            # npu_smi_log/npu_smi.log
            ParseConfig(ToolLogCollectionDataType.SPOD_INFO, ToolLogCollectionSplitType.DEVICE_CHIP_ID, "spod_info",
                        cls._NPU_SMI_PATH),
            ParseConfig(ToolLogCollectionDataType.HCCS, ToolLogCollectionSplitType.DEVICE_CHIP_ID, "hccs",
                        cls._NPU_SMI_PATH),
            # pcie_log/pcie.log
            ParseConfig(ToolLogCollectionDataType.NPU_TYPE, ToolLogCollectionSplitType.NONE, "pcie", cls._PCIE_PATH)
        ]

class ParseConfigCollectionV3(ParseConfigCollection):
    _LLDP_LOG = "lldp.log"
    _OPTICAL_LOG = "optical.log"

    @classmethod
    def get_configs(cls) -> List[ParseConfig]:
        return [
            # lldp.log
            ParseConfig(ToolLogCollectionDataType.LLDP, ToolLogCollectionSplitType.DEVICE_ID, "lldp", cls._LLDP_LOG),
            # optical.log
            ParseConfig(ToolLogCollectionDataType.SPEED, ToolLogCollectionSplitType.DEVICE_ID, "speed info",
                        cls._OPTICAL_LOG),
            ParseConfig(ToolLogCollectionDataType.OPTICAL, ToolLogCollectionSplitType.DEVICE_ID, "optical",
                        cls._OPTICAL_LOG),
            ParseConfig(ToolLogCollectionDataType.LINK_STAT, ToolLogCollectionSplitType.DEVICE_ID, "link stat",
                        cls._OPTICAL_LOG),
            ParseConfig(ToolLogCollectionDataType.HCCN_LINK_STATUS, ToolLogCollectionSplitType.DEVICE_ID, "link info",
                        cls._OPTICAL_LOG),
        ]
