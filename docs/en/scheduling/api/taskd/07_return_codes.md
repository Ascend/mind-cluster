# Return Code Description<a name="ZH-CN_TOPIC_0000002511426711"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:44:01.011Z pushedAt=2026-06-09T02:05:50.659Z -->

The TaskD return codes are shown in the following table.

**Table 1** TaskD return codes

|Return Code|Value|Description|
|--|--|--|
|OK|0|The API call is successful.|
|UnRegistry|400|The Job ID is not registered.|
|OrderMix|401|The request does not comply with the state machine order.|
|JobNotExist|402|The Job ID does not exist.|
|ProcessRescheduleOff|403|The process-level recovery is disabled.|
|ProcessNotReady|404|The training process is not started.|
|RecoverableRetryError|405|Recovery failed due to a device cleanup failure.|
|UnRecoverableRetryError|406|Recovery failed due to a device stop failure.|
|DumpError|407|Failed to save the dying gasps.|
|UnInit|408|Initialization is not called.|
|ClientError|499|Other failure reasons.|
|OutOfMaxServeJobs|500|The maximum number of service jobs has been exceeded.|
|OperateConfigMapError|501|Failed to operate ConfigMap.|
|OperatePodGroupError|502|Failed to operate PodGroup.|
|ScheduleTimeout|503|Pod scheduling timed out.|
|SignalQueueBusy|504|Failed to enqueue the control signal.|
|EventQueueBusy|505|Failed to enqueue the state machine event.|
|ControllerEventCancel|506|The state machine has exited.|
|WaitReportTimeout|507|Timed out waiting for the client to call the API.|
|WaitPlatStrategyTimeout|508|Timed out waiting for the AI platform to prepare the recovery strategy.|
|WriteConfirmFaultOrWaitPlatResultFault|509|AI platform fault information error.|
|ServerInnerError|599|Internal server error.|
