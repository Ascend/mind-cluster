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

import importlib
import pkgutil
import sys
from pathlib import Path
from typing import List, Type

from diag_tool.core.fault_analyzer.base import Analyzer
from diag_tool.core.inspection.base import InspectionCheckItem
from diag_tool.utils import logger  

_CONSOLE_LOGGER = logger.CONSOLE_LOGGER

ANALYZER_STORE: List[Type[Analyzer]] = []
INSPECTION_STORE: List[Type[InspectionCheckItem]] = []


def register_analyzer(cls):
    # 确保被装饰的类是 Analyzer 的子类（可选校验）
    if not issubclass(cls, Analyzer):
        raise TypeError(f"类 {cls.__name__} 必须是 {Analyzer.__name__} 的子类")
    if cls not in ANALYZER_STORE:
        ANALYZER_STORE.append(cls)

    return cls  # 装饰器需返回原类，不改变其功能


def register_inspection_check_item(cls):
    # 确保被装饰的类是 Analyzer 的子类（可选校验）
    if not issubclass(cls, InspectionCheckItem):
        raise TypeError(f"类 {cls.__name__} 必须是 {InspectionCheckItem.__name__} 的子类")
    if cls not in INSPECTION_STORE:
        INSPECTION_STORE.append(cls)

    return cls  # 装饰器需返回原类，不改变其功能


def recursive_scan_and_register(root_module: str) -> None:
    """
    递归扫描指定根模块下的所有子模块，并自动导入触发注册

    参数:
        root_module: 根模块名称（如 "my_analyzers"）或模块路径
    """
    # 解析根模块的路径和名称
    if Path(root_module).exists():
        # 若传入路径，转换为模块名称（假设根目录在 sys.path 中）
        root_path = Path(root_module).resolve()
        root_name = root_path.name
        # 将根目录添加到 Python 路径，确保能被导入
        if str(root_path.parent) not in sys.path:
            sys.path.append(str(root_path.parent))
    else:
        # 若传入模块名，获取其路径
        root_module_obj = importlib.import_module(root_module)
        root_path = Path(root_module_obj.__file__).parent
        root_name = root_module

    # 递归遍历所有子模块
    for module_info in pkgutil.walk_packages(
            path=[str(root_path)],  # 模块路径
            prefix=f"{root_name}.",  # 模块名称前缀（确保导入路径正确）
            onerror=lambda x: None  # 忽略导入错误
    ):
        try:
            # 导入模块（触发模块内类的定义和注册）
            importlib.import_module(module_info.name)
        except ImportError as e:
            _CONSOLE_LOGGER.info(f"导入模块 {module_info.name} 失败: {e}")
