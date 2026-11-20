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

"""HTTP contracts: query envelopes and the original diagnosis JSON."""

# pylint: disable=redefined-outer-name,duplicate-code  # pytest fixtures are injected by parameter name; shared test prologue

import pytest
from agent_core import query_api
from agent_core.tools import diagnosis as tools
from agent_core.datasources.query_context import QueryContext
from agent_core.datasources.gateway import SourceReadResult
from agent_core.query_schemas import QueryResult
from agent_core.tools.registry import TOOL_REGISTRY
from fastapi import FastAPI
from fastapi.testclient import TestClient

NPU = "huawei.com/Ascend910"


@pytest.fixture
def client():
    app = FastAPI()
    app.include_router(query_api.router)
    return TestClient(app)


class StubGateway:
    """Node/Pod listing plus empty source reads: the minimum cluster_resource needs."""

    def __init__(self, ctx):
        self.ctx = ctx

    def list_nodes(self, names=None):
        items = [
            {
                "metadata": {"name": "worker-a"},
                "status": {
                    "capacity": {"cpu": "16", NPU: "8"},
                    "allocatable": {"cpu": "16", NPU: "8"},
                    "conditions": [{"type": "Ready", "status": "True"}],
                },
            }
        ]
        return SourceReadResult(status="ok", source_id="s1", items=items)

    def list_pods(self, namespace=None):
        return SourceReadResult(status="empty", source_id="s2", items=[])

    def list_cms(self, namespace=None, label_selector=None, name_prefix=None):
        return SourceReadResult(status="empty", source_id="s3", items=[])

    def get_cm(self, namespace, name, absent_expected=False, *, use_cache=True):
        # Every ConfigMap source reports "no record": legitimate for absent_expected
        # reads and a logged-only miss for the rest, which is enough for this stub.
        status = "absent_expected" if absent_expected else "missing"
        return SourceReadResult(status=status, source_id="s4")


@pytest.fixture
def stub_cluster(monkeypatch):
    class StubContext(QueryContext):
        def __init__(self):
            super().__init__()
            self.gateway = StubGateway(self)

    monkeypatch.setattr(query_api, "QueryContext", StubContext)


def test_success_returns_result_envelope(client, stub_cluster):
    response = client.post("/tools/cluster_resource", json={})

    assert response.status_code == 200
    body = response.json()
    assert body["tool"] == "cluster_resource"
    assert body["status"] == "ok"
    assert isinstance(body["elapsed_ms"], int) and body["elapsed_ms"] >= 0
    node = body["data"]["node"]["worker-a"]
    assert node["cpu"]["capacity_cores"] == 16.0
    # cpu and the whole-card bucket are sibling keys (no resources wrapper).
    assert node["Ascend910"] == {
        "capacity": 8,
        "allocatable": 8,
        "used": 0,
        "idle": 8,
        "fault": None,
        "isolated": 0,
    }


def test_resource_name_is_rejected_as_an_unknown_argument(client, stub_cluster):
    # `resource_name` was never in the RFC: the server only accepts `node`.
    response = client.post("/tools/cluster_resource", json={"resource_name": "huawei.com/npu"})
    assert response.status_code == 422
    issue = response.json()["errors"][0]
    assert issue["code"] == "INVALID_ARGUMENT"
    assert issue["path"] == "resource_name"


def test_unknown_argument_is_422(client):
    response = client.post("/tools/faulty_device", json={"unexpected": 1})

    assert response.status_code == 422
    body = response.json()
    assert body["status"] == "error"
    assert body["errors"][0]["code"] == "INVALID_ARGUMENT"
    assert body["errors"][0]["path"] == "unexpected"


def test_body_must_be_json_object(client):
    response = client.post("/tools/isolated_device", json=[1, 2])

    assert response.status_code == 422
    assert response.json()["errors"][0]["code"] == "INVALID_ARGUMENT"


def test_malformed_json_uses_the_tool_envelope(client):
    response = client.post(
        "/tools/task_status",
        content='{"job": ',
        headers={"content-type": "application/json"},
    )

    assert response.status_code == 422
    body = response.json()
    assert "detail" not in body
    assert body["tool"] == "task_status"
    assert body["status"] == "error"
    assert body["errors"][0]["code"] == "INVALID_ARGUMENT"
    assert "not valid JSON" in body["errors"][0]["message"]
    assert "detail" not in response.text


def test_discovery_needs_no_configuration(client):
    # The query configuration surface was removed: discovery is unconditional.
    response = client.get("/tools")

    assert response.status_code == 200
    assert [tool["name"] for tool in response.json()["tools"]] == [
        "cluster_resource",
        "diagnose",
        "faulty_device",
        "isolated_device",
        "task_status",
    ]


def test_unexpected_error_maps_to_500_without_traceback(client, monkeypatch):
    def raise_crash(tool, payload, ctx):
        raise RuntimeError("boom: secret detail")

    monkeypatch.setattr(TOOL_REGISTRY, "execute", raise_crash)
    response = client.post("/tools/faulty_device", json={})

    assert response.status_code == 500
    body = response.json()
    assert body["errors"][0]["code"] == "INTERNAL_ERROR"
    assert body["request_id"] in body["errors"][0]["message"]
    assert "secret detail" not in response.text


def test_unknown_post_tool_is_404(client):
    response = client.post("/tools/not_registered", json={})
    assert response.status_code == 404
    body = response.json()
    assert body["status"] == "error"
    assert body["errors"][0]["code"] == "TOOL_NOT_FOUND"


def test_partial_result_stays_200(client, monkeypatch):
    monkeypatch.setattr(
        TOOL_REGISTRY,
        "execute",
        lambda tool, payload, ctx: QueryResult(request_id="r1", tool=tool, status="partial"),
    )
    response = client.post("/tools/faulty_device", json={})
    assert response.status_code == 200
    assert response.json()["status"] == "partial"
    assert response.json()["errors"] == []


def test_all_requested_nodes_absent_is_404_target_not_found(client, stub_cluster):
    response = client.post("/tools/faulty_device", json={"node": ["worker-z"]})

    assert response.status_code == 404
    body = response.json()
    assert body["status"] == "error"
    assert body["data"] is None
    issue = body["errors"][0]
    assert issue["code"] == "TARGET_NOT_FOUND"
    assert issue["path"] == "node"
    assert issue["message"] == "所有节点都找不到"


def test_some_requested_nodes_absent_is_partial_with_the_missing_node(client, stub_cluster):
    response = client.post("/tools/isolated_device", json={"node": ["worker-a", "worker-z"]})

    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "partial"
    assert body["errors"] == [
        {
            "code": "TARGET_NOT_FOUND",
            "source_id": None,
            "path": "node",
            "message": "节点 worker-z 不存在",
            "retryable": False,
        }
    ]


def _diagnosis_envelope(body, status, expect_data=True):
    assert body["tool"] == "diagnose"
    assert body["status"] == status
    if expect_data:
        assert isinstance(body["data"], dict)
        assert not {"task", "execution", "report", "collection", "presentation"} & body["data"].keys()


def test_diagnosis_busy_maps_to_500(client, monkeypatch):
    monkeypatch.setattr(
        tools,
        "diagnose_cached",
        lambda **kwargs: {"error": "job is being diagnosed", "pods": [], "_status": "busy"},
    )
    response = client.post("/tools/diagnose", json={"job": "train-a"})
    assert response.status_code == 500
    body = response.json()
    _diagnosis_envelope(body, "error")
    assert body["data"]["error"] == "job is being diagnosed"
    assert [issue["code"] for issue in body["errors"]] == ["INTERNAL_ERROR"]


@pytest.mark.parametrize(
    "internal,status,http_status,extra,code",
    [
        ("ok", "ok", 200, {}, None),
        ("ok", "ok", 200, {"cached": True, "final_text": "已有总结", "cached_note": "use --refresh"}, None),
        ("partial", "partial", 200, {"incomplete_nodes": ["node-b"], "empty_nodes": ["node-c"]}, None),
        ("not_found", "error", 500, {"error": "task not found"}, "INTERNAL_ERROR"),
        ("timeout", "error", 500, {"error": "diagnosis timed out"}, "INTERNAL_ERROR"),
        ("failed", "error", 500, {"error": "collection failed"}, "INTERNAL_ERROR"),
    ],
)
def test_diagnosis_result_preserves_flat_fields(client, monkeypatch, internal, status, http_status, extra, code):
    legacy = {
        "error": None,
        "pods": [{"pod_name": "worker-0"}],
        "diag_report": {"root": "node-a"},
        "diag_report_text": "原始报告",
        "extension": {"future_field": [1, None]},
        **extra,
    }
    calls = []

    def diagnose_cached(**kwargs):
        calls.append(kwargs)
        return {**legacy, "_status": internal}

    monkeypatch.setattr(tools, "diagnose_cached", diagnose_cached)
    response = client.post("/tools/diagnose", json={"job": "train-a", "namespace": "ns", "refresh": True})
    assert response.status_code == http_status
    body = response.json()
    _diagnosis_envelope(body, status)
    data = body["data"]
    for key, value in legacy.items():
        if key not in {"final_text", "llm_error"}:
            assert data[key] == value
    assert data["final_text"]  # presentation is applied automatically
    if "final_text" in extra:
        assert data["final_text"].startswith(extra["final_text"])
    assert [issue["code"] for issue in body["errors"]] == ([code] if code else [])
    assert calls == [{"job": "train-a", "namespace": "ns", "refresh": True}]


@pytest.mark.parametrize("payload", [{}, {"job": ""}, {"job": "train-a", "unexpected": 1}, [1, 2]])
def test_invalid_diagnosis_input_uses_tool_envelope(client, monkeypatch, payload):
    def unexpected_diagnosis(**kwargs):
        pytest.fail("invalid input must not trigger diagnosis")

    monkeypatch.setattr(tools, "diagnose_cached", unexpected_diagnosis)
    response = client.post("/tools/diagnose", json=payload)
    assert response.status_code == 422
    body = response.json()
    _diagnosis_envelope(body, "error", expect_data=False)
    assert body["errors"][0]["code"] == "INVALID_ARGUMENT"
    assert body["errors"][0]["message"]


def test_diagnosis_unexpected_error_uses_tool_envelope(client, monkeypatch):
    def crash(*args):
        raise RuntimeError("private detail")

    monkeypatch.setattr(TOOL_REGISTRY, "execute", crash)
    response = client.post("/tools/diagnose", json={"job": "train-a"})
    assert response.status_code == 500
    body = response.json()
    _diagnosis_envelope(body, "error", expect_data=False)
    assert body["errors"][0]["code"] == "INTERNAL_ERROR"
    assert "private detail" not in response.text


def test_diagnosis_applies_legacy_presentation_automatically(client, monkeypatch):
    from agent_core import llm_summary

    raw = {"diag_report": {"root": "n"}, "error": None, "pods": []}
    monkeypatch.setattr(tools, "diagnose_cached", lambda **kwargs: dict(raw))
    monkeypatch.setattr(llm_summary, "summarize_report", lambda report: "中文总结")
    written = []
    monkeypatch.setattr(tools, "cache_diag_final_text", lambda *args: written.append(args))

    response = client.post("/tools/diagnose", json={"job": "train-a"})
    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ok"
    presentation = body["data"]
    assert presentation["final_text"].startswith("中文总结")
    assert "--json" in presentation["final_text"]
    assert written and written[0][2].startswith("中文总结")

    # No opt-in switch: every tool call applies the same automatic presentation.
    repeated = client.post("/tools/diagnose", json={"job": "train-a"})
    assert repeated.json()["data"]["final_text"].startswith("中文总结")
