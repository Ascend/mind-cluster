# Supported Product Forms

The products supported by the ascend-fd-tk tool are listed in the following table.

<table>
<thead>
<tr>
<th>Product Series</th>
<th>Product Name</th>
</tr>
</thead>
<tbody>
<tr>
<td>Atlas inference products</td>
<td>Atlas 300I Duo inference card</td>
</tr>
<tr>
<td rowspan="2">Atlas A2 products</td>
<td>Atlas A2 training products: <ul><li>Atlas 800T A2</li><li>Atlas 900 A2 PoD</li><li>Atlas 200T A2 Box16</li></ul></td>
</tr>
<tr>
<td>Atlas A2 inference products: <ul><li>Atlas 800I A2 (32 GB, HCCS)</li><li>Atlas 800I A2 (32 GB, PCIe)</li><li>Atlas 800I A2 (64 GB, HCCS)</li></ul></td>
</tr>
<tr>
<td rowspan="2">Atlas A3 products</td>
<td>Atlas A3 training products: <ul><li>Atlas 800T A3</li><li>Atlas 900 A3 SuperPoD</li><li>Atlas 9000 A3 SuperPoD</li><li>A200T A3 Box8</li></ul></td>
</tr>
<tr><td>Atlas A3 inference products:<p>Atlas 800I A3</p></td></tr>
<tr>
<td rowspan="2">Ascend 950PR&950DT products</td>
<td>Ascend 950PR products:<ul><li>Atlas 350 accelerator cards</li><li>Atlas 850</li><li>Atlas 650</li></ul></td>
</tr>
<tr><td>Ascend 950DT products:<ul><li>Atlas 850E</li><li>Atlas 650E</li><li>Atlas 950 SuperPoD</li></ul></td></tr>
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
