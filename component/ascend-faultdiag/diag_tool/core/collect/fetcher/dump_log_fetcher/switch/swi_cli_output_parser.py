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

import os

from diag_tool.core.collect.fetcher.dump_log_fetcher.switch.cli_output_txt.cli_output_txt_parser import \
    SwiCliOutputTxtParser
from diag_tool.core.collect.fetcher.dump_log_fetcher.switch.diag_info_output.diag_info_output_parser import \
    SwitchDiagnoseInformationParser


class SwiCliOutputParser:
    _DIAG_INFO_START = "======"

    @classmethod
    def parse(cls, file_path: str) -> dict:
        if not os.path.exists(file_path):
            return {}
        try:
            with open(file_path, 'r', encoding="utf8") as f:
                content = f.read()
        except UnicodeDecodeError:
            return {}
        if not content:
            return {}
        if content.startswith(cls._DIAG_INFO_START):
            data = SwitchDiagnoseInformationParser(content).parse()
        else:
            data = SwiCliOutputTxtParser(content).parse()
        return data
