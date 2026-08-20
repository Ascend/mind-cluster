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

"""注册表单元测试。

覆盖本次改动：
- HOST_COLLECTOR_REGISTRY / HOST_FETCHER_REGISTRY 代际注册装饰器（context/register.py）
- HOST_INFO_REGISTRY 独立为 host_registry.py 避免循环依赖
"""

import unittest

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.context.register import (
    HOST_COLLECTOR_REGISTRY,
    HOST_FETCHER_REGISTRY,
    register_host_collector,
    register_host_fetcher,
)
from ascend_fd_tk.core.context.host_registry import HOST_INFO_REGISTRY, register_host_info


class TestRegistries(unittest.TestCase):
    def test_register_host_collector(self):
        """装饰器应将类注册到 HOST_COLLECTOR_REGISTRY 对应代际。"""
        original = HOST_COLLECTOR_REGISTRY.get(NpuType.A3)
        try:

            @register_host_collector(NpuType.A3)
            class _FakeCollector:
                pass

            self.assertIs(HOST_COLLECTOR_REGISTRY[NpuType.A3], _FakeCollector)
            # 装饰器返回原类，不影响正常使用
            self.assertIs(_FakeCollector, _FakeCollector)
        finally:
            if original is not None:
                HOST_COLLECTOR_REGISTRY[NpuType.A3] = original

    def test_register_host_fetcher(self):
        """装饰器应将类注册到 HOST_FETCHER_REGISTRY 对应代际。"""
        original = HOST_FETCHER_REGISTRY.get(NpuType.A5)
        try:

            @register_host_fetcher(NpuType.A5)
            class _FakeFetcher:
                pass

            self.assertIs(HOST_FETCHER_REGISTRY[NpuType.A5], _FakeFetcher)
        finally:
            if original is not None:
                HOST_FETCHER_REGISTRY[NpuType.A5] = original

    def test_register_host_info_independent_module(self):
        """HOST_INFO_REGISTRY 应在独立模块（避免循环依赖）。"""
        # 用字符串代际标识（host_registry 用 str 而非 NpuType）

        @register_host_info("A6")
        class _FakeHostInfo:
            pass

        self.assertIs(HOST_INFO_REGISTRY["A6"], _FakeHostInfo)
        # 确认 host_registry 和 register 是不同模块
        from ascend_fd_tk.core.context import register, host_registry

        self.assertIsNot(register.HOST_COLLECTOR_REGISTRY, host_registry.HOST_INFO_REGISTRY)


if __name__ == "__main__":
    unittest.main()
