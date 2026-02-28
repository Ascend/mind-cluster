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

import multiprocessing
import shlex
import sys
from typing import List

from ascend_fd_tk.core.cli_module.cli_model import build_cli_ctx
from ascend_fd_tk.core.context.diag_ctx import DiagCtx
from ascend_fd_tk.utils import logger

_CONSOLE_LOGGER = logger.CONSOLE_LOGGER


def _shlex_handle(input_str: str) -> List[str]:
    # 临时用文件名中不会出现的字符替换
    temp_input = input_str.replace('\\', '\\\\\\')  # \ -> \\\
    parts = shlex.split(temp_input.strip())
    parts = [part.replace('\\\\', '\\') for part in parts]  # \\\ -> \
    return parts


class DiagToolCLI:
    def __init__(self):
        self.diag_ctx = DiagCtx()
        self.cli_ctx = build_cli_ctx(self.diag_ctx)

    def process_command(self, input_str: str):
        """处理用户输入的命令"""
        if not input_str.strip():
            return
        parts = _shlex_handle(input_str)
        cmd = parts[0].lower()
        args = parts[1:] if len(parts) > 1 else []
        if not self.cli_ctx.is_cmd_valid(cmd):
            _CONSOLE_LOGGER.info(f"未知命令: {cmd}")
            _CONSOLE_LOGGER.info("使用 'help' 查看可用命令")
            return
        try:
            _CONSOLE_LOGGER.info(self.cli_ctx.run_cmd(cmd, *args))
        except Exception as e:
            _CONSOLE_LOGGER.info(f"执行异常: {e}")
            import traceback
            traceback.print_exc()

    def run(self):
        """运行主循环"""
        _CONSOLE_LOGGER.info("=== MindCluster ascend-faultdiag-toolkit诊断工具 ===")
        _CONSOLE_LOGGER.info("-- 交互式命令行模式 --")
        _CONSOLE_LOGGER.info("具体使用细节可以执行 ' guide ' 命令查看使用向导(首次使用请务必查看!!)")
        _CONSOLE_LOGGER.info("")
        _CONSOLE_LOGGER.info(self.cli_ctx.show_help())

        while self.cli_ctx.is_running:
            try:
                # 使用input获取用户输入
                user_input = input(">>> ")
                self.process_command(user_input)
            except EOFError:
                # 处理Ctrl+Z (Windows) 或 Ctrl+D (Unix)
                _CONSOLE_LOGGER.info("\n检测到文件结束符，退出程序")
                break
            except KeyboardInterrupt:
                # 处理Ctrl+C
                _CONSOLE_LOGGER.info("\n中断程序")
                choice = input("确定要退出吗？(y/n): ").lower()
                if choice == 'y':
                    break
            except Exception as e:
                _CONSOLE_LOGGER.info(f"程序错误: {e}")
                continue


def run_parser(cli: DiagToolCLI, args):
    len_args = len(args)
    for k, cli_model in cli.cli_ctx.cli_model_map.items():
        for i, arg in enumerate(args):
            if i >= len_args:
                break
            if arg != k:
                continue
            param = None
            if i + 1 < len_args and args[i + 1] not in cli.cli_ctx.cli_model_map:
                param = args[i + 1]
            _CONSOLE_LOGGER.info(cli_model.run(param))
            break


def main():
    """主函数"""
    multiprocessing.freeze_support()
    # 创建CLI实例
    cli = DiagToolCLI()

    parts = sys.argv
    if not parts:
        return
    if len(parts) == 1:
        cli.run()
    else:
        run_parser(cli, parts[1:])


if sys.platform == "win32":
    # Windows必须用spawn
    multiprocessing.freeze_support()
    multiprocessing.set_start_method("spawn", force=True)

if __name__ == "__main__":
    main()
