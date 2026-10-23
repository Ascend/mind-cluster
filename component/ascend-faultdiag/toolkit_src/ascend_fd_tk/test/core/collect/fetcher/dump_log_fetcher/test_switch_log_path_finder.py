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

"""交换机日志路径查找器单元测试：定位 diag info 目录与过滤回显 TXT。"""

import os
import tempfile
import unittest
from pathlib import Path

from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.switch_log_path_finder import SwitchLogPathFinder


class TestSwitchLogPathFinder(unittest.TestCase):
    def test_find_diag_info_and_cli_txt(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = os.path.join(tmp, "dump")
            slot = os.path.join(root, "node", "slot_1_0")
            os.makedirs(slot)
            cli_dir = os.path.join(root, "cli_files")
            os.makedirs(cli_dir)
            with open(os.path.join(cli_dir, "cli_output.txt"), "w", encoding="utf-8") as f:
                f.write("content")
            with open(os.path.join(cli_dir, "diag.zip"), "w", encoding="utf-8") as f:
                f.write("zip")
            diag_info_dirs, cli_txt_paths = SwitchLogPathFinder.find(root)
            self.assertEqual(diag_info_dirs, [str(Path(os.path.dirname(slot)).resolve())])
            self.assertEqual(len(cli_txt_paths), 1)
            self.assertTrue(cli_txt_paths[0].endswith("cli_output.txt"))

    def test_find_cli_output_txt_excludes_folders(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = os.path.join(tmp, "dump")
            exclude = str(Path(os.path.join(root, "slot_1_0")).resolve())
            keep = os.path.join(root, "other")
            os.makedirs(os.path.join(exclude, "sub"))
            os.makedirs(keep)
            with open(os.path.join(exclude, "sub", "skip.txt"), "w", encoding="utf-8") as f:
                f.write("x")
            with open(os.path.join(keep, "keep.txt"), "w", encoding="utf-8") as f:
                f.write("x")
            with open(os.path.join(keep, "keep.dat"), "w", encoding="utf-8") as f:
                f.write("x")
            result = SwitchLogPathFinder._find_cli_output_txt(root, [exclude])
            self.assertEqual(len(result), 1)
            self.assertTrue(result[0].endswith("keep.txt"))


if __name__ == "__main__":
    unittest.main()
