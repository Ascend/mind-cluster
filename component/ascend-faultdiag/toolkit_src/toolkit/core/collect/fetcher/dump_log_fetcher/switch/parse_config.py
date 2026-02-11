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

from collections.abc import Callable
from typing import List

from toolkit.core.collect.collect_config import SwiCliOutputDataType
from toolkit.utils import list_tool
from toolkit.utils import helpers


class SwiCliOutputParseConfig:

    def __init__(self, primary_keys: List[str], data_type: SwiCliOutputDataType, desc="",
                 multi_parts_judge_func: Callable = None, is_multi_item=False):
        self.primary_keys = primary_keys
        self.data_type = data_type
        self.desc = desc
        # 存在多个可能片段时,需要有这个函数做细化区分
        self.multi_parts_judge_func = multi_parts_judge_func
        # 是否为同类多项
        self.is_multi_item = is_multi_item


def _display_alarm_judge_func(content: str):
    lines = content.splitlines()
    if not lines:
        return False
    separators = [">", "]"]
    separator = list_tool.find_first(list(lines[0]), lambda char: char in separators)
    if not separator:
        return False
    # 找到h, 则说明是历史(history)alarm
    search = helpers.find_pattern_after_substrings(lines[0], [separator], "h")
    return bool(search)


def _display_alarm_active_judge_func(content: str):
    return not _display_alarm_judge_func(content)


def _display_alarm_history_judge_func(content: str):
    return _display_alarm_judge_func(content)


PARSE_CONFIGS = [
    SwiCliOutputParseConfig(["AlarmId", "AlarmName", "AlarmType", "State : active"],
                            SwiCliOutputDataType.ALARM_ACTIVE_VERBOSE, "display alarm active verbose"),
    SwiCliOutputParseConfig(["AlarmId", "AlarmName", "AlarmType", "State : cleared"],
                            SwiCliOutputDataType.ALARM_HISTORY_VERBOSE, "display alarm active verbose"),
    SwiCliOutputParseConfig(["Sequence", "AlarmId", "Severity", "Date Time", "Description"],
                            SwiCliOutputDataType.ALARM_ACTIVE, "display alarm active",
                            _display_alarm_active_judge_func),
    SwiCliOutputParseConfig(["Sequence", "AlarmId", "Severity", "Date Time", "Description"],
                            SwiCliOutputDataType.ALARM_HISTORY, "display alarm history",
                            _display_alarm_history_judge_func),
    SwiCliOutputParseConfig(["Local Interface", "Exptime(s)", "Neighbor Interface", "Neighbor Device"],
                            SwiCliOutputDataType.LLDP_NEI_B),
    SwiCliOutputParseConfig(["Items", "Value", "HighAlarm", "HighWarn", "LowAlarm", "Status"],
                            SwiCliOutputDataType.OPTICAL_MODULE),
    SwiCliOutputParseConfig(["Interface", "PHY", "Protocol", "InUti", "OutUti", "inErrors", "outErrors"],
                            SwiCliOutputDataType.IF_BRIEF),
    SwiCliOutputParseConfig(["Current state", "Speed"], SwiCliOutputDataType.BIT_ERR_RATE),
    SwiCliOutputParseConfig(["current state", "Description", "Port Mode"], SwiCliOutputDataType.IF_INFO),
    SwiCliOutputParseConfig(["MainBoard", "ESN"], SwiCliOutputDataType.LICENSE_ESN),
    SwiCliOutputParseConfig(["clock", "Time Zone"], SwiCliOutputDataType.CLOCK),
    SwiCliOutputParseConfig(["transceiver information:"], SwiCliOutputDataType.IF_TRANSCEIVER_INFO),
    SwiCliOutputParseConfig(["Interface", "IfIndex", "TB", "TP", "Chip", "Port", "Core"],
                            SwiCliOutputDataType.PORT_MAPPING),
    # hccs
    SwiCliOutputParseConfig(["Interface", "RemoteProxyMiss", "RemoteProxyRxTimeout", "RemoteProxyTxTimeout",
                             "LocalProxyMiss", "LocalProxyRxTimeout", "LocalProxyTxTimeout"],
                            SwiCliOutputDataType.HCCS_PROXY_RESP_STATISTIC),
    SwiCliOutputParseConfig(["ProxyType", "ResponseType", "Address", "CollectTime"],
                            SwiCliOutputDataType.HCCS_PROXY_RESP_DETAIL),
    SwiCliOutputParseConfig(['Interface', 'RpDirection', 'LpDirection', 'NcDirection'],
                            SwiCliOutputDataType.HCCS_ROUTE_MISS),
    SwiCliOutputParseConfig(["[ GET LINK STATUS ] Link status record"], SwiCliOutputDataType.HCCS_PORT_LINK_STATUS),
    SwiCliOutputParseConfig(['Dfx_StatName', 'Dfx_Result'], SwiCliOutputDataType.HCCS_PORT_STATISTIC_CHIP_INFO),
    SwiCliOutputParseConfig(["Ub-instance", "link-group", "RPLP", "NC"], SwiCliOutputDataType.HCCS_PORT_INVALID_DROP),
    SwiCliOutputParseConfig(["Interface", "VL", "Back-pressure Counts", "Last-time"],
                            SwiCliOutputDataType.HCCS_PORT_CREDIT_BACK_PRESSURES_STATISTIC),
    SwiCliOutputParseConfig(["Interface", "StartAddr", "EndAddr", "BaseEid"], SwiCliOutputDataType.HCCS_MAP_TABLE),
    SwiCliOutputParseConfig(["interfaceName", "lane1", "lane2", "lane3", "lane4"], SwiCliOutputDataType.HCCS_IF_SNR),
    SwiCliOutputParseConfig(["interfaceName", "running-lane-num", "real-lane-num"],
                            SwiCliOutputDataType.HCCS_IF_LANE_INFO),
    SwiCliOutputParseConfig(["PORT SNR ]"], SwiCliOutputDataType.HCCS_PORT_SNR)

]
