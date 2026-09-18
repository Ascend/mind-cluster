# Prerequisites

<!-- md-trans-meta sourceCommit=14372229050641e3df56cf2ec68c3468e5c3727f translatedAt=2026-08-24T02:26:31.612Z pushedAt=2026-08-24T02:51:43.228Z -->

Before using ascend-fd, you are advised to understand the following basic knowledge.

## Basic Knowledge

### Linux System Administration

ascend-fd runs on Linux servers, so you need to be familiar with basic Linux operations.

### Python Basics

ascend-fd is a tool developed in Python. Basic Python knowledge and the package manager pip are required during installation and usage.

## Recommended Knowledge

### Ascend Computing Related Concepts

- **NPU**: Neural-network Processing Unit. An NPU is a hardware chip specifically designed for neural network inference. Through parallel computing architecture and instruction set optimization, it can execute core neural network operations such as matrix multiplication and convolution with extremely low power consumption and extremely high efficiency.
- **[CANN](https://www.hiascend.com/eng/cann)**: Compute Architecture for Neural Networks (CANN) is a heterogeneous computing architecture launched by Huawei for AI scenarios. It supports multiple AI frameworks on the upper layer and serves AI processors and programming on the lower layer, playing a key bridging role and serving as a critical platform for improving the computing efficiency of Ascend AI processors.
- **[HCCL](https://gitcode.com/cann/hccl)**: The Huawei Collective Communication Library (HCCL) is a high-performance collective communication library based on Ascend AI processors, providing high-performance and highly reliable communication solutions for computing clusters.
- **[MindIE](https://www.hiascend.com/en/developer/software/mindie)**: Mind Inference Engine (MindIE) is an inference acceleration suite launched by Huawei Ascend for all AI scenarios. By opening up AI capabilities in a layered manner, it supports diverse AI business requirements, enables a wide range of models and scenarios, and unleashes the computing power of Ascend hardware devices.
- **[AMCT](https://gitcode.com/cann/amct)**: Ascend Model Compression Toolkit, abbreviated as AMCT, is a deep learning model compression toolkit optimized for Ascend AI processors, providing multiple model quantization and compression features. After compression, the model size is reduced. Deploying it on Ascend AI processors enables low-bit operations, improving computational efficiency and achieving the goal of performance improvement.
- **LCNE**: UnifiedBus Computing Network Engine, which supports domain-based centralized topology discovery, routing management, and forwarding control for large-scale compute networks.
- **BMC**: Baseboard Management Controller, a controller used to monitor and manage server hardware, such as temperature, voltage, and fans.
- **[vLLM](https://docs.vllm.ai/en/latest/)**: A Python library for inference and serving deployment of large language models (LLMs).
- **[MindCluster](https://www.hiascend.com/eng/developer/software/mindcluster)**: A deep learning system that supports NPU (Ascend AI processor) training and inference hardware, enabling end-to-end cluster operation and providing functions such as NPU cluster job scheduling, O&M monitoring, and fault recovery.
- **[MindSpore](https://www.mindspore.cn/en)**: A full-scenario deep learning framework.
- **[MindIE-PyMotor](https://gitcode.com/Ascend/MindIE-PyMotor)**: An inference cluster management framework independently developed by Ascend. It provides one-click PD disaggregation and PD hybrid deployment, flexibly adapts to multiple inference engines (vLLM and SGLang) based on a cloud-native pluggable architecture, and combines high-performance scheduling and load balancing capabilities to build highly available and scalable large-scale inference services.
- **[ModelArts](https://www.huaweicloud.com/intl/en-us/product/modelarts.html)**: A one-stop AI development platform for developers. It provides massive data preprocessing and semi-automated labeling, large-scale distributed training, automated model generation, and on-demand edge-cloud model deployment for machine learning and deep learning, helping users quickly create and deploy models and manage the full lifecycle of AI workflows.
- **[SmartKit](https://support.huawei.com/enterprise/en/flash-storage/smartkit-pid-8576706)**: A unified service tool platform for products in the three major fields of storage, servers, and cloud computing.
- **[CCAE](https://www.hiascend.com/en/software/ccae)**: Cluster Computing Autonomous Engine (CCAE) is an intelligent O&M platform for cluster computing service scenarios. It mainly provides basic management capabilities (resources, performance, and alarms) for computing, networking, and storage, as well as advanced capabilities such as cluster health assessment, cluster fault diagnosis, job quality assurance, cluster digital map, and UnifiedBus management.
