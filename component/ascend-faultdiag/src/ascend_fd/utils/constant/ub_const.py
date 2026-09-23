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

PRECHECK_PREFIX = "PRECHECK_"

# ubctl 日志目录/文件名
UBCTL_LOG_FILE = "ubctl_log.txt"
UBCTL_DIR = "ubctl"
UB_INFO_DIR = "ub_info"


# 时间格式
TIME_FORMAT = "%Y-%m-%d %H:%M:%S"

# side 取值
SIDE_BEFORE = "before"
SIDE_AFTER = "after"

UBMEM_TIMEOUT_KEYWORD = "set ubmem timeout"
AGE_PERIOD_KEYWORD = "age_period"

# ubctl_log.txt 指标关键字（需要行首匹配时用 f"{KEY}:"）
RC_FULL_QUEUE_CNT = "rc_full_queue_cnt"
TP_RRP_ERR_FLG_0 = "tp_rrp_err_flg_0"
RC_NOT_ENOUGH_0X5 = "rc_not_enough_0x5"
RQE_NOT_ENOUGH_0X5 = "rqe_not_enough_0x5"
RC_NOT_ENOUGH_0X2 = "rc_not_enough_0x2"
TWP_AE_DFX = "twp_ae_dfx"
TAI_COMPACT_ALARM = "tai_compact_alarm"
LQC_TAI_DFX_ALARM = "lqc_tai_dfx_alarm"
DAM_INTF_ALARM = "dam_intf_alarm"
TAI_COMPACT_TOP_ALARM = "tai_compact_top_alarm"
DFX_TM_CRD_CTRL = "dfx_tm_crd_ctrl"
TAACK_ABNORM_SSN = "taack_abnorm_ssn"
TAACK_ABNORM_HEADER = "taack_abnorm_header"

LOST_PKG_UBMEM = "lost_pkg_ubmem"
LOST_PKG_0X5 = "lost_pkg_0x5"
PHY_REINIT_CNT = "phy_reinit_cnt"
UDIE_MAX_PORT_NUM = 9

# -m 参数：dump 为指标采集数据块；其他值表示新类型采集数据，dump 块已读完，整个文件不再解析
UBCTL_DUMP_MODE = "dump"

RX_VL6_PKT_NUM = "rx_vl6_pkt_num"
TX_VL6_PKT_NUM = "tx_vl6_pkt_num"
RX_VL7_PKT_NUM = "rx_vl7_pkt_num"
TX_VL7_PKT_NUM = "tx_vl7_pkt_num"
RX_VL10_PKT_NUM = "rx_vl10_pkt_num"
TX_VL10_PKT_NUM = "tx_vl10_pkt_num"
RX_VL11_PKT_NUM = "rx_vl11_pkt_num"
TX_VL11_PKT_NUM = "tx_vl11_pkt_num"

# lost_pkg 的 vl 指标名
PKT_TX_RX_VL10_VL11_KEYS = (TX_VL10_PKT_NUM, RX_VL10_PKT_NUM, TX_VL11_PKT_NUM, RX_VL11_PKT_NUM)
PKT_TX_RX_VL6_VL7_KEYS = (TX_VL6_PKT_NUM, RX_VL6_PKT_NUM, TX_VL7_PKT_NUM, RX_VL7_PKT_NUM)

# 需要采集的全部指标关键字（统一提取用，行首匹配时用 f"{KEY}:"）
ALL_UBCTL_KEYS = (
    (
        RC_FULL_QUEUE_CNT,
        TP_RRP_ERR_FLG_0,
        TWP_AE_DFX,
        TAI_COMPACT_ALARM,
        LQC_TAI_DFX_ALARM,
        DAM_INTF_ALARM,
        TAI_COMPACT_TOP_ALARM,
        DFX_TM_CRD_CTRL,
        PHY_REINIT_CNT,
        TAACK_ABNORM_SSN,
        TAACK_ABNORM_HEADER,
    )
    + PKT_TX_RX_VL10_VL11_KEYS
    + PKT_TX_RX_VL6_VL7_KEYS
)

# plog 中需要匹配的环境变量日志关键字。日志行示例：
# [HCCL_ENV] HCOMM_TA_CTP_UB_TIMEOUT set by env to [8]
HCOMM_TA_CTP_UB_TIMEOUT = "HCOMM_TA_CTP_UB_TIMEOUT"

PRECHECK_UBCTL_DATA = "PRECHECK_UBCTL_DATA"
# 不重复的 PRECHECK 自定义故障码，按实际需求替换
PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT = "PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT"
# PRECHECK 事件码：diaglog 中 ubmem timeout 配置行，进 root_causes（按实际需求替换）
PRECHECK_UBMEM_TIMEOUT = "PRECHECK_UBMEM_TIMEOUT"
# PRECHECK 事件码：hisi kernel.log 中 UDMA AE 报错行，进 root_causes；不依赖 kg-config.json（按实际需求替换）
PRECHECK_KERNEL_AE_TYPE23 = "PRECHECK_Hisi_log_kernel_ae_type23"
# PRECHECK 事件码：A5 hccn_tool -g -stat 回显中 rxdma icrc 错误计数，仅承载 before/after 原始数据，
# 由规则层 MergePrecheckCause 对比 before/after 判断增长并置 flag（不依赖 kg-config.json）
PRECHECK_RXDMA_ICRC_DATA = "PRECHECK_RXDMA_ICRC_DATA"

# A5 -stat 回显中待观测的 rxdma icrc 错误计数指标名
ICRC_ERR_COUNT_METRICS = ["rxdma_icrc_err_cnt_queue_id%d" % i for i in range(4)]


# rule_flags 的规则 key
RULE_RC_FULL_QUEUE_NONZERO = "rc_full_queue_nonzero"
RULE_TP_RRP_ERR_BIT28 = "tp_rrp_err_bit28"
RULE_TWP_AE_DFX_BIT3 = "twp_ae_dfx_bit3"
RULE_ROUTE_NO_CFG_BIT = "route_no_cfg_bit"
RULE_TAI_COMPACT_TOP_BIT6 = "tai_compact_top_bit6"
RULE_PHY_REINIT_CNT_EXCEED = "phy_reinit_cnt_exceed"
RULE_VLAN10_11_BALANCE = "vlan10_11_balance"
RULE_VLAN6_7_BALANCE = "vlan6_7_balance"
RULE_TAACK_ABNORM_SSN_INCREASE = "taack_abnorm_ssn_increase"
RULE_TAACK_ABNORM_HEADER_INCREASE = "taack_abnorm_header_increase"
RULE_TP_RRP_ERR_BIT25 = "tp_rrp_err_bit25"
RULE_TP_RRP_ERR_BIT27_28 = "tp_rrp_err_bit27_28"
RULE_DFX_TM_CRD_CTRL_INVALID = "dfx_tm_crd_ctrl_invalid"
RULE_LQC_TAI_DFX_ALARM_BIT45 = "lqc_tai_dfx_alarm_bit45"
RULE_HCOMM_TA_CTP_UB_TIMEOUT = "hcomm_ta_ctp_ub_timeout_low"
# PRECHECK ubmem 事件 age_period（微秒）换算秒后小于 4 秒
RULE_UBMEM_TIMEOUT_LOW = "ubmem_timeout_low"
# PRECHECK rxdma icrc 事件 before/after 原始计数对比，任一指标 after > before
RULE_RXDMA_ICRC_ERR_INCREASE = "rxdma_icrc_err_increase"

UB_RAS_CODES = [
    "0x81AF8009",
    "0x81B18603",
    "0x08520006",
    "0x081320C6",
    "0x081320C7",
    "0x08132247",
    "0x08130054",
    "0x0813002e",
    "Net_Param_Dev_david_Chip_01",
    "0x08130056",
    "Net_Param_Dev_5808_Chip_02",
    "Net_Param_Dev_Switch_02",
    "0x08130059",
    "0x08520004",
    "0x8520006",
    "0x81320C6",
    "0x81320C7",
    "0x8132247",
    "0x8130054",
    "0x813002e",
    "0x8130056",
    "0x8130059",
    "0x8520004",
]

Comp_CANN_HCCL_Custom_CQE0x2 = "Comp_CANN_HCCL_Custom_CQE0x2"
Comp_CANN_HCCL_Custom_CQE0x5 = "Comp_CANN_HCCL_Custom_CQE0x5"
REMOTE_MEM_ACCESS_NET_TIMEOUT = "0x81AFAA02"
