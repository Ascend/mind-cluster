# 配置静默故障检测

## 概述

静默故障指ClusterD根据软件故障与硬件故障信息综合判定的特殊公共故障。本章节介绍如何通过ConfigMap（clusterd-config-cm）中的silent_fault_policy.conf配置项，开启静默故障检测并调整检测与释放参数。静默故障的定义、判定条件与检测原理请参见[静默故障](../02_fault_types/07_silent_faults.md)。

## 配置文件说明

静默故障检测默认关闭，需用户显式开启。ClusterD根据ConfigMap（clusterd-config-cm）中silent_fault_policy.conf配置项的内容，执行静默故障检测。silent_fault_policy.conf参数说明请参见表1。

>[!NOTE]
>
>静默故障的故障码与级别在publicFaultConfiguration.json中配置，用户可通过publicCustomization.json调整级别，参考[配置公共故障](./08_public_faults.md)。

**表 1**  silent_fault_policy.conf的参数说明

|一级参数|二级参数|类型|说明|
|--|--|--|--|
|enabled|-|bool|静默故障检测开关。取值包括：<ul><li>true：开启静默故障检测功能。</li><li>false：关闭静默故障检测功能。</li></ul><p>默认配置为false。关闭该开关时，ClusterD会清除clusterd-manual-info-cm和statistic-fault-info ConfigMap中的的静默故障。</p>|
|detect|min_task_cards|int|静默故障判定需要满足的任务总NPU数量。取值范围为(0, 10000000)，默认配置为16。|
|detect|consecutive_times|int|判定为静默故障需要在判定时间窗内达到的有效重调度次数。取值范围为(0, 10000)，默认配置为3。|
|detect|hardware_fault_window_seconds|int|重调度前后无硬件故障窗口。取值范围为(3, 86400)，默认配置为30，单位为s（秒）。|
|detect|window_seconds|int|静默故障判定时间窗。取值范围为[30, 31536000)，默认配置为10800（3小时），单位为s（秒）。|
|release|fault_free_seconds|int|静默故障自动释放时长。取值包括：<ul><li>-1：不自动解除。</li><li>(0, 31536000)：显式释放时长，单位为s（秒）。</li></ul><p>默认配置为172800（48小时）。判定为静默故障后再次命中会刷新计时，释放时间从最近一次命中重新计算。</p>|

>[!NOTE]
>
>若enabled字段缺失，ClusterD识别为false；若其他字段不配置或取值不合法，ClusterD使用该字段的默认值。

## （可选）开启静默故障检测

**操作步骤**

以开启静默故障检测为例。

1. 登录环境，执行以下命令，查询当前配置。

    ```shell
    kubectl describe cm -n cluster-system clusterd-config-cm
    ```

    - 如果存在clusterd-config-cm，则执行步骤3进行编辑。
    - 如果不存在clusterd-config-cm，则执行步骤2进行创建。

    >[!NOTE]
    >正常情况下存在clusterd-config-cm，因为安装ClusterD时已默认创建了该ConfigMap。若不存在，建议确认ClusterD的安装过程是否存在错误。

2. 创建静默故障检测所需的ConfigMap（clusterd-config-cm）。

    将以下内容保存为文件cm.yaml：

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: clusterd-config-cm
      namespace: cluster-system
    data:
      silent_fault_policy.conf: |
        enabled: true
        detect:
          min_task_cards: 16
          consecutive_times: 3
          hardware_fault_window_seconds: 30
          window_seconds: 10800
        release:
          fault_free_seconds: 172800
    ```

    执行以下命令：

    ```shell
    kubectl apply -f cm.yaml
    ```

    回显示例如下，说明创建成功。

    ```ColdFusion
    configmap/clusterd-config-cm created
    ```

3. 执行以下命令，编辑clusterd-config-cm。

    ```shell
    kubectl edit cm -n cluster-system clusterd-config-cm
    ```

    根据实际情况，修改静默故障检测配置。参数说明请参见表1。

    ```yaml
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
      silent_fault_policy.conf: |
        # 修改静默故障检测开关
        enabled: true
        detect:
          # 修改静默故障判定需要满足的任务总NPU数量
          min_task_cards: 16
          # 修改判定为静默故障需要在判定时间窗内达到的有效重调度次数
          consecutive_times: 3
          # 修改重调度前后无硬件故障窗口
          hardware_fault_window_seconds: 30
          # 修改静默故障判定时间窗
          window_seconds: 10800
        release:
          # 修改静默故障自动释放时长
          fault_free_seconds: 172800
    kind: ConfigMap
    metadata:
      name: clusterd-config-cm
      namespace: cluster-system
    ```

4. 修改完成后，按“Esc”键，输入:wq!保存并退出。
5. 等clusterd-config-cm更新生效（ClusterD的检测周期为300s）后，查看操作是否成功。

    1. 执行以下命令，查询ClusterD组件名称。

        ```shell
        kubectl get pods -A | grep clusterd
        ```

        回显示例如下：

        ```ColdFusion
        mindx-dl      clusterd-559bf4bd6-z9hv4   1/1     Running   0             4m23s
        ```

    2. 通过查询到的组件名称，查询ClusterD的组件日志信息。

        ```shell
        kubectl logs -f -n mindx-dl clusterd-559bf4bd6-z9hv4
        ```

        >[!NOTE]
        >若日志出现“load silent fault policy config success”，表示静默故障检测配置加载成功。

## （可选）解除静默故障

静默故障支持以下两种解除方式。

### 自动释放

命中静默故障后累计fault_free_seconds时长（默认48小时）自动解除；隔离期内再次命中会刷新计时、顺延释放时间，持续反复命中则持续顺延，fault_free_seconds时长内无新命中才真正解除。

### 手动解除

静默故障按整节点隔离，需删除clusterd-manual-info-cm中该节点的所有芯片才会解除。静默故障条目的识别方式为Detail中FaultLevel取值为SilentFault。

1. 执行以下命令，编辑ConfigMap clusterd-manual-info-cm。

    ```shell
    kubectl edit cm -n cluster-system clusterd-manual-info-cm
    ```

2. 删除该节点Total字段value值（该节点全部芯片列表）中的所有芯片（Total清空）。

3. 修改完成后，按“Esc”键，输入:wq!保存并退出。
4. 等待15s后，执行以下命令，查看clusterd-manual-info-cm中该节点的静默故障条目是否还存在于Total和Detail字段中。若不存在，则静默故障解除成功，可继续正常使用该节点。

    ```shell
    kubectl describe cm -n cluster-system clusterd-manual-info-cm
    ```

    >[!NOTE]
    >
    >- 只删除部分芯片时不生效，下一轮ClusterD会把完整节点自动写回。
    >- 若节点上同时存在静默故障和人工隔离的芯片，若要解除静默故障需清除所有芯片，此时静默故障和人工隔离会一并解除。
    >- clusterd-manual-info-cm的详细说明请参见[clusterd-manual-info-cm](../../../06_api/04_clusterd/00_cluster_resources.md#clusterd-manual-info-cm)。
