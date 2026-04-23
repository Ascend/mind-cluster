# 任务信息<a name="ZH-CN_TOPIC_0000002511426769"></a>

## job-summary-<任务名称\><a name="section24017282404"></a>

**表 1**  job-summary-任务名称 ConfigMap字段说明

|参数|说明|取值|
|--|--|--|
|hccl.json|任务使用的芯片通信信息。可转义为JSON格式，字段说明如下：<ul><li>status：任务RankTable是否已经生成。</li><ul><li>initializing：还在为任务分配设备，RankTable未生成。</li><li>complete：当RankTable生成后，状态会立即变为complete，同步出现server_list等其他字段。</li></ul><li>server_list：任务设备分配情况。</li><ul><li>device：记录NPU分配，NPU IP和rank_id信息。</li><ul><li>device_id：NPU 的设备 ID。</li><li>device_ip：NPU 的设备 IP。</li><li>rank_id：NPU 对应的训练 rank ID。</li><li>super_device_id：超节点内 NPU 的唯一标识。</li></ul><li>server_id：AI Server标识，全局唯一。</li><li>server_name：节点名称。</li><li>server_sn：节点的SN号。需要保证设备的SN存在。若不存在，请联系华为技术支持。</li><li>host_ip：主机 ip。</li><li>super_pod_id：超节点 id。</li><li>pod_name：pod 名称。</li><li>container_ids：pod 所有容器的 id 映射表。</li></ul><li>server_count：任务使用的节点数量。</li><li>version：版本信息。</li><li>total：configmap 个数。</li></ul>|字符串|
|job_id|任务的K8s ID信息。|字符串|
|operator|<ul><li>add：接收到添加任务命令后状态更新为add。</li><li>delete：接收到删除任务命令后状态更新为delete。</li></ul>|字符串|
|deleteTime|任务被删除的时间。|字符串|
|sharedTorIp|任务使用的共享交换机信息。|字符串|
|masterAddr|PyTorch训练时指定的MASTER_ADDR值。|字符串|
|total|ConfigMap的个数。|字符串|
|time|任务开始时间。|字符串|
|framework|任务使用的框架。|字符串|
|job_status|任务状态，存在以下几种状态。<ul><li>pending</li><li>running</li><li>complete</li><li>failed</li></ul>|字符串|
|job_name|任务名称。|字符串|
|cm_index|当前ConfigMap的序号。|字符串|
|sid|用户自定义任务 id|字符串|

## current-job-statistic<a name="section39901331194218"></a>

用于展示集群中当前任务的统计信息，详细信息记录在/var/log/mindx-dl/clusterd/event\_job.log日志文件中。由于K8s的ConfigMap容量大小限制，最大支持统计集群任务数量约为1w条。当日志文件达到20M时，触发自动转储，最多保存5份转储日志，转储日志最长保留时间为40天。

|参数|说明|
|--|--|
|data|-|
|- ID|K8s集群分配的Job ID。|
|- customID|用户自定义的Job ID，如果内容为空则不展示。|
|- cardNum|任务使用的卡的数量，如果内容为空则不展示。|
|- podFirstRunTime|任务Pod第一次全部running的时间，如果内容为空则不展示。|
|- stopTime|任务Pod全部complete或者被强行删除的时间，如果内容为空则不展示。|
|- podLastRunTime|任务Pod上一次全部恢复running的时间，如果内容为空则不展示。|
|- podLastFaultTime|任务Pod上一次部分或者全部failed的时间，如果内容为空则不展示。|
|- podFaultTimes|任务故障导致Pod重调度的次数，如果次数为0则不展示。|
|totalJob|当前集群中的总任务数。|

## scheduling-exception-report<a name="section_scheduling_exception_report"></a>

该ConfigMap位于cluster-system命名空间下。用于展示集群中调度异常的任务信息，帮助用户快速定位任务调度失败的原因。

**表 7**  scheduling-exception-report ConfigMap字段说明

|参数|说明|取值|
|--|--|--|
|\<jobName\>.\<jobUID\>|任务异常信息的key，由任务名称和任务UID组成。|字符串|
|- jobName|任务名称。|字符串|
|- jobType|任务类型，例如vcjob、acjob等。|字符串|
|- nameSpace|任务所在的命名空间。|字符串|
|- conditions|任务异常条件详情。|对象|
|-- status|任务状态。<ul><li>JobEmptyStatus：任务状态为空。</li><li>JobInitialized：任务已初始化。</li><li>JobFailed：任务失败。</li><li>PodGroupCreated：PodGroup已创建。</li><li>PodGroupPending：PodGroup处于Pending状态。</li><li>PodGroupInqueue：PodGroup处于Inqueue状态。</li><li>PodGroupUnknown：PodGroup状态未知。</li><li>PodGroupRunning：PodGroup处于Running状态。</li></ul>|字符串|
|-- reason|异常原因，例如JobEnqueueFailed、JobValidateFailed、NodePredicateFailed、BatchOrderFailed、NotEnoughResources、PodPending、PodFailed、PgNotInitialized、JobNoInitialized等。|字符串|
|-- message|异常详细信息，包含故障描述和排查建议。|字符串|
