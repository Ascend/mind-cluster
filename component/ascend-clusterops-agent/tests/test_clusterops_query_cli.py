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

"""kubectl clusterops CLI: parameter mapping, JSON purity, exit codes and transport errors."""

# pylint: disable=redefined-outer-name,duplicate-code  # pytest fixtures are injected by parameter name; shared CLI test prologue

from __future__ import annotations

import http.client
import importlib.util
import json
import subprocess
import sys
from importlib.machinery import SourceFileLoader
from pathlib import Path

import pytest

PLUGIN_DIR = Path(__file__).resolve().parents[1] / "kubectl-plugin"
sys.path.insert(0, str(PLUGIN_DIR))


def _load(name: str, filename: str):
    # the CLI script has no .py suffix, so the loader must be given explicitly
    loader = SourceFileLoader(name, str(PLUGIN_DIR / filename))
    spec = importlib.util.spec_from_loader(name, loader)
    module = importlib.util.module_from_spec(spec)
    sys.modules[name] = module
    loader.exec_module(module)
    return module


client = _load("clusterops_query_client", "clusterops_query_client.py")
cli = _load("kubectl_clusterops", "kubectl-clusterops")


class _FakeForward:
    local = 12345

    def __init__(self, _service):
        pass

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb):
        return None


@pytest.fixture
def fake_http(monkeypatch, catalog):
    calls = {}

    def fake_request(local, method, path, payload=None, timeout=None):
        calls.update(local=local, method=method, path=path, payload=payload, timeout=timeout)
        if method == "GET" and path.startswith("/tools"):
            return 200, {"request_id": "r1", "tools": list(catalog.values())}
        return calls.pop("response", (200, _result("ok")))

    monkeypatch.setattr(cli.client, "PortForward", _FakeForward)
    monkeypatch.setattr(cli.client, "request_json", fake_request)
    return calls


def _result(status: str, tool: str = "faulty_device", data=None):
    return {
        "request_id": "r1",
        "tool": tool,
        "status": status,
        "data": data if data is not None else {"devices": [], "summary": {"observed_faulty_npu_count": 0}},
        "errors": [] if status != "error" else [{"code": "SOURCE_NOT_FOUND", "message": "missing"}],
    }


def _run(monkeypatch, capsys, argv):
    monkeypatch.setattr(sys, "argv", ["kubectl-clusterops"] + argv)
    code = cli.main()
    return code, capsys.readouterr()


def test_parse_service_defaults_and_explicit():
    assert client.parse_service(None) == ("agent-core", "mindx-dl", 9700)
    assert client.parse_service("svc.ns:1234") == ("svc", "ns", 1234)
    with pytest.raises(ValueError):
        client.parse_service(".ns:1")


def test_exit_code_and_error_envelope():
    assert client.exit_code(200, {"status": "ok"}) == 0
    assert client.exit_code(200, {"status": "partial"}) == 3
    assert client.exit_code(502, {"status": "error"}) == 1

    envelope = client.transport_error_envelope("diagnose", "down")
    assert envelope["status"] == "error"
    assert envelope["data"] is None
    assert envelope["elapsed_ms"] == 0
    assert set(envelope) == {"request_id", "tool", "status", "elapsed_ms", "data", "errors"}
    assert "sources" not in envelope
    assert "warnings" not in envelope
    assert envelope["errors"][0]["code"] == "TRANSPORT_ERROR"
    assert envelope["errors"][0]["retryable"] is False


@pytest.mark.parametrize(
    "argv",
    [
        ["faulty-device"],
        ["isolated-card"],
        ["cluster-resource"],
        ["task-status", "--job", "train-a"],
        ["diagnose", "--job", "train-a"],
        ["faults"],
        ["isolated"],
        ["resources"],
        ["status", "--job", "train-a"],
    ],
)
def test_removed_alias_subcommands_are_rejected(argv, fake_http, monkeypatch):
    """Only `tools` and `tool` remain; the five per-tool aliases must not be reachable."""
    monkeypatch.setattr(sys, "argv", ["kubectl-clusterops"] + argv)
    with pytest.raises(SystemExit) as exc:
        cli.main()
    assert exc.value.code == 2
    assert fake_http == {}


def test_faults_json_is_one_object_and_exit_zero(fake_http, monkeypatch, capsys):
    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device", "--json"])
    assert code == 0
    body = json.loads(captured.out)
    assert body["tool"] == "faulty_device"
    assert fake_http["path"] == "/tools/faulty_device"
    assert fake_http["method"] == "POST"
    assert fake_http["payload"] == {}


def test_partial_result_exits_three(fake_http, monkeypatch, capsys):
    fake_http["response"] = (200, _result("partial"))
    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device", "--json"])
    assert code == 3
    assert json.loads(captured.out)["status"] == "partial"
    # stdout stays exactly one JSON object; the envelope has no warnings to spill on stderr.
    assert captured.err == ""


def test_partial_table_summary_points_at_request_id_and_logs(fake_http, monkeypatch, capsys):
    fake_http["response"] = (200, _result("partial"))
    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device"])

    assert code == 3
    assert "partial result (request_id=r1)" in captured.err
    assert "see agent-core logs" in captured.err
    assert "see response for details" not in captured.err
    assert "warning:" not in captured.out


def test_http_error_body_exits_one(fake_http, monkeypatch, capsys):
    fake_http["response"] = (502, _result("error"))
    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device"])
    assert code == 1
    assert "SOURCE_NOT_FOUND" in captured.err


def test_transport_error_uses_common_envelope(fake_http, monkeypatch, capsys):
    def _boom(*_args, **_kwargs):
        raise client.TransportError("cannot connect")

    monkeypatch.setattr(cli.client, "request_json", _boom)
    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device", "--json"])
    assert code == 1
    body = json.loads(captured.out)
    assert body["errors"][0]["code"] == "TRANSPORT_ERROR"
    assert body["data"] is None


def test_request_json_translates_remote_disconnected(monkeypatch):
    """urlopen() raises RemoteDisconnected directly when the peer closes the connection without a response."""
    message = "Remote end closed connection without response"

    def _closed(*_args, **_kwargs):
        raise http.client.RemoteDisconnected(message)

    monkeypatch.setattr(client.urllib.request, "urlopen", _closed)
    with pytest.raises(client.TransportError) as exc:
        client.request_json(12345, "GET", "/tools")
    assert "agent-core closed the connection" in str(exc.value)


def test_remote_disconnected_prints_json_and_no_traceback(fake_http, monkeypatch, capsys):
    """The CLI must fall back to a TRANSPORT_ERROR envelope even if a raw HTTPException escapes the client."""
    message = "Remote end closed connection without response"

    def _closed(*_args, **_kwargs):
        raise http.client.RemoteDisconnected(message)

    monkeypatch.setattr(cli.client, "request_json", _closed)
    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device", "--json"])
    assert code == 1
    assert captured.out
    assert "Traceback" not in captured.err
    body = json.loads(captured.out)
    assert body["status"] == "error"
    assert body["data"] is None
    assert body["errors"][0]["code"] == "TRANSPORT_ERROR"
    assert message in body["errors"][0]["message"]


def test_generic_call_is_not_a_cli_command(fake_http, monkeypatch):
    monkeypatch.setattr(sys, "argv", ["kubectl-clusterops", "call", "diagnose", "--args-json", "{}"])
    with pytest.raises(SystemExit) as exc:
        cli.main()
    assert exc.value.code == 2
    assert fake_http == {}


def test_management_and_business_subcommands_are_mutually_exclusive(monkeypatch):
    monkeypatch.setattr(sys, "argv", ["kubectl-clusterops", "--clear-llm-config", "tool"])
    with pytest.raises(SystemExit) as exc:
        cli.main()
    assert exc.value.code == 2


def test_faults_table_shows_node_records_and_node_health(fake_http, monkeypatch, capsys):
    data = {
        "node": {
            "worker-a": {
                "devices": [
                    {
                        "device_id": "3",
                        "device_type": "npu",
                        "chip_type": "Ascend910",
                        "fault_type": "CardUnhealthy",
                        "faults": [
                            {
                                "fault_code": "8",
                                "fault_handling": "NotHandleFault",
                                "last_time": "2026-10-30T02:15:00Z",
                            },
                            {"fault_code": "9", "fault_handling": "SeparateNPU", "last_time": None},
                        ],
                        "superpod": {"super_pod_id": 3, "rack_id": 1, "server_index": 2},
                    },
                    {
                        "device_id": "",
                        "device_type": "psu",
                        "chip_type": None,
                        "fault_type": "PSU",
                        "faults": [{"fault_code": "99", "fault_handling": "NotHandleFault", "last_time": None}],
                    },
                ],
                "node_health": {"node_status": "UnHealthy"},
            }
        }
    }
    fake_http["response"] = (200, _result("ok", data=data))
    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device"])

    assert code == 0
    assert "NODE\tSOURCE\tDEVICE\tTYPE\tHANDLING\tCODES\tLAST_TIME\tEXTRA" in captured.out
    assert (
        "worker-a\tdevices\t3\tCardUnhealthy\tNotHandleFault\t8\t2026-10-30T02:15:00Z\tsuperpod=3/1/2" in captured.out
    )
    assert "worker-a\tdevices\t3\tCardUnhealthy\tSeparateNPU\t9\t\tsuperpod=3/1/2" in captured.out
    assert "worker-a\tdevices\t-\tPSU\tNotHandleFault\t99\t\t" in captured.out
    assert "NODE_HEALTH" in captured.out
    assert "worker-a\tUnHealthy" in captured.out


def test_faults_table_shows_switch_dpu_and_public_faults(fake_http, monkeypatch, capsys):
    data = {
        "node": {
            "worker-a": {
                "devices": [],
                "switch": [
                    {
                        "device_id": "switch:0:16",
                        "fault_handling": "SubHealthFault",
                        "fault_code": "[0x00f10509,132332,npu,na]",
                        "last_time": "2026-11-16T02:00:00.000Z",
                    }
                ],
                "dpu": [
                    {
                        "device_id": "dpu:0x8200",
                        "fault_handling": "SubHealthFault",
                        "fault_code": "22000025",
                        "last_time": "2026-11-16T08:28:55.731Z",
                        "affected_npu": ["0", "1"],
                    },
                    {
                        "device_id": "",
                        "fault_handling": "Busy",
                        "fault_code": "1",
                        "last_time": None,
                    },
                ],
                "public_faults": [
                    {
                        "device_id": "PSU",
                        "fault_type": "PublicFault",
                        "fault_handling": "NotHandleFault",
                        "fault_code": "0300000D",
                        "last_time": "2026-11-16T00:00:00Z",
                    }
                ],
                "node_health": {"node_status": "Healthy"},
            }
        }
    }
    fake_http["response"] = (200, _result("ok", data=data))
    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device"])

    assert code == 0
    assert (
        "worker-a\tswitch\tswitch:0:16\t\tSubHealthFault\t"
        "[0x00f10509,132332,npu,na]\t2026-11-16T02:00:00.000Z\t" in captured.out
    )
    assert (
        "worker-a\tdpu\tdpu:0x8200\t\tSubHealthFault\t22000025\t"
        "2026-11-16T08:28:55.731Z\taffected_npu=0,1" in captured.out
    )
    # Node-level DPU events carry no affected_npu relation.
    assert "worker-a\tdpu\t-\t\tBusy\t1\t\t" in captured.out
    assert "worker-a\tpublic_faults\tPSU\tPublicFault\tNotHandleFault\t0300000D\t2026-11-16T00:00:00Z\t" in captured.out
    assert "worker-a\tHealthy" in captured.out


def test_resources_table_shows_cpu_and_card_buckets(fake_http, monkeypatch, capsys):
    data = {
        "node": {
            "worker-a": {
                "cpu": {
                    "capacity_cores": 96.0,
                    "allocatable_cores": 96.0,
                    "used_cores": 0.5,
                    "idle_cores": 95.5,
                },
                "Ascend910": {
                    "capacity": 16,
                    "allocatable": 16,
                    "used": 0,
                    "idle": 16,
                    "fault": 1,
                    "isolated": None,
                },
                "npu": {
                    "capacity": 4,
                    "allocatable": 4,
                    "used": 1,
                    "idle": 3,
                    "fault": 0,
                    "isolated": 0,
                },
            }
        }
    }
    fake_http["response"] = (200, _result("ok", tool="cluster_resource", data=data))
    code, captured = _run(monkeypatch, capsys, ["tool", "cluster_resource"])

    assert code == 0
    assert "NODE\tRESOURCE\tCAPACITY\tALLOCATABLE\tUSED\tIDLE\tFAULT\tISOLATED" in captured.out
    assert "worker-a\tcpu\t96.0\t96.0\t0.5\t95.5" in captured.out
    # unknown (None) counts render as an empty cell, never as "None"
    assert "worker-a\tAscend910\t16\t16\t0\t16\t1\t" in captured.out
    assert "worker-a\tnpu\t4\t4\t1\t3\t0\t0" in captured.out


def test_tools_list_calls_discovery_endpoint(fake_http, monkeypatch, capsys, catalog):
    code, captured = _run(monkeypatch, capsys, ["tools", "--json"])
    assert code == 0
    assert fake_http["method"] == "GET"
    assert fake_http["path"] == "/tools"
    assert [item["name"] for item in json.loads(captured.out)["tools"]] == list(catalog)


# --------------------------------------------------------------------------- #
# Generic `tool <name>` entry: schema-driven parameters and dispatch
# --------------------------------------------------------------------------- #
NODES_SCHEMA = {
    "type": "object",
    "properties": {
        "node": {
            "anyOf": [{"type": "array", "items": {"type": "string"}}, {"type": "null"}],
            "default": None,
        },
    },
    "additionalProperties": False,
}


def _parse(schema, params=None, params_json=None):
    return cli._parse_tool_params(schema, params or [], params_json)


def _error_envelope(tool="faulty_device", code="TOOL_NOT_FOUND", message="missing"):
    body = _result("error", tool=tool)
    body["errors"] = [{"code": code, "source_id": None, "path": None, "message": message, "retryable": False}]
    return body


@pytest.fixture(scope="module")
def catalog():
    # The CLI itself only needs the stdlib; importing the registry pulls the agent-core
    # dependency stack, so catalog-driven tests skip when that stack is not installed.
    registry = pytest.importorskip("agent_core.tools.registry")
    return {descriptor.name: descriptor.model_dump(mode="json") for descriptor in registry.TOOL_REGISTRY.describe()}


@pytest.fixture
def fake_tool_api(monkeypatch, catalog):
    """Fake agent-core: GET serves the real descriptor, POST returns a configurable envelope."""
    state = {"calls": [], "responses": {}}

    def fake_request(local, method, path, payload=None, timeout=None):
        state["calls"].append({"method": method, "path": path, "payload": payload, "timeout": timeout})
        if (method, path) in state["responses"]:
            return state["responses"][(method, path)]
        if method == "GET":
            return 200, {"request_id": "r1", "tools": list(catalog.values())}
        return 200, _result("ok", tool=path.rsplit("/", 1)[-1])

    monkeypatch.setattr(cli.client, "PortForward", _FakeForward)
    monkeypatch.setattr(cli.client, "request_json", fake_request)
    return state


def test_parse_tool_params_converts_scalar_types():
    schema = {
        "type": "object",
        "properties": {
            "job": {"type": "string"},
            "refresh": {"type": "boolean"},
            "retries": {"type": "integer"},
            "ratio": {"type": "number"},
        },
        "required": ["job"],
    }
    payload, error = _parse(schema, ["job=train-a", "refresh=TRUE", "retries=3", "ratio=0.5"])

    assert error is None
    assert payload == {"job": "train-a", "refresh": True, "retries": 3, "ratio": 0.5}


def test_parse_tool_params_boolean_and_integer_are_strict():
    schema = {"type": "object", "properties": {"refresh": {"type": "boolean"}, "n": {"type": "integer"}}}

    for params, fragment in (
        (["refresh=yes"], "true or false"),
        (["n=1.5"], "expected an integer"),
        (["n=abc"], "expected an integer"),
    ):
        payload, error = _parse(schema, params)
        assert payload is None
        assert fragment in error


def test_parse_tool_params_array_repeat_single_and_null():
    assert _parse(NODES_SCHEMA, ["node=worker-a", "node=worker-b"]) == ({"node": ["worker-a", "worker-b"]}, None)
    assert _parse(NODES_SCHEMA, ["node=worker-a"]) == ({"node": ["worker-a"]}, None)
    assert _parse(NODES_SCHEMA, ["node=null"]) == ({"node": None}, None)


def test_parse_tool_params_empty_value_is_empty_string_not_null():
    # Documented in the design: use --params-json '{"node": []}' for an empty scope.
    assert _parse(NODES_SCHEMA, ["node="]) == ({"node": [""]}, None)


def test_parse_tool_params_enum_is_checked_locally():
    schema = {
        "type": "object",
        "properties": {"type": {"type": "string", "enum": ["all", "manual", "silent", "unknown"], "default": "all"}},
    }

    assert _parse(schema, ["type=manual"]) == ({"type": "manual"}, None)
    payload, error = _parse(schema, ["type=automatic"])
    assert payload is None
    assert "allowed: all, manual, silent, unknown" in error


def test_parse_tool_params_string_constraints_are_checked_locally():
    schema = {
        "type": "object",
        "properties": {
            "node": {
                "type": "string",
                "minLength": 1,
                "maxLength": 8,
                "pattern": "^[a-z0-9](?:[-a-z0-9]*[a-z0-9])?$",
            },
        },
    }

    assert _parse(schema, ["node=worker-a"]) == ({"node": "worker-a"}, None)
    for params in (["node=Worker-A"], ["node="], ["node=worker-a-b"]):
        payload, error = _parse(schema, params)
        assert payload is None
        assert "invalid value for --param node" in error


def test_parse_tool_params_rejects_unknown_malformed_and_missing():
    schema = {"type": "object", "properties": {"job": {"type": "string"}}, "required": ["job"]}

    for params, fragment in (
        (["node=worker-a"], "unknown parameter(s): node"),
        (["job"], "expected NAME=VALUE"),
        (["=value"], "empty parameter name"),
        ([], "missing required parameter(s): job"),
    ):
        payload, error = _parse(schema, params)
        assert payload is None
        assert fragment in error


def test_parse_tool_params_rejects_repeated_scalar_and_array_single_only():
    schema = {"type": "object", "properties": {"job": {"type": "string"}}}

    payload, error = _parse(schema, ["job=a", "job=b"])
    assert payload is None
    assert "does not accept multiple values" in error

    array_schema = {"type": "object", "properties": {"node": {"type": "array", "items": {"type": "string"}}}}
    assert _parse(array_schema, ["node=a", "node=b"]) == ({"node": ["a", "b"]}, None)


def test_parse_tool_params_object_fields_require_params_json():
    schema = {"type": "object", "properties": {"labels": {"type": "object"}}}

    payload, error = _parse(schema, ["labels=a=b"])
    assert payload is None
    assert "use --params-json" in error


def test_parse_tool_params_params_json_supports_empty_array_and_null():
    assert _parse(NODES_SCHEMA, params_json='{"node": []}') == ({"node": []}, None)
    assert _parse(NODES_SCHEMA, params_json='{"node": null}') == ({"node": None}, None)


def test_parse_tool_params_params_json_errors():
    for params_json, fragment in (
        ("not-json", "invalid --params-json"),
        ("[1, 2]", "must contain a JSON object"),
        ('{"node": [], "unknown": 1}', "unknown parameter(s): unknown"),
    ):
        payload, error = _parse(NODES_SCHEMA, params_json=params_json)
        assert payload is None
        assert fragment in error


def test_parse_tool_params_conflicting_sources_are_rejected():
    payload, error = _parse(NODES_SCHEMA, ["node=a"], '{"node": []}')
    assert payload is None
    assert "mutually exclusive" in error


def test_generic_tool_maps_params_to_get_then_post(fake_tool_api, monkeypatch, capsys):
    code, captured = _run(
        monkeypatch,
        capsys,
        ["tool", "faulty_device", "--param", "node=worker-a", "--param", "node=worker-b", "--json"],
    )

    assert code == 0
    assert [call["method"] for call in fake_tool_api["calls"]] == ["GET", "POST"]
    assert fake_tool_api["calls"][0]["path"] == "/tools"
    assert fake_tool_api["calls"][1]["path"] == "/tools/faulty_device"
    assert fake_tool_api["calls"][1]["payload"] == {"node": ["worker-a", "worker-b"]}
    assert fake_tool_api["calls"][1]["timeout"] == client.QUERY_TIMEOUT
    assert json.loads(captured.out)["tool"] == "faulty_device"


def test_generic_tool_expression_covers_every_schema_field(fake_tool_api, catalog, monkeypatch, capsys):
    for name, descriptor in catalog.items():
        fake_tool_api["calls"].clear()
        argv = ["tool", name, "--json"] + _schema_argv(descriptor["input_schema"])
        code, captured = _run(monkeypatch, capsys, argv)

        assert code == 0, name
        assert [call["method"] for call in fake_tool_api["calls"]] == ["GET", "POST"], name
        assert fake_tool_api["calls"][0]["path"] == "/tools", name
        assert fake_tool_api["calls"][1]["path"] == f"/tools/{name}"
        properties = set(descriptor["input_schema"].get("properties") or {})
        assert set(fake_tool_api["calls"][1]["payload"]) == properties, name
        assert json.loads(captured.out)["status"] == "ok"


@pytest.mark.parametrize(
    "tool, args, expected",
    [
        ("faulty_device", ["--param", "node=a", "--param", "node=b"], {"node": ["a", "b"]}),
        ("faulty_device", ["--param", "node=null"], {"node": None}),
        ("isolated_device", ["--param", "node=a"], {"node": ["a"]}),
        ("cluster_resource", ["--params-json", '{"node":[]}'], {"node": []}),
        ("task_status", ["--param", "job=train-a"], {"job": "train-a"}),
        ("task_status", [], None),
        ("faulty_device", ["--param", "refresh=true"], None),
        ("faulty_device", ["--param", "bogus=true"], None),
        # RFC-less filters are not part of the schema: the plugin rejects them locally.
        ("isolated_device", ["--param", "type=manual"], None),
        ("isolated_device", ["--param", "type=invalid"], None),
        ("cluster_resource", ["--param", "resource_name=huawei.com/npu"], None),
        ("cluster_resource", ["--param", "node=a", "--params-json", "{}"], None),
        ("task_status", ["--param", "namespace=test"], None),
    ],
)
def test_generic_tool_contract(tool, args, expected, fake_tool_api, monkeypatch, capsys):
    path = f"/tools/{tool}"
    code, captured = _run(monkeypatch, capsys, ["tool", tool] + args + ["--json"])
    assert code == (2 if expected is None else 0)
    calls = fake_tool_api["calls"]
    assert calls[0]["method"] == "GET" and calls[0]["path"] == "/tools"
    if expected is None:
        assert len(calls) == 1
    else:
        assert (calls[1]["method"], calls[1]["path"]) == ("POST", path)
        assert calls[1]["payload"] == expected
        assert json.loads(captured.out)["tool"] == tool


def test_generic_tool_params_json_empty_nodes(fake_tool_api, monkeypatch, capsys):
    code, _ = _run(monkeypatch, capsys, ["tool", "faulty_device", "--params-json", '{"node": []}', "--json"])

    assert code == 0
    assert fake_tool_api["calls"][1]["payload"] == {"node": []}


def test_generic_tool_rejects_the_removed_filters_locally(fake_tool_api, monkeypatch, capsys):
    """`type` and `resource_name` are gone from the schema, so the plugin never sends them."""
    for argv in (
        ["tool", "isolated_device", "--param", "type=manual", "--json"],
        ["tool", "cluster_resource", "--param", "resource_name=huawei.com/npu", "--json"],
    ):
        fake_tool_api["calls"].clear()
        code, captured = _run(monkeypatch, capsys, argv)
        assert code == 2
        assert [call["method"] for call in fake_tool_api["calls"]] == ["GET"]
        assert "unknown parameter" in captured.err


def test_collect_manifest_for_diagnose_writes_configmap_and_skips_agent(monkeypatch, capsys, tmp_path):
    manifest = tmp_path / "collect_manifest.yaml"
    manifest.write_text("version: 3\n", encoding="utf-8")
    calls = []

    def fake_run(args, **kwargs):
        calls.append((args, kwargs))
        if "create" in args:
            return subprocess.CompletedProcess(args, 0, "apiVersion: v1\nkind: ConfigMap\n", "")
        return subprocess.CompletedProcess(args, 0, "", "")

    monkeypatch.setattr(cli, "_kubectl", lambda: "kubectl")
    monkeypatch.setattr(cli.subprocess, "run", fake_run)

    code, captured = _run(monkeypatch, capsys, ["tool", "diagnose", "--param", f"collect_manifest={manifest}"])

    assert code == 0
    assert "collect manifest written to ConfigMap" in captured.out
    assert [call[0][1:3] for call in calls] == [["create", "configmap"], ["apply", "-f"]]
    assert "--dry-run=client" in calls[0][0]
    assert calls[1][1]["input"].startswith("apiVersion: v1")


def test_collect_manifest_accepts_param_alias(monkeypatch, capsys, tmp_path):
    manifest = tmp_path / "collect_manifest.yaml"
    manifest.write_text("version: 3\n", encoding="utf-8")
    calls = []

    def fake_run(args, **kwargs):
        calls.append((args, kwargs))
        if "create" in args:
            return subprocess.CompletedProcess(args, 0, "apiVersion: v1\nkind: ConfigMap\n", "")
        return subprocess.CompletedProcess(args, 0, "", "")

    monkeypatch.setattr(cli, "_kubectl", lambda: "kubectl")
    monkeypatch.setattr(cli.subprocess, "run", fake_run)

    code, captured = _run(monkeypatch, capsys, ["tool", "diagnose", "--param", f"collect-manifest={manifest}"])

    assert code == 0
    assert "collect manifest written to ConfigMap" in captured.out
    assert len(calls) == 2


def test_collect_manifest_flag_rejects_non_diagnose_tool(monkeypatch, capsys):
    monkeypatch.setattr(sys, "argv", ["kubectl-clusterops", "tool", "faulty_device", "--collect-manifest", "x.yaml"])

    with pytest.raises(SystemExit) as exc:
        cli.main()

    assert exc.value.code == 2
    assert "unrecognized arguments: --collect-manifest" in capsys.readouterr().err


def test_collect_manifest_missing_file_returns_one(monkeypatch, capsys, tmp_path):
    monkeypatch.setattr(cli, "_kubectl", lambda: "kubectl")
    code, captured = _run(
        monkeypatch, capsys, ["tool", "diagnose", "--param", f"collect_manifest={tmp_path / 'missing.yaml'}"]
    )

    assert code == 1
    assert "does not exist" in captured.err


def test_generic_diagnose_uses_diag_timeout_without_summary_option(fake_tool_api, monkeypatch, capsys):
    code, _ = _run(monkeypatch, capsys, ["tool", "diagnose", "--param", "job=train-a", "--json"])

    assert code == 0
    assert fake_tool_api["calls"][1]["payload"] == {"job": "train-a"}
    assert fake_tool_api["calls"][1]["timeout"] == client.DIAG_TIMEOUT


@pytest.mark.parametrize("status, expected", [("ok", 0), ("partial", 3), ("error", 1)])
def test_generic_query_exit_codes(status, expected, fake_tool_api, monkeypatch, capsys):
    fake_tool_api["responses"][("POST", "/tools/faulty_device")] = (200 if status != "error" else 502, _result(status))

    code, _ = _run(monkeypatch, capsys, ["tool", "faulty_device", "--json"])
    assert code == expected


@pytest.mark.parametrize("status, expected", [("ok", 0), ("partial", 0), ("error", 1)])
def test_generic_diagnosis_exit_codes(status, expected, fake_tool_api, monkeypatch, capsys):
    fake_tool_api["responses"][("POST", "/tools/diagnose")] = (200 if status != "error" else 500, _result(status))

    code, _ = _run(monkeypatch, capsys, ["tool", "diagnose", "--param", "job=train-a", "--json"])
    assert code == expected


def test_generic_json_output_is_the_http_body(fake_tool_api, monkeypatch, capsys):
    body = _result("ok", tool="isolated_device", data={"node": {}})
    fake_tool_api["responses"][("POST", "/tools/isolated_device")] = (200, body)

    code, captured = _run(monkeypatch, capsys, ["tool", "isolated_device", "--param", "node=a", "--json"])

    assert code == 0
    assert json.loads(captured.out) == body


def test_generic_unknown_tool_exits_two_without_post(fake_tool_api, monkeypatch, capsys):
    code, captured = _run(monkeypatch, capsys, ["tool", "does_not_exist", "--json"])

    assert code == 2
    assert [call["method"] for call in fake_tool_api["calls"]] == ["GET"]
    assert captured.out == ""
    assert "unknown tool" in captured.err


def test_generic_local_param_error_exits_two_without_post(fake_tool_api, monkeypatch, capsys):
    code, captured = _run(
        monkeypatch, capsys, ["tool", "diagnose", "--param", "job=train-a", "--param", "refresh=yes", "--json"]
    )

    assert code == 2
    assert [call["method"] for call in fake_tool_api["calls"]] == ["GET"]
    assert captured.out == ""
    assert "true or false" in captured.err


def test_generic_descriptor_transport_error_returns_envelope(fake_tool_api, monkeypatch, capsys):
    def _boom(*_args, **_kwargs):
        raise client.TransportError("cannot connect")

    monkeypatch.setattr(cli.client, "request_json", _boom)
    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device", "--json"])

    assert code == 1
    body = json.loads(captured.out)
    assert body["status"] == "error"
    assert body["data"] is None
    assert body["errors"][0]["code"] == "TRANSPORT_ERROR"


def test_generic_catalog_failure_exits_one(fake_tool_api, monkeypatch, capsys):
    fake_tool_api["responses"][("GET", "/tools")] = (
        500,
        _error_envelope(tool="catalog", code="SOURCE_NOT_FOUND", message="catalog unavailable"),
    )

    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device", "--json"])

    assert code == 1
    assert [call["method"] for call in fake_tool_api["calls"]] == ["GET"]
    assert "SOURCE_NOT_FOUND" in captured.out


def test_generic_tool_forbidden_exits_one(fake_tool_api, monkeypatch, capsys):
    fake_tool_api["responses"][("POST", "/tools/task_status")] = (
        403,
        _error_envelope(tool="task_status", code="SOURCE_FORBIDDEN", message="forbidden by policy"),
    )

    code, captured = _run(monkeypatch, capsys, ["tool", "task_status", "--param", "job=job-x"])

    assert code == 1
    assert "SOURCE_FORBIDDEN" in captured.err
    assert [call["method"] for call in fake_tool_api["calls"]] == ["GET", "POST"]


def test_generic_human_output_uses_the_known_renderer(fake_tool_api, monkeypatch, capsys):
    data = {"devices": [], "node_health": []}
    fake_tool_api["responses"][("POST", "/tools/faulty_device")] = (200, _result("ok", data=data))

    code, captured = _run(monkeypatch, capsys, ["tool", "faulty_device"])

    assert code == 0
    assert "NODE\tSOURCE\tDEVICE\tTYPE\tHANDLING\tCODES\tLAST_TIME\tEXTRA" in captured.out
    assert "no fault devices or node health records" in captured.out


def test_generic_human_output_falls_back_for_unknown_renderer(fake_tool_api, catalog, monkeypatch, capsys):
    future = {
        "name": "future_tool",
        "description": "not registered yet",
        "input_schema": {"type": "object", "properties": {"flag": {"type": "boolean"}}},
    }
    fake_tool_api["responses"][("GET", "/tools")] = (
        200,
        {"request_id": "r1", "tools": list(catalog.values()) + [future]},
    )
    fake_tool_api["responses"][("POST", "/tools/future_tool")] = (
        200,
        _result("ok", tool="future_tool", data={"answer": 42}),
    )

    code, captured = _run(monkeypatch, capsys, ["tool", "future_tool", "--param", "flag=true"])

    assert code == 0
    assert "status: ok" in captured.out
    assert '"answer": 42' in captured.out


def _sample_arg_text(field):
    """A textual --param value that exercises one schema property (None means object)."""
    if isinstance(field.get("enum"), list) and field["enum"]:
        return str(field["enum"][0])
    for key in ("anyOf", "oneOf"):
        if isinstance(field.get(key), list):
            branches = field[key]
            for branch in branches:
                if branch.get("type") == "array":
                    return "worker-a"
            for branch in branches:
                if branch.get("type") != "null":
                    return _sample_arg_text(branch)
            return "null"
    value_type = field.get("type")
    if value_type == "boolean":
        return "true"
    if value_type in ("integer", "number"):
        return "1"
    if value_type == "array":
        return "worker-a"
    if value_type == "object":
        return None
    return "job-a"


def _schema_argv(schema):
    """Express every property through --param, or switch to --params-json when objects exist."""
    properties = schema.get("properties") or {}
    sample = {name: _sample_arg_text(field) for name, field in properties.items()}
    if any(value is None for value in sample.values()):
        complete = {name: value if value is not None else {} for name, value in sample.items()}
        return ["--params-json", json.dumps(complete)]
    argv = []
    for name, value in sample.items():
        argv += ["--param", f"{name}={value}"]
    return argv


def test_portforward_stops_process_before_reading_stderr(monkeypatch):
    class FakeProcess:
        stopped = False

        def __init__(self):
            self.stderr = self

        def read(self):
            assert self.stopped, "stderr must not be read while port-forward is still alive"
            return "mock timeout"

        def terminate(self):
            self.stopped = True

        def wait(self, timeout):
            return 0

    process = FakeProcess()
    monkeypatch.setattr(client, "_kubectl", lambda: "kubectl")
    monkeypatch.setattr(client, "_free_port", lambda: 12345)
    monkeypatch.setattr(client, "_wait_port", lambda *_: False)
    monkeypatch.setattr(client.subprocess, "Popen", lambda *args, **kwargs: process)

    with pytest.raises(client.TransportError):
        with client.PortForward(None):
            pass
    assert process.stopped


@pytest.mark.parametrize("scenario", ["success", "missing", "empty", "no_kubectl", "render_failed", "apply_failed"])
def test_collect_manifest_matches_ascend_diag(scenario, monkeypatch, capsys, tmp_path):
    legacy = _load("kubectl_ascend_diag_manifest", "kubectl-ascend_diag")
    manifest = tmp_path / "collect_manifest.yaml"
    if scenario != "missing":
        manifest.write_text("" if scenario == "empty" else "version: 3\n", encoding="utf-8")
    calls = []

    def fake_run(args, **kwargs):
        calls.append((args, kwargs))
        failed = (scenario == "render_failed" and "create" in args) or (scenario == "apply_failed" and "apply" in args)
        return subprocess.CompletedProcess(args, int(failed), "rendered manifest", "failure" if failed else "")

    monkeypatch.setattr(cli.subprocess, "run", fake_run)
    monkeypatch.setattr(cli, "_kubectl", lambda: None if scenario == "no_kubectl" else "kubectl")
    monkeypatch.setattr(legacy.shutil, "which", lambda name: None if scenario == "no_kubectl" else "kubectl")
    monkeypatch.setattr(cli, "_call", lambda *a, **k: pytest.fail("must not contact agent-core"))
    monkeypatch.setattr(sys, "argv", ["kubectl-ascend_diag", "--collect-manifest", str(manifest)])
    expected_code = legacy.main()
    expected_output = capsys.readouterr()
    expected_calls = list(calls)
    calls.clear()
    code, output = _run(monkeypatch, capsys, ["tool", "diagnose", "--param", f"collect_manifest={manifest}"])
    assert (code, output.out, output.err) == (expected_code, expected_output.out, expected_output.err)
    assert calls == expected_calls
