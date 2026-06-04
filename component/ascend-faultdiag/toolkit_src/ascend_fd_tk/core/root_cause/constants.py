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

# 根因判定结果
IS_ROOT_CAUSE = "是"
NOT_ROOT_CAUSE = "否"
UNKNOWN_ROOT_CAUSE = "未知"

# 根因部件类型
ROOT_CAUSE_NPU_CPU = "NPU/CPU"
ROOT_CAUSE_CABLE_TRAY = "线缆桥"
ROOT_CAUSE_L1_OPTICAL = "L1光模块"
ROOT_CAUSE_L1_SWITCH_BOARD = "L1交换板"
ROOT_CAUSE_L2_OPTICAL = "L2光模块"
ROOT_CAUSE_L2_SWITCH_BOARD = "L2交换板"

# 链路状态
LINK_STATUS_NORMAL = "正常"
LINK_STATUS_ABNORMAL = "异常"
LINK_STATUS_UNANALYZED = "未分析"

# 端口速率
PORT_SPEED_200G = "200G"
PORT_SPEED_400G = "400G"

# SNR规则名称
RULE_L1_200G_HILINK = "l1_200g_hilink"
RULE_L1_400G_HILINK = "l1_400g_hilink"
RULE_L1_OPTICAL_HOST = "l1_optical_host"
RULE_L1_OPTICAL_MEDIA = "l1_optical_media"
RULE_L2_200G_HILINK = "l2_200g_hilink"
RULE_L2_OPTICAL_MEDIA = "l2_optical_media"

# 信噪比类型
TYPE_HOST_SNR = "host_snr"
TYPE_MEDIA_SNR = "media_snr"

# 芯片类型
NPU = "NPU"
CPU = "CPU"
