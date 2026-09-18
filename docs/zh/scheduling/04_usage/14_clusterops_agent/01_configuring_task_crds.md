# 配置任务类型<a name="ZH-CN_TOPIC_00000026faultdiagnosis02"></a>

Agent Core通过ConfigMap `agent-core-task-crds`（cluster-system命名空间）配置可检测的任务CR GVK清单。Agent Core启动时经K8s API读取一次该ConfigMap，仅跟踪这些任务CR所管理的Pod（用于任务存在性判定和Pod采集调度）。

## 默认支持的任务类型<a name="sectionfaultdiagnosisdefaultcr"></a>

Agent Core默认支持以下任务类型，无需任何配置：

| 任务类型 | CRD | 资源对象 |
|---|---|---|
| acjob（AscendJob训练任务） | mindxdl.gitee.com/v1 | AscendJob |
| inferjob（InferServiceSet推理任务） | mindcluster.huawei.com/v1 | InferServiceSet |

> [!NOTE]
> 若使用上述两种任务类型，可直接跳过本章节，进入[配置日志采集](./02_configuring_log_collection.md)章节。

**步骤1：修改`agent-core.yaml`中的ConfigMap `agent-core-task-crds`**

在`agent-core.yaml`顶部内联的ConfigMap `agent-core-task-crds`的`task_crds`列表中新增任务CR的`api_version`和`kind`。以新增Volcano Job（`batch.volcano.sh/v1alpha1`，kind为`Job`）为例，修改后的`data.task_crds.yaml`内容：

```yaml
# agent-core.yaml中ConfigMap agent-core-task-crds的data.task_crds.yaml
task_crds:
  - api_version: mindxdl.gitee.com/v1   # AscendJob
    kind: AscendJob
  - api_version: mindcluster.huawei.com/v1  # InferServiceSet (推理任务)
    kind: InferServiceSet
  - api_version: batch.volcano.sh/v1alpha1  # Volcano Job (新增)
    kind: Job
```

**步骤2：重启Agent Core使配置生效**

```shell
kubectl rollout restart deployment agent-core -n mindx-dl
```

> [!NOTE]
> 新增任务类型后无需修改ClusterRole。任务存在性检查与Pod采集调度均依赖relcache（中心关系缓存，按`agent-core-task-crds`过滤任务Pod），Agent Core不直接查询任务CR。

## 验证配置<a name="sectionfaultdiagnosisverify"></a>

Agent Core启动后，查看Agent Core日志，确认支持检测的任务CR清单已包含新配置的任务类型：

```shell
kubectl logs -f deployment/agent-core -n mindx-dl | grep "task CRs"
```

日志示例如下：

```text
INFO ... task CRs supported by agent-core: ["mindxdl.gitee.com/v1/AscendJob", "mindcluster.huawei.com/v1/InferServiceSet", "batch.volcano.sh/v1alpha1/Job"]
```
