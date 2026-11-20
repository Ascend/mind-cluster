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

"""Packaging, deployment manifest and user guide contract.

With an explicit packages list setuptools does not pick up subpackages on its own, so a
wheel can import-fail on agent_core.adapters/api/chat/skill inside the container. These tests
pin the subpackage list and the package-data patterns that carry the WebUI (Web User
Interface, 网页用户界面) and the built-in skills, the env the deployment manifest documents
(and the capabilities it must not advertise), the user guide with its menu/README links, and
- by building the wheel once - the zipfile listing with webui and skills inside.
"""

# pylint: disable=redefined-outer-name  # pytest fixture names double as test params
from __future__ import annotations

import re
import shutil
import subprocess
import sys
import zipfile
from pathlib import Path
from typing import List

import pytest

COMPONENT_ROOT = Path(__file__).resolve().parent.parent
AGENT_CORE = COMPONENT_ROOT / "agent-core"
PYPROJECT = AGENT_CORE / "pyproject.toml"
MANIFEST = COMPONENT_ROOT / "build" / "agent-core.yaml"
GUIDE = (
    Path(__file__).resolve().parents[3]
    / "docs"
    / "zh"
    / "scheduling"
    / "04_usage"
    / "14_clusterops_agent"
    / "06_using_webui.md"
)
MENU = GUIDE.parent / "menu_fault_diagnosis.md"
README = COMPONENT_ROOT / "README.md"
API_DOC = Path(__file__).resolve().parents[3] / "docs" / "zh" / "scheduling" / "06_api" / "20_clusterops_agent.md"

# The WebUI Service port contract: the node port stays put, the target port serves the app.
NODE_PORT = 30443
TARGET_PORT = 9700

SUBPACKAGES = (
    "agent_core",
    "agent_core.adapters",
    "agent_core.api",
    "agent_core.chat",
    "clusterops_common",
    "diagproto",
)
WEBUI_FILES = (
    "webui/index.html",
    "webui/style.css",
    "webui/app.js",
    "webui/sse.js",
    "webui/views/chat.js",
    "webui/views/toolcard.js",
    "webui/views/richtext.js",
)
PACKAGE_DATA_PATTERNS = (
    "webui/*.html",
    "webui/*.css",
    "webui/*.js",
    "webui/views/*.js",
    "skills/*/SKILL.md",
    "skills/README.md",
)

# The documents describing what ships with the agent; none may promise on-site skill injection.
SKILL_CLAIM_DOCS = (
    ("agent-core.yaml", MANIFEST),
    ("skills README", AGENT_CORE / "skills" / "README.md"),
    ("user guide", GUIDE),
)


def _toml_text() -> str:
    return PYPROJECT.read_text(encoding="utf-8")


def test_pyproject_lists_every_agent_core_subpackage():
    text = _toml_text()
    start = text.index("packages = [")
    end = text.index("]", start)
    declared = re.findall(r'"([^"]+)"', text[start:end])
    for package in SUBPACKAGES:
        assert package in declared, f"pyproject packages lacks {package}: import fails in the container"


def test_pyproject_package_data_covers_webui_and_skills():
    text = _toml_text()
    start = text.index("package-data")
    end = text.index("]", text.index("]", start) + 1)  # close of the patterns list
    block = text[start:end]
    for pattern in PACKAGE_DATA_PATTERNS:
        assert f'"{pattern}"' in block, f"package-data lacks {pattern}: static files never reach the wheel"


def test_declared_static_files_exist_in_the_source_tree():
    for name in WEBUI_FILES:
        assert (AGENT_CORE / name).is_file(), f"missing static file: {name}"
    assert (AGENT_CORE / "skills" / "README.md").is_file()
    assert (AGENT_CORE / "skills").is_dir(), "the skills tree is missing from the package root"
    for skill_dir in (AGENT_CORE / "skills").iterdir():
        if skill_dir.is_dir():
            assert (skill_dir / "SKILL.md").is_file(), f"skill without SKILL.md: {skill_dir}"


def test_deployment_manifest_documents_the_optional_env():
    yaml_text = MANIFEST.read_text(encoding="utf-8")
    for env_name in ("AGENT_WEBUI_ACCESS", "AGENT_WEBUI_PORT"):
        assert env_name in yaml_text, f"agent-core.yaml does not document {env_name}"
    # The tool layer is called in process and routing has one mode: no capability gateway,
    # no routing switch.
    for retired in ("CAPABILITIES_MODE", "CAPABILITIES_BASE", "AGENT_MODE"):
        assert retired not in yaml_text, f"agent-core.yaml still offers the retired {retired}"


@pytest.mark.parametrize("label, path", SKILL_CLAIM_DOCS)
def test_docs_promise_no_on_site_skill_injection(label: str, path: Path):
    """Skills ship inside the wheel: no document may advertise a site directory or a reload."""
    text = path.read_text(encoding="utf-8")
    assert "AGENT_SKILLS_DIR" not in text, f"{label} must not advertise the reserved AGENT_SKILLS_DIR"
    for claim in ("热加载", "热更新", "预留能力", "自动重扫", "无需重启或重打镜像"):
        assert claim not in text, f"{label} must not promise skill {claim}"


def test_deployment_manifest_carries_the_webui_nodeport():
    """The WebUI Service keeps its node port and serves the app on the target port."""
    import yaml

    docs = list(yaml.safe_load_all(MANIFEST.read_text(encoding="utf-8")))
    services = {d["metadata"]["name"]: d for d in docs if d and d.get("kind") == "Service"}
    assert "agent-core-ui" in services, "the WebUI Service is missing"
    spec = services["agent-core-ui"]["spec"]
    assert spec["type"] == "NodePort"
    port = spec["ports"][0]
    assert port["nodePort"] == NODE_PORT
    assert port["targetPort"] == port["port"] == TARGET_PORT


def test_user_guide_documents_the_webui_nodeport():
    text = GUIDE.read_text(encoding="utf-8")
    assert str(NODE_PORT) in text, "the guide must document the node port"
    assert "agent-core-ui" in text, "the guide must name the Service"


def test_docs_do_not_retire_the_nodeport():
    """The node port never retires: no document may claim it goes away."""
    for path in (API_DOC, GUIDE, MANIFEST):
        for line in path.read_text(encoding="utf-8").splitlines():
            # the documented promise is worded "30443 不下线"; any other line pairing the node
            # port with 下线 claims a retirement
            claim = line.replace("不会下线", "").replace("不下线", "")
            assert not (str(NODE_PORT) in claim and "下线" in claim), f"{path.name} retires the node port: {line}"


def test_user_guide_covers_access_chat_and_guide():
    text = GUIDE.read_text(encoding="utf-8")
    assert "http" in text and "/ui/" in text, "the guide must explain how to reach the page"
    # The page shows one fixed sentence; the guide quotes it and points at the configuration
    # procedure instead of carrying an in-page kubectl block.
    assert "未配置大模型，对话不可用；请配置模型后使用" in text, "the guide must quote the page sentence"
    assert "kubectl clusterops --create-llm-config" in text, "the guide must point at the config command"
    assert "04_configuring_llm.md" in text, "the guide must link the LLM configuration page"
    for slash in ("/faults", "/isolated", "/diagnose"):
        assert slash not in text, f"the guide still mentions the retired {slash} slash command"


def test_menu_and_readme_link_the_user_guide():
    menu = MENU.read_text(encoding="utf-8")
    assert "06_using_webui.md" in menu, "the feature menu lacks the WebUI guide entry"
    readme = README.read_text(encoding="utf-8")
    assert "06_using_webui.md" in readme or "14_clusterops_agent" in readme, (
        "the component README must link the user guide"
    )


@pytest.fixture(scope="module")
def wheel_path(tmp_path_factory) -> Path:
    """Build the agent-core wheel once (offline: --no-deps --no-build-isolation).

    在一次性拷贝里构建：pip wheel 会在源码树原地留 build/ 与 *.egg-info，拷贝到
    tmp 既能保持检出干净，也避免并发跑同一份源码时相互覆盖中间产物。agent-core 与
    common 两棵一起拷：package-dir 用 ../common 相对路径内嵌共享代码与 gRPC 桩。
    """
    if sys.version_info < (3, 10):
        pytest.skip("pyproject 声明 requires-python >= 3.10：py3.9 上 pip wheel 不出包，打包门禁跳过")
    out = tmp_path_factory.mktemp("wheel")
    src = tmp_path_factory.mktemp("src")
    shutil.copytree(AGENT_CORE, src / "agent-core")
    shutil.copytree(AGENT_CORE.parent / "common", src / "common")
    proc = subprocess.run(
        [
            sys.executable,
            "-m",
            "pip",
            "wheel",
            "--no-deps",
            "--no-build-isolation",
            "--wheel-dir",
            str(out),
            str(src / "agent-core"),  # absolute path: a bare name would resolve against the index
        ],
        check=False,
        capture_output=True,
        timeout=300,
    )
    assert proc.returncode == 0, "wheel build failed:\n" + proc.stderr.decode(errors="replace")[-2000:]
    wheels: List[Path] = sorted(out.glob("agent_core-*.whl"))
    assert len(wheels) == 1, f"expected exactly one agent_core wheel, got {wheels}"
    return wheels[0]


def test_wheel_includes_subpackages_and_static_data(wheel_path: Path):
    with zipfile.ZipFile(wheel_path) as wheel:
        names = set(wheel.namelist())
    for module in (
        "agent_core/adapters/tool_adapter.py",
        "agent_core/api/stream_api.py",
        "agent_core/chat/tool_control.py",
        "agent_core/chat/events.py",
        "agent_core/chat/graph.py",
        "agent_core/chat/skill_loader.py",
        "agent_core/chat/tool_control.py",
    ):
        assert module in names, f"subpackage module missing from the wheel: {module}"
    for name in WEBUI_FILES:
        assert f"agent_core/{name}" in names, f"webui file missing from the wheel: {name}"
    assert "agent_core/skills/README.md" in names
    # every built-in skill must reach the wheel with its SKILL.md
    for skill_dir in (AGENT_CORE / "skills").iterdir():
        if skill_dir.is_dir():
            assert f"agent_core/skills/{skill_dir.name}/SKILL.md" in names


def test_wheel_keeps_tests_out(wheel_path: Path):
    with zipfile.ZipFile(wheel_path) as wheel:
        names = wheel.namelist()
    assert not [name for name in names if name.startswith("tests/") or "/__pycache__/" in name]
