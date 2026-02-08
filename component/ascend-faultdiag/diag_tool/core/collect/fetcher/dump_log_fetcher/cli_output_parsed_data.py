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

from typing import List

from diag_tool.core.collect.collect_config import ToolLogCollectionDataType
from diag_tool.utils.multi_level_dict_tool import MultiLevelDict


class CliOutputParsedData:

    def __init__(self, data_dict=None):
        self._data = MultiLevelDict(data_dict)

    def get_data_dict(self) -> dict:
        return self._data.get_target_data()

    def add_data(self, keys: List[str], data):
        self._data.write(keys, data)

    def append_str_data(self, keys: List[str], data: str):
        cur = self._data.read(keys)
        if cur is None:
            self._data.write(keys, data)
        else:
            self._data.write(keys, cur + data)

    def fetch_data(self, keys: List[str], default=""):
        return self._data.read(keys, default)

    def fetch_data_by_name(self, name: str, default=""):
        return self._data.read([name], default)

    def fetch_data_by_chip_phy_id(self, key: ToolLogCollectionDataType, chip_phy_id: str):
        return self._data.read([key.name, chip_phy_id], "")

    def fetch_data_by_device_chip_id(self, key: ToolLogCollectionDataType, device_id: str, chip_id: str):
        return self._data.read([key.name, device_id, chip_id], "")
