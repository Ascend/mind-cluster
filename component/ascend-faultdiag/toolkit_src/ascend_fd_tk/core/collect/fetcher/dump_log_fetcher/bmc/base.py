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
from enum import Enum


class BmcDumpLogDataType(Enum):
    BMC_IP = "bmc_ip"
    SN_NUM = "sn_num"
    SEL_INFO = "sel_info"
    SENSOR_INFO = "sensor_info"
    HEALTH_EVENTS = "health_events"
    OP_HISTORY_INFO_LOG = "optical_history_info_log"
