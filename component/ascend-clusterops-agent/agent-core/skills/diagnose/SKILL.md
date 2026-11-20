---
name: diagnose
description: 针对某个任务执行确定性故障诊断链（采集日志、清洗、诊断），返回诊断报告、结论正文与证据缺口。当用户问"帮我诊断这个任务""任务为什么失败""根因是什么"时使用。
allowed-tools: [diagnose]
---

# 任务诊断

## 什么时候用

- 用户明确要求"诊断""根因""失败原因"，且提供了任务名（可带命名空间）。
- 这是耗时最长的一步（可能要几分钟）：先确认任务名再调用；**同一任务同时只允许一次诊断**，并发重复请求会被拒并提示"正在诊断中"，此时如实转述、不要立刻重试。
- 只看状态不诊断 → `task-status`；只看设备故障 → `faulty-device`。

## 调用工具与参数

工具名 `diagnose`（下划线）：

| 参数 | 类型 | 规则 |
|---|---|---|
| `job` | 字符串，**必填** | 任务名 |
| `namespace` | 字符串，可选 | 默认 `default` |
| `refresh` | 布尔，可选 | 默认 `false`：命中缓存就复用（`data.cached=true`）；`true` 表示强制重新采集 |

```json
{"job": "train-xxx"}
```

摘要与展示由工具层完成，没有额外开关：配了模型就出模型摘要，没配就给确定性报告正文。

## 返回字段与解读

**返回是工具层原始信封 JSON**，不经过任何摘要或改写：`request_id`、`tool`、`status`、`elapsed_ms`、`data`、`errors[]`。下面的字段路径就是 JSON 里的实际路径，直接用路径引用，不要把整段 JSON 贴给用户。返回超过长度上限时末尾会带一句「已截断」声明：被截断的部分没有输出，不代表为空或不存在；诊断报告很长时最容易截断，若 `final_text` 或 `diag_report` 被截断，先说结论，再说明"报告正文被截断，完整内容请用 kubectl 查看"，不要补写。

`data` 是旧 `/diag` 入口的扁平业务字段，与 CLI 看到的完全一致：

- `data.final_text`：**结论正文**（模型摘要，或没有模型时的确定性报告文本），对话链路的主要结论来源
- `data.diag_report` / `data.diag_report_text`：确定性诊断报告的分节结构与纯文本
- `data.incomplete_nodes[]` / `data.empty_nodes[]`：**证据缺口**——未完成采集的节点、采集到但无数据的节点
- `data.pods[]`：诊断涉及的 Pod 记录
- `data.cached`：是否命中历史缓存（`true` 表示可能包含既往产出，不代表本次实时结论）
- `data.error`：诊断链自身错误，可能为空
- `data.llm_error`：**只在模型摘要失败时出现**（此时 `final_text` 已回退为确定性报告正文）；没有这个键就是摘要正常或未配置模型，不要因为缺少它就说"诊断没有结论"

解读要点：

- 结论优先来自 `final_text`，它缺失或为空时用 `diag_report` 的分节自行归纳；有缺口时必须带上"哪些节点证据缺失"，不能说成"已完整诊断"。
- 诊断的重复执行冲突、超时、内部异常**都归为 `INTERNAL_ERROR`**：按 `errors[].code` 说明原因，不要自动重试。
- `cached=true` 时提示"来自历史缓存，可能含既往结论"，需要实时结论就用 `refresh=true` 再查一次。
- 需要看任务状态或 Pod 分布时改用 `task_status` 查询，不要在本工具的返回里找 Pod 状态明细。

## 无数据与异常提示

- 外层 `status` 是**查询质量**而非任务成败：`ok` 说明本次诊断完整；`partial` 说明只诊断了一部分——先给已观察到的结论，再说缺口，**节点级缺口明细就在 `data.incomplete_nodes[]` / `empty_nodes[]` 里**（照此给出缺失节点清单，不得说成"已完整诊断"；`errors[]` 只承载执行失败、不含缺口，`request_id` 仅用于执行失败的日志追溯）；`error` 说明执行失败（`data` 可仍带已采集内容，按 `errors[].code` 说明原因，不要回答"诊断完成/没有问题"）。
- `errors[]` 只承载执行失败，诊断类失败码只有 `INTERNAL_ERROR`（`INVALID_ARGUMENT` / `TOOL_NOT_FOUND` 同其他工具）；条目含 `code` 与 `message`，`retryable` 恒为 `false`，不要据此自动重试。
- `status=ok` 即本次诊断完整（缺口字段此时必为空）；若 `ok` 而 `incomplete_nodes` / `empty_nodes` 非空，以缺口字段为准声明覆盖范围、不得说"已完整覆盖"。
- 本技能有副作用（采集日志、写报告与缓存），但不对集群做修复操作；需要动作时引导用户走命令行或变更流程。

## 回答格式（三段式，缺一不可）

1. **总结**：1~3 句直接回答，先给结论（根因是什么、是否诊断完整）。
2. **依据**：逐条列出结论出处，写清工具名与字段路径（如 `diagnose` → `data.final_text`、`data.diag_report` 的某分节、`data.incomplete_nodes[]`），结论必须能在返回里找到出处。
3. **说明与建议**：先交代本次缺口与限制（`partial`、截断、`incomplete_nodes`/`empty_nodes`、`cached`、`llm_error`），再给下一步建议（重跑用 `refresh=true`、看状态用 `task_status`、需要动作走变更流程）。

三段都要写，也不要把整段信封 JSON 贴给用户。

## 禁止事项

- 不编造返回里没有的字段、节点名或结论；缺口未说明不能说"已完整诊断"。
- 不把技能名当工具名：工具名必须是 `diagnose`（与技能名相同）。
- 不跨技能臆测：设备故障查 `faulty_device`，隔离名单查 `isolated_device`，资源查 `cluster_resource`，任务状态查 `task_status`，任务日志诊断查 `diagnose`。

契约依据：`agent-core/tools/builtin.py`（工具注册）、`query_schemas.py`（入参与返回模型，与其余四份技能同一处）。
