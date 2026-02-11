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

import argparse
import json
import os


def recursively_remove_sensitives_in_python_file(base_path: str, keyword_replacements: dict):
    for root, _, files in os.walk(base_path):
        for file in files:
            if file.endswith(".py"):
                file_path = os.path.join(root, file)
                # remove npu_info_parser.py since tons of sensitive words
                if file_path.endswith("npu_info_parser.py"):
                    os.remove(file_path)
                    continue
                if file_path.endswith("package_parser.py"):
                    # remove the import and usages of npu_info_parser.py
                    npu_info_parser_skip_kwds = [["from", "import", "NpuInfoParser"]]
                    npu_info_import_replacement = {", NpuInfoParser]": "]"}
                    replace_keywords(file_path, npu_info_import_replacement, npu_info_parser_skip_kwds)
                if file_path.endswith("regular_table.py"):
                    # remove the constant of NpuInfoParser
                    regular_table_skip_kwds = [['"EnvInfoSaver"', '[NPU_INFO_SOURCE]']]
                    replace_keywords(file_path, {}, regular_table_skip_kwds)
                # remove with a filter of copyright
                replace_keywords(file_path, keyword_replacements, [["huawei", "copyright"]])


def replace_keywords(file_path, keyword_replacements: dict, sensitive_keywords: list):
    filtered_lines = []
    with open(file_path, "r", encoding="utf-8") as f:
        for line in f:
            if sensitive_keywords and keywords_in_line(line, sensitive_keywords):
                continue
            filtered_lines.append(line)
    full_text = "".join(filtered_lines)
    for old_key, new_key in keyword_replacements.items():
        full_text = full_text.replace(old_key, new_key)
    with open(file_path, "w", encoding="utf-8") as f:
        f.write(full_text)


def keywords_in_line(line: str, keywords: list):
    lower_line = line.lower()
    for keywords_group in keywords:
        lower_keywords = [kwd.lower() for kwd in keywords_group]
        if all(kwd in lower_line for kwd in lower_keywords):
            return True
    return False


def remove_chinese_config(path):
    with open(path, "r", encoding="utf-8") as f:
        config = json.load(f)

    for attr in config["knowledge-repository"].values():
        if "attribute" in attr:
            attribute = attr["attribute"]
            attribute.pop("cause_zh", None)
            attribute.pop("description_zh", None)
            attribute.pop("suggestion_zh", None)

    with open(path, 'w', encoding='utf-8') as file:
        json.dump(config, file, ensure_ascii=False, indent=4)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Replace import paths recursively.")
    parser.add_argument("--path", required=True)
    parser.add_argument("--old", required=True)
    parser.add_argument("--new", required=True)
    args = parser.parse_args()

    config_path = os.path.join(args.path, args.old, "configuration", "config.py")
    replace_map = {
        '"zh"': '"en"',
        '"ascend"': '"alan"',
        '"ascend-fd"': '"alan-fd"',
        '"~/.ascend_faultdiag/"': '"~/.alan_faultdiag/"',
        '"ASCEND_FD_HOME_PATH"': '"ALAN_FD_HOME_PATH"',
        '"custom-ascend-kg-config.json"': '"custom-alan-kg-config.json"',
        '"ascend_faultdiag_{}.log"': '"alan_faultdiag_{}.log"',
        '"ascend_faultdiag_operation.log"': '"alan_faultdiag_operation.log"',
        '"ascend-rc-parser.json"': '"rc-parser.json"',
        '"ascend-kg-parser.json"': '"kg-parser.json"',
        '"ascend-kg-analyzer.json"': '"kg-analyzer.json"',
        '"aicore-error-code-config-zh.json"': '"aicore-error-code-config-en.json"'
    }
    replace_keywords(config_path, replace_map, [])

    import_prefix_replacements = {
        f"from {args.old}.": f"from {args.new}.",
        f"import {args.old}.": f"import {args.new}."
    }
    recursively_remove_sensitives_in_python_file(args.path, import_prefix_replacements)

    manifest_path = os.path.join(args.path, "MANIFEST.in")
    replace_keywords(manifest_path, {f"{args.old}/": f"{args.new}/"}, [["include", "aicore-error-code-config-zh"]])

    kg_config_path = os.path.join(args.path, args.old, "configuration", "kg-config.json")
    remove_chinese_config(kg_config_path)
    kg_config_replacements = {
        "_Ascend": "",
        "_HW": "",
        '"Ascend-Docker-Runtime"': '"Docker-Runtime"',
        '"AscendBackend"': '"Backend"'
    }
    replace_keywords(kg_config_path, kg_config_replacements, [])

    i18n_path = os.path.join(args.path, args.old, "utils", "i18n.py")
    default_language_replacement = {
        'return lang if lang in ["zh", "en"] else "zh"': 'return lang if lang in ["zh", "en"] else "en"'
    }
    replace_keywords(i18n_path, default_language_replacement, [])
