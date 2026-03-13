# tft\_notify\_controller\_change\_strategy<a name="ZH-CN_TOPIC_0000002511721166"></a>

## 接口功能<a name="zh-cn_topic_0000001975861586_section1346094194320"></a>

提供给MindCluster调用，通知MindIO TFT要执行的修复策略。

>**说明：** 
>该修复策略必须在MindCluster和MindIO TFT协商的可选修复策略范围内。

## 接口格式<a name="zh-cn_topic_0000001975861586_section1272104915439"></a>

```
mindio_ttp.controller_ttp.tft_notify_controller_change_strategy(strategy: str, params: str = "")
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
<tbody><tr id="row364941564412"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p1129417476504"><a name="p1129417476504"></a><a name="p1129417476504"></a>strategy</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p15292947105015"><a name="p15292947105015"></a><a name="p15292947105015"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p1529013472508"><a name="p1529013472508"></a><a name="p1529013472508"></a>通知<span id="ph64128315443"><a name="ph64128315443"></a><a name="ph64128315443"></a>MindIO TFT</span>修复策略。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p243564419268"><a name="p243564419268"></a><a name="p243564419268"></a>str，支持的修复策略如下：</p>
<a name="ul12287105882616"></a><a name="ul12287105882616"></a><ul id="ul12287105882616"><li>retry</li><li>downgrade</li><li>upgrade</li><li>recover</li><li>dump</li><li>continue</li><li>migration</li><li>exit</li></ul>
</td>
</tr>
<tr id="row141431860417"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p1114418619414"><a name="p1114418619414"></a><a name="p1114418619414"></a>params</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><a name="ul1813291418178"></a><a name="ul1813291418178"></a><ul id="ul1813291418178"><li>降级训练必选</li><li>其他可选</li></ul>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p1893544143"><a name="p1893544143"></a><a name="p1893544143"></a>降级训练参数。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p4144156114111"><a name="p4144156114111"></a><a name="p4144156114111"></a>str，默认值：""。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

-   0：调用成功
-   1：调用失败

