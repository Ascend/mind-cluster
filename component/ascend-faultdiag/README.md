# MindCluster Ascend FaultDiag

## 目录

- [MindCluster Ascend FaultDiag](#mindcluster-ascend-faultdiag)
  - [目录](#目录)
  - [简介](#简介)
  - [软件架构](#软件架构)
    - [上下游依赖](#上下游依赖)
  - [编译指南](#编译指南)
  - [安装部署](#安装部署)
  - [使用指南](#使用指南)
  - [说明](#说明)

## 简介

- MindCluster Ascend FaultDiag（故障诊断工具，命令行名为ascend-fd）是一款面向昇腾AI集群的日志故障诊断工具，部署在集群各节点上，用于提取训练及推理过程相关日志的关键信息，并分析故障根因节点以及故障事件。
- 主要功能：
  - 日志清洗：提取训练及推理过程相关日志的关键信息，覆盖主机OS日志、CANN应用侧与Device侧日志、MindCluster组件日志、MindIE组件日志等，输出结构化清洗结果。
  - 故障诊断：根据集群所有节点清洗后的关键信息，分析故障根因节点以及故障事件，支持根因节点分析、故障事件分析、设备资源分析、网络拥塞分析四类诊断。
  - 命令行与SDK双入口：提供ascend-fd命令行工具，同时通过ascend-faultdiag-toolkit包提供Python SDK，便于AI运维平台集成清洗与诊断能力。

## 软件架构

### 上下游依赖

![](../../docs/zh/faultdiag/figures/ascend-faultdiag/全量应用场景方案.png "全量应用场景方案")

1. 上游输入为各节点采集的原始日志：训练/推理控制台日志、CANN应用侧与Device侧日志、主机OS日志、MindCluster组件日志、MindIE组件日志以及NPU网口、环境信息等。
2. 各节点分别执行日志清洗，提取error、trace等关键信息，输出结构化清洗结果。
3. 将各节点的清洗结果集中转储至同一台设备，执行故障诊断。
4. 下游输出为故障诊断结果（根因节点、故障事件、故障描述），供用户定位问题，或供AI运维平台做后续处理。

## 编译指南

1. 通过git拉取源码，获得ascend-faultdiag。

   示例：源码放在/home/mind-cluster/component/ascend-faultdiag目录下。

2. 执行以下命令，安装编译所需的三方依赖库（要求Python版本不低于3.7.5）。

   ```shell
   cd /home/mind-cluster/component/ascend-faultdiag
   pip3 install -r src/requirements.txt && pip3 install 'setuptools>=60.3.0' 'wheel>=0.45.1'
   ```

3. 执行构建脚本，在"output"目录下生成组件whl包和SDK whl包。

   ```shell
   bash build/build.sh
   ```

4. 执行以下命令，查看 **output** 目录生成的软件列表。

   ```shell
   ll /home/mind-cluster/component/ascend-faultdiag/output
   ```

   ```text
   ascend_faultdiag-<version>-py3-none-linux_{arch}.whl
   alan_faultdiag-<version>-py3-none-linux_{arch}.whl
   ascend_faultdiag_toolkit-<version>-py3-none-any.whl
   ```

   其中`{arch}`为软件包架构（x86_64或aarch64，可通过`arch`命令查看），ascend_faultdiag为组件中文版安装包，alan_faultdiag为组件英文版安装包，ascend_faultdiag_toolkit为SDK工具包。

## 安装部署

当前MindCluster Ascend FaultDiag组件的安装部署支持两种方式：

1. Whl包安装（推荐），请参见[MindCluster Ascend FaultDiag安装部署指南 - 安装](../../docs/zh/faultdiag/ascend-faultdiag/04_installation_guide/01_installation.md)。
2. 使用MindCluster Ascend Deployer安装，请参见[MindCluster Ascend Deployer - 安装软件](https://gitcode.com/Ascend/ascend-deployer/blob/dev/docs/zh/05_installation_and_upgrade/02_install_softwares.md)。

## 使用指南

ascend-fd的典型使用流程为“日志清洗→清洗结果转储→故障诊断”，快速上手请参见[快速入门](../../docs/zh/faultdiag/ascend-faultdiag/03_quick_start/quick_start.md)。基于自身关键特性的使用指导，请参见MindCluster Ascend FaultDiag用户指南对应章节：

- 日志清洗（提取原始日志和监测指标信息中的有效信息），请参见[日志清洗](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/03_log_parsing.md)。
- 故障诊断（分析故障根因节点和故障事件），请参见[故障诊断](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/04_fault_diagnosis.md)。
- 单机场景的一站式清洗与诊断，请参见[单机故障诊断](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/05_single_server_diagnosis.md)。
- 超节点场景的故障诊断（含交换机日志分析），请参见[超节点故障诊断](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/06_superpod_diagnosis.md)。
- 屏蔽无效CANN日志，请参见[屏蔽故障日志](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/08_fault_log_masking.md)。
- 自定义清洗配置与自定义故障实体，请参见[自定义配置文件](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/09_custom_configuration.md)与[自定义故障实体](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/07_custom_fault_entities.md)。
- 基于SDK的清洗与诊断（业务日志清洗、根因节点、故障事件），请参见[业务日志清洗](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/10_service_flow_parsing.md)、[根因节点清洗及诊断](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/11_root_cause_parsing_diagnosis.md)与[故障事件清洗及诊断](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/12_fault_event_parsing_diagnosis.md)。
- 命令行与SDK的详细参数说明，请参见[API参考](../../docs/zh/faultdiag/ascend-faultdiag/06_api/menu_api.md)。

## 说明

- 请将ascend-fd独立部署在各服务器上使用。如果部署在共享目录中供多台服务器共用，可能导致功能异常或性能问题。
- ascend-fd仅支持对整机满卡训练或推理任务提供故障诊断，非满卡场景执行诊断可能导致故障根因定位错误或失败。
- 请同步各训练/推理服务器的系统时间（含Host与Device、宿主机与容器），时间不一致可能导致分析结果不准确。
- 使用诊断功能时，因Linux系统最大文件描述符数限制（默认为1024），集群规模建议不超过128台服务器（1024卡），超过时需使用`ulimit -n <值>`命令调整。
- 各类日志的版本配套要求请参见[使用说明](../../docs/zh/faultdiag/ascend-faultdiag/05_usage/01_usage_overview.md)，常见问题请参见[FAQ](../../docs/zh/faultdiag/ascend-faultdiag/07_references/02_faq.md)。
