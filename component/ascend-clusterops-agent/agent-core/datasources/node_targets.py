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

"""Explicit node target existence validation for the node-scoped query tools.

A requested node that is not a Node object of this cluster is a request problem, not a
coverage gap:

* every requested node missing -> the query has no target at all, so it fails with
  ``TARGET_NOT_FOUND`` / HTTP 404 and the message ``所有节点都找不到``;
* only some requested nodes missing -> the surviving nodes are still answered, the result
  is marked ``partial`` and one ``TARGET_NOT_FOUND`` issue per missing node carries
  ``节点 <name> 不存在``.

Existence means "visible as a Node to agent-core". If the Node list itself cannot be read
completely, existence cannot be asserted, so the check is skipped and the normal
source-coverage behaviour applies.
"""

from __future__ import annotations

import logging
from typing import List

from agent_core.datasources.query_context import QueryContext, get_gateway
from agent_core.query_schemas import QueryFailure, QueryIssue

logger = logging.getLogger(__name__)

ALL_NODES_MISSING = "所有节点都找不到"


def _missing_issue(name: str) -> QueryIssue:
    return QueryIssue(code="TARGET_NOT_FOUND", path="node", message=f"节点 {name} 不存在")


def validate_target_nodes(ctx: QueryContext) -> List[QueryIssue]:
    """Validate explicitly requested nodes and return one issue per missing node.

    Raises ``QueryFailure`` when every requested node is missing. When only some are
    missing, the missing nodes are removed from ``ctx.requested_nodes`` so no source is
    read for them, and the caller must publish ``partial`` together with these issues.
    ``node=None`` (full cluster) and ``node=[]`` (explicit empty scope) are not validated.
    """
    requested = ctx.requested_nodes
    if not requested:
        return []
    result = get_gateway(ctx).list_nodes(None)
    if not result.ok or not result.listing_complete:
        logger.warning(
            "node target validation skipped: node listing status=%s complete=%s",
            result.status,
            result.listing_complete,
        )
        return []
    present = {(item.get("metadata") or {}).get("name") or "" for item in result.items}
    missing = [name for name in requested if name not in present]
    if not missing:
        return []
    if len(missing) == len(requested):
        raise QueryFailure("TARGET_NOT_FOUND", ALL_NODES_MISSING, http_status=404, path="node")
    ctx.requested_nodes = [name for name in requested if name not in missing]
    return [_missing_issue(name) for name in missing]
