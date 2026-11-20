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

"""SSE 线格式：六个事件、截断声明与心跳帧。"""

import json

from agent_core.chat import events


def test_event_names_are_the_protocol_six():
    actual = [events.EVENT_START, events.EVENT_TOKEN, events.EVENT_TOOL_CALL]
    actual += [events.EVENT_TOOL_RESULT, events.EVENT_DONE, events.EVENT_ERROR]
    assert actual == ["start", "token", "tool_call", "tool_result", "done", "error"]


def test_frame_encodes_one_event_and_one_json_line():
    frame = events.frame(events.EVENT_TOOL_CALL, {"name": "faulty_device", "args": {"node": "n1"}})
    assert frame.startswith("event: tool_call\n")
    assert frame.endswith("\n\n")
    data_line = frame.splitlines()[1]
    assert data_line.startswith("data: ")
    assert json.loads(data_line[len("data: ") :]) == {"name": "faulty_device", "args": {"node": "n1"}}


def test_clip_keeps_a_short_result_as_is():
    assert events.clip_tool_output("x" * 10) == "x" * 10
    assert events.TOOL_OUTPUT_MAX_CHARS == 9000


def test_clip_declares_the_truncation_to_the_model():
    """超限裁剪必须在文本里声明截断，且整段（含声明）不越过上限。"""
    clipped = events.clip_tool_output("y" * (events.TOOL_OUTPUT_MAX_CHARS * 2))
    assert len(clipped) == events.TOOL_OUTPUT_MAX_CHARS
    assert "已截断" in clipped
    assert "不代表为空" in clipped


def test_heartbeat_is_a_comment_frame_not_an_event():
    """保活帧必须保持可忽略：没有事件名、没有 data 行。"""
    assert events.HEARTBEAT_FRAME.startswith(":")
    assert "event:" not in events.HEARTBEAT_FRAME
