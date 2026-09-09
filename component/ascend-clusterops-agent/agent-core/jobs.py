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

"""Agent-side collection job aggregation (JobTracker): tracks expected nodes, collects
reports, and wakes up when all nodes report or the timeout expires.

In the asynchronous report model, the agent is both the sender (TriggerCollect) and
the receiver (UploadResult):
  - diagnose triggers all nodes concurrently, then blocks on JobTracker.wait;
  - the UploadResult handler calls complete once per received node result;
  - when all nodes report or the timeout expires, the waiter wakes up and runs the
    centralized diagnosis.
"""

from __future__ import annotations

import threading
import time

from typing import Any


class JobTracker:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._cv = threading.Condition(self._lock)
        self._jobs: dict[str, dict[str, Any]] = {}

    def register(self, job: str, nodes: list[str]) -> None:
        """Register the expected node set for a job (overwrites an existing job with the same name)."""
        with self._cv:
            self._jobs[job] = {"expected": {n for n in nodes if n}, "results": {}}
            self._cv.notify_all()

    def complete(self, job: str, node: str, result: dict) -> None:
        """Record the final result of one node (trigger failure / report success / report failure all go through here)."""
        with self._cv:
            j = self._jobs.get(job)
            if j is None:
                return
            j["results"][node] = result
            self._cv.notify_all()

    def get(self, job: str) -> dict | None:
        """Return an {expected, results} snapshot; None when the job does not exist."""
        with self._cv:
            j = self._jobs.get(job)
            if j is None:
                return None
            return {"expected": set(j["expected"]), "results": dict(j["results"])}

    def wait(self, job: str, timeout: float) -> dict[str, dict]:
        """Wait until all expected nodes report or the timeout expires; return node->result. Missing nodes are patched by the caller."""
        with self._cv:
            j = self._jobs.get(job)
            if j is None:
                return {}
            end = time.monotonic() + timeout
            while len(j["results"]) < len(j["expected"]) and time.monotonic() < end:
                self._cv.wait(max(0.0, end - time.monotonic()))
            return dict(j["results"])


tracker = JobTracker()
