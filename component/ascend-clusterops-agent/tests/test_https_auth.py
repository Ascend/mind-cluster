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

"""Middleware responses: entry boundaries, 200 on auth failure, 404/429, headers, containment."""

import asyncio
import base64

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from agent_core.https import rate_limit
from agent_core.https.auth_middleware import (
    AUTH_FAILURE_CODE,
    AUTH_RESULT_HEADER,
    AUTH_RESULT_UNAUTHENTICATED,
    CSP,
    WebUIAuthMiddleware,
)


def _assert_unauthenticated(response):
    """Valid path with bad credentials: HTTP-level success, the reason is in header and body."""
    assert response.status_code == 200
    assert response.headers[AUTH_RESULT_HEADER] == AUTH_RESULT_UNAUTHENTICATED


def _route_registered(app, path):
    """Whether the application registers the path; /ui and /tools come from other branches."""
    return any(getattr(route, "path", None) == path for route in app.router.routes)


@pytest.fixture(autouse=True)
def isolated_rate_limit_env(monkeypatch):
    """Dev machines and CI may define these variables; the default window assertions isolate them."""
    monkeypatch.delenv("WEBUI_RATELIMIT_MAX_REQUESTS", raising=False)
    monkeypatch.delenv("WEBUI_RATELIMIT_WINDOW_SECONDS", raising=False)


def _app(token="valid-token"):
    app = FastAPI()
    app.state.webui_token = token
    app.add_middleware(WebUIAuthMiddleware)

    @app.get("/{path:path}")
    async def any_path(path):
        if path == "boom":
            raise RuntimeError("failure")
        return {"ok": True}

    return app


def _app_with_fixed_routes(token="valid-token"):
    """Register explicit paths only, to verify that unknown paths return 404 before authentication."""
    app = FastAPI()
    app.state.webui_token = token
    app.add_middleware(WebUIAuthMiddleware)

    @app.get("/api/v1/status")
    async def status():
        return {"ok": True}

    @app.get("/ui/")
    async def ui_page():
        return {"page": True}

    return app


def _headers(response, https):
    assert response.headers["x-content-type-options"] == "nosniff"
    assert response.headers["x-frame-options"] == "DENY"
    assert response.headers["cache-control"] == "no-store"
    assert response.headers["content-security-policy"] == CSP
    if https:
        assert response.headers["strict-transport-security"] == "max-age=31536000"
    else:
        assert "strict-transport-security" not in response.headers


@pytest.mark.parametrize("path", ["/diag", "/tools", "/tools/faulty_device"])
def test_http_keeps_only_exact_cli_paths(path):
    with TestClient(_app(), base_url="http://testserver") as client:
        response = client.get(path)
    assert response.status_code == 200
    _headers(response, False)


@pytest.mark.parametrize(
    "path",
    ["/", "/diag/", "/tools/", "/tools/a/b", "/ui", "/ui/app.js", "/chat/stream", "/api/v1/status", "/docs"],
)
def test_http_hides_every_non_cli_path_even_with_token_and_forwarded_scheme(path):
    with TestClient(_app(), base_url="http://testserver") as client:
        response = client.get(path, headers={"Authorization": "Bearer valid-token", "X-Forwarded-Proto": "https"})
    assert response.status_code == 404
    _headers(response, False)


@pytest.mark.parametrize(
    "authorization", [None, "Bearer wrong", "Bearer  valid-token", "Bearer valid-token ", "bearer valid-token"]
)
def test_https_business_reports_invalid_tokens_with_200(authorization):
    headers = {} if authorization is None else {"Authorization": authorization}
    headers["X-Forwarded-Proto"] = "http"
    with TestClient(_app(), base_url="https://testserver") as client:
        response = client.get("/api/v1/status", headers=headers)
    _assert_unauthenticated(response)
    assert response.json()["ok"] is False
    assert response.json()["error"]["code"] == AUTH_FAILURE_CODE
    assert "www-authenticate" not in response.headers
    _headers(response, True)


def test_unknown_path_is_404_before_authentication():
    """Unknown paths are decided before authentication, so 200 + auth failure proves the path exists."""
    with TestClient(_app_with_fixed_routes(""), base_url="https://testserver") as client:
        response = client.get("/no-such-path")
    assert response.status_code == 404
    assert AUTH_RESULT_HEADER not in response.headers
    _headers(response, True)


def test_bad_token_tells_known_and_unknown_paths_apart():
    with TestClient(_app_with_fixed_routes(), base_url="https://testserver") as client:
        known = client.get("/api/v1/status", headers={"Authorization": "Bearer wrong"})
        unknown = client.get("/no-such-path", headers={"Authorization": "Bearer wrong"})
    _assert_unauthenticated(known)
    assert unknown.status_code == 404


@pytest.mark.parametrize("path", ["/ui", "/ui/", "/ui/index.html"])
def test_https_ui_entry_returns_login_html_without_a_token(path):
    with TestClient(_app(""), base_url="https://testserver") as client:
        response = client.get(path)
    _assert_unauthenticated(response)
    assert response.headers["content-type"].startswith("text/html")
    assert "login-form" in response.text
    _headers(response, True)


@pytest.mark.parametrize("path", ["/ui", "/ui/", "/ui/index.html"])
def test_https_ui_entry_reaches_the_app_with_a_token(path):
    with TestClient(_app(), base_url="https://testserver") as client:
        response = client.get(path, headers={"Authorization": "Bearer valid-token"})
    assert response.status_code == 200
    _headers(response, True)


@pytest.mark.parametrize(
    "path",
    [
        "/ui/login",
        "/ui/login.html",
        "/ui/login.css",
        "/ui/auth.js",
        "/ui/login.js",
        "/ui/style.css",
        "/ui/app.js",
        "/ui/sse.js",
        "/ui/views/chat.js",
        "/ui/views/toolcard.js",
        "/ui/vendor/marked.esm.js",
        "/ui/vendor/purify.min.js",
    ],
)
def test_https_login_and_static_assets_are_anonymous(path):
    with TestClient(_app(""), base_url="https://testserver") as client:
        response = client.get(path)
    assert response.status_code == 200
    _headers(response, True)


def test_https_login_route_is_mapped_to_the_login_file():
    captured = []

    async def application(scope, _receive, _send):
        captured.append(scope["path"])

    async def receive():
        return {"type": "http.request", "body": b""}

    async def send(_message):
        return None

    scope = {"type": "http", "scheme": "https", "path": "/ui/login", "headers": [], "app": None}
    asyncio.run(WebUIAuthMiddleware(application)(scope, receive, send))
    assert captured == ["/ui/login.html"]


def test_static_code_is_anonymous_but_data_and_pages_are_not():
    """Rule-based allow list: new front-end files pass, HTML and data paths still need auth."""
    with TestClient(_app(), base_url="https://testserver") as client:
        anonymous = client.get("/ui/views/brand-new-view.js")
        assert anonymous.status_code == 200
        assert AUTH_RESULT_HEADER not in anonymous.headers
        for path in ("/uievil", "/ui/docs/README.md", "/ui/private.json"):
            _assert_unauthenticated(client.get(path))


def test_correct_token_and_empty_expected_token_fail_closed(caplog):
    app = _app()
    with TestClient(app, base_url="https://testserver") as client:
        assert client.get("/api/v1/status", headers={"Authorization": "Bearer valid-token"}).status_code == 200
        app.state.webui_token = ""
        response = client.get("/api/v1/status", headers={"Authorization": "Bearer "})
    _assert_unauthenticated(response)
    assert "token is empty" in caplog.text
    assert "valid-token" not in caplog.text


@pytest.mark.parametrize("token", ["令牌中文", "  leading-and-trailing  ", "same-token"])
def test_explicit_base64_transport_compares_original_utf8_bytes(token):
    encoded = base64.b64encode(token.encode("utf-8")).decode("ascii")
    headers = {"Authorization": "Bearer " + encoded, "X-WebUI-Token-Encoding": "base64"}
    with TestClient(_app(token), base_url="https://testserver") as client:
        assert client.get("/api/v1/status", headers=headers).status_code == 200
        altered = base64.b64encode((token + " ").encode("utf-8")).decode("ascii")
        headers["Authorization"] = "Bearer " + altered
        _assert_unauthenticated(client.get("/api/v1/status", headers=headers))


def test_raw_utf8_bearer_remains_compatible():
    token = "中文令牌"
    with TestClient(_app(token), base_url="https://testserver") as client:
        response = client.get("/api/v1/status", headers=[(b"authorization", b"Bearer " + token.encode("utf-8"))])
    assert response.status_code == 200


@pytest.mark.parametrize(
    "encoding,value",
    [("BASE64", "dmFsaWQtdG9rZW4="), ("base64", "%%%"), ("base64", "dmFsaWQ=\n"), ("unknown", "valid-token")],
)
def test_unknown_or_malformed_transport_encoding_is_rejected(encoding, value):
    """Drive ASGI directly so the HTTP client cannot reject deliberately malformed headers."""
    messages = []
    middleware = WebUIAuthMiddleware(lambda *_args: None)
    scope = {
        "type": "http",
        "scheme": "https",
        "path": "/api/v1/status",
        "headers": [
            (b"authorization", b"Bearer " + value.encode("ascii")),
            (b"x-webui-token-encoding", encoding.encode("ascii")),
        ],
    }

    async def receive():
        return {"type": "http.request", "body": b""}

    async def send(message):
        messages.append(message)

    asyncio.run(middleware(scope, receive, send))
    assert messages[0]["status"] == 200
    assert dict(messages[0]["headers"])[
        AUTH_RESULT_HEADER.lower().encode("ascii")
    ] == AUTH_RESULT_UNAUTHENTICATED.encode("ascii")


def test_duplicate_authorization_is_rejected():
    with TestClient(_app(), base_url="https://testserver") as client:
        response = client.get(
            "/api/v1/status", headers=[("Authorization", "Bearer valid-token"), ("Authorization", "Bearer valid-token")]
        )
    _assert_unauthenticated(response)


def test_unexpected_error_keeps_security_headers_and_omits_exception():
    with TestClient(_app(), base_url="https://testserver") as client:
        response = client.get("/boom", headers={"Authorization": "Bearer valid-token"})
    assert response.status_code == 500
    assert response.json() == {"detail": "Internal Server Error"}
    _headers(response, True)


@pytest.mark.parametrize("scheme", [None, "", "ftp", "HTTPS", "ws", "wss"])
def test_unknown_transport_scheme_fails_closed_even_for_cli(scheme):
    messages = []
    calls = []

    async def application(_scope, _receive, _send):
        calls.append(True)

    async def receive():
        return {"type": "http.request", "body": b""}

    async def send(message):
        messages.append(message)

    scope = {"type": "http", "scheme": scheme, "path": "/tools", "headers": []}
    asyncio.run(WebUIAuthMiddleware(application)(scope, receive, send))
    assert not calls
    assert messages[0]["status"] == 404
    headers = dict(messages[0]["headers"])
    assert headers[b"cache-control"] == b"no-store"
    assert b"strict-transport-security" not in headers


def test_rate_limit_rejects_protected_paths_but_keeps_login_reachable(monkeypatch):
    now = [0.0]
    monkeypatch.setenv("WEBUI_RATELIMIT_MAX_REQUESTS", "2")
    monkeypatch.setenv("WEBUI_RATELIMIT_WINDOW_SECONDS", "10")
    monkeypatch.setattr(rate_limit, "monotonic", lambda: now[0])

    with TestClient(_app(), base_url="https://testserver") as client:
        token = {"Authorization": "Bearer valid-token"}
        assert client.get("/api/v1/status", headers=token).status_code == 200
        assert client.get("/api/v1/status", headers=token).status_code == 200
        rejected = client.get("/api/v1/status", headers=token)
        assert rejected.status_code == 429
        assert rejected.json() == {"detail": "Too Many Requests"}
        _headers(rejected, True)
        assert client.get("/ui/login").status_code == 200
        assert client.get("/ui/login.css").status_code == 200
        now[0] = 11.0
        assert client.get("/api/v1/status", headers=token).status_code == 200


def test_sliding_window_has_no_boundary_burst(monkeypatch):
    """A fixed window resets at t=10; the sliding window still counts the last 10 seconds."""
    now = [0.0]
    monkeypatch.setattr(rate_limit, "monotonic", lambda: now[0])
    limiter = rate_limit.SlidingWindowRateLimiter(2, 10)
    now[0] = 9.9
    assert limiter.allow_request() is True
    now[0] = 9.95
    assert limiter.allow_request() is True
    now[0] = 10.0
    assert limiter.allow_request() is False
    now[0] = 19.96
    assert limiter.allow_request() is True
    now[0] = 20.0
    assert limiter.allow_request() is True
    now[0] = 20.1
    assert limiter.allow_request() is False


def test_rejected_requests_do_not_extend_the_lockout(monkeypatch):
    """Rejected requests are not recorded, so a flood cannot lock the quota permanently."""
    now = [0.0]
    monkeypatch.setattr(rate_limit, "monotonic", lambda: now[0])
    limiter = rate_limit.SlidingWindowRateLimiter(1, 10)
    assert limiter.allow_request() is True
    for second in (1, 2, 3, 4, 5, 6, 7, 8, 9):
        now[0] = float(second)
        assert limiter.allow_request() is False
    now[0] = 10.1
    assert limiter.allow_request() is True


@pytest.mark.parametrize(
    "max_requests,window",
    [("0", "0"), ("-1", "-1"), ("1000001", "3601"), ("not-a-number", "not-a-number")],
)
def test_invalid_rate_limit_env_falls_back_with_warning(monkeypatch, caplog, max_requests, window):
    monkeypatch.setenv("WEBUI_RATELIMIT_MAX_REQUESTS", max_requests)
    monkeypatch.setenv("WEBUI_RATELIMIT_WINDOW_SECONDS", window)
    middleware = WebUIAuthMiddleware(lambda *_args: None)
    assert middleware._limiter.max_requests == 600
    assert middleware._limiter.window_seconds == 60.0
    assert "WEBUI_RATELIMIT_MAX_REQUESTS" in caplog.text
    assert "WEBUI_RATELIMIT_WINDOW_SECONDS" in caplog.text


def test_real_application_routes_protect_auth_check_and_keep_cli(monkeypatch):
    from contextlib import asynccontextmanager

    # agent_core.main 的导入链包含对话面（api.stream_api → chat.graph → langchain）；
    # 流水线旧镜像没有 langchain 时跳过本条，而不是失败。
    pytest.importorskip("langchain")

    from agent_core import main

    @asynccontextmanager
    async def no_cluster_startup(_app):
        yield

    monkeypatch.setattr(main.app.router, "lifespan_context", no_cluster_startup)
    monkeypatch.setattr(main.tools, "diagnose_cached", lambda **_kwargs: {"error": "not found"})
    monkeypatch.setattr(main.app.state, "webui_token", "test-token", raising=False)
    with TestClient(main.app, base_url="https://testserver") as client:
        invalid_check = client.post("/api/v1/auth/check")
        _assert_unauthenticated(invalid_check)
        assert invalid_check.json()["ok"] is False
        response = client.post("/api/v1/auth/check", headers={"Authorization": "Bearer test-token"})
        assert response.status_code == 200 and response.json() == {"ok": True}
        # /ui is mounted by the WebUI baseline; without it every /ui/* path is unknown and returns 404.
        if _route_registered(main.app, "/ui"):
            _assert_unauthenticated(client.get("/ui/auth"))
            assert client.get("/ui/auth", headers={"Authorization": "Bearer test-token"}).status_code == 404
        _assert_unauthenticated(client.post("/diag", json={"job": "missing"}))
        response = client.post("/diag", json={"job": "missing"}, headers={"Authorization": "Bearer test-token"})
        assert response.status_code == 200 and response.json()["error"] == "not found"
    with TestClient(main.app, base_url="http://testserver") as client:
        assert client.post("/api/v1/auth/check").status_code == 404
        assert client.post("/diag", json={"job": "missing"}).status_code == 200
        if _route_registered(main.app, "/tools"):
            assert client.get("/tools").status_code == 200
