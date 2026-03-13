# tft\_can\_do\_uce\_repair<a name="ZH-CN_TOPIC_0000002511721156"></a>

## 接口功能<a name="zh-cn_topic_0000001975861586_section1346094194320"></a>

提供给MindSpore调用，根据L2 Cache触发的UCE故障时间和优化器更新前后时间，判断优化器数据在时间维度是否有被污染可能，进而返回是否能修复的判断结果。

>**说明：** 
>该接口仅从时间区间交集上判断优化器数据是否有被污染可能，无法根据内存地址判断。

## 接口格式<a name="zh-cn_topic_0000001975861586_section1272104915439"></a>

```
mindio_ttp.framework_ttp.tft_can_do_uce_repair(hbm_error_time: int, start_time: int = None, end_time: int = None)
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
<tbody><tr id="row364941564412"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p14656133220213"><a name="p14656133220213"></a><a name="p14656133220213"></a>hbm_error_time</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p15292947105015"><a name="p15292947105015"></a><a name="p15292947105015"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p1529013472508"><a name="p1529013472508"></a><a name="p1529013472508"></a>L2 Cache触发的UCE故障时间。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p54005509211"><a name="p54005509211"></a><a name="p54005509211"></a>int</p>
</td>
</tr>
<tr id="row1531111310216"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p106151153935"><a name="p106151153935"></a><a name="p106151153935"></a>start_time</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p553171310213"><a name="p553171310213"></a><a name="p553171310213"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p75311132027"><a name="p75311132027"></a><a name="p75311132027"></a>优化器在本地更新前从device获取的时间。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p2077511131132"><a name="p2077511131132"></a><a name="p2077511131132"></a>int</p>
</td>
</tr>
<tr id="row18126219226"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p189941857234"><a name="p189941857234"></a><a name="p189941857234"></a>end_time</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p412781916219"><a name="p412781916219"></a><a name="p412781916219"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p012711192210"><a name="p012711192210"></a><a name="p012711192210"></a>优化器在本地更新后从device获取的时间。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p852611150316"><a name="p852611150316"></a><a name="p852611150316"></a>int</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

bool值，根据时间交集判断是否可以进行UCE快恢的判断结果。

