# 使用前必读

MindCluster集群调度组件支持用户通过Infer Operator部署推理任务进行调度和故障实例重调度。

本章节仅说明相关特性原理及对应配置示例。用户可以参考配置示例部署Infer Operator推理任务。

**前提条件**

在部署Infer Operator推理任务前，需要确保相关组件已经安装，若没有安装，可以参考[安装部署](../../installation_guide/03_installation.md)章节进行操作。

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Infer Operator
- ClusterD
- NodeD（可选）

**支持的产品形态**

- Atlas 800I A2 推理服务器
- Atlas 800I A3 超节点服务器

**使用方式**

MindCluster集群调度组件支持通过以下方式部署Infer Operator推理任务。

- [基于vLLM Proxy部署Infer Operator推理任务](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md)。支持通过以下两种方式部署：
  - [通过命令行使用](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#通过命令行使用)：通过配置的YAML文件部署任务。
  - [通过MindCluster社区部署工具一键部署使用](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#通过mindcluster社区部署工具一键部署使用)：通过自动化脚本参考设计部署任务。
- [基于MindIE PyMotor部署Infer Operator推理任务](./02_deploying_infer_operator_inference_job_with_mindie_pymotor.md)。
