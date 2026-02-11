import base64
import os

from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC
from cryptography.hazmat.backends import default_backend
import hashlib
import hmac


class RootKeyCrypto:
    _STATIC_KEY_PART = "ImKeyPart"

    def __init__(self, master_key, salt=None, iterations=100000):
        """
        初始化加解密器

        参数:
            master_key: 固定根密钥（字符串或字节）
            salt: 盐值（可选，默认随机生成）
            iterations: PBKDF2迭代次数
        """
        if isinstance(master_key, str):
            master_key = (self._STATIC_KEY_PART + master_key).encode('utf-8')

        self.master_key = master_key
        self.salt = salt or self._generate_random_component()
        self.iterations = iterations

        # 使用PBKDF2从主密钥派生密钥
        kdf = PBKDF2HMAC(
            algorithm=hashes.SHA256(),
            length=32,  # 32字节 = 256位
            salt=self.salt,
            iterations=self.iterations,
            backend=default_backend()
        )
        self.derived_key = kdf.derive(self.master_key)

    @staticmethod
    def _generate_random_component(length=16):
        """生成随机部分"""
        return os.urandom(length)

    def encrypt(self, plaintext):
        """
        加密数据

        参数:
            plaintext: 要加密的文本（字符串）

        返回:
            包含随机部分、IV和密文的base64编码字符串
        """
        if isinstance(plaintext, str):
            plaintext = plaintext.encode('utf-8')

        # 1. 生成随机部分
        random_component = self._generate_random_component()

        # 2. 生成最终加密密钥
        encryption_key = self._combine_keys(random_component)

        # 3. 生成IV
        iv = os.urandom(16)

        # 4. 加密数据
        cipher = Cipher(
            algorithms.AES(encryption_key),
            modes.GCM(iv),
            backend=default_backend()
        )
        encryptor = cipher.encryptor()

        ciphertext = encryptor.update(plaintext) + encryptor.finalize()

        # 5. 组合所有组件：随机部分 + IV + 密文 + 认证标签
        combined = (
                random_component +
                iv +
                ciphertext +
                encryptor.tag
        )

        # 6. 返回base64编码结果
        return base64.b64encode(combined).decode('utf-8')

    def decrypt(self, encrypted_data):
        """
        解密数据

        参数:
            encrypted_data: 加密后的base64编码字符串

        返回:
            解密后的文本（字符串）
        """
        # 1. 解码base64
        data = base64.b64decode(encrypted_data)

        # 2. 解析组件
        random_component = data[:16]  # 16字节随机部分
        iv = data[16:32]  # 16字节IV
        ciphertext = data[32:-16]  # 密文
        tag = data[-16:]  # 16字节认证标签

        # 3. 重新生成加密密钥
        encryption_key = self._combine_keys(random_component)

        # 4. 解密数据
        cipher = Cipher(
            algorithms.AES(encryption_key),
            modes.GCM(iv, tag),
            backend=default_backend()
        )
        decryptor = cipher.decryptor()

        plaintext = decryptor.update(ciphertext) + decryptor.finalize()

        return plaintext.decode('utf-8')

    def encrypt_with_salt(self, plaintext):
        """
        加密并包含盐值（用于长期存储）
        """
        ciphertext = self.encrypt(plaintext)

        # 将盐值和密文组合
        combined = self.salt + ciphertext.encode('utf-8')
        return base64.b64encode(combined).decode('utf-8')

    def decrypt_with_salt(self, encrypted_data):
        """
        解密包含盐值的数据
        """
        data = base64.b64decode(encrypted_data)

        # 分离盐值和密文
        salt = data[:16]
        ciphertext = data[16:].decode('utf-8')

        # 使用相同的盐值初始化新的实例
        temp_crypto = RootKeyCrypto(
            master_key=self.master_key,
            salt=salt,
            iterations=self.iterations
        )

        return temp_crypto.decrypt(ciphertext)

    def _combine_keys(self, random_component):
        """
        结合固定密钥和随机部分生成最终加密密钥
        使用HMAC确保完整性
        """
        # 使用HMAC-SHA256结合密钥
        hmac_obj = hmac.new(
            self.derived_key,
            random_component,
            hashlib.sha256
        )
        return hmac_obj.digest()
