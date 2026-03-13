# tft\_set\_optimizer\_replica<a name="ZH-CN_TOPIC_0000002543201127"></a>

## 接口功能<a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_section190115624413"></a>

设置rank对应的优化器状态数据副本关系。

## 接口格式<a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_section4717142204511"></a>

```
mindio_ttp.framework_ttp.tft_set_optimizer_replica(rank: int, replica_info: list)
```

## 接口参数<a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_section1232552884519"></a>

<a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="29.959999999999997%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30.04%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000002012461173_p54600331599"><a name="zh-cn_topic_0000002012461173_p54600331599"></a><a name="zh-cn_topic_0000002012461173_p54600331599"></a>rank</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p1265420391361"><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p1265420391361"></a><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p1265420391361"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="29.959999999999997%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000002012461173_p175213010591"><a name="zh-cn_topic_0000002012461173_p175213010591"></a><a name="zh-cn_topic_0000002012461173_p175213010591"></a>当前执行训练任务的NPU卡号。</p>
</td>
<td class="cellrowborder" valign="top" width="30.04%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000002012461173_p5751183085911"><a name="zh-cn_topic_0000002012461173_p5751183085911"></a><a name="zh-cn_topic_0000002012461173_p5751183085911"></a>int，[0, 100000)。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002012461173_row129665813187"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000002012461173_p14460333155914"><a name="zh-cn_topic_0000002012461173_p14460333155914"></a><a name="zh-cn_topic_0000002012461173_p14460333155914"></a>replica_info</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000002012461173_p99661989186"><a name="zh-cn_topic_0000002012461173_p99661989186"></a><a name="zh-cn_topic_0000002012461173_p99661989186"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="29.959999999999997%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000002012461173_p0751193055915"><a name="zh-cn_topic_0000002012461173_p0751193055915"></a><a name="zh-cn_topic_0000002012461173_p0751193055915"></a>副本关系list，其中每个元素是一个字典，字典按照ATTENTION（0）、MOE（1）的索引顺序排列。</p>
</td>
<td class="cellrowborder" valign="top" width="30.04%" headers="mcps1.1.5.1.4 "><pre class="screen" id="screen9897135514207"><a name="screen9897135514207"></a><a name="screen9897135514207"></a>[
{
"rank_list":list,对应的一组副本关系rank列表，PyTorch场景为DP组rank list,MindSpore场景为该卡对应的所有副本卡的list
"replica_cnt":int，副本数，PyTorch场景为副本数，MindSpore场景为rank_list的长度
"replica_shift":int，PyTorch场景有效
},
]</pre>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_section3787164144816"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

