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

import re
from typing import List, Dict, Tuple


class TableParser(object):

    @staticmethod
    def format_titles_dict(titles_dict):
        if isinstance(titles_dict, list) or isinstance(titles_dict, set) or isinstance(titles_dict, tuple):
            return {value: value for value in titles_dict}
        return titles_dict

    @staticmethod
    def _parse_value(src_line: str, scope: Tuple[int, int], col_separator: str, both_strip: bool) -> str:
        value = src_line[scope[0]:scope[1]].rstrip().strip(col_separator).rstrip()
        if both_strip:
            value = value.strip()
        return value

    @classmethod
    def parse(cls, recv: str, titles_dict: dict, regex_replace_dict: dict = None, separate_title_content_lines_num=0,
              end_sign='', col_separator="", both_strip=True) -> List[Dict]:
        """
        将内容和标题左对齐的命令行列表式的回显处理成字典列表的filter
        @param recv 回显
        @param titles_dict 参数标题字典，key为预期的字段名，title为预期的标题名
        @param regex_replace_dict 正则替换的字典
        @param separate_title_content_lines_num 标题与内容间隔几行，一般0或1
        @param end_sign 结束标记
        @param col_separator 列间隔符号
        解析与标题名称左对齐的命令行

        入参样例
        recv = '
            (R) Router, (B) Bridge, (T) Telephone, (C) DOCSIS Cable Device
            (W) WLAN Access Point, (P) Repeater, (S) Station, (O) Other
            System Name                 Local Intf          Port ID                          Capability   Aging-time
            RJEOR02                     TF0/1               TFGigabitEthernet 0/1            P, B, R      1minutes 35seconds
            RJEOR02                     TF0/2               TFGigabitEthernet 0/2            P, B, R      1minutes 35seconds
            Total entries displayed: 2
        '
        titles_dict = {'system_name': 'System Name', 'local_intf': 'Local Intf', 'port_id': 'Port ID',
        'capability': 'Capability', 'aging_time': 'Aging-time'}

        regex_replace_dict = {'local_intf': {'pattern': '\s', 'repl': ''},
                                port_id': {'pattern': '[^0-9a-zA-Z]','repl': ''}}

        end_sign = 'Total entries displayed'

        输出样例
        [{
            'system_name': 'RJEOR02',
            'local_intf': 'TF0/1',
            'port_id': 'TFGigabitEthernet 0/1',
            'capability': 'P, B, R',
            'aging_time': '1minutes 35seconds'
        }, {
            'system_name': 'RJEOR02',
            'local_intf': 'TF0/2',
            'port_id': 'TFGigabitEthernet 0/2',
            'capability': 'P, B, R',
            'aging_time': '1minutes 35seconds'
        }]
        """
        if not recv.strip():
            return []
        regex_replace_dict = regex_replace_dict or {}
        titles_dict = cls.format_titles_dict(titles_dict)
        lines = [line.rstrip() for line in re.split('[\r\n]+', recv) if line.rstrip()]
        if end_sign and recv.find(end_sign) > -1:
            index = [index for index, line in enumerate(lines) if line.find(end_sign) > -1][-1]
            lines = lines[:index]
        title_size = len(titles_dict)
        title_line, title_line_index = cls.get_title_start_line(title_size,
                                                                titles_dict, lines)
        if title_line_index < 0:
            return []
        title_line_index += separate_title_content_lines_num
        if title_line_index + 1 >= len(lines):
            return []
        titles_start_index_list = [title_line.find(item) for item in titles_dict.values()] + [None]
        titles_scope_dict = {}
        for index, key in enumerate(titles_dict.keys()):
            titles_scope_dict.update({key: (titles_start_index_list[index], titles_start_index_list[index + 1])})

        result_list = [{k: cls._parse_value(line, v, col_separator, both_strip) for k, v in titles_scope_dict.items()}
                       for line in lines[title_line_index + 1:]]
        return cls.regex_replace_dict_value(result_list, regex_replace_dict)

    @classmethod
    def get_title_start_line(cls, title_size, titles_dict, lines: List[str]):
        title_line = ''
        title_line_index = -1
        for index, line in enumerate(lines):
            find_title_result = [True for item in titles_dict.values() if item and line.find(item) > -1]
            if len(find_title_result) == title_size:
                title_line = line
                title_line_index = index
                break
        return title_line, title_line_index

    @classmethod
    def regex_replace_dict_value(cls, result_list, regex_replace_dict):
        if not regex_replace_dict:
            return result_list
        for result in result_list:
            for title_key, regex_info in regex_replace_dict.items():
                result[title_key] = re.sub(regex_info['pattern'], regex_info['repl'], result[title_key])
        return result_list
