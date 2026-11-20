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

"""自研答案渲染器必须认得回答里真正出现的语法，且绝不写 HTML。

在 node 里跑 `views/richtext.js`：垫一份最小 DOM，把渲染出的节点树拉回 Python 断言结构，
因此钉住的是真实渲染结果（没有 node 就跳过）。
"""

# 这里的 node 脚手架与工具卡用例同形，但两处刻意各自独立（分属不同 MR，跨文件共享会引入依赖）。
# pylint: disable=duplicate-code
from __future__ import annotations

import json
import os
import shutil
import subprocess
import tempfile

import pytest

UI = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "agent-core", "webui")
RICHTEXT = os.path.join(UI, "views", "richtext.js")
NODE = shutil.which("node")

NODE_TIMEOUT_S = 60  # node 卡死时用例以超时失败，不拖死整个测试作业
STDERR_TAIL_CHARS = 2000  # 失败信息里回显的 stderr 尾部长度
TEXT_TAG = "#text"

RUNNER = """
import { readFileSync } from "fs";

function makeElement(tag) {
  return {
    tagName: tag,
    className: "",
    children: [],
    _text: "",
    appendChild(child) { this.children.push(child); return child; },
    set textContent(value) { this._text = String(value); this.children = []; },
    get textContent() {
      if (this.children.length) { return this.children.map((c) => c.textContent).join(""); }
      return this._text;
    },
  };
}

function toTree(node) {
  if (node.tagName === "#text") { return { tag: "#text", text: node.textContent }; }
  return { tag: node.tagName, text: node.textContent, children: node.children.map(toTree) };
}

async function main() {
  const mod = await import(process.argv[2]);
  globalThis.document = {
    createElement: (tag) => makeElement(tag),
    createTextNode: (value) => ({ tagName: "#text", children: [], get textContent() { return String(value); } }),
  };
  const spec = JSON.parse(readFileSync(process.argv[3], "utf-8"));
  process.stdout.write(JSON.stringify(spec.texts.map((text) => {
    const root = makeElement("div");
    mod.renderAnswer(root, text);
    return { tree: toTree(root), text: root.textContent };
  })));
}

main().catch((error) => {
  process.stderr.write(String((error && error.stack) || error));
  process.exit(1);
});
"""


def _render(*texts):
    """对随包 richtext.js 跑一次 runner，返回每段文本的渲染结果。"""
    with open(RICHTEXT, encoding="utf-8") as handle:
        source = handle.read()
    with tempfile.TemporaryDirectory() as tmp:
        module_path = os.path.join(tmp, "richtext.mjs")  # .mjs：让 node 按模块解析
        runner_path = os.path.join(tmp, "runner.mjs")
        spec_path = os.path.join(tmp, "spec.json")
        with open(module_path, "w", encoding="utf-8") as handle:
            handle.write(source)
        with open(runner_path, "w", encoding="utf-8") as handle:
            handle.write(RUNNER)
        with open(spec_path, "w", encoding="utf-8") as handle:
            json.dump({"texts": list(texts)}, handle, ensure_ascii=False)
        completed = subprocess.run(
            [NODE, runner_path, "file://" + module_path, spec_path],
            capture_output=True,
            text=True,
            check=False,
            encoding="utf-8",
            timeout=NODE_TIMEOUT_S,
        )
    assert completed.returncode == 0, "node 跑不动渲染器: " + completed.stderr[-STDERR_TAIL_CHARS:]
    return json.loads(completed.stdout)


def _tags(node):
    """树里出现过的标签名（含自身）。"""
    found = [node["tag"]]
    for child in node.get("children", []):
        found.extend(_tags(child))
    return found


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端渲染门禁跳过")
def test_pipe_table_becomes_a_real_table():
    rendered = _render(
        "| 设备 | 故障码 | 类型 |\n|:--|--:|:--:|\n| npu-01 | 200001002 | SubHealth |\n| npu-02 | 81078603 |"
    )[0]
    blocks = rendered["tree"]["children"]
    assert [block["tag"] for block in blocks] == ["table"]
    head, body = blocks[0]["children"]
    assert head["tag"] == "thead" and body["tag"] == "tbody"
    assert [cell["text"] for cell in head["children"][0]["children"]] == ["设备", "故障码", "类型"]
    rows = body["children"]
    assert [[cell["text"] for cell in row["children"]] for row in rows] == [
        ["npu-01", "200001002", "SubHealth"],
        ["npu-02", "81078603", ""],  # 缺列补空，表格不塌
    ]
    assert all(cell["tag"] == "td" for row in rows for cell in row["children"])


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端渲染门禁跳过")
def test_a_table_missing_its_separator_row_stays_text():
    """流式中途只有表头时按段落输出：不抛异常，也不吃掉原文。"""
    source = "| 设备 | 故障码 |\n| npu-01 | 200001002 |"
    rendered = _render(source)[0]
    assert "table" not in _tags(rendered["tree"])
    assert rendered["text"] == source


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端渲染门禁跳过")
def test_bullet_and_numbered_lines_become_lists():
    bullets, numbered = _render("- 卡 1 故障\n- 卡 2 隔离", "1. 查资源\n2. 查任务")
    assert [block["tag"] for block in bullets["tree"]["children"]] == ["ul"]
    assert [item["text"] for item in bullets["tree"]["children"][0]["children"]] == ["卡 1 故障", "卡 2 隔离"]
    assert [block["tag"] for block in numbered["tree"]["children"]] == ["ol"]
    assert [item["text"] for item in numbered["tree"]["children"][0]["children"]] == ["查资源", "查任务"]


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端渲染门禁跳过")
def test_bold_and_inline_code_are_the_only_inline_markup():
    rendered = _render("**结论**：卡 `npu-01` 状态 SubHealth，其余正常")[0]
    paragraph = rendered["tree"]["children"][0]
    assert paragraph["tag"] == "p"
    pieces = paragraph["children"]
    assert [piece["tag"] for piece in pieces] == ["strong", TEXT_TAG, "code", TEXT_TAG]
    assert [pieces[0]["text"], pieces[2]["text"]] == ["结论", "npu-01"]
    assert rendered["text"] == "结论：卡 npu-01 状态 SubHealth，其余正常"


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端渲染门禁跳过")
def test_headings_one_to_four_map_to_h_tags():
    rendered = _render("# 一\n## 二\n### 三\n#### 四")[0]
    blocks = rendered["tree"]["children"]
    assert [(block["tag"], block["text"]) for block in blocks] == [
        ("h1", "一"),
        ("h2", "二"),
        ("h3", "三"),
        ("h4", "四"),
    ]


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端渲染门禁跳过")
def test_unclosed_markup_stays_text():
    source = "未闭合的 **加粗 与 `行内码"
    rendered = _render(source)[0]
    assert "strong" not in _tags(rendered["tree"])
    assert "code" not in _tags(rendered["tree"])
    assert rendered["text"] == source


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端渲染门禁跳过")
def test_html_in_the_answer_stays_text():
    source = "设备 <img src=x onerror=alert(1)> 已隔离"
    rendered = _render(source)[0]
    assert "img" not in _tags(rendered["tree"])
    assert rendered["text"] == source


@pytest.mark.skipif(NODE is None, reason="node 不可用：前端渲染门禁跳过")
def test_plain_text_survives_verbatim_and_an_empty_answer_clears():
    plain, empty = _render("集群共 8 张卡，其中 3 张可用（故障码 81078603）。", "")
    assert [block["tag"] for block in plain["tree"]["children"]] == ["p"]
    assert plain["text"] == "集群共 8 张卡，其中 3 张可用（故障码 81078603）。"
    assert empty["tree"]["children"] == [] and empty["text"] == ""
