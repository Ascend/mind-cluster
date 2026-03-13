# Action<a name="ZH-CN_TOPIC_0000002511721170"></a>

## 接口功能<a name="zh-cn_topic_0000001975861594_section106131225152112"></a>

主线程上报异常后的动作类型枚举。

## 接口格式<a name="zh-cn_topic_0000001975861594_section5166205542118"></a>

```
mindio_ttp.framework_ttp.Action
```

## 接口参数<a name="zh-cn_topic_0000001975861594_section28331950124012"></a>

<a name="zh-cn_topic_0000001975861594_table191585974013"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001975861594_row171517595400"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001975861594_p915759154011"><a name="zh-cn_topic_0000001975861594_p915759154011"></a><a name="zh-cn_topic_0000001975861594_p915759154011"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001975861594_p21525904019"><a name="zh-cn_topic_0000001975861594_p21525904019"></a><a name="zh-cn_topic_0000001975861594_p21525904019"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001975861594_p1915175994014"><a name="zh-cn_topic_0000001975861594_p1915175994014"></a><a name="zh-cn_topic_0000001975861594_p1915175994014"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001975861594_p1815159104017"><a name="zh-cn_topic_0000001975861594_p1815159104017"></a><a name="zh-cn_topic_0000001975861594_p1815159104017"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001975861594_row151545954011"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001975861594_p1415175914017"><a name="zh-cn_topic_0000001975861594_p1415175914017"></a><a name="zh-cn_topic_0000001975861594_p1415175914017"></a>Action</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001975861594_p815125916409"><a name="zh-cn_topic_0000001975861594_p815125916409"></a><a name="zh-cn_topic_0000001975861594_p815125916409"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001975861594_p41505912408"><a name="zh-cn_topic_0000001975861594_p41505912408"></a><a name="zh-cn_topic_0000001975861594_p41505912408"></a>区分主线程上报异常后的动作类型，具体如下：</p>
<a name="ul545122118311"></a><a name="ul545122118311"></a><ul id="ul545122118311"><li>RETRY：修复成功后续训。</li><li>EXIT：退出。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><a name="ul3622192314314"></a><a name="ul3622192314314"></a><ul id="ul3622192314314"><li>RETRY：0</li><li>EXIT：1</li></ul>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

无

