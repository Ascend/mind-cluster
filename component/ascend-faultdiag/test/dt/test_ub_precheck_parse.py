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
from datetime import datetime
import os
import tempfile
import unittest
import zipfile
from unittest.mock import patch

from ascend_fd.pkg.parse.knowledge_graph.parser.cann_log_parser import PrecheckEnvParser
from ascend_fd.pkg.parse.knowledge_graph.parser.lcne_parser import LCNEParser
from ascend_fd.pkg.parse.knowledge_graph.parser.npu_device_parse import NpuHistoryLogParser, UbctlLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.npu_info_parser import HcclStatInfoParser
from ascend_fd.utils.constant.str_const import UNKNOWN_DEVICE_ID
from ascend_fd.utils.constant.ub_const import (
    AGE_PERIOD_KEYWORD,
    HCOMM_TA_CTP_UB_TIMEOUT,
    ICRC_ERR_COUNT_METRICS,
    NIC_TX_L2_ERR_DROP_PKT_CNT,
    PHY_REINIT_CNT,
    PPP_RLT_DROP_NUM,
    PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT,
    PRECHECK_KERNEL_AE_TYPE23,
    PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA,
    PRECHECK_RXDMA_ICRC_DATA,
    PRECHECK_UBCTL_DATA,
    PRECHECK_UBMEM_TIMEOUT,
    RC_FULL_QUEUE_CNT,
    SIDE_AFTER,
    SIDE_BEFORE,
    SNAPSHOOT_TAG,
    TP_RRP_ERR_FLG_0,
    TX_VL6_PKT_NUM,
)

EMPTY_PARAMS = {"default_conf": {}, "user_conf": {}}
SNAP_PREFIX = "[INFO]" + SNAPSHOOT_TAG
SNAP_LOG = "/tmp/lcne_log/diagnostic_information/nsdk_hccl_shapshot.log"


class TestNpuHistoryPrecheckLine(unittest.TestCase):
    def setUp(self):
        self.parser = NpuHistoryLogParser(EMPTY_PARAMS)

    def test_match_udma_ae_line(self):
        line = "2026-01-01 12:00:00 UDMA 0x100 ae 0x200, event type is 2, sub type is 3"
        ret = self.parser._match_precheck_line(line)
        self.assertEqual(ret["event_code"], PRECHECK_KERNEL_AE_TYPE23)
        self.assertTrue(ret["key_info"])
        self.assertEqual(ret["attribute"], {})
        self.assertFalse(ret["is_custom_event"])

    def test_missing_one_keyword_not_match(self):
        line = "2026-01-01 12:00:00 UDMA 0x100 err 0x200, event type is 2, sub type is 3"
        self.assertEqual(self.parser._match_precheck_line(line), {})
        # 空行同样命中负匹配分支
        self.assertEqual(self.parser._match_precheck_line(""), {})


class TestLcnePrecheckLine(unittest.TestCase):
    def setUp(self):
        self.parser = LCNEParser(EMPTY_PARAMS)

    def test_match_age_period_line(self):
        line = "May 27 2025 11:25:00 set ubmem timeout 0x1f400000 age_period = 40000, done"
        ret = self.parser._match_precheck_line(line)
        self.assertEqual(ret["event_code"], PRECHECK_UBMEM_TIMEOUT)
        self.assertTrue(ret["key_info"])
        self.assertEqual(ret["attribute"], {AGE_PERIOD_KEYWORD: "40000"})
        self.assertFalse(ret["is_custom_event"])

    def test_missing_keyword_not_match(self):
        self.assertEqual(self.parser._match_precheck_line("set ubmem timeout 0x1f400000 done"), {})
        # 空行同样命中负匹配分支
        self.assertEqual(self.parser._match_precheck_line(""), {})

    def test_age_period_without_digits_not_match(self):
        line = "set ubmem timeout 0x1f400000 age_period = unknown done"
        self.assertEqual(self.parser._match_precheck_line(line), {})


class TestPrecheckEnvParser(unittest.TestCase):
    def setUp(self):
        self.parser = PrecheckEnvParser()

    def test_parse_hcomm_env_line(self):
        line = "[HCCL_ENV] HCOMM_TA_CTP_UB_TIMEOUT set by env to [8]"
        self.parser.parse_line(line, "plog-0.log")
        self.assertIn(HCOMM_TA_CTP_UB_TIMEOUT, self.parser.env_info_dict)
        self.assertEqual(self.parser.env_info_dict[HCOMM_TA_CTP_UB_TIMEOUT]["env_value"], "8")
        self.assertEqual(self.parser.source_file, "plog-0.log")

    def test_get_precheck_event(self):
        line = "[HCCL_ENV] HCOMM_TA_CTP_UB_TIMEOUT set by env to [4]"
        self.parser.parse_line(line, "plog-0.log")
        events = self.parser.get_precheck_event()
        self.assertEqual(len(events), 1)
        event = events[0]
        self.assertEqual(event["event_code"], PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT)
        self.assertEqual(event["attribute"], {HCOMM_TA_CTP_UB_TIMEOUT: "4"})
        self.assertEqual(event["source_device"], UNKNOWN_DEVICE_ID)
        self.assertTrue(event["key_info"])

    def test_unrelated_line_ignored(self):
        # 无相关行：env 信息保持为空，且不产生 PRECHECK 事件
        self.parser.parse_line("[HCCL_ENV] HCCL_CONNECT_TIMEOUT set by env to [120]", "plog-0.log")
        self.assertEqual(self.parser.env_info_dict, {})
        self.assertEqual(self.parser.get_precheck_event(), [])

    def test_keyword_without_pattern_ignored(self):
        self.parser.parse_line("HCOMM_TA_CTP_UB_TIMEOUT something wrong", "plog-0.log")
        self.assertEqual(self.parser.get_precheck_event(), [])


class TestUbctlFeedLine(unittest.TestCase):
    def setUp(self):
        self.parser = UbctlLogParser(EMPTY_PARAMS)
        self.parser.calc_pkg = UbctlLogParser._CalcPkg()

    def feed(self, *lines):
        for line in lines:
            self.parser.feed_line(line)

    def test_blank_line_ignored(self):
        self.parser.feed_line("")
        self.parser.feed_line("   ")
        self.assertIsNone(self.parser.calc_pkg.current_port)
        self.assertFalse(self.parser.calc_pkg.parse_finished)

    def test_command_block_metric_collected_by_udie_port(self):
        # U1-1：Command 行以 -d/-p 精确定位 (udie, port)；指标行格式为「指标名: 0x值」（冒号后带空格）
        self.feed('"Command: ubctl -c 0 -d 0 -m dump -p 2"', "rc_full_queue_cnt: 0x0")
        self.assertEqual(self.parser.calc_pkg.udie0[2][RC_FULL_QUEUE_CNT], "0x0")

    def test_block_switch_follows_latest_command(self):
        # U1-2：后一个 Command 块接管所属 (udie, port)，各 port 数据独立累积
        self.feed('"Command: ubctl -c 0 -d 0 -m dump -p 2"', "rc_full_queue_cnt: 0x1")
        self.feed('"Command: ubctl -c 0 -d 0 -m dump -p 3"', "rc_full_queue_cnt: 0x2")
        self.assertEqual(self.parser.calc_pkg.udie0[2][RC_FULL_QUEUE_CNT], "0x1")
        self.assertEqual(self.parser.calc_pkg.udie0[3][RC_FULL_QUEUE_CNT], "0x2")

    def test_udie1_determined_by_d1(self):
        # U1-3：-d 1 精确定位 udie1，不再靠 port 序列推断
        self.feed('"Command: ubctl -c 0 -d 1 -m dump -p 1"', "tp_rrp_err_flg_0: 0x2")
        self.assertEqual(self.parser.calc_pkg.udie1[1][TP_RRP_ERR_FLG_0], "0x2")
        self.assertNotIn(TP_RRP_ERR_FLG_0, self.parser.calc_pkg.udie0[1])

    def test_command_without_d_p_ignored(self):
        # U1-4：缺 -d/-p 的 Command 行：该块忽略，不影响后续块
        self.feed('"Command: ubctl -c 0 -m dump"', "rc_full_queue_cnt: 0x1")
        for port in range(9):
            self.assertEqual(self.parser.calc_pkg.udie0[port], {})
            self.assertEqual(self.parser.calc_pkg.udie1[port], {})
        self.feed('"Command: ubctl -c 0 -d 0 -m dump -p 4"', "rc_full_queue_cnt: 0x3")
        self.assertEqual(self.parser.calc_pkg.udie0[4][RC_FULL_QUEUE_CNT], "0x3")

    def test_metric_without_command_context_ignored(self):
        # U1-5：无 Command 上下文（未进入合法数据块）的指标行忽略
        self.parser.feed_line("rc_full_queue_cnt: 0x1")
        for port in range(9):
            self.assertEqual(self.parser.calc_pkg.udie0[port], {})
            self.assertEqual(self.parser.calc_pkg.udie1[port], {})

    def test_out_of_range_port_block_discarded(self):
        # U1-6：-p 9 越界（>8）：该块整体丢弃，当前归属复位
        self.feed('"Command: ubctl -c 0 -d 0 -m dump -p 9"', "rc_full_queue_cnt: 0x1")
        self.assertIsNone(self.parser.calc_pkg.current_port)
        for port in range(9):
            self.assertEqual(self.parser.calc_pkg.udie0[port], {})
            self.assertEqual(self.parser.calc_pkg.udie1[port], {})

    def test_same_metric_overwrites(self):
        # U1-7：同 (udie, port) 同指标重复出现取最后一次
        self.feed(
            '"Command: ubctl -c 0 -d 0 -m dump -p 0"',
            "rc_full_queue_cnt: 0x1",
            "rc_full_queue_cnt: 0x2",
        )
        self.assertEqual(self.parser.calc_pkg.udie0[0][RC_FULL_QUEUE_CNT], "0x2")

    def test_metric_value_with_space_stripped(self):
        # U1-8：冒号后带空格（真实日志格式），strip 后原样保存
        self.feed('"Command: ubctl -c 0 -d 0 -m dump -p 0"', "rc_full_queue_cnt: 0x0")
        self.assertEqual(self.parser.calc_pkg.udie0[0][RC_FULL_QUEUE_CNT], "0x0")

    def test_unknown_or_legacy_lines_ignored(self):
        # U1-9：非指标关键字行、旧 port_id 行均忽略
        self.feed('"Command: ubctl -c 0 -d 0 -m dump -p 0"', "foo: 0x1", "port_id: 0x2")
        self.assertEqual(self.parser.calc_pkg.udie0[0], {})

    def test_non_dump_mode_stops_parsing(self):
        # U1-10：-m 非 dump ⇒ parse_finished 置位，后续行不再采集
        self.feed('"Command: ubctl -c 0 -d 0 -m dump -p 0"', "rc_full_queue_cnt: 0x1")
        self.assertFalse(self.parser.calc_pkg.parse_finished)
        self.feed('"Command: ubctl -c 0 -d 0 -m collect -p 0"', "rc_full_queue_cnt: 0x2")
        self.assertTrue(self.parser.calc_pkg.parse_finished)
        self.assertEqual(self.parser.calc_pkg.udie0[0][RC_FULL_QUEUE_CNT], "0x1")

    def test_malformed_command_line_ignored(self):
        # 乱序参数 / udie 越界等整行不匹配固定格式：该块忽略
        self.feed("Command: ubctl -c 0 -p 0 -m dump", "rc_full_queue_cnt: 0x1")
        self.feed("Command: ubctl -d 2 -m dump -p 0", "rc_full_queue_cnt: 0x2")
        for port in range(9):
            self.assertEqual(self.parser.calc_pkg.udie0[port], {})
            self.assertEqual(self.parser.calc_pkg.udie1[port], {})


class TestUbctlTimestampAndSelect(unittest.TestCase):
    def setUp(self):
        self.parser = UbctlLogParser(EMPTY_PARAMS)

    def test_parse_timestamp_dir_valid(self):
        dt = self.parser._parse_timestamp_dir_to_datetime("2026_08_23_13_20_00")
        self.assertEqual(dt, datetime(2026, 8, 23, 13, 20, 0))

    def test_parse_timestamp_dir_invalid(self):
        self.assertIsNone(self.parser._parse_timestamp_dir_to_datetime("2026_08_23"))
        self.assertIsNone(self.parser._parse_timestamp_dir_to_datetime("abc"))
        self.assertIsNone(self.parser._parse_timestamp_dir_to_datetime("2026_13_23_13_20_00"))

    def _pairs(self):
        return [
            (datetime(2026, 8, 23, 13, 10, 0), "t1", "p1", "0"),
            (datetime(2026, 8, 23, 13, 20, 0), "t2", "p2", "0"),
            (datetime(2026, 8, 23, 13, 30, 0), "t3", "p3", "0"),
            (datetime(2026, 8, 23, 13, 40, 0), "t4", "p4", "0"),
            (datetime(2026, 8, 23, 13, 30, 0), "t5", "p5", "1"),
        ]

    def test_before_picks_latest_before_start(self):
        self.parser.start_time = "2026-08-23 13:25:00"
        self.parser.end_time = "2026-08-23 13:35:00"
        self.parser.resuming_training_time = ""
        ret = self.parser._select_by_train_time(self._pairs(), SIDE_BEFORE)
        self.assertEqual(ret["0"], ("p2", "2026-08-23 13:20:00"))

    def test_after_picks_earliest_after_end(self):
        self.parser.start_time = "2026-08-23 13:15:00"
        self.parser.end_time = "2026-08-23 13:30:00"
        self.parser.resuming_training_time = ""
        ret = self.parser._select_by_train_time(self._pairs(), SIDE_AFTER)
        # dev 0 after 选择满足 end 的最早文件
        self.assertEqual(ret["0"], ("p3", "2026-08-23 13:30:00"))
        self.assertEqual(ret["1"], ("p5", "2026-08-23 13:30:00"))

    def test_resuming_time_filters_before(self):
        self.parser.start_time = "2026-08-23 13:20:00"
        self.parser.end_time = "2026-08-23 13:45:00"
        self.parser.resuming_training_time = "2026-08-23 13:20:00"
        ret = self.parser._select_by_train_time(self._pairs(), SIDE_BEFORE)
        # 断点后（>=resuming）的 before 数据应被过滤，取 <=resuming 的最近一条
        self.assertEqual(ret["0"], ("p2", "2026-08-23 13:20:00"))

    def test_no_valid_pair_returns_empty(self):
        self.parser.start_time = "2026-08-23 13:00:00"
        self.parser.end_time = "2026-08-23 13:05:00"
        self.parser.resuming_training_time = ""
        ret = self.parser._select_by_train_time(self._pairs(), SIDE_BEFORE)
        self.assertEqual(ret, {})


class TestUbctlAssembleEvents(unittest.TestCase):
    def setUp(self):
        self.parser = UbctlLogParser(EMPTY_PARAMS)

    def _collect_result(self, udie0, udie1, path, occur_time):
        return {"udie0": udie0, "udie1": udie1, "source_path": path, "occur_time": occur_time}

    def test_assemble_before_after_events(self):
        before = {"0": self._collect_result({0: {RC_FULL_QUEUE_CNT: "0x0"}}, {}, "before.txt", "2026-08-23 13:20:00")}
        after = {"0": self._collect_result({0: {RC_FULL_QUEUE_CNT: "0x1"}}, {}, "after.txt", "2026-08-23 13:40:00")}
        events = self.parser._assemble_precheck_events(None, before, after)
        self.assertEqual(len(events), 1)
        event = events[0]
        self.assertEqual(event["event_code"], PRECHECK_UBCTL_DATA)
        self.assertEqual(event["source_device"], "0")
        self.assertEqual(event["occur_time"], "2026-08-23 13:40:00")
        self.assertEqual(event["attribute"]["before"]["udie0"][0][RC_FULL_QUEUE_CNT], "0x0")
        self.assertEqual(event["attribute"]["after"]["udie0"][0][RC_FULL_QUEUE_CNT], "0x1")
        self.assertTrue(event["is_custom_event"] is False)

    def test_assemble_merges_devices_in_both_side(self):
        before = {
            "0": self._collect_result({}, {}, "b0.txt", "2026-08-23 13:20:00"),
            "1": self._collect_result({}, {}, "b1.txt", "2026-08-23 13:20:00"),
        }
        after = {
            "0": self._collect_result({}, {}, "a0.txt", "2026-08-23 13:40:00"),
            "2": self._collect_result({}, {}, "a2.txt", "2026-08-23 13:40:00"),
        }
        events = self.parser._assemble_precheck_events(None, before, after)
        self.assertEqual({e["source_device"] for e in events}, {"0", "1", "2"})
        dev0 = next(e for e in events if e["source_device"] == "0")
        self.assertEqual(dev0["occur_time"], "2026-08-23 13:40:00")


class TestUbctlKernelMerge(unittest.TestCase):
    """以 ubctl_log.txt 为底稿、kernel.log 覆盖/补齐的指标级双源合并"""

    def setUp(self):
        self.parser = UbctlLogParser(EMPTY_PARAMS)

    @staticmethod
    def _dev(udie0, udie1, path, occur_time=""):
        return {"udie0": udie0, "udie1": udie1, "source_path": path, "occur_time": occur_time}

    def test_ubctl_base_with_kernel_override_and_fill(self):
        ubctl = {
            "0": self._dev({0: {TX_VL6_PKT_NUM: "0x9", RC_FULL_QUEUE_CNT: "0x3"}}, {}, "ubctl_log.txt"),
        }
        kernel = {
            "0": self._dev(
                {0: {TX_VL6_PKT_NUM: "0x5", PHY_REINIT_CNT: "0x1"}, 1: {RC_FULL_QUEUE_CNT: "0x4"}},
                {0: {TP_RRP_ERR_FLG_0: "0x10000000"}},
                "kernel.log",
            )
        }
        self.parser._merge_kernel_result(ubctl, kernel)
        udie0_port0 = ubctl["0"]["udie0"][0]
        # 两源都有的指标用 kernel 覆盖 ubctl
        self.assertEqual(udie0_port0[TX_VL6_PKT_NUM], "0x5")
        # ubctl 独有的指标保持不动
        self.assertEqual(udie0_port0[RC_FULL_QUEUE_CNT], "0x3")
        # kernel 独有的指标补入
        self.assertEqual(udie0_port0[PHY_REINIT_CNT], "0x1")
        # ubctl 未解析到的 port / udie 由 kernel 补齐
        self.assertEqual(ubctl["0"]["udie0"][1][RC_FULL_QUEUE_CNT], "0x4")
        self.assertEqual(ubctl["0"]["udie1"][0][TP_RRP_ERR_FLG_0], "0x10000000")
        # 事件溯源包含 ubctl 底稿与 kernel 数据源
        self.assertEqual(ubctl["0"]["source_path"].splitlines(), ["ubctl_log.txt", "kernel.log"])

    def test_device_not_covered_by_ubctl_uses_kernel(self):
        ubctl = {"0": self._dev({0: {TX_VL6_PKT_NUM: "0x5"}}, {}, "ubctl_log.txt")}
        kernel = {"1": self._dev({0: {RC_FULL_QUEUE_CNT: "0x2"}}, {}, "kernel.log")}
        self.parser._merge_kernel_result(ubctl, kernel)
        # ubctl 未覆盖的设备整体采用 kernel 数据
        self.assertIn("1", ubctl)
        self.assertEqual(ubctl["1"]["source_path"], "kernel.log")
        self.assertEqual(ubctl["1"]["udie0"][0][RC_FULL_QUEUE_CNT], "0x2")
        # 已有 ubctl 数据的设备不被 kernel 覆盖
        self.assertEqual(ubctl["0"]["udie0"][0][TX_VL6_PKT_NUM], "0x5")

    def test_parse_emits_single_event_with_merged_metrics(self):
        kernel_before = {"0": self._dev({0: {TX_VL6_PKT_NUM: "0x5", PHY_REINIT_CNT: "0x1"}}, {}, "kernel.log")}
        kernel_after = {"0": self._dev({0: {TX_VL6_PKT_NUM: "0x7", PHY_REINIT_CNT: "0x2"}}, {}, "kernel.log")}
        ubctl_before = {"0": self._dev({0: {TX_VL6_PKT_NUM: "0x9", RC_FULL_QUEUE_CNT: "0x1"}}, {}, "ubctl_before.txt")}
        ubctl_after = {"0": self._dev({0: {TX_VL6_PKT_NUM: "0xB", RC_FULL_QUEUE_CNT: "0x2"}}, {}, "ubctl_after.txt")}
        ctx = KGParseCtxStub()
        ctx.parse_file_path = type(
            "PathStub", (), {"ubctl_before_files": ["before.txt"], "ubctl_after_files": ["after.txt"]}
        )()
        with (
            patch.object(self.parser, "_collect_kernel_dfx", return_value=(kernel_before, kernel_after)),
            patch.object(
                self.parser,
                "_pick_ubctl_file",
                side_effect=[
                    {"0": ("ubctl_before.txt", "2026-08-23 13:20:00")},
                    {"0": ("ubctl_after.txt", "2026-08-23 13:40:00")},
                ],
            ),
            patch.object(self.parser, "_parse_ubctl_content", side_effect=[ubctl_before, ubctl_after]),
        ):
            events, err = self.parser.parse(ctx, "task_id")
        self.assertEqual(err, {})
        self.assertEqual(len(events), 1)
        event = events[0]
        self.assertEqual(event["event_code"], PRECHECK_UBCTL_DATA)
        # 两源共有的指标用 kernel 覆盖 ubctl
        self.assertEqual(event["attribute"]["after"]["udie0"][0][TX_VL6_PKT_NUM], "0x7")
        # ubctl 独有的指标保持不动
        self.assertEqual(event["attribute"]["after"]["udie0"][0][RC_FULL_QUEUE_CNT], "0x2")
        # kernel 独有的指标补入
        self.assertEqual(event["attribute"]["after"]["udie0"][0][PHY_REINIT_CNT], "0x2")
        self.assertEqual(event["attribute"]["before"]["udie0"][0][PHY_REINIT_CNT], "0x1")
        # 事件溯源包含 ubctl 底稿和 kernel 数据源
        self.assertEqual(set(event["source_file"].split("\n")), {"kernel.log", "ubctl_before.txt", "ubctl_after.txt"})

    def test_kernel_filled_metric_triggers_precheck_rule(self):
        # 8.1-6②：ubctl 缺失的指标由 kernel 补齐后，规则层应能正常触发
        from ascend_fd.pkg.parse.knowledge_graph.prechecker.merge_precheck_cause import MergePrecheckCause
        from ascend_fd.utils.constant.ub_const import RULE_RC_FULL_QUEUE_NONZERO

        ubctl = {"0": self._dev({0: {TX_VL6_PKT_NUM: "0x5"}}, {}, "ubctl_log.txt")}
        kernel = {"0": self._dev({0: {RC_FULL_QUEUE_CNT: "0x1"}}, {}, "kernel.log")}
        self.parser._merge_kernel_result(ubctl, kernel)
        events = self.parser._assemble_precheck_events(None, ubctl, ubctl)
        self.assertEqual(len(events), 1)
        checker = MergePrecheckCause(schema=None, precheck_info={})
        checker.single_device_precheck_event = {PRECHECK_UBCTL_DATA: events[0]}
        checker._prepare_ubctl_log_rule()
        self.assertTrue(checker.rule_flags[RULE_RC_FULL_QUEUE_NONZERO]["value"])

    def test_both_sources_missing_returns_no_event(self):
        # 8.1-6④：kernel.log 无 aubdfx 数据且无 ubctl 文件时，不产出事件、不报错
        ctx = type("Ctx", (), {})()
        ctx.parse_file_path = type("PathStub", (), {"ubctl_before_files": [], "ubctl_after_files": []})()
        ctx.is_sdk_input = False
        ctx.resuming_training_time = ""
        with (
            patch.object(self.parser, "_collect_kernel_dfx", return_value=({}, {})),
            patch.object(self.parser, "find_log", return_value=[]),
        ):
            events, err = self.parser.parse(ctx, "task_id")
        self.assertEqual(events, [])
        self.assertEqual(err, {})


class TestHcclStatInfoParser(unittest.TestCase):
    def setUp(self):
        self.parser = HcclStatInfoParser(end_time="2026-08-23 13:40:00")

    def _sample_message(self, q0, q1):
        return (
            "hccn_tool -g -stat -i 0 -u 1 -p 4\n"
            "rxdma_icrc_err_cnt_queue_id0: %d\n"
            "rxdma_icrc_err_cnt_queue_id1: %d\n" % (q0, q1)
        )

    def test_extract_metrics(self):
        msg = self._sample_message(5, 6)
        self.assertEqual(
            self.parser._extract_metrics(msg), {"rxdma_icrc_err_cnt_queue_id0": 5, "rxdma_icrc_err_cnt_queue_id1": 6}
        )

    def test_parse_and_processing_info(self):
        before_msg = self._sample_message(5, 6)
        after_msg = self._sample_message(7, 6)
        events = self.parser.parse({"before": [before_msg], "after": [after_msg]})
        self.assertEqual(len(events), 1)
        event = events[0]
        self.assertEqual(event["event_code"], PRECHECK_RXDMA_ICRC_DATA)
        self.assertEqual(event["source_device"], "0")
        self.assertEqual(event["source_file"], "npu_info_before/after.txt")
        self.assertEqual(event["attribute"]["before"]["1"]["4"][ICRC_ERR_COUNT_METRICS[0]], 5)
        self.assertEqual(event["attribute"]["after"]["1"]["4"][ICRC_ERR_COUNT_METRICS[0]], 7)
        self.assertEqual(event["attribute"]["after"]["1"]["4"][ICRC_ERR_COUNT_METRICS[1]], 6)

    def test_invalid_message_no_events(self):
        # 非 A5 命令行，或 A5 命令但无有效指标：均不产生 PRECHECK 事件
        self.assertEqual(self.parser.parse({"before": ["some other npu info line"]}), [])
        msg = "hccn_tool -g -stat -i 0 -u 1 -p 4\nno metric here"
        self.assertEqual(self.parser.parse({"before": [msg]}), [])


class TestKernelDfxParse(unittest.TestCase):
    """kernel.log aubdfx 网络 DFX 解析（UB 维测主数据源）"""

    def setUp(self):
        self.parser = UbctlLogParser(EMPTY_PARAMS)

    @staticmethod
    def _kernel_group(side, group_time, value, source_path="kernel.log"):
        """直接构造 kernel 数据块（组时间已换算），避免依赖真实 device_os.log"""
        return {
            "side": side,
            "udie": "udie0",
            "ports": {0: {TX_VL6_PKT_NUM: hex(value)}},
            "die_metrics": {},
            "source_path": source_path,
            "time": group_time,
        }

    def test_rescheduling_picks_first_start_and_end(self):
        # 重调度时间线：start1(10:00) → resuming(10:20) → start2(10:30) → end(11:00)
        self.parser.start_time = "2026-02-14 09:50:00"
        self.parser.end_time = "2026-02-14 11:10:00"
        self.parser.resuming_training_time = "2026-02-14 10:20:00"
        groups = [
            self._kernel_group(SIDE_BEFORE, "2026-02-14 10:00:00.000000", 1, "start1.log"),
            self._kernel_group(SIDE_BEFORE, "2026-02-14 10:30:00.000000", 2, "start2.log"),
            self._kernel_group(SIDE_AFTER, "2026-02-14 11:00:00.000000", 3, "end.log"),
        ]
        before, after, source_path, _ = self.parser._pair_kernel_groups(groups)
        # before 取续训后的首次建链（start2），after 取窗口内的 end
        self.assertEqual(before["udie0"][0][TX_VL6_PKT_NUM], hex(2))
        self.assertEqual(after["udie0"][0][TX_VL6_PKT_NUM], hex(3))
        self.assertIn("start2.log", source_path)
        self.assertIn("end.log", source_path)
        self.assertNotIn("start1.log", source_path)

    def test_groups_outside_window_discarded(self):
        # 所有组都在窗口之前，不能取窗口外数据
        self.parser.start_time = "2026-02-14 09:50:00"
        self.parser.end_time = "2026-02-14 11:10:00"
        self.parser.resuming_training_time = ""
        groups = [
            self._kernel_group(SIDE_BEFORE, "2026-02-14 08:00:00.000000", 1),
            self._kernel_group(SIDE_AFTER, "2026-02-14 08:30:00.000000", 2),
        ]
        before, after, source_path, _ = self.parser._pair_kernel_groups(groups)
        self.assertEqual(before, {})
        self.assertEqual(after, {})
        self.assertEqual(source_path, "")

    def test_no_window_falls_back_to_latest_group(self):
        # 无 start/end 时间窗：退化为每个 die 最后一份 start/end 组（按组时间取最大）
        groups = [
            self._kernel_group(SIDE_BEFORE, "2026-02-14 09:00:00.000000", 1),
            self._kernel_group(SIDE_BEFORE, "2026-02-14 10:30:00.000000", 2),
            self._kernel_group(SIDE_AFTER, "2026-02-14 11:00:00.000000", 3),
        ]
        before, after, _, _ = self.parser._pair_kernel_groups(groups)
        self.assertEqual(before["udie0"][0][TX_VL6_PKT_NUM], hex(2))
        self.assertEqual(after["udie0"][0][TX_VL6_PKT_NUM], hex(3))

    def test_no_window_discards_reversed_groups(self):
        # 无 start/end：最后一份 before 来自新任务建链组、最后一份 after 是旧任务销毁组，倒序须丢弃
        groups = [
            self._kernel_group(SIDE_BEFORE, "2026-02-14 10:00:00.000000", 1),
            self._kernel_group(SIDE_AFTER, "2026-02-14 10:30:00.000000", 2),
            self._kernel_group(SIDE_BEFORE, "2026-02-14 11:00:00.000000", 3),
        ]
        self.assertEqual(self.parser._pair_kernel_groups(groups)[:2], ({}, {}))

    def test_window_picks_earliest_start_and_latest_end(self):
        # 无断点续训：before 取 >= start_time 的最早建链组，after 取 <= end_time 的最晚销毁组
        self.parser.start_time = "2026-02-14 09:50:00"
        self.parser.end_time = "2026-02-14 11:10:00"
        self.parser.resuming_training_time = ""
        groups = [
            self._kernel_group(SIDE_BEFORE, "2026-02-14 10:00:00.000000", 1),
            self._kernel_group(SIDE_BEFORE, "2026-02-14 10:10:00.000000", 2),
            self._kernel_group(SIDE_AFTER, "2026-02-14 10:50:00.000000", 3),
            self._kernel_group(SIDE_AFTER, "2026-02-14 11:00:00.000000", 4),
        ]
        before, after, _, _ = self.parser._pair_kernel_groups(groups)
        self.assertEqual(before["udie0"][0][TX_VL6_PKT_NUM], hex(1))
        self.assertEqual(after["udie0"][0][TX_VL6_PKT_NUM], hex(4))

    def test_duplicate_groups_deduped_before_selection(self):
        # 累积快照中重复的组（side+udie+time 相同）只保留一份；不同 side 不去重
        groups = [
            self._kernel_group(SIDE_BEFORE, "2026-02-14 10:00:00.000000", 1, "a.log"),
            self._kernel_group(SIDE_BEFORE, "2026-02-14 10:00:00.000000", 2, "b.log"),
            self._kernel_group(SIDE_AFTER, "2026-02-14 10:00:00.000000", 3, "c.log"),
        ]
        unique_groups = self.parser._dedup_groups(groups)
        self.assertEqual(len(unique_groups), 2)
        # 去重保留先出现的组，选组结果使用第一份数据的值
        self.parser.start_time = ""
        self.parser.end_time = ""
        self.parser.resuming_training_time = ""
        before, after, _, _ = self.parser._pair_kernel_groups(groups)
        self.assertEqual(before["udie0"][0][TX_VL6_PKT_NUM], hex(1))
        self.assertEqual(after["udie0"][0][TX_VL6_PKT_NUM], hex(3))

    def test_single_side_kernel_falls_back_to_ubctl(self):
        # kernel.log 仅有 status:start 单侧组、无法配对：parse() 必须回退 ubctl 数据
        kernel = "/tmp/hisi_logs/device-0/20260214100000/log/kernel.log"
        before = "/tmp/ub_info/dev-os-0/ubctl/2026_08_23_13_20_00/ubctl_log.txt"
        after = "/tmp/ub_info/dev-os-0/ubctl/2026_08_23_13_40_00/ubctl_log.txt"
        lines = {
            kernel: [
                "2026-02-14 10:00:00 aubdfx_net_dfx_info_print [die id:0] [status:start]",
                "2026-02-14 10:00:00 aubdfx_net_dfx_info_print [die id:0] logic_port_id = 0, pmu_tx_vl_6 = 10",
            ],
            before: ['"Command: ubctl -c 0 -d 0 -m dump -p 0"', "rc_full_queue_cnt: 0x1"],
            after: ['"Command: ubctl -c 0 -d 0 -m dump -p 0"', "rc_full_queue_cnt: 0x2"],
        }
        parser = UbctlLogParser({"start_time": "2026-08-23 13:30:00", "end_time": "2026-08-23 13:35:00"})
        ctx = KGParseCtxStub()
        ctx.parse_file_path = type(
            "PathStub", (), dict(hisi_logs_path=[kernel], ubctl_before_files=[before], ubctl_after_files=[after])
        )()
        with (
            patch.object(parser, "_yield_log", side_effect=lambda path: lines[path]),
            patch.object(parser, "find_log", return_value=[]),
        ):
            events, err = parser.parse(ctx, "task_id")
        self.assertEqual(err, {})
        self.assertEqual(len(events), 1)
        self.assertEqual(events[0]["event_code"], PRECHECK_UBCTL_DATA)
        self.assertEqual(events[0]["attribute"]["after"]["udie0"][0][RC_FULL_QUEUE_CNT], "0x2")


class TestSnapshootParse(unittest.TestCase):
    """nsdk_hccl_shapshot.log 解析"""

    HEADER = f"2026-11-03 09:29:06.920 {SNAP_PREFIX}devid = 0, dieid = {{die}}, start_end = {{side}}"

    def setUp(self):
        self.parser = LCNEParser(EMPTY_PARAMS)

    def _parse_groups(self, lines, parser=None):
        parser = parser or self.parser
        groups = {SIDE_BEFORE: [], SIDE_AFTER: []}
        current = {}
        for line in lines:
            current = parser._parse_snapshoot_line(line, current, groups, "nsdk_hccl_shapshot.log")
        parser._finish_group(current, groups)
        return groups[SIDE_BEFORE] + groups[SIDE_AFTER]

    def _select_trees(self, groups, parser=None):
        parser = parser or self.parser
        before_groups, after_groups = parser._select_groups(groups)
        before = {udie: group["ports_metrics"] for udie, group in before_groups.items()}
        after = {udie: group["ports_metrics"] for udie, group in after_groups.items()}
        return before, after

    @staticmethod
    def _table(prefix, drop, nic, nl):
        return [
            f"{prefix}port                  0            1           2",
            f"{prefix}PPP_RLT_DROP_NUM      {drop[0]}           {drop[1]}          {drop[2]}",
            f"{prefix}NIC_TX_L2_ERR_DROP_PKT_CNT   {nic[0]}      {nic[1]}     {nic[2]}",
            f"{prefix}nl                    10           11          12",
            f"{prefix}PPP_RLT_DROP_NUM      100          200         300",
        ]

    def test_before_after_pair(self):
        lines = [self.HEADER.format(die=1, side=0)]
        lines.extend(self._table(f"2026-11-03 09:29:06.920 {SNAP_PREFIX}", (1, 2, 3), (4, 5, 6), (1, 2, 3)))
        lines.append(self.HEADER.format(die=1, side=1))
        lines.extend(self._table(f"2026-11-03 10:29:06.920 {SNAP_PREFIX}", (2, 2, 4), (5, 5, 6), (1, 2, 3)))
        groups = self._parse_groups(lines)
        self.assertEqual(len(groups), 2)
        # 组时间取组头行绝对时间戳并补全微秒
        self.assertEqual(groups[0]["time"], "2026-11-03 09:29:06.920000")
        self.assertEqual(groups[1]["time"], "2026-11-03 09:29:06.920000")
        before, after = self._select_trees(groups)
        self.assertEqual(before["udie1"][0][PPP_RLT_DROP_NUM], "1")
        self.assertEqual(before["udie1"][2][NIC_TX_L2_ERR_DROP_PKT_CNT], "6")
        self.assertEqual(after["udie1"][0][PPP_RLT_DROP_NUM], "2")
        self.assertEqual(after["udie1"][2][PPP_RLT_DROP_NUM], "4")
        # nl 表下的同名指标不覆盖 port 表结果
        self.assertEqual(after["udie1"][0][PPP_RLT_DROP_NUM], "2")

    def test_multiple_start_groups_pair_with_latest(self):
        start1_header = f"2026-11-03 09:00:00.000 {SNAP_PREFIX}devid = 0, dieid = 0, start_end = 0"
        start2_header = f"2026-11-03 09:10:00.000 {SNAP_PREFIX}devid = 0, dieid = 0, start_end = 0"
        end_header = f"2026-11-03 09:20:00.000 {SNAP_PREFIX}devid = 0, dieid = 0, start_end = 1"
        lines = [start1_header]
        lines.append(f"2026-11-03 09:00:00 {SNAP_PREFIX}port 0")
        lines.append(f"2026-11-03 09:00:00 {SNAP_PREFIX}PPP_RLT_DROP_NUM 5")
        lines.append(start2_header)
        lines.append(f"2026-11-03 09:10:00 {SNAP_PREFIX}port 0")
        lines.append(f"2026-11-03 09:10:00 {SNAP_PREFIX}PPP_RLT_DROP_NUM 8")
        lines.append(end_header)
        lines.append(f"2026-11-03 09:20:00 {SNAP_PREFIX}port 0")
        lines.append(f"2026-11-03 09:20:00 {SNAP_PREFIX}PPP_RLT_DROP_NUM 9")
        before, after = self._select_trees(self._parse_groups(lines))
        self.assertEqual(before["udie0"][0][PPP_RLT_DROP_NUM], "8")
        self.assertEqual(after["udie0"][0][PPP_RLT_DROP_NUM], "9")

    def test_single_side_no_trees(self):
        lines = [self.HEADER.format(die=0, side=1)]
        lines.append(f"2026-11-03 09:20:00 {SNAP_PREFIX}port 0")
        lines.append(f"2026-11-03 09:20:00 {SNAP_PREFIX}PPP_RLT_DROP_NUM 9")
        before, after = self._select_trees(self._parse_groups(lines))
        self.assertEqual(before, {})
        self.assertEqual(after, {})

    def test_groups_outside_window_produce_no_event(self):
        # 所有组都在窗口之前：不产出事件
        parser = LCNEParser(
            {
                "default_conf": {},
                "user_conf": {},
                "start_time": "2026-11-03 09:50:00",
                "end_time": "2026-11-03 11:10:00",
            }
        )
        lines = [self.HEADER.format(die=0, side=0)]
        lines.append(f"2026-11-03 09:00:00 {SNAP_PREFIX}port 0")
        lines.append(f"2026-11-03 09:00:00 {SNAP_PREFIX}PPP_RLT_DROP_NUM 1")
        lines.append(self.HEADER.format(die=0, side=1))
        lines.append(f"2026-11-03 09:30:00 {SNAP_PREFIX}port 0")
        lines.append(f"2026-11-03 09:30:00 {SNAP_PREFIX}PPP_RLT_DROP_NUM 2")
        with (
            patch.object(parser, "_yield_log", return_value=lines),
            patch.object(parser, "find_log", return_value=[SNAP_LOG]),
        ):
            events, err = parser.parse(KGParseCtxStub(), "task_id")
        self.assertEqual(events, [])
        self.assertEqual(err, {})

    def test_parse_emits_snapshoot_event(self):
        file_lines = [self.HEADER.format(die=0, side=0)]
        file_lines.append(f"2026-11-03 09:00:00 {SNAP_PREFIX}port 0")
        file_lines.append(f"2026-11-03 09:00:00 {SNAP_PREFIX}PPP_RLT_DROP_NUM 1")
        file_lines.append(self.HEADER.format(die=0, side=1))
        file_lines.append(f"2026-11-03 09:20:00 {SNAP_PREFIX}port 0")
        file_lines.append(f"2026-11-03 09:20:00 {SNAP_PREFIX}PPP_RLT_DROP_NUM 2")
        with (
            patch.object(self.parser, "_yield_log", return_value=file_lines),
            patch.object(self.parser, "find_log", return_value=[SNAP_LOG]),
        ):
            events, err = self.parser.parse(KGParseCtxStub(), "task_id")
        self.assertEqual(err, {})
        self.assertEqual(len(events), 1)
        self.assertEqual(events[0]["event_code"], PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA)
        self.assertEqual(events[0]["source_device"], "0")
        self.assertEqual(events[0]["attribute"]["after"]["udie0"][0][PPP_RLT_DROP_NUM], "2")
        self.assertEqual(events[0]["occurrence"], [])

    def test_selected_groups_source_lines_and_occur_time(self):
        # occur_time 取选中 after 组的组头时间
        before_header = f"2026-11-03 09:00:00.100 {SNAP_PREFIX}devid = 0, dieid = 0, start_end = 0"
        before_lines = [
            before_header,
            f"2026-11-03 09:00:00.100 {SNAP_PREFIX}port 0",
            f"2026-11-03 09:00:00.100 {SNAP_PREFIX}PPP_RLT_DROP_NUM 1",
        ]
        after_header = f"2026-11-03 10:00:00.200 {SNAP_PREFIX}devid = 0, dieid = 0, start_end = 1"
        after_lines = [
            after_header,
            f"2026-11-03 10:00:00.200 {SNAP_PREFIX}port 0",
            f"2026-11-03 10:00:00.200 {SNAP_PREFIX}PPP_RLT_DROP_NUM 2",
        ]
        with (
            patch.object(self.parser, "_yield_log", return_value=before_lines + after_lines),
            patch.object(self.parser, "find_log", return_value=[SNAP_LOG]),
        ):
            events, _ = self.parser.parse(KGParseCtxStub(), "task_id")
        # 与 ubctl 事件结构一致
        self.assertEqual(events[0]["attribute"]["source_lines"], {"before": [], "after": []})
        self.assertEqual(events[0]["occur_time"], "2026-11-03 10:00:00.200000")

    def test_metric_name_exact_match_only(self):
        # 精确认指标名：正确写法命中，带后缀的近似名不误匹配
        prefix = f"2026-11-03 09:29:06.920 {SNAP_PREFIX}"
        lines = [
            self.HEADER.format(die=0, side=0),
            f"{prefix}port                  0            1",
            f"{prefix}NIC_TX_L2_ERR_DROP_PKT_CNT     7            8",
            f"{prefix}NIC_TX_L2_ERR_DROP_PKT_CNT_OLD     1          1",
            f"{prefix}PPP_RLT_DROP_NUM_EXTRA     2            3",
        ]
        groups = self._parse_groups(lines)
        metrics = groups[0]["ports_metrics"]
        self.assertEqual(metrics[0][NIC_TX_L2_ERR_DROP_PKT_CNT], "7")
        self.assertEqual(metrics[1][NIC_TX_L2_ERR_DROP_PKT_CNT], "8")
        self.assertNotIn(PPP_RLT_DROP_NUM, metrics[0])

    def test_short_metric_row_fills_only_declared_ports(self):
        # 指标行列数少于表头 port 数时，只填充已声明的列，避免越界或错列
        prefix = f"2026-11-03 09:29:06.920 {SNAP_PREFIX}"
        lines = [
            self.HEADER.format(die=0, side=0),
            f"{prefix}port                  0            1           2",
            f"{prefix}PPP_RLT_DROP_NUM     2            3",
        ]
        groups = self._parse_groups(lines)
        metrics = groups[0]["ports_metrics"]
        self.assertEqual(metrics[0][PPP_RLT_DROP_NUM], "2")
        self.assertEqual(metrics[1][PPP_RLT_DROP_NUM], "3")
        self.assertNotIn(2, metrics)


class KGParseCtxStub:
    is_sdk_input = False
    parse_file_path = None
    resuming_training_time = ""


class TestLcneSaverCollectSnapshoot(unittest.TestCase):
    """LCNELogSaver 需要收上 nsdk_hccl_shapshot.log（含 varlog/slot_1 层级）"""

    def test_filter_log_collects_snapshoot(self):
        from ascend_fd.pkg.parse.parser_saver import LCNELogSaver

        with tempfile.TemporaryDirectory() as tmp_dir:
            snap_dir = os.path.join(tmp_dir, "diagnostic_information", "varlog", "slot_1", "varlog")
            os.makedirs(snap_dir)
            snap_file = os.path.join(snap_dir, "nsdk_hccl_shapshot.log")
            with open(snap_file, "w", encoding="utf-8") as file:
                file.write("snapshoot")
            saver = LCNELogSaver()
            saver.filter_log(tmp_dir)
            # snapshoot 与普通 lcne 日志统一进 lcne 列表，由 LCNEParser 内部分发
            self.assertIn(snap_file, saver.get_lcne_log_list())

    def test_filter_log_collects_snapshoot_from_varlog_zip(self):
        """slot_1 下是 varlog.zip 压缩包：解压后仍需收集 snapshoot，且 cpdt_checkcc.log 不受影响"""
        from ascend_fd.pkg.parse.parser_saver import LCNELogSaver

        with tempfile.TemporaryDirectory() as tmp_dir:
            slot_dir = os.path.join(tmp_dir, "diagnostic_information", "varlog", "slot_1")
            os.makedirs(slot_dir)
            zip_path = os.path.join(slot_dir, "varlog.zip")
            with zipfile.ZipFile(zip_path, "w") as zip_file:
                zip_file.writestr("cpdt_checkcc.log", "cpdt")
                zip_file.writestr("nsdk_hccl_shapshot.log", SNAPSHOOT_TAG + "devid = 0, dieid = 0, start_end = 0")
            saver = LCNELogSaver()
            saver.filter_log(tmp_dir)

            snap_list = [
                path
                for path in saver.get_lcne_log_list()
                if path.endswith(os.path.join("varlog", "nsdk_hccl_shapshot.log"))
            ]
            self.assertEqual(len(snap_list), 1)
            self.assertTrue(snap_list[0].endswith(os.path.join("varlog", "nsdk_hccl_shapshot.log")))
            self.assertTrue(os.path.isfile(snap_list[0]))
            self.assertTrue(saver.get_cpdt_checkcc_log().endswith("cpdt_checkcc.log"))
            self.assertTrue(os.path.isfile(saver.get_cpdt_checkcc_log()))
