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
from typing import List

from ascend_fd.model.parse_info import KGParseFilePath
from ascend_fd.utils.json_dict import JsonObj
from ascend_fd.utils.regular_table import KG_MIN_TIME


class KGParseCtx(JsonObj):
    def __init__(self, parse_file_path: KGParseFilePath = None,
                 resuming_training_time: str = KG_MIN_TIME,
                 is_sdk_input: bool = False,
                 custom_info_list: List = None):
        self.parse_file_path = parse_file_path
        self.resuming_training_time = resuming_training_time
        self.is_sdk_input = is_sdk_input
        self.custom_info_list = custom_info_list or []
