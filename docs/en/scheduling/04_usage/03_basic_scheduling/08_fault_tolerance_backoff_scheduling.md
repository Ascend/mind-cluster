# Degraded Scheduling with Fault Backoff Tolerance<a name="ZH-CN_TOPIC_0000002480000001"></a>

## Before You Start<a name="ZH-CN_TOPIC_0000002480000002"></a>

Some chips and resources in a cluster do not affect the chip health status, but may cause jobs to fail repeatedly. The following three types are examples:

- Silent fault chips: When software faults occur repeatedly on a node but no hardware fault is reported within the detection period, ClusterD comprehensively determines the node as having a silent fault (`SilentFault`). A silent fault only appends to the fault list and does not change the chip health status, so job-level, Pod-level, or process-level rescheduling is not triggered.
- Pod error chips refer to the chips used by a Pod in its current run when the Pod fails due to various faults (including hardware faults) during job execution.
- Subhealthy resources: Include su-healthy NPU chips (including chips involved in public faults) and nodes with sub-healthy switches.

For such chips, Volcano provides a duration-based backoff tolerance mechanism. A job can configure the backoff duration through annotations. Within the backoff window, the scheduler no longer allocates such chips to the job (degraded scheduling, that is, these chips are not preferentially scheduled). After the window ends, their use is automatically restored, thereby preventing the job from being repeatedly scheduled onto faulty chips and causing continuous failures.

**Prerequisites<a name="section166381652174501"></a>**

To use the degraded scheduling with fault backoff tolerance feature, ensure that the following components have been installed. If they are not installed, see [Installation and Deployment](../../03_installation_guide/02_installation/00_helm_installation.md).

- Volcano (This feature supports only Volcano as the scheduler. Other schedulers are not supported.)
- Ascend Device Plugin
- ClusterD (The detection and reporting of silent fault chips depend on ClusterD. This component is optional when only Pod error chip backoff is used.)

**Usage Constraints<a name="section117804491812701"></a>**

- This function is supported only in 26.3.0 and later versions.
- Pod error chip backoff and sub-healthy backoff are both temporary backoff strategies, and they cannot be configured as permanent prohibition (`-1`).
- Subhealthy backoff takes effect only when the `subHealthyStrategy` label of the job is set to `ignore` or is not configured. This annotation takes effect only for full-NPU scheduling jobs, and vNPU jobs do not support sub-healthy backoff.

**Supported Products<a name="section169961844182901"></a>**

**Degraded scheduling with fault backoff tolerance** is supported on the following products.

- <term>Atlas inference products</term>
- <term>Atlas training products</term>
- <term>Atlas A2 products</term>
- <term>Atlas A3 products</term>
- <term>Ascend 950PR&950DT products</term>

**Principles<a name="section579011371711101"></a>**

Fault backoff tolerance takes effect at the chip granularity (rather than the node granularity), and the three types of backoff are independent of each other.

1. Silent fault chips are reported by ClusterD to the Volcano scheduler.
2. When a Pod of a job fails, the scheduler records the chips used by the Pod as the Pod error chips of the current job. Pod error backoff takes effect at the job level: only the job that encounters the error avoids these chips within the backoff window, and other jobs are not affected.
3. During scheduling, the scheduler determines, according to the backoff duration configured for the job, whether the silent fault chips, Pod error chips, and sub-healthy resources are still within the backoff window. If so, these chips are removed from the candidate chip set (nodes with sub-healthy switches are blocked as a whole during the preselection phase), implementing degraded scheduling. Use is automatically restored after the window ends.
4. When the silent fault backoff duration configured for a job is `-1` (permanent prohibition), if the node where a running job resides becomes a silent fault node and the job uses such a chip, the scheduler evicts the job and triggers rescheduling.
5. The default window of sub-healthy backoff is 10s. When the job waiting time is within the window, the scheduler does not schedule the job to sub-healthy resources. After the window is exceeded, sub-healthy resources are released.

## Command-Line Usage (Volcano)<a name="ZH-CN_TOPIC_0000002480000003"></a>

Corresponding configuration items are added for fault backoff tolerance in this chapter on the basis of [Full-NPU Scheduling](./03_full_npu_scheduling.md).

### Image Creation<a name="ZH-CN_TOPIC_0000002480000004"></a>

For details, see [Image Creation](./03_full_npu_scheduling.md#image-creation) in *Full-NPU Scheduling*.

### Script Adaptation<a name="ZH-CN_TOPIC_0000002480000005"></a>

For details, see [Script Adaptation](./03_full_npu_scheduling.md#script-adaptation) in *Full-NPU Scheduling*.

### Configuring Backoff Tolerance<a name="configuring-backoff-tolerance"></a>

Based on the job YAML in [Full-NPU Scheduling](./03_full_npu_scheduling.md#preparing-the-job-yaml), add the following fields to `metadata.annotations`.

```yaml
...
metadata:
  annotations:
    huawei.com/silent.unhealthy-tolerance.after-time: "1800"   # Backoff tolerance duration for chips with silent faults, in second
    huawei.com/pod.error-tolerance.after-time: "60"            # Backoff tolerance duration for Pod error chips, in second
    huawei.com/subhealthy.tolerance.after-time: "60"           # Backoff tolerance duration for sub-healthy resources, in second
...
```

**Table 1** Value list of backoff tolerance configuration items

<a name="table039616264491601"></a>

| Parameter | Default Value | Value Range | Description |
|--|--|--|--|
| huawei.com/silent.unhealthy-tolerance.after-time | 0 | [-1, 1800] | Backoff tolerance duration for chips with silent faults, in second.<ul><li>0: allowed immediately without backoff.</li><li>-1: permanently prohibited. Jobs are not scheduled to chips with silent faults, and a running job that uses a chip with a silent fault is evicted and rescheduled.</li><li>N (N&gt;0): jobs can use chips with silent faults only after being queued for N seconds. If N exceeds 1800, 1800 is used.</li></ul> |
| huawei.com/pod.error-tolerance.after-time | 0 | [0, 300] | Backoff tolerance duration for Pod error chips, in second.<ul><li>0: allowed immediately without backoff.</li><li>N (N&gt;0): within N seconds after a Pod failure, the job avoids the chip that caused the failure. If N exceeds 300, 300 is used.</li></ul><p>Pod error backoff is a temporary backoff, and -1 (permanent prohibition) is not supported.</p> |
| huawei.com/subhealthy.tolerance.after-time | 10 | [0, 300] | Backoff tolerance duration for sub-healthy resources, in second. When the job waiting time is within the tolerance duration, the scheduler does not schedule the job to sub-healthy resources. After the tolerance duration is exceeded, sub-healthy resources are released, ensuring that the job can still be scheduled when no healthy resources are available. For details about the values, see [Pod Annotation Usage](../../06_api/12_k8s.md#p_subhealthy_tolerance_k8s). |

>[!NOTE]
>
>- When Pod rescheduling occurs, escalation is performed level by level in the order of "single-Pod scheduling -> rescheduling of all Pods in the logical rack -> logical SuperPoD rescheduling -> job rescheduling". The single-Pod scheduling is performed first; if no redundant resources are available for rack affinity within 3 sessions, all Pods in the entire logical rack are rescheduled; if no redundant logical racks are available in the SuperPoD within 6 sessions, the entire logical SuperPoD is rescheduled; if scheduling still fails within 12 sessions, job rescheduling is used. See the fault rescheduling description in [Logical SuperPoD Affinity Scheduling](./01_affinity_scheduling/04_node_based_affinity.md#logical-superpod-affinity-scheduling) and [Logical Rack Affinity Scheduling](./01_affinity_scheduling/04_node_based_affinity.md#logical-rack-affinity-scheduling).
>- Pod error chip backoff is a temporary backoff strategy. If the window is too short, the job may still be scheduled to the original faulty node after the window ends. If the window is too long and no redundant resources are available in the rack or the SuperPoD, job rescheduling (overall rescheduling) is used. This may affect the recovery duration of process-level rescheduling scenarios.
>- After the ascend-for-volcano component is upgraded from a version earlier than 26.3.0 to 26.3.0 or later, scheduling to sub-healthy resources is not allowed within 10s by default. In scenarios without redundancy, this may cause a job scheduling delay of 10s. In Pod-level rescheduling scenarios, the rescheduling may be degraded to job-level rescheduling.

### Delivering a Job<a name="ZH-CN_TOPIC_0000002480000006"></a>

For details, see [Delivering a Job](./03_full_npu_scheduling.md#delivering-a-job) in *Full-NPU Scheduling*.

### Viewing Job Progress<a name="ZH-CN_TOPIC_0000002480000007"></a>

For details, see [Viewing Job Progress](./03_full_npu_scheduling.md#viewing-job-progress) in *Full-NPU Scheduling*.

### Verifying Backoff Tolerance<a name="verifying-backoff-tolerance"></a>

1. Construct a silent fault chip.

   See [Configuring Silent Fault Detection](../04_fault_detection_and_diagnosis/03_configuration/09_silent_faults.md) to enable silent fault detection, and wait for ClusterD to identity a silent fault chip.

2. Deliver a job configured with backoff tolerance configured.

   Modify the job YAML as described in [Configuring Backoff Tolerance](#configuring-backoff-tolerance) in this chapter, and then deliver the job.

3. Check the scheduling result.

   - When the silent fault backoff duration configured for the job is a positive number, the job is not scheduled to silent fault chips within the backoff window. After the backoff duration is exceeded, the job can be scheduled to these chips normally.
   - When the Pod error backoff duration configured for the job is a positive number and when the job is scheduled again after a Pod failure, the chip that caused the failure is avoided within the backoff window.
   - When the `subHealthyStrategy` label of the job is set to `ignore` or is not configured, and the sub-healthy backoff duration is a positive number, the job is not scheduled to sub-healthy resources within the backoff window. After the backoff duration is exceeded, the job can be scheduled to these resources normally.
