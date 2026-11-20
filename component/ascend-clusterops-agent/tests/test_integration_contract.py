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

"""工具注册表与本侧假设的一致性：工具侧一移动就红。

名字、参数、模型与错误码由工具层持有；读不到的是「本侧怎么假设它们」，所以写死在这里。
红了要人工复核受影响的消费方（提示词 / 卡片 / 技能正文），而不是放宽断言。
"""

import pytest

from agent_core.adapters import tool_adapter as adapter
from agent_core.query_schemas import (
    FaultData,
    IsolationData,
    JobData,
    ResourceData,
    ToolEnvelope,
)
from agent_core.tools import TOOL_REGISTRY
from tests.fixtures.demo_adapter import demo_envelope

# Optional[X] 的 __args__ 是 (X, NoneType)：过滤 NoneType 要按类比较，is not None 恒真。
_NONE_TYPE = type(None)

# 本侧认识的五个能力。
DOCUMENTED_TOOLS = ["cluster_resource", "diagnose", "faulty_device", "isolated_device", "task_status"]

# Required/optional input fields per tool, exactly as the prompt and the graph call them.
# "None" means the tool takes no argument; a listed name must exist on the input model.
DOCUMENTED_INPUTS = {
    "faulty_device": {"node"},
    "isolated_device": {"node"},
    "cluster_resource": {"node"},
    "task_status": {"job", "namespace"},
    "diagnose": {"job", "namespace", "refresh"},
}

# 每个工具信封必须携带的数据模型（卡片与技能正文的字段路径都挂在它上面）。
DOCUMENTED_DATA_MODELS = {
    "faulty_device": FaultData,
    "isolated_device": IsolationData,
    "cluster_resource": ResourceData,
    "task_status": JobData,
}

# 卡片与技能正文读取的信封字段：新增必填字段会打断所有消费方，改名会让正文里的路径静默失效。
DOCUMENTED_ENVELOPE_FIELDS = {"request_id", "tool", "status", "elapsed_ms", "data", "errors"}
DOCUMENTED_STATUSES = {"ok", "partial", "error"}

# One public code per failure kind: the three envelope codes plus the two framework-owned ones.
DOCUMENTED_ERROR_CODES = {
    adapter.INVALID_ARGUMENT,
    adapter.TOOL_NOT_FOUND,
    adapter.INTERNAL_ERROR,
    adapter.TOOL_UNAVAILABLE,
    adapter.TOOL_TIMEOUT,
}


def test_registry_registers_exactly_the_documented_tools():
    assert sorted(TOOL_REGISTRY.names()) == DOCUMENTED_TOOLS


@pytest.mark.parametrize("name", DOCUMENTED_TOOLS)
def test_input_fields_match_the_documented_parameters(name):
    definition = TOOL_REGISTRY.get(name)
    assert set(definition.input_model.model_fields) == DOCUMENTED_INPUTS[name]


@pytest.mark.parametrize("name", DOCUMENTED_TOOLS)
def test_output_model_carries_the_documented_data_model(name):
    definition = TOOL_REGISTRY.get(name)
    envelope_model = definition.output_model
    assert "data" in envelope_model.model_fields, f"{name}: output model is not an envelope"
    if name in DOCUMENTED_DATA_MODELS:
        annotation = envelope_model.model_fields["data"].annotation
        args = [item for item in getattr(annotation, "__args__", ()) if item is not _NONE_TYPE]
        assert DOCUMENTED_DATA_MODELS[name] in args, (
            f"{name}: data annotation drifted from {DOCUMENTED_DATA_MODELS[name].__name__}"
        )


def test_every_tool_uses_the_shared_envelope():
    for name in DOCUMENTED_TOOLS:
        model = TOOL_REGISTRY.get(name).output_model
        shared = set(ToolEnvelope.model_fields) & set(model.model_fields)
        assert shared == DOCUMENTED_ENVELOPE_FIELDS, f"{name}: envelope fields drifted ({sorted(shared)})"


def test_envelope_status_literal_keeps_three_states():
    annotation = ToolEnvelope.model_fields["status"].annotation
    values = set(getattr(annotation, "__args__", ()))
    assert values == DOCUMENTED_STATUSES


def test_every_documented_error_code_is_exported():
    for code in DOCUMENTED_ERROR_CODES:
        assert getattr(adapter, code) == code, f"the adapter lost the {code} constant"
    # The headline picker must know every code, otherwise a failure degrades to a generic text.
    assert DOCUMENTED_ERROR_CODES - {adapter.TOOL_UNAVAILABLE} <= set(adapter._ERROR_PRIORITY)


@pytest.mark.parametrize("name", DOCUMENTED_TOOLS)
def test_tool_descriptions_are_single_line(name):
    description = TOOL_REGISTRY.get(name).description
    assert "\n\n" not in description.strip(), f"{name}: description must not carry paragraph breaks"


@pytest.mark.parametrize("name", DOCUMENTED_TOOLS)
def test_tools_are_read_only_except_diagnose(name):
    definition = TOOL_REGISTRY.get(name)
    if name == "diagnose":
        assert definition.read_only is False, "diagnose writes a report and must not claim read-only"
    else:
        assert definition.read_only is True, f"{name}: a query tool must stay read-only"


@pytest.mark.parametrize("name", ["faulty_device", "isolated_device", "cluster_resource", "task_status", "diagnose"])
def test_demo_payloads_validate_against_the_tooling_models(name):
    """演示夹具必须是工具层真能返回的载荷。"""
    envelope = demo_envelope(name)
    validated = TOOL_REGISTRY.get(name).output_model.model_validate(envelope)
    assert set(validated.model_dump(mode="json")) == set(envelope)
    assert validated.model_dump(mode="json")["status"] == "ok"


def test_diagnose_data_field_set_is_an_allowlist_not_a_presence_guarantee():
    """DiagnosisData 是 extra=allow 且带序列化器：没有的键就是不存在，不是 null。"""
    model = TOOL_REGISTRY.get("diagnose").output_model
    minimal = {"request_id": "r", "tool": "diagnose", "status": "ok", "elapsed_ms": 1, "data": {}, "errors": []}
    dumped = model.model_validate(minimal).model_dump()
    assert "final_text" not in dumped and "llm_error" not in dumped
