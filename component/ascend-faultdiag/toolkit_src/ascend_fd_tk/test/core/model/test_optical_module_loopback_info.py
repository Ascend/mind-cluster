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

"""光模块环回信息数据模型单元测试：十六进制能力与控制项解析。"""

import unittest

from ascend_fd_tk.core.model.optical_module_loopback_info import OpticalModuleLoopbackInfo


class TestOpticalModuleLoopbackInfo(unittest.TestCase):
    def test_parse_hex_capabilities_and_controls(self):
        info = OpticalModuleLoopbackInfo(
            npu_id="0",
            loopback_capabilities="1A",
            loopback_controls=["1", "0x2", "ff"],
        )
        self.assertEqual(info.loopback_capabilities, 26)
        self.assertEqual(info.loopback_controls, [1, 2, 255])

    def test_default_and_invalid_hex(self):
        info = OpticalModuleLoopbackInfo()
        self.assertEqual(info.loopback_capabilities, 0)
        self.assertEqual(info.loopback_controls, [])
        info_invalid = OpticalModuleLoopbackInfo(loopback_capabilities="xyz", loopback_controls=["xyz"])
        self.assertEqual(info_invalid.loopback_capabilities, 0)
        self.assertEqual(info_invalid.loopback_controls, [0])


if __name__ == "__main__":
    unittest.main()
