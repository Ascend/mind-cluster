# tft\_notify\_controller\_stop\_train<a name="ZH-CN_TOPIC_0000002543281123"></a>

## 接口功能<a name="zh-cn_topic_0000001975861586_section1346094194320"></a>

提供给MindCluster调用，通知MindIO TFT主动停止训练，并告知MindIO TFT发生故障的卡信息。

## 接口格式<a name="zh-cn_topic_0000001975861586_section1272104915439"></a>

```
mindio_ttp.controller_ttp.tft_notify_controller_stop_train(fault_ranks: dict, stop_type: str = "stop", timeout: int = None)
```

## 接口参数<a name="zh-cn_topic_0000001975861586_section1148355114320"></a>

<a name="zh-cn_topic_0000001975861586_table173251148163716"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001975861586_row14325848123711"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001975861586_p6325184810378"><a name="zh-cn_topic_0000001975861586_p6325184810378"></a><a name="zh-cn_topic_0000001975861586_p6325184810378"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001975861586_p1132534853714"><a name="zh-cn_topic_0000001975861586_p1132534853714"></a><a name="zh-cn_topic_0000001975861586_p1132534853714"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001975861586_p17325144811374"><a name="zh-cn_topic_0000001975861586_p17325144811374"></a><a name="zh-cn_topic_0000001975861586_p17325144811374"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001975861586_p1325048183713"><a name="zh-cn_topic_0000001975861586_p1325048183713"></a><a name="zh-cn_topic_0000001975861586_p1325048183713"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row364941564412"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p1129417476504"><a name="p1129417476504"></a><a name="p1129417476504"></a>fault_ranks</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p15292947105015"><a name="p15292947105015"></a><a name="p15292947105015"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p1529013472508"><a name="p1529013472508"></a><a name="p1529013472508"></a>发生故障的卡信息。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p1949974519218"><a name="p1949974519218"></a><a name="p1949974519218"></a>&lt;int key, int errorType&gt;字典：</p>
<a name="ul736975113211"></a><a name="ul736975113211"></a><ul id="ul736975113211"><li>key为故障卡的rank号</li><li>errorType为故障类型：<a name="ul16212145011220"></a><a name="ul16212145011220"></a><ul id="ul16212145011220"><li>0：UCE故障</li><li>1：非UCE故障</li></ul>
</li></ul>
</td>
</tr>
<tr id="row377293264712"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p10801318101314"><a name="p10801318101314"></a><a name="p10801318101314"></a>stop_type</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p14801618101311"><a name="p14801618101311"></a><a name="p14801618101311"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p1980111187137"><a name="p1980111187137"></a><a name="p1980111187137"></a>停止训练的类型。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p880191881315"><a name="p880191881315"></a><a name="p880191881315"></a>字符串，支持以下两种方式：</p>
<a name="ul4520162184819"></a><a name="ul4520162184819"></a><ul id="ul4520162184819"><li>"stop"：暂停训练，taskabort方式。</li><li>"pause"：暂停训练，非taskabort方式。</li></ul>
</td>
</tr>
<tr id="row1877253215478"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p51331924181311"><a name="p51331924181311"></a><a name="p51331924181311"></a>timeout</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p9133924181311"><a name="p9133924181311"></a><a name="p9133924181311"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p1813392417136"><a name="p1813392417136"></a><a name="p1813392417136"></a>暂停训练之后等待<span id="ph1945563021510"><a name="ph1945563021510"></a><a name="ph1945563021510"></a>MindCluster</span>做下一步通知的超时时间。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p1133924121319"><a name="p1133924121319"></a><a name="p1133924121319"></a>非负整数，单位：s。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

-   0：调用成功
-   1：调用失败

