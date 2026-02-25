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

from typing import Dict, Any, Optional, Sequence


class MultiLevelDict:
    """
    简化版多级字典读写工具（仅支持字典嵌套，强制覆盖值）
    特性：target_data传入构造函数复用，自动创建中间层级，读写操作简洁
    """

    def __init__(self, target_data: Optional[Dict[str, Any]] = None):
        """
        初始化工具，绑定目标字典
        Args:
            target_data: 目标多级字典（默认自动创建空字典）
        """
        self.target_data = target_data or {}

    def write(self, keys: Sequence[str], value: Any) -> None:
        """
        按key序列写入值（强制覆盖已有值，自动创建中间层级）
        Args:
            keys: 多级key序列（如 ["a", "b", "c"] 对应 target_data["a"]["b"]["c"]）
            value: 要写入的值（任意类型，会覆盖已有值）
        Raises:
            ValueError: key序列为空时抛出
            TypeError: 中间层级不是字典时抛出（确保嵌套结构为纯字典）
        """
        if not keys:
            raise ValueError("key序列不能为空")

        current = self.target_data
        # 遍历除最后一个key外的中间层级，自动创建空字典
        for key in keys[:-1]:
            if not isinstance(current, dict):
                raise TypeError(f"中间层级必须是字典（当前key: {key}，实际类型: {type(current)}）")
            # 不存在的key自动初始化空字典
            if key not in current:
                current[key] = {}
            current = current[key]

        # 写入最终值（强制覆盖）
        final_key = keys[-1]
        if not isinstance(current, dict):
            raise TypeError(f"最终写入层级必须是字典（当前key: {final_key}，实际类型: {type(current)}）")
        current[final_key] = value

    def read(self, keys: Sequence[str], default: Optional[Any] = None) -> Any:
        """
        按key序列读取值（不存在时返回默认值，不报错）
        Args:
            keys: 多级key序列
            default: 不存在时的默认返回值（默认None）
        Returns: 找到的值或默认值
        Raises:
            ValueError: key序列为空时抛出
        """
        if not keys:
            raise ValueError("key序列不能为空")

        current = self.target_data
        for key in keys:
            # 中间层级不是字典或key不存在，直接返回默认值
            if not isinstance(current, dict) or key not in current:
                return default
            current = current[key]

        return current

    def delete(self, keys: Sequence[str]) -> bool:
        """
        按key序列删除值（删除成功返回True，不存在返回False）
        Args:
            keys: 多级key序列
        Returns: bool - 删除结果
        Raises:
            ValueError: key序列为空时抛出
        """
        if not keys:
            raise ValueError("key序列不能为空")

        current = self.target_data
        parent = None  # 父层级对象
        parent_key = None  # 父层级中当前key

        for key in keys:
            # 中间层级不是字典或key不存在，删除失败
            if not isinstance(current, dict) or key not in current:
                return False
            parent = current
            parent_key = key
            current = current[key]

        # 执行删除（确保父层级是字典且存在该key）
        if isinstance(parent, dict) and parent_key in parent:
            del parent[parent_key]
            return True
        return False

    def get_target_data(self) -> Dict[str, Any]:
        """获取当前完整的目标字典（便于外部查看或后续处理）"""
        return self.target_data
