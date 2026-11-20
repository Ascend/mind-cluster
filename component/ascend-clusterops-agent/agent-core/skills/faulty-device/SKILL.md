---
name: faulty-device
description: 按节点查询故障设备、交换机告警、DPU 故障、公共故障与节点健康，解释故障类型、逐故障码的处置建议与时间。当用户问"哪些卡有故障""某节点有没有故障""这个故障码什么意思"时使用。
allowed-tools: [faulty_device]
---

# 故障设备查询

## 什么时候用

- 全集群或指定节点范围内列出现存故障：设备/节点级故障、交换机告警、DPU 故障、公共故障与节点健康，各类独立呈现。
- 典型问法："集群里哪些卡有故障？""node-1 上有什么故障？""81078603 这个码是什么意思？"
- 追问某个明确节点时把节点名放进 `node` 数组；对象不明确（既没说节点也没说集群）时先澄清，不要自己挑一个节点。
- 只读查询。任务日志诊断用 `diagnose`；隔离名单用 `isolated-device`；容量与空闲卡用 `cluster-resource`。

## 调用工具与参数

工具名 `faulty_device`（下划线）。参数只有 `node` 一个：

| 参数 | 类型 | 规则 |
|---|---|---|
| `node` | 字符串数组，可选 | 省略或 null = 不限定节点（全量读取）；传数组 = 只看这些节点；传空数组 `[]` = 显式空范围（不读来源，直接返回空 `data.node`）。一律传数组（单个字符串也会被服务端归一为单元素数组，公网 CLI 单值形态由此兼容） |

全集群查询：

```json
{}
```

指定节点查询（node-1、node-2 换成真实节点名）：

```json
{"node": ["node-1", "node-2"]}
```

## 返回字段与解读

**返回是工具层原始信封 JSON**，不经过任何摘要或改写：`request_id`、`tool`、`status`、`elapsed_ms`、`data`、`errors[]`。下面的字段路径就是 JSON 里的实际路径，直接用路径引用，不要把整段 JSON 贴给用户。返回超过长度上限时末尾会带一句「已截断」声明：被截断的部分没有输出，不代表为空或不存在，需要完整数据请缩小 `node` 范围或改用 kubectl 查询，并把这一点写进结论。

数据入口是 `data.node{}`：按节点名索引的字典（`{}` 是节点名占位），节点按名称排序；空范围或没有记录时 `data.node` 为空对象。

每个节点对象内五类内容**并列**：

- `data.node{}.devices[]`：设备与节点级故障。**一条记录 = 一个设备 + 一个 `fault_type`**：同一设备的多种故障类型是多条记录，报数说"N 条记录"而不是"N 台坏卡"。字段：`device_id`（**节点级故障为空字符串 `""`**，不能按非空过滤）、`device_type`（如 `npu`/`psu`）、`chip_type`（资源前缀，如 `Ascend910`，节点级为 null）、`fault_type`、`faults[]`、`superpod`。
- `data.node{}.devices[].faults[]`：**逐故障码明细**，每条 `fault_code`（单码，不拼接）、`fault_handling`（该码的处置建议，如 `PreSeparateNPU`/`NotHandleFault`，是建议、不代表已隔离）、`last_time`（该码最近故障时间，可能为 null）。源条目含多个码时展开成多条明细。
- `data.node{}.switch[]` / `data.node{}.dpu[]` / `data.node{}.public_faults[]`：交换机活动告警、DPU 故障与节点事件、公共故障，三个**独立列表**，与 `devices[]` 并列、互不并入，报数分开。每条含 `device_id`、`fault_handling`、`fault_code`、`last_time`；`switch[]`/`dpu[]` **没有 `fault_type` 字段**（源无故障类型维度，不要臆造或说"类型未给出"），`public_faults[]` 另含 `fault_type`（可为 null）；DPU 记录可带 `affected_npu`。
- `data.node{}.public_faults[].device_id` 是**芯片物理 ID**（如 `5`），与 `devices[].device_id`（设备名，如 `Ascend910-0`）命名不同，不能直接等值关联，需自行映射。
- `data.node{}.node_health`：节点健康**对象** `{node_status}`（如 `Healthy`），只在可读时出现，**键可能整个缺失**；谈节点健康只能引用这里，不从故障记录反推，也不把 `Healthy` 说成异常。
- `data.node{}.devices[].superpod`：超节点拓扑，可能为 null；**非 null 不代表拓扑完整**——`super_pod_id`、`rack_id`、`server_index` 要逐子键判空（子键 null 表示未知；`0` 是合法值，不是未知）。

字段出现规则：`devices[]` 只在有故障记录时出现，只有交换机/DPU/公共故障记录的节点不补 `devices[]`；`node_health` 只在节点健康可读时出现。

解读要点：节点数、`devices[]` 记录数、`faults[]` 明细数分开说，不做跨列表相加；`fault_handling` 是建议的处置方式，判断"是否已被隔离"要用 `isolated-device` 的名单。

## 无数据与异常提示

- 外层 `status` 是**查询质量**而非集群健康：`ok` 说明本次结果完整；`partial` 说明只查到一部分——先给已观察到的结果，再说缺口。**缺口分两层**：已列出节点的 `node_health` 键缺失、明细字段为 null 即该来源不可读（按各自规则声明）；点名了却未出现在 `data.node` 中的节点 = 未能读取或不存在（返回中不可见，凭 `request_id` 查服务端日志）；`error` 说明执行失败（`data` 为 null，按 `errors[].code` 说明原因，不要回答"没有故障/一切正常"）。
- `errors[]` 只承载执行失败（`INVALID_ARGUMENT` / `TOOL_NOT_FOUND` / `INTERNAL_ERROR`），条目含 `code` 与 `message`，`retryable` 恒为 `false`，不要据此自动重试。
- 仅 `status=ok` 时**记录列表**为空/缺失（`devices[]` / `switch[]` / `dpu[]` / `public_faults[]`）才可读作"该范围内无记录"；`node_health` 缺键不是"无记录"，是健康来源不可读（见上）。显式空范围 `node: []` 产生的空返回应表述为"未查询任何节点"，不在此列；`partial` 下不能读作"没有故障"。
- 只读技能：不执行隔离、解隔离、重启或其他修复；需要动作时引导用户走命令行或变更流程。

## 回答格式（三段式，缺一不可）

1. **总结**：1~3 句直接回答，先给结论（有无故障、涉及几个节点、几类记录）。
2. **依据**：逐条列出结论出处，写清工具名与字段路径（如 `faulty_device` → `data.node{}.devices[].faults[]` 里的 `fault_code`），数值必须与返回一致，找不到出处的结论不写。
3. **说明与建议**：先交代本次缺口与限制（`partial`、截断、`node_health` 缺键、null 字段、点名了却没返回的节点），再给下一步建议（该用哪个工具、该走哪个流程）。

三段都要写，也不要把整段信封 JSON 贴给用户。

## 禁止事项

- 不编造返回里没有的字段、设备名、故障码或计数；空结果不等于设备或集群健康。
- 不把技能名当工具名：工具名必须是 `faulty_device`（本技能名是 `faulty-device`）。
- 不跨技能臆测：设备故障查 `faulty_device`，隔离名单查 `isolated_device`，资源查 `cluster_resource`，任务状态查 `task_status`，任务日志诊断查 `diagnose`。

契约依据：`agent-core/tools/builtin.py`（工具注册）、`query_schemas.py`（入参与返回模型）。
