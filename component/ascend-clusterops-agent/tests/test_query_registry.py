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

"""Registry boundaries: unique names, fixed policy, input/output validation."""

# pylint: disable=duplicate-code

from __future__ import annotations

import pytest
from pydantic import BaseModel, ConfigDict, ValidationError

from agent_core.query_schemas import QueryFailure
from agent_core.tools.registry import (
    DIAGNOSIS_POLICY,
    QUERY_POLICY,
    TOOL_REGISTRY,
    ToolDefinition,
    ToolInvocation,
    ToolRegistry,
)


class _Input(BaseModel):
    model_config = ConfigDict(extra="forbid", strict=True)
    value: int = 0


class _Output(BaseModel):
    value: int
    elapsed_ms: int = 0


def _handler(request: _Input, _ctx):
    return _Output(value=request.value)


def _definition(name="tool", read_only=True, policy=QUERY_POLICY, handler=_handler):
    return ToolDefinition(
        name=name,
        description="test tool",
        input_model=_Input,
        output_model=_Output,
        handler=handler,
        execution_policy=policy,
        read_only=read_only,
    )


def test_duplicate_registration_fails():
    with pytest.raises(ValueError):
        ToolRegistry([_definition("dup"), _definition("dup")])


def test_policy_must_match_read_only():
    with pytest.raises(ValueError):
        ToolRegistry([_definition("q", read_only=False, policy=QUERY_POLICY)])
    with pytest.raises(ValueError):
        ToolRegistry([_definition("d", read_only=True, policy=DIAGNOSIS_POLICY)])


def test_unknown_tool_is_tool_not_found():
    registry = ToolRegistry([_definition()])
    with pytest.raises(QueryFailure) as exc:
        registry.get("missing")
    assert exc.value.code == "TOOL_NOT_FOUND"
    assert exc.value.http_status == 404


def test_extra_field_is_rejected_before_handler():
    called = []
    registry = ToolRegistry(
        [_definition(handler=lambda request, ctx: called.append(request) or _Output(value=request.value))]
    )
    with pytest.raises(ValidationError):
        registry.execute("tool", {"unexpected": 1}, ToolInvocation())
    assert not called


def test_output_shape_mismatch_is_internal_error():
    registry = ToolRegistry([_definition(handler=lambda request, ctx: {"value": request.value})])
    with pytest.raises(QueryFailure) as exc:
        registry.execute("tool", {}, ToolInvocation())
    assert exc.value.code == "INTERNAL_ERROR"
    assert exc.value.http_status == 500


def test_execution_policy_builds_a_query_context_per_call():
    seen = []

    def handler(request, ctx):
        seen.append(ctx)
        return _Output(value=request.value)

    registry = ToolRegistry([_definition(handler=handler)])
    first = registry.execute("tool", {"value": 1}, ToolInvocation())
    second = registry.execute("tool", {"value": 2}, ToolInvocation())
    assert [item.value for item in (first, second)] == [1, 2]
    assert seen[0] is not seen[1]
    assert seen[0].request_id != seen[1].request_id


def test_register_after_construction_and_duplicate_guard():
    registry = ToolRegistry()
    definition = _definition("late")
    registry.register(definition)
    assert registry.get("late") is definition

    with pytest.raises(ValueError):
        registry.register(_definition("late"))

    replacement = _definition("late")
    registry.register(replacement, replace=True)
    assert registry.get("late") is replacement


def test_builtin_registry_exposes_five_tools():
    assert TOOL_REGISTRY.names() == [
        "cluster_resource",
        "diagnose",
        "faulty_device",
        "isolated_device",
        "task_status",
    ]


def test_tools_package_exports_the_shared_registry():
    from agent_core import tools as package

    assert package.TOOL_REGISTRY is TOOL_REGISTRY
    assert package.diagnose_cached.__module__ == "agent_core.tools.diagnosis"
