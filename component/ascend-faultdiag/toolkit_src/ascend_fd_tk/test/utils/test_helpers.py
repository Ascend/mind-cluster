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

"""utils.helpers 通用辅助函数单元测试。"""

import unittest

from ascend_fd_tk.utils.helpers import (
    split_str,
    split_multiline,
    to_int,
    parse_hex,
    trans_date_fmt,
    to_float,
    find_pattern_after_substrings,
    camel_to_separated,
    mw_to_dbm,
    dbm_to_mw,
    strip_keys_from_parentheses,
    combine_32_to_64,
)


class TestSplit(unittest.TestCase):
    def test_split_str(self):
        data = "header\nSTART\na=1\nb=2\nSTART\nc=3\n"
        blocks = split_str(data, "START")
        self.assertEqual(len(blocks), 2)
        self.assertIn("a=1", blocks[0])
        self.assertIn("c=3", blocks[1])
        self.assertEqual(len(split_str(data, r"===", regex=True)), 0)
        self.assertEqual(split_str("plain text", "MARK"), [])

    def test_split_multiline(self):
        data = "aa\n+-\nb1\nb2\n+-\n+-\nc1\n+-\n+-\nd1\n"
        self.assertGreaterEqual(len(split_multiline(data, ["+-", "+-"])), 2)
        self.assertEqual(split_multiline("", ["+-"]), [])
        self.assertEqual(split_multiline("   ", ["+-"]), [])
        self.assertEqual(split_multiline("abc", []), ["abc"])


class TestNumberConversion(unittest.TestCase):
    def test_conversions(self):
        self.assertEqual(to_int("42"), 42)
        self.assertEqual(to_int(7), 7)
        self.assertEqual(to_int("abc", default=-1), -1)
        self.assertEqual(to_int("1.5"), 0)
        self.assertEqual(parse_hex("ff"), 255)
        self.assertEqual(parse_hex(" 1A "), 26)
        self.assertEqual(parse_hex("zz", default=9), 9)
        ok, val = to_float("12.5")
        self.assertTrue(ok)
        self.assertEqual(val, 12.5)
        self.assertFalse(to_float("")[0])
        self.assertFalse(to_float(123)[0])


class TestFormatConversion(unittest.TestCase):
    def test_date_and_pattern(self):
        self.assertEqual(trans_date_fmt("2026-01-02", "%Y-%m-%d", "%Y%m%d"), "20260102")
        self.assertEqual(trans_date_fmt("bad", "%Y-%m-%d", "%Y%m%d"), "")
        match = find_pattern_after_substrings("pre KEY1 mid KEY2 val=9 end", ["KEY1", "KEY2"], r"val=(\d+)")
        self.assertEqual(match.group(1), "9")
        self.assertIsNone(find_pattern_after_substrings("pre KEY1", ["KEY1", "MISSING"], r"\d+"))

    def test_camel_to_separated(self):
        self.assertEqual(camel_to_separated("HttpRequest"), "http_request")
        self.assertEqual(camel_to_separated("HTTPRequest"), "http_request")
        self.assertEqual(camel_to_separated(""), "")
        self.assertEqual(camel_to_separated("simple"), "simple")

    def test_mw_dbm(self):
        self.assertEqual(mw_to_dbm("1"), "0.0")
        self.assertEqual(mw_to_dbm("0"), "-40")
        self.assertEqual(mw_to_dbm("bad"), "-40")
        self.assertEqual(dbm_to_mw("10"), "10.0")
        self.assertEqual(dbm_to_mw("0"), "0.0001")
        self.assertEqual(dbm_to_mw("-999"), "0.0001")


class TestDictAndBitOps(unittest.TestCase):
    def test_dict_and_bits(self):
        src = {"device[a]": 1, "port": 2, 3: 3}
        result = strip_keys_from_parentheses(src)
        self.assertIn("device", result)
        self.assertEqual(result["port"], 2)
        self.assertNotIn(3, result)
        self.assertEqual(combine_32_to_64("1", "2"), str((1 << 32) | 2))
        self.assertEqual(combine_32_to_64("", "2"), "")
        self.assertEqual(combine_32_to_64("1", ""), "")


if __name__ == "__main__":
    unittest.main()
