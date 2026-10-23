# Configuring Job Types<a name="ZH-CN_TOPIC_00000026faultdiagnosis02"></a>

Agent Core configures the list of CR GVKs that can be detected through the ConfigMap `agent-core-task-crds` (in the `cluster-system` namespace). When Agent Core starts, it reads this ConfigMap once through the K8s API and tracks only the Pods managed by these CRs (used for job existence determination and Pod collection scheduling).

## Default Supported Job Types<a name="sectionfaultdiagnosisdefaultcr"></a>

Agent Core supports the following job types by default, without any configuration:

| Type | CRD | Resource Object |
|---|---|---|
| acjob (AscendJob training job) | mindxdl.gitee.com/v1 | AscendJob |
| inferjob (InferServiceSet inference job) | mindcluster.huawei.com/v1 | InferServiceSet |

> [!NOTE]
> If you use the preceding two job types, skip this section and go to [Configuring Log Collection](./02_configuring_log_collection.md).

**Step 1: Modify the ConfigMap `agent-core-task-crds` in `agent-core.yaml`**

Add the `api_version` and `kind` of the CR to the `task_crds` list of the ConfigMap `agent-core-task-crds` inlined at the top of `agent-core.yaml`. The following example adds a Volcano Job (`batch.volcano.sh/v1alpha1`, kind `Job`). The modified `data.task_crds.yaml` content is as follows:

```yaml
# data.task_crds.yaml of the ConfigMap agent-core-task-crds in agent-core.yaml
task_crds:
  - api_version: mindxdl.gitee.com/v1   # AscendJob
    kind: AscendJob
  - api_version: mindcluster.huawei.com/v1  # InferServiceSet (inference job)
    kind: InferServiceSet
  - api_version: batch.volcano.sh/v1alpha1  # Volcano Job (newly added)
    kind: Job
```

**Step 2: Restart Agent Core for the configuration to take effect**

```shell
kubectl rollout restart deployment agent-core -n mindx-dl
```

> [!NOTE]
> After adding a job type, the ClusterRole does not need to be modified. Both job existence checks and Pod collection scheduling rely on relcache (the central relationship cache, which filters job Pods by `agent-core-task-crds`). Agent Core does not query CRs directly.

## Verifying the Configuration<a name="sectionfaultdiagnosisverify"></a>

After Agent Core starts, check the Agent Core logs to confirm that the list of detectable CRs contains the newly configured job type:

```shell
kubectl logs -f deployment/agent-core -n mindx-dl | grep "task CRs"
```

The following is an example log:

```text
INFO ... task CRs supported by agent-core: ["mindxdl.gitee.com/v1/AscendJob", "mindcluster.huawei.com/v1/InferServiceSet", "batch.volcano.sh/v1alpha1/Job"]
```
