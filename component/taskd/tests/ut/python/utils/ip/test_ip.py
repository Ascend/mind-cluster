#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

import unittest

from taskd.python.utils.ip import pre_handle_ip


class TestPreHandleIp(unittest.TestCase):
    """
    Test pre_handle_ip function
    """

    def setUp(self):
        super().setUp()

    def tearDown(self):
        super().tearDown()

    def test_should_return_unchanged_when_address_already_has_brackets(self):
        test_cases = [
            "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
            "[::1]",
            "[2001:db8::1]",
            "[fe80::1]",
        ]
        for address in test_cases:
            with self.subTest(address=address):
                result, err = pre_handle_ip(address)
                self.assertIsNone(err, f"Expected no error for {address}")
                self.assertEqual(result, address,
                                 f"Expected {address} to be unchanged")

    def test_should_add_brackets_when_address_is_valid_ipv6(self):
        test_cases = [
            ("2001:0db8:85a3:0000:0000:8a2e:0370:7334", "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]"),
            ("2001:db8:85a3::8a2e:370:7334", "[2001:db8:85a3::8a2e:370:7334]"),
            ("2001:db8:85a3:0:0:8a2e:370:7334", "[2001:db8:85a3:0:0:8a2e:370:7334]"),
            ("::1", "[::1]"),
            ("2001:db8::1", "[2001:db8::1]"),
            ("fe80::1", "[fe80::1]"),
            ("::", "[::]"),
            ("2001:0db8:0000:0000:0000:0000:0000:0001", "[2001:0db8:0000:0000:0000:0000:0000:0001]"),
        ]
        for address, expected in test_cases:
            with self.subTest(address=address):
                result, err = pre_handle_ip(address)
                self.assertIsNone(err, f"Expected no error for {address}")
                self.assertEqual(result, expected,
                                 f"Expected {address} to be wrapped with brackets")

    def test_should_return_unchanged_when_address_is_valid_ipv4(self):
        test_cases = [
            "192.168.1.1",
            "10.0.0.1",
            "127.0.0.1",
            "0.0.0.0",
            "255.255.255.255",
        ]
        for address in test_cases:
            with self.subTest(address=address):
                result, err = pre_handle_ip(address)
                self.assertIsNone(err, f"Expected no error for {address}")
                self.assertEqual(result, address,
                                 f"Expected {address} to be unchanged")

    def test_should_return_error_when_address_is_empty_string(self):
        result, err = pre_handle_ip("")
        self.assertIsNotNone(err, "Expected error for empty string")
        self.assertEqual(result, "")

    def test_should_return_error_when_address_is_invalid_format(self):
        test_cases = [
            "not an ip address",
            "192.168.1",
            "2001:db8:85a3::8a2e:370:7334:extra",
            "2001::db8::1",
            "2001:db8:85a3:0:0:8a2e:370:7334:extra",
            "256.256.256.256",
            "192.168.1.1.1",
        ]
        for address in test_cases:
            with self.subTest(address=address):
                result, err = pre_handle_ip(address)
                self.assertIsNotNone(err, f"Expected error for invalid address: {address}")
                self.assertEqual(result, "")


if __name__ == "__main__":
    unittest.main()
