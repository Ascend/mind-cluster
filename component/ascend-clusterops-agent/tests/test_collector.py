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

# pylint: disable=no-member,redefined-outer-name  # diag_pb2 dynamic members / pytest fixture names as params

from __future__ import annotations

import io
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


def test_run_commands_swallows_errors(tmp_path, client, command_stubs, monkeypatch):
    # whitelisted command fails with no output -> silently skipped, no files produced
    _install_stub(tmp_path, monkeypatch, "msnpureport", "exit 1")
    monkeypatch.setattr(collector, "ALLOWED_COMMANDS", frozenset({"msnpureport --bad"}))
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
    # the host side of a mount pair contains the plog keyword -> collect that host path (first hit wins)
    src = tmp_path / "var-log-plog"
    src.mkdir()
    (src / "p.log").write_text("plog")
    pairs = [f"{tmp_path}/other:/data", f"{src}:/var/log/plog"]
    fake_pm = SimpleNamespace(all_pairs=lambda uid: pairs)
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "p.log").read_text() == "plog"


def test_collect_mount_keywords_scans_subdirs(tmp_path, monkeypatch, client):
    # no pair carries the keyword: scan the host-path subdirs under /var/log for one containing plog
    host = tmp_path / "var-log"
    host.mkdir()
    (host / "plog").mkdir()
    (host / "plog" / "p.log").write_text("plog-sub")
    (host / "messages").write_text("ignored")
    pairs = [f"{host}:/var/log"]
    fake_pm = SimpleNamespace(all_pairs=lambda uid: pairs)
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "mount_keywords": ["plog"]}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._collect_mount_keywords(dst, entity, pods)
    assert (dst / "p.log").read_text() == "plog-sub"
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
# _gather: matches all_pairs against entity paths/mount_keywords + special-cases host_log
# --------------------------------------------------------------------------- #
def test_gather_env_falls_back_to_mount_keywords(tmp_path, monkeypatch, client):
    # process_log (env+mount_keywords): env not recorded (pod_env returns None) -> fall back to mount_keywords over all pairs
    src = tmp_path / "plog"
    src.mkdir()
    (src / "p.log").write_text("plog")
    fake_pm = SimpleNamespace(all_pairs=lambda uid: [f"{src}:/var/log/plog"], pod_env=lambda uid, n: None)
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    manifest = {"entities": [{"name": "process_log", "env": "ASCEND_PROCESS_LOG_PATH", "mount_keywords": ["plog"]}]}
    collect_dir = tmp_path / "collect"
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._gather(manifest, collect_dir, pods)
    assert (collect_dir / "process_log" / "p.log").read_text() == "plog"


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
    src.mkdir(parents=True)
    (src / "p.log").write_text("plog")
    fake_pm = _fake_pm([f"{tmp_path}/host:/var/log"], env="/var/log/mindx-dl/plog")
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)

    dst = tmp_path / "process_log"
    dst.mkdir()
    entity = {"name": "process_log", "env": "ASCEND_PROCESS_LOG_PATH"}
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    ok = client._collect_env_entity(dst, entity, pods)
    assert ok is True
    assert (dst / "p.log").read_text() == "plog"


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
    src.mkdir(parents=True)
    (src / "p.log").write_text("plog")
    fake_pm = _fake_pm([f"{tmp_path}/host:/var/log"], env="/var/log/plog")
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    manifest = {"entities": [{"name": "process_log", "env": "ASCEND_PROCESS_LOG_PATH", "mount_keywords": ["plog"]}]}
    collect_dir = tmp_path / "collect"
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._gather(manifest, collect_dir, pods)
    assert (collect_dir / "process_log" / "p.log").read_text() == "plog"


def test_gather_env_success_skips_commands(tmp_path, monkeypatch, client, command_stubs):
    # env is the highest priority: when env succeeds, the equal-priority commands must not run
    src = tmp_path / "host" / "plog"
    src.mkdir(parents=True)
    (src / "p.log").write_text("plog")
    fake_pm = _fake_pm([f"{tmp_path}/host:/var/log"], env="/var/log/plog")
    monkeypatch.setattr(collector, "get_pathmap", lambda: fake_pm)
    monkeypatch.setattr(collector, "ALLOWED_COMMANDS", frozenset({"dmesg > host_info.txt"}))
    manifest = {
        "entities": [{"name": "process_log", "env": "ASCEND_PROCESS_LOG_PATH", "commands": ["dmesg > host_info.txt"]}]
    }
    collect_dir = tmp_path / "collect"
    pods = [collector.PodRef(ns="ns", name="p", pod_uid="u0")]
    client._gather(manifest, collect_dir, pods)
    assert (collect_dir / "process_log" / "p.log").read_text() == "plog"
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
    assert captured["meta"] == {"node": "node-a", "job": "job-x", "namespace": "default"}
    assert "parse boom" in captured["error"]


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
    assert captured["meta"] == {"node": "node-a", "job": "job-x", "namespace": "default"}
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
        ok = collector.collector_client.upload_result({"node": node, "job": job, "namespace": "testns"}, buf.read())
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
        assert Path(res["artifacts_tar"]).name == "parse-result-job-e2e-node-a.tar.gz"
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
