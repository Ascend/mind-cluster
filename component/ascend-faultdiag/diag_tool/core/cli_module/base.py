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

import abc
from typing import Dict, List, final

from diag_tool.core.context.diag_ctx import DiagCtx


class CliCtx:
    _HELP_KEY = "help"

    def __init__(self):
        self.is_running = True
        self.cli_model_map: Dict[str, CliModel] = {}

    def update_cli_models(self, cli_models: List["CliModel"]):
        self.cli_model_map.update({cli_model.get_key(): cli_model for cli_model in cli_models})

    def is_cmd_valid(self, cmd: str) -> bool:
        return cmd.strip().lower() in self.cli_model_map

    def run_cmd(self, cmd: str, *args) -> str:
        cmd = cmd.strip().lower()
        cli_model = self.cli_model_map.get(cmd)
        if not cli_model:
            return ""
        return cli_model.run(*args)

    def show_help(self) -> str:
        model = self.cli_model_map.get(self._HELP_KEY)
        if model:
            return model.run()
        return ""


class CliModel(abc.ABC):

    def __init__(self, diag_ctx: DiagCtx, cli_ctx: CliCtx):
        self.diag_ctx = diag_ctx
        self.cli_ctx = cli_ctx

    # 命令需要的key
    @classmethod
    @abc.abstractmethod
    def get_key(cls) -> str:
        pass

    # 大致帮助, 用于在help展示
    @abc.abstractmethod
    def get_help(self) -> str:
        pass

    # 详细说明, 用于指导详细操作
    @abc.abstractmethod
    def run_task(self, *args) -> str:
        pass

    # 添加命令行参数
    def add_arguments(self, parser):
        parser.add_argument("action", metavar='actions', nargs='?', choices=['?', '？'], default=None,
                            help=f"?(？)=查看{self.get_key()}详细信息；无参数={self.get_help()}")

    @final
    def run(self, *args) -> str:
        args = list(filter(bool, args))
        if args and args[0] in ("?", "？"):
            return self.get_detail()
        return self.run_task(*args)

    # 执行任务
    def get_detail(self) -> str:
        return self.get_help()


class DetailedCliModel(CliModel, abc.ABC):

    @abc.abstractmethod
    def get_detail(self) -> str:
        return self.get_help()
