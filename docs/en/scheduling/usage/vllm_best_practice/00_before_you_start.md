# Before You Start<a name="ZH-CN_TOPIC_0000002516292409"></a>

MindCluster supports deploying vLLM inference jobs for scheduling and fault instance rescheduling through the [StormService](https://aibrix.readthedocs.io/latest/designs/aibrix-stormservice.html) workload defined by the [AIBrix](https://github.com/vllm-project/aibrix) framework. The currently adapted AIBrix version is [v0.5.0](https://github.com/vllm-project/aibrix/tree/v0.5.0); the adapted [vLLM-Ascend](https://github.com/vllm-project/vllm-ascend) version is the main branch with commit ID [41fbc5e](https://github.com/vllm-project/vllm-ascend/commit/41fbc5ebc9b35bb81f3f14dbe55a76539f6675f5) and later versions.

This section describes the principles of related features and corresponding configuration examples. You can refer to the configuration examples to deploy AIBrix-based vLLM inference jobs.

## Prerequisites<a name="zh-cn_topic_0000002322062116_section52051339787"></a>

Before deploying vLLM inference jobs, ensure that the relevant components have been installed. If they are not installed, refer to the [Installation and Deployment](../../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) section for instructions.

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD (Optional)

## Supported Product Forms<a name="zh-cn_topic_0000002322062116_section169961844182917"></a>

- Atlas 800I A2 inference server
- Atlas 800I A3 SuperPoD server

## Usage Methods<a name="zh-cn_topic_0000002322062116_section6771194616104"></a>

MindCluster supports containerized deployment and fault rescheduling of vLLM inference services through the following methods. This section only introduces the methods of using the command line and one-click deployment via scripts.

- [Using via Command Line](./01_deploying_vllm_inference_job.md#using-via-command-line): Deploy jobs through configured YAML files.
- [One-Click Deployment via Script](./01_deploying_vllm_inference_job.md#deploying-inference-jobs-using-a-script-in-one-click-mode): Deploy jobs using automated script reference designs.
- Using after integration: Integrate cluster scheduling components into an existing third-party AI platform or an AI platform developed based on these components.
