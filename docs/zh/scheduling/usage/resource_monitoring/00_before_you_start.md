# 使用前必读<a name="ZH-CN_TOPIC_0000002479387018"></a>

资源监测主要包含2个方面的实时监测：对虚拟NPU（vNPU）的AI Core利用率、vNPU总内存和vNPU使用中内存进行监测；对训练或者推理任务中NPU资源各种数据信息的实时监测，即实时获取昇腾AI处理器利用率、温度、电压、内存，以及昇腾AI处理器在容器中的分配状况等信息。

资源监测特性是一个基础特性，不区分训练或者推理场景；同时也不区分使用Volcano调度器或者使用其他调度器场景。资源监测特性需要用户配合Prometheus或Telegraf中的一种使用，如果配合Prometheus使用，则需要在部署Prometheus后通过调用NPU Exporter相关接口，实现资源监测，如果配合Telegraf使用，则需要部署和运行Telegraf，实现资源监测。

- Prometheus是一个开源的完整监测解决方案，具有易管理、高效、可扩展、可视化等特点，搭配NPU Exporter组件使用，可实现对昇腾AI处理器利用率、温度、电压、内存，以及昇腾AI处理器在容器中的分配状况等信息的实时监测。支持对虚拟NPU（vNPU）的AI Core利用率、vNPU总内存和vNPU使用中内存进行监测。
- Telegraf用于收集系统和服务的统计数据，具有内存占用小和支持其他服务的扩展等功能。搭配NPU Exporter组件使用，可以在环境上通过回显查看上报的昇腾AI处理器的相关信息。

**前提条件<a name="section1632062465010"></a>**

- 在使用资源监测特性前，需要确保NPU Exporter组件已经安装，若没有安装，可以参考[安装部署](../../installation_guide/03_installation.md)章节进行操作。
- NPU Exporter启动前，请确保NPU卡在位。

**使用说明<a name="section44381612353"></a>**

资源监测可以和训练场景下的所有特性一起使用，也可以和推理场景的所有特性一起使用。

**支持的产品形态<a name="section169961844182917"></a>**

支持以下产品使用资源监测。

- Atlas 训练系列产品
- Atlas A2 训练系列产品
- Atlas A3 训练系列产品
- 推理服务器（插Atlas 300I 推理卡）
- Atlas 推理系列产品
- Atlas 800I A2 推理服务器
- A200I A2 Box 异构组件
- Atlas 800I A3 超节点服务器
- Atlas 350 加速卡
- Atlas 950 SuperPoD
