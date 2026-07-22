# 使用前必读<a name="ZH-CN_TOPIC_0000002518340700"></a>

在无K8s的场景下，可以通过Container Manager组件实现NPU硬件故障的检测、处理以及容器的自动恢复。

## 前提条件<a name="section1632062465010"></a>

在使用本特性前，需要确保如下组件已经安装，若没有安装，可以参考[安装部署](../../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md)章节进行操作。

- Container Manager
- Ascend Docker Runtime（可选）

## 使用说明<a name="section44381612353"></a>

- 本特性适用于无K8s的场景，不依赖K8s调度器。
- 本特性不适用于算力虚拟化场景，不支持共享设备特性及混插模式。
- 特权容器需通过设备配置或ASCEND_VISIBLE_DEVICES环境变量显式挂载NPU才会被管理。

## 支持的产品形态<a name="section169961844182917"></a>

支持以下产品使用故障管理和故障容器的自动恢复功能：
    - <term>Atlas 训练系列产品</term>
    - <term>Atlas A2 训练系列产品</term>
    - <term>Atlas A3 训练系列产品</term>
    - <term>Atlas 推理系列产品</term>
    - <term>Atlas A2 推理系列产品</term>
    - <term>Atlas A3 推理系列产品</term>
