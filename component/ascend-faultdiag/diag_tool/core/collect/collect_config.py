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

from enum import Enum, auto


class ToolLogCollectionSplitType(Enum):
    NONE = auto()
    DEVICE_ID = auto()
    DEVICE_CHIP_ID = auto()


class ToolLogCollectionDataType(Enum):
    HOST_ID = auto()
    NPU_TYPE = auto()
    SN = auto()
    LLDP = auto()
    SPEED = auto()
    OPTICAL = auto()
    LINK_STAT = auto()
    STAT = auto()
    SPOD_INFO = auto()
    HCCS = auto()
    MS_NPU_REPORT = auto()
    NET_HEALTH = auto()
    HCCN_LINK_STATUS = auto()
    CDR_SNR = auto()
    DFX_CFG = auto()


class SwiCliOutputDataType(Enum):
    SWI_NAME = auto()
    SWI_IP = auto()
    ALARM_ACTIVE = auto()
    ALARM_HISTORY = auto()
    ALARM_ACTIVE_VERBOSE = auto()
    ALARM_HISTORY_VERBOSE = auto()
    LLDP_NEI_B = auto()
    BIT_ERR_RATE = auto()
    OPTICAL_MODULE = auto()
    IF_INFO = auto()
    IF_BRIEF = auto()
    IF_TRANSCEIVER_INFO = auto()
    LICENSE_ESN = auto()
    CLOCK = auto()
    PORT_MAPPING = auto()
    # hccs
    HCCS_PROXY_RESP_STATISTIC = auto()
    HCCS_PROXY_RESP_DETAIL = auto()
    HCCS_ROUTE_MISS = auto()
    HCCS_PORT_LINK_STATUS = auto()
    HCCS_PORT_STATISTIC_CHIP_INFO = auto()
    HCCS_PORT_INVALID_DROP = auto()
    HCCS_PORT_CREDIT_BACK_PRESSURES_STATISTIC = auto()
    HCCS_MAP_TABLE = auto()
    HCCS_IF_SNR = auto()
    HCCS_IF_LANE_INFO = auto()
    HCCS_PORT_SNR = auto()
    # log
    DIAG_INFO_LOG = auto()
    PORT_DOWN_STATUS = auto()
