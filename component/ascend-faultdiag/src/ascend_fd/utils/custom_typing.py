#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
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
import sys
from typing import Any, Tuple

# 导入各版本所需的内部类，暂不处理可变参数泛型
if sys.version_info >= (3, 8):
    from typing import _GenericAlias, _SpecialForm
else:  # 3.7
    from typing import _GenericAlias

    _SpecialForm = None

# 处理Python 3.10+中的类型泛化
if sys.version_info >= (3, 10):
    from types import GenericAlias


def get_origin(tp: Any) -> Any:
    """
    获取类型的原始类型（去除参数的版本）
    支持泛型类型、Callable、Tuple、Union等
    """
    # 处理Python 3.10+中list[int]这样的泛型别名
    if sys.version_info >= (3, 10) and isinstance(tp, GenericAlias):
        return tp.__origin__

    # 处理_GenericAlias实例（大多数泛型类型）
    if isinstance(tp, _GenericAlias):
        return tp.__origin__

    # 处理特殊形式（如Union、Tuple等）
    if _SpecialForm is not None and isinstance(tp, _SpecialForm):
        return tp
    # 处理Python 3.9+中的list、dict等原生泛型
    if sys.version_info >= (3, 9):
        if tp is list:
            return list
        if tp is dict:
            return dict
        if tp is set:
            return set
        if tp is tuple:
            return tuple

    return None


def get_args(tp: Any) -> Tuple[Any, ...]:
    """
    获取类型的参数元组，已执行所有替换
    对于Union类型，会执行Union构造函数使用的基本简化
    """
    # 处理Python 3.10+中list[int]这样的泛型别名
    if sys.version_info >= (3, 10) and isinstance(tp, GenericAlias):
        return tp.__args__ if tp.__args__ is not None else ()

    # 处理_GenericAlias实例
    if isinstance(tp, _GenericAlias):
        res = tp.__args__
        return res if res is not None else ()

    # 处理特殊形式
    if _SpecialForm is not None and isinstance(tp, _SpecialForm):
        return ()

    # 处理Python 3.9+中的原生泛型
    if sys.version_info >= (3, 9):
        if isinstance(tp, (list, dict, set, tuple)) and hasattr(tp, '__args__'):
            return tp.__args__ if tp.__args__ is not None else ()

    return ()