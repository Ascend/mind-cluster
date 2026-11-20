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

"""POST /chat/stream 的 SSE 线格式：六个事件，多一个都不加。

`start` -> (`token`* / `tool_call` -> `tool_result`)* -> `done`，降级路径在 `done` 前多一个
`error`。会话状态在进程内存里，因此没有 run id、没有重放；本模块只管线上格式。
"""

from __future__ import annotations

import json
from typing import Any, Awaitable, Callable, Dict, Optional

EVENT_START = "start"  # 协议事件名，线上逐字使用
EVENT_TOKEN = "token"  # nosec B105 - SSE 事件名，不是凭据
EVENT_TOOL_CALL = "tool_call"
EVENT_TOOL_RESULT = "tool_result"
EVENT_DONE = "done"
EVENT_ERROR = "error"

# 单条工具结果的字符上限：整份工具信封原样给模型，超限截断并声明（截断不等于没有数据）。
TOOL_OUTPUT_MAX_CHARS = 9000
TRUNCATION_NOTE = (
    "\n（本段返回共 {total} 字符，超过 {limit} 字符上限已截断：缺失部分没有输出，不代表为空或不存在；"
    "需要完整数据请缩小范围（如指定 node）或改用 kubectl 查询。）"
)

# 心跳不是事件：SSE 注释帧，客户端会忽略，只用来避开代理的空闲超时。
HEARTBEAT_SECONDS = 15.0
HEARTBEAT_FRAME = ": keep-alive\n\n"

# 逐轮发布器：(事件名, 载荷) -> awaitable；None 表示这一轮没有监听者。
Emitter = Callable[[str, Dict[str, Any]], Awaitable[None]]


def clip_tool_output(text: str, limit: int = TOOL_OUTPUT_MAX_CHARS) -> str:
    """按上限裁剪一条工具结果；被裁时补一句截断声明，让模型知道数据不全。"""
    body = "" if text is None else str(text)
    if len(body) <= limit:
        return body
    note = TRUNCATION_NOTE.format(total=len(body), limit=limit)
    return body[: limit - len(note)] + note


def frame(event: str, payload: Optional[Dict[str, Any]] = None) -> str:
    """一个 SSE 帧：event 行 + 一行 JSON data。"""
    data = json.dumps(payload or {}, ensure_ascii=False, separators=(",", ":"))
    return f"event: {event}\ndata: {data}\n\n"
