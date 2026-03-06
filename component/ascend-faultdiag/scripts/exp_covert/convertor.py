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
import os
import json
import inspect
import argparse


class TroubleShootingExpConvertor:
    REGEX_KEY_ORIGIN = {"Key Log Information": "log_info_regex",
                        "Key Python Stack Information": "python_stack_regex"}
    REGEX_KEY_REPLACE = {"Key C++ Stack Information": "cpp_stack_regex",
                         "Code Path": "code_path_regex"}
    NEW_KEY = {
        "Fault Name": "cause_zh",
        "Fault Cause": "description_zh",
        "Modification Suggestion": "suggestion_zh",
        "Fault Case": "reference",
        "Error Case": "error_case",
        "Fixed Case": "fixed_case",
        "Fixed Code": "fixed_case"
    }
    REMOVE_KEY = ["Test Case", "Device Type", "ID", "Sink Mode"]

    def __init__(self, lib_list, suggestion_path: str, regex_path: str, fault_code_path: str):
        self.lib_list = lib_list
        self.json_path_list = [suggestion_path, regex_path, fault_code_path]
        self.sug_dict, self.reg_dict, self.fault_code_dict = [self.read_dict_from_json(file_path)
                                                              for file_path in self.json_path_list]

    @staticmethod
    def read_dict_from_json(file_path: str):
        """
        Read the existing exp from json file
        :param file_path: json file path
        :return: json dict
        """
        file_path = os.path.realpath(file_path)
        if not os.path.exists(file_path):
            return dict()
        with open(file_path, encoding="utf-8") as file_stream:
            return json.load(file_stream)

    @staticmethod
    def write_dict_to_json(json_dict: dict, file_path: str):
        """
        Write the exp to the json file
        :param json_dict: exp json dict
        :param file_path: json file path
        """
        file_path = os.path.realpath(file_path)
        with os.fdopen(os.open(file_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o640), "w",
                       encoding="utf-8") as file_stream:
            file_stream.write(json.dumps(json_dict, ensure_ascii=False, indent=4))

    @staticmethod
    def add_newline_symbol(origin_msg: str):
        """
        Add new line symbol for origin message,
        maybe cannot add newline for all situation or add wrong,
        please check the final json and manual modification
        :param origin_msg: origin msg
        :return: new msg
        """
        for i in range(2, 10):
            split_symbol = f"{i}."
            msg_list = origin_msg.split(split_symbol)
            # skip 'r2.0.0' version info and x.html
            if len(msg_list) == 2 and msg_list[0][-1] != "r" and msg_list[1][:4] != "html":
                origin_msg = f"\n{i}.".join(msg_list)
        return origin_msg

    def job(self):
        """
        Start convert job and save exp json file
        """
        self._convert()
        for json_dict, file_path in zip([self.sug_dict, self.reg_dict, self.fault_code_dict], self.json_path_list):
            self.write_dict_to_json(json_dict, file_path)

    def _convert(self):
        """
        Convert the lib exp to ascend-fd exp json
        """
        for lib in self.lib_list:
            for name, obj in inspect.getmembers(lib):
                reg_list = []
                if "experience" not in name:
                    continue
                for exp_dict in obj:
                    is_general = True if "general" in name else False
                    suggestion_exp, regex_exp = self._add_one_exp(exp_dict, is_general)
                    self.sug_dict.update(suggestion_exp)
                    reg_list.append(regex_exp)
                self.reg_dict.update({name: reg_list})

    def _add_one_exp(self, exp_lib_dict: dict, is_general=False):
        """
        Add one exp from exp lib
        :param exp_lib_dict: mindspore exp lib dict
        :param is_general: Indicates whether the type is general.
        :return: one experience dict, one regex dict
        """
        exp_id = exp_lib_dict.get("ID")
        if not exp_id:
            raise ValueError(f"The {exp_lib_dict.keys()} don't contain ID.")
        if exp_id == "common_id_1":  # default故障不生成解析器
            return dict(), dict()

        if not self.fault_code_dict.values():
            fault_code = 21001  # 第一个fault code值
        elif f"Mindspore_{exp_id}" in self.fault_code_dict:
            fault_code = self.fault_code_dict.get(f"Mindspore_{exp_id}")  # 如果该fault id存在了，则取出该id对应的fault_code
        else:
            fault_code = int(max(self.fault_code_dict.values())) + 1  # 如果该fault id不存在，则使用当前最大fault_code + 1

        regex_exp = {"name": f"Mindspore_{exp_id}", "is_general": is_general}
        self.fault_code_dict.update({f"Mindspore_{exp_id}": fault_code})
        suggestion_exp, regex_base = self._get_one_exp_dict(fault_code, exp_lib_dict)
        regex_exp.update(regex_base)
        suggestion_exp = {str(fault_code): suggestion_exp}
        return suggestion_exp, regex_exp

    def _get_one_exp_dict(self, fault_code: int, ori_exp: dict):
        """
        Convert one exp to suggestion and regex
        :param fault_code: fault code
        :param ori_exp: origin mindspore exp lib
        :return: suggestion exp, regex exp
        """
        suggestion_exp = {"code": fault_code}
        regex_exp = dict()
        for key, value in ori_exp.items():
            if key in self.REGEX_KEY_ORIGIN:
                regex_exp.update({self.REGEX_KEY_ORIGIN.get(key): value})
                continue
            if key in self.REGEX_KEY_REPLACE:  # 对于这两类正则规则，将其反斜杠replace为正斜杠
                regex_exp.update({self.REGEX_KEY_REPLACE.get(key): value.replace("\\", "/")})
                continue
            if key in self.NEW_KEY:
                if key in ["Error Case", "Fixed Case", "Fixed Code"]:
                    suggestion_exp.update({self.NEW_KEY.get(key): value.strip("\n")})
                    continue
                if key in ["Modification Suggestion", "Fault Case"]:
                    suggestion_exp.update({self.NEW_KEY.get(key): self.add_newline_symbol(value)})
                    continue
                suggestion_exp.update({self.NEW_KEY.get(key): value})
                continue
            if key in self.REMOVE_KEY:
                continue
            raise KeyError(f"Key '{key}' don't convert, maybe new key add, please check.")
        return suggestion_exp, regex_exp


def command_line():
    """
    The command line interface. Commands contain:
    -s, --suggestion_path, the input path of suggestion json file
    -r, --regex_path, the input path of regex json file
    -f, --fault_code_path, the input path of fault_code json file
    """
    args = argparse.ArgumentParser(add_help=True, description="MS Exp Convertor")
    args.add_argument("-s", "--suggestion_path", type=str, required=True, metavar='',
                      help="the input path of suggestion json file. "
                           "The file will be created when it doesn't exist.")
    args.add_argument("-r", "--regex_path", type=str, required=True, metavar='',
                      help="the input path of regex json file."
                           "The file will be created when it doesn't exist.")
    args.add_argument("-f", "--fault_code_path", type=str, required=True, metavar='',
                      help="the input path of fault_code json file."
                           "The file will be created when it doesn't exist.")
    return args.parse_args()


if __name__ == "__main__":
    exp_lib_list = []  # 导入所有exp lib 文件
    arg_cmd = command_line()
    convertor = TroubleShootingExpConvertor(exp_lib_list, arg_cmd.suggestion_path,
                                            arg_cmd.regex_path, arg_cmd.fault_code_path)
    convertor.job()
