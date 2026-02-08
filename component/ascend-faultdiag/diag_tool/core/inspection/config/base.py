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

import enum

from diag_tool.core.common.json_obj import JsonObj


class OpticalThreshold(JsonObj):

    def __init__(self, media_snr_error_threshold: int,
                 txrx_power_threshold: float = 0,
                 txrx_power_diff_threshold: float = 0):
        self.media_snr_error_threshold = media_snr_error_threshold
        self.txrx_power_threshold = txrx_power_threshold
        self.txrx_power_diff_threshold = txrx_power_diff_threshold


class BerThreshold(JsonObj):

    def __init__(self, operations_phase_err: float = 0,
                 operations_phase_warn: float = 0,
                 delivery_phase_err: float = 0):
        self.operations_phase_err = operations_phase_err
        self.operations_phase_warn = operations_phase_warn
        self.delivery_phase_err = delivery_phase_err


class InspectionConfig(JsonObj):

    def __init__(self, hccs_swi_optical_threshold: OpticalThreshold = None,
                 roce_swi_optical_threshold: OpticalThreshold = None,
                 host_roce_optical_threshold: OpticalThreshold = None,
                 ber_threshold: BerThreshold = None):
        self.hccs_swi_optical_threshold = hccs_swi_optical_threshold
        self.roce_swi_optical_threshold = roce_swi_optical_threshold
        self.host_roce_optical_threshold = host_roce_optical_threshold
        self.ber_threshold = ber_threshold


class FaultDescTemplate(enum.Enum):
    MEDIA_SNR_WARN = "链路亚健康, A端模块lane{} SNR值{}db小于告警值{}db"
    LOW_POWER_WARN = "链路亚健康, A端光模块lane{} {}光功率{}db较低, 低于{}db"
    LANE_POWER_DIFF_FAULT = "{}端口Lane最大值和最小值差值大于{}db, 实际最大值lane{}: {}db, 最小值lane{}: {}db"
