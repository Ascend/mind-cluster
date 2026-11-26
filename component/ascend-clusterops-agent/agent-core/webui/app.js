/* Copyright 2026 Huawei Technologies Co., Ltd.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
// WebUI 装配：唯一页面状态 = 当前会话 id（浏览器生成）。
// 页面能力：对话气泡 token 流式、工具调用折叠、新会话按钮、无模型引导；api-key 只存服务端。
// 首屏「有没有模型」来自页面内联的 boot 对象（GET /ui 注入），页面不单独请求状态接口。
// 空会话：主区带 is-welcome，灰色问候语 + 输入卡垂直居中；发出首条消息后移除 is-welcome
// 回到底部布局，「＋ 新建会话」恢复空态。
// 不做的：会话列表、历史点回、刷新恢复——会话是进程内记忆，重启即失效。
// 版式（左栏 / 主区 / 悬浮输入卡）见 style.css 分节注释。

import { streamChat } from "./sse.js";
import { appendUserBubble, createAssistantTurn, appendNotice } from "./views/chat.js";

const pane = document.getElementById("chat-pane");
const flow = document.getElementById("chat-flow");
const form = document.getElementById("chat-form");
const input = document.getElementById("chat-input");
const sendBtn = document.getElementById("send-btn");
const badge = document.getElementById("model-badge");
const newSessionBtn = document.getElementById("new-session-btn");

// 无模型提示：占位符与服务端 error 事件共用同一句话，字面一致由静态门禁钉住。
const NO_LLM_HINT = "未配置大模型，对话不可用；请配置模型后使用";

const READY_PLACEHOLDER = "输入运维问题，Enter 发送；Shift+Enter 换行";
const INPUT_MAX_HEIGHT = 200;
const SESSION_ID_MAX_CHARS = 64;

const state = { sessionId: newSessionId(), busy: false };

function newSessionId() {
  const raw =
    typeof crypto !== "undefined" && crypto.randomUUID
      ? crypto.randomUUID()
      : "s" + Date.now().toString(36) + Math.random().toString(36).slice(2);
  return String(raw).replace(/[^A-Za-z0-9_-]/g, "").slice(0, SESSION_ID_MAX_CHARS) || "s" + Date.now().toString(36);
}

function setBadge(available) {
  if (available === null) {
    badge.textContent = "LLM：状态未知";
    badge.className = "badge badge-unknown";
    return;
  }
  badge.textContent = available ? "LLM：就绪" : "LLM：未配置";
  badge.className = "badge " + (available ? "badge-ok" : "badge-unknown");
}

// 首屏状态随页面下发（服务端 GET /ui 内联 window.__AGENT_BOOT__）：页面不发状态请求，
// 没有 boot 对象时（例如直接打开本地文件）角标显示「状态未知」、占位符用通用文案。
function applyBootState() {
  const boot = typeof window === "undefined" ? null : window.__AGENT_BOOT__;
  const available = boot && typeof boot.llm_available === "boolean" ? boot.llm_available : null;
  setBadge(available);
  input.placeholder = available === false ? NO_LLM_HINT : READY_PLACEHOLDER;
}

// 悬浮输入卡随内容自动变高：先回到 auto 再量 scrollHeight，封顶后交给内部滚动。
function fitInput() {
  input.style.height = "auto";
  input.style.height = Math.min(input.scrollHeight, INPUT_MAX_HEIGHT) + "px";
}

input.addEventListener("input", fitInput);
fitInput();

input.addEventListener("keydown", (event) => {
  // Enter 发送、Shift+Enter 换行；输入法组词中的 Enter 不当发送。
  if (event.key === "Enter" && !event.shiftKey && !event.isComposing) {
    event.preventDefault();
    form.requestSubmit();
  }
});

function setBusy(busy) {
  state.busy = busy;
  sendBtn.disabled = busy;
  input.disabled = busy;
  if (busy) {
    input.value = "";
    fitInput();
  }
  input.focus();
}

async function submitTurn(text) {
  // 发出首条消息即离开空态：输入卡回到底部，页面展示回答或无模型提示。
  pane.classList.remove("is-welcome");
  appendUserBubble(flow, text);
  const turn = createAssistantTurn(flow);
  setBusy(true);
  let interrupted = false;
  try {
    await streamChat(
      { session_id: state.sessionId, message: text },
      {
        token: (payload) => turn.appendToken(payload && payload.text),
        tool_call: (payload) => turn.addToolCall(payload),
        tool_result: (payload) => turn.setToolResult(payload),
        // error 帧是服务端已明确告知的状态（如无模型引导），提示行即说明，不标中断
        error: (payload) => appendNotice(flow, (payload && payload.message) || "本轮失败", "notice-error"),
      }
    );
  } catch (err) {
    interrupted = true;
    appendNotice(flow, String((err && err.message) || err), "notice-error");
  } finally {
    turn.finish(interrupted ? "（本轮中断）" : undefined);
    setBusy(false);
  }
}

form.addEventListener("submit", (event) => {
  event.preventDefault();
  if (state.busy) {
    return;
  }
  const text = input.value.trim();
  if (!text) {
    return;
  }
  submitTurn(text);
});

newSessionBtn.addEventListener("click", () => {
  if (state.busy) {
    return;
  }
  state.sessionId = newSessionId();
  flow.textContent = "";
  pane.classList.add("is-welcome");
  input.focus();
});

applyBootState();
