# tft\_end\_updating\_os<a name="ZH-CN_TOPIC_0000002543201129"></a>

## 接口功能<a name="zh-cn_topic_0000002012581733_section1484723013148"></a>

在优化器状态更新完成后，调用该接口以更新optimizer state为Updated。

## 接口格式<a name="zh-cn_topic_0000002012581733_section17562123821418"></a>

```
mindio_ttp.framework_ttp.tft_end_updating_os(step: int)
```

## 接口参数<a name="zh-cn_topic_0000002012581733_section13536743131418"></a>

<a name="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002012581733_zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000002012581733_p1595892583714"><a name="zh-cn_topic_0000002012581733_p1595892583714"></a><a name="zh-cn_topic_0000002012581733_p1595892583714"></a>step</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000002012581733_p887182717011"><a name="zh-cn_topic_0000002012581733_p887182717011"></a><a name="zh-cn_topic_0000002012581733_p887182717011"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000002012581733_p118701727407"><a name="zh-cn_topic_0000002012581733_p118701727407"></a><a name="zh-cn_topic_0000002012581733_p118701727407"></a>当前的step。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000002012581733_p187017276016"><a name="zh-cn_topic_0000002012581733_p187017276016"></a><a name="zh-cn_topic_0000002012581733_p187017276016"></a>正整数，范围[1, 9223372036854775807)。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000002012581733_section1777516402588"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

