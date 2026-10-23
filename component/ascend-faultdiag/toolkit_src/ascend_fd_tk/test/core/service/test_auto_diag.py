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

"""AutoDiag 诊断服务测试：按代际分流 analyzer 并延迟应用阈值配置"""

import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.service.auto_diag import AutoDiag


class _FakeAnalyzer:
    def __init__(self, cache):
        self.cache = cache

    def analyse(self):
        return [{"result": "fake"}]


class TestAutoDiag(unittest.IsolatedAsyncioTestCase):
    async def _run(self, generation):
        ctx = MagicMock()
        ctx.cache.generation = generation
        ctx.threshold_loader = MagicMock()
        ctx.diag_result = []
        with (
            patch("ascend_fd_tk.core.service.auto_diag.recursive_scan_and_register") as mock_scan,
            patch("ascend_fd_tk.core.service.auto_diag.get_analyzers") as mock_get,
        ):
            mock_get.return_value = [_FakeAnalyzer]
            await AutoDiag(ctx).run()
        return ctx, mock_scan, mock_get

    async def test_run_default_a3(self):
        ctx, mock_scan, mock_get = await self._run("")
        mock_scan.assert_called_once_with("ascend_fd_tk.core.fault_analyzer")
        mock_get.assert_called_once_with(NpuType.A3)
        ctx.threshold_loader.apply.assert_called_once_with()
        self.assertEqual(ctx.diag_result, [{"result": "fake"}])

    async def test_run_a5_generation(self):
        ctx, mock_scan, mock_get = await self._run("A5")
        mock_get.assert_called_once_with(NpuType.A5)
        ctx.threshold_loader.apply.assert_called_once_with()
        self.assertEqual(ctx.diag_result, [{"result": "fake"}])


if __name__ == "__main__":
    unittest.main()
