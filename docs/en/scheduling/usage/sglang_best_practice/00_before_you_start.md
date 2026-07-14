# Before You Start<a name="ZH-CN_TOPIC_0000002512753445"></a>

MindCluster supports deploying SGLang inference jobs through OME (Open Model Engine) for scheduling and fault instance rescheduling.

This chapter describes the relevant feature principles and corresponding configuration examples. You can refer to the configuration examples to deploy OME-based SGLang inference jobs.

## Prerequisites<a name="zh-cn_topic_0000002322062116_section52051339787"></a>

Before deploying the SGLang inference service, ensure that the relevant components have been installed. If they are not installed, refer to the [Installation and Deployment](../../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) chapter for instructions.

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD (Optional)

## Supported Product Forms<a name="zh-cn_topic_0000002322062116_section169961844182917"></a>

- Atlas 800I A2 inference server
- Atlas 800I A3 SuperPoD server

## Usage Methods<a name="zh-cn_topic_0000002322062116_section6771194616104"></a>

MindCluster supports containerized deployment and fault rescheduling of SGLang inference service through the following methods. This section only introduces the command-line usage and the one-click script deployment method.

- [Using via Command Line](./01_deploying_ome_sglang_inference_job.md#using-via-command-line): Deploy jobs through configured YAML files.
- [Using via One-Click Script Deployment](./01_deploying_ome_sglang_inference_job.md#deploying-inference-jobs-using-a-script-in-one-click-mode): Deploy jobs through automated script reference designs.
- Using after integration: Integrate cluster scheduling components into an existing third-party AI platform or an AI platform developed based on these components.
