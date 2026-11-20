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

"""Read-only Kubernetes gateway: get/list through the existing client, with a classified
source status, fixed paging and a per-object evidence id. No write verbs exist.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Callable, Dict, List, Optional

from agent_core.datasources.query_context import CALL_TIMEOUT_SECONDS, PAGE_SIZE, QueryContext
from kubernetes import client

# Connection-level failures that mean "upstream unreachable" rather than "bad data".
_TRANSIENT_EXCEPTION_NAMES = frozenset(
    {
        "ProtocolError",
        "MaxRetryError",
        "NewConnectionError",
        "RemoteDisconnected",
        "ConnectionResetError",
        "ConnectionRefusedError",
        "BrokenPipeError",
    }
)


@dataclass
class SourceReadResult:
    status: str  # ok/empty/absent_expected/missing/forbidden/invalid/unavailable
    source_id: str = ""
    items: List[Dict[str, Any]] = field(default_factory=list)
    missing: List[str] = field(default_factory=list)
    resource_version: Optional[str] = None
    error_code: Optional[str] = None
    message: Optional[str] = None
    listing_complete: bool = True
    observed: int = 0

    @property
    def ok(self) -> bool:
        return self.status in ("ok", "empty")


def _to_dict(obj: Any) -> Dict[str, Any]:
    if isinstance(obj, dict):
        return obj
    try:
        return client.ApiClient().sanitize_for_serialization(obj)
    except Exception:  # noqa: BLE001  # simple namespace / test doubles
        value = getattr(obj, "__dict__", None)
        if isinstance(value, dict):
            return value
        return {"value": str(obj)}


def _classify_exception(exc: BaseException, absent_expected: bool = False):
    """Map one API failure to (source status, error code).

    Timeouts and transport failures share ``SOURCE_UNAVAILABLE``: there is no separate
    "deadline" or "limit" outcome any more.
    """
    status = getattr(exc, "status", None)
    if status == 404:
        return ("absent_expected" if absent_expected else "missing", "SOURCE_NOT_FOUND")
    if status == 403:
        return "forbidden", "SOURCE_FORBIDDEN"
    if status in (408, 429, 504) or isinstance(exc, TimeoutError) or "Timeout" in type(exc).__name__:
        return "unavailable", "SOURCE_UNAVAILABLE"
    if isinstance(status, int) and status >= 500:
        return "unavailable", "SOURCE_UNAVAILABLE"
    if any(cls.__name__ in _TRANSIENT_EXCEPTION_NAMES for cls in type(exc).__mro__):
        return "unavailable", "SOURCE_UNAVAILABLE"
    return "invalid", "SOURCE_INVALID"


@dataclass
class _Outcome:
    """Result of one API call: either a response or a classified failure."""

    response: Any = None
    status: str = "ok"
    error_code: Optional[str] = None
    message: Optional[str] = None

    @property
    def ok(self) -> bool:
        return self.response is not None


@dataclass
class _Page:
    """Everything collected by one paginated list read, plus the first failed page."""

    items: List[Any]
    resource_version: Optional[str]
    complete: bool
    failure: Optional[_Outcome]


def _continue_token(metadata: Any) -> Optional[str]:
    """Next-page token of one list response (the client exposes ``_continue`` or ``continue``)."""
    return getattr(metadata, "_continue", None) or getattr(metadata, "continue", None)


class KubernetesReadGateway:
    """Thin read-only wrapper over CoreV1Api; callers must not assume cross-object atomicity."""

    def __init__(self, core_client, ctx: QueryContext, custom_client: Any = None):
        self._v1 = core_client
        self._custom_client = custom_client
        self._ctx = ctx

    # ---- evidence ids ---- #
    def _record(self) -> str:
        """Allocate the evidence id for one source read (the request keeps the order)."""
        return self._ctx.next_source_id()

    def _timeout(self) -> float:
        return CALL_TIMEOUT_SECONDS

    def _invoke_read(self, call: Callable[[], Any], absent_expected: bool = False) -> _Outcome:
        """Call one read-only API operation exactly once and classify a failure."""
        try:
            return _Outcome(response=call())
        except Exception as exc:  # noqa: BLE001  # classified into the source contract
            status, code = _classify_exception(exc, absent_expected)
            return _Outcome(status=status, error_code=code, message=str(exc))

    # ---- paginated list ---- #
    def _paged(self, call: Callable[[Optional[str]], Any]) -> "_Page":
        """Run one paginated list call to completion and classify the first failed page."""
        items: List[Any] = []
        resource_version: Optional[str] = None
        token: Optional[str] = None
        while True:
            outcome = self._invoke_read(lambda: call(token))
            if not outcome.ok:
                return _Page(items, resource_version, False, outcome)
            metadata = getattr(outcome.response, "metadata", None)
            resource_version = resource_version or getattr(metadata, "resource_version", None)
            items.extend(getattr(outcome.response, "items", []) or [])
            token = _continue_token(metadata)
            if not token:
                return _Page(items, resource_version, True, None)

    def _source_result_from_page(self, page: "_Page", items: List[Dict[str, Any]], **extra: Any) -> SourceReadResult:
        """Build the source result of one list read from its page outcome."""
        failure = page.failure
        if failure is not None:
            return SourceReadResult(
                status=failure.status,
                source_id=self._record(),
                items=items,
                resource_version=page.resource_version,
                error_code=failure.error_code,
                message=failure.message,
                listing_complete=False,
                observed=len(items),
                **extra,
            )
        return SourceReadResult(
            status="ok" if items else "empty",
            source_id=self._record(),
            items=items,
            resource_version=page.resource_version,
            observed=len(items),
            **extra,
        )

    # ---- ConfigMaps ---- #
    def get_cm(
        self,
        namespace: str,
        name: str,
        absent_expected: bool = False,
        *,
        use_cache: bool = True,
    ) -> SourceReadResult:
        """Read a CM; uncached reads neither consult nor replace shared request slots."""
        key = f"cm:{namespace}:{name}"

        def loader() -> SourceReadResult:
            outcome = self._invoke_read(
                lambda: self._v1.read_namespaced_config_map(name, namespace, _request_timeout=self._timeout()),
                absent_expected,
            )
            source_id = self._record()
            if not outcome.ok:
                return SourceReadResult(
                    status=outcome.status,
                    source_id=source_id,
                    error_code=outcome.error_code,
                    message=outcome.message,
                )
            obj = _to_dict(outcome.response)
            metadata = obj.get("metadata") or {}
            data = obj.get("data") or {}
            status = "ok" if data else "empty"
            return SourceReadResult(
                status=status,
                source_id=source_id,
                items=[{"name": name, "data": data, "metadata": metadata}],
                resource_version=metadata.get("resourceVersion"),
                observed=len(data),
            )

        return self._ctx.read_once(key, loader) if use_cache else loader()

    def list_cms(
        self,
        namespace: str,
        label_selector: Optional[str] = None,
        name_prefix: Optional[str] = None,
    ) -> SourceReadResult:
        key = f"cms:{namespace}:{label_selector or '*'}:{name_prefix or '*'}"
        return self._ctx.read_once(key, lambda: self._list_cms_once(namespace, label_selector, name_prefix))

    def _list_cms_once(self, namespace: str, selector: Optional[str], name_prefix: Optional[str]) -> SourceReadResult:
        page = self._paged(
            lambda token: self._v1.list_namespaced_config_map(
                namespace,
                label_selector=selector,
                limit=PAGE_SIZE,
                _continue=token,
                _request_timeout=self._timeout(),
            )
        )
        items: List[Dict[str, Any]] = []
        for cm in page.items:
            obj = _to_dict(cm)
            name = (obj.get("metadata") or {}).get("name") or ""
            if name_prefix and not name.startswith(name_prefix):
                continue
            items.append({"name": name, "data": obj.get("data") or {}, "metadata": obj.get("metadata") or {}})
        return self._source_result_from_page(page, items)

    # ---- Nodes ---- #
    def list_nodes(self, names: Optional[List[str]] = None) -> SourceReadResult:
        if names is not None:
            key = "nodes:" + ",".join(names)
            return self._ctx.read_once(key, lambda: self._get_nodes(names))
        key = "nodes:all"
        return self._ctx.read_once(key, self._list_nodes_once)

    def _get_nodes(self, names: List[str]) -> SourceReadResult:
        items: List[Dict[str, Any]] = []
        missing: List[str] = []
        status = "ok"
        error_code: Optional[str] = None
        message: Optional[str] = None
        complete = True
        for name in names:
            outcome = self._invoke_read(
                lambda node_name=name: self._v1.read_node(node_name, _request_timeout=self._timeout())
            )
            if not outcome.ok:
                if outcome.status == "missing":
                    missing.append(name)
                    continue
                status, error_code, message = outcome.status, outcome.error_code, outcome.message
                complete = False
                break
            items.append(_to_dict(outcome.response))
        return SourceReadResult(
            status=status,
            source_id=self._record(),
            items=items,
            missing=missing,
            error_code=error_code,
            message=message,
            listing_complete=complete,
            observed=len(items),
        )

    def _list_nodes_once(self) -> SourceReadResult:
        page = self._paged(
            lambda token: self._v1.list_node(limit=PAGE_SIZE, _continue=token, _request_timeout=self._timeout())
        )
        return self._source_result_from_page(page, [_to_dict(node) for node in page.items])

    # ---- Pods ---- #
    def list_pods(self, namespace: Optional[str] = None) -> SourceReadResult:
        key = f"pods:{namespace or '*'}"
        return self._ctx.read_once(key, lambda: self._list_pods_once(namespace))

    def _list_pods_once(self, namespace: Optional[str]) -> SourceReadResult:
        def page_call(token: Optional[str]):
            if namespace:
                return self._v1.list_namespaced_pod(
                    namespace, limit=PAGE_SIZE, _continue=token, _request_timeout=self._timeout()
                )
            return self._v1.list_pod_for_all_namespaces(
                limit=PAGE_SIZE, _continue=token, _request_timeout=self._timeout()
            )

        page = self._paged(page_call)
        return self._source_result_from_page(page, [_to_dict(pod) for pod in page.items])

    # ---- custom objects (task CRDs) ---- #
    def _custom(self):
        if self._custom_client is None:
            from agent_core.k8s import K8s

            self._custom_client = K8s.custom()
        return self._custom_client

    def get_custom_object(self, group: str, version: str, plural: str, namespace: str, name: str) -> SourceReadResult:
        """Read one namespaced custom object by exact name; a missing CRD/object is ``missing``."""
        key = f"custom:{group}/{version}:{namespace}:{plural}/{name}"
        return self._ctx.read_once(key, lambda: self._get_custom_object_once(group, version, plural, namespace, name))

    def _get_custom_object_once(
        self, group: str, version: str, plural: str, namespace: str, name: str
    ) -> SourceReadResult:
        outcome = self._invoke_read(
            lambda: self._custom().get_namespaced_custom_object(
                group, version, namespace, plural, name, _request_timeout=self._timeout()
            )
        )
        source_id = self._record()
        if not outcome.ok:
            return SourceReadResult(
                status=outcome.status,
                source_id=source_id,
                error_code=outcome.error_code,
                message=outcome.message,
            )
        obj = _to_dict(outcome.response)
        metadata = obj.get("metadata") or {}
        return SourceReadResult(
            status="ok",
            source_id=source_id,
            items=[obj],
            resource_version=metadata.get("resourceVersion"),
            observed=1,
        )
