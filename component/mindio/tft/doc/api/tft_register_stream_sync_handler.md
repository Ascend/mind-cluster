# tft\_register\_stream\_sync\_handler<a name="ZH-CN_TOPIC_0000002543281115"></a>

## 接口功能<a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_section719820274302"></a>

注册同步回调函数。

>**说明：** 
>对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

## 接口格式<a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_section14612105214308"></a>

```
mindio_ttp.framework_ttp.tft_register_stream_sync_handler(func: Callable, ctx=None)
```

## 接口参数<a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_section228315213599"></a>

<a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_table5690105618279"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_row146900565278"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p166901756172719"><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p166901756172719"></a><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p166901756172719"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p869035652719"><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p869035652719"></a><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p869035652719"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p069095611271"><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p069095611271"></a><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p069095611271"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p8690356122714"><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p8690356122714"></a><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p8690356122714"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_row11690105610273"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p46901256142718"><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p46901256142718"></a><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p46901256142718"></a>func</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p126901556192712"><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p126901556192712"></a><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p126901556192712"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000002330163997_p1015714219179"><a name="zh-cn_topic_0000002330163997_p1015714219179"></a><a name="zh-cn_topic_0000002330163997_p1015714219179"></a>同步回调函数，完成训练暂停后同步操作。避免在暂停训练后算子队列有残留算子未执行完。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p06901456172711"><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p06901456172711"></a><a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_p06901456172711"></a>回调函数，不为空。回调函数无参数，约定该回调函数无返回值，执行失败抛出异常。</p>
</td>
</tr>
<tr id="row154510916168"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p128871833174415"><a name="p128871833174415"></a><a name="p128871833174415"></a>ctx</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p18871133174412"><a name="p18871133174412"></a><a name="p18871133174412"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p18887143394417"><a name="p18887143394417"></a><a name="p18887143394417"></a>回调函数上下文。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p1388743344416"><a name="p1388743344416"></a><a name="p1388743344416"></a>由注册方决定。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000002330163997_zh-cn_topic_0000001976021318_section16811972329"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

