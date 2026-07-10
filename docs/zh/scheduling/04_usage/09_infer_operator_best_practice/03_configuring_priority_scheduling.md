# 配置推理任务优先级调度

Infer Operator支持配置多角色的优先级大小，以确保高优先级角色优先被创建并调度资源。

## 前提条件

已完成Infer Operator服务部署，详细请参见[部署Infer Operator任务](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md)。

## 优先级调度原理

Infer Operator在部署不同角色的实例时，会为每一个实例创建Deployment/StatefulSet（具体类型取决于配置项workload）。当开启优先级调度特性，并配置各角色优先级大小后，Infer Operator会基于各角色的优先级，从高到低依次创建对应实例。

当需要适配其他服务化平台（例如：MindIE PyMotor）开启缩P保D特性时，需要同时配置优先级调度与实例级重调度。

## 配置优先级调度

Infer operator任务配置实例级重调度示例如下，需修改以下加粗部分配置。相关配置项说明请参见[YAML参数说明](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#yaml参数说明)。
<pre codetype="yaml">
apiVersion: mindcluster.huawei.com/v1
kind: InferServiceSet
metadata:
  name: "my-test"
  namespace: default
spec:
  replicas: 1 # 推理服务副本数
  template:
    <strong>schedulingStrategy: # 将type配置为Priority开启优先级调度
      type: Priority</strong>
    roles:
    - name: prefill # prefill定义
      replicas: 1   # prefill副本数
      <strong>priority: 2   # prefill的优先级配置，值越小优先级越高，仅开启优先级调度时生效</strong>
      workload:     # prefill中实例的CRD类型信息
        apiVersion: apps/v1
        kind: StatefulSet # workload类型，当前支持StatefulSet/Deployment
      metadata:
        labels:
          infer.huawei.com/gang-schedule: 'false' # 关闭gang调度，开启时会为每一个workload实例创建PodGroup
      spec:
        replicas: 1 # prefill中workload的pod副本数
        podManagementPolicy: Parallel # 此配置可不填，当workload为StatefulSet，且infer.huawei.com/gang-schedule为true时，需配置为Parallel
        selector:
          matchLabels:
            app: test-prefill # 用户自定义，需要与下面labels中app配置保持一致
        template:
          metadata:
            labels:
              app: test-prefill # 用户自定义，需要与下面labels中app配置保持一致
              fault-scheduling: 'external-force' # 开启实例级重调度
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # 标识产品类型
            annotations:
              huawei.com/schedule_policy: chip8-node8 # 根据硬件形态设置
          spec:
            schedulerName: volcano # 指定调度器为Volcano
            containers:
            - name: prefill
              image: vllm-ascend:xxx # 自定义vllm镜像名
              ...
              resources:
                requests:
                  huawei.com/Ascend910: 8
                limits:
                  huawei.com/Ascend910: 8
              ... # 补充容器必要的挂载项与运行命令
    - name: decode  # decode定义
      replicas: 1   # decode副本数
      <strong>priority: 1   # decode的优先级配置，值越小优先级越高，仅开启优先级调度时生效</strong>
      workload:     # decode中实例的CRD类型信息
        apiVersion: apps/v1
        kind: StatefulSet # workload类型，当前支持StatefulSet/Deployment
      metadata:
        labels:
          infer.huawei.com/gang-schedule: 'false' # 关闭gang调度，开启时会为每一个workload实例创建PodGroup
      spec:
        replicas: 1 # decode中workload的pod副本数
        podManagementPolicy: Parallel # 此配置可不填，当workload为StatefulSet，且infer.huawei.com/gang-schedule为true时，需配置为Parallel
        selector:
          matchLabels:
            app: test-decode # 用户自定义，需要与下面labels中app配置保持一致
        template:
          metadata:
            labels:
              app: test-decode # 用户自定义，需要与下面labels中app配置保持一致
              fault-scheduling: 'external-force' # 开启实例级重调度
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # 标识产品类型
            annotations:
              huawei.com/schedule_policy: chip8-node8 # 根据硬件形态设置
          spec:
            schedulerName: volcano # 指定调度器为Volcano
            containers:
            - name: decode
              image: vllm-ascend:xxx # 自定义vllm镜像名
              ...
              resources:
                requests:
                  huawei.com/Ascend910: 8
                limits:
                  huawei.com/Ascend910: 8
              ... # 补充容器必要的挂载项与运行命令
    - name: router  # router定义
      replicas: 1   # router副本数
      <strong>priority: 3   # router的优先级配置，值越小优先级越高，仅开启优先级调度时生效</strong>
      services:     # router services定义，此处定义的service在一个角色范围内仅创建一个
      - name: vllm-router-service
        spec:
          ports:    # service的端口定义
          - port: 1026
            protocol: TCP
            targetPort: 1026
          selector:
            app: test-router # 用户自定义，需要与下面labels中app配置保持一致
          type: ClusterIP
      workload:     # router中实例的CRD类型信息
        apiVersion: apps/v1
        kind: Deployment # workload类型，当前支持StatefulSet/Deployment
      spec:
        replicas: 1 # router中workload的pod副本数
        selector:
          matchLabels:
            app: test-router # 用户自定义，需要与下面labels中app配置保持一致
        template:
          metadata:
            labels:
              app: test-router # 用户自定义，需要与下面labels中app配置保持一致
          spec:
            schedulerName: volcano # 指定调度器为Volcano
            containers:
            - name: router
              image: xxx:yyy # 自定义镜像名
              ... # 补充容器必要的挂载项与运行命令
</pre>
