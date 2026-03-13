# tft\_set\_dp\_group\_info<a name="ZH-CN_TOPIC_0000002543201101"></a>

## 接口功能<a name="zh-cn_topic_0000001975861586_section1346094194320"></a>

训练框架向Processor注册DP组信息。

## 接口格式<a name="zh-cn_topic_0000001975861586_section1272104915439"></a>

```
mindio_ttp.controller_ttp.tft_set_dp_group_info(rank: int, dp_rank_list: list)
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
<tbody><tr id="row364941564412"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p1129417476504"><a name="p1129417476504"></a><a name="p1129417476504"></a>rank</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p15292947105015"><a name="p15292947105015"></a><a name="p15292947105015"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p1529013472508"><a name="p1529013472508"></a><a name="p1529013472508"></a>当前rank。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="entry13445154165814p0"><a name="entry13445154165814p0"></a><a name="entry13445154165814p0"></a>大于或等于0。</p>
</td>
</tr>
<tr id="row141431860417"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p1114418619414"><a name="p1114418619414"></a><a name="p1114418619414"></a>dp_rank_list</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p1514413610415"><a name="p1514413610415"></a><a name="p1514413610415"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p0144186174120"><a name="p0144186174120"></a><a name="p0144186174120"></a>DP组信息。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p4144156114111"><a name="p4144156114111"></a><a name="p4144156114111"></a>非空。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_section16811972329"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

