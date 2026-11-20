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

"""Diagnosis agent entry point.

Usage:
  HTTP: uvicorn agent_core.main:app --host 0.0.0.0 --port 9700
        curl -X POST localhost:9700/diag -d '{"job":"job-x","namespace":"default"}'
"""

from __future__ import annotations

import json
import logging
from contextlib import asynccontextmanager
from typing import Optional

from agent_core import llm, pathmap, relcache
from agent_core.diagnosis_presentation import present_diagnosis, summarize_diagnosis
from agent_core.query_api import router as query_router
from agent_core.tools import diagnosis as tools
from agent_core.tools.diagnosis_adapter import diagnosis_failure
from agent_core.upload import start_upload_server
from clusterops_common.log import init_logging
from fastapi import FastAPI
from pydantic import BaseModel

logger = logging.getLogger(__name__)

_upload_server = None


@asynccontextmanager
async def lifespan(_app):
    global _upload_server
    init_logging(log_dir="/var/log/mindx-dl/agent-core", log_file="agent-core.log")  # write logs to disk
    relcache.init_cache()  # start central relation cache (pod metadata informer + CM snapshot)
    pathmap.init_writer()  # start pathmap-writer (watch all pods -> mount pairs -> global CM)
    llm.init_watch()  # start llm-secret watch (real-time LLM config, no per-call k8s read)
    _upload_server = start_upload_server()  # collector -> agent upload gRPC service (:9710)
    logger.info("agent-core started: relcache + pathmap-writer + llm-secret-watch + upload gRPC :9710")
    yield
    logger.info("agent-core stopping: shutting down upload gRPC / pathmap-writer / relcache / llm-secret-watch")
    if _upload_server is not None:
        _upload_server.stop(0)
    llm.shutdown_watch()
    pathmap.shutdown_writer()
    relcache.shutdown_cache()


app = FastAPI(lifespan=lifespan)
app.include_router(query_router)


class DiagReq(BaseModel):
    job: str  # jobname (task CR name), required
    namespace: Optional[str] = "default"
    refresh: Optional[bool] = False  # True: ignore cache, force re-run and refresh


@app.post("/diag")
def diag(req: DiagReq) -> dict:
    """Legacy /diag entry: the original deterministic chain for ``kubectl ascend_diag``.

    Unchanged original path: ``diagnose_cached`` -> legacy presentation. No tool
    registry, no tool argument model, no result envelope; the dict returned by
    ``diagnose_cached`` is presented in place and returned as-is.
    """
    logger.info("diag request received: %s", json.dumps(req.model_dump(), ensure_ascii=False))
    try:
        r = tools.diagnose_cached(job=req.job, namespace=req.namespace or "default", refresh=bool(req.refresh))
    except Exception as e:  # noqa: BLE001  # fallback: turn any unexpected exception into a readable error, avoid bare 500
        logger.exception("/diag handling failed: job=%s ns=%s", req.job, req.namespace)
        return diagnosis_failure(e)
    # The V2 chain adds internal execution state; the legacy chain had no such field
    # and it must never appear in the /diag body.
    r.pop("_status", None)
    return present_diagnosis(summarize_diagnosis(r, req.job, req.namespace or "default"))
