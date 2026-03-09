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

from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.utils import helpers


class ThresholdStatus(Enum):
    NORMAL = auto()
    LOW_THRESHOLD_ALARM = auto()
    LOW_THRESHOLD_WARN = auto()
    HIGH_THRESHOLD_ALARM = auto()
    HIGH_THRESHOLD_WARN = auto()
    NOT_EQUAL_THRESHOLD_ALARM = auto()
    NOT_EQUAL_THRESHOLD_WARN = auto()


class Threshold(JsonObj):
    _STR_MAPPING = {
        ThresholdStatus.HIGH_THRESHOLD_WARN: "{}实际值：{}，高于告警阈值：{}，单位：{}",
        ThresholdStatus.HIGH_THRESHOLD_ALARM: "{}实际值：{}，高于故障阈值：{}，单位：{}",
        ThresholdStatus.LOW_THRESHOLD_WARN: "{}实际值：{}，低于告警阈值：{}，单位：{}",
        ThresholdStatus.LOW_THRESHOLD_ALARM: "{}实际值：{}，低于故障阈值：{}，单位：{}",
        ThresholdStatus.NOT_EQUAL_THRESHOLD_WARN: "{}实际值：{}，不等于期望阈值：{}，单位：{}",
        ThresholdStatus.NOT_EQUAL_THRESHOLD_ALARM: "{}实际值：{}，不等于期望阈值：{}，单位：{}",
    }

    def __init__(self, low_threshold_alarm: str = "", high_threshold_alarm: str = "", low_threshold_warn: str = "",
                 high_threshold_warn: str = "", normal_value_alarm: str = "", normal_value_warn: str = "",
                 desc="", unit=""):
        self.low_alarm_th = low_threshold_alarm
        self._has_low_alarm_th, self._low_alarm_th_f = helpers.to_float(low_threshold_alarm)
        self.high_alarm_th = high_threshold_alarm
        self._has_high_alarm_th, self._high_alarm_th_f = helpers.to_float(high_threshold_alarm)
        self.low_warn_th = low_threshold_warn
        self._has_low_warn_th, self._low_warn_th_f = helpers.to_float(low_threshold_warn)
        self.high_warn_th = high_threshold_warn
        self._has_high_warn_th, self._high_warn_th_f = helpers.to_float(high_threshold_warn)
        self.normal_alarm_th = normal_value_alarm  # 正常阈值（异常级别）
        self.normal_warn_th = normal_value_warn  # 正常阈值（警告级别）
        self.desc = desc
        self.unit = unit

    def check_value(self, value: str) -> Tuple[ThresholdStatus, str]:
        # 首先尝试数值比较
        success, value_f = helpers.to_float(value)
        if success:
            if self._has_high_alarm_th and self._high_alarm_th_f < value_f:
                return ThresholdStatus.HIGH_THRESHOLD_ALARM, self.high_alarm_th
            if self._has_high_warn_th and self._high_warn_th_f < value_f:
                return ThresholdStatus.HIGH_THRESHOLD_WARN, self.high_warn_th
            if self._has_low_alarm_th and self._low_alarm_th_f > value_f:
                return ThresholdStatus.LOW_THRESHOLD_ALARM, self.low_alarm_th
            if self._has_low_warn_th and self._low_warn_th_f > value_f:
                return ThresholdStatus.LOW_THRESHOLD_WARN, self.low_warn_th
        else:
            # 数值转换失败，尝试字符串相等比较（忽略大小写）
            # 逻辑：只有等于normal_alarm_th或normal_warn_th才是正常，其他均为异常
            # 优先级：异常级别高于警告级别
            if value:
                # 检查是否设置了异常级别的正常阈值
                if self.normal_alarm_th:
                    if value.strip().lower() != self.normal_alarm_th.strip().lower():
                        return ThresholdStatus.NOT_EQUAL_THRESHOLD_ALARM, self.normal_alarm_th
                # 检查是否设置了警告级别的正常阈值
                elif self.normal_warn_th:
                    if value.strip().lower() != self.normal_warn_th.strip().lower():
                        return ThresholdStatus.NOT_EQUAL_THRESHOLD_WARN, self.normal_warn_th
        return ThresholdStatus.NORMAL, ""

    def check_value_str(self, value: str) -> str:
        th_type, th_value = self.check_value(value)
        if th_type in self._STR_MAPPING:
            return self._STR_MAPPING[th_type].format(self.desc, value, th_value, self.unit)
        return ""
