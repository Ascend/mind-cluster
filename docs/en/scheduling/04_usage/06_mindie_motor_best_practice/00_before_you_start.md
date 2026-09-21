# Before You Start<a name="ZH-CN_TOPIC_0000002511346371"></a>

MindCluster support containerized deployment, fault rescheduling, and elastic scaling of MindIE Motor by generating inference jobs of acjob type.

This section only describes the principles of related features and provides corresponding configuration examples. The provided YAML examples are not sufficient to complete the deployment of MindIE jobs. For details about the complete deployment process of MindIE Motor, see the [MindIE Motor Development Guide](https://www.hiascend.com/document/detail/en/mindie/310/mindiellm/llmdev/mindie_motor_cpp/user_guide/introduction.md).

## Prerequisites<a name="zh-cn_topic_0000002322062116_section52051339787"></a>

Before deploying MindIE Motor, ensure that the relevant components have been installed. If they are not installed, refer to the [Installation and Deployment](../../03_installation_guide/02_installation/00_helm_installation.md) section for instructions.

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- Ascend Operator; the startup parameter [enableGangScheduling](../../05_developer_guide/00_installation_deployment/00_manual_installation/08_ascend_operator.md) must be set to `true`
- ClusterD
- NodeD

## Supported Product Forms<a name="zh-cn_topic_0000002322062116_section169961844182917"></a>

- Atlas 800I A2 inference server
- Atlas 800I A3 SuperPoD server

## Usage Methods<a name="zh-cn_topic_0000002322062116_section6771194616104"></a>

MindCluster supports containerized deployment, fault rescheduling, and elastic scaling of MindIE Motor through the following two methods. This section only describes the method of using the command line.

- [Using the command line](./01_deploying_mindie_motor.md#using-via-command-line): Deploy jobs through configured YAML files.
- Using after integration: Integrate the cluster scheduling components into an existing third-party AI platform or an AI platform developed based on these components.
