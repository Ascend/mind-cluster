# Before You Start<a name="ZH-CN_TOPIC_0000002511427169"></a>

Containerization is a technical capability that packages an application and its dependencies into an independent, portable environment (container).

## Prerequisites<a name="section1632062465010"></a>

Before using the containerization feature, ensure that the relevant components are installed. If they are not installed, refer to the [Installation and Deployment](../../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md) section for instructions.

- Ascend Docker Runtime

## Usage Instructions<a name="section44381612353"></a>

- Containerization can be used together with all features in training scenarios, and can also be used together with all features in inference scenarios.
- If Volcano is used for job scheduling, it is not recommended to create/mount containers with NPUs through Docker or Containerd commands, as this may trigger Volcano scheduling issues.

## Supported Product Forms<a name="section169961844182917"></a>

The following products support containerization.

- Atlas training products
- <term>Atlas A2 training products</term>
- <term>Atlas A3 training products</term>
- Inference server (with Atlas 300I inference card)
- <term>Atlas 200/300/500 inference products</term>
- <term>Atlas 200I/500 A2 inference products</term>
- <term>Atlas inference products</term>
- Atlas 800I A2 inference server
- A200I A2 Box heterogeneous subrack
- Atlas 800I A3 SuperPoD server
- Atlas 350 accelerator card
- Atlas 850E SuperPoD
- Atlas 650E server
- Atlas 950 SuperPoD

## Use Cases<a name="section124697813416"></a>

Ascend Docker Runtime supports containerization in the following 2 scenarios.

- [Using with Docker Client](./02_usage_on_the_docker_client.md)
- [Using with Containerd Client](./03_usage_on_the_containerd_client.md)
