# 集群调度组件 NPU Exporter

> [English](./OVERVIEW.md) | 中文

## 快速参考

- NPU Exporter 由 [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)  维护
- 从哪里获取帮助
    - [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster 昇腾社区](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [问题反馈](https://gitcode.com/Ascend/mind-cluster/issues)

---

## NPU Exporter

NPU Exporter 是 MindCluster 集群调度组件之一，部署在计算节点上，用于上报芯片的各项数据信息，支持 Prometheus 和 Telegraf 两种监控集成方式。

### 应用场景

在任务运行过程中，除芯片故障外，往往需要关注芯片的网络和算力使用情况，以便确认任务运行过程中的性能瓶颈，找到提升任务性能的方向。MindCluster 提供了部署在计算节点的 NPU Exporter 组件，用于上报芯片的各项数据信息。

### 组件功能

- 从驱动中获取芯片、网络的各项数据信息。
- 适配 Prometheus 钩子函数，提供标准的接口供 Prometheus 服务调用。
- 适配 Telegraf 钩子函数，提供标准的接口供 Telegraf 服务调用。
- 支持对昇腾 AI 处理器利用率、温度、电压、内存等数据信息的实时监测。
- 支持对虚拟 NPU（vNPU）的 AI Core 利用率、vNPU 总内存和 vNPU 使用中内存进行监测。
- 支持自定义指标开发，用户可参考提供的 demo 开发自定义指标插件。

### 组件上下游依赖

1. 从驱动中获取芯片以及网络信息，并放入本地缓存。
2. 从 K8s 标准化接口 CRI 中获取容器信息，并放入本地缓存。
3. 实现 Prometheus 或者 Telegraf 的接口，供二者周期性获取缓存中的数据信息。

---

## 支持的 Tags 及 Dockerfile 链接

### Tag 规范

Tag 遵循以下格式：

```shell
<版本>
```

| 字段   | 示例值                 | 说明                       |
|------|---------------------|--------------------------|
| `版本`   | `v26.1.0`     | NPU Exporter组件版本   |
| `操作系统` | `ubuntu22.04` | NPU Exporter镜像操作系统 |

### NPU Exporter 26.1.0

| Tag       | Dockerfile | 镜像内容                      |
|-----------| ----------- |---------------------------|
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | NPU Exporter组件v26.1.0版本操作系统为ubuntu22.04的镜像    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | NPU Exporter组件v26.1.0版本操作系统为openeuler24.03的镜像 |

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称 | 支持的版本 | 安装位置 | 说明 |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x（推荐使用1.19.x及以上版本） | 所有节点 | 了解 K8s 的使用请参见 [Kubernetes 文档](https://kubernetes.io/zh-cn/docs/) |
| Prometheus | 建议使用最新稳定版本 | 监控节点 | NPU Exporter 适配 Prometheus 钩子函数提供监控数据 |
| 昇腾AI处理器驱动和固件 | 请参见版本配套表 | 计算节点 | 请参见《CANN 软件安装指南》中的"安装NPU驱动和固件"章节 |

#### 硬件规格要求

| 名称 | 要求 |
| -- | -- |
| CPU | 1核 |
| 内存 | 1GB |

### 如何本地构建

```bash
docker build --no-cache -t npu-exporter:{tag} ./ -f Dockerfile.{os}
```

### 部署 NPU Exporter

1. 拉取镜像

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:{tag}
```

2. 修改镜像标签

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:{tag} npu-exporter:{tag}
```

3. 启动 NPU Exporter

将 npu-exporter-{version}.yaml 文件中镜像的 `{tag}` 替换为实际标签。

```bash
kubectl apply -f npu-exporter-{version}.yaml
```

4. 验证部署

```bash
kubectl get pods -A | grep npu-exporter
```

5. 访问监控指标

```bash
curl http://<pod-ip>:8082/metrics
```

---

## 支持的硬件

## 支持的硬件

当前支持的昇腾硬件型号说明，请参考官方文档：
[支持的产品形态和OS清单](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/document/detail/zh/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
