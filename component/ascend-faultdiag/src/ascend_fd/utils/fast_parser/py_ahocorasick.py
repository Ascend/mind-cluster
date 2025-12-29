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

class Node:
    def __init__(self, char=""):
        self.char = char
        self.children = {}  # 子节点字典 {字符: 节点}
        self.fail = None  # 失败指针
        self.output = []  # 匹配的模式串列表


class AhoCorasick:

    def __init__(self):
        self.root = Node()  # 根节点

    def add_pattern(self, pattern, value):
        """添加模式串到Trie树"""
        node = self.root
        for char in pattern:
            if char not in node.children:
                node.children[char] = Node(char)
            node = node.children[char]
        node.output.append(value)  # 叶节点标记模式串

    def build_failure(self):
        """构建失败指针"""
        queue = []
        # 初始化根节点的子节点的失败指针
        for child in self.root.children.values():
            child.fail = self.root
            queue.append(child)

        while queue:
            current = queue.pop(0)
            # 遍历所有子节点
            for char, child_node in current.children.items():
                queue.append(child_node)
                # 回溯失败指针
                fail_node = current.fail
                while fail_node and char not in fail_node.children:
                    fail_node = fail_node.fail
                # 设置子节点的失败指针
                child_node.fail = fail_node.children[char] if fail_node else self.root
                # 合并输出
                child_node.output.extend(child_node.fail.output)

    def search(self, text):
        """在文本中查找所有匹配的模式串"""
        results = []
        current = self.root
        for char in text:
            # 沿失败指针回溯直到找到匹配或根节点
            while current and char not in current.children:
                current = current.fail
            if not current:
                current = self.root
                continue
            # 匹配成功，转移到子节点
            current = current.children[char]
            # 收集所有匹配的模式串
            results.extend(current.output)
        return results
