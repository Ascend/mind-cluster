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

"""工具层适配：agent-core 侧唯一接触 `agent_core.tools` 内部的文件。

发现、入参/返回模型、调用与错误收敛都在这里。异常不向编排层抛 traceback；
工具层缺失（未合入或坏包）时清单为空、调用答 TOOL_UNAVAILABLE，启动照常。
"""

from __future__ import annotations

import copy
import logging
import uuid
from dataclasses import dataclass
from typing import Any, Dict, List, Optional

from pydantic import ValidationError

logger = logging.getLogger(__name__)

# 信封里的三个失败码 + 框架侧两个：调用超时、工具层不可用。
INVALID_ARGUMENT = "INVALID_ARGUMENT"
TOOL_NOT_FOUND = "TOOL_NOT_FOUND"
INTERNAL_ERROR = "INTERNAL_ERROR"
TOOL_TIMEOUT = "TOOL_TIMEOUT"
TOOL_UNAVAILABLE = "TOOL_UNAVAILABLE"

# errors[] 有多条时谁当头条：按此顺序取第一个出现的码。
_ERROR_PRIORITY = (INVALID_ARGUMENT, TOOL_NOT_FOUND, INTERNAL_ERROR, TOOL_TIMEOUT, TOOL_UNAVAILABLE)
_MESSAGE_LIMIT = 300  # 一条错误消息的字符上限，更长的文本对模型是噪声
_MAX_VALIDATION_ERRORS = 3
_REQUEST_ID_CHARS = 12
_DIAGNOSE = "diagnose"
_SKIP_SCHEMA_KEYS = ("title", "$defs")


def _error_entry(code: str, message: Any) -> Dict[str, str]:
    """一条错误只带 {code, message}：消息按上限截断，缺消息时退回错误码本身。"""
    return {"code": code, "message": str(message or code)[:_MESSAGE_LIMIT]}


@dataclass(frozen=True)
class ToolOutcome:
    """一次调用的结果：成功携带工具信封，失败携带 {code, message}。"""

    ok: bool
    result: Optional[Dict[str, Any]]
    error: Optional[Dict[str, str]]

    @classmethod
    def failure(cls, code: str, message: Any, result: Optional[Dict[str, Any]] = None) -> "ToolOutcome":
        """构造一条失败结果；信封（result）可留作排查，判断失败只看 ok 与 error。"""
        return cls(ok=False, result=result, error=_error_entry(code, message))


@dataclass(frozen=True)
class _Invocation:
    """调用凭据：工具层只从中读 request_id，不依赖它的具体类型。"""

    request_id: str


def primary_error(entries: List[Any]) -> Optional[Dict[str, str]]:
    """按 _ERROR_PRIORITY 挑头条错误：码都不在表里就取第一条，条目缺码按 INTERNAL_ERROR。

    返回值恒为规整后的 {code, message}；信封里的完整 errors[] 不受影响。
    """
    normalized: List[Dict[str, str]] = []
    for entry in entries:
        if isinstance(entry, dict):
            code = str(entry.get("code") or INTERNAL_ERROR)
            normalized.append(_error_entry(code, entry.get("message") or entry.get("detail")))
        else:
            normalized.append(_error_entry(INTERNAL_ERROR, entry))
    for code in _ERROR_PRIORITY:
        for entry in normalized:  # 同一个码出现多次时取先出现的那条
            if entry["code"] == code:
                return entry
    return normalized[0] if normalized else None


def failure_from_exception(exc: BaseException, tool: str = "", request_id: str = "") -> ToolOutcome:
    """把上游异常翻译成失败结果（失败码 + 可读原因）；traceback 不越过这道缝。"""
    if isinstance(exc, (ImportError, ModuleNotFoundError)):
        return ToolOutcome.failure(TOOL_UNAVAILABLE, f"工具层不可用：{exc}")
    if isinstance(exc, ValidationError):
        problems = "; ".join(
            f"{'.'.join(str(part) for part in error.get('loc') or [])}: {error.get('msg')}"
            for error in exc.errors()[:_MAX_VALIDATION_ERRORS]
        )
        return ToolOutcome.failure(INVALID_ARGUMENT, f"参数校验失败：{problems}")
    code = getattr(exc, "code", None)
    if isinstance(code, str) and code in _ERROR_PRIORITY:
        # 工具层自己的 QueryFailure 之类已带机读码，直接用；非字符串 code 走下面的类型映射。
        return ToolOutcome.failure(code, getattr(exc, "message", None) or exc)
    if isinstance(exc, TimeoutError):
        # 查询超时算框架侧超时；诊断的失败面统一是 INTERNAL_ERROR（与工具层一致）。
        if tool == _DIAGNOSE:
            return ToolOutcome.failure(INTERNAL_ERROR, f"诊断调用超时：{exc}")
        return ToolOutcome.failure(TOOL_TIMEOUT, "调用超时，建议缩小查询范围后重试")
    if isinstance(exc, ConnectionError):
        return ToolOutcome.failure(INTERNAL_ERROR, f"上游连接不可达：{exc}")
    hint = f"（request_id={request_id}，详见 agent-core 日志）" if request_id else "（详见 agent-core 日志）"
    budget = max(0, _MESSAGE_LIMIT - len("内部错误：") - len(hint))
    return ToolOutcome.failure(INTERNAL_ERROR, "内部错误：" + f"{type(exc).__name__}: {exc}"[:budget] + hint)


def _simplify(node: Any, defs: Dict[str, Any], keys_are_names: bool = False) -> Any:
    """把一段 pydantic schema 化简成函数调用能直接吃的样子（递归处理子树）。

    三件事：$ref 换成 $defs 里的实际定义、丢掉 title 这类纯展示关键字、把 Optional[X] 的
    anyOf[X, null] 折成 X。keys_are_names 表示这一层的键是参数名而非 schema 关键字，原样保留。
    """
    if isinstance(node, list):
        return [_simplify(item, defs) for item in node]
    if not isinstance(node, dict):
        return node
    ref = node.get("$ref")
    if isinstance(ref, str) and ref.startswith("#/$defs/"):
        target = defs.get(ref[len("#/$defs/") :])
        if isinstance(target, dict):
            merged = dict(target)
            merged.update({key: value for key, value in node.items() if key != "$ref"})
            return _simplify(merged, defs, keys_are_names)
    if keys_are_names:
        return {key: _simplify(value, defs) for key, value in node.items()}
    cleaned = {
        key: _simplify(value, defs, key == "properties") for key, value in node.items() if key not in _SKIP_SCHEMA_KEYS
    }
    branches = cleaned.get("anyOf")
    if isinstance(branches, list):
        real = [item for item in branches if not (isinstance(item, dict) and item.get("type") == "null")]
        if len(real) == 1 and isinstance(real[0], dict):
            collapsed = dict(real[0])
            collapsed.update({key: value for key, value in cleaned.items() if key != "anyOf" and key not in collapsed})
            return collapsed
    return cleaned


def clean_schema(schema: Any) -> Dict[str, Any]:
    """函数调用友好的入参 schema：无 `$defs`、无 `anyOf[null]`。"""
    if not isinstance(schema, dict):
        return {"type": "object", "properties": {}}
    return _simplify(schema, dict(schema.get("$defs") or {}))


class ToolAdapter:
    """进程内适配 TOOL_REGISTRY：工具清单、入参/返回模型、调用与错误收敛。"""

    def __init__(self, registry: Any = None) -> None:
        self._registry = registry

    def _import_registry(self) -> Optional[Any]:
        """工具层本体；导入失败时保持 None（延迟到首次使用，缺层也能启动）。"""
        if self._registry is None:
            try:
                from agent_core.tools import TOOL_REGISTRY
            except ImportError:
                logger.debug("tool layer unavailable", exc_info=True)
                return None
            self._registry = TOOL_REGISTRY
        return self._registry

    def _require_registry(self) -> Any:
        """取注册表；工具层不在时抛 ImportError，由 failure_from_exception 翻成 TOOL_UNAVAILABLE。"""
        registry = self._import_registry()
        if registry is None:
            raise ImportError("tool layer is absent")
        return registry

    def _get_definition(self, name: str) -> Optional[Any]:
        """按名字取工具定义（ToolDefinition）；工具层缺失或名字不在册时返回 None。

        registry.get() 对未知名字会抛 TOOL_NOT_FOUND，而这里要的是「查不到就给 None」。
        """
        registry = self._import_registry()
        if registry is None:
            return None
        return next((definition for definition in registry.definitions() if definition.name == name), None)

    def list_tools(self) -> List[Dict[str, Any]]:
        """注册表的三个公开字段（name / description / input_schema），逐字镜像。"""
        registry = self._import_registry()
        if registry is None:
            return []
        return [item if isinstance(item, dict) else copy.deepcopy(item.model_dump()) for item in registry.describe()]

    def get_input_model(self, name: str) -> Optional[Any]:
        """注册的入参模型，用作 langchain 工具的 args_schema（不另立第二份契约）。"""
        definition = self._get_definition(name)
        return None if definition is None else definition.input_model

    def get_output_model(self, name: str) -> Optional[Any]:
        """注册的返回模型：技能正文引用的字段路径对它校验。"""
        definition = self._get_definition(name)
        return None if definition is None else definition.output_model

    def invoke(self, name: str, args: Dict[str, Any]) -> ToolOutcome:
        """执行一个工具；任何失败都落成结构化结果，绝不抛异常。"""
        request_id = uuid.uuid4().hex[:_REQUEST_ID_CHARS]
        try:
            raw = self._require_registry().execute(name, dict(args or {}), _Invocation(request_id))
            return _adapt_result(raw)
        except Exception as exc:  # noqa: BLE001 - 这道缝的职责就是把上游失败收敛成一条错误
            logger.exception("tool %s invocation failed (request_id=%s)", name, request_id)
            return failure_from_exception(exc, name, request_id)


def _adapt_result(raw: Any) -> ToolOutcome:
    """工具返回模型 -> ToolOutcome：status=error 即失败，失败原因取 errors[] 的头条。"""
    result = raw if isinstance(raw, dict) else raw.model_dump(mode="json")
    if result.get("status") != "error":
        return ToolOutcome(ok=True, result=result, error=None)
    headline = primary_error(result.get("errors") or [])
    if headline is None:
        return ToolOutcome.failure(INTERNAL_ERROR, "工具失败但没有错误条目", result)
    return ToolOutcome(ok=False, result=result, error=headline)
