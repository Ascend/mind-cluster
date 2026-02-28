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

import itertools
from typing import List, Callable, TypeVar, Dict, Any, Optional

T = TypeVar('T')
U = TypeVar('U')


def find_first(lst: List[T], predicate: Callable[[T], bool], default=None):
    for item in lst:
        if predicate(item):
            return item
    return default


def group_by_to_dict(data_list: List[T], key: Callable[[T], U]) -> Dict[U, List[T]]:
    sorted_data = sorted(data_list, key=key)
    return {k: list(v) for k, v in itertools.groupby(sorted_data, key=key)}


def group_by_to_list(data_list: List[T], key: Callable[[T], U]) -> List[List[T]]:
    return list(group_by_to_dict(data_list, key).values())


def safe_get(lst: List[T], index: int, default: Optional[T] = None) -> Optional[T]:
    """
    使用切片安全获取索引（避免异常处理）
    """
    if not lst:
        return default

    list_len = len(lst)
    positive_index = index % list_len  # 取模运算自动处理正负索引转换

    if positive_index < 0 or positive_index >= list_len:
        return default

    result = lst[positive_index:positive_index + 1]
    return result[0] if result else default


def list_of_lists_to_dict_list(data: List[List[Any]], headers: List[str] = None) -> List[Dict[str, Any]]:
    """
    将List[List]结构转换为List[Dict]结构

    Args:
        data: 二维列表数据
        headers: 可选的头部列表，如果不提供则使用第一行作为头部

    Returns:
        字典列表，每个字典代表一行数据
    """
    if not data:
        return []

    # 确定头部
    if headers is None:
        headers = data[0]
        rows = data[1:]  # 剩余行作为数据
    else:
        rows = data  # 所有行都是数据

    # 转换为字典列表
    result = []
    for row in rows:
        # 创建字典，键来自headers，值来自row
        row_dict = {}
        for i, header in enumerate(headers):
            if i < len(row):
                row_dict[header] = row[i]
            else:
                row_dict[header] = None  # 如果行长度不够，填充None
        result.append(row_dict)

    return result
