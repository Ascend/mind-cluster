# Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511426737"></a>

## 芯片资源<a name="ZH-CN_TOPIC_0000002511346781"></a>

**mindx-dl-deviceinfo-<nodename\><a name="section11555858123711"></a>**

Ascend Device Plugin上报的NPU芯片信息如[表1](#table13817185391117)所示。

**表 1**  DeviceInfoCfg

<a name="table13817185391117"></a>
|名称|含义|说明|
|--|--|--|
|huawei.com/Ascend910|标记当前节点可用的芯片名称信息，存在多个时用英文逗号拼接。|该字段正在日落，后续版本该字段不再呈现。默认情况下，节点的可用芯片由Volcano维护，该字段不生效。如果需要生效，可以修改Volcano的配置参数“self-maintain-available-card”值为false。|
|huawei.com/Ascend910-NetworkUnhealthy|标记当前节点网络不健康的芯片名称信息，存在多个时用英文逗号拼接。|-|
|huawei.com/Ascend910-Unhealthy|标记当前芯片不健康的芯片名称信息，存在多个时用英文逗号拼接。|-|
|huawei.com/Ascend910-Recovering|标记当前节点正在进行恢复的芯片，存在多个时用英文逗号拼接。|-|
|huawei.com/Ascend910-Fault|记录芯片具体的故障信息。|数组对象，对象包含fault_type、npu_name、large_model_fault_level、fault_level、fault_handling、fault_code和fault_time_and_level_map这7个字段。|
|-fault_type|故障类型。|<ul><li>CardUnhealthy：芯片故障</li><li>CardNetworkUnhealthy：芯片网络故障</li><li>NodeUnhealthy：节点故障</li></ul>|
|-npu_name|故障的芯片名称，节点故障时为空。|字符串|
|<p>-large_model_fault_level</p><p>-fault_level</p><p>-fault_handling</p>|故障处理类型，节点故障时取值为空。|<ul><li>NotHandleFault：不做处理</li><li>RestartRequest：推理场景需要重新执行推理请求，训练场景重新执行训练业务</li><li>RestartBusiness：需要重新执行业务</li><li>FreeRestartNPU：影响业务执行，待芯片空闲时需复位芯片</li><li>RestartNPU：直接复位芯片并重新执行业务</li><li>SeparateNPU：隔离芯片</li><li>PreSeparateNPU：预隔离芯片，会根据训练任务实际运行情况判断是否重调度</li></ul>说明：<ul><li>large_model_fault_level、fault_level和fault_handling参数功能一致，推荐使用fault_handling。</li><li>若推理任务订阅了故障信息，任务使用的推理卡上发生RestartRequest故障且故障持续时间未超过60秒，则不执行任务重调度；若故障持续时间超过60秒仍未恢复，则隔离芯片，进行任务重调度。</li></ul>|
|-fault_code|故障码，英文逗号拼接的字符串。|Disconnected表示芯片网络不连通故障。heartbeatTimeOut表示节点状态丢失故障。|
|-fault_time_and_level_map|故障码、故障发生时间及故障处理等级。|-|
|SuperPodID|超节点ID。|字符串|
|ServerIndex|当前节点在超节点中的相对位置。|<ul><li>驱动上报的SuperPodID或ServerIndex的值为0xffffffff时，SuperPodID或ServerIndex的取值为-1。</li><li>存在以下情况，SuperPodID或ServerIndex的取值为-2。<ul><li>当前设备不支持查询超节点信息。</li><li>因驱动问题导致获取超节点信息失败。</li></ul></li></ul>|
|CheckCode|校验码。|-|

Ascend Device Plugin上报的灵衢总线设备故障信息如[表2](#table13455135662318)所示。

**表 2**  SwitchInfoCfg参数说明

<a name="table13455135662318"></a>
|名称|含义|说明|
|--|--|--|
|FaultCode|当前节点的灵衢总线设备故障码列表。|数组对象，包含EventType、AssembledFaultCode、PeerPortDevice、PeerPortId、SwitchChipId、SwitchPortId、Severity、Assertion、AlarmRaisedTime等字段。|
|-EventType|告警ID。|-|
|-AssembledFaultCode|故障码。|-|
|-PeerPortDevice|对接设备类型。|<ul><li>0：CPU</li><li>1：NPU</li><li>2：SW</li><li>0xFFFF：NA</li></ul>|
|-PeerPortId|对接设备ID。|-|
|-SwitchChipId|灵衢故障芯片ID。|从0开始编号。|
|-SwitchPortId|灵衢故障端口ID。|从0开始编号。|
|-Severity|故障等级。|<ul><li>0：提示</li><li>1：次要</li><li>2：重要</li><li>3：紧急</li></ul>|
|-Assertion|事件类型。|<ul><li>0：故障恢复</li><li>1：故障产生</li><li>2：通知类事件</li></ul>|
|FaultLevel|当前节点故障处理等级。|取FaultCode中所有故障中等级最高的故障等级，取值包含：NotHandle、SubHealthFault、Separate和RestartRequest。|
|UpdateTime|故障上报刷新时间。|-|
|NodeStatus|当前节点健康状态。|对应FaultLevel取值，NotHandle:Healthy、SubHealthFault:SubHealthy、Separate:UnHealthy和RestartRequest:UnHealthy。|
|FaultTimeAndLevelMap|故障发生时间及故障处理等级列表。|数组对象，包含故障码、灵衢故障芯片ID、灵衢故障端口ID、fault_time和fault_level字段。键值为故障码、灵衢故障芯片ID、灵衢故障端口ID，由下划线连接组成。|
|-fault_time|故障发生时间。|-|
|-fault_level|故障处理等级。|-|

Ascend Device Plugin的ConfigMap上报的人工干预的故障级别芯片信息如[表3](#table9710232)所示。

**表 3**  ManuallySeparateNPU说明

<a name="table9710232"></a>
|名称|含义|说明|
|--|--|--|
|ManuallySeparateNPU|因芯片多次故障，触发频率型故障升级策略，被ConfigMap记录到此键中。|多个芯片名称使用英文逗号分隔。|

Ascend Device Plugin的ConfigMap上报的故障策略升级原因如[表4](#table9710233)所示。

**表 4**  UpgradeFaultReason说明

<a name="table9710233"></a>
|名称|含义|说明|
|--|--|--|
|UpgradeFaultReason|故障码配置了频率型策略和持续型策略后，当触发故障升级时，记录故障升级原因和升级时间。|JSON的Map形式，键为芯片名称，值为导致该芯片故障升级的原因。|
|-fault_code|芯片故障升级的故障码。|-|
|-fault_level|升级后的故障级别。|-|
|-upgrade_type|故障升级类型。|<ul><li>频率型升级：FaultFrequency</li><li>持续型升级：FaultDuration</li><li>自动填充型升级：FaultAutofill</li></ul>|
|-upgrade_time|故障升级的时间点。|-|
|<p>注：</p><ul><li>当Ascend Device Plugin从26.0.0之前版本升级到26.0.0及之后版本时，若Ascend Device Plugin的ConfigMap中已有的芯片故障升级到ManuallySeparateNPU，则会为该芯片隔离自动填充原因。其中-fault_code值为AutofillFaultCode，-upgrade_type为FaultAutofill。</li><li>故障升级原因会随着故障降级而删除，同时删除事件会记录到K8s的kube-system命名空间下的event事件中。</li></ul>|

Ascend Device Plugin的ConfigMap中的描述信息如[表5](#table97108314503)所示。

**表 5**  Description说明

<a name="table97108314503"></a>
|名称|含义|说明|
|--|--|--|
|Description|描述信息。|此ConfigMap中的节点的可用芯片信息正在日落。默认情况下，节点的可用芯片由Volcano维护，此ConfigMap中维护的不生效。如果需要生效，可以修改Volcano的配置参数“self-maintain-available-card”值为false。|

Ascend Device Plugin上报的NPU设备故障信息如[表6](#table68216761214)所示。对象名称是<device-plugin-pod-name\>.<上报时间\><故障芯片ID\>，对象类型为Event。

>[!NOTE] 说明 
>下表仅展示与MindCluster业务相关的字段说明，更多字段的说明详细请参见[Event core](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#event-v1-core)。

**表 6**  NPU设备故障信息

<a name="table68216761214"></a>
|名称|含义|说明|
|--|--|--|
|type|事件的级别。|唯一值：Warning|
|message|事件的内容，包括节点名称、芯片编号、故障的产生或者恢复类型、故障码和故障级别信息。|字符串|
|reason|事件上报的原因。|<ul><li>Recovery：故障恢复</li><li>Occur：故障产生</li><li>Notice：一次性通知故障</li></ul>|
|action|故障的级别。|字符串。详细说明请参见<a href="#自定义芯片故障">表1</a>。|
|source|故障产生的源头。|结构体。表明故障产生的节点。|
|eventTime|故障产生的时间。|时间戳|
|involvedObject|故障绑定展示的对象。|结构体。通过Kind、Namespace和Name指向当前Ascend Device Plugin的Pod名称。指定后除了可以直接通过Event对象查询之外，查询当前的Pod详情时也能看到该事件。|
|reportingComponent|事件的控制者。|唯一值：device-plugin|
|reportingInstance|事件的上报实例。|字符串。取当前Ascend Device Plugin的Pod名称。|

**deviceNameCustomization.json<a name="section579455712489"></a>**

deviceNameCustomization.json支持自定义设备名称。编译Ascend Device Plugin镜像时，将该文件放在二进制包的同级目录下，即可将Ascend Device Plugin对外展示的资源类型、资源名称修改为自定义的名称。

**表 7**  deviceNameCustomization.json支持自定义设备名称

<a name="table76189511522121"></a>
|名称|含义|说明|
|--|--|--|
|ResourceType|设备的初始名称，必填。|仅支持Ascend910、Ascend310和Ascend310P中的一种。|
|DevicePublicType|设备对外展示的类型，例如huawei.com/Ascend910，必填。|仅支持xxx.xxx/xxx格式，xxx可以为大小写字母及数字，长度范围为10~32个字符。|
|DevicePublicNamePre|设备对外展示的名称前缀，例如Ascend910-。实际展示的名称，Ascend Device Plugin会在前缀后面拼接芯片的物理ID，必填。|可以包含大小写字母、中划线（-）、数字，必须以大小写字母开头，长度范围为2~16个字符。|
|PodConfigurationName|Pod的annotation上展示的挂载芯片信息详情，ResourceType为Ascend910时必填。|可以包含大小写字母、中划线（-）、/、点（.）、数字，必须以大小写字母开头，大小写字母数字结尾，长度范围为10~63个字符。|


## 任务信息<a name="ZH-CN_TOPIC_0000002479226860"></a>

**fault-config-<任务名称\><a name="section1786481083812"></a>**

**表 1**  fault-config-任务名称

<a name="table68216761214"></a>
|字段名称|含义|取值|备注|
|--|--|--|--|
|fault-npus|故障任务使用的故障芯片的rank信息。|字符串|-|
|checkCode|校验码。|字符串|-|

**reset-config-<任务名称\><a name="section3394547123916"></a>**

**表 2**  reset-config-_<job-name\>_

<a name="table1213115712136"></a>
<table><thead align="left"><tr id="row3132772132"><th class="cellrowborder" valign="top" width="15.950000000000001%" id="mcps1.2.6.1.1"><p id="p1022487193411"><a name="p1022487193411"></a><a name="p1022487193411"></a>字段名称</p>
</th>
<th class="cellrowborder" valign="top" width="14.69%" id="mcps1.2.6.1.2"><p id="p1313212741314"><a name="p1313212741314"></a><a name="p1313212741314"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="24.23%" id="mcps1.2.6.1.3"><p id="p513317151314"><a name="p513317151314"></a><a name="p513317151314"></a>含义</p>
</th>
<th class="cellrowborder" valign="top" width="28.82%" id="mcps1.2.6.1.4"><p id="p313315721314"><a name="p313315721314"></a><a name="p313315721314"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="16.31%" id="mcps1.2.6.1.5"><p id="p1313327191318"><a name="p1313327191318"></a><a name="p1313327191318"></a>备注</p>
</th>
</tr>
</thead>
<tbody><tr id="row41336711317"><td class="cellrowborder" rowspan="13" valign="top" width="15.950000000000001%" headers="mcps1.2.6.1.1 "><p id="p20565164533410"><a name="p20565164533410"></a><a name="p20565164533410"></a>reset.json</p>
<p id="p111446396589"><a name="p111446396589"></a><a name="p111446396589"></a></p>
<p id="p1811413311215"><a name="p1811413311215"></a><a name="p1811413311215"></a></p>
<p id="p0452951162310"><a name="p0452951162310"></a><a name="p0452951162310"></a></p>
</td>
<td class="cellrowborder" valign="top" width="14.69%" headers="mcps1.2.6.1.2 "><p id="p813420781315"><a name="p813420781315"></a><a name="p813420781315"></a>RankList</p>
</td>
<td class="cellrowborder" valign="top" width="24.23%" headers="mcps1.2.6.1.3 "><p id="p121346712134"><a name="p121346712134"></a><a name="p121346712134"></a>芯片列表</p>
</td>
<td class="cellrowborder" valign="top" width="28.82%" headers="mcps1.2.6.1.4 "><p id="p5134137121315"><a name="p5134137121315"></a><a name="p5134137121315"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="16.31%" headers="mcps1.2.6.1.5 "><p id="p1513427131320"><a name="p1513427131320"></a><a name="p1513427131320"></a>-</p>
</td>
</tr>
<tr id="row21341174135"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p171346791316"><a name="p171346791316"></a><a name="p171346791316"></a>-RankId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p3134177131313"><a name="p3134177131313"></a><a name="p3134177131313"></a>故障任务使用的Rank信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1413587161310"><a name="p1413587161310"></a><a name="p1413587161310"></a>整数类型</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1413511721318"><a name="p1413511721318"></a><a name="p1413511721318"></a>-</p>
</td>
</tr>
<tr id="row1713512717138"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p161352712139"><a name="p161352712139"></a><a name="p161352712139"></a>-LogicId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1135127181319"><a name="p1135127181319"></a><a name="p1135127181319"></a>芯片逻辑ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15135157131311"><a name="p15135157131311"></a><a name="p15135157131311"></a>32位整数类型</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p181366715137"><a name="p181366715137"></a><a name="p181366715137"></a>-</p>
</td>
</tr>
<tr id="row013914719136"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1313927191317"><a name="p1313927191317"></a><a name="p1313927191317"></a>-Status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8139177171315"><a name="p8139177171315"></a><a name="p8139177171315"></a>芯片状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul8436530113"></a><a name="ul8436530113"></a><ul id="ul8436530113"><li>unrecovered：未恢复</li><li>recovered：恢复成功</li><li>failed：恢复失败</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11394791316"><a name="p11394791316"></a><a name="p11394791316"></a>-</p>
</td>
</tr>
<tr id="row814016761315"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1814015719134"><a name="p1814015719134"></a><a name="p1814015719134"></a>-Policy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p11140676132"><a name="p11140676132"></a><a name="p11140676132"></a>热复位策略</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul1156918243817"></a><a name="ul1156918243817"></a><ul id="ul1156918243817"><li>empty：无故障</li><li>ignore：忽略故障</li><li>restart_request：重新执行当前请求</li><li>restart：重新执行训练任务</li><li>free_reset：当NPU上没有任务时，需要重启设备</li><li>reset：需要重启设备</li><li>isolate：需要隔离设备</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11140672134"><a name="p11140672134"></a><a name="p11140672134"></a>-</p>
</td>
</tr>
<tr id="row151401717139"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p101413711136"><a name="p101413711136"></a><a name="p101413711136"></a>-InitialPolicy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12141176132"><a name="p12141176132"></a><a name="p12141176132"></a>初始热复位策略</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul16378161213281"></a><a name="ul16378161213281"></a><ul id="ul16378161213281"><li>empty：无故障</li><li>ignore：忽略故障</li><li>restart_request：重新执行当前请求</li><li>restart：重新执行训练任务</li><li>free_reset：当NPU上没有任务时，需要重启设备</li><li>reset：需要重启设备</li><li>isolate：需要隔离设备</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p171419712133"><a name="p171419712133"></a><a name="p171419712133"></a>-</p>
</td>
</tr>
<tr id="row2141187121312"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3141197161312"><a name="p3141197161312"></a><a name="p3141197161312"></a>-ErrorCode</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19141576138"><a name="p19141576138"></a><a name="p19141576138"></a>十进制故障码</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p151429710139"><a name="p151429710139"></a><a name="p151429710139"></a>64位整型数组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1314257131311"><a name="p1314257131311"></a><a name="p1314257131311"></a>-</p>
</td>
</tr>
<tr id="row14142137191314"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p171421973132"><a name="p171421973132"></a><a name="p171421973132"></a>-ErrorCodeHex</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p0142577133"><a name="p0142577133"></a><a name="p0142577133"></a>十六进制故障码</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p8142177131320"><a name="p8142177131320"></a><a name="p8142177131320"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p31421070133"><a name="p31421070133"></a><a name="p31421070133"></a>-</p>
</td>
</tr>
<tr id="row41431139195820"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3233191110537"><a name="p3233191110537"></a><a name="p3233191110537"></a>GracefulExit</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p321511543920"><a name="p321511543920"></a><a name="p321511543920"></a>管理训练进程</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p33363655012"><a name="p33363655012"></a><a name="p33363655012"></a>0或1</p>
<a name="ul7532185975011"></a><a name="ul7532185975011"></a><ul id="ul7532185975011"><li>取值为1，杀死所有训练进程</li><li>取值为0，不做处理</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p921615511390"><a name="p921615511390"></a><a name="p921615511390"></a>-</p>
</td>
</tr>
<tr id="row167775084714"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p108353401829"><a name="p108353401829"></a><a name="p108353401829"></a>UpdateTime</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p118356401224"><a name="p118356401224"></a><a name="p118356401224"></a>ConfigMap的更新时间</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p58359402214"><a name="p58359402214"></a><a name="p58359402214"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p7835114017213"><a name="p7835114017213"></a><a name="p7835114017213"></a>-</p>
</td>
</tr>
<tr id="row189371153471"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p2066862744"><a name="p2066862744"></a><a name="p2066862744"></a>RetryTime</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p126681521149"><a name="p126681521149"></a><a name="p126681521149"></a>Pod重调度的次数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p66683214418"><a name="p66683214418"></a><a name="p66683214418"></a>整数类型</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p18668142844"><a name="p18668142844"></a><a name="p18668142844"></a>-</p>
</td>
</tr>
<tr id="row13113203322"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1254115251666"><a name="p1254115251666"></a><a name="p1254115251666"></a>FaultFlushing</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p7541192512618"><a name="p7541192512618"></a><a name="p7541192512618"></a>告知<span id="ph14256162281217"><a name="ph14256162281217"></a><a name="ph14256162281217"></a>Elastic Agent</span>当前是否有故障正在刷新</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p13813147101216"><a name="p13813147101216"></a><a name="p13813147101216"></a>取值为true或false</p>
<a name="ul1563191521213"></a><a name="ul1563191521213"></a><ul id="ul1563191521213"><li>true：表示有故障正在刷新</li><li>false：表示当前无故障刷新</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p19951631131314"><a name="p19951631131314"></a><a name="p19951631131314"></a><span id="ph952618296564"><a name="ph952618296564"></a><a name="ph952618296564"></a>Elastic Agent</span>需要等待该字段为false且故障RankList无本节点故障时才会拉起训练进程</p>
</td>
</tr>
<tr id="row18452151202319"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p64521951162319"><a name="p64521951162319"></a><a name="p64521951162319"></a><span>RestartFaultProcess</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17453851172311"><a name="p17453851172311"></a><a name="p17453851172311"></a><span>告知</span><span id="ph262783362516"><a name="ph262783362516"></a><a name="ph262783362516"></a>Elastic Agent</span><span>当前是否仅重启本节点故障进程</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2012431813258"><a name="p2012431813258"></a><a name="p2012431813258"></a><span>取值true或false</span></p>
<a name="ul5650162018256"></a><a name="ul5650162018256"></a><ul id="ul5650162018256"><li><span>true：表示不退出</span><span id="ph8849103812259"><a name="ph8849103812259"></a><a name="ph8849103812259"></a>Elastic Agent</span><span>，仅重启本节点故障进程</span></li><li><span>false：当本节点有故障进程时，退出</span><span id="ph1888614312613"><a name="ph1888614312613"></a><a name="ph1888614312613"></a>Elastic Agent</span></li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p94534513233"><a name="p94534513233"></a><a name="p94534513233"></a>-</p>
</td>
</tr>
<tr id="row859053413417"><td class="cellrowborder" valign="top" width="15.950000000000001%" headers="mcps1.2.6.1.1 "><p id="p844941513297"><a name="p844941513297"></a><a name="p844941513297"></a>restartType</p>
</td>
<td class="cellrowborder" valign="top" width="14.69%" headers="mcps1.2.6.1.2 "><p id="p220916992912"><a name="p220916992912"></a><a name="p220916992912"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="24.23%" headers="mcps1.2.6.1.3 "><p id="p1820909182911"><a name="p1820909182911"></a><a name="p1820909182911"></a>reset.json更新的类型</p>
</td>
<td class="cellrowborder" valign="top" width="28.82%" headers="mcps1.2.6.1.4 "><p id="p15209596295"><a name="p15209596295"></a><a name="p15209596295"></a>podReschedule或hotReset</p>
</td>
<td class="cellrowborder" valign="top" width="16.31%" headers="mcps1.2.6.1.5 "><p id="p95471047133013"><a name="p95471047133013"></a><a name="p95471047133013"></a>单pod重调度情况下取值为podReschedule，热恢复场景下取值为hotReset</p>
</td>
</tr>
<tr id="row165081157153910"><td class="cellrowborder" valign="top" width="15.950000000000001%" headers="mcps1.2.6.1.1 "><p id="p750805713392"><a name="p750805713392"></a><a name="p750805713392"></a>checkCode</p>
</td>
<td class="cellrowborder" valign="top" width="14.69%" headers="mcps1.2.6.1.2 "><p id="p0508145711393"><a name="p0508145711393"></a><a name="p0508145711393"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="24.23%" headers="mcps1.2.6.1.3 "><p id="p250845783917"><a name="p250845783917"></a><a name="p250845783917"></a>校验码</p>
</td>
<td class="cellrowborder" valign="top" width="28.82%" headers="mcps1.2.6.1.4 "><p id="p750835713919"><a name="p750835713919"></a><a name="p750835713919"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="16.31%" headers="mcps1.2.6.1.5 "><p id="p175081157113917"><a name="p175081157113917"></a><a name="p175081157113917"></a>-</p>
</td>
</tr>
</tbody>
</table>

**data-trace-<任务名称\><a name="section19954856135618"></a>**

存储当前任务的各类打点类型的开关状态，由Ascend Device Plugin挂载到计算节点存储，训练容器挂载该文件后，由TaskD读取后对各类打点数据进行开关。

**表 3**  data-trace-<任务名称\> ConfigMap字段说明

<a name="table97521457610"></a>
|字段名称|含义|取值|类型|
|--|--|--|--|
|Communication|标识通信算子。|on/off|string|
|Step|标识Step时延。|on/off|string|
|SaveCheckpoint|标识SaveCheckpoint耗时。|on/off|string|
|FP|标识前向传播数据。|on/off|string|
|DataLoader|标识DataLoader耗时。|on/off|string|

>[!NOTE] 说明 
>-   该ConfigMap需要和训练任务在同一命名空间，且命名为data-trace-<任务名称\>，包括标签reset=true。
>-   该ConfigMap由Ascend Device Plugin挂载到训练节点的/user/cluster-info/datatrace-config/命名空间.data-trace-任务名称/\*的文件夹下，文件名为profilingSwitch。
>-   如用户未创建该ConfigMap，在首次调用gRPC接口ModifyTrainingDataTraceSwitch时，ClusterD将尝试自动创建该ConfigMap。
>-   用户如需使用该功能，应将节点上的profilingSwitch文件，使用hostPath方式挂载进入容器内的/user/cluster-info/datatrace-config/目录。
>-   当前Step、SaveCheckpoint、FP、DataLoader为默认开启，且四类只能同步开启关闭，当五类数据全为off时关闭所有打点，否则默认开启上述四类，同时根据通信算子开关状态对其进行开启或关闭。

**steptime-dtpgroup<a name="section1146122513469"></a>**

存储任务的迭代时延和分组信息的保存路径和启停开关，启动任务时用户可通过CCAE管理平台配置ConfigMap参数进行任务是否劣化的判定。

**表 4**  steptime-dtpgroup ConfigMap字段说明

<a name="table3610611144615"></a>
|一级参数名称|二级参数名称|含义|取值|备注|
|--|--|--|--|--|
|data|PerfDumpPath|迭代时延和分组信息保存路径。|字符串|-|
|-|PerfDumpConfig|迭代时延和分组信息启停开关。|字符串|-|


## 自定义芯片故障<a name="ZH-CN_TOPIC_0000002511346805"></a>

**faultCode.json中的故障级别<a name="section579455712489"></a>**

断点续训针对芯片故障的不同级别进行分级处理。若用户需要修改故障码的故障级别，操作指导请参见[（可选）配置芯片故障级别](../usage/resumable_training.md#可选配置芯片故障级别)。

Ascend Device Plugin从驱动获取到芯片故障码后，将根据故障码对设备及业务的影响将故障划分为以下几种级别，详细说明请参见[表1](#table7618951152212)。

**表 1**  故障级别及处理说明

<a name="table7618951152212"></a>
<table><thead align="left"><tr id="row461812518228"><th class="cellrowborder" valign="top" width="19.06%" id="mcps1.2.5.1.1"><p id="p12618851162220"><a name="p12618851162220"></a><a name="p12618851162220"></a>故障处理策略</p>
</th>
<th class="cellrowborder" valign="top" width="35.78%" id="mcps1.2.5.1.2"><p id="p16618125162219"><a name="p16618125162219"></a><a name="p16618125162219"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="20.349999999999998%" id="mcps1.2.5.1.3"><p id="p1163819316544"><a name="p1163819316544"></a><a name="p1163819316544"></a>重调度处理</p>
</th>
<th class="cellrowborder" valign="top" width="24.81%" id="mcps1.2.5.1.4"><p id="p171971327125410"><a name="p171971327125410"></a><a name="p171971327125410"></a>优雅容错处理</p>
</th>
</tr>
</thead>
<tbody><tr id="row961811511228"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p7618125114229"><a name="p7618125114229"></a><a name="p7618125114229"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="p1261835110227"><a name="p1261835110227"></a><a name="p1261835110227"></a>对业务无影响的故障，无需处理。</p>
</td>
<td class="cellrowborder" valign="top" width="20.349999999999998%" headers="mcps1.2.5.1.3 "><p id="p10638123115414"><a name="p10638123115414"></a><a name="p10638123115414"></a>暂不处理</p>
</td>
<td class="cellrowborder" valign="top" width="24.81%" headers="mcps1.2.5.1.4 "><p id="p719714273546"><a name="p719714273546"></a><a name="p719714273546"></a>暂不处理</p>
</td>
</tr>
<tr id="row116184515226"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p5618751102216"><a name="p5618751102216"></a><a name="p5618751102216"></a>RestartRequest</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="p05771854113911"><a name="p05771854113911"></a><a name="p05771854113911"></a>影响业务执行，需要重新执行业务请求。</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="20.349999999999998%" headers="mcps1.2.5.1.3 "><p id="p13855131912555"><a name="p13855131912555"></a><a name="p13855131912555"></a>隔离芯片，进行任务重调度</p>
<div class="note" id="note11901123612819"><a name="note11901123612819"></a><a name="note11901123612819"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="zh-cn_topic_0000002479386448_p1069261722310"><a name="zh-cn_topic_0000002479386448_p1069261722310"></a><a name="zh-cn_topic_0000002479386448_p1069261722310"></a>若推理任务订阅<span id="zh-cn_topic_0000002479386448_ph4356222144812"><a name="zh-cn_topic_0000002479386448_ph4356222144812"></a><a name="zh-cn_topic_0000002479386448_ph4356222144812"></a>了</span>故障信息，任务使用的推理卡上发生RestartRequest故障且故障持续时间未超过60秒，则不执行任务重调度；若故障持续时间超过60秒仍未恢复，则隔离芯片，进行任务重调度。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="24.81%" headers="mcps1.2.5.1.4 "><p id="p9145165785517"><a name="p9145165785517"></a><a name="p9145165785517"></a>推理场景重新执行推理请求，训练场景重新执行训练业务</p>
</td>
</tr>
<tr id="row14618105116225"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p15618851132212"><a name="p15618851132212"></a><a name="p15618851132212"></a>RestartBusiness</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p3618851182216"><a name="p3618851182216"></a><a name="p3618851182216"></a>影响业务执行，需要重新执行业务。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1419712272549"><a name="p1419712272549"></a><a name="p1419712272549"></a>重新执行业务</p>
</td>
</tr>
<tr id="row561825132214"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p66188511222"><a name="p66188511222"></a><a name="p66188511222"></a>FreeRestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p661865162211"><a name="p661865162211"></a><a name="p661865162211"></a>影响业务执行，待芯片空闲时需复位芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p178789204535"><a name="p178789204535"></a><a name="p178789204535"></a>等待芯片空闲后复位芯片</p>
</td>
</tr>
<tr id="row14618125152210"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17618155116227"><a name="p17618155116227"></a><a name="p17618155116227"></a>RestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p108302057102114"><a name="p108302057102114"></a><a name="p108302057102114"></a>影响业务执行，需立即复位芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p969972925312"><a name="p969972925312"></a><a name="p969972925312"></a>立即停止训练业务，复位芯片后重新执行业务</p>
</td>
</tr>
<tr id="row1061895115227"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p961885142215"><a name="p961885142215"></a><a name="p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p18618151202216"><a name="p18618151202216"></a><a name="p18618151202216"></a>无法恢复，需要隔离芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p019742745411"><a name="p019742745411"></a><a name="p019742745411"></a>隔离芯片，进行任务重调度</p>
</td>
</tr>
<tr id="row1930365771212"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p16227299522"><a name="zh-cn_topic_0000002171521445_p16227299522"></a><a name="zh-cn_topic_0000002171521445_p16227299522"></a>PreSeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p546081915499"><a name="zh-cn_topic_0000002171521445_p546081915499"></a><a name="zh-cn_topic_0000002171521445_p546081915499"></a>暂不影响业务，后续不再调度任务到该芯片。</p>
</td>
<td class="cellrowborder" valign="top" width="20.349999999999998%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p222102912521"><a name="zh-cn_topic_0000002171521445_p222102912521"></a><a name="zh-cn_topic_0000002171521445_p222102912521"></a>预隔离芯片</p>
</td>
<td class="cellrowborder" valign="top" width="24.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002171521445_p12221329155217"><a name="zh-cn_topic_0000002171521445_p12221329155217"></a><a name="zh-cn_topic_0000002171521445_p12221329155217"></a>预隔离芯片</p>
</td>
</tr>
<tr id="row89346317136"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002171521445_p835213245522"><a name="zh-cn_topic_0000002171521445_p835213245522"></a><a name="zh-cn_topic_0000002171521445_p835213245522"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.78%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002171521445_p1354813311915"></a><a name="zh-cn_topic_0000002171521445_p1354813311915"></a>根据任务YAML中配置的subHealthyStrategy参数取值进行处理，详细请参见<a href="../api/ascend_operator.md">Ascend Operator</a>中YAML参数说明（acjob任务）。</p>
</td>
<td class="cellrowborder" valign="top" width="20.349999999999998%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002171521445_p3352524125220"></a><a name="zh-cn_topic_0000002171521445_p3352524125220"></a>当芯片出现亚健康故障时，需根据<a href="../usage/resumable_training.md#任务yaml配置示例">配置yaml</a>策略进行处理。</p>
<div class="note" id="zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002171521445_note7936204710536"></a><a name="zh-cn_topic_0000002171521445_note7936204710536"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002171521445_p15222114115810"></a><a name="zh-cn_topic_0000002171521445_p15222114115810"></a>如果后续芯片出现其他级别故障，此时SubHealthFault</p>
<p id="zh-cn_topic_0000002171521445_p109369476532"><a name="zh-cn_topic_0000002171521445_p109369476532"></a><a name="zh-cn_topic_0000002171521445_p109369476532"></a>处理策略不影响其他级别的故障处理。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="24.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002171521445_p8352172425218"><a name="zh-cn_topic_0000002171521445_p8352172425218"></a><a name="zh-cn_topic_0000002171521445_p8352172425218"></a>根据策略进行处理</p>
</td>
</tr>
</tbody>
</table>

**faultCustomization.json参数说明<a name="section33036167576"></a>**

用户不手动修改faultCustomization.json文件时，Ascend Device Plugin按照faultCustomization.json的默认配置（默认值）进行故障处理。

**表 2**  faultCustomization.json文件参数说明

<a name="table1519814413572"></a>
|一级参数名称|二级参数名称|说明|
|--|--|--|
|GraceTolerance|-|优雅容错相关配置。<p>GraceTolerance及其子参数不存在或者超出取值范围，则使用默认值。</p>|
|-|WaitProcessReadCMTime|使用优雅容错模式时，等待管理进程读取ConfigMap文件的时间，单位为秒，取值范围为5~90，默认值为30。|
|-|WaitDeviceResetTime|使用优雅容错模式时，等待芯片重启的最大时长，单位为秒，取值范围为60~180，默认值为150。|
|-|WaitFaultSelfHealingTime|使用优雅容错模式时，等待RestartBusiness级别故障恢复时间，单位为秒，取值范围为1~30，默认值为15。|
|FaultFrequency|-|自定义故障频率，即某一故障在时间窗口内出现次数达到次数上限时，根据配置的故障处理策略进行处理。<ul><li>FaultFrequency及其子参数取值范围不正确，则忽略该条配置。</li><li>FaultFrequency及其子参数数据格式不正确，则会使用默认配置。</li></ul>|
|-|EventId|故障码ID。<p>每个故障码（EventId）只允许配置一个FaultFrequency参数，如果配置了多个，则只有第一条正确的会生效。</p>|
|-|TimeWindow|时间窗口，即统计当前时间减去TimeWindow的时间至当前时间，这段时间范围内的故障次数，单位为秒，取值范围为60~864000。|
|-|Times|任务支持的断点续训最大次数，即同一个故障出现的次数上限，取值范围为1~100。如果在时间窗口内该故障出现次数大于或等于该值，则按照FaultHandling中定义的策略处理和上报。|
|-|FaultHandling|<p>达到断点续训最大次数后故障的处理策略，支持配置不同级别的故障处理策略。其中只有ManuallySeparateNPU级别会持续维持，若配置了ReleaseTimeWindow，则可达到条件后自动释放。其他故障处理级别只会升级处理一次，然后自动降级为故障原级别。</p><ul><li>PreSeparateNPU：大模型的故障处理策略。该故障处理模式为预隔离芯片，根据训练任务实际运行情况判断是否重调度。</li><li>ManuallySeparateNPU：需人工干预的故障处理策略。<ul><li>出现该策略时，将直接上报K8s该芯片不健康并将芯片名字写入device-info-cm。</li><li>芯片名称只要保存于该字段中，即使故障恢复也仍然隔离芯片，直到运维人员手动在该字段中删除芯片名称。</li><li>该字段只允许Ascend Device Plugin新增或修改，维护人员只能删除该字段中的芯片名称。</li><li>faultCode.json暂不支持该策略。</li></ul></li></ul>|
|-|ReleaseTimeWindow|当频率型故障处理策略为ManuallySeparateNPU时，若故障已经恢复，并持续超过ReleaseTimeWindow的时间窗没有再发送该故障，则ManuallySeparateNPU自动降级。该参数的取值范围为60~int32最大值，单位为秒。若不配置该参数，则表示策略升级为ManuallySeparateNPU后不降级。|
|FaultDuration|-|自定义故障超时策略，当某一故障持续时间达到配置上限时，该故障会按照指定的故障处理策略进行处理。<ul><li>FaultDuration及其子参数取值范围不正确，则忽略该条配置。</li><li>FaultDuration及其子参数数据格式不正确，则会使用默认配置。</li></ul>|
|-|EventId|故障ID。<p>每个故障码（EventId）只允许配置一个FaultDuration参数，如果配置了多个，则只有第一条正确的会生效。</p>|
|-|FaultTimeout|故障持续时间超过该值，则按照FaultHandling中定义的故障处理策略进行处理，单位为秒，取值范围为0~600，默认值说明如下。<ul><li>故障ID为81078603的参数面网络故障默认值为20。</li><li>故障ID为80E01801的片上内存多Bit故障默认值为30。</li><li>其余故障默认值为0。</li></ul>|
|-|RecoverTimeout|故障恢复时间超过该值，则上报故障恢复，单位为秒，取值范围为0~86400，默认值说明如下。<ul><li>故障ID为81078603的参数面网络故障默认值为60。不建议设置为0，建议大于listWatchPeriod健康状态检查周期。关于listWatchPeriod的详细说明请参见<a href="../installation_guide.md#ascend-device-plugin">Ascend Device Plugin</a>中"Ascend Device Plugin启动参数"表。</li><li>其余故障默认值为0。</li></ul>|
|-|FaultHandling|<p>超过故障持续时间后的故障处理策略，支持配置不同级别的故障处理策略，同时还支持配置PreSeparateNPU故障处理策略。</p><p>超过故障持续时间后的故障处理策略，建议高于故障本身的故障处理策略，否则配置不生效。</p><p>不支持配置ManuallySeparateNPU策略，配置不生效。</p>|
|注：<ul><li>如果一个故障码同时配置了故障频率（FaultFrequency）和故障超时策略（FaultDuration），该故障码在TimeWindow时间窗口中超时次数达到任务支持的最大次数，则采用以下三者中最严重的等级进行处理。这三者分别为：故障本身的故障处理策略、FaultFrequency和FaultDuration中配置的故障处理策略。</li><li>如果一个故障码同时配置了故障频率和故障超时策略，只有当故障超时后，故障频次才会增加一次。故障恢复超过RecoverTimeout才算恢复，恢复后再次故障超时才能累积下一次计数。</li><li>故障ID为81078603的网络故障只支持配置为NotHandleFault、PreSeparateNPU或SeparateNPU三种故障处理策略，若配置为其他策略则使用默认配置NotHandleFault。</li><li>当Ascend Device Plugin从26.0.0之前版本升级到26.0.0及之后版本时，若Ascend Device Plugin的ConfigMap中已经包含了ManuallySeparateNPU键值，则其降级的时间窗为faultCustomization.json中最大的ReleaseTimeWindow值，若没有任何故障码配置ReleaseTimeWindow，则ConfigMap已有的ManuallySeparateNPU不降级。</li></ul>|


## 自定义灵衢设备故障<a name="ZH-CN_TOPIC_0000002511426735"></a>

断点续训针对灵衢总线设备故障的不同级别进行分级处理。若用户需要修改故障码的故障级别，操作指导请参见[（可选）配置总线设备故障级别](../usage/resumable_training.md#可选配置总线设备故障级别)。

Ascend Device Plugin从驱动获取到故障码后，将根据故障码对设备及业务的影响将故障划分为以下五种级别并进行相应的重调度处理，详细说明请参见[表1](#table212253274720)。

**表 1**  故障级别及处理说明

<a name="table212253274720"></a>
|故障类型|说明|重调度处理|
|--|--|--|
|NotHandleFault|暂不影响业务，可以自行恢复，无需处理。|暂不处理。|
|SubHealthFault|影响业务运行性能，需要排查亚健康原因。|当出现亚健康故障时，需根据<a href="../api/ascend_operator.md">Ascend Operator</a>中"YAML参数说明（acjob任务）"中subHealthyStrategy参数所指定的亚健康策略进行处理。|
|RestartRequestFault|业务运行失败，需要重新执行业务请求。|停止当前训练任务，隔离节点，进行任务重调度。|
|ResetFault|业务运行失败。|停止当前训练任务，隔离节点，进行任务重调度。|
|SeparateFault|业务运行失败，需更换器件或板卡。|停止当前训练任务，隔离节点，进行任务重调度。|


