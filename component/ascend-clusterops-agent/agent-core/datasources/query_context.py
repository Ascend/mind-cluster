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

"""Per-request query context: declared scope and in-request read reuse.

Nothing here is shared across requests or written to the cluster. The limits below are
not user-tunable: they only keep a single request's loop bounded (one call timeout, one
list page, one target list), so they are constants.
"""

from __future__ import annotations

import uuid
from typing import Any, Callable, Dict, List, Optional

from agent_core.query_schemas import NodeQuery, QueryFailure

# Fixed, non-configurable safety limits (implementation detail, not part of the contract).
CALL_TIMEOUT_SECONDS = 5.0
PAGE_SIZE = 500
MAX_TARGETS = 20000


class QueryContext:
    """Request-scoped state handed to every tool handler."""

    def __init__(self, request: Any = None, request_id: Optional[str] = None):
        self.request = request
        self.request_id = request_id or uuid.uuid4().hex
        self.requested_nodes: Optional[List[str]] = None
        self.gateway = None  # injected in tests; otherwise built lazily from K8s.core()
        self._reads: Dict[str, Any] = {}
        self._source_seq = 0

    # ---- evidence pointer ids ---- #
    def next_source_id(self) -> str:
        self._source_seq += 1
        return f"s{self._source_seq}"

    # ---- scope ---- #
    def check_scope(self, request: Any) -> None:
        if isinstance(request, NodeQuery) and request.node is not None:
            if len(request.node) > MAX_TARGETS:
                raise QueryFailure(
                    "INVALID_ARGUMENT", f"node exceeds the limit of {MAX_TARGETS}", http_status=422, path="node"
                )
            # de-duplicate while preserving order; the caller's model is not mutated
            self.requested_nodes = list(dict.fromkeys(request.node))

    # ---- in-request reuse ---- #
    def read_once(self, key: str, loader: Callable[[], Any]) -> Any:
        """Read one source at most once per request and reuse the result."""
        if key in self._reads:
            return self._reads[key]
        value = loader()
        self._reads[key] = value
        return value


def get_gateway(ctx: QueryContext):
    """Lazily build the read-only gateway from the existing memoized K8s client."""
    if ctx.gateway is None:
        from agent_core.datasources.gateway import KubernetesReadGateway
        from agent_core.k8s import K8s

        ctx.gateway = KubernetesReadGateway(K8s.core(), ctx)
    return ctx.gateway
