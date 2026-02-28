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

from ascend_fd_tk.core.collect.collect_config import SwiCliOutputDataType
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.base import SwitchOutputParser
from ascend_fd_tk.utils import helpers


class SwitchDiagnoseInformationParser(SwitchOutputParser):
    _SPLIT_PATTERN = '=' * 79

    def __init__(self, file_content: str):
        super().__init__()
        self.file_content = file_content

    def parse(self) -> dict:
        self.find_name(self.file_content)
        self.find_ip(self.file_content)
        parts = helpers.split_str(self.file_content, self._SPLIT_PATTERN)
        if not parts or len(parts) % 2 != 0:
            return {}
        new_parts = [parts[i] + parts[i + 1] for i in range(0, len(parts), 2)]
        for part in new_parts:
            self.parse_cli_output_part(part)
        return self.parse_data.get_data_dict()

    # 找名字
    def find_name(self, file_content: str):
        search = helpers.find_pattern_after_substrings(
            file_content, ["display current-configuration", "sysname"], r".*"
        )
        if search:
            self.parse_data.add_data([SwiCliOutputDataType.SWI_NAME.name], search.group().strip())
