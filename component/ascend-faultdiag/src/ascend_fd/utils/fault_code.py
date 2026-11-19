#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
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


# KG_DIAGNOSIS_NORMAL, Knowledge graph diagnosis normally
KG_DIAGNOSIS_NORMAL = "NORMAL_OR_UNSUPPORTED"
# HCCL_SOCKET_FAULT, HCCL get socket error
HCCL_SOCKET_FAULT = "AISW_CANN_HCCL_017"
# HCCL_P2P_FAULT, HCCL get P2P status error
HCCL_P2P_FAULT = "AISW_CANN_HCCL_016"
# HCCL_ROOT_RANK_SOCKET_FAULT, HCCL root rank socket error
HCCL_ROOT_RANK_SOCKET_FAULT = "AISW_CANN_HCCL_004"
# HCCL_NOT_ROOT_RANK_SOCKET_FAULT, HCCL not root rank socket error
HCCL_NOT_ROOT_RANK_SOCKET_FAULT = "AISW_CANN_HCCL_005"
# HCCL_NOT_ROOT_RANK_RECV_FAULT, HCCL not root rank receive error
HCCL_NOT_ROOT_RANK_RECV_FAULT = "AISW_CANN_HCCL_006"
# HCCL_SOCKET_RECV_FAULT, HCCL socket receive error
HCCL_SOCKET_RECV_FAULT = "AISW_CANN_HCCL_018"
# HCCL_NOTIFY_FAULT, HCCL get NOTIFY status error
HCCL_NOTIFY_FAULT = "AISW_CANN_HCCL_027"
# RUNTIME_EVENT_WAIT_FAULT, event wait timeout
RUNTIME_EVENT_WAIT_FAULT = "AISW_CANN_Runtime_020"
# RUNTIME_NOTIFY_WAIT_FAULT, notify wait timeout
RUNTIME_NOTIFY_WAIT_FAULT = "AISW_CANN_Runtime_021"
# RUNTIME_AICORE_EXECUTE_FAULT, AI core execute failed
RUNTIME_AICORE_EXECUTE_FAULT = "AISW_CANN_Runtime_032"
# FRAMEWORK, Training framework report NOTIFY status error
FRAMEWORK_NOTIFY_FAULT = "AISW_CANN_ERRMSG_Custom_03"
# FRAMEWORK, Training framework report gsocket error
FRAMEWORK_SOCKET_FAULT = "AISW_CANN_ERRMSG_Custom_04"
# FRAMEWORK, Training framework report P2P status error
FRAMEWORK_P2P_FAULT = "AISW_CANN_ERRMSG_Custom_05"
# DEVICE_CQE_FAULT, Roce get cqe 0x15 error
DEVICE_CQE_FAULT = "AISW_CANN_HCCL_029"
# LINK_DOWN_FAULT, Network port link down frequently
LINK_DOWN_FAULT = "Comp_Network_Custom_01"
# LINK_STATUS_CHANGE, Network port link status change
LINK_STATUS_CHANGE = "0x81078603"
# CANN_ERRCODE_CUSTOM, CANN custom event
CANN_ERRCODE_CUSTOM = "AISW_CANN_ERRCODE_Custom"
# HBM_HALVED_FAULT HBM halved compared to rated HBM
HBM_ABNORMAL_FAULT = "AISW_CANN_DRV_Custom_01"
# ABNORMAL_FEC_MODE_FAULT, abnormal fec mode detected
ABNORMAL_FEC_MODE_FAULT = "Comp_Network_Custom_02"
# GENERAL_NET_HEALTH_FAULT, a general fault implies net health fault, in which Socket failed, Receive timeout,
# Unreachable and Detect ip set are sub-fault that would be presented finally.
GENERAL_NET_HEALTH_FAULT = "Comp_Network_Custom_03"
# OPTICAL_MODULE_NOT_PRESENT, the optical module of a non-pod board not present
OPTICAL_MODULE_NOT_PRESENT = "Comp_Network_Custom_05"
# NPU_DRIVER_FAULT, the NPU driver is missing if the npu-smi info command is not supported
NPU_DRIVER_FAULT = "Comp_Network_Custom_06"
# IP_NOT_CONFIG_FAULT, no IP address or netmask configured
IP_NOT_CONFIG_FAULT = "Comp_Network_Custom_07"
# OPTICAL_POWER_FAULT, the optical power is abnormal
OPTICAL_POWER_FAULT = "Comp_Network_Custom_08"
# OPTICAL_MODULE_NOT_RX_OR_TX_FAULT, the RX/TX of the optical module does not receive or transmit signals
OPTICAL_MODULE_NOT_RX_OR_TX_FAULT = "Comp_Network_Custom_09"
# OPTICAL_MODULE_OUT_OF_LOCK_FAULT, the receive and transmit signals of the optical module are out of lock
OPTICAL_MODULE_OUT_OF_LOCK_FAULT = "Comp_Network_Custom_10"
# FIBER_OR_COPPER_LINK_FAULT, fiber/copper Link fault
FIBER_OR_COPPER_LINK_FAULT = "Comp_Network_Custom_11"
# PHYSICAL_CARD_DROPPING, physical card dropping detected
PHYSICAL_CARD_DROPPING = "Comp_NPU_DRV_Custom_01"
# SOFTWARE_CARD_DROPPING, software card dropping detected
SOFTWARE_CARD_DROPPING = "Comp_NPU_DRV_Custom_02"
# PRE_TRACEBACK_FAULT, Prefix of fault code of python traceback
PRE_TRACEBACK_FAULT = "AISW_TRACEBACK"
# PRE_SEG_FAULT, Prefix of fault code of segmentation fault
PRE_SEG_FAULT = "AISW_SEG"
# PRE_SWITCH_FAULT, Prefix of fault code of Switch
PRE_SWITCH_FAULT = "Comp_Switch"

# NODE_DIAGNOSIS_NORMAL, node diagnosis normally
NODE_DIAGNOSIS_NORMAL = "NODE_RES_NORMAL"
# NPU_STATUS_ABNORMAL, npu status abnormal
NPU_STATUS_ABNORMAL = "NODE_RES_ABNORMAL_01"
# NPU_OVER_TEMPERATURE, fan speed too low
NPU_OVER_TEMPERATURE = "NODE_RES_ABNORMAL_02"

# NET_DIAGNOSIS_NORMAL, Net congestion diagnosis normally
NET_DIAGNOSIS_NORMAL = "NET_CONGESTION_NORMAL"
# NET_LINK_CONGESTION_FAULT, Some npu links are congested due to conflict
NET_LINK_CONGESTION_FAULT = "NET_CONGESTION_ABNORMAL_01"

# ALL_PROCESS_PREEMPT_FAULT, all process preemption
ALL_PROCESS_PREEMPT_FAULT = "NODE_RES_ABNORMAL_03"
# SINGLE_PROCESS_PREEMPT_FAULT, single process preemption
SINGLE_PROCESS_PREEMPT_FAULT = "NODE_RES_ABNORMAL_04"
# PART_PROCESS_PREEMPT_FAULT, partials process preemption
PART_PROCESS_PREEMPT_FAULT = "NODE_RES_ABNORMAL_05"

# PYTORCH_ERRCODE_COMMON, PYTORCH ERRCODE fault event
PYTORCH_ERRCODE_COMMON = "AISW_PyTorch_ERRCODE_Common"

# MINDIE_ERRCODE_COMMON, MindIE ERRCODE fault event
MINDIE_ERRCODE_COMMON = "AISW_MindIE_ERRCODE_Common"

# CANN fault events about out of memory
AISW_CANN_MEMORY_INFO = "AISW_CANN_Memory_Info_Custom"
AISW_CANN_HDC_SEND_FAILED_FAULT = "AISW_CANN_DRV_HDC_014"
AISW_CANN_HDC_MANY_FILE_FAULT = "AISW_CANN_DRV_HDC_021"
AISW_CANN_GETNEXT_TIMEOUT_FAULT = "AISW_CANN_Runtime_Drv_005"
AISW_CANN_NON_TRANS_FAULT = "AISW_CANN_DRV_PCIE_021"
AISW_CANN_RDMA_QP_FAULT = "AISW_CANN_HCCL_023"
SYSTEM_MEMORY_EXCEED_FAULT = "0x8C2FA001"
AISW_CANN_FAILED_APPLY_MEMORY_FAULT = "AISW_CANN_Runtime_054"

# HCCL FAULT LIST
HCCL_FAULT_LIST = [
    HCCL_SOCKET_FAULT,
    HCCL_P2P_FAULT,
    HCCL_NOTIFY_FAULT,
    FRAMEWORK_NOTIFY_FAULT,
    FRAMEWORK_SOCKET_FAULT,
    FRAMEWORK_P2P_FAULT,
    HCCL_ROOT_RANK_SOCKET_FAULT,
    HCCL_NOT_ROOT_RANK_SOCKET_FAULT,
    HCCL_NOT_ROOT_RANK_RECV_FAULT,
    HCCL_SOCKET_RECV_FAULT,
    RUNTIME_EVENT_WAIT_FAULT,
    RUNTIME_NOTIFY_WAIT_FAULT,
]

NORMAL_CODE_LIST = [KG_DIAGNOSIS_NORMAL, NODE_DIAGNOSIS_NORMAL, NET_DIAGNOSIS_NORMAL]

# NODE_AND_NETWORK_CODE_LIST, node compute fault and network congestion fault code list
NODE_AND_NETWORK_CODE_LIST = [
    NPU_STATUS_ABNORMAL,
    NPU_OVER_TEMPERATURE,
    NET_LINK_CONGESTION_FAULT,
    ALL_PROCESS_PREEMPT_FAULT,
    SINGLE_PROCESS_PREEMPT_FAULT,
    PART_PROCESS_PREEMPT_FAULT,
]

FAULT_WITH_COMPLEMENT_LIST = [GENERAL_NET_HEALTH_FAULT, PHYSICAL_CARD_DROPPING, SOFTWARE_CARD_DROPPING]

OOM_CANN_FAULT_LIST = [
    AISW_CANN_HDC_SEND_FAILED_FAULT,
    AISW_CANN_HDC_MANY_FILE_FAULT,
    AISW_CANN_GETNEXT_TIMEOUT_FAULT,
    AISW_CANN_NON_TRANS_FAULT,
    AISW_CANN_RDMA_QP_FAULT,
    SYSTEM_MEMORY_EXCEED_FAULT,
    AISW_CANN_FAILED_APPLY_MEMORY_FAULT,
]

# 掉卡类故障模式诊断（RFC《掉卡类故障模式诊断能力》）：联合判定产出的细分码
# 模式1 HBM主副空间同时故障（SEL掉卡锚点＋FDM_DMC＋FDM_MIRROR，掉卡前1min内，BMC同侧清洗阶段判定）
COMP_LOST_DEV_HBM_MAIN_MIRROR_CUSTOM = "Comp_LOST_DEV_HBM_MAIN_MIRROR_Custom"
# 模式2 HBM多bit UCE＋Kernel panic（SEL掉卡锚点＋FDM MATA UCE＋KERNEL_PANIC，掉卡前1min内，跨侧diag阶段判定）
COMP_LOST_DEV_UCE_PANIC_CUSTOM = "Comp_LOST_DEV_UCE_PANIC_Custom"
# 模式3 HBM多bit ECC＋热隔离失败伴OS_OOM（SEL掉卡锚点＋sel 0x56000003+0x80e01801＋bbox 0xA4040001，掉卡前1min内，跨侧diag阶段判定）
COMP_LOST_DEV_HBM_THERMAL_OSOOM_CUSTOM = "Comp_LOST_DEV_HBM_THERMAL_OSOOM_Custom"
# 模式4 AI Core多bit ECC（SEL掉卡锚点＋SEL_AIC 0x80c98008，BMC同侧清洗阶段判定）
COMP_LOST_DEV_AIC_ECC_CUSTOM = "Comp_LOST_DEV_AIC_ECC_Custom"
# 模式5 AICPU Cache CE风暴（SEL掉卡锚点＋FDM_AICPU CE＋无HBM ECC告警，掉卡前1min内，BMC同侧清洗阶段判定）
COMP_LOST_DEV_AICPU_CE_CUSTOM = "Comp_LOST_DEV_AICPU_CE_Custom"
# 模式6 SMMU CMD_SYNC timeout（SEL掉卡锚点＋KERNEL svm_smmu CMD_SYNC timeout，掉卡前1min内，跨侧diag阶段判定）
COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM = "Comp_LOST_DEV_SMMU_CMD_TIMEOUT_Custom"
# 模式7 MATA多bit ECC（SEL掉卡锚点＋FDM MATA DDRC多bit ECC＋KERNEL error_type:3 multi-bit，掉卡前1min内，跨侧diag阶段判定）
COMP_LOST_DEV_MATA_ECC_CUSTOM = "Comp_LOST_DEV_MATA_ECC_Custom"
# 模式10 HBM OOM（SEL掉卡锚点＋os_info 0xA4040001＋KERNEL oom huge page，掉卡前1min内，跨侧diag阶段判定）
COMP_LOST_DEV_HBM_OOM_CUSTOM = "Comp_LOST_DEV_HBM_OOM_Custom"
# 模式11 非法踩内存地址（SEL掉卡锚点＋KERNEL non-executable memory，掉卡前1min内，跨侧diag阶段判定）
COMP_LOST_DEV_ILLEGAL_ADDR_CUSTOM = "Comp_LOST_DEV_ILLEGAL_ADDR_Custom"
# 模式12 Device OS HardLockup（SEL掉卡锚点＋KERNEL hard LOCKUP，掉卡前1min内，跨侧diag阶段判定）
COMP_LOST_DEV_HARDLOCKUP_CUSTOM = "Comp_LOST_DEV_HARDLOCKUP_Custom"
# 模式13 Device OS CoreLockup（SEL掉卡锚点＋KERNEL panic Core LOCKUP，掉卡前1min内，跨侧diag阶段判定）
COMP_LOST_DEV_CORELOCKUP_CUSTOM = "Comp_LOST_DEV_CORELOCKUP_Custom"
# 模式14 Device OS Softlockup（SEL掉卡锚点＋DEVICE_INFO heartbeat＋KERNEL panic softlockup＋watchdog，掉卡前1min内，跨侧diag阶段判定）
COMP_LOST_DEV_SOFTLOCKUP_CUSTOM = "Comp_LOST_DEV_SOFTLOCKUP_Custom"

# 掉卡类联合判定产物码集合（同侧清洗阶段 + 跨侧diag阶段）
LOST_DEV_FAULT_LIST = [
    COMP_LOST_DEV_HBM_MAIN_MIRROR_CUSTOM,
    COMP_LOST_DEV_UCE_PANIC_CUSTOM,
    COMP_LOST_DEV_HBM_THERMAL_OSOOM_CUSTOM,
    COMP_LOST_DEV_AIC_ECC_CUSTOM,
    COMP_LOST_DEV_AICPU_CE_CUSTOM,
    COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM,
    COMP_LOST_DEV_MATA_ECC_CUSTOM,
    COMP_LOST_DEV_HBM_OOM_CUSTOM,
    COMP_LOST_DEV_ILLEGAL_ADDR_CUSTOM,
    COMP_LOST_DEV_HARDLOCKUP_CUSTOM,
    COMP_LOST_DEV_CORELOCKUP_CUSTOM,
    COMP_LOST_DEV_SOFTLOCKUP_CUSTOM,
]
