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

from ascend_fd.utils.fast_parser.fast_parser import FastParser, ParseItem


class TestFastParser(unittest.TestCase):

    def test(self):
        parser = FastParser([ParseItem(["test"], "", "1")])
        res = parser.fast_parse_lines(["my test line"])
        self.assertEqual(res[0].parse_item.data, "1")
