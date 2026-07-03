# Volcano<a name="ZH-CN_TOPIC_0000002479226814"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:44:29.569Z pushedAt=2026-06-09T02:05:50.675Z -->

## Obtaining Cluster Scheduling Component Information<a name="ZH-CN_TOPIC_0000002479386860"></a>

- The VolcanoJob API is provided by the open-source component Volcano. MindCluster has modified the `Annotations` field of the VolcanoJob API, as shown in [Table 1](#table177621954014). Other APIs remain unchanged. For details about open-source Volcano, see the Volcano open-source community.

    **Table 1**  Annotations

    <a name="table177621954014"></a>

    |Parameter|Description|Value|
    |--|--|--|
    |distributed|Written and used by Resilience Controller to mark whether a job is a distributed task.|true|

- The following describes the interfaces exposed by the volcano-scheduler and volcano-controller component Pods (defined by the open-source components themselves).

    **Table 2** List of interfaces exposed by MindCluster Volcano

    <a name="zh-cn_topic_0000001446965056_table173071368477"></a>

    |Access Method|Protocol|Method|Description|Component|
    |--|--|--|--|--|
    |`http://podIP:11251/healthz`|http|Get|Health check port|volcano-controller|
    |`http://podIP:11251/healthz`|http|Get|Health check port|volcano-scheduler|
    |`http://volcano-scheduler-serviceIP:8080/metrics`|http|Get|Prometheus metrics collection port|volcano-scheduler|

    >[!NOTE]
    >- To ensure normal access to the Volcano health check port and the Prometheus metrics collection port, set the `--enable-healthz` and `--enable-metrics` parameters to `true` in the YAML file when installing Volcano. For details about how to modify these parameters, see "Step 7" in [Installing Volcano](../developer_guide/installation_deployment/manual_installation/05_volcano.md).
    >- The HUAWEI CLOUD CCI service provides more detailed VolcanoJob instructions. For details, see the "[Creating a VolcanoJob](https://support.huaweicloud.com/intl/en-us/api-cci/createBatchVolcanoShV1alpha1NamespacedJob.html)" section in *Cloud Container Instance*.

## PodGroup<a name="ZH-CN_TOPIC_0000002479226832"></a>

**Table 1** PodGroup labels used by cluster scheduling components

<a name="table143562050699"></a>

|Name|Description|Value|Component|
|--|--|--|--|
|ring-controller.atlas|Identifies Atlas Pods|<ul><li>ascend-npu</li><li>ascend-910</li><li>ascend-<span><em>{xxx}</em></span>b</li></ul>|Ascend Device Plugin, Ascend Operator, Volcano|
|fault-scheduling|Rescheduling upon job faults|grace, force, off|Volcano, Resilience Controller|
|elastic-scheduling|Job elastic scheduling |on|Volcano, Resilience Controller|
|fault-retry-times|Number of times a job can be rescheduled when a service plane fault occurs|0-100|Volcano, Ascend Operator|
|tor-affinity|Switch affinity policy|<ul><li>normal-schema</li><li>large-model-schema</li><li>null</li></ul>|Volcano|
|npu-310-strategy|Marks the scheduling policy for inference servers (with Atlas 300I inference cards)|<ul><li>card</li><li>chip</li></ul>|Volcano|
|pod-rescheduling|Enable Pod-level rescheduling.|<ul><li>on: Enable Pod-level rescheduling</li><li>Other values or if this field is not used: Disable Pod-level rescheduling</li></ul>|Volcano|
|process-recover-enable|Enable process-level rescheduling.|<ul><li>on: Enable process-level rescheduling</li><li>Other values or if this field is not used: Disable process-level rescheduling</li></ul>|Volcano|
|subHealthyStrategy|Sub-health handling policy.|<ul><li>ignore: Ignore the sub-healthy node. Subsequent jobs will not prioritize this node in affinity scheduling.</li><li>graceExit: Do not use the sub-healthy node. Save the checkpoint file and then perform rescheduling. Subsequent jobs will not be scheduled to this node.</li><li>forceExit: Do not use the sub-healthy node. Exit the jobs without saving and perform rescheduling. Subsequent jobs will not be scheduled to this node.</li><li>hotSwitch: Perform a hot switching. After the backup Pod is started, pause the training jobs and use a new node to restart training.</li></ul>|Volcano|
|huawei.com/scheduler.softShareDev.aicoreQuota|Percentage of AICore requested by the soft partitioning job.|[1, 100]|Volcano|
|huawei.com/scheduler.softShareDev.hbmQuota|Amount of high-bandwidth memory requested by the soft partitioning job.|<p>[1, maxHBM]</p><p>maxHBM is the HBM value in HBM-Usage(MB) queried using the <b>npu-smi info</b> command.</p>|Volcano|
|huawei.com/scheduler.softShareDev.policy|Policy for the soft partitioning job.|<ul><li>fixed-share</li><li>elastic</li><li>best-effort</li></ul>|Volcano|

**Table 2** PodGroup annotations used by cluster scheduling components

<a name="table87117712413"></a>
<table><thead align="left"><tr id="row167127122419"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p17127152415"><a name="p17127152415"></a><a name="p17127152415"></a>Name</p>
</th>
<th class="cellrowborder" valign="top" width="24.169999999999998%" id="mcps1.2.5.1.2"><p id="p14713722416"><a name="p14713722416"></a><a name="p14713722416"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="27.450000000000003%" id="mcps1.2.5.1.3"><p id="p471127192414"><a name="p471127192414"></a><a name="p471127192414"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="23.380000000000003%" id="mcps1.2.5.1.4"><p id="p57187142419"><a name="p57187142419"></a><a name="p57187142419"></a>Component</p>
</th>
</tr>
</thead>
<tbody><tr id="row47177202416"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p576592911262"><a name="p576592911262"></a><a name="p576592911262"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p167655293262"><a name="p167655293262"></a><a name="p167655293262"></a>Specifies the number of chips in a logical SuperPoD.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p77651729202614"><a name="p77651729202614"></a><a name="p77651729202614"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p1072973244"><a name="p1072973244"></a><a name="p1072973244"></a><span id="ph197215716249"><a name="ph197215716249"></a><a name="ph197215716249"></a>Volcano</span>, <span id="ph17212711245"><a name="ph17212711245"></a><a name="ph17212711245"></a>Ascend Operator</span></p>
</td>
</tr>
<tr id="row972875243"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p20765429172612"><a name="p20765429172612"></a><a name="p20765429172612"></a>huawei.com/schedule_policy</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p1276572913269"><a name="p1276572913269"></a><a name="p1276572913269"></a>Specifies the scheduling policy.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p1389015013273"><a name="p1389015013273"></a><a name="p1389015013273"></a>Currently supports the configurations in <a href="#table1120511613153">Table 3</a>.</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p197214711249"><a name="p197214711249"></a><a name="p197214711249"></a><span id="ph972477246"><a name="ph972477246"></a><a name="ph972477246"></a>Volcano</span></p>
</td>
</tr>
<tr id="row572178247"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p13765229182617"><a name="p13765229182617"></a><a name="p13765229182617"></a>sp-fit</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p1276582913269"><a name="p1276582913269"></a><a name="p1276582913269"></a>SuperPoD scheduling policy.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><ul><li>idlest: Logical SuperPoDs are scheduled to more idle physical SuperPoDs.</li><li>Non-idlest: Logical SuperPoDs preferentially fill up physical SuperPoDs.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p77220742415"><a name="p77220742415"></a><a name="p77220742415"></a><span id="ph1372071243"><a name="ph1372071243"></a><a name="ph1372071243"></a>Volcano</span></p>
</td>
</tr>
<tr id="row1721472248"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p3766152922612"><a name="p3766152922612"></a><a name="p3766152922612"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p576614298267"><a name="p576614298267"></a><a name="p576614298267"></a>Minimum number of replicas required for a job to be scheduled.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p10766102912261"><a name="p10766102912261"></a><a name="p10766102912261"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p1972147182413"><a name="p1972147182413"></a><a name="p1972147182413"></a><span id="ph57212720245"><a name="ph57212720245"></a><a name="ph57212720245"></a>Volcano</span></p>
</td>
</tr>
<tr id="row6729792413"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p19766129192612"><a name="p19766129192612"></a><a name="p19766129192612"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p376652917262"><a name="p376652917262"></a><a name="p376652917262"></a>Job rescheduling policy.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p1178352611283"><a name="p1178352611283"></a><a name="p1178352611283"></a>pod: Only supports Pod-level rescheduling, and will not escalate to the Job level. (When using vcjob, you need to configure this policy: policies: -event:PodFailed -action:RestartTask)</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p13726762413"><a name="p13726762413"></a><a name="p13726762413"></a><span id="ph1672771246"><a name="ph1672771246"></a><a name="ph1672771246"></a>Volcano</span></p>
</td>
</tr>
<tr id="row2032944619369"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p6523192073819"><a name="p6523192073819"></a><a name="p6523192073819"></a>huawei.com/schedule_enable_dequeue</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p95238206385"><a name="p95238206385"></a><a name="p95238206385"></a>Enables or disables the job dequeue function (transitioning from Inqueue to Pending state). Manual configuration is required.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><a name="ul1452313209384"></a><a name="ul1452313209384"></a><ul id="ul1452313209384"><li>"on": Enable</li><li>Other values: Disable</li></ul>
<p id="p184512184913"><a name="p184512184913"></a><a name="p184512184913"></a>If not configured, it is disabled by default.</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p19523102023811"><a name="p19523102023811"></a><a name="p19523102023811"></a><span id="ph16444326193819"><a name="ph16444326193819"></a><a name="ph16444326193819"></a>Volcano</span></p>
</td>
</tr>
<tr id="row1450448133619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p195231920133819"><a name="p195231920133819"></a><a name="p195231920133819"></a>huawei.com/schedule_dequeue_frequency</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p4523172063820"><a name="p4523172063820"></a><a name="p4523172063820"></a>Records the number of times a job is dequeued. Automatically updated by <span id="ph5862824114017"><a name="ph5862824114017"></a><a name="ph5862824114017"></a>Volcano</span>.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p125231320193818"><a name="p125231320193818"></a><a name="p125231320193818"></a>This value increments by 1 each time the job is dequeued.</p>
<div class="note" id="note10987851174216"><a name="note10987851174216"></a><div class="notebody"><p id="p698710511425"><a name="p698710511425"></a><a name="p698710511425"></a>This value is deleted when the job is not in the Inqueue or Pending state.</p>
</div></div>
<p id="p105231520203811"><a name="p105231520203811"></a><a name="p105231520203811"></a></p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p12523520123817"><a name="p12523520123817"></a><a name="p12523520123817"></a><span id="ph497462713812"><a name="ph497462713812"></a><a name="ph497462713812"></a>Volcano</span></p>
</td>
</tr>
<tr id="row16233175083617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p5523142033819"><a name="p5523142033819"></a><a name="p5523142033819"></a>huawei.com/schedule_enqueue_time</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p152312015383"><a name="p152312015383"></a><a name="p152312015383"></a>Records the time when a job is enqueued (transitioning from Pending to Inqueue state). Automatically updated by <span id="ph19470113214427"><a name="ph19470113214427"></a><a name="ph19470113214427"></a>Volcano</span>.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p1152332013385"><a name="p1152332013385"></a><a name="p1152332013385"></a>Millisecond-level timestamp.</p>
<div class="note" id="note813021515431"><a name="note813021515431"></a><div class="notebody"><a name="ul1115755417436"></a><a name="ul1115755417436"></a><ul id="ul1115755417436"><li>If a job has been enqueued for more than 5 minutes and the dequeue function is enabled, when other jobs need to be enqueued, this job will be dequeued to release resources so that other jobs can be enqueued.</li><li>This value is deleted when the job is not in the Inqueue state.</li></ul>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p1752310207382"><a name="p1752310207382"></a><a name="p1752310207382"></a><span id="ph10540629193811"><a name="ph10540629193811"></a><a name="ph10540629193811"></a>Volcano</span></p>
</td>
</tr>
<tr id="row16233175083617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p5523142033819"><a name="p5523142033819"></a><a name="p5523142033819"></a>huawei.com/affinity-config</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p>Configures the affinity hierarchy for multi-level scheduling of jobs.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p>level1=x,level2=y,...</p><p>Where x, y... are the subjob sizes for the corresponding network hierarchy.</p><p>The format must be a concatenation of strings in the format leveli=ni, separated by commas. Here, i is the network hierarchy sequence number, and ni is the number of replicas for the subjob at that network hierarchy. For example, for a job with a total of 8 replicas, "level1=2,level2=4" means that every 2 Pods in the job are assigned to nodes with the same level1 label, and every 4 Pods are assigned to nodes with the same level2 label.</p><p>The network hierarchy configuration must meet the following requirements: <ul><li>When the job hierarchy has more than 1 level, the value of level n must be an integer multiple of n-1.</li><li>The total number of job replicas must be an integer multiple of all hierarchy levels.</li><li>The job hierarchy configuration must start from level1 and be consecutive in ascending order.</li></ul></p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p><span>Volcano</span></p>
</td>
</tr>
</tbody>
</table>

**Table 3** huawei.com/schedule_policy configurations

<a name="table1120511613153"></a>

| Configuration | Description |
|--|--|
| chip4-node8 | 1 node with 8 chips, every 4 chips form an interconnect ring. For example, Atlas 800 training server (model 9000)/Atlas 800 training server (model 9010)/Atlas 350 PCIe card has 8 chips total, every 4 chips are connected via UB mezzanine. |
| chip1-node2 | 1 node with 2 chips. For example, the Atlas 300T training card supports up to 1 chip, and 1 node supports up to 2 chips. |
| chip4-node4 | 1 node with 4 chips, forming 1 interconnect ring, for example, Atlas 800 training server (models 9000, 9010) half-configuration chip scenario. |
| chip8-node8 | 1 node with 8 chips. All 8 chips form 1 interconnect ring. For example, Atlas 800T A2 training server/Atlas 850 series hardware products. |
| chip8-node16 | 1 node with 16 chips, every 8 chips on 1 interconnect ring. For example, Atlas 200T A2 Box16 heterogeneous subrack. |
| chip2-node8 | 1 node with 8 chips, every 2 chips on 1 interconnect ring. |
| chip2-node16 | 1 node with 16 chips, every 2 chips on 1 interconnect ring. For example, Atlas 800T A3 SuperPoD. |
| chip2-node8-sp | 1 node with 8 chips, every 2 chips on 1 interconnect ring. Multiple servers form a SuperPoD, for example, Atlas 9000 A3 SuperPoD cluster computing system. |
| chip2-node16-sp | 1 node with 16 chips, every 2 chip on 1 interconnect ring. Multiple servers form a SuperPoD, for example, Atlas 900 A3 SuperPoD. |
| chip4-node16 | 1 node with 16 chips, every 4 chips on 1 interconnect ring. For example, Atlas 350 PCIe card with 16 chips total, every 4 chips connected via UB mezzanine. |
| chip1-node8 | 1 node with 8 chips, no interconnect between chips. For example, Atlas 350 PCIe card with 8 chips total, no interconnect between chips. |
| chip1-node16 | 1 node with 16 chips, no interconnect between chips. For example, Atlas 350 PCIe card with 16 chips total, no interconnect between chips. |
| chip8-node8-sp | 1 node with 8 chips, all 8 chips on 1 interconnect ring. Multiple servers form a SuperPoD, for example, Atlas 850 SuperPoD. |
| chip8-node8-ra64-sp | 1 node with 8 chips, all 8 chips on 1 interconnect ring, 64 nodes form a computing cabinet. Multiple cabinets form a SuperPoD, for example, Atlas 950 SuperPoD. |
| chip1-softShareDev | Dedicated scheduling policy for soft partition-based virtualization. |
| multilevel | Used in multi-level scheduling scenarios. For details on how to use multi-level scheduling, see [Multi-level Scheduling](../usage/basic_scheduling/05_multi_level_scheduling.md). |

## Pod<a name="ZH-CN_TOPIC_0000002484428552"></a>

**Table 1** Pod labels used by cluster scheduling components

<a name="table143562050699"></a>
<table><thead align="left"><tr id="row23564507918"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p1535615011914"><a name="p1535615011914"></a><a name="p1535615011914"></a>Name</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.2"><p id="p83576501093"><a name="p83576501093"></a><a name="p83576501093"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="p235719501097"><a name="p235719501097"></a><a name="p235719501097"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p435716507913"><a name="p435716507913"></a><a name="p435716507913"></a>Component</p>
</th>
</tr>
</thead>
<tbody><tr id="row7357125010917"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0357155019917"><a name="p0357155019917"></a><a name="p0357155019917"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p1035711508917"><a name="p1035711508917"></a><a name="p1035711508917"></a>Identifies the Atlas <span id="ph13571150291"><a name="ph13571150291"></a><a name="ph13571150291"></a>Pod</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul835765011916"></a><a name="ul835765011916"></a><ul id="ul835765011916"><li>ascend-910</li><li>ascend-<span id="ph19358150597"><a name="ph19358150597"></a><a name="ph19358150597"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li><li>huawei.com/npu</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10471193985417"><a name="p10471193985417"></a><a name="p10471193985417"></a><span id="ph1035865018915"><a name="ph1035865018915"></a><a name="ph1035865018915"></a>Ascend Device Plugin</span>, <span id="ph446593975417"><a name="ph446593975417"></a><a name="ph446593975417"></a>Ascend Operator</span>, <span id="ph635885012911"><a name="ph635885012911"></a><a name="ph635885012911"></a>Volcano</span></p>
</td>
</tr>
<tr id="row135825013910"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p203581150493"><a name="p203581150493"></a><a name="p203581150493"></a>fault-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p935815502914"><a name="p935815502914"></a><a name="p935815502914"></a>Switch for job rescheduling upon faults</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p5358950291"><a name="p5358950291"></a><a name="p5358950291"></a>grace, force, off</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10358105012913"><a name="p10358105012913"></a><a name="p10358105012913"></a><span id="ph635812501497"><a name="ph635812501497"></a><a name="ph635812501497"></a>Volcano</span>, <span id="ph183581350898"><a name="ph183581350898"></a><a name="ph183581350898"></a>Resilience Controller</span></p>
</td>
</tr>
<tr id="row03591501297"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p143599501394"><a name="p143599501394"></a><a name="p143599501394"></a>elastic-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p1235917508916"><a name="p1235917508916"></a><a name="p1235917508916"></a>Switch for job elastic scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p1135910503918"><a name="p1135910503918"></a><a name="p1135910503918"></a>on</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1735916501991"><a name="p1735916501991"></a><a name="p1735916501991"></a><span id="ph93613501992"><a name="ph93613501992"></a><a name="ph93613501992"></a>Resilience Controller</span>, <span id="ph53614501198"><a name="ph53614501198"></a><a name="ph53614501198"></a>Volcano</span></p>
</td>
</tr>
<tr id="row103614504912"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p6361950695"><a name="p6361950695"></a><a name="p6361950695"></a>fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p183613501791"><a name="p183613501791"></a><a name="p183613501791"></a>Number of times a job can be rescheduled when a service-plane fault occurs</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p1836115501693"><a name="p1836115501693"></a><a name="p1836115501693"></a>0-100</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p636212501391"><a name="p636212501391"></a><a name="p636212501391"></a><span id="ph1036213501896"><a name="ph1036213501896"></a><a name="ph1036213501896"></a>Volcano</span>, <span id="ph436285014913"><a name="ph436285014913"></a><a name="ph436285014913"></a>Ascend Operator</span></p>
</td>
</tr>
<tr id="row1336212502091"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p103625502094"><a name="p103625502094"></a><a name="p103625502094"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p43628508911"><a name="p43628508911"></a><a name="p43628508911"></a>Switch affinity policy</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul143629507913"></a><a name="ul143629507913"></a><ul id="ul143629507913"><li>normal-schema</li><li>large-model-schema</li><li>null</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p436310506914"><a name="p436310506914"></a><a name="p436310506914"></a><span id="ph73631050690"><a name="ph73631050690"></a><a name="ph73631050690"></a>Volcano</span></p>
</td>
</tr>
<tr id="row1136411501898"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0364135010913"><a name="p0364135010913"></a><a name="p0364135010913"></a>npu-310-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p636425010918"><a name="p636425010918"></a><a name="p636425010918"></a>Marks the scheduling policy for the inference server (with <span id="ph1436410501390"><a name="ph1436410501390"></a><a name="ph1436410501390"></a>Atlas 300I inference card</span>)</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul73644501797"></a><a name="ul73644501797"></a><ul id="ul73644501797"><li>card</li><li>chip</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1636419501398"><a name="p1636419501398"></a><a name="p1636419501398"></a><span id="ph163652501393"><a name="ph163652501393"></a><a name="ph163652501393"></a>Volcano</span></p>
</td>
</tr>
<tr id="row7970125593620"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p65914912716"><a name="p65914912716"></a><a name="p65914912716"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p3591694276"><a name="p3591694276"></a><a name="p3591694276"></a>Whether to enable Pod-level rescheduling.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul186101614131"></a><a name="ul186101614131"></a><ul id="ul186101614131"><li>on: Enable Pod-level rescheduling</li><li>Other values or when this field is not used: Disable Pod-level rescheduling</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1372045172812"><a name="p1372045172812"></a><a name="p1372045172812"></a><span id="ph2072005192818"><a name="ph2072005192818"></a><a name="ph2072005192818"></a>Volcano</span></p>
</td>
</tr>
<tr id="row209101813153710"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p2417162275410"><a name="p2417162275410"></a><a name="p2417162275410"></a>process-recover-enable</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p131172319273"><a name="p131172319273"></a><a name="p131172319273"></a>Whether to enable process-level rescheduling.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul71592205015"></a><a name="ul71592205015"></a><ul id="ul71592205015"><li>on: Enable process-level rescheduling</li><li>Other values or when this field is not used: Disable process-level rescheduling</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p181177342718"><a name="p181177342718"></a><a name="p181177342718"></a><span id="ph102814152910"><a name="ph102814152910"></a><a name="ph102814152910"></a>Volcano</span></p>
</td>
</tr>
<tr id="row8889122663714"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p172891224132816"><a name="p172891224132816"></a><a name="p172891224132816"></a>subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p728915247282"><a name="p728915247282"></a><a name="p728915247282"></a>Sub-health handling strategy</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul18716519102210"></a><a name="ul18716519102210"></a><ul id="ul18716519102210"><li>ignore: Ignore the sub-healthy node, and subsequent jobs will not prioritize this node in affinity scheduling.</li><li>graceExit: Do not use the sub-healthy node, save the dying gasp CKPT file, and then perform rescheduling. Subsequent jobs will not be scheduled to this node.</li><li>forceExit: Do not use the sub-healthy node, exit the job directly without saving, and perform rescheduling. Subsequent jobs will not be scheduled to this node.</li><li>hotSwitch: Perform a hot switching, start the backup Pod, pause the training job, and use a new node to restart the training.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p929902143119"><a name="p929902143119"></a><a name="p929902143119"></a><span id="ph6299326312"><a name="ph6299326312"></a><a name="ph6299326312"></a>Volcano</span></p>
</td>
</tr>
</tbody>
</table>

**Table 2** Pod annotations used by cluster scheduling components

<a name="table87117712413"></a>
<table><thead align="left"><tr id="row167127122419"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p17127152415"><a name="p17127152415"></a><a name="p17127152415"></a>Name</p>
</th>
<th class="cellrowborder" valign="top" width="24.169999999999998%" id="mcps1.2.5.1.2"><p id="p14713722416"><a name="p14713722416"></a><a name="p14713722416"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="27.450000000000003%" id="mcps1.2.5.1.3"><p id="p471127192414"><a name="p471127192414"></a><a name="p471127192414"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="23.380000000000003%" id="mcps1.2.5.1.4"><p id="p57187142419"><a name="p57187142419"></a><a name="p57187142419"></a>Component</p>
</th>
</tr>
</thead>
<tbody><tr id="row47177202416"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p576592911262"><a name="p576592911262"></a><a name="p576592911262"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p167655293262"><a name="p167655293262"></a><a name="p167655293262"></a>Specifies the number of chips in a logical SuperPoD.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p77651729202614"><a name="p77651729202614"></a><a name="p77651729202614"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p1072973244"><a name="p1072973244"></a><a name="p1072973244"></a><span id="ph197215716249"><a name="ph197215716249"></a><a name="ph197215716249"></a>Volcano</span>, <span id="ph17212711245"><a name="ph17212711245"></a><a name="ph17212711245"></a>Ascend Operator</span></p>
</td>
</tr>
<tr id="row972875243"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p20765429172612"><a name="p20765429172612"></a><a name="p20765429172612"></a>huawei.com/schedule_policy</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p1276572913269"><a name="p1276572913269"></a><a name="p1276572913269"></a>Specifies the scheduling policy.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p1389015013273"><a name="p1389015013273"></a><a name="p1389015013273"></a>Currently supports the configurations in <a href="#table1120511613153">Table 3</a>.</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p197214711249"><a name="p197214711249"></a><a name="p197214711249"></a><span id="ph972477246"><a name="ph972477246"></a><a name="ph972477246"></a>Volcano</span></p>
</td>
</tr>
<tr id="row572178247"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p13765229182617"><a name="p13765229182617"></a><a name="p13765229182617"></a>sp-fit</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p1276582913269"><a name="p1276582913269"></a><a name="p1276582913269"></a>SuperPoD scheduling policy.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><ul><li>idlest: The logical SuperPoD is scheduled to a more idle physical SuperPoD.</li><li>Non-idlest: The logical SuperPoD preferentially fills up a physical SuperPoD.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p77220742415"><a name="p77220742415"></a><a name="p77220742415"></a><span id="ph1372071243"><a name="ph1372071243"></a><a name="ph1372071243"></a>Volcano</span></p>
</td>
</tr>
<tr id="row1721472248"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p3766152922612"><a name="p3766152922612"></a><a name="p3766152922612"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p576614298267"><a name="p576614298267"></a><a name="p576614298267"></a>The minimum number of replicas required for the job to be scheduled.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p10766102912261"><a name="p10766102912261"></a><a name="p10766102912261"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p1972147182413"><a name="p1972147182413"></a><a name="p1972147182413"></a><span id="ph57212720245"><a name="ph57212720245"></a><a name="ph57212720245"></a>Volcano</span></p>
</td>
</tr>
<tr id="row6729792413"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p19766129192612"><a name="p19766129192612"></a><a name="p19766129192612"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p376652917262"><a name="p376652917262"></a><a name="p376652917262"></a>Job rescheduling policy.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p1178352611283"><a name="p1178352611283"></a><a name="p1178352611283"></a>pod: Only Pod-level rescheduling is supported, and it will not be escalated to the Job level. (When using vcjob, you need to configure this policy: policies: -event:PodFailed -action:RestartTask)</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p13726762413"><a name="p13726762413"></a><a name="p13726762413"></a><span id="ph1672771246"><a name="ph1672771246"></a><a name="ph1672771246"></a>Volcano</span></p>
</td>
</tr>
<tr id="row16233175083617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p5523142033819"><a name="p5523142033819"></a><a name="p5523142033819"></a>huawei.com/affinity-config</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p>Configures the affinity hierarchy for multi-level scheduling of the job.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p>level1=x,level2=y,...</p><p>Where x, y... are the subjob sizes at the corresponding network hierarchy.</p><p>The format must be a concatenation of strings in the format leveli=ni, separated by commas. Here, i is the network hierarchy sequence number, and ni is the number of replicas for the subjob at that network hierarchy. For example, for a job with a total of 8 replicas, "level1=2,level2=4" means that every 2 Pods in the job are assigned to nodes with the same level1 label, and every 4 Pods are assigned to nodes with the same level2 label.</p><p>The network hierarchy configuration must meet the following requirements: <ul><li>When the job hierarchy has more than 1 level, the value of level n must be an integer multiple of n-1.</li><li>The total number of job replicas must be an integer multiple of all hierarchy levels.</li><li>The job hierarchy configuration must start from level1 and be consecutive from small to large.</li></ul></p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p><span>Volcano</span></p>
</td>
</tr>
<tr id="row16233175083617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p5523142033819"><a name="p5523142033819"></a><a name="p5523142033819"></a>huawei.com/skip-ascend-plugin</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p>Allows some special jobs (such as jobs that do not require NPU resources) to bypass the default check logic of Ascend-for-volcano.</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p>Not set or set to "enabled"</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p><span>Volcano</span></p>
</td>
</tr>
</tbody>
</table>

**Table 3** PodGroup Status.Conditions used by cluster scheduling components

<a name="table_podgroup_conditions"></a>
<table><thead align="left"><tr id="row_podgroup_conditions_header"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p_podgroup_conditions_type"><a name="p_podgroup_conditions_type"></a><a name="p_podgroup_conditions_type"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p_podgroup_conditions_status"><a name="p_podgroup_conditions_status"></a><a name="p_podgroup_conditions_status"></a>Status</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.3"><p id="p_podgroup_conditions_reason"><a name="p_podgroup_conditions_reason"></a><a name="p_podgroup_conditions_reason"></a>Reason</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p_podgroup_conditions_message"><a name="p_podgroup_conditions_message"></a><a name="p_podgroup_conditions_message"></a>Message</p>
</th>
</tr>
</thead>
<tr id="row_podgroup_conditions_5"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p_podgroup_conditions_type_5"><a name="p_podgroup_conditions_type_5"></a><a name="p_podgroup_conditions_type_5"></a>Unschedulable</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p_podgroup_conditions_status_5"><a name="p_podgroup_conditions_status_5"></a><a name="p_podgroup_conditions_status_5"></a>True</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p_podgroup_conditions_reason_5"><a name="p_podgroup_conditions_reason_5"></a><a name="p_podgroup_conditions_reason_5"></a>JobEnqueueFailed</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p_podgroup_conditions_message_5"><a name="p_podgroup_conditions_message_5"></a><a name="p_podgroup_conditions_message_5"></a>The job failed to enqueue due to insufficient cluster NPU resources.</p>
</td>
</tr>
<tr id="row_podgroup_conditions_6"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p_podgroup_conditions_type_6"><a name="p_podgroup_conditions_type_6"></a><a name="p_podgroup_conditions_type_6"></a>Unschedulable</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p_podgroup_conditions_status_6"><a name="p_podgroup_conditions_status_6"></a><a name="p_podgroup_conditions_status_6"></a>True</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p_podgroup_conditions_reason_6"><a name="p_podgroup_conditions_reason_6"></a><a name="p_podgroup_conditions_reason_6"></a>JobValidateFailed</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p_podgroup_conditions_message_6"><a name="p_podgroup_conditions_message_6"></a><a name="p_podgroup_conditions_message_6"></a>job validation failed, typically due to an insufficient number of pods, non-compliant resource requests, or an unsupported job type.</p>
</td>
</tr>
<tr id="row_podgroup_conditions_7"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p_podgroup_conditions_type_7"><a name="p_podgroup_conditions_type_7"></a><a name="p_podgroup_conditions_type_7"></a>Unschedulable</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p_podgroup_conditions_status_7"><a name="p_podgroup_conditions_status_7"></a><a name="p_podgroup_conditions_status_7"></a>True</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p_podgroup_conditions_reason_7"><a name="p_podgroup_conditions_reason_7"></a><a name="p_podgroup_conditions_reason_7"></a>NodePredicateFailed</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p_podgroup_conditions_message_7"><a name="p_podgroup_conditions_message_7"></a><a name="p_podgroup_conditions_message_7"></a>Node filtering failed, containing detailed node filtering error information.</p>
</td>
</tr>
<tr id="row_podgroup_conditions_8"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p_podgroup_conditions_type_8"><a name="p_podgroup_conditions_type_8"></a><a name="p_podgroup_conditions_type_8"></a>Unschedulable</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p_podgroup_conditions_status_8"><a name="p_podgroup_conditions_status_8"></a><a name="p_podgroup_conditions_status_8"></a>True</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.3 "><p id="p_podgroup_conditions_reason_8"><a name="p_podgroup_conditions_reason_8"></a><a name="p_podgroup_conditions_reason_8"></a>BatchOrderFailed</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p_podgroup_conditions_message_8"><a name="p_podgroup_conditions_message_8"></a><a name="p_podgroup_conditions_message_8"></a>Batch node ordering failed, containing detailed ordering error information.</p>
</td>
</tr>
</table>

## Job Information<a name="ZH-CN_TOPIC_0000002479386798"></a>

**tor-share-cm<a name="section98191810400"></a>**

**Table 1**  tor-share-cm

<a name="table185653715301"></a>
<table><thead align="left"><tr id="row1857037203020"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p385703733012"><a name="p385703733012"></a><a name="p385703733012"></a>Name</p>
</th>
<th class="cellrowborder" valign="top" width="26.75%" id="mcps1.2.5.1.2"><p id="p28571437193018"><a name="p28571437193018"></a><a name="p28571437193018"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="23.25%" id="mcps1.2.5.1.3"><p id="p1385803713018"><a name="p1385803713018"></a><a name="p1385803713018"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p68588374304"><a name="p68588374304"></a><a name="p68588374304"></a>Remarks</p>
</th>
</tr>
</thead>
<tbody><tr id="row685863711309"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p138581437143012"><a name="p138581437143012"></a><a name="p138581437143012"></a>IsHealthy</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p1485818378307"><a name="p1485818378307"></a><a name="p1485818378307"></a>Switch status corresponding to the node</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p17859193717305"><a name="p17859193717305"></a><a name="p17859193717305"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1085918371305"><a name="p1085918371305"></a><a name="p1085918371305"></a>-</p>
</td>
</tr>
<tr id="row585918375300"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p585973723020"><a name="p585973723020"></a><a name="p585973723020"></a>IsSharedTor</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p1185913715307"><a name="p1185913715307"></a><a name="p1185913715307"></a>Switch attribute corresponding to the node</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p198596377302"><a name="p198596377302"></a><a name="p198596377302"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p8860637123014"><a name="p8860637123014"></a><a name="p8860637123014"></a>-</p>
</td>
</tr>
<tr id="row1286018378301"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p19860337183019"><a name="p19860337183019"></a><a name="p19860337183019"></a>NodeIP</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p18605378309"><a name="p18605378309"></a><a name="p18605378309"></a>Node IP</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p186043718309"><a name="p186043718309"></a><a name="p186043718309"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p58611837193020"><a name="p58611837193020"></a><a name="p58611837193020"></a>-</p>
</td>
</tr>
<tr id="row08615377305"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1186283713305"><a name="p1186283713305"></a><a name="p1186283713305"></a>NodeName</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p68629374305"><a name="p68629374305"></a><a name="p68629374305"></a>Node name</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1286383711307"><a name="p1286383711307"></a><a name="p1286383711307"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p158631337173013"><a name="p158631337173013"></a><a name="p158631337173013"></a>-</p>
</td>
</tr>
<tr id="row8863337133015"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p18863153743016"><a name="p18863153743016"></a><a name="p18863153743016"></a>JobName</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p19863143743015"><a name="p19863143743015"></a><a name="p19863143743015"></a>Job name</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p78631337113011"><a name="p78631337113011"></a><a name="p78631337113011"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p086473713017"><a name="p086473713017"></a><a name="p086473713017"></a>-</p>
</td>
</tr>
</tbody>
</table>

**vcjob-fault-npu-cm<a name="section1731892963620"></a>**

**Table 2** vcjob-fault-npu-cm field description

<a name="table153041817110"></a>
<table><thead align="left"><tr id="row4530818101120"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p1653191871120"><a name="p1653191871120"></a><a name="p1653191871120"></a><span id="ph135612450384"><a name="ph135612450384"></a><a name="ph135612450384"></a>Name</span></p>
</th>
<th class="cellrowborder" valign="top" width="26.75%" id="mcps1.2.5.1.2"><p id="p353111818113"><a name="p353111818113"></a><a name="p353111818113"></a><span id="ph4571459382"><a name="ph4571459382"></a><a name="ph4571459382"></a>Description</span></p>
</th>
<th class="cellrowborder" valign="top" width="23.25%" id="mcps1.2.5.1.3"><p id="p6531101821116"><a name="p6531101821116"></a><a name="p6531101821116"></a><span id="ph12579458385"><a name="ph12579458385"></a><a name="ph12579458385"></a>Value</span></p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p1753115188111"><a name="p1753115188111"></a><a name="p1753115188111"></a><span id="ph658045153811"><a name="ph658045153811"></a><a name="ph658045153811"></a>Remarks</span></p>
</th>
</tr>
</thead>
<tbody><tr id="row14547818131118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1454791821114"><a name="p1454791821114"></a><a name="p1454791821114"></a><span id="ph1158545163819"><a name="ph1158545163819"></a><a name="ph1158545163819"></a>fault-node</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p5547131815113"><a name="p5547131815113"></a><a name="p5547131815113"></a><span id="ph1359154516384"><a name="ph1359154516384"></a><a name="ph1359154516384"></a>Faulty node information</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1454719188113"><a name="p1454719188113"></a><a name="p1454719188113"></a><span id="ph11601045193815"><a name="ph11601045193815"></a><a name="ph11601045193815"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p154731817113"><a name="p154731817113"></a><a name="p154731817113"></a><span id="ph260204519383"><a name="ph260204519383"></a><a name="ph260204519383"></a>-</span></p>
</td>
</tr>
<tr id="row0547118101117"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p155476181111"><a name="p155476181111"></a><a name="p155476181111"></a><span id="ph186084513387"><a name="ph186084513387"></a><a name="ph186084513387"></a>- NodeName</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p254841814114"><a name="p254841814114"></a><a name="p254841814114"></a><span id="ph1161174511388"><a name="ph1161174511388"></a><a name="ph1161174511388"></a>Node Name</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p054815183116"><a name="p054815183116"></a><a name="p054815183116"></a><span id="ph15611245113815"><a name="ph15611245113815"></a><a name="ph15611245113815"></a>String</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p2054851813118"><a name="p2054851813118"></a><a name="p2054851813118"></a><span id="ph1621745143812"><a name="ph1621745143812"></a><a name="ph1621745143812"></a>-</span></p>
</td>
</tr>
<tr id="row55481518151111"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p954816185116"><a name="p954816185116"></a><a name="p954816185116"></a><span id="ph36311451382"><a name="ph36311451382"></a><a name="ph36311451382"></a>- UpdateTime</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p1554814184117"><a name="p1554814184117"></a><a name="p1554814184117"></a><span id="ph1964154513813"><a name="ph1964154513813"></a><a name="ph1964154513813"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p105487180118"><a name="p105487180118"></a><a name="p105487180118"></a><span id="ph3665457384"><a name="ph3665457384"></a><a name="ph3665457384"></a>64-bit integer</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p11548151841113"><a name="p11548151841113"></a><a name="p11548151841113"></a><span id="ph1682459383"><a name="ph1682459383"></a><a name="ph1682459383"></a>-</span></p>
</td>
</tr>
<tr id="row11549151819118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p75491618121111"><a name="p75491618121111"></a><a name="p75491618121111"></a><span id="ph2069144503818"><a name="ph2069144503818"></a><a name="ph2069144503818"></a>- UnhealthyNPU</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p17549518101114"><a name="p17549518101114"></a><a name="p17549518101114"></a><span id="ph57194583812"><a name="ph57194583812"></a><a name="ph57194583812"></a>Set of faulty chips on the faulty node</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p35499187114"><a name="p35499187114"></a><a name="p35499187114"></a><span id="ph8711945133815"><a name="ph8711945133815"></a><a name="ph8711945133815"></a>String slice</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p17549318151111"><a name="p17549318151111"></a><a name="p17549318151111"></a><span id="ph18721545163813"><a name="ph18721545163813"></a><a name="ph18721545163813"></a>-</span></p>
</td>
</tr>
<tr id="row95491186111"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p25501518181120"><a name="p25501518181120"></a><a name="p25501518181120"></a><span id="ph573164511386"><a name="ph573164511386"></a><a name="ph573164511386"></a>- NetworkUnhealthyNPU</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p2550161841120"><a name="p2550161841120"></a><a name="p2550161841120"></a><span id="ph873174553816"><a name="ph873174553816"></a><a name="ph873174553816"></a>Set of chips with network faults on the faulty node</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1955041814116"><a name="p1955041814116"></a><a name="p1955041814116"></a><span id="ph12741045163816"><a name="ph12741045163816"></a><a name="ph12741045163816"></a>String slice</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p65505186116"><a name="p65505186116"></a><a name="p65505186116"></a><span id="ph67494593817"><a name="ph67494593817"></a><a name="ph67494593817"></a>-</span></p>
</td>
</tr>
<tr id="row7551201831116"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p6551181831118"><a name="p6551181831118"></a><a name="p6551181831118"></a><span id="ph127544519384"><a name="ph127544519384"></a><a name="ph127544519384"></a>- NodeDEnable</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p17551318111111"><a name="p17551318111111"></a><a name="p17551318111111"></a><span id="ph776114513812"><a name="ph776114513812"></a><a name="ph776114513812"></a>Whether the node status detection switch is enabled</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><a name="ul55510181111"></a><a name="ul55510181111"></a><ul id="ul55510181111"><li><span id="ph078174514388"><a name="ph078174514388"></a><a name="ph078174514388"></a>True</span></li><li><span id="ph138054563812"><a name="ph138054563812"></a><a name="ph138054563812"></a>False</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p3552171818117"><a name="p3552171818117"></a><a name="p3552171818117"></a><span id="ph081545103815"><a name="ph081545103815"></a><a name="ph081545103815"></a>-</span></p>
</td>
</tr>
<tr id="row95521718111118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8552111817116"><a name="p8552111817116"></a><a name="p8552111817116"></a><span id="ph17811545153817"><a name="ph17811545153817"></a><a name="ph17811545153817"></a>- NodeHealthState</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p14552191817115"><a name="p14552191817115"></a><a name="p14552191817115"></a><span id="ph1082184563820"><a name="ph1082184563820"></a><a name="ph1082184563820"></a>Node health status</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p355221821111"><a name="p355221821111"></a><a name="p355221821111"></a><span id="ph882184593816"><a name="ph882184593816"></a><a name="ph882184593816"></a>String</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p255214185116"><a name="p255214185116"></a><a name="p255214185116"></a><span id="ph1883545163815"><a name="ph1883545163815"></a><a name="ph1883545163815"></a>-</span></p>
</td>
</tr>
<tr id="row356761891116"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p956814189118"><a name="p956814189118"></a><a name="p956814189118"></a><span id="ph1593145103817"><a name="ph1593145103817"></a><a name="ph1593145103817"></a>FaultDeviceList</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p35681218131114"><a name="p35681218131114"></a><a name="p35681218131114"></a><span id="ph1693134583816"><a name="ph1693134583816"></a><a name="ph1693134583816"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1756811861117"><a name="p1756811861117"></a><a name="p1756811861117"></a><span id="ph0931545163813"><a name="ph0931545163813"></a><a name="ph0931545163813"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p12568201801117"><a name="p12568201801117"></a><a name="p12568201801117"></a><span id="ph16941545183819"><a name="ph16941545183819"></a><a name="ph16941545183819"></a>-</span></p>
</td>
</tr>
<tr id="row1056811831111"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p15568151819118"><a name="p15568151819118"></a><a name="p15568151819118"></a><span id="ph199418457387"><a name="ph199418457387"></a><a name="ph199418457387"></a>- fault_type</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p4568118151117"><a name="p4568118151117"></a><a name="p4568118151117"></a><span id="ph195104520386"><a name="ph195104520386"></a><a name="ph195104520386"></a>Fault type object</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><a name="ul15568201841114"></a><a name="ul15568201841114"></a><ul id="ul15568201841114"><li><span id="ph179524583815"><a name="ph179524583815"></a><a name="ph179524583815"></a>CardUnhealthy: Chip fault</span></li><li><span id="ph596845183816"><a name="ph596845183816"></a><a name="ph596845183816"></a>CardNetworkUnhealthy: Chip network fault</span></li><li><span id="ph139684533810"><a name="ph139684533810"></a><a name="ph139684533810"></a>NodeUnhealthy: Node Failure</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p85692183111"><a name="p85692183111"></a><a name="p85692183111"></a><span id="ph1497114514381"><a name="ph1497114514381"></a><a name="ph1497114514381"></a>-</span></p>
</td>
</tr>
<tr id="row4569191813112"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p4569171841117"><a name="p4569171841117"></a><a name="p4569171841117"></a><span id="ph59720456384"><a name="ph59720456384"></a><a name="ph59720456384"></a>- npu_name</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p256931831110"><a name="p256931831110"></a><a name="p256931831110"></a><span id="ph189811454383"><a name="ph189811454383"></a><a name="ph189811454383"></a>Name of the faulty chip. Empty in case of a node failure.</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p65691318151118"><a name="p65691318151118"></a><a name="p65691318151118"></a><span id="ph169812450388"><a name="ph169812450388"></a><a name="ph169812450388"></a>String</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p356931818115"><a name="p356931818115"></a><a name="p356931818115"></a><span id="ph129994517380"><a name="ph129994517380"></a><a name="ph129994517380"></a>-</span></p>
</td>
</tr>
<tr id="row11570131817115"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1057018180118"><a name="p1057018180118"></a><a name="p1057018180118"></a><span id="ph1899445113813"><a name="ph1899445113813"></a><a name="ph1899445113813"></a>- fault_level</span></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p1257021816119"><a name="p1257021816119"></a><a name="p1257021816119"></a><span id="ph510084533811"><a name="ph510084533811"></a><a name="ph510084533811"></a>Fault handling type. Empty in case of a node failure.</span></p>
<p id="p7570111819115"><a name="p7570111819115"></a><a name="p7570111819115"></a></p>
<p id="p12570171881115"><a name="p12570171881115"></a><a name="p12570171881115"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><a name="ul1157001871115"></a><a name="ul1157001871115"></a><ul id="ul1157001871115"><li>NotHandleFault: No handling</li><li>RestartRequest: Re-execute inference requests for inference scenarios; re-execute training services for training scenarios</li><li>RestartBusiness: Re-execute the service</li><li>FreeRestartNPU: Affects service execution. Reset the chip when it becomes idle.</li><li>RestartNPU: Directly reset the chip and re-execute the service</li><li>SeparateNPU: Isolate the chip</li><li>PreSeparateNPU: Pre-isolate the chip. Determine whether to reschedule based on the actual running status of the training task.</li></ul>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.5.1.4 "><div class="note" id="note11570618121119"><a name="note11570618121119"></a><div class="notebody"><a name="ul17072011133917"></a><a name="ul17072011133917"></a><ul id="ul17072011133917"><li><span id="ph181001745123813"><a name="ph181001745123813"></a><a name="ph181001745123813"></a>The fault_level, fault_handling, and large_model_fault_level parameters have the same functionality. It is recommended to use fault_handling.</span></li><li>If an inference task subscribes to fault information and a RestartRequest fault occurs on the inference card used by the task, and the fault duration does not exceed 60 seconds, task rescheduling will not be performed. If the fault duration exceeds 60 seconds and is not recovered, the chip will be isolated and task rescheduling will be performed.</li></ul>
</div></div>
</td>
</tr>
<tr id="row195701318101112"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p19571111810111"><a name="p19571111810111"></a><a name="p19571111810111"></a><span id="ph141016453386"><a name="ph141016453386"></a><a name="ph141016453386"></a>- fault_handling</span></p>
</td>
</tr>
<tr id="row957116185118"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p6571111812118"><a name="p6571111812118"></a><a name="p6571111812118"></a><span id="ph2010174510386"><a name="ph2010174510386"></a><a name="ph2010174510386"></a>- large_model_fault_level</span></p>
</td>
</tr>
<tr id="row657171818113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p7571518141116"><a name="p7571518141116"></a><a name="p7571518141116"></a><span id="ph5103164513389"><a name="ph5103164513389"></a><a name="ph5103164513389"></a>- fault_code</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p857117181110"><a name="p857117181110"></a><a name="p857117181110"></a><span id="ph1410384543813"><a name="ph1410384543813"></a><a name="ph1410384543813"></a>Fault code. A string concatenated by commas.</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p057261817110"><a name="p057261817110"></a><a name="p057261817110"></a><span id="ph171041450384"><a name="ph171041450384"></a><a name="ph171041450384"></a>String</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><a name="ul057261815113"></a><a name="ul057261815113"></a><ul id="ul057261815113"><li><span id="ph41041545133816"><a name="ph41041545133816"></a><a name="ph41041545133816"></a>Disconnected: Chip network disconnection fault.</span></li><li><span id="ph31051045203810"><a name="ph31051045203810"></a><a name="ph31051045203810"></a>heartbeatTimeOut: Node status loss fault</span></li></ul>
</td>
</tr>
<tr id="row1757216185116"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1357251820116"><a name="p1357251820116"></a><a name="p1357251820116"></a><span id="ph181051445163811"><a name="ph181051445163811"></a><a name="ph181051445163811"></a>remain-retry-times</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p35725186119"><a name="p35725186119"></a><a name="p35725186119"></a><span id="ph17106945133816"><a name="ph17106945133816"></a><a name="ph17106945133816"></a>Remaining reschedulable information for the job</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p11573181820118"><a name="p11573181820118"></a><a name="p11573181820118"></a><span id="ph12106134593818"><a name="ph12106134593818"></a><a name="ph12106134593818"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p857381812119"><a name="p857381812119"></a><a name="p857381812119"></a><span id="ph181064454387"><a name="ph181064454387"></a><a name="ph181064454387"></a>-</span></p>
</td>
</tr>
<tr id="row1057312188112"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p35731618131111"><a name="p35731618131111"></a><a name="p35731618131111"></a><span id="ph1510784515383"><a name="ph1510784515383"></a><a name="ph1510784515383"></a>- UUID</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p11573141813118"><a name="p11573141813118"></a><a name="p11573141813118"></a><span id="ph8107104543812"><a name="ph8107104543812"></a><a name="ph8107104543812"></a>Task UID</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p3573141861113"><a name="p3573141861113"></a><a name="p3573141861113"></a><span id="ph1122184510383"><a name="ph1122184510383"></a><a name="ph1122184510383"></a>String</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p15731518101119"><a name="p15731518101119"></a><a name="p15731518101119"></a><span id="ph9123154533816"><a name="ph9123154533816"></a><a name="ph9123154533816"></a>-</span></p>
</td>
</tr>
<tr id="row1457316187116"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p12573151841115"><a name="p12573151841115"></a><a name="p12573151841115"></a><span id="ph512334517386"><a name="ph512334517386"></a><a name="ph512334517386"></a>- Times</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p185741118141119"><a name="p185741118141119"></a><a name="p185741118141119"></a><span id="ph17123134517384"><a name="ph17123134517384"></a><a name="ph17123134517384"></a>Remaining reschedulable times for the job</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p18574151841113"><a name="p18574151841113"></a><a name="p18574151841113"></a><span id="ph8124545193810"><a name="ph8124545193810"></a><a name="ph8124545193810"></a>Integer</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p457421819116"><a name="p457421819116"></a><a name="p457421819116"></a><span id="ph191247456380"><a name="ph191247456380"></a><a name="ph191247456380"></a>-</span></p>
</td>
</tr>
</tbody>
</table>

**reset-config-<job-name\><a name="section3394547123916"></a>**

The MindCluster cluster scheduling component writes information such as device and training task status into the reset-config-<job-name\> ConfigMap through K8s and maps it into the container. The Elastic Agent reads it and performs corresponding fault detection and processing.

**Table 3**  reset-config-_<job-name\>_

<a name="table1213115712136"></a>
<table><thead align="left"><tr id="row3132772132"><th class="cellrowborder" valign="top" width="13.940000000000003%" id="mcps1.2.6.1.1"><p id="p207081513112812"><a name="p207081513112812"></a><a name="p207081513112812"></a>Field Name</p>
</th>
<th class="cellrowborder" valign="top" width="16.700000000000003%" id="mcps1.2.6.1.2"><p id="p1313212741314"><a name="p1313212741314"></a><a name="p1313212741314"></a>Name</p>
</th>
<th class="cellrowborder" valign="top" width="22.28%" id="mcps1.2.6.1.3"><p id="p513317151314"><a name="p513317151314"></a><a name="p513317151314"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="31.220000000000002%" id="mcps1.2.6.1.4"><p id="p313315721314"><a name="p313315721314"></a><a name="p313315721314"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="15.860000000000001%" id="mcps1.2.6.1.5"><p id="p1313327191318"><a name="p1313327191318"></a><a name="p1313327191318"></a>Remarks</p>
</th>
</tr>
</thead>
<tbody><tr id="row41336711317"><td class="cellrowborder" rowspan="11" valign="top" width="13.940000000000003%" headers="mcps1.2.6.1.1 "><p id="p4472165312280"><a name="p4472165312280"></a><a name="p4472165312280"></a>reset.json</p>
</td>
<td class="cellrowborder" valign="top" width="16.700000000000003%" headers="mcps1.2.6.1.2 "><p id="p813420781315"><a name="p813420781315"></a><a name="p813420781315"></a>RankList</p>
</td>
<td class="cellrowborder" valign="top" width="22.28%" headers="mcps1.2.6.1.3 "><p id="p121346712134"><a name="p121346712134"></a><a name="p121346712134"></a>Chip list</p>
</td>
<td class="cellrowborder" valign="top" width="31.220000000000002%" headers="mcps1.2.6.1.4 "><p id="p5134137121315"><a name="p5134137121315"></a><a name="p5134137121315"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="15.860000000000001%" headers="mcps1.2.6.1.5 "><p id="p1513427131320"><a name="p1513427131320"></a><a name="p1513427131320"></a>-</p>
</td>
</tr>
<tr id="row21341174135"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p171346791316"><a name="p171346791316"></a><a name="p171346791316"></a>RankId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p3134177131313"><a name="p3134177131313"></a><a name="p3134177131313"></a>Rank information used by the faulty task</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1413587161310"><a name="p1413587161310"></a><a name="p1413587161310"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1413511721318"><a name="p1413511721318"></a><a name="p1413511721318"></a>-</p>
</td>
</tr>
<tr id="row1713512717138"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p161352712139"><a name="p161352712139"></a><a name="p161352712139"></a>LogicId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1135127181319"><a name="p1135127181319"></a><a name="p1135127181319"></a>Chip logic ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15135157131311"><a name="p15135157131311"></a><a name="p15135157131311"></a>32-bit integer</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p181366715137"><a name="p181366715137"></a><a name="p181366715137"></a>-</p>
</td>
</tr>
<tr id="row013914719136"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1313927191317"><a name="p1313927191317"></a><a name="p1313927191317"></a>Status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8139177171315"><a name="p8139177171315"></a><a name="p8139177171315"></a>Chip status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul8436530113"></a><a name="ul8436530113"></a><ul id="ul8436530113"><li>unrecovered: Not recovered</li><li>recovered: Recovery succeeded</li><li>failed: Recovery failed</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11394791316"><a name="p11394791316"></a><a name="p11394791316"></a>-</p>
</td>
</tr>
<tr id="row814016761315"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1814015719134"><a name="p1814015719134"></a><a name="p1814015719134"></a>Policy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p11140676132"><a name="p11140676132"></a><a name="p11140676132"></a>Hot reset policy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul1156918243817"></a><a name="ul1156918243817"></a><ul id="ul1156918243817"><li>empty: No fault</li><li>ignore: Ignore the fault</li><li>restart_request: Re-execute the current request</li><li>restart: Re-execute the training task</li><li>free_reset: The device needs to be restarted when there is no task on the NPU</li><li>reset: The device needs to be restarted</li><li>isolate: The device needs to be isolated</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11140672134"><a name="p11140672134"></a><a name="p11140672134"></a>-</p>
</td>
</tr>
<tr id="row151401717139"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p101413711136"><a name="p101413711136"></a><a name="p101413711136"></a>InitialPolicy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12141176132"><a name="p12141176132"></a><a name="p12141176132"></a>Initial hot reset policy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul1918372817111"></a><a name="ul1918372817111"></a><ul id="ul1918372817111"><li>empty: No fault</li><li>ignore: Ignore the fault</li><li>restart_request: Re-execute the current request</li><li>restart: Re-execute the training task</li><li>free_reset: The device needs to be restarted when there is no task on the NPU</li><li>reset: The device needs to be restarted</li><li>isolate: The device needs to be isolated</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p171419712133"><a name="p171419712133"></a><a name="p171419712133"></a>-</p>
</td>
</tr>
<tr id="row2141187121312"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3141197161312"><a name="p3141197161312"></a><a name="p3141197161312"></a>ErrorCode</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19141576138"><a name="p19141576138"></a><a name="p19141576138"></a>Decimal fault code</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p151429710139"><a name="p151429710139"></a><a name="p151429710139"></a>64-bit integer array</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1314257131311"><a name="p1314257131311"></a><a name="p1314257131311"></a>-</p>
</td>
</tr>
<tr id="row1721555113913"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p414625113526"><a name="p414625113526"></a><a name="p414625113526"></a>GracefulExit</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p321511543920"><a name="p321511543920"></a><a name="p321511543920"></a>Manages training processes</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p33363655012"><a name="p33363655012"></a><a name="p33363655012"></a>0 or 1</p>
<a name="ul7532185975011"></a><a name="ul7532185975011"></a><ul id="ul7532185975011"><li>Value 1: Terminate all training processes</li><li>Value 0: No action</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p921615511390"><a name="p921615511390"></a><a name="p921615511390"></a>-</p>
</td>
</tr>
<tr id="row45409251618"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1254115251666"><a name="p1254115251666"></a><a name="p1254115251666"></a>FaultFlushing</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p7541192512618"><a name="p7541192512618"></a><a name="p7541192512618"></a>Informs <span id="ph14256162281217"><a name="ph14256162281217"></a><a name="ph14256162281217"></a>Elastic Agent</span> whether a fault is currently being flushed</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p13813147101216"><a name="p13813147101216"></a><a name="p13813147101216"></a>Value is true or false</p>
<a name="ul1563191521213"></a><a name="ul1563191521213"></a><ul id="ul1563191521213"><li>true: Indicates that a fault is being flushed</li><li>false: Indicates that no fault is being flushed</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p19951631131314"><a name="p19951631131314"></a><a name="p19951631131314"></a><span id="ph952618296564"><a name="ph952618296564"></a><a name="ph952618296564"></a>Elastic Agent</span> needs to wait until this field is false and the fault RankList does not contain a fault for this node before starting the training process</p>
</td>
</tr>
<tr id="row141375594377"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p64521951162319"><a name="p64521951162319"></a><a name="p64521951162319"></a><span>RestartFaultProcess</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17453851172311"><a name="p17453851172311"></a><a name="p17453851172311"></a><span>Informs </span><span id="ph262783362516"><a name="ph262783362516"></a><a name="ph262783362516"></a>Elastic Agent</span><span> whether to restart only the faulty process on this node</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2012431813258"><a name="p2012431813258"></a><a name="p2012431813258"></a><span>Value is true or false</span></p>
<a name="ul14729113619013"></a><a name="ul14729113619013"></a><ul id="ul14729113619013"><li><span>true: When a fault occurs on this node, restart only the faulty process on this node</span></li><li><span>false: When a fault occurs on this node, exit all processes on this node and exit </span><span id="ph205211257104"><a name="ph205211257104"></a><a name="ph205211257104"></a>Elastic Agent</span></li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p94534513233"><a name="p94534513233"></a><a name="p94534513233"></a>This field takes effect only when the fault RankList contains a fault for this node</p>
</td>
</tr>
<tr id="row14142137191314"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p171421973132"><a name="p171421973132"></a><a name="p171421973132"></a>ErrorCodeHex</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p0142577133"><a name="p0142577133"></a><a name="p0142577133"></a>Hexadecimal fault code</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p8142177131320"><a name="p8142177131320"></a><a name="p8142177131320"></a>String</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p31421070133"><a name="p31421070133"></a><a name="p31421070133"></a>-</p>
</td>
</tr>
<tr id="row1209179172911"><td class="cellrowborder" valign="top" width="13.940000000000003%" headers="mcps1.2.6.1.1 "><p id="p844941513297"><a name="p844941513297"></a><a name="p844941513297"></a>restartType</p>
</td>
<td class="cellrowborder" valign="top" width="16.700000000000003%" headers="mcps1.2.6.1.2 "><p id="p220916992912"><a name="p220916992912"></a><a name="p220916992912"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.28%" headers="mcps1.2.6.1.3 "><p id="p1820909182911"><a name="p1820909182911"></a><a name="p1820909182911"></a>Type of reset.json update</p>
</td>
<td class="cellrowborder" valign="top" width="31.220000000000002%" headers="mcps1.2.6.1.4 "><p id="p15209596295"><a name="p15209596295"></a><a name="p15209596295"></a>podReschedule or hotReset</p>
</td>
<td class="cellrowborder" valign="top" width="15.860000000000001%" headers="mcps1.2.6.1.5 "><p id="p95471047133013"><a name="p95471047133013"></a><a name="p95471047133013"></a>The value is podReschedule for single Pod rescheduling, and hotReset for hot recovery scenarios</p>
</td>
</tr>
</tbody>
</table>

**mindx-dl/job-reschedule-reason<a name="section20866121155814"></a>**

This ConfigMap is used to record the historical information of job rescheduling. By default, it saves the ten most recent rescheduling records of a job. When the ConfigMap content exceeds 950 KB, the earliest record of each job will be deleted in sequence.

**Table 4**  Job field description

<a name="table589619361579"></a>
<table><thead align="left"><tr id="row15897183618711"><th class="cellrowborder" valign="top" width="7.5200000000000005%" id="mcps1.2.6.1.1"><p id="p1389711365712"><a name="p1389711365712"></a><a name="p1389711365712"></a>Field</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.2"><p id="p20897836479"><a name="p20897836479"></a><a name="p20897836479"></a>Name</p>
</th>
<th class="cellrowborder" valign="top" width="24.740000000000002%" id="mcps1.2.6.1.3"><p id="p15897113614715"><a name="p15897113614715"></a><a name="p15897113614715"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="21.5%" id="mcps1.2.6.1.4"><p id="p889712361771"><a name="p889712361771"></a><a name="p889712361771"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.5"><p id="p12897153610715"><a name="p12897153610715"></a><a name="p12897153610715"></a>Remarks</p>
</th>
</tr>
</thead>
<tbody><tr id="row289716363716"><td class="cellrowborder" rowspan="4" valign="top" width="7.5200000000000005%" headers="mcps1.2.6.1.1 "><p id="p6386733753"><a name="p6386733753"></a><a name="p6386733753"></a>Job namespace/Job name</p>
<p id="p489813612714"><a name="p489813612714"></a><a name="p489813612714"></a></p>
<p id="p148981361575"><a name="p148981361575"></a><a name="p148981361575"></a></p>
<p id="p389883620715"><a name="p389883620715"></a><a name="p389883620715"></a></p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.2 "><p id="p114390307520"><a name="p114390307520"></a><a name="p114390307520"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="24.740000000000002%" headers="mcps1.2.6.1.3 "><p id="p178971736978"><a name="p178971736978"></a><a name="p178971736978"></a>Marks the name of the job that performs rescheduling.</p>
</td>
<td class="cellrowborder" valign="top" width="21.5%" headers="mcps1.2.6.1.4 "><p id="p88981936971"><a name="p88981936971"></a><a name="p88981936971"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.5 "><p id="p13898436378"><a name="p13898436378"></a><a name="p13898436378"></a>-</p>
</td>
</tr>
<tr id="row19898636879"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1489810360719"><a name="p1489810360719"></a><a name="p1489810360719"></a>JobID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p98985368719"><a name="p98985368719"></a><a name="p98985368719"></a>Job ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p18898736674"><a name="p18898736674"></a><a name="p18898736674"></a>String</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p389803617718"><a name="p389803617718"></a><a name="p389803617718"></a>-</p>
</td>
</tr>
<tr id="row48980368719"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p188988361372"><a name="p188988361372"></a><a name="p188988361372"></a>TotalRescheduleTimes</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1489812361714"><a name="p1489812361714"></a><a name="p1489812361714"></a>The total number of rescheduling times recorded for this job in the current lifecycle of <span id="ph525518226126"><a name="ph525518226126"></a><a name="ph525518226126"></a>Volcano</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p289810362715"><a name="p289810362715"></a><a name="p289810362715"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p989820364719"><a name="p989820364719"></a><a name="p989820364719"></a>-</p>
</td>
</tr>
<tr id="row19898163611716"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p109891552520"><a name="p109891552520"></a><a name="p109891552520"></a>RescheduleRecords</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p138991936178"><a name="p138991936178"></a><a name="p138991936178"></a>Records the specific information of the rescheduling of this task.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p32675459545"><a name="p32675459545"></a><a name="p32675459545"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p178998369716"><a name="p178998369716"></a><a name="p178998369716"></a>-</p>
</td>
</tr>
</tbody>
</table>

**Table 5**  RescheduleRecords description

<a name="table1578964348"></a>
<table><thead align="left"><tr id="row4327416646"><th class="cellrowborder" valign="top" width="7.5200000000000005%" id="mcps1.2.6.1.1"><p id="p112821723049"><a name="p112821723049"></a><a name="p112821723049"></a>Field Name</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.2"><p id="p16282823942"><a name="p16282823942"></a><a name="p16282823942"></a>Name</p>
</th>
<th class="cellrowborder" valign="top" width="24.740000000000002%" id="mcps1.2.6.1.3"><p id="p4282723240"><a name="p4282723240"></a><a name="p4282723240"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="21.5%" id="mcps1.2.6.1.4"><p id="p328212231943"><a name="p328212231943"></a><a name="p328212231943"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.5"><p id="p102821623248"><a name="p102821623248"></a><a name="p102821623248"></a>Remarks</p>
</th>
</tr>
</thead>
<tbody><tr id="row67891741047"><td class="cellrowborder" rowspan="3" valign="top" width="7.5200000000000005%" headers="mcps1.2.6.1.1 "><p id="p28117451834"><a name="p28117451834"></a><a name="p28117451834"></a>RescheduleRecords</p>
<p id="p101914156594"><a name="p101914156594"></a><a name="p101914156594"></a></p>
<p id="p161613154594"><a name="p161613154594"></a><a name="p161613154594"></a></p>
<p id="p182501417586"><a name="p182501417586"></a><a name="p182501417586"></a></p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.2 "><p id="p212744105514"><a name="p212744105514"></a><a name="p212744105514"></a>LogFileFormatTime</p>
</td>
<td class="cellrowborder" valign="top" width="24.740000000000002%" headers="mcps1.2.6.1.3 "><p id="p37891541546"><a name="p37891541546"></a><a name="p37891541546"></a>Rescheduling time recorded in <span id="ph769722034511"><a name="ph769722034511"></a><a name="ph769722034511"></a>Volcano</span> log format</p>
</td>
<td class="cellrowborder" valign="top" width="21.5%" headers="mcps1.2.6.1.4 "><p id="p13790104440"><a name="p13790104440"></a><a name="p13790104440"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.5 "><p id="p67904410410"><a name="p67904410410"></a><a name="p67904410410"></a>-</p>
</td>
</tr>
<tr id="row2790043411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1540831717595"><a name="p1540831717595"></a><a name="p1540831717595"></a>RescheduleTimeStamp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p117908411411"><a name="p117908411411"></a><a name="p117908411411"></a>Timestamp when rescheduling occurred</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p379064641"><a name="p379064641"></a><a name="p379064641"></a>String</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p3790144148"><a name="p3790144148"></a><a name="p3790144148"></a>-</p>
</td>
</tr>
<tr id="row8790941340"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p17859751501"><a name="p17859751501"></a><a name="p17859751501"></a>ReasonOfTask</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p198239143588"><a name="p198239143588"></a><a name="p198239143588"></a>Records the specific information of this rescheduling.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p666117055520"><a name="p666117055520"></a><a name="p666117055520"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p12817714165811"><a name="p12817714165811"></a><a name="p12817714165811"></a>-</p>
</td>
</tr>
</tbody>
</table>

**Table 6**  ReasonOfTask Description

<a name="table8680019155817"></a>
<table><thead align="left"><tr id="row165113075818"><th class="cellrowborder" valign="top" width="7.5200000000000005%" id="mcps1.2.6.1.1"><p id="p25183025820"><a name="p25183025820"></a><a name="p25183025820"></a>Field Name</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.2"><p id="p0514308584"><a name="p0514308584"></a><a name="p0514308584"></a>Name</p>
</th>
<th class="cellrowborder" valign="top" width="24.740000000000002%" id="mcps1.2.6.1.3"><p id="p175193045819"><a name="p175193045819"></a><a name="p175193045819"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="21.5%" id="mcps1.2.6.1.4"><p id="p1051103025812"><a name="p1051103025812"></a><a name="p1051103025812"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.5"><p id="p10511530195815"><a name="p10511530195815"></a><a name="p10511530195815"></a>Remarks</p>
</th>
</tr>
</thead>
<tbody><tr id="row1568121995819"><td class="cellrowborder" rowspan="4" valign="top" width="7.5200000000000005%" headers="mcps1.2.6.1.1 "><p id="p1868191919589"><a name="p1868191919589"></a><a name="p1868191919589"></a>ReasonOfTask</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.2 "><p id="p17681171915582"><a name="p17681171915582"></a><a name="p17681171915582"></a>RescheduleReason</p>
</td>
<td class="cellrowborder" valign="top" width="24.740000000000002%" headers="mcps1.2.6.1.3 "><p id="p1968114193584"><a name="p1968114193584"></a><a name="p1968114193584"></a>Reason for rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="21.5%" headers="mcps1.2.6.1.4 "><p id="p58011850228"><a name="p58011850228"></a><a name="p58011850228"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.5 "><p id="p1968191912589"><a name="p1968191912589"></a><a name="p1968191912589"></a>-</p>
</td>
</tr>
<tr id="row9681171915810"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1681191915581"><a name="p1681191915581"></a><a name="p1681191915581"></a>PodName</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p668219192585"><a name="p668219192585"></a><a name="p668219192585"></a>The pod that first triggered this rescheduling</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p186822019105812"><a name="p186822019105812"></a><a name="p186822019105812"></a>String</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11682131910582"><a name="p11682131910582"></a><a name="p11682131910582"></a>-</p>
</td>
</tr>
<tr id="row76821198589"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p17682171915812"><a name="p17682171915812"></a><a name="p17682171915812"></a>NodeName</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p26821819125817"><a name="p26821819125817"></a><a name="p26821819125817"></a>Node name</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p19682201917586"><a name="p19682201917586"></a><a name="p19682201917586"></a>String</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p186825190582"><a name="p186825190582"></a><a name="p186825190582"></a>The node that first triggered this rescheduling.</p>
</td>
</tr>
<tr id="row10682151985811"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p8682919115816"><a name="p8682919115816"></a><a name="p8682919115816"></a>NodeRankIndex</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19682191985813"><a name="p19682191985813"></a><a name="p19682191985813"></a>The rank of the node that first triggered this rescheduling in the training</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11682191995818"><a name="p11682191995818"></a><a name="p11682191995818"></a>String</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1269842159"><a name="p1269842159"></a><a name="p1269842159"></a>-</p>
</td>
</tr>
</tbody>
</table>

## Parameter Plane Network Topology<a name="ZH-CN_TOPIC_0000002479386820"></a>

**basic-tor-node-cm<a name="section18148132883914"></a>**

**Table 1**  basic-tor-node-cm

<a name="table18901255141213"></a>
<table><thead align="left"><tr id="row11911955131210"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p4911455191214"><a name="p4911455191214"></a><a name="p4911455191214"></a>Name</p>
</th>
<th class="cellrowborder" valign="top" width="26.75%" id="mcps1.2.5.1.2"><p id="p119105512124"><a name="p119105512124"></a><a name="p119105512124"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="23.25%" id="mcps1.2.5.1.3"><p id="p291205519126"><a name="p291205519126"></a><a name="p291205519126"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p4911955171219"><a name="p4911955171219"></a><a name="p4911955171219"></a>Remarks</p>
</th>
</tr>
</thead>
<tbody><tr id="row392125551217"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p892655181213"><a name="p892655181213"></a><a name="p892655181213"></a>version</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p8921055191217"><a name="p8921055191217"></a><a name="p8921055191217"></a>Version of basic-tor-node-cm</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1392455101212"><a name="p1392455101212"></a><a name="p1392455101212"></a>1.0</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1392125581216"><a name="p1392125581216"></a><a name="p1392125581216"></a>-</p>
</td>
</tr>
<tr id="row1921555151218"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p292355131211"><a name="p292355131211"></a><a name="p292355131211"></a>tor_count</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p2931155131219"><a name="p2931155131219"></a><a name="p2931155131219"></a>Number of nodes under the switch in the cluster</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p119395551211"><a name="p119395551211"></a><a name="p119395551211"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p3931455181211"><a name="p3931455181211"></a><a name="p3931455181211"></a>-</p>
</td>
</tr>
<tr id="row293165511123"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1693115512128"><a name="p1693115512128"></a><a name="p1693115512128"></a>server_list</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p893855191217"><a name="p893855191217"></a><a name="p893855191217"></a>Collection of cluster nodes grouped by switch</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p17931955191211"><a name="p17931955191211"></a><a name="p17931955191211"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p169417559127"><a name="p169417559127"></a><a name="p169417559127"></a>-</p>
</td>
</tr>
<tr id="row139425511215"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8941755121217"><a name="p8941755121217"></a><a name="p8941755121217"></a>- tor_id</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p7943559126"><a name="p7943559126"></a><a name="p7943559126"></a>Sequence number of the switch</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p109412553124"><a name="p109412553124"></a><a name="p109412553124"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p4940552122"><a name="p4940552122"></a><a name="p4940552122"></a>-</p>
</td>
</tr>
<tr id="row1194165581219"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1095195511125"><a name="p1095195511125"></a><a name="p1095195511125"></a>- tor_ip</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p19953555121"><a name="p19953555121"></a><a name="p19953555121"></a>IP address of the switch</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p99515520126"><a name="p99515520126"></a><a name="p99515520126"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p129535541220"><a name="p129535541220"></a><a name="p129535541220"></a>-</p>
</td>
</tr>
<tr id="row189575510122"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p18955559121"><a name="p18955559121"></a><a name="p18955559121"></a>server</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p595255171210"><a name="p595255171210"></a><a name="p595255171210"></a>Node information under the switch</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1796185511219"><a name="p1796185511219"></a><a name="p1796185511219"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p15971155131217"><a name="p15971155131217"></a><a name="p15971155131217"></a>-</p>
</td>
</tr>
<tr id="row199718557121"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1697555161219"><a name="p1697555161219"></a><a name="p1697555161219"></a>- server_ip</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p0986557122"><a name="p0986557122"></a><a name="p0986557122"></a>IP address of the node</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p16983555129"><a name="p16983555129"></a><a name="p16983555129"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10981755141218"><a name="p10981755141218"></a><a name="p10981755141218"></a>-</p>
</td>
</tr>
<tr id="row89805591220"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p109815561219"><a name="p109815561219"></a><a name="p109815561219"></a>- npu_count</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p799355141210"><a name="p799355141210"></a><a name="p799355141210"></a>Number of NPU chips on the node</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1399455161212"><a name="p1399455161212"></a><a name="p1399455161212"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p699145517127"><a name="p699145517127"></a><a name="p699145517127"></a>-</p>
</td>
</tr>
<tr id="row12993553125"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p9100155518126"><a name="p9100155518126"></a><a name="p9100155518126"></a>- slice_id</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p151001155141214"><a name="p151001155141214"></a><a name="p151001155141214"></a>Sequence number of the node under the switch</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p19100655111211"><a name="p19100655111211"></a><a name="p19100655111211"></a>Integer</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p131011555171218"><a name="p131011555171218"></a><a name="p131011555171218"></a>-</p>
</td>
</tr>
</tbody>
</table>

## Volcano Configuration<a name="ZH-CN_TOPIC_0000002511346767"></a>

**volcano-scheduler-configmap<a name="section42181344193715"></a>**

**Table 1**  Field description of volcano-scheduler-configmap

<a name="table1864354211112"></a>
<table><thead align="left"><tr id="row464317427117"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p10644942131117"><a name="p10644942131117"></a><a name="p10644942131117"></a>Name</p>
</th>
<th class="cellrowborder" valign="top" width="26.75%" id="mcps1.2.5.1.2"><p id="p96441642191113"><a name="p96441642191113"></a><a name="p96441642191113"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="23.25%" id="mcps1.2.5.1.3"><p id="p14644124281118"><a name="p14644124281118"></a><a name="p14644124281118"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p196441342201112"><a name="p196441342201112"></a><a name="p196441342201112"></a>Remarks</p>
</th>
</tr>
</thead>
<tbody><tr id="row964444217112"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1664520421116"><a name="p1664520421116"></a><a name="p1664520421116"></a>actions</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p764518425111"><a name="p764518425111"></a><a name="p764518425111"></a>Actions used in the scheduling process</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p8645942191119"><a name="p8645942191119"></a><a name="p8645942191119"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p2645154210118"><a name="p2645154210118"></a><a name="p2645154210118"></a>ascend-volcano-plugin uses three scheduling actions: enqueue, allocate, and backfill</p>
</td>
</tr>
<tr id="row5645184251114"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p14646114221110"><a name="p14646114221110"></a><a name="p14646114221110"></a>plugins</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p064613424117"><a name="p064613424117"></a><a name="p064613424117"></a>Set of plugins used in the scheduling process</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p16646194251117"><a name="p16646194251117"></a><a name="p16646194251117"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p664610425116"><a name="p664610425116"></a><a name="p664610425116"></a>-</p>
</td>
</tr>
<tr id="row16460421119"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p15647174213117"><a name="p15647174213117"></a><a name="p15647174213117"></a>- name</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p17647174251112"><a name="p17647174251112"></a><a name="p17647174251112"></a>Name of the plugin in use</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p164784221110"><a name="p164784221110"></a><a name="p164784221110"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1364764219113"><a name="p1364764219113"></a><a name="p1364764219113"></a>ascend-volcano-plugin uses</p>
<p id="p6647144211119"><a name="p6647144211119"></a><a name="p6647144211119"></a>the following scheduling plugins: priority, gang, conformance, volcano-npu_<em id="i14647134215110"><a name="i14647134215110"></a><a name="i14647134215110"></a>{version}</em>_linux-<em id="i9647144211116"><a name="i9647144211116"></a><a name="i9647144211116"></a>{arch}</em>, drf, predicates, proportion, nodeorder, and binpack</p>
</td>
</tr>
<tr id="row11647174219115"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p96481942141116"><a name="p96481942141116"></a><a name="p96481942141116"></a>configurations</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p964816428110"><a name="p964816428110"></a><a name="p964816428110"></a>Configuration information for scheduler initialization</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p10648042161117"><a name="p10648042161117"></a><a name="p10648042161117"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1364812426115"><a name="p1364812426115"></a><a name="p1364812426115"></a>-</p>
</td>
</tr>
<tr id="row16648134251119"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1664816426116"><a name="p1664816426116"></a><a name="p1664816426116"></a>- name</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p14648144261120"><a name="p14648144261120"></a><a name="p14648144261120"></a>Configuration information name</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p18649542171116"><a name="p18649542171116"></a><a name="p18649542171116"></a>init-params</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1264914219119"><a name="p1264914219119"></a><a name="p1264914219119"></a>-</p>
</td>
</tr>
<tr id="row136491042141119"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p164914420116"><a name="p164914420116"></a><a name="p164914420116"></a>- arguments</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p17649174210112"><a name="p17649174210112"></a><a name="p17649174210112"></a>Configuration information content</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1564964215115"><a name="p1564964215115"></a><a name="p1564964215115"></a>Key-value pair set</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p96491242191116"><a name="p96491242191116"></a><a name="p96491242191116"></a>-</p>
</td>
</tr>
</tbody>
</table>
