# 静默故障

## 概述

静默故障是指任务的Pod反复主动异常退出、但节点无硬件故障上报时，ClusterD根据软件故障和硬件故障信息综合判定的特殊公共故障。判定为静默故障后，ClusterD仅在cluster-info-device-xxx ConfigMap的故障列表中追加该故障，不改变芯片健康状态，因此不会触发重调度；同时，还会将该故障写入公共故障缓存（statistic-fault-info），并在clusterd-manual-info-cm中记录静默故障条目。

## 判定机制

ClusterD通过监听job-reschedule-reason ConfigMap获取任务的重调度记录，并结合节点的硬件故信息，综合判定静默故障。判定为静默故障需同时满足以下条件（以下阈值均为默认值，可通过配置调整，参数说明请参见[配置静默故障检测](../03_configuration/09_silent_faults.md)）：

- 重调度任务的总NPU数量不小于16，且为整机任务。
- 单次重调度仅1个pod-failed，批量pod重调度不统计。
- 重调度发生前后30秒内，节点无硬件故障上报。硬件故障包括芯片级、节点级、交换机级与外部公共故障，静默故障自身除外。
- 按节点维度统计，在最近3小时内连续出现3条有效软件故障首报错记录。

## 所需组件

为保证静默故障检测功能的正常使用，需要安装以下组件。

- 必选组件：Volcano、Ascend Device Plugin、ClusterD
- 可选组件：NodeD

## 相关操作

- [配置静默故障检测](../03_configuration/09_silent_faults.md)
- [配置公共故障](../03_configuration/08_public_faults.md)
