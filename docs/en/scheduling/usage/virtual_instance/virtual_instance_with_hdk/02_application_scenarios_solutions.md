# Application Scenarios and Solutions<a name="ZH-CN_TOPIC_0000002511426823"></a>

## Application Scenarios<a name="section198715461917"></a>

The HDK-based virtual instance function is suitable for scenarios with multiple users and parallel tasks, where the computing power requirement of each task is relatively small. For large model tasks with high computing power requirements, Ascend virtual instances are not supported.

## Virtualization Scenarios<a name="section1618382307"></a>

[Table 1](#table197838103018) describes supported virtualization scenarios when the Ascend virtual instance function is used on the physical machine or virtual machine. This section mainly describes the supported scenarios and methods for dividing vNPUs on Ascend devices.

There are two ways to partition vNPUs.

- Static virtualization: Multiple vNPUs are manually created using the `npu-smi` tool. Both physical and virtual machine scenarios support static virtualization.
- Dynamic virtualization: With software configuration, vNPUs are dynamically and automatically created upon receiving virtualization task requests, with the task then attached and vNPUs reclaimed after completion.

**Table 1** Usage scenarios

<a name="table197838103018"></a>
<table><thead align="left"><tr id="row16723873015"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p871338103019"><a name="p871338103019"></a><a name="p871338103019"></a>Supported Scenarios for Ascend Virtual Instance</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.2"><p id="p14014521402"><a name="p14014521402"></a><a name="p14014521402"></a>Operation Procedure</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.4"><p id="p18893873015"><a name="p18893873015"></a><a name="p18893873015"></a>Supported Virtualization Mode</p>
</th>
</tr>
</thead>
<tbody><tr id="row158123818304"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1819384303"><a name="p1819384303"></a><a name="p1819384303"></a>partition vNPUs on the physical machine and mount vNPUs to the virtual machine.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><p id="p1290518155817"><a name="p1290518155817"></a><a name="p1290518155817"></a>Partition vNPUs on the physical machine and mount vNPUs to the virtual machine.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p10921030123711"><a name="p10921030123711"></a><a name="p10921030123711"></a>Static Virtualization</p>
<p id="p333261621717"><a name="p333261621717"></a><a name="p333261621717"></a></p>
</td>
</tr>
<tr id="row89138123014"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p391138203014"><a name="p391138203014"></a><a name="p391138203014"></a>partition vNPUs on the physical machine and mount vNPUs to the container.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><a name="ol4232523123116"></a><a name="ol4232523123116"></a><ol id="ol4232523123116"><li>For the steps to partition vNPUs on the physical machine, see <a href="./04_creating_vnpu.md">Creating a NPUs</a>.</li><li>For the steps to mount vNPUs to the container, see <a href="./06_mounting_vnpu.md">Mounting vNPUs</a>.</li></ol>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><ul><li>Static Virtualization</li><li>Dynamic Virtualization: <ul><li>Mount using Ascend Docker Runtime</li><li>Mount using Kubernetes</li></ul></li></ul>
</td>
</tr>
<tr id="row131012387307"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1010133833013"><a name="p1010133833013"></a><a name="p1010133833013"></a>partition vNPUs on the physical machine, mount vNPUs to the virtual machine, and then mount vNPUs to containers within the virtual machine.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><a name="ol14307634103119"></a><a name="ol14307634103119"></a><ol id="ol14307634103119"><li>Partition vNPUs on the physical machine and mount vNPUs to the virtual machine.</li><li>For the steps to mount vNPUs to containers within the virtual machine, see <a href="./06_mounting_vnpu.md">Mounting vNPUs</a>.</li></ol>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><p id="p13911193234713"><a name="p13911193234713"></a><a name="p13911193234713"></a>Static Virtualization</p>
</td>
</tr>
<tr id="row3124381309"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p20127385307"><a name="p20127385307"></a><a name="p20127385307"></a>Pass through the NPU from the physical machine to the virtual machine, partition vNPUs within the virtual machine, and then mount the vNPUs to containers inside the virtual machine.</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.2 "><a name="ol441318447318"></a><a name="ol441318447318"></a><ol id="ol441318447318"><li>Passthrough NPU to a VM on the physical machine.</li><li>For the steps to partition vNPU within the VM, see <a href="./04_creating_vnpu.md">Creating vNPUs</a>.</li><li>For the steps to mount vNPUs to containers within the virtual machine, see <a href="./06_mounting_vnpu.md">Mounting vNPUs</a>.</li></ol>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.4 "><ul><li>Static Virtualization</li><li>Dynamic Virtualization: <ul><li>Mount using Ascend Docker Runtime</li><li>Mount using Kubernetes</li></ul></li></ul>
</td>
</tr>
</tbody>
</table>

## vNPU Mounting to Container Solutions<a name="section84114107544"></a>

The following solutions are available for mounting vNPUs to containers:

- Native Docker: Only static virtualization is supported (creating multiple vNPUs using the `npu-smi` tool). vNPUs are mounted to containers when containers are started using Docker.

    >[!NOTE]
    >Mounting vNPUs to containers when starting containers using native Containerd is not supported.

- Combined with MindCluster components:
    - Ascend Docker Runtime (container engine plugin): Used independently. Both static virtualization and dynamic virtualization are supported. vNPUs are mounted to containers when containers are started using Ascend Docker Runtime.
    - Kubernetes: Combined with Ascend Device Plugin and Volcano. Both static virtualization and dynamic virtualization are supported. vNPUs are mounted to containers when containers are started through Kubernetes.
        - Static virtualization: Multiple vNPUs are created in advance using the `npu-smi` tool. When users need to use vNPU resources, Ascend Device Plugin provides device discovery, device allocation, and device health status reporting functions to allocate vNPU resources for upper-layer users. In this solution, Volcano of the cluster scheduling components is optional.
        - Dynamic virtualization: Ascend Device Plugin  reports the number of available AICores on its node. After a virtualization task is submitted, Volcano calculates and schedules the task to a node that meets its requirements. The Ascend Device Plugin on that node automatically creates the vNPU device upon receiving the request and mounts it for the task, completing the dynamic virtualization process. This process does not require users to pre‑partition vNPUs, and the vNPUs are automatically reclaimed after task completion, effectively supporting scenarios with fluctuating computing power requirements.
