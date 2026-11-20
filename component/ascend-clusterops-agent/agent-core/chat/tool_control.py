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

"""工具集与调用控制：模型侧的工具对象，加上「一次工具调用怎么被管」的时限与诊断锁。

工具结果原样交给模型（工具层信封 JSON，超限截断并声明），不挑字段、不改写；
时限与诊断互斥也放在这里——它们描述的都是工具被调用的方式，分开两个文件反而看不出全貌。
"""

from __future__ import annotations

import asyncio
import json
import logging
import time
from typing import Any, Dict, List, Optional, Set, Tuple

from langchain_core.tools import BaseTool
from pydantic import BaseModel, Field

from agent_core.adapters.tool_adapter import INTERNAL_ERROR, TOOL_TIMEOUT, ToolOutcome
from agent_core.chat import events as protocol
from agent_core.chat import prompts
from agent_core.chat.events import EVENT_TOOL_CALL, EVENT_TOOL_RESULT, clip_tool_output

logger = logging.getLogger(__name__)

QUERY_CALL_TIMEOUT_SECONDS = 30.0
DIAGNOSE_CALL_TIMEOUT_SECONDS = 1200.0
DIAGNOSE_TOOL_NAME = "diagnose"

SKILL_TOOL_NAME = "load_skill"
SKILL_READ_TIMEOUT = 3.0
SKILL_TOOL_DESCRIPTION = "读取一个技能的使用指引正文。仅在需要某技能的详细说明时调用；未知名称会返回可用技能清单。"


def envelope_text(name: str, outcome: ToolOutcome) -> str:
    """工具层信封原样序列化；失败时给模型一个最小错误信封（与成功信封同字段）。"""
    if outcome.ok:
        payload: Dict[str, Any] = outcome.result if isinstance(outcome.result, dict) else {"data": outcome.result}
    else:
        payload = {"tool": name, "status": "error", "data": None, "errors": [outcome.error or {}]}
    return json.dumps(payload, ensure_ascii=False, default=str)


def _task_key(args: Dict[str, Any]) -> Tuple[str, str]:
    """诊断的锁粒度：一个任务（命名空间, 任务名），不同任务互不阻塞。"""
    return str(args.get("namespace") or "default"), str(args.get("job") or "")


def _diagnose_busy(key: Tuple[str, str]) -> ToolOutcome:
    """同一任务已有诊断在跑：如实回「正在诊断中」，别让模型空等、也别重复打一遍。"""
    namespace, job = key
    logger.info("diagnose already running: ns=%s job=%s", namespace, job)
    return ToolOutcome(
        ok=False,
        result=None,
        error={
            "code": INTERNAL_ERROR,
            "message": f"任务 {job}（命名空间 {namespace}）正在诊断中，本次未重复执行，请稍后再试",
        },
    )


class ToolControl:
    """调用时限 + 诊断锁。单进程单事件循环，集合本身即锁，无需额外 asyncio 原语。"""

    def __init__(self) -> None:
        self._diagnosing: Set[Tuple[str, str]] = set()

    async def call(self, adapter: Any, name: str, args: Dict[str, Any]) -> ToolOutcome:
        """在时限内调用一条能力；超时转成 TOOL_TIMEOUT，模型收到结果而不是空等。"""
        if name != DIAGNOSE_TOOL_NAME:
            return await self._run(adapter, name, args, QUERY_CALL_TIMEOUT_SECONDS)
        key = _task_key(args)
        if key in self._diagnosing:
            return _diagnose_busy(key)
        self._diagnosing.add(key)
        try:
            return await self._run(adapter, name, args, DIAGNOSE_CALL_TIMEOUT_SECONDS)
        finally:
            self._diagnosing.discard(key)

    async def _run(self, adapter: Any, name: str, args: Dict[str, Any], timeout: float) -> ToolOutcome:
        """底座调用放到线程里跑：同步的 registry 调用不占事件循环。"""
        try:
            return await asyncio.wait_for(asyncio.to_thread(adapter.invoke, name, args), timeout=timeout)
        except asyncio.TimeoutError:
            logger.warning("tool call timed out: name=%s timeout=%.0fs", name, timeout)
            return ToolOutcome(
                ok=False,
                result=None,
                error={"code": TOOL_TIMEOUT, "message": f"工具调用超过 {timeout:.0f} 秒上限，已中止本轮等待"},
            )


class LoadSkillArgs(BaseModel):
    """技能名入参。必须声明 args_schema，否则 langchain 推导不出 name 参数。"""

    name: str = Field(description="技能名（见系统提示词中的可用技能清单）")


class _AgentTool(BaseTool):
    """只走 async 路径、入参不二次校验、帧经 emitter 属性播出去（ContextVar 穿不过工具执行）。"""

    def _run(self, *args: Any, **kwargs: Any) -> Any:
        raise NotImplementedError("工具只走异步路径")

    def _parse_input(self, tool_input: Any, tool_call_id: Optional[str]) -> Any:  # type: ignore[override]
        """参数直交注册表校验：模型写错要收到 INVALID_ARGUMENT，而不是 langchain 的报错文本。"""
        return dict(tool_input) if isinstance(tool_input, dict) else {}

    async def _announce(self, event: str, payload: Dict[str, Any]) -> None:
        emitter: Optional[protocol.Emitter] = getattr(self, "emitter", None)
        if emitter is not None:
            await emitter(event, payload)


class RegistryTool(_AgentTool):  # pylint: disable=abstract-method
    """一个注册项：播 tool_call -> 受控调用 -> 播 tool_result，模型与卡片看到同一份文本。"""

    name: str = ""
    description: str = ""
    args_schema: Any = None
    adapter: Any = None
    control: Any = None
    emitter: Any = None

    async def _arun(self, **kwargs: Any) -> str:  # pylint: disable=arguments-differ
        args = dict(kwargs)
        await self._announce(EVENT_TOOL_CALL, {"name": self.name, "args": args})
        started = time.monotonic()
        outcome: ToolOutcome = await self.control.call(self.adapter, self.name, args)
        output = clip_tool_output(envelope_text(self.name, outcome))
        logger.info(
            "tool call: name=%s ok=%s elapsed_ms=%d output_chars=%d",
            self.name,
            outcome.ok,
            int((time.monotonic() - started) * 1000),
            len(output),
        )
        await self._announce(EVENT_TOOL_RESULT, {"name": self.name, "output": output})
        return output


class LoadSkillTool(_AgentTool):  # pylint: disable=abstract-method
    """技能正文按需读取：有界、只读、硬超时；读取失败降级成指引，不卡住本轮。"""

    name: str = SKILL_TOOL_NAME
    description: str = SKILL_TOOL_DESCRIPTION
    args_schema: Any = LoadSkillArgs
    skills: Any = None
    emitter: Any = None

    async def _arun(self, **kwargs: Any) -> str:  # pylint: disable=arguments-differ
        args = dict(kwargs)
        skill_name = str(args.get("name") or "")
        await self._announce(EVENT_TOOL_CALL, {"name": SKILL_TOOL_NAME, "args": args})
        try:
            output = str(await asyncio.wait_for(asyncio.to_thread(self.skills.read, skill_name), SKILL_READ_TIMEOUT))
        except asyncio.TimeoutError:
            logger.warning("skill read timed out: name=%s", skill_name)
            output = f"读取技能正文超时（{SKILL_READ_TIMEOUT:.0f} 秒上限），请基于已有信息回答。"
        except OSError:
            logger.warning("skill read failed: name=%s", skill_name, exc_info=True)
            output = "读取技能正文失败，请基于已有信息回答。"
        await self._announce(EVENT_TOOL_RESULT, {"name": SKILL_TOOL_NAME, "output": clip_tool_output(output)})
        return output


class AgentTools:
    """一次会话的工具集：能力工具 + 技能工具，系统提示词与工具清单同源生成。"""

    def __init__(self, adapter: Any, control: ToolControl, skills: Any = None, emitter: Any = None) -> None:
        self._adapter = adapter
        self._control = control
        self._skills = skills
        self._emitter = emitter

    def with_emitter(self, emitter: Any) -> "AgentTools":
        """逐轮用的副本：帧经该 emitter 播出（进程级实例没有 emitter）。"""
        return AgentTools(self._adapter, self._control, self._skills, emitter)

    def system_prompt(self) -> str:
        """身份 + 工具纪律 + 能力清单 + 技能清单。"""
        skill_lines = list(self._skills.prompt_lines()) if self._skills is not None else None
        return prompts.system_prompt(self._adapter.list_tools(), skill_lines)

    def _registry_tool(self, descriptor: Dict[str, Any]) -> RegistryTool:
        """一条能力对应的模型侧工具对象。"""
        name = str(descriptor.get("name") or "")
        return RegistryTool(
            name=name,
            description=str(descriptor.get("description") or name),
            args_schema=self._adapter.get_input_model(name),
            adapter=self._adapter,
            control=self._control,
            emitter=self._emitter,
        )

    def as_langchain_tools(self) -> List[BaseTool]:
        """五个能力的工具对象 + 技能工具（create_agent 直接绑定）。"""
        tools: List[BaseTool] = [self._registry_tool(item) for item in self._adapter.list_tools()]
        if self._skills is not None:
            tools.append(LoadSkillTool(skills=self._skills, emitter=self._emitter))
        return tools
