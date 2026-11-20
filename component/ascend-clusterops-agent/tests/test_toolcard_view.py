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

"""随包的工具卡解析器必须认得服务端真正发出的 tool_result 文本。

卡片不认识任何具体工具：能力调用是工具层信封 JSON，技能正文是 Markdown。这里在 node
里跑 `views/toolcard.js`，因此解析口径与前端视图由同一条门禁钉住（没有 node 就跳过）。
"""

# 这里的 node 脚手架与富文本用例同形，但两处刻意各自独立（分属不同 MR，跨文件共享会引入依赖）。
# pylint: disable=duplicate-code
from __future__ import annotations

import json
import os
import shutil
import subprocess
import tempfile

import pytest

from tests.fixtures.demo_adapter import demo_envelope

UI = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "agent-core", "webui")
TOOLCARD = os.path.join(UI, "views", "toolcard.js")
NODE = shutil.which("node")

NODE_TIMEOUT_S = 60  # node 卡死时用例以超时失败，不拖死整个测试作业
STDERR_TAIL_CHARS = 2000  # 失败信息里回显的 stderr 尾部长度
TOOLS = ("faulty_device", "isolated_device", "cluster_resource", "task_status", "diagnose")

# 技能正文就是 load_skill 的返回：Markdown，不是 JSON。
SKILL_BODY = "# 隔离设备查询\n\n## 什么时候用\n\n- 列出 ClusterD 手工隔离名单上的设备。"

RUNNER = """
import { readFileSync } from "fs";

async function main() {
  const mod = await import(process.argv[2]);
  const spec = JSON.parse(readFileSync(process.argv[3], "utf-8"));
  process.stdout.write(JSON.stringify({
    parsed: spec.texts.map((text) => mod.parseEnvelope(text)),
    statuses: spec.statuses.map((s) => mod.statusText(s)),
    tones: spec.tones.map((s) => mod.statusTone(s)),
    ms: spec.ms.map((v) => mod.formatMs(v)),
  }));
}

main().catch((error) => {
  process.stderr.write(String((error && error.stack) || error));
  process.exit(1);
});
"""


def _run(spec):
    """对随包 toolcard.js 跑一次 runner，返回它的 JSON 输出。"""
    with open(TOOLCARD, encoding="utf-8") as handle:
        source = handle.read()
    with tempfile.TemporaryDirectory() as tmp:
        module_path = os.path.join(tmp, "toolcard.mjs")  # .mjs：让 node 按模块解析
        runner_path = os.path.join(tmp, "runner.mjs")
        spec_path = os.path.join(tmp, "spec.json")
        with open(module_path, "w", encoding="utf-8") as handle:
            handle.write(source)
        with open(runner_path, "w", encoding="utf-8") as handle:
            handle.write(RUNNER)
        with open(spec_path, "w", encoding="utf-8") as handle:
            json.dump(spec, handle, ensure_ascii=False)
        completed = subprocess.run(
            [NODE, runner_path, "file://" + module_path, spec_path],
            capture_output=True,
            text=True,
            check=False,
            encoding="utf-8",
            timeout=NODE_TIMEOUT_S,
        )
    assert completed.returncode == 0, "node 跑不动卡片模块: " + completed.stderr[-STDERR_TAIL_CHARS:]
    return json.loads(completed.stdout)


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端解析门禁跳过")
@pytest.mark.parametrize("tool", TOOLS)
def test_card_parses_every_tool_envelope(tool):
    envelope = demo_envelope(tool)
    parsed = _run({"texts": [json.dumps(envelope, ensure_ascii=False)], "statuses": [], "tones": [], "ms": []})[
        "parsed"
    ][0]
    assert parsed["envelope"] == envelope
    assert parsed["tool"] == tool and parsed["status"] == "ok"
    assert parsed["requestId"] == envelope["request_id"]
    assert json.loads(parsed["show"]) == envelope  # 展示用的是美化后的同一份信封
    assert "\n  \"" in parsed["show"]


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端解析门禁跳过")
def test_card_flags_a_failed_envelope():
    envelope = demo_envelope(
        "diagnose", status="error", errors=[{"code": "INTERNAL_ERROR", "message": "training/inference task not found"}]
    )
    parsed = _run({"texts": [json.dumps(envelope, ensure_ascii=False)], "statuses": [], "tones": [], "ms": []})[
        "parsed"
    ][0]
    assert parsed["status"] == "error"


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端解析门禁跳过")
def test_card_leaves_a_skill_body_alone():
    parsed = _run({"texts": [SKILL_BODY], "statuses": [], "tones": [], "ms": []})["parsed"][0]
    assert parsed["envelope"] is None
    assert parsed["show"] == SKILL_BODY and parsed["raw"] == SKILL_BODY


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端解析门禁跳过")
def test_card_survives_an_empty_or_missing_output():
    parsed = _run({"texts": ["", None], "statuses": [], "tones": [], "ms": []})["parsed"]
    assert [item["envelope"] for item in parsed] == [None, None]
    assert [item["raw"] for item in parsed] == ["", ""]


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端解析门禁跳过")
def test_helper_maps_and_fallbacks_are_pinned():
    """STATUS_TABLE 是 null 原型查表：constructor 一类原型链键必须与未知键同值回退；
    statusTone 是三值映射；formatMs 对非数值/负值回空串。
    """
    out = _run(
        {
            "texts": [],
            "statuses": ["constructor", "hasownproperty", "__no_such_status__", "ok", "partial"],
            "tones": ["ok", "partial", "error", "whatever"],
            "ms": [1500, None, -5],
        }
    )
    assert out["statuses"][0] == out["statuses"][2] and out["statuses"][1] == out["statuses"][2]
    assert out["statuses"][3:] == ["成功", "部分"]
    assert out["tones"] == ["ok", "warn", "error", ""]
    assert out["ms"] == ["1.5s", "0ms", ""]
