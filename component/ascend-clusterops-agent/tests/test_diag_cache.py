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

"""diagnose_cached dedup + cache tests (mocks k8s, no real cluster needed)."""

from __future__ import annotations

import json
import threading

import pytest

from agent_core import tools


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

    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "diagnose", fake_diagnose)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)

    first = {}
    t = threading.Thread(target=lambda: first.update(r=tools.diagnose_cached("job1", "default")))
    t.start()
    assert started.wait(2)

    dup = tools.diagnose_cached("job1", "default")
    assert dup["error"] and "is being diagnosed" in dup["error"]

    release.set()
    t.join(5)
    assert first["r"]["error"] is None

    # the guard is released after completion, so diagnosis can run again
    assert tools.diagnose_cached("job1", "default")["error"] is None


# --------------------------------------------------------------------------- #
# cache read: any cached result (task stopped or running) is returned directly, no re-collect
# --------------------------------------------------------------------------- #
def test_cache_hit_when_stopped(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    tools._write_diag_cache(  # pylint: disable=protected-access
        "job1", "default", {"pods": ["cached"], "diag_report": {"s": 1}, "error": None}
    )
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    r = tools.diagnose_cached("job1", "default")
    assert r["cached"] is True
    assert "--refresh" in r["cached_note"]
    assert r["pods"] == ["cached"]
    assert not calls  # no re-collect


def test_cache_hit_when_running(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    tools._write_diag_cache(  # pylint: disable=protected-access
        "job1", "default", {"pods": ["cached"], "diag_report": {}, "error": None}
    )
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    r = tools.diagnose_cached("job1", "default")
    assert r["cached"] is True  # running tasks are served from the cache too
    assert "--refresh" in r["cached_note"]
    assert r["pods"] == ["cached"]
    assert not calls  # no re-collect


def test_stale_cache_not_served_when_task_missing(tmp_path, monkeypatch):
    # a cached result exists on disk, but relcache has no record of the task -> the stale
    # cache must NOT be served; return "task not found" directly.
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    tools._write_diag_cache(  # pylint: disable=protected-access
        "job1", "default", {"pods": ["cached"], "diag_report": {"s": 1}, "error": None}
    )
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: False)
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    r = tools.diagnose_cached("job1", "default")
    assert "Training/inference task not found" in r["error"]
    assert r.get("cached") is not True
    assert not calls  # no collect/diagnose steps


def test_refresh_bypasses_cache(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    tools._write_diag_cache(  # pylint: disable=protected-access
        "job1", "default", {"pods": ["cached"], "diag_report": {}, "error": None}
    )
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
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    tools.diagnose_cached("job1", "default")
    assert calls == ["job1"]
    second = tools.diagnose_cached("job1", "default")
    assert calls == ["job1"]  # no re-collect
    assert second["cached"] is True
    assert "--refresh" in second["cached_note"]


def test_recent_cache_expired_persistent_cache_hit(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "RECENT_TTL", 0.0)  # recent cache expires immediately
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    tools.diagnose_cached("job1", "default")
    second = tools.diagnose_cached("job1", "default")
    assert calls == ["job1"]  # second call is served by the persistent cache, no re-collect
    assert second["cached"] is True
    assert "--refresh" in second["cached_note"]


def test_recent_cache_bypassed_by_refresh(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    tools.diagnose_cached("job1", "default")
    r = tools.diagnose_cached("job1", "default", refresh=True)
    assert calls == ["job1", "job1"]
    assert r.get("cached") is not True


# --------------------------------------------------------------------------- #
# cache write: every successful diagnosis is cached, failed ones are not
# --------------------------------------------------------------------------- #
def test_cache_written_after_successful_diagnosis(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: _make_result())

    tools.diagnose_cached("job1", "default")
    cached = tmp_path / "default_job1.json"
    assert cached.exists()
    assert cached.read_text(encoding="utf-8")  # parseable non-empty json
    r = tools.diagnose_cached("job1", "default")
    assert r["cached"] is True  # the persistent cache serves the repeat call


def test_failed_diagnosis_not_cached(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: _make_result(error="collect failed"))

    tools.diagnose_cached("job1", "default")
    assert not (tmp_path / "default_job1.json").exists()


# --------------------------------------------------------------------------- #
# job_exists: the task is diagnosable iff relcache still holds any of its pods
# --------------------------------------------------------------------------- #
def test_job_exists_true_when_relcache_has_pods(monkeypatch):
    # live pods, or deleted-within-TTL pods whose logs remain on nodes / shared storage -> diagnosable
    from agent_core import relcache

    monkeypatch.setattr(relcache, "lookup", lambda job, ns: [{"pod_name": "p0"}])
    assert tools.job_exists("job1", "default") is True


def test_job_exists_false_when_relcache_empty(monkeypatch):
    # no pods -> wrong job name or the task's pods are all gone -> not found
    from agent_core import relcache

    monkeypatch.setattr(relcache, "lookup", lambda job, ns: [])
    assert tools.job_exists("job1", "default") is False


# --------------------------------------------------------------------------- #
# diagnose_cached: job name does not exist -> short-circuit with a "task not found" message
# --------------------------------------------------------------------------- #
def test_diag_missing_job_short_circuits(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: False)
    calls = []
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: calls.append(job) or _make_result())

    r = tools.diagnose_cached("nonexistent", "default")
    assert "Training/inference task not found" in r["error"]
    assert not calls  # no collect/diagnose steps


def test_diag_job_exists_runs_diagnose(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)
    monkeypatch.setattr(tools, "diagnose", lambda job, ns: _make_result())

    r = tools.diagnose_cached("job1", "default")
    assert r["error"] is None
    assert r["diag_report"] == {"summary": "ok"}


def test_diag_cached_catches_unexpected_error(tmp_path, monkeypatch):
    # unexpected error in the flow (k8s failure / file IO) -> wrapped as a readable error, not a 500
    monkeypatch.setattr(tools, "_cache_root", lambda: tmp_path)
    monkeypatch.setattr(tools, "job_exists", lambda job, ns: True)

    def _boom(job, ns):
        raise RuntimeError("k8s api unavailable")

    monkeypatch.setattr(tools, "diagnose", _boom)
    r = tools.diagnose_cached("job1", "default")
    assert "Diagnosis error" in r["error"]
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
