# 集群调度组件 Infer Operator

> [English](./OVERVIEW.md) | 中文

## 快速参考

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:v26.1.0-ubuntu22.04
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:v26.1.0-openeuler24.03
```

---

## Infer Operator

Infer Operator 是 MindCluster 集群调度组件之一，部署在管理节点上，是一个 Kubernetes Operator，用于部署和管理多角色合作的推理任务。Infer Operator 定义了 InferServiceSet、InferService 和 InstanceSet 三种 CRD，并实现了三种资源的控制器用于调谐三种资源实例状态。

### 应用场景

MindCluster 提供 Infer Operator 组件，根据推理服务的实例配置，拉起推理服务，并支持推理实例的手动扩缩容。

### 组件功能

- 创建推理实例 Workload 与 Service。
- 推理实例的手动扩缩容。

### 组件上下游依赖

1. 基于用户配置的任务 YAML 创建推理实例 Workload。
2. Workload Controller 创建 Pod 后，Volcano 进行资源的最终选定。
3. 若 Workload 申请占用 NPU 卡，Ascend Device Plugin 获取 NPU 信息，完成设备的挂载。

---

## 支持的 Tags 及 Dockerfile 链接

### Tag 规范

Tag 遵循以下格式：

```shell
<版本>
```

| 字段   | 示例值               | 说明                       |
|------|-------------------|--------------------------|
| `版本`   | `v26.1.0`     | Infer Operator组件版本   |
| `操作系统` | `ubuntu22.04` | Infer Operator镜像操作系统 |

### Infer Operator 26.1.0

| Tag | Dockerfile | 镜像内容 |
| --- | ----------- | -------- |
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | Infer Operator组件v26.1.0版本操作系统为ubuntu22.04的镜像    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | Infer Operator组件v26.1.0版本操作系统为openeuler24.03的镜像 |

---

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称 | 支持的版本 | 安装位置 | 说明 |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x（推荐使用1.19.x及以上版本） | 所有节点 | 了解 K8s 的使用请参见 [Kubernetes 文档](https://kubernetes.io/zh-cn/docs/) |
| Volcano | 请参见 [Volcano 官网中对应的 Kubernetes 版本](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility) | 管理节点 | Infer Operator 依赖 Volcano 进行资源调度 |
| Ascend Device Plugin | 与 Infer Operator 同版本 | 计算节点 | 推理任务占用 NPU 时需要 |

#### 硬件规格要求

| 名称 | 要求 |
| -- | -- |
| CPU | 2核 |
| 内存 | 2GB |

### 如何本地构建

```bash
docker build --no-cache -t infer-operator:{tag} ./ -f Dockerfile.{os}
```

### 部署 Infer Operator

1. 拉取镜像

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:{tag}
```

2. 修改镜像标签

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:{tag} infer-operator:{tag}
```

3. 启动 Infer Operator

将 infer-operator-{version}.yaml 文件中镜像的 `{tag}` 替换为实际标签。

```bash
kubectl apply -f infer-operator-{version}.yaml
```

4. 验证部署

```bash
kubectl get pods -A | grep infer-operator
```

---

## 支持的硬件

当前支持的昇腾硬件型号说明，请参考官方文档：
[支持的产品形态和OS清单](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/document/detail/zh/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
