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

"""Secret bootstrap: creation, reuse, renewal, concurrent creation and degradation.

Kubernetes is replaced by a lightweight fake API returning real V1Secret models, so the
production code needs no dict/object dual entry points for tests.
"""
# pylint: disable=duplicate-code  # shared TLS fixtures across listener and bootstrap tests

# pylint: disable=redefined-outer-name  # pytest fixtures share names with test parameters

import base64
import copy
import ipaddress
import logging
import os
import re
import stat
from datetime import datetime, timedelta, timezone
from pathlib import Path
from types import SimpleNamespace

import pytest
from cryptography import x509
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import ec, rsa
from cryptography.x509.oid import ExtendedKeyUsageOID, NameOID
from kubernetes.client import V1ObjectMeta, V1Secret

from agent_core.https import secret_bootstrap
from agent_core.https import tls

NS = "mindx-dl"
IPS = {"10.0.0.1", "10.0.0.2", "fd00::1"}


class ApiFailure(Exception):
    def __init__(self, status):
        self.status = status
        super().__init__("sensitive API response must not be logged")


def secret_object(body: dict) -> V1Secret:
    """Turn a request body into a real V1Secret so the fake API matches what production reads."""
    metadata = body.get("metadata") or {}
    return V1Secret(
        api_version=body.get("apiVersion"),
        kind=body.get("kind"),
        type=body.get("type"),
        metadata=V1ObjectMeta(
            name=metadata.get("name"),
            namespace=metadata.get("namespace"),
            resource_version=metadata.get("resourceVersion"),
            annotations=metadata.get("annotations"),
        ),
        data=dict(body.get("data") or {}),
    )


class FakeApi:
    def __init__(self, values=None):
        self.values = copy.deepcopy(values or {})
        self.calls = []
        self.conflicts = {}
        self.failures = {}
        self.nodes = None

    def _call(self, method, name, kwargs):
        assert kwargs.get("_request_timeout") == secret_bootstrap.KUBERNETES_API_TIMEOUT_SECONDS
        self.calls.append((method, name))
        if (method, name) in self.failures:
            raise ApiFailure(self.failures[(method, name)])

    def list_node(self, **kwargs):
        self._call("list", "nodes", kwargs)
        if self.nodes is not None:
            return SimpleNamespace(items=self.nodes)
        return SimpleNamespace(
            items=[
                SimpleNamespace(
                    status=SimpleNamespace(
                        addresses=[
                            SimpleNamespace(type="InternalIP", address="10.0.0.1"),
                            SimpleNamespace(type="InternalIP", address="fd00::1"),
                            SimpleNamespace(type="ExternalIP", address="203.0.113.20"),
                        ]
                    )
                ),
                SimpleNamespace(
                    status=SimpleNamespace(
                        addresses=[SimpleNamespace(type="InternalIP", address="10.0.0.2")],
                    )
                ),
            ]
        )

    def read_namespaced_secret(self, name, namespace, **kwargs):
        assert namespace == NS
        self._call("get", name, kwargs)
        if name not in self.values:
            raise ApiFailure(404)
        return copy.deepcopy(self.values[name])

    def create_namespaced_secret(self, namespace, body, **kwargs):
        assert namespace == NS
        name = body["metadata"]["name"]
        self._call("create", name, kwargs)
        if name in self.conflicts:
            self.values[name] = copy.deepcopy(self.conflicts.pop(name))
            raise ApiFailure(409)
        if name in self.values:
            raise ApiFailure(409)
        body = copy.deepcopy(body)
        body["metadata"]["resourceVersion"] = "1"
        self.values[name] = secret_object(body)
        return copy.deepcopy(self.values[name])

    def patch_namespaced_secret(self, name, namespace, body, **kwargs):
        assert namespace == NS
        assert name == secret_bootstrap.TLS_SECRET_NAME, "token must never be updated"
        self._call("patch", name, kwargs)
        old = self.values[name]
        assert body["metadata"]["resourceVersion"] == old.metadata.resource_version
        body = copy.deepcopy(body)
        body["metadata"]["annotations"] = dict(
            old.metadata.annotations or {}, **(body["metadata"].get("annotations") or {})
        )
        body["metadata"]["resourceVersion"] = str(int(old.metadata.resource_version) + 1)
        self.values[name] = secret_object(body)
        return copy.deepcopy(self.values[name])


@pytest.fixture(scope="module")
def key():
    return rsa.generate_private_key(public_exponent=65537, key_size=4096)


@pytest.fixture(autouse=True)
def shared_rsa_key(monkeypatch, key):
    """Sign with real RSA-4096 but share one key, so per-test generation is not repeated."""
    monkeypatch.setattr(tls.rsa, "generate_private_key", lambda **_kwargs: key)


@pytest.fixture(autouse=True)
def isolated_san_environment(monkeypatch):
    """Dev machines and CI may define these variables; the default SAN assertions isolate them."""
    monkeypatch.delenv(secret_bootstrap.TLS_SANS_ENV, raising=False)
    monkeypatch.delenv(secret_bootstrap.HOST_IP_ENV, raising=False)


def token_secret(value: str) -> V1Secret:
    return V1Secret(
        metadata=V1ObjectMeta(name=secret_bootstrap.TOKEN_SECRET_NAME, resource_version="1"),
        data={"token": base64.b64encode(value.encode("utf-8")).decode("ascii")},
    )


def certificate_secret(private_key, *, days=365, complete_san=True, eku=True, future=False):
    now = datetime.now(timezone.utc)
    subject = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, "test-webui")])
    builder = (
        x509.CertificateBuilder()
        .subject_name(subject)
        .issuer_name(subject)
        .public_key(private_key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(now + timedelta(days=1) if future else now - timedelta(days=366))
        .not_valid_after(now + timedelta(days=days))
    )
    names = [x509.DNSName("custom.example")]
    if complete_san:
        names += [x509.IPAddress(ipaddress.ip_address(ip)) for ip in IPS]
        names += [x509.DNSName(name) for name in tls.get_service_dns_names(NS)]
    builder = builder.add_extension(x509.SubjectAlternativeName(names), critical=False)
    if eku:
        builder = builder.add_extension(x509.ExtendedKeyUsage([ExtendedKeyUsageOID.SERVER_AUTH]), critical=False)
    certificate = builder.sign(private_key, hashes.SHA256())
    return V1Secret(
        metadata=V1ObjectMeta(
            name=secret_bootstrap.TLS_SECRET_NAME,
            resource_version="7",
            annotations={"existing-note": "preserve"},
        ),
        type="kubernetes.io/tls",
        data={
            "tls.crt": base64.b64encode(certificate.public_bytes(serialization.Encoding.PEM)).decode("ascii"),
            "tls.key": base64.b64encode(
                private_key.private_bytes(
                    serialization.Encoding.PEM, serialization.PrivateFormat.PKCS8, serialization.NoEncryption()
                )
            ).decode("ascii"),
        },
    )


def configured_api(private_key, **kwargs):
    return FakeApi(
        {
            secret_bootstrap.TLS_SECRET_NAME: certificate_secret(private_key, **kwargs),
            secret_bootstrap.TOKEN_SECRET_NAME: token_secret("shared-token"),
        }
    )


def run_bootstrap(monkeypatch, api, directory):
    """Production takes only a namespace; tests replace the module collaborators and the directory."""
    monkeypatch.setattr(secret_bootstrap, "_create_core_api_client", lambda: api)
    monkeypatch.setattr(secret_bootstrap, "TLS_MATERIAL_DIRECTORY", Path(directory))
    return secret_bootstrap.bootstrap_https_material(NS)


def changes(api):
    return [call for call in api.calls if call[0] in ("create", "patch")]


def test_bootstrap_creation_real_crypto_sans_eku_entropy_and_idempotence(tmp_path, key, monkeypatch):
    api = FakeApi()
    real_hex = secret_bootstrap.secrets.token_hex
    requested = []

    def random_value(size):
        requested.append(size)
        return real_hex(size)

    monkeypatch.setattr(secret_bootstrap.secrets, "token_hex", random_value)
    result = run_bootstrap(monkeypatch, api, tmp_path / "first")
    assert result is not None
    assert requested == [secret_bootstrap.TOKEN_BYTES]
    assert re.fullmatch(r"[0-9a-f]{4}(?:-[0-9a-f]{4}){15}", result.token)
    certificate = x509.load_pem_x509_certificate((tmp_path / "first" / "tls.crt").read_bytes())
    assert certificate.public_key().key_size == 4096
    assert certificate.signature_hash_algorithm.name == "sha256"
    assert certificate.subject == certificate.issuer
    assert certificate.extensions.get_extension_for_class(x509.ExtendedKeyUsage).value == x509.ExtendedKeyUsage(
        [ExtendedKeyUsageOID.SERVER_AUTH]
    )
    san = certificate.extensions.get_extension_for_class(x509.SubjectAlternativeName).value
    assert set(san.get_values_for_type(x509.IPAddress)) == {ipaddress.ip_address(ip) for ip in IPS}
    assert set(san.get_values_for_type(x509.DNSName)) == tls.get_service_dns_names(NS)
    assert "203.0.113.20" not in [str(ip) for ip in san.get_values_for_type(x509.IPAddress)]
    assert stat.S_IMODE((tmp_path / "first" / "tls.key").stat().st_mode) == 0o600
    previous = copy.deepcopy(api.values)
    api.calls.clear()
    again = run_bootstrap(monkeypatch, api, tmp_path / "second")
    assert again.token == result.token
    assert api.values == previous
    assert changes(api) == []
    assert result.token not in repr(result)


def test_partial_tls_failure_still_creates_token(tmp_path, key, monkeypatch):
    tls = certificate_secret(key)
    tls.data["tls.crt"] = base64.b64encode(b"not a certificate").decode("ascii")
    api = FakeApi({secret_bootstrap.TLS_SECRET_NAME: tls})
    assert run_bootstrap(monkeypatch, api, tmp_path) is None
    assert secret_bootstrap.TOKEN_SECRET_NAME in api.values
    assert ("create", secret_bootstrap.TOKEN_SECRET_NAME) in api.calls


def test_partial_token_failure_still_creates_tls(tmp_path, key, monkeypatch):
    api = FakeApi({secret_bootstrap.TOKEN_SECRET_NAME: token_secret("")})
    assert run_bootstrap(monkeypatch, api, tmp_path) is None
    assert secret_bootstrap.TLS_SECRET_NAME in api.values
    assert ("create", secret_bootstrap.TLS_SECRET_NAME) in api.calls


def test_default_tls_directory_is_fixed_and_private(tmp_path, monkeypatch):
    target = tmp_path / "webui-tls"
    result = run_bootstrap(monkeypatch, FakeApi(), target)
    assert result is not None
    assert target.is_dir()
    assert result.certfile == target / "tls.crt"
    assert result.keyfile == target / "tls.key"
    assert stat.S_IMODE((target / "tls.key").stat().st_mode) == 0o600
    if os.name != "nt":
        assert stat.S_IMODE(target.stat().st_mode) == 0o700


@pytest.mark.skipif(os.name == "nt", reason="symlink semantics differ on Windows")
def test_tls_directory_symlink_is_rejected_before_touching_the_api(tmp_path, monkeypatch):
    real = tmp_path / "real"
    real.mkdir()
    link = tmp_path / "link"
    link.symlink_to(real, target_is_directory=True)
    api = FakeApi()
    assert run_bootstrap(monkeypatch, api, link) is None
    assert not api.calls


def test_tls_sans_entry_count_is_capped(monkeypatch, caplog):
    values = ",".join(f"10.1.0.{index}" for index in range(1, 102))
    monkeypatch.setenv(secret_bootstrap.TLS_SANS_ENV, values)
    with caplog.at_level(logging.WARNING):
        ips, dns = secret_bootstrap.read_extra_sans_from_env()
    assert len(ips) == secret_bootstrap.MAX_TLS_SANS
    assert not dns
    assert "beyond 100" in caplog.text


def test_invalid_host_ip_is_ignored(monkeypatch, caplog):
    monkeypatch.setenv(secret_bootstrap.HOST_IP_ENV, "not-an-ip")
    with caplog.at_level(logging.WARNING):
        ips, dns = secret_bootstrap.read_extra_sans_from_env()
    assert not ips and not dns
    assert secret_bootstrap.HOST_IP_ENV in caplog.text


def test_invalid_node_internal_ip_is_skipped(caplog):
    """One bad node must not fail the whole SAN computation and disable HTTPS."""
    api = FakeApi()
    api.nodes = [
        SimpleNamespace(
            status=SimpleNamespace(
                addresses=[
                    SimpleNamespace(type="InternalIP", address="10.0.0.1"),
                    SimpleNamespace(type="InternalIP", address="not-an-ip"),
                    SimpleNamespace(type="InternalIP", address=None),
                    SimpleNamespace(type="ExternalIP", address="203.0.113.20"),
                ]
            )
        )
    ]
    with caplog.at_level(logging.WARNING):
        assert secret_bootstrap.list_node_internal_ips(api) == {ipaddress.ip_address("10.0.0.1")}
    assert "InternalIP" in caplog.text


def test_extra_sans_and_downward_host_ip_are_included(tmp_path, key, monkeypatch, caplog):
    monkeypatch.setenv(secret_bootstrap.TLS_SANS_ENV, "198.51.100.7, extra.example, bad/name")
    monkeypatch.setenv(secret_bootstrap.HOST_IP_ENV, "10.0.0.99")
    with caplog.at_level(logging.WARNING):
        assert run_bootstrap(monkeypatch, FakeApi(), tmp_path) is not None
    certificate = x509.load_pem_x509_certificate((tmp_path / "tls.crt").read_bytes())
    san = certificate.extensions.get_extension_for_class(x509.SubjectAlternativeName).value
    ips = {str(ip) for ip in san.get_values_for_type(x509.IPAddress)}
    dns = set(san.get_values_for_type(x509.DNSName))
    assert {"198.51.100.7", "10.0.0.99"}.issubset(ips)
    assert "extra.example" in dns
    assert "bad/name" not in dns
    assert "invalid" in caplog.text.lower()


@pytest.mark.parametrize("days,complete_san", [(-1, True), (365, False)])
def test_expiry_or_san_drift_renews_all_required_names(tmp_path, key, monkeypatch, days, complete_san):
    api = configured_api(key, days=days, complete_san=complete_san)
    old = copy.deepcopy(api.values[secret_bootstrap.TLS_SECRET_NAME])
    result = run_bootstrap(monkeypatch, api, tmp_path)
    assert result is not None
    assert changes(api) == [("patch", secret_bootstrap.TLS_SECRET_NAME)]
    assert api.values[secret_bootstrap.TLS_SECRET_NAME].data != old.data
    assert api.values[secret_bootstrap.TLS_SECRET_NAME].metadata.annotations["existing-note"] == "preserve"
    certificate = x509.load_pem_x509_certificate((tmp_path / "tls.crt").read_bytes())
    san = certificate.extensions.get_extension_for_class(x509.SubjectAlternativeName).value
    assert {str(ip) for ip in san.get_values_for_type(x509.IPAddress)} == IPS
    assert set(san.get_values_for_type(x509.DNSName)) == tls.get_service_dns_names(NS)
    assert result.token == "shared-token"


def test_near_expiry_renews(tmp_path, key, monkeypatch, caplog):
    api = configured_api(key, days=20)
    before = copy.deepcopy(api.values)
    with caplog.at_level(logging.INFO):
        assert run_bootstrap(monkeypatch, api, tmp_path) is not None
    assert changes(api) == [("patch", secret_bootstrap.TLS_SECRET_NAME)]
    assert api.values[secret_bootstrap.TLS_SECRET_NAME].data != before[secret_bootstrap.TLS_SECRET_NAME].data
    assert api.values[secret_bootstrap.TLS_SECRET_NAME].metadata.annotations["existing-note"] == "preserve"
    assert "expired or near expiry" in caplog.text
    assert "Renewed WebUI TLS certificate" in caplog.text


@pytest.mark.parametrize("case", ["pem", "mismatch", "eku", "future", "base64"])
def test_invalid_certificate_disables_without_overwrite(tmp_path, key, monkeypatch, case):
    api = configured_api(key, eku=case != "eku", future=case == "future")
    tls = api.values[secret_bootstrap.TLS_SECRET_NAME]
    if case == "pem":
        tls.data["tls.crt"] = base64.b64encode(b"not a certificate").decode()
    elif case == "base64":
        tls.data["tls.crt"] = "%%%invalid%%%"
    elif case == "mismatch":
        wrong = ec.generate_private_key(ec.SECP256R1())
        tls.data["tls.key"] = base64.b64encode(
            wrong.private_bytes(
                serialization.Encoding.PEM, serialization.PrivateFormat.PKCS8, serialization.NoEncryption()
            )
        ).decode()
    before = copy.deepcopy(api.values)
    assert run_bootstrap(monkeypatch, api, tmp_path) is None
    assert api.values == before and changes(api) == []


def test_create_409_reuses_existing_certificate_and_token(tmp_path, key, monkeypatch):
    tls = certificate_secret(key)
    token = token_secret("  Admin-Existing-Token\n")
    api = FakeApi()
    api.conflicts = {secret_bootstrap.TLS_SECRET_NAME: tls, secret_bootstrap.TOKEN_SECRET_NAME: token}
    result = run_bootstrap(monkeypatch, api, tmp_path)
    assert result is not None and result.token == "  Admin-Existing-Token\n"
    assert api.values[secret_bootstrap.TLS_SECRET_NAME] == tls
    assert api.values[secret_bootstrap.TOKEN_SECRET_NAME] == token
    assert not any(call[0] == "patch" for call in api.calls)


@pytest.mark.parametrize("value", ["single", " ABCD-ef12 \n", "  ", "token-without-fixed-format"])
def test_existing_token_is_exact_and_never_updated(tmp_path, key, monkeypatch, value):
    api = configured_api(key)
    api.values[secret_bootstrap.TOKEN_SECRET_NAME] = token_secret(value)
    before = copy.deepcopy(api.values)
    result = run_bootstrap(monkeypatch, api, tmp_path)
    assert result is not None and result.token == value
    assert api.values == before and changes(api) == []


def test_empty_token_warns_and_fails_closed_without_regeneration(tmp_path, key, monkeypatch, caplog):
    api = configured_api(key)
    api.values[secret_bootstrap.TOKEN_SECRET_NAME] = token_secret("")
    before = copy.deepcopy(api.values)
    with caplog.at_level(logging.WARNING):
        assert run_bootstrap(monkeypatch, api, tmp_path) is None
    assert "empty" in caplog.text
    assert api.values == before and changes(api) == []


def test_non_utf8_token_disables_https_without_raising(tmp_path, key, monkeypatch, caplog):
    """Invalid encoding and empty content are the same: the token is unavailable, returned as None."""
    api = configured_api(key)
    secret = token_secret("placeholder")
    secret.data["token"] = base64.b64encode(b"\xff\xfe").decode("ascii")
    api.values[secret_bootstrap.TOKEN_SECRET_NAME] = secret
    before = copy.deepcopy(api.values)
    with caplog.at_level(logging.WARNING):
        assert run_bootstrap(monkeypatch, api, tmp_path) is None
    assert "UTF-8" in caplog.text
    assert api.values == before and changes(api) == []


@pytest.mark.parametrize(
    "method,name",
    [
        ("list", "nodes"),
        ("get", secret_bootstrap.TLS_SECRET_NAME),
        ("get", secret_bootstrap.TOKEN_SECRET_NAME),
        ("patch", secret_bootstrap.TLS_SECRET_NAME),
    ],
)
def test_api_or_rbac_fault_disables_https_without_leaking_response(tmp_path, key, monkeypatch, caplog, method, name):
    api = configured_api(key, days=-1 if method == "patch" else 365)
    api.failures[(method, name)] = 403
    before = copy.deepcopy(api.values)
    with caplog.at_level(logging.WARNING):
        assert run_bootstrap(monkeypatch, api, tmp_path) is None
    assert api.values == before
    assert "sensitive API response" not in caplog.text
    assert "shared-token" not in caplog.text


@pytest.mark.parametrize("name", [secret_bootstrap.TLS_SECRET_NAME, secret_bootstrap.TOKEN_SECRET_NAME])
def test_create_rbac_denied_is_https_only_failure(tmp_path, key, monkeypatch, name):
    if name == secret_bootstrap.TLS_SECRET_NAME:
        api = FakeApi()
    else:
        api = FakeApi({secret_bootstrap.TLS_SECRET_NAME: certificate_secret(key)})
    api.failures[("create", name)] = 403
    assert run_bootstrap(monkeypatch, api, tmp_path) is None


def test_missing_resource_version_never_blindly_overwrites_tls(tmp_path, key, monkeypatch):
    api = configured_api(key, days=-1)
    api.values[secret_bootstrap.TLS_SECRET_NAME].metadata.resource_version = None
    before = copy.deepcopy(api.values)
    assert run_bootstrap(monkeypatch, api, tmp_path) is None
    assert api.values == before and changes(api) == []
