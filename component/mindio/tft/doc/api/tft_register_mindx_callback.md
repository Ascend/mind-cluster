# tft\_register\_mindx\_callback<a name="ZH-CN_TOPIC_0000002511721164"></a>

## 接口功能<a name="zh-cn_topic_0000001975861586_section1346094194320"></a>

提供给MindCluster调用，向MindIO TFT注册修复流程回调函数接口。

## 接口格式<a name="zh-cn_topic_0000001975861586_section1272104915439"></a>

```
mindio_ttp.controller_ttp.tft_register_mindx_callback(action: str, func: Callable)
```

## 接口参数<a name="zh-cn_topic_0000001975861586_section1148355114320"></a>

<a name="zh-cn_topic_0000001975861586_table173251148163716"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001975861586_row14325848123711"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001975861586_p6325184810378"><a name="zh-cn_topic_0000001975861586_p6325184810378"></a><a name="zh-cn_topic_0000001975861586_p6325184810378"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001975861586_p1132534853714"><a name="zh-cn_topic_0000001975861586_p1132534853714"></a><a name="zh-cn_topic_0000001975861586_p1132534853714"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="27.73%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001975861586_p17325144811374"><a name="zh-cn_topic_0000001975861586_p17325144811374"></a><a name="zh-cn_topic_0000001975861586_p17325144811374"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="32.269999999999996%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001975861586_p1325048183713"><a name="zh-cn_topic_0000001975861586_p1325048183713"></a><a name="zh-cn_topic_0000001975861586_p1325048183713"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row364941564412"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p0846164618288"><a name="p0846164618288"></a><a name="p0846164618288"></a>action</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p15292947105015"><a name="p15292947105015"></a><a name="p15292947105015"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="27.73%" headers="mcps1.1.5.1.3 "><p id="p1529013472508"><a name="p1529013472508"></a><a name="p1529013472508"></a>回调函数要注册的动作名。</p>
</td>
<td class="cellrowborder" valign="top" width="32.269999999999996%" headers="mcps1.1.5.1.4 "><p id="p7137183212291"><a name="p7137183212291"></a><a name="p7137183212291"></a>str，支持的动作名如下：</p>
<a name="ul5436026132911"></a><a name="ul5436026132911"></a><ul id="ul5436026132911"><li>report_fault_ranks</li><li>report_stop_complete</li><li>report_strategies</li><li>report_result</li></ul>
</td>
</tr>
<tr id="row19917102062820"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p6917182017282"><a name="p6917182017282"></a><a name="p6917182017282"></a>func</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p491732017282"><a name="p491732017282"></a><a name="p491732017282"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="27.73%" headers="mcps1.1.5.1.3 "><p id="p14917172018284"><a name="p14917172018284"></a><a name="p14917172018284"></a>要注册的函数。</p>
</td>
<td class="cellrowborder" valign="top" width="32.269999999999996%" headers="mcps1.1.5.1.4 "><p id="p19917020182820"><a name="p19917020182820"></a><a name="p19917020182820"></a>回调函数，不为空，回调函数入参详情请参见<a href="#table1543916174616">表1</a> ~ <a href="#table459011249361">表4</a>。</p>
</td>
</tr>
</tbody>
</table>

**表 1**  action为report\_fault\_ranks时回调函数参数

<a name="table1543916174616"></a>
<table><thead align="left"><tr id="row04391172064"><th class="cellrowborder" valign="top" width="19.98%" id="mcps1.2.5.1.1"><p id="p943912177615"><a name="p943912177615"></a><a name="p943912177615"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20.02%" id="mcps1.2.5.1.2"><p id="p143914170610"><a name="p143914170610"></a><a name="p143914170610"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.3"><p id="p04391317161"><a name="p04391317161"></a><a name="p04391317161"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p3439131716610"><a name="p3439131716610"></a><a name="p3439131716610"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row743917171568"><td class="cellrowborder" valign="top" width="19.98%" headers="mcps1.2.5.1.1 "><p id="p17439117667"><a name="p17439117667"></a><a name="p17439117667"></a>error_rank_dict</p>
</td>
<td class="cellrowborder" valign="top" width="20.02%" headers="mcps1.2.5.1.2 "><p id="p124390171366"><a name="p124390171366"></a><a name="p124390171366"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p333834413245"><a name="p333834413245"></a><a name="p333834413245"></a>发生故障的卡信息。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1949974519218"><a name="p1949974519218"></a><a name="p1949974519218"></a>&lt;int key, int errorType&gt;字典：</p>
<a name="ul736975113211"></a><a name="ul736975113211"></a><ul id="ul736975113211"><li>key为故障卡的rank号。</li><li>errorType为故障类型：<a name="ul16212145011220"></a><a name="ul16212145011220"></a><ul id="ul16212145011220"><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul>
</li></ul>
</td>
</tr>
</tbody>
</table>

**表 2**  action为report\_stop\_complete时回调函数参数

<a name="table1512319011317"></a>
<table><thead align="left"><tr id="row01233013316"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p11231014312"><a name="p11231014312"></a><a name="p11231014312"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p20123404313"><a name="p20123404313"></a><a name="p20123404313"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.3"><p id="p6123402312"><a name="p6123402312"></a><a name="p6123402312"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p6123140133111"><a name="p6123140133111"></a><a name="p6123140133111"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row19124701316"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p82666505341"><a name="p82666505341"></a><a name="p82666505341"></a>code</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p526675015343"><a name="p526675015343"></a><a name="p526675015343"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p1026655011340"><a name="p1026655011340"></a><a name="p1026655011340"></a>action执行结果。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><a name="ul1880654263916"></a><a name="ul1880654263916"></a><ul id="ul1880654263916"><li>0：成功。</li><li>400：普通错误。</li><li>401：<span id="ph1975922310265"><a name="ph1975922310265"></a><a name="ph1975922310265"></a>MindCluster</span> task id不存在。</li><li>402：模型错误。</li><li>403：顺序错误。</li><li>404：Processor未全部准备就绪。</li></ul>
</td>
</tr>
<tr id="row169413541344"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p136103017352"><a name="p136103017352"></a><a name="p136103017352"></a>msg</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p13941135463412"><a name="p13941135463412"></a><a name="p13941135463412"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p59411954193419"><a name="p59411954193419"></a><a name="p59411954193419"></a>训练是否停止消息。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p5941185410340"><a name="p5941185410340"></a><a name="p5941185410340"></a>str。</p>
</td>
</tr>
<tr id="row1885085633414"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p1253815341354"><a name="p1253815341354"></a><a name="p1253815341354"></a>error_rank_dict</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p485085620346"><a name="p485085620346"></a><a name="p485085620346"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p171781518354"><a name="p171781518354"></a><a name="p171781518354"></a>发生故障的卡信息。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p4178751123518"><a name="p4178751123518"></a><a name="p4178751123518"></a>&lt;int key, int errorType&gt;字典：</p>
<a name="ul171781151183510"></a><a name="ul171781151183510"></a><ul id="ul171781151183510"><li>key为故障卡的rank号</li><li>errorType为故障类型：<a name="ul15178051153516"></a><a name="ul15178051153516"></a><ul id="ul15178051153516"><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul>
</li></ul>
</td>
</tr>
</tbody>
</table>

**表 3**  action为report\_strategies时回调函数参数

<a name="table116012273617"></a>
<table><thead align="left"><tr id="row1460172213615"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p36042213617"><a name="p36042213617"></a><a name="p36042213617"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p1260162215363"><a name="p1260162215363"></a><a name="p1260162215363"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.3"><p id="p5604227367"><a name="p5604227367"></a><a name="p5604227367"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p860122243619"><a name="p860122243619"></a><a name="p860122243619"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row1760142263618"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p12601822113615"><a name="p12601822113615"></a><a name="p12601822113615"></a>error_rank_dict</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p166162213614"><a name="p166162213614"></a><a name="p166162213614"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p16113222368"><a name="p16113222368"></a><a name="p16113222368"></a>发生故障的卡信息。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p8611622183615"><a name="p8611622183615"></a><a name="p8611622183615"></a>&lt;int key, int errorType&gt;字典：</p>
<a name="ul106172216368"></a><a name="ul106172216368"></a><ul id="ul106172216368"><li>key为故障卡的rank号。</li><li>errorType为故障类型：<a name="ul36182215363"></a><a name="ul36182215363"></a><ul id="ul36182215363"><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul>
</li></ul>
</td>
</tr>
<tr id="row13416359173617"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p134160590368"><a name="p134160590368"></a><a name="p134160590368"></a>strategy_list</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1441695915363"><a name="p1441695915363"></a><a name="p1441695915363"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p511513405373"><a name="p511513405373"></a><a name="p511513405373"></a>基于当前可用的副本信息，<span id="ph144711415443"><a name="ph144711415443"></a><a name="ph144711415443"></a>MindIO TFT</span>支持的修复策略列表。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p243564419268"><a name="p243564419268"></a><a name="p243564419268"></a>list，支持的修复策略可选值如下（str）：</p>
<a name="ul12287105882616"></a><a name="ul12287105882616"></a><ul id="ul12287105882616"><li>retry：执行UCE修复。</li><li>recover：执行ARF修复。</li><li>dump：执行临终遗言。</li><li>exit：退出。</li></ul>
</td>
</tr>
</tbody>
</table>

**表 4**  action为report\_result时回调函数参数

<a name="table459011249361"></a>
<table><thead align="left"><tr id="row18590162413365"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p359072419367"><a name="p359072419367"></a><a name="p359072419367"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p359072414369"><a name="p359072414369"></a><a name="p359072414369"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.3"><p id="p14590102414368"><a name="p14590102414368"></a><a name="p14590102414368"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p35901242366"><a name="p35901242366"></a><a name="p35901242366"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row529461533914"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p429531513391"><a name="p429531513391"></a><a name="p429531513391"></a>code</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p172959154394"><a name="p172959154394"></a><a name="p172959154394"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p329511583917"><a name="p329511583917"></a><a name="p329511583917"></a>action的执行结果。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><a name="ul4679192464120"></a><a name="ul4679192464120"></a><ul id="ul4679192464120"><li>0：修复成功。</li><li>405：retry修复失败，支持做recover、dump、exit修复策略。</li><li>406：修复失败，支持做dump或exit修复策略。</li><li>499：修复失败，仅支持exit策略。</li></ul>
</td>
</tr>
<tr id="row5949718163914"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p77183173910"><a name="p77183173910"></a><a name="p77183173910"></a>msg</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p159509185392"><a name="p159509185392"></a><a name="p159509185392"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p39501218193914"><a name="p39501218193914"></a><a name="p39501218193914"></a>修复成功或失败的消息。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p7950171818398"><a name="p7950171818398"></a><a name="p7950171818398"></a>str</p>
</td>
</tr>
<tr id="row1591162416368"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p359152412367"><a name="p359152412367"></a><a name="p359152412367"></a>error_rank_dict</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1859118247367"><a name="p1859118247367"></a><a name="p1859118247367"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p85912245362"><a name="p85912245362"></a><a name="p85912245362"></a>发生故障的卡信息。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p85915243360"><a name="p85915243360"></a><a name="p85915243360"></a>&lt;int key, int errorType&gt;字典：</p>
<a name="ul1859162411368"></a><a name="ul1859162411368"></a><ul id="ul1859162411368"><li>key为故障卡的rank号。</li><li>errorType为故障类型：<a name="ul10591192413614"></a><a name="ul10591192413614"></a><ul id="ul10591192413614"><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul>
</li></ul>
</td>
</tr>
<tr id="row1025534083913"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p8256194019396"><a name="p8256194019396"></a><a name="p8256194019396"></a>curr_strategy</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p112561440163918"><a name="p112561440163918"></a><a name="p112561440163918"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p1625644013393"><a name="p1625644013393"></a><a name="p1625644013393"></a>本次修复策略。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p209125518405"><a name="p209125518405"></a><a name="p209125518405"></a>str，支持的修复策略取值范围为<a href="#table116012273617">表3</a>中的strategy_list。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

-   0：调用成功
-   1：调用失败

