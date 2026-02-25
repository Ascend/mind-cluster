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

import json
import os
import platform
from pathlib import Path
from typing import List

from toolkit.utils import logger

MAX_SIZE = 512 * 1024 * 1024
MB_SHIFT = 20
_CONSOLE_LOGGER = logger.CONSOLE_LOGGER


def safe_read_open(file_path: str, *args, **kwargs):
    if not os.path.islink(file_path):
        raise Exception(f"The {os.path.basename(file_path)} should not be a symbolic link file.")
    file_real_path = os.path.realpath(file_path)
    file_stream = open(file_real_path, *args, **kwargs)
    file_info = os.stat(file_stream.fileno())
    if file_info.st_size > MAX_SIZE:
        file_stream.close()
        raise Exception(f"The size of {os.path.basename(file_path)} should be less than {MAX_SIZE >> MB_SHIFT} MB.")
    return file_stream


def safe_read_json(file_path: str, *args, **kwargs):
    with safe_read_open(file_path, "r", encoding="utf-8") as file_stream:
        try:
            return json.load(file_stream, *args, **kwargs)
        except Exception as error:
            logger.DIAG_LOGGER.error("Read json file %s failed, error: %s", os.path.basename(file_path), error)
            return {}


def find_all_sub_paths(
        root_dir: str,
        target_path_pattern: str,
        max_depth: int = 2
) -> List[str]:
    """优化版：用 os.scandir 提升大规模目录遍历性能"""
    root_path = Path(root_dir)
    if not root_path.exists() or not root_path.is_dir():
        _CONSOLE_LOGGER.log(f"警告：根目录不存在或不是合法目录：{root_dir}")
        return []

    collect_dirs = []
    # 按深度生成匹配规则：如 max_depth=3 → ["collect", "*", "collect", "*/*", "collect"]
    for depth in range(1, max_depth + 1):
        # 构造匹配模式：depth=1 → "collect"（root_dir/collect）；depth=2 → "*"/"collect"（root_dir/任意/collect）
        pattern_parts = ["*"] * (depth - 1) + [target_path_pattern]
        pattern = "/".join(pattern_parts)
        # 匹配当前深度的 collect 目录
        for dir_path in root_path.glob(pattern):
            collect_dirs.append(str(dir_path.resolve()))  # 存储绝对路径

    # 去重（避免同一目录被多层匹配重复捕获，理论上不会发生，保险起见）
    return list(set(collect_dirs))


def jump_up_directories(current_dir, levels=1):
    """
    向上跳转指定层级的目录

    参数:
    current_dir: 当前目录路径
    levels: 向上跳转的层级数，默认1（父目录）

    返回:
    跳转后的目录路径字符串
    """
    if not current_dir or levels <= 0:
        return current_dir

    # 使用pathlib处理
    current_path = Path(current_dir).resolve()

    # 逐级向上跳转
    target_path = current_path
    for _ in range(levels):
        if target_path.parent == target_path:  # 已到达根目录
            break
        target_path = target_path.parent

    return str(target_path)


def convert_log_path(input_path: str) -> str:
    os_name = platform.system().lower()
    abs_input_path = os.path.abspath(input_path)
    output_path = Path(f"\\\\?\\{abs_input_path}") if os_name == "windows" else Path(abs_input_path)
    if not output_path.exists() or not output_path.is_dir():
        return ""
    return str(output_path)
