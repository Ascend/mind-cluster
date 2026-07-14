# Affinity Scheduling Integration<a name="ZH-CN_TOPIC_0000002516224533"></a>

To decouple the scheduling layer from task resource types, the Ascend-for-volcano scheduling plugin now supports the configuration of Pod-level scheduling policies. You can directly configure scheduling-related parameters in the `metadata.labels` or `metadata.annotations` of a Pod without relying on PodGroup. This supports all Pod types, including `acjob`, `vcjob`, `Job`, `Deployment`, and `StatefulSet`.

## Feature Description <a name="section112161354155714"></a>

By adding specific labels or annotations to the Pod template of a K8s resource, you can control the core scheduling behavior of Volcano, including but not limited to:

- Affinity scheduling for Ascend AI processors
- Switch affinity scheduling
- Logical SuperPoD affinity scheduling
- Fault rescheduling

## Prerequisites<a name="section46282421720"></a>

Ensure that the Kubernetes cluster has been correctly deployed and configured with Volcano, and that the related scheduling plugin Ascend-for-volcano is enabled.

## Scheduling Policy Configuration Example<a name="section5997169155814"></a>

Taking `StatefulSet` as an example, all scheduling-related labels/annotations must be configured under `StatefulSet.spec.template.metadata` to ensure that the scheduler can correctly read them from the Pod instance.

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
  <strong>podManagementPolicy: Parallel   # Supports OrderedReady and Parallel modes. "OrderedReady" only supports intra-node affinity scheduling and huawei.com/schedule_minAvailable can only be 1. "Parallel" supports both intra-node and inter-node affinity scheduling.</strong>
  serviceName: service-headliness
  selector:
    matchLabels:
      app: mindspore
  <strong>template:</strong>
    <strong>metadata:</strong>
      <strong>labels:</strong>
        app: mindspore
        ring-controller.atlas: ascend-910
        <strong>fault-scheduling: force   # Fault rescheduling feature switch</strong>
        <strong>pod-rescheduling: "on"   # Pod-level rescheduling feature switch</strong>
        <strong>fault-retry-times: "85"    # Number of service plane fault rescheduling attempts</strong>
        <strong>tor-affinity: large-model-schema  # Switch affinity scheduling switch</strong>
        <strong>deploy-name: mindx-dls-test # This label must be added to generate rankTable. The value must be consistent with the task name.</strong>
      <strong>annotations:</strong>
        <strong>sp-block: "128"         # Logical SuperPoD affinity scheduling switch</strong>
        <strong>huawei.com/recover_policy_path: pod    # Switch to prevent Pod-level rescheduling from escalating to Job-level (when using vcjob, you need to configure this policy: policies: -event:PodFailed -action:RestartTask)</strong>
        <strong>huawei.com/schedule_minAvailable: "16"  # Minimum number of replicas for task scheduling. It is recommended to keep it consistent with the number of task replicas.</strong>
        <strong>huawei.com/skip-ascend-plugin: "enabled"    # When enabled, allows some special tasks (such as tasks that do not require NPU resources) to bypass the default check logic of Ascend-for-volcano.</strong>
    spec:
      schedulerName: volcano         # Use Volcano to schedule jobs.
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
>- If a PodGroup is created, the scheduling configuration in the `spec` will override the labels/annotations configuration on the generated Pods.
>- For resources that can generate a PodGroup, configuring the corresponding scheduling policy on the PodGroup can also enable affinity scheduling.
>- For a comparison table of commonly used Labels and Annotations, see [PodGroup](../../../api/volcano.md#podgroup)/[Pod](../../../api/volcano.md#pod).
