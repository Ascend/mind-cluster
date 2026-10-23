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

"""port_mapping_config 单元测试：L1 地址/端口映射解析与查询。"""

import unittest

from ascend_fd_tk.core.config.port_mapping_config import (
    L1GlobalAddrMapping,
    L1InterfacePortMapping,
    L1LocalAddrMapping,
    PortMappingConfig,
)


def build_config() -> PortMappingConfig:
    global_map = {"0-0-0": "0x300000000000-0x300fffffffff", "1-0-0": "0x310000000000-0x310fffffffff"}
    local_map = {"cpu-0-0": "0x000000000000-0x07ffffffffff", "npu-0-0": "0x200040000000-0x20ffffffffff"}
    interface_map = {"200GHL1/1/1": "12-25-NPU3_D0-7", "200GHL1/2/1": "16-17-CPU1_NB-5"}
    return PortMappingConfig(
        [L1GlobalAddrMapping.from_config(k, v) for k, v in global_map.items()],
        [L1LocalAddrMapping.from_config(k, v) for k, v in local_map.items()],
        {k: L1InterfacePortMapping.from_config(k, v) for k, v in interface_map.items()},
    )


class TestMappingFromConfig(unittest.TestCase):
    def test_global_addr_from_config(self):
        m = L1GlobalAddrMapping.from_config("0-0-0", "0x300000000000-0x300fffffffff")
        self.assertEqual((m.server_id, m.npu_id, m.chip_id), ("0", "0", "0"))
        self.assertEqual((m.addr_start, m.addr_end), ("0x300000000000", "0x300fffffffff"))

    def test_local_addr_from_config(self):
        m = L1LocalAddrMapping.from_config("npu-0-0", "0x200040000000-0x20ffffffffff")
        self.assertEqual((m.xpu, m.xpu_id, m.chip_id), ("npu", "0", "0"))

    def test_interface_from_config(self):
        m = L1InterfacePortMapping.from_config("200GHL1/1/1", "12-25-NPU3_D0-7")
        self.assertEqual(m.swi_port, "200GHL1/1/1")
        self.assertEqual(m.swi_chip_id, "0")
        self.assertEqual(m.phy_id, "12")
        self.assertEqual(m.logic_id, "25")
        self.assertEqual(m.xpu, "npu")
        self.assertEqual(m.xpu_id, "2")
        self.assertEqual(m.chip_id, "0")
        self.assertEqual(m.macro, "7")


class TestPortMappingConfig(unittest.TestCase):
    def setUp(self):
        self.config = build_config()

    def test_is_local_addr(self):
        self.assertTrue(self.config.is_local_addr("0x200000000000"))
        self.assertFalse(self.config.is_local_addr("0x300000000000"))
        self.assertTrue(self.config.is_local_addr("not-hex"))

    def test_find_global_addr_port_mapping(self):
        m = self.config.find_global_addr_port_mapping("0x300000000001")
        self.assertEqual(m.server_id, "0")
        m = self.config.find_global_addr_port_mapping("0x310000000001")
        self.assertEqual(m.server_id, "1")
        self.assertIsNone(self.config.find_global_addr_port_mapping("0x2"))

    def test_find_local_addr_port_mapping(self):
        m = self.config.find_local_addr_port_mapping("0x1000000")
        self.assertEqual(m.xpu, "cpu")
        m = self.config.find_local_addr_port_mapping("0x210000000000")
        self.assertEqual(m.xpu, "npu")

    def test_find_swi_port(self):
        m = self.config.find_swi_port("0", phy_id="12")
        self.assertEqual(m.swi_port, "200GHL1/1/1")
        self.assertIsNone(self.config.find_swi_port("9", phy_id="12"))

    def test_find_interface(self):
        self.assertEqual(self.config.find_interface("200GHL1/1/1", "2", "0"), "200GHL1/1/1")
        self.assertEqual(self.config.find_interface("bad", "2", "0"), "")

    def test_find_swi_port_by_cpu_macro(self):
        self.assertEqual(self.config.find_swi_port_by_cpu_macro("1", "B", "5"), "200GHL1/2/1")
        self.assertEqual(self.config.find_swi_port_by_cpu_macro("9", "9", "x"), "")

    def test_find_port_mapping_by_name(self):
        self.assertEqual(self.config.find_port_mapping_by_name("200GHL1/1/1").phy_id, "12")
        self.assertIsNone(self.config.find_port_mapping_by_name("9/9/9"))


if __name__ == "__main__":
    unittest.main()
