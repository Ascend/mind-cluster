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
import unittest

from ascend_fd.pkg.parse.knowledge_graph.parser.cann_log_parser import PrecheckEnvParser
from ascend_fd.pkg.parse.knowledge_graph.parser.lcne_parser import LCNEParser
from ascend_fd.pkg.parse.knowledge_graph.parser.npu_device_parse import NpuHistoryLogParser, UbctlLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.npu_info_parser import HcclStatInfoParser
from ascend_fd.utils.constant.str_const import UNKNOWN_DEVICE_ID
from ascend_fd.utils.constant.ub_const import (
    AGE_PERIOD_KEYWORD,
    HCOMM_TA_CTP_UB_TIMEOUT,
    ICRC_ERR_COUNT_METRICS,
    PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT,
    PRECHECK_KERNEL_AE_TYPE23,
    PRECHECK_RXDMA_ICRC_DATA,
    PRECHECK_UBCTL_DATA,
    PRECHECK_UBMEM_TIMEOUT,
    RC_FULL_QUEUE_CNT,
    SIDE_AFTER,
    SIDE_BEFORE,
    TP_RRP_ERR_FLG_0,
)

EMPTY_PARAMS = {"default_conf": {}, "user_conf": {}}


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
