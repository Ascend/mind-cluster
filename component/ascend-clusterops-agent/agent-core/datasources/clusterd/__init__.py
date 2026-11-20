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

"""Parsers for the ClusterD summary ConfigMaps, one module per producer document.

``device_summary``  cluster-info-device-{m}   per-node DeviceList / topology
``node_summary``    cluster-info-node-cm      node health and node-level faults
``switch_summary``  cluster-info-switch-{x}   Lingqu switch faults
``dpu_summary``     cluster-info-dpu-{n}      DPU faults (device and node scope)

Every module maps one producer document into model objects; the shared CM envelope
(one JSON shard per data key) is decoded by ``agent_core.datasources.cm_documents``.
"""
