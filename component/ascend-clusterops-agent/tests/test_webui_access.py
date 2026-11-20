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
"""WebUI page entry: the bare /ui redirect and the first-paint state.

The access plane is decided by the listener port (9700 serves the CLI paths only, 9743 is
the token-protected HTTPS surface), so the requests here speak HTTPS with a valid token;
the boundaries themselves are covered by test_https_auth.py.
"""

# pylint: disable=redefined-outer-name  # pytest fixtures share names with test parameters

from __future__ import annotations

import re

import pytest
from fastapi.testclient import TestClient

# agent_core.main 的导入链需要 langchain/langgraph：流水线镜像若没有（旧镜像为 py3.9），
# 整文件跳过即可，不让收集阶段报 ImportError。
pytest.importorskip("langchain")

from agent_core import main  # noqa: E402  # isort: skip  (guard must run before the import)

TOKEN_HEADERS = {"Authorization": "Bearer test-token"}


@pytest.fixture
def client(monkeypatch):
    monkeypatch.setattr(main.app.state, "webui_token", "test-token", raising=False)
    return TestClient(main.app, base_url="https://testserver")  # no context: lifespan must not run


def test_the_bare_page_url_redirects_so_relative_assets_resolve(client):
    """页面资源是相对路径：裸 /ui 必须跳到 /ui/，否则浏览器把 style.css 解析到根目录而全部 404。"""
    bare = client.get("/ui", headers=TOKEN_HEADERS, follow_redirects=False)
    assert bare.status_code == 307
    assert bare.headers["location"].endswith("/ui/")
    refs = re.findall(r'(?:src|href)="([^"]+)"', client.get("/ui/", headers=TOKEN_HEADERS).text)
    assert refs and not [ref for ref in refs if ref.startswith("/")], "资源路径必须保持相对，尾斜杠入口才兜得住"


def test_the_page_carries_the_first_paint_state_inline(client, monkeypatch):
    """首屏的「有没有模型」随页面内联下发，页面不再另请求状态接口。"""
    monkeypatch.setattr(main, "_llm_ready", lambda: True)
    page = client.get("/ui/", headers=TOKEN_HEADERS)
    assert page.status_code == 200
    assert "window.__AGENT_BOOT__" in page.text and '"llm_available": true' in page.text
    assert "<!--AGENT_BOOT-->" not in page.text, "the marker must be replaced, not shipped verbatim"
    assert page.headers.get("cache-control") == "no-store"
    monkeypatch.setattr(main, "_llm_ready", lambda: False)
    assert '"llm_available": false' in client.get("/ui/index.html", headers=TOKEN_HEADERS).text
