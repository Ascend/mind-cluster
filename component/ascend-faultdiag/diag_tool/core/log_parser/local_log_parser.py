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

from typing import List, Dict

from diag_tool.core.log_parser.base import LogParser, LogParsePattern, FindResult
from diag_tool.core.log_parser.parse_script import process_pattern_group


class LocalLogParser(LogParser):
    """
    使用python脚本远程清洗
    """

    async def find(self, parse_dir="", log_pattern_map: Dict[str, LogParsePattern] = None) -> List[FindResult]:
        if not parse_dir or not log_pattern_map:
            return []
        pattern_group = []
        for v in log_pattern_map.values():
            pattern_group.append(
                (v.keyword_config.pattern_key, v.keyword_config.keyword_pattern, v.keyword_config.filepath_pattern)
            )
        parsed_list = process_pattern_group(parse_dir, pattern_group)
        result = [FindResult.from_dict(item) for item in parsed_list]
        result = self.fill_search_info(result, log_pattern_map)
        return result
