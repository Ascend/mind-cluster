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

"""WebUI HTTPS material bootstrap: the certificate and the shared token.

Both materials are independent; any failure returns None and only HTTPS is disabled.
Decisions return None, parsing and cluster I/O raise fixed-message errors, and
``_run_bootstrap_step`` is the only catch point. Logs never carry secret material.
"""

from __future__ import annotations

import base64
import binascii
import ipaddress
import logging
import os
import secrets
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Mapping, Optional

from agent_core.k8s import K8s
from agent_core.https.tls import (
    RequiredSanSet,
    CertificateRenewalReason,
    TlsKeypair,
    TlsMaterialError,
    build_certificate_common_name,
    build_required_sans,
    check_certificate_renewal,
    generate_tls_certificate,
    parse_san_value,
    parse_tls_certificate,
)

logger = logging.getLogger(__name__)

TLS_SECRET_NAME = "agent-core-webui-tls"  # nosec B105
TOKEN_SECRET_NAME = "agent-core-webui-token"  # nosec B105
TLS_SANS_ENV = "WEBUI_TLS_SANS"
HOST_IP_ENV = "HOST_IP"
MAX_TLS_SANS = 100
TOKEN_BYTES = 32
TOKEN_GROUP_SIZE = 4
KUBERNETES_API_TIMEOUT_SECONDS = 5
TLS_MATERIAL_DIRECTORY = Path("/tmp/webui-tls")  # nosec B108
PRIVATE_DIRECTORY_MODE = 0o700
PRIVATE_FILE_MODE = 0o600


@dataclass(frozen=True)
class StoredSecret:
    """Normalized Secret view: decoded bytes plus the resourceVersion needed for renewal."""

    data: Mapping[str, bytes]
    resource_version: str


@dataclass(frozen=True)
class HttpsMaterial:
    """Runtime HTTPS material: local PEM paths plus the shared token (kept out of repr)."""

    certfile: Path
    keyfile: Path
    token: str = field(repr=False)


class HttpsBootstrapError(RuntimeError):
    """Bootstrap failure with a fixed message; never carries Secret content."""


def _create_core_api_client():
    """Build the Kubernetes CoreV1Api client."""
    return K8s.core()


def _normalize_secret(secret) -> StoredSecret:
    """Normalize a V1Secret to bytes and a resourceVersion so callers ignore the model type."""
    raw_data = getattr(secret, "data", None) or {}
    try:
        data = {key: base64.b64decode(value, validate=True) for key, value in raw_data.items()}
    except (binascii.Error, TypeError, ValueError):
        raise HttpsBootstrapError("Secret value is not valid base64") from None
    metadata = getattr(secret, "metadata", None)
    resource_version = (getattr(metadata, "resource_version", "") or "") if metadata is not None else ""
    return StoredSecret(data=data, resource_version=resource_version)


def read_secret(api, name: str, namespace: str) -> Optional[StoredSecret]:
    """Read a Secret; None when it is absent, every other error becomes a fixed message."""
    try:
        secret = api.read_namespaced_secret(name, namespace, _request_timeout=KUBERNETES_API_TIMEOUT_SECONDS)
    except Exception as error:
        if getattr(error, "status", None) == 404:
            return None
        raise HttpsBootstrapError("Kubernetes Secret read failed") from None
    if secret is None:
        raise HttpsBootstrapError("Kubernetes Secret read returned no object")
    return _normalize_secret(secret)


def create_secret_if_absent(api, namespace: str, body: dict) -> StoredSecret:
    """Create a Secret; on a concurrent create (409) the existing object wins."""
    try:
        created = api.create_namespaced_secret(namespace, body, _request_timeout=KUBERNETES_API_TIMEOUT_SECONDS)
    except Exception as error:
        if getattr(error, "status", None) == 409:
            existing = read_secret(api, body["metadata"]["name"], namespace)
            if existing is not None:
                return existing
            raise HttpsBootstrapError("Secret conflict could not be reread") from None
        raise HttpsBootstrapError("Kubernetes Secret create failed") from None
    if created is None:
        raise HttpsBootstrapError("Kubernetes Secret create returned no object")
    return _normalize_secret(created)


def _patch_secret(api, namespace: str, body: dict) -> StoredSecret:
    """Patch the TLS Secret with its resourceVersion; the API rejects stale versions."""
    try:
        patched = api.patch_namespaced_secret(
            TLS_SECRET_NAME, namespace, body, _request_timeout=KUBERNETES_API_TIMEOUT_SECONDS
        )
    except Exception:
        raise HttpsBootstrapError("Kubernetes TLS Secret renewal failed") from None
    if patched is None:
        raise HttpsBootstrapError("Kubernetes TLS Secret renewal returned no object")
    return _normalize_secret(patched)


def read_extra_sans_from_env() -> tuple[set, set]:
    """Read WEBUI_TLS_SANS and HOST_IP as (extra IPs, extra DNS names).

    Invalid entries are skipped with a warning instead of blocking startup.
    """
    ip_addresses, dns_names = set(), set()
    accepted = 0
    for entry in (os.environ.get(TLS_SANS_ENV) or "").split(","):
        value = entry.strip()
        if not value:
            continue
        accepted += 1
        if accepted > MAX_TLS_SANS:
            logger.warning("Ignoring entries beyond %d in %s", MAX_TLS_SANS, TLS_SANS_ENV)
            break
        parsed = parse_san_value(value)
        if parsed is None:
            logger.warning("Ignoring an invalid %s entry", TLS_SANS_ENV)
            continue
        ip_addresses |= parsed.ip_addresses
        dns_names |= parsed.dns_names
    host_ip = (os.environ.get(HOST_IP_ENV) or "").strip()
    if host_ip:
        try:
            ip_addresses.add(ipaddress.ip_address(host_ip))
        except ValueError:
            logger.warning("Ignoring an invalid %s value", HOST_IP_ENV)
    return ip_addresses, dns_names


def list_node_internal_ips(api) -> set:
    """Enumerate node InternalIPs, which drive renewal on node growth.

    An invalid single address is skipped with a warning: one bad node must not disable
    the whole WebUI.
    """
    try:
        result = api.list_node(_request_timeout=KUBERNETES_API_TIMEOUT_SECONDS)
    except Exception:
        raise HttpsBootstrapError("Kubernetes node InternalIP enumeration failed") from None
    ip_addresses = set()
    for node in getattr(result, "items", None) or []:
        status = getattr(node, "status", None)
        for address in getattr(status, "addresses", None) or []:
            if getattr(address, "type", None) != "InternalIP":
                continue
            try:
                ip_addresses.add(ipaddress.ip_address(getattr(address, "address", "")))
            except (TypeError, ValueError):
                logger.warning("Ignoring an invalid node InternalIP")
    return ip_addresses


def collect_cluster_required_sans(api, namespace: str) -> RequiredSanSet:
    """Required SANs of the cluster: node InternalIPs, extra entries and Service DNS names."""
    extra_ips, extra_dns = read_extra_sans_from_env()
    return build_required_sans(namespace, list_node_internal_ips(api) | extra_ips, extra_dns)


def _secret_body(name: str, namespace: str, secret_type: str, data: dict) -> dict:
    return {
        "apiVersion": "v1",
        "kind": "Secret",
        "type": secret_type,
        "metadata": {"name": name, "namespace": namespace},
        "data": data,
    }


def _build_tls_secret_body(namespace: str, required: RequiredSanSet) -> dict:
    """Build a new TLS Secret body holding the self-signed certificate and key."""
    material = generate_tls_certificate(build_certificate_common_name(namespace), required, datetime.now(timezone.utc))
    return _secret_body(
        TLS_SECRET_NAME,
        namespace,
        "kubernetes.io/tls",
        {"tls.crt": _encode_secret_value(material.cert_pem), "tls.key": _encode_secret_value(material.key_pem)},
    )


def _encode_secret_value(value: bytes) -> str:
    return base64.b64encode(value).decode("ascii")


def _require_secret_value(secret: StoredSecret, key: str) -> bytes:
    """Return a required Secret field; field names are fixed and safe to log."""
    try:
        return secret.data[key]
    except KeyError:
        raise HttpsBootstrapError(f"Secret is missing the required field {key}") from None


def _load_keypair_from_secret(secret: StoredSecret):
    """Decode and validate the certificate and key from the Secret."""
    return parse_tls_certificate(_require_secret_value(secret, "tls.crt"), _require_secret_value(secret, "tls.key"))


def _renew_tls_secret(api, namespace: str, secret: StoredSecret, required: RequiredSanSet) -> StoredSecret:
    """Renew and patch the TLS Secret with its resourceVersion to avoid clobbering updates."""
    if not secret.resource_version:
        raise HttpsBootstrapError("TLS Secret has no resourceVersion for safe renewal")
    body = _build_tls_secret_body(namespace, required)
    body["metadata"]["resourceVersion"] = secret.resource_version
    return _patch_secret(api, namespace, body)


def load_or_renew_tls_keypair(api, namespace: str) -> Optional[TlsKeypair]:
    """Return a usable key pair; create it when missing and renew it when stale.

    None means the material is unusable (a not-yet-valid certificate cannot be renewed).
    Renewal logs distinguish a validity problem from missing SANs. A renewed certificate
    that still needs renewal is treated as a bootstrap failure.
    """
    required = collect_cluster_required_sans(api, namespace)
    secret = read_secret(api, TLS_SECRET_NAME, namespace)
    if secret is None:
        secret = create_secret_if_absent(api, namespace, _build_tls_secret_body(namespace, required))
    material = _load_keypair_from_secret(secret)

    renewal_reason = check_certificate_renewal(material.certificate, required, datetime.now(timezone.utc))
    if renewal_reason is CertificateRenewalReason.NOT_YET_VALID:
        logger.warning("WebUI TLS certificate is not yet valid; HTTPS is disabled")
        return None
    if renewal_reason is None:
        return material
    if renewal_reason in (CertificateRenewalReason.EXPIRED, CertificateRenewalReason.NEAR_EXPIRY):
        logger.warning("WebUI TLS certificate is expired or near expiry; renewing")
    else:
        logger.warning("WebUI TLS certificate SAN is incomplete; renewing")

    secret = _renew_tls_secret(api, namespace, secret, required)
    renewed = _load_keypair_from_secret(secret)
    if check_certificate_renewal(renewed.certificate, required, datetime.now(timezone.utc)) is not None:
        raise HttpsBootstrapError("Renewed TLS certificate failed validation")
    logger.info("Renewed WebUI TLS certificate")
    return renewed


def _generate_token_value() -> str:
    """256-bit random token, grouped in four hex digits for manual entry."""
    raw = secrets.token_hex(TOKEN_BYTES)
    return "-".join(raw[index : index + TOKEN_GROUP_SIZE] for index in range(0, len(raw), TOKEN_GROUP_SIZE))


def _build_token_secret_body(namespace: str, value: str) -> dict:
    return _secret_body(TOKEN_SECRET_NAME, namespace, "Opaque", {"token": _encode_secret_value(value.encode())})


def load_or_create_token(api, namespace: str) -> Optional[str]:
    """Return the shared token, generating it only when the Secret is missing.

    The stored value is used verbatim (no trimming or case folding). An empty or
    non-UTF-8 token means "unavailable" and returns None instead of raising.
    """
    secret = read_secret(api, TOKEN_SECRET_NAME, namespace)
    if secret is None:
        secret = create_secret_if_absent(api, namespace, _build_token_secret_body(namespace, _generate_token_value()))
    try:
        value = _require_secret_value(secret, "token").decode("utf-8")
    except UnicodeDecodeError:
        logger.warning("WebUI token is not valid UTF-8; HTTPS is disabled")
        return None
    if value == "":
        logger.warning("WebUI token is empty; HTTPS is disabled")
        return None
    return value


def _write_private_file(path: Path, data: bytes) -> None:
    """Create the file as 0600 without following symlinks."""
    flags = os.O_WRONLY | os.O_CREAT | os.O_TRUNC | getattr(os, "O_NOFOLLOW", 0)
    descriptor = os.open(path, flags, PRIVATE_FILE_MODE)
    with os.fdopen(descriptor, "wb") as handle:
        os.fchmod(handle.fileno(), PRIVATE_FILE_MODE)
        handle.write(data)


def _prepare_tls_directory() -> Optional[Path]:
    """Prepare the private directory and check its owner before writing the key."""
    target = TLS_MATERIAL_DIRECTORY
    if target.is_symlink():
        logger.warning("WebUI HTTPS bootstrap disabled: TLS directory must not be a symlink")
        return None
    try:
        target.mkdir(parents=True, exist_ok=True, mode=PRIVATE_DIRECTORY_MODE)
        st_uid = getattr(target.stat(), "st_uid", None)
        if st_uid is not None and hasattr(os, "getuid") and st_uid != os.getuid():
            logger.warning("WebUI HTTPS bootstrap disabled: TLS directory is not owned by this process")
            return None
        os.chmod(target, PRIVATE_DIRECTORY_MODE)
    except OSError as error:
        logger.warning("WebUI HTTPS bootstrap disabled by %s", type(error).__name__)
        return None
    return target


def _run_bootstrap_step(step: str, loader, api, namespace: str):
    """Run one bootstrap step; a failure only logs, so the other material still proceeds."""
    try:
        return loader(api, namespace)
    except (HttpsBootstrapError, TlsMaterialError) as error:
        logger.warning("WebUI %s bootstrap failed: %s", step, error)
    except Exception as error:
        logger.warning("WebUI %s bootstrap failed by %s", step, type(error).__name__)
    return None


def bootstrap_https_material(namespace: str) -> Optional[HttpsMaterial]:
    """Prepare the HTTPS material; any failure returns None and only HTTPS is disabled."""
    directory = _prepare_tls_directory()
    if directory is None:
        return None
    try:
        api = _create_core_api_client()
        keypair = _run_bootstrap_step("TLS", load_or_renew_tls_keypair, api, namespace)
        token = _run_bootstrap_step("token", load_or_create_token, api, namespace)
        if keypair is None or token is None:
            return None
        certfile, keyfile = directory / "tls.crt", directory / "tls.key"
        _write_private_file(certfile, keypair.cert_pem)
        _write_private_file(keyfile, keypair.key_pem)
        return HttpsMaterial(certfile=certfile, keyfile=keyfile, token=token)
    except Exception as error:
        logger.warning("WebUI HTTPS bootstrap disabled by %s", type(error).__name__)
    return None
