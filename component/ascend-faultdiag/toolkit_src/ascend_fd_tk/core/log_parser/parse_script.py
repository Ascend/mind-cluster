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

import argparse
import json
import os
import platform
import re
import sys
from pathlib import Path
from typing import List, Tuple, Dict, Generator


def convert_log_path(input_path: str) -> str:
    os_name = platform.system().lower()
    abs_input_path = os.path.abspath(input_path)
    output_path = Path(f"\\\\?\\{abs_input_path}") if os_name == "windows" else Path(abs_input_path)
    if not output_path.exists() or output_path.is_dir():
        return ""
    return str(output_path)


# 找对应的py文件
SCRIPT_PATH = str(os.path.abspath(__file__)).rstrip("c")


def convert_path_regex(linux_style_regex):
    """
    将Linux风格的路径正则表达式转换为当前操作系统对应的风格

    参数:
        linux_style_regex: 包含Linux风格(/)分隔符的路径正则表达式

    返回:
        转换后适配当前系统的路径正则表达式
    """
    # 获取当前系统的路径分隔符
    current_sep = os.sep

    # 对于Windows系统，正则中需要用双反斜杠表示一个实际的反斜杠
    if current_sep == '\\':
        current_sep = r'\\\\'

    # 替换Linux风格的分隔符(包括可能的转义情况\/)
    # 先处理已经转义的/(\/)，再处理未转义的/
    converted = re.sub(r'\\/', current_sep, linux_style_regex)
    converted = re.sub(r'(?<!\\)/', current_sep, converted)

    return converted


def find_matching_files(parse_dir: str, filepath_pattern: str) -> Generator[str, None, None]:
    """查找匹配路径模式的文件"""
    pattern = re.compile(convert_path_regex(filepath_pattern), re.IGNORECASE)
    for root, _, files in os.walk(parse_dir):
        for file in files:
            file_path = os.path.join(root, file)
            rel_path = os.path.relpath(file_path, parse_dir)
            if pattern.search(rel_path):
                yield file_path


def search_keywords_in_file(file_path: str, keyword_patterns: List[Tuple[str, str]]) -> List[Tuple[str, str]]:
    """在文件中搜索匹配关键字的行"""
    matches = []
    patterns = [(pattern[0], re.compile(pattern[1])) for pattern in keyword_patterns]
    abs_file_path = convert_log_path(file_path)
    if not abs_file_path:
        return matches
    try:
        with open(abs_file_path, 'r', encoding='utf-8', errors='ignore') as f:
            for _, line in enumerate(f, 1):
                for pattern in patterns:
                    if pattern[1].search(line):
                        matches.append((pattern[0], line.strip()))
                        break
    except Exception as e:
        print(f"处理文件 {file_path} 错误: {str(e)}", file=sys.stderr)
    return matches


def process_pattern_group(parse_dir: str, pattern_group: List[Tuple[str, str, str]]) -> List[Dict]:
    """处理模式组并返回结果"""
    results = []
    path_patterns = {}
    for pattern_key, keyword_pattern, path_pattern in pattern_group:
        path_patterns.setdefault(path_pattern, []).append((pattern_key, keyword_pattern))

    for path_pattern, keyword_patterns in path_patterns.items():
        for file_path in find_matching_files(parse_dir, path_pattern):
            matches = search_keywords_in_file(file_path, keyword_patterns)
            for pattern_key, line_content in matches:
                results.append({
                    "pattern_key": pattern_key,
                    "log_path": os.path.relpath(file_path, parse_dir),
                    "logline": line_content,
                })
    return results


def load_patterns_from_config(config_path: str) -> List[Tuple[str, str, str]]:
    """从JSON配置文件加载模式组"""
    try:
        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)
        # 配置文件格式：[{"pattern_key": "pattern key", "keyword": "正则", "filepath": "正则"}, ...]
        return [(item["pattern_key"], item["keyword_pattern"], item["filepath_pattern"]) for item in config]
    except Exception as e:
        print(f"加载配置文件失败: {str(e)}", file=sys.stderr)
        sys.exit(1)


def main():
    """命令行入口（支持JSON配置和输出）"""
    parser = argparse.ArgumentParser(description='日志关键字清洗工具（JSON模式）')
    parser.add_argument('parse_dir', help='要解析的根目录')
    default_config = os.path.expanduser("~/.ascend-fd-tk/parse_config.json")
    parser.add_argument('-c', '--config', default=default_config,
                        help='JSON配置文件路径（必填）')
    # 输出路径默认值：~/.ascend-fd-tk/parse_result.json
    default_output = os.path.expanduser("~/.ascend-fd-tk/parse_result.json")
    parser.add_argument('-o', '--output', default=default_output,
                        help=f'输出JSON文件路径（默认：{default_output}）')

    args = parser.parse_args()

    # 确保输出目录存在
    output_dir = os.path.dirname(args.output)
    os.makedirs(output_dir, exist_ok=True)

    # 加载模式组并处理
    pattern_group = load_patterns_from_config(args.config)
    results = process_pattern_group(args.parse_dir, pattern_group)

    # 写入JSON结果
    try:
        with open(args.output, 'w', encoding='utf-8') as f:
            json.dump(results, f, ensure_ascii=False, indent=2)
        print(f"清洗完成，结果已保存到: {args.output}")
    except Exception as e:
        print(f"写入结果文件失败: {str(e)}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
