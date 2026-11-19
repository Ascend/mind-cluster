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

"""Decode read ConfigMaps into their JSON documents.

The ClusterD summaries share one envelope: one data key per shard whose value is a JSON
string (an empty string means a cleared shard). This module owns that decode contract so
each producer parser only maps fields.
"""

from __future__ import annotations

from typing import Any, Dict, Iterator, List, Tuple

from agent_core.datasources.utils import decode_json_string
from agent_core.query_schemas import QueryIssue, SourcePointer


def _document_value(item: Dict[str, Any], name: str) -> Any:
    """The shard value of one CM item; a single-key CM may carry it under another key."""
    data = item.get("data") or {}
    if name in data:
        return data.get(name)
    if len(data) == 1:
        return next(iter(data.values()))
    return None


def iter_cm_documents(
    items: List[Dict[str, Any]], issues: List[QueryIssue], source_id: str = ""
) -> Iterator[Tuple[str, SourcePointer, Dict[str, Any]]]:
    """Yield ``(name, source, document)`` for every readable CM in ``items``.

    A value that is not a JSON object is reported as ``SOURCE_INVALID`` and skipped, so
    callers never need to decode the same envelope themselves.
    """
    for item in items:
        name = item.get("name") or ""
        source = SourcePointer(source_id=item.get("source_id") or source_id, data_key=name, path=name)
        document = decode_json_string(_document_value(item, name), name, issues, source.source_id, allow_empty=True)
        if not isinstance(document, dict):
            if document is not None:
                issues.append(
                    QueryIssue(
                        code="SOURCE_INVALID", source_id=source.source_id, path=name, message="expected an object"
                    )
                )
            continue
        yield name, source, document
