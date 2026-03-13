# tft\_exception\_handler<a name="ZH-CN_TOPIC_0000002543281131"></a>

## 接口功能<a name="zh-cn_topic_0000001975861558_section2664124672218"></a>

装饰器，对MindSpeed-LLM的train方法进行装饰，捕获训练状态异常以及上报处理，对于用户的其他训练框架，本接口仅提供参考示例功能。

## 接口格式<a name="zh-cn_topic_0000001975861558_section259611294404"></a>

```
mindio_ttp.framework_ttp.tft_exception_handler(func: Callable)
```

## 接口参数<a name="zh-cn_topic_0000001975861558_section1620335184016"></a>

<a name="zh-cn_topic_0000001975861558_table814272973018"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001975861558_row15142192993018"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001975861558_p16142529193020"><a name="zh-cn_topic_0000001975861558_p16142529193020"></a><a name="zh-cn_topic_0000001975861558_p16142529193020"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001975861558_p5142122933010"><a name="zh-cn_topic_0000001975861558_p5142122933010"></a><a name="zh-cn_topic_0000001975861558_p5142122933010"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001975861558_p714217293309"><a name="zh-cn_topic_0000001975861558_p714217293309"></a><a name="zh-cn_topic_0000001975861558_p714217293309"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001975861558_p14142129153011"><a name="zh-cn_topic_0000001975861558_p14142129153011"></a><a name="zh-cn_topic_0000001975861558_p14142129153011"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001975861558_row514242993017"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001975861558_p121423298309"><a name="zh-cn_topic_0000001975861558_p121423298309"></a><a name="zh-cn_topic_0000001975861558_p121423298309"></a>func</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001975861558_p1314212983012"><a name="zh-cn_topic_0000001975861558_p1314212983012"></a><a name="zh-cn_topic_0000001975861558_p1314212983012"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001975861558_p4142162973014"><a name="zh-cn_topic_0000001975861558_p4142162973014"></a><a name="zh-cn_topic_0000001975861558_p4142162973014"></a>函数作为参数。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001975861558_p181421629173010"><a name="zh-cn_topic_0000001975861558_p181421629173010"></a><a name="zh-cn_topic_0000001975861558_p181421629173010"></a>框架的train方法。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861558_section1777516402588"></a>

装饰器返回的func。

