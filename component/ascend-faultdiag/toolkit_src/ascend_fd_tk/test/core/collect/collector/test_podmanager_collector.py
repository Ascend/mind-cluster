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

"""PoDManagerCollector 单元测试：串行切换槽位复用 A5 采集，逐槽位补写 swi_id/slot_id。"""

import asyncio
import unittest
from unittest.mock import AsyncMock, MagicMock, call, patch

from ascend_fd_tk.core.collect.collector.podmanager_collector import PoDManagerCollector
from ascend_fd_tk.core.collect.collector.switch_collector_a5 import SwitchCollectorA5
from ascend_fd_tk.core.model.switch import SwitchInfo


def _make_fetcher(slot_ids=(61, 62, 63)) -> MagicMock:
    """构造绑定一组槽位的 mock fetcher。"""
    fetcher = MagicMock()
    fetcher.slot_ids = list(slot_ids)
    fetcher.executor = MagicMock(host="10.1.1.1")
    fetcher.switch_slot = AsyncMock()
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
        fetcher.switch_slot.assert_has_awaits([call(61), call(62), call(63)])

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
        fetcher.switch_slot.assert_has_awaits([call(61), call(62), call(63)])


if __name__ == "__main__":
    unittest.main()
