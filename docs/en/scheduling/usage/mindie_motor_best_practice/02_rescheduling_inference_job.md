# Configuring Inference Task Rescheduling<a name="ZH-CN_TOPIC_0000002479386400"></a>

When node, chip, or other faults occur in an inference job, the MindCluster scheduling components can isolate the faulty resources and automatically perform rescheduling. For details about the fault detection principle, see the [Fault Detection](../resumable_training/01_solutions_principles.md#fault-detection) section.

## Prerequisites<a name="zh-cn_topic_0000002356060805_section19119249163119"></a>

[Deploying MindIE Motor](./01_deploying_mindie_motor.md) has been completed.

## Supported Fault Types<a name="section121201333144919"></a>

- MindIE Server: node, chip, or other faults
- MindIE MS: node fault

## Rescheduling Principles<a name="zh-cn_topic_0000002356060805_section4253197539"></a>

- Job-level rescheduling: Supported by both MindIE Server and MindIE MS. When a fault occurs on MindIE Server or MindIE MS, the corresponding MindIE Server instance or MindIE MS stops all pods, recreates and reschedules all pods, and then pushes the latest `global-ranktable.json` to the MS Controller, restarting the inference job.

    In a Prefill-Decode disaggregation scenario for MindIE Server, for example, if MindIE Server contains one Prefill instance and one Decode instance, and the Prefill instance fails, only all pods of the Prefill instance are stopped, without affecting other normally running instances.

- Pod-level rescheduling: Only supported by MindIE MS. In scenarios where the active/standby switchover feature is enabled, the number of pods corresponding to the MS Controller or MS Coordinator is greater than 1. When a node fails, only the pod corresponding to that node is stopped. For example, the MS Coordinator includes an active MS Coordinator and a standby MS Coordinator. When the active MS Coordinator fails, only the pod corresponding to the active MS Coordinator is stopped, and the standby MS Coordinator is not affected.

    >[!NOTE]
    >If pod-level rescheduling recovery fails, it will fall back to the Job-level rescheduling method.

## Configuring Job-Level Rescheduling<a name="zh-cn_topic_0000002356060805_section20633874524"></a>

Job-level rescheduling is enabled by default. You only need to complete the steps for preparing the job YAML. The following uses MindIE Server as an example to describe the configuration of Job-level rescheduling.

<pre codetype="yaml">
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-server-0
  namespace: mindie
  labels:
    framework: pytorch
    app: mindie-ms-server        # Indicates the role of MindIE Motor in Ascend Job. Do not modify.
    jobID: mindie-ms-test        # The unique ID of the current MindIE Motor inference job in the cluster. Users can configure it based on actual conditions.
    <strong>fault-scheduling: force      # Enable the rescheduling feature</strong>
    fault-retry-times: "10000"     # Enable service-plane fault rescheduling. The value is the number of rescheduling attempts upon a service-plane fault.
    ring-controller.atlas: ascend-910b
spec:
  schedulerName: volcano   # The scheduler selected when Ascend Operator enables "gang" scheduling.
  runPolicy:
    schedulingPolicy:      # This field takes effect only when Ascend Operator enables "gang" scheduling and the scheduler is Volcano.
      minAvailable: 2      # Total number of replicas for job running
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:</pre>

## Configuring Pod-Level Rescheduling<a name="section5620411141"></a>

Pod-level rescheduling currently supports only MS Controller and MS Coordinator. It is recommended for use in scenarios where the active/standby switchover feature is enabled. The following uses enabling active/standby switchover for MS Coordinator as an example to illustrate the configuration of pod-level rescheduling.

<pre codetype="yaml">
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-coordinator
  namespace: mindie
  labels:
    framework: pytorch
    app: mindie-ms-coordinator        # Indicates the role of MindIE Motor in Ascend Job. Do not modify.
    jobID: mindie-ms-test             # The unique identifier of the current MindIE Motor inference task in the cluster. Configure this based on actual conditions.
    <strong>fault-scheduling: force          # Enable the rescheduling feature</strong>
    <strong>pod-rescheduling: "on"           # Enable pod-level rescheduling</strong>
    ring-controller.atlas: ascend-910b
spec:
  schedulerName: volcano   # The scheduler selected when Ascend Operator enables "gang" scheduling
  runPolicy:
    schedulingPolicy:      # This field takes effect only when Ascend Operator enables "gang" scheduling and the scheduler is Volcano
      <strong>minAvailable: 2      # Total number of replicas for job execution</strong>
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:</pre>
