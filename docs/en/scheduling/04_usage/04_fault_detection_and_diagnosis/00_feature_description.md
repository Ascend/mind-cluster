# Fault Detection Feature Description

## Fault Description

The fault detection capability obtains the fault status of the cluster and of training jobs, so that downstream features such as resumable training can handle faults based on the detection result. Currently, the fault detection feature mainly provides the following fault detection capabilities: Ascend hardware faults, training job faults, and faults reported by other fault providers.

## Supported Fault Modes

Currently, more than 200 faults can be detected. For supported fault types, see [Table 1](#zh-cn_topic_0000002039699773_table9980135316395). For detailed fault descriptions, see [Typical_Faults.xlsx](../../../resource/Typical_Faults.xlsx).

**Table 1** Fault types

<a name="zh-cn_topic_0000002039699773_table9980135316395"></a>

|Fault Type|Fault Description|
|--|--|
|Node faults|<p>Include node health status, node hardware faults, and shared storage faults.</p><p>For fault code descriptions, see [Node Fault Code References](../../07_references/05_appendix.md#node-fault-code-references).</p><p>If a node hardware fault causes the node to go down or restart, NodeD cannot detect the specific fault type or report it.</p>|
|Chip faults|<p>Include chip faults reported through the DCMI and chip network faults detected by the device network detection tool `hccn_tool`.</p><p>For fault code descriptions, see [Chip Fault Code References](../../07_references/05_appendix.md#chip-fault-code-references).</p>|
|Parameter plane network faults|Include chip network-related faults and UnifiedBus (UB) device faults.<ul><li>Chip network-related faults: The dedicated network used for parameter exchange between chips is faulty, for example, an NPU network port fault.</li><li>UnifiedBus device faults: The UnifiedBus devices of <term>Atlas A3 training products</term> are faulty.</li></ul>|
|Service plane faults|<p>A training job exits unexpectedly, causing the status of the Pod to become "Failed".</p><p>You can run the <strong>kubectl describe pod <em>{pod_name} </em>-n <em>\{NAMESPACE\}</em> \|grep Status:</strong> command to check whether the status of the current Pod is "Failed". The following is an example:<pre class="screen"><strong>Status:       Failed</strong></pre></p>|
|Public faults|Public faults are faults provided by other fault providers (non-MindCluster components). Public faults include the following types: NPU faults, node faults, network faults, and storage faults.|
|pingmesh UnifiedBus network faults|Detected NPU network faults within the HCCS network inside a SuperPoD (including intra-node and inter-node).|
|Performance degradation faults|MindStudio provides the profiling capability. MindCluster uses it to provide diagnosis for performance degradation faults (slow nodes) in a cluster. This feature provides dynamic start of profile points and persistence of profile data. It can be dynamically started and stopped, and diagnosis does not require restarting the job, causing no loss to training.|
|Task hang faults|A task hang fault is an abnormal state in which a process is running on an NPU but has made no substantive progress for a long time due to computing hang, memory hang, or communication hang.|
|Optical link member port faults|Refer to faults occurred on the parameter plane optical link member port on an NPU, including union faults, 5808 faults, and UBOE port faults. Because Ascend Device Plugin cannot sense remote faults, it can only make a comprehensive judgment based on the fault code and the port status of the target device, and therefore fault codes of multiple levels are included.|
|UB NIC faults|Refer to faults of UB NICs used in Atlas 950 SuperPoD, Atlas 850E, Atlas 850, Atlas 650E, and Atlas 650, including UB port faults, UB lane degradation, HCA port faults, Bond member faults, and DPU loss faults. They are detected by the K8s RDMA Shared Dev Plugin component.|
