# Feature Description

## Containerization Support<a name="ZH-CN_TOPIC_0000002479386930"></a>

**Feature Introduction<a name="section1788818281655"></a>**

- Provides NPU containerization support for all training or inference jobs, automatically mounts the required files and device dependencies, and simplifies the container startup command.
- Supports automatic creation and mounting of vNPUs.
- Supports Docker, Containerd, and CRI-O.

**Required Components<a name="section15655185785119"></a>**

Ascend Docker Runtime

**Usage Instructions<a name="section1245612501584"></a>**

1. For component installation, see [Installation and Deployment](../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md).
2. For feature usage guidance, see [Containerization Feature Guide](../04_usage/00_containerization/00_before_you_start.md).

## Resource Monitoring<a name="ZH-CN_TOPIC_0000002479386910"></a>

**Feature Introduction<a name="section1788818281655"></a>**

NPU Resource Monitoring

- Supports real-time monitoring of various data information of Ascend AI Processor resources while executing training or inference jobs. It can obtain in real time information such as Ascend AI Processor utilization, temperature, voltage, memory, and the allocation status of Ascend AI Processors in containers, thereby enabling real-time resource monitoring.
- Supports reporting other metrics through custom plugins.

DPU Resource Monitoring

- Supports monitoring the running status and various statistics of the DPU, including DPU global metrics such as RoCE packet errors, packet loss, received ECN, sent/received CNP, and abnormal PSN retransmission, as well as interface-level metrics such as the link running status, send/receive traffic, and abnormal error status of each NIC port.

**Required Components<a name="section15655185785119"></a>**

Required component for NPU resource monitoring:

- NPU Exporter

Required component for DPU resource monitoring:

- DPU Exporter

**Usage Instructions<a name="section1245612501584"></a>**

1. For component installation, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage guidance, see [Resource Monitoring Feature Guide](../04_usage/01_resource_monitoring/menu_resource_monitoring.md).

## Fault Detection<a name="ZH-CN_TOPIC_FAULT_DETECTION"></a>

**Feature Introduction<a name="section_fault_detection_features"></a>**

The fault detection capability obtains the fault status of the cluster and of training or inference jobs, so that downstream features such as resumable training, chip fault recovery, and rescheduling can handle faults based on the detection result. Currently, the fault detection feature mainly provides the detection of Ascend hardware faults, training job faults, and faults reported by other fault providers.

**Required Components<a name="section_fault_detection_components"></a>**

The required components vary by fault type. For details, see [Fault Detection Feature Guide](../04_usage/04_fault_detection_and_diagnosis/00_feature_description.md).

**Usage Instructions<a name="section_fault_detection_usage"></a>**

1. For component installation, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage guidance, see [Fault Detection Feature Guide](../04_usage/04_fault_detection_and_diagnosis/00_feature_description.md).

## Virtual Instance<a name="ZH-CN_TOPIC_0000002511346855"></a>

A virtual instance can divide one NPU into multiple parts and allocate them to different jobs. Based on the division method, virtual instances are classified into HDK-based virtual instances and vCANN-RT-based virtual instances.

### HDK-Based Virtual Instance<a name="ZH-CN_TOPIC_0000002511346855hdk"></a>

**Feature Introduction<a name="section1337420477275"></a>**

This feature refers to dividing NPUs (Ascend AI Processors) configured on a physical machine or virtual machine into multiple vNPUs (virtual NPUs) through resource virtualization and mounting them into containers for use. The virtualization management approach enables the allocation and reclamation of resources of different specifications, satisfying the resource operation requests of multiple users repeatedly applying for and releasing resources.

**Required Components<a name="ZH-CN_TOPIC_0000002479226932"></a>**

- Static virtualization: **manually** create multiple vNPUs using the npu-smi tool. Scheduling is performed based on vNPUs of fixed size and fixed quantity.
- Dynamic virtualization: Volcano works with Ascend Device Plugin to dynamically and **automatically** create vNPUs, which are automatically destroyed before the container is destroyed.

Required components for static virtualization:

- Ascend Docker Runtime
- Ascend Device Plugin

Required components for dynamic virtualization:

- Ascend Docker Runtime
- Ascend Device Plugin
- Volcano

**Usage Instructions<a name="section1350915844811"></a>**

- For installation, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
- For feature usage instructions, see [HDK-Based Virtual Instance](../04_usage/02_virtual_instance/00_virtual_instance_with_hdk/01_description.md).

### vCANN-RT-Based Virtual Instance<a name="ZH-CN_TOPIC_0000002511346855vcann"></a>

**Feature Introduction<a name="section1337420477275vcann"></a>**

This feature based on vCANN-RT mounts NPUs (Ascend AI Processors) configured on a physical machine into a container by providing a soft partitioning configuration file to vCANN-RT. The virtualization management approach enables allocation and reclamation of resources of different specifications, satisfying the operational requests of multiple users repeatedly applying for and releasing resources.

**Required Components<a name="ZH-CN_TOPIC_0000002479226932vcann"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD

**Usage Instructions<a name="section1350915844811vcann"></a>**

1. For installation components, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage instructions, see [vCANN-RT-Based Virtual Instance](../04_usage/02_virtual_instance/01_virtual_instance_with_vcann_rt/00_description.md).

## Basic Scheduling<a name="ZH-CN_TOPIC_0000002511346871"></a>

### Full-NPU Scheduling<a name="ZH-CN_TOPIC_0000002479386926"></a>

**Feature Introduction<a name="section1788818281655"></a>**

When users run training or inference jobs, the jobs can be scheduled to an entire NPU on a node, exclusively occupying the whole NPU to execute the training or inference jobs. The full-NPU scheduling feature leverages the basic scheduling capabilities supported by Kubernetes (K8s), works with Volcano or other schedulers, and selects appropriate NPU devices based on the physical topology of NPU devices, thereby maximizing NPU performance and achieving optimal allocation of NPU cards and other resources for training or inference jobs.

The more complex the network topology between chips, the more complex the scheduling logic. For details, see [Affinity Scheduling](../04_usage/03_basic_scheduling/01_affinity_scheduling/00_solution_description.md).

Preempt and Reclaim Action operations are supported. Preempt is used for resource preemption between jobs in the same queue. When a high-priority job needs resources, it can preempt resources from a low-priority job. Reclaim Action is used for resource reclamation between different queues. When a job in a queue needs resources and the queue has not exceeded its resource quota, resources can be reclaimed from other reclaimable queues. Both enable dynamic adjustment and optimized allocation of resources. For details about Preempt and Reclaim Action, see [relevant information on the official Volcano website](https://volcano.sh/docs/Scheduler/Actions).

With the Ascend Dynamic Resource Allocation component, you can request a whole NPU through the Kubernetes Dynamic Resource Allocation (DRA) mechanism, and service containers request NPU resources through ResourceClaim. DRA-based scheduling does not currently support advanced scheduling features such as affinity scheduling and dynamic virtualization; only basic full-NPU scheduling based on K8s and Volcano 1.15 is supported.

**Required Components<a name="section15655185785119"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- Infer Operator
- ClusterD
- NodeD

**Usage Instructions<a name="section1245612501584"></a>**

1. For component installation, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage instructions, see [Full-NPU scheduling](../04_usage/03_basic_scheduling/03_full_npu_scheduling.md).
3. For usage examples of Preempt and Reclaim Action, see [Tidal Scheduling Best Practices](../04_usage/11_tidal_scheduling/00_before_you_start.md).
4. For usage examples of DRA-based scheduling, see [DRA Scheduling Best Practices](../04_usage/13_ascend_dynamic_resource_allocation_best_practice/00_before_you_start.md).

### Multi-Level Scheduling<a name="ZH-CN_TOPIC_0000002511346873"></a>

**Feature Highlights<a name="section1788818281655"></a>**

Multi-level scheduling is an advanced scheduling strategy of the ascend-for-volcano plugin, designed specifically for NPU clusters with complex network topologies. Multi-level scheduling is a special scenario of full-NPU scheduling. It abstracts cluster resources into a multi-level structure based on the hierarchical relationships of the NPU network topology, providing efficient, flexible, and reliable scheduling capabilities for NPU clusters. When users run training jobs, they select appropriate NPU devices based on the network topology to maximize NPU performance and achieve optimal allocation of NPU scheduling for training jobs.

**Required Components<a name="section15655185785119"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD
- NodeD

**Usage Instructions<a name="section1245612501584"></a>**

1. For installation components, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage guidance, see [Multi-level Scheduling](../04_usage/03_basic_scheduling/04_multi_level_scheduling.md).

### Chip Fault Recovery<a name="ZH-CN_TOPIC_0000002479226952"></a>

**Feature Introduction<a name="section113779818313"></a>**

When NPU resources managed by the cluster scheduling components fail, a hot reset is performed on the faulty resource (the corresponding NPU) to restore the NPU to health.

**Required Components<a name="section143231032154719"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD

**Usage Instructions<a name="section74221327111220"></a>**

1. For installation components, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage instructions, see [Chip Fault Recovery](../04_usage/05_fault_recovery/00_chip_fault_recovery.md).

### Rescheduling<a name="ZH-CN_TOPIC_0000002511346875"></a>

**Feature Introduction<a name="section119259203315"></a>**

After a fault occurs on an NPU resource managed by the cluster scheduling component, the cluster scheduling component isolates the faulty resource (the corresponding NPU) and automatically reschedules it.

**Required Components<a name="section15655185785119"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD
- Ascend Operator
- Infer Operator

**Usage Instructions<a name="section18894171918127"></a>**

1. For component installation, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage guidance, see [Rescheduling](../04_usage/03_basic_scheduling/05_rescheduling.md).

## Resumable Training<a name="ZH-CN_TOPIC_0000002511346867"></a>

**Feature Introduction<a name="section1788818281655"></a>**

When a fault occurs on a training job, resumable training relies on the fault events, fault levels, and resource status provided by the fault detection feature to reschedule the job to a healthy device to continue training, or to automatically recover the faulty chip.

- **Fault detection (prerequisite)**: Job faults are obtained through the fault detection feature. For details, see [Fault Detection Feature Guide](../04_usage/04_fault_detection_and_diagnosis/00_feature_description.md).
- **Fault handling**: After a fault occurs, fault handling is performed based on the reported fault information. It is divided into the following two modes.
  - **Rescheduling mode**: After a fault occurs, the job is rescheduled to other healthy devices to continue running.
  - **Graceful fault tolerance mode**: When a chip fails during training, the system attempts to automatically recover the faulty chip.

- **Recovery acceleration**: After the job is rescheduled, the training job uses the checkpoint automatically saved before the fault to restart the training job and continue training.

**Required Components<a name="section15655185785119"></a>**

- Volcano
- Ascend Operator
- Ascend Device Plugin
- Ascend Docker Runtime
- NodeD
- ClusterD
- TaskD
- MindIO ACP (optional)
- MindIO TFT (optional)

**Usage Instructions<a name="section1245612501584"></a>**

1. For component installation, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage instructions, see [Resumable Training](../04_usage/05_fault_recovery/01_resumable_training/00_feature_description.md).
3. TaskD must be installed inside the container. For details, see [Building an Image](../04_usage/05_fault_recovery/01_resumable_training/03_examples_and_verification/01_pytorch_examples_and_verification.md#building-an-image).
4. For a detailed introduction to MindIO ACP and its installation steps, see [Checkpoint Saving and Loading Optimization](../07_references/01_optimizing_saving_and_loading_checkpoints/01_product_description.md).
5. For a detailed introduction to MindIO TFT and its installation steps, see [Fault Recovery Acceleration](../07_references/00_fault_recovery_acceleration/01_product_description.md).

## Container Recovery<a name="ZH-CN_TOPIC_0000002492192948"></a>

**Feature Introduction<a name="section1788818281655"></a>**

In scenarios without K8s, when a training or inference process becomes abnormal, container fault recovery can be performed by configuring the container recovery feature.

- **Fault detection**: Detects job faults through the Container Manager component.
- **Fault handling**: After a fault occurs, the faulty device is automatically recovered without manual intervention.
- **Container recovery**: When a fault occurs, the container is stopped, and after the fault is recovered, the container is restarted.

**Required Components<a name="section15655185785119"></a>**

Container Manager

**Usage Instructions<a name="section1245612501584"></a>**

1. For installation components, see [Installation and Deployment](../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md).
2. For feature usage instructions, see [Appliance Feature Guide](../04_usage/06_appliance/01_npu_hardware_fault_detection_and_rectification.md).

## Container Snapshot<a name="ZH-CN_TOPIC_0000002511346881"></a>

**Feature Introduction<a name="section1788818281655"></a>**

This feature implements the container snapshot capability for inference services, supporting rapid startup of large-model inference services and fast recovery in fault scenarios. Through the collaboration of MindCluster's Infer Operator, NodeD, and Ascend Docker Runtime components, host-side and device-side snapshots are generated after the inference job completes warm up, and services are quickly recovered through snapshots after a Pod is abnormally deleted, reducing the inference service startup time from over 30 minutes to the minute level.

**Required Components<a name="section15655185785119"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD
- Infer Operator

**Usage Instructions<a name="section1245612501584"></a>**

1. For installation components, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage guidance, see [Container Snapshot Deployment and Usage](../04_usage/10_infer_operator_best_practice/06_container_snapshot_usage.md).

## Inference High Availability<a name="ZH-CN_TOPIC_0000002511346885"></a>

**Feature Introduction<a name="section2288818281659"></a>**

The inference high availability module provides best practices for inference jobs, covering scenarios such as deployment, rescheduling, and elastic scaling of multiple inference engines, and supports users in achieving high-availability deployment of inference services in production environments. Its main functions include:

- **Inference engine support**: Supports best practices for multiple inference engines such as vLLM, MindIE CMotor, and SGLang.
- **Scheduling capability**: Supports priority scheduling configuration and switch affinity configuration for inference jobs.
- **Fault handling**: Supports fault rescheduling, elastic scaling, and fault isolation for inference jobs.

**Required Components<a name="section33655185785111"></a>**

- Infer Operator
- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD

**Usage Instructions<a name="section3245612501586"></a>**

1. For installation components, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For Infer Operator inference job best practices, see [Infer Operator Inference Job Best Practices](../04_usage/10_infer_operator_best_practice/00_before_you_start.md).
3. For MindIE CMotor inference job best practices (sunset), see [MindIE CMotor Inference Job Best Practices (Sunset)](../04_usage/07_mindie_motor_best_practice/00_before_you_start.md).
4. For SGLang inference job best practices, see [SGLang Inference Job Best Practices](../04_usage/08_sglang_best_practice/00_before_you_start.md).
5. For vLLM inference job best practices, see [vLLM Inference Job Best Practices](../04_usage/09_vllm_best_practice/00_before_you_start.md).

## ClusterOps Agent<a name="ZH-CN_TOPIC_0000002524312690"></a>

**Feature Introduction<a name="section1788818281655"></a>**

When a training or inference job fails during running, the fault root cause must be located quickly. Jobs are usually distributed across multiple nodes, and fault-related logs (system logs, device logs, service logs, plog, and so on) are scattered on the host machine of each node, making manual troubleshooting costly and inefficient. MindCluster provides the ClusterOps Agent feature, which triggers diagnosis for a specific job with the `kubectl ascend-diag` command and automatically completes the whole process of node log collection, log parsing, and centralized diagnosis, helping users quickly locate the fault root cause of a job.

- **One-click diagnosis**: Trigger diagnosis for a specific job with the `kubectl ascend-diag` command, without logging in to each node to manually collect logs.
- **Automatic collection scheduling**: Agent Core acts as the centralized control center for fault diagnosis, automatically maintains the mapping between jobs and Pods, and delivers collection instructions to the Node Collector on each compute node.
- **Local parsing and reporting**: Node Collector collects job logs from the host machine according to the collection contract, locally calls `ascend-fd parse` to parse the logs, and packages and reports the parsing result to Agent Core.
- **Centralized diagnosis**: Agent Core aggregates the parsing results reported by each node, assembles them into a diagnosis input directory, and calls `ascend-fd diag` to perform centralized diagnosis and generate a diagnosis report.
- **Result caching**: The diagnosis result is cached after each diagnosis. Repeated diagnosis directly returns the cached result, and the `--refresh` parameter can be used to force a refresh.
- **Intelligent summarization**: Optionally connects to an LLM service to intelligently summarize the diagnosis report and output a root cause report.

**Required Components<a name="section15655185785119"></a>**

- Agent Core
- Node Collector
- Kubectl Plugin (kubectl-ascend_diag, kubectl-clusterops)
- ascend-faultdiag
- LLM service (optional)

**Usage Instructions<a name="section1245612501584"></a>**

1. For component installation, see [Manual Installation](../03_installation_guide/02_installation/01_manual_installation.md).
2. For feature usage guidance, see [ClusterOps Agent](../04_usage/14_clusterops_agent/00_before_you_start.md).
