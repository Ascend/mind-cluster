# Agent Core、Node Collector与Kubectl Plugin<a name="ZH-CN_TOPIC_00000026ascendclusterops01"></a>

集群运维Agent特性包含以下三个组件，Agent Core是诊断的集中控制中心，Node Collector负责节点日志采集与清洗，Kubectl Plugin提供`kubectl ascend_diag`和`kubectl clusterops`命令入口。

- **Agent Core**：Deployment形态部署，维护任务与Pod的中心关系缓存，向各节点Node Collector下发采集指令，聚合清洗产物并调用ascend-fd diag执行集中诊断，生成诊断报告。
- **Node Collector**：DaemonSet形态部署，每个NPU节点一份，按采集契约从宿主机采集任务日志，本地调用ascend-fd parse完成清洗并上报。
- **Kubectl Plugin**：纯标准库实现，复用本机kubeconfig鉴权，通过`kubectl port-forward`访问Agent Core服务。

## ConfigMap说明<a name="sectionascendclusteropscm"></a>

Agent Core启动后，会创建如下ConfigMap：

- clusterops-pathmap，详细说明请参见[表1](#tableclusteropspathmap)。
- agent-core-relcache，详细说明请参见[表2](#tableagentcorerelcache)。

**表1**  clusterops-pathmap

<a name="tableclusteropspathmap"></a>

该ConfigMap位于cluster-system命名空间，数据键为`pathmap.json`，用于记录任务Pod的挂载关系，供Node Collector反查宿主机路径采集日志。

|参数|说明|
|--|--|
|profiles|任务Pod的挂载对集合。键为profile_id（挂载对内容哈希），值为挂载对列表`host:container`，同一任务挂载相同的Pod共享一个profile。|
|pods|任务Pod条目。键为pod UID，仅保留`profile`、`env`、`deleted_at`字段。|
|-profile|该Pod所属的profile_id。|
|-env|任务Pod的字面env值（跳过valueFrom引用），用于解析env类采集实体（如`ASCEND_PROCESS_LOG_PATH`）。|
|-deleted_at|Pod删除时间戳，存活Pod为null。|

>[!NOTE]
>Node Collector监听该ConfigMap并在本地缓存，采集时按采集契约的实体类型匹配挂载对、反查宿主机路径后读取日志。已删除Pod条目保留7天（POD_TTL），Pod删除后仍可据此反查宿主路径。采用变更驱动同步：内存表更新后立即同步，失败时自动退避重试。

**表2**  agent-core-relcache

<a name="tableagentcorerelcache"></a>

该ConfigMap位于cluster-system命名空间，数据键为`snapshot`，用于保存任务与Pod的中心关系缓存快照，Agent Core重启时恢复关系数据。

|参数|说明|
|--|--|
|pods|任务Pod列表，每个元素包含以下字段。|
|-pod_name|Pod名称。|
|-pod_uid|Pod UID。|
|-namespace|Pod命名空间。|
|-node|Pod所在节点。|
|-rank|Pod在任务中的Rank（来自`ascend/job-rank`或`rank`注解）。|
|-owner_name|所属任务（CR）名称。|
|-deleted_at|Pod删除时间戳，存活Pod为null。|
|-phase|Pod当前阶段（Running、Succeeded、Failed等）。|

>[!NOTE]
>Agent Core启动时读取该快照恢复任务→Pod关系数据，诊断时按任务名查询该关系表获取Pod及所在节点，再向各节点Node Collector下发采集指令。已删除Pod条目保留7天（POD_TTL）。

## 相关ConfigMap/Secret说明<a name="sectionascendclusteropsrelated"></a>

以下ConfigMap/Secret与集群运维Agent密切相关：

**表3**  相关ConfigMap/Secret

|资源|命名空间|写入方|读取方|作用|
|--|--|--|--|--|
|agent-core-task-crds|cluster-system|用户|Agent Core|任务CR GVK配置，Agent Core启动时读取一次，仅跟踪这些任务CR管理的Pod。新增任务类型时需同时在该ConfigMap增补GVK并在ClusterRole中增补对应get权限，详见[配置任务类型](../04_usage/14_clusterops_agent/01_configuring_task_crds.md)。|
|collect-manifest|cluster-system|Kubectl Plugin（`kubectl ascend_diag --collect-manifest`）|Node Collector|采集契约，Node Collector每次采集时读取，详见[配置日志采集](../04_usage/14_clusterops_agent/02_configuring_log_collection.md)。|
|llm-secret|mindx-dl|Kubectl Plugin（`kubectl clusterops`）|Agent Core|LLM配置（api-key、base-url、model），Agent Core实时监听，详见[配置和清除LLM](../04_usage/14_clusterops_agent/04_configuring_llm.md)。|

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
| diag_report | object | ascend-fd诊断报告。 |
| final_text | string | 诊断报告文本，为LLM智能总结或确定性诊断报告。 |
| error | string | 错误信息，诊断失败（任务不存在、采集失败等）时存在。 |
| cached | bool | 是否命中缓存。 |
| cached_note | string | 命中缓存时的提示信息。 |
| llm_error | string | LLM总结失败原因，配置LLM且调用失败时存在。 |

请求示例：

```shell
curl -X POST localhost:9700/diag -H 'Content-Type: application/json' \
     -d '{"job":"job-x","namespace":"default","refresh":false}'
```

## 诊断结果缓存<a name="sectionascendclusteropscache"></a>

Agent Core提供两级诊断结果缓存：近期结果缓存（内存）和持久结果缓存（磁盘）。

**近期结果缓存（内存，10秒）**

同一任务（ns, job）在10秒内重复诊断时，直接返回最近一次结果（标记cached并提示"这是缓存数据，需要实时数据，请加 --refresh"），不再重新采集。保留时长可通过环境变量DIAG_RECENT_TTL调整。

**持久结果缓存（磁盘）**

任务处于停止状态时，诊断结果写入持久缓存文件；再次诊断同一任务时，若命中缓存且任务仍处于停止状态，直接返回缓存结果（提示"结果来自缓存，如需最新诊断，请执行 --refresh强制刷新"）。

**表7**  缓存文件说明

| 项目 | 说明 |
|------|------|
| 缓存文件 | `{缓存目录}/cache/<ns>_<job>.json`，按任务命名空间与任务名命名，跨天可命中。 |
| 缓存目录 | 默认`{AGENT_WORK_ROOT}/cache`，即`/user/clusterops/agent-core/cache`（宿主导出目录），可通过环境变量`AGENT_CACHE_DIR`覆盖。 |
| 保留期 | 7天（POD_TTL默认值，可通过环境变量`POD_TTL`调整），任务结束或Pod删除后7天内仍可执行诊断并命中缓存。 |

**表8**  缓存数据说明

| 参数 | 说明 |
|------|------|
| pods | 任务涉及的Pod信息。 |
| diag_report | ascend-fd诊断报告。 |
| final_text | 诊断报告文本（LLM智能总结或确定性诊断报告），LLM总结生成后写回缓存，命中缓存时直接返回，不重复调用LLM。 |
| error | 错误信息。 |
| cached_at | 缓存写入时间戳。 |
| task_state | 缓存时任务状态（stopped）。 |

**缓存条件**

- 仅任务处于停止状态时缓存，运行中的任务不缓存。停止状态通过查询任务CR的`status.conditions`判定（大小写不敏感），终态包括JobSucceeded、JobFailed、Succeeded、Failed、Stopped、Complete、Completed、Terminated、Aborted。
- 任务CR不存在（404，任务已删除）视为停止状态，结果可缓存。
- 任务CR状态读取不到或滞后时，若任务的所有Pod均已达到终态（Succeeded/Failed），结果也可缓存。
- 任务CR状态查询失败（RBAC权限缺失、网络异常等）时判定为unknown，此时不缓存（但不阻塞诊断）。

>[!NOTE]
>缓存与诊断产物存放在Agent Core的`WORK_ROOT`目录（默认`/user/clusterops/agent-core`，宿主导出）。若需在Pod重建（重调度）后仍保留缓存与产物，需要为Agent Core配置PVC挂载`WORK_ROOT`目录；hostPath挂载仅支持同一节点上的Pod重建保留。

## gRPC接口<a name="sectionascendclusteropsgrpc"></a>

**表9**  gRPC接口说明

| 服务 | 端口 | 接口 | 方向 | 说明 |
|------|------|------|------|------|
| agent-core | 9710 | UploadResult | node-collector → agent-core | 接收Node Collector上报的采集清洗结果（tar.gz流式回传）。 |
| node-collector | 9720 | TriggerCollect | agent-core → node-collector | 接收Agent Core下发的采集指令，立即返回受理结果并后台异步执行采集。 |

## Kubectl Plugin<a name="sectionascendclusteropskubectl"></a>

### kubectl ascend_diag<a name="sectionascendclusteropsascenddiag"></a>

按任务维度触发集群运维Agent。

**表10**  参数说明

| 参数 | 说明 |
|------|------|
| --job | 任务名，即任务CR名称（必填）。 |
| -n / --namespace | 任务所在命名空间，默认default。 |
| --refresh | 忽略缓存，强制重新执行诊断并刷新缓存。 |
| --json | 输出完整JSON响应。 |
| --collect-manifest | 将本地采集契约写入集群ConfigMap后退出。 |
| --agent-core-service | 指定agent-core服务，格式svc.ns:9700，默认agent-core.mindx-dl:9700。 |

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
| --create-llm-config | 配置LLM Secret（llm-secret）。 |
| --clear-llm-config | 清除全部LLM配置（删除llm-secret）。 |
| --base-url | LLM Base URL，可通过环境变量LLM_BASE_URL或交互输入指定。 |
| --model | LLM模型名，可通过环境变量LLM_MODEL或交互输入指定。 |

>[!NOTE]
>`--create-llm-config`与`--clear-llm-config`互斥，不能同时使用。更新或清除后实时生效，无需重启Agent Core。

使用示例：

```shell
kubectl clusterops --create-llm-config
kubectl clusterops --create-llm-config --base-url https://open.bigmodel.cn/api/paas/v4 --model glm-4-plus
kubectl clusterops --clear-llm-config
```
