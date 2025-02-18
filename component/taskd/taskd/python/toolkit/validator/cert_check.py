#!/usr/bin/python3
# -*- coding: utf-8 -*-

#   Copyright (C)  2025. Huawei Technologies Co., Ltd. All rights reserved.

import os
from datetime import datetime
from enum import Enum

from OpenSSL import crypto

from taskd.python.toolkit.logger.log import run_log


class PubkeyType(Enum):
    EVP_PKEY_RSA = 6
    EVP_PKEY_DSA = 116
    EVP_PKEY_DH = 28
    EVP_PKEY_EC = 408


class ParseCertInfo:
    """解析根证书信息类"""

    def __init__(self, cert_buff: bytes):
        if not cert_buff:
            raise ValueError("Cert buffer is null.")

        self.cert_info = crypto.load_certificate(crypto.FILETYPE_PEM, cert_buff)
        self.serial_num = hex(self.cert_info.get_serial_number())[2:].upper()
        self.subject_components = self.cert_info.get_subject().get_components()
        self.issuer_components = self.cert_info.get_issuer().get_components()
        self.fingerprint = self.cert_info.digest("sha256").decode()
        self.start_time = datetime.strptime(self.cert_info.get_notBefore().decode(), '%Y%m%d%H%M%SZ')
        self.end_time = datetime.strptime(self.cert_info.get_notAfter().decode(), '%Y%m%d%H%M%SZ')
        self.signature_algorithm = self.cert_info.get_signature_algorithm().decode()
        self.signature_len = self.cert_info.get_pubkey().bits()
        self.cert_version = self.cert_info.get_version() + 1
        self.pubkey_type = self.cert_info.get_pubkey().type()
        self.ca_pub_key = self.cert_info.get_pubkey().to_cryptography_key()
        self.extensions = {}
        for i in range(self.cert_info.get_extension_count()):
            ext = self.cert_info.get_extension(i)
            ext_name = ext.get_short_name().decode()
            try:
                self.extensions[ext_name] = str(ext)
            except Exception as e:
                run_log.warning(f"format '{ext_name}' str info in certificate failed: {e}")
                continue


class CertContentsChecker:
    X509_V3 = 3
    RSA_LEN_LIMIT = 3072
    # 椭圆曲线密钥长度
    EC_LEN_LIMIT = 256
    # 允许的签名算法
    SAFE_SIGNATURE_ALGORITHM = ("sha256WithRSAEncryption", "sha512WithRSAEncryption", "ecdsa-with-SHA256")

    def check_cert_info(self, cert_bytes: bytes):
        cert_info = ParseCertInfo(cert_bytes)
        time_now = datetime.utcnow()
        if time_now <= cert_info.start_time or time_now >= cert_info.end_time:
            raise ValueError("Cert contents checker: invalid cert validity period.")

        if cert_info.cert_version != self.X509_V3:
            raise ValueError(f"Cert contents checkers: check cert version '{cert_info.cert_version}' is not safe.")

        if cert_info.pubkey_type not in (PubkeyType.EVP_PKEY_RSA.value, PubkeyType.EVP_PKEY_EC.value):
            raise ValueError("Cert contents checkers: check cert pubkey type is not safe.")

        if cert_info.pubkey_type == PubkeyType.EVP_PKEY_RSA.value and cert_info.signature_len < self.RSA_LEN_LIMIT:
            raise ValueError("Cert contents checkers: check cert RSA pubkey length is not safe.")

        if cert_info.pubkey_type == PubkeyType.EVP_PKEY_EC.value and cert_info.signature_len < self.EC_LEN_LIMIT:
            raise ValueError("Cert contents checkers: check cert EC pubkey length is not safe.")

        if cert_info.signature_algorithm not in self.SAFE_SIGNATURE_ALGORITHM:
            raise ValueError("Cert contents checkers: check signature algorithm is not safe.")

        basic_constraints = cert_info.extensions.get("basicConstraints", "")
        if "CA:" not in basic_constraints:
            raise ValueError("Cert contents checkers: 'CA' not found in basic constraints.")

        key_usage = cert_info.extensions.get("keyUsage", "")
        if "Digital Signature" not in key_usage:
            raise ValueError("Cert contents checkers: 'Digital Signature' not found in key usage.")
        domain_name = ""
        for title, name in cert_info.subject_components:
            if title.decode() == "CN":
                domain_name = name.decode()
        return domain_name
