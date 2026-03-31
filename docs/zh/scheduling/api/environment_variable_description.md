# 环境变量说明<a name="ZH-CN_TOPIC_0000002479226386"></a>

**MindCluster组件使用的环境变量<a name="section1562121818463"></a>**

MindCluster组件使用的环境变量说明如[表1](#table1132513610543)所示。

**表 1**  环境变量说明

<a name="table1132513610543"></a>

|环境变量名称|来源|是否必选|取值|作用|
|--|--|--|--|--|
|POD_IP|部署组件的YAML中写入|是|当前容器所在Pod的Pod IP|ClusterD、TaskD用于启动gRPC服务|
|POD_UID|部署组件的YAML中写入|否|当前容器所在Pod的PodUID|用于解析ranktable文件的server_id字段|
|ASCEND_DOCKER_RUNTIME|容器创建时Ascend Docker Runtime写入|否|"true"|Ascend Device Plugin用于判断当前节点容器默认运行时是否是Ascend Docker Runtime|
|HOSTNAME|K8s创建容器时写入|是|当前容器所在Pod的Pod名称|Ascend Device Plugin用于获取当前Pod名称|
|NODE_NAME|部署组件的YAML中写入|是|当前容器所在节点的节点名称|Ascend Device Plugin、NodeD、ClusterD用于获取当前节点名称|
|LD_LIBRARY_PATH|Dockerfile中写入|是|文件路径|Ascend Device Plugin和NPU Exporter用于初始化DCMI|
|BATCH_BIND_NUM|-|否|数字字符串|指定Volcano设置批量绑定Pod的数量|
|MULTI_SCHEDULER_ENABLE|-|否|"true"或者"false"|指定Volcano是否是多调度器场景|
|SCHEDULER_POD_NAME|-|否|字符串|指定Volcano调度器Pod名称|
|SCHEDULER_NUM|-|否|数字字符串|指定Volcano调度器数量|
|PANIC_ON_ERROR|-|否|"true"或者"false"|指定Volcano调度器发生错误时是否需要panic|
|KUBECONFIG|-|否|文件路径|指定Volcano连接K8s api-server的kubeconfig路径|
|HOME|K8s创建容器时写入|是|文件夹路径|指定Volcano获取当前用户home路径|
|DEBUG_SOCKET_DIR|-|否|socket文件路径|指定Volcano侦听的socket路径|
|HCCL_CONNECT_TIMEOUT|训练脚本中写入|否|HCCL建链的超时时间|表示建链超时时间|
|TTP_PORT|部署组件的YAML中写入|是|MindIO TTP用的通信端口|用于启动MindIO Controller|
|SSH_CLIENT|SSH 服务器设置的环境变量，它包含有关客户端连接的信息|是|当前客户端连接的信息|安装Ascend Docker Runtime时，记录该信息到操作日志中|
|TASKD_LOG_PATH|-|否|字符串|表示TaskD组件运行日志的落盘路径|
|MINDX_SERVER_IP|容器创建时由Ascend Operator写入|是|字符串|表示任务与ClusterD通信的IP地址，同时也是clusterd-grpc-svc的svc IP|
|MINDX_SERVER_DOMAIN|容器创建时由Ascend Operator写入|是|字符串|表示任务与ClusterD通信的域名，默认值为"clusterd-grpc-svc.mindx-dl.svc.cluster.local"|
|MINDX_TASK_ID|容器创建时Ascend Operator写入|否|MindIE推理任务场景下，取值为acjob任务中label字段下jobID字段的值|Elastic Agent/TaskD向ClusterD注册gRPC服务和TaskD profiling功能保存日志需要提供MINDX_TASK_ID信息|
|GROUP_BASE_DIR|任务启动脚本中写入|否|文件夹路径|表示TaskD组件的并行域信息导出路径|
|MINDIO_WAIT_MINDX_TIME|任务YAML中写入|否|数字字符串，取值范围为[1, 3600]|不开启进程级重调度，开启弹性训练时等待故障Pod调度的超时时间|
|RAS_NET_ROOT_PATH|用户配置|否|ClusterD和NodeD共享目录的根路径|在慢网络诊断场景下ClusterD和NodeD通过共享存储进行交互，详细请参见[慢网络诊断](../usage/resumable_training.md#慢网络诊断)|
|REPLICA_TYPE|容器创建时由Ascend Operator写入|是|Master、Scheduler、Chief或Worker|Pod副本类型|

**Ascend Operator环境变量说明<a name="section1272862810184"></a>**

Ascend Operator为不同AI框架的分布式训练任务（acjob）提供相应的环境变量，该环境变量的相关说明如下表所示。

**表 2** Ascend Operator注入的训练环境变量

<a name="table154271816163912"></a>
<table><thead align="left"><tr id="row2428151693919"><th class="cellrowborder" valign="top" width="12.379999999999999%" id="mcps1.2.6.1.1"><p id="p13428016113914"><a name="p13428016113914"></a><a name="p13428016113914"></a>框架名称</p>
</th>
<th class="cellrowborder" valign="top" width="16.869999999999997%" id="mcps1.2.6.1.2"><p id="p194281416103914"><a name="p194281416103914"></a><a name="p194281416103914"></a>环境变量名称</p>
</th>
<th class="cellrowborder" valign="top" width="27.77%" id="mcps1.2.6.1.3"><p id="p1342841653915"><a name="p1342841653915"></a><a name="p1342841653915"></a>功能</p>
</th>
<th class="cellrowborder" valign="top" width="19.79%" id="mcps1.2.6.1.4"><p id="p18871191318405"><a name="p18871191318405"></a><a name="p18871191318405"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="23.189999999999998%" id="mcps1.2.6.1.5"><p id="p64281016193910"><a name="p64281016193910"></a><a name="p64281016193910"></a>说明</p>
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
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p4428181673917"><a name="p4428181673917"></a><a name="p4428181673917"></a>与Master节点通信的IP地址</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><p id="p5871413104013"><a name="p5871413104013"></a><a name="p5871413104013"></a>合法的IP地址，格式为字符串，要求为常规IPv4或IPv6格式</p>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><a name="ul695319973016"></a><a name="ul695319973016"></a><ul id="ul695319973016"><li>Master Pod中设置为podIP。</li><li>Worker Pod中设置为Master Pod对应svc的clusterIP。</li></ul>
</td>
</tr>
<tr id="row84281516153918"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p17428016193912"><a name="p17428016193912"></a><a name="p17428016193912"></a>MASTER_PORT</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p642871613915"><a name="p642871613915"></a><a name="p642871613915"></a>与Master节点通信的端口</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p196391835162819"><a name="p196391835162819"></a><a name="p196391835162819"></a>支持配置为字符串、数字，取值范围为0~65520</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p182011320145513"><a name="p182011320145513"></a><a name="p182011320145513"></a>Master Pod对应svc中名称为ascendjob-port的值，默认为2222。</p>
</td>
</tr>
<tr id="row1542861610390"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p13428161673916"><a name="p13428161673916"></a><a name="p13428161673916"></a>WORLD_SIZE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p124284165399"><a name="p124284165399"></a><a name="p124284165399"></a>任务使用的总NPU数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1481964632819"><a name="p1481964632819"></a><a name="p1481964632819"></a>大于0的整数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p442812163396"><a name="p442812163396"></a><a name="p442812163396"></a>任务使用的总卡数，例如64个NPU任务，则取值为64。</p>
</td>
</tr>
<tr id="row1428216163912"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p204282016153919"><a name="p204282016153919"></a><a name="p204282016153919"></a>RANK</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p210753716532"><a name="p210753716532"></a><a name="p210753716532"></a>本节点Pod的Node Rank</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p0871121315406"><a name="p0871121315406"></a><a name="p0871121315406"></a>大于或等于0的整数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p44288167393"><a name="p44288167393"></a><a name="p44288167393"></a>Master为0，Worker从1开始逐一增加。</p>
</td>
</tr>
<tr id="row205661943184311"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p175665433439"><a name="p175665433439"></a><a name="p175665433439"></a>LOCAL_WORLD_SIZE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p55661843164319"><a name="p55661843164319"></a><a name="p55661843164319"></a>每个节点Pod使用的NPU数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1087181334010"><a name="p1087181334010"></a><a name="p1087181334010"></a>大于或等于0的整数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p3566204318433"><a name="p3566204318433"></a><a name="p3566204318433"></a>例如Pod使用4个NPU，则配置为4。</p>
</td>
</tr>
<tr id="row138804664312"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1788154612438"><a name="p1788154612438"></a><a name="p1788154612438"></a>LOCAL_RANK</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p10677635145611"><a name="p10677635145611"></a><a name="p10677635145611"></a>每个节点Pod使用的NPU的逻辑ID列表</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1687119132409"><a name="p1687119132409"></a><a name="p1687119132409"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p688746194315"><a name="p688746194315"></a><a name="p688746194315"></a>根据Pod使用NPU数量进行配置，从0开始。例如，Pod使用4个NPU，则配置为{0,1,2,3}。</p>
</td>
</tr>
<tr id="row16916943102412"><td class="cellrowborder" rowspan="6" valign="top" width="12.379999999999999%" headers="mcps1.2.6.1.1 "><p id="p9425933142120"><a name="p9425933142120"></a><a name="p9425933142120"></a><span id="ph3425633192112"><a name="ph3425633192112"></a><a name="ph3425633192112"></a>PyTorch</span>、MindSpore、<span id="ph1742583312216"><a name="ph1742583312216"></a><a name="ph1742583312216"></a>TensorFlow</span></p>
<p id="p37048619498"><a name="p37048619498"></a><a name="p37048619498"></a></p>
<p id="p7868336195017"><a name="p7868336195017"></a><a name="p7868336195017"></a></p>
<p id="p18298181492719"><a name="p18298181492719"></a><a name="p18298181492719"></a></p>
</td>
<td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p19161443172416"><a name="p19161443172416"></a><a name="p19161443172416"></a>HostNetwork</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p2093218562259"><a name="p2093218562259"></a><a name="p2093218562259"></a>表示当前任务YAML的hostNetwork字段的值。</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><a name="ul960434424111"></a><a name="ul960434424111"></a><ul id="ul960434424111"><li>true：使用HostIP创建Pod。</li><li>false：不使用HostIP创建Pod。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p474032616460"><a name="p474032616460"></a><a name="p474032616460"></a>当集群规模较大（节点数量&gt;1000时），推荐使用HostIP创建Pod。</p>
</td>
</tr>
<tr id="row11721153544311"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p99031145174312"><a name="p99031145174312"></a><a name="p99031145174312"></a><span>MINDX_SERVER_IP</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p49794118443"><a name="p49794118443"></a><a name="p49794118443"></a>表示<span>任务与</span><span id="ph767616278495"><a name="ph767616278495"></a><a name="ph767616278495"></a>ClusterD</span><span>通信的IP地址，同时也是</span>clusterd-grpc-svc的svc ip。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2021818824510"><a name="p2021818824510"></a><a name="p2021818824510"></a>合法的IP地址，格式为字符串，要求为常规IPv4或IPv6格式。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p872210353431"><a name="p872210353431"></a><a name="p872210353431"></a>-</p>
</td>
</tr>
<tr id="row99115919216"><td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p1189132935"><a name="p1189132935"></a><a name="p1189132935"></a><span>HCCL_LOGIC_SUPERPOD_ID</span></p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p148917321033"><a name="p148917321033"></a><a name="p148917321033"></a>相同ID的芯片间使用灵衢网络通信，不同ID的芯片间使用RoCE网络通信。</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><p id="p19891532930"><a name="p19891532930"></a><a name="p19891532930"></a>大于或等于0的整数</p>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p1189232739"><a name="p1189232739"></a><a name="p1189232739"></a>HCCL使用此环境变量用于动态组网，限制芯片间网络通信方式。</p>
<div class="note" id="note4836193915520"><a name="note4836193915520"></a><div class="notebody"><p id="p143153051215"><a name="p143153051215"></a><a name="p143153051215"></a>当前环境变量仅支持在以下条件下使用：</p>
<a name="ul29353417120"></a><a name="ul29353417120"></a><ul id="ul29353417120"><li>硬件：<span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span>。</li><li>软件：MindCluster 7.0.RC1及以上版本、CANN 8.0.0及以上版本。</li></ul>
</div></div>
</td>
</tr>
<tr id="row0703116194918"><td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p1022716206493"><a name="p1022716206493"></a><a name="p1022716206493"></a>MINDX_TASK_ID</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p922752054917"><a name="p922752054917"></a><a name="p922752054917"></a><span id="ph159662220018"><a name="ph159662220018"></a><a name="ph159662220018"></a>Elastic Agent</span>/<span id="ph126107511246"><a name="ph126107511246"></a><a name="ph126107511246"></a>TaskD</span>向<span id="ph1722782017491"><a name="ph1722782017491"></a><a name="ph1722782017491"></a>ClusterD</span>注册gRPC服务需要提供MINDX_TASK_ID信息。</p>
<p id="p11227154910536"><a name="p11227154910536"></a><a name="p11227154910536"></a>MindIE推理任务场景下，取值为acjob任务中label字段下jobID字段的值。</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><p id="p6227142012497"><a name="p6227142012497"></a><a name="p6227142012497"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p1622752054919"><a name="p1622752054919"></a><a name="p1622752054919"></a>任务的UID。</p>
<p id="p7227102014916"><a name="p7227102014916"></a><a name="p7227102014916"></a></p>
</td>
</tr>
<tr id="row1586823610504"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p186863665020"><a name="p186863665020"></a><a name="p186863665020"></a>APP_TYPE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p128682367506"><a name="p128682367506"></a><a name="p128682367506"></a>取值为acjob任务中label字段下app字段的值。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3868133612507"><a name="p3868133612507"></a><a name="p3868133612507"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p3868173675015"><a name="p3868173675015"></a><a name="p3868173675015"></a>-</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p>REPLICA_TYPE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p>Pod副本类型。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p>字符串，取值为Master、Scheduler、Chief或Worker</p>
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
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p13907155719207"><a name="p13907155719207"></a><a name="p13907155719207"></a>标记当前Pod是否挂载了芯片。</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><a name="ul590735782017"></a><a name="ul590735782017"></a><ul id="ul590735782017"><li>true：当前pod已挂载芯片。</li><li>false：当前pod未挂载芯片。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p1090715782017"><a name="p1090715782017"></a><a name="p1090715782017"></a>-</p>
</td>
</tr>
<tr id="row2871057114311"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p8871057144312"><a name="p8871057144312"></a><a name="p8871057144312"></a>MS_SERVER_NUM</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19871813154015"><a name="p19871813154015"></a><a name="p19871813154015"></a>指定角色为MS_PSERVER的进程数量</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2864162515474"><a name="p2864162515474"></a><a name="p2864162515474"></a>0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1874575436"><a name="p1874575436"></a><a name="p1874575436"></a><p>暂不支持PS模式，设置固定值0。</p><p>关于MS_PSERVER和PS模式的详细说明请参见MindSpore相关文档。</p></p>
</td>
</tr>
<tr id="row9716135318434"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p371613538438"><a name="p371613538438"></a><a name="p371613538438"></a>MS_WORKER_NUM</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p0716115364317"><a name="p0716115364317"></a><a name="p0716115364317"></a>任务使用的总NPU数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p787121312405"><a name="p787121312405"></a><a name="p787121312405"></a>大于0的整数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p9871142312514"><a name="p9871142312514"></a><a name="p9871142312514"></a>任务使用的总NPU数，例如64个NPU任务，则取值为64。</p>
</td>
</tr>
<tr id="row15416851194316"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1641613512434"><a name="p1641613512434"></a><a name="p1641613512434"></a>MS_LOCAL_WORKER</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1871114029"><a name="p1871114029"></a><a name="p1871114029"></a>每个节点Pod使用的NPU数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p10871213144015"><a name="p10871213144015"></a><a name="p10871213144015"></a>大于0的整数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1599443785216"><a name="p1599443785216"></a><a name="p1599443785216"></a>例如Pod使用4个NPU，则配置为4。</p>
</td>
</tr>
<tr id="row611695124512"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4117195134517"><a name="p4117195134517"></a><a name="p4117195134517"></a>MS_SCHED_HOST</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p115148526440"><a name="p115148526440"></a><a name="p115148526440"></a>指定Scheduler的IP地址</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15871161312408"><a name="p15871161312408"></a><a name="p15871161312408"></a>合法的IP地址，格式为字符串，要求为常规IPv4或IPv6格式</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><a name="ul134891523153513"></a><a name="ul134891523153513"></a><ul id="ul134891523153513"><li>Scheduler Pod中设置为podIP</li><li>Worker Pod设置为Scheduler Pod对应svc的clusterIP。</li></ul>
</td>
</tr>
<tr id="row1471013244518"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1271016264511"><a name="p1271016264511"></a><a name="p1271016264511"></a>MS_SCHED_PORT</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1171010224518"><a name="p1171010224518"></a><a name="p1171010224518"></a>与Scheduler通信的端口</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p13871613144013"><a name="p13871613144013"></a><a name="p13871613144013"></a>1024～65535范围内的端口号。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p136821316145311"><a name="p136821316145311"></a><a name="p136821316145311"></a>Scheduler Pod对应svc中名称为ascendjob-port的值，默认取值为2222。</p>
</td>
</tr>
<tr id="row55726034515"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p17573120164511"><a name="p17573120164511"></a><a name="p17573120164511"></a>MS_ROLE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1457320184519"><a name="p1457320184519"></a><a name="p1457320184519"></a>指定本进程角色</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul9226735436"></a><a name="ul9226735436"></a><ul id="ul9226735436"><li>MS_SCHED：表示Scheduler进程，一个训练任务只启动一个Scheduler，负责组网，容器恢复等，<strong id="b18226143174315"><a name="b18226143174315"></a><a name="b18226143174315"></a>不会执行训练代码</strong>。</li><li>MS_WORKER：表示Worker进程，一般设置分布式训练进程为此角色。</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1457316017453"><a name="p1457316017453"></a><a name="p1457316017453"></a>Worker进程会向Scheduler进程注册从而完成组网。</p>
</td>
</tr>
<tr id="row9477165812440"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1747775874419"><a name="p1747775874419"></a><a name="p1747775874419"></a>MS_NODE_RANK</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p91701339918"><a name="p91701339918"></a><a name="p91701339918"></a>本节点Pod的Node Rank</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3871413144015"><a name="p3871413144015"></a><a name="p3871413144015"></a>大于或等于0的整数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p5312103318576"><a name="p5312103318576"></a><a name="p5312103318576"></a>Scheduler Pod设置为0。</p>
<a name="ul350115366586"></a><a name="ul350115366586"></a><ul id="ul350115366586"><li>当Scheduler挂载芯片时，Worker Pod从1开始递增。</li><li>当Scheduler不挂载芯片时，Worker Pod从0开始递增。</li></ul>
</td>
</tr>
<tr id="row736875617444"><td class="cellrowborder" rowspan="7" valign="top" width="12.379999999999999%" headers="mcps1.2.6.1.1 "><p id="p10368156114420"><a name="p10368156114420"></a><a name="p10368156114420"></a><span id="ph12195638125217"><a name="ph12195638125217"></a><a name="ph12195638125217"></a>TensorFlow</span></p>
<p id="p147091538135013"><a name="p147091538135013"></a><a name="p147091538135013"></a></p>
<p id="p1182154174413"><a name="p1182154174413"></a><a name="p1182154174413"></a></p>
<p id="p15496174405014"><a name="p15496174405014"></a><a name="p15496174405014"></a></p>
<p id="p16608736115011"><a name="p16608736115011"></a><a name="p16608736115011"></a></p>
<p id="p58121518449"><a name="p58121518449"></a><a name="p58121518449"></a></p>
<p id="p37734944410"><a name="p37734944410"></a><a name="p37734944410"></a></p>
<p id="p893002673510"><a name="p893002673510"></a><a name="p893002673510"></a></p>
</td>
<td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p8263121415503"><a name="p8263121415503"></a><a name="p8263121415503"></a>CM_CHIEF_IP</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p17368115654418"><a name="p17368115654418"></a><a name="p17368115654418"></a>与CHIEF通信的IP</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><p id="p79867231273"><a name="p79867231273"></a><a name="p79867231273"></a>合法的IP地址，格式为字符串，要求为常规IPv4或IPv6格式</p>
</td>
<td class="cellrowborder" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><a name="ul102323572366"></a><a name="ul102323572366"></a><ul id="ul102323572366"><li>chief Pod中设置为podIP。</li><li>Worker Pod设置为chief Pod对应svc的clusterIP。</li></ul>
</td>
</tr>
<tr id="row18709738135012"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p18709238115018"><a name="p18709238115018"></a><a name="p18709238115018"></a>CM_CHIEF_PORT</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p187091238165016"><a name="p187091238165016"></a><a name="p187091238165016"></a>与CHIEF通信的端口</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p6871111384013"><a name="p6871111384013"></a><a name="p6871111384013"></a>支持配置为字符串、数字，取值范围0~65520</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p2558810068"><a name="p2558810068"></a><a name="p2558810068"></a>Scheduler Pod对应svc中名称为ascendjob-port的值，默认取值为2222。</p>
</td>
</tr>
<tr id="row01811354154420"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1418215547442"><a name="p1418215547442"></a><a name="p1418215547442"></a>CM_CHIEF_DEVICE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p718285417445"><a name="p718285417445"></a><a name="p718285417445"></a>用于指定CHIEF节点中统计Server端集群信息的Device逻辑ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p4871121315402"><a name="p4871121315402"></a><a name="p4871121315402"></a>0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p13182195412441"><a name="p13182195412441"></a><a name="p13182195412441"></a>取值固定取值为0。</p>
</td>
</tr>
<tr id="row1949611447503"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p16496114485016"><a name="p16496114485016"></a><a name="p16496114485016"></a>CM_WORKER_SIZE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1149634410506"><a name="p1149634410506"></a><a name="p1149634410506"></a>任务使用的总NPU数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p140293415346"><a name="p140293415346"></a><a name="p140293415346"></a>取值范围为0~32768</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p14401120121415"><a name="p14401120121415"></a><a name="p14401120121415"></a>任务使用的总卡数，例如64个NPU任务，则取值为64。</p>
</td>
</tr>
<tr id="row18608103620502"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p860893611501"><a name="p860893611501"></a><a name="p860893611501"></a>CM_LOCAL_WORKER</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19608103665014"><a name="p19608103665014"></a><a name="p19608103665014"></a>每个Pod使用的NPU数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1987113139401"><a name="p1987113139401"></a><a name="p1987113139401"></a>大于0的整数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p182881124181417"><a name="p182881124181417"></a><a name="p182881124181417"></a>例如Pod使用4个NPU，则配置为4。</p>
</td>
</tr>
<tr id="row88121951194420"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p4812951174410"><a name="p4812951174410"></a><a name="p4812951174410"></a>CM_WORKER_IP</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1790162214197"><a name="p1790162214197"></a><a name="p1790162214197"></a>Pod的podIP</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p124691910181618"><a name="p124691910181618"></a><a name="p124691910181618"></a>合法的IP地址，格式为字符串，要求为常规IPv4或IPv6格式</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p98128515444"><a name="p98128515444"></a><a name="p98128515444"></a>当前Pod的podIP。</p>
</td>
</tr>
<tr id="row177104964413"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1477164984418"><a name="p1477164984418"></a><a name="p1477164984418"></a>CM_RANK</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p137794919443"><a name="p137794919443"></a><a name="p137794919443"></a>本节点Pod的Node Rank</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1871713194015"><a name="p1871713194015"></a><a name="p1871713194015"></a>大于或等于0的整数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><a name="ul2345519387"></a><a name="ul2345519387"></a><ul id="ul2345519387"><li>chief设置为0</li><li>worker从1开始递增</li></ul>
</td>
</tr>
<tr id="row1058205923118"><td class="cellrowborder" valign="top" width="12.379999999999999%" headers="mcps1.2.6.1.1 "><p id="p1181155914329"><a name="p1181155914329"></a><a name="p1181155914329"></a><span id="ph1551815244211"><a name="ph1551815244211"></a><a name="ph1551815244211"></a>PyTorch</span>、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p318119595325"><a name="p318119595325"></a><a name="p318119595325"></a>PROCESS_RECOVER</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p1318195912322"><a name="p1318195912322"></a><a name="p1318195912322"></a>进程级别重调度、进程级在线恢复及弹性训练总开关。</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><a name="ul65501334344"></a><a name="ul65501334344"></a><ul id="ul65501334344"><li>on：开启本功能。</li><li>off：关闭本功能。</li></ul>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p682072453313"><a name="p682072453313"></a><a name="p682072453313"></a>进程级别重调度、进程级在线恢复、进程级原地恢复和弹性训练场景下注入该环境变量。</p>
</td>
</tr>
<tr id="row242413586587"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p311314413594"><a name="p311314413594"></a><a name="p311314413594"></a><span id="ph611313425919"><a name="ph611313425919"></a><a name="ph611313425919"></a>PyTorch</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p151133465910"><a name="p151133465910"></a><a name="p151133465910"></a>HIGH_AVAILABILITY</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1911315414598"><a name="p1911315414598"></a><a name="p1911315414598"></a>MindSpeed-LLM进程级恢复功能开关。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p61131485917"><a name="p61131485917"></a><a name="p61131485917"></a>任务可用恢复策略。</p>
<a name="ul2113204145911"></a><a name="ul2113204145911"></a><ul id="ul2113204145911"><li>retry：进程级在线恢复。</li><li>recover：进程级别重调度。</li><li>dump：保存临终遗言。</li><li>elastic-training：弹性训练。</li></ul>
</td>
</tr>
<tr id="row83631024143218"><td class="cellrowborder" valign="top" width="12.379999999999999%" headers="mcps1.2.6.1.1 "><p id="p1018165914322"><a name="p1018165914322"></a><a name="p1018165914322"></a><span id="ph12546182614219"><a name="ph12546182614219"></a><a name="ph12546182614219"></a>PyTorch</span>、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" width="16.869999999999997%" headers="mcps1.2.6.1.2 "><p id="p118117596323"><a name="p118117596323"></a><a name="p118117596323"></a>ELASTIC_PROCESS_RECOVER_ENABLE</p>
</td>
<td class="cellrowborder" valign="top" width="27.77%" headers="mcps1.2.6.1.3 "><p id="p1518175973216"><a name="p1518175973216"></a><a name="p1518175973216"></a><span id="ph1072282311518"><a name="ph1072282311518"></a><a name="ph1072282311518"></a>Elastic Agent</span>侧进程级别重调度、进程级在线恢复、临终CKPT恢复功能开关。</p>
</td>
<td class="cellrowborder" valign="top" width="19.79%" headers="mcps1.2.6.1.4 "><a name="ul167945693511"></a><a name="ul167945693511"></a><ul id="ul167945693511"><li>取值为1：开启本功能。</li><li>取值为其他值：关闭本功能。关闭本功能时，MindIO侧相关功能需同时关闭。</li></ul>
</td>
<td class="cellrowborder" rowspan="4" valign="top" width="23.189999999999998%" headers="mcps1.2.6.1.5 "><p id="p923972485920"><a name="p923972485920"></a><a name="p923972485920"></a>进程级别重调度、进程级在线恢复、进程级原地恢复场景下注入该环境变量。</p>
</td>
</tr>
<tr id="row0765193853210"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1918295993216"><a name="p1918295993216"></a><a name="p1918295993216"></a><span id="ph5168103016219"><a name="ph5168103016219"></a><a name="ph5168103016219"></a>PyTorch</span>、MindSpore</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17182145910321"><a name="p17182145910321"></a><a name="p17182145910321"></a>ENABLE_RESTART_FAULT_PROCESS</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p71821759193216"><a name="p71821759193216"></a><a name="p71821759193216"></a><span id="ph249610307518"><a name="ph249610307518"></a><a name="ph249610307518"></a>Elastic Agent</span>/<span id="ph1513354715617"><a name="ph1513354715617"></a><a name="ph1513354715617"></a>TaskD</span>组件开启故障进程级原地恢复功能开关。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><a name="ul1032320399361"></a><a name="ul1032320399361"></a><ul id="ul1032320399361"><li>on：开启本功能。</li><li>其他值：关闭本功能。</li></ul>
<div class="note" id="note21949542365"><a name="note21949542365"></a><a name="note21949542365"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul11833105863616"></a><a name="ul11833105863616"></a><ul id="ul11833105863616"><li><span id="ph1982841320618"><a name="ph1982841320618"></a><a name="ph1982841320618"></a>PyTorch</span>框架下，本功能由<span id="ph193151321661"><a name="ph193151321661"></a><a name="ph193151321661"></a>Elastic Agent/TaskD</span>提供。</li><li>MindSpore框架下，本功能由<span id="ph5518105017616"><a name="ph5518105017616"></a><a name="ph5518105017616"></a>TaskD</span>提供。</li></ul>
</div></div>
</td>
</tr>
<tr id="row1276511386323"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p618295953218"><a name="p618295953218"></a><a name="p618295953218"></a>MindSpore</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p518245917328"><a name="p518245917328"></a><a name="p518245917328"></a>MINDIO_FOR_MINDSPORE</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p818275973217"><a name="p818275973217"></a><a name="p818275973217"></a>MindIO使能MindSpore开关。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1418212592323"><a name="p1418212592323"></a><a name="p1418212592323"></a>取值为1：开启MindIO使能MindSpore开关。</p>
</td>
</tr>
<tr id="row10116337329"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p14182125983213"><a name="p14182125983213"></a><a name="p14182125983213"></a>MindSpore</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1518285953219"><a name="p1518285953219"></a><a name="p1518285953219"></a>MS_ENABLE_TFT</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p18182105963210"><a name="p18182105963210"></a><a name="p18182105963210"></a>MindSpore使能进程级恢复开关。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><pre class="screen" id="screen15182185953212"><a name="screen15182185953212"></a><a name="screen15182185953212"></a>'{TTP:1,UCE:1,ARF:1,HCCE:1,RSC:1}'     # 分别开启临终遗言、<span id="ph15161018131912"><a name="ph15161018131912"></a><a name="ph15161018131912"></a>片上内存</span>故障的进程级在线恢复、进程级别重调度、网络故障的进程级在线恢复和Pod级别重调度</pre>
</td>
</tr>
</tbody>
</table>

**Ascend Docker Runtime环境变量说明<a name="section109964810209"></a>**

Ascend Docker Runtime为容器注入相应的环境变量。

<a name="table974781182117"></a>

|环境变量名称|功能|取值|说明|
|--|--|--|--|
|ASCEND_DOCKER_RUNTIME|标识当前环境是否安装了Ascend Docker Runtime插件。|True|当未安装Ascend Docker Runtime时不存在该环境变量。|

**Ascend Device Plugin环境变量说明<a name="section1419516175219"></a>**

Ascend Device Plugin为容器注入相应的环境变量，该环境变量的相关说明请参见下表。

**表 3**  Ascend Device Plugin向容器中注入的环境变量

<a name="table4446195872218"></a>

|环境变量名称|功能|取值|说明|
|--|--|--|--|
|ASCEND_VISIBLE_DEVICES|如果任务需要使用NPU设备，必须使用ASCEND_VISIBLE_DEVICES指定被挂载至容器中的NPU设备，否则挂载NPU设备失败。使用设备序号指定设备时，支持单个和范围指定且支持混用；使用芯片名称指定设备时，支持同时指定多个同类型的芯片名称。|<ul><li>挂载物理芯片（NPU）<ul><li>ASCEND_VISIBLE_DEVICES=0表示将0号NPU设备（/dev/davinci0）挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=1,3表示将1号和3号NPU设备挂载入容器中。</li></ul></li><li>挂载虚拟芯片（vNPU）</li><ul><li>**静态虚拟化**：和物理芯片使用方式相同，只需要把物理芯片ID换成虚拟芯片ID（vNPU ID）即可。</li><li>**动态虚拟化**：ASCEND_VISIBLE_DEVICES=0表示从0号NPU设备中划分出一定数量的AI Core。</li></ul></ul>|-|
|ASCEND_ALLOW_LINK|是否允许挂载的文件或目录中存在软链接，在Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件场景下必须指定该参数。|<ul><li>ASCEND_ALLOW_LINK=True，表示在Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件场景下允许挂载带有软链接的驱动文件。</li><li>ASCEND_ALLOW_LINK=False或者不指定该参数，Atlas 500 A2 智能小站、Atlas 200I A2 加速模块和Atlas 200I DK A2 开发者套件将无法使用Ascend Docker Runtime。</li></ul>|-|
|ASCEND_RUNTIME_OPTIONS|对参数ASCEND_VISIBLE_DEVICES中指定的芯片ID作出限制：<ul><li>NODRV：表示不挂载驱动相关目录。</li><li>VIRTUAL：表示挂载的是虚拟芯片。</li><li>NODRV,VIRTUAL：表示挂载的是虚拟芯片，并且不挂载驱动相关目录。</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul>|-|
|WORLD_SIZE|任务使用的总NPU数|大于或等于0的整数|仅在动态vNPU调度场景下写入|
|LOCAL_WORLD_SIZE|每个节点Pod使用的NPU数|大于或等于0的整数|仅在动态vNPU调度场景下写入|
|LOCAL_RANK|每个节点Pod使用的NPU的逻辑ID列表|字符串|仅在动态vNPU调度场景下写入。从0开始。例如，Pod使用4个NPU，则配置为{0,1,2,3}。|
|CM_WORKER_SIZE|任务使用的总NPU数|大于或等于0的整数|仅在动态vNPU调度场景下写入|
|CM_LOCAL_WORKER|每个节点Pod使用的NPU数|大于或等于0的整数|仅在动态vNPU调度场景下写入|
|MS_WORKER_NUM|任务使用的总NPU数|大于或等于0的整数|仅在动态vNPU调度场景下写入|
|MS_LOCAL_WORKER|每个节点Pod使用的NPU数|大于或等于0的整数|仅在动态vNPU调度场景下写入|
|PERF_DUMP_PATH|迭代时延和分组信息保存路径|字符串|仅在慢节点检测场景下写入|
|PERF_DUMP_CONFIG|迭代时延和分组信息启停开关|字符串|仅在慢节点检测场景下写入|
|KUBELET_PORT|指定当前节点kubelet默认端口号（若用户未自定义kubelet端口，则无需配置）。|0~65535的整数|若用户修改kubelet默认端口，需要设置该环境变量的值为自定义端口号。<p>若用户未修改kubelet默认端口，则忽略该环境变量。</p>|
|HOST_IP|指定当前节点的物理IP。|合法的IP地址，格式为字符串，要求为常规IPv4或IPv6格式|固定配置项，初始YAML文件已提供。|

**Elastic Agent环境变量说明<a name="section8853192413411"></a>**

>[!NOTE] 
>Elastic Agent组件已经日落，相关资料将于2026年12月30日的版本删除。

使用Elastic Agent组件时可以配置的环境变量。如需了解其他来自源码的环境变量，请参见[PyTorch相关资料](https://pytorch.ac.cn/#google_vignette)。

**表 4** Elastic Agent环境变量说明

<a name="table159711045543"></a>
<table><thead align="left"><tr id="row109717454411"><th class="cellrowborder" valign="top" width="19.25%" id="mcps1.2.5.1.1"><p id="p1897164513415"><a name="p1897164513415"></a><a name="p1897164513415"></a>环境变量名称</p>
</th>
<th class="cellrowborder" valign="top" width="20.72%" id="mcps1.2.5.1.2"><p id="p49711245944"><a name="p49711245944"></a><a name="p49711245944"></a>功能</p>
</th>
<th class="cellrowborder" valign="top" width="14.19%" id="mcps1.2.5.1.3"><p id="p119716457417"><a name="p119716457417"></a><a name="p119716457417"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="45.839999999999996%" id="mcps1.2.5.1.4"><p id="p797164512417"><a name="p797164512417"></a><a name="p797164512417"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row1097110458417"><td class="cellrowborder" valign="top" width="19.25%" headers="mcps1.2.5.1.1 "><p id="p16501154910413"><a name="p16501154910413"></a><a name="p16501154910413"></a>ELASTIC_LOG_PATH</p>
</td>
<td class="cellrowborder" valign="top" width="20.72%" headers="mcps1.2.5.1.2 "><p id="p174982492048"><a name="p174982492048"></a><a name="p174982492048"></a><span id="ph1272719340716"><a name="ph1272719340716"></a><a name="ph1272719340716"></a>Elastic Agent</span>组件运行日志的落盘路径。</p>
</td>
<td class="cellrowborder" valign="top" width="14.19%" headers="mcps1.2.5.1.3 "><p id="p19494949546"><a name="p19494949546"></a><a name="p19494949546"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="45.839999999999996%" headers="mcps1.2.5.1.4 "><p id="p1249154918417"><a name="p1249154918417"></a><a name="p1249154918417"></a>配置时需区分该日志的节点名称。参考示例：</p>
<pre class="screen" id="screen1333264631214"><a name="screen1333264631214"></a><a name="screen1333264631214"></a>ELASTIC_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/elasticlogs/elastic-log$XDL_IP-$RANK 
请将<strong id="b1623882751713"><a name="b1623882751713"></a><a name="b1623882751713"></a>$XDL_IP</strong>替换成实际使用的节点IP。
请将<strong id="b016573418154"><a name="b016573418154"></a><a name="b016573418154"></a>$RANK</strong>替换成实际使用的节点RANK。</pre>
</td>
</tr>
<tr id="row397995514414"><td class="cellrowborder" valign="top" width="19.25%" headers="mcps1.2.5.1.1 "><p id="p16980185594110"><a name="p16980185594110"></a><a name="p16980185594110"></a>ELASTIC_PROCESS_RECOVER_ENABLE</p>
</td>
<td class="cellrowborder" valign="top" width="20.72%" headers="mcps1.2.5.1.2 "><p id="p1098015594113"><a name="p1098015594113"></a><a name="p1098015594113"></a><span id="ph1764143514215"><a name="ph1764143514215"></a><a name="ph1764143514215"></a>Elastic Agent</span>侧进程级别重调度、进程级在线恢复、临终CheckPoint恢复功能开关。</p>
</td>
<td class="cellrowborder" valign="top" width="14.19%" headers="mcps1.2.5.1.3 "><p id="p09801355144111"><a name="p09801355144111"></a><a name="p09801355144111"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="45.839999999999996%" headers="mcps1.2.5.1.4 "><a name="ul1436916744312"></a><a name="ul1436916744312"></a><ul id="ul1436916744312"><li>取值为1：开启本功能。</li><li>其他值：关闭本功能。<p id="p121131599216"><a name="p121131599216"></a><a name="p121131599216"></a>关闭本功能时，MindIO侧相关功能需同时关闭。</p>
<div class="note" id="note2086611177509"><a name="note2086611177509"></a><div class="notebody"><p id="p1478116510530"><a name="p1478116510530"></a><a name="p1478116510530"></a>MindIO侧相关功能开关说明如下：</p>
<a name="ul208627229515"></a><a name="ul208627229515"></a><ul id="ul208627229515"><li>enable-high-availability：故障快速恢复特性开关，默认关闭，配置后即开启临终遗言功能。</li><li>enable-worker-reboot：进程级别重调度功能开关，默认关闭，配置后在发生一般性故障时，进行进程级别调度，继续训练。本开关开启时，需同时开启enable-high-availability。</li></ul>
</div></div>
</li></ul>
</td>
</tr>
<tr id="row0615134982716"><td class="cellrowborder" valign="top" width="19.25%" headers="mcps1.2.5.1.1 "><p id="p8615124913271"><a name="p8615124913271"></a><a name="p8615124913271"></a>ENABLE_RESTART_FAULT_PROCESS</p>
</td>
<td class="cellrowborder" valign="top" width="20.72%" headers="mcps1.2.5.1.2 "><p id="p8615749132716"><a name="p8615749132716"></a><a name="p8615749132716"></a><span id="ph97981526102819"><a name="ph97981526102819"></a><a name="ph97981526102819"></a>Elastic Agent</span>组件开启进程级别原地恢复功能的开关。</p>
</td>
<td class="cellrowborder" valign="top" width="14.19%" headers="mcps1.2.5.1.3 "><p id="p1761620496274"><a name="p1761620496274"></a><a name="p1761620496274"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="45.839999999999996%" headers="mcps1.2.5.1.4 "><p id="p20864131183019"><a name="p20864131183019"></a><a name="p20864131183019"></a>取值为“on”或“其他值”。</p>
<a name="ul0406214203015"></a><a name="ul0406214203015"></a><ul id="ul0406214203015"><li>on：开启本功能</li><li>其他值：关闭本功能</li></ul>
</td>
</tr>
<tr id="row9124131755016"><td class="cellrowborder" valign="top" width="19.25%" headers="mcps1.2.5.1.1 "><p id="p12124161735013"><a name="p12124161735013"></a><a name="p12124161735013"></a>RESTART_FAULT_PROCESS_TYPE</p>
</td>
<td class="cellrowborder" valign="top" width="20.72%" headers="mcps1.2.5.1.2 "><p id="p5124191714502"><a name="p5124191714502"></a><a name="p5124191714502"></a><span id="ph17327438155018"><a name="ph17327438155018"></a><a name="ph17327438155018"></a>Elastic Agent</span>通知MindIO重启故障进程的类型。</p>
</td>
<td class="cellrowborder" valign="top" width="14.19%" headers="mcps1.2.5.1.3 "><p id="p1412431745018"><a name="p1412431745018"></a><a name="p1412431745018"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="45.839999999999996%" headers="mcps1.2.5.1.4 "><p id="p2027133085115"><a name="p2027133085115"></a><a name="p2027133085115"></a>取值为“worker”或“pod”。</p>
<a name="ul125707417516"></a><a name="ul125707417516"></a><ul id="ul125707417516"><li>worker：不退出Pod，只重启故障进程</li><li>pod：重启Pod</li></ul>
</td>
</tr>
<tr id="row1720115818520"><td class="cellrowborder" valign="top" width="19.25%" headers="mcps1.2.5.1.1 "><p id="p18202188165214"><a name="p18202188165214"></a><a name="p18202188165214"></a>RANK_TABLE_FILE</p>
</td>
<td class="cellrowborder" valign="top" width="20.72%" headers="mcps1.2.5.1.2 "><p id="p320288195211"><a name="p320288195211"></a><a name="p320288195211"></a>RankTable文件路径</p>
</td>
<td class="cellrowborder" valign="top" width="14.19%" headers="mcps1.2.5.1.3 "><p id="p152028835217"><a name="p152028835217"></a><a name="p152028835217"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="45.839999999999996%" headers="mcps1.2.5.1.4 "><p id="p13202589520"><a name="p13202589520"></a><a name="p13202589520"></a>hccl.json文件的路径</p>
</td>
</tr>
<tr id="row1223545210523"><td class="cellrowborder" valign="top" width="19.25%" headers="mcps1.2.5.1.1 "><p id="p4235135275219"><a name="p4235135275219"></a><a name="p4235135275219"></a>PROCESS_RECOVER</p>
</td>
<td class="cellrowborder" valign="top" width="20.72%" headers="mcps1.2.5.1.2 "><p id="p72360528528"><a name="p72360528528"></a><a name="p72360528528"></a>进程级别重调度或进程级在线恢复开关</p>
</td>
<td class="cellrowborder" valign="top" width="14.19%" headers="mcps1.2.5.1.3 "><p id="p18236165285211"><a name="p18236165285211"></a><a name="p18236165285211"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="45.839999999999996%" headers="mcps1.2.5.1.4 "><p id="p4482132175410"><a name="p4482132175410"></a><a name="p4482132175410"></a>取值为“on”或“其他值”。</p>
<a name="ul4482132110546"></a><a name="ul4482132110546"></a><ul id="ul4482132110546"><li>on：开启本功能</li><li>其他值：关闭本功能</li></ul>
</td>
</tr>
</tbody>
</table>

**TaskD环境变量说明<a name="section6616275583"></a>**

使用TaskD组件时可以配置的环境变量。如需了解其他来自源码的环境变量，请参见[PyTorch相关资料](https://pytorch.ac.cn/#google_vignette)。

**表 5** TaskD环境变量说明

<a name="table13568156155815"></a>

|环境变量名称|功能|取值|说明|
|--|--|--|--|
|TASKD_LOG_PATH|指定TaskD组件运行日志的落盘路径。|字符串|如未指定使用默认的./taskd_log/taskd.log-worker-{*RANK*}，即当前执行路径下的taskd_log目录。<p>*{RANK}*为当前训练进程的全局rank号。</p>|
|TASKD_FILE_LOG_LEVEL|指定需要记录到日志文件的日志等级。|字符串|-|
|TASKD_STD_LOG_LEVEL|指定需要打印的日志等级。|字符串|-|
|TASKD_LOG_STDOUT|指定日志是否需要打印。|bool|取值为True或False。|
|ENABLE_RESTART_FAULT_PROCESS|TaskD组件开启进程级别原地恢复功能的开关。|字符串|取值为“on”或“其他值”。<ul><li>on：开启本功能</li><li>其他值：关闭本功能</li></ul>|
|RESTART_FAULT_PROCESS_TYPE|TaskD通知MindIO重启故障进程的类型。|字符串|取值为“worker”或“pod”。<ul><li>worker：不退出Pod，只重启故障进程</li><li>pod：重启Pod</li></ul>|
|TASKD_PROCESS_ENABLE|TaskD组件开启进程级别重调度、进程级在线恢复、进程级别原地恢复和弹性训练功能的开关。|字符串|取值为“on”或“off”。<ul><li>on：开启本功能</li><li>off：关闭本功能</li></ul>|
|LOCAL_PROXY_ENABLE|是否开启本地代理（安全加固需要）。|字符串|取值为“on”或“off”。<ul><li>on：开启本功能</li><li>off：关闭本功能</li></ul>默认值为“off”，通信安全加固场景需要设置为“on”。|
|HCCL_ASYNC_ERROR_HANDLING|是否开启watchdog功能。|字符串|取值如下：<ul><li>0：表示关闭故障检测和进程退出功能。</li><li>1：表示开启故障检测和进程退出功能。</li><li>2：表示仅开启故障检测功能。</li></ul>默认值为1。|
|TASKD_PROCESS_INTERVAL|设置TaskD Manager主流程处理间隔。|字符串|取值范围为100~1000，单位为毫秒。|

**NodeD环境变量说明<a name="section10131935141216"></a>**

**表 6** NodeD环境变量说明

<a name="table11131133571214"></a>

|环境变量名称|功能|取值|说明|
|--|--|--|--|
|XDL_IP|用于获取Pod所在host的IP地址，慢节点使用，用于记录、匹配慢节点信息。|合法的IP地址，格式为字符串，要求为常规IPv4或IPv6格式。|部署NodeD组件的YAML中写入该环境变量。|
