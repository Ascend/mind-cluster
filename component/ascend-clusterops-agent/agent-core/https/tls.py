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

"""Pure WebUI HTTPS certificate logic: SAN sets, validation, self-signed generation, renewal.

No environment, cluster access or file writes; Secret I/O lives in ``secret_bootstrap``.
"""

from __future__ import annotations

import ipaddress
import re
from dataclasses import dataclass
from datetime import datetime, timedelta
from enum import Enum
from typing import FrozenSet, Iterable, Optional, Set, Union

from cryptography import x509
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.x509.oid import ExtendedKeyUsageOID, NameOID

IPAddress = Union[ipaddress.IPv4Address, ipaddress.IPv6Address]

SERVICE_NAME = "agent-core-ui"
CERTIFICATE_VALIDITY_DAYS = 365
RENEWAL_WARNING_DAYS = 30
RSA_KEY_BITS = 4096
RSA_PUBLIC_EXPONENT = 65537
MAX_DNS_NAME_LENGTH = 253
MAX_DNS_LABEL_LENGTH = 63
_DNS_LABEL_PATTERN = re.compile(r"[A-Za-z0-9]([-A-Za-z0-9]*[A-Za-z0-9])?")


@dataclass(frozen=True)
class RequiredSanSet:
    """SANs the certificate must cover: node InternalIPs, Service DNS names and extra entries."""

    ip_addresses: FrozenSet[IPAddress]
    dns_names: FrozenSet[str]


@dataclass(frozen=True)
class TlsKeypair:
    """A validated certificate and key pair; the PEM bytes are written as-is."""

    certificate: "x509.Certificate"
    cert_pem: bytes
    key_pem: bytes


class TlsMaterialError(RuntimeError):
    """The certificate material is unusable. Messages are fixed and never contain keys."""


class CertificateRenewalReason(Enum):
    """Certificate verdict; the member order is the priority order.

    NOT_YET_VALID is a material problem that renewal cannot fix, so HTTPS stays off.
    The other three are renewable: validity window or missing SANs.
    """

    NOT_YET_VALID = "not-yet-valid"
    EXPIRED = "expired"
    NEAR_EXPIRY = "near-expiry"
    SAN_DRIFT = "san-drift"


def get_service_dns_names(namespace: str) -> Set[str]:
    """The four resolvable forms of the Service name inside its namespace."""
    qualified = f"{SERVICE_NAME}.{namespace}"
    return {SERVICE_NAME, qualified, f"{qualified}.svc", f"{qualified}.svc.cluster.local"}


def is_valid_dns_name(value: str) -> bool:
    """Validate DNS length and each label; Unicode and other invalid characters are rejected."""
    if not value or len(value) > MAX_DNS_NAME_LENGTH:
        return False
    for label in value.split("."):
        if not 1 <= len(label) <= MAX_DNS_LABEL_LENGTH:
            return False
        if _DNS_LABEL_PATTERN.fullmatch(label) is None:
            return False
    return True


def parse_san_value(value: str) -> Optional[RequiredSanSet]:
    """Parse one SAN entry into a set holding only that entry; None when invalid.

    The return type matches what the caller accumulates, so no type dispatch is needed.
    """
    try:
        return RequiredSanSet(ip_addresses=frozenset({ipaddress.ip_address(value)}), dns_names=frozenset())
    except ValueError:
        if not is_valid_dns_name(value):
            return None
        return RequiredSanSet(ip_addresses=frozenset(), dns_names=frozenset({value}))


def build_required_sans(
    namespace: str,
    extra_ip_addresses: Iterable[IPAddress] = (),
    extra_dns_names: Iterable[str] = (),
) -> RequiredSanSet:
    """Build the full required SAN set: extra entries plus the four Service DNS names."""
    return RequiredSanSet(
        ip_addresses=frozenset(extra_ip_addresses),
        dns_names=frozenset(extra_dns_names) | frozenset(get_service_dns_names(namespace)),
    )


def parse_tls_certificate(cert_pem: bytes, key_pem: bytes) -> TlsKeypair:
    """Parse and validate the certificate and key: valid PEM, matching key, serverAuth EKU."""
    try:
        certificate = x509.load_pem_x509_certificate(cert_pem)
        private_key = serialization.load_pem_private_key(key_pem, password=None)
        cert_public = certificate.public_key().public_bytes(
            serialization.Encoding.DER, serialization.PublicFormat.SubjectPublicKeyInfo
        )
        key_public = private_key.public_key().public_bytes(
            serialization.Encoding.DER, serialization.PublicFormat.SubjectPublicKeyInfo
        )
        if cert_public != key_public:
            raise TlsMaterialError("TLS certificate and private key do not match")
        usages = certificate.extensions.get_extension_for_class(x509.ExtendedKeyUsage).value
        if ExtendedKeyUsageOID.SERVER_AUTH not in usages:
            raise TlsMaterialError("TLS certificate lacks serverAuth EKU")
    except TlsMaterialError:
        raise
    except Exception:
        raise TlsMaterialError("TLS PEM, private key or serverAuth EKU is invalid") from None
    return TlsKeypair(certificate=certificate, cert_pem=cert_pem, key_pem=key_pem)


def get_certificate_valid_from(certificate: "x509.Certificate") -> datetime:
    """Certificate start of validity (UTC)."""
    return certificate.not_valid_before_utc


def get_certificate_expiry(certificate: "x509.Certificate") -> datetime:
    """Certificate expiry (UTC)."""
    return certificate.not_valid_after_utc


def has_missing_required_sans(certificate: "x509.Certificate", required: RequiredSanSet) -> bool:
    """Whether the certificate misses a required SAN (node growth or a Service rename)."""
    try:
        names = certificate.extensions.get_extension_for_class(x509.SubjectAlternativeName).value
        present_ips = set(names.get_values_for_type(x509.IPAddress))
        present_dns = set(names.get_values_for_type(x509.DNSName))
    except x509.ExtensionNotFound:
        present_ips, present_dns = set(), set()
    return not required.ip_addresses.issubset(present_ips) or not required.dns_names.issubset(present_dns)


def check_certificate_renewal(
    certificate: "x509.Certificate", required: RequiredSanSet, now: datetime
) -> Optional[CertificateRenewalReason]:
    """Return the certificate verdict; None means it is usable as-is.

    Every conclusion is a return value: the caller decides to reuse, renew or disable
    HTTPS. A not-yet-valid certificate is reported separately because renewal cannot fix it.
    """
    if get_certificate_valid_from(certificate) > now:
        return CertificateRenewalReason.NOT_YET_VALID
    not_after = get_certificate_expiry(certificate)
    if not_after <= now:
        return CertificateRenewalReason.EXPIRED
    if not_after - now < timedelta(days=RENEWAL_WARNING_DAYS):
        return CertificateRenewalReason.NEAR_EXPIRY
    if has_missing_required_sans(certificate, required):
        return CertificateRenewalReason.SAN_DRIFT
    return None


def generate_tls_certificate(
    common_name: str,
    required: RequiredSanSet,
    now: datetime,
    validity_days: int = CERTIFICATE_VALIDITY_DAYS,
) -> TlsKeypair:
    """Generate a self-signed leaf certificate (RSA, SHA-256, serverAuth, not a CA)."""
    private_key = rsa.generate_private_key(public_exponent=RSA_PUBLIC_EXPONENT, key_size=RSA_KEY_BITS)
    subject = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, common_name)])
    san = [x509.IPAddress(ip) for ip in sorted(required.ip_addresses, key=lambda ip: (ip.version, ip.packed))]
    san += [x509.DNSName(name) for name in sorted(required.dns_names)]
    certificate = (
        x509.CertificateBuilder()
        .subject_name(subject)
        .issuer_name(subject)
        .public_key(private_key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(now)
        .not_valid_after(now + timedelta(days=validity_days))
        .add_extension(x509.SubjectAlternativeName(san), critical=False)
        .add_extension(x509.ExtendedKeyUsage([ExtendedKeyUsageOID.SERVER_AUTH]), critical=False)
        .add_extension(x509.BasicConstraints(ca=False, path_length=None), critical=True)
        .sign(private_key, hashes.SHA256())
    )
    cert_pem = certificate.public_bytes(serialization.Encoding.PEM)
    key_pem = private_key.private_bytes(
        serialization.Encoding.PEM, serialization.PrivateFormat.PKCS8, serialization.NoEncryption()
    )
    return TlsKeypair(certificate=certificate, cert_pem=cert_pem, key_pem=key_pem)


def build_certificate_common_name(namespace: str) -> str:
    """Common name of the self-signed certificate: the in-cluster Service FQDN."""
    return f"{SERVICE_NAME}.{namespace}.svc"
