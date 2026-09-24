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

"""node-collector + pathmap unit tests: resolver / pathmap / Collect gRPC (mocks ascend-fd)."""

# pylint: disable=no-member,redefined-outer-name,too-many-lines  # diag_pb2 dynamic members / pytest fixture names as params / long E2E file

from __future__ import annotations

import io
import logging
import os
import tarfile
import time
from pathlib import Path
from types import SimpleNamespace

import grpc
import pytest

from node_collector import collector
from diagproto import diag_pb2


@pytest.fixture
def client():
    """Collector executor instance (tests the internal collect/parse methods; default work_root, no uploads)."""
    return collector.CollectorClient()


def _install_stub(tmp_path, monkeypatch, name: str, body: str) -> Path:
    """Put an executable stub named <name> (sh script <body>) on PATH for real shell execution."""
    d = tmp_path / "stubs"
    d.mkdir(exist_ok=True)
    f = d / name
    f.write_text(f"#!/bin/sh\n{body}\n", encoding="utf-8")
    f.chmod(0o755)
    monkeypatch.setenv("PATH", f"{d}:{os.environ.get('PATH', '')}")
    return f


@pytest.fixture
def command_stubs(tmp_path, monkeypatch):
    """Install whitelisted command stubs (dmesg/dmidecode/msnpureport) onto PATH."""
    _install_stub(tmp_path, monkeypatch, "dmesg", 'printf "DMESG-OUT %s\\n" "$*"')
    _install_stub(tmp_path, monkeypatch, "dmidecode", 'printf "DMIDECODE-OUT %s\\n" "$*"')
    _install_stub(tmp_path, monkeypatch, "msnpureport", 'printf "MSNPUREPORT-OUT %s\\n" "$*"')


# --------------------------------------------------------------------------- #
# _collect_host_path / _run_commands (commands: run in a tmp dir shell + full copy + stdout capture)
# --------------------------------------------------------------------------- #
def test_collect_host_path_copies_files(tmp_path, client):
    src = tmp_path / "src"
    src.mkdir()
    (src / "a.log").write_text("hello")
    (src / "b.log").write_text("world")
    dst = tmp_path / "device_log"
    dst.mkdir()
    client._collect_host_path(dst, [str(src / "*.log")])
    assert (dst / "a.log").read_text() == "hello"
    assert (dst / "b.log").read_text() == "world"


def test_collect_host_path_preserves_empty_dirs(tmp_path, client):
    # whole-tree directory copy: empty subdirs (e.g. security/) must be kept even without files
    src = tmp_path / "plog"
    (src / "run").mkdir(parents=True)
    (src / "debug").mkdir()
    (src / "security").mkdir()
    (src / "run" / "a.log").write_text("hello")
    dst = tmp_path / "process_log"
    dst.mkdir()
    client._collect_host_path(dst, [str(src / "**")])
    assert (dst / "run" / "a.log").read_text() == "hello"
    assert (dst / "debug").is_dir()  # empty dirs are copied too
    assert (dst / "security").is_dir()  # empty dirs are copied too


def test_warn_if_no_files(tmp_path, client, caplog):
    # pattern matching nothing, matched dir without files, and normal file -> warnings for the first two only
    existing = tmp_path / "messages"
    existing.write_text("log")
    empty_dir = tmp_path / "empty-component"
    empty_dir.mkdir()
    patterns = [str(tmp_path / "no-such-file*"), str(tmp_path / "empty-component"), str(existing)]
    with caplog.at_level(logging.WARNING):
        client._warn_if_no_files(patterns, "host_log")
    warnings = [r.message for r in caplog.records if r.levelno == logging.WARNING]
    assert any("matches nothing" in w and "no-such-file" in w for w in warnings)
    assert any("contains no files" in w and "empty-component" in w for w in warnings)
    assert not any(str(existing) in w for w in warnings)


def test_run_commands_redirects_generated_file(tmp_path, client, command_stubs, monkeypatch):
    # whole-string whitelist hit: the redirect generates a file that is copied to dst
    monkeypatch.setattr(collector, "ALLOWED_COMMANDS", frozenset({"dmesg -T > dmesg.txt"}))
    dst = tmp_path / "host_log"
    dst.mkdir()
    client._run_commands(dst, ["dmesg -T > dmesg.txt"], "host_log")
    assert "DMESG-OUT" in (dst / "dmesg.txt").read_text()
    # the temp dir is cleaned up: nothing remains except the collect result
    assert not list(tmp_path.glob("ascend-fd-cmd-*"))


def test_run_commands_captures_stdout_without_redirect(tmp_path, client, command_stubs, monkeypatch):
    # command without a > redirect: stdout/stderr is captured as {name}_stdout.txt
    monkeypatch.setattr(collector, "ALLOWED_COMMANDS", frozenset({"dmidecode"}))
    dst = tmp_path / "device_log"
    dst.mkdir()
    client._run_commands(dst, ["dmidecode"], "device_log")
    assert "DMIDECODE-OUT" in (dst / "device_log_stdout.txt").read_text()


def test_run_commands_swallows_errors(tmp_path, client, monkeypatch):
    # whitelisted command returns no output -> silently skipped, no files produced
    monkeypatch.setattr(collector, "ALLOWED_COMMANDS", frozenset({"msnpureport --bad"}))
    # stub subprocess.run: a failed command with empty stdout/stderr must not emit a file.
    # Real shell execution is prone to process-level stderr noise (e.g. crashpad), so mock it here.
    monkeypatch.setattr(
        collector.subprocess,
        "run",
        lambda *a, **k: SimpleNamespace(returncode=1, stdout="", stderr=""),
    )
    dst = tmp_path / "device_log"
    dst.mkdir()
    client._run_commands(dst, ["msnpureport --bad"], "device_log")
    assert not any(dst.iterdir())


def test_run_commands_multiple_preserves_files(tmp_path, client, command_stubs, monkeypatch):
    # multiple whitelisted commands each produce their file; all are kept
    monkeypatch.setattr(collector, "ALLOWED_COMMANDS", frozenset({"dmesg > a.txt", "dmidecode > b.txt"}))
    dst = tmp_path / "host_log"
    dst.mkdir()
    client._run_commands(dst, ["dmesg > a.txt", "dmidecode > b.txt"], "host_log")
    assert "DMESG-OUT" in (dst / "a.txt").read_text()
    assert "DMIDECODE-OUT" in (dst / "b.txt").read_text()


def test_run_commands_skips_disallowed(tmp_path, client):
    # non-whitelisted commands are rejected and produce no files
    dst = tmp_path / "host_log"
    dst.mkdir()
    client._run_commands(dst, ["echo HAXX", "curl http://evil", "rm -rf /"], "host_log")
    assert not any(dst.iterdir())


def test_run_commands_whole_string_whitelist(tmp_path, client):
    # whole-string whitelist: tampered strings (extra &&/;/$()) never match, even if the first word is allowed
    dst = tmp_path / "host_log"
    dst.mkdir()
    client._run_commands(
        dst,
        [
            "/usr/bin/dmesg -T | tail -n 100000 > dmesg && touch pwned",
            "/usr/sbin/dmidecode > dmidecode.txt; touch pwned2",
            "/usr/bin/msnpureport --docker $(id)",
        ],
        "host_log",
    )
    assert not any(dst.iterdir())


# --------------------------------------------------------------------------- #
# _collect_paths: paths entities resolve the host path via the all_pairs container-side prefix match (mocks get_pathmap)
# --------------------------------------------------------------------------- #
def test_collect_paths_matches_container_paths(tmp_path, monkeypatch, client):
    src = tmp_path / "plog"
    src.mkdir()
    (src / "p.log").write_text("plog")
    pairs = [f"{src}:/home/hwMindX/plog", f"{tmp_path}/other:/data"]
    fake_pm = SimpleNamespace(all_pairs=lambda uid: pairs if uid == "u0" else [])
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "paths": ["/home/hwMindX/plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_paths(dst, "process_log", entity, pods)
    assert (dst / "p.log").read_text() == "plog"


def test_collect_paths_dl_log_reads_host_paths_directly(tmp_path, monkeypatch, client):
    # dl_log reads host paths directly, keeping the basename subdir: {collect_dir}/dl_log/noded/***.log.
    # Even if the task pod happens to mount a matching container path, dl_log must NOT use that
    # mount pair (a task pod mounting the same path would otherwise pollute dl_log).
    noded = tmp_path / "noded"
    noded.mkdir()
    (noded / "noded.log").write_text("noded-log")
    dp = tmp_path / "devicePlugin"
    dp.mkdir()
    (dp / "dp.log").write_text("dp-log")
    fake_pm = SimpleNamespace(all_pairs=lambda uid: [f"{tmp_path}/fake-plog:/var/log/mindx-dl/noded"])
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "dl_log"
    dst.mkdir()
    entity = {"name": "dl_log", "paths": [str(noded), str(dp)]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_paths(dst, "dl_log", entity, pods)
    assert (dst / "noded" / "noded.log").read_text() == "noded-log"
    assert (dst / "devicePlugin" / "dp.log").read_text() == "dp-log"
    assert not (dst / "fake-plog").exists()


def test_collect_paths_dl_log_ignores_pathmap(tmp_path, monkeypatch, client):
    # dl_log paths are host paths; the pathmap is not consulted at all for dl_log
    fake_pm = SimpleNamespace(all_pairs=lambda uid: [f"{tmp_path}/other:/var/log/mindx-dl/noded"])
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    host = tmp_path / "logs"
    host.mkdir()
    (host / "m.log").write_text("host-log")
    dst = tmp_path / "dl_log"
    dst.mkdir()
    entity = {"name": "dl_log", "paths": [str(host)]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_paths(dst, "dl_log", entity, pods)
    assert (dst / "logs" / "m.log").read_text() == "host-log"
    assert not (dst / "other").exists()


# --------------------------------------------------------------------------- #
# _collect_mount_keywords: keyword match (first hit wins) + subdir scan
# --------------------------------------------------------------------------- #
def test_collect_mount_keywords_matches_pairs(tmp_path, monkeypatch, client):
    # the host side of a mount pair contains the plog keyword -> collect that plog dir's
    # run/debug/security subdirs (first hit wins, top-level files skipped)
    src = tmp_path / "var-log-plog"
    (src / "run").mkdir(parents=True)
    (src / "debug").mkdir()
    (src / "security").mkdir()
    (src / "run" / "p.log").write_text("plog")
    (src / "plog-113_123.log").write_text("dup")  # top-level duplicate, must be skipped
    pairs = [f"{tmp_path}/other:/data", f"{src}:/var/log/plog"]
    fake_pm = SimpleNamespace(all_pairs=lambda uid: pairs)
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "run" / "p.log").read_text() == "plog"
    assert not (dst / "plog-113_123.log").exists()


def test_collect_mount_keywords_scans_subdirs(tmp_path, monkeypatch, client):
    # no pair carries the keyword: scan the host-path subdirs under /var/log for one containing plog
    host = tmp_path / "var-log"
    host.mkdir()
    (host / "plog" / "run").mkdir(parents=True)
    (host / "plog" / "debug").mkdir()
    (host / "plog" / "security").mkdir()
    (host / "plog" / "run" / "p.log").write_text("plog-sub")
    (host / "messages").write_text("ignored")
    pairs = [f"{host}:/var/log"]
    fake_pm = SimpleNamespace(all_pairs=lambda uid: pairs)
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "run" / "p.log").read_text() == "plog-sub"
    assert not (dst / "messages").exists()


def test_collect_mount_keywords_no_pairs_noop(tmp_path, monkeypatch, client):
    fake_pm = SimpleNamespace(all_pairs=lambda uid: [])
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert not list(dst.iterdir())


# --------------------------------------------------------------------------- #
# _collect_mount_keywords: static shared-storage mount (protocol-agnostic, survives pod deletion)
# --------------------------------------------------------------------------- #
def test_collect_mount_keywords_shared_storage(tmp_path, monkeypatch, client):
    # the site statically mounts the shared storage at SHARED_STORAGE_ROOT (NFS or any
    # CSI-backed protocol); the collector reads the mounted dir directly (no mount command),
    # narrowed to this pod by task id + pod ip (alllogs/<task_id>/plogs/<ip>), then copies
    # the plog dir's run/debug/security subdirs
    shared = tmp_path / "shared-storage"
    other = shared / "alllogs" / "task-OTHER" / "plogs" / "9.9.9.9"
    mine = shared / "alllogs" / "task-42" / "plogs" / "10.0.0.5"
    (other / "run").mkdir(parents=True)
    (mine / "run").mkdir(parents=True)
    (mine / "debug").mkdir()
    (mine / "security").mkdir()
    (other / "run" / "o.log").write_text("other-task")
    (mine / "run" / "p.log").write_text("my-plog")
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", shared)
    fake_pm = SimpleNamespace(
        all_pairs=lambda uid: [],
        pod_env=lambda uid, name: "task-42" if (uid, name) == ("u0", "MINDX_TASK_ID") else None,
        pod_ip=lambda uid: "10.0.0.5" if uid == "u0" else "",
    )
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "run" / "p.log").read_text() == "my-plog"  # own task/pod plog collected
    assert (dst / "debug").is_dir()  # plog dir's subdirs land directly under process_log
    assert (dst / "security").is_dir()
    assert not (dst / "run" / "o.log").exists()  # another task's plog skipped


def test_collect_mount_keywords_shared_storage_reversed_layout(tmp_path, monkeypatch, client):
    # reversed layout alllogs/<ip>/plogs/<task>: the ip scopes this pod, take the plog dir;
    # the single <task> level is descended so run/debug/security land directly under process_log
    shared = tmp_path / "shared-storage"
    log_dir = shared / "alllogs" / "10.0.0.5" / "plogs" / "task-42"
    (log_dir / "run").mkdir(parents=True)
    (log_dir / "debug").mkdir()
    (log_dir / "security").mkdir()
    (log_dir / "run" / "p.log").write_text("reversed-layout")
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", shared)
    fake_pm = SimpleNamespace(
        all_pairs=lambda uid: [],
        pod_env=lambda uid, name: "task-42" if (uid, name) == ("u0", "MINDX_TASK_ID") else None,
        pod_ip=lambda uid: "10.0.0.5" if uid == "u0" else "",
    )
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "run" / "p.log").read_text() == "reversed-layout"
    assert not (dst / "task-42").exists()  # the extra <task> level is stripped


def test_collect_mount_keywords_shared_storage_ip_only_layout(tmp_path, monkeypatch, client):
    # layout alllogs/plog/<ip>: only the pod ip occupies the directory below the keyword
    # dir (no task id in the path) -> narrowed by the direct ip subdir (ip fallback, no task id)
    shared = tmp_path / "shared-storage"
    other = shared / "job" / "code" / "alllogs" / "plog" / "9.9.9.9"
    mine = shared / "job" / "code" / "alllogs" / "plog" / "10.0.0.5"
    (other / "run").mkdir(parents=True)
    (mine / "run").mkdir(parents=True)
    (mine / "debug").mkdir()
    (mine / "security").mkdir()
    (other / "run" / "o.log").write_text("other-pod")
    (mine / "run" / "p.log").write_text("my-plog")
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", shared)
    fake_pm = SimpleNamespace(
        all_pairs=lambda uid: [],
        pod_env=lambda uid, name: None,  # task id not recorded -> ip-only narrowing
        pod_ip=lambda uid: "10.0.0.5" if uid == "u0" else "",
    )
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "run" / "p.log").read_text() == "my-plog"  # own pod plog collected
    assert not (dst / "run" / "o.log").exists()  # another pod's plog skipped


def test_collect_mount_keywords_shared_storage_scoped_by_task_id(tmp_path, monkeypatch, client):
    # hostNetwork pods share the node ip across tasks: the ip subdir alone matches every
    # task that ran on the node; with a known task id only this task's dir is collected
    shared = tmp_path / "shared-storage"
    mine = shared / "alllogs" / "task-42" / "plogs" / "10.0.0.5"
    other = shared / "alllogs" / "task-OTHER" / "plogs" / "10.0.0.5"  # same ip, other task
    (mine / "run").mkdir(parents=True)
    (other / "run").mkdir(parents=True)
    (mine / "debug").mkdir()
    (mine / "security").mkdir()
    (mine / "run" / "p.log").write_text("my-plog")
    (other / "run" / "o.log").write_text("other-task")
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", shared)
    fake_pm = SimpleNamespace(
        all_pairs=lambda uid: [],
        pod_env=lambda uid, name: "task-42" if (uid, name) == ("u0", "MINDX_TASK_ID") else None,
        pod_ip=lambda uid: "10.0.0.5" if uid == "u0" else "",
    )
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "run" / "p.log").read_text() == "my-plog"  # own task collected
    assert not (dst / "run" / "o.log").exists()  # same-ip other-task dir excluded


def test_collect_mount_keywords_shared_storage_single_candidate_ip_fallback(tmp_path, monkeypatch, client):
    # task-less layout /job/code/plogs/<ip>: the task id appears nowhere; a single
    # ip-narrowed candidate falls back to the ip (the site does not organize by task)
    shared = tmp_path / "shared-storage"
    mine = shared / "job" / "code" / "plogs" / "10.0.0.5"
    (mine / "run").mkdir(parents=True)
    (mine / "debug").mkdir()
    (mine / "security").mkdir()
    (mine / "run" / "p.log").write_text("my-plog")
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", shared)
    fake_pm = SimpleNamespace(
        all_pairs=lambda uid: [],
        pod_env=lambda uid, name: "task-42" if (uid, name) == ("u0", "MINDX_TASK_ID") else None,
        pod_ip=lambda uid: "10.0.0.5" if uid == "u0" else "",
    )
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "run" / "p.log").read_text() == "my-plog"  # single candidate collected by ip


def test_collect_mount_keywords_shared_storage_bare_plog_dir(tmp_path, monkeypatch, client):
    # bare layout /job/code/plogs/{run,debug,security}: no task/ip layer (the pod's own
    # host mount, ip-isolated) -> the keyword dir itself is this pod's and collected
    shared = tmp_path / "shared-storage"
    mine = shared / "job" / "code" / "plogs"
    (mine / "run").mkdir(parents=True)
    (mine / "debug").mkdir()
    (mine / "security").mkdir()
    (mine / "run" / "p.log").write_text("my-plog")
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", shared)
    fake_pm = SimpleNamespace(
        all_pairs=lambda uid: [],
        pod_env=lambda uid, name: "task-42" if (uid, name) == ("u0", "MINDX_TASK_ID") else None,
        pod_ip=lambda uid: "10.0.0.5" if uid == "u0" else "",
    )
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "run" / "p.log").read_text() == "my-plog"
    assert (dst / "debug").is_dir()
    assert (dst / "security").is_dir()


def test_collect_mount_keywords_shared_storage_multi_candidate_no_match_noop(tmp_path, monkeypatch, client):
    # multiple same-ip candidates under different task dirs and none carries our task id:
    # ambiguous which task owns them -> nothing is collected
    shared = tmp_path / "shared-storage"
    a = shared / "alllogs" / "task-A" / "plogs" / "10.0.0.5"
    b = shared / "alllogs" / "task-B" / "plogs" / "10.0.0.5"
    (a / "run").mkdir(parents=True)
    (b / "run").mkdir(parents=True)
    (a / "run" / "a.log").write_text("task-a")
    (b / "run" / "b.log").write_text("task-b")
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", shared)
    fake_pm = SimpleNamespace(
        all_pairs=lambda uid: [],
        pod_env=lambda uid, name: "task-42" if (uid, name) == ("u0", "MINDX_TASK_ID") else None,
        pod_ip=lambda uid: "10.0.0.5" if uid == "u0" else "",
    )
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert not list(dst.iterdir())  # ambiguous -> no plog collected


def test_collect_mount_keywords_shared_storage_arbitrary_prefix(tmp_path, monkeypatch, client):
    # the shared-storage prefix is site-specific (e.g. /job/code), not fixed to alllogs:
    # any path carrying the plogs keyword dir + task id + pod ip is collected
    shared = tmp_path / "shared-storage"
    other = shared / "job" / "code" / "logs" / "task-OTHER" / "plogs" / "9.9.9.9"
    mine = shared / "job" / "code" / "logs" / "task-42" / "plogs" / "10.0.0.5"
    (other / "run").mkdir(parents=True)
    (mine / "run").mkdir(parents=True)
    (mine / "debug").mkdir()
    (mine / "security").mkdir()
    (other / "run" / "o.log").write_text("other-task")
    (mine / "run" / "p.log").write_text("my-plog")
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", shared)
    fake_pm = SimpleNamespace(
        all_pairs=lambda uid: [],
        pod_env=lambda uid, name: "task-42" if (uid, name) == ("u0", "MINDX_TASK_ID") else None,
        pod_ip=lambda uid: "10.0.0.5" if uid == "u0" else "",
    )
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "run" / "p.log").read_text() == "my-plog"  # own task/pod plog collected
    assert not (dst / "run" / "o.log").exists()  # another task's plog skipped


def test_collect_mount_keywords_shared_storage_not_mounted_noop(tmp_path, monkeypatch, client):
    # SHARED_STORAGE_ROOT absent (shared storage not mounted): nothing is collected
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", tmp_path / "not-mounted")
    fake_pm = SimpleNamespace(
        all_pairs=lambda uid: [],
        pod_env=lambda uid, name: "task-42" if (uid, name) == ("u0", "MINDX_TASK_ID") else None,
        pod_ip=lambda uid: "10.0.0.5" if uid == "u0" else "",
    )
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert not list(dst.iterdir())


# --------------------------------------------------------------------------- #
# _gather: matches all_pairs against entity paths/mount_keywords + special-cases host_log
# --------------------------------------------------------------------------- #
def test_gather_env_falls_back_to_mount_keywords(tmp_path, monkeypatch, client):
    # process_log (env+mount_keywords): env not recorded (pod_env returns None) -> fall back
    # to mount_keywords over all pairs; plog dir's run/debug/security subdirs are collected
    src = tmp_path / "plog"
    (src / "run").mkdir(parents=True)
    (src / "debug").mkdir()
    (src / "security").mkdir()
    (src / "run" / "p.log").write_text("plog")
    fake_pm = SimpleNamespace(all_pairs=lambda uid: [f"{src}:/var/log/plog"], pod_env=lambda uid, n: None)
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    manifest = {"entities": [{"name": "process_log", "env": "ASCEND_PROCESS_LOG_PATH", "mount_keywords": ["plog"]}]}
    collect_dir = tmp_path / "collect"
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._gather(manifest, collect_dir, pods)
    assert (collect_dir / "process_log" / "run" / "p.log").read_text() == "plog"


# --------------------------------------------------------------------------- #
# env mode: node-collector reads the pod env from pathmap (recorded by agent-core) and resolves the host path
# --------------------------------------------------------------------------- #
def _fake_pm(pairs, env=None):
    """Pathmap stand-in with all_pairs + pod_env."""
    return SimpleNamespace(all_pairs=lambda uid: pairs, pod_env=lambda uid, name: env)


def test_resolve_container_to_host_longest_prefix(tmp_path, client):
    pairs = [f"{tmp_path}/host:/var/log", f"{tmp_path}/app:/var/log/mindx-dl"]
    # /var/log/mindx-dl/plog -> the most specific mount /var/log/mindx-dl wins
    host = client._resolve_container_to_host(pairs, "/var/log/mindx-dl/plog")
    assert host == str(tmp_path / "app" / "plog")
    # /var/log/messages -> only /var/log matches
    host2 = client._resolve_container_to_host(pairs, "/var/log/messages")
    assert host2 == str(tmp_path / "host" / "messages")
    assert client._resolve_container_to_host(pairs, "/etc/passwd") is None


def test_collect_env_entity_reads_pod_env(tmp_path, monkeypatch, client):
    # node-collector reads the pod env from pathmap -> resolves the host path (nested subdirs) -> collects
    src = tmp_path / "host" / "mindx-dl" / "plog"
    (src / "run").mkdir(parents=True)
    (src / "debug").mkdir()
    (src / "security").mkdir()
    (src / "run" / "p.log").write_text("plog")
    fake_pm = _fake_pm([f"{tmp_path}/host:/var/log"], env="/var/log/mindx-dl/plog")
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "env": "ASCEND_PROCESS_LOG_PATH", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    ok = client._collect_env_entity(dst, entity, pods)
    assert ok is True
    assert (dst / "run" / "p.log").read_text() == "plog"


def test_glob_base():
    assert collector.CollectorClient._glob_base("/var/log/mindx-dl/plog/**") == "/var/log/mindx-dl/plog"
    assert collector.CollectorClient._glob_base("/var/log/messages*") == "/var/log"
    assert collector.CollectorClient._glob_base(str(Path("/tmp/src/*.log"))) == "/tmp/src"


def test_collect_host_path_preserves_plog_subdirs(tmp_path, client):
    # plog contains run/debug/security subdirs -> the subdir structure is preserved after collection
    plog = tmp_path / "plog"
    for sub in ("run", "debug", "security"):
        d = plog / sub
        d.mkdir(parents=True)
        (d / f"{sub}.log").write_text(sub)
    dst = tmp_path / "process_log"
    dst.mkdir()
    client._collect_host_path(dst, [str(plog / "**")])
    assert (dst / "run" / "run.log").read_text() == "run"
    assert (dst / "debug" / "debug.log").read_text() == "debug"
    assert (dst / "security" / "security.log").read_text() == "security"


# --------------------------------------------------------------------------- #
# plog special handling: _collect_subdirs copies run/debug/security (top-level files skipped),
# a single-subdir chain (e.g. <jobname>) is descended so the plog subdirs land directly
# --------------------------------------------------------------------------- #
def test_collect_subdirs_copies_subdirs_only(tmp_path, client):
    # top-level plog files duplicate run/plog + debug/plog -> only the subdirs are kept
    src = tmp_path / "plog"
    (src / "run" / "plog").mkdir(parents=True)
    (src / "debug" / "plog").mkdir(parents=True)
    (src / "security").mkdir()
    (src / "plog-113_20260918105102886.log").write_text("dup-top")
    (src / "run" / "plog" / "p.log").write_text("run-plog")
    (src / "debug" / "plog" / "d.log").write_text("debug-plog")
    dst = tmp_path / "process_log"
    dst.mkdir()
    client._collect_subdirs(dst, str(src))
    assert (dst / "run" / "plog" / "p.log").read_text() == "run-plog"
    assert (dst / "debug" / "plog" / "d.log").read_text() == "debug-plog"
    assert (dst / "security").is_dir()
    assert not (dst / "plog-113_20260918105102886.log").exists()  # top-level duplicate skipped


def test_collect_subdirs_descends_single_subdir(tmp_path, client):
    # plog under one extra job-level dir: run/debug/security land directly under dst
    root = tmp_path / "plog" / "jobname"
    (root / "run").mkdir(parents=True)
    (root / "debug").mkdir()
    (root / "security").mkdir()
    (root / "run" / "p.log").write_text("x")
    dst = tmp_path / "process_log"
    dst.mkdir()
    assert client._collect_subdirs(dst, str(tmp_path / "plog")) is True
    assert (dst / "run" / "p.log").read_text() == "x"
    assert (dst / "debug").is_dir()
    assert (dst / "security").is_dir()
    assert not (dst / "jobname").exists()  # the extra level is stripped


def test_collect_plog_dir_warns_when_matched_dir_empty(tmp_path, client, caplog):
    # matched plog dir holds no run/debug/security -> warning names the dir
    src = tmp_path / "plogs"
    src.mkdir()
    (src / "not-plog.txt").write_text("stray file")
    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pod = collector.PodRef(ns="ns", name="p", pod_uid="u0")
    with caplog.at_level(logging.WARNING):
        client._collect_plog_dir(dst, str(src), entity, pod)
    assert any("no plog logs" in rec.message and str(src) in rec.message for rec in caplog.records)


def test_collect_mount_keywords_plog_no_match_warns(tmp_path, monkeypatch, client, caplog):
    # plog entity: no mount pair, no shared-storage hit -> collect-failed warning
    shared = tmp_path / "shared-storage"
    shared.mkdir()
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", shared)
    fake_pm = SimpleNamespace(all_pairs=lambda uid: [], pod_env=lambda uid, name: None, pod_ip=lambda uid: "")
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    with caplog.at_level(logging.WARNING):
        client._collect_mount_keywords(dst, entity, pods)
    assert any("collect plog logs failed" in rec.message for rec in caplog.records)


def test_collect_mount_keywords_shared_storage_prunes_nested_plog(tmp_path, monkeypatch, client):
    # nested keyword dirs (.../plogs/<ip>/run/plog) are pruned: only the ip dir's subdirs
    # are collected once, no flattening of the nested plog dir as a separate hit
    shared = tmp_path / "shared-storage"
    mine = shared / "alllogs" / "task-42" / "plogs" / "10.0.0.5"
    (mine / "run" / "plog").mkdir(parents=True)
    (mine / "debug").mkdir()
    (mine / "security").mkdir()
    (mine / "run" / "plog" / "p.log").write_text("my-plog")
    monkeypatch.setattr(collector, "SHARED_STORAGE_ROOT", shared)
    fake_pm = SimpleNamespace(
        all_pairs=lambda uid: [],
        pod_env=lambda uid, name: "task-42" if (uid, name) == ("u0", "MINDX_TASK_ID") else None,
        pod_ip=lambda uid: "10.0.0.5" if uid == "u0" else "",
    )
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "run" / "plog" / "p.log").read_text() == "my-plog"
    assert not (dst / "p.log").exists()  # nested plog dir not flattened to the dst root


def test_collect_env_entity_falls_back_when_no_value(tmp_path, monkeypatch, client):
    # pod env not recorded (e.g. task deleted) -> returns False (the caller falls back to mount_keywords)
    fake_pm = _fake_pm([f"{tmp_path}/host:/var/log"], env=None)
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "env": "ASCEND_PROCESS_LOG_PATH"}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    assert client._collect_env_entity(dst, entity, pods) is False


def test_gather_env_entity_prefers_env(tmp_path, monkeypatch, client):
    # process_log (env) resolves via env; a host side without the plog keyword is fine
    src = tmp_path / "host" / "plog"
    (src / "run").mkdir(parents=True)
    (src / "debug").mkdir()
    (src / "security").mkdir()
    (src / "run" / "p.log").write_text("plog")
    fake_pm = _fake_pm([f"{tmp_path}/host:/var/log"], env="/var/log/plog")
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    manifest = {"entities": [{"name": "process_log", "env": "ASCEND_PROCESS_LOG_PATH", "mount_keywords": ["plog"]}]}
    collect_dir = tmp_path / "collect"
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._gather(manifest, collect_dir, pods)
    assert (collect_dir / "process_log" / "run" / "p.log").read_text() == "plog"


def test_gather_env_success_skips_commands(tmp_path, monkeypatch, client, command_stubs):
    # env is the highest priority: when env succeeds, the equal-priority commands must not run
    src = tmp_path / "host" / "plog"
    (src / "run").mkdir(parents=True)
    (src / "debug").mkdir()
    (src / "security").mkdir()
    (src / "run" / "p.log").write_text("plog")
    fake_pm = _fake_pm([f"{tmp_path}/host:/var/log"], env="/var/log/plog")
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    monkeypatch.setattr(collector, "ALLOWED_COMMANDS", frozenset({"dmesg > host_info.txt"}))
    manifest = {
        "entities": [{"name": "process_log", "env": "ASCEND_PROCESS_LOG_PATH", "commands": ["dmesg > host_info.txt"]}]
    }
    collect_dir = tmp_path / "collect"
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._gather(manifest, collect_dir, pods)
    assert (collect_dir / "process_log" / "run" / "p.log").read_text() == "plog"
    assert not (collect_dir / "process_log" / "host_info.txt").exists()


def test_gather_env_failure_runs_commands(tmp_path, monkeypatch, client, command_stubs):
    # env not recorded -> lower-priority fields run, including the equal-priority commands
    fake_pm = _fake_pm([f"{tmp_path}/host:/var/log"], env=None)
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    monkeypatch.setattr(collector, "ALLOWED_COMMANDS", frozenset({"dmesg > host_info.txt"}))
    manifest = {
        "entities": [{"name": "process_log", "env": "ASCEND_PROCESS_LOG_PATH", "commands": ["dmesg > host_info.txt"]}]
    }
    collect_dir = tmp_path / "collect"
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._gather(manifest, collect_dir, pods)
    assert "DMESG-OUT" in (collect_dir / "process_log" / "host_info.txt").read_text()


def test_gather_paths_falls_back_to_host_paths(tmp_path, monkeypatch, client):
    # dl_log task pod has no matching mount pair -> fall back to reading host paths (subdir per basename)
    fake_pm = SimpleNamespace(all_pairs=lambda uid: [])
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    src = tmp_path / "host"
    src.mkdir()
    (src / "messages.log").write_text("syslog")
    manifest = {"entities": [{"name": "dl_log", "paths": [str(src)]}]}
    collect_dir = tmp_path / "collect"
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._gather(manifest, collect_dir, pods)
    assert (collect_dir / "dl_log" / "host" / "messages.log").read_text() == "syslog"


def test_gather_host_log_command_and_host_paths(tmp_path, monkeypatch, client, command_stubs):
    # host_log: commands + direct host-path reads, independent of pathmap
    src = tmp_path / "host"
    src.mkdir()
    (src / "messages.log").write_text("syslog")
    fake_pm = SimpleNamespace(all_pairs=lambda uid: [])
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    monkeypatch.setattr(collector, "ALLOWED_COMMANDS", frozenset({"dmesg > host_info.txt"}))
    manifest = {
        "entities": [{"name": "host_log", "commands": ["dmesg > host_info.txt"], "paths": [str(src / "*.log")]}]
    }
    collect_dir = tmp_path / "collect"
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._gather(manifest, collect_dir, pods)
    assert (collect_dir / "host_log" / "messages.log").read_text() == "syslog"
    assert "DMESG-OUT" in (collect_dir / "host_log" / "host_info.txt").read_text()


def test_gather_command_only_entity_runs_commands(tmp_path, monkeypatch, client, command_stubs):
    # command-only entity (device_log, no env/paths/mount_keywords): runs its commands
    fake_pm = SimpleNamespace(all_pairs=lambda uid: [])
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    monkeypatch.setattr(collector, "ALLOWED_COMMANDS", frozenset({"dmidecode > npu_info.txt"}))
    manifest = {"entities": [{"name": "device_log", "commands": ["dmidecode > npu_info.txt"]}]}
    collect_dir = tmp_path / "collect"
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._gather(manifest, collect_dir, pods)
    assert "DMIDECODE-OUT" in (collect_dir / "device_log" / "npu_info.txt").read_text()


# --------------------------------------------------------------------------- #
# async collect + upload: TriggerCollect returns immediately, collects in the background, reports via UploadResult
# --------------------------------------------------------------------------- #
def _fake_subprocess_run():
    import subprocess as _sp

    _real = _sp.run

    def _run(cmd, check=False, **kwargs):
        if isinstance(cmd, list) and "parse" in cmd and "-o" in cmd:
            out = Path(cmd[cmd.index("-o") + 1])
            (out / "server-info.json").write_text("{}")
            (out / "ascend-rc-parser.json").write_text('{"root":"n"}')
            return SimpleNamespace(returncode=0, stdout="", stderr="")
        return _real(cmd, check=check, **kwargs)

    return _run


class _Ctx:
    """Mock gRPC context: abort raises so the failure branch is assertable in tests."""

    def __init__(self):
        self.aborted = None

    def abort(self, code, detail):
        self.aborted = (code, detail)
        raise grpc.RpcError(detail)


def _base_manifest(tmp_path):
    src = tmp_path / "device"
    src.mkdir()
    (src / "d.log").write_text("dl")
    return {
        "entities": [
            {"name": "host_log", "paths": [str(src / "*.log")]},  # direct host-path read
            {"name": "device_log", "commands": ["/usr/bin/msnpureport --docker"]},
        ]
    }


def test_collect_tar_produces_archive(tmp_path, monkeypatch):
    """Collect + parse into a tar.gz; the content includes server-info.json / ascend-rc-parser.json."""
    monkeypatch.setattr(collector, "_manifest", _base_manifest(tmp_path))
    monkeypatch.setattr(collector.subprocess, "run", _fake_subprocess_run())

    client = collector.CollectorClient(work_root=tmp_path)
    data = client.collect("node-a", "default", "job-x", [collector.PodRef(ns="default", name="p", pod_uid="u")])
    assert data
    with tarfile.open(fileobj=io.BytesIO(data), mode="r:gz") as tf:
        names = tf.getnames()
    assert any("server-info.json" in n for n in names)
    assert any("ascend-rc-parser.json" in n for n in names)


def test_trigger_collect_spawns_async(monkeypatch):
    """TriggerCollect returns accepted immediately; a background thread takes over collect + upload."""
    captured = {}
    monkeypatch.setattr(collector.collector_client, "run_and_upload", lambda req: captured.update(req=req))
    req = diag_pb2.CollectRequest(
        node="node-a",
        job="job-x",
        pods=[diag_pb2.PodRef(ns="default", name="p", pod_uid="u")],
    )
    ack = collector.CollectorServicer().TriggerCollect(req, _Ctx())
    assert ack.accepted is True
    for _ in range(50):
        if "req" in captured:
            break
        time.sleep(0.05)
    assert captured["req"].node == "node-a"
    assert captured["req"].job == "job-x"


def test_run_and_upload_reports_failure(tmp_path, monkeypatch):
    """parse failure -> the background thread reports the error (nothing is written to disk)."""
    captured = {}
    monkeypatch.setattr(collector.collector_client, "work_root", tmp_path)
    monkeypatch.setattr(collector, "_manifest", {"entities": []})
    monkeypatch.setattr(
        collector.collector_client, "upload_failure", lambda meta, error: captured.update(meta=meta, error=error)
    )

    def _run(cmd, **kwargs):
        raise collector.subprocess.CalledProcessError(returncode=1, cmd=cmd, output="", stderr="parse boom")

    monkeypatch.setattr(collector.subprocess, "run", _run)

    req = diag_pb2.CollectRequest(node="node-a", job="job-x", pods=[])
    collector.collector_client.run_and_upload(req)
    assert captured["meta"] == {"node": "node-a", "job": "job-x", "namespace": "default", "host_ip": ""}
    assert "parse boom" in captured["error"]


def test_run_and_upload_meta_host_ip_from_env_and_mounts(tmp_path, monkeypatch):
    """run_and_upload reports the collector host IP: HOST_IP env wins, mounts.pod_ip is the fallback."""
    captured = {}
    monkeypatch.setattr(collector.collector_client, "work_root", tmp_path)
    monkeypatch.setattr(
        collector.collector_client, "upload_failure", lambda meta, error: captured.update(meta=meta, error=error)
    )
    monkeypatch.setattr(
        collector.subprocess,
        "run",
        lambda *a, **k: (_ for _ in ()).throw(
            collector.subprocess.CalledProcessError(returncode=1, cmd=[], output="", stderr="boom")
        ),
    )
    req = diag_pb2.CollectRequest(
        node="node-a",
        job="job-x",
        pods=[],
        mounts=[diag_pb2.PodMount(pod_uid="u", pod_ip="10.0.0.5")],
    )
    collector.collector_client.run_and_upload(req)
    assert captured["meta"]["host_ip"] == "10.0.0.5"  # falls back to the dispatched mount pod_ip


def test_run_and_upload_meta_host_ip_prefers_env(monkeypatch):
    """When HOST_IP is injected via downward API, it beats the mount fallback."""
    captured = {}
    monkeypatch.setenv("HOST_IP", "172.16.0.9")
    monkeypatch.setattr(collector.collector_client, "collect", lambda node, ns, job, pods: b"fake-data")
    monkeypatch.setattr(
        collector.collector_client, "upload_result", lambda meta, data: captured.update(meta=meta) or True
    )
    req = diag_pb2.CollectRequest(
        node="node-a",
        job="job-x",
        pods=[],
        mounts=[diag_pb2.PodMount(pod_uid="u", pod_ip="10.0.0.5")],
    )
    collector.collector_client.run_and_upload(req)
    assert captured["meta"]["host_ip"] == "172.16.0.9"  # HOST_IP env wins over the mount fallback


def test_run_and_upload_reports_upload_failure(tmp_path, monkeypatch):
    """upload_result returns False -> report failure so agent-core does not wait for the collect timeout."""
    captured = {}
    monkeypatch.setattr(collector.collector_client, "work_root", tmp_path)
    monkeypatch.setattr(collector.collector_client, "collect", lambda node, ns, job, pods: b"fake-data")
    monkeypatch.setattr(collector.collector_client, "upload_result", lambda meta, data: False)
    monkeypatch.setattr(
        collector.collector_client, "upload_failure", lambda meta, error: captured.update(meta=meta, error=error)
    )

    req = diag_pb2.CollectRequest(node="node-a", job="job-x", pods=[])
    collector.collector_client.run_and_upload(req)
    assert captured["meta"] == {"node": "node-a", "job": "job-x", "namespace": "default", "host_ip": ""}
    assert captured["error"] == "upload failed after retries"


# --------------------------------------------------------------------------- #
# upload path E2E: collector.upload_result -> agent.upload.UploaderServicer
# (real loopback gRPC, tar.gz stored and extracted, JobTracker aggregates)
# --------------------------------------------------------------------------- #
def _start_upload_server(tmp_path, monkeypatch):
    from concurrent.futures import ThreadPoolExecutor

    import agent_core.upload as au
    from diagproto import diag_pb2_grpc

    monkeypatch.setattr(au, "WORK_ROOT", tmp_path)
    server = grpc.server(ThreadPoolExecutor(max_workers=2))
    diag_pb2_grpc.add_UploaderServicer_to_server(au.UploaderServicer(), server)
    port = server.add_insecure_port("127.0.0.1:0")
    server.start()
    monkeypatch.setattr(collector.collector_client, "agent_addr", f"127.0.0.1:{port}")
    return server


def test_upload_result_end_to_end(tmp_path, monkeypatch):
    from agent_core.jobs import tracker

    job, node = "job-e2e", "node-a"
    tracker.register(job, [node])
    from agent_core import tools

    tools._active_diags[("testns", job)] = time.time()
    server = _start_upload_server(tmp_path, monkeypatch)
    try:
        buf = io.BytesIO()
        with tarfile.open(fileobj=buf, mode="w:gz") as tf:
            info = tarfile.TarInfo("server-info.json")
            payload = b'{"ok":true}'
            info.size = len(payload)
            tf.addfile(info, io.BytesIO(payload))
        buf.seek(0)
        ok = collector.collector_client.upload_result(
            {"node": node, "job": job, "namespace": "testns", "host_ip": "10.0.0.5"}, buf.read()
        )
        assert ok is True
        st = tracker.get(job)
        res = st["results"][node]
        assert res["ok"] is True
        worker = Path(res["worker_dir"])
        assert (worker / "server-info.json").exists()
        # organized by ns_job: .../{YYYYMMDD}/testns_{job}/diag-input/worker-{node}
        assert "testns_job-e2e" in worker.parts
        assert "diag-input" in worker.parts
        assert worker.name == "worker-node-a"
        assert Path(res["artifacts_tar"]).name == "parse-result-testns-job-e2e-10.0.0.5.tar.gz"
    finally:
        server.stop(0)


def test_upload_failure_records_error(tmp_path, monkeypatch):
    from agent_core.jobs import tracker

    job, node = "job-fail", "node-b"
    tracker.register(job, [node])
    from agent_core import tools

    tools._active_diags[("default", job)] = time.time()
    server = _start_upload_server(tmp_path, monkeypatch)
    try:
        collector.collector_client.upload_failure({"node": node, "job": job}, "parse boom")
        res = tracker.get(job)["results"][node]
        assert res["ok"] is False
        assert "parse boom" in res["error"]
    finally:
        server.stop(0)


def test_upload_rejects_non_active_job(tmp_path, monkeypatch):
    from diagproto import diag_pb2_grpc

    server = _start_upload_server(tmp_path, monkeypatch)
    try:
        with pytest.raises(grpc.RpcError) as exc:
            with grpc.insecure_channel(collector.collector_client.agent_addr) as ch:
                diag_pb2_grpc.UploaderStub(ch).UploadResult(
                    collector.collector_client._upload_stream(
                        {"node": "node-c", "job": "job-ghost", "namespace": "testns", "ok": True}, b"data"
                    ),
                    timeout=5,
                )
        assert exc.value.code() == grpc.StatusCode.PERMISSION_DENIED
    finally:
        server.stop(0)


def test_upload_rejects_path_traversal(tmp_path, monkeypatch):
    from agent_core import tools
    from diagproto import diag_pb2_grpc

    job, node = "job-trav", "node-d"
    tools._active_diags[("testns", job)] = time.time()
    server = _start_upload_server(tmp_path, monkeypatch)
    try:
        buf = io.BytesIO()
        with tarfile.open(fileobj=buf, mode="w:gz") as tf:
            info = tarfile.TarInfo("../evil.txt")
            payload = b"evil"
            info.size = len(payload)
            tf.addfile(info, io.BytesIO(payload))
        buf.seek(0)
        with pytest.raises(grpc.RpcError) as exc:
            with grpc.insecure_channel(collector.collector_client.agent_addr) as ch:
                diag_pb2_grpc.UploaderStub(ch).UploadResult(
                    collector.collector_client._upload_stream(
                        {"node": node, "job": job, "namespace": "testns", "ok": True}, buf.read()
                    ),
                    timeout=5,
                )
        assert exc.value.code() == grpc.StatusCode.INTERNAL
        assert not (tmp_path.parent / "evil.txt").exists()
    finally:
        server.stop(0)
