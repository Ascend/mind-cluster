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
# ============================================================================

"""task_status: the task CR is the only status source; Pods only fill ``pods[]``."""

# pylint: disable=redefined-outer-name  # pytest fixtures are injected by parameter name

import pytest
from agent_core import query_api
from agent_core.datasources.query_context import QueryContext
from agent_core.datasources.gateway import SourceReadResult
from agent_core.query_schemas import JobQuery
from agent_core.tools.task_status import task_status
from fastapi import FastAPI
from fastapi.testclient import TestClient

JOB_UID = "uid-job"


class FakeGateway:
    """Hands out one task object per CRD plural plus a single namespace-wide Pod list."""

    def __init__(self, ctx, objects=None, pods=(), pods_result=None):
        self.ctx = ctx
        self.objects = dict(objects or {})
        self.pods = list(pods)
        self.pods_result = pods_result
        self.pod_reads = 0

    def _source_id(self):
        return self.ctx.next_source_id()

    def get_custom_object(self, group, version, plural, namespace, name):
        value = self.objects.get((plural, name))
        if isinstance(value, SourceReadResult):
            return value
        if value is None:
            return SourceReadResult(status="missing", source_id=self._source_id(), error_code="SOURCE_NOT_FOUND")
        return SourceReadResult(status="ok", source_id=self._source_id(), items=[value], observed=1)

    def list_pods(self, namespace=None):
        self.pod_reads += 1
        if self.pods_result is not None:
            return self.pods_result
        return SourceReadResult(
            status="ok" if self.pods else "empty", source_id=self._source_id(), items=list(self.pods)
        )


def acjob(uid=JOB_UID, conditions=(("Running", "True"),)):
    return {
        "metadata": {"uid": uid, "name": "train-a"},
        "status": {"conditions": [{"type": kind, "status": value} for kind, value in conditions]},
    }


def vcjob(uid=JOB_UID, phase="Running"):
    return {"metadata": {"uid": uid, "name": "train-a"}, "status": {"state": {"phase": phase}}}


def pod(name, uid, owner_uid=JOB_UID, phase="Running", waiting=None, init_waiting=None):
    metadata = {"name": name, "namespace": "default", "uid": uid}
    if owner_uid:
        metadata["ownerReferences"] = [{"controller": True, "uid": owner_uid, "name": "train-a"}]
    status = {"phase": phase}
    if waiting:
        status["containerStatuses"] = [{"state": {"waiting": {"reason": waiting}}}]
    if init_waiting:
        status["initContainerStatuses"] = [{"state": {"waiting": {"reason": init_waiting}}}]
    return {"metadata": metadata, "status": status}


def run(objects=None, pods=(), pods_result=None):
    request = JobQuery(job="train-a", namespace="default")
    ctx = QueryContext(request)
    ctx.gateway = FakeGateway(ctx, objects=objects, pods=pods, pods_result=pods_result)
    return task_status(request, ctx), ctx


@pytest.mark.parametrize(
    "conditions,expected",
    [
        ((("Running", "True"),), "running"),
        ((("Restarting", "True"),), "running"),
        ((("Created", "True"),), "pending"),
        ((("Succeeded", "True"),), "completed"),
        ((("Failed", "True"),), "failed"),
        ((("Running", "True"), ("Failed", "True")), "failed"),
        ((("Running", "False"),), "unknown"),
        ((), "unknown"),
    ],
)
def test_acjob_conditions_normalize_to_the_public_status(conditions, expected):
    result, _ = run(objects={("ascendjobs", "train-a"): acjob(conditions=conditions)})

    assert result.data.job_type == "acjob"
    assert result.data.status == expected
    assert result.status == ("partial" if expected == "unknown" else "ok")


@pytest.mark.parametrize(
    "phase,expected",
    [
        ("Completed", "completed"),
        ("Failed", "failed"),
        ("Aborted", "failed"),
        ("Running", "running"),
        ("Pending", "pending"),
        ("Enqueued", "pending"),
        ("InPlaceRestarting", "unknown"),
    ],
)
def test_vcjob_phase_normalizes_to_the_public_status(phase, expected):
    result, _ = run(objects={("jobs", "train-a"): vcjob(phase=phase)})

    assert result.data.job_type == "vcjob"
    assert result.data.status == expected
    assert result.status == ("partial" if expected == "unknown" else "ok")


def test_pods_are_matched_by_controller_uid_deduplicated_and_sorted():
    pods = [
        pod("worker-1", "pod-b", phase="Succeeded"),
        pod("worker-0", "pod-a", phase="Running"),
        pod("worker-0", "pod-a", phase="Running"),  # same Pod seen twice: one record
        pod("old-worker", "pod-old", owner_uid="uid-previous", phase="Succeeded"),
        pod("unrelated", "pod-x", owner_uid=None, phase="Failed"),
    ]
    result, _ = run(objects={("ascendjobs", "train-a"): acjob()}, pods=pods)

    assert result.status == "ok"
    assert [(record.pod_name, record.status) for record in result.data.pods] == [
        ("worker-0", "running"),
        ("worker-1", "completed"),
    ]


@pytest.mark.parametrize("field", ["containerStatuses", "initContainerStatuses"])
def test_crash_loop_backoff_is_published_verbatim_for_a_running_pod(field):
    crashing = pod("worker-0", "pod-a", phase="Running")
    crashing["status"][field] = [{"state": {"waiting": {"reason": "CrashLoopBackOff"}}}]
    result, _ = run(objects={("ascendjobs", "train-a"): acjob()}, pods=[crashing])

    assert result.data.status == "running"  # the job status never carries a container reason
    assert result.data.pods[0].status == "CrashLoopBackOff"
    assert result.status == "ok"


def test_non_crash_waiting_reason_keeps_the_phase_status():
    pulling = pod("worker-0", "pod-a", phase="Pending", waiting="ImagePullBackOff")
    failed = pod("worker-1", "pod-b", phase="Failed")
    result, _ = run(objects={("ascendjobs", "train-a"): acjob()}, pods=[pulling, failed])

    assert [(record.pod_name, record.status) for record in result.data.pods] == [
        ("worker-0", "pending"),
        ("worker-1", "failed"),
    ]


def test_missing_task_object_is_unknown_and_never_reads_pods():
    result, ctx = run()

    assert result.status == "partial"
    assert result.data.model_dump() == {
        "namespace": "default",
        "job_name": "train-a",
        "job_type": None,
        "status": "unknown",
        "pods": [],
    }
    assert ctx.gateway.pod_reads == 0


def test_a_name_matching_both_task_types_is_a_conflict_not_a_guess():
    result, ctx = run(objects={("ascendjobs", "train-a"): acjob(), ("jobs", "train-a"): vcjob()})

    assert result.status == "partial"
    assert result.data.job_type is None
    assert result.data.status == "unknown"
    assert result.data.pods == []
    assert ctx.gateway.pod_reads == 0


@pytest.mark.parametrize("failure", ["forbidden", "unavailable"])
def test_unreadable_other_task_type_still_reports_the_unique_hit_as_partial(failure):
    objects = {
        ("ascendjobs", "train-a"): acjob(),
        ("jobs", "train-a"): SourceReadResult(status=failure, source_id="s1", error_code="SOURCE_" + failure.upper()),
    }
    result, _ = run(objects=objects)

    assert result.status == "partial"
    assert result.data.job_type == "acjob"
    assert result.data.status == "running"


def test_incomplete_pod_listing_keeps_observed_pods_and_is_partial():
    observed = pod("worker-0", "pod-a", phase="Running")
    pods_result = SourceReadResult(
        status="unavailable",
        source_id="s3",
        items=[observed],
        error_code="SOURCE_UNAVAILABLE",
        listing_complete=False,
        observed=1,
    )
    result, _ = run(objects={("ascendjobs", "train-a"): acjob()}, pods_result=pods_result)

    assert result.status == "partial"
    assert [(record.pod_name, record.status) for record in result.data.pods] == [("worker-0", "running")]


def test_http_exposes_only_public_job_fields(monkeypatch):
    request = JobQuery(job="train-a", namespace="default")
    ctx = QueryContext(request)
    ctx.gateway = FakeGateway(
        ctx,
        objects={("ascendjobs", "train-a"): acjob()},
        pods=[pod("worker-0", "pod-a", phase="Running")],
    )
    monkeypatch.setattr(query_api, "QueryContext", lambda *_args, **_kwargs: ctx)
    app = FastAPI()
    app.include_router(query_api.router)
    response = TestClient(app).post("/tools/task_status", json=request.model_dump())

    assert response.status_code == 200
    payload = response.json()
    assert payload["status"] == "ok"
    assert set(payload["data"]) == {"namespace", "job_name", "job_type", "status", "pods"}
    assert payload["data"]["pods"] == [{"pod_name": "worker-0", "status": "running"}]
