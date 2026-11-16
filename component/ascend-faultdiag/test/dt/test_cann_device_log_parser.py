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
import unittest
from unittest import mock

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.knowledge_graph.parser import cann_log_parser
from ascend_fd.pkg.parse.knowledge_graph.parser.cann_log_parser import CANNDeviceLogParser

TEST_PID = "16241"
TEST_DEVICE_LOG = "device-16241_20261114145648.log"
TEST_EVENT = {"event_code": "AISW_CANN_AICPU_07", "source_device": "Unknown"}
TEST_TIME = ("2026-11-14 14:56:07", "2026-11-14 14:56:08")


class FakeMultiProcessJob:
    """Run security jobs serially in current process for non-sdk input test."""

    def __init__(self, *args, **kwargs):
        self.result = dict()

    def add_security_job(self, name, func, *args):
        self.result.update({name: func(*args)})

    def join_and_get_results(self):
        return dict(self.result), dict()


class TestCANNDeviceLogParserPidFilter(unittest.TestCase):
    def setUp(self):
        self.params = dict()
        self.parser = CANNDeviceLogParser(self.params)
        self.parser.find_log = mock.Mock(return_value={TEST_PID: [TEST_DEVICE_LOG]})
        self.parser._parse_files_of_pid = mock.Mock(return_value=([TEST_EVENT], TEST_TIME, (TEST_PID, "")))

    @staticmethod
    def _parse(parser, is_sdk_input):
        parse_ctx = KGParseCtx(is_sdk_input=is_sdk_input)
        events, err_dict = parser.parse(parse_ctx, "")
        return events, err_dict

    def test_sdk_input_without_useful_pid_keeps_events(self):
        with self.assertLogs("KNOWLEDGE_GRAPH", level="WARNING") as log_ctx:
            events, _ = self._parse(self.parser, True)
        self.assertEqual([TEST_EVENT], events)
        self.assertIn("keep all CANN_Device events", log_ctx.output[0])

    def test_sdk_input_with_matched_useful_pid_keeps_events(self):
        self.params["useful_pid"] = [TEST_PID]
        events, _ = self._parse(self.parser, True)
        self.assertEqual([TEST_EVENT], events)

    def test_sdk_input_with_unmatched_useful_pid_keeps_events(self):
        self.params["useful_pid"] = ["999"]
        with self.assertLogs("KNOWLEDGE_GRAPH", level="WARNING") as log_ctx:
            events, _ = self._parse(self.parser, True)
        self.assertEqual([TEST_EVENT], events)
        self.assertIn("keep all CANN_Device events", log_ctx.output[0])

    def test_sdk_input_with_empty_useful_pid_keeps_events(self):
        self.params["useful_pid"] = []
        events, _ = self._parse(self.parser, True)
        self.assertEqual([TEST_EVENT], events)

    def test_sdk_input_with_allocated_fake_pid_keeps_events(self):
        self.params["useful_pid"] = [0]
        with self.assertLogs("KNOWLEDGE_GRAPH", level="WARNING"):
            events, _ = self._parse(self.parser, True)
        self.assertEqual([TEST_EVENT], events)

    def test_non_sdk_input_without_useful_pid_drops_events(self):
        with mock.patch.object(cann_log_parser, "MultiProcessJob", FakeMultiProcessJob):
            events, _ = self._parse(self.parser, False)
        self.assertEqual([], events)

    def test_non_sdk_input_with_unmatched_useful_pid_drops_events(self):
        self.params["useful_pid"] = ["999"]
        with mock.patch.object(cann_log_parser, "MultiProcessJob", FakeMultiProcessJob):
            events, _ = self._parse(self.parser, False)
        self.assertEqual([], events)

    def test_non_sdk_input_with_matched_useful_pid_keeps_events(self):
        self.params["useful_pid"] = [TEST_PID]
        with mock.patch.object(cann_log_parser, "MultiProcessJob", FakeMultiProcessJob):
            events, _ = self._parse(self.parser, False)
        self.assertEqual([TEST_EVENT], events)


if __name__ == "__main__":
    unittest.main()
