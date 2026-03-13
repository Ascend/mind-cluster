# tft\_init\_processor<a name="ZH-CN_TOPIC_0000002511561170"></a>

## 接口功能<a name="zh-cn_topic_0000001976021294_section188071357132719"></a>

初始化MindIO TFT  Processor模块。

## 接口格式<a name="zh-cn_topic_0000001976021294_section1642012205289"></a>

```
mindio_ttp.framework_ttp.tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=True, tls_info='', enable_uce=True, enable_arf=False, enable_zit=False)
```

## 接口参数<a name="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_section1232552884519"></a>

<a name="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001976021294_zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001976021294_p101082013608"><a name="zh-cn_topic_0000001976021294_p101082013608"></a><a name="zh-cn_topic_0000001976021294_p101082013608"></a>rank</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001976021294_p31086138010"><a name="zh-cn_topic_0000001976021294_p31086138010"></a><a name="zh-cn_topic_0000001976021294_p31086138010"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001976021294_p4108713207"><a name="zh-cn_topic_0000001976021294_p4108713207"></a><a name="zh-cn_topic_0000001976021294_p4108713207"></a>当前执行训练任务NPU卡号。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001976021294_p3107151318014"><a name="zh-cn_topic_0000001976021294_p3107151318014"></a><a name="zh-cn_topic_0000001976021294_p3107151318014"></a>int，[0, world_size)。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001976021294_row86871818192112"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001976021294_p36871818132120"><a name="zh-cn_topic_0000001976021294_p36871818132120"></a><a name="zh-cn_topic_0000001976021294_p36871818132120"></a>world_size</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001976021294_p19687151812218"><a name="zh-cn_topic_0000001976021294_p19687151812218"></a><a name="zh-cn_topic_0000001976021294_p19687151812218"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001976021294_p5687111882111"><a name="zh-cn_topic_0000001976021294_p5687111882111"></a><a name="zh-cn_topic_0000001976021294_p5687111882111"></a>参与训练任务的集群卡数。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001976021294_p6687518192119"><a name="zh-cn_topic_0000001976021294_p6687518192119"></a><a name="zh-cn_topic_0000001976021294_p6687518192119"></a>int，[1, 100000]。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001976021294_row1095512167347"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001976021294_p12955121613416"><a name="zh-cn_topic_0000001976021294_p12955121613416"></a><a name="zh-cn_topic_0000001976021294_p12955121613416"></a>enable_local_copy</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001976021294_p99551816123412"><a name="zh-cn_topic_0000001976021294_p99551816123412"></a><a name="zh-cn_topic_0000001976021294_p99551816123412"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001976021294_p18955141611344"><a name="zh-cn_topic_0000001976021294_p18955141611344"></a><a name="zh-cn_topic_0000001976021294_p18955141611344"></a>是否启用local copy。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><a name="zh-cn_topic_0000001976021294_ul123600322417"></a><a name="zh-cn_topic_0000001976021294_ul123600322417"></a><ul id="zh-cn_topic_0000001976021294_ul123600322417"><li>False：关闭</li><li>True：启用</li></ul>
</td>
</tr>
<tr id="row133600144267"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p024015741618"><a name="p024015741618"></a><a name="p024015741618"></a>enable_tls</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p524015751612"><a name="p524015751612"></a><a name="p524015751612"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p32405575168"><a name="p32405575168"></a><a name="p32405575168"></a>TLS加密传输开关。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><a name="ul89771122142714"></a><a name="ul89771122142714"></a><ul id="ul89771122142714"><li>False：关闭</li><li>True：启用</li></ul>
<p id="p2024010577168"><a name="p2024010577168"></a><a name="p2024010577168"></a>默认为True。</p>
</td>
</tr>
<tr id="row633610102292"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p155211091713"><a name="p155211091713"></a><a name="p155211091713"></a>tls_info</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p165220071717"><a name="p165220071717"></a><a name="p165220071717"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p1252901172"><a name="p1252901172"></a><a name="p1252901172"></a>TLS的证书配置。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p05216091710"><a name="p05216091710"></a><a name="p05216091710"></a>默认为空，当开启TLS认证时，需要配置证书信息，具体字段应以键值对形式组织。具体配置指导见<a href="导入TLS证书.md">导入TLS证书</a>。</p>
</td>
</tr>
<tr id="row735874715915"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p193582471495"><a name="p193582471495"></a><a name="p193582471495"></a>enable_uce</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p183583471394"><a name="p183583471394"></a><a name="p183583471394"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p63586477914"><a name="p63586477914"></a><a name="p63586477914"></a><span id="ph10595622285"><a name="ph10595622285"></a><a name="ph10595622285"></a>MindIO UCE</span>特性开关。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><a name="ul88286481114"></a><a name="ul88286481114"></a><ul id="ul88286481114"><li>False：关闭</li><li>True：启用</li></ul>
<p id="p639592610183"><a name="p639592610183"></a><a name="p639592610183"></a>默认为True。</p>
</td>
</tr>
<tr id="row10557145820144"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p17558135813142"><a name="p17558135813142"></a><a name="p17558135813142"></a>enable_arf</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p75581858101416"><a name="p75581858101416"></a><a name="p75581858101416"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p855895812145"><a name="p855895812145"></a><a name="p855895812145"></a><span id="ph1088320775915"><a name="ph1088320775915"></a><a name="ph1088320775915"></a>MindIO ARF</span>特性开关。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><a name="ul207883816151"></a><a name="ul207883816151"></a><ul id="ul207883816151"><li>False：关闭</li><li>True：启用</li></ul>
<p id="p20221631181818"><a name="p20221631181818"></a><a name="p20221631181818"></a>默认为False。</p>
</td>
</tr>
<tr id="row5612453192910"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p628604312720"><a name="p628604312720"></a><a name="p628604312720"></a>enable_zit</p>
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

## 返回值<a name="zh-cn_topic_0000001976021294_section8785165291317"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

