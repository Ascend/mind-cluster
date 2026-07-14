# Before You Start<a name="ZH-CN_TOPIC_0000002479387018"></a>

Resource monitoring mainly includes real-time monitoring in two aspects: monitoring the AICore utilization, total memory, and used memory of virtual NPUs (vNPUs); and real-time monitoring of various NPU resource data during training or inference jobs, that is, obtaining information such as Ascend AI processor utilization, temperature, voltage, memory, and the allocation status of Ascend AI processors in containers in real time.

The resource monitoring feature is a basic feature that does not distinguish between training or inference scenarios, nor does it distinguish between scenarios using Volcano or other schedulers. This feature needs to be used together with either Prometheus or Telegraf. If used with Prometheus, resource monitoring is implemented by calling NPU Exporter-related interfaces after Prometheus is deployed. If used with Telegraf, Telegraf needs to be deployed and run to implement resource monitoring.

- Prometheus is an open-source monitoring solution featuring easy management, high efficiency, scalability, and visualization. Used with NPU Exporter, it enables real-time monitoring of information such as Ascend AI processor utilization, temperature, voltage, memory, and the allocation status of Ascend AI processors in containers. It supports monitoring the AICore utilization, total memory, and used memory of vNPUs.
- Telegraf is used to collect statistical data of systems and services, featuring low memory usage and support for extension with other services. Used with NPU Exporter, it allows you to view reported information about Ascend AI processors through command outputs in the environment.

## Prerequisites<a name="section1632062465010"></a>

- Before using the resource monitoring feature, ensure that NPU Exporter has been installed. If it is not installed, see [Installation and Deployment](../../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md) for instructions.
- Before starting NPU Exporter, ensure that NPUs are available.

## Usage Instructions<a name="section44381612353"></a>

Resource monitoring can be used together with all features in training scenarios and all features in inference scenarios.

## Supported Product Forms<a name="section169961844182917"></a>

The following products support resource monitoring.

- Atlas training series products
- <term>Atlas A2 training series products</term>
- <term>Atlas A3 training series products</term>
- Inference server (with Atlas 300I inference card)
- Atlas inference series products
- Atlas 800I A2 inference server
- A200I A2 Box heterogeneous subrack
- Atlas 800I A3 SuperPoD server
- Atlas 350 PCIe card
- Atlas 950 SuperPoD
