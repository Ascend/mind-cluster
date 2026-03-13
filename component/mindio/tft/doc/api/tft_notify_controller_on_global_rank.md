# tft\_notify\_controller\_on\_global\_rank<a name="ZH-CN_TOPIC_0000002511721136"></a>

## 接口功能<a name="zh-cn_topic_0000001975861586_section1346094194320"></a>

提供给MindCluster调用，通知MindIO TFT全局的故障卡信息。

## 接口格式<a name="zh-cn_topic_0000001975861586_section1272104915439"></a>

```
mindio_ttp.controller_ttp.tft_notify_controller_on_global_rank(fault_ranks: dict,time:int=1)
```

## 接口参数<a name="zh-cn_topic_0000001975861586_section1148355114320"></a>

<a name="zh-cn_topic_0000001975861586_table173251148163716"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001975861586_row14325848123711"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001975861586_p6325184810378"><a name="zh-cn_topic_0000001975861586_p6325184810378"></a><a name="zh-cn_topic_0000001975861586_p6325184810378"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001975861586_p1132534853714"><a name="zh-cn_topic_0000001975861586_p1132534853714"></a><a name="zh-cn_topic_0000001975861586_p1132534853714"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001975861586_p17325144811374"><a name="zh-cn_topic_0000001975861586_p17325144811374"></a><a name="zh-cn_topic_0000001975861586_p17325144811374"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001975861586_p1325048183713"><a name="zh-cn_topic_0000001975861586_p1325048183713"></a><a name="zh-cn_topic_0000001975861586_p1325048183713"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row364941564412"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p63381044122417"><a name="p63381044122417"></a><a name="p63381044122417"></a>fault_ranks</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p1333824422413"><a name="p1333824422413"></a><a name="p1333824422413"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p333834413245"><a name="p333834413245"></a><a name="p333834413245"></a>发生故障的卡信息。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p1949974519218"><a name="p1949974519218"></a><a name="p1949974519218"></a>&lt;int key, int errorType&gt;字典：</p>
<a name="ul736975113211"></a><a name="ul736975113211"></a><ul id="ul736975113211"><li>key为故障卡的rank号</li><li>errorType为故障类型：<a name="ul16212145011220"></a><a name="ul16212145011220"></a><ul id="ul16212145011220"><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul>
</li></ul>
</td>
</tr>
<tr id="row15379836104319"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p9513281516"><a name="p9513281516"></a><a name="p9513281516"></a>time</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p13522813517"><a name="p13522813517"></a><a name="p13522813517"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p35152813515"><a name="p35152813515"></a><a name="p35152813515"></a>根据环境变量设置，决定与MindCluster的修复策略交互的最大时间。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p9582819514"><a name="p9582819514"></a><a name="p9582819514"></a>int，取值范围：[1, 3600]，默认值：1，单位：s。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

-   0：调用成功
-   1：调用失败

