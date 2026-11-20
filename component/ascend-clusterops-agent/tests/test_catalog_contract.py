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

"""Catalog contract: one request_id, five sorted entries, exactly three descriptor fields.

The third field (``input_schema``) must be generated from the registered input model, so
the catalog can never drift from the validation the server actually applies.
"""

# pylint: disable=redefined-outer-name  # pytest fixtures are injected by parameter name

from __future__ import annotations

from pathlib import Path

import pytest
from agent_core import query_api
from agent_core.tools.registry import TOOL_REGISTRY
from fastapi import FastAPI
from fastapi.testclient import TestClient

EXPECTED_TOOLS = [
    "cluster_resource",
    "diagnose",
    "faulty_device",
    "isolated_device",
    "task_status",
]

# agent-core 目录同时是 agent_core 包的根目录 (pyproject 的 package-dir 映射),
# 这里只跳过打包器文件、构建中间产物与构建脚本的临时备份。
PACKAGE_DIR = Path(__file__).resolve().parents[1] / "agent-core"
PACKAGE_PACKAGING_FILES = {"pyproject.toml"}
IGNORED_PARTS = {"build", "__pycache__", ".pytest_cache"}


@pytest.fixture
def client():
    app = FastAPI()
    app.include_router(query_api.router)
    return TestClient(app)


def test_catalog_has_one_request_id_and_five_entries(client):
    response = client.get("/tools")
    assert response.status_code == 200

    body = response.json()
    assert set(body) == {"request_id", "tools"}
    assert isinstance(body["request_id"], str) and body["request_id"]
    assert [tool["name"] for tool in body["tools"]] == EXPECTED_TOOLS


def test_every_descriptor_has_exactly_the_three_public_fields(client):
    for tool in client.get("/tools").json()["tools"]:
        assert list(tool) == ["name", "description", "input_schema"]
        assert isinstance(tool["name"], str) and tool["name"]
        assert isinstance(tool["description"], str) and tool["description"].strip()
        assert tool["input_schema"]["type"] == "object"


def test_input_schema_is_the_registered_model_schema(client):
    catalog = {tool["name"]: tool for tool in client.get("/tools").json()["tools"]}
    for name in TOOL_REGISTRY.names():
        definition = TOOL_REGISTRY.get(name)
        assert catalog[name]["input_schema"] == definition.input_model.model_json_schema()


def test_catalog_is_sorted_and_stable(client):
    first = client.get("/tools").json()["tools"]
    second = client.get("/tools").json()["tools"]
    assert [tool["name"] for tool in first] == sorted(tool["name"] for tool in first)
    assert [(tool["name"], tool["input_schema"]) for tool in first] == [
        (tool["name"], tool["input_schema"]) for tool in second
    ]


def test_package_ships_code_only():
    """查询工具不暴露配置面：agent_core 包内不得随包投放数据文件.

    该不变量原先只写在构建脚本的产物校验里；放进契约测试后，包布局变化只需改一处，
    且不需要先构建即可回归。构建器自身的 pyproject 与构建脚本的 *.orig 备份不算随包数据。
    """
    data_files = []
    for path in PACKAGE_DIR.rglob("*"):
        if not path.is_file():
            continue
        relative = path.relative_to(PACKAGE_DIR)
        if any(part in IGNORED_PARTS or part.endswith(".egg-info") for part in relative.parts[:-1]):
            continue
        if path.suffix == ".py" or relative.name in PACKAGE_PACKAGING_FILES or relative.name.endswith(".orig"):
            continue
        data_files.append(str(relative))
    assert not data_files, "agent_core 只能包含代码, 发现随包数据文件: " + ", ".join(sorted(data_files))
