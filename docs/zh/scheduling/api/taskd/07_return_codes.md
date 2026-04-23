# 返回码说明<a name="ZH-CN_TOPIC_0000002511426711"></a>

TaskD返回码如下表所示。

**表 1** TaskD返回码

|返回码|值|含义|
|--|--|--|
|OK|0|接口调用正常。|
|UnRegistry|400|Job ID未注册。|
|OrderMix|401|请求不符合状态机顺序。|
|JobNotExist|402|Job ID不存在。|
|ProcessRescheduleOff|403|未打开进程级恢复开关。|
|ProcessNotReady|404|训练进程未拉起。|
|RecoverableRetryError|405|恢复失败，失败原因为clean device失败。|
|UnRecoverableRetryError|406|恢复失败，失败原因为stop device失败。|
|DumpError|407|临终遗言保存失败。|
|UnInit|408|未调用初始化。|
|ClientError|499|其他失败原因。|
|OutOfMaxServeJobs|500|超过最大服务任务数。|
|OperateConfigMapError|501|操作ConfigMap失败。|
|OperatePodGroupError|502|操作PodGroup失败。|
|ScheduleTimeout|503|Pod调度超时。|
|SignalQueueBusy|504|控制信号入队失败。|
|EventQueueBusy|505|状态机事件入队失败。|
|ControllerEventCancel|506|状态机已退出。|
|WaitReportTimeout|507|等待客户端调用接口超时。|
|WaitPlatStrategyTimeout|508|等待AI平台准备恢复策略超时。|
|WriteConfirmFaultOrWaitPlatResultFault|509|AI平台故障信息错误。|
|ServerInnerError|599|服务端内部错误。|
