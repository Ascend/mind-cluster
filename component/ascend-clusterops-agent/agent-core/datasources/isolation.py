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

"""Isolation parsing for ``clusterd-manual-info-cm``, the only isolation source.

``data[node]`` decodes to ``{"Total": [device], "Detail": {device: [reason]}}``. ``Total``
is the single roster of isolated devices; ``Detail`` only supplements roster members with
fault level, fault code and last separate time. Detail-only entries never expand the
roster, and the DP ConfigMap is never read for isolation facts.
"""

from __future__ import annotations

from typing import Any, Dict, List, Optional, Set, Tuple

from agent_core.datasources.utils import (
    decode_json_string,
    epoch_seconds,
    join_code,
    positive_number,
    timestamp_value_inferred,
)
from agent_core.query_schemas import IsolatedDeviceRecord, NodeIsolation, QueryIssue

# The producer documents LastSeparateTime in milliseconds; a seconds-magnitude value is
# still rendered (magnitude fallback) but logged as a contract deviation.
_CONTRACT_UNIT = "ms"


def _invalid(issues: List[QueryIssue], source_id: str, path: str, message: str) -> None:
    issues.append(QueryIssue(code="SOURCE_INVALID", source_id=source_id, path=path, message=message))


def _later_time(
    raw: Any,
    current: Optional[Tuple[float, str]],
    path: str,
    source_id: str,
    issues: List[QueryIssue],
) -> Optional[Tuple[float, str]]:
    """Keep the most recent valid ``LastSeparateTime`` as (epoch seconds, UTC string)."""
    if raw is None:
        return current
    if not positive_number(raw):
        _invalid(
            issues, source_id, f"{path}.LastSeparateTime", f"last separate time {raw!r} is not a positive epoch value"
        )
        return current
    value = timestamp_value_inferred(
        raw, _CONTRACT_UNIT, "last_separate_time", issues, f"{path}.LastSeparateTime", source_id
    )
    seconds = epoch_seconds(raw, value.unit)
    if value.utc is None or seconds is None:
        return current  # unconvertible values are already reported by path
    if current is None or seconds > current[0]:
        return (seconds, value.utc)
    return current


def _device_record(
    name: str,
    reasons_raw: Any,
    path: str,
    source_id: str,
    issues: List[QueryIssue],
) -> IsolatedDeviceRecord:
    """Aggregate one roster member's Detail entries into a single public record."""
    entries: List[Any]
    if isinstance(reasons_raw, dict):
        entries = [reasons_raw]
    elif isinstance(reasons_raw, list):
        entries = reasons_raw
        if not entries:
            _invalid(issues, source_id, path, "detail is missing or empty")
    elif reasons_raw is None:
        entries = []
        _invalid(issues, source_id, path, "detail is missing or empty")
    else:
        entries = []
        _invalid(issues, source_id, path, "detail is not an object or array of objects")
    levels: List[str] = []
    codes: List[str] = []
    latest: Optional[Tuple[float, str]] = None
    for entry in entries:
        if not isinstance(entry, dict):
            _invalid(issues, source_id, path, "detail entry is not an object")
            continue
        level = entry.get("FaultLevel")
        if isinstance(level, str) and level.strip() and level.strip() not in levels:
            levels.append(level.strip())
        code = join_code(entry.get("FaultCode"))
        if code and code not in codes:
            codes.append(code)
        latest = _later_time(entry.get("LastSeparateTime"), latest, path, source_id, issues)
    return IsolatedDeviceRecord(
        device_id=name,
        fault_level=",".join(levels) if levels else "unknown",
        fault_code=",".join(codes) if codes else None,
        last_separate_time=latest[1] if latest is not None else None,
    )


def parse_manual_isolation(
    data: Dict[str, Any],
    nodes: Optional[List[str]],
    source_id: str,
    issues: List[QueryIssue],
    data_key: str = "clusterd-manual-info-cm",
) -> Tuple[Dict[str, NodeIsolation], Set[str]]:
    """Parse the CM data and return ``(per-node containers, broken node names)``.

    A node is broken when its roster could not be fully determined: the value does not
    decode, is not an object, ``Total`` is missing or not an array, or a member entry is
    invalid. Broken nodes still expose their valid members; nodes absent from the CM are
    legitimate empty rosters and appear in neither structure.
    """
    per_node: Dict[str, NodeIsolation] = {}
    broken: Set[str] = set()
    for node, raw in data.items():
        if nodes is not None and node not in nodes:
            continue
        path = f"{data_key}[{node}]"
        decoded = decode_json_string(raw, path, issues, source_id)
        if decoded is None:
            broken.add(node)
            continue
        if not isinstance(decoded, dict):
            _invalid(issues, source_id, path, "expected an object")
            broken.add(node)
            continue
        detail = decoded.get("Detail")
        if detail is None:
            detail = {}
        if not isinstance(detail, dict):
            _invalid(issues, source_id, f"{path}.Detail", "expected an object")
            detail = {}
        total = decoded.get("Total")
        if total is None:
            _invalid(issues, source_id, f"{path}.Total", "Total is missing")
            broken.add(node)
            continue
        if not isinstance(total, (list, tuple)):
            _invalid(issues, source_id, f"{path}.Total", "Total is not an array")
            broken.add(node)
            continue
        names: List[str] = []
        invalid_member = False
        for member in total:
            if isinstance(member, str) and member.strip():
                name = member.strip()
                if name not in names:
                    names.append(name)
            else:
                invalid_member = True
        if invalid_member:
            _invalid(issues, source_id, f"{path}.Total", "Total contains invalid member entries")
            broken.add(node)
        # Detail-only entries never expand the roster; they are logged and dropped.
        extra = sorted(str(device) for device in detail if str(device) not in set(names))
        if extra:
            _invalid(issues, source_id, f"{path}.Detail", "entries outside the Total roster: " + ", ".join(extra))
        devices = [
            _device_record(name, detail.get(name), f"{path}.Detail[{name}]", source_id, issues)
            for name in sorted(names)
        ]
        per_node[node] = NodeIsolation(devices=devices)
    return per_node, broken
