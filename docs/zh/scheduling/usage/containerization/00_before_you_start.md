# 使用前必读<a name="ZH-CN_TOPIC_0000002511427169"></a>

容器化支持是一种将应用程序及其依赖项打包到一个独立、可移植的环境（容器）中的技术支持。

## 前提条件<a name="section1632062465010"></a>

在使用容器化支持特性前，需要确保相关组件已经安装，若没有安装，可以参考[安装部署](../../developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)章节进行操作。

- Ascend Docker Runtime

## 使用说明<a name="section44381612353"></a>

- 容器化支持可以和训练场景下的所有特性一起使用，也可以和推理场景的所有特性一起使用。
- 若使用Volcano进行任务调度，则不建议通过Docker或Containerd指令创建/挂载NPU卡的容器，否则可能会触发Volcano调度问题。

## 支持的产品形态<a name="section169961844182917"></a>

支持以下产品使用容器化支持。

- Atlas 训练系列产品
- <term>Atlas A2 训练系列产品</term>
- <term>Atlas A3 训练系列产品</term>
- 推理服务器（插Atlas 300I 推理卡）
- <term>Atlas 200/300/500 推理产品</term>
- <term>Atlas 200I/500 A2 推理产品</term>
- Atlas 推理系列产品
- Atlas 800I A2 推理服务器
- A200I A2 Box 异构组件
- Atlas 800I A3 超节点服务器
- Atlas 350 标卡
- Atlas 850 系列硬件产品
- Atlas 950 SuperPoD

## 使用场景<a name="section124697813416"></a>

Ascend Docker Runtime组件支持在以下2种场景下使用容器化支持功能。

- [在Docker客户端使用](./02_usage_on_the_docker_client.md)
- [在Containerd客户端使用](./03_usage_on_the_containerd_client.md)
