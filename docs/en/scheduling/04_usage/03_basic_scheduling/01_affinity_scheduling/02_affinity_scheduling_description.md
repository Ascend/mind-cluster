# Affinity Scheduling Description<a name="ZH-CN_TOPIC_0000002511426833"></a>

Affinity scheduling refers to maximizing the computing power of Ascend AI processors by reducing resource fragmentation and reducing network congestion.

- Reduce resource fragmentation

    After task deployment, concentrate the remaining Ascend AI processors as much as possible in units with smoother network connectivity (such as nodes, SuperPoDs, or nodes under the same switch); avoid scenarios where the total number of Ascend AI processors is sufficient, but tasks cannot be scheduled due to resource fragmentation.

- Reduce network congestion

There are multiple network connection modes between Ascend AI processors. The interconnection modes of Ascend AI processors vary across different products; they also differ depending on the networking between products; and the network bandwidth varies significantly among different connection modes. Selecting different invocation strategies based on the interconnection modes of Ascend AI processors can reduce network congestion.

## Affinity Scheduling Based on Ascend AI Processors<a name="section926413474357"></a>

Within a hardware product, there are three chip link modes. Their scheduling priorities are: first, schedule tasks to Ascend AI processors within the same inference card or training card; second, schedule to Ascend AI processors interconnected via HCCS; finally, schedule to Ascend AI processors interconnected via PCIe.

> [!NOTE]
> HCCS (Huawei Cache Coherence System) is the hardware form factor of HCCL (Huawei Collective Communication Library). HCCL provides high-performance collective communication functions between servers in deep learning training scenarios.

**Figure 1** Ascend AI processor interconnection modes<a name="fig19751440101616"></a>

![](../../../../figures/scheduling/ascend-ai-processor-interconnection-modes.png)

Different hardware products may contain one or more of these three interconnection modes. The specific scheduling policies are as follows:

**Table 1** **Affinity scheduling based on Ascend AI processors**

|Hardware Form Factor|Ascend AI processor Interconnection Method|Reduce Network Congestion|Reduce Resource Fragmentation|
|--|--|--|--|
|<term>Atlas training products</term>|4 Ascend AI processors interconnected via HCCS; Ascend AI processors between HCCS rings interconnected via PCIe.|Jobs requesting 4 or fewer Ascend AI processors are scheduled onto the same HCCS ring.|If two resources have the same network conditions, the resource that results in less fragmentation after scheduling is selected.|
|<p>Atlas 200T A2 Box16 heterogeneous subrack</p><p>Atlas 200I A2 Box16 heterogeneous subrack</p>|88 Ascend AI processors interconnected via HCCS; Ascend AI processors between HCCS rings interconnected via PCIe.|<ul><li>Jobs requesting 8 or fewer Ascend AI processors are scheduled onto the same HCCS ring.</li><li>Jobs requesting more than 8 Ascend AI processors are scheduled evenly across two rings</li></ul>|If two resources have the same network conditions, the resource that results in less fragmentation after scheduling is selected.|
|<p>Atlas 900 A3 SuperPoD</p><p>A200T A3 Box8 SuperPoD server</p><p>Atlas 800I A3 SuperPoD server</p><p>Atlas 800T A3 SuperPoD server</p>|2 Ascend AI processors interconnected via SIO, forming 8 HiAM modules; each HiAM module is interconnected via HCCS.|When the number of requested Ascend AI processors is even, they must be scheduled onto the same HiAM module.|-|
|Atlas 800 inference server (model 3000) (with Atlas 300I inference card inserted)|4 Ascend AI processors interconnected within each inference card; inference cards are not interconnected with each other.|When the number of requested Ascend AI processors is less than 4 and scheduling by inference card is configured, the job will definitely be scheduled onto a single inference card.|If two resources have the same network conditions, the resource that results in less fragmentation after scheduling is selected.|
|Atlas 800 inference server (model 3000) (with Atlas 300I Duo inference card inserted)|2 Ascend AI processors interconnected via HCCS within each inference card; inference cards are interconnected via PCIe.|<p>For distributed inference scheduling, the job must be scheduled onto a full Atlas 300I Duo inference card.</p><p>If the number of Ascend AI processors required by the job is odd, the portion using a single Ascend AI processor will be preferentially scheduled to an Atlas 300I Duo inference card that has exactly 1 remaining Ascend AI processor.</p>|If two resources have the same network conditions, the resource that results in less fragmentation after scheduling is selected.|
|Server (with Atlas 350 accelerator card inserted) (4P mesh 8/16 processors)|Each Atlas 350 Acceleration Card contains 1 Ascend AI processor; every 4 Atlas 350 Acceleration Cards form a FullMesh interconnection via mezzanine boards.|<p>For single-node/distributed job scheduling, the job must be scheduled onto one or more 4P groups within a full Atlas 350 acceleration card.</p><p>If the number of Ascend AI processors required by the job is odd, the portion using a single Ascend AI processor will be preferentially scheduled to an Atlas 350 acceleration card that has exactly 1 remaining Ascend AI processor.</p>|If two resources have the same network conditions, the resource that results in less fragmentation after scheduling is selected.|
|Atlas 650E server|Within each server, 8 Ascend AI processors form an 8P mesh via UBC; servers are interconnected via PCIe.|<p>For single-node job scheduling, the job must be scheduled onto a specific server.</p><p>If the number of Ascend AI processors required by the job is odd, the portion using a single Ascend AI processor will be preferentially scheduled to an Atlas 650E server that has exactly 1 remaining Ascend AI processor.</p>|If two resources have the same network conditions, the resource that results in less fragmentation after scheduling is selected.|

## Node-Based Affinity Scheduling<a name="section1144182323712"></a>

Nodes are connected through RoCE networks or UnifiedBus devices + RoCE networks. When scheduling tasks, the UnifiedBus device network is used preferentially. The RoCE network adopts a Spine-Leaf network architecture, prioritizing network traffic control within the Leaf layer. When the Spine layer must be used, traffic is evenly distributed across all Spine layers.

- Products using RoCE connections: Atlas 800T A2 training server, Atlas 800I A2 inference server, A200I A2 Box heterogeneous subrack, Atlas 200T A2 Box16 heterogeneous subrack, Atlas 200I A2 Box16 heterogeneous subrack, Atlas 800 training server (model 9000), and Atlas 800 training server (model 9010)
- Products using single-layer RoCE connections: Atlas 800I A2 inference server, A200I A2 Box heterogeneous subrack
- Products using UnifiedBus + RoCE connections: Atlas 900 A3 SuperPoD, Atlas 850E SuperPoD, Atlas 950 SuperPoD

**Figure 2** Inter-node network<a name="fig1728811518184"></a>

![](../../../../figures/scheduling/inter-node-network.png)

**Table 2** **Inter-node affinity scheduling**

|Interconnection Mode|Ascend AI Processor Interconnection Mode|Scheduling Mode|Reduce Network Congestion|Reduce Networking Costs|Reduce Resource Fragmentation|
|--|--|--|--|--|--|
|RoCE-connected dual-layer interconnection|Global dual-layer interconnection via Spine-Leaf|Switch affinity scheduling 1.0|<ul><li>Prioritize node resources under one Leaf.</li><li>When using cross-Leaf resources, ensure even distribution of upstream traffic to each Spine.</li><li>Among multiple tasks under one Leaf, at most one task can use Spine traffic; other tasks are small tasks within the Leaf.</li></ul>|-|If the network conditions of two resources are the same, select the resource that produces less resource fragmentation after scheduling.|
|RoCE-connected dual-layer interconnection|Global dual-layer interconnection via Spine-Leaf|Switch affinity scheduling 2.0|<ul><li>Prioritize node resources under one Leaf.</li><li>When using cross-Leaf resources, ensure even distribution of upstream traffic to each Spine.</li><li>Allow multiple tasks under a specific number of Leaves to use Spine traffic.</li><li>Among multiple tasks under one Leaf, at most one task can use Spine traffic; other tasks are small tasks within the Leaf.</li></ul>|-|If the network conditions of two resources are the same, select the resource that produces less resource fragmentation after scheduling.|
|RoCE-connected single-layer connection|Single-layer connection via Leaf|Single-layer switch affinity scheduling|-|A single-layer network can meet parameter plane interconnection requirements, greatly reducing networking costs.|If the network conditions of two resources are the same, select the resource that produces less resource fragmentation after scheduling.|
|UnifiedBus + RoCE|Global interconnection via Spine-Leaf, forming multiple SuperPoDs through the UnifiedBus network|Logical SuperPoD affinity scheduling|Based on the task partitioning strategy, obtain network affinity units with high network communication requirements. Ensure that each network affinity unit is distributed under one UnifiedBus network.|-|If the network conditions of two resources are the same, select the resource that produces less resource fragmentation after scheduling.|
