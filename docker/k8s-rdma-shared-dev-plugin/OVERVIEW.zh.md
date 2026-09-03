# K8s RDMA 共享设备插件

> [English](./OVERVIEW.md) | 中文

## 快速参考

- K8s RDMA 共享设备插件由 [MindCluster 代码仓库](https://gitcode.com/Ascend/mind-cluster) 维护
- 获取帮助
    - [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster 昇腾社区](https://www.hiascend.com/document/detail/zh/mindcluster/latest/clustersched/dlug/docs/zh/scheduling/01_introduction/00_overview.md)
    - [问题反馈](https://gitcode.com/Ascend/mind-cluster/issues)

---

## K8s RDMA 共享设备插件

K8s RDMA 共享设备插件是一个 Kubernetes 设备插件，用于以共享方式管理 RDMA 设备。它允许容器共享 RDMA 设备，为分布式应用提供高性能网络。

### 应用场景

在运行需要 RDMA（远程直接内存访问）的分布式训练或高性能计算工作负载时，K8s RDMA 共享设备插件允许多个容器高效共享 RDMA 设备。

### 功能特性

- 管理 Kubernetes 节点上的 RDMA 设备，支持 UB 设备类型
- 支持多个容器之间的设备共享
- 支持基于总线（buses）、供应商、设备 ID、驱动程序和接口名称的设备选择
- 支持共享与独占两种工作模式，独占模式下按 NPU 与 DPU 的映射关系分配设备
- 与 Kubernetes 设备插件框架集成
- 支持 UB 设备故障检测，并将故障信息写入 ConfigMap
- 向节点写入 DPU 资源注解，向 Pod 写入设备状态注解

### 上下游依赖

1. 检测计算节点上的 RDMA 设备，并执行周期性故障检测
2. 向 Kubernetes kubelet 设备插件框架注册
3. 向 Kubernetes 调度器报告设备可用性
4. 以 ConfigMap 形式向 Kubernetes 写入故障检测信息
5. 向节点写入 DPU 资源注解 `huawei.com/dpu.resource.name`
6. 独占模式下按 NPU 与 DPU 的映射关系为 Pod 分配设备，并将结果写入 `k8s.v1.cni.cncf.io/device-status` 注解供 Multus CNI 使用

---

## 支持的 Tags 及 Dockerfile 链接

### 标签约定

标签遵循以下格式：

```shell
<版本>-<操作系统>
```

| 字段     | 示例值           | 说明                      |
|--------|---------------|-------------------------|
| `版本`   | `v26.1.0`     | K8s RDMA 共享设备插件组件版本号    |
| `操作系统` | `ubuntu22.04` | K8s RDMA 共享设备插件组件镜像操作系统 |

### K8s RDMA 共享设备插件 26.1.0

| Tag                      | Dockerfile                                                                                                                                 | 镜像内容                                           |
|--------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------|
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/k8s-rdma-shared-dev-plugin/v26.1.0/Dockerfile.ubuntu)       | K8s RDMA 共享设备插件 v26.1.0 (基础镜像 Ubuntu 22.04)    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/k8s-rdma-shared-dev-plugin/v26.1.0/Dockerfile.openeuler) | K8s RDMA 共享设备插件 v26.1.0 (基础镜像 openEuler 24.03) |

---

## 快速入门

### 前提条件

#### 软件依赖

| 软件         | 支持版本                           | 安装位置 | 描述                                              |
|------------|--------------------------------|------|-------------------------------------------------|
| Kubernetes | 1.17.x~1.34.x（建议 1.19.x 或更高版本） | 所有节点 | 参见 [Kubernetes 文档](https://kubernetes.io/docs/) |
| RDMA 驱动    | OFED 5.6 或更高版本                 | 计算节点 | RDMA 设备驱动                                       |

#### 硬件要求

| 资源  | 要求     |
|-----|--------|
| CPU | 0.1 核  |
| 内存  | 0.1 GB |

### 本地构建

示例场景：构建 linux-aarch64 架构、v26.1.0 版本、基于 Ubuntu 22.04 的 K8s RDMA 共享设备插件镜像。

1. 获取对应架构的 Dockerfile

   前往支持的 Tags 及 Dockerfile 链接章节（如 [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/k8s-rdma-shared-dev-plugin/v26.1.0/Dockerfile.ubuntu)），打开目标版本对应的 Dockerfile.ubuntu 链接，保存文件至 aarch64 架构环境的本地目录。

2. 本地构建 Docker 镜像（禁用缓存，保证构建纯净度）

   ```bash
   docker build --no-cache -t k8s-rdma-shared-dev-plugin:v26.1.0 ./ -f Dockerfile.ubuntu
   ```

> [!NOTE]
>
>若 Docker 版本低于 18.09，或未手动开启 BuildKit，构建镜像时将无法读取 TARGETPLATFORM 变量，会造成镜像构建失败。
>
> 1. TARGETPLATFORM 为 Docker BuildKit 内置全局变量，用于识别当前构建目标平台，示例：linux/amd64、linux/arm64。
> 2. 该变量仅在 BuildKit 启用后自动注入；老旧 Docker 环境、默认关闭 BuildKit 的环境无法使用此参数。
> 3. 构建前可执行以下命令临时开启 BuildKit：
>
> ```bash
> export DOCKER_BUILDKIT=1
> ```

### 部署 K8s RDMA 共享设备插件

1. 拉取镜像

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/k8s-rdma-shared-dev-plugin:{tag}
   ```

2. 重新打标签

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/k8s-rdma-shared-dev-plugin:{tag} k8s-rdma-shared-dev-plugin:{version}
   ```

3. 创建配置文件

   创建包含设备插件配置的 ConfigMap。

4. 使用 DaemonSet 部署

   ```bash
   kubectl apply -f k8s-rdma-shared-dev-plugin-{version}.yaml
   ```

5. 验证部署

   ```bash
   kubectl get pods -A | grep k8s-rdma-shared-dev-plugin
   ```

---

## 配置说明

### 启动参数

K8s RDMA 共享设备插件支持以下启动参数：

| 参数                | 描述                                              | 默认值                                                              |
|-------------------|-------------------------------------------------|------------------------------------------------------------------|
| `--config-file`   | 配置文件路径                                          | `/k8s-rdma-shared-dev-plugin/config.json`                        |
| `--use-cdi`       | 使用 CDI 将设备暴露到容器中（UB 设备不支持 CDI）                  | `false`                                                           |
| `--ub-excl-mode`  | 开启 UB 设备独占模式（默认共享模式）                            | `false`                                                           |
| `--logLevel`      | 日志级别，`-1` 调试、`0` 信息、`1` 警告、`2` 错误、`3` 严重     | `0`                                                               |
| `--maxBackups`    | 备份日志文件的最大数量，范围 `(0, 180]`                       | `3`                                                               |
| `--maxAge`        | 备份日志文件的保留天数，范围 `[7, 700]`                      | `7`                                                               |
| `--logFile`       | 日志文件路径，文件大小超过 20MB 时自动轮转                         | `/var/log/mindx-dl/k8s-rdma-shared-dp/k8s-rdma-shared-dp.log`    |
| `--version`/`-v`  | 显示应用版本信息                                        | -                                                                |

### 配置文件参数

K8s RDMA 共享设备插件支持以下配置参数：

| 参数                       | 类型     | 描述                                  | 默认值                  |
|--------------------------|--------|-------------------------------------|----------------------|
| `periodicUpdateInterval` | int    | 定期设备更新间隔（秒），未配置时默认 60 秒，设置为 0 时禁用 | 60                   |
| `faultDetectPeriod`      | int    | 定期故障检测间隔（秒），未配置时禁用故障检测，配置时最小为 1   | 0（禁用）               |
| `configList`             | array  | 设备配置列表                              | []                   |
| `resourceName`           | string | 设备插件的资源名称                           | rdma                 |
| `resourcePrefix`         | string | 资源前缀                                | rdma-ub |
| `rdmaHcaMax`             | int    | RDMA HCA 设备的最大数量                    | 1000                 |
| `devices`                | array  | 要包含的设备名称列表（已废弃，推荐使用 `selectors.ifNames`） | []                   |
| `selectors.buses`        | array  | 用于过滤设备的总线类型（例如，`ub` 用于启用 UB 设备）      | []                   |
| `selectors.vendors`      | array  | 用于过滤设备的供应商 ID                       | []                   |
| `selectors.deviceIDs`    | array  | 用于过滤设备的设备 ID                        | []                   |
| `selectors.drivers`      | array  | 用于过滤设备的驱动程序名称                       | []                   |
| `selectors.ifNames`      | array  | 用于过滤设备的接口名称                         | []                   |
| `selectors.linkTypes`    | array  | 用于过滤设备的链路类型                         | []                   |

配置示例：

```json
{
  "periodicUpdateInterval": 300,
  "faultDetectPeriod": 5,
  "configList": [
    {
      "resourceName": "rdma-ub-devices",
      "rdmaHcaMax": 1000,
      "selectors": {
        "buses": ["ub"]
      }
    }
  ]
}
```

### 工作模式

- **共享模式（默认）**：集群内所有 Pod 共享节点上的全部 RDMA 设备。共享模式下根据 `rdmaHcaMax` 参数创建虚拟设备并上报给 kubelet，Pod 请求该资源后即可使用任一设备。
- **独占模式**：每个 NPU 与特定的 DPU 设备一一对应，通过 NPU-NIC 映射关系完成设备分配。独占模式下上报节点上真实发现的 UB 设备，Pod 申请资源时，组件根据 Pod 申请的 NPU，通过映射文件（`/etc/rdma-plugin/npu-nic-mapping.json`）查找对应的 DPU 设备并挂载到容器，实现设备级隔离。
  - **使能方式**：启动组件时添加 `--ub-excl-mode` 参数开启独占模式。
  - **分配结果**：独占模式下组件将分配的 DPU 设备写入 Pod 的 `k8s.v1.cni.cncf.io/device-status` 注解，Multus CNI 通过该注解获取设备并传递给 UB Host Device CNI 完成设备挂载。

---

## 支持的硬件

支持 UB 类型的 RDMA 网卡

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的 [许可证信息](https://www.hiascend.com/zh/legal/softlicense)。

与所有容器镜像一样，预安装的软件包（Python、系统库等）可能受其各自许可协议的约束。
