# load接口<a name="ZH-CN_TOPIC_0000002468993980"></a>

## 接口功能<a name="section188071357132719"></a>

从文件中加载save/multi\_save接口持久化的对象。

## 接口格式<a name="section1642012205289"></a>

```python
mindio_acp.load(path, open_way='memfs', map_location=None)
```

## 接口参数<a name="zh-cn_topic_0000001671257765_section1232552884519"></a>

<a name="zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="12.43%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="11.64%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="50.93%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="12.43%" headers="mcps1.1.5.1.1 "><p id="p113710115016"><a name="p113710115016"></a><a name="p113710115016"></a>path</p>
</td>
<td class="cellrowborder" valign="top" width="11.64%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001671257765_p1265420391361"><a name="zh-cn_topic_0000001671257765_p1265420391361"></a><a name="zh-cn_topic_0000001671257765_p1265420391361"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="50.93%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001671257765_p13247141681315"><a name="zh-cn_topic_0000001671257765_p13247141681315"></a><a name="zh-cn_topic_0000001671257765_p13247141681315"></a>加载路径。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001671257765_p20482175915402"><a name="zh-cn_topic_0000001671257765_p20482175915402"></a><a name="zh-cn_topic_0000001671257765_p20482175915402"></a>有效文件路径。</p>
</td>
</tr>
<tr id="row84334589522"><td class="cellrowborder" valign="top" width="12.43%" headers="mcps1.1.5.1.1 "><p id="p17434145813523"><a name="p17434145813523"></a><a name="p17434145813523"></a>open_way</p>
</td>
<td class="cellrowborder" valign="top" width="11.64%" headers="mcps1.1.5.1.2 "><p id="p3434958125212"><a name="p3434958125212"></a><a name="p3434958125212"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="50.93%" headers="mcps1.1.5.1.3 "><p id="p13681313113"><a name="p13681313113"></a><a name="p13681313113"></a>加载方式。</p>
<a name="ul18492928512"></a><a name="ul18492928512"></a><ul id="ul18492928512"><li>memfs：使用<span id="ph185121283267"><a name="ph185121283267"></a><a name="ph185121283267"></a>MindIO ACP</span>的高性能MemFS保存数据。</li><li>fopen：调用C标准库中的文件操作函数保存数据，通常作为memfs方式的备份存在。</li></ul>
<p id="p189230513417"><a name="p189230513417"></a><a name="p189230513417"></a>默认值：memfs。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.4 "><a name="ul1349173411131"></a><a name="ul1349173411131"></a><ul id="ul1349173411131"><li>memfs</li><li>fopen</li></ul>
</td>
</tr>
<tr id="row145471244168"><td class="cellrowborder" valign="top" width="12.43%" headers="mcps1.1.5.1.1 "><p id="p17674949677"><a name="p17674949677"></a><a name="p17674949677"></a>map_location</p>
</td>
<td class="cellrowborder" valign="top" width="11.64%" headers="mcps1.1.5.1.2 "><p id="p1354716444615"><a name="p1354716444615"></a><a name="p1354716444615"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="50.93%" headers="mcps1.1.5.1.3 "><p id="p754794412610"><a name="p754794412610"></a><a name="p754794412610"></a>加载时需要映射到的设备。默认值：None。</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.1.5.1.4 "><a name="ul918011557718"></a><a name="ul918011557718"></a><ul id="ul918011557718"><li>None</li><li>cpu</li></ul>
</td>
</tr>
</tbody>
</table>

## 使用样例<a name="section1550852911320"></a>

```python
>>> # load from file
>>> mindio_acp.load('/mnt/dpc01/checkpoint/rank-0.pt')
```

## 返回值<a name="section8785165291317"></a>

Any

>**须知：**
>如同PyTorch的load接口，本接口内部也使用pickle模块，存在被恶意构造的数据在unpickle期间攻击的风险。需要保证被加载的数据来源是安全存储的，仅可以load可信的数据。
