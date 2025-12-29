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

from ascend_fd.utils.tool import white_check

LINE_WHITE_LIST = string.digits + string.ascii_letters + string.punctuation + " "
PARA_WHITE_LIST = string.digits + string.ascii_letters + string.punctuation + " \n"


def valid_int(attr: str, val: int):
    """
    Validate integer greater than 0
    :param attr: attribute name
    :param val: attribute value
    """
    if val <= 0:
        raise ValueError(f"Invalid value for {attr}: {val}. Must be an integer greater than 0.")


def valid_list_len(attr: str, val: list, max_len=10):
    """
    Validate integer greater than 0
    :param attr: attribute name
    :param val: attribute value
    :param max_len: max len of val
    """
    if len(val) > max_len:
        raise ValueError(f"Invalid value for {attr}: {val}. Supports up to {max_len}.")


def char_check(input_value, length_range, white_list, allow_zh=False):
    """
    Verify the character validity of any string. Default allow white_list char, support chinese char by allow_zh para
    :param input_value: str, input string value
    :param length_range: Tuple[int, int], string length range
    :param white_list: str, char white list
    :param allow_zh: bool, whether Chinese characters are supported
    :return: Indicates whether the value is valid. bool value.
    """
    if not isinstance(input_value, str):
        return False
    if len(input_value) < length_range[0] or len(input_value) > length_range[1]:
        return False
    return white_check(input_value, white_list, allow_zh)
