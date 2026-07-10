# 推理服务亲和性调度

在推理业务场景下，单个推理服务通常由多个推理实例构成，包含若干 Prefill 实例与若干 Decode 实例。
对于 Atlas 950 SuperPoD产品，同框内节点间网络通信时延最低、吞吐性能最优，同一超节点内不同框的节点之间通信性能次之；对于 Atlas 850 Server超节点，同一超节点内节点之间的网络通信表现最佳。
基于上述网络特性，Ascend-for-volcano调度插件新增支持配置推理服务亲和性调度策略：对于Atlas 950 SuperPoD产品，优先将同一推理服务的全部实例调度至同一框内，无法满足时，优先将同服务实例调度至同一超节点；对于Atlas 850 Server超节点，优先将同服务实例调度至同一超节点，以此充分发挥网络优势，整体提升推理服务运行性能。

## 前置条件

确保Kubernetes集群已经正确部署并配置了Volcano调度器，并且相关的调度插件Ascend-for-volcano已启用。

## 配置推理服务亲和性调度策略

通过给K8s资源添加特定Label与Annotation，即可配置推理服务亲和性调度策略。

对于Atlas 950 SuperPoD产品，以Deployment资源为例，需要在其Pod模板的labels与annotations中添加如下示例中加粗部分的内容：

<pre codetype="yaml">
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vllm-prefill-0
  labels:
    app: vllm-prefill-0
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vllm-prefill-0
  template:
    metadata:
      labels:
        app: vllm-prefill-0
        <strong>inferserviceid: vllm-test # 推理服务ID，表征当前实例属于哪个推理服务</strong>
        host-arch: huawei-arm
        ring-controller.atlas: ascend-npu
      annotations:
        sp-block: "8" # 指定逻辑超节点芯片数量，设置成请求的NPU数量
        ra-block: "8" # 指定逻辑框大小，设置成请求的NPU数量
        huawei.com/schedule_policy: "chip8-node8-ra64-sp" # Atlas 950 SuperPoD产品对应的调度策略
    spec:
      schedulerName: volcano
      automountServiceAccountToken: false
      nodeSelector:
        host-arch: huawei-arm
      containers:
      - image: ubuntu:22.04
        imagePullPolicy: IfNotPresent
        name: prefill
        command: ["/bin/bash", "-c", "sleep 3000"]
        env:
        - name: ASCEND_VISIBLE_DEVICES
          valueFrom:
            fieldRef:
              fieldPath: metadata.annotations['huawei.com/npu']
        resources:
          requests:
            huawei.com/npu: 8 # 请求的NPU数量
          limits:
            huawei.com/npu: 8 # 需要和requests保持一致
        volumeMounts:
        - name: slog
          mountPath: /var/log/npu/conf/slog/
        - name: localtime
          mountPath: /etc/localtime
      volumes:
      - name: slog
        hostPath:
          path: /var/log/npu/conf/slog/
      - name: localtime
        hostPath:
          path: /etc/localtime
</pre>

对于Atlas 850 Server超节点，以Deployment资源为例，配置推理服务亲和性调度策略与Atlas 950 SuperPoD产品类似，添加如下示例中加粗部分的内容：

<pre codetype="yaml">
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vllm-prefill-0
  labels:
    app: vllm-prefill-0
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vllm-prefill-0
  template:
    metadata:
      labels:
        app: vllm-prefill-0
        <strong>inferserviceid: vllm-test # 推理服务ID，表征当前实例属于哪个推理服务</strong>
        host-arch: huawei-arm
        ring-controller.atlas: ascend-npu
      annotations:
        sp-block: "8" # 指定逻辑超节点芯片数量，设置成请求的NPU数量
        huawei.com/schedule_policy: "chip8-node8-sp" # Atlas 850 Server超节点对应的调度策略
    spec:
      schedulerName: volcano
      automountServiceAccountToken: false
      nodeSelector:
        host-arch: huawei-arm
      containers:
      - image: ubuntu:22.04
        imagePullPolicy: IfNotPresent
        name: prefill
        command: ["/bin/bash", "-c", "sleep 3000"]
        env:
        - name: ASCEND_VISIBLE_DEVICES
          valueFrom:
            fieldRef:
              fieldPath: metadata.annotations['huawei.com/npu']
        resources:
          requests:
            huawei.com/npu: 8 # 请求的NPU数量
          limits:
            huawei.com/npu: 8 # 需要和requests保持一致
        volumeMounts:
        - name: slog
          mountPath: /var/log/npu/conf/slog/
        - name: localtime
          mountPath: /etc/localtime
      volumes:
      - name: slog
        hostPath:
          path: /var/log/npu/conf/slog/
      - name: localtime
        hostPath:
          path: /etc/localtime
</pre>

> [!NOTE]
>
>- 推理服务亲和性调度策略仅支持Atlas 950 SuperPoD产品与Atlas 850 Server超节点。
>- 使用其他类型的k8s资源部署推理服务示例请参见[推理任务类型与硬件型号对应YAML文件](../03_full_npu_scheduling.md#准备任务yaml)，添加对应的Label与Annotation即可开启推理服务亲和性调度策略。
>- 对于可以生成PodGroup的资源，在PodGroup上添加相应字段也可以实现推理服务亲和性调度。
>- 常用的Label与Annotation对照表请参见[PodGroup](../../../06_api/01_volcano.md#podgroup)/[Pod](../../../06_api/01_volcano.md#pod)。
