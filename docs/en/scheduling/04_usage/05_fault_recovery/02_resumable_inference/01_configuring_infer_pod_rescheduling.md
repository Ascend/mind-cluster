# Configuring Pod-Level Rescheduling for Inference Jobs

When a node, chip, or other fault occurs in an Infer Operator inference job, the MindCluster cluster scheduling components can isolate the faulty resource and automatically perform rescheduling. For details about the fault detection principle, see [Fault Detection](../../04_fault_detection_and_diagnosis/01_working_principle.md#overall-architecture-of-fault-detection).

Unlike instance-level rescheduling (which reschedules the entire instance corresponding to a faulty Pod), Pod-level rescheduling deletes and reschedules only the faulty Pod when a fault occurs, and other normal Pods in the instance are not affected. Pod-level rescheduling applies to the service plane fault scenario (a process in a Pod exits with a non-zero code). In the case of a hardware fault, the faulty Pod is not deleted, and the service is allowed to continue running on the remaining available NPU cards in the faulty Pod. For details about how to configure instance-level rescheduling, see [Configuring Instance-Level Rescheduling for Inference Jobs](../../10_infer_operator_best_practice/04_configuring_rescheduling_inference_job.md).

## Prerequisites

Infer Operator services have been deployed. For details, see [Deploying Infer Operator Jobs](../../10_infer_operator_best_practice/01_deploying_infer_operator_inference_job_with_vllm_proxy.md).

## Rescheduling Principle

When Infer Operator deploys instances of different roles, it creates `Deployment` or `StatefulSet` instances (the specific type depends on the workload configuration item). After Pod-level rescheduling is enabled, when a service plane fault occurs on a Pod, Infer Operator deletes only the faulty Pod, the corresponding Controller recreates the deleted Pod, and Volcano reschedules the newly created Pod.

The following example yaml describes the meaning of each configuration item:

- `fault-scheduling: external-force-pod-failed`: Handles only service plane faults (a process in a Pod exits with a non-zero code). In the case of a hardware fault, the Pod on the faulty node is not deleted.
- `pod-rescheduling: on`: When the value is on, faulty Pods are deleted at the Pod granularity. When the value is another value, they are deleted at the instance granularity.
- `fault-retry-times`: Number of retries upon service plane faults.

## Configuring Pod-Level Rescheduling

The following is an example of configuring Pod-level rescheduling for an Infer Operator job. Modify the bold part in the configuration. For details about the related configuration items, see [YAML Parameter Description](../../10_infer_operator_best_practice/01_deploying_infer_operator_inference_job_with_vllm_proxy.md#yaml-parameter-description).

<pre codetype="yaml">
apiVersion: mindcluster.huawei.com/v1
kind: InferServiceSet
metadata:
  name: "my-test"
  namespace: default
spec:
  replicas: 1 # Number of inference service replicas
  template:
    roles:
      - name: prefill # prefill definition
        replicas: 1   # Number of prefill replicas
        workload:     # CRD type information of the instance in prefill
          apiVersion: apps/v1
          kind: StatefulSet # workload type. Currently, StatefulSet and Deployment are supported
        metadata:
          labels:
            infer.huawei.com/gang-schedule: 'true' # Enable gang scheduling. When it is enabled, a PodGroup is created for each workload instance
        spec:
          replicas: 2 # Number of Pod replicas of the workload in prefill
          podManagementPolicy: Parallel # When the workload is a StatefulSet and infer.huawei.com/gang-schedule is true, this parameter must be set to Parallel
          selector:
            matchLabels:
              app: test-prefill # User-defined. It must be consistent with the app configuration in labels below
          template:
            metadata:
              labels:
                app: test-prefill # User-defined. It must be consistent with the app configuration in labels below
                <strong>fault-scheduling: 'external-force-pod-failed' # Enable rescheduling. Only service plane faults (a process in a Pod exits with a non-zero code) are handled</strong>
                <strong>pod-rescheduling: 'on' # Enable Pod-level rescheduling</strong>
                <strong>fault-retry-times: '10' # Number of retries upon service plane faults</strong>
                ring-controller.atlas: ascend-910b # Identifies the product type
              annotations:
                huawei.com/schedule_policy: chip8-node8 # Set based on the hardware form
            spec:
              schedulerName: volcano # Specify Volcano as the scheduler
              containers:
                - name: prefill
                  image: vllm-ascend:xxx # Customized vLLM image name
                  imagePullPolicy: IfNotPresent
                  command: [...]
                  resources:
                    requests:
                      huawei.com/Ascend910: 8
                    limits:
                      huawei.com/Ascend910: 8
                  ... # Add the necessary mounts and running commands for the container
      - name: decode  # decode definition
        replicas: 1   # Number of decode replicas
        workload:     # CRD type information of the instance in decode
          apiVersion: apps/v1
          kind: StatefulSet # workload type. Currently, StatefulSet and Deployment are supported
        metadata:
          labels:
            infer.huawei.com/gang-schedule: 'true' # Enable gang scheduling. When it is enabled, a PodGroup is created for each workload instance
        spec:
          replicas: 1 # Number of Pod replicas of the workload in decode
          podManagementPolicy: Parallel # When the workload is a StatefulSet and infer.huawei.com/gang-schedule is true, this parameter must be set to Parallel
          selector:
            matchLabels:
              app: test-decode # User-defined. It must be consistent with the app configuration in labels below
          template:
            metadata:
              labels:
                app: test-decode # User-defined. It must be consistent with the app configuration in labels below
                <strong>fault-scheduling: 'external-force-pod-failed' # Enable rescheduling. Only service plane faults (a process in a Pod exits with a non-zero code) are handled</strong>
                <strong>pod-rescheduling: 'on' # Enable Pod-level rescheduling</strong>
                <strong>fault-retry-times: '10'</strong>
                ring-controller.atlas: ascend-910b # Identifies the product type
              annotations:
                huawei.com/schedule_policy: chip8-node8 # Set based on the hardware form
            spec:
              schedulerName: volcano # Specify Volcano as the scheduler
              containers:
                - name: decode
                  image: vllm-ascend:xxx # Customized vLLM image name
                  imagePullPolicy: IfNotPresent
                  command: [...]
                  resources:
                    requests:
                      huawei.com/Ascend910: 8
                    limits:
                      huawei.com/Ascend910: 8
                  ... # Add the necessary mounts and running commands for the container
</pre>
