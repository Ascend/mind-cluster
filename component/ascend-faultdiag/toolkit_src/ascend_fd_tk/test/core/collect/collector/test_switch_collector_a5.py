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

"""SwitchCollectorA5 单元测试：A5 代际写入、800G 光口过滤与光模块解析合并。"""

import asyncio
import unittest
from unittest.mock import AsyncMock, MagicMock, patch

from ascend_fd_tk.core.collect.collector.switch_collector_a5 import SwitchCollectorA5
from ascend_fd_tk.core.collect.collector.switch_collector import SwitchCollector
from ascend_fd_tk.core.model.switch import InterfaceBrief, SwitchInfo


def _make_fetcher() -> MagicMock:
    """构造具备全部 fetch 方法的 mock fetcher。"""
    fetcher = MagicMock()
    fetcher.host = "10.1.1.1"
    fetcher.init_fetcher = AsyncMock()
    fetcher.get_switch_name = AsyncMock(return_value="swi1")
    fetcher.fetch_id = AsyncMock(return_value="10.1.1.1")
    for name in [
        "fetch_serial_num",
        "fetch_interface_brief",
        "fetch_optical_module_info",
        "fetch_switch_log_info",
        "fetch_interface_port_mapping",
        "fetch_active_alarms",
        "fetch_active_alarms_verbose",
        "fetch_history_alarms",
        "fetch_history_alarms_verbose",
        "fetch_interface_info",
        "fetch_lldp_nei_brief",
        "fetch_datetime",
        "fetch_bit_error_rate",
        "fetch_transceiver_info",
    ]:
        setattr(fetcher, name, AsyncMock(return_value="raw"))
    return fetcher


def _make_parser() -> MagicMock:
    """构造全部 parse 方法默认返回空结果的 mock parser。"""
    parser = MagicMock()
    parser.parse_esn = MagicMock(return_value="SN123")
    parser.parse_datetime = MagicMock(return_value="2026-01-01")
    for name in [
        "parse_interface_brief",
        "parse_opt_module_info_from_table",
        "parse_opt_module_info_from_line",
        "parse_port_mapping",
        "parse_alarms",
        "parse_alarm_verbose",
        "parse_interface_info",
        "parse_lldp_nei_brief",
        "parse_bit_err_rate",
        "parse_transceiver_info",
    ]:
        setattr(parser, name, MagicMock(return_value=[]))
    return parser


class TestSwitchCollectorA5(unittest.TestCase):
    def test_collect_sets_generation_a5(self):
        """collect() 在基类结果基础上写入 A5 代际标识。"""
        with patch.object(SwitchCollector, "collect", new=AsyncMock(return_value=SwitchInfo("swi1", "10.1.1.1"))):
            info = asyncio.run(SwitchCollectorA5(MagicMock()).collect())
        self.assertEqual(info.generation, "A5")
        self.assertEqual(info.name, "swi1")

    def test_coll_optical_module_info_parses_table_and_line(self):
        """继承的在线 table 解析与离线 line 解析结果应合并，table 在前。"""
        fetcher = MagicMock()
        fetcher.fetch_optical_module_info = AsyncMock(return_value="table_raw")
        fetcher.fetch_switch_log_info = AsyncMock(return_value="log_raw")
        fetcher.fetch_interface_port_mapping = AsyncMock(return_value="pm_raw")
        parser = MagicMock()
        parser.parse_opt_module_info_from_table = MagicMock(return_value=["T"])
        parser.parse_opt_module_info_from_line = MagicMock(return_value=["L"])
        parser.parse_port_mapping = MagicMock(return_value={})
        collector = SwitchCollectorA5(fetcher)
        collector.parser = parser
        result = asyncio.run(collector.coll_optical_module_info([]))
        self.assertEqual(result, ["T", "L"])
        parser.parse_port_mapping.assert_called_once_with("pm_raw")

    def test_get_optical_interface_filters_800g(self):
        """A5 光模块采集前应只保留 800G 光口。"""
        collector = SwitchCollectorA5(MagicMock())
        briefs = [
            InterfaceBrief(interface="800G-1/1/1"),
            InterfaceBrief(interface="400G-1/1/2"),
            InterfaceBrief(interface="GE0/0/1"),
            InterfaceBrief(interface="800G-2/1/1"),
        ]
        result = asyncio.run(collector.get_optical_interface(briefs))
        self.assertEqual([brief.interface for brief in result], ["800G-1/1/1", "800G-2/1/1"])

    def test_base_collect_dispatches_to_a5_optical_override(self):
        """基类 collect() 应调用被 A5 重写的 get_optical_interface（800G 过滤）。"""

        async def boom(self, interface_briefs):  # pylint: disable=unused-argument
            raise AssertionError("collect 应分发到 A5 重写，而非基类 get_optical_interface")

        fetcher = _make_fetcher()
        fetcher.last_quit = AsyncMock()
        parser = _make_parser()
        parser.parse_interface_brief = MagicMock(
            return_value=[
                InterfaceBrief(interface="800G-1/1/1"),
                InterfaceBrief(interface="GE0/0/1"),
            ]
        )
        with patch.object(SwitchCollector, "get_optical_interface", new=boom):
            collector = SwitchCollectorA5(fetcher)
            collector.parser = parser
            info = asyncio.run(collector.collect())
        self.assertEqual(info.generation, "A5")
        # 800G 光口被 A5 重写过滤后传入 fetch_optical_module_info
        fetcher.fetch_optical_module_info.assert_awaited_once()
        briefs = fetcher.fetch_optical_module_info.await_args.args[0]
        self.assertEqual([brief.interface for brief in briefs], ["800G-1/1/1"])


if __name__ == "__main__":
    unittest.main()
