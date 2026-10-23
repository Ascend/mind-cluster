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

"""diag_ctx 单元测试：诊断上下文默认状态与资源清理。"""

import tempfile
import unittest

from ascend_fd_tk.core.config.dump_log_dir_config import DumpLogDirConfig
from ascend_fd_tk.core.config.location_config import LocationConfig
from ascend_fd_tk.core.context.diag_ctx import DiagCtx
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache


class TestDiagCtx(unittest.TestCase):
    def test_initialization_defaults(self):
        ctx = DiagCtx()
        try:
            self.assertIsNone(ctx.conn_config)
            self.assertIsInstance(ctx.dump_log_dir_config, DumpLogDirConfig)
            self.assertIsInstance(ctx.location_config, LocationConfig)
            self.assertIsInstance(ctx.cache, ClusterInfoCache)
            self.assertEqual(ctx.diag_result, [])
            self.assertEqual(ctx.inspection_result, [])
            self.assertIsNotNone(ctx.process_pool)
        finally:
            ctx.close()

    def test_close_shuts_down_process_pool(self):
        ctx = DiagCtx()
        ctx.close()
        ctx.close()

    def test_read_from_dir_without_config(self):
        ctx = DiagCtx()
        try:
            with tempfile.TemporaryDirectory() as tmp:
                ctx.read_from_dir(tmp)
            self.assertIsInstance(ctx.location_config, LocationConfig)
            self.assertEqual(ctx.threshold_loader._overrides, {})
        finally:
            ctx.close()


if __name__ == "__main__":
    unittest.main()
