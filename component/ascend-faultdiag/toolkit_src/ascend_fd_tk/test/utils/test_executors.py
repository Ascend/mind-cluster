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

import asyncio
import os
import tempfile
import unittest

from ascend_fd_tk.utils.executors import CommandResult, CmdTask, AsyncCmdExecutor, AsyncSSHExecutor


class TestCommandResult(unittest.TestCase):
    """测试 CommandResult 数据封装"""

    def test_success_result(self):
        """测试成功结果"""
        result = CommandResult(cmd="echo hello", returncode=0, stdout="hello", stderr="")
        self.assertTrue(result.is_success())
        self.assertEqual(result.stdout, "hello")
        self.assertEqual(result.returncode, 0)

    def test_failed_result(self):
        """测试失败结果"""
        result = CommandResult(cmd="false", returncode=1, stdout="", stderr="error")
        self.assertFalse(result.is_success())
        self.assertEqual(result.stderr, "error")


class TestCmdTask(unittest.TestCase):
    """测试 CmdTask 任务封装"""

    def test_cmd_task_init(self):
        """测试 CmdTask 初始化"""
        task = CmdTask(cmd="ls -l", timeout=10)
        self.assertEqual(task.cmd, "ls -l")
        self.assertEqual(task.timeout, 10)
        self.assertIsNotNone(task.on_failed)  # 默认错误回调


class TestAsyncCmdExecutor(unittest.TestCase):
    """测试 AsyncCmdExecutor 本地命令执行器（ST 业务接口级别）"""

    def test_run_cmd_success(self):
        """测试执行单条本地命令 - 正常调用"""
        executor = AsyncCmdExecutor()
        task = CmdTask(cmd="echo st_test", timeout=5)
        result = asyncio_run(executor.run_cmd(task))
        self.assertTrue(result.is_success())
        self.assertEqual(result.stdout, "st_test")

    def test_run_parallel(self):
        """测试并行执行多条命令 - 正常调用"""
        executor = AsyncCmdExecutor()
        tasks = [
            CmdTask(cmd="echo cmd1", timeout=5),
            CmdTask(cmd="echo cmd2", timeout=5),
        ]
        results = asyncio_run(executor.run_parallel(tasks))
        self.assertEqual(len(results), 2)
        outputs = {r.stdout for r in results}
        self.assertEqual(outputs, {"cmd1", "cmd2"})

    def test_upload_and_download_file(self):
        """测试上传（复制）和下载（复制）文件 - 正常调用"""
        executor = AsyncCmdExecutor()
        with tempfile.TemporaryDirectory() as tmpdir:
            src_file = os.path.join(tmpdir, "src.txt")
            mid_file = os.path.join(tmpdir, "mid", "mid.txt")
            dst_file = os.path.join(tmpdir, "dst", "dst.txt")
            with open(src_file, "w", encoding="utf-8") as f:
                f.write("st_file_content")

            # upload: src -> mid
            asyncio_run(executor.upload_file(src_file, mid_file))
            self.assertTrue(os.path.exists(mid_file))

            # download: mid -> dst
            asyncio_run(executor.download_file(mid_file, dst_file))
            self.assertTrue(os.path.exists(dst_file))

            with open(dst_file, "r", encoding="utf-8") as f:
                self.assertEqual(f.read(), "st_file_content")

    def test_close(self):
        """测试关闭执行器 - 正常调用"""
        executor = AsyncCmdExecutor()
        asyncio_run(executor.close())


class TestAsyncSSHExecutorConstruction(unittest.TestCase):
    """测试 AsyncSSHExecutor 构造（不涉及真实网络连接）"""

    def test_construct_without_key(self):
        """测试无私钥构造 - 正常调用"""
        executor = AsyncSSHExecutor(host="127.0.0.1", port=22, username="root", password="test", timeout=5)
        self.assertEqual(executor.host, "127.0.0.1")
        self.assertEqual(executor.port, 22)
        self.assertIsNone(executor.private_key_obj)
        self.assertIsNone(executor.ssh_client)

    def test_construct_with_nonexistent_key(self):
        """测试传入不存在的私钥路径 - 正常调用（应优雅返回 None）"""
        executor = AsyncSSHExecutor(host="127.0.0.1", port=22, username="root", private_key="/nonexistent/key")
        self.assertIsNone(executor.private_key_obj)


def asyncio_run(coro):
    """辅助函数：运行异步协程"""
    return asyncio.run(coro)


if __name__ == "__main__":
    unittest.main()
