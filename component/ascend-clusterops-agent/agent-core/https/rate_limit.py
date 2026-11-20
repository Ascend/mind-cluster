#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""In-process global rate limiting: all HTTPS requests share one sliding window.

NodePort SNAT hides the client address, so per-source limiting would degrade to a
global window. A sliding window avoids the double burst that a fixed window allows
at its boundaries.
"""

from __future__ import annotations

import threading
from collections import deque
from time import monotonic


class SlidingWindowRateLimiter:
    """Sliding-window limiter that keeps the timestamps of accepted requests only."""

    def __init__(self, max_requests: int, window_seconds: float):
        self.max_requests = max_requests
        self.window_seconds = window_seconds
        self._timestamps = deque()
        self._lock = threading.Lock()

    def allow_request(self) -> bool:
        """Record one request; True when it is within the quota.

        Rejected requests are not recorded, so the window recovers and the queue stays
        bounded by ``max_requests``.
        """
        now = monotonic()
        with self._lock:
            cutoff = now - self.window_seconds
            while self._timestamps and self._timestamps[0] <= cutoff:
                self._timestamps.popleft()
            if len(self._timestamps) >= self.max_requests:
                return False
            self._timestamps.append(now)
            return True
