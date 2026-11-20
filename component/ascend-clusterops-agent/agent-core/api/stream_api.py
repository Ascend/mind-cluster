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

"""POST /chat/stream：一轮对话一条 SSE 流，任何降级路径都是 HTTP 200。

帧序 `start` -> (`token`* / `tool_call` -> `tool_result`)* -> `done`，无模型时 `error` + `done`；
会话记忆归图的 checkpointer，本层只管传输：每 `HEARTBEAT_SECONDS` 发一个注释帧防代理空闲断开。
"""

from __future__ import annotations

import asyncio
import logging
import re
from typing import Any, AsyncIterator, Callable, Dict, Optional

from fastapi import APIRouter, HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field, field_validator

from agent_core.chat import events, prompts
from agent_core.chat.graph import astream_chat_events, build_graph

logger = logging.getLogger(__name__)

SESSION_ID_MAX_CHARS = 128
SESSION_ID_PATTERN = re.compile(r"^[A-Za-z0-9_-]{1,%d}$" % SESSION_ID_MAX_CHARS)
CHAT_TEXT_MAX_BYTES = 16 * 1024  # 单条消息的接口上限，按 UTF-8 字节计
RUN_TIMEOUT_SECONDS = 1800.0  # 单轮上限；单次诊断本身就可能跑满 1200 秒

router = APIRouter(tags=["chat-stream"])

_llm_provider: Optional[Callable[[], Any]] = None
_tools: Any = None
_checkpointer: Any = None


def init_stream(llm_provider: Callable[[], Any], tools: Any, checkpointer: Any) -> None:
    """装配 lifespan 持有的三项：模型探针、工具集与记忆。"""
    global _llm_provider, _tools, _checkpointer
    _llm_provider = llm_provider
    _tools = tools
    _checkpointer = checkpointer


class ChatTurn(BaseModel):
    """The request body: a session key and one message."""

    session_id: str = Field(min_length=1, max_length=SESSION_ID_MAX_CHARS)
    message: str = Field(min_length=1, max_length=CHAT_TEXT_MAX_BYTES)

    @field_validator("session_id")
    @classmethod
    def _session_charset(cls, value: str) -> str:
        """Charset guard: the key reaches logs and the checkpointer, so keep it boring.

        fullmatch, not match: a lone '$' also matches before a trailing newline, and a
        session_id ending in a newline would split log lines and checkpointer keys.
        """
        if not SESSION_ID_PATTERN.fullmatch(value):
            raise ValueError("session_id must match [A-Za-z0-9_-] and be at most 128 characters")
        return value

    @field_validator("message")
    @classmethod
    def _message_budget(cls, value: str) -> str:
        """The cap is 16KB of UTF-8 bytes, not 16K characters."""
        size = len(value.encode("utf-8"))
        if size > CHAT_TEXT_MAX_BYTES:
            raise ValueError(f"message exceeds the {CHAT_TEXT_MAX_BYTES}-byte UTF-8 budget ({size} bytes)")
        return value


@router.post("/chat/stream")
async def post_chat_stream(turn: ChatTurn) -> StreamingResponse:
    """Answer one turn as an SSE stream; HTTP status is 200 on every degraded path."""
    if _llm_provider is None or _tools is None:
        raise HTTPException(status_code=503, detail="chat is not initialized")
    try:
        llm = _llm_provider()
        if llm is None:
            # 没有模型是正常状态，不是服务端错误。
            return _sse(_guide_stream(turn.session_id))
        # 每轮一个队列，两头各接一端：工具集往里写帧，翻译器从里读帧（显式交接，不用 ContextVar）。
        queue: asyncio.Queue = asyncio.Queue()
        graph = build_graph(llm, _tools.with_emitter(_queue_emitter(queue)), _checkpointer)
    except Exception:
        # 装配失败（密钥格式错、图校验失败等）和「没有模型」同样是逐轮状态：HTTP 200 + error 帧，
        # 绝不能是 500——前端的 ReadableStream 解析不了非 SSE 的错误页。
        logger.warning("chat stream build failed (session=%s)", _short_key(turn.session_id), exc_info=True)
        return _sse(_degraded_stream(turn.session_id))
    return _sse(_turn_stream(graph, turn.session_id, turn.message, queue))


def _queue_emitter(queue: asyncio.Queue):
    """本轮发布器：(事件, 载荷) -> 放进队列。"""

    async def _publish(event: str, payload: Dict[str, Any]) -> None:
        await queue.put((event, dict(payload)))

    return _publish


def _sse(agen: AsyncIterator[str]) -> StreamingResponse:
    """给 SSE 生成器套上流式响应需要的响应头。"""
    return StreamingResponse(
        agen,
        media_type="text/event-stream",
        headers={"Cache-Control": "no-cache", "X-Accel-Buffering": "no"},
    )


async def _guide_stream(session_id: str) -> AsyncIterator[str]:
    """没有模型时的路径：确认、说明、收尾，不创建任何会话线程。"""
    yield events.frame(events.EVENT_START, {"session_id": session_id})
    yield events.frame(events.EVENT_ERROR, {"message": prompts.GUIDANCE_TEXT})
    yield events.frame(events.EVENT_DONE)


async def _degraded_stream(session_id: str) -> AsyncIterator[str]:
    """装配失败的路径：与「没有模型」同形，start + error + done（接口承诺降级一律 HTTP 200）。"""
    yield events.frame(events.EVENT_START, {"session_id": session_id})
    yield events.frame(
        events.EVENT_ERROR,
        {"message": "本轮处理未能启动（模型或执行图构建失败），请稍后重试或检查配置。"},
    )
    yield events.frame(events.EVENT_DONE)


async def _forward_events(agent_events: AsyncIterator[Any], event_queue: asyncio.Queue) -> None:
    """把执行图产出的事件搬进队列，让主循环只在「等队列」上计时。

    生成器由本任务独占：主循环绝不 wait_for 生成器的 __anext__，那样超时会取消生成器并静默截断
    回答（例如一次 1200 秒的诊断）。正常读完放一个 None 哨兵，抛错则把异常本身放进队列，
    由主循环转成 error 帧。
    """
    try:
        async for item in agent_events:
            event_queue.put_nowait(item)
    except StopAsyncIteration:  # pragma: no cover - async-for 结束时不会抛它
        pass
    except BaseException as exc:  # noqa: BLE001 - 交给主循环，由它转成 error 帧
        event_queue.put_nowait(exc)
    finally:
        event_queue.put_nowait(None)  # 唯一哨兵：真事件永远不是 None


def _failure_frame(exc: BaseException, session_id: str) -> str:
    """本轮失败的 error 帧：超步上限与其它失败各自固定文案，原始异常只进服务端日志。"""
    if _is_recursion_error(exc):
        logger.warning("chat turn hit the superstep ceiling (session=%s)", _short_key(session_id))
        return events.frame(events.EVENT_ERROR, {"message": "本轮超出最大推理步数，请缩小问题范围后重试"})
    # 异常首行可能带内网地址与路径，所以帧里只给固定文案。
    logger.warning("chat turn failed (%s): %s", _short_error(exc), exc, exc_info=True)
    return events.frame(events.EVENT_ERROR, {"message": "模型服务不可用，请稍后重试或查看组件日志。"})


async def _stop_forwarding(forwarder: Optional[asyncio.Task]) -> None:
    """收尾：取消转发任务并等它退出（取消以 athrow 形式传到生成器，它的 finally 负责关流）。"""
    if forwarder is None or forwarder.done():
        return
    forwarder.cancel()
    try:
        await forwarder
    except asyncio.CancelledError:
        pass
    except BaseException:  # noqa: BLE001 - 收尾阶段不得再抛
        logger.warning("chat turn event forwarding raised while closing", exc_info=True)


def _timeout_frame(session_id: str) -> str:
    """整轮时钟到点：给一句可操作的提示，别让长任务把连接挂死。"""
    logger.warning("chat turn hit the %.0fs ceiling (session=%s)", RUN_TIMEOUT_SECONDS, _short_key(session_id))
    return events.frame(
        events.EVENT_ERROR,
        {"message": f"本轮处理超时（超过 {RUN_TIMEOUT_SECONDS:.0f} 秒），请缩小问题范围或改用 kubectl CLI"},
    )


def _queued_failure(item: BaseException) -> BaseException:
    """队列里的异常：转发任务转来的取消变成普通错误，其余原样上抛。

    取消这个协程本身不会进队列，所以真正的取消仍按取消向上传播。
    """
    if isinstance(item, asyncio.CancelledError):
        return RuntimeError("本轮被内部任务取消")
    return item


async def _turn_stream(
    graph: Any, session_id: str, message: str, queue: Optional[asyncio.Queue] = None
) -> AsyncIterator[str]:
    """把一轮图运行翻译成帧，任何失败都收敛成 error + done。

    心跳与整轮时限只计时「等队列」；生成器归转发任务独占，本循环只负责计时与发帧。
    """
    yield events.frame(events.EVENT_START, {"session_id": session_id})
    finished = False
    surfaced = False
    forwarder: Optional[asyncio.Task] = None
    event_queue: asyncio.Queue = asyncio.Queue()
    try:
        agent_events = astream_chat_events(graph, session_id, message, queue)
        forwarder = asyncio.ensure_future(_forward_events(agent_events, event_queue))
        loop = asyncio.get_running_loop()
        deadline = loop.time() + RUN_TIMEOUT_SECONDS
        while True:
            remaining = deadline - loop.time()
            if remaining <= 0:
                surfaced = True
                yield _timeout_frame(session_id)
                break
            try:
                item = await asyncio.wait_for(event_queue.get(), timeout=min(events.HEARTBEAT_SECONDS, remaining))
            except asyncio.TimeoutError:
                yield events.HEARTBEAT_FRAME
                continue
            if item is None:  # 转发任务把生成器读完了：这一轮自己结束
                finished = True
                break
            if isinstance(item, BaseException):  # 生成器抛错：作为本轮的 error 帧上报
                raise _queued_failure(item)
            yield events.frame(item[0], item[1])
    except Exception as exc:  # pylint: disable=broad-except
        surfaced = True  # 两种失败都会发 error 帧，所以不算「提前断开」
        yield _failure_frame(exc, session_id)
    finally:
        # 「提前断开」只留给真正走开的客户端：超时或异常结束的轮次已经上报过原因。
        if not finished and not surfaced:
            logger.info("chat turn closed early: session=%s", _short_key(session_id))
        await _stop_forwarding(forwarder)
    yield events.frame(events.EVENT_DONE)


def _is_recursion_error(exc: BaseException) -> bool:
    """超步上限触发时 langgraph 抛 GraphRecursionError。"""
    return exc.__class__.__name__ in ("GraphRecursionError", "RecursionError")


def _short_error(exc: BaseException) -> str:
    """压成一行可读文本，绝不上报 traceback（线上不暴露内部结构）。"""
    text = str(exc).strip().splitlines()
    return (text[0] if text else exc.__class__.__name__)[:200]


def _short_key(session_id: str) -> str:
    """日志里的会话号截短，完整 id 不进日志行。"""
    return session_id[:8] if len(session_id) > 8 else session_id
