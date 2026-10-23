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

"""RootCauseAnalysis 根因分析服务测试"""

import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.service.root_cause_analysis import RootCauseAnalysis


class TestRootCauseAnalysis(unittest.IsolatedAsyncioTestCase):
    async def test_run_builds_root_cause_filter(self):
        ctx = MagicMock()
        with patch("ascend_fd_tk.core.service.root_cause_analysis.RootCauseFilter") as mock_filter:
            mock_filter.return_value.build_and_analyze = MagicMock()
            await RootCauseAnalysis(ctx).run()
            mock_filter.assert_called_once_with(ctx.cache)
            self.assertIs(ctx.root_cause_filter, mock_filter.return_value)
            mock_filter.return_value.build_and_analyze.assert_called_once()


if __name__ == "__main__":
    unittest.main()
