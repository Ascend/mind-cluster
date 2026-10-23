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

"""信息采集服务测试：BMC/Host/Switch/回环/链路HCCS 采集与 cache 更新"""

import unittest
from unittest.mock import AsyncMock, MagicMock, patch

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.service.collect_bmc_info import CollectBmcsInfo
from ascend_fd_tk.core.service.collect_bmc_log import CollectBmcLogService
from ascend_fd_tk.core.service.collect_host_info import CollectHostsInfo
from ascend_fd_tk.core.service.collect_l1_hccs_info import CollectL1HccsInfo
from ascend_fd_tk.core.service.collect_loopback_info import CollectLoopbackInfo
from ascend_fd_tk.core.service.collect_swi_info import CollectSwiInfo

_IP = "10.1.1.1"


class TestCollectServices(unittest.IsolatedAsyncioTestCase):
    def setUp(self):
        self.ctx = MagicMock()
        self.cache = MagicMock()
        self.ctx.cache = self.cache

    async def test_collect_bmcs(self):
        self.cache.bmcs_info = {}
        self.ctx.bmcs_fetchers = {}
        await CollectBmcsInfo(self.ctx).run()
        self.assertEqual(self.cache.bmcs_info, {})
        self.ctx.bmcs_fetchers = {_IP: MagicMock()}
        bmc_info = MagicMock(bmc_id=_IP)
        with patch("ascend_fd_tk.core.service.collect_bmc_info.BmcCollector") as mock_collector:
            mock_collector.return_value.collect = AsyncMock(return_value=bmc_info)
            await CollectBmcsInfo(self.ctx).run()
            self.assertEqual(self.cache.bmcs_info, {_IP: bmc_info})

    async def test_collect_bmc_log(self):
        self.ctx.bmcs_fetchers = {}
        with patch("ascend_fd_tk.core.service.collect_bmc_log.DIAG_LOGGER") as mock_logger:
            await CollectBmcLogService(self.ctx).run()
            mock_logger.error.assert_called()
        self.ctx.bmcs_fetchers = {_IP: MagicMock()}
        with (
            patch("ascend_fd_tk.core.service.collect_bmc_log.BmcLogCollector") as mock_collector,
            patch("ascend_fd_tk.core.service.collect_bmc_log.DIAG_LOGGER"),
        ):
            mock_collector.return_value.collect = AsyncMock(return_value=None)
            await CollectBmcLogService(self.ctx).run()
            mock_collector.return_value.collect.assert_awaited_once()

    async def test_collect_hosts(self):
        self.cache.hosts_info = {}
        self.ctx.host_fetchers = {}
        with patch("ascend_fd_tk.core.service.collect_host_info.create_host_collector") as mock_create:
            await CollectHostsInfo(self.ctx).run()
            mock_create.assert_not_called()
        self.ctx.host_fetchers = {_IP: MagicMock(generation=NpuType.A5)}
        host_info = MagicMock(host_id=_IP)
        collector = MagicMock()
        collector.collect = AsyncMock(return_value=host_info)
        with patch("ascend_fd_tk.core.service.collect_host_info.create_host_collector", return_value=collector):
            await CollectHostsInfo(self.ctx).run()
            self.ctx.location_config.enrich_host_info.assert_called_once_with(host_info)
            self.assertEqual(host_info.generation, "A5")
            self.assertEqual(self.cache.hosts_info, {_IP: host_info})
        self.ctx.host_fetchers = {_IP: MagicMock(generation=None)}
        host_info = MagicMock(host_id=_IP)
        collector = MagicMock()
        collector.collect = AsyncMock(return_value=host_info)
        with patch("ascend_fd_tk.core.service.collect_host_info.create_host_collector", return_value=collector):
            await CollectHostsInfo(self.ctx).run()
            self.assertEqual(host_info.generation, "A3")

    async def test_collect_l1_hccs(self):
        self.ctx.switch_fetchers = {}
        with patch("ascend_fd_tk.core.service.collect_l1_hccs_info.HccsCollector") as mock_collector:
            await CollectL1HccsInfo(self.ctx).run()
            mock_collector.assert_not_called()
        self.ctx.switch_fetchers = {_IP: MagicMock()}
        switch_info = MagicMock()
        self.cache.swis_info.get.return_value = switch_info
        hccs_info = MagicMock()
        with patch("ascend_fd_tk.core.service.collect_l1_hccs_info.HccsCollector") as mock_collector:
            mock_collector.return_value.collect = AsyncMock(return_value=hccs_info)
            await CollectL1HccsInfo(self.ctx).run()
            self.assertIs(switch_info.hccs_info, hccs_info)
        self.cache.swis_info.get.return_value = None
        with (
            patch("ascend_fd_tk.core.service.collect_l1_hccs_info.HccsCollector") as mock_collector,
            patch("ascend_fd_tk.core.service.collect_l1_hccs_info._DIAG_LOGGER") as mock_logger,
        ):
            mock_collector.return_value.collect = AsyncMock(return_value=MagicMock())
            await CollectL1HccsInfo(self.ctx).run()
            mock_logger.warning.assert_called()

    async def test_collect_loopback(self):
        self.cache.hosts_info = {}
        self.ctx.host_fetchers = {}
        with patch("ascend_fd_tk.core.service.collect_loopback_info.HostLoopbackCollector") as mock_collector:
            await CollectLoopbackInfo(self.ctx).run()
            mock_collector.assert_not_called()
        self.ctx.host_fetchers = {_IP: MagicMock()}
        host_info = MagicMock(host_id=_IP)
        with patch("ascend_fd_tk.core.service.collect_loopback_info.HostLoopbackCollector") as mock_collector:
            mock_collector.return_value.collect = AsyncMock(return_value=host_info)
            await CollectLoopbackInfo(self.ctx).run()
            self.assertEqual(self.cache.hosts_info, {_IP: host_info})

    async def test_collect_swis(self):
        self.cache.swis_info = {}
        self.ctx.switch_fetchers = {}
        with patch("ascend_fd_tk.core.service.collect_swi_info.SwitchCollector") as mock_collector:
            await CollectSwiInfo(self.ctx).run()
            mock_collector.assert_not_called()
        self.ctx.switch_fetchers = {_IP: MagicMock()}
        switch_info = MagicMock(swi_id=_IP)
        with patch("ascend_fd_tk.core.service.collect_swi_info.SwitchCollector") as mock_collector:
            mock_collector.return_value.collect = AsyncMock(return_value=switch_info)
            await CollectSwiInfo(self.ctx).run()
            self.ctx.location_config.enrich_switch_info.assert_called_once_with(switch_info)
            self.assertEqual(self.cache.swis_info, {_IP: switch_info})


if __name__ == "__main__":
    unittest.main()
