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
import string

from ascend_fd.utils.comm_valid import char_check, PARA_WHITE_LIST, LINE_WHITE_LIST
from ascend_fd.utils.tool import CODE_WHITE_LIST
from ascend_fd.sdk.fd_tool import FDTool
from ascend_fd.utils.i18n import LANG

RULE_DST_NAME = "dst_code"
RULE_EXP_NAME = "expression"


def code_check(input_value):
    """
    Verify the validity of code value
    Used by [code]
    :param input_value: str, fault code value
    :return: Indicates whether the value is valid. bool value.
    """
    return char_check(input_value, length_range=(1, 50), white_list=CODE_WHITE_LIST)


def paragraph_check(input_value):
    """
    Verify the validity of paragraph
    Used by [attribute.description, attribute.suggestion, attribute.error_case, attribute.fixed_case]
    :param input_value: Union[str, List[str]], paragraph
    :return: Indicates whether the value is valid. bool value.
    """
    if isinstance(input_value, str):
        # Only digits, uppercase and lowercase letters, punctuations, " ", "\n" and Chinese characters.
        return char_check(input_value, length_range=(1, 2000), white_list=PARA_WHITE_LIST, allow_zh=True)
    if isinstance(input_value, list):
        # Only digits, uppercase and lowercase letters, punctuations, " " and Chinese characters.
        for single_str in input_value:
            if not char_check(single_str, length_range=(1, 200), white_list=LINE_WHITE_LIST, allow_zh=True):
                return False
        return True
    return False


def rule_check(code, input_value, all_user_entities):
    """
    Verify the validity of rule list.
    rule = [{'dst_code': 'xxxxx', 'expression': 'xxxxxx'}...]
    Used by [rule]
    :param code: customized entity code
    :param input_value: List[Dict[str, str]], rule list
    :param all_user_entities: set, all entity codes set, contain origin_entity and user_entity
    :return: Indicates whether the value is valid. bool value.
    """
    if not isinstance(input_value, list):
        return False
    for single_rule in input_value:
        if not isinstance(single_rule, dict):
            return False
        if RULE_DST_NAME not in single_rule:
            return False
        dst_code = single_rule.get(RULE_DST_NAME)
        if dst_code not in all_user_entities and not FDTool().is_code_exist(dst_code):
            return False
        if dst_code == code:
            return False
        if RULE_EXP_NAME in single_rule and \
                not char_check(single_rule.get(RULE_EXP_NAME), length_range=(1, 200), white_list=LINE_WHITE_LIST):
            return False
    return True


def source_check(input_value, source_file_num=10):
    """
    Verify the validity of source_file value
    Used by [source_file]
    :param input_value: str, source_file value
    :param source_file_num: int, the number of source_file
    :return: Indicates whether the value is valid. bool value.
    """
    value_list = input_value.split("|")
    if len(value_list) > source_file_num:
        return False
    for each_str in value_list:
        if not char_check(each_str, length_range=(1, 50), white_list=LINE_WHITE_LIST, allow_zh=False):
            return False
    return True


def in_check(input_value):
    """
    Verify the validity of "in" rule
    Used by [regex.in]
    :param input_value: Union[List[str], List[List[str]]], "in" rule list
    :return: Indicates whether the value is valid. bool value.
    """
    if not isinstance(input_value, list):
        return False
    if not input_value:
        return False
    if not isinstance(input_value[0], list):
        input_value = [input_value]
    for single_value in input_value:
        if not single_value or not isinstance(single_value, list):
            return False
        for each_str in single_value:
            if not char_check(each_str, length_range=(1, 200), white_list=LINE_WHITE_LIST, allow_zh=True):
                return False
    return True


def line_check_func_factor(length_range, allow_zh=False):
    """
    Produce a character detection function
    :param length_range: Tuple[int, int], string length range
    :param allow_zh: bool, whether Chinese characters are supported
    :return: function
    """

    def check_func(input_value):
        return char_check(input_value, length_range=length_range, white_list=LINE_WHITE_LIST, allow_zh=allow_zh)

    return check_func


def check_missing_attribute_when_add(contain_attribute):
    """
    Check the missing attributes when add a new fault code
    :param contain_attribute: set, the set of contain attribute
    :return: the missing required attributes set.
    """
    required_attributes = {
        "attribute.class", "attribute.component", "attribute.module", f"attribute.cause_{LANG}",
        f"attribute.description_{LANG}", f"attribute.suggestion_{LANG}", "source_file", "regex.in"
    }
    return required_attributes - contain_attribute


# rule check func need two parameters
CHECK_MAP = {
    "attribute.class": line_check_func_factor(length_range=(1, 50)),
    "attribute.component": line_check_func_factor(length_range=(1, 50)),
    "attribute.module": line_check_func_factor(length_range=(1, 50)),
    "attribute.cause_zh": line_check_func_factor(length_range=(1, 200), allow_zh=True),
    "attribute.description_zh": paragraph_check,
    "attribute.suggestion_zh": paragraph_check,
    "attribute.cause_en": line_check_func_factor(length_range=(1, 200)),
    "attribute.description_en": paragraph_check,
    "attribute.suggestion_en": paragraph_check,
    "attribute.error_case": paragraph_check,
    "attribute.fixed_case": paragraph_check,
    "rule": rule_check,
    "source_file": source_check,
    "regex.in": in_check
}
