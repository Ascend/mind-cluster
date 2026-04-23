# 产品描述

## 产品介绍

MindCluster MindIO Async Checkpoint Persistence（下文简称MindIO ACP）加速大模型Checkpoint功能主要针对大模型训练中的Checkpoint的保存及加载进行加速，Checkpoint的数据先写入训练服务器的内存系统中，再异步写入后端的可靠性存储设备中。本文档主要介绍纵向加速部分，包含Checkpoint在本系统中的写入及读取过程。

## 产品价值

LLM（Large Language Model，大语言模型）是全球当前科技界竞争的焦点，LLM模型的训练往往需要长达数十天、甚至数月。Checkpoint是模型中断训练后恢复的关键点，Checkpoint的密集程度、保存和恢复的性能较为关键，它可以提高训练系统的有效吞吐率。MindIO ACP针对Checkpoint的加速方案，支持昇腾产品在LLM模型领域扩展市场空间。

该方案提升昇腾平台上LLM模型的训练吞吐量，性能超越[Microsoft Azure Nebula方案](https://learn.microsoft.com/zh-cn/azure/machine-learning/reference-checkpoint-performance-for-large-models?view=azureml-api-2&tabs=PYTORCH)。

## MindIO ACP架构

![](../../figures/scheduling/mindio_acp架构.png)

MindIO ACP加速LLM Checkpoint保存和加载的4个关键点如下：

- 异步持久化。训练框架通过MindIO ACP的save/load接口或MindSpore框架将Checkpoint保存到MindIO ACP后，直接返回继续训练，该时间为秒级；MindIO ACP会异步将Checkpoint写入持久化的分布式存储，该过程为分钟级。
- 高性能MemFS（Memory File System，内存文件系统）。MindIO ACP为实现Checkpoint极速写入，实现了全用户态的以内存为介质的文件系统；消除各种标准文件系统的系统调用和用户态到内核态的内存拷贝。
- 高效Checkpoint保存和加载。MindIO ACP为实现Checkpoint极速写入和恢复，研发了高效Checkpoint保存、加载方式。
- MindIO ACP具备自动容错能力。当MindIO ACP服务异常导致数据读写失败、超时等异常时，能自动切换到原生数据存储方式，保证业务不中断。

    > [!CAUTION]注意
    > MindIO ACP仅保存训练过程中的Checkpoint数据，暂不支持敏感数据的保存和处理。若涉及敏感数据存储，请在前序流程完成相关脱敏操作，避免造成信息安全问题。
