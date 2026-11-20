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

"""系统提示词：能力清单按注册表描述渲染，纪律与三段式回答格式在这里钉住。

能力清单不是自由文本：其中一条用例把它对齐到 GET /tools 的真实目录，工具层一改就红。
"""

import re

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from agent_core import query_api
from agent_core.adapters.tool_adapter import ToolAdapter
from agent_core.chat import prompts
from agent_core.chat.tool_control import AgentTools
from agent_core.chat.tool_control import ToolControl


@pytest.fixture(name="descriptors")
def descriptors_fixture():
    return ToolAdapter().list_tools()


@pytest.fixture(name="catalog_client")
def catalog_client_fixture():
    app = FastAPI()
    app.include_router(query_api.router)
    return TestClient(app)


def test_prompt_lists_every_capability_with_its_parameters(descriptors):
    if not descriptors:
        pytest.skip("工具层未合入：能力清单为空")
    text = prompts.system_prompt(descriptors)
    for descriptor in descriptors:
        assert descriptor["name"] in text
        assert descriptor["description"] in text
    # 参数取自 input_schema，模型看到的参数名与校验器用的是同一份
    assert "job（字符串，必填）" in text
    assert "namespace（字符串，可选）" in text
    assert "node（字符串数组，可选）" in text


def test_prompt_states_the_discipline_rules(descriptors):
    """四条语义底线：只准引用返回值、status 是查询质量、null 是未知、截断不等于没有数据。"""
    text = prompts.system_prompt(descriptors)
    for phrase in (
        "只能依据工具返回的字段和数值作答",
        "外层 status 是「查询质量」",
        "null 一律读作「未知」",
        "被截断的部分没有输出，不得当成空或不存在",
        "data.node",
    ):
        assert phrase in text, f"纪律条目缺失: {phrase}"


def test_prompt_asks_for_a_three_part_answer(descriptors):
    """回答格式固定三段：总结 / 依据 / 说明与建议。"""
    text = prompts.system_prompt(descriptors)
    for part in ("总结", "依据", "说明与建议"):
        assert part in text, f"三段式回答缺少 {part} 段"
    assert "全程中文" in text and "不要中英混排" in text


def test_skill_listing_is_names_and_descriptions_only(descriptors):
    with_skill = prompts.system_prompt(descriptors, ["fault-device-triage: 用户问哪些设备有故障时使用"])
    assert "- fault-device-triage: 用户问哪些设备有故障时使用" in with_skill
    assert "load_skill(name)" in with_skill
    assert "load_skill(name)" not in prompts.system_prompt(descriptors)


PARAM_PATTERN = re.compile(r"([A-Za-z_][A-Za-z0-9_]*)（([^，]+)，(必填|可选)）")


def _rendered_capabilities(prompt: str) -> dict:
    """解析【可用能力清单】块，得到 {工具名: (描述, [(参数名, 形状, 必填/可选), ...])}。"""
    block = prompt.split("【可用能力清单】", 1)[1].split("【可用技能】", 1)[0]
    rendered: dict = {}
    current = None
    for raw in block.splitlines():
        line = raw.strip()
        if line.startswith("- "):
            name, _, description = line[2:].partition("：")
            current = name.strip()
            rendered[current] = (description.strip(), [])
        elif line.startswith("参数：") and current:
            rendered[current] = (rendered[current][0], PARAM_PATTERN.findall(line[3:]))
    return rendered


def test_system_prompt_matches_the_tool_catalog(catalog_client):
    """能力清单 == GET /tools：名字、顺序、参数与必填标记都对齐。"""
    adapter = ToolAdapter()
    rendered = _rendered_capabilities(AgentTools(adapter, ToolControl()).system_prompt())
    tools = catalog_client.get("/tools").json()["tools"]
    catalog = {tool["name"]: tool for tool in tools}

    assert list(rendered) == [tool["name"] for tool in tools]
    for name, (description, params) in rendered.items():
        schema = catalog[name]["input_schema"]
        properties = list(schema.get("properties") or {})
        required = [str(item) for item in (schema.get("required") or [])]
        assert description == catalog[name]["description"].strip()
        assert [param for param, _, _ in params] == properties
        assert [necessity for _, _, necessity in params] == [
            "必填" if param in required else "可选" for param in properties
        ]


def test_guidance_is_one_fixed_sentence():
    """无模型提示是唯一一句固定文案，页面占位符与 error 事件共用它。"""
    assert prompts.GUIDANCE_TEXT == "未配置大模型，对话不可用；请配置模型后使用"
