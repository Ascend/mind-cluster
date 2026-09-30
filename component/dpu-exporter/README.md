# DPU Exporter

# 目录

- [简介](#简介)
- [软件架构](#软件架构)
  - [上下游依赖](#上下游依赖)
  - [组件架构](#组件架构)
- [编译指南](#编译指南)
- [安装部署](#安装部署)
- [使用指南](#使用指南)
- [说明](#说明)

## 简介

- 在任务运行过程中，DPU的健康状态直接影响任务的稳定性。MindCluster提供DPU Exporter组件用于监测DPU的运行状态与各项统计指标。当前提供的指标可分为两类：
  - **全局指标**：涵盖RoCE错包、丢包、接收ECN、发送/接收CNP及PSN异常重传等统计指标。
  - **Interface级指标**：涵盖每个网卡端口的链路运行状态、收发流量与异常错误状态等指标。
- 主要功能：
  - 从网卡管理工具与文件系统接口分别获取DPU的全局指标与Interface级指标。
  - 提供Prometheus指标接口，用于监控DPU的运行状态与各项统计指标。

## 软件架构

### 上下游依赖

![](../../docs/zh/figures/scheduling/组件上下游依赖-9.png "组件上下游依赖-9")

1. 从网卡管理工具hinicadm5获取DPU全局指标，放入缓存。
2. 从sysfs文件系统获取Interface级指标，放入缓存。
3. 实现Prometheus的指标接口，供其周期性获取缓存中的数据信息。

### 组件架构

DPU Exporter 组件在 Kubernetes 集群中以容器方式（DaemonSet）部署在配置了 DPU 的节点上，模块层级如下：

```mermaid
%%{init: {
  "theme": "base",
  "themeVariables": {
    "edgeLabelBackground": "transparent",
    "labelBoxBkgColor": "transparent",
    "labelTextColor": "#000000",
    "edgeLabelFontWeight": "bold",
    "primaryTextColor": "#1f2937",
    "titleColor": "#1f2937"
  },
  "themeCSS": ".edgeLabel, .edgeLabel .label, .edgeLabel p, .edgeLabel span, .edgeLabel text { color: #000000 !important; fill: #000000 !important; } .edgeLabel rect, .edgeLabel foreignObject div { fill: transparent !important; background-color: transparent !important; }"
}}%%
flowchart BT
    %% ===== 最底层：硬件 =====
    DPU["DPU"]

    %% ===== 数据来源：网卡管理工具与文件接口 =====
    TOOL["网卡管理工具 hinicadm5"]
    SYS["sysfs 文件接口<br/>/sys/class/net/"]

    subgraph op["DPU Exporter（DaemonSet，部署于DPU节点）"]
        direction BT

        subgraph ext["扩展能力层"]
            direction LR
            E1["配置管理与热加载<br/>采集周期"]
            E2["指标白名单<br/>metricWhiteList"]
            E3["日志<br/>Logger"]
        end

        subgraph collect["数据采集层（DpuCollector 采集引擎周期调度）"]
            direction LR
            DM["DeviceManager<br/>设备管理"]
            C1["Hinicadm5 采集器<br/>全局指标"]
            C2["Sysfs 采集器<br/>Interface 级指标"]
        end

        CACHE["DpuCache 指标缓存"]

        subgraph serve["指标服务层"]
            direction LR
            S1["Prometheus 服务<br/>HTTP /metrics"]
            S2["Healthz 服务<br/>存活探针"]
        end
    end

    %% ===== 最顶层：数据消费方 =====
    PROM["Prometheus"]
    K8S["Kubernetes"]

    %% ===== 自底向上的数据流 =====
    DPU -->|"① 卡级 RoCE 计数器"| TOOL
    DPU -->|"② 端口状态与统计"| SYS
    TOOL -->|"③ exec 调用"| DM
    SYS -->|"④ 文件读取"| DM
    DM --> C1
    DM --> C2
    C1 --> CACHE
    C2 --> CACHE
    CACHE --> S1
    S1 -->|"⑤ HTTP 拉取 metrics"| PROM
    S2 -->|"⑥ 存活探针检查"| K8S

    style op fill:#ffffff,stroke:#9ca3af,stroke-width:2px,color:#111827
    style collect fill:#eaf3fc,stroke:#7fb2e5,stroke-width:1.5px,color:#1e3a5f
    style serve fill:#fdf0dd,stroke:#e5b06e,stroke-width:1.5px,color:#7c2d12
    style ext fill:#f1e9fc,stroke:#b79ae0,stroke-width:1.5px,color:#4c1d95
    style CACHE fill:#e3f4ec,stroke:#7ac9a3,stroke-width:2px,color:#14532d

    classDef hardware fill:#c9f2e2,stroke:#18a058,stroke-width:2px,color:#0b4a2c
    classDef upstream fill:#cfe6f7,stroke:#2b7de9,stroke-width:2px,color:#0f3a66
    classDef downstream fill:#fce8e4,stroke:#e07b5f,stroke-width:2px,color:#7c2d12
    classDef collector fill:#cfe6f7,stroke:#2b7de9,stroke-width:2px,color:#0f3a66
    classDef server fill:#fbe3bd,stroke:#e08c00,stroke-width:2px,color:#6b3d00
    classDef extension fill:#e4d6f7,stroke:#7c3aed,stroke-width:2px,color:#3b1e73

    class DPU hardware
    class TOOL,SYS upstream
    class PROM,K8S downstream
    class DM,C1,C2 collector
    class S1,S2 server
    class E1,E2,E3 extension
```

## 编译指南

1. 通过 git 拉取源码，获得 dpu-exporter。

    示例：源码放在 /home/mind-cluster/component/dpu-exporter 目录下。

2. 执行以下命令，进入构建目录，执行构建脚本，在"output"目录下生成二进制 dpu-exporter、yaml 文件和 Dockerfile。

    ```shell
    cd /home/mind-cluster/component/dpu-exporter/build/
    chmod +x build.sh
    ./build.sh
    ```

3. 执行以下命令，查看 **output** 目录生成的软件列表。

    ```shell
    ll /home/mind-cluster/component/dpu-exporter/output
    ```

    ```text
    -rw-r--r--. 1 root root      128 Aug 22 17:31 config.json
    -rw-r--r--. 1 root root      761 Aug 22 17:31 Dockerfile
    -rw-r--r--. 1 root root      833 Aug 22 17:31 Dockerfile.openeuler
    -rw-r--r--. 1 root root 11010350 Aug 22 17:31 dpu-exporter
    -rw-r--r--. 1 root root     4640 Aug 22 17:31 dpu-exporter-v26.2.0.yaml
    ```

## 安装部署

当前 DPU Exporter 组件的安装部署支持两种方式：镜像部署和二进制部署（包含安装前置检查、使用约束、部署及安装验证等），请参见[手动安装DPU Exporter](../../docs/zh/scheduling/05_developer_guide/00_installation_deployment/00_manual_installation/12_dpu_exporter.md)。

## 使用指南

DPU Exporter 组件的使用指南，请参见[DPU资源监测](../../docs/zh/scheduling/04_usage/01_resource_monitoring/01_dpu_resource_monitoring/00_before_you_start.md)。

## 说明

- 当前 DPU Exporter 仅支持 HTTP 启动，如果需要使用 HTTPS 启动，请自行完成代码修改并适配 Prometheus。
- 组件以 hostNetwork 方式部署，指标端口（默认 8083）会在部署节点的所有网络接口上侦听，请根据实际组网安全策略做好访问控制。
