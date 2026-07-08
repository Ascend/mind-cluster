# register\_checker接口<a name="ZH-CN_TOPIC_0000002502153981"></a>

## 接口功能<a name="section23222518147"></a>

注册异步回调函数。

## 接口格式<a name="section1585114371516"></a>

```
mindio_acp.register_checker(callback, check_dict, user_context, timeout_sec)
```

## 接口参数<a name="section522915911518"></a>

<a name="zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="16.259999999999998%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="11.51%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="44.16%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="28.07%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="16.259999999999998%" headers="mcps1.1.5.1.1 "><p id="p113710115016"><a name="p113710115016"></a><a name="p113710115016"></a>callback</p>
</td>
<td class="cellrowborder" valign="top" width="11.51%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001671257765_p1265420391361"><a name="zh-cn_topic_0000001671257765_p1265420391361"></a><a name="zh-cn_topic_0000001671257765_p1265420391361"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="44.16%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001671257765_p13247141681315"><a name="zh-cn_topic_0000001671257765_p13247141681315"></a><a name="zh-cn_topic_0000001671257765_p13247141681315"></a>回调函数（第一个参数result为数据完整性校验的结果，0为成功，其他为失败；第二个参数为user_context）。</p>
</td>
<td class="cellrowborder" valign="top" width="28.07%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001671257765_p20482175915402"><a name="zh-cn_topic_0000001671257765_p20482175915402"></a><a name="zh-cn_topic_0000001671257765_p20482175915402"></a>有效函数名。</p>
</td>
</tr>
<tr id="row84334589522"><td class="cellrowborder" valign="top" width="16.259999999999998%" headers="mcps1.1.5.1.1 "><p id="p17434145813523"><a name="p17434145813523"></a><a name="p17434145813523"></a>check_dict</p>
</td>
<td class="cellrowborder" valign="top" width="11.51%" headers="mcps1.1.5.1.2 "><p id="p3434958125212"><a name="p3434958125212"></a><a name="p3434958125212"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="44.16%" headers="mcps1.1.5.1.3 "><p id="p1643405818528"><a name="p1643405818528"></a><a name="p1643405818528"></a>数据完整性校验条件，类型为dict，用来校验指定path下的文件个数是否符合要求。</p>
</td>
<td class="cellrowborder" valign="top" width="28.07%" headers="mcps1.1.5.1.4 "><a name="ul192704523505"></a><a name="ul192704523505"></a><ul id="ul192704523505"><li>key：path，数据路径。</li><li>value：对应key路径下的文件个数。</li></ul>
</td>
</tr>
<tr id="row1289032519197"><td class="cellrowborder" valign="top" width="16.259999999999998%" headers="mcps1.1.5.1.1 "><p id="p1389172561919"><a name="p1389172561919"></a><a name="p1389172561919"></a>user_context</p>
</td>
<td class="cellrowborder" valign="top" width="11.51%" headers="mcps1.1.5.1.2 "><p id="p4891202571916"><a name="p4891202571916"></a><a name="p4891202571916"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="44.16%" headers="mcps1.1.5.1.3 "><p id="p10891142511191"><a name="p10891142511191"></a><a name="p10891142511191"></a>回调函数的第二个参数。</p>
</td>
<td class="cellrowborder" valign="top" width="28.07%" headers="mcps1.1.5.1.4 "><p id="p10891152561912"><a name="p10891152561912"></a><a name="p10891152561912"></a>-</p>
</td>
</tr>
<tr id="row379985321917"><td class="cellrowborder" valign="top" width="16.259999999999998%" headers="mcps1.1.5.1.1 "><p id="p187996531195"><a name="p187996531195"></a><a name="p187996531195"></a>timeout_sec</p>
</td>
<td class="cellrowborder" valign="top" width="11.51%" headers="mcps1.1.5.1.2 "><p id="p1579915351919"><a name="p1579915351919"></a><a name="p1579915351919"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="44.16%" headers="mcps1.1.5.1.3 "><p id="p68001653191912"><a name="p68001653191912"></a><a name="p68001653191912"></a>回调超时时间，单位：秒。</p>
<div class="note" id="note49311948132417"><a name="note49311948132417"></a><a name="note49311948132417"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="p186064413342"><a name="p186064413342"></a><a name="p186064413342"></a>如果训练客户端日志中提示："watching checkpoint failed"，则需要调大该参数。</p>
<p id="p16931134819241"><a name="p16931134819241"></a><a name="p16931134819241"></a>代码在mindio_acp实际安装路径（<span class="filepath" id="filepath145216262620"><a name="filepath145216262620"></a><a name="filepath145216262620"></a>“mindio_acp/acc_checkpoint/framework_acp.py”</span>）下的async_write_tracker_file函数中。</p>
</div></div>
</td>
<td class="cellrowborder" valign="top" width="28.07%" headers="mcps1.1.5.1.4 "><p id="p188001153181914"><a name="p188001153181914"></a><a name="p188001153181914"></a>[1, 3600]</p>
</td>
</tr>
</tbody>
</table>

## 使用样例<a name="section1550852911320"></a>

```
>>> def callback(result, user_context):
>>>    if result == 0:
>>>        print("success")
>>>    else:
>>>        print("fail")
>>> context_obj = None
>>> check_dict = {'/mnt/dpc01/checkpoint-last': 4}
>>> mindio_acp.register_checker(callback, check_dict, context_obj, 1000)
```

## 返回值<a name="section8785165291317"></a>

-   None：失败。
-   1：成功。
