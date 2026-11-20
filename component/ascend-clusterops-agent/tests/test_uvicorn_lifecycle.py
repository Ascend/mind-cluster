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

"""uvicorn listener lifecycle: config building, single start, socket cleanup after failures.

These cases cover the listener layer only; the HTTP/HTTPS ordering lives in test_serve.py.
"""
# pylint: disable=duplicate-code  # shared TLS fixtures across listener and bootstrap tests

# pylint: disable=redefined-outer-name  # pytest fixtures share names with test parameters

import asyncio
import ssl
from contextlib import asynccontextmanager
from datetime import datetime, timedelta, timezone
from types import SimpleNamespace

import pytest
from cryptography import x509
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.x509.oid import NameOID
from fastapi import FastAPI

from agent_core import uvicorn_lifecycle


@pytest.fixture
def material(tmp_path):
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


async def _until(predicate):
    async def wait():
        while not predicate():
            await asyncio.sleep(0.001)

    await asyncio.wait_for(wait(), timeout=2)


def test_http_config_enables_lifespan_and_ignores_proxy_headers():
    config = uvicorn_lifecycle.build_http_config(FastAPI(), "0.0.0.0", 9700)
    assert config.lifespan == "on"
    assert config.proxy_headers is False


def test_https_config_uses_tls_12_and_disables_lifespan(material):
    config = uvicorn_lifecycle.build_https_config(FastAPI(), "0.0.0.0", 9743, material)
    assert config.lifespan == "off"
    assert config.proxy_headers is False
    assert config.ssl.minimum_version == ssl.TLSVersion.TLSv1_2


def test_managed_server_does_not_steal_signals():
    """The coordinator owns process signals, so the child listener capture must be a no-op."""
    server = uvicorn_lifecycle.ManagedUvicornServer(uvicorn_lifecycle.build_http_config(FastAPI(), "127.0.0.1", 0))
    assert server.install_signal_handlers() is None


@pytest.mark.parametrize("failure", ["exception", "cancel"])
def test_bound_https_failure_closes_socket_and_keeps_http_serving(material, failure):
    """Failure after the TLS bind: the socket must close and the HTTP listener keeps serving."""
    import socket

    import httpx

    async def scenario():
        counts = {"start": 0, "stop": 0}

        @asynccontextmanager
        async def lifespan(_app):
            counts["start"] += 1
            yield
            counts["stop"] += 1

        app = FastAPI(lifespan=lifespan)

        @app.get("/health")
        async def health():
            return {"ok": True}

        ready = asyncio.Event()
        release = asyncio.Event()

        class FailedAfterBind(uvicorn_lifecycle.ManagedUvicornServer):
            async def main_loop(self):
                ready.set()
                await release.wait()
                raise RuntimeError("deliberate post-bind failure")

        http = uvicorn_lifecycle.ManagedUvicornServer(uvicorn_lifecycle.build_http_config(app, "127.0.0.1", 0))
        secure = FailedAfterBind(uvicorn_lifecycle.build_https_config(app, "127.0.0.1", 0, material))
        http_task = asyncio.create_task(uvicorn_lifecycle.run_server_until_stopped(http))
        secure_task = None
        try:
            await _until(lambda: http.started)
            http_port = http.servers[0].sockets[0].getsockname()[1]
            secure_task = asyncio.create_task(uvicorn_lifecycle.run_server_until_stopped(secure))
            await asyncio.wait_for(ready.wait(), timeout=2)
            secure_port = secure.servers[0].sockets[0].getsockname()[1]
            async with httpx.AsyncClient(verify=False, trust_env=False) as client:
                before = await client.get(f"https://127.0.0.1:{secure_port}/health", headers={"Connection": "close"})
                assert before.status_code == 200
                if failure == "cancel":
                    secure_task.cancel()
                    with pytest.raises(asyncio.CancelledError):
                        await asyncio.wait_for(secure_task, timeout=2)
                else:
                    release.set()
                    with pytest.raises(RuntimeError):
                        await asyncio.wait_for(secure_task, timeout=2)
                with socket.socket() as probe:
                    probe.settimeout(1)
                    assert probe.connect_ex(("127.0.0.1", secure_port)) != 0
                assert all(not listener.is_serving() for listener in secure.servers)
                after = await client.get(f"http://127.0.0.1:{http_port}/health")
                assert after.status_code == 200
                assert not http_task.done()
                assert counts == {"start": 1, "stop": 0}
        finally:
            if secure_task is not None and not secure_task.done():
                secure_task.cancel()
                await asyncio.gather(secure_task, return_exceptions=True)
            http.should_exit = True
            await asyncio.wait_for(http_task, timeout=3)
        assert counts == {"start": 1, "stop": 1}

    asyncio.run(scenario())


def test_stop_during_http_startup_closes_socket_and_runs_lifespan_once():
    """uvicorn skips its own cleanup when asked to exit during startup; it must be done here."""
    counts = {"start": 0, "stop": 0}
    server = None

    @asynccontextmanager
    async def lifespan(_app):
        counts["start"] += 1
        server.should_exit = True
        yield
        counts["stop"] += 1

    async def scenario():
        nonlocal server
        app = FastAPI(lifespan=lifespan)
        server = uvicorn_lifecycle.ManagedUvicornServer(uvicorn_lifecycle.build_http_config(app, "127.0.0.1", 0))
        await asyncio.wait_for(uvicorn_lifecycle.run_server_until_stopped(server), timeout=3)

    asyncio.run(scenario())
    assert counts == {"start": 1, "stop": 1}
    assert server.servers and all(not listener.is_serving() for listener in server.servers)
