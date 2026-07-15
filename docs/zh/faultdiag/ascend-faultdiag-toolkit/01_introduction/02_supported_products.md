# 支持的产品形态

ascend-fd-tk 工具支持的服务器产品如下表所示：

<table>
<thead>
<tr>
<th>产品系列</th>
<th>产品名称</th>
</tr>
</thead>
<tbody>
<tr>
<td rowspan="3">Atlas A2 训练系列产品</td>
<td>Atlas 200T A2 Box16 异构子框</td>
</tr>
<tr>
<td>Atlas 800T A2 训练服务器</td>
</tr>
<tr>
<td>Atlas 900 A2 PoD 集群基础单元</td>
</tr>
<tr>
<td>Atlas 推理系列产品</td>
<td>Atlas 300I Duo 推理卡</td>
</tr>
<tr>
<td rowspan="3">Atlas 800I A2 推理产品</td>
<td>Atlas 800I A2 推理服务器（32GB HCCS 款）</td>
</tr>
<tr>
<td>Atlas 800I A2 推理服务器（32GB PCIe 款）</td>
</tr>
<tr>
<td>Atlas 800I A2 推理服务器（64GB HCCS 款）</td>
</tr>
<tr>
<td rowspan="4">Atlas A3 训练系列产品</td>
<td>Atlas 900 A3 SuperPoD 超节点</td>
</tr>
<tr>
<td>Atlas 9000 A3 SuperPoD 集群算力系统</td>
</tr>
<tr>
<td>Atlas 800T A3 超节点服务器</td>
</tr>
<tr>
<td>A200T A3 Box8 超节点服务器</td>
</tr>
<tr>
<td>Atlas A3 推理系列产品</td>
<td>Atlas 800I A3 超节点服务器</td>
</tr>
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

> 交换机需运行 VRP（Versatile Routing Platform）操作系统，支持 `display diagnostic-information`、`display interface`、`display interface transceiver verbose` 等命令。
