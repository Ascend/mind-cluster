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

"""Request-surface decisions and credential parsing (pure functions), without ASGI."""

import base64

import pytest

from agent_core.https.access_policy import (
    RequestSurface,
    classify_surface,
    is_anonymous_https_path,
    is_cli_path,
    read_bearer_token,
    token_matches,
)


@pytest.mark.parametrize("path", ["/diag", "/tools", "/tools/faulty_device"])
def test_http_cli_paths_are_exactly_matched(path):
    assert is_cli_path(path)
    assert classify_surface("http", path) is RequestSurface.CLI


@pytest.mark.parametrize("path", ["/", "/diag/", "/tools/", "/tools/a/b", "/ui", "/ui/app.js", "/chat/stream"])
def test_http_hides_every_non_cli_path(path):
    assert not is_cli_path(path)
    assert classify_surface("http", path) is RequestSurface.NOT_FOUND


@pytest.mark.parametrize("scheme", [None, "", "ftp", "HTTPS", "ws", "wss"])
def test_unknown_schemes_fail_closed(scheme):
    assert classify_surface(scheme, "/tools") is RequestSurface.NOT_FOUND


@pytest.mark.parametrize("path", ["/ui", "/ui/", "/ui/index.html"])
def test_ui_page_entries_are_protected(path):
    assert classify_surface("https", path) is RequestSurface.PROTECTED_UI_ENTRY


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
        "/ui/vendor/marked-LICENSE.md",
    ],
)
def test_login_page_and_static_assets_are_anonymous(path):
    assert is_anonymous_https_path(path)


@pytest.mark.parametrize(
    "path",
    ["/ui/", "/ui/index.html", "/ui/private.json", "/ui/docs/README.md", "/ui/extra.html", "/uievil"],
)
def test_html_and_data_paths_are_never_anonymous(path):
    assert not is_anonymous_https_path(path)


def test_new_frontend_file_needs_no_whitelist_change():
    """Why the allow list is rule-based: new WebUI front-end files need no backend change."""
    assert is_anonymous_https_path("/ui/views/brand-new-view.js")
    assert is_anonymous_https_path("/ui/vendor/new-library.min.css")


def test_login_route_and_business_paths_classification():
    assert classify_surface("https", "/ui/login") is RequestSurface.LOGIN_ENTRY
    assert classify_surface("https", "/ui/app.js") is RequestSurface.ANONYMOUS
    assert classify_surface("https", "/api/v1/status") is RequestSurface.PROTECTED
    assert classify_surface("https", "/chat/stream") is RequestSurface.PROTECTED
    assert classify_surface("https", "/diag") is RequestSurface.PROTECTED


def test_plain_bearer_token_is_returned_verbatim():
    headers = [(b"authorization", b"Bearer   Token-With-Space ")]
    assert read_bearer_token(headers) == b"  Token-With-Space "


def test_base64_transport_preserves_unicode_and_spaces():
    token = "令牌 带空格 "
    encoded = base64.b64encode(token.encode("utf-8"))
    headers = [(b"authorization", b"Bearer " + encoded), (b"x-webui-token-encoding", b"base64")]
    assert read_bearer_token(headers) == token.encode("utf-8")


@pytest.mark.parametrize(
    "headers",
    [
        [],
        [(b"authorization", b"Token abc")],
        [(b"authorization", b"Bearer abc"), (b"authorization", b"Bearer abc")],
        [(b"authorization", b"Bearer abc"), (b"x-webui-token-encoding", b"BASE64")],
        [(b"authorization", b"Bearer %%%"), (b"x-webui-token-encoding", b"base64")],
        [(b"authorization", b"Bearer abc"), (b"x-webui-token-encoding", b"unknown")],
    ],
)
def test_malformed_or_duplicate_credentials_are_rejected(headers):
    assert read_bearer_token(headers) is None


def test_token_comparison_is_exact_and_fails_closed():
    assert token_matches(b"token", b"token")
    assert not token_matches(b"token", b"token ")
    assert not token_matches(b"", b"token")
    assert not token_matches(b"token", None)
