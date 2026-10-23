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

"""host_registry 单元测试：HostInfo 代际注册表。"""

import unittest

from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.core.context.host_registry import HOST_INFO_REGISTRY, register_host_info


class _DummyHostInfo(JsonObj):
    def __init__(self, name=""):
        self.name = name


class TestHostRegistry(unittest.TestCase):
    def tearDown(self):
        if "TEST" in HOST_INFO_REGISTRY:
            del HOST_INFO_REGISTRY["TEST"]

    def test_register_host_info(self):
        register_host_info("TEST")(_DummyHostInfo)
        self.assertIs(HOST_INFO_REGISTRY["TEST"], _DummyHostInfo)


if __name__ == "__main__":
    unittest.main()
