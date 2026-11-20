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

"""Public fault parser for statistic-fault-info (cluster-system)."""

from __future__ import annotations

from typing import Any, Dict, List, Optional, Set

from agent_core.datasources.utils import dedupe_records, decode_json_string, join_code, timestamp_value_inferred
from agent_core.query_schemas import PublicFaultRecord, QueryIssue

_PUBLIC_TYPES = {"public", "publicfault", "public_fault"}


def _text(value: Any) -> Optional[str]:
    return value.strip() if isinstance(value, str) and value.strip() else None


def _fault_type(value: Any) -> str:
    text = _text(value)
    if text is None or text.lower() in _PUBLIC_TYPES:
        return "PublicFault"
    return text


def _public_key(record: PublicFaultRecord):
    """The five public fields must all match for two records to be the same fault."""
    return (record.device_id, record.fault_type, record.fault_handling, record.fault_code, record.last_time)


def parse_public_faults(
    data: Dict[str, Any],
    nodes: Optional[Set[str]],
    source_id: str,
    issues: List[QueryIssue],
    data_key: str = "statistic-fault-info",
) -> Dict[str, List[PublicFaultRecord]]:
    """Return node -> public-fault records, one record per ``devIds`` element.

    ``PublicFaults`` is decoded strictly (an empty string means truncated data, not an
    empty set). ``FaultNum`` only explains producer-side truncation, so it is not parsed;
    a non-empty ``Description`` is logged but never changes the tool status.
    """
    public_faults = decode_json_string(data.get("PublicFaults"), "PublicFaults", issues, source_id)
    if public_faults is not None and not isinstance(public_faults, dict):
        issues.append(
            QueryIssue(code="SOURCE_INVALID", source_id=source_id, path="PublicFaults", message="expected an object")
        )
        public_faults = {}

    out: Dict[str, List[PublicFaultRecord]] = {}
    for node, raw_items in (public_faults or {}).items():
        name = str(node)
        if nodes is not None and name not in nodes:
            continue
        if not isinstance(raw_items, list):
            issues.append(
                QueryIssue(
                    code="SOURCE_INVALID",
                    source_id=source_id,
                    path=f"PublicFaults[{node}]",
                    message="expected an array",
                )
            )
            continue
        records: List[PublicFaultRecord] = []
        for index, raw in enumerate(raw_items):
            if not isinstance(raw, dict):
                issues.append(
                    QueryIssue(
                        code="SOURCE_INVALID",
                        source_id=source_id,
                        path=f"PublicFaults[{node}][{index}]",
                        message="fault entry is not an object",
                    )
                )
                continue
            dev_ids = raw.get("devIds")
            if not isinstance(dev_ids, list):
                issues.append(
                    QueryIssue(
                        code="SOURCE_INVALID",
                        source_id=source_id,
                        path=f"PublicFaults[{node}][{index}].devIds",
                        message="expected an array",
                    )
                )
                continue
            last_time: Optional[str] = None
            if raw.get("faultTime") is not None:
                when = timestamp_value_inferred(
                    raw.get("faultTime"),
                    "s",
                    "public_fault_time",
                    issues,
                    f"PublicFaults[{node}][{index}].faultTime",
                    source_id,
                )
                last_time = when.utc
            for dev_id in dev_ids or [None]:
                records.append(
                    PublicFaultRecord(
                        device_id="" if dev_id in (None, "") else str(dev_id),
                        fault_type=_fault_type(raw.get("type")),
                        fault_handling=_text(raw.get("level")),
                        fault_code=join_code(raw.get("faultCode")),
                        last_time=last_time,
                        resource=_text(raw.get("resource")),
                    )
                )
        if records:
            out[name] = dedupe_records(records, _public_key)

    description = data.get("Description")
    if isinstance(description, str) and description.strip():
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                source_id=source_id,
                path=f"{data_key}.Description",
                message="public fault details are truncated by the producer",
            )
        )
    return out
