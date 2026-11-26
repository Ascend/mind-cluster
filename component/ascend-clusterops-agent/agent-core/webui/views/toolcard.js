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
// 工具调用折叠卡：只有一种卡，前端不认识任何具体工具。
//
// tool_result 只带 name + output：能力调用是工具层信封 JSON，技能正文是 Markdown。
// 能解析成 JSON 的按缩进美化展示，其余原样展示；两者都能整段复制。
//
// 纯函数写在文件上半部，import 时不接触 document：可在 node 里直接跑解析断言。

const MS_PER_SECOND = 1000;
const FINE_SECONDS_MS = 10000; // 低于这个值，秒数保留一位小数才读得准
const COPY_FEEDBACK_MS = 1200; // 复制按钮回显「已复制」的时长
const INDENT = 2; // 信封 JSON 的展示缩进

// null 原型：status 来自工具返回，查表键如 constructor/toString 不得沿原型链误命中
const STATUS_TABLE = Object.freeze(
  Object.assign(Object.create(null), {
    ok: { text: "成功", tone: "ok" },
    partial: { text: "部分", tone: "warn" },
    error: { text: "失败", tone: "error" },
    unknown: { text: "未知", tone: "" },
  })
);

function statusEntry(status) {
  return STATUS_TABLE[String(status || "").toLowerCase()] || STATUS_TABLE.unknown;
}

/** 毫秒 → 人类可读（客户端计时用）。 */
export function formatMs(ms) {
  const value = Number(ms);
  if (!Number.isFinite(value) || value < 0) {
    return "";
  }
  if (value < MS_PER_SECOND) {
    return Math.round(value) + "ms";
  }
  return (value / MS_PER_SECOND).toFixed(value < FINE_SECONDS_MS ? 1 : 0) + "s";
}

/** 状态 → 中文。 */
export function statusText(status) {
  return statusEntry(status).text;
}

/** 状态 → 卡片色调类名后缀（ok / warn / error / 空）。 */
export function statusTone(status) {
  return statusEntry(status).tone;
}

/**
 * 解析一段 tool_result 文本。工具信封是 JSON（status / tool / request_id 都在里面），
 * 技能正文是 Markdown；解析不出对象时 envelope 为 null，show 就是原文。
 */
export function parseEnvelope(output) {
  const raw = String(output === undefined || output === null ? "" : output);
  const parsed = { raw: raw, envelope: null, tool: "", status: "unknown", requestId: "", show: raw };
  let value = null;
  try {
    value = JSON.parse(raw);
  } catch (err) {
    return parsed;
  }
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return parsed;
  }
  parsed.envelope = value;
  parsed.tool = String(value.tool || "");
  parsed.status = String(value.status || "unknown").toLowerCase();
  parsed.requestId = String(value.request_id || "");
  parsed.show = JSON.stringify(value, null, INDENT);
  return parsed;
}

function nowMs() {
  if (typeof performance !== "undefined" && performance.now) {
    return performance.now();
  }
  return Date.now();
}

function element(tag, className, text) {
  const node = document.createElement(tag);
  if (className) {
    node.className = className;
  }
  if (text !== undefined && text !== null) {
    node.textContent = String(text);
  }
  return node;
}

function tag(key, value) {
  const holder = element("span", "tool-tag");
  holder.appendChild(element("span", "tool-tag-key", key));
  holder.appendChild(element("span", "tool-tag-val", value));
  return holder;
}

function fillTags(view, parsed) {
  view.textContent = "";
  view.hidden = parsed.envelope === null;
  if (parsed.envelope === null) {
    return;
  }
  view.appendChild(tag("状态", statusText(parsed.status)));
  if (parsed.tool) {
    view.appendChild(tag("工具", parsed.tool));
  }
  if (parsed.requestId) {
    view.appendChild(tag("请求", parsed.requestId));
  }
}

/**
 * 建一张卡并挂到容器里。options.index 是「本轮第几次能力调用」（1 起，0 表示不显示编号）。
 * 返回 {card, setResult(output)}：结果到达时补状态、耗时与返回内容。
 */
export function createToolCard(container, payload, options) {
  const opts = options || {};
  const index = Number(opts.index);
  const name = String((payload && payload.name) || "工具");
  const started = nowMs();

  const card = document.createElement("div");
  card.className = "tool-card tool-pending";

  const head = document.createElement("button");
  head.type = "button";
  head.className = "tool-card-head";
  head.appendChild(element("span", "tool-icon", "🔧"));
  head.appendChild(element("span", "tool-block-title", name));
  if (Number.isFinite(index) && index > 0) {
    head.appendChild(element("span", "tool-index", "#" + index));
  }
  const status = element("span", "tool-card-status running", "调用中…");
  head.appendChild(status);
  const elapsed = element("span", "tool-elapsed", "");
  head.appendChild(elapsed);

  const body = element("div", "tool-card-body");
  body.hidden = true;

  const argsBlock = element("div", "tool-block");
  argsBlock.appendChild(element("div", "tool-block-title", "入参"));
  const argsPre = element("pre", "tool-args");
  let argsText = "";
  try {
    argsText = JSON.stringify((payload && payload.args) || {}, null, INDENT);
  } catch (err) {
    argsText = String((payload && payload.args) || "");
  }
  argsPre.textContent = argsText;
  argsBlock.appendChild(argsPre);
  body.appendChild(argsBlock);

  const outBlock = element("div", "tool-block");
  outBlock.appendChild(element("div", "tool-block-title", "返回"));
  const view = element("div", "tool-view");
  view.hidden = true;
  const outPre = element("pre", "tool-summary");
  outPre.textContent = "（等待返回…）";
  outBlock.appendChild(view);
  outBlock.appendChild(outPre);
  body.appendChild(outBlock);

  const foot = element("div", "tool-card-foot");
  const copy = element("button", "tool-copy", "复制返回");
  copy.type = "button";
  copy.disabled = true;
  foot.appendChild(copy);
  body.appendChild(foot);

  head.addEventListener("click", () => {
    body.hidden = !body.hidden;
  });

  let parsed = null;

  copy.addEventListener("click", () => {
    const text = parsed ? parsed.raw : outPre.textContent;
    const done = () => {
      copy.textContent = "已复制";
      setTimeout(() => {
        copy.textContent = "复制返回";
      }, COPY_FEEDBACK_MS);
    };
    if (typeof navigator !== "undefined" && navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(done, () => {
        copy.textContent = "复制失败";
      });
      return;
    }
    copy.textContent = "不支持复制";
  });

  card.appendChild(head);
  card.appendChild(body);
  container.appendChild(card);

  return {
    card,
    setResult(output) {
      parsed = parseEnvelope(output);
      outPre.textContent = parsed.show;
      fillTags(view, parsed);
      elapsed.textContent = formatMs(nowMs() - started);
      card.classList.remove("tool-pending");
      const tone = statusTone(parsed.status);
      card.classList.add("tool-" + (tone || "unknown"));
      status.textContent = parsed.envelope === null ? "完成" : statusText(parsed.status);
      const toneClass = { error: "err", warn: "warn", ok: "ok" }[tone] || "";
      status.className = ["tool-card-status", toneClass].filter(Boolean).join(" ");
      copy.disabled = false;
    },
  };
}
