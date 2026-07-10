# 使用前必读<a name="ZH-CN_TOPIC_000000_task_alternation_overview"></a>

## 概述<a name="section_overview_alternation"></a>

在大模型推理与训练共享NPU集群的场景中，推理任务根据业务流量动态扩缩容，训练任务长期占用NPU资源。当推理高峰期需要更多NPU时，可通过Volcano的Preempt（抢占）或Reclaim（回收）机制从训练任务回收资源；推理低峰期释放资源后，训练任务恢复运行。

本最佳实践描述如何配置推理与训练任务的交替运行，并利用**重建Pod优先调度回原节点**特性，当两种任务的Pod重建时都优先回到此前运行的节点，复用节点上已缓存的容器镜像，将启动时间从5~20分钟降至秒级。

**核心收益：**

| 场景 | 无回原节点偏好               | 启用回原节点偏好 |
|------|-----------------------|----------------|
| 训练Pod被Preempt/Reclaim驱逐后恢复 | 随机分配节点，重新拉取20~40 GB镜像 | 优先回原节点，复用镜像缓存 |
| 推理Pod缩容后再扩容 | 随机分配节点，重新拉取15~30 GB镜像 | 优先回原节点，复用镜像缓存 |
| 启动延迟 | 分钟级                   | 秒级 |
| 镜像拉取带宽 | 每次重建都重新拉取             | 仅首次拉取 |

## 前提条件<a name="section_prerequisites_alternation"></a>

- 已完成Volcano、Ascend Device Plugin、Ascend Docker Runtime、Ascend Operator、ClusterD、NodeD的安装部署，详细请参见[安装部署](../../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md)。
- ascend-volcano-plugin版本 ≥ 26.1.0（包含回原节点偏好功能）。
- NPU集群节点已配置镜像缓存。

## Preempt与Reclaim的区别<a name="section_preempt_vs_reclaim"></a>

| 特性 | Preempt（抢占） | Reclaim（回收） |
|------|---------------|----------------|
| 触发条件 | 高优先级任务找不到足够资源 | 高权重队列资源不足，从低权重队列回收 |
| 资源释放粒度 | 任务级别（按PriorityClass） | 队列级别（按Queue weight） |
| 配置方式 | PriorityClass + gang.enablePreemptable | Queue.weight + Queue.reclaimable |
| 适合场景 | 推理SLO严格，训练可接受中断 | 按业务线管理资源优先级 |

更多关于Preempt和Reclaim的详细说明，请参见[Volcano官方文档Actions](https://volcano.sh/docs/Scheduler/Actions)。

## 训练任务重启机制<a name="section_restart"></a>

训练任务配置`minAvailable`等于`replicas`以保证gang完整性。当Pod被Preempt或Reclaim驱逐后，Pod数量低于`minAvailable`，`fault-scheduling: grace`标签触发重调度模块自动级联清理剩余Pod并重启Job。Job重启后Pod重新进入调度，通过回原节点特性优先回到原节点。

## 支持的产品形态<a name="section_products_alternation"></a>

- Atlas 800 训练服务器
- Atlas 800I A2 推理服务器
- Atlas 900 A3 SuperPoD 超节点
- Atlas 9000 A3 SuperPoD 集群算力系统
- Atlas 推理系列产品
- A200I A2 Box 异构组件
- Atlas 350 标卡
