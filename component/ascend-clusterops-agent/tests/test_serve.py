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

"""Startup orchestration: HTTP first, one lifespan run, HTTPS failures only degrade.

The uvicorn servers are created by uvicorn_lifecycle.create_uvicorn_server, so replacing
that function verifies the ordering without opening real sockets.
"""
# pylint: disable=duplicate-code  # shared TLS fixtures across listener and bootstrap tests

# pylint: disable=redefined-outer-name  # pytest fixtures share names with test parameters

import asyncio
from contextlib import asynccontextmanager
from datetime import datetime, timedelta, timezone
from types import SimpleNamespace

import pytest
from cryptography import x509
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.x509.oid import NameOID
from fastapi import FastAPI

# agent_core.serve 的导入链经过 agent_core.main → api.stream_api → chat.graph → langchain；
# 流水线旧镜像没有 langchain 时整文件跳过，避免收集阶段中断整个 UT 任务。
pytest.importorskip("langchain")

from agent_core import uvicorn_lifecycle
from agent_core.serve import ClusterOpsServer, ListenOptions, is_valid_namespace, validate_listen_ports


@pytest.fixture
def material(tmp_path):
    """HTTPS material needs a real PEM: uvicorn loads the certificate while building the config."""
    key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
    name = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, "localhost")])
    now = datetime.now(timezone.utc)
    certificate = (
        x509.CertificateBuilder()
        .subject_name(name)
        .issuer_name(name)
        .public_key(key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(now - timedelta(minutes=1))
        .not_valid_after(now + timedelta(days=1))
        .sign(key, hashes.SHA256())
    )
    certfile = tmp_path / "cert.pem"
    keyfile = tmp_path / "key.pem"
    certfile.write_bytes(certificate.public_bytes(serialization.Encoding.PEM))
    keyfile.write_bytes(
        key.private_bytes(serialization.Encoding.PEM, serialization.PrivateFormat.PKCS8, serialization.NoEncryption())
    )
    return SimpleNamespace(certfile=certfile, keyfile=keyfile, token="test-token")


class FakeServer:
    """Minimal uvicorn stand-in that records the started flag."""

    def __init__(self, config, fail_secure=False):
        self.config = config
        self.started = False
        self.should_exit = False
        self.fail_secure = fail_secure

    async def _wait(self):
        self.started = True
        while not self.should_exit:
            await asyncio.sleep(0.001)

    async def serve(self):
        if self.config.lifespan == "on":
            async with self.config.app.router.lifespan_context(self.config.app):
                await self._wait()
        elif self.fail_secure:
            self.started = True
            raise SystemExit(1)
        else:
            await self._wait()


async def _until(predicate):
    async def wait():
        while not predicate():
            await asyncio.sleep(0.001)

    await asyncio.wait_for(wait(), timeout=2)


def _fixture(fail_secure=False):
    counts = {"start": 0, "stop": 0}

    @asynccontextmanager
    async def lifespan(_app):
        counts["start"] += 1
        yield
        counts["stop"] += 1

    app = FastAPI(lifespan=lifespan)
    servers = []

    def factory(config):
        server = FakeServer(config, fail_secure=fail_secure)
        servers.append(server)
        return server

    return app, servers, counts, factory


def install_fakes(monkeypatch, fail_secure=False):
    """Replace listener construction and HTTPS bootstrap; return (app, servers, counts)."""
    app, servers, counts, factory = _fixture(fail_secure)
    monkeypatch.setattr(uvicorn_lifecycle, "create_uvicorn_server", factory)
    return app, servers, counts


def test_dual_listener_bootstraps_after_http_and_runs_lifespan_once(monkeypatch, material):
    app, servers, counts = install_fakes(monkeypatch)
    attempts = []

    def bootstrap(namespace):
        assert namespace == "test-ns"
        assert servers[0].started and counts["start"] == 1
        attempts.append(True)
        return material

    monkeypatch.setattr(uvicorn_lifecycle, "bootstrap_https_material", bootstrap)

    async def scenario():
        server = ClusterOpsServer(app, ListenOptions(namespace="test-ns"))
        runner = asyncio.create_task(server.run())
        await _until(lambda: len(servers) == 2 and servers[1].started)
        assert servers[0].config.app is servers[1].config.app is app
        assert servers[0].config.lifespan == "on"
        assert servers[1].config.lifespan == "off"
        assert not servers[0].config.proxy_headers and not servers[1].config.proxy_headers
        assert app.state.webui_token == material.token
        assert counts == {"start": 1, "stop": 0}
        await server.stop()
        await asyncio.wait_for(runner, timeout=2)
        assert counts == {"start": 1, "stop": 1}

    asyncio.run(scenario())
    assert attempts


@pytest.mark.parametrize("failure", ["none", "exception", "empty", "bad-certificate"])
def test_bootstrap_failure_preserves_http_and_single_lifespan(monkeypatch, material, failure):
    app, servers, counts = install_fakes(monkeypatch)
    attempted = []

    def bootstrap(_namespace):
        attempted.append(True)
        if failure == "exception":
            raise RuntimeError("private bootstrap failure")
        if failure == "empty":
            return SimpleNamespace(certfile=material.certfile, keyfile=material.keyfile, token="")
        if failure == "bad-certificate":
            return SimpleNamespace(certfile="/missing/cert.pem", keyfile="/missing/key.pem", token="test-token")
        return None

    monkeypatch.setattr(uvicorn_lifecycle, "bootstrap_https_material", bootstrap)

    async def scenario():
        server = ClusterOpsServer(app, ListenOptions(namespace="test-ns"))
        runner = asyncio.create_task(server.run())
        await _until(lambda: bool(attempted))
        await asyncio.sleep(0.03)
        assert servers[0].started and not runner.done()
        assert len(servers) == 1
        assert app.state.webui_token == ""
        await server.stop()
        await asyncio.wait_for(runner, timeout=2)
        assert counts == {"start": 1, "stop": 1}

    asyncio.run(scenario())


def test_https_bind_failure_never_stops_http(monkeypatch, material):
    app, servers, counts = install_fakes(monkeypatch, fail_secure=True)
    monkeypatch.setattr(uvicorn_lifecycle, "bootstrap_https_material", lambda _namespace: material)

    async def scenario():
        server = ClusterOpsServer(app, ListenOptions(namespace="test-ns"))
        runner = asyncio.create_task(server.run())
        await _until(lambda: len(servers) == 2 and servers[1].started)
        await asyncio.sleep(0.02)
        assert servers[0].started and not runner.done()
        await server.stop()
        await asyncio.wait_for(runner, timeout=2)
        assert counts == {"start": 1, "stop": 1}

    asyncio.run(scenario())


@pytest.mark.parametrize(
    "http_port,https_port",
    [(0, 9743), (1023, 9743), (70000, 9743), (9700, 9700), ("x", 9743)],
)
def test_invalid_or_conflicting_ports_are_rejected(http_port, https_port):
    """Ports below 1024 need root privileges, so they are rejected."""
    with pytest.raises(ValueError):
        validate_listen_ports(http_port, https_port)


def test_port_range_boundaries_are_accepted():
    assert validate_listen_ports("1024", "65535") == (1024, 65535)


def test_valid_ports_are_normalized():
    assert validate_listen_ports("9700", "9743") == (9700, 9743)


def test_namespace_validation():
    assert is_valid_namespace("mindx-dl")
    assert is_valid_namespace("test-ns")
    assert not is_valid_namespace("")
    assert not is_valid_namespace("Mindx_DL")
    assert not is_valid_namespace("-bad")
    assert not is_valid_namespace("a" * 64)


def test_options_keep_an_explicit_namespace():
    assert ListenOptions(namespace="test-ns").resolved().namespace == "test-ns"
