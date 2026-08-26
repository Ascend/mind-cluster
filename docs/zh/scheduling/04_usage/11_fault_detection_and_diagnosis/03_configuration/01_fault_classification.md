# 配置故障级别

## 配置说明<a name="ZH-CN_TOPIC_0000002479386448"></a>

故障检测特性针对节点故障中**节点硬件故障**、**芯片故障、灵衢总线设备故障**和**公共故障**的不同故障码，提供了默认的故障级别和对应级别的故障处理策略；**芯片故障**还提供了默认的故障频率和时长，以及对应的故障处理策略。

若用户需要修改故障处理策略可参见本章节。若无特殊需求，请勿随意修改。

**支持配置的故障级别说明<a name="section257513292065"></a>**

不同类型的故障支持配置的故障级别如下表所示。

**表 1**  支持配置的故障级别

<a name="table4710459145316"></a>
<table><thead align="left"><tr id="row37104590534"><th class="cellrowborder" valign="top" id="mcps1.2.5.1.1"><p id="p7710135925316"><a name="p7710135925316"></a>故障名称</p>
</th>
<th class="cellrowborder" colspan="3" valign="top" id="mcps1.2.5.1.2"><p id="p11175192213564"><a name="p11175192213564"></a>支持配置的故障级别</p>
</th>
</tr>
</thead>
<tbody><tr id="row271045905320"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p271015916536"><a name="p271015916536"></a>节点故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p66711187562"><a name="p66711187562"></a>NotHandleFault、PreSeparateFault、SeparateFault</p>
</td>
</tr>
<tr id="row3710165935311"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17710125955315"><a name="p17710125955315"></a>芯片故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p21371428713"><a name="p21371428713"></a>NotHandleFault、RestartRequest、RestartBusiness、FreeRestartNPU、RestartNPU、SeparateNPU、PreSeparateNPU、SubHealthFault</p>
</td>
</tr>
<tr id="row5710125913537"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p10710959185319"><a name="p10710959185319"></a>灵衢总线设备故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p6631112135616"><a name="p6631112135616"></a>NotHandleFault、SubHealthFault、ResetFault、SeparateFault<span id="ph51441721217"><a name="ph51441721217"></a>、</span><span id="ph375517710129"><a name="ph375517710129"></a>RestartRequestFault</span></p>
</td>
</tr>
<tr id="row416145918513"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p116115913517"><a name="p116115913517"></a>公共故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p147536536717"><a name="p147536536717"></a>NotHandleFault、SeparateNPU、SubHealthFault<span id="ph632635517598"><a name="ph632635517598"></a>、PreSeparateNPU</span></p>
</td>
</tr>
<tr id="row416145918513"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p116115913528"><a name="p116115913528"></a>任务卡死故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p147536536728"><a name="p147536536728"></a>NotHandleFault、SeparateNPU、PreSeparateNPU</p>
</td>
</tr>
<tr id="row416145918513"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p116115913528"><a name="p116115913528-duplicate-2"></a>光链路成员端口故障</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p147536536728"><a name="p147536536728-duplicate-2"></a>SeparateNPU、PreSeparateNPU、SubHealthFault</p>
</td>
</tr>
</tbody>
</table>

在以上表格中，每种故障级别的处理策略说明如下。

**表 2**  故障级别及处理说明

<a name="table103716651410"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row461812518228"><th class="cellrowborder" valign="top" width="19.06%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"></a>故障处理策略</p>
</th>
<th class="cellrowborder" valign="top" width="35.74%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="23.39%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"></a>重调度处理</p>
</th>
<th class="cellrowborder" valign="top" width="21.81%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"></a>优雅容错处理</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row961811511228"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"></a>对业务无影响的故障，无需处理。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"></a>暂不处理</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"></a>暂不处理</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row116184515226"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"></a>RestartRequest</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"></a>影响业务执行，需要重新执行业务请求。</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"></a>隔离芯片，进行任务重调度。</p>
<div class="note" id="note11901123612819"><a name="note11901123612819"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1069261722310"><a name="p1069261722310"></a>若推理任务订阅<span id="ph4356222144812"><a name="ph4356222144812"></a>了</span>故障信息，任务使用的推理卡上发生RestartRequest故障且故障持续时间未超过60秒，则不执行任务重调度；若故障持续时间超过60秒仍未恢复，则隔离芯片，进行任务重调度。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"></a>推理场景重新执行推理请求，训练场景重新执行训练业务。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row14618105116225"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"></a>RestartBusiness</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"></a>影响业务执行，需要重新执行业务。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"></a>重新执行业务</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row561825132214"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"></a>FreeRestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"></a>影响业务执行，待芯片空闲时需复位芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"></a>等待芯片空闲后复位芯片。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row14618125152210"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"></a>RestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"></a>影响业务执行，需立即复位芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"></a>立即停止训练业务，复位芯片后重新执行业务。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row1061895115227"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"></a>无法恢复，需要隔离芯片。</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"></a>隔离芯片，进行任务重调度。</p>
</td>
</tr>
<tr id="row870814247412"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p5708202454117"><a name="p5708202454117"></a>SeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p12708162474117"><a name="p12708162474117"></a>任务一定会受到影响。</p>
<div class="note" id="note1521013164613"><a name="note1521013164613"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p92101114465"><a name="p92101114465"></a>灵衢总线设备故障级别为SeparateFault时，表示业务运行失败，需更换器件或板卡。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p0708624204112"><a name="p0708624204112"></a>任务重调度</p>
<div class="note" id="note44451347164716"><a name="note44451347164716"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p64453471479"><a name="p64453471479"></a>灵衢总线设备故障下，本故障级别代表的故障处理策略为停止当前训练任务，隔离节点，进行任务重调度。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p137081824174117"><a name="p137081824174117"></a>-</p>
</td>
</tr>
<tr id="row5706333131216"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p177061833201220"><a name="p177061833201220"></a><span id="ph141513510124"><a name="ph141513510124"></a>RestartRequestFault</span></p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p070623351220"><a name="p070623351220"></a><span id="ph18501459184"><a name="ph18501459184"></a>业务运行失败，需要重新执行业务请求。</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p770653313124"><a name="p770653313124"></a><span id="ph38912127169"><a name="ph38912127169"></a>停止当前训练任务，隔离节点，进行任务重调度。</span></p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p6706113331213"><a name="p6706113331213"></a>推理场景重新执行推理请求，训练场景重新执行训练业务。</p>
</td>
</tr>
<tr id="row3938182254418"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p39381822174417"><a name="p39381822174417"></a>ResetFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p1193862274418"><a name="p1193862274418"></a>业务运行失败</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p184323519501"><a name="p184323519501"></a>停止当前训练任务，隔离节点，进行任务重调度。</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p18938822204411"><a name="p18938822204411"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row102215292529"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"></a>PreSeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"></a>暂不影响业务，后续不再调度任务到该芯片。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"></a>预隔离芯片</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"></a>预隔离芯片</p>
</td>
</tr>
<tr id="row84541721401"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p174559214016"><a name="p174559214016"></a>PreSeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p145562114011"><a name="p145562114011"></a>可能导致任务受到影响。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p54556214409"><a name="p54556214409"></a>该节点上有任务则不处理，后续调度时不调度任务到该节点。</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p1245572144015"><a name="p1245572144015"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row0352224175218"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"></a>根据任务YAML中配置的subHealthyStrategy参数取值进行处理，详细请参见<a href="../../../06_api/15_yaml_configuration.md#yaml_configuration">YAML配置说明</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"></a>当芯片出现亚健康故障时，需根据<a href="../../04_resumable_training/03_configuration/02_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002511426471">配置亚健康热切</a>进行处理。</p>
<div class="note" id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"></a>如果后续芯片出现其他级别故障，此时SubHealthFault处理策略不影响其他级别的故障处理。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"></a>暂不处理</p>
</td>
</tr>
</tbody>
</table>
