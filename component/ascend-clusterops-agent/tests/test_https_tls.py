#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Certificate logic: SAN assembly, DNS validation, self-signed generation/parsing, renewal.

These cases touch no Kubernetes, environment or filesystem, so no test doubles are needed.
"""

# pylint: disable=redefined-outer-name  # pytest fixtures share names with test parameters

import ipaddress
from datetime import datetime, timedelta, timezone

import pytest
from cryptography import x509
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import ec, rsa
from cryptography.x509.oid import ExtendedKeyUsageOID, NameOID

from agent_core.https import tls

NS = "mindx-dl"
IPS = {ipaddress.ip_address(value) for value in ("10.0.0.1", "10.0.0.2", "fd00::1")}
NOW = datetime(2026, 11, 18, tzinfo=timezone.utc)


@pytest.fixture(scope="module")
def key():
    return rsa.generate_private_key(public_exponent=65537, key_size=4096)


@pytest.fixture(autouse=True)
def shared_rsa_key(monkeypatch, key):
    """Sign for real but reuse one key, so RSA generation is not repeated per test."""
    monkeypatch.setattr(tls.rsa, "generate_private_key", lambda **_kwargs: key)


def required_with(extra_dns=()):
    return tls.build_required_sans(NS, IPS, extra_dns)


def build_certificate(cert_key, *, days=365, not_before_days=-1, complete_san=True, server_auth=True):
    not_before = NOW + timedelta(days=not_before_days)
    subject = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, "test-webui")])
    names = [x509.DNSName("custom.example")]
    if complete_san:
        names += [x509.IPAddress(ip) for ip in sorted(IPS, key=lambda ip: (ip.version, ip.packed))]
        names += [x509.DNSName(name) for name in tls.get_service_dns_names(NS)]
    builder = (
        x509.CertificateBuilder()
        .subject_name(subject)
        .issuer_name(subject)
        .public_key(cert_key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(not_before)
        .not_valid_after(not_before + timedelta(days=days))
        .add_extension(x509.SubjectAlternativeName(names), critical=False)
    )
    if server_auth:
        builder = builder.add_extension(x509.ExtendedKeyUsage([ExtendedKeyUsageOID.SERVER_AUTH]), critical=False)
    return builder.sign(cert_key, hashes.SHA256())


def pem(certificate) -> bytes:
    return certificate.public_bytes(serialization.Encoding.PEM)


def test_service_dns_names_cover_all_four_forms():
    assert tls.get_service_dns_names(NS) == {
        "agent-core-ui",
        "agent-core-ui.mindx-dl",
        "agent-core-ui.mindx-dl.svc",
        "agent-core-ui.mindx-dl.svc.cluster.local",
    }


@pytest.mark.parametrize("value", ["example.com", "a", "svc.cluster.local", "agent-core-ui.mindx-dl.svc"])
def test_valid_dns_names_are_accepted(value):
    assert tls.is_valid_dns_name(value)


@pytest.mark.parametrize(
    "value",
    ["", "-bad.example", "bad-.example", "bad/name", "bad name", "a..b", "中文.example", "x" * 64 + ".example"],
)
def test_invalid_dns_names_are_rejected(value):
    """Chinese domain names pass str.isalnum() but are invalid SANs, rejected here."""
    assert not tls.is_valid_dns_name(value)


def test_parse_san_value_returns_a_single_entry_set():
    """The return type matches what the caller accumulates, so no type dispatch is needed."""
    assert tls.parse_san_value("10.0.0.7") == tls.RequiredSanSet({ipaddress.ip_address("10.0.0.7")}, frozenset())
    assert tls.parse_san_value("extra.example") == tls.RequiredSanSet(frozenset(), {"extra.example"})
    assert tls.parse_san_value("bad/name") is None


def test_required_sans_union_keeps_service_names():
    required = required_with(extra_dns=["extra.example"])
    assert required.ip_addresses == frozenset(IPS)
    assert required.dns_names == frozenset(tls.get_service_dns_names(NS)) | {"extra.example"}


def test_generated_certificate_is_self_signed_with_required_extensions():
    material = tls.generate_tls_certificate("agent-core-ui.mindx-dl.svc", required_with(), NOW)
    certificate = material.certificate
    assert certificate.subject == certificate.issuer
    assert certificate.public_key().key_size == 4096
    assert certificate.signature_hash_algorithm.name == "sha256"
    assert certificate.extensions.get_extension_for_class(x509.ExtendedKeyUsage).value == x509.ExtendedKeyUsage(
        [ExtendedKeyUsageOID.SERVER_AUTH]
    )
    constraints = certificate.extensions.get_extension_for_class(x509.BasicConstraints).value
    assert constraints.ca is False
    san = certificate.extensions.get_extension_for_class(x509.SubjectAlternativeName).value
    assert set(san.get_values_for_type(x509.IPAddress)) == IPS
    assert set(san.get_values_for_type(x509.DNSName)) == tls.get_service_dns_names(NS)
    assert material.cert_pem == pem(certificate)
    assert b"PRIVATE KEY" in material.key_pem


def test_loaded_certificate_keeps_original_pem_bytes(key):
    certificate = build_certificate(key)
    material = tls.parse_tls_certificate(pem(certificate), key_pem_of(key))
    assert material.cert_pem == pem(certificate)
    assert material.certificate.subject == certificate.subject


def key_pem_of(private_key) -> bytes:
    return private_key.private_bytes(
        serialization.Encoding.PEM, serialization.PrivateFormat.PKCS8, serialization.NoEncryption()
    )


@pytest.mark.parametrize("case", ["garbage", "mismatch", "missing-eku"])
def test_invalid_certificate_material_is_rejected(key, case):
    if case == "garbage":
        cert_pem, key_pem = b"not a certificate", key_pem_of(key)
    elif case == "mismatch":
        cert_pem, key_pem = pem(build_certificate(key)), key_pem_of(ec.generate_private_key(ec.SECP256R1()))
    else:
        cert_pem, key_pem = pem(build_certificate(key, server_auth=False)), key_pem_of(key)
    with pytest.raises(tls.TlsMaterialError):
        tls.parse_tls_certificate(cert_pem, key_pem)


def test_fresh_complete_certificate_needs_no_renewal(key):
    assert tls.check_certificate_renewal(build_certificate(key), required_with(), NOW) is None


def test_expired_certificate_is_renewed(key):
    certificate = build_certificate(key, days=30, not_before_days=-400)
    assert tls.check_certificate_renewal(certificate, required_with(), NOW) is tls.CertificateRenewalReason.EXPIRED


def test_near_expiry_certificate_is_renewed(key):
    certificate = build_certificate(key, days=365, not_before_days=-345)
    assert tls.check_certificate_renewal(certificate, required_with(), NOW) is tls.CertificateRenewalReason.NEAR_EXPIRY


def test_missing_san_triggers_renewal(key):
    certificate = build_certificate(key, complete_san=False)
    assert tls.check_certificate_renewal(certificate, required_with(), NOW) is tls.CertificateRenewalReason.SAN_DRIFT


def test_not_yet_valid_certificate_is_reported_without_raising(key):
    """A not-yet-valid certificate is a material problem, returned like the other verdicts."""
    certificate = build_certificate(key, not_before_days=1)
    assert (
        tls.check_certificate_renewal(certificate, required_with(), NOW) is tls.CertificateRenewalReason.NOT_YET_VALID
    )


def test_renewal_window_uses_named_constant():
    """Pin the renewal window and validity period so the constants are not changed casually."""
    assert tls.RENEWAL_WARNING_DAYS == 30
    assert tls.CERTIFICATE_VALIDITY_DAYS == 365
