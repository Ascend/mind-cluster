# 亲和性调度对接说明<a name="ZH-CN_TOPIC_0000002516224533"></a>

为了实现调度层与任务资源类型解耦，Ascend-for-volcano调度插件新增支持Pod级别调度策略的配置。用户可直接在Pod的metadata.labels或metadata.annotations中配置调度相关参数，无需依赖PodGroup，支持acjob、vcjob、Job、Deployment、StatefulSet等所有Pod类型。

## 功能介绍<a name="section112161354155714"></a>

通过在K8s资源的Pod模板中添加特定Label或Annotation，可控制Volcano的核心调度行为，包括但不限于：

- 昇腾AI处理器的亲和性调度
- 交换机亲和性调度
- 逻辑超节点亲和性调度
- 故障重调度

## 前提条件<a name="section46282421720"></a>

确保Kubernetes集群已经正确部署并配置了Volcano调度器，并且相关的调度插件Ascend-for-volcano已启用。

## 调度策略配置示例<a name="section5997169155814"></a>

以StatefulSet为例，所有调度相关的labels/annotations均需配置在StatefulSet.spec.template.metadata下，确保调度器可以从Pod实例中正确读取。

<pre codetype="yaml">
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mindx-dls-test               # The value of this parameter must be consistent with the name of ConfigMap.
  labels:
    app: mindspore
    ring-controller.atlas: ascend-910
spec:
  replicas: 16                        # The value of replicas is 1 in a single-node scenario and N in an N-node scenario. The number of NPUs in the requests field is 8 in an N-node scenario.
  <strong>podManagementPolicy: Parallel   # 支持OrderedReady和Parallel两种模式。“OrderedReady”仅支持节点内亲和调度并且huawei.com/schedule_minAvailable只能为1。“Parallel”支持节点内和节点间亲和调度</strong>
  serviceName: service-headliness
  selector:
    matchLabels:
      app: mindspore
  <strong>template:</strong>
    <strong>metadata:</strong>
      <strong>labels:</strong>
        app: mindspore
        ring-controller.atlas: ascend-910
        <strong>fault-scheduling: force   # 故障重调度功能开关</strong>
        <strong>pod-rescheduling: "on"   # Pod级别重调度功能开关</strong>
        <strong>fault-retry-times: "85"    # 业务面故障重调度次数</strong>
        <strong>tor-affinity: large-model-schema  # 交换机亲和性调度开关</strong>
        <strong>deploy-name: mindx-dls-test # 生成rankTable必须增加该标签，取值和任务名称保持一致</strong>
      <strong>annotations:</strong>
        <strong>sp-block: "128"         # 逻辑超节点亲和性调度开关</strong>
        <strong>huawei.com/recover_policy_path: pod    # Pod级别重调度不升级为Job级开关（当使用vcjob时，需要配置该策略：policies: -event:PodFailed -action:RestartTask）</strong>
        <strong>huawei.com/schedule_minAvailable: "16"  # 任务调度的最小副本数，建议与任务副本数保持一致</strong>
        <strong>huawei.com/skip-ascend-plugin: "enabled"    # 开启后将允许一些特殊任务（如不需要NPU资源的任务）绕过Ascend-for-volcano的默认检查逻辑</strong>
    spec:
      schedulerName: volcano         # Use the Volcano scheduler to schedule jobs.
      nodeSelector:
        host-arch: huawei-arm        # Configure the label based on the actual job.
      containers:
        - image: ubuntu:18.04      # Training framework image, which can be modified.
          name: mindspore
          resources:
            requests:
              huawei.com/Ascend910: 16                                               # Number of required NPUs. The maximum value is 16. You can add lines below to configure resources such as memory and CPU
            limits:
              huawei.com/Ascend910: 16                                                # The value must be consistent with that in requests.</pre>

> [!NOTE] 
>
>- 如果一个PodGroup被创建，则spec中的调度配置将覆盖其生成的Pod上的labels/annotations配置。
>- 对于可以生成PodGroup的资源，在PodGroup上配置对应的调度策略也可以实现亲和性调度能力。
>- 常用的Label与Annotation对照表请参见[PodGroup](../../../api/volcano.md#podgroup)/[Pod](../../../api/volcano.md#pod)。
