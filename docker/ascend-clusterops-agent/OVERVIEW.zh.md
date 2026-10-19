# 集群调度组件 Ascend ClusterOps Agent

> [English](./OVERVIEW.md) | 中文

## 快速参考

- Ascend ClusterOps Agent 由 [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)维护
- 从哪里获取帮助
    - [MindCluster 代码仓](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster 昇腾社区](https://www.hiascend.com/document/detail/zh/mindcluster/latest/clustersched/dlug/docs/zh/scheduling/01_introduction/00_overview.md)
    - [问题反馈](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Ascend ClusterOps Agent

Ascend ClusterOps Agent（集群运维Agent）是 MindCluster 集群调度组件之一，把ascend-faultdiag（离线日志清洗+诊断）包装成集群诊断Agent。用户只需指定异常任务名，即可自动完成**采集→清洗→诊断→出报告**的全流程。

组件包含三个模块，其中Agent Core和Node Collector共用同一个Docker镜像（通过`SERVICE`环境变量区分启动入口）：

- **Agent Core**：诊断主控，部署在管理节点。维护任务Pod中心关系缓存与挂载关系表，并发向各节点下发采集指令，聚合清洗产物后调用`ascend-fd diag`生成诊断报告，并按需落缓存。
- **Node Collector**：节点采集端，部署在计算节点。按采集契约从宿主机与静态挂载的共享盘采集任务日志，本地执行`ascend-fd parse`进行清洗。
- **Kubectl插件**（`kubectl-ascend_diag`/`kubectl-clusterops`）：面向用户的入口，安装在用户机。`ascend_diag`经port-forward调用Agent Core；`clusterops`管理LLM配置。

### 应用场景

训练或推理任务在运行过程中出现故障时，需要快速定位故障根因。任务通常分布在多个节点，故障相关日志（系统日志、设备日志、业务日志、plog等）分散在各节点的宿主机上，人工排查成本高、效率低。MindCluster提供集群运维Agent特性，通过`kubectl ascend_diag`命令按任务维度一键触发诊断，自动完成节点日志采集、日志清洗、集中诊断的全流程，帮助用户快速定位任务故障根因。

### 组件功能

- 按任务维度一键触发诊断，自动完成节点日志采集调度、本地清洗与集中诊断。
- 维护任务Pod中心关系缓存与挂载关系表，支持plog等任务日志的挂载路径反查。
- 每次诊断之后缓存诊断结果，重复诊断直接返回缓存，支持通过`--refresh`参数强制刷新。
- 可选对接LLM服务，对诊断报告进行智能总结，输出根因报告。

### 组件上下游依赖

1. Agent Core从K8s API Server获取任务请求。
2. Agent Core从ConfigMap中获取任务缓存信息、训练/推理任务的挂载对和env信息。
3. Agent Core从训练/推理任务容器中获取挂载信息、env信息和元数据。
4. Agent Core发送采集指令给Node Collector。
5. Node Collector将清洗结果回传给Agent Core。

---

## 支持的 Tags 及 Dockerfile 链接

### Tag 规范

Tag 遵循以下格式：

```text
<版本>
```

| 字段   | 示例值    | 说明                             |
|--------|----------|----------------------------------|
| `版本` | `v26.2.0` | Ascend ClusterOps Agent 版本号 |

> [!NOTE]
> 与ClusterD等组件不同，本组件镜像基于统一的python:3.13-slim基础镜像构建（组件whl为纯Python包，不依赖架构），因此Tag中不包含操作系统字段。

### Ascend ClusterOps Agent 最新版本 26.2.0

如下所示是Ascend ClusterOps Agent最新发布的26.2.0版本的所有镜像，历史版本所有Tag请参考[supported_tags.md](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-clusterops-agent/supported_tags.md)。

| Tag       | Dockerfile                                                                                                                              | 镜像内容                                                          |
|-----------|-----------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------|
| `v26.2.0` | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-clusterops-agent/v26.2.0/Dockerfile) | Ascend ClusterOps Agent v26.2.0（基础镜像python 3.13-slim） |

---

## 快速开始

### 前置要求

#### 软件依赖

| 软件名称    | 支持的版本                            | 安装位置 | 说明                                                                     |
|-------------|---------------------------------------|----------|--------------------------------------------------------------------------|
| Kubernetes  | 1.17.x~1.34.x（推荐使用1.19.x及以上版本） | 所有节点 | 了解K8s的使用请参见[Kubernetes文档](https://kubernetes.io/zh-cn/docs/)     |
| kubectl     | 与Kubernetes版本配套                  | 用户机   | Kubectl插件的运行依赖                                                    |
| socat       | -                                     | 用户机   | 确保`kubectl port-forward`功能正常                                       |

#### 硬件规格要求

| 名称  | 管理节点（Agent Core） | 计算节点（Node Collector） |
|-------|------------------------|----------------------------|
| CPU   | 2核                    | 4核                        |
| 内存  | 2GB                    | 4GB                        |

### 在线获取 Ascend ClusterOps Agent 镜像

1. 拉取官方镜像

   拉取昇腾镜像仓库提供的Ascend ClusterOps Agent镜像，替换{tag}为实际版本号。

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-clusterops-agent:{tag}
   ```

2. 修改镜像标签

   为拉取的官方镜像重新打本地标签，统一本地镜像命名规范，方便后续运维管理。

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-clusterops-agent:{tag} ascend-clusterops-agent:{tag}
   ```

### 本地构建（可选）

示例场景：构建 linux-aarch64 架构、v26.2.0 版本的Ascend ClusterOps Agent组件镜像。

1. 获取Dockerfile

   前往[支持的 Tags 及 Dockerfile 链接](#支持的-tags-及-dockerfile-链接)章节，打开目标版本对应的Dockerfile链接，保存文件至aarch64架构环境的本地目录。

2. 本地构建 Docker 镜像（禁用缓存，保证构建纯净度）

   ```bash
   docker build --no-cache -t ascend-clusterops-agent:v26.2.0 ./
   ```

> [!NOTE]
>
> - 若Docker版本低于18.09，或未手动开启BuildKit，构建镜像时将无法读取TARGETPLATFORM变量，会造成镜像构建失败。构建前可执行以下命令临时开启BuildKit：
>
>   ```bash
>   export DOCKER_BUILDKIT=1
>   ```
>
> - 构建过程需要联网：Dockerfile会自动从MindCluster发布件下载组件软件包（`Ascend-mindxdl-ascend-clusterops-agent_{version}_linux.zip`）和日志清洗依赖软件包（`Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip`），并按目标架构自动选择匹配的wheel；镜像内pip会自动安装grpcio、fastapi等运行时依赖。

### 部署 Ascend ClusterOps Agent

部署前需将YAML文件内的镜像`{tag}`替换为实际使用的镜像版本。

1. 部署Agent Core（Deployment，部署在管理节点），操作请参见[MindCluster 集群调度组件开发指南-手动安装-Agent Core](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/05_developer_guide/00_installation_deployment/00_manual_installation/15_agent_core.md)。

2. 部署Node Collector（DaemonSet，每个NPU节点一份），操作请参见[MindCluster 集群调度组件开发指南-手动安装-Node Collector](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/05_developer_guide/00_installation_deployment/00_manual_installation/16_node_collector.md)。

3. 安装Kubectl插件（安装在用户机），操作请参见[MindCluster 集群调度组件开发指南-手动安装-Kubectl Plugin](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/05_developer_guide/00_installation_deployment/00_manual_installation/14_kubectl_plugin.md)。

4. 验证部署

   ```bash
   kubectl get pods -A | grep -E "agent-core|node-collector"
   ```

   预期结果：`agent-core`相关Pod和各计算节点上的`node-collector`相关Pod状态为Running。

---

## 许可证

查看这些镜像中包含的 Mind 系列软件的[许可证信息](https://www.hiascend.com/zh/legal/softlicense)。

与所有容器镜像一样，预装软件包（Python、系统库等）可能受其自身许可证约束。
