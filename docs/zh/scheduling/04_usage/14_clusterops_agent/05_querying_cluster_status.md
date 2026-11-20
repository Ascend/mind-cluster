# 查询集群与任务状态<a name="ZH-CN_TOPIC_00000026clusteropsquery01"></a>

集群运维Agent除故障诊断外，还提供故障设备、隔离卡、集群资源和任务状态四类只读查询。无需登录节点或逐个查看ConfigMap，通过`kubectl clusterops`即可获取统一格式的查询结果。

## 前提条件<a name="sectionclusteropsqueryprereq"></a>

- 已完成Agent Core、Node Collector和Kubectl Plugin的安装部署，操作请参见[手动安装](../../03_installation_guide/02_installation/01_manual_installation.md)。
- 本机已安装kubectl，且kubeconfig具备访问集群的权限；`kubectl port-forward`可用（依赖socat）。
- Agent Core运行正常，且ServiceAccount具备查询所需的ConfigMap与任务CR只读权限（`agent-core.yaml`默认已包含，权限要求参见[查询工具接口](../../06_api/21_clusterops_agent_tools.md#sectionclusteropstoolspermissions)）。

## 查看能力清单<a name="sectionclusteropsquerycatalog"></a>

```shell
kubectl clusterops tools
```

命令输出当前Agent Core注册的全部工具及入参概要。加`--json`可查看完整描述符（工具名、说明、入参schema），用于确认参数名称与类型。

## 查询故障设备<a name="sectionclusteropsqueryfaulty"></a>

```shell
kubectl clusterops tool faulty_device
kubectl clusterops tool faulty_device --param node=worker-a
```

不指定`node`参数（缺省为`null`）时查询全集群；指定时只查询该节点。表格模式输出故障设备表与节点健康表，列含义如下。

| 列 | 说明 |
|--|--|
| NODE | 节点名 |
| SOURCE | 记录来源：`devices`（设备/节点级故障）、`switch`（交换机告警）、`dpu`（DPU故障或节点事件）、`public_faults`（公共故障） |
| DEVICE | 设备名；设备摘要中节点级故障为空，节点汇总`FaultDevList`路径下为设备ID |
| TYPE | 故障类型（`fault_type`），缺失时取记录类别（`device_type`），如`CardUnhealthy`、`NodeUnhealthy`、`psu` |
| HANDLING | 故障处理等级，如`SeparateNPU`、`PreSeparateNPU` |
| CODES | 故障码，每个故障码一行 |
| LAST_TIME | 该故障码最近一次故障时间（UTC） |
| EXTRA | 关联信息：设备记录为`superpod=超节点ID/框ID/节点位置`，DPU设备记录为`affected_npu=...` |

示例（数据为示意）：

```text
NODE      SOURCE   DEVICE        TYPE             HANDLING      CODES      LAST_TIME                  EXTRA
worker-a  devices  Ascend910-3   CardUnhealthy    SeparateNPU   00776711   2026-11-09T02:11:26Z       superpod=1/0/3
NODE_HEALTH
NODE      STATUS
worker-a  UnHealthy
```

>[!NOTE]
>记录（包括`NotHandleFault`）表示“上报过故障”，不代表设备已被隔离或不可调度；隔离状态请使用`tool isolated_device`查询。
>节点因只有健康状态、或只有Switch/DPU/公共故障记录而出现时，不会补空的`devices`行。

## 查询隔离卡<a name="sectionclusteropsqueryisolated"></a>

```shell
kubectl clusterops tool isolated_device
kubectl clusterops tool isolated_device --param node=worker-a
```

表格模式逐卡输出隔离记录（不输出汇总数）：

```text
NODE      DEVICE        FAULT_LEVEL           FAULT_CODE   LAST_SEPARATE_TIME
worker-a  Ascend910-6   ManuallySeparateNPU   00776711     2026-11-09T02:11:26Z
```

| 列 | 说明 |
|--|--|
| DEVICE | 隔离设备名 |
| FAULT_LEVEL | 隔离状态（`Detail`中的`FaultLevel`原值），如`ManuallySeparateNPU`、`SilentFault`；多个明细的不同等级按英文逗号拼接；无有效明细时为`unknown` |
| FAULT_CODE | 触发隔离的故障码，多个以逗号拼接 |
| LAST_SEPARATE_TIME | 最近一次隔离/升级时间（UTC） |

>[!NOTE]
>隔离名单只来自ClusterD的`clusterd-manual-info-cm`，不读取DP的`ManuallySeparateNPU`；只有`SeparateNPU`/`PreSeparateNPU`而未进入该名单的设备不会显示。

## 查询集群资源<a name="sectionclusteropsqueryresource"></a>

```shell
kubectl clusterops tool cluster_resource
kubectl clusterops tool cluster_resource --param node=worker-a
```

表格模式按节点、按资源类型输出明细（不输出全集群汇总）：

```text
NODE      RESOURCE   CAPACITY   ALLOCATABLE   USED   IDLE   FAULT   ISOLATED
worker-a  cpu        128.0      120.0         2.0    118.0
worker-a  Ascend910  8          7             1      6      1       1
worker-b  Ascend910  8          8             0      8      0       0
```

| 列 | 说明 |
|--|--|
| NODE | 节点名 |
| RESOURCE | 资源类型：`cpu`，或整卡资源桶`Ascend910`/`npu`（同一个节点可以同时有多个桶） |
| CAPACITY | Node `status.capacity` 中该资源的值；CPU按核、NPU按卡 |
| ALLOCATABLE | Node `status.allocatable` 中该资源的值 |
| USED | 该节点上Running Pod的普通容器`requests`之和；不计initContainer、limits，也不读实际利用率 |
| IDLE | `allocatable−used`；allocatable不可知时为空，负数保留 |
| FAULT / ISOLATED | 故障卡数 / 隔离卡数；按设备身份去重、两列相互独立（同一张卡可同时计入）。证据不可知时为空，读到且没有对应身份时为`0` |

>[!NOTE]
>
>- `FAULT+ISOLATED`不是`capacity−allocatable`的拆分，工具也不做这层“闭合”校验；不要用该列做硬断言。
>- 故障/隔离来源读取失败只让对应列为空、顶层状态可能仍为`ok`（该工具的状态只看Node/Pod与资源字段完整性）；具体原因查agent-core日志。
>- 只统计带受支持整卡资源键的节点；纯CPU节点不出现。

## 查询任务状态<a name="sectionclusteropsquerytask"></a>

```shell
kubectl clusterops tool task_status --param job=train-a
kubectl clusterops tool task_status --param job=train-a --param namespace=training
```

表格模式输出任务行与Pod清单：

```text
job: train-a ns=default type=acjob status=running
pod                status
train-a-worker-0   running
train-a-worker-1   running
```

| 字段 | 说明 |
|--|--|
| job / ns / type | 任务名、命名空间、任务类型（`acjob`/`vcjob`，无法唯一确认时为`-`） |
| status | 任务状态：`running`、`completed`、`failed`、`pending`、`unknown` |
| pod / status | 当前任务实例的Pod名与归一状态（`running`、`pending`、`completed`、`failed`、`unknown`；Running但容器处于`CrashLoopBackOff`时原样显示） |

>[!NOTE]
>任务状态只由ACJob/VCJob对象自身的`status`归一，不从Pod推断；任务对象不存在、同名类型冲突或状态不可判定时显示`unknown`，信封为`partial`。

## 参数写法<a name="sectionclusteropsquerytool"></a>

`tool <name>`是唯一的业务调用入口，参数由服务端返回的`input_schema`驱动。常用参数与多节点、空范围写法如下：

```shell
kubectl clusterops tool faulty_device --param node=worker-a
kubectl clusterops tool faulty_device --param node=worker-a --param node=worker-b
kubectl clusterops tool isolated_device --param node=worker-a --param node=worker-b
kubectl clusterops tool isolated_device --params-json '{"node": []}'
kubectl clusterops tool task_status --param job=train-a --param namespace=training
kubectl clusterops tool diagnose --param job=train-a
```

- `--param NAME=VALUE`可重复，同名多次表示数组；`--params-json`直接传完整参数对象，用于空数组、嵌套对象和`null`，与`--param`互斥。
- 参数按服务端返回的`input_schema`转换；未知参数和类型错误在本地报错（退出码2），不发起请求。
- `node`取`null`（缺省）表示全集群，取`[]`表示显式空范围（不触发全集群读取）。
- 显式`node`会先校验节点是否在当前集群：全部不存在直接报错；部分不存在仍查询存在的节点，并在`errors[]`中逐节点提示。

## 理解返回结果<a name="sectionclusteropsqueryresult"></a>

所有查询返回统一结构：`request_id`、`tool`、`status`、`elapsed_ms`、`data`、`errors`（共6个字段，没有`warnings`）。

- `status=ok`：本次范围内读取完整且解析无误；`data`中的空数组、空节点对象表示“范围内确实没有记录”。
- `status=partial`：查询已执行，但结果不完整（如部分节点未被设备事实覆盖、节点缺健康状态、分片读取失败、资源字段缺失），或显式`node`中有节点不在集群内（`errors[]`逐节点给出`节点 <name> 不存在`）；已读数据保留，其他缺口原因只写服务端日志。
- `status=error`：工具没有执行完成（参数非法、工具不存在、内部异常），或显式`node`**全部**不在集群内（`errors[0].message=所有节点都找不到`）；`data`为`null`。

`--json`输出完整JSON，便于脚本处理；退出码约定如下。

| 场景 | 退出码 |
|--|--|
| 查询成功（`ok`） | 0 |
| 查询不完整（`partial`） | 3 |
| 查询失败（`error`） | 1 |
| 本地参数错误 | 2 |

常见返回情况：

| 场景 | 表现 |
|--|--|
| 显式空范围（`node=[]`） | `status=ok`，`data`中节点对象为空，不读取集群 |
| 部分节点未读到或分片缺失 | `status=partial`，stderr打印`warning: partial result (request_id=...); see agent-core logs for the uncovered targets`，已读数据保留 |
| 权限不足 | `status=partial`（数据缺失）、`errors=[]`；服务端日志记`SOURCE_FORBIDDEN` |
| 显式节点全部不存在 | `status=error`，`errors[0].code=TARGET_NOT_FOUND`、message`所有节点都找不到`，`data=null` |
| 显式节点部分不存在 | `status=partial`，`errors[]`每个缺失节点一条`节点 <name> 不存在`，存在的节点正常返回 |
| 指定任务不存在 | `task_status`为`status=partial`、`data.status=unknown`、`pods=[]`；服务端日志记`SOURCE_NOT_FOUND` |
| 参数或工具错误 | `status=error`，`errors[].code`为`INVALID_ARGUMENT`、`TOOL_NOT_FOUND`或`INTERNAL_ERROR` |

`errors[]`在工具执行失败、以及显式`node`部分不存在时非空，条目含`code`、`source_id`、`path`、`message`、`retryable`；其他业务层来源问题（`SOURCE_*`）只写服务端日志，不改变HTTP状态。

## 常见问题<a name="sectionclusteropsqueryfaq"></a>

| 现象 | 可能原因 | 处理建议 |
|--|--|--|
| `status=partial` | 部分节点未被设备事实覆盖、缺健康状态、分片读取失败或资源字段缺失 | 已读数据仍可参考；按`tool`与时间查agent-core日志（业务层日志不含`request_id`，不能直接用响应中的`request_id`检索） |
| 服务端日志出现`SOURCE_FORBIDDEN` | Agent Core的ServiceAccount缺少对应ConfigMap/Node/Pod的读权限 | 核对`agent-core.yaml`中的ClusterRole与RoleBinding，确认只读权限已授予 |
| 响应`errors[]`出现`TARGET_NOT_FOUND` | 显式`node`里有节点不在当前集群 | 全部不存在是`error`、部分不存在是`partial`；用`kubectl get nodes`核对节点名 |
| 服务端日志出现`SOURCE_NOT_FOUND` | 对应ConfigMap不存在或任务对象不存在 | 核对ClusterD/DP是否已写入对应CM；查询任务时确认`namespace` |
| 本地参数错误（退出码2） | 参数名、类型或取值不符合服务端返回的`input_schema` | 先执行`kubectl clusterops tools --json`查看参数定义；显式空范围用`--params-json '{"node": []}'` |
| `cluster_resource`的`FAULT`/`ISOLATED`为空 | 该资源桶没有可归属的故障/隔离身份，或对应来源不可读 | 属正常口径：两列独立统计，不按`capacity−allocatable`拆分；需要隔离名单时用`tool isolated_device`核对 |
| 命令提示连接失败 | 本机`kubectl port-forward`不可用（缺`socat`）或本地端口被占用 | 安装`socat`、释放被占用端口，并确认Agent Core的Pod处于`Running` |
| 提示工具不存在 | 使用了旧的连字符写法，或当前Agent Core版本不提供该工具 | 执行`kubectl clusterops tools`查看注册的工具名；工具名均为下划线形式（如`isolated_device`） |

## 参考<a name="sectionclusteropsqueryref"></a>

- 接口契约（参数、返回字段、错误码、行为边界）：[查询工具接口](../../06_api/21_clusterops_agent_tools.md)
- 诊断流程与缓存：[使用故障诊断](./03_using_fault_diagnosis.md)
