# NPU Exporter

## 目录

- [简介](#简介)
- [软件架构](#软件架构)
  - [上下游依赖](#上下游依赖)
  - [组件架构](#组件架构)
- [编译指南](#编译指南)
- [安装部署](#安装部署)
- [使用指南](#使用指南)
- [说明](#说明)

## 简介

- NPU Exporter 是华为自研的专门收集昇腾NPU各种监测信息和指标，并将监测对象的指标转换为 Prometheus、Telegraf 能够识别的数据格式，供上层监测系统集成的服务组件。
- 在任务运行过程中，需要密切关注芯片的网络和算力使用情况，为任务的调优提供数据支持。NPU Exporter 用于上报芯片的各项指标状态。
- 主要功能：
  - 从驱动中获取芯片、网络的各项数据信息，并支持 Prometheus、Telegraf 两种方式上报。
  - 插件化管理，支持通过插件开发新的指标，支持针对不同类型的指标配置不同的采集周期。

## 软件架构

### 上下游依赖

  ![](../../docs/zh/figures/scheduling/组件上下游依赖-0.png "组件上下游依赖-0")

  - NPU Exporter 从驱动中获取芯片信息与网络信息，并放入本地缓存。
  - 从 K8s 标准化接口 CRI 中获取容器信息，并放入本地缓存。
  - 实现 Prometheus 或 Telegraf 的接口，供二者周期性获取缓存中的数据信息。

  具体数据获取方式如下：

  1. 通过 gRPC 服务调用 K8s 中的标准化接口 CRI，获取容器相关信息。
  2. 通过 exec 调用 hccn_tool 工具，获取芯片的网络信息。
  3. 通过 dlopen/dlsym 调用 DCMI 接口，获取芯片信息，并上报给 Prometheus。

### 组件架构

NPU Exporter 组件在 Kubernetes 集群中以容器方式（DaemonSet）部署在计算节点上，模块层级如下：

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
    NPU["昇腾 NPU"]

    %% ===== 数据来源：驱动与容器运行时接口 =====
    DRV["dcmi / hccn_tool"]
    CRI["K8s 标准化接口 CRI"]

    subgraph op["NPU Exporter（DaemonSet，部署于计算节点）"]
        direction BT

        subgraph ext["扩展能力层"]
            direction LR
            E1["插件管理<br/>注册自定义指标"]
            E2["动态配置加载<br/>采集开关与周期"]
            E3["健康探针<br/>livenessProbe"]
        end

        subgraph collect["数据采集层"]
            direction LR
            C1["DCMI 采集器<br/>dlopen/dlsym 调用<br/>芯片信息"]
            C2["hccn_tool 采集器<br/>exec 调用<br/>网络信息"]
            C3["CRI 客户端<br/>gRPC 调用<br/>容器信息"]
            C4["自定义指标<br/>插件采集器"]
        end

        CACHE["本地缓存（按指标组组织）<br/>version · utilization · npu · ddr · sio · hbm · hccs · pcie<br/>vnpu · nodeBase · ub · roce · optical · network_bandwidth · network_link"]

        subgraph serve["指标服务层"]
            direction LR
            S1["Prometheus 服务<br/>HTTP /metrics"]
            S2["Telegraf 服务<br/>execd 插件"]
        end
    end

    %% ===== 最顶层：数据消费方 =====
    PROM["Prometheus"]
    TELE["Telegraf"]

    %% ===== 自底向上的数据流 =====
    NPU -->|"① 芯片与网络资源"| DRV
    DRV -->|"② 芯片数据（DCMI）"| C1
    DRV -->|"③ 网络数据（hccn_tool）"| C2
    CRI -->|"④ 容器信息（CRI）"| C3

    C1 --> CACHE
    C2 --> CACHE
    C3 --> CACHE
    C4 --> CACHE
    CACHE --> S1
    CACHE --> S2

    S1 -->|"⑤ HTTP 拉取 metrics"| PROM
    S2 -->|"⑥ execd 输出数据"| TELE

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

    class NPU hardware
    class DRV,CRI upstream
    class PROM,TELE downstream
    class C1,C2,C3,C4 collector
    class S1,S2 server
    class E1,E2,E3 extension
```

## 编译指南

1. 通过 git 拉取源码，获得 npu-exporter。

    示例：源码放在 /home/mind-cluster/component/npu-exporter 目录下。

2. 执行以下命令，进入构建目录，执行构建脚本，在 "output" 目录下生成二进制 npu-exporter、yaml 文件、配置文件和 Dockerfile 等文件。

    ```shell
    cd /home/mind-cluster/component/npu-exporter/build/
    chmod +x build.sh
    ./build.sh
    ```

    构建版本号默认取自 `service_config.ini` 文件，若该文件不存在则默认使用 v6.0.0。

3. 执行以下命令，查看 "output" 目录生成的软件列表。

    ```shell
    ll /home/mind-cluster/component/npu-exporter/output
    ```

    ```text
    drwxr-xr-x 2 root root     4096 Jan 29 19:12 ./
    drwxr-xr-x 9 root root     4096 Jan 29 19:09 ../
    -r-x------ 1 root root 25481072 Jan 29 19:09 npu-exporter
    -r-------- 1 root root     3438 Jan 29 19:09 npu-exporter-v6.0.0.yaml
    -r-------- 1 root root     3438 Jan 29 19:09 npu-exporter-310P-1usoc-v6.0.0.yaml
    -r-------- 1 root root      623 Jan 29 19:09 metricConfiguration.json
    -r-------- 1 root root      623 Jan 29 19:09 pluginConfiguration.json
    -r-------- 1 root root      623 Jan 29 19:09 agreement.txt
    -r-------- 1 root root      623 Jan 29 19:09 Dockerfile
    -r-------- 1 root root      623 Jan 29 19:09 Dockerfile.openeuler
    -r-------- 1 root root      623 Jan 29 19:09 Dockerfile-310P-1usoc
    -r-------- 1 root root      623 Jan 29 19:09 Dockerfile-310P-1usoc.openeuler
    -r-x------ 1 root root     2579 Jan 29 19:09 run_for_310P_1usoc.sh
    ```

## 安装部署

当前 NPU Exporter 组件的安装部署支持两种方式：

1. Helm 安装，请参考 [使用 Helm 安装](../../docs/zh/scheduling/03_installation_guide/02_installation/00_helm_installation.md)。

2. 手动安装与部署（包含使用约束、镜像/二进制方式部署、启动参数及动态配置加载等），请参见 [手动安装NPU Exporter](../../docs/zh/scheduling/05_developer_guide/00_installation_deployment/00_manual_installation/03_npu_exporter.md)。

## 使用指南

NPU Exporter 组件的最佳实践（包含使用前必读、实现原理，以及通过 Prometheus 或 Telegraf 使用资源监测的完整流程），请参见 [资源监测](../../docs/zh/scheduling/04_usage/01_resource_monitoring/menu_resource_monitoring.md)。

## 说明

- 当前 NPU Exporter 默认以 HTTP 方式提供服务，如需使用 HTTPS，可通过启动参数 `--tls-cert-file` 与 `--tls-private-key-file` 配置证书，具体请参见 [健康探针安全加固](../../docs/zh/scheduling/07_references/04_security_hardening.md#ZH-CN_TOPIC_0000002511346467)。
- 使用 Telegraf 集成时，在 `plugins/inputs/all/npu.go` 中添加 `import _ "github.com/influxdata/telegraf/plugins/inputs/npu"`。
- Kubernetes 场景下组件以 ServiceAccount 方式认证鉴权，该方式的 token 以明文显示，建议用户自行进行安全加强。
