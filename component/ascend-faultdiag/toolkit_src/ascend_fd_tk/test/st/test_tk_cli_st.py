#!/usr/bin/env python3
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

import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

TOOLKIT_SRC = Path(__file__).resolve().parents[3]


class TestToolkitOfflineCliST(unittest.TestCase):
    """Process-level ST for toolkit interactive/offline commands."""

    def setUp(self):
        self.tmp_home = tempfile.mkdtemp(prefix="tk_home_")
        self.workspace = tempfile.mkdtemp(prefix="tk_workspace_")

    def tearDown(self):
        shutil.rmtree(self.tmp_home, ignore_errors=True)
        shutil.rmtree(self.workspace, ignore_errors=True)

    def run_cli(self, *args, timeout=180):
        env = os.environ.copy()
        env["HOME"] = self.tmp_home
        env["PYTHONPATH"] = str(TOOLKIT_SRC) + os.pathsep + env.get("PYTHONPATH", "")
        result = subprocess.run(
            [sys.executable, "-m", "ascend_fd_tk.cli", *args],
            cwd=self.workspace,
            env=env,
            capture_output=True,
            text=True,
            timeout=timeout,
            check=False,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn("Traceback", result.stderr)
        return result.stdout + result.stderr

    def make_dump_dir(self, name):
        dump_dir = os.path.join(self.workspace, name)
        os.makedirs(dump_dir, exist_ok=True)
        return dump_dir

    def test_guide_mentions_required_commands(self):
        output = self.run_cli("guide")
        for command in ("set_conn_config", "set_host_dump_log", "auto_collect_diag", "clear_cache"):
            self.assertIn(command, output)

    def test_set_dump_log_invalid_and_valid(self):
        output = self.run_cli("set_host_dump_log", "no_such_dir_xyz")
        self.assertIn("不存在", output)

        dump_dir = self.make_dump_dir("host_dump_log")
        output = self.run_cli("set_host_dump_log", dump_dir)
        self.assertIn("设置成功", output)

    def test_offline_auto_collect_and_diag(self):
        dump_dir = self.make_dump_dir("host_dump_log")
        self.assertIn("设置成功", self.run_cli("set_host_dump_log", dump_dir))
        self.assertIn("收集完成", self.run_cli("auto_collect"))
        self.assertIn("诊断完成", self.run_cli("auto_collect_diag"))

    def test_auto_collect_diag_twice_without_clear_cache(self):
        dump_dir = self.make_dump_dir("host_dump_log")
        self.run_cli("set_host_dump_log", dump_dir)
        self.assertIn("诊断完成", self.run_cli("auto_collect_diag"))
        self.assertIn("诊断完成", self.run_cli("auto_collect_diag"))

    def test_batch_collect_then_unified_diag(self):
        first_dir = self.make_dump_dir("batch_1")
        second_dir = self.make_dump_dir("batch_2")

        self.run_cli("set_host_dump_log", first_dir)
        self.assertIn("收集完成", self.run_cli("auto_collect"))
        self.run_cli("set_host_dump_log", second_dir)
        self.assertIn("收集完成", self.run_cli("auto_collect"))

        self.assertIn("诊断完成", self.run_cli("auto_diag"))

    def test_clear_cache_after_diag(self):
        dump_dir = self.make_dump_dir("host_dump_log")
        self.run_cli("set_host_dump_log", dump_dir)
        self.run_cli("auto_collect_diag")
        self.assertIn("清理完成", self.run_cli("clear_cache"))


if __name__ == "__main__":
    unittest.main()
