# Before You Start

MindCluster allows you to deploy inference jobs through Infer Operator for scheduling and rescheduling of faulty instances.

This chapter only describes the principles of related features and corresponding configuration examples. You can refer to the configuration examples to deploy Infer Operator inference jobs.

## Prerequisites

Before deploying an Infer Operator inference job, ensure that the related components have been installed. If they are not installed, refer to the [Installation and Deployment](../../03_installation_guide/02_installation/00_helm_installation.md) chapter for instructions.

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Infer Operator
- ClusterD
- NodeD (Optional)

## Supported Product Forms

- Atlas 800I A2
- Atlas 800I A3
- Atlas 350 accelerator card
- Atlas 850E
- Atlas 850
- Atlas 650E
- Atlas 650
- Atlas 950 SuperPoD

## Usage Methods

MindCluster supports deploying Infer Operator inference jobs in the following ways.

- [Deploying an Infer Operator Inference Job Based on vLLM Proxy](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md). The following two deployment methods are supported:
  - [Using the Command Line](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#usage-via-command-line): Deploy the job using a configured YAML file.
  - [One-Click Deployment Using the MindCluster Deployment Tool](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#one-click-deployment-and-usage-via-mindcluster-deployment-tool): Deploy the job using an automated script based on a reference design.
- [Deploying an Infer Operator Inference Job Based on MindIE PyMotor](./02_deploying_infer_operator_inference_job_with_mindie_pymotor.md).

>[!NOTE]
>If you have not configured a RoCE network or the RoCE network is unhealthy:
>
>- In a non-SuperPoD scheduling scenario, a single-node inference instance can be scheduled normally, but KV transfers between inference instances may be abnormal, causing the inference job to fail to run normally.
>- In a SuperPoD scheduling scenario, if the number of logical SuperPoDs of the inference instance is 1, the inference instance can be scheduled normally, but KV transfers between inference instances may be abnormal, causing the inference job to fail to run normally.
>
> Infer Operator supports deploying the following roles within the same InferServiceSet:
>
> - Inference roles that request NPU resources (such as Prefill and Decode)
> - General-computing roles that do not request NPU resources (such as Router, MetaService, etc.)
>
> The deployment, priority, and rescheduling configurations of general-computing roles are the same as those of inference roles and reuse the existing YAML parameters. The differences are:
>
> - General-computing roles do not request NPU resources, and the `huawei.com/` field does not need to be configured in `resources.requests`. The NPU scheduling plugin of ascend-for-volcano returns early in the filtering, scoring, and reservation phases and does not participate in NPU allocation. Pods are scheduled to nodes (including nodes without NPUs) by the default Volcano plugins.
> - Fault rescheduling of general-computing roles is triggered only by node failures and service-plane failures. NPU failures do not trigger fault rescheduling of general-computing roles.
