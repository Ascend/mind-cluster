# (Optional) Configuring Fault Detection Levels<a name="ZH-CN_TOPIC_0000002479386556"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T08:01:36.375Z pushedAt=2026-06-09T09:02:55.502Z -->

## Configuration Description<a name="ZH-CN_TOPIC_0000002479386448"></a>

For different fault codes of **hardware faults**, **chip faults, UnifiedBus device faults**, and **common faults**, resumable training provides default fault levels and corresponding fault handling policies. For **chip faults**, it also provides default fault frequency and duration settings, along with the corresponding handling policies.

If you need to modify the fault handling policy, refer to this section. Do not modify it unless there are special requirements.

**Supported Configurable Fault Levels<a name="section257513292065"></a>**

The fault levels that can be configured for different types of faults are shown in the table below.

**Table 1**  Configurable fault levels

<a name="table4710459145316"></a>
<table><thead align="left"><tr id="row37104590534"><th class="cellrowborder" valign="top" id="mcps1.2.5.1.1"><p id="p7710135925316"><a name="p7710135925316"></a><a name="p7710135925316"></a>Fault Name</p>
</th>
<th class="cellrowborder" colspan="3" valign="top" id="mcps1.2.5.1.2"><p id="p11175192213564"><a name="p11175192213564"></a><a name="p11175192213564"></a>Configurable Fault Levels</p>
</th>
</tr>
</thead>
<tbody><tr id="row271045905320"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p271015916536"><a name="p271015916536"></a><a name="p271015916536"></a>Node fault</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p66711187562"><a name="p66711187562"></a><a name="p66711187562"></a>NotHandleFault, PreSeparateFault, SeparateFault</p>
</td>
</tr>
<tr id="row3710165935311"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p17710125955315"><a name="p17710125955315"></a><a name="p17710125955315"></a>Chip fault</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p21371428713"><a name="p21371428713"></a><a name="p21371428713"></a>NotHandleFault, RestartRequest, RestartBusiness, FreeRestartNPU, RestartNPU, SeparateNPU, PreSeparateNPU, SubHealthFault</p>
</td>
</tr>
<tr id="row5710125913537"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p10710959185319"><a name="p10710959185319"></a><a name="p10710959185319"></a>UnifiedBus device fault</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p6631112135616"><a name="p6631112135616"></a><a name="p6631112135616"></a>NotHandleFault, SubHealthFault, ResetFault, SeparateFault<span id="ph51441721217"><a name="ph51441721217"></a><a name="ph51441721217"></a>, </span><span id="ph375517710129"><a name="ph375517710129"></a><a name="ph375517710129"></a>RestartRequestFault</span></p>
</td>
</tr>
<tr id="row416145918513"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p116115913517"><a name="p116115913517"></a><a name="p116115913517"></a>Common fault</p>
</td>
<td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.5.1.2 "><p id="p147536536717"><a name="p147536536717"></a><a name="p147536536717"></a>NotHandleFault, SeparateNPU, SubHealthFault<span id="ph632635517598"><a name="ph632635517598"></a><a name="ph632635517598"></a>, PreSeparateNPU</span></p>
</td>
</tr>
</tbody>
</table>

In the table above, the handling policy for each fault level is described as follows.

**Table 2**  Fault levels and handling policies

<a name="table103716651410"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row461812518228"><th class="cellrowborder" valign="top" width="19.06%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12618851162220"></a>Fault Handling Policy</p>
</th>
<th class="cellrowborder" valign="top" width="35.74%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16618125162219"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="23.39%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1163819316544"></a>Rescheduling</p>
</th>
<th class="cellrowborder" valign="top" width="21.81%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p171971327125410"></a>Graceful Fault Tolerance</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row961811511228"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p7618125114229"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1261835110227"></a>Has no service impact and requires no handling.</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p10638123115414"></a>Not handled for now</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p719714273546"></a>Not handled for now</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row116184515226"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p5618751102216"></a>RestartRequest</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p05771854113911"></a>Affects service execution and requires re-executing the service request.</p>
</td>
<td class="cellrowborder" rowspan="5" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p13855131912555"></a>Isolate the chip and reschedule the job.</p>
<div class="note" id="note11901123612819"><a name="note11901123612819"></a><a name="note11901123612819"></a><span class="notetitle">Note:</span><div class="notebody"><p id="p1069261722310"><a name="p1069261722310"></a><a name="p1069261722310"></a>If the inference job subscribes to fault information, and a RestartRequest fault occurs on the inference card used by the job with a fault duration not exceeding 60 seconds, job rescheduling will not be performed. If the fault duration exceeds 60 seconds without recovery, the chip will be isolated and job rescheduling will be performed.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p9145165785517"></a>In inference scenarios, re-execute the inference request; in training scenarios, re-execute the training service.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row14618105116225"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15618851132212"></a>RestartBusiness</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3618851182216"></a>Affects service execution and requires re-executing the service.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1419712272549"></a>Re-execute the service.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row561825132214"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p66188511222"></a>FreeRestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p661865162211"></a>Affects service execution and requires resetting the chip when it becomes idle.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p178789204535"></a>Wait for the chip to become idle and then reset it.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row14618125152210"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p17618155116227"></a>RestartNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p108302057102114"></a>Affects service execution and requires resetting the chip immediately.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p969972925312"></a>Immediately stop the training service, reset the chip, and then re-execute the service.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row1061895115227"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p18618151202216"></a>Unrecoverable and requires isolating the chip.</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p019742745411"></a>Isolate the chip and reschedule the job.</p>
</td>
</tr>
<tr id="row870814247412"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p5708202454117"><a name="p5708202454117"></a><a name="p5708202454117"></a>SeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p12708162474117"><a name="p12708162474117"></a><a name="p12708162474117"></a>The job will definitely be affected.</p>
<div class="note" id="note1521013164613"><a name="note1521013164613"></a><a name="note1521013164613"></a><span class="notetitle">Note:</span><div class="notebody"><p id="p92101114465"><a name="p92101114465"></a><a name="p92101114465"></a>When the UnifiedBus device fault level is SeparateFault, it indicates that the service has failed to run, and the component or board needs to be replaced.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p0708624204112"><a name="p0708624204112"></a><a name="p0708624204112"></a>Reschedule the job.</p>
<div class="note" id="note44451347164716"><a name="note44451347164716"></a><a name="note44451347164716"></a><span class="notetitle">[!NOTE] Note</span><div class="notebody"><p id="p64453471479"><a name="p64453471479"></a><a name="p64453471479"></a>For UnifiedBus device faults, the fault handling policy represented by this fault level is to stop the current training job, isolate the node, and reschedule the job.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p137081824174117"><a name="p137081824174117"></a><a name="p137081824174117"></a>-</p>
</td>
</tr>
<tr id="row5706333131216"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p177061833201220"><a name="p177061833201220"></a><a name="p177061833201220"></a><span id="ph141513510124"><a name="ph141513510124"></a><a name="ph141513510124"></a>RestartRequestFault</span></p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p070623351220"><a name="p070623351220"></a><a name="p070623351220"></a><span id="ph18501459184"><a name="ph18501459184"></a><a name="ph18501459184"></a>The service has failed to run and requires re-executing the service request.</span></p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p770653313124"><a name="p770653313124"></a><a name="p770653313124"></a><span id="ph38912127169"><a name="ph38912127169"></a><a name="ph38912127169"></a>Stop the current training job, isolate the node, and reschedule the job.</span></p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p6706113331213"><a name="p6706113331213"></a><a name="p6706113331213"></a>In inference scenarios, re-execute the inference request; in training scenarios, re-execute the training service.</p>
</td>
</tr>
<tr id="row3938182254418"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p39381822174417"><a name="p39381822174417"></a><a name="p39381822174417"></a>ResetFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p1193862274418"><a name="p1193862274418"></a><a name="p1193862274418"></a>The service has failed to run.</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p184323519501"><a name="p184323519501"></a><a name="p184323519501"></a>Stop the current training job, isolate the node, and reschedule the job.</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p18938822204411"><a name="p18938822204411"></a><a name="p18938822204411"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row102215292529"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p16227299522"></a>PreSeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p546081915499"></a>Does not affect the service for now, but jobs will no longer be scheduled to this chip subsequently.</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p222102912521"></a>Pre-isolate the chip.</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p12221329155217"></a>Pre-isolate the chip.</p>
</td>
</tr>
<tr id="row84541721401"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="p174559214016"><a name="p174559214016"></a><a name="p174559214016"></a>PreSeparateFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="p145562114011"><a name="p145562114011"></a><a name="p145562114011"></a>May cause the job to be affected.</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="p54556214409"><a name="p54556214409"></a><a name="p54556214409"></a>If there is a job on this node, it will not be handled. During subsequent scheduling, jobs will not be scheduled to this node.</p>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="p1245572144015"><a name="p1245572144015"></a><a name="p1245572144015"></a>-</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_row0352224175218"><td class="cellrowborder" valign="top" width="19.06%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p835213245522"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="35.74%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p1354813311915"></a>Handled based on the value of the subHealthyStrategy parameter configured in the job YAML. For details, see <a href="../../api/">YAML Configuration Description</a>.</p>
</td>
<td class="cellrowborder" valign="top" width="23.39%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p3352524125220"></a>When a sub-health fault occurs on the chip, it needs to be handled according to the <a href="./06_configuring_the_job_yaml_file.md">Configuring YAML</a>.</p>
<div class="note" id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_note7936204710536"></a><span class="notetitle">Note:</span><div class="notebody"><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p15222114115810"></a>If a fault of another level occurs on the chip subsequently, the SubHealthFault handling policy will not affect the handling of faults at other levels.</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="21.81%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"></a><a name="zh-cn_topic_0000002395188553_zh-cn_topic_0000002171521445_p8352172425218"></a>Handled according to the policy.</p>
</td>
</tr>
</tbody>
</table>

## Node Hardware Faults<a name="ZH-CN_TOPIC_0000002479226584"></a>

### Configuration File Description<a name="ZH-CN_TOPIC_0000002479226562"></a>

Resumable training performs hierarchical processing for different levels of **node hardware faults**. NodeD obtains the fault code of the current fault and processes the fault accordingly based on the fault level configured for the fault code in `NodeDConfiguration.json`. The supported fault levels and handling methods for node hardware faults are described as follows.

The NodeD configuration file `NodeDConfiguration.json` is a system configuration file. Do not modify it arbitrarily unless you have special requirements. If you need to modify the fault level of a fault code, you can do so through the `mindx-dl-node-fault-config` file created from `NodeDConfiguration.json`. For operation instructions, see [(Optional) Configuring Node Hardware Fault Levels](#optional-configuring-node-hardware-fault-levels).For fault level descriptions and node status descriptions, see [Customizing Node Faults](../../api/noded.md#customizing-node-faults).

### (Optional) Configuring Node Hardware Fault Levels<a name="ZH-CN_TOPIC_0000002511346507"></a>

When creating a NodeD image, the fault level configuration file `NodeDConfiguration.json` is built into the image. When NodeD starts, it reads the default configuration from this file as the basis for current fault handling.

If you want to customize fault levels, create a ConfigMap file (`mindx-dl-node-fault-config`) in the cluster.

- If `mindx-dl-node-fault-config` exists in the cluster when NodeD starts, NodeD will prioritize the content configured in the existing `mindx-dl-node-fault-config` as the basis for current fault handling.
- If `mindx-dl-node-fault-config` exists in the cluster after reinstalling NodeD, NodeD's default `NodeDConfiguration.json` will not take effect, and the existing `mindx-dl-node-fault-config` in the cluster will be used. If you want to use the default configuration of `NodeDConfiguration.json`, delete `mindx-dl-node-fault-config` so that NodeD reads the default `NodeDConfiguration.json` file.
- If there are issues such as format errors in the content of `mindx-dl-node-fault-config`, NodeD will read the content of the `NodeDConfiguration.json` file built into the image by default as the basis for current fault handling.

**Procedure<a name="section25164134219"></a>**

Taking fault code `0100001D` as an example, the following shows how to modify the handling policy for the current fault from `NotHandleFault` (no handling required) to `PreSeparateFault` (do not handle if there are jobs on the node, and do not schedule subsequent jobs to the node).

1. Log in to the environment and go to the NodeD decompression directory.
2. Run the following command to create the ConfigMap file (`mindx-dl-node-fault-config`) required for dynamic fault level configuration.

    ```shell
    kubectl create cm mindx-dl-node-fault-config -n mindx-dl  --from-file=./NodeDConfiguration.json
    ```

    Command output:

    ```ColdFusion
    configmap/mindx-dl-node-fault-config created
    ```

    **Table 1** Parameter description

    <a name="table1925220306444"></a>
    <table><thead align="left"><tr id="row172531430134411"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p16253163094420"><a name="p16253163094420"></a><a name="p16253163094420"></a>Parameter Name</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p152534301443"><a name="p152534301443"></a><a name="p152534301443"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1325318306446"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p15214952162210"><a name="p15214952162210"></a><a name="p15214952162210"></a>mindx-dl-node-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p621417523229"><a name="p621417523229"></a><a name="p621417523229"></a>Name of the created <span id="ph188631730142314"><a name="ph188631730142314"></a><a name="ph188631730142314"></a>ConfigMap</span> file. This file name cannot be modified.</p>
    </td>
    </tr>
    <tr id="row925343011442"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p82141952122212"><a name="p82141952122212"></a><a name="p82141952122212"></a>mindx-dl</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p0214952142217"><a name="p0214952142217"></a><a name="p0214952142217"></a>Namespace name. This namespace cannot be modified.</p>
    </td>
    </tr>
    <tr id="row1253183012444"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p182141521222"><a name="p182141521222"></a><a name="p182141521222"></a>NodeDConfiguration.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p22148525226"><a name="p22148525226"></a><a name="p22148525226"></a>Used to configure fault codes and their corresponding fault levels. Must be consistent with the NodeDConfiguration.json file name.</p>
    </td>
    </tr>
    </tbody>
    </table>

3. Run the following command to edit the `mindx-dl-node-fault-config` file.

    ```shell
    kubectl edit cm -n mindx-dl mindx-dl-node-fault-config
    ```

4. In the `mindx-dl-node-fault-config` file, locate the fault code `0100001D`.

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
    >During fault level customization, if the following issues occur accidentally, this modification will be invalid, and NodeD will use the last saved configuration for processing.
    >- The file format is abnormal or the fault code is incorrect. The fault code can only be an 8-character string containing digits and letters.
    >- The same fault code is configured in multiple fault levels at the same time.

5. Delete the fault code `0100001D` from `NotHandleFaultCodes` and add it to `PreSeparateFaultCodes`.

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

6. After the modification is complete, press `Esc`, enter `:wq!` to save and exit.
7. After the `mindx-dl-node-fault-config` file is updated, check whether the operation is successful.
    1. Run the following command to query the log name of NodeD.

        ```shell
        kubectl get pods -A | grep noded
        ```

        Command output:

        ```ColdFusion
        mindx-dl      noded-c5f52   1/1     Running   0               2m16s
        ```

    2. Query the log information of NodeD by using the queried log name.

        ```shell
        kubectl logs noded-c5f52 -n mindx-dl -f
        ```

        If the log contains "update fault config success", it indicates that the dynamic fault code configuration operation is successful.

## Chip Faults<a name="ZH-CN_TOPIC_0000002479226466"></a>

### Overview<a name="ZH-CN_TOPIC_0000002511346521_0101"></a>

Both Ascend Device Plugin and ClusterD provide the capability to manually isolate chips based on fault frequency. The functional differences between the two are as follows:

- Ascend Device Plugin determines faults based on the node dimension and counts the frequency of actual faults that occur.
- ClusterD determines faults based on the job dimension.
    - If multiple chips under a single job experience the same fault simultaneously within 30 seconds, it excludes a hardware fault, and does not count the fault frequency. This judgment rule applies to most scenarios. For scenarios such as a Pod being deleted but residual processes remaining, the fault frequency counts may have deviations.
    - Only new faults can trigger the judgment of whether the fault frequency for manual chip isolation has reached the upper limit. If the configured threshold is adjusted to the current count, isolation will not be triggered immediately; the judgment logic will only be triggered when the next fault occurs.
    - After ClusterD restarts, the frequency count information will be lost, and the fault frequency for manual chip isolation will start counting from zero.
    - If job scheduling does not meet expectations after removing the isolation, check whether the node has the label `huawei.com/scheduler.chip1softsharedev.enable=false`. If this label exists, delete it.

The fault codes involved in the manual chip isolation feature of Ascend Device Plugin and ClusterD theoretically do not need to be duplicated. If you do not want to use the isolation feature of Ascend Device Plugin, see the [(Optional) Configuring Chip Fault Frequency and Duration](#optional-configuring-chip-fault-frequency-and-duration) section to delete the manual chip isolation-related configuration in the `faultCustomization.json` file. If you do not want to use the isolation feature of ClusterD, see the [(Optional) Configuring Chip Fault Frequency](#optional-configuring-chip-fault-frequency) section to disable manual chip isolation.

If both Ascend Device Plugin and ClusterD have manually isolated the same chip, the isolation must be removed separately for each. For the method to remove isolation in Ascend Device Plugin, see "Manually Recovering Force-Isolated Chips" in [(Optional) Configuring Chip Fault Frequency and Duration](#optional-configuring-chip-fault-frequency-and-duration). For the method to remove isolation in ClusterD, see "Manually Recovering Manually Isolated Chips" in [(Optional) Configuring Chip Fault Frequency](#optional-configuring-chip-fault-frequency).

### Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511346521_02"></a>

#### Configuration File Description<a name="ZH-CN_TOPIC_0000002511346521"></a>

For **chip faults**, resumable training supports processing based on fault level, fault frequency, and fault duration configuration.

- When performing hierarchical processing for **different levels** of chip faults, Ascend Device Plugin obtains the fault code of the current fault and processes the fault accordingly based on the fault level configured for the fault code in `faultCode.json`.
- When processing based on the **fault frequency and duration** of chip faults, Ascend Device Plugin obtains the fault code of the current fault and processes the fault accordingly based on the fault frequency and duration configured for the fault in `faultCustomization.json`.

`faultCode.json` and `faultCustomization.json` are system configuration files. Do not modify them arbitrarily unless you have special requirements. If the default frequency fault configuration of Ascend Device Plugin contains faults that can be triggered by software reasons, you can delete the corresponding fault code yourself. (Software reasons may cause a certain fault to repeatedly occur a large number of times within a short period under a single job, causing Ascend Device Plugin to detect that the fault has reached the fault frequency and place a large number of devices into manual isolation state.)

If you need to modify the fault level corresponding to a fault code, you can do so through the `mindx-dl-fault-config` file created from `faultCode.json` and `faultCustomization.json`.

>[!NOTE]
>
>- For the fault code corresponding to each fault, see the [Chip Fault Code References](../../references/appendix.md#chip-fault-code-references) section.
>- For the fault levels that can be configured for chip faults, see [Fault Levels](#zh-cn_topic_0000002171521445_section5245155017242).
>- For the fault frequency and duration that can be configured for chip faults, see [Fault Frequency and Duration](#zh-cn_topic_0000002171521445_section115842029104220).

**Fault Levels in faultCode.json<a name="zh-cn_topic_0000002171521445_section5245155017242"></a>**

Resumable training performs hierarchical handling for different levels of chip faults. If you need to modify the fault level of a fault code, see [(Optional) Configuring Chip Fault Levels](#optional-configuring-chip-fault-levels) for operation instructions.

After Ascend Device Plugin obtains the chip fault code from the driver, it classifies the fault into several levels based on the impact of the fault code on the device and service. For details, see [Table 1](../../api/ascend_device_plugin.md#custom-chip-faults).

>[!NOTE]
>
>- The training process must be stopped before chip reset; otherwise, the reset will fail.
>- If Ascend Device Plugin receives an unrecognized fault code (not saved in `faultCode.json`) through subscription, it performs fault handling according to the handling suggestion provided by the subscription interface by default. If the fault level received by the subscription interface is `Hint` or `Minor`, it is handled at the `NotHandleFault` level; if the fault level is any other level, it is handled at the `SeparateNPU` level.

**Fault Frequency and Duration<a name="zh-cn_topic_0000002171521445_section115842029104220"></a>**

Resumable training handles the fault frequency and duration of chip faults. Certain hardware faults may occur repeatedly during a single training job, causing the training job to be interrupted and rescheduled repeatedly. The cluster scheduling components provide an initialization configuration file, `faultCustomization.json`, to elevate the fault level for the fault codes corresponding to these faults.

- The relationship between the initialization configuration in the `faultCustomization.json` file and fault types is described in [Initialization Configuration and Fault Types](#zh-cn_topic_0000002171521445_section13684172919539).
- For the default configuration (default values) of the `faultCustomization.json` file, see [Table 2](../../api/ascend_device_plugin.md#custom-chip-faults).
- If you need to modify the fault frequency and duration configuration, see [(Optional) Configuring Chip Fault Frequency and Duration](#optional-configuring-chip-fault-frequency-and-duration) for operation instructions.

**Initialization Configuration and Fault Types<a name="zh-cn_topic_0000002171521445_section13684172919539"></a>**

The current `faultCustomization.json` file only provides initialization configuration for upgrading the fault level of identifiable hardware faults.

If the following fault occurs three times within 24 hours, the chip fault level is upgraded to `ManuallySeparateNPU`, a fault level that requires manual intervention. For details, see [faultCustomization.json Parameter Description](#zh-cn_topic_0000002171521445_section33036167576).

The following example uses the fault name H`BMC Ca Parity Error`, corresponding to fault code `80E18005`, to escalate the current fault level to `ManuallySeparateNPU` (a fault level that requires manual intervention).

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
>- When the fault handling policy is ManuallySeparateNPU, you can refer to the steps in "Manually Recovering Force-Isolated Chips" in [(Optional) Configuring Chip Fault Frequency and Duration](#optional-configuring-chip-fault-frequency-and-duration) for processing.
>- In addition to identifiable hardware faults, the `faultCustomization.json` file also includes the following types of faults.
>     - Faults that do not require handling: These faults do not affect training jobs or devices, and no initialization configuration for upgrading the fault level is provided.
>     - Faults that cannot be identified as hardware or software faults: These faults cannot be accurately identified as hardware or software faults and will affect training jobs. No initialization configuration for upgrading the fault level is provided for such faults. You are advised to manually configure the maximum number of resumable training times supported and the fault handling policy after the maximum number is reached based on the actual situation. For details, see [(Optional) Configuring Chip Fault Frequency and Duration](#optional-configuring-chip-fault-frequency-and-duration).
>     - Software configuration faults: These faults are software configuration issues and do not occur under normal circumstances. No initialization configuration for upgrading the fault level is provided for such faults. You are advised to check whether the software versions are compatible.

**faultCustomization.json Parameter Description<a name="zh-cn_topic_0000002171521445_section33036167576"></a>**

When there is no need to manually modify the `faultCustomization.json` file, Ascend Device Plugin performs fault handling according to the default configuration (default values) of `faultCustomization.json`. For the parameter description of the `faultCustomization.json` file, see [Table 2](../../api/ascend_device_plugin.md#custom-chip-faults).

#### (Optional) Configuring Chip Fault Levels

When the Ascend Device Plugin image is built, the `faultCode.json` and `faultCustomization.json` configuration files are built into the image. When Ascend Device Plugin starts, it reads the default configurations of these two files as the basis for current fault handling. For the description of `faultCode.json` and `faultCustomization.json`, see [Configuration File Description](#ZH-CN_TOPIC_0000002511346521).

If you want to customize the fault level or graceful fault tolerance related configurations, create a ConfigMap file (`mindx-dl-fault-config`) in the cluster.

- If `mindx-dl-fault-config` exists in the cluster when Ascend Device Plugin starts, Ascend Device Plugin will prioritize the content configured in the existing `mindx-dl-fault-config` as the basis for current fault handling.
- If `mindx-dl-fault-config` exists in the cluster after Ascend Device Plugin is reinstalled, the default `faultCode.json` of Ascend Device Plugin will not take effect, and the existing `mindx-dl-fault-config` in the cluster will be used instead.
- If you want to use the default configuration of `faultCode.json` or `faultCustomization.json`, you can delete `mindx-dl-fault-config` so that Ascend Device Plugin reads the default `faultCode.json`, `SwitchFaultCode.json`, or `faultCustomization.json` files.
- If there are issues such as format errors in the ConfigMap file content, Ascend Device Plugin will read the content of the built-in ConfigMap file in the image by default as the basis for current fault handling.

**Configuring Fault Levels Using faultCode.json<a name="zh-cn_topic_0000001951258609_section112139052513"></a>**

Take the fault `dmp_daemon` (node status detection anomaly), corresponding to fault code `80E21007`, as an example. The following shows the operation to modify the current fault handling policy from `NotHandleFaultCodes` (no handling required) to `RestartNPUCodes` (isolate the chip and perform job rescheduling).

1. Log in to the environment and go to the Ascend Device Plugin decompression directory.
2. Run the following command to create the ConfigMap file (`mindx-dl-fault-config`) required for dynamically configuring fault codes.

    ```shell
    kubectl create cm mindx-dl-fault-config -n kube-system --from-literal="PollInterval=300" --from-file=./faultCode.json
    ```

    Command output:

    ```ColdFusion
    configmap/mindx-dl-fault-config created
    ```

    **Table 1** Parameter description

    <a name="zh-cn_topic_0000001951258609_table16314861918"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001951258609_row763648161914"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001951258609_p16631548171910"><a name="zh-cn_topic_0000001951258609_p16631548171910"></a><a name="zh-cn_topic_0000001951258609_p16631548171910"></a>Parameter Name</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001951258609_p1663144816197"><a name="zh-cn_topic_0000001951258609_p1663144816197"></a><a name="zh-cn_topic_0000001951258609_p1663144816197"></a>Required</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001951258609_p775918210209"><a name="zh-cn_topic_0000001951258609_p775918210209"></a><a name="zh-cn_topic_0000001951258609_p775918210209"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001951258609_row36354871915"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1863164816197"><a name="zh-cn_topic_0000001951258609_p1863164816197"></a><a name="zh-cn_topic_0000001951258609_p1863164816197"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p1063194861910"><a name="zh-cn_topic_0000001951258609_p1063194861910"></a><a name="zh-cn_topic_0000001951258609_p1063194861910"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p157595292015"><a name="zh-cn_topic_0000001951258609_p157595292015"></a><a name="zh-cn_topic_0000001951258609_p157595292015"></a>The name of the <span id="zh-cn_topic_0000001951258609_ph126311642183015"><a name="zh-cn_topic_0000001951258609_ph126311642183015"></a><a name="zh-cn_topic_0000001951258609_ph126311642183015"></a>ConfigMap</span> file used for dynamically configuring fault codes. This file name cannot be modified.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row763184812192"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1963194819195"><a name="zh-cn_topic_0000001951258609_p1963194819195"></a><a name="zh-cn_topic_0000001951258609_p1963194819195"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p76316488192"><a name="zh-cn_topic_0000001951258609_p76316488192"></a><a name="zh-cn_topic_0000001951258609_p76316488192"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p276092142019"><a name="zh-cn_topic_0000001951258609_p276092142019"></a><a name="zh-cn_topic_0000001951258609_p276092142019"></a>The namespace where mindx-dl-fault-config resides. This namespace name cannot be modified.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row86314881910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p964144891914"><a name="zh-cn_topic_0000001951258609_p964144891914"></a><a name="zh-cn_topic_0000001951258609_p964144891914"></a>PollInterval</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p1164748191916"><a name="zh-cn_topic_0000001951258609_p1164748191916"></a><a name="zh-cn_topic_0000001951258609_p1164748191916"></a>No</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p876012211206"><a name="zh-cn_topic_0000001951258609_p876012211206"></a><a name="zh-cn_topic_0000001951258609_p876012211206"></a>If this parameter is not specified, the default value is 300s. It specifies the polling interval for checking whether the mindx-dl-fault-config file has been updated. The unit is seconds, and the value range is 30 to 3600. Changes to PollInterval will take effect in the next polling cycle.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row176474851911"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1964748141915"><a name="zh-cn_topic_0000001951258609_p1964748141915"></a><a name="zh-cn_topic_0000001951258609_p1964748141915"></a>faultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p10641648191915"><a name="zh-cn_topic_0000001951258609_p10641648191915"></a><a name="zh-cn_topic_0000001951258609_p10641648191915"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p147602211206"><a name="zh-cn_topic_0000001951258609_p147602211206"></a><a name="zh-cn_topic_0000001951258609_p147602211206"></a>Used to store fault codes. It must be consistent with the faultCode.json file name.</p>
    </td>
    </tr>
    </tbody>
    </table>

3. Run the following command to edit the `mindx-dl-fault-config` file.

    ```shell
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

4. In the `mindx-dl-fault-config` file, locate the fault code `80E21007`.

    ```json
    "NotHandleFaultCodes":[

    "80E21007","80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    ...
      ],
    ...
    ```

    >[!NOTE]
    >If the same fault code is configured in multiple fault levels, the configuration will be displayed as successful, but the fault will be handled according to the higher-level fault by default.

5. Delete the fault code `80E21007` from `NotHandleFaultCodes` and add it to `RestartNPUCodes`.

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

6. After modification, press `Esc`, type `:wq!` to save and exit.
7. Wait for the `mindx-dl-fault-config` file update to take effect (`PollInterval` defaulted to `300s` if not specified), then check whether the operation is successful.
    1. Run the following command to query the log name of Ascend Device Plugin.

        ```shell
        kubectl get pods -A | grep ascend-device-plugin
        ```

        Command output:

        ```ColdFusion
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2. Use the queried log name to query the component log information of Ascend Device Plugin.

        ```shell
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        If the log shows "load fault code from configmap success", it indicates that the manual fault code configuration operation was successful.

#### (Optional) Configuring Chip Fault Frequency and Duration<a name="ZH-CN_TOPIC_0000002511426473"></a>

When the Ascend Device Plugin image is built, the `faultCode.json` and `faultCustomization.json` configuration files are built into the image. When Ascend Device Plugin starts, it reads the default configurations from these two files as the basis for current fault handling. For descriptions of `faultCode.json` and `faultCustomization.json`, see [Configuration File Description](#ZH-CN_TOPIC_0000002511346521).

If you want to customize the chip fault frequency and duration, create a ConfigMap file (`mindx-dl-fault-config`) in the cluster.

- If `mindx-dl-fault-config` exists in the cluster when Ascend Device Plugin starts, Ascend Device Plugin will prioritize the content configured in the existing `mindx-dl-fault-config` as the basis for current fault handling.
- If `mindx-dl-fault-config` exists in the cluster after reinstalling Ascend Device Plugin, the default `faultCustomization.json` of Ascend Device Plugin will not take effect, and the existing `mindx-dl-fault-config` in the cluster will be used. If you want to use the default configuration of `faultCustomization.json`, you can delete `mindx-dl-fault-config` so that Ascend Device Plugin reads the default `faultCustomization.json` file.
- If there are issues such as format errors in the ConfigMap file content, Ascend Device Plugin will read the content of the built-in ConfigMap file in the image as the basis for current fault handling by default.

>[!CAUTION]
>Modifying the fault frequency is a high-risk operation. Improper modification may cause chips to be mistakenly isolated. For example, software faults caused by job errors may occur repeatedly in large numbers within a short period, causing Ascend Device Plugin to detect that the fault frequency has been reached, and then place a large number of chips into the manual isolation state, making a large number of nodes unschedulable.

**Procedure<a name="section141902103110"></a>**

Taking fault code `80CB8002` as an example, if this fault occurs repeatedly on a certain chip, causing the training service to be rescheduled repeatedly, you can manually configure the supported maximum number of resumable training times within 24 hours to 2. The fault handling policy is `ManuallySeparateNPU` after the maximum number is reached.

1. Log in to the environment and go to the Ascend Device Plugin decompression directory.
2. Run the following command to check whether `mindx-dl-fault-config` has been created based on the `faultCode.json` file.

    ```shell
    kubectl describe cm -n kube-system mindx-dl-fault-config
    ```

    - If `mindx-dl-fault-config` exists and contains the relevant fields of `faultCustomization.json`, perform [Step 4](#zh-cn_topic_0000002136360238_li38432520129) to edit the file.
    - If `mindx-dl-fault-config` exists but does not contain the relevant fields of `faultCustomization.json`, save the content of `mindx-dl-fault-config` first, delete the `mindx-dl-fault-config` file, and then perform [Step 3](#zh-cn_topic_0000002136360238_li1946014413123) to create the file.
    - If `mindx-dl-fault-config` does not exist, perform [Step 3](#zh-cn_topic_0000002136360238_li1946014413123) to create it.

3. <a name="zh-cn_topic_0000002136360238_li1946014413123"></a>Run the following command to create the ConfigMap file (`mindx-dl-fault-config`) required for chip fault frequency configuration.

    ```shell
    kubectl create cm mindx-dl-fault-config -n kube-system --from-literal="PollInterval=300" --from-file=./faultCode.json --from-file=./faultCustomization.json
    ```

    Command output:

    ```ColdFusion
    configmap/mindx-dl-fault-config created
    ```

    **Table 1**  Parameter description

    <a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_table16314861918"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row763648161914"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"></a>Parameter Name</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"></a>Required</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row36354871915"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"></a>The <span id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"></a>ConfigMap</span> file name required for dynamically configuring fault codes. This file name cannot be modified.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row763184812192"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"></a>The namespace where mindx-dl-fault-config resides. This namespace name cannot be modified.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row86314881910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"></a>PollInterval</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"></a>No</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"></a>If this parameter is not specified, the default value is 300s. It specifies the polling interval for checking whether the mindx-dl-fault-config file is updated, in seconds. The value ranges from 30 to 3600. Modifications to PollInterval take effect in the next polling cycle.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row176474851911"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"></a>faultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"></a>Used to store fault codes. The file name must be consistent with faultCode.json.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_row9289716194614"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"></a>faultCustomization.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"></a>No</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"></a><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"></a>Used to customize configurations such as graceful fault tolerance time, fault frequency, and fault duration (only parameter plane network faults are supported). If this parameter is not specified, there is no fault frequency configuration, and other configurations are processed using default values. The file name must be consistent with faultCustomization.json.</p>
    </td>
    </tr>
    </tbody>
    </table>

4. <a name="zh-cn_topic_0000002136360238_li38432520129"></a>Run the following command to edit the `mindx-dl-fault-config` file.

    ```shell
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

    Modify the chip fault frequency and duration based on the actual situation.

    ```json
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
    PollInterval: "300"
    # Modify the fault level of chip faults
    faultCode.json: |
    {
    "NotHandleFaultCodes":[
    ...
    }
    # Modify the fault frequency and duration of chip faults
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
    ```

5. In the `mindx-dl-fault-config` file, add the following code under the `FaultFrequency` field to set the maximum number of resumable training times supported for fault `80CB8002` within 24 hours to 2, and the handling policy for the fault after reaching the maximum number to `ManuallySeparateNPU`.

    ```json
    {
      "EventId": ["80CB8002"],
      "TimeWindow": 86400,
      "Times": 2,
      "FaultHandling": "ManuallySeparateNPU"
    }
    ```

6. After modification, press `Esc` and enter `:wq!` to save and exit.
7. After the `mindx-dl-fault-config` file update takes effect (`PollInterval` defaulted to `300s` if not specified), check whether the operation is successful.
    1. Run the following command to query the log name of Ascend Device Plugin.

        ```shell
        kubectl get pods -A | grep ascend-device-plugin
        ```

        Command output:

        ```ColdFusion
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2. Use the queried log name to query the log information of Ascend Device Plugin.

        ```shell
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        >[!NOTE]
        >- If the log contains "load fault customization from configmap complete", it indicates that the manual chip fault frequency configuration operation is successful.
        >- If the log contains "modify  _xxx_  success", it indicates that the <i>xxx</i> parameter in faultCustomization.json in the ConfigMap is set successfully.
        >- If the log contains "insert fault frequency success", it indicates that the occurrence time of a frequency fault has been recorded. Within the frequency window, after the number of fault records for that chip reaches the fault frequency trigger threshold, the corresponding fault level will be reported.

8. (Optional) Manually restore a forcibly isolated chip. When the fault handling policy is `ManuallySeparateNPU`, the chip remains isolated after fault recovery. If the release conditions are not met and you need to manually restore the forcibly isolated chip, do as follows.
    1. Run the following command to find `device-info-cm` reported by Ascend Device Plugin of this node.

        ```shell
        kubectl get cm -n kube-system | grep deviceinfo | grep {nodeName}
        ```

    2. Run the following command to edit `device-info-cm`.

        ```shell
        kubectl edit cm -n kube-system {configMapName}
        ```

    3. Delete the name of the recovered, healthy chip following `ManuallySeparateNPU` under `data`.

        ```Yaml
        apiVersion: v1
        kind: ConfigMap
        data:
          DeviceInfoCfg: '{"DeviceInfo":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-Fault":"[]","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Unhealthy":""},"UpdateTime":1718702470},"CheckCode":"4f00cf1d220da26a8fdbeb5ba163a751d4b264c48b81d22149257e272ae3b413"}'
          ManuallySeparateNPU: Ascend910-0
        ```

        >[!NOTE]
        >Delete all chip names after the `ManuallySeparateNPU` field and set the value to empty `""`.

    4. After modification, press `Esc` and enter `:wq!` to save and exit.
    5. Wait for one reporting cycle (if device information changes, it will be reported within the health status check cycle; if device information does not change, the reporting cycle is fixed at 5 minutes), then run the following command to check whether the chip name just deleted exists in `ManuallySeparateNPU` in `device-info-cm`. If it does not exist, the chip has successfully recovered to a healthy state and can continue to be used normally.

        ```shell
        kubectl describe cm -n kube-system {configMapName}
        ```

### ClusterD<a name="ZH-CN_TOPIC_0000002511346521_03"></a>

#### Configuration Description<a name="ZH-CN_TOPIC_0000002511346521_04"></a>

Resumable training can handle chip faults based on the fault frequency configuration.

When performing hierarchical processing for different levels of chip faults, ClusterD obtains the fault code and fault level of the current fault. For faults at levels other than `NotHandleFault` and `SubHealthFault`, the chip status is set to manual isolation based on the fault frequency configured in the ConfigMap (`clusterd-config-cm`). For parameter descriptions of this ConfigMap, see [Table 1](../../developer_guide/installation_deployment/manual_installation/06_clusterd.md).

>[!NOTE]
>
>- `clusterd-config-cm` is a system configuration. Do not modify it arbitrarily unless you have special requirements. If you need to modify the detection switch for manually isolated chips, fault frequency, isolation removal time, etc., you modify this ConfigMap by referring to [(Optional) Configuring Chip Fault Frequency Configuration](#optional-configuring-chip-fault-frequency).
>- Configuring the fault code detection range is not supported. ClusterD makes judgments based on the fault level reported by Ascend Device Plugin. For faults at levels other than `NotHandleFault` and `SubHealthFault`, they will all be included in the manual isolation chip detection process.

#### (Optional) Configuring Chip Fault Frequency

When ClusterD is installed, the ConfigMap (c`lusterd-config-cm`) is automatically created as the detection basis for manually isolated chips. For parameter descriptions of this ConfigMap, see [Table 1](../../developer_guide/installation_deployment/manual_installation/06_clusterd.md).

If you want to customize the chip fault frequency, you can modify this ConfigMap. If the modified ConfigMap content has format errors or other issues, ClusterD will retain the last successfully read configuration as the detection basis for manual chip isolation. If the ConfigMap content read by ClusterD at startup is incorrect, the manual chip isolation detection mechanism will be disabled by default until the format and content are correct.

**Procedure<a name="section14190101"></a>**

Take adjusting a manual chip isolation threshold from the default value of 3 occurrences to 5 occurrences within 24 hours as an example.

1. Log in to the environment and run the following command to query the current configuration.

    ```shell
    kubectl describe cm -n cluster-system clusterd-config-cm
    ```

    - If `clusterd-config-cm` exists, proceed to [Step 3](#li01010203) for editing.
    - If `clusterd-config-cm` does not exist, proceed to [Step 2](#li010102) for creation.

    >[!NOTE]
    >Under normal circumstances, `clusterd-config-cm` exists. If it does not exist, check whether there are errors in the ClusterD installation process.

2. <a name="li010102"></a>Create `clusterd-config-cm` required for manual chip isolation detection.

    Save the following content as the file `cm.yaml`:

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

    Run the following command:

    ```shell
    kubectl apply -f cm.yaml
    ```

    The following example output indicates a successful creation.

    ```ColdFusion
    configmap/clusterd-config-cm created
    ```

3. <a name="li01010203"></a>Run the following command to edit `clusterd-config-cm`.

    ```shell
    kubectl edit cm -n cluster-system clusterd-config-cm
    ```

    Modify the fault frequency for manual chip isolation based on the actual situation. For parameter descriptions, see [Table 1](../../developer_guide/installation_deployment/manual_installation/06_clusterd.md).

    ```Yaml
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
      manually_separate_policy.conf: |
        # Modify the detection switch for manually isolated chips
        enabled: true
        separate:
          # Modify the fault frequency of the manually isolated chip
          fault_window_hours: 24
          fault_threshold: 5   # Change from 3 to 5
        release:
          # Modify the De-isolation Time
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

4. After the modification is complete, press the `Esc` key, enter `:wq!` to save and exit.
5. Wait for the `clusterd-config-cm` update to take effect (the detection cycle of ClusterD is 300s), and then check whether the operation is successful.
    1. Run the following command to query the log name of ClusterD.

        ```shell
        kubectl get pods -A | grep clusterd
        ```

        Command output:

        ```ColdFusion
        mindx-dl      clusterd-559bf4bd6-z9hv4   1/1     Running   0             4m23s
        ```

    2. Use the queried component log name to query the log information of ClusterD.

        ```shell
        kubectl logs -f -n mindx-dl clusterd-559bf4bd6-z9hv4
        ```

        >[!NOTE]
        >- If the log shows "load manually separate policy config success", it indicates that the operation to manually modify the fault frequency for manual chip isolation was successful.
        >- If the log shows "node: xx, dev: xx, code: xx is not found in manual fault cache, add", it indicates that this fault triggers manual isolation.
        >- If the log shows "node: xx, dev: xx, code: xx is found in manual fault cache, update last separate time", it indicates that a fault triggering manual chip isolation has once again reached the fault frequency for manual isolation, and `LastSeparateTime` in `clusterd-manual-info-cm` will be updated. For a description of `clusterd-manual-info-cm`, see [clusterd-manual-info-cm](../../api/clusterd/00_cluster_resources.md#clusterd-manual-info-cm).

6. (Optional) Manually recover a manually isolated chip. When the fault handling policy is `ManuallySeparateNPU`, the chip remains in an isolated state after fault recovery, and you can manually recover the manually isolated chip.

    1. Run the following command to edit the ConfigMap `clusterd-manual-info-cm`.

        ```shell
        kubectl edit cm -n cluster-system clusterd-manual-info-cm
        ```

    2. Delete the name of the chip to be removed from manual isolation following the `Total` field under `Data`, for example, `Ascend910-2`.

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

    3. After the modification is complete, press `Esc` and enter `:wq!` to save and exit.
    4. After waiting for 15 seconds, run the following command to check whether `Ascend910-2` still exists in the `Total` and `Detail` fields of `clusterd-manual-info-cm`. Also, check whether the `ManuallySeparateNPU` fault of this chip exists in `cluster-info-device-\${m}`. If it does not exist, the chip has been successfully removed from manual isolation and can continue to be used normally.

        ```shell
        kubectl describe cm -n cluster-system clusterd-manual-info-cm
        ```

        >[!NOTE]
        >- Only deletion of chips from the `Total` field is supported; manual addition is not supported. Modification of other content is not supported.
        >- After manually recovering a chip from manual isolation, the fault count of the chip will be cleared. Manual isolation will be triggered again only when the frequency is reached again.
        >- If you need to delete all manually isolated chips on a node, you must delete all chip names following the `Total` field and set the value to `[]`. To remove all manually isolated chips at once, you can directly delete `clusterd-manual-info-cm`.
        >- Within 15 seconds after ClusterD starts, do not modify `clusterd-manual-info-cm` temporarily to avoid data errors.

## Parameter Plane Network Fault<a name="ZH-CN_TOPIC_0000002479226486"></a>

### Bus Device Fault<a name="ZH-CN_TOPIC_0000002511346423"></a>

#### Configuration File Description<a name="ZH-CN_TOPIC_0000002511346513"></a>

When performing hierarchical processing for different levels of **bus device** faults, Ascend Device Plugin obtains the fault code of the current fault and processes the fault according to the fault level configured for the fault code in `SwitchFaultCode.json`. `SwitchFaultCode.json` is a system configuration file. Do not modify it arbitrarily unless you have special requirements. If you need to modify the fault level corresponding to a fault code, you can do so through the `mindx-dl-fault-config` file created from `faultCode.json` and `SwitchFaultCode.json`.

>[!NOTE]
>Only Atlas A3 training series products have **bus devices**, and the fault codes for such devices can be viewed in the `SwitchFaultCode.json` file.

**Fault Levels in SwitchFaultCode.json<a name="section681495612012"></a>**

Resumable training supports hierarchical processing for different levels of **bus device** faults. If you need to modify the fault level of a fault code, see [(Optional) Configuring Bus Device Fault Levels](#optional-configuring-bus-device-fault-levels) for operation instructions.

After Ascend Device Plugin obtains the fault code from the driver, it classifies the fault into several levels based on the impact of the fault code on the device and service, and performs corresponding rescheduling processing. For details, see [Table. Fault levels and handling policies](../../api/ascend_device_plugin.md#custom-unifiedbus-device-faults).

#### (Optional) Configuring Bus Device Fault Levels<a name="ZH-CN_TOPIC_0000002511426433"></a>

When building the Ascend Device Plugin image, the fault level configuration file `SwitchFaultCode.json` is built into the image. When Ascend Device Plugin starts, it reads the default configuration of this file as the basis for current fault handling.

If you want to customize the fault level or graceful fault tolerance related configuration, create a ConfigMap file (`mindx-dl-fault-config`) in the cluster.

- If `mindx-dl-fault-config` exists in the cluster when Ascend Device Plugin starts, Ascend Device Plugin will preferentially use the content configured in the existing `mindx-dl-fault-config` as the basis for current fault handling.
- If `mindx-dl-fault-config` exists in the cluster after Ascend Device Plugin is reinstalled, the default `SwitchFaultCode.json` of Ascend Device Plugin will not take effect, and the existing `mindx-dl-fault-config` in the cluster will be used.
- If `mindx-dl-fault-config` exists in the cluster after Ascend Device Plugin is reinstalled and the `SwitchFaultCode.json` field exists in this ConfigMap, the default `SwitchFaultCode.json` of Ascend Device Plugin will not take effect, and the existing `mindx-dl-fault-config` in the cluster will be used.
- If you want to use the default `SwitchFaultCode.json` configuration, you can delete `mindx-dl-fault-config` so that Ascend Device Plugin reads the default `SwitchFaultCode.json` file.
- If there are issues such as format errors in the ConfigMap file content, Ascend Device Plugin will read the content of the built-in ConfigMap file in the image by default as the basis for current fault handling.

**Using SwitchFaultCode.json to Configure Fault Levels<a name="section067783615137"></a>**

Take the bus device fault code `[0x00f1ff09,155913,cpu,na]` as an example. This fault code consists of four parts: alarm ID, fault ID, peer device type, and port number, as shown in [Table 1 Fault code description](#zh-cn_topic_0000002007978080_table167355241939).

**Table 1**  Fault code description

<a name="zh-cn_topic_0000002007978080_table167355241939"></a>

|Parameter|Description|Value|
|--|--|--|
|Alarm ID|In the above example, the alarm ID is 0x00f1ff09.|Values in-band and out-of-band must be consistent.|
|Fault ID|In the above example, the fault ID is 155913.|Values in-band and out-of-band must be consistent.|
|Peer device type|The peer device type corresponding to this fault. In the above example, the peer device type is cpu.|<ul><li>na: This fault is a chip fault and does not involve a peer device.</li><li>cpu: The peer device corresponding to this fault is a CPU.</li><li>npu: The peer device corresponding to this fault is an NPU.</li><li>L2: The peer device corresponding to this fault is an L2.</li></ul>|
|Port number|In the above example, the port number is na.|The value can only be na.|

The following is an example of changing the handling policy for the current fault from `NotHandleFaultCodes` (no handling required) to `SeparateFaultCodes` (isolate the chip and perform job rescheduling).

1. Log in to the environment and go to the decompression directory of Ascend Device Plugin.
2. Run the following command to check whether `mindx-dl-fault-config` has been created based on the `SwitchFaultCode.json` file.

    ```shell
    kubectl describe cm -n kube-system mindx-dl-fault-config
    ```

    - If `mindx-dl-fault-config` exists and contains the relevant fields of `SwitchFaultCode.json`, perform [Step 4](#zh-cn_topic_0000002007978080_li1014819812423) to edit the file.
    - If `mindx-dl-fault-config` exists but does not contain the relevant fields of `SwitchFaultCode.json`, save the content of `mindx-dl-fault-config` first, then delete the `mindx-dl-fault-config` file, and then perform [Step 3](#zh-cn_topic_0000002007978080_li14147485427) to create the file.
    - If `mindx-dl-fault-config` does not exist, perform [Step 3](#zh-cn_topic_0000002007978080_li14147485427) to create the file.

3. <a name="zh-cn_topic_0000002007978080_li14147485427"></a>Run the following command to create the `mindx-dl-fault-config` required for dynamic fault code configuration.

    ```shell
    kubectl create cm mindx-dl-fault-config -n kube-system  --from-file=./faultCode.json --from-file=./SwitchFaultCode.json --from-literal="PollInterval=300"
    ```

    Command output:

    ```ColdFusion
    configmap/mindx-dl-fault-config created
    ```

    **Table 2** Parameter description

    <a name="zh-cn_topic_0000002007978080_table14147138184211"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002007978080_row1814716812426"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002007978080_p141471483423"><a name="zh-cn_topic_0000002007978080_p141471483423"></a><a name="zh-cn_topic_0000002007978080_p141471483423"></a>Parameter Name</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002007978080_p101477811428"><a name="zh-cn_topic_0000002007978080_p101477811428"></a><a name="zh-cn_topic_0000002007978080_p101477811428"></a>Required</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002007978080_p1014718154210"><a name="zh-cn_topic_0000002007978080_p1014718154210"></a><a name="zh-cn_topic_0000002007978080_p1014718154210"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002007978080_row1514810811424"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002007978080_p714817819421"><a name="zh-cn_topic_0000002007978080_p714817819421"></a><a name="zh-cn_topic_0000002007978080_p714817819421"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002007978080_p201488804220"><a name="zh-cn_topic_0000002007978080_p201488804220"></a><a name="zh-cn_topic_0000002007978080_p201488804220"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002007978080_p161481689426"><a name="zh-cn_topic_0000002007978080_p161481689426"></a><a name="zh-cn_topic_0000002007978080_p161481689426"></a>The name of the <span id="zh-cn_topic_0000002007978080_ph214819813425"><a name="zh-cn_topic_0000002007978080_ph214819813425"></a><a name="zh-cn_topic_0000002007978080_ph214819813425"></a>ConfigMap</span> file required for dynamic fault code configuration. This file name cannot be modified.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002007978080_row814819819422"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002007978080_p101481589424"><a name="zh-cn_topic_0000002007978080_p101481589424"></a><a name="zh-cn_topic_0000002007978080_p101481589424"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002007978080_p214814819427"><a name="zh-cn_topic_0000002007978080_p214814819427"></a><a name="zh-cn_topic_0000002007978080_p214814819427"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002007978080_p1614814815424"><a name="zh-cn_topic_0000002007978080_p1614814815424"></a><a name="zh-cn_topic_0000002007978080_p1614814815424"></a>The namespace where mindx-dl-fault-config resides. This namespace name cannot be modified.</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002007978080_row1714868114215"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002007978080_p182611591222"><a name="zh-cn_topic_0000002007978080_p182611591222"></a><a name="zh-cn_topic_0000002007978080_p182611591222"></a>SwitchFaultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002007978080_p1314868184217"><a name="zh-cn_topic_0000002007978080_p1314868184217"></a><a name="zh-cn_topic_0000002007978080_p1314868184217"></a>Yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002007978080_p17148118174218"><a name="zh-cn_topic_0000002007978080_p17148118174218"></a><a name="zh-cn_topic_0000002007978080_p17148118174218"></a>Used to store fault codes. Must be consistent with the SwitchFaultCode.json file name.</p>
    </td>
    </tr>
    </tbody>
    </table>

4. <a name="zh-cn_topic_0000002007978080_li1014819812423"></a>Run the following command to edit the `mindx-dl-fault-config` file.

    ```shell
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

5. In the `mindx-dl-fault-config` file, locate the fault code `[0x00f1ff09,155913,cpu,na]`.

    ```json
    Data
    ====
    SwitchFaultCode.json:
    ----
    {"NotHandleFaultCodes":[0x00f1ff09,155913,cpu,na],
    ...
    ```

6. Delete the fault code from `NotHandleFaultCodes` and add it to `SeparateFaultCodes`.

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

7. After the modification is complete, press the `Esc`, enter `:wq!` to save and exit.
8. After the `mindx-dl-fault-config` file update takes effect (`PollInterval` defaulted to `300s` if not specified), check whether the operation is successful.
    1. Run the following command to query the log name of Ascend Device Plugin.

        ```shell
        kubectl get pods -A | grep ascend-device-plugin
        ```

        The echo example is as follows:

        ```ColdFusion
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2. Query the log information of Ascend Device Plugin by using the queried component log name.

        ```shell
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        If the log displays "load switch fault code from configmap success", it indicates that the manual fault code configuration is successful.

### Associated Faults<a name="ZH-CN_TOPIC_0000002511426403"></a>

#### Configuration File Description<a name="ZH-CN_TOPIC_0000002479386560"></a>

For associated faults (special faults may trigger other related faults), it is necessary to ignore the included accompanying faults. ClusterD can detect special faults and perform special processing on faulty jobs according to the associated fault policies configured in the `relationFaultCustomization.json` and `faultDuration.json` files.

`relationFaultCustomization.json` and `faultDuration.json` are system configuration files. Do not modify them arbitrarily unless you have special requirements.

**Table 1**  relationFaultCustomization file description

<a name="zh-cn_topic_0000002157130117_table5148194813113"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002157130117_row1614914482114"><th class="cellrowborder" valign="top" width="13.701370137013702%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002157130117_p278365710116"><a name="zh-cn_topic_0000002157130117_p278365710116"></a><a name="zh-cn_topic_0000002157130117_p278365710116"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="69.05690569056905%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002157130117_p127832571915"><a name="zh-cn_topic_0000002157130117_p127832571915"></a><a name="zh-cn_topic_0000002157130117_p127832571915"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="17.241724172417243%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002157130117_p47831857912"><a name="zh-cn_topic_0000002157130117_p47831857912"></a><a name="zh-cn_topic_0000002157130117_p47831857912"></a>Value</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002157130117_row1514912481715"><td class="cellrowborder" valign="top" width="13.701370137013702%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p14783115717117"><a name="zh-cn_topic_0000002157130117_p14783115717117"></a><a name="zh-cn_topic_0000002157130117_p14783115717117"></a>TriggerFault</p>
</td>
<td class="cellrowborder" valign="top" width="69.05690569056905%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p1878313577120"><a name="zh-cn_topic_0000002157130117_p1878313577120"></a><a name="zh-cn_topic_0000002157130117_p1878313577120"></a>Accompanying fault code. Currently supports fault codes configured in faultCode.json and SwitchFaultCode.json.</p>
</td>
<td class="cellrowborder" valign="top" width="17.241724172417243%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p117831557615"><a name="zh-cn_topic_0000002157130117_p117831557615"></a><a name="zh-cn_topic_0000002157130117_p117831557615"></a>String</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row1714944814110"><td class="cellrowborder" valign="top" width="13.701370137013702%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p6783657411"><a name="zh-cn_topic_0000002157130117_p6783657411"></a><a name="zh-cn_topic_0000002157130117_p6783657411"></a>RelationFaults</p>
</td>
<td class="cellrowborder" valign="top" width="69.05690569056905%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p278411575113"><a name="zh-cn_topic_0000002157130117_p278411575113"></a><a name="zh-cn_topic_0000002157130117_p278411575113"></a>List of faults to be associated, which can be one or more fault codes. Currently supports fault codes configured in faultCode.json and SwitchFaultCode.json.</p>
</td>
<td class="cellrowborder" valign="top" width="17.241724172417243%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p178414571018"><a name="zh-cn_topic_0000002157130117_p178414571018"></a><a name="zh-cn_topic_0000002157130117_p178414571018"></a>String list</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row111493481216"><td class="cellrowborder" valign="top" width="13.701370137013702%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p578414571818"><a name="zh-cn_topic_0000002157130117_p578414571818"></a><a name="zh-cn_topic_0000002157130117_p578414571818"></a>FaultStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="69.05690569056905%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p178405710112"><a name="zh-cn_topic_0000002157130117_p178405710112"></a><a name="zh-cn_topic_0000002157130117_p178405710112"></a>Handling policy for the corresponding job when the associated fault is successfully matched.</p>
<a name="zh-cn_topic_0000002157130117_ul17849570118"></a><a name="zh-cn_topic_0000002157130117_ul17849570118"></a><ul id="zh-cn_topic_0000002157130117_ul17849570118"><li>Separate: job isolation</li><li>SubHealth: job sub-health</li></ul>
</td>
<td class="cellrowborder" valign="top" width="17.241724172417243%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p1378413577119"><a name="zh-cn_topic_0000002157130117_p1378413577119"></a><a name="zh-cn_topic_0000002157130117_p1378413577119"></a>String</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row84116191226"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p114317681616"><a name="zh-cn_topic_0000002157130117_p114317681616"></a><a name="zh-cn_topic_0000002157130117_p114317681616"></a>Note:</p>
<p id="zh-cn_topic_0000002157130117_p47413216213"><a name="zh-cn_topic_0000002157130117_p47413216213"></a><a name="zh-cn_topic_0000002157130117_p47413216213"></a>When the configured RelationFaults occur on a device, <span id="zh-cn_topic_0000002157130117_ph12291515161616"><a name="zh-cn_topic_0000002157130117_ph12291515161616"></a><a name="zh-cn_topic_0000002157130117_ph12291515161616"></a>ClusterD</span> will add the corresponding faults to the pending fault code queue. Within the configured TimeOutInterval, if a fault corresponding to TriggerFault occurs, the job will be processed according to the user-configured FaultStrategy policy. If the configured TimeOutInterval is exceeded, bus device faults will be processed using SubHealth policy; chip faults or parameter plane network faults will be ignored.</p>
</td>
</tr>
</tbody>
</table>

**Table 2**  faultDuration.json file description

<a name="zh-cn_topic_0000002157130117_table1484617498414"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002157130117_row1284615492415"><th class="cellrowborder" valign="top" width="13.36133613361336%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002157130117_p116699222514"><a name="zh-cn_topic_0000002157130117_p116699222514"></a><a name="zh-cn_topic_0000002157130117_p116699222514"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="70.36703670367037%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002157130117_p56691922055"><a name="zh-cn_topic_0000002157130117_p56691922055"></a><a name="zh-cn_topic_0000002157130117_p56691922055"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="16.271627162716275%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002157130117_p466911221257"><a name="zh-cn_topic_0000002157130117_p466911221257"></a><a name="zh-cn_topic_0000002157130117_p466911221257"></a>Value</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002157130117_row084615491413"><td class="cellrowborder" valign="top" width="13.36133613361336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p066920221954"><a name="zh-cn_topic_0000002157130117_p066920221954"></a><a name="zh-cn_topic_0000002157130117_p066920221954"></a>FaultCode</p>
</td>
<td class="cellrowborder" valign="top" width="70.36703670367037%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p96702227514"><a name="zh-cn_topic_0000002157130117_p96702227514"></a><a name="zh-cn_topic_0000002157130117_p96702227514"></a>Fault code. Currently supports fault codes configured in faultCode.json and SwitchFaultCode.json.</p>
</td>
<td class="cellrowborder" valign="top" width="16.271627162716275%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p56701922954"><a name="zh-cn_topic_0000002157130117_p56701922954"></a><a name="zh-cn_topic_0000002157130117_p56701922954"></a>String</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row18467491043"><td class="cellrowborder" valign="top" width="13.36133613361336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p167022212517"><a name="zh-cn_topic_0000002157130117_p167022212517"></a><a name="zh-cn_topic_0000002157130117_p167022212517"></a>FaultType</p>
</td>
<td class="cellrowborder" valign="top" width="70.36703670367037%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p667020225515"><a name="zh-cn_topic_0000002157130117_p667020225515"></a><a name="zh-cn_topic_0000002157130117_p667020225515"></a>Fault type:</p>
<a name="zh-cn_topic_0000002157130117_ul1367017221559"></a><a name="zh-cn_topic_0000002157130117_ul1367017221559"></a><ul id="zh-cn_topic_0000002157130117_ul1367017221559"><li>faultDevice: chip fault or parameter plane network fault</li><li>faultSwitch: bus device fault</li></ul>
</td>
<td class="cellrowborder" valign="top" width="16.271627162716275%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p967010221359"><a name="zh-cn_topic_0000002157130117_p967010221359"></a><a name="zh-cn_topic_0000002157130117_p967010221359"></a>String</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002157130117_row208478499416"><td class="cellrowborder" valign="top" width="13.36133613361336%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002157130117_p16713225511"><a name="zh-cn_topic_0000002157130117_p16713225511"></a><a name="zh-cn_topic_0000002157130117_p16713225511"></a>TimeOutInterval</p>
</td>
<td class="cellrowborder" valign="top" width="70.36703670367037%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002157130117_p36711221159"><a name="zh-cn_topic_0000002157130117_p36711221159"></a><a name="zh-cn_topic_0000002157130117_p36711221159"></a>Maximum time for which the fault code can be associated. Unit: seconds.</p>
</td>
<td class="cellrowborder" valign="top" width="16.271627162716275%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002157130117_p186718221450"><a name="zh-cn_topic_0000002157130117_p186718221450"></a><a name="zh-cn_topic_0000002157130117_p186718221450"></a>Integer</p>
</td>
</tr>
</tbody>
</table>

#### (Optional) Configure the Handling Policy for Associated Faults<a name="ZH-CN_TOPIC_0000002479226478"></a>

When the ClusterD image is built, the two configuration files for associated faults are built into the image. When ClusterD starts, it reads the default configurations of these two files as the basis for current fault handling.

If you want to customize the associated fault codes and corresponding handling policies, modify the corresponding `relationFaultCustomization.json` and `faultDuration.json` files when creating the ClusterD image.

**Procedure<a name="zh-cn_topic_0000002157048501_section2086912531189"></a>**

Take `RelationFaults` with code `81078603` and `TriggerFault` with code `8C1F8609` as an example. If the fault `81078603` occurs, the fault `8C1F8609` should be ignored when it appears within the subsequent 60 seconds, and the job where the fault `81078603` occurred should be isolated. You can manually configure the handling policy for associated faults to `Separate`.

1. Log in to the environment and go to the directory where ClusterD is decompressed.
2. Run the `vi relationFaultCustomization.json` command to edit the configuration file.

    ```shell
    vi relationFaultCustomization.json
    ```

    Associate the two faults. After modification, press `Esc` and enter `:wq!` to save and exit.

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

3. Run the `vi faultDuration.json` command to edit the configuration file.

    ```shell
    vi faultDuration.json
    ```

    Configure fault types, fault association time, etc. After modification, press `Esc` and enter `:wq!` to save and exit.

    ```json
    …
      {
        "FaultCode": "81078603",
        "FaultType": "faultDevice",
        "TimeOutInterval": 60
      }
    …
    ```

## Common Faults<a name="ZH-CN_TOPIC_0000002479386564"></a>

### Configuration File Description<a name="ZH-CN_TOPIC_0000002511346487"></a>

Resumable training performs hierarchical processing for different levels of common faults. ClusterD obtains the fault code of the current fault and processes the fault accordingly based on the fault level configured for the fault code in the `publicFaultConfiguration.json` file. In special cases, if ClusterD receives an unrecognized fault code (not saved in the configuration file), it will discard this fault.

[publicFaultConfiguration.json](#zh-cn_topic_0000002181110120_table8202741102717) is the system configuration file for common faults. Do not modify it arbitrarily unless you have special requirements. If you need to modify the level and sender of common faults, you can do so by writing a custom configuration file named `publicCustomization.json` to `/user1/mindx-dl/clusterd`. The path to this file is configurable. The configuration method is as follows:

>[!NOTE]
>
>- `publicCustomization.json` is located at `/user1/mindx-dl/clusterd` inside the container. Modification and soft links are not supported. The default host path is `/user1/mindx-dl/clusterd`.
>- You can configure the host path based on actual conditions: Modify the host mount path of the volume named `config-clusterd` in the ClusterD startup YAML.
>- In a multi-master scenario, it is recommended to synchronize the latest `publicCustomization.json` file on each master node. This prevents the issue of losing the custom fault configuration file if ClusterD is rescheduled to another master node after a restart.

**Table 1**  Fault levels and handling policies

<a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_table169151711124319"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row19916131120434"><th class="cellrowborder" valign="top" width="15.09499941718149%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1291621144314"></a>Fault Level</p>
</th>
<th class="cellrowborder" valign="top" width="42.54575125305979%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1694364414313"></a>Fault Handling Policy</p>
</th>
<th class="cellrowborder" valign="top" width="42.35924932975871%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002181110120_p2218314171716"><a name="zh-cn_topic_0000002181110120_p2218314171716"></a><a name="zh-cn_topic_0000002181110120_p2218314171716"></a>Rescheduling Handling</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row6916711144312"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p1240123404316"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119431441431"></a>No handling required</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p119435448430"></a>Not handled for now</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row1991661104316"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p18618151202216"></a>Unrecoverable, chip isolation required</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p12165431710"><a name="zh-cn_topic_0000002181110120_p12165431710"></a><a name="zh-cn_topic_0000002181110120_p12165431710"></a>Isolate the chip and perform job rescheduling.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_row191716112431"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521329_p172401834194316"></a>SubHealthFault</p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p1354813311915"></a>Handling is based on the value of the subHealthyStrategy parameter configured in the job YAML. For details, see <a href="../../api/">YAML Configuration Description</a>.</p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p3352524125220"></a>When a chip experiences a sub-health fault, handling must follow the policy in <a href="./06_configuring_the_job_yaml_file.md">Job YAML Configuration</a>.</p>
<div class="note" id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_note7936204710536"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_note7936204710536"></a><div class="notebody"><p id="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"></a><a name="zh-cn_topic_0000002181110120_zh-cn_topic_0000002171521445_p15222114115810"></a>If the chip subsequently experiences faults of other levels, the SubHealthFault handling policy does not affect the handling of those other-level faults.</p>
</div></div>
</td>
</tr>
<tr id="row16800523414"><td class="cellrowborder" valign="top" width="15.09499941718149%" headers="mcps1.2.4.1.1 "><p id="p88011823817"><a name="p88011823817"></a><a name="p88011823817"></a><span id="ph1339214581915"><a name="ph1339214581915"></a><a name="ph1339214581915"></a>PreSeparateNPU</span></p>
</td>
<td class="cellrowborder" valign="top" width="42.54575125305979%" headers="mcps1.2.4.1.2 "><p id="p980117231413"><a name="p980117231413"></a><a name="p980117231413"></a><span id="ph739245817113"><a name="ph739245817113"></a><a name="ph739245817113"></a>No immediate service impact; subsequent jobs will no longer be scheduled to this chip.</span></p>
</td>
<td class="cellrowborder" valign="top" width="42.35924932975871%" headers="mcps1.2.4.1.3 "><p id="p1280114235116"><a name="p1280114235116"></a><a name="p1280114235116"></a><span id="ph3392758212"><a name="ph3392758212"></a><a name="ph3392758212"></a>Pre-isolate the chip.</span></p>
</td>
</tr>
</tbody>
</table>

**Table 2**  publicFaultConfiguration.json field description

<a name="zh-cn_topic_0000002181110120_table8202741102717"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row18202164117272"><th class="cellrowborder" valign="top" width="28.93%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002181110120_p1120213413271"><a name="zh-cn_topic_0000002181110120_p1120213413271"></a><a name="zh-cn_topic_0000002181110120_p1120213413271"></a>Parameter Name</p>
</th>
<th class="cellrowborder" valign="top" width="71.07%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002181110120_p22024417279"><a name="zh-cn_topic_0000002181110120_p22024417279"></a><a name="zh-cn_topic_0000002181110120_p22024417279"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row172028412278"><td class="cellrowborder" valign="top" width="28.93%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p1220219412279"><a name="zh-cn_topic_0000002181110120_p1220219412279"></a><a name="zh-cn_topic_0000002181110120_p1220219412279"></a><a href="#zh-cn_topic_0000002181110120_table1689274753416">publicFaultCode</a></p>
</td>
<td class="cellrowborder" valign="top" width="71.07%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p220284110271"><a name="zh-cn_topic_0000002181110120_p220284110271"></a><a name="zh-cn_topic_0000002181110120_p220284110271"></a>Configuration related to common fault codes.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row14606121802219"><td class="cellrowborder" valign="top" width="28.93%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p1760617182224"><a name="zh-cn_topic_0000002181110120_p1760617182224"></a><a name="zh-cn_topic_0000002181110120_p1760617182224"></a>publicFaultResource</p>
</td>
<td class="cellrowborder" valign="top" width="71.07%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p1606118102218"><a name="zh-cn_topic_0000002181110120_p1606118102218"></a><a name="zh-cn_topic_0000002181110120_p1606118102218"></a>Configuration of the common fault sender.</p>
</td>
</tr>
</tbody>
</table>

**Table 3** publicFaultCode field description

<a name="zh-cn_topic_0000002181110120_table1689274753416"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row16892144733413"><th class="cellrowborder" valign="top" width="28.849999999999998%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000002181110120_p689264723412"><a name="zh-cn_topic_0000002181110120_p689264723412"></a><a name="zh-cn_topic_0000002181110120_p689264723412"></a>Parameter Name</p>
</th>
<th class="cellrowborder" valign="top" width="71.15%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000002181110120_p889274783418"><a name="zh-cn_topic_0000002181110120_p889274783418"></a><a name="zh-cn_topic_0000002181110120_p889274783418"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row28921647103410"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p48921847143412"><a name="zh-cn_topic_0000002181110120_p48921847143412"></a><a name="zh-cn_topic_0000002181110120_p48921847143412"></a>NotHandleFaultCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p58921747183416"><a name="zh-cn_topic_0000002181110120_p58921747183416"></a><a name="zh-cn_topic_0000002181110120_p58921747183416"></a>Fault codes with the fault level NotHandleFault (no handling required).</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row989224719346"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p118928476343"><a name="zh-cn_topic_0000002181110120_p118928476343"></a><a name="zh-cn_topic_0000002181110120_p118928476343"></a>SubHealthFaultCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p17892947113410"><a name="zh-cn_topic_0000002181110120_p17892947113410"></a><a name="zh-cn_topic_0000002181110120_p17892947113410"></a>Fault codes with the fault level SubHealthFault.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row289264713349"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000002181110120_p38921547193418"><a name="zh-cn_topic_0000002181110120_p38921547193418"></a><a name="zh-cn_topic_0000002181110120_p38921547193418"></a>SeparateNPUCodes</p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000002181110120_p689274714341"><a name="zh-cn_topic_0000002181110120_p689274714341"></a><a name="zh-cn_topic_0000002181110120_p689274714341"></a>Fault codes with the fault level SeparateNPU (unrecoverable, chip isolation required).</p>
</td>
</tr>
<tr id="row107385344217"><td class="cellrowborder" valign="top" width="28.849999999999998%" headers="mcps1.2.3.1.1 "><p id="p187397341724"><a name="p187397341724"></a><a name="p187397341724"></a><span id="ph791817016319"><a name="ph791817016319"></a><a name="ph791817016319"></a>PreSeparateNPUCodes</span></p>
</td>
<td class="cellrowborder" valign="top" width="71.15%" headers="mcps1.2.3.1.2 "><p id="p15739113415210"><a name="p15739113415210"></a><a name="p15739113415210"></a><span id="ph8918120234"><a name="ph8918120234"></a><a name="ph8918120234"></a>Fault codes with the fault level </span><span id="ph491890639"><a name="ph491890639"></a><a name="ph491890639"></a>PreSeparateNPU</span><span id="ph6918601336"><a name="ph6918601336"></a><a name="ph6918601336"></a> (no immediate service impact, but no further jobs will be scheduled to this chip).</span></p>
</td>
</tr>
</tbody>
</table>

**Fault Code Description<a name="zh-cn_topic_0000002181110120_section1440314273418"></a>**

The fault code for common faults is 9 digits, as described below.

**Table 4** Fault code description

<a name="table1237891465117"></a>
<table><thead align="left"><tr id="row1137891413516"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="p1937816143519"><a name="p1937816143519"></a><a name="p1937816143519"></a>Bit</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="p837812144514"><a name="p837812144514"></a><a name="p837812144514"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="p14378201455110"><a name="p14378201455110"></a><a name="p14378201455110"></a>Value</p>
</th>
</tr>
</thead>
<tbody><tr id="row1137861419517"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p123782149514"><a name="p123782149514"></a><a name="p123782149514"></a>1</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p15378914185120"><a name="p15378914185120"></a><a name="p15378914185120"></a>Fault type</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p10378161419517"><a name="p10378161419517"></a><a name="p10378161419517"></a>0: chip fault</p>
<p id="p037871414515"><a name="p037871414515"></a><a name="p037871414515"></a>1: node fault</p>
<p id="p33781414125113"><a name="p33781414125113"></a><a name="p33781414125113"></a>2: network fault</p>
<p id="p10379101414516"><a name="p10379101414516"></a><a name="p10379101414516"></a>3: storage fault</p>
</td>
</tr>
<tr id="row337901415519"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p2379181475111"><a name="p2379181475111"></a><a name="p2379181475111"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p133796146513"><a name="p133796146513"></a><a name="p133796146513"></a>Default fault level</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p103791114185115"><a name="p103791114185115"></a><a name="p103791114185115"></a>0: NotHandleFault</p>
<p id="p193791214175112"><a name="p193791214175112"></a><a name="p193791214175112"></a>1: SubHealthFault</p>
<p id="p737991475119"><a name="p737991475119"></a><a name="p737991475119"></a>2: SeparateNPU</p>
</td>
</tr>
<tr id="row1737917147519"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p133793145514"><a name="p133793145514"></a><a name="p133793145514"></a>3, 4</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1137901435119"><a name="p1137901435119"></a><a name="p1137901435119"></a>Reserved extension bits</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p43795142516"><a name="p43795142516"></a><a name="p43795142516"></a>Temporarily 00</p>
</td>
</tr>
<tr id="row1337961495114"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17379121416515"><a name="p17379121416515"></a><a name="p17379121416515"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p17379141465112"><a name="p17379141465112"></a><a name="p17379141465112"></a>Whether the fault code in bits 6-9 is user-defined to avoid conflicts</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1237917146517"><a name="p1237917146517"></a><a name="p1237917146517"></a>0: defined in the release package</p>
<p id="p12379191418513"><a name="p12379191418513"></a><a name="p12379191418513"></a>1: user-defined</p>
</td>
</tr>
<tr id="row1937911425114"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p12379161465115"><a name="p12379161465115"></a><a name="p12379161465115"></a>6-9</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1437931425115"><a name="p1437931425115"></a><a name="p1437931425115"></a>Specific decimal fault code</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p8379121413512"><a name="p8379121413512"></a><a name="p8379121413512"></a>Example: 1001</p>
</td>
</tr>
<tr id="row6379214165114"><td class="cellrowborder" colspan="3" valign="top" headers="mcps1.2.4.1.1 mcps1.2.4.1.2 mcps1.2.4.1.3 "><p id="p1137911410515"><a name="p1137911410515"></a><a name="p1137911410515"></a>Examples are as follows:</p>
<p id="p1837941416513"><a name="p1837941416513"></a><a name="p1837941416513"></a>0100 01001: chip fault, SubHealthFault, defined in the release package, fault 1001.</p>
<p id="p1037911455117"><a name="p1037911455117"></a><a name="p1037911455117"></a>1000 11002: node fault, NotHandleFault, user-defined, fault 1002.</p>
<p id="p8379181455115"><a name="p8379181455115"></a><a name="p8379181455115"></a>2200 01003: network fault, SeparateNPU, defined in the release package, fault 1003.</p>
</td>
</tr>
</tbody>
</table>

**Known Common Faults<a name="zh-cn_topic_0000002181110120_section4960201383813"></a>**

**Table 5** Known common faults

<a name="zh-cn_topic_0000002181110120_table31451934163811"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002181110120_row514523493819"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002181110120_p1114523420389"><a name="zh-cn_topic_0000002181110120_p1114523420389"></a><a name="zh-cn_topic_0000002181110120_p1114523420389"></a>Fault Code</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002181110120_p9145143412387"><a name="zh-cn_topic_0000002181110120_p9145143412387"></a><a name="zh-cn_topic_0000002181110120_p9145143412387"></a>Fault Description</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002181110120_p15145193413388"><a name="zh-cn_topic_0000002181110120_p15145193413388"></a><a name="zh-cn_topic_0000002181110120_p15145193413388"></a>Default Fault Level</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002181110120_row1514593415388"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_p181451134193811"><a name="zh-cn_topic_0000002181110120_p181451134193811"></a><a name="zh-cn_topic_0000002181110120_p181451134193811"></a>010001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_p814593412386"><a name="zh-cn_topic_0000002181110120_p814593412386"></a><a name="zh-cn_topic_0000002181110120_p814593412386"></a>Optical link contamination (chip fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p414533483811"><a name="zh-cn_topic_0000002181110120_p414533483811"></a><a name="zh-cn_topic_0000002181110120_p414533483811"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row175241157181818"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p580896101918"><a name="p580896101918"></a><a name="p580896101918"></a>210001007</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p15808166121915"><a name="p15808166121915"></a><a name="p15808166121915"></a>Optical link contamination (network fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1180917617197"><a name="p1180917617197"></a><a name="p1180917617197"></a>SubHealthFault</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002181110120_row131782214434"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002181110120_p41752216438"><a name="zh-cn_topic_0000002181110120_p41752216438"></a><a name="zh-cn_topic_0000002181110120_p41752216438"></a>220001001</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002181110120_p1171822134316"><a name="zh-cn_topic_0000002181110120_p1171822134316"></a><a name="zh-cn_topic_0000002181110120_p1171822134316"></a>NPU card <span id="ph17233131243911"><a name="ph17233131243911"></a><a name="ph17233131243911"></a>HCCS</span> network fault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002181110120_p1566710511444"><a name="zh-cn_topic_0000002181110120_p1566710511444"></a><a name="zh-cn_topic_0000002181110120_p1566710511444"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row192881812184715"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3289131210473"><a name="p3289131210473"></a><a name="p3289131210473"></a>010001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1628951244719"><a name="p1628951244719"></a><a name="p1628951244719"></a>Optical link loosening (chip fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p828971254715"><a name="p828971254715"></a><a name="p828971254715"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row38601828161910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p6168163671911"><a name="p6168163671911"></a><a name="p6168163671911"></a>210001008</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p316816364194"><a name="p316816364194"></a><a name="p316816364194"></a>Optical link loosening (network fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2168436121911"><a name="p2168436121911"></a><a name="p2168436121911"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row172051674711"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p127201168472"><a name="p127201168472"></a><a name="p127201168472"></a>310001005</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1572071644717"><a name="p1572071644717"></a><a name="p1572071644717"></a>DPC client failure</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p141495491488"><a name="p141495491488"></a><a name="p141495491488"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row4720816104713"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17720131674712"><a name="p17720131674712"></a><a name="p17720131674712"></a>200001006</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1972020169475"><a name="p1972020169475"></a><a name="p1972020169475"></a>Suspected optical link sub-health</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1572061684719"><a name="p1572061684719"></a><a name="p1572061684719"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row191121122184715"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p7112152234711"><a name="p7112152234711"></a><a name="p7112152234711"></a>210001009</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p10112152210476"><a name="p10112152210476"></a><a name="p10112152210476"></a>Optical module component sub-health</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2011213229474"><a name="p2011213229474"></a><a name="p2011213229474"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row19731102610435"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1324180124413"><a name="p1324180124413"></a><a name="p1324180124413"></a>220001002</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p13241019443"><a name="p13241019443"></a><a name="p13241019443"></a>Non-existent backup rack resources used for scheduling in a back SuperPoD scenario</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p12161558145818"><a name="p12161558145818"></a><a name="p12161558145818"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row13731626174317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3241309446"><a name="p3241309446"></a><a name="p3241309446"></a>220001003</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1424190104412"><a name="p1424190104412"></a><a name="p1424190104412"></a>Resource port fault in a backup rack</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p17362234428"><a name="p17362234428"></a><a name="p17362234428"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row127318268434"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p52416011444"><a name="p52416011444"></a><a name="p52416011444"></a>220001004</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p124800442"><a name="p124800442"></a><a name="p124800442"></a>Job ID occupancy conflict in a backup rack</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p191615588586"><a name="p191615588586"></a><a name="p191615588586"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row20731826154318"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5241103443"><a name="p5241103443"></a><a name="p5241103443"></a>220001005</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p824705444"><a name="p824705444"></a><a name="p824705444"></a>NetMind failure</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1016110586589"><a name="p1016110586589"></a><a name="p1016110586589"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row673142624317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p162412019440"><a name="p162412019440"></a><a name="p162412019440"></a>220001006</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p10241109444"><a name="p10241109444"></a><a name="p10241109444"></a>Suspected partial failure of backup rack link port</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p916119583580"><a name="p916119583580"></a><a name="p916119583580"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row673116264438"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p9247064419"><a name="p9247064419"></a><a name="p9247064419"></a>220001007</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p6249018447"><a name="p6249018447"></a><a name="p6249018447"></a>Optical link adjustment failure</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p132215501211"><a name="p132215501211"></a><a name="p132215501211"></a>SeparateNPU</p>
</td>
</tr>
<tr id="row8926105693315"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1792695643311"><a name="p1792695643311"></a><a name="p1792695643311"></a>200001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p18926165614336"><a name="p18926165614336"></a><a name="p18926165614336"></a>Slow network generated/recovered within a node (slow network fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p24091634153411"><a name="p24091634153411"></a><a name="p24091634153411"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row10526205273417"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p3526052153416"><a name="p3526052153416"></a><a name="p3526052153416"></a>200001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p5804154074418"><a name="p5804154074418"></a><a name="p5804154074418"></a>Slow network generated/recovered between nodes within a SuperPoD (slow network fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p352695212349"><a name="p352695212349"></a><a name="p352695212349"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row663164316353"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p17634437355"><a name="p17634437355"></a><a name="p17634437355"></a>200001012</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p18631743123513"><a name="p18631743123513"></a><a name="p18631743123513"></a>Slow network not caused by chip fault (slow network fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p101021310163612"><a name="p101021310163612"></a><a name="p101021310163612"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row178327182364"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p383221833611"><a name="p383221833611"></a><a name="p383221833611"></a>110001010</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1683231816361"><a name="p1683231816361"></a><a name="p1683231816361"></a>Slow node fault</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p6832131833614"><a name="p6832131833614"></a><a name="p6832131833614"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row1179514189380"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p979511810389"><a name="p979511810389"></a><a name="p979511810389"></a>100001011</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1579521818381"><a name="p1579521818381"></a><a name="p1579521818381"></a>Degradation recovered (slow node fault)</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p27883220394"><a name="p27883220394"></a><a name="p27883220394"></a>NotHandleFault</p>
</td>
</tr>
<tr id="row121732048142813"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p4359165915289"><a name="p4359165915289"></a><a name="p4359165915289"></a>110001020</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1817494822816"><a name="p1817494822816"></a><a name="p1817494822816"></a>Shared storage DPC process exception</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p121741348132817"><a name="p121741348132817"></a><a name="p121741348132817"></a>SubHealthFault</p>
</td>
</tr>
<tr id="row7277115416280"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p5277135492816"><a name="p5277135492816"></a><a name="p5277135492816"></a>110001021</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7277854132820"><a name="p7277854132820"></a><a name="p7277854132820"></a>Insufficient memory for Shared storage DPC</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p182781654132815"><a name="p182781654132815"></a><a name="p182781654132815"></a>SubHealthFault</p>
</td>
</tr>
</tbody>
</table>

### (Optional) Configuring the Level and Sender of Common Faults<a name="ZH-CN_TOPIC_0000002479226494"></a>

When the ClusterD image is created, the fault level configuration file `publicFaultConfiguration.jso`n is built into the image. When ClusterD starts, it reads the default configuration of this file as the basis for current fault handling.

If you want to customize fault levels, create the `/user1/mindx-dl/clusterd/publicCustomization.json` file on the host.

- If this file exists when ClusterD starts, ClusterD will prioritize the content configured in the existing file as the basis for current fault handling.
- If this file exists after ClusterD is reinstalled, the default `publicFaultConfiguration.json` of ClusterD will not take effect, and the existing `publicCustomization.json` file will be used. If you want to use the default configuration of `publicFaultConfiguration.json`, you can delete the existing `publicCustomization.json` file so that ClusterD reads the default `publicFaultConfiguration.json` file.
- If the content of the `publicCustomization.json` file has issues such as format errors, ClusterD will read the content of the built-in `publicFaultConfiguration.json` file in the image by default as the basis for current fault handling.

**Configuring the Level of Common Fault Codes<a name="zh-cn_topic_0000002180950420_section1384121854711"></a>**

Configuring the level of common fault codes is divided into the following two scenarios.

- Adjusting the level of existing fault codes.
- Adding new fault codes and their fault levels.

    The following uses fault code `010001008` as an example to describe how to configure a common fault code level.

1. Log in to the environment and go to the `/user1/mindx-dl/clusterd` directory.
2. Run the `vi publicCustomization.json` command to edit the file. For detailed description of `publicCustomization.json`, see [Table 2](#ZH-CN_TOPIC_0000002511346487).

    >[!NOTE]
    >- After creating the `publicCustomization.json` file, ensure that the file has the read permission for the ClusterD user `hwMindX`. For example, if the user permission is `root`, the file permission is recommended to be set to `644`.
    >- Ensure the security of file permissions. Excessive permissions may pose a security risk.

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

3. After the modification is complete, press `Esc`, enter `:wq!` to save and exit.
4. After a few seconds, the file takes effect. Check whether the operation is successful.

    If the log displays "load fault config from <publicCustomization.json\> success", the manual fault code configuration is successful.

**Configuring the Sender of Common Faults<a name="zh-cn_topic_0000002180950420_section5532327614"></a>**

The following uses the new fault sender XXX as an example to describe the steps for configuring the sender of common fault codes.

1. Log in to the environment and go to the `/user1/mindx-dl/clusterd` directory.
2. Run the `vi publicCustomization.json` command to edit the file. For detailed description of `publicCustomization.json`, see [Table 2](#ZH-CN_TOPIC_0000002511346487).

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

3. After modification, press `Esc` and enter `:wq!` to save and exit.
4. The file takes effect after a few seconds. Check whether the operation is successful.

    If "load fault config from <publicCustomization.json\> success" appears in the log, the configuration is successful.
