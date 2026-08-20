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

"""HostParserA5 单元测试。

覆盖本次改动：
- 正则使用 {n,} 替代 * 和 +，验证 lane 动态匹配
- parse_nic_sfp_info 按 lane 号动态收集（不硬编码 lane 数量）
"""

import unittest

from ascend_fd_tk.core.collect.parser.host_parser_a5 import HostParserA5


class TestHostParserA5(unittest.TestCase):
    def test_parse_nic_sfp_info_dynamic_lanes(self):
        """验证 lane 号动态收集，适配不同 lane 数量。"""
        cmd_res = """Internally measured Tx1 bias current monitor: 6.523 mA
Internally measured Tx2 bias current monitor: 6.618 mA
Internally measured Tx1 output optical power: 0.49 mW / -3.1 dBm
Internally measured Tx2 output optical power: 0.51 mW / -2.9 dBm
Internally measured Rx1 input optical power: 0.81 mW / -0.91 dBm
Latched Tx LOS flag(OK), lane 1
Latched Rx LOS flag(OK), media lane 1
Latched Tx CDR LOL flag(OK), lane 1
Host side SNR lane 1: 20.32 dB
Media side SNR lane 1: 18.45 dB
|----host1("""

        result = HostParserA5.parse_nic_sfp_info(cmd_res, port_id="0")
        self.assertEqual(result.port_id, "0")
        # 2 个 lane（按 max_lane_id 填充，缺失的填空串）
        self.assertEqual(len(result.bias_lanes), 2)
        self.assertEqual(result.bias_lanes[0], "6.523")
        self.assertEqual(result.bias_lanes[1], "6.618")
        self.assertEqual(result.tx_power_lanes[0], "-3.1")
        self.assertEqual(result.tx_power_lanes[1], "-2.9")
        self.assertEqual(result.rx_power_lanes[0], "-0.91")
        # tx_los 只有 lane 1 有值，lane 2 填空串
        self.assertEqual(len(result.tx_los_lanes), 2)
        self.assertEqual(result.tx_los_lanes[0], "OK")
        self.assertEqual(result.tx_los_lanes[1], "")
        self.assertEqual(result.host_snr_lanes[0], "20.32")
        self.assertEqual(result.media_snr_lanes[0], "18.45")

    def test_parse_nic_sfp_info_empty(self):
        """空输入应返回空 lane 列表。"""
        result = HostParserA5.parse_nic_sfp_info("", port_id="0")
        self.assertEqual(result.bias_lanes, [])
        self.assertEqual(result.tx_power_lanes, [])


if __name__ == "__main__":
    unittest.main()
