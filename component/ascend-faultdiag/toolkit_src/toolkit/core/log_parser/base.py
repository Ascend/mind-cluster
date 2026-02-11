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

import abc
import re
from typing import List, Dict

from toolkit.core.common.json_obj import JsonObj


class KeywordPattern(JsonObj):

    def __init__(self, pattern_key: str, keyword_pattern: str, filepath_pattern: str):
        self.pattern_key = pattern_key
        self.keyword_pattern = keyword_pattern
        self.filepath_pattern = filepath_pattern

    def __repr__(self):
        return f"{self.filepath_pattern}_{self.keyword_pattern}"


class LogParsePattern(JsonObj):

    def __init__(self, keyword_config: KeywordPattern, info_dict_pattern: str = ""):
        self.keyword_config = keyword_config
        self.info_dict_pattern = info_dict_pattern

    @classmethod
    def build(cls, pattern_key: str, keyword_pattern: str, filepath_pattern: str, info_dict_pattern: str = None):
        return cls(KeywordPattern(pattern_key, keyword_pattern, filepath_pattern), info_dict_pattern or "")


class FindResult(JsonObj):

    def __init__(self, pattern_key: str, logline: str, log_path: str, info_dict: dict = None):
        self.pattern_key = pattern_key
        self.logline = logline
        self.log_path = log_path
        self.info_dict = info_dict


class LogParser(abc.ABC):

    @staticmethod
    def fill_search_info(find_results: List[FindResult], log_pattern_map: Dict[str, LogParsePattern]) -> List[
        FindResult]:
        results = []
        for find_result in find_results:
            info_dict_pattern = log_pattern_map[find_result.pattern_key].info_dict_pattern
            if info_dict_pattern:
                search = re.search(info_dict_pattern, find_result.logline)
                find_result.info_dict = search and search.groupdict()
            results.append(find_result)
        return results

    @abc.abstractmethod
    async def find(self, parse_dir="", log_pattern_map: Dict[str, LogParsePattern] = None) -> List[FindResult]:
        pass
