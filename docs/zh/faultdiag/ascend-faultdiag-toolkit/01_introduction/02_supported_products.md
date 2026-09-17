# 支持的产品形态

ascend-fd-tk工具支持的产品如下表所示：

<table>
<thead>
<tr>
<th>产品系列</th>
<th>产品型号</th>
</tr>
</thead>
<tbody>
<tr>
<td>Atlas 推理系列产品</td>
<td>Atlas 300I Duo 推理卡</td>
</tr>
<tr>
<td rowspan="2">Atlas A2系列产品</td>
<td>Atlas A2 训练系列产品：<ul><li>Atlas 800T A2 训练服务器</li><li>Atlas 900 A2 PoD 集群基础单元</li><li>Atlas 200T A2 Box16 异构子框</li></ul></td>
</tr>
<tr>
<td>Atlas A2 推理系列产品：<ul><li>Atlas 800I A2 推理服务器（32GB HCCS款）</li><li>Atlas 800I A2 推理服务器（32GB PCIe款）</li><li>Atlas 800I A2 推理服务器（64GB HCCS款）</li></ul></td>
</tr>
<tr>
<td rowspan="2">Atlas A3系列产品</td>
<td>Atlas A3 训练系列产品：<ul><li>Atlas 800T A3 超节点服务器</li><li>Atlas 900 A3 SuperPoD 超节点</li><li>Atlas 9000 A3 SuperPoD 集群算力系统</li><li>A200T A3 Box8 超节点服务器</li></ul></td>
</tr>
<tr><td>Atlas A3 推理系列产品：<p>Atlas 800I A3 超节点服务器</p></td></tr>
<tr>
<td rowspan="2">Ascend 950PR&950DT系列产品</td>
<td>Ascend 950PR系列产品：<ul><li>Atlas 350 加速卡</li><li>Atlas 850 超节点</li><li>Atlas 650 服务器</li></ul></td>
</tr>
<tr><td>Ascend 950DT系列产品：<ul><li>Atlas 850E 超节点</li><li>Atlas 650E 服务器</li><li>Atlas 950 SuperPoD 超节点</li></ul></td></tr>
</tbody>
</table>

支持的交换机按网络平面分为以下三类，均需支持 VRP `display` 系列命令：

<table>
<thead>
<tr>
<th>交换机类型</th>
<th>所属网络平面</th>
<th>说明</th>
</tr>
</thead>
<tbody>
<tr>
<td>灵衢 L1 交换机</td>
<td rowspan="2">灵衢交换平面</td>
<td>实现单机内多 NPU 高速互通</td>
</tr>
<tr>
<td>灵衢 L2 交换机</td>
<td>完成跨机柜算力节点互联</td>
</tr>
<tr>
<td>RoCE 交换机</td>
<td>RoCE 参数平面</td>
<td>承载参数同步、数据读写等业务流量</td>
</tr>
</tbody>
</table>

>[!NOTE]
>交换机需运行 VRP（Versatile Routing Platform）操作系统，支持 `display diagnostic-information`、`display interface`、`display interface transceiver verbose` 等命令。
