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

"""CLI 命令模型单元测试。"""

import os
import tempfile
import unittest
from unittest import mock

from ascend_fd_tk.core.cli_module.base import CliCtx
from ascend_fd_tk.core.cli_module.cli_model import (
    AboutCliModel,
    AutoDiagCliModel,
    AutoInspection,
    AutoSingleDiagCliModel,
    ClearCacheCliModel,
    ClearCliModel,
    ExitCliModel,
    GenerateCsvPermissionErr,
    HelpCliModel,
    SetBmcDumpLogDirCliModel,
    SetConfigDirCliModel,
    SetConnConfigCliModel,
    SetHostDumpLogDirCliModel,
    SetSwiDumpLogDirCliModel,
)
from ascend_fd_tk.utils.file_tool import convert_log_path


class _CliTestBase(unittest.TestCase):
    def setUp(self):
        self.diag_ctx = mock.MagicMock()
        self.cli_ctx = CliCtx()
        self.cli_ctx.cli_model_map = {
            "help": HelpCliModel(self.diag_ctx, self.cli_ctx),
            "exit": ExitCliModel(self.diag_ctx, self.cli_ctx),
            "set_config_dir": SetConfigDirCliModel(self.diag_ctx, self.cli_ctx),
        }


class TestBasicCommands(_CliTestBase):
    def test_basic_commands(self):
        self.assertIn("显示帮助信息", HelpCliModel(self.diag_ctx, self.cli_ctx).run_task())
        exit_model = ExitCliModel(self.diag_ctx, self.cli_ctx)
        self.assertEqual(exit_model.run_task(), "再见!")
        self.assertFalse(self.cli_ctx.is_running)
        with mock.patch("ascend_fd_tk.core.cli_module.cli_model.os.system") as mock_system:
            self.assertEqual(ClearCliModel(self.diag_ctx, self.cli_ctx).run_task(), "")
            mock_system.assert_called_once()
        self.diag_ctx.tool_config.version = "1.0.0"
        self.assertIn("1.0.0", AboutCliModel(self.diag_ctx, self.cli_ctx).run_task())


class TestSetConfigDir(_CliTestBase):
    def test_check_input_path(self):
        self.assertEqual(SetConfigDirCliModel.check_input_path(), "目录路径为空，请重新设置")
        self.assertIn("不存在", SetConfigDirCliModel.check_input_path("no_such_dir_xyz"))
        with tempfile.TemporaryDirectory() as tmp:
            file_path = os.path.join(tmp, "f.txt")
            with open(file_path, "w", encoding="utf-8") as f:
                f.write("x")
            self.assertIn("不是目录", SetConfigDirCliModel.check_input_path(file_path))
            self.assertEqual(SetConfigDirCliModel.check_input_path(tmp), "")

    def test_run_task(self):
        with tempfile.TemporaryDirectory() as tmp:
            model = SetConfigDirCliModel(self.diag_ctx, self.cli_ctx)
            self.assertIn("设置成功", model.run_task(tmp))
            self.diag_ctx.read_from_dir.assert_called_once_with(convert_log_path(tmp))
            self.assertIn("不存在", model.run_task("no_such_dir_xyz"))


class TestSetConnConfig(_CliTestBase):
    def test_check_input_path(self):
        self.assertEqual(SetConnConfigCliModel.check_input_path(), "地址为空，请重新设置")
        self.assertIn("不存在", SetConnConfigCliModel.check_input_path("no_such_file_xyz"))
        with tempfile.TemporaryDirectory() as tmp:
            dir_path = os.path.join(tmp, "subdir")
            os.mkdir(dir_path)
            self.assertIn("不是文件", SetConnConfigCliModel.check_input_path(dir_path))
            file_path = os.path.join(tmp, "conn.ini")
            with open(file_path, "w", encoding="utf-8") as f:
                f.write("x")
            self.assertEqual(SetConnConfigCliModel.check_input_path(file_path), "")

    def test_run_task(self):
        model = SetConnConfigCliModel(self.diag_ctx, self.cli_ctx)
        with tempfile.TemporaryDirectory() as tmp:
            file_path = os.path.join(tmp, "conn.ini")
            with open(file_path, "w", encoding="utf-8") as f:
                f.write("x")
            self.diag_ctx.load_conn_config.return_value = ""
            self.assertIn("设置成功", model.run_task(file_path))
            self.diag_ctx.encrypt_conn_config.assert_called_once_with(file_path)
            self.diag_ctx.load_conn_config.return_value = "boom"
            self.assertIn("boom", model.run_task(file_path))


class TestSetDumpLogDir(_CliTestBase):
    def test_set_dump_log_dirs(self):
        with tempfile.TemporaryDirectory() as tmp:
            SetHostDumpLogDirCliModel(self.diag_ctx, self.cli_ctx).run_task(tmp)
            self.assertEqual(self.diag_ctx.dump_log_dir_config.host_dump_log_dir, tmp)
            SetBmcDumpLogDirCliModel(self.diag_ctx, self.cli_ctx).run_task(tmp)
            self.assertEqual(self.diag_ctx.dump_log_dir_config.bmc_dump_log_dir, tmp)
            SetSwiDumpLogDirCliModel(self.diag_ctx, self.cli_ctx).run_task(tmp)
            self.assertEqual(self.diag_ctx.dump_log_dir_config.switch_dump_log_dir, tmp)


class TestAutoCommands(_CliTestBase):
    def test_auto_inspection(self):
        model = AutoInspection(self.diag_ctx, self.cli_ctx)
        with mock.patch("ascend_fd_tk.core.cli_module.cli_model.asyncio.run") as mock_run:
            self.assertEqual(model.run_task(), "巡检完成")
            mock_run.assert_called_once()
        self.assertIn("不支持的客户类型", model.run_task("not_a_customer"))

    def test_auto_diag(self):
        model = AutoDiagCliModel(self.diag_ctx, self.cli_ctx)
        with mock.patch("ascend_fd_tk.core.cli_module.cli_model.asyncio.run"):
            self.assertEqual(model.run_task(), "诊断完成")
        with mock.patch(
            "ascend_fd_tk.core.cli_module.cli_model.asyncio.run", side_effect=GenerateCsvPermissionErr("occupied")
        ):
            self.assertIn("生成报告失败", model.run_task())

    def test_auto_single_diag(self):
        model = AutoSingleDiagCliModel(self.diag_ctx, self.cli_ctx)
        self.assertEqual(AutoSingleDiagCliModel.max_param_count(), 2)
        self.assertIn("参数不足", model.run_task("1.1.1.1"))
        self.diag_ctx.diag_result = [object()]
        model = AutoSingleDiagCliModel(self.diag_ctx, self.cli_ctx)
        with (
            mock.patch("ascend_fd_tk.core.cli_module.cli_model.asyncio.run"),
            mock.patch("ascend_fd_tk.core.cli_module.cli_model.AutoSingleDiag") as mock_cls,
        ):
            mock_cls.return_value.error_msg = ""
            self.assertIn("1 条故障", model.run_task("1.1.1.1", "5"))


class TestClearCache(_CliTestBase):
    def test_clear_cache(self):
        model = ClearCacheCliModel(self.diag_ctx, self.cli_ctx)
        with (
            mock.patch("ascend_fd_tk.core.cli_module.cli_model.os.path.exists", return_value=False),
            mock.patch("ascend_fd_tk.core.cli_module.cli_model.shutil.rmtree") as mock_rmtree,
            mock.patch("ascend_fd_tk.core.cli_module.cli_model.os.remove") as mock_remove,
        ):
            self.assertEqual(model.run_task(), "清理完成")
            mock_rmtree.assert_not_called()
            mock_remove.assert_not_called()
        with mock.patch("ascend_fd_tk.core.cli_module.cli_model.os.path.exists", side_effect=OSError("denied")):
            self.assertIn("异常", model.run_task())


if __name__ == "__main__":
    unittest.main()
