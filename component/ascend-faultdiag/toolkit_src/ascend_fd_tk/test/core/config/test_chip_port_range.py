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

"""chip_port_range 单元测试：天宸芯片各端口分组的数据一致性校验。"""

import unittest

from ascend_fd_tk.core.config import chip_port_range as cpr


class TestChipPortRange(unittest.TestCase):
    def test_all_chips_have_expected_lengths(self):
        for idx in range(7):
            cpu_port = cpr.tiancheng_cpu_port_list[idx]
            npu_port = cpr.tiancheng_npu_port_list[idx]
            xpu_port = cpr.tiancheng_xpu_port_list[idx]
            sw_port = getattr(cpr, f"tiancheng_chip_{idx}_sw_port")
            all_port = getattr(cpr, f"tiancheng_chip_{idx}_all_port")
            self.assertEqual(len(npu_port), 16)
            self.assertEqual(len(sw_port), 16)
            self.assertEqual(len(xpu_port), len(cpu_port) + len(npu_port))
            self.assertEqual(len(all_port), len(sw_port) + len(xpu_port))

    def test_xpu_port_is_union_of_cpu_and_npu(self):
        for idx in range(7):
            self.assertEqual(
                cpr.tiancheng_xpu_port_list[idx],
                cpr.tiancheng_cpu_port_list[idx] + cpr.tiancheng_npu_port_list[idx],
            )

    def test_all_port_has_no_duplicates(self):
        for idx in range(7):
            all_port = getattr(cpr, f"tiancheng_chip_{idx}_all_port")
            self.assertEqual(len(all_port), len(set(all_port)))

    def test_chip0_port_values(self):
        self.assertEqual(cpr.tiancheng_chip_0_npu_port, list(range(12, 20)) + list(range(36, 44)))
        self.assertEqual(cpr.tiancheng_chip_0_cpu_port, [20, 21, 22, 23])
        self.assertEqual(
            cpr.tiancheng_chip_0_sw_port,
            list(range(0, 4)) + list(range(6, 10)) + list(range(24, 28)) + list(range(30, 34)),
        )


if __name__ == "__main__":
    unittest.main()
