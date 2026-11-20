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

"""Tool: task_status.

The job's own object is the only status source: ``AcJob.status.conditions`` for
AscendJob and ``Job.status.state.phase`` for Volcano Job. Live Pods are read only to
publish ``pods[]``; they never feed the job status. There is no fallback source: no
job-summary, hccl.json or relcache read.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Tuple

from agent_core.datasources.query_context import QueryContext, get_gateway
from agent_core.query_schemas import JobData, JobPod, JobQuery, JobQueryResult
from agent_core.tools.envelope import build_result, log_source_issue

TOOL = "task_status"

# Container waiting reasons are k8s constants: ``CrashLoopBackOff`` is published verbatim
# (no case/underscore/hyphen conversion) as a Pod-only status value.
CRASH_LOOP_BACKOFF = "CrashLoopBackOff"


@dataclass(frozen=True)
class _TaskObject:
    """One task CR type read by exact namespace/name get."""

    job_type: str
    group: str
    version: str
    plural: str


# Read order is fixed; a name hitting both types is a conflict, never a guess.
_TASK_OBJECTS = (
    _TaskObject("acjob", "mindxdl.gitee.com", "v1", "ascendjobs"),
    _TaskObject("vcjob", "batch.volcano.sh", "v1alpha1", "jobs"),
)

# AscendJob conditions: only status=True is meaningful; terminal wins over active.
_ACJOB_TERMINAL = (("Failed", "failed"), ("Succeeded", "completed"))
_ACJOB_ACTIVE = (("Running", "running"), ("Restarting", "running"), ("Created", "pending"))
# Volcano Job state.phase.
_VCJOB_PHASES = {
    "Completed": "completed",
    "Failed": "failed",
    "Aborted": "failed",
    "Running": "running",
    "Pending": "pending",
    "Enqueued": "pending",
}
_POD_PHASES = {
    "Running": "running",
    "Pending": "pending",
    "Succeeded": "completed",
    "Failed": "failed",
    "Unknown": "unknown",
}


def _controller_uid(pod: Dict[str, Any]) -> Optional[str]:
    for ref in (pod.get("metadata") or {}).get("ownerReferences") or []:
        if (ref or {}).get("controller") and (ref or {}).get("uid"):
            return ref["uid"]
    return None


def _has_crash_loop(status: Dict[str, Any]) -> bool:
    """``CrashLoopBackOff`` is a container waiting reason, never a Pod phase."""
    for key in ("containerStatuses", "initContainerStatuses"):
        for container in status.get(key) or []:
            waiting = ((container or {}).get("state") or {}).get("waiting") or {}
            if waiting.get("reason") == CRASH_LOOP_BACKOFF:
                return True
    return False


def _pod_status(pod: Dict[str, Any]) -> str:
    status = pod.get("status") or {}
    phase = status.get("phase")
    if phase == "Running" and _has_crash_loop(status):
        return CRASH_LOOP_BACKOFF
    return _POD_PHASES.get(phase, "unknown")


def _job_pods(pods: List[Dict[str, Any]], job_uid: Optional[str]) -> List[JobPod]:
    """Return the live Pods owned by the current job instance, sorted by pod_name."""
    if not job_uid:
        return []  # without the job UID, ownership cannot be verified
    by_uid: Dict[str, JobPod] = {}
    for pod in pods or []:
        if _controller_uid(pod) != job_uid:
            continue
        metadata = pod.get("metadata") or {}
        key = metadata.get("uid") or metadata.get("name") or ""
        by_uid[key] = JobPod(pod_name=metadata.get("name") or "", status=_pod_status(pod))
    return sorted(by_uid.values(), key=lambda record: record.pod_name)


def _acjob_status(obj: Dict[str, Any]) -> str:
    status = obj.get("status") or {}
    true_conditions = {
        condition.get("type")
        for condition in status.get("conditions") or []
        if str((condition or {}).get("status")) == "True"
    }
    for condition_type, value in _ACJOB_TERMINAL + _ACJOB_ACTIVE:
        if condition_type in true_conditions:
            return value
    return "unknown"


def _vcjob_status(obj: Dict[str, Any]) -> str:
    phase = ((obj.get("status") or {}).get("state") or {}).get("phase")
    return _VCJOB_PHASES.get(phase, "unknown")


def _read_task_object(
    request: JobQuery, ctx: QueryContext
) -> Tuple[Optional[Tuple[_TaskObject, Dict[str, Any]]], bool]:
    """Read both task types; return the unique hit and whether a type read gap exists."""
    gateway = get_gateway(ctx)
    found: List[Tuple[_TaskObject, Dict[str, Any]]] = []
    gap = False
    for spec in _TASK_OBJECTS:
        read = gateway.get_custom_object(spec.group, spec.version, spec.plural, request.namespace, request.job)
        log_source_issue(TOOL, spec.job_type, read.status, read.error_code, read.message)
        if read.ok and read.items:
            found.append((spec, read.items[0]))
        elif read.status != "missing":
            gap = True  # forbidden/timeout: cannot rule out this type, but not "not found"
    if len(found) != 1:
        log_source_issue(
            TOOL,
            "target",
            "missing",
            "SOURCE_NOT_FOUND",
            f"{request.namespace}/{request.job!r} matches {len(found)} task objects; job_type stays null",
        )
        return None, gap
    return found[0], gap


def task_status(request: JobQuery, ctx: QueryContext) -> JobQueryResult:
    unique, read_gap = _read_task_object(request, ctx)
    if unique is None:
        # No type can be confirmed: the Pods cannot be attributed, so they are not read.
        return build_result(
            JobQueryResult,
            TOOL,
            ctx,
            JobData(namespace=request.namespace, job_name=request.job, status="unknown"),
            complete=False,
        )

    spec, obj = unique
    job_uid = (obj.get("metadata") or {}).get("uid")
    job_status = _acjob_status(obj) if spec.job_type == "acjob" else _vcjob_status(obj)

    pods = get_gateway(ctx).list_pods(request.namespace)
    log_source_issue(TOOL, "pods", pods.status, pods.error_code, pods.message)
    pods_complete = pods.ok and pods.listing_complete

    data = JobData(
        namespace=request.namespace,
        job_name=request.job,
        job_type=spec.job_type,
        status=job_status,
        pods=_job_pods(pods.items, job_uid),
    )
    complete = not read_gap and bool(job_uid) and job_status != "unknown" and pods_complete
    return build_result(JobQueryResult, TOOL, ctx, data, complete=complete)
