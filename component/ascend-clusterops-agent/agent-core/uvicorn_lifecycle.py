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

"""uvicorn listener lifecycle and the single HTTPS start attempt.

HTTPS never cancels or restarts the HTTP listener: a bootstrap or bind failure only
logs a warning and is retried on the next restart.
"""

from __future__ import annotations

import asyncio
import logging
import ssl
from contextlib import contextmanager
from typing import Optional

import uvicorn

from agent_core.https.secret_bootstrap import HttpsMaterial, bootstrap_https_material

logger = logging.getLogger(__name__)

DEFAULT_HTTP_PORT = 9700
DEFAULT_HTTPS_PORT = 9743
GRACEFUL_SHUTDOWN_SECONDS = 10
BOOTSTRAP_TIMEOUT_SECONDS = 30.0
LISTENER_SHUTDOWN_TIMEOUT_SECONDS = 15
LISTENER_CLOSE_TIMEOUT_SECONDS = 2
MINIMUM_TLS_VERSION = ssl.TLSVersion.TLSv1_2


class ManagedUvicornServer(uvicorn.Server):
    """The coordinator owns process signals; child listeners must not replace the handlers."""

    def __init__(self, config):
        super().__init__(config)
        # uvicorn sets this in its own __init__; declared here so linters see the attribute.
        self.should_exit = False

    @contextmanager
    def capture_signals(self):
        yield

    def install_signal_handlers(self):
        """Compatibility for uvicorn versions without capture_signals."""
        return None


def create_uvicorn_server(config: uvicorn.Config) -> uvicorn.Server:
    """The only place a uvicorn.Server is built; tests replace it to avoid real sockets."""
    return ManagedUvicornServer(config)


def build_http_config(app, host: str, port: int) -> uvicorn.Config:
    """HTTP listener: carries the application lifespan (relcache/pathmap/llm/upload start here).

    Forwarded headers are not trusted; the entry is decided by the real listener scheme.
    """
    return uvicorn.Config(
        app,
        host=host,
        port=port,
        lifespan="on",
        proxy_headers=False,
        timeout_graceful_shutdown=GRACEFUL_SHUTDOWN_SECONDS,
    )


def build_https_config(app, host: str, port: int, material: HttpsMaterial) -> uvicorn.Config:
    """HTTPS listener: reuses the same app without running the lifespan again."""
    config = uvicorn.Config(
        app,
        host=host,
        port=port,
        lifespan="off",
        proxy_headers=False,
        ssl_certfile=str(material.certfile),
        ssl_keyfile=str(material.keyfile),
        ssl_version=ssl.PROTOCOL_TLS_SERVER,
        timeout_graceful_shutdown=GRACEFUL_SHUTDOWN_SECONDS,
    )
    config.load()
    if not isinstance(config.ssl, ssl.SSLContext):
        raise RuntimeError("TLS context was not initialized")
    config.ssl.minimum_version = MINIMUM_TLS_VERSION
    return config


def has_open_listeners(server: uvicorn.Server) -> bool:
    """Whether any socket is still serving."""
    return any(listener.is_serving() for listener in getattr(server, "servers", ()))


async def close_server_listeners(server: uvicorn.Server) -> None:
    """Stop accepting connections first, then bound and isolate the close phase."""
    listeners = tuple(getattr(server, "servers", ()))
    for listener in listeners:
        try:
            listener.close()
        except Exception:
            logger.warning("Listener close failed during shutdown", exc_info=True)
    shutdown = getattr(server, "shutdown", None)
    if listeners and shutdown is not None:
        try:
            await asyncio.wait_for(shutdown(), timeout=LISTENER_SHUTDOWN_TIMEOUT_SECONDS)
        except (Exception, SystemExit):
            logger.warning("Listener cleanup did not complete normally", exc_info=True)
    for listener in listeners:
        try:
            await asyncio.wait_for(listener.wait_closed(), timeout=LISTENER_CLOSE_TIMEOUT_SECONDS)
        except Exception:
            logger.warning("Listener close wait did not complete normally", exc_info=True)


async def run_server_until_stopped(server: uvicorn.Server) -> None:
    """Run one listener until it stops on its own.

    uvicorn exits via sys.exit on startup failures, which is converted to RuntimeError so
    callers only handle ordinary exceptions. A normal stop may skip uvicorn's own cleanup,
    so listeners that still serve are closed here.
    """
    failed = False
    try:
        await server.serve()
    except SystemExit:
        failed = True
        raise RuntimeError("listener exited during startup") from None
    except BaseException:
        failed = True
        raise
    finally:
        if failed or has_open_listeners(server):
            await close_server_listeners(server)


async def _load_https_material(namespace: str) -> Optional[HttpsMaterial]:
    """Run the bootstrap in a thread (Kubernetes calls block); None means HTTPS is unavailable.

    Missing material and an empty token are both expressed as None by
    ``bootstrap_https_material``; only unexpected failures such as a timeout raise.
    """
    material = await asyncio.wait_for(
        asyncio.to_thread(bootstrap_https_material, namespace), timeout=BOOTSTRAP_TIMEOUT_SECONDS
    )
    if material is None or not material.token:
        return None
    return material


async def start_https_listener_once(app, host: str, port: int, namespace: str, stop: asyncio.Event) -> None:
    """Start the HTTPS listener once; failures only log a warning.

    The token is published through app.state and cleared on exit, so no usable credential
    stays behind while HTTPS is unavailable.
    """
    if stop.is_set():
        return
    server = None
    try:
        material = await _load_https_material(namespace)
        if material is None:
            logger.warning("HTTPS unavailable; HTTP CLI remains available")
            return
        server = create_uvicorn_server(build_https_config(app, host, port, material))
        app.state.webui_token = material.token
        await run_server_until_stopped(server)
        if not stop.is_set():
            logger.warning("HTTPS listener stopped unexpectedly; HTTP CLI remains available")
    except (Exception, SystemExit):
        logger.warning("HTTPS unavailable; HTTP CLI remains available", exc_info=True)
    finally:
        if server is not None:
            server.should_exit = True
        app.state.webui_token = ""  # nosec B105
