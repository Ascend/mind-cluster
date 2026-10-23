# Configuring Public Faults<a name="ZH-CN_TOPIC_0000002479386564"></a>

## Configuration File Description<a name="ZH-CN_TOPIC_0000002511346487"></a>

The fault detection feature handles public faults by level. ClusterD obtains the fault code of the current fault and handles the fault accordingly based on the fault level configured for the fault code in the `publicFaultConfiguration.json` file. In special cases, if ClusterD receives an unrecognized fault code (one that is not saved in the configuration file), the fault is discarded.

[publicFaultConfiguration.json](#zh-cn_topic_0000002181110120_table8202741102717) is the system configuration file of public faults. Do not modify it unless necessary. If you need to change the level and sender of a public fault, you can write `publicCustomization.json` in `/user1/mindx-dl/clusterd`. The path of this file can be configured as follows.

>[!NOTE]
>
>- The path of the `publicCustomization.json` file in the container is `/user1/mindx-dl/clusterd`. It cannot be modified, and symbolic links are not supported. The default host path is `/user1/mindx-dl/clusterd`.
>- The host path can be configured by users based on the actual situation: modify the host mount path of the volume named `config-clusterd` in the ClusterD startup YAML.
>- In multi-master scenarios, you are advised to synchronize the latest publicCustomization.json file to each master node. This prevents the custom fault configuration file from being lost when ClusterD is scheduled to another master node after a restart.

**Table 1** Fault level and handling description

<a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_table169151711124319"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row19916131120434"><th class="cellrowborder" valign="top" width="15.09499941718149%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"></a>Fault Level</p>
</th>
<th class="cellrowborder" valign="top" width="42.54575125305979%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"></a>Fault Handling Policy</p>
</th>
<th class="cellrowborder" valign="top" width="42.35924932975871%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002181110120_p2218314171716"><a name="zh-cn_topic_0000002181110120_p2218314171716"></a>Rescheduling Handling</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row6916711144312"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"></a>No handling required</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"></a>Not handled for now</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row1991661104316"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"></a>Unrecoverable; the chip needs to be isolated</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p12165431710"><a name="zh-cn_topic_0000002181110120_p12165431710"></a>Isolate the chip and perform job rescheduling.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row191716112431"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"></a>Handled according to the value of the subHealthyStrategy parameter configured in the job YAML. For details, see <a href="../../../06_api/15_yaml_configuration.md#yaml_configuration">Job YAML Configuration</a>.</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"></a>When a sub-healthy fault occurs on a chip, handle it according to <a href="../../05_fault_recovery/01_resumable_training/02_configuration/01_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002511426471">Configuring Sub-Health Hot Switchover</a>.</p>
<div class="note" id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_note7936204710536"></a><div class="notebody"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"></a>If the chip subsequently has faults of other levels, the SubHealthFault handling policy does not affect the handling of faults of other levels.</p>
</div></div>
</td>
</tr>
<tr id="row16800523414"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="p88011823817"><a name="p88011823817"></a><span id="ph1339214581915"><a name="ph1339214581915"></a>PreSeparateNPU</span></p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="p980117231413"><a name="p980117231413"></a><span id="ph739245817113"><a name="ph739245817113"></a>Does not affect services for now; jobs are no longer scheduled to this chip.</span></p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="p1280114235116"><a name="p1280114235116"></a><span id="ph3392758212"><a name="ph3392758212"></a>Pre-isolate the chip.</span></p>
</td>
</tr>
</tbody>
</table>

**Table 2** publicFaultConfiguration.json field description

<a name="zh-cn_topic_0000002181110120_table8202741102717"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row18202164117272"><th class="cellrowborder" valign="top" width="28.93%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002181110120_p1120213413271"><a name="zh-cn_topic_0000002181110120_p1120213413271"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="71.07%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002181110120_p22024417279"><a name="zh-cn_topic_0000002181110120_p22024417279"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row172028412278"><td class="cellrowborder" valign="top" width="28.93%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p1220219412279"><a name="zh-cn_topic_0000002181110120_p1220219412279"></a><a href="#zh-cn_topic_0000002181110120_table1689274753416">publicFaultCode</a></p>
</td>
<td class="cellrowborder" valign="top" width="71.07%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p220284110271"><a name="zh-cn_topic_0000002181110120_p220284110271"></a>Configuration related to public fault codes.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row14606121802219"><td class="cellrowborder" valign="top" width="28.93%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p1760617182224"><a name="zh-cn_topic_0000002181110120_p1760617182224"></a>publicFaultResource</p>
</td>
<td class="cellrowborder" valign="top" width="71.07%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p1606118102218"><a name="zh-cn_topic_0000002181110120_p1606118102218"></a>Configuration of public fault senders.</p>
</td>
</tr>
</tbody>
</table>

**Table 3** publicFaultCode field description

<a name="zh-cn_topic_0000002181110120_table1689274753416"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row16892144733413"><th class="cellrowborder" valign="top" width="28.849999999999998%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002181110120_p689264723412"><a name="zh-cn_topic_0000002181110120_p689264723412"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="71.15%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002181110120_p889274783418"><a name="zh-cn_topic_0000002181110120_p889274783418"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row28921647103410"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p48921847143412"><a name="zh-cn_topic_0000002181110120_p48921847143412"></a>NotHandleFaultCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p58921747183416"><a name="zh-cn_topic_0000002181110120_p58921747183416"></a>Fault codes whose fault level is NotHandleFault (no handling required).</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row989224719346"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p118928476343"><a name="zh-cn_topic_0000002181110120_p118928476343"></a>SubHealthFaultCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p17892947113410"><a name="zh-cn_topic_0000002181110120_p17892947113410"></a>Fault codes whose fault level is SubHealthFault (sub-healthy).</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row289264713349"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p38921547193418"><a name="zh-cn_topic_0000002181110120_p38921547193418"></a>SeparateNPUCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p689274714341"><a name="zh-cn_topic_0000002181110120_p689274714341"></a>Fault codes whose fault level is SeparateNPU (unrecoverable; the chip needs to be isolated).</p>
</td>
</tr>
<tr id="row107385344217"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="p187397341724"><a name="p187397341724"></a><span id="ph791817016319"><a name="ph791817016319"></a>PreSeparateNPUCodes</span></p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="p15739113415210"><a name="p15739113415210"></a><span id="ph8918120234"><a name="ph8918120234"></a>Fault codes whose fault level is </span><span id="ph491890639"><a name="ph491890639"></a>PreSeparateNPU</span><span id="ph6918601336"><a name="ph6918601336"></a>(does not affect services for now; jobs are no longer scheduled to this chip).</span></p>
</td>
</tr>
</tbody>
</table>

## Fault Code Description<a name="zh-cn_topic_0000002181110120_section1440314273418"></a>

A public fault code has nine digits, as described below.

**Table 4** Fault code description

<a name="table1237891465117"></a>
<table><thead align="left"><tr id="row1137891413516"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="p1937816143519"><a name="p1937816143519"></a>Digit</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="p837812144514"><a name="p837812144514"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="p14378201455110"><a name="p14378201455110"></a>Value</p>
</th>
</tr>
</thead>
<tbody><tr id="row1137861419517"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p123782149514"><a name="p123782149514"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p15378914185120"><a name="p15378914185120"></a>Fault type</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p10378161419517"><a name="p10378161419517"></a>0: Chip fault</p>
<p id="p037871414515"><a name="p037871414515"></a>1: Node fault</p>
<p id="p33781414125113"><a name="p33781414125113"></a>2: Network fault</p>
<p id="p10379101414516"><a name="p10379101414516"></a>3: Storage fault</p>
</td>
</tr>
<tr id="row337901415519"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p2379181475111"><a name="p2379181475111"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p133796146513"><a name="p133796146513"></a>Default fault level</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p103791114185115"><a name="p103791114185115"></a>0: NotHandleFault</p>
<p id="p193791214175112"><a name="p193791214175112"></a>1: SubHealthFault</p>
<p id="p737991475119"><a name="p737991475119"></a>2: SeparateNPU</p>
</td>
</tr>
<tr id="row1737917147519"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p133793145514"><a name="p133793145514"></a>3, 4</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1137901435119"><a name="p1137901435119"></a>Reserved extension bit</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p43795142516"><a name="p43795142516"></a>Currently 00</p>
</td>
</tr>
<tr id="row1337961495114"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17379121416515"><a name="p17379121416515"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p17379141465112"><a name="p17379141465112"></a>Whether digits 6 to 9 of the fault code are user-defined, to avoid conflicts</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1237917146517"><a name="p1237917146517"></a>0: Defined in the release package</p>
<p id="p12379191418513"><a name="p12379191418513"></a>1: User-defined</p>
</td>
</tr>
<tr id="row1937911425114"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p12379161465115"><a name="p12379161465115"></a>6-9</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1437931425115"><a name="p1437931425115"></a>Specific decimal fault code</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p8379121413512"><a name="p8379121413512"></a>Example: 1001</p>
</td>
</tr>
<tr id="row6379214165114"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p1137911410515"><a name="p1137911410515"></a>Examples:</p>
<p id="p1837941416513"><a name="p1837941416513"></a>0100 01001: Chip fault, SubHealthFault, defined in the release package, fault 1001.</p>
<p id="p1037911455117"><a name="p1037911455117"></a>1000 11002: Node fault, NotHandleFault, user-defined, fault 1002.</p>
<p id="p8379181455115"><a name="p8379181455115"></a>2200 01003: Network fault, SeparateNPU, defined in the release package, fault 1003.</p>
</td>
</tr>
</tbody>
</table>

## Supported Public Faults<a name="zh-cn_topic_0000002181110120_section4960201383813"></a>

**Table 5** Supported public faults

<a name="zh-cn_topic_0000002181110120_table31451934163811"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row514523493819"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002181110120_p1114523420389"><a name="zh-cn_topic_0000002181110120_p1114523420389"></a>Fault Code</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002181110120_p9145143412387"><a name="zh-cn_topic_0000002181110120_p9145143412387"></a>Fault Description</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002181110120_p15145193413388"><a name="zh-cn_topic_0000002181110120_p15145193413388"></a>Default Fault Level</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row1514593415388"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_p181451134193811"><a name="zh-cn_topic_0000002181110120_p181451134193811"></a>010001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_p814593412386"><a name="zh-cn_topic_0000002181110120_p814593412386"></a>Dirty optical link (chip fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p414533483811"><a name="zh-cn_topic_0000002181110120_p414533483811"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row175241157181818"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p580896101918"><a name="p580896101918"></a>210001007</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p15808166121915"><a name="p15808166121915"></a>Dirty optical link (network fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1180917617197"><a name="p1180917617197"></a>SubHealthFault</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row131782214434"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_p41752216438"><a name="zh-cn_topic_0000002181110120_p41752216438"></a>220001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_p1171822134316"><a name="zh-cn_topic_0000002181110120_p1171822134316"></a>NPU card <span id="ph17233131243911"><a name="ph17233131243911"></a>HCCS</span> network fault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p1566710511444"><a name="zh-cn_topic_0000002181110120_p1566710511444"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row192881812184715"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3289131210473"><a name="p3289131210473"></a>010001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1628951244719"><a name="p1628951244719"></a>Loose optical link (chip fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p828971254715"><a name="p828971254715"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row38601828161910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p6168163671911"><a name="p6168163671911"></a>210001008</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p316816364194"><a name="p316816364194"></a>Loose optical link (network fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2168436121911"><a name="p2168436121911"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row172051674711"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p127201168472"><a name="p127201168472"></a>310001005</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1572071644717"><a name="p1572071644717"></a>DPC client failure</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p141495491488"><a name="p141495491488"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row4720816104713"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17720131674712"><a name="p17720131674712"></a>200001006</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1972020169475"><a name="p1972020169475"></a>Suspected sub-healthy optical link</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1572061684719"><a name="p1572061684719"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row191121122184715"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p7112152234711"><a name="p7112152234711"></a>210001009</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p10112152210476"><a name="p10112152210476"></a>Sub-healthy optical module component</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2011213229474"><a name="p2011213229474"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row19731102610435"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1324180124413"><a name="p1324180124413"></a>220001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p13241019443"><a name="p13241019443"></a>In a backup super node scenario, a non-existent backup frame resource is used for scheduling.</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p12161558145818"><a name="p12161558145818"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row13731626174317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3241309446"><a name="p3241309446"></a>220001003</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1424190104412"><a name="p1424190104412"></a>Backup frame resource port fault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p17362234428"><a name="p17362234428"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row127318268434"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p52416011444"><a name="p52416011444"></a>220001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p124800442"><a name="p124800442"></a>Backup frame job ID occupation conflict</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p191615588586"><a name="p191615588586"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row20731826154318"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5241103443"><a name="p5241103443"></a>220001005</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p824705444"><a name="p824705444"></a>NetMind failure</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1016110586589"><a name="p1016110586589"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row673142624317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p162412019440"><a name="p162412019440"></a>220001006</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p10241109444"><a name="p10241109444"></a>Suspected partial failure of backup frame link ports</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p916119583580"><a name="p916119583580"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row673116264438"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p9247064419"><a name="p9247064419"></a>220001007</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p6249018447"><a name="p6249018447"></a>Optical link adjustment failure</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p132215501211"><a name="p132215501211"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row8926105693315"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1792695643311"><a name="p1792695643311"></a>200001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p18926165614336"><a name="p18926165614336"></a>Slow network generated/recovered within a node (slow network fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p24091634153411"><a name="p24091634153411"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row10526205273417"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3526052153416"><a name="p3526052153416"></a>200001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p5804154074418"><a name="p5804154074418"></a>Slow network generated/recovered between nodes in a super node (slow network fault).</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p352695212349"><a name="p352695212349"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row663164316353"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17634437355"><a name="p17634437355"></a>200001012</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p18631743123513"><a name="p18631743123513"></a>Slow network not caused by a card fault (slow network fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p101021310163612"><a name="p101021310163612"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row178327182364"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p383221833611"><a name="p383221833611"></a>110001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1683231816361"><a name="p1683231816361"></a>Slow node fault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p6832131833614"><a name="p6832131833614"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row1179514189380"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p979511810389"><a name="p979511810389"></a>100001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1579521818381"><a name="p1579521818381"></a>Degradation recovered (slow node fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p27883220394"><a name="p27883220394"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row121732048142813"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p4359165915289"><a name="p4359165915289"></a>110001020</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1817494822816"><a name="p1817494822816"></a>Shared storage DPC process exception</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p121741348132817"><a name="p121741348132817"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row7277115416280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5277135492816"><a name="p5277135492816"></a>110001021</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7277854132820"><a name="p7277854132820"></a>Shared storage DPC out-of-memory exception</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182781654132815"><a name="p182781654132815"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row7277125616280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5277123492816"><a name="p5277123492816"></a>120001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7277855532820"><a name="p7277855532820"></a>Insufficient file system capacity</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182123454132815"><a name="p182123454132815"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row7277125966280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5277258692816"><a name="p5277258692816"></a>010001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7277819132820"><a name="p7277819132820"></a>HBM multi-bit ECC error (restart)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723412132815"><a name="p182723412132815"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row1233115416280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5837135492816"><a name="p5837135492816"></a>020001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7272364132820"><a name="p7272364132820"></a>HBM multi-bit ECC error (isolation)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723454132815"><a name="p182723454132815"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row1233115416281"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5837135492817"><a name="p5837135492817"></a>020001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7272364132821"><a name="p7272364132821"></a>Any one severe HBM multi-bit ECC error within the detection period (the count is adjustable)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723454132816"><a name="p182723454132816"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row1233115416282"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5837135492818"><a name="p5837135492818"></a>020001003</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7272364132822"><a name="p7272364132822"></a>Any two minor alarms of NPU chip health status within the detection period (the count is adjustable)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723454132817"><a name="p182723454132817"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row1233115416283"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5837135492819"><a name="p5837135492819"></a>120001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7272364132823"><a name="p7272364132823"></a>Any one severe alarm of CPU chip health status within the detection period</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723454132818"><a name="p182723454132818"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row1233115416284"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5837135492820"><a name="p5837135492820"></a>120001003</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7272364132824"><a name="p7272364132824"></a>A minor alarm of CPU chip health status lasting 10 seconds within the detection period (the duration is adjustable)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182723454132819"><a name="p182723454132819"></a>SeparateNPU</p>
</td>
</tr>
<tr ><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p ><a name="p5277135492816-duplicate-2"></a>110001022</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p><a name="p7277854132820-duplicate-2"></a>Shared storage DTFS process exception</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p><a name="p182781654132815-duplicate-2"></a>PreSeparateNPU</p>
</td>
</tr>
<tr><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p><a name="p5277135492816-duplicate-3"></a>110001023</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p><a name="p7277854132820-duplicate-3"></a>Shared storage DTFS link exception</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p><a name="p182781654132815-duplicate-3"></a>PreSeparateNPU</p>
</td>
</tr>
<tr id="row_hangfault001"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p_hangfault001_code"><a name="p_hangfault001_code"></a>200001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p_hangfault001_desc"><a name="p_hangfault001_desc"></a>Job hang fault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p_hangfault001_level"><a name="p_hangfault001_level"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row_hangfault001"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p_hangfault001_code"><a name="p_hangfault001_code-duplicate-2"></a>020000002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p_hangfault001_desc"><a name="p_hangfault001_desc-duplicate-2"></a>Hyper-plane optical link member port fault (route can converge)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p_hangfault001_level"><a name="p_hangfault001_level-duplicate-2"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row_hangfault001"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p_hangfault001_code"><a name="p_hangfault001_code-duplicate-3"></a>020001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p_hangfault001_desc"><a name="p_hangfault001_desc-duplicate-3"></a>Hyper-plane optical link member port fault (route cannot converge)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p_hangfault001_level"><a name="p_hangfault001_level-duplicate-3"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row_hangfault001"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p_hangfault001_code"><a name="p_hangfault001_code-duplicate-4"></a>110001024</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p_hangfault001_desc"><a name="p_hangfault001_desc-duplicate-4"></a>Parameter plane optical link member port fault (UBoE fault, route cannot converge)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p_hangfault001_level"><a name="p_hangfault001_level-duplicate-4"></a>PreSeparateNPU</p>
</td>
</tr>
<tr id="row_hangfault001"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p_hangfault001_code"><a name="p_hangfault001_code-duplicate-5"></a>110000002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p_hangfault001_desc"><a name="p_hangfault001_desc-duplicate-5"></a>Parameter plane optical link member port fault (UBoE fault, route can converge)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p_hangfault001_level"><a name="p_hangfault001_level-duplicate-5"></a>SubHealthFault</p>
</td>
</tr>
</tbody>
</table>

## (Optional) Configuring the Level and Sender of Public Faults<a name="ZH-CN_TOPIC_0000002479226494"></a>

When the ClusterD image is built, the fault level configuration file `publicFaultConfiguration.json` is built into the image. When ClusterD is started, the default configuration of this file is read as the basis for current fault handling.

If you want to customize fault levels, you can create the `/user1/mindx-dl/clusterd/publicCustomization.json` file on the host.

- If this file already exists when ClusterD is started, ClusterD preferentially uses the content configured in the existing file as the basis for current fault handling.
- If this file already exists after ClusterD is reinstalled, the default publicFaultConfiguration.json of ClusterD does not take effect, and the existing `publicCustomization.json` file is used. If you want to use the default configuration of `publicFaultConfiguration.json`, delete the existing `publicCustomization.json` file so that ClusterD reads the default `publicFaultConfiguration.json` file.
- If the content of the publicCustomization.json file has problems such as format errors, ClusterD reads the content of the `publicFaultConfiguration.json` file built into the image by default as the basis for current fault handling.

### Configuring the Fault Level of a Public Fault Code<a name="zh-cn_topic_0000002180950420_section1384121854711"></a>

The fault level of a public fault code can be configured in the following two scenarios.

- Adjusting the fault level of an existing fault code.
- Adding a fault code and its fault level.

    The following uses fault code `010001008` as an example to describe how to configure the level of a public fault code.

1. Log in to the environment and go to the `/user1/mindx-dl/clusterd` directory.
2. Run the `vi publicCustomization.json` command to edit the file. For details about `publicCustomization.json`, see [Table 2](#ZH-CN_TOPIC_0000002511346487).

    >[!NOTE]
    >- After creating the `publicCustomization.json` file, ensure that the file is readable by the ClusterD user `hwMindX`. For example, if the current user is root, you are advised to set the file permission to 644.
    >- The user must ensure file permission security. If the permission is too broad, security risks may exist.

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

3. After the modification, press `Esc`, type `:wq!`, and press `Enter` to save and exit.
4. The file takes effect after a few seconds. Check whether the operation is successful.

    If "load fault config from <publicCustomization.json> success" appears in the log, the manual fault code configuration is successful.

### Configuring the Sender of a Public Fault<a name="zh-cn_topic_0000002180950420_section5532327614"></a>

The following uses adding the fault sender XXX as an example to describe how to configure the sender of a public fault code.

1. Log in to the environment and go to the `/user1/mindx-dl/clusterd` directory.
2. Run the `vi publicCustomization.json`* command to edit the file. For details about publicCustomization.json, see [Table 2](#ZH-CN_TOPIC_0000002511346487).

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

3. After the modification, press `Esc`, type `:wq!`, and press `Enter` to save and exit.
4. The file takes effect after a few seconds. Check whether the operation is successful.

    If "load fault config from <publicCustomization.json> success" appears in the log, the manual fault code configuration is successful.
