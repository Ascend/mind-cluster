# 特性介绍

## 容器化支持<a name="ZH-CN_TOPIC_0000002479386930"></a>

**功能特点<a name="section1788818281655"></a>**

为所有的训练或推理作业提供NPU容器化支持，自动挂载所需文件和设备依赖，使用户AI作业能够以Docker容器的方式平滑运行在昇腾设备之上。

**所需组件<a name="section15655185785119"></a>**

Ascend Docker Runtime

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[容器化支持](../usage/containerization/00_before_you_start.md)章节进行操作。

## 资源监测<a name="ZH-CN_TOPIC_0000002479386910"></a>

**功能特点<a name="section1788818281655"></a>**

支持在执行训练或者推理任务时，对昇腾AI处理器资源各种数据信息的实时监测，可实时获取昇腾AI处理器利用率、温度、电压、内存，以及昇腾AI处理器在容器中的分配状况等信息，实现资源的实时监测。支持对虚拟NPU（vNPU）的AI Core利用率、vNPU总内存和vNPU使用中内存进行监测。

**所需组件<a name="section15655185785119"></a>**

NPU Exporter

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[资源监测](../usage/resource_monitoring/00_before_you_start.md)章节进行操作。

## 虚拟化实例<a name="ZH-CN_TOPIC_0000002511346855"></a>

### 基于HDK的虚拟化实例<a name="ZH-CN_TOPIC_0000002511346855hdk"></a>

**功能介绍<a name="section1337420477275"></a>**

基于HDK的虚拟化实例功能是指通过资源虚拟化的方式将物理机或虚拟机配置的NPU（昇腾AI处理器）切分成若干份vNPU（虚拟NPU）挂载到容器中使用，虚拟化管理方式能够实现统一不同规格资源的分配和回收处理，满足多用户反复申请/释放的资源操作请求。

昇腾基于HDK的虚拟化实例功能的优点是可实现多个用户共同使用一台服务器，用户可以按需申请vNPU，降低了用户使用NPU算力的门槛和成本。多个用户共同使用一台服务器的NPU，并借助容器进行资源隔离，资源隔离性好，保证运行环境的稳定和安全，且资源分配，资源回收过程统一，方便多租户管理。

**所需组件<a name="ZH-CN_TOPIC_0000002479226932"></a>**

根据创建或挂载vNPU的方式不同，所需组件不同，可以参考如下内容。

创建vNPU所需组件：

创建vNPU有以下两种方式。

- 静态虚拟化：通过npu-smi工具**手动**创建多个vNPU。
- 动态虚拟化：通过MindCluster中的以下组件创建vNPU。
    - 方式一：通过Ascend Docker Runtime**手动**创建vNPU，容器进程结束时，自动销毁vNPU。
    - 方式二：通过Volcano和Ascend Device Plugin动态地**自动**创建vNPU，容器进程结束时，自动销毁vNPU。

挂载vNPU所需组件：

根据创建vNPU的方式的不同，将vNPU挂载到容器的方式也不同，说明如下：

- 基于原生Docker挂载vNPU（只支持静态虚拟化）
- 基于MindCluster组件挂载vNPU（支持静态虚拟化和动态虚拟化）
    - 方式一：通过Ascend Docker Runtime+Docker方式挂载vNPU（此方式相比只使用原生Docker易用性更高）。
    - 方式二：通过Kubernetes挂载vNPU。

**使用说明<a name="section1350915844811"></a>**

- 驱动安装后会默认安装npu-smi工具，安装操作请参见《CANN 软件安装指南》中的“<a href="https://www.hiascend.com/document/detail/zh/canncommercial/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=Debian">安装NPU驱动和固件</a>”章节（商用版）或“<a href="https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=openEuler">安装NPU驱动和固件</a>”章节（社区版）；安装成功后，npu-smi放置在“/usr/local/sbin/”和“/usr/local/bin/”路径下。
- 安装MindCluster中的Ascend Docker Runtime、Ascend Device Plugin和Volcano组件，请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
- 安装Docker，请参见[安装Docker](https://docs.docker.com/engine/install/)。
- 安装Kubernetes，请参见[安装Kubernetes](https://kubernetes.io/zh/docs/setup/production-environment/tools/)。
- 特性使用指导请参见[基于HDK的虚拟化实例](../usage/virtual_instance/virtual_instance_with_hdk/01_description.md)章节进行操作。

### 基于vCANN-RT的虚拟化实例<a name="ZH-CN_TOPIC_0000002511346855vcann"></a>

**功能介绍<a name="section1337420477275vcann"></a>**

基于vCANN-RT的虚拟化实例功能是指通过向vCANN-RT提供软切分配置文件的方式将物理机配置的NPU（昇腾AI处理器）挂载到容器中使用，虚拟化管理方式能够实现统一不同规格资源的分配和回收处理，满足多用户反复申请/释放资源的操作请求。

昇腾基于vCANN-RT的虚拟化实例功能的优点是可实现多个用户共同使用一台服务器，用户可以按需申请NPU的资源，降低了用户使用NPU算力的门槛和成本。多个用户共同使用一台服务器的NPU，并借助容器进行资源隔离，资源隔离性好，保证运行环境的稳定和安全，且资源分配与回收过程统一，从而方便多租户管理。

**所需组件<a name="ZH-CN_TOPIC_0000002479226932vcann"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD

**使用说明<a name="section1350915844811vcann"></a>**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[基于vCANN-RT的虚拟化实例](../usage/virtual_instance/virtual_instance_with_vcann_rt/00_description.md)章节进行操作。

## 基础调度<a name="ZH-CN_TOPIC_0000002511346871"></a>

### 整卡调度<a name="ZH-CN_TOPIC_0000002479386926"></a>

**功能特点<a name="section1788818281655"></a>**

支持用户运行训练或者推理任务时，将训练或推理任务调度到节点的整张NPU卡上，独占整张卡执行训练或者推理任务。整卡调度特性借助Kubernetes（以下简称K8s）支持的基础调度功能，配合Volcano或者其他调度器，根据NPU设备物理拓扑，选择合适的NPU设备，最大化发挥NPU性能，实现训练或者推理任务的NPU卡的调度和其他资源的最佳分配。

使用集群调度组件提供的Volcano组件，可以实现交换机亲和性调度和昇腾AI处理器亲和性调度。Volcano是基于昇腾AI处理器的互联拓扑结构和处理逻辑，实现了昇腾AI处理器最佳利用的调度器组件，可以最大化发挥昇腾AI处理器计算性能。关于交换机亲和性调度和昇腾AI处理器亲和性调度的详细说明，可以参见[亲和性调度](../usage/basic_scheduling/01_affinity_scheduling/00_solution_description.md)。

**所需组件<a name="section15655185785119"></a>**

- 调度器（Volcano或其他调度器）
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- Infer Operator
- ClusterD
- NodeD

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[整卡调度或静态vNPU调度（训练）](../usage/basic_scheduling/03_full_npu_scheduling_and_static_vnpu_scheduling_training.md)章节进行操作。

### 静态vNPU调度<a name="ZH-CN_TOPIC_0000002511426831"></a>

**功能特点<a name="section1788818281655"></a>**

支持用户运行训练或者推理任务时，将训练或推理任务调度到节点的vNPU卡上，使用vNPU执行训练或者推理任务。静态vNPU调度特性借助Kubernetes（以下简称K8s）支持的基础调度功能，配合Volcano或者其他调度器，实现训练或者推理任务的vNPU卡的调度和其他资源的最佳分配。

**所需组件<a name="section15655185785119"></a>**

训练任务及推理任务下需要安装以下组件：

- 调度器（Volcano或其他调度器）
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- Infer Operator
- ClusterD
- NodeD

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[整卡调度或静态vNPU调度（训练）](../usage/basic_scheduling/03_full_npu_scheduling_and_static_vnpu_scheduling_training.md)章节进行操作。

### 多级调度

**功能特点**

多级调度是ascend-for-volcano插件中的一种高级调度策略，专为具有复杂网络拓扑的NPU集群设计。多级调度是整卡调度的一种特殊场景，它根据NPU的网络拓扑层级关系将集群资源抽象为多层级结构，为NPU集群提供高效、灵活、可靠的调度能力。用户运行训练任务时，根据网络拓扑选择合适的NPU设备，最大化发挥NPU性能，实现训练任务的NPU卡调度的最佳分配。

**所需组件**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD
- NodeD

**使用说明**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[多级调度](../usage/basic_scheduling/05_multi_level_scheduling.md)章节进行操作。

### 动态vNPU调度<a name="ZH-CN_TOPIC_0000002479226956"></a>

**功能特点<a name="section1788818281655"></a>**

动态vNPU调度需要Ascend Device Plugin组件上报其所在节点的可用AI Core数目。虚拟化任务上报后，Volcano经过计算将该任务调度到满足其要求的节点。该节点的Ascend Device Plugin在收到请求后自动切分出vNPU设备并挂载该任务，从而完成整个动态虚拟化过程。该过程不需要用户提前切分vNPU，在任务使用完成后又能自动回收，支持用户算力需求不断变化的场景。

**所需组件<a name="section15655185785119"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[动态vNPU调度（推理）](../usage/basic_scheduling/06_dynamic_vnpu_scheduling_inference.md)章节进行操作。

### 软切分调度

**功能特点**

软切分虚拟化调度需要Ascend Device Plugin组件上报其所在节点的可用芯片的AI Core总百分比信息。虚拟化任务上报后，Volcano经过计算将该任务调度到满足其要求的节点，由Ascend Device Plugin根据任务的配置信息生成软切分配置文件并挂载到任务容器中供[vCANN-RT](https://gitcode.com/openeuler/ubs-virt/blob/master/ubs-virt-enpu/vcann-rt/README.md)使用。该功能通过向[vCANN-RT](https://gitcode.com/openeuler/ubs-virt/blob/master/ubs-virt-enpu/vcann-rt/README.md)提供软切分配置文件的方式，使用户可以按需申请NPU（昇腾AI处理器）资源，实现对NPU资源的精细化管理和动态分配，支持多个用户共同使用一台服务器的NPU资源。该虚拟化管理方式具有统一资源分配与回收能力，可满足多租户场景下反复申请、释放资源的动态操作需求，提升资源利用率。

**所需组件**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD

**使用说明**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[软切分调度（推理）](../usage/basic_scheduling/07_soft_allocation_scheduling_inference.md)章节进行操作。

### 弹性训练<a name="ZH-CN_TOPIC_0000002479226936"></a>

>[!NOTE]
>本章节描述的是基于Resilience Controller组件的弹性训练，该组件已经日落，相关资料将于2026年9月30日的版本删除。最新的弹性训练能力请参见[弹性训练](../usage/resumable_training/01_solutions_principles.md#弹性训练)。

**功能特点<a name="section1788818281655"></a>**

训练节点出现故障后，集群调度组件将对故障节点进行隔离，并根据任务预设的规模和当前集群中可用的节点数，重新设置任务副本数，然后进行重调度和重训练（需进行脚本适配）。

**所需组件<a name="section15655185785119"></a>**

- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- Volcano
- NodeD
- Resilience Controller
- ClusterD

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[弹性训练](../usage/basic_scheduling/08_elastic_training.md)章节进行操作。

### 推理卡故障恢复<a name="ZH-CN_TOPIC_0000002479226952"></a>

**功能特点<a name="section113779818313"></a>**

集群调度组件管理的推理NPU资源出现故障后，将对故障资源（对应NPU）进行热复位操作，使NPU恢复健康。

**所需组件<a name="section143231032154719"></a>**

- 调度器（Volcano或其他调度器）
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD

**使用说明<a name="section74221327111220"></a>**

- 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
- 特性使用指导请参见[推理卡故障恢复](../usage/basic_scheduling/10_recovery_of_inference_card_faults.md)章节进行操作。

### 推理卡故障重调度<a name="ZH-CN_TOPIC_0000002511346875"></a>

**功能特点<a name="section119259203315"></a>**

集群调度组件管理的推理NPU资源出现故障后，集群调度组件将对故障资源（对应NPU）进行隔离并自动进行重调度。

**所需组件<a name="section15655185785119"></a>**

- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- Infer Operator
- Volcano
- ClusterD
- NodeD

**使用说明<a name="section18894171918127"></a>**

- 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
- 特性使用指导请参见[推理卡故障重调度](../usage/basic_scheduling/09_rescheduling_upon_inference_card_faults.md)章节进行操作。

## 断点续训<a name="ZH-CN_TOPIC_0000002511346867"></a>

**功能特点<a name="section1788818281655"></a>**

当训练任务出现故障时，将任务重调度到健康设备上继续训练或者对故障芯片进行自动恢复。

- **故障检测**：通过Ascend Device Plugin、Volcano、ClusterD和NodeD四个组件，发现任务故障。
- **故障处理**：故障发生后，根据上报的故障信息进行故障处理。分为以下两种模式。
    - **重调度模式**：故障发生后将任务重调度到其他健康设备上继续运行。
    - **优雅容错模式**：当训练时芯片出现故障后，系统将尝试对故障芯片进行自动恢复。

- **恢复加速**：在任务重新调度之后，训练任务会使用故障前自动保存的CKPT，重新拉起训练任务继续训练。

**所需组件<a name="section15655185785119"></a>**

- Volcano
- Ascend Operator
- Ascend Device Plugin
- Ascend Docker Runtime
- NodeD
- ClusterD
- TaskD
- MindIO ACP（可选）
- MindIO TFT（可选）

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[断点续训](../usage/resumable_training/00_feature_description.md)章节进行操作。
3. TaskD需安装在容器内，详细请参见[制作镜像](../usage/resumable_training/07_using_resumable_training_on_the_cli.md#制作镜像)章节。
4. MindIO ACP的详细介绍及安装步骤请参见[Checkpoint保存与加载优化](../optimizing_saving_and_loading_checkpoints/01_product_description.md)章节。
5. MindIO TFT的详细介绍及安装步骤请参见[故障恢复加速](../fault_recovery_acceleration/01_product_description.md)。

## 容器恢复<a name="ZH-CN_TOPIC_0000002492192948"></a>

**功能特点<a name="section1788818281655"></a>**

在无K8s的场景下，训练或推理进程异常后，通过配置容器恢复功能，可以进行容器故障恢复。

- **故障检测**：通过Container Manager组件，发现任务故障。
- **故障处理**：故障发生后，不需要人工介入就可自动恢复故障设备。
- **容器恢复**：故障发生时，将容器停止，故障恢复后重新将容器拉起。

**所需组件<a name="section15655185785119"></a>**

Container Manager

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[一体机特性指南](../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md)章节进行操作。
