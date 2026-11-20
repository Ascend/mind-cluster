# 查询工具接口<a name="ZH-CN_TOPIC_00000026clusteropstools01"></a>

集群运维Agent在诊断能力之外，提供4个只读查询工具（`faulty_device`、`isolated_device`、`cluster_resource`、`task_status`）和1个诊断工具（`diagnose`），并以统一契约对外暴露：同一份工具定义同时供HTTP接口与Kubectl Plugin使用，外部智能体经这两个入口复用同一份定义，新增工具无需改动客户端。

本文描述上述接口的发现、执行、参数、返回体、错误码与行为边界。诊断接口的既有说明（`/diag`、诊断结果缓存、gRPC、Kubectl Plugin安装）请参见[Agent Core、Node Collector与Kubectl Plugin](20_clusterops_agent.md)。

>[!NOTE]
>9700为集群内部端口，不对外暴露；集群外请通过Kubectl Plugin（`kubectl clusterops`，经`kubectl port-forward`访问）调用。

## 工具清单<a name="sectionclusteropstoolstools"></a>

**表1**  工具清单

|工具名|读写|用途|入参|
|--|--|--|--|
|`faulty_device`|只读|查询NPU故障设备：以ClusterD设备摘要（`cluster-info-device-{m}`）为主源，节点未被覆盖时回退到该节点的DP ConfigMap；同时汇总节点健康、Switch/DPU/公共故障|`node`|
|`isolated_device`|只读|查询隔离卡：只读ClusterD隔离名单（`clusterd-manual-info-cm`）的`Total`，`Detail`补充故障等级、故障码和最近隔离时间|`node`|
|`cluster_resource`|只读|按节点、按芯片类型统计CPU与整卡的容量、可分配、Running请求占用、空闲，并独立统计故障/隔离卡数|`node`|
|`task_status`|只读|查询任务状态：以AscendJob/Volcano Job对象自身的`status`为唯一状态源，附当前任务实例的实时Pod|`job`、`namespace`|
|`diagnose`|非只读|对任务执行采集、清洗与诊断，返回诊断结果；HTTP/CLI入口有模型时自动总结，无模型时返回确定性报告文本；Registry/Agent直接调用返回原始诊断data|`job`、`namespace`、`refresh`|

## 权限要求<a name="sectionclusteropstoolspermissions"></a>

查询工具在集群内直接读取任务对象与ConfigMap，依赖以下RBAC规则（`agent-core.yaml`中已包含，部署时无需手工创建）：

- **ClusterRole `agent-core`（集群级）**：`configmaps`的`get`；`ascendjobs`（`mindxdl.gitee.com`）和`jobs`（`batch.volcano.sh`）的`get`。读取任务对象取得`status`，并读取任务所在命名空间的ConfigMap：任务命名空间不固定、对象名称动态，无法用命名空间或`resourceNames`收敛。
- **Role `agent-core-query-read`（`mindx-dl`、`kube-system`、`cluster-system`）**：`configmaps`的`get/list`。设备摘要分片（`cluster-info-device-{m}`）、交换机与DPU故障分片、kube-system下按节点命名的DP设备信息回退源等，需按名称前缀或标签发现，分片序号与节点名动态，因此需要`list`；`list`仅在上述固定命名空间内授予。

>[!NOTE]
>若集群使用自定义RBAC，请保留上述权限，否则相关工具会因权限不足返回`status=partial`（服务端日志记`SOURCE_FORBIDDEN`），响应中没有对应错误码。

## 能力发现（GET /tools）<a name="sectionclusteropstoolscatalog"></a>

**功能说明<a name="sectionclusteropstoolscatalogdesc"></a>**

返回当前Agent Core注册的全部工具及其入参schema。客户端按名称选择工具，并按`input_schema`转换参数，因此新增工具无需升级客户端。

**接口说明<a name="sectionclusteropstoolscatalogapi"></a>**

**表2**  能力发现接口

|项目|说明|
|--|--|
|路径|`/tools`|
|方法|GET|
|默认端口|9700|
|响应|`request_id`、`tools[]`|

**响应说明<a name="sectionclusteropstoolscatalogresp"></a>**

**表3**  响应字段

|字段|类型|说明|
|--|--|--|
|request_id|string|服务端生成的本次请求标识。|
|tools|array|工具描述符列表，按工具名排序。|
|tools[].name|string|工具名，用于`POST /tools/{name}`。|
|tools[].description|string|工具用途说明。|
|tools[].input_schema|object|入参JSON Schema（与请求模型同源，不随服务端校验漂移）。|

**响应示例<a name="sectionclusteropstoolscatalogexample"></a>**

```json
{"request_id": "9d63c02f90e74e05bd2f5ebd81498819",
 "tools": [{"name": "cluster_resource",
            "description": "Report per-node CPU cores and per-chip-type card accounting (capacity, allocatable, used, idle) with independent fault/isolated counts.",
            "input_schema": {"additionalProperties": false,
                             "properties": {"node": {"anyOf": [
                               {"items": {"maxLength": 253, "minLength": 1,
                                          "pattern": "^[a-z0-9](?:[-a-z0-9]*[a-z0-9])?(?:\\.[a-z0-9](?:[-a-z0-9]*[a-z0-9])?)*$",
                                          "type": "string"}, "type": "array"},
                               {"type": "null"}], "default": null, "title": "Node"}},
                             "title": "NodeQuery", "type": "object"}}]}
```

>[!NOTE]
>仅支持`GET /tools`与`POST /tools/{name}`；`GET /tools/{name}`已删除，访问返回405。

## 统一执行（POST /tools/{name}）<a name="sectionclusteropstoolsexec"></a>

**功能说明<a name="sectionclusteropstoolsexecdesc"></a>**

执行一个已注册工具。请求体直接是工具参数对象（不套`arguments`字段），服务端按工具的入参模型校验后执行，并返回统一工具信封。

**接口说明<a name="sectionclusteropstoolsexecapi"></a>**

**表4**  执行接口

|项目|说明|
|--|--|
|路径|`/tools/{name}`|
|方法|POST|
|默认端口|9700|
|请求体|工具参数对象，见[请求参数](#sectionclusteropstoolsparams)|
|响应|工具执行信封，见[返回体](#sectionclusteropstoolsenvelope)|
|未知工具|404，`TOOL_NOT_FOUND`|
|参数校验失败|422，`INVALID_ARGUMENT`（`path`给出字段路径）|

## 请求参数<a name="sectionclusteropstoolsparams"></a>

**表5**  公共参数

|参数|类型|适用工具|默认/约束|
|--|--|--|--|
|node|string[]、string或null|除`task_status`、`diagnose`外的三类查询|null表示ServiceAccount可见的全量节点；`[]`表示显式空范围（不读全集群）；单个字符串等价于单元素数组；每项为DNS-1123子域名（1～253字符）；去重保序；上限20000项。|
|job|string|`task_status`、`diagnose`|必填；任务名（任务CR名称），DNS-1123子域名（1～253字符）。|
|namespace|string|`task_status`、`diagnose`|默认`default`；DNS-1123标签（1～63字符），不允许点。|
|refresh|boolean|仅`diagnose`|默认false；忽略缓存强制重新诊断。查询工具传入按未知参数拒绝。|

**表6**  各工具适用参数

|工具名|node|job|namespace|refresh|
|--|--|--|--|--|
|`faulty_device`|是|否|否|否|
|`isolated_device`|是|否|否|否|
|`cluster_resource`|是|否|否|否|
|`task_status`|否|是|是|否|
|`diagnose`|否|是|是|是|

>[!NOTE]
>
>- 查询工具不接受`refresh`：只读查询直接读取当前数据，没有缓存刷新语义。
>- 未知参数、类型错误和缺失必填参数统一返回`INVALID_ARGUMENT`，不执行工具。

## 返回体（工具信封）<a name="sectionclusteropstoolsenvelope"></a>

所有工具（含`diagnose`）返回同一信封，共6个字段。

**表7**  信封字段

|字段|类型|说明|
|--|--|--|
|request_id|string|服务端生成的本次请求标识；业务层日志当前不带该字段，不能直接用它检索日志。|
|tool|string|工具名。|
|status|string|`ok`、`partial`或`error`：表示执行与完整性，不表示设备健康或任务成败。|
|elapsed_ms|int|服务端耗时（毫秒）。|
|data|object/null|业务数据：4个查询工具见[查询结果data](#sectionclusteropstoolsdata)，`diagnose`见[诊断工具](#sectionclusteropstoolsdiagnose)；查询工具`status=error`时为null，`diagnose`的业务错误仍保留扁平业务字段（参数/工具错误时为null）。|
|errors|Issue[]|工具执行失败，或显式`node`中有节点不存在（此时`status=partial`）时非空；其他业务层查询问题不写入，见[错误码与状态映射](#sectionclusteropstoolserrors)。|

**表8**  Issue字段<a name="tableclusteropsissue"></a>

|字段|类型|说明|
|--|--|--|
|code|string|执行失败码，见[错误码与状态映射](#sectionclusteropstoolserrors)。|
|source_id|string/null|执行失败涉及来源时保留读取网关标识；HTTP边界错误为null。|
|path|string/null|请求参数或执行路径（如`node.0`）；没有具体路径时为null。|
|message|string|可读说明。|
|retryable|boolean|是否可重试，当前固定为false。|

**status取值说明<a name="sectionclusteropstoolsenvelopestatus"></a>**

- `ok`：本次范围内读取完整且解析无误；空业务数组/空节点对象表示“范围内确实没有记录”。
- `partial`：查询已执行，但有可用数据不完整（如部分节点未被设备事实覆盖、节点缺少健康状态、分片读取失败、资源字段缺失等），或显式`node`中有节点不在集群内（`errors[]`逐节点给出`TARGET_NOT_FOUND`、`message`为`节点 <name> 不存在`）。`data`保留已读部分；其他缺口原因只记录在服务端日志。
- `error`：工具没有执行完成（参数非法、工具不存在、未预期异常），或显式`node`**全部**不在集群内（`TARGET_NOT_FOUND`，message`所有节点都找不到`）；`errors`说明原因。查询工具此时`data=null`；`diagnose`的业务错误（任务不存在、并发、超时、失败）仍返回扁平`data`（含`error`/`final_text`）。

>[!NOTE]
>读取返回体的固定顺序：先看`status`，再看`errors`，最后看`data`。只有`status=ok`时，空业务数组才可解释为“本次完整范围内没有记录”；`partial`时`errors`可能给出不存在的节点，其余缺口原因需查服务端日志（业务层日志不带`request_id`，只能按`tool`与时间定位）。

## 查询结果data<a name="sectionclusteropstoolsdata"></a>

三类节点查询的节点级数据统一挂在`data.node`下，键为节点名、按节点名排序；`task_status`按任务聚合。节点对象只包含该工具实际存在的属性。

### faulty_device<a name="sectionclusteropstoolsfaulty"></a>

**表9**  faulty_device的data字段

|字段|类型|说明|
|--|--|--|
|node|object|以节点名为键的节点容器，见下表。|

`data.node.<节点名>`的字段：

|字段|类型|说明|
|--|--|--|
|devices|array|设备/节点级故障记录，见[表10](#tableclusteropsfaultrecord)。节点有设备事实（含已读但名册为空，此时为`[]`）或有节点级故障记录时出现；节点完全没有设备事实时省略该字段。|
|switch|array|交换机活动告警，字段见[表11](#tableclusteropsswitchrecord)；无记录时为`[]`。|
|dpu|array|DPU设备故障与DPU节点事件，字段见[表12](#tableclusteropsdpurecord)；无记录时为`[]`。|
|public_faults|array|公共故障，字段见[表13](#tableclusteropspublicfaultrecord)；无记录时为`[]`。|
|node_health|object|节点健康，字段见[表14](#tableclusteropsnodehealth)；节点健康状态不可读时省略。|

**表10**  devices[]字段<a name="tableclusteropsfaultrecord"></a>

|字段|类型|说明|
|--|--|--|
|device_id|string|设备名（如`Ascend910-3`）；设备摘要中`npu_name`为空的节点级记录为空字符串；节点汇总`FaultDevList`路径下为`DeviceId`原值（可能非空）。|
|device_type|string/null|芯片级记录固定为`npu`；节点级记录为来源类别串的小写形式（如`node`、`nodeunhealthy`、`psu`、`cpu`）；类别不可判定时为null。|
|chip_type|string/null|设备所属整卡资源类型：`Ascend910`或`npu`；节点级记录为null。|
|fault_type|string/null|故障类型原值，如`CardUnhealthy`、`CardNetworkUnhealthy`、`NodeUnhealthy`、`PublicFault`；节点汇总路径为`DeviceType`原值。|
|faults|array|按故障码展开的明细，见下；来源确有故障但无可用故障码时给出一条`fault_code=null`的明细；仅健康异常名单（`-Unhealthy`等）产生的记录为`[]`。|
|superpod|object/null|超节点拓扑`{super_pod_id,rack_id,server_index}`；无拓扑时为null，子字段可能为null。|

`faults[]`字段：

|字段|类型|说明|
|--|--|--|
|fault_code|string/null|故障码。|
|fault_handling|string/null|故障处理等级（如`SeparateNPU`、`PreSeparateNPU`）；优先取故障码级等级，缺失时取条目级等级。|
|last_time|string/null|该故障码最近一次故障时间（UTC，RFC3339）；来源未提供可识别时间时为null。|

**表11**  switch[]字段<a name="tableclusteropsswitchrecord"></a>

|字段|类型|说明|
|--|--|--|
|device_id|string|交换机告警标识，如`switch:0:3`。|
|fault_handling|string/null|故障处理等级，取交换机来源的等级（如`PreSeparateNPU`、`NotHandle`）。|
|fault_code|string/null|故障码。|
|last_time|string/null|最近一次告警时间（UTC）。|

>[!NOTE]
>Switch来源没有故障类型维度，该记录不输出`fault_type`字段（字段不存在，不是null）。

**表12**  dpu[]字段<a name="tableclusteropsdpurecord"></a>

|字段|类型|说明|
|--|--|--|
|device_id|string|DPU设备标识，或DPU节点事件的标识。|
|fault_handling|string/null|故障等级。|
|fault_code|string/null|故障码。|
|last_time|string/null|最近一次故障时间（UTC）。|
|affected_npu|array|仅DPU设备故障携带：受该DPU故障影响的NPU设备ID列表；DPU节点事件不输出该字段。|

**表13**  public_faults[]字段<a name="tableclusteropspublicfaultrecord"></a>

|字段|类型|说明|
|--|--|--|
|device_id|string|生产者原值（芯片物理ID，如`5`），与`devices[]`的`npu_name`命名不同。|
|fault_type|string/null|公共故障类型；`public`/`publicfault`/`public_fault`归一为`PublicFault`，`NPU`/`Node`/`Network`/`Storage`等原样保留。|
|fault_handling|string/null|故障级别。|
|fault_code|string/null|故障码。|
|last_time|string/null|故障产生时间（UTC）。|

**表14**  node_health字段<a name="tableclusteropsnodehealth"></a>

|字段|类型|说明|
|--|--|--|
|node_status|string|节点健康状态，如`Healthy`、`UnHealthy`；取自`cluster-info-node-cm`的`NodeStatus`。|

### isolated_device<a name="sectionclusteropstoolsisolated"></a>

**表15**  isolated_device的data字段

|字段|类型|说明|
|--|--|--|
|node|object|以节点名为键的节点容器；`data.node.<节点名>.devices[]`为隔离设备记录，见[表16](#tableclusteropsisolationcard)。隔离CM中没有该节点的key时不出现在`data.node`；key存在但`Total`为空时出现且`devices`为`[]`。|

**表16**  devices[]字段<a name="tableclusteropsisolationcard"></a>

|字段|类型|说明|
|--|--|--|
|device_id|string|设备名，如`Ascend910-6`。|
|fault_level|string|隔离状态，取`Detail`中的`FaultLevel`原值（如`ManuallySeparateNPU`、`SilentFault`）；多个`Detail`条目的不同等级按英文逗号拼接；`Total`中有成员但无有效`Detail`时为`unknown`。|
|fault_code|string/null|触发隔离的故障码，多个以英文逗号拼接；无明细时为null。|
|last_separate_time|string/null|最近一次隔离/升级时间（UTC）；无明细时为null。|

>[!NOTE]
>隔离记录不输出`chip_type`、`summary`等字段；设备属于哪种芯片资源，请在`cluster_resource`的`data.node.<节点名>.<资源桶>`中查看。只有`SeparateNPU`/`PreSeparateNPU`而未进入ClusterD名单的设备不产生隔离记录。

### cluster_resource<a name="sectionclusteropstoolsresource"></a>

**表17**  cluster_resource的data字段

|字段|类型|说明|
|--|--|--|
|node|object|以节点名为键的节点容器。只包含Node上带有受支持整卡资源键（`huawei.com/Ascend910`、`huawei.com/npu`）的节点；纯CPU节点不出现。|

`data.node.<节点名>`是与`cpu`并列的对象：

|字段|类型|说明|
|--|--|--|
|cpu|object|CPU资源块，见[表18](#tableclusteropscpu)。|
|Ascend910 / npu|object|每个存在的整卡资源键各建一个桶，见[表19](#tableclusteropsnoderesource)；两个键都存在时两个桶同时输出，不合并、不换算。|

**表18**  cpu字段<a name="tableclusteropscpu"></a>

|字段|类型|说明|
|--|--|--|
|capacity_cores|float/null|Node `status.capacity[cpu]`，按核输出。|
|allocatable_cores|float/null|Node `status.allocatable[cpu]`，按核输出。|
|used_cores|float|该节点上Running Pod的普通容器`requests[cpu]`之和，按核输出；无请求时为0。|
|idle_cores|float/null|`allocatable_cores − used_cores`；allocatable不可知时为null，负数保留并记日志。|

**表19**  整卡资源桶字段<a name="tableclusteropsnoderesource"></a>

|字段|类型|说明|
|--|--|--|
|capacity|int/null|Node `status.capacity[该资源键]`。|
|allocatable|int/null|Node `status.allocatable[该资源键]`。|
|used|int|该节点上Running Pod的普通容器`requests[该资源键]`之和；无请求时为0。|
|idle|int/null|`allocatable − used`；allocatable不可知时为null，负数保留并记日志。|
|fault|int/null|该桶的故障卡数：故障记录中`device_type=npu`、`device_id`非空且`chip_type`等于该桶的设备身份去重计数。证据不可知时为null，读到且无对应身份时为0。|
|isolated|int/null|该桶的隔离卡数：ClusterD隔离名单成员按名册（基础可分配 + `-Unhealthy`/`-NetworkUnhealthy`/`-Recovering` + `-Fault`名称并集）归属到该桶后去重计数。隔离来源不可读或成员无法归属时为null，读到且无成员时为0。|

>[!NOTE]
>
>- 不输出跨节点/全集群汇总，也没有`capacity−allocatable`与故障/隔离数量的“闭合”校验。
>- `fault`与`isolated`相互独立：同一张卡既故障又隔离时两个字段都计数，`fault`不扣减`isolated`。
>- `used`只计Running Pod的普通容器`requests`；不计`initContainers`、`limits`、`overhead`，也不读取设备实际利用率。

**响应示例<a name="sectionclusteropstoolsresourceexample"></a>**

```json
{"request_id": "3e6ca7b69e224443beb154543b7cbfb9", "tool": "cluster_resource",
 "status": "ok", "elapsed_ms": 118,
 "data": {"node": {"ascend-142": {
   "cpu": {"capacity_cores": 128.0, "allocatable_cores": 120.0, "used_cores": 2.0, "idle_cores": 118.0},
   "Ascend910": {"capacity": 8, "allocatable": 7, "used": 1, "idle": 6, "fault": 1, "isolated": 1}}}},
 "errors": []}
```

### task_status<a name="sectionclusteropstoolstask"></a>

**表20**  task_status的data字段

|字段|类型|说明|
|--|--|--|
|namespace|string|任务命名空间。|
|job_name|string|任务名。|
|job_type|string/null|`acjob`或`vcjob`；按名称同时命中两类、或无任务对象时为null。|
|status|string|任务状态：`running`、`completed`、`failed`、`pending`、`unknown`。|
|pods|array|当前任务实例的实时Pod，见[表21](#tableclusteropspods)；恒为数组，无可核验Pod时为`[]`。|

**表21**  pods[]字段<a name="tableclusteropspods"></a>

|字段|类型|说明|
|--|--|--|
|pod_name|string|Pod名称。|
|status|string|归一后的Pod状态：`running`、`pending`、`completed`、`failed`、`unknown`；Running但容器处于`CrashLoopBackOff`时原样输出`CrashLoopBackOff`。|

>[!NOTE]
>任务状态只由ACJob/VCJob对象的`status`归一，**不从Pod推断**；Pod只用于展示。对象状态不可判定时`status=unknown`且信封为`partial`。

## 诊断工具（diagnose）<a name="sectionclusteropstoolsdiagnose"></a>

**表22**  诊断工具入参

|参数|类型|必填|说明|
|--|--|--|--|
|job|string|是|任务名（任务CR名称）。|
|namespace|string|否|任务所在命名空间，默认`default`。|
|refresh|boolean|否|忽略缓存强制重新诊断，默认false。|

`diagnose`的`data`与旧`/diag`的扁平业务字段一一对应：`pods`、`collected`、`diag_report`、`diag_report_text`、`diag_input_dir`、`diag_output_dir`、`incomplete_nodes`、`empty_nodes`、`final_text`、`cached`、`cached_note`、`llm_error`、`error`；字段含义与[Agent Core、Node Collector与Kubectl Plugin](20_clusterops_agent.md#sectionascendclusteropshttp)中的`/diag`响应说明一致。

`POST /tools/diagnose`（`kubectl clusterops tool diagnose`）在业务结果之上自动执行旧版展示：有模型时总结，无模型时回退为确定性报告文本，因此可执行诊断业务的HTTP响应中`final_text`始终存在；`cached=true`时优先复用缓存中的`final_text`。通过`TOOL_REGISTRY.execute()`直接调用时，只返回原始业务`data`，不执行总结，`final_text`仅在缓存结果本身带有该字段时保留。

**表23**  诊断工具的status与错误码

|场景|status|error code|
|--|--|--|
|诊断成功|`ok`|—|
|部分节点未采集或上报为空|`partial`|—|
|任务不存在|`error`|`INTERNAL_ERROR`|
|同一任务诊断进行中|`error`|`INTERNAL_ERROR`|
|诊断失败|`error`|`INTERNAL_ERROR`|
|诊断超时|`error`|`INTERNAL_ERROR`|

## 错误码与状态映射<a name="sectionclusteropstoolserrors"></a>

**表24**  `errors[]`错误码

|错误码|HTTP状态|典型场景|
|--|--|--|
|`INVALID_ARGUMENT`|422|参数类型、取值或未知参数校验失败；`node`超过20000项。|
|`TOOL_NOT_FOUND`|404|工具名不在目录中。|
|`TARGET_NOT_FOUND`|404|显式`node`全部不在集群内（`status=error`，message`所有节点都找不到`）；部分节点不存在时为`status=partial`，每个缺失节点一条`节点 <name> 不存在`。|
|`INTERNAL_ERROR`|500|未预期内部异常；诊断失败、超时、并发冲突也归此码。|

>[!NOTE]
>除显式目标不存在外，业务层查询问题**不会**出现在`errors[]`，也不改变HTTP状态，只记服务端日志。日志中的业务码包括：`SOURCE_NOT_FOUND`（来源缺失）、`SOURCE_FORBIDDEN`（无权限）、`SOURCE_INVALID`（内容损坏、字段非法）、`SOURCE_UNAVAILABLE`（超时/不可达/分页中断）；`cluster_resource`对读取间隙仍会记日志`TARGET_NOT_FOUND`。它们的影响由`status=partial`和字段null表达。

## 查询行为边界<a name="sectionclusteropstoolsbounds"></a>

- **读取超时与分页**：每次Kubernetes读取使用固定的5秒单次超时、每页500条分页，跟随`continue`翻页直至读完；超时与不可达统一归类为`SOURCE_UNAVAILABLE`，不重试。查询不设时间/字节预算，也不暴露配置项。
- **范围语义**：`node=null`表示ServiceAccount可见的全量节点；`node=[]`表示显式空范围，返回空结果并标记`ok`，不触发全集群读取；显式节点列表去重保序，上限20000项。
- **显式节点的存在性**：`node`非空时，先按ServiceAccount可见的Node列表校验。全部节点不存在 → `status=error`、HTTP 404、`errors[].code=TARGET_NOT_FOUND`、message`所有节点都找不到`，不读取任何来源；部分节点不存在 → 查询仍执行，缺失节点从读取范围剔除，`status=partial`并逐节点返回`节点 <name> 不存在`。Node列表本身读不全（权限/超时/分页中断）时跳过存在性校验，按原有覆盖规则降级。
- **完整性判定**：
  - `faulty_device`：全集群时要求每个带整卡资源键的节点都被设备事实覆盖且都有健康状态；指定节点时要求每个请求节点都被覆盖且有健康状态。回退成功不影响状态（仍`ok`），回退后仍未被覆盖才为`partial`。Switch/DPU/公共故障是补充源，读取失败不影响状态。
  - `isolated_device`：隔离CM可读且节点名单可完整解析即`ok`；CM不存在视为合法空集。
  - `cluster_resource`：Node/Pod列表完整、请求节点都存在、`capacity/allocatable`可解析且`capacity≥allocatable`即`ok`；请求节点存在但没有任何整卡资源键时按`ok`处理（确认无NPU资源、不产生节点条目），只带无法识别的NPU类资源键（310、vNPU等）时为`partial`。故障/隔离源读取失败不改变状态，只让对应字段为null。
  - `task_status`：任务类型唯一、任务对象UID可核验、状态可判定且Pod列表完整才为`ok`；否则`partial`。
- **读取中断**：分页在某一页失败后停止翻页并保留已读数据；显式节点逐个读取时，遇到非404失败即停止后续读取并保留已读节点。以上均体现在`status=partial`，不进入`errors`。
- **节点级故障与芯片级故障**：设备摘要中`npu_name`为空的条目为节点级事实（`chip_type=null`），`npu_name`非空为芯片级；节点汇总`FaultDevList[]`全部是节点级事实。节点级补源只在设备事实中没有节点级条目时读取。
- **故障/隔离统计**：`cluster_resource`的`fault`与`isolated`是两套独立证据；不按`capacity−allocatable`拆分或闭合。同一张卡可同时计入两列。
- **预隔离**：`fault_handling=PreSeparateNPU`保留在`faulty_device`的`devices[].faults[]`中，不单独计数，也不计为已隔离卡（隔离只认ClusterD名单）。
- **任务状态推断**：只读任务对象自身`status`；Pod不参与判定。对象不存在或类型冲突时`status=unknown`、`pods=[]`、信封`partial`。

## Kubectl Plugin使用<a name="sectionclusteropstoolscli"></a>

**表25**  查询相关子命令

|命令|说明|
|--|--|
|`kubectl clusterops tools [--json]`|列出能力目录（对应`GET /tools`）。|
|`kubectl clusterops tool <name> [--param NAME=VALUE ...] [--params-json JSON]`|通用调用入口（对应`POST /tools/{name}`），也是唯一的业务调用子命令。|

**参数写法<a name="sectionclusteropstoolscliparams"></a>**

- `--param NAME=VALUE`可重复；同名出现多次表示数组，单值也会被包装为一项数组。
- `--params-json '<JSON>'`直接传完整参数对象，用于空数组、嵌套对象和精确的`null`；与`--param`互斥。
- `--param NAME=`表示空字符串；`null`用`--param NAME=null`；空数组请用`--params-json '{"node": []}'`。
- `VALUE`按`input_schema`转换：布尔只接受`true`/`false`，整数/数字严格解析，未知参数和类型错误在本地报错（退出码2），不发起执行请求。
- 通用入口每次先读取descriptor再执行，不缓存schema；服务端仍是最终校验方。

**示例<a name="sectionclusteropstoolscliexample"></a>**

```shell
kubectl clusterops tools
kubectl clusterops tools --json
kubectl clusterops tool faulty_device --param node=worker-a
kubectl clusterops tool isolated_device --param node=worker-a --json
kubectl clusterops tool cluster_resource
kubectl clusterops tool task_status --param job=train-a --param namespace=training
kubectl clusterops tool faulty_device --param node=worker-a --param node=worker-b
kubectl clusterops tool isolated_device --params-json '{"node": []}'
kubectl clusterops tool diagnose --param job=train-a --json
```

**退出码<a name="sectionclusteropstoolscliexit"></a>**

|场景|退出码|
|--|--|
|查询`status=ok`|0|
|查询`status=partial`|3|
|查询或诊断`status=error`|1|
|本地参数/转换失败|2|

>[!NOTE]
>`--json`模式stdout只输出一个JSON对象；表格模式下`partial`会在stderr打印提示（含`request_id`）并逐条打印`errors[]`中的缺失节点，其余缺口原因需查agent-core日志。表格列与各工具的对应关系见[查询集群与任务状态](../04_usage/14_clusterops_agent/05_querying_cluster_status.md)。

## 与旧/diag接口的关系<a name="sectionclusteropstoolsdiag"></a>

- `POST /diag`保持不变：返回旧诊断JSON，不带工具信封，供`kubectl ascend_diag`使用。
- `POST /tools/diagnose`是同一诊断链的信封化入口，HTTP响应的`data`字段与`/diag`业务字段一一对应（含`final_text`），供`kubectl clusterops tool diagnose`使用；直接调用Registry（`TOOL_REGISTRY.execute()`）得到不含新总结的原始`data`。
- 两者共用同一诊断服务函数；`/diag`不经过工具注册表，因此不受工具参数模型约束。
