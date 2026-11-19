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

"""Request and result models for the four deterministic read-only query tools.

The public result contract is ``QueryResult`` (one envelope for every tool); each tool uses a
concrete ``data`` model so the HTTP boundary never validates arbitrary dicts.
"""

from __future__ import annotations

from typing import Annotated, Any, Dict, List, Literal, Optional

from pydantic import BaseModel, ConfigDict, Field, field_validator, model_serializer

# --------------------------------------------------------------------------- #
# Requests
# --------------------------------------------------------------------------- #
# Kubernetes DNS-1123 subdomains (Node/task names) and labels (namespaces).
# Keep constraints in the schema so discovery and Agent inputs describe the same rules.
_DNS_LABEL = r"[a-z0-9](?:[-a-z0-9]*[a-z0-9])?"
KubernetesName = Annotated[
    str, Field(strict=True, min_length=1, max_length=253, pattern=rf"^{_DNS_LABEL}(?:\.{_DNS_LABEL})*$")
]
NamespaceName = Annotated[str, Field(strict=True, min_length=1, max_length=63, pattern=rf"^{_DNS_LABEL}$")]


class QueryRequest(BaseModel):
    """Common validation policy for query inputs.

    Query tools are read-only snapshots of the current sources and expose no
    cache-bypass switch, so this base declares no shared business parameter.
    Each concrete model declares only the parameters its handler consumes.
    """

    model_config = ConfigDict(extra="forbid", strict=True)


class NodeQuery(QueryRequest):
    # None -> all nodes allowed by deployment; [] -> empty scope (no full-cluster read).
    # A single node name is accepted for compatibility with the public CLI form.
    node: Optional[List[KubernetesName]] = None

    @field_validator("node", mode="before")
    @classmethod
    def _normalize_node(cls, value: Any) -> Any:
        if isinstance(value, str):
            return [value]
        return value


class JobQuery(QueryRequest):
    job: KubernetesName
    namespace: NamespaceName = "default"


# The three node-scoped tools take exactly the same input: the RFC lists only `node`,
# so the former per-tool filters (`type`, `resource_name`) are gone and the names stay
# as aliases for the handler signatures and the registered input models.
FaultQuery = NodeQuery
IsolationQuery = NodeQuery
ResourceQuery = NodeQuery


# --------------------------------------------------------------------------- #
# Common result fragments
# --------------------------------------------------------------------------- #
class TimestampValue(BaseModel):
    raw: Optional[Any] = None
    unit: Optional[str] = None  # "s" | "ms" | None (unknown)
    utc: Optional[str] = None
    semantics: Optional[str] = None


class SourcePointer(BaseModel):
    source_id: str
    data_key: Optional[str] = None
    path: Optional[str] = None


class QueryIssue(BaseModel):
    code: str
    source_id: Optional[str] = None
    path: Optional[str] = None
    message: str
    retryable: bool = False


# --------------------------------------------------------------------------- #
# Tool 1: faulty devices
# --------------------------------------------------------------------------- #
class SuperPod(BaseModel):
    """The three producer fields that make up public `superpod` data."""

    super_pod_id: Optional[Any] = None
    rack_id: Optional[Any] = None
    server_index: Optional[Any] = None


class FaultDetail(BaseModel):
    """One fault code of a device record."""

    fault_code: Optional[str] = None
    fault_handling: Optional[str] = None
    last_time: Optional[str] = None


class DeviceFaultRecord(BaseModel):
    """One source fault entry; several codes stay in one record as several details."""

    device_id: str = ""
    device_type: Optional[str] = None
    chip_type: Optional[str] = None
    fault_type: Optional[str] = None
    faults: List[FaultDetail] = Field(default_factory=list)
    superpod: Optional[SuperPod] = None


class SwitchFaultRecord(BaseModel):
    """One active switch alarm.

    The ClusterD switch summary has no fault-type dimension, so this record carries no
    ``fault_type`` key: a field whose value never depends on the source is not evidence.
    """

    device_id: str = ""
    fault_handling: Optional[str] = None
    fault_code: Optional[str] = None
    last_time: Optional[str] = None


class DpuFaultRecord(BaseModel):
    """One DPU device fault or DPU node event; neither source carries a fault type."""

    device_id: str = ""
    fault_handling: Optional[str] = None
    fault_code: Optional[str] = None
    last_time: Optional[str] = None
    # Only DPU device entries carry this relation; node events leave it unset.
    affected_npu: Optional[List[str]] = None

    @model_serializer(mode="wrap")
    def _omit_unset_relation(self, handler):
        return {key: value for key, value in handler(self).items() if key != "affected_npu" or value is not None}


class PublicFaultRecord(BaseModel):
    device_id: str = ""
    fault_type: Optional[str] = None
    fault_handling: Optional[str] = None
    fault_code: Optional[str] = None
    last_time: Optional[str] = None
    # Internal only: cluster_resource classifies identities until the resource batch replaces it.
    resource: Optional[str] = Field(default=None, exclude=True)


class NodeHealth(BaseModel):
    """Raw NodeStatus from cluster-info-node-cm."""

    node_status: Optional[str] = None


class NodeFaults(BaseModel):
    """One node container; supplement-only nodes omit ``devices`` and ``node_health``."""

    devices: Optional[List[DeviceFaultRecord]] = None
    switch: List[SwitchFaultRecord] = Field(default_factory=list)
    dpu: List[DpuFaultRecord] = Field(default_factory=list)
    public_faults: List[PublicFaultRecord] = Field(default_factory=list)
    node_health: Optional[NodeHealth] = None

    @model_serializer(mode="wrap")
    def _omit_absent(self, handler):
        return {key: value for key, value in handler(self).items() if value is not None}


# --------------------------------------------------------------------------- #
# Tool 2: isolated devices
# --------------------------------------------------------------------------- #
class IsolatedDeviceRecord(BaseModel):
    """One ClusterD-confirmed isolated device; its Detail facts are aggregated per device."""

    device_id: str
    fault_level: str = "unknown"
    fault_code: Optional[str] = None
    last_separate_time: Optional[str] = None


class NodeIsolation(BaseModel):
    """One node container; only ClusterD ``Total`` members populate ``devices``."""

    devices: List[IsolatedDeviceRecord] = Field(default_factory=list)


# --------------------------------------------------------------------------- #
# Tool 3: cluster resource stats
# --------------------------------------------------------------------------- #
class CpuResource(BaseModel):
    """CPU block: millicores are accumulated internally, cores are published."""

    capacity_cores: Optional[float] = None
    allocatable_cores: Optional[float] = None
    used_cores: float = 0.0
    idle_cores: Optional[float] = None


class ResourceBucket(BaseModel):
    """One whole-card bucket (``Ascend910``/``npu``) with independent fault/isolated counts."""

    capacity: Optional[int] = None
    allocatable: Optional[int] = None
    used: int = 0
    idle: Optional[int] = None
    fault: Optional[int] = None
    isolated: Optional[int] = None


class NodeResourceData(BaseModel):
    """One node container: the CPU block plus one bucket per chip type.

    ``cpu`` and every whole-card bucket (``Ascend910``/``npu``) are sibling keys, not
    nested under a ``resources`` wrapper: CPU is itself a resource, only with a different
    field name and unit. Buckets are carried as extra fields so a node can expose any
    chip type without a fixed schema.
    """

    model_config = ConfigDict(extra="allow")

    cpu: CpuResource = Field(default_factory=CpuResource)

    def buckets(self) -> Dict[str, ResourceBucket]:
        """Return the whole-card buckets; ``cpu`` is a sibling key, never a bucket."""
        return {key: ResourceBucket.model_validate(value) for key, value in (self.model_extra or {}).items()}


# --------------------------------------------------------------------------- #
# Tool 4: job status
# --------------------------------------------------------------------------- #
class JobPod(BaseModel):
    """One live Pod of the job: only the name and the normalized Pod status."""

    pod_name: str
    status: str


# --------------------------------------------------------------------------- #
# Public query data contract (only externally exposed business fields)
# --------------------------------------------------------------------------- #
class FaultData(BaseModel):
    node: Dict[str, NodeFaults] = Field(default_factory=dict)


class IsolationData(BaseModel):
    node: Dict[str, NodeIsolation] = Field(default_factory=dict)


class ResourceData(BaseModel):
    node: Dict[str, NodeResourceData] = Field(default_factory=dict)


class JobData(BaseModel):
    """The job object's own status plus its live Pods; ``pods`` is always an array."""

    namespace: str
    job_name: str
    job_type: Optional[str] = None
    status: str
    pods: List[JobPod] = Field(default_factory=list)


# --------------------------------------------------------------------------- #
# Result envelope
# --------------------------------------------------------------------------- #
class ToolEnvelope(BaseModel):
    """Fields shared by every tool response, query and diagnosis alike.

    Kept in this module (rather than ``tools.schemas``) because it depends on
    ``QueryIssue``; ``tools.schemas`` builds the diagnosis envelope on top of it.
    The envelope carries execution state only: format/version metadata was removed
    to keep the response free of fields the caller cannot act on.
    """

    request_id: str
    tool: str
    status: Literal["ok", "partial", "error"] = "ok"
    elapsed_ms: int = 0
    data: Optional[Any] = None
    errors: List[QueryIssue] = Field(default_factory=list)


class QueryResult(ToolEnvelope):
    pass


class FaultQueryResult(QueryResult):
    data: Optional[FaultData] = None


class IsolationQueryResult(QueryResult):
    data: Optional[IsolationData] = None


class ResourceQueryResult(QueryResult):
    data: Optional[ResourceData] = None


class JobQueryResult(QueryResult):
    data: Optional[JobData] = None


class QueryFailure(Exception):
    """Machine-readable query failure mapped to a uniform HTTP/CLI result."""

    def __init__(self, code: str, message: str, http_status: int = 500, path: Optional[str] = None):
        super().__init__(message)
        self.code = code
        self.message = message
        self.http_status = http_status
        self.path = path
