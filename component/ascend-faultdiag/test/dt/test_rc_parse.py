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
import unittest
import os

from ascend_fd.pkg.parse.root_cluster.parser import PidFileParser

TEST_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
TESTCASE_KG_PARSE_INPUT = os.path.join(TEST_DIR, "st_module_testcase", "rc_parse")


class RcParseTestCase(unittest.TestCase):

    def test_base_info_parser_func(self):
        pid_file_parser = PidFileParser("test_pid", {})
        pid_file_parser.parse_log(os.path.join(TESTCASE_KG_PARSE_INPUT, "example.log"))
        result = pid_file_parser.get_result()

        base_result = result.base
        error_result = result.error
        self.assertEqual("2024-08-01-15:45:28.498874", result.lagging_time)
        self.assertEqual(base_result.logic_device_id, "0")
        self.assertEqual(base_result.timeout_param.get("CONNECT_TIMEOUT"), 120)
        self.assertEqual(base_result.timeout_param.get("EXEC_TIMEOUT"), 120)
        self.assertEqual(base_result.timeout_param.get("RDMA_TIMEOUT"), 20)
        self.assertEqual(base_result.timeout_param.get("RDMA_RETRY_CNT"), 7)
        self.assertIn("172.16.13.183%eth0_64000_0_1721821172092650", base_result.rank_map)
        self.assertEqual(base_result.server_id, "172.16.13.183")
        self.assertEqual("2024-03-28-10:25:48.427201", error_result.first_error_time)
        self.assertEqual("HCCL", error_result.first_error_module)
        self.assertIn("1.1.1.1", error_result.cqe_links)
        self.assertIn("2.2.2.2", error_result.cqe_links)
        for event_dict in error_result.timeout_error_events_list:
            if event_dict.get("error_type") == "Notify":
                self.assertEqual("3", event_dict.get("remote_rank"))
                self.assertEqual("AllReduce_10.136.181.175%enp179s0f0_60000_0_1712529353144389", event_dict.get("tag"))
                self.assertEqual("3", event_dict.get("index"))
