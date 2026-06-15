# 集群调度组件 Infer Operator

> [English](./OVERVIEW.md) | 中文

## 快速参考

- Infer Operator 由 [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)  维护
- 从哪里获取帮助
    - [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster 昇腾社区](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [问题反馈](https://gitcode.com/Ascend/mind-cluster/issues)

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
| `版本`   | `v26.0.0`     | Infer Operator组件版本   |

### Infer Operator 26.1.0

| Tag | Dockerfile | 镜像内容                                        |
| --- | ----------- |---------------------------------------------|
| `v26.1.0`    | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/v26.0.0/component/infer-operator/build/Dockerfile) | Infer Operator v26.1.0(基础操作系统Ubuntu 22.04) |

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

### 在线获取 Infer Operator 镜像

1. 拉取官方镜像

拉取昇腾镜像仓库提供的 Infer Operator 镜像，替换 {tag} 为实际版本号（推荐 v26.0.0）。
```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:{tag}
```

2. 修改镜像标签

为拉取的官方镜像重新打本地标签，统一本地镜像命名规范，方便后续运维管理。
```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:{tag} infer-operator:{tag}
```

### 本地构建（可选）

以下以 linux-aarch64 架构、v26.0.0 版本为例，提供完整的本地镜像构建步骤:

1. 下载官方发布的组件安装包

```shell
wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-infer-operator_26.0.0_linux-aarch64.zip
```

2. 解压安装包至自定义目录

```shell
unzip Ascend-mindxdl-infer-operator_26.0.0_linux-aarch64.zip -d Ascend-mindxdl-infer-operator_26.0.0_linux-aarch64
```

3. 进入解压后的工作目录

```shell
cd Ascend-mindxdl-infer-operator_26.0.0_linux-aarch64
```

4. 本地构建 Docker 镜像（禁用缓存，保证构建纯净度）
```bash
docker build --no-cache -t infer-operator:v26.0.0 ./ -f Dockerfile
```

### 部署 Infer Operator

1. 启动 Infer Operator

将 infer-operator-{version}.yaml 文件中镜像的 `{tag}` 替换为实际标签。

```bash
kubectl apply -f infer-operator-{version}.yaml
```

2. 验证部署

```bash
kubectl get pods -A | grep infer-operator
```

预期结果：对应命名空间下的 infer-operator 相关 Pod 状态为 Running。

---

## 支持的硬件

当前支持的昇腾硬件型号说明，请参考官方文档：
[支持的产品形态和OS清单](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/zh/legal/softlicense)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
