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

"""Per-node DP ConfigMap reads used by the faulty-device query.

The producer publishes one ConfigMap per node (``mindx-dl-deviceinfo-<node>``). A scoped
read fetches exactly the requested nodes; a full-cluster read lists the prefix once
instead of issuing one GET per node. The read plan lives here so the fallback selection
rules stay in one place.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Dict, Iterable, List, Optional, Sequence, Tuple

from agent_core.datasources.query_context import QueryContext, get_gateway
from agent_core.datasources.gateway import SourceReadResult


@dataclass
class DpDocument:
    """One DP ConfigMap that was read successfully: node, name, evidence id and data."""

    node: str
    name: str
    source_id: str
    data: Dict[str, Any]


def read_dp_configmaps(
    ctx: QueryContext,
    namespace: str,
    prefix: str,
    *,
    nodes: Optional[Sequence[str]] = None,
    skip: Iterable[str] = (),
) -> Tuple[Optional[SourceReadResult], List[DpDocument], List[SourceReadResult]]:
    """Read the DP ConfigMaps in scope.

    ``nodes`` is the declared target scope (None means the whole cluster). A node in
    ``skip`` is never read: its primary source already answered for it.

    Returns ``(listing, documents, failed)``:

    * ``listing``   - the ``list_cms`` result in whole-cluster mode, otherwise None;
    * ``documents`` - every readable ConfigMap, in scope order;
    * ``failed``    - per-target reads that did not succeed (always empty in list mode,
      where a failed listing is reported once through ``listing``).
    """
    skip_set = set(skip)
    gateway = get_gateway(ctx)
    if nodes is None:
        listing = gateway.list_cms(namespace, None, prefix)
        documents: List[DpDocument] = []
        for item in listing.items:
            name = item.get("name") or ""
            if not name.startswith(prefix):
                continue
            node = name[len(prefix) :]
            if node in skip_set:
                continue
            documents.append(DpDocument(node=node, name=name, source_id=listing.source_id, data=item.get("data") or {}))
        return listing, documents, []

    documents = []
    failed: List[SourceReadResult] = []
    for node in nodes:
        if node in skip_set:
            continue
        name = f"{prefix}{node}"
        read = gateway.get_cm(namespace, name)
        if not read.ok:
            # An empty ConfigMap is readable and simply contributes no record.
            failed.append(read)
            continue
        documents.append(
            DpDocument(node=node, name=name, source_id=read.source_id, data=(read.items or [{}])[0].get("data") or {})
        )
    return None, documents, failed
