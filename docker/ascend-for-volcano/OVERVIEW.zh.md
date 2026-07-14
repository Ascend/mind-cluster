# 集群调度组件 Ascend for Volcano

> [English](./OVERVIEW.md) | 中文

## 快速参考

- Ascend for Volcano 由 [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)  维护
- 从哪里获取帮助
    - [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster 昇腾社区](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [问题反馈](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Ascend for Volcano

Ascend for Volcano 是基于开源 Volcano 调度的插件机制，增加了昇腾 AI 处理器（NPU）的亲和性调度、虚拟设备调度等特性，部署在管理节点上。Volcano 是基于昇腾 AI 处理器的互联拓扑结构和处理逻辑，实现了昇腾 AI 处理器最佳利用的调度器组件，可以最大化发挥昇腾 AI 处理器计算性能。

### 应用场景

K8s 基础调度仅能通过感知昇腾芯片的数量进行资源调度。为实现亲和性调度，最大化资源利用，需要感知昇腾芯片之间的网络连接方式，选择网络最优的资源。MindCluster 提供了部署在管理节点的 Volcano 服务，针对不同的昇腾设备和组网方式提供网络亲和性调度。

### 组件功能

- **可用设备计算**：根据集群调度底层组件上报的故障信息及节点信息计算集群的可用设备信息。（`self-maintain-available-card` 默认开启。关闭时从集群调度底层组件获取集群的可用设备信息。）
- **最优资源分配**：从 K8s 的任务对象中获取用户期望的资源数量，结合集群的设备数量、设备类型和设备组网方式，选择最优资源分配给任务。
- **故障重调度**：任务资源故障时，重新调度任务。
- **NPU 亲和性调度**：基于昇腾 AI 处理器的互联拓扑结构，优先将任务调度到同一张卡内的处理器，其次调度到 HCCS 互联的处理器，最后调度到 PCIe 互联的处理器，减少资源碎片和网络拥塞。
- **交换机亲和性调度**：基于交换机下节点的组网配置和参数面网络配置，实现节点的最佳利用。支持 Spine-Leaf 双层互联、单层交换机互联等多种组网方式。
- **逻辑超节点亲和性调度**：对物理超节点根据切分策略划分出逻辑超节点，实现节点的最佳利用。
- **多级调度策略**：根据 NPU 的网络拓扑层级关系将集群资源抽象为多层级结构，支持通过 `resource-level-config` 参数配置。
- **多种调度模式**：支持整卡调度、静态 vNPU 调度、动态 vNPU 调度和软切分调度。

### 亲和性策略

针对昇腾 910 AI 处理器的特征和资源利用的规则，制定以下亲和性策略（按优先级排列）：

1. **HCCS 亲和性调度原则**：申请的昇腾 910 AI 处理器必须在同一个 HCCS 环内，优先选择剩余可用处理器数量最匹配的 HCCS。
2. **优先占满调度原则**：优先分配已经分配过昇腾 910 AI 处理器的 AI 服务器，减少碎片。
3. **剩余偶数优先原则**：优先选择满足上述条件的 HCCS，然后选择剩余处理器数量为偶数的 HCCS。

### 组件上下游依赖

1. 根据 ClusterD 上报的信息计算集群资源信息（默认使用 ClusterD 的场景）。
2. 接收第三方下发的任务拉起配置，根据集群资源信息，选择最优节点资源。
3. 向计算节点的 Ascend Device Plugin 传递具体的资源选中信息，完成设备挂载。

### 镜像说明

Ascend for Volcano 包含两个镜像：

- **volcano-scheduler**：Volcano 调度器镜像，包含昇腾 NPU 亲和性调度插件（`volcano-npu_*.so`）。
- **volcano-controller**：Volcano 控制器镜像。

---

## 支持的 Tags 及 Dockerfile 链接

### Tag 规范

自昇腾NPU调度插件v26.1.0版本开始 Tag 遵循以下格式：

```text
<组件版本>-<昇腾调度插件版本>-<操作系统>
```

| 字段         | 示例值            | 说明            |
|------------|----------------|---------------|
| `组件版本`     | `v1.7.0`       | Volcano 组件版本  |
| `昇腾调度插件版本` | `v26.1.0`      | 昇腾NPU调度插件版本   |
| `操作系统`     | `alpinelatest` | Volcano镜像操作系统 |

### Ascend for Volcano 26.1.0（Volcano v1.12.0）

| Tag                              | Dockerfile                                                                                                                                                               | 镜像内容                                                                     |
|----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| `v1.12.0-v26.1.0-alpinelatest`   | [Dockerfile-scheduler.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.12.0/v26.1.0/Dockerfile-scheduler.alpine)         | Volcano调度器v26.1.0版本镜像（含昇腾NPU调度插件，基于Volcano v1.12.0，基础镜像 Alpine latest）   |
| `v1.12.0-v26.1.0-alpinelatest`   | [Dockerfile-controller.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.12.0/v26.1.0/Dockerfile-controller.alpine)       | Volcano控制器v26.1.0版本镜像（基于Volcano v1.12.0，基础镜像 Alpine latest）              |
| `v1.12.0-v26.1.0-openeuler24.03` | [Dockerfile-scheduler.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.12.0/v26.1.0/Dockerfile-scheduler.openeuler)   | Volcano调度器v26.1.0版本镜像（含昇腾NPU调度插件，基于Volcano v1.12.0，基础镜像 openEuler 24.03） |
| `v1.12.0-v26.1.0-openeuler24.03` | [Dockerfile-controller.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.12.0/v26.1.0/Dockerfile-controller.openeuler) | Volcano控制器v26.1.0版本镜像（基于Volcano v1.12.0，基础镜像 openEuler 24.03）            |

### Ascend for Volcano 26.1.0（Volcano v1.9.0）

| Tag                             | Dockerfile                                                                                                                                                              | 镜像内容                                                                    |
|---------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| `v1.9.0-v26.1.0-alpinelatest`   | [Dockerfile-scheduler.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.9.0/v26.1.0/Dockerfile-scheduler.alpine)         | Volcano调度器v26.1.0版本镜像（含昇腾NPU调度插件，基于Volcano v1.9.0，基础镜像 Alpine latest）   |
| `v1.9.0-v26.1.0-alpinelatest`   | [Dockerfile-controller.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.9.0/v26.1.0/Dockerfile-controller.alpine)       | Volcano控制器v26.1.0版本镜像（基于Volcano v1.9.0，基础镜像 Alpine latest）              |
| `v1.9.0-v26.1.0-openeuler24.03` | [Dockerfile-scheduler.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.9.0/v26.1.0/Dockerfile-scheduler.openeuler)   | Volcano调度器v26.1.0版本镜像（含昇腾NPU调度插件，基于Volcano v1.9.0，基础镜像 openEuler 24.03） |
| `v1.9.0-v26.1.0-openeuler24.03` | [Dockerfile-controller.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.9.0/v26.1.0/Dockerfile-controller.openeuler) | Volcano控制器v26.1.0版本镜像（基于Volcano v1.9.0，基础镜像 openEuler 24.03）            |

### Ascend for Volcano 26.1.0（Volcano v1.7.0）

| Tag                             | Dockerfile                                                                                                                                                              | 镜像内容                                                                    |
|---------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| `v1.7.0-v26.1.0-alpinelatest`   | [Dockerfile-scheduler.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.7.0/v26.1.0/Dockerfile-scheduler.alpine)         | Volcano调度器v26.1.0版本镜像（含昇腾NPU调度插件，基于Volcano v1.7.0，基础镜像 Alpine latest）   |
| `v1.7.0-v26.1.0-alpinelatest`   | [Dockerfile-controller.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.7.0/v26.1.0/Dockerfile-controller.alpine)       | Volcano控制器v26.1.0版本镜像（基于Volcano v1.7.0，基础镜像 Alpine latest）              |
| `v1.7.0-v26.1.0-openeuler24.03` | [Dockerfile-scheduler.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.7.0/v26.1.0/Dockerfile-scheduler.openeuler)   | Volcano调度器v26.1.0版本镜像（含昇腾NPU调度插件，基于Volcano v1.7.0，基础镜像 openEuler 24.03） |
| `v1.7.0-v26.1.0-openeuler24.03` | [Dockerfile-controller.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.7.0/v26.1.0/Dockerfile-controller.openeuler) | Volcano控制器v26.1.0版本镜像（基于Volcano v1.7.0，基础镜像 openEuler 24.03）            |

---

昇腾NPU调度插件v26.0.0及以前版本的 Tag 遵循以下格式：

```shell
<组件版本>-<昇腾调度插件版本>
```

| 字段         | 示例值       | 说明           |
|------------|-----------|--------------|
| `组件版本`     | `v1.7.0`  | Volcano 组件版本 |
| `昇腾调度插件版本` | `v26.0.0` | 昇腾NPU调度插件版本  |

### Ascend for Volcano 26.0.0（Volcano v1.9.0）

以 linux-aarch64 架构为例： Ascend for Volcano组件安装包下载：[Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip](https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip)

| Tag              | Dockerfile(安装包内文件路径)                 | 镜像内容                                                                  |
|------------------|--------------------------------------|-----------------------------------------------------------------------|
| `v1.9.0-v26.0.0` | volcano-v1.9.0/Dockerfile-scheduler  | Volcano调度器v26.0.0版本镜像（含昇腾NPU调度插件，基于Volcano v1.9.0，基础镜像 Alpine latest） |
| `v1.9.0-v26.0.0` | volcano-v1.9.0/Dockerfile-controller | Volcano控制器v26.0.0版本镜像（基于Volcano v1.9.0，基础镜像 Alpine latest）            |

### Ascend for Volcano 26.0.0（Volcano v1.7.0）

以 linux-aarch64 架构为例： Ascend for Volcano组件安装包下载：[Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip](https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip)

| Tag              | Dockerfile(安装包内文件路径)                 | 镜像内容                                                                  |
|------------------|--------------------------------------|-----------------------------------------------------------------------|
| `v1.7.0-v26.0.0` | volcano-v1.7.0/Dockerfile-scheduler  | Volcano调度器v26.0.0版本镜像（含昇腾NPU调度插件，基于Volcano v1.7.0，基础镜像 Alpine latest） |
| `v1.7.0-v26.0.0` | volcano-v1.7.0/Dockerfile-controller | Volcano控制器v26.0.0版本镜像（基于Volcano v1.7.0，基础镜像 Alpine latest）            |

---

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称 | 支持的版本 | 安装位置 | 说明 |
| -- | -- | -- | -- |
| Kubernetes | 1.19.x~1.34.x | 所有节点 | 了解 K8s 的使用请参见 [Kubernetes 文档](https://kubernetes.io/zh-cn/docs/) |
| Ascend Device Plugin | 与 Volcano 同版本 | 计算节点 | Volcano 依赖 Ascend Device Plugin 上报设备信息 |
| ClusterD | 与 Volcano 同版本 | 管理节点 | Volcano 依赖 ClusterD 汇总集群故障信息 |

#### 硬件规格要求

| 名称 | 100节点以内 | 500节点 | 1000节点 |
| -- | -- | -- | -- |
| Volcano Scheduler CPU | 2.5核 | 4核 | 5.5核 |
| Volcano Scheduler 内存 | 2.5GB | 5GB | 8GB |
| Volcano Controller CPU | 2核 | 2核 | 2.5核 |
| Volcano Controller 内存 | 2.5GB | 3GB | 4GB |

### 在线获取 Ascend for Volcano 镜像

1. 拉取镜像

   拉取昇腾镜像仓库提供的 Ascend for Volcano 相关镜像，替换 {tag} 为实际版本对应的Tag。

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:{tag}
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:{tag}
   ```

2. 修改镜像标签

   为拉取的官方镜像重新打本地标签，统一本地镜像命名规范，方便后续运维管理。

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:{tag} volcanosh/vc-scheduler:{tag}
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:{tag} volcanosh/vc-controller-manager:{tag}
   ```

### 本地构建(可选)

#### 昇腾 NPU 调度插件 v26.1.0 及更高版本本地镜像构建流程

示例场景：构建基于 Alpine latest、架构为 linux-aarch64 的 Volcano v1.9.0 组件镜像，镜像内置昇腾 NPU 调度插件 v26.1.0。

1. 获取对应架构的 Dockerfile

   前往[支持的 Tags 及 Dockerfile 链接](#支持的-Tags-及-Dockerfile-链接)章节，打开目标版本对应的
   Dockerfile-scheduler.alpine 和 Dockerfile-controller.alpine 链接，保存文件至 aarch64 架构环境的本地目录。

2. 本地构建 Docker 镜像（禁用缓存，保证构建纯净度）

   ```bash
   # 构建调度器镜像
   docker build --no-cache -t  volcanosh/vc-scheduler:v1.9.0 ./  -f Dockerfile-scheduler.alpine

   # 构建控制器镜像
   docker build --no-cache -t volcanosh/vc-controller-manager:v1.9.0 ./  -f Dockerfile-controller.alpine
   ```

> **重要注意事项**
> 若 Docker 版本低于 18.09，或未手动开启 BuildKit，构建镜像时将无法读取 TARGETPLATFORM 变量，会造成镜像构建失败。
> 1. TARGETPLATFORM 为 Docker BuildKit 内置全局变量，用于识别当前构建目标平台，示例：linux/amd64、linux/arm64。
> 2. 该变量仅在 BuildKit 启用后自动注入；老旧 Docker 环境、默认关闭 BuildKit 的环境无法使用此参数。
> 3. 构建前可执行以下命令临时开启 BuildKit：
> ```bash
> export DOCKER_BUILDKIT=1
> ```

#### 昇腾 NPU 调度插件 v26.0.0 及更早版本本地镜像构建流程

示例场景：构建基于 Alpine latest、架构为 linux-aarch64 的 Volcano v1.9.0 组件镜像，镜像内置昇腾 NPU 调度插件 v26.0.0。

1. 下载官方发布的组件安装包

   ```shell
   wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip
   ```

2. 解压安装包至自定义目录

   ```shell
   unzip Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip -d Ascend-mindxdl-volcano_26.0.0_linux-aarch64
   ```

3. 进入解压后的工作目录

   ```shell
   cd Ascend-mindxdl-volcano_26.0.0_linux-aarch64/volcano-v1.9.0
   ```

4. 本地构建 Docker 镜像（禁用缓存，保证构建纯净度）

   ```bash
   # 构建调度器镜像
   docker build --no-cache -t  volcanosh/vc-scheduler:v1.9.0 ./  -f Dockerfile-scheduler

   # 构建控制器镜像
   docker build --no-cache -t volcanosh/vc-controller-manager:v1.9.0 ./  -f Dockerfile-controller
   ```

> **注意**：
> - TARGETPLATFORM 是 Docker BuildKit 提供的全局内置参数，用于获取当前构建的目标平台（如 linux/amd64、linux/arm64）。
> - 只有启用 BuildKit，才会自动注入这个变量。旧版 Docker / 默认关闭 BuildKit
    的环境，构建时不存在这个变量，需要在运行构建指令前通过 <b>export DOCKER_BUILDKIT=1</b> 临时启用。

### 部署 Ascend for Volcano

1. 启动 Volcano

   YAML文件名中 `{version}` 替换为实际版本（当前使用的volcano版本为 v1.9.0），部署前需将 YAML 文件内的镜像 `{tag}` 替换为实际使用的镜像版本。

   ```bash
   kubectl apply -f volcano-{version}.yaml
   ```

2. 验证部署

   ```bash
   kubectl get pods -A | grep volcano
   ```

   预期结果：对应命名空间下的 volcano 相关 Pod 状态为 Running。

---

## 支持的硬件

当前支持的昇腾硬件型号说明，请参考官方文档：
[支持的产品形态和OS清单](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/01_introduction/03_supported_product_models_and_os.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/zh/legal/softlicense)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
