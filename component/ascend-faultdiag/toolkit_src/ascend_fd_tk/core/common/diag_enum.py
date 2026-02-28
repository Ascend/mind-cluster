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

from enum import Enum, auto, EnumMeta
from typing import TypeVar, Type, Optional


# 设备类型枚举
class DeviceType(Enum):
    SERVER = "服务器"
    BMC = "BMC"
    NPU = "NPU"
    XPU = "XPU"
    SWITCH = "交换机"
    L1_SWITCH = "L1 交换机"
    L2_SWITCH = "L2 交换机"
    ROCE_SWITCH = "ROCE 交换机"
    SWI_CHIP = "交换板芯片"
    SWI_PORT = "交换机端口"
    OPTICAL_MODULE = "光模块"
    ELECTRIC_PORT = "电口"
    TX_PORT = "tx"
    RX_PORT = "rx"
    CHIP = "chip"

    def __str__(self):
        return self.value


class XPU(Enum):
    CPU = "cpu"
    NPU = "npu"


class LinkType(Enum):
    FIBER = "Fiber"  # 光纤
    COPPER = "Copper"  # 铜缆


class SwitchType(Enum):
    L1 = "l1"
    L2 = "l2"
    ROCE = "roce"


class NpuType(Enum):
    A2 = "A2"
    A3 = "A3"


# 设备状态枚举
class DeviceStatus(Enum):
    NORMAL = auto()
    WARNING = auto()
    ERROR = auto()

    def __str__(self):
        return self.name.lower()


class HCCSProxyModule(Enum):
    RP = "10"
    LP = "11"
    VOQ = "1"


class ProxyType(Enum):
    REMOTE_PROXY = auto()
    LOCAL_PROXY = auto()


class ResponseType(Enum):
    TX_TIMEOUT = auto()
    RX_TIMEOUT = auto()


class HccsPackErrorCnt(Enum):
    RP_PACK_STUACK = "rp_id_using_cnt"  # rp窝包
    LP_PACK_STUACK = "lp_id_using_cnt"  # lp窝包
    VOQ_PACK_DROP = "pkt_drp_cnt_l"  # VOQ丢包


# 光模块协议
class OpticalModuleProtXsfpId(Enum):
    QSFP_DD = ["QSFP_DD", "QSFP+ or later"]
    QSFP = ["QSFP", "QSFPPLUS", "QSFP28 or later"]  # 该类协议无法启用回环


class OpticalModuleLoopbackCapabilityBit(Enum):
    HOST_INPUT_LOOPBACK = 0b1000  # 电测内环能力(-t 1)
    MEDIA_OUTPUT_LOOPBACK = 0b0001  # 光侧内环能力(-t 2)
    HOST_OUTPUT_LOOPBACK = 0b0100  # 电测外环能力(-t 3)
    MEDIA_INPUT_LOOPBACK = 0b0010  # 光侧外环能力(-t 4)


class OpticalModuleLoopbackMode(Enum):
    NO_LOOPBACK = "no loopback mode"
    HOST_SIDE_INPUT = "host side input loopback mode"
    MEDIA_SIDE_OUTPUT = "media side output loopback mode"
    HOST_SIDE_OUTPUT = "host side output loopback mode"
    MEDIA_SIDE_INPUT = "media side input loopback mode"


class OpticalLoopbackMode(Enum):
    HOST_SIDE_INPUT = 1
    MEDIA_SIDE_OUTPUT = 2
    HOST_SIDE_OUTPUT = 3
    MEDIA_SIDE_INPUT = 4
    NO_LOOPBACK = 0


class TimeFormat(Enum):
    TYPE_CLOCK = "%Y-%m-%d %H:%M:%S%z"
    DEFAULT_TIME_FMT = "%Y-%m-%d %H:%M:%S:%f"
    BMC_DATE_FMT = "%Y-%m-%d %H:%M:%S"
    BMC_TAR_FILE = "%Y%m%d-%H%M"
    NPU_LINK_STAT_TIME = "%a %b %d %H:%M:%S %Y"


class FaultLevel(Enum):
    ERROR_FAULT = "故障态"
    SUB_ERROR_FAULT = "次故障态"
    UNHEALTH = "亚健康态"


class Customer(Enum):
    Mayi = "mayi"


class CollectType(Enum):
    ALL = auto()
    SSH = auto()
    LOCAL = auto()


class PowerUnitType(Enum):
    DBM = "dBm"
    MW = "mW"


class PowerStatType(Enum):
    # 范围内最大值最小值, 适用于bmc
    MAX = auto()
    MIN = auto()
    # 实时功率
    CUR = auto()


T = TypeVar('T', bound=Enum)


def get_enum(enum_cls: Type[T], name: str = "", value: str = "", case_sensitive: bool = True) -> Optional[T]:
    """
    通过枚举成员的名称（name）获取对应的枚举成员

    参数:
        enum_cls: 枚举类（必须是Enum的子类）
        name: 要查找的枚举成员名称（字符串）
        case_sensitive: 是否区分大小写（默认True，严格匹配大小写）

    返回:
        对应的枚举成员

    异常:
        ValueError: 当枚举类中不存在指定名称的成员时
        TypeError: 当传入的enum_cls不是Enum的子类时
    """
    # 校验输入的枚举类是否合法
    if not isinstance(enum_cls, EnumMeta):
        return None

    # 处理不区分大小写的情况
    if not case_sensitive:
        # 遍历枚举成员，匹配名称（忽略大小写）
        for member in enum_cls:
            if name and member.name.lower() == name.lower():
                return member
            if value and member.value.lower() == value.lower():
                return member
        # 未找到时抛出异常
        return None
    else:
        # 区分大小写：直接使用枚举的__getitem__方法（或EnumMeta.__getitem__）
        try:
            return enum_cls[name]
        except KeyError:
            return None
