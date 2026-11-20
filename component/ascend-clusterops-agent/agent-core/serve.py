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

"""Process entry point: runs the HTTP and HTTPS listeners and owns process signals.

HTTP starts first and carries the application lifespan; HTTPS bootstraps in the
background, so a failure there only disables port 9743.
"""

from __future__ import annotations

import argparse
import asyncio
import logging
import os
import re
import signal
import time
from dataclasses import dataclass, replace
from typing import List, Optional

from agent_core import uvicorn_lifecycle
from agent_core.k8s import K8s
from agent_core.uvicorn_lifecycle import DEFAULT_HTTP_PORT, DEFAULT_HTTPS_PORT
from agent_core.main import app

logger = logging.getLogger(__name__)

POD_NAMESPACE_ENV = "POD_NAMESPACE"
HTTP_PORT_ENV = "AGENT_PORT"
HTTPS_PORT_ENV = "WEBUI_HTTPS_PORT"
LISTEN_HOST = "0.0.0.0"
# Ports below 1024 need root privileges and the component runs as hwMindX.
MIN_PORT, MAX_PORT = 1024, 65535
MAX_NAMESPACE_LENGTH = 63
_NAMESPACE_PATTERN = re.compile(r"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")
SHUTDOWN_TIMEOUT_SECONDS = (
    uvicorn_lifecycle.GRACEFUL_SHUTDOWN_SECONDS + uvicorn_lifecycle.LISTENER_CLOSE_TIMEOUT_SECONDS + 1
)


@dataclass(frozen=True)
class ListenOptions:
    """Listener address, ports and namespace (CLI and environment input).

    The address defaults to ``LISTEN_HOST``: the container must listen on every
    interface to be reachable through the Service and NodePort. ``--host`` is only
    needed to pin a single interface.
    """

    http_port: int = DEFAULT_HTTP_PORT
    https_port: int = DEFAULT_HTTPS_PORT
    namespace: Optional[str] = None
    host: str = LISTEN_HOST

    def resolved(self) -> "ListenOptions":
        """Validate the ports and resolve the namespace; a missing namespace only disables HTTPS."""
        http_port, https_port = validate_listen_ports(self.http_port, self.https_port)
        return replace(
            self,
            http_port=http_port,
            https_port=https_port,
            namespace=resolve_namespace(self.namespace),
        )


def validate_listen_ports(http_port, https_port) -> tuple[int, int]:
    """Ports must be integers in 1024..65535 and must differ."""
    try:
        http = int(http_port)
        https = int(https_port)
    except (TypeError, ValueError):
        raise ValueError("http_port and https_port must be integers") from None
    for name, port in (("http_port", http), ("https_port", https)):
        if not MIN_PORT <= port <= MAX_PORT:
            raise ValueError(f"{name} must be between {MIN_PORT} and {MAX_PORT}")
    if http == https:
        raise ValueError("http_port and https_port must differ")
    return http, https


def is_valid_namespace(value: str) -> bool:
    """Check a Kubernetes namespace name (lowercase alphanumerics and hyphens)."""
    return bool(value) and len(value) <= MAX_NAMESPACE_LENGTH and _NAMESPACE_PATTERN.match(value) is not None


def resolve_namespace(declared: Optional[str]) -> Optional[str]:
    """Resolve the namespace: explicit argument > Downward API > service account.

    None means HTTPS is disabled; HTTP keeps serving.
    """
    candidate = declared if declared is not None else (os.environ.get(POD_NAMESPACE_ENV) or "")
    if candidate and not is_valid_namespace(candidate):
        logger.warning("Invalid namespace; falling back to the service account namespace")
        candidate = ""
    if not candidate:
        try:
            candidate = K8s.namespace()
        except Exception:
            logger.warning("Namespace unavailable; HTTPS is disabled and HTTP remains available")
            return None
    if not is_valid_namespace(candidate):
        logger.warning("Invalid namespace; HTTPS is disabled and HTTP remains available")
        return None
    return candidate


def handle_stop_signal(stop: asyncio.Event, signum: int) -> None:
    """Log the signal that requested shutdown, then set the stop event.

    A normal shutdown raises nothing, so this line is the only direct evidence of a
    manual stop: a crashed listener never goes through the handler.
    """
    logger.info("received %s; shutting down", signal.Signals(signum).name)
    stop.set()


def install_signal_handlers(stop: asyncio.Event, loop: asyncio.AbstractEventLoop) -> List[int]:
    """Turn SIGINT/SIGTERM into a stop event.

    An unsupported platform only logs a warning, but never stays silent: a process
    that ignores SIGTERM looks like a hung shutdown in the logs.
    """
    registered = []
    for signum in (signal.SIGINT, signal.SIGTERM):
        try:
            loop.add_signal_handler(signum, handle_stop_signal, stop, signum)
        except (NotImplementedError, RuntimeError, ValueError):
            logger.warning(
                "Signal %s is unavailable; the process cannot be stopped by it",
                signal.Signals(signum).name,
            )
            continue
        registered.append(signum)
    return registered


def remove_signal_handlers(registered: List[int], loop: asyncio.AbstractEventLoop) -> None:
    for signum in registered:
        loop.remove_signal_handler(signum)


class ClusterOpsServer:
    """Lifecycle of both listeners: start, wait and stop are driven by the caller."""

    def __init__(self, web_app, options: ListenOptions):
        self._app = web_app
        self._options = options.resolved()
        self._stop = asyncio.Event()
        self._http_server = None
        self._http_task: Optional[asyncio.Task] = None
        self._https_task: Optional[asyncio.Task] = None
        self._stop_task: Optional[asyncio.Task] = None
        self._registered_signals: List[int] = []

    async def start(self) -> None:
        """Start HTTP and wait until it is ready; HTTPS then bootstraps in the background.

        The token is cleared first so a stale value from a previous process is never reused.
        """
        self._app.state.webui_token = ""  # nosec B105
        self._http_server = uvicorn_lifecycle.create_uvicorn_server(
            uvicorn_lifecycle.build_http_config(self._app, self._options.host, self._options.http_port)
        )
        self._http_task = asyncio.create_task(uvicorn_lifecycle.run_server_until_stopped(self._http_server))
        self._stop_task = asyncio.create_task(self._stop.wait())
        while not self._http_server.started and not self._stop.is_set():
            if self._http_task.done():
                self._http_task.result()
                raise RuntimeError("HTTP listener did not start")
            await asyncio.sleep(0.01)
        if self._stop.is_set():
            return
        self._start_https()

    def _start_https(self) -> None:
        """Start HTTPS once in the background; skip it when the namespace is unknown."""
        if self._options.namespace is None:
            logger.warning("HTTPS disabled: namespace is unavailable")
            return
        self._https_task = asyncio.create_task(
            uvicorn_lifecycle.start_https_listener_once(
                self._app,
                self._options.host,
                self._options.https_port,
                self._options.namespace,
                self._stop,
            )
        )

    async def wait(self) -> None:
        """Wait for a stop request; an HTTP exit is fatal, an HTTPS exit only degrades."""
        watched = {self._http_task, self._stop_task}
        if self._https_task is not None:
            watched.add(self._https_task)
        while not self._stop.is_set():
            done, _pending = await asyncio.wait(watched, return_when=asyncio.FIRST_COMPLETED)
            if self._stop_task in done:
                return
            if self._http_task in done:
                self._http_task.result()
                raise RuntimeError("HTTP listener stopped unexpectedly")
            if self._https_task is not None and self._https_task in done:
                watched.remove(self._https_task)
                logger.warning("HTTPS unavailable; HTTP CLI remains available")

    async def stop(self) -> None:
        """Stop both listeners and wait for cleanup (idempotent).

        HTTP exits through uvicorn so the lifespan shutdown runs; HTTPS carries no
        lifespan and is cancelled. The two log lines show whether cleanup finished.
        """
        started_at = time.monotonic()
        logger.info("stopping listeners")
        self._stop.set()
        if self._http_server is not None:
            self._http_server.should_exit = True
        if self._https_task is not None:
            self._https_task.cancel()
            await asyncio.gather(self._https_task, return_exceptions=True)
        if self._http_task is not None:
            try:
                await asyncio.wait_for(self._http_task, timeout=SHUTDOWN_TIMEOUT_SECONDS)
            except Exception:
                self._http_task.cancel()
                await asyncio.gather(self._http_task, return_exceptions=True)
        if self._stop_task is not None:
            self._stop_task.cancel()
            await asyncio.gather(self._stop_task, return_exceptions=True)
        logger.info("shutdown complete in %.2fs", time.monotonic() - started_at)

    async def run(self) -> None:
        """Start and wait for a stop request (no process signals; for embedding and tests)."""
        await self.start()
        await self.wait()

    async def run_until_signal(self) -> None:
        """Install process signals and run until SIGINT/SIGTERM or an HTTP failure."""
        loop = asyncio.get_running_loop()
        self._registered_signals = install_signal_handlers(self._stop, loop)
        try:
            await self.run()
        finally:
            try:
                await self.stop()
            finally:
                remove_signal_handlers(self._registered_signals, loop)
                self._registered_signals = []


def parse_port_argument(value: str) -> int:
    port = int(value)
    if not MIN_PORT <= port <= MAX_PORT:
        raise argparse.ArgumentTypeError(f"port must be between {MIN_PORT} and {MAX_PORT}")
    return port


def main() -> None:
    parser = argparse.ArgumentParser(description="ClusterOps HTTP CLI and HTTPS WebUI")
    parser.add_argument(
        "--port", type=parse_port_argument, default=os.environ.get(HTTP_PORT_ENV, str(DEFAULT_HTTP_PORT))
    )
    parser.add_argument(
        "--https-port",
        type=parse_port_argument,
        default=os.environ.get(HTTPS_PORT_ENV, str(DEFAULT_HTTPS_PORT)),
    )
    parser.add_argument("--namespace", default=None)
    parser.add_argument("--host", default=LISTEN_HOST)
    args = parser.parse_args()

    options = ListenOptions(http_port=args.port, https_port=args.https_port, namespace=args.namespace, host=args.host)
    asyncio.run(ClusterOpsServer(app, options).run_until_signal())


if __name__ == "__main__":
    main()
