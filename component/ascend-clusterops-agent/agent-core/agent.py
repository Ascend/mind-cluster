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

"""Optional LLM helpers for the diagnosis report.

The main flow is the deterministic diagnose tool (see ``tools/diagnosis.py``). The LLM is
only used to summarize an already deterministic diagnosis report; that logic lives in
``llm_summary.py`` and is re-exported here for callers using the historical
``agent_core.agent`` import path.
"""

from __future__ import annotations

from agent_core.llm_summary import LLMSummaryError, summarize_report

__all__ = ["LLMSummaryError", "summarize_report"]
