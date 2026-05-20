# 推理卡故障恢复<a name="ZH-CN_TOPIC_0000002479227136"></a>

**推理卡故障恢复特性**需要搭配**整卡调度特性**一起使用，开启推理卡故障恢复特性只需要将Ascend Device Plugin的启动参数“-hotReset”取值设置为“0”或“2”（默认为“-1”，不支持故障恢复功能）。具体使用方式请参考[整卡调度或静态vNPU调度（推理）](./04_full_npu_scheduling_and_static_vnpu_scheduling_inference.md)。

Atlas 800I A2 推理服务器、A200I A2 Box 异构组件使用**推理卡故障恢复特性**，仅支持下发单机单卡任务，不支持分布式任务，且需要单独使用[infer-vcjob-910-hotreset.yaml](https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/infer-vcjob-910-hotreset.yaml)示例下发任务。

>[!NOTE]
>Atlas 800I A2 推理服务器存在以下两种故障恢复方式，一台Atlas 800I A2 推理服务器只能使用一种故障恢复方式，由集群调度组件自动识别使用哪种故障恢复方式。
>
>- 方式一：若设备上不存在HCCS环，执行推理任务中，当NPU出现故障，Ascend Device Plugin等待该NPU空闲后，对该NPU进行复位操作。
>- 方式二：若设备上存在HCCS环，执行推理任务中，当服务器出现一个或多个故障NPU，Ascend Device Plugin等待环上的NPU全部空闲后，一次性复位环上所有的NPU。
