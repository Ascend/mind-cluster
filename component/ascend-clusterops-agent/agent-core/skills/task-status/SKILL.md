---
name: task-status
description: 查询某个训练/推理任务的权威状态与逐 Pod 状态：任务类型（可为 null）、归一化任务状态、恒为数组的 Pod 列表。当用户问"任务起了吗""任务卡在哪""这个任务的 Pod 什么状态"时使用。
allowed-tools: [task_status]
---

# 任务状态查询

## 什么时候用

- 追问某一个具体任务的权威状态与 Pod 明细：任务类型、归一化任务状态、每个关联 Pod 的状态。
- 典型问法："train-xxx 这个任务起来了吗？""任务卡在哪个环节？""worker Pod 现在什么状态？"
- 需要看日志或做故障定位时用 `diagnose`；需要看节点还剩多少卡用 `cluster-resource`。

## 调用工具与参数

工具名 `task_status`（下划线）：

| 参数 | 类型 | 规则 |
|---|---|---|
| `job` | 字符串，**必填** | 任务名（不需要带命名空间前缀） |
| `namespace` | 字符串，可选 | 任务所在命名空间，默认 `default` |

```json
{"job": "train-xxx"}
```

```json
{"job": "train-xxx", "namespace": "train-ns"}
```

- 按名称自动识别 ACJob / VCJob 两类任务对象，**没有类型入参**；任务名必须来自用户输入或上一轮返回，**不要猜**，用户没给任务名时先问。
- 用户没说命名空间时可用 `default`；如果返回空态或缺口，要主动提示"任务可能在别的命名空间，请确认"。

## 返回字段与解读

**返回是工具层原始信封 JSON**，不经过任何摘要或改写：`request_id`、`tool`、`status`、`elapsed_ms`、`data`、`errors[]`。下面的字段路径就是 JSON 里的实际路径，直接用路径引用，不要把整段 JSON 贴给用户。返回超过长度上限时末尾会带一句「已截断」声明：被截断的部分没有输出，不代表为空或不存在，需要完整数据请缩小查询范围或改用 kubectl 查询，并把这一点写进结论。

`data` 是任务级对象（无节点分组）：

- `data.job_name` / `data.namespace`：任务名与命名空间回显
- `data.job_type`：任务类型归一值（`acjob` / `vcjob`），**可能为 null**
- `data.status`：**任务状态**，由任务对象状态归一，取值 `pending` / `running` / `completed` / `failed` / `unknown`
- `data.pods[]`：关联 Pod 列表，每项 `pod_name`、`status`，按名称排序

解读要点：

- **不要把外层 `status` 当任务状态**：外层是查询质量（`ok`/`partial`/`error`），任务状态只看 `data.status`；任务失败（`failed`）不代表查询失败。
- **`data.status=unknown` 的来源**：① ACJob 与 VCJob 同名双命中（工具不猜类型）；② 两类对象都不存在；③ 对象读取失败无法确认（权限/超时）。三种来源在返回中都表现为 `job_type` 为 null、`data.pods` 为空数组，且**外层 `status` 恒为 `partial`（不会是 `ok`）**——「查询正常完成但状态未知」的组合不存在。回答必须说「无法确认任务类型与状态」并列出可能原因，不得说成「任务不存在」或「没有关联 Pod」（①②③的区分信息在服务端日志，不要臆断其中一种）。对象存在但状态无法判定时同样返回 `unknown`（此时 `job_type` 有值、Pod 列表可非空，外层同样至少为 `partial`）。
- **`data.pods` 恒为数组**：目标不存在、类型冲突或归属无法核验时是 `[]`，字段不会缺失，可直接读取。
- **`CrashLoopBackOff` 只出现在 `pods[].status`，是 Pod 级状态，不是任务状态**：任务状态只随任务对象变化，不要用 Pod 状态反推任务状态，也不要把任务本身说成"CrashLoopBackOff"。
- 旧口径的 ranktable 状态、rank/server 计数、server 与设备分布、调度原因列表都不在返回中；用户追问时如实说明本工具不提供，不要编造。

## 无数据与异常提示

- 外层 `status` 是**查询质量**而非任务成败：`ok` 说明本次结果完整；`partial` 说明只查到一部分——先给已观察到的结果，再说缺口，**缺口明细不在返回中**（`errors[]` 仍为空，只能凭 `request_id` 查服务端日志）；`error` 说明执行失败（`data` 为 null，按 `errors[].code` 说明原因，不要回答"任务不存在/一切正常"）。
- `errors[]` 只承载执行失败（`INVALID_ARGUMENT` / `TOOL_NOT_FOUND` / `INTERNAL_ERROR`），条目含 `code` 与 `message`，`retryable` 恒为 `false`，不要据此自动重试。
- 仅 `status=ok` 时空 Pod 列表才可读作"当前没有关联 Pod"。
- 只读技能：不执行重启、删除或其他变更；需要动作时引导用户走命令行或变更流程。

## 回答格式（三段式，缺一不可）

1. **总结**：1~3 句直接回答，先给结论（任务是 running/failed/无法确认）。
2. **依据**：逐条列出结论出处，写清工具名与字段路径（如 `task_status` → `data.status`、`data.pods[].status`），数值与状态必须与返回一致。
3. **说明与建议**：先交代本次缺口与限制（`unknown` 的三种来源、`partial`、截断、命名空间可能不对），再给下一步建议（要查失败原因用 `diagnose`、要看节点资源用 `cluster_resource`）。

三段都要写，也不要把整段信封 JSON 贴给用户。

## 禁止事项

- 不编造返回里没有的字段、Pod 名或计数；空 Pod 列表不等于任务健康。
- 不把技能名当工具名：工具名必须是 `task_status`（本技能名是 `task-status`）。
- 不跨技能臆测：设备故障查 `faulty_device`，隔离名单查 `isolated_device`，资源查 `cluster_resource`，任务状态查 `task_status`，任务日志诊断查 `diagnose`。

契约依据：`agent-core/tools/builtin.py`（工具注册）、`query_schemas.py`（入参与返回模型）。
