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

import unittest
from types import SimpleNamespace
from unittest.mock import AsyncMock, patch

from ascend_fd_tk.core.service.collect_bmc_info import CollectBmcsInfo
from ascend_fd_tk.core.service.collect_swi_info import CollectSwiInfo


def _fake_ctx(existing, fetchers_key, fetchers):
    return SimpleNamespace(
        **{
            fetchers_key: fetchers,
            "location_config": SimpleNamespace(
                enrich_switch_info=lambda *a, **k: None, enrich_host_info=lambda *a, **k: None
            ),
            "cache": SimpleNamespace(swis_info=existing, bmcs_info=existing),
        }
    )


class TestCollectInfoMerge(unittest.IsolatedAsyncioTestCase):
    async def test_collect_swi_info_merges_not_replaces(self):
        """模拟 PoDManager 先合并完成的时序：CollectSwiInfo 应增量合并，保留既有槽位条目。"""
        existing = {"1.1.1.1_61": "podmanager_slot61"}
        ctx = _fake_ctx(existing, "switch_fetchers", {"1.1.1.2": object()})
        fake_swi = SimpleNamespace(swi_id="1.1.1.2")
        with patch("ascend_fd_tk.core.service.collect_swi_info.SwitchCollector") as collector_cls:
            collector_cls.return_value.collect = AsyncMock(return_value=fake_swi)
            await CollectSwiInfo(ctx).run()
        self.assertEqual(ctx.cache.swis_info, {"1.1.1.1_61": "podmanager_slot61", "1.1.1.2": fake_swi})

    async def test_collect_bmc_info_merges_not_replaces(self):
        """模拟 PoDManager 先合并完成的时序：CollectBmcsInfo 应增量合并，保留既有槽位条目。"""
        existing = {"1.1.1.1_61": "podmanager_bmc_slot61"}
        ctx = _fake_ctx(existing, "bmcs_fetchers", {"1.1.1.9": object()})
        fake_bmc = SimpleNamespace(bmc_id="1.1.1.9")
        with patch("ascend_fd_tk.core.service.collect_bmc_info.BmcCollector") as collector_cls:
            collector_cls.return_value.collect = AsyncMock(return_value=fake_bmc)
            await CollectBmcsInfo(ctx).run()
        self.assertEqual(ctx.cache.bmcs_info, {"1.1.1.1_61": "podmanager_bmc_slot61", "1.1.1.9": fake_bmc})


if __name__ == "__main__":
    unittest.main()
