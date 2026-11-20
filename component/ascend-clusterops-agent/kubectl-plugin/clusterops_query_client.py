#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

"""Shared kubectl client helpers: port-forward, HTTP transport, JSON output, exit codes.

The plugin scripts import this module from their own directory (installed next to
kubectl-clusterops). It never shells out with a command string: port-forward runs
through a subprocess argv array with shell=False and listens on 127.0.0.1 only.
"""

from __future__ import annotations

import http.client
import json
import os
import socket
import subprocess
import time
import urllib.error
import urllib.request
import uuid
from typing import Any, Dict, Optional, Tuple

DEFAULT_SVC = os.environ.get("ASCEND_CLUSTEROPS_AGENT_SVC", "agent-core.mindx-dl:9700")
AGENT_CORE_PORT = 9700
QUERY_TIMEOUT = 60.0  # one query is a short read; the server also bounds each API call
DIAG_TIMEOUT = 1200.0  # diagnosis transport wait is independent of the query timeout

EXIT_OK = 0
EXIT_FAILED = 1
EXIT_ARGUMENT = 2
EXIT_PARTIAL = 3


class TransportError(Exception):
    """Client-side transport failure; rendered as a TRANSPORT_ERROR envelope."""


def parse_service(service: Optional[str]) -> Tuple[str, str, int]:
    """Parse ``svc.ns:port`` (defaults: mindx-dl:9700). Raises ValueError on bad input."""
    value = service or DEFAULT_SVC
    name, namespace, port = value, "mindx-dl", AGENT_CORE_PORT
    try:
        if "." in value:
            name, rest = value.split(".", 1)
            if ":" in rest:
                namespace, raw_port = rest.split(":", 1)
                port = int(raw_port) if raw_port else AGENT_CORE_PORT
            else:
                namespace = rest
    except (TypeError, ValueError) as exc:
        raise ValueError(f"invalid agent-core service {value!r}") from exc
    if not name or not namespace:
        raise ValueError(f"invalid agent-core service {value!r}")
    return name, namespace, port


def _kubectl() -> Optional[str]:
    import shutil

    return shutil.which("kubectl") or shutil.which("kubectl.exe")


def _free_port() -> int:
    sock = socket.socket()
    sock.bind(("127.0.0.1", 0))
    port = sock.getsockname()[1]
    sock.close()
    return port


def _wait_port(proc: subprocess.Popen, local: int, timeout: float = 30.0) -> bool:
    end = time.time() + timeout
    while time.time() < end:
        if proc.poll() is not None:
            return False
        with socket.socket() as probe:
            probe.settimeout(0.5)
            try:
                probe.connect(("127.0.0.1", local))
                return True
            except OSError:
                time.sleep(0.3)
    return False


class PortForward:
    """Temporary 127.0.0.1-only port-forward; always terminated/killed on exit."""

    def __init__(self, service: Optional[str]):
        self.service = service
        self.proc: Optional[subprocess.Popen] = None
        self.local = 0

    def __enter__(self) -> "PortForward":
        kubectl = _kubectl()
        if not kubectl:
            raise TransportError("kubectl not found on PATH")
        name, namespace, port = parse_service(self.service)
        self.local = _free_port()
        cmd = [kubectl, "port-forward", "-n", namespace, f"svc/{name}", f"{self.local}:{port}"]
        self.proc = subprocess.Popen(  # pylint: disable=consider-using-with
            cmd, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE, text=True, shell=False
        )
        if not _wait_port(self.proc, self.local):
            error = self._stop()
            raise TransportError(f"port-forward setup failed: {error.strip()}")
        return self

    def _stop(self) -> str:
        proc = self.proc
        if proc is None:
            return ""
        self.proc = None
        proc.terminate()
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill()
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                return ""
        return proc.stderr.read() if proc.stderr else ""

    def __exit__(self, exc_type, exc, tb) -> None:
        self._stop()


def request_json(
    local: int,
    method: str,
    path: str,
    payload: Optional[Dict[str, Any]] = None,
    timeout: float = QUERY_TIMEOUT,
) -> Tuple[int, Dict[str, Any]]:
    """Call agent-core; HTTP error bodies are preserved as structured results."""
    url = f"http://127.0.0.1:{local}{path}"
    body = json.dumps(payload).encode() if payload is not None else None
    request = urllib.request.Request(url, data=body, method=method, headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            return response.status, _decode(response.read())
    except urllib.error.HTTPError as exc:
        raw = exc.read()
        try:
            return exc.code, _decode(raw)
        except TransportError:
            raise TransportError(f"agent-core returned HTTP {exc.code} without a JSON body") from exc
    except urllib.error.URLError as exc:
        raise TransportError(f"cannot connect to agent-core ({exc.reason})") from exc
    except http.client.HTTPException as exc:
        # urlopen() raises these directly (not wrapped in URLError) when the peer accepts a connection and closes it
        reason = str(exc) or type(exc).__name__
        raise TransportError(f"agent-core closed the connection ({reason})") from exc
    except (TimeoutError, socket.timeout):
        raise TransportError("connection to agent-core timed out") from None
    except OSError as exc:
        raise TransportError(f"cannot connect to agent-core ({exc})") from exc
    except json.JSONDecodeError:
        raise TransportError("agent-core returned invalid JSON") from None


def _decode(raw: bytes) -> Dict[str, Any]:
    data = json.loads(raw.decode("utf-8"))
    if not isinstance(data, dict):
        raise TransportError("agent-core returned a non-object JSON payload")
    return data


def transport_error_envelope(tool: str, message: str) -> Dict[str, Any]:
    """Common error fields for a client-side failure; never invents source evidence."""
    return {
        "request_id": uuid.uuid4().hex,
        "tool": tool,
        "status": "error",
        "elapsed_ms": 0,
        "data": None,
        "errors": [
            {"code": "TRANSPORT_ERROR", "source_id": None, "path": None, "message": message, "retryable": False}
        ],
    }


def exit_code(http_status: int, payload: Dict[str, Any]) -> int:
    status = payload.get("status")
    if status == "partial":
        return EXIT_PARTIAL
    if 200 <= http_status < 300 and status in (None, "ok"):
        return EXIT_OK
    return EXIT_FAILED


def emit(payload: Dict[str, Any], json_mode: bool) -> None:
    """--json prints exactly one JSON object on stdout; human extras go to stderr."""
    if json_mode:
        print(json.dumps(payload, ensure_ascii=False, indent=2))
