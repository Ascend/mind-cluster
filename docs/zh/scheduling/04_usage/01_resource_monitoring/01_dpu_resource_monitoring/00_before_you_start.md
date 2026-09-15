# 使用前必读<a name="ZH-CN_TOPIC_0000002524312666"></a>

DPU资源监测提供对DPU运行状态与统计指标的实时监测，监测内容分为两类：

- **全局指标（DPU卡级）**：涵盖RoCE错包、丢包、接收ECN、发送/接收CNP及PSN异常重传等统计指标。
- **Interface级指标**：涵盖每个网卡端口的链路运行状态、收发流量与异常错误状态等指标。

DPU资源监测特性是一个基础特性，不区分训练或者推理场景；同时也不区分使用Volcano调度器或者使用其他调度器场景。该特性需要用户配合Prometheus使用，即在部署Prometheus后通过调用DPU Exporter相关接口，实现资源监测。

- Prometheus是一个开源的完整监测解决方案，具有易管理、高效、可扩展、可视化等特点。Prometheus配合DPU Exporter组件使用，以可视化的方式展示上报的DPU相关信息。

## 前提条件<a name="section1672062465010"></a>

- 在使用DPU资源监测特性前，需要确保DPU Exporter组件已经安装，若没有安装，可以参考[安装部署](../../../03_installation_guide/02_installation/00_helm_installation.md)章节进行操作；若以二进制方式部署，可以参考[DPU Exporter安装部署](../../../05_developer_guide/00_installation_deployment/00_manual_installation/13_dpu_exporter.md)章节进行操作。
- DPU Exporter启动前，请确保DPU卡在位。
- 宿主机上已安装DPU驱动与网卡管理工具hinicadm5，且hinicadm5位于"/usr/sbin/"目录下。

## 使用说明<a name="section45381612353"></a>

- DPU资源监测可以和训练场景下的所有特性一起使用，也可以和推理场景的所有特性一起使用。
- DPU Exporter组件周期性调用hinicadm5工具并读取sysfs文件接口获取指标，用户可以根据自身关注的指标，通过配置文件中的metricWhiteList指标白名单控制采集范围，参考[DPU Exporter安装部署](../../../05_developer_guide/00_installation_deployment/00_manual_installation/13_dpu_exporter.md)章节。
- 监测指标数据格式的相关说明，请参见[Prometheus Metrics接口](../../../06_api/16_dpu_exporter.md)章节。

## 支持的产品形态<a name="section170961844182917"></a>

- 昇腾950PR系列产品
- 昇腾950DT系列产品
