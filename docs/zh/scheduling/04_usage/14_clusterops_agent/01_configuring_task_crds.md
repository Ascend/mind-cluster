# 配置任务类型<a name="ZH-CN_TOPIC_00000026faultdiagnosis02"></a>

Agent Core通过ConfigMap `agent-core-task-crds`（cluster-system命名空间）配置可检测的任务CR GVK清单。Agent Core启动时经K8s API读取一次该ConfigMap，仅跟踪这些任务CR所管理的Pod（用于任务存在性判定、任务状态判定和Pod采集调度）。

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

**步骤2：修改`agent-core.yaml`中的ClusterRole `agent-core`**

在`agent-core.yaml`的ClusterRole `agent-core`中新增对应API组的`get`权限。以上述Volcano Job为例：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: agent-core
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list"]
  - apiGroups: ["mindxdl.gitee.com"]
    resources: ["ascendjobs"]
    verbs: ["get"]
  - apiGroups: ["mindcluster.huawei.com"]
    resources: ["inferservicesets"]
    verbs: ["get"]
  - apiGroups: ["batch.volcano.sh"]      # 新增: Volcano Job
    resources: ["jobs"]
    verbs: ["get"]
```

> [!NOTE]
> 若任务CRD不止AscendJob和InferServiceSet，必须在ClusterRole中为该CR增补对应的`get`权限。否则Agent Core无法查询该任务的状态，任务状态判定为unknown，诊断结果不会被缓存（但诊断本身不受影响）。

**步骤3：重启Agent Core使配置生效**

```shell
kubectl rollout restart deployment agent-core -n mindx-dl
```

## 验证配置<a name="sectionfaultdiagnosisverify"></a>

Agent Core启动后，查看Agent Core日志，确认支持检测的任务CR清单已包含新配置的任务类型：

```shell
kubectl logs -f deployment/agent-core -n mindx-dl | grep "task CRs"
```

日志示例如下：

```text
INFO ... task CRs supported by agent-core: ["mindxdl.gitee.com/v1/AscendJob", "mindcluster.huawei.com/v1/InferServiceSet", "batch.volcano.sh/v1alpha1/Job"]
```
