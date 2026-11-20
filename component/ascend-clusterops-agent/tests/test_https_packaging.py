#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""HTTPS deployment contracts, independent of the optional chat modules."""

from pathlib import Path
import yaml

ROOT = Path(__file__).resolve().parents[1]


def resources():
    return list(yaml.safe_load_all((ROOT / "build/agent-core.yaml").read_text()))


def test_dedicated_nodeport_keeps_cli_service_internal():
    services = {v["metadata"]["name"]: v for v in resources() if v["kind"] == "Service"}
    cli = services["agent-core"]["spec"]
    assert cli.get("type", "ClusterIP") == "ClusterIP"
    assert [(v["port"], v["targetPort"]) for v in cli["ports"]] == [(9700, 9700), (9710, 9710)]
    ui = services["agent-core-ui"]["spec"]
    assert ui["type"] == "NodePort"
    assert [(v["port"], v["targetPort"], v["nodePort"]) for v in ui["ports"]] == [(9743, 9743, 30443)]


def test_secrets_use_api_and_least_privilege_bootstrap():
    docs = resources()
    deployment = next(v for v in docs if v["kind"] == "Deployment")
    pod = deployment["spec"]["template"]["spec"]
    assert all("secret" not in v and "projected" not in v for v in pod["volumes"])
    container = pod["containers"][0]
    assert 9743 in [v["containerPort"] for v in container["ports"]]
    assert all("secretKeyRef" not in v.get("valueFrom", {}) for v in container["env"])
    role = next(v for v in docs if v["kind"] == "Role" and v["metadata"]["name"] == "agent-core-webui-secrets")
    rules = role["rules"]
    assert next(v for v in rules if v["verbs"] == ["get"])["resourceNames"] == [
        "agent-core-webui-tls",
        "agent-core-webui-token",
    ]
    assert next(v for v in rules if v["verbs"] == ["patch"])["resourceNames"] == ["agent-core-webui-tls"]
    assert next(v for v in rules if v["verbs"] == ["create"]).get("resourceNames") is None
    assert all("update" not in v["verbs"] and "delete" not in v["verbs"] for v in rules)


def test_https_env_injects_port_namespace_and_host_ip():
    deployment = next(v for v in resources() if v["kind"] == "Deployment")
    container = deployment["spec"]["template"]["spec"]["containers"][0]
    env = {item["name"]: item for item in container["env"]}
    assert env["WEBUI_HTTPS_PORT"]["value"] == "9743"
    assert env["POD_NAMESPACE"]["valueFrom"]["fieldRef"]["fieldPath"] == "metadata.namespace"
    assert env["HOST_IP"]["valueFrom"]["fieldRef"]["fieldPath"] == "status.hostIP"
    assert env["WEBUI_TLS_SANS"]["value"] == ""
    assert env["WEBUI_RATELIMIT_MAX_REQUESTS"]["value"] == "600"
    assert env["WEBUI_RATELIMIT_WINDOW_SECONDS"]["value"] == "60"


def test_package_contains_tls_dependency_and_entrypoint():
    project = (ROOT / "agent-core/pyproject.toml").read_text()
    assert '"cryptography>=43.0"' in project
    assert 'clusterops-agent = "agent_core.serve:main"' in project
    docker = (ROOT / "build/Dockerfile").read_text()
    assert "exec python -m agent_core.serve" in docker
    assert "9743" in next(v for v in docker.splitlines() if v.startswith("EXPOSE "))
    assert "cryptography>=43.0" in (ROOT / "requirements.txt").read_text()
