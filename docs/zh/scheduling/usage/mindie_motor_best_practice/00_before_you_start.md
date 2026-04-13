# 使用前必读<a name="ZH-CN_TOPIC_0000002511346371"></a>

MindCluster集群调度组件支持用户通过生成acjob推理任务的方式进行MindIE Motor的容器化部署、故障重调度和弹性扩缩容。

本章节仅说明相关特性原理及对应配置示例，所提供的YAML示例不足以完成MindIE任务的部署。了解MindIE Motor的详细部署流程请参见《[MindIE Motor开发指南](https://www.hiascend.com/document/detail/zh/mindie/230/mindiemotor/motordev/mindie_service0001.html)》。

**前提条件<a name="zh-cn_topic_0000002322062116_section52051339787"></a>**

在部署MindIE Motor前，需要确保相关组件已经安装，若没有安装，可以参考[安装部署](../../installation_guide/03_installation.md)章节进行操作。

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator，且需将启动参数[enableGangScheduling](../../installation_guide/03_installation.md#ascend-operator)的取值设置为true
- ClusterD
- NodeD

**支持的产品形态<a name="zh-cn_topic_0000002322062116_section169961844182917"></a>**

- Atlas 800I A2 推理服务器
- Atlas 800I A3 超节点服务器

**使用方式<a name="zh-cn_topic_0000002322062116_section6771194616104"></a>**

MindCluster集群调度组件支持用户通过以下2种方式进行MindIE Motor的容器化部署、故障重调度和弹性扩缩容。本章节仅介绍通过命令行使用这种方式。

- 通过命令行使用：通过配置的YAML文件部署任务。
- 集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。
