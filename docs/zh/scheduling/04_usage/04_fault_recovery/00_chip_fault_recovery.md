# 芯片故障恢复<a name="ZH-CN_TOPIC_0000002479227136"></a>

芯片故障恢复特性利指的是Ascend Device Plugin的离线热复位功能，当集群中的芯片资源发生部分硬件故障时，能够自动恢复芯片到健康状态。

## 使用前必读<a name="section_precondition"></a>

**前提条件<a name="section_prerequisite"></a>**

- 离线热复位特性一般搭配**整卡调度特性**和**重调度**一起使用，具体使用方式请参考[整卡调度](../03_basic_scheduling/03_full_npu_scheduling.md)和[重调度](../03_basic_scheduling/05_rescheduling.md)。
- 开启离线热复位特性只必须将安装Ascend Device Plugin，并且将启动参数“-hotReset”取值设置为“0”或“2”（默认为“-1”，不支持离线热复位功能）参考[手动安装Ascend Device Plugin](../../05_developer_guide/00_installation_deployment/00_manual_installation/04_ascend_device_plugin.md)。
- 仅芯片故障支持热复位恢复。

## 支持的设备类型

- Atlas 800 训练服务器（型号 9000）（NPU满配）
- Atlas 800 训练服务器（型号 9010）（NPU满配）
- Atlas 900T PoD Lite
- Atlas 900 PoD（型号 9000）
- Atlas 800T A2 训练服务器
- Atlas 900 A2 PoD 集群基础单元
- Atlas 900 A3 SuperPoD 超节点
- Atlas 800T A3 超节点服务器
- Atlas 850E 超节点
- Atlas 650E 服务器
- Atlas 950 SuperPoD 超节点
- Atlas 350 加速卡
- Atlas 300I Pro 推理卡
- Atlas 300V 视频解析卡
- Atlas 300V Pro 视频解析卡
- Atlas 300I Duo 推理卡
- Atlas 300I 推理卡（型号 3000）（整卡）
- Atlas 300I 推理卡（型号 3010）
- Atlas 800I A2 推理服务器
- A200I A2 Box 异构组件
- Atlas 800I A3 超节点服务器

## 原理说明

支持执行芯片复位的故障级别如下：

- RestartRequest
- RestartBusiness
- FreeRestartNPU
- RestartNPU

故障级别说明参考[故障级别](../../06_api/02_ascend_device_plugin.md#自定义芯片故障)。

### 离线热复位流程

1. Ascend Device Plugin启动后，会通过驱动的DCMI接口获取芯片健康状态，若芯片发生了故障则能够从驱动接口拿获取到故障信息，参考[芯片故障](../11_fault_detection_and_diagnosis/02_fault_types/02_chip_faults.md)。

2. Ascend Device Plugin根据获取到的故障信息，判断故障所属的故障级别，进而判断是否需要对芯片进行热复位操作。

3. 部分芯片故障理论上在任务结束后就会恢复，故Ascend Device Plugin会等待60秒后，检查故障芯片是否已经恢复健康状态。

4. 若故障芯片在60秒内未恢复健康状态，则Ascend Device Plugin会根据不同服务器类型，检查故障芯片及与其关联的需要一并复位的芯片是否空闲。

5. 若故障芯片及与其关联的需要一并复位的芯片均空闲，则Ascend Device Plugin会调用驱动接口对其进行热复位。

>[!NOTE]
>
>- 若芯片发生故障时，其已经被调度并挂载到了任务容器中，则若要使其空闲，则需要任务配置重调度策略将故障Pod重新调度到其他空闲的NPU上，参考[重调度](../03_basic_scheduling/05_rescheduling.md)。
>
>- Atlas 800I A2 推理服务器存在以下两种故障恢复方式，一台Atlas 800I A2 推理服务器只能使用一种故障恢复方式，由集群调度组件自动识别使用哪种故障恢复方式。
>
>    - 方式一：若设备上不存在HCCS环，执行推理任务中，当NPU出现故障，Ascend Device Plugin等待该NPU空闲后，对该NPU进行复位操作。
>
>    - 方式二：若设备上存在HCCS环，执行推理任务中，当服务器出现一个或多个故障NPU，Ascend Device Plugin等待环上的NPU全部空闲后，一次性复位环上所有的NPU。
>
>- 对于Atlas 9000 A3 SuperPoD 集群算力系统会复位指定芯片所在的NPU模组；对于Atlas 900 A3 SuperPoD 超节点、Atlas 800T A3 超节点、Atlas 800I A3 系列硬件产品、A200T A3 Box8 系列硬件产品会复位指定芯片所在的NPU模组及与其具备网口互助关系的NPU模组。
>
>- Atlas A3系列产品在调用驱动的带内热复位接口执行热复位失败后，会再次尝试调用带外热复位接口进行热复位。
>
>- 热复位恢复无法覆盖所有故障，部分故障可能恢复失败，例如，故障导致掉卡，device OS挂死等故障。
