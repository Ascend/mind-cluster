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

import datetime
import math
import re
from typing import List, Tuple, Any, Dict

from toolkit.core.common import constants
from toolkit.utils.logger import DIAG_LOGGER


def split_str(data: str, input_pattern: str, regex=False) -> List[str]:
    result = []
    current_block = []
    for line in data.splitlines(keepends=True):
        is_input_valid = not regex and input_pattern in line
        is_search_valid = regex and re.search(input_pattern, line)
        if is_input_valid or is_search_valid:
            # 如果当前块不为空，先保存之前的块
            if current_block:
                result.append(''.join(current_block))
                current_block = []
            # 添加当前分隔符行到新块
            current_block.append(line)
            continue
        # 非分隔符行，加入当前块
        if current_block:  # 只处理已有分隔符开头的块（避免空行干扰）
            current_block.append(line)
    # 处理最后一个块
    if current_block:
        result.append(''.join(current_block))
    return [block.strip() for block in result]


def to_int(data, default=0) -> int:
    if isinstance(data, str) and data.isdigit():
        return int(data)
    elif isinstance(data, int):
        return data
    return default


def parse_hex(hex_str: str, default=0) -> int:
    hex_str = hex_str.strip()
    try:
        return int(hex_str, 16)
    except ValueError:
        return default


def trans_date_fmt(date_str: str, src_fmt: str, target_fmt: str) -> str:
    try:
        date = datetime.datetime.strptime(date_str, src_fmt)
        return date.strftime(target_fmt)
    except Exception as e:
        DIAG_LOGGER.error(f"trans date fmt failed: {date_str} {src_fmt} to {target_fmt}, error: {e}")
        return ""


def to_float(s: str, default=float(constants.SYS_INT_MIN_SIZE)) -> Tuple[bool, float]:
    # 先判断是否为字符串（避免非字符串输入报错）
    if not isinstance(s, str):
        return False, default
    # 空字符串直接返回 False
    if s.strip() == "":
        return False, default
    try:
        # 尝试转换，成功则返回 True
        return True, float(s)
    except (ValueError, TypeError):
        # 转换失败（非法格式或非字符串），返回 False
        return False, default


def find_pattern_after_substrings(long_string: str, substrings: List[str], pattern: str, end_sign="\n"):
    """
    高性能版本：迭代搜索子串位置，然后在剩余文本中匹配正则
    """
    current_pos = 0

    # 顺序查找所有子串
    for substr in substrings:
        pos = long_string.find(substr, current_pos)
        if pos == -1:
            return None  # 未找到某个子串
        current_pos = pos + len(substr)  # 移动到子串之后
    # 结束标记
    next_sep = long_string.find(end_sign, current_pos)
    # 在最后一个子串之后的文本中搜索正则
    remaining_text = long_string[current_pos:next_sep]
    match = re.search(pattern, remaining_text)
    return match


def camel_to_separated(camel_str: str, separator='_'):
    """
    将驼峰格式字符串转换为分割模式

    参数:
    camel_str: 驼峰格式字符串
    separator: 分隔符，默认为下划线'_'

    返回:
    str: 分割后的字符串
    """
    if not camel_str or not isinstance(camel_str, str):
        return camel_str

    def get_replacement_template(sep: str) -> str:
        return r'\1{0}\2'.format(sep)

    # 先把非字母数字统一转为分隔符
    converted = re.sub(r'[^a-zA-Z0-9]+', separator, camel_str).strip(separator)
    replace_template = get_replacement_template(separator)
    # 同时处理连续大写字母（如HTTPRequest -> HTTP_Request）
    # 先在大写字母后跟小写字母的位置插入分隔符
    converted = re.sub(r'([A-Z]+)([A-Z][a-z])', replace_template, converted)
    # 然后在小写字母后跟大写字母的位置插入分隔符
    converted = re.sub(r'([a-z])([A-Z])', replace_template, converted)
    # 全部转换为小写
    return converted.lower()


def mw_to_dbm(mw: str, default="-40") -> str:
    success, mw_f = to_float(mw)
    if not success or mw_f <= 0:
        return default
    res = 10 * math.log10(mw_f)
    return str(round(res, 2))


def dbm_to_mw(dbm: str, default="0.0001") -> str:
    success, dbm_f = to_float(dbm)
    if not success or dbm_f <= 0:
        return default
    res = 10 ** (dbm_f / 10)
    return str(round(res, 2))


def strip_keys_from_parentheses(src_dict: Dict[str, Any]) -> Dict[str, Any]:
    new_dict = {}
    for key, value in src_dict.items():
        if not isinstance(key, str):
            continue
        key = re.sub(r"\W.*\W", "", key)
        new_dict[key] = value
    return new_dict
