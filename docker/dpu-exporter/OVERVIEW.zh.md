# 集群调度组件 DPU Exporter

> [English](./OVERVIEW.md) | 中文

## 快速参考

- DPU Exporter 由 [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)  维护
- 从哪里获取帮助
    - [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster 昇腾社区](https://www.hiascend.com/document/detail/zh/mindcluster/latest/clustersched/dlug/docs/zh/scheduling/01_introduction/00_overview.md)
    - [问题反馈](https://gitcode.com/Ascend/mind-cluster/issues)

---

## DPU Exporter

### 应用场景

在任务运行过程中，DPU的健康状态直接影响任务的稳定性。MindCluster提供DPU Exporter组件用于监测DPU的运行状态与统计指标。

### 组件功能

- 从网卡管理工具与文件接口获取DPU的运行状态与统计指标。
- 提供Prometheus指标接口，用于监控DPU的运行状态与统计指标。

### 组件上下游依赖

1. 从网卡管理工具和文件接口分别获取DPU全局指标和interface级指标。
2. 将获取到的指标转换为Prometheus指标格式。
3. 提供Prometheus指标接口，用于监控DPU的运行状态与统计指标。

---

## 支持的 Tags 及 Dockerfile 链接

### Tag 规范

Tag 遵循以下格式：

```text
<版本>-<操作系统>
```

| 字段     | 示例值           | 说明                  |
|--------|---------------|---------------------|
| `版本`   | `v26.2.0`     | DPU Exporter 版本号   |
| `操作系统` | `ubuntu22.04` | DPU Exporter 镜像操作系统 |

### DPU Exporter 26.2.0

| Tag                      | Dockerfile                                                                                                                   | 镜像内容                                        |
|--------------------------|------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------|
| `v26.2.0-ubuntu22.04`    | [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/dpu-exporter/v26.2.0/Dockerfile.ubuntu)                     | DPU Exporter v26.2.0 (基础镜像 Ubuntu 22.04)    |
| `v26.2.0-openeuler24.03` | [Dockerfile.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/dpu-exporter/v26.2.0/Dockerfile.openeuler)  | DPU Exporter v26.2.0 (基础镜像 openEuler 24.03) |

---

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称              | 支持的版本                          | 安装位置 | 说明                                                               |
|-------------------|--------------------------------|------|------------------------------------------------------------------|
| Kubernetes        | 1.17.x~1.34.x（推荐使用1.19.x及以上版本） | 所有节点 | 了解 K8s 的使用请参见 [Kubernetes 文档](https://kubernetes.io/zh-cn/docs/) |
| Prometheus        | 建议使用最新稳定版本                     | 监控节点 | DPU Exporter 适配 Prometheus 钩子函数提供监控数据                          |
| DPU驱动、固件及hinicadm5 | 请参见版本配套表                       | 计算节点 | hinicadm5 为 DPU 驱动配套交付的管理工具                                   |

#### 硬件规格要求

| 名称  | 要求   |
|-----|------|
| CPU | 1核   |
| 内存 | 512 MB |

### 在线获取 DPU Exporter 镜像

1. 拉取官方镜像

   拉取昇腾镜像仓库提供的 DPU Exporter 镜像，替换 {tag} 为实际版本号。

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/dpu-exporter:{tag}
   ```

2. 修改镜像标签

   为拉取的官方镜像重新打本地标签，统一本地镜像命名规范，方便后续运维管理。

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/dpu-exporter:{tag} dpu-exporter:{tag}
   ```

### 本地构建（可选）

示例场景：构建 linux-aarch64 架构、v26.2.0 版本、基于 Ubuntu 22.04 的 DPU Exporter 组件镜像。

1. 获取构建产物

   在源码仓的 `component/dpu-exporter/build` 目录下执行 `build.sh`，产物目录 `output` 中将生成
   `dpu-exporter` 二进制、`config.json` 以及部署 YAML 文件。

2. 获取对应架构的 Dockerfile

   前往支持的 Tags 及 Dockerfile 链接章节，打开目标版本对应的 Dockerfile 链接，将其与构建产物一同保存至
   aarch64 架构环境的本地目录。

3. 本地构建 Docker 镜像（禁用缓存，保证构建纯净度）

   ```bash
   docker build --no-cache -t dpu-exporter:v26.2.0 ./ -f Dockerfile
   ```

### 部署 DPU Exporter

1. 给 Kubernetes 节点打标签

   为对应节点添加标签，用于集群调度匹配，替换 `<node-name>` 为实际节点名称。

   ```bash
   kubectl label nodes <node-name> workerselector=dls-worker-node
   ```

2. 启动 DPU Exporter

   部署前需将 YAML 文件内的镜像 `{tag}` 替换为实际使用的镜像版本。

   ```bash
   kubectl apply -f dpu-exporter-{version}.yaml
   ```

3. 验证部署

   ```bash
   kubectl get pods -A | grep dpu-exporter
   ```

   预期结果：对应命名空间下的 dpu-exporter 相关 Pod 状态为 Running。

4. 访问监控指标

   ```bash
   curl http://<pod-ip>:8080/metrics
   ```

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/zh/legal/softlicense)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
