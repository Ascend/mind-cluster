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
import logging
from unittest.mock import MagicMock

from ascend_fd.utils.comm_valid import process_device_id


class TestProcessDeviceId(unittest.TestCase):
    """Test cases for process_device_id utility function"""

    def setUp(self):
        """Set up test fixtures"""
        self.mock_logger = MagicMock(spec=logging.Logger)
        self.test_line = "device_id: 0, some log content"

    def test_valid_non_negative_integer_zero(self):
        """Test that zero is accepted as a valid device ID"""
        result = process_device_id("0", self.test_line, "devPhyId", -1, self.mock_logger)
        self.assertEqual(result, "0")
        self.mock_logger.warning.assert_not_called()

    def test_valid_positive_integer(self):
        """Test that positive integers are accepted as valid device IDs"""
        test_values = ["1", "5", "100", "999999"]
        for value in test_values:
            with self.subTest(value=value):
                result = process_device_id(value, self.test_line, "deviceLogicId", "-1", self.mock_logger)
                self.assertEqual(result, value)
                self.mock_logger.warning.assert_not_called()

    def test_empty_value_returns_empty_string(self):
        """Test that empty string returns empty string without warning"""
        result = process_device_id("", self.test_line, "devPhyId", -1, self.mock_logger)
        self.assertEqual(result, "")
        self.mock_logger.warning.assert_not_called()

    def test_none_value_returns_empty_string(self):
        """Test that None value returns empty string without warning"""
        result = process_device_id(None, self.test_line, "devPhyId", -1, self.mock_logger)
        self.assertEqual(result, "")
        self.mock_logger.warning.assert_not_called()

    def test_invalid_value_negative_one_integer(self):
        """Test that integer -1 (NEGATIVE_ONE) returns empty string"""
        result = process_device_id(-1, self.test_line, "devPhyId", -1, self.mock_logger)
        self.assertEqual(result, "")
        self.mock_logger.warning.assert_not_called()

    def test_invalid_value_negative_one_string(self):
        """Test that string '-1' (INVALID_ID) returns empty string"""
        result = process_device_id("-1", self.test_line, "deviceLogicId", "-1", self.mock_logger)
        self.assertEqual(result, "")
        self.mock_logger.warning.assert_not_called()

    def test_invalid_value_custom_string(self):
        """Test custom invalid value returns empty string"""
        result = process_device_id("INVALID", self.test_line, "devPhyId", "INVALID", self.mock_logger)
        self.assertEqual(result, "")
        self.mock_logger.warning.assert_not_called()

    def test_non_numeric_string_triggers_warning(self):
        """Test that non-numeric strings trigger warning but still return value"""
        non_numeric_values = ["dadada", "abc", "12.34", "xyz123"]
        for value in non_numeric_values:
            with self.subTest(value=value):
                self.mock_logger.reset_mock()
                result = process_device_id(value, self.test_line, "devPhyId", -1, self.mock_logger)
                self.assertEqual(result, value)
                self.mock_logger.warning.assert_called_once()
                call_args = self.mock_logger.warning.call_args[0]
                self.assertEqual(call_args[0], "Except %s non-negative integer, but got: %s, origin line: %s")
                self.assertEqual(call_args[1], "devPhyId")
                self.assertEqual(call_args[2], value)

    def test_negative_number_triggers_warning(self):
        """Test that negative numbers trigger warning but still return value"""
        negative_values = ["-100", "-999"]
        for value in negative_values:
            with self.subTest(value=value):
                self.mock_logger.reset_mock()
                result = process_device_id(value, self.test_line, "deviceLogicId", "-1", self.mock_logger)
                self.assertEqual(result, value)
                self.mock_logger.warning.assert_called_once()
                call_args = self.mock_logger.warning.call_args[0]
                self.assertEqual(call_args[0], "Except %s non-negative integer, but got: %s, origin line: %s")
                self.assertEqual(call_args[1], "deviceLogicId")
                self.assertEqual(call_args[2], value)

    def test_warning_message_format(self):
        """Test that warning message contains correct information"""
        test_value = "dadada"
        id_name = "phydevId"
        line_content = "some error log line here"

        process_device_id(test_value, line_content, id_name, -1, self.mock_logger)

        self.mock_logger.warning.assert_called_once()
        args, kwargs = self.mock_logger.warning.call_args

        self.assertEqual(args[0], "Except %s non-negative integer, but got: %s, origin line: %s")
        self.assertEqual(args[1], id_name)
        self.assertEqual(args[2], test_value)
        self.assertEqual(args[3], line_content.strip())

    def test_with_whitespace_in_value(self):
        """Test handling of values with whitespace"""
        result = process_device_id(" 123 ", self.test_line, "devPhyId", -1, self.mock_logger)
        self.assertEqual(result, " 123 ")
        self.mock_logger.warning.assert_not_called()

    def test_different_loggers(self):
        """Test that function works with different logger instances"""
        mock_kg_logger = MagicMock(spec=logging.Logger)
        mock_rc_logger = MagicMock(spec=logging.Logger)

        process_device_id("dadada", self.test_line, "devPhyId", -1, mock_kg_logger)
        process_device_id("abc", self.test_line, "deviceLogicId", "-1", mock_rc_logger)

        mock_kg_logger.warning.assert_called_once()
        mock_rc_logger.warning.assert_called_once()

        kg_call_args = mock_kg_logger.warning.call_args[0]
        rc_call_args = mock_rc_logger.warning.call_args[0]

        self.assertEqual(kg_call_args[1], "devPhyId")
        self.assertEqual(rc_call_args[1], "deviceLogicId")

    def test_long_log_line_handling(self):
        """Test handling of long log lines in warning message"""
        long_line = "x" * 10000
        process_device_id("invalid", long_line, "devPhyId", -1, self.mock_logger)

        self.mock_logger.warning.assert_called_once()
        args, kwargs = self.mock_logger.warning.call_args
        self.assertEqual(args[3], long_line.strip())

    def test_special_characters_in_value(self):
        """Test handling of special characters in device ID value"""
        special_values = ["@#$%", "!@#", "test-123"]
        for value in special_values:
            with self.subTest(value=value):
                self.mock_logger.reset_mock()
                result = process_device_id(value, self.test_line, "devPhyId", -1, self.mock_logger)
                self.assertEqual(result, value)
                self.mock_logger.warning.assert_called_once()


class TestProcessDeviceIdEdgeCases(unittest.TestCase):
    """Edge case tests for process_device_id"""

    def setUp(self):
        self.mock_logger = MagicMock(spec=logging.Logger)

    def test_unicode_characters(self):
        """Test handling of unicode characters"""
        result = process_device_id("中文测试", "test line", "devPhyId", -1, self.mock_logger)
        self.assertEqual(result, "中文测试")
        self.mock_logger.warning.assert_called_once()

    def test_very_long_number_string(self):
        """Test very long numeric string"""
        long_number = "9" * 100
        result = process_device_id(long_number, "test", "devPhyId", -1, self.mock_logger)
        self.assertEqual(result, long_number)
        self.mock_logger.warning.assert_not_called()

    def test_float_like_strings(self):
        """Test float-like strings should trigger warning"""
        float_values = ["3.14", "0.0", "-1.5", ".5"]
        for value in float_values:
            with self.subTest(value=value):
                self.mock_logger.reset_mock()
                result = process_device_id(value, "test", "devPhyId", -1, self.mock_logger)
                self.assertEqual(result, value)
                self.mock_logger.warning.assert_called_once()

    def test_scientific_notation(self):
        """Test scientific notation strings should trigger warning"""
        sci_values = ["1e10", "1E5", "2.5e3"]
        for value in sci_values:
            with self.subTest(value=value):
                self.mock_logger.reset_mock()
                result = process_device_id(value, "test", "devPhyId", -1, self.mock_logger)
                self.assertEqual(result, value)
                self.mock_logger.warning.assert_called_once()

    def test_boolean_values(self):
        """Test boolean values behavior"""
        for value in [True, False]:
            with self.subTest(value=value):
                self.mock_logger.reset_mock()
                result = process_device_id(value, "test", "devPhyId", -1, self.mock_logger)
                if value:
                    self.assertEqual(result, value)
                    self.mock_logger.warning.assert_called_once()
                else:
                    self.assertEqual(result, "")


if __name__ == "__main__":
    unittest.main()
