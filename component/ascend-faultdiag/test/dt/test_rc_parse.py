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
import os
import tempfile
import unittest
import unittest.mock

from ascend_fd.pkg.parse.root_cluster.parser import PidFileParser, BaseInfoParser, ErrorParser
from ascend_fd.pkg.parse.root_cluster.rc_parse_job import parse_npu_info_file
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

    def test_parse_line_multiple_updates_takes_latest(self):
        first_line = (
            "[RootInfoDetect] nRanks[4], rank[0] entry flat topo detect, "
            "rootinfo: host ip[10.0.0.1] port[30000] netMode[HrtNetworkMode::HDC] "
            "identifier[test_id_1], deviceLogicId[3], devPhyId[9]"
        )
        second_line = (
            "[RootInfoDetect] nRanks[4], rank[1] entry flat topo detect, "
            "rootinfo: host ip[10.0.0.2] port[40000] netMode[HrtNetworkMode::HDC] "
            "identifier[test_id_2], deviceLogicId[7], devPhyId[15]"
        )
        parser = self._create_parser()
        parser.parse_line(first_line)
        self.assertEqual(parser.logic_device_id, "3")
        self.assertEqual(parser.phy_device_id, "9")
        parser.parse_line(second_line)
        self.assertEqual(parser.logic_device_id, "7")
        self.assertEqual(parser.phy_device_id, "15")

    def test_same_logic_id_different_phy_id_takes_latest(self):
        first_line = (
            "[RootInfoDetect] nRanks[4], rank[0] entry flat topo detect, "
            "rootinfo: host ip[10.0.0.1] port[30000] netMode[HrtNetworkMode::HDC] "
            "identifier[test_id_1], deviceLogicId[3], devPhyId[9]"
        )
        second_line = (
            "[RootInfoDetect] nRanks[4], rank[1] entry flat topo detect, "
            "rootinfo: host ip[10.0.0.2] port[40000] netMode[HrtNetworkMode::HDC] "
            "identifier[test_id_2], deviceLogicId[3], devPhyId[21]"
        )
        parser = self._create_parser()
        parser.parse_line(first_line)
        self.assertEqual(parser.logic_device_id, "3")
        self.assertEqual(parser.phy_device_id, "9")
        parser.parse_line(second_line)
        self.assertEqual(parser.logic_device_id, "3")
        self.assertEqual(parser.phy_device_id, "21")

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


class TestPhyDeviceIdValidation(unittest.TestCase):
    """Test cases for phy_device_id non-negative integer validation in BaseInfoParser"""

    def _create_parser(self):
        return BaseInfoParser({}, BlackListManager())

    def test_valid_positive_phy_device_id_accepted(self):
        """Test that valid positive phy_device_id values are accepted"""
        from ascend_fd.utils.regular_table import SOCKET_PHY_ID_INFO

        parser = self._create_parser()
        valid_ids = ["1", "5", "10", "100", "999"]

        for current_phy_id in valid_ids:
            with self.subTest(phy_id=current_phy_id):
                parser.re_init()
                with unittest.mock.patch(
                    'ascend_fd.pkg.parse.root_cluster.parser.filter_single_rank_info'
                ) as mock_filter:

                    def mock_side_effect(line, pattern, phy_id=current_phy_id):
                        if pattern == SOCKET_PHY_ID_INFO:
                            return phy_id
                        return ""

                    mock_filter.side_effect = mock_side_effect
                    parser._parse_a5_root_info("test line")
                    self.assertEqual(
                        parser.phy_device_id, current_phy_id, f"phy_device_id {current_phy_id} should be accepted"
                    )

    def test_zero_phy_device_id_accepted(self):
        """Test that zero is accepted as a valid non-negative integer (no warning)"""
        parser = self._create_parser()
        parser.re_init()

        with unittest.mock.patch('ascend_fd.pkg.parse.root_cluster.parser.filter_single_rank_info') as mock_filter:
            mock_filter.return_value = "0"
            parser._parse_a5_root_info("test line devPhyId[0]")
            self.assertEqual(parser.phy_device_id, "0", "Zero should be accepted as a valid device ID")

    def test_negative_phy_device_id_rejected(self):
        """Test that -1 is rejected, but other negative numbers return value with warning"""
        parser = self._create_parser()

        parser.re_init()
        with unittest.mock.patch('ascend_fd.pkg.parse.root_cluster.parser.filter_single_rank_info') as mock_filter:
            mock_filter.return_value = "-1"
            parser._parse_a5_root_info("test line devPhyId[-1]")
            self.assertEqual(parser.phy_device_id, "", "-1 should be rejected as INVALID_ID")

        parser.re_init()
        with unittest.mock.patch('ascend_fd.pkg.parse.root_cluster.parser.filter_single_rank_info') as mock_filter:
            mock_filter.return_value = "-100"
            with self.assertLogs('ROOT_CLUSTER', level='WARNING') as cm:
                parser._parse_a5_root_info("test line devPhyId[-100]")
            self.assertEqual(parser.phy_device_id, "-100", "Other negative values are returned with warning")
            self.assertTrue(
                any("Except devPhyId non-negative integer" in msg and "-100" in msg for msg in cm.output),
                "Warning should be logged for negative value",
            )

    def test_invalid_id_constant_rejected(self):
        """Test that INVALID_ID (-1) is rejected"""
        from ascend_fd.pkg.parse.root_cluster.parser import INVALID_ID

        parser = self._create_parser()
        parser.re_init()

        with unittest.mock.patch('ascend_fd.pkg.parse.root_cluster.parser.filter_single_rank_info') as mock_filter:
            mock_filter.return_value = INVALID_ID
            parser._parse_a5_root_info("test line devPhyId[-1]")
            self.assertEqual(parser.phy_device_id, "", "INVALID_ID (-1) should be rejected")

    def test_non_numeric_phy_device_id_rejected(self):
        """Test that non-numeric strings are returned with warning (except empty/None)"""
        parser = self._create_parser()
        non_numeric_values = ["abc", "", None, "12.34", "NaN"]

        for value in non_numeric_values:
            with self.subTest(value=value):
                parser.re_init()
                with unittest.mock.patch(
                    'ascend_fd.pkg.parse.root_cluster.parser.filter_single_rank_info'
                ) as mock_filter:
                    mock_filter.return_value = value
                    if not value:  # Empty or None should be rejected
                        parser._parse_a5_root_info("test line")
                        self.assertEqual(parser.phy_device_id, "", f"Empty/None value {repr(value)} should be rejected")
                    else:  # Non-empty non-numeric values are returned with warning
                        with self.assertLogs('ROOT_CLUSTER', level='WARNING'):
                            parser._parse_a5_root_info("test line")
                        expected_result = str(value)
                        self.assertEqual(
                            parser.phy_device_id,
                            expected_result,
                            f"Non-numeric value {repr(value)} should be returned with warning",
                        )

    def test_zero_phy_device_id_accepted_no_warning(self):
        """Test that zero is accepted without triggering warning log"""
        parser = self._create_parser()
        parser.re_init()

        with unittest.mock.patch('ascend_fd.pkg.parse.root_cluster.parser.filter_single_rank_info') as mock_filter:
            mock_filter.return_value = "0"
            parser._parse_a5_root_info("test line devPhyId[0]")
            self.assertEqual(parser.phy_device_id, "0", "Zero should be accepted")

    def test_non_numeric_phy_device_id_with_warning_log(self):
        """Test that non-numeric values trigger warning logs with details"""
        parser = self._create_parser()
        non_numeric_values = ["abc", "12.34"]

        for value in non_numeric_values:
            with self.subTest(value=value):
                parser.re_init()
                with unittest.mock.patch(
                    'ascend_fd.pkg.parse.root_cluster.parser.filter_single_rank_info'
                ) as mock_filter:
                    mock_filter.return_value = value
                    with self.assertLogs('ROOT_CLUSTER', level='WARNING') as cm:
                        parser._parse_a5_root_info(f"test line devPhyId[{value}]")
                    self.assertEqual(parser.phy_device_id, value, f"Non-numeric {repr(value)} should be returned")
                    self.assertTrue(
                        any("Except devPhyId non-negative integer" in msg and str(value) in msg for msg in cm.output),
                        f"Warning should contain value {repr(value)}",
                    )


class TestTransportErrorPattern(unittest.TestCase):
    def _create_parser(self):
        return ErrorParser(BlackListManager())

    def test_ipv4_match_sets_transport_error_remote(self):
        parser = self._create_parser()
        line = "[ERROR] HCCL Transport init error: remoteIpAddr[192.168.1.100/3]"
        parser._filter_transport_error_from_log(line)

        self.assertTrue(parser.transport_init_error_happened)
        self.assertIsNotNone(parser.transport_error_remote)
        self.assertEqual(parser.transport_error_remote.server_ip, "192.168.1.100")
        self.assertEqual(parser.transport_error_remote.phy_device_id, "3")

    def test_ipv6_match_sets_transport_error_remote(self):
        parser = self._create_parser()
        line = "[ERROR] HCCL Transport init error: remoteIpAddr[2001:db8::1/5]"
        parser._filter_transport_error_from_log(line)

        self.assertTrue(parser.transport_init_error_happened)
        self.assertIsNotNone(parser.transport_error_remote)
        self.assertEqual(parser.transport_error_remote.server_ip, "2001:db8::1")
        self.assertEqual(parser.transport_error_remote.phy_device_id, "5")

    def test_ipv6_full_address_match(self):
        parser = self._create_parser()
        line = "[ERROR] HCCL Transport init error: remoteIpAddr[fe80:0000:0000:0000:0202:b3ff:fe1e:8329/2]"
        parser._filter_transport_error_from_log(line)

        self.assertTrue(parser.transport_init_error_happened)
        self.assertIsNotNone(parser.transport_error_remote)
        self.assertEqual(parser.transport_error_remote.server_ip, "fe80:0000:0000:0000:0202:b3ff:fe1e:8329")
        self.assertEqual(parser.transport_error_remote.phy_device_id, "2")

    def test_no_transport_init_error_keyword_skipped(self):
        parser = self._create_parser()
        line = "[ERROR] HCCL some other error: remoteIpAddr[192.168.1.100/3]"
        parser._filter_transport_error_from_log(line)

        self.assertFalse(parser.transport_init_error_happened)
        self.assertIsNone(parser.transport_error_remote)

    def test_transport_error_remote_already_set_skipped(self):
        parser = self._create_parser()
        line1 = "[ERROR] HCCL Transport init error: remoteIpAddr[192.168.1.100/3]"
        parser._filter_transport_error_from_log(line1)

        line2 = "[ERROR] HCCL Transport init error: remoteIpAddr[10.0.0.1/7]"
        parser._filter_transport_error_from_log(line2)

        self.assertEqual(parser.transport_error_remote.server_ip, "192.168.1.100")
        self.assertEqual(parser.transport_error_remote.phy_device_id, "3")

    def test_ipv4_match_invalid_ip_filtered_out(self):
        parser = self._create_parser()
        line = "[ERROR] HCCL Transport init error: remoteIpAddr[999.999.999.999/3]"
        parser._filter_transport_error_from_log(line)

        self.assertTrue(parser.transport_init_error_happened)
        self.assertIsNone(parser.transport_error_remote)

    def test_ipv6_match_invalid_ip_filtered_out(self):
        parser = self._create_parser()
        line = "[ERROR] HCCL Transport init error: remoteIpAddr[gggg::1/3]"
        parser._filter_transport_error_from_log(line)

        self.assertTrue(parser.transport_init_error_happened)
        self.assertIsNone(parser.transport_error_remote)

    def test_no_remote_ip_addr_pattern_no_match(self):
        parser = self._create_parser()
        line = "[ERROR] HCCL Transport init error: no remote ip address here"
        parser._filter_transport_error_from_log(line)

        self.assertTrue(parser.transport_init_error_happened)
        self.assertIsNone(parser.transport_error_remote)


class TestServerInfoValidation(unittest.TestCase):
    def _create_parser(self):
        return BaseInfoParser({}, BlackListManager())

    def test_valid_ipv4_server_info(self):
        parser = self._create_parser()
        line = "[HCCL_TRACE] rankNum[4], rank[0], rootInfo identifier[test_id], server[192.168.1.100], logicDevId[3]"
        parser._parse_common_init_info(line)

        self.assertEqual(parser.server_id, "192.168.1.100")

    def test_valid_ipv6_server_info(self):
        parser = self._create_parser()
        line = "[HCCL_TRACE] rankNum[4], rank[0], rootInfo identifier[test_id], server[2001:db8::1], logicDevId[3]"
        parser._parse_common_init_info(line)

        self.assertEqual(parser.server_id, "2001:db8::1")

    def test_invalid_ip_server_info_set_to_empty(self):
        parser = self._create_parser()
        line = "[HCCL_TRACE] rankNum[4], rank[0], rootInfo identifier[test_id], server[not_an_ip], logicDevId[3]"
        parser._parse_common_init_info(line)

        self.assertEqual(parser.server_id, "")

    def test_server_info_with_network_adapter_split(self):
        parser = self._create_parser()
        line = (
            "[HCCL_TRACE] rankNum[4], rank[0], rootInfo identifier[test_id], server[192.168.1.100%eth0], logicDevId[3]"
        )
        parser._parse_common_init_info(line)

        self.assertEqual(parser.server_id, "192.168.1.100")

    def test_server_info_ipv6_with_network_adapter(self):
        parser = self._create_parser()
        line = "[HCCL_TRACE] rankNum[4], rank[0], rootInfo identifier[test_id], server[fe80::1%eth0], logicDevId[3]"
        parser._parse_common_init_info(line)

        self.assertEqual(parser.server_id, "fe80::1")

    def test_no_server_info_fallbacks_to_server_id_info(self):
        parser = self._create_parser()
        line = (
            "hcclCommInitInfo:commId[test], rank[0], totalRanks[4], "
            "serverId[192.168.1.100], deviceType[1], logicDevId[3], identifier[test_id]"
        )
        parser._parse_common_init_info(line)

        self.assertEqual(parser.server_id, "192.168.1.100")


class TestParseNpuInfoFile(unittest.TestCase):
    def test_ipv4_ipaddr_extraction(self):
        content = (
            "hccn_tool -i 0 -ip -g\n"
            "ipaddr:192.168.1.100\n"
            "netmask:255.255.255.0\n"
            "\n"
            "hccn_tool -i 1 -ip -g\n"
            "ipaddr:10.0.0.1\n"
            "netmask:255.255.0.0\n"
        )
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False, encoding="utf-8") as tmp_file:
            tmp_file.write(content)

        try:
            result = parse_npu_info_file(tmp_file.name)

            self.assertEqual(result, {"0": "192.168.1.100", "1": "10.0.0.1"})
        finally:
            os.unlink(tmp_file.name)

    def test_ipv6_ipaddr_extraction(self):
        content = (
            "hccn_tool -i 0 -ip -inet6 -g\n"
            "ipv6_address:2001:db8::1\n"
            "prefix_length:64\n"
            "\n"
            "hccn_tool -i 2 -ip -inet6 -g\n"
            "ipv6_address:fe80::1\n"
            "prefix_length:64\n"
        )
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False, encoding="utf-8") as tmp_file:
            tmp_file.write(content)

        try:
            result = parse_npu_info_file(tmp_file.name)

            self.assertEqual(result, {"0": "2001:db8::1", "2": "fe80::1"})
        finally:
            os.unlink(tmp_file.name)

    def test_invalid_ip_filtered_out(self):
        content = (
            "hccn_tool -i 0 -ip -g\n"
            "ipaddr:not_an_ip\n"
            "netmask:255.255.255.0\n"
            "\n"
            "hccn_tool -i 1 -ip -g\n"
            "ipaddr:192.168.1.100\n"
            "netmask:255.255.255.0\n"
        )
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False, encoding="utf-8") as tmp_file:
            tmp_file.write(content)

        try:
            result = parse_npu_info_file(tmp_file.name)

            self.assertEqual(result, {"1": "192.168.1.100"})
        finally:
            os.unlink(tmp_file.name)

    def test_mixed_ipv4_ipv6_extraction(self):
        content = (
            "hccn_tool -i 0 -ip -g\n"
            "ipaddr:192.168.1.100\n"
            "netmask:255.255.255.0\n"
            "\n"
            "hccn_tool -i 1 -ip -inet6 -g\n"
            "ipv6_address:fe80::1\n"
            "prefix_length:64\n"
        )
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False, encoding="utf-8") as tmp_file:
            tmp_file.write(content)

        try:
            result = parse_npu_info_file(tmp_file.name)

            self.assertEqual(result, {"0": "192.168.1.100", "1": "fe80::1"})
        finally:
            os.unlink(tmp_file.name)

    def test_no_ipaddr_section_skipped(self):
        content = "hccn_tool -i 0 -g\nchip_id:0\n\nhccn_tool -i 1 -g\nchip_id:1\n"
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False, encoding="utf-8") as tmp_file:
            tmp_file.write(content)

        try:
            result = parse_npu_info_file(tmp_file.name)

            self.assertEqual(result, {})
        finally:
            os.unlink(tmp_file.name)

    def test_ipv6_inet6_ipv6_address_extraction(self):
        content = (
            "hccn_tool -i 0 -ip -inet6 -g\n"
            "ipv6_address:2001:0db8:85a3:0000:0000:8a2e:c0a8:01c5\n"
            "prefix_length:64\n"
            "\n"
            "hccn_tool -i 2 -ip -inet6 -g\n"
            "ipv6_address:2001:db8::1\n"
            "prefix_length:64\n"
        )
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False, encoding="utf-8") as tmp_file:
            tmp_file.write(content)

        try:
            result = parse_npu_info_file(tmp_file.name)

            self.assertEqual(result, {"0": "2001:0db8:85a3:0000:0000:8a2e:c0a8:01c5", "2": "2001:db8::1"})
        finally:
            os.unlink(tmp_file.name)

    def test_mixed_ipv4_and_inet6_extraction(self):
        content = (
            "hccn_tool -i 0 -ip -g\n"
            "ipaddr:192.168.1.100\n"
            "netmask:255.255.255.0\n"
            "\n"
            "hccn_tool -i 1 -ip -inet6 -g\n"
            "ipv6_address:2001:0db8:85a3:0000:0000:8a2e:c0a8:01c6\n"
            "prefix_length:64\n"
        )
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False, encoding="utf-8") as tmp_file:
            tmp_file.write(content)

        try:
            result = parse_npu_info_file(tmp_file.name)

            self.assertEqual(result, {"0": "192.168.1.100", "1": "2001:0db8:85a3:0000:0000:8a2e:c0a8:01c6"})
        finally:
            os.unlink(tmp_file.name)


class TestSocketTimeoutIpMatch(unittest.TestCase):
    """Test cases for IPv4/IPv6 address extraction in socket timeout error parsing"""

    def _create_error_parser(self):
        return ErrorParser(BlackListManager())

    def test_ipv4_no_connect_extracts_remote_info(self):
        parser = self._create_error_parser()
        # First set the flag via connection failure line
        parser._filter_socket_timeout_error_from_log(
            "[ERROR] HCCL(26252):the connection failure between this device",
            "2025-01-24-11:24:17.476916",
        )
        line = "   |  192.168.1.197(1)   |   192.168.1.198(0)   |  client  | no connect |"
        parser._filter_socket_timeout_error_from_log(line, "2025-01-24-11:24:17.476916")

        self.assertTrue(parser.socket_error_data_cache.exist_data_flag)
        self.assertEqual(len(parser.socket_error_data_cache.remote_info), 1)
        self.assertEqual(parser.socket_error_data_cache.remote_info[0].device_ip, "192.168.1.197")
        self.assertEqual(parser.socket_error_data_cache.remote_info[0].phy_device_id, "")

    def test_ipv4_time_out_extracts_remote_info(self):
        parser = self._create_error_parser()
        line = "   |  10.0.0.5(3)   |   10.0.0.6(2)   |  server  | time out |"
        parser._filter_socket_timeout_error_from_log(line, "2025-01-24-11:25:00.000000")

        self.assertEqual(len(parser.socket_error_data_cache.remote_info), 1)
        self.assertEqual(parser.socket_error_data_cache.remote_info[0].device_ip, "10.0.0.5")

    def test_ipv6_full_format_no_connect_extracts_remote_info(self):
        parser = self._create_error_parser()
        parser._filter_socket_timeout_error_from_log(
            "[ERROR] HCCL(26252):the connection failure between this device",
            "2025-01-24-11:24:17.476943",
        )
        line = (
            "   |  2001:0db8:85a3:0000:0000:8a2e:c0a8:01c5(1)"
            "   |   2001:0db8:85a3:0000:0000:8a2e:c0a8:01c6(0)   |  client  | no connect |"
        )
        parser._filter_socket_timeout_error_from_log(line, "2025-01-24-11:24:17.476943")

        self.assertTrue(parser.socket_error_data_cache.exist_data_flag)
        self.assertEqual(len(parser.socket_error_data_cache.remote_info), 1)
        self.assertEqual(
            parser.socket_error_data_cache.remote_info[0].device_ip,
            "2001:0db8:85a3:0000:0000:8a2e:c0a8:01c5",
        )

    def test_ipv6_compressed_format_no_connect_extracts_remote_info(self):
        parser = self._create_error_parser()
        parser._filter_socket_timeout_error_from_log(
            "[ERROR] HCCL(26252):the connection failure between this device",
            "2025-01-24-11:24:17.476943",
        )
        line = "   |  2001:db8::c0a8:01c5(1)   |   2001:db8::c0a8:01c6(0)   |  client  | no connect |"
        parser._filter_socket_timeout_error_from_log(line, "2025-01-24-11:24:17.476943")

        self.assertEqual(len(parser.socket_error_data_cache.remote_info), 1)
        self.assertEqual(parser.socket_error_data_cache.remote_info[0].device_ip, "2001:db8::c0a8:01c5")

    def test_ipv6_compressed_leading_double_colon_no_connect(self):
        parser = self._create_error_parser()
        parser._filter_socket_timeout_error_from_log(
            "[ERROR] HCCL(26252):the connection failure between this device",
            "2025-01-24-11:24:17.476943",
        )
        line = "   |  2001:db8:0:0:0:ff00:42:8329(1)   |   2001:db8:0:0:0:ff00:42:8330(0)   |  client  | no connect |"
        parser._filter_socket_timeout_error_from_log(line, "2025-01-24-11:24:17.476943")

        self.assertEqual(len(parser.socket_error_data_cache.remote_info), 1)
        self.assertEqual(parser.socket_error_data_cache.remote_info[0].device_ip, "2001:db8:0:0:0:ff00:42:8329")

    def test_ipv6_compressed_trailing_no_connect(self):
        parser = self._create_error_parser()
        parser._filter_socket_timeout_error_from_log(
            "[ERROR] HCCL(26252):the connection failure between this device",
            "2025-01-24-11:24:17.476943",
        )
        line = "   |  fe80::1(1)   |   fe80::2(0)   |  client  | no connect |"
        parser._filter_socket_timeout_error_from_log(line, "2025-01-24-11:24:17.476943")

        self.assertEqual(len(parser.socket_error_data_cache.remote_info), 1)
        self.assertEqual(parser.socket_error_data_cache.remote_info[0].device_ip, "fe80::1")

    def test_no_match_keyword_skipped(self):
        parser = self._create_error_parser()
        line = "   |  192.168.1.197(1)   |   192.168.1.198(0)   |  client  | connected |"
        parser._filter_socket_timeout_error_from_log(line, "2025-01-24-11:24:17.476943")

        self.assertFalse(parser.socket_error_data_cache.exist_data_flag)
        self.assertEqual(len(parser.socket_error_data_cache.remote_info), 0)

    def test_single_ip_match_early_return(self):
        parser = self._create_error_parser()
        line = "   |  192.168.1.197(1)   |   unknown(0)   |  client  | no connect |"
        parser._filter_socket_timeout_error_from_log(line, "2025-01-24-11:24:17.476943")

        # Only 1 match (not 2), should early return without adding remote_info
        self.assertEqual(len(parser.socket_error_data_cache.remote_info), 0)

    def test_connection_failure_line_sets_flag_only(self):
        parser = self._create_error_parser()
        line = "[ERROR] HCCL(26252):the connection failure between this device and target device may be due to..."
        parser._filter_socket_timeout_error_from_log(line, "2025-01-24-11:24:17.476943")

        self.assertTrue(parser.socket_error_data_cache.exist_data_flag)
        self.assertEqual(len(parser.socket_error_data_cache.remote_info), 0)

    def test_p2p_timeout_with_phy_id_adds_remote_info(self):
        parser = self._create_error_parser()
        line = "[ERROR] connected p2p timeout remote physic id:5"
        parser._filter_socket_timeout_error_from_log(line, "2025-01-24-11:24:17.476943")

        self.assertTrue(parser.socket_error_data_cache.exist_data_flag)
        self.assertEqual(len(parser.socket_error_data_cache.remote_info), 1)
        self.assertEqual(parser.socket_error_data_cache.remote_info[0].device_ip, "")
        self.assertEqual(parser.socket_error_data_cache.remote_info[0].phy_device_id, "5")
