# tft\_start\_updating\_os<a name="ZH-CN_TOPIC_0000002511561152"></a>

## 接口功能<a name="zh-cn_topic_0000001976021342_section13306165312920"></a>

在优化器状态更新前，调用该接口以更新optimizer state为Updating。

## 接口格式<a name="zh-cn_topic_0000001976021342_section1219942751312"></a>

```
mindio_ttp.framework_ttp.tft_start_updating_os(backup_step: int)
```

## 接口参数<a name="zh-cn_topic_0000001976021342_section02383393131"></a>

<a name="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001976021342_zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001976021342_p1595892583714"><a name="zh-cn_topic_0000001976021342_p1595892583714"></a><a name="zh-cn_topic_0000001976021342_p1595892583714"></a>backup_step</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001976021342_p887182717011"><a name="zh-cn_topic_0000001976021342_p887182717011"></a><a name="zh-cn_topic_0000001976021342_p887182717011"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001976021342_p118701727407"><a name="zh-cn_topic_0000001976021342_p118701727407"></a><a name="zh-cn_topic_0000001976021342_p118701727407"></a>备份的step。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001976021342_p098214207241"><a name="zh-cn_topic_0000001976021342_p098214207241"></a><a name="zh-cn_topic_0000001976021342_p098214207241"></a>-1或自然数，范围[-1, 9223372036854775807)。</p>
<a name="zh-cn_topic_0000001976021342_ul186934273248"></a><a name="zh-cn_topic_0000001976021342_ul186934273248"></a><ul id="zh-cn_topic_0000001976021342_ul186934273248"><li>-1：表示不使用备份step。</li><li>自然数：优化器更新前，备份的优化器状态数据对应的step。</li></ul>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001976021342_section1777516402588"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

