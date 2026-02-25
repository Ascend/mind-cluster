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
import base64
from toolkit.core.crypto.crypto import RootKeyCrypto


class TestRootKeyCrypto(unittest.TestCase):
    """测试RootKeyCrypto类的功能"""

    def setUp(self):
        """设置测试环境"""
        self.master_key = "test_master_key"
        self.test_data = "测试数据123!@#"
        self.long_test_data = "这是一个更长的测试数据，用于验证加密解密功能是否正常工作。" * 10

    def test_encrypt_decrypt(self):
        """测试加密解密功能"""
        crypto = RootKeyCrypto(self.master_key)
        encrypted = crypto.encrypt(self.test_data)
        decrypted = crypto.decrypt(encrypted)
        self.assertEqual(decrypted, self.test_data)

    def test_encrypt_decrypt_long_data(self):
        """测试长数据加密解密功能"""
        crypto = RootKeyCrypto(self.master_key)
        encrypted = crypto.encrypt(self.long_test_data)
        decrypted = crypto.decrypt(encrypted)
        self.assertEqual(decrypted, self.long_test_data)

    def test_encrypt_with_salt_decrypt_with_salt(self):
        """测试包含盐值的加密解密功能"""
        crypto = RootKeyCrypto(self.master_key)
        encrypted = crypto.encrypt_with_salt(self.test_data)
        decrypted = crypto.decrypt_with_salt(encrypted)
        self.assertEqual(decrypted, self.test_data)

    def test_encrypt_decrypt_with_custom_salt(self):
        """测试自定义盐值的加密解密功能"""
        custom_salt = base64.b64decode("dGVzdF9zYWx0MTIzNDU2Nzg5MA==")[:16]
        crypto = RootKeyCrypto(self.master_key, salt=custom_salt)
        encrypted = crypto.encrypt(self.test_data)
        decrypted = crypto.decrypt(encrypted)
        self.assertEqual(decrypted, self.test_data)

    def test_encrypt_decrypt_with_different_iterations(self):
        """测试不同迭代次数的加密解密功能"""
        crypto = RootKeyCrypto(self.master_key, iterations=50000)
        encrypted = crypto.encrypt(self.test_data)
        decrypted = crypto.decrypt(encrypted)
        self.assertEqual(decrypted, self.test_data)

    def test_string_master_key(self):
        """测试字符串类型的主密钥"""
        crypto = RootKeyCrypto(self.master_key)
        encrypted = crypto.encrypt(self.test_data)
        decrypted = crypto.decrypt(encrypted)
        self.assertEqual(decrypted, self.test_data)

    def test_bytes_master_key(self):
        """测试字节类型的主密钥"""
        bytes_master_key = ("ImKeyPart" + self.master_key).encode('utf-8')
        crypto = RootKeyCrypto(bytes_master_key)
        encrypted = crypto.encrypt(self.test_data)
        decrypted = crypto.decrypt(encrypted)
        self.assertEqual(decrypted, self.test_data)


if __name__ == "__main__":
    unittest.main()
