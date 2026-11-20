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

"""The offline scripted chat model shared by the graph and skill tests.

It lives here so both can import it without a collection cycle.
"""

from __future__ import annotations

import json
from typing import Any, List, Optional

from langchain_core.language_models.chat_models import BaseChatModel
from langchain_core.messages import AIMessageChunk
from langchain_core.outputs import ChatGeneration, ChatGenerationChunk, ChatResult
from pydantic import Field

# Each streamed chunk carries this many characters, so a reader sees real incremental frames.
CHUNK_CHARS = 3


class ScriptedModel(BaseChatModel):
    """One scripted response per stream, cut into CHUNK_CHARS-sized chunks."""

    responses: List[Any]
    index: int = 0
    calls: int = 0
    seen: List[List[Any]] = Field(default_factory=list)
    bound_tools: Optional[List[Any]] = None

    @property
    def _llm_type(self) -> str:
        return "scripted-skill-test"

    def bind_tools(self, tools, **kwargs):  # noqa: ARG002
        self.bound_tools = list(tools)
        return self

    def _generate(self, messages, stop=None, run_manager=None, **kwargs):  # noqa: ARG002
        self.calls += 1
        self.seen.append(list(messages))
        message = self.responses[min(self.index, len(self.responses) - 1)]
        self.index += 1
        return ChatResult(generations=[ChatGeneration(message=message)])

    async def _astream(self, messages, stop=None, run_manager=None, **kwargs):  # noqa: ARG002
        self.calls += 1
        self.seen.append(list(messages))
        message = self.responses[min(self.index, len(self.responses) - 1)]
        self.index += 1
        text = message.content if isinstance(message.content, str) else ""
        for i in range(0, len(text), CHUNK_CHARS):
            yield ChatGenerationChunk(message=AIMessageChunk(content=text[i : i + CHUNK_CHARS]))
        for call in message.tool_calls or []:
            yield ChatGenerationChunk(
                message=AIMessageChunk(
                    content="",
                    tool_call_chunks=[
                        {
                            "name": call["name"],
                            "args": json.dumps(call.get("args") or {}, ensure_ascii=False),
                            "id": call.get("id") or "call-1",
                            "type": "tool_call_chunk",
                        }
                    ],
                )
            )
