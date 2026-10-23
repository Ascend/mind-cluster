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

"""CLI 回显解析数据结构单元测试：写入、追加与多级读取。"""

import unittest

from ascend_fd_tk.core.collect.collect_config import ToolLogCollectionDataType
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.cli_output_parsed_data import CliOutputParsedData


class TestCliOutputParsedData(unittest.TestCase):
    def test_add_data_and_get_data_dict(self):
        data = CliOutputParsedData()
        data.add_data(["LLDP", "0"], "data0")
        data.add_data(["LLDP", "1"], "data1")
        self.assertEqual(data.get_data_dict(), {"LLDP": {"0": "data0", "1": "data1"}})

    def test_append_str_data(self):
        data = CliOutputParsedData()
        data.append_str_data(["SPEED", "0"], "100")
        data.append_str_data(["SPEED", "0"], "1000")
        self.assertEqual(data.fetch_data(["SPEED", "0"]), "1001000")

    def test_fetch_data_with_default(self):
        data = CliOutputParsedData()
        self.assertEqual(data.fetch_data(["NO", "KEY"], "x"), "x")
        self.assertEqual(data.fetch_data_by_name("NO_KEY", "y"), "y")

    def test_fetch_by_chip_and_device_id(self):
        data = CliOutputParsedData()
        data.add_data(["OPTICAL", "0"], "optical0")
        data.add_data(["HCCS", "0", "1"], "hccs01")
        self.assertEqual(data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.OPTICAL, "0"), "optical0")
        self.assertEqual(data.fetch_data_by_device_chip_id(ToolLogCollectionDataType.HCCS, "0", "1"), "hccs01")
        self.assertEqual(data.fetch_data_by_chip_phy_id(ToolLogCollectionDataType.HCCS, "9"), "")


if __name__ == "__main__":
    unittest.main()
