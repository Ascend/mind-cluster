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

"""Shared agent-core constants (single source of truth, tunable via environment)."""

import os
from pathlib import Path

# tunable via the POD_TTL env.
POD_TTL = int(os.environ.get("POD_TTL", str(7 * 86400)))

# Root directory for diagnosis artifacts / caches (host-exported hostPath, not /tmp);
# AGENT_WORK_ROOT overrides. All modules read it from here to avoid duplicated defaults.
WORK_ROOT = Path(os.environ.get("AGENT_WORK_ROOT") or "/user/clusterops/agent-core")

# Namespace that hosts all agent-core ConfigMaps (task-crds config, relcache/pathmap snapshots).
CLUSTER_SYSTEM_NS = "cluster-system"

# Task CR GVK config CM: declares which CR kinds own the diagnosis-target pods.
# Read by k8s.load_task_crds_config (relcache/pathmap pod filtering).
TASK_CRDS_CM_NAME = "agent-core-task-crds"
TASK_CRDS_CM_KEY = "task_crds.yaml"

# agent-core snapshot CMs share one sharding layout: each shard stays below the
# 1MiB ConfigMap data limit, so the total capacity scales linearly with SNAPSHOT_SHARDS
# (100 shards = 100MiB); shards are filled sequentially and the in-memory byte cap equals
# the snapshot capacity, so the memory table never outgrows the snapshot.
RELCACHE_CM_NAME = "agent-core-relcache"
PATHMAP_CM_NAME = "clusterops-pathmap"
SNAPSHOT_SHARDS = 100
SNAPSHOT_MAX_BYTES = SNAPSHOT_SHARDS * 1024 * 1024
SNAPSHOT_SHARD_FILL_LIMIT = 1024 * 1024 - 16 * 1024
