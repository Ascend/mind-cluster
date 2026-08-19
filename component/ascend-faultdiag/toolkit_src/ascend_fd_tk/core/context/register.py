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
from typing import Callable, Dict, List, Optional, Tuple, Type

from ascend_fd_tk.core.collect.base import Collector
from ascend_fd_tk.core.collect.fetcher.base import Fetcher
from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.inspection.base import InspectionCheckItem
from ascend_fd_tk.utils import logger

_CONSOLE_LOGGER = logger.CONSOLE_LOGGER

# (analyzer 类, 适用代际列表)；未显式指定时默认 [NpuType.A3]
ANALYZER_STORE: List[Tuple[Type[Analyzer], List[NpuType]]] = []
INSPECTION_STORE: List[Type[InspectionCheckItem]] = []

# HostCollector 代际注册表：代际 -> 采集器类
HOST_COLLECTOR_REGISTRY: Dict[NpuType, Type[Collector]] = {}


def register_host_collector(generation: NpuType) -> Callable[[Type[Collector]], Type[Collector]]:
    """HostCollector 代际注册装饰器。

    新增代际只需新建 Collector 子类并加 @register_host_collector(NpuType.A6) 装饰器，
    无需修改 host_collector_factory，符合开闭原则。
    """

    def wrap(cls: Type[Collector]) -> Type[Collector]:
        HOST_COLLECTOR_REGISTRY[generation] = cls
        return cls

    return wrap


# HostSshFetcher 代际注册表：代际 -> fetcher 类
HOST_FETCHER_REGISTRY: Dict[NpuType, Type[Fetcher]] = {}


def register_host_fetcher(generation: NpuType) -> Callable[[Type[Fetcher]], Type[Fetcher]]:
    """HostSshFetcher 代际注册装饰器。

    新增代际只需新建 Fetcher 子类并加 @register_host_fetcher(NpuType.A6) 装饰器，
    """

    def wrap(cls: Type[Fetcher]) -> Type[Fetcher]:
        HOST_FETCHER_REGISTRY[generation] = cls
        return cls

    return wrap


def register_analyzer(
    cls=None,
    *,
    generation: Optional[List[NpuType]] = None,
):
    """注册 analyzer。

    Args:
        generation: 适用代际列表，统一为 List[NpuType]。
                    None（默认）按 [A3] 处理；单代际用单元素列表（如 [NpuType.A5]）；
                    多代际用多元素列表（如 [NpuType.A3, NpuType.A5]），
                    该 analyzer 在任一匹配代际集群下都会被实例化。
    """

    def wrap(target_cls):
        if not issubclass(target_cls, Analyzer):
            raise TypeError(f"类 {target_cls.__name__} 必须是 {Analyzer.__name__} 的子类")
        gens = _dedup_generations(generation)
        entry = (target_cls, gens)
        if entry not in ANALYZER_STORE:
            ANALYZER_STORE.append(entry)
        return target_cls

    return wrap(cls) if cls is not None else wrap


def _dedup_generations(generation: Optional[List[NpuType]]) -> List[NpuType]:
    """归一化 generation 为去重后的代际列表；None 默认 [NpuType.A3]。"""
    if not generation:
        return [NpuType.A3]
    gens: List[NpuType] = []
    for gen in generation:
        if gen not in gens:
            gens.append(gen)
    return gens


def get_analyzers(generation: NpuType) -> List[Type[Analyzer]]:
    """按代际过滤返回匹配的 analyzer 类列表。

    只要 analyzer 声明的代际列表包含目标代际，即视为匹配。
    """
    return [cls for cls, gens in ANALYZER_STORE if generation in gens]


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
        onerror=lambda x: None,  # 忽略导入错误
    ):
        try:
            # 导入模块（触发模块内类的定义和注册）
            importlib.import_module(module_info.name)
        except ImportError as e:
            _CONSOLE_LOGGER.info("导入模块 %s 失败：%s", module_info.name, str(e))
