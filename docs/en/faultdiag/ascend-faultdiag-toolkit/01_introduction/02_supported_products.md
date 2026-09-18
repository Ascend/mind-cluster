# Supported Product Forms

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:15:00.330Z pushedAt=2026-08-24T02:22:41.507Z -->

The servers supported by the ascend-fd-tk tool are listed in the following table.

<table>
<thead>
<tr>
<th>Product Series</th>
<th>Product Name</th>
</tr>
</thead>
<tbody>
<tr>
<td rowspan="3">Atlas A2 training products</td>
<td>Atlas 200T A2 Box16 heterogeneous subrack</td>
</tr>
<tr>
<td>Atlas 800T A2 training server</td>
</tr>
<tr>
<td>Atlas 900 A2 PoD cluster base unit</td>
</tr>
<tr>
<td rowspan="4">Atlas A3 training products</td>
<td>Atlas 900 A3 SuperPoD</td>
</tr>
<tr>
<td>Atlas 9000 A3 SuperPoD cluster computing system</td>
</tr>
<tr>
<td>Atlas 800T A3 SuperPoD server</td>
</tr>
<tr>
<td>A200T A3 Box8 SuperPoD server</td>
</tr>
<tr>
<td>Atlas inference products</td>
<td>Atlas 300I Duo inference card</td>
</tr>
<tr>
<td rowspan="3">Atlas A2 inference products</td>
<td>Atlas 800I A2 inference server (32 GB; HCCS)</td>
</tr>
<tr>
<td>Atlas 800I A2 inference server (32 GB; PCIe)</td>
</tr>
<tr>
<td>Atlas 800I A2 inference server (64 GB; HCCS)</td>
</tr>
<tr>
<td>Atlas A3 inference products</td>
<td>Atlas 800I A3 SuperPoD server</td>
</tr>
</tbody>
</table>

The supported switches are classified into the following three types by network plane, all of which must support the Versatile Routing Platform (VRP) commands, for example, `display`.

<table>
<thead>
<tr>
<th>Switch Type</th>
<th>Network Plane</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>UnifiedBus L1 switch</td>
<td rowspan="2">UnifiedBus switching plane</td>
<td>Enables high-speed interconnection of multiple NPUs within a single node</td>
</tr>
<tr>
<td>UnifiedBus L2 switch</td>
<td>Interconnects compute nodes across cabinets</td>
</tr>
<tr>
<td>RoCE switch</td>
<td>RoCE parameter plane</td>
<td>Carries service traffic such as parameter synchronization and data read/write</td>
</tr>
</tbody>
</table>

>[!NOTE]
>Switches must run the VRP OS and support commands such as `display diagnostic-information`, `display interface`, and `display interface transceiver verbose`.
