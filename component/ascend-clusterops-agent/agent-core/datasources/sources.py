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

"""Where each query reads its data: a fixed table of namespaces, ConfigMap names and labels.

The RFC defines no configuration surface, so these are source constants rather than a
packaged/overridable profile: changing a namespace or a ConfigMap name is a code change
and a release, exactly like any other production behaviour change.

Every value is the producer contract of the ClusterD / DP versions supported today; a
missing producer side is expressed by ``absent_expected`` (read is normal, "no record").
"""

from __future__ import annotations

from typing import NamedTuple, Optional


class CmSource(NamedTuple):
    """One ConfigMap addressed by namespace + name."""

    namespace: str
    name: str
    absent_expected: bool = False


class CmPrefixSource(NamedTuple):
    """ConfigMap shards addressed by namespace + name prefix (+ optional label selector)."""

    namespace: str
    name_prefix: str
    label_selector: Optional[str] = None


# ClusterD device summary shards (cluster-info-device-0 ...); the label is a deployment
# prerequisite, so no unlabeled full-namespace fallback is attempted.
CLUSTER_DEVICES = CmPrefixSource("mindx-dl", "cluster-info-device", "mx-consumer-volcano=true")
# Per-node DP ConfigMap (mindx-dl-deviceinfo-<node>): faulty_device fallback source.
DP_DEVICES = CmPrefixSource("kube-system", "mindx-dl-deviceinfo-")
# ClusterD node summary: node health and node-level faults.
NODE_CM = CmSource("mindx-dl", "cluster-info-node-cm")
# ClusterD switch / DPU fault shards.
SWITCH = CmPrefixSource("mindx-dl", "cluster-info-switch")
DPU = CmPrefixSource("mindx-dl", "cluster-info-dpu")
# Producer-wide public fault records.
PUBLIC_FAULTS = CmSource("cluster-system", "statistic-fault-info", True)
# Manual/silent isolation records (ClusterD).
MANUAL_ISOLATION = CmSource("cluster-system", "clusterd-manual-info-cm", True)
