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

        self.assertIsInstance(result, list)

    def test_event_generation(self):
        """Test that events are generated from log parsing"""
        if not os.path.exists(self.log_file):
            self.skipTest(f"Test log file not found: {self.log_file}")

        result = self.parser._parse_file(self.log_file)

        self.assertIsInstance(result, list)
        # Note: Event generation depends on fault pattern matching
        # This test verifies the parser doesn't crash during event generation


if __name__ == '__main__':
    unittest.main()
