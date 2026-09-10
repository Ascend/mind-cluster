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

"""Shared disk-quota enforcement for agent-core / node-collector work roots.

Both components write artifacts under the per-day, per-job layout
    {work_root}/{YYYYMMDD}/{ns}_{job}/...
and the agent additionally keeps per-job cache files
    {work_root}/cache/{ns}_{job}.json

Diagnosis output would otherwise grow without bound. This module prunes the oldest
job-level entries until the total size stays within a configurable limit (default 10GB).
"""

from __future__ import annotations

import logging
import os
import shutil
import threading
from pathlib import Path

logger = logging.getLogger(__name__)

DEFAULT_MAX_BYTES = 10 * 1024**3  # 10GB
MAX_BYTES = int(os.environ.get("WORK_ROOT_MAX_BYTES", str(DEFAULT_MAX_BYTES)))

_prune_lock = threading.Lock()


def _dir_size(path: Path) -> int:
    """Total bytes of a directory tree (0 when missing/empty)."""
    total = 0
    for p in path.rglob("*"):
        try:
            if p.is_file():
                total += p.stat().st_size
        except OSError:
            continue
    return total


def _mtime(path: Path) -> float:
    try:
        return path.stat().st_mtime
    except OSError:
        return 0.0


def _job_dirs(root: Path) -> list[Path]:
    """job-level dirs {root}/{YYYYMMDD}/{ns}_{job} (cache files are skipped, not dirs)."""
    jobs: list[Path] = []
    if not root.is_dir():
        return jobs
    for day in root.iterdir():
        if not day.is_dir():
            continue
        for job in day.iterdir():
            if job.is_dir():
                jobs.append(job)
    return jobs


def _cache_files(cache_dir: Path) -> list[Path]:
    if not cache_dir.is_dir():
        return []
    return [p for p in cache_dir.iterdir() if p.is_file()]


def enforce_quota(root: Path, max_bytes: int = MAX_BYTES, cache_dir: Path | None = None) -> int:
    """Delete oldest job-level entries until total size <= max_bytes; returns the number deleted.

    Units are job dirs under {root}/{YYYYMMDD}/{ns}_{job} plus cache files under
    cache_dir (default {root}/cache when present). Best-effort: never raises.
    """
    if cache_dir is None:
        cache_dir = root / "cache"
    with _prune_lock:
        return _enforce_quota_locked(root, max_bytes, cache_dir)


def _enforce_quota_locked(root: Path, max_bytes: int, cache_dir: Path) -> int:
    units: list[tuple[Path, float, int]] = []
    for jd in _job_dirs(root):
        units.append((jd, _mtime(jd), _dir_size(jd)))
    for cf in _cache_files(cache_dir):
        try:
            size = cf.stat().st_size
        except OSError:
            size = 0
        units.append((cf, _mtime(cf), size))

    total = sum(u[2] for u in units)
    if total <= max_bytes:
        return 0

    units.sort(key=lambda u: u[1])  # oldest first
    removed = 0
    for path, _mt, size in units:
        if total <= max_bytes:
            break
        try:
            if path.is_dir():
                shutil.rmtree(path, ignore_errors=True)
            else:
                path.unlink(missing_ok=True)
        except OSError as e:  # noqa: BLE001  # best-effort prune, skip the stuck entry
            logger.warning("quota prune failed for %s: %s", path, e)
            continue
        total -= size
        removed += 1
        logger.info("quota pruned oldest entry: %s (reclaimed %d bytes)", path, size)
    return removed
