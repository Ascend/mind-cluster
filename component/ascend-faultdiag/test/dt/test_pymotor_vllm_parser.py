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
import tempfile
import os

from ascend_fd.pkg.parse.knowledge_graph.parser.pymotor_vllm_parser import PyMotorVLLMParser


class TestPyMotorVLLMParser(unittest.TestCase):
    """Test cases for PYMOTOR VLLM parser"""

    def _create_parser(self):
        return PyMotorVLLMParser({})

    def _create_temp_file(self, content, prefix="vllm-test"):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.log', prefix=prefix + "-", delete=False) as tmp_file:
            tmp_file.write(content)
            return tmp_file.name

    def test_parse_vllm_device_info(self):
        """Test parsing vllm device info from log line"""
        log_line = "2026-05-26 15:15:49,051 - INFO - device_id: 0, device_ip_info: ['ipaddr:10.0.245.11\\n', 'netmask:255.255.255.0\\n']"
        parser = self._create_parser()
        device_info = parser.find_device_info(log_line)

        self.assertIsNotNone(device_info)
        self.assertEqual(device_info.device_id, "0")
        self.assertEqual(device_info.device_ip, "10.0.245.11")

    def test_parse_motor_log_time(self):
        """Test parsing motor format timestamp"""
        log_line = "2026-05-26 15:16:20  [INFO][root][logger.py:170][proc:MainProcess]  Internal logs of pod"
        parser = self._create_parser()
        time_str = parser.get_log_time(log_line)

        self.assertTrue(time_str.startswith("2026-05-26 15:16:20"))

    def test_parse_vllm_log_time(self):
        """Test parsing vllm format timestamp"""
        log_line = "2026-05-26 15:15:48,667 - INFO - start /mnt/configmap/hccl_tools.py"
        parser = self._create_parser()
        time_str = parser.get_log_time(log_line)

        self.assertTrue(time_str.startswith("2026-05-26 15:15:48"))

    def test_extract_host_pod_ip(self):
        """Test extracting host_ip and pod_ip"""
        log_line_host = "2026-05-26 15:15:48,694 - INFO - host_ip: 90.90.97.42"
        log_line_pod = "2026-05-26 15:15:48,694 - INFO - pod_ip: 192.168.222.203"

        parser = self._create_parser()
        parser.extract_host_pod_ip(log_line_host)
        parser.extract_host_pod_ip(log_line_pod)

        self.assertEqual(parser.host_ip, "90.90.97.42")
        self.assertEqual(parser.pod_ip, "192.168.222.203")

    def test_invalid_ip_rejected(self):
        """Test that invalid IP addresses are rejected"""
        log_line = "2026-05-26 15:15:49,051 - INFO - device_id: 0, device_ip_info: ['ipaddr:invalid_ip\\n', 'netmask:255.255.255.0\\n']"
        parser = self._create_parser()
        device_info = parser.find_device_info(log_line)

        self.assertIsNone(device_info)

    def test_multiple_devices_parsed(self):
        """Test parsing multiple devices from log content"""
        log_content = """2026-05-26 15:15:49,051 - INFO - device_id: 0, device_ip_info: ['ipaddr:10.0.245.11\\n', 'netmask:255.255.255.0\\n']
2026-05-26 15:15:50,285 - INFO - device_id: 1, device_ip_info: ['ipaddr:10.0.244.11\\n', 'netmask:255.255.255.0\\n']
2026-05-26 15:15:51,175 - INFO - device_id: 10, device_ip_info: ['ipaddr:10.0.245.16\\n', 'netmask:255.255.255.0\\n']"""

        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)

        try:
            result = parser._parse_file(file_path)

            self.assertEqual(len(result.device_info_list), 3)
            device_ids = {d.device_id for d in result.device_info_list}
            self.assertEqual(device_ids, {"0", "1", "10"})
        finally:
            os.unlink(file_path)

    def test_container_ip_from_pod_ip(self):
        """Test that container IP is set from pod_ip"""
        log_content = """2026-05-26 15:15:48,694 - INFO - host_ip: 90.90.97.42
2026-05-26 15:15:48,694 - INFO - pod_ip: 192.168.222.203
2026-05-26 15:15:49,051 - INFO - device_id: 0, device_ip_info: ['ipaddr:10.0.245.11\\n', 'netmask:255.255.255.0\\n']"""

        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)

        try:
            result = parser._parse_file(file_path)

            self.assertEqual(result.container_ip, "192.168.222.203")
        finally:
            os.unlink(file_path)


class TestPyMotorVLLMParserIPv6(unittest.TestCase):
    """Test cases for PYMOTOR VLLM parser with IPv6 addresses"""

    def _create_parser(self):
        return PyMotorVLLMParser({})

    def _create_temp_file(self, content, prefix="vllm-"):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.log', prefix=prefix + "-", delete=False) as tmp_file:
            tmp_file.write(content)
            return tmp_file.name

    def test_find_device_info_ipv6(self):
        """Test parsing device info with IPv6 address"""
        log_line = (
            "2026-06-08 10:00:00,000 - INFO - "
            "device_id: 0, device_ip_info: ['ipaddr:fe80::1c01:3c0\\n', 'netmask:ffff:ffff::\\n']"
        )
        parser = self._create_parser()
        device_info = parser.find_device_info(log_line)

        self.assertIsNotNone(device_info)
        self.assertEqual(device_info.device_id, "0")
        self.assertEqual(device_info.device_ip, "fe80::1c01:3c0")

    def test_find_device_info_ipv6_full(self):
        """Test parsing device info with full-format IPv6 address"""
        log_line = (
            "2026-06-08 10:00:01,000 - INFO - "
            "device_id: 3, device_ip_info: ['ipaddr:2001:0db8:85a3:0000:0000:8a2e:0370:7334\\n', "
            "'netmask:ffff:ffff:ffff::\\n']"
        )
        parser = self._create_parser()
        device_info = parser.find_device_info(log_line)

        self.assertIsNotNone(device_info)
        self.assertEqual(device_info.device_id, "3")
        self.assertEqual(device_info.device_ip, "2001:0db8:85a3:0000:0000:8a2e:0370:7334")

    def test_extract_host_pod_ip_ipv6(self):
        """Test extracting host_ip and pod_ip with IPv6"""
        log_line_host = "2026-06-08 10:00:02,000 - INFO - host_ip: fe80::1c01:3c0"
        log_line_pod = "2026-06-08 10:00:02,000 - INFO - pod_ip: 2001:db8::1"

        parser = self._create_parser()
        parser.extract_host_pod_ip(log_line_host)
        parser.extract_host_pod_ip(log_line_pod)

        self.assertEqual(parser.host_ip, "fe80::1c01:3c0")
        self.assertEqual(parser.pod_ip, "2001:db8::1")

    def test_container_ip_from_pod_ip_ipv6(self):
        """Test that container IP is set from IPv6 pod_ip"""
        log_content = """2026-06-08 10:00:03,000 - INFO - host_ip: fe80::1c01:3c0
2026-06-08 10:00:03,000 - INFO - pod_ip: 2001:db8::1
2026-06-08 10:00:04,000 - INFO - device_id: 0, device_ip_info: ['ipaddr:fe80::1c01:3c0\\n', 'netmask:ffff:ffff::\\n']"""

        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)

        try:
            result = parser._parse_file(file_path)

            self.assertEqual(result.container_ip, "2001:db8::1")
        finally:
            os.unlink(file_path)

    def test_multiple_devices_parsed_ipv6(self):
        """Test parsing multiple devices with IPv6 addresses"""
        log_content = (
            "2026-06-08 10:00:05,000 - INFO - device_id: 0, device_ip_info: ['ipaddr:fe80::1c01:3c0\\n', 'netmask:ffff:ffff::\\n']\n"
            "2026-06-08 10:00:06,000 - INFO - device_id: 1, device_ip_info: ['ipaddr:2001:db8::2\\n', 'netmask:ffff:ffff::\\n']\n"
            "2026-06-08 10:00:07,000 - INFO - device_id: 2, device_ip_info: ['ipaddr:fd00::100\\n', 'netmask:ffff:ffff::\\n']\n"
        )

        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)

        try:
            result = parser._parse_file(file_path)

            self.assertEqual(len(result.device_info_list), 3)
            device_ids = {d.device_id for d in result.device_info_list}
            self.assertEqual(device_ids, {"0", "1", "2"})
            ips = {d.device_ip for d in result.device_info_list}
            self.assertIn("fe80::1c01:3c0", ips)
            self.assertIn("2001:db8::2", ips)
            self.assertIn("fd00::100", ips)
        finally:
            os.unlink(file_path)


class TestPyMotorVLLMParserMultiLineFaultMode(unittest.TestCase):
    """Regression tests for cross-line fault mode matching (all/opt/max_lines)"""

    CONFIG_PATH = os.path.join(
        os.path.dirname(__file__), '..', '..', 'src', 'ascend_fd', 'configuration', 'kg-config.json'
    )

    def _load_config(self):
        from ascend_fd.utils.load_kg_config import ParseRegexMap

        parse_regex = ParseRegexMap(config_pkg_list=[self.CONFIG_PATH]).get_parse_regex()
        return parse_regex.get(PyMotorVLLMParser.SOURCE_FILE, {})

    def _create_parser(self):
        config = self._load_config()
        return PyMotorVLLMParser(
            {
                "default_conf": {PyMotorVLLMParser.SOURCE_FILE: config},
                "user_conf": {},
            }
        )

    def _create_temp_file(self, content, prefix="vllm-mooncake"):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.log', prefix=prefix + "-", delete=False) as tmp_file:
            tmp_file.write(content)
            return tmp_file.name

    def test_mooncake_002_cross_line_match(self):
        """MOONCAKE_002 opt/all groups on different lines must match (regression)"""
        log_content = (
            "\x1b[93m2026-06-24 10:46:50.237365 WARNING  \x1b[0m\x1b[0K[3591] "
            "[client_pool.hpp:445] send request to 80.48.37.140:50060 failed. connection refused. \n"
            "E20260624 10:46:50.237413  3591 master_client.cpp:234] Client not available \n"
            "(Worker_TP1 pid=3193) ERROR 06-24 10:46:50 [mooncake_backend.py:146] "
            "[vllm-ascend] [distributed] - Initialize mooncake failed. ret=-600, "
            "metadata_server=P2PHANDSHAKE. Check mooncake config and network. \n"
        )
        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)
        try:
            result = parser._parse_file(file_path)
            event_codes = {ev.get("event_code") for ev in result.event_list}
            self.assertIn("AISW_VLLM_ASCEND_KV_POOL_MOONCAKE_002", event_codes)
        finally:
            os.unlink(file_path)

    def test_mooncake_002_no_match_when_group_missing(self):
        """MOONCAKE_002 must not match when one all-group is absent"""
        log_content = (
            "\x1b[93m2026-06-24 10:46:50.237365 WARNING  \x1b[0m\x1b[0K[3591] "
            "[client_pool.hpp:445] send request to 80.48.37.140:50060 failed. connection refused. \n"
            "E20260624 10:46:50.237413  3591 master_client.cpp:234] Client not available \n"
        )
        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)
        try:
            result = parser._parse_file(file_path)
            event_codes = {ev.get("event_code") for ev in result.event_list}
            self.assertNotIn("AISW_VLLM_ASCEND_KV_POOL_MOONCAKE_002", event_codes)
        finally:
            os.unlink(file_path)


if __name__ == '__main__':
    unittest.main()
