# （可选）配置故障检测级别<a name="ZH-CN_TOPIC_0000002479386556"></a>

## 配置说明<a name="ZH-CN_TOPIC_0000002479386448"></a>

断点续训针对节点故障中**节点硬件故障**、**芯片故障、灵衢总线设备故障**和**公共故障**的不同故障码，提供了默认的故障级别和对应级别的故障处理策略；**芯片故障**还提供了默认的故障频率和时长，以及对应的故障处理策略。

若用户需要修改故障处理策略可参见本章节。若无特殊需求，请勿随意修改。

**支持配置的故障级别说明<a name="section257513292065"></a>**

不同类型的故障支持配置的故障级别如下表所示。

**表 1**  支持配置的故障级别

<a name="table4710459145316"></a>
<table><thead align="left"><tr id="row37104590534"><th class="cellrowborder" valign="top" id="mcps1.2.5.1.1"><p id="p7710135925316"><a name="p7710135925316"></a><a name="p7710135925316"></a>故障名称</p>
</th>
<th class="cellrowborder" colspan="3" valign="top" id="mcps1.2.5.1.2"><p id="p11175192213564"><a name="p11175192213564"></a><a name="p11175192213564"></a>支持配置的故障级别</p>
</th>
</tr>
</thead>
<tbody><tr id="row271045905320"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p271015916536"><a name="p271015916536"></a><a name="p271015916536"></a>节点故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p66711187562"><a name="p66711187562"></a><a name="p66711187562"></a>NotHandleFault、PreSeparateFault、SeparateFault</p>
</td>
</tr>
<tr id="row3710165935311"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17710125955315"><a name="p17710125955315"></a><a name="p17710125955315"></a>芯片故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p21371428713"><a name="p21371428713"></a><a name="p21371428713"></a>NotHandleFault、RestartRequest、RestartBusiness、FreeRestartNPU、RestartNPU、SeparateNPU、PreSeparateNPU、SubHealthFault</p>
</td>
</tr>
<tr id="row5710125913537"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p10710959185319"><a name="p10710959185319"></a><a name="p10710959185319"></a>灵衢总线设备故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p6631112135616"><a name="p6631112135616"></a><a name="p6631112135616"></a>NotHandleFault、SubHealthFault、ResetFault、SeparateFault<span id="ph51441721217"><a name="ph51441721217"></a><a name="ph51441721217"></a>、</span><span id="ph375517710129"><a name="ph375517710129"></a><a name="ph375517710129"></a>RestartRequestFault</span></p>
</td>
</tr>
<tr id="row416145918513"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p116115913517"><a name="p116115913517"></a><a name="p116115913517"></a>公共故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p147536536717"><a name="p147536536717"></a><a name="p147536536717"></a>NotHandleFault、SeparateNPU、SubHealthFault<span id="ph632635517598"><a name="ph632635517598"></a><a name="ph632635517598"></a>、PreSeparateNPU</span></p>
</td>
</tr>
</tbody>
</table>

在以上表格中，每种故障级别的处理策略说明如下。

**表 2**  故障级别及处理说明

<a name="table103716651410"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row461812518228"><th class="cellrowborder" valign="top" width="19.06%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"></a>故障处理策略</p>
</th>
<th class="cellrowborder" valign="top" width="35.74%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="23.39%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"></a>重调度处理</p>
</th>
<th class="cellrowborder" valign="top" width="21.81%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"></a>优雅容错处理</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row961811511228"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"></a>对业务无影响的故障，无需处理。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"></a>暂不处理</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"></a>暂不处理</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row116184515226"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"></a>RestartRequest</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"></a>影响业务执行，需要重新执行业务请求。</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"></a>隔离芯片，进行任务重调度。</p>
<div class="note" id="note11901123612819"><a name="note11901123612819"></a><a name="note11901123612819"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1069261722310"><a name="p1069261722310"></a><a name="p1069261722310"></a>若推理任务订阅<span id="ph4356222144812"><a name="ph4356222144812"></a><a name="ph4356222144812"></a>了</span>故障信息，任务使用的推理卡上发生RestartRequest故障且故障持续时间未超过60秒，则不执行任务重调度；若故障持续时间超过60秒仍未恢复，则隔离芯片，进行任务重调度。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"></a>推理场景重新执行推理请求，训练场景重新执行训练业务。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row14618105116225"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"></a>RestartBusiness</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"></a>影响业务执行，需要重新执行业务。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"></a>重新执行业务</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row561825132214"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"></a>FreeRestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"></a>影响业务执行，待芯片空闲时需复位芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"></a>等待芯片空闲后复位芯片。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row14618125152210"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"></a>RestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"></a>影响业务执行，需立即复位芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"></a>立即停止训练业务，复位芯片后重新执行业务。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row1061895115227"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"></a>无法恢复，需要隔离芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"></a>隔离芯片，进行任务重调度。</p>
</td>
</tr>
<tr id="row870814247412"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p5708202454117"><a name="p5708202454117"></a><a name="p5708202454117"></a>SeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p12708162474117"><a name="p12708162474117"></a><a name="p12708162474117"></a>任务一定会受到影响。</p>
<div class="note" id="note1521013164613"><a name="note1521013164613"></a><a name="note1521013164613"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p92101114465"><a name="p92101114465"></a><a name="p92101114465"></a>灵衢总线设备故障级别为SeparateFault时，表示业务运行失败，需更换器件或板卡。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p0708624204112"><a name="p0708624204112"></a><a name="p0708624204112"></a>任务重调度</p>
<div class="note" id="note44451347164716"><a name="note44451347164716"></a><a name="note44451347164716"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p64453471479"><a name="p64453471479"></a><a name="p64453471479"></a>灵衢总线设备故障下，本故障级别代表的故障处理策略为停止当前训练任务，隔离节点，进行任务重调度。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p137081824174117"><a name="p137081824174117"></a><a name="p137081824174117"></a>-</p>
</td>
</tr>
<tr id="row5706333131216"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p177061833201220"><a name="p177061833201220"></a><a name="p177061833201220"></a><span id="ph141513510124"><a name="ph141513510124"></a><a name="ph141513510124"></a>RestartRequestFault</span></p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p070623351220"><a name="p070623351220"></a><a name="p070623351220"></a><span id="ph18501459184"><a name="ph18501459184"></a><a name="ph18501459184"></a>业务运行失败，需要重新执行业务请求。</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p770653313124"><a name="p770653313124"></a><a name="p770653313124"></a><span id="ph38912127169"><a name="ph38912127169"></a><a name="ph38912127169"></a>停止当前训练任务，隔离节点，进行任务重调度。</span></p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p6706113331213"><a name="p6706113331213"></a><a name="p6706113331213"></a>推理场景重新执行推理请求，训练场景重新执行训练业务。</p>
</td>
</tr>
<tr id="row3938182254418"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p39381822174417"><a name="p39381822174417"></a><a name="p39381822174417"></a>ResetFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p1193862274418"><a name="p1193862274418"></a><a name="p1193862274418"></a>业务运行失败</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p184323519501"><a name="p184323519501"></a><a name="p184323519501"></a>停止当前训练任务，隔离节点，进行任务重调度。</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p18938822204411"><a name="p18938822204411"></a><a name="p18938822204411"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row102215292529"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"></a>PreSeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"></a>暂不影响业务，后续不再调度任务到该芯片。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"></a>预隔离芯片</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"></a>预隔离芯片</p>
</td>
</tr>
<tr id="row84541721401"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p174559214016"><a name="p174559214016"></a><a name="p174559214016"></a>PreSeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p145562114011"><a name="p145562114011"></a><a name="p145562114011"></a>可能导致任务受到影响。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p54556214409"><a name="p54556214409"></a><a name="p54556214409"></a>该节点上有任务则不处理，后续调度时不调度任务到该节点。</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p1245572144015"><a name="p1245572144015"></a><a name="p1245572144015"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row0352224175218"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"></a>根据任务YAML中配置的subHealthyStrategy参数取值进行处理，详细请参见<a href="./06_configuring_the_job_yaml_file.md#yaml参数说明">YAML参数说明</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"></a>当芯片出现亚健康故障时，需根据<a href="./06_configuring_the_job_yaml_file.md#任务yaml配置示例">配置YAML</a>策略进行处理。</p>
<div class="note" id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"></a>如果后续芯片出现其他级别故障，此时SubHealthFault处理策略不影响其他级别的故障处理。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"></a>根据策略进行处理。</p>
</td>
</tr>
</tbody>
</table>

## 节点硬件故障<a name="ZH-CN_TOPIC_0000002479226584"></a>

### 配置文件说明<a name="ZH-CN_TOPIC_0000002479226562"></a>

断点续训针对节点故障中**节点硬件故障**的不同级别进行分级处理。NodeD组件会获取到当前故障的故障码，根据NodeDConfiguration.json中故障码配置的故障级别，对故障进行相应处理。节点硬件故障支持的故障级别和处理方式说明如下。

NodeD组件的配置文件NodeDConfiguration.json为系统配置文件，若用户无特殊需求，请勿随意修改。若用户需要修改故障码的故障级别，可以通过由NodeDConfiguration.json创建的mindx-dl-node-fault-config文件实现，操作指导请参见[（可选）配置节点硬件故障级别](#可选配置节点硬件故障级别)。故障级别说明及节点状态说明请参见[自定义节点故障](../../api/noded.md#自定义节点故障)。

### （可选）配置节点硬件故障级别<a name="ZH-CN_TOPIC_0000002511346507"></a>

在制作NodeD镜像时，会将故障级别配置文件NodeDConfiguration.json内置在镜像中，启动NodeD时会读取这个文件的默认配置，作为当前故障处理依据。

如果用户想要自定义故障级别，可以在集群中创建ConfigMap文件（mindx-dl-node-fault-config）。

- 如果NodeD启动时，集群中已经存在该mindx-dl-node-fault-config，NodeD会优先按照已存在的mindx-dl-node-fault-config中配置的内容，作为当前故障处理依据。
- 如果重新安装NodeD后，集群中已经存在mindx-dl-node-fault-config，NodeD的默认NodeDConfiguration.json将不会生效，使用集群中已经存在mindx-dl-node-fault-config。若想要使用NodeDConfiguration.json的默认配置，可以删除mindx-dl-node-fault-config，使NodeD读取默认的NodeDConfiguration.json文件。
- 如果mindx-dl-node-fault-config内容存在格式错误等问题，NodeD会默认读取镜像中内置的NodeDConfiguration.json文件的内容，作为当前故障处理依据。

**操作步骤<a name="section25164134219"></a>**

以故障码0100001D为例，将当前故障的处理策略NotHandleFault（无需处理）修改为PreSeparateFault（该节点上有任务则不处理，后续不调度任务到该节点）的操作示例如下。

1. 登录环境，进入NodeD解压目录。
2. 执行以下命令，创建动态配置故障级别所需ConfigMap文件（mindx-dl-node-fault-config）。

    ```shell
    kubectl create cm mindx-dl-node-fault-config -n mindx-dl  --from-file=./NodeDConfiguration.json
    ```

    回显示例如下：

    ```ColdFusion
    configmap/mindx-dl-node-fault-config created
    ```

    **表 1**  参数说明

    <a name="table1925220306444"></a>
    <table><thead align="left"><tr id="row172531430134411"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p16253163094420"><a name="p16253163094420"></a><a name="p16253163094420"></a>参数名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p152534301443"><a name="p152534301443"></a><a name="p152534301443"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1325318306446"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p15214952162210"><a name="p15214952162210"></a><a name="p15214952162210"></a>mindx-dl-node-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p621417523229"><a name="p621417523229"></a><a name="p621417523229"></a>创建的<span id="ph188631730142314"><a name="ph188631730142314"></a><a name="ph188631730142314"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="row925343011442"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p82141952122212"><a name="p82141952122212"></a><a name="p82141952122212"></a>mindx-dl</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p0214952142217"><a name="p0214952142217"></a><a name="p0214952142217"></a>命名空间名称，不能修改该命名空间。</p>
    </td>
    </tr>
    <tr id="row1253183012444"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p182141521222"><a name="p182141521222"></a><a name="p182141521222"></a>NodeDConfiguration.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p22148525226"><a name="p22148525226"></a><a name="p22148525226"></a>用于配置故障码以及对应的故障级别，必须与NodeDConfiguration.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

3. 执行以下命令，编辑mindx-dl-node-fault-config文件。

    ```shell
    kubectl edit cm -n mindx-dl mindx-dl-node-fault-config
    ```

4. 在mindx-dl-node-fault-config文件中，找到故障码0100001D。

    ```json
     "FaultTypeCode": {
            "NotHandleFaultCodes":[
              "0100001D","03000009","03000013","0300000D","03000011"
            ],
    ...
      ],
    ...
    ```

    >[!NOTE] 
    >自定义故障级别时，若不小心导致出现以下问题，则本次修改无效，NodeD将会使用上一次保存的配置进行处理。
    >- 文件格式异常或故障码取值错误，故障码只能为8位的包含数字和字母的字符串。
    >- 同一故障码同时配置在多个故障级别中。

5. 将故障码0100001D在**NotHandleFaultCodes**中删除，并添加到**PreSeparateFaultCodes**中。

    ```json
     "FaultTypeCode": {
            "NotHandleFaultCodes":[
             "03000009","03000013","0300000D","03000011"
            ],
            "PreSeparateFaultCodes":[
              "28000037","00000011", "0100001D"
    ...
            ],
    ...
    ```

6. 修改完成后，按“Esc”键，输入:wq!保存并退出。
7. 等mindx-dl-node-fault-config文件更新后，查看操作是否成功。
    1. 执行以下命令，查询NodeD组件日志名称。

        ```shell
        kubectl get pods -A | grep noded
        ```

        回显示例如下：

        ```ColdFusion
        mindx-dl      noded-c5f52   1/1     Running   0               2m16s
        ```

    2. 通过查询到的组件日志名称，查询NodeD的组件日志信息。

        ```shell
        kubectl logs noded-c5f52 -n mindx-dl -f
        ```

        若日志出现“update fault config success”，表示动态配置故障码操作成功。

## 芯片故障<a name="ZH-CN_TOPIC_0000002479226466"></a>

### 概述<a name="ZH-CN_TOPIC_0000002511346521_0101"></a>

Ascend Device Plugin和ClusterD均提供了按照故障频率进行人工隔离芯片的能力，两者功能差异如下：

- Ascend Device Plugin基于节点维度进行故障判定，根据实际发生的故障进行频率计数。
- ClusterD基于任务维度进行故障判定。
    - 若一个任务下30s内多张卡同时出现同一个故障，则认为不是硬件故障导致，不会进行故障频率计数。该判断规则适用于大多数场景。对于Pod被删除但是有残留进程等场景，故障频率计数可能存在偏差。
    - 只有新故障才能触发人工隔离芯片的故障频率是否达到上限的判断。如果将配置的阈值调整为当前的计数，无法立刻触发隔离，需要等下一次故障发生时触发判断逻辑。
    - ClusterD重启后，频率计数信息会丢失，人工隔离芯片的故障频率会从零开始计数。
    - 如果解除隔离后，任务调度不符合预期，可查看节点上是否打了标签huawei.com/scheduler.chip1softsharedev.enable=false。如果打了该标签，需要删除。

Ascend Device Plugin和ClusterD的人工隔离芯片功能，理论上涉及的故障码不需要重复。若不想使用Ascend Device Plugin的隔离功能，请参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)章节，将faultCustomization.json文件中的人工隔离芯片相关的配置删除；若不想使用ClusterD的隔离功能，请参见[（可选）配置芯片故障频率](#可选配置芯片故障频率)章节，将人工隔离芯片功能开关关闭。

若Ascend Device Plugin和ClusterD对同一张芯片都进行了人工隔离，需要各自解除隔离。Ascend Device Plugin解除隔离的方法请参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)中"手动恢复强制隔离的芯片"步骤；ClusterD解除隔离的方法请参见[（可选）配置芯片故障频率](#可选配置芯片故障频率)中"手动恢复人工隔离的芯片"步骤。

### Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511346521_02"></a>

#### 配置文件说明<a name="ZH-CN_TOPIC_0000002511346521"></a>

断点续训针对**芯片故障**，支持按故障级别、故障频率和故障时长的配置进行处理。

- 针对芯片故障的**不同级别**进行分级处理时，Ascend Device Plugin组件会获取到当前故障的故障码，根据**faultCode.json**中故障码配置的故障级别，对故障进行相应处理。
- 针对芯片故障的**故障频率及时长**进行处理时，Ascend Device Plugin组件会获取到当前故障的故障码，根据**faultCustomization.json**中故障配置的故障频率和时长，对故障进行相应处理。

faultCode.json、faultCustomization.json为系统配置文件，若用户无特殊需求，请勿随意修改。若Ascend Device Plugin默认的频率故障配置中有由软件原因可以触发的故障，用户可自行将该故障码删除。（软件原因会导致一个任务下短时间内反复大量出现某个故障，导致Ascend Device Plugin侧感知到该故障达到了故障频率，将大量设备置为人工隔离状态。）

若用户需要修改故障码对应的故障级别，可以通过由faultCode.json和faultCustomization.json创建的**mindx-dl-fault-config**文件实现。

>[!NOTE] 
>
>- 每个故障对应的故障码请参见[芯片故障码参考文档](../../appendix.md#芯片故障码参考文档)章节。
>- 芯片故障支持配置的故障级别参见[故障级别](#zh-cn_topic_0000002171521445_section5245155017242)。
>- 芯片故障支持配置的故障频率和时长参见[故障频率及时长](#zh-cn_topic_0000002171521445_section115842029104220)。

**faultCode.json中的故障级别<a name="zh-cn_topic_0000002171521445_section5245155017242"></a>**

断点续训针对芯片故障的不同级别进行分级处理。若用户需要修改故障码的故障级别，操作指导请参见[（可选）配置芯片故障级别](#可选配置芯片故障级别)。

Ascend Device Plugin从驱动获取到芯片故障码后，将根据故障码对设备及业务的影响将故障划分几个级别，详细说明请参见[表1](../../api/ascend_device_plugin.md#自定义芯片故障)。

>[!NOTE] 
>
>- 复位芯片前需要停止训练进程，否则复位将失败。
>- 若Ascend Device Plugin通过订阅的方式收到了无法识别的故障码（未保存在faultCode.json中），默认按照订阅接口给的处理意见进行故障处理。若订阅接口收到的故障等级为“提示”或“次要”，则按照NotHandleFault级别处理；若故障等级为其他等级，则按照SeparateNPU级别处理。

**故障频率及时长<a name="zh-cn_topic_0000002171521445_section115842029104220"></a>**

断点续训针对芯片故障的故障频率及时长进行处理。某些硬件类故障可能在一次训练任务中反复出现，导致训练任务中断反复进行重调度。集群调度组件针对这些故障对应的故障码，提供了提升故障级别的初始化配置文件faultCustomization.json。

- faultCustomization.json文件提供的初始化配置和故障类型关系如下[初始化配置和故障类型](#zh-cn_topic_0000002171521445_section13684172919539)。
- faultCustomization.json文件的默认配置（默认值）请参见[表2](../../api/ascend_device_plugin.md#自定义芯片故障)。
- 若用户需要修改故障频率及时长配置，操作指导请参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)。

**初始化配置和故障类型<a name="zh-cn_topic_0000002171521445_section13684172919539"></a>**

当前faultCustomization.json文件中仅提供对可识别的硬件类故障进行提升故障级别的初始化配置。

24小时内发生3次以下故障，则将芯片故障级别提升至需要人工干预的故障级别ManuallySeparateNPU，详细说明请参见[faultCustomization.json参数说明](#zh-cn_topic_0000002171521445_section33036167576)。

下面将以故障名称HBMC Ca Parity错误，对应故障码80E18005为例，将当前的故障级别提升至ManuallySeparateNPU（需要人工干预的故障级别），示例如下。

```json
  "FaultFrequency": [
    {
      "EventId": [
        "80C98000","80B78000","80B58000","80A18008","80A38008","80A58008","80B98000","80B98008","80BB8000",
        "80BB8008","80BD8000","80BD8008","80C78008","80C98008","80CB8008","80CD8008","80CF8008","80D98008",
        "80DF8008","80DE1801","80E01801","80E18008","80E38008","80E39200","80E3A202","80E3A203","80E78000",
        "80E78008","80F18000","80F18008","80F38008","80F78008","81318008","81338008","813B8008","81478008",
        "81578008","815F8008","81938008","81958008","81978008"
      ],
      "TimeWindow": 86400,
      "Times": 2,
      "FaultHandling": "ManuallySeparateNPU"
    },
    {
      "EventId": ["80E18005"],
      "TimeWindow": 86400,
      "Times": 3,
      "FaultHandling": "ManuallySeparateNPU"
    }
  ],
```

>[!NOTE]
>
>- 故障的处理策略为ManuallySeparateNPU时，可以参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)中“手动恢复强制隔离的芯片”步骤进行处理。
>- 除可以识别的硬件故障外，faultCustomization.json文件中还包含以下几类故障。
>     - 无需处理的故障：该类故障出现不影响训练任务及设备，不提供提升故障级别的初始化配置。
>     - 无法识别出是硬件还是软件类故障：该类故障无法准确识别是硬件还是软件故障，且会影响训练任务。该类故障不提供提升故障级别的初始化配置，建议用户根据实际情况手动配置任务支持的断点续训最大次数和达到最大次数后故障的处理策略，可以参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)进行配置。
>     - 软件配置类故障：该类故障为软件配置类问题，正常情况下不会出现。该类故障不提供提升故障级别的初始化配置，建议用户检查软件版本是否配套。

**faultCustomization.json参数说明<a name="zh-cn_topic_0000002171521445_section33036167576"></a>**

用户不手动修改faultCustomization.json文件时，Ascend Device Plugin按照faultCustomization.json的默认配置（默认值）进行故障处理。faultCustomization.json文件参数说明请参见[表2](../../api/ascend_device_plugin.md#自定义芯片故障)。

#### （可选）配置芯片故障级别<a name="ZH-CN_TOPIC_0000002479226532"></a>

在制作Ascend Device Plugin镜像时，会将faultCode.json和faultCustomization.json配置文件内置在镜像中，启动Ascend Device Plugin时会读取这两个文件的默认配置，作为当前故障处理依据。faultCode.json和faultCustomization.json的说明请参见[配置文件说明](#ZH-CN_TOPIC_0000002511346521)。

如果用户想要自定义故障级别或者优雅容错相关配置，可以在集群中创建ConfigMap文件（mindx-dl-fault-config）。

- 如果Ascend Device Plugin启动时，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin会优先按照已存在的mindx-dl-fault-config中配置的内容，作为当前故障处理依据。
- 如果重新安装Ascend Device Plugin后，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin的默认faultCode.json将不会生效，使用集群中已经存在的mindx-dl-fault-config。
- 若想要使用faultCode.json或faultCustomization.json的默认配置，可以删除mindx-dl-fault-config，使Ascend Device Plugin读取默认faultCode.json、SwitchFaultCode.json或faultCustomization.json文件。
- 如果ConfigMap文件内容存在格式错误等问题，Ascend Device Plugin会默认读取镜像中内置的ConfigMap文件的内容，作为当前故障处理依据。

**使用faultCode.json配置故障级别<a name="zh-cn_topic_0000001951258609_section112139052513"></a>**

以故障名称dmp\_daemon节点状态检测异常，对应故障码80E21007为例。将当前故障的处理策略NotHandleFaultCodes（无需处理）修改为RestartNPUCodes（隔离芯片，进行任务重调度）的操作示例如下。

1. 登录环境，进入Ascend Device Plugin解压目录。
2. 执行以下命令，创建动态配置故障码所需ConfigMap文件（mindx-dl-fault-config）。

    ```shell
    kubectl create cm mindx-dl-fault-config -n kube-system --from-literal="PollInterval=300" --from-file=./faultCode.json
    ```

    回显示例如下。

    ```ColdFusion
    configmap/mindx-dl-fault-config created
    ```

    **表 1**  参数说明

    <a name="zh-cn_topic_0000001951258609_table16314861918"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001951258609_row763648161914"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001951258609_p16631548171910"><a name="zh-cn_topic_0000001951258609_p16631548171910"></a><a name="zh-cn_topic_0000001951258609_p16631548171910"></a>参数名</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001951258609_p1663144816197"><a name="zh-cn_topic_0000001951258609_p1663144816197"></a><a name="zh-cn_topic_0000001951258609_p1663144816197"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001951258609_p775918210209"><a name="zh-cn_topic_0000001951258609_p775918210209"></a><a name="zh-cn_topic_0000001951258609_p775918210209"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001951258609_row36354871915"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1863164816197"><a name="zh-cn_topic_0000001951258609_p1863164816197"></a><a name="zh-cn_topic_0000001951258609_p1863164816197"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p1063194861910"><a name="zh-cn_topic_0000001951258609_p1063194861910"></a><a name="zh-cn_topic_0000001951258609_p1063194861910"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p157595292015"><a name="zh-cn_topic_0000001951258609_p157595292015"></a><a name="zh-cn_topic_0000001951258609_p157595292015"></a>动态配置故障码所需的<span id="zh-cn_topic_0000001951258609_ph126311642183015"><a name="zh-cn_topic_0000001951258609_ph126311642183015"></a><a name="zh-cn_topic_0000001951258609_ph126311642183015"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row763184812192"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1963194819195"><a name="zh-cn_topic_0000001951258609_p1963194819195"></a><a name="zh-cn_topic_0000001951258609_p1963194819195"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p76316488192"><a name="zh-cn_topic_0000001951258609_p76316488192"></a><a name="zh-cn_topic_0000001951258609_p76316488192"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p276092142019"><a name="zh-cn_topic_0000001951258609_p276092142019"></a><a name="zh-cn_topic_0000001951258609_p276092142019"></a>mindx-dl-fault-config所在命名空间，不能修改该命名空间名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row86314881910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p964144891914"><a name="zh-cn_topic_0000001951258609_p964144891914"></a><a name="zh-cn_topic_0000001951258609_p964144891914"></a>PollInterval</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p1164748191916"><a name="zh-cn_topic_0000001951258609_p1164748191916"></a><a name="zh-cn_topic_0000001951258609_p1164748191916"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p876012211206"><a name="zh-cn_topic_0000001951258609_p876012211206"></a><a name="zh-cn_topic_0000001951258609_p876012211206"></a>不指定该参数则默认取值为300s。用于指定查询mindx-dl-fault-config文件是否更新的周期时间，单位为秒，取值范围为30~3600。PollInterval的修改将在下一个周期生效。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row176474851911"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1964748141915"><a name="zh-cn_topic_0000001951258609_p1964748141915"></a><a name="zh-cn_topic_0000001951258609_p1964748141915"></a>faultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p10641648191915"><a name="zh-cn_topic_0000001951258609_p10641648191915"></a><a name="zh-cn_topic_0000001951258609_p10641648191915"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p147602211206"><a name="zh-cn_topic_0000001951258609_p147602211206"></a><a name="zh-cn_topic_0000001951258609_p147602211206"></a>用于保存故障码，必须与faultCode.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

3. 执行以下命令，编辑mindx-dl-fault-config文件。

    ```shell
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

4. 在mindx-dl-fault-config文件中，找到故障码80E21007。

    ```json
    "NotHandleFaultCodes":[
        
    "80E21007","80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    ...
      ],
    ...
    ```

    >[!NOTE] 
    >同一故障码配置在多个故障级别中，会显示设置成功，但默认按照高等级故障处理。

5. 将故障码80E21007从NotHandleFaultCodes中删除，并添加到RestartNPUCodes中。

    ```json
    "NotHandleFaultCodes":[ 
         "80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    ...
      ],
    ...
    "RestartNPUCodes":[
       "8C204E00","A8028802","A4302003","A4302004","A4302005","A4302006","A4302009","A430200A","80CF8009","80CF8008","80E21007",... 
    ...
       ],
    ```

6. 修改完成后，按“Esc”键，输入:wq!保存并退出。
7. 等mindx-dl-fault-config文件更新生效（PollInterval取值，不指定则为300s）后，查看操作是否成功。
    1. 执行以下命令，查询Ascend Device Plugin组件日志名称。

        ```shell
        kubectl get pods -A | grep ascend-device-plugin
        ```

        回显示例如下：

        ```ColdFusion
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2. 通过查询到的组件日志名称，查询Ascend Device Plugin的组件日志信息。

        ```shell
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        若日志出现“load fault code from configmap success”，表示手动配置故障码操作成功。

#### （可选）配置芯片故障频率及时长<a name="ZH-CN_TOPIC_0000002511426473"></a>

在制作Ascend Device Plugin镜像时，会将faultCode.json和faultCustomization.json配置文件内置在镜像中，启动Ascend Device Plugin时会读取这两个文件的默认配置，作为当前故障处理依据。faultCode.json和faultCustomization.json的说明请参见[配置文件说明](#ZH-CN_TOPIC_0000002511346521)。

如果用户想要自定义芯片故障频率及时长，可以在集群中创建ConfigMap文件（mindx-dl-fault-config）。

- 如果Ascend Device Plugin启动时，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin会优先按照已存在的mindx-dl-fault-config中配置的内容，作为当前故障处理依据。
- 如果重新安装Ascend Device Plugin后，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin的默认faultCustomization.json将不会生效，使用集群中已经存在的mindx-dl-fault-config。若想要使用faultCustomization.json的默认配置，可以删除mindx-dl-fault-config，使Ascend Device Plugin读取默认faultCustomization.json文件。
- 如果ConfigMap文件内容存在格式错误等问题，Ascend Device Plugin会默认读取镜像中内置的ConfigMap文件的内容，作为当前故障处理依据。

>[!CAUTION]
>修改故障频率为高危操作，如果修改不当，会导致芯片被误隔离。例如，由于任务发生错误导致的软件故障，会短时间内反复大量出现，使Ascend Device Plugin侧感知到该故障达到了故障频率，将大量芯片置为人工隔离状态，导致大量节点无法调度。

**操作步骤<a name="section141902103110"></a>**

以故障码80CB8002为例，如果某张芯片反复发生80CB8002故障，导致训练业务反复重调度，可以手动配置24小时内任务支持的断点续训最大次数为2，达到最大次数后故障的处理策略为ManuallySeparateNPU。

1. 登录环境，进入Ascend Device Plugin解压目录。
2. 执行以下命令，查询是否已经基于faultCode.json文件创建了mindx-dl-fault-config。

    ```shell
    kubectl describe cm -n kube-system mindx-dl-fault-config
    ```

    - 如果mindx-dl-fault-config已经存在，且存在faultCustomization.json的相关字段，执行[步骤4](#zh-cn_topic_0000002136360238_li38432520129)编辑该文件。
    - 如果mindx-dl-fault-config已经存在，但是不存在faultCustomization.json的相关字段，需要先保存mindx-dl-fault-config内容，再删除mindx-dl-fault-config文件后，执行[步骤3](#zh-cn_topic_0000002136360238_li1946014413123)创建该文件。
    - 如果不存在mindx-dl-fault-config，执行[步骤3](#zh-cn_topic_0000002136360238_li1946014413123)创建该文件。

3. <a name="zh-cn_topic_0000002136360238_li1946014413123"></a>执行以下命令，创建配置芯片故障频率及时长所需ConfigMap文件（mindx-dl-fault-config）。

    ```shell
    kubectl create cm mindx-dl-fault-config -n kube-system --from-literal="PollInterval=300" --from-file=./faultCode.json --from-file=./faultCustomization.json
    ```

    回显示例如下。

    ```ColdFusion
    configmap/mindx-dl-fault-config created
    ```

    **表 1**  参数说明

    <a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_table16314861918"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row763648161914"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"></a>参数名</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row36354871915"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"></a>动态配置故障码所需的<span id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row763184812192"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"></a>mindx-dl-fault-config所在命名空间，不能修改该命名空间名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row86314881910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"></a>PollInterval</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"></a>不指定该参数则默认取值为300s。用于指定查询mindx-dl-fault-config文件是否更新的周期时间，单位为秒，取值范围为30~3600。PollInterval的修改将在下一个周期生效。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row176474851911"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"></a>faultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"></a>用于保存故障码，必须与faultCode.json文件名称保持一致。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_row9289716194614"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"></a>faultCustomization.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"></a>用于自定义优雅容错时间、故障频率、故障持续时间（仅支持参数面网络故障）等配置，不指定该参数则没有故障频率配置，其余配置使用默认值进行处理。必须与faultCustomization.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

4. <a name="zh-cn_topic_0000002136360238_li38432520129"></a>执行以下命令，编辑mindx-dl-fault-config文件。

    ```shell
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

    根据实际情况，修改芯片的故障频率和时长。

    ```json
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
    PollInterval: "300"
    # 修改芯片故障的故障级别
    faultCode.json: |
    {
    "NotHandleFaultCodes":[
    ...
    }
    # 修改芯片故障的故障频率和时长
    faultCustomization.json: |
    {
     "GraceTolerance": {
     "WaitProcessReadCMTime": 30,
     "WaitDeviceResetTime": 150,
     "WaitFaultSelfHealingTime": 15
    },
    "FaultFrequency": [
     {
      "EventId": [
        "80C98000","80B78000","80B58000","80A18008","80A38008","80A58008","80B98000","80B98008","80BB8000",
        "80BB8008","80BD8000","80BD8008","80C78008","80C98008","80CB8008","80CD8008","80CF8008","80D98008",
        "80DF8008","80DE1801","80E01801","80E18008","80E38008","80E39200","80E3A202","80E3A203","80E78000",
        "80E78008","80F18000","80F18008","80F38008","80F78008","81318008","81338008","813B8008","81478008",
        "81578008","815F8008","81938008","81958008","81978008"
      ],
      "TimeWindow": 86400,
      "Times": 2,
      "FaultHandling": "ManuallySeparateNPU"
     },
     {
      "EventId": ["80E18005"],
      "TimeWindow": 86400,
      "Times": 3,
      "FaultHandling": "ManuallySeparateNPU"
     },
     {
      "EventId": ["81078603"],
      "TimeWindow": 86400,
      "Times": 5,
      "FaultHandling": "ManuallySeparateNPU",
      "ReleaseTimeWindow": 172800
     }
    ],
    "FaultDuration": [
     {
      "EventId": ["81078603"],
      "FaultTimeout": 20,
      "RecoverTimeout": 60,
      "FaultHandling": "PreSeparateNPU"
     },
     {
      "EventId": ["81B18603"],
      "FaultTimeout": 5,
      "RecoverTimeout": 60,
      "FaultHandling": "PreSeparateNPU"
     }
    ]
   }
    kind: ConfigMap
    metadata:
    creationTimestamp: "2024-06-20T10:12:07Z"
    name: mindx-dl-fault-config
    namespace: kube-system
    resourceVersion: "52893696"
    selfLink: /api/v1/namespaces/kube-system/configmaps/mindx-dl-fault-config
    uid: bba9e17f-41dd-43b3-848e-3d29cb8c595a
    ```

5. 在mindx-dl-fault-config文件中，在FaultFrequency字段下新增以下代码，设置80CB8002故障在24小时内任务支持的断点续训最大次数为2，达到最大次数后故障的处理策略为ManuallySeparateNPU。

    ```json
    {
      "EventId": ["80CB8002"],
      "TimeWindow": 86400,
      "Times": 2,      
      "FaultHandling": "ManuallySeparateNPU"
    }
    ```

6. 修改完成后，按“Esc”键，输入:wq!保存并退出。
7. 等mindx-dl-fault-config文件更新生效（PollInterval取值，不指定则为300s）后，查看操作是否成功。
    1. 执行以下命令，查询Ascend Device Plugin组件日志名称。

        ```shell
        kubectl get pods -A | grep ascend-device-plugin
        ```

        回显示例如下：

        ```ColdFusion
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2. 通过查询到的组件日志名称，查询Ascend Device Plugin的组件日志信息。

        ```shell
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        >[!NOTE] 
        >- 若日志出现“load fault customization from configmap complete”，表示手动配置故障频率操作成功。
        >- 若日志出现“modify  _xxx_  success”，表示ConfigMap中faultCustomization.json里的<i>xxx</i>参数设置成功。
        >- 若日志出现“insert fault frequency success”，表示记录了一次频率故障发生时间，在频率窗口内，该卡的该故障记录次数达到频率故障触发次数以后，就会上报频率故障对应的故障级别。

8. （可选）手动恢复强制隔离的芯片。故障的处理策略为ManuallySeparateNPU时，故障恢复后该芯片也处于隔离状态，在未达到释放条件时若需要手动恢复强制隔离的芯片。
    1. 执行以下命令，查找该节点的Ascend Device Plugin上报的device-info-cm。

        ```shell
        kubectl get cm -n kube-system | grep deviceinfo | grep {nodeName}
        ```

    2. 执行以下命令，编辑该device-info-cm。

        ```shell
        kubectl edit cm -n kube-system {configMapName}
        ```

    3. 将data下面的ManuallySeparateNPU后面已恢复健康的芯片名称删除。

        ```Yaml
        apiVersion: v1
        kind: ConfigMap
        data:
          DeviceInfoCfg: '{"DeviceInfo":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-Fault":"[]","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Unhealthy":""},"UpdateTime":1718702470},"CheckCode":"4f00cf1d220da26a8fdbeb5ba163a751d4b264c48b81d22149257e272ae3b413"}'
          ManuallySeparateNPU: Ascend910-0  
        ```

        >[!NOTE] 
        >删除ManuallySeparateNPU字段后所有芯片名称，并将取值设置为空“”。

    4. 修改完成后，按“Esc”键，输入:wq!保存并退出。
    5. 等待1个上报周期（若设备信息有变化，那么在健康状态检查周期内就会上报，如果设备信息没有变化，那么上报周期固定为5分钟）后，执行以下命令，查看device-info-cm中ManuallySeparateNPU是否存在刚才删除的芯片名称。若不存在，则芯片恢复健康成功，可继续正常使用该芯片。

        ```shell
        kubectl describe cm -n kube-system {configMapName}
        ```

### ClusterD<a name="ZH-CN_TOPIC_0000002511346521_03"></a>

#### 配置说明<a name="ZH-CN_TOPIC_0000002511346521_04"></a>

断点续训针对芯片故障，支持按故障频率的配置进行处理。

针对芯片故障的不同级别进行分级处理时，ClusterD组件会获取到当前故障的故障码和故障级别，对于除了NotHandleFault和SubHealthFault级别之外的故障，根据ConfigMap（clusterd-config-cm）中配置的故障频率，将芯片状态置为人工隔离。该ConfigMap的参数说明请参见[表1](../../installation_guide/03_installation/manual_installation/06_clusterd.md)。

>[!NOTE] 
>
>- ConfigMap（clusterd-config-cm）为系统配置，若用户无特殊需求，请勿随意修改。若用户需要修改人工隔离芯片检测开关及故障频率、解除隔离时间等，可以通过修改该ConfigMap实现，修改方法请参见[（可选）配置芯片故障频率](#可选配置芯片故障频率)。
>- 不支持配置故障码检测范围，ClusterD会基于Ascend Device Plugin上报的故障级别进行判断。对于除了NotHandleFault和SubHealthFault级别之外的故障，都会进入人工隔离芯片检测流程。

#### （可选）配置芯片故障频率<a name="ZH-CN_TOPIC_0000002511426473_01"></a>

在安装ClusterD时，会自动创建ConfigMap（clusterd-config-cm），作为当前人工隔离芯片的检测依据。该ConfigMap的参数说明请参见[表1](../../installation_guide/03_installation/manual_installation/06_clusterd.md)。

如果用户想要自定义芯片故障频率，可以通过修改该ConfigMap实现。如果修改后的ConfigMap内容存在格式错误等问题，ClusterD会保留上一次读取成功的配置作为当前人工隔离芯片的检测依据。若ClusterD启动时，读取到的ConfigMap内容错误，则人工隔离芯片检测机制会默认关闭，直到格式和内容正确。

**操作步骤<a name="section14190101"></a>**

以人工隔离芯片的阈值由默认的24小时内出现3次调整为24小时内出现5次为例。

1. 登录环境，执行以下命令，查询当前配置。

    ```shell
    kubectl describe cm -n cluster-system clusterd-config-cm
    ```

    - 如果存在clusterd-config-cm，则执行[步骤3](#li01010203)进行编辑。
    - 如果不存在clusterd-config-cm，则执行[步骤2](#li010102)进行创建。

    >[!NOTE] 
    >正常情况下存在clusterd-config-cm。若不存在，需确认ClusterD的安装过程是否存在错误。
    
2. <a name="li010102"></a>创建人工隔离芯片检测所需的ConfigMap（clusterd-config-cm）。
    
    将以下内容保存为文件cm.yaml：

    ```Yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: clusterd-config-cm
          namespace: cluster-system
        data:
          manually_separate_policy.conf: |
            enabled: true
            separate:
              fault_window_hours: 24
              fault_threshold: 3
            release:
              fault_free_hours: 48

    ```

    执行以下命令：

    ```shell
    kubectl apply -f cm.yaml
    ```

    回显示例如下，说明创建成功。

    ```ColdFusion
    configmap/clusterd-config-cm created
    ```

3. <a name="li01010203"></a>执行以下命令，编辑clusterd-config-cm。

    ```shell
    kubectl edit cm -n cluster-system clusterd-config-cm
    ```

    根据实际情况，修改人工隔离芯片的故障频率。参数说明请参见[表1](../../installation_guide/03_installation/manual_installation/06_clusterd.md)。

    ```Yaml
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
      manually_separate_policy.conf: |
        # 修改人工隔离芯片检测开关
        enabled: true
        separate:
          # 修改人工隔离芯片的故障频率
          fault_window_hours: 24
          fault_threshold: 5   # 由3修改为5
        release:
          # 修改解除隔离时间
          fault_free_hours: 48
    kind: ConfigMap
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: |
          {"apiVersion":"v1","data":{"manually_separate_policy.conf":"enabled: true\nseparate:\n  fault_window_hours: 24\n  fault_threshold: 3\nrelease:\n  fault_free_hours: 48\n"},"kind":"ConfigMap","metadata":{"annotations":{},"name":"clusterd-config-cm","namespace":"cluster-system"}}
      creationTimestamp: "2026-02-24T11:25:19Z"
      name: clusterd-config-cm
      namespace: cluster-system
      resourceVersion: "3344125"
      selfLink: /api/v1/namespaces/cluster-system/configmaps/clusterd-config-cm
      uid: 68210bfc-f742-4765-a497-b61e9cc6b1a6
    ```

4. 修改完成后，按“Esc”键，输入:wq!保存并退出。
5. 等clusterd-config-cm更新生效（ClusterD的检测周期为300s）后，查看操作是否成功。
    1. 执行以下命令，查询ClusterD组件日志名称。

        ```shell
        kubectl get pods -A | grep clusterd
        ```

        回显示例如下：

        ```ColdFusion
        mindx-dl      clusterd-559bf4bd6-z9hv4   1/1     Running   0             4m23s
        ```

    2. 通过查询到的组件日志名称，查询ClusterD的组件日志信息。

        ```shell
        kubectl logs -f -n mindx-dl clusterd-559bf4bd6-z9hv4
        ```

        >[!NOTE] 
        >- 若日志出现“load manually separate policy config success”，表示手动修改人工隔离芯片的故障频率操作成功。
        >- 若日志出现“node: xx, dev: xx, code: xx is not found in manual fault cache, add”，表示该故障触发人工隔离。
        >- 若日志出现“node: xx, dev: xx, code: xx is found in manual fault cache, update last separate time”，表示已经触发人工隔离芯片的故障，再一次达到了人工隔离的故障频率，会刷新clusterd-manual-info-cm中的LastSeparateTime。clusterd-manual-info-cm的说明请参见[clusterd-manual-info-cm](../../api/clusterd/00_cluster_resources.md#clusterd-manual-info-cm)。

6. （可选）手动恢复人工隔离的芯片。故障的处理策略为ManuallySeparateNPU时，故障恢复后该芯片也处于隔离状态，可以手动恢复人工隔离的芯片。

    1. 执行以下命令，编辑ConfigMap clusterd-manual-info-cm。

        ```shell
        kubectl edit cm -n cluster-system clusterd-manual-info-cm
        ```

    2. 将Data下面的Total字段后面需要解除人工隔离的芯片名称删除。例如：Ascend910-2。

        ```json
        Name:         clusterd-manual-info-cm
        Namespace:    cluster-system
        Labels:       <none>
        Annotations:  <none>
         
        Data
        ====
        localhost.localdomain:
        ----
        {"Total":["Ascend910-0","Ascend910-2","Ascend910-3"],"Detail":{"Ascend910-0":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}],"Ascend910-2":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}],"Ascend910-3":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}]}}
         
        Events:  <none> 
        ```

    3. 修改完成后，按“Esc”键，输入:wq!保存并退出。
    4. 等待15s后，执行以下命令，查看clusterd-manual-info-cm中Ascend910-2是否还存在于Total和Detail字段中。同时，需要查看该芯片的ManuallySeparateNPU故障是否存在于cluster-info-device-\${m}中。若不存在，则芯片解除人工隔离成功，可继续正常使用该芯片。

        ```shell
        kubectl describe cm -n cluster-system clusterd-manual-info-cm
        ```

        >[!NOTE] 
        >- 仅支持删除Total字段中的芯片，不支持手动添加。其他内容不支持修改。
        >- 手动恢复人工隔离的芯片后，该芯片的故障计数会清零，再次达到频率时才会再次触发人工隔离。
        >- 若需要删除节点上所有的人工隔离芯片，则需删除Total字段后面的所有芯片名称，并将取值设置为空[]。如果想一次性解除所有的人工隔离芯片，可以直接将clusterd-manual-info-cm删除。
        >- ClusterD启动后15s内，暂时先不要修改clusterd-manual-info-cm，以免发生数据错误。

## 参数面网络故障<a name="ZH-CN_TOPIC_0000002479226486"></a>

### 总线设备故障<a name="ZH-CN_TOPIC_0000002511346423"></a>

#### 配置文件说明<a name="ZH-CN_TOPIC_0000002511346513"></a>

针对**总线设备**故障的**不同级别**进行分级处理时，Ascend Device Plugin组件会获取到当前故障的故障码，根据**SwitchFaultCode.json**中故障码配置的故障级别，对故障进行相应处理。SwitchFaultCode.json为系统配置文件，若用户无特殊需求，请勿随意修改。若用户需要修改故障码对应的故障级别，可以通过由faultCode.json和SwitchFaultCode.json创建的**mindx-dl-fault-config**文件实现。

>[!NOTE] 
>只有Atlas A3 训练系列产品存在**总线设备**，该设备的故障码可以查看SwitchFaultCode.json文件。

**SwitchFaultCode.json中的故障级别<a name="section681495612012"></a>**

断点续训针对**总线设备**故障的不同级别进行分级处理。若用户需要修改故障码的故障级别，操作指导请参见[（可选）配置总线设备故障级别](#可选配置总线设备故障级别)。

Ascend Device Plugin从驱动获取到故障码后，将根据故障码对设备及业务的影响将故障划分为几种级别并进行相应的重调度处理，详细说明请参见[故障级别及处理说明表](../../api/ascend_device_plugin.md#自定义灵衢设备故障)。

#### （可选）配置总线设备故障级别<a name="ZH-CN_TOPIC_0000002511426433"></a>

在制作Ascend Device Plugin镜像时，会将故障级别配置文件**SwitchFaultCode.json**内置在镜像中，启动Ascend Device Plugin时会读取这个文件的默认配置，作为当前故障处理依据。

如果用户想要自定义故障级别或者优雅容错相关配置，可以在集群中创建ConfigMap文件（mindx-dl-fault-config）。

- 如果Ascend Device Plugin启动时，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin会优先按照已存在的mindx-dl-fault-config中配置的内容，作为当前故障处理依据。
- 如果重新安装Ascend Device Plugin后，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin的默认**SwitchFaultCode.json**将不会生效，使用集群中已经存在的mindx-dl-fault-config。
- 如果重新安装Ascend Device Plugin后，集群中已经存在mindx-dl-fault-config且该ConfigMap中存在SwitchFaultCode.json字段，Ascend Device Plugin的默认SwitchFaultCode.json将不会生效，使用集群中已经存在的mindx-dl-fault-config。
- 若想要使用SwitchFaultCode.json默认配置，可以删除mindx-dl-fault-config，使Ascend Device Plugin读取默认SwitchFaultCode.json文件。
- 如果ConfigMap文件内容存在格式错误等问题，Ascend Device Plugin会默认读取镜像中内置的ConfigMap文件的内容，作为当前故障处理依据。

**使用SwitchFaultCode.json配置故障级别<a name="section067783615137"></a>**

以总线设备故障码\[0x00f1ff09,155913,cpu,na\]为例。该故障码由四部分组成：告警ID、故障ID、对端设备类型、端口号，如[表1 故障码说明](#zh-cn_topic_0000002007978080_table167355241939)所示。

**表 1**  故障码说明

<a name="zh-cn_topic_0000002007978080_table167355241939"></a>

|参数|说明|取值|
|--|--|--|
|告警ID|在以上示例中，告警ID为0x00f1ff09。|带内带外一致。|
|故障ID|在以上示例中，故障ID为155913。|带内带外一致。|
|对端设备类型|该故障所对应的对端设备类型。在以上示例中，对端设备类型为cpu。|<ul><li>取值为na：该故障为芯片故障，不涉及对端设备。</li><li>取值为cpu：该故障所对应的对端设备为CPU。</li><li>取值为npu：该故障所对应的对端设备为NPU。</li><li>取值为L2：该故障所对应的对端设备为L2。</li></ul>|
|端口号|在以上示例中，端口号为na。|取值只能为na。|

将当前故障的处理策略NotHandleFaultCodes（无需处理）修改为SeparateFaultCodes（隔离芯片，进行任务重调度）的操作示例如下。

1. 登录环境，进入Ascend Device Plugin解压目录。
2. 执行以下命令，查询是否已经基于SwitchFaultCode.json文件创建了mindx-dl-fault-config。

    ```shell
    kubectl describe cm -n kube-system mindx-dl-fault-config
    ```

    - 如果mindx-dl-fault-config已经存在，且存在SwitchFaultCode.json的相关字段，执行[步骤4](#zh-cn_topic_0000002007978080_li1014819812423)编辑该文件。
    - 如果mindx-dl-fault-config已经存在，但是不存在SwitchFaultCode.json的相关字段，需要先保存mindx-dl-fault-config内容，再删除mindx-dl-fault-config文件后，执行[步骤3](#zh-cn_topic_0000002007978080_li14147485427)创建该文件。
    - 如果不存在mindx-dl-fault-config，执行[步骤3](#zh-cn_topic_0000002007978080_li14147485427)创建该文件。

3. <a name="zh-cn_topic_0000002007978080_li14147485427"></a>执行以下命令，创建动态配置故障码所需ConfigMap文件（mindx-dl-fault-config）。

    ```shell
    kubectl create cm mindx-dl-fault-config -n kube-system  --from-file=./faultCode.json --from-file=./SwitchFaultCode.json --from-literal="PollInterval=300"
    ```

    回显示例如下。

    ```ColdFusion
    configmap/mindx-dl-fault-config created
    ```

    **表 2**  参数说明

    <a name="zh-cn_topic_0000002007978080_table14147138184211"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002007978080_row1814716812426"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002007978080_p141471483423"><a name="zh-cn_topic_0000002007978080_p141471483423"></a><a name="zh-cn_topic_0000002007978080_p141471483423"></a>参数名</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002007978080_p101477811428"><a name="zh-cn_topic_0000002007978080_p101477811428"></a><a name="zh-cn_topic_0000002007978080_p101477811428"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002007978080_p1014718154210"><a name="zh-cn_topic_0000002007978080_p1014718154210"></a><a name="zh-cn_topic_0000002007978080_p1014718154210"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002007978080_row1514810811424"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002007978080_p714817819421"><a name="zh-cn_topic_0000002007978080_p714817819421"></a><a name="zh-cn_topic_0000002007978080_p714817819421"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002007978080_p201488804220"><a name="zh-cn_topic_0000002007978080_p201488804220"></a><a name="zh-cn_topic_0000002007978080_p201488804220"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002007978080_p161481689426"><a name="zh-cn_topic_0000002007978080_p161481689426"></a><a name="zh-cn_topic_0000002007978080_p161481689426"></a>动态配置故障码所需的<span id="zh-cn_topic_0000002007978080_ph214819813425"><a name="zh-cn_topic_0000002007978080_ph214819813425"></a><a name="zh-cn_topic_0000002007978080_ph214819813425"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002007978080_row814819819422"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002007978080_p101481589424"><a name="zh-cn_topic_0000002007978080_p101481589424"></a><a name="zh-cn_topic_0000002007978080_p101481589424"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002007978080_p214814819427"><a name="zh-cn_topic_0000002007978080_p214814819427"></a><a name="zh-cn_topic_0000002007978080_p214814819427"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002007978080_p1614814815424"><a name="zh-cn_topic_0000002007978080_p1614814815424"></a><a name="zh-cn_topic_0000002007978080_p1614814815424"></a>mindx-dl-fault-config所在命名空间，不能修改该命名空间名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002007978080_row1714868114215"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002007978080_p182611591222"><a name="zh-cn_topic_0000002007978080_p182611591222"></a><a name="zh-cn_topic_0000002007978080_p182611591222"></a>SwitchFaultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002007978080_p1314868184217"><a name="zh-cn_topic_0000002007978080_p1314868184217"></a><a name="zh-cn_topic_0000002007978080_p1314868184217"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002007978080_p17148118174218"><a name="zh-cn_topic_0000002007978080_p17148118174218"></a><a name="zh-cn_topic_0000002007978080_p17148118174218"></a>用于保存故障码，必须与SwitchFaultCode.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

4. <a name="zh-cn_topic_0000002007978080_li1014819812423"></a>执行以下命令，编辑mindx-dl-fault-config文件。

    ```shell
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

5. 在mindx-dl-fault-config文件中，找到故障码\[0x00f1ff09,155913,cpu,na\]。

    ```json
    Data
    ====
    SwitchFaultCode.json:
    ----
    {"NotHandleFaultCodes":[0x00f1ff09,155913,cpu,na],
    ...
    ```

6. 将故障码从NotHandleFaultCodes中删除，并添加到SeparateFaultCodes中。

    ```json
    Data
    ====
    SwitchFaultCode.json:
    ----
    {"NotHandleFaultCodes":[],
    ```

    ```json
    ...
    "SeparateFaultCodes":["0x00f1ff09,155913,cpu,na","[0x00f103b0,155907,na,na]"…]
    }
    ```

7. 修改完成后，按“Esc”键，输入:wq!保存并退出。
8. 等mindx-dl-fault-config文件更新生效后（PollInterval取值，不指定则为300s），查看操作是否成功。
    1. 执行以下命令，查询Ascend Device Plugin组件日志名称。

        ```shell
        kubectl get pods -A | grep ascend-device-plugin
        ```

        回显示例如下：

        ```ColdFusion
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2. 通过查询到的组件日志名称，查询Ascend Device Plugin的组件日志信息。

        ```shell
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        若日志出现“load switch fault code from configmap success”，表示手动配置故障码操作成功。

### 关联故障<a name="ZH-CN_TOPIC_0000002511426403"></a>

#### 配置文件说明<a name="ZH-CN_TOPIC_0000002479386560"></a>

断点续训针对关联故障（特殊故障会伴生其他相关联的故障场景），需要忽略特殊故障诱发的伴生故障。ClusterD组件会获取到特殊故障，根据**relationFaultCustomization.json**和**faultDuration.json**文件中配置的关联故障策略对故障任务进行特殊处理。

relationFaultCustomization.json、faultDuration.json为系统配置文件，若用户无特殊需求，请勿随意修改。

**表 1**  relationFaultCustomization文件说明

<a name="zh-cn_topic_0000002157130117_table5148194813113"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002157130117_row1614914482114"><th class="cellrowborder" valign="top" width="13.701370137013702%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002157130117_p278365710116"><a name="zh-cn_topic_0000002157130117_p278365710116"></a><a name="zh-cn_topic_0000002157130117_p278365710116"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="69.05690569056905%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002157130117_p127832571915"><a name="zh-cn_topic_0000002157130117_p127832571915"></a><a name="zh-cn_topic_0000002157130117_p127832571915"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="17.241724172417243%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002157130117_p47831857912"><a name="zh-cn_topic_0000002157130117_p47831857912"></a><a name="zh-cn_topic_0000002157130117_p47831857912"></a>取值</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002157130117_row1514912481715"><td class="cellrowborder" valign="top" width="13.701370137013702%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p14783115717117"><a name="zh-cn_topic_0000002157130117_p14783115717117"></a><a name="zh-cn_topic_0000002157130117_p14783115717117"></a>TriggerFault</p>
</td>
<td class="cellrowborder" valign="top" width="69.05690569056905%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p1878313577120"><a name="zh-cn_topic_0000002157130117_p1878313577120"></a><a name="zh-cn_topic_0000002157130117_p1878313577120"></a>伴生故障码，当前支持faultCode.json和SwitchFaultCode.json配置的故障码。</p>
</td>
<td class="cellrowborder" valign="top" width="17.241724172417243%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p117831557615"><a name="zh-cn_topic_0000002157130117_p117831557615"></a><a name="zh-cn_topic_0000002157130117_p117831557615"></a>字符串</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row1714944814110"><td class="cellrowborder" valign="top" width="13.701370137013702%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p6783657411"><a name="zh-cn_topic_0000002157130117_p6783657411"></a><a name="zh-cn_topic_0000002157130117_p6783657411"></a>RelationFaults</p>
</td>
<td class="cellrowborder" valign="top" width="69.05690569056905%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p278411575113"><a name="zh-cn_topic_0000002157130117_p278411575113"></a><a name="zh-cn_topic_0000002157130117_p278411575113"></a>需要被关联的故障列表，可以是一个或多个故障码。当前支持faultCode.json和SwitchFaultCode.json配置的故障码。</p>
</td>
<td class="cellrowborder" valign="top" width="17.241724172417243%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p178414571018"><a name="zh-cn_topic_0000002157130117_p178414571018"></a><a name="zh-cn_topic_0000002157130117_p178414571018"></a>字符串列表</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row111493481216"><td class="cellrowborder" valign="top" width="13.701370137013702%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p578414571818"><a name="zh-cn_topic_0000002157130117_p578414571818"></a><a name="zh-cn_topic_0000002157130117_p578414571818"></a>FaultStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="69.05690569056905%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p178405710112"><a name="zh-cn_topic_0000002157130117_p178405710112"></a><a name="zh-cn_topic_0000002157130117_p178405710112"></a>关联故障匹配成功时对应任务的处理策略。</p>
<a name="zh-cn_topic_0000002157130117_ul17849570118"></a><a name="zh-cn_topic_0000002157130117_ul17849570118"></a><ul id="zh-cn_topic_0000002157130117_ul17849570118"><li>Separate：任务隔离</li><li>SubHealth：任务亚健康</li></ul>
</td>
<td class="cellrowborder" valign="top" width="17.241724172417243%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p1378413577119"><a name="zh-cn_topic_0000002157130117_p1378413577119"></a><a name="zh-cn_topic_0000002157130117_p1378413577119"></a>字符串</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row84116191226"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p114317681616"><a name="zh-cn_topic_0000002157130117_p114317681616"></a><a name="zh-cn_topic_0000002157130117_p114317681616"></a>注：</p>
<p id="zh-cn_topic_0000002157130117_p47413216213"><a name="zh-cn_topic_0000002157130117_p47413216213"></a><a name="zh-cn_topic_0000002157130117_p47413216213"></a>当设备发生配置的RelationFaults时，<span id="zh-cn_topic_0000002157130117_ph12291515161616"><a name="zh-cn_topic_0000002157130117_ph12291515161616"></a><a name="zh-cn_topic_0000002157130117_ph12291515161616"></a>ClusterD</span>会将对应的故障加入待处理的故障码队列。在配置的TimeOutInterval时间内，如果发生了TriggerFault对应的故障，会按照用户配置的FaultStrategy策略对任务进行处理。如果超过配置的TimeOutInterval时间，总线设备故障类型，按照任务亚健康进行处理，芯片故障或者参数面网络故障，会忽略该故障。</p>
</td>
</tr>
</tbody>
</table>

**表 2**  faultDuration.json文件说明

<a name="zh-cn_topic_0000002157130117_table1484617498414"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002157130117_row1284615492415"><th class="cellrowborder" valign="top" width="13.36133613361336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002157130117_p116699222514"><a name="zh-cn_topic_0000002157130117_p116699222514"></a><a name="zh-cn_topic_0000002157130117_p116699222514"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="70.36703670367037%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002157130117_p56691922055"><a name="zh-cn_topic_0000002157130117_p56691922055"></a><a name="zh-cn_topic_0000002157130117_p56691922055"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="16.271627162716275%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002157130117_p466911221257"><a name="zh-cn_topic_0000002157130117_p466911221257"></a><a name="zh-cn_topic_0000002157130117_p466911221257"></a>取值</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002157130117_row084615491413"><td class="cellrowborder" valign="top" width="13.36133613361336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p066920221954"><a name="zh-cn_topic_0000002157130117_p066920221954"></a><a name="zh-cn_topic_0000002157130117_p066920221954"></a>FaultCode</p>
</td>
<td class="cellrowborder" valign="top" width="70.36703670367037%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p96702227514"><a name="zh-cn_topic_0000002157130117_p96702227514"></a><a name="zh-cn_topic_0000002157130117_p96702227514"></a>故障码，当前支持faultCode.json和SwitchFaultCode.json配置的故障码。</p>
</td>
<td class="cellrowborder" valign="top" width="16.271627162716275%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p56701922954"><a name="zh-cn_topic_0000002157130117_p56701922954"></a><a name="zh-cn_topic_0000002157130117_p56701922954"></a>字符串</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row18467491043"><td class="cellrowborder" valign="top" width="13.36133613361336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p167022212517"><a name="zh-cn_topic_0000002157130117_p167022212517"></a><a name="zh-cn_topic_0000002157130117_p167022212517"></a>FaultType</p>
</td>
<td class="cellrowborder" valign="top" width="70.36703670367037%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p667020225515"><a name="zh-cn_topic_0000002157130117_p667020225515"></a><a name="zh-cn_topic_0000002157130117_p667020225515"></a>故障类型：</p>
<a name="zh-cn_topic_0000002157130117_ul1367017221559"></a><a name="zh-cn_topic_0000002157130117_ul1367017221559"></a><ul id="zh-cn_topic_0000002157130117_ul1367017221559"><li>faultDevice：芯片故障或者参数面网络故障</li><li>faultSwitch：总线设备故障</li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.271627162716275%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p967010221359"><a name="zh-cn_topic_0000002157130117_p967010221359"></a><a name="zh-cn_topic_0000002157130117_p967010221359"></a>字符串</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row208478499416"><td class="cellrowborder" valign="top" width="13.36133613361336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p16713225511"><a name="zh-cn_topic_0000002157130117_p16713225511"></a><a name="zh-cn_topic_0000002157130117_p16713225511"></a>TimeOutInterval</p>
</td>
<td class="cellrowborder" valign="top" width="70.36703670367037%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p36711221159"><a name="zh-cn_topic_0000002157130117_p36711221159"></a><a name="zh-cn_topic_0000002157130117_p36711221159"></a>故障码最长被关联时间。单位为秒。</p>
</td>
<td class="cellrowborder" valign="top" width="16.271627162716275%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p186718221450"><a name="zh-cn_topic_0000002157130117_p186718221450"></a><a name="zh-cn_topic_0000002157130117_p186718221450"></a>整数</p>
</td>
</tr>
</tbody>
</table>

#### （可选）配置关联故障的处理策略<a name="ZH-CN_TOPIC_0000002479226478"></a>

在制作ClusterD镜像时，会将关联故障的两个配置文件内置在镜像中，启动ClusterD会读取这两个文件的默认配置，作为当前故障处理依据。

如果用户想要自定义关联的故障码以及对应的处理策略。可以在制作ClusterD镜像时，修改对应的relationFaultCustomization.json和faultDuration.json配置文件。

**操作步骤<a name="zh-cn_topic_0000002157048501_section2086912531189"></a>**

以RelationFaults为故障码81078603，TriggerFault为故障码8C1F8609为例。如果发生了芯片81078603的故障码，需要在后面60s内出现8C1F8609故障时忽略8C1F8609故障，并且隔离发生的81078603故障的任务。可以手动配置关联故障的处理策略为Separate。

1. 登录环境，进入ClusterD解压后的目录。
2. 执行**vi relationFaultCustomization.json**命令编辑配置文件。

    ```shell
    vi relationFaultCustomization.json
    ```

    将2个故障进行关联。修改完成后，按“Esc”键，输入:wq!保存并退出。

    ```json
    …
      {
        "TriggerFault": "8C1F8609",
        "RelationFaults": [
          "81078603"
        ],
        "FaultStrategy": "Separate"
      }
    …
    ```

3. 执行**vi faultDuration.json**命令编辑配置文件。

    ```shell
    vi faultDuration.json
    ```

    配置故障类型、故障关联时间等。修改完成后，按“Esc”键，输入:wq!保存并退出。

    ```json
    …
      {
        "FaultCode": "81078603",
        "FaultType": "faultDevice",
        "TimeOutInterval": 60
      }
    …
    ```

## 公共故障<a name="ZH-CN_TOPIC_0000002479386564"></a>

### 配置文件说明<a name="ZH-CN_TOPIC_0000002511346487"></a>

断点续训针对公共故障的不同级别进行分级处理。ClusterD组件会获取到当前故障的故障码，根据publicFaultConfiguration.json文件中故障码配置的故障级别，对故障进行相应处理。特殊情况下，若ClusterD收到了无法识别的故障码（未保存在配置文件中），会将此故障丢弃。

[publicFaultConfiguration.json](#zh-cn_topic_0000002181110120_table8202741102717)为公共故障的系统配置文件，若用户无特殊需求，请勿随意修改。若用户需要修改公共故障的级别和发送方，可以通过在/user1/mindx-dl/clusterd写入自定义配置文件publicCustomization.json实现。该文件路径支持配置，配置方式如下所示。

>[!NOTE] 
>
>- 文件publicCustomization.json在容器内路径为/user1/mindx-dl/clusterd，不支持修改，不支持软链接；主机路径默认为/user1/mindx-dl/clusterd。
>- 主机路径可由用户根据实际情况自行配置：在ClusterD的启动YAML中修改挂载卷名称为config-clusterd的主机挂载路径。
>- 多master场景下，建议每个master节点上都同步一份最新的publicCustomization.json文件。避免重启ClusterD后，ClusterD被调度到其他master节点，从而导致自定义故障配置文件丢失的问题。

**表 1**  故障级别及处理说明

<a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_table169151711124319"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row19916131120434"><th class="cellrowborder" valign="top" width="15.09499941718149%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"></a>故障级别</p>
</th>
<th class="cellrowborder" valign="top" width="42.54575125305979%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"></a>故障处理策略</p>
</th>
<th class="cellrowborder" valign="top" width="42.35924932975871%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002181110120_p2218314171716"><a name="zh-cn_topic_0000002181110120_p2218314171716"></a><a name="zh-cn_topic_0000002181110120_p2218314171716"></a>重调度处理</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row6916711144312"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"></a>无需处理</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"></a>暂不处理</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row1991661104316"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"></a>无法恢复，需要隔离芯片</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p12165431710"><a name="zh-cn_topic_0000002181110120_p12165431710"></a><a name="zh-cn_topic_0000002181110120_p12165431710"></a>隔离芯片，进行任务重调度。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row191716112431"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"></a>根据任务YAML中配置的subHealthyStrategy参数取值进行处理，详细请参见<a href="./06_configuring_the_job_yaml_file.md#yaml参数说明">YAML参数说明</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"></a>当芯片出现亚健康故障时，需根据<a href="./06_configuring_the_job_yaml_file.md#任务yaml配置示例">任务YAML配置示例</a>策略进行处理。</p>
<div class="note" id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_note7936204710536"></a><div class="notebody"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"></a>如果后续芯片出现其他级别故障，此时SubHealthFault处理策略不影响其他级别的故障处理。</p>
</div></div>
</td>
</tr>
<tr id="row16800523414"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="p88011823817"><a name="p88011823817"></a><a name="p88011823817"></a><span id="ph1339214581915"><a name="ph1339214581915"></a><a name="ph1339214581915"></a>PreSeparateNPU</span></p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="p980117231413"><a name="p980117231413"></a><a name="p980117231413"></a><span id="ph739245817113"><a name="ph739245817113"></a><a name="ph739245817113"></a>暂不影响业务，后续不再调度任务到该芯片。</span></p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="p1280114235116"><a name="p1280114235116"></a><a name="p1280114235116"></a><span id="ph3392758212"><a name="ph3392758212"></a><a name="ph3392758212"></a>预隔离芯片。</span></p>
</td>
</tr>
</tbody>
</table>

**表 2**  publicFaultConfiguration.json字段说明

<a name="zh-cn_topic_0000002181110120_table8202741102717"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row18202164117272"><th class="cellrowborder" valign="top" width="28.93%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002181110120_p1120213413271"><a name="zh-cn_topic_0000002181110120_p1120213413271"></a><a name="zh-cn_topic_0000002181110120_p1120213413271"></a>参数名称</p>
</th>
<th class="cellrowborder" valign="top" width="71.07%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002181110120_p22024417279"><a name="zh-cn_topic_0000002181110120_p22024417279"></a><a name="zh-cn_topic_0000002181110120_p22024417279"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row172028412278"><td class="cellrowborder" valign="top" width="28.93%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p1220219412279"><a name="zh-cn_topic_0000002181110120_p1220219412279"></a><a name="zh-cn_topic_0000002181110120_p1220219412279"></a><a href="#zh-cn_topic_0000002181110120_table1689274753416">publicFaultCode</a></p>
</td>
<td class="cellrowborder" valign="top" width="71.07%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p220284110271"><a name="zh-cn_topic_0000002181110120_p220284110271"></a><a name="zh-cn_topic_0000002181110120_p220284110271"></a>公共故障码相关配置。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row14606121802219"><td class="cellrowborder" valign="top" width="28.93%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p1760617182224"><a name="zh-cn_topic_0000002181110120_p1760617182224"></a><a name="zh-cn_topic_0000002181110120_p1760617182224"></a>publicFaultResource</p>
</td>
<td class="cellrowborder" valign="top" width="71.07%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p1606118102218"><a name="zh-cn_topic_0000002181110120_p1606118102218"></a><a name="zh-cn_topic_0000002181110120_p1606118102218"></a>公共故障发送方配置。</p>
</td>
</tr>
</tbody>
</table>

**表 3**  publicFaultCode字段说明

<a name="zh-cn_topic_0000002181110120_table1689274753416"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row16892144733413"><th class="cellrowborder" valign="top" width="28.849999999999998%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002181110120_p689264723412"><a name="zh-cn_topic_0000002181110120_p689264723412"></a><a name="zh-cn_topic_0000002181110120_p689264723412"></a>参数名称</p>
</th>
<th class="cellrowborder" valign="top" width="71.15%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002181110120_p889274783418"><a name="zh-cn_topic_0000002181110120_p889274783418"></a><a name="zh-cn_topic_0000002181110120_p889274783418"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row28921647103410"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p48921847143412"><a name="zh-cn_topic_0000002181110120_p48921847143412"></a><a name="zh-cn_topic_0000002181110120_p48921847143412"></a>NotHandleFaultCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p58921747183416"><a name="zh-cn_topic_0000002181110120_p58921747183416"></a><a name="zh-cn_topic_0000002181110120_p58921747183416"></a>故障级别为NotHandleFault（无需处理）的故障码。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row989224719346"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p118928476343"><a name="zh-cn_topic_0000002181110120_p118928476343"></a><a name="zh-cn_topic_0000002181110120_p118928476343"></a>SubHealthFaultCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p17892947113410"><a name="zh-cn_topic_0000002181110120_p17892947113410"></a><a name="zh-cn_topic_0000002181110120_p17892947113410"></a>故障级别为SubHealthFault（亚健康）的故障码。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row289264713349"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p38921547193418"><a name="zh-cn_topic_0000002181110120_p38921547193418"></a><a name="zh-cn_topic_0000002181110120_p38921547193418"></a>SeparateNPUCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p689274714341"><a name="zh-cn_topic_0000002181110120_p689274714341"></a><a name="zh-cn_topic_0000002181110120_p689274714341"></a>故障级别为SeparateNPU（无法恢复，需要隔离芯片）的故障码。</p>
</td>
</tr>
<tr id="row107385344217"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="p187397341724"><a name="p187397341724"></a><a name="p187397341724"></a><span id="ph791817016319"><a name="ph791817016319"></a><a name="ph791817016319"></a>PreSeparateNPUCodes</span></p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="p15739113415210"><a name="p15739113415210"></a><a name="p15739113415210"></a><span id="ph8918120234"><a name="ph8918120234"></a><a name="ph8918120234"></a>故障级别为</span><span id="ph491890639"><a name="ph491890639"></a><a name="ph491890639"></a>PreSeparateNPU</span><span id="ph6918601336"><a name="ph6918601336"></a><a name="ph6918601336"></a>（暂不影响业务，后续不再调度任务到该芯片）的故障码。</span></p>
</td>
</tr>
</tbody>
</table>

**故障码说明<a name="zh-cn_topic_0000002181110120_section1440314273418"></a>**

公共故障的故障码为9位，说明如下。

**表 4**  故障码说明

<a name="table1237891465117"></a>
<table><thead align="left"><tr id="row1137891413516"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="p1937816143519"><a name="p1937816143519"></a><a name="p1937816143519"></a>位数</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="p837812144514"><a name="p837812144514"></a><a name="p837812144514"></a>描述</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="p14378201455110"><a name="p14378201455110"></a><a name="p14378201455110"></a>取值</p>
</th>
</tr>
</thead>
<tbody><tr id="row1137861419517"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p123782149514"><a name="p123782149514"></a><a name="p123782149514"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p15378914185120"><a name="p15378914185120"></a><a name="p15378914185120"></a>故障类型</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p10378161419517"><a name="p10378161419517"></a><a name="p10378161419517"></a>0：芯片故障</p>
<p id="p037871414515"><a name="p037871414515"></a><a name="p037871414515"></a>1：节点故障</p>
<p id="p33781414125113"><a name="p33781414125113"></a><a name="p33781414125113"></a>2：网络故障</p>
<p id="p10379101414516"><a name="p10379101414516"></a><a name="p10379101414516"></a>3：存储故障</p>
</td>
</tr>
<tr id="row337901415519"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p2379181475111"><a name="p2379181475111"></a><a name="p2379181475111"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p133796146513"><a name="p133796146513"></a><a name="p133796146513"></a>故障默认的级别</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p103791114185115"><a name="p103791114185115"></a><a name="p103791114185115"></a>0: NotHandleFault</p>
<p id="p193791214175112"><a name="p193791214175112"></a><a name="p193791214175112"></a>1: SubHealthFault</p>
<p id="p737991475119"><a name="p737991475119"></a><a name="p737991475119"></a>2: SeparateNPU</p>
</td>
</tr>
<tr id="row1737917147519"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p133793145514"><a name="p133793145514"></a><a name="p133793145514"></a>3、4</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1137901435119"><a name="p1137901435119"></a><a name="p1137901435119"></a>预留扩展位</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p43795142516"><a name="p43795142516"></a><a name="p43795142516"></a>暂为00</p>
</td>
</tr>
<tr id="row1337961495114"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17379121416515"><a name="p17379121416515"></a><a name="p17379121416515"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p17379141465112"><a name="p17379141465112"></a><a name="p17379141465112"></a>第6-9位的故障码是否为用户自定义，避免冲突</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1237917146517"><a name="p1237917146517"></a><a name="p1237917146517"></a>0：发布包中定义</p>
<p id="p12379191418513"><a name="p12379191418513"></a><a name="p12379191418513"></a>1：用户自定义</p>
</td>
</tr>
<tr id="row1937911425114"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p12379161465115"><a name="p12379161465115"></a><a name="p12379161465115"></a>6-9</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1437931425115"><a name="p1437931425115"></a><a name="p1437931425115"></a>具体的十进制故障码</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p8379121413512"><a name="p8379121413512"></a><a name="p8379121413512"></a>示例：1001</p>
</td>
</tr>
<tr id="row6379214165114"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p1137911410515"><a name="p1137911410515"></a><a name="p1137911410515"></a>示例如下：</p>
<p id="p1837941416513"><a name="p1837941416513"></a><a name="p1837941416513"></a>0100 01001：芯片故障，SubHealthFault，发布包中定义，故障1001。</p>
<p id="p1037911455117"><a name="p1037911455117"></a><a name="p1037911455117"></a>1000 11002：节点故障，NotHandleFault，用户自定义，故障1002。</p>
<p id="p8379181455115"><a name="p8379181455115"></a><a name="p8379181455115"></a>2200 01003：网络故障，SeparateNPU，发布包中定义，故障1003。</p>
</td>
</tr>
</tbody>
</table>

**已支持的公共故障<a name="zh-cn_topic_0000002181110120_section4960201383813"></a>**

**表 5**  已支持的公共故障

<a name="zh-cn_topic_0000002181110120_table31451934163811"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row514523493819"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002181110120_p1114523420389"><a name="zh-cn_topic_0000002181110120_p1114523420389"></a><a name="zh-cn_topic_0000002181110120_p1114523420389"></a>故障码</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002181110120_p9145143412387"><a name="zh-cn_topic_0000002181110120_p9145143412387"></a><a name="zh-cn_topic_0000002181110120_p9145143412387"></a>故障说明</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002181110120_p15145193413388"><a name="zh-cn_topic_0000002181110120_p15145193413388"></a><a name="zh-cn_topic_0000002181110120_p15145193413388"></a>默认故障级别</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row1514593415388"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_p181451134193811"><a name="zh-cn_topic_0000002181110120_p181451134193811"></a><a name="zh-cn_topic_0000002181110120_p181451134193811"></a>010001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_p814593412386"><a name="zh-cn_topic_0000002181110120_p814593412386"></a><a name="zh-cn_topic_0000002181110120_p814593412386"></a>光链路脏污（芯片故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p414533483811"><a name="zh-cn_topic_0000002181110120_p414533483811"></a><a name="zh-cn_topic_0000002181110120_p414533483811"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row175241157181818"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p580896101918"><a name="p580896101918"></a><a name="p580896101918"></a>210001007</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p15808166121915"><a name="p15808166121915"></a><a name="p15808166121915"></a>光链路脏污（网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1180917617197"><a name="p1180917617197"></a><a name="p1180917617197"></a>SubHealthFault</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row131782214434"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_p41752216438"><a name="zh-cn_topic_0000002181110120_p41752216438"></a><a name="zh-cn_topic_0000002181110120_p41752216438"></a>220001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_p1171822134316"><a name="zh-cn_topic_0000002181110120_p1171822134316"></a><a name="zh-cn_topic_0000002181110120_p1171822134316"></a>NPU卡<span id="ph17233131243911"><a name="ph17233131243911"></a><a name="ph17233131243911"></a>HCCS</span>网络故障</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p1566710511444"><a name="zh-cn_topic_0000002181110120_p1566710511444"></a><a name="zh-cn_topic_0000002181110120_p1566710511444"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row192881812184715"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3289131210473"><a name="p3289131210473"></a><a name="p3289131210473"></a>010001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1628951244719"><a name="p1628951244719"></a><a name="p1628951244719"></a>光链路松动（芯片故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p828971254715"><a name="p828971254715"></a><a name="p828971254715"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row38601828161910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p6168163671911"><a name="p6168163671911"></a><a name="p6168163671911"></a>210001008</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p316816364194"><a name="p316816364194"></a><a name="p316816364194"></a>光链路松动（网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2168436121911"><a name="p2168436121911"></a><a name="p2168436121911"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row172051674711"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p127201168472"><a name="p127201168472"></a><a name="p127201168472"></a>310001005</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1572071644717"><a name="p1572071644717"></a><a name="p1572071644717"></a>DPC客户端失效</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p141495491488"><a name="p141495491488"></a><a name="p141495491488"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row4720816104713"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17720131674712"><a name="p17720131674712"></a><a name="p17720131674712"></a>200001006</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1972020169475"><a name="p1972020169475"></a><a name="p1972020169475"></a>疑似光链路亚健康</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1572061684719"><a name="p1572061684719"></a><a name="p1572061684719"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row191121122184715"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p7112152234711"><a name="p7112152234711"></a><a name="p7112152234711"></a>210001009</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p10112152210476"><a name="p10112152210476"></a><a name="p10112152210476"></a>光模块器件亚健康</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2011213229474"><a name="p2011213229474"></a><a name="p2011213229474"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row19731102610435"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1324180124413"><a name="p1324180124413"></a><a name="p1324180124413"></a>220001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p13241019443"><a name="p13241019443"></a><a name="p13241019443"></a>备份超节点场景下，调度使用不存在的备份框资源。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p12161558145818"><a name="p12161558145818"></a><a name="p12161558145818"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row13731626174317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3241309446"><a name="p3241309446"></a><a name="p3241309446"></a>220001003</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1424190104412"><a name="p1424190104412"></a><a name="p1424190104412"></a>备份框资源端口故障</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p17362234428"><a name="p17362234428"></a><a name="p17362234428"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row127318268434"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p52416011444"><a name="p52416011444"></a><a name="p52416011444"></a>220001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p124800442"><a name="p124800442"></a><a name="p124800442"></a>备份框任务ID占用冲突</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p191615588586"><a name="p191615588586"></a><a name="p191615588586"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row20731826154318"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5241103443"><a name="p5241103443"></a><a name="p5241103443"></a>220001005</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p824705444"><a name="p824705444"></a><a name="p824705444"></a>NetMind失效</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1016110586589"><a name="p1016110586589"></a><a name="p1016110586589"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row673142624317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p162412019440"><a name="p162412019440"></a><a name="p162412019440"></a>220001006</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p10241109444"><a name="p10241109444"></a><a name="p10241109444"></a>疑似备份框链路端口部分失效</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p916119583580"><a name="p916119583580"></a><a name="p916119583580"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row673116264438"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p9247064419"><a name="p9247064419"></a><a name="p9247064419"></a>220001007</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p6249018447"><a name="p6249018447"></a><a name="p6249018447"></a>光链路调整失败</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p132215501211"><a name="p132215501211"></a><a name="p132215501211"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row8926105693315"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1792695643311"><a name="p1792695643311"></a><a name="p1792695643311"></a>200001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p18926165614336"><a name="p18926165614336"></a><a name="p18926165614336"></a>某节点内产生/恢复慢网络（慢网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p24091634153411"><a name="p24091634153411"></a><a name="p24091634153411"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row10526205273417"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3526052153416"><a name="p3526052153416"></a><a name="p3526052153416"></a>200001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p5804154074418"><a name="p5804154074418"></a><a name="p5804154074418"></a>超节点内的节点间产生/恢复慢网络。（慢网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p352695212349"><a name="p352695212349"></a><a name="p352695212349"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row663164316353"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17634437355"><a name="p17634437355"></a><a name="p17634437355"></a>200001012</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p18631743123513"><a name="p18631743123513"></a><a name="p18631743123513"></a>不是卡故障导致的慢网络（慢网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p101021310163612"><a name="p101021310163612"></a><a name="p101021310163612"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row178327182364"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p383221833611"><a name="p383221833611"></a><a name="p383221833611"></a>110001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1683231816361"><a name="p1683231816361"></a><a name="p1683231816361"></a>慢节点故障</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p6832131833614"><a name="p6832131833614"></a><a name="p6832131833614"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row1179514189380"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p979511810389"><a name="p979511810389"></a><a name="p979511810389"></a>100001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1579521818381"><a name="p1579521818381"></a><a name="p1579521818381"></a>劣化已恢复（慢节点故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p27883220394"><a name="p27883220394"></a><a name="p27883220394"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row121732048142813"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p4359165915289"><a name="p4359165915289"></a><a name="p4359165915289"></a>110001020</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1817494822816"><a name="p1817494822816"></a><a name="p1817494822816"></a>共享存储DPC进程异常</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p121741348132817"><a name="p121741348132817"></a><a name="p121741348132817"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row7277115416280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5277135492816"><a name="p5277135492816"></a><a name="p5277135492816"></a>110001021</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7277854132820"><a name="p7277854132820"></a><a name="p7277854132820"></a>共享存储DPC内存不足异常</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182781654132815"><a name="p182781654132815"></a><a name="p182781654132815"></a>SubHealthFault</p>
</td>
</tr>
</tbody>
</table>

### （可选）配置公共故障的级别和发送方<a name="ZH-CN_TOPIC_0000002479226494"></a>

在制作ClusterD镜像时，会将故障级别配置文件publicFaultConfiguration.json内置在镜像中，启动ClusterD时会读取这个文件的默认配置，作为当前故障处理依据。

如果用户想要自定义故障级别，可以在主机上创建/user1/mindx-dl/clusterd/publicCustomization.json文件。

- 如果ClusterD启动时，已经存在该文件，ClusterD会优先按照已存在的文件中配置的内容，作为当前故障处理依据。
- 如果重新安装ClusterD后，已经存在该文件，ClusterD的默认publicFaultConfiguration.json将不会生效，使用已经存在的publicCustomization.json文件。若想要使用publicFaultConfiguration.json的默认配置，可以删除已存在的publicCustomization.json文件，使ClusterD读取默认的publicFaultConfiguration.json文件。
- 如果publicCustomization.json文件内容存在格式错误等问题，ClusterD会默认读取镜像中内置的publicFaultConfiguration.json文件的内容，作为当前故障处理依据。

**配置公共故障码的故障级别<a name="zh-cn_topic_0000002180950420_section1384121854711"></a>**

配置公共故障码的故障级别分为以下2种场景。

- 对已有故障码的故障级别进行调整。
- 新增故障码及其故障级别。

    下面将以故障码010001008为例，介绍公共故障码故障级别的配置步骤。

1. 登录环境，进入/user1/mindx-dl/clusterd目录。
2. 执行**vi publicCustomization.json**命令，编辑文件。publicCustomization.json的详细说明请参见[表2](#ZH-CN_TOPIC_0000002511346487)。

    >[!NOTE] 
    >- 创建文件publicCustomization.json之后，用户需要保证该文件有ClusterD用户hwMindX的可读权限。例如，如果用户权限为root，该文件权限建议设置为644。
    >- 文件权限安全需要用户保证，如果权限过大，可能存在安全风险。

    ```json
    {
      "publicFaultCode": {
        "NotHandleFaultCodes":[],
        "SubHealthFaultCodes":[],
        "SeparateNPUCodes":["010001008"],
        "PreSeparateNPUCodes":[]
      },
      "publicFaultResource": [
        "CCAE", "fd-online", "pingmesh", "Netmind", "dpcStorage"
      ]
    }
    ```

3. 修改完成后，按“Esc”键，输入:wq!保存并退出。
4. 几秒钟后，文件生效。查看操作是否成功。

    若日志出现“load fault config from <publicCustomization.json\> success”，表示手动配置故障码操作成功。

**配置公共故障的发送方<a name="zh-cn_topic_0000002180950420_section5532327614"></a>**

下面将以新增故障发送方XXX为例，介绍公共故障码发送方的配置步骤。

1. 登录环境，进入/user1/mindx-dl/clusterd目录。
2. 执行**vi publicCustomization.json**命令，编辑文件。publicCustomization.json的详细说明请参见[表2](#ZH-CN_TOPIC_0000002511346487)。

    ```json
    {
      "publicFaultCode": {
        "NotHandleFaultCodes":[],
        "SubHealthFaultCodes":[],
        "SeparateNPUCodes":[],
        "PreSeparateNPUCodes":[]
      },
      "publicFaultResource": [
        "CCAE", "fd-online", "pingmesh", "Netmind", "dpcStorage", "XXX"
      ]
    }
    ```

3. 修改完成后，按“Esc”键，输入:wq!保存并退出。
4. 几秒钟后，文件生效。查看操作是否成功。

    若日志出现“load fault config from <publicCustomization.json\> success”，表示手动配置故障码操作成功。
