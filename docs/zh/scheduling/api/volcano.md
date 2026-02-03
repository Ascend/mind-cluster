# Volcano<a name="ZH-CN_TOPIC_0000002479226814"></a>

## 获取集群调度组件信息<a name="ZH-CN_TOPIC_0000002479386860"></a>

-   VolcanoJob接口由开源组件Volcano提供，MindCluster修改了VolcanoJob接口的Annotations字段，如[表1](#table177621954014)所示。其他接口未改动，了解开源Volcano的详细说明请参见Volcano开源社区。

    **表 1**  Annotations参数说明

    <a name="table177621954014"></a>
    <table><thead align="left"><tr id="row1577601924020"><th class="cellrowborder" valign="top" width="17.85178517851785%" id="mcps1.2.4.1.1"><p id="p5776119134011"><a name="p5776119134011"></a><a name="p5776119134011"></a>参数名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="66.05660566056606%" id="mcps1.2.4.1.2"><p id="p16776119164019"><a name="p16776119164019"></a><a name="p16776119164019"></a>说明</p>
    </th>
    <th class="cellrowborder" valign="top" width="16.09160916091609%" id="mcps1.2.4.1.3"><p id="p18776121934012"><a name="p18776121934012"></a><a name="p18776121934012"></a>取值</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row4776111944019"><td class="cellrowborder" valign="top" width="17.85178517851785%" headers="mcps1.2.4.1.1 "><p id="p167761919104014"><a name="p167761919104014"></a><a name="p167761919104014"></a>distributed</p>
    </td>
    <td class="cellrowborder" valign="top" width="66.05660566056606%" headers="mcps1.2.4.1.2 "><p id="p1429543818494"><a name="p1429543818494"></a><a name="p1429543818494"></a>由<span id="ph829115811272"><a name="ph829115811272"></a><a name="ph829115811272"></a>Resilience Controller</span>写入和使用，标记job是否为分布式任务。</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.09160916091609%" headers="mcps1.2.4.1.3 "><p id="p11776101954011"><a name="p11776101954011"></a><a name="p11776101954011"></a>True</p>
    </td>
    </tr>
    </tbody>
    </table>

-   对于volcano-scheduler和volcano-controller组件Pod开放的接口（开源组件本身定义），做出如下说明。

    **表 2** 集群调度Volcano组件开放接口列表

    <a name="zh-cn_topic_0000001446965056_table173071368477"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001446965056_row153077618473"><th class="cellrowborder" valign="top" width="34.68346834683469%" id="mcps1.2.6.1.1"><p id="zh-cn_topic_0000001446965056_p3307116134715"><a name="zh-cn_topic_0000001446965056_p3307116134715"></a><a name="zh-cn_topic_0000001446965056_p3307116134715"></a>访问方式</p>
    </th>
    <th class="cellrowborder" valign="top" width="6.0906090609060906%" id="mcps1.2.6.1.2"><p id="zh-cn_topic_0000001446965056_p1525244211493"><a name="zh-cn_topic_0000001446965056_p1525244211493"></a><a name="zh-cn_topic_0000001446965056_p1525244211493"></a>协议</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.741174117411742%" id="mcps1.2.6.1.3"><p id="zh-cn_topic_0000001446965056_p04543391867"><a name="zh-cn_topic_0000001446965056_p04543391867"></a><a name="zh-cn_topic_0000001446965056_p04543391867"></a>方法</p>
    </th>
    <th class="cellrowborder" valign="top" width="16.89168916891689%" id="mcps1.2.6.1.4"><p id="zh-cn_topic_0000001446965056_p23071468473"><a name="zh-cn_topic_0000001446965056_p23071468473"></a><a name="zh-cn_topic_0000001446965056_p23071468473"></a>作用</p>
    </th>
    <th class="cellrowborder" valign="top" width="30.59305930593059%" id="mcps1.2.6.1.5"><p id="zh-cn_topic_0000001446965056_p730796134713"><a name="zh-cn_topic_0000001446965056_p730796134713"></a><a name="zh-cn_topic_0000001446965056_p730796134713"></a>所属组件</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001446965056_row23070613479"><td class="cellrowborder" valign="top" width="34.68346834683469%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001446965056_p1730717615477"><a name="zh-cn_topic_0000001446965056_p1730717615477"></a><a name="zh-cn_topic_0000001446965056_p1730717615477"></a>http://podIP:11251/healthz</p>
    </td>
    <td class="cellrowborder" valign="top" width="6.0906090609060906%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001446965056_p10252142154917"><a name="zh-cn_topic_0000001446965056_p10252142154917"></a><a name="zh-cn_topic_0000001446965056_p10252142154917"></a>http</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.741174117411742%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001446965056_p64546391612"><a name="zh-cn_topic_0000001446965056_p64546391612"></a><a name="zh-cn_topic_0000001446965056_p64546391612"></a>Get</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p151727316530"><a name="zh-cn_topic_0000001446965056_p151727316530"></a><a name="zh-cn_topic_0000001446965056_p151727316530"></a>健康检查端口</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.59305930593059%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001446965056_p103089610475"><a name="zh-cn_topic_0000001446965056_p103089610475"></a><a name="zh-cn_topic_0000001446965056_p103089610475"></a>volcano-controller</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001446965056_row1308176144715"><td class="cellrowborder" valign="top" width="34.68346834683469%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001446965056_p2308865475"><a name="zh-cn_topic_0000001446965056_p2308865475"></a><a name="zh-cn_topic_0000001446965056_p2308865475"></a>http://podIP:11251/healthz</p>
    </td>
    <td class="cellrowborder" valign="top" width="6.0906090609060906%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001446965056_p162523428499"><a name="zh-cn_topic_0000001446965056_p162523428499"></a><a name="zh-cn_topic_0000001446965056_p162523428499"></a>http</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.741174117411742%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001446965056_p245433916610"><a name="zh-cn_topic_0000001446965056_p245433916610"></a><a name="zh-cn_topic_0000001446965056_p245433916610"></a>Get</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p53084617475"><a name="zh-cn_topic_0000001446965056_p53084617475"></a><a name="zh-cn_topic_0000001446965056_p53084617475"></a>健康检查端口</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.59305930593059%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001446965056_p14308176104718"><a name="zh-cn_topic_0000001446965056_p14308176104718"></a><a name="zh-cn_topic_0000001446965056_p14308176104718"></a>volcano-scheduler</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001446965056_row830812614472"><td class="cellrowborder" valign="top" width="34.68346834683469%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001446965056_p19308116104714"><a name="zh-cn_topic_0000001446965056_p19308116104714"></a><a name="zh-cn_topic_0000001446965056_p19308116104714"></a>http://volcano-scheduler-serviceIP:8080/metrics</p>
    </td>
    <td class="cellrowborder" valign="top" width="6.0906090609060906%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001446965056_p10252104224912"><a name="zh-cn_topic_0000001446965056_p10252104224912"></a><a name="zh-cn_topic_0000001446965056_p10252104224912"></a>http</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.741174117411742%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001446965056_p134543391765"><a name="zh-cn_topic_0000001446965056_p134543391765"></a><a name="zh-cn_topic_0000001446965056_p134543391765"></a>Get</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p193087624718"><a name="zh-cn_topic_0000001446965056_p193087624718"></a><a name="zh-cn_topic_0000001446965056_p193087624718"></a>Prometheus信息收集端口</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.59305930593059%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001446965056_p3308166154716"><a name="zh-cn_topic_0000001446965056_p3308166154716"></a><a name="zh-cn_topic_0000001446965056_p3308166154716"></a>volcano-scheduler</p>
    </td>
    </tr>
    <tr id="row92108376505"><td class="cellrowborder" colspan="5" valign="top" headers="mcps1.2.6.1.1 mcps1.2.6.1.2 mcps1.2.6.1.3 mcps1.2.6.1.4 mcps1.2.6.1.5 "><p id="p57959426507"><a name="p57959426507"></a><a name="p57959426507"></a>注：</p>
    <p id="p142394495508"><a name="p142394495508"></a><a name="p142394495508"></a>为保证Volcano健康检查端口和Prometheus信息收集端口的正常访问，请在安装<span id="ph175881448132716"><a name="ph175881448132716"></a><a name="ph175881448132716"></a>Volcano</span>时，将YAML中的--enable-healthz参数和--enable-metrics参数的值设置为“true”，详细修改方法可参见<a href="../installation_guide.md#安装volcano">步骤7</a>。</p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE] 说明 
    >华为云的CCI服务提供了更为详细的VolcanoJob说明，可参见《云容器实例 API参考》中“[创建Volcano Job](https://support.huaweicloud.com/api-cci/createBatchVolcanoShV1alpha1NamespacedJob.html)”章节了解相关内容。


## PodGroup<a name="ZH-CN_TOPIC_0000002479226832"></a>

**表 1** 集群调度组件对PodGroup label使用说明

<a name="table143562050699"></a>
<table><thead align="left"><tr id="row23564507918"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p1535615011914"><a name="p1535615011914"></a><a name="p1535615011914"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.2"><p id="p83576501093"><a name="p83576501093"></a><a name="p83576501093"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="p235719501097"><a name="p235719501097"></a><a name="p235719501097"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p435716507913"><a name="p435716507913"></a><a name="p435716507913"></a>使用组件</p>
</th>
</tr>
</thead>
<tbody><tr id="row7357125010917"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0357155019917"><a name="p0357155019917"></a><a name="p0357155019917"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p1035711508917"><a name="p1035711508917"></a><a name="p1035711508917"></a>标识Atlas的<span id="ph13571150291"><a name="ph13571150291"></a><a name="ph13571150291"></a>Pod</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul835765011916"></a><a name="ul835765011916"></a><ul id="ul835765011916"><li>ascend-910</li><li>ascend-<span id="ph19358150597"><a name="ph19358150597"></a><a name="ph19358150597"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li><li>huawei.com/npu</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10471193985417"><a name="p10471193985417"></a><a name="p10471193985417"></a><span id="ph1035865018915"><a name="ph1035865018915"></a><a name="ph1035865018915"></a>Ascend Device Plugin</span>、<span id="ph446593975417"><a name="ph446593975417"></a><a name="ph446593975417"></a>Ascend Operator</span>、<span id="ph635885012911"><a name="ph635885012911"></a><a name="ph635885012911"></a>Volcano</span></p>
</td>
</tr>
<tr id="row135825013910"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p203581150493"><a name="p203581150493"></a><a name="p203581150493"></a>fault-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p935815502914"><a name="p935815502914"></a><a name="p935815502914"></a>任务故障重调度开关</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p5358950291"><a name="p5358950291"></a><a name="p5358950291"></a>grace、force、off</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10358105012913"><a name="p10358105012913"></a><a name="p10358105012913"></a><span id="ph635812501497"><a name="ph635812501497"></a><a name="ph635812501497"></a>Volcano</span>、<span id="ph183581350898"><a name="ph183581350898"></a><a name="ph183581350898"></a>Resilience Controller</span></p>
</td>
</tr>
<tr id="row03591501297"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p143599501394"><a name="p143599501394"></a><a name="p143599501394"></a>elastic-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p1235917508916"><a name="p1235917508916"></a><a name="p1235917508916"></a>任务弹性调度开关</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p1135910503918"><a name="p1135910503918"></a><a name="p1135910503918"></a>on</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1735916501991"><a name="p1735916501991"></a><a name="p1735916501991"></a><span id="ph93613501992"><a name="ph93613501992"></a><a name="ph93613501992"></a>Resilience Controller</span>、<span id="ph53614501198"><a name="ph53614501198"></a><a name="ph53614501198"></a>Volcano</span></p>
</td>
</tr>
<tr id="row103614504912"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p6361950695"><a name="p6361950695"></a><a name="p6361950695"></a>fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p183613501791"><a name="p183613501791"></a><a name="p183613501791"></a>任务发生业务面故障可以重调度的次数</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p1836115501693"><a name="p1836115501693"></a><a name="p1836115501693"></a>0-100</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p636212501391"><a name="p636212501391"></a><a name="p636212501391"></a><span id="ph1036213501896"><a name="ph1036213501896"></a><a name="ph1036213501896"></a>Volcano</span>、<span id="ph436285014913"><a name="ph436285014913"></a><a name="ph436285014913"></a>Ascend Operator</span></p>
</td>
</tr>
<tr id="row1336212502091"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p103625502094"><a name="p103625502094"></a><a name="p103625502094"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p43628508911"><a name="p43628508911"></a><a name="p43628508911"></a>交换机亲和性策略</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul143629507913"></a><a name="ul143629507913"></a><ul id="ul143629507913"><li>normal-schema</li><li>large-model-schema</li><li>null</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p436310506914"><a name="p436310506914"></a><a name="p436310506914"></a><span id="ph73631050690"><a name="ph73631050690"></a><a name="ph73631050690"></a>Volcano</span></p>
</td>
</tr>
<tr id="row1136411501898"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0364135010913"><a name="p0364135010913"></a><a name="p0364135010913"></a>npu-310-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p636425010918"><a name="p636425010918"></a><a name="p636425010918"></a>标记推理服务器（插<span id="ph1436410501390"><a name="ph1436410501390"></a><a name="ph1436410501390"></a>Atlas 300I 推理卡</span>）调度策略</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul73644501797"></a><a name="ul73644501797"></a><ul id="ul73644501797"><li>card</li><li>chip</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1636419501398"><a name="p1636419501398"></a><a name="p1636419501398"></a><span id="ph163652501393"><a name="ph163652501393"></a><a name="ph163652501393"></a>Volcano</span></p>
</td>
</tr>
<tr id="row7970125593620"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p65914912716"><a name="p65914912716"></a><a name="p65914912716"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p3591694276"><a name="p3591694276"></a><a name="p3591694276"></a>是否启用Pod级别重调度。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul186101614131"></a><a name="ul186101614131"></a><ul id="ul186101614131"><li>on：开启Pod级别重调度</li><li>其他值或不使用该字段：关闭Pod级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1372045172812"><a name="p1372045172812"></a><a name="p1372045172812"></a><span id="ph2072005192818"><a name="ph2072005192818"></a><a name="ph2072005192818"></a>Volcano</span></p>
</td>
</tr>
<tr id="row209101813153710"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p2417162275410"><a name="p2417162275410"></a><a name="p2417162275410"></a>process-recover-enable</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p131172319273"><a name="p131172319273"></a><a name="p131172319273"></a>是否启用进程级别重调度。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul71592205015"></a><a name="ul71592205015"></a><ul id="ul71592205015"><li>on：开启进程级别重调度</li><li>其他值或不使用该字段：关闭进程级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p181177342718"><a name="p181177342718"></a><a name="p181177342718"></a><span id="ph102814152910"><a name="ph102814152910"></a><a name="ph102814152910"></a>Volcano</span></p>
</td>
</tr>
<tr id="row8889122663714"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p172891224132816"><a name="p172891224132816"></a><a name="p172891224132816"></a>subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p728915247282"><a name="p728915247282"></a><a name="p728915247282"></a>亚健康处理策略</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul18716519102210"></a><a name="ul18716519102210"></a><ul id="ul18716519102210"><li>ignore：忽略该亚健康节点，后续任务在亲和性调度上不优先调度该节点。</li><li>graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。</li><li>forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。</li><li>hotSwitch：执行亚健康热切，拉起备份Pod后，暂停训练任务，并使用新节点重新拉起训练。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p929902143119"><a name="p929902143119"></a><a name="p929902143119"></a><span id="ph6299326312"><a name="ph6299326312"></a><a name="ph6299326312"></a>Volcano</span></p>
</td>
</tr>
</tbody>
</table>

**表 2** 集群调度组件对PodGroup annotations使用说明

<a name="table87117712413"></a>
<table><thead align="left"><tr id="row167127122419"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p17127152415"><a name="p17127152415"></a><a name="p17127152415"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="24.169999999999998%" id="mcps1.2.5.1.2"><p id="p14713722416"><a name="p14713722416"></a><a name="p14713722416"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="27.450000000000003%" id="mcps1.2.5.1.3"><p id="p471127192414"><a name="p471127192414"></a><a name="p471127192414"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="23.380000000000003%" id="mcps1.2.5.1.4"><p id="p57187142419"><a name="p57187142419"></a><a name="p57187142419"></a>使用组件</p>
</th>
</tr>
</thead>
<tbody><tr id="row47177202416"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p576592911262"><a name="p576592911262"></a><a name="p576592911262"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p167655293262"><a name="p167655293262"></a><a name="p167655293262"></a>指定逻辑超节点芯片数量。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p77651729202614"><a name="p77651729202614"></a><a name="p77651729202614"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p1072973244"><a name="p1072973244"></a><a name="p1072973244"></a><span id="ph197215716249"><a name="ph197215716249"></a><a name="ph197215716249"></a>Volcano</span>、<span id="ph17212711245"><a name="ph17212711245"></a><a name="ph17212711245"></a>Ascend Operator</span></p>
</td>
</tr>
<tr id="row972875243"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p20765429172612"><a name="p20765429172612"></a><a name="p20765429172612"></a>huawei.com/schedule_policy</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p1276572913269"><a name="p1276572913269"></a><a name="p1276572913269"></a>指定调度策略。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p1389015013273"><a name="p1389015013273"></a><a name="p1389015013273"></a>目前支持<a href="#table1120511613153">表3</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p197214711249"><a name="p197214711249"></a><a name="p197214711249"></a><span id="ph972477246"><a name="ph972477246"></a><a name="ph972477246"></a>Volcano</span></p>
</td>
</tr>
<tr id="row572178247"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p13765229182617"><a name="p13765229182617"></a><a name="p13765229182617"></a>sp-fit</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p1276582913269"><a name="p1276582913269"></a><a name="p1276582913269"></a>超节点调度策略。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p67657296267"><a name="p67657296267"></a><a name="p67657296267"></a>idlest：逻辑超节点会往更空闲的物理超节点调度。</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p77220742415"><a name="p77220742415"></a><a name="p77220742415"></a><span id="ph1372071243"><a name="ph1372071243"></a><a name="ph1372071243"></a>Volcano</span></p>
</td>
</tr>
<tr id="row1721472248"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p3766152922612"><a name="p3766152922612"></a><a name="p3766152922612"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p576614298267"><a name="p576614298267"></a><a name="p576614298267"></a>任务能够调度的最小副本数。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p10766102912261"><a name="p10766102912261"></a><a name="p10766102912261"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p1972147182413"><a name="p1972147182413"></a><a name="p1972147182413"></a><span id="ph57212720245"><a name="ph57212720245"></a><a name="ph57212720245"></a>Volcano</span></p>
</td>
</tr>
<tr id="row6729792413"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p19766129192612"><a name="p19766129192612"></a><a name="p19766129192612"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p376652917262"><a name="p376652917262"></a><a name="p376652917262"></a>任务重调度策略。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p1178352611283"><a name="p1178352611283"></a><a name="p1178352611283"></a>pod：只支持Pod级重调度，不升级为Job级别。</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p13726762413"><a name="p13726762413"></a><a name="p13726762413"></a><span id="ph1672771246"><a name="ph1672771246"></a><a name="ph1672771246"></a>Volcano</span></p>
</td>
</tr>
<tr id="row2032944619369"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p6523192073819"><a name="p6523192073819"></a><a name="p6523192073819"></a>huawei.com/schedule_enable_dequeue</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p95238206385"><a name="p95238206385"></a><a name="p95238206385"></a>是否启动任务可出队（从Inqueue变为Pending状态）功能。需手动配置。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><a name="ul1452313209384"></a><a name="ul1452313209384"></a><ul id="ul1452313209384"><li>“on”：开启</li><li>其他取值：关闭</li></ul>
<p id="p184512184913"><a name="p184512184913"></a><a name="p184512184913"></a>不配置则默认关闭。</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p19523102023811"><a name="p19523102023811"></a><a name="p19523102023811"></a><span id="ph16444326193819"><a name="ph16444326193819"></a><a name="ph16444326193819"></a>Volcano</span></p>
</td>
</tr>
<tr id="row1450448133619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p195231920133819"><a name="p195231920133819"></a><a name="p195231920133819"></a>huawei.com/schedule_dequeue_frequency</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p4523172063820"><a name="p4523172063820"></a><a name="p4523172063820"></a>记录任务出队次数。<span id="ph5862824114017"><a name="ph5862824114017"></a><a name="ph5862824114017"></a>Volcano</span>自动更新。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p125231320193818"><a name="p125231320193818"></a><a name="p125231320193818"></a>任务出队1次，该值加1。</p>
<div class="note" id="note10987851174216"><a name="note10987851174216"></a><div class="notebody"><p id="p698710511425"><a name="p698710511425"></a><a name="p698710511425"></a>任务不处于Inqueue、Pending状态时，删除该值。</p>
</div></div>
<p id="p105231520203811"><a name="p105231520203811"></a><a name="p105231520203811"></a></p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p12523520123817"><a name="p12523520123817"></a><a name="p12523520123817"></a><span id="ph497462713812"><a name="ph497462713812"></a><a name="ph497462713812"></a>Volcano</span></p>
</td>
</tr>
<tr id="row16233175083617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p5523142033819"><a name="p5523142033819"></a><a name="p5523142033819"></a>huawei.com/schedule_enqueue_time</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p152312015383"><a name="p152312015383"></a><a name="p152312015383"></a>记录任务入队（从Pending变为Inqueue状态）时间。<span id="ph19470113214427"><a name="ph19470113214427"></a><a name="ph19470113214427"></a>Volcano</span>自动更新。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p1152332013385"><a name="p1152332013385"></a><a name="p1152332013385"></a>毫秒级时间戳。</p>
<div class="note" id="note813021515431"><a name="note813021515431"></a><div class="notebody"><a name="ul1115755417436"></a><a name="ul1115755417436"></a><ul id="ul1115755417436"><li>若任务入队超5分钟且开启了可出队功能，当有其他任务需要入队时，此任务会出队释放资源，以便其他任务可以入队。</li><li>任务不处于Inqueue状态时，删除该值。</li></ul>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p1752310207382"><a name="p1752310207382"></a><a name="p1752310207382"></a><span id="ph10540629193811"><a name="ph10540629193811"></a><a name="ph10540629193811"></a>Volcano</span></p>
</td>
</tr>
</tbody>
</table>

**表 3**  huawei.com/schedule\_policy配置说明

<a name="table1120511613153"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002511347099_row192066612155"><th class="cellrowborder" valign="top" width="22.3%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002511347099_p132062614153"><a name="zh-cn_topic_0000002511347099_p132062614153"></a><a name="zh-cn_topic_0000002511347099_p132062614153"></a>配置</p>
</th>
<th class="cellrowborder" valign="top" width="77.7%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002511347099_p5206126181520"><a name="zh-cn_topic_0000002511347099_p5206126181520"></a><a name="zh-cn_topic_0000002511347099_p5206126181520"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002511347099_row201261346162"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p457945418181"><a name="zh-cn_topic_0000002511347099_p457945418181"></a><a name="zh-cn_topic_0000002511347099_p457945418181"></a>chip4-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p7579105411817"><a name="zh-cn_topic_0000002511347099_p7579105411817"></a><a name="zh-cn_topic_0000002511347099_p7579105411817"></a>1个节点8张芯片，每4个芯片形成1个互联环。例如，<span id="zh-cn_topic_0000002511347099_ph18314192319429"><a name="zh-cn_topic_0000002511347099_ph18314192319429"></a><a name="zh-cn_topic_0000002511347099_ph18314192319429"></a>Atlas 800 训练服务器（型号 9000）</span>/<span id="zh-cn_topic_0000002511347099_ph631452384213"><a name="zh-cn_topic_0000002511347099_ph631452384213"></a><a name="zh-cn_topic_0000002511347099_ph631452384213"></a>Atlas 800 训练服务器（型号 9010）</span>芯片的整模块场景 /Atlas 350 推理卡内部共8张卡，每4张卡通过UB扣板连接。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row102574171610"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p205801254151810"><a name="zh-cn_topic_0000002511347099_p205801254151810"></a><a name="zh-cn_topic_0000002511347099_p205801254151810"></a>chip1-node2</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p65801354101816"><a name="zh-cn_topic_0000002511347099_p65801354101816"></a><a name="zh-cn_topic_0000002511347099_p65801354101816"></a>1个节点2张芯片。例如，<span id="zh-cn_topic_0000002511347099_ph97657495514"><a name="zh-cn_topic_0000002511347099_ph97657495514"></a><a name="zh-cn_topic_0000002511347099_ph97657495514"></a>Atlas 300T 训练卡</span>的插卡场景，1张卡最多插1个芯片，1个节点最多插2张卡。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row825811151619"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p17580854201815"><a name="zh-cn_topic_0000002511347099_p17580854201815"></a><a name="zh-cn_topic_0000002511347099_p17580854201815"></a>chip4-node4</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p858019546184"><a name="zh-cn_topic_0000002511347099_p858019546184"></a><a name="zh-cn_topic_0000002511347099_p858019546184"></a>1个节点4张芯片，形成1个互联环。例如，<span id="zh-cn_topic_0000002511347099_ph1165491719811"><a name="zh-cn_topic_0000002511347099_ph1165491719811"></a><a name="zh-cn_topic_0000002511347099_ph1165491719811"></a>Atlas 800 训练服务器（型号 9000）</span>/<span id="zh-cn_topic_0000002511347099_ph15654111712815"><a name="zh-cn_topic_0000002511347099_ph15654111712815"></a><a name="zh-cn_topic_0000002511347099_ph15654111712815"></a>Atlas 800 训练服务器（型号 9010）</span>芯片的半配场景。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p2580654131819"><a name="zh-cn_topic_0000002511347099_p2580654131819"></a><a name="zh-cn_topic_0000002511347099_p2580654131819"></a>chip8-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p85801654181818"><a name="zh-cn_topic_0000002511347099_p85801654181818"></a><a name="zh-cn_topic_0000002511347099_p85801654181818"></a>1个节点8张卡，8张卡都在1个互联环上。例如，<span id="zh-cn_topic_0000002511347099_ph14314162316427"><a name="zh-cn_topic_0000002511347099_ph14314162316427"></a><a name="zh-cn_topic_0000002511347099_ph14314162316427"></a>Atlas 800T A2 训练服务器</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row1820613612158"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p1358111544185"><a name="zh-cn_topic_0000002511347099_p1358111544185"></a><a name="zh-cn_topic_0000002511347099_p1358111544185"></a>chip8-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p9581135461815"><a name="zh-cn_topic_0000002511347099_p9581135461815"></a><a name="zh-cn_topic_0000002511347099_p9581135461815"></a>1个节点16张卡，每8张卡在1个互联环上。例如，<span id="zh-cn_topic_0000002511347099_ph1831422311424"><a name="zh-cn_topic_0000002511347099_ph1831422311424"></a><a name="zh-cn_topic_0000002511347099_ph1831422311424"></a>Atlas 200T A2 Box16 异构子框</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row2020613616154"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p2581854121811"><a name="zh-cn_topic_0000002511347099_p2581854121811"></a><a name="zh-cn_topic_0000002511347099_p2581854121811"></a>chip2-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p758125481813"><a name="zh-cn_topic_0000002511347099_p758125481813"></a><a name="zh-cn_topic_0000002511347099_p758125481813"></a>1个节点16张卡，每2张卡在1个互联环上。例如，<span id="zh-cn_topic_0000002511347099_ph855133261011"><a name="zh-cn_topic_0000002511347099_ph855133261011"></a><a name="zh-cn_topic_0000002511347099_ph855133261011"></a>Atlas 800T A3 超节点服务器</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row22064621511"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p558111549188"><a name="zh-cn_topic_0000002511347099_p558111549188"></a><a name="zh-cn_topic_0000002511347099_p558111549188"></a>chip2-node16-sp</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p258115548187"><a name="zh-cn_topic_0000002511347099_p258115548187"></a><a name="zh-cn_topic_0000002511347099_p258115548187"></a>1个节点16张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，<span id="zh-cn_topic_0000002511347099_ph1990844161011"><a name="zh-cn_topic_0000002511347099_ph1990844161011"></a><a name="zh-cn_topic_0000002511347099_ph1990844161011"></a>Atlas 900 A3 SuperPoD 超节点</span>。</p>
</td>
</tr>
<tr id="row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2580654131819"><a name="p2580654131819"></a><a name="p2580654131819"></a>chip4-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p85801654181818"><a name="p85801654181818"></a><a name="p85801654181818"></a>1个节点16张卡，每4张卡都在1个互联环上。例如，<span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 350 推理卡内部共16张卡，每4张卡通过UB扣板连接</span>。</p>
</td>
</tr>
<tr id="row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2580654131819"><a name="p2580654131819"></a><a name="p2580654131819"></a>chip1-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p85801654181818"><a name="p85801654181818"></a><a name="p85801654181818"></a>1个节点8张卡，每张卡之间无互联。例如，<span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 350 推理卡内部共8张卡，每张卡之间无互联</span>。</p>
</td>
</tr>
<tr id="row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2580654131819"><a name="p2580654131819"></a><a name="p2580654131819"></a>chip1-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p85801654181818"><a name="p85801654181818"></a><a name="p85801654181818"></a>1个节点16张卡，每张卡之间无互联。例如，<span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 350 推理卡内部共16张卡，每张卡之间无互联</span>。</p>
</td>
</tr>
</tbody>
</table>


## Pod<a name="ZH-CN_TOPIC_0000002484428552"></a>

**表 1** 集群调度组件对Pod label使用说明

<a name="table143562050699"></a>
<table><thead align="left"><tr id="row23564507918"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p1535615011914"><a name="p1535615011914"></a><a name="p1535615011914"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.2"><p id="p83576501093"><a name="p83576501093"></a><a name="p83576501093"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="p235719501097"><a name="p235719501097"></a><a name="p235719501097"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p435716507913"><a name="p435716507913"></a><a name="p435716507913"></a>使用组件</p>
</th>
</tr>
</thead>
<tbody><tr id="row7357125010917"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0357155019917"><a name="p0357155019917"></a><a name="p0357155019917"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p1035711508917"><a name="p1035711508917"></a><a name="p1035711508917"></a>标识Atlas的<span id="ph13571150291"><a name="ph13571150291"></a><a name="ph13571150291"></a>Pod</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul835765011916"></a><a name="ul835765011916"></a><ul id="ul835765011916"><li>ascend-910</li><li>ascend-<span id="ph19358150597"><a name="ph19358150597"></a><a name="ph19358150597"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li><li>huawei.com/npu</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10471193985417"><a name="p10471193985417"></a><a name="p10471193985417"></a><span id="ph1035865018915"><a name="ph1035865018915"></a><a name="ph1035865018915"></a>Ascend Device Plugin</span>、<span id="ph446593975417"><a name="ph446593975417"></a><a name="ph446593975417"></a>Ascend Operator</span>、<span id="ph635885012911"><a name="ph635885012911"></a><a name="ph635885012911"></a>Volcano</span></p>
</td>
</tr>
<tr id="row135825013910"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p203581150493"><a name="p203581150493"></a><a name="p203581150493"></a>fault-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p935815502914"><a name="p935815502914"></a><a name="p935815502914"></a>任务故障重调度开关</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p5358950291"><a name="p5358950291"></a><a name="p5358950291"></a>grace、force、off</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10358105012913"><a name="p10358105012913"></a><a name="p10358105012913"></a><span id="ph635812501497"><a name="ph635812501497"></a><a name="ph635812501497"></a>Volcano</span>、<span id="ph183581350898"><a name="ph183581350898"></a><a name="ph183581350898"></a>Resilience Controller</span></p>
</td>
</tr>
<tr id="row03591501297"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p143599501394"><a name="p143599501394"></a><a name="p143599501394"></a>elastic-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p1235917508916"><a name="p1235917508916"></a><a name="p1235917508916"></a>任务弹性调度开关</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p1135910503918"><a name="p1135910503918"></a><a name="p1135910503918"></a>on</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1735916501991"><a name="p1735916501991"></a><a name="p1735916501991"></a><span id="ph93613501992"><a name="ph93613501992"></a><a name="ph93613501992"></a>Resilience Controller</span>、<span id="ph53614501198"><a name="ph53614501198"></a><a name="ph53614501198"></a>Volcano</span></p>
</td>
</tr>
<tr id="row103614504912"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p6361950695"><a name="p6361950695"></a><a name="p6361950695"></a>fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p183613501791"><a name="p183613501791"></a><a name="p183613501791"></a>任务发生业务面故障可以重调度的次数</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p1836115501693"><a name="p1836115501693"></a><a name="p1836115501693"></a>0-100</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p636212501391"><a name="p636212501391"></a><a name="p636212501391"></a><span id="ph1036213501896"><a name="ph1036213501896"></a><a name="ph1036213501896"></a>Volcano</span>、<span id="ph436285014913"><a name="ph436285014913"></a><a name="ph436285014913"></a>Ascend Operator</span></p>
</td>
</tr>
<tr id="row1336212502091"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p103625502094"><a name="p103625502094"></a><a name="p103625502094"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p43628508911"><a name="p43628508911"></a><a name="p43628508911"></a>交换机亲和性策略</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul143629507913"></a><a name="ul143629507913"></a><ul id="ul143629507913"><li>normal-schema</li><li>large-model-schema</li><li>null</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p436310506914"><a name="p436310506914"></a><a name="p436310506914"></a><span id="ph73631050690"><a name="ph73631050690"></a><a name="ph73631050690"></a>Volcano</span></p>
</td>
</tr>
<tr id="row1136411501898"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0364135010913"><a name="p0364135010913"></a><a name="p0364135010913"></a>npu-310-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p636425010918"><a name="p636425010918"></a><a name="p636425010918"></a>标记推理服务器（插<span id="ph1436410501390"><a name="ph1436410501390"></a><a name="ph1436410501390"></a>Atlas 300I 推理卡</span>）调度策略</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul73644501797"></a><a name="ul73644501797"></a><ul id="ul73644501797"><li>card</li><li>chip</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1636419501398"><a name="p1636419501398"></a><a name="p1636419501398"></a><span id="ph163652501393"><a name="ph163652501393"></a><a name="ph163652501393"></a>Volcano</span></p>
</td>
</tr>
<tr id="row7970125593620"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p65914912716"><a name="p65914912716"></a><a name="p65914912716"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p3591694276"><a name="p3591694276"></a><a name="p3591694276"></a>是否启用Pod级别重调度。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul186101614131"></a><a name="ul186101614131"></a><ul id="ul186101614131"><li>on：开启Pod级别重调度</li><li>其他值或不使用该字段：关闭Pod级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1372045172812"><a name="p1372045172812"></a><a name="p1372045172812"></a><span id="ph2072005192818"><a name="ph2072005192818"></a><a name="ph2072005192818"></a>Volcano</span></p>
</td>
</tr>
<tr id="row209101813153710"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p2417162275410"><a name="p2417162275410"></a><a name="p2417162275410"></a>process-recover-enable</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p131172319273"><a name="p131172319273"></a><a name="p131172319273"></a>是否启用进程级别重调度。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul71592205015"></a><a name="ul71592205015"></a><ul id="ul71592205015"><li>on：开启进程级别重调度</li><li>其他值或不使用该字段：关闭进程级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p181177342718"><a name="p181177342718"></a><a name="p181177342718"></a><span id="ph102814152910"><a name="ph102814152910"></a><a name="ph102814152910"></a>Volcano</span></p>
</td>
</tr>
<tr id="row8889122663714"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p172891224132816"><a name="p172891224132816"></a><a name="p172891224132816"></a>subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p728915247282"><a name="p728915247282"></a><a name="p728915247282"></a>亚健康处理策略</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><a name="ul18716519102210"></a><a name="ul18716519102210"></a><ul id="ul18716519102210"><li>ignore：忽略该亚健康节点，后续任务在亲和性调度上不优先调度该节点。</li><li>graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。</li><li>forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。</li><li>hotSwitch：执行亚健康热切，拉起备份Pod后，暂停训练任务，并使用新节点重新拉起训练。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p929902143119"><a name="p929902143119"></a><a name="p929902143119"></a><span id="ph6299326312"><a name="ph6299326312"></a><a name="ph6299326312"></a>Volcano</span></p>
</td>
</tr>
</tbody>
</table>

**表 2** 集群调度组件对Pod annotations使用说明

<a name="table87117712413"></a>
<table><thead align="left"><tr id="row167127122419"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p17127152415"><a name="p17127152415"></a><a name="p17127152415"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="24.169999999999998%" id="mcps1.2.5.1.2"><p id="p14713722416"><a name="p14713722416"></a><a name="p14713722416"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="27.450000000000003%" id="mcps1.2.5.1.3"><p id="p471127192414"><a name="p471127192414"></a><a name="p471127192414"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="23.380000000000003%" id="mcps1.2.5.1.4"><p id="p57187142419"><a name="p57187142419"></a><a name="p57187142419"></a>使用组件</p>
</th>
</tr>
</thead>
<tbody><tr id="row47177202416"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p576592911262"><a name="p576592911262"></a><a name="p576592911262"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p167655293262"><a name="p167655293262"></a><a name="p167655293262"></a>指定逻辑超节点芯片数量。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p77651729202614"><a name="p77651729202614"></a><a name="p77651729202614"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p1072973244"><a name="p1072973244"></a><a name="p1072973244"></a><span id="ph197215716249"><a name="ph197215716249"></a><a name="ph197215716249"></a>Volcano</span>、<span id="ph17212711245"><a name="ph17212711245"></a><a name="ph17212711245"></a>Ascend Operator</span></p>
</td>
</tr>
<tr id="row972875243"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p20765429172612"><a name="p20765429172612"></a><a name="p20765429172612"></a>huawei.com/schedule_policy</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p1276572913269"><a name="p1276572913269"></a><a name="p1276572913269"></a>指定调度策略。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p1389015013273"><a name="p1389015013273"></a><a name="p1389015013273"></a>目前支持<a href="#table1120511613153">表3</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p197214711249"><a name="p197214711249"></a><a name="p197214711249"></a><span id="ph972477246"><a name="ph972477246"></a><a name="ph972477246"></a>Volcano</span></p>
</td>
</tr>
<tr id="row572178247"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p13765229182617"><a name="p13765229182617"></a><a name="p13765229182617"></a>sp-fit</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p1276582913269"><a name="p1276582913269"></a><a name="p1276582913269"></a>超节点调度策略。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p67657296267"><a name="p67657296267"></a><a name="p67657296267"></a>idlest：逻辑超节点会往更空闲的物理超节点调度。</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p77220742415"><a name="p77220742415"></a><a name="p77220742415"></a><span id="ph1372071243"><a name="ph1372071243"></a><a name="ph1372071243"></a>Volcano</span></p>
</td>
</tr>
<tr id="row1721472248"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p3766152922612"><a name="p3766152922612"></a><a name="p3766152922612"></a>huawei.com/schedule_minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p576614298267"><a name="p576614298267"></a><a name="p576614298267"></a>任务能够调度的最小副本数。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p10766102912261"><a name="p10766102912261"></a><a name="p10766102912261"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p1972147182413"><a name="p1972147182413"></a><a name="p1972147182413"></a><span id="ph57212720245"><a name="ph57212720245"></a><a name="ph57212720245"></a>Volcano</span></p>
</td>
</tr>
<tr id="row6729792413"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p19766129192612"><a name="p19766129192612"></a><a name="p19766129192612"></a>huawei.com/recover_policy_path</p>
</td>
<td class="cellrowborder" valign="top" width="24.169999999999998%" headers="mcps1.2.5.1.2 "><p id="p376652917262"><a name="p376652917262"></a><a name="p376652917262"></a>任务重调度策略。</p>
</td>
<td class="cellrowborder" valign="top" width="27.450000000000003%" headers="mcps1.2.5.1.3 "><p id="p1178352611283"><a name="p1178352611283"></a><a name="p1178352611283"></a>pod：只支持Pod级重调度，不升级为Job级别。</p>
</td>
<td class="cellrowborder" valign="top" width="23.380000000000003%" headers="mcps1.2.5.1.4 "><p id="p13726762413"><a name="p13726762413"></a><a name="p13726762413"></a><span id="ph1672771246"><a name="ph1672771246"></a><a name="ph1672771246"></a>Volcano</span></p>
</td>
</tr>
</tbody>
</table>

**表 3**  huawei.com/schedule\_policy配置说明

<a name="table1120511613153"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002511347099_row192066612155"><th class="cellrowborder" valign="top" width="22.3%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002511347099_p132062614153"><a name="zh-cn_topic_0000002511347099_p132062614153"></a><a name="zh-cn_topic_0000002511347099_p132062614153"></a>配置</p>
</th>
<th class="cellrowborder" valign="top" width="77.7%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002511347099_p5206126181520"><a name="zh-cn_topic_0000002511347099_p5206126181520"></a><a name="zh-cn_topic_0000002511347099_p5206126181520"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002511347099_row201261346162"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p457945418181"><a name="zh-cn_topic_0000002511347099_p457945418181"></a><a name="zh-cn_topic_0000002511347099_p457945418181"></a>chip4-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p7579105411817"><a name="zh-cn_topic_0000002511347099_p7579105411817"></a><a name="zh-cn_topic_0000002511347099_p7579105411817"></a>1个节点8张芯片，每4个芯片形成1个互联环。例如，<span id="zh-cn_topic_0000002511347099_ph18314192319429"><a name="zh-cn_topic_0000002511347099_ph18314192319429"></a><a name="zh-cn_topic_0000002511347099_ph18314192319429"></a>Atlas 800 训练服务器（型号 9000）</span>/<span id="zh-cn_topic_0000002511347099_ph631452384213"><a name="zh-cn_topic_0000002511347099_ph631452384213"></a><a name="zh-cn_topic_0000002511347099_ph631452384213"></a>Atlas 800 训练服务器（型号 9010）</span>芯片的整模块场景 /Atlas 350 推理卡内部共8张卡，每4张卡通过UB扣板连接。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row102574171610"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p205801254151810"><a name="zh-cn_topic_0000002511347099_p205801254151810"></a><a name="zh-cn_topic_0000002511347099_p205801254151810"></a>chip1-node2</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p65801354101816"><a name="zh-cn_topic_0000002511347099_p65801354101816"></a><a name="zh-cn_topic_0000002511347099_p65801354101816"></a>1个节点2张芯片。例如，<span id="zh-cn_topic_0000002511347099_ph97657495514"><a name="zh-cn_topic_0000002511347099_ph97657495514"></a><a name="zh-cn_topic_0000002511347099_ph97657495514"></a>Atlas 300T 训练卡</span>的插卡场景，1张卡最多插1个芯片，1个节点最多插2张卡。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row825811151619"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p17580854201815"><a name="zh-cn_topic_0000002511347099_p17580854201815"></a><a name="zh-cn_topic_0000002511347099_p17580854201815"></a>chip4-node4</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p858019546184"><a name="zh-cn_topic_0000002511347099_p858019546184"></a><a name="zh-cn_topic_0000002511347099_p858019546184"></a>1个节点4张芯片，形成1个互联环。例如，<span id="zh-cn_topic_0000002511347099_ph1165491719811"><a name="zh-cn_topic_0000002511347099_ph1165491719811"></a><a name="zh-cn_topic_0000002511347099_ph1165491719811"></a>Atlas 800 训练服务器（型号 9000）</span>/<span id="zh-cn_topic_0000002511347099_ph15654111712815"><a name="zh-cn_topic_0000002511347099_ph15654111712815"></a><a name="zh-cn_topic_0000002511347099_ph15654111712815"></a>Atlas 800 训练服务器（型号 9010）</span>芯片的半配场景。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p2580654131819"><a name="zh-cn_topic_0000002511347099_p2580654131819"></a><a name="zh-cn_topic_0000002511347099_p2580654131819"></a>chip8-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p85801654181818"><a name="zh-cn_topic_0000002511347099_p85801654181818"></a><a name="zh-cn_topic_0000002511347099_p85801654181818"></a>1个节点8张卡，8张卡都在1个互联环上。例如，<span id="zh-cn_topic_0000002511347099_ph14314162316427"><a name="zh-cn_topic_0000002511347099_ph14314162316427"></a><a name="zh-cn_topic_0000002511347099_ph14314162316427"></a>Atlas 800T A2 训练服务器</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row1820613612158"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p1358111544185"><a name="zh-cn_topic_0000002511347099_p1358111544185"></a><a name="zh-cn_topic_0000002511347099_p1358111544185"></a>chip8-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p9581135461815"><a name="zh-cn_topic_0000002511347099_p9581135461815"></a><a name="zh-cn_topic_0000002511347099_p9581135461815"></a>1个节点16张卡，每8张卡在1个互联环上。例如，<span id="zh-cn_topic_0000002511347099_ph1831422311424"><a name="zh-cn_topic_0000002511347099_ph1831422311424"></a><a name="zh-cn_topic_0000002511347099_ph1831422311424"></a>Atlas 200T A2 Box16 异构子框</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row2020613616154"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p2581854121811"><a name="zh-cn_topic_0000002511347099_p2581854121811"></a><a name="zh-cn_topic_0000002511347099_p2581854121811"></a>chip2-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p758125481813"><a name="zh-cn_topic_0000002511347099_p758125481813"></a><a name="zh-cn_topic_0000002511347099_p758125481813"></a>1个节点16张卡，每2张卡在1个互联环上。例如，<span id="zh-cn_topic_0000002511347099_ph855133261011"><a name="zh-cn_topic_0000002511347099_ph855133261011"></a><a name="zh-cn_topic_0000002511347099_ph855133261011"></a>Atlas 800T A3 超节点服务器</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002511347099_row22064621511"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002511347099_p558111549188"><a name="zh-cn_topic_0000002511347099_p558111549188"></a><a name="zh-cn_topic_0000002511347099_p558111549188"></a>chip2-node16-sp</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002511347099_p258115548187"><a name="zh-cn_topic_0000002511347099_p258115548187"></a><a name="zh-cn_topic_0000002511347099_p258115548187"></a>1个节点16张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，<span id="zh-cn_topic_0000002511347099_ph1990844161011"><a name="zh-cn_topic_0000002511347099_ph1990844161011"></a><a name="zh-cn_topic_0000002511347099_ph1990844161011"></a>Atlas 900 A3 SuperPoD 超节点</span>。</p>
</td>
</tr>
<tr id="row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2580654131819"><a name="p2580654131819"></a><a name="p2580654131819"></a>chip4-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p85801654181818"><a name="p85801654181818"></a><a name="p85801654181818"></a>1个节点16张卡，每4张卡都在1个互联环上。例如，<span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 350 推理卡内部共16张卡，每4张卡通过UB扣板连接</span>。</p>
</td>
</tr>
<tr id="row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2580654131819"><a name="p2580654131819"></a><a name="p2580654131819"></a>chip1-node8</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p85801654181818"><a name="p85801654181818"></a><a name="p85801654181818"></a>1个节点8张卡，每张卡之间无互联。例如，<span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 350 推理卡内部共8张卡，每张卡之间无互联</span>。</p>
</td>
</tr>
<tr id="row1925831181613"><td class="cellrowborder" valign="top" width="22.3%" headers="mcps1.2.3.1.1 "><p id="p2580654131819"><a name="p2580654131819"></a><a name="p2580654131819"></a>chip1-node16</p>
</td>
<td class="cellrowborder" valign="top" width="77.7%" headers="mcps1.2.3.1.2 "><p id="p85801654181818"><a name="p85801654181818"></a><a name="p85801654181818"></a>1个节点16张卡，每张卡之间无互联。例如，<span id="ph14314162316427"><a name="ph14314162316427"></a><a name="ph14314162316427"></a>Atlas 350 推理卡内部共16张卡，每张卡之间无互联</span>。</p>
</td>
</tr>
</tbody>
</table>


## 任务信息<a name="ZH-CN_TOPIC_0000002479386798"></a>

**tor-share-cm<a name="section98191810400"></a>**

**表 1**  tor-share-cm

<a name="table185653715301"></a>
<table><thead align="left"><tr id="row1857037203020"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p385703733012"><a name="p385703733012"></a><a name="p385703733012"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="26.75%" id="mcps1.2.5.1.2"><p id="p28571437193018"><a name="p28571437193018"></a><a name="p28571437193018"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="23.25%" id="mcps1.2.5.1.3"><p id="p1385803713018"><a name="p1385803713018"></a><a name="p1385803713018"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p68588374304"><a name="p68588374304"></a><a name="p68588374304"></a>备注</p>
</th>
</tr>
</thead>
<tbody><tr id="row685863711309"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p138581437143012"><a name="p138581437143012"></a><a name="p138581437143012"></a>IsHealthy</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p1485818378307"><a name="p1485818378307"></a><a name="p1485818378307"></a>节点对应的交换机状态</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p17859193717305"><a name="p17859193717305"></a><a name="p17859193717305"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1085918371305"><a name="p1085918371305"></a><a name="p1085918371305"></a>-</p>
</td>
</tr>
<tr id="row585918375300"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p585973723020"><a name="p585973723020"></a><a name="p585973723020"></a>IsSharedTor</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p1185913715307"><a name="p1185913715307"></a><a name="p1185913715307"></a>节点对应的交换机属性</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p198596377302"><a name="p198596377302"></a><a name="p198596377302"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p8860637123014"><a name="p8860637123014"></a><a name="p8860637123014"></a>-</p>
</td>
</tr>
<tr id="row1286018378301"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p19860337183019"><a name="p19860337183019"></a><a name="p19860337183019"></a>NodeIP</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p18605378309"><a name="p18605378309"></a><a name="p18605378309"></a>节点IP</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p186043718309"><a name="p186043718309"></a><a name="p186043718309"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p58611837193020"><a name="p58611837193020"></a><a name="p58611837193020"></a>-</p>
</td>
</tr>
<tr id="row08615377305"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1186283713305"><a name="p1186283713305"></a><a name="p1186283713305"></a>NodeName</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p68629374305"><a name="p68629374305"></a><a name="p68629374305"></a>节点名称</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1286383711307"><a name="p1286383711307"></a><a name="p1286383711307"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p158631337173013"><a name="p158631337173013"></a><a name="p158631337173013"></a>-</p>
</td>
</tr>
<tr id="row8863337133015"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p18863153743016"><a name="p18863153743016"></a><a name="p18863153743016"></a>JobName</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p19863143743015"><a name="p19863143743015"></a><a name="p19863143743015"></a>任务名称</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p78631337113011"><a name="p78631337113011"></a><a name="p78631337113011"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p086473713017"><a name="p086473713017"></a><a name="p086473713017"></a>-</p>
</td>
</tr>
</tbody>
</table>

**vcjob-fault-npu-cm<a name="section1731892963620"></a>**

**表 2** vcjob-fault-npu-cm字段说明

<a name="table153041817110"></a>
<table><thead align="left"><tr id="row4530818101120"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p1653191871120"><a name="p1653191871120"></a><a name="p1653191871120"></a><span id="ph135612450384"><a name="ph135612450384"></a><a name="ph135612450384"></a>名称</span></p>
</th>
<th class="cellrowborder" valign="top" width="26.75%" id="mcps1.2.5.1.2"><p id="p353111818113"><a name="p353111818113"></a><a name="p353111818113"></a><span id="ph4571459382"><a name="ph4571459382"></a><a name="ph4571459382"></a>作用</span></p>
</th>
<th class="cellrowborder" valign="top" width="23.25%" id="mcps1.2.5.1.3"><p id="p6531101821116"><a name="p6531101821116"></a><a name="p6531101821116"></a><span id="ph12579458385"><a name="ph12579458385"></a><a name="ph12579458385"></a>取值</span></p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p1753115188111"><a name="p1753115188111"></a><a name="p1753115188111"></a><span id="ph658045153811"><a name="ph658045153811"></a><a name="ph658045153811"></a>备注</span></p>
</th>
</tr>
</thead>
<tbody><tr id="row14547818131118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1454791821114"><a name="p1454791821114"></a><a name="p1454791821114"></a><span id="ph1158545163819"><a name="ph1158545163819"></a><a name="ph1158545163819"></a>fault-node</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p5547131815113"><a name="p5547131815113"></a><a name="p5547131815113"></a><span id="ph1359154516384"><a name="ph1359154516384"></a><a name="ph1359154516384"></a>故障节点信息</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1454719188113"><a name="p1454719188113"></a><a name="p1454719188113"></a><span id="ph11601045193815"><a name="ph11601045193815"></a><a name="ph11601045193815"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p154731817113"><a name="p154731817113"></a><a name="p154731817113"></a><span id="ph260204519383"><a name="ph260204519383"></a><a name="ph260204519383"></a>-</span></p>
</td>
</tr>
<tr id="row0547118101117"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p155476181111"><a name="p155476181111"></a><a name="p155476181111"></a><span id="ph186084513387"><a name="ph186084513387"></a><a name="ph186084513387"></a>- NodeName</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p254841814114"><a name="p254841814114"></a><a name="p254841814114"></a><span id="ph1161174511388"><a name="ph1161174511388"></a><a name="ph1161174511388"></a>节点名称</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p054815183116"><a name="p054815183116"></a><a name="p054815183116"></a><span id="ph15611245113815"><a name="ph15611245113815"></a><a name="ph15611245113815"></a>字符串</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p2054851813118"><a name="p2054851813118"></a><a name="p2054851813118"></a><span id="ph1621745143812"><a name="ph1621745143812"></a><a name="ph1621745143812"></a>-</span></p>
</td>
</tr>
<tr id="row55481518151111"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p954816185116"><a name="p954816185116"></a><a name="p954816185116"></a><span id="ph36311451382"><a name="ph36311451382"></a><a name="ph36311451382"></a>- UpdateTime</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p1554814184117"><a name="p1554814184117"></a><a name="p1554814184117"></a><span id="ph1964154513813"><a name="ph1964154513813"></a><a name="ph1964154513813"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p105487180118"><a name="p105487180118"></a><a name="p105487180118"></a><span id="ph3665457384"><a name="ph3665457384"></a><a name="ph3665457384"></a>64位整数类型</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p11548151841113"><a name="p11548151841113"></a><a name="p11548151841113"></a><span id="ph1682459383"><a name="ph1682459383"></a><a name="ph1682459383"></a>-</span></p>
</td>
</tr>
<tr id="row11549151819118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p75491618121111"><a name="p75491618121111"></a><a name="p75491618121111"></a><span id="ph2069144503818"><a name="ph2069144503818"></a><a name="ph2069144503818"></a>- UnhealthyNPU</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p17549518101114"><a name="p17549518101114"></a><a name="p17549518101114"></a><span id="ph57194583812"><a name="ph57194583812"></a><a name="ph57194583812"></a>故障节点上芯片故障的芯片集合</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p35499187114"><a name="p35499187114"></a><a name="p35499187114"></a><span id="ph8711945133815"><a name="ph8711945133815"></a><a name="ph8711945133815"></a>字符串切片</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p17549318151111"><a name="p17549318151111"></a><a name="p17549318151111"></a><span id="ph18721545163813"><a name="ph18721545163813"></a><a name="ph18721545163813"></a>-</span></p>
</td>
</tr>
<tr id="row95491186111"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p25501518181120"><a name="p25501518181120"></a><a name="p25501518181120"></a><span id="ph573164511386"><a name="ph573164511386"></a><a name="ph573164511386"></a>- NetworkUnhealthyNPU</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p2550161841120"><a name="p2550161841120"></a><a name="p2550161841120"></a><span id="ph873174553816"><a name="ph873174553816"></a><a name="ph873174553816"></a>故障节点上网络故障的芯片集合</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1955041814116"><a name="p1955041814116"></a><a name="p1955041814116"></a><span id="ph12741045163816"><a name="ph12741045163816"></a><a name="ph12741045163816"></a>字符串切片</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p65505186116"><a name="p65505186116"></a><a name="p65505186116"></a><span id="ph67494593817"><a name="ph67494593817"></a><a name="ph67494593817"></a>-</span></p>
</td>
</tr>
<tr id="row7551201831116"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p6551181831118"><a name="p6551181831118"></a><a name="p6551181831118"></a><span id="ph127544519384"><a name="ph127544519384"></a><a name="ph127544519384"></a>- NodeDEnable</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p17551318111111"><a name="p17551318111111"></a><a name="p17551318111111"></a><span id="ph776114513812"><a name="ph776114513812"></a><a name="ph776114513812"></a>节点状态检测开关是否打开</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><a name="ul55510181111"></a><a name="ul55510181111"></a><ul id="ul55510181111"><li><span id="ph078174514388"><a name="ph078174514388"></a><a name="ph078174514388"></a>True</span></li><li><span id="ph138054563812"><a name="ph138054563812"></a><a name="ph138054563812"></a>False</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p3552171818117"><a name="p3552171818117"></a><a name="p3552171818117"></a><span id="ph081545103815"><a name="ph081545103815"></a><a name="ph081545103815"></a>-</span></p>
</td>
</tr>
<tr id="row95521718111118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8552111817116"><a name="p8552111817116"></a><a name="p8552111817116"></a><span id="ph17811545153817"><a name="ph17811545153817"></a><a name="ph17811545153817"></a>- NodeHealthState</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p14552191817115"><a name="p14552191817115"></a><a name="p14552191817115"></a><span id="ph1082184563820"><a name="ph1082184563820"></a><a name="ph1082184563820"></a>节点健康状态</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p355221821111"><a name="p355221821111"></a><a name="p355221821111"></a><span id="ph882184593816"><a name="ph882184593816"></a><a name="ph882184593816"></a>字符串</span></p>
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
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p4568118151117"><a name="p4568118151117"></a><a name="p4568118151117"></a><span id="ph195104520386"><a name="ph195104520386"></a><a name="ph195104520386"></a>故障类型对象</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><a name="ul15568201841114"></a><a name="ul15568201841114"></a><ul id="ul15568201841114"><li><span id="ph179524583815"><a name="ph179524583815"></a><a name="ph179524583815"></a>CardUnhealthy：芯片故障</span></li><li><span id="ph596845183816"><a name="ph596845183816"></a><a name="ph596845183816"></a>CardNetworkUnhealthy：芯片网络故障</span></li><li><span id="ph139684533810"><a name="ph139684533810"></a><a name="ph139684533810"></a>NodeUnhealthy：节点故障</span></li></ul>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p85692183111"><a name="p85692183111"></a><a name="p85692183111"></a><span id="ph1497114514381"><a name="ph1497114514381"></a><a name="ph1497114514381"></a>-</span></p>
</td>
</tr>
<tr id="row4569191813112"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p4569171841117"><a name="p4569171841117"></a><a name="p4569171841117"></a><span id="ph59720456384"><a name="ph59720456384"></a><a name="ph59720456384"></a>- npu_name</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p256931831110"><a name="p256931831110"></a><a name="p256931831110"></a><span id="ph189811454383"><a name="ph189811454383"></a><a name="ph189811454383"></a>故障的芯片名称，节点故障时为空</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p65691318151118"><a name="p65691318151118"></a><a name="p65691318151118"></a><span id="ph169812450388"><a name="ph169812450388"></a><a name="ph169812450388"></a>字符串</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p356931818115"><a name="p356931818115"></a><a name="p356931818115"></a><span id="ph129994517380"><a name="ph129994517380"></a><a name="ph129994517380"></a>-</span></p>
</td>
</tr>
<tr id="row11570131817115"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1057018180118"><a name="p1057018180118"></a><a name="p1057018180118"></a><span id="ph1899445113813"><a name="ph1899445113813"></a><a name="ph1899445113813"></a>- fault_level</span></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p1257021816119"><a name="p1257021816119"></a><a name="p1257021816119"></a><span id="ph510084533811"><a name="ph510084533811"></a><a name="ph510084533811"></a>故障处理类型，节点故障时取值为空</span></p>
<p id="p7570111819115"><a name="p7570111819115"></a><a name="p7570111819115"></a></p>
<p id="p12570171881115"><a name="p12570171881115"></a><a name="p12570171881115"></a></p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><a name="ul1157001871115"></a><a name="ul1157001871115"></a><ul id="ul1157001871115"><li>NotHandleFault：不做处理</li><li>RestartRequest：推理场景需要重新执行推理请求，训练场景重新执行训练业务</li><li>RestartBusiness：需要重新执行业务</li><li>FreeRestartNPU：影响业务执行，待芯片空闲时需复位芯片</li><li>RestartNPU：直接复位芯片并重新执行业务</li><li>SeparateNPU：隔离芯片</li><li>PreSeparateNPU：预隔离芯片，会根据训练任务实际运行情况判断是否重调度</li></ul>
</td>
<td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.5.1.4 "><div class="note" id="note11570618121119"><a name="note11570618121119"></a><div class="notebody"><a name="ul17072011133917"></a><a name="ul17072011133917"></a><ul id="ul17072011133917"><li><span id="ph181001745123813"><a name="ph181001745123813"></a><a name="ph181001745123813"></a>fault_level、fault_handling和large_model_fault_level参数功能一致，推荐使用fault_handling。</span></li><li>若推理任务订阅了故障信息，任务使用的推理卡上发生RestartRequest故障且故障持续时间未超过60秒，则不执行任务重调度；若故障持续时间超过60秒仍未恢复，则隔离芯片，进行任务重调度。</li></ul>
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
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p857117181110"><a name="p857117181110"></a><a name="p857117181110"></a><span id="ph1410384543813"><a name="ph1410384543813"></a><a name="ph1410384543813"></a>故障码，由英文逗号拼接而成的字符串</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p057261817110"><a name="p057261817110"></a><a name="p057261817110"></a><span id="ph171041450384"><a name="ph171041450384"></a><a name="ph171041450384"></a>字符串</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><a name="ul057261815113"></a><a name="ul057261815113"></a><ul id="ul057261815113"><li><span id="ph41041545133816"><a name="ph41041545133816"></a><a name="ph41041545133816"></a>Disconnected：芯片网络不连通故障。</span></li><li><span id="ph31051045203810"><a name="ph31051045203810"></a><a name="ph31051045203810"></a>heartbeatTimeOut：节点状态丢失故障</span></li></ul>
</td>
</tr>
<tr id="row1757216185116"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1357251820116"><a name="p1357251820116"></a><a name="p1357251820116"></a><span id="ph181051445163811"><a name="ph181051445163811"></a><a name="ph181051445163811"></a>remain-retry-times</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p35725186119"><a name="p35725186119"></a><a name="p35725186119"></a><span id="ph17106945133816"><a name="ph17106945133816"></a><a name="ph17106945133816"></a>任务剩余可重调度信息</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p11573181820118"><a name="p11573181820118"></a><a name="p11573181820118"></a><span id="ph12106134593818"><a name="ph12106134593818"></a><a name="ph12106134593818"></a>-</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p857381812119"><a name="p857381812119"></a><a name="p857381812119"></a><span id="ph181064454387"><a name="ph181064454387"></a><a name="ph181064454387"></a>-</span></p>
</td>
</tr>
<tr id="row1057312188112"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p35731618131111"><a name="p35731618131111"></a><a name="p35731618131111"></a><span id="ph1510784515383"><a name="ph1510784515383"></a><a name="ph1510784515383"></a>- UUID</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p11573141813118"><a name="p11573141813118"></a><a name="p11573141813118"></a><span id="ph8107104543812"><a name="ph8107104543812"></a><a name="ph8107104543812"></a>任务UID</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p3573141861113"><a name="p3573141861113"></a><a name="p3573141861113"></a><span id="ph1122184510383"><a name="ph1122184510383"></a><a name="ph1122184510383"></a>字符串</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p15731518101119"><a name="p15731518101119"></a><a name="p15731518101119"></a><span id="ph9123154533816"><a name="ph9123154533816"></a><a name="ph9123154533816"></a>-</span></p>
</td>
</tr>
<tr id="row1457316187116"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p12573151841115"><a name="p12573151841115"></a><a name="p12573151841115"></a><span id="ph512334517386"><a name="ph512334517386"></a><a name="ph512334517386"></a>- Times</span></p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p185741118141119"><a name="p185741118141119"></a><a name="p185741118141119"></a><span id="ph17123134517384"><a name="ph17123134517384"></a><a name="ph17123134517384"></a>任务剩余可重调度次数</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p18574151841113"><a name="p18574151841113"></a><a name="p18574151841113"></a><span id="ph8124545193810"><a name="ph8124545193810"></a><a name="ph8124545193810"></a>整数类型</span></p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p457421819116"><a name="p457421819116"></a><a name="p457421819116"></a><span id="ph191247456380"><a name="ph191247456380"></a><a name="ph191247456380"></a>-</span></p>
</td>
</tr>
</tbody>
</table>

**reset-config-<任务名称\><a name="section3394547123916"></a>**

MindCluster集群调度组件通过K8s将设备和训练任务状态等信息写入reset-config-<任务名称\> ConfigMap中，并映射到容器内。Elastic Agent读取后进行相应的故障检测与处理。

**表 3**  reset-config-_<job-name\>_

<a name="table1213115712136"></a>
<table><thead align="left"><tr id="row3132772132"><th class="cellrowborder" valign="top" width="13.940000000000003%" id="mcps1.2.6.1.1"><p id="p207081513112812"><a name="p207081513112812"></a><a name="p207081513112812"></a>字段名称</p>
</th>
<th class="cellrowborder" valign="top" width="16.700000000000003%" id="mcps1.2.6.1.2"><p id="p1313212741314"><a name="p1313212741314"></a><a name="p1313212741314"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="22.28%" id="mcps1.2.6.1.3"><p id="p513317151314"><a name="p513317151314"></a><a name="p513317151314"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="31.220000000000002%" id="mcps1.2.6.1.4"><p id="p313315721314"><a name="p313315721314"></a><a name="p313315721314"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="15.860000000000001%" id="mcps1.2.6.1.5"><p id="p1313327191318"><a name="p1313327191318"></a><a name="p1313327191318"></a>备注</p>
</th>
</tr>
</thead>
<tbody><tr id="row41336711317"><td class="cellrowborder" rowspan="11" valign="top" width="13.940000000000003%" headers="mcps1.2.6.1.1 "><p id="p4472165312280"><a name="p4472165312280"></a><a name="p4472165312280"></a>reset.json</p>
</td>
<td class="cellrowborder" valign="top" width="16.700000000000003%" headers="mcps1.2.6.1.2 "><p id="p813420781315"><a name="p813420781315"></a><a name="p813420781315"></a>RankList</p>
</td>
<td class="cellrowborder" valign="top" width="22.28%" headers="mcps1.2.6.1.3 "><p id="p121346712134"><a name="p121346712134"></a><a name="p121346712134"></a>芯片列表</p>
</td>
<td class="cellrowborder" valign="top" width="31.220000000000002%" headers="mcps1.2.6.1.4 "><p id="p5134137121315"><a name="p5134137121315"></a><a name="p5134137121315"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="15.860000000000001%" headers="mcps1.2.6.1.5 "><p id="p1513427131320"><a name="p1513427131320"></a><a name="p1513427131320"></a>-</p>
</td>
</tr>
<tr id="row21341174135"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p171346791316"><a name="p171346791316"></a><a name="p171346791316"></a>RankId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p3134177131313"><a name="p3134177131313"></a><a name="p3134177131313"></a>故障任务使用的Rank信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1413587161310"><a name="p1413587161310"></a><a name="p1413587161310"></a>整数类型</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1413511721318"><a name="p1413511721318"></a><a name="p1413511721318"></a>-</p>
</td>
</tr>
<tr id="row1713512717138"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p161352712139"><a name="p161352712139"></a><a name="p161352712139"></a>LogicId</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1135127181319"><a name="p1135127181319"></a><a name="p1135127181319"></a>芯片逻辑ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p15135157131311"><a name="p15135157131311"></a><a name="p15135157131311"></a>32位整数类型</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p181366715137"><a name="p181366715137"></a><a name="p181366715137"></a>-</p>
</td>
</tr>
<tr id="row013914719136"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1313927191317"><a name="p1313927191317"></a><a name="p1313927191317"></a>Status</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8139177171315"><a name="p8139177171315"></a><a name="p8139177171315"></a>芯片状态</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul8436530113"></a><a name="ul8436530113"></a><ul id="ul8436530113"><li>unrecovered：未恢复</li><li>recovered：恢复成功</li><li>failed：恢复失败</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11394791316"><a name="p11394791316"></a><a name="p11394791316"></a>-</p>
</td>
</tr>
<tr id="row814016761315"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1814015719134"><a name="p1814015719134"></a><a name="p1814015719134"></a>Policy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p11140676132"><a name="p11140676132"></a><a name="p11140676132"></a>热复位策略</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul1156918243817"></a><a name="ul1156918243817"></a><ul id="ul1156918243817"><li>empty：无故障</li><li>ignore：忽略故障</li><li>restart_request：重新执行当前请求</li><li>restart：重新执行训练任务</li><li>free_reset：当NPU上没有任务时，需要重启设备</li><li>reset：需要重启设备</li><li>isolate：需要隔离设备</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11140672134"><a name="p11140672134"></a><a name="p11140672134"></a>-</p>
</td>
</tr>
<tr id="row151401717139"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p101413711136"><a name="p101413711136"></a><a name="p101413711136"></a>InitialPolicy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12141176132"><a name="p12141176132"></a><a name="p12141176132"></a>初始热复位策略</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><a name="ul1918372817111"></a><a name="ul1918372817111"></a><ul id="ul1918372817111"><li>empty：无故障</li><li>ignore：忽略故障</li><li>restart_request：重新执行当前请求</li><li>restart：重新执行训练任务</li><li>free_reset：当NPU上没有任务时，需要重启设备</li><li>reset：需要重启设备</li><li>isolate：需要隔离设备</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p171419712133"><a name="p171419712133"></a><a name="p171419712133"></a>-</p>
</td>
</tr>
<tr id="row2141187121312"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3141197161312"><a name="p3141197161312"></a><a name="p3141197161312"></a>ErrorCode</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19141576138"><a name="p19141576138"></a><a name="p19141576138"></a>十进制故障码</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p151429710139"><a name="p151429710139"></a><a name="p151429710139"></a>64位整型数组</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1314257131311"><a name="p1314257131311"></a><a name="p1314257131311"></a>-</p>
</td>
</tr>
<tr id="row1721555113913"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p414625113526"><a name="p414625113526"></a><a name="p414625113526"></a>GracefulExit</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p321511543920"><a name="p321511543920"></a><a name="p321511543920"></a>管理训练进程</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p33363655012"><a name="p33363655012"></a><a name="p33363655012"></a>0或1</p>
<a name="ul7532185975011"></a><a name="ul7532185975011"></a><ul id="ul7532185975011"><li>取值为1，杀死所有训练进程</li><li>取值为0，不做处理</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p921615511390"><a name="p921615511390"></a><a name="p921615511390"></a>-</p>
</td>
</tr>
<tr id="row45409251618"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1254115251666"><a name="p1254115251666"></a><a name="p1254115251666"></a>FaultFlushing</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p7541192512618"><a name="p7541192512618"></a><a name="p7541192512618"></a>告知<span id="ph14256162281217"><a name="ph14256162281217"></a><a name="ph14256162281217"></a>Elastic Agent</span>当前是否有故障正在刷新</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p13813147101216"><a name="p13813147101216"></a><a name="p13813147101216"></a>取值为true或false</p>
<a name="ul1563191521213"></a><a name="ul1563191521213"></a><ul id="ul1563191521213"><li>true：表示有故障正在刷新</li><li>false表示当前无故障刷新</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p19951631131314"><a name="p19951631131314"></a><a name="p19951631131314"></a><span id="ph952618296564"><a name="ph952618296564"></a><a name="ph952618296564"></a>Elastic Agent</span>需要等待该字段为false且故障RankList无本节点故障时才会拉起训练进程</p>
</td>
</tr>
<tr id="row141375594377"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p64521951162319"><a name="p64521951162319"></a><a name="p64521951162319"></a><span>RestartFaultProcess</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17453851172311"><a name="p17453851172311"></a><a name="p17453851172311"></a><span>告知</span><span id="ph262783362516"><a name="ph262783362516"></a><a name="ph262783362516"></a>Elastic Agent</span><span>当前是否仅重启本节点故障进程</span></p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p2012431813258"><a name="p2012431813258"></a><a name="p2012431813258"></a><span>取值true或false</span></p>
<a name="ul14729113619013"></a><a name="ul14729113619013"></a><ul id="ul14729113619013"><li><span>true：当本节点有故障时，仅重启本节点故障进程</span></li><li><span>false：当本节点有故障时，退出本节点所有进程且退出</span><span id="ph205211257104"><a name="ph205211257104"></a><a name="ph205211257104"></a>Elastic Agent</span></li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p94534513233"><a name="p94534513233"></a><a name="p94534513233"></a>当故障RankList有本节点故障时此字段才生效</p>
</td>
</tr>
<tr id="row14142137191314"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p171421973132"><a name="p171421973132"></a><a name="p171421973132"></a>ErrorCodeHex</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p0142577133"><a name="p0142577133"></a><a name="p0142577133"></a>十六进制故障码</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p8142177131320"><a name="p8142177131320"></a><a name="p8142177131320"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p31421070133"><a name="p31421070133"></a><a name="p31421070133"></a>-</p>
</td>
</tr>
<tr id="row1209179172911"><td class="cellrowborder" valign="top" width="13.940000000000003%" headers="mcps1.2.6.1.1 "><p id="p844941513297"><a name="p844941513297"></a><a name="p844941513297"></a>restartType</p>
</td>
<td class="cellrowborder" valign="top" width="16.700000000000003%" headers="mcps1.2.6.1.2 "><p id="p220916992912"><a name="p220916992912"></a><a name="p220916992912"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="22.28%" headers="mcps1.2.6.1.3 "><p id="p1820909182911"><a name="p1820909182911"></a><a name="p1820909182911"></a>reset.json更新的类型</p>
</td>
<td class="cellrowborder" valign="top" width="31.220000000000002%" headers="mcps1.2.6.1.4 "><p id="p15209596295"><a name="p15209596295"></a><a name="p15209596295"></a>podReschedule或hotReset</p>
</td>
<td class="cellrowborder" valign="top" width="15.860000000000001%" headers="mcps1.2.6.1.5 "><p id="p95471047133013"><a name="p95471047133013"></a><a name="p95471047133013"></a>单Pod重调度情况下取值为podReschedule，热恢复场景下取值为hotReset</p>
</td>
</tr>
</tbody>
</table>

**mindx-dl/job-reschedule-reason<a name="section20866121155814"></a>**

该ConfigMap用于记录任务重调度历史信息，默认情况下会保存任务最近的十次重调度记录，当ConfigMap内容超过950Kb时会依次删减每个任务中发生时间最早的记录。

**表 4**  任务字段说明

<a name="table589619361579"></a>
<table><thead align="left"><tr id="row15897183618711"><th class="cellrowborder" valign="top" width="7.5200000000000005%" id="mcps1.2.6.1.1"><p id="p1389711365712"><a name="p1389711365712"></a><a name="p1389711365712"></a>字段名称</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.2"><p id="p20897836479"><a name="p20897836479"></a><a name="p20897836479"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="24.740000000000002%" id="mcps1.2.6.1.3"><p id="p15897113614715"><a name="p15897113614715"></a><a name="p15897113614715"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="21.5%" id="mcps1.2.6.1.4"><p id="p889712361771"><a name="p889712361771"></a><a name="p889712361771"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.5"><p id="p12897153610715"><a name="p12897153610715"></a><a name="p12897153610715"></a>备注</p>
</th>
</tr>
</thead>
<tbody><tr id="row289716363716"><td class="cellrowborder" rowspan="4" valign="top" width="7.5200000000000005%" headers="mcps1.2.6.1.1 "><p id="p6386733753"><a name="p6386733753"></a><a name="p6386733753"></a>任务ns/任务名</p>
<p id="p489813612714"><a name="p489813612714"></a><a name="p489813612714"></a></p>
<p id="p148981361575"><a name="p148981361575"></a><a name="p148981361575"></a></p>
<p id="p389883620715"><a name="p389883620715"></a><a name="p389883620715"></a></p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.2 "><p id="p114390307520"><a name="p114390307520"></a><a name="p114390307520"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="24.740000000000002%" headers="mcps1.2.6.1.3 "><p id="p178971736978"><a name="p178971736978"></a><a name="p178971736978"></a>标记执行重调度的任务名称。</p>
</td>
<td class="cellrowborder" valign="top" width="21.5%" headers="mcps1.2.6.1.4 "><p id="p88981936971"><a name="p88981936971"></a><a name="p88981936971"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.5 "><p id="p13898436378"><a name="p13898436378"></a><a name="p13898436378"></a>-</p>
</td>
</tr>
<tr id="row19898636879"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1489810360719"><a name="p1489810360719"></a><a name="p1489810360719"></a>JobID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p98985368719"><a name="p98985368719"></a><a name="p98985368719"></a>任务ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p18898736674"><a name="p18898736674"></a><a name="p18898736674"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p389803617718"><a name="p389803617718"></a><a name="p389803617718"></a>-</p>
</td>
</tr>
<tr id="row48980368719"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p188988361372"><a name="p188988361372"></a><a name="p188988361372"></a>TotalRescheduleTimes</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p1489812361714"><a name="p1489812361714"></a><a name="p1489812361714"></a>该任务在<span id="ph525518226126"><a name="ph525518226126"></a><a name="ph525518226126"></a>Volcano</span>本次生命周期内记录的重调度总次数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p289810362715"><a name="p289810362715"></a><a name="p289810362715"></a>整数</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p989820364719"><a name="p989820364719"></a><a name="p989820364719"></a>-</p>
</td>
</tr>
<tr id="row19898163611716"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p109891552520"><a name="p109891552520"></a><a name="p109891552520"></a>RescheduleRecords</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p138991936178"><a name="p138991936178"></a><a name="p138991936178"></a>记录本任务重调度的具体信息。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p32675459545"><a name="p32675459545"></a><a name="p32675459545"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p178998369716"><a name="p178998369716"></a><a name="p178998369716"></a>-</p>
</td>
</tr>
</tbody>
</table>

**表 5**  RescheduleRecords说明

<a name="table1578964348"></a>
<table><thead align="left"><tr id="row4327416646"><th class="cellrowborder" valign="top" width="7.5200000000000005%" id="mcps1.2.6.1.1"><p id="p112821723049"><a name="p112821723049"></a><a name="p112821723049"></a>字段名称</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.2"><p id="p16282823942"><a name="p16282823942"></a><a name="p16282823942"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="24.740000000000002%" id="mcps1.2.6.1.3"><p id="p4282723240"><a name="p4282723240"></a><a name="p4282723240"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="21.5%" id="mcps1.2.6.1.4"><p id="p328212231943"><a name="p328212231943"></a><a name="p328212231943"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.5"><p id="p102821623248"><a name="p102821623248"></a><a name="p102821623248"></a>备注</p>
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
<td class="cellrowborder" valign="top" width="24.740000000000002%" headers="mcps1.2.6.1.3 "><p id="p37891541546"><a name="p37891541546"></a><a name="p37891541546"></a>按<span id="ph769722034511"><a name="ph769722034511"></a><a name="ph769722034511"></a>Volcano</span>日志格式记录的重调度时间</p>
</td>
<td class="cellrowborder" valign="top" width="21.5%" headers="mcps1.2.6.1.4 "><p id="p13790104440"><a name="p13790104440"></a><a name="p13790104440"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.5 "><p id="p67904410410"><a name="p67904410410"></a><a name="p67904410410"></a>-</p>
</td>
</tr>
<tr id="row2790043411"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1540831717595"><a name="p1540831717595"></a><a name="p1540831717595"></a>RescheduleTimeStamp</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p117908411411"><a name="p117908411411"></a><a name="p117908411411"></a>重调度发生的时间戳</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p379064641"><a name="p379064641"></a><a name="p379064641"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p3790144148"><a name="p3790144148"></a><a name="p3790144148"></a>-</p>
</td>
</tr>
<tr id="row8790941340"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p17859751501"><a name="p17859751501"></a><a name="p17859751501"></a>ReasonOfTask</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p198239143588"><a name="p198239143588"></a><a name="p198239143588"></a>记录本次重调度的具体信息。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p666117055520"><a name="p666117055520"></a><a name="p666117055520"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p12817714165811"><a name="p12817714165811"></a><a name="p12817714165811"></a>-</p>
</td>
</tr>
</tbody>
</table>

**表 6**  ReasonOfTask说明

<a name="table8680019155817"></a>
<table><thead align="left"><tr id="row165113075818"><th class="cellrowborder" valign="top" width="7.5200000000000005%" id="mcps1.2.6.1.1"><p id="p25183025820"><a name="p25183025820"></a><a name="p25183025820"></a>字段名称</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.2"><p id="p0514308584"><a name="p0514308584"></a><a name="p0514308584"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="24.740000000000002%" id="mcps1.2.6.1.3"><p id="p175193045819"><a name="p175193045819"></a><a name="p175193045819"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="21.5%" id="mcps1.2.6.1.4"><p id="p1051103025812"><a name="p1051103025812"></a><a name="p1051103025812"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.5"><p id="p10511530195815"><a name="p10511530195815"></a><a name="p10511530195815"></a>备注</p>
</th>
</tr>
</thead>
<tbody><tr id="row1568121995819"><td class="cellrowborder" rowspan="4" valign="top" width="7.5200000000000005%" headers="mcps1.2.6.1.1 "><p id="p1868191919589"><a name="p1868191919589"></a><a name="p1868191919589"></a>ReasonOfTask</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.2 "><p id="p17681171915582"><a name="p17681171915582"></a><a name="p17681171915582"></a>RescheduleReason</p>
</td>
<td class="cellrowborder" valign="top" width="24.740000000000002%" headers="mcps1.2.6.1.3 "><p id="p1968114193584"><a name="p1968114193584"></a><a name="p1968114193584"></a>重调度原因</p>
</td>
<td class="cellrowborder" valign="top" width="21.5%" headers="mcps1.2.6.1.4 "><p id="p58011850228"><a name="p58011850228"></a><a name="p58011850228"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.5 "><p id="p1968191912589"><a name="p1968191912589"></a><a name="p1968191912589"></a>-</p>
</td>
</tr>
<tr id="row9681171915810"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1681191915581"><a name="p1681191915581"></a><a name="p1681191915581"></a>PodName</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p668219192585"><a name="p668219192585"></a><a name="p668219192585"></a>本次重调度首先触发的pod</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p186822019105812"><a name="p186822019105812"></a><a name="p186822019105812"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p11682131910582"><a name="p11682131910582"></a><a name="p11682131910582"></a>-</p>
</td>
</tr>
<tr id="row76821198589"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p17682171915812"><a name="p17682171915812"></a><a name="p17682171915812"></a>NodeName</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p26821819125817"><a name="p26821819125817"></a><a name="p26821819125817"></a>节点名称</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p19682201917586"><a name="p19682201917586"></a><a name="p19682201917586"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p186825190582"><a name="p186825190582"></a><a name="p186825190582"></a>本次重调度首先触发的node。</p>
</td>
</tr>
<tr id="row10682151985811"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p8682919115816"><a name="p8682919115816"></a><a name="p8682919115816"></a>NodeRankIndex</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p19682191985813"><a name="p19682191985813"></a><a name="p19682191985813"></a>本次重调度首先触发的node在训练中所属rank</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p11682191995818"><a name="p11682191995818"></a><a name="p11682191995818"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1269842159"><a name="p1269842159"></a><a name="p1269842159"></a>-</p>
</td>
</tr>
</tbody>
</table>


## 参数面网络拓扑配置<a name="ZH-CN_TOPIC_0000002479386820"></a>

**basic-tor-node-cm<a name="section18148132883914"></a>**

**表 1**  basic-tor-node-cm

<a name="table18901255141213"></a>
<table><thead align="left"><tr id="row11911955131210"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p4911455191214"><a name="p4911455191214"></a><a name="p4911455191214"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="26.75%" id="mcps1.2.5.1.2"><p id="p119105512124"><a name="p119105512124"></a><a name="p119105512124"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="23.25%" id="mcps1.2.5.1.3"><p id="p291205519126"><a name="p291205519126"></a><a name="p291205519126"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p4911955171219"><a name="p4911955171219"></a><a name="p4911955171219"></a>备注</p>
</th>
</tr>
</thead>
<tbody><tr id="row392125551217"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p892655181213"><a name="p892655181213"></a><a name="p892655181213"></a>version</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p8921055191217"><a name="p8921055191217"></a><a name="p8921055191217"></a>basic-tor-node-cm的版本</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1392455101212"><a name="p1392455101212"></a><a name="p1392455101212"></a>1.0</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1392125581216"><a name="p1392125581216"></a><a name="p1392125581216"></a>-</p>
</td>
</tr>
<tr id="row1921555151218"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p292355131211"><a name="p292355131211"></a><a name="p292355131211"></a>tor_count</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p2931155131219"><a name="p2931155131219"></a><a name="p2931155131219"></a>集群中交换机下的节点数</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p119395551211"><a name="p119395551211"></a><a name="p119395551211"></a>整数类型</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p3931455181211"><a name="p3931455181211"></a><a name="p3931455181211"></a>-</p>
</td>
</tr>
<tr id="row293165511123"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1693115512128"><a name="p1693115512128"></a><a name="p1693115512128"></a>server_list</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p893855191217"><a name="p893855191217"></a><a name="p893855191217"></a>集群节点按交换机为单位划分的集合</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p17931955191211"><a name="p17931955191211"></a><a name="p17931955191211"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p169417559127"><a name="p169417559127"></a><a name="p169417559127"></a>-</p>
</td>
</tr>
<tr id="row139425511215"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8941755121217"><a name="p8941755121217"></a><a name="p8941755121217"></a>- tor_id</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p7943559126"><a name="p7943559126"></a><a name="p7943559126"></a>交换机的序号</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p109412553124"><a name="p109412553124"></a><a name="p109412553124"></a>整数类型</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p4940552122"><a name="p4940552122"></a><a name="p4940552122"></a>-</p>
</td>
</tr>
<tr id="row1194165581219"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1095195511125"><a name="p1095195511125"></a><a name="p1095195511125"></a>- tor_ip</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p19953555121"><a name="p19953555121"></a><a name="p19953555121"></a>交换机的IP地址</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p99515520126"><a name="p99515520126"></a><a name="p99515520126"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p129535541220"><a name="p129535541220"></a><a name="p129535541220"></a>-</p>
</td>
</tr>
<tr id="row189575510122"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p18955559121"><a name="p18955559121"></a><a name="p18955559121"></a>server</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p595255171210"><a name="p595255171210"></a><a name="p595255171210"></a>交换机下的节点信息</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1796185511219"><a name="p1796185511219"></a><a name="p1796185511219"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p15971155131217"><a name="p15971155131217"></a><a name="p15971155131217"></a>-</p>
</td>
</tr>
<tr id="row199718557121"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1697555161219"><a name="p1697555161219"></a><a name="p1697555161219"></a>- server_ip</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p0986557122"><a name="p0986557122"></a><a name="p0986557122"></a>节点的IP地址</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p16983555129"><a name="p16983555129"></a><a name="p16983555129"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10981755141218"><a name="p10981755141218"></a><a name="p10981755141218"></a>-</p>
</td>
</tr>
<tr id="row89805591220"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p109815561219"><a name="p109815561219"></a><a name="p109815561219"></a>- npu_count</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p799355141210"><a name="p799355141210"></a><a name="p799355141210"></a>节点上NPU芯片的数量</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1399455161212"><a name="p1399455161212"></a><a name="p1399455161212"></a>整数类型</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p699145517127"><a name="p699145517127"></a><a name="p699145517127"></a>-</p>
</td>
</tr>
<tr id="row12993553125"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p9100155518126"><a name="p9100155518126"></a><a name="p9100155518126"></a>- slice_id</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p151001155141214"><a name="p151001155141214"></a><a name="p151001155141214"></a>节点在交换机下的编号</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p19100655111211"><a name="p19100655111211"></a><a name="p19100655111211"></a>整数类型</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p131011555171218"><a name="p131011555171218"></a><a name="p131011555171218"></a>-</p>
</td>
</tr>
</tbody>
</table>


## Volcano调度器配置<a name="ZH-CN_TOPIC_0000002511346767"></a>

**volcano-scheduler-configmap<a name="section42181344193715"></a>**

**表 1**  volcano-scheduler-configmap字段说明

<a name="table1864354211112"></a>
<table><thead align="left"><tr id="row464317427117"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p10644942131117"><a name="p10644942131117"></a><a name="p10644942131117"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="26.75%" id="mcps1.2.5.1.2"><p id="p96441642191113"><a name="p96441642191113"></a><a name="p96441642191113"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="23.25%" id="mcps1.2.5.1.3"><p id="p14644124281118"><a name="p14644124281118"></a><a name="p14644124281118"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p196441342201112"><a name="p196441342201112"></a><a name="p196441342201112"></a>备注</p>
</th>
</tr>
</thead>
<tbody><tr id="row964444217112"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1664520421116"><a name="p1664520421116"></a><a name="p1664520421116"></a>actions</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p764518425111"><a name="p764518425111"></a><a name="p764518425111"></a>调度流程使用的动作</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p8645942191119"><a name="p8645942191119"></a><a name="p8645942191119"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p2645154210118"><a name="p2645154210118"></a><a name="p2645154210118"></a>ascend-volcano-plugin使用了enqueue、allocate、backfill三个调度动作</p>
</td>
</tr>
<tr id="row5645184251114"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p14646114221110"><a name="p14646114221110"></a><a name="p14646114221110"></a>plugins</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p064613424117"><a name="p064613424117"></a><a name="p064613424117"></a>调度流程使用的插件集合</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p16646194251117"><a name="p16646194251117"></a><a name="p16646194251117"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p664610425116"><a name="p664610425116"></a><a name="p664610425116"></a>-</p>
</td>
</tr>
<tr id="row16460421119"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p15647174213117"><a name="p15647174213117"></a><a name="p15647174213117"></a>- name</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p17647174251112"><a name="p17647174251112"></a><a name="p17647174251112"></a>使用的插件名称</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p164784221110"><a name="p164784221110"></a><a name="p164784221110"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1364764219113"><a name="p1364764219113"></a><a name="p1364764219113"></a>ascend-volcano-plugin使用了</p>
<p id="p6647144211119"><a name="p6647144211119"></a><a name="p6647144211119"></a>priority、gang、conformance、volcano-npu_<em id="i14647134215110"><a name="i14647134215110"></a><a name="i14647134215110"></a>{version}</em>_linux-<em id="i9647144211116"><a name="i9647144211116"></a><a name="i9647144211116"></a>{arch}</em>、drf、predicates、proportion、nodeorder、binpack几种调度插件</p>
</td>
</tr>
<tr id="row11647174219115"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p96481942141116"><a name="p96481942141116"></a><a name="p96481942141116"></a>configurations</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p964816428110"><a name="p964816428110"></a><a name="p964816428110"></a>调度器初始化的配置信息</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p10648042161117"><a name="p10648042161117"></a><a name="p10648042161117"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1364812426115"><a name="p1364812426115"></a><a name="p1364812426115"></a>-</p>
</td>
</tr>
<tr id="row16648134251119"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1664816426116"><a name="p1664816426116"></a><a name="p1664816426116"></a>- name</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p14648144261120"><a name="p14648144261120"></a><a name="p14648144261120"></a>配置信息名称</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p18649542171116"><a name="p18649542171116"></a><a name="p18649542171116"></a>init-params</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p1264914219119"><a name="p1264914219119"></a><a name="p1264914219119"></a>-</p>
</td>
</tr>
<tr id="row136491042141119"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p164914420116"><a name="p164914420116"></a><a name="p164914420116"></a>- arguments</p>
</td>
<td class="cellrowborder" valign="top" width="26.75%" headers="mcps1.2.5.1.2 "><p id="p17649174210112"><a name="p17649174210112"></a><a name="p17649174210112"></a>配置信息内容</p>
</td>
<td class="cellrowborder" valign="top" width="23.25%" headers="mcps1.2.5.1.3 "><p id="p1564964215115"><a name="p1564964215115"></a><a name="p1564964215115"></a>键值对集合</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p96491242191116"><a name="p96491242191116"></a><a name="p96491242191116"></a>-</p>
</td>
</tr>
</tbody>
</table>


