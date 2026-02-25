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
import asyncio
import collections
import os.path
import re
from typing import List, Dict, Tuple

from toolkit.core.collect.collect_config import ToolLogCollectionDataType, ToolLogCollectionSplitType
from toolkit.core.collect.fetcher.dump_log_fetcher.base import DumpLogDirParser
from toolkit.core.collect.fetcher.dump_log_fetcher.cli_output_parsed_data import CliOutputParsedData
from toolkit.core.collect.fetcher.dump_log_fetcher.host.dump_log_file_parse_config import ParseConfig, \
    ParseConfigCollectionV1, ParseConfigCollectionV2, ParseConfigCollectionV3
from toolkit.core.log_parser.base import FindResult
from toolkit.core.log_parser.local_log_parser import LocalLogParser
from toolkit.core.log_parser.parse_config import msnpureport_log_config


class HostDumpLogParser(DumpLogDirParser, abc.ABC):
    _TITLE_TEMPLATE = r'([=-]){3,20}\s*{title}\s*\1{3,20}'
    _END_SIGNS = ["----------", "======"]
    _TIME_FMT_PATTERN = re.compile(r"\d{4}(-\d{1,2}){5}")

    def __init__(self, root_dir: str, parse_dir: str):
        super().__init__(root_dir, parse_dir)

    # 返回标准字典, 便于作为多进程子任务
    def parse(self) -> dict:
        parse_data = CliOutputParsedData()
        path_cache: Dict[str, List[ParseConfig]] = collections.defaultdict(list)
        for config in self.get_parse_configs():
            path_cache[config.file_path].append(config)
        for path, configs in path_cache.items():
            log_file_path = os.path.join(self.parse_dir, path)
            if not os.path.exists(log_file_path):
                continue
            with open(log_file_path, "r", encoding="utf-8") as f:
                file_lines = f.readlines()
            for config in configs:
                part_lines = self._find_part_by_title(config.title, file_lines)
                split_parts = self._split_lines(part_lines, config.split_type)
                for part in split_parts:
                    keys = [config.data_type.name] + part[0]
                    parse_data.add_data(keys, part[1])
        parse_data.add_data([ToolLogCollectionDataType.MS_NPU_REPORT.name], self._parse_msnpureport_log())
        if not parse_data.get_data_dict():
            return {}
        self._find_log_name(parse_data)
        return parse_data.get_data_dict()

    @abc.abstractmethod
    def get_parse_configs(self) -> List[ParseConfig]:
        return []

    @abc.abstractmethod
    def get_pattern_map(self) -> Dict:
        return {}

    @abc.abstractmethod
    def get_name_dir(self) -> str:
        pass

    def _find_log_name(self, parse_data: CliOutputParsedData):
        if parse_data.fetch_data_by_name(ToolLogCollectionDataType.HOST_ID.name):
            return
        log_dir = self.get_name_dir()
        root_dir = os.path.abspath(self.root_dir)
        if os.path.exists(log_dir) and log_dir != root_dir:
            relpath = os.path.relpath(log_dir, root_dir).replace(os.sep, "_")
            parse_data.add_data([ToolLogCollectionDataType.HOST_ID.name], relpath)

    def _find_part_by_title(self, title: str, file_lines: List[str]) -> List[str]:
        pattern = self._TITLE_TEMPLATE.replace("{title}", title)
        res_list = []
        found_flag = False
        for file_line in file_lines:
            if not found_flag and re.search(pattern, file_line):
                found_flag = True
            elif found_flag and any(sign in file_line for sign in self._END_SIGNS):
                break
            elif found_flag and file_line.strip():
                res_list.append(file_line.strip())
        return res_list

    def _split_lines(self, file_lines: List[str],
                     split_type: ToolLogCollectionSplitType) -> List[Tuple[List[str], str]]:
        pattern = self.get_pattern_map().get(split_type)
        # 不按device_id分割场景, 直接返回全部行, 作为数组首位
        if not pattern:
            return [([], "\n".join(file_lines))]
        res_list = []
        temp_list = []
        last_search_res = None
        for line in file_lines:
            search = pattern.search(line)
            if search:
                if temp_list and last_search_res:
                    res_list.append((last_search_res, "\n".join(temp_list)))
                temp_list = []
                last_search_res = list(search.groups())
            elif line.strip():
                temp_list.append(line.strip())
        if last_search_res:
            res_list.append((last_search_res, "\n".join(temp_list)))
        return res_list

    def _parse_msnpureport_log(self) -> List[FindResult]:
        pattern_map = {}
        for config in msnpureport_log_config.MS_NPU_REPORT_PARSE_CONFIG:
            pattern_map[config.keyword_config.pattern_key] = config
        result = []
        for file in os.listdir(self.parse_dir):
            if not self._TIME_FMT_PATTERN.match(file.strip()):
                continue
            log_dir = os.path.join(self.parse_dir, file)
            if not os.path.isdir(log_dir):
                continue
            result.extend(asyncio.run(LocalLogParser().find(log_dir, pattern_map)))
        return result


class HostDumpLogParserV1(HostDumpLogParser):
    _PATTERN_MAP = {
        ToolLogCollectionSplitType.DEVICE_ID: re.compile(r"hccn_tool -i (\d{1,2})"),
        ToolLogCollectionSplitType.DEVICE_CHIP_ID: re.compile(r"npu-smi info .{0,30}-i (\d{1,2}) -c (\d{1,2})")
    }

    def get_name_dir(self) -> str:
        return os.path.abspath(self.parse_dir)

    def get_parse_configs(self) -> List[ParseConfig]:
        return ParseConfigCollectionV1.get_configs()

    def get_pattern_map(self) -> Dict:
        return self._PATTERN_MAP


class HostDumpLogParserV2(HostDumpLogParser):
    _PATTERN_MAP = {
        ToolLogCollectionSplitType.DEVICE_ID: re.compile(r"device_id=(\d{1,2})"),
        ToolLogCollectionSplitType.DEVICE_CHIP_ID: re.compile(r"device_id=(\d{1,2})_c(\d{1,2})")
    }

    def get_name_dir(self) -> str:
        return os.path.abspath(os.path.dirname(self.parse_dir))

    def get_parse_configs(self) -> List[ParseConfig]:
        return ParseConfigCollectionV2.get_configs()

    def get_pattern_map(self) -> Dict:
        return self._PATTERN_MAP


class HostDumpLogParserV3(HostDumpLogParser):
    _PATTERN_MAP = {
        ToolLogCollectionSplitType.DEVICE_ID: re.compile(r"====>\s*(\d{1,2})")
    }

    def get_name_dir(self) -> str:
        return os.path.abspath(self.parse_dir)

    def get_parse_configs(self) -> List[ParseConfig]:
        return ParseConfigCollectionV3.get_configs()

    def get_pattern_map(self) -> Dict:
        return self._PATTERN_MAP
