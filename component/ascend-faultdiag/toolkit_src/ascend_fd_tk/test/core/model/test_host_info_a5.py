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

"""HostInfoA5 单元测试。

覆盖本次改动：
- @register_host_info("A5") 代际注册
- chip_generation 默认 "A5"
- 与 A3 完全独立（不继承 HostInfo）
- get_msn_logs_by_type 过滤
"""

import unittest

from ascend_fd_tk.core.context.host_registry import HOST_INFO_REGISTRY
from ascend_fd_tk.core.model.host import HostInfo
from ascend_fd_tk.core.model.host_a5 import HostInfoA5
from ascend_fd_tk.core.log_parser.base import FindResult


def _make_find_result(pattern_key: str) -> FindResult:
    return FindResult(pattern_key=pattern_key, logline="dummy", log_path="/tmp/dummy")


class TestHostInfoA5Registration(unittest.TestCase):
    def test_generation_registration(self):
        """HostInfoA5 应注册到 HOST_INFO_REGISTRY["A5"]。"""
        self.assertIs(HOST_INFO_REGISTRY["A5"], HostInfoA5)

    def test_not_inherit_a3(self):
        """HostInfoA5 与 A3 HostInfo 完全独立，不继承。"""
        self.assertFalse(issubclass(HostInfoA5, HostInfo))


class TestHostInfoA5Defaults(unittest.TestCase):
    def test_default_chip_generation(self):
        """HostInfoA5 默认 chip_generation 为 'A5'。"""
        info = HostInfoA5(host_id="h1", sn_num="SN1")
        self.assertEqual(info.chip_generation, "A5")
        # 可选字段默认值
        self.assertEqual(info.hostname, "")
        self.assertEqual(info.npu_chip_info, {})
        self.assertEqual(info.nic_info_list, [])
        self.assertEqual(info.msnpureport_log, [])

    def test_get_msn_logs_by_type(self):
        """get_msn_logs_by_type 按 pattern_key 过滤 msnpureport_log。"""
        info = HostInfoA5(
            host_id="h1",
            sn_num="SN1",
            msnpureport_log=[
                _make_find_result("error"),
                _make_find_result("warning"),
                _make_find_result("error"),
            ],
        )
        errors = info.get_msn_logs_by_type("error")
        self.assertEqual(len(errors), 2)
        self.assertEqual(info.get_msn_logs_by_type("nonexistent"), [])
        self.assertEqual(info.get_msn_logs_by_type(""), [])


if __name__ == "__main__":
    unittest.main()
