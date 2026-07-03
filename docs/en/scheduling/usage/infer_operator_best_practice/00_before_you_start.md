# Before You Start

MindCluster allows you to deploy inference jobs through Infer Operator for scheduling and rescheduling of faulty instances.

This chapter only describes the principles of related features and corresponding configuration examples. You can refer to the configuration examples to deploy Infer Operator inference jobs.

## Prerequisites

Before deploying an Infer Operator inference job, ensure that the related components have been installed. If they are not installed, refer to the [Installation and Deployment](../../developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md) chapter for instructions.

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Infer Operator
- ClusterD
- NodeD (Optional)

## Supported Product Forms

- Atlas 800I A2 inference server
- Atlas 800I A3 SuperPoD server

## Usage Methods

MindCluster supports deploying Infer Operator inference jobs in the following ways.

- [Deploying an Infer Operator Inference Job Based on vLLM Proxy](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md). The following two deployment methods are supported:
  - [Using the Command Line](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#usage-via-command-line): Deploy the job using a configured YAML file.
  - [One-Click Deployment Using the MindCluster Deployment Tool](./01_deploying_infer_operator_inference_job_with_vllm_proxy.md#one-click-deployment-and-usage-via-mindcluster-deployment-tool): Deploy the job using an automated script based on a reference design.
- [Deploying an Infer Operator Inference Job Based on MindIE PyMotor](./02_deploying_infer_operator_inference_job_with_mindie_pymotor.md).
