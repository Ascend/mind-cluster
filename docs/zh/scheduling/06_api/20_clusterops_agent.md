# Agent Core、Node Collector与Kubectl Plugin<a name="ZH-CN_TOPIC_00000026ascendclusterops01"></a>

集群运维Agent特性包含以下三个组件，Agent Core是诊断的集中控制中心，Node Collector负责节点日志采集与清洗，Kubectl Plugin提供`kubectl ascend_diag`和`kubectl clusterops`命令入口。

- **Agent Core**：Deployment形态部署，维护任务与Pod的中心关系缓存，向各节点Node Collector下发采集指令，聚合清洗产物并调用ascend-fd diag执行集中诊断，生成诊断报告。
- **Node Collector**：DaemonSet形态部署，每个NPU节点一份，按采集契约从宿主机采集任务日志，本地调用ascend-fd parse完成清洗并上报。
- **Kubectl Plugin**：纯标准库实现，复用本机kubeconfig鉴权，通过`kubectl port-forward`访问Agent Core服务。

## ConfigMap说明<a name="sectionascendclusteropscm"></a>

### Agent Core创建ConfigMap

- clusterops-pathmap，详细说明请参见[表1](#tableclusteropspathmap)。
- agent-core-relcache，详细说明请参见[表2](#tableagentcorerelcache)。

**表1**  clusterops-pathmap

<a name="tableclusteropspathmap"></a>

该ConfigMap位于cluster-system命名空间，数据键为`pathmap.json`，用于记录任务Pod的挂载对、env值和主机IP，供Agent Core重启时会从中读取内容恢复关系数据，并在TriggerCollect采集指令中随Pod列表下发。快照按顺序填充分片写入（`clusterops-pathmap`、`clusterops-pathmap-1`…`clusterops-pathmap-N`，分片0写满后写分片1，依此类推；分片数固定为100），每个分片不超过ConfigMap的1MB数据上限。

|参数|说明|
|--|--|
|profiles|任务Pod的挂载对集合。键为profile_id（挂载对内容哈希），值为挂载对列表`pairs`（`host:container`格式），同一任务挂载相同的Pod共享一个profile。|
|pods|任务Pod条目。键为pod UID，包含以下字段。|
|-profile|该Pod所属的profile_id。|
|-env|任务Pod的字面env值（跳过valueFrom引用），用于解析env类采集实体（如`ASCEND_PROCESS_LOG_PATH`）。|
|-pod_ip|Pod主机IP（hostIP，为空时回退Pod IP），用于共享存储plog目录匹配。|
|-namespace|Pod命名空间。|
|-owner_name|所属任务（CR）名称。|
|-owner_uid|所属任务（CR）UID，用于任务重建时清理旧实例条目。|
|-deleted_at|Pod删除时间戳，存活Pod为null。|
|-shard|该条目所属分片号（内部路由字段）。|

**表2**  agent-core-relcache

<a name="tableagentcorerelcache"></a>

该ConfigMap位于cluster-system命名空间，数据键为`snapshot`，用于保存任务与Pod的中心关系缓存快照，Agent Core重启时会从中读取内容恢复关系数据。快照按顺序填充分片写入（`agent-core-relcache`、`agent-core-relcache-1`…`agent-core-relcache-N`，分片0写满后写分片1，依此类推；分片数固定为100），每个分片不超过ConfigMap的1MB数据上限。

|参数|说明|
|--|--|
|pods|任务Pod列表，每个元素包含以下字段。|
|-pod_name|Pod名称。|
|-pod_uid|Pod UID。|
|-namespace|Pod命名空间。|
|-node|Pod所在节点。|
|-host_ip|Pod主机IP，用于共享存储plog目录匹配。|
|-rank|Pod在任务中的Rank（来自`ascend/job-rank`或`rank`注解）。|
|-owner_name|所属任务（CR）名称。|
|-owner_uid|所属任务（CR）UID，用于任务重建时清理旧实例条目。|
|-deleted_at|Pod删除时间戳，存活Pod为null。|
|-phase|Pod当前阶段，取值Pending、Running、Succeeded、Failed、Unknown。该阶段由Kubernetes（kubelet）依据容器退出状态与重启策略判定并写入`pod.status.phase`，Agent Core通过watch机制读取：所有容器退出码为0时为Succeeded；任一容器退出码非0或被系统终止（如OOM）且不再重启时为Failed。|
|-shard|该条目所属分片号（内部路由字段）。|

>[!NOTE]
>Agent Core启动时读取该快照恢复任务→Pod关系数据，诊断时按任务名查询该关系表获取Pod及所在节点，再向各节点Node Collector下发采集指令。已删除Pod条目默认保留7天（可通过agent-core.yaml中环境变量POD_TTL进行修改）；内存表容量上限与分片快照总容量一致（100MB），超出后按最老优先淘汰。

### 其他业务相关ConfigMap/Secret

**表3**  其他业务相关ConfigMap/Secret

|资源|命名空间|写入方|读取方|作用|
|--|--|--|--|--|
|agent-core-task-crds|cluster-system|用户|Agent Core|任务CR GVK配置，Agent Core启动时读取一次，仅跟踪这些任务CR管理的Pod。新增任务类型时需同时在该ConfigMap增补GVK，详见[配置任务类型](../04_usage/14_clusterops_agent/01_configuring_task_crds.md)。|
|collect-manifest|cluster-system|Kubectl Plugin（`kubectl ascend_diag --collect-manifest`）|Node Collector|采集契约，Node Collector每次采集时读取，详见[配置日志采集](../04_usage/14_clusterops_agent/02_configuring_log_collection.md)。|
|llm-secret|mindx-dl|Kubectl Plugin（`kubectl clusterops --create-llm-config`）|Agent Core|LLM配置（api-key、base-url、model），Agent Core实时监听，详见[配置和清除LLM](../04_usage/14_clusterops_agent/04_configuring_llm.md)。|

### 容量与历史数据保留<a name="sectionascendclusteropscapacity"></a>

relcache/pathmap快照按顺序填充分片存储，分片0写满后写分片1，依此类推，总容量=分片数×1MB，分片数固定为100（100MB）；本地内存表上限与快照容量一致，超出后按最老优先淘汰。

按每任务8卡估算，100MB容量可容纳的数据量如下：

|任务Pod形态|单条记录大小|100MB可存记录数|可容纳任务数|
|--|--|--|--|
|每任务1个Pod、每Pod8卡|约264B|约39.7万条|约4个十万卡任务|

- 昇腾单集群上限10万卡（约1.25万个节点）。
- 4个十万卡任务包含删除后保留的默认7天内的任务。
- 历史任务过期：已删除Pod条目默认保留**7天**（`POD_TTL`，可在agent-core.yaml中调整，单位秒），超期后从内存表与分片快照中删除；内存占用超过容量上限时，按最老优先淘汰。

## HTTP接口<a name="sectionascendclusteropshttp"></a>

**表4**  诊断接口

| 项目 | 说明 |
|------|------|
| 路径 | `/diag` |
| 方法 | POST |
| 默认端口 | 9700 |
| 协议 | HTTP |

**表5**  请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| job | string | 是 | 任务名，即任务CR名称。 |
| namespace | string | 否 | 任务所在命名空间，默认default。 |
| refresh | bool | 否 | 是否忽略缓存强制刷新，默认false。 |

**表6**  响应说明

| 参数 | 类型 | 说明 |
|------|------|------|
| pods | array | 任务涉及的Pod信息。 |
| collected | array | 各节点采集与清洗结果（节点、host_ip、是否成功等）。 |
| diag_report | object | ascend-fd诊断报告。 |
| diag_report_text | string | 确定性诊断报告文本（含未完成节点等补充说明）。 |
| incomplete_nodes | array | 采集不完整的节点列表。 |
| empty_nodes | array | 未采集到日志的节点列表。 |
| final_text | string | 诊断报告文本，为LLM智能总结或确定性诊断报告。 |
| error | string | 错误信息，诊断失败（任务不存在、采集失败等）时存在。 |
| cached | bool | 是否命中缓存。 |
| cached_note | string | 命中缓存时的提示信息。 |
| llm_error | string | LLM总结失败原因，配置LLM且调用失败时存在。 |

>[!NOTE]
>
> - 9700为集群内部端口，不对外暴露；集群外请通过Kubectl Plugin（`kubectl ascend_diag`，经`kubectl port-forward`访问）调用该接口，参见[Kubectl Plugin](#sectionascendclusteropskubectl)。

## 诊断结果缓存<a name="sectionascendclusteropscache"></a>

Agent Core提供两级诊断结果缓存：近期结果缓存（内存）和持久结果缓存（磁盘）。

**近期结果缓存（内存，10秒）**

同一任务（ns, job）在10秒内重复诊断时，直接返回最近一次结果（标记cached并提示"This is cached data; add --refresh for real-time data"），不再重新采集。保留时长可通过环境变量DIAG_RECENT_TTL调整。

**持久结果缓存（磁盘）**

每次诊断成功后，诊断结果即写入持久缓存文件（任务运行中或停止均缓存）；再次诊断同一任务时，若命中缓存，直接返回缓存结果（提示"Result from cache; run --refresh for the latest diagnosis"）。

**表7**  缓存文件说明

| 项目 | 说明 |
|------|------|
| 缓存文件 | `{缓存目录}/cache/<ns>_<job>.json`，按任务命名空间与任务名命名，跨天可命中。 |
| 缓存目录 | 默认`{AGENT_WORK_ROOT}/cache`，即`/user/clusterops/agent-core/cache`（宿主导出目录），可通过环境变量`AGENT_CACHE_DIR`覆盖。 |
| 淘汰机制 | 缓存文件不设固定保留期，每次诊断后按`WORK_ROOT`目录总容量配额（默认10GB，可通过环境变量`WORK_ROOT_MAX_BYTES`调整）淘汰，超限时连同诊断产物按最老优先删除。 |
| 可诊断窗口 | 任务结束或Pod删除后7天内（`POD_TTL`默认值，可通过环境变量`POD_TTL`调整）relcache中仍有该任务记录，仍可执行诊断并命中缓存；超期后任务记录过期，返回任务不存在。 |

**表8**  缓存数据说明

| 参数 | 说明 |
|------|------|
| pods | 任务涉及的Pod信息。 |
| diag_report | ascend-fd诊断报告。 |
| diag_report_text | 确定性诊断报告文本（含未完成节点等补充说明）。 |
| incomplete_nodes | 采集不完整的节点列表。 |
| empty_nodes | 未采集到日志的节点列表。 |
| final_text | 诊断报告文本（LLM智能总结或确定性诊断报告），LLM总结生成后写回缓存，命中缓存时直接返回，不重复调用LLM。 |
| error | 错误信息。 |
| cached_at | 缓存写入时间戳。 |

**缓存条件**

- 每次诊断成功后均缓存诊断结果，不依赖任务CR状态；命中缓存时直接返回缓存结果，提示使用 --refresh强制刷新获取最新结果。
- 任务CR不存在（404，任务已删除）时：若relcache中仍有该任务Pod记录（删除TTL内，日志仍保留在节点/共享盘上），仍可执行诊断并缓存；仅当relcache中也没有该任务记录时，直接返回任务不存在。

>[!NOTE]
>缓存与诊断产物存放在Agent Core的`WORK_ROOT`目录（默认`/user/clusterops/agent-core`，宿主导出）。若需在Pod重建（重调度）后仍保留缓存与产物，需要为Agent Core配置PVC挂载`WORK_ROOT`目录；hostPath挂载仅支持同一节点上的Pod重建保留。

## gRPC接口<a name="sectionascendclusteropsgrpc"></a>

**表9**  gRPC接口说明

| 服务 | 端口 | 接口 | 方向 | 说明 |
|------|------|------|------|------|
| agent-core | 9710 | UploadResult | node-collector → agent-core | 接收Node Collector上报的采集清洗结果（tar.gz流式回传）。 |
| node-collector | 9720 | TriggerCollect | agent-core → node-collector | 接收Agent Core下发的采集指令，立即返回受理结果并后台异步执行采集。 |

>[!NOTE]
>
> - 9710和9720为集群内部端口，不对外暴露，仅供Agent Core与Node Collector在集群内互相调用。

## Kubectl Plugin<a name="sectionascendclusteropskubectl"></a>

### kubectl ascend_diag<a name="sectionascendclusteropsascenddiag"></a>

按任务维度触发集群运维Agent。

**表10**  参数说明

| 参数 | 说明 |
|------|------|
| --job | 任务名，即任务CR名称（必填）。和--collect-manifest互斥。 |
| -n / --namespace | 任务所在命名空间，默认default。搭配--job使用。 |
| --refresh | 忽略缓存，强制重新执行诊断并刷新缓存。搭配--job使用。 |
| --json | 输出完整JSON响应。搭配--job使用。 |
| --collect-manifest | 将本地采集契约写入集群ConfigMap后退出，单独使用和其他命令互斥。 |
| --agent-core-service | 指定agent-core服务，格式svc.ns:9700，默认agent-core.mindx-dl:9700。搭配--job使用。 |

使用示例：

```shell
kubectl ascend_diag --job job-x -n training
kubectl ascend_diag --job job-x --refresh
kubectl ascend_diag --collect-manifest ./collect_manifest.yaml
```

### kubectl clusterops<a name="sectionascendclusteropsclusterops"></a>

配置和清除LLM服务配置。

**表11**  参数说明

| 参数 | 说明 |
|------|------|
| --create-llm-config | 配置LLM Secret（llm-secret）。和--clear-llm-config互斥。|
| --clear-llm-config | 清除全部LLM配置（删除llm-secret）。和--create-llm-config、--base-url、--model互斥。|
| --base-url | LLM Base URL，可通过环境变量LLM_BASE_URL或交互输入指定。需搭配--create-llm-config使用。 |
| --model | LLM模型名，可通过环境变量LLM_MODEL或交互输入指定。需搭配--create-llm-config使用。 |

>[!NOTE]
>更新或清除后实时生效，无需重启Agent Core。

使用示例：

```shell
kubectl clusterops --create-llm-config
kubectl clusterops --create-llm-config --base-url https://open.bigmodel.cn/api/paas/v4 --model glm-4-plus
kubectl clusterops --clear-llm-config
```
