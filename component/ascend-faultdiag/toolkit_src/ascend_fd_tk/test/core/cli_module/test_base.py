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

"""CLI 基础框架单元测试：CliCtx 命令分发与 CliModel 参数/路径校验。"""

import os
import tempfile
import unittest
from unittest import mock

from ascend_fd_tk.core.cli_module.base import CliCtx, CliModel


class _FakeCliModel(CliModel):
    @classmethod
    def get_key(cls):
        return "fake"

    def get_help(self):
        return "fake help"

    def run_task(self, *args):
        return f"ran:{','.join(args)}"


class _ParamCliModel(_FakeCliModel):
    @classmethod
    def get_key(cls):
        return "param"

    @staticmethod
    def is_support_param():
        return True


class TestCliCtx(unittest.TestCase):
    def setUp(self):
        self.diag_ctx = mock.MagicMock()
        self.cli_ctx = CliCtx()
        self.cli_ctx.update_cli_models(
            [_FakeCliModel(self.diag_ctx, self.cli_ctx), _ParamCliModel(self.diag_ctx, self.cli_ctx)]
        )

    def test_is_cmd_valid(self):
        self.assertTrue(self.cli_ctx.is_cmd_valid(" fake "))
        self.assertFalse(self.cli_ctx.is_cmd_valid("unknown"))

    def test_run_cmd_unknown(self):
        self.assertEqual(self.cli_ctx.run_cmd("unknown"), "")

    def test_run_cmd_no_param_model(self):
        self.assertEqual(self.cli_ctx.run_cmd("fake", "x"), "Unsupported arguments: x")
        self.assertEqual(self.cli_ctx.run_cmd("fake", "?"), "fake help")
        self.assertEqual(self.cli_ctx.run_cmd("fake"), "ran:")

    def test_run_cmd_param_model(self):
        self.assertEqual(self.cli_ctx.run_cmd("param", "a", "b"), "Unsupported arguments: b")
        self.assertEqual(self.cli_ctx.run_cmd("param", "a"), "ran:a")

    def test_show_help(self):
        self.assertEqual(self.cli_ctx.show_help(), "")
        self.cli_ctx.update_cli_models([_FakeCliModel(self.diag_ctx, self.cli_ctx)])
        self.cli_ctx.cli_model_map["help"] = _FakeCliModel(self.diag_ctx, self.cli_ctx)
        self.assertIn("ran:", self.cli_ctx.show_help())


class TestCliModel(unittest.TestCase):
    def test_max_param_count(self):
        self.assertEqual(CliModel.max_param_count(), 0)
        self.assertEqual(_ParamCliModel.max_param_count(), 1)

    def test_run_filters_and_help(self):
        model = _FakeCliModel(mock.MagicMock(), CliCtx())
        self.assertEqual(model.run("", "a"), "ran:a")
        self.assertEqual(model.run("?"), "fake help")
        self.assertEqual(model.run("？"), "fake help")
        self.assertEqual(model.run(), "ran:")
        self.assertEqual(model.get_detail(), "fake help")

    def test_check_input_path(self):
        self.assertEqual(CliModel.check_input_path(), "地址为空，请重新设置")
        self.assertIn("不存在", CliModel.check_input_path(os.path.join(tempfile.gettempdir(), "no_such_dir_xyz")))
        with tempfile.TemporaryDirectory() as tmp:
            file_path = os.path.join(tmp, "f.txt")
            with open(file_path, "w", encoding="utf-8") as f:
                f.write("x")
            self.assertIn("非文件夹", CliModel.check_input_path(file_path))
            self.assertEqual(CliModel.check_input_path(tmp), "")


if __name__ == "__main__":
    unittest.main()
