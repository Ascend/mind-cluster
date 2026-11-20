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

"""五份技能正文必须描述工具真正实现的契约（过期技能比没有技能更糟）。

路由、字段引用与语义全部由注册表推导：allowed-tools 是注册名，正文引用的每条
`data.<path>` 必须能在 output_model 上走通，每份正文必须教会形状、原始信封与三段式回答格式。
"""

import re
import typing
from typing import Any, Dict

import pytest
import yaml

from agent_core.chat.skill_loader import SkillLoader
from agent_core.tools import TOOL_REGISTRY

SKILLS_ROOT = __file__.rsplit("/tests/", 1)[0] + "/agent-core/skills"
SKILL_DIRS = ("cluster-resource", "diagnose", "faulty-device", "isolated-device", "task-status")

# 技能目录名 -> 正文唯一允许路由到的工具（隔离工具注册名是下划线形式）。
EXPECTED_TOOL = {
    "cluster-resource": "cluster_resource",
    "diagnose": "diagnose",
    "faulty-device": "faulty_device",
    "isolated-device": "isolated_device",
    "task-status": "task_status",
}

# 工具永不产出的词表。分片拼接构造，避免本守卫自己成为残留词。
_BY_NODE = "by" + "_node"
_RANKTABLE = "ranktable" + "_status"
_RANK_COUNT = "rank" + "_count"
_SERVER_COUNT = "server" + "_count"
_ISOLATED_CARD = "isolated" + "_card"

# 工具永不产出的字段名与参数名。禁止事项一节可以点名（它的职责就是禁止），正文其余部分不行。
ABSENT_TOKENS = (
    "coverage",
    "sources",
    "is_inferred",
    "status_source",
    "live_pods",
    "node_faults",
    "switch_faults",
    "dpu_faults",
    "observed_faulty_npu_count",
    "TRANSPORT_ERROR",
    "TOOL_DISABLED",
    "SCOPE_FORBIDDEN",
    "DEADLINE_EXCEEDED",
    "warnings",
    "summary",
    "cards",
    "servers",
    "reasons",
    "state",
    "since",
    "summarize",
    _BY_NODE,
    _RANKTABLE,
    _RANK_COUNT,
    _SERVER_COUNT,
    _ISOLATED_CARD,
)
PROHIBITION_HEADING = "## 禁止事项"

# 每份正文必须教会的东西：查询质量语义、原始信封、截断声明与三段式回答格式。
COMMON_PHRASES = ("查询质量", "partial", "信封", "截断", "总结", "依据", "说明与建议")

# 各工具自己的形状要点（短语与字段名，不钉整句）。
REQUIRED_PHRASES = {
    "faulty-device": ("data.node", "devices[]", "faults[]", "switch", "dpu", "public_faults", "node_health"),
    "isolated-device": ("data.node", "devices[]", "fault_level", "last_separate_time", "由明细统计得出"),
    "cluster-resource": ("data.node", "cpu", "fault", "isolated", "由明细求和"),
    "task-status": ("job_type", "data.pods", "CrashLoopBackOff", "恒为数组"),
    "diagnose": ("final_text", "llm_error", "incomplete_nodes", "INTERNAL_ERROR"),
}

# 整份正文（含禁止事项）都不能出现的无效字段名。
FORBIDDEN_ANYWHERE = {
    "isolated-device": ("summary",),  # 隔离数据没有汇总计数
    "task-status": (_RANKTABLE, _RANK_COUNT, _SERVER_COUNT, "servers", "reasons"),
}

_DATA_PATH_RE = re.compile(
    r"data((?:\.[a-z_][a-z0-9_-]*(?:\[\]|\{\})?)+)"
)  # [] 与 {} 段都要穿透：{} 是节点名/动态键占位，叶子写错也必须红
_JSON_NODES_RE = re.compile(r'"nodes"\s*:')
_PARAM_ROW_RE = re.compile(r"^\| `([a-z_][a-z0-9_-]*)` \|", re.MULTILINE)
_NONE_TYPE = type(None)


def _load(skill_dir: str) -> Dict[str, Any]:
    """把一份 SKILL.md 拆成 frontmatter 与正文（禁止事项之前的操作用文本）。"""
    path = f"{SKILLS_ROOT}/{skill_dir}/SKILL.md"
    with open(path, encoding="utf-8") as handle:
        raw = handle.read()
    parts = raw.split("---", 2)
    assert len(parts) == 3, f"{skill_dir}: SKILL.md 必须以 frontmatter 开头"
    return {
        "meta": yaml.safe_load(parts[1]) or {},
        "body": parts[2],
        "operative": parts[2].split(PROHIBITION_HEADING, 1)[0],
        "path": path,
    }


def _walk(model: Any, path: str) -> bool:
    """沿 pydantic 模型走一条字段路径；False 表示这段路径不存在。

    路径段：`[]` 是列表项，`{}` 是按节点名索引的字典，`*` 是 extra=allow 的动态键。
    """
    current = model
    for raw in path.split("."):
        if current is None:
            return False
        if raw == "*":
            config = getattr(current, "model_config", None) or {}
            return config.get("extra") == "allow"
        is_list = raw.endswith("[]")
        is_dict = raw.endswith("{}")
        field = (getattr(current, "model_fields", None) or {}).get(raw[:-2] if is_list or is_dict else raw)
        if field is None:
            return False
        annotation = field.annotation
        if typing.get_origin(annotation) is typing.Union:  # Optional[X] -> X
            args = [item for item in typing.get_args(annotation) if item is not _NONE_TYPE]
            annotation = args[0] if len(args) == 1 else annotation
        if is_list or is_dict:
            if typing.get_origin(annotation) is not (list if is_list else dict):
                return False
            args = typing.get_args(annotation)
            index = 0 if is_list else 1
            annotation = args[index] if args else None
        current = annotation if hasattr(annotation, "model_fields") else None
    return True  # 每一段都在字段表里存在；标量叶子之后的继续下探已在上面的 None 分支判负


@pytest.mark.parametrize("skill_dir", SKILL_DIRS)
def test_frontmatter_routes_to_the_expected_tool(skill_dir):
    meta = _load(skill_dir)["meta"]
    assert meta["name"] == skill_dir, "frontmatter name 必须与目录名一致"
    assert meta["allowed-tools"] == [EXPECTED_TOOL[skill_dir]]
    assert set(meta["allowed-tools"]) <= set(TOOL_REGISTRY.names()) | {"load_skill"}


def test_the_citation_gate_penetrates_array_and_dict_items():
    """正则必须穿透 [] 与 {} 段：节点容器下的叶子字段要完整抽出。"""
    assert _DATA_PATH_RE.findall("读 data.node{}.devices[].faults[].fault_code 字段") == [
        ".node{}.devices[].faults[].fault_code"
    ]


def test_the_shipped_skills_all_parse_through_the_runtime_loader():
    """测试与运行时不得两套解析规则：SkillLoader 必须收下随包的每一份技能，且不多不少。"""
    assert sorted(SkillLoader(SKILLS_ROOT).names()) == sorted(SKILL_DIRS)


@pytest.mark.parametrize("skill_dir", SKILL_DIRS)
def test_every_cited_data_path_exists_in_the_output_model(skill_dir):
    """正文引用的每条 data 路径都必须在工具返回模型上真实存在。"""
    cited = sorted({match.group(0) for match in _DATA_PATH_RE.finditer(_load(skill_dir)["body"])})
    assert cited, f"{skill_dir}: 正文没有引用任何 data 字段"
    model = TOOL_REGISTRY.get(EXPECTED_TOOL[skill_dir]).output_model
    for path in cited:
        assert _walk(model, path), f"{skill_dir}: {path} 不存在于 {EXPECTED_TOOL[skill_dir]} 的返回模型"


@pytest.mark.parametrize("skill_dir", SKILL_DIRS)
def test_nonexistent_contract_vocabulary_is_confined_to_the_prohibition_section(skill_dir):
    operative = _load(skill_dir)["operative"]
    for token in ABSENT_TOKENS:
        assert token not in operative, f"{skill_dir}: {token} 不属于工具契约"
    assert not _JSON_NODES_RE.search(operative), f"{skill_dir}: 正文仍在展示 nodes 入参"


@pytest.mark.parametrize("skill_dir", SKILL_DIRS)
def test_documented_parameters_match_the_input_model(skill_dir):
    """正文参数表里写到的每个参数都必须存在于工具的入参模型。"""
    fields = set(TOOL_REGISTRY.get(EXPECTED_TOOL[skill_dir]).input_model.model_fields)
    documented = set(_PARAM_ROW_RE.findall(_load(skill_dir)["operative"]))
    assert documented, f"{skill_dir}: 参数表为空"
    assert documented == fields, f"{skill_dir}: 参数表与入参模型不一致"


@pytest.mark.parametrize("skill_dir", SKILL_DIRS)
def test_bodies_teach_the_envelope_and_the_answer_format(skill_dir):
    """简报删除后，正文要承担解读职责：原始信封、截断声明与三段式回答格式。"""
    body = _load(skill_dir)["body"]
    for phrase in (*COMMON_PHRASES, *REQUIRED_PHRASES[skill_dir]):
        assert phrase in body, f"{skill_dir}: 正文没有教会 {phrase!r}"


def test_nonexistent_vocabulary_stays_out_of_the_whole_body():
    """隔离与任务状态两份正文的无效字段连「禁止事项」都不能点名。"""
    for skill_dir, tokens in FORBIDDEN_ANYWHERE.items():
        body = _load(skill_dir)["body"]
        for token in tokens:
            assert token not in body, f"{skill_dir}: {token!r} 不得出现在正文任何位置"
