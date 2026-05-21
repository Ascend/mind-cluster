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

from ascend_fd.pkg.parse.root_cluster.parser import PidFileParser, BaseInfoParser
from ascend_fd.pkg.parse.blacklist.blacklist_op import BlackListManager

TEST_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
TESTCASE_KG_PARSE_INPUT = os.path.join(TEST_DIR, "st_module_testcase", "rc_parse")

A5_ROOT_INFO_DETECT_LINE = (
    "[RootInfoDetect] nRanks[16], rank[15] entry flat topo detect, "
    "rootinfo: host ip[192.168.0.1] port[30000] netMode[HrtNetworkMode::HDC] "
    "identifier[192.168.0.1_30000_0_947147113161045], deviceLogicId[7], devPhyId[15]"
)


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


class TestA5RootInfoDetect(unittest.TestCase):
    def _create_parser(self, device_info_map=None):
        return BaseInfoParser({}, BlackListManager(), device_info_map)

    def test_parse_full_line(self):
        parser = self._create_parser()
        parser.parse_line(A5_ROOT_INFO_DETECT_LINE)

        self.assertEqual(parser.logic_device_id, "7")
        self.assertEqual(parser.phy_device_id, "15")
        self.assertEqual(parser.server_id, "192.168.0.1")
        self.assertIn("192.168.0.1_30000_0_947147113161045", parser.rank_map)
        self.assertEqual(parser.rank_map["192.168.0.1_30000_0_947147113161045"]["rank_id"], "15")
        self.assertEqual(parser.rank_map["192.168.0.1_30000_0_947147113161045"]["rank_num"], 16)

    def test_parse_line_wildcard_identifier(self):
        line = (
            "[RootInfoDetect] nRanks[8], rank[0] entry flat topo detect, "
            "rootinfo: host ip[10.0.0.1] port[20000] netMode[HrtNetworkMode::HDC] "
            "identifier[*], deviceLogicId[3], devPhyId[9]"
        )
        parser = self._create_parser()
        parser.parse_line(line)

        self.assertIn("*", parser.rank_map)
        self.assertEqual(parser.rank_map["*"]["rank_id"], "0")
        self.assertEqual(parser.rank_map["*"]["rank_num"], 8)
        self.assertEqual(parser.logic_device_id, "3")
        self.assertEqual(parser.phy_device_id, "9")

    def test_parse_line_invalid_logic_id(self):
        line = (
            "[RootInfoDetect] nRanks[4], rank[2] entry flat topo detect, "
            "rootinfo: host ip[10.0.0.2] port[30000] netMode[HrtNetworkMode::HDC] "
            "identifier[test_id], deviceLogicId[-1], devPhyId[5]"
        )
        parser = self._create_parser()
        parser.parse_line(line)

        self.assertEqual(parser.logic_device_id, "")
        self.assertEqual(parser.phy_device_id, "5")

    def test_parse_line_invalid_host_ip(self):
        line = (
            "[RootInfoDetect] nRanks[4], rank[2] entry flat topo detect, "
            "rootinfo: host ip[0.0.0.0] port[30000] netMode[HrtNetworkMode::HDC] "
            "identifier[test_id], deviceLogicId[3], devPhyId[5]"
        )
        parser = self._create_parser()
        parser.parse_line(line)

        self.assertEqual(parser.device_ip, "")

    def test_parse_line_rank_num_not_integer(self):
        line = (
            "[RootInfoDetect] nRanks[abc], rank[2] entry flat topo detect, "
            "rootinfo: host ip[10.0.0.2] port[30000] netMode[HrtNetworkMode::HDC] "
            "identifier[test_id], deviceLogicId[3], devPhyId[5]"
        )
        parser = self._create_parser()
        parser.parse_line(line)

        self.assertEqual(parser.logic_device_id, "3")
        self.assertEqual(parser.rank_map["test_id"]["rank_num"], -1)

    def test_parse_generation_info_a5(self):
        line = "[HCCL_TRACE]V950 Entry-HcclCommInitRootInfo:ranks[16], rank[0]"
        parser = self._create_parser()
        parser.parse_line(line)

        self.assertEqual(parser.generation_info, "A5")

    def test_parse_generation_info_default(self):
        line = "[HCCL_TRACE]Entry-HcclCommInitRootInfo:ranks[16], rank[0]"
        parser = self._create_parser()
        parser.parse_line(line)

        self.assertEqual(parser.generation_info, "A2/A3")

    def test_parse_base_info_a5(self):
        line = "Entry-HcclCommInitRootInfoConfigV2:ranks[8], rank[2], rootinfo: host ip[1.1.1.1] , identifier[group_0]"
        parser = self._create_parser()
        parser.parse_line(line)

        self.assertEqual(parser.rank_map["group_0"]["rank_num"], 8)
        self.assertEqual(parser.rank_map["group_0"]["rank_id"], "2")

    def test_parse_a5_timeout_info(self):
        line = "[HCCL_TRACE]Env config hcclSocketFamily[*], linkTimeOut[180]s, execTimeOut[300]s"
        parser = self._create_parser()
        parser.parse_line(line)

        self.assertEqual(parser.timeout_params.get("CONNECT_TIMEOUT"), 180)
        self.assertEqual(parser.timeout_params.get("EXEC_TIMEOUT"), 300)

    def test_parse_eid_plane_info(self):
        line = "[HCCL_TRACE]Net2PeerLink rankId[0] eid[0x1234] planeId[1]"
        parser = self._create_parser()
        parser.parse_line(line)

        self.assertIn("0", parser.rank_eid_plane_info)
        self.assertEqual(len(parser.rank_eid_plane_info["0"]), 1)


class TestResolvePhyDeviceId(unittest.TestCase):
    def _create_parser(self, device_info_map=None):
        return BaseInfoParser({}, BlackListManager(), device_info_map)

    def test_phy_already_set_not_overridden(self):
        parser = self._create_parser()
        parser.phy_device_id = "15"
        parser.logic_device_id = "7"
        parser._resolve_phy_device_id()

        self.assertEqual(parser.phy_device_id, "15")

    def test_no_logic_id_phy_stays_empty(self):
        parser = self._create_parser()
        parser.phy_device_id = ""
        parser.logic_device_id = ""
        parser._resolve_phy_device_id()

        self.assertEqual(parser.phy_device_id, "")

    def test_resolve_from_device_info_map(self):
        parser = self._create_parser({"7": "15", "3": "9"})
        parser.phy_device_id = ""
        parser.logic_device_id = "7"
        parser._resolve_phy_device_id()

        self.assertEqual(parser.phy_device_id, "15")

    def test_no_mapping_in_map_fallbacks_to_logic_id(self):
        parser = self._create_parser({"3": "9"})
        parser.phy_device_id = ""
        parser.logic_device_id = "7"
        parser._resolve_phy_device_id()
        self.assertEqual(parser.phy_device_id, "7")

    def test_no_device_info_map_fallbacks_to_logic_id(self):
        parser = self._create_parser()
        parser.phy_device_id = ""
        parser.logic_device_id = "7"
        parser._resolve_phy_device_id()
        self.assertEqual(parser.phy_device_id, "7")
