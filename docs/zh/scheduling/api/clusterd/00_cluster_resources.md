# 集群资源<a name="ZH-CN_TOPIC_0000002511346785"></a>

## ConfigMap说明<a name="section17868183824213"></a>

ClusterD启动后，会创建如下ConfigMap：

- cluster-info-node-cm，详细说明请参见[表1](#table25031946405)。
- cluster-info-device-$\{m\}，详细说明请参见[表2](#table915714719368)。m为从0开始递增的整数。集群规模每增加1000个节点，则会新增一个该ConfigMap文件。
- cluster-info-switch-$\{x\}，详细说明请参见[表3](#table9246232250)。x为从0开始递增的整数。集群规模每增加2000个节点，则会新增一个该ConfigMap文件。

**表 1**  cluster-info-node-cm

<a name="table25031946405"></a>

|参数|说明|
|--|--|
|mindx-dl-nodeinfo-*\<kwok-node-0\>*|前缀为固定的mindx-dl-nodeinfo，kwok-node-0是节点名称，方便定位故障的具体节点。|
|NodeInfo|节点维度的故障信息。|
|FaultDevList|节点故障设备列表。|
|- DeviceType|故障设备类型。|
|- DeviceId|故障设备ID。|
|- FaultCode|故障码，由英文和数组拼接而成的字符串，字符串表示故障码的十六进制。|
|- FaultLevel|故障处理等级。<ul><li>NotHandleFault：无需处理。</li><li>PreSeparateFault：该节点上有任务则不处理，后续调度时不调度任务到该节点。</li><li>SeparateFault：任务重调度。</li></ul>|
|NodeStatus|节点健康状态，由本节点故障处理等级最严重的设备决定。<ul><li>Healthy：该节点故障处理等级存在且不超过NotHandleFault，该节点为健康节点，可以正常训练。若该节点故障处理等级为PreSeparateFault，且节点有NPU卡正在使用，则该节点为健康节点。任务执行完成后，该节点将变为故障节点。</li><li>UnHealthy：该节点故障处理等级存在SeparateFault，该节点为故障节点，将影响训练任务，立即将任务调离该节点。若该节点故障处理等级为PreSeparateFault，且节点无NPU卡正在使用，则该节点为故障节点，不可将任务调度到该节点。</li></ul>|

**表 2** cluster-info-device-$\{m\}

<a name="table915714719368"></a>

|参数|说明|
|--|--|
|mindx-dl-deviceinfo-*\<kwok-node-0\>*|前缀为固定的mindx-dl-deviceinfo，kwok-node-0是节点名称，用于定位故障的具体节点。|
|huawei.com/Ascend910|<ul><li>当前节点可用的芯片名称信息，存在多个时用英文逗号拼接。</li><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD使用huawei.com/npu作为参数名称。</li></ul>|
|huawei.com/Ascend910-NetworkUnhealthy|<ul><li>当前节点网络不健康的芯片名称信息，存在多个时用英文逗号拼接。</li><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD使用huawei.com/npu-NetworkUnhealthy作为参数名称。</li></ul>|
|huawei.com/Ascend910-Unhealthy|<ul><li>当前芯片不健康的芯片名称信息，存在多个时用英文逗号拼接。</li><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD使用huawei.com/npu-Unhealthy作为参数名称。</li></ul>|
|huawei.com/Ascend910-Fault|<ul><li>数组对象，对象包含fault_type、npu_name、large_model_fault_level、 fault_level、fault_handling、fault_code和fault_time_and_level_map字段。</li><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD使用huawei.com/npu-Fault作为参数名称。</li></ul>|
|- fault_type|故障类型。<ul><li>CardUnhealthy：芯片故障</li><li>CardNetworkUnhealthy：参数面网络故障（芯片网络相关故障）</li><li>NodeUnhealthy：节点故障</li><li>PublicFault：公共故障</li></ul>|
|- npu_name|故障的芯片名称，节点故障时为空。|
|<p>- large_model_fault_level</p><p>- fault_level</p><p>- fault_handling</p>|故障处理类型，节点故障时取值为空。<ul><li>NotHandleFault：不做处理</li><li>RestartRequest：推理场景需要重新执行推理请求，训练场景重新执行训练业务</li><li>RestartBusiness：需要重新执行业务</li><li>FreeRestartNPU：影响业务执行，待芯片空闲时需复位芯片</li><li>RestartNPU：直接复位芯片并重新执行业务</li><li>SeparateNPU：隔离芯片</li><li>PreSeparateNPU：预隔离芯片，会根据训练任务实际运行情况判断是否重调度</li><li>ManuallySeparateNPU：人工隔离芯片。当达到Ascend Device Plugin和ClusterD各自的故障频率，Ascend Device Plugin和ClusterD会将故障芯片进行人工隔离。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>large_model_fault_level、fault_handling和fault_level参数功能一致，推荐使用fault_handling。</li><li>若推理任务订阅了故障信息，任务使用的推理卡上发生RestartRequest故障且故障持续时间未超过60秒，则不执行任务重调度；若故障持续时间超过60秒仍未恢复，则隔离芯片，进行任务重调度。</li></ul></div></div>|
|- fault_code|故障码，英文逗号拼接的字符串。|
|- fault_time_and_level_map|故障码、故障发生时间及故障处理等级。|
|UpdateTime|当前节点信息的更新时间，格式为时间戳，用于标识故障信息或设备状态的最新上报时间。|
|CmName|该ConfigMap的NAME，即该节点对应的配置在集群中的ConfigMap名称。|
|SuperPodID|超节点ID。|
|RackID|框ID。|
|ServerIndex|当前节点在超节点中的相对位置。<ul><li>驱动上报的SuperPodID或ServerIndex的值为0xffffffff时，SuperPodID或ServerIndex的取值为-1。</li><li>存在以下情况，SuperPodID或ServerIndex的取值为-2。</li><ul><li>当前设备不支持查询超节点信息。</li><li>因驱动问题导致获取超节点信息失败。</li></ul></ul>|

**表 3**  cluster-info-switch-$\{x\}

<a name="table9246232250"></a>

|参数|说明|
|--|--|
|FaultCode|当前节点的灵衢总线设备故障码列表。数组对象包含EventType、AssembledFaultCode、PeerPortDevice、PeerPortId、SwitchChipId、SwitchPortId、Severity、Assertion、AlarmRaisedTime等字段。|
|-EventType|告警ID。|
|-AssembledFaultCode|故障码。|
|-PeerPortDevice|对接设备类型。<ul><li>0：CPU</li><li>1：NPU</li><li>2：SW</li><li>0xFFFF：NA</li></ul>|
|-PeerPortId|对接设备ID。|
|-SwitchChipId|灵衢故障芯片ID。从0开始编号。|
|-SwitchPortId|灵衢故障端口ID。从0开始编号。|
|-Severity|故障等级。<ul><li>0：提示</li><li>1：次要</li><li>2：重要</li><li>3：紧急</li></ul>|
|-Assertion|事件类型。<ul><li>0：故障恢复</li><li>1：故障产生</li><li>2：通知类事件</li></ul>|
|-AlarmRaisedTime|故障/事件产生时间。|
|FaultLevel|当前节点故障处理等级。<p>取FaultCode中所有故障中等级最高的故障等级，取值包含：NotHandle、SubHealthFault、Separate和RestartRequest。</p>|
|UpdateTime|故障上报刷新时间。|
|NodeStatus|当前节点健康状态。<p>对应FaultLevel取值，NotHandle:Healthy、SubHealthFault:SubHealthy、Separate:UnHealthy和RestartRequest:UnHealthy。</p>|
|FaultTimeAndLevelMap|故障发生时间及故障处理等级列表。数组对象包含故障码、灵衢故障芯片ID、灵衢故障端口ID、fault_time和fault_level字段。键值为故障码、灵衢故障芯片ID、灵衢故障端口ID，由下划线连接组成。|
|-fault_time|故障发生时间。|
|-fault_level|故障处理等级。|

## statistic-fault-info<a name="section1153232554520"></a>

该ConfigMap位于用户创建的cluster-system命名空间下，Label为mc-statistic-fault=true。用于展示集群中的故障信息（当前仅展示公共故障信息）。

**表 4**  Data数据信息说明

|参数|说明|
|--|--|
|PublicFaults|公共故障详情。故障数量过大时，不再更新本字段内容。以下各字段的详细说明请参见[故障信息说明表](./03_public_fault_apis.md#configmap)。|
|-<i>\<node name></i>|故障节点名称|
|-resource|故障发送方<p>默认配置为CCAE、fd-online、pingmesh、Netmind。</p>|
|-devIds|故障芯片物理ID|
|-faultId|故障实例ID|
|-type|故障类型<ul><li>NPU：芯片故障。</li><li>Node：节点故障。</li><li>Network：网络故障。</li><li>Storage：存储故障。</li></ul>|
|-faultCode|故障码|
|-level|故障级别<ul><li>NotHandleFault：暂不处理。</li><li>SubHealthFault：亚健康。</li><li>SeparateNPU：无法恢复，需要隔离芯片。</li><li>PreSeparateNPU：暂不影响业务，后续不再调度任务到该芯片。</li></ul>|
|-faultTime|故障产生时间|
|FaultNum|故障数量|
|-publicFaultNum|所有节点的公共故障数量之和。|
|Description|公共故障数量过大时的提示信息。|

>[!NOTE]
>公共故障对外展示1M数据，大约4500条。超过4500条时，部分数据不再对外展示，ConfigMap中会新增Description内容进行提示，内部缓存正常运行。

## super-pod-<super-pod-id\><a name="section53741611135414"></a>

该ConfigMap位于用户创建的cluster-system命名空间下，Label为app=pingmesh。

**表 5**  super-pod-<super-pod-id\>

|参数|说明|
|--|--|
|app|NodeD识别ConfigMap所需的Label key，取值为pingmesh。|
|superPodDevice|超节点信息的key。|
|SuperPodID|超节点ID|
|NodeDeviceMap|超节点中包含的所有节点信息。|
|NodeName|节点名称|
|DeviceMap|节点中的所有NPU信息，格式为physicID: superDeviceID。|

## fault-job-info<a name="section1548342116513"></a>

该ConfigMap位于用户创建的cluster-system命名空间下。用于展示集群中需要强制释放通信资源的故障任务信息。仅在Atlas 900 A3 SuperPoD 超节点进行进程级别重调度时生效。

**表 6**  fault-job-info

|参数|说明|取值|
|--|--|--|
|SdIds|故障卡的SDID。|字符串序列|
|NodeNames|需要强制释放资源的节点名。|字符串序列|
|FaultTimes|发生故障的时间。|64位整数类型|
|JobId|任务的UID。|字符串|

## clusterd-manual-info-cm<a name="section15483421165190"></a>

该ConfigMap位于用户创建的cluster-system命名空间下。用于展示集群中人工隔离的芯片及故障信息。

示例如下：

```json
Name:         clusterd-manual-info-cm
Namespace:    cluster-system
Labels:       <none>
Annotations:  <none>
         
Data
====
localhost.localdomain:
----
{"Total":["Ascend910-0","Ascend910-2","Ascend910-3"],"Detail":{"Ascend910-0":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}],"Ascend910-2":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}],"Ascend910-3":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}]}}
         
Events:  <none> 
```

**表 7**  clusterd-manual-info-cm

|参数|说明|
|--|--|
|<i>localhost.localdomain</i>|节点名称，例如示例中的localhost.localdomain。|
|Total|故障的芯片名称。|
|Detail|芯片故障信息。|
|-<i>Ascend910-0</i>|芯片名称，例如示例中的Ascend910-0。|
|-FaultCode|故障码。|
|-FaultLevel|故障级别。|
|-LastSeparateTime|达到人工隔离频率时的最后一次故障时间。如果已经触发人工隔离芯片的故障，再一次达到了人工隔离频率，将刷新该时间。|
