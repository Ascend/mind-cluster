# Configuring Load-Based Elastic Scaling

<!-- md-trans-meta sourceCommit=09f26edcdf94b50c4f44371fc258b3798dbd04c2 translatedAt=2026-09-01T08:02:27.872Z pushedAt=2026-09-01T08:03:28.194Z -->

Infer Operator supports configuring elastic scaling policies for inference instances, enabling automatic adjustment of the number of inference instances based on their load.

## Prerequisites

- Infer Operator has been [installed and deployed](../../03_installation_guide/02_installation/00_helm_installation.md).
- If you need to configure the scaling metric type as External, you must first implement and deploy the corresponding External Metrics Adaptor. This Adaptor needs to provide the load metrics of the inference instance (such as request queue length, request processing time, etc.). You can refer to the [example](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/infer-operator-metrics-adaptor) for implementation. In the scenario of [deploying Infer Operator inference jobs based on MindIE PyMotor](./02_deploying_infer_operator_inference_job_with_mindie_pymotor.md), you can directly deploy and use Metrics Adaptor provided by this example.

> [!NOTE]
> In failure scenarios, automatic or manual elastic scaling is not supported. Before elastic scaling starts, you must ensure that all instances are in the `Running` state.

## Elastic Scaling Principles

Based on the elastic scaling configuration of an inference instance, Infer Operator creates the corresponding scaling controller resources (for example, Horizontal Pod Autoscaler (HPA)) for that instance. The scaling controller then automatically adjusts the desired replica count of the instance according to its load conditions.

## Configuring the Elastic Scaling Policy

The following example shows how to configure an elastic scaling policy for an inference instance. You need to add the configuration in bold. For details about the configuration items, see [YAML Parameter Description](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#yaml-parameter-description).

<pre codetype="yaml">
apiVersion: mindcluster.huawei.com/v1
kind: InferServiceSet
metadata:
  name: "my-test"
  namespace: default
spec:
  replicas: 1 # Replica count of the inference service
  template:
    roles:
    - name: prefill # Prefill definition
      replicas: 1   # Replica count of prefill
      workload:     # CRD type information of the instance in prefill
        apiVersion: apps/v1
        kind: StatefulSet # Workload type. Currently, StatefulSet and Deployment are supported.
      <strong>scalingPolicy:</strong>
        <strong>type: HPA # Elastic scaling policy type. Currently, only HPA is supported.</strong>
        <strong>spec: # HPA configuration</strong>
          <strong>minReplicas: 1 # Minimum replicas for scale-in</strong>
          <strong>maxReplicas: 4 # Maximum replicas for scale-out</strong>
          <strong>metrics: # List of HPA scaling metric configurations</strong>
          <strong>- type: External # Metric type: external custom metric (provided by External Metrics Adapter)</strong>
            <strong>external:</strong>
              <strong>metric:</strong>
                <strong>name: num_requests_waiting # External metric name</strong>
              <strong>target: # Target value configuration</strong>
                <strong>type: AverageValue</strong>
                <strong>averageValue: "5"</strong>
          <strong>... # Other HPA configuration items. Add them as needed and ensure they comply with the HPA configuration specifications.</strong>
      metadata:
        labels:
          infer.huawei.com/gang-schedule: 'false' # Disable gang scheduling. When enabled, a PodGroup is created for each workload instance.
      spec:
        replicas: 1 # Number of pod replicas of the workload in prefill
        podManagementPolicy: Parallel # This configuration is optional. When the workload is a StatefulSet and infer.huawei.com/gang-schedule is true, set this to Parallel.
        selector:
          matchLabels:
            app: test-prefill # User-defined. Must be consistent with the app configuration in labels below.
        template:
          metadata:
            labels:
              app: test-prefill # User-defined. Must be consistent with the app configuration in labels below.
              fault-scheduling: 'grace' # Enable rescheduling
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # Product type identifier
            annotations:
              huawei.com/schedule_policy: chip8-node8 # Set based on the hardware form
          spec:
            schedulerName: volcano # Specify Volcano as the scheduler
            nodeSelector:
              example-key: example-value    # Example value. You can configure nodeSelector based on your scheduling intent.
            containers:
            - name: prefill
              image: vllm-ascend:xxx # Custom vllm image name
              ...
              resources:
                requests:
                  huawei.com/Ascend910: 8
                limits:
                  huawei.com/Ascend910: 8
              ... # Necessary mount items and run commands for the supplementary container
    - name: decode  # decode definition
      replicas: 1   # decode replica count
      workload:     # CRD type information of the instance in decode
        apiVersion: apps/v1
        kind: StatefulSet # Workload type. StatefulSet/Deployment is currently supported
      <strong>scalingPolicy:</strong>
        <strong>type: HPA # Elastic scaling policy type. HPA is currently supported</strong>
        <strong>spec: # HPA configuration</strong>
          <strong>minReplicas: 1 # Minimum replicas for scale-in</strong>
          <strong>maxReplicas: 4 # Maximum replicas for scale-out</strong>
          <strong>metrics: # HPA scaling metric configuration list</strong>
          <strong>- type: External # Metric type: external custom metric (provided by External Metrics Adapter)</strong>
            <strong>external:</strong>
              <strong>metric:</strong>
                <strong>name: generation_tokens_per_second # External metric name</strong>
              <strong>target: # Target value configuration</strong>
                <strong>type: AverageValue</strong>
                <strong>averageValue: "10"</strong>
          <strong>... # Other HPA configuration items, added as needed and compliant with the HPA configuration specification</strong>
      metadata:
        labels:
          infer.huawei.com/gang-schedule: 'false' # Disable gang scheduling. When enabled, a PodGroup is created for each workload instance
      spec:
        replicas: 1 # Pod replica count of the workload in decode
        podManagementPolicy: Parallel # This configuration is optional. When the workload is a StatefulSet and infer.huawei.com/gang-schedule is true, it must be set to Parallel
        selector:
          matchLabels:
            app: test-decode # User-defined. Must maintain configuration consistency with the app in labels below
        template:
          metadata:
            labels:
              app: test-decode # User-defined. Must maintain configuration consistency with the app in labels below
              fault-scheduling: 'grace' # Enable rescheduling
              fault-retry-times: '10'
              ring-controller.atlas: ascend-910b # Identifies the product type
            annotations:
              huawei.com/schedule_policy: chip8-node8 # Set according to the hardware form
          spec:
            schedulerName: volcano # Specify Volcano as the scheduler
            containers:
            - name: decode
              image: vllm-ascend:xxx # Custom vllm image name
              ...
              resources:
                requests:
                  huawei.com/Ascend910: 8
                limits:
                  huawei.com/Ascend910: 8
              ... # Supplementary container's required mount items and run commands
    - name: router  # router definition
      replicas: 1   # router replica count
      services:     # router services definition. A service defined here is created only once within a role
      - name: vllm-router-service
        spec:
          ports:    # service port definition
          - port: 1026
            protocol: TCP
            targetPort: 1026
          selector:
            app: test-router # User-defined; must maintain configuration consistency with the app in the labels below
          type: ClusterIP
      workload:     # CRD type information of the instance in the router
        apiVersion: apps/v1
        kind: Deployment # Workload type; StatefulSet/Deployment are currently supported
      spec:
        replicas: 1 # Replica count of the workload pods in the router
        selector:
          matchLabels:
            app: test-router # User-defined; must maintain configuration consistency with the app in the labels below
        template:
          metadata:
            labels:
              app: test-router # User-defined; must maintain configuration consistency with the app in the labels below
          spec:
            schedulerName: volcano # Specify Volcano as the scheduler
            containers:
            - name: router
              image: xxx:yyy # Custom image name
              ... # Add the necessary mount items and run commands for the supplementary container
</pre>

> [!NOTE]
>
>- This feature supports Kubernetes 1.23 and later.
