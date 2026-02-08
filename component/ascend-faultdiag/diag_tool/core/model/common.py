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
from typing import Tuple

from diag_tool.core.common.json_obj import JsonObj
from diag_tool.utils import helpers


class ThresholdStatus(Enum):
    NORMAL = auto()
    LOW_THRESHOLD_ALARM = auto()
    LOW_THRESHOLD_WARN = auto()
    HIGH_THRESHOLD_ALARM = auto()
    HIGH_THRESHOLD_WARN = auto()


class Threshold(JsonObj):
    _STR_MAPPING = {
        ThresholdStatus.HIGH_THRESHOLD_WARN: "{}实际值: {}，高于告警阈值: {}",
        ThresholdStatus.HIGH_THRESHOLD_ALARM: "{}实际值: {}，高于故障阈值: {}",
        ThresholdStatus.LOW_THRESHOLD_WARN: "{}实际值: {}，低于告警阈值: {}",
        ThresholdStatus.LOW_THRESHOLD_ALARM: "{}实际值: {}，低于故障阈值: {}",
    }

    def __init__(self, low_threshold_alarm: str = "", high_threshold_alarm: str = "", low_threshold_warn: str = "",
                 high_threshold_warn: str = "", desc=""):
        self.low_alarm_th = low_threshold_alarm
        self._has_low_alarm_th, self._low_alarm_th_f = helpers.to_float(low_threshold_alarm)
        self.high_alarm_th = high_threshold_alarm
        self._has_high_alarm_th, self._high_alarm_th_f = helpers.to_float(high_threshold_alarm)
        self.low_warn_th = low_threshold_warn
        self._has_low_warn_th, self._low_warn_th_f = helpers.to_float(low_threshold_warn)
        self.high_warn_th = high_threshold_warn
        self._has_high_warn_th, self._high_warn_th_f = helpers.to_float(high_threshold_warn)
        self.desc = desc

    def check_value(self, value: str) -> Tuple[ThresholdStatus, str]:
        success, value_f = helpers.to_float(value)
        if not success:
            return ThresholdStatus.NORMAL, ""
        if self._has_high_alarm_th and self._high_alarm_th_f < value_f:
            return ThresholdStatus.HIGH_THRESHOLD_ALARM, self.high_alarm_th
        if self._has_high_warn_th and self._high_warn_th_f < value_f:
            return ThresholdStatus.HIGH_THRESHOLD_ALARM, self.high_warn_th
        if self._has_low_alarm_th and self._low_alarm_th_f > value_f:
            return ThresholdStatus.LOW_THRESHOLD_ALARM, self.low_alarm_th
        if self._has_low_warn_th and self._low_warn_th_f > value_f:
            return ThresholdStatus.LOW_THRESHOLD_WARN, self.low_warn_th
        return ThresholdStatus.NORMAL, ""

    def check_value_str(self, value: str) -> str:
        th_type, th_value = self.check_value(value)
        if th_type in self._STR_MAPPING:
            return self._STR_MAPPING[th_type].format(self.desc, value, th_value)
        return ""
