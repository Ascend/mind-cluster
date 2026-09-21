# Configuring Priority Scheduling for Inference Jobs

<!-- md-trans-meta sourceCommit=b3ad2e48550e7547214af4f421007ec4d2082590 translatedAt=2026-09-01T08:02:35.474Z pushedAt=2026-09-01T08:03:28.200Z -->

Infer Operator supports configuring priority levels for multiple roles to ensure that high-priority roles are created and scheduled first.

## Prerequisites

The Infer Operator service has been deployed. For details, see [Deploying an Infer Operator Inference Job with vLLM Proxy](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md).

## Priority Scheduling Principles

When deploying instances of different roles, Infer Operator creates a `Deployment`/`StatefulSet` for each instance (the specific type depends on the `workload` configuration item). After the priority scheduling feature is enabled and the priority of each role is configured, Infer Operator creates the corresponding instances in descending order of role priority.

When adapting to other serving platforms (for example, MindIE PyMotor) that enable scaling down prefill to preserve decode, you need to configure both priority scheduling and instance-level rescheduling.

## Configuring Priority Scheduling

The following is an example of configuring priority scheduling for an Infer Operator job. You need to modify the configuration in the bold parts below. For descriptions of related configuration items, see [YAML Parameter Description](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#yaml-parameter-description).
<pre codetype="yaml">
apiVersion: mindcluster.huawei.com/v1
kind: InferServiceSet
metadata:
  name: "my-test"
  namespace: default
spec:
  replicas: 1 # Replica count of the inference service
  template:
    <strong>schedulingStrategy: # Set type to Priority to enable priority scheduling</strong>
      <strong>type: Priority</strong>
    roles:
    - name: prefill # prefill definition
      replicas: 1   # prefill replica count
      <strong>priority: 2   # Priority configuration of prefill. A smaller value indicates a higher priority. This takes effect only when priority scheduling is enabled.</strong>
      workload:     # CRD type information of the instance in prefill
        apiVersion: apps/v1
        kind: StatefulSet # Workload type. StatefulSet/Deployment is currently supported.
      metadata:
        labels:
          infer.huawei.com/gang-schedule: 'false' # Disable gang scheduling. When enabled, a PodGroup is created for each workload instance.
      spec:
        replicas: 1 # Pod replica count of the workload in prefill
        podManagementPolicy: Parallel # This configuration is optional. When the workload is a StatefulSet and infer.huawei.com/gang-schedule is true, set it to Parallel.
        selector:
          matchLabels:
            app: test-prefill # User-defined. Must be consistent with the app configuration in labels below.
        template:
          metadata:
            labels:
              app: test-prefill # User-defined. Must be consistent with the app configuration in labels below.
              fault-scheduling: 'external-force' # Enable instance-level rescheduling
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # Identify the product type
            annotations:
              huawei.com/schedule_policy: chip8-node8 # Set based on the hardware form
          spec:
            schedulerName: volcano # Specify Volcano as the scheduler
            containers:
            - name: prefill
              image: vllm-ascend:xxx # Custom vLLM image name.
              ...
              resources:
                requests:
                  huawei.com/Ascend910: 8
                limits:
                  huawei.com/Ascend910: 8
              ... # Add the necessary mount items and run commands for the supplementary container.
    - name: decode  # decode definition
      replicas: 1   # decode replica count
      <strong>priority: 1   # Priority configuration for decode. A smaller value indicates a higher priority. It takes effect only when priority scheduling is enabled.</strong>
      workload:     # CRD type information of the instance in decode
        apiVersion: apps/v1
        kind: StatefulSet # Workload type. StatefulSet/Deployment is currently supported.
      metadata:
        labels:
          infer.huawei.com/gang-schedule: 'false' # Disable gang scheduling. When enabled, a PodGroup is created for each workload instance.
      spec:
        replicas: 1 # Pod replica count of the workload in decode
        podManagementPolicy: Parallel # This configuration is optional. When the workload is a StatefulSet and infer.huawei.com/gang-schedule is true, set it to Parallel.
        selector:
          matchLabels:
            app: test-decode # User-defined. Keep it consistent with the app configuration in labels below.
        template:
          metadata:
            labels:
              app: test-decode # User-defined. Keep it consistent with the app configuration in labels below.
              fault-scheduling: 'external-force' # Enable instance-level rescheduling.
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # Identify the product type.
            annotations:
              huawei.com/schedule_policy: chip8-node8 # Set it based on the hardware form.
          spec:
            schedulerName: volcano # Specify Volcano as the scheduler.
            containers:
            - name: decode
              image: vllm-ascend:xxx # Custom vllm image name.
              ...
              resources:
                requests:
                  huawei.com/Ascend910: 8
                limits:
                  huawei.com/Ascend910: 8
              ... # Add the necessary mount items and run commands for the supplementary container.
    - name: router  # router definition
      replicas: 1   # router replica count
      <strong>priority: 3   # router priority configuration. A smaller value indicates a higher priority. It takes effect only when priority scheduling is enabled.</strong>
      services:     # router services definition. Only one service is created within a role scope for the service defined here.
      - name: vllm-router-service
        spec:
          ports:    # service port definition
          - port: 1026
            protocol: TCP
            targetPort: 1026
          selector:
            app: test-router # User-defined. Must be consistent with the app configuration in the labels below.
          type: ClusterIP
      workload:     # CRD type information of the instance in router
        apiVersion: apps/v1
        kind: Deployment # Workload type. StatefulSet/Deployment is currently supported.
      spec:
        replicas: 1 # Pod replica count of the workload in router
        selector:
          matchLabels:
            app: test-router # User-defined. Must be consistent with the app configuration in the labels below.
        template:
          metadata:
            labels:
              app: test-router # User-defined. Must be consistent with the app configuration in the labels below.
          spec:
            schedulerName: volcano # Specify Volcano as the scheduler.
            containers:
            - name: router
              image: xxx:yyy # Custom image name.
              ... # Add the necessary mount items and run commands for the supplementary container.
</pre>
