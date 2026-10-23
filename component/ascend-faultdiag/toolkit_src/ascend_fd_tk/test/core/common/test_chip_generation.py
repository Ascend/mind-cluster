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

"""generation 单元测试：npu-smi info -m 回显解析出代际与 npu_mapping。"""

import unittest

from ascend_fd_tk.core.common.generation import parse_npu_mapping_and_generation
from ascend_fd_tk.core.common.diag_enum import NpuType

A3_STDOUT = """Device count : 1
NPU ID   Chip ID   Chip Logic ID   Chip Phy-ID  Chip Name
0        0         0               0             Ascend910
1        0         1               1             Ascend910
"""

A5_STDOUT = """Device count : 1
NPU ID   Slot ID   Chip ID   Chip Phy-ID  Chip Name
0        0         0               0             Ascend950
1        0         1               1             Ascend950
"""

COMMON = """2        0         2               2             Mcu
-        0         3               3             Ascend910
"""


class TestParseNpuMappingAndGeneration(unittest.TestCase):
    def test_a3_generation_and_mapping(self):
        generation, mapping = parse_npu_mapping_and_generation(A3_STDOUT)
        self.assertEqual(generation, NpuType.A3)
        self.assertEqual(set(mapping.keys()), {"0", "1"})
        self.assertEqual(mapping["0"].npu_id, "0")
        self.assertEqual(mapping["0"].chip_id, "0")
        self.assertEqual(mapping["1"].chip_phy_id, "1")

    def test_a5_generation_and_mapping(self):
        generation, mapping = parse_npu_mapping_and_generation(A5_STDOUT)
        self.assertEqual(generation, NpuType.A5)
        self.assertEqual(mapping["0"].chip_id, "0")

    def test_skip_mcu_and_dash_lines(self):
        stout = A3_STDOUT + COMMON
        generation, mapping = parse_npu_mapping_and_generation(stout)
        self.assertEqual(set(mapping.keys()), {"0", "1"})

    def test_skip_short_lines(self):
        stdout = "NPU ID   Chip ID\n0        0\n"
        generation, mapping = parse_npu_mapping_and_generation(stdout)
        self.assertEqual(generation, NpuType.A3)
        self.assertEqual(mapping, {})


if __name__ == "__main__":
    unittest.main()
