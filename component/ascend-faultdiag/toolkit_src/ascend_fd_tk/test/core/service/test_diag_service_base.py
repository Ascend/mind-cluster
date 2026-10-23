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

"""DiagService 抽象基类测试"""

import unittest
from abc import ABC

from ascend_fd_tk.core.service.base import DiagService


class TestDiagService(unittest.TestCase):
    def test_is_abstract(self):
        self.assertTrue(issubclass(DiagService, ABC))
        with self.assertRaises(TypeError):
            # pylint: disable=abstract-class-instantiated
            DiagService(object)

    def test_init_stores_diag_ctx(self):
        class ConcreteDiagService(DiagService):
            async def run(self):
                pass

        diag_ctx = object()
        service = ConcreteDiagService(diag_ctx)
        self.assertIs(service.diag_ctx, diag_ctx)


if __name__ == "__main__":
    unittest.main()
