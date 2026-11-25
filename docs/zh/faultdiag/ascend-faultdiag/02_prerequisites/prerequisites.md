# 前置知识

使用ascend-fd之前，建议用户了解以下基础知识。

## 必备知识

### Linux系统管理

ascend-fd运行在Linux服务器上，用户需要会基本的Linux操作。

### Python基础

ascend-fd是使用Python开发的工具，安装和使用过程中会用到Python基础知识和包管理工具pip。

## 推荐知识

### 昇腾计算相关概念

- **NPU**：神经网络处理器单元，NPU是专门为神经网络推理设计的硬件芯片，其通过并行计算架构和指令集优化，能够以极低的功耗与极高的效率执行矩阵乘法、卷积等神经网络核心运算。
- **[CANN](https://www.hiascend.com/cann/document)**：CANN（Compute Architecture for Neural Networks）是华为针对AI场景推出的异构计算架构，对上支持多种AI框架，对下服务AI处理器与编程，发挥承上启下的关键作用，是提升昇腾AI处理器计算效率的关键平台。
- **[HCCL](https://gitcode.com/cann/hccl)**：华为集合通信库（Huawei Collective Communication Library，简称HCCL）是基于昇腾AI处理器的高性能集合通信库，为计算集群提供高性能、高可靠的通信方案。
- **[MindIE](https://www.hiascend.com/cn/developer/software/mindie)**：MindIE（Mind Inference Engine，昇腾推理引擎）是华为昇腾针对AI全场景业务的推理加速套件。通过分层开放AI能力，支撑用户多样化的AI业务需求，使能百模千态，释放昇腾硬件设备算力。
- **[AMCT](https://gitcode.com/cann/amct)**：Ascend Model Compression Toolkit，简称 AMCT，是一款昇腾AI处理器亲和的深度学习模型压缩工具包，提供多种模型量化压缩特性。压缩后模型体积变小，部署到昇腾AI处理器可使能低比特运算，提高计算效率，达到性能提升的目标。
- **LCNE**：灵衢计算网络引擎，支持大规模计算网络的分域集中式拓扑发现、路由管理、转发控制。
- **BMC**：基板管理控制器，用于监控和管理服务器硬件的控制器，如温度、电压、风扇等。
- **[vLLM](https://docs.vllm.ai/en/latest/)**：是用于大语言模型（LLM）的推理和服务部署的Python库。
- **[MindCluster](https://www.hiascend.com/cn/developer/software/mindcluster)**：是支持NPU（昇腾AI处理器）训练和推理硬件的深度学习组件，使能构建集群全流程运行，提供NPU集群作业调度、运维监测、故障恢复等功能。
- **[MindSpore](https://www.mindspore.cn/)**：昇思MindSpore是一个全场景深度学习框架。
- **[MindIE-PyMotor](https://gitcode.com/Ascend/MindIE-PyMotor)**：昇腾自研推理集群管理框架，提供一键式PD分离与PD混部部署，基于云原生插件化架构灵活适配多种推理引擎（vLLM、SGLang），结合高性能调度与负载均衡能力，构建高可用、可扩展的大规模推理服务。
- **[ModelArts](https://www.huaweicloud.com/product/modelarts.html)**：面向开发者的一站式AI开发平台，为机器学习与深度学习提供海量数据预处理及半自动化标注、大规模分布式训练、自动化模型生成，及端-边-云模型按需部署能力，帮助用户快速创建和部署模型，管理全周期AI工作流。
- **[SmartKit](https://support.huawei.com/enterprise/zh/flash-storage/smartkit-pid-8576706)**：为存储、服务器、云计算三大领域的产品提供了统一的服务工具平台。
- **[CCAE](https://www.hiascend.com/software/ccae)**：集群自智引擎（Cluster Computing Autonomous Engine），是一款面向集群计算业务场景的智能运维平台，主要提供计算、网络、存储的基础（资源、性能、告警）管理能力，以及集群健康评估、集群故障诊断、作业质量保障、集群数字地图、灵衢总线管理的高阶能力。
