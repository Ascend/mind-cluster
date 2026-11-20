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

"""ReAct 编排：交给 langchain 的 `create_agent`，本模块只做边界与事件翻译。

记忆是进程级 MemorySaver（thread_id = session_id，重启即丢）；推理步数上限用 langgraph
自己的 recursion_limit。消息流翻译成 token 帧，工具帧经同一队列汇入。
"""

from __future__ import annotations

import asyncio
import logging
from typing import Any, AsyncIterator, Dict, List, Optional, Tuple

from langchain.agents import create_agent
from langchain_core.messages import AIMessageChunk, HumanMessage

from agent_core.chat.events import EVENT_TOKEN

logger = logging.getLogger(__name__)

# langgraph 按超步计数，25 是它自己的默认值，显式写出来让上限可见。
RECURSION_LIMIT = 25

Frame = Tuple[str, Dict[str, Any]]


def build_graph(llm: Any, agent_tools: Any, checkpointer: Any = None) -> Any:
    """编译 ReAct 图：提示词与工具对象都取自 agent_tools（模型每次调用时重新绑定工具）。"""
    return create_agent(
        llm,
        agent_tools.as_langchain_tools(),
        system_prompt=agent_tools.system_prompt(),
        checkpointer=checkpointer,
        name="clusterops-react",
    )


def _next_chunk(stream: Any) -> asyncio.Task:
    """任务包住的是迭代器的一步，所以必须显式调 __anext__，不能 async-for。"""
    return asyncio.ensure_future(stream.__anext__())  # pylint: disable=unnecessary-dunder-call


def _drain_ready(queue: asyncio.Queue) -> List[Any]:
    """取出队列里此刻已就绪的项，不等待。"""
    items: List[Any] = []
    while True:
        try:
            items.append(queue.get_nowait())
        except asyncio.QueueEmpty:
            return items


def _token_event(chunk_task: asyncio.Task) -> Optional[Frame]:
    """把一块模型消息翻成 token 帧；非文本块（工具调用块等）没有可播报的内容，返回 None。

    消息流结束时 result() 抛 StopAsyncIteration，由调用方据此收尾。
    """
    message_chunk, _metadata = chunk_task.result()
    usage = getattr(message_chunk, "usage_metadata", None) or {}
    if usage:
        logger.info("llm tokens: input=%s output=%s", usage.get("input_tokens"), usage.get("output_tokens"))
    text = getattr(message_chunk, "content", None) if isinstance(message_chunk, AIMessageChunk) else None
    if isinstance(text, str) and text:
        return (EVENT_TOKEN, {"text": text})
    return None


def _advance_stream(stream: Any, chunk_task: asyncio.Task) -> Tuple[Optional[Frame], Optional[asyncio.Task]]:
    """读完消息流的一步：返回这一步的 token 帧与下一步的任务；流结束时任务为 None。"""
    try:
        token = _token_event(chunk_task)
    except StopAsyncIteration:
        return None, None
    return token, _next_chunk(stream)


async def _cancel(task: Optional[asyncio.Task]) -> None:
    """取消并等任务退出：收尾阶段不得抛异常，也不得覆盖原始错误。"""
    if task is None or task.done():
        return
    task.cancel()
    try:
        await task
    except (asyncio.CancelledError, Exception):  # noqa: BLE001 - 收尾不得覆盖原始错误
        pass


async def _finish_events(event_task: asyncio.Task, queue: asyncio.Queue) -> List[Frame]:
    """消息流结束后收尾：取走最后一个工具帧，再把队列里剩的取完。

    流已结束，工具不可能再发帧，所以那个「等队列」的任务可以安全取消。
    """
    frames: List[Frame] = []
    if event_task.done() and not event_task.cancelled():
        frames.append(event_task.result())
    else:
        await _cancel(event_task)
    frames.extend(_drain_ready(queue))
    return frames


async def astream_chat_events(
    graph: Any,
    session_id: str,
    message: str,
    queue: asyncio.Queue,
) -> AsyncIterator[Frame]:
    """把一次图运行翻译成 token / tool_call / tool_result 事件。

    两个来源按到达顺序合并：消息流的文本成为 token，工具帧经 AgentTools.with_emitter 写进同一
    队列。start/done/error 归调用方；历史由 checkpointer 回填，这里只送本轮用户消息。
    """
    run_config: Dict[str, Any] = {"configurable": {"thread_id": session_id}, "recursion_limit": RECURSION_LIMIT}
    stream = graph.astream({"messages": [HumanMessage(content=message)]}, run_config, stream_mode="messages")
    chunk_task: Optional[asyncio.Task] = _next_chunk(stream)
    event_task: asyncio.Task = asyncio.ensure_future(queue.get())
    try:
        while chunk_task is not None:
            done, _pending = await asyncio.wait({chunk_task, event_task}, return_when=asyncio.FIRST_COMPLETED)
            # 先排空已就绪的工具帧：工具帧一定先于它之后的模型文本
            while event_task.done():
                yield event_task.result()
                event_task = asyncio.ensure_future(queue.get())
            for frame in _drain_ready(queue):
                yield frame
            if chunk_task in done:
                token, chunk_task = _advance_stream(stream, chunk_task)
                if token is not None:
                    yield token
            if chunk_task is None:
                for frame in await _finish_events(event_task, queue):
                    yield frame
                return
    finally:
        # 先等被取消的 __anext__ 退出再关流，否则 aclose() 会撞上正在运行的生成器
        await _cancel(chunk_task)
        await _cancel(event_task)
        try:
            await stream.aclose()  # 干净地停掉 langgraph 的后台流任务
        except (asyncio.CancelledError, RuntimeError):
            pass  # 被取消，或生成器仍被 __anext__ 持有：收尾不得抛异常
