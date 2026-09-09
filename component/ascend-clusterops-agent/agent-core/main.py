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
  CLI : python -m agent_core.main --job <jobname> [-n namespace]
  HTTP: uvicorn agent_core.main:app --host 0.0.0.0 --port 9700
        curl -X POST localhost:9700/diag -d '{"job":"job-x","namespace":"default"}'
"""

from __future__ import annotations

import json
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from pydantic import BaseModel

from agent_core import llm, pathmap, relcache, tools
from agent_core.agent import LLMSummaryError, format_diag_report, summarize_report
from agent_core.upload import start_upload_server
from clusterops_common.log import init_logging

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


class DiagReq(BaseModel):
    job: str  # jobname (task CR name), required
    namespace: str | None = "default"
    refresh: bool | None = False  # True: ignore cache, force re-run and refresh


@app.post("/diag")
def diag(req: DiagReq) -> dict:
    """Deterministic main flow: diagnose_cached(job) with guard + cache + (optional) LLM summary."""
    logger.info("diag request received: %s", json.dumps(req.model_dump(), ensure_ascii=False))
    try:
        r = tools.diagnose_cached(job=req.job, namespace=req.namespace or "default", refresh=bool(req.refresh))
    except Exception as e:  # noqa: BLE001  # fallback: turn any unexpected exception into a readable error, avoid bare 500
        logger.exception("/diag handling failed: job=%s ns=%s", req.job, req.namespace)
        return {"error": f"诊断错误，诊断过程发生异常，请查看 agent-core 日志 ({e})", "pods": [], "diag_report": {}}
    if r.get("error"):
        # Diagnosis failed (task missing / collect failed / diag failed / timeout): use the reason as the final text
        r["final_text"] = r["error"]
        return r
    if r.get("final_text"):
        # Cache hit with a cached LLM summary -> reuse it directly
        return r
    report_text = format_diag_report(r.get("diag_report") or {})
    llm_error = None
    try:
        text = summarize_report(r.get("diag_report") or {})
        final_text = text or report_text
    except LLMSummaryError as e:  # mis-configured/unreachable LLM -> tell the user, keep the deterministic report
        llm_error = str(e)
        final_text = f"{e}\n\n{report_text}"
    except Exception as e:  # noqa: BLE001  # unexpected summarizer bug -> keep the deterministic report, never 500
        logger.exception("LLM report summarization crashed: %s", e)
        llm_error = "诊断错误，报告总结过程发生异常，请查看 agent-core 日志"
        final_text = report_text
    r["final_text"] = final_text
    if llm_error:
        r["llm_error"] = llm_error
    # Write back the LLM summary into the cache for later cache hits
    tools.cache_diag_final_text(req.job, req.namespace or "default", final_text)
    return r


def main() -> None:
    import argparse

    ap = argparse.ArgumentParser(prog="agent_core.main", description="Ascend 故障诊断 (按 jobname 查中心关系表)")
    ap.add_argument("--job", required=True, help="任务名 (jobname, 任务 CR 名)")
    ap.add_argument("-n", "--namespace", default="default", help="命名空间")
    ap.add_argument("--refresh", action="store_true", help="忽略缓存强制重跑并刷新")
    args = ap.parse_args()
    # CLI mode: logs to disk + start the relation cache locally (kubeconfig when debugging outside the cluster)
    init_logging(log_dir="/var/log/mindx-dl/agent-core", log_file="agent-core.log")
    relcache.init_cache()
    try:
        logger.info("CLI diagnosis started: job=%s ns=%s refresh=%s", args.job, args.namespace, args.refresh)
        r = tools.diagnose_cached(job=args.job, namespace=args.namespace, refresh=args.refresh)
        if r.get("error"):
            # Diagnosis failed (task missing / collect failed / diag failed / timeout): log it
            logger.error("CLI diagnosis failed: %s", r["error"])
        else:
            if r.get("final_text"):
                text = r["final_text"]  # cache hit with a cached LLM summary -> reuse it
            else:
                report_text = format_diag_report(r.get("diag_report") or {})
                try:
                    text = summarize_report(r.get("diag_report") or {})
                except (
                    LLMSummaryError
                ) as e:  # mis-configured/unreachable LLM -> show the reason, keep the deterministic report
                    text = f"{e}\n\n{report_text}"
                except Exception as e:  # noqa: BLE001  # unexpected summarizer bug -> keep the deterministic report
                    logger.exception("LLM report summarization crashed: %s", e)
                    text = report_text
                tools.cache_diag_final_text(
                    args.job, args.namespace, text
                )  # write back the LLM summary for later cache hits
            logger.info("CLI diagnosis result: %s", text or json.dumps(r, ensure_ascii=False, default=str)[:4000])
        logger.info("CLI diagnosis finished: job=%s error=%s", args.job, r.get("error"))
    finally:
        relcache.shutdown_cache()


if __name__ == "__main__":
    main()
