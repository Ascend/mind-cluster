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

"""agent.tools unit tests (mocks k8s/HTTP/ascend-fd, no real cluster needed)."""

from __future__ import annotations

from pathlib import Path
from types import SimpleNamespace

import pytest

from agent_core import k8s, tools


# --------------------------------------------------------------------------- #
# assemble_diag_input
# --------------------------------------------------------------------------- #
def test_assemble_diag_input_copies_ok_workers(tmp_path, monkeypatch):
    monkeypatch.setattr(tools, "WORK_ROOT", tmp_path)
    w1 = tmp_path / "w1"
    w1.mkdir()
    (w1 / "server-info.json").write_text("{}")
    (w1 / "ascend-rc-parser.json").write_text('{"k":1}')
    w2 = tmp_path / "w2"
    w2.mkdir()
    (w2 / "server-info.json").write_text("{}")
    collected = [
        {"ok": True, "node": "node-a", "worker_dir": str(w1)},
        {"ok": True, "node": "node-b", "worker_dir": str(w2)},
        {"ok": False, "node": "node-c", "worker_dir": None},
    ]
    diag_input = Path(tools.assemble_diag_input(collected, "job1", "default"))
    assert (diag_input / "worker0" / "server-info.json").exists()
    assert (diag_input / "worker0" / "ascend-rc-parser.json").exists()
    assert (diag_input / "worker1" / "server-info.json").exists()
    assert not (diag_input / "worker2").exists()
    assert not (diag_input / "worker-node-a").exists()  # not named by node anymore
    assert not w1.exists()  # rename moves, does not copy; the original dir is moved away


# --------------------------------------------------------------------------- #
# run_diag: monkeypatch subprocess.run (cross-platform, no real ascend-fd needed)
# --------------------------------------------------------------------------- #
def test_run_diag_reads_report(tmp_path, monkeypatch):
    import subprocess

    monkeypatch.setattr(tools, "WORK_ROOT", tmp_path)

    def fake_run(cmd, **kwargs):
        out = Path(cmd[cmd.index("-o") + 1])
        (out / "fault_diag_result").mkdir(parents=True, exist_ok=True)
        (out / "fault_diag_result" / "diag_report.json").write_text('{"root_cause":"n1","events":[]}')
        return SimpleNamespace(returncode=0, stdout="", stderr="")

    monkeypatch.setattr(subprocess, "run", fake_run)

    in_dir = tmp_path / "default_job1" / "diag-input"
    in_dir.mkdir(parents=True)
    out, report = tools.run_diag(str(in_dir))
    assert report["root_cause"] == "n1"
    assert Path(out).exists()


def test_run_diag_missing_report_returns_empty(tmp_path, monkeypatch):
    import subprocess

    monkeypatch.setattr(tools, "WORK_ROOT", tmp_path)
    monkeypatch.setattr(subprocess, "run", lambda *a, **k: SimpleNamespace(returncode=0, stdout="", stderr=""))
    in_dir = tmp_path / "default_job1" / "diag-input"
    in_dir.mkdir(parents=True)
    _out, report = tools.run_diag(str(in_dir))
    assert report == {}


def test_run_diag_timeout_raises_diagerror(tmp_path, monkeypatch):
    # ascend-fd diag timeout -> wrapped as _DiagError (with a timeout hint), not a leaked TimeoutExpired
    import subprocess

    monkeypatch.setattr(tools, "WORK_ROOT", tmp_path)

    def _timeout(*_a, **_k):
        raise subprocess.TimeoutExpired(cmd=["ascend-fd"], timeout=1800)

    monkeypatch.setattr(subprocess, "run", _timeout)
    in_dir = tmp_path / "default_job1" / "diag-input"
    in_dir.mkdir(parents=True)
    with pytest.raises(tools._DiagError) as ei:  # pylint: disable=protected-access
        tools.run_diag(str(in_dir))
    assert "诊断超时" in str(ei.value)


# --------------------------------------------------------------------------- #
# dispatch_collect: groups by node and issues TriggerCollect, body has {pod_uid,ns,name},
# no host_paths (resolved by the local pathmap)
# --------------------------------------------------------------------------- #
def test_dispatch_collect_body_no_host_paths(monkeypatch):
    captured = {}

    def fake_trigger(v1, node, job, refs):
        captured[node] = {"job": job, "refs": refs}
        return {"ok": True, "node": node, "error": None, "worker_dir": "/w", "artifacts_tar": None}

    monkeypatch.setattr(tools, "_trigger_node", fake_trigger)
    monkeypatch.setattr(k8s.K8s, "core", classmethod(lambda cls: SimpleNamespace))
    monkeypatch.setattr(tools.tracker, "wait", lambda job, timeout: {})  # do not wait for real uploads
    pods = [
        {"pod_name": "p0", "pod_uid": "u0", "node": "n0", "rank": "0", "namespace": "default"},
        {"pod_name": "p1", "pod_uid": "u1", "node": "n0", "rank": "1", "namespace": "default"},
        {"pod_name": "p2", "pod_uid": "u2", "node": "n1", "rank": "0", "namespace": "default"},
    ]
    tools.dispatch_collect("job-x", pods)
    assert set(captured) == {"n0", "n1"}
    assert all(v["job"] == "job-x" for v in captured.values())
    n0 = captured["n0"]["refs"]
    assert {p["pod_uid"] for p in n0} == {"u0", "u1"}
    assert all("host_paths" not in p for p in n0)  # host_paths are not dispatched
    assert all(set(p) == {"pod_uid", "ns", "name"} for p in n0)


# --------------------------------------------------------------------------- #
# _group_by_node: dedup rescheduled same-name pods on the same node (keep the newest UID)
# --------------------------------------------------------------------------- #
def _pod(name, uid, node, deleted_at=None):
    return {
        "pod_name": name,
        "pod_uid": uid,
        "node": node,
        "rank": "0",
        "namespace": "default",
        "deleted_at": deleted_at,
    }


def test_group_by_node_dedups_same_pod_on_same_node():
    # the same logical pod rescheduled 4 times (4 UIDs); the newest is alive -> keep only it
    pods = [
        _pod("taskmgr-npu-019-default-test-0", "u1", "n0", deleted_at=100.0),
        _pod("taskmgr-npu-019-default-test-0", "u2", "n0", deleted_at=200.0),
        _pod("taskmgr-npu-019-default-test-0", "u3", "n0", deleted_at=300.0),
        _pod("taskmgr-npu-019-default-test-0", "u4", "n0", deleted_at=None),
    ]
    by_node = tools._group_by_node(pods)  # pylint: disable=protected-access
    refs = by_node["n0"]
    assert [p["pod_uid"] for p in refs] == ["u4"]  # the newest alive instance
    assert refs[0]["name"] == "taskmgr-npu-019-default-test-0"


def test_group_by_node_all_dead_keeps_latest_deleted():
    # all deleted -> keep the one with the newest deleted_at
    pods = [
        _pod("p", "u1", "n0", deleted_at=100.0),
        _pod("p", "u2", "n0", deleted_at=300.0),
        _pod("p", "u3", "n0", deleted_at=200.0),
    ]
    by_node = tools._group_by_node(pods)  # pylint: disable=protected-access
    assert [p["pod_uid"] for p in by_node["n0"]] == ["u2"]


def test_group_by_node_keeps_same_name_on_different_nodes():
    # same-name pod rescheduled to another node -> keep one per node (historic node logs are still collectable)
    pods = [
        _pod("p", "u1", "n0", deleted_at=100.0),
        _pod("p", "u2", "n1", deleted_at=None),
    ]
    by_node = tools._group_by_node(pods)  # pylint: disable=protected-access
    assert set(by_node) == {"n0", "n1"}
    assert [p["pod_uid"] for p in by_node["n0"]] == ["u1"]
    assert [p["pod_uid"] for p in by_node["n1"]] == ["u2"]


# --------------------------------------------------------------------------- #
# diagnose: error branches + normal path (mocks relcache.lookup)
# --------------------------------------------------------------------------- #
def test_diagnose_no_pods(monkeypatch):
    import agent_core.relcache as rc

    monkeypatch.setattr(rc, "lookup", lambda job, namespace="default": [])
    r = tools.diagnose(job="x", namespace="default")
    assert "没有可诊断的 pod" in r["error"]
    assert r["pods"] == []


def test_diagnose_all_collect_fail(monkeypatch):
    import agent_core.relcache as rc

    monkeypatch.setattr(
        rc,
        "lookup",
        lambda job, namespace="default": [
            {"pod_name": "p", "pod_uid": "u", "node": "n", "rank": "", "namespace": "default"}
        ],
    )
    monkeypatch.setattr(
        tools, "dispatch_collect", lambda job, pods: [{"ok": False, "node": "n", "error": "boom", "worker_dir": None}]
    )
    r = tools.diagnose(job="x")
    assert "节点采集全部失败" in r["error"]
    assert r["pods"] and not r["collected"][0]["ok"]


def test_diagnose_ok(monkeypatch, tmp_path):
    import agent_core.relcache as rc

    pods = [{"pod_name": "p", "pod_uid": "u", "node": "n", "rank": "0", "namespace": "default"}]
    monkeypatch.setattr(rc, "lookup", lambda job, namespace="default": pods)
    worker_dir = tmp_path / "w"
    worker_dir.mkdir()
    (worker_dir / "server-info.json").write_text("{}")
    monkeypatch.setattr(
        tools,
        "dispatch_collect",
        lambda job, pods: [
            {"ok": True, "node": "n", "error": None, "worker_dir": str(worker_dir), "artifacts_tar": None}
        ],
    )
    monkeypatch.setattr(tools, "WORK_ROOT", tmp_path)
    monkeypatch.setattr(tools, "run_diag", lambda d: (str(tmp_path / "out"), {"root_cause": "n"}))
    r = tools.diagnose(job="x")
    assert r["error"] is None
    assert r["diag_report"] == {"root_cause": "n"}
    assert r["pods"] == pods
    assert Path(r["diag_input_dir"], "worker0", "server-info.json").exists()
