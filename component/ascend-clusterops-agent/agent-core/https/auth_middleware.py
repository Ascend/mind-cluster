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

"""HTTPS authentication middleware: entry decisions, token checks, rate limiting, security headers.

Surface decisions and credential parsing live in ``access_policy`` (pure functions),
window counting in ``rate_limit``; this module only builds responses.

Failed authentication does not use an HTTP error code: a valid path with bad credentials
returns 200 with the reason in the body and the ``X-WebUI-Auth`` header. Only 404, 429
and 500 use error status codes.
"""

from __future__ import annotations

import logging
from pathlib import Path
from time import monotonic

from starlette.responses import HTMLResponse, JSONResponse
from starlette.routing import Match

from agent_core.https.bounded_env import env_float, env_int
from agent_core.https.access_policy import (
    LOGIN_PAGE,
    RequestSurface,
    classify_surface,
    read_bearer_token,
    token_matches,
)
from agent_core.https.rate_limit import SlidingWindowRateLimiter

logger = logging.getLogger(__name__)

CSP = (
    "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; "
    "img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; "
    "frame-ancestors 'none'; form-action 'self'"
)
HSTS_MAX_AGE_SECONDS = 31536000
_SECURITY_HEADERS = (
    (b"x-content-type-options", b"nosniff"),
    (b"x-frame-options", b"DENY"),
    (b"content-security-policy", CSP.encode("ascii")),
    (b"cache-control", b"no-store"),
)
_SECURITY_HEADER_NAMES = {name for name, _value in _SECURITY_HEADERS} | {b"strict-transport-security"}

HTTP_STATUS_OK = 200
HTTP_STATUS_NOT_FOUND = 404
HTTP_STATUS_TOO_MANY_REQUESTS = 429
HTTP_STATUS_INTERNAL_SERVER_ERROR = 500

AUTH_RESULT_HEADER = "X-WebUI-Auth"
AUTH_RESULT_UNAUTHENTICATED = "unauthenticated"
AUTH_FAILURE_CODE = "UNAUTHENTICATED"
AUTH_FAILURE_MESSAGE = "共享 token 无效或缺失"
AUTH_FAILURE_LOG_INTERVAL_SECONDS = 60.0
_AUTH_FAILURE_HEADERS = {AUTH_RESULT_HEADER: AUTH_RESULT_UNAUTHENTICATED}

RATE_LIMIT_MAX_ENV = "WEBUI_RATELIMIT_MAX_REQUESTS"
RATE_LIMIT_WINDOW_ENV = "WEBUI_RATELIMIT_WINDOW_SECONDS"
DEFAULT_MAX_REQUESTS = 600
DEFAULT_WINDOW_SECONDS = 60.0
MIN_MAX_REQUESTS, MAX_MAX_REQUESTS = 1, 1_000_000
MIN_WINDOW_SECONDS, MAX_WINDOW_SECONDS = 1.0, 3600.0

LOGIN_PAGE_PATH = Path(__file__).resolve().parent.parent / "webui" / "login.html"
_FALLBACK_LOGIN_HTML = (
    '<!doctype html><html lang="zh-CN"><meta charset="utf-8">'
    "<title>ClusterOps HTTPS</title><p>请通过 <code>/ui/login</code> 登录。</p></html>"
)


def load_login_page_html() -> str:
    """Return the login page, falling back to a self-explanatory page when it is missing."""
    try:
        return LOGIN_PAGE_PATH.read_text(encoding="utf-8")
    except OSError:
        return _FALLBACK_LOGIN_HTML


def read_rate_limit_settings() -> tuple[int, float]:
    """Read the rate-limit settings; invalid values fall back to the defaults with a warning."""
    max_requests = env_int(RATE_LIMIT_MAX_ENV, DEFAULT_MAX_REQUESTS, MIN_MAX_REQUESTS, MAX_MAX_REQUESTS)
    window_seconds = env_float(RATE_LIMIT_WINDOW_ENV, DEFAULT_WINDOW_SECONDS, MIN_WINDOW_SECONDS, MAX_WINDOW_SECONDS)
    return max_requests, window_seconds


def add_security_headers(message: dict, https: bool) -> dict:
    """Append security headers, dropping existing ones with the same name.

    HSTS is sent on HTTPS responses only.
    """
    if message["type"] != "http.response.start":
        return message
    headers = [
        (name, value) for name, value in message.get("headers", []) if name.lower() not in _SECURITY_HEADER_NAMES
    ]
    headers.extend(_SECURITY_HEADERS)
    if https:
        headers.append((b"strict-transport-security", f"max-age={HSTS_MAX_AGE_SECONDS}".encode("ascii")))
    return {**message, "headers": headers}


def rewrite_to_login_page(scope) -> dict:
    """StaticFiles only resolves real files, so the explicit login route is rewritten."""
    rewritten = dict(scope)
    rewritten["path"] = LOGIN_PAGE
    rewritten["raw_path"] = LOGIN_PAGE.encode("ascii")
    return rewritten


def route_exists(scope) -> bool:
    """Whether the path is registered in the application.

    Unknown paths return 404 before authentication, so a 200 with a bad token proves the
    path exists. Routes are matched one by one (``router.matches`` only covers mounted
    sub-routers). When matching fails the path is treated as existing.
    """
    router = getattr(scope.get("app"), "router", None)
    if router is None:
        return True
    try:
        return any(route.matches(scope)[0] is not Match.NONE for route in router.routes)
    except Exception:
        return True


class WebUIAuthMiddleware:
    """ASGI middleware that decides by listener scheme.

    Everything comes from runtime state: the token from ``app.state.webui_token`` and the
    rate-limit settings from the environment, so no constructor arguments are needed.
    """

    def __init__(self, app):
        self.app = app
        self._login_html = load_login_page_html()
        self._warned_empty_token = False
        self._last_auth_failure_log = -AUTH_FAILURE_LOG_INTERVAL_SECONDS
        max_requests, window_seconds = read_rate_limit_settings()
        self._limiter = SlidingWindowRateLimiter(max_requests, window_seconds)

    async def __call__(self, scope, receive, send):
        """Handle one ASGI request.

        Non-HTTP messages pass through. Every HTTPS request counts against the rate limit,
        but only protected paths are rejected with 429, so login and static assets stay
        reachable.
        """
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        scheme = scope.get("scheme")
        https = scheme == "https"
        surface = classify_surface(scheme, scope.get("path", ""))

        async def secure_send(message):
            await send(add_security_headers(message, https))

        if surface is RequestSurface.NOT_FOUND:
            await self._send_error_response(scope, receive, secure_send, HTTP_STATUS_NOT_FOUND, "Not Found")
            return
        if surface is RequestSurface.CLI:
            await self._forward_to_application(scope, receive, secure_send)
            return

        allowed = self._limiter.allow_request()
        if surface in (RequestSurface.ANONYMOUS, RequestSurface.LOGIN_ENTRY):
            target = rewrite_to_login_page(scope) if surface is RequestSurface.LOGIN_ENTRY else scope
            await self._forward_to_application(target, receive, secure_send)
            return
        if not allowed:
            await self._send_error_response(
                scope, receive, secure_send, HTTP_STATUS_TOO_MANY_REQUESTS, "Too Many Requests"
            )
            return
        if not route_exists(scope):
            await self._send_error_response(scope, receive, secure_send, HTTP_STATUS_NOT_FOUND, "Not Found")
            return
        if not self._has_valid_credentials(scope):
            self._log_authentication_failure()
            await self._send_authentication_required_response(scope, receive, secure_send, surface)
            return
        await self._forward_to_application(scope, receive, secure_send)

    def _has_valid_credentials(self, scope) -> bool:
        expected = self._current_expected_token(scope)
        return token_matches(expected, read_bearer_token(scope.get("headers", [])))

    def _current_expected_token(self, scope) -> bytes:
        """Token in effect; an empty token warns once and denies every business request."""
        state = getattr(scope.get("app"), "state", None)
        token = getattr(state, "webui_token", "")
        try:
            encoded = token.encode("utf-8") if isinstance(token, str) else b""
        except UnicodeEncodeError:
            encoded = b""
        if not encoded and not self._warned_empty_token:
            logger.warning("WebUI token is empty; HTTPS business requests are denied")
            self._warned_empty_token = True
        return encoded

    async def _forward_to_application(self, scope, receive, send) -> None:
        """Forward to the application; failures before the response start become a 500."""
        started = False

        async def tracking_send(message):
            nonlocal started
            if message["type"] == "http.response.start":
                started = True
            await send(message)

        try:
            await self.app(scope, receive, tracking_send)
        except Exception:
            if started:
                raise
            logger.error("Request failed before a response was started")
            await self._send_error_response(
                scope, receive, send, HTTP_STATUS_INTERNAL_SERVER_ERROR, "Internal Server Error"
            )

    async def _send_authentication_required_response(self, scope, receive, send, surface: RequestSurface) -> None:
        """Valid path with invalid credentials: 200 with the reason, login HTML for pages."""
        if surface is RequestSurface.PROTECTED_UI_ENTRY:
            response = HTMLResponse(self._login_html, status_code=HTTP_STATUS_OK, headers=_AUTH_FAILURE_HEADERS)
        else:
            response = JSONResponse(
                {"ok": False, "error": {"code": AUTH_FAILURE_CODE, "message": AUTH_FAILURE_MESSAGE}},
                status_code=HTTP_STATUS_OK,
                headers=_AUTH_FAILURE_HEADERS,
            )
        await response(scope, receive, send)

    def _log_authentication_failure(self) -> None:
        """Log authentication failures at a fixed interval to keep log flooding down."""
        now = monotonic()
        if now - self._last_auth_failure_log < AUTH_FAILURE_LOG_INTERVAL_SECONDS:
            return
        self._last_auth_failure_log = now
        logger.warning("WebUI request denied: invalid or missing token")

    async def _send_error_response(self, scope, receive, send, status_code: int, detail: str) -> None:
        await JSONResponse({"detail": detail}, status_code=status_code)(scope, receive, send)
