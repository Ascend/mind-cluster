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

"""技能宿主：清单注入系统提示词，正文按需读取（渐进披露）。

唯一来源是随包的 skills/ 目录——启动时扫一次，不读环境变量、不热重扫。坏技能记 warning
后跳过，不影响启动；正文读取有字节上限，超限截断并声明。
"""

from __future__ import annotations

import logging
import re
from dataclasses import dataclass
from pathlib import Path
from typing import Any, List, Optional, Tuple, Union

import yaml

logger = logging.getLogger(__name__)

SKILL_FILE_NAME = "SKILL.md"
DEFAULT_MAX_BODY_BYTES = 8192  # 单份正文上限：一份失控的技能不能吃掉模型上下文
# 清单每技能固定一行且注入系统提示词：多行描述会伪造提示词段落，无界描述会挤爆上下文。
LISTING_DESCRIPTION_MAX_CHARS = 200
DEFAULT_SKILLS_ROOT = Path(__file__).resolve().parent.parent / "skills"

# 技能名与 allowed-tools 条目同一形状：有界的标识符（小写字母开头）。
_SAFE_IDENTIFIER_RE = re.compile(r"^[a-z][a-z0-9_-]{0,63}$")
_REQUIRED_FIELDS = ("name", "description")
_FRONTMATTER_DELIM = "---"
_TRUNCATION_NOTE = "\n\n（技能正文超过 {limit} 字节上限，已截断。）"


@dataclass(frozen=True)
class SkillCard:
    """一份通过校验的技能：frontmatter 字段、文件路径与清单展示用的相对位置。"""

    name: str
    description: str
    path: Path
    allowed_tools: Optional[Tuple[str, ...]] = None
    rel_path: str = ""


def _parse_skill_file(path: Path) -> Optional[SkillCard]:
    """解析一份 SKILL.md；不合法就记 warning 返回 None，绝不抛异常。"""
    skill_dir = path.parent.name
    try:
        text = path.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError) as exc:
        logger.warning("skill %s skipped: SKILL.md unreadable (%s)", skill_dir, exc)
        return None
    lines = text.splitlines()
    if not lines or lines[0].strip() != _FRONTMATTER_DELIM:
        logger.warning("skill %s skipped: frontmatter must start with a --- line", skill_dir)
        return None
    closing = next((index for index in range(1, len(lines)) if lines[index].strip() == _FRONTMATTER_DELIM), None)
    if closing is None:
        logger.warning("skill %s skipped: frontmatter has no closing --- line", skill_dir)
        return None
    try:
        meta = yaml.safe_load("\n".join(lines[1:closing]))
    except yaml.YAMLError as exc:
        logger.warning("skill %s skipped: frontmatter is not valid YAML (%s)", skill_dir, exc)
        return None
    if not isinstance(meta, dict):
        logger.warning("skill %s skipped: frontmatter is not a mapping", skill_dir)
        return None
    fields = {}
    for key in _REQUIRED_FIELDS:
        value = meta.get(key)
        if isinstance(value, str) and value.strip():
            fields[key] = value.strip()
        else:
            logger.warning("skill %s skipped: frontmatter field %r is missing or empty", skill_dir, key)
            return None
    name = fields["name"]
    if not _SAFE_IDENTIFIER_RE.match(name):
        logger.warning("skill %s skipped: name %r is not a safe identifier (lowercase snake/kebab)", skill_dir, name)
        return None
    return SkillCard(
        name=name,
        description=fields["description"],
        path=path,
        allowed_tools=_parse_allowed_tools(meta.get("allowed-tools"), skill_dir),
        rel_path=f"skills/{skill_dir}",
    )


def _parse_allowed_tools(value: Any, skill_dir: str) -> Optional[Tuple[str, ...]]:
    """解析可选的 allowed-tools 路由提示：非法条目丢弃并记 warning，不与注册表比对（提示不鉴权）。"""
    if value is None:
        return None
    if not isinstance(value, list):
        logger.warning("skill %s: allowed-tools is not a list, the hint is ignored", skill_dir)
        return None
    kept: List[str] = []
    for item in value:
        if not isinstance(item, str) or not item.strip():
            logger.warning("skill %s: allowed-tools drops a non-string entry %r", skill_dir, item)
            continue
        tool_name = item.strip()
        if not _SAFE_IDENTIFIER_RE.match(tool_name):
            logger.warning("skill %s: allowed-tools drops the unsafe name %r", skill_dir, tool_name)
            continue
        if tool_name not in kept:
            kept.append(tool_name)
    return tuple(kept) or None


def _body_from_text(text: str) -> str:
    """取结束定界符之后的正文；没有 frontmatter 时返回空串。"""
    lines = text.splitlines()
    if lines and lines[0].strip() == _FRONTMATTER_DELIM:
        for index in range(1, len(lines)):
            if lines[index].strip() == _FRONTMATTER_DELIM:
                return "\n".join(lines[index + 1 :]).strip()
    return ""


def _listing_description(text: str, limit: int = LISTING_DESCRIPTION_MAX_CHARS) -> str:
    """把描述压成一行并有界（清单是一技能一行）。"""
    collapsed = " ".join(text.split())
    return collapsed if len(collapsed) <= limit else collapsed[: limit - 1] + "…"


def _truncate_utf8(text: str, limit: int) -> str:
    """按字节上限裁剪且不切碎 UTF-8 字符，裁了就补一句声明。"""
    encoded = text.encode("utf-8")
    if len(encoded) <= limit:
        return text
    return encoded[:limit].decode("utf-8", errors="ignore") + _TRUNCATION_NOTE.format(limit=limit)


class SkillLoader:
    """渐进披露的技能宿主：构造时扫一次清单并打摘要日志，之后只读内存。"""

    def __init__(
        self,
        builtin_root: Optional[Union[Path, str]] = None,
        max_body_bytes: int = DEFAULT_MAX_BODY_BYTES,
    ) -> None:
        self._builtin_root = Path(builtin_root) if builtin_root is not None else DEFAULT_SKILLS_ROOT
        self._max_body_bytes = int(max_body_bytes)
        self._skipped = 0
        self._cards: List[SkillCard] = self._scan()
        self._log_summary()

    def _scan(self) -> List[SkillCard]:
        cards: List[SkillCard] = []
        seen = set()
        if not self._builtin_root.is_dir():
            logger.warning("skills tree %s is absent: the listing stays empty", self._builtin_root)
            return cards
        for child in sorted(self._builtin_root.iterdir()):
            skill_file = child / SKILL_FILE_NAME
            if not child.is_dir() or not skill_file.is_file():
                continue
            card = _parse_skill_file(skill_file)
            if card is None:
                self._skipped += 1  # 已经记过 warning：坏技能不阻断启动
                continue
            if card.name in seen:
                logger.warning("skill %s redeclared under %s: the first declaration wins", card.name, child)
                self._skipped += 1
                continue
            seen.add(card.name)
            cards.append(card)
        return cards

    def _log_summary(self) -> None:
        """一行汇总：清单里有什么、跳过了几份。"""
        loaded = [card.name for card in self._cards]
        logger.info(
            "skill host scan summary: loaded=%d [%s] skipped=%d",
            len(loaded),
            ", ".join(loaded) or "-",
            self._skipped,
        )

    def names(self) -> List[str]:
        """清单中的技能名，按目录名排序。"""
        return [card.name for card in self._cards]

    def prompt_lines(self) -> List[str]:
        """系统提示词里的技能清单：每份一行 `name: description（工具：…）（位于 …）`。"""
        lines: List[str] = []
        for card in self._cards:
            line = f"{card.name}: {_listing_description(card.description)}"
            if card.allowed_tools:
                line += f"（工具：{'、'.join(card.allowed_tools)}）"
            line += f"（位于 {card.rel_path}）"
            lines.append(line)
        return lines

    def read(self, name: str) -> str:
        """一份技能正文（有界）；未知名字返回可用清单。"""
        card = next((item for item in self._cards if item.name == name), None)
        if card is None:
            return self._unknown_listing(name)
        logger.info("skill body loaded: %s (source: %s)", card.name, card.path)
        try:
            text = card.path.read_text(encoding="utf-8")
        except (OSError, UnicodeDecodeError) as exc:
            logger.warning("skill %s unreadable: %s", name, exc)
            return f"技能 {name} 暂时无法读取（{exc.__class__.__name__}），请基于已有信息回答。"
        body = _body_from_text(text)
        return _truncate_utf8(body, self._max_body_bytes) if body else f"技能 {name} 暂无正文。"

    def _unknown_listing(self, name: str) -> str:
        """未知技能的答复：把清单本身给模型。"""
        lines = self.prompt_lines()
        listing = "\n".join("- " + line for line in lines) if lines else "（当前没有可用技能。）"
        return f"未知技能：{name}。当前可用技能清单：\n{listing}\n需要正文时请用清单中的名称调用 load_skill。"
