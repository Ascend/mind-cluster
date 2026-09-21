# Configuring Inference Job Rescheduling<a name="ZH-CN_TOPIC_0000002479386400"></a>

When a node, chip, or other fault occurs in an inference job, the MindCluster cluster scheduling components can isolate the faulty resource and automatically perform rescheduling. To learn about the fault detection principle, see the [Fault Detection](../04_resumable_training/01_solutions_principles.md#fault-detection) section.

## Prerequisites<a name="zh-cn_topic_0000002356060805_section19119249163119"></a>

You have completed [deploying MindIE Motor](./01_deploying_mindie_motor.md).

## Supported Fault Types<a name="section121201333144919"></a>

- MindIE Server: node, chip, or other faults
- MindIE MS: node faults

## Rescheduling Principles<a name="zh-cn_topic_0000002356060805_section4253197539"></a>

- Job-level rescheduling: supported by both MindIE Server and MindIE MS. When a fault occurs on MindIE Server or MindIE MS, the corresponding MindIE Server instance or MindIE MS stops all Pods, recreates and reschedules all Pods, and then pushes the latest `global-ranktable.json` to MS Controller again, restarting the inference job.

    In a prefill-decode disaggregation scenario of MindIE Server, for example, when MindIE Server contains one Prefill instance and one Decode instance, if a fault occurs on the Prefill instance, only all Pods of the Prefill instance are stopped, without affecting other normally running instances.

- Pod-level rescheduling: supported only by MindIE MS. In the scenario where the active/standby switchover function is enabled, the number of Pods corresponding to MS Controller or MS Coordinator is greater than 1. When a fault occurs on a node, only the Pod corresponding to that node is stopped. For example, when MS Coordinator contains an active MS Coordinator and a standby MS Coordinator, if a fault occurs on the active MS Coordinator, only the Pod corresponding to the active MS Coordinator is stopped, without affecting the standby MS Coordinator.

    >[!NOTE]
    >If Pod-level rescheduling fails to recover, it falls back to the Job-level rescheduling handling method.

## Configuring Job-Level Rescheduling<a name="zh-cn_topic_0000002356060805_section20633874524"></a>

Job-level rescheduling is enabled by default. Users only need to complete the step of preparing the job YAML. The following uses MindIE Server as an example to describe the configuration of Job-level rescheduling.

<pre codetype="yaml">
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-server-0
  namespace: mindie
  labels:
    framework: pytorch
    app: mindie-ms-server        # Indicates the role of MindIE Motor in AscendJob. This value cannot be modified.
    jobID: mindie-ms-test        # The unique ID of the current MindIE Motor inference job in the cluster. Users can configure it based on actual conditions.
    <strong>fault-scheduling: force      # Enable rescheduling.</strong>
    fault-retry-times: "10000"     # Enable service-plane fault rescheduling. The value is the number of rescheduling attempts when a service-plane fault occurs.
    ring-controller.atlas: ascend-910b
spec:
  schedulerName: volcano   # The scheduler selected when Ascend Operator enables "gang" scheduling.
  runPolicy:
    schedulingPolicy:      # This field takes effect only when Ascend Operator enables "gang" scheduling and the scheduler is Volcano.
      minAvailable: 2      # Total number of replicas for job running.
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:</pre>

### Dispatching a Job<a name="ZH-CN_TOPIC_0000002511427027"></a>

In the path where the sample YAML is located on the management node, run the following command to dispatch the inference job using the YAML.

```shell
kubectl apply -f XXX.yaml
```

For example:

```shell
kubectl apply -f infer-job.yaml
```

Command output:

```ColdFusion
ascendjob.mindxdl.gitee.com/mindie-server-0 created
```

>[!NOTE]
>If you modify the job YAML after the job is successfully dispatched, run the kubectl delete -f _XXX_.yaml command to delete the original job first, and then dispatch the job again.

### Viewing Job Process<a name="ZH-CN_TOPIC_0000002511427025"></a>

Run the following command to view the Pod running status.

```shell
kubectl get pod --all-namespaces
```

Command output:

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          mindie-server-master-0                     1/1     Running   0          20m
...
```

### Viewing Inference Card Fault Rescheduling Results<a name="ZH-CN_TOPIC_0000002511347069"></a>

When a fault occurs while an inference job is running (an error can be actively triggered through the service code), Volcano schedules the job to another NPU.

Run the following command to view the job running status.

```shell
kubectl get pod --all-namespaces
```

View the job Pod. You can see that the job is recreated after being deleted. The original Pod is deleted after its status becomes `Error`, and a new job Pod is created. The Pod status changes from `Pending` to `ContainerCreating` and then to `Running`, and AGE starts from 0 seconds, indicating that the fault rescheduling feature runs successfully.

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          mindie-server-0-master-0                   1/1     Running   0          1s
...
```

### Deleting a Job<a name="ZH-CN_TOPIC_0000002479387108"></a>

In the path where the sample YAML is located, run the following command to delete the corresponding inference job.

```shell
kubectl delete -f XXX.yaml
```

For example:

```shell
kubectl delete -f infer-job.yaml
```

Command output:

```ColdFusion
ascendjob.mindxdl.gitee.com "mindie-server-0" deleted
```

## Configuring Pod-Level Rescheduling<a name="section5620411141"></a>

Pod-level rescheduling currently supports only MS Controller and MS Coordinator, and is recommended for use in scenarios where the active/standby switchover function is enabled. The following uses MS Coordinator with the active/standby switchover function enabled as an example to describe the configuration of Pod-level rescheduling.

<pre codetype="yaml">
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-coordinator
  namespace: mindie
  labels:
    framework: pytorch
    app: mindie-ms-coordinator        # Indicates the role of MindIE Motor in AscendJob. This value cannot be modified.
    jobID: mindie-ms-test             # Unique identifier of the current MindIE Motor inference job in the cluster. Configure it based on the actual situation.
    <strong>fault-scheduling: force          # Enable the rescheduling function.</strong>
    <strong>pod-rescheduling: "on"           # Enable Pod-level rescheduling.</strong>
    ring-controller.atlas: ascend-910b
spec:
  schedulerName: volcano   # Scheduler selected when Ascend Operator enables gang scheduling.
  runPolicy:
    schedulingPolicy:      # This field takes effect only when Ascend Operator enables gang scheduling and the scheduler is Volcano.
      <strong>minAvailable: 2      # Total number of replicas for the job.</strong>
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:</pre>

### Dispatching a Job<a name="ZH-CN_TOPIC_0000002511427027"></a>

In the path where the sample YAML is located on the management node, run the following command to use the YAML to dispatch a multi-Pod inference job (only in a multi-Pod job scenario can you see the difference from Job-level rescheduling in k8s).

```shell
kubectl apply -f XXX.yaml
```

For example:

```shell
kubectl apply -f infer-job.yaml
```

Command output:

```ColdFusion
ascendjob.mindxdl.gitee.com/mindie-coordinator created
```

>[!NOTE]
>If you modify the job YAML after the job is successfully dispatched, run the kubectl delete -f _XXX_.yaml command to delete the original job first, and then dispatch the job again.

### Viewing Job Process<a name="ZH-CN_TOPIC_0000002511427025"></a>

Run the following command to view the running status of the Pods.

```shell
kubectl get pod --all-namespaces
```

Command output:

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          mindie-coordinator-master-0                1/1     Running   0          20m
default          mindie-coordinator-worker-1                1/1     Running   0          20m
...
```

### Viewing Inference Card Fault Rescheduling Results<a name="ZH-CN_TOPIC_0000002511347069"></a>

When a fault occurs on a Pod while the inference job is running (which can be triggered proactively by the service code reporting an error), Volcano reschedules that Pod to another NPU.

Run the following command to view the job running status.

```shell
watch -n 1 kubectl get pod --all-namespaces
```

View the job Pods. You can see that the error Pod is deleted and then recreated: the error Pod is deleted after its status becomes `Error`, a new Pod is created, and the Pod status transitions from `Pending` to `ContainerCreating` and then to `Running`, with AGE starting from 0 seconds. This indicates that the fault rescheduling feature has run successfully.

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          mindie-coordinator-master-0                1/1     Running   0          20m
default          mindie-coordinator-worker-1                1/1     Running   0          1s
...
```

### Deleting a Job<a name="ZH-CN_TOPIC_0000002479387108"></a>

In the path where the sample YAML is located, run the following command to delete the corresponding inference job.

```shell
kubectl delete -f XXX.yaml
```

For example:

```shell
kubectl delete -f infer-job.yaml
```

Command output:

```ColdFusion
ascendjob.mindxdl.gitee.com "mindie-coordinator" deleted
```
