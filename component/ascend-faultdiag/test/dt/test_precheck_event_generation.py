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
"""Unit tests for the event-generation layer of device-loss fault modes: multi-line pairing / fdm_log group determination / sel.txt device-loss state machine.

Covers the PRECHECK event generation logic newly added in this requirement (previously untested):
1. NpuHistoryLogParser kernel.log: direct-match types (SMMU / multi-bit ECC / OOM / non-executable /
   hard lockup / core lockup / AE type23 / bbox / os_info) and delayed-pairing types
   (Kernel panic + panic reason: softlockup; watchdog stuck + rlock SOFT-WATCHDOG);
2. BMCAppDumpParser fdm_log: data-group settlement; MATA DDRC / MATA UEE / CPU_CORE /
   DMC main & mirror each produce independent events, and missing key fields produce nothing;
3. BMCAppDumpParser sel.txt: device-loss (assert without deassert) state machine, multi-bit ECC, AIC.
"""

import os
import shutil
import tempfile
import unittest

from ascend_fd.pkg.parse.knowledge_graph.parser.bmc_parser import BMCAppDumpParser
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import EventStorage
from ascend_fd.pkg.parse.knowledge_graph.parser.npu_device_parse import NpuHistoryLogParser
from ascend_fd.utils.constant.dev_log_const import (
    PRECHECK_BBOX_OS_OOM_EXCEPTION,
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
from ascend_fd.utils.constant.ub_const import PRECHECK_KERNEL_AE_TYPE23

T = "2026-06-04 03:44:24.000000"


class TestNpuHistoryLogParserPrecheckMatch(unittest.TestCase):
    """PRECHECK events produced by direct matching in kernel.log / bbox_info.txt / os_info.txt"""

    # Sample lines for common bbox/system exceptions (bbox_info.txt and os_info.txt both contain exception codes)
    BBOX_LINE = "TimeStamp [20260604034424-123456] 0xA4040001 system exception code"
    SMMU_LINE = "svm_smmu 0x1: CMD_SYNC timeout"
    ECC_LINE = "error_type: 3, multi-bit ECC at address 0x1234"
    OOM_LINE = "oom space, huge page memory is exhausted"
    EXEC_LINE = (
        "Unable to handle kernel execute of pKdma_span at 0xffff80001234  non-executable memory at virtual address"
    )
    HARD_LOCKUP_LINE = "NMI watchdog: Watchdog detected hard LOCKUP on cpu 34"
    CORE_LOCKUP_LINE = "panic reason:Core LOCKUP detected on CPU"
    AE_LINE = "UDMA 0x1 ae xxx, event type is 2, sub type is 3"
    KP_LINE = "Kernel panic - not syncing: Fatal exception"
    SOFTLOCKUP_REASON_LINE = "panic reason: softlockup"
    WATCHDOG_STUCK_LINE = "watchdog: CPU 8 self-detected hard/soft lockup"
    RLOCK_WATCHDOG_LINE = "rlock reason:SOFT-WATCHDOG"

    def setUp(self):
        self.parser = NpuHistoryLogParser({})
        self.parser.resuming_training_time = "2020-01-01 00:00:00.000000"
        self.parser.start_time = None
        self.parser.end_time = None
        self.parser.is_sdk_input = False
        self.tmp_dir = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.tmp_dir)

    def _kernel_file(self):
        return os.path.join(self.tmp_dir, "kernel.log")

    def _match(self, line, source_file=None):
        """Call _match_precheck_line (bypassing the file loop) and return the matched event or None."""
        result = self.parser._match_precheck_line(line, T, source_file or self._kernel_file())
        return result or {}

    def test_match_smmu_timeout(self):
        self.assertEqual(PRECHECK_SMMU_TIMEOUT, self._match(self.SMMU_LINE).get("event_code"))

    def test_match_kernel_multi_bit_ecc(self):
        self.assertEqual(PRECHECK_KERNEL_ERR_TYPE3_MULTI_BIT_ECC, self._match(self.ECC_LINE).get("event_code"))

    def test_match_kernel_oom_huge_page(self):
        self.assertEqual(PRECHECK_KERNEL_OOM_HUGE_PAGE_EXHAUSTED, self._match(self.OOM_LINE).get("event_code"))

    def test_match_kernel_execute_non_executable(self):
        self.assertEqual(PRECHECK_KERNEL_EXECUTE_NON_EXECUTABLE, self._match(self.EXEC_LINE).get("event_code"))

    def test_match_kernel_hard_lockup(self):
        self.assertEqual(PRECHECK_KERNEL_HARD_LOCKUP, self._match(self.HARD_LOCKUP_LINE).get("event_code"))

    def test_match_kernel_core_lockup(self):
        self.assertEqual(PRECHECK_KERNEL_CORE_LOCKUP, self._match(self.CORE_LOCKUP_LINE).get("event_code"))

    def test_match_kernel_ae_type23(self):
        self.assertEqual(PRECHECK_KERNEL_AE_TYPE23, self._match(self.AE_LINE).get("event_code"))

    def test_match_bbox_os_oom(self):
        """system exception in bbox_info.txt → BBOX OS_OOM event"""
        bbox_file = os.path.join(self.tmp_dir, "bbox_info.txt")
        self.assertEqual(PRECHECK_BBOX_OS_OOM_EXCEPTION, self._match(self.BBOX_LINE, bbox_file).get("event_code"))

    def test_match_os_info_os_oom(self):
        """system exception in os_info.txt → OS_INFO OS_OOM event"""
        os_info_file = os.path.join(self.tmp_dir, "os_info.txt")
        self.assertEqual(PRECHECK_OS_INFO_OS_OOM_EXCEPTION, self._match(self.BBOX_LINE, os_info_file).get("event_code"))

    def test_no_match_returns_empty(self):
        self.assertEqual({}, self._match("normal log line without any alarm"))


class TestNpuHistoryLogParserPairedEvents(unittest.TestCase):
    """kernel.log multi-line pairing (flushed at end of file):
    - Kernel panic (without softlockup reason) → only the PANIC event;
    - panic + "panic reason: softlockup" within 1 min → additionally emits PANIC_SOFTLOCKUP;
    - watchdog stuck + rlock SOFT-WATCHDOG within 1 min → emits WATCHDOG_TIMEOUT (two lines concatenated);
    - unpaired watchdog stuck / rlock emit no event.
    """

    KP_LINE = "Kernel panic - not syncing: Fatal exception"
    SOFTLOCKUP_REASON_LINE = "panic reason: softlockup"
    WATCHDOG_STUCK_LINE = "watchdog: CPU 8 self-detected soft lockup"
    RLOCK_WATCHDOG_LINE = "rlock reason:SOFT-WATCHDOG"

    def setUp(self):
        self.parser = NpuHistoryLogParser({})
        self.parser.resuming_training_time = "2020-01-01 00:00:00.000000"
        self.parser.start_time = None
        self.parser.end_time = None
        self.parser.is_sdk_input = False
        self.event_storage = EventStorage()
        self.source_file = os.path.join(tempfile.mkdtemp(), "kernel.log")

    def _record(self, line, occur_time):
        """Feed lines into the parser one by one (equivalent to the per-line matching of _parse_single_file)."""
        self.parser._match_precheck_line(line, occur_time, self.source_file)

    def _flush_and_codes(self):
        self.parser._flush_kernel_panic_prechecks(self.event_storage, self.source_file, "0")
        self.parser._flush_watchdog_prechecks(self.event_storage, self.source_file, "0")
        return [e["event_code"] for e in self.event_storage.generate_event_list()]

    def test_kernel_panic_alone(self):
        self._record(self.KP_LINE, T)
        self.assertEqual([PRECHECK_KERNEL_PANIC], self._flush_and_codes())

    def test_kernel_panic_with_softlockup_reason(self):
        """panic and softlockup reason within 1 min → two events: PANIC + PANIC_SOFTLOCKUP"""
        self._record(self.KP_LINE, T)
        self._record(self.SOFTLOCKUP_REASON_LINE, "2026-06-04 03:44:50.000000")
        codes = self._flush_and_codes()
        self.assertEqual({PRECHECK_KERNEL_PANIC, PRECHECK_KERNEL_PANIC_SOFTLOCKUP}, set(codes))
        soft_events = [
            e for e in self.event_storage.generate_event_list() if e["event_code"] == PRECHECK_KERNEL_PANIC_SOFTLOCKUP
        ]
        self.assertIn(self.SOFTLOCKUP_REASON_LINE, soft_events[0]["key_info"])

    def test_kernel_panic_softlockup_out_of_window(self):
        """panic and softlockup reason more than 1 min apart → only PANIC"""
        self._record(self.KP_LINE, T)
        self._record(self.SOFTLOCKUP_REASON_LINE, "2026-06-04 03:45:30.000000")
        self.assertEqual([PRECHECK_KERNEL_PANIC], self._flush_and_codes())

    def test_watchdog_with_rlock(self):
        """watchdog stuck + rlock SOFT-WATCHDOG within 1 min → WATCHDOG_TIMEOUT (key_info contains both lines)"""
        self._record(self.WATCHDOG_STUCK_LINE, T)
        self._record(self.RLOCK_WATCHDOG_LINE, "2026-06-04 03:44:55.000000")
        codes = self._flush_and_codes()
        self.assertEqual([PRECHECK_KERNEL_WATCHDOG_TIMEOUT], codes)
        events = self.event_storage.generate_event_list()
        self.assertIn(self.WATCHDOG_STUCK_LINE, events[0]["key_info"])
        self.assertIn(self.RLOCK_WATCHDOG_LINE, events[0]["key_info"])

    def test_watchdog_without_rlock_not_emitted(self):
        """watchdog stuck without a paired rlock → no WATCHDOG_TIMEOUT emitted"""
        self._record(self.WATCHDOG_STUCK_LINE, T)
        self.assertEqual([], self._flush_and_codes())

    def test_rlock_alone_not_emitted(self):
        """rlock SOFT-WATCHDOG line alone, without watchdog stuck → no event emitted"""
        self._record(self.RLOCK_WATCHDOG_LINE, T)
        self.assertEqual([], self._flush_and_codes())


class TestBmcAppDumpFdmLog(unittest.TestCase):
    """fdm_log data-group settlement: MATA / CPU_CORE / DMC events are each emitted independently"""

    def setUp(self):
        self.params = {
            "default_conf": {},
            "user_conf": {},
            "start_time": "2026-01-01 00:00:00.000000",
            "end_time": "2026-12-31 23:59:59.000000",
        }
        self.parser = BMCAppDumpParser(self.params)
        self.parser.is_sdk_input = False
        self.tmp_dir = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.tmp_dir)
        self.fdm_path = None

    def _write_fdm(self, lines: list):
        self.fdm_path = os.path.join(self.tmp_dir, "fdm_log")
        with open(self.fdm_path, "w", encoding="utf-8") as f:
            f.write("\n".join(lines))
        return self.fdm_path

    def _parse(self, lines):
        path = self._write_fdm(lines)
        return self.parser._parse_file(path)

    def _mata_group(self, lines):
        """MATA group sample: contains the Time/Socket/HAM/Module/MATA keywords."""
        group = [
            "[Error Log Number]",
            "Time: 2026-06-04 03:44:24",
            "Socket: NPU2",
            "HAM: 2",
            "Module: MATA",
        ]
        return group + lines

    def test_mata_ddrc_and_uee(self):
        """MATA group hitting both IERRCODE DDRC and Error Type UER → two independent events"""
        lines = self._mata_group(
            [
                "IERRCODE:0x1 (Response Error From DDRC)",
                "Error Type:Uncorrected Error (UER)",
            ]
        )
        events = self._parse(lines)
        codes = sorted(e["event_code"] for e in events)
        self.assertEqual([PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC, PRECHECK_FDM_MATA_UNCORRECTED_ERROR], codes)
        # Device number conversion: HAM 2 × Socket NPU2(=2) - 1 = 3
        self.assertEqual("3", events[0]["source_device"])

    def test_mata_no_error_matches_nothing(self):
        """MATA group without DDRC/UER/UEO/UEU keywords → no events"""
        lines = self._mata_group(["IERRCODE:0x0 (No Error)"])
        self.assertEqual([], self._parse(lines))

    def test_cpu_core_ce_error(self):
        """CPU_CORE with all three sentences hit (CE + associative-memory SERR + IERR Undefined) → AICPU HBM event"""
        lines = [
            "[Error Log Number]",
            "Time: 2026-06-04 03:44:24",
            "Socket: NPU2",
            "HAM: 2",
            "Module: CPU_CORE",
            "Error Type:At least one error was corrected.(CE)",
            "SERRCODE:0x2 (Data error from associative memory)",
            "IERRCODE:0x3 (Undefined)",
        ]
        events = self._parse(lines)
        codes = [e["event_code"] for e in events]
        self.assertEqual([PRECHECK_FDM_AICPU_HBM_ERROR], codes)

    def test_cpu_core_missing_one_sentence_not_emitted(self):
        """CPU_CORE missing the IERRCODE(Undefined) sentence → no event emitted"""
        lines = [
            "[Error Log Number]",
            "Time: 2026-06-04 03:44:24",
            "Socket: NPU2",
            "HAM: 2",
            "Module: CPU_CORE",
            "Error Type:At least one error was corrected.(CE)",
            "SERRCODE:0x2 (Data error from associative memory)",
        ]
        self.assertEqual([], self._parse(lines))

    def test_dmc_main_and_mirror(self):
        """fdm DMC main and mirror each hit → two independent events"""
        lines = self._mata_group(["dmc_rdata_main_error=1", "dmc_rdata_mirror_error=1"])
        events = self._parse(lines)
        codes = sorted(e["event_code"] for e in events)
        self.assertEqual([PRECHECK_FDM_DMC_MAIN_ERROR, PRECHECK_FDM_DMC_MIRROR_ERROR], codes)

    def test_group_missing_required_fields_not_emitted(self):
        """Group missing any of Socket/HAM/Time/Module → the whole group emits no events"""
        lines = [
            "[Error Log Number]",
            "Time: 2026-06-04 03:44:24",
            "Module: MATA",
            "IERRCODE:0x1 (Response Error From DDRC)",
        ]
        self.assertEqual([], self._parse(lines))

    def test_missing_required_field_each(self):
        """Batch-verify that all four scenarios of missing Time / Socket / HAM / Module emit nothing."""
        base = [
            "Time: 2026-06-04 03:44:24",
            "Socket: NPU2",
            "HAM: 2",
            "Module: MATA",
            "IERRCODE:0x1 (Response Error From DDRC)",
        ]
        for drop_field in ("Time:", "Socket:", "HAM:", "Module:"):
            with self.subTest(drop=drop_field):
                lines = ["[Error Log Number]"] + [ln for ln in base if not ln.startswith(drop_field)]
                self.assertEqual([], self._parse(lines))

    def test_two_groups_each_flushed(self):
        """Two groups in one file: after the first group settles and flushes, the second group hits independently; both batches are emitted."""
        group1 = self._mata_group(["IERRCODE:0x1 (Response Error From DDRC)"])
        group2 = self._mata_group(["Error Type:Uncorrected Error (UEU)"])
        events = self._parse(group1 + group2)
        codes = sorted(e["event_code"] for e in events)
        self.assertEqual([PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC, PRECHECK_FDM_MATA_UNCORRECTED_ERROR], codes)

    def test_duplicate_dmc_keyword_dedup(self):
        """The same dmc keyword repeated within one group produces only one event."""
        lines = self._mata_group(["dmc_rdata_main_error=1", "dmc_rdata_main_error=2"])
        events = self._parse(lines)
        codes = [e["event_code"] for e in events]
        self.assertEqual([PRECHECK_FDM_DMC_MAIN_ERROR], codes)
        self.assertEqual(1, len([e for e in events if e["event_code"] == PRECHECK_FDM_DMC_MAIN_ERROR]))


class TestBmcAppDumpSelTxt(unittest.TestCase):
    """sel.txt device-loss / multi-bit ECC / AIC event generation"""

    SEL_LOST_LINE = "1 | 2026-06-04 03:44:24 | hello | 0x56000001 | Asserted | NPU1 device NPU connection has been lost"
    SEL_LOST_REVERTED_LINE = (
        "2 | 2026-06-04 03:50:00 | hello | 0x56000001 | Deasserted | NPU1 device NPU connection has been lost"
    )
    SEL_MULTIBIT_ECC_LINE = (
        "3 | 2026-06-04 03:44:24 | hello | 0x56000003 | Asserted | 0x80e01801 NPU2 device health critical multi-bit ECC"
    )
    SEL_AIC_ECC_LINE = "4 | 2026-06-04 03:44:24 | hello | 0x56000001 | Asserted | NPU3 0x80c98008 AIC multi-bit ECC"

    def setUp(self):
        self.params = {
            "default_conf": {},
            "user_conf": {},
            "start_time": "2026-01-01 00:00:00.000000",
            "end_time": "2026-12-31 23:59:59.000000",
        }
        self.parser = BMCAppDumpParser(self.params)
        self.parser.is_sdk_input = False
        self.tmp_dir = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.tmp_dir)

    def _write_sel(self, lines):
        path = os.path.join(self.tmp_dir, "sel.txt")
        with open(path, "w", encoding="utf-8") as f:
            f.write("\n".join(lines))
        return self.parser._parse_file(path)

    def test_sel_connection_lost_asserted(self):
        """Device-loss Asserted line (without a Deasserted recovery) → device-loss event, NPU1 converts to device 0"""
        events = self._write_sel([self.SEL_LOST_LINE])
        self.assertEqual(1, len(events))
        self.assertEqual(PRECHECK_SEL_LOG_LOST_NPU, events[0]["event_code"])
        self.assertEqual("0", events[0]["source_device"])

    def test_sel_connection_lost_recovered_not_emitted(self):
        """The latest device-loss line (first in reverse order) is a Deasserted recovery boundary → no device-loss event"""
        events = self._write_sel([self.SEL_LOST_REVERTED_LINE, self.SEL_LOST_LINE])
        self.assertEqual([], events)

    def test_sel_multibit_ecc(self):
        """Event Code 0x56000003 + description 0x80e01801 → health multi-bit ECC"""
        events = self._write_sel([self.SEL_MULTIBIT_ECC_LINE])
        self.assertEqual(1, len(events))
        self.assertEqual(PRECHECK_SEL_MULTIBIT_ECC, events[0]["event_code"])
        self.assertEqual("1", events[0]["source_device"])

    def test_sel_aic_multibit_ecc(self):
        """description 0x80c98008 → AIC multi-bit ECC"""
        events = self._write_sel([self.SEL_AIC_ECC_LINE])
        self.assertEqual(1, len(events))
        self.assertEqual(PRECHECK_SEL_AIC_MULTIBIT_ECC, events[0]["event_code"])
        self.assertEqual("2", events[0]["source_device"])

    def test_sel_connection_lost_multi_assert_merged(self):
        """Multiple Asserted lines for the same NPU (during a fault) are all recorded, and occurrence merges the time groups."""
        lines = [
            "4 | 2026-06-04 03:44:24 | hello | 0x56000001 | Asserted | NPU1 device NPU connection has been lost",
            "3 | 2026-06-04 03:45:24 | hello | 0x56000001 | Asserted | NPU1 device NPU connection has been lost",
        ]
        events = self._write_sel(lines)
        self.assertEqual(1, len(events))
        self.assertEqual(PRECHECK_SEL_LOG_LOST_NPU, events[0]["event_code"])
        # occurrence collects both lines
        self.assertEqual(2, len(events[0]["occurrence"]))
        event_times = {e[0] for e in events[0]["occurrence"]}
        self.assertEqual({"2026-06-04 03:44:24", "2026-06-04 03:45:24"}, event_times)

    def test_sel_multibit_ecc_multi_assert_merged(self):
        """Multiple Asserted lines for multi-bit ECC → single event with merged occurrence."""
        lines = [
            "5 | 2026-06-04 03:44:24 | hello | 0x56000003 | Asserted | "
            "0x80e01801 NPU2 device health critical multi-bit ECC",
            "6 | 2026-06-04 03:46:01 | hello | 0x56000003 | Asserted | "
            "0x80e01801 NPU2 device health critical multi-bit ECC",
        ]
        events = self._write_sel(lines)
        self.assertEqual(1, len(events))
        self.assertEqual(PRECHECK_SEL_MULTIBIT_ECC, events[0]["event_code"])
        self.assertEqual(2, len(events[0]["occurrence"]))


if __name__ == "__main__":
    unittest.main()
