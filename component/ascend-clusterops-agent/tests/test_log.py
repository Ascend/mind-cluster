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

"""agent-core / node-collector shared log-to-disk tests (clusterops_common.log: RotatingFileHandler + stdout)."""

from __future__ import annotations

import logging

import pytest

from clusterops_common import log as common_log


@pytest.fixture
def _reset_logging():
    """Reset the handlers attached by init_logging before and after each test to avoid interference."""
    common_log.reset_logging()
    yield
    common_log.reset_logging()


def test_agent_core_log_writes_to_file(tmp_path, monkeypatch, _reset_logging):
    monkeypatch.setenv("LOG_DIR", str(tmp_path))
    common_log.init_logging(log_dir="/var/log/mindx-dl/agent-core", log_file="agent-core.log")
    logging.getLogger("agent_core.tools").info("agent-core-test-marker")
    log_file = tmp_path / "agent-core.log"
    assert log_file.exists()
    assert "agent-core-test-marker" in log_file.read_text(encoding="utf-8")


def test_node_collector_log_writes_to_file(tmp_path, monkeypatch, _reset_logging):
    monkeypatch.setenv("LOG_DIR", str(tmp_path))
    common_log.init_logging(log_dir="/var/log/mindx-dl/node-collector", log_file="node-collector.log")
    logging.getLogger("node_collector.collector").info("node-collector-test-marker")
    log_file = tmp_path / "node-collector.log"
    assert log_file.exists()
    assert "node-collector-test-marker" in log_file.read_text(encoding="utf-8")


def test_init_logging_idempotent(tmp_path, monkeypatch, _reset_logging):
    monkeypatch.setenv("LOG_DIR", str(tmp_path))
    common_log.reset_logging()
    common_log.init_logging(log_dir=str(tmp_path), log_file="agent-core.log")
    handlers = list(common_log._APP_HANDLERS)
    common_log.init_logging(
        log_dir=str(tmp_path), log_file="agent-core.log"
    )  # a second call must not attach duplicate handlers
    assert list(common_log._APP_HANDLERS) == handlers


def test_init_logging_falls_back_when_dir_unwritable(monkeypatch, tmp_path, _reset_logging):
    # the parent dir is missing and its parent is a file, so mkdir must fail -> fall back to stdout without raising
    blocker = tmp_path / "blocker"
    blocker.write_text("x")
    monkeypatch.setenv("LOG_DIR", str(blocker / "sub" / "log"))
    common_log.reset_logging()
    common_log.init_logging(log_dir=str(blocker / "sub" / "log"), log_file="agent-core.log")
    assert not (blocker / "sub" / "log" / "agent-core.log").exists()
