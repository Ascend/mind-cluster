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

from ascend_fd_tk.core.model.threshold import Threshold


class OpticalModuleThreshold:
    # 功率阈值(mW)
    TX_POWER_THRESHOLD_CONFIG_MW = Threshold(
        low_threshold_alarm="0.2", high_threshold_alarm="2.5", desc="tx power", unit="mW"
    )
    RX_POWER_THRESHOLD_CONFIG_MW = Threshold(
        low_threshold_alarm="0.1445", low_threshold_warn="0.6", high_threshold_alarm="2.3", desc="rx power", unit="mW"
    )
    # 功率阈值(dBm)
    TX_POWER_THRESHOLD_CONFIG_DBM = Threshold(
        low_threshold_alarm="-9.60",
        high_threshold_alarm="7.00",
        low_threshold_warn="-7.00",
        high_threshold_warn="5.50",
        desc="tx power",
        unit="dBm",
    )
    RX_POWER_THRESHOLD_CONFIG_DBM = Threshold(
        low_threshold_alarm="-10.00",
        high_threshold_alarm="7.00",
        low_threshold_warn="-6.50",
        high_threshold_warn="5.50",
        desc="rx power",
        unit="dBm",
    )
    # 电流
    TX_BIAS_MA = Threshold(low_threshold_alarm="6", high_threshold_alarm="10", desc="tx bias", unit="mA")
    # snr
    HOST_SNR_DB = Threshold(low_threshold_warn="20", low_threshold_alarm="18", desc="host snr", unit="dB")
    MEDIA_SNR_DB = Threshold(low_threshold_warn="20", low_threshold_alarm="18", desc="media snr", unit="dB")
    # cdr snr
    CDR_HOST_SNR_DB = Threshold(low_threshold_alarm="20", desc="cdr host snr", unit="dB")
    CDR_MEDIA_SNR_DB = Threshold(low_threshold_alarm="20", desc="cdr media snr", unit="dB")
    # 直接信噪比, 约等于56db
    CDR_HOST_SNR_LINE = Threshold(low_threshold_alarm="400000", desc="cdr host snr", unit="")
    CDR_MEDIA_SNR_LINE = Threshold(low_threshold_alarm="400000", desc="cdr media snr", unit="")

    # 网络状态阈值（字符串相等判断）
    # 逻辑：只有等于normal_value_alarm的才是正常，其他均为异常
    DUPLEX_THRESHOLD = Threshold(normal_value_alarm="Full", desc="duplex mode", unit="")  # 只有Full是正常的
    NET_HEALTH_THRESHOLD = Threshold(
        normal_value_alarm="Success", desc="network health", unit=""
    )  # 只有Success是正常的
    LINK_STATUS_THRESHOLD = Threshold(normal_value_alarm="UP", desc="link status", unit="")  # 只有UP是正常的
    OPTICAL_PRESENT_THRESHOLD = Threshold(normal_value_alarm="present", desc="optical module status", unit="")

    # A5 代际光模块阈值
    A5_TX_BIAS_MA = Threshold(
        low_threshold_alarm="0.5",
        low_threshold_warn="1",
        high_threshold_alarm="12",
        high_threshold_warn="11",
        desc="bias",
        unit="mA",
    )
    A5_TX_POWER_DBM = Threshold(
        low_threshold_alarm="-7.5995",
        low_threshold_warn="-4.6005",
        high_threshold_alarm="7",
        high_threshold_warn="5.5",
        desc="tx power",
        unit="dBm",
    )
    A5_RX_POWER_DBM = Threshold(
        low_threshold_alarm="-9.4006",
        low_threshold_warn="-6.3997",
        high_threshold_alarm="7",
        high_threshold_warn="5.5",
        desc="rx power",
        unit="dBm",
    )
    A5_HOST_SNR_DB = Threshold(low_threshold_warn="20", low_threshold_alarm="18", desc="host snr", unit="dB")
    A5_MEDIA_SNR_DB = Threshold(low_threshold_warn="20", low_threshold_alarm="18", desc="media snr", unit="dB")

    # A5 代际 NIC SFP 阈值
    NIC_TX_BIAS_MA = Threshold(
        low_threshold_alarm="0.5",
        low_threshold_warn="1",
        high_threshold_alarm="12",
        high_threshold_warn="11",
        desc="nic bias",
        unit="mA",
    )
    NIC_TX_POWER_DBM = Threshold(
        low_threshold_alarm="-7.5995",
        low_threshold_warn="-4.6005",
        high_threshold_alarm="7",
        high_threshold_warn="5.5",
        desc="nic tx power",
        unit="dBm",
    )
    NIC_RX_POWER_DBM = Threshold(
        low_threshold_alarm="-9.4006",
        low_threshold_warn="-6.3997",
        high_threshold_alarm="7",
        high_threshold_warn="5.5",
        desc="nic rx power",
        unit="dBm",
    )
    NIC_HOST_SNR_DB = Threshold(low_threshold_warn="20", low_threshold_alarm="18", desc="nic host snr", unit="dB")
    NIC_MEDIA_SNR_DB = Threshold(low_threshold_warn="20", low_threshold_alarm="18", desc="nic media snr", unit="dB")
    # NIC lane flag 正常值（"0" 为正常，其他为异常）
    NIC_LANE_FLAG = Threshold(normal_value_alarm="0", desc="nic lane flag", unit="")
