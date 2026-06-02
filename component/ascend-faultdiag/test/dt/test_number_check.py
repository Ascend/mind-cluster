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

from ascend_fd.utils.number_check import NumberCheck


class TestNumberCheck(unittest.TestCase):
    """Test cases for NumberCheck utility class"""

    def test_is_non_negative_integer_valid_cases(self):
        """Test is_non_negative_integer with valid non-negative integers (including zero)"""
        valid_values = ["0", "1", "5", "100", "999999", "+10"]
        for value in valid_values:
            with self.subTest(value=value):
                self.assertTrue(NumberCheck.is_non_negative_integer(value), f"{value} should be a non-negative integer")

    def test_is_non_negative_integer_valid_zero(self):
        """Test is_non_negative_integer accepts zero as a valid value"""
        self.assertTrue(NumberCheck.is_non_negative_integer("0"), "Zero should be considered a non-negative integer")

    def test_is_non_negative_integer_invalid_negative(self):
        """Test is_non_negative_integer rejects negative numbers"""
        invalid_negatives = ["-1", "-100", "-999"]
        for value in invalid_negatives:
            with self.subTest(value=value):
                self.assertFalse(NumberCheck.is_non_negative_integer(value), f"{value} should not be accepted")

    def test_is_non_negative_integer_empty_and_none(self):
        """Test is_non_negative_integer handles empty and None values"""
        self.assertFalse(NumberCheck.is_non_negative_integer(""))
        self.assertFalse(NumberCheck.is_non_negative_integer(None))

    def test_is_non_negative_integer_non_numeric(self):
        """Test is_non_negative_integer rejects non-numeric strings"""
        non_numeric = ["abc", "12.34", "1.5", "", " ", "NaN", "inf"]
        for value in non_numeric:
            with self.subTest(value=value):
                self.assertFalse(NumberCheck.is_non_negative_integer(value), f"'{value}' should not be accepted")

    def test_is_non_negative_integer_with_whitespace(self):
        """Test is_non_negative_integer strips whitespace from numeric strings"""
        self.assertTrue(NumberCheck.is_non_negative_integer(" 123"), "Should accept leading whitespace")
        self.assertTrue(NumberCheck.is_non_negative_integer("123 "), "Should accept trailing whitespace")
        self.assertTrue(NumberCheck.is_non_negative_integer(" 123 "), "Should accept surrounding whitespace")
        self.assertFalse(NumberCheck.is_non_negative_integer("   "), "Should reject whitespace-only string")


if __name__ == "__main__":
    unittest.main()
