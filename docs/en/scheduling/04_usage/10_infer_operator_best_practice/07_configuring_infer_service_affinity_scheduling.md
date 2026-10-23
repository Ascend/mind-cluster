# Configuring Inference Service Affinity Scheduling

When an Infer Operator inference service is deployed in a SuperPoD scenario, Infer Operator supports automatically configuring the inference service affinity label (`inferserviceid`) for all instances under the same inference service, so that the Ascend-for-volcano scheduling plugin preferentially schedules all instances of the same inference service to the same rack or the same SuperPoD, giving full play to network advantages and improving the overall inference service running performance. For details about the inference service affinity scheduling policy, see [Inference Service Affinity Scheduling](../03_basic_scheduling/01_affinity_scheduling/06_infer_affinity_scheduling.md).

## Prerequisites

- Infer Operator services have been deployed. For details, see [Deploying Infer Operator Jobs](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md).

## Supported Product Forms

- Atlas 950 SuperPoD
- Atlas 850E
- Atlas 850
- Atlas 900 A3 SuperPoD
- Atlas 9000 A3 SuperPoD

## Automatic Affinity Label Configuration Principle

1. In the role of the InferServiceSet job YAML, configure the `huawei.com/schedule_policy` annotation as a SuperPoD scheduling policy (the value ends with `-sp`, for example, `chip8-node8-sp`, `chip8-node16-sp`, `chip8-node8-ra64-sp`, `chip2-node16-sp`, and `chip2-node8-sp`).
2. When Infer Operator creates an InstanceSet, after detecting the SuperPoD scheduling policy, it automatically injects the `inferserviceid` label into the InstanceSets of all roles under the inference service, with the value being the unique identifier of the current inference service (InferService). All instances under the same inference service have the same `inferserviceid` value.
3. The `inferserviceid` label is passed to the corresponding PodGroup and workload along with the InstanceSet. Based on this label, Ascend-for-volcano preferentially schedules the instances of the same inference service to the same rack (Atlas 950 SuperPoD) or the same SuperPoD (Atlas 850E, Atlas 850, and Atlas 900 A3 SuperPoD), and falls back to other SuperPoDs when resources are insufficient.

If you have manually configured `inferserviceid` in the labels of the role metadata, your configuration prevails, and Infer Operator does not overwrite it.

## Configuring Inference Service Affinity Scheduling

The following is an example of configuring inference service affinity scheduling for an Infer Operator job. Modify the bold part in the configuration. For details about the related configuration items, see [YAML Parameter Description](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#yaml-parameter-description).

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
          <strong>infer.huawei.com/gang-schedule: 'true' # Gang scheduling must be enabled in the affinity scheduling scenario</strong>
      spec:
        replicas: 1 # Number of Pod replicas of the workload in prefill
        <strong>podManagementPolicy: Parallel # When the workload is a StatefulSet and gang scheduling is enabled, this parameter must be set to Parallel</strong>
        selector:
          matchLabels:
            app: test-prefill # User-defined. It must be consistent with the app configuration in labels below
        template:
          metadata:
            labels:
              app: test-prefill # User-defined. It must be consistent with the app configuration in labels below
              fault-scheduling: 'grace' # Enable rescheduling
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # Identifies the product type
            <strong>annotations:</strong>
              <strong>sp-block: "8" # Number of chips in the logical SuperPoD. Set it to the total number of NPUs requested by the instance</strong>
              <strong>huawei.com/schedule_policy: chip8-node8-sp # SuperPoD scheduling policy. The value ends with -sp. Set it based on the hardware form</strong>
          spec:
            schedulerName: volcano # Specify Volcano as the scheduler
            containers:
            - name: prefill
              image: vllm-ascend:xxx # Customized vLLM image name
              ...
              resources:
                requests:
                  huawei.com/npu: 8
                limits:
                  huawei.com/npu: 8
              ... # Add the necessary mounts and running commands for the container
    - name: decode  # decode definition
      replicas: 1   # Number of decode replicas
      workload:     # CRD type information of the instance in decode
        apiVersion: apps/v1
        kind: StatefulSet # workload type. Currently, StatefulSet and Deployment are supported
      metadata:
        labels:
          <strong>infer.huawei.com/gang-schedule: 'true' # Gang scheduling must be enabled in the affinity scheduling scenario</strong>
      spec:
        replicas: 1 # Number of Pod replicas of the workload in decode
        <strong>podManagementPolicy: Parallel # When the workload is a StatefulSet and gang scheduling is enabled, this parameter must be set to Parallel</strong>
        selector:
          matchLabels:
            app: test-decode # User-defined. It must be consistent with the app configuration in labels below
        template:
          metadata:
            labels:
              app: test-decode # User-defined. It must be consistent with the app configuration in labels below
              fault-scheduling: 'grace' # Enable rescheduling
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # Identifies the product type
            <strong>annotations:</strong>
              <strong>sp-block: "8" # Number of chips in the logical SuperPoD. Set it to the total number of NPUs requested by the instance</strong>
              <strong>huawei.com/schedule_policy: chip8-node8-sp # SuperPoD scheduling policy. The value ends with -sp. Set it based on the hardware form</strong>
          spec:
            schedulerName: volcano # Specify Volcano as the scheduler
            containers:
            - name: decode
              image: vllm-ascend:xxx # Customized vLLM image name
              ...
              resources:
                requests:
                  huawei.com/npu: 8
                limits:
                  huawei.com/npu: 8
              ... # Add the necessary mounts and running commands for the container
</pre>

After the job is delivered, Infer Operator automatically configures the same `inferserviceid` label for the instances of the Prefill and Decode roles. You can run the following command to view the automatically injected label. `<namespace>` is the user-defined namespace, and `<instance-set-name>` is the instance set name.

```shell
kubectl get instanceset <instance-set-name> -n <namespace> -o jsonpath='{.metadata.labels.inferserviceid}'
```

>[!NOTE]
>
>- Inference service affinity scheduling relies on the PodGroup to pass the `inferserviceid` label to the Volcano scheduler. Therefore, `infer.huawei.com/gang-schedule` must be set to `'true'` to enable gang scheduling.
>- The `huawei.com/schedule_policy` and `sp-block` annotations must be configured in metadata.annotations of the role (as shown in the preceding example). If they are configured elsewhere, automatic injection of the affinity label is not triggered.
>- If you do not want to enable inference service affinity scheduling, configure `huawei.com/schedule_policy` as a non-SuperPoD scheduling policy (for example, `chip8-node8`), or configure `inferserviceid` yourself in the labels of the role metadata.
