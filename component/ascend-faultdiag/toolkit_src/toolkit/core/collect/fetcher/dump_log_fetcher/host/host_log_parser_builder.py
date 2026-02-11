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

from typing import List, Callable

from toolkit.core.collect.fetcher.dump_log_fetcher.host.host_dump_log_parser import \
    HostDumpLogParser, HostDumpLogParserV1, HostDumpLogParserV3, HostDumpLogParserV2
from toolkit.utils import file_tool


class ParserTypeConfig:

    def __init__(self, search_path: str, parser_type: Callable, up_level=1, judge_func: Callable[[str], bool] = None):
        # 该类型导出日志的唯一标志地址, 请谨慎选择不会重复的目录或者文件
        self.search_path = search_path
        self.parser_type = parser_type
        self.up_level = up_level
        # 判断函数, 入参为目录地址, 返回bool
        self.judge_func = judge_func


_PARSER_TYPE_CONFIGS = [
    ParserTypeConfig("ascend-dmi", HostDumpLogParserV1),
    ParserTypeConfig("npu_smi_log", HostDumpLogParserV2),
    ParserTypeConfig("hilink_down.log", HostDumpLogParserV3),
]


class HostLogParserBuilder:
    _DEFAULT_DEEP = 4

    @classmethod
    def build(cls, root_dir: str) -> List[HostDumpLogParser]:
        res = []
        for config in _PARSER_TYPE_CONFIGS:
            sub_dirs = file_tool.find_all_sub_paths(root_dir, config.search_path,
                                                    cls._DEFAULT_DEEP)
            for sub_dir in sub_dirs:
                res.append(config.parser_type(root_dir, file_tool.jump_up_directories(sub_dir, config.up_level)))
        return res
