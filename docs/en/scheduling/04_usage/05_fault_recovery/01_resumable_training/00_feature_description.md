# Feature Description<a name="ZH-CN_TOPIC_0000002511346415"></a>

## Application Scenarios<a name="ZH-CN_TOPIC_0000002511346439"></a>

As neural networks and datasets continue to grow in scale, a single server can no longer handle large-scale training jobs. To address this challenge, multiple servers (equipped with more AI chips) are typically used to form a high-density training cluster for long-duration distributed training. However, as the number of hardware devices increases, the probability of device failures also rises, leading to more frequent training interruptions. Therefore, improving cluster availability has become a critical issue that needs to be addressed urgently.

Improving cluster availability requires reducing the fault recovery cost after each training interruption. Currently, fault recovery often involves manual troubleshooting of hardware failures or software anomalies, incurring significant labor costs. Additionally, isolating the faulty device and then re-launching a training job takes a considerable amount of time, impacting overall efficiency.

To solve these problems, resumable training is introduced, covering the following features. Relying on fault detection features to provide fault events, fault levels, and resource status, it can effectively handle faults during training and reduce recovery time, thereby significantly improving cluster availability and stability.

## Capability Overview

**Key Features<a name="section15584171017252"></a>**

<a name="table1866285218270"></a>

| Feature | Description | Configuration Steps |
|--|--|--|
| Fault Detection (Prerequisite) | <p>Resumable training relies on the fault detection feature to provide fault events, fault levels, and resource status for the cluster and training services.</p><p>For details on functions and principles, see [Fault Detection](../../04_fault_detection_and_diagnosis/00_feature_description.md).</p> | [Configuring Fault Detection Levels](../../04_fault_detection_and_diagnosis/03_configuration/01_fault_classification.md) |
| Fault Handling | <p>Automatically isolates faulty devices without manual intervention after a fault occurs.</p><p>For details on functions and principles, see [Fault Handling](./01_solutions_principles/01_fault_handling.md).</p> | [Configuring Fault Handling](./02_configuration/01_configuring_fault_handling_policies.md) |
| Recovery Acceleration | <p>Accelerates training recovery. You can customize acceleration policies to reduce training launch time.</p><p>For details on functions and principles, see [Recovery Acceleration](./01_solutions_principles/02_recovery_acceleration.md).</p> | [Configuring Recovery Acceleration](./02_configuration/02_configuring_training_recovery.md) |

**Use Cases<a name="section1498618364358"></a>**

<a name="table1716618439356"></a>

|Scenario |Main Business|Business Value|
|--|--|--|
|AI training|Supports monitoring of computing, network, and storage device resources, health checks of the AI environment, and AI job fault diagnosis.|<ul><li>Overall monitoring of cluster environment resources.</li><li>Improve the success rate of AI training jobs.</li><li>Reduce the handling and recovery time of AI job training faults.</li></ul>|

>[!NOTE]
>
>- For smaller-scale model tasks with short training duration (< 1h), hardware failures occur less frequently, and it is not recommended to use the resumable training feature.
>- This feature is not applicable to computing power virtualization scenarios.

## Overall Architecture<a name="ZH-CN_TOPIC_0000002479226568"></a>

When a training job fails in a Kubernetes (K8s) cluster, the resumable training feature enables the system to perceive the fault, handle or isolate the faulty resources, and reallocate resources as needed by the training job. It then relaunches the training job using periodically saved or dying gasp checkpoint, thereby reducing lost time.

**Overall Architecture<a name="section1483121685613"></a>**

The architecture principle of resumable training is shown in [Figure 1](#fig1285977919).

**Figure 1**  Overall architecture<a name="fig1285977919"></a>

![](../../../../figures/scheduling/overall-architecture.png)

The capabilities of each component are as follows:

1. Ascend Device Plugin: Detects faults by providing capabilities such as NPU resource management, NPU chip fault and NPU network fault reporting, and chip hot reset.
2. NodeD: Detects faults by providing capabilities for reporting node health status, node hardware (including CPU, memory, chip, and other components) faults, UnifiedBus network faults, and DPC shared storage faults.
3. Volcano: Handles faults by rescheduling faulty jobs.
4. Ascend Operator: Generates the corresponding environment variables for distributed training jobs of different AI frameworks; provides the RankTable information required for static networking collective communication.
5. ClusterD: Obtains all the data reported by Ascend Device Plugins and NodeD in the cluster, organizes it, and sends it to Volcano.
6. TaskD: Provides communication functionality with ClusterD to complete resumable training; provides training status monitoring and training status control capabilities for training and inference jobs on Ascend devices.
7. MindIO TTP: After a fault occurs during large model training, it verifies the integrity and consistency of intermediate state data, generates dying gasp checkpoint data, and can recover through this checkpoint data when resuming training, reducing the loss of training iterations caused by the fault.
8. Training model code: Adaptation operations related to resumable training capabilities are required.

**End-to-End Flow<a name="section11999131581210"></a>**

The resumable training feature relies on the fault events, fault levels, and resource status provided by the fault detection feature, and resumes training after fault handling and recovery acceleration.

**Figure 2** End-to-end flow<a name="fig1968122593112"></a>

![](../../../../figures/scheduling/end-to-end-flow.png)

The description of each step is as follows:

1. Query the device status through polling. Ascend Device Plugin obtains the NPU status from the DCMI, as well as the node health status, node hardware fault information, and UnifiedBus network fault information reported by NodeD. ClusterD consolidates all fault information, determines the final fault status, and reports it to Volcano.
2. After a node or chip fault is detected, isolate the faulty node or chip to prevent it from being scheduled again.
3. Stop the training process and exit the training container.
4. After a node or chip fault occurs, the system reschedules the training job to a healthy device and restarts the training container. When resources are selected for the rescheduled training job, nodes that did not cause the current training job rescheduling are preferred.
5. The training script relaunches the training process.
6. O&M personnel can determine whether a hot reset is feasible based on the fault type of the node or chip.
7. Perform a hot reset to restore the device to a healthy state.
8. The recovered device automatically rejoins the cluster.
9. Unrecoverable devices report alarms through the O&M monitoring system.
10. Perform offline manual repair and part replacement for unrecoverable devices.

>[!NOTE]
>Resumable training triggered by service-plane faults will only execute steps 3 to 5 above.

**Component Invocation Process<a name="section1726018439396"></a>**

The component invocation process of resumable training is shown in [Figure 3](#fig1710473818543).

**Figure 3** Architecture principle<a name="fig1710473818543"></a>

![](../../../../figures/scheduling/architectures%20and%20principles.png)

The description of each step is as follows:

1. Ascend Device Plugin discovers and reports faults and health status.
2. NodeD updates node hardware fault information so that Volcano can accurately determine the node fault type.
3. ClusterD determines whether the chip is healthy based on the chip information provided by Ascend Device Plugin.
4. ClusterD obtains the fault information reported by NodeD.
5. ClusterD aggregates the collected chip and node information and places it into a ConfigMap.
6. Volcano obtains the device information of the entire cluster. If there is fault information on the device used by the job, Volcano will schedule the job to other healthy devices.
7. Volcano selects nodes and chips based on affinity rules, and after Ascend Operator creates a new Pod, it schedules the training job to nodes that meet the requirements.
8. Ascend Device Plugin allocates chips based on the chip IDs specified by Volcano on the Pod, and writes the chip IP information into the container.
9. Before the container starts, Ascend Docker Runtime automatically mounts NPU-related devices, driver so files, and directories for the training container.
10. Ascend Operator writes the relevant environment variables required by the training job (such as collective communication information and training configuration) into the container. It also obtains the chip information on the training container and automatically generates the collective communication information required for the distributed training job.

**Usage Conditions<a name="section179221342141518"></a>**

- For details about the components required to use the resumable training feature, see the [Required Components](../../../01_introduction/02_feature_description.md#resumable-training) section.

- The resumable training feature is an advanced feature based on the MindCluster cluster scheduling components. It combines the Ascend full-stack software and hardware to implement training fault recovery. Before using this feature, the following prerequisites must be met.

  - Complete basic performance tuning of the K8s cluster. For details, see[K8s Cluster Basic Performance Tuning](../../../07_references/05_appendix.md#kubernetes-cluster-basic-performance-tuning).

  - Have a shared storage system.

      Some processes of the resumable training feature rely on reading stored data, such as loading CKPT, starting training, and compilation cache loading. Storage performance affects the overall recovery time of resumable training. To avoid deterioration of training recovery time, it is recommended to optimize the storage performance configuration. The recommended configuration provided below takes a 10,000-card scale cluster as an example.
      - 8k I/O read IOPS: \>1024W
      - 8k I/O write IOPS: \>128W
      - Large file sequential read bandwidth: \>288GB/s
      - Large file sequential write bandwidth: \>173GB/s

  - (Optional) Expand shared memory: When using the resumable training feature, it is recommended to expand memory. Please add parameters as commented, as shown in the example below.

      ```yaml
      ...
              volumeMounts:                             # Resource expansion for resumable training
               - name: shm
                 mountPath: /dev/shm
              volumes:
              - name: shm
                emptyDir:
                  medium: Memory
                  sizeLimit: 16Gi
      ...
      ```

  - (Optional) Configure CPU and memory resources: If you need to configure CPU and Memory resources, refer to the following example to manually add the `cpu`"` and `memory` parameters and their corresponding values. Configure the specific values according to actual conditions.

      ```yaml
      ...
                resources:
                  requests:
                    huawei.com/Ascend910: 8
                    cpu: 1000m
                    memory: 100Gi
                  limits:
                    huawei.com/Ascend910: 8
                    cpu: 1000m
                    memory: 100Gi
      ...
      ```

## Fault Recovery Time Description<a name="ZH-CN_TOPIC_0000002479386472"></a>

For the fault recovery stages, composition of training loss, and the definitions and calculation methods of T<sub>0</sub> to T<sub>4</sub>, see [Fault Recovery Time and Training Loss](./01_solutions_principles/02_recovery_acceleration.md#fault-recovery-time-and-training-loss).

**Training Recovery Duration Reference<a name="zh-cn_topic_0000002003001306_section1672017599123"></a>**

Taking the GPT-3 model under the PyTorch framework as an example, this model runs a single-node eight-processor job with parameter sizes of 3B or 15B, under the condition that the NFS storage read and write speeds are 2.7 GB/s for writing and 4.8 GB/s for reading. **(The fault handling mode is rescheduling. If the graceful fault tolerance mode is used, this metric may not be referenced.)**

- For a parameter size of 3B, as shown in [Figure 4](#zh-cn_topic_0000002003001306_fig175521679432), the checkpoint write-to-disk time for this model is approximately **30** seconds. The fault discovery phase of resumable training takes less than **5** seconds, the fault processing phase takes less than **30** seconds, the training restart phase takes approximately **70** seconds, and the checkpoint loading function during the training restart phase takes approximately **3** seconds.
- For a parameter size of 15B, as shown in [Figure 5](#zh-cn_topic_0000002003001306_fig10995142020518), the checkpoint write time for this model is approximately **120** seconds. During resumable training, the fault discovery phase takes less than **5** seconds, the fault processing phase takes less than **30** seconds, the training restart phase takes approximately **210** seconds, and checkpoint loading during the training restart phase takes approximately **90** seconds.

**Figure 4** 3B model time metrics<a name="zh-cn_topic_0000002003001306_fig175521679432"></a>

![](../../../../figures/scheduling/time-metrics-of-3b-model.png)

**Figure 5**  15B model time metrics<a name="zh-cn_topic_0000002003001306_fig10995142020518"></a>

![](../../../../figures/scheduling/time-metrics-of-15b-model.png)
