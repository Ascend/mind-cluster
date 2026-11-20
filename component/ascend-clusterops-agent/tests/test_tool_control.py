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

"""chat/tool_control.py：模型侧工具集（信封直传、技能工具）与调用控制（时限、诊断按任务锁）。"""

from __future__ import annotations

import asyncio
import json
import time

from agent_core.adapters.tool_adapter import INTERNAL_ERROR, TOOL_TIMEOUT
from agent_core.chat import tool_control as tc
from agent_core.chat import events
from tests.fixtures.demo_adapter import DemoAdapter


class FakeSkills:
    """技能宿主替身：清单行、技能名与一段有界正文。"""

    def __init__(self, names=("ok-skill",), body="技能正文"):
        self._names = list(names)
        self._body = body

    def names(self):
        return list(self._names)

    def prompt_lines(self):
        return [f"- {name}: 说明" for name in self._names]

    def read(self, name):
        if name not in self._names:
            return "未知技能：" + name
        return self._body


class Collector:
    """顶替本轮队列：收下工具播出的帧。"""

    def __init__(self):
        self.frames = []

    async def __call__(self, event, payload):
        self.frames.append((event, payload))


def _tools(skills=None, emitter=None, adapter=None):
    return tc.AgentTools(adapter or DemoAdapter(), tc.ToolControl(), skills=skills, emitter=emitter)


def _tool(tools, name):
    return next(tool for tool in tools.as_langchain_tools() if tool.name == name)


def test_tools_and_prompt_cover_registry_plus_skill_tool():
    adapter = DemoAdapter()
    tools = _tools(skills=FakeSkills(), adapter=adapter)
    assert [tool.name for tool in tools.as_langchain_tools()] == adapter.names() + [tc.SKILL_TOOL_NAME]
    assert "ok-skill" in tools.system_prompt()


def test_no_skill_host_means_no_skill_tool():
    tools = _tools()
    assert [tool.name for tool in tools.as_langchain_tools()] == DemoAdapter().names()
    assert "load_skill(name)" not in tools.system_prompt()


def test_skill_tool_declares_the_name_parameter():
    """一个必填入参，否则模型没法点名技能。"""
    skill_tool = _tool(_tools(skills=FakeSkills()), tc.SKILL_TOOL_NAME)
    schema = skill_tool.args_schema.model_json_schema()
    assert list(schema["properties"]) == ["name"]
    assert schema["required"] == ["name"]


def test_registry_tool_hands_the_model_the_raw_envelope():
    """能力结果就是工具层信封本身：不挑字段、没有简报头，帧里的文本与模型拿到的一致。"""
    sink = Collector()
    adapter = DemoAdapter()
    tool = _tool(_tools(emitter=sink, adapter=adapter), "faulty_device")
    output = asyncio.run(tool._arun(node=["node-1"]))
    assert json.loads(output)["data"] == adapter.invoke("faulty_device", {"node": ["node-1"]}).result["data"]
    assert not output.startswith("[faulty_device]"), "简报头已废弃"
    assert [name for name, _payload in sink.frames] == [events.EVENT_TOOL_CALL, events.EVENT_TOOL_RESULT]
    assert sink.frames[0][1] == {"name": "faulty_device", "args": {"node": ["node-1"]}}
    assert sink.frames[1][1] == {"name": "faulty_device", "output": output}


def test_with_emitter_is_what_wires_the_frames():
    """进程级工具集没有 emitter；逐轮副本才有。"""
    plain = _tools()
    sink = Collector()
    scoped = plain.with_emitter(sink)
    assert scoped is not plain
    assert _tool(scoped, "faulty_device").emitter is sink
    assert asyncio.run(_tool(plain, "faulty_device")._arun(node=["n1"]))
    assert not sink.frames


def test_load_skill_returns_the_body_and_announces_frames():
    sink = Collector()
    tool = _tool(_tools(skills=FakeSkills(body="先看日志"), emitter=sink), tc.SKILL_TOOL_NAME)
    assert asyncio.run(tool._arun(name="ok-skill")) == "先看日志"
    assert sink.frames == [
        (events.EVENT_TOOL_CALL, {"name": tc.SKILL_TOOL_NAME, "args": {"name": "ok-skill"}}),
        (events.EVENT_TOOL_RESULT, {"name": tc.SKILL_TOOL_NAME, "output": "先看日志"}),
    ]


def test_load_skill_read_failure_degrades_instead_of_escaping():
    class BrokenSkills(FakeSkills):
        def read(self, name):
            raise OSError("volume is gone")

    sink = Collector()
    tool = _tool(_tools(skills=BrokenSkills(), emitter=sink), tc.SKILL_TOOL_NAME)
    assert asyncio.run(tool._arun(name="ok-skill")) == "读取技能正文失败，请基于已有信息回答。"
    assert sink.frames[1][0] == events.EVENT_TOOL_RESULT  # 帧不缺席，卡片不会停在「调用中」


def test_query_timeout_becomes_a_structured_error(monkeypatch):
    class SlowAdapter(DemoAdapter):
        def invoke(self, name, args):
            time.sleep(0.2)
            return super().invoke(name, args)

    monkeypatch.setattr(tc, "QUERY_CALL_TIMEOUT_SECONDS", 0.01)
    outcome = asyncio.run(tc.ToolControl().call(SlowAdapter(), "faulty_device", {}))
    assert outcome.ok is False
    assert outcome.error["code"] == TOOL_TIMEOUT


def test_same_task_diagnose_is_rejected_while_running():
    """诊断按任务串行：同一任务重复请求直接回提示，不同任务照常并行。"""

    class SlowDiagnose(DemoAdapter):
        def invoke(self, name, args):
            time.sleep(0.1)
            return super().invoke(name, args)

    control = tc.ToolControl()
    adapter = SlowDiagnose()

    async def scenario():
        running = asyncio.create_task(control.call(adapter, "diagnose", {"job": "j1", "namespace": "ns"}))
        await asyncio.sleep(0.02)
        blocked = await control.call(adapter, "diagnose", {"job": "j1", "namespace": "ns"})
        other = await control.call(adapter, "diagnose", {"job": "j2", "namespace": "ns"})
        return await running, blocked, other

    running, blocked, other = asyncio.run(scenario())
    assert running.ok is True and other.ok is True
    assert blocked.ok is False
    assert blocked.error["code"] == INTERNAL_ERROR
    assert "正在诊断中" in blocked.error["message"]
    assert [name for name, _args in adapter.calls].count("diagnose") == 2  # 被拒的那次没进底座
