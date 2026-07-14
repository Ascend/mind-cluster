# Environment Variable Description<a name="ZH-CN_TOPIC_0000002479226386"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-26T11:48:22.724Z pushedAt=2026-06-27T00:32:25.613Z -->

## Environment Variables Used by MindCluster Components<a name="section1562121818463"></a>

The environment variables used by MindCluster components are described in [Table 1](#table1132513610543).

**Table 1** Environment variables

<a name="table1132513610543"></a>

|Environment Variable Name|Source|Required|Value|Description|
|--|--|--|--|--|
|POD_IP|Written in the YAML of the deployment component|Yes|IP of the Pod where the current container resides|Used by ClusterD and TaskD to start the gRPC service|
|POD_UID|Written in the YAML of the deployment component|No|UID of the Pod where the current container resides|Used to parse the server_id field of the RankTable file|
|ASCEND_DOCKER_RUNTIME|Written by Ascend Docker Runtime during container creation|No|"true"|Used by Ascend Device Plugin to determine whether the default container runtime on the current node is Ascend Docker Runtime|
|HOSTNAME|Written by K8s during container creation|Yes|Name of the pod where the current container resides|Used by Ascend Device Plugin to obtain the current pod name|
|NODE_NAME|Written in the YAML of the deployment component|Yes|Name of the node where the current container resides|Used by Ascend Device Plugin, NodeD, and ClusterD to obtain the current node name|
|LD_LIBRARY_PATH|Written in the Dockerfile|Yes|File path|Used by Ascend Device Plugin and NPU Exporter to initialize DCMI|
|BATCH_BIND_NUM|-|No|Numeric string|Specifies the number of pods for Volcano to bind in batch|
|MULTI_SCHEDULER_ENABLE|-|No|"true" or "false"|Specifies whether Volcano is used in a multi-scheduler scenario|
|SCHEDULER_POD_NAME|-|No|String|Specifies the Volcano scheduler pod name|
|SCHEDULER_NUM|-|No|Numeric string|Specifies the number of Volcano schedulers|
|PANIC_ON_ERROR|-|No|"true" or "false"|Specifies whether the Volcano scheduler needs to panic when an error occurs|
|KUBECONFIG|-|No|File path|Specifies the kubeconfig path for Volcano to connect to the K8s api-server|
|HOME|Written by K8s during container creation|Yes|Folder path|Specifies the current user home path obtained by Volcano|
|DEBUG_SOCKET_DIR|-|No|Socket file path|Specifies the socket path that Volcano listens on|
|HCCL_CONNECT_TIMEOUT|Written in the training script|No|HCCL link establishment timeout|Indicates the link establishment timeout|
|TTP_PORT|Written in the YAML of the deployment component|Yes|Communication port used by MindIO TTP|Used to start MindIO Controller|
|SSH_CLIENT|Environment variable set by the SSH server, containing information about the client connection|Yes|Information about the current client connection|Records this information in the operation log when installing Ascend Docker Runtime|
|TASKD_LOG_PATH|-|No|String|Indicates the disk path for TaskD running logs|
|MINDX_SERVER_IP|Written by Ascend Operator during container creation|Yes|String|Indicates the IP address for communication between a job and ClusterD, which is also the svc IP of clusterd-grpc-svc|
|MINDX_SERVER_DOMAIN|Written by Ascend Operator during container creation|Yes|String|Indicates the domain name for communication between a job and ClusterD. Default Value is "clusterd-grpc-svc.mindx-dl.svc.cluster.local"|
|MINDX_TASK_ID|Written by Ascend Operator during container creation|No|For MindIE inference jobs, the value is the value of the jobID field under the label field in an acjob |Required for Elastic Agent/TaskD to register the gRPC service with ClusterD and for the TaskD profiling feature to save logs|
|GROUP_BASE_DIR|Written in the job startup script|No|Folder path|Indicates the parallel domain information export path of TaskD |
|MINDIO_WAIT_MINDX_TIME|Written in the job YAML|No|Numeric string, with a value range of [1, 3600]|Timeout for waiting for faulty pod scheduling when process-level rescheduling is not enabled and elastic training is enabled|
|RAS_NET_ROOT_PATH|User configuration|No|Root path of the shared directory between ClusterD and NodeD|In the slow network diagnosis scenario, ClusterD and NodeD interact through shared storage. For details, see [Slow Network Diagnosis](../usage/resumable_training/01_solutions_principles.md)|
|REPLICA_TYPE|Written by Ascend Operator during container creation|Yes|Master, Scheduler, Chief, or Worker|Pod replica type|

## Ascend Operator Environment Variables<a name="section1272862810184"></a>

Ascend Operator provides corresponding environment variables for distributed training jobs (acjob) of different AI frameworks. For details, see the table below

**Table 2** Training environment variables injected by Ascend Operator

<a name="table154271816163912"></a>
<table><thead align="left"><tr id="row2428151693919"><th class="cellrowborder" valign="top" width="12.379999999999999%" id="mcps1.2.6.1.1"><p id="p13428016113914"><a name="p13428016113914"></a><a name="p13428016113914"></a>Framework</p>
</th>
<th class="cellrowborder" valign="top" width="16.869999999999997%" id="mcps1.2.6.1.2"><p id="p194281416103914"><a name="p194281416103914"></a><a name="p194281416103914"></a>Environment Variable</p>
</th>
<th class="cellrowborder" valign="top" width="27.77%" id="mcps1.2.6.1.3"><p id="p1342841653915"><a name="p1342841653915"></a><a name="p1342841653915"></a>Function</p>
</th>
<th class="cellrowborder" valign="top" width="19.79%" id="mcps1.2.6.1.4"><p id="p18871191318405"><a name="p18871191318405"></a><a name="p18871191318405"></a>Value</p>
</th>
<th class="cellrowborder" valign="top" width="23.189999999999998%" id="mcps1.2.6.1.5"><p id="p64281016193910"><a name="p64281016193910"></a><a name="p64281016193910"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row7428171663918"><td class="cellrowborder" rowspan="6" valign="top" width="12.379999999999999%" headers="mcps1.2.6.1.1 "><p id="p542811620396"><a name="p542811620396"></a><a name="p542811620396"></a><span id="ph19355165113512"><a name="ph19355165113512"></a><a name="ph19355165113512"></a>PyTorch</span></p>
<p id="p7428416183920"><a name="p7428416183920"></a><a name="p7428416183920"></a></p>
<p id="p134282016123915"><a name="p134282016123915"></a><a name="p134282016123915"></a></p>
<p id="p154281016143919"><a name="p154281016143919"></a><a name="p154281016143919"></a></p>
<p id="p756674313435"><a name="p756674313435"></a><a name="p756674313435"></a></p>
<p id="p788164613431"><a name="p788164613431"></a><a name="p788164613431"></a></p>
<p id="p397553410353"><a name="p397553410353"></a><a name="p397553410353"></a></p>
</td>
<td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p16428101633914"><a name="p16428101633914"></a><a name="p16428101633914"></a>MASTER_ADDR</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p4428181673917"><a name="p4428181673917"></a><a name="p4428181673917"></a>IP address for communicating with the master node</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><p id="p5871413104013"><a name="p5871413104013"></a><a name="p5871413104013"></a>Valid IP address in the format of string; must be in standard IPv4 or IPv6 format</p>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><a name="ul695319973016"></a><a name="ul695319973016"></a><p id="ul695319973016">clusterIP of the service corresponding to the master pod.</p>
</td>
</tr>
<tr id="row84281516153918"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p17428016193912"><a name="p17428016193912"></a><a name="p17428016193912"></a>MASTER_PORT</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p642871613915"><a name="p642871613915"></a><a name="p642871613915"></a>Port for communicating with the master node</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p196391835162819"><a name="p196391835162819"></a><a name="p196391835162819"></a>String or number, with a value range of 0 to 65520</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p182011320145513"><a name="p182011320145513"></a><a name="p182011320145513"></a>The value of the ascendjob-port field in the svc corresponding to the master pod. The default value is 2222.</p>
</td>
</tr>
<tr id="row1542861610390"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p13428161673916"><a name="p13428161673916"></a><a name="p13428161673916"></a>WORLD_SIZE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p124284165399"><a name="p124284165399"></a><a name="p124284165399"></a>Total number of NPUs used by the job</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1481964632819"><a name="p1481964632819"></a><a name="p1481964632819"></a>Integer greater than 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p442812163396"><a name="p442812163396"></a><a name="p442812163396"></a>Total number of NPUs used by the job. For example, if there are 64 NPUs, the value is 64.</p>
</td>
</tr>
<tr id="row1428216163912"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p204282016153919"><a name="p204282016153919"></a><a name="p204282016153919"></a>RANK</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p210753716532"><a name="p210753716532"></a><a name="p210753716532"></a>Node rank of the pod on this node</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p0871121315406"><a name="p0871121315406"></a><a name="p0871121315406"></a>Integer greater than or equal to 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p44288167393"><a name="p44288167393"></a><a name="p44288167393"></a>The value is 0 for the master node, and the value increases one by one for worker nodes.</p>
</td>
</tr>
<tr id="row205661943184311"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p175665433439"><a name="p175665433439"></a><a name="p175665433439"></a>LOCAL_WORLD_SIZE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p55661843164319"><a name="p55661843164319"></a><a name="p55661843164319"></a>Number of NPUs used per node Pod</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1087181334010"><a name="p1087181334010"></a><a name="p1087181334010"></a>Integer greater than or equal to 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p3566204318433"><a name="p3566204318433"></a><a name="p3566204318433"></a>For example, if a Pod uses 4 NPUs, configure it as 4.</p>
</td>
</tr>
<tr id="row138804664312"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1788154612438"><a name="p1788154612438"></a><a name="p1788154612438"></a>LOCAL_RANK</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p10677635145611"><a name="p10677635145611"></a><a name="p10677635145611"></a>List of logical IDs of NPUs used per node Pod</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1687119132409"><a name="p1687119132409"></a><a name="p1687119132409"></a>String</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p688746194315"><a name="p688746194315"></a><a name="p688746194315"></a>Configured based on the number of NPUs used by the Pod, starting from 0. For example, if a Pod uses 4 NPUs, the configuration is {0,1,2,3}.</p>
</td>
</tr>
<tr id="row16916943102412"><td class="cellrowborder" rowspan="6" valign="top" width="12.379999999999999%" headers="mcps1.2.6.1.1 "><p id="p9425933142120"><a name="p9425933142120"></a><a name="p9425933142120"></a><span id="ph3425633192112"><a name="ph3425633192112"></a><a name="ph3425633192112"></a>PyTorch</span>, MindSpore</p>
<p id="p37048619498"><a name="p37048619498"></a><a name="p37048619498"></a></p>
<p id="p7868336195017"><a name="p7868336195017"></a><a name="p7868336195017"></a></p>
<p id="p18298181492719"><a name="p18298181492719"></a><a name="p18298181492719"></a></p>
</td>
<td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p19161443172416"><a name="p19161443172416"></a><a name="p19161443172416"></a>HostNetwork</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p2093218562259"><a name="p2093218562259"></a><a name="p2093218562259"></a>Value of the hostNetwork field in the job YAML.</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><a name="ul960434424111"></a><a name="ul960434424111"></a><ul id="ul960434424111"><li>true: Creates a Pod using HostIP.</li><li>false: Creates a Pod without using HostIP.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p474032616460"><a name="p474032616460"></a><a name="p474032616460"></a>When the cluster is large (number of nodes &gt; 1000), it is recommended to create Pods using HostIP.</p>
</td>
</tr>
<tr id="row11721153544311"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p99031145174312"><a name="p99031145174312"></a><a name="p99031145174312"></a><span>MINDX_SERVER_IP</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p49794118443"><a name="p49794118443"></a><a name="p49794118443"></a>IP address for communication between the <span>job and</span> <span id="ph767616278495"><a name="ph767616278495"></a><a name="ph767616278495"></a>ClusterD</span><span>, which is also the svc ip of clusterd-grpc-svc.</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2021818824510"><a name="p2021818824510"></a><a name="p2021818824510"></a>A valid IP address. The format is a string and must be in standard IPv4 or IPv6 format.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p872210353431"><a name="p872210353431"></a><a name="p872210353431"></a>-</p>
</td>
</tr>
<tr id="row99115919216"><td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p1189132935"><a name="p1189132935"></a><a name="p1189132935"></a><span>HCCL_LOGIC_SUPERPOD_ID</span></p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p148917321033"><a name="p148917321033"></a><a name="p148917321033"></a>Chips with the same ID communicate using the UnifiedBus network, while chips with different IDs communicate using the RoCE network.</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><p id="p19891532930"><a name="p19891532930"></a><a name="p19891532930"></a>Integer greater than or equal to 0</p>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p1189232739"><a name="p1189232739"></a><a name="p1189232739"></a>HCCL uses this environment variable for dynamic networking to restrict the network communication mode between chips.</p>
<div class="note" id="note4836193915520"><a name="note4836193915520"></a><div class="notebody"><p id="p143153051215"><a name="p143153051215"></a><a name="p143153051215"></a>This environment variable is only supported under the following conditions:</p>
<a name="ul29353417120"></a><a name="ul29353417120"></a><ul id="ul29353417120"><li>Hardware: <span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD</span>.</li><li>Software: MindCluster 7.0.RC1 or later, CANN 8.0.0 or later.</li></ul>
</div></div>
</td>
</tr>
<tr id="row0703116194918"><td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p1022716206493"><a name="p1022716206493"></a><a name="p1022716206493"></a>MINDX_TASK_ID</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p922752054917"><a name="p922752054917"></a><a name="p922752054917"></a><span id="ph159662220018"><a name="ph159662220018"></a><a name="ph159662220018"></a>Elastic Agent</span>/<span id="ph126107511246"><a name="ph126107511246"></a><a name="ph126107511246"></a>TaskD</span> needs to provide the MINDX_TASK_ID information when registering the gRPC service with <span id="ph1722782017491"><a name="ph1722782017491"></a><a name="ph1722782017491"></a>ClusterD</span>.</p>
<p id="p11227154910536"><a name="p11227154910536"></a><a name="p11227154910536"></a>For the MindIE inference jobs, the value is the value of the jobID field under the label field in an acjob.</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><p id="p6227142012497"><a name="p6227142012497"></a><a name="p6227142012497"></a>String</p>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p1622752054919"><a name="p1622752054919"></a><a name="p1622752054919"></a>Job UID</p>
<p id="p7227102014916"><a name="p7227102014916"></a><a name="p7227102014916"></a></p>
</td>
</tr>
<tr id="row1586823610504"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p186863665020"><a name="p186863665020"></a><a name="p186863665020"></a>APP_TYPE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p128682367506"><a name="p128682367506"></a><a name="p128682367506"></a>The value is the value of the app field under the label field in an acjob.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3868133612507"><a name="p3868133612507"></a><a name="p3868133612507"></a>String</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p3868173675015"><a name="p3868173675015"></a><a name="p3868173675015"></a>-</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p>REPLICA_TYPE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p>Pod replica type.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p>String. The value is Master, Scheduler, Chief, or Worker.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p>-</p>
</td>
</tr>
<tr id="row8906345192017"><td class="cellrowborder" rowspan="8" valign="top" width="12.379999999999999%" headers="mcps1.2.6.1.1 "><p id="p687175715434"><a name="p687175715434"></a><a name="p687175715434"></a>MindSpore</p>
<p id="p16201203117487"><a name="p16201203117487"></a><a name="p16201203117487"></a></p>
<p id="p204163510439"><a name="p204163510439"></a><a name="p204163510439"></a></p>
<p id="p1711725164512"><a name="p1711725164512"></a><a name="p1711725164512"></a></p>
<p id="p1971017224517"><a name="p1971017224517"></a><a name="p1971017224517"></a></p>
<p id="p75734064516"><a name="p75734064516"></a><a name="p75734064516"></a></p>
<p id="p1477358184417"><a name="p1477358184417"></a><a name="p1477358184417"></a></p>
<p id="p1351443318348"><a name="p1351443318348"></a><a name="p1351443318348"></a></p>
<p id="p156585919429"><a name="p156585919429"></a><a name="p156585919429"></a></p>
</td>
<td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p11907257102019"><a name="p11907257102019"></a><a name="p11907257102019"></a>NPU_POD</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p13907155719207"><a name="p13907155719207"></a><a name="p13907155719207"></a>Marks whether the current pod has a chip mounted.</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><a name="ul590735782017"></a><a name="ul590735782017"></a><ul id="ul590735782017"><li>true: The current pod has a chip mounted.</li><li>false: The current pod does not have a chip mounted.</li></ul>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p1090715782017"><a name="p1090715782017"></a><a name="p1090715782017"></a>-</p>
</td>
</tr>
<tr id="row2871057114311"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p8871057144312"><a name="p8871057144312"></a><a name="p8871057144312"></a>MS_SERVER_NUM</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19871813154015"><a name="p19871813154015"></a><a name="p19871813154015"></a>Specifies the number of processes with the role MS_PSERVER.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2864162515474"><a name="p2864162515474"></a><a name="p2864162515474"></a>0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1874575436"><a name="p1874575436"></a><a name="p1874575436"></a></p><p>The PS mode is not currently supported. Set this to a fixed value of 0.</p><p>For detailed information about MS_PSERVER and the PS mode, see the relevant MindSpore documentation.</p>
</td>
</tr>
<tr id="row9716135318434"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p371613538438"><a name="p371613538438"></a><a name="p371613538438"></a>MS_WORKER_NUM</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p0716115364317"><a name="p0716115364317"></a><a name="p0716115364317"></a>Total number of NPUs used by the job</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p787121312405"><a name="p787121312405"></a><a name="p787121312405"></a>Integer greater than 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p9871142312514"><a name="p9871142312514"></a><a name="p9871142312514"></a>Total number of NPUs used by the job. For example, if a job uses 64 NPUs, the value is 64.</p>
</td>
</tr>
<tr id="row15416851194316"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1641613512434"><a name="p1641613512434"></a><a name="p1641613512434"></a>MS_LOCAL_WORKER</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1871114029"><a name="p1871114029"></a><a name="p1871114029"></a>Number of NPUs used per node Pod</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p10871213144015"><a name="p10871213144015"></a><a name="p10871213144015"></a>Integer greater than 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1599443785216"><a name="p1599443785216"></a><a name="p1599443785216"></a>For example, if a Pod uses 4 NPUs, configure it as 4.</p>
</td>
</tr>
<tr id="row611695124512"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4117195134517"><a name="p4117195134517"></a><a name="p4117195134517"></a>MS_SCHED_HOST</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p115148526440"><a name="p115148526440"></a><a name="p115148526440"></a>IP address of the Scheduler</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15871161312408"><a name="p15871161312408"></a><a name="p15871161312408"></a>Valid IP address, in the format of string; must be in standard IPv4 or IPv6 format.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><a name="ul134891523153513"></a><a name="ul134891523153513"></a><ul id="ul134891523153513"><li>In the Scheduler Pod, set to podIP.</li><li>In the Worker Pod, set to the clusterIP of the Scheduler Pod's corresponding svc.</li></ul>
</td>
</tr>
<tr id="row1471013244518"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1271016264511"><a name="p1271016264511"></a><a name="p1271016264511"></a>MS_SCHED_PORT</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1171010224518"><a name="p1171010224518"></a><a name="p1171010224518"></a>Port for communicating with the Scheduler</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p13871613144013"><a name="p13871613144013"></a><a name="p13871613144013"></a>Port number in the range of 1024 to 65535.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p136821316145311"><a name="p136821316145311"></a><a name="p136821316145311"></a>Value of the ascendjob-port field in the corresponding Scheduler Pod's svc. The default value is 2222.</p>
</td>
</tr>
<tr id="row55726034515"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p17573120164511"><a name="p17573120164511"></a><a name="p17573120164511"></a>MS_ROLE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1457320184519"><a name="p1457320184519"></a><a name="p1457320184519"></a>Specifies the role of this process.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul9226735436"></a><a name="ul9226735436"></a><ul id="ul9226735436"><li>MS_SCHED: Scheduler process. Only one Scheduler is started per training job, responsible for networking, container recovery, etc., <strong id="b18226143174315"><a name="b18226143174315"></a><a name="b18226143174315"></a>and does not execute training code.</strong>.</li><li>MS_WORKER: Worker process. Distributed training processes are generally set to this role.</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1457316017453"><a name="p1457316017453"></a><a name="p1457316017453"></a>The Worker process registers with the Scheduler process to complete networking.</p>
</td>
</tr>
<tr id="row9477165812440"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1747775874419"><a name="p1747775874419"></a><a name="p1747775874419"></a>MS_NODE_RANK</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p91701339918"><a name="p91701339918"></a><a name="p91701339918"></a>Node rank of the Pod on this node</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3871413144015"><a name="p3871413144015"></a><a name="p3871413144015"></a>Integer greater than or equal to 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p5312103318576"><a name="p5312103318576"></a><a name="p5312103318576"></a>0 for the Scheduler Pod.</p>
<a name="ul350115366586"></a><a name="ul350115366586"></a><ul id="ul350115366586"><li>When the Scheduler mounts chips, Worker Pods start incrementing from 1.</li><li>When the Scheduler does not mount chips, Worker Pods start incrementing from 0.</li></ul>
</td>
</tr>
<tr id="row1058205923118"><td class="cellrowborder" valign="top" width="12.379999999999999%" headers="mcps1.2.6.1.1 "><p id="p1181155914329"><a name="p1181155914329"></a><a name="p1181155914329"></a><span id="ph1551815244211"><a name="ph1551815244211"></a><a name="ph1551815244211"></a>PyTorch</span>, MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p318119595325"><a name="p318119595325"></a><a name="p318119595325"></a>PROCESS_RECOVER</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p1318195912322"><a name="p1318195912322"></a><a name="p1318195912322"></a>Master switch for process-level rescheduling, process-level online recovery, and elastic training.</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><a name="ul65501334344"></a><a name="ul65501334344"></a><ul id="ul65501334344"><li>on: Enable this feature.</li><li>off: Disable this feature.</li></ul>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p682072453313"><a name="p682072453313"></a><a name="p682072453313"></a>Inject this environment variable in process-level rescheduling, process-level online recovery, process-level in-place recovery, and elastic training scenarios.</p>
</td>
</tr>
<tr id="row242413586587"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p311314413594"><a name="p311314413594"></a><a name="p311314413594"></a><span id="ph611313425919"><a name="ph611313425919"></a><a name="ph611313425919"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p151133465910"><a name="p151133465910"></a><a name="p151133465910"></a>HIGH_AVAILABILITY</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1911315414598"><a name="p1911315414598"></a><a name="p1911315414598"></a>Switch for the MindSpeed-LLM process-level recovery feature.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p61131485917"><a name="p61131485917"></a><a name="p61131485917"></a>Available recovery strategies.</p>
<a name="ul2113204145911"></a><a name="ul2113204145911"></a><ul id="ul2113204145911"><li>retry: Process-level online recovery</li><li>recover: Process-level rescheduling</li><li>dump: Save dying gasps</li><li>elastic-training: Elastic training</li></ul>
</td>
</tr>
<tr id="row83631024143218"><td class="cellrowborder" valign="top" width="12.379999999999999%" headers="mcps1.2.6.1.1 "><p id="p1018165914322"><a name="p1018165914322"></a><a name="p1018165914322"></a><span id="ph12546182614219"><a name="ph12546182614219"></a><a name="ph12546182614219"></a>PyTorch</span>, MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p118117596323"><a name="p118117596323"></a><a name="p118117596323"></a>ELASTIC_PROCESS_RECOVER_ENABLE</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p1518175973216"><a name="p1518175973216"></a><a name="p1518175973216"></a><span id="ph1072282311518"><a name="ph1072282311518"></a><a name="ph1072282311518"></a>Elastic Agent</span>-side switch for process-level rescheduling, process-level online recovery, and dying gasp checkpoint features.</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><a name="ul167945693511"></a><a name="ul167945693511"></a><ul id="ul167945693511"><li>Value is 1: Enable.</li><li>Value is other values: Disable. When this variable feature is, the related features on the MindIO side must also be disabled.</li></ul>
</td>
<td class="cellrowborder" rowspan="4" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p923972485920"><a name="p923972485920"></a><a name="p923972485920"></a>Inject this environment variable in process-level rescheduling, process-level online recovery, and process-level in-place recovery scenarios.</p>
</td>
</tr>
<tr id="row0765193853210"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1918295993216"><a name="p1918295993216"></a><a name="p1918295993216"></a><span id="ph5168103016219"><a name="ph5168103016219"></a><a name="ph5168103016219"></a>PyTorch</span>, MindSpore</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17182145910321"><a name="p17182145910321"></a><a name="p17182145910321"></a>ENABLE_RESTART_FAULT_PROCESS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p71821759193216"><a name="p71821759193216"></a><a name="p71821759193216"></a><span id="ph249610307518"><a name="ph249610307518"></a><a name="ph249610307518"></a>Process-level</span>/<span id="ph1513354715617"><a name="ph1513354715617"></a><a name="ph1513354715617"></a>in-place</span> recovery switch for Elastic Agent and TaskD.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><a name="ul1032320399361"></a><a name="ul1032320399361"></a><ul id="ul1032320399361"><li>on: Enable this feature.</li><li>Other values: Disable this feature.</li></ul>
<div class="note" id="note21949542365"><a name="note21949542365"></a><a name="note21949542365"></a><span class="notetitle">Note:</span><div class="notebody"><a name="ul11833105863616"></a><a name="ul11833105863616"></a><ul id="ul11833105863616"><li>Under the <span id="ph1982841320618"><a name="ph1982841320618"></a><a name="ph1982841320618"></a>PyTorch</span> framework, this feature is provided by <span id="ph193151321661"><a name="ph193151321661"></a><a name="ph193151321661"></a>Elastic Agent/TaskD</span>.</li><li>Under the MindSpore framework, this feature is provided by <span id="ph5518105017616"><a name="ph5518105017616"></a><a name="ph5518105017616"></a>TaskD</span>.</li></ul>
</div></div>
</td>
</tr>
<tr id="row1276511386323"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p618295953218"><a name="p618295953218"></a><a name="p618295953218"></a>MindSpore</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p518245917328"><a name="p518245917328"></a><a name="p518245917328"></a>MINDIO_FOR_MINDSPORE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p818275973217"><a name="p818275973217"></a><a name="p818275973217"></a>Switch for MindIO to start MindSpore.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1418212592323"><a name="p1418212592323"></a><a name="p1418212592323"></a><ul><li>Value is 1: Enable the switch for MindIO to start MindSpore.</li><li>Value is not 1: Disable the switch for MindIO to start MindSpore.</li></ul></p>
</td>
</tr>
<tr id="row10116337329"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p14182125983213"><a name="p14182125983213"></a><a name="p14182125983213"></a>MindSpore</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1518285953219"><a name="p1518285953219"></a><a name="p1518285953219"></a>MS_ENABLE_TFT</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p18182105963210"><a name="p18182105963210"></a><a name="p18182105963210"></a>MindSpore process-level recovery switch.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><pre class="screen" id="screen15182185953212"><a name="screen15182185953212"></a><a name="screen15182185953212"></a>{TTP:1,UCE:1,ARF:1,HCCE:1,RSC:1}    # String type, respectively enable dying gasp, process-level online recovery for <span id="ph15161018131912"><a name="ph15161018131912"></a><a name="ph15161018131912"></a>on-chip memory</span> faults, process-level rescheduling, process-level online recovery for network faults, and pod-level rescheduling</pre><div class="note"><span class="notetitle">Note:</span><div class="notebody">For detailed descriptions of the above fields, see the "MS_ENABLE_TFT" environment variable in the <a href="https://www.mindspore.cn/docs/en/stable/api_python/env_var_list.html#%E4%B8%89%E6%96%B9%E5%BA%93">MindSpore documentation</a>.</div></div>
</td>
</tr>
</tbody>
</table>

## Ascend Docker Runtime Environment Variable<a name="section109964810209"></a>

Ascend Docker Runtime injects the environment variable below into the container.

<a name="table974781182117"></a>

| Environment Variable | Function | Value | Description |
|--|--|--|--|
| ASCEND_DOCKER_RUNTIME | Identifies whether Ascend Docker Runtime is installed in the current environment. | True | This environment variable does not exist when Ascend Docker Runtime is not installed. |

## Ascend Device Plugin Environment Variables<a name="section1419516175219"></a>

Ascend Device Plugin injects corresponding environment variables into the container. See the following table for descriptions of these environment variables.

**Table 3** Environment variables injected by Ascend Device Plugin into the container

<a name="table4446195872218"></a>

| Environment Variable | Function | Value | Description |
|--|--|--|--|
| ASCEND_VISIBLE_DEVICES | If a job requires NPU devices, ASCEND_VISIBLE_DEVICES must be used to specify the NPU devices to be mounted into the container; otherwise, NPU device mounting fails. When specifying devices by device index, both individual and range specifications are supported, and they can be used together. When specifying devices by chip name, multiple chip names of the same type can be specified simultaneously. | <ul><li>Mounting physical chips (NPUs)<ul><li>ASCEND_VISIBLE_DEVICES=0 indicates that NPU device 0 (/dev/davinci0) is mounted into the container.</li><li>ASCEND_VISIBLE_DEVICES=1,3 indicates that NPU devices 1 and 3 are mounted into the container.</li></ul></li><li>Mounting virtual chips (vNPUs)</li><ul><li>**Static virtualization**: The usage is the same as for physical chips; simply replace the physical chip ID with the virtual chip ID (vNPU ID).</li><li>**Dynamic virtualization**: ASCEND_VISIBLE_DEVICES=0 indicates that a certain number of AICores are allocated from NPU device 0.</li></ul></ul> | - |
| ASCEND_ALLOW_LINK | Whether to allow soft links in mounted files or directories. This parameter must be specified in Atlas 500 A2 Intelligent Station, Atlas 200I A2 Accelerator Module, and Atlas 200I DK A2 scenarios. | <ul><li>ASCEND_ALLOW_LINK=True indicates that mounting driver files with soft links is allowed in Atlas 500 A2 Intelligent Station, Atlas 200I A2 Accelerator Module, and Atlas 200I DK A2 scenarios.</li><li>ASCEND_ALLOW_LINK=False or if this parameter is not specified, Ascend Docker Runtime cannot be used on Atlas 500 A2 Intelligent Station, Atlas 200I A2 Accelerator Module, and Atlas 200I DK A2.</li></ul> | - |
| ASCEND_RUNTIME_OPTIONS | Imposes restrictions on the chip IDs specified in the ASCEND_VISIBLE_DEVICES parameter: <ul><li>NODRV: Indicates that driver-related directories are not mounted.</li><li>VIRTUAL: Indicates that the mounted chip is a virtual chip.</li><li>NODRV,VIRTUAL: Indicates that the mounted chip is a virtual chip and driver-related directories are not mounted.</li></ul> | <ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul> | - |
| WORLD_SIZE | Total number of NPUs used by a job | Integer greater than or equal to 0 | Written only in dynamic vNPU scheduling scenarios |
| LOCAL_WORLD_SIZE | Number of NPUs used by each node's Pod | Integer greater than or equal to 0 | Written only in dynamic vNPU scheduling scenarios |
| LOCAL_RANK | List of logical IDs of NPUs used by each node's Pod | String | Written only in dynamic vNPU scheduling scenarios. The value starts from 0. For example, if a Pod uses 4 NPUs, the configuration is {0,1,2,3}. |
| CM_WORKER_SIZE | Total number of NPUs used by a job | Integer greater than or equal to 0 | Written only in dynamic vNPU scheduling scenarios |
| CM_LOCAL_WORKER | Number of NPUs used by each node's Pod | Integer greater than or equal to 0 | Written only in dynamic vNPU scheduling scenarios |
| MS_WORKER_NUM | Total number of NPUs used by a job | Integer greater than or equal to 0 | Written only in dynamic vNPU scheduling scenarios |
| MS_LOCAL_WORKER | Number of NPUs used by each node's Pod | Integer greater than or equal to 0 | Written only in dynamic vNPU scheduling scenarios |
| PERF_DUMP_PATH | Path for saving iteration latency and grouping information | String | Written only in slow node detection scenarios |
| PERF_DUMP_CONFIG | Start/stop switch for iteration latency and grouping information | String | Written only in slow node detection scenarios |
| KUBELET_PORT | Specifies the default port number of kubelet on the current node (if the user has not customized the kubelet port, no configuration is required). | Integer from 0 to 65535 | If the user modifies the default kubelet port, the value of this environment variable must be set to the custom port number.<p>If the user has not modified the default kubelet port, this environment variable is ignored.</p> |
| HOST_IP | Specifies the physical IP address of the current node. | Valid IP address in the format of string; must be in standard IPv4 or IPv6 format | Fixed configuration item, provided in the initial YAML file. |

## Elastic Agent Environment Variables<a name="section8853192413411"></a>

>[!NOTE]
>Elastic Agent has reached its end of life, and related information will be removed in the version released on December 30, 2026.

The table below describes environment variables that can be configured when Elastic Agent is used. For other environment variables from the source code, see [PyTorch Documentation](https://pytorch.ac.cn/#google_vignette).

**Table 4** Elastic Agent environment variables

<a name="table159711045543"></a>

| Environment Variable | Function | Value | Description |
| --- | --- | --- | --- |
| ELASTIC\_LOG\_PATH | Specifies the disk path for Elastic Agent running logs. | String | When configuring, differentiate the node name for this log. Reference example: <pre class="screen">ELASTIC_LOG_PATH=/job/code/alllogs/\$MINDX_TASK_ID/elasticlogs/elastic-log\$XDL_IP-\$RANK</pre><ul><li>Replace \$XDL_IP with the actual node IP.</li><li>Replace \$RANK with the actual node RANK.</li></ul> |
| ELASTIC\_PROCESS\_RECOVER\_ENABLE | Switch for Elastic Agent-side process-level rescheduling, process-level online recovery, and dying gasp checkpoint recovery. | String | <ul><li>1: Enable this feature.</li><li>Other values: Disable this feature.</li></ul> When this feature is disabled, the related features on the MindIO side must also be disabled. |
| ENABLE\_RESTART\_FAULT\_PROCESS | Switch for Elastic Agent to enable the process-level in-place recovery feature. | String | <ul><li>on: Enable this feature.</li><li>Other values: Disable this feature.</li></ul> |
| RESTART\_FAULT\_PROCESS\_TYPE | Type of faulty process restart that Elastic Agent notifies MindIO to perform. | String | <ul><li>worker: Do not exit the Pod, and only restart the faulty process.</li><li>pod: Restart the Pod.</li></ul> |
| RANK\_TABLE\_FILE | RankTable file path. | String | Path to the hccl.json file. |
| PROCESS\_RECOVER | Switch for process-level rescheduling or process-level online recovery. | String | <ul><li>on: Enable this feature.</li><li>Other values: Disable this feature.</li></ul> |

## TaskD Environment Variables<a name="section6616275583"></a>

The table below describes environment variables that can be configured when TaskD is used. For other environment variables from the source code, see [PyTorch Documentation](https://pytorch.ac.cn/#google_vignette).

**Table 5** TaskD environment variables

<a name="table13568156155815"></a>

| Environment Variable | Function | Value | Description |
|--|--|--|--|
| TASKD_LOG_PATH | Specifies the disk path for storing TaskD running logs. | String | If not specified, the default path is used: ./taskd_log/, which is the taskd_log directory under the current execution path. The following logs are generated based on different node configurations: <ul><li>manager.log: TaskD Manager log</li><li>taskd.log: TaskD Python-side log</li><li>agent-{*RANK*}.log: TaskD Agent log</li><li>taskd-proxy-{*RANK*}-{*TIMESTAMP*}.log: TaskD Proxy log</li><li>taskd-worker-{*RANK*}.log: TaskD Worker log</li></ul><p><i>\{RANK\}</i> is the global rank number of the current training process, and <i>\{TIMESTAMP\}</i> is the timestamp.</p> |
| TASKD_FILE_LOG_LEVEL | Specifies the log level to be recorded in the log file. | String | Value: <ul><li>DEBUG: Debug information</li><li>INFO: General information (default level)</li><li>WARNING: Warning information</li><li>ERROR: Error information</li></ul> |
| TASKD_STD_LOG_LEVEL | Specifies the log level to be printed. | String | Value is: <ul><li>DEBUG: Debug information</li><li>INFO: General information (default level)</li><li>WARNING: Warning information</li><li>ERROR: Error information</li></ul> |
| TASKD_LOG_STDOUT | Specifies whether logs need to be printed. | bool | Value is True or False. Defaults to True if not configured. |
| ENABLE_RESTART_FAULT_PROCESS | Switch for the TaskD component to enable the process-level in-place recovery feature. | String | Value: <ul><li>on: Enable this feature</li><li>Other Values: Disable this feature</li></ul> |
| RESTART_FAULT_PROCESS_TYPE | Type of notification from TaskD to MindIO for restarting the faulty process. | String | Value: <ul><li>worker: Do not exit the pod, and only restart the faulty process</li><li>pod: Restart the pod</li></ul> |
| TASKD_PROCESS_ENABLE | Switch for TaskD to enable process-level rescheduling, process-level online recovery, process-level in-place recovery, and elastic training features. | String | Value: <ul><li>on: Enable this feature</li><li>off: Disable this feature</li></ul> |
| LOCAL_PROXY_ENABLE | Whether to enable the local proxy (required for security hardening). | String | Value: <ul><li>on: Enable this feature</li><li>off: Disable this feature</li></ul>Default Value is "off". It must be set to "on" for communication security hardening scenarios. |
| HCCL_ASYNC_ERROR_HANDLING | Whether to enable the watchdog feature. | String | Value: <ul><li>0: Disables fault detection and process exit features.</li><li>1: Enables fault detection and process exit features.</li><li>2: Enables only the fault detection feature.</li></ul>Default Value is 1. |
| TASKD_PROCESS_INTERVAL | Sets the processing interval for the TaskD Manager main process. | String | Value ranging from 100 to 1000, in milliseconds. |
| TASKD_REPORT_FAULT_TIMEOUT | Timeout for TaskD Agent to report service faults to TaskD Manager. If the fault persists after the timeout, TaskD Agent exits. | String | Value ranging from 300 to 600, in seconds. |

## NodeD Environment Variable<a name="section10131935141216"></a>

**Table 6** NodeD environment variable

<a name="table11131133571214"></a>

| Environment Variable | Function | Value | Description |
|--|--|--|--|
| XDL_IP | Used to obtain the IP address of the host where the pod resides. Used by slow nodes to record and match slow node information. | A valid IP address in string format, either regular IPv4 or IPv6. | Write this environment variable in the YAML for deploying NodeD. |
