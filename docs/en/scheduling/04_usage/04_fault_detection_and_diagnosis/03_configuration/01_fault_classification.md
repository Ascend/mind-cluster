# Configuring Fault Levels

## Configuration Description<a name="ZH-CN_TOPIC_0000002479386448"></a>

For different fault codes of **node hardware faults**, **chip faults and UnifiedBus device faults**, and **public faults**, the fault detection feature provides default fault levels and fault handling policies for the corresponding levels. For **chip faults**, it also provides the default fault frequency and duration, as well as the corresponding fault handling policy.

If you need to modify a fault handling policy, see this chapter. Do not modify it at will unless you have special requirements.

**Supported Fault Levels<a name="section257513292065"></a>**

The configurable fault levels supported for different types of faults are listed in the following table.

**Table 1** Configurable fault levels

<a name="table4710459145316"></a>
<table><thead align="left"><tr id="row37104590534"><th class="cellrowborder" valign="top" id="mcps1.2.5.1.1"><p id="p7710135925316"><a name="p7710135925316"></a>Fault</p>
</th>
<th class="cellrowborder" colspan="3" valign="top" id="mcps1.2.5.1.2"><p id="p11175192213564"><a name="p11175192213564"></a>Configurable Fault Levels</p>
</th>
</tr>
</thead>
<tbody><tr id="row271045905320"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p271015916536"><a name="p271015916536"></a>Node faults</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p66711187562"><a name="p66711187562"></a>NotHandleFault, PreSeparateFault, SeparateFault</p>
</td>
</tr>
<tr id="row3710165935311"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17710125955315"><a name="p17710125955315"></a>Chip faults</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p21371428713"><a name="p21371428713"></a>NotHandleFault, RestartRequest, RestartBusiness, FreeRestartNPU, RestartNPU, SeparateNPU, PreSeparateNPU, SubHealthFault</p>
</td>
</tr>
<tr id="row5710125913537"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p10710959185319"><a name="p10710959185319"></a>UnifiedBus device faults</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p6631112135616"><a name="p6631112135616"></a>NotHandleFault, SubHealthFault, ResetFault, SeparateFault<span id="ph51441721217"><a name="ph51441721217"></a>, </span><span id="ph375517710129"><a name="ph375517710129"></a>RestartRequestFault</span></p>
</td>
</tr>
<tr id="row416145918513"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p116115913517"><a name="p116115913517"></a>Public faults</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p147536536717"><a name="p147536536717"></a>NotHandleFault, SeparateNPU, SubHealthFault<span id="ph632635517598"><a name="ph632635517598"></a>, PreSeparateNPU</span></p>
</td>
</tr>
<tr id="row416145918513"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p116115913528"><a name="p116115913528"></a>Task hang faults</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p147536536728"><a name="p147536536728"></a>NotHandleFault, SeparateNPU, PreSeparateNPU</p>
</td>
</tr>
<tr id="row416145918513"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p116115913528"><a name="p116115913528-duplicate-2"></a>Optical link member port faults</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p147536536728"><a name="p147536536728-duplicate-2"></a>SeparateNPU, PreSeparateNPU, SubHealthFault</p>
</td>
</tr>
</tbody>
</table>

In the preceding table, the handling policy of each fault level is described as follows.

**Table 2** Fault levels and handling description

<a name="table103716651410"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row461812518228"><th class="cellrowborder" valign="top" width="19.06%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"></a>Fault Handling Policy</p>
</th>
<th class="cellrowborder" valign="top" width="35.74%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="23.39%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"></a>Rescheduling Handling</p>
</th>
<th class="cellrowborder" valign="top" width="21.81%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"></a>Graceful Fault Tolerance Handling</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row961811511228"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"></a>A fault that does not affect services and does not need to be handled.</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"></a>Not handled for now</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"></a>Not handled for now</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row116184515226"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"></a>RestartRequest</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"></a>Affect service execution and requires re-execution of the service request.</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"></a>Isolate the chip and perform job rescheduling.</p>
<div class="note" id="note11901123612819"><a name="note11901123612819"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="p1069261722310"><a name="p1069261722310"></a>If an inference job has subscribed to fault information, and a RestartRequest fault occurs on the inference card used by the job and the fault duration does not exceed 60 seconds, job rescheduling is not performed. If the fault persists for more than 60 seconds, the chip is isolated and job rescheduling is performed.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"></a>In the inference scenario, re-execute the inference request. In the training scenario, re-execute the training service.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row14618105116225"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"></a>RestartBusiness</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"></a>Affect service execution and requires re-execution of the service.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"></a>Re-execute the service</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row561825132214"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"></a>FreeRestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"></a>Affect service execution. The chip needs to be reset when it is idle.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"></a>Wait until the chip is idle and then reset the chip.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row14618125152210"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"></a>RestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"></a>Affect service execution and requires immediate reset of the chip.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"></a>Stop the training service immediately, reset the chip, and then re-execute the service.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row1061895115227"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"></a>The fault cannot be recovered and the chip needs to be isolated.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"></a>Isolate the chip and perform job rescheduling.</p>
</td>
</tr>
<tr id="row870814247412"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p5708202454117"><a name="p5708202454117"></a>SeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p12708162474117"><a name="p12708162474117"></a>The job is definitely affected.</p>
<div class="note" id="note1521013164613"><a name="note1521013164613"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="p92101114465"><a name="p92101114465"></a>When the fault level of a UnifiedBus device fault is SeparateFault, the service fails to run, and the component or board card needs to be replaced.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p0708624204112"><a name="p0708624204112"></a>Job rescheduling</p>
<div class="note" id="note44451347164716"><a name="note44451347164716"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="p64453471479"><a name="p64453471479"></a>For a UnifiedBus device fault, the fault handling policy represented by this fault level is to stop the current training job, isolate the node, and perform job rescheduling.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p137081824174117"><a name="p137081824174117"></a>-</p>
</td>
</tr>
<tr id="row5706333131216"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p177061833201220"><a name="p177061833201220"></a><span id="ph141513510124"><a name="ph141513510124"></a>RestartRequestFault</span></p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p070623351220"><a name="p070623351220"></a><span id="ph18501459184"><a name="ph18501459184"></a>The service fails to run and the service request needs to be re-executed.</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p770653313124"><a name="p770653313124"></a><span id="ph38912127169"><a name="ph38912127169"></a>Stop the current training job, isolate the node, and perform job rescheduling.</span></p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p6706113331213"><a name="p6706113331213"></a>In the inference scenario, re-execute the inference request. In the training scenario, re-execute the training service.</p>
</td>
</tr>
<tr id="row3938182254418"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p39381822174417"><a name="p39381822174417"></a>ResetFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p1193862274418"><a name="p1193862274418"></a>The service fails to run</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p184323519501"><a name="p184323519501"></a>Stop the current training job, isolate the node, and perform job rescheduling.</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p18938822204411"><a name="p18938822204411"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row102215292529"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"></a>PreSeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"></a>Services are not affected for the time being, and jobs are no longer scheduled to this chip subsequently.</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"></a>Pre-isolate the chip</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"></a>Pre-isolate the chip</p>
</td>
</tr>
<tr id="row84541721401"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p174559214016"><a name="p174559214016"></a>PreSeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p145562114011"><a name="p145562114011"></a>May cause the job to be affected.</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p54556214409"><a name="p54556214409"></a>If the node has jobs, do not handle. Subsequently, do not schedule jobs to the node.</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p1245572144015"><a name="p1245572144015"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row0352224175218"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"></a>Handled based on the value of the subHealthyStrategy parameter configured in the job YAML. For details, see <a href="../../../06_api/15_yaml_configuration.md#yaml_configuration">YAML Configuration Description</a>.</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"></a>When a sub-health fault occurs on a chip, handle it according to <a href="../../05_fault_recovery/01_resumable_training/02_configuration/01_configuring_fault_handling_policies.md#ZH-CN_TOPIC_0000002511426471">Configuring Sub-Health Hot Switchover</a>.</p>
<div class="note" id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"></a>If another level of fault occurs on the chip subsequently, the SubHealthFault handling policy does not affect the handling of faults of other levels.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"></a>Not handled for now</p>
</td>
</tr>
</tbody>
</table>
