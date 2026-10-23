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

"""utils.compress_tool 递归解压工具单元测试。"""

import io
import os
import tarfile
import tempfile
import unittest
import zipfile

from ascend_fd_tk.utils.compress_tool import CompressTool


class TestExtractZipRecursive(unittest.TestCase):
    def test_extract_simple_zip(self):
        with tempfile.TemporaryDirectory() as tmp:
            zip_path = os.path.join(tmp, "a.zip")
            with zipfile.ZipFile(zip_path, "w") as zf:
                zf.writestr("inner.txt", "data")
            result = CompressTool.extract_zip_recursive(zip_path, max_depth=1)
            self.assertEqual(result, 0)
            self.assertTrue(os.path.exists(os.path.join(tmp, "a", "inner.txt")))

    def test_extract_target_already_exists(self):
        with tempfile.TemporaryDirectory() as tmp:
            zip_path = os.path.join(tmp, "a.zip")
            os.makedirs(os.path.join(tmp, "a"))
            self.assertEqual(CompressTool.extract_zip_recursive(zip_path, max_depth=1), 0)

    def test_invalid_zip_returns_zero(self):
        with tempfile.TemporaryDirectory() as tmp:
            zip_path = os.path.join(tmp, "bad.zip")
            with open(zip_path, "w", encoding="utf-8") as f:
                f.write("not a zip")
            self.assertEqual(CompressTool.extract_zip_recursive(zip_path, max_depth=1), 0)


class TestExtractTarGz(unittest.TestCase):
    def test_extract_tar_gz(self):
        with tempfile.TemporaryDirectory() as tmp:
            tar_path = os.path.join(tmp, "a.tar.gz")
            data = b"content"
            info = tarfile.TarInfo("file.txt")
            info.size = len(data)
            with tarfile.open(tar_path, "w:gz") as tar:
                tar.addfile(info, io.BytesIO(data))
            result = CompressTool.extract_tar_gz(tar_path, tmp)
            self.assertTrue(result)
            self.assertTrue(os.path.exists(os.path.join(tmp, "a", "file.txt")))

    def test_invalid_tar_gz_returns_false(self):
        with tempfile.TemporaryDirectory() as tmp:
            tar_path = os.path.join(tmp, "bad.tar.gz")
            with open(tar_path, "w", encoding="utf-8") as f:
                f.write("bad")
            self.assertFalse(CompressTool.extract_tar_gz(tar_path, tmp))


if __name__ == "__main__":
    unittest.main()
