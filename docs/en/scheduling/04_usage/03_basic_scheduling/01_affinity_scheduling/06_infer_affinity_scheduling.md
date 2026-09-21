# Inference Service Affinity Scheduling

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-29T07:59:58.321Z pushedAt=2026-08-29T08:01:03.070Z -->

In inference service scenarios, a single inference service typically consists of multiple inference instances, including several Prefill instances and several Decode instances.
For the Atlas 950 SuperPoD, network communication between nodes within the same rack has the lowest latency and optimal throughput performance, while communication between nodes in different racks within the same SuperPoD is the next best; for the Atlas 850E SuperPoD, network communication between nodes within the same SuperPoD performs best.
Based on the preceding network characteristics, the Ascend-for-volcano scheduling plugin supports configuring inference service affinity scheduling policies: for the Atlas 950 SuperPoD, all instances of the same inference service are preferentially scheduled to the same rack, and when this cannot be satisfied, instances of the same service are preferentially scheduled to the same SuperPoD; for the Atlas 850E SuperPoD, instances of the same service are preferentially scheduled to the same SuperPoD, thereby fully leveraging the network advantages and improving the overall running performance of the inference service.

## Prerequisites

Ensure that the Kubernetes cluster has been properly deployed and configured with the Volcano scheduler, and that Ascend-for-volcano is enabled.

## Configuring the Inference Service Affinity Scheduling Policy

You can configure the inference service affinity scheduling policy by adding specific labels to Kubernetes resources.

For the Atlas 950 SuperPoD, taking a Deployment resource as an example, add the following content in bold to the labels of its Pod template:

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
        <strong>inferserviceid: vllm-test # Inference service ID, indicating which inference service the current instance belongs to.</strong>
        host-arch: huawei-arm
        ring-controller.atlas: ascend-npu
      annotations:
        sp-block: "8" # Specify the number of chips in the logical SuperPoD. Set it to the total count of NPUs requested by this instance.
        ra-block: "8" # Specify the logical rack size. Set it to the total count of NPUs requested by this instance.
        huawei.com/schedule_policy: "chip8-node8-ra64-sp" # Scheduling policy corresponding to the Atlas 950 SuperPoD.
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
            huawei.com/npu: 8 # Number of NPUs requested by a single Pod
          limits:
            huawei.com/npu: 8 # Keep consistent with requests
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

For the Atlas 850E SuperPoD, taking the Deployment resource as an example, the configuration of the inference service affinity scheduling policy is similar to that of the Atlas 950 SuperPoD. Add the following content in bold:

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
        <strong>inferserviceid: vllm-test # Inference service ID, indicating which inference service the current instance belongs to</strong>
        host-arch: huawei-arm
        ring-controller.atlas: ascend-npu
      annotations:
        sp-block: "8" # Specifies the number of chips in the logical SuperPoD. Set to the total count of NPUs requested by this instance
        huawei.com/schedule_policy: "chip8-node8-sp" # Scheduling policy corresponding to the Atlas 850E SuperPoD
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
            huawei.com/npu: 8 # Number of NPUs requested by a single Pod.
          limits:
            huawei.com/npu: 8 # Must be consistent with requests.
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
>- The inference service affinity scheduling policy supports only Atlas 950 SuperPoDs and Atlas 850E SuperPoDs.
>- For Atlas 950 SuperPoDs, if the inference service affinity scheduling feature is enabled, the current version mandates that a single instance must not be scheduled across racks. As a result, the following situation may occur: although the total idle node resources across multiple racks can satisfy the requirements of an instance, these idle nodes belong to different racks, causing the instance to remain in the `Pending` state because it cannot span racks. To enable successful scheduling of the instance, delete the `inferserviceid` label from `labels` to disable inference affinity, and change `huawei.com/schedule_policy` to `chip8-node8-sp`, thereby ensuring that a single instance is not scheduled across SuperPoDs.
>- For examples of deploying inference services using other types of K8s resources, see [YAML Files for Different Job Types and Hardware Models](../03_full_npu_scheduling.md#preparing-the-job-yaml). Add the corresponding labels to enable the inference service affinity scheduling policy.
>- For resources that can generate a PodGroup, adding the corresponding fields to the PodGroup can also implement inference service affinity scheduling.
>- For a comparison table of commonly used labels and annotations, see [PodGroup](../../../06_api/01_volcano.md#podgroup)/[Pod](../../../06_api/01_volcano.md#pod).
