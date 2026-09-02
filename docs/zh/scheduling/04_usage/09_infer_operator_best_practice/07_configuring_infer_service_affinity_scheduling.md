# 配置推理服务亲和性调度

在超节点场景下部署Infer Operator推理服务时，Infer Operator支持自动为同一推理服务下的所有实例配置推理服务亲和性标签（inferserviceid），使Ascend-for-volcano调度插件优先将同一推理服务的全部实例调度至同一框或同一超节点内，充分发挥网络优势，整体提升推理服务运行性能。推理服务亲和性调度策略的详细说明请参见[推理服务亲和性调度](../03_basic_scheduling/01_affinity_scheduling/06_infer_affinity_scheduling.md)。

## 前提条件

- 已完成Infer Operator服务部署，详细请参见[部署Infer Operator任务](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md)。

## 支持的产品形态

- Atlas 950 SuperPoD 超节点
- Atlas 850E 超节点
- Atlas 900 A3 SuperPoD 超节点
- Atlas 9000 A3 SuperPoD 集群算力系统

## 亲和性标签自动配置原理

1. 用户在InferServiceSet任务YAML的角色（role）中将`huawei.com/schedule_policy`注解配置为超节点调度策略（取值以-sp结尾，例如chip8-node8-sp、chip8-node16-sp、chip8-node8-ra64-sp、chip2-node16-sp、chip2-node8-sp）。
2. Infer Operator创建InstanceSet时，检测到超节点调度策略后，自动为该推理服务下所有角色的InstanceSet注入`inferserviceid`标签，取值为当前推理服务（InferService）的唯一标识。同一推理服务下所有实例的`inferserviceid`取值相同。
3. `inferserviceid`标签随InstanceSet传递至对应的PodGroup与工作负载，Ascend-for-volcano调度插件基于该标签将同一推理服务的实例优先调度至同一框（Atlas 950 SuperPoD 超节点）或同一超节点（Atlas 850E 超节点、Atlas 900 A3 SuperPoD 超节点），资源不足时回退至其他超节点。

若用户已在角色metadata的labels中手动配置`inferserviceid`，则以用户配置为准，Infer Operator不会覆盖。

## 配置推理服务亲和性调度

Infer Operator任务配置推理服务亲和性调度示例如下，需修改以下加粗部分配置。相关配置项说明请参见[YAML参数说明](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#yaml参数说明)。

<pre codetype="yaml">
apiVersion: mindcluster.huawei.com/v1
kind: InferServiceSet
metadata:
  name: "my-test"
  namespace: default
spec:
  replicas: 1 # 推理服务副本数
  template:
    roles:
    - name: prefill # prefill定义
      replicas: 1   # prefill副本数
      workload:     # prefill中实例的CRD类型信息
        apiVersion: apps/v1
        kind: StatefulSet # workload类型，当前支持StatefulSet/Deployment
      metadata:
        labels:
          <strong>infer.huawei.com/gang-schedule: 'true' # 亲和性调度场景要求开启gang调度</strong>
      spec:
        replicas: 1 # prefill中workload的pod副本数
        <strong>podManagementPolicy: Parallel # workload为StatefulSet且开启gang调度时，需配置为Parallel</strong>
        selector:
          matchLabels:
            app: test-prefill # 用户自定义，需要与下面labels中app配置保持一致
        template:
          metadata:
            labels:
              app: test-prefill # 用户自定义，需要与下面labels中app配置保持一致
              fault-scheduling: 'grace' # 开启重调度
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # 标识产品类型
            <strong>annotations:</strong>
              <strong>sp-block: "8" # 指定逻辑超节点芯片数量，设置成该实例请求的NPU总数</strong>
              <strong>huawei.com/schedule_policy: chip8-node8-sp # 超节点调度策略，取值以-sp结尾，根据硬件形态设置</strong>
          spec:
            schedulerName: volcano # 指定调度器为Volcano
            containers:
            - name: prefill
              image: vllm-ascend:xxx # 自定义vllm镜像名
              ...
              resources:
                requests:
                  huawei.com/npu: 8
                limits:
                  huawei.com/npu: 8
              ... # 补充容器必要的挂载项与运行命令
    - name: decode  # decode定义
      replicas: 1   # decode副本数
      workload:     # decode中实例的CRD类型信息
        apiVersion: apps/v1
        kind: StatefulSet # workload类型，当前支持StatefulSet/Deployment
      metadata:
        labels:
          <strong>infer.huawei.com/gang-schedule: 'true' # 亲和性调度场景要求开启gang调度</strong>
      spec:
        replicas: 1 # decode中workload的pod副本数
        <strong>podManagementPolicy: Parallel # workload为StatefulSet且开启gang调度时，需配置为Parallel</strong>
        selector:
          matchLabels:
            app: test-decode # 用户自定义，需要与下面labels中app配置保持一致
        template:
          metadata:
            labels:
              app: test-decode # 用户自定义，需要与下面labels中app配置保持一致
              fault-scheduling: 'grace' # 开启重调度
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # 标识产品类型
            <strong>annotations:</strong>
              <strong>sp-block: "8" # 指定逻辑超节点芯片数量，设置成该实例请求的NPU总数</strong>
              <strong>huawei.com/schedule_policy: chip8-node8-sp # 超节点调度策略，取值以-sp结尾，根据硬件形态设置</strong>
          spec:
            schedulerName: volcano # 指定调度器为Volcano
            containers:
            - name: decode
              image: vllm-ascend:xxx # 自定义vllm镜像名
              ...
              resources:
                requests:
                  huawei.com/npu: 8
                limits:
                  huawei.com/npu: 8
              ... # 补充容器必要的挂载项与运行命令
</pre>

任务下发后，Infer Operator自动为Prefill与Decode角色的实例配置相同的`inferserviceid`标签，可通过以下命令查看自动注入的标签。其中，`<namespace>`为用户定义的命名空间，`<instance-set-name>`为实例集名称。

```shell
kubectl get instanceset <instance-set-name> -n <namespace> -o jsonpath='{.metadata.labels.inferserviceid}'
```

>[!NOTE]
>
>- 推理服务亲和性调度依赖PodGroup将`inferserviceid`标签传递给Volcano调度器，因此需将`infer.huawei.com/gang-schedule`配置为'true'以开启gang调度。
>- `huawei.com/schedule_policy`与`sp-block`注解需配置在角色的metadata.annotations中（如上示例所示），配置在其他位置时不会触发亲和性标签的自动注入。
>- 若不希望开启推理服务亲和性调度，可将`huawei.com/schedule_policy`配置为非超节点调度策略（如chip8-node8），或在角色metadata的labels中自行配置`inferserviceid`。
