# Ascend Operator<a name="ZH-CN_TOPIC_0000002511346797"></a>

**YAML参数说明（acjob任务）<a name="section1660111420312"></a>**

如果是acjob任务，在配置YAML前，请先了解相关YAML参数说明，详细说明如下表所示。

每个acjob任务YAML中包含一些固定字段，例如apiVersion、kind等，如果想了解这些字段的详细说明请参见[acjob关键字段说明](../appendix.md#acjob关键字段说明)。

**表 1**  YAML参数说明

<a name="table7602101418317"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001951418201_row1460212146313"><th class="cellrowborder" valign="top" width="27.18%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001951418201_p196029147318"><a name="zh-cn_topic_0000001951418201_p196029147318"></a><a name="zh-cn_topic_0000001951418201_p196029147318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="36.26%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001951418201_p1560213143314"><a name="zh-cn_topic_0000001951418201_p1560213143314"></a><a name="zh-cn_topic_0000001951418201_p1560213143314"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001951418201_p106023141317"><a name="zh-cn_topic_0000001951418201_p106023141317"></a><a name="zh-cn_topic_0000001951418201_p106023141317"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001951418201_row260211141136"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1660311140313"><a name="zh-cn_topic_0000001951418201_p1660311140313"></a><a name="zh-cn_topic_0000001951418201_p1660311140313"></a>framework</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="ul4975113512712"></a><a name="ul4975113512712"></a><ul id="ul4975113512712"><li>mindspore</li><li>pytorch</li><li>tensorflow</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p4389131101318"><a name="p4389131101318"></a><a name="p4389131101318"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row10436102842510"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p2436102814254"><a name="zh-cn_topic_0000001951418201_p2436102814254"></a><a name="zh-cn_topic_0000001951418201_p2436102814254"></a>jobID</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="p1619843317517"><a name="p1619843317517"></a><a name="p1619843317517"></a>当前MindIE Motor任务在集群中的唯一识别ID，用户可根据实际情况进行配置。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1146401944917"><a name="p1146401944917"></a><a name="p1146401944917"></a>该参数仅支持在<span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span>、<span id="ph2472821203013"><a name="ph2472821203013"></a><a name="ph2472821203013"></a>Atlas 800I A3 超节点服务器</span>上使用。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row16523123316254"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p13524833182513"><a name="zh-cn_topic_0000001951418201_p13524833182513"></a><a name="zh-cn_topic_0000001951418201_p13524833182513"></a>app</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p5524103317257"><a name="zh-cn_topic_0000001951418201_p5524103317257"></a><a name="zh-cn_topic_0000001951418201_p5524103317257"></a>表示当前MindIE Motor在Ascend Job任务中的角色，取值包括mindie-ms-controller、mindie-ms-coordinator、mindie-ms-server。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><div class="note" id="zh-cn_topic_0000001951418201_note4367125713295"><a name="zh-cn_topic_0000001951418201_note4367125713295"></a><div class="notebody"><a name="ul139591420161415"></a><a name="ul139591420161415"></a><ul id="ul139591420161415"><li>acjob的任务YAML同时包含jobID和app这2个字段时，<span id="zh-cn_topic_0000001951418201_ph1566531814589"><a name="zh-cn_topic_0000001951418201_ph1566531814589"></a><a name="zh-cn_topic_0000001951418201_ph1566531814589"></a>Ascend Operator</span>组件会自动传入环境变量MINDX_TASK_ID、APP_TYPE及MINDX_SERVICE_IP，并将其标识为MindIE推理任务。</li><li>关于以上环境变量的详细说明请参见<a href="../appendix.md#环境变量说明">环境变量说明</a>中"Ascend Operator注入的训练环境变量"表。</li><li>该参数仅支持在<span id="ph249120134413"><a name="ph249120134413"></a><a name="ph249120134413"></a>Atlas 800I A2 推理服务器</span>、<span id="ph2790182618303"><a name="ph2790182618303"></a><a name="ph2790182618303"></a>Atlas 800I A3 超节点服务器</span>上使用。</li></ul>
</div></div>
</td>
</tr>
<tr id="row1541745918171"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="p143891219113814"><a name="p143891219113814"></a><a name="p143891219113814"></a><span>mind-cluster/scaling-rule: scaling-rule</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="p3389151918383"><a name="p3389151918383"></a><a name="p3389151918383"></a>标记扩缩容规则对应的ConfigMap名称。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p15266582437"><a name="p15266582437"></a><a name="p15266582437"></a>仅支持MindIE Motor推理任务在<span id="ph2026614834316"><a name="ph2026614834316"></a><a name="ph2026614834316"></a>Atlas 800I A2 推理服务器</span>、<span id="ph1434121484214"><a name="ph1434121484214"></a><a name="ph1434121484214"></a>Atlas 800I A3 超节点服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="row172898336182"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="p16388101913817"><a name="p16388101913817"></a><a name="p16388101913817"></a><span>mind-cluster/group-name: group0</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="p13387171983812"><a name="p13387171983812"></a><a name="p13387171983812"></a>标记扩缩容规则中对应的group名称。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p5144163418438"><a name="p5144163418438"></a><a name="p5144163418438"></a>仅支持MindIE Motor推理任务在<span id="ph1114333494317"><a name="ph1114333494317"></a><a name="ph1114333494317"></a>Atlas 800I A2 推理服务器</span>、<span id="ph12417527144218"><a name="ph12417527144218"></a><a name="ph12417527144218"></a>Atlas 800I A3 超节点服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="row1996920561501"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="p20969356175010"><a name="p20969356175010"></a><a name="p20969356175010"></a>podAffinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="p196965635015"><a name="p196965635015"></a><a name="p196965635015"></a>表示逻辑超节点会往具有更多亲和性Pod的物理超节点调度。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p3969155611509"><a name="p3969155611509"></a><a name="p3969155611509"></a>仅支持MindIE Motor推理任务<span id="ph249517547298"><a name="ph249517547298"></a><a name="ph249517547298"></a>Atlas 800I A3 超节点服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="row22810219519"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="p128102185119"><a name="p128102185119"></a><a name="p128102185119"></a>sp-fit</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="p72820245118"><a name="p72820245118"></a><a name="p72820245118"></a>超节点调度策略。</p>
<a name="ul86791134125215"></a><a name="ul86791134125215"></a><ul id="ul86791134125215"><li>idlest：逻辑超节点会往更空闲的物理超节点调度。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p755841425314"><a name="p755841425314"></a><a name="p755841425314"></a>仅支持MindIE Motor推理任务<span id="ph1858015143594"><a name="ph1858015143594"></a><a name="ph1858015143594"></a>Atlas 800I A3 超节点服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1060320149314"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1860317149314"><a name="zh-cn_topic_0000001951418201_p1860317149314"></a><a name="zh-cn_topic_0000001951418201_p1860317149314"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="ul16230203215710"></a><a name="ul16230203215710"></a><ul id="ul16230203215710"><li><span id="ph20976435102713"><a name="ph20976435102713"></a><a name="ph20976435102713"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>、<span id="ph163483412215"><a name="ph163483412215"></a><a name="ph163483412215"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph136651315478"><a name="ph136651315478"></a><a name="ph136651315478"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span>、<span id="ph12174764117"><a name="ph12174764117"></a><a name="ph12174764117"></a>Atlas 800I A3 超节点服务器</span>取值为：ascend-<span id="ph11976935122715"><a name="ph11976935122715"></a><a name="ph11976935122715"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li><li>Atlas 800 训练服务器，服务器（插<span id="ph2099203201811"><a name="ph2099203201811"></a><a name="ph2099203201811"></a>Atlas 300T 训练卡</span>）取值为：ascend-910</li><li>Atlas A5 系列产品、Atlas 350 标卡、Atlas 850 服务器、Atlas 950 SuperPod 超节点</span>取值为：huawei.com/npu</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p43811639112614"><a name="zh-cn_topic_0000001951418201_p43811639112614"></a><a name="zh-cn_topic_0000001951418201_p43811639112614"></a>标识任务使用的芯片的产品类型。</p>
<p id="zh-cn_topic_0000001951418201_p4409148135"><a name="zh-cn_topic_0000001951418201_p4409148135"></a><a name="zh-cn_topic_0000001951418201_p4409148135"></a>需要在<span id="zh-cn_topic_0000001951418201_ph7409748837"><a name="zh-cn_topic_0000001951418201_ph7409748837"></a><a name="zh-cn_topic_0000001951418201_ph7409748837"></a>ConfigMap</span>和任务task中配置。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1960421417318"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p460415143318"><a name="zh-cn_topic_0000001951418201_p460415143318"></a><a name="zh-cn_topic_0000001951418201_p460415143318"></a>schedulerName</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p56045145317"><a name="zh-cn_topic_0000001951418201_p56045145317"></a><a name="zh-cn_topic_0000001951418201_p56045145317"></a>默认值为<span class="parmvalue" id="zh-cn_topic_0000001951418201_parmvalue10604111417319"><a name="zh-cn_topic_0000001951418201_parmvalue10604111417319"></a><a name="zh-cn_topic_0000001951418201_parmvalue10604111417319"></a>“volcano”</span>，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p46041014430"><a name="zh-cn_topic_0000001951418201_p46041014430"></a><a name="zh-cn_topic_0000001951418201_p46041014430"></a><span id="zh-cn_topic_0000001951418201_ph6604131419312"><a name="zh-cn_topic_0000001951418201_ph6604131419312"></a><a name="zh-cn_topic_0000001951418201_ph6604131419312"></a>Ascend Operator</span>启用“gang”调度时所选择的调度器。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row19604714936"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p46047147315"><a name="zh-cn_topic_0000001951418201_p46047147315"></a><a name="zh-cn_topic_0000001951418201_p46047147315"></a>minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p106041114132"><a name="zh-cn_topic_0000001951418201_p106041114132"></a><a name="zh-cn_topic_0000001951418201_p106041114132"></a>默认值为任务总副本数</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p660420141935"><a name="zh-cn_topic_0000001951418201_p660420141935"></a><a name="zh-cn_topic_0000001951418201_p660420141935"></a><span id="zh-cn_topic_0000001951418201_ph16604181418316"><a name="zh-cn_topic_0000001951418201_ph16604181418316"></a><a name="zh-cn_topic_0000001951418201_ph16604181418316"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="zh-cn_topic_0000001951418201_ph156050141033"><a name="zh-cn_topic_0000001951418201_ph156050141033"></a><a name="zh-cn_topic_0000001951418201_ph156050141033"></a>Volcano</span>时，任务运行总副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row86054141139"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p16051014839"><a name="zh-cn_topic_0000001951418201_p16051014839"></a><a name="zh-cn_topic_0000001951418201_p16051014839"></a>queue</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1760581413313"><a name="zh-cn_topic_0000001951418201_p1760581413313"></a><a name="zh-cn_topic_0000001951418201_p1760581413313"></a>默认值为<span class="parmvalue" id="zh-cn_topic_0000001951418201_parmvalue1260516142036"><a name="zh-cn_topic_0000001951418201_parmvalue1260516142036"></a><a name="zh-cn_topic_0000001951418201_parmvalue1260516142036"></a>“default”</span>，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p13605414637"><a name="zh-cn_topic_0000001951418201_p13605414637"></a><a name="zh-cn_topic_0000001951418201_p13605414637"></a><span id="zh-cn_topic_0000001951418201_ph10605114231"><a name="zh-cn_topic_0000001951418201_ph10605114231"></a><a name="zh-cn_topic_0000001951418201_ph10605114231"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="zh-cn_topic_0000001951418201_ph1660520141632"><a name="zh-cn_topic_0000001951418201_ph1660520141632"></a><a name="zh-cn_topic_0000001951418201_ph1660520141632"></a>Volcano</span>时，任务所属队列。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row18605114739"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p136051014737"><a name="zh-cn_topic_0000001951418201_p136051014737"></a><a name="zh-cn_topic_0000001951418201_p136051014737"></a>（可选）successPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul6605121420317"></a><a name="zh-cn_topic_0000001951418201_ul6605121420317"></a><ul id="zh-cn_topic_0000001951418201_ul6605121420317"><li>默认值为空，若用户不填写该参数，则默认取空值。</li><li>AllWorkers</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p126052141730"><a name="zh-cn_topic_0000001951418201_p126052141730"></a><a name="zh-cn_topic_0000001951418201_p126052141730"></a>表明任务成功的前提。空值代表只需要一个<span id="zh-cn_topic_0000001951418201_ph46053142034"><a name="zh-cn_topic_0000001951418201_ph46053142034"></a><a name="zh-cn_topic_0000001951418201_ph46053142034"></a>Pod</span>成功，整个任务判定为成功。取值为<span class="parmvalue" id="zh-cn_topic_0000001951418201_parmvalue460514143318"><a name="zh-cn_topic_0000001951418201_parmvalue460514143318"></a><a name="zh-cn_topic_0000001951418201_parmvalue460514143318"></a>“AllWorkers”</span>表示所有<span id="zh-cn_topic_0000001951418201_ph46065141434"><a name="zh-cn_topic_0000001951418201_ph46065141434"></a><a name="zh-cn_topic_0000001951418201_ph46065141434"></a>Pod</span>都成功，任务才判定为成功。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row560612147315"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p560613141833"><a name="zh-cn_topic_0000001951418201_p560613141833"></a><a name="zh-cn_topic_0000001951418201_p560613141833"></a>container.name</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1360671419316"><a name="zh-cn_topic_0000001951418201_p1360671419316"></a><a name="zh-cn_topic_0000001951418201_p1360671419316"></a>ascend</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p860691418319"><a name="zh-cn_topic_0000001951418201_p860691418319"></a><a name="zh-cn_topic_0000001951418201_p860691418319"></a>训练容器的名称必须是<span class="parmvalue" id="zh-cn_topic_0000001951418201_parmvalue1260612147320"><a name="zh-cn_topic_0000001951418201_parmvalue1260612147320"></a><a name="zh-cn_topic_0000001951418201_parmvalue1260612147320"></a>“ascend”</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row116068141134"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p5606151413318"><a name="zh-cn_topic_0000001951418201_p5606151413318"></a><a name="zh-cn_topic_0000001951418201_p5606151413318"></a>（可选）ports</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p176065144312"><a name="zh-cn_topic_0000001951418201_p176065144312"></a><a name="zh-cn_topic_0000001951418201_p176065144312"></a>若用户未进行设置，系统默认填写以下参数：</p>
<a name="zh-cn_topic_0000001951418201_ul106061214438"></a><a name="zh-cn_topic_0000001951418201_ul106061214438"></a><ul id="zh-cn_topic_0000001951418201_ul106061214438"><li>name：ascendjob-port</li><li>containerPort：2222</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p11607414537"><a name="zh-cn_topic_0000001951418201_p11607414537"></a><a name="zh-cn_topic_0000001951418201_p11607414537"></a>分布式训练集合通讯端口。<span class="parmname" id="zh-cn_topic_0000001951418201_parmname160711141237"><a name="zh-cn_topic_0000001951418201_parmname160711141237"></a><a name="zh-cn_topic_0000001951418201_parmname160711141237"></a>“containerPort”</span>用户可根据实际情况设置，若未进行设置则采用默认端口2222。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row3607151417314"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p560721419320"><a name="zh-cn_topic_0000001951418201_p560721419320"></a><a name="zh-cn_topic_0000001951418201_p560721419320"></a>replicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul156070141730"></a><a name="zh-cn_topic_0000001951418201_ul156070141730"></a><ul id="zh-cn_topic_0000001951418201_ul156070141730"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p196071814834"><a name="zh-cn_topic_0000001951418201_p196071814834"></a><a name="zh-cn_topic_0000001951418201_p196071814834"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row86071144316"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p660791413315"><a name="zh-cn_topic_0000001951418201_p660791413315"></a><a name="zh-cn_topic_0000001951418201_p660791413315"></a>image</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1060731418315"><a name="zh-cn_topic_0000001951418201_p1060731418315"></a><a name="zh-cn_topic_0000001951418201_p1060731418315"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p560811141335"><a name="zh-cn_topic_0000001951418201_p560811141335"></a><a name="zh-cn_topic_0000001951418201_p560811141335"></a>训练镜像名称，请根据实际修改。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row260820141037"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1860814141536"><a name="zh-cn_topic_0000001951418201_p1860814141536"></a><a name="zh-cn_topic_0000001951418201_p1860814141536"></a>（可选）host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p56089142319"><a name="zh-cn_topic_0000001951418201_p56089142319"></a><a name="zh-cn_topic_0000001951418201_p56089142319"></a><span id="zh-cn_topic_0000001951418201_ph5608814330"><a name="zh-cn_topic_0000001951418201_ph5608814330"></a><a name="zh-cn_topic_0000001951418201_ph5608814330"></a>Arm</span>环境：<span id="ph27942615713"><a name="ph27942615713"></a><a name="ph27942615713"></a>huawei-arm</span></p>
<p id="zh-cn_topic_0000001951418201_p1960819141639"><a name="zh-cn_topic_0000001951418201_p1960819141639"></a><a name="zh-cn_topic_0000001951418201_p1960819141639"></a><span id="zh-cn_topic_0000001951418201_ph186088141531"><a name="zh-cn_topic_0000001951418201_ph186088141531"></a><a name="zh-cn_topic_0000001951418201_ph186088141531"></a>x86_64</span>环境：<span id="ph27919313716"><a name="ph27919313716"></a><a name="ph27919313716"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1060801414315"><a name="zh-cn_topic_0000001951418201_p1060801414315"></a><a name="zh-cn_topic_0000001951418201_p1060801414315"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000001951418201_p76084142313"><a name="zh-cn_topic_0000001951418201_p76084142313"></a><a name="zh-cn_topic_0000001951418201_p76084142313"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="row1191123318297"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="p1377484145110"><a name="p1377484145110"></a><a name="p1377484145110"></a><span id="ph19817116185120"><a name="ph19817116185120"></a><a name="ph19817116185120"></a>huawei.com/schedule_policy</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="p1877410425111"><a name="p1877410425111"></a><a name="p1877410425111"></a>目前支持<a href="#table1120511613153">表3</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p153031739509"><a name="p153031739509"></a><a name="p153031739509"></a>配置任务需要调度的AI芯片布局形态。<span id="zh-cn_topic_0000002511347099_ph204811934163414"><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
<div class="note" id="note1230363125010"><a name="note1230363125010"></a><a name="note1230363125010"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002511347099_p1767434372512"><a name="zh-cn_topic_0000002511347099_p1767434372512"></a><a name="zh-cn_topic_0000002511347099_p1767434372512"></a>仅支持在<span id="ph1331492318423"><a name="ph1331492318423"></a><a name="ph1331492318423"></a>Atlas 训练系列产品</span>、<span id="ph2314323124211"><a name="ph2314323124211"></a><a name="ph2314323124211"></a><term id="zh-cn_topic_0000001519959665_term57208119917_1"><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a>Atlas A2 训练系列产品</term></span>、<span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span>、<span id="ph996833614580"><a name="ph996833614580"></a><a name="ph996833614580"></a><term id="zh-cn_topic_0000001094307702_term99602034117"><a name="zh-cn_topic_0000001094307702_term99602034117"></a><a name="zh-cn_topic_0000001094307702_term99602034117"></a>Atlas A5 训练系列产品</term></span>、<span id="ph996833614580"><a name="ph996833614580"></a><a name="ph996833614580"></a><term id="zh-cn_topic_0000001094307702_term99602034117"><a name="zh-cn_topic_0000001094307702_term99602034117"></a><a name="zh-cn_topic_0000001094307702_term99602034117"></a>Atlas A2 推理系列产品</term></span>、<span id="ph791742714211"><a name="ph791742714211"></a><a name="ph791742714211"></a><term id="zh-cn_topic_0000001519959665_term176419491615"><a name="zh-cn_topic_0000001519959665_term176419491615"></a><a name="zh-cn_topic_0000001519959665_term176419491615"></a>Atlas A3 推理系列产品</term></span>、<span id="ph791742714211"><a name="ph791742714211"></a><a name="ph791742714211"></a><term id="zh-cn_topic_0000001519959665_term176419491615"><a name="zh-cn_topic_0000001519959665_term176419491615"></a><a name="zh-cn_topic_0000001519959665_term176419491615"></a>Atlas A5 推理系列产品</term></span>中使用该字段。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row56081214237"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1960818141031"><a name="zh-cn_topic_0000001951418201_p1960818141031"></a><a name="zh-cn_topic_0000001951418201_p1960818141031"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p14755536454"><a name="zh-cn_topic_0000002039339953_p14755536454"></a><a name="zh-cn_topic_0000002039339953_p14755536454"></a>指定逻辑超节点芯片数量。</p>
<a name="zh-cn_topic_0000002039339953_ul10451144414619"></a><a name="zh-cn_topic_0000002039339953_ul10451144414619"></a><ul id="zh-cn_topic_0000002039339953_ul10451144414619"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p1670155202912"><a name="p1670155202912"></a><a name="p1670155202912"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>若用户未指定该字段，</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p19701652112917"><a name="p19701652112917"></a><a name="p19701652112917"></a>了解详细说明请参见<a href="../references.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note47015215291"><a name="note47015215291"></a><a name="note47015215291"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569"><li>仅支持在<span id="zh-cn_topic_0000002511347099_ph34244153594"><a name="zh-cn_topic_0000002511347099_ph34244153594"></a><a name="zh-cn_topic_0000002511347099_ph34244153594"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="../faq.md#任务申请的总芯片数量为32sp-block设置为32可以正常训练sp-block设置为16无法完成训练训练容器报错提示初始化连接失败">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row760915145311"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p06096144316"><a name="zh-cn_topic_0000001951418201_p06096144316"></a><a name="zh-cn_topic_0000001951418201_p06096144316"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul206095141139"></a><a name="zh-cn_topic_0000001951418201_ul206095141139"></a><ul id="zh-cn_topic_0000001951418201_ul206095141139"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note" id="zh-cn_topic_0000001951418201_note66098141039"><a name="zh-cn_topic_0000001951418201_note66098141039"></a><a name="zh-cn_topic_0000001951418201_note66098141039"></a><span class="notetitle"> [!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p8609121419312"><a name="zh-cn_topic_0000001951418201_p8609121419312"></a><a name="zh-cn_topic_0000001951418201_p8609121419312"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p660912140310"><a name="zh-cn_topic_0000001951418201_p660912140310"></a><a name="zh-cn_topic_0000001951418201_p660912140310"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note16091141837"><a name="zh-cn_topic_0000001951418201_note16091141837"></a><a name="zh-cn_topic_0000001951418201_note16091141837"></a><span class="notetitle">[!NOTE] 说明 </span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul176092014030"></a><a name="zh-cn_topic_0000001951418201_ul176092014030"></a><ul id="zh-cn_topic_0000001951418201_ul176092014030"><li>交换机亲和性调度1.0版本支持<span id="zh-cn_topic_0000001951418201_ph1157665817140"><a name="zh-cn_topic_0000001951418201_ph1157665817140"></a><a name="zh-cn_topic_0000001951418201_ph1157665817140"></a>Atlas 训练系列产品</span>和<span id="zh-cn_topic_0000001951418201_ph168598363399"><a name="zh-cn_topic_0000001951418201_ph168598363399"></a><a name="zh-cn_topic_0000001951418201_ph168598363399"></a>Atlas A2 训练系列产品</span>；支持<span id="zh-cn_topic_0000001951418201_ph4181625925"><a name="zh-cn_topic_0000001951418201_ph4181625925"></a><a name="zh-cn_topic_0000001951418201_ph4181625925"></a>PyTorch</span>和<span id="zh-cn_topic_0000001951418201_ph61882510210"><a name="zh-cn_topic_0000001951418201_ph61882510210"></a><a name="zh-cn_topic_0000001951418201_ph61882510210"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="zh-cn_topic_0000001951418201_ph311717506401"><a name="zh-cn_topic_0000001951418201_ph311717506401"></a><a name="zh-cn_topic_0000001951418201_ph311717506401"></a>Atlas A2 训练系列产品</span>；支持<span id="zh-cn_topic_0000001951418201_ph17383182419412"><a name="zh-cn_topic_0000001951418201_ph17383182419412"></a><a name="zh-cn_topic_0000001951418201_ph17383182419412"></a>PyTorch</span>框架。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row46101144312"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1861010140316"><a name="zh-cn_topic_0000001951418201_p1861010140316"></a><a name="zh-cn_topic_0000001951418201_p1861010140316"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul186101614131"></a><a name="zh-cn_topic_0000001951418201_ul186101614131"></a><ul id="zh-cn_topic_0000001951418201_ul186101614131"><li>on：开启Pod级别重调度</li><li>其他值或不使用该字段：关闭Pod级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p661016141437"><a name="zh-cn_topic_0000001951418201_p661016141437"></a><a name="zh-cn_topic_0000001951418201_p661016141437"></a>Pod级别重调度，表示任务发生故障后，不会删除所有任务Pod，而是将发生故障的Pod进行删除，重新创建新Pod后进行重调度。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note1561010145316"><a name="zh-cn_topic_0000001951418201_note1561010145316"></a><a name="zh-cn_topic_0000001951418201_note1561010145316"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul461013147314"></a><a name="zh-cn_topic_0000001951418201_ul461013147314"></a><ul id="zh-cn_topic_0000001951418201_ul461013147314"><li>重调度模式默认为任务级重调度，若需要开启Pod级别重调度，需要新增该字段。</li><li><span id="zh-cn_topic_0000001951418201_ph1061091414318"><a name="zh-cn_topic_0000001951418201_ph1061091414318"></a><a name="zh-cn_topic_0000001951418201_ph1061091414318"></a>TensorFlow</span>暂不支持Pod级别重调度。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row5888113641512"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p10888436121514"><a name="zh-cn_topic_0000001951418201_p10888436121514"></a><a name="zh-cn_topic_0000001951418201_p10888436121514"></a>recover-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p186621822171620"><a name="zh-cn_topic_0000001951418201_p186621822171620"></a><a name="zh-cn_topic_0000001951418201_p186621822171620"></a>任务可用恢复策略。</p>
<a name="zh-cn_topic_0000001951418201_ul20208182771618"></a><a name="zh-cn_topic_0000001951418201_ul20208182771618"></a><ul id="zh-cn_topic_0000001951418201_ul20208182771618"><li>retry：进程级在线恢复。</li><li>recover：进程级别重调度。</li><li>recover-in-place：进程级原地恢复。</li><li>elastic-training：弹性训练。</li><li>dump：保存临终遗言。</li><li>exit：退出训练</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000001951418201_ul18941121318614"></a><a name="zh-cn_topic_0000001951418201_ul18941121318614"></a>recover-strategy配置在任务YAML annotations下，取值为6种策略的随意组合，策略之间由逗号分割。
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row114562211115"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1399351045011"><a name="zh-cn_topic_0000001951418201_p1399351045011"></a><a name="zh-cn_topic_0000001951418201_p1399351045011"></a>process-recover-enable</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul71592205015"></a><a name="zh-cn_topic_0000001951418201_ul71592205015"></a><ul id="zh-cn_topic_0000001951418201_ul71592205015"><li>on：开启进程级别重调度及进程级在线恢复。<p>进程级别重调度和优雅容错不能同时开启，若同时开启，断点续训将通过job级重调度恢复训练。</p>
</li><li>pause：暂时关闭进程级别重调度及进程级在线恢复。</li><li>off或不使用该字段：关闭进程级别重调度及进程级在线恢复。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p10523132473119"><a name="zh-cn_topic_0000001951418201_p10523132473119"></a><a name="zh-cn_topic_0000001951418201_p10523132473119"></a>Ascend Operator会根据用户配置的recover-strategy自动给任务打上process-recover-enable=on标签，无需用户手动指定。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row205285218207"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p95281922209"><a name="zh-cn_topic_0000001951418201_p95281922209"></a><a name="zh-cn_topic_0000001951418201_p95281922209"></a>subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul18716519102210"></a><a name="zh-cn_topic_0000001951418201_ul18716519102210"></a><ul id="zh-cn_topic_0000001951418201_ul18716519102210"><li>ignore：忽略该亚健康节点，后续任务在亲和性调度上不优先调度该节点。</li><li>graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。</li><li>forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。</li><li>hotSwitch：执行亚健康热切，拉起备份Pod后，暂停训练任务，并使用新节点重新拉起训练。</li><li>默认取值为ignore。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p3641163291319"><a name="zh-cn_topic_0000001951418201_p3641163291319"></a><a name="zh-cn_topic_0000001951418201_p3641163291319"></a>节点状态为亚健康（SubHealthy）的节点的处理策略。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note635023119333"><a name="zh-cn_topic_0000001951418201_note635023119333"></a><a name="zh-cn_topic_0000001951418201_note635023119333"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002246684301_ul24832019017"></a><a name="zh-cn_topic_0000002246684301_ul24832019017"></a><ul id="zh-cn_topic_0000002246684301_ul24832019017"><li>使用graceExit策略时，需保证训练框架能够接收SIGTERM信号并保存CKPT文件。</li><li>hotSwitch策略的使用约束请参见<a href="../usage/resumable_training.md#亚健康热切">使用约束</a>。</li><li>Verl强化学习任务只支持ignore和forceExit。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row36114148312"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1461116146318"><a name="zh-cn_topic_0000001951418201_p1461116146318"></a><a name="zh-cn_topic_0000001951418201_p1461116146318"></a>accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul461118141037"></a><a name="zh-cn_topic_0000001951418201_ul461118141037"></a><ul id="zh-cn_topic_0000001951418201_ul461118141037"><li><span id="zh-cn_topic_0000001951418201_ph136117141331"><a name="zh-cn_topic_0000001951418201_ph136117141331"></a><a name="zh-cn_topic_0000001951418201_ph136117141331"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li><li><span id="zh-cn_topic_0000001951418201_ph26111143315"><a name="zh-cn_topic_0000001951418201_ph26111143315"></a><a name="zh-cn_topic_0000001951418201_ph26111143315"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li><li>服务器（插<span id="zh-cn_topic_0000001951418201_ph26117141237"><a name="zh-cn_topic_0000001951418201_ph26117141237"></a><a name="zh-cn_topic_0000001951418201_ph26117141237"></a>Atlas 300T 训练卡</span>）：card</li><li><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000001951418201_ph061115145320"><a name="zh-cn_topic_0000001951418201_ph061115145320"></a><a name="zh-cn_topic_0000001951418201_ph061115145320"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="zh-cn_topic_0000001951418201_ph1661112141731"><a name="zh-cn_topic_0000001951418201_ph1661112141731"></a><a name="zh-cn_topic_0000001951418201_ph1661112141731"></a><em id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b-8</li><li><span id="zh-cn_topic_0000001951418201_ph1161214145319"><a name="zh-cn_topic_0000001951418201_ph1161214145319"></a><a name="zh-cn_topic_0000001951418201_ph1161214145319"></a>Atlas 200T A2 Box16 异构子框</span><span id="ph172491011154612"><a name="ph172491011154612"></a><a name="ph172491011154612"></a>和</span><span id="ph10949202261219"><a name="ph10949202261219"></a><a name="ph10949202261219"></a>Atlas 200I A2 Box16 异构子框</span>：module-<span id="zh-cn_topic_0000001951418201_ph116121514934"><a name="zh-cn_topic_0000001951418201_ph116121514934"></a><a name="zh-cn_topic_0000001951418201_ph116121514934"></a><em id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b-16</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 350 标卡（可选）</span>：350-Atlas-8、350-Atlas-16、350-Atlas-4p-8、350-Atlas-4p-16</li><li>Atlas 850 服务器（可选）：850-Atlas-8p-8、850-SuperPod-Atlas-8</li><li>Atlas 950 PoD 超节点（可选）：950-SuperPod-Atlas-8</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p6612914039"><a name="zh-cn_topic_0000001951418201_p6612914039"></a><a name="zh-cn_topic_0000001951418201_p6612914039"></a>根据需要运行训练任务的节点类型，选取不同的值。如果节点是<span id="zh-cn_topic_0000001951418201_ph1961291412314"><a name="zh-cn_topic_0000001951418201_ph1961291412314"></a><a name="zh-cn_topic_0000001951418201_ph1961291412314"></a>Atlas 800 训练服务器（NPU满配）</span>，可以省略该标签。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note1861316141738"><a name="zh-cn_topic_0000001951418201_note1861316141738"></a><a name="zh-cn_topic_0000001951418201_note1861316141738"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p1027616512420"><a name="zh-cn_topic_0000001951418201_p1027616512420"></a><a name="zh-cn_topic_0000001951418201_p1027616512420"></a><span id="zh-cn_topic_0000001951418201_ph9014016509"><a name="zh-cn_topic_0000001951418201_ph9014016509"></a><a name="zh-cn_topic_0000001951418201_ph9014016509"></a>芯片型号的数值可通过<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_b168254314713"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_b168254314713"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_b168254314713"></a>npu-smi info</strong>命令查询，返回的“Name”字段对应信息为芯片型号，下文的{<em id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1914312018209"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1914312018209"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1914312018209"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row18613714439"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p196131140315"><a name="zh-cn_topic_0000001951418201_p196131140315"></a><a name="zh-cn_topic_0000001951418201_p196131140315"></a>huawei.com/Ascend910</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><div class="p" id="zh-cn_topic_0000001951418201_p361318141733"><a name="zh-cn_topic_0000001951418201_p361318141733"></a><a name="zh-cn_topic_0000001951418201_p361318141733"></a><span id="zh-cn_topic_0000001951418201_ph106131514134"><a name="zh-cn_topic_0000001951418201_ph106131514134"></a><a name="zh-cn_topic_0000001951418201_ph106131514134"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000001951418201_ul126147145311"></a><a name="zh-cn_topic_0000001951418201_ul126147145311"></a><ul id="zh-cn_topic_0000001951418201_ul126147145311"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、8</li><li>分布式任务：1、2、4、8</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000001951418201_p3614151416317"><a name="zh-cn_topic_0000001951418201_p3614151416317"></a><a name="zh-cn_topic_0000001951418201_p3614151416317"></a><span id="zh-cn_topic_0000001951418201_ph261416147313"><a name="zh-cn_topic_0000001951418201_ph261416147313"></a><a name="zh-cn_topic_0000001951418201_ph261416147313"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="zh-cn_topic_0000001951418201_ul1961418141132"></a><a name="zh-cn_topic_0000001951418201_ul1961418141132"></a><ul id="zh-cn_topic_0000001951418201_ul1961418141132"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4</li><li>分布式任务：1、2、4</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000001951418201_p10614214538"><a name="zh-cn_topic_0000001951418201_p10614214538"></a><a name="zh-cn_topic_0000001951418201_p10614214538"></a>服务器（插<span id="zh-cn_topic_0000001951418201_ph13615131417315"><a name="zh-cn_topic_0000001951418201_ph13615131417315"></a><a name="zh-cn_topic_0000001951418201_ph13615131417315"></a>Atlas 300T 训练卡</span>）：<a name="zh-cn_topic_0000001951418201_ul1261519142311"></a><a name="zh-cn_topic_0000001951418201_ul1261519142311"></a><ul id="zh-cn_topic_0000001951418201_ul1261519142311"><li>单机单芯片任务：1</li><li>单机多芯片任务：2</li><li>分布式任务：2</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000001951418201_p11615161416311"><a name="zh-cn_topic_0000001951418201_p11615161416311"></a><a name="zh-cn_topic_0000001951418201_p11615161416311"></a><span id="ph422512673816"><a name="ph422512673816"></a><a name="ph422512673816"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000001951418201_ph46156141634"><a name="zh-cn_topic_0000001951418201_ph46156141634"></a><a name="zh-cn_topic_0000001951418201_ph46156141634"></a>Atlas 900 A2 PoD 集群基础单元</span>：<a name="zh-cn_topic_0000001951418201_ul1961514143314"></a><a name="zh-cn_topic_0000001951418201_ul1961514143314"></a><ul id="zh-cn_topic_0000001951418201_ul1961514143314"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8</li><li>分布式任务：1、2、3、4、5、6、7、8</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000001951418201_p15616514739"><a name="zh-cn_topic_0000001951418201_p15616514739"></a><a name="zh-cn_topic_0000001951418201_p15616514739"></a><span id="zh-cn_topic_0000001951418201_ph161611419319"><a name="zh-cn_topic_0000001951418201_ph161611419319"></a><a name="zh-cn_topic_0000001951418201_ph161611419319"></a>Atlas 200T A2 Box16 异构子框</span><span id="ph14474722144613"><a name="ph14474722144613"></a><a name="ph14474722144613"></a>和</span><span id="ph05982314466"><a name="ph05982314466"></a><a name="ph05982314466"></a>Atlas 200I A2 Box16 异构子框</span>：<a name="zh-cn_topic_0000001951418201_ul661611418316"></a><a name="zh-cn_topic_0000001951418201_ul661611418316"></a><ul id="zh-cn_topic_0000001951418201_ul661611418316"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式任务：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000001951418201_p136171314938"><a name="zh-cn_topic_0000001951418201_p136171314938"></a><a name="zh-cn_topic_0000001951418201_p136171314938"></a><span id="ph747840144217"><a name="ph747840144217"></a><a name="ph747840144217"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph1573918472719"><a name="ph1573918472719"></a><a name="ph1573918472719"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph1173994172715"><a name="ph1173994172715"></a><a name="ph1173994172715"></a>Atlas 800T A3 超节点服务器</span>：<a name="zh-cn_topic_0000001951418201_ul261751412316"></a><a name="zh-cn_topic_0000001951418201_ul261751412316"></a><ul id="zh-cn_topic_0000001951418201_ul261751412316"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、6、8、10、12、14、16</li><li>分布式任务：2、4、6、8、10、12、14、16</li><li>针对<span id="ph1474218267273"><a name="ph1474218267273"></a><a name="ph1474218267273"></a>Atlas 900 A3 SuperPoD 超节点</span>的逻辑超节点亲和任务：16</li></ul>
</div>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p561713141331"><a name="zh-cn_topic_0000001951418201_p561713141331"></a><a name="zh-cn_topic_0000001951418201_p561713141331"></a>请求的NPU数量，请根据实际修改。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row11621414533"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p894317013244"><a name="zh-cn_topic_0000001951418201_p894317013244"></a><a name="zh-cn_topic_0000001951418201_p894317013244"></a>(.kind=="AscendJob").spec.replicaSpecs.[Master|Scheduler|Worker].template.spec.containers[0].env[name==ASCEND_VISIBLE_DEVICES].valueFrom.fieldRef.fieldPath</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p7622914235"><a name="zh-cn_topic_0000001951418201_p7622914235"></a><a name="zh-cn_topic_0000001951418201_p7622914235"></a>取值为metadata.annotations['huawei.com/AscendXXX']，其中XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p136226142031"><a name="zh-cn_topic_0000001951418201_p136226142031"></a><a name="zh-cn_topic_0000001951418201_p136226142031"></a><span id="zh-cn_topic_0000001951418201_ph1062212140315"><a name="zh-cn_topic_0000001951418201_ph1062212140315"></a><a name="zh-cn_topic_0000001951418201_ph1062212140315"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note462214141730"><a name="zh-cn_topic_0000001951418201_note462214141730"></a><a name="zh-cn_topic_0000001951418201_note462214141730"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p186225141637"><a name="zh-cn_topic_0000001951418201_p186225141637"></a><a name="zh-cn_topic_0000001951418201_p186225141637"></a>该参数只支持使用<span id="zh-cn_topic_0000001951418201_ph962251412315"><a name="zh-cn_topic_0000001951418201_ph962251412315"></a><a name="zh-cn_topic_0000001951418201_ph962251412315"></a>Volcano</span>调度器的整卡调度特性，使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row662216141939"><td class="cellrowborder" rowspan="5" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p106221514533"><a name="zh-cn_topic_0000001951418201_p106221514533"></a><a name="zh-cn_topic_0000001951418201_p106221514533"></a>fault-scheduling</p>
<p id="zh-cn_topic_0000001951418201_p176223141233"><a name="zh-cn_topic_0000001951418201_p176223141233"></a><a name="zh-cn_topic_0000001951418201_p176223141233"></a></p>
<p id="zh-cn_topic_0000001951418201_p106222141235"><a name="zh-cn_topic_0000001951418201_p106222141235"></a><a name="zh-cn_topic_0000001951418201_p106222141235"></a></p>
<p id="zh-cn_topic_0000001951418201_p17622171414319"><a name="zh-cn_topic_0000001951418201_p17622171414319"></a><a name="zh-cn_topic_0000001951418201_p17622171414319"></a></p>
<p id="zh-cn_topic_0000001951418201_p86225141938"><a name="zh-cn_topic_0000001951418201_p86225141938"></a><a name="zh-cn_topic_0000001951418201_p86225141938"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p206221814637"><a name="zh-cn_topic_0000001951418201_p206221814637"></a><a name="zh-cn_topic_0000001951418201_p206221814637"></a>grace</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1462216142314"><a name="zh-cn_topic_0000001951418201_p1462216142314"></a><a name="zh-cn_topic_0000001951418201_p1462216142314"></a>配置任务采用优雅删除模式，并在过程中先优雅删除原<span id="zh-cn_topic_0000001951418201_ph19623131417313"><a name="zh-cn_topic_0000001951418201_ph19623131417313"></a><a name="zh-cn_topic_0000001951418201_ph19623131417313"></a>Pod</span>，15分钟后若还未成功，使用强制删除原<span id="zh-cn_topic_0000001951418201_ph96231114734"><a name="zh-cn_topic_0000001951418201_ph96231114734"></a><a name="zh-cn_topic_0000001951418201_ph96231114734"></a>Pod</span>。</p>
<p id="p1462216142314"><a name="p1462216142314"></a><a name="p1462216142314"></a>进程级别重调度和进程级在线恢复场景，需将本参数配置为grace。</p>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1262313144314"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p762371419310"><a name="zh-cn_topic_0000001951418201_p762371419310"></a><a name="zh-cn_topic_0000001951418201_p762371419310"></a>force</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p562301420319"><a name="zh-cn_topic_0000001951418201_p562301420319"></a><a name="zh-cn_topic_0000001951418201_p562301420319"></a>配置任务采用强制删除模式，在过程中强制删除原<span id="zh-cn_topic_0000001951418201_ph19623151420318"><a name="zh-cn_topic_0000001951418201_ph19623151420318"></a><a name="zh-cn_topic_0000001951418201_ph19623151420318"></a>Pod</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row46230146312"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p126231144312"><a name="zh-cn_topic_0000001951418201_p126231144312"></a><a name="zh-cn_topic_0000001951418201_p126231144312"></a>off</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p26233141317"><a name="zh-cn_topic_0000001951418201_p26233141317"></a><a name="zh-cn_topic_0000001951418201_p26233141317"></a>该任务不使用断点续训特性，<span id="zh-cn_topic_0000001951418201_ph8623191418313"><a name="zh-cn_topic_0000001951418201_ph8623191418313"></a><a name="zh-cn_topic_0000001951418201_ph8623191418313"></a>K8s</span>的maxRetry仍然生效。</p>
<p id="zh-cn_topic_0000001951418201_p186239141631"><a name="zh-cn_topic_0000001951418201_p186239141631"></a><a name="zh-cn_topic_0000001951418201_p186239141631"></a></p>
<p id="zh-cn_topic_0000001951418201_p1623191419310"><a name="zh-cn_topic_0000001951418201_p1623191419310"></a><a name="zh-cn_topic_0000001951418201_p1623191419310"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row7623191419310"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p11624181414311"><a name="zh-cn_topic_0000001951418201_p11624181414311"></a><a name="zh-cn_topic_0000001951418201_p11624181414311"></a>无（无fault-scheduling字段）</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row106241614036"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p9624191420310"><a name="zh-cn_topic_0000001951418201_p9624191420310"></a><a name="zh-cn_topic_0000001951418201_p9624191420310"></a>其他值</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row76241014637"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p262451412319"><a name="zh-cn_topic_0000001951418201_p262451412319"></a><a name="zh-cn_topic_0000001951418201_p262451412319"></a>fault-retry-times</p>
<p id="zh-cn_topic_0000001951418201_p96241714134"><a name="zh-cn_topic_0000001951418201_p96241714134"></a><a name="zh-cn_topic_0000001951418201_p96241714134"></a></p>
<p id="zh-cn_topic_0000001951418201_p662413146319"><a name="zh-cn_topic_0000001951418201_p662413146319"></a><a name="zh-cn_topic_0000001951418201_p662413146319"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p76249141830"><a name="zh-cn_topic_0000001951418201_p76249141830"></a><a name="zh-cn_topic_0000001951418201_p76249141830"></a>0 &lt; fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p9624101418314"><a name="zh-cn_topic_0000001951418201_p9624101418314"></a><a name="zh-cn_topic_0000001951418201_p9624101418314"></a>处理业务面故障，必须配置业务面无条件重试的次数。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note136241514039"><a name="zh-cn_topic_0000001951418201_note136241514039"></a><a name="zh-cn_topic_0000001951418201_note136241514039"></a><span class="notetitle"> 说明： </span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul13624161415314"></a><a name="zh-cn_topic_0000001951418201_ul13624161415314"></a><ul id="zh-cn_topic_0000001951418201_ul13624161415314"><li>使用无条件重试功能需保证训练进程异常时会导致容器异常退出，若容器未异常退出则无法成功重试。</li><li>当前仅<span id="ph19925121012"><a name="ph19925121012"></a><a name="ph19925121012"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000001951418201_ph166251314730"><a name="zh-cn_topic_0000001951418201_ph166251314730"></a><a name="zh-cn_topic_0000001951418201_ph166251314730"></a>Atlas 900 A2 PoD 集群基础单元</span>支持无条件重试功能。</li><li>进行进程级恢复时，将会触发业务面故障，如需使用进程级恢复，必须配置此参数。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row06256141536"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1462511141837"><a name="zh-cn_topic_0000001951418201_p1462511141837"></a><a name="zh-cn_topic_0000001951418201_p1462511141837"></a>无（无fault-retry-times）或0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p12625914632"><a name="zh-cn_topic_0000001951418201_p12625914632"></a><a name="zh-cn_topic_0000001951418201_p12625914632"></a>该任务不使用无条件重试功能，无法感知业务面故障，vcjob的maxRetry仍然生效。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row146252141832"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p17625214933"><a name="zh-cn_topic_0000001951418201_p17625214933"></a><a name="zh-cn_topic_0000001951418201_p17625214933"></a>backoffLimit</p>
<p id="zh-cn_topic_0000001951418201_p106261014332"><a name="zh-cn_topic_0000001951418201_p106261014332"></a><a name="zh-cn_topic_0000001951418201_p106261014332"></a></p>
<p id="zh-cn_topic_0000001951418201_p146262147310"><a name="zh-cn_topic_0000001951418201_p146262147310"></a><a name="zh-cn_topic_0000001951418201_p146262147310"></a></p>
<p id="zh-cn_topic_0000001951418201_p1962631414310"><a name="zh-cn_topic_0000001951418201_p1962631414310"></a><a name="zh-cn_topic_0000001951418201_p1962631414310"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p66263141230"><a name="zh-cn_topic_0000001951418201_p66263141230"></a><a name="zh-cn_topic_0000001951418201_p66263141230"></a>0 &lt; backoffLimit</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p18626614532"><a name="zh-cn_topic_0000001951418201_p18626614532"></a><a name="zh-cn_topic_0000001951418201_p18626614532"></a>任务重调度次数。任务故障时，可以重调度的次数，当已经重调度次数与backoffLimit取值相同时，任务将不再进行重调度。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note15626214934"><a name="zh-cn_topic_0000001951418201_note15626214934"></a><a name="zh-cn_topic_0000001951418201_note15626214934"></a><span class="notetitle"> [!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p362641413318"><a name="zh-cn_topic_0000001951418201_p362641413318"></a><a name="zh-cn_topic_0000001951418201_p362641413318"></a>同时配置了backoffLimit和fault-retry-times参数时，当已经重调度次数与backoffLimit或fault-retry-times取值有一个相同时，将不再进行重调度。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row662614145317"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1062612142315"><a name="zh-cn_topic_0000001951418201_p1062612142315"></a><a name="zh-cn_topic_0000001951418201_p1062612142315"></a>无（无backoffLimit）或backoffLimit ≤ 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p46267141139"><a name="zh-cn_topic_0000001951418201_p46267141139"></a><a name="zh-cn_topic_0000001951418201_p46267141139"></a>不限制总重调度次数。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note7627191419311"><a name="zh-cn_topic_0000001951418201_note7627191419311"></a><a name="zh-cn_topic_0000001951418201_note7627191419311"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p962712140314"><a name="zh-cn_topic_0000001951418201_p962712140314"></a><a name="zh-cn_topic_0000001951418201_p962712140314"></a>若不配置backoffLimit，但是配置了fault-retry-times参数，则使用fault-retry-times的重调度次数。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1662711144310"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p462713149315"><a name="zh-cn_topic_0000001951418201_p462713149315"></a><a name="zh-cn_topic_0000001951418201_p462713149315"></a>restartPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul36271614531"></a><a name="zh-cn_topic_0000001951418201_ul36271614531"></a><ul id="zh-cn_topic_0000001951418201_ul36271614531"><li>Never：从不重启</li><li>Always：总是重启</li><li>OnFailure：失败时重启</li><li>ExitCode：根据进程退出码决定是否重启Pod，错误码是1~127时不重启，128~255时重启Pod。<div class="note" id="zh-cn_topic_0000001951418201_note156280141037"><a name="zh-cn_topic_0000001951418201_note156280141037"></a><a name="zh-cn_topic_0000001951418201_note156280141037"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p17628131415316"><a name="zh-cn_topic_0000001951418201_p17628131415316"></a><a name="zh-cn_topic_0000001951418201_p17628131415316"></a>vcjob类型的训练任务不支持ExitCode。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p762813148318"><a name="zh-cn_topic_0000001951418201_p762813148318"></a><a name="zh-cn_topic_0000001951418201_p762813148318"></a>容器重启策略。当配置业务面故障无条件重试时，容器重启策略取值必须为<span class="parmvalue" id="zh-cn_topic_0000001951418201_parmvalue18628191413311"><a name="zh-cn_topic_0000001951418201_parmvalue18628191413311"></a><a name="zh-cn_topic_0000001951418201_parmvalue18628191413311"></a>“Never”</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row14628131414314"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1362811415312"><a name="zh-cn_topic_0000001951418201_p1362811415312"></a><a name="zh-cn_topic_0000001951418201_p1362811415312"></a>terminationGracePeriodSeconds</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1862811412316"><a name="zh-cn_topic_0000001951418201_p1862811412316"></a><a name="zh-cn_topic_0000001951418201_p1862811412316"></a>0 &lt; terminationGracePeriodSeconds &lt; <strong id="zh-cn_topic_0000001951418201_b1962881410316"><a name="zh-cn_topic_0000001951418201_b1962881410316"></a><a name="zh-cn_topic_0000001951418201_b1962881410316"></a>grace-over-time</strong>参数取值</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p4628914935"><a name="zh-cn_topic_0000001951418201_p4628914935"></a><a name="zh-cn_topic_0000001951418201_p4628914935"></a>容器收到SIGTERM到被<span id="zh-cn_topic_0000001951418201_ph176283140314"><a name="zh-cn_topic_0000001951418201_ph176283140314"></a><a name="zh-cn_topic_0000001951418201_ph176283140314"></a>K8s</span>强制停止经历的时间，该时间需要大于0且小于volcano-v<em id="zh-cn_topic_0000001951418201_i1562819141131"><a name="zh-cn_topic_0000001951418201_i1562819141131"></a><a name="zh-cn_topic_0000001951418201_i1562819141131"></a>{version}</em>.yaml文件中“<strong id="zh-cn_topic_0000001951418201_b10628161415313"><a name="zh-cn_topic_0000001951418201_b10628161415313"></a><a name="zh-cn_topic_0000001951418201_b10628161415313"></a>grace-over-time</strong>”参数取值，同时还需要保证能够保存CKPT文件，请根据实际情况修改。具体说明请参考<span id="zh-cn_topic_0000001951418201_ph136288141334"><a name="zh-cn_topic_0000001951418201_ph136288141334"></a><a name="zh-cn_topic_0000001951418201_ph136288141334"></a>K8s</span>官网<a href="https://kubernetes.io/zh/docs/concepts/containers/container-lifecycle-hooks/" target="_blank" rel="noopener noreferrer">容器生命周期回调</a>。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note146291714533"><a name="zh-cn_topic_0000001951418201_note146291714533"></a><a name="zh-cn_topic_0000001951418201_note146291714533"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p1962991416310"><a name="zh-cn_topic_0000001951418201_p1962991416310"></a><a name="zh-cn_topic_0000001951418201_p1962991416310"></a>只有当fault-scheduling配置为grace时，该字段才生效；fault-scheduling配置为force时，该字段无效。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row962814644010"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1062816462407"><a name="zh-cn_topic_0000001951418201_p1062816462407"></a><a name="zh-cn_topic_0000001951418201_p1062816462407"></a>hostNetwork</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul960434424111"></a><a name="zh-cn_topic_0000001951418201_ul960434424111"></a><ul id="zh-cn_topic_0000001951418201_ul960434424111"><li>true：使用HostIP创建Pod。</li><li>false：不使用HostIP创建Pod。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000001951418201_ul14611159182815"></a><a name="zh-cn_topic_0000001951418201_ul14611159182815"></a><ul id="zh-cn_topic_0000001951418201_ul14611159182815"><li>当集群规模较大（节点数量&gt;1000时），推荐使用HostIP创建Pod。</li><li>不传入此参数时，默认不使用HostIP创建Pod。<div class="note" id="zh-cn_topic_0000001951418201_note1423653119592"><a name="zh-cn_topic_0000001951418201_note1423653119592"></a><a name="zh-cn_topic_0000001951418201_note1423653119592"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p461933317584"><a name="zh-cn_topic_0000001951418201_p461933317584"></a><a name="zh-cn_topic_0000001951418201_p461933317584"></a>当HostNetwork取值为true时，若当前任务YAML挂载了RankTable文件路径，则可以通过在训练脚本中解析RankTable文件获取Pod的hostIP来实现建链。若任务YAML未挂载RankTable文件路径，则与原始保持一致，使用serviceIP来实现建链。</p>
</div></div>
</li></ul>
</td>
</tr>
</tbody>
</table>

**YAML参数说明（deploy任务或vcjob任务）<a name="section18126121013410"></a>**

**表 2**  YAML参数说明

<a name="zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_table1565872494511"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_row1465822412450"><th class="cellrowborder" valign="top" width="27.18%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p13658124194513"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p13658124194513"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p13658124194513"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="36.26%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p4658152420459"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p4658152420459"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p4658152420459"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="36.559999999999995%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p8302202619484"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p8302202619484"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p8302202619484"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_row8658102464518"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p19658152414451"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p19658152414451"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p19658152414451"></a>minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1531417539259"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1531417539259"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1531417539259"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p11302326164814"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p11302326164814"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p11302326164814"></a>N为节点个数，Deployment类型的任务不需要该参数，该参数建议与replicas保持一致。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_row1065822419459"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p5658142413455"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p5658142413455"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p5658142413455"></a>replicas</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul122461585257"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul122461585257"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul122461585257"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p3302102644813"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p3302102644813"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p3302102644813"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_row9658152417458"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p12658132454515"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p12658132454515"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p12658132454515"></a>image</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p3658162417453"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p3658162417453"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p3658162417453"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1930210269483"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1930210269483"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1930210269483"></a>训练镜像名称，请根据实际修改（用户在制作镜像章节制作的镜像名称）。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_row186581324154511"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p16581924144516"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p16581924144516"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p16581924144516"></a>（可选）host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1650105613241"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1650105613241"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1650105613241"></a><span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph16676195493717"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph16676195493717"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph16676195493717"></a>Arm</span>环境：<span id="ph7409155415919"><a name="ph7409155415919"></a><a name="ph7409155415919"></a>huawei-arm</span></p>
<p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p0658124184512"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p0658124184512"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p0658124184512"></a><span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1274682034217"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1274682034217"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1274682034217"></a>x86_64</span>环境：<span id="ph47819414512"><a name="ph47819414512"></a><a name="ph47819414512"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1261514892612"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1261514892612"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1261514892612"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p173021526124817"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p173021526124817"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p173021526124817"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="row120054210308"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="p7617184343017"><a name="p7617184343017"></a><a name="p7617184343017"></a><span id="ph7617443123017"><a name="ph7617443123017"></a><a name="ph7617443123017"></a>huawei.com/schedule_policy</span></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="p186115560296"><a name="p186115560296"></a><a name="p186115560296"></a>目前支持<a href="#table1120511613153">表3</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p161115610295"><a name="p161115610295"></a><a name="p161115610295"></a>配置任务需要调度的AI芯片布局形态。<span id="zh-cn_topic_0000002511347099_ph204811934163414_1"><a name="zh-cn_topic_0000002511347099_ph204811934163414_1"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414_1"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
<div class="note" id="note15611185610297"><a name="note15611185610297"></a><a name="note15611185610297"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p561185642913"><a name="p561185642913"></a><a name="p561185642913"></a>仅支持在<span id="ph12611956122914"><a name="ph12611956122914"></a><a name="ph12611956122914"></a>Atlas 训练系列产品</span>、<span id="ph961145652910"><a name="ph961145652910"></a><a name="ph961145652910"></a><term id="zh-cn_topic_0000001519959665_term57208119917_2"><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a>Atlas A2 训练系列产品</term></span>、<span id="ph761217563298"><a name="ph761217563298"></a><a name="ph761217563298"></a><term id="zh-cn_topic_0000001519959665_term26764913715_1"><a name="zh-cn_topic_0000001519959665_term26764913715_1"></a><a name="zh-cn_topic_0000001519959665_term26764913715_1"></a>Atlas A3 训练系列产品</term></span>、<span id="ph186126569292"><a name="ph186126569292"></a><a name="ph186126569292"></a><term id="zh-cn_topic_0000001519959665_term176419491615_1"><a name="zh-cn_topic_0000001519959665_term176419491615_1"></a><a name="zh-cn_topic_0000001519959665_term176419491615_1"></a>Atlas A5 训练系列产品</term></span>、<span id="ph156121566299"><a name="ph156121566299"></a><a name="ph156121566299"></a><term id="zh-cn_topic_0000001094307702_term99602034117_1"><a name="zh-cn_topic_0000001094307702_term99602034117_1"></a><a name="zh-cn_topic_0000001094307702_term99602034117_1"></a>Atlas A2 推理系列产品</term></span>、<span id="ph186126569292"><a name="ph186126569292"></a><a name="ph186126569292"></a><term id="zh-cn_topic_0000001519959665_term176419491615_1"><a name="zh-cn_topic_0000001519959665_term176419491615_1"></a><a name="zh-cn_topic_0000001519959665_term176419491615_1"></a>Atlas A3 推理系列产品</term></span>、<span id="ph186126569292"><a name="ph186126569292"></a><a name="ph186126569292"></a><term id="zh-cn_topic_0000001519959665_term176419491615_1"><a name="zh-cn_topic_0000001519959665_term176419491615_1"></a><a name="zh-cn_topic_0000001519959665_term176419491615_1"></a>Atlas A5 推理系列产品</term></span>中使用该字段。</p>
</div></div>
</td>
</tr>
<tr id="row3642125014314"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="p414573134413"><a name="p414573134413"></a><a name="p414573134413"></a>sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="p514533164418"><a name="p514533164418"></a><a name="p514533164418"></a>指定逻辑超节点芯片数量。</p>
<a name="ul1514518315442"></a><a name="ul1514518315442"></a><ul id="ul1514518315442"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="p131459314416"><a name="p131459314416"></a><a name="p131459314416"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="zh-cn_topic_0000002511347099_ph521204025916_1"><a name="zh-cn_topic_0000002511347099_ph521204025916_1"></a><a name="zh-cn_topic_0000002511347099_ph521204025916_1"></a>若用户未指定该字段，</span><span id="zh-cn_topic_0000002511347099_ph172121408590_1"><a name="zh-cn_topic_0000002511347099_ph172121408590_1"></a><a name="zh-cn_topic_0000002511347099_ph172121408590_1"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911_1"><a name="zh-cn_topic_0000002511347099_ph192121140135911_1"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911_1"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p9145037441"><a name="p9145037441"></a><a name="p9145037441"></a>了解详细说明请参见<a href="../references.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note191456316446"><a name="note191456316446"></a><a name="note191456316446"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569_1"></a><a name="zh-cn_topic_0000002511347099_ul546892712569_1"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569_1"><li>仅支持在<span id="zh-cn_topic_0000002511347099_ph34244153594_1"><a name="zh-cn_topic_0000002511347099_ph34244153594_1"></a><a name="zh-cn_topic_0000002511347099_ph34244153594_1"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="../faq.md#任务申请的总芯片数量为32sp-block设置为32可以正常训练sp-block设置为16无法完成训练训练容器报错提示初始化连接失败">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row12322917182117"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p132726845716"><a name="zh-cn_topic_0000001951418201_p132726845716"></a><a name="zh-cn_topic_0000001951418201_p132726845716"></a>tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul1427218195710"></a><a name="zh-cn_topic_0000001951418201_ul1427218195710"></a><ul id="zh-cn_topic_0000001951418201_ul1427218195710"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note" id="zh-cn_topic_0000001951418201_note32586245294"><a name="zh-cn_topic_0000001951418201_note32586245294"></a><a name="zh-cn_topic_0000001951418201_note32586245294"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p5258102462916"><a name="zh-cn_topic_0000001951418201_p5258102462916"></a><a name="zh-cn_topic_0000001951418201_p5258102462916"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p32732087577"><a name="zh-cn_topic_0000001951418201_p32732087577"></a><a name="zh-cn_topic_0000001951418201_p32732087577"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note13620817512"><a name="zh-cn_topic_0000001951418201_note13620817512"></a><a name="zh-cn_topic_0000001951418201_note13620817512"></a><span class="notetitle"> [!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul768510427538"></a><a name="zh-cn_topic_0000001951418201_ul768510427538"></a><ul id="zh-cn_topic_0000001951418201_ul768510427538"><li>交换机亲和性调度1.0版本只支持<span id="zh-cn_topic_0000001951418201_ph188485434318"><a name="zh-cn_topic_0000001951418201_ph188485434318"></a><a name="zh-cn_topic_0000001951418201_ph188485434318"></a>Atlas 训练系列产品</span>和<span id="zh-cn_topic_0000001951418201_ph488445419430"><a name="zh-cn_topic_0000001951418201_ph488445419430"></a><a name="zh-cn_topic_0000001951418201_ph488445419430"></a>Atlas A2 训练系列产品</span>的<span id="zh-cn_topic_0000001951418201_ph1588485411430"><a name="zh-cn_topic_0000001951418201_ph1588485411430"></a><a name="zh-cn_topic_0000001951418201_ph1588485411430"></a>PyTorch</span>和<span id="zh-cn_topic_0000001951418201_ph38841054154313"><a name="zh-cn_topic_0000001951418201_ph38841054154313"></a><a name="zh-cn_topic_0000001951418201_ph38841054154313"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本只支持<span id="zh-cn_topic_0000001951418201_ph1188475414430"><a name="zh-cn_topic_0000001951418201_ph1188475414430"></a><a name="zh-cn_topic_0000001951418201_ph1188475414430"></a>Atlas A2 训练系列产品</span><span id="zh-cn_topic_0000001951418201_ph14885195444317"><a name="zh-cn_topic_0000001951418201_ph14885195444317"></a><a name="zh-cn_topic_0000001951418201_ph14885195444317"></a>PyTorch</span>框架。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_row15494422131"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1449413229314"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1449413229314"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p1449413229314"></a>accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p7665323173618"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p7665323173618"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p7665323173618"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul14200073713"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul14200073713"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul14200073713"><li><span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1881218064513"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1881218064513"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1881218064513"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li><li><span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1284164912438"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1284164912438"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1284164912438"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li></ul>
<a name="zh-cn_topic_0000001951418201_ul14320202612113"></a><a name="zh-cn_topic_0000001951418201_ul14320202612113"></a><ul id="zh-cn_topic_0000001951418201_ul14320202612113"><li><span id="ph03118298215"><a name="ph03118298215"></a><a name="ph03118298215"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000001951418201_ph18254135471314"><a name="zh-cn_topic_0000001951418201_ph18254135471314"></a><a name="zh-cn_topic_0000001951418201_ph18254135471314"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="zh-cn_topic_0000001951418201_ph4487202241512"><a name="zh-cn_topic_0000001951418201_ph4487202241512"></a><a name="zh-cn_topic_0000001951418201_ph4487202241512"></a><em id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_2"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_2"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_2"></a>{xxx}</em></span>b-8</li><li><span id="zh-cn_topic_0000001951418201_ph1114211211203"><a name="zh-cn_topic_0000001951418201_ph1114211211203"></a><a name="zh-cn_topic_0000001951418201_ph1114211211203"></a>Atlas 200T A2 Box16 异构子框</span><span id="ph3895131615489"><a name="ph3895131615489"></a><a name="ph3895131615489"></a>和</span><span id="ph1242701717485"><a name="ph1242701717485"></a><a name="ph1242701717485"></a>Atlas 200I A2 Box16 异构子框</span>：module-<span id="zh-cn_topic_0000001951418201_ph165491143158"><a name="zh-cn_topic_0000001951418201_ph165491143158"></a><a name="zh-cn_topic_0000001951418201_ph165491143158"></a><em id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_3"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_3"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1489729141619_3"></a>{xxx}</em></span>b-16</li><li><span id="ph196887651011"><a name="ph196887651011"></a><a name="ph196887651011"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 350 标卡（可选）</span>：350-Atlas-8、350-Atlas-16、350-Atlas-4p-8、350-Atlas-4p-16</li><li>Atlas 850 服务器（可选）：850-Atlas-8p-8、850-SuperPod-Atlas-8</li><li>Atlas 950 PoD 超节点（可选）：950-SuperPod-Atlas-8</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p134948221318"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p134948221318"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p134948221318"></a>根据需要运行训练任务的节点类型，选取不同的值。如果节点是<span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph329515587456"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph329515587456"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph329515587456"></a>Atlas 800 训练服务器（NPU满配）</span>，可以省略该标签。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note13585141217151"><a name="zh-cn_topic_0000001951418201_note13585141217151"></a><a name="zh-cn_topic_0000001951418201_note13585141217151"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p551372465715"><a name="zh-cn_topic_0000001951418201_p551372465715"></a><a name="zh-cn_topic_0000001951418201_p551372465715"></a><span id="zh-cn_topic_0000001951418201_ph145131424115718"><a name="zh-cn_topic_0000001951418201_ph145131424115718"></a><a name="zh-cn_topic_0000001951418201_ph145131424115718"></a>芯片型号的数值可通过<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_b168254314713_1"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_b168254314713_1"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_b168254314713_1"></a>npu-smi info</strong>命令查询，返回的“Name”字段对应信息为芯片型号，下文的{<em id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1914312018209_1"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1914312018209_1"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001519959665_i1914312018209_1"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_row1725618216467"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p15256112124619"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p15256112124619"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p15256112124619"></a>huawei.com/Ascend910</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p370843110385"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p370843110385"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p370843110385"></a>根据所使用芯片类型不同，取值如下：</p>
<a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul4403181216571"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul4403181216571"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ul4403181216571"><li><span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph141901927154611"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph141901927154611"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph141901927154611"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul169264817234"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul169264817234"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、4、8</li><li>分布式：1、2、4、8</li></ul>
</li><li><span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1312973814465"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1312973814465"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ph1312973814465"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1713712328597"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1713712328597"></a><ul id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_ul1713712328597"><li>单机单芯片：1</li><li>单机多芯片：2、4</li><li>分布式：1、2、4</li></ul>
</li><li><span id="ph157984201135"><a name="ph157984201135"></a><a name="ph157984201135"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000001951418201_ph745323894316"><a name="zh-cn_topic_0000001951418201_ph745323894316"></a><a name="zh-cn_topic_0000001951418201_ph745323894316"></a>Atlas 900 A2 PoD 集群基础单元</span><a name="zh-cn_topic_0000001951418201_ul169264817234"></a><a name="zh-cn_topic_0000001951418201_ul169264817234"></a><ul id="zh-cn_topic_0000001951418201_ul169264817234"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8</li><li>分布式：1、2、3、4、5、6、7、8</li></ul>
</li><li><span id="zh-cn_topic_0000001951418201_ph419517625020"><a name="zh-cn_topic_0000001951418201_ph419517625020"></a><a name="zh-cn_topic_0000001951418201_ph419517625020"></a>Atlas 200T A2 Box16 异构子框</span><span id="ph1891953184717"><a name="ph1891953184717"></a><a name="ph1891953184717"></a>和</span><span id="ph1149713543472"><a name="ph1149713543472"></a><a name="ph1149713543472"></a>Atlas 200I A2 Box16 异构子框</span>：<a name="zh-cn_topic_0000001951418201_ul191955617509"></a><a name="zh-cn_topic_0000001951418201_ul191955617509"></a><ul id="zh-cn_topic_0000001951418201_ul191955617509"><li>单机单芯片：1</li><li>单机多芯片：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p530216266485"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p530216266485"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_zh-cn_topic_0000001609074269_p530216266485"></a>请求的NPU数量，请根据实际修改，请求整卡时不能再同时请求vNPU。</p>
<div class="note" id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001621472369_note10624141372118"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001621472369_note10624141372118"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001621472369_note10624141372118"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul54321224184319"></a><a name="zh-cn_topic_0000001951418201_ul54321224184319"></a><ul id="zh-cn_topic_0000001951418201_ul54321224184319"><li><strong id="zh-cn_topic_0000001951418201_b16213840172320"><a name="zh-cn_topic_0000001951418201_b16213840172320"></a><a name="zh-cn_topic_0000001951418201_b16213840172320"></a>优雅容错模式</strong>支持<span id="zh-cn_topic_0000001951418201_ph158146714142"><a name="zh-cn_topic_0000001951418201_ph158146714142"></a><a name="zh-cn_topic_0000001951418201_ph158146714142"></a>Atlas 800 训练服务器</span>，且资源请求数量只能为4N、8N，N为训练节点数。</li><li><strong id="zh-cn_topic_0000001951418201_b1091614581433"><a name="zh-cn_topic_0000001951418201_b1091614581433"></a><a name="zh-cn_topic_0000001951418201_b1091614581433"></a>优雅容错模式</strong>支持<span id="ph184881417142314"><a name="ph184881417142314"></a><a name="ph184881417142314"></a>Atlas 800T A2 训练服务器</span>或<span id="zh-cn_topic_0000001951418201_ph9246916444"><a name="zh-cn_topic_0000001951418201_ph9246916444"></a><a name="zh-cn_topic_0000001951418201_ph9246916444"></a>Atlas 900 A2 PoD 集群基础单元</span>，且资源请求数量只能为8N，N为训练节点数。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_row171754462391"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p15220101916253"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p15220101916253"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p15220101916253"></a>ring-controller.atlas</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="ul1350933784715"></a><a name="ul1350933784715"></a><ul id="ul1350933784715"><li><span id="ph1550973764719"><a name="ph1550973764719"></a><a name="ph1550973764719"></a><term id="zh-cn_topic_0000001519959665_term57208119917_3"><a name="zh-cn_topic_0000001519959665_term57208119917_3"></a><a name="zh-cn_topic_0000001519959665_term57208119917_3"></a>Atlas A2 训练系列产品</term></span>、<span id="ph165091737134717"><a name="ph165091737134717"></a><a name="ph165091737134717"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="ph7509937124714"><a name="ph7509937124714"></a><a name="ph7509937124714"></a>Atlas 800T A3 超节点服务器</span>取值为：ascend-<span id="ph3509183744719"><a name="ph3509183744719"></a><a name="ph3509183744719"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b</li><li>Atlas 800 训练服务器，服务器（插<span id="ph2509143710470"><a name="ph2509143710470"></a><a name="ph2509143710470"></a>Atlas 300T 训练卡</span>）取值为：ascend-910</li><li>Atlas A5 系列产品、Atlas 350 标卡、Atlas 850 服务器、Atlas 950 SuperPod 超节点</span>取值为：huawei.com/npu</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p19220131902512"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p19220131902512"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p19220131902512"></a>用于标识任务使用的芯片的类型。需要在<span id="zh-cn_topic_0000001951418201_ph12290749162911"><a name="zh-cn_topic_0000001951418201_ph12290749162911"></a><a name="zh-cn_topic_0000001951418201_ph12290749162911"></a>ConfigMap</span>和任务task中配置。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1532024714421"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p10781181822210"><a name="zh-cn_topic_0000001951418201_p10781181822210"></a><a name="zh-cn_topic_0000001951418201_p10781181822210"></a>metadata.annotations['huawei.com/AscendXXX']</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p178151812224"><a name="zh-cn_topic_0000001951418201_p178151812224"></a><a name="zh-cn_topic_0000001951418201_p178151812224"></a>XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境的实际芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p5781181818226"><a name="zh-cn_topic_0000001951418201_p5781181818226"></a><a name="zh-cn_topic_0000001951418201_p5781181818226"></a><span id="zh-cn_topic_0000001951418201_ph1378141872210"><a name="zh-cn_topic_0000001951418201_ph1378141872210"></a><a name="zh-cn_topic_0000001951418201_ph1378141872210"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row96491720161519"><td class="cellrowborder" rowspan="5" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p19309122810151"><a name="zh-cn_topic_0000001951418201_p19309122810151"></a><a name="zh-cn_topic_0000001951418201_p19309122810151"></a>fault-scheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1058012817249"><a name="zh-cn_topic_0000001951418201_p1058012817249"></a><a name="zh-cn_topic_0000001951418201_p1058012817249"></a>grace</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1432806105912"><a name="zh-cn_topic_0000001951418201_p1432806105912"></a><a name="zh-cn_topic_0000001951418201_p1432806105912"></a>配置任务采用优雅删除模式，并在过程中先优雅删除原<span id="zh-cn_topic_0000001951418201_ph0511135011612"><a name="zh-cn_topic_0000001951418201_ph0511135011612"></a><a name="zh-cn_topic_0000001951418201_ph0511135011612"></a>Pod</span>，15分钟后若还未成功，使用强制删除原<span id="zh-cn_topic_0000001951418201_ph2149141202811"><a name="zh-cn_topic_0000001951418201_ph2149141202811"></a><a name="zh-cn_topic_0000001951418201_ph2149141202811"></a>Pod</span>。</p>
<p id="p185506120153"><a name="p185506120153"></a><a name="p185506120153"></a>进程级别重调度和进程级在线恢复场景，需将本参数配置为grace。</p>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_row328274210471"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p2032819617590"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p2032819617590"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p2032819617590"></a>force</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p113286645910"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p113286645910"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p113286645910"></a>配置任务采用强制删除模式，在过程中强制删除原<span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph38454178285"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph38454178285"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph38454178285"></a>Pod</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_row1598711439475"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p153287615911"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p153287615911"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p153287615911"></a>off</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p1832817655911"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p1832817655911"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p1832817655911"></a>该任务不使用断点续训特性，<span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph1319220540374"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph1319220540374"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph1319220540374"></a>K8s</span>的maxRetry仍然生效。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_row1066644916496"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p8667174914916"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p8667174914916"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p8667174914916"></a>无（无fault-scheduling字段）</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_row11602175216493"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p4602135219491"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p4602135219491"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p4602135219491"></a>其他值</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row4635558201210"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1499116019135"><a name="zh-cn_topic_0000001951418201_p1499116019135"></a><a name="zh-cn_topic_0000001951418201_p1499116019135"></a>recover-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p599118017133"><a name="zh-cn_topic_0000001951418201_p599118017133"></a><a name="zh-cn_topic_0000001951418201_p599118017133"></a>任务可用恢复策略。</p>
<a name="zh-cn_topic_0000001951418201_ul139911803137"></a><a name="zh-cn_topic_0000001951418201_ul139911803137"></a><ul id="zh-cn_topic_0000001951418201_ul139911803137"><li>retry：进程级在线恢复。</li><li>recover：进程级别重调度。</li><li>recover-in-place：进程级原地恢复。</li><li>dump：保存临终遗言。</li><li>exit：退出训练。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000001951418201_ul169911906135"></a><a name="zh-cn_topic_0000001951418201_ul169911906135"></a>recover-strategy配置在任务YAML annotations下，取值为5种策略的随意组合，策略之间由逗号分割。
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row10152132415157"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p10821192541514"><a name="zh-cn_topic_0000001951418201_p10821192541514"></a><a name="zh-cn_topic_0000001951418201_p10821192541514"></a>pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul5821162501510"></a><a name="zh-cn_topic_0000001951418201_ul5821162501510"></a><ul id="zh-cn_topic_0000001951418201_ul5821162501510"><li>on：开启Pod级别重调度</li><li>其他值或不使用该字段：关闭Pod级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p78221125201514"><a name="zh-cn_topic_0000001951418201_p78221125201514"></a><a name="zh-cn_topic_0000001951418201_p78221125201514"></a>Pod级别重调度，表示任务发生故障后，不会删除所有任务Pod，而是将发生故障的Pod进行删除，重新创建新Pod后进行重调度。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note5822925151516"><a name="zh-cn_topic_0000001951418201_note5822925151516"></a><a name="zh-cn_topic_0000001951418201_note5822925151516"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul17822112517158"></a><a name="zh-cn_topic_0000001951418201_ul17822112517158"></a><ul id="zh-cn_topic_0000001951418201_ul17822112517158"><li>重调度模式默认为任务级重调度，若需要开启Pod级别重调度，需要新增该字段。</li><li><span id="zh-cn_topic_0000001951418201_ph19822625101514"><a name="zh-cn_topic_0000001951418201_ph19822625101514"></a><a name="zh-cn_topic_0000001951418201_ph19822625101514"></a>TensorFlow</span>暂不支持Pod级别重调度。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row576132216324"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1772202423212"><a name="zh-cn_topic_0000001951418201_p1772202423212"></a><a name="zh-cn_topic_0000001951418201_p1772202423212"></a>subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul972624133214"></a><a name="zh-cn_topic_0000001951418201_ul972624133214"></a><ul id="zh-cn_topic_0000001951418201_ul972624133214"><li>ignore：忽略该亚健康节点，后续任务在亲和性调度上不优先调度该节点。</li><li>graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。</li><li>forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。</li><li>默认取值为ignore。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1973102463218"><a name="zh-cn_topic_0000001951418201_p1973102463218"></a><a name="zh-cn_topic_0000001951418201_p1973102463218"></a>节点状态为亚健康（SubHealthy）的节点的处理策略。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note173271703519"><a name="zh-cn_topic_0000001951418201_note173271703519"></a><a name="zh-cn_topic_0000001951418201_note173271703519"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p163271901355"><a name="zh-cn_topic_0000001951418201_p163271901355"></a><a name="zh-cn_topic_0000001951418201_p163271901355"></a>使用graceExit策略时，需保证训练框架能够接收SIGTERM信号并保存CKPT文件。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1314311835012"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p123205151739"><a name="zh-cn_topic_0000001951418201_p123205151739"></a><a name="zh-cn_topic_0000001951418201_p123205151739"></a>fault-retry-times</p>
<p id="zh-cn_topic_0000001951418201_p196969196112"><a name="zh-cn_topic_0000001951418201_p196969196112"></a><a name="zh-cn_topic_0000001951418201_p196969196112"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1192310597344"><a name="zh-cn_topic_0000001951418201_p1192310597344"></a><a name="zh-cn_topic_0000001951418201_p1192310597344"></a>0 &lt; fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p109232597342"><a name="zh-cn_topic_0000001951418201_p109232597342"></a><a name="zh-cn_topic_0000001951418201_p109232597342"></a>处理业务面故障，必须配置业务面可无条件重试的次数。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note15571815115017"><a name="zh-cn_topic_0000001951418201_note15571815115017"></a><a name="zh-cn_topic_0000001951418201_note15571815115017"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul15238182410364"></a><a name="zh-cn_topic_0000001951418201_ul15238182410364"></a><ul id="zh-cn_topic_0000001951418201_ul15238182410364"><li>使用无条件重试功能需保证训练进程异常时会导致容器异常退出，若容器未异常退出则无法成功重试。</li><li>当前仅<span id="ph1377171612516"><a name="ph1377171612516"></a><a name="ph1377171612516"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000001951418201_ph14104952376"><a name="zh-cn_topic_0000001951418201_ph14104952376"></a><a name="zh-cn_topic_0000001951418201_ph14104952376"></a>Atlas 900 A2 PoD 集群基础单元</span>支持无条件重试功能。</li><li>进行进程级恢复时，将会触发业务面故障，如需使用进程级恢复，必须配置此参数。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row260912190502"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p2966613113520"><a name="zh-cn_topic_0000001951418201_p2966613113520"></a><a name="zh-cn_topic_0000001951418201_p2966613113520"></a>无（无fault-retry-times）或0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p2096618130353"><a name="zh-cn_topic_0000001951418201_p2096618130353"></a><a name="zh-cn_topic_0000001951418201_p2096618130353"></a>该任务不使用无条件重试功能，无法感知业务面故障，vcjob的maxRetry仍然生效。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row128551542131510"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p10285161985210"><a name="zh-cn_topic_0000001951418201_p10285161985210"></a><a name="zh-cn_topic_0000001951418201_p10285161985210"></a>policies</p>
<p id="zh-cn_topic_0000001951418201_p490916512164"><a name="zh-cn_topic_0000001951418201_p490916512164"></a><a name="zh-cn_topic_0000001951418201_p490916512164"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p056810252162"><a name="zh-cn_topic_0000001951418201_p056810252162"></a><a name="zh-cn_topic_0000001951418201_p056810252162"></a>event，取值如下：</p>
<a name="zh-cn_topic_0000001951418201_ul1781384818238"></a><a name="zh-cn_topic_0000001951418201_ul1781384818238"></a><ul id="zh-cn_topic_0000001951418201_ul1781384818238"><li>PodFailed：Pod失败</li><li>PodEvicted：Pod被驱逐</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p180717598243"><a name="zh-cn_topic_0000001951418201_p180717598243"></a><a name="zh-cn_topic_0000001951418201_p180717598243"></a>Pod状态。与action字段搭配使用，表示当Pod处于某种状态时，<span id="zh-cn_topic_0000001951418201_ph525518226126"><a name="zh-cn_topic_0000001951418201_ph525518226126"></a><a name="zh-cn_topic_0000001951418201_ph525518226126"></a>Volcano</span>的处理策略。默认值为PodEvicted。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1390814541612"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1590911581611"><a name="zh-cn_topic_0000001951418201_p1590911581611"></a><a name="zh-cn_topic_0000001951418201_p1590911581611"></a>action，取值如下：</p>
<a name="zh-cn_topic_0000001951418201_ul17824133752420"></a><a name="zh-cn_topic_0000001951418201_ul17824133752420"></a><ul id="zh-cn_topic_0000001951418201_ul17824133752420"><li>RestartJob：重新启动训练任务。</li><li>Ignore：<span id="zh-cn_topic_0000001951418201_ph141051824104819"><a name="zh-cn_topic_0000001951418201_ph141051824104819"></a><a name="zh-cn_topic_0000001951418201_ph141051824104819"></a>忽略。开源Volcano</span>不做任何处理，由<span id="zh-cn_topic_0000001951418201_ph631119334409"><a name="zh-cn_topic_0000001951418201_ph631119334409"></a><a name="zh-cn_topic_0000001951418201_ph631119334409"></a>Ascend-volcano-plugin</span>插件进行处理。</li></ul>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p20698124111427"><a name="zh-cn_topic_0000001951418201_p20698124111427"></a><a name="zh-cn_topic_0000001951418201_p20698124111427"></a><span id="zh-cn_topic_0000001951418201_ph10699341154214"><a name="zh-cn_topic_0000001951418201_ph10699341154214"></a><a name="zh-cn_topic_0000001951418201_ph10699341154214"></a>Volcano</span>对处于某种状态的Pod的处理策略。默认值为RestartJob。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note128691230174312"><a name="zh-cn_topic_0000001951418201_note128691230174312"></a><a name="zh-cn_topic_0000001951418201_note128691230174312"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000001951418201_ul13166894452"></a><a name="zh-cn_topic_0000001951418201_ul13166894452"></a><ul id="zh-cn_topic_0000001951418201_ul13166894452"><li>开启Pod级别重调度需要删除policies及其子参数event和action。</li><li>使用业务面故障无条件重试时（或同时使用Pod级别重调度和业务面故障无条件重试），需要将event配置为PodFailed；action配置为Ignore。</li><li>如果不使用集群调度组件的<span id="zh-cn_topic_0000001951418201_ph8224175173014"><a name="zh-cn_topic_0000001951418201_ph8224175173014"></a><a name="zh-cn_topic_0000001951418201_ph8224175173014"></a>Volcano</span>或者开源<span id="zh-cn_topic_0000001951418201_ph83286473313"><a name="zh-cn_topic_0000001951418201_ph83286473313"></a><a name="zh-cn_topic_0000001951418201_ph83286473313"></a>Volcano</span>没有集成<span id="zh-cn_topic_0000001951418201_ph617718254597"><a name="zh-cn_topic_0000001951418201_ph617718254597"></a><a name="zh-cn_topic_0000001951418201_ph617718254597"></a>Ascend-volcano-plugin</span>插件，需要参考<a href="../faq.md#使用volcano和ascend-operator组件场景下业务面故障的任务所有pod的status全部变为failed任务无法触发无条件重试重调度">使用Volcano和Ascend Operator组件场景下，业务面故障的任务所有Pod的Status全部变为Failed，任务无法触发无条件重试重调度</a>修改开源Volcano代码。</li><li>开源Volcano还提供了policies的其他取值，不建议用户修改为其他取值，否则可能影响断点续训功能的正常使用。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row1458223119296"><td class="cellrowborder" rowspan="2" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p7582183112296"><a name="zh-cn_topic_0000001951418201_p7582183112296"></a><a name="zh-cn_topic_0000001951418201_p7582183112296"></a>maxRetry</p>
<p id="zh-cn_topic_0000001951418201_p1758196165112"><a name="zh-cn_topic_0000001951418201_p1758196165112"></a><a name="zh-cn_topic_0000001951418201_p1758196165112"></a></p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1026835111"><a name="zh-cn_topic_0000001951418201_p1026835111"></a><a name="zh-cn_topic_0000001951418201_p1026835111"></a>0&lt; maxRetry</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p2216813515"><a name="zh-cn_topic_0000001951418201_p2216813515"></a><a name="zh-cn_topic_0000001951418201_p2216813515"></a>任务重调度次数。任务故障时，可以重调度的次数，当已经重调度次数与maxRetry取值相同时，任务将不再进行重调度。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note394611531302"><a name="zh-cn_topic_0000001951418201_note394611531302"></a><a name="zh-cn_topic_0000001951418201_note394611531302"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p0947553193013"><a name="zh-cn_topic_0000001951418201_p0947553193013"></a><a name="zh-cn_topic_0000001951418201_p0947553193013"></a>同时配置了maxRetry和fault-retry-times参数时，当已经重调度次数与maxRetry或fault-retry-times取值有一个相同时，将不再进行重调度。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row11581962517"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p13882123719515"><a name="zh-cn_topic_0000001951418201_p13882123719515"></a><a name="zh-cn_topic_0000001951418201_p13882123719515"></a>无（无maxRetry）或maxRetry等于0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p1637895110"><a name="zh-cn_topic_0000001951418201_p1637895110"></a><a name="zh-cn_topic_0000001951418201_p1637895110"></a>不配置maxRetry或配置maxRetry取值为0时，系统默认进行3次重调度。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_row11217021145014"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p1929464718814"><a name="zh-cn_topic_0000001951418201_p1929464718814"></a><a name="zh-cn_topic_0000001951418201_p1929464718814"></a>restartPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000001951418201_ul193373071216"></a><a name="zh-cn_topic_0000001951418201_ul193373071216"></a><ul id="zh-cn_topic_0000001951418201_ul193373071216"><li>Never：从不重启</li><li>Always：总是重启</li><li>OnFailure：失败时重启</li><li>ExitCode：根据进程退出码决定是否重启Pod，错误码是1~127时不重启，128~255时重启Pod。<div class="note" id="zh-cn_topic_0000001951418201_note278954373014"><a name="zh-cn_topic_0000001951418201_note278954373014"></a><a name="zh-cn_topic_0000001951418201_note278954373014"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000001951418201_p14789194311309"><a name="zh-cn_topic_0000001951418201_p14789194311309"></a><a name="zh-cn_topic_0000001951418201_p14789194311309"></a>vcjob类型的训练任务不支持ExitCode。</p>
</div></div>
</li></ul>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_p1129434710811"><a name="zh-cn_topic_0000001951418201_p1129434710811"></a><a name="zh-cn_topic_0000001951418201_p1129434710811"></a>容器重启策略。当配置业务面故障无条件重试时，容器重启策略取值必须为<span class="parmvalue" id="zh-cn_topic_0000001951418201_parmvalue182751614652"><a name="zh-cn_topic_0000001951418201_parmvalue182751614652"></a><a name="zh-cn_topic_0000001951418201_parmvalue182751614652"></a>“Never”</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_row1116371844811"><td class="cellrowborder" valign="top" width="27.18%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p246371419493"></a>terminationGracePeriodSeconds</p>
</td>
<td class="cellrowborder" valign="top" width="36.26%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p9919805116"></a>0 &lt; terminationGracePeriodSeconds &lt;<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1192168195110"></a> grace-over-time</strong>参数取值</p>
</td>
<td class="cellrowborder" valign="top" width="36.559999999999995%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_p3929811514"></a>容器收到SIGTERM到被<span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph20922835119"></a>K8s</span>强制停止经历的时间，该时间需要大于0且小于volcano-v<em id="zh-cn_topic_0000001951418201_i1645121221719"><a name="zh-cn_topic_0000001951418201_i1645121221719"></a><a name="zh-cn_topic_0000001951418201_i1645121221719"></a>{version}</em>.yaml文件中“<strong id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_b1292208135117"></a>grace-over-time</strong>”参数取值，同时还需要保证能够保存CKPT文件，请根据实际情况修改。具体说明请参考<span id="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"></a><a name="zh-cn_topic_0000001951418201_zh-cn_topic_0000001570873348_ph7921589510"></a>K8s</span>官网<a href="https://kubernetes.io/zh/docs/concepts/containers/container-lifecycle-hooks/" target="_blank" rel="noopener noreferrer">容器生命周期回调</a>。</p>
<div class="note" id="zh-cn_topic_0000001951418201_note17641176363"><a name="zh-cn_topic_0000001951418201_note17641176363"></a><a name="zh-cn_topic_0000001951418201_note17641176363"></a><div class="notebody"><p id="zh-cn_topic_0000001951418201_p97641517103616"><a name="zh-cn_topic_0000001951418201_p97641517103616"></a><a name="zh-cn_topic_0000001951418201_p97641517103616"></a>只有当fault-scheduling配置为grace时，该字段才生效；fault-scheduling配置为force时，该字段无效。</p>
</div></div>
</td>
</tr>
</tbody>
</table>

**表 3**  huawei.com/schedule\_policy配置说明

<a name="table1120511613153"></a>
|配置|说明|
|--|--|
|chip4-node8|1个节点8张芯片，每4个芯片形成1个互联环。例如，Atlas 800 训练服务器（型号 9000）/Atlas 800 训练服务器（型号 9010）芯片的整模块场景/Atlas 350 标卡共8张卡，每4张卡通过UB扣板连接。|
|chip1-node2|1个节点2张芯片。例如，Atlas 300T 训练卡的插卡场景，1张卡最多插1个芯片，1个节点最多插2张卡。|
|chip4-node4|1个节点4张芯片，形成1个互联环。例如，Atlas 800 训练服务器（型号 9000）/Atlas 800 训练服务器（型号 9010）芯片的半配场景。|
|chip8-node8|1个节点8张卡，8张卡都在1个互联环上。例如，Atlas 800T A2 训练服务器 /Atlas 850 服务器。|
|chip8-node16|1个节点16张卡，每8张卡在1个互联环上。例如，Atlas 200T A2 Box16 异构子框。|
|chip2-node16|1个节点16张卡，每2张卡在1个互联环上。例如，Atlas 800T A3 超节点服务器。|
|chip2-node16-sp|1个节点16张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，Atlas 900 A3 SuperPoD 超节点。|
|chip4-node16|1个节点16张卡，每4张卡都在1个互联环上。例如，Atlas 350 标卡共16张卡，每4张卡通过UB扣板连接。|
|chip1-node8|1个节点8张卡，每张卡之间无互联。例如，Atlas 350 标卡共8张卡，每张卡之间无互联。|
|chip1-node16|1个节点16张卡，每张卡之间无互联。例如，Atlas 350 标卡共16张卡，每张卡之间无互联。|
|chip8-node8-sp|1个节点8张卡，8张卡都在1个互联环上，多个服务器形成超节点。例如，Atlas 850 超节点服务器。|
|chip8-node8-ra64-sp|1个节点8张卡，8张卡都在1个互联环上，64个节点组成一个计算框，多个框形成超节点。例如，Atlas 950 PoD 超节点集群。|

**rings-config-<任务名称\><a name="section1377115581385"></a>**

**表 4**  rings-config-任务名称

<a name="table1328211233126"></a>
<table><thead align="left"><tr id="row2028442312122"><th class="cellrowborder" valign="top" width="9.99%" id="mcps1.2.6.1.1"><p id="p0566161515246"><a name="p0566161515246"></a><a name="p0566161515246"></a>字段名称</p>
</th>
<th class="cellrowborder" valign="top" width="20.630000000000003%" id="mcps1.2.6.1.2"><p id="p428442317128"><a name="p428442317128"></a><a name="p428442317128"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="27.779999999999998%" id="mcps1.2.6.1.3"><p id="p32851623121215"><a name="p32851623121215"></a><a name="p32851623121215"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="18.48%" id="mcps1.2.6.1.4"><p id="p122851233123"><a name="p122851233123"></a><a name="p122851233123"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.5"><p id="p2028522371219"><a name="p2028522371219"></a><a name="p2028522371219"></a>备注</p>
</th>
</tr>
</thead>
<tbody><tr id="row142851523101217"><td class="cellrowborder" rowspan="9" valign="top" width="9.99%" headers="mcps1.2.6.1.1 "><p id="p383082613247"><a name="p383082613247"></a><a name="p383082613247"></a>hccl.json</p>
</td>
<td class="cellrowborder" valign="top" width="20.630000000000003%" headers="mcps1.2.6.1.2 "><p id="p20285172315128"><a name="p20285172315128"></a><a name="p20285172315128"></a>version</p>
</td>
<td class="cellrowborder" valign="top" width="27.779999999999998%" headers="mcps1.2.6.1.3 "><p id="p628512301218"><a name="p628512301218"></a><a name="p628512301218"></a>RankTable使用的格式版本</p>
</td>
<td class="cellrowborder" valign="top" width="18.48%" headers="mcps1.2.6.1.4 "><p id="p728612315129"><a name="p728612315129"></a><a name="p728612315129"></a>1.0</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.5 "><p id="p192861223151219"><a name="p192861223151219"></a><a name="p192861223151219"></a>-</p>
</td>
</tr>
<tr id="row92861423161214"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p16286723111214"><a name="p16286723111214"></a><a name="p16286723111214"></a>server_count</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p528602331212"><a name="p528602331212"></a><a name="p528602331212"></a>任务使用的节点数量</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p10286623111214"><a name="p10286623111214"></a><a name="p10286623111214"></a>整数类型</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p13286122371216"><a name="p13286122371216"></a><a name="p13286122371216"></a>-</p>
</td>
</tr>
<tr id="row1628711236125"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p528710237122"><a name="p528710237122"></a><a name="p528710237122"></a>server_list</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12287423101219"><a name="p12287423101219"></a><a name="p12287423101219"></a>任务使用的节点信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p428716231129"><a name="p428716231129"></a><a name="p428716231129"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p52876233123"><a name="p52876233123"></a><a name="p52876233123"></a>-</p>
</td>
</tr>
<tr id="row228712311218"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p172881238128"><a name="p172881238128"></a><a name="p172881238128"></a>- server_id</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p228820231122"><a name="p228820231122"></a><a name="p228820231122"></a>AI Server标识，全局唯一</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p928812351211"><a name="p928812351211"></a><a name="p928812351211"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p112882023131216"><a name="p112882023131216"></a><a name="p112882023131216"></a>-</p>
</td>
</tr>
<tr id="row526819617575"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p331911115717"><a name="p331911115717"></a><a name="p331911115717"></a>- host_ip</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17314111571"><a name="p17314111571"></a><a name="p17314111571"></a>AI Server的Host IP地址</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p731131111573"><a name="p731131111573"></a><a name="p731131111573"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p73171113571"><a name="p73171113571"></a><a name="p73171113571"></a>-</p>
</td>
</tr>
<tr id="row1128892381211"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p6288523171213"><a name="p6288523171213"></a><a name="p6288523171213"></a>device</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p128916238124"><a name="p128916238124"></a><a name="p128916238124"></a>任务使用的芯片信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p14289152311122"><a name="p14289152311122"></a><a name="p14289152311122"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p028902319121"><a name="p028902319121"></a><a name="p028902319121"></a>-</p>
</td>
</tr>
<tr id="row528919236120"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1428972321213"><a name="p1428972321213"></a><a name="p1428972321213"></a>- device_id</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5289423131210"><a name="p5289423131210"></a><a name="p5289423131210"></a>任务使用的芯片的物理ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p0289172316122"><a name="p0289172316122"></a><a name="p0289172316122"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1328912313128"><a name="p1328912313128"></a><a name="p1328912313128"></a>-</p>
</td>
</tr>
<tr id="row10290172361218"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p122909232126"><a name="p122909232126"></a><a name="p122909232126"></a>- device_ip</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14290023151213"><a name="p14290023151213"></a><a name="p14290023151213"></a>任务使用的芯片的IP地址</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1929042321217"><a name="p1929042321217"></a><a name="p1929042321217"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p62901323161220"><a name="p62901323161220"></a><a name="p62901323161220"></a>-</p>
</td>
</tr>
<tr id="row1429013237126"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p6290123131214"><a name="p6290123131214"></a><a name="p6290123131214"></a>- rank_id</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12291202391210"><a name="p12291202391210"></a><a name="p12291202391210"></a>任务使用的芯片的Rank号</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p17291223121215"><a name="p17291223121215"></a><a name="p17291223121215"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p52911623171220"><a name="p52911623171220"></a><a name="p52911623171220"></a>-</p>
</td>
</tr>
<tr id="row115483483241"><td class="cellrowborder" valign="top" width="9.99%" headers="mcps1.2.6.1.1 "><p id="p6549948102419"><a name="p6549948102419"></a><a name="p6549948102419"></a>version</p>
</td>
<td class="cellrowborder" valign="top" width="20.630000000000003%" headers="mcps1.2.6.1.2 "><p id="p12549174817242"><a name="p12549174817242"></a><a name="p12549174817242"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="27.779999999999998%" headers="mcps1.2.6.1.3 "><p id="p165492487247"><a name="p165492487247"></a><a name="p165492487247"></a>任务使用hccl.json的版本</p>
</td>
<td class="cellrowborder" valign="top" width="18.48%" headers="mcps1.2.6.1.4 "><p id="p145492048112416"><a name="p145492048112416"></a><a name="p145492048112416"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.5 "><p id="p65521848132412"><a name="p65521848132412"></a><a name="p65521848132412"></a>-</p>
</td>
</tr>
</tbody>
</table>

