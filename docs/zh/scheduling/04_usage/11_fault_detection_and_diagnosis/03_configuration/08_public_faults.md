# 配置公共故障<a name="ZH-CN_TOPIC_0000002479386564"></a>

## 配置文件说明<a name="ZH-CN_TOPIC_0000002511346487"></a>

故障检测特性针对公共故障的不同级别进行分级处理。ClusterD组件会获取到当前故障的故障码，根据publicFaultConfiguration.json文件中故障码配置的故障级别，对故障进行相应处理。特殊情况下，若ClusterD收到了无法识别的故障码（未保存在配置文件中），会将此故障丢弃。

[publicFaultConfiguration.json](#zh-cn_topic_0000002181110120_table8202741102717)为公共故障的系统配置文件，若用户无特殊需求，请勿随意修改。若用户需要修改公共故障的级别和发送方，可以通过在/user1/mindx-dl/clusterd写入自定义配置文件publicCustomization.json实现。该文件路径支持配置，配置方式如下所示。

>[!NOTE]
>
>- 文件publicCustomization.json在容器内路径为/user1/mindx-dl/clusterd，不支持修改，不支持软链接；主机路径默认为/user1/mindx-dl/clusterd。
>- 主机路径可由用户根据实际情况自行配置：在ClusterD的启动YAML中修改挂载卷名称为config-clusterd的主机挂载路径。
>- 多master场景下，建议每个master节点上都同步一份最新的publicCustomization.json文件。避免重启ClusterD后，ClusterD被调度到其他master节点，从而导致自定义故障配置文件丢失的问题。

**表 1**  故障级别及处理说明

<a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_table169151711124319"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row19916131120434"><th class="cellrowborder" valign="top" width="15.09499941718149%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"></a>故障级别</p>
</th>
<th class="cellrowborder" valign="top" width="42.54575125305979%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"></a>故障处理策略</p>
</th>
<th class="cellrowborder" valign="top" width="42.35924932975871%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002181110120_p2218314171716"><a name="zh-cn_topic_0000002181110120_p2218314171716"></a>重调度处理</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row6916711144312"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"></a>无需处理</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"></a>暂不处理</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row1991661104316"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"></a>无法恢复，需要隔离芯片</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p12165431710"><a name="zh-cn_topic_0000002181110120_p12165431710"></a>隔离芯片，进行任务重调度。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row191716112431"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"></a>根据任务YAML中配置的subHealthyStrategy参数取值进行处理，详细请参见<a href="../../../06_api/15_yaml_configuration.md#yaml_configuration">YAML配置说明</a>。</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"></a>当芯片出现亚健康故障时，需根据<a href="../../04_fault_recovery/01_resumable_training/03_configuration/01_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002511426471">配置亚健康热切</a>进行处理。</p>
<div class="note" id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_note7936204710536"></a><div class="notebody"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"></a>如果后续芯片出现其他级别故障，此时SubHealthFault处理策略不影响其他级别的故障处理。</p>
</div></div>
</td>
</tr>
<tr id="row16800523414"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="p88011823817"><a name="p88011823817"></a><span id="ph1339214581915"><a name="ph1339214581915"></a>PreSeparateNPU</span></p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="p980117231413"><a name="p980117231413"></a><span id="ph739245817113"><a name="ph739245817113"></a>暂不影响业务，后续不再调度任务到该芯片。</span></p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="p1280114235116"><a name="p1280114235116"></a><span id="ph3392758212"><a name="ph3392758212"></a>预隔离芯片。</span></p>
</td>
</tr>
</tbody>
</table>

**表 2**  publicFaultConfiguration.json字段说明

<a name="zh-cn_topic_0000002181110120_table8202741102717"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row18202164117272"><th class="cellrowborder" valign="top" width="28.93%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002181110120_p1120213413271"><a name="zh-cn_topic_0000002181110120_p1120213413271"></a>参数名称</p>
</th>
<th class="cellrowborder" valign="top" width="71.07%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002181110120_p22024417279"><a name="zh-cn_topic_0000002181110120_p22024417279"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row172028412278"><td class="cellrowborder" valign="top" width="28.93%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p1220219412279"><a name="zh-cn_topic_0000002181110120_p1220219412279"></a><a href="#zh-cn_topic_0000002181110120_table1689274753416">publicFaultCode</a></p>
</td>
<td class="cellrowborder" valign="top" width="71.07%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p220284110271"><a name="zh-cn_topic_0000002181110120_p220284110271"></a>公共故障码相关配置。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row14606121802219"><td class="cellrowborder" valign="top" width="28.93%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p1760617182224"><a name="zh-cn_topic_0000002181110120_p1760617182224"></a>publicFaultResource</p>
</td>
<td class="cellrowborder" valign="top" width="71.07%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p1606118102218"><a name="zh-cn_topic_0000002181110120_p1606118102218"></a>公共故障发送方配置。</p>
</td>
</tr>
</tbody>
</table>

**表 3**  publicFaultCode字段说明

<a name="zh-cn_topic_0000002181110120_table1689274753416"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row16892144733413"><th class="cellrowborder" valign="top" width="28.849999999999998%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002181110120_p689264723412"><a name="zh-cn_topic_0000002181110120_p689264723412"></a>参数名称</p>
</th>
<th class="cellrowborder" valign="top" width="71.15%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002181110120_p889274783418"><a name="zh-cn_topic_0000002181110120_p889274783418"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row28921647103410"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p48921847143412"><a name="zh-cn_topic_0000002181110120_p48921847143412"></a>NotHandleFaultCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p58921747183416"><a name="zh-cn_topic_0000002181110120_p58921747183416"></a>故障级别为NotHandleFault（无需处理）的故障码。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row989224719346"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p118928476343"><a name="zh-cn_topic_0000002181110120_p118928476343"></a>SubHealthFaultCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p17892947113410"><a name="zh-cn_topic_0000002181110120_p17892947113410"></a>故障级别为SubHealthFault（亚健康）的故障码。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row289264713349"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p38921547193418"><a name="zh-cn_topic_0000002181110120_p38921547193418"></a>SeparateNPUCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p689274714341"><a name="zh-cn_topic_0000002181110120_p689274714341"></a>故障级别为SeparateNPU（无法恢复，需要隔离芯片）的故障码。</p>
</td>
</tr>
<tr id="row107385344217"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="p187397341724"><a name="p187397341724"></a><span id="ph791817016319"><a name="ph791817016319"></a>PreSeparateNPUCodes</span></p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="p15739113415210"><a name="p15739113415210"></a><span id="ph8918120234"><a name="ph8918120234"></a>故障级别为</span><span id="ph491890639"><a name="ph491890639"></a>PreSeparateNPU</span><span id="ph6918601336"><a name="ph6918601336"></a>（暂不影响业务，后续不再调度任务到该芯片）的故障码。</span></p>
</td>
</tr>
</tbody>
</table>

## 故障码说明<a name="zh-cn_topic_0000002181110120_section1440314273418"></a>

公共故障的故障码为9位，说明如下。

**表 4**  故障码说明

<a name="table1237891465117"></a>
<table><thead align="left"><tr id="row1137891413516"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="p1937816143519"><a name="p1937816143519"></a>位数</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="p837812144514"><a name="p837812144514"></a>描述</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="p14378201455110"><a name="p14378201455110"></a>取值</p>
</th>
</tr>
</thead>
<tbody><tr id="row1137861419517"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p123782149514"><a name="p123782149514"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p15378914185120"><a name="p15378914185120"></a>故障类型</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p10378161419517"><a name="p10378161419517"></a>0：芯片故障</p>
<p id="p037871414515"><a name="p037871414515"></a>1：节点故障</p>
<p id="p33781414125113"><a name="p33781414125113"></a>2：网络故障</p>
<p id="p10379101414516"><a name="p10379101414516"></a>3：存储故障</p>
</td>
</tr>
<tr id="row337901415519"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p2379181475111"><a name="p2379181475111"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p133796146513"><a name="p133796146513"></a>故障默认的级别</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p103791114185115"><a name="p103791114185115"></a>0: NotHandleFault</p>
<p id="p193791214175112"><a name="p193791214175112"></a>1: SubHealthFault</p>
<p id="p737991475119"><a name="p737991475119"></a>2: SeparateNPU</p>
</td>
</tr>
<tr id="row1737917147519"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p133793145514"><a name="p133793145514"></a>3、4</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1137901435119"><a name="p1137901435119"></a>预留扩展位</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p43795142516"><a name="p43795142516"></a>暂为00</p>
</td>
</tr>
<tr id="row1337961495114"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17379121416515"><a name="p17379121416515"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p17379141465112"><a name="p17379141465112"></a>第6-9位的故障码是否为用户自定义，避免冲突</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1237917146517"><a name="p1237917146517"></a>0：发布包中定义</p>
<p id="p12379191418513"><a name="p12379191418513"></a>1：用户自定义</p>
</td>
</tr>
<tr id="row1937911425114"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p12379161465115"><a name="p12379161465115"></a>6-9</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1437931425115"><a name="p1437931425115"></a>具体的十进制故障码</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p8379121413512"><a name="p8379121413512"></a>示例：1001</p>
</td>
</tr>
<tr id="row6379214165114"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p1137911410515"><a name="p1137911410515"></a>示例如下：</p>
<p id="p1837941416513"><a name="p1837941416513"></a>0100 01001：芯片故障，SubHealthFault，发布包中定义，故障1001。</p>
<p id="p1037911455117"><a name="p1037911455117"></a>1000 11002：节点故障，NotHandleFault，用户自定义，故障1002。</p>
<p id="p8379181455115"><a name="p8379181455115"></a>2200 01003：网络故障，SeparateNPU，发布包中定义，故障1003。</p>
</td>
</tr>
</tbody>
</table>

## 已支持的公共故障<a name="zh-cn_topic_0000002181110120_section4960201383813"></a>

**表 5**  已支持的公共故障

<a name="zh-cn_topic_0000002181110120_table31451934163811"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row514523493819"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002181110120_p1114523420389"><a name="zh-cn_topic_0000002181110120_p1114523420389"></a>故障码</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002181110120_p9145143412387"><a name="zh-cn_topic_0000002181110120_p9145143412387"></a>故障说明</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002181110120_p15145193413388"><a name="zh-cn_topic_0000002181110120_p15145193413388"></a>默认故障级别</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row1514593415388"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_p181451134193811"><a name="zh-cn_topic_0000002181110120_p181451134193811"></a>010001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_p814593412386"><a name="zh-cn_topic_0000002181110120_p814593412386"></a>光链路脏污（芯片故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p414533483811"><a name="zh-cn_topic_0000002181110120_p414533483811"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row175241157181818"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p580896101918"><a name="p580896101918"></a>210001007</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p15808166121915"><a name="p15808166121915"></a>光链路脏污（网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1180917617197"><a name="p1180917617197"></a>SubHealthFault</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row131782214434"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_p41752216438"><a name="zh-cn_topic_0000002181110120_p41752216438"></a>220001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_p1171822134316"><a name="zh-cn_topic_0000002181110120_p1171822134316"></a>NPU卡<span id="ph17233131243911"><a name="ph17233131243911"></a>HCCS</span>网络故障</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p1566710511444"><a name="zh-cn_topic_0000002181110120_p1566710511444"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row192881812184715"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3289131210473"><a name="p3289131210473"></a>010001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1628951244719"><a name="p1628951244719"></a>光链路松动（芯片故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p828971254715"><a name="p828971254715"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row38601828161910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p6168163671911"><a name="p6168163671911"></a>210001008</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p316816364194"><a name="p316816364194"></a>光链路松动（网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2168436121911"><a name="p2168436121911"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row172051674711"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p127201168472"><a name="p127201168472"></a>310001005</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1572071644717"><a name="p1572071644717"></a>DPC客户端失效</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p141495491488"><a name="p141495491488"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row4720816104713"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17720131674712"><a name="p17720131674712"></a>200001006</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1972020169475"><a name="p1972020169475"></a>疑似光链路亚健康</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1572061684719"><a name="p1572061684719"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row191121122184715"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p7112152234711"><a name="p7112152234711"></a>210001009</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p10112152210476"><a name="p10112152210476"></a>光模块器件亚健康</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2011213229474"><a name="p2011213229474"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row19731102610435"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1324180124413"><a name="p1324180124413"></a>220001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p13241019443"><a name="p13241019443"></a>备份超节点场景下，调度使用不存在的备份框资源。</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p12161558145818"><a name="p12161558145818"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row13731626174317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3241309446"><a name="p3241309446"></a>220001003</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1424190104412"><a name="p1424190104412"></a>备份框资源端口故障</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p17362234428"><a name="p17362234428"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row127318268434"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p52416011444"><a name="p52416011444"></a>220001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p124800442"><a name="p124800442"></a>备份框任务ID占用冲突</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p191615588586"><a name="p191615588586"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row20731826154318"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5241103443"><a name="p5241103443"></a>220001005</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p824705444"><a name="p824705444"></a>NetMind失效</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1016110586589"><a name="p1016110586589"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row673142624317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p162412019440"><a name="p162412019440"></a>220001006</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p10241109444"><a name="p10241109444"></a>疑似备份框链路端口部分失效</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p916119583580"><a name="p916119583580"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row673116264438"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p9247064419"><a name="p9247064419"></a>220001007</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p6249018447"><a name="p6249018447"></a>光链路调整失败</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p132215501211"><a name="p132215501211"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row8926105693315"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1792695643311"><a name="p1792695643311"></a>200001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p18926165614336"><a name="p18926165614336"></a>某节点内产生/恢复慢网络（慢网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p24091634153411"><a name="p24091634153411"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row10526205273417"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3526052153416"><a name="p3526052153416"></a>200001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p5804154074418"><a name="p5804154074418"></a>超节点内的节点间产生/恢复慢网络。（慢网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p352695212349"><a name="p352695212349"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row663164316353"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17634437355"><a name="p17634437355"></a>200001012</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p18631743123513"><a name="p18631743123513"></a>不是卡故障导致的慢网络（慢网络故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p101021310163612"><a name="p101021310163612"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row178327182364"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p383221833611"><a name="p383221833611"></a>110001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1683231816361"><a name="p1683231816361"></a>慢节点故障</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p6832131833614"><a name="p6832131833614"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row1179514189380"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p979511810389"><a name="p979511810389"></a>100001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1579521818381"><a name="p1579521818381"></a>劣化已恢复（慢节点故障）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p27883220394"><a name="p27883220394"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row121732048142813"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p4359165915289"><a name="p4359165915289"></a>110001020</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1817494822816"><a name="p1817494822816"></a>共享存储DPC进程异常</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p121741348132817"><a name="p121741348132817"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row7277115416280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5277135492816"><a name="p5277135492816"></a>110001021</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7277854132820"><a name="p7277854132820"></a>共享存储DPC内存不足异常</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182781654132815"><a name="p182781654132815"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row7277125616280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5277123492816"><a name="p5277123492816"></a>120001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7277855532820"><a name="p7277855532820"></a>文件系统容量不足</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182123454132815"><a name="p182123454132815"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row7277125966280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5277258692816"><a name="p5277258692816"></a>010001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7277819132820"><a name="p7277819132820"></a>HBM多bit ECC错误（重启）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723412132815"><a name="p182723412132815"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row1233115416280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5837135492816"><a name="p5837135492816"></a>020001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7272364132820"><a name="p7272364132820"></a>HBM多bit ECC错误（隔离）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723454132815"><a name="p182723454132815"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row1233115416281"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5837135492817"><a name="p5837135492817"></a>020001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7272364132821"><a name="p7272364132821"></a>检测周期内任意一次HBM多bit ECC严重错误（次数可调整）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723454132816"><a name="p182723454132816"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row1233115416282"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5837135492818"><a name="p5837135492818"></a>020001003</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7272364132822"><a name="p7272364132822"></a>检测周期内任意两次NPU芯片健康状态次要告警（次数可调整）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723454132817"><a name="p182723454132817"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row1233115416283"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5837135492819"><a name="p5837135492819"></a>120001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7272364132823"><a name="p7272364132823"></a>检测周期内任意一次CPU芯片健康状态严重告警</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723454132818"><a name="p182723454132818"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row1233115416284"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5837135492820"><a name="p5837135492820"></a>120001003</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7272364132824"><a name="p7272364132824"></a>检测周期内任意一个CPU芯片健康状态次要告警持续10秒（时长可调整）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723454132819"><a name="p182723454132819"></a>SeparateNPU</p>
</td>
</tr>
<tr ><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p ><a name="p5277135492816-duplicate-2"></a>110001022</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p><a name="p7277854132820-duplicate-2"></a>共享存储DTFS进程异常</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p><a name="p182781654132815-duplicate-2"></a>PreSeparateNPU</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p><a name="p5277135492816-duplicate-3"></a>110001023</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p><a name="p7277854132820-duplicate-3"></a>共享存储DTFS链接异常</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p><a name="p182781654132815-duplicate-3"></a>PreSeparateNPU</p>
</td>
</tr>
<tr id="row_hangfault001"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p_hangfault001_code"><a name="p_hangfault001_code"></a>200001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p_hangfault001_desc"><a name="p_hangfault001_desc"></a>任务卡死故障</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p_hangfault001_level"><a name="p_hangfault001_level"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row_hangfault001"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p_hangfault001_code"><a name="p_hangfault001_code-duplicate-2"></a>020000002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p_hangfault001_desc"><a name="p_hangfault001_desc-duplicate-2"></a>超平面光链路成员端口故障（路由可收敛）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p_hangfault001_level"><a name="p_hangfault001_level-duplicate-2"></a>SubHealthFaultCodes</p>
</td>
</tr>
<tr id="row_hangfault001"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p_hangfault001_code"><a name="p_hangfault001_code-duplicate-3"></a>020001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p_hangfault001_desc"><a name="p_hangfault001_desc-duplicate-3"></a>超平面光链路成员端口故障（路由不可收敛）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p_hangfault001_level"><a name="p_hangfault001_level-duplicate-3"></a>SeparateNPUCodes</p>
</td>
</tr>
<tr id="row_hangfault001"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p_hangfault001_code"><a name="p_hangfault001_code-duplicate-4"></a>110001024</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p_hangfault001_desc"><a name="p_hangfault001_desc-duplicate-4"></a>参数面光链路成员端口故障（UBOE故障，路由不可收敛）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p_hangfault001_level"><a name="p_hangfault001_level-duplicate-4"></a>PreSeparateNPUCodes</p>
</td>
</tr>
<tr id="row_hangfault001"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p_hangfault001_code"><a name="p_hangfault001_code-duplicate-5"></a>110000002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p_hangfault001_desc"><a name="p_hangfault001_desc-duplicate-5"></a>参数面光链路成员端口故障（UBOE故障，路由可收敛）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p_hangfault001_level"><a name="p_hangfault001_level-duplicate-5"></a>SubHealthFaultCodes</p>
</td>
</tr>
</tbody>
</table>

## （可选）配置公共故障的级别和发送方<a name="ZH-CN_TOPIC_0000002479226494"></a>

在制作ClusterD镜像时，会将故障级别配置文件publicFaultConfiguration.json内置在镜像中，启动ClusterD时会读取这个文件的默认配置，作为当前故障处理依据。

如果用户想要自定义故障级别，可以在主机上创建/user1/mindx-dl/clusterd/publicCustomization.json文件。

- 如果ClusterD启动时，已经存在该文件，ClusterD会优先按照已存在的文件中配置的内容，作为当前故障处理依据。
- 如果重新安装ClusterD后，已经存在该文件，ClusterD的默认publicFaultConfiguration.json将不会生效，使用已经存在的publicCustomization.json文件。若想要使用publicFaultConfiguration.json的默认配置，可以删除已存在的publicCustomization.json文件，使ClusterD读取默认的publicFaultConfiguration.json文件。
- 如果publicCustomization.json文件内容存在格式错误等问题，ClusterD会默认读取镜像中内置的publicFaultConfiguration.json文件的内容，作为当前故障处理依据。

### 配置公共故障码的故障级别<a name="zh-cn_topic_0000002180950420_section1384121854711"></a>

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
        "CCAE", "fd-online", "pingmesh", "Netmind", "dpcStorage", "dtfsStorage"
      ]
    }
    ```

3. 修改完成后，按“Esc”键，输入:wq!保存并退出。
4. 几秒钟后，文件生效。查看操作是否成功。

    若日志出现“load fault config from <publicCustomization.json> success”，表示手动配置故障码操作成功。

### 配置公共故障的发送方<a name="zh-cn_topic_0000002180950420_section5532327614"></a>

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
        "CCAE", "fd-online", "pingmesh", "Netmind", "dpcStorage", "dtfsStorage", "XXX"
      ]
    }
    ```

3. 修改完成后，按“Esc”键，输入:wq!保存并退出。
4. 几秒钟后，文件生效。查看操作是否成功。

    若日志出现“load fault config from <publicCustomization.json> success”，表示手动配置故障码操作成功。
