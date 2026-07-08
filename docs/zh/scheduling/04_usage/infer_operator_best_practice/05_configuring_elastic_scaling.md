# 配置基于负载的弹性扩缩容

Infer Operator 支持给推理实例配置弹性扩缩容策略，从而实现基于推理实例的负载情况自动调整推理实例数量的功能。

## 前置准备

1. 已完成Infer Operator的[安装部署](../../05_developer_guide/installation_deployment/manual_installation/07_infer_operator.md)。
2. 如需配置扩缩容指标类型为External，需先实现并部署相应的External Metrics Adaptor，该Adaptor需要提供推理实例的负载指标（例如请求队列长度、请求处理时间等），可参考[示例](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/infer-operator-metrics-adaptor)进行实现。若为[基于MindIE PyMotor部署Infer Operator推理任务](./02_deploying_infer_operator_inference_job_with_mindie_pymotor.md)场景，可直接部署使用该示例提供的Metrics Adaptor。

## 弹性扩缩容原理

Infer Operator 会根据推理实例的弹性扩缩容配置，为对应实例创建相应的扩缩容控制器资源（例如Horizontal Pod Autoscaler（HPA）），由扩缩容控制器根据实例的负载情况，自动调整实例的期望数量。

## 配置弹性扩缩容策略

给推理实例配置弹性扩缩容策略的示例如下，需添加以下加粗部分配置。相关配置项说明请参见[YAML参数说明](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#yaml参数说明)。

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
      <strong>scalingPolicy:
        type: HPA # 弹性扩缩容策略类型，当前支持HPA
        spec: # HPA配置
          minReplicas: 1 # 缩容下限
          maxReplicas: 4 # 扩容上限
          metrics: # HPA扩缩容指标配置列表
          - type: External # 指标类型：外部自定义指标（由External Metrics Adapter提供）
            external:
              metric:
                name: num_requests_waiting # 外部指标名称
              target: # 目标值配置
                type: AverageValue
                averageValue: "5"
          ... # 其他HPA配置项，根据需要添加，需符合HPA配置规范</strong>
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
              fault-scheduling: 'grace' # 开启重调度
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # 标识产品类型
            annotations:
              huawei.com/schedule_policy: chip8-node8 # 根据硬件形态设置
          spec:
            schedulerName: volcano # 指定调度器为Volcano
            nodeSelector:
              example-key: example-value    # 示例值，用户可根据调度意图自行配置nodeSelector
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
      workload:     # decode中实例的CRD类型信息
        apiVersion: apps/v1
        kind: StatefulSet # workload类型，当前支持StatefulSet/Deployment
      <strong>scalingPolicy:
        type: HPA # 弹性扩缩容策略类型，当前支持HPA
        spec: # HPA配置
          minReplicas: 1 # 缩容下限
          maxReplicas: 4 # 扩容上限
          metrics: # HPA扩缩容指标配置列表
          - type: External # 指标类型：外部自定义指标（由External Metrics Adapter提供）
            external:
              metric:
                name: generation_tokens_per_second # 外部指标名称
              target: # 目标值配置
                type: AverageValue
                averageValue: "10"
          ... # 其他HPA配置项，根据需要添加，需符合HPA配置规范</strong>
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
              fault-scheduling: 'grace' # 开启重调度
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
