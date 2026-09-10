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

"""quota module tests: prune the oldest job-level entries until total size <= limit."""

from __future__ import annotations

import os
from pathlib import Path

from clusterops_common import quota

OLD = 1_000_000


def _mk_job(root: Path, day: str, ns_job: str, content: bytes, mtime: float = OLD) -> Path:
    """Create {root}/{day}/{ns_job}/data.bin with a fixed mtime and return the job dir."""
    job = root / day / ns_job
    job.mkdir(parents=True, exist_ok=True)
    (job / "data.bin").write_bytes(content)
    os.utime(job, (mtime, mtime))
    return job


def test_no_op_when_under_limit(tmp_path):
    _mk_job(tmp_path, "20260901", "ns_a", b"x" * 100, OLD)
    _mk_job(tmp_path, "20260902", "ns_b", b"y" * 200, OLD + 1000)
    removed = quota.enforce_quota(tmp_path, max_bytes=10_000)
    assert removed == 0
    assert (tmp_path / "20260901" / "ns_a").exists()
    assert (tmp_path / "20260902" / "ns_b").exists()


def test_prune_oldest_job_dir_when_over_limit(tmp_path):
    _mk_job(tmp_path, "20260901", "ns_a", b"x" * 100, OLD)  # older
    _mk_job(tmp_path, "20260902", "ns_b", b"y" * 100, OLD + 1000)  # newer
    # limit 150: dropping only the oldest (100 bytes) brings total to 100 <= 150
    removed = quota.enforce_quota(tmp_path, max_bytes=150)
    assert removed == 1
    assert not (tmp_path / "20260901" / "ns_a").exists()
    assert (tmp_path / "20260902" / "ns_b").exists()


def test_mtime_decides_oldest_not_lexicographic(tmp_path):
    # older mtime has a larger lex name; lexicographic sort alone would prune the wrong one
    a = _mk_job(tmp_path, "20260901", "zz_older", b"a" * 200, OLD)
    b = _mk_job(tmp_path, "20260901", "aa_newer", b"b" * 100, OLD + 1000)
    quota.enforce_quota(tmp_path, max_bytes=150)
    assert not a.exists()
    assert b.exists()


def test_prune_cache_files_oldest_first(tmp_path):
    cache = tmp_path / "cache"
    cache.mkdir()
    c1 = cache / "ns_a.json"
    c1.write_bytes(b"1" * 100)
    os.utime(c1, (OLD, OLD))
    c2 = cache / "ns_b.json"
    c2.write_bytes(b"2" * 100)
    os.utime(c2, (OLD + 1000, OLD + 1000))
    removed = quota.enforce_quota(tmp_path, max_bytes=150, cache_dir=cache)
    assert removed == 1
    assert not c1.exists()
    assert c2.exists()


def test_prune_job_dirs_and_cache_together(tmp_path):
    _mk_job(tmp_path, "20260901", "ns_a", b"j" * 300, OLD)  # oldest
    cache = tmp_path / "cache"
    cache.mkdir()
    c = cache / "ns_c.json"
    c.write_bytes(b"k" * 100)
    os.utime(c, (OLD + 1000, OLD + 1000))  # newest
    # total 400 bytes, limit 150 -> drop the oldest job dir (300 bytes) first
    removed = quota.enforce_quota(tmp_path, max_bytes=150, cache_dir=cache)
    assert not (tmp_path / "20260901" / "ns_a").exists()
    assert c.exists()
    assert removed == 1
