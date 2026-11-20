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

"""WebUI 参考页面按文本做门禁。

页面固定五件事：流式气泡、可折叠工具调用卡、新建会话按钮、唯一一句无模型提示、零构建。
这些用例钉住该功能集合，以及「没有会话列表、不重载历史、没有类型化矩阵卡、没有前端引导文案」。
"""

# pylint 把跨文件重复代码报在最后解析的模块上，而重复内容在富文本与工具卡两个用例里：全量跑时
# 这条消息会落在这个文件，与那两处同步抑制。
# pylint: disable=duplicate-code
from __future__ import annotations

import os
import re

from agent_core.chat import events as protocol

UI = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "agent-core", "webui")

# The page files that must ship.
PAGE_FILES = (
    "index.html",
    "style.css",
    "app.js",
    "sse.js",
    "views/chat.js",
    "views/toolcard.js",
    "views/richtext.js",
)

# The capability names the tools registry exposes; the page and the card must know none.
TOOL_NAMES = ("faulty_device", "isolated_device", "cluster_resource", "task_status", "diagnose", "load_skill")

# The sidebar is dropped below this width, with no drawer and no hamburger in its place.
NARROW_SCREEN_MAX_WIDTH = 900


def read(name: str) -> str:
    with open(os.path.join(UI, name), encoding="utf-8") as handle:
        return handle.read()


def test_every_page_file_reads_as_utf8():
    for name in PAGE_FILES:
        assert os.path.isfile(os.path.join(UI, name)), "missing " + name
        assert read(name).strip(), name + " is empty"


def test_retired_page_modules_are_gone():
    for name in ("views/session.js", "views/guide.js"):
        assert not os.path.exists(os.path.join(UI, name)), name + " should not ship"
    # 页面零第三方脚本：不引 CDN、不引 npm，也不许再放回 vendor 目录
    assert not os.path.exists(os.path.join(UI, "vendor")), "the page must not ship a vendor directory"


def test_index_is_a_sidebar_plus_main_pane_page():
    """The script-facing ids survive and the sidebar carries its four parts."""
    html = read("index.html")
    for required in (
        'id="chat-flow"',
        'id="chat-form"',
        'id="chat-input"',
        'id="send-btn"',
        'id="new-session-btn"',
        'id="model-badge"',
        'id="welcome-hint"',
        'name="message"',
        "layout-sidebar",
        'id="sidebar"',
        'id="history-empty"',
        "新建会话",
        "历史会话",
    ):
        assert required in html, "index.html misses " + required
    # no session list, no second/third pane, and no client-side guidance copy
    for retired in (
        'id="session-list"',
        'id="sessions-pane"',
        'id="context-pane"',
        "kubectl clusterops",
        "/api/v1/sessions",
        "EventSource",
        "model-banner",
    ):
        assert retired not in html, "index.html still carries " + retired
    # the model badge lives in the sidebar foot, not in the input bar
    sidebar = html[html.index('id="sidebar"') : html.index("</aside>")]
    input_bar = html[html.index('id="chat-form"') : html.index("</form>")]
    assert 'id="model-badge"' in sidebar, "the badge belongs to the sidebar"
    assert 'id="model-badge"' not in input_bar, "the badge moved out of the input bar"
    assert 'id="new-session-btn"' in sidebar, "the new-session button moved into the sidebar"


def test_the_empty_session_is_a_centered_composer():
    """A fresh session centers the composer under a greeting that sits above it."""
    html = read("index.html")
    css = read("style.css")
    js = read("app.js")
    assert 'class="pane pane-chat is-welcome"' in html, "a fresh page must start centered"
    greeting = html[html.index('id="welcome-hint"') : html.index("</p>", html.index('id="welcome-hint"'))]
    assert "你好，我是 MindCluster 运维助手" in greeting, "the greeting line is missing"
    # the centered state is a styled state, and sending/new-session must switch it on and off
    assert ".pane-chat.is-welcome .welcome-hint" in css, "the greeting needs its centered style"
    assert ".pane-chat.is-welcome .input-bar" in css, "the centered composer needs its style"
    assert 'classList.remove("is-welcome")' in js, "sending must return the composer to the bottom"
    assert 'classList.add("is-welcome")' in js, "new session must restore the centered state"
    # the greeting shares the centering with the card and aligns to the card's inner text edge
    start = css.index(".pane-chat.is-welcome .welcome-hint")
    hint_block = css[start : css.index("}", start)]
    assert "margin: auto" in hint_block, "the greeting+card group must share the centering"
    assert "text-align: left" in hint_block, "the greeting must align to the card's text edge"


def test_the_history_section_is_a_dead_end():
    """The sidebar's history line is text only - nothing behind it, in page or script."""
    html = read("index.html")
    empty = html[html.index('id="history-empty"') : html.index("</p>")]
    for forbidden in ("onclick", "href=", "data-", "button"):
        assert forbidden not in empty, "the empty state grew an interaction: " + forbidden
    for name in ("app.js", "views/chat.js", "views/toolcard.js", "sse.js"):
        assert "history-empty" not in read(name), name + " wires the empty history state"


def test_no_tool_name_or_guidance_text_on_the_page():
    """The page still knows no tool, and the guidance copy is still the server's."""
    html = read("index.html")
    for tool in TOOL_NAMES:
        assert tool not in html, "index.html learned the tool name " + tool


def test_style_sheet_hides_the_sidebar_on_narrow_screens():
    """Narrow screens drop the sidebar entirely (no drawer, no hamburger)."""
    css = read("style.css")
    media_query = f"@media (max-width: {NARROW_SCREEN_MAX_WIDTH}px)"
    assert media_query in css, "the narrow-screen breakpoint is missing"
    narrow = css[css.index(media_query) :]
    assert ".sidebar { display: none; }" in narrow, "the narrow-screen rule must hide the sidebar"
    for drawer in ("drawer", "hamburger", "overlay"):
        assert drawer not in css, "a drawer interaction crept in: " + drawer


def test_the_page_loads_nothing_from_the_network():
    """Zero build, zero CDN - every asset is local, so the offline image still renders."""
    html, css = read("index.html"), read("style.css")
    for name, text, extra in (
        ("index.html", html, ()),
        ("style.css", css, ("@import", "url(")),
    ):
        for scheme in ("http://", "https://", "//cdn", "fonts.googleapis", *extra):
            assert scheme not in text, name + " reaches the network: " + scheme
    assert "tailwind" not in html.lower(), "the reference project's Tailwind must not come along"


def test_sse_js_covers_the_six_events_and_posts():
    source = read("sse.js")
    assert "/chat/stream" in source
    assert "POST" in source
    handled = (
        protocol.EVENT_START,
        protocol.EVENT_TOKEN,
        protocol.EVENT_TOOL_CALL,
        protocol.EVENT_TOOL_RESULT,
        protocol.EVENT_ERROR,
        protocol.EVENT_DONE,
    )
    for name in handled:
        assert '"' + name + '"' in source, "sse.js does not handle event " + name
    # comment frames (the heartbeat) are skipped, never treated as events
    assert 'startsWith(":")' in source
    # no EventSource: the turn is a POST, which EventSource cannot carry
    assert "EventSource" not in source


def test_app_js_wires_the_views_and_never_writes_html():
    source = read("app.js")
    for module in ('"./sse.js"', '"./views/chat.js"'):
        assert module in source, "app.js does not use " + module
    for retired in ("localStorage", "sessionStorage", "listSessions", "getSession"):
        assert retired not in source, "app.js still references " + retired
    # the turn body is exactly the protocol pair
    assert "session_id" in source and "message" in source


def test_new_session_button_rolls_the_key():
    source = read("app.js")
    assert "newSessionId" in source
    assert "new-session-btn" in source
    # the page clears the flow instead of reloading history from the server
    assert "flow.textContent" in source


def test_no_page_script_writes_html():
    """页面 JS 一律 createElement + textContent：没有 innerHTML 就没有要消毒的东西。"""
    offenders = [n for n in PAGE_FILES if n.endswith(".js") and "innerHTML" in read(n)]
    assert not offenders, "page script writes HTML: " + str(offenders)


def test_toolcard_is_one_generic_card_with_name_args_output():
    source = read("views/toolcard.js")
    assert "createToolCard" in source
    for label in ("入参", "返回"):
        assert label in source, "toolcard.js misses the " + label + " block"
    # name + args + output only: no typed dispatch table, no matrix rendering
    for retired in ("RENDERERS", "renderers", "matrix", "tool-table", "presentation"):
        assert retired not in source, "toolcard.js still carries " + retired


def test_toolcard_reads_the_envelope_without_learning_tool_names():
    """卡片只依赖 name + output：把信封 JSON 美化展示，其余原样展示，不认识任何具体工具。"""
    source = read("views/toolcard.js")
    for symbol in ("parseEnvelope", "statusText", "statusTone", "formatMs"):
        assert "export function " + symbol in source, "toolcard.js misses " + symbol
    for piece in ("tool-tag", "tool-elapsed", "tool-index", "tool-copy", "tool-summary"):
        assert piece in source, "toolcard.js misses the " + piece + " piece"
    for tool in TOOL_NAMES:
        assert tool not in source, "toolcard.js learned the tool name " + tool


def test_chat_js_numbers_the_cards_within_one_turn():
    source = read("views/chat.js")
    assert "index: calls" in source, "chat.js must tell the card which call this is (第几次调用)"
    assert "calls += 1" in source


def test_chat_js_streams_tokens_and_shows_tool_cards():
    source = read("views/chat.js")
    for symbol in ("appendUserBubble", "createAssistantTurn", "appendNotice", "renderAnswer"):
        assert symbol in source, "chat.js misses " + symbol
    for action in ("appendToken", "addToolCall", "setToolResult"):
        assert action in source, "chat.js misses " + action


def test_the_no_model_hint_is_one_sentence_shared_with_the_server():
    """The placeholder and the error event carry the same single sentence, pinned to
    prompts.GUIDANCE_TEXT verbatim.
    """
    from agent_core.chat import prompts

    source = read("app.js")
    match = re.search(r'const NO_LLM_HINT = "([^"]+)"', source)
    assert match, "app.js must keep the shared no-model hint in one constant"
    assert match.group(1) == prompts.GUIDANCE_TEXT
    assert source.count("NO_LLM_HINT") >= 2, "the hint must back the placeholder"
    for retired in ("kubectl", "见下方引导", "确定性查询与诊断不受影响"):
        assert retired not in source, "app.js still carries retired guidance copy: " + retired


def test_the_page_reads_the_boot_state_the_server_inlines():
    """首屏状态取自页面内联的 boot 对象：服务端在 AGENT_BOOT 标记处注入，页面不再请求状态接口。"""
    assert "AGENT_BOOT" in read("index.html"), "index.html lost the boot marker the server replaces"
    source = read("app.js")
    assert "window.__AGENT_BOOT__" in source
    assert "llm_available" in source
    for retired in ("/api/v1/status", "./api.js", "fetch("):
        assert retired not in source, "app.js still depends on a status request: " + retired
