# 集群调度组件 Ascend Device Plugin

> [English](./OVERVIEW.md) | 中文

## 快速参考

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.1.0
```

---

## Ascend Device Plugin

Ascend Device Plugin 是 MindCluster 集群调度组件的核心组件之一，部署在计算节点上，用于提供适合昇腾设备的资源发现和上报策略。

### 应用场景

Kubernetes 需要感知资源信息来实现对资源信息的调度。除基础的 CPU 和内存信息以外，需通过 Kubernetes 提供的设备插件机制，供用户自定义新的资源类型，从而定制个性化的资源发现和上报策略。MindCluster 提供了部署在计算节点的 Ascend Device Plugin 服务，用于提供适合昇腾设备的资源发现和上报策略。

### 组件功能

- **设备发现**：从驱动中获取芯片的类型及型号，并上报给 kubelet 和资源调度的上层服务 ClusterD。支持从昇腾设备驱动中发现设备个数，将其发现的设备个数上报到 Kubernetes 系统中。支持发现拆分物理设备得到的虚拟设备并上报 Kubernetes 系统。

- **健康检查**：从驱动中订阅芯片故障信息，并将芯片状态上报给 kubelet，同时将芯片状态和具体故障信息上报给资源调度的上层服务。支持检测昇腾设备的健康状态，当设备处于不健康状态时，上报到 Kubernetes 系统中，Kubernetes 系统会自动将不健康设备从可用列表中剔除。虚拟设备健康状态由拆分这些虚拟设备的物理设备决定。

- **设备分配**：支持在 Kubernetes 系统中分配昇腾设备；支持 NPU 设备重调度功能，设备故障后会自动拉起新容器，挂载健康设备，并重建训练任务。在资源挂载阶段，负责获取集群调度选中的芯片信息，并通过环境变量传递给 Ascend Docker Runtime 挂载。

- **故障处理**：可配置故障的处理级别，且可在故障反复发生，或者长时间连续存在的情况下提升故障处理级别。若故障芯片处于空闲状态，且重启后可恢复，对芯片执行热复位。

- **网络故障监控**：从灵衢驱动中订阅灵衢网络故障信息，并将网络状态上报给 kubelet，同时将灵衢网络状态和具体故障信息上报给资源调度的上层服务。

### 组件上下游依赖

1. 从 DCMI 中获取芯片的类型、数量、健康状态信息，或者下发芯片复位命令。
2. 上报芯片的类型、数量和状态给 kubelet。
3. 上报芯片的类型、数量和具体故障信息给 ClusterD。
4. 将调度器选中的芯片信息，以环境变量的方式告知给 Ascend Docker Runtime。

---

## 支持的 Tags 及 Dockerfile 链接

### Tag 规范

Tag 遵循以下格式：

```shell
<版本>-<操作系统>
```

| 字段     | 示例值           | 说明                         |
|--------|---------------|----------------------------|
| `版本`   | `v26.1.0`     | Ascend Device Plugin组件版本   |
| `操作系统` | `ubuntu22.04` | Ascend Device Plugin镜像操作系统 |

### Ascend Device Plugin 26.1.0

| Tag                      | Dockerfile                                      | 镜像内容                                                  |
|--------------------------|-------------------------------------------------|-------------------------------------------------------|
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | Ascend Device Plugin组件v26.1.0版本操作系统为ubuntu22.04的镜像    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | Ascend Device Plugin组件v26.1.0版本操作系统为openeuler24.03的镜像 |

---

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称 | 支持的版本 | 安装位置 | 说明 |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x（推荐使用1.19.x及以上版本） | 所有节点 | 了解 K8s 的使用请参见 [Kubernetes 文档](https://kubernetes.io/zh-cn/docs/) |
| Docker | 18.09.x~28.5.1 | 所有节点 | 可从 [Docker 社区或官网](https://docs.docker.com/engine/install/) 获取 |
| Containerd | 1.4.x~2.1.4（推荐使用1.6.x版本） | 所有节点 | 可从 Containerd 的 [官网](https://containerd.io/downloads/) 获取 |
| 昇腾AI处理器驱动和固件 | 请参见版本配套表 | 计算节点 | 请参见《CANN 软件安装指南》中的"安装NPU驱动和固件"章节 |

#### 硬件规格要求

| 名称 | 要求 |
| -- | -- |
| CPU | 0.5核 |
| 内存 | 0.5GB |

#### 安装驱动

宿主机已安装驱动和固件，详情请参见《CANN 软件安装指南》中的"[安装NPU驱动和固件](https://www.hiascend.com/document/detail/zh/canncommercial/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=Debian)"章节（商用版）或"[安装NPU驱动和固件](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=openEuler)"章节（社区版）。

### 如何本地构建

```bash
docker build --no-cache -t ascend-k8sdeviceplugin:{tag} ./ -f Dockerfile.{os}
```

### 部署 Ascend Device Plugin

1. 拉取镜像

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:{version}
```

2. 修改镜像标签

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:{version} ascend-k8sdeviceplugin:{tag}
```

3. 给节点打标签

根据昇腾处理器型号查询需要打的标签：

```bash
# 示例：为 Ascend 910 节点打标签
kubectl label nodes <node-name> accelerator=huawei-Ascend910
kubectl label nodes <node-name> host-arch=huawei-arm
```

4. 启动 Ascend Device Plugin

根据昇腾处理器型号选择对应的 YAML 文件：

将 YAML 文件中镜像的 `{tag}` 替换为实际标签。

```bash
# Ascend 310 处理器
kubectl apply -f device-plugin-310-{version}.yaml

# Ascend 910 处理器（不使用 Volcano）
kubectl apply -f device-plugin-910-{version}.yaml

# Ascend 910 处理器（使用 Volcano）
kubectl apply -f device-plugin-volcano-{version}.yaml
```

5. 验证部署

```bash
kubectl get pods -n kube-system | grep ascend-device-plugin
```

6. 检查节点资源

```bash
kubectl describe node <npu-node-name> | grep "huawei.com/Ascend"
```

---

## 支持的硬件

所有昇腾设备通用

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/document/detail/zh/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
