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

HISI_LOG_DIR = "hisi_logs"
SLOG_DIR = "slog"

BEFORE_TASK_DIR = "before_task"
AFTER_TASK_DIR = "after_task"

HISI_OS_INFO_FILE_NAME = "os_info.txt"
HISI_DEVICE_INFO_FILE_NAME = "device_info.txt"

HIST_DEVICE_OS_PATH_ARRAY = ("log", "slog", "debug", "device_os.log")

# PRECHECK 事件码：hisi kernel.log 中 Kernel panic。1 分钟窗口内同时出现
# "panic reason: softlockup" 用 PRECHECK_KERNEL_PANIC_SOFTLOCKUP，否则用 PRECHECK_KERNEL_PANIC
PRECHECK_KERNEL_PANIC = "PRECHECK_Hisi_log_kernel_panic"
PRECHECK_KERNEL_PANIC_SOFTLOCKUP = "PRECHECK_Hisi_log_kernel_panic_softlockup"
# PRECHECK 事件码：hisi_logs 下 bbox_info.txt 中 system exception code 异常行（如 OS_OOM）
PRECHECK_BBOX_OS_OOM_EXCEPTION = "PRECHECK_bbox_info_os_oom_exception"
# PRECHECK 事件码：hisi_logs 下 os_info.txt 中同名 system exception 行，与 bbox_info.txt 区分
PRECHECK_OS_INFO_OS_OOM_EXCEPTION = "PRECHECK_os_info_os_oom_exception"
# PRECHECK 事件码：hisi kernel.log 中 SMMU CMD_SYNC timeout（svm_smmu controller）
PRECHECK_SMMU_TIMEOUT = "PRECHECK_hisi_log_smmu_cmd_sync_timeout"
# PRECHECK 事件码：hisi kernel.log 中 error_type: 3 的 multi-bit ECC
# 与 sel.txt 系多比特事件（PRECHECK_SEL_MULTIBIT_ECC / PRECHECK_SEL_AIC_MULTIBIT_ECC）区分
PRECHECK_KERNEL_ERR_TYPE3_MULTI_BIT_ECC = "PRECHECK_kernel_log_err_type3_multi_bit_ecc"
# PRECHECK 事件码：hisi kernel.log 中 "oom ... huge page memory is exhausted"
PRECHECK_KERNEL_OOM_HUGE_PAGE_EXHAUSTED = "PRECHECK_kernel_log_oom_huge_page_exhausted"
# PRECHECK 事件码：hisi kernel.log 中 "Unable ... kernel execute ... non-executable memory at virtual address"
PRECHECK_KERNEL_EXECUTE_NON_EXECUTABLE = "PRECHECK_kernel_log_execute_non_executable_memory"
# PRECHECK 事件码：hisi kernel.log 中 "Watchdog detected hard LOCKUP" / "panic reason:Core LOCKUP"
PRECHECK_KERNEL_HARD_LOCKUP = "PRECHECK_kernel_log_watchdog_hard_lockup"
PRECHECK_KERNEL_CORE_LOCKUP = "PRECHECK_kernel_log_panic_core_lockup"
# PRECHECK 事件码：hisi_logs/device_info.txt 中状态为 "heartbeat lost" 的卡
PRECHECK_DEVICE_INFO_HEARTBEAT_LOST = "PRECHECK_device_info_heartbeat_lost"
# PRECHECK 事件码：hisi kernel.log 中 watchdog stuck + rlock SOFT-WATCHDOG 1 分钟内配对
PRECHECK_KERNEL_WATCHDOG_TIMEOUT = "PRECHECK_kernel_log_watchdog_stuck_timeout"
# PRECHECK 事件码：BMC sel.txt 掉卡 / 多bit ECC（0x56000003+0x80e01801 健康紧急 / 0x80c98008 AIC）
PRECHECK_SEL_LOG_LOST_NPU = "PRECHECK_SEL_LOG_LOST_NPU"
PRECHECK_SEL_MULTIBIT_ECC = "PRECHECK_SEL_MULTIBIT_ECC"
PRECHECK_SEL_AIC_MULTIBIT_ECC = "PRECHECK_SEL_AIC_MULTIBIT_ECC"
# PRECHECK 事件码：BMC fdm_log 中 dmc main/mirror 主副空间同时故障、MATA 模块不可纠正错误（UER/UEO/UEU）
PRECHECK_FDM_DMC_MAIN_ERROR = "PRECHECK_FDM_DMC_MAIN_ERROR"
PRECHECK_FDM_DMC_MIRROR_ERROR = "PRECHECK_FDM_DMC_MIRROR_ERROR"
PRECHECK_FDM_MATA_UNCORRECTED_ERROR = "PRECHECK_FDM_MATA_UNCORRECTED_ERROR"
# PRECHECK 事件码：BMC fdm_log 中 MATA 模块 DDRC 多比特 / CPU_CORE CE（AICPU 访问 HBM 报错）
PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC = "PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC"
PRECHECK_FDM_AICPU_HBM_ERROR = "PRECHECK_FDM_AICPU_HBM_ERROR"
# 掉卡类联合判定时间窗（秒）：锚点（BMC sel.txt 掉卡）与证据事件的时间窗阈值，同侧/跨侧共用
LOST_DEV_TIME_WINDOW_SEC = 60
