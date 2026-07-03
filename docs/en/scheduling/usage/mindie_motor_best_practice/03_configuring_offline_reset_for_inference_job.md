# Configuring Offline Reset for Inference Jobs<a name="ZH-CN_TOPIC_0000002479226442"></a>

Currently, offline reset is only supported for Atlas 800I A2 inference servers and Atlas 800I A3 SuperPoD servers. When this function is enabled, a hot reset operation is performed after a chip fault occurs, restoring the chip to a healthy state.

To enable the offline reset function for MindIE Motor inference jobs, you only need to set the Ascend Device Plugin startup parameter `-hotReset` to `0` or `2`.

**Table 1**  Parameter description

<a name="table173461839165111"></a>

|Parameter|Type|Default Value|Description|
|--|--|--|--|
|-hotReset|int|-1|Device hot reset. When this function is enabled, if a chip fault occurs, Ascend Device Plugin performs a hot reset operation to restore the chip to a healthy state.<ul><li>-1: Disable the chip reset function</li><li>0: Enable the inference device reset function</li><li>1: Enable the training device online reset function</li><li>2: Enable the training/inference device offline reset function</li></ul><div class="note"><span class="notetitle">[!NOTE] Note</span><div class="notebody"><p>The function corresponding to the value 1 has been deprecated. Configure other values.</p></div></div>Supported training devices:<ul><li>Atlas 800 training server (model 9000) (fully configured with NPUs)</li><li>Atlas 800 training server (model 9010) (fully configured with NPUs)</li><li>Atlas 900T PoD Lite</li><li>Atlas 900 PoD (model 9000)</li><li>Atlas 800T A2 training server</li><li>Atlas 900 A2 PoD cluster basic unit</li><li>Atlas 900 A3 SuperPoD</li><li>Atlas 800T A3 SuperPoD server</li></ul>Supported inference devices:<ul><li>Atlas 300I Pro inference card</li><li>Atlas 300V video analysis card</li><li>Atlas 300V Pro video analysis card</li><li>Atlas 300I Duo inference card</li><li>Atlas 300I inference card (model 3000)</li><li>Atlas 300I inference card (model 3010)</li><li>Atlas 800I A2 inference server</li><li>A200I A2 Box heterogeneous subrack</li><li>Atlas 800I A3 SuperPoD server</li></ul>|

>[!NOTE]
>Atlas 800I A2 inference server supports the following two fault recovery methods. An Atlas 800I A2 inference server can use only one fault recovery method, which is automatically identified by cluster scheduling components.
>
>- Method 1: If no HCCS ring exists on the device, during inference job execution, when an NPU fault occurs, Ascend Device Plugin waits for the NPU to become idle and then resets the NPU.
>- Method 2: If an HCCS ring exists on the device, during inference job execution, when one or more faulty NPUs occur on the server, Ascend Device Plugin waits for all NPUs on the ring to become idle and then resets all NPUs on the ring at once.
