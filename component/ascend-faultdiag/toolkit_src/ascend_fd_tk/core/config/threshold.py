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

from ascend_fd_tk.core.model.common import Threshold


class OpticalModuleThreshold:
    # 功率阈值(mW)
    TX_POWER_THRESHOLD_CONFIG_MW = Threshold(low_threshold_alarm="0.2", high_threshold_alarm="2.5", desc="tx power(mW)")
    RX_POWER_THRESHOLD_CONFIG_MW = Threshold(low_threshold_alarm="0.1445", low_threshold_warn="0.6",
                                             high_threshold_alarm="2.3", desc="rx power(mW)")
    # 功率阈值(dBm)
    TX_POWER_THRESHOLD_CONFIG_DBM = Threshold(low_threshold_alarm="-9.60", high_threshold_alarm="7.00",
                                              low_threshold_warn="-7.00", high_threshold_warn="5.50",
                                              desc="tx power(dBm)")
    RX_POWER_THRESHOLD_CONFIG_DBM = Threshold(low_threshold_alarm="-10.00", high_threshold_alarm="7.00",
                                              low_threshold_warn="-6.50", high_threshold_warn="5.50",
                                              desc="rx power(dBm)")
    # 电流
    TX_BIAS_MA = Threshold(low_threshold_alarm="6", high_threshold_alarm="10", desc="tx bias(mA)")
    # snr
    HOST_SNR_DB = Threshold(low_threshold_warn="20", low_threshold_alarm="18", desc="host snr(dB)")
    MEDIA_SNR_DB = Threshold(low_threshold_warn="20", low_threshold_alarm="18", desc="media snr(dB)")
    # cdr snr
    CDR_HOST_SNR_DB = Threshold(low_threshold_alarm="20", desc="cdr host snr(dB)")
    CDR_MEDIA_SNR_DB = Threshold(low_threshold_alarm="20", desc="cdr media snr(dB)")
    # 直接信噪比, 约等于56db
    CDR_HOST_SNR_LINE = Threshold(low_threshold_alarm="400000", desc="cdr host snr")
    CDR_MEDIA_SNR_LINE = Threshold(low_threshold_alarm="400000", desc="cdr media snr")
