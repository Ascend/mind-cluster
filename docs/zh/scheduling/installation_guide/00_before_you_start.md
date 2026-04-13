# 安装

## 安装前必读<a name="ZH-CN_TOPIC_0000002511426285"></a>

在安装组件前，用户需详细阅读[简介](../introduction.md#概述)章节，了解集群调度各组件功能详细的说明，并根据要使用的特性选择安装相应的组件。

Elastic Agent和TaskD组件需部署在容器内，详细安装步骤请参见[制作镜像](../usage/resumable_training.md#制作镜像)。

>[!NOTE] 
>Resilience Controller和Elastic Agent组件已经日落，Resilience Controller相关内容将于2026年9月30日的版本删除；Elastic Agent相关内容将于2026年12月30日的版本删除。

**使用约束<a name="section933252483715"></a>**

- 请确保根目录有足够的磁盘空间，根目录的磁盘空间利用率高于85%会触发kubelet的资源驱逐机制，将导致服务不可用。磁盘空间要求说明请参见[表1](./01_environment_dependencies.md#软硬件规格要求)；驱逐策略请查看[Kubernetes官方文档](https://kubernetes.io/zh-cn/docs/concepts/scheduling-eviction/node-pressure-eviction/)。
- 为保证MindCluster集群调度组件的正常安装及使用，同一集群下，不同训练服务器的系统时间请保持一致。
- ARM架构和x86\_64架构使用的集群调度组件镜像不能相互兼容。
- K8s默认的证书有效期为365天，到期前需要用户自行更新。

**组件部署说明<a name="section1563217510232"></a>**

安装部署集群调度组件时，可以参考[图1](#fig87391254145620)，将相应的集群调度组件或其他第三方软件安装到相应的节点上。大部分组件都使用容器化方式部署；Ascend Docker Runtime使用二进制方式部署；只有NPU Exporter组件既可以使用容器化方式部署，又可以使用二进制方式部署。

**图 1**  组件安装部署<a name="fig87391254145620"></a>  
![](../../figures/scheduling/installation_guide_001.PNG "installation_guide_001")

>[!NOTE] 
>MindCluster提供Volcano组件，该组件在开源Volcano上集成了昇腾插件Ascend-volcano-plugin。

**日志路径说明<a name="section4837236204914"></a>**

- Ascend Docker Runtime日志路径为“/var/log/ascend-docker-runtime/”。
- 其他集群调度组件日志路径可参考[创建日志目录](./03_installation.md#创建日志目录)章节。
