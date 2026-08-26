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


class BaseThreshold:
    """基础阈值Profile（A3 默认值 + 跨代际公共阈值），各代际统一字段名，子类按需同名覆盖

    与代际无关的阈值（网络状态/lane差值等）只定义在本类，所有代际继承生效；
    代际相关阈值（功率/偏置/SNR等）子类以相同属性名覆盖。
    """

    # ==================== 跨代际公共阈值 ====================
    # cdr snr
    CDR_HOST_SNR_DB = Threshold(low_value_alarm="20", desc="cdr host snr", unit="dB")
    CDR_MEDIA_SNR_DB = Threshold(low_value_alarm="20", desc="cdr media snr", unit="dB")
    # 直接信噪比, 约等于56db
    CHIP_CPU_PORT_SNR_LINE = Threshold(low_value_alarm="290000", desc="CPU与L1间端口 snr", unit="")
    CHIP_NPU_PORT_SNR_LINE = Threshold(low_value_alarm="400000", desc="NPU与L1间端口 snr", unit="")
    SWITCH_PORT_SNR_LINE = Threshold(low_value_alarm="400000", desc="L1与L2间端口 snr", unit="")

    # 网络状态阈值（字符串相等判断）
    # 逻辑：只有等于normal_value_alarm的才是正常，其他均为异常
    DUPLEX_THRESHOLD = Threshold(normal_value_alarm="Full", desc="duplex mode", unit="")  # 只有Full是正常的
    NET_HEALTH_THRESHOLD = Threshold(
        normal_value_alarm="Success", desc="network health", unit=""
    )  # 只有Success是正常的
    LINK_STATUS_THRESHOLD = Threshold(normal_value_alarm="UP", desc="link status", unit="")  # 只有UP是正常的
    OPTICAL_PRESENT_THRESHOLD = Threshold(normal_value_alarm="present", desc="optical module status", unit="")

    # SNR lane间差值阈值
    SNR_LANE_DIFF_DB = Threshold(high_value_alarm="3", desc="snr lane diff", unit="dB")
    # 功率lane间差值阈值
    POWER_LANE_DIFF_DB = Threshold(high_value_alarm="3", desc="power lane diff", unit="dBm")
    # 24小时内NPU链路down次数阈值（超过告警阈值判定亚健康，超过故障阈值判定异常）
    HCCN_LINK_DOWN_CNT = Threshold(high_value_warn="3", high_value_alarm="5", desc="24h内link down次数", unit="次")

    # ==================== A3 默认值（代际阈值，子类按需同名覆盖） ====================
    # 功率阈值(mW)
    TX_POWER_MW = Threshold(low_value_alarm="0.2", high_value_alarm="2.5", desc="tx power", unit="mW")
    RX_POWER_MW = Threshold(
        low_value_alarm="0.1445", low_value_warn="0.6", high_value_alarm="2.3", desc="rx power", unit="mW"
    )
    # 功率阈值(dBm)
    TX_POWER_DBM = Threshold(
        low_value_alarm="-9.60",
        high_value_alarm="7.00",
        low_value_warn="-7.00",
        high_value_warn="5.50",
        desc="tx power",
        unit="dBm",
    )
    RX_POWER_DBM = Threshold(
        low_value_alarm="-10.00",
        high_value_alarm="7.00",
        low_value_warn="-6.50",
        high_value_warn="5.50",
        desc="rx power",
        unit="dBm",
    )
    # 电流
    TX_BIAS_MA = Threshold(low_value_alarm="6", high_value_alarm="10", desc="tx bias", unit="mA")
    # snr
    HOST_SNR_DB = Threshold(low_value_warn="20", low_value_alarm="18", desc="host snr", unit="dB")
    MEDIA_SNR_DB = Threshold(low_value_warn="20", low_value_alarm="18", desc="media snr", unit="dB")


class A5Threshold(BaseThreshold):
    """A5 代际阈值：仅覆盖与 A3 不同的项，相同的继承 BaseThreshold"""

    TX_BIAS_MA = Threshold(
        low_value_alarm="0.5",
        low_value_warn="1",
        high_value_alarm="12",
        high_value_warn="11",
        desc="bias",
        unit="mA",
    )
    # 功率阈值(dBm)与 A3 数值不同，同名覆盖；mW 口径 A5 不使用，继承 Base
    TX_POWER_DBM = Threshold(
        low_value_alarm="-7.5995",
        low_value_warn="-4.6005",
        high_value_alarm="7",
        high_value_warn="5.5",
        desc="tx power",
        unit="dBm",
    )
    RX_POWER_DBM = Threshold(
        low_value_alarm="-9.4006",
        low_value_warn="-6.3997",
        high_value_alarm="7",
        high_value_warn="5.5",
        desc="rx power",
        unit="dBm",
    )

    # ==================== A5 新增网卡（NIC SFP）阈值 ====================
    NIC_TX_BIAS_MA = Threshold(
        low_value_alarm="0.5",
        low_value_warn="1",
        high_value_alarm="12",
        high_value_warn="11",
        desc="nic bias",
        unit="mA",
    )
    NIC_TX_POWER_DBM = Threshold(
        low_value_alarm="-7.5995",
        low_value_warn="-4.6005",
        high_value_alarm="7",
        high_value_warn="5.5",
        desc="nic tx power",
        unit="dBm",
    )
    NIC_RX_POWER_DBM = Threshold(
        low_value_alarm="-9.4006",
        low_value_warn="-6.3997",
        high_value_alarm="7",
        high_value_warn="5.5",
        desc="nic rx power",
        unit="dBm",
    )
    NIC_HOST_SNR_DB = Threshold(low_value_warn="20", low_value_alarm="18", desc="nic host snr", unit="dB")
    NIC_MEDIA_SNR_DB = Threshold(low_value_warn="20", low_value_alarm="18", desc="nic media snr", unit="dB")
    # NIC lane flag 正常值（"0" 为正常，其他为异常）
    NIC_LANE_FLAG = Threshold(normal_value_alarm="0", desc="nic lane flag", unit="")
