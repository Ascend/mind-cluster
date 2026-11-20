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

"""工具层适配：清单镜像、入参/返回模型、调用收敛与错误码翻译。"""

import asyncio
import inspect
import sys
from typing import Any, List

import pytest
from pydantic import ValidationError

from agent_core.adapters import tool_adapter
from tests.fixtures.demo_adapter import demo_envelope

REGISTRY_NAMES = ["cluster_resource", "diagnose", "faulty_device", "isolated_device", "task_status"]

# 工具层未合入时（如 v5-ReAct-WebUI 分支）适配器清单为空：整文件跳过，而不是假装通过。
pytestmark = pytest.mark.skipif(not tool_adapter.ToolAdapter().list_tools(), reason="工具层未合入：无清单可校验")


@pytest.fixture(name="adapter")
def adapter_fixture() -> tool_adapter.ToolAdapter:
    return tool_adapter.ToolAdapter()


def _has_null_branch(node: Any) -> bool:
    """schema 子树里是否残留 null 分支（Optional 没折干净）。"""
    if isinstance(node, dict):
        if any(isinstance(item, dict) and item.get("type") == "null" for item in node.get("anyOf") or []):
            return True
        return any(_has_null_branch(value) for value in node.values())
    if isinstance(node, list):
        return any(_has_null_branch(item) for item in node)
    return False


def test_descriptors_are_the_registry_three_fields(adapter):
    descriptors = adapter.list_tools()
    assert [item["name"] for item in descriptors] == REGISTRY_NAMES
    for item in descriptors:
        assert set(item) == {"name", "description", "input_schema"}, "适配层不得增加第四个字段"
        assert item["input_schema"]["type"] == "object"


def test_descriptors_are_copies_not_shared_state(adapter):
    first = adapter.list_tools()[0]
    first["description"] = "mutated"
    first["input_schema"]["properties"]["injected"] = {"type": "string"}
    fresh = adapter.list_tools()[0]
    assert fresh["description"] != "mutated"
    assert "injected" not in fresh["input_schema"]["properties"]


def test_every_descriptor_resolves_to_its_models(adapter):
    for name in REGISTRY_NAMES:
        assert adapter.get_input_model(name) is not None, f"{name}: 入参模型缺失"
        assert hasattr(adapter.get_output_model(name), "model_fields"), f"{name}: 返回模型不是 pydantic 模型"


def test_node_parameter_is_an_array_of_names(adapter):
    by_name = {item["name"]: tool_adapter.clean_schema(item["input_schema"]) for item in adapter.list_tools()}
    for name in ("faulty_device", "isolated_device", "cluster_resource"):
        node = by_name[name]["properties"]["node"]
        assert node["type"] == "array", f"{name}.node 必须是节点名数组"
        assert node["items"]["type"] == "string"


def test_clean_schema_keeps_a_parameter_named_title():
    """回归：properties 的键是参数名，名为 title 的入参不得被关键字过滤吞掉。"""
    schema = {
        "type": "object",
        "title": "查询",
        "properties": {"title": {"type": "string", "title": "节点名"}, "node": {"type": "string"}},
    }
    cleaned = tool_adapter.clean_schema(schema)
    assert set(cleaned["properties"]) == {"title", "node"}
    assert cleaned["properties"]["title"] == {"type": "string"}
    assert "title" not in cleaned


def test_clean_schema_handles_refs_and_nested_optionals():
    raw = {
        "type": "object",
        "properties": {
            "inner": {"anyOf": [{"$ref": "#/$defs/Inner"}, {"type": "null"}]},
            "plain": {"$ref": "#/$defs/Inner"},
        },
        "$defs": {"Inner": {"type": "object", "title": "Inner", "properties": {"x": {"type": "string"}}}},
    }
    cleaned = tool_adapter.clean_schema(raw)
    assert "$defs" not in cleaned
    assert cleaned["properties"]["inner"]["type"] == "object"
    assert "title" not in cleaned["properties"]["inner"]
    assert cleaned["properties"]["plain"]["type"] == "object"


def test_registry_schemas_clean_up_without_null_branches(adapter):
    """随包入参模型清洗后必须能直接给函数调用：无 $defs/$ref、Optional 已折叠。"""
    for descriptor in adapter.list_tools():
        parameters = tool_adapter.clean_schema(descriptor["input_schema"])
        assert "$defs" not in parameters and "$ref" not in repr(parameters)
        assert not _has_null_branch(parameters), f"{descriptor['name']}: Optional 未折叠"


class _FakeRegistry:
    """注册表替身：只实现适配层用到的四个方法。"""

    def __init__(self, result: Any = None, raises: BaseException = None) -> None:
        self.result = result
        self.raises = raises
        self.seen: List[tuple] = []

    def describe(self):
        return _FakeRegistry.descriptors()

    def definitions(self):
        return _FakeRegistry.descriptors()

    @staticmethod
    def descriptors():
        from types import SimpleNamespace

        return [SimpleNamespace(name=name, input_schema={}, description=name) for name in REGISTRY_NAMES]

    def execute(self, name, args, invocation):
        self.seen.append((name, args, invocation))
        if self.raises is not None:
            raise self.raises
        return self.result


def test_ok_envelope_passes_through():
    envelope = demo_envelope("task_status")
    fake = _FakeRegistry(result=envelope)
    outcome = tool_adapter.ToolAdapter(registry=fake).invoke("task_status", {"job": "train-stub"})
    assert outcome.ok is True and outcome.error is None
    assert outcome.result["data"]["job_name"] == "train-stub"
    name, args, invocation = fake.seen[0]
    assert name == "task_status" and args == {"job": "train-stub"}
    assert getattr(invocation, "request_id", ""), "每次调用都要带 request_id"


def test_error_status_becomes_a_structured_error():
    envelope = demo_envelope(
        "isolated_device",
        status="error",
        errors=[{"code": "INVALID_ARGUMENT", "message": "节点名不合法"}],
    )
    outcome = tool_adapter.ToolAdapter(registry=_FakeRegistry(result=envelope)).invoke("isolated_device", {})
    assert outcome.ok is False
    assert outcome.error == {"code": "INVALID_ARGUMENT", "message": "节点名不合法"}
    assert outcome.result["status"] == "error", "失败信封仍要留给调用方（卡片要能看到）"


def test_primary_error_prefers_the_most_actionable_code():
    entries = [
        {"code": "INTERNAL_ERROR", "message": "boom"},
        {"code": "TOOL_NOT_FOUND", "message": "没有这个能力"},
        {"code": "TOOL_TIMEOUT", "message": "调用超时"},
    ]
    assert tool_adapter.primary_error(entries)["code"] == "TOOL_NOT_FOUND"
    assert tool_adapter.primary_error([]) is None
    assert tool_adapter.primary_error([{"message": "no code"}])["code"] == tool_adapter.INTERNAL_ERROR


def test_unknown_tool_is_tool_not_found(adapter):
    outcome = adapter.invoke("no_such_tool", {})
    assert outcome.ok is False
    assert outcome.error["code"] == tool_adapter.TOOL_NOT_FOUND


@pytest.mark.parametrize("args", [{"job": ""}, {}], ids=["empty-value", "missing-key"])
def test_bad_arguments_are_invalid_argument(adapter, args):
    outcome = adapter.invoke("task_status", args)
    assert outcome.ok is False
    assert outcome.error["code"] == tool_adapter.INVALID_ARGUMENT
    assert outcome.error["message"]


def test_unexpected_exception_becomes_internal_error():
    outcome = tool_adapter.ToolAdapter(registry=_FakeRegistry(raises=RuntimeError("kaboom"))).invoke(
        "task_status", {"job": "x"}
    )
    assert outcome.ok is False
    assert outcome.error["code"] == tool_adapter.INTERNAL_ERROR
    assert "kaboom" in outcome.error["message"]
    assert "request_id=" in outcome.error["message"], "内部错误必须带回日志追溯句柄"


def test_a_result_that_is_not_an_envelope_is_internal_error():
    outcome = tool_adapter.ToolAdapter(registry=_FakeRegistry(result=["not", "a", "mapping"])).invoke(
        "task_status", {"job": "x"}
    )
    assert outcome.ok is False
    assert outcome.error["code"] == tool_adapter.INTERNAL_ERROR


class _QueryFailure(Exception):
    def __init__(self, code: str, message: str) -> None:
        super().__init__(message)
        self.code = code
        self.message = message


@pytest.mark.parametrize(
    ("exc", "tool", "expected"),
    [
        (ImportError("no tools"), "task_status", tool_adapter.TOOL_UNAVAILABLE),
        (ModuleNotFoundError("no agent_core.tools"), "task_status", tool_adapter.TOOL_UNAVAILABLE),
        (TimeoutError(), "diagnose", tool_adapter.INTERNAL_ERROR),
        (TimeoutError(), "task_status", tool_adapter.TOOL_TIMEOUT),
        (ConnectionError("refused"), "task_status", tool_adapter.INTERNAL_ERROR),
        (_QueryFailure("INTERNAL_ERROR", "内部错误"), "diagnose", tool_adapter.INTERNAL_ERROR),
        (_QueryFailure("TOOL_NOT_FOUND", "没有这个能力"), "faulty_device", tool_adapter.TOOL_NOT_FOUND),
        (ValueError("plain"), "task_status", tool_adapter.INTERNAL_ERROR),
    ],
)
def test_error_translation_table(exc, tool, expected):
    outcome = tool_adapter.failure_from_exception(exc, tool, "req-9")
    assert outcome.ok is False
    assert outcome.error["code"] == expected


def test_validation_error_lists_the_offending_field():
    model = tool_adapter.ToolAdapter().get_input_model("task_status")
    with pytest.raises(ValidationError) as caught:
        model(job="ok", nope=1)
    error = tool_adapter.failure_from_exception(caught.value, "task_status").error
    assert error["code"] == tool_adapter.INVALID_ARGUMENT
    assert "nope" in error["message"]


def test_import_failure_is_reported_not_raised():
    """缺工具层的部署：模型查询降级为 None，调用答 TOOL_UNAVAILABLE，不抛异常。"""
    broken = tool_adapter.ToolAdapter(registry=None)
    saved = {name: sys.modules[name] for name in list(sys.modules) if name == "agent_core.tools"}
    try:
        sys.modules["agent_core.tools"] = None
        outcome = broken.invoke("task_status", {"job": "x"})
        assert broken.get_input_model("task_status") is None
        assert broken.get_output_model("task_status") is None
        assert broken.list_tools() == []
    finally:
        sys.modules.update(saved)
    assert outcome.ok is False
    assert outcome.error["code"] == tool_adapter.TOOL_UNAVAILABLE


def test_invoke_is_synchronous_so_it_can_run_in_a_worker_thread():
    assert not asyncio.iscoroutinefunction(tool_adapter.ToolAdapter.invoke)
    assert not inspect.iscoroutinefunction(tool_adapter.ToolAdapter.list_tools)
