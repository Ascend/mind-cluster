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

import re
from typing import List

from ascend_fd.utils.fast_parser.py_ahocorasick import AhoCorasick


class ParseItem:

    def __init__(self, line_patterns, regex_pattern: str, data):
        self.line_patterns = line_patterns or []
        self.regex_pattern = regex_pattern and re.compile(regex_pattern)
        self.data = data


class MatchingRes:

    def __init__(self, parse_item: ParseItem, logline: str):
        self.parse_item = parse_item
        self.logline = logline


class FastParser:

    def __init__(self, parse_items: List[ParseItem]):
        self.searcher = self._build_searcher(parse_items)

    @staticmethod
    def _build_searcher(parse_items: List[ParseItem]) -> AhoCorasick:
        searcher = AhoCorasick()
        for item in parse_items:
            if item.line_patterns:
                searcher.add_pattern(item.line_patterns[0], item)
        searcher.build_failure()
        return searcher

    @staticmethod
    def _is_subsequence_ordered(patterns: list, line: str) -> bool:
        if not patterns:
            return True  # 空列表视为满足条件
        index = 0
        for pattern in patterns:
            index = line.find(pattern, index)
            if index == -1:
                return False
            index += len(pattern)  # 移动到当前元素之后
        return True

    def fast_parse_lines(self, log_lines: List[str]) -> List[MatchingRes]:
        res = []
        for line in log_lines:
            res.extend(self.fast_parse_line(line))
        return res

    def fast_parse_line(self, line) -> List[MatchingRes]:
        searches = self.searcher.search(line)
        if not searches:
            return []
        res = []
        for parse_item in searches:
            if self._is_subsequence_ordered(parse_item.line_patterns, line):
                res.append(MatchingRes(parse_item, line))
        return res
