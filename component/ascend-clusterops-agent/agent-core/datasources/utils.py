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

"""Small pure helpers shared by the query datasources (no network, no model, no cache writes)."""

# pylint: disable=duplicate-code  # small defensive parsing blocks intentionally mirror sibling datasources

from __future__ import annotations

import json
from datetime import datetime, timezone
from typing import Any, Callable, Iterable, List, Optional, Tuple

from agent_core.query_schemas import QueryIssue, TimestampValue

# Whole-card NPU resource keys the accounting understands; a cluster may contain several
# at once. Kept as an explicit constant so the accepted keys no longer double as a request
# parameter (the RFC only defines `node`).
RESOURCE_PREFIXES = ("huawei.com/Ascend910", "huawei.com/npu")
_DEVICE_LIST_SUFFIXES = ("-Fault", "-NetworkUnhealthy", "-Unhealthy", "-Recovering")

# Producer epoch values are either seconds or milliseconds. Epoch seconds stay below 1e11
# until year 5138; epoch milliseconds exceed it after 1973-03. Every plausible fault/report
# timestamp therefore falls cleanly on one side, without trusting a per-field assumption.
_EPOCH_MS_THRESHOLD = 10**11

# Topology sentinels published by the producers instead of an absent value.
_INVALID_SERVER_INDEX = 1023
_INVALID_RACK_IDS = (-1, -2)


def infer_time_unit(raw: Any) -> Optional[str]:
    """Classify an untyped producer epoch timestamp as ``s`` or ``ms`` by magnitude.

    ``fault_time`` is genuinely mixed: hang-detection faults are written in seconds while
    DCMI hardware events are written in milliseconds. Deciding per value keeps both
    readable instead of forcing one unit on the whole field.
    """
    if raw is None or isinstance(raw, bool):
        return None
    try:
        number = float(raw)
    except (TypeError, ValueError):
        return None
    if number <= 0:
        return None
    return "ms" if number >= _EPOCH_MS_THRESHOLD else "s"


def positive_number(raw: Any) -> bool:
    """True when ``raw`` parses to a strictly positive number (bool is not a timestamp)."""
    if raw is None or isinstance(raw, bool):
        return False
    try:
        return float(raw) > 0
    except (TypeError, ValueError):
        return False


def epoch_seconds(raw: Any, unit: Optional[str]) -> Optional[float]:
    """Normalize a timestamp with a *known* unit to epoch seconds; None otherwise."""
    if raw is None or unit not in ("s", "ms") or isinstance(raw, bool):
        return None
    try:
        number = float(raw)
    except (TypeError, ValueError):
        return None
    if number <= 0:
        return None
    return number if unit == "s" else number / 1000.0


def timestamp_value(raw: Any, unit: Optional[str] = None, semantics: Optional[str] = None) -> TimestampValue:
    """Build a TimestampValue; the UTC string is only produced for explicitly known units."""
    value = TimestampValue(raw=raw, unit=unit, semantics=semantics)
    seconds = epoch_seconds(raw, unit)
    if seconds is None:
        return value
    try:
        moment = datetime.fromtimestamp(seconds, tz=timezone.utc)
    except (OverflowError, OSError, ValueError):
        return value
    # Whole seconds stay `...:SSZ`; a fractional source is rendered with exactly 3 digits
    # so the public string never carries microseconds.
    timespec = "milliseconds" if moment.microsecond else "seconds"
    value.utc = moment.isoformat(timespec=timespec).replace("+00:00", "Z")
    return value


def timestamp_value_inferred(
    raw: Any,
    expected_unit: Optional[str],
    semantics: Optional[str],
    issues: List[QueryIssue],
    path: str,
    source_id: Optional[str] = None,
) -> TimestampValue:
    """Infer the unit, render the UTC value, and record any contract deviation.

    ``expected_unit`` is the producer's documented unit. A mismatch is kept visible as a
    warning but never blocks the value: the magnitude is the only thing we can verify.
    """
    unit = infer_time_unit(raw)
    value = timestamp_value(raw, unit, semantics=semantics)
    if unit is None:
        if positive_number(raw):
            issues.append(
                QueryIssue(
                    code="SOURCE_INVALID",
                    source_id=source_id,
                    path=path,
                    message=f"timestamp {raw!r} is not a recognizable epoch second/millisecond value",
                )
            )
        return value
    if value.utc is None:
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                source_id=source_id,
                path=path,
                message=f"timestamp {raw!r} (inferred {unit}) cannot be converted to a UTC datetime",
            )
        )
    elif expected_unit and unit != expected_unit:
        issues.append(
            QueryIssue(
                code="SOURCE_INVALID",
                source_id=source_id,
                path=path,
                message=f"timestamp unit inferred as {unit}, documented contract is {expected_unit}",
            )
        )
    return value


def normalize_topology(raw: Optional[dict]) -> dict:
    """Normalize a node topology block and drop the producer sentinel values.

    Input uses the normalized snake_case keys (``super_pod_id`` / ``server_index`` /
    ``rack_id``), not the producer keys: callers map ``SuperPodID`` / ``ServerIndex`` /
    ``RackID`` first, and a producer-shaped block silently normalizes to ``{}``. Values may
    be ints or digit strings; anything else counts as absent.

    ``super_pod_id`` is the gate: a missing or negative id means "no usable topology"
    (``-1`` is the documented "not in a superpod", ``-2`` is "unsupported or driver read
    failure") and the whole block is omitted. Inside a real superpod, ``server_index`` /
    ``rack_id`` are 0-based, so only the sentinels are removed per sub-key and ``0`` must be
    kept. ``1023`` for ``server_index`` is an observed driver value when the slot is
    unknown, not a documented MindCluster API sentinel.
    """
    if not isinstance(raw, dict):
        return {}

    def _as_int(value: Any) -> Optional[int]:
        if isinstance(value, bool) or value is None:
            return None
        if isinstance(value, int):
            return value
        if isinstance(value, str) and value.strip().lstrip("-").isdigit():
            return int(value)
        return None

    super_pod = _as_int(raw.get("super_pod_id"))
    if super_pod is None or super_pod < 0:
        return {}
    topology = {"super_pod_id": super_pod}
    server_index = _as_int(raw.get("server_index"))
    if server_index is not None and server_index >= 0 and server_index != _INVALID_SERVER_INDEX:
        topology["server_index"] = server_index
    rack_id = _as_int(raw.get("rack_id"))
    if rack_id is not None and rack_id >= 0 and rack_id not in _INVALID_RACK_IDS:
        topology["rack_id"] = rack_id
    return topology


def decode_json_string(
    value: Any,
    path: str,
    issues: List[QueryIssue],
    source_id: Optional[str] = None,
    allow_empty: bool = False,
) -> Any:
    """Decode a CM data value that should be JSON (tolerates an already-decoded object).

    ``allow_empty`` is only for sources whose producer explicitly writes an empty string to
    mean "no data" (ClusterD clears unused/empty shards). Other callers must keep the strict
    behavior so a truncated job summary or public-fault document still reports SOURCE_INVALID.
    """
    if value is None:
        return None
    if isinstance(value, (dict, list, int, float, bool)):
        return value
    if isinstance(value, str) and allow_empty and not value.strip():
        return None
    if not isinstance(value, str):
        issues.append(
            QueryIssue(code="SOURCE_INVALID", source_id=source_id, path=path, message="value is not a string")
        )
        return None
    try:
        return json.loads(value)
    except (json.JSONDecodeError, ValueError):
        issues.append(
            QueryIssue(code="SOURCE_INVALID", source_id=source_id, path=path, message="value is not valid JSON")
        )
        return None


def join_code(value: Any) -> Optional[str]:
    """Normalize a producer code that may be a string, an array or a scalar."""
    if value is None:
        return None
    if isinstance(value, (list, tuple, set)):
        parts = [str(item).strip() for item in value if str(item).strip()]
        return ",".join(parts) if parts else None
    if isinstance(value, str):
        return value.strip() or None
    return str(value)


def dedupe_records(records: Iterable[Any], key: Callable[[Any], Any]) -> List[Any]:
    """Keep the first occurrence of each key, preserving source order."""
    unique: List[Any] = []
    seen = set()
    for record in records:
        marker = key(record)
        if marker in seen:
            continue
        seen.add(marker)
        unique.append(record)
    return unique


def split_csv(value: Any) -> List[str]:
    if value is None:
        return []
    if isinstance(value, (list, tuple)):
        return [str(v).strip() for v in value if str(v).strip()]
    if not isinstance(value, str):
        return []
    return [item.strip() for item in value.split(",") if item.strip()]


def strip_device_list_suffix(key: str) -> Tuple[str, str]:
    """Split a DP DeviceList key into (resource key, suffix) where suffix is '' for the base key."""
    for suffix in _DEVICE_LIST_SUFFIXES:
        if key.endswith(suffix):
            return key[: -len(suffix)], suffix
    return key, ""


def supported_whole_card_key(key: str) -> Optional[str]:
    return key if key in RESOURCE_PREFIXES else None


def chip_type(resource_key: str) -> str:
    """The chip/bucket name of a whole-card key (``huawei.com/npu`` -> ``npu``)."""
    return resource_key.split("/", 1)[1] if "/" in resource_key else resource_key


def first_non_empty(mapping: dict, keys: Iterable[str]) -> Tuple[Optional[str], Optional[str]]:
    """Return (value, source_field) of the first non-empty key; unknown enums are kept verbatim."""
    for key in keys:
        value = mapping.get(key)
        if isinstance(value, str) and value.strip():
            return value, key
    return None, None


def node_from_cm_name(name: str, prefixes: Iterable[str]) -> str:
    """Strip a known CM prefix to obtain the node name; never split on the last hyphen."""
    for prefix in prefixes:
        if name.startswith(prefix):
            return name[len(prefix) :]
    return name


def parse_non_negative_number(value: Any) -> Optional[float]:
    """Parse a non-negative number from an int/float/numeric string; bool/None/invalid -> None."""
    if isinstance(value, bool) or value is None:
        return None
    if isinstance(value, (int, float)):
        number = float(value)
    elif isinstance(value, str):
        text = value.strip()
        if not text:
            return None
        try:
            number = float(text)
        except ValueError:
            return None
    else:
        return None
    return number if number >= 0 else None


def to_int_quantity(value: Any) -> Optional[int]:
    """Parse a Kubernetes quantity/card count; returns None for bool/negative/fraction/invalid."""
    number = parse_non_negative_number(value)
    if number is None or not number.is_integer():
        return None
    return int(number)
