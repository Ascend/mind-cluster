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

"""Bounded environment parsing for HTTPS settings (integers, floats, nan/inf rejection)."""

from __future__ import annotations

import logging
import math
import os

logger = logging.getLogger(__name__)


def env_int(name: str, default: int, minimum: int, maximum: int) -> int:
    """Return the integer env value, falling back to ``default`` when missing or invalid."""
    raw = os.environ.get(name)
    if raw is None:
        return default
    try:
        value = int(raw)
    except (TypeError, ValueError):
        logger.warning("%s is not an integer; using %d", name, default)
        return default
    if value < minimum or value > maximum:
        logger.warning("%s is outside [%d, %d]; using %d", name, minimum, maximum, default)
        return default
    return value


def env_float(name: str, default: float, minimum: float, maximum: float) -> float:
    """Return the float env value, falling back to ``default`` when missing or invalid.

    ``nan``/``inf`` are rejected first: they compare False against every bound and would
    slip through the range check (for example a NaN rate-limit window).
    """
    raw = os.environ.get(name)
    if raw is None:
        return default
    try:
        value = float(raw)
    except (TypeError, ValueError):
        logger.warning("%s is not a number; using %g", name, default)
        return default
    if not math.isfinite(value) or value < minimum or value > maximum:
        logger.warning("%s is outside [%g, %g]; using %g", name, minimum, maximum, default)
        return default
    return value
