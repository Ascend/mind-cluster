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

"""Capability discovery contract: complete/sorted catalog and the three-field descriptor."""

# pylint: disable=redefined-outer-name,duplicate-code  # pytest fixtures are injected by parameter name; shared test prologue

from __future__ import annotations

import pytest
from agent_core import query_api
from agent_core.tools.registry import TOOL_REGISTRY
from fastapi import FastAPI
from fastapi.testclient import TestClient

ALL_TOOLS = [
    "cluster_resource",
    "diagnose",
    "faulty_device",
    "isolated_device",
    "task_status",
]
NODE_TOOLS = ["cluster_resource", "faulty_device", "isolated_device"]


@pytest.fixture
def client():
    app = FastAPI()
    app.include_router(query_api.router)
    return TestClient(app)


def test_catalog_lists_all_five_tools_sorted(client, monkeypatch):
    def _boom(*_args, **_kwargs):
        raise AssertionError("discovery must not execute a handler")

    monkeypatch.setattr(TOOL_REGISTRY, "execute", _boom)
    response = client.get("/tools")

    assert response.status_code == 200
    body = response.json()
    assert body["request_id"]
    names = [tool["name"] for tool in body["tools"]]
    assert names == ALL_TOOLS


def test_descriptor_carries_only_name_description_and_input_schema(client):
    body = client.get("/tools").json()
    by_name = {tool["name"]: tool for tool in body["tools"]}

    for name in ALL_TOOLS:
        descriptor = by_name[name]
        assert set(descriptor) == {"name", "description", "input_schema"}
        assert descriptor["description"]
        assert descriptor["input_schema"]["type"] == "object"

    for name in NODE_TOOLS:
        assert set(by_name[name]["input_schema"]["properties"]) == {"node"}
    assert set(by_name["task_status"]["input_schema"]["properties"]) == {"job", "namespace"}
    diagnose = by_name["diagnose"]
    assert set(diagnose["input_schema"]["properties"]) == {"job", "namespace", "refresh"}


def test_descriptor_input_schema_comes_from_the_registered_model(client):
    catalog = {item["name"]: item for item in client.get("/tools").json()["tools"]}
    for name in TOOL_REGISTRY.names():
        expected = TOOL_REGISTRY.get(name).input_model.model_json_schema()
        assert catalog[name]["input_schema"] == expected


def test_catalog_needs_no_query_parameter(client):
    """Discovery takes no parameter: an unexpected query string cannot change the catalog."""
    assert client.get("/tools?unexpected=true").json()["tools"] == client.get("/tools").json()["tools"]


def test_catalog_entry_matches_definition(client):
    catalog = client.get("/tools").json()
    by_name = {tool["name"]: tool for tool in catalog["tools"]}
    descriptor = by_name["faulty_device"]
    assert descriptor["name"] == "faulty_device"
    assert "node" in descriptor["input_schema"]["properties"]
    assert "refresh" not in descriptor["input_schema"]["properties"]


def test_unknown_tool_execution_is_404(client):
    response = client.post("/tools/does_not_exist", json={})
    assert response.status_code == 404
    body = response.json()
    assert body["status"] == "error"
    assert body["errors"][0]["code"] == "TOOL_NOT_FOUND"


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
