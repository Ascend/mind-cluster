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
"""Unit tests for combined determination of device-loss fault modes (1-14, excluding the integrated FPGA/UCMI).

Covers the 12 LostDeviceChecker subclasses (modes 1-7, 10-14) in
prechecker/lost_device_checker.py, verifying the "anchor + evidence event + time window /
forbid mutual exclusion" hit paths and the various rejection paths
(missing anchor / missing evidence / out of window / forbid present).
"""

import unittest

from ascend_fd.pkg.parse.knowledge_graph.prechecker.merge_precheck_cause import MergePrecheckCause
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
from ascend_fd.utils.load_kg_config import Schema, SchemaEntity
from ascend_fd.pkg.parse.knowledge_graph.prechecker.lost_device_checker import LostDeviceChecker

# Anchor base time for device loss: evidence in window at +30s, out of window at +2min
ANCHOR_TIME = "2026-01-01 00:00:00"
IN_WINDOW_TIME = "2026-01-01 00:00:30"
OUT_WINDOW_TIME = "2026-01-01 00:02:00"


def make_schema():
    """Build a minimal schema: register all device-loss output codes for easy checking of entities_attribute."""
    schema = Schema(config_pkg_list=[])
    for code in (
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
    ):
        schema.add_custom_event_to_schema(
            code,
            SchemaEntity(
                code,
                {
                    "class": "3",
                    "component": "NPU",
                    "module": "NPU",
                    "cause_zh": "故障_" + code,
                    "cause_en": "fault_" + code,
                    "description_zh": code,
                    "description_en": code,
                },
            ),
        )
    return schema


def lost_dev_event(event_code, source_device="0", occur_time=None, occurrence=None):
    """Build a PRECHECK event (occur_time may be None, simulating timestamp-less events such as heartbeat lost)."""
    event = {
        "event_code": event_code,
        "source_device": source_device,
        "is_custom_event": False,
        "type": "NPU",
        "source_file": "",
        "key_info": "precheck " + event_code,
        "occurrence": occurrence or [],
    }
    if occur_time is not None:
        event["occur_time"] = occur_time
    return event


class LostDeviceCheckerTestBase(unittest.TestCase):
    """Common device-loss joint determination cases (hit / missing anchor / missing evidence / out of window / forbid)."""

    out_code = None
    anchor_code = PRECHECK_SEL_LOG_LOST_NPU
    evidence_codes = ()
    forbid_codes = ()
    window_sec = 0
    # Evidence codes present but excluded from time-window checks (no timestamp, e.g., heartbeat lost)
    no_window_codes = ()
    # Time used for each evidence on in-window hit (default anchor +30s)
    in_window_time = IN_WINDOW_TIME

    def setUp(self):
        if not self.out_code:
            self.skipTest("abstract base class, no out_code")
        self.schema = make_schema()
        self.source_device = "0"

    def run_analyze(self, events):
        device_causes = {}
        merge = MergePrecheckCause(self.schema, {"0": events})
        merge.single_device_analyze(self.source_device, device_causes)
        return device_causes

    def _full_events(self):
        anchor = lost_dev_event(self.anchor_code, occur_time=ANCHOR_TIME)
        evidence = [
            lost_dev_event(code, occur_time=self.in_window_time)
            if code not in self.no_window_codes
            else lost_dev_event(code)
            for code in self.evidence_codes
        ]
        return [anchor] + evidence

    def assert_full_hit(self, result):
        self.assertIn(self.out_code, result)
        entry = result[self.out_code]
        self.assertEqual(entry["code"], self.out_code)
        # Events loaded into anchor + evidence (no_window codes are not subject to the time window but are loaded as evidence)
        expected_events = [self.anchor_code] + list(self.evidence_codes)
        self.assertEqual([e["event_code"] for e in entry["events_attribute"]], expected_events)

    def test_full_hit(self):
        """Anchor + all evidence (in window) present: emit the device-loss refined code."""
        result = self.run_analyze(self._full_events())
        self.assert_full_hit(result)

    def test_missing_anchor(self):
        """Missing device-loss anchor: no refined code is emitted."""
        events = [lost_dev_event(code, occur_time=self.in_window_time) for code in self.evidence_codes]
        result = self.run_analyze(events)
        self.assertNotIn(self.out_code, result)

    def test_missing_anchor_only_event(self):
        """Anchor only, no evidence: no refined code is emitted."""
        result = self.run_analyze([lost_dev_event(self.anchor_code, occur_time=ANCHOR_TIME)])
        self.assertNotIn(self.out_code, result)

    def test_missing_evidence(self):
        """Any evidence event missing: no refined code is emitted."""
        full = self._full_events()
        for drop_code in self.evidence_codes:
            events = [e for e in full if e["event_code"] != drop_code]
            result = self.run_analyze(events)
            self.assertNotIn(self.out_code, result, "should not hit when %s missing" % drop_code)

    def test_out_of_window(self):
        """Out of time window (> time_window_sec): no refined code is emitted. Only valid for modes with a time window.

        Modes without a time window (e.g., AIC ECC) are not time-constrained and are skipped here.
        """
        if not self.window_sec:
            self.skipTest("no time window for this mode")
        anchor = lost_dev_event(self.anchor_code, occur_time=ANCHOR_TIME)
        evidence = [
            lost_dev_event(code, occur_time=OUT_WINDOW_TIME)
            if code not in self.no_window_codes
            else lost_dev_event(code)
            for code in self.evidence_codes
        ]
        result = self.run_analyze([anchor] + evidence)
        self.assertNotIn(self.out_code, result)

    def test_forbid_event(self):
        """Forbid evidence present: reject, no refined code is emitted."""
        if not self.forbid_codes:
            self.skipTest("no forbid codes for this mode")
        events = self._full_events() + [lost_dev_event(self.forbid_codes[0], occur_time=self.in_window_time)]
        result = self.run_analyze(events)
        self.assertNotIn(self.out_code, result)


class TestHbmMainMirrorLostChecker(LostDeviceCheckerTestBase):
    """Mode 1: HBM main space (DMC) and mirror space (MIRROR) both fault within 1 minute before device loss"""

    out_code = COMP_LOST_DEV_HBM_MAIN_MIRROR_CUSTOM
    evidence_codes = (PRECHECK_FDM_DMC_MAIN_ERROR, PRECHECK_FDM_DMC_MIRROR_ERROR)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestUcePanicLostChecker(LostDeviceCheckerTestBase):
    """Mode 2: device loss + fdm MATA UCE + kernel panic"""

    out_code = COMP_LOST_DEV_UCE_PANIC_CUSTOM
    evidence_codes = (PRECHECK_FDM_MATA_UNCORRECTED_ERROR, PRECHECK_KERNEL_PANIC)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestThermalOomLostChecker(LostDeviceCheckerTestBase):
    """Mode 3: bbox over-temperature/OOM + sel multi-bit ECC within 1 minute before device loss"""

    out_code = COMP_LOST_DEV_HBM_THERMAL_OSOOM_CUSTOM
    evidence_codes = (PRECHECK_BBOX_OS_OOM_EXCEPTION, PRECHECK_SEL_MULTIBIT_ECC)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestAicEccLostChecker(LostDeviceCheckerTestBase):
    """Mode 4: device loss accompanied by AI Core multi-bit ECC (0x80c98008), no time window"""

    out_code = COMP_LOST_DEV_AIC_ECC_CUSTOM
    evidence_codes = (PRECHECK_SEL_AIC_MULTIBIT_ECC,)
    window_sec = 0


class TestAicpuCeLostChecker(LostDeviceCheckerTestBase):
    """Mode 5: device loss + AICPU (CPU_CORE) CE storm, with no HBM ECC alarm (forbid)"""

    out_code = COMP_LOST_DEV_AICPU_CE_CUSTOM
    evidence_codes = (PRECHECK_FDM_AICPU_HBM_ERROR,)
    forbid_codes = (PRECHECK_SEL_MULTIBIT_ECC, PRECHECK_FDM_MATA_UNCORRECTED_ERROR)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestSmmuTimeoutLostChecker(LostDeviceCheckerTestBase):
    """Mode 6: device loss + kernel SMMU CMD_SYNC timeout"""

    out_code = COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM
    evidence_codes = (PRECHECK_SMMU_TIMEOUT,)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestMataEccLostChecker(LostDeviceCheckerTestBase):
    """Mode 7: device loss + fdm MATA DDRC multi-bit ECC + kernel error_type3 multi-bit ECC"""

    out_code = COMP_LOST_DEV_MATA_ECC_CUSTOM
    evidence_codes = (PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC, PRECHECK_KERNEL_ERR_TYPE3_MULTI_BIT_ECC)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestHbmOomLostChecker(LostDeviceCheckerTestBase):
    """Mode 10: device loss + os_info OOM + kernel oom huge page exhausted"""

    out_code = COMP_LOST_DEV_HBM_OOM_CUSTOM
    evidence_codes = (PRECHECK_OS_INFO_OS_OOM_EXCEPTION, PRECHECK_KERNEL_OOM_HUGE_PAGE_EXHAUSTED)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestIllegalAddrLostChecker(LostDeviceCheckerTestBase):
    """Mode 11: device loss + kernel executes non-executable memory"""

    out_code = COMP_LOST_DEV_ILLEGAL_ADDR_CUSTOM
    evidence_codes = (PRECHECK_KERNEL_EXECUTE_NON_EXECUTABLE,)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestHardlockupLostChecker(LostDeviceCheckerTestBase):
    """Mode 12: device loss + kernel watchdog hard lockup"""

    out_code = COMP_LOST_DEV_HARDLOCKUP_CUSTOM
    evidence_codes = (PRECHECK_KERNEL_HARD_LOCKUP,)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestCorelockupLostChecker(LostDeviceCheckerTestBase):
    """Mode 13: device loss + kernel panic Core LOCKUP"""

    out_code = COMP_LOST_DEV_CORELOCKUP_CUSTOM
    evidence_codes = (PRECHECK_KERNEL_CORE_LOCKUP,)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestSoftlockupLostChecker(LostDeviceCheckerTestBase):
    """Mode 14:
    heartbeat lost (no timestamp, not subject to the time window) + kernel panic softlockup + watchdog timeout
    """

    # pylint: disable=R0801
    out_code = COMP_LOST_DEV_SOFTLOCKUP_CUSTOM
    evidence_codes = (
        PRECHECK_DEVICE_INFO_HEARTBEAT_LOST,
        PRECHECK_KERNEL_PANIC_SOFTLOCKUP,
        PRECHECK_KERNEL_WATCHDOG_TIMEOUT,
    )
    no_window_codes = (PRECHECK_DEVICE_INFO_HEARTBEAT_LOST,)
    window_sec = LOST_DEV_TIME_WINDOW_SEC


class TestNoAnchorCodeBranch(unittest.TestCase):
    """No-anchor branch: when anchor_code=None, _time_within_window short-circuits to True and skips the time-window check.

    Drives LostDeviceChecker subclasses directly (bypassing the fixed checker_list of single_device_analyze),
    verifying that with anchor_code=None and a non-zero time_window_sec, evidence presence alone hits,
    and arbitrary evidence time (including out of window) is not rejected.
    """

    def setUp(self):
        self.schema = make_schema()
        self.source_device = "0"
        self.merge = MergePrecheckCause(self.schema, {})

    def _make_checker(self, evidence_codes):
        class NoAnchorLostChecker(LostDeviceChecker):
            out_code = COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM
            anchor_code = None
            time_window_sec = LOST_DEV_TIME_WINDOW_SEC

        NoAnchorLostChecker.evidence_codes = evidence_codes
        return NoAnchorLostChecker(self.source_device, self.merge)

    def _run(self, checker, events):
        self.merge.single_device_precheck_event = {e["event_code"]: e for e in events}
        checker.single_device_precheck_event = self.merge.single_device_precheck_event
        result = {}
        checker.analyze(result)
        return result

    def test_full_hit_without_anchor(self):
        """No anchor: out-of-window evidence still hits (time window short-circuited); events_attribute contains only evidence."""
        checker = self._make_checker((PRECHECK_SMMU_TIMEOUT,))
        events = [lost_dev_event(PRECHECK_SMMU_TIMEOUT, occur_time=OUT_WINDOW_TIME)]
        result = self._run(checker, events)
        self.assertIn(COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM, result)
        entry = result[COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM]
        self.assertEqual([PRECHECK_SMMU_TIMEOUT], [e["event_code"] for e in entry["events_attribute"]])

    def test_no_anchor_missing_evidence_not_hit(self):
        """No anchor but evidence missing: evidence conditions still apply, no hit."""
        checker = self._make_checker((PRECHECK_SMMU_TIMEOUT,))
        result = self._run(checker, [])
        self.assertNotIn(COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM, result)

    def test_no_anchor_multi_evidence_all_required(self):
        """No anchor with multiple evidence: missing any one evidence still yields no hit."""
        checker = self._make_checker((PRECHECK_SMMU_TIMEOUT, PRECHECK_KERNEL_HARD_LOCKUP))
        events = [lost_dev_event(PRECHECK_SMMU_TIMEOUT, occur_time=OUT_WINDOW_TIME)]
        result = self._run(checker, events)
        self.assertNotIn(COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM, result)
