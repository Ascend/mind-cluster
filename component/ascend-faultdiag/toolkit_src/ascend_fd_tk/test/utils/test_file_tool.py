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

"""utils.file_tool 文件安全读写与路径工具单元测试。"""

import json
import os
import platform
import tempfile
import unittest
from pathlib import Path
from unittest import mock

from ascend_fd_tk.utils import file_tool


# pylint: disable=R0801
class TestSafeReadOpen(unittest.TestCase):
    def test_open_normal_file(self):
        with tempfile.NamedTemporaryFile(delete=False, suffix=".txt") as f:
            f.write(b"hello world")
            tmp = f.name
        try:
            with file_tool.safe_read_open(tmp, "r", encoding="utf-8") as stream:
                self.assertEqual(stream.read(), "hello world")
        finally:
            os.unlink(tmp)

    def test_open_symbolic_link_raises(self):
        with mock.patch("ascend_fd_tk.utils.file_tool.os.path.islink", return_value=True):
            # pylint: disable=R0801
            with self.assertRaises(Exception) as ctx:
                file_tool.safe_read_open("dummy.txt")
            self.assertIn("symbolic link", str(ctx.exception))

    def test_open_oversize_file_raises(self):
        with tempfile.NamedTemporaryFile(delete=False, suffix=".txt") as f:
            f.write(b"data")
            tmp = f.name
        try:
            with mock.patch("ascend_fd_tk.utils.file_tool.os.stat") as mock_stat:
                mock_stat.return_value.st_size = file_tool.MAX_SIZE + 1
                with self.assertRaises(Exception) as ctx:
                    # pylint: disable=R0801
                    file_tool.safe_read_open(tmp, "r", encoding="utf-8")
                self.assertIn("less than", str(ctx.exception))
        finally:
            os.unlink(tmp)

    def test_open_text_mode_invalid_utf8_replaced(self):
        """文本模式默认 errors='replace'，非法字节被替换为占位符而非抛异常"""
        with tempfile.NamedTemporaryFile(delete=False, suffix=".txt") as f:
            f.write(b"ok\xff\xfeend")
            tmp = f.name
        try:
            # pylint: disable=R0801
            with file_tool.safe_read_open(tmp, "r", encoding="utf-8") as stream:
                content = stream.read()
            self.assertTrue(content.startswith("ok"))
            self.assertTrue(content.endswith("end"))
            self.assertIn("\ufffd", content)
        finally:
            os.unlink(tmp)

    def test_open_text_mode_explicit_errors_not_overridden(self):
        """显式传入 errors='strict' 时不应被默认值覆盖，非法字节抛 UnicodeDecodeError"""
        with tempfile.NamedTemporaryFile(delete=False, suffix=".txt") as f:
            f.write(b"\xff\xfe")
            tmp = f.name
        try:
            with self.assertRaises(UnicodeDecodeError):
                with file_tool.safe_read_open(tmp, "r", encoding="utf-8", errors="strict") as stream:
                    stream.read()
        finally:
            os.unlink(tmp)

    def test_open_binary_mode_returns_bytes(self):
        """二进制模式读取返回 bytes，不注入 errors 参数"""
        with tempfile.NamedTemporaryFile(delete=False, suffix=".bin") as f:
            # pylint: disable=R0801
            f.write(b"\xff\xfe\x00data")
            tmp = f.name
        try:
            # pylint: disable=R0801
            with file_tool.safe_read_open(tmp, "rb") as stream:
                self.assertEqual(stream.read(), b"\xff\xfe\x00data")
        finally:
            os.unlink(tmp)

    def test_open_default_mode_is_text(self):
        """不传 mode 时默认按文本 'r' 打开，非法字节被替换"""
        with tempfile.NamedTemporaryFile(delete=False, suffix=".txt") as f:
            f.write(b"abc\xff")
            tmp = f.name
        try:
            # pylint: disable=R0801
            with file_tool.safe_read_open(tmp, encoding="utf-8") as stream:
                content = stream.read()
            self.assertTrue(content.startswith("abc"))
            self.assertIn("\ufffd", content)
        finally:
            os.unlink(tmp)


class TestSafeReadJson(unittest.TestCase):
    def test_valid_json(self):
        with tempfile.NamedTemporaryFile(delete=False, suffix=".json", mode="w", encoding="utf-8") as f:
            json.dump({"a": 1}, f)
            tmp = f.name
        try:
            self.assertEqual(file_tool.safe_read_json(tmp), {"a": 1})
        finally:
            os.unlink(tmp)

    def test_invalid_json_returns_empty(self):
        with tempfile.NamedTemporaryFile(delete=False, suffix=".json", mode="w", encoding="utf-8") as f:
            f.write("not json")
            tmp = f.name
        try:
            self.assertEqual(file_tool.safe_read_json(tmp), {})
        finally:
            os.unlink(tmp)


class TestFindAllSubPaths(unittest.TestCase):
    def test_find_multi_depth(self):
        with tempfile.TemporaryDirectory() as tmp:
            Path(tmp, "collect").mkdir()
            Path(tmp, "a", "collect").mkdir(parents=True)
            result = file_tool.find_all_sub_paths(tmp, "collect", max_depth=2)
            self.assertEqual(len(result), 2)

    def test_invalid_root_returns_empty(self):
        with mock.patch("ascend_fd_tk.utils.file_tool._CONSOLE_LOGGER"):
            result = file_tool.find_all_sub_paths(os.path.join(tempfile.gettempdir(), "no_such_dir_xyz"), "collect")
            self.assertEqual(result, [])


class TestJumpUpDirectories(unittest.TestCase):
    @staticmethod
    def _resolve(path):
        return str(Path(path).resolve())

    def test_jump_up(self):
        current = os.path.join(tempfile.gettempdir(), "a", "b", "c")
        self.assertEqual(
            file_tool.jump_up_directories(current), self._resolve(os.path.join(tempfile.gettempdir(), "a", "b"))
        )
        self.assertEqual(
            file_tool.jump_up_directories(current, levels=2), self._resolve(os.path.join(tempfile.gettempdir(), "a"))
        )
        self.assertEqual(file_tool.jump_up_directories("", 1), "")
        self.assertEqual(
            file_tool.jump_up_directories(os.path.join(tempfile.gettempdir(), "a"), 0),
            os.path.join(tempfile.gettempdir(), "a"),
        )


class TestConvertLogPath(unittest.TestCase):
    def test_existing_path(self):
        with tempfile.NamedTemporaryFile(delete=False, suffix=".txt") as f:
            tmp = f.name
        try:
            result = file_tool.convert_log_path(tmp)
            self.assertTrue(result)
            if platform.system().lower() != "windows":
                self.assertEqual(result, os.path.abspath(tmp))
        finally:
            os.unlink(tmp)

    def test_missing_path_returns_empty(self):
        self.assertEqual(file_tool.convert_log_path(os.path.join(tempfile.gettempdir(), "no_such_file_xyz.txt")), "")


if __name__ == "__main__":
    unittest.main()
