# Before You Start<a name="ZH-CN_TOPIC_0000002516292419"></a>

MindCluster cluster scheduling components allow users to deploy and schedule verl reinforcement learning (RL) jobs through the acjob workload, and support job rescheduling and faulty Pod rescheduling after a fault. Faulty Pod rescheduling must be used together with the `fault_recover` function in `verl-ascend-recipe` (see [verl-ascend-recipe/fault_recover](https://github.com/verl-project/verl-ascend-recipe/tree/main/fault_recover)).

This chapter describes the principles of related features and the corresponding configuration examples. Users can refer to the configuration examples to deploy verl reinforcement learning jobs.

## Prerequisites<a name="zh-cn_topic_0000002322062116_section52051339787"></a>

Before deploying a verl reinforcement learning job, ensure that the related components have been installed. If they are not installed, see [Installation and Deployment](../../../03_installation_guide/02_installation/00_helm_installation.md).

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator
- ClusterD
- NodeD (optional)

## Supported Product Forms<a name="zh-cn_topic_0000002322062116_section169961844182917"></a>

- Atlas 800T A2
- Atlas 900 A3 SuperPoD

## Usage<a name="zh-cn_topic_0000002322062116_section6771194616104"></a>

MindCluster cluster scheduling components allow users to deploy containerized verl reinforcement learning jobs and perform fault rescheduling in the following ways. This chapter describes only the command line method.

- [Using via the Command Line](./01_deploying_verl_reinforcement_learning_job.md#using-via-command-line): Deploy a job through a configured YAML file.
- Integration: Integrate the cluster scheduling components into an existing third-party AI platform or an AI platform developed based on the cluster scheduling components.
