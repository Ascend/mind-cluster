# 集群调度组件 NodeD

> [English](./OVERVIEW.md) | 中文

## 快速参考

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.1.0
```

---

## NodeD

NodeD 是 MindCluster 集群调度组件之一，部署在计算节点上，用于检测节点的异常状态，从 IPMI 获取计算节点的 CPU、内存、硬盘的故障信息，并上报给 ClusterD。

### 应用场景

节点的 CPU、内存或硬盘发生某些故障后，训练任务会失败。为了让训练任务在节点故障情况下快速退出，并且后续的新任务不再调度到故障节点上，MindCluster 提供了 NodeD 组件，用于检测节点的异常。

### 组件功能

- 从 IPMI 中获取节点异常，并上报给资源调度的上层服务。
- 定时发送节点故障信息给资源调度的上层服务。

### 组件上下游依赖

1. 从 IPMI 中获取计算节点的 CPU、内存、硬盘的故障信息。
2. 将计算节点的 CPU、内存、硬盘的故障信息上报给 ClusterD。

---

## 支持的 Tags 及 Dockerfile 链接

### Tag 规范

Tag 遵循以下格式：

```shell
<版本>
```

| 字段   | 示例值       | 说明                       |
|------|-----------|--------------------------|
| `版本`   | `v26.1.0`     | NodeD组件版本   |
| `操作系统` | `ubuntu22.04` | NodeD镜像操作系统 |

### NodeD 26.1.0

| Tag       | Dockerfile | 镜像内容               |
|-----------| ----------- |--------------------|
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | NodeD组件v26.1.0版本操作系统为ubuntu22.04的镜像    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | NodeD组件v26.1.0版本操作系统为openeuler24.03的镜像 |

---

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称 | 支持的版本 | 安装位置 | 说明 |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x（推荐使用1.19.x及以上版本） | 所有节点 | 了解 K8s 的使用请参见 [Kubernetes 文档](https://kubernetes.io/zh-cn/docs/) |
| ClusterD | 与 NodeD 同版本 | 管理节点 | NodeD 上报的故障信息由 ClusterD 汇总处理 |

#### 硬件规格要求

| 名称 | 要求 |
| -- | -- |
| CPU | 0.5核 |
| 内存 | 0.3GB |

### 如何本地构建

```bash
docker build --no-cache -t noded:{tag} ./ -f Dockerfile.{os}
```

### 部署 NodeD

1. 拉取镜像

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:{version}
```

2. 修改镜像标签

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:{version} noded:{version}
```

3. 启动 NodeD

将 noded-{version}.yaml 文件中镜像的 `{tag}` 替换为实际标签。

```bash
kubectl apply -f noded-{version}.yaml
```

4. 验证部署

```bash
kubectl get pods -n kube-system | grep noded
```

---

## 支持的硬件

所有昇腾设备通用

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/document/detail/zh/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
