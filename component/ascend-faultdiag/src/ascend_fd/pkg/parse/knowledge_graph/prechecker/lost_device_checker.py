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
import copy
import logging

from ascend_fd.pkg.parse.knowledge_graph.prechecker.ub_cqe_checker import Checker
from ascend_fd.utils.constant.dev_log_const import (
    LOST_DEV_TIME_WINDOW_SEC,
    PRECHECK_BBOX_OS_OOM_EXCEPTION,
    PRECHECK_DEVICE_INFO_HEARTBEAT_LOST,
    PRECHECK_FDM_AICPU_HBM_ERROR,
    PRECHECK_FDM_DMC_MAIN_ERROR,
    PRECHECK_FDM_DMC_MIRROR_ERROR,
    PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC,
    PRECHECK_FDM_MATA_UNCORRECTED_ERROR,
    PRECHECK_KERNEL_CORE_LOCKUP,
    PRECHECK_KERNEL_ERR_TYPE3_MULTI_BIT_ECC,
    PRECHECK_KERNEL_EXECUTE_NON_EXECUTABLE,
    PRECHECK_KERNEL_HARD_LOCKUP,
    PRECHECK_KERNEL_OOM_HUGE_PAGE_EXHAUSTED,
    PRECHECK_KERNEL_PANIC,
    PRECHECK_KERNEL_PANIC_SOFTLOCKUP,
    PRECHECK_KERNEL_WATCHDOG_TIMEOUT,
    PRECHECK_OS_INFO_OS_OOM_EXCEPTION,
    PRECHECK_SEL_AIC_MULTIBIT_ECC,
    PRECHECK_SEL_LOG_LOST_NPU,
    PRECHECK_SEL_MULTIBIT_ECC,
    PRECHECK_SMMU_TIMEOUT,
)
from ascend_fd.utils.fault_code import (
    COMP_LOST_DEV_AIC_ECC_CUSTOM,
    COMP_LOST_DEV_AICPU_CE_CUSTOM,
    COMP_LOST_DEV_CORELOCKUP_CUSTOM,
    COMP_LOST_DEV_HARDLOCKUP_CUSTOM,
    COMP_LOST_DEV_HBM_MAIN_MIRROR_CUSTOM,
    COMP_LOST_DEV_HBM_OOM_CUSTOM,
    COMP_LOST_DEV_HBM_THERMAL_OSOOM_CUSTOM,
    COMP_LOST_DEV_ILLEGAL_ADDR_CUSTOM,
    COMP_LOST_DEV_MATA_ECC_CUSTOM,
    COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM,
    COMP_LOST_DEV_SOFTLOCKUP_CUSTOM,
    COMP_LOST_DEV_UCE_PANIC_CUSTOM,
)
from ascend_fd.utils.tool import get_event_occur_times, times_within_window

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class LostDeviceChecker(Checker):
    """掉卡类联合判定基类：多 PRECHECK 事件组合（含锚点）→ Comp_LOST_DEV_*_Custom 码。

    子类需定义：
        out_code: 输出的 Comp 码
        anchor_code: 掉卡锚点 PRECHECK 码（None 表示无需锚点，单条件判定）
        evidence_codes: 需同时存在的 PRECHECK 码（可为空）
        forbid_codes: 出现即否决的 PRECHECK 码（可为空）
        time_window_sec: 锚点与各证据的时间窗（秒），0 表示不校验
    """

    out_code = None
    anchor_code = None
    evidence_codes = ()
    forbid_codes = ()
    # 需存在但不参与时间窗判定的证据码（如 heartbeat lost 无时间戳只有状态）
    no_window_codes = ()
    time_window_sec = 0

    def analyze(self, device_causes):
        if not self.out_code:
            return
        if self.out_code in device_causes:
            return
        precheck = self.single_device_precheck_event
        if self.anchor_code and not precheck.get(self.anchor_code):
            kg_logger.info(
                "LostDeviceChecker: [%s] %s skipped, anchor event %s not exist",
                self.source_device,
                self.out_code,
                self.anchor_code,
            )
            return
        for code in self.evidence_codes:
            if not precheck.get(code):
                kg_logger.info(
                    "LostDeviceChecker: [%s] %s skipped, evidence event %s not exist",
                    self.source_device,
                    self.out_code,
                    code,
                )
                return
        for code in self.forbid_codes:
            if precheck.get(code):
                kg_logger.info(
                    "LostDeviceChecker: [%s] %s skipped, forbid event %s exists",
                    self.source_device,
                    self.out_code,
                    code,
                )
                return
        if self.time_window_sec and not self._time_within_window():
            kg_logger.info(
                "LostDeviceChecker: [%s] %s skipped, no occurrence within %ss window",
                self.source_device,
                self.out_code,
                self.time_window_sec,
            )
            return
        self._build_lost_device_cause(device_causes)

    def _time_within_window(self) -> bool:
        """锚点与证据事件存在一组 occur_time 差值 ≤ time_window_sec 即命中。

        与 diag 阶段共用 tool.times_within_window；按故障码分别解析 occurrence 时间，
        锚点与各证据码各自保留时间组。无锚点（单条件判定）时不校验时间窗。
        """
        # 无锚点（单条件判定）不校验时间窗
        if not self.anchor_code:
            return True
        anchor_times = get_event_occur_times(self.single_device_precheck_event.get(self.anchor_code, {}))
        evidence_times_by_code = {
            code: get_event_occur_times(self.single_device_precheck_event.get(code, {}))
            for code in self.evidence_codes
            if code not in self.no_window_codes
        }
        return times_within_window(anchor_times, evidence_times_by_code, self.time_window_sec)

    def _build_lost_device_cause(self, device_causes):
        """构建并写入掉卡类细分码，events_attribute 装填锚点＋证据事件。"""
        code = self.out_code
        # 先拿 code 的实体属性；kg-config 未注册时回退空属性，避免清洗阶段崩溃
        code_entity = self.schema.get_schema_entity(code) if self.schema else None
        entities_attribute = code_entity.attribute.to_json() if code_entity else {}

        event_codes = []
        if self.anchor_code:
            event_codes.append(self.anchor_code)
        event_codes.extend(self.evidence_codes)
        events_attribute = [copy.deepcopy(self.single_device_precheck_event.get(ec, {})) for ec in event_codes]

        device_causes.update(
            {
                code: {
                    "code": code,
                    "entities_attribute": entities_attribute,
                    "events_attribute": events_attribute,
                    "chains": {},
                }
            }
        )


class HbmMainMirrorLostChecker(LostDeviceChecker):
    """模式1：掉卡前1分钟内 HBM 主空间（DMC）与副空间（MIRROR）同时故障"""

    out_code = COMP_LOST_DEV_HBM_MAIN_MIRROR_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_FDM_DMC_MAIN_ERROR, PRECHECK_FDM_DMC_MIRROR_ERROR)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC


class AicEccLostChecker(LostDeviceChecker):
    """模式4：掉卡且伴生 AI Core 多bit ECC（0x80c98008）"""

    out_code = COMP_LOST_DEV_AIC_ECC_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_SEL_AIC_MULTIBIT_ECC,)


class AicpuCeLostChecker(LostDeviceChecker):
    """模式5：掉卡 + AICPU（CPU_CORE）CE 风暴（掉卡前1分钟内），且无 HBM ECC 告警"""

    out_code = COMP_LOST_DEV_AICPU_CE_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_FDM_AICPU_HBM_ERROR,)
    forbid_codes = (PRECHECK_SEL_MULTIBIT_ECC, PRECHECK_FDM_MATA_UNCORRECTED_ERROR)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC


class UcePanicLostChecker(LostDeviceChecker):
    """模式2：掉卡 + fdm MATA UCE + kernel panic"""

    out_code = COMP_LOST_DEV_UCE_PANIC_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_FDM_MATA_UNCORRECTED_ERROR, PRECHECK_KERNEL_PANIC)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC


class ThermalOomLostChecker(LostDeviceChecker):
    """模式3：掉卡前1分钟内 bbox 过温/OOM + sel 多bit ECC"""

    out_code = COMP_LOST_DEV_HBM_THERMAL_OSOOM_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_BBOX_OS_OOM_EXCEPTION, PRECHECK_SEL_MULTIBIT_ECC)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC


class SmmuTimeoutLostChecker(LostDeviceChecker):
    """模式6：掉卡 + kernel SMMU CMD_SYNC timeout"""

    out_code = COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_SMMU_TIMEOUT,)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC


class MataEccLostChecker(LostDeviceChecker):
    """模式7：掉卡 + fdm MATA DDRC 多bit ECC + kernel error_type3 多bit ECC"""

    out_code = COMP_LOST_DEV_MATA_ECC_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC, PRECHECK_KERNEL_ERR_TYPE3_MULTI_BIT_ECC)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC


class HbmOomLostChecker(LostDeviceChecker):
    """模式10：掉卡 + os_info OOM + kernel oom huge page exhausted"""

    out_code = COMP_LOST_DEV_HBM_OOM_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_OS_INFO_OS_OOM_EXCEPTION, PRECHECK_KERNEL_OOM_HUGE_PAGE_EXHAUSTED)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC


class IllegalAddrLostChecker(LostDeviceChecker):
    """模式11：掉卡 + kernel 执行不可执行内存"""

    out_code = COMP_LOST_DEV_ILLEGAL_ADDR_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_KERNEL_EXECUTE_NON_EXECUTABLE,)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC


class HardlockupLostChecker(LostDeviceChecker):
    """模式12：掉卡 + kernel watchdog hard lockup"""

    out_code = COMP_LOST_DEV_HARDLOCKUP_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_KERNEL_HARD_LOCKUP,)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC


class CorelockupLostChecker(LostDeviceChecker):
    """模式13：掉卡 + kernel panic Core LOCKUP"""

    out_code = COMP_LOST_DEV_CORELOCKUP_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (PRECHECK_KERNEL_CORE_LOCKUP,)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC


class SoftlockupLostChecker(LostDeviceChecker):
    """模式14：heartbeat lost（无时间戳仅判状态，不参与时间窗）+ kernel panic softlockup + watchdog timeout"""

    out_code = COMP_LOST_DEV_SOFTLOCKUP_CUSTOM
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = (
        PRECHECK_DEVICE_INFO_HEARTBEAT_LOST,
        PRECHECK_KERNEL_PANIC_SOFTLOCKUP,
        PRECHECK_KERNEL_WATCHDOG_TIMEOUT,
    )
    no_window_codes = (PRECHECK_DEVICE_INFO_HEARTBEAT_LOST,)
    time_window_sec = LOST_DEV_TIME_WINDOW_SEC
