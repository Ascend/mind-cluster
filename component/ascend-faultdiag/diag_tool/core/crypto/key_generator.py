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
            use_*: 使用的字符集
        """
        # 构建字符集
        charset = list(chain(*self.CHARSETS.values()))
        password = []

        # 填充剩余字符
        remaining = length - len(password)
        for _ in range(remaining):
            password.append(secrets.choice(charset))

        # 随机打乱
        secrets.SystemRandom().shuffle(password)

        return ''.join(password)
