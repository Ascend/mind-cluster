# tft\_start\_controller<a name="ZH-CN_TOPIC_0000002543201121"></a>

## 接口功能<a name="zh-cn_topic_0000001976021314_section94051932861"></a>

在初始化Controller模块成功后，调用该接口以启动MindIO TFT  Controller模块服务。

## 接口格式<a name="zh-cn_topic_0000001976021314_section1884425917718"></a>

```
mindio_ttp.framework_ttp.tft_start_controller(bind_ip: str, port: int, enable_tls=True, tls_info='')
```

## 接口参数<a name="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_section1232552884519"></a>

<a name="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="29.959999999999997%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30.04%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001976021314_zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001976021314_p65181146115911"><a name="zh-cn_topic_0000001976021314_p65181146115911"></a><a name="zh-cn_topic_0000001976021314_p65181146115911"></a>bind_ip</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001976021314_p95183465595"><a name="zh-cn_topic_0000001976021314_p95183465595"></a><a name="zh-cn_topic_0000001976021314_p95183465595"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="29.959999999999997%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001976021314_p051718465599"><a name="zh-cn_topic_0000001976021314_p051718465599"></a><a name="zh-cn_topic_0000001976021314_p051718465599"></a>Controller对外提供服务的IP地址或域名。</p>
</td>
<td class="cellrowborder" valign="top" width="30.04%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001976021314_p351713468599"><a name="zh-cn_topic_0000001976021314_p351713468599"></a><a name="zh-cn_topic_0000001976021314_p351713468599"></a>符合IP地址规范的IPv4地址，位于集群节点IP地址中，禁止全零IP地址，支持域名。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001976021314_row84334589522"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001976021314_p951717468592"><a name="zh-cn_topic_0000001976021314_p951717468592"></a><a name="zh-cn_topic_0000001976021314_p951717468592"></a>port</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001976021314_p4516134618598"><a name="zh-cn_topic_0000001976021314_p4516134618598"></a><a name="zh-cn_topic_0000001976021314_p4516134618598"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="29.959999999999997%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001976021314_p16516184605911"><a name="zh-cn_topic_0000001976021314_p16516184605911"></a><a name="zh-cn_topic_0000001976021314_p16516184605911"></a>Controller侦听端口号。</p>
</td>
<td class="cellrowborder" valign="top" width="30.04%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001976021314_p3515164620599"><a name="zh-cn_topic_0000001976021314_p3515164620599"></a><a name="zh-cn_topic_0000001976021314_p3515164620599"></a>[1024, 65535]</p>
</td>
</tr>
<tr id="row13239457171618"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p024015741618"><a name="p024015741618"></a><a name="p024015741618"></a>enable_tls</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p524015751612"><a name="p524015751612"></a><a name="p524015751612"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="29.959999999999997%" headers="mcps1.1.5.1.3 "><p id="p32405575168"><a name="p32405575168"></a><a name="p32405575168"></a>TLS加密传输开关。</p>
</td>
<td class="cellrowborder" valign="top" width="30.04%" headers="mcps1.1.5.1.4 "><a name="ul89771122142714"></a><a name="ul89771122142714"></a><ul id="ul89771122142714"><li>False：关闭</li><li>True：启用</li></ul>
<p id="p2024010577168"><a name="p2024010577168"></a><a name="p2024010577168"></a>默认为True。</p>
</td>
</tr>
<tr id="row105110018179"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p155211091713"><a name="p155211091713"></a><a name="p155211091713"></a>tls_info</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p165220071717"><a name="p165220071717"></a><a name="p165220071717"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="29.959999999999997%" headers="mcps1.1.5.1.3 "><p id="p1252901172"><a name="p1252901172"></a><a name="p1252901172"></a>TLS的证书配置。</p>
</td>
<td class="cellrowborder" valign="top" width="30.04%" headers="mcps1.1.5.1.4 "><p id="p05216091710"><a name="p05216091710"></a><a name="p05216091710"></a>默认为空，当开启TLS认证时，需要配置证书信息，具体字段应以键值对形式组织。具体配置指导见<a href="导入TLS证书.md">导入TLS证书</a>。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001976021314_section8785165291317"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

