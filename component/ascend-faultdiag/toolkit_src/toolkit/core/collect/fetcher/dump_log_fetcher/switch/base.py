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
from typing import List

from toolkit.core.collect.collect_config import SwiCliOutputDataType
from toolkit.core.collect.fetcher.dump_log_fetcher.cli_output_parsed_data import CliOutputParsedData
from toolkit.core.collect.fetcher.dump_log_fetcher.switch.parse_config import PARSE_CONFIGS, \
    SwiCliOutputParseConfig
from toolkit.utils import helpers


class SwitchOutputParser(abc.ABC):

    _IP_PATTERN = r"(\d{1,3}(\.\d{1,3}){3})"

    def __init__(self):
        self.parse_data = CliOutputParsedData()

    def parse_cli_output_part(self, content: str):
        possible_configs: List[SwiCliOutputParseConfig] = []
        for config in PARSE_CONFIGS:
            if all(key in content for key in config.primary_keys):
                possible_configs.append(config)
                if not config.multi_parts_judge_func:
                    break
        if len(possible_configs) == 1:
            self.parse_data.add_data([possible_configs[0].data_type.name], content)
        elif len(possible_configs) > 1:
            for config in possible_configs:
                if config.multi_parts_judge_func and config.multi_parts_judge_func(content):
                    self.parse_data.add_data([config.data_type.name], content)
                    break

    # 找ip
    def find_ip(self, file_content: str):

        search = helpers.find_pattern_after_substrings(file_content, ["interface MEth0/0/0", "ip address"],
                                                       self._IP_PATTERN)
        if search:
            self.parse_data.add_data([SwiCliOutputDataType.SWI_IP.name], search.group(1))
