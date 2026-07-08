# convert接口<a name="ZH-CN_TOPIC_0000002502033987"></a>

## 接口功能<a name="section188071357132719"></a>

将MindIO ACP格式的Checkpoint文件转换为Torch原生保存的格式。

## 接口格式<a name="section1642012205289"></a>

```python
mindio_acp.convert(src, dst)
```

## 接口参数<a name="zh-cn_topic_0000001671257765_section1232552884519"></a>

<a name="zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="10.81%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="11.27%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="49%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="28.92%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="10.81%" headers="mcps1.1.5.1.1 "><p id="p113710115016"><a name="p113710115016"></a><a name="p113710115016"></a>src</p>
</td>
<td class="cellrowborder" valign="top" width="11.27%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001671257765_p1265420391361"><a name="zh-cn_topic_0000001671257765_p1265420391361"></a><a name="zh-cn_topic_0000001671257765_p1265420391361"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="49%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001671257765_p13247141681315"><a name="zh-cn_topic_0000001671257765_p13247141681315"></a><a name="zh-cn_topic_0000001671257765_p13247141681315"></a>待转换的源路径或源文件，源路径或源文件必须存在。</p>
</td>
<td class="cellrowborder" valign="top" width="28.92%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001671257765_p20482175915402"><a name="zh-cn_topic_0000001671257765_p20482175915402"></a><a name="zh-cn_topic_0000001671257765_p20482175915402"></a>有效文件路径，不能包含软链接。</p>
</td>
</tr>
<tr id="row84334589522"><td class="cellrowborder" valign="top" width="10.81%" headers="mcps1.1.5.1.1 "><p id="p17434145813523"><a name="p17434145813523"></a><a name="p17434145813523"></a>dst</p>
</td>
<td class="cellrowborder" valign="top" width="11.27%" headers="mcps1.1.5.1.2 "><p id="p3434958125212"><a name="p3434958125212"></a><a name="p3434958125212"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="49%" headers="mcps1.1.5.1.3 "><p id="p1698125181913"><a name="p1698125181913"></a><a name="p1698125181913"></a>待转换的目标路径或目标文件。指定路径的父目录必须存在，如果文件已存在，则会被覆盖。</p>
</td>
<td class="cellrowborder" valign="top" width="28.92%" headers="mcps1.1.5.1.4 "><p id="p4501191611019"><a name="p4501191611019"></a><a name="p4501191611019"></a>有效文件路径，不能包含软链接。</p>
</td>
</tr>
</tbody>
</table>

## 使用样例<a name="section1550852911320"></a>

```python
>>> mindio_acp.convert('/mnt/dpc01/iter_0000050/mp_rank_00/distrib_optim.pt', '/mnt/dpc02/iter_0000050/mp_rank_00/distrib_optim.pt')
```

## 返回值<a name="section8785165291317"></a>

- 0：转换成功。
- -1：转换失败。
