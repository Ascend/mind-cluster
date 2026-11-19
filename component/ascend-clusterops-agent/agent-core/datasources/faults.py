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

"""Shared construction of per-code fault details (device list and node summary)."""

from __future__ import annotations

from typing import Iterable, List, Optional

from agent_core.query_schemas import FaultDetail


def fault_detail(fault_code: Optional[str], handling: Optional[str], last_time: Optional[str]) -> FaultDetail:
    """One public fault-code detail; None fields stay None, matching the producer gaps."""
    return FaultDetail(fault_code=fault_code, fault_handling=handling, last_time=last_time)


def build_fault_details(codes: Iterable[str], handling: Optional[str], last_time: Optional[str]) -> List[FaultDetail]:
    """One detail per code; a fault without any code still yields one placeholder detail."""
    details = [fault_detail(code, handling, last_time) for code in codes]
    return details or [fault_detail(None, handling, last_time)]
