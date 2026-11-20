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

"""G01: invalid tool arguments stop before any cluster reads or diagnosis side effects."""

# pylint: disable=redefined-outer-name,shadowed-import  # pytest fixtures are injected by name; the fixture intentionally shadows the tools module

from __future__ import annotations

from unittest.mock import Mock

import pytest
from agent_core import query_api, tools
from agent_core.datasources.query_context import QueryContext
from agent_core.k8s import K8s
from agent_core.query_schemas import FaultQuery, JobQuery, NodeQuery, QueryFailure
from agent_core.tools.registry import TOOL_REGISTRY
from agent_core.tools.schemas import DiagnosisQuery
from fastapi import FastAPI
from fastapi.testclient import TestClient
from pydantic import ValidationError

NODE_TOOLS = ("faulty_device", "isolated_device", "cluster_resource")
JOB_TOOLS = ("task_status", "diagnose")
INVALID_NAMES = (
    "",
    " ",
    "Upper",
    "bad/name",
    "bad_name",
    "bad;cmd",
    "a..b",
    "a.-b",
    "a-.b",
    ".a",
    "a.",
    "-a",
    "a-",
    "a\n",
    "a\r",
    "a\x00",
    "任务",
    "a" * 254,
    123,
    True,
    None,
)
INVALID_NAMESPACES = (*INVALID_NAMES, "a.b", "a" * 64)

INVALID_CALLS = (
    [(tool, {"node": ["worker-a", value]}, "node.1") for tool in NODE_TOOLS for value in INVALID_NAMES]
    + [(tool, {"job": value}, "job") for tool in JOB_TOOLS for value in INVALID_NAMES]
    + [
        (tool, {"job": "train-a", "namespace": value}, "namespace")
        for tool in JOB_TOOLS
        for value in INVALID_NAMESPACES
    ]
    + [
        ("faulty_device", {"node": [], "refresh": True}, "refresh"),
        ("isolated_device", {"node": [], "refresh": True}, "refresh"),
        ("cluster_resource", {"node": [], "refresh": True}, "refresh"),
        ("task_status", {"job": "train-a", "refresh": True}, "refresh"),
        # RFC-less filters are gone: the server must reject them as unknown parameters.
        ("isolated_device", {"node": [], "type": "manual"}, "type"),
        ("cluster_resource", {"node": [], "resource_name": "huawei.com/npu"}, "resource_name"),
    ]
)


@pytest.fixture(autouse=True)
def no_business_io(monkeypatch):
    cluster = Mock(side_effect=AssertionError("unexpected Kubernetes access"))
    diagnosis = Mock(side_effect=AssertionError("unexpected diagnosis execution"))
    monkeypatch.setattr(K8s, "core", cluster)
    monkeypatch.setattr(tools, "diagnose_cached", diagnosis)
    yield
    cluster.assert_not_called()
    diagnosis.assert_not_called()


@pytest.fixture
def client():
    app = FastAPI()
    app.include_router(query_api.router)
    with TestClient(app) as result:
        yield result


@pytest.mark.parametrize("tool,payload,path", INVALID_CALLS)
def test_invalid_arguments_are_rejected_by_registry_and_http(client, tool, payload, path):
    with pytest.raises(ValidationError) as error:
        TOOL_REGISTRY.execute(tool, payload)
    assert ".".join(map(str, error.value.errors()[0]["loc"])) == path

    response = client.post(f"/tools/{tool}", json=payload)
    assert response.status_code == 422
    body = response.json()
    assert body["status"] == "error"
    assert body["data"] is None
    assert body["errors"][0]["code"] == "INVALID_ARGUMENT"
    assert body["errors"][0]["path"] == path


@pytest.mark.parametrize("name", ["a", "0", "train-a-0", "worker.example.com", "a" * 253])
@pytest.mark.parametrize("model", [NodeQuery, JobQuery, DiagnosisQuery])
def test_valid_names_and_maximum_length_are_preserved(model, name):
    field = "job" if model in (JobQuery, DiagnosisQuery) else "node"
    payload = {field: name if field == "job" else [name]}
    assert model.model_validate(payload).model_dump()[field] == payload[field]


@pytest.mark.parametrize("namespace", ["default", "0", "train-ns", "n" * 63])
@pytest.mark.parametrize("model", [JobQuery, DiagnosisQuery])
def test_namespace_boundaries(model, namespace):
    request = model(job="train-a", namespace=namespace)
    assert request.namespace == namespace
    assert model(job="train-a").namespace == "default"


@pytest.mark.parametrize("tool", NODE_TOOLS)
def test_empty_scope_stays_empty_for_every_node_tool(tool):
    result = TOOL_REGISTRY.execute(tool, {"node": []})
    assert result.status == "ok"
    assert result.data is not None
    assert not result.data.model_dump().get("devices")
    assert not result.data.model_dump().get("cards")
    assert result.data.model_dump().get("node", {}) == {}


def test_node_limit_is_checked_before_deduplication_and_io(monkeypatch):
    from agent_core.datasources import query_context as query_context_module

    monkeypatch.setattr(query_context_module, "MAX_TARGETS", 2)
    ctx = QueryContext()
    with pytest.raises(QueryFailure) as error:
        TOOL_REGISTRY.execute("faulty_device", {"node": ["a", "a", "a"]}, ctx)
    assert (error.value.code, error.value.http_status, error.value.path) == ("INVALID_ARGUMENT", 422, "node")
    original = ["b", "a", "b"]
    request = FaultQuery(node=original)
    monkeypatch.setattr(query_context_module, "MAX_TARGETS", 3)
    ctx.check_scope(request)
    assert ctx.requested_nodes == ["b", "a"]
    assert request.node == original == ["b", "a", "b"]
    ctx.check_scope(FaultQuery(node=[]))
    assert ctx.requested_nodes == []
    assert QueryContext().requested_nodes is None


def test_discovery_exposes_name_constraints_and_the_registry_enforces_them(client):
    catalog = {item["name"]: item for item in client.get("/tools").json()["tools"]}
    for name in TOOL_REGISTRY.names():
        descriptor = catalog[name]
        declared = descriptor["input_schema"]["properties"]
        if name in JOB_TOOLS:
            for field, limit in (("job", 253), ("namespace", 63)):
                assert declared[field]["minLength"] == 1
                assert declared[field]["maxLength"] == limit
            invalid = {"job": "train-a", "namespace": "BAD/NAMESPACE"}
        else:
            items = next(item for item in declared["node"]["anyOf"] if item["type"] == "array")["items"]
            assert items["maxLength"] == 253
            assert items["minLength"] == 1
            invalid = {"node": ["bad/node"]}
        with pytest.raises(ValidationError):
            TOOL_REGISTRY.get(name).input_model.model_validate(invalid)


def test_legacy_diag_passes_names_through_without_tool_validation(monkeypatch):
    # The legacy entry predates the tool input models: it must hand the job name to
    # the chain unchanged instead of validating it as a Kubernetes name.
    from agent_core.main import DiagReq, diag
    from agent_core.tools import diagnosis as tools

    seen = []

    def _stub(**kwargs):
        seen.append(kwargs)
        return {"error": "training/inference task not found", "pods": [], "diag_report": {}}

    monkeypatch.setattr(tools, "diagnose_cached", _stub)
    result = diag(DiagReq(job="bad/job", namespace="default"))
    assert result["error"] == "training/inference task not found"
    assert seen == [{"job": "bad/job", "namespace": "default", "refresh": False}]
