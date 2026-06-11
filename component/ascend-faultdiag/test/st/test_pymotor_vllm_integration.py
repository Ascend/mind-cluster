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

from ascend_fd.pkg.parse.knowledge_graph.parser.pymotor_vllm_parser import PyMotorVLLMParser


class TestPyMotorVLLMIntegration(unittest.TestCase):
    """Integration tests for PYMOTOR VLLM parser with real log data"""

    @classmethod
    def setUpClass(cls):
        cls.test_data_dir = os.path.join(
            os.path.dirname(__file__), '..', 'st_module_testcase', 'kg_parse', 'pymotor_vllm'
        )
        cls.log_file = os.path.join(cls.test_data_dir, 'vllm-p0-cbf7496f5-qc255_node-97-40.log')

    def setUp(self):
        self.parser = PyMotorVLLMParser({})

    def test_log_file_exists(self):
        """Test that the test log file exists"""
        self.assertTrue(os.path.exists(self.log_file), f"Test log file not found: {self.log_file}")

    def test_parse_real_log_file(self):
        """Test parsing a real PYMOTOR VLLM log file"""
        if not os.path.exists(self.log_file):
            self.skipTest(f"Test log file not found: {self.log_file}")

        result = self.parser._parse_file(self.log_file)

        self.assertIsNotNone(result)
        self.assertIsInstance(result.device_info_list, list)
        self.assertGreater(len(result.device_info_list), 0, "Should parse at least one device from log file")

    def test_extract_multiple_devices(self):
        """Test that multiple devices are extracted from real log"""
        if not os.path.exists(self.log_file):
            self.skipTest(f"Test log file not found: {self.log_file}")

        result = self.parser._parse_file(self.log_file)

        device_ids = {d.device_id for d in result.device_info_list}
        expected_ids = {"0", "1", "10", "11", "12"}

        self.assertTrue(expected_ids.issubset(device_ids), f"Expected devices {expected_ids}, but got {device_ids}")

    def test_container_ip_extraction(self):
        """Test that container IP is correctly extracted"""
        if not os.path.exists(self.log_file):
            self.skipTest(f"Test log file not found: {self.log_file}")

        result = self.parser._parse_file(self.log_file)

        self.assertEqual(result.container_ip, "192.168.222.203", "Container IP should be set from pod_ip")

    def test_device_ip_validation(self):
        """Test that all extracted device IPs are valid"""
        if not os.path.exists(self.log_file):
            self.skipTest(f"Test log file not found: {self.log_file}")

        result = self.parser._parse_file(self.log_file)

        for device_info in result.device_info_list:
            self.assertIsNotNone(device_info.device_ip, f"Device {device_info.device_id} should have IP")
            self.assertTrue(
                device_info.device_ip.startswith("10.0."), f"Device {device_info.device_id} IP should start with 10.0."
            )

    def test_event_generation(self):
        """Test that events are generated from log parsing"""
        if not os.path.exists(self.log_file):
            self.skipTest(f"Test log file not found: {self.log_file}")

        result = self.parser._parse_file(self.log_file)

        self.assertIsInstance(result.event_list, list)
        # Note: Event generation depends on fault pattern matching
        # This test verifies the parser doesn't crash during event generation


if __name__ == '__main__':
    unittest.main()
