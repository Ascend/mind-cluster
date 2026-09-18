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

import sys

DEFAULT_LOGGER_NAME = "ascend_fd_tk"
DEFAULT_CONSOLE_LOGGER_NAME = "toolkit_console"
PACKAGE_NAME = "ascend-faultdiag-toolkit"

# 闪断关键字
NPU_LINK_DOWN = "LINK DOWN"
NPU_LINK_UP = "LINK UP"
# cmd指令执行成功标识
CMD_EXEC_SUCCESS = "Cmd executed successfully"
# iic错误发生关键字
KEY_IIC_ERROR = "iic_error_check"
# 不可纠误码发生关键字
KEY_UNCORR_CW = "uncorr_cw_cnt_check"
# 闪断关键字
KEY_PCS_LINK = "pcs_link_check"
# 高功率模式开启状态
HIGH_POWER_ENABLE = "enabled"
HIGH_POWER_ENABLE_A5 = "enable"

# 光模块开光状态
OP_TX_DISABLE_STATUS = "0x0"

# hccn_tool 回显中光口固定标识 "Optical"：media_type 取值前缀、光模块信息回显列名
OPTICAL_FLAG = "Optical"

# npu长时down间隔
NPU_LONG_DOWN_TIME = 28
L1_CHIP_NUM = 7
L1_EACH_CHIP_PORT_NUM = 48

# 误码率阈值
BIT_ERROR_RATE_LIMIT = 5.00e-06

# 最大NPU Phy-ID
MAX_NPU_PHY_SIZE = 16
# 不可救无法连续发现的判定时间（单位s）
UNCORR_CW_THRESHOLD = 11 * 60
# 1天（单位s）
ONE_DAY = 60 * 60 * 24

# 进程池最大工作线程数 = CPU核心数 * 进程池CPU利用率系数
CPU_UTILIZATION_RATIO = 0.8

# 常用极限数值
SYS_INT_MAX_SIZE = sys.maxsize
SYS_INT_MIN_SIZE = -sys.maxsize - 1

SYS_FLOAT_MAX_SIZE = float('inf')
SYS_FLOAT_MIN_SIZE = float('-inf')

# 故障所属类型
FAULT_TYPE_BMC = "bmc"
FAULT_TYPE_HOST = "host"
FAULT_TYPE_SWITCH = "switch"

# 通用路径
TOOL_BMC_LOG_COLLECT_DIR_NAME = "dump_info"

# PoDManager 内置槽位表：一个 PoDManager IP 按分片建立多个 SSH 连接，连接内串行切换槽位执行命令
# SFU 槽位 61-64
POD_MANAGER_SFU_SLOT_IDS = ["61", "62", "63", "64"]
# NPU 槽位
POD_MANAGER_NPU_SLOT_IDS = ["18", "19", "20", "21", "23", "24", "25", "26"]
# NPU 槽位内芯片数（chip_id 取值 0-15）
POD_MANAGER_NPU_SLOT_CHIP_NUM = 16
# PoDManager 连接数：槽位均分到各连接，连接内串行执行（可按环境调整）
POD_MANAGER_CONN_NUM = 4
