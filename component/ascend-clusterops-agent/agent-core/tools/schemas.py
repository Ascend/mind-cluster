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

"""Shared tool contract: catalog/descriptor discovery plus diagnosis result models.

All tools expose one typed envelope. Diagnosis data preserves the flat /diag
field names and nested payloads used by kubectl ascend_diag. Optional legacy
fields are emitted only when present; no renamed or synthetic fields are added.
"""

from __future__ import annotations

from typing import Any, Dict, List, Optional

from agent_core.query_schemas import KubernetesName, NamespaceName, ToolEnvelope
from pydantic import BaseModel, ConfigDict, Field, model_serializer


# --------------------------------------------------------------------------- #
# Diagnosis input / output
# --------------------------------------------------------------------------- #
class DiagnosisQuery(BaseModel):
    """Strict input of the unified diagnose tool (the legacy /diag request stays separate)."""

    model_config = ConfigDict(extra="forbid", strict=True)

    job: KubernetesName
    namespace: NamespaceName = "default"
    refresh: bool = False


class DiagnosisData(BaseModel):
    """The unchanged /diag business payload, including producer extension fields."""

    model_config = ConfigDict(extra="allow", strict=True)

    error: Optional[str] = None
    pods: Optional[List[Dict[str, Any]]] = None
    collected: Optional[List[Dict[str, Any]]] = None
    diag_report: Optional[Dict[str, Any]] = None
    diag_report_text: Optional[str] = None
    diag_input_dir: Optional[str] = None
    diag_output_dir: Optional[str] = None
    incomplete_nodes: Optional[List[str]] = None
    empty_nodes: Optional[List[str]] = None
    cached: Optional[bool] = None
    cached_note: Optional[str] = None
    final_text: Optional[str] = None
    llm_error: Optional[str] = None

    @model_serializer(mode="wrap")
    def serialize_present_fields(self, handler: Any) -> Dict[str, Any]:
        """Preserve absent vs null through Registry validation and nested serialization."""
        payload = handler(self)
        return {key: value for key, value in payload.items() if key in self.model_fields_set}


class DiagnosisResult(ToolEnvelope):
    """Public execution envelope of the diagnose tool; data matches the flat /diag body."""

    data: Optional[DiagnosisData] = None


def tool_response_payload(result: BaseModel) -> dict[str, Any]:
    """Serialize the public tool contract (one envelope for every registered tool)."""
    return result.model_dump(mode="json")


# --------------------------------------------------------------------------- #
# Discovery (catalog / descriptor)
# --------------------------------------------------------------------------- #
class ToolDescriptor(BaseModel):
    """One catalog entry: the tool name, what it does, and its parameter schema."""

    name: str
    description: str
    input_schema: Dict[str, Any] = Field(default_factory=dict)


class ToolCatalog(BaseModel):
    request_id: str
    tools: List[ToolDescriptor] = Field(default_factory=list)
