// Copyright 2026 Huawei Technologies Co., Ltd. Licensed under the Apache License, Version 2.0.
// 对话区渲染：用户气泡、助手气泡（token 流式追加 + 富文本渲染）、工具调用折叠区、错误提示。
// 所有文本一律走 textContent 或 views/richtext.js 的 DOM 渲染，页面不写 HTML。

import { renderAnswer } from "./richtext.js";
import { createToolCard } from "./toolcard.js";

function bubble(flow, role) {
  const el = document.createElement("div");
  el.className = "bubble bubble-" + role;
  flow.appendChild(el);
  return el;
}

/** 追加一条用户气泡。 */
export function appendUserBubble(flow, text) {
  const el = bubble(flow, "user");
  const body = document.createElement("div");
  body.className = "msg-text";
  body.textContent = text;
  el.appendChild(body);
  flow.scrollTop = flow.scrollHeight;
  return el;
}

/**
 * 开一轮助手回答：气泡 + （按需出现的）工具调用区。返回 turn 对象给流式回调使用。
 */
export function createAssistantTurn(flow) {
  const el = bubble(flow, "assistant");
  const text = document.createElement("div");
  text.className = "msg-text md-text";
  el.appendChild(text);
  let toolsWrap = null;
  let toolsList = null;
  let pending = [];
  let buffer = "";
  let calls = 0;

  function ensureTools() {
    if (toolsWrap) {
      return toolsList;
    }
    toolsWrap = document.createElement("div");
    toolsWrap.className = "msg-tools";
    const head = document.createElement("div");
    head.className = "msg-tools-head";
    head.textContent = "过程中的能力调用";
    toolsList = document.createElement("div");
    toolsList.className = "msg-tools-list";
    toolsWrap.appendChild(head);
    toolsWrap.appendChild(toolsList);
    el.insertBefore(toolsWrap, text);
    return toolsList;
  }

  return {
    appendToken(chunk) {
      buffer += String(chunk || "");
      renderAnswer(text, buffer);
      flow.scrollTop = flow.scrollHeight;
    },
    addToolCall(payload) {
      calls += 1;
      const card = createToolCard(ensureTools(), payload, { index: calls });
      pending.push({ name: String((payload && payload.name) || ""), card: card });
      flow.scrollTop = flow.scrollHeight;
    },
    setToolResult(payload) {
      const name = String((payload && payload.name) || "");
      const index = pending.findIndex((item) => item.name === name);
      // 名字对不上就不回填：按顺序兜底会把结果写到无关工具卡上，误导用户
      const entry = index >= 0 ? pending.splice(index, 1)[0] : null;
      if (entry) {
        entry.card.setResult(payload && payload.output);
      }
      flow.scrollTop = flow.scrollHeight;
    },
    finish(note) {
      if (!buffer && !toolsWrap) {
        text.textContent = note || "（本轮没有返回内容）";
      }
    },
  };
}

/** 一条独立提示（错误、引导、中断），不进气泡正文。 */
export function appendNotice(flow, message, kind) {
  const el = document.createElement("div");
  el.className = "notice" + (kind ? " " + kind : "");
  el.textContent = String(message || "");
  flow.appendChild(el);
  flow.scrollTop = flow.scrollHeight;
  return el;
}
