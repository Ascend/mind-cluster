# K8s RDMA Shared Device Plugin

> **说明：** 本文档为中文版本说明。英文原版 README 请参考 [README-EN.md](README-EN.md)，该文档为开源组件 [Mellanox/k8s-rdma-shared-dev-plugin](https://github.com/Mellanox/k8s-rdma-shared-dev-plugin) 的原始文档。

# 组件介绍

k8s-rdma-shared-dev-plugin 是一个 RDMA 设备插件，支持 InfiniBand (IB) 和 RoCE HCA 设备。该插件以 DaemonSet 方式运行在 Kubernetes 集群中，为容器提供 RDMA 网络设备的共享访问能力。

主要功能：

- 支持 IB 和 RoCE HCA 设备的发现与管理
- 支持设备资源的共享分配（多个 Pod 可共享同一 RDMA 设备）
- 支持通过配置文件灵活管理多个 RDMA 设备资源
- 支持 CDI（Container Device Interface）规范

# 编译

1. 通过 git 拉取源码，获得 k8s-rdma-shared-dev-plugin。

   示例：源码放在 /home/mind-cluster/component/k8s-rdma-shared-dev-plugin 目录下

2. 执行以下命令，进入构建目录，执行构建脚本，在"output"目录下生成二进制文件、yaml 文件和 Dockerfile 等文件。

   **cd** _/home/mind-cluster/component/_**k8s-rdma-shared-dev-plugin/build/**

       chmod +x build.sh

       ./build.sh

3. 执行以下命令，查看 **output** 生成的软件列表。

   **ls** _/home/mind-cluster/component/_**k8s-rdma-shared-dev-plugin/output**

# 使用说明

## 部署设备插件

**1.** 配置 CNI 插件（如 Contiv、Calico、Cluster）

确保将 ib0 或适当的 IPoIB 网络设备配置为创建 overlay/虚拟网络设备的父设备。

**2.** 创建 ConfigMap 并部署设备插件

部署设备插件并创建 config map 来描述"hca"模式配置（按节点配置）：

```bash
cd deployment/k8s/base
kubectl apply -k .
```

**3.** 创建测试 Pod

创建请求 1 个 vhca 资源的测试 Pod：

```bash
kubectl create -f example/test-hca-pod.yaml
```

## 使用 CDI 支持部署

要使用支持 [CDI](https://github.com/cncf-tags/container-device-interface) 的设备插件：

```bash
cd deployment/k8s/base/overlay
kubectl apply -k .
```

## 配置说明

插件支持以下配置字段：

```json
{
  "periodicUpdateInterval": 300,
  "configList": [{
      "resourceName": "hca_shared_devices_a",
      "resourcePrefix": "example_prefix",
      "rdmaHcaMax": 1000,
      "devices": ["ib0", "ib1"]
    },
    {
      "resourceName": "hca_shared_devices_b",
      "rdmaHcaMax": 500,
      "selectors": {
        "vendors": ["15b3"],
        "deviceIDs": ["1017"],
        "ifNames": ["ib3", "ib4"]
      }
    }
  ]
}
```

### 配置字段说明

| 字段 | 必填 | 说明 | 类型 | 默认值 | 示例 |
|------|------|------|------|--------|------|
| resourceName | 是 | 资源端点名称，不能包含特殊字符，在资源前缀范围内必须唯一 | string | - | "hca_shared_devices_a" |
| resourcePrefix | 否 | 资源端点前缀，不能包含特殊字符 | string | "rdma" | "example_prefix" |
| rdmaHcaMax | 是 | 设备插件可提供的最大 RDMA 资源数量 | Integer | - | 1000 |
| selectors | 否 | 设备选择器映射，用于过滤设备 | json object | - | {"vendors": ["15b3"]} |
| devices | 否 | 设备名称列表，等同于 ifNames 选择器 | string list | - | ["ib0", "ib1"] |

**注意：** 对于给定资源，必须指定 `selectors` 或 `devices` 其中之一，推荐使用 `selectors`。

### 设备选择器

| 字段 | 说明 | 类型 | 示例 |
|------|------|------|------|
| vendors | 目标设备的厂商十六进制代码 | string list | "vendors": ["15b3"] |
| deviceIDs | 目标设备的设备十六进制代码 | string list | "deviceIDs": ["1017"] |
| drivers | 目标设备驱动名称 | string list | "drivers": ["mlx5_core"] |
| ifNames | 目标设备名称 | string list | "ifNames": ["enp2s2f0"] |
| linkTypes | PCI 设备关联的网络设备链路类型 | string list | "linkTypes": ["ether"] |

### 选择器匹配逻辑

设备插件根据提供的选择器过滤主机设备。如果存在缺失的选择器，插件会忽略它们。在特定选择器内执行逻辑 OR 操作，在不同选择器之间执行逻辑 AND 操作。

## 节点标签部署

RDMA 共享设备插件应部署在满足以下条件的节点上：

1. 具有 RDMA 功能的硬件
2. RDMA 内核栈已加载

可使用 [Node Feature Discovery (NFD)](https://github.com/kubernetes-sigs/node-feature-discovery) 发现节点功能并将其作为节点标签暴露：

1. 部署 NFD（v0.6.0 或更新版本）

```bash
export NFD_VERSION=v0.6.0
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/$NFD_VERSION/nfd-master.yaml.template
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/$NFD_VERSION/nfd-worker-daemonset.yaml.template
```

2. 检查节点上的新标签

```bash
kubectl get nodes --show-labels
```

然后可以在具有 `feature.node.kubernetes.io/custom-rdma.available=true` 标签的节点上部署 RDMA 设备插件，该标签表示节点具有 RDMA 功能且 RDMA 模块已加载。

# 说明

1. 本组件为开源组件 [Mellanox/k8s-rdma-shared-dev-plugin](https://github.com/Mellanox/k8s-rdma-shared-dev-plugin) 的集成版本
2. 容器镜像默认使用 `alpine` 基础镜像构建
3. 当前容器方式部署本组件，本组件的认证鉴权方式为 ServiceAccount，该认证鉴权方式为 ServiceAccount 的 token 明文显示，建议用户自行进行安全加强
