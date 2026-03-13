# tft\_init\_controller<a name="ZH-CN_TOPIC_0000002543201103"></a>

## 接口功能<a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_section190115624413"></a>

初始化MindIO TFT  Controller模块。

## 接口格式<a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_section4717142204511"></a>

```
mindio_ttp.framework_ttp.tft_init_controller(rank: int, world_size: int, enable_local_copy: bool, enable_arf=False, enable_zit=False)
```

## 接口参数<a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_section1232552884519"></a>

<a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000002012461173_p54600331599"><a name="zh-cn_topic_0000002012461173_p54600331599"></a><a name="zh-cn_topic_0000002012461173_p54600331599"></a>rank</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p1265420391361"><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p1265420391361"></a><a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_p1265420391361"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000002012461173_p175213010591"><a name="zh-cn_topic_0000002012461173_p175213010591"></a><a name="zh-cn_topic_0000002012461173_p175213010591"></a>当前执行训练任务的NPU卡号。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000002012461173_p5751183085911"><a name="zh-cn_topic_0000002012461173_p5751183085911"></a><a name="zh-cn_topic_0000002012461173_p5751183085911"></a>int，[-1, world_size)。<span id="ph195091850104019"><a name="ph195091850104019"></a><a name="ph195091850104019"></a>MindCluster</span>在Torch Agent进程拉起Controller时rank值取-1。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002012461173_row129665813187"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000002012461173_p14460333155914"><a name="zh-cn_topic_0000002012461173_p14460333155914"></a><a name="zh-cn_topic_0000002012461173_p14460333155914"></a>world_size</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000002012461173_p99661989186"><a name="zh-cn_topic_0000002012461173_p99661989186"></a><a name="zh-cn_topic_0000002012461173_p99661989186"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000002012461173_p0751193055915"><a name="zh-cn_topic_0000002012461173_p0751193055915"></a><a name="zh-cn_topic_0000002012461173_p0751193055915"></a>整个集群参与训练任务的卡数。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000002012461173_p177511630115913"><a name="zh-cn_topic_0000002012461173_p177511630115913"></a><a name="zh-cn_topic_0000002012461173_p177511630115913"></a>int，[1, 100000]。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002012461173_row6568195852615"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000002012461173_p1656875811264"><a name="zh-cn_topic_0000002012461173_p1656875811264"></a><a name="zh-cn_topic_0000002012461173_p1656875811264"></a>enable_local_copy</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000002012461173_p2569135802613"><a name="zh-cn_topic_0000002012461173_p2569135802613"></a><a name="zh-cn_topic_0000002012461173_p2569135802613"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000002012461173_p95692058112618"><a name="zh-cn_topic_0000002012461173_p95692058112618"></a><a name="zh-cn_topic_0000002012461173_p95692058112618"></a>表示是否启用local copy。优化器更新前，先对优化器做一次备份。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><a name="zh-cn_topic_0000002012461173_ul1914021520236"></a><a name="zh-cn_topic_0000002012461173_ul1914021520236"></a><ul id="zh-cn_topic_0000002012461173_ul1914021520236"><li>False：关闭</li><li>True：启用</li></ul>
</td>
</tr>
<tr id="row295112201153"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p1895111203153"><a name="p1895111203153"></a><a name="p1895111203153"></a>enable_arf</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p15951320151510"><a name="p15951320151510"></a><a name="p15951320151510"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p39511420171511"><a name="p39511420171511"></a><a name="p39511420171511"></a><span id="ph1088320775915"><a name="ph1088320775915"></a><a name="ph1088320775915"></a>MindIO ARF</span>特性开关。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><a name="ul89771122142714"></a><a name="ul89771122142714"></a><ul id="ul89771122142714"><li>False：关闭</li><li>True：启用</li></ul>
<p id="p1695112205159"><a name="p1695112205159"></a><a name="p1695112205159"></a>默认为False。</p>
</td>
</tr>
<tr id="row396915350278"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p628604312720"><a name="p628604312720"></a><a name="p628604312720"></a>enable_zit</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p1028610439271"><a name="p1028610439271"></a><a name="p1028610439271"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p1528634332719"><a name="p1528634332719"></a><a name="p1528634332719"></a>MindIO ZIT特性开关。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><a name="ul42872432272"></a><a name="ul42872432272"></a><ul id="ul42872432272"><li>False：关闭</li><li>True：启用</li></ul>
<p id="p1028711431272"><a name="p1028711431272"></a><a name="p1028711431272"></a>默认为False。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000002012461173_zh-cn_topic_0000001671257765_section3787164144816"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

