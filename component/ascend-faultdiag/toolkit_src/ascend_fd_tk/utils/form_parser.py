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

from typing import List


class FormParser:
    SPACE_INDENT = " "

    def __init__(self, key_separator=":", multi_key_in_line_separator: List[str] = None, append_multi_line=False,
                 skip_sign=""):
        self.key_separator = key_separator
        self.multi_key_in_line_separator = multi_key_in_line_separator or [","]
        self.append_multi_line = append_multi_line
        self.skip_sign = skip_sign

    def get_key_value(self, line_part: str):
        separator = self._find_separator(line_part)
        parts = line_part.split(separator)
        return parts[0].strip(), parts[1].strip()

    def cnt_space_indent(self, line: str):
        for i, char in enumerate(line):
            if char != self.SPACE_INDENT:
                return i
        return 0

    def parse(self, content: str):
        root = {}
        route_stack = [(-1, root)]
        last_key = ""
        cur_node = {}
        for line in content.splitlines():
            line = line.rstrip()
            if not line:
                continue
            if self.skip_sign and self.skip_sign in line:
                continue
            if self.key_separator not in line:
                is_str_last_key = isinstance(cur_node.get(last_key), str)
                is_param_exist = cur_node and last_key
                if self.append_multi_line and is_param_exist and is_str_last_key:
                    cur_node[last_key] = f"{cur_node.get(last_key, '')}\n{line.strip()}"
                continue
            # 记录缩进数量
            cur_indent = self.cnt_space_indent(line)
            # 缩进回退, 退出子节点
            new_stack = []
            for item in route_stack:
                if item[0] < cur_indent:
                    new_stack.append(item)
            route_stack = new_stack
            cur_node = route_stack[-1][1]

            # 新增子节点
            if self.key_separator == line[-1]:
                key = line[:-1].strip()
                cur_node[key] = {}
                last_key = key
                route_stack.append((cur_indent, cur_node[key]))

            else:
                parts = [line]
                for separator in self.multi_key_in_line_separator:
                    if separator in line:
                        parts = []
                        for part in line.split(separator):
                            if self.key_separator in part:
                                parts.append(part)
                            elif len(parts) > 0:
                                parts[-1] = parts[-1] + part
                        break
                for part in parts:
                    key, value = self.get_key_value(part)
                    last_key = key
                    cur_node[key] = value
        return root

    def _find_separator(self, line) -> str:
        separators = [
            self.SPACE_INDENT + self.key_separator + self.SPACE_INDENT,
            self.key_separator + self.SPACE_INDENT,
            self.SPACE_INDENT + self.key_separator
        ]
        for separator in separators:
            if separator in line:
                return separator
        return self.key_separator
