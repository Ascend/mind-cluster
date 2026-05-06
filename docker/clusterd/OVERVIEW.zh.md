# 集群调度组件 ClusterD

> [English](./OVERVIEW.md) | 中文

## 快速参考

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v26.1.0-ubuntu22.04
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v26.1.0-openeuler24.03
```

---

## ClusterD

ClusterD 是 MindCluster 集群调度组件之一，部署在管理节点上，用于收集并汇总集群任务、资源和故障信息及影响范围，从任务、芯片和故障维度统计分析，统一判定故障处理级别和策略。

### 应用场景

一个节点可能发生多个故障，如果由各个节点自发进行故障处理，会造成任务同时处于多种恢复策略的场景。为了协调任务的处理级别，MindCluster 提供了部署在管理节点的 ClusterD 服务。ClusterD 收集并汇总集群任务、资源和故障信息及影响范围，从任务、芯片和故障维度统计分析，统一判定故障处理级别和策略。

### 组件功能

- 从 Ascend Device Plugin 和 NodeD 组件获取芯片、节点和网络信息，从 ConfigMap 或 gRPC 获取公共故障信息。
- 汇总以上故障信息，供集群调度上层服务调用。
- 与训练容器内部建立连接，控制训练进程进行重计算动作。
- 与带外服务交互，传输任务信息。

### 组件上下游依赖

1. 从各个计算节点的 Ascend Device Plugin 中获取芯片的信息。
2. 从各个计算节点的 NodeD 中获取计算节点的 CPU、内存和硬盘的健康状态信息、节点 DPC 共享存储故障信息和灵衢网络故障信息。
3. 从 ConfigMap 或 gRPC 获取公共故障信息。
4. 汇总整个集群的资源信息，上报给 Ascend-volcano-plugin。
5. 侦听集群的任务信息，将任务状态、资源使用情况等信息上报给 CCAE。
6. 与容器内进程交互，控制训练进程进行重计算。

---

## 支持的 Tags 及 Dockerfile 链接

### Tag 规范

Tag 遵循以下格式：

```shell
<版本>
```

| 字段   | 示例值        | 说明                       |
|------|------------|--------------------------|
| `版本`   | `v26.1.0`     | ClusterD组件版本   |
| `操作系统` | `ubuntu22.04` | ClusterD镜像操作系统 |


### ClusterD 26.1.0

| Tag | Dockerfile                                      | 镜像内容                  |
| --- |-------------------------------------------------|-----------------------|
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | ClusterD组件v26.1.0版本操作系统为ubuntu22.04的镜像    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | ClusterD组件v26.1.0版本操作系统为openeuler24.03的镜像 |

---

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称 | 支持的版本 | 安装位置 | 说明 |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x（推荐使用1.19.x及以上版本） | 所有节点 | 了解 K8s 的使用请参见 [Kubernetes 文档](https://kubernetes.io/zh-cn/docs/) |
| Ascend Device Plugin | 与 ClusterD 同版本 | 计算节点 | ClusterD 依赖 Ascend Device Plugin 上报芯片信息 |
| NodeD | 与 ClusterD 同版本 | 计算节点 | ClusterD 依赖 NodeD 上报节点故障信息 |

#### 硬件规格要求

| 名称 | 100节点以内 | 500节点 | 1000节点 |
| -- | -- | -- | -- |
| CPU | 1核 | 2核 | 4核 |
| 内存 | 1GB | 2GB | 8GB |

### 如何本地构建

```bash
docker build --no-cache -t ascend-k8sclusterd:{tag} ./ -f Dockerfile.{os}
```

### 部署 ClusterD

1. 拉取镜像

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sclusterd:{tag}
```

2. 修改镜像标签

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sclusterd:{tag} ascend-k8sclusterd:{tag}
```

3. 启动 ClusterD

将 clusterd-{version}.yaml 文件中镜像的 `{tag}` 替换为实际标签。

```bash
kubectl apply -f clusterd-{version}.yaml
```

4. 验证部署

```bash
kubectl get pods -A | grep clusterd
```

---

## 支持的硬件

当前支持的昇腾硬件型号说明，请参考官方文档：
[支持的产品形态和OS清单](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/document/detail/zh/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
