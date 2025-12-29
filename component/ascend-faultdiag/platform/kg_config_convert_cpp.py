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
import os.path
import logging
import sys

echo_handler = logging.StreamHandler(sys.stdout)
echo_handler.setFormatter(logging.Formatter('%(message)s'))
echo_logger = logging.getLogger("ECHO")
echo_logger.addHandler(echo_handler)

IN_CATE = "in"
ATTR_CATE = "attribute"


def escape_string(s: str) -> str:
    return (s.replace("\\", "\\\\").replace("\"", "\\\"").replace("\n", "\\n"))


def refactor_attribute_dict(value_dict):
    """
    处理 attribute 字典：
    1. 将真实换行符 \n 转义成 \\n
    2. 将反斜杠 \ 转义成 \\
    3. 将双引号 " 转义成 \"
    4. 列表拼接成单行文本，使用 \\n 分隔
    """
    new_attribute = {}
    for key, value in value_dict.items():
        if isinstance(value, str):
            new_attribute[key] = escape_string(value)
        elif isinstance(value, list):
            # 列表 -> 转义每个元素 -> 拼接
            new_attribute[key] = "\\n".join(escape_string(str(item)) for item in value)
        else:
            new_attribute[key] = str(value)

    return new_attribute

def refactor_knowledge_repository(json_data):
    data = json_data.setdefault("knowledge-repository", {})
    for _, attr_dict in data.items():
        attribute_dict = attr_dict.get(ATTR_CATE, {})
        if attribute_dict:
            new_attribute = refactor_attribute_dict(attribute_dict)
            attr_dict[ATTR_CATE] = new_attribute
        regex_dict = attr_dict.get("regex", {})
        if IN_CATE in regex_dict:
            in_list = regex_dict.get(IN_CATE, [])
            if not in_list:
                continue
            if isinstance(in_list[0], str):
                in_list = [in_list]
            attr_dict["regex"] = {IN_CATE: in_list}


def filter_chinese_config(event_attributes: dict):
    attribute = event_attributes.get("attribute", {})
    for k in ["cause_zh", "description_zh", "suggestion_zh"]:
        attribute.pop(k, None)


def kg_config_convert_cpp():
    """
    Convert the kg file to cpp file
    """
    parser = argparse.ArgumentParser()
    parser.add_argument('--input_file', type=str, help='This is a parameter')
    parser.add_argument('--output_path', type=str, help='This is a parameter')
    parser.add_argument("--mode", type=str, default="zh", choices=["zh", "en"], help="select the language")
    args = parser.parse_args()

    if not args.input_file or not args.output_path:
        echo_logger.info("No parameter provided.")
        return
    echo_logger.info(f"The value of --input_file is: %s", args.input_file)
    echo_logger.info(f"The value of --output_path is: %s", args.output_path)

    with open(args.input_file, 'r', encoding="utf-8") as file:
        data = json.load(file)
    refactor_knowledge_repository(data)
    fault_diag_lib_cpp = "\nvoid FaultDiagSpace::Init()\n{\n"
    for event_code, event_attribute in data.get("knowledge-repository", {}).items():
        if args.mode == "en":
            filter_chinese_config(event_attribute)

        # 生成JSON
        json_str = json.dumps(event_attribute, ensure_ascii=False, separators=(',', ':'))
        json_str = json_str.replace("\n", "\\n")  # json解析多行字符串时还原为真正的换行符\n

        # 转成UTF-8字节流，再逐字节+128
        utf8_bytes = json_str.encode("utf-8")
        json_char_list = [str(b + 128) for b in utf8_bytes]
        json_char_len = len(json_char_list)
        code_name = event_code.replace("_", "")
        fault_diag_lib_cpp += f"\tstatic int char{code_name}[] = {{{', '.join(json_char_list)}}};\n"
        fault_diag_lib_cpp += f"\tg_faultDiag[\"{event_code}\"] = " \
                              f"FaultDiagSpace::FaultDiagEvent(char{code_name}, {json_char_len});\n"
        source_file = event_attribute.get("source_file")
        if source_file:
            fault_diag_lib_cpp += f"\tg_faultCode[\"{source_file}\"].push_back(\"{event_code}\");\n"
    fault_diag_lib_cpp += f"\tInitAcSearchers();\n"
    fault_diag_lib_cpp += "}\n"

    # 强制指定 UTF-8 写文件，避免 ARM/x86 差异
    out_file = os.path.join(args.output_path, 'fault_diag.cpp')
    with open(out_file, 'a', encoding="utf-8") as file:
        file.write(fault_diag_lib_cpp)


if __name__ == "__main__":
    kg_config_convert_cpp()
