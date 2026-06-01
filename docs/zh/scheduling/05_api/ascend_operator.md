# Ascend Operator<a name="ZH-CN_TOPIC_0000002511346797"></a>

## rings-config-<任务名称\><a name="section1377115581385"></a>

**表 1**  rings-config-任务名称

<a name="table1328211233126"></a>
<table><thead align="left"><tr id="row2028442312122"><th class="cellrowborder" valign="top" width="9.99%" id="mcps1.2.6.1.1"><p id="p0566161515246"><a name="p0566161515246"></a><a name="p0566161515246"></a>字段名称</p>
</th>
<th class="cellrowborder" valign="top" width="20.630000000000003%" id="mcps1.2.6.1.2"><p id="p428442317128"><a name="p428442317128"></a><a name="p428442317128"></a>名称</p>
</th>
<th class="cellrowborder" valign="top" width="27.779999999999998%" id="mcps1.2.6.1.3"><p id="p32851623121215"><a name="p32851623121215"></a><a name="p32851623121215"></a>作用</p>
</th>
<th class="cellrowborder" valign="top" width="18.48%" id="mcps1.2.6.1.4"><p id="p122851233123"><a name="p122851233123"></a><a name="p122851233123"></a>取值</p>
</th>
<th class="cellrowborder" valign="top" width="23.119999999999997%" id="mcps1.2.6.1.5"><p id="p2028522371219"><a name="p2028522371219"></a><a name="p2028522371219"></a>备注</p>
</th>
</tr>
</thead>
<tbody><tr id="row142851523101217"><td class="cellrowborder" rowspan="9" valign="top" width="9.99%" headers="mcps1.2.6.1.1 "><p id="p383082613247"><a name="p383082613247"></a><a name="p383082613247"></a>hccl.json</p>
</td>
<td class="cellrowborder" valign="top" width="20.630000000000003%" headers="mcps1.2.6.1.2 "><p id="p20285172315128"><a name="p20285172315128"></a><a name="p20285172315128"></a>version</p>
</td>
<td class="cellrowborder" valign="top" width="27.779999999999998%" headers="mcps1.2.6.1.3 "><p id="p628512301218"><a name="p628512301218"></a><a name="p628512301218"></a>RankTable使用的格式版本</p>
</td>
<td class="cellrowborder" valign="top" width="18.48%" headers="mcps1.2.6.1.4 "><p id="p728612315129"><a name="p728612315129"></a><a name="p728612315129"></a>1.0</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.5 "><p id="p192861223151219"><a name="p192861223151219"></a><a name="p192861223151219"></a>-</p>
</td>
</tr>
<tr id="row92861423161214"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p16286723111214"><a name="p16286723111214"></a><a name="p16286723111214"></a>server_count</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p528602331212"><a name="p528602331212"></a><a name="p528602331212"></a>任务使用的节点数量</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p10286623111214"><a name="p10286623111214"></a><a name="p10286623111214"></a>整数类型</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p13286122371216"><a name="p13286122371216"></a><a name="p13286122371216"></a>-</p>
</td>
</tr>
<tr id="row1628711236125"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p528710237122"><a name="p528710237122"></a><a name="p528710237122"></a>server_list</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12287423101219"><a name="p12287423101219"></a><a name="p12287423101219"></a>任务使用的节点信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p428716231129"><a name="p428716231129"></a><a name="p428716231129"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p52876233123"><a name="p52876233123"></a><a name="p52876233123"></a>-</p>
</td>
</tr>
<tr id="row228712311218"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p172881238128"><a name="p172881238128"></a><a name="p172881238128"></a>- server_id</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p228820231122"><a name="p228820231122"></a><a name="p228820231122"></a>AI Server标识，全局唯一</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p928812351211"><a name="p928812351211"></a><a name="p928812351211"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p112882023131216"><a name="p112882023131216"></a><a name="p112882023131216"></a>-</p>
</td>
</tr>
<tr id="row526819617575"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p331911115717"><a name="p331911115717"></a><a name="p331911115717"></a>- host_ip</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p17314111571"><a name="p17314111571"></a><a name="p17314111571"></a>AI Server的Host IP地址</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p731131111573"><a name="p731131111573"></a><a name="p731131111573"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p73171113571"><a name="p73171113571"></a><a name="p73171113571"></a>-</p>
</td>
</tr>
<tr id="row1128892381211"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p6288523171213"><a name="p6288523171213"></a><a name="p6288523171213"></a>device</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p128916238124"><a name="p128916238124"></a><a name="p128916238124"></a>任务使用的芯片信息</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p14289152311122"><a name="p14289152311122"></a><a name="p14289152311122"></a>-</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p028902319121"><a name="p028902319121"></a><a name="p028902319121"></a>-</p>
</td>
</tr>
<tr id="row528919236120"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1428972321213"><a name="p1428972321213"></a><a name="p1428972321213"></a>- device_id</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p5289423131210"><a name="p5289423131210"></a><a name="p5289423131210"></a>任务使用的芯片的物理ID</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p0289172316122"><a name="p0289172316122"></a><a name="p0289172316122"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1328912313128"><a name="p1328912313128"></a><a name="p1328912313128"></a>-</p>
</td>
</tr>
<tr id="row10290172361218"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p122909232126"><a name="p122909232126"></a><a name="p122909232126"></a>- device_ip</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14290023151213"><a name="p14290023151213"></a><a name="p14290023151213"></a>任务使用的芯片的IP地址</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1929042321217"><a name="p1929042321217"></a><a name="p1929042321217"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p62901323161220"><a name="p62901323161220"></a><a name="p62901323161220"></a>-</p>
</td>
</tr>
<tr id="row1429013237126"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p6290123131214"><a name="p6290123131214"></a><a name="p6290123131214"></a>- rank_id</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p12291202391210"><a name="p12291202391210"></a><a name="p12291202391210"></a>任务使用的芯片的Rank号</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p17291223121215"><a name="p17291223121215"></a><a name="p17291223121215"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p52911623171220"><a name="p52911623171220"></a><a name="p52911623171220"></a>-</p>
</td>
</tr>
<tr id="row115483483241"><td class="cellrowborder" valign="top" width="9.99%" headers="mcps1.2.6.1.1 "><p id="p6549948102419"><a name="p6549948102419"></a><a name="p6549948102419"></a>version</p>
</td>
<td class="cellrowborder" valign="top" width="20.630000000000003%" headers="mcps1.2.6.1.2 "><p id="p12549174817242"><a name="p12549174817242"></a><a name="p12549174817242"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="27.779999999999998%" headers="mcps1.2.6.1.3 "><p id="p165492487247"><a name="p165492487247"></a><a name="p165492487247"></a>任务使用hccl.json的版本</p>
</td>
<td class="cellrowborder" valign="top" width="18.48%" headers="mcps1.2.6.1.4 "><p id="p145492048112416"><a name="p145492048112416"></a><a name="p145492048112416"></a>字符串</p>
</td>
<td class="cellrowborder" valign="top" width="23.119999999999997%" headers="mcps1.2.6.1.5 "><p id="p65521848132412"><a name="p65521848132412"></a><a name="p65521848132412"></a>-</p>
</td>
</tr>
</tbody>
</table>
