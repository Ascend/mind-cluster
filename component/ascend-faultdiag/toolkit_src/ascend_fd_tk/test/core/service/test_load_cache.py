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

"""LoadCache 缓存加载服务测试"""

import json
import os
import tempfile
import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.model.host import HostInfo
from ascend_fd_tk.core.model.host_a5 import HostInfoA5
from ascend_fd_tk.core.service.load_cache import LoadCache


class TestResolveHostClass(unittest.TestCase):
    def test_a3_resolves_host_info(self):
        self.assertIs(LoadCache._resolve_host_class({"generation": "A3"}), HostInfo)

    def test_a5_resolves_host_info_a5(self):
        self.assertIs(LoadCache._resolve_host_class({"generation": "A5"}), HostInfoA5)

    def test_missing_generation_defaults_a3(self):
        self.assertIs(LoadCache._resolve_host_class({}), HostInfo)


class TestLoadCacheFile(unittest.TestCase):
    def test_load_cache_reads_json_files(self):
        with tempfile.TemporaryDirectory() as tmp_dir:
            with open(os.path.join(tmp_dir, "1.1.1.1.json"), "w", encoding="utf-8") as f:
                json.dump({"host_id": "1.1.1.1", "sn_num": "SN1"}, f)
            cache_map = {}
            LoadCache._load_cache(tmp_dir, HostInfo, cache_map)
            self.assertIn("1.1.1.1", cache_map)
            self.assertEqual(cache_map["1.1.1.1"].host_id, "1.1.1.1")

    def test_load_cache_skips_non_json_and_empty(self):
        with tempfile.TemporaryDirectory() as tmp_dir:
            with open(os.path.join(tmp_dir, "a.txt"), "w", encoding="utf-8") as f:
                f.write("not json")
            with open(os.path.join(tmp_dir, "empty.json"), "w", encoding="utf-8") as f:
                f.write("")
            cache_map = {}
            LoadCache._load_cache(tmp_dir, HostInfo, cache_map)
            self.assertEqual(cache_map, {})


class TestLoadCacheRun(unittest.IsolatedAsyncioTestCase):
    async def test_run_loads_and_sets_generation(self):
        with tempfile.TemporaryDirectory() as tmp_dir:
            host_json = {"host_id": "1.1.1.1", "sn_num": "SN1", "generation": "A5"}
            with open(os.path.join(tmp_dir, "1.1.1.1.json"), "w", encoding="utf-8") as f:
                json.dump(host_json, f)
            cache = MagicMock()
            cache.hosts_info = {}
            cache.bmcs_info = {}
            cache.swis_info = {}
            ctx = MagicMock()
            ctx.cache = cache
            with (
                patch("ascend_fd_tk.core.common.path.CommonPath.COLLECT_BMC_CACHE_DIR", tmp_dir),
                patch("ascend_fd_tk.core.common.path.CommonPath.COLLECT_SWITCH_CACHE_DIR", tmp_dir),
                patch("ascend_fd_tk.core.common.path.CommonPath.COLLECT_HOST_CACHE_DIR", tmp_dir),
            ):
                await LoadCache(ctx).run()
            cache.sort_info.assert_called_once()
            cache.init_diag_data.assert_called_once()
            self.assertEqual(cache.generation, "A5")


if __name__ == "__main__":
    unittest.main()
