# 集群调度组件 NodeD

> [English](./OVERVIEW.md) | 中文

## 快速参考

- NodeD 由 [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)  维护
- 从哪里获取帮助
    - [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster 昇腾社区](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [问题反馈](https://gitcode.com/Ascend/mind-cluster/issues)

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

自v26.1.0版本开始 Tag 遵循以下格式：

```text
<版本>-<操作系统>
```

| 字段     | 示例值           | 说明           |
|--------|---------------|--------------|
| `版本`   | `v26.1.0`     | NodeD 版本号    |
| `操作系统` | `ubuntu22.04` | NodeD 镜像操作系统 |

### NodeD 26.1.0

| Tag                      | Dockerfile                                                                                                            | 镜像内容                                 |
|--------------------------|-----------------------------------------------------------------------------------------------------------------------|--------------------------------------|
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/noded/v26.1.0/Dockerfile.ubuntu)       | NodeD v26.1.0 (基础镜像 Ubuntu 22.04)    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/noded/v26.1.0/Dockerfile.openeuler) | NodeD v26.1.0 (基础镜像 openEuler 24.03) |

---

v26.0.0及以前版本的 Tag 遵循以下格式：

```shell
<版本>
```

| 字段   | 示例值       | 说明        |
|------|-----------|-----------|
| `版本` | `v26.0.0` | NodeD 版本号 |

### NodeD 26.0.0

| Tag       | Dockerfile                                                                                          | 镜像内容                              |
|-----------|-----------------------------------------------------------------------------------------------------|-----------------------------------|
| `v26.0.0` | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/v26.0.0/component/noded/build/Dockerfile) | NodeD v26.0.0 (基础镜像 Ubuntu 22.04) |

---

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称       | 支持的版本                          | 安装位置 | 说明                                                               |
|------------|--------------------------------|------|------------------------------------------------------------------|
| Kubernetes | 1.17.x~1.34.x（推荐使用1.19.x及以上版本） | 所有节点 | 了解 K8s 的使用请参见 [Kubernetes 文档](https://kubernetes.io/zh-cn/docs/) |
| ClusterD   | 与 NodeD 同版本                    | 管理节点 | NodeD 上报的故障信息由 ClusterD 汇总处理                                     |
| UMDK软件包    | 请参见版本配套表                       | 计算节点 | 针对Atlas 850 系列硬件产品、Atlas 950 SuperPod产品，在构建镜像时需要                 |

#### 硬件规格要求

| 名称 | 要求 |
| -- | -- |
| CPU | 0.5核 |
| 内存 | 0.3GB |

### 在线获取 NodeD 镜像

1. 拉取官方镜像

   拉取昇腾镜像仓库提供的 NodeD 镜像，替换 {tag} 为实际版本号（推荐 v26.0.0）。

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:{tag}
   ```

2. 修改镜像标签

   为拉取的官方镜像重新打本地标签，统一本地镜像命名规范，方便后续运维管理。

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:{tag} noded:{tag}
   ```

### 本地构建（可选）

#### v26.1.0 及更高版本本地镜像构建流程

示例场景：构建 linux-aarch64 架构、v26.1.0 版本、基于 Ubuntu 22.04 的 NodeD 组件镜像。

1. 获取对应架构的 Dockerfile

   前往[支持的 Tags 及 Dockerfile 链接](#支持的-Tags-及-Dockerfile-链接)章节，打开目标版本对应的 Dockerfile.ubuntu
   链接，保存文件至 aarch64 架构环境的本地目录。

2. 本地构建 Docker 镜像（禁用缓存，保证构建纯净度）

   ```bash
   docker build --no-cache -t noded:v26.1.0 ./ -f Dockerfile.ubuntu
   ```

> **重要注意事项**
> 若 Docker 版本低于 18.09，或未手动开启 BuildKit，构建镜像时将无法读取 TARGETPLATFORM 变量，会造成镜像构建失败。
> 1. TARGETPLATFORM 为 Docker BuildKit 内置全局变量，用于识别当前构建目标平台，示例：linux/amd64、linux/arm64。
> 2. 该变量仅在 BuildKit 启用后自动注入；老旧 Docker 环境、默认关闭 BuildKit 的环境无法使用此参数。
> 3. 构建前可执行以下命令临时开启 BuildKit：
> ```bash
> export DOCKER_BUILDKIT=1
> ```

#### v26.0.0 及更早版本本地镜像构建流程

示例场景：构建 linux-aarch64 架构、v26.0.0 版本、基于 Ubuntu 22.04 的 NodeD 组件镜像。

1. 下载官方发布的组件安装包

   ```shell
   wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-noded_26.0.0_linux-aarch64.zip
   ```

2. 解压安装包至自定义目录

   ```shell
   unzip Ascend-mindxdl-noded_26.0.0_linux-aarch64.zip -d Ascend-mindxdl-noded_26.0.0_linux-aarch64
   ```

3. 进入解压后的工作目录

   ```shell
   cd Ascend-mindxdl-noded_26.0.0_linux-aarch64
   ```

4. 本地构建 Docker 镜像（禁用缓存，保证构建纯净度）

   ```bash
   docker build --no-cache -t noded:v26.0.0 ./ -f Dockerfile
   ```

### 部署 NodeD

1. 给 Kubernetes 节点打标签

   为对应节点添加标签，用于集群调度匹配，替换 `<node-name>` 为实际节点名称。

   ```bash
   kubectl label nodes <node-name> workerselector=dls-worker-node
   ```

2. 启动 NodeD

   部署前需将 YAML 文件内的镜像 `{tag}` 替换为实际使用的镜像版本。

   ```bash
   kubectl apply -f noded-{version}.yaml
   ```

3. 验证部署

   ```bash
   kubectl get pods -A | grep noded
   ```

   预期结果：对应命名空间下的 noded 相关 Pod 状态为 Running。

---

## 支持的硬件

当前支持的昇腾硬件型号说明，请参考官方文档：
[支持的产品形态和OS清单](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/01_introduction/03_supported_product_models_and_os.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/zh/legal/softlicense)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
