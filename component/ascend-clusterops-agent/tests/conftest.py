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

"""pytest setup: lets tests/ import the agent_core / node_collector / clusterops_common / diagproto packages.

The source directory names contain hyphens (agent-core/ node-collector/) and cannot be imported
directly as Python packages; production uses the pyproject `package-dir` mapping (agent_core = ".").
The test environment does not install the wheels, so the same mapping is mounted as aliases in the
import system here; nothing else changes.
"""

import importlib.util
import os
import sys
from types import SimpleNamespace

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def _mount_package(src_dir: str, pkg_name: str) -> None:
    """Mount <ROOT>/<src_dir> as <pkg_name> (equivalent to the pyproject package-dir mapping)."""
    if pkg_name in sys.modules:
        return
    path = os.path.join(ROOT, src_dir)
    spec = importlib.util.spec_from_file_location(
        pkg_name, os.path.join(path, "__init__.py"), submodule_search_locations=[path]
    )
    mod = importlib.util.module_from_spec(spec)
    sys.modules[pkg_name] = mod
    if spec.loader is not None:
        spec.loader.exec_module(mod)


def make_pod_metadata(uid, name, ns, owner_references):
    """Build the pod metadata SimpleNamespace shared by the test pod builders."""
    return SimpleNamespace(
        uid=uid,
        name=name,
        namespace=ns,
        annotations={},
        owner_references=owner_references,
    )


_mount_package("common", "clusterops_common")
_mount_package("common/diagproto", "diagproto")
_mount_package("agent-core", "agent_core")
_mount_package("node-collector", "node_collector")
