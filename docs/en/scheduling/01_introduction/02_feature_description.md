# Feature Description

<!-- md-trans-meta sourceCommit=c1a0dfa4983002031992997886051605eb062522 translatedAt=2026-08-27T02:19:13.810Z pushedAt=2026-08-27T02:26:43.513Z -->

## Containerization Support<a name="ZH-CN_TOPIC_0000002479386930"></a>

**Feature Introduction<a name="section1788818281655"></a>**

- Provides NPU containerization support for all training or inference jobs, automatically mounts the required files and device dependencies, and simplifies the container startup command.
- Supports automatic creation and mounting of vNPUs.
- Supports Docker and Containerd.

**Required Components<a name="section15655185785119"></a>**

Ascend Docker Runtime

**Usage Instructions<a name="section1245612501584"></a>**

1. For component installation, see [Installation and Deployment](../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md).
2. For feature usage guidance, see [Containerization Feature Guide](../04_usage/00_containerization/00_before_you_start.md).

## Resource Monitoring<a name="ZH-CN_TOPIC_0000002479386910"></a>

**Feature Introduction<a name="section1788818281655"></a>**

- Supports real-time monitoring of various data information of Ascend AI Processor resources while executing training or inference tasks. It can obtain in real time information such as Ascend AI Processor utilization, temperature, voltage, memory, and the allocation status of Ascend AI Processors in containers, thereby enabling real-time resource monitoring.
- Supports reporting other metrics through custom plugins.

**Required Components<a name="section15655185785119"></a>**

NPU Exporter

**Usage Instructions<a name="section1245612501584"></a>**

1. For component installation, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage guidance, see [Resource Monitoring](../04_usage/01_resource_monitoring/00_before_you_start.md).

## Virtual Instance<a name="ZH-CN_TOPIC_0000002511346855"></a>

A virtual instance can divide one NPU into multiple parts and allocate them to different tasks. Based on the division method, virtual instances are classified into HDK-based virtual instances and vCANN-RT-based virtual instances.

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

When users run training or inference tasks, the tasks can be scheduled to an entire NPU on a node, exclusively occupying the whole NPU to execute the training or inference task. The full-NPU scheduling feature leverages the basic scheduling capabilities supported by Kubernetes (K8s), works with Volcano or other schedulers, and selects appropriate NPU devices based on the physical topology of NPU devices, thereby maximizing NPU performance and achieving optimal allocation of NPU cards and other resources for training or inference tasks.

The more complex the network topology between chips, the more complex the scheduling logic. For details, see [Affinity Scheduling](../04_usage/03_basic_scheduling/01_affinity_scheduling/00_solution_description.md).

Preempt and Reclaim Action operations are supported. Preempt is used for resource preemption between tasks in the same queue. When a high-priority task needs resources, it can preempt resources from a low-priority task. Reclaim is used for resource reclamation between different queues. When a task in a queue needs resources and the queue has not exceeded its resource quota, resources can be reclaimed from other reclaimable queues. Both enable dynamic adjustment and optimized allocation of resources. For details about Preempt and Reclaim Action, see [relevant information on the official Volcano website](https://volcano.sh/docs/Scheduler/Actions).

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
3. For usage examples of Preempt and Reclaim Action, see [Tidal Scheduling Best Practices](../04_usage/10_tidal_scheduling/00_before_you_start.md).
4. The Ascend plugin selects eviction targets based on the task's specific affinity scheduling policy, but the final eviction targets are determined jointly by multiple plugins in the Volcano framework. Therefore, the resources released by the final eviction targets may not meet the requirements of the current task, causing the eviction to fail.

### Multi-Level Scheduling<a name="ZH-CN_TOPIC_0000002511346873"></a>

**Feature Highlights<a name="section1788818281655"></a>**

Multi-level scheduling is an advanced scheduling strategy of the ascend-for-volcano plugin, designed specifically for NPU clusters with complex network topologies. Multi-level scheduling is a special scenario of full-NPU scheduling. It abstracts cluster resources into a multi-level structure based on the hierarchical relationships of the NPU network topology, providing efficient, flexible, and reliable scheduling capabilities for NPU clusters. When users run training tasks, they select appropriate NPU devices based on the network topology to maximize NPU performance and achieve optimal allocation of NPU scheduling for training tasks.

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

### Recovery of Inference Card Faults<a name="ZH-CN_TOPIC_0000002479226952"></a>

**Feature Introduction<a name="section113779818313"></a>**

When an inference NPU resource managed by the cluster scheduling component fails, a hot reset is performed on the faulty resource (the corresponding NPU) to restore the NPU to a healthy state.

**Required Components<a name="section143231032154719"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD

**Usage Instructions<a name="section74221327111220"></a>**

- For installation components, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
- For feature usage instructions, see [Recovery of Inference Card Faults](../04_usage/03_basic_scheduling/06_recovery_of_inference_card_faults.md).

### Rescheduling upon Inference Card Faults<a name="ZH-CN_TOPIC_0000002511346875"></a>

**Feature Introduction<a name="section119259203315"></a>**

After a fault occurs on an inference NPU resource managed by the cluster scheduling component, the cluster scheduling component isolates the faulty resource (the corresponding NPU) and automatically performs rescheduling.

**Required Components<a name="section15655185785119"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD
- Ascend Operator
- Infer Operator

**Usage Instructions<a name="section18894171918127"></a>**

- For component installation, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
- For feature usage guidance, see [Rescheduling upon Inference Card Faults](../04_usage/03_basic_scheduling/05_rescheduling_upon_inference_card_faults.md).

## Resumable Training<a name="ZH-CN_TOPIC_0000002511346867"></a>

**Feature Introduction<a name="section1788818281655"></a>**

When a training task encounters a fault, the task is rescheduled to a healthy device to continue training, or the faulty chip is automatically recovered.

- **Fault detection**: Task faults are detected through Ascend Device Plugin, Volcano, ClusterD, and NodeD.
- **Fault handling**: After a fault occurs, fault handling is performed based on the reported fault information. It is divided into the following two modes.
  - **Rescheduling mode**: After a fault occurs, the task is rescheduled to other healthy devices to continue running.
  - **Graceful fault tolerance mode**: When a chip fails during training, the system attempts to automatically recover the faulty chip.

- **Recovery acceleration**: After the task is rescheduled, the training task uses the checkpoint automatically saved before the fault to restart the training task and continue training.

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
2. For feature usage instructions, see [Resumable Training](../04_usage/04_resumable_training/00_feature_description.md).
3. TaskD must be installed inside the container. For details, see [Building an Image](../04_usage/04_resumable_training/04_using_resumable_training_on_the_cli.md#building-an-image).
4. For a detailed introduction to MindIO ACP and its installation steps, see [Checkpoint Saving and Loading Optimization](../07_references/01_optimizing_saving_and_loading_checkpoints/01_product_description.md).
5. For a detailed introduction to MindIO TFT and its installation steps, see [Fault Recovery Acceleration](../07_references/00_fault_recovery_acceleration/01_product_description.md).

## Container Recovery<a name="ZH-CN_TOPIC_0000002492192948"></a>

**Feature Introduction<a name="section1788818281655"></a>**

In scenarios without K8s, when a training or inference process becomes abnormal, container fault recovery can be performed by configuring the container recovery feature.

- **Fault detection**: Detects task faults through the Container Manager component.
- **Fault handling**: After a fault occurs, the faulty device is automatically recovered without manual intervention.
- **Container recovery**: When a fault occurs, the container is stopped, and after the fault is recovered, the container is restarted.

**Required Components<a name="section15655185785119"></a>**

Container Manager

**Usage Instructions<a name="section1245612501584"></a>**

1. For installation components, see [Installation and Deployment](../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md).
2. For feature usage instructions, see [Appliance Feature Guide](../04_usage/05_appliance/01_npu_hardware_fault_detection_and_rectification.md).

## Container Snapshot<a name="ZH-CN_TOPIC_0000002511346881"></a>

**Feature Introduction<a name="section1788818281655"></a>**

This feature implements the container snapshot capability for inference services, supporting rapid startup of large-model inference services and fast recovery in fault scenarios. Through the collaboration of MindCluster's Infer Operator, NodeD, and Ascend Docker Runtime components, host-side and device-side snapshots are generated after the inference task completes warm up, and services are quickly recovered through snapshots after a Pod is abnormally deleted, reducing the inference service startup time from over 30 minutes to the minute level.

**Required Components<a name="section15655185785119"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD
- Infer Operator

**Usage Instructions<a name="section1245612501584"></a>**

1. For installation components, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For feature usage guidance, see [Container Snapshot Deployment and Usage](../04_usage/09_infer_operator_best_practice/06_container_snapshot_usage.md).

## Inference High Availability<a name="ZH-CN_TOPIC_0000002511346885"></a>

**Feature Introduction<a name="section2288818281659"></a>**

The inference high availability module provides best practices for inference tasks, covering scenarios such as deployment, rescheduling, and elastic scaling of multiple inference engines, and supports users in achieving high-availability deployment of inference services in production environments. Its main functions include:

- **Inference engine support**: Supports best practices for multiple inference engines such as vLLM, MindIE Motor, and SGLang.
- **Scheduling capability**: Supports priority scheduling configuration and switch affinity configuration for inference tasks.
- **Fault handling**: Supports fault rescheduling, elastic scaling, and fault isolation for inference tasks.

**Required Components<a name="section33655185785111"></a>**

- Infer Operator
- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD

**Usage Instructions<a name="section3245612501586"></a>**

1. For installation components, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).
2. For Infer Operator inference task best practices, see [Infer Operator Inference Task Best Practices](../04_usage/09_infer_operator_best_practice/00_before_you_start.md).
3. For MindIE Motor inference task best practices, see [MindIE Motor Inference Task Best Practices](../04_usage/06_mindie_motor_best_practice/00_before_you_start.md).
4. For SGLang inference task best practices, see [SGLang Inference Task Best Practices](../04_usage/07_sglang_best_practice/00_before_you_start.md).
5. For vLLM inference task best practices, see [vLLM Inference Task Best Practices](../04_usage/08_vllm_best_practice/00_before_you_start.md).
