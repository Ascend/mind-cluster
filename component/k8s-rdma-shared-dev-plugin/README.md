# K8s RDMA Shared Dev Plugin

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

k8s-rdma-shared-dev-plugin 是 Kubernetes 设备管理插件，支持以下设备使用资源监测：

- UB RDMA DPU

组件拥有以下功能：

- 设备发现：支持发现计算节点上的 UB RDMA 设备，将发现的设备上报到 Kubernetes 系统中。支持共享与独占两种工作模式，共享模式下将节点上的 RDMA 设备作为共享资源上报；独占模式下按 NPU 与 DPU 的映射关系为 Pod 分配节点上发现的真实 UB 设备。
- 健康检查：支持检测 UB 设备的健康状态，当设备处于不健康状态时，将故障信息写入 Kubernetes ConfigMap。
- 设备分配：支持在 Kubernetes 系统中分配 RDMA 设备；独占模式下按 NPU 与 DPU 的映射关系为 Pod 分配设备，并将分配结果写入 `k8s.v1.cni.cncf.io/device-status` 注解，供 Multus CNI 与 UB Host Device CNI 使用完成设备挂载。

组件 UB 功能示意：

```mermaid
flowchart TD
    DEV[节点 UB RDMA 设备]

    subgraph discover[设备发现]
        direction TB
        D1[扫描 /sys/bus/ub/devices<br/>发现节点上的 UB 设备] --> D2[结合 /sys/class/infiniband<br/>与 /sys/class/net 获取设备信息<br/>上报 kubelet]
    end

    subgraph shared[共享模式]
        direction TB
        S1[设备作为共享资源上报] --> S2[Pod 申请资源后<br/>挂载节点上全部 RDMA 设备]
    end

    subgraph exclusive[独占模式]
        direction TB
        E1[读取 Pod 的 huawei.com/npu 注解<br/>获取 NPU ID] --> E2[按 NPU-DPU 映射<br/>分配真实 UB 设备] --> E3[分配结果写入 device-status 注解<br/>供 Multus CNI 与 UB Host Device CNI 挂载]
    end

    subgraph fault[故障检测]
        direction TB
        F1[故障采集：按 faultDetectPeriod 周期执行 fault_detection.sh<br/>调用 hinicadm5 采集 UB 设备故障信息] --> F2[故障上报：故障信息写入 dpuinfo-node ConfigMap<br/>供 ClusterD 故障汇总]
    end

    DEV --> D1
    D2 --> shared
    D2 --> exclusive
    DEV --> F1
```

## 软件架构

### 上下游依赖

![](../../docs/zh/figures/scheduling/01_introduction/01_component_description/k8s-rdma-shared-dev-plugin.png "组件上下游依赖")

- 从设备目录读取设备信息。
- 上报RDMA设备的类型、数量和状态给kubelet。
- 通过hinicadm5工具进行RDMA设备的故障检测。
- RDMA设备故障上报，支持ClusterD进行故障汇总。
- 独占模式从业务Pod获取Volcano写入的NPU卡信息。

### 组件架构

```mermaid
flowchart TD
    MAIN[cmd/k8s-rdma-shared-dev-plugin<br/>程序入口：参数解析与初始化]

    CORE[pkg/resources/core<br/>服务生命周期管理、配置加载与校验]

    subgraph resources[pkg/resources 资源管理核心]
        SERVER[server.go<br/>共享模式 gRPC server]
        WATCHER[watcher.go<br/>设备发现与监听]
        UB[ub_device/<br/>UB 独占模式 server 与 NPU 分配器]
        COMMON[common/<br/>公共工具与常量]
    end

    FAULT[pkg/fault<br/>故障检测与上报]
    UTILS[pkg/utils<br/>Node 注解、NPU-NIC 映射]
    TYPES[pkg/types<br/>类型定义]

    MAIN --> CORE
    MAIN --> FAULT
    CORE --> SERVER
    CORE --> UB
    SERVER --> WATCHER
    UB --> WATCHER
    UB --> UTILS
    WATCHER --> COMMON
    SERVER --> COMMON
    UB --> COMMON
    FAULT --> UTILS
    SERVER -.-> TYPES
    UB -.-> TYPES
    FAULT -.-> TYPES
```

## 编译指南

1. 通过 git 拉取源码，并切换 master 分支，获得 k8s-rdma-shared-dev-plugin。

   示例：源码放在 `/home/mind-cluster/component/k8s-rdma-shared-dev-plugin` 目录下

2. 执行以下命令，进入构建目录，执行构建脚本，在 `output` 目录下生成二进制组件包、yaml 文件、Dockerfile 等文件。

    ```shell
    cd /home/mind-cluster/component/k8s-rdma-shared-dev-plugin/build
    ```

    ```bash
    chmod +x build.sh
    ./build.sh
    ```

3. 执行以下命令，查看 **output** 生成的软件列表。

    ```shell
    ll /home/mind-cluster/component/k8s-rdma-shared-dev-plugin/output
    ```

    ```text
    k8s-rdma-shared-dp
    k8s-rdma-shared-dp-v26.1.0.yaml
    config.json
    Dockerfile
    Dockerfile.openeuler
    agreement.txt
    fault_code.json
    fault_detection.sh
    npu-nic-mapping.json
    ```

## 安装部署

组件以 DaemonSet 方式部署在集群的每个计算节点上

1. Helm安装，请参考[MindCluster 安装部署 - 使用Helm安装](../../docs/zh/scheduling/03_installation_guide/02_installation/00_helm_installation.md)
2. 手动安装与部署（包含安装前置检查、镜像准备、yaml 部署及安装验证等），请参见[MindCluster 集群调度组件开发指南 - 手动安装](../../docs/zh/scheduling/05_developer_guide/00_installation_deployment/00_manual_installation/12_k8s_rdma_shared_dev_plugin.md)

## 使用指南

组件的具体使用配置请参见[K8s RDMA Shared Dev Plugin 安装指导](../../docs/zh/scheduling/05_developer_guide/00_installation_deployment/00_manual_installation/12_k8s_rdma_shared_dev_plugin.md)。

## 说明

当前容器方式部署本组件，本组件的认证鉴权方式为 ServiceAccount，该认证鉴权方式为 ServiceAccount 的 token 明文显示，建议用户自行进行安全加强。
