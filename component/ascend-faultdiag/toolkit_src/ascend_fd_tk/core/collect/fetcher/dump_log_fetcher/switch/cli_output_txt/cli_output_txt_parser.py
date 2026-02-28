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

import re

from ascend_fd_tk.core.collect.collect_config import SwiCliOutputDataType
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.base import SwitchOutputParser


class SwiCliOutputTxtParser(SwitchOutputParser):
    _NAME_PATTERN = re.compile(r"<([^>]+)>")
    _SPLIT_PATTERN = re.compile(r"[\[<]")

    def __init__(self, file_content: str):
        super().__init__()
        self.file_content = file_content

    def parse(self) -> dict:
        self.find_ip(self.file_content)
        # 找不到ip视为没有dis cur分不清内容, 直接跳过
        if not self.parse_data.get_data_dict():
            return {}
        self.find_name(self.file_content)
        self.file_content = self.file_content.replace("[ PORT SNR ]", "PORT SNR ]")
        parts = self._SPLIT_PATTERN.split(self.file_content)
        for part in parts:
            self.parse_cli_output_part(part)
        return self.parse_data.get_data_dict()

    # 找名字
    def find_name(self, file_content: str):
        search = self._NAME_PATTERN.search(file_content)
        if search:
            self.parse_data.add_data([SwiCliOutputDataType.SWI_NAME.name], search.group(1))
