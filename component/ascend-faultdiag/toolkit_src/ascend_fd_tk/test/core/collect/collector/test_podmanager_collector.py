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

import asyncio
import unittest
from unittest.mock import AsyncMock, MagicMock, call, patch

from ascend_fd_tk.core.collect.collector.podmanager_collector import PoDManagerCollector
from ascend_fd_tk.core.collect.collector.switch_collector_a5 import SwitchCollectorA5
from ascend_fd_tk.core.model.switch import SwitchInfo


def _make_fetcher(sfu_slot_ids=("61", "62", "63"), npu_slot_ids=()) -> MagicMock:
    """构造绑定一组槽位的 mock fetcher。"""
    fetcher = MagicMock()
    fetcher.sfu_slot_ids = list(sfu_slot_ids)
    fetcher.npu_slot_ids = list(npu_slot_ids)
    fetcher.executor = MagicMock(host="10.1.1.1")
    fetcher.switch_slot = AsyncMock()
    fetcher.init_fetcher = AsyncMock()
    fetcher.last_quit = AsyncMock()
    fetcher.get_switch_name = AsyncMock(return_value="swi_name")
    fetcher.fetch_qos_credit = AsyncMock(return_value="")
    return fetcher


class TestPoDManagerCollector(unittest.TestCase):
    def test_get_id_format(self):
        """get_id 应拼接 host 与槽位列表。"""
        self.assertEqual(asyncio.run(PoDManagerCollector(_make_fetcher()).get_id()), "10.1.1.1_slots_61,62,63")

    def test_collect_iterates_slots_and_marks_a5(self):
        """collect() 应按序切换槽位并逐槽位调用 A5 采集器，补写 swi_id/slot_id。"""
        fetcher = _make_fetcher()
        collector = PoDManagerCollector(fetcher)
        with patch.object(
            SwitchCollectorA5, "collect", new=AsyncMock(side_effect=[SwitchInfo("s", "i") for _ in range(3)])
        ):
            switch_list, bmc_list = asyncio.run(collector.collect())
        self.assertEqual(
            [(s.swi_id, s.slot_id) for s in switch_list], [("10.1.1.1", "61"), ("10.1.1.1", "62"), ("10.1.1.1", "63")]
        )
        self.assertEqual(bmc_list, [])
        fetcher.switch_slot.assert_has_awaits([call("61"), call("62"), call("63")])

    def test_collect_isolates_single_slot_failure(self):
        """单个槽位采集失败时应跳过该槽位，其余槽位正常返回。"""
        fetcher = _make_fetcher()
        collector = PoDManagerCollector(fetcher)
        with patch.object(
            SwitchCollectorA5,
            "collect",
            new=AsyncMock(side_effect=[SwitchInfo("s", "i"), RuntimeError("slot down"), SwitchInfo("s", "i")]),
        ):
            switch_list, bmc_list = asyncio.run(collector.collect())
        self.assertEqual([(s.swi_id, s.slot_id) for s in switch_list], [("10.1.1.1", "61"), ("10.1.1.1", "63")])
        self.assertEqual(bmc_list, [])
        fetcher.switch_slot.assert_has_awaits([call("61"), call("62"), call("63")])

    def test_collect_npu_slots_qos_credit(self):
        """NPU 槽位应跳转后采集基础信息与 QoS Credit，每个槽位生成一个 SwitchInfo。"""
        from ascend_fd_tk.core.model.switch import QosCreditInfo

        fetcher = _make_fetcher(sfu_slot_ids=("61",), npu_slot_ids=("18", "19"))
        credit_infos = [QosCreditInfo(slot_id="18", chip_id="0", ports=[])]
        collector = PoDManagerCollector(fetcher)
        with (
            patch.object(SwitchCollectorA5, "collect", new=AsyncMock(return_value=SwitchInfo("s", "i"))),
            patch.object(SwitchCollectorA5, "coll_serial_num", new=AsyncMock(return_value="SN123")),
            patch.object(SwitchCollectorA5, "coll_datetime", new=AsyncMock(return_value="2026-09-17")),
            patch.object(SwitchCollectorA5, "coll_interface_brief", new=AsyncMock(return_value=[])),
            patch.object(collector.parser, "parse_qos_credit", return_value=credit_infos) as mock_parse,
        ):
            switch_list, _ = asyncio.run(collector.collect())
        # SFU 槽位 + 每个 NPU 槽位各自生成 SwitchInfo（两个 NPU 槽位均命中 mock 解析结果）
        self.assertEqual(
            [(s.swi_id, s.slot_id) for s in switch_list],
            [("10.1.1.1", "61"), ("10.1.1.1", "18"), ("10.1.1.1", "19")],
        )
        npu_switch = switch_list[1]
        self.assertEqual(
            (npu_switch.name, npu_switch.sn, npu_switch.date_time, npu_switch.generation),
            ("swi_name", "SN123", "2026-09-17", "A5"),
        )
        self.assertIs(npu_switch.qos_credit_infos, credit_infos)
        self.assertEqual(mock_parse.call_count, 2)
        fetcher.switch_slot.assert_has_awaits([call("61"), call("18"), call("19")])
        self.assertEqual(fetcher.last_quit.await_count, 2)

    def test_collect_npu_slot_skips_when_parse_empty(self):
        """NPU 槽位解析结果为空时不生成 SwitchInfo，也不影响其余槽位。"""
        fetcher = _make_fetcher(sfu_slot_ids=(), npu_slot_ids=("18", "19"))
        collector = PoDManagerCollector(fetcher)
        with patch.object(collector.parser, "parse_qos_credit", return_value=[]):
            switch_list, _ = asyncio.run(collector.collect())
        self.assertEqual(switch_list, [])
        fetcher.switch_slot.assert_has_awaits([call("18"), call("19")])


if __name__ == "__main__":
    unittest.main()
