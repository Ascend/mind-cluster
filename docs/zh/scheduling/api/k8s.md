# K8s原生对象说明<a name="ZH-CN_TOPIC_0000002511346725"></a>

**Service label<a name="section17127184555719"></a>**

**表 1**  集群调度对Service的使用说明

|名称|作用|取值|使用组件|
|--|--|--|--|
|group-name|标记Pod对应的acjob的group名称|mindxdl.gitee.com|Volcano、Ascend Operator|
|job-name|标记Pod对应的acjob名称|字符串|Ascend Operator|
|replica-index|标记Pod序号（后续将删除）|[0-{Pod数量-1}]|Ascend Operator|
|replica-type|标记Pod类型（后续将删除）|<ul><li>master</li><li>chief</li><li>scheduler</li><li>worker</li></ul>|Ascend Operator|
|training.kubeflow.org/job-name|标记Pod对应的acjob名称|字符串|Ascend Operator|
|training.kubeflow.org/operator-name|标记创建Pod的operator名称|ascendjob-controller|Ascend Operator|
|training.kubeflow.org/replica-index|标记Pod序号|[0-{Pod数量-1}]|Ascend Operator|
|training.kubeflow.org/replica-type|标记Pod类型|<ul><li>master</li><li>chief</li><li>scheduler</li><li>worker</li></ul>|Ascend Operator|


**Job label<a name="section3960559173617"></a>**

**表 2**  集群调度对Job label使用说明

|Job label名称|作用|取值|使用组件|
|--|--|--|--|
|mind-cluster/scaling-rule: scaling-rule|标记扩缩容规则对应的ConfigMap名称。|字符串|Ascend Operator|
|mind-cluster/group-name: group0|标记扩缩容规则中对应的group名称。|字符串|Ascend Operator|


**Node label<a name="section121401114162912"></a>**

**表 3**  集群调度对Node label使用说明

|node label名称|作用|取值|使用组件|
|--|--|--|--|
|accelerator|标识节点的处理芯片|huawei-Ascend910、huawei-Ascend310、huawei-Ascend310P|Ascend Device Plugin|
|host-arch|标识节点的CPU架构|<ul><li>huawei-x86</li><li>huawei-arm</li></ul>|Volcano|
|masterselector|标识MindCluster的管理节点|dls-master-node|Volcano、Ascend Operator、Resilience Controller、ClusterD|
|node.kubernetes.io/npu.chip.name|上报当前芯片的具体类型|<ul><li>310</li><li>310P1</li><li>310P2</li><li>310P3</li><li>310P4</li><li>{xxx}A</li><li>910PremiumA</li><li>910ProA</li><li>910ProB</li><li>{xxx}Bx（x可取值为1、2、3、4）</li></ul>|Ascend Device Plugin<p><span>说明：</span></p>芯片型号的数值可通过**npu-smi info**命令查询，返回的“Name”字段对应信息为芯片型号，下文的{*xxx*}即取“910”字符作为芯片型号数值。|
|nodeDEnable|NodeD节点启动的开关|on|Volcano、Resilience Controller<p><span>说明：</span></p><ul><li>nodeDEnable=on标签表示启用NodeD的节点状态监测功能，用于获取节点的状态信息并用于判断节点是否故障。</li><li>取值为off或无该参数表示仅上报节点信息，不判断节点是否故障。</li><li>使用**容器化支持**或者**资源监测**时，可以不配置该标签；其他特性必须配置该标签。</li></ul>|
|workerselector|标识MindCluster的计算节点|dls-worker-node|Ascend Device Plugin、NodeD、NPU Exporter|
|accelerator-type|标识Atlas服务器类型|<ul><li>card</li><li>module</li><li>half</li><li>module-{xxx}b-8</li><li>module-{xxx}b-16</li><li>card-{xxx}b-2</li><li>card-{xxx}b-infer</li><li>module-a3-16</li><li>module-a3-16-super-pod</li></ul>|Ascend Device Plugin、Volcano|
|servertype|Atlas 200I SoC A1 核心板标识|<ul><li>soc</li><li>Ascend910-{aicore核数}</li><li>Ascend310P-{aicore核数}</li></ul>|Volcano、Ascend Device Plugin|
|huawei.com/Ascend910-Recover|Atlas 训练系列产品故障恢复标识|故障芯片ID|Ascend Device Plugin|
|huawei.com/Ascend910-NetworkRecover|Atlas 训练系列产品网络故障恢复标识|故障芯片ID|Ascend Device Plugin|
|infer-card-type|由Ascend Device Plugin写入，表明节点推理卡类型。|card-300i-duo|Volcano|
|mind-cluster/npu-chip-memory|芯片片上内存|mind-cluster/npu-chip-memory=64G|Volcano、Ascend Device Plugin|


**Pod  label<a name="section1019341142914"></a>**

**表 4** 集群调度组件对Pod  label使用说明

|名称|作用|取值|使用组件|
|--|--|--|--|
|ring-controller.atlas|标识Atlas的Pod|<li>ascend-910</li><li>ascend-{xxx}b</li><li>huawei.com/npu</li>|Ascend Device Plugin|
|vnpu-dvpp|标记Pod设置的DVPP|<li>yes：该Pod使用DVPP。</li><li>no：该Pod不使用DVPP。</li><li>null：默认值。不关注是否使用DVPP。</li>|Volcano|
|vnpu-level|标记选择虚拟化实例模板的等级|<li>low：低配，默认值。</li><li>high：性能优先。</li>|Volcano|
|version|标记Pod的版本|字符串|Ascend Operator|
|volcano.sh/job-name|标记Pod对应vcjob名称|字符串|Volcano|
|volcano.sh/job-namespace|标记Pod对应vcjob命名空间|字符串|Volcano|
|volcano.sh/queue-name|标记Pod对应queue名称|字符串|Volcano|
|volcano.sh/task-spec|标记Pod对应task名称|字符串|Volcano|
|fault-type|标记Pod故障处理策略|<ul><li>SubHealth</li><li>Separate</li></ul>|Volcano|
|deploy-name|标记Pod对应的deployment名称|字符串|Ascend Operator|
|group-name|标记Pod对应的acjob的group名称|mindxdl.gitee.com|Volcano、Ascend Operator|
|job-name|标记Pod对应的acjob名称|字符串|Ascend Operator|
|replica-index|标记Pod序号（后续将删除）|[0-{Pod数量-1}]|Ascend Operator|
|replica-type|标记Pod类型（后续将删除）|<ul><li>master</li><li>chief</li><li>scheduler</li><li>worker</li></ul>|Ascend Operator|
|training.kubeflow.org/job-name|标记Pod对应的acjob名称|字符串|Ascend Operator|
|training.kubeflow.org/job-role|标记Pod类型|master|Ascend Operator|
|training.kubeflow.org/operator-name|标记创建Pod的operator名称|ascendjob-controller|Ascend Operator|
|training.kubeflow.org/replica-index|标记Pod序号|[0-{Pod数量-1}]|Ascend Operator|
|training.kubeflow.org/replica-type|标记Pod类型|masterchiefschedulerworker|Ascend Operator|
|super-pod-affinity|超节点任务使用的亲和性调度策略|softhard|Ascend Operator、Volcano|


**Pod  annotation<a name="section16927154663513"></a>**

**表 5** 集群调度组件对Pod  annotation使用说明

|名称|作用|取值|使用组件|
|--|--|--|--|
|ascend.kubectl.kubernetes.io/ascend-910-configuration|Ascend Operator生成hccl.json的数据来源|字符串map|Ascend Device Plugin、Ascend Operator|
|super_pod_id|为Ascend Operator提供超节点ID信息|数字|Ascend Operator|
|hccl/rankIndex|断点续训中保持原rankId的依据|[0,1000]|Volcano、Ascend Operator|
|distributed-job|标记训练任务类型|<ul><li>true：当前任务为分布式任务</li><li>false：当前任务为单机任务</li></ul>|Volcano|
|huawei.com/Ascend910|Ascend Device Plugin为Pod分配芯片的依据|字符串|Volcano、Ascend Device Plugin|
|huawei.com/AscendReal|Ascend Device Plugin为Pod实际分配芯片的记录|字符串|Volcano、Ascend Device Plugin|
|huawei.com/npu-core|标记Pod使用的npu卡物理ID及切分模板|字符串|Volcano、Ascend Device Plugin|
|huawei.com/kltDev|kubelet为Pod分配芯片的记录|字符串|Ascend Device Plugin|
|huawei.com/recover_policy_path|任务重调度策略|pod：只支持Pod级重调度，不升级为Job级别|Volcano|
|huawei.com/schedule_minAvailable|任务能够调度的最小副本数|整数|Volcano|
|predicate-time|Ascend Device Plugin为Pod分配芯片的顺序依据|字符串|Volcano、Ascend Device Plugin|
|isSharedTor|标记Pod对应的交换机属性|整数|Volcano|
|isHealthy|标记Pod对应的交换机状态|整数|Volcano|
|scheduling.k8s.io/group-name|标记Pod对应podGroup名称|字符串|Volcano|
|volcano.sh/job-name|标记Pod对应的vcjob名称|字符串|Volcano|
|volcano.sh/job-version|标记Pod对应的vcjob版本|字符串|Volcano|
|volcano.sh/queue-name|标记Pod对应的queue版本|字符串|Volcano|
|volcano.sh/task-spec|标记Pod对应task名称|字符串|Volcano|
|volcano.sh/template-uid|标记Pod对应pod-template名称|字符串|Volcano|
|sharedTorIp|标记任务使用的共享交换机信息|字符串|Volcano、ClusterD|
|fault-job-delete|标记job的rank信息|字符串|Volcano|
|mind-cluster/hardware-type=800I-A2-xx|xx表示当前节点的片上内存，例如mind-cluster/hardware-type=800I-A2-64G|字符串|Volcano|
|super-pod-rank|任务的逻辑超节点rank|数字|Ascend Operator、Volcano|
|inHotSwitchFlow|标记当前Pod（故障Pod和备份Pod）处于亚健康热切流程中|true|ClusterD、Ascend Operator|
|backupNewPodName|标记当前故障Pod拉起的备份Pod名称|对应的备份Pod名称|ClusterD、Ascend Operator|
|backupSourcePodName|标记当前备份Pod对应的原Pod名称|对应的原Pod名称|Ascend Operator|
|needOperatorOpe|标记当前Pod需要Ascend Operator进行处理|<ul><li>create：需要Ascend Operator基于当前Pod创建备份Pod</li><li>delete：需要Ascend Operator删除当前Pod</li></ul>|ClusterD、Ascend Operator|
|needVolcanoOpe|标记当前Pod需要Volcano进行处理|delete：需要Volcano删除当前Pod|ClusterD、Volcano|
|podType|标记当前Pod是备份Pod|backup|ClusterD、Ascend Operator|


**Node annotation<a name="section9144358124519"></a>**

**表 6** 集群调度组件对Node annotation使用说明

|名称|作用|取值|使用组件|
|--|--|--|--|
|baseDeviceInfos|展示芯片的基础信息，例如IP，供Volcano调度时使用|字符串|Volcano|
|product-serial-number|NodeD通过IPMI接口获取节点SN号并写入annotation，供ClusterD接收公共故障时使用。|字符串|ClusterD|
|superPodID|表示该节点所属的超节点的ID。|字符串|ClusterD|
|ResetInfo|展示Ascend Device Plugin自动复位失败的芯片信息，如芯片的物理ID、Card ID等。|字符串|Ascend Device Plugin|


ResetInfo的内容格式如下所示。

```
{
    "ThirdPartyResetDevs": [
        {
            "CardId": 0,
            "DeviceId": 0,
            "AssociatedCardId": 4,
            "PhyID": 0,
            "LogicID": 0
        }
    ],
    "ManualResetDevs": [
        {
            "CardId": 1,
            "DeviceId": 0,
            "AssociatedCardId": 5,
            "PhyID": 2,
            "LogicID": 2
        }
    ]
}
```

**K8s的ServiceAccount<a name="section168254015405"></a>**

**表 7**  组件在K8s中创建的ServiceAccount列表

|账号名|说明|
|--|--|
|volcano-controllers|开源Volcano的controller组件在K8s中创建的用户。|
|volcano-scheduler|开源Volcano的scheduler组件在K8s中创建的用户。|
|<p>ascend-device-plugin-sa-910</p><p>ascend-device-plugin-sa-310p</p><p>ascend-device-plugin-sa-310</p>|使用YAML启动服务，将会在K8s中创建该用户，不同型号的设备使用的账号名不同。|
|ascend-operator-manager|使用YAML启动服务，将会在K8s中创建该用户，如：ascend-operator-v{version}.yaml。|
|resilience-controller|建议安全加固启动，使用带without-token的YAML启动服务，在K8s中创建并使用resilience-controller账号，同时为该账号授予适当权限。|
|noded|使用YAML启动服务，将会在K8s中创建该用户，如：noded-v{version}.yaml。|
|clusterd|使用YAML启动服务，将会在K8s中创建该用户，如：clusterd-v{version}.yaml。|
|default|MindCluster组件或开源Volcano部署时会在K8s中自动创建的用户。|


