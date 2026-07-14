# Feature Introduction

## Containerization<a name="ZH-CN_TOPIC_0000002479386930"></a>

**Overview<a name="section1788818281655"></a>**

The NPU containerization is provided for all training or inference jobs to automatically mount required files and device dependencies, so that AI jobs can smoothly run as Docker containers on Ascend devices.

**Required Components<a name="section15655185785119"></a>**

Ascend Docker Runtime

**How to Use<a name="section1245612501584"></a>**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
2. See [Containerization](../usage/containerization/00_before_you_start.md) for detailed instructions.

## Resource Monitoring<a name="ZH-CN_TOPIC_0000002479386910"></a>

**Overview<a name="section1788818281655"></a>**

The Ascend AI processor resources can be monitored in real time during training or inference job execution, including usage, temperature, voltage, memory, and allocation status in containers. It can also monitor the vNPU AI Core usage, total vNPU memory, and used vNPU memory. Currently, NPU Exporter can only monitor vNPU resources of Atlas inference product.

**Required Components<a name="section15655185785119"></a>**

NPU Exporter

**How to Use<a name="section1245612501584"></a>**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
2. See [Resource Monitoring](../usage/resource_monitoring/00_before_you_start.md) for detailed instructions.

## Virtual Instance<a name="ZH-CN_TOPIC_0000002511346855"></a>

### HDK-based Virtual Instance<a name="ZH-CN_TOPIC_0000002511346855hdk"></a>

**Overview<a name="section1337420477275"></a>**

This feature allows multiple users to share one server and to allocate vNPU resources as needed, making the NPU computing power more accessible and affordable.

**Required Components<a name="ZH-CN_TOPIC_0000002479226932"></a>**

The required components vary depending on the vNPU creation or mounting method.

Components required for creating vNPUs:

You can create vNPUs in either of the following ways:

- Static virtualization: Use the npu-smi tool to **manually** create vNPUs.
- Dynamic virtualization: Use the following MindCluster components to create vNPUs.
    - Method 1: Use Ascend Docker Runtime to **manually** create vNPUs. When the container process ends, the vNPUs are automatically destroyed.
    - Method 2: Use Volcano and Ascend Device Plugin to **automatically** create vNPUs. When the container process ends, the vNPUs are automatically destroyed.

Components required for mounting vNPUs:

You can use either of following methods to mount vNPUs to a container.

- Native Docker (only for static virtualization)
- MindCluster components (for static and dynamic virtualization)
    - Method 1: Use Ascend Docker Runtime and Docker to mount vNPUs. (This method is more convenient than using only the native Docker.)
    - Method 2: Use Kubernetes to mount vNPUs.

**How to Use<a name="section1350915844811"></a>**

- Install Docker by referring to [Install Docker Engine](https://docs.docker.com/engine/install/).
- Install Kubernetes by referring to [Installing Kubernetes with deployment tools](https://kubernetes.io/docs/setup/production-environment/tools/).
- For details about how to use this feature, see [HDK-based Virtual Instance](../usage/virtual_instance/virtual_instance_with_hdk/01_description.md).

### vCANN-RT-based Virtual Instance<a name="ZH-CN_TOPIC_0000002511346855vcann"></a>

**Overview<a name="section1337420477275vcann"></a>**

It allows multiple users to share one server. Users can allocate NPU resources on demand, lowering the threshold and cost for using NPU computing power.

**Required Components<a name="ZH-CN_TOPIC_0000002479226932vcann"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD

**How to Use<a name="section1350915844811vcann"></a>**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
2. For details about how to use this feature, see [vCANN-RT-based Virtual Instance](../usage/virtual_instance/virtual_instance_with_vcann_rt/00_description.md).

## Basic Scheduling<a name="ZH-CN_TOPIC_0000002511346871"></a>

### Full-NPU Scheduling

**Overview<a name="section1788818281655"></a>**

When running a training or inference job, you can schedule it to the entire NPU of a node and exclusively occupy the NPU to execute the job. This feature uses the basic scheduling function supported by Kubernetes and works with Volcano or other schedulers to select proper NPUs based on the physical topology of NPUs. This maximizes NPU performance, schedules NPUs for training or inference jobs, and optimally allocates other resources.

Volcano can be used to implement switch affinity scheduling and Ascend AI Processor-based affinity scheduling. Volcano is a scheduler that is developed based on the interconnection topology and processing logic of Ascend AI Processors and fully utilizes Ascend AI Processor capabilities. It can maximize the computing performance of Ascend AI Processors. For details about switch affinity scheduling and Ascend AI Processor-based affinity scheduling, see [Affinity Scheduling](../usage/basic_scheduling/01_affinity_scheduling/00_solution_description.md).

**Required Components<a name="section15655185785119"></a>**

- Volcano or other schedulers
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- Infer Operator
- ClusterD
- NodeD

**How to Use<a name="section1245612501584"></a>**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
2. For details about how to use this feature, see [Full-NPU Scheduling or Static vNPU Scheduling (Training)](../usage/basic_scheduling/03_full_npu_scheduling_and_static_vnpu_scheduling_training.md).

### Static vNPU Scheduling<a name="ZH-CN_TOPIC_0000002511426831"></a>

**Overview<a name="section1788818281655"></a>**

When running a training or inference job, you can schedule the job to vNPUs of a node for training or inference. This feature uses the basic scheduling function supported by Kubernetes and works with Volcano or other schedulers to schedule vNPUs for training or inference jobs and optimally allocate other resources.

**Required Components<a name="section15655185785119"></a>**

The following components need to be installed for training and inference jobs:

- Volcano or other schedulers
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- Infer Operator
- ClusterD
- NodeD

**How to Use<a name="section1245612501584"></a>**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
2. For details about how to use this feature, see [Full-NPU Scheduling or Static vNPU Scheduling (Training)](../usage/basic_scheduling/03_full_npu_scheduling_and_static_vnpu_scheduling_training.md).

### Multi-Level Scheduling

**Overview**

Multi-level scheduling is an advanced scheduling policy introduced by the ascend-for-volcano plugin, specifically designed for NPU clusters with complex network topologies. As a special case of full-NPU scheduling, it abstracts cluster resources into a multi-level structure based on the network topology hierarchy of NPUs, providing efficient, flexible, and reliable scheduling capabilities for NPU clusters. When running a training job, users can select appropriate NPUs based on the network topology to maximize NPU performance and achieve optimal NPU allocation for the training job.

**Required Component**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD
- NodeD

**How to Use**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
2. For details about how to use this feature, see [Multi-Level Scheduling](../usage/basic_scheduling/05_multi_level_scheduling.md).

### Dynamic vNPU Scheduling<a name="ZH-CN_TOPIC_0000002479226956"></a>

**Overview<a name="section1788818281655"></a>**

This feature requires Ascend Device Plugin to report the available number of AI Cores of the node where it is installed. After a virtualization task is reported, Volcano schedules the task to a node that meets the task requirements. After receiving the request, Ascend Device Plugin of the node automatically splits vNPUs and mounts the task to complete the entire dynamic virtualization process. In this process, you do not need to split vNPUs in advance, and vNPUs can be automatically reclaimed after the task is finished. This process supports scenarios where your requirements on computing power change continuously.

**Required Components<a name="section15655185785119"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD

**How to Use<a name="section1245612501584"></a>**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
2. See [Dynamic vNPU Scheduling (Inference)](../usage/virtual_instance/virtual_instance_with_hdk/01_description.md) for detailed instructions.

### Soft Partitioning-based Scheduling

**Overview**

This feature requires Ascend Device Plugin to the available AI Core percentage of the node where it is installed. After a virtualization task is reported, Volcano schedules the task to a node that meets the task requirements. Ascend Device Plugin generates a soft partitioning configuration file based on the task configuration and mounts the file to the task container for [vCANN-RT](https://gitcode.com/openeuler/ubs-virt/blob/master/ubs-virt-enpu/vcann-rt/README.md) to use. With this feature, users can allocate NPU resources on demand, allowing for refined management and dynamic allocation and enabling multiple users to share NPU resources of a single server. This virtualization management mode provides unified resource allocation and reclamation capabilities, meeting the dynamic operation requirements of repeatedly allocating and deallocating resources in multi-tenant scenarios and improving resource utilization.

**Required Component**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD

**How to Use**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
2. See [Soft Partitioning-based Scheduling (Inference)](../usage/virtual_instance/virtual_instance_with_vcann_rt/00_description.md) for detailed instructions.

### Elastic Training<a name="ZH-CN_TOPIC_0000002479226936"></a>

>[!NOTE]
>This section describes the elastic training capabilities based on Resilience Controller. However, this component has been deprecated and its documentation will be removed in the version released on September 30, 2026. For details about the latest elastic training capabilities, see [Elastic Training](../usage/resumable_training/01_solutions_principles.md#elastic-training).

**Overview<a name="section1788818281655"></a>**

If a training node fails, the cluster scheduling components isolate the failed node, reset the number of job replicas based on the preset job scale and the number of available nodes in the current cluster, and perform rescheduling and retraining (script adaptation is required).

**Required Components<a name="section15655185785119"></a>**

- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- Volcano
- NodeD
- Resilience Controller
- ClusterD

**How to Use<a name="section1245612501584"></a>**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.

### Recovery of Inference Card Faults<a name="ZH-CN_TOPIC_0000002479226952"></a>

**Overview<a name="section113779818313"></a>**

If an inference NPU resource managed by the cluster scheduling components is faulty, a hot reset is performed on the faulty resource to restore the NPU.

**Required Components<a name="section143231032154719"></a>**

- Volcano or other schedulers
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD

**How to Use<a name="section74221327111220"></a>**

- See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
- See *Recovery of Inference Card Faults* for detailed instructions.

### Rescheduling Upon Inference Card Faults<a name="ZH-CN_TOPIC_0000002511346875"></a>

**Overview<a name="section119259203315"></a>**

If an inference NPU resource managed by the cluster scheduling components is faulty, the faulty NPU is isolated and automatically rescheduled.

**Required Components<a name="section15655185785119"></a>**

- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- Infer Operator
- Volcano
- ClusterD
- NodeD

**How to Use<a name="section18894171918127"></a>**

- See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
- See *Rescheduling Upon Inference Card Faults* for detailed instructions.

## Resumable Training<a name="ZH-CN_TOPIC_0000002511346867"></a>

**Overview<a name="section1788818281655"></a>**

When a training job is faulty, the job can be rescheduled to a healthy device for training or the faulty chip can be automatically recovered.

- Fault detection: Ascend Device Plugin, Volcano, ClusterD, and NodeD are used to detect job faults.
- Fault handling: After a fault occurs, rectify the fault based on the reported fault information. The following two fault handling modes are supported:
    - Rescheduling mode: After a fault occurs, jobs are rescheduled to other healthy devices.
    - Graceful fault tolerance mode: If a chip is faulty during training, the system attempts to automatically recover the faulty chip.
- Training recovery: After a training job is rescheduled, the checkpoint that is automatically saved before the fault occurs is used to resume training.

**Required Components<a name="section15655185785119"></a>**

- Volcano
- Ascend Operator
- Ascend Device Plugin
- Ascend Docker Runtime
- NodeD
- ClusterD
- TaskD
- (Optional) MindIO ACP
- (Optional) MindIO TFT

**How to Use<a name="section1245612501584"></a>**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
2. For details about how to use this feature, see [Resumable Training](../usage/resumable_training/00_feature_description.md).
3. TaskD must be installed in a container. For details, see [Creating an Image](../usage/resumable_training/07_using_resumable_training_on_the_cli.md).
4. For details about MindIO ACP and its installation procedure, see [Optimizing Checkpoint Saving and Loading](../optimizing_saving_and_loading_checkpoints/01_product_description.md).
5. For details about MindIO TFT and its installation procedure, see [Fault Recovery Acceleration](../fault_recovery_acceleration/01_product_description.md).

## Container Recovery<a name="ZH-CN_TOPIC_0000002492192948"></a>

**Overview<a name="section1788818281655"></a>**

In the scenario where Kubernetes is not deployed, if the training or inference process is abnormal, you can configure the container recovery function to recover containers.

- Fault detection: Container Manager is used to detect job faults.
- Fault rectification: After a fault occurs, the faulty device can be automatically recovered without manual intervention.
- Container recovery: When a fault occurs, the container is stopped. After the fault is rectified, the container is started again.

**Required Components<a name="section15655185785119"></a>**

Container Manager

**How to Use<a name="section1245612501584"></a>**

1. See [Installation and Deployment](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for detailed instructions.
2. For details about how to use this feature, see [Appliance Feature Guide](../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md).
