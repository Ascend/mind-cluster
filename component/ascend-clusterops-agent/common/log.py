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

"""Shared logging for agent-core / node-collector: file persistence (RotatingFileHandler
10MB x 10) + stdout.

Refer to ascend-faultdiag's log persistence (same DETAIL_FORMAT).
The log directory is pre-created on the host by each role's init container and chowned to
the running user; if file persistence fails (unwritable dir / volume not mounted), it falls
back to stdout only without blocking app startup.
The LOG_DIR env var can override the log directory (for local debugging/testing).
"""

from __future__ import annotations

import logging
import os
import sys
from logging.handlers import RotatingFileHandler
from pathlib import Path

LOG_MAX_SIZE = 10 * 1024 * 1024  # 10MB
LOG_MAX_BACKUP_COUNT = 10
DETAIL_FORMAT = (
    '[%(asctime)s] %(levelname)-7s[%(name)s][Process %(processName)s:%(process)d][%(filename)s:%(lineno)d] %(message)s'
)

logger = logging.getLogger(__name__)

_configured = False
_APP_HANDLERS: list[logging.Handler] = []


def init_logging(log_dir: str, log_file: str, level: int = logging.INFO) -> logging.Logger:
    """Initialize logging (file persistence + stdout), idempotent.

    log_dir: log directory (overridable via the LOG_DIR env var), log_file: log file name.
    Handlers are attached to the root logger: logs from app child loggers and third-party
    libraries (k8s/grpc etc.) are persisted as well.
    Returns the root logger.
    """
    global _configured
    if _configured:
        return logging.getLogger()
    log_path = Path(os.environ.get("LOG_DIR", log_dir))
    formatter = logging.Formatter(DETAIL_FORMAT)
    handlers = [logging.StreamHandler(sys.stdout)]
    try:
        log_path.mkdir(parents=True, exist_ok=True)
        file_handler = RotatingFileHandler(log_path / log_file, maxBytes=LOG_MAX_SIZE, backupCount=LOG_MAX_BACKUP_COUNT)
        handlers.append(file_handler)
    except OSError as e:  # unwritable dir / volume not mounted -> stdout only, no startup block
        logger.warning("log file unavailable (%s), fallback to stdout only", e)

    root = logging.getLogger()
    root.setLevel(level)
    for h in handlers:
        h.setFormatter(formatter)
        h.setLevel(level)
        root.addHandler(h)
    _APP_HANDLERS.extend(handlers)
    _configured = True
    return root


def reset_logging() -> None:
    """Remove the handlers attached by init_logging and reset the flag (for tests)."""
    global _configured
    root = logging.getLogger()
    for h in list(root.handlers):
        if h in _APP_HANDLERS:
            root.removeHandler(h)
    _APP_HANDLERS.clear()
    _configured = False
