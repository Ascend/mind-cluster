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


class OpticalThresholdView:
    """光模块类型阈值视图：属性访问时按 {TYPE}_{METRIC} 优先解析，未定义时回退基础阈值

    用法：诊断侧逐光模块选阈值，view = A5Threshold.get_optical_view(optical_type)，
    之后 view.TX_POWER_DBM 等属性访问方式与 Profile 类一致，可直接传入原有检查逻辑。
    """

    __slots__ = ("_profile", "_optical_type")

    def __init__(self, profile: type, optical_type: str = ""):
        self._profile = profile
        self._optical_type = optical_type

    def __getattr__(self, metric_name: str) -> Threshold:
        return self._profile.get_optical_metric(metric_name, self._optical_type)


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

    # 网络状态阈值（字符串相等判断，只有等于normal_value_alarm的才是正常，其他均为异常）
    DUPLEX_THRESHOLD = Threshold(normal_value_alarm="Full", desc="duplex mode", unit="")  # 只有Full是正常的
    NET_HEALTH_THRESHOLD = Threshold(normal_value_alarm="Success", desc="network health", unit="")
    LINK_STATUS_THRESHOLD = Threshold(normal_value_alarm="UP", desc="link status", unit="")

    # 24小时内NPU链路down次数阈值（超过告警阈值判定亚健康，超过故障阈值判定异常）
    HCCN_LINK_DOWN_CNT = Threshold(high_value_warn="3", high_value_alarm="5", desc="24h内link down次数", unit="次")

    # 光模块在位状态
    OPTICAL_PRESENT_THRESHOLD = Threshold(normal_value_alarm="present", desc="optical module status", unit="")
    # 电流
    TX_BIAS_MA = Threshold(low_value_alarm="6", high_value_alarm="10", desc="tx bias", unit="mA")
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
    # snr
    HOST_SNR_DB = Threshold(low_value_warn="20", low_value_alarm="18", desc="host snr", unit="dB")
    MEDIA_SNR_DB = Threshold(low_value_warn="20", low_value_alarm="18", desc="media snr", unit="dB")
    # SNR lane间差值阈值
    SNR_LANE_DIFF_DB = Threshold(high_value_alarm="3", desc="snr lane diff", unit="dB")
    # 功率lane间差值阈值
    POWER_LANE_DIFF_DB = Threshold(high_value_alarm="3", desc="power lane diff", unit="dBm")

    @classmethod
    def get_optical_metric(cls, metric_name: str, optical_type: str = "") -> Threshold:
        """按光模块类型取指标阈值：类型为 LPO 且Profile定义了 LPO_{METRIC} 时优先，否则使用基础阈值（即 ODSP 阈值）

        规则：optical_type 为空、ODSP 或其他未定义类型化阈值的类型，一律回退基础阈值；
        扩展新类型只需按 {TYPE}_{METRIC} 命名在对应代际Profile上定义差异阈值（如 LPO_TX_POWER_DBM），
        本方法与 threshold_config.json 覆盖机制（配置键即属性名）自动生效，无需改代码。
        """
        type_key = str(optical_type or "").strip().upper()
        if type_key:
            typed_th = getattr(cls, f"{type_key}_{metric_name}", None)
            if isinstance(typed_th, Threshold):
                return typed_th
        return getattr(cls, metric_name)

    @classmethod
    def get_optical_view(cls, optical_type: str = "") -> "OpticalThresholdView":
        """返回按光模块类型解析阈值的视图（属性访问方式与Profile一致，便于逐模块选阈值）"""
        return OpticalThresholdView(cls, optical_type)


class A5Threshold(BaseThreshold):
    """A5 代际阈值：仅覆盖与 A3 不同的项，相同的继承 BaseThreshold"""

    TX_BIAS_MA = Threshold(
        low_value_alarm="4.00",
        low_value_warn="6.00",
        high_value_alarm="13.00",
        high_value_warn="11.00",
        desc="tx bias",
        unit="mA",
    )
    # 功率阈值(dBm)与 A3 数值不同，同名覆盖；mW 口径 A5 不使用，继承 Base
    RX_POWER_DBM = Threshold(
        low_value_alarm="-9.40",
        low_value_warn="-6.40",
        high_value_alarm="7.00",
        high_value_warn="5.50",
        desc="rx power",
        unit="dBm",
    )
    TX_POWER_DBM = Threshold(
        low_value_alarm="-7.60",
        low_value_warn="-4.60",
        high_value_alarm="7.00",
        high_value_warn="5.50",
        desc="tx power",
        unit="dBm",
    )
    # snr
    HOST_SNR_DB = Threshold(low_value_warn="19", low_value_alarm="18", desc="host snr", unit="dB")
    MEDIA_SNR_DB = Threshold(low_value_warn="19", low_value_alarm="18", desc="media snr", unit="dB")
    # 功率lane间差值阈值
    POWER_LANE_DIFF_DB = Threshold(high_value_alarm="5", desc="power lane diff", unit="dBm")

    # ==================== A5 光模块类型阈值（LPO，命名约定 {TYPE}_{METRIC}） ====================
    # 同一集群可能混插不同类型光模块（ODSP/LPO），诊断时按模块上报的 optical_type 选择阈值，
    # 取值逻辑见 BaseThreshold.get_optical_metric：
    #   - ODSP 为默认类型，直接使用本类基础阈值（如 TX_BIAS_MA），无需重复定义；
    #   - LPO 仅定义与默认值有差异的指标，未定义的指标回退基础阈值。LPO光模块不支持 SNR。
    # 扩展新类型只需按 {TYPE}_{METRIC} 命名定义差异阈值即可自动生效；
    # 运行期可用 threshold_config.json 按同名键（如 "LPO_TX_BIAS_MA"）覆盖。
    LPO_TX_BIAS_MA = Threshold(
        low_value_alarm="4.00",
        low_value_warn="6.50",
        high_value_alarm="11.00",
        high_value_warn="9.50",
        desc="lpo tx bias",
        unit="mA",
    )
    LPO_RX_POWER_DBM = Threshold(
        low_value_alarm="-7.40",
        low_value_warn="-4.40",
        high_value_alarm="7.50",
        high_value_warn="6.00",
        desc="lpo rx power",
        unit="dBm",
    )
    LPO_TX_POWER_DBM = Threshold(
        low_value_alarm="-5.70",
        low_value_warn="-2.70",
        high_value_alarm="7.50",
        high_value_warn="6.00",
        desc="lpo tx power",
        unit="dBm",
    )
    LPO_POWER_LANE_DIFF_DB = Threshold(high_value_alarm="3", desc="lpo power lane diff", unit="dBm")

    # ==================== A5 新增网卡（NIC SFP）阈值 ====================
    NIC_TX_BIAS_MA = Threshold(
        low_value_alarm="0.5",
        low_value_warn="1",
        high_value_alarm="12",
        high_value_warn="11",
        desc="nic tx bias",
        unit="mA",
    )
    NIC_RX_POWER_DBM = Threshold(
        low_value_alarm="-9.4006",
        low_value_warn="-6.3997",
        high_value_alarm="7",
        high_value_warn="5.5",
        desc="nic rx power",
        unit="dBm",
    )
    NIC_TX_POWER_DBM = Threshold(
        low_value_alarm="-7.5995",
        low_value_warn="-4.6005",
        high_value_alarm="7",
        high_value_warn="5.5",
        desc="nic tx power",
        unit="dBm",
    )
    NIC_HOST_SNR_DB = Threshold(low_value_warn="19", low_value_alarm="18", desc="nic host snr", unit="dB")
    NIC_MEDIA_SNR_DB = Threshold(low_value_warn="19", low_value_alarm="18", desc="nic media snr", unit="dB")
    # NIC lane flag 正常值（"0" 为正常，其他为异常）
    NIC_LANE_FLAG = Threshold(normal_value_alarm="0", desc="nic lane flag", unit="")
