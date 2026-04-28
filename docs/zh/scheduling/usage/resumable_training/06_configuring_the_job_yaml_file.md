# 配置任务YAML<a name="ZH-CN_TOPIC_0000002479226518"></a>

## YAML参数说明<a name="ZH-CN_TOPIC_0000002479386550"></a>

如果是acjob任务，在配置YAML前，请先了解相关YAML参数说明，详细说明如[表1](#zh-cn_topic_0000002039339953_table11351193062117)所示。

每个acjob任务YAML中包含一些固定字段，例如apiVersion、kind等，如果想了解这些字段的详细说明请参见[acjob关键字段说明](../../api/ascend_job.md)。

**表 1**  YAML参数说明

<a name="zh-cn_topic_0000002039339953_table11351193062117"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002039339953_row635183013217"><th class="cellrowborder" valign="top" width="25.042504250425047%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002039339953_p94514441221"><a name="zh-cn_topic_0000002039339953_p94514441221"></a><a name="zh-cn_topic_0000002039339953_p94514441221"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="24.76247624762476%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002039339953_p1045113449226"><a name="zh-cn_topic_0000002039339953_p1045113449226"></a><a name="zh-cn_topic_0000002039339953_p1045113449226"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="50.1950195019502%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002039339953_p18451154472214"><a name="zh-cn_topic_0000002039339953_p18451154472214"></a><a name="zh-cn_topic_0000002039339953_p18451154472214"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002039339953_row43521630112117"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p64516449226"><a name="zh-cn_topic_0000002039339953_p64516449226"></a><a name="zh-cn_topic_0000002039339953_p64516449226"></a>(.kind=="AscendJob").metadata.labels.framework</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul745174412210"></a><a name="zh-cn_topic_0000002039339953_ul745174412210"></a><ul id="zh-cn_topic_0000002039339953_ul745174412210"><li>mindspore</li><li>pytorch</li><li>tensorflow</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p1745112442227"><a name="zh-cn_topic_0000002039339953_p1745112442227"></a><a name="zh-cn_topic_0000002039339953_p1745112442227"></a>框架类型，目前只支持三种。</p>
</td>
</tr>
<tr id="row133645515273"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p2436102814254"><a name="zh-cn_topic_0000001951418201_p2436102814254"></a><a name="zh-cn_topic_0000001951418201_p2436102814254"></a>(.kind=="AscendJob").metadata.labels.jobID</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="p1619843317517"><a name="p1619843317517"></a><a name="p1619843317517"></a>当前MindIE Motor任务在集群中的唯一识别ID，用户可根据实际情况进行配置。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p9111039132818"><a name="p9111039132818"></a><a name="p9111039132818"></a>该参数仅支持在<span id="ph12174764117"><a name="ph12174764117"></a><a name="ph12174764117"></a>Atlas 800I A3 超节点服务器</span>和<span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span>上使用。</p>
</td>
</tr>
<tr id="row1199016622817"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951418201_p13524833182513"><a name="zh-cn_topic_0000001951418201_p13524833182513"></a><a name="zh-cn_topic_0000001951418201_p13524833182513"></a>(.kind=="AscendJob").metadata.labels.app</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951418201_p5524103317257"><a name="zh-cn_topic_0000001951418201_p5524103317257"></a><a name="zh-cn_topic_0000001951418201_p5524103317257"></a>表明MindIE Motor任务在Ascend Job中的角色，取值包括mindie-ms-controller、mindie-ms-coordinator、mindie-ms-server。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><div class="note" id="zh-cn_topic_0000001951418201_note4367125713295"><a name="zh-cn_topic_0000001951418201_note4367125713295"></a><div class="notebody"><a name="ul139591420161415"></a><a name="ul139591420161415"></a><ul id="ul139591420161415"><li>acjob的任务YAML同时包含jobID和app这2个字段时，<span id="zh-cn_topic_0000001951418201_ph1566531814589"><a name="zh-cn_topic_0000001951418201_ph1566531814589"></a><a name="zh-cn_topic_0000001951418201_ph1566531814589"></a>Ascend Operator</span>组件会自动传入环境变量MINDX_TASK_ID、APP_TYPE、MINDX_SERVER_IP及MINDX_SERVER_DOMAIN，并将其标识为MindIE推理任务。</li><li>关于以上环境变量的详细说明请参见<a href="../../api/environment_variable_description.md">Ascend Operator注入的训练环境变量</a>。</li><li>该参数仅支持在<span id="ph1493312176292"><a name="ph1493312176292"></a><a name="ph1493312176292"></a>Atlas 800I A3 超节点服务器</span>和<span id="ph1893331752914"><a name="ph1893331752914"></a><a name="ph1893331752914"></a>Atlas 800I A2 推理服务器</span>上使用。</li></ul>
</div></div>
</td>
</tr>
<tr id="row1553412124289"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="p143891219113814"><a name="p143891219113814"></a><a name="p143891219113814"></a><span>(.kind=="AscendJob").metadata.labels.mind-cluster/scaling-rule: scaling-rule</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="p3389151918383"><a name="p3389151918383"></a><a name="p3389151918383"></a>标记扩缩容规则对应的ConfigMap名称。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p10101730182112"><a name="p10101730182112"></a><a name="p10101730182112"></a>仅支持MindIE Motor推理任务在<span id="ph13640202812297"><a name="ph13640202812297"></a><a name="ph13640202812297"></a>Atlas 800I A3 超节点服务器</span>和<span id="ph136401628192912"><a name="ph136401628192912"></a><a name="ph136401628192912"></a>Atlas 800I A2 推理服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="row9133171112813"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="p16388101913817"><a name="p16388101913817"></a><a name="p16388101913817"></a><span>(.kind=="AscendJob").metadata.labels.mind-cluster/group-name: group0</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="p13387171983812"><a name="p13387171983812"></a><a name="p13387171983812"></a>标记扩缩容规则中对应的group名称。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p14289143313188"><a name="p14289143313188"></a><a name="p14289143313188"></a>仅支持MindIE Motor推理任务在<span id="ph156731391296"><a name="ph156731391296"></a><a name="ph156731391296"></a>Atlas 800I A3 超节点服务器</span>和<span id="ph26731039182916"><a name="ph26731039182916"></a><a name="ph26731039182916"></a>Atlas 800I A2 推理服务器</span>上使用本参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row7208139102014"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1447163824215"><a name="zh-cn_topic_0000002039339953_p1447163824215"></a><a name="zh-cn_topic_0000002039339953_p1447163824215"></a>(.kind=="AscendJob").metadata.labels."ring-controller.atlas"</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul16300163019516"></a><a name="zh-cn_topic_0000002039339953_ul16300163019516"></a>
    <ul id="zh-cn_topic_0000002039339953_ul16300163019516">
        <li>Atlas 800 训练服务器：ascend-910</li>
        <li><span id="zh-cn_topic_0000002039339953_ph760316141835"><a name="zh-cn_topic_0000002039339953_ph760316141835"></a><a name="zh-cn_topic_0000002039339953_ph760316141835"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>、<span id="zh-cn_topic_0000002039339953_ph163483412215"><a name="zh-cn_topic_0000002039339953_ph163483412215"></a><a name="zh-cn_topic_0000002039339953_ph163483412215"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph10355115144111"><a name="ph10355115144111"></a><a name="ph10355115144111"></a>Atlas 800T A3 超节点服务器</span>、<span id="ph1086225213421"><a name="ph1086225213421"></a><a name="ph1086225213421"></a>Atlas 800I A3 超节点服务器</span>和<span id="zh-cn_topic_0000002039339953_ph960313141731"><a name="zh-cn_topic_0000002039339953_ph960313141731"></a><a name="zh-cn_topic_0000002039339953_ph960313141731"></a>Atlas 900 A3 SuperPoD 超节点</span>：ascend-<span id="zh-cn_topic_0000002039339953_ph1360341413317"><a name="zh-cn_topic_0000002039339953_ph1360341413317"></a><a name="zh-cn_topic_0000002039339953_ph1360341413317"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li>
        <li>（可选）Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD：ascend-npu</li>
    </ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p43811639112614"><a name="zh-cn_topic_0000002039339953_p43811639112614"></a><a name="zh-cn_topic_0000002039339953_p43811639112614"></a>标识任务使用的芯片的产品类型。</p>
<p id="zh-cn_topic_0000002039339953_zh-cn_topic_0000001570873348_p19220131902512"><a name="zh-cn_topic_0000002039339953_zh-cn_topic_0000001570873348_p19220131902512"></a><a name="zh-cn_topic_0000002039339953_zh-cn_topic_0000001570873348_p19220131902512"></a>需要在<span id="zh-cn_topic_0000002039339953_ph12290749162911"><a name="zh-cn_topic_0000002039339953_ph12290749162911"></a><a name="zh-cn_topic_0000002039339953_ph12290749162911"></a>ConfigMap</span>和任务task中配置。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1735283013214"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p8451174482214"><a name="zh-cn_topic_0000002039339953_p8451174482214"></a><a name="zh-cn_topic_0000002039339953_p8451174482214"></a>(.kind=="AscendJob").metadata.labels.tor-affinity</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul645114462218"></a><a name="zh-cn_topic_0000002039339953_ul645114462218"></a><ul id="zh-cn_topic_0000002039339953_ul645114462218"><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度</li></ul>
<div class="note" id="zh-cn_topic_0000002039339953_note62680445222"><a name="zh-cn_topic_0000002039339953_note62680445222"></a><a name="zh-cn_topic_0000002039339953_note62680445222"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339953_p14582042192513"><a name="zh-cn_topic_0000002039339953_p14582042192513"></a><a name="zh-cn_topic_0000002039339953_p14582042192513"></a>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p>
</div></div>
<p id="zh-cn_topic_0000002039339953_p112971504251"><a name="zh-cn_topic_0000002039339953_p112971504251"></a><a name="zh-cn_topic_0000002039339953_p112971504251"></a></p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p6452124472216"><a name="zh-cn_topic_0000002039339953_p6452124472216"></a><a name="zh-cn_topic_0000002039339953_p6452124472216"></a>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note1528312443229"><a name="zh-cn_topic_0000002039339953_note1528312443229"></a><div class="notebody"><a name="zh-cn_topic_0000002039339953_ul176092014030"></a><a name="zh-cn_topic_0000002039339953_ul176092014030"></a><ul id="zh-cn_topic_0000002039339953_ul176092014030"><li>交换机亲和性调度1.0版本支持<span id="zh-cn_topic_0000002039339953_ph1157665817140"><a name="zh-cn_topic_0000002039339953_ph1157665817140"></a><a name="zh-cn_topic_0000002039339953_ph1157665817140"></a>Atlas 训练系列产品</span>和<span id="zh-cn_topic_0000002039339953_ph168598363399"><a name="zh-cn_topic_0000002039339953_ph168598363399"></a><a name="zh-cn_topic_0000002039339953_ph168598363399"></a><term id="zh-cn_topic_0000001519959665_term57208119917_1"><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a>Atlas A2 训练系列产品</term></span>；支持<span id="zh-cn_topic_0000002039339953_ph4181625925"><a name="zh-cn_topic_0000002039339953_ph4181625925"></a><a name="zh-cn_topic_0000002039339953_ph4181625925"></a>PyTorch</span>和<span id="zh-cn_topic_0000002039339953_ph61882510210"><a name="zh-cn_topic_0000002039339953_ph61882510210"></a><a name="zh-cn_topic_0000002039339953_ph61882510210"></a>MindSpore</span>框架。</li><li>交换机亲和性调度2.0版本支持<span id="zh-cn_topic_0000002039339953_ph311717506401"><a name="zh-cn_topic_0000002039339953_ph311717506401"></a><a name="zh-cn_topic_0000002039339953_ph311717506401"></a><term id="zh-cn_topic_0000001519959665_term57208119917_2"><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a><a name="zh-cn_topic_0000001519959665_term57208119917_2"></a>Atlas A2 训练系列产品</term></span>；支持<span id="zh-cn_topic_0000002039339953_ph17383182419412"><a name="zh-cn_topic_0000002039339953_ph17383182419412"></a><a name="zh-cn_topic_0000002039339953_ph17383182419412"></a>PyTorch</span>框架。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row83521230102117"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1745254442212"><a name="zh-cn_topic_0000002039339953_p1745254442212"></a><a name="zh-cn_topic_0000002039339953_p1745254442212"></a>(.kind=="AscendJob").metadata.labels.pod-rescheduling</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul14521244102210"></a><a name="zh-cn_topic_0000002039339953_ul14521244102210"></a><ul id="zh-cn_topic_0000002039339953_ul14521244102210"><li>on：开启Pod级别重调度</li><li>其他值或不使用该字段：关闭Pod级别重调度</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p12453044102217"><a name="zh-cn_topic_0000002039339953_p12453044102217"></a><a name="zh-cn_topic_0000002039339953_p12453044102217"></a>Pod级别重调度，表示任务发生故障后，不会删除所有任务Pod，而是将发生故障的Pod进行删除，重新创建新Pod后进行重调度。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note1430334413223"><a name="zh-cn_topic_0000002039339953_note1430334413223"></a><div class="notebody"><a name="zh-cn_topic_0000002039339953_ul461013147314"></a><a name="zh-cn_topic_0000002039339953_ul461013147314"></a><ul id="zh-cn_topic_0000002039339953_ul461013147314"><li>重调度模式默认为任务级重调度，若需要开启Pod级别重调度，需要新增该字段。</li><li><span id="zh-cn_topic_0000002039339953_ph1061091414318"><a name="zh-cn_topic_0000002039339953_ph1061091414318"></a><a name="zh-cn_topic_0000002039339953_ph1061091414318"></a>TensorFlow</span>暂不支持Pod级别重调度。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row435215305211"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p18454184442214"><a name="zh-cn_topic_0000002039339953_p18454184442214"></a><a name="zh-cn_topic_0000002039339953_p18454184442214"></a>(.kind=="AscendJob").metadata.labels.process-recover-enable</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul71592205015"></a><a name="zh-cn_topic_0000002039339953_ul71592205015"></a><ul id="zh-cn_topic_0000002039339953_ul71592205015"><li>on：开启进程级别重调度及进程级在线恢复。<p>进程级别重调度和优雅容错不能同时开启，若同时开启，断点续训将通过Job级别重调度恢复训练。</p>
</li><li>pause：暂时关闭进程级别重调度及进程级在线恢复。</li><li>off或不使用该字段：关闭进程级别重调度及进程级在线恢复。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p10523132473119"><a name="zh-cn_topic_0000002039339953_p10523132473119"></a><a name="zh-cn_topic_0000002039339953_p10523132473119"></a>Ascend Operator会根据用户配置的recover-strategy自动给任务打上process-recover-enable=on标签，无需用户手动指定。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1635217304212"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p10454144492212"><a name="zh-cn_topic_0000002039339953_p10454144492212"></a><a name="zh-cn_topic_0000002039339953_p10454144492212"></a>(.kind=="AscendJob").metadata.annotations.recover-strategy</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p156641331161715"><a name="zh-cn_topic_0000002039339953_p156641331161715"></a><a name="zh-cn_topic_0000002039339953_p156641331161715"></a>任务可用恢复策略。</p>
<a name="zh-cn_topic_0000002039339953_ul6665173119177"></a><a name="zh-cn_topic_0000002039339953_ul6665173119177"></a><ul id="zh-cn_topic_0000002039339953_ul6665173119177"><li>retry：进程级在线恢复。</li><li>recover：进程级别重调度。</li><li><span>recover-in-place：进程级原地恢复</span>。</li><li>elastic-training：弹性训练。</li><li>dump：保存临终遗言。</li><li>exit：退出训练。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039339953_ul18941121318614"></a><a name="zh-cn_topic_0000002039339953_ul18941121318614"></a>recover-strategy配置在任务YAML的annotations下，取值为6种策略的随意组合，策略之间由逗号分割。
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row5353133052111"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1745411440228"><a name="zh-cn_topic_0000002039339953_p1745411440228"></a><a name="zh-cn_topic_0000002039339953_p1745411440228"></a>(.kind=="AscendJob").metadata.labels.subHealthyStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul18716519102210"></a><a name="zh-cn_topic_0000002039339953_ul18716519102210"></a><ul id="zh-cn_topic_0000002039339953_ul18716519102210"><li>ignore：忽略该亚健康节点，后续任务在亲和性调度上不优先调度该节点。</li><li>graceExit：不使用亚健康节点，并保存临终CKPT文件后，进行重调度，后续任务不会调度到该节点。</li><li>forceExit：不使用亚健康节点，不保存任务直接退出，进行重调度，后续任务不会调度到该节点。</li><li>hotSwitch：执行亚健康热切，拉起备份Pod后，暂停训练任务，并使用新节点重新拉起训练。</li><li>默认取值为ignore。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p10455114482217"><a name="zh-cn_topic_0000002039339953_p10455114482217"></a><a name="zh-cn_topic_0000002039339953_p10455114482217"></a>节点状态为亚健康（SubHealthy）的节点的处理策略。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note1751515456204"><a name="zh-cn_topic_0000002039339953_note1751515456204"></a><div class="notebody"><a name="ul24832019017"></a><a name="ul24832019017"></a><ul id="ul24832019017"><li>使用graceExit策略时，需保证任务开启了临终CKPT保存功能。</li><li>hotSwitch策略的使用约束请参见<a href="./01_solutions_principles.md#亚健康热切">使用约束</a>。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row163537304214"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p10887172213238"><a name="zh-cn_topic_0000002039339953_p10887172213238"></a><a name="zh-cn_topic_0000002039339953_p10887172213238"></a>(.kind=="AscendJob").specs.schedulerName</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1888718226237"><a name="zh-cn_topic_0000002039339953_p1888718226237"></a><a name="zh-cn_topic_0000002039339953_p1888718226237"></a>默认值为“volcano”，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p17887322122310"><a name="zh-cn_topic_0000002039339953_p17887322122310"></a><a name="zh-cn_topic_0000002039339953_p17887322122310"></a><span id="zh-cn_topic_0000002039339953_ph6604131419312"><a name="zh-cn_topic_0000002039339953_ph6604131419312"></a><a name="zh-cn_topic_0000002039339953_ph6604131419312"></a>Ascend Operator</span>启用“gang”调度时所选择的调度器。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row11548142102118"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p15887122212235"><a name="zh-cn_topic_0000002039339953_p15887122212235"></a><a name="zh-cn_topic_0000002039339953_p15887122212235"></a>(.kind=="AscendJob").spec.runPolicy.schedulingPolicy.minAvailable</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1188782292317"><a name="zh-cn_topic_0000002039339953_p1188782292317"></a><a name="zh-cn_topic_0000002039339953_p1188782292317"></a>默认值为任务总副本数</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p660420141935"><a name="zh-cn_topic_0000002039339953_p660420141935"></a><a name="zh-cn_topic_0000002039339953_p660420141935"></a><span id="zh-cn_topic_0000002039339953_ph16604181418316"><a name="zh-cn_topic_0000002039339953_ph16604181418316"></a><a name="zh-cn_topic_0000002039339953_ph16604181418316"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="zh-cn_topic_0000002039339953_ph156050141033"><a name="zh-cn_topic_0000002039339953_ph156050141033"></a><a name="zh-cn_topic_0000002039339953_ph156050141033"></a>Volcano</span>时，任务运行总副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row6549642112110"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p108872226233"><a name="zh-cn_topic_0000002039339953_p108872226233"></a><a name="zh-cn_topic_0000002039339953_p108872226233"></a>(.kind=="AscendJob").spec.runPolicy.schedulingPolicy.queue</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p19887112222310"><a name="zh-cn_topic_0000002039339953_p19887112222310"></a><a name="zh-cn_topic_0000002039339953_p19887112222310"></a>默认值为“default”，用户需根据自身情况填写</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p3887422182315"><a name="zh-cn_topic_0000002039339953_p3887422182315"></a><a name="zh-cn_topic_0000002039339953_p3887422182315"></a><span id="zh-cn_topic_0000002039339953_ph10605114231"><a name="zh-cn_topic_0000002039339953_ph10605114231"></a><a name="zh-cn_topic_0000002039339953_ph10605114231"></a>Ascend Operator</span>启用“gang”调度生效，且调度器为<span id="zh-cn_topic_0000002039339953_ph1660520141632"><a name="zh-cn_topic_0000002039339953_ph1660520141632"></a><a name="zh-cn_topic_0000002039339953_ph1660520141632"></a>Volcano</span>时，任务所属队列。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1054916421215"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1388862217239"><a name="zh-cn_topic_0000002039339953_p1388862217239"></a><a name="zh-cn_topic_0000002039339953_p1388862217239"></a>（可选）(.kind=="AscendJob").spec.successPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul688815228238"></a><a name="zh-cn_topic_0000002039339953_ul688815228238"></a><ul id="zh-cn_topic_0000002039339953_ul688815228238"><li>默认值为空，若用户不填写该参数，则默认取空值。</li><li>AllWorkers</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p1688812219234"><a name="zh-cn_topic_0000002039339953_p1688812219234"></a><a name="zh-cn_topic_0000002039339953_p1688812219234"></a>表明任务成功的前提。空值代表只需要一个Pod成功，整个任务判定为成功。取值为“AllWorkers”表示所有Pod都成功，任务才判定为成功。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row15549114252110"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p178882222231"><a name="zh-cn_topic_0000002039339953_p178882222231"></a><a name="zh-cn_topic_0000002039339953_p178882222231"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.containers[0].name</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p28887225231"><a name="zh-cn_topic_0000002039339953_p28887225231"></a><a name="zh-cn_topic_0000002039339953_p28887225231"></a>ascend</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p0888172216231"><a name="zh-cn_topic_0000002039339953_p0888172216231"></a><a name="zh-cn_topic_0000002039339953_p0888172216231"></a>容器的名称必须是“ascend”。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row8549242152117"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p688811220233"><a name="zh-cn_topic_0000002039339953_p688811220233"></a><a name="zh-cn_topic_0000002039339953_p688811220233"></a>（可选）(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.containers[0].ports</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p188882022132320"><a name="zh-cn_topic_0000002039339953_p188882022132320"></a><a name="zh-cn_topic_0000002039339953_p188882022132320"></a>若用户未进行设置，系统默认填写以下参数：</p>
<a name="zh-cn_topic_0000002039339953_ul1488862272310"></a><a name="zh-cn_topic_0000002039339953_ul1488862272310"></a><ul id="zh-cn_topic_0000002039339953_ul1488862272310"><li>name: ascendjob-port</li><li>containerPort: 2222</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p2980183592716"><a name="p2980183592716"></a><a name="p2980183592716"></a>分布式训练集合通信端口。<span class="parmname" id="parmname1198063542717"><a name="parmname1198063542717"></a><a name="parmname1198063542717"></a>“name”</span>取值只能为<span class="parmvalue" id="parmvalue17980153515270"><a name="parmvalue17980153515270"></a><a name="parmvalue17980153515270"></a>“ascendjob-port”</span>，<span class="parmname" id="parmname8980135102711"><a name="parmname8980135102711"></a><a name="parmname8980135102711"></a>“containerPort”</span>用户可根据实际情况设置，若未进行设置则采用默认端口2222。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1054994210210"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p14889142214238"><a name="zh-cn_topic_0000002039339953_p14889142214238"></a><a name="zh-cn_topic_0000002039339953_p14889142214238"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.replicas</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul28892022132316"></a><a name="zh-cn_topic_0000002039339953_ul28892022132316"></a><ul id="zh-cn_topic_0000002039339953_ul28892022132316"><li>单机：1</li><li>分布式：N</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p788952218239"><a name="zh-cn_topic_0000002039339953_p788952218239"></a><a name="zh-cn_topic_0000002039339953_p788952218239"></a>N为任务副本数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row13550142102114"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p188915220236"><a name="zh-cn_topic_0000002039339953_p188915220236"></a><a name="zh-cn_topic_0000002039339953_p188915220236"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.containers[0].image</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p11889192242311"><a name="zh-cn_topic_0000002039339953_p11889192242311"></a><a name="zh-cn_topic_0000002039339953_p11889192242311"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p188942210238"><a name="zh-cn_topic_0000002039339953_p188942210238"></a><a name="zh-cn_topic_0000002039339953_p188942210238"></a>训练镜像名称，请根据实际修改。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row2256185652210"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p10889022132310"><a name="zh-cn_topic_0000002039339953_p10889022132310"></a><a name="zh-cn_topic_0000002039339953_p10889022132310"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec. nodeSelector.host-arch</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1889192292316"><a name="zh-cn_topic_0000002039339953_p1889192292316"></a><a name="zh-cn_topic_0000002039339953_p1889192292316"></a>Arm环境：<span id="ph27942615713"><a name="ph27942615713"></a><a name="ph27942615713"></a>huawei-arm</span></p>
<p id="zh-cn_topic_0000002039339953_p18889822162315"><a name="zh-cn_topic_0000002039339953_p18889822162315"></a><a name="zh-cn_topic_0000002039339953_p18889822162315"></a>x86_64环境：<span id="ph27919313716"><a name="ph27919313716"></a><a name="ph27919313716"></a>huawei-x86</span></p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p7889152202315"><a name="zh-cn_topic_0000002039339953_p7889152202315"></a><a name="zh-cn_topic_0000002039339953_p7889152202315"></a>需要运行训练任务的节点架构，请根据实际修改。</p>
<p id="zh-cn_topic_0000002039339953_p168891122112316"><a name="zh-cn_topic_0000002039339953_p168891122112316"></a><a name="zh-cn_topic_0000002039339953_p168891122112316"></a>分布式任务中，请确保运行训练任务的节点架构相同。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row13257165619223"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p11889152272316"><a name="zh-cn_topic_0000002039339953_p11889152272316"></a><a name="zh-cn_topic_0000002039339953_p11889152272316"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec. nodeSelector.accelerator-type</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul461118141037"></a><a name="zh-cn_topic_0000002039339953_ul461118141037"></a>
    <ul id="zh-cn_topic_0000002039339953_ul461118141037">
        <li><span id="zh-cn_topic_0000002039339953_ph136117141331"><a name="zh-cn_topic_0000002039339953_ph136117141331"></a><a name="zh-cn_topic_0000002039339953_ph136117141331"></a>Atlas 800 训练服务器（NPU满配）</span>：module</li>
        <li><span id="zh-cn_topic_0000002039339953_ph26111143315"><a name="zh-cn_topic_0000002039339953_ph26111143315"></a><a name="zh-cn_topic_0000002039339953_ph26111143315"></a>Atlas 800 训练服务器（NPU半配）</span>：half</li>
        <li>服务器（插<span id="zh-cn_topic_0000002039339953_ph26117141237"><a name="zh-cn_topic_0000002039339953_ph26117141237"></a><a name="zh-cn_topic_0000002039339953_ph26117141237"></a>Atlas 300T 训练卡</span>）：card</li>
        <li><span id="ph157633217501"><a name="ph157633217501"></a><a name="ph157633217501"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000002039339953_ph061115145320"><a name="zh-cn_topic_0000002039339953_ph061115145320"></a><a name="zh-cn_topic_0000002039339953_ph061115145320"></a>Atlas 900 A2 PoD 集群基础单元</span>：module-<span id="zh-cn_topic_0000002039339953_ph1661112141731"><a name="zh-cn_topic_0000002039339953_ph1661112141731"></a><a name="zh-cn_topic_0000002039339953_ph1661112141731"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_1"><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_1"></a>{xxx}</em></span>b-8</li>
        <li><span id="zh-cn_topic_0000002039339953_ph1161214145319"><a name="zh-cn_topic_0000002039339953_ph1161214145319"></a><a name="zh-cn_topic_0000002039339953_ph1161214145319"></a>Atlas 200T A2 Box16 异构子框</span>：module-<span id="zh-cn_topic_0000002039339953_ph116121514934"><a name="zh-cn_topic_0000002039339953_ph116121514934"></a><a name="zh-cn_topic_0000002039339953_ph116121514934"></a><em id="zh-cn_topic_0000001519959665_i1489729141619_2"><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a><a name="zh-cn_topic_0000001519959665_i1489729141619_2"></a>{xxx}</em></span>b-16</li>
        <li><span id="zh-cn_topic_0000002039339953_ph1514953013253"><a name="zh-cn_topic_0000002039339953_ph1514953013253"></a><a name="zh-cn_topic_0000002039339953_ph1514953013253"></a>A200T A3 Box8 超节点服务器</span>：module-a3-16</li><li>（可选）<span id="ph7730165573912"><a name="ph7730165573912"></a><a name="ph7730165573912"></a>Atlas 800 训练服务器（NPU满配）</span>可以省略该标签</li><li><span id="ph1973065563912"><a name="ph1973065563912"></a><a name="ph1973065563912"></a>Atlas 900 A3 SuperPoD 超节点</span>：module-a3-16-super-pod</li>
        <li>（可选）Atlas 350 标卡：350-Atlas-8、350-Atlas-16、350-Atlas-4p-8、350-Atlas-4p-16</li>
        <li>（可选）Atlas 850 系列硬件产品：850-Atlas-8p-8、850-SuperPod-Atlas-8</li>
        <li>（可选）Atlas 950 SuperPoD：950-SuperPod-Atlas-8</li>
    </ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><div class="p" id="zh-cn_topic_0000002039339953_p1989014223239"><a name="zh-cn_topic_0000002039339953_p1989014223239"></a><a name="zh-cn_topic_0000002039339953_p1989014223239"></a>根据需要运行训练任务的节点类型，选取不同的值。<div class="note" id="zh-cn_topic_0000002039339953_note1861316141738"><a name="zh-cn_topic_0000002039339953_note1861316141738"></a><a name="zh-cn_topic_0000002039339953_note1861316141738"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339953_p1027616512420"><a name="zh-cn_topic_0000002039339953_p1027616512420"></a><a name="zh-cn_topic_0000002039339953_p1027616512420"></a><span id="zh-cn_topic_0000002039339953_ph9014016509"><a name="zh-cn_topic_0000002039339953_ph9014016509"></a><a name="zh-cn_topic_0000002039339953_ph9014016509"></a>芯片型号的数值可通过<strong id="zh-cn_topic_0000001519959665_b168254314713"><a name="zh-cn_topic_0000001519959665_b168254314713"></a><a name="zh-cn_topic_0000001519959665_b168254314713"></a>npu-smi info</strong>命令查询，返回的“Name”字段对应信息为芯片型号，下文的{<em id="zh-cn_topic_0000001519959665_i1914312018209"><a name="zh-cn_topic_0000001519959665_i1914312018209"></a><a name="zh-cn_topic_0000001519959665_i1914312018209"></a>xxx</em>}即取“910”字符作为芯片型号数值。</span></p>
</div></div>
</div>
</td>
</tr>
<tr id="row17952121918262"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="p149528198263"><a name="p149528198263"></a><a name="p149528198263"></a><span id="ph12817122692620"><a name="ph12817122692620"></a><a name="ph12817122692620"></a>(.kind=="AscendJob").metadata.annotations.</span><span id="ph5346181126"><a name="ph5346181126"></a><a name="ph5346181126"></a>"</span><span id="ph19817116185120"><a name="ph19817116185120"></a><a name="ph19817116185120"></a>huawei.com/schedule_policy</span><span id="ph9572712181216"><a name="ph9572712181216"></a><a name="ph9572712181216"></a>"</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="p1877410425111"><a name="p1877410425111"></a><a name="p1877410425111"></a><span id="ph135426519519"><a name="ph135426519519"></a><a name="ph135426519519"></a>目前支持</span><a href="#table1120511613153">表2</a>中的配置。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p153031739509"><a name="p153031739509"></a><a name="p153031739509"></a>配置任务需要调度的AI芯片布局形态。<span id="zh-cn_topic_0000002511347099_ph204811934163414"><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a><a name="zh-cn_topic_0000002511347099_ph204811934163414"></a>Volcano</span>会根据该字段选择合适的调度策略。若不配置，则根据accelerator-type选择调度策略。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row142575564228"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p8290194641017"><a name="zh-cn_topic_0000002039339953_p8290194641017"></a><a name="zh-cn_topic_0000002039339953_p8290194641017"></a>(.kind=="AscendJob").metadata.annotations.sp-block</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p14755536454"><a name="zh-cn_topic_0000002039339953_p14755536454"></a><a name="zh-cn_topic_0000002039339953_p14755536454"></a>指定逻辑超节点芯片数量。</p>
<a name="zh-cn_topic_0000002039339953_ul10451144414619"></a><a name="zh-cn_topic_0000002039339953_ul10451144414619"></a><ul id="zh-cn_topic_0000002039339953_ul10451144414619"><li>单机时需要和任务请求的芯片数量一致。</li><li>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p1670155202912"><a name="p1670155202912"></a><a name="p1670155202912"></a>指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的亲和性调度。<span id="zh-cn_topic_0000002511347099_ph521204025916"><a name="zh-cn_topic_0000002511347099_ph521204025916"></a><a name="zh-cn_topic_0000002511347099_ph521204025916"></a>若用户未指定该字段，</span><span id="zh-cn_topic_0000002511347099_ph172121408590"><a name="zh-cn_topic_0000002511347099_ph172121408590"></a><a name="zh-cn_topic_0000002511347099_ph172121408590"></a>Volcano</span><span id="zh-cn_topic_0000002511347099_ph192121140135911"><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a><a name="zh-cn_topic_0000002511347099_ph192121140135911"></a>调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。</span></p>
<p id="p19701652112917"><a name="p19701652112917"></a><a name="p19701652112917"></a>了解详细说明请参见<a href="../basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a>。</p>
<div class="note" id="note47015215291"><a name="note47015215291"></a><a name="note47015215291"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><a name="zh-cn_topic_0000002511347099_ul546892712569"></a><ul id="zh-cn_topic_0000002511347099_ul546892712569"><li>仅支持在<span id="zh-cn_topic_0000002511347099_ph34244153594"><a name="zh-cn_topic_0000002511347099_ph34244153594"></a><a name="zh-cn_topic_0000002511347099_ph34244153594"></a>Atlas 900 A3 SuperPoD 超节点</span>中使用该字段。</li><li>使用了该字段后，不需要额外配置tor-affinity字段。</li><li>FAQ：<a href="../../faq.md#任务申请的总芯片数量为32sp-block设置为32可以正常训练sp-block设置为16无法完成训练训练容器报错提示初始化连接失败">任务申请的总芯片数量为32，sp-block设置为32可以正常训练，sp-block设置为16无法完成训练，训练容器报错提示初始化连接失败</a></li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row17257145682219"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1993916082411"><a name="zh-cn_topic_0000002039339953_p1993916082411"></a><a name="zh-cn_topic_0000002039339953_p1993916082411"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.containers[0].resources.requests.<span id="ph8846537141214"><a name="ph8846537141214"></a><a name="ph8846537141214"></a>"</span>huawei.com/Ascend910<span id="ph1636632921214"><a name="ph1636632921214"></a><a name="ph1636632921214"></a>"</span></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><div class="p" id="zh-cn_topic_0000002039339953_p361318141733"><a name="zh-cn_topic_0000002039339953_p361318141733"></a><a name="zh-cn_topic_0000002039339953_p361318141733"></a><span id="zh-cn_topic_0000002039339953_ph106131514134"><a name="zh-cn_topic_0000002039339953_ph106131514134"></a><a name="zh-cn_topic_0000002039339953_ph106131514134"></a>Atlas 800 训练服务器（NPU满配）</span>：<a name="zh-cn_topic_0000002039339953_ul126147145311"></a><a name="zh-cn_topic_0000002039339953_ul126147145311"></a><ul id="zh-cn_topic_0000002039339953_ul126147145311"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、8</li><li>分布式任务：1、2、4、8</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002039339953_p3614151416317"><a name="zh-cn_topic_0000002039339953_p3614151416317"></a><a name="zh-cn_topic_0000002039339953_p3614151416317"></a><span id="zh-cn_topic_0000002039339953_ph261416147313"><a name="zh-cn_topic_0000002039339953_ph261416147313"></a><a name="zh-cn_topic_0000002039339953_ph261416147313"></a>Atlas 800 训练服务器（NPU半配）</span>：<a name="zh-cn_topic_0000002039339953_ul1961418141132"></a><a name="zh-cn_topic_0000002039339953_ul1961418141132"></a><ul id="zh-cn_topic_0000002039339953_ul1961418141132"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4</li><li>分布式任务：1、2、4</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002039339953_p10614214538"><a name="zh-cn_topic_0000002039339953_p10614214538"></a><a name="zh-cn_topic_0000002039339953_p10614214538"></a>服务器（插<span id="zh-cn_topic_0000002039339953_ph13615131417315"><a name="zh-cn_topic_0000002039339953_ph13615131417315"></a><a name="zh-cn_topic_0000002039339953_ph13615131417315"></a>Atlas 300T 训练卡</span>）：<a name="zh-cn_topic_0000002039339953_ul1261519142311"></a><a name="zh-cn_topic_0000002039339953_ul1261519142311"></a><ul id="zh-cn_topic_0000002039339953_ul1261519142311"><li>单机单芯片任务：1</li><li>单机多芯片任务：2</li><li>分布式任务：2</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002039339953_p11615161416311"><a name="zh-cn_topic_0000002039339953_p11615161416311"></a><a name="zh-cn_topic_0000002039339953_p11615161416311"></a><span id="ph9683124520355"><a name="ph9683124520355"></a><a name="ph9683124520355"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000002039339953_ph46156141634"><a name="zh-cn_topic_0000002039339953_ph46156141634"></a><a name="zh-cn_topic_0000002039339953_ph46156141634"></a>Atlas 900 A2 PoD 集群基础单元</span>：<a name="zh-cn_topic_0000002039339953_ul1961514143314"></a><a name="zh-cn_topic_0000002039339953_ul1961514143314"></a><ul id="zh-cn_topic_0000002039339953_ul1961514143314"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8</li><li>分布式任务：1、2、3、4、5、6、7、8</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002039339953_p15616514739"><a name="zh-cn_topic_0000002039339953_p15616514739"></a><a name="zh-cn_topic_0000002039339953_p15616514739"></a><span id="zh-cn_topic_0000002039339953_ph161611419319"><a name="zh-cn_topic_0000002039339953_ph161611419319"></a><a name="zh-cn_topic_0000002039339953_ph161611419319"></a>Atlas 200T A2 Box16 异构子框</span>：<a name="zh-cn_topic_0000002039339953_ul661611418316"></a><a name="zh-cn_topic_0000002039339953_ul661611418316"></a><ul id="zh-cn_topic_0000002039339953_ul661611418316"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、3、4、5、6、7、8、10、12、14、16</li><li>分布式任务：1、2、3、4、5、6、7、8、10、12、14、16</li></ul>
</div>
<div class="p" id="zh-cn_topic_0000002039339953_p136171314938"><a name="zh-cn_topic_0000002039339953_p136171314938"></a><a name="zh-cn_topic_0000002039339953_p136171314938"></a><span id="zh-cn_topic_0000002039339953_ph161712141313"><a name="zh-cn_topic_0000002039339953_ph161712141313"></a><a name="zh-cn_topic_0000002039339953_ph161712141313"></a>Atlas 900 A3 SuperPoD 超节点</span>、<span id="zh-cn_topic_0000002039339953_ph188291824164611"><a name="zh-cn_topic_0000002039339953_ph188291824164611"></a><a name="zh-cn_topic_0000002039339953_ph188291824164611"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph83001907446"><a name="ph83001907446"></a><a name="ph83001907446"></a>Atlas 800T A3 超节点服务器</span>：<a name="zh-cn_topic_0000002039339953_ul261751412316"></a><a name="zh-cn_topic_0000002039339953_ul261751412316"></a><ul id="zh-cn_topic_0000002039339953_ul261751412316"><li>单机单芯片任务：1</li><li>单机多芯片任务：2、4、6、8、10、12、14、16</li><li>分布式任务：2、4、6、8、10、12、14、16</li><li>针对<span id="ph6583110111012"><a name="ph6583110111012"></a><a name="ph6583110111012"></a>Atlas 900 A3 SuperPoD 超节点</span>的逻辑超节点亲和任务：16</li></ul>
</div>
<div class="p"><a name=""></a><a name=""></a><span>Atlas 350 标卡（无互联节点内8卡）</span>：<a name=""></a><a name=""></a><ul><li>单机：1、2、3、4、5、6、7、8</li><li>分布式：1、2、3、4、5、6、7、8</li></ul>
</div>
<div class="p"><a name=""></a><a name=""></a><span>Atlas 350 标卡（无互联节点内16卡）</span>：<a name=""></a><a name=""></a><ul><li>单机：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li><li>分布式：1、2、3、4、5、6、7、8、9、10、11、12、13、14、15、16</li></ul>
</div>
<div class="p"><a name=""></a><a name=""></a><span>Atlas 350 标卡（4P mesh 8卡）</span>：<a name=""></a><a name=""></a><ul><li>单机（满足亲和性）：1、2、3、4、8</li><li>单机（不保证亲和性）：5、6、7</li><li>分布式（满足亲和性）：1、2、3、4、8</li><li>分布式（不保证亲和性）：5、6、7</li></ul>
</div>
<div class="p"><a name=""></a><a name=""></a><span>Atlas 350 标卡（4P mesh 16卡）</span>：<a name=""></a><a name=""></a><ul><li>单机（满足亲和性）：1、2、3、4、8、12、16</li><li>单机（不保证亲和性）：5、6、7、9、10、11、13、14、15</li><li>分布式（满足亲和性）：1、2、3、4、8、12、16</li><li>分布式（不保证亲和性）：5、6、7、9、10、11、13、14、15</li></ul>
</div>
<div class="p"><a name=""></a><a name=""></a><span>Atlas 850 系列硬件产品（普通集群）</span>：<a name=""></a><a name=""></a><ul><li>单机：1、2、4、8</li><li>分布式：1、2、4、8</li></ul>
</div>
<div class="p"><a name=""></a><a name=""></a><span>Atlas 850 系列硬件产品（超节点集群）</span>：<a name=""></a><a name=""></a><ul><li>单机：1、2、4、8（sp-block参数取值与其保持一致）</li><li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li></ul>
</div>
<div class="p"><a name=""></a><a name=""></a><span>Atlas 950 SuperPoD</span>：<a name=""></a><a name=""></a><ul><li>单机：1、2、3、4、5、6、7、8（sp-block参数取值与其保持一致）</li><li>分布式：8（sp-block参数取值需为8或8的倍数，且能被任务所需总卡数整除，且不能大于物理超节点大小）</li></ul>
</div>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 ">
    <p id="zh-cn_topic_0000002039339953_p3943130162412"><a name="zh-cn_topic_0000002039339953_p3943130162412"></a><a name="zh-cn_topic_0000002039339953_p3943130162412"></a>请求的NPU数量，请根据实际修改。</p>
    <div class="note">
        <span class="notetitle">[!NOTE] 说明</span>
        <div class="notebody">
            <p>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需修改参数名称为huawei.com/npu。</p>
        </div>
    </div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row11257156102214"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p894317013244"><a name="zh-cn_topic_0000002039339953_p894317013244"></a><a name="zh-cn_topic_0000002039339953_p894317013244"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.containers[0].env[name==ASCEND_VISIBLE_DEVICES].valueFrom.fieldRef.fieldPath</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p159431102243"><a name="zh-cn_topic_0000002039339953_p159431102243"></a><a name="zh-cn_topic_0000002039339953_p159431102243"></a>取值为metadata.annotations['huawei.com/AscendXXX']，其中XXX表示芯片的型号，支持的取值为910，310和310P。取值需要和环境上实际的芯片类型保持一致。</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 ">
    <p id="zh-cn_topic_0000002039339953_p136226142031"><a name="zh-cn_topic_0000002039339953_p136226142031"></a><a name="zh-cn_topic_0000002039339953_p136226142031"></a><span id="zh-cn_topic_0000002039339953_ph1062212140315"><a name="zh-cn_topic_0000002039339953_ph1062212140315"></a><a name="zh-cn_topic_0000002039339953_ph1062212140315"></a>Ascend Docker Runtime</span>会获取该参数值，用于给容器挂载相应类型的NPU。</p>          
    <div class="note" id="zh-cn_topic_0000002039339953_note462214141730"><a name="zh-cn_topic_0000002039339953_note462214141730"></a><a name="zh-cn_topic_0000002039339953_note462214141730"></a>
        <span class="notetitle">[!NOTE] 说明</span>
        <div class="notebody">
            <ul>
                <li>
                        <p id="zh-cn_topic_0000002039339953_p186225141637"><a name="zh-cn_topic_0000002039339953_p186225141637"></a><a name="zh-cn_topic_0000002039339953_p186225141637"></a>该参数只支持使用<span id="zh-cn_topic_0000002039339953_ph962251412315"><a name="zh-cn_topic_0000002039339953_ph962251412315"></a><a name="zh-cn_topic_0000002039339953_ph962251412315"></a>Volcano</span>调度器的整卡调度特性，使用静态vNPU调度和其他调度器的用户需要删除示例YAML中该参数的相关字段。</p>
                </li>
                <li>
                    <p>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD需配置为metadata.annotations['huawei.com/npu']。</p>
                </li>
            </ul>
        </div>
    </div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row925815602217"><td class="cellrowborder" rowspan="5" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p39441803247"><a name="zh-cn_topic_0000002039339953_p39441803247"></a><a name="zh-cn_topic_0000002039339953_p39441803247"></a>(.kind=="AscendJob").metadata.labels.fault-scheduling</p>
<p id="zh-cn_topic_0000002039339953_p54121202412"><a name="zh-cn_topic_0000002039339953_p54121202412"></a><a name="zh-cn_topic_0000002039339953_p54121202412"></a></p>
<p id="zh-cn_topic_0000002039339953_p13419162418"><a name="zh-cn_topic_0000002039339953_p13419162418"></a><a name="zh-cn_topic_0000002039339953_p13419162418"></a></p>
<p id="zh-cn_topic_0000002039339953_p23118240"><a name="zh-cn_topic_0000002039339953_p23118240"></a><a name="zh-cn_topic_0000002039339953_p23118240"></a></p>
<p id="zh-cn_topic_0000002039339953_p8211122417"><a name="zh-cn_topic_0000002039339953_p8211122417"></a><a name="zh-cn_topic_0000002039339953_p8211122417"></a></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p494450122411"><a name="zh-cn_topic_0000002039339953_p494450122411"></a><a name="zh-cn_topic_0000002039339953_p494450122411"></a>grace</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p1028610425357"><a name="p1028610425357"></a><a name="p1028610425357"></a>配置任务采用优雅删除模式，并在过程中先优雅删除原<span id="ph19623131417313"><a name="ph19623131417313"></a><a name="ph19623131417313"></a>Pod</span>，15分钟后若还未成功，使用强制删除原<span id="ph96231114734"><a name="ph96231114734"></a><a name="ph96231114734"></a>Pod</span>。</p>
<p id="p1462216142314"><a name="p1462216142314"></a><a name="p1462216142314"></a>进程级别重调度和进程级在线恢复场景，需将本参数配置为grace。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1258135615221"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p8944120192414"><a name="zh-cn_topic_0000002039339953_p8944120192414"></a><a name="zh-cn_topic_0000002039339953_p8944120192414"></a>force</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p562301420319"><a name="p562301420319"></a><a name="p562301420319"></a>配置任务采用强制删除模式，在过程中强制删除原<span id="ph19623151420318"><a name="ph19623151420318"></a><a name="ph19623151420318"></a>Pod</span>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1125885682215"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p4944603241"><a name="zh-cn_topic_0000002039339953_p4944603241"></a><a name="zh-cn_topic_0000002039339953_p4944603241"></a>off</p>
</td>
<td class="cellrowborder" rowspan="3" valign="top" headers="mcps1.2.4.1.2 "><p id="p26233141317"><a name="p26233141317"></a><a name="p26233141317"></a>该任务不使用断点续训特性，<span id="ph8623191418313"><a name="ph8623191418313"></a><a name="ph8623191418313"></a>K8s</span>的maxRetry仍然生效。</p>
<p id="p186239141631"><a name="p186239141631"></a><a name="p186239141631"></a></p>
<p id="p1623191419310"><a name="p1623191419310"></a><a name="p1623191419310"></a></p>
<p id="p10269114741310"><a name="p10269114741310"></a><a name="p10269114741310"></a></p>
<p id="p1326917477139"><a name="p1326917477139"></a><a name="p1326917477139"></a></p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1225812563228"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1794470102412"><a name="zh-cn_topic_0000002039339953_p1794470102412"></a><a name="zh-cn_topic_0000002039339953_p1794470102412"></a>无（无fault-scheduling字段）</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row870173811239"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p1494460142412"><a name="zh-cn_topic_0000002039339953_p1494460142412"></a><a name="zh-cn_topic_0000002039339953_p1494460142412"></a>其他值</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row97116382233"><td class="cellrowborder" rowspan="2" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p149441205245"><a name="zh-cn_topic_0000002039339953_p149441205245"></a><a name="zh-cn_topic_0000002039339953_p149441205245"></a>(.kind=="AscendJob").metadata.labels.fault-retry-times</p>
<p id="zh-cn_topic_0000002039339953_p18018142420"><a name="zh-cn_topic_0000002039339953_p18018142420"></a><a name="zh-cn_topic_0000002039339953_p18018142420"></a></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1394470122411"><a name="zh-cn_topic_0000002039339953_p1394470122411"></a><a name="zh-cn_topic_0000002039339953_p1394470122411"></a>0 &lt; fault-retry-times</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p13944180132415"><a name="zh-cn_topic_0000002039339953_p13944180132415"></a><a name="zh-cn_topic_0000002039339953_p13944180132415"></a>处理业务面故障，必须配置业务面可无条件重试的次数。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note146647015242"><a name="zh-cn_topic_0000002039339953_note146647015242"></a><div class="notebody"><a name="zh-cn_topic_0000002039339953_ul13624161415314"></a><a name="zh-cn_topic_0000002039339953_ul13624161415314"></a><ul id="zh-cn_topic_0000002039339953_ul13624161415314"><li>使用无条件重试功能需保证训练进程异常时会导致容器异常退出，若容器未异常退出则无法成功重试。</li><li>当前仅<span id="ph41689437364"><a name="ph41689437364"></a><a name="ph41689437364"></a>Atlas 800T A2 训练服务器</span>和<span id="zh-cn_topic_0000002039339953_ph166251314730"><a name="zh-cn_topic_0000002039339953_ph166251314730"></a><a name="zh-cn_topic_0000002039339953_ph166251314730"></a>Atlas 900 A2 PoD 集群基础单元</span>支持无条件重试功能。</li><li>进行进程级恢复时，将会触发业务面故障，如需使用进程级恢复，必须配置此参数。</li></ul>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row17113822319"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p159451022413"><a name="zh-cn_topic_0000002039339953_p159451022413"></a><a name="zh-cn_topic_0000002039339953_p159451022413"></a>无（无fault-retry-times）或0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p894515062415"><a name="zh-cn_topic_0000002039339953_p894515062415"></a><a name="zh-cn_topic_0000002039339953_p894515062415"></a>该任务不使用无条件重试功能，无法感知业务面故障，vcjob的maxRetry仍然生效。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row12722038182310"><td class="cellrowborder" rowspan="2" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p39451803247"><a name="zh-cn_topic_0000002039339953_p39451803247"></a><a name="zh-cn_topic_0000002039339953_p39451803247"></a>(.kind=="AscendJob").spec.runPolicy.backoffLimit</p>
<p id="zh-cn_topic_0000002039339953_p1999716016245"><a name="zh-cn_topic_0000002039339953_p1999716016245"></a><a name="zh-cn_topic_0000002039339953_p1999716016245"></a></p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1894519012417"><a name="zh-cn_topic_0000002039339953_p1894519012417"></a><a name="zh-cn_topic_0000002039339953_p1894519012417"></a>0 &lt; backoffLimit</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p1194560112412"><a name="zh-cn_topic_0000002039339953_p1194560112412"></a><a name="zh-cn_topic_0000002039339953_p1194560112412"></a>任务重调度次数。任务故障时，可以重调度的次数，当已经重调度次数与backoffLimit取值相同时，任务将不再进行重调度。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note1680005244"><a name="zh-cn_topic_0000002039339953_note1680005244"></a><div class="notebody"><p id="zh-cn_topic_0000002039339953_p59459062418"><a name="zh-cn_topic_0000002039339953_p59459062418"></a><a name="zh-cn_topic_0000002039339953_p59459062418"></a>同时配置了backoffLimit和fault-retry-times参数时，当已经重调度次数与backoffLimit或fault-retry-times取值有一个相同时，将不再进行重调度。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row167283811232"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p6945904245"><a name="zh-cn_topic_0000002039339953_p6945904245"></a><a name="zh-cn_topic_0000002039339953_p6945904245"></a>无（无backoffLimit）或backoffLimit ≤ 0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1594511016244"><a name="zh-cn_topic_0000002039339953_p1594511016244"></a><a name="zh-cn_topic_0000002039339953_p1594511016244"></a>不限制总重调度次数。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note176907017241"><a name="zh-cn_topic_0000002039339953_note176907017241"></a><div class="notebody"><p id="zh-cn_topic_0000002039339953_p159468016247"><a name="zh-cn_topic_0000002039339953_p159468016247"></a><a name="zh-cn_topic_0000002039339953_p159468016247"></a>若不配置backoffLimit，但是配置了fault-retry-times参数，则使用fault-retry-times的重调度次数。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row1372163816239"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p4946190192415"><a name="zh-cn_topic_0000002039339953_p4946190192415"></a><a name="zh-cn_topic_0000002039339953_p4946190192415"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.restartPolicy</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul59469052410"></a><a name="zh-cn_topic_0000002039339953_ul59469052410"></a><ul id="zh-cn_topic_0000002039339953_ul59469052410"><li>Never：从不重启</li><li>Always：总是重启</li><li>OnFailure：失败时重启</li><li>ExitCode：根据进程退出码决定是否重启Pod，错误码是1~127时不重启，128~255时重启Pod。</li></ul>
<div class="note" id="zh-cn_topic_0000002039339953_note8696110162419"><a name="zh-cn_topic_0000002039339953_note8696110162419"></a><a name="zh-cn_topic_0000002039339953_note8696110162419"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339953_p1694680112418"><a name="zh-cn_topic_0000002039339953_p1694680112418"></a><a name="zh-cn_topic_0000002039339953_p1694680112418"></a>vcjob类型的训练任务不支持ExitCode。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p1594690102418"><a name="zh-cn_topic_0000002039339953_p1594690102418"></a><a name="zh-cn_topic_0000002039339953_p1594690102418"></a>容器重启策略。当配置业务面故障无条件重试时，容器重启策略取值必须为“Never”。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row18731938162313"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p10946140152416"><a name="zh-cn_topic_0000002039339953_p10946140152416"></a><a name="zh-cn_topic_0000002039339953_p10946140152416"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.terminationGracePeriodSeconds</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002039339953_p1694690152418"><a name="zh-cn_topic_0000002039339953_p1694690152418"></a><a name="zh-cn_topic_0000002039339953_p1694690152418"></a>0 &lt; terminationGracePeriodSeconds &lt; <strong id="zh-cn_topic_0000002039339953_b09468052417"><a name="zh-cn_topic_0000002039339953_b09468052417"></a><a name="zh-cn_topic_0000002039339953_b09468052417"></a>grace-over-time</strong>参数取值</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002039339953_p1494610072416"><a name="zh-cn_topic_0000002039339953_p1494610072416"></a><a name="zh-cn_topic_0000002039339953_p1494610072416"></a>容器收到SIGTERM到被K8s强制停止经历的时间，该时间需要大于0且小于volcano-v<em id="zh-cn_topic_0000002039339953_i1394616092412"><a name="zh-cn_topic_0000002039339953_i1394616092412"></a><a name="zh-cn_topic_0000002039339953_i1394616092412"></a>{version}</em>.yaml文件中“<strong id="zh-cn_topic_0000002039339953_b794617013242"><a name="zh-cn_topic_0000002039339953_b794617013242"></a><a name="zh-cn_topic_0000002039339953_b794617013242"></a>grace-over-time</strong>”参数取值，同时还需要保证能够保存CKPT文件，请根据实际情况修改。具体说明请参见K8s官网<a href="https://kubernetes.io/zh/docs/concepts/containers/container-lifecycle-hooks/" target="_blank" rel="noopener noreferrer">容器生命周期回调</a>。</p>
<div class="note" id="zh-cn_topic_0000002039339953_note5717204249"><a name="zh-cn_topic_0000002039339953_note5717204249"></a><div class="notebody"><p id="zh-cn_topic_0000002039339953_p1894750112418"><a name="zh-cn_topic_0000002039339953_p1894750112418"></a><a name="zh-cn_topic_0000002039339953_p1894750112418"></a>只有当fault-scheduling配置为grace时，该字段才生效；fault-scheduling配置为force时，该字段无效。</p>
</div></div>
</td>
</tr>
<tr id="zh-cn_topic_0000002039339953_row15963544152013"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002039339953_p7418049182319"><a name="zh-cn_topic_0000002039339953_p7418049182319"></a><a name="zh-cn_topic_0000002039339953_p7418049182319"></a>(.kind=="AscendJob").spec.replicaSpecs.{Master|Scheduler|Worker}.template.spec.hostNetwork</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><a name="zh-cn_topic_0000002039339953_ul960434424111"></a><a name="zh-cn_topic_0000002039339953_ul960434424111"></a><ul id="zh-cn_topic_0000002039339953_ul960434424111"><li>true：使用HostIP创建Pod。</li><li>false：不使用HostIP创建Pod。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><a name="zh-cn_topic_0000002039339953_ul14611159182815"></a><a name="zh-cn_topic_0000002039339953_ul14611159182815"></a><ul id="zh-cn_topic_0000002039339953_ul14611159182815"><li>当集群规模较大（节点数量&gt;1000时），推荐使用HostIP创建Pod。</li><li>不传入此参数时，默认不使用HostIP创建Pod。<div class="note" id="zh-cn_topic_0000002039339953_note1423653119592"><a name="zh-cn_topic_0000002039339953_note1423653119592"></a><a name="zh-cn_topic_0000002039339953_note1423653119592"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002039339953_p461933317584"><a name="zh-cn_topic_0000002039339953_p461933317584"></a><a name="zh-cn_topic_0000002039339953_p461933317584"></a>当HostNetwork取值为true时，若当前任务YAML挂载了<span id="ph01944310814"><a name="ph01944310814"></a><a name="ph01944310814"></a>RankTable</span>文件路径，则可以通过在训练脚本中解析<span id="ph158525327162"><a name="ph158525327162"></a><a name="ph158525327162"></a>RankTable</span>文件获取Pod的hostIP来实现建链。若任务YAML未挂载<span id="ph094714211613"><a name="ph094714211613"></a><a name="ph094714211613"></a>RankTable</span>文件路径，则与原始保持一致，使用serviceIP来实现建链。</p>
</div></div>
</li></ul>
</td>
</tr>
<tr id="row13351447164012"><td class="cellrowborder" valign="top" width="25.042504250425047%" headers="mcps1.2.4.1.1 "><p id="p6351124764011"><a name="p6351124764011"></a><a name="p6351124764011"></a>(.kind=="AscendJob").metadata.annotations.wait-reschedule-timeout</p>
</td>
<td class="cellrowborder" valign="top" width="24.76247624762476%" headers="mcps1.2.4.1.2 "><p id="p1235154718404"><a name="p1235154718404"></a><a name="p1235154718404"></a>30~270</p>
</td>
<td class="cellrowborder" valign="top" width="50.1950195019502%" headers="mcps1.2.4.1.3 "><p id="p8351747134014"><a name="p8351747134014"></a><a name="p8351747134014"></a>进程级别重调度处理时等待故障节点重调度的超时时间，单位为秒，默认值为270。</p>
</td>
</tr>
</tbody>
</table>

**表 2**  huawei.com/schedule\_policy配置说明

<a name="table1120511613153"></a>

|配置|说明|
|--|--|
|chip4-node8|1个节点8张芯片，每4个芯片形成1个互联环。例如，Atlas 800 训练服务器（型号 9000）/Atlas 800 训练服务器（型号 9010）芯片的整模块场景/Atlas 350 标卡共8张卡，每4张卡通过UB扣板连接。|
|chip1-node2|1个节点2张芯片。例如，Atlas 300T 训练卡的插卡场景，1张卡最多插1个芯片，1个节点最多插2张卡。|
|chip4-node4|1个节点4张芯片，形成1个互联环。例如，Atlas 800 训练服务器（型号 9000）/Atlas 800 训练服务器（型号 9010）芯片的半配场景。|
|chip8-node8|1个节点8张卡，8张卡都在1个互联环上。例如，Atlas 800T A2 训练服务器 /Atlas 850 系列硬件产品。|
|chip8-node16|1个节点16张卡，每8张卡在1个互联环上。例如，Atlas 200T A2 Box16 异构子框。|
|chip2-node8|1个节点8张卡，每2张卡在1个互联环上。|
|chip2-node16|1个节点16张卡，每2张卡在1个互联环上。例如，Atlas 800T A3 超节点服务器。|
|chip2-node8-sp|1个节点8张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，Atlas 9000 A3 SuperPoD 集群算力系统。|
|chip2-node16-sp|1个节点16张卡，每2张卡在1个互联环上，多个服务器形成超节点。例如，Atlas 900 A3 SuperPoD 超节点。|
|chip4-node16|1个节点16张卡，每4张卡都在1个互联环上。例如，Atlas 350 标卡共16张卡，每4张卡通过UB扣板连接。|
|chip1-node8|1个节点8张卡，每张卡之间无互联。例如，Atlas 350 标卡共8张卡，每张卡之间无互联。|
|chip1-node16|1个节点16张卡，每张卡之间无互联。例如，Atlas 350 标卡共16张卡，每张卡之间无互联。|
|chip8-node8-sp|1个节点8张卡，8张卡都在1个互联环上，多个服务器形成超节点。例如，Atlas 850 系列硬件产品 超节点服务器。|
|chip8-node8-ra64-sp|1个节点8张卡，8张卡都在1个互联环上，64个节点组成一个计算框，多个框形成超节点。例如，Atlas 950 SuperPoD。|
|chip1-softShareDev|软切分虚拟化专用调度策略。|
|multilevel|多级调度场景使用，多级调度的详细使用方法请参见[多级调度](../basic_scheduling/05_multi_level_scheduling.md)。|

## 任务YAML配置示例<a name="ZH-CN_TOPIC_0000002511346461"></a>

重调度模式和优雅容错模式可参见如下[操作步骤](#zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section18181655154219)配置示例。当**subHealthyStrategy**取值为graceExit时，需要参见[（可选）配置亚健康故障保存临终遗言](#zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section1048332432310)完成启动脚本与任务YAML的适配，以确保任务因亚健康故障被重调度前能够正常保存CKPT文件。

**前提条件<a name="zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section7585519135117"></a>**

用户已创建[hccl.json](../../api/hccl.json_file_description.md)文件的具体挂载路径，详细操作步骤请参见[Ascend Operator](../../installation_guide/03_installation/manual_installation/08_ascend_operator.md)中的“步骤4”。

**操作步骤<a name="zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section18181655154219"></a>**

1. 将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。
    - 以a800\_AscendJob\_<i>\{xxx\}</i>b.yaml为例，在一台Atlas 200T A2 Box16 异构子框节点创建**分布式训练**任务，任务使用2\*4个芯片，修改示例如下。

        ```Yaml
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-mindspore
          labels:
            framework: mindspore  # 训练框架名称
            fault-scheduling: "grace"     # 开启优雅删除模式
            ring-controller.atlas: ascend-{xxx}b
            fault-retry-times: "3"            # 开启业务面故障无条件重试能力，同时需要将restartPolicy取值设置为Never
            tor-affinity: "normal-schema" #该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不使用该特性。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
            pod-rescheduling: "on"     # 开启Pod级别重调度
            subHealthyStrategy: "ignore"  # 忽略健康状态为亚健康的节点，后续任务在亲和性调度上不优先调度该节点
        spec:
          schedulerName: volcano    # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            backoffLimit: 3      # 任务重调度次数
            schedulingPolicy:
              minAvailable: 3       # 任务总副本数
              queue: default     # 任务所属队列
          successPolicy: AllWorkers  # 任务成功的前提
          replicaSpecs:
            Scheduler:
              replicas: 1            #只能为1
              restartPolicy:  Never   #容器重启策略
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
                spec:
                  terminationGracePeriodSeconds: 360  #容器收到SIGTERM到被K8s强制停止经历的时间
                  nodeSelector:                       
                    host-arch: huawei-x86          # Atlas 200T A2 Box16 异构子框只有x86_64架构
                    accelerator-type: module-{xxx}b-16   # 节点类型
                  containers:
                  - name: ascend     # 不能修改
        ...
                    ports:                     # 可选，分布式训练集合通信端口
                      - containerPort: 2222    
                        name: ascendjob-port 
                    volumeMounts:
        ...
          
            Worker:
              replicas: 2
              restartPolicy: Never  # 容器重启策略
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
                spec:
                  terminationGracePeriodSeconds: 360   #容器收到SIGTERM到被K8s强制停止经历的时间
                  affinity:
        ...
                  nodeSelector:           
                    host-arch: huawei-x86      # Atlas 200T A2 Box16 异构子框只有x86_64架构
                    accelerator-type: module-{xxx}b-16   # 节点类型
                  containers:
                  - name: ascend      # 不能修改
        ...
                    env:
                    - name: ASCEND_VISIBLE_DEVICES
                      valueFrom:
                        fieldRef:
                          fieldPath: metadata.annotations['huawei.com/Ascend910']         # 需要和下面resources和requests保持一致     
        ...
        
                    ports:        # 可选，分布式训练集合通信端口
                      - containerPort: 2222    
                        name: ascendjob-port  
                    resources:
                      limits:
                        huawei.com/Ascend910: 4      # 需要的NPU芯片个数为4
                      requests:
                        huawei.com/Ascend910: 4       # 与limits取值一致
        ```

    - 以a800\_vcjob.yaml为例，在一台Atlas 800 训练服务器节点创建**单机训练**任务，任务使用8个芯片，修改示例如下。

        ```Yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
        ...
          labels:
            ring-controller.atlas: ascend-910  # 标识产品类型
        ...
        ---
        apiVersion: batch.volcano.sh/v1alpha1   # 不可修改。必须使用Volcano的API。
        kind: Job                               # 目前只支持Job类型
        metadata:
          name: mindx-dls-test                  # 任务名，可自定义
          labels:
            ring-controller.atlas: ascend-910   
            fault-scheduling: "grace"        # 开启优雅删除模式
            fault-retry-times: "3"            # 开启业务面故障无条件重试能力，同时需要将restartPolicy取值设置为Never；并将policies的event设置为PodFailed，action设置为Ignore
            tor-affinity: "normal-schema" #该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不使用该特性。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
            pod-rescheduling: "on"     # 开启Pod级别重调度
            subHealthyStrategy: "ignore"     # 忽略健康状态为亚健康的节点，后续任务在亲和性调度上不优先调度该节点
        ...
        spec:
          policies:  # 使用重调度功能时，无需修改 policies 内容
            - event: PodFailed
              action: Ignore
        ...
          minAvailable: 1                  # 单机为1
        ...
          maxRetry: 3              # 重调度次数
        ...
          - name: "default-test"
              replicas: 1                  # 单机为1
              template:
                metadata:
        ...
                spec:
                  terminationGracePeriodSeconds: 360  #容器收到SIGTERM到被K8s强制停止经历的时间
        ...
                    env:
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources和requests保持一致
        ...
                    resources:  
                      requests:
                        huawei.com/Ascend910: 8          # 需要的NPU芯片个数为8。可在下方添加行，配置memory、cpu等资源
                      limits:
                        huawei.com/Ascend910: 8          # 目前需要和上面requests保持一致
        ...
                    nodeSelector:
                      host-arch: huawei-arm               # 可选值，根据实际情况填写
                      accelerator-type: module      #调度到Atlas 800 训练服务器节点上
        ...
                restartPolicy: Never   # 容器重启策略
        ```

2. 配置MindIO的通信地址。在代码中新增以下内容。

    ```Yaml
    ...
       Master:
    ...
                env:        
                  - name: POD_IP
                    valueFrom:
                      fieldRef:
                        fieldPath: status.podIP             # 用于MindIO通信，如果不配置此参数会影响训练任务的正常拉起。
    ```

3. （可选）如果开启了临终遗言，需要在训练YAML中增加临终遗言通信的端口信息，以pytorch\_multinodes\_acjob\_<i>\{xxx\}</i>b.yaml为例，新增以下加粗内容。

    <pre codetype="yaml">
    ...
       Master:
    ...
              <strong>env:</strong>
                  <strong>- name: TTP_PORT</strong>                  
                    <strong>value: "8000"     # 用于临终遗言通信，请注意上下保持一致</strong>
    ...
                ports:                         
                    - containerPort: 2222        
                      name: ascendjob-port       
                    <strong>- containerPort: 8000     # 用于临终遗言通信，请注意上下保持一致</strong>
                      <strong>name: ttp-port</strong>
                    <strong>- containerPort: 9601     # TaskD Pod间通信端口</strong>
                      <strong>name: taskd-port</strong>
    ...
       Worker:
    ...
              <strong>env:</strong>
                  <strong>- name: TTP_PORT</strong>                  
                    <strong>value: "8000"            # 用于临终遗言通信，请注意上下保持一致</strong>
    ...
                ports:                          
                    - containerPort: 2222         
                      name: ascendjob-port       
                    <strong>- containerPort: 8000     # 用于临终遗言通信，请注意上下保持一致</strong>
                      <strong>name: ttp-port</strong>
                    <strong>- containerPort: 9601     # TaskD Pod间通信端口</strong>
                      <strong>name: taskd-port</strong>
    
    ...</pre>

4. （可选）如果使用临终遗言和进程级恢复，需要在训练YAML中增加临终遗言通信的端口信息和进程级恢复开关等信息，以pytorch\_multinodes\_acjob\_<i>\{xxx\}</i>b.yaml为例，新增以下加粗内容。

    <pre codetype="yaml">
    ...
      labels:    
           framework: pytorch   
           ring-controller.atlas: ascend-{xxx}b    
           <strong>fault-scheduling: "grace"</strong>    
           <strong>fault-retry-times: "10"   // 开启无条件重试</strong>
           <strong>pod-rescheduling: "on"   // 开启Pod级重调度</strong>
           tor-affinity: "null" # 该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不适用。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
    ...
      annotations:  
         ...  
         <strong>recover-strategy: "recover,dump"</strong>
      replicaSpecs:    
          Master:     
            replicas: 1      
            <strong>restartPolicy: Never</strong>      
            template:        
                metadata:
    ...
               <strong>- name: TTP_PORT</strong>
                 <strong>value: "8000"  # 用于MindIO通信，请注意上下保持一致</strong>
            command:                           # training command, which can be modified             
              - /bin/bash              
              - -c            
            args:
              - | 
                cd /job/code; 
                chmod +x scripts/train_start.sh; 
                bash scripts/train_start.sh
             ports:                          # default value 
               - containerPort: 2222 
                 name: ascendjob-port              
               <strong>- containerPort: 8000    # 用于MindIO通信，请注意上下保持一致</strong>
                 <strong>name: ttp-port</strong>
               <strong>- containerPort: 9601    # TaskD Pod间通信端口</strong>
                 <strong>name: taskd-port</strong>
    ...
    
    ...
      replicaSpecs:    
          Worker:     
            replicas: 1      
            <strong>restartPolicy: Never</strong>      
            template:        
                metadata:
    ...
                <strong>- name: TTP_PORT</strong>
                <strong>value: "8000"  # 用于MindIO通信，请注意上下保持一致</strong>
            command:                           # training command, which can be modified             
              - /bin/bash              
              - -c            
            args:
              - | 
                cd /job/code; 
                chmod +x scripts/train_start.sh; 
                bash scripts/train_start.sh
             ports:                          # default value 
               - containerPort: 2222 
                 name: ascendjob-port              
               <strong>- containerPort: 8000    # 用于MindIO通信，请注意上下保持一致</strong>
                 <strong>name: ttp-port</strong>
               <strong>- containerPort: 9601    # TaskD Pod间通信端口</strong>
                 <strong>name: taskd-port</strong>
    ...</pre>

5. 使用断点续训功能，建议扩展内存，请按注释添加参数，示例如下。

    ```Yaml
    ...
              volumeMounts:                             #断点续训扩容
             - name: shm
               mountPath: /dev/shm
            volumes:
            - name: shm
              emptyDir:
                medium: Memory
                sizeLimit: 16Gi
    ...
    ```

6. 若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu”和“memory”参数和对应的参数值，具体数值请根据实际情况配置。

    ```Yaml
    ...
              resources:  
                requests:
                  huawei.com/Ascend910: 8
                  cpu: 100m               
                  memory: 100Gi           
                limits:
                  huawei.com/Ascend910: 8
                  cpu: 100m
                  memory: 100Gi
    ...
    ```

7. 修改训练脚本、代码的挂载路径。

    从昇腾镜像仓库拉取的基础镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。

    ```Yaml
              volumeMounts:
              - name: ascend-910-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径        
    ```

8. （可选）如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及TensorFlow框架模型文件）、启动脚本相对代码目录的路径（PyTorch命令参数不涉及启动脚本）。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。

    >[!NOTE] 
    >使用**优雅容错模式**可跳过该步骤。

    - **TensorFlow命令参数**

        ```shell
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ tensorflow/resnet_ctl_imagenet_main.py --data_dir=/job/data/imagenet_TF --distribution_strategy=one_device --use_tf_while_loop=true --epochs_between_evals=1 --skip_eval --enable_checkpoint_and_export;"
        ...
        ```

    - **PyTorch命令参数**

        ```shell
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --lr=1.6 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=1024 --resume=true;"
        ...
        ```

    - 使用**MindSpore架构**的模型，包括ResNet50模型和Pangu\_alpha模型需要跳过此步骤。

9. 选择存储方式。
    - （可选）NFS场景需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

        >[!NOTE] 
        >请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。

        ```Yaml
        ...
                  volumeMounts:
                  - name: ascend-910-config
                    mountPath: /user/serverid/devindex/config
                  - name: code
                    mountPath: /job/code                     # 容器中训练脚本路径
                  - name: data
                    mountPath: /job/data                      # 容器中训练数据集路径
                  - name: output
                    mountPath: /job/output                    # 容器中训练输出路径
        ...
                   # 可选，使用Ascend Operator组件为训练任务生成RankTable文件，需要新增以下字段，设置容器中hccl.json文件保存路径，该路径不可修改。
                  - name: ranktable        
                    mountPath: /user/serverid/devindex/config
        ...
                volumes:
        ...
                - name: code
                  nfs:
                    server: 127.0.0.1        # NFS服务器IP地址
                    path: "xxxxxx"           # 配置训练脚本路径
                - name: data
                  nfs:
                    server: 127.0.0.1
                    path: "xxxxxx"           # 配置训练集路径
                - name: output
                  nfs:
                    server: 127.0.0.1
                    path: "xxxxxx"           # 设置脚本相关模型的保存路径
        ...
                   # 可选，使用组件为PyTorch框架生成RankTable文件，需要新增以下字段，设置hccl.json文件保存路径
                - name: ranktable         #请勿修改此参数的默认值，Ascend Operator会用于检查是否开启文件挂载hccl.json。
                  hostPath:                    #请使用hostpath挂载或NFS挂载
                    path: /user/mindx-dl/ranktable/default.default-test-pytorch   # 共享存储或者本地存储路径，/user/mindx-dl/ranktable/为前缀路径，必须和Ascend Operator挂载的Ranktable根目录保持一致。default.default-test-pytorch为后缀路径，建议改为:namespace.job-name。
        ...
        ```

    - （可选）如果使用本地存储的挂载方式，需要将YAML中的NFS方式改为hostPath。

        ```Yaml
                  volumes:
                  - name: code
                    hostPath:                                                        # 修改为本地存储
                      path: "/data/atlas_dls/code/resnet/"
                  - name: data
                    hostPath:                                                        # 修改为本地存储
                      path: "/data/atlas_dls/public/dataset/"
                  - name: output
                    hostPath:                                                        # 修改为本地存储
                      path: "/data/atlas_dls/output/"
                  - name: ascend-driver
                    hostPath:
                      path: /usr/local/Ascend/driver
                  - name: dshm
                    emptyDir:
                      medium: Memory
                  - name: localtime
                    hostPath:
                      path: /etc/localtime
        ```

**（可选）配置亚健康故障保存临终遗言<a name="zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section1048332432310"></a>**

如果希望任务发生亚健康故障时保存临终遗言，需修改任务YAML，配置亚健康策略为“graceExit”，故障恢复策略为“dump”，其余启动脚本、任务YAML配置可参见[配置临终CKPT保存](./05_configuring_training_recovery.md#配置临终ckpt保存)修改。此功能需确保TaskD和ClusterD可以正常使用。

```Yaml
...  
  labels:  
     ... 
     subHealthyStrategy: "graceExit"  # 配置亚健康策略
...
  annotations:  
    ...  
    recover-strategy: "dump"  # 任务可用恢复策略为保存临终遗言
...
```
