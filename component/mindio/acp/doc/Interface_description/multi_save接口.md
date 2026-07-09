# multi\_save接口<a name="ZH-CN_TOPIC_0000002468993994"></a>

## 接口功能<a name="section94051932861"></a>

将同一个数据保存到多个文件中。

## 接口格式<a name="section1884425917718"></a>

```python
mindio_acp.multi_save(obj, path_list)
```

## 接口参数<a name="zh-cn_topic_0000001671257765_section1232552884519"></a>

<a name="zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="25%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.1 "><p id="p113710115016"><a name="p113710115016"></a><a name="p113710115016"></a>obj</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001671257765_p1265420391361"><a name="zh-cn_topic_0000001671257765_p1265420391361"></a><a name="zh-cn_topic_0000001671257765_p1265420391361"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001671257765_p13247141681315"><a name="zh-cn_topic_0000001671257765_p13247141681315"></a><a name="zh-cn_topic_0000001671257765_p13247141681315"></a>需要保存的对象。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001671257765_p20482175915402"><a name="zh-cn_topic_0000001671257765_p20482175915402"></a><a name="zh-cn_topic_0000001671257765_p20482175915402"></a>有效数据对象。</p>
</td>
</tr>
<tr id="row84334589522"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.1 "><p id="p17434145813523"><a name="p17434145813523"></a><a name="p17434145813523"></a>path_list</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.2 "><p id="p3434958125212"><a name="p3434958125212"></a><a name="p3434958125212"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.3 "><p id="p1643405818528"><a name="p1643405818528"></a><a name="p1643405818528"></a>数据保存路径列表。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.4 "><p id="p743495885211"><a name="p743495885211"></a><a name="p743495885211"></a>有效文件路径列表。</p>
</td>
</tr>
</tbody>
</table>

## 使用样例<a name="section1550852911320"></a>

```python
>>> # Save to file
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> path_list = ["/mnt/dpc01/dir1/rank_1.pt","/mnt/dpc01/dir2/rank_1.pt"]
>>> mindio_acp.multi_save(x, path_list)
```

## 返回值<a name="section8785165291317"></a>

- None：失败。
- 0：通过原生torch.save方式实现保存。
- 1：通过memfs方式实现保存。
- 2：通过fopen方式实现保存。
