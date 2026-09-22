# 配置推理任务Pod级重调度

当Infer Operator推理任务中出现节点、芯片或其他故障时，MindCluster集群调度组件可以对故障资源进行隔离并自动进行重调度。如需了解故障的检测原理，请参见[故障检测](../../04_fault_detection_and_diagnosis/01_working_principle.md#故障检测整体架构)章节。

与实例级重调度（重调度故障Pod对应的整个实例）不同，Pod级重调度在故障发生时仅删除并重新调度发生故障的Pod，实例内其他正常的Pod不受影响。Pod级重调度适用于业务面故障（Pod内进程非零退出）场景，硬件故障时不删除故障Pod，允许业务继续使用故障Pod内剩余可用的NPU卡运行。实例级重调度的配置方式请参见[配置推理任务实例级重调度](../../10_infer_operator_best_practice/04_configuring_rescheduling_inference_job.md)。

## 前提条件

已完成Infer Operator服务部署，详细请参见[部署Infer Operator任务](../../10_infer_operator_best_practice/01_deploying_infer_operator_inference_job_with_vllm_proxy.md)。

## 重调度原理

Infer Operator在部署不同角色的实例时，会创建Deployment/StatefulSet（具体类型取决于配置项workload）对应实例。开启Pod级重调度后，当某个Pod发生业务面故障时，Infer Operator仅删除故障Pod，由对应的Controller重新创建被删除的Pod，并由Volcano重新调度新创建Pod。

以示例yaml为例，其中各配置项含义如下：

- `fault-scheduling: external-force-pod-failed`：仅处理业务面故障（Pod内进程非零退出），硬件故障时不删除故障节点Pod。
- `pod-rescheduling: on`：取值为on时按Pod粒度删除故障Pod；取值为其他值时按实例粒度删除。
- `fault-retry-times`：业务面故障的重试次数。

## 配置Pod级重调度

Infer Operator任务配置Pod级重调度示例如下，需修改以下加粗部分配置。相关配置项说明请参见[YAML参数说明](../../10_infer_operator_best_practice/01_deploying_infer_operator_inference_job_with_vllm_proxy.md#YAML参数说明)。

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
            infer.huawei.com/gang-schedule: 'true' # 开启gang调度。开启时会为每一个workload实例创建PodGroup
        spec:
          replicas: 2 # prefill中workload的pod副本数
          podManagementPolicy: Parallel # 当workload为StatefulSet且infer.huawei.com/gang-schedule为true时，需配置为Parallel
          selector:
            matchLabels:
              app: test-prefill # 用户自定义，需要与下面labels中app配置保持一致
          template:
            metadata:
              labels:
                app: test-prefill # 用户自定义，需要与下面labels中app配置保持一致
                <strong>fault-scheduling: 'external-force-pod-failed' # 开启重调度，仅处理业务面故障（Pod内进程非零退出）</strong>
                <strong>pod-rescheduling: 'on' # 开启Pod级重调度</strong>
                <strong>fault-retry-times: '10' # 业务面故障的重试次数</strong>
                ring-controller.atlas: ascend-910b # 标识产品类型
              annotations:
                huawei.com/schedule_policy: chip8-node8 # 根据硬件形态设置
            spec:
              schedulerName: volcano # 指定调度器为Volcano
              containers:
                - name: prefill
                  image: vllm-ascend:xxx # 自定义vllm镜像名
                  imagePullPolicy: IfNotPresent
                  command: [...]
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
        metadata:
          labels:
            infer.huawei.com/gang-schedule: 'true' # 开启gang调度。开启时会为每一个workload实例创建PodGroup
        spec:
          replicas: 1 # decode中workload的pod副本数
          podManagementPolicy: Parallel # 当workload为StatefulSet且infer.huawei.com/gang-schedule为true时，需配置为Parallel
          selector:
            matchLabels:
              app: test-decode # 用户自定义，需要与下面labels中app配置保持一致
          template:
            metadata:
              labels:
                app: test-decode # 用户自定义，需要与下面labels中app配置保持一致
                <strong>fault-scheduling: 'external-force-pod-failed' # 开启重调度，仅处理业务面故障（Pod内进程非零退出）</strong>
                <strong>pod-rescheduling: 'on' # 开启Pod级重调度</strong>
                <strong>fault-retry-times: '10'</strong>
                ring-controller.atlas: ascend-910b # 标识产品类型
              annotations:
                huawei.com/schedule_policy: chip8-node8 # 根据硬件形态设置
            spec:
              schedulerName: volcano # 指定调度器为Volcano
              containers:
                - name: decode
                  image: vllm-ascend:xxx # 自定义vllm镜像名
                  imagePullPolicy: IfNotPresent
                  command: [...]
                  resources:
                    requests:
                      huawei.com/Ascend910: 8
                    limits:
                      huawei.com/Ascend910: 8
                  ... # 补充容器必要的挂载项与运行命令
</pre>
