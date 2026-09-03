# 集群调度组件 Ascend Dynamic Resource Allocation

> English | 中文

## 快速参考

- Ascend Dynamic Resource Allocation 由 [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster) 维护
- 从哪里获取帮助
    - [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster 昇腾社区](https://www.hiascend.com/document/detail/zh/mindcluster/latest/clustersched/dlug/docs/zh/scheduling/01_introduction/00_overview.md)
    - [问题反馈](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Ascend Dynamic Resource Allocation

Ascend Dynamic Resource Allocation（Ascend DRA）是 MindCluster 集群调度套件的核心组件之一，以 kubelet 插件方式部署在每个计算节点上，基于 Kubernetes 动态资源分配（Dynamic Resource Allocation，DRA）机制，提供 NPU 设备的发现、上报与分配能力。

### 应用场景

传统的 Kubernetes Device Plugin 机制通过 kubelet 上报扩展资源并按整数数量进行粗粒度分配，无法表达多卡训练、vNPU 切分、HCCS 拓扑感知、精细化健康过滤等场景下调度器所需的设备属性。为此 Kubernetes 引入 DRA 机制：由节点级 driver 通过 `ResourceSlice` 上报每张设备的属性，由工作负载通过 `ResourceClaim` 声明式申请设备，并由 kubelet 回调 driver 在容器启动前完成设备准备。MindCluster 提供 Ascend DRA 在昇腾 NPU 硬件上实现该机制。

> [!NOTE]
> Ascend DRA 与 Ascend Device Plugin 是两种不同的设备接入方式，互斥使用。使用 DRA 方式的节点无需部署 Ascend Device Plugin；若集群中同时存在两种接入方式的节点，请通过节点标签区分。

### 组件功能

- **设备发现与上报**：启动时通过 DCMI 枚举本节点全部 NPU，并以 `ResourceSlice` 形式上报给 API Server，
  作用域为本节点。每张设备携带稳定属性（`type=npu`、`physicId`、`chipName`），使调度器可按真实硬件属性
  过滤而非按不透明整数数量调度。

- **设备分配（Prepare）**：实现 kubelet DRA 回调 `PrepareResourceClaims`。当工作负载的 `ResourceClaim`
  绑定到本 driver 且 Pod 调度到本节点时，kubelet 调用该回调。driver 从本地清单中分配设备，将分配结果
  持久化到 checkpoint（重启后幂等），写入 CDI spec 文件，并返回完全限定的 CDI 设备 ID 给 kubelet。

- **设备释放（Unprepare）**：实现 kubelet DRA 回调 `UnprepareResourceClaims`。Pod 删除时 kubelet 调用该
  回调，driver 移除对应的 CDI spec 文件并清理本地 checkpoint 条目。

- **CDI Spec 生成**：将分配到的设备名（如 `npu-0`）转换为完全限定的 CDI 设备 ID，并在 CDI 根目录下为
  每个 claim 写入 CDI spec 文件。容器运行时随后按 CDI spec 合并并注入 `/dev/davinci<id>`、共享设备节点、
  `LD_LIBRARY_PATH` 环境变量以及驱动库挂载到业务容器。

- **多代际支持**：设备相关逻辑通过 `DraGenerationInterface` 隔离，每代芯片（Ascend 910、Ascend 950）
  独立实现枚举、属性上报和 ID 转换；新增芯片代际只需新增一代实现，无需修改公共代码。

- **健康检查**：暴露 HTTP `healthz` 端点（默认 `:11251`，路径 `/`），供 kubelet liveness 探针使用。
  服务与业务逻辑解耦，支持 HTTP 与 HTTPS，限流 1 QPS、突发上限 5。

> [!NOTE]
> DRA 机制当前仅支持基于 Kubernetes 和 Volcano 1.15 的基础整卡调度，暂不支持亲和性调度、动态虚拟化等高级调度特性。

---

## 支持的 Tags 及 Dockerfile 链接

### Tag 规范

自v26.2.0版本开始 Tag 遵循以下格式：

```text
<版本>-<操作系统>
```

| 字段     | 示例值           | 说明                                     |
|--------|---------------|----------------------------------------|
| `版本`   | `v26.2.0`     | Ascend DRA 版本号 |
| `操作系统` | `ubuntu22.04` | Ascend DRA 镜像操作系统                      |

### Ascend Dynamic Resource Allocation 26.2.0

| Tag                      | Dockerfile                                                                                                                                          | 镜像内容                                                |
|--------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------|
| `v26.2.0-ubuntu22.04`    | [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-dynamic-resource-allocation/v26.2.0/Dockerfile.ubuntu)        | Ascend DRA v26.2.0 (基础镜像 Ubuntu 22.04)            |
| `v26.2.0-openeuler24.03` | [Dockerfile.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-dynamic-resource-allocation/v26.2.0/Dockerfile.openeuler)   | Ascend DRA v26.2.0 (基础镜像 openEuler 24.03)         |

---

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称         | 支持的版本                          | 安装位置 | 说明                                                                                                              |
|--------------|--------------------------------|------|-----------------------------------------------------------------------------------------------------------------|
| Kubernetes   | 1.36.x  | 所有节点 | 了解 K8s 的使用请参见 [Kubernetes 文档](https://kubernetes.io/zh-cn/docs/) |
| Containerd   | 2.0+    | 所有节点 | 必须支持 CDI。可从 Containerd 的 [官网](https://containerd.io/downloads/) 获取                                              |
| 昇腾AI处理器驱动和固件 | 请参见版本配套表 | 计算节点 | 请参见《[CANN 软件安装指南](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/910/softwareinst/instg/instg_0005.html?OS=openEuler&InstallType=local)》中的"安装NPU驱动和固件"章节                                                                                   |
| UMDK软件包      | 请参见版本配套表 | 计算节点 | 针对Atlas 850E 超节点、Atlas 650E服务器、Atlas 950 SuperPod超节点，在构建镜像时需要，详情请参见[容器基础镜像集成UMDK安装指导](https://www.hiascend.com/document/detail/zh/mindcluster/latest/clustersched/dlug/docs/zh/scheduling/07_references/02_common_operations.html#容器基础镜像集成umdk安装指导)                                                                 |

#### 硬件规格要求

| 名称  | 要求    |
|-----|-------|
| CPU | 0.5核  |
| 内存  | 0.5GB |

#### 安装驱动

宿主机已安装驱动和固件，详情请参见《[CANN 软件安装指南](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/910/softwareinst/instg/instg_0005.html?OS=openEuler&InstallType=local)》中的"安装NPU驱动和固件"章节。

### 在线获取 Ascend DRA 镜像

1. 拉取官方镜像

   拉取昇腾镜像仓库提供的 Ascend DRA 镜像，替换 {tag} 为实际版本号（推荐最新版本）。

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-dra:{tag}
   ```

2. 修改镜像标签

   为拉取的官方镜像重新打本地标签，统一本地镜像命名规范，方便后续运维管理。

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-dra:{tag} ascend-dra:{tag}
   ```

### 本地构建（可选）

示例场景：构建 linux-aarch64 架构、v26.2.0 版本、基于 Ubuntu 22.04 的 Ascend DRA 组件镜像。

1. 获取对应架构的 Dockerfile

   前往[支持的 Tags 及 Dockerfile 链接](#支持的-tags-及-dockerfile-链接)章节，打开目标版本对应的 Dockerfile.ubuntu 链接，保存文件至 aarch64 架构环境的本地目录。

2. 本地构建 Docker 镜像（禁用缓存，保证构建纯净度）

   ```bash
   docker build --no-cache -t ascend-dra:v26.2.0 ./ -f Dockerfile.ubuntu
   ```

> [!NOTE]
> 若 Docker 版本低于 18.09，或未手动开启 BuildKit，构建镜像时将无法读取 TARGETPLATFORM 变量，会造成镜像构建失败。
>
> 1. TARGETPLATFORM 为 Docker BuildKit 内置全局变量，用于识别当前构建目标平台，示例：linux/amd64、linux/arm64。
> 2. 该变量仅在 BuildKit 启用后自动注入；老旧 Docker 环境、默认关闭 BuildKit 的环境无法使用此参数。
> 3. 构建前可执行以下命令临时开启 BuildKit：
>
> ```bash
> export DOCKER_BUILDKIT=1
> ```

### 部署 Ascend DRA

1. 部署 Ascend DRA

   部署清单为 `kube-system` 命名空间下的 DaemonSet，并附带 ServiceAccount、ClusterRole、ClusterRoleBinding、
   `DeviceClass`（`npu.huawei.com`）与 `ValidatingAdmissionPolicy`。部署前需将 `ascend-dra-driver.yaml`
   内的镜像 `{tag}` 替换为实际使用的镜像版本。

   ```bash
   kubectl apply -f ascend-dra-driver.yaml
   ```

   回显示例如下：

   ```text
   serviceaccount/ascend-dra-driver-service-account created
   clusterrole.rbac.authorization.k8s.io/ascend-dra-driver-role created
   clusterrolebinding.rbac.authorization.k8s.io/ascend-dra-driver-role-binding created
   daemonset.apps/ascend-dra-driver-kubeletplugin created
   deviceclass.resource.k8s.io/npu.huawei.com created
   validatingadmissionpolicy.admissionregistration.k8s.io/resourceslices-policy-ascend-dra-driver created
   validatingadmissionpolicybinding.admissionregistration.k8s.io/resourceslices-policy-ascend-dra-driver created
   ```

2. 验证部署

   ```bash
   kubectl get pods -n kube-system | grep ascend-dra-driver
   ```

   预期结果：每个计算节点上的 `ascend-dra-driver-kubeletplugin` Pod 状态为 `Running`。

   ```text
   NAME                                        READY   STATUS    RESTARTS   AGE
   ascend-dra-driver-kubeletplugin-5m2xv       1/1     Running   0          74s
   ```

3. 验证 ResourceSlice 上报

   ```bash
   kubectl get deviceclass npu.huawei.com
   kubectl get resourceslices
   ```

   预期结果：可查到名为 `npu.huawei.com` 的 DeviceClass，且每个计算节点存在对应的 ResourceSlice，
   `DEVICECOUNT` 列展示发现的 NPU 数量。

   ```text
   NAME            AGE
   npu.huawei.com  2m

   NAME                                  POOLNAME     DEVICECOUNT   AGE
   npu.huawei.com-node1-1a2b3c           node1        8             2m
   ```

> 关于组件上报的 `ResourceSlice` 字段格式、设备属性语义、健康检查端点行为以及组件启动参数完整列表，
> 请参考官方文档：[Ascend Dynamic Resource Allocation 接口说明](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/06_api/17_ascend_dynamic_resource_allocation_.md)
> 与[手动安装](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/05_developer_guide/00_installation_deployment/00_manual_installation/14_ascend_dynamic_resource_allocation.md)。
> 完整的 Prepare/Unprepare 工作流演示请参见
> [DRA 最佳实践](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/04_usage/13_ascend_dynamic_resource_allocation_best_practice/01_deploying_dra_and_running_task.md)。

---

## 支持的硬件

Ascend DRA 组件支持的产品形态如下：

- Atlas A2 训练系列产品
- Atlas A2 推理系列产品
- Atlas A3 训练系列产品
- Atlas A3 推理系列产品
- Ascend 950 系列产品

当前支持的昇腾硬件型号说明，请参考官方文档：
[支持的产品形态和OS清单](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/01_introduction/03_supported_product_models_and_os.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/zh/legal/softlicense)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
