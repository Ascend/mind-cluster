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

"""OutputCache 缓存落盘服务测试"""

import os
import tempfile
import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.model.host import HostInfo
from ascend_fd_tk.core.service.output_cache import OutputCache


class TestOutputCacheFile(unittest.TestCase):
    def test_output_cache_writes_json_files(self):
        with tempfile.TemporaryDirectory() as tmp_dir:
            host = HostInfo("1.1.1.1", "SN1")
            cache_map = {"1.1.1.1": host}
            OutputCache._output_cache(tmp_dir, cache_map)
            file_path = os.path.join(tmp_dir, "1.1.1.1.json")
            self.assertTrue(os.path.exists(file_path))
            with open(file_path, "r", encoding="utf-8") as f:
                content = f.read()
            self.assertIn("1.1.1.1", content)

    def test_output_cache_creates_dir(self):
        with tempfile.TemporaryDirectory() as tmp_dir:
            target = os.path.join(tmp_dir, "sub", "dir")
            OutputCache._output_cache(target, {})
            self.assertTrue(os.path.exists(target))


class TestOutputCacheRun(unittest.IsolatedAsyncioTestCase):
    async def test_run_outputs_all_caches(self):
        ctx = MagicMock()
        ctx.cache = MagicMock()
        ctx.cache.bmcs_info = {}
        ctx.cache.swis_info = {}
        ctx.cache.hosts_info = {}
        with (
            tempfile.TemporaryDirectory() as tmp_dir,
            patch("ascend_fd_tk.core.common.path.CommonPath.COLLECT_BMC_CACHE_DIR", tmp_dir),
            patch("ascend_fd_tk.core.common.path.CommonPath.COLLECT_SWITCH_CACHE_DIR", tmp_dir),
            patch("ascend_fd_tk.core.common.path.CommonPath.COLLECT_HOST_CACHE_DIR", tmp_dir),
            patch.object(OutputCache, "_output_cache") as mock_output,
        ):
            await OutputCache(ctx).run()
            self.assertEqual(mock_output.call_count, 3)


if __name__ == "__main__":
    unittest.main()
