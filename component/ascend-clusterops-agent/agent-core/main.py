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
from pathlib import Path
from typing import Optional

from agent_core import llm, pathmap, relcache
from agent_core.adapters.tool_adapter import ToolAdapter
from agent_core.api import stream_api
from agent_core.chat.skill_loader import SkillLoader
from agent_core.chat.tool_control import AgentTools, ToolControl
from agent_core.diagnosis_presentation import present_diagnosis, summarize_diagnosis
from agent_core.https.auth_middleware import WebUIAuthMiddleware
from agent_core.query_api import router as query_router
from agent_core.tools import diagnosis as tools
from agent_core.tools.diagnosis_adapter import diagnosis_failure
from agent_core.upload import start_upload_server
from clusterops_common.log import init_logging
from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse, RedirectResponse
from fastapi.staticfiles import StaticFiles
from langgraph.checkpoint.memory import MemorySaver
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
    try:
        # 对话链路自持一套装配：工具层适配 + 调用控制 + 技能宿主。
        chat_tools = AgentTools(adapter=ToolAdapter(), control=ToolControl(), skills=SkillLoader())
        # 记忆用 langgraph 自己的进程级 checkpointer：thread_id = session_id，重启即丢（已知限制）。
        stream_api.init_stream(llm_provider=llm.get_llm, tools=chat_tools, checkpointer=MemorySaver())
    except Exception:
        # 对话装配失败只让对话面降级（未装配的流答 503），绝不能把已有的 /diag 入口一起拖下来。
        logger.warning("chat assembly failed; the chat surface stays offline", exc_info=True)
    logger.info(
        "agent-core started: relcache + pathmap-writer + llm-secret-watch + upload gRPC :9710 "
        "+ chat POST /chat/stream + skill host"
    )
    yield
    logger.info("agent-core stopping: shutting down upload gRPC / pathmap-writer / relcache / llm-secret-watch")
    if _upload_server is not None:
        _upload_server.stop(0)
    llm.shutdown_watch()
    pathmap.shutdown_writer()
    relcache.shutdown_cache()


app = FastAPI(lifespan=lifespan)
app.add_middleware(WebUIAuthMiddleware)


@app.post("/api/v1/auth/check", include_in_schema=False)
async def webui_authentication() -> dict:
    """Probe for the shared HTTPS token.

    The middleware validates first, so reaching this function means the credentials are
    correct. Invalid credentials are answered with 200 and ``{"ok": false, ...}`` before
    the request gets here; the token is never echoed back.
    """
    return {"ok": True}


app.include_router(query_router)

app.include_router(stream_api.router)

# WebUI 页面与静态资源。浏览器唯一能得知「有没有配模型」的地方就是页面本身：index.html 里留了
# 标记，请求时替换成一段内联 boot 对象。不额外开接口——这一面只有 GET /ui 一个页面入口。
_UI_ROOT = Path(__file__).resolve().parent / "webui"
_UI_INDEX = _UI_ROOT / "index.html"
_UI_BOOT_MARKER = "<!--AGENT_BOOT-->"


def _llm_ready() -> bool:
    """此刻是否已配置模型（llm-secret watch 让它保持实时）。"""
    try:
        return llm.get_llm() is not None
    except Exception:  # 密钥坏了不能把页面一起弄挂：按「没有模型」渲染
        logger.warning("llm probe raised while rendering the page", exc_info=True)
        return False


def render_ui() -> str:
    """index.html 内联首屏状态（角标与输入框占位符都取自它）。"""
    page = _UI_INDEX.read_text(encoding="utf-8")
    boot = json.dumps({"llm_available": _llm_ready()}, ensure_ascii=False)
    return page.replace(_UI_BOOT_MARKER, "<script>window.__AGENT_BOOT__ = " + boot + ";</script>")


if _UI_INDEX.is_file():

    @app.get("/ui", include_in_schema=False)
    def webui_page_entry(request: Request) -> RedirectResponse:
        """裸 /ui 跳到 /ui/，页面里的相对资源路径才解析得对。

        页面用相对路径引用 style.css / app.js：停在裸 /ui 时浏览器会按站点根解析，全部 404。
        """
        return RedirectResponse(
            str(request.url.replace(path="/ui/")), status_code=307, headers={"Cache-Control": "no-store"}
        )

    @app.get("/ui/", include_in_schema=False)
    @app.get("/ui/index.html", include_in_schema=False)
    def webui_page() -> HTMLResponse:
        """WebUI 入口页面（boot 状态是实时的，页面一律不缓存）。"""
        return HTMLResponse(render_ui(), headers={"Cache-Control": "no-store"})


# /ui 下的静态资源（上面两条页面路由先接住 HTML 入口）。
if _UI_ROOT.is_dir():
    app.mount("/ui", StaticFiles(directory=str(_UI_ROOT), html=True), name="ui")


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
