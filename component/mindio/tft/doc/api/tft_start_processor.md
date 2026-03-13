# tft\_start\_processor<a name="ZH-CN_TOPIC_0000002543281083"></a>

## 接口功能<a name="zh-cn_topic_0000002012581705_section188071357132719"></a>

在初始化Processor模块成功后，调用该接口以启动MindIO TFT  Processor模块服务。

## 接口格式<a name="zh-cn_topic_0000002012581705_section1642012205289"></a>

```
mindio_ttp.framework_ttp.tft_start_processor(master_ip: str, port: int, local_ip='')
```

## 接口参数<a name="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_section1232552884519"></a>

<a name="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002012581705_zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000002012581705_p987113271408"><a name="zh-cn_topic_0000002012581705_p987113271408"></a><a name="zh-cn_topic_0000002012581705_p987113271408"></a>master_ip</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000002012581705_p887182717011"><a name="zh-cn_topic_0000002012581705_p887182717011"></a><a name="zh-cn_topic_0000002012581705_p887182717011"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000002012581705_p118701727407"><a name="zh-cn_topic_0000002012581705_p118701727407"></a><a name="zh-cn_topic_0000002012581705_p118701727407"></a>Controller所在节点IP地址或域名。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000002012581705_p187017276016"><a name="zh-cn_topic_0000002012581705_p187017276016"></a><a name="zh-cn_topic_0000002012581705_p187017276016"></a>符合IP地址规范的IPv4地址，位于集群节点IP地址中，禁止全零IP地址，支持域名。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002012581705_row84334589522"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000002012581705_p487018274013"><a name="zh-cn_topic_0000002012581705_p487018274013"></a><a name="zh-cn_topic_0000002012581705_p487018274013"></a>port</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000002012581705_p108691627208"><a name="zh-cn_topic_0000002012581705_p108691627208"></a><a name="zh-cn_topic_0000002012581705_p108691627208"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000002012581705_p138691427707"><a name="zh-cn_topic_0000002012581705_p138691427707"></a><a name="zh-cn_topic_0000002012581705_p138691427707"></a>Controller侦听端口号。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000002012581705_p686817272020"><a name="zh-cn_topic_0000002012581705_p686817272020"></a><a name="zh-cn_topic_0000002012581705_p686817272020"></a>[1024, 65535]</p>
</td>
</tr>
<tr id="row10420423133019"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p14201823113013"><a name="p14201823113013"></a><a name="p14201823113013"></a>local_ip</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p14420162318302"><a name="p14420162318302"></a><a name="p14420162318302"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p12515911154512"><a name="p12515911154512"></a><a name="p12515911154512"></a>K8s中Processor所在节点的Service IP地址或域名。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p2042032314303"><a name="p2042032314303"></a><a name="p2042032314303"></a>符合IP地址规范的IPv4地址，位于集群节点IP地址中，禁止全零IP地址，支持域名。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000002012581705_section8785165291317"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

