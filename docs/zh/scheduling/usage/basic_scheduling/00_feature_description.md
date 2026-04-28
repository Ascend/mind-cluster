# 特性说明<a name="ZH-CN_TOPIC_0000002511347091"></a>

基础调度包含如下特性：

- 训练任务：[整卡调度](../../introduction.md#整卡调度)、[静态vNPU调度](../../introduction.md#静态vnpu调度)、[多级调度](../../introduction.md#多级调度)和[弹性训练](../../introduction.md#弹性训练)。若使用断点续训请参见[断点续训](../../usage/resumable_training/00_feature_description.md)。
- 推理任务：[整卡调度](../../introduction.md#整卡调度)、[静态vNPU调度](../../introduction.md#静态vnpu调度)、[动态vNPU调度](../../introduction.md#动态vnpu调度)、[软切分调度](../../introduction.md#软切分调度)、[推理卡故障恢复](../../introduction.md#推理卡故障恢复)和[推理卡故障重调度](../../introduction.md#推理卡故障重调度)。

    不同的特性依赖不同的组件，详细介绍请参见[基础调度](../../introduction.md#基础调度)章节。

本文档演示如何基于某模型部署并执行使用NPU的训练或推理任务。生产环境与示例存在差异，本章节内示例仅做参考，用户需要根据实际生产环境做修改。

## 任务类型<a name="section14151030191813"></a>

Ascend Operator提供以下2种方式配置资源信息：

- 通过环境变量配置资源信息：为不同AI框架的分布式训练任务提供相应的环境变量，请参见[环境变量说明](../../api/environment_variable_description.md)中"Ascend Operator环境变量说明"。使用此方式的用户仅支持创建Ascend Job（以下简称acjob）对象。
- 通过文件配置资源信息：训练任务集合通信配置文件（RankTable File，也叫[hccl.json](../../api/hccl.json_file_description.md)）。使用此方式的用户支持创建以下3种类型的对象：Volcano Job（以下简称vcjob）、Ascend Job（以下简称acjob）和Deployment（以下简称deploy）。
    - （推荐）Ascend Job：简称acjob，是MindCluster自定义的一种任务类型，当前支持通过环境变量配置资源信息及文件配置资源信息这2种方式拉起训练或推理任务。

        每个acjob任务YAML中包含一些固定字段，例如apiVersion、kind等，如果想了解这些字段的详细说明请参见[acjob关键字段说明](../../api/ascend_job.md)。

    - Volcano Job：简称vcjob，适用于批处理任务，任务有完成状态。
    - Deployment：简称deploy，适用于后台常驻任务，任务没有完成状态。在需要持续训练任务、持续占用资源，调试训练任务，或者提供推理服务接口的时候选用。

        >[!NOTE] 
        >不支持Deployment的更新操作，如果需要更新，请先删除再创建。

## 调度时间说明<a name="section12177114564719"></a>

Volcano在多任务或者单任务场景下，在Atlas 800T A2 训练服务器设备上acjob任务的调度参考时间说明如下。若要达到以下参考时间，需要确保CPU的频率至少为2.60GHz，API Server时延不超过80毫秒。其中调度时间是指任务下发到Pod状态为Running的时间。

- 多任务调度时间说明。
    - 并发创建多个单机单卡任务数量的峰值为100个，即用100个任务YAML同时创建100个单机单卡任务，这100个单机单卡任务的调度时间为107秒。
    - 每秒稳定创建单机单卡任务数为5个，连续稳定创建1分钟后，可以创建300个单机单卡任务，这300个单机单卡任务的调度时间为293秒。

- 单任务调度时间说明如[表1](#table18378013481)所示。

    **表 1**  单任务多Pod调度说明

    <a name="table18378013481"></a>

    |集群节点数|Pod数量|调度时间|
    |--|--|--|
    |100|100|14秒|
    |500|500|57秒|
    |1000|1000|114秒|
    |2000|2000|228秒|
    |3000|3000|269秒|
    |4000|4000|300秒|
    |5000|5000|400秒|

    >[!NOTE] 
    >- 单任务多Pod场景即用1个任务YAML创建多个Pod，比如1个任务YAML创建100个Pod，这100个Pod分别调度到100个节点上的调度时间为14秒。
    >- 若想要达到4000或5000节点的优化调度参考时间，需要参见[安装Volcano](../../installation_guide/03_installation/manual_installation/05_volcano.md#安装volcano)中调度时间性能调优步骤进行相应修改。
    >- 当前vcjob任务的调度规格最大支持1000节点。
