# 公共故障接口<a name="ZH-CN_TOPIC_0000002479226838"></a>

## ConfigMap<a name="ZH-CN_TOPIC_0000002479386788"></a>

**功能说明<a name="section359310211618"></a>**

接收公共故障的ConfigMap信息，接入断点续训流程。

>[!NOTE]
>
>- 实际的ConfigMap中的参数如果与定义的取值范围不相符，ClusterD会将故障信息丢弃，不作处理。
>- 通过ConfigMap或者gRPC接口注入的公共故障，所有节点的故障数量之和上限为5w。当故障数量超过5w时，再次注入故障，ClusterD会将故障信息丢弃，不作处理。
>- ConfigMap的Label需要为mc-consumer-publicfault=true，Data的key需要为PublicFault。
>- 通过ConfigMap方式发送公共故障时，单次数据量不能超过1M大小，否则ConfigMap会更新失败。

**参数说明<a name="section4809204015614"></a>**

具体的参数说明见下表。

**表 1**  故障信息说明

|参数名称|含义|取值|类型|是否必填|
|--|--|--|--|--|
|id|消息唯一标识|8到128个字符的字符串，支持大小写字母、数字、中划线（-）、下划线（_）和点（.），保证唯一性。|string|是|
|timestamp|消息发送的时间戳|时间戳（单位：ms），13位数字，必须在2025-01-01T00:00:00Z之后。|int64|是|
|version|消息版本号|取值为1.0。|string|是|
|resource|故障发送方|默认配置为CCAE、fd-online、pingmesh、Netmind、dpcStorage。<ul><li>公共故障的故障发送方，必须存在于故障配置文件的publicFaultResource中。</li><li>对于新增的故障发送方，需要将其手动配置到故障配置文件中。详细说明请参见[（可选）配置公共故障的级别和发送方](../../usage/resumable_training/03_configuring_fault_detection_levels.md#可选配置公共故障的级别和发送方)。</li></ul>|string|是|
|faults|故障内容|切片，长度>0且≤100。|[]object, [fault](#fault0023698)|是|

**表 2**  fault字段说明

<a name="fault0023698"></a>

|参数名称|含义|取值|类型|是否必填|
|--|--|--|--|--|
|faultId|故障实例ID|8到128个字符的字符串，支持大小写字母、数字、中划线（-）、下划线（_）和点（.），保证唯一性。<p>同一个故障实例，faultId需要保证唯一性。</p>|string|是|
|faultType|故障类型|取值为NPU、Node、Network或Storage。<ul><li>NPU：芯片故障。</li><li>Node：节点故障。</li><li>Network：网络故障。</li><li>Storage：存储故障。</li></ul>该字段在cluster-info-cm中展示为“PublicFault”。|string|是|
|faultCode|故障码|用户可以自定义，9位唯一即可。<ul><li>接入断点续训的故障码，必须存在于故障配置文件的publicFaultCode中。</li><li>对于新增的故障码，需要在故障配置文件配置其故障级别。详细说明请参见[（可选）配置公共故障的级别和发送方](../../usage/resumable_training/03_configuring_fault_detection_levels.md#可选配置公共故障的级别和发送方)。</li><li>故障码建议遵循故障码说明表中的规则定义，方便后续维护。</li><li>若一张NPU先后出现两个相同的故障码，在cluster-info-cm中fault_code字段将同时记录2个相同的故障码。</li></ul>|string|是|
|faultTime|故障产生时间|时间戳（单位：ms），13位数字，必须在2025-01-01T00:00:00Z之后。<ul><li>无论是故障产生还是故障消除，该字段均为故障产生时间。</li><li>该字段在cluster-info-cm中以秒为单位展示。</li></ul>|int64|是|
|assertion|故障状态|取值为occur、recover或once。<ul><li>occur：故障产生。</li><li>recover：故障恢复。</li><li>once：一次性事件。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>公共故障消除需要将相应故障的recover事件写入ConfigMap中，不能通过删除ConfigMap的形式实现。</li><li>对于一次性事件，几秒钟之后故障会自动清除。</li></ul></div></div>|string|是|
|faultLocation|故障定位信息|故障源信息，长度≤10，map的key长度≤16，value长度≤128。eg. key: npuIp, value: ip|map[string]string|否|
|influence|故障影响的范围|切片，长度>0且≤1000。|[]object, [faultInfo](#faultinfo0023698)|是|
|description|故障描述|0~512个字符。包含非空白字符和空格。|string|否|

**表 3**  faultInfo字段说明

<a name="faultinfo0023698"></a>

|参数名称|含义|取值|类型|是否必填|
|--|--|--|--|--|
|nodeName|节点名称。可通过**kubectl get nodes -owide**命令查询。|1到253个字符的字符串，支持小写字母、数字、中划线（-）和点（.），必须以字母数字开头和结尾。该字段存在时，就不使用nodeSN。<p>如果节点名称不存在于K8s集群中，ClusterD不会提示节点名称错误，但是不会将该故障信息写入cluster-info-device-cm。</p>|string|nodeName与nodeSN二选一|
|nodeSN|节点SN号|节点的SN号。取值为NodeD写入的节点annotation，key为product-serial-number。<p>若使用该字段而不使用nodeName，需要提前安装NodeD组件。</p>|string|nodeName与nodeSN二选一|
|deviceIds|芯片物理ID|长度(0, 32]，每个元素的取值[0, 32)，且不允许重复。<ul><li>如果无法准确找到故障的芯片，需要填入节点上的所有芯片物理ID。</li><li>如果传入一个节点上不存在的芯片物理ID，ClusterD也会将其展示在cluster-info-device-cm中。</li></ul>|[]int32|是|

## gRPC接口<a name="ZH-CN_TOPIC_0000002479226854"></a>

**功能说明<a name="section125411749115817"></a>**

接收处理gRPC客户端的公共故障发送请求，接入断点续训流程。

>[!NOTE]
>
>- 实际的gRPC请求参数如果与定义的取值范围不相符，ClusterD会将故障信息丢弃，不作处理。
>- 通过ConfigMap或者gRPC接口注入的公共故障，所有节点的故障数量之和上限为5w。当故障数量超过5w时，再次注入故障，ClusterD会将故障信息丢弃，不作处理。
>- 公共故障消除需要将相应故障的recover事件通过gRPC接口发送给ClusterD。

**函数原型<a name="section1698941035919"></a>**

```proto
rpc SendPublicFault(PublicFaultRequest) returns (RespStatus){}
```

**输入参数说明<a name="section52771657118"></a>**

|参数|类型（Protobuf定义）|说明|
|--|--|--|
|PublicFaultRequest|<p>message PublicFaultRequest{<p>string id = 1;</p><p>int64 timestamp = 2;</p><p>string version = 3;</p><p>string resource = 4;</p><p>repeated Fault faults = 5;</p>}</p><p>message Fault{<p>string faultId = 1;</p><p>string faultType = 2;</p><p>string faultCode = 3;</p><p>int64 faultTime = 4;</p><p>string assertion = 5;</p><p>map<string, string> faultLocation = 6;</p><p>repeated PubFaultInfo influence = 7;</p><p>string description = 8;</p>}</p><p>message PubFaultInfo{<p>string nodeName = 1;</p><p>string nodeSN = 2;</p><p>repeated int32 deviceIds = 3;</p>}</p>|<p>**PublicFaultRequest.id**：消息唯一标识</p><p>**PublicFaultRequest.timestamp**：消息发送的时间戳</p><p>**PublicFaultRequest.version**：消息版本号</p><p>**PublicFaultRequest.resource**：故障发送方</p><p>**PublicFaultRequest.faults**：故障内容</p><p>**Fault.faultId**：故障实例ID</p><p>**Fault.faultType**：故障类型</p><p>**Fault.faultCode**：故障码</p><p>**Fault.faultTime**：故障产生时间</p><p>**Fault.assertion**：故障状态</p><p>**Fault.faultLocation**：故障定位信息</p><p>**Fault.influence**：故障影响的范围</p><p>**Fault.description**：故障描述</p><p>**PubFaultInfo.nodeName**：节点名称</p><p>**PubFaultInfo.nodeSN**：节点SN号</p><p>**PubFaultInfo.deviceIds**：芯片物理ID</p><p>以上参数的详细说明及取值情况请参见[ConfigMap](#configmap)。</p>|

**返回值说明<a name="section521319321415"></a>**

|返回值|类型（Protobuf定义）|说明|
|--|--|--|
|RespStatus|message RespStatus{<p>int32 code = 1;</p><p>string info = 2;</p>}|**RespStatus.code**：返回码。<ul><li>取值为0：表示故障发送成功。</li><li>其他值：表示故障发送失败。409表示请求参数有误，410表示消息发送频率超限。</li></ul>**RespStatus.info**：返回信息描述。|
