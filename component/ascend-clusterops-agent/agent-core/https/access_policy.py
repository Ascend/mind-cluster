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

"""Pure HTTPS request-surface rules: entry classification, anonymous assets, Bearer parsing.

Only decisions live here (no responses, logging or state), so the boundaries can be
verified without an ASGI server.
"""

from __future__ import annotations

import base64
import binascii
import hmac
import re
from enum import Enum, auto
from typing import Optional, Sequence, Tuple

CLI_PATHS = frozenset({"/diag", "/tools"})
LOGIN_ROUTE = "/ui/login"
LOGIN_PAGE = "/ui/login.html"
PROTECTED_UI_ENTRY_PATHS = frozenset({"/ui", "/ui/", "/ui/index.html"})
_UI_PREFIX = "/ui/"
_TOOL_PATH = re.compile(r"/tools/[^/]+\Z")
_ANONYMOUS_SUFFIXES = (".js", ".css")
_VENDOR_PREFIX = "vendor/"

BEARER_PREFIX = b"Bearer "
TOKEN_ENCODING_HEADER = b"x-webui-token-encoding"
BASE64_TOKEN_ENCODING = b"base64"


class RequestSurface(Enum):
    """Handling surface of one request.

    CLI and NOT_FOUND belong to the HTTP 9700 entry; the rest belong to HTTPS. An
    unauthenticated page entry returns the login HTML, other protected paths return a
    JSON error; both use 200 because the path itself is valid.
    """

    CLI = auto()
    NOT_FOUND = auto()
    LOGIN_ENTRY = auto()
    ANONYMOUS = auto()
    PROTECTED_UI_ENTRY = auto()
    PROTECTED = auto()


def is_cli_path(path: str) -> bool:
    """HTTP 9700 serves exact CLI paths only; no prefix matching."""
    return path in CLI_PATHS or _TOOL_PATH.fullmatch(path) is not None


def is_anonymous_https_path(path: str) -> bool:
    """Paths that need no token over HTTPS.

    Login pages, JS/CSS under ``/ui`` and vendor assets are anonymous; HTML entries and
    every other extension require authentication, so new front-end files need no
    whitelist update.
    """
    if path in (LOGIN_ROUTE, LOGIN_PAGE):
        return True
    if not path.startswith(_UI_PREFIX):
        return False
    relative = path[len(_UI_PREFIX) :]
    if not relative or relative.endswith("/") or relative.endswith(".html"):
        return False
    if relative.startswith(_VENDOR_PREFIX):
        return True
    return relative.endswith(_ANONYMOUS_SUFFIXES)


def classify_surface(scheme: Optional[str], path: str) -> RequestSurface:
    """Classify the surface by listener scheme and path (Host and forwarded headers are ignored)."""
    if scheme == "http":
        return RequestSurface.CLI if is_cli_path(path) else RequestSurface.NOT_FOUND
    if scheme != "https":
        return RequestSurface.NOT_FOUND
    if path == LOGIN_ROUTE:
        return RequestSurface.LOGIN_ENTRY
    if is_anonymous_https_path(path):
        return RequestSurface.ANONYMOUS
    if path in PROTECTED_UI_ENTRY_PATHS:
        return RequestSurface.PROTECTED_UI_ENTRY
    return RequestSurface.PROTECTED


def read_bearer_token(headers: Sequence[Tuple[bytes, bytes]]) -> Optional[bytes]:
    """Return the raw token bytes; multiple Authorization headers are rejected.

    A base64 value must carry ``X-WebUI-Token-Encoding: base64``; otherwise the bytes
    are compared as provided.
    """
    authorizations = [value for name, value in headers if name.lower() == b"authorization"]
    if len(authorizations) != 1 or not authorizations[0].startswith(BEARER_PREFIX):
        return None
    provided = authorizations[0][len(BEARER_PREFIX) :]
    encodings = [value for name, value in headers if name.lower() == TOKEN_ENCODING_HEADER]
    if not encodings:
        return provided
    if encodings != [BASE64_TOKEN_ENCODING]:
        return None
    try:
        return base64.b64decode(provided, validate=True)
    except (ValueError, binascii.Error):
        return None


def token_matches(expected: bytes, provided: Optional[bytes]) -> bool:
    """Constant-time comparison; an empty expected token always denies."""
    if not expected or provided is None:
        return False
    return hmac.compare_digest(provided, expected)
