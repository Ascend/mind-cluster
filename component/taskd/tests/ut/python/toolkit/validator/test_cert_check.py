#!/usr/bin/python3
# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

from unittest.mock import patch, MagicMock
from unittest import TestCase
from taskd.python.toolkit.validator.cert_check import CertContentsChecker, ParseCertInfo
from datetime import datetime, timedelta

class TestParseCertInfo(TestCase):
    def test_init(self):
        cert_bytes = b'''
-----BEGIN CERTIFICATE-----
MIIFazCCA1OgAwIBAgIUWQZaTY2WwhGptBkIezChAjF2VE0wDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yNTA3MTIxMDA5MTFaFw0zNTA3
MTAxMDA5MTFaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQCyYKM3MTLY7QwTUVsTruye5czVReENzu+UGCJDPYLx
JOPJfh+Kz0NQePPDAVT58uAcs0VLtIR1mEe+JQWxPdyLU2wuxAA2KjQ2YdJmq8li
BPr1qBjTWNwHG4PbyFPRPtr2SZs1F3kTgTvy0YNczhh2iEVyUbAoZ1RL7L2FRZq/
ZjeZeTL0S5n+eGlwYyUN+lOnmvIWgmVn6XU+/rKYmKu6vPYG25So1IpoTPzoLCxJ
hqkCMJZdmpkq0iXpwJy6NqWgug7axsVNVzXBth1JOxFHG4EI+VRaBhJ0GtO7xZUO
UgZzQIts8J/QIJrVKhNBTzV5oHVN75msBAZ9q2GWBgxJuIzyvK9rhi7qysPc42Vb
UIHdbPCwV6aVpy7EdXEQ6cbSIIW3Z7JqeslJ/BUMwd/9vRn35pnd5z/M5UQgrQ5C
2jB0zAmr3o4h4Pn2Y80LUNcrIkohMceDevr9IPcF01v9HnAAyDIy00hdMAZbiSlB
+TCpjBMXXEL3CIEXqKkxXktnaFPf/MtznlAzbjezhIcvaKDPXVZ93O/LARWafqOB
V94wbvDGdgy75axT/zZvIVgYHf8ShgR+K3cSK4TykL2K4gVABQr0jAAqavB2L/X0
emD2MO3shAJkU5J4r0Mn14b17FxYZwvlwg3H9Tgbcb5s1BuVGs3h185uhmhq0Z/w
VQIDAQABo1MwUTAdBgNVHQ4EFgQUzJWFzidbqglAZRcHnY14t8mD6VEwHwYDVR0j
BBgwFoAUzJWFzidbqglAZRcHnY14t8mD6VEwDwYDVR0TAQH/BAUwAwEB/zANBgkq
hkiG9w0BAQsFAAOCAgEAETMGsj7Nw/r8riNk/FXpyK1XehuC7N9UpBy3EH72Yd0P
q+ppCcP6FUozuN7eSmOA441GT9ouapisK3q/8cXN1QIG0Em2Cr97cHyQQX9RhQZY
PaUf76ymC8+iFzr2+Lf+rq22SdyoohytpvxKSBFr6uEq1M4bekenjFUBKNNLGkor
z7AFmMvxwedc6A7YD6zwrnM4igWGQ6dCyA01lsOYd9kQ7d25J9OPzLCA/LJSxZWP
kq3hDpDPLq8K5fta6oQpQgONYNDd5ObfxdU4YANww4PvjGMH4vB239Mo4hd+WgLq
zaNPQCgOzqAX67/XvVI8FpyflV6687l4xn4UF9HQ1B7tqzVolrBQw3r1VWTO351Q
8o+NrW+jfFOhbJfDxsFsabVrkJlfvxX16cB8QfqR++TcuDNZvf5SpPmdEZoIN7Yn
Dv5t0wGO95HPIXmOh1HxAm3V0ZtDRlJkO8hIEr9RYsfv0tbPAXzt+cV3MnVANW1V
9OttH6rW2VpGLG3hR2DMBJxv5rdLHY6yv6uICDdLJtgc1bl9avenRB7AVrWhJcW4
1qPRa++ag1B0ufTMMgrjNoqhLxIY4iT/o8ABPCHix4XBbtTYVHPGmpgFBBzxiqvb
6X1otVeY2WF+z2KR6nFLJBU5aGQlHYGCYm1vcD+qdB6BvZICo3MIxC2BPF2a/ig=
-----END CERTIFICATE-----
'''
        parser = ParseCertInfo(cert_bytes)
        self.assertEqual(len(parser.extensions), 3)


class TestCertContentsChecker(TestCase):

    @patch('taskd.python.toolkit.validator.cert_check.ParseCertInfo')
    def test_check_cert_info(self, mock_parse_cert_info):
        cert_checker = CertContentsChecker()
        cert_bytes = b'test_cert_bytes'

        # Mocking the ParseCertInfo class
        mock_cert_info = MagicMock()
        mock_parse_cert_info.return_value = mock_cert_info

        # Testing valid certificate
        mock_cert_info.start_time = datetime.utcnow() - timedelta(days=1)
        mock_cert_info.end_time = datetime.utcnow() + timedelta(days=1)
        mock_cert_info.cert_version = cert_checker.X509_V3
        mock_cert_info.pubkey_type = 6
        mock_cert_info.signature_len = cert_checker.RSA_LEN_LIMIT + 1
        mock_cert_info.signature_algorithm = 'sha256WithRSAEncryption'
        mock_cert_info.extensions = {'basicConstraints': 'CA:TRUE', 'keyUsage': 'Digital Signature'}
        title = "CN"
        title_byte = title.encode('utf-8')
        name = "test.com"
        name_byte = name.encode('utf-8')
        mock_cert_info.subject_components = [(title_byte, name_byte)]

        domain_name = cert_checker.check_cert_info(cert_bytes)
        self.assertEqual(domain_name, 'test.com')

        # Testing invalid certificate validity period
        mock_cert_info.start_time = datetime.utcnow() + timedelta(days=1)
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing invalid certificate version
        mock_cert_info.start_time = datetime.utcnow() - timedelta(days=1)
        mock_cert_info.cert_version = 'Invalid Version'
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing invalid certificate pubkey type
        mock_cert_info.cert_version = cert_checker.X509_V3
        mock_cert_info.pubkey_type = 'Invalid Type'
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing invalid RSA pubkey length
        mock_cert_info.pubkey_type = 'EVP_PKEY_RSA'
        mock_cert_info.signature_len = cert_checker.RSA_LEN_LIMIT - 1
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing invalid EC pubkey length
        mock_cert_info.pubkey_type = 'EVP_PKEY_EC'
        mock_cert_info.signature_len = cert_checker.EC_LEN_LIMIT - 1
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing invalid signature algorithm
        mock_cert_info.signature_len = cert_checker.RSA_LEN_LIMIT + 1
        mock_cert_info.signature_algorithm = 'Invalid Algorithm'
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing missing 'CA' in basic constraints
        mock_cert_info.signature_algorithm = 'SHA256'
        mock_cert_info.extensions = {'basicConstraints': '', 'keyUsage': 'Digital Signature'}
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)

        # Testing missing 'Digital Signature' in key usage
        mock_cert_info.extensions = {'basicConstraints': 'CA:TRUE', 'keyUsage': ''}
        with self.assertRaises(ValueError):
            cert_checker.check_cert_info(cert_bytes)