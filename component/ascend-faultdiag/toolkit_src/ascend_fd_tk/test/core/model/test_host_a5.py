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

"""NpuChipInfoA5 单元测试。"""

import unittest

from ascend_fd_tk.core.model.host_a5 import (
    NpuChipInfoA5,
    HCCNOpticalInfoA5,
    OpticalModuleMonitorItem,
)
from ascend_fd_tk.core.model.optical_module import OpticalModuleInfo


def _make_monitor_item(items: str, value: str) -> OpticalModuleMonitorItem:
    return OpticalModuleMonitorItem(items=items, value=value)


def _make_optical_info(udie_id: str, port_id: str, lane_values: dict) -> HCCNOpticalInfoA5:
    """构造 HCCNOpticalInfoA5，lane_values 形如 {"1": {"TxPower": "-3.1", "RxPower": "-0.91"}}。"""
    monitor_items = []
    for lane_id, fields in lane_values.items():
        for field, value in fields.items():
            items_name = f"{field} Lane{lane_id}"
            monitor_items.append(_make_monitor_item(items_name, value))
    return HCCNOpticalInfoA5(udie_id=udie_id, port_id=port_id, monitor_item=monitor_items)


class TestNpuChipInfoA5Optical(unittest.TestCase):
    def test_multi_optical_modules(self):
        """A5 单 NPU 多光模块：get_optical_module_info 返回所有光模块列表。"""
        optical_info_0 = _make_optical_info("0", "0", {"1": {"TxPower": "-3.1", "RxPower": "-0.91", "Bias": "6.523"}})
        optical_info_1 = _make_optical_info("0", "1", {"1": {"TxPower": "-2.9", "RxPower": "-0.85", "Bias": "6.618"}})
        chip = NpuChipInfoA5(hccn_optical_info=[optical_info_0, optical_info_1])

        infos = chip.get_optical_module_info()
        self.assertEqual(len(infos), 2)
        self.assertEqual(infos[0].udie_id, "0")
        self.assertEqual(infos[0].port_id, "0")
        self.assertEqual(infos[1].udie_id, "0")
        self.assertEqual(infos[1].port_id, "1")
        # 第一个光模块 lane1 的 tx_power
        self.assertEqual(infos[0].lane_power_infos[0].tx_power, "-3.1")
        self.assertEqual(infos[0].lane_power_infos[0].rx_power, "-0.91")
        self.assertEqual(infos[0].lane_power_infos[0].bias, "6.523")

    def test_single_optical_module(self):
        """仅一个光模块时返回单元素列表。"""
        optical_info_0 = _make_optical_info("0", "0", {"1": {"TxPower": "-3.1"}})
        chip = NpuChipInfoA5(hccn_optical_info=[optical_info_0])

        infos = chip.get_optical_module_info()
        self.assertEqual(len(infos), 1)
        self.assertIsInstance(infos[0], OpticalModuleInfo)
        self.assertEqual(infos[0].udie_id, "0")
        self.assertEqual(infos[0].port_id, "0")

    def test_empty_optical_info(self):
        """无光模块数据时返回空列表。"""
        chip = NpuChipInfoA5(hccn_optical_info=None)
        self.assertEqual(chip.get_optical_module_info(), [])


if __name__ == "__main__":
    unittest.main()
