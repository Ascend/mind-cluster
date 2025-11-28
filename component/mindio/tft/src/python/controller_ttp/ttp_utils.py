#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
import os
import site
import ttp_logger


def input_ip_transform(input_ip: str):
    if is_valid_ip(input_ip) or input_ip == '':
        return input_ip

    try:
        import socket
        ip = socket.gethostbyname(input_ip)
        ttp_logger.LOGGER.info(f"transform {input_ip} to {ip}")
        return ip
    except socket.error as e:
        ttp_logger.LOGGER.error(f"input neither ip nor hostname, {str(e)}")
        return input_ip


def is_valid_ip(ip_str):
    if len(ip_str) > 15:
        ttp_logger.LOGGER.warning(f"input illegal ipv4 address: length exceeds 15.")
        return False
    import re
    ip_pattern = r'(^((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}$)'
    return bool(re.match(ip_pattern, ip_str))


def is_zero_ip(ip_str):
    # if ip_str exceeds max DOMAIN length, directly regard as all-zero ip.
    if len(ip_str) > 253:
        ttp_logger.LOGGER.warning(f"input illegal ipv4 address or domain: length exceeds 253.")
        return True
    import re
    zero_pattern = '^0+\\.0+\\.0+\\.0+$'
    return bool(re.match(zero_pattern, ip_str))


def get_env_var_int_safely(env_var_name, default_value, min_value=None, max_value=None):
    try:
        temp_var = os.getenv(env_var_name, str(default_value))
        if len(temp_var) > 1000:
            temp_var = default_value
        temp_var = int(temp_var)
        if min_value is not None and temp_var < min_value:
            ttp_logger.LOGGER.warning(f"Environment variable {env_var_name} is below the minimum value, "
                                      f"using default value: {default_value}")
            return default_value
        if max_value is not None and temp_var > max_value:
            ttp_logger.LOGGER.warning(f"Environment variable {env_var_name} is above the maximum value, "
                                      f"using default value: {default_value}")
            return default_value
        result = temp_var
        return result
    except Exception as e:
        ttp_logger.LOGGER.warning(f"invalid env variable:{env_var_name}, use default value {default_value}: {e}")
        return default_value
