需求背景
MindCluster当前对ManuallySeparateNPU类故障的隔离策略是将故障写入cm文件的ManuallySeparateNPU字段内。故障维修完成后需要手动将故障卡从cm文件中删除，增加了客户的操作和华为侧的解释成本。

提议方案
1、ManuallySeparateNPU故障发生后，支持配置解除隔离时间，默认0（表示不解除）。也可以手动提前清除。清除时长可配。
2、dp的配置中增加解除手动隔离时间字段，作为自动解除隔离的时间判断依据。不配置，则默认不主动解除手工隔离状态
3、在device-info-cm记录手工隔离的设备时，同时在cm中记录进入手工隔离状态的时间。 在到达解除隔离的时间后，删除cm中的手工隔离设备信息，同时在configmap中记录删除手工隔离信息事件。

目前的故障升级配置如下，针对每个故障码都可以配置其升级策略。根据和客户沟通，可以针对每个故障码配置其手工隔离的解除时间，比如通过字段ReleaseTimeWindow配置80E18005无故障24小时后自动解除升级。此处需要客户自行配置维修时间窗，防止过长或者过短。
```json
{
    "EventId": ["80E18005"],
    "TimeWindow": 86400,
    "Times": 3,
    "FaultHandling": "ManuallySeparateNPU",
    "ReleaseTimeWindow": 86400
}
```

ReleaseTimeWindow只在频率型故障处生效，持续型故障已有自动释放功能。ReleaseTimeWindow是根据最后一次故障消失时间开始计算，直到超过ReleaseTimeWindow就释放。
故障升级时，需要在DP日志中打印，因为哪几次故障导致的什么故障导致的升级，便于定位。
故障解除升级时，需要在DP日志中打印相关日志，同时在Event事件中添加解除升级事件。

device-info-cm设计：cm属于对外接口，需要维持兼容性设计，原有的ManuallySeparateNPU字段其示例如下,维持其格式不变，新增字段ManuallySeparateReason来展示被隔离的原因是什么，防止现网日志被冲刷后找不到原因。若一个故障码频率型和持续型都配置了，则记录SeparateTime是以频次超次那次故障发生时间为准。用户依然维持可以手工删除，还是删除ManuallySeparateNPU的值，删除后ManuallySeparateReason对应的键值也会被删除。
```
ManuallySeparateNPU:Ascend910-1
UpgradeFaultReason: // 针对每张卡的故障升级原因
"Ascend910-1": [{ // 每张卡可能配置了多个升级策略，此处展示针对每个升级策略的原因
    upgrade_time: 123456, // 故障升级时间
    fault_code: "xxx", // 什么故障码导致的升级
    fault_level: "ManuallySeparate"// 什么级别
    upgrade_type: "FaultAutofill" // "FaultFrequency"// "FaultDuration"
}],
"Ascend910-2": [{
    upgrade_time: 123456,
    fault_code: "a008",
    fault_level: "ManuallySeparate",
    upgrade_type: "FaultFrequency"
}]
```
真实示例
```
ManuallySeparateNPU: Ascend910-1
UpgradeFaultReason: '
{"Ascend910-1":[
    {
        "upgrade_time":1770433127279,
        "fault_code":"AutofillFaultCode",
        "fault_level":"ManuallySeparateNPU",
        "upgrade_type":"FaultAutofill"},
    {
        "upgrade_time":1770433162519,
        "fault_code":"81978008",
        "fault_level":"ManuallySeparateNPU",
        "upgrade_type":"FaultFrequency"}
    ]
}'
```
Dfx
1、 升级后的故障自动释放时，故障原因也会被删除。需要在Event事件中记录一下故障已经释放（event k8s只保留30min），event需要限流，通过限流处理器。
2、 如果在更新了faultCustom配置，已缓存的故障信息不会被清空。会按照新的升级策略来判断已缓存的故障信息是否要升级。

性能
在已有的流程中控制不增加K8s的访问，每轮更新设备信息只需要读一次CM，写一次CM。

兼容性
老版本CM中已有了ManuallySeparateNPU 26.0.0以及以后的DP需要读入后自动设置下UpgradeFaultReason 不涉及

异常场景测试
1、 频率型故障和持续型故障反复故障，根据故障的发生和消失时间与释放时间窗整体判断，可能会升级->降级->升级->降级。可能是维持一直处于升级状态，但是fault_time_and_level_map中的fault_time字段会刷新为最后一次故障时间，升级原因中的UpgradeTime（upgrade_time）也会不断刷新为最后的故障时间。
2、 针对老的CM格式，需要自动补齐Reason。