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

"""diagnose_cached dedup + cache + task-state tests (mocks k8s, no real cluster needed)."""

from __future__ import annotations

import json
import threading

import pytest
from kubernetes.client.exceptions import ApiException

from agent_core import k8s, tools


@pytest.fixture(autouse=True)
def _clear_recent_cache():
    tools._recent_results.clear()  # pylint: disable=protected-access
    yield


def _make_result(**kw):
    base = {
        "pods": [{"name": "p0"}],
        "collected": [{"node": "n0"}],
        "diag_input_dir": "/tmp/in",
        "diag_output_dir": "/tmp/out",
        "diag_report": {"summary": "ok"},
        "error": None,
    }
    base.update(kw)
    return base


def _noop_kube_config(monkeypatch):
    # job_exists/task_state load config via K8s.core(); mock it to avoid a real cluster
    monkeypatch.setattr(k8s.K8s, "core", classmethod(lambda cls: None))


# --------------------------------------------------------------------------- #
# dedup: a concurrent diagnosis of the same task is rejected
# --------------------------------------------------------------------------- #
def test_duplicate_diag_rejected(tmp_path, monkeypatch):
    started = threading.Event()
    release = threading.Event()

    def fake_diagnose(job, namespace):
        started.set()
        release.wait(5)
        return _make_result()

    monkeypatch.setattr(tools, "diagnose", fake_diagnose)
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "running")
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)

    first = {}
    t = threading.Thread(target=lambda: first.update(r=tools.diagnose_cached("job1", "default")))
    t.start()
    assert started.wait(2)

    dup = tools.diagnose_cached("job1", "default")
    assert dup["error"] and "正在诊断中" in dup["error"]

    release.set()
    t.join(5)
    assert first["r"]["error"] is None

    # the guard is released after completion, so diagnosis can run again
    assert tools.diagnose_cached("job1", "default")["error"] is None


# --------------------------------------------------------------------------- #
# cache read: stopped + cache exists -> return directly, no re-collect
# --------------------------------------------------------------------------- #
def test_cache_hit_when_stopped(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    tools._write_diag_cache(  # pylint: disable=protected-access
        "job1", "default", {"pods": ["cached"], "diag_report": {"s": 1}, "error": None}
    )
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "stopped")
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    r = tools.diagnose_cached("job1", "default")
    assert r["cached"] is True
    assert r["pods"] == ["cached"]
    assert not calls  # no re-collect


def test_cache_not_hit_when_running(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    tools._write_diag_cache(  # pylint: disable=protected-access
        "job1", "default", {"pods": ["cached"], "diag_report": {}, "error": None}
    )
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "running")
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: _make_result())

    r = tools.diagnose_cached("job1", "default")
    assert r.get("cached") is not True
    assert r["diag_report"] == {"summary": "ok"}  # running -> re-run, cache not hit


def test_refresh_bypasses_cache(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    tools._write_diag_cache(  # pylint: disable=protected-access
        "job1", "default", {"pods": ["cached"], "diag_report": {}, "error": None}
    )
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "stopped")
    calls = []
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    r = tools.diagnose_cached("job1", "default", refresh=True)
    assert calls == ["job1"]
    assert r["diag_report"] == {"summary": "ok"}


# --------------------------------------------------------------------------- #
# recent cache: same task within 10s returns directly, hinting --refresh
# --------------------------------------------------------------------------- #
def test_recent_cache_hit_within_ttl(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "running")
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    tools.diagnose_cached("job1", "default")
    assert calls == ["job1"]
    second = tools.diagnose_cached("job1", "default")
    assert calls == ["job1"]  # no re-collect
    assert second["cached"] is True
    assert "--refresh" in second["cached_note"]


def test_recent_cache_expired_reruns(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "running")
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "RECENT_TTL", 0.0)  # expires immediately
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    tools.diagnose_cached("job1", "default")
    second = tools.diagnose_cached("job1", "default")
    assert calls == ["job1", "job1"]
    assert second.get("cached") is not True


def test_recent_cache_bypassed_by_refresh(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "running")
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    tools.diagnose_cached("job1", "default")
    r = tools.diagnose_cached("job1", "default", refresh=True)
    assert calls == ["job1", "job1"]
    assert r.get("cached") is not True


# --------------------------------------------------------------------------- #
# fallback: when the CR status is unreadable/lagged, pod terminal phase from relcache decides cacheability
# --------------------------------------------------------------------------- #
def test_cache_written_when_pods_terminal_even_if_cr_unknown(tmp_path, monkeypatch):
    import agent_core.relcache as rc

    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "unknown")  # CR unreadable (missing RBAC etc.)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: _make_result())
    monkeypatch.setattr(rc, "lookup", lambda job, ns: [{"pod_name": "p0", "phase": "Succeeded"}])

    tools.diagnose_cached("job1", "default")
    assert (tmp_path / "default_job1.json").exists()
    r = tools.diagnose_cached("job1", "default")
    assert r["cached"] is True


def test_cache_not_written_when_pods_still_running(tmp_path, monkeypatch):
    import agent_core.relcache as rc

    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "unknown")
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: _make_result())
    monkeypatch.setattr(rc, "lookup", lambda job, ns: [{"pod_name": "p0", "phase": "Running"}])

    tools.diagnose_cached("job1", "default")
    assert not (tmp_path / "default_job1.json").exists()


# --------------------------------------------------------------------------- #
# cache write: only stopped tasks are cached, running ones are not
# --------------------------------------------------------------------------- #
def test_cache_written_only_when_stopped(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: _make_result())

    monkeypatch.setattr(tools, "task_state", lambda job, ns: "running")
    tools.diagnose_cached("job1", "default")
    assert not (tmp_path / "default_job1.json").exists()

    monkeypatch.setattr(tools, "task_state", lambda job, ns: "stopped")
    tools.diagnose_cached("job2", "default")
    cached = tmp_path / "default_job2.json"
    assert cached.exists()
    assert cached.read_text(encoding="utf-8")  # parseable non-empty json


# --------------------------------------------------------------------------- #
# task_state: determine the task CR status
# --------------------------------------------------------------------------- #
class _FakeCustom:
    """CustomObjectsApi stand-in returning a preset response."""

    def __init__(self, response, captured=None):
        self._resp = response
        self._captured = captured

    def get_namespaced_custom_object_status(self, **kw):
        if self._captured is not None:
            self._captured.update(kw)
        if isinstance(self._resp, Exception):
            raise self._resp
        return self._resp


def test_task_state_stopped_when_succeeded(monkeypatch):
    obj = {
        "status": {
            "conditions": [
                {"type": "JobRunning", "status": "True"},
                {"type": "JobSucceeded", "status": "True"},
            ]
        }
    }
    monkeypatch.setattr(tools, "_load_task_crds", lambda: [("mindxdl.gitee.com", "v1", "AscendJob")])
    monkeypatch.setattr(tools.client, "CustomObjectsApi", lambda: _FakeCustom(obj))
    _noop_kube_config(monkeypatch)
    assert tools.task_state("job1", "default") == "stopped"


def test_task_state_running(monkeypatch):
    obj = {"status": {"conditions": [{"type": "JobRunning", "status": "True"}]}}
    monkeypatch.setattr(tools, "_load_task_crds", lambda: [("mindxdl.gitee.com", "v1", "AscendJob")])
    monkeypatch.setattr(tools.client, "CustomObjectsApi", lambda: _FakeCustom(obj))
    _noop_kube_config(monkeypatch)
    assert tools.task_state("job1", "default") == "running"


def test_task_state_stopped_when_volcano_completed(monkeypatch):
    # Volcano Job terminal condition is "Completed"; it must be judged stopped to be cacheable
    obj = {"status": {"conditions": [{"type": "Running", "status": "True"}, {"type": "Completed", "status": "True"}]}}
    monkeypatch.setattr(tools, "_load_task_crds", lambda: [("batch.volcano.sh", "v1alpha1", "Job")])
    monkeypatch.setattr(tools.client, "CustomObjectsApi", lambda: _FakeCustom(obj))
    _noop_kube_config(monkeypatch)
    assert tools.task_state("job1", "default") == "stopped"


def test_task_state_terminal_aliases_and_case_insensitive(monkeypatch):
    # bare Succeeded / stopped / case-insensitive matches must all be judged stopped
    for cond_type, cond_status in (
        ("Succeeded", "True"),
        ("stopped", "True"),
        ("succeeded", "true"),
        ("FAILED", "true"),
    ):
        obj = {"status": {"conditions": [{"type": cond_type, "status": cond_status}]}}
        monkeypatch.setattr(tools, "_load_task_crds", lambda: [("mindxdl.gitee.com", "v1", "AscendJob")])
        monkeypatch.setattr(tools.client, "CustomObjectsApi", lambda obj=obj: _FakeCustom(obj))
        _noop_kube_config(monkeypatch)
        assert tools.task_state("job1", "default") == "stopped", f"{cond_type}={cond_status}"


def test_task_state_missing_cr_stopped(monkeypatch):
    def _custom_api():
        return _FakeCustom(ApiException(status=404))

    monkeypatch.setattr(tools, "_load_task_crds", lambda: [("mindxdl.gitee.com", "v1", "AscendJob")])
    monkeypatch.setattr(tools.client, "CustomObjectsApi", _custom_api)
    _noop_kube_config(monkeypatch)
    assert tools.task_state("job1", "default") == "stopped"


def test_task_state_uses_cr_params(monkeypatch):
    captured = {}

    class _Fake:
        def get_namespaced_custom_object_status(self, **kw):
            captured.update(kw)
            return {"status": {"conditions": [{"type": "JobRunning", "status": "True"}]}}

    monkeypatch.setattr(tools, "_load_task_crds", lambda: [("mindxdl.gitee.com", "v1", "AscendJob")])

    def _custom_api():
        return _Fake()

    monkeypatch.setattr(tools.client, "CustomObjectsApi", _custom_api)
    _noop_kube_config(monkeypatch)
    assert tools.task_state("job1", "myns") == "running"
    assert captured == {
        "group": "mindxdl.gitee.com",
        "version": "v1",
        "namespace": "myns",
        "plural": "ascendjobs",
        "name": "job1",
    }


# --------------------------------------------------------------------------- #
# job_exists: task CR existence check (short-circuits when the job name does not exist)
# --------------------------------------------------------------------------- #
class _FakeGet:
    """get_namespaced_custom_object stand-in: preset response/exception."""

    def __init__(self, response, captured=None):
        self._resp = response
        self._captured = captured

    def get_namespaced_custom_object(self, **kw):
        if self._captured is not None:
            self._captured.update(kw)
        if isinstance(self._resp, Exception):
            raise self._resp
        return self._resp


def test_job_exists_true(monkeypatch):
    monkeypatch.setattr(tools, "_load_task_crds", lambda: [("mindxdl.gitee.com", "v1", "AscendJob")])
    monkeypatch.setattr(tools.client, "CustomObjectsApi", lambda: _FakeGet({}))
    _noop_kube_config(monkeypatch)
    assert tools.job_exists("job1", "default") is True


def test_job_exists_false_on_404(monkeypatch):
    def _custom_api():
        return _FakeGet(ApiException(status=404))

    monkeypatch.setattr(tools, "_load_task_crds", lambda: [("mindxdl.gitee.com", "v1", "AscendJob")])
    monkeypatch.setattr(tools.client, "CustomObjectsApi", _custom_api)
    _noop_kube_config(monkeypatch)
    assert tools.job_exists("job1", "default") is False


def test_job_exists_unknown_error_treated_as_exists(monkeypatch):
    # non-404 errors (RBAC/network etc.) -> treat as existing, avoid a false "task not found"
    def _custom_api():
        return _FakeGet(ApiException(status=403))

    monkeypatch.setattr(tools, "_load_task_crds", lambda: [("mindxdl.gitee.com", "v1", "AscendJob")])
    monkeypatch.setattr(tools.client, "CustomObjectsApi", _custom_api)
    _noop_kube_config(monkeypatch)
    assert tools.job_exists("job1", "default") is True


# --------------------------------------------------------------------------- #
# diagnose_cached: job name does not exist -> short-circuit with a "task not found" message
# --------------------------------------------------------------------------- #
def test_diag_missing_job_short_circuits(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: False)
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    r = tools.diagnose_cached("nonexistent", "default")
    assert "训练/推理任务不存在" in r["error"]
    assert not calls  # no collect/diagnose steps


def test_diag_job_exists_runs_diagnose(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "running")
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: _make_result())

    r = tools.diagnose_cached("job1", "default")
    assert r["error"] is None
    assert r["diag_report"] == {"summary": "ok"}


def test_diag_cached_catches_unexpected_error(tmp_path, monkeypatch):
    # unexpected error in the flow (k8s failure / file IO) -> wrapped as a readable error, not a 500
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "task_state", lambda job, ns: "running")

    def _boom(job, ns):
        raise RuntimeError("k8s api unavailable")

    monkeypatch.setattr(tools, "diagnose", _boom)
    r = tools.diagnose_cached("job1", "default")
    assert "诊断错误" in r["error"]
    assert r["diag_report"] == {}


# --------------------------------------------------------------------------- #
# LLM summary writeback: cache_diag_final_text writes final_text into an existing cache file
# --------------------------------------------------------------------------- #
def test_cache_diag_final_text_writeback(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    tools._write_diag_cache(  # pylint: disable=protected-access
        "job1", "default", {"pods": ["p"], "diag_report": {"s": 1}, "error": None}
    )
    tools.cache_diag_final_text("job1", "default", "LLM 总结")
    data = json.loads((tmp_path / "default_job1.json").read_text(encoding="utf-8"))
    assert data["final_text"] == "LLM 总结"


def test_cache_diag_final_text_no_cache_file_is_noop(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    tools.cache_diag_final_text("job1", "default", "LLM 总结")  # no cache file -> no error
    assert not (tmp_path / "default_job1.json").exists()
