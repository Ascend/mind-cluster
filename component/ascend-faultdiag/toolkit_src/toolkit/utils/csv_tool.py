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

import csv
import os
from typing import List, Dict, Any, Optional

from toolkit.core.common.errors import GenerateCsvPermissionErr
from toolkit.utils import logger

_CONSOLE_LOGGER = logger.CONSOLE_LOGGER


def flatten_dict(d: Dict[str, Any], parent_key: str = '', sep: str = '_') -> Dict[str, Any]:
    """
    递归展平嵌套字典（如 {'a': {'b': 1}} -> {'a_b': 1}）
    :param d: 可能包含嵌套结构的字典
    :param parent_key: 父键名（用于拼接嵌套键）
    :param sep: 键名分隔符
    :return: 展平后的字典
    """
    items: List[tuple] = []
    for k, v in d.items():
        new_key = f"{parent_key}{sep}{k}" if parent_key else k
        if isinstance(v, dict):
            # 递归展平嵌套字典
            items.extend(flatten_dict(v, new_key, sep=sep).items())
        else:
            items.append((new_key, v))
    return dict(items)


def get_all_columns(dict_list: List[Dict[str, Any]], flatten: bool = True, sep: str = '_') -> List[str]:
    """
    获取所有字典中的键（列名），自动去重并排序
    :param dict_list: 字典列表
    :param flatten: 是否展平嵌套字典
    :param sep: 嵌套键分隔符
    :return: 所有列名的列表
    """
    columns = set()
    for d in dict_list:
        if flatten:
            # 展平后获取键
            flat_d = flatten_dict(d, sep=sep)
            columns.update(flat_d.keys())
        else:
            # 不展平，直接获取顶层键
            columns.update(d.keys())
    # 排序列名，确保输出顺序一致
    return sorted(columns)


def dict_list_to_csv(
        dict_list: List[Dict[str, Any]],
        output_path: str,
        columns: Optional[List[str]] = None,
        flatten: bool = False,
        sep: str = '_',
        na_rep: str = '',
        encoding: str = 'utf-8-sig',
        **csv_kwargs
) -> None:
    """
    将 list[dict] 转换为 CSV 文件
    :param dict_list: 输入的字典列表（每个字典代表一行）
    :param output_path: 输出 CSV 文件路径
    :param columns: 自定义列顺序（None 则自动获取所有列）
    :param flatten: 是否展平嵌套字典（如 {'a': {'b': 1}} -> 'a_b' 列）
    :param sep: 嵌套键的连接符（展平时使用）
    :param na_rep: 空值替换字符串（默认空字符串）
    :param encoding: 文件编码（默认 utf-8）
    :param csv_kwargs: 传递给 csv.writer 的额外参数（如 delimiter、quotechar 等）
    """
    if not dict_list:
        raise ValueError("输入的字典列表不能为空")

    # 确保输出目录存在
    output_dir = os.path.dirname(output_path)
    if output_dir and not os.path.exists(output_dir):
        os.makedirs(output_dir, exist_ok=True)

    # 处理嵌套字典（展平）
    processed_list = []
    for d in dict_list:
        if flatten:
            processed_list.append(flatten_dict(d, sep=sep))
        else:
            processed_list.append(d.copy())

    # 确定列名（自定义列或自动获取）
    if columns is None:
        columns = get_all_columns(processed_list, flatten=False)  # 已展平，无需再次处理
    else:
        # 检查自定义列是否存在（可选：跳过不存在的列或报错）
        all_cols = get_all_columns(processed_list, flatten=False)
        invalid_cols = [col for col in columns if col not in all_cols]
        if invalid_cols:
            _CONSOLE_LOGGER.info(f"警告：自定义列中存在不存在的键：{invalid_cols}（将输出空值）")

    # 写入 CSV
    try:
        with open(output_path, 'w', encoding=encoding, newline='') as f:
            writer = csv.DictWriter(f, fieldnames=columns, **csv_kwargs)
            # 写入表头
            writer.writeheader()
            # 写入行数据
            for row in processed_list:
                # 处理空值，确保所有列都有值（不存在的键用 na_rep 填充）
                row_data = {col: row.get(col, na_rep) for col in columns}
                writer.writerow(row_data)

        _CONSOLE_LOGGER.info(f"CSV 文件已生成：{output_path}（共 {len(dict_list)} 行，{len(columns)} 列）")
    except Exception as e:
        raise GenerateCsvPermissionErr(f"生成CSV文件到: {output_path} 失败, 可能是已打开文件占用, 异常: {e}") from e


# 便捷工具函数：从 CSV 读取回 list[dict]（可选功能）
def csv_to_dict_list(
        csv_path: str = "",
        flatten: bool = False,
        sep: str = '_',
        encoding: str = 'utf-8',
        **csv_kwargs
) -> List[Dict[str, Any]]:
    """将 CSV 文件读回 list[dict]（支持还原展平的嵌套结构）"""

    with open(csv_path, 'r', encoding=encoding, newline='') as f:
        reader = csv.DictReader(f, **csv_kwargs)
        dict_list = [row for row in reader]

    if not flatten:
        return dict_list

    # 还原嵌套结构（将 'a_b' 转为 {'a': {'b': ...}}）
    restored_list = []
    for d in dict_list:
        restored = {}
        for k, v in d.items():
            keys = k.split(sep)
            current = restored
            for i, key in enumerate(keys):
                if i == len(keys) - 1:
                    current[key] = v  # 最后一个键赋值
                else:
                    # 中间键创建嵌套字典
                    current[key] = current.get(key, {})
                    current = current[key]
        restored_list.append(restored)
    return restored_list


def csv_to_list_of_lists(filepath: str, delimiter: str = ',',
                         encoding: str = 'utf-8') -> List[List[Any]]:
    """
    读取CSV文件转为List[List]结构

    Args:
        filepath: CSV文件路径
        delimiter: 分隔符，默认逗号
        encoding: 文件编码，默认utf-8

    Returns:
        二维列表，每行一个子列表
    """
    data = []

    try:
        with open(filepath, 'r', encoding=encoding) as file:
            # 使用csv.reader
            csv_reader = csv.reader(file, delimiter=delimiter)

            for row in csv_reader:
                data.append(row)

        return data

    except FileNotFoundError:
        _CONSOLE_LOGGER.info(f"错误：找不到文件 {filepath}")
        return []
    except Exception as e:
        _CONSOLE_LOGGER.info(f"读取CSV文件时出错: {e}")
        return []
