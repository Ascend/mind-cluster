// Copyright 2026 Huawei Technologies Co., Ltd. Licensed under the Apache License, Version 2.0.
// POST /chat/stream 的流式客户端。
//
// 事件集固定为 start / token / tool_call / tool_result / error / done。
// 用 fetch + ReadableStream 手动分帧（浏览器原生的事件源只支持 GET，承载不了 POST 体），
// 按空行切帧、按 "event:" / "data:" 解析；以 ":" 开头的注释帧（心跳 keep-alive）直接忽略。

const PATH = "/chat/stream";
// 服务端每 15 秒发一帧心跳注释；连续 3 个周期一帧未到即视为流卡死：抛错让上层复位 busy，
// 而不是让页面永远等下去。
const HEARTBEAT_SECONDS = 15;
const STALL_AFTER_MS = 3 * HEARTBEAT_SECONDS * 1000;
const STALL_TEXT = "服务器超过 " + STALL_AFTER_MS / 1000 + " 秒没有下发任何内容，本轮已中止";

function formatDetail(detail) {
  if (typeof detail === "string") {
    return detail;
  }
  if (Array.isArray(detail)) {
    return detail.map((item) => (item && item.msg ? item.msg : JSON.stringify(item))).join("；");
  }
  return JSON.stringify(detail);
}
const EVENTS = ["start", "token", "tool_call", "tool_result", "error", "done"];

function stallError() {
  return new Error(STALL_TEXT);
}

function parseFrame(raw) {
  let event = "";
  const data = [];
  for (const line of raw.split("\n")) {
    if (!line || line.startsWith(":")) {
      continue;
    }
    if (line.startsWith("event:")) {
      event = line.slice(6).trim();
    } else if (line.startsWith("data:")) {
      data.push(line.slice(5).trim());
    }
  }
  if (!event || !EVENTS.includes(event)) {
    return null;
  }
  let payload = {};
  if (data.length) {
    try {
      payload = JSON.parse(data.join("\n"));
    } catch (err) {
      payload = {};
    }
  }
  return { event, payload };
}

/**
 * 发送一轮对话并逐帧回调。handlers = {start, token, tool_call, tool_result, error, done}，
 * 缺失的回调被忽略；网络或解析异常时抛错，由调用方收尾。
 */
export async function streamChat(body, handlers) {
  // 首包看门狗：读循环有自己的逐帧看门狗，但首包 fetch 与错误体读取在此之前——
  // 服务端接了连接却迟迟不下发响应头（反代挂起/处理函数阻塞）时 fetch 永不
  // settle，页面会永久 busy。同一 AbortController 覆盖 fetch 与错误体 json()，
  // 拿到响应体句柄后由读循环的看门狗接管。
  const controller = new AbortController();
  const firstByteTimer = setTimeout(() => controller.abort(), STALL_AFTER_MS);
  let response;
  try {
    response = await fetch(PATH, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      signal: controller.signal,
    });
    if (!response.ok) {
      let detail = String(response.status);
      try {
        const payload = await response.json();
        if (payload && payload.detail) {
          detail = formatDetail(payload.detail);
        }
      } catch (err) {
        if (err && err.name === "AbortError") {
          throw stallError();
        }
        detail = String(response.status);
      }
      throw new Error("请求未被接受：" + detail);
    }
  } catch (err) {
    if (err && err.name === "AbortError") {
      throw stallError();
    }
    throw err;
  } finally {
    clearTimeout(firstByteTimer);
  }
  if (!response.body) {
    throw new Error("浏览器不支持流式响应");
  }
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  try {
    for (;;) {
      let timer = 0;
      const watchdog = new Promise((_, reject) => {
        timer = setTimeout(() => reject(stallError()), STALL_AFTER_MS);
      });
      let chunk;
      try {
        chunk = await Promise.race([reader.read(), watchdog]);
      } finally {
        clearTimeout(timer);
      }
      if (chunk.done) {
        break;
      }
      buffer += decoder.decode(chunk.value, { stream: true });
      let index = buffer.indexOf("\n\n");
      while (index >= 0) {
        const raw = buffer.slice(0, index);
        buffer = buffer.slice(index + 2);
        const frame = parseFrame(raw);
        if (frame && typeof handlers[frame.event] === "function") {
          handlers[frame.event](frame.payload);
        }
        index = buffer.indexOf("\n\n");
      }
    }
  } finally {
    try {
      await reader.cancel();
    } catch (err) {
      // 关流失败（连接已由对端关闭等）不再放大异常：上层只关心帧序列本身
    }
  }
  const tail = parseFrame(buffer);
  if (tail && typeof handlers[tail.event] === "function") {
    handlers[tail.event](tail.payload);
  }
}
