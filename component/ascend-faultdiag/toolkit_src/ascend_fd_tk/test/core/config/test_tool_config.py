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

"""tool_config 单元测试：版本号获取与默认值回退。"""

import unittest
from unittest import mock

from ascend_fd_tk.core.config.tool_config import ToolConfig


class TestToolConfig(unittest.TestCase):
    @mock.patch("ascend_fd_tk.core.config.tool_config.version", side_effect=Exception("boom"))
    def test_get_version_unknown_error_returns_empty(self, _mock_version):
        self.assertEqual(ToolConfig.get_version(), "")

    @mock.patch("ascend_fd_tk.core.config.tool_config.version", return_value="v1.2.3")
    def test_get_version_success(self, _mock_version):
        self.assertEqual(ToolConfig.get_version(), "v1.2.3")

    def test_init_fallback_to_default_version(self):
        tool_config = ToolConfig(version_info="v0.10")
        self.assertTrue(tool_config.version)
        self.assertEqual(tool_config.version, tool_config.get_version() or "v0.10")


if __name__ == "__main__":
    unittest.main()
