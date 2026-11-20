// Copyright 2026 Huawei Technologies Co., Ltd. Licensed under the Apache License, Version 2.0.
// 回答的富文本渲染：把模型输出里的 Markdown 子集拼成 DOM 节点。
// 支持：段落、1~4 级标题、单层有序/无序列表、GFM 管道表格、**加粗**、行内代码。
// 其余标记（斜体/链接/图片/HTML/围栏）一律按纯文本原样输出。全程 createElement +
// textContent 拼节点，页面从不拼 HTML 字符串，因此不需要消毒器。

const HEADING = /^(#{1,4})\s+(.*)$/;
const BULLET_ITEM = /^\s*[-*+]\s+(.*)$/;
const ORDERED_ITEM = /^\s*\d+[.)]\s+(.*)$/;
const SEPARATOR_CELL = /^:?-+:?$/;
const PIPE = "|";

function element(tag) {
  return document.createElement(tag);
}

function withInline(tag, text) {
  const node = element(tag);
  appendInline(node, text);
  return node;
}

/** 表格行拆单元格：去掉首尾竖线，按竖线切分，两端空白裁掉。 */
function splitRow(line) {
  let text = line.trim();
  if (text.startsWith(PIPE)) {
    text = text.slice(1);
  }
  if (text.endsWith(PIPE)) {
    text = text.slice(0, -1);
  }
  return text.split(PIPE).map((cell) => cell.trim());
}

/** 分隔行：每个单元格都是 --- / :-- / --: 这类，才是表头下面那条杠。 */
function isSeparatorRow(line) {
  return line.includes(PIPE) && splitRow(line).every((cell) => SEPARATOR_CELL.test(cell));
}

/** 表格块：表头 + 分隔行 + 若干数据行；缺分隔行或列数对不上就不算表格。 */
function tableBlock(lines, index) {
  const next = lines[index + 1];
  if (!lines[index].includes(PIPE) || next === undefined || !isSeparatorRow(next)) {
    return null;
  }
  const header = splitRow(lines[index]);
  if (header.length !== splitRow(next).length) {
    return null;
  }
  const rows = [];
  let cursor = index + 2;
  while (cursor < lines.length && lines[cursor].trim() && lines[cursor].includes(PIPE)) {
    rows.push(splitRow(lines[cursor]));
    cursor += 1;
  }
  return { block: { kind: "table", header: header, rows: rows }, next: cursor };
}

function headingBlock(lines, index) {
  const matched = HEADING.exec(lines[index]);
  if (!matched) {
    return null;
  }
  return { block: { kind: "heading", level: matched[1].length, text: matched[2].trim() }, next: index + 1 };
}

function listBlock(lines, index, ordered) {
  const pattern = ordered ? ORDERED_ITEM : BULLET_ITEM;
  const items = [];
  let cursor = index;
  while (cursor < lines.length) {
    const matched = pattern.exec(lines[cursor]);
    if (!matched) {
      break;
    }
    items.push(matched[1].trim());
    cursor += 1;
  }
  return { block: { kind: "list", ordered: ordered, items: items }, next: cursor };
}

/** 段落块：连续非空、又不开启新块的行，按原行序保留（浏览器把换行折成空格）。 */
function paragraphBlock(lines, index) {
  const parts = [];
  let cursor = index;
  while (cursor < lines.length && lines[cursor].trim() && !startsBlock(lines, cursor)) {
    parts.push(lines[cursor].trim());
    cursor += 1;
  }
  return { block: { kind: "paragraph", text: parts.join("\n") }, next: cursor };
}

function startsBlock(lines, index) {
  const line = lines[index];
  return HEADING.test(line) || BULLET_ITEM.test(line) || ORDERED_ITEM.test(line) || tableBlock(lines, index) !== null;
}

/** 切块：空行分段；表格/标题/列表各自成块；其余非空行并成段落。 */
function parseBlocks(lines) {
  const blocks = [];
  let index = 0;
  while (index < lines.length) {
    if (!lines[index].trim()) {
      index += 1;
      continue;
    }
    const parsed =
      tableBlock(lines, index) ||
      headingBlock(lines, index) ||
      (BULLET_ITEM.test(lines[index]) ? listBlock(lines, index, false) : null) ||
      (ORDERED_ITEM.test(lines[index]) ? listBlock(lines, index, true) : null) ||
      paragraphBlock(lines, index);
    blocks.push(parsed.block);
    index = parsed.next;
  }
  return blocks;
}

/** 行内：反引号先于 ** 处理（代码里的星号不加粗）；未闭合的标记按原文输出。 */
function appendInline(parent, text) {
  let cursor = 0;
  let plain = "";
  const flush = () => {
    if (plain) {
      parent.appendChild(document.createTextNode(plain));
      plain = "";
    }
  };
  while (cursor < text.length) {
    if (text[cursor] === "`") {
      const end = text.indexOf("`", cursor + 1);
      if (end > cursor + 1) {
        flush();
        const code = element("code");
        code.textContent = text.slice(cursor + 1, end);
        parent.appendChild(code);
        cursor = end + 1;
        continue;
      }
    }
    if (text.startsWith("**", cursor)) {
      const end = text.indexOf("**", cursor + 2);
      if (end > cursor + 2) {
        flush();
        const strong = element("strong");
        appendInline(strong, text.slice(cursor + 2, end));
        parent.appendChild(strong);
        cursor = end + 2;
        continue;
      }
    }
    plain += text[cursor];
    cursor += 1;
  }
  flush();
}

function renderTable(block) {
  const table = element("table");
  const head = element("thead");
  const headRow = element("tr");
  for (const cell of block.header) {
    headRow.appendChild(withInline("th", cell));
  }
  head.appendChild(headRow);
  table.appendChild(head);
  const body = element("tbody");
  for (const row of block.rows) {
    const rowNode = element("tr");
    // 按表头列数对齐：缺的补空单元格、多的丢掉，表格不塌
    for (let column = 0; column < block.header.length; column += 1) {
      rowNode.appendChild(withInline("td", row[column] || ""));
    }
    body.appendChild(rowNode);
  }
  table.appendChild(body);
  return table;
}

function renderBlock(block) {
  if (block.kind === "heading") {
    return withInline("h" + block.level, block.text);
  }
  if (block.kind === "list") {
    const list = element(block.ordered ? "ol" : "ul");
    for (const item of block.items) {
      list.appendChild(withInline("li", item));
    }
    return list;
  }
  if (block.kind === "table") {
    return renderTable(block);
  }
  return withInline("p", block.text);
}

/** 渲染一段回答文本到 el（先清空）。流式调用时每次传入增长后的全文。 */
export function renderAnswer(el, text) {
  const source = text === undefined || text === null ? "" : String(text);
  el.textContent = "";
  try {
    for (const block of parseBlocks(source.split(/\r?\n/))) {
      el.appendChild(renderBlock(block));
    }
  } catch (err) {
    // 兜底：解析异常一律退回原文，回答不能因为格式怪就变空白
    el.textContent = source;
  }
}
