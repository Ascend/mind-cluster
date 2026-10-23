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

"""errors / constants / path 单元测试：公共异常类与常量/路径定义。"""

import os
import unittest

from ascend_fd_tk.core.common.constants import (
    FAULT_TYPE_BMC,
    FAULT_TYPE_HOST,
    FAULT_TYPE_SWITCH,
    NPU_LINK_DOWN,
    NPU_LONG_DOWN_TIME,
    PACKAGE_NAME,
)
from ascend_fd_tk.core.common.errors import GenerateCsvPermissionErr
from ascend_fd_tk.core.common.path import CommonPath, ConfigPath


class TestErrors(unittest.TestCase):
    def test_generate_csv_permission_err_is_exception(self):
        """GenerateCsvPermissionErr 应继承 Exception。"""
        self.assertTrue(issubclass(GenerateCsvPermissionErr, Exception))
        with self.assertRaises(GenerateCsvPermissionErr):
            raise GenerateCsvPermissionErr("no permission")


class TestConstants(unittest.TestCase):
    def test_key_constants(self):
        """关键常量应符合预期。"""
        self.assertEqual(PACKAGE_NAME, "ascend-faultdiag-toolkit")
        self.assertEqual(NPU_LINK_DOWN, "LINK DOWN")
        self.assertEqual(NPU_LONG_DOWN_TIME, 28)
        self.assertEqual((FAULT_TYPE_BMC, FAULT_TYPE_HOST, FAULT_TYPE_SWITCH), ("bmc", "host", "switch"))


class TestPath(unittest.TestCase):
    def test_common_path_structure(self):
        """CommonPath 关键路径应基于 TOOL_HOME / ROOT_DIR 构建。"""
        self.assertTrue(os.path.isabs(CommonPath.ROOT_DIR))
        self.assertTrue(CommonPath.TOOL_HOME.endswith(".ascend-faultdiag-toolkit"))
        self.assertTrue(CommonPath.LOG_DIR.startswith(CommonPath.TOOL_HOME))
        self.assertTrue(CommonPath.REPORT_DIR.startswith(CommonPath.TOOL_HOME))

    def test_config_path_locations(self):
        """ConfigPath 应定位到配置目录下的 JSON 文件。"""
        self.assertTrue(ConfigPath.L1_GLOBAL_ADDR_CONFIG_PATH.startswith(ConfigPath.CONFIG_DIR))
        self.assertEqual(ConfigPath.CONFIG_DIR, os.path.join(CommonPath.ROOT_DIR, "core", "config"))


if __name__ == "__main__":
    unittest.main()
