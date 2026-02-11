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

import threading
from typing import Any, Iterable, Optional, Iterator


class SyncList:
    def __init__(self, init_data: Optional[Iterable] = None):
        """初始化线程安全列表，支持传入初始数据"""
        self._list = list(init_data) if init_data else []
        self._lock = threading.Lock()  # 互斥锁保证线程安全

    # ------------------------------
    # 核心：支持下标和切片操作
    # ------------------------------
    def __getitem__(self, index):
        """支持下标取值（如 list[0]）和切片（如 list[1:3]）"""
        with self._lock:
            return self._list[index]

    def __setitem__(self, index, value):
        """支持下标赋值（如 list[0] = 10）和切片赋值（如 list[1:3] = [20, 30]）"""
        with self._lock:
            self._list[index] = value

    def __delitem__(self, index):
        """支持删除下标元素（如 del list[0]）和切片（如 del list[1:3]）"""
        with self._lock:
            del self._list[index]

    def __len__(self) -> int:
        """支持 len(list) 获取长度"""
        with self._lock:
            return len(self._list)

    def __repr__(self) -> str:
        """打印列表时显示内容（如 print(list)）"""
        with self._lock:
            return repr(self._list)

    def __contains__(self, item: Any) -> bool:
        """支持 'item in list' 判断元素是否存在"""
        with self._lock:
            return item in self._list

    def __iter__(self) -> Iterator[Any]:
        """支持 for 循环遍历（返回线程安全的迭代器）"""
        # 先获取当前列表的快照（避免迭代中列表被修改导致错乱）
        with self._lock:
            snapshot = self._list.copy()

        # 生成器迭代快照（确保迭代过程中数据稳定）
        for item in snapshot:
            yield item

    # ------------------------------
    # 常用列表方法封装（线程安全）
    # ------------------------------
    def append(self, item: Any) -> None:
        """尾部添加元素"""
        with self._lock:
            self._list.append(item)

    def extend(self, items: Iterable) -> None:
        """批量添加可迭代对象中的元素"""
        with self._lock:
            self._list.extend(items)

    def insert(self, index: int, item: Any) -> None:
        """在指定位置插入元素"""
        with self._lock:
            self._list.insert(index, item)

    def remove(self, item: Any) -> None:
        """删除第一个匹配的元素（不存在则抛 ValueError）"""
        with self._lock:
            self._list.remove(item)

    def pop(self, index: int = -1) -> Any:
        """删除并返回指定位置的元素（默认最后一个）"""
        with self._lock:
            return self._list.pop(index)

    def index(self, item: Any, start: int = 0, end: Optional[int] = None) -> int:
        """返回元素第一次出现的索引（不存在则抛 ValueError）"""
        with self._lock:
            return self._list.index(item, start, end) if end is not None else self._list.index(item, start)

    def count(self, item: Any) -> int:
        """统计元素出现的次数"""
        with self._lock:
            return self._list.count(item)

    def sort(self, *args, **kwargs) -> None:
        """排序（支持 list.sort() 的所有参数，如 key、reverse）"""
        with self._lock:
            self._list.sort(*args, **kwargs)

    def reverse(self) -> None:
        """反转列表"""
        with self._lock:
            self._list.reverse()

    def clear(self) -> None:
        """清空列表"""
        with self._lock:
            self._list.clear()

    # ------------------------------
    # 其他常用功能
    # ------------------------------
    def copy(self) -> 'SyncList':
        """返回列表的浅拷贝（线程安全）"""
        with self._lock:
            return SyncList(self._list.copy())
