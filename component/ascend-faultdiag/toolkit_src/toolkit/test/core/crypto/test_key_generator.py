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

import unittest

from toolkit.core.crypto.key_generator import KeyGenerator


class TestKeyGenerator(unittest.TestCase):
    """测试KeyGenerator类的功能"""

    def setUp(self):
        """设置测试环境"""
        self.key_generator = KeyGenerator()

    def test_generate_complex_password_default_length(self):
        """测试默认长度的复杂密码生成"""
        password = self.key_generator.generate_complex_password()
        self.assertEqual(len(password), 16)
        self.assertIsInstance(password, str)

    def test_generate_complex_password_custom_length(self):
        """测试自定义长度的复杂密码生成"""
        custom_length = 24
        password = self.key_generator.generate_complex_password(length=custom_length)
        self.assertEqual(len(password), custom_length)
        self.assertIsInstance(password, str)

    def test_generate_complex_password_different_outputs(self):
        """测试生成的密码具有随机性（不同输出）"""
        password1 = self.key_generator.generate_complex_password()
        password2 = self.key_generator.generate_complex_password()
        self.assertNotEqual(password1, password2)

    def test_generate_complex_password_includes_different_charsets(self):
        """测试生成的密码包含不同类型的字符"""
        # 生成多个密码以增加测试的可信度
        for _ in range(10):
            password = self.key_generator.generate_complex_password(length=32)
            # 检查是否包含小写字母
            self.assertRegex(password, r'[a-z]')
            # 检查是否包含大写字母
            self.assertRegex(password, r'[A-Z]')
            # 检查是否包含数字
            self.assertRegex(password, r'[0-9]')
            # 检查是否包含特殊字符
            self.assertRegex(password, r'[!@#$%^&*()_+-=\[\]{}|;:,.<>?]')

    def test_generate_complex_password_min_length(self):
        """测试生成最小长度的密码"""
        min_length = 1
        password = self.key_generator.generate_complex_password(length=min_length)
        self.assertEqual(len(password), min_length)

    def test_generate_complex_password_long_length(self):
        """测试生成长长度的密码"""
        long_length = 128
        password = self.key_generator.generate_complex_password(length=long_length)
        self.assertEqual(len(password), long_length)


if __name__ == "__main__":
    unittest.main()
