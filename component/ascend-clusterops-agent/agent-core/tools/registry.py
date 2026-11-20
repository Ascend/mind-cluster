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

"""Single registry for the public business capabilities.

Four read-only query tools plus the existing diagnosis tool are registered here.
HTTP and CLI all resolve names/definitions from this registry, so nothing else
keeps a second name/parameter table. Query and diagnosis execution
keep their own policies: queries reuse ``QueryContext`` (scope/evidence ids),
diagnosis reuses the existing cached/mutex-protected chain.
"""

from __future__ import annotations

import time
import uuid
from dataclasses import dataclass
from typing import Any, Callable, Dict, List, Optional, Type

from agent_core.datasources.query_context import QueryContext
from agent_core.query_schemas import QueryFailure
from agent_core.tools.diagnosis_adapter import DiagnosisContext
from agent_core.tools.schemas import ToolDescriptor
from pydantic import BaseModel, ValidationError

QUERY_POLICY = "query"
DIAGNOSIS_POLICY = "diagnosis"


@dataclass(frozen=True)
class ToolDefinition:
    name: str
    description: str
    input_model: Type[BaseModel]
    output_model: Type[BaseModel]
    handler: Callable[[BaseModel, Any], BaseModel]
    execution_policy: str
    read_only: bool


@dataclass(frozen=True)
class ToolInvocation:
    """A non-query invocation: only carries the request id into the execution policy."""

    request_id: str = ""

    def __post_init__(self) -> None:
        if not self.request_id:
            object.__setattr__(self, "request_id", uuid.uuid4().hex)


# --------------------------------------------------------------------------- #
# Execution policies
# --------------------------------------------------------------------------- #
class QueryPolicy:
    """Queries reuse the existing QueryContext scope/evidence behaviour."""

    def create_context(self, invocation: Any, request: BaseModel) -> QueryContext:
        if isinstance(invocation, QueryContext):
            ctx = invocation
        else:
            ctx = QueryContext(request_id=getattr(invocation, "request_id", None) or uuid.uuid4().hex)
        ctx.request = request
        return ctx

    def before_execute(self, ctx: QueryContext, request: BaseModel) -> None:
        ctx.check_scope(request)


class DiagnosisPolicy:
    """Diagnosis keeps the existing timeout/mutex/cache chain."""

    def create_context(self, invocation: Any, request: BaseModel) -> DiagnosisContext:
        if isinstance(invocation, DiagnosisContext):
            return invocation
        return DiagnosisContext(request_id=getattr(invocation, "request_id", None) or uuid.uuid4().hex)

    def before_execute(self, ctx: DiagnosisContext, request: BaseModel) -> None:
        return None


# --------------------------------------------------------------------------- #
# Registry
# --------------------------------------------------------------------------- #
class ToolRegistry:
    def __init__(
        self,
        definitions: Optional[List[ToolDefinition]] = None,
        policies: Optional[Dict[str, Any]] = None,
    ):
        self._definitions: Dict[str, ToolDefinition] = {}
        self._policies = policies or {QUERY_POLICY: QueryPolicy(), DIAGNOSIS_POLICY: DiagnosisPolicy()}
        for definition in definitions or []:
            self.register(definition)

    def register(self, definition: ToolDefinition, *, replace: bool = False) -> None:
        """Add one capability. Duplicate names and policy mismatches fail fast."""
        if definition.execution_policy not in self._policies:
            raise ValueError(f"tool {definition.name!r} has no execution policy")
        if definition.execution_policy == QUERY_POLICY and not definition.read_only:
            raise ValueError(f"query tool {definition.name!r} must be read_only")
        if definition.execution_policy == DIAGNOSIS_POLICY and definition.read_only:
            raise ValueError(f"diagnosis tool {definition.name!r} must not be read_only")
        if definition.name in self._definitions and not replace:
            raise ValueError(f"duplicate tool name: {definition.name!r}")
        self._definitions[definition.name] = definition

    def get(self, name: str) -> ToolDefinition:
        definition = self._definitions.get(name)
        if definition is None:
            raise QueryFailure("TOOL_NOT_FOUND", f"unknown tool: {name!r}", http_status=404)
        return definition

    def names(self) -> List[str]:
        return sorted(self._definitions)

    def definitions(self) -> List[ToolDefinition]:
        return [self._definitions[name] for name in self.names()]

    def describe(self) -> List[ToolDescriptor]:
        return [self._descriptor(self._definitions[name]) for name in self.names()]

    @staticmethod
    def _descriptor(definition: ToolDefinition) -> ToolDescriptor:
        return ToolDescriptor(
            name=definition.name,
            description=definition.description,
            input_schema=definition.input_model.model_json_schema(),
        )

    def execute(
        self,
        name: str,
        arguments: Dict[str, Any],
        invocation: Any = None,
    ) -> BaseModel:
        """Validate and execute one registered tool.

        Every public entry point (HTTP, CLI, Agent) goes through here, so input
        validation is identical everywhere.
        """
        started = time.monotonic()
        definition = self.get(name)
        request = definition.input_model.model_validate(arguments)
        policy = self._policies[definition.execution_policy]
        ctx = policy.create_context(invocation, request)
        policy.before_execute(ctx, request)
        result = definition.handler(request, ctx)
        try:
            payload = result.model_dump(mode="python")
            output = definition.output_model.model_validate(payload)
        except (ValidationError, AttributeError, TypeError) as exc:
            # A handler returning the wrong shape is a server bug, not an invalid argument.
            raise QueryFailure(
                "INTERNAL_ERROR", "tool result does not match its output schema", http_status=500
            ) from exc
        if hasattr(output, "elapsed_ms"):
            output.elapsed_ms = int((time.monotonic() - started) * 1000)
        return output


TOOL_REGISTRY = ToolRegistry()
