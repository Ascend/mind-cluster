# preload接口<a name="ZH-CN_TOPIC_0000002469154016"></a>

## 接口功能<a name="zh-cn_topic_0000002112429502_section107101141937"></a>

从文件中预加载使用torch保存的数据对象，并将其保存为MindIO ACP的高性能MemFS数据。

## 接口格式<a name="zh-cn_topic_0000002112429502_section13362162011417"></a>

```
mindio_acp.preload(*path)
```

## 接口参数<a name="zh-cn_topic_0000002112429502_section171201830749"></a>

<a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="10.81%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="13.94%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="45.43%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="29.82%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="10.81%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000002112429502_p7772205131814"><a name="zh-cn_topic_0000002112429502_p7772205131814"></a><a name="zh-cn_topic_0000002112429502_p7772205131814"></a>path</p>
</td>
<td class="cellrowborder" valign="top" width="13.94%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p1265420391361"><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p1265420391361"></a><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p1265420391361"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="45.43%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p13247141681315"><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p13247141681315"></a><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p13247141681315"></a>预加载的源文件，源文件必须存在。</p>
</td>
<td class="cellrowborder" valign="top" width="29.82%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p20482175915402"><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p20482175915402"></a><a name="zh-cn_topic_0000002112429502_zh-cn_topic_0000001671257765_p20482175915402"></a>有效文件路径或路径集合。</p>
</td>
</tr>
</tbody>
</table>

## 使用样例<a name="zh-cn_topic_0000002112429502_section81115380412"></a>

```
>>> # preload from file
>>> mindio_acp.preload('/mnt/dpc01/checkpoint/rank-0.pt')
```

## 返回值<a name="zh-cn_topic_0000002112429502_section17538071458"></a>

-   0：预加载成功。
-   1：预加载失败。
