# 硬件故障恢复<a name="ZH-CN_TOPIC_0000002479227136"></a>

硬件故障恢复特性利用Ascend Device Plugin的离线热复位功能，当集群中的芯片资源发生部分硬件故障时，能够自动恢复芯片到健康状态。

## 使用前必读<a name="section_precondition"></a>

**前提条件<a name="section_prerequisite"></a>**

- 离线热复位特性需要搭配**整卡调度特性**一起使用，具体使用方式请参考[整卡调度](./03_full_npu_scheduling.md)。
- 开启离线热复位特性只需将Ascend Device Plugin的启动参数“-hotReset”取值设置为“0”或“2”（默认为“-1”，不支持离线热复位功能），开启方式请参考[手动安装Ascend Device Plugin](../../05_developer_guide/00_installation_deployment/00_manual_installation/04_ascend_device_plugin.md#li_config_hotreset)的步骤6。

**离线热复位流程**：

1、使用离线热复位特性时，任务需要配置重调度策略，故障重调度特性请参考[故障重调度](./05_rescheduling_upon_inference_card_faults.md)。

2、当芯片发生故障后，集群调度组件会先将故障芯片上运行的Pod重调度到其他健康节点，待故障芯片上的Pod全部调度走（即故障芯片处于空闲状态）后，Ascend Device Plugin再对故障芯片进行热复位操作。

>[!NOTE]
>
>- Atlas 800I A2 推理服务器存在以下两种故障恢复方式，一台Atlas 800I A2 推理服务器只能使用一种故障恢复方式，由集群调度组件自动识别使用哪种故障恢复方式。
>
>    - 方式一：若设备上不存在HCCS环，执行推理任务中，当NPU出现故障，Ascend Device Plugin等待该NPU空闲后，对该NPU进行复位操作。
>
>    - 方式二：若设备上存在HCCS环，执行推理任务中，当服务器出现一个或多个故障NPU，Ascend Device Plugin等待环上的NPU全部空闲后，一次性复位环上所有的NPU。
>
>- 对于Atlas 9000 A3 SuperPoD 集群算力系统会复位指定芯片所在的NPU模组；对于Atlas 900 A3 SuperPoD 超节点、Atlas 800T A3 超节点、Atlas 800I A3 系列硬件产品、A200T A3 Box8 系列硬件产品会复位指定芯片所在的NPU模组及与其具备网口互助关系的NPU模组。
>
>- 热复位恢复无法覆盖所有故障，部分故障可能恢复失败，例如，故障导致掉卡，device OS挂死等故障。
