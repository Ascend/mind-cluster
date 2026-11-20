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

"""Diagnosis CLI contract: removed alias subcommands stay rejected; the flat fallback printer still works.

Cases that drive the diagnosis chain end to end (ascend_diag legacy output, tool diagnose envelope,
collect-manifest passthrough) import agent_core.main, which pulls the ReAct chat stack and therefore needs
langchain. They are temporarily parked in branch v5-ReAct-WebUI-allTests until the UT image ships
Python 3.10 + langchain.
"""

# pylint: disable=redefined-outer-name  # pytest fixtures are injected by parameter name

from __future__ import annotations

import importlib.util
import json
import sys
from importlib.machinery import SourceFileLoader
from pathlib import Path

import pytest

PLUGIN_DIR = Path(__file__).resolve().parents[1] / "kubectl-plugin"
sys.path.insert(0, str(PLUGIN_DIR))


def _load(filename):
    loader = SourceFileLoader(filename.replace("-", "_"), str(PLUGIN_DIR / filename))
    module = importlib.util.module_from_spec(importlib.util.spec_from_loader(loader.name, loader))
    loader.exec_module(module)
    return module


clusterops = _load("kubectl-clusterops")


@pytest.mark.parametrize("report_text", [None, "raw report body"])
def test_print_diagnosis_result_flat_fallback(capsys, report_text):
    """The fallback printer still shows deterministic text or the original report JSON."""
    report = {"Root_Cluster": {"fault_description": {"code": 109, "string": "AICPU timeout"}}}
    clusterops._print_diagnosis_result(
        {"status": "ok", "data": {"diag_report": report, "diag_report_text": report_text}}
    )
    stdout = capsys.readouterr().out
    if report_text:
        assert report_text in stdout
    else:
        assert json.loads(stdout) == report
    assert "(no output)" not in stdout


@pytest.mark.parametrize(
    "argv",
    [
        ["diagnose", "--job", "train-a"],
        ["--collect-manifest", "collect.yaml"],
        ["faulty-device"],
        ["isolated-card"],
        ["cluster-resource"],
        ["task-status", "--job", "train-a"],
    ],
)
def test_removed_alias_subcommands_are_rejected(argv, monkeypatch):
    monkeypatch.setattr(sys, "argv", ["kubectl-clusterops"] + argv)
    with pytest.raises(SystemExit) as exc:
        clusterops.main()
    assert exc.value.code == 2
