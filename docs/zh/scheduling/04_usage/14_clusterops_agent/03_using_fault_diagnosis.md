# 使用故障诊断<a name="ZH-CN_TOPIC_00000026faultdiagnosis04"></a>

当训练/推理任务发生异常时，可通过`kubectl ascend_diag`命令按任务维度一键触发集群运维Agent。

## 诊断流程<a name="sectionfaultdiagnosisprocess"></a>

1. Agent Core按任务名查询中心关系表，获取任务的所有Pod及所在节点。
2. Agent Core向各节点的Node Collector下发采集指令，按采集契约从宿主机采集任务日志（plog、系统日志、设备日志等）。
3. Node Collector本地调用`ascend-fd parse`对日志进行清洗，将清洗产物打包上报给Agent Core。
4. Agent Core聚合各节点清洗产物，组装为诊断输入目录，调用`ascend-fd diag`执行集中诊断，生成诊断报告（可选对接LLM进行智能总结）。

## 执行诊断<a name="sectionfaultdiagnosisexecute"></a>

在任务发生异常后，执行以下命令发起诊断（默认命名空间为default，可通过`-n`指定）：

```shell
kubectl ascend_diag --job job-x -n training
```

其中：

| 参数 | 说明 |
|---|---|
| `--job` | 任务名，即任务CR名称（必填） |
| `-n` / `--namespace` | 任务所在命名空间，默认`default` |
| `--refresh` | 忽略缓存，强制重新执行诊断并刷新缓存 |
| `--json` | 输出完整JSON响应 |
| `--collect-manifest` | 将本地采集契约写入集群ConfigMap后退出 |
| `--agent-core-service` | 指定Agent Core服务，格式`svc.ns:9700`（默认`agent-core.mindx-dl:9700`） |

**诊断输出样例**

```text
$ kubectl ascend_diag --job job-x -n training
The diag job starts. Please wait. Job id: [20260918163512389528_b1b293d4-1c70-4e11-916e-a9023da3660a], run log file is [ascend_faultdiag_36.log].
+--------------------------------------------------------------------------------------------------------------------------------------------+
|                                                        Ascend-fd Fault-Diag Report                                                         |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
|   版本信息   |    类型    | 版本                                                                                                           |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
|              | Fault-Diag | 26.1.0                                                                                                         |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
| 根因节点分析 |    类型    | 描述                                                                                                           |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
|              |  根因节点  | ['worker0 device-7']                                                                                           |
|              |  现象描述  | 所有节点的Plog都没有记录超时类错误日志。日志中有报错的节点为疑似根因节点，请排查。                             |
|              |  首错节点  | worker0 device-7: 2026-09-18-12:19:30.141986                                                                   |
|              |  尾错节点  | worker0 device-7: 2026-09-18-12:19:30.141986                                                                   |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
| 故障事件分析 |    类型    | 描述                                                                                                           |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
| 疑似根因故障 |   状态码   | Comp_OS_Service_Container_04                                                                                   |
|              |  故障分类  | 类型:Software 组件:HostOS 模块:OS                                                                              |
|              |  故障设备  | ['worker0']                                                                                                    |
|              |  故障名称  | 容器存储异常                                                                                                   |
|              |  故障描述  | docker存储时延。                                                                                               |
|              |  建议方案  | 1. 更新存储磁盘，重启创建容器；                                                                                |
|              |  关键日志  | Sep 18 12:19:24 localhost /usr/sbin/irqbalance[3091]: Cannot change IRQ 1145 affinity: No space left on device |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
The diag job is complete.

Tip: to output the full diagnosis JSON, add --json after the command
```

使用`--json`参数可查看完整的JSON诊断响应（包含采集到的Pod信息、诊断报告等）：

```shell
kubectl ascend_diag --job job-x -n training --json
```

## 重复执行与缓存提示<a name="sectionfaultdiagnosisrepeat"></a>

同一任务在**10秒内**重复执行诊断时，会直接返回最近一次诊断结果（缓存数据），并提示使用`--refresh`获取实时数据：

```text
$ kubectl ascend_diag --job job-x -n training
Result from cache; run --refresh for the latest diagnosis
The diag job starts. Please wait. Job id: [20260918163512389528_b1b293d4-1c70-4e11-916e-a9023da3660a], run log file is [ascend_faultdiag_36.log].
+--------------------------------------------------------------------------------------------------------------------------------------------+
|                                                        Ascend-fd Fault-Diag Report                                                         |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
|   版本信息   |    类型    | 版本                                                                                                           |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
|              | Fault-Diag | 26.1.0                                                                                                         |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
| 根因节点分析 |    类型    | 描述                                                                                                           |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
|              |  根因节点  | ['worker0 device-7']                                                                                           |
|              |  现象描述  | 所有节点的Plog都没有记录超时类错误日志。日志中有报错的节点为疑似根因节点，请排查。                             |
|              |  首错节点  | worker0 device-7: 2026-09-18-12:19:30.141986                                                                   |
|              |  尾错节点  | worker0 device-7: 2026-09-18-12:19:30.141986                                                                   |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
| 故障事件分析 |    类型    | 描述                                                                                                           |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
| 疑似根因故障 |   状态码   | Comp_OS_Service_Container_04                                                                                   |
|              |  故障分类  | 类型:Software 组件:HostOS 模块:OS                                                                              |
|              |  故障设备  | ['worker0']                                                                                                    |
|              |  故障名称  | 容器存储异常                                                                                                   |
|              |  故障描述  | docker存储时延。                                                                                               |
|              |  建议方案  | 1. 更新存储磁盘，重启创建容器；                                                                                |
|              |  关键日志  | Sep 18 12:19:24 localhost /usr/sbin/irqbalance[3091]: Cannot change IRQ 1145 affinity: No space left on device |
+--------------+------------+----------------------------------------------------------------------------------------------------------------+
The diag job is complete.

Tip: to output the full diagnosis JSON, add --json after the command
```

> [!NOTE]
>
> - 同一任务正在诊断中再次执行时，会提示`job=job-x is being diagnosed (started at HH:MM:SS), please do not run it again`，不会重复采集。
> - 若任务不存在（任务CR已删除且relcache中无该任务pod记录），会直接返回`Training/inference task not found: job=job-x in ns=training`并终止诊断；任务CR已删除但relcache中仍有该任务pod记录（删除TTL内，日志仍保留在节点/共享盘上），仍会继续诊断。
> - 诊断结果缓存的详细说明请参见[诊断结果缓存](../../06_api/20_clusterops_agent.md#诊断结果缓存)章节。

## 数据落盘与空间回收<a name="sectionfaultdiagnosisstorage"></a>

Agent Core与Node Collector分别在各自的工作目录下按任务维度落盘采集与诊断产物，具体路径如下：

| 组件 | 工作目录 | 说明 |
|---|---|---|
| Agent Core | `/user/clusterops/agent-core` | 诊断输入、诊断输出及结果缓存 |
| Node Collector | `/user/clusterops/node-collector` | 宿主机日志采集与清洗产物 |

### Agent Core落盘内容<a name="sectionfaultdiagnosisagentstorage"></a>

```text
/user/clusterops/agent-core/
├── {YYYYMMDD}/
│   └── {namespace}_{job}/
│       ├── parse-result-{job}-{node}.tar.gz          # 各节点上报的采集归档
│       ├── diag-input/
│       │   └── {host_ip}/                             # 各节点清洗产物，以机器IP命名（含 server-info.json 等）
│       └── diag-output/
│           └── fault_diag_result/
│               └── diag_report.json                   # 集中诊断报告
└── cache/
    └── {namespace}_{job}.json                         # 诊断结果缓存
```

- `parse-result-{job}-{node}.tar.gz`：Node Collector按节点上报的采集归档。
- `diag-input/{host_ip}/`：各节点 `ascend-fd parse` 清洗产物，以机器IP命名（host_ip缺失时回退`worker{N}`），组装为集中诊断输入。
- `diag-output/`：`ascend-fd diag` 生成的诊断报告目录。
- `cache/{namespace}_{job}.json`：每次诊断成功后缓存的诊断结果，重复诊断直接返回，支持`--refresh`强制刷新。

### Node Collector落盘内容<a name="sectionfaultdiagnosiscollectorstorage"></a>

```text
/user/clusterops/node-collector/
└── {YYYYMMDD}/
    └── {namespace}_{job}/
        ├── collect/                                   # 从宿主机采集的原始日志
        └── parse-output/                              # ascend-fd parse 清洗产物
```

- `collect/`：按采集契约从宿主机采集的原始日志（plog、系统日志、设备日志等）。
- `parse-output/`：本地 `ascend-fd parse` 清洗后的产物，上传后作为诊断输入。

### 保留策略<a name="sectionfaultdiagnosisretention"></a>

两个组件的工作目录默认限制为 **10GB**。

- Agent Core在每次诊断结束后、Node Collector在每次采集上传结束后触发空间回收。
- 当工作目录总占用超过阈值时，按各job目录及缓存文件的修改时间从旧到新删除，直至总占用降到阈值以内。
- 删除粒度为job级别：Agent Core删除整个 `{YYYYMMDD}/{namespace}_{job}` 目录及对应的 `{namespace}_{job}.json` 缓存文件；Node Collector删除整个 `{YYYYMMDD}/{namespace}_{job}` 目录。

> [!NOTE]
>
> - 诊断结果缓存与诊断产物统一纳入Agent Core的10GB配额管理，超限时同样按最旧优先删除。
> - 缓存目录默认位于 `/user/clusterops/agent-core/cache/`。
