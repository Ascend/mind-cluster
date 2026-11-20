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

The package layout is the same kind of contract: ``agent-core`` ships code plus the static
payload named by the pyproject package-data whitelist, and nothing else.
"""

# pylint: disable=redefined-outer-name  # pytest fixtures are injected by parameter name

from __future__ import annotations

import re
from fnmatch import fnmatch
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
IGNORED_PARTS = {"build", "dist", ".eggs", "__pycache__", ".pytest_cache", ".mypy_cache", ".ruff_cache", ".tox"}


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


def _package_data_whitelist() -> list[str]:
    """The pyproject package-data patterns: the only static payload the package may ship."""
    text = (PACKAGE_DIR / "pyproject.toml").read_text(encoding="utf-8")
    block = text.split("[tool.setuptools.package-data]", 1)[1].split("\n[", 1)[0]
    patterns = re.findall(r'"([^"]+)"', block)
    assert patterns, "pyproject declares no package-data: the static payload would never ship"
    return patterns


def _data_files():
    """Every file under the package root, build litter and caches excluded."""
    for path in PACKAGE_DIR.rglob("*"):
        relative = path.relative_to(PACKAGE_DIR)
        if not path.is_file():
            continue
        if any(part in IGNORED_PARTS or part.endswith(".egg-info") for part in relative.parts[:-1]):
            continue
        yield path, relative


def test_package_ships_code_only():
    """agent_core 包内只投放代码：pyproject 的 package-data 白名单之外不得有随包数据文件。

    白名单即 WebUI 与技能正文这两类静态载荷的唯一来源；构建器 pyproject 与 *.orig 备份不算随包数据。
    """
    whitelist = _package_data_whitelist()
    assert all(pattern.startswith(("webui/", "skills/")) for pattern in whitelist), (
        "the package-data whitelist may only cover the WebUI and the skill bodies"
    )
    data_files = []
    for path, relative in _data_files():
        posix = relative.as_posix()
        if any(fnmatch(posix, pattern) for pattern in whitelist):
            continue
        build_artifact = path.suffix in {".py", ".whl"} or relative.name.endswith(".orig")
        if build_artifact or relative.name in PACKAGE_PACKAGING_FILES:
            continue
        data_files.append(posix)
    assert not data_files, "agent_core 只能包含代码, 发现随包数据文件: " + ", ".join(sorted(data_files))
