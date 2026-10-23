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

"""cluster_mapping 单元测试：机框映射查找与默认 NPU 映射构建。"""

import unittest

from ascend_fd_tk.core.model.cluster_mapping import (
    ChassisMapping,
    DEFAULT_NPU_MAPPING,
    DEFAULT_NPU_PHY_MAPPING,
    DefaultNPUInfo,
    L1SwiServerMapping,
)


class TestChassisMapping(unittest.TestCase):
    def setUp(self):
        self.mappings = [
            L1SwiServerMapping("1.1.1.1", "sw0", "spod-0", "10.0.0.1", "10.0.0.100"),
            L1SwiServerMapping("1.1.1.2", "sw1", "spod-1", "10.0.0.2", "10.0.0.101"),
        ]
        self.chassis = ChassisMapping(self.mappings)

    def test_find_by_bmc_ip(self):
        mapping = self.chassis.find_mapping_by_bmc_ip("10.0.0.101")
        self.assertEqual(mapping.l1_swi_ip, "1.1.1.2")

    def test_find_by_l1_swi_ip(self):
        mapping = self.chassis.find_mapping_by_l1_swi_ip("1.1.1.1")
        self.assertEqual(mapping.server_ip, "10.0.0.1")

    def test_find_by_server_ip(self):
        mapping = self.chassis.find_mapping_by_server_ip("10.0.0.2")
        self.assertEqual(mapping.bmc_ip, "10.0.0.101")

    def test_find_missing_returns_none(self):
        self.assertIsNone(self.chassis.find_mapping_by_server_ip("10.0.0.9"))


class TestDefaultNPUInfo(unittest.TestCase):
    def test_build_default_npu_mapping_size(self):
        npu_mapping, npu_phy_mapping = DefaultNPUInfo.build_default_npu_mapping()
        self.assertEqual(len(npu_mapping), 16)
        self.assertEqual(len(npu_phy_mapping), 16)

    def test_build_default_npu_mapping_relation(self):
        npu_mapping, npu_phy_mapping = DefaultNPUInfo.build_default_npu_mapping()
        self.assertEqual(npu_mapping["0"].npu_id, "0")
        self.assertEqual(npu_mapping["0"].chip_id, "0")
        self.assertEqual(npu_mapping["1"].npu_id, "0")
        self.assertEqual(npu_mapping["1"].chip_id, "1")
        self.assertEqual(npu_phy_mapping[("0", "1")].chip_phy_id, "1")

    def test_module_globals_consistent(self):
        self.assertEqual(len(DEFAULT_NPU_MAPPING), 16)
        self.assertEqual(DEFAULT_NPU_PHY_MAPPING[("3", "0")].chip_phy_id, "6")


if __name__ == "__main__":
    unittest.main()
