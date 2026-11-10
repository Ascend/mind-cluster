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

"""
简单安全的主密钥生成模块
提供多种密钥生成方式，不依赖外部加密库
"""

import secrets
from itertools import chain


class KeyGenerator:
    """简单安全的主密钥生成器"""

    def __init__(self):
        # 字符集定义
        self.CHARSETS = {
            'lower': 'abcdefghijklmnopqrstuvwxyz',
            'upper': 'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
            'digits': '0123456789',
            'special': '!@#$%^&*()_+-=[]{}|;:,.<>?',
            'hex': '0123456789abcdef',
        }

    def generate_complex_password(self, length: int = 16) -> str:
        """
        生成复杂密码
        参数:
            length: 密码长度
        """
        # 构建字符集
        charset = list(chain(*self.CHARSETS.values()))
        required_charsets = [
            self.CHARSETS['lower'],
            self.CHARSETS['upper'],
            self.CHARSETS['digits'],
            self.CHARSETS['special'],
        ]
        if length < len(required_charsets):
            # 长度不足以覆盖全部字符集时，直接从总字符集随机生成
            return ''.join(secrets.choice(charset) for _ in range(length))

        # 每个字符集至少取一个字符，保证密码复杂度
        password = [secrets.choice(charset_item) for charset_item in required_charsets]

        # 填充剩余字符
        remaining = length - len(password)
        for _ in range(remaining):
            password.append(secrets.choice(charset))

        # 随机打乱
        secrets.SystemRandom().shuffle(password)

        return ''.join(password)
