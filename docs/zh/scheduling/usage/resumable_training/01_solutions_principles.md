# 方案和原理<a name="ZH-CN_TOPIC_0000002511346509"></a>

## 故障检测<a name="ZH-CN_TOPIC_0000002479226514"></a>

### 故障说明<a name="ZH-CN_TOPIC_0000002511426413"></a>

断点续训基于故障检测能力获取集群和训练业务的故障状态，根据检测结果进行故障处理。当前，断点续训特性主要提供以下几个方面的故障检测能力：昇腾硬件故障、训练业务故障、其他故障发送方的故障。

MindCluster集群调度组件Ascend Device Plugin提供NPU芯片故障检测能力及NPU参数面网络故障检测能力，NodeD提供服务器节点故障、DPC共享存储故障和灵衢网络故障检测能力，ClusterD提供公共故障检测能力，Volcano提供业务面容器异常检测能力，故障检测整体架构如下图所示。

![](../../../figures/scheduling/250411110432760.png)

1. 计算服务器上的Ascend Device Plugin通过驱动获取NPU芯片故障以及参数面网络故障后，将故障信息上报到管理服务器。
2. 计算服务器上的NodeD通过驱动获取服务器节点故障、DPC共享存储故障和灵衢网络故障信息后，将故障信息上报到管理服务器。
3. 计算服务器上的K8s监测训练容器状态，训练容器异常后上报到K8s中，管理服务器上的Volcano通过K8s获取训练容器的故障信息。
4. 管理服务器上的ClusterD通过公共故障接口获取公共故障后，将接收到的信息进行汇总写入cluster-info-device-cm。
5. （可选）管理服务器上的ClusterD汇总集群内所有Ascend Device Plugin和NodeD上报的故障信息。

**支持的故障模式<a name="zh-cn_topic_0000002039699773_section8301627182117"></a>**

当前已支持200+故障的检测。支持的故障类型请参见[表1](#zh-cn_topic_0000002039699773_table9980135316395)，详细的故障说明请参见[典型故障.xlsx](../../../resource/典型故障.xlsx)。

**表 1**  故障类型说明

<a name="zh-cn_topic_0000002039699773_table9980135316395"></a>

|故障类型|故障说明|
|--|--|
|节点故障|<p>包括节点健康状态、节点硬件故障和DPC共享存储故障。</p><p>故障码说明请参见[节点故障码参考文档](../../appendix.md#节点故障码参考文档)。</p><p>若节点的硬件故障导致节点宕机或重启，则NodeD无法检测到具体的故障类型并上报。</p>|
|芯片故障|<p>DCMI接口上报的芯片故障和设备网络探测工具hccn_tool检测到的芯片网络故障。</p><p>故障码说明请参见[芯片故障码参考文档](../../appendix.md#芯片故障码参考文档)。</p>|
|参数面网络故障|包括芯片网络相关故障和灵衢总线设备故障。<ul><li>芯片网络相关故障：芯片之间进行参数交换的专用网络出现故障，如NPU网口故障。</li><li>灵衢总线设备故障：<term>Atlas A3 训练系列产品</term>的灵衢总线设备发生故障。</li></ul>|
|业务面故障|<p>训练任务异常退出，导致Pod的Status变为Failed状态。</p><p>可执行<strong>kubectl describe pod <em>{pod名称} </em>-n <em>\{NAMESPACE\}</em> \|grep Status:</strong>命令，查看当前Pod的Status是否为Failed状态。回显示例如下：<pre class="screen"><strong>Status:       Failed</strong></pre></p>|
|公共故障|公共故障指的是其他故障发现者（非MindCluster组件）提供的故障，公共故障包括以下几种类型：NPU故障、节点故障、网络故障和存储故障。|
|pingmesh灵衢网络故障|灵衢网络故障是针对超节点内部（包括节点内和节点间）的HCCS网络提供的NPU网络故障检测。|
|性能劣化故障|MindCluster结合MindStudio提供的profiling能力对集群中的性能劣化故障（慢节点）提供诊断功能。该功能提供动态使能打点和打点数据持久化功能、可动态启停，无需重启任务进行诊断，对训练无损耗。|

**ConfigMap说明<a name="zh-cn_topic_0000002039699773_section49901206282"></a>**

- 每个计算节点的Ascend Device Plugin均会创建记录本节点NPU和灵衢总线设备信息的ConfigMap文件。该ConfigMap文件名为mindx-dl-deviceinfo-_<nodename\>_（以下简称device-info-cm），故障信息会通过该ConfigMap进行上报。该ConfigMap文件中各字段的说明，请参见[DeviceInfoCfg](../../api/ascend_device_plugin.md#芯片资源)表。
- 当节点上存在节点故障时，每个计算节点的NodeD会创建记录本节点设备信息的ConfigMap文件。该ConfigMap文件名为mindx-dl-nodeinfo-_<nodename\>_（以下简称node-info-cm），节点故障信息会通过该ConfigMap进行上报。该ConfigMap文件中各字段的说明，请参见[mindx-dl-nodeinfo-<nodename\>](../../api/noded.md#节点资源)表。
- ClusterD会创建记录本集群设备信息的ConfigMap文件，该ConfigMap文件名为cluster-info-<device/switch\>-<\[0-5\]\>、cluster-info-node-cm（以下简称cluster-info-cm）。节点及芯片故障信息会通过[cluster-info-cm](../../api/clusterd/00_cluster_resources.md)进行上报。
- 创建每个任务时，需要在YAML中配置ConfigMap文件，该ConfigMap文件名称为reset-config-_<job-name\>_（以下简称reset-info-cm）。该ConfigMap挂载到容器的“/user/restore/reset/config”路径下。Ascend Device Plugin会自动将ConfigMap挂载到本节点的“/user/restore/reset/<job-namespace\>.<job-name\>”路径下。

    也可以将节点上/user/restore/reset/<job-namespace\>.<job-name\>替代ConfigMap，挂载到容器的“/user/restore/reset/config”路径下。该ConfigMap文件字段说明，请参见[reset-config-<job-name\>](../../api/ascend_device_plugin.md#任务信息)表。

### 节点故障<a name="ZH-CN_TOPIC_0000002479386528"></a>

节点故障的发现主要通过NodeD组件实现。节点故障包括节点健康状态和节点硬件故障、节点DPC共享存储故障，详细说明如下：

- 节点健康状态

    NodeD完成当前节点的节点状态诊断后，收集本节点内的故障信息。当节点发生故障时，通过节点状态上报机制不断向Volcano发送节点状态（当前仅收集本节点内的硬件故障信息）。

- 节点硬件故障

    针对节点硬件故障，NodeD通过IPMI驱动向iBMC发送故障查询请求，iBMC将当前硬件告警信息响应给NodeD。NodeD收集硬件告警信息后，将节点硬件状态上报给Volcano。

- 节点DPC共享存储故障

    针对使用Scale-Out Storage DPC产品的节点，可以使用NodeD安装包下的noded-dpc-\{version\}.yaml启动NodeD服务。开启对DPC的进程异常及内存不足异常的检测和上报。

    >[!NOTE] 
    >当节点发生故障时，NodeD会上报节点健康状态和节点硬件故障。无故障时，默认节点健康。

**图 1**  节点故障上报<a name="fig1329112151382"></a>  
![](../../../figures/scheduling/节点故障上报.png "节点故障上报")

- 当节点发生故障时，NodeD最短5秒（默认）更新本节点的node-info-cm内容，其中字段说明见[mindx-dl-nodeinfo-<nodename\>](../../api/noded.md#节点资源)表。
- NodeD每隔60秒（默认）从iBMC查询故障信息。当查到的故障信息相比上次查到的有变化或与上次上报的时间间隔30分钟以上时，会在1秒内上报到node-info-cm中。

**所需组件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

为保证节点故障检测功能的正常使用，需要安装以下组件：Volcano、Ascend Operator、NodeD、ClusterD

**使用约束<a name="section16867482102"></a>**

- NodeD的节点硬件故障上报能力仅支持以下产品型号：Atlas 800T A2 训练服务器、Atlas 900 A2 PoD 集群基础单元、Atlas 900 A3 SuperPoD 超节点。
- 仅V2 3.15.0.1及以上版本或者V2 3.10.02.55版本的iBMC，且安装了IPMC驱动的产品，支持NodeD的节点硬件故障上报能力。低版本的iBMC或IPMI获取节点故障信息失败时，将只上报节点健康状态。
- 如需使用超节点故障检测功能，需使用V3 5.8.3.35及以上版本的iBMC。
- 如需使用DPC故障检测功能，需使用Scale-Out Storage DPC 24.2.0及以上版本。

**支持的故障处理类型<a name="section099935818571"></a>**

Job级别重调度、Pod级别重调度、进程级别重调度

**（可选）配置故障检测的级别<a name="section1343172016386"></a>**

断点续训针对节点故障中**节点硬件故障**的不同故障码，提供了默认的故障级别和对应级别的故障处理策略。若用户需要修改故障处理策略，可参见[节点硬件故障](./03_configuring_fault_detection_levels.md#节点硬件故障)。若无特殊需求，请勿随意修改。

### 芯片故障<a name="ZH-CN_TOPIC_0000002511346395"></a>

芯片故障指的是NPU出现的基础软件类故障和芯片硬件类故障。断点续训特性中芯片故障的检测和上报由设备管理组件Ascend Device Plugin负责。

**NPU上报机制<a name="section15950121613265"></a>**

NPU发生故障时，故障管理框架获取到故障信息后，将该信息上传给NPU驱动的故障管理框架。故障管理框架收到故障信息后，通过DCMI接口上报给Ascend Device Plugin，如[图1](#fig3951191610267)所示。

Ascend Device Plugin通过DCMI接口获取芯片健康状态。当前提供如下两种获取模式：

- 故障订阅模式。Ascend Device Plugin启动时会先调用DCMI故障订阅接口注册监测，故障发生时驱动通过该接口将故障事件上报给Ascend Device Plugin。故障恢复时通过该接口将恢复事件上报给Ascend Device Plugin。
- 故障轮询模式。每隔固定时间，通过故障查询接口查询芯片故障状态，当设备驱动不支持订阅能力时将切换该模式。

**图 1**  芯片故障上报<a name="fig3951191610267"></a>  
![](../../../figures/scheduling/芯片故障上报.png "芯片故障上报")

**Ascend Device Plugin上报机制<a name="section0951116132615"></a>**

Ascend Device Plugin获取到芯片故障信息后，通过ConfigMap的形式上报给K8s。Ascend Device Plugin的故障上报机制如下：

**图 2**  上报故障到K8s<a name="fig10951101692610"></a>  
![](../../../figures/scheduling/上报故障到K8s.png "上报故障到K8s")

对于不同故障处理模式，上报的路径会有一定差别。

- 重调度模式：Ascend Device Plugin获取到芯片故障后，将芯片故障信息写入该节点所属的device-info-cm中，其中字段说明见[DeviceInfoCfg](../../api/ascend_device_plugin.md#芯片资源)表。ClusterD读取每个节点的device-info-cm感知芯片故障并上报给调度器。
- 优雅容错模式：Ascend Device Plugin获取到可恢复的芯片故障后，将芯片故障信息写入该任务所属的reset-info-cm中，业务容器通过将reset-info-cm挂载为文件的形式，读取文件感知芯片故障。

    >[!NOTE] 
    >若优雅容错模式处理故障失败，回退至重调度模式后，故障上报的路径则按照重调度模式进行上报。

**所需组件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

为保证芯片故障检测功能的正常使用，需要安装以下组件：Volcano、Ascend Operator、Ascend Device Plugin、ClusterD

**（可选）配置故障检测的级别<a name="section1343172016386"></a>**

断点续训针对**芯片故障**提供了默认的故障频率、时长、故障级别以及对应级别的故障处理策略。若用户需要修改故障处理策略，可参见[芯片故障](./03_configuring_fault_detection_levels.md#ZH-CN_TOPIC_0000002511346521_0101)。若无特殊需求，请勿随意修改。

**支持的故障处理类型<a name="section099935818571"></a>**

Job级别重调度、Pod级别重调度、进程级别重调度、进程级在线恢复、优雅容错

>[!NOTE] 
>仅片上内存出现的不可纠正错误支持进程级在线恢复，其他类型的芯片故障不支持进程级在线恢复。

### 参数面网络故障<a name="ZH-CN_TOPIC_0000002511426381"></a>

NPU的参数面网络故障包括芯片网络相关故障和灵衢总线设备故障。

参数面网络出现故障时，将导致训练任务中断或者训练任务性能较差。灵衢总线设备发生故障后，MindCluster集群调度组件将根据故障级别进行相应的重调度处理。

>[!NOTE]
>
>- 参数面网络故障不会直接触发任务重调度，当参数面故障导致训练任务异常中断时才触发任务重调度。
>- 如果需要对参数面网络故障进行故障处理，需要同时开启业务面故障无条件重试能力。

参数面网络故障检测由设备管理组件Ascend Device Plugin负责，详细原理如[图1](#fig68743107307)所示。

**图 1**  故障检测<a name="fig68743107307"></a>  
![](../../../figures/scheduling/故障检测.png "故障检测")

**关键步骤说明<a name="section1787471017308"></a>**

**芯片网络故障**：

1. NPU定时检测和网关地址的通信是否正常，探测周期为2.5秒，通过故障管理框架上报结果。
2. RoCE驱动实时监测NPU网口Link状态，通过故障管理框架上报Linkdown或Linkup事件。
3. Ascend Device Plugin通过DCMI接口从故障管理框架获取信息，通过轮询的方式查询网关探测结果，并实时订阅网口Linkdown或Linkup事件并进行上报。Ascend Device Plugin统计网关检测异常持续时间、Linkdown持续时间。如果小于或等于RoCE网络超时时间（默认为20秒）则标记为NPU网络故障（默认不处理，可能会引起参数面网络故障）；如果大于20秒，则升级成配置的故障等级。

**灵衢总线设备故障**：

1. 灵衢总线设备将设备发生的故障写入本地队列中。
2. 灵衢查询接口通过查询上述队列，将故障缓存至查询接口，并进行汇总处理。
3. Ascend Device Plugin通过订阅或轮询的方式调用接口获取灵衢总线设备相关故障，并写入device-info-cm进行上报。

**故障上报机制<a name="section1874141093019"></a>**

- **芯片发生网络故障时**，NPU故障管理框架获取故障信息后，将该信息上报给NPU驱动。NPU驱动收到故障信息后，通过DCMI接口上报给Ascend Device Plugin。Ascend Device Plugin通过DCMI接口获取芯片健康状态。当前提供如下两种获取模式：
    - 故障订阅模式。Ascend Device Plugin启动时会先调用DCMI故障订阅接口注册监测，故障发生或恢复时，驱动通过该接口将故障发生或恢复事件上报给Ascend Device Plugin。
    - 故障轮询模式。每隔固定时间，通过故障查询接口查询芯片故障状态。当设备驱动不支持订阅能力时将切换该模式。

- **灵衢总线设备发生故障时**，Ascend Device Plugin通过灵衢查询接口获取故障信息，当前故障查询提供两种模式：
    - 故障订阅模式：在Ascend Device Plugin启动过程中向灵衢查询接口注册故障处理回调。故障发生后，该回调被调用后将故障上报给Ascend Device Plugin，故障恢复时通过该接口上报恢复事件。
    - 故障轮询模式：Ascend Device Plugin每隔5分钟调用一次全量故障查询接口。

**Ascend Device Plugin上报机制<a name="section1875111093017"></a>**

Ascend Device Plugin获取到参数面网络故障后，将故障信息写入到device-info-cm中，并通过ConfigMap的形式上报给K8s。device-info-cm中各字段的说明，请参见[DeviceInfoCfg](../../api/ascend_device_plugin.md#芯片资源)表。

Ascend Device Plugin的故障上报机制如[图2](#fig1587571063011)所示。

**图 2**  故障上报<a name="fig1587571063011"></a>  
![](../../../figures/scheduling/故障上报.png "故障上报")

**watchdog故障检测<a name="section4599926103917"></a>**

参数面网络链路异常（参数面网络故障）可能导致任务中正常NPU无法与故障NPU通信，使所有NPU集合通信陷入超时等待状态；并使任务集合通信出现等待超时异常后才退出（默认为30分钟）。

开启watchdog功能（且开启了业务面故障无条件重试能力）可以在参数面网络链路异常发生后，隔离故障NPU，将任务重调度到健康的NPU上，从而实现6分钟内使任务快速退出。

>[!NOTE] 
>仅支持在PyTorch及MindSpore框架下使用watchdog功能。

**所需组件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

为保证参数面网络故障检测功能的正常使用，需要安装以下组件：Volcano、Ascend Operator、Ascend Device Plugin、ClusterD

**支持的故障处理类型<a name="section099935818571"></a>**

Job级别重调度、Pod级别重调度、进程级别重调度

**（可选）配置故障检测的级别<a name="section1343172016386"></a>**

断点续训针对**参数面故障**提供了默认的故障级别以及对应级别的故障处理策略，若用户需要修改故障处理策略可参见[参数面网络故障](./03_configuring_fault_detection_levels.md#ZH-CN_TOPIC_0000002479226486)。若无特殊需求，请勿随意修改。

### 业务面故障<a name="ZH-CN_TOPIC_0000002479386512"></a>

断点续训特性支持通过Volcano调度器感知并处理因业务面故障导致的任务失败。业务面故障是因容器内的训练进程均异常退出后引起容器异常退出，导致Pod的Status变为Failed状态。在使用Ascend Operator的场景下，业务面故障仅支持任务的部分Pod发生故障的场景，若任务所有Pod在几秒内Status都转变为Failed，任务不会发生重调度，认定任务为失败状态。

业务面故障发现原理如[图1](#fig1761563615337)所示。

**图 1**  发现原理<a name="fig1761563615337"></a>  
![](../../../figures/scheduling/发现原理.png "发现原理")

调度器不断轮询地查询每个任务的Pod状态，从而感知到业务面故障并上报该故障。用户可根据具体业务需求对业务面故障做处理。断点续训获取到业务面故障后，Volcano会检测是否开启无条件重试功能，开启后会将任务重新调度到未导致本次训练任务重调度的新节点，并重新执行训练任务，重试次数减1；当重试次数为0或者没有开启无条件重试功能时，不会对业务容器故障进行处理。

>[!NOTE]
>
>- 如需使用无条件重试功能，需在任务YAML中配置以下3个参数：fault-retry-times，restartPolicy及policies，详细参数说明请参见[YAML参数说明](./06_configuring_the_job_yaml_file.md#yaml参数说明)（policies 是 vcjob 原生字段）。
>- 在使用Ascend Operator的场景下，若希望任务所有Pod的Status在转变为Failed后仍发生重调度，可参考[使用Volcano和Ascend Operator组件场景下，业务面故障的任务所有Pod的Status全部变为Failed，任务无法触发无条件重试重调度](../../faq.md#使用volcano和ascend-operator组件场景下业务面故障的任务所有pod的status全部变为failed任务无法触发无条件重试重调度)。

**watchdog故障检测<a name="section59641929143117"></a>**

NPU上Task执行异常（业务面故障）可能导致任务中正常NPU无法与故障NPU通信，使正常NPU集合通信陷入超时等待状态，任务集合通信出现等待超时异常后才退出（默认为30分钟）。开启watchdog功能（需同时开启业务面故障无条件重试能力），可以在该异常发生后，隔离故障NPU，将任务重调度到健康的NPU上，从而实现6分钟内使任务快速退出。

>[!NOTE] 
>NPU上Task执行异常仅支持Atlas A2 训练系列产品的PyTorch框架使用watchdog功能。

**所需组件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

为保证业务面故障检测功能的正常使用，需要安装以下组件：Volcano、Ascend Operator

**支持的故障处理类型<a name="section099935818571"></a>**

Job级别重调度、Pod级别重调度、进程级别重调度、优雅容错

### 公共故障<a name="ZH-CN_TOPIC_0000002511426387"></a>

公共故障指的是其他故障发送方（非MindCluster组件）上报的故障，公共故障包括以下几种类型：NPU故障、节点故障、网络故障和存储故障。

>[!NOTE] 
>ClusterD支持接收公共故障的前提是需要在节点上安装Ascend Device Plugin，并且生成了相应的device-info-cm。

**上报机制<a name="zh-cn_topic_0000002216292813_section64469192378"></a>**

公共故障发送方发现故障后，将通过ConfigMap或gRPC方式，将获取到的故障信息发送给ClusterD。ClusterD会将接收到的信息进行汇总写入cluster-info-device-cm，再上报给Ascend-volcano-plugin。

- 通过ConfigMap获取。故障发现者将故障信息写入ConfigMap中，然后由ClusterD获取故障信息。用户可通过调用ConfigMap接口的方式来注入公共故障，详细说明请参见[ConfigMap](../../api/clusterd/03_public_fault_apis.md#configmap)。
- 通过gRPC获取。故障发现者将故障信息通过gRPC通道发送给ClusterD，然后由ClusterD获取故障信息。用户可通过调用gRPC接口的方式来注入公共故障，说明请参见[gRPC接口](../../api/clusterd/03_public_fault_apis.md#grpc接口)。

**图 1**  公共故障上报<a name="fig72618571585"></a>  
![](../../../figures/scheduling/公共故障上报.png "公共故障上报")

**所需组件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

为保证公共故障检测功能的正常使用，需要安装以下组件。

- 必选组件：Volcano、Ascend Operator、Ascend Device Plugin、ClusterD
- 可选组件：NodeD

**支持的故障处理类型<a name="zh-cn_topic_0000002216292813_section177211923175116"></a>**

Job级别重调度、Pod级别重调度、进程级别重调度

**（可选）配置故障检测的级别和发送方<a name="zh-cn_topic_0000002216292813_section1343172016386"></a>**

断点续训针对**公共故障**提供了默认的故障级别以及支持的故障发送方。若用户需要修改公共故障的级别及故障发送方，可参见[公共故障](./03_configuring_fault_detection_levels.md#ZH-CN_TOPIC_0000002479386564)。若无特殊需求，请勿随意修改。

### pingmesh灵衢网络故障<a name="ZH-CN_TOPIC_0000002511426437"></a>

灵衢网络故障是针对超节点内部（包括节点内和节点间）的HCCS网络提供的NPU网络故障检测。

**上报机制<a name="zh-cn_topic_0000002193288232_section68367256347"></a>**

NodeD调用DCMI接口启动pingmesh任务，并周期性查询pingmesh结果，将该结果写入文件<nodename\>.log。该文件所在目录在容器中为固定路径：/user/mind-cluster/pingmesh，物理机默认目录/user/mind-cluster/pingmesh。物理机路径可以修改，修改方式如以下说明所示。

>[!NOTE]
>
>- <nodename\>非固定值，为K8s中查询到的节点名称。
>- <nodename\>.log文件物理机路径可由用户根据实际情况自行配置：在NodeD的启动YAML中修改挂载卷名称为pingmesh-result的物理机挂载路径。

获取pingmesh结果后，ClusterD会对结果进行初步分析，将故障信息写入到名为[pingmesh-fault-<nodename\>](#zh-cn_topic_0000002193288232_table2371535113510)的ConfigMap文件中。ClusterD会侦听该ConfigMap信息，并将故障汇总后上报给Volcano，由Volcano进行调度。

**前提条件<a name="zh-cn_topic_0000002193288232_section8281518121516"></a>**

- （必选）已[创建命名空间](../../installation_guide/03_installation.md#创建命名空间)
- 在相应节点上完成以下组件的安装：[NodeD](../../installation_guide/03_installation.md#noded)（必选）、[Ascend Device Plugin](../../installation_guide/03_installation.md#ascend-device-plugin)（可选）、[ClusterD](../../installation_guide/03_installation.md#clusterd)（可选）
- （必选）已[配置NodeD启动参数resultMaxAge](../../installation_guide/03_installation.md#noded)

**使用约束<a name="zh-cn_topic_0000002193288232_section156679598384"></a>**

本功能仅支持在以下产品型号中使用：Atlas 900 A3 SuperPoD 超节点。

**配置灵衢网络检测<a name="zh-cn_topic_0000002193288232_section18190175418362"></a>**

配置灵衢网络检测，需执行以下步骤。

1. 配置共享存储。

    ClusterD和NodeD通过共享存储进行交互，两者的共享存储根路径需要保持一致。共享目录的根路径属主为9000用户，与ClusterD运行用户一致。

    1. 配置server。

        ![](../../../figures/scheduling/zh-cn_image_0000002479386634.png)

    2. 修改NodeD配置。

        ![](../../../figures/scheduling/zh-cn_image_0000002479386638.png)

    3. 如果存在ClusterD，则需修改ClusterD配置。

        ![](../../../figures/scheduling/zh-cn_image_0000002511346583.png)

    4. 执行**kubectl get pods -o wide -A**命令出现如下示例，则表示已完成共享存储配置。

        ![](../../../figures/scheduling/zh-cn_image_0000002479226664.png)

2. 启用或关闭灵衢网络检测。
    - （推荐）已安装Ascend Device Plugin和ClusterD
        1. 登录环境，进入NodeD解压目录。
        2. 执行以下命令创建名为pingmesh-config的ConfigMap文件。

            pingmesh-config.yaml为pingmesh配置文件，可从NodeD安装包中获取。

            ```shell
            kubectl apply -f pingmesh-config.yaml  
            ```

            回显示例如下。

            ```ColdFusion
            configmap/pingmesh-config created
            ```

        3. 执行以下命令编辑pingmesh-config文件。该文件中各参数的填写说明如[表1](#zh-cn_topic_0000002193288232_table985012534578)所示。

            ```shell
            kubectl edit cm -n cluster-system pingmesh-config
            ```

            **表 1**  pingmesh-config cm

            <a name="zh-cn_topic_0000002193288232_table985012534578"></a>

            |参数|说明|取值|
            |--|--|--|
            |app|ConfigMap其中一个label的key。|pingmesh|
            |global|集群配置信息。|-|
            |"1"|超节点ID为1的配置示例，用户可根据实际情况进行修改或新增。当配置了某个超节点后，NodeD会采用超节点的配置信息而忽略global配置信息。|超节点ID|
            |activate|是否启用pingmesh功能。|on或off|
            |task_interval|pingmesh任务间隔。单位为秒。|[1~60]|

    - 未安装Ascend Device Plugin和ClusterD

        自行生成名为cluster-system的命名空间，name为super-pod-<superPodID\>、label为app=pingmesh的ConfigMap。且该ConfigMap中各字段需按照[super-pod-<super-pod-id\>](../../api/clusterd/00_cluster_resources.md)表填写。示例如下。

        ```Yaml
        apiVersion: v1
        data:
          superPodDevice: '{"SuperPodID":"0","NodeDeviceMap":{"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"62914560","1":"62980097","10":"64225290","11":"64290827","12":"64487436","13":"64552973","14":"64749582","15":"64815119","2":"63176706","3":"63242243","4":"63438852","5":"63504389","6":"63700998","7":"63766535","8":"63963144","9":"64028681"}},"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"67108864","1":"67174401","10":"68419594","11":"68485131","12":"68681740","13":"68747277","14":"68943886","15":"69009423","2":"67371010","3":"67436547","4":"67633156","5":"67698693","6":"67895302","7":"67960839","8":"68157448","9":"68222985"}},"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"104857600","1":"104923137","10":"106168330","11":"106233867","12":"106430476","13":"106496013","14":"106692622","15":"106758159","2":"105119746","3":"105185283","4":"105381892","5":"105447429","6":"105644038","7":"105709575","8":"105906184","9":"105971721"}},"node-**-*":{"NodeName":"node-**-*","DeviceMap":{"0":"4194304","1":"4259841","10":"5505034","11":"5570571","12":"5767180","13":"5832717","14":"6029326","15":"6094863","2":"4456450","3":"4521987","4":"4718596","5":"4784133","6":"4980742","7":"5046279","8":"5242888","9":"5308425"}},"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"142606336","1":"142671873","10":"143917066","11":"143982603","12":"144179212","13":"144244749","14":"144441358","15":"144506895","2":"142868482","3":"142934019","4":"143130628","5":"143196165","6":"143392774","7":"143458311","8":"143654920","9":"143720457"}},"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"146800640","1":"146866177","10":"148111370","11":"148176907","12":"148373516","13":"148439053","14":"148635662","15":"148701199","2":"147062786","3":"147128323","4":"147324932","5":"147390469","6":"147587078","7":"147652615","8":"147849224","9":"147914761"}},"node-**-**":{"NodeName":"node-**-**","DeviceMap":{"0":"83886080","1":"83951617","10":"85196810","11":"85262347","12":"85458956","13":"85524493","14":"85721102","15":"85786639","2":"84148226","3":"84213763","4":"84410372","5":"84475909","6":"84672518","7":"84738055","8":"84934664","9":"85000201"}}}}'
        kind: ConfigMap
        metadata:
          labels:
            app: pingmesh
          name: super-pod-0       # 0为超节点ID
          namespace: cluster-system
        ```

**查看检测结果信息<a name="zh-cn_topic_0000002193288232_section772614207398"></a>**

>[!NOTE] 
>检测结果查询周期为配置参数“task\_interval”的10倍。

灵衢网络检测的pingmesh结果写入文件<nodename\>.log中。该文件中各字段的详细说明如下表所示。

**表 2**  <nodename\>.log

<a name="zh-cn_topic_0000002193288232_table313985322113"></a>

|参数|说明|取值|
|--|--|--|
|uid|该次pingmesh任务的ID。|长度为64的字符串|
|config|该次pingmesh任务的用户配置。|字符串|
|physicID|NPU卡物理ID。|[0~15]|
|taskID|任务ID，0代表节点内部、1代表节点间。|0或1|
|DestNum|本次pingmesh目标地址数量。|[0~47]|
|source_addr|源地址|IPv4网络地址|
|target_addr|目标地址|IPv4网络地址|
|suc_pkt_num|发送成功的包数量。|-|
|fail_pkt_num|发送失败的包数量。|-|
|max_time|最长响应时间。|<ul><li>ping失败的时候，值为-1。</li><li>正常情况下为非负值。</li></ul>|
|min_time|最短响应时间。|<ul><li>ping失败的时候，值为-1。</li><li>正常情况下为非负值。</li></ul>|
|avg_time|平均响应时间。|<ul><li>ping失败的时候，值为-1。</li><li>正常情况下为非负值。</li></ul>|
|tp95_time|处于95%位置的响应时间。|<ul><li>ping失败的时候，值为-1。</li><li>正常情况下为非负值。</li></ul>|
|reply_stat_num|本次查询到的响应数量。|-|
|ping_total_num|本次任务累计的响应数量。|-|

**查看故障信息<a name="zh-cn_topic_0000002193288232_section7712929183110"></a>**

在管理节点上执行以下命令，查看灵衢网络检测的故障信息。

```shell
kubectl describe cm -n cluster-system  pingmesh-fault-<nodename>
```

故障信息中各字段的详细说明如下所示。

**表 3**  pingmesh-fault-<nodename\>

<a name="zh-cn_topic_0000002193288232_table2371535113510"></a>

|参数|说明|取值|
|--|--|--|
|mc-consumer-publicfault|ClusterD侦听所需的label key|true|
|PublicFault|公共故障信息key|详细说明请参见[fault字段说明](../../api/clusterd/03_public_fault_apis.md#configmap)表。|

**已支持的灵衢网络故障<a name="zh-cn_topic_0000002193288232_section4960201383813"></a>**

<a name="zh-cn_topic_0000002193288232_table31451934163811"></a>

|故障码|故障说明|故障级别|
|--|--|--|
|220001001|NPU卡HCCS网络故障|<p>SeparateNPU</p><p>该故障级别不支持自行配置。</p>|

### 性能劣化故障<a name="ZH-CN_TOPIC_0000002479386488"></a>

#### 使用7.1.RC1及以上版本TaskD<a name="ZH-CN_TOPIC_0000002511346475"></a>

MindCluster集群调度组件结合MindStudio提供的profiling能力，对集群中的性能劣化故障（慢节点）提供诊断功能。该功能提供动态打点和打点数据持久化功能、可动态启停训练任务打点功能，无需重启任务进行诊断，对训练无损耗。

当前支持的打点数据如[表1](#zh-cn_topic_0000002194466236_table5530103025919)所示。

**表 1**  打点数据说明

<a name="zh-cn_topic_0000002194466236_table5530103025919"></a>

|打点数据的类型|支持的AI框架|提供支持的组件|
|--|--|--|
|<p>FP</p><p>（标识前向传播数据）</p>|<p>PyTorch</p><p>仅支持单算子场景。</p>|mstx_torch_plugin|
|<p>Step</p><p>（标识Step时延）</p>|PyTorch、MindSpore|<ul><li>PyTorch<ul><li>原生优化器场景：若torch_npu为7.1.RC1版本，需使用mstx_torch_plugin；若torch_npu为7.1.RC1以上版本，无需使用mstx_torch_plugin，torch_npu自带Step打点。</li><li>自定义优化器场景：手动增加打点数据。</li></ul></li><li>MindSpore<ul><li>MindFormers场景：Step打点数据由MindFormers提供。</li><li>MindSpeed场景：不提供Step打点数据。</li></ul></li></ul>|
|<p>Communication</p><p>（标识通信算子）</p>|PyTorch、MindSpore|<ul><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架</li></ul>|
|<p>SaveCheckpoint</p><p>（标识SaveCheckpoint耗时）</p>|PyTorch、MindSpore|<ul><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架</li></ul>|
|<p>DataLoader</p><p>（标识DataLoader耗时）</p>|PyTorch、MindSpore|<ul><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架</li></ul>|

**使用约束<a name="zh-cn_topic_0000002194466236_section487603614"></a>**

- 当前Step、SaveCheckpoint、FP、DataLoader仅支持同步开启。如需关闭以上四类打点数据，需同时关闭Communication。
- Communication通信算子数据支持单独开启、关闭。
- 动态轻量打点功能与MindStudio的全量打点功能不可同时开启，开启全量打点功能会导致性能劣化故障不能正常采集数据。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- （可选）已安装[ClusterD](../../installation_guide/03_installation.md#clusterd)、[Ascend Device Plugin](../../installation_guide/03_installation.md#ascend-device-plugin)和[Volcano](../../installation_guide/03_installation.md#volcano)（以上MindCluster组件版本均需与TaskD配套）
- 在容器内安装[torch\_npu](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（**可选**，PyTorch场景需安装、版本号≥7.1.RC1）、MindSpore（**可选**，MindSpore场景需安装、版本号≥2.7.0）、[CANN](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（**必选**，版本号≥8.2.RC1）、[TaskD](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（**必选**）

**准备软件包<a name="zh-cn_topic_0000002194466236_section8281518121516"></a>**

**表 2**  准备软件包

<a name="zh-cn_topic_0000002194466236_table232305471415"></a>

|软件包|是否必选|说明|获取方法|使用场景|
|--|--|--|--|--|
|mstx_torch_plugin|否|<p>Ascend PyTorch Profiler中的[采集并解析msproftx数据](https://www.hiascend.com/document/detail/zh/canncommercial/800/devaids/devtools/profiling/atlasprofiling_16_0033.html)功能已经内置了通信算子的打点。为了方便用户在不修改业务代码的基础上获取更多关键阶段的耗时数据，mstx_torch_plugin在Ascend PyTorch Profiler内置了dataloader、forward、step、save_checkpoint这四个关键阶段函数的打点。</p><ul><li>如需使用FP打点数据，需安装mstx_torch_plugin。其他场景下无需安装。</li><li>需使用1.0及以上版本的mstx_torch_plugin。</li></ul>|[获取链接](https://gitee.com/link?target=https://ptdbg.obs.myhuaweicloud.com/profiler/example/1.0/mstx_torch_plugin-1.0-py3-none-any.whl)|PyTorch|

**配置性能劣化故障检测<a name="section1831691464111"></a>**

本方案仅针对7.1.RC1及以上版本的TaskD组件。如使用7.1.RC1以下版本的组件请参见[使用其他版本TaskD](#使用其他版本taskd)章节进行操作。

- **PyTorch场景**

  1. 以下两种方式请根据实际需要进行二选一。
      - 在容器内安装mstx\_torch\_plugin。
          1. 下载mstx\_torch\_plugin的whl包。whl包链接：[mstx\_torch\_plugin](https://gitee.com/link?target=https%3A%2F%2Fptdbg.obs.myhuaweicloud.com%2Fprofiler%2Fexample%2F1.0%2Fmstx_torch_plugin-1.0-py3-none-any.whl)。
          2. 安装软件包。

              ```shell
              pip install mstx_torch_plugin-1.0-py3-none-any.whl
              ```

          3. 在AI任务执行脚本中import导入该whl包。

              需保证import的顺序在import torch和import torch\_npu后面，示例如下。

              ```shell
              import torch 
              import torch_npu  
              import mstx_torch_plugin
              ```

      - 非原生优化器或不使用mstx\_torch\_plugin的情况下，为获取训练的Step耗时数据需修改训练脚本中的训练迭代循环，需增加Step打点代码。

          以下示例为PyTorch-MindSpeed场景，需修改./mindspeed\_llm/training/training.py文件，增加如下加粗字段。

          <pre codetype="Python">
          def train(forward_step_func, model, optimizer, opt_param_scheduler,
                    train_data_iterator, valid_data_iterator,
                    process_non_loss_data_func, config):
                      # Cache into one-logger for callback
              ……
              ……
              if is_profile_enabled():
                  prof = get_profiler()
                  prof.start()
              <strong>step_id = iteration</strong>
              while iteration < args.train_iters:
                  <strong>stream = torch.npu.current_stream()      # 获取当前环境的执行流，用于获取NPU侧时间</strong>
                  <strong>range_id = torch.npu.mstx.range_start(f"step {step_id}", stream) # 标识当前训练step的开始</strong>
                  ……
                  ……
                  if args.manual_gc:
                      if args.manual_gc_interval != 0 and iteration % args.manual_gc_interval == 0:
                          gc.collect()
          
                  if is_profile_enabled():
                      prof.step()
                  <strong>step_id +=1  # 训练step加一，用于标识下一step</strong>
                  <strong>torch.npu.mstx.range_end(range_id) # 标识当前训练step的结束</strong></pre>

  2. 在容器内，以CANN软件包的运行用户登录环境，执行source \$\{install\_path\}/set\_env.sh命令设置环境变量。其中\$\{install\_path\}为CANN软件的安装目录。示例如下。

      ```shell
      source /usr/local/Ascend/cann/set_env.sh
      ```

  3. 训练启动前，在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。

      ```shell
      export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
      ```

      - libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。

      - libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。

          TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。

          ```shell
          pip show taskd
          ```

  4. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在管理进程中拉起TaskD Proxy，在训练进程内部拉起TaskD  Worker。
      1. <a name="li399811541"></a>（可选）拉起TaskD  Manager和TaskD  Proxy。若通过gRPC接口方式开启轻量profiling获取落盘数据，则需执行如下步骤；若通过ConfigMap方式开启轻量profiling获取落盘数据，则跳过该步骤。
          1. 创建manager.py文件，放在调用训练脚本时的当前目录下。manager.py文件内容如下所示。

              ```Python
              from taskd.api import init_taskd_manager, start_taskd_manager
              import os
              
              job_id=os.getenv("MINDX_TASK_ID")
              node_nums=XX         # 用户填入任务节点总数
              proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
              
              init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
              start_taskd_manager()
              ```

              >[!NOTE] 
              >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

          2. 在训练脚本中增加以下代码，拉起TaskD Manager和TaskD Proxy。

              ```Python
              sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
              
              if [[ "${RANK}" -eq 0 ]]; then
                  export MASTER_ADDR=${POD_IP}
                  python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &      # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
              fi
                  
              torchrun ...
              ```

      2. <a name="li23023"></a>拉起TaskD Worker。

         以下示例为PyTorch-MindSpeed场景，需修改QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py文件，在代码中增加如下加粗字段。

          <pre codetype="Python">
          def pretrain(train_valid_test_dataset_provider,
                        model_provider,
                        model_type,
                        forward_step_func,
                        process_non_loss_data_func=None,
                        extra_args_provider=None,
                        args_defaults={}):
              print_rank_0('time to initialize megatron (seconds): {:.3f}'.format(
                  time.time() - _TRAIN_START_TIME))
              print_datetime('after megatron is initialized')
              <strong>import torch.distributed as dist</strong>
              <strong>if dist.is_initialized():</strong>
                  <strong>rank = dist.get_rank()</strong>
                  <strong>from taskd.api.taskd_worker_api import init_taskd_worker</strong>
                  <strong>from taskd.api.taskd_worker_api import start_taskd_worker</strong>
                  <strong>init_taskd_worker(rank,5000)</strong>
                  <strong>start_taskd_worker()</strong>
              app_metrics['app_model_init_finish_time'] = one_logger_utils.get_timestamp_in_ms()
              one_logger_utils.on_pretrain_start()</pre>

         >[!NOTE] 
         >以上代码init_taskd_worker(rank,5000)中的入参5000为/user/cluster-info/profiling的上限大小，详细说明请参见[def init\_taskd\_worker\(rank\_id: int, upper\_limit\_of\_disk\_in\_mb: int = 5000, framework: str = "pt"\) -\> bool](../../api/taskd/01_taskd_worker_apis.md#def-init_taskd_workerrank_id-int-upper_limit_of_disk_in_mb-int--5000-framework-str--pt---bool)中“upper\_limit\_of\_disk\_in\_mb”参数。

  5. <a name="li5236yaml"></a>修改任务YAML。
      1. 修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

          ```Yaml
          ...
                spec:
          ...
                  containers:
          ...
                    ports:                          
                    - containerPort: 9601              
                      name: taskd-port
          ...       
          ```

      2. 挂载文件。
          1. 挂载轻量profiling配置文件：需将宿主机上任务对应的data-trace ConfigMap落盘到/user/cluster-info/datatrace-config/命名空间.data-trace-任务名称/文件夹下。将名为profilingSwitch的文件挂载到容器指定路径：/user/cluster-info/datatrace-config/。
          2. 挂载轻量profiling落盘文件：轻量profiling数据写在容器内的/user/cluster-info/profiling路径下。如需在宿主机获取，请修改任务YAML，将该路径挂出。
              - 容器内YAML挂载示例如下。

                  ```Yaml
                  volumeMounts:
                  - name: profilingdata
                    mountPath: /user/cluster-info/
                  - name: profileswitch
                    mountPath: /user/cluster-info/datatrace-config
                  ```

              - 宿主机内YAML挂载示例如下。

                  ```Yaml
                  volumes:
                  - name: profileswitch
                    hostPath:
                      path: /user/cluster-info/datatrace-config/default.data-trace-default-test-pytorch-fault-mixtral
                  - name: profilingdata
                    hostPath:
                      path: /home/profilingdatapath
                  ```

  6. <a name="li52986profiling"></a>开启轻量profiling获取落盘数据。支持如下两种方式：
      - 修改ClusterD提供的gRPC接口：若配置了[4.1](#li399811541)，需要使用该方式开启。详细接口信息请参见[ModifyTrainingDataTraceSwitch](../../api/clusterd/04_performance_degradation_apis.md#modifytrainingdatatraceswitch)。

          >[!NOTE] 
          >通过ClusterD提供的gRPC接口开启或修改轻量profiling获取落盘数据，创建的data-trace-<任务名称\> ConfigMap的生命周期会随着任务的删除而删除。当任务不存在时，该接口会调用失败。

      - 修改任务对应的data-trace ConfigMap：若未配置[4.1](#li399811541)，需要使用该方式开启。具体操作步骤如下：

          以default命名空间下的名为default-test-pytorch-fault-mixtral的任务为例，以编辑ConfigMap的方式开启轻量profiling获取落盘数据，示例如下。

          1. 在master节点执行以下命令查询该任务对应的配置ConfigMap。

              ```shell
              kubectl get cm
              ```

              - 如果data-trace-default-test-pytorch-fault-mixtral cm已经存在，执行[步骤3](#zh-cn_topic_0000002194466236_li4751182133418)编辑该文件。

                  回显示例如下。

                  ```ColdFusion
                  NAME                                              DATA   AGE
                  data-trace-default-test-pytorch-fault-mixtral     1      18h
                  ```

              - 如果data-trace-default-test-pytorch-fault-mixtral cm不存在，执行[步骤2](#zh-cn_topic_0000002194466236_li1633768104412)创建该文件。

          2. <a name="zh-cn_topic_0000002194466236_li1633768104412"></a>执行以下命令，创建配置轻量profiling获取落盘数据所需ConfigMap文件。
              1. 将以下内容写入datacm.yaml。

                  ```Yaml
                  apiVersion: v1
                  kind: ConfigMap
                  metadata:
                    name: data-trace-default-test-pytorch-fault-mixtral  # cm的名字需以data-trace为前缀+任务名     
                    labels:
                      reset: "true"
                  data:
                    profilingSwitch: '{"CommunicationOperator":"off","Step":"on","SaveCheckpoint":"on","FP":"on","DataLoader":"on"}'
                  ```

              2. 在master节点执行以下命令，创建ConfigMap。

                  ```shell
                  kubectl apply -f datacm.yaml
                  ```

                  回显如下所示，表示ConfigMap创建成功。

                  ```ColdFusion 
                  configmap/data-trace-default-test-pytorch-fault-mixtral created
                  ```

          3. <a name="zh-cn_topic_0000002194466236_li4751182133418"></a>执行以下命令编辑ConfigMap文件。

              ```shell
              kubectl edit cm data-trace-default-test-pytorch-fault-mixtral
              ```

          4. 如需开启通信算子，请将CommunicationOperator字段的取值改为“on”。

              ```Yaml
              apiVersion: v1
              data:
                profilingSwitch: '{"CommunicationOperator":"on","Step":"on","SaveCheckpoint":"on","FP":"on","DataLoader":"on"}'
              ```

              >[!NOTE] 
              >开启通信算子后可能造成训练性能下降，不建议常态开启通信算子。

          5. 按“Esc”键，输入:wq!保存并退出。

- **MindSpore场景**

  1. 在容器内，以CANN软件包的运行用户登录环境，执行source \$\{install\_path\}/set\_env.sh命令设置环境变量。其中\$\{install\_path\}为CANN软件的安装目录。示例如下。

      ```shell
      source /usr/local/Ascend/cann/set_env.sh
      ```

  2. 训练启动前，在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。

      ```shell
      export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
      ```

      - libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。

      - libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。

          TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。

          ```shell
          pip show taskd
          ```

  3. 在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练脚本中拉起TaskD  Manager，在管理进程中拉起TaskD Proxy，在训练进程内部拉起TaskD  Worker。
      1. <a name="li399811541"></a>（可选）拉起TaskD  Manager和TaskD  Proxy。若通过gRPC接口方式开启轻量profiling获取落盘数据，则需执行如下步骤；若通过ConfigMap方式开启轻量profiling获取落盘数据，则跳过该步骤。

          1. 创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

              ```Python
              from taskd.api import init_taskd_manager, start_taskd_manager
              import os
              
              job_id=os.getenv("MINDX_TASK_ID")
              node_nums=XX         # 用户填入任务节点总数
              proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
              
              init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
              start_taskd_manager()
              ```

              >[!NOTE] 
              >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../../api/taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool)。

          2. 在训练脚本中增加以下代码拉起TaskD  Manager。

              ```Python
              if [[ "${MS_SCHED_HOST}" -eq "${POD_IP}" ]]; then
                  python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &       # manager.py具体执行路径由当前路径决定，error.log日志路径需提前创建
              fi
                  
              msrun ...
              ```

          3. 修改mindspore/python/mindspore/parallel/cluster/process\_entity/\_api.py文件，拉起TaskD  Proxy。示例如下。

              <pre codetype="Python">
              ...
                if ("TTP:1" in tft_env) or ("UCE:1" in tft_env) or ("ARF:1" in tft_env):
                          try:
                              from taskd.python.framework.agent.ms_mgr.msrun_plugin import MSRunPlugin
                              <strong>from taskd.api.taskd_proxy_api import init_taskd_proxy</strong>
                              <strong>from taskd.python.framework.common.type import CONFIG_UPSTREAMIP_KEY, LOCAL_HOST</strong>
                              <strong>import threading</strong>
                              <strong>proxy = threading.Thread(target=init_taskd_proxy, args=({CONFIG_UPSTREAMIP_KEY : os.getenv("MS_SCHED_HOST", LOCAL_HOST)},))</strong>
                              <strong>proxy.daemon = True</strong>
                              <strong>proxy.start()</strong>
                              self.msmgr = MSRunPlugin()
                              self.msmgr.register_callbacks("KILL_WORKER", self.kill_workers)
                              self.msmgr.register_callbacks("START_ALL_WORKER", self.start_all_workers)
                              self.msmgr.register_callbacks("START_WORKER_LIST", self.start_worker_list)
                              self.msmgr.register_callbacks("MONITOR", self.monitor_rank_status)
                              self.enable_mindx = True
                              os.environ["MS_ENABLE_RECOVERY"] = str(1)
              ...</pre>

      2. <a name="li2302301"></a>拉起TaskD  Worker。

          以下示例为MindSpore-MindFormers场景，需修改./mindformers/trainer/base\_trainer.py文件，在代码中增加如下加粗字段。

          <pre codetype="Python">
              def training_process(
                      self,
                      config: Optional[Union[dict, MindFormerConfig, ConfigArguments, TrainingArguments]] = None,
                      network: Optional[Union[Cell, PreTrainedModel]] = None,
                      dataset: Optional[Union[BaseDataset, GeneratorDataset]] = None,
                      optimizer: Optional[Optimizer] = None,
                      callbacks: Optional[Union[Callback, List[Callback]]] = None,
                      compute_metrics: Optional[Union[dict, set]] = None,
                      **kwargs):
                  ……
                  ……
          
                  logger.info(".........Starting Training Model..........")
                  if get_real_rank() % 8 == 0:
                      pprint(config)
                  logger.info(".........Model Compiling, Please Wait a Moment...........")
                  <strong>try:</strong>
                      <strong>rank = get_rank()</strong>
                      <strong>from taskd.api.taskd_worker_api import init_taskd_worker</strong>
                      <strong>from taskd.api.taskd_worker_api import start_taskd_worker</strong>
                      <strong>init_taskd_worker(rank,5000)</strong>
                      <strong>start_taskd_worker()</strong>
                  <strong>except Exception as e:</strong>
                      <strong>print("failed to call mindcluster taskd")</strong>
                  model.train(config.runner_config.epochs, dataset,
                              callbacks=callbacks,
                              dataset_sink_mode=config.runner_config.sink_mode,
                              sink_size=config.runner_config.sink_size,
                              initial_epoch=config.runner_config.initial_epoch)</pre>

          >[!NOTE] 
          >以上代码init_taskd_worker(rank,5000)中的入参5000为/user/cluster-info/profiling的上限大小，详细说明请参见[def init\_taskd\_worker\(rank\_id: int, upper\_limit\_of\_disk\_in\_mb: int = 5000, framework: str = "pt"\) -\> bool](../../api/taskd/01_taskd_worker_apis.md#def-init_taskd_workerrank_id-int-upper_limit_of_disk_in_mb-int--5000-framework-str--pt---bool)中“upper\_limit\_of\_disk\_in\_mb”参数。

  4. 修改任务YAML。详细请参见[PyTorch场景的步骤5](#li5236yaml)。
  5. 开启轻量profiling获取落盘数据。详细请参见[PyTorch场景的步骤6](#li52986profiling)。

**获取性能劣化故障检测数据<a name="zh-cn_topic_0000002194466236_section435518556912"></a>**

- 落盘数据按rank进行分类，轻量profiling数据写在容器内的/user/cluster-info/profiling路径。
- 对于存在环境变量[MINDX\_TASK\_ID](../../api/environment_variable_description.md)的Pod，rank 0数据在容器内的路径为/user/cluster-info/profiling/$MINDX\_TASK\_ID/0。

    >[!NOTE] 
    >- 如无该环境变量，默认会落盘到名为default\_task\_id\_<i>时间戳</i>的文件夹内。
    >- /user/cluster-info/profiling达到配置的上限大小（PyTorch场景参考[4.2](#li23023)；MindSpore场景参考[3.2](#li2302301)）后，将进行文件老化，默认每次删除修改时间最早的20%个文件。老化过程中仅删除profiling目录下rank文件夹中的以数字命名的文件，建议不手动添加其他文件到profiling文件夹下。如果用户手动添加其他文件，TaskD不会将该文件删除，但该文件会占用空间。
    >- 轻量profiling文件以时间戳命名，各条记录以换行分割，每次追加写入rank下最新文件。最新文件大小超过10MB时，TaskD会新建profiling文件。如果使用NFS等网络存储方式，当数据同步较慢时，可能存在文件大小未达到10MB即创建新文件的情况。

#### 使用其他版本TaskD<a name="ZH-CN_TOPIC_0000002511346483"></a>

MindCluster集群调度组件结合MindStudio提供的profiling能力，对集群中的性能劣化故障（慢节点）提供诊断功能。该功能提供动态打点和打点数据持久化功能、可动态启停训练任务打点功能，无需重启任务进行诊断，对训练无损耗。

当前支持的打点数据如[表1](#zh-cn_topic_0000002194466236_table5530103025919)所示。

**表 1**  打点数据说明

<a name="zh-cn_topic_0000002194466236_table5530103025919"></a>

|打点数据的类型|支持的AI框架|提供支持的组件|
|--|--|--|
|<p>FP</p><p>（标识前向传播数据）</p>|<p>PyTorch</p><p>仅支持单算子场景。</p>|mstx_torch_plugin|
|<p>Step</p><p>（标识Step时延）</p>|PyTorch、MindSpore|<ul><li>PyTorch<ul><li>原生优化器场景：若torch_npu为7.1.RC1及以下版本，需使用mstx_torch_plugin；若torch_npu为7.1.RC1以上版本，无需使用mstx_torch_plugin，torch_npu自带Step打点。</li><li>自定义优化器场景：手动增加打点数据。</li></ul></li><li>MindSpore<ul><li>MindFormers场景：Step打点数据由MindFormers提供。</li><li>MindSpeed场景：不提供Step打点数据。</li></ul></li></ul>|
|<p>Communication</p><p>（标识通信算子）</p>|PyTorch、MindSpore|<ul><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架</li></ul>|
|<p>SaveCheckpoint</p><p>（标识SaveCheckpoint耗时）</p>|PyTorch、MindSpore|<ul><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架</li></ul>|
|<p>DataLoader</p><p>（标识DataLoader耗时）</p>|PyTorch、MindSpore|<ul><li>PyTorch：torch_npu</li><li>MindSpore：MindSpore框架</li></ul>|

**使用约束<a name="zh-cn_topic_0000002194466236_section487603614"></a>**

- 当前Step、SaveCheckpoint、FP、DataLoader仅支持同步开启。如需关闭以上四类打点数据，需同时关闭Communication。
- Communication通信算子数据支持单独开启、关闭。
- 动态轻量打点功能与MindStudio的全量打点功能不可同时开启，开启全量打点功能会导致性能劣化故障不能正常采集数据。

**前提条件<a name="zh-cn_topic_0000002194466236_section138036504533"></a>**

- （可选）已安装[ClusterD](../../installation_guide/03_installation.md#clusterd)、[Ascend Device Plugin](../../installation_guide/03_installation.md#ascend-device-plugin)和[Volcano](../../installation_guide/03_installation.md#volcano)（以上MindCluster组件版本均需与TaskD配套）
- 在容器内安装[torch\_npu](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（**可选**，PyTorch场景需安装、版本号≥7.0.0）、MindSpore（**可选**，MindSpore场景需安装、版本号≥2.6.RC1）、[CANN](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（**必选**，版本号≥8.1.RC1）、[TaskD](./07_using_resumable_training_on_the_cli.md#制作mindspeed-llm训练镜像pytorch框架)（**必选**，版本号≥7.0.RC1）

**准备软件包<a name="zh-cn_topic_0000002194466236_section8281518121516"></a>**

**表 2**  准备软件包

<a name="zh-cn_topic_0000002194466236_table232305471415"></a>

|软件包|是否必选|说明|获取方法|使用场景|
|--|--|--|--|--|
|mstx_torch_plugin|否|<p>Ascend PyTorch Profiler中的[采集并解析msproftx数据](https://www.hiascend.com/document/detail/zh/canncommercial/800/devaids/devtools/profiling/atlasprofiling_16_0033.html)功能已经内置了通信算子的打点。为了方便用户在不修改业务代码的基础上获取更多关键阶段的耗时数据，mstx_torch_plugin在Ascend PyTorch Profiler内置了dataloader、forward、step、save_checkpoint这四个关键阶段函数的打点。</p><ul><li>如需使用FP打点数据，需安装mstx_torch_plugin。其他场景下无需安装。</li><li>需使用1.0及以上版本的mstx_torch_plugin。</li></ul>|[获取链接](https://gitee.com/link?target=https://ptdbg.obs.myhuaweicloud.com/profiler/example/1.0/mstx_torch_plugin-1.0-py3-none-any.whl)|PyTorch|

**配置性能劣化故障检测<a name="section167141313174510"></a>**

本方案仅针对7.1.RC1以下版本的TaskD组件。如使用7.1.RC1及以上版本的组件请参见[使用7.1.RC1及以上版本TaskD](#使用71rc1及以上版本taskd)章节。

- **PyTorch场景**

  1. （可选）在容器内安装mstx\_torch\_plugin。
      1. 下载mstx\_torch\_plugin的whl包。whl包链接：[mstx\_torch\_plugin](https://gitee.com/link?target=https%3A%2F%2Fptdbg.obs.myhuaweicloud.com%2Fprofiler%2Fexample%2F1.0%2Fmstx_torch_plugin-1.0-py3-none-any.whl)。
      2. 安装软件包。

          ```shell
          pip install mstx_torch_plugin-1.0-py3-none-any.whl
          ```

      3. 在AI任务执行脚本中import导入该whl包。

          需保证import的顺序在import torch和import torch\_npu后面，示例如下。

          ```shell
          import torch 
          import torch_npu  
          import mstx_torch_plugin
          ```

  2. （可选）非原生优化器或不使用mstx\_torch\_plugin的情况下，为获取训练的Step耗时数据需修改训练脚本中的训练迭代循环，需增加Step打点代码。

      以下示例为PyTorch-MindSpeed场景，需修改./mindspeed\_llm/training/training.py文件，增加如下加粗字段。

      <pre codetype="Python">
      def train(forward_step_func, model, optimizer, opt_param_scheduler,
                train_data_iterator, valid_data_iterator,
                process_non_loss_data_func, config):
                  # Cache into one-logger for callback
          ……
          ……
          if is_profile_enabled():
              prof = get_profiler()
              prof.start()
          <strong>step_id = iteration</strong>
          while iteration < args.train_iters:
             <strong>stream = torch.npu.current_stream()      # 获取当前环境的执行流，用于获取NPU侧时间</strong>
              <strong>range_id = torch.npu.mstx.range_start(f"step {step_id}", stream) # 标识当前训练step的开始</strong>
              ……
              ……
              if args.manual_gc:
                  if args.manual_gc_interval != 0 and iteration % args.manual_gc_interval == 0:
                      gc.collect()
      
              if is_profile_enabled():
                  prof.step()
              <strong>step_id +=1  # 训练step加一，用于标识下一step</strong>
              <strong>torch.npu.mstx.range_end(range_id) # 标识当前训练step的结束</strong></pre>

  3. 在容器内，以CANN软件包的运行用户登录环境，执行source \$\{install\_path\}/set\_env.sh命令设置环境变量。其中\$\{install\_path\}为CANN软件的安装目录。示例如下。

      ```shell
      source /usr/local/Ascend/cann/set_env.sh
      ```

  4. 训练启动前，在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。

      ```shell
      export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
      ```

      - libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。

      - libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。

          TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。

          ```shell
          pip show taskd
          ```

  5. <a name="li230238965"></a>在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练进程内部拉起TaskD  Worker。

      以下示例为PyTorch-MindSpeed场景，需修改QWEN3\_for\_PyTorch\_2.7\_code/mindspeed\_llm/training/training.py文件，在代码中增加如下加粗字段。

        <pre codetype="Python">
        def pretrain(train_valid_test_dataset_provider,
                      model_provider,
                      model_type,
                      forward_step_func,
                      process_non_loss_data_func=None,
                      extra_args_provider=None,
                      args_defaults={}):
            print_rank_0('time to initialize megatron (seconds): {:.3f}'.format(
                time.time() - _TRAIN_START_TIME))
            print_datetime('after megatron is initialized')
            <strong>import torch.distributed as dist</strong>
            <strong>if dist.is_initialized():</strong>
                <strong>rank = dist.get_rank()</strong>
                <strong>from taskd.api.taskd_worker_api import init_taskd_worker</strong>
                <strong>from taskd.api.taskd_worker_api import start_taskd_worker</strong>
                <strong>init_taskd_worker(rank,5000)</strong>
                <strong>start_taskd_worker()</strong>
            app_metrics['app_model_init_finish_time'] = one_logger_utils.get_timestamp_in_ms()
            one_logger_utils.on_pretrain_start()</pre>

        >[!NOTE] 
        >以上代码init_taskd_worker(rank,5000)中的入参5000为/user/cluster-info/profiling的上限大小，详细说明请参见[def init\_taskd\_worker\(rank\_id: int, upper\_limit\_of\_disk\_in\_mb: int = 5000, framework: str = "pt"\) -\> bool](../../api/taskd/01_taskd_worker_apis.md#def-init_taskd_workerrank_id-int-upper_limit_of_disk_in_mb-int--5000-framework-str--pt---bool)中“upper\_limit\_of\_disk\_in\_mb”参数。

  6. <a name="li5236890yaml"></a>修改任务YAML。
      1. 挂载轻量profiling配置文件：需将宿主机上任务对应的data-trace ConfigMap落盘到/user/cluster-info/datatrace-config/命名空间.data-trace-任务名称/文件夹下。将名为profilingSwitch的文件挂载到容器指定路径：/user/cluster-info/datatrace-config/。
      2. 挂载轻量profiling落盘文件：轻量profiling数据写在容器内的/user/cluster-info/profiling路径下。如需在宿主机获取，请修改任务YAML，将该路径挂出。
          - 容器内YAML挂载示例如下。

              ```Yaml
              volumeMounts:
              - name: profilingdata
                mountPath: /user/cluster-info/
              - name: profileswitch
                mountPath: /user/cluster-info/datatrace-config
              ```

          - 宿主机内YAML挂载示例如下。

              ```Yaml
              volumes:
              - name: profileswitch
                hostPath:
                  path: /user/cluster-info/datatrace-config/default.data-trace-default-test-pytorch-fault-mixtral
              - name: profilingdata
                hostPath:
                  path: /home/profilingdatapath
              ```

  7. <a name="li52986890profiling"></a>开启轻量profiling获取落盘数据。修改任务对应的data-trace ConfigMap或ClusterD提供的gRPC接口，接口信息见[ModifyTrainingDataTraceSwitch](../../api/clusterd/04_performance_degradation_apis.md#modifytrainingdatatraceswitch)，动态开启或关闭轻量profiling能力。

      以default命名空间下的名为default-test-pytorch-fault-mixtral的任务为例，以编辑ConfigMap的方式开启轻量profiling获取落盘数据，示例如下。

      1. 在master节点执行以下命令查询该任务对应的配置ConfigMap。

          ```shell
          kubectl get cm
          ```

          - 如果data-trace-default-test-pytorch-fault-mixtral cm已经存在，执行[步骤3](#zh-cn_topic_0000002194466236_li47511821334189)编辑该文件。

              回显示例如下。

              ```ColdFusion
              NAME                                              DATA   AGE
              data-trace-default-test-pytorch-fault-mixtral     1      18h
              ```

          - 如果data-trace-default-test-pytorch-fault-mixtral cm不存在，执行[步骤2](#zh-cn_topic_0000002194466236_li16337681044126)创建该文件。

      2. <a name="zh-cn_topic_0000002194466236_li16337681044126"></a>执行以下命令，创建配置轻量profiling获取落盘数据所需ConfigMap文件。
          1. 将以下内容写入datacm.yaml。

              ```Yaml
              apiVersion: v1
              kind: ConfigMap
              metadata:
                name: data-trace-default-test-pytorch-fault-mixtral  # cm的名字需以data-trace为前缀+任务名     
                labels:
                  reset: "true"
              data:
                profilingSwitch: '{"CommunicationOperator":"off","Step":"on","SaveCheckpoint":"on","FP":"on","DataLoader":"on"}'
              ```

          2. 在master节点执行以下命令，创建ConfigMap。

              ```shell
              kubectl apply -f datacm.yaml
              ```

              回显如下所示，表示ConfigMap创建成功。

              ```ColdFusion
              configmap/data-trace-default-test-pytorch-fault-mixtral created
              ```

      3. <a name="zh-cn_topic_0000002194466236_li47511821334189"></a>执行以下命令编辑ConfigMap文件。

          ```shell
          kubectl edit cm data-trace-default-test-pytorch-fault-mixtral
          ```

      4. 如需开启通信算子，请将CommunicationOperator字段的取值改为“on”。

          ```Yaml
          apiVersion: v1
          data:
            profilingSwitch: '{"CommunicationOperator":"on","Step":"on","SaveCheckpoint":"on","FP":"on","DataLoader":"on"}'
          ```

          >[!NOTE] 
          >开启通信算子后可能造成训练性能下降，不建议常态开启通信算子。

      5. 按“Esc”键，输入:wq!保存并退出。

- **MindSpore场景**

  1. 在容器内，以CANN软件包的运行用户登录环境，执行source \$\{install\_path\}/set\_env.sh命令设置环境变量。其中\$\{install\_path\}为CANN软件的安装目录。示例如下。

      ```shell
      source /usr/local/Ascend/cann/set_env.sh
      ```

  2. 训练启动前，在训练脚本中导入LD\_PRELOAD环境变量。该环境变量允许系统提前加载指定的so文件。示例如下。

      ```shell
      export LD_PRELOAD=/usr/local/Ascend/cann/lib64/libmspti.so:/usr/local/python3.10.5/lib/python3.10/site-packages/taskd/python/cython_api/libs/libtaskd.so
      ```

      - libmspti.so：该so由MindStudio提供，集成在CANN包内。当使用默认安装路径时，路径为：/usr/local/Ascend/cann/lib64/libmspti.so。

      - libtaskd.so：该so由TaskD组件提供，安装该whl包后，路径为：TaskD所在路径/taskd/python/cython\_api/libs/libtaskd.so。

          TaskD所在路径可通过以下命令进行查询。回显中的Location字段即为TaskD所在路径。

          ```shell
          pip show taskd
          ```

  3. <a name="li23023896501"></a>在分布式环境初始化完成，能够获取到全局rank之后，修改训练脚本，在训练进程内部拉起TaskD  Worker。

      以下示例为MindSpore-MindFormers场景，需修改./mindformers/trainer/base\_trainer.py文件，在代码中增加如下加粗字段。

        <pre codetype="Python">
            def training_process(
                    self,
                    config: Optional[Union[dict, MindFormerConfig, ConfigArguments, TrainingArguments]] = None,
                    network: Optional[Union[Cell, PreTrainedModel]] = None,
                    dataset: Optional[Union[BaseDataset, GeneratorDataset]] = None,
                    optimizer: Optional[Optimizer] = None,
                    callbacks: Optional[Union[Callback, List[Callback]]] = None,
                    compute_metrics: Optional[Union[dict, set]] = None,
                    **kwargs):
                ……
                ……
        
                logger.info(".........Starting Training Model..........")
                if get_real_rank() % 8 == 0:
                    pprint(config)
                logger.info(".........Model Compiling, Please Wait a Moment...........")
                <strong>try:</strong>
                    <strong>rank = get_rank()</strong>
                    <strong>from taskd.api.taskd_worker_api import init_taskd_worker</strong>
                    <strong>from taskd.api.taskd_worker_api import start_taskd_worker</strong>
                    <strong>init_taskd_worker(rank,5000)</strong>
                    <strong>start_taskd_worker()</strong>
                <strong>except Exception as e:</strong>
                    <strong>print("failed to call mindcluster taskd")</strong>
                model.train(config.runner_config.epochs, dataset,
                            callbacks=callbacks,
                            dataset_sink_mode=config.runner_config.sink_mode,
                            sink_size=config.runner_config.sink_size,
                            initial_epoch=config.runner_config.initial_epoch)</pre>

        >[!NOTE] 
        >以上代码init_taskd_worker(rank,5000)中的入参5000为/user/cluster-info/profiling的上限大小，详细说明请参见[def init\_taskd\_worker\(rank\_id: int, upper\_limit\_of\_disk\_in\_mb: int = 5000, framework: str = "pt"\) -\> bool](../../api/taskd/01_taskd_worker_apis.md#def-init_taskd_workerrank_id-int-upper_limit_of_disk_in_mb-int--5000-framework-str--pt---bool)中“upper\_limit\_of\_disk\_in\_mb”参数。

  4. 修改任务YAML。详细请参见[PyTorch场景的步骤6](#li5236890yaml)。
  5. 开启轻量profiling获取落盘数据。详细请参见[PyTorch场景的步骤7]。(#li52986890profiling)。

**获取性能劣化故障检测数据<a name="zh-cn_topic_0000002194466236_section435518556912"></a>**

- 落盘数据按rank进行分类，轻量profiling数据写在容器内的/user/cluster-info/profiling路径。
- 对于存在环境变量[MINDX\_TASK\_ID](../../api/environment_variable_description.md)的Pod，rank 0数据在容器内的路径为/user/cluster-info/profiling/$MINDX\_TASK\_ID/0。

    >[!NOTE] 
    >- 如无该环境变量，默认会落盘到名为default\_task\_id\_<i>时间戳</i>的文件夹内。
    >- /user/cluster-info/profiling达到配置的上限大小（PyTorch场景参考[步骤5](#li230238965)；MindSpore场景参考[步骤3](#li23023896501)）后，将进行文件老化，默认每次删除修改时间最早的20%个文件。老化过程中仅删除profiling目录下rank文件夹中的以数字命名的文件，建议不手动添加其他文件到profiling文件夹下。如果用户手动添加其他文件，TaskD不会将该文件删除，但该文件会占用空间。
    >- 轻量profiling文件以时间戳命名，各条记录以换行分割，每次追加写入rank下最新文件。最新文件大小超过10MB时，TaskD会新建profiling文件。如果使用NFS等网络存储方式，当数据同步较慢时，可能存在文件大小未达到10MB即创建新文件的情况。

### 慢节点&慢网络故障<a name="ZH-CN_TOPIC_0000002511426421"></a>

#### 简介<a name="ZH-CN_TOPIC_0000002532640773"></a>

MindCluster集群调度组件结合MindCluster Ascend FaultDiag（故障诊断工具）提供的在线诊断能力，为集群中的慢节点&慢网络故障提供诊断功能。

**使用前准备<a name="zh-cn_topic_0000002333550505_section420815439315"></a>**

使用慢节点&慢网络故障诊断功能前，需增加NodeD中CPU和内存的资源大小，在NodeD启动YAML文件中更改资源信息。

当前YAML文件内容如下：

```Yaml
resources:
            requests:
              memory: 300Mi
              cpu: 500m
            limits:
              memory: 300Mi
              cpu: 500m
```

修改后YAML文件内容如下：

```Yaml
resources:
            requests:
              memory: 10Gi
              cpu: 5000m
            limits:
              memory: 10Gi
              cpu: 5000m
```

**部署形态<a name="zh-cn_topic_0000002333550505_section1048011118418"></a>**

ClusterD与FD-OL（Fault Diagnose Online）框架在同一进程中，都部署在管理节点。ClusterD启动时将自动拉起FD-OL框架。

#### 慢节点诊断<a name="ZH-CN_TOPIC_0000002500880704"></a>

**功能说明<a name="zh-cn_topic_0000002278667326_section27999216294"></a>**

对于AI集群中出现的节点训练性能劣化现象，提供支持实时检测计算域问题或网络导致的慢节点，以便用户通过切换或其他方式隔离慢节点。

当前仅支持与ClusterD和NodeD集成进行在线部署，请参见[安装部署](../../installation_guide/03_installation.md)章节完成ClusterD和NodeD部署。

- 慢节点算法：基于训练场景关键性能指标，感知实时劣化状态；针对通信算子、计算算子同步关系，实现慢计算卡、慢通信域问题定界。
- 慢节点清洗：对节点内部增量数据转化并清洗，生成清洗结果csv文件。
- 慢节点调度：调度慢节点整体流程，控制数据清洗和慢节点算法。

**使用示例<a name="zh-cn_topic_0000002278667326_section19867823600"></a>**

启动慢节点诊断任务。

1. 为获取并行域信息，需在训练脚本的训练迭代循环中增加获取并行域信息的函数调用。以下示例为PyTorch-MindSpeed场景，需在./mindspeed\_llm/training/training.py文件中增加如下加粗字段。

    <pre codetype="Python">
    def train(forward_step_func, model, optimizer, opt_param_scheduler,
              train_data_iterator, valid_data_iterator,
              process_non_loss_data_func, config):
        ……
        if is_profile_enabled():
            prof = get_profiler()
            prof.start()
        <strong>m_iter = 0</strong>
        while iteration < args.train_iters:
            ……
            args.curr_iteration = iteration
            loss_dict, skipped_iter, grad_norm, num_zeros_in_grad = \
                train_step(forward_step_func,
                           train_data_iterator,
                           model,
                           optimizer,
                           opt_param_scheduler,
                           config)
            iteration += 1
            <strong>m_iter += 1</strong>
            <strong>if m_iter == 5:</strong>
                <strong>from taskd.python.adaptor.pytorch.group_info import dump_group_info</strong>
                <strong>dump_group_info()</strong>
            batch_size = mpu.get_data_parallel_world_size() * \
                         args.micro_batch_size * \
                         get_num_microbatches()</pre>

2. 完成[使用前准备](#zh-cn_topic_0000002333550505_section420815439315)和[部署形态](#zh-cn_topic_0000002333550505_section1048011118418)。
3. 使用**kubectl apply -f ajob-2pod-16npu.yaml**命令，创建慢节点诊断任务写入configMap。

    ![](../../../figures/scheduling/zh-cn_image_0000002333860285.png)

4. ajob-2pod-16npu.yaml内容如下所示，各回显数据说明请见[表1](#zh-cn_topic_0000002278667326_table1834456175114)。

    ![](../../../figures/scheduling/zh-cn_image_0000002509443757.png)

    以下为YAML示例，不可以直接拷贝编译运行，仅供参考。

    ```Yaml
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: ras-feature-slownode-default-test-pytorch-2pod-16npu    # The value of JobName must be the same as the name attribute of the following job. The prefix ras-feature-slownode- cannot be modified.
      namespace: mindx-dl
      labels:
        fd-ol-slow-node: "true"
    data:
      FeatConf: |
        {"jobName":"default-test-pytorch-2pod-16npu","jobNamespace":"default","normalNumber":20,"nSigma":3,"degradationPercentage":0.3,"nConsecAnomaliesSignifySlow":3,"nSecondsDoOneDetection":30,"clusterMeanDistance":1.3,"cardOneNode":16,"SlowNode":1}
    ---
    ```

    **表 1**  YAML文件回显说明

    <a name="zh-cn_topic_0000002278667326_table1834456175114"></a>

    |字段名|默认值|说明|
    |--|--|--|
    |jobNamespace|default|任务所在的namespace。|
    |jobName|-|任务名。|
    |normalNumber|20|计算初始阈值（正常数量）。|
    |nSigma|3个|设置σ的个数以计算其上下界。|
    |degradationPercentage|0.3|阈值，劣化的百分比，0.3表示劣化30%。|
    |nConsecAnomaliesSignifySlow|3次|设置异常次数，连续出现多次异常后才进行检测。|
    |nSecondsDoOneDetection|30秒|设置间隔时长，进行检测，单位为秒。|
    |clusterMeanDistance|1.3|聚类后，两个类别之间的阈值距离（mean1、mean2）。|
    |cardOneNode|16张卡|一个节点的卡片数量。|
    |slowNode|默认为1，开启任务。|<p>是否开启任务。</p><ul><li>1：开启任务。</li><li>0：关闭任务。</li></ul>|

**查询慢节点诊断结果<a name="zh-cn_topic_0000002278667326_section208199121010"></a>**

在创建慢节点任务后，可通过查询ClusterD和NodeD的日志查看其诊断任务详情。

**方式一：通过K8s日志查询集群侧慢节点诊断日志**

1. 通过**kubectl get pods -n mindx-dl**命令，查询启动的ClusterD和NodeD节点数据。

    ![](../../../figures/scheduling/zh-cn_image_0000002477523808.png)

2. 再使用<b>kubectl logs -n mindx-dl clusterd-7d5db546d8-kdslz | grep "got degradation, slow rank"</b>查询日志数据。
3. 若日志中出现如下图所示，则表明出现节点劣化。

    ![](../../../figures/scheduling/zh-cn_image_0000002457147010.png)

**方式二：通过落盘日志查询集群侧慢节点诊断日志**

1. 使用<b>cat /var/log/mindx-dl.clusterd.clusterd.log | grep "got degradation, slow rank"</b>命令查询日志数据。
2. 若日志中出现如下图所示，则表明出现节点劣化。

    ![](../../../figures/scheduling/zh-cn_image_0000002490267057.png)

**方式三：查询节点侧的慢节点诊断日志。**

使用<b>kubectl logs -n mindx-dl node-9ld8k | grep "is degradation"</b>命令进行查询，若日志中出现如下图所示数据，则表明出现节点劣化。

![](../../../figures/scheduling/zh-cn_image_0000002457149146.png)

**已支持的慢节点网络故障<a name="zh-cn_topic_0000002278667326_section10496211245"></a>**

<a name="zh-cn_topic_0000002278667326_table4804164084414"></a>

|故障码|故障说明|故障级别|
|--|--|--|
|110001010|慢节点故障，一次性消息上报。|SubHealthFault：亚健康故障。|
|100001011|故障劣化已恢复。|NotHandleFault：暂不处理故障。|

#### 慢网络诊断<a name="ZH-CN_TOPIC_0000002500720860"></a>

**功能说明<a name="zh-cn_topic_0000002313236861_section27999216294"></a>**

支持提供参数面网络连通性检测，实时进行网络监测和异常上报，辅助故障分析和定界定位，提前预警网络故障和亚健康风险信息，保障集群网络的长期稳定运行。

当前仅支持与ClusterD和NodeD集成进行在线部署，请参见[安装部署](../../installation_guide/03_installation.md)章节完成ClusterD和NodeD部署。

- 慢网络算法：对节点之间的网络拨测数据进行分析、检测，并输出网络诊断结果。
- 慢网络调度：把控探测任务启停，上报故障结果，调度慢网络整体流程。

**使用示例<a name="zh-cn_topic_0000002313236861_section1969604665710"></a>**

1. 配置共享存储。

    ClusterD和NodeD通过共享存储进行交互，两者的共享存储根路径需要保持一致。共享目录的根路径属主为9000用户，与ClusterD运行用户一致。

    1. 配置server。

        ![](../../../figures/scheduling/zh-cn_image_0000002300566136.png)

    2. 修改NodeD配置。

        ![](../../../figures/scheduling/zh-cn_image_0000002384880596.png)

    3. 修改ClusterD配置。

        ![](../../../figures/scheduling/zh-cn_image_0000002385041140.png)

    4. 执行**kubectl get pods -o wide -A**命令出现如下示例，则表示已完成共享存储配置。

        ![](../../../figures/scheduling/zh-cn_image_0000002300409300.png)

2. 开启故障检测开关。
    1. 登录环境，进入NodeD解压目录。
    2. 执行以下命令创建名为pingmesh-config的ConfigMap文件。pingmesh-config.yaml为pingmesh配置文件，可从NodeD安装包中获取。

        ```shell
        kubectl apply -f pingmesh-config.yaml
        ```

        回显示例如下：

        ```ColdFusion
        configmap/pingmesh-config created
        ```

    3. 执行以下命令编辑pingmesh-config文件，该文件中各参数的填写说明如下表所示。

        ```shell
        kubectl edit cm -n cluster-system pingmesh-config
        ```

        **表 1**  pingmesh-config文件参数说明

        <a name="zh-cn_topic_0000002313236861_table15591134151811"></a>

        |参数|取值|说明|
        |--|--|--|
        |app|pingmesh|ConfigMap其中一个label的key。|
        |global|-|集群配置信息。|
        |"1"|超节点ID|超节点ID为1的配置示例，用户可根据实际情况进行修改或新增。当配置了某个超节点后，NodeD会采用超节点的配置信息而忽略global配置信息。|
        |activate|<ul><li>on：开启</li><li>off：关闭</li></ul>|是否启用pingmesh功能。|
        |task_interval|[1~60]|pingmesh任务间隔时间。单位为秒。|

**查看检测结果<a name="zh-cn_topic_0000002313236861_section74321914202214"></a>**

网络检测的pingmesh结果将写入文件<nodename\>.log中，该文件中各字段的详细说明如下表所示。

**表 2**  <nodename\>.log文件参数说明

<a name="zh-cn_topic_0000002313236861_table1485915561131"></a>

|参数|取值|说明|
|--|--|--|
|uid|长度为64的字符串|本次pingmesh任务的ID。|
|config|字符串|本次pingmesh任务的用户配置。|
|physicID|[0~15]|NPU卡物理ID。|
|taskID|<ul><li>节点内部的任务：0</li><li>节点间的任务：1</li></ul>|任务ID。|
|DestNum|[0~47]|本次pingmesh目标地址数量。|
|source_addr|IPv4网络地址|源地址。|
|target_addr|IPv4网络地址|目标地址。|
|suc_pkt_num|-|发送成功的包数量。|
|fail_pkt_num|-|发送失败的包数量。|
|max_time|<ul><li>正常情况：非负值</li><li>ping失败：-1</li></ul>|最长响应时间。|
|min_time|<ul><li>正常情况：非负值</li><li>ping失败：-1</li></ul>|最短响应时间。|
|avg_time|<ul><li>正常情况：非负值</li><li>ping失败：-1</li></ul>|平均响应时间。|
|tp95_time|<ul><li>正常情况：非负值</li><li>ping失败：-1</li></ul>|处于95%位置时的响应时间。|
|reply_stat_num|-|本次查询到的响应数量。|
|ping_total_num|-|本次任务累计的响应数量。|

**查看gRPC上报结果<a name="zh-cn_topic_0000002313236861_section28851054410"></a>**

慢网络诊断到故障，会通过gRPC上报至ClusterD的公共故障管理中心。

ConfigMap文件会显示相关信息，5秒钟之后自动清除。

![](../../../figures/scheduling/zh-cn_image_0000002300581874.png)

**已支持的慢网络故障<a name="zh-cn_topic_0000002313236861_section19919834124518"></a>**

<a name="zh-cn_topic_0000002313236861_table4804164084414"></a>

|故障码|故障说明|故障级别|
|--|--|--|
|200001010|某节点中产生/恢复慢网络。|NotHandleFault：暂不处理故障。|
|200001011|超节点内的节点间产生/恢复慢网络。|NotHandleFault：暂不处理故障。|
|200001012|不是卡故障导致的慢网络。|NotHandleFault：暂不处理故障。|

## 故障处理<a name="ZH-CN_TOPIC_0000002511346405"></a>

### 故障决策说明<a name="ZH-CN_TOPIC_0000002511346435"></a>

在故障检测完成后，针对每一种故障模式，断点续训通过故障处理或故障容错来恢复训练业务。断点续训特性根据恢复粒度由粗到细提供Job级别重调度、Pod级别重调度、进程级别重调度、弹性训练、进程级在线恢复、算子级在线恢复多层故障处理系统。用户可根据实际情况选择使用对应的子特性。

**图 1**  故障决策说明<a name="fig2639326192019"></a>  
![](../../../figures/scheduling/故障决策说明.png "故障决策说明")

上图中，容错速度代表故障发生到故障恢复的速度，成功率代表故障发生后故障完成恢复的成功率，易用性代表用户使用或集成的成本。

Job级别重调度、Pod级别重调度、进程级别重调度可支持当前断点续训支持的全部故障模式，但依赖存在备份冗余计算服务器资源。如果存在不可修复的硬件故障且无备份冗余计算服务器时，可以通过配置弹性训练功能进行缩容训练。进程级在线恢复当前支持片上内存故障和网络故障。算子级在线恢复当前支持芯片网络故障和灵衢网络故障。

断点续训多层故障处理系统不同层级根据恢复粒度由细到粗可以逐级回退，如[图2](#fig477415371217)所示，如果上一层恢复失败则可以回退到下一层处理方式。

**图 2**  恢复失败说明<a name="fig477415371217"></a>  
![](../../../figures/scheduling/恢复失败说明.png "恢复失败说明")

**重调度模式<a name="zh-cn_topic_0000002198051753_section1536115719358"></a>**

1. 重调度模式：将任务调度到健康的芯片上，并隔离故障芯片。

    重调度模式默认为**Job级别重调度**，每次故障会停止所有的Pod。但在大规模任务中，停止所有Pod后再重调度的成本较高，存在故障恢复时间过长的问题。除此之外断点续训还提供**Pod级别重调度**功能，用户可根据任务规模配置，在故障时刻只停止故障相关的Pod后重调度少量Pod，从而达成故障的快速恢复。为了进一步缩短故障恢复时间、降低故障影响范围，断点续训还提供进程级别重调度及进程级在线恢复功能。

    **表 1**  各种重调度级别的差异

    <a name="zh-cn_topic_0000002198051753_table18771108163419"></a>

    |重调度的级别|恢复训练耗时|配置步骤|说明|
    |--|--|--|--|
    |Job级别重调度|Job级重调度的恢复时间较长，随着任务规模增加恢复时间超线性劣化。|<p>Job级重调度操作步骤简单，使用MindCluster的用户仅打开配置开关即可使用。</p><p>关键配置步骤请参见[配置Job级别重调度](./04_configuring_fault_handling_policies.md#配置job级别重调度)。</p>|为了进一步降低恢复中资源调度时间，用户可以选择在Job级重调度上开启Pod级重调度能力。|
    |Pod级别重调度|Pod级重调度可以将资源调度时间缩短，且与任务规模无关。但是，Pod级重调度并不能优化训练初始化过程中的时间开销，整体恢复时间仍然会随着任务规模增加而超线性劣化。|<p>Pod级重调度用户需要额外在训练容器中集成训练进程管理能力，使用MindCluster的用户具备对应进程管理能力后即可使用。</p><p>关键配置步骤请参见[配置Pod级别重调度](./04_configuring_fault_handling_policies.md#配置pod级别重调度)。</p>|为了进一步降低训练初始化中的恢复时间，用户可以选择在Pod级重调度上开启进程级重调度能力。|
    |进程级别重调度（进程级恢复）|进程级重调度可以减少训练初始化时间，将整体恢复时间缩短，且与任务规模无关或者弱相关。|<p>相比Pod级重调度，进程级重调度用户需要额外在训练框架中集成高可用训练能力，使用MindCluster的用户需要修改训练脚本，并开启对应配置开关后使用。</p><p>关键配置步骤请参见[配置进程级别重调度](./04_configuring_fault_handling_policies.md#配置进程级别重调度)。</p>|为了解决大规模场景下MTBF时间较短的问题，进一步降低整体恢复时间，用户可以选择在进程级重调度上开启进程级在线恢复能力。|
    |进程级在线恢复|进程级在线恢复比起进程级重调度，恢复训练耗时更低。|<p>相比进程级重调度，进程级在线恢复用户需要配置对应的配置开关后使用。</p><p>关键配置步骤请参见[配置进程级在线恢复](./04_configuring_fault_handling_policies.md#配置进程级在线恢复)。</p>|当前进程级在线恢复支持片上内存故障和网络故障，其余故障场景将回退其他处理方式。|
    |算子级在线恢复|-|关键配置步骤请参见[配置算子级在线恢复](./04_configuring_fault_handling_policies.md#配置算子级在线恢复)。|-|

2. 重调度模式存在以下两种重调度策略。

    - **直接重调度**：训练过程中发生集群调度组件可以探测到的硬件故障，系统将故障节点或芯片进行隔离，直接对任务进行重调度。
    - **无条件重试**：训练过程中发生集群调度组件不能探测到的故障，导致任务容器异常退出，系统无条件对任务进行重调度。

    **表 2**  重调度策略说明

    <a name="zh-cn_topic_0000002198051753_table37727194382"></a>
    
    |重调度策略|说明|支持的故障类型|
    |--|--|--|
    |直接重调度|系统将故障的节点或芯片进行隔离，然后直接对任务进行重调度。|已知的节点故障或重调度处理级别芯片故障。|
    |无条件重试|<p>系统对配置了无条件重试次数的任务，进行指定次数内的重调度。</p><p>成功重调度后，任务可重试次数将减1，当可重试次数为0时无法再次触发重调度。</p><p>如需使用无条件重试功能，需在YAML中配置fault-retry-times参数，详细参数说明请参见[YAML参数说明](./06_configuring_the_job_yaml_file.md#yaml参数说明)。</p>|由于参数面网络故障或者训练相关软件故障等，导致任务异常退出，Pod的Status变为Failed状态的相关故障。|

### Job级别重调度<a name="ZH-CN_TOPIC_0000002479226586"></a>

**Job级别重调度**即每次故障停止所有Pod，重新创建并重调度所有Pod后，重启训练任务。重调度模式默认为**Job级别重调度**。

了解Job级别重调度的关键配置步骤，请参见[配置Job级别重调度](./04_configuring_fault_handling_policies.md#配置job级别重调度)。

**使用约束<a name="zh-cn_topic_0000002039194017_section1178044918127"></a>**

- 本功能仅支持在6.0.RC2及以上版本中使用。
- 大规模K8s集群场景下，ConfigMap映射时延不可控，建议RankTable使用共享存储方式。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039194017_section140112935318"></a>**

**表 1**  Job级别重调度支持的产品和框架

<a name="zh-cn_topic_0000002039194017_table6198201175416"></a>

|产品类型|硬件形态|训练框架|
|--|--|--|
|Atlas 训练系列产品|<ul><li>Atlas 800 训练服务器（型号 9000）</li><li>Atlas 800 训练服务器（型号 9010）</li></ul><p>若Atlas 800 训练服务器的芯片工作模式为SMP模式，且每个Pod申请的NPU数量为1、2时，不支持使用重调度模式。查询和设置NPU芯片工作模式的详细介绍请参见《Atlas 800 训练服务器 iBMC用户指南（型号 9000）》中的“[查询和设置NPU芯片工作模式（npuworkmode）](https://support.huawei.com/enterprise/zh/doc/EDOC1100136583/b6e6ed5a)”章节。</p>|<ul><li>MindSpore</li><li>TensorFlow</li><li>PyTorch</li></ul>|
|Atlas A2 训练系列产品|<ul><li>Atlas 800T A2 训练服务器</li><li>Atlas 200T A2 Box16 异构子框</li><li>Atlas 900 A2 PoD 集群基础单元</li></ul>|<ul><li>MindSpore</li><li>TensorFlow</li><li>PyTorch</li></ul>|
|Atlas A3 训练系列产品|<ul><li>Atlas 900 A3 SuperPoD 超节点</li><li>Atlas 800T A3 超节点服务器</li></ul>|<ul><li>MindSpore</li><li>TensorFlow</li><li>PyTorch</li></ul>|
|A200T A3 Box8 超节点服务器|A200T A3 Box8 超节点服务器|<ul><li>MindSpore</li><li>TensorFlow</li><li>PyTorch</li></ul>|

**重调度原理<a name="zh-cn_topic_0000002039194017_section57901137171110"></a>**

训练过程中如果出现了软硬件故障，将导致训练状态异常。Job级别重调度首先销毁所有的训练容器，然后隔离故障设备，再重新将训练容器调度启动。训练容器重新启动后重新拉起训练，该行为类似训练首次拉起过程。

**图 1**  原理图<a name="fig18343114924113"></a>  
![](../../../figures/scheduling/原理图.png "原理图")

在以上原理图中，各个步骤的说明如下。

1. 检测到故障后，首先删除当前任务所有的Pod和容器。
2. 隔离故障所在的设备，防止再次使用该设备。
3. 重新创建和调度训练Pod和容器。
4. 容器启动后，拉起训练进程恢复训练任务。

### Pod级别重调度<a name="ZH-CN_TOPIC_0000002511346429"></a>

**Pod级别重调度**即每次故障只停止故障相关的Pod，重新创建并重调度故障相关的Pod后，重启训练任务。如果当前故障不能恢复，则回退至Job级重调度模式。相比于Job级别重调度，Pod级别重调度会减少部分资源调度、Pod创建的时间。

了解Pod级别重调度的关键配置步骤，请参见[配置Pod级别重调度](./04_configuring_fault_handling_policies.md#配置pod级别重调度)。

**使用约束<a name="zh-cn_topic_0000002003034876_section11983145119441"></a>**

- 在大集群训练任务中使用**Pod级别重调度**时，建议设置open files参数（可以打开的最大文件数目）足够大，设置过小可能导致Pod重调度出现异常。例如执行**ulimit -n 100000**命令，将open files参数设置为100000。
- 当训练任务的annotation中hccl/rankIndex字段为0的Pod发生故障时，不触发Pod级别重调度和进程级别重调度，直接触发Job级别重调度。
- 请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002003034876_section48174410591"></a>**

**表 1**  重调度支持的产品和框架

<a name="zh-cn_topic_0000002003034876_table1991711954417"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002003034876_row1091711912447"><th class="cellrowborder" valign="top" width="20.462046204620464%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002003034876_p199171819164417"><a name="zh-cn_topic_0000002003034876_p199171819164417"></a><a name="zh-cn_topic_0000002003034876_p199171819164417"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="63.10631063106311%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002003034876_p2917819114420"><a name="zh-cn_topic_0000002003034876_p2917819114420"></a><a name="zh-cn_topic_0000002003034876_p2917819114420"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="16.43164316431643%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002003034876_p27578257424"><a name="zh-cn_topic_0000002003034876_p27578257424"></a><a name="zh-cn_topic_0000002003034876_p27578257424"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002003034876_row12917151994410"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003034876_p339114714459"><a name="zh-cn_topic_0000002003034876_p339114714459"></a><a name="zh-cn_topic_0000002003034876_p339114714459"></a><span id="zh-cn_topic_0000002003034876_ph327965117217"><a name="zh-cn_topic_0000002003034876_ph327965117217"></a><a name="zh-cn_topic_0000002003034876_ph327965117217"></a>Atlas 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002003034876_ul17412295261"></a><a name="zh-cn_topic_0000002003034876_ul17412295261"></a><ul id="zh-cn_topic_0000002003034876_ul17412295261"><li><span id="ph1179307345"><a name="ph1179307345"></a><a name="ph1179307345"></a>Atlas 800 训练服务器（型号 9000）</span></li><li><span id="zh-cn_topic_0000002039194017_ph1627888115712"><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a>Atlas 800 训练服务器（型号 9010）</span><div class="note" id="zh-cn_topic_0000002003034876_note186291241356"><a name="zh-cn_topic_0000002003034876_note186291241356"></a><a name="zh-cn_topic_0000002003034876_note186291241356"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002003034876_p86294411854"><a name="zh-cn_topic_0000002003034876_p86294411854"></a><a name="zh-cn_topic_0000002003034876_p86294411854"></a>若<span id="zh-cn_topic_0000002003034876_ph1162924110518"><a name="zh-cn_topic_0000002003034876_ph1162924110518"></a><a name="zh-cn_topic_0000002003034876_ph1162924110518"></a>Atlas 800 训练服务器</span>的芯片工作模式为SMP模式，且每个Pod申请的NPU数量为1、2时，不支持使用重调度模式。查询和设置NPU芯片工作模式的详细介绍请参见<span id="zh-cn_topic_0000002003034876_ph66296417518"><a name="zh-cn_topic_0000002003034876_ph66296417518"></a><a name="zh-cn_topic_0000002003034876_ph66296417518"></a>《Atlas 800 训练服务器 iBMC用户指南（型号 9000）》中的“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100136583/b6e6ed5a" target="_blank" rel="noopener noreferrer">查询和设置NPU芯片工作模式（npuworkmode）</a>”</span>章节。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002003034876_ul353572894311"></a><a name="zh-cn_topic_0000002003034876_ul353572894311"></a><ul id="zh-cn_topic_0000002003034876_ul353572894311"><li><span id="zh-cn_topic_0000002003034876_ph2075216585425"><a name="zh-cn_topic_0000002003034876_ph2075216585425"></a><a name="zh-cn_topic_0000002003034876_ph2075216585425"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002003034876_ph12195638125217"><a name="zh-cn_topic_0000002003034876_ph12195638125217"></a><a name="zh-cn_topic_0000002003034876_ph12195638125217"></a>TensorFlow</span></li><li><span id="zh-cn_topic_0000002003034876_ph19355165113512"><a name="zh-cn_topic_0000002003034876_ph19355165113512"></a><a name="zh-cn_topic_0000002003034876_ph19355165113512"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002003034876_row6171182004512"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003034876_p153913472453"><a name="zh-cn_topic_0000002003034876_p153913472453"></a><a name="zh-cn_topic_0000002003034876_p153913472453"></a><span id="zh-cn_topic_0000002003034876_ph151431757142112"><a name="zh-cn_topic_0000002003034876_ph151431757142112"></a><a name="zh-cn_topic_0000002003034876_ph151431757142112"></a>Atlas A2 训练系列产品</span></p>
<p id="p15647160165615"><a name="p15647160165615"></a><a name="p15647160165615"></a></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002003034876_ul1843217118563"></a><a name="zh-cn_topic_0000002003034876_ul1843217118563"></a><ul id="zh-cn_topic_0000002003034876_ul1843217118563"><li><span id="ph2153181425619"><a name="ph2153181425619"></a><a name="ph2153181425619"></a>Atlas 800T A2 训练服务器</span></li><li><span id="zh-cn_topic_0000002003034876_ph1114211211203"><a name="zh-cn_topic_0000002003034876_ph1114211211203"></a><a name="zh-cn_topic_0000002003034876_ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span></li><li><span id="zh-cn_topic_0000002003034876_ph495114991519"><a name="zh-cn_topic_0000002003034876_ph495114991519"></a><a name="zh-cn_topic_0000002003034876_ph495114991519"></a>Atlas 900 A2 PoD 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002003034876_ul693112434815"></a><a name="zh-cn_topic_0000002003034876_ul693112434815"></a><ul id="zh-cn_topic_0000002003034876_ul693112434815"><li><span id="zh-cn_topic_0000002003034876_ph1393112494820"><a name="zh-cn_topic_0000002003034876_ph1393112494820"></a><a name="zh-cn_topic_0000002003034876_ph1393112494820"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002003034876_ph2932182416487"><a name="zh-cn_topic_0000002003034876_ph2932182416487"></a><a name="zh-cn_topic_0000002003034876_ph2932182416487"></a>TensorFlow</span></li><li><span id="zh-cn_topic_0000002003034876_ph2093210246488"><a name="zh-cn_topic_0000002003034876_ph2093210246488"></a><a name="zh-cn_topic_0000002003034876_ph2093210246488"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002003034876_row62157458147"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003034876_p18222246142212"><a name="zh-cn_topic_0000002003034876_p18222246142212"></a><a name="zh-cn_topic_0000002003034876_p18222246142212"></a><span id="zh-cn_topic_0000002003034876_ph18411121792018"><a name="zh-cn_topic_0000002003034876_ph18411121792018"></a><a name="zh-cn_topic_0000002003034876_ph18411121792018"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002003034876_ul1367372444211"></a><a name="zh-cn_topic_0000002003034876_ul1367372444211"></a><ul id="zh-cn_topic_0000002003034876_ul1367372444211"><li><p id="p14426829306"><a name="p14426829306"></a><a name="p14426829306"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039194017_ul7201511105411"></a><a name="zh-cn_topic_0000002039194017_ul7201511105411"></a><ul id="zh-cn_topic_0000002039194017_ul7201511105411"><li><span id="zh-cn_topic_0000002039194017_ph52034113546"><a name="zh-cn_topic_0000002039194017_ph52034113546"></a><a name="zh-cn_topic_0000002039194017_ph52034113546"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002039194017_ph13203811195412"><a name="zh-cn_topic_0000002039194017_ph13203811195412"></a><a name="zh-cn_topic_0000002039194017_ph13203811195412"></a>TensorFlow</span></li><li><span id="zh-cn_topic_0000002039194017_ph620418118547"><a name="zh-cn_topic_0000002039194017_ph620418118547"></a><a name="zh-cn_topic_0000002039194017_ph620418118547"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row999211122017"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="p09912115201"><a name="p09912115201"></a><a name="p09912115201"></a><span id="ph126247155413"><a name="ph126247155413"></a><a name="ph126247155413"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="63.10631063106311%" headers="mcps1.2.4.1.2 "><p id="p49961172020"><a name="p49961172020"></a><a name="p49961172020"></a><span id="ph6124114710214"><a name="ph6124114710214"></a><a name="ph6124114710214"></a>A200T A3 Box8 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="16.43164316431643%" headers="mcps1.2.4.1.3 "><a name="ul5581185452113"></a><a name="ul5581185452113"></a><ul id="ul5581185452113"><li><span id="ph19581195472117"><a name="ph19581195472117"></a><a name="ph19581195472117"></a>MindSpore</span></li><li><span id="ph1758125422113"><a name="ph1758125422113"></a><a name="ph1758125422113"></a>TensorFlow</span></li><li><span id="ph8581154132114"><a name="ph8581154132114"></a><a name="ph8581154132114"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**重调度原理<a name="zh-cn_topic_0000002003034876_section19557184814234"></a>**

训练过程中如果出现了软硬件故障，将导致训练状态异常。Pod级别重调度首先销毁当前任务中故障的Pod和容器，并通知其他训练容器中的管理进程销毁所有训练进程，然后隔离故障设备，再重新将训练容器调度启动。待训练容器重新启动后，通知所有容器中的管理进程重新拉起训练进程恢复训练。

1. 检测到故障后，仅删除当前任务中故障的Pod和容器，销毁所有训练进程。
2. 隔离故障所在的设备，防止再次使用该设备。
3. 重新创建和调度训练Pod和容器。
4. 容器启动后，拉起训练进程恢复训练。

### 进程级别重调度<a name="ZH-CN_TOPIC_0000002511346457"></a>

进程级别重调度即每次故障只停止故障相关节点的进程，根据配置策略判断是否退出故障节点。

- recover策略：将故障节点的容器迁移到健康节点；
- recover-in-place策略：对于发生以下两类故障的节点，仅重启故障进程，不迁移故障节点的容器。若多个节点同时发生故障，则只发生以下两类故障的节点仅重启故障进程，不迁移容器，发生其他类型故障的节点会迁移容器。若多个节点发生故障的类型只包含业务进程异常故障，则所有故障节点均会迁移容器。
    - 业务进程异常故障。
    - RestartRequest和RestartBusiness级别的芯片故障。

不能恢复则回退至Job级或Pod级重调度模式。相比于Pod级别重调度，本功能仅重调度故障进程，减少了大量进程间不同步的等待耗时。同时利用了新的HCCL建链方案大大降低了建链耗时，且通过NPU卡间的参数面高速网络P2P传递CKPT信息，避免了CKPT保存和加载的耗时。

了解进程级别重调度的关键配置步骤，请参见[配置进程级别重调度](./04_configuring_fault_handling_policies.md#配置进程级别重调度)。

>[!NOTE] 
>
>- 参数面传递CKPT信息依赖故障卡中的全量优化器副本，如果不存在全量优化器副本则回退为加载存储中的CKPT文件恢复参数。
>- 优化器副本依赖额外的显存占用，如果用户的显存较为紧张，可选择本地加载模式，无论是否存在优化器副本都直接加载存储中的CKPT文件恢复参数。

**使用约束<a name="zh-cn_topic_0000002039353153_section514611624316"></a>**

- 对于PyTorch训练框架，需配套MindSpeed版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
- 对于MindSpore训练框架，需配套MindFormers版本使用，版本配套请参见[MindSpore MindFormers](https://gitcode.com/mindspore/mindformers/tree/master)。
- 当训练任务的annotation中hccl/rankIndex字段为0的Pod发生故障，且需迁移容器时，不触发Pod级别重调度和进程级别重调度，直接触发Job级别重调度。
- 不能和优雅容错功能同时开启。若同时开启，断点续训将通过Job级别重调度恢复训练。
- MindSpore场景下，为保证本功能的正常使用，请将MindSpore和MindIO安装在同一路径下。
- MindSpore场景下，受框架机制限制，进程级重调度存在极小概率失败风险。
- 请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。
- PyTorch只支持单算子模式，只支持基于Megatron框架的模型，只支持acjob类型训练任务。
- 只支持单容器迁移，不支持按照亲和性迁移。
- 不支持多模态模型。
- 不支持开启watchdog功能。
- 不支持在保存Checkpoint期间触发进程级别重调度。
- Atlas A3 训练系列产品场景下，若发生NPU掉卡类、OS断连类的故障，可导致进程级别重调度失败。
- 当故障发生在HCCL建链阶段时，会导致进程级别重调度失败。如果除训练初始化的HCCL建链外，还存在其他训练阶段的HCCL建链，可参考[配置HCCL主动触发建链](./05_configuring_training_recovery.md#配置hccl主动触发建链)章节进行提前建链，防止故障出现在HCCL建链阶段。
- 暂不支持在IPv6场景下使用。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039353153_section136131584164"></a>**

**表 1**  重调度支持的产品和框架

<a name="zh-cn_topic_0000002039353153_table1991711954417"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039353153_row1091711912447"><th class="cellrowborder" valign="top" width="20.462046204620464%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039353153_p199171819164417"><a name="zh-cn_topic_0000002039353153_p199171819164417"></a><a name="zh-cn_topic_0000002039353153_p199171819164417"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="66.2966296629663%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039353153_p2917819114420"><a name="zh-cn_topic_0000002039353153_p2917819114420"></a><a name="zh-cn_topic_0000002039353153_p2917819114420"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="13.24132413241324%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039353153_p27578257424"><a name="zh-cn_topic_0000002039353153_p27578257424"></a><a name="zh-cn_topic_0000002039353153_p27578257424"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039353153_row6171182004512"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p153913472453"><a name="zh-cn_topic_0000002039353153_p153913472453"></a><a name="zh-cn_topic_0000002039353153_p153913472453"></a><span id="zh-cn_topic_0000002039353153_ph151431757142112"><a name="zh-cn_topic_0000002039353153_ph151431757142112"></a><a name="zh-cn_topic_0000002039353153_ph151431757142112"></a>Atlas A2 训练系列产品</span></p>
<p id="p737515258512"><a name="p737515258512"></a><a name="p737515258512"></a></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039353153_ul1843217118563"></a><a name="zh-cn_topic_0000002039353153_ul1843217118563"></a><ul id="zh-cn_topic_0000002039353153_ul1843217118563"><li><p id="p1546725019404"><a name="p1546725019404"></a><a name="p1546725019404"></a><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span></p>
</li><li><span id="zh-cn_topic_0000002039353153_ph1114211211203"><a name="zh-cn_topic_0000002039353153_ph1114211211203"></a><a name="zh-cn_topic_0000002039353153_ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span></li><li><span id="zh-cn_topic_0000002039353153_ph495114991519"><a name="zh-cn_topic_0000002039353153_ph495114991519"></a><a name="zh-cn_topic_0000002039353153_ph495114991519"></a>Atlas 900 A2 PoD 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039353153_ul693112434815"></a><a name="zh-cn_topic_0000002039353153_ul693112434815"></a><ul id="zh-cn_topic_0000002039353153_ul693112434815"><li><span id="zh-cn_topic_0000002039353153_ph1393112494820"><a name="zh-cn_topic_0000002039353153_ph1393112494820"></a><a name="zh-cn_topic_0000002039353153_ph1393112494820"></a>MindSpore</span></li><li><span id="zh-cn_topic_0000002039353153_ph2093210246488"><a name="zh-cn_topic_0000002039353153_ph2093210246488"></a><a name="zh-cn_topic_0000002039353153_ph2093210246488"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039353153_row62157458147"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p18222246142212"><a name="zh-cn_topic_0000002039353153_p18222246142212"></a><a name="zh-cn_topic_0000002039353153_p18222246142212"></a><span id="zh-cn_topic_0000002039353153_ph18411121792018"><a name="zh-cn_topic_0000002039353153_ph18411121792018"></a><a name="zh-cn_topic_0000002039353153_ph18411121792018"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><a name="ul61561253231"></a><a name="ul61561253231"></a><ul id="ul61561253231"><li><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><a name="ul18946810161311"></a><a name="ul18946810161311"></a><ul id="ul18946810161311"><li><span id="ph99461100137"><a name="ph99461100137"></a><a name="ph99461100137"></a>MindSpore</span><p id="p664545214"><a name="p664545214"></a><a name="p664545214"></a><span id="ph294661010130"><a name="ph294661010130"></a><a name="ph294661010130"></a></span></p>
</li><li><span id="ph99469109139"><a name="ph99469109139"></a><a name="ph99469109139"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**重调度原理<a name="zh-cn_topic_0000002039353153_section12206164333619"></a>**

训练过程中如果出现了软硬件故障，将导致训练状态异常。进程级重调度根据配置策略首先销毁故障的训练进程或容器，并通知其他训练容器中的训练进程暂停当前训练任务，然后隔离故障设备，再重新将训练容器调度启动。故障训练容器重新启动后，通知所有容器中的训练进程进行集合通信重建链。建链完成后，将CKPT通过参数面发送给新拉起的训练进程恢复参数，恢复后所有进程重新执行当前step恢复训练。

**图 1**  进程级别重调度原理示意图<a name="fig1373016583373"></a>  
![](../../../figures/scheduling/进程级别重调度原理示意图.png "进程级别重调度原理示意图")

在以上原理图中，各个步骤的说明如下。

1. 设备出现硬件故障后，MindCluster在服务器上的检测组件上报故障信息到ClusterD中，软件故障由容器内MindIO Controller感知并上报到ClusterD。
2. ClusterD将故障服务器上的任务容器退出故障训练进程，重新调度到备用的服务器上。
3. ClusterD通知Master节点上的MindIO Controller进行容错，容错流程包括通知停止训练、通知全局故障、通知恢复策略。
4. MindIO Controller通知每个训练进程中的MindIO Processor，MindIO Processor调用PTA强制停止训练进程。MindIO Processor清理正常节点的资源，销毁通信域，清理后等待新进程加入。
5. 备用服务器上的管理进程拉起训练进程后，创建新的MindIO Processor，MindIO Controller通知每个训练进程中的MindIO Processor恢复训练。
6. 各个进程进行集合通信建链。
7. 正常服务器上的NPU通过参数面将CKPT传递到备用服务器上，完成参数状态恢复后继续训练。

**功能适配点<a name="section1446615300284"></a>**

在进程级别重调度中，集群大脑会根据全局故障信息决策恢复策略并将策略下发到MindIO，调度器需要支持故障Pod调度，而非整个任务重调度，支持恢复策略依次回退。在训练容器中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保障模型参数的冗余备份。在异常发生时，通过异常捕获装饰器捕获故障模式，在恢复时执行算子资源清理，节点重启后触发通信重建。通过参数面在线修复和状态回滚，完成进程级重调度恢复。

对于非MindSpeed-LLM和MindCluster平台用户，需在框架侧完成[表2](#table1995514113610)的功能适配。

**表 2**  进程级别重调度框架适配功能点

<a name="table1995514113610"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="16.77167716771677%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="43.23432343234324%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="18.13181318131813%" id="mcps1.2.5.1.3"><p id="p104301715185316"><a name="p104301715185316"></a><a name="p104301715185316"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="21.862186218621858%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="16.77167716771677%" headers="mcps1.2.5.1.1 "><p id="p18221046175418"><a name="p18221046175418"></a><a name="p18221046175418"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="43.23432343234324%" headers="mcps1.2.5.1.2 "><p id="p14221746205412"><a name="p14221746205412"></a><a name="p14221746205412"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="18.13181318131813%" headers="mcps1.2.5.1.3 "><p id="p5119132211596"><a name="p5119132211596"></a><a name="p5119132211596"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="21.862186218621858%" headers="mcps1.2.5.1.4 "><p id="p252632095917"><a name="p252632095917"></a><a name="p252632095917"></a><a href="../../fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1122104645414"><a name="p1122104645414"></a><a name="p1122104645414"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p322446155419"><a name="p322446155419"></a><a name="p322446155419"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p152294645418"><a name="p152294645418"></a><a name="p152294645418"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p6221046125412"><a name="p6221046125412"></a><a name="p6221046125412"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row144014113397"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p52294618541"><a name="p52294618541"></a><a name="p52294618541"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p72284615410"><a name="p72284615410"></a><a name="p72284615410"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row74014111391"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p522194614547"><a name="p522194614547"></a><a name="p522194614547"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p2221346125415"><a name="p2221346125415"></a><a name="p2221346125415"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row74025111392"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9229467542"><a name="p9229467542"></a><a name="p9229467542"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p102218466541"><a name="p102218466541"></a><a name="p102218466541"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row19531411367"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p422846105412"><a name="p422846105412"></a><a name="p422846105412"></a>节点重启及通信重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p922194612545"><a name="p922194612545"></a><a name="p922194612545"></a>通过注册重建回调实现健康节点与故障节点重建通信域。</p>
</td>
</tr>
<tr id="row1708112845416"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p722164611549"><a name="p722164611549"></a><a name="p722164611549"></a>参数面在线修复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1122246145417"><a name="p1122246145417"></a><a name="p1122246145417"></a>通过回调函数完成副本卡与恢复卡恢复处理。</p>
</td>
</tr>
<tr id="row1911610240547"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p92214463547"><a name="p92214463547"></a><a name="p92214463547"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p222154613543"><a name="p222154613543"></a><a name="p222154613543"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row1311652445414"><td class="cellrowborder" valign="top" width="16.77167716771677%" headers="mcps1.2.5.1.1 "><p id="p202220467541"><a name="p202220467541"></a><a name="p202220467541"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="43.23432343234324%" headers="mcps1.2.5.1.2 "><p id="p1022184612549"><a name="p1022184612549"></a><a name="p1022184612549"></a>根据全局故障信息决策恢复策略，并下发到MindIO，支持恢复策略回退，进程级重调度失败回退到Pod级别、Job级别重调度。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="18.13181318131813%" headers="mcps1.2.5.1.3 "><p id="p488619172591"><a name="p488619172591"></a><a name="p488619172591"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="21.862186218621858%" headers="mcps1.2.5.1.4 "><p id="p1211652412545"><a name="p1211652412545"></a><a name="p1211652412545"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.0.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row18952145365"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p72274605415"><a name="p72274605415"></a><a name="p72274605415"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p522104615410"><a name="p522104615410"></a><a name="p522104615410"></a>调度故障Pod，支持调度恢复策略回退。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p11417425315"><a name="p11417425315"></a><a name="p11417425315"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.0.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

### 进程级在线恢复<a name="ZH-CN_TOPIC_0000002479386460"></a>

进程级在线恢复（Step级别重计算恢复）主要针对以下2种故障类型进行故障处理：

- 网络故障：当前仅支持以下两种场景。
    - HCCS L1-L2端口或链路故障时，BGP切路后，若开启算子级在线恢复且执行失败后进行Step级重试，实现进程不退出的故障快速恢复；若关闭算子级在线恢复，则对训练进程进行Step级重试，实现进程不退出的故障快速恢复。
    - RoCE到上级端口或链路故障，且开启算子级在线恢复并执行失败时，对训练进程进行Step级重试，实现进程不退出的故障快速恢复。

- 片上内存故障：片上内存上出现的不可纠正错误（如故障码0x80E01801），先隔离故障片上内存空间，然后对训练进程进行Step级重试，实现进程不退出的故障快速恢复。

在以上2种场景下，如果故障不能恢复，则回退至**重调度模式**。

相比于进程级别重调度，进程级在线恢复不会重调度故障进程，减少了大量进程间不同步的等待耗时。同时通过NPU卡间的参数面高速网络P2P传递CKPT信息，避免了CKPT保存和加载的耗时。

该故障处理模式默认关闭，若要开启请参见[（可选）配置组件](./07_using_resumable_training_on_the_cli.md#可选配置组件)。

了解进程级在线恢复的关键配置步骤，请参见[配置进程级在线恢复](./04_configuring_fault_handling_policies.md#配置进程级在线恢复)。

>[!NOTE] 
>
>- 参数面传递CKPT信息依赖未故障卡中的全量优化器副本，如果不存在全量优化器副本，则回退为加载存储中的CKPT文件恢复参数。
>- 优化器副本依赖额外的显存占用，如果用户的显存较为紧张，可选择本地加载模式，无论是否存在优化器副本都直接加载存储中的CKPT文件恢复参数。

**使用约束<a name="zh-cn_topic_0000002003193196_section17145122992213"></a>**

- 对于PyTorch训练框架，需配套MindSpeed版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
- 对于MindSpore训练框架，需配套MindFormers版本使用，版本配套请参见[MindSpore MindFormers](https://gitcode.com/mindspore/mindformers/tree/master)。
- 依赖于PyTorch的内存管理机制，仅在PYTORCH\_NO\_NPU\_MEMORY\_CACHING未配置时才能使用此功能。
- 针对部分片上内存故障场景无法生效，例如HCCL集合通信使用的内存地址故障，仍需通过进程级重调度或更上层的容错方案恢复。
- 针对MindSpeed-LLM、MindSpeed等模型或训练脚本中定义的全局变量发生故障的场景，详细处理策略请参见[FAQ](../../faq.md#启用进程级在线恢复后报错there-is-unsafe-data-in-the-input-tensor恢复失败)。
- 与优雅容错不能同时开启。若同时开启，断点续训将通过Job级别重调度恢复训练。
- MindSpore场景下，为保证本功能的正常使用，请将MindSpore和MindIO安装在同一路径下。
- MindSpore场景下，需要在启动TaskD  Manager前设置export TASKD\_PROCESS\_ENABLE="on"。
- 请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。
- 不支持多模态模型。
- 不支持MC2开启场景。
- 不支持开启watchdog功能。
- 当故障发生在HCCL建链阶段时，会导致进程级在线恢复失败。如果除训练初始化的HCCL建链外，还存在其他训练阶段的HCCL建链，可参考[配置HCCL主动触发建链](./05_configuring_training_recovery.md#配置hccl主动触发建链)章节进行提前建链，防止故障出现在HCCL建链阶段。
- 暂不支持在IPv6场景下使用。

**支持的产品型号及AI框架<a name="zh-cn_topic_0000002003193196_section108582044132214"></a>**

**表 1**  网络故障进程级在线恢复支持的产品和框架

<a name="zh-cn_topic_0000002003193196_table18104314924"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002003193196_row81042144212"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002003193196_p51041814022"><a name="zh-cn_topic_0000002003193196_p51041814022"></a><a name="zh-cn_topic_0000002003193196_p51041814022"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002003193196_p91041414627"><a name="zh-cn_topic_0000002003193196_p91041414627"></a><a name="zh-cn_topic_0000002003193196_p91041414627"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002003193196_p11040145218"><a name="zh-cn_topic_0000002003193196_p11040145218"></a><a name="zh-cn_topic_0000002003193196_p11040145218"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002003193196_row1910518141229"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002003193196_p191051114524"><a name="zh-cn_topic_0000002003193196_p191051114524"></a><a name="zh-cn_topic_0000002003193196_p191051114524"></a><span id="zh-cn_topic_0000002003193196_ph19105814420"><a name="zh-cn_topic_0000002003193196_ph19105814420"></a><a name="zh-cn_topic_0000002003193196_ph19105814420"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul18927338231"></a><a name="ul18927338231"></a><ul id="ul18927338231"><li><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul17506112910131"></a><a name="ul17506112910131"></a><ul id="ul17506112910131"><li><span id="ph135064298139"><a name="ph135064298139"></a><a name="ph135064298139"></a>MindSpore</span></li></ul>
<a name="ul7506132918139"></a><a name="ul7506132918139"></a><ul id="ul7506132918139"><li><span id="ph550610294136"><a name="ph550610294136"></a><a name="ph550610294136"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**表 2** 片上内存故障进程级在线恢复支持的产品和框架

<a name="table0630917154413"></a>
<table><thead align="left"><tr id="row13630161784418"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="p963031734417"><a name="p963031734417"></a><a name="p963031734417"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="p663151714415"><a name="p663151714415"></a><a name="p663151714415"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="p13631111710444"><a name="p13631111710444"></a><a name="p13631111710444"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="row5631517114410"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p166312178442"><a name="p166312178442"></a><a name="p166312178442"></a><span id="ph1463121734416"><a name="ph1463121734416"></a><a name="ph1463121734416"></a>Atlas A2 训练系列产品</span></p>
<p id="p12631191713449"><a name="p12631191713449"></a><a name="p12631191713449"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul0631181774417"></a><a name="ul0631181774417"></a><ul id="ul0631181774417"><li><span id="ph46319177449"><a name="ph46319177449"></a><a name="ph46319177449"></a>Atlas 800T A2 训练服务器</span></li><li><span id="ph1463131724413"><a name="ph1463131724413"></a><a name="ph1463131724413"></a>Atlas 900 A2 PoD 集群基础单元</span></li><li><span id="ph46311417154417"><a name="ph46311417154417"></a><a name="ph46311417154417"></a>Atlas 900 A2 PoDc 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul3631151714415"></a><a name="ul3631151714415"></a><ul id="ul3631151714415"><li><span id="ph36311817154419"><a name="ph36311817154419"></a><a name="ph36311817154419"></a>MindSpore</span></li></ul>
<a name="ul1263181794418"></a><a name="ul1263181794418"></a><ul id="ul1263181794418"><li><span id="ph1263191704413"><a name="ph1263191704413"></a><a name="ph1263191704413"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="row16631181714416"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p563111714440"><a name="p563111714440"></a><a name="p563111714440"></a><span id="ph363111714444"><a name="ph363111714444"></a><a name="ph363111714444"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul1763161764415"></a><a name="ul1763161764415"></a><ul id="ul1763161764415"><li><span id="ph1963121720449"><a name="ph1963121720449"></a><a name="ph1963121720449"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph1363115172443"><a name="ph1363115172443"></a><a name="ph1363115172443"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul96311517144415"></a><a name="ul96311517144415"></a><ul id="ul96311517144415"><li><span id="ph96310177449"><a name="ph96310177449"></a><a name="ph96310177449"></a>MindSpore</span></li></ul>
<a name="ul7631141712447"></a><a name="ul7631141712447"></a><ul id="ul7631141712447"><li><span id="ph1563101734413"><a name="ph1563101734413"></a><a name="ph1563101734413"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**进程级在线恢复原理<a name="zh-cn_topic_0000002003193196_section961210366427"></a>**

训练过程中如果出现了片上内存故障或网络故障，将导致训练状态异常。进程级在线恢复首先通知所有训练进程停止当前训练，然后保留当前训练信息并修复故障。修复完成后，所有训练进程回退训练状态到当前上一个Step结束时，正常服务器通过参数面将CKPT传递到故障服务器上，完成参数恢复后重新执行当前Step，然后恢复训练任务。

**图 1**  进程级在线恢复原理<a name="fig37536398327"></a>  
![](../../../figures/scheduling/进程级在线恢复原理.png "进程级在线恢复原理")

在以上原理图中，各个步骤的说明如下。

1. 设备出现片上内存故障或网络故障后，MindCluster在服务器上的检测组件上报故障信息到集群大脑ClusterD中。
2. 片上内存故障或网络故障被CANN软件感知，经训练框架上报给MindIO  Processor和MindIO  Controller。
3. MindIO  Controller向集群大脑请求决策是否进行Step级别重计算恢复，集群大脑综合集群其他节点的健康状态给出决策。
4. MindIO  Controller通知每个训练进程中的MindIO  Processor，调用训练框架停止任务、修复故障，保留通信域信息。
5. 正常服务器上的NPU通过参数面将CKPT传递到故障（已修复）服务器上，完成参数状态恢复后继续训练，重新启动当前Step计算。

**适配功能点<a name="section1446615300284"></a>**

在进程级在线恢复中，集群大脑根据故障信息识别网络故障和片上内存故障，下发对应恢复策略，支持恢复策略回退。在训练容器中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保障模型参数的冗余备份。在异常发生时，通过异常捕获装饰器捕获故障模式，在恢复时针对不同故障执行算子资源清理、UCE模型优化器重建、参数面在线修复、状态回滚，完成进程级在线恢复。

对于非MindSpeed-LLM、MindCluster平台用户，针对不同故障需在框架侧完成以下功能适配。

**表 3**  进程级在线恢复针对网络故障框架适配功能点

<a name="table19955141136101"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.61186118611861%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="36.72367236723672%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="17.981798179817982%" id="mcps1.2.5.1.3"><p id="p1912785111610"><a name="p1912785111610"></a><a name="p1912785111610"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="26.68266826682668%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row199614191876"><td class="cellrowborder" valign="top" width="18.61186118611861%" headers="mcps1.2.5.1.1 "><p id="p174797321974"><a name="p174797321974"></a><a name="p174797321974"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="36.72367236723672%" headers="mcps1.2.5.1.2 "><p id="p1847910326710"><a name="p1847910326710"></a><a name="p1847910326710"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="17.981798179817982%" headers="mcps1.2.5.1.3 "><p id="p12303135518715"><a name="p12303135518715"></a><a name="p12303135518715"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="26.68266826682668%" headers="mcps1.2.5.1.4 "><p id="p1878873515913"><a name="p1878873515913"></a><a name="p1878873515913"></a><a href="../../fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row149661916713"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1947943212711"><a name="p1947943212711"></a><a name="p1947943212711"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p04796323710"><a name="p04796323710"></a><a name="p04796323710"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row1239411299541"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p94791332771"><a name="p94791332771"></a><a name="p94791332771"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p5479103211720"><a name="p5479103211720"></a><a name="p5479103211720"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row13395629115418"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17479113217716"><a name="p17479113217716"></a><a name="p17479113217716"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p10479532172"><a name="p10479532172"></a><a name="p10479532172"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row7395142913549"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1447912327718"><a name="p1447912327718"></a><a name="p1447912327718"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1847916321477"><a name="p1847916321477"></a><a name="p1847916321477"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row539519296541"><td class="cellrowborder" valign="top" width="18.61186118611861%" headers="mcps1.2.5.1.1 "><p id="p114808324711"><a name="p114808324711"></a><a name="p114808324711"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="36.72367236723672%" headers="mcps1.2.5.1.2 "><p id="p248011324715"><a name="p248011324715"></a><a name="p248011324715"></a>根据故障信息识别网络故障或片上内存故障，下发对应恢复策略，支持恢复策略回退。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="17.981798179817982%" headers="mcps1.2.5.1.3 "><p id="p16303135517718"><a name="p16303135517718"></a><a name="p16303135517718"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="26.68266826682668%" headers="mcps1.2.5.1.4 "><p id="p19472244965"><a name="p19472244965"></a><a name="p19472244965"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.0.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row7396029145419"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9480632573"><a name="p9480632573"></a><a name="p9480632573"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p84806321578"><a name="p84806321578"></a><a name="p84806321578"></a>调度故障Pod，支持调度恢复策略回退。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p12472134412615"><a name="p12472134412615"></a><a name="p12472134412615"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.0.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

**表 4**  进程级在线恢复针对片上内存故障框架适配功能点

<a name="table14662336155516"></a>
<table><thead align="left"><tr id="row866213619553"><th class="cellrowborder" valign="top" width="17.119999999999997%" id="mcps1.2.5.1.1"><p id="p36629367550"><a name="p36629367550"></a><a name="p36629367550"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="38.769999999999996%" id="mcps1.2.5.1.2"><p id="p6662103635520"><a name="p6662103635520"></a><a name="p6662103635520"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="17.43%" id="mcps1.2.5.1.3"><p id="p1857674501116"><a name="p1857674501116"></a><a name="p1857674501116"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="26.68%" id="mcps1.2.5.1.4"><p id="p966243617552"><a name="p966243617552"></a><a name="p966243617552"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row19662436145518"><td class="cellrowborder" valign="top" width="17.119999999999997%" headers="mcps1.2.5.1.1 "><p id="p339173741211"><a name="p339173741211"></a><a name="p339173741211"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="38.769999999999996%" headers="mcps1.2.5.1.2 "><p id="p1739537151219"><a name="p1739537151219"></a><a name="p1739537151219"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="17.43%" headers="mcps1.2.5.1.3 "><p id="p9527145711216"><a name="p9527145711216"></a><a name="p9527145711216"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="26.68%" headers="mcps1.2.5.1.4 "><p id="p7146223174212"><a name="p7146223174212"></a><a name="p7146223174212"></a><a href="../../fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row566215364551"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p23923716123"><a name="p23923716123"></a><a name="p23923716123"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p11397374123"><a name="p11397374123"></a><a name="p11397374123"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row06621936185512"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p639183714129"><a name="p639183714129"></a><a name="p639183714129"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p15391437141215"><a name="p15391437141215"></a><a name="p15391437141215"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row2662133617558"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p73913781212"><a name="p73913781212"></a><a name="p73913781212"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1039113741210"><a name="p1039113741210"></a><a name="p1039113741210"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row066213685511"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p143953791220"><a name="p143953791220"></a><a name="p143953791220"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p203910376122"><a name="p203910376122"></a><a name="p203910376122"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row666243613555"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1396379125"><a name="p1396379125"></a><a name="p1396379125"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p143993761219"><a name="p143993761219"></a><a name="p143993761219"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row14662143645516"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p639113711122"><a name="p639113711122"></a><a name="p639113711122"></a>UCE模型优化器重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p83953718128"><a name="p83953718128"></a><a name="p83953718128"></a>通过回调函数完成故障卡模型优化器对象操作清理、重建操作。</p>
</td>
</tr>
<tr id="row43068171121"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p139737201218"><a name="p139737201218"></a><a name="p139737201218"></a>参数面在线修复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p539103711214"><a name="p539103711214"></a><a name="p539103711214"></a>通过回调函数完成副本卡与恢复卡恢复处理。</p>
</td>
</tr>
<tr id="row17307161715127"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p183923781214"><a name="p183923781214"></a><a name="p183923781214"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p114017372126"><a name="p114017372126"></a><a name="p114017372126"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row966233613558"><td class="cellrowborder" valign="top" width="17.119999999999997%" headers="mcps1.2.5.1.1 "><p id="p114023721211"><a name="p114023721211"></a><a name="p114023721211"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="38.769999999999996%" headers="mcps1.2.5.1.2 "><p id="p124083761213"><a name="p124083761213"></a><a name="p124083761213"></a>根据故障信息识别网络故障或片上内存故障，下发对应恢复策略，支持恢复策略回退。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="17.43%" headers="mcps1.2.5.1.3 "><p id="p65272572124"><a name="p65272572124"></a><a name="p65272572124"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="26.68%" headers="mcps1.2.5.1.4 "><p id="p14571125414116"><a name="p14571125414116"></a><a name="p14571125414116"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.0.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row16621936105516"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17407378128"><a name="p17407378128"></a><a name="p17407378128"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p140173701218"><a name="p140173701218"></a><a name="p140173701218"></a>调度故障Pod，支持调度恢复策略回退。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p957195451114"><a name="p957195451114"></a><a name="p957195451114"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.0.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

### 算子级在线恢复<a name="ZH-CN_TOPIC_0000002479386484"></a>

Atlas A3 训练系列产品支持在发生参数面网络故障时，HCCL会执行通信算子重传。在故障进程不退出的情况下，算子级在线恢复可容忍更长时间的网络异常，训练任务不中断。

若网络故障的算子级在线恢复（HCCL通信算子重执行）执行失败，则回退至进程级在线恢复

了解算子级在线恢复的关键配置步骤，请参见[配置算子级在线恢复](./04_configuring_fault_handling_policies.md#配置算子级在线恢复)。

>[!NOTE] 
>HCCL（Huawei Collective Communication Library，华为集合通信库）是华为专为昇腾（Ascend）AI处理器设计的分布式通信库，旨在优化多设备（如NPU/GPU）间的高效协作，以加速深度学习模型的分布式训练，适用于需要大规模算力的AI场景。在分布式训练中，HCCL负责协调多个昇腾处理器之间的数据同步（如梯度聚合、参数更新），减少通信开销，提升训练效率。

**使用场景<a name="section4314241154917"></a>**

当前支持在以下2种故障场景下使用算子级在线恢复功能。

- 对于芯片网络相关故障，当算子重传成功时，Volcano会将任务作为亚健康任务处理。当算子重传失败时，Volcano触发重调度处理。
- 对于灵衢总线设备相关故障，HCCL执行算子级在线恢复后，Volcano会将任务作为亚健康任务处理。

**使用约束<a name="section1915719315116"></a>**

- 本特性不支持MC2开启场景。
- 不支持开启watchdog功能。

**算子级在线恢复支持的产品和框架<a name="section996215473410"></a>**

**表 1**  支持的产品和框架

<a name="table11647101624213"></a>
<table><thead align="left"><tr id="row17647111614214"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="p1664831610428"><a name="p1664831610428"></a><a name="p1664831610428"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="p1664816167422"><a name="p1664816167422"></a><a name="p1664816167422"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="p17648141664214"><a name="p17648141664214"></a><a name="p17648141664214"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="row14649101615422"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="p8649101644216"><a name="p8649101644216"></a><a name="p8649101644216"></a><span id="ph96491216144210"><a name="ph96491216144210"></a><a name="ph96491216144210"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><p id="p10649816134219"><a name="p10649816134219"></a><a name="p10649816134219"></a><span id="ph264911612426"><a name="ph264911612426"></a><a name="ph264911612426"></a>Atlas 900 A3 SuperPoD 集群算力系统</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><p id="p1664981614218"><a name="p1664981614218"></a><a name="p1664981614218"></a>-</p>
</td>
</tr>
</tbody>
</table>

**算子级在线恢复原理<a name="section41453583611"></a>**

**图 1**  原理图<a name="fig151851746103612"></a>  
![](../../../figures/scheduling/原理图-8.png "原理图-8")

在以上原理图中，各个步骤的说明如下。

1. 训练过程中，发生HCCS网络平面LinkDown故障或RoCE网络平面LinkDown故障。
2. CANN检测到网络故障，当前算子终止后，进行网络链路恢复（HCCS网络平面进行BGP切路，RoCE网络平面进行借轨通信），通信链路恢复后进行网络算子重执行。
3. 算子重执行成功后，恢复训练迭代。

### 借轨通信任务暂停与回切<a name="ZH-CN_TOPIC_0000002479226530"></a>

Atlas A3 训练系列产品场景下，MindCluster集群调度组件提供训练任务借轨通信的暂停与回切功能。即在训练过程中，使用主动借轨回切接口，可自由切换NPU芯片使用的RoCE网口。

使用借轨回切功能时，NPU芯片的组网关系可参考《Ascend Training Solution 25.1.RC1 组网指南（Atlas A3训练产品）》中的“网络平面介绍 \> 参数面网络 \> [端口对接策略](https://support.huawei.com/enterprise/zh/doc/EDOC1100494585/3e6a1479?idPath=23710424|251366513|22892968|252309113|258915853)”章节。

了解借轨通信任务暂停与回切功能的详细配置方法，请参见[配置借轨通信任务暂停与回切](./04_configuring_fault_handling_policies.md#配置借轨通信任务暂停与回切)。

- 调用[借轨回切接口](../../api/clusterd/08_link_failover_and_switchback_apis.md)执行借轨回切动作前，请先了解NPU芯片组网关系，保证目标NPU的网络链路正常，如果目标NPU为linkdown状态会导致操作失败。
- 以上述组网指南中的接口对接关系为例，对于以下几种情况，调用SwitchNicTrack接口时，指定的dev与op如下：
    1. 若将device0，device8从QDD8借轨切到QDD7，传参dev为\[device0，device8\]，op为\[true，true\]
    2. 若将device0，device8从QDD7回切到QDD8，传参dev为\[device0，device8\]，op为\[false，false\]
    3. 如果单独将device0从QDD8的PortA借轨切到QDD7的PortA，传参dev为\[device0\]，op为\[true\]
    4. 如果单独将device0从QDD7的PortA回切到QDD8的PortA，传参dev为\[device0\]，op为\[false\]
    5. 如果将Leaf1下的全部device借轨切到Leaf2下，传参dev为\[device0，device8，device2，device10，device4，device12，device6，device14 \]，op为\[true，true，true，true，true，true，true，true\]
    6. 如果将Leaf2下的全部device回切到Leaf1下，传参dev为\[device0，device8，device2，device10，device4，device12，device6，device14 \]，op为\[false，false，false，false，false，false，false，false\]

    **图 1**  接口对接关系<a name="fig111354543222"></a>  
    ![](../../../figures/scheduling/接口对接关系.png "接口对接关系")

**使用场景<a name="section14336140104818"></a>**

当前支持在以下2种场景下使用借轨通信任务暂停与回切功能。

- 交换机升级场景：人工触发借轨后升级交换机，再回切。
- 故障处理场景：发生借轨的故障端口在修复完成后，再做人工回切。

**使用约束<a name="section620412554441"></a>**

- 请在训练正常迭代后，再进行借轨或回切指令的下发。
- 确保已开启进程级恢复相关功能特性。
- 暂不支持在IPv6场景下使用。
- 仅支持Pod间为Roce通信的场景。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002098609234_section4771115416256"></a>**

**表 1**  支持的产品和框架

<a name="zh-cn_topic_0000002098609234_table1526819106465"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002098609234_row22681310134611"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002098609234_p137295354447"><a name="zh-cn_topic_0000002098609234_p137295354447"></a><a name="zh-cn_topic_0000002098609234_p137295354447"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002098609234_p1172993554412"><a name="zh-cn_topic_0000002098609234_p1172993554412"></a><a name="zh-cn_topic_0000002098609234_p1172993554412"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002098609234_p97299357449"><a name="zh-cn_topic_0000002098609234_p97299357449"></a><a name="zh-cn_topic_0000002098609234_p97299357449"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002098609234_row71691214122315"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p112681620231"><a name="zh-cn_topic_0000002098609234_p112681620231"></a><a name="zh-cn_topic_0000002098609234_p112681620231"></a><span id="zh-cn_topic_0000002098609234_ph9126121617231"><a name="zh-cn_topic_0000002098609234_ph9126121617231"></a><a name="zh-cn_topic_0000002098609234_ph9126121617231"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul13725194132419"></a><a name="ul13725194132419"></a><ul id="ul13725194132419"><li><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul7583132019396"></a><a name="ul7583132019396"></a><ul id="ul7583132019396"><li><span id="ph135835207394"><a name="ph135835207394"></a><a name="ph135835207394"></a>MindSpore</span></li></ul>
<a name="ul75831320173911"></a><a name="ul75831320173911"></a><ul id="ul75831320173911"><li><span id="ph13583142013394"><a name="ph13583142013394"></a><a name="ph13583142013394"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**借轨通信任务暂停与回切原理<a name="section56986212179"></a>**

**图 2**  原理图<a name="fig9336113210132"></a>  
![](../../../figures/scheduling/原理图-9.png "原理图-9")

在以上原理图中，各个步骤的说明如下。

1. AI平台集成ClusterD，调用ClusterD的gRPC接口下发切换操作，指定需要切换的NPU卡。
2. ClusterD通知MindIO暂停训练。
3. TaskD Manager通知所有TaskD Worker调用训练框架接口执行切换操作。
4. 训练框架按照通信域逐一调用CANN接口执行切换操作。
5. ClusterD判断所有NPU卡的切换操作完成后，再由TaskD通知MindIO在切换完成后继续执行下一个Step训练。

**适配功能点<a name="section1446615300284"></a>**

在借轨通信任务暂停与回切中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。通过主动调用优雅暂停机制，完成当前卡上任务暂停和任务切换。集群大脑需提供对外接口，接收切换指令并管理借轨通信流程。

对于非MindSpeed-LLM、MindCluster平台用户，需在框架侧完成[表2](#table19955141136102)的功能适配。

**表 2**  借轨通信任务暂停与回切框架适配功能点

<a name="table19955141136102"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.87%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="43.419999999999995%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="14.719999999999999%" id="mcps1.2.5.1.3"><p id="p10978953142414"><a name="p10978953142414"></a><a name="p10978953142414"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="22.99%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="18.87%" headers="mcps1.2.5.1.1 "><p id="p1987424102519"><a name="p1987424102519"></a><a name="p1987424102519"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="43.419999999999995%" headers="mcps1.2.5.1.2 "><p id="p14351731182511"><a name="p14351731182511"></a><a name="p14351731182511"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="14.719999999999999%" headers="mcps1.2.5.1.3 "><p id="p922524114255"><a name="p922524114255"></a><a name="p922524114255"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="22.99%" headers="mcps1.2.5.1.4 "><p id="p7146223174212"><a name="p7146223174212"></a><a name="p7146223174212"></a><a href="../../fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p9871924102517"><a name="p9871924102517"></a><a name="p9871924102517"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p16810326255"><a name="p16810326255"></a><a name="p16810326255"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p12878242257"><a name="p12878242257"></a><a name="p12878242257"></a>优雅暂停</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p687122472518"><a name="p687122472518"></a><a name="p687122472518"></a>训练迭代循环最尾部增加MindIO函数调用，实现主动暂停功能。</p>
</td>
</tr>
<tr id="row1297881015253"><td class="cellrowborder" valign="top" width="18.87%" headers="mcps1.2.5.1.1 "><p id="p168711249252"><a name="p168711249252"></a><a name="p168711249252"></a>借轨切换过程管理</p>
</td>
<td class="cellrowborder" valign="top" width="43.419999999999995%" headers="mcps1.2.5.1.2 "><p id="p168762416258"><a name="p168762416258"></a><a name="p168762416258"></a>提供借轨切换请求下发能力，控制训练进程暂停与重启。</p>
</td>
<td class="cellrowborder" valign="top" width="14.719999999999999%" headers="mcps1.2.5.1.3 "><p id="p10461144315257"><a name="p10461144315257"></a><a name="p10461144315257"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="22.99%" headers="mcps1.2.5.1.4 "><p id="p10979110172511"><a name="p10979110172511"></a><a name="p10979110172511"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v26.0.0/component/clusterd/pkg/application/recover/om_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

### （可选）优雅容错<a name="ZH-CN_TOPIC_0000002479226564"></a>

>[!NOTE] 
>该功能已经日落。PyTorch框架在7.2.RC1之后的版本不再支持；MindSpore框架在7.1.RC1之后的版本不再支持。

当用户进行没有备用资源的训练任务，或者期望设备自动恢复时，可以选择使用**优雅容错**功能。即当训练时的芯片设备出现故障后，系统将尝试对故障芯片进行自动恢复，如果可以恢复，则在保持Pod运行状态下，将任务原地拉起继续训练，不能恢复则回退至**重调度模式**。

优雅容错功能无需进行资源调度，即可自动将故障设备恢复。但是它无法降低训练初始化中的恢复时间，通常情况下，优雅容错所需恢复时间大于进程级重调度和进程级在线恢复功能。

了解优雅容错的关键配置步骤，请参见[配置优雅容错](./04_configuring_fault_handling_policies.md#配置优雅容错)。

**使用约束<a name="zh-cn_topic_0000002098609234_section1137610139461"></a>**

- 当前只支持芯片故障使用优雅容错功能。
- 优雅容错功能与进程级别重调度、进程级在线恢复功能不能同时开启。若同时开启，断点续训将通过Job级别重调度恢复训练。
- 暂不支持在IPv6场景下使用。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002098609234_section4771115416256"></a>**

**表 1**  优雅容错支持的产品和框架

<a name="zh-cn_topic_0000002098609234_table1526819106465"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002098609234_row22681310134611"><th class="cellrowborder" valign="top" width="33.333333333333336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002098609234_p137295354447"><a name="zh-cn_topic_0000002098609234_p137295354447"></a><a name="zh-cn_topic_0000002098609234_p137295354447"></a>产品系列</p>
</th>
<th class="cellrowborder" valign="top" width="33.29332933293329%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002098609234_p1172993554412"><a name="zh-cn_topic_0000002098609234_p1172993554412"></a><a name="zh-cn_topic_0000002098609234_p1172993554412"></a>产品名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.373337333733375%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002098609234_p97299357449"><a name="zh-cn_topic_0000002098609234_p97299357449"></a><a name="zh-cn_topic_0000002098609234_p97299357449"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002098609234_row17268131014613"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p889791444417"><a name="zh-cn_topic_0000002098609234_p889791444417"></a><a name="zh-cn_topic_0000002098609234_p889791444417"></a><span id="zh-cn_topic_0000002098609234_ph289810142442"><a name="zh-cn_topic_0000002098609234_ph289810142442"></a><a name="zh-cn_topic_0000002098609234_ph289810142442"></a>Atlas 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039353153_ul17412295261"></a><a name="zh-cn_topic_0000002039353153_ul17412295261"></a><ul id="zh-cn_topic_0000002039353153_ul17412295261"><li><span id="ph1638757114220"><a name="ph1638757114220"></a><a name="ph1638757114220"></a>Atlas 800 训练服务器（型号 9000）</span></li><li><span id="zh-cn_topic_0000002039194017_ph1627888115712"><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a><a name="zh-cn_topic_0000002039194017_ph1627888115712"></a>Atlas 800 训练服务器（型号 9010）</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002098609234_ul1381333331316"></a><a name="zh-cn_topic_0000002098609234_ul1381333331316"></a><ul id="zh-cn_topic_0000002098609234_ul1381333331316"><li><span id="zh-cn_topic_0000002098609234_ph1246144904420"><a name="zh-cn_topic_0000002098609234_ph1246144904420"></a><a name="zh-cn_topic_0000002098609234_ph1246144904420"></a>MindSpore</span></li></ul>
<a name="zh-cn_topic_0000002098609234_ul10570112811135"></a><a name="zh-cn_topic_0000002098609234_ul10570112811135"></a><ul id="zh-cn_topic_0000002098609234_ul10570112811135"><li><span id="zh-cn_topic_0000002098609234_ph473115306133"><a name="zh-cn_topic_0000002098609234_ph473115306133"></a><a name="zh-cn_topic_0000002098609234_ph473115306133"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002098609234_row181221631185611"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p128991832165620"><a name="zh-cn_topic_0000002098609234_p128991832165620"></a><a name="zh-cn_topic_0000002098609234_p128991832165620"></a><span id="zh-cn_topic_0000002098609234_ph13899123211565"><a name="zh-cn_topic_0000002098609234_ph13899123211565"></a><a name="zh-cn_topic_0000002098609234_ph13899123211565"></a>Atlas A2 训练系列产品</span></p>
<p id="p96481557151918"><a name="p96481557151918"></a><a name="p96481557151918"></a></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002098609234_ul13899193245613"></a><a name="zh-cn_topic_0000002098609234_ul13899193245613"></a><ul id="zh-cn_topic_0000002098609234_ul13899193245613"><li><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span></li><li><span id="zh-cn_topic_0000002098609234_ph189001332105615"><a name="zh-cn_topic_0000002098609234_ph189001332105615"></a><a name="zh-cn_topic_0000002098609234_ph189001332105615"></a>Atlas 900 A2 PoD 集群基础单元</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002098609234_ul664419915495"></a><a name="zh-cn_topic_0000002098609234_ul664419915495"></a><ul id="zh-cn_topic_0000002098609234_ul664419915495"><li><span id="zh-cn_topic_0000002098609234_ph146444924919"><a name="zh-cn_topic_0000002098609234_ph146444924919"></a><a name="zh-cn_topic_0000002098609234_ph146444924919"></a>MindSpore</span></li></ul>
<a name="zh-cn_topic_0000002098609234_ul36445934915"></a><a name="zh-cn_topic_0000002098609234_ul36445934915"></a><ul id="zh-cn_topic_0000002098609234_ul36445934915"><li><span id="zh-cn_topic_0000002098609234_ph364489174917"><a name="zh-cn_topic_0000002098609234_ph364489174917"></a><a name="zh-cn_topic_0000002098609234_ph364489174917"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002098609234_row71691214122315"><td class="cellrowborder" valign="top" width="33.333333333333336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002098609234_p112681620231"><a name="zh-cn_topic_0000002098609234_p112681620231"></a><a name="zh-cn_topic_0000002098609234_p112681620231"></a><span id="zh-cn_topic_0000002098609234_ph9126121617231"><a name="zh-cn_topic_0000002098609234_ph9126121617231"></a><a name="zh-cn_topic_0000002098609234_ph9126121617231"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.29332933293329%" headers="mcps1.2.4.1.2 "><a name="ul13725194132419"></a><a name="ul13725194132419"></a><ul id="ul13725194132419"><li><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></li><li><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="33.373337333733375%" headers="mcps1.2.4.1.3 "><a name="ul7583132019396"></a><a name="ul7583132019396"></a><ul id="ul7583132019396"><li><span id="ph135835207394"><a name="ph135835207394"></a><a name="ph135835207394"></a>MindSpore</span></li></ul>
<a name="ul75831320173911"></a><a name="ul75831320173911"></a><ul id="ul75831320173911"><li><span id="ph13583142013394"><a name="ph13583142013394"></a><a name="ph13583142013394"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**优雅容错原理<a name="zh-cn_topic_0000002098609234_section882584011262"></a>**

在节点或芯片故障处理过程中，若使用重调度模式，需要运维人员手动恢复设备。若任务恢复不及时可能导致训练集群中出现大量散点故障，降低集群算力利用率。因此，断点续训在**重调度模式**上增加了**优雅容错**功能，用于优化NPU芯片的部分故障容错能力。

NPU芯片故障中的部分故障可以通过退出芯片上的训练进程以及热复位芯片来恢复，优雅容错功能即针对这部分故障进行恢复处理，不需要重调度任务。

Ascend Device Plugin负责故障的上报以及设备的恢复，管理进程（PyTorch场景下为Elastic Agent组件，MindSpore场景下为TaskD组件）根据Ascend Device Plugin上报的信息进行训练进程的停止与重新拉起，完成故障恢复（不能恢复则回退至**重调度模式**）。集成优雅容错模式需要在业务容器中添加管理进程，管理进程需要具备故障感知、停止训练任务和重启训练任务等能力。

优雅容错模式直接将故障上报到业务容器内的管理进程中（通常通过挂载文件的方式），容器内的管理进程读取故障文件信息获取到故障信息，获取故障信息的流程如[图1](#zh-cn_topic_0000002098609234_fig135111361314)所示。

**图 1**  获取故障信息<a name="zh-cn_topic_0000002098609234_fig135111361314"></a>  
![](../../../figures/scheduling/获取故障信息.png "获取故障信息")

优雅容错模式将故障区分为以下四类，**无需处理**、**重新执行业务**、**需要复位芯片**和**需要重调度**，对于每类故障的处理如[图2](#zh-cn_topic_0000002098609234_fig12620181591012)所示。

**图 2**  优雅容错故障处理流程<a name="zh-cn_topic_0000002098609234_fig12620181591012"></a>  
![](../../../figures/scheduling/优雅容错故障处理流程.png "优雅容错故障处理流程")

### 在线压测<a name="ZH-CN_TOPIC_0000002479226572"></a>

MindCluster支持训练在线压测特性，即在训练过程中可以调用在线压测接口，暂停指定训练任务，对任务使用的节点进行硬件P2P或AIC压力测试。若不存在故障则恢复训练；若存在故障则隔离故障节点，触发断点续训。

**使用约束<a name="zh-cn_topic_0000002039194017_section1178044918127"></a>**

- 对于PyTorch训练框架，需配合MindSpeed-LLM  2.3.0版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
- 对于MindSpore训练框架，需配合MindFormers master版本使用，版本配套请参见[MindSpore MindFormers](https://gitcode.com/mindspore/mindformers/tree/master)。
- 请在训练正常迭代后，再进行在线压测指令的下发。
- 确保已开启进程级恢复相关功能特性。
- 压测过程中不支持重启ClusterD，如果ClusterD异常重启，需要重启训练并下发压测任务。
- 压测过程中，需要关闭热复位功能。
- P2P压测需确保device侧有10G以上的空闲内存。
- 需要在节点增加nodeDEnable=on标签，保证出现压测的节点可以隔离。
- 对于MindSpore训练框架，需要在启动TaskD  Manager前设置export TASKD\_PROCESS\_ENABLE="on"。
- 暂不支持在IPv6场景下使用。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039194017_section140112935318"></a>**

**表 1**  在线压测支持的产品和框架

<a name="zh-cn_topic_0000002039194017_table6198201175416"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039194017_row111997118547"><th class="cellrowborder" valign="top" width="25.172517251725168%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039194017_p91998117543"><a name="zh-cn_topic_0000002039194017_p91998117543"></a><a name="zh-cn_topic_0000002039194017_p91998117543"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="43.834383438343835%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039194017_p3199161115419"><a name="zh-cn_topic_0000002039194017_p3199161115419"></a><a name="zh-cn_topic_0000002039194017_p3199161115419"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="30.993099309930994%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039194017_p5199011125416"><a name="zh-cn_topic_0000002039194017_p5199011125416"></a><a name="zh-cn_topic_0000002039194017_p5199011125416"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039194017_row920001115417"><td class="cellrowborder" valign="top" width="25.172517251725168%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p192011311155411"><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><span id="ph2314323124211"><a name="ph2314323124211"></a><a name="ph2314323124211"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span></p>
<p id="p773278122616"><a name="p773278122616"></a><a name="p773278122616"></a></p>
</td>
<td class="cellrowborder" valign="top" width="43.834383438343835%" headers="mcps1.2.4.1.2 "><p id="p17354133423610"><a name="p17354133423610"></a><a name="p17354133423610"></a><span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 800T A2 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="30.993099309930994%" headers="mcps1.2.4.1.3 "><a name="ul15879359132214"></a><a name="ul15879359132214"></a><ul id="ul15879359132214"><li><span id="ph135835207394"><a name="ph135835207394"></a><a name="ph135835207394"></a>MindSpore</span></li><li><span id="ph19425111582712"><a name="ph19425111582712"></a><a name="ph19425111582712"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039194017_row13204101125410"><td class="cellrowborder" valign="top" width="25.172517251725168%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p172044116542"><a name="zh-cn_topic_0000002039194017_p172044116542"></a><a name="zh-cn_topic_0000002039194017_p172044116542"></a><span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="43.834383438343835%" headers="mcps1.2.4.1.2 "><p id="p4897194703620"><a name="p4897194703620"></a><a name="p4897194703620"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="30.993099309930994%" headers="mcps1.2.4.1.3 "><a name="ul13821123132320"></a><a name="ul13821123132320"></a><ul id="ul13821123132320"><li><span id="ph19127156230"><a name="ph19127156230"></a><a name="ph19127156230"></a>MindSpore</span></li><li><span id="ph310231710274"><a name="ph310231710274"></a><a name="ph310231710274"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**在线压测原理<a name="section56986212179"></a>**

**图 1**  原理图<a name="fig9336113210132"></a>  
![](../../../figures/scheduling/原理图-10.png "原理图-10")

在以上原理图中，各个步骤的说明如下。

1. AI平台集成ClusterD，调用ClusterD的gRPC接口下发压测操作，指定需要压测的节点。
2. ClusterD通知MindIO暂停训练。
3. TaskD Manager通知指定TaskD Worker调用训练框架接口执行压测操作。
4. 训练框架调用指定NPU卡上的CANN接口执行压测操作。
5. ClusterD判断指定NPU卡的压测操作完成后，再由TaskD通知MindIO在压测完成后继续执行下一个Step训练。

**适配功能点<a name="section1446615300284"></a>**

在在线压测中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。通过主动调用优雅暂停机制，完成当前卡上任务暂停，暂停后进行硬件压力测试，测试完成后继续训练。集群大脑需提供对外接口，接收压测指令并管理压测流程。

对于非MindSpeed-LLM、MindCluster平台用户，需在框架侧完成[表2](#table19955141136103)的功能适配。

**表 2**  在线压测框架适配功能点

<a name="table19955141136103"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.98%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="39.26%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="18.01%" id="mcps1.2.5.1.3"><p id="p106021527183014"><a name="p106021527183014"></a><a name="p106021527183014"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="23.75%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="18.98%" headers="mcps1.2.5.1.1 "><p id="p0609650313"><a name="p0609650313"></a><a name="p0609650313"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="39.26%" headers="mcps1.2.5.1.2 "><p id="p195191085319"><a name="p195191085319"></a><a name="p195191085319"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="18.01%" headers="mcps1.2.5.1.3 "><p id="p1855311819317"><a name="p1855311819317"></a><a name="p1855311819317"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="23.75%" headers="mcps1.2.5.1.4 "><p id="p10701822403"><a name="p10701822403"></a><a name="p10701822403"></a><a href="../../fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1960918515317"><a name="p1960918515317"></a><a name="p1960918515317"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p156091533120"><a name="p156091533120"></a><a name="p156091533120"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p76093513110"><a name="p76093513110"></a><a name="p76093513110"></a>优雅暂停</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p136091519311"><a name="p136091519311"></a><a name="p136091519311"></a>训练迭代循环最尾部增加MindIO函数调用，实现主动暂停功能。</p>
</td>
</tr>
<tr id="row46026594305"><td class="cellrowborder" valign="top" width="18.98%" headers="mcps1.2.5.1.1 "><p id="p26091514318"><a name="p26091514318"></a><a name="p26091514318"></a>在线压测过程管理</p>
</td>
<td class="cellrowborder" valign="top" width="39.26%" headers="mcps1.2.5.1.2 "><p id="p14609155183114"><a name="p14609155183114"></a><a name="p14609155183114"></a>提供在线压测请求下发能力，控制训练进程暂停与恢复。</p>
</td>
<td class="cellrowborder" valign="top" width="18.01%" headers="mcps1.2.5.1.3 "><p id="p6553121803118"><a name="p6553121803118"></a><a name="p6553121803118"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="23.75%" headers="mcps1.2.5.1.4 "><p id="p1660265933015"><a name="p1660265933015"></a><a name="p1660265933015"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v26.0.0/component/clusterd/pkg/application/recover/om_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

### 亚健康热切<a name="ZH-CN_TOPIC_0000002479386544"></a>

训练任务配置为亚健康热切策略（hotSwitch）后，当发生亚健康故障时，拉起备份节点后暂停训练进程，再使用备份节点重新拉起训练任务。

**使用约束<a name="zh-cn_topic_0000002039194017_section1178044918127"></a>**

- 对于PyTorch训练框架，需配合MindSpeed-LLM 2.3.0版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
- 对于MindSpore训练框架，需配合MindFormers master版本使用，版本配套请参见[MindSpore MindFormers](https://gitcode.com/mindspore/mindformers/tree/master)。
- 只支持PyTorch单算子模式、基于Megatron框架的模型以及acjob类型训练任务。
- MindSpore场景下，为保证本功能的正常使用，请将MindSpore和MindIO安装在同一路径下。
- 不支持多模态模型。
- 不支持开启watchdog功能。
- 训练任务未出迭代时触发热切，可能会造成MindIO阻塞，最后触发Job级别重调度。
- 当训练任务的annotation中hccl/rankIndex字段为0的Pod发生亚健康故障时，不支持触发亚健康热切。
- 以下异常情况会回退至Job级别重调度，且任务亚健康处理策略降级为ignore，不再处理亚健康故障：
    - 备份Pod拉起后，训练暂停失败。
    - 备份Pod拉起后，MindCluster等待上报训练暂停状态超时（15分钟）。
    - 备份Pod运行失败。
    - 原Pod删除后，训练恢复失败。
    - 原Pod删除后，MindCluster等待上报训练恢复状态超时（15分钟）。

- 配置亚健康热切策略后，会自动增加进程级恢复开关，若发生非亚健康故障，将触发进程级恢复流程。
- 无备节点场景下，无法完成热切流程，任务亚健康处理策略降级为ignore，不再处理亚健康故障。
- 暂不支持在IPv6场景下使用。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039194017_section140112935318"></a>**

**表 1**  亚健康热切支持的产品和框架

<a name="zh-cn_topic_0000002039194017_table6198201175416"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039194017_row111997118547"><th class="cellrowborder" valign="top" width="25.172517251725175%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039194017_p91998117543"><a name="zh-cn_topic_0000002039194017_p91998117543"></a><a name="zh-cn_topic_0000002039194017_p91998117543"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="59.82598259825983%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039194017_p3199161115419"><a name="zh-cn_topic_0000002039194017_p3199161115419"></a><a name="zh-cn_topic_0000002039194017_p3199161115419"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="15.001500150015001%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039194017_p5199011125416"><a name="zh-cn_topic_0000002039194017_p5199011125416"></a><a name="zh-cn_topic_0000002039194017_p5199011125416"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039194017_row920001115417"><td class="cellrowborder" valign="top" width="25.172517251725175%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p192011311155411"><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><a name="zh-cn_topic_0000002039194017_p192011311155411"></a><span id="ph2314323124211"><a name="ph2314323124211"></a><a name="ph2314323124211"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span></p>
<p id="p773278122616"><a name="p773278122616"></a><a name="p773278122616"></a></p>
</td>
<td class="cellrowborder" valign="top" width="59.82598259825983%" headers="mcps1.2.4.1.2 "><p id="p3799611168"><a name="p3799611168"></a><a name="p3799611168"></a><span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 800T A2 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.001500150015001%" headers="mcps1.2.4.1.3 "><a name="ul15879359132214"></a><a name="ul15879359132214"></a><ul id="ul15879359132214"><li><span id="ph135835207394"><a name="ph135835207394"></a><a name="ph135835207394"></a>MindSpore</span></li><li><span id="ph19425111582712"><a name="ph19425111582712"></a><a name="ph19425111582712"></a>PyTorch</span></li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000002039194017_row13204101125410"><td class="cellrowborder" valign="top" width="25.172517251725175%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039194017_p172044116542"><a name="zh-cn_topic_0000002039194017_p172044116542"></a><a name="zh-cn_topic_0000002039194017_p172044116542"></a><span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span></p>
</td>
<td class="cellrowborder" valign="top" width="59.82598259825983%" headers="mcps1.2.4.1.2 "><p id="p13693112166"><a name="p13693112166"></a><a name="p13693112166"></a><span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="15.001500150015001%" headers="mcps1.2.4.1.3 "><a name="ul6531100274"></a><a name="ul6531100274"></a><ul id="ul6531100274"><li><span id="ph1053216019718"><a name="ph1053216019718"></a><a name="ph1053216019718"></a>MindSpore</span></li><li><span id="ph35321001570"><a name="ph35321001570"></a><a name="ph35321001570"></a>PyTorch</span></li></ul>
</td>
</tr>
</tbody>
</table>

**亚健康热切原理<a name="zh-cn_topic_0000002039194017_section57901137171110"></a>**

**图 1**  原理图<a name="fig1770171514241"></a>  
![](../../../figures/scheduling/原理图-11.png "原理图-11")

在以上原理图中，各个步骤的说明如下。

1. ClusterD通过Ascend Device Plugin感知到亚健康故障。
2. ClusterD根据配置策略决策是否进行亚健康热切恢复。
3. ClusterD通知Ascend Operator拉起备份Pod。
4. Volcano调度备份Pod。
5. 备份Pod中创建新的MindIO Processor，MindIO Processor向MindIO Controller发起注册。
6. MindIO Controller下发训练暂停通知。
7. MindIO Controller通知ClusterD训练暂停。
8. ClusterD通知Volcano删除故障Pod。
9. ClusterD通知MindIO恢复训练。

**适配功能点<a name="section1446615300284"></a>**

在亚健康热切中，集群大脑根据亚健康故障信息，为故障Pod设置注解，拉起并调度备份Pod，通知热切策略到MindIO，训练切换到备份Pod后恢复训练。在训练容器中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。在异常发生时，通过异常捕获装饰器捕获故障模式。在新节点启动后，正常节点暂停训练，之后重建通信域，完成新节点参数面恢复，训练状态完成后完成节点热切换。

对于非MindSpeed-LLM、MindCluster平台用户，需在框架侧完成[表2](#table19955141136104)的功能适配。

**表 2**  亚健康热切框架适配功能点

<a name="table19955141136104"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="18.200000000000003%" id="mcps1.2.5.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="39.330000000000005%" id="mcps1.2.5.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="19.670000000000005%" id="mcps1.2.5.1.3"><p id="p237216122367"><a name="p237216122367"></a><a name="p237216122367"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="22.800000000000004%" id="mcps1.2.5.1.4"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="18.200000000000003%" headers="mcps1.2.5.1.1 "><p id="p1698525618364"><a name="p1698525618364"></a><a name="p1698525618364"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="39.330000000000005%" headers="mcps1.2.5.1.2 "><p id="p117503011375"><a name="p117503011375"></a><a name="p117503011375"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="19.670000000000005%" headers="mcps1.2.5.1.3 "><p id="p444112643720"><a name="p444112643720"></a><a name="p444112643720"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="9" valign="top" width="22.800000000000004%" headers="mcps1.2.5.1.4 "><p id="p7146223174212"><a name="p7146223174212"></a><a name="p7146223174212"></a><a href="../../fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1598625612366"><a name="p1598625612366"></a><a name="p1598625612366"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p2986125603612"><a name="p2986125603612"></a><a name="p2986125603612"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1798635633611"><a name="p1798635633611"></a><a name="p1798635633611"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p898615619362"><a name="p898615619362"></a><a name="p898615619362"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row191961528155914"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p6986115693618"><a name="p6986115693618"></a><a name="p6986115693618"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p39861056113613"><a name="p39861056113613"></a><a name="p39861056113613"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row111971728195915"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p99861756103611"><a name="p99861756103611"></a><a name="p99861756103611"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p19986125623612"><a name="p19986125623612"></a><a name="p19986125623612"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row1519712855916"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p698615618367"><a name="p698615618367"></a><a name="p698615618367"></a>节点重启及通信重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p11986856193616"><a name="p11986856193616"></a><a name="p11986856193616"></a>通过注册重建回调实现健康节点与故障节点重建通信域。</p>
</td>
</tr>
<tr id="row1375943411593"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p10986256103613"><a name="p10986256103613"></a><a name="p10986256103613"></a>参数面在线修复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p198635693613"><a name="p198635693613"></a><a name="p198635693613"></a>通过回调函数完成副本卡与恢复卡恢复处理。</p>
</td>
</tr>
<tr id="row876023415918"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p49861556183618"><a name="p49861556183618"></a><a name="p49861556183618"></a>状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p1398655643610"><a name="p1398655643610"></a><a name="p1398655643610"></a>通过回调函数完成数据迭代器重建、框架变量重置。</p>
</td>
</tr>
<tr id="row17605341596"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p11986656153611"><a name="p11986656153611"></a><a name="p11986656153611"></a>优雅暂停</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p199862056113617"><a name="p199862056113617"></a><a name="p199862056113617"></a>训练迭代循环最尾部增加MindIO函数调用，实现主动暂停功能。</p>
</td>
</tr>
<tr id="row144412445361"><td class="cellrowborder" valign="top" width="18.200000000000003%" headers="mcps1.2.5.1.1 "><p id="p129861056183611"><a name="p129861056183611"></a><a name="p129861056183611"></a>热切流程控制</p>
</td>
<td class="cellrowborder" valign="top" width="39.330000000000005%" headers="mcps1.2.5.1.2 "><p id="p14986125653610"><a name="p14986125653610"></a><a name="p14986125653610"></a>管理热切恢复流程，通过设置注解方式管理备份Pod和故障Pod。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="19.670000000000005%" headers="mcps1.2.5.1.3 "><p id="p1045122693710"><a name="p1045122693710"></a><a name="p1045122693710"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="22.800000000000004%" headers="mcps1.2.5.1.4 "><p id="p64451744113612"><a name="p64451744113612"></a><a name="p64451744113612"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v26.0.0/component/clusterd/pkg/application/recover/hot_switch_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row14716101112393"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1371681114396"><a name="p1371681114396"></a><a name="p1371681114396"></a>Pod创建删除</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p071681117390"><a name="p071681117390"></a><a name="p071681117390"></a>通过识别特定注解删除和创建Pod。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p071621117393"><a name="p071621117393"></a><a name="p071621117393"></a><a href="https://gitcode.com/Ascend/mind-cluster/blob/branch_v26.0.0/component/ascend-operator/pkg/controllers/v1/ascendjob_controller.go" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

### 弹性训练<a name="ZH-CN_TOPIC_0000002479226542"></a>

当出现硬件故障，且K8s集群中无可用备份资源时，MindCluster会先按照数据并行域缩容部分节点继续训练，当集群中有可用空闲资源时，再触发扩容恢复原有规模训练。相比于进程级别重调度，解决了集群中无可用备份资源被重调度的问题。

**使用约束<a name="zh-cn_topic_0000002039353153_section514611624316"></a>**

- 仅支持PyTorch配合MindSpeed-LLM 2.3.0版本使用，版本配套请参见[MindSpeed-LLM](https://gitcode.com/Ascend/MindSpeed-LLM/tree/2.3.0)。
- 仅支持acjob类型训练任务。
- 依赖于MindIO的优化器副本，需要存在全量优化器副本，故需要安装MindIO和TaskD配合使用。
- 不能和优雅容错功能同时开启。
- 当训练任务的annotation中hccl/rankIndex字段为0的Pod发生故障时，不支持触发弹性训练。
- 不支持多模态模型。
- 不支持开启watchdog功能。
- 由于弹性训练会额外创建新的通信组，因此可能会导致片上内存占用增加。
- 暂不支持在IPv6场景下使用。

    增加内存大小计算公式：增加内存最大值（MB）= HCCL\_BUFFSIZE \* 2 \* 9，其中，HCCL\_BUFFSIZE默认为200MB，HCCL\_BUFFSIZE的说明详细请参见《CANN 环境变量参考》中的“[HCCL_BUFFSIZE](https://www.hiascend.com/document/detail/zh/canncommercial/850/maintenref/envvar/envref_07_0080.html)”章节。

更多使用约束可参考[MindSpeed-LLM弹性训练功能使用约束](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/docs/pytorch/features/high_availability.md)。

**支持的产品型号和AI框架<a name="zh-cn_topic_0000002039353153_section136131584164"></a>**

**表 1**  弹性训练支持的产品和框架

<a name="zh-cn_topic_0000002039353153_table1991711954417"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039353153_row1091711912447"><th class="cellrowborder" valign="top" width="20.462046204620464%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039353153_p199171819164417"><a name="zh-cn_topic_0000002039353153_p199171819164417"></a><a name="zh-cn_topic_0000002039353153_p199171819164417"></a>产品类型</p>
</th>
<th class="cellrowborder" valign="top" width="66.2966296629663%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039353153_p2917819114420"><a name="zh-cn_topic_0000002039353153_p2917819114420"></a><a name="zh-cn_topic_0000002039353153_p2917819114420"></a>硬件形态</p>
</th>
<th class="cellrowborder" valign="top" width="13.24132413241324%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039353153_p27578257424"><a name="zh-cn_topic_0000002039353153_p27578257424"></a><a name="zh-cn_topic_0000002039353153_p27578257424"></a>训练框架</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039353153_row6171182004512"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p153913472453"><a name="zh-cn_topic_0000002039353153_p153913472453"></a><a name="zh-cn_topic_0000002039353153_p153913472453"></a><span id="zh-cn_topic_0000002039353153_ph151431757142112"><a name="zh-cn_topic_0000002039353153_ph151431757142112"></a><a name="zh-cn_topic_0000002039353153_ph151431757142112"></a>Atlas A2 训练系列产品</span></p>
<p id="p737515258512"><a name="p737515258512"></a><a name="p737515258512"></a></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><p id="p697681955215"><a name="p697681955215"></a><a name="p697681955215"></a><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><p id="p139316519435"><a name="p139316519435"></a><a name="p139316519435"></a><span id="zh-cn_topic_0000002039353153_ph2093210246488"><a name="zh-cn_topic_0000002039353153_ph2093210246488"></a><a name="zh-cn_topic_0000002039353153_ph2093210246488"></a>PyTorch</span></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039353153_row62157458147"><td class="cellrowborder" valign="top" width="20.462046204620464%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039353153_p18222246142212"><a name="zh-cn_topic_0000002039353153_p18222246142212"></a><a name="zh-cn_topic_0000002039353153_p18222246142212"></a><span id="zh-cn_topic_0000002039353153_ph18411121792018"><a name="zh-cn_topic_0000002039353153_ph18411121792018"></a><a name="zh-cn_topic_0000002039353153_ph18411121792018"></a>Atlas A3 训练系列产品</span></p>
</td>
<td class="cellrowborder" valign="top" width="66.2966296629663%" headers="mcps1.2.4.1.2 "><p id="p1711620216528"><a name="p1711620216528"></a><a name="p1711620216528"></a><span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span></p>
</td>
<td class="cellrowborder" valign="top" width="13.24132413241324%" headers="mcps1.2.4.1.3 "><p id="p16887149174313"><a name="p16887149174313"></a><a name="p16887149174313"></a><span id="ph99469109139"><a name="ph99469109139"></a><a name="ph99469109139"></a>PyTorch</span></p>
</td>
</tr>
</tbody>
</table>

**弹性训练原理<a name="section3841210162013"></a>**

**图 1**  原理图<a name="fig130013397201"></a>  
![](../../../figures/scheduling/原理图-12.png "原理图-12")

以上示意图仅以缩容1个DP域为例，实际弹性训练过程中可能会一次缩容多个DP域。图中每个方格代表一个rank。

1. 按照TP（Tensor Parallelism，张量并行）、PP（Pipeline Parallelism，流水线并行）、DP（Data Parallelism，数据并行）正常进行分布式训练。
2. 训练到某一时刻，若某张卡发生故障，且集群中无更多空闲资源可被调度进行断点续训，则按照DP域缩容，即缩容1个DP域对应的Pod（可能包含多个Pod）后继续训练。
3. 缩容训练到某一时刻，集群中有空闲资源时，缩容的Pod会被重新调度，扩容恢复到原有规模继续训练。

**图 2**  流程图<a name="fig7783192415293"></a>  
![](../../../figures/scheduling/流程图.png "流程图")

在以上流程图中，各个步骤的说明如下。

1. 设备出现硬件故障后，MindCluster在服务器上的检测组件上报故障信息到ClusterD中，软件故障由容器内MindIO Controller感知并上报到ClusterD。
2. ClusterD将故障服务器上的任务容器销毁。
3. 若没有备份节点调度新容器，ClusterD通知Master节点上的MindIO Controller进行缩容训练。
4. MindIO Controller通知每个训练进程中的MindIO Processor，MindIO Processor调用PTA停止训练进程，清理正常节点的资源。
5. MindIO Controller通知正常的训练进程中的MindIO Processor执行通信组重建等缩容流程，进行缩容训练。
6. 检测到缩容时删除的Pod重调度成功。
7. ClusterD通过TaskD  Manager通知MindIO Controller执行扩容。
8. MindIO Controller通知每个训练进程中的MindIO Processor，MindIO Processor调用PTA停止训练进程，清理正常节点的资源。
9. 各个进程进行集合通信建链。
10. 正常服务器上的NPU通过参数面将CKPT传递到备用服务器上，完成参数状态恢复后继续训练。

**适配功能点<a name="section1446615300284"></a>**

在弹性训练中，集群大脑会根据全局故障信息决策恢复策略，并将策略下发到MindIO。调度器需要支持故障Pod调度，而非整个任务重调度，支持恢复策略依次回退。在训练容器中，框架首先初始化MindIO服务。启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保证模型参数的冗余备份。当异常发生时，通过异常捕获装饰器捕获故障模式，并由MindIO上报给集群大脑决策。

- 当集群大脑检测到故障，且无冗余备份资源时，下发缩容策略到MindIO，执行算子资源清理、缩容重建，以缩容状态继续训练。
- 当集群大脑检测到有可用资源且新节点成功拉起时，下发扩容策略到MindIO，执行算子资源清理、扩容通信重建、扩容参数面恢复和扩容状态回滚，完成弹性扩容恢复原有规模继续训练。

对于非MindSpeed-LLM和MindCluster平台用户，需在框架侧完成[表2](#table19955141136107)的功能适配。

**表 2**  弹性训练框架适配功能点

<a name="table19955141136107"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="7.520000000000001%" id="mcps1.2.6.1.1"><p id="p4637165993110"><a name="p4637165993110"></a><a name="p4637165993110"></a>序号</p>
</th>
<th class="cellrowborder" valign="top" width="18.810000000000002%" id="mcps1.2.6.1.2"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="34.39%" id="mcps1.2.6.1.3"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="18.190000000000005%" id="mcps1.2.6.1.4"><p id="p237216122367"><a name="p237216122367"></a><a name="p237216122367"></a>适配组件</p>
</th>
<th class="cellrowborder" valign="top" width="21.090000000000003%" id="mcps1.2.6.1.5"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="7.520000000000001%" headers="mcps1.2.6.1.1 "><p id="p26376591313"><a name="p26376591313"></a><a name="p26376591313"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="18.810000000000002%" headers="mcps1.2.6.1.2 "><p id="p1142119117913"><a name="p1142119117913"></a><a name="p1142119117913"></a>初始化启动</p>
</td>
<td class="cellrowborder" valign="top" width="34.39%" headers="mcps1.2.6.1.3 "><p id="p112827185916"><a name="p112827185916"></a><a name="p112827185916"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="16" valign="top" width="18.190000000000005%" headers="mcps1.2.6.1.4 "><p id="p444112643720"><a name="p444112643720"></a><a name="p444112643720"></a>分布式训练框架</p>
</td>
<td class="cellrowborder" rowspan="6" valign="top" width="21.090000000000003%" headers="mcps1.2.6.1.5 "><p id="p7146223174212"><a name="p7146223174212"></a><a name="p7146223174212"></a><a href="../../fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">表2</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p106371759163113"><a name="p106371759163113"></a><a name="p106371759163113"></a>2</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1942113117919"><a name="p1942113117919"></a><a name="p1942113117919"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p92821518193"><a name="p92821518193"></a><a name="p92821518193"></a>优化器更新前上报优化器更新的开始和结束状态。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p363765912314"><a name="p363765912314"></a><a name="p363765912314"></a>3</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p164211711596"><a name="p164211711596"></a><a name="p164211711596"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p22829180917"><a name="p22829180917"></a><a name="p22829180917"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row191961528155914"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7637175903115"><a name="p7637175903115"></a><a name="p7637175903115"></a>4</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p134219118919"><a name="p134219118919"></a><a name="p134219118919"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p192829181594"><a name="p192829181594"></a><a name="p192829181594"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row111971728195915"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1963725993118"><a name="p1963725993118"></a><a name="p1963725993118"></a>5</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1542111118913"><a name="p1542111118913"></a><a name="p1542111118913"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p112826181914"><a name="p112826181914"></a><a name="p112826181914"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row1519712855916"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p363711591310"><a name="p363711591310"></a><a name="p363711591310"></a>6</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p6421121796"><a name="p6421121796"></a><a name="p6421121796"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11282181811916"><a name="p11282181811916"></a><a name="p11282181811916"></a>通过回调函数完成算子资源清理。</p>
</td>
</tr>
<tr id="row1375943411593"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p06375599316"><a name="p06375599316"></a><a name="p06375599316"></a>7</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p34212017916"><a name="p34212017916"></a><a name="p34212017916"></a>弹性训练回调注册</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1528212181493"><a name="p1528212181493"></a><a name="p1528212181493"></a>将弹性训练各个回调函数注册到MindIO。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1163581571719"><a name="p1163581571719"></a><a name="p1163581571719"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_register.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row876023415918"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4637259163118"><a name="p4637259163118"></a><a name="p4637259163118"></a>8</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1342113114912"><a name="p1342113114912"></a><a name="p1342113114912"></a>缩容重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p528210181396"><a name="p528210181396"></a><a name="p528210181396"></a>重建缩容后的通信组、数据迭代器、记录并更新部分框架变量等。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p106351815121720"><a name="p106351815121720"></a><a name="p106351815121720"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_scale_in_rebuild.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row17605341596"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1763711599312"><a name="p1763711599312"></a><a name="p1763711599312"></a>9</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p74211911599"><a name="p74211911599"></a><a name="p74211911599"></a>扩容通信重建</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p52821818299"><a name="p52821818299"></a><a name="p52821818299"></a>新节点与缩容节点重建通信组。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p126358155177"><a name="p126358155177"></a><a name="p126358155177"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_scale_out_rebuild.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row144412445361"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p56378590319"><a name="p56378590319"></a><a name="p56378590319"></a>10</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p64221014919"><a name="p64221014919"></a><a name="p64221014919"></a>扩容参数面恢复</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p728213181097"><a name="p728213181097"></a><a name="p728213181097"></a>通过副本rank与新拉rank参数传输恢复新节点优化器等参数。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p935519615208"><a name="p935519615208"></a><a name="p935519615208"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_repair.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row14716101112393"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1963713597315"><a name="p1963713597315"></a><a name="p1963713597315"></a>11</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p842217115916"><a name="p842217115916"></a><a name="p842217115916"></a>扩容状态回滚</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p72821618294"><a name="p72821618294"></a><a name="p72821618294"></a>恢复缩容时更改框架变量、重建数据集等。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p135516622010"><a name="p135516622010"></a><a name="p135516622010"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/core/high_availability/elastic_training_rollback.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row164994019817"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p563705923117"><a name="p563705923117"></a><a name="p563705923117"></a>12</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17422911918"><a name="p17422911918"></a><a name="p17422911918"></a>新拉起节点torch通信适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1228218181798"><a name="p1228218181798"></a><a name="p1228218181798"></a>新拉起节点恢复前跳过通信。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.4 "><p id="p627084962414"><a name="p627084962414"></a><a name="p627084962414"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py#:~text=def pre_register_patches(self, patch_manager, args):" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row5499401185"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p5637135973120"><a name="p5637135973120"></a><a name="p5637135973120"></a>13</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p11422518917"><a name="p11422518917"></a><a name="p11422518917"></a>缩容训练全局组通信适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p32829185918"><a name="p32829185918"></a><a name="p32829185918"></a>缩容训练时使用缩容后全局组替换原全局组通信。</p>
</td>
</tr>
<tr id="row1550640684"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3637125953118"><a name="p3637125953118"></a><a name="p3637125953118"></a>14</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14221517919"><a name="p14221517919"></a><a name="p14221517919"></a>缩容训练副本组通信适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p6282018699"><a name="p6282018699"></a><a name="p6282018699"></a>缩容训练时副本rank替代故障rank与故障rank所在副本组通信。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.4 "><p id="p1320441192517"><a name="p1320441192517"></a><a name="p1320441192517"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接</a></p>
</td>
</tr>
<tr id="row1650540489"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p06374591313"><a name="p06374591313"></a><a name="p06374591313"></a>15</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5422415915"><a name="p5422415915"></a><a name="p5422415915"></a>缩容训练参数适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p112827184913"><a name="p112827184913"></a><a name="p112827184913"></a>缩容训练时修改num_microbatches、world_size、global_batch_size等参数。</p>
</td>
</tr>
<tr id="row7501940584"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p6637165923110"><a name="p6637165923110"></a><a name="p6637165923110"></a>16</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5422211095"><a name="p5422211095"></a><a name="p5422211095"></a>梯度精度计算适配</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p5282161815916"><a name="p5282161815916"></a><a name="p5282161815916"></a>适配因缩容num_micro_batches等变化导致的精度梯度变化。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p17653121214259"><a name="p17653121214259"></a><a name="p17653121214259"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py" target="_blank" rel="noopener noreferrer">LLM仓参考链接1</a></p>
<p id="p116531412172515"><a name="p116531412172515"></a><a name="p116531412172515"></a><a href="https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/pretrain_gpt.py#:~text=if args.enable_elastic_training:" target="_blank" rel="noopener noreferrer">LLM仓参考链接2</a></p>
</td>
</tr>
<tr id="row145017409813"><td class="cellrowborder" valign="top" width="7.520000000000001%" headers="mcps1.2.6.1.1 "><p id="p17637165943120"><a name="p17637165943120"></a><a name="p17637165943120"></a>17</p>
</td>
<td class="cellrowborder" valign="top" width="18.810000000000002%" headers="mcps1.2.6.1.2 "><p id="p15422131799"><a name="p15422131799"></a><a name="p15422131799"></a>恢复策略决策</p>
</td>
<td class="cellrowborder" valign="top" width="34.39%" headers="mcps1.2.6.1.3 "><p id="p1628291814912"><a name="p1628291814912"></a><a name="p1628291814912"></a>根据全局故障信息决策恢复策略，并将策略下发到MindIO。支持恢复策略回退，弹性训练失败回退到临终遗言等策略。</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="18.190000000000005%" headers="mcps1.2.6.1.4 "><p id="p1504404816"><a name="p1504404816"></a><a name="p1504404816"></a>AI平台</p>
</td>
<td class="cellrowborder" valign="top" width="21.090000000000003%" headers="mcps1.2.6.1.5 "><p id="p20447192572312"><a name="p20447192572312"></a><a name="p20447192572312"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.0.0/component/clusterd/pkg/application/recover" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
<tr id="row155014017818"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p063755903111"><a name="p063755903111"></a><a name="p063755903111"></a>18</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1042215113910"><a name="p1042215113910"></a><a name="p1042215113910"></a>故障Pod调度</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p18282181818913"><a name="p18282181818913"></a><a name="p18282181818913"></a>调度故障Pod。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p10446112592310"><a name="p10446112592310"></a><a name="p10446112592310"></a><a href="https://gitcode.com/Ascend/mind-cluster/tree/branch_v26.0.0/component/ascend-for-volcano/internal/rescheduling" target="_blank" rel="noopener noreferrer">链接</a></p>
</td>
</tr>
</tbody>
</table>

[表2](#table19955141136107)中序号为1-6的适配项为MindIO TFT（MindCluster MindIO Training Fault Tolerance）公共逻辑，序号为17-18的适配项为断点续训公共逻辑，本章节不再详细描述。以下针对弹性训练特有功能点，基于Megatron 0.12.1版本进行简要介绍。

- 弹性训练回调注册

    在训练拉起初始化时调用，将弹性训练缩容和扩容恢复过程中需要执行的回调函数注册到MindIO中，进而在恢复过程中被调用。

- 缩容重建
    1. 基于缩容后成员创建新的全局通信组并记录，后续将替代原全局通信组进行通信。
    2. 记录框架原始DP size、num\_microbatches等参数作为后续扩容恢复使用，并更新为缩容后数据。
    3. 基于故障Rank信息重建缩容后其他局部通信组，并更新模型、优化器等实例对象中的通信组。
    4. 重建数据集、重新初始化部分框架实例、参数等。

- 扩容通信重建
    1. 重建扩容后全局和局部通信组，并更新模型、优化器等实例对象中的通信组。
    2. 恢复框架DP size等参数、重新初始化部分框架实例等。

- 扩容参数面恢复
    1. 为新拉起的rank训练进程和备份rank训练进程创建通信组，用于发送和接收优化器参数等。
    2. 备份rank训练进程向新拉起的rank训练进程发送恢复所需的优化器参数。
    3. 新拉起的rank训练进程接收优化器参数后，按需更新optimizer、opt\_param\_scheduler、全局args等参数。

- 扩容状态回滚
    1. 恢复框架num\_microbatches等参数。
    2. 恢复训练前将优化器参数拷贝到模型参数中，并在对应DP域内进行一次all\_gather通信操作，确保模型参数为最新状态。
    3. 修复打印训练迭代日志。
    4. 重建数据集，重新初始化部分框架实例、参数等。
    5. 销毁恢复过程中发送和接收参数的通信组。

- 新拉起节点torch通信适配
    1. 对于重启节点，从pretrain启动流程到进入train之间，会下发通信算子，但正常训练rank在该阶段并未与重启节点配套重建通信域，集合通信无法成功，因此直接跳过。
    2. 对于重启节点，从pretrain启动流程到进入train之间，会创建并行通信域，但正常训练rank在该阶段并未与重启节点配套重建通信域，对于gloo组会报错，因此直接跳过新建gloo通信组。

- 缩容训练全局组通信适配

    在缩容训练过程中，由于故障节点已经被删除，因此使用原全局通信组通信会失败，需替换为缩容后的全局通信组。

- 缩容训练副本组通信适配

    在[LLM仓参考链接](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py)中，start\_param\_sync\_wrapper、get\_grad\_norm\_fp32\_wrapper、get\_parameter\_state\_dp\_zero\_wrapper等是为了适配缩容训练时副本组通信而patch，下面以get\_parameter\_state\_dp\_zero\_wrapper为例介绍副本组适配原理：

    假设当前tp=8、pp=1、dp=4。DP组分别为rank \[0,8,16,24\]、\[1,9,17,25\]、\[2,10,18,26\]、…、\[7,15,23,31\]，按照副本优化器原理，副本组分别为rank \[0,8\]、\[16,24\]、\[1,9\]、\[17,25\]、\[2,10\]、\[18,26\]、…、\[7,15\]、\[23,31\]，rank 0-15与rank 16-31互为副本。rank 31故障后，将rank 24-31对应DP域删除继续缩容训练。

    原生Megatron会使用优化器实例的data\_parallel\_group\_gloo成员变量对应的group（即DP组，在使用MindIO的优化器副本时为副本组）进行通信。缩容后不包含删除的rank 24-31的副本组，继续按照原有通信组进行通信，包含缩容rank的副本组使用组内正常rank与缩容rank对应的副本rank组成的缩容组进行通信，例如副本组rank \[23,31\]缩容后，通信使用的通信组为rank \[23,15\]。

- 缩容训练参数适配

    在[LLM仓参考链接](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py)中，patch\_world\_size\_func\_wrapper、log\_wrapper、is\_last\_rank\_wrapper、optimizer\_param\_scheduler\_step\_wrapper、track\_app\_tag\_wrapper、print\_rank\_last\_wrapper、num\_floating\_point\_operations\_wrapper等是为了适配global\_batch\_size、world\_size等训练中使用的参数而patch。例如：原生使用dp\_size\*micro\_batch\_size\*num\_microbatches，缩容后各个DP内num\_microbatches可能不一样，因此直接使用args.globatch\_size。缩容后判断是否最后一个rank使用缩容后的全局组；全局组大小修改为缩容后的大小等。

- 梯度精度计算适配

    在[LLM仓参考链接1](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/mindspeed_llm/features_manager/high_availability/high_availability.py)中，start\_grad\_sync\_wrapper、forward\_step\_wrapper、elastic\_training\_get\_forward\_backward\_func\_wrapper以及[LLM仓参考链接2](https://gitcode.com/Ascend/MindSpeed-LLM/blob/2.3.0/pretrain_gpt.py#:~text=if%20args.enable_elastic_training%3A)所指向的loss\_func的代码是为了适配因缩容导致的精度梯度变化而patch或修改。

    - loss\_func由每个micro\_batch都要进行DP组内all\_reduce通信修改为缩容训练时不进行通信，原因是缩容后每个DP域内num\_micro\_batches数量可能不一样，导致前几个DP会多执行一次all\_reduce而卡住。
    - start\_grad\_sync\_wrapper中将梯度缩放因子gradient\_scaling\_factor修改为1.0 / \(arguments.global\_batch\_size / arguments.micro\_batch\_size\)，即在原1/dp\_size基础上再除以num\_micro\_batches。
    - forward\_step\_wrapper将入参num\_microbatches修改为1，目的是loss计算时不再除以num\_microbatches，因为在start\_grad\_sync\_wrapper中已经除以了num\_microbatches。
    - elastic\_training\_get\_forward\_backward\_func\_wrapper因为loss\_func没有执行DP组内all\_reduce，原生forward\_backward\_func执行完成后，在最后一个PP时将losses\_reduced每个key的和（即所有micro\_batch的lm loss相加）在DP组内执行all\_reduce操作求和。

## 训练恢复<a name="ZH-CN_TOPIC_0000002511426359"></a>

### 训练恢复原理说明<a name="ZH-CN_TOPIC_0000002479226500"></a>

在完成故障处理后，训练进程会被重新拉起，拉起的训练进程需要完成模型权重的保存和加载，才能回到任务中断时的训练状态。在正常训练中，每隔一段时间保存训练模型权重的CKPT（Checkpoint）文件，在任务中断后，新拉起的进程可以加载之前保存的CKPT文件，从而恢复到之前保存点的模型权重状态，减少训练时间。对于不同框架，保存和加载CKPT的方法不一样，以下给出了TensorFlow、PyTorch、MindSpore保存和加载CKPT的示例，用户需按照示例修改自己的**训练模型脚本**。

**PyTorch<a name="section77915151121"></a>**

1. 保存CKPT。

    ```Python
    def save_checkpoint(state, is_best, args, filename='checkpoint.pth.tar'):
        filename2 = os.path.join(args.save_ckpt_path, filename)
        torch.save(state, filename2)
        if is_best:
            shutil.copyfile(filename2, os.path.join(args.save_ckpt_path, 'model_best.pth.tar'))
    ```

2. 加载CKPT。

    ```Python
    checkpoint = torch.load(args.checkpoint_path, map_location=loc)
    args.start_epoch = checkpoint['epoch']
    best_acc1 = checkpoint['best_acc1']
    model.load_state_dict(checkpoint['state_dict'])
    optimizer.load_state_dict(checkpoint['optimizer'])
    ```

**MindSpore<a name="section104642081315"></a>**

1. 保存CKPT。

    ```Python
    ms.save_checkpoint(net, "./lenet.ckpt",
                       choice_func=lambda x: x.startswith("conv") and not x.startswith("conv1"))
    ```

2. 加载CKPT。

    ```Python
    param_dict = ms.load_checkpoint("./lenet.ckpt")
    ```

**TensorFlow<a name="section20353419915"></a>**

1. 使用tf.compat.v1.train.CheckpointManager接口进行CKPT管理。

    ```Python
      checkpoint_manager = tf.train.CheckpointManager(
          runnable.checkpoint,
          directory=flags_obj.model_dir,
          max_to_keep=10,
          step_counter=runnable.global_step,
          checkpoint_interval=checkpoint_interval)
    ```

2. 保存CKPT（创建一个新的CKPT）。

    ```Python
    Save(
        Checkpoint_number=None, check_internal=True, options=None
    )
    ```

3. 加载保存的CKPT（尝试加载从目录中的最新的CKPT）。

    ```Python
    Restore_or_initialize()
    ```

### 周期性CKPT保存<a name="ZH-CN_TOPIC_0000002479386434"></a>

现有大规模集群训练主要通过CKPT（Checkpoint）机制，即在训练过程中周期性保存训练过程数据（模型参数等）作为CKPT。当业务平台检测到故障发生后，可退出当前训练任务，通过重新加载CKPT数据，从CKPT保存时刻开始恢复训练，避免从头开始重新进行训练。

周期性CKPT保存分为2个部分：异步CKPT保存以及内存CKPT加载。

- **异步CKPT保存**

    MindIO ACP提供异步保存周期性CKPT的能力。未使用MindIO ACP时，需要将需要保存的参数从设备拷贝到主机侧，再从主机侧落盘到存储中，这一时间通常在分钟级。MindIO ACP提供异步落盘的能力，当需要保存的参数从设备拷贝到主机侧后，通过异步进程进行落盘到存储，不会阻塞训练进程，落盘的过程中训练可以继续进行。

- **内存CKPT加载**

    MindIO ACP提供基于内存的周期性CKPT加载的能力。在训练恢复时，通常需要从存储加载之前保存的周期性CKPT，加载完成后恢复训练状态再继续训练。但是，由于数据量较大和存储性能限制，大模型任务通常加载时间在分钟级。为了降低CKPT加载时间，从而降低训练恢复的时间，MindIO ACP提供基于内存的周期性CKPT加载机制，故障后直接基于内存加载，将降低大量加载的时间。

**推荐配置<a name="section883116216236"></a>**

在使用故障重调度的CKPT保存能力时，需根据实际情况选择周期性保存CKPT频率，用户可参考如[图1](#fig41241253101)所示的推荐频率。

**图 1**  周期性CKPT保存频率推荐<a name="fig41241253101"></a>  
![](../../../figures/scheduling/周期性CKPT保存频率推荐.png "周期性CKPT保存频率推荐")

使用周期CKPT恢复能力，训练恢复后将丢失上一次周期保存点到故障点这一时间段的训练状态。因此，如果想要降低每次故障导致的训练状态损失，需要降低周期性保存的间隔。但是，每次保存需要中断训练后将CKPT从设备侧落盘到存储侧，这浪费了大量的训练时间。如果降低周期性保存的间隔，将导致训练时间的浪费，从而也会带来训练时间的损失。综上所述，如果单次保存时间恒定，通常需要作出保存损失和故障损失的综合权衡。

为了降低上述损失，需要降低单次保存时间。单次保存时间受到保存数据量及存储性能的影响，通常难以改变这两者。本产品提供MindIO ACP产品解决周期性CKPT恢复损失高的问题。

### 临终CKPT保存<a name="ZH-CN_TOPIC_0000002511426397"></a>

尽管通过异步保存周期性CKPT能够降低周期性保存间隔，从而降低每次故障的损失，但是由于仍然具有保存开销，难以做到秒级的故障损失。因此，MindCluster集群调度组件提供临终保存CKPT能力，在故障时刻保存当前step初始的参数状态，从而将训练恢复的状态损失降低到一个“step”以内。

MindCluster MindIO Try To Persist（下文简称MindIO TTP）提供临终CKPT能力，帮助用户在故障时刻保存临终时刻CKPT。

了解临终CKPT保存的详细介绍，请参见[故障恢复加速](../../fault_recovery_acceleration/01_product_description.md)。

了解临终CKPT保存的配置步骤，请参见[配置临终CKPT保存](./05_configuring_training_recovery.md#配置临终ckpt保存)。

**适配功能点<a name="section1446615300284"></a>**

在临终CKPT中，框架首先初始化MindIO服务，启动服务后优化器更新时会上报对应状态到MindIO。随后，创建DP副本组和优化器副本，以保障模型参数的冗余备份。在异常发生时，通过异常捕获装饰器捕获故障模式，之后执行算子资源清理，基于副本完成临终CKPT保存。

对于非MindSpeed-LLM用户，需在框架侧完成[表1](#table19955141136109)的功能适配。

**表 1**  临终CKPT保存框架适配功能点

<a name="table19955141136109"></a>
<table><thead align="left"><tr id="row169591493619"><th class="cellrowborder" valign="top" width="20.632063206320634%" id="mcps1.2.4.1.1"><p id="p46603387387"><a name="p46603387387"></a><a name="p46603387387"></a>适配功能点</p>
</th>
<th class="cellrowborder" valign="top" width="50.51505150515051%" id="mcps1.2.4.1.2"><p id="p176601638153816"><a name="p176601638153816"></a><a name="p176601638153816"></a>功能简述</p>
</th>
<th class="cellrowborder" valign="top" width="28.852885288528853%" id="mcps1.2.4.1.3"><p id="p4660113823812"><a name="p4660113823812"></a><a name="p4660113823812"></a>参考链接</p>
</th>
</tr>
</thead>
<tbody><tr id="row893618158397"><td class="cellrowborder" valign="top" width="20.632063206320634%" headers="mcps1.2.4.1.1 "><p id="p19821165420516"><a name="p19821165420516"></a><a name="p19821165420516"></a>初始化拉起</p>
</td>
<td class="cellrowborder" valign="top" width="50.51505150515051%" headers="mcps1.2.4.1.2 "><p id="p5821185419518"><a name="p5821185419518"></a><a name="p5821185419518"></a>训练框架初始化时拉起MindIO服务。</p>
</td>
<td class="cellrowborder" rowspan="7" valign="top" width="28.852885288528853%" headers="mcps1.2.4.1.3 "><p id="p7146223174212"><a name="p7146223174212"></a><a name="p7146223174212"></a><a href="../../fault_recovery_acceleration/03_usage_guidance.md#对接非mindspeed-llm框架">对接非MindSpeed-LLM框架</a></p>
</td>
</tr>
<tr id="row1793717157396"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p6821754125118"><a name="p6821754125118"></a><a name="p6821754125118"></a>上报优化器更新状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p182111545511"><a name="p182111545511"></a><a name="p182111545511"></a>优化器更新前上报优化器更新开始和结束。</p>
</td>
</tr>
<tr id="row193701523914"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p082105435116"><a name="p082105435116"></a><a name="p082105435116"></a>创建DP副本组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p11821254135115"><a name="p11821254135115"></a><a name="p11821254135115"></a>新增dp_cp/dp_ep副本组及gloo组创建逻辑，在原生Megatron分布式并行组创建后创建相关副本组。</p>
</td>
</tr>
<tr id="row191961528155914"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p1282145475118"><a name="p1282145475118"></a><a name="p1282145475118"></a>优化器副本</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p128211854175117"><a name="p128211854175117"></a><a name="p128211854175117"></a>接管、继承相关Megatron原生优化器功能，嵌入MindIO优化器副本管理逻辑。</p>
</td>
</tr>
<tr id="row111971728195915"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p7821115445113"><a name="p7821115445113"></a><a name="p7821115445113"></a>异常捕获装饰器</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p138216541513"><a name="p138216541513"></a><a name="p138216541513"></a>使用异常捕获装饰器装饰train函数捕获故障模式。</p>
</td>
</tr>
<tr id="row1519712855916"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p6821754135112"><a name="p6821754135112"></a><a name="p6821754135112"></a>算子资源清理</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p14822105475115"><a name="p14822105475115"></a><a name="p14822105475115"></a>通过回调函数完成算子清理、恢复算子下发能力。</p>
</td>
</tr>
<tr id="row1375943411593"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p138221254105112"><a name="p138221254105112"></a><a name="p138221254105112"></a>临终CKPT</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p1582210547513"><a name="p1582210547513"></a><a name="p1582210547513"></a>通过新增回调函数、优化器副本dump方法完成临终CKPT保存。</p>
</td>
</tr>
</tbody>
</table>

### 参数面CKPT传输恢复<a name="ZH-CN_TOPIC_0000002511426371"></a>

通过临终CKPT能力可以将每次训练由于CKPT回滚机制导致的训练回滚损失降到一个“step”内，但是在故障时刻时需要进行落盘保存，并在容错完成训练恢复后需要加载存储中的CKPT进行恢复，将导致整体故障恢复时间延长。因此，为了降低故障恢复时间，MindCluster集群调度组件提供参数面CKPT传输恢复能力。

在故障时刻将参数状态保持在设备侧，在容错完成训练恢复时将正常卡内的参数状态通过参数面网络传输到容错处理的卡上，从而快速恢复容错处理卡的参数状态。当前该能力需要结合进程级别重调度和进程级在线恢复使用，不支持用户独立使用。

了解参数面CKPT的配置步骤，请参见[配置参数面传参恢复](./05_configuring_training_recovery.md#配置参数面传参恢复)。
