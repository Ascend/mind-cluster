# Before You Start<a name="ZH-CN_TOPIC_0000002524312666"></a>

DPU resource monitoring provides real-time monitoring of the running status and statistics of the DPU. The monitored content falls into two categories:

- **Global metrics (DPU level)**: Covers statistics such as RoCE packet errors, packet loss, received ECN, sent/received CNP, and abnormal PSN retransmission.
- **Interface-level metrics**: Covers metrics such as the link running status, send/receive traffic, and abnormal error status of each NIC port.

DPU resource monitoring is a basic feature. It does not distinguish between training and inference scenarios, nor between scenarios that use the Volcano scheduler and scenarios that use other schedulers. This feature must be used together with Prometheus. That is, after Prometheus is deployed, call the DPU Exporter APIs to monitor resources.

- Prometheus is a complete open-source monitoring solution featuring easy management, high efficiency, scalability, and visualization. It works with DPU Exporter to display the reported DPU information in a visualized manner.

## Prerequisites<a name="section1672062465010"></a>

- Before using the DPU resource monitoring feature, ensure that DPU Exporter has been installed. If it is not installed, see [Installation and Deployment](../../../03_installation_guide/02_installation/00_helm_installation.md). If it is deployed in binary mode, see [DPU Exporter Installation and Deployment](../../../05_developer_guide/00_installation_deployment/00_manual_installation/12_dpu_exporter.md).
- Before starting DPU Exporter, ensure that the DPU is in position.
- The DPU driver and the NIC management tool hinicadm5 have been installed on the host machine, and hinicadm5 is located in the "/usr/sbin/" directory.

## Usage Instructions<a name="section45381612353"></a>

- DPU resource monitoring can be used together with all features in training scenarios, and also with all features in inference scenarios.
- DPU Exporter periodically calls the hinicadm5 tool and reads the sysfs file interface to obtain metrics. You can control the collection scope based on the metrics you care about through `metricWhiteList` in the configuration file. For details, see [DPU Exporter Installation and Deployment](../../../05_developer_guide/00_installation_deployment/00_manual_installation/12_dpu_exporter.md).
- For details about the format of the monitored metric data, see [Prometheus Metrics API](../../../06_api/16_dpu_exporter.md).

## Supported Product Forms<a name="section170961844182917"></a>

- Ascend 950PR Products
- Ascend 950DT Products
