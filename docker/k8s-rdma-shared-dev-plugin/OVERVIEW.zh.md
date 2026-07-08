# K8s RDMA 共享设备插件

> [English](./OVERVIEW.md) | 中文

## 快速参考

- K8s RDMA 共享设备插件由 [MindCluster 代码仓库](https://gitcode.com/Ascend/mind-cluster) 维护
- 获取帮助
    - [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster 昇腾社区](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [问题反馈](https://gitcode.com/Ascend/mind-cluster/issues)

---

## K8s RDMA 共享设备插件

K8s RDMA 共享设备插件是一个 Kubernetes 设备插件，用于以共享方式管理 RDMA 设备。它允许容器共享 RDMA 设备，为分布式应用提供高性能网络。

### 应用场景

在运行需要 RDMA（远程直接内存访问）的分布式训练或高性能计算工作负载时，K8s RDMA 共享设备插件允许多个容器高效共享 RDMA 设备。

### 功能特性

- 管理 Kubernetes 节点上的 RDMA 设备
- 支持多个容器之间的设备共享
- 支持基于供应商、设备 ID、驱动程序和接口名称的设备选择
- 与 Kubernetes 设备插件框架集成
- 支持容器设备接口（CDI）
- 支持 UB 设备故障检测

### 上下游依赖

1. 检测计算节点上的 RDMA 设备，并执行周期性故障检测
2. 向 Kubernetes kubelet 设备插件框架注册
3. 向 Kubernetes 调度器报告设备可用性
4. 以 configMap 形式向 Kubernetes 写入故障检测信息

---

## 支持的标签和 Dockerfile 链接

### 标签约定

标签遵循以下格式：

```shell
<version>-<os>
```

| 字段 | 示例 | 描述 |
| -- | -- | -- |
| `version` | `v26.1.0` | K8s RDMA 共享设备插件组件版本 |
| `os` | `ubuntu22.04` | 镜像操作系统 |

### K8s RDMA 共享设备插件 26.1.0

| 标签 | Dockerfile | 镜像内容 |
| --- | ----------- | -------- |
| `v26.1.0-ubuntu22.04` | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | K8s RDMA 共享设备插件 v26.1.0 Ubuntu 22.04 镜像 |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | K8s RDMA 共享设备插件 v26.1.0 openEuler 24.03 镜像 |

---

## 快速入门

### 前提条件

#### 软件依赖

| 软件 | 支持版本 | 安装位置 | 描述 |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x（建议 1.19.x 或更高版本） | 所有节点 | 参见 [Kubernetes 文档](https://kubernetes.io/docs/) |
| RDMA 驱动 | OFED 5.6 或更高版本 | 计算节点 | RDMA 设备驱动 |

#### 硬件要求

| 资源 | 要求 |
| -- | -- |
| CPU | 0.1 核 |
| 内存 | 0.1 GB |

### 本地构建

```bash
docker build --no-cache -t k8s-rdma-shared-dev-plugin:{tag} ./ -f Dockerfile.{os}
```

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

K8s RDMA 共享设备插件支持以下配置参数：

| 参数                       | 类型 | 描述                              | 默认值        |
|--------------------------| -- |---------------------------------|------------|
| `periodicUpdateInterval` | int | 定期设备更新间隔（秒）                     | 0（禁用）      |
| `faultDetectPeriod`      | int | 定期故障检测间隔（秒）                     | 5（最小配置为1）  |
| `configList`             | array | 设备配置列表                          | []         |
| `resourceName`           | string | 设备插件的资源名称                       | rdma       |
| `resourcePrefix`         | string | 资源前缀                            | huawei.com |
| `rdmaHcaMax`             | int | RDMA HCA 设备的最大数量                | 1000       |
| `devices`                | array | 要包含的设备名称列表                      | []         |
| `selectors.buses`        | array | 用于过滤设备的总线类型（例如，"ub" 用于启用 UB 设备） | []         |
| `selectors.vendors`      | array | 用于过滤设备的供应商 ID                   | []         |
| `selectors.deviceIDs`    | array | 用于过滤设备的设备 ID                    | []         |
| `selectors.drivers`      | array | 用于过滤设备的驱动程序名称                   | []         |
| `selectors.ifNames`      | array | 用于过滤设备的接口名称                     | []         |
| `selectors.linkTypes`    | array | 用于过滤设备的链路类型                     | []         |

---

## 支持的硬件

PCI和UB类型的DPU网卡

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的 [许可证信息](https://www.hiascend.com/zh/legal/softlicense)。

与所有容器镜像一样，预安装的软件包（Python、系统库等）可能受其各自许可协议的约束。
