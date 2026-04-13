# 使用前必读<a name="ZH-CN_TOPIC_0000002512753445"></a>

MindCluster集群调度组件支持用户通过OME（Open Model Engine）部署SGLang推理任务进行调度和故障实例重调度。

本章节说明相关特性原理及对应配置示例，用户可以参考配置示例部署基于OME的SGLang推理任务。

**前提条件<a name="zh-cn_topic_0000002322062116_section52051339787"></a>**

在部署SGLang推理服务前，需要确保相关组件已经安装，若没有安装，可以参考[安装部署](../../installation_guide/03_installation.md)章节进行操作。

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD（可选）

**支持的产品形态<a name="zh-cn_topic_0000002322062116_section169961844182917"></a>**

- Atlas 800I A2 推理服务器
- Atlas 800I A3 超节点服务器

**使用方式<a name="zh-cn_topic_0000002322062116_section6771194616104"></a>**

MindCluster集群调度组件支持用户通过以下方式进行SGLang推理服务的容器化部署、故障重调度。本章节仅介绍通过命令行使用和通过脚本一键式部署使用方式。

- 通过命令行使用：通过配置的YAML文件部署任务。
- 通过脚本一键式部署使用：通过自动化脚本参考设计部署任务。
- 集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。
