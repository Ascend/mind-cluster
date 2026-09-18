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

"""HostCollectorA5 单元测试：A5 代际注册、多光模块与多网卡多端口采集。"""

import asyncio
import unittest
from unittest.mock import AsyncMock, MagicMock

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.common.constants import OPTICAL_FLAG
from ascend_fd_tk.core.context.register import HOST_COLLECTOR_REGISTRY
from ascend_fd_tk.core.collect.collector.host_collector_a5 import HostCollectorA5
from ascend_fd_tk.core.model.host_a5 import (
    HostInfoA5,
    NpuChipInfoA5,
    DevInfo,
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
    # 设备信息原始回显，由 parser 解析
    fetcher.fetch_dev_info = AsyncMock(return_value="dev_info_raw")
    # 端口 Credit 原始回显，由 parser 解析
    fetcher.fetch_credit_info = AsyncMock(return_value="credit_info_raw")
    # 端口状态回显（含光模块类型），由 parser 解析
    fetcher.fetch_port_state_info = AsyncMock(return_value="port_state_raw")
    # 每个光模块端口信息采集返回非空字符串，由 parser 解析
    fetcher.fetch_optical_port_info = AsyncMock(return_value="optical_info_raw")
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
        # 设备信息：NPU 0 下 UDie 0 有 3 个端口，其中 2 个 Optical、1 个非 Optical
        self.collector.parser.parse_dev_info = MagicMock(
            return_value=[
                DevInfo(udie_id="0", port_id="0", media_type=OPTICAL_FLAG),
                DevInfo(udie_id="0", port_id="1", media_type=OPTICAL_FLAG),
                DevInfo(udie_id="0", port_id="2", media_type="Copper"),
            ]
        )
        # 端口状态信息（media_type 用于填充光模块类型）
        self.collector.parser.parse_port_state_info = MagicMock(return_value=MagicMock(media_type="400G_OCP"))
        # 光模块信息（返回非 None）
        self.collector.parser.parse_optical_info_port = MagicMock(return_value=MagicMock(spec=[]))
        # 网卡解析
        self.collector.parser.parse_nic_card_names = MagicMock(return_value=["eth0"])
        self.collector.parser.parse_nic_port_num = MagicMock(return_value="2")
        self.collector.parser.parse_nic_sfp_info = MagicMock(return_value=MagicMock(spec=[]))

    def test_collect_multi_optical_modules(self):
        """collect 应遍历每个 Optical 端口并填充 npu_chip_info。"""
        host_info = asyncio.run(self.collector.collect())

        self.assertIsInstance(host_info, HostInfoA5)
        self.assertEqual(host_info.host_id, "host1")
        self.assertEqual(host_info.hostname, "node-1")
        self.assertEqual(host_info.sn_num, "SN12345")
        self.assertEqual(host_info.generation, "A5")
        # NPU 0 下有 2 个光模块
        self.assertIn("0", host_info.npu_chip_info)
        chip_info = host_info.npu_chip_info["0"]
        self.assertIsInstance(chip_info, NpuChipInfoA5)
        self.assertEqual(len(chip_info.hccn_optical_info), 2)
        # fetch_optical_port_info 应被调用 2 次（每个 Optical 端口一次）
        self.assertEqual(self.fetcher.fetch_optical_port_info.call_count, 2)
        # 调用参数按 UDie + Port 传递
        self.collector.parser.parse_optical_info_port.assert_any_call("optical_info_raw", "0", "0")
        self.collector.parser.parse_optical_info_port.assert_any_call("optical_info_raw", "0", "1")

    def test_collect_credit_info_all_ports(self):
        """collect 应遍历每个端口（含非 Optical 端口）采集 Credit 信息并填充 credit_info_list。"""
        host_info = asyncio.run(self.collector.collect())
        chip_info = host_info.npu_chip_info["0"]
        # 3 个端口（2 个 Optical + 1 个非 Optical）均采集 Credit
        self.assertEqual(self.fetcher.fetch_credit_info.call_count, 3)
        self.collector.parser.parse_credit_info.assert_any_call("credit_info_raw", "0", "0")
        self.collector.parser.parse_credit_info.assert_any_call("credit_info_raw", "0", "1")
        self.collector.parser.parse_credit_info.assert_any_call("credit_info_raw", "0", "2")
        self.assertEqual(len(chip_info.credit_info_list), 3)

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
