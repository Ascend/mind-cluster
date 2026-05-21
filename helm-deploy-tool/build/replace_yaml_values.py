#!/usr/bin/env python3

# Perform  replace yaml values
# Copyright @ Huawei Technologies CO., Ltd. 2026. All rights reserved
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================
import os
import glob
import re
import argparse


CONTAINER_NAME_MAP = {
    "volcano-scheduler": "scheduler",
    "volcano-controllers": "controller",
}

REPLACE_RULES = [
    {
        "pattern": re.compile(r"^(\s*-?\s*)image:\s+([^:\s]+):(\S+)\s*$"),
        "replace": lambda m, vp, ver: (
            m.group(1)
            + 'image: {{ .Values.'
            + vp
            + 'image.repository | default "'
            + m.group(2)
            + '" }}:{{ .Values.'
            + vp
            + 'image.tag | default "'
            + (m.group(3) + "-" + ver if "volcano" in m.group(2) else ver)
            + '" }}'
        ),
    },
    {
        "pattern": re.compile(r'^(\s+)imagePullPolicy:\s*["\']?(\S+?)["\']?\s*$'),
        "replace": lambda m, vp, _ver: (
            m.group(1) + 'imagePullPolicy: {{ .Values.' + vp + 'image.pullPolicy | default "' + m.group(2) + '" }}'
        ),
    },
]


def find_values_prefix(result_lines, container_map):
    for line in reversed(result_lines):
        m = re.match(r"^\s*-?\s*name:\s+(\S+)", line)
        if m:
            name = m.group(1)
            if name in container_map:
                return container_map[name] + "."
    return ""


def replace_resources_block(lines, i, result_lines, container_map):
    line = lines[i]
    m = re.match(r"^(\s+)(limits|requests):\s*$", line)
    if not m or i + 2 >= len(lines):
        return None

    indent = m.group(1)
    kind = m.group(2)
    next1 = lines[i + 1]
    next2 = lines[i + 2]

    m1 = re.match(r"^(\s+)(cpu|memory):\s*(\S+)\s*$", next1)
    m2 = re.match(r"^(\s+)(cpu|memory):\s*(\S+)\s*$", next2)
    if not m1 or not m2 or m1.group(2) == m2.group(2):
        return None

    sub_indent = m1.group(1)
    vp = find_values_prefix(result_lines, container_map)

    replaced = [indent + kind + ":"]
    for mg in [m1, m2]:
        key = mg.group(2)
        val = mg.group(3)
        replaced.append(
            sub_indent + key + ': {{ .Values.' + vp + 'resources.' + kind + '.' + key + ' | default "' + val + '" }}'
        )
    return replaced


def collect_block_lines(lines, i, keyword):
    line = lines[i]
    if "{{" in line:
        return None
    m = re.match(r"^(\s+)" + keyword + r":\s*(.*)", line)
    if not m:
        return None

    indent = m.group(1)
    rest = m.group(2)
    collected = [line]

    if "]" in rest:
        return {"indent": indent, "lines": collected, "count": 1}

    j = i + 1
    while j < len(lines):
        collected.append(lines[j])
        if "]" in lines[j]:
            break
        j += 1

    return {"indent": indent, "lines": collected, "count": len(collected)}


def extract_block_content(block_lines, keyword):
    full = "\n".join(block_lines)
    m = re.match(r"^(\s+)" + keyword + r":\s*\[\s*(.*?)\s*\]\s*$", full, re.DOTALL)
    if not m:
        return None
    return m.group(2)


def replace_block(lines, i, result_lines, container_map, keyword):
    info = collect_block_lines(lines, i, keyword)
    if info is None:
        return None

    content = extract_block_content(info["lines"], keyword)
    if content is None:
        return None

    vp = find_values_prefix(result_lines, container_map)
    indent = info["indent"]

    original_block = "\n".join(info["lines"])

    replaced = [
        indent + "{{- if .Values." + vp + keyword + " }}",
        indent + keyword + ": {{ .Values." + vp + keyword + " }}",
        indent + "{{- else }}",
        original_block,
        indent + "{{- end }}",
    ]

    return {"replaced": replaced, "count": info["count"]}


def has_following_args(lines, i, keyword_indent):
    j = i + 1
    while j < len(lines):
        line = lines[j]
        stripped = line.lstrip()
        if not stripped or stripped.startswith("#"):
            j += 1
            continue
        current_indent = len(line) - len(stripped)
        if current_indent < keyword_indent:
            return False
        if re.match(r"^\s*args:\s", line) or re.match(r"^\s*\{\{-\s*if\s+\.Values.*args\s*\}\}", stripped):
            return True
        break
    return False


def skip_template_block(lines, i):
    line = lines[i].strip()
    if not re.match(r"^\{\{-\s*if\s+\.Values\.(\w+\.)*(args|command)\s*\}\}$", line):
        return None

    depth = 1
    j = i + 1
    while j < len(lines) and depth > 0:
        line_content = lines[j].strip()
        if re.match(r"^\{\{-\s*if\s+", line_content):
            depth += 1
        elif re.match(r"^\{\{-\s*end\s*\}\}$", line_content):
            depth -= 1
        j += 1

    return j - 1


def process_file(file_path, container_map=None, version=None):
    if container_map is None:
        container_map = CONTAINER_NAME_MAP

    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    lines = content.split("\n")
    result = []
    i = 0

    while i < len(lines):
        line = lines[i]

        skip_end = skip_template_block(lines, i)
        if skip_end is not None:
            while i <= skip_end:
                result.append(lines[i])
                i += 1
            continue

        replaced = False
        for rule in REPLACE_RULES:
            m = rule["pattern"].match(line)
            if m:
                vp = find_values_prefix(result, container_map)
                result.append(rule["replace"](m, vp, version))
                i += 1
                replaced = True
                break

        if replaced:
            continue

        res_block = replace_resources_block(lines, i, result, container_map)
        if res_block:
            result.extend(res_block)
            i += 3
            continue

        args_result = replace_block(lines, i, result, container_map, "args")
        if args_result:
            result.extend(args_result["replaced"])
            i += args_result["count"]
            continue

        cmd_result = replace_block(lines, i, result, container_map, "command")
        if cmd_result:
            cmd_indent = len(lines[i]) - len(lines[i].lstrip())
            if not has_following_args(lines, i, cmd_indent):
                result.extend(cmd_result["replaced"])
                i += cmd_result["count"]
                continue

        result.append(line)
        i += 1

    with open(file_path, "w", encoding="utf-8") as f:
        f.write("\n".join(result))


def process_charts_dir(charts_path, version=None):
    for chart_dir in sorted(glob.glob(os.path.join(charts_path, "*"))):
        if not os.path.isdir(chart_dir):
            continue
        chart_name = os.path.basename(chart_dir)
        for yaml_file in sorted(glob.glob(os.path.join(chart_dir, "yamls", "**", "*.yaml"), recursive=True)):
            print(f"Processing [{chart_name}]: {yaml_file}")
            process_file(yaml_file, version=version)
    print("Done.")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Replace YAML values with Helm template placeholders")
    parser.add_argument("-v", "--version", help="Version tag to use as default for image.tag")
    parser.add_argument("path", nargs="?", help="Chart directory or YAML file path")
    args = parser.parse_args()

    script_dir = os.path.dirname(os.path.abspath(__file__))
    charts_dir = os.path.join(script_dir, "../app/charts")

    if args.path:
        if os.path.isdir(args.path):
            charts_dir = args.path
            process_charts_dir(charts_dir, version=args.version)
        else:
            process_file(args.path, version=args.version)
    else:
        process_charts_dir(charts_dir, version=args.version)
