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

"""LoopbackDiag 回环诊断服务测试"""

import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.service.loopback_diag import LoopbackDiag


class TestLoopbackDiag(unittest.IsolatedAsyncioTestCase):
    async def test_run_appends_loopback_results(self):
        ctx = MagicMock()
        ctx.diag_result = []
        fake_result = MagicMock()
        with patch("ascend_fd_tk.core.service.loopback_diag.HostLoopbackAnalyzer") as mock_analyzer:
            mock_analyzer.return_value.analyse.return_value = [fake_result]
            await LoopbackDiag(ctx).run()
            self.assertEqual(ctx.diag_result, [fake_result])


if __name__ == "__main__":
    unittest.main()
