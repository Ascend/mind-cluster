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

"""HostCollectorA5 单元测试。

覆盖本次改动：
- @register_host_collector(NpuType.A5) 代际注册
- collect() 多光模块采集流程：每个 NPU 遍历 optical_top_headline，逐个采集光模块信息
- collect_nic_info() 多网卡多端口遍历
"""

import asyncio
import unittest
from unittest.mock import AsyncMock, MagicMock

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.context.register import HOST_COLLECTOR_REGISTRY
from ascend_fd_tk.core.collect.collector.host_collector_a5 import HostCollectorA5
from ascend_fd_tk.core.model.host_a5 import (
    HostInfoA5,
    NpuChipInfoA5,
    OpticalTopHeadline,
    NICInfoA5,
)


def _build_mock_fetcher():
    """构造 mock fetcher，预设各 fetch 方法返回值。"""
    fetcher = MagicMock()

    # 基础信息
    fetcher.fetch_id = AsyncMock(return_value="host1")
    fetcher.fetch_hostname = AsyncMock(return_value="node-1")
    fetcher.fetch_sn_num = AsyncMock(return_value="SN12345")
    fetcher.fetch_msnpureport_log = AsyncMock(return_value=[])
    fetcher.fetch_npu_type = AsyncMock(return_value="910B5")
    fetcher.fetch_npu_mapping = AsyncMock(return_value={"0": {"chip_phy_id": "0"}})

    # 光模块顶部标题：NPU 0 有 2 个光模块（id 0 和 1）
    fetcher.fetch_optical_top_headline = AsyncMock(
        return_value=[
            OpticalTopHeadline(npu_id="0", optical_silk_screen_num="1", optical_id="0"),
            OpticalTopHeadline(npu_id="0", optical_silk_screen_num="2", optical_id="1"),
        ]
    )
    # 每个光模块信息采集返回非空字符串，由 parser 解析
    fetcher.fetch_optical_info_a5 = AsyncMock(return_value="optical_info_raw")

    # 网卡：返回 1 张网卡名，端口数 "2"
    fetcher.fetch_nic_info = AsyncMock(return_value="nic_list_raw")
    fetcher.fetch_nic_port_num = AsyncMock(return_value="port_num_raw")
    fetcher.fetch_nic_sfp_info = AsyncMock(return_value="sfp_raw")
    return fetcher


class TestHostCollectorA5Registration(unittest.TestCase):
    def test_generation_registration(self):
        """A5 collector 应注册到 HOST_COLLECTOR_REGISTRY。"""
        self.assertIs(HOST_COLLECTOR_REGISTRY[NpuType.A5], HostCollectorA5)


class TestHostCollectorA5Collect(unittest.TestCase):
    def setUp(self):
        self.fetcher = _build_mock_fetcher()
        self.collector = HostCollectorA5(self.fetcher)
        # mock parser 以隔离采集流程
        self.collector.parser = MagicMock()
        self.collector.parser.parse_npu_type = MagicMock(return_value="A5")
        # 顶部标题：NPU 0 有 2 个光模块
        self.collector.parser.parse_optical_top_headline = MagicMock(
            return_value=[
                OpticalTopHeadline(npu_id="0", optical_silk_screen_num="1", optical_id="0"),
                OpticalTopHeadline(npu_id="0", optical_silk_screen_num="2", optical_id="1"),
            ]
        )
        # 光模块信息（返回非 None）
        self.collector.parser.parse_optical_info_a5 = MagicMock(return_value=MagicMock(spec=[]))
        # 网卡解析
        self.collector.parser.parse_nic_card_names = MagicMock(return_value=["eth0"])
        self.collector.parser.parse_nic_port_num = MagicMock(return_value="2")
        self.collector.parser.parse_nic_sfp_info = MagicMock(return_value=MagicMock(spec=[]))

    def test_collect_multi_optical_modules(self):
        """collect 应遍历每个光模块并填充 npu_chip_info。"""
        host_info = asyncio.run(self.collector.collect())

        self.assertIsInstance(host_info, HostInfoA5)
        self.assertEqual(host_info.host_id, "host1")
        self.assertEqual(host_info.hostname, "node-1")
        self.assertEqual(host_info.sn_num, "SN12345")
        self.assertEqual(host_info.chip_generation, "A5")
        # NPU 0 下有 2 个光模块
        self.assertIn("0", host_info.npu_chip_info)
        chip_info = host_info.npu_chip_info["0"]
        self.assertIsInstance(chip_info, NpuChipInfoA5)
        self.assertEqual(len(chip_info.hccn_optical_info), 2)
        # fetch_optical_info_a5 应被调用 2 次（每个光模块一次）
        self.assertEqual(self.fetcher.fetch_optical_info_a5.call_count, 2)

    def test_collect_nic_info_multi_ports(self):
        """collect 应遍历每张网卡的每个端口采集 SFP 信息。"""
        host_info = asyncio.run(self.collector.collect())

        self.assertEqual(len(host_info.nic_info_list), 1)
        nic_info = host_info.nic_info_list[0]
        self.assertIsInstance(nic_info, NICInfoA5)
        self.assertEqual(nic_info.card_name, "eth0")
        self.assertEqual(nic_info.port_num, "2")
        # 每个端口调用一次 parse_nic_sfp_info（端口 0 和 1）
        self.assertEqual(self.collector.parser.parse_nic_sfp_info.call_count, 2)


if __name__ == "__main__":
    unittest.main()
