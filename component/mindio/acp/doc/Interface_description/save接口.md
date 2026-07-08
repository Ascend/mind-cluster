# save接口<a name="ZH-CN_TOPIC_0000002468993986"></a>

## 接口功能<a name="zh-cn_topic_0000001671257765_section190115624413"></a>

将数据保存到指定的路径下。

## 接口格式<a name="zh-cn_topic_0000001671257765_section4717142204511"></a>

```
mindio_acp.save(obj, path, open_way='memfs')
```

## 接口参数<a name="zh-cn_topic_0000001671257765_section1232552884519"></a>

<a name="zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="13.13%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="12.45%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="49.36%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="25.06%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="13.13%" headers="mcps1.1.5.1.1 "><p id="p113710115016"><a name="p113710115016"></a><a name="p113710115016"></a>obj</p>
</td>
<td class="cellrowborder" valign="top" width="12.45%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001671257765_p1265420391361"><a name="zh-cn_topic_0000001671257765_p1265420391361"></a><a name="zh-cn_topic_0000001671257765_p1265420391361"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="49.36%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001671257765_p13247141681315"><a name="zh-cn_topic_0000001671257765_p13247141681315"></a><a name="zh-cn_topic_0000001671257765_p13247141681315"></a>需要保存的对象。</p>
</td>
<td class="cellrowborder" valign="top" width="25.06%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001671257765_p20482175915402"><a name="zh-cn_topic_0000001671257765_p20482175915402"></a><a name="zh-cn_topic_0000001671257765_p20482175915402"></a>有效数据对象。</p>
</td>
</tr>
<tr id="row129665813187"><td class="cellrowborder" valign="top" width="13.13%" headers="mcps1.1.5.1.1 "><p id="p1596610811817"><a name="p1596610811817"></a><a name="p1596610811817"></a>path</p>
</td>
<td class="cellrowborder" valign="top" width="12.45%" headers="mcps1.1.5.1.2 "><p id="p99661989186"><a name="p99661989186"></a><a name="p99661989186"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="49.36%" headers="mcps1.1.5.1.3 "><p id="p5966689182"><a name="p5966689182"></a><a name="p5966689182"></a>数据保存路径。</p>
</td>
<td class="cellrowborder" valign="top" width="25.06%" headers="mcps1.1.5.1.4 "><p id="p139661484188"><a name="p139661484188"></a><a name="p139661484188"></a>有效文件路径。</p>
</td>
</tr>
<tr id="row84334589522"><td class="cellrowborder" valign="top" width="13.13%" headers="mcps1.1.5.1.1 "><p id="p17434145813523"><a name="p17434145813523"></a><a name="p17434145813523"></a>open_way</p>
</td>
<td class="cellrowborder" valign="top" width="12.45%" headers="mcps1.1.5.1.2 "><p id="p3434958125212"><a name="p3434958125212"></a><a name="p3434958125212"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="49.36%" headers="mcps1.1.5.1.3 "><p id="p1849891111466"><a name="p1849891111466"></a><a name="p1849891111466"></a>保存方式。</p>
<a name="ul134756212465"></a><a name="ul134756212465"></a><ul id="ul134756212465"><li>memfs：使用<span id="ph185121283267"><a name="ph185121283267"></a><a name="ph185121283267"></a>MindIO ACP</span>的高性能MemFS保存数据。</li><li>fopen：调用C标准库中的文件操作函数保存数据，通常作为memfs方式的备份存在。</li></ul>
<p id="p17108134119401"><a name="p17108134119401"></a><a name="p17108134119401"></a>默认值：memfs。</p>
</td>
<td class="cellrowborder" valign="top" width="25.06%" headers="mcps1.1.5.1.4 "><a name="ul528292511012"></a><a name="ul528292511012"></a><ul id="ul528292511012"><li>memfs</li><li>fopen</li></ul>
</td>
</tr>
</tbody>
</table>

## 使用样例<a name="zh-cn_topic_0000001671257765_section5115161344717"></a>

```
>>> # Save to file
>>> x = torch.tensor([0, 1, 2, 3, 4])
>>> mindio_acp.save(x, '/mnt/dpc01/tensor.pt')
```

## 返回值<a name="zh-cn_topic_0000001671257765_section3787164144816"></a>

-   -1：保存失败。
-   0：通过原生torch.save方式实现保存。
-   1：通过memfs方式实现保存。
-   2：通过fopen方式实现保存。
