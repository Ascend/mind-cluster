# 使用前必读<a name="ZH-CN_TOPIC_0000002516292419"></a>

MindCluster集群调度组件支持用户通过acjob工作负载部署verl强化学习任务进行调度，并支持故障后Job重调度以及故障Pod重调度。其中故障Pod重调度需要搭配verl-ascend-recipe中的fault_recover功能使用（详细请参见[verl-ascend-recipe/fault_recover](https://github.com/verl-project/verl-ascend-recipe/tree/main/fault_recover)）。

本章节说明相关特性原理及对应配置示例，用户可以参考配置示例部署基于verl的强化学习任务。

## 前提条件<a name="zh-cn_topic_0000002322062116_section52051339787"></a>

在部署verl强化学习任务前，需要确保相关组件已经安装，若没有安装，可以参考[安装部署](../../03_installation_guide/02_installation/00_helm_installation.md)章节进行操作。

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD
- NodeD（可选）

## 支持的产品形态<a name="zh-cn_topic_0000002322062116_section169961844182917"></a>

- Atlas 800T A2 训练服务器
- Atlas 900 A3 SuperPoD 超节点

## 使用方式<a name="zh-cn_topic_0000002322062116_section6771194616104"></a>

MindCluster集群调度组件支持用户通过以下方式进行verl强化学习任务的容器化部署、故障重调度。本章节仅介绍通过命令行使用。

- [通过命令行使用](./01_deploying_verl_reinforcement_learning_job.md#通过命令行使用)：通过配置的YAML文件部署任务。
- 集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。
