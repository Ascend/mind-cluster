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
import re

from ascend_fd_tk.core.common.json_obj import JsonObj


class BmcErrCodeHardwareInfo(JsonObj):

    def __init__(self, npu: str = "", chip: str = ""):
        self.npu = npu
        self.chip = chip


class BmcErrCodeEvent(JsonObj):

    def __init__(self, err_code: int, err_desc: str, handle_suggestion: str, err_info_pattern: re.Pattern = None,
                 keywords: str = ""):
        self.err_code = err_code
        self.err_desc = err_desc
        self.handle_suggestion = handle_suggestion
        self.err_info_pattern = err_info_pattern
        self.keywords = keywords


_HARDWARE_PATTERN = re.compile(r"NPU Board(?P<npu>\d{1,2})( NPU\d-(?P<chip>\d{1,2}))?|NPU(?P<npu1>\d{1,2})|AI Module("
                               r"?P<npu2>\d{1,2})")

_ERR_CODE_EVENT_LIST = [
    BmcErrCodeEvent(0x80e01801, "发生多Bit ECC故障", "请对该NPU HBM进行压测", _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x80e18402, "多Bit ECC故障, 隔离行已满64", "请立即更换NPU备件", _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x80cb800a, "AIV算子超时, NPU热复位", "建议对硬件做AICode压测", _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x80cb8009, "AIV总线访问错误", "建议对硬件做AICode压测", _HARDWARE_PATTERN),

    # 掉卡故障
    BmcErrCodeEvent(0x56000003, "NPU健康状态紧急告警",
                    "1、检查芯片温度是否过高（可能是散热异常、环境温度过高或进风口/出风口堵塞）；\n"
                    "2、或者检查是否是地址异常或者内存泄漏等软件问题。\n"
                    "3、若无法解决，请联系技术支持", _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x56000005, "NPU connection has been lost告警",
                    "发生掉卡故障，请联系运维处理", _HARDWARE_PATTERN),

    BmcErrCodeEvent(0x2C000007, "AI模组异常下电告警，过流保护异常下电（V_1V2_DVDD_HBM02_FIX）",
                    "AC重启，不恢复则更换对应模组", _HARDWARE_PATTERN, "V_1V2_DVDD_HBM02_FIX"),

    BmcErrCodeEvent(0x2C000007, "系统异常下电告警，AIC电源过流保护掉电（V_0V9_AIC_DVFS_DA）",
                    "AC重启，不恢复则更换对应模组", _HARDWARE_PATTERN, "V_0V9_AIC_DVFS_DA"),
    BmcErrCodeEvent(0x2C000007, "系统异常下电告警，主板有电压跌落（V_AVDD12_HVCC）",
                    "20A PSIP故障或者12V电容失效，请联系运维处理", _HARDWARE_PATTERN, "V_AVDD12_HVCC"),
    BmcErrCodeEvent(0x2C000007, "NPU异常下电告警，NPU异常掉电（V_AVDD08_LVCC）",
                    "20A PSIP故障或者12V电容失效，请联系运维处理", _HARDWARE_PATTERN, "V_AVDD08_LVCC"),
    BmcErrCodeEvent(0x2C000007, "NPU异常下电告警，电源芯片阻抗异常（MOS器件EOS早期失效），导致掉电（V_0V8_DVDD_SIOE）",
                    "电源芯片阻抗异常，请联系运维处理", _HARDWARE_PATTERN, "V_0V8_DVDD_SIOE"),

    BmcErrCodeEvent(0x2C00002B, "上电超时告警，主板有电压跌落（V_AVDD12_HVCC）",
                    "1、检查外部供电时候满足服务器整机功耗要求；\n"
                    "2、或者通过拔插电源线或拔插单板，将服务器彻底下电再上电，检查告警是否清楚；\n"
                    "3、若无法解决，请联系技术支持更换可能涉及的部件",
                    _HARDWARE_PATTERN, "V_AVDD12_HVCC"),
    BmcErrCodeEvent(0x5D000005, "PSU过温告警", "PSU温度过高，请联系运维处理", _HARDWARE_PATTERN),

    BmcErrCodeEvent(0x5D00001D, "54V上电超时告警，NPU的54V链路上的某器件失效（54V0_HAM）",
                    "NPU的54V链路上的某器件失效，请联系运维处理", _HARDWARE_PATTERN, "54V0_HAM"),

    BmcErrCodeEvent(0x5D00001D, "上电超时告警，6A PSIP故障或者12V电容失效（V_DVDD25_2V5_HBM_FIX）",
                    "6A PSIP故障或者12V电容失效，请联系运维处理", _HARDWARE_PATTERN, "V_DVDD25_2V5_HBM_FIX"),
    BmcErrCodeEvent(0x5D00001D, "上电超时告警，NPU异常掉电（V_DVDD075_HBMPHY_FIX）",
                    "20A PSIP故障或者12V电容失效，请联系运维处理", _HARDWARE_PATTERN,
                    "V_DVDD075_HBMPHY_FIX"),
    BmcErrCodeEvent(0x5D00001D, "上电超时告警，NPU异常掉电（V_AVDD08_LVCC）",
                    "20A PSIP故障或者12V电容失效，请联系运维处理", _HARDWARE_PATTERN, "V_AVDD08_LVCC"),

    BmcErrCodeEvent(0x5D00001F, "NPU异常掉电告警，负载剧烈变化导致保护性下电（V_DVDD09_BUS_DVFS）",
                    "AC重启，不恢复则更换对应模组", _HARDWARE_PATTERN, "V_DVDD09_BUS_DVFS"),
    BmcErrCodeEvent(0x5D00001F, "NPU异常掉电告警，模组PSU过温导致12V掉电（PG_12V0_）",
                    "NPU异常掉电告警，请联系运维处理", _HARDWARE_PATTERN, "PG_12V0_"),
    BmcErrCodeEvent(0x5D00001F, "NPU异常掉电告警，6A PSIP故障或者12V电容失效（V_DVDD25_2V5_HBM_FIX）",
                    "6A PSIP故障或者12V电容失效，请联系运维处理", _HARDWARE_PATTERN, "V_DVDD25_2V5_HBM_FIX"),
    BmcErrCodeEvent(0x5D00001F, "NPU异常下电告警，NPU异常掉电（V_DVDD075_HBMPHY_FIX）",
                    "20A PSIP故障或者12V电容失效，请联系运维处理", _HARDWARE_PATTERN,
                    "V_DVDD075_HBMPHY_FIX"),
    BmcErrCodeEvent(0x5D00001F, "NPU异常掉电告警，NPU的54V链路上的某器件失效（PG_54V0_HAM）",
                    "NPU的54V链路上的某器件失效，请联系运维处理", _HARDWARE_PATTERN, "PG_54V0_HAM"),
    BmcErrCodeEvent(0x5D00001F, "异常掉电告警，模组电压异常掉电（V_DRMOS）",
                    "电池砖高温触发保护掉电，请联系运维处理", _HARDWARE_PATTERN, "V_DRMOS"),

    BmcErrCodeEvent(0x56000009, "NPU 过热关机", "NPU 过热关机，请联系运维处理", _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x12000023, "液冷装置发生漏液", "液冷装置发生漏液，请联系运维处理", _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x120000C3, "液冷装置(LAAC)异常，液冷泵不在位", "液冷泵不在位，请联系运维处理",
                    _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x120000C7, "液冷装置(LAAC)异常，液冷泵转速异常", "液冷泵转速异常，请联系运维处理",
                    _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x120000C9, "液冷装置(LAAC)异常，液冷泵故障", "液冷泵故障，请联系运维处理",
                    _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x04000005, "风冷散热模块故障，风扇冗余失效", "风扇冗余失效，请联系运维处理",
                    _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x18000003, "风冷散热模块故障，风扇背板电源故障", "风扇背板电源故障，请联系运维处理",
                    _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x1800000D, "风冷散热模块故障，风扇背板 MCU 自检异常",
                    "风扇背板 MCU 自检异常，请联系运维处理", _HARDWARE_PATTERN),
    BmcErrCodeEvent(0x04000007, "风冷散热模块故障，风扇转速偏差大", "风扇转速偏差大，请联系运维处理",
                    _HARDWARE_PATTERN),
]
