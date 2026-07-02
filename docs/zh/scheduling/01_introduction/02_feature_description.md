# 特性介绍

## 容器化支持<a name="ZH-CN_TOPIC_0000002479386930"></a>

**功能特点<a name="section1788818281655"></a>**

- 为所有的训练或推理作业提供NPU容器化支持，自动挂载所需文件和设备依赖，简化容器拉起命令。
- 支持vNPU的自动创建和挂载。
- 支持Docker及Containerd。

**所需组件<a name="section15655185785119"></a>**

Ascend Docker Runtime

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[容器化支持特性指南](../04_usage/containerization/00_before_you_start.md)章节进行操作。

## 资源监测<a name="ZH-CN_TOPIC_0000002479386910"></a>

**功能特点<a name="section1788818281655"></a>**

- 支持在执行训练或者推理任务时，对昇腾AI处理器资源各种数据信息的实时监测，可实时获取昇腾AI处理器利用率、温度、电压、内存，以及昇腾AI处理器在容器中的分配状况等信息，实现资源的实时监测。
- 支持通过自定义插件上报其他指标。

**所需组件<a name="section15655185785119"></a>**

NPU Exporter

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[资源监测](../04_usage/resource_monitoring/00_before_you_start.md)章节进行操作。

## 虚拟化实例<a name="ZH-CN_TOPIC_0000002511346855"></a>

虚拟化实例可将一张NPU切分成多份，分给不同的任务使用。按照切分的方式，分为基于HDK的虚拟化实例和基于vCANN-RT的虚拟化实例。

### 基于HDK的虚拟化实例<a name="ZH-CN_TOPIC_0000002511346855hdk"></a>

**功能介绍<a name="section1337420477275"></a>**

基于HDK的虚拟化实例功能是指通过资源虚拟化的方式将物理机或虚拟机配置的NPU（昇腾AI处理器）切分成若干份vNPU（虚拟NPU）挂载到容器中使用，虚拟化管理方式能够实现不同规格资源的分配和回收处理，满足多用户反复申请/释放的资源操作请求。

**所需组件<a name="ZH-CN_TOPIC_0000002479226932"></a>**

- 静态虚拟化：通过npu-smi工具**手动**创建多个vNPU。基于固定大小、固定数量的vNPU进行调度
- 动态虚拟化：通过Volcano和Ascend Device Plugin配合，动态地**自动**创建vNPU，容器销毁前，自动销毁vNPU。

静态虚拟化实例所需组件：

- Ascend Docker Runtime
- Ascend Device Plugin

动态虚拟化实例所需组件：

- Ascend Docker Runtime
- Ascend Device Plugin
- Volcano

**使用说明<a name="section1350915844811"></a>**

- 安装请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
- 特性使用指导请参见[基于HDK的虚拟化实例](../04_usage/virtual_instance/virtual_instance_with_hdk/01_description.md)章节进行操作。

### 基于vCANN-RT的虚拟化实例<a name="ZH-CN_TOPIC_0000002511346855vcann"></a>

**功能介绍<a name="section1337420477275vcann"></a>**

基于vCANN-RT的虚拟化实例功能是指通过向vCANN-RT提供软切分配置文件的方式将物理机配置的NPU（昇腾AI处理器）挂载到容器中使用，虚拟化管理方式能够实现不同规格资源的分配和回收处理，满足多用户反复申请/释放资源的操作请求。

**所需组件<a name="ZH-CN_TOPIC_0000002479226932vcann"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD

**使用说明<a name="section1350915844811vcann"></a>**

1. 安装组件请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[基于vCANN-RT的虚拟化实例](../04_usage/virtual_instance/virtual_instance_with_vcann_rt/00_description.md)章节进行操作。

## 基础调度<a name="ZH-CN_TOPIC_0000002511346871"></a>

### 整卡调度<a name="ZH-CN_TOPIC_0000002479386926"></a>

**功能特点<a name="section1788818281655"></a>**

支持用户运行训练或者推理任务时，将训练或推理任务调度到节点的整张NPU卡上，独占整张卡执行训练或者推理任务。整卡调度特性借助Kubernetes（以下简称K8s）支持的基础调度功能，配合Volcano或者其他调度器，根据NPU设备物理拓扑，选择合适的NPU设备，最大化发挥NPU性能，实现训练或者推理任务的NPU卡的调度和其他资源的最佳分配。

芯片间的网络拓扑越复杂，调度逻辑越复杂，详细可以参见[亲和性调度](../04_usage/basic_scheduling/01_affinity_scheduling/00_solution_description.md)。

支持Preempt（抢占）和Reclaim Action（回收）操作。Preempt用于同一个队列中任务之间的资源抢占，当高优先级任务需要资源时，可以抢占低优先级任务的资源；Reclaim用于不同队列之间的资源回收，当某个队列中的任务需要资源且该队列资源未超用时，可以从其他可回收队列中回收资源。两者均可实现资源的动态调整和优化分配。关于Preempt和Reclaim Action的详细说明，请参见[Volcano官方网站相关信息](https://volcano.sh/zh/docs/Scheduler/Actions)。

**所需组件<a name="section15655185785119"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- Infer Operator
- ClusterD
- NodeD

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[整卡调度](../04_usage/basic_scheduling/03_full_npu_scheduling.md)章节进行操作。
3. Preempt和Reclaim Action的使用样例请参见[任务交替运行最佳实践](../04_usage/task_alternation/00_before_you_start.md)章节进行操作。

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

1. 安装组件请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[多级调度](../04_usage/basic_scheduling/05_multi_level_scheduling.md)章节进行操作。

### 推理卡故障恢复<a name="ZH-CN_TOPIC_0000002479226952"></a>

**功能特点<a name="section113779818313"></a>**

集群调度组件管理的推理NPU资源出现故障后，将对故障资源（对应NPU）进行热复位操作，使NPU恢复健康。

**所需组件<a name="section143231032154719"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD

**使用说明<a name="section74221327111220"></a>**

- 安装组件请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
- 特性使用指导请参见[推理卡故障恢复](../04_usage/basic_scheduling/08_recovery_of_inference_card_faults.md)章节进行操作。

### 推理卡故障重调度<a name="ZH-CN_TOPIC_0000002511346875"></a>

**功能特点<a name="section119259203315"></a>**

集群调度组件管理的推理NPU资源出现故障后，集群调度组件将对故障资源（对应NPU）进行隔离并自动进行重调度。

**所需组件<a name="section15655185785119"></a>**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD
- Ascend Operator
- Infer Operator

**使用说明<a name="section18894171918127"></a>**

- 安装组件请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
- 特性使用指导请参见[推理卡故障重调度](../04_usage/basic_scheduling/07_rescheduling_upon_inference_card_faults.md)章节进行操作。

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

1. 安装组件请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[断点续训](../04_usage/resumable_training/00_feature_description.md)章节进行操作。
3. TaskD需安装在容器内，详细请参见[制作镜像](../04_usage/resumable_training/03_using_resumable_training_on_the_cli.md#制作镜像)章节。
4. MindIO ACP的详细介绍及安装步骤请参见[Checkpoint保存与加载优化](../07_references/optimizing_saving_and_loading_checkpoints/01_product_description.md)章节。
5. MindIO TFT的详细介绍及安装步骤请参见[故障恢复加速](../07_references/fault_recovery_acceleration/01_product_description.md)。

## 容器恢复<a name="ZH-CN_TOPIC_0000002492192948"></a>

**功能特点<a name="section1788818281655"></a>**

在无K8s的场景下，训练或推理进程异常后，通过配置容器恢复功能，可以进行容器故障恢复。

- **故障检测**：通过Container Manager组件，发现任务故障。
- **故障处理**：故障发生后，不需要人工介入就可自动恢复故障设备。
- **容器恢复**：故障发生时，将容器停止，故障恢复后重新将容器拉起。

**所需组件<a name="section15655185785119"></a>**

Container Manager

**使用说明<a name="section1245612501584"></a>**

1. 安装组件请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[一体机特性指南](../04_usage/appliance/01_npu_hardware_fault_detection_and_rectification.md)章节进行操作。

## 容器快照

**功能特点**

本特性实现推理服务的容器快照能力，支持大模型推理服务快速启动和故障场景下的快速恢复。通过MindCluster的Infer Operator、NodeD和Ascend Docker Runtime组件协作，在推理任务完成warm up后生成Host和Device侧快照，在异常删除Pod后通过快照快速恢复服务，将推理服务启动时间从30分钟以上缩短至分钟级。

**所需组件**

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD
- Infer Operator

**使用说明**

1. 安装组件请参见[安装部署](../05_developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。
2. 特性使用指导请参见[容器快照部署及使用](../05_developer_guide/container_snapshot_usage.md)章节进行操作。
