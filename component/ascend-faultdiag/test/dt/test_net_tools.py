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

from ascend_fd.utils.net_tools import IPAddress


class TestIPAddressIPv4(unittest.TestCase):
    def test_valid_ipv4(self):
        self.assertTrue(IPAddress.is_ipv4("192.168.1.1"))
        self.assertTrue(IPAddress.is_ipv4("10.0.0.1"))
        self.assertTrue(IPAddress.is_ipv4("172.16.0.1"))
        self.assertTrue(IPAddress.is_ipv4("255.255.255.255"))
        self.assertTrue(IPAddress.is_ipv4("1.2.3.4"))

    def test_zero_ipv4(self):
        self.assertTrue(IPAddress.is_ipv4("0.0.0.0"))
        self.assertTrue(IPAddress.is_ipv4("0.0.0.1"))
        self.assertTrue(IPAddress.is_ipv4("127.0.0.1"))

    def test_empty_or_none(self):
        self.assertFalse(IPAddress.is_ipv4(""))

    def test_octet_out_of_range(self):
        self.assertFalse(IPAddress.is_ipv4("256.1.1.1"))
        self.assertFalse(IPAddress.is_ipv4("1.256.1.1"))
        self.assertFalse(IPAddress.is_ipv4("1.1.256.1"))
        self.assertFalse(IPAddress.is_ipv4("1.1.1.256"))
        self.assertFalse(IPAddress.is_ipv4("999.999.999.999"))

    def test_incomplete_ipv4(self):
        self.assertFalse(IPAddress.is_ipv4("192.168.1"))
        self.assertFalse(IPAddress.is_ipv4("192.168"))
        self.assertFalse(IPAddress.is_ipv4("192"))

    def test_extra_octets(self):
        self.assertFalse(IPAddress.is_ipv4("192.168.1.1.1"))

    def test_non_numeric(self):
        self.assertFalse(IPAddress.is_ipv4("abc.def.ghi.jkl"))
        self.assertFalse(IPAddress.is_ipv4("192.168.1.x"))

    def test_ipv6_not_ipv4(self):
        self.assertFalse(IPAddress.is_ipv4("2001:db8::1"))
        self.assertFalse(IPAddress.is_ipv4("::1"))
        self.assertFalse(IPAddress.is_ipv4("fe80::1"))

    def test_hostname_not_ipv4(self):
        self.assertFalse(IPAddress.is_ipv4("localhost"))


class TestIPAddressIPv6(unittest.TestCase):
    def test_valid_full_ipv6(self):
        self.assertTrue(IPAddress.is_ipv6("2001:0db8:85a3:0000:0000:8a2e:0370:7334"))
        self.assertTrue(IPAddress.is_ipv6("2001:db8:85a3:0:0:8a2e:370:7334"))

    def test_valid_compressed_ipv6(self):
        self.assertTrue(IPAddress.is_ipv6("2001:db8::1"))
        self.assertTrue(IPAddress.is_ipv6("::1"))
        self.assertTrue(IPAddress.is_ipv6("::"))
        self.assertTrue(IPAddress.is_ipv6("fe80::1"))
        self.assertTrue(IPAddress.is_ipv6("2001:db8:85a3::8a2e:370:7334"))

    def test_empty_string(self):
        self.assertFalse(IPAddress.is_ipv6(""))

    def test_ipv4_not_ipv6(self):
        self.assertFalse(IPAddress.is_ipv6("192.168.1.1"))
        self.assertFalse(IPAddress.is_ipv6("10.0.0.1"))

    def test_invalid_ipv6(self):
        self.assertFalse(IPAddress.is_ipv6("gggg::1"))
        self.assertFalse(IPAddress.is_ipv6("12345::1"))


class TestIPAddressIsValid(unittest.TestCase):
    def test_ipv4_is_valid(self):
        self.assertTrue(IPAddress.is_valid_ip("192.168.1.1"))
        self.assertTrue(IPAddress.is_valid_ip("10.0.0.1"))

    def test_ipv6_is_valid(self):
        self.assertTrue(IPAddress.is_valid_ip("2001:db8::1"))
        self.assertTrue(IPAddress.is_valid_ip("::1"))

    def test_invalid_is_not_valid(self):
        self.assertFalse(IPAddress.is_valid_ip(""))
        self.assertFalse(IPAddress.is_valid_ip("invalid"))
        self.assertFalse(IPAddress.is_valid_ip("256.256.256.256"))

    def test_zero_ipv4_is_valid(self):
        self.assertTrue(IPAddress.is_valid_ip("0.0.0.0"))
