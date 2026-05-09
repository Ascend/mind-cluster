# 集群调度组件 Ascend Operator

> [English](./OVERVIEW.md) | 中文

## 快速参考

- Ascend Operator 由 [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)  维护
- 从哪里获取帮助
    - [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster 昇腾社区](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [问题反馈](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Ascend Operator

Ascend Operator 是 MindCluster 集群调度组件之一，部署在管理节点上，支持 MindSpore、PyTorch、TensorFlow 三个 AI 框架在 Kubernetes 上进行分布式训练。CRD（Custom Resource Definition）中定义了 AscendJob 任务，用户只需配置 YAML 文件，即可轻松实现分布式训练。

### 应用场景

MindCluster 提供 Ascend Operator 组件，输入集合通信所需的主进程 IP、静态组网集合通信所需的 RankTable 信息、当前 Pod 的 rankId 等信息。

### 组件功能

- 创建 Pod，并将集合通信参数按照环境变量的方式注入。
- 创建 RankTable 文件，并按照共享存储或 ConfigMap 的方式挂载到容器，优化集合通信建链性能。

### 组件上下游依赖

1. 通过 Volcano 感知当前任务所需资源是否满足。
2. 资源满足后，针对任务创建对应的 Pod 并注入集合通信参数的环境变量。
3. Pod 创建完成后，Volcano 进行资源的最终选定。
4. 从 Ascend Device Plugin 获取任务的芯片编号、IP、rankId 信息，汇总后生成集合通信文件。
5. 通过共享存储或 ConfigMap，将集合通信文件挂载到容器内。

---

## 支持的 Tags 及 Dockerfile 链接

### Tag 规范

Tag 遵循以下格式：

```shell
<版本>
```

| 字段   | 示例值       | 说明                       |
|------|-----------|--------------------------|
| `版本`   | `v26.1.0`     | Ascend Operator组件版本   |
| `操作系统` | `ubuntu22.04` | Ascend Operator镜像操作系统 |

### Ascend Operator 26.1.0

| Tag       | Dockerfile | 镜像内容                         |
|-----------| ----------- |------------------------------|
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | Ascend Operator组件v26.1.0版本操作系统为ubuntu22.04的镜像    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | Ascend Operator组件v26.1.0版本操作系统为openeuler24.03的镜像 |

---

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称 | 支持的版本 | 安装位置 | 说明 |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x（推荐使用1.19.x及以上版本） | 所有节点 | 了解 K8s 的使用请参见 [Kubernetes 文档](https://kubernetes.io/zh-cn/docs/) |
| Volcano | 请参见 [Volcano 官网中对应的 Kubernetes 版本](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility) | 管理节点 | Ascend Operator 依赖 Volcano 进行资源调度 |

#### 硬件规格要求

| 名称 | 要求 |
| -- | -- |
| CPU | 2核 |
| 内存 | 2.5GB |

### 如何本地构建

```bash
docker build --no-cache -t ascend-k8soperator:{tag} ./ -f Dockerfile.{os}
```

### 部署 Ascend Operator

1. 拉取镜像

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8soperator:{tag}
```

2. 修改镜像标签

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8soperator:{tag} ascend-k8soperator:{tag}
```

3. 启动 Ascend Operator

将 ascend-operator-{version}.yaml 文件中镜像的 `{tag}` 替换为实际标签。

```bash
kubectl apply -f ascend-operator-{version}.yaml
```

4. 验证部署

```bash
kubectl get pods -A | grep ascend-operator
```

---

## 支持的硬件

当前支持的昇腾硬件型号说明，请参考官方文档：
[支持的产品形态和OS清单](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/document/detail/zh/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
