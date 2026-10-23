# Rescheduling<a name="ZH-CN_TOPIC_0000002479387124"></a>

## Before You Start<a name="ZH-CN_TOPIC_0000002479387116"></a>

The rescheduling feature means that when a hardware or software fault occurs during job running, the cluster scheduling components can reschedule the job to a new healthy node or chip, ensuring that the job can run again properly.
The default rescheduling mode is job-level rescheduling, that is, on each fault, all Pods are stopped, all Pods are recreated and rescheduled, and then the job is restarted.

**Prerequisites<a name="section166381652174516"></a>**

To use the rescheduling feature, ensure that the following components have been installed. If they are not installed, perform operations by referring to [Installation and Deployment](../../03_installation_guide/02_installation/00_helm_installation.md).

- Volcano (this feature supports only Volcano as the scheduler; other schedulers are not supported.)
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator (must be installed when AscendJob is used)
- ClusterD
- NodeD
- Infer Operator (must be installed when an InferServiceSet job is used)

**Constraints<a name="section1178044918127"></a>**

- This function is supported only in 6.0.RC2 and later versions.
- In large-scale K8s cluster scenarios, the ConfigMap mapping latency is uncontrollable. You are advised to use the shared storage mode for RankTable.

**Supported Product Forms<a name="section169961844182917"></a>**

The following products support **rescheduling**.

- <term>Atlas inference products</term>
- <term>Atlas training products</term>
- <term>Atlas A2 products</term>
- <term>Atlas A3 products</term>
- <term>Ascend 950PR&950DT products</term>

**Rescheduling Principle<a name="section57901137171110"></a>**

If a hardware or software fault occurs during job running, the job status becomes abnormal. Job-level rescheduling first destroys all job containers, then isolates the faulty device, and finally reschedules and starts the job containers again. After the job containers are restarted, the job is started again, which is similar to the first startup of the job.

**Figure 1** Working principle<a name="fig18343114924113"></a>

![Job rescheduling principle](../../../figures/scheduling/principles.png)

The steps in the preceding working principle diagram are described as follows:

1. After a fault is detected, delete all Pods of the current job first.
2. Isolate the device where the hardware fault occurs to prevent the device from being used again.
3. Recreate and schedule the job Pods.
4. After the container is started, start the job process to recover the job.

## Command-line Usage (Volcano)<a name="ZH-CN_TOPIC_0000002511427039"></a>

The rescheduling feature configuration in this section adds the corresponding configuration items on the basis of full-NPU scheduling.

### Image Creation<a name="ZH-CN_TOPIC_0000002511427053"></a>

For details, see [Creating an Image](./03_full_npu_scheduling.md#image-creation) in *Full-NPU Scheduling*.

### Script Adaptation<a name="ZH-CN_TOPIC_0000002479227172"></a>

For details, see [Script Adaptation](./03_full_npu_scheduling.md#script-adaptation) in *Full-NPU Scheduling*.

### Configuring Job-Level Rescheduling

Job-level rescheduling is enabled by default. Based on the job YAML in [Full-NPU Scheduling](./03_full_npu_scheduling.md#preparing-the-job-yaml), add the following fields to `metadata.labels` to enable job-level rescheduling.

```yaml
...
metadata:
  labels:
    ...
    fault-scheduling: "grace"      # You can set this parameter to force or grace based on the actual situation. When it is set to force, it is not recommended that the Pod use the host network, and rescheduling may fail and be triggered multiple times until it succeeds.
    fault-retry-times: "100"         # Number of unconditional attempts. This parameter needs to be configured in the software fault scenario.
```

The meanings of the values of the `fault-scheduling` configuration item are as follows.

**Table 2** fault-scheduling configuration values

<a name="table0396162644916"></a>

|Parameter|Value|Meaning|
|--|--|--|
|fault-scheduling|grace|The job uses the rescheduling switch and gracefully deletes the original Pod during the process.|
|fault-scheduling|force|The job uses the forcible deletion mode and forcibly deletes the original Pod during the process.|

### Delivering a Job<a name="ZH-CN_TOPIC_0000002511427027"></a>

For details, see [Delivering a Job](./03_full_npu_scheduling.md#delivering-a-job) in *Full-NPU Scheduling*.

### Viewing Job Progress<a name="ZH-CN_TOPIC_0000002511427025"></a>

For details, see [Viewing Job Progress](./03_full_npu_scheduling.md#viewing-job-progress) in *Full-NPU Scheduling*.

### Verifying Job-Level Rescheduling

1. Construct a fault.

   1. Query the job process on the node where the job is running.

      ```bash
      npu-smi info|grep python|awk '{print $5}'
      ```

      Command output:

      ```ColdFusion
      2398104
      2398105
      2398107
      ```

   2. Terminate the process to simulate a fault.

      ```bash
      kill -9 2398104
      ```

2. Observe the rescheduling process.

   Monitor the Pod status changes of the job.

   ```bash
   kubectl get pod -n <namespace> -o wide -w | grep <pod-name>
   ```

   The following is an example of the Pod history status of the job. By observing the changes of the bold fields, you can find that the Pod of the job goes through the `Terminating` -> `Pending` -> `ContainerCreating` -> `Running` stages and then runs properly, indicating that the job rescheduling is successful:

   <pre codetype="ColdFusion">
   trjob            taskmgr-npu-020-default-test-0                  1/1     Running             0          2s      xx.xx.xx.xx       node173                 &lt;none&gt;           &lt;none&gt;
   // ===================== Inject a fault ======================
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Terminating</strong>         0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          0s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          1s      &lt;none&gt;            &lt;none&gt;                  &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          43s     &lt;none&gt;                 node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>Pending</strong>             0          43s     &lt;none&gt;                 node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>ContainerCreating</strong>   0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 0/1     <strong>ContainerCreating</strong>   0          43s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   trjob            <strong>taskmgr-npu-020-default-test-0</strong>                 1/1     <strong>Running</strong>             0          45s     xx.xx.xx.xx      node173                 &lt;none&gt;           &lt;none&gt;
   </pre>

## Usage After Integration

For details, see [Usage After Integration](./03_full_npu_scheduling.md#usage-after-integration) in *Full-NPU Scheduling*, and add the rescheduling configuration items on top of it.
