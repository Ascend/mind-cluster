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
[Root Cause] HCCS链路异常，导致集合通信超时
Root cause device: node-3:3
Note: 任务启动后设备间通信持续超时，多节点日志同步报错
[1] Fault: 940000010
    Type: hw npu hbm
    Name: HBM单粒子翻转（SEU）
    Desc: 检测到HBM错误并触发复位
    Suggestion: 1. 检查故障设备所在节点硬件状态
    Log: [ERROR] HBM ...
```

使用`--json`参数可查看完整的JSON诊断响应（包含采集到的Pod信息、诊断报告等）：

```shell
kubectl ascend_diag --job job-x -n training --json
```

## 重复执行与缓存提示<a name="sectionfaultdiagnosisrepeat"></a>

同一任务在**10秒内**重复执行诊断时，会直接返回最近一次诊断结果（缓存数据），并提示使用`--refresh`获取实时数据：

```text
$ kubectl ascend_diag --job job-x -n training
这是缓存数据，需要实时数据，请加 --refresh
[Root Cause] HCCS链路异常，导致集合通信超时
```

> [!NOTE]
>
> - 同一任务正在诊断中再次执行时，会提示`job=job-x正在诊断中(开始于HH:MM:SS), 请勿重复执行`，不会重复采集。
> - 若任务不存在，会直接返回`训练/推理任务不存在: job=job-x in ns=training`并终止诊断。
> - 诊断结果缓存的详细说明请参见[诊断结果缓存](../../06_api/18_clusterops_agent.md#诊断结果缓存)章节。

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
│       │   └── worker{N}/                             # 各节点清洗产物（含 server-info.json 等）
│       └── diag-output/
│           └── fault_diag_result/
│               └── diag_report.json                   # 集中诊断报告
└── cache/
    └── {namespace}_{job}.json                         # 诊断结果缓存
```

- `parse-result-{job}-{node}.tar.gz`：Node Collector按节点上报的采集归档。
- `diag-input/worker{N}/`：各节点 `ascend-fd parse` 清洗产物，组装为集中诊断输入。
- `diag-output/`：`ascend-fd diag` 生成的诊断报告目录。
- `cache/{namespace}_{job}.json`：任务停止后缓存的诊断结果，用于后续查询直接返回。

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
