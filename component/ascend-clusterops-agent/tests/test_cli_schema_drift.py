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

"""CLI schema drift guard.

The plugin converts ``--param``/``--params-json`` through the catalog ``input_schema``
and the server validates with the pydantic model behind that schema. This module proves
the two agree: whenever the plugin accepts a value, the server model accepts the same
value with the same result, and whenever the plugin rejects locally the model rejects it
too (the plugin is never stricter or laxer in a way that changes the answer).
"""

from __future__ import annotations

import importlib.util
import sys
from importlib.machinery import SourceFileLoader
from pathlib import Path

import pytest
from agent_core.tools.registry import TOOL_REGISTRY
from pydantic import ValidationError

PLUGIN_DIR = Path(__file__).resolve().parents[1] / "kubectl-plugin"
sys.path.insert(0, str(PLUGIN_DIR))


def _load(name: str, filename: str):
    # the CLI script has no .py suffix, so the loader must be given explicitly
    loader = SourceFileLoader(name, str(PLUGIN_DIR / filename))
    spec = importlib.util.spec_from_loader(name, loader)
    module = importlib.util.module_from_spec(spec)
    sys.modules.setdefault(name, module)
    loader.exec_module(module)
    return module


_load("clusterops_query_client", "clusterops_query_client.py")
cli = _load("kubectl_clusterops_drift", "kubectl-clusterops")

NODE_TOOLS = ("faulty_device", "isolated_device", "cluster_resource")

ACCEPTED = [
    # (tool, --param values, --params-json, the equivalent complete request body)
    ("faulty_device", ["node=worker-a"], None, {"node": ["worker-a"]}),
    ("faulty_device", ["node=worker-a", "node=worker-b"], None, {"node": ["worker-a", "worker-b"]}),
    ("faulty_device", ["node=null"], None, {"node": None}),
    ("faulty_device", [], None, {}),
    ("faulty_device", [], '{"node": []}', {"node": []}),
    ("isolated_device", ["node=train-a-0"], None, {"node": ["train-a-0"]}),
    ("cluster_resource", ["node=worker.example.com"], None, {"node": ["worker.example.com"]}),
    ("task_status", ["job=train-a"], None, {"job": "train-a"}),
    ("task_status", ["job=train-a", "namespace=training"], None, {"job": "train-a", "namespace": "training"}),
    ("diagnose", ["job=train-a"], None, {"job": "train-a"}),
    ("diagnose", ["job=train-a", "refresh=true"], None, {"job": "train-a", "refresh": True}),
]

REJECTED = [
    # (tool, --param values, --params-json, the request body the server would also refuse)
    ("isolated_device", ["type=manual"], None, {"type": "manual"}),
    ("cluster_resource", ["resource_name=huawei.com/npu"], None, {"resource_name": "huawei.com/npu"}),
    ("faulty_device", ["bogus=1"], None, {"bogus": 1}),
    ("diagnose", ["refresh=yes", "job=train-a"], None, {"job": "train-a", "refresh": "yes"}),
    ("task_status", ["namespace=training"], None, {"namespace": "training"}),
    ("task_status", ["job=train-a", "namespace=BAD/NAMESPACE"], None, {"job": "train-a", "namespace": "BAD/NAMESPACE"}),
]


def _schema(tool: str):
    return TOOL_REGISTRY.get(tool).input_model.model_json_schema()


@pytest.mark.parametrize("tool,params,params_json,equivalent", ACCEPTED)
def test_plugin_conversion_matches_server_model(tool, params, params_json, equivalent):
    payload, error = cli._parse_tool_params(_schema(tool), params, params_json)
    assert error is None
    assert payload == equivalent

    model = TOOL_REGISTRY.get(tool).input_model
    assert model.model_validate(payload).model_dump() == model.model_validate(equivalent).model_dump()


@pytest.mark.parametrize("tool,params,params_json,server_body", REJECTED)
def test_local_rejection_matches_server_rejection(tool, params, params_json, server_body):
    payload, error = cli._parse_tool_params(_schema(tool), params, params_json)
    assert payload is None and error

    with pytest.raises(ValidationError):
        TOOL_REGISTRY.get(tool).input_model.model_validate(server_body)


@pytest.mark.parametrize("tool", NODE_TOOLS)
def test_node_array_and_empty_scope_keep_their_meaning(tool):
    schema = _schema(tool)
    assert cli._parse_tool_params(schema, [], '{"node": []}')[0] == {"node": []}
    assert cli._parse_tool_params(schema, ["node=null"], None)[0] == {"node": None}
    assert cli._parse_tool_params(schema, ["node=a", "node=b"], None)[0] == {"node": ["a", "b"]}


def test_plugin_uses_only_the_public_endpoints():
    source = (PLUGIN_DIR / "kubectl-clusterops").read_text(encoding="utf-8")
    assert '"/tools"' in source
    assert "/tools/{" in source
    assert "include_unavailable" not in source
