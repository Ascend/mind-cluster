# tft\_set\_step\_args<a name="ZH-CN_TOPIC_0000002543281129"></a>

## 接口功能<a name="zh-cn_topic_0000001975861586_section1346094194320"></a>

训练框架设置的参数集合。

>**说明：** 
>对于MindSpeed-LLM训练框架，设置功能已经由MindIO TFT完成适配，不需要调用。

## 接口格式<a name="zh-cn_topic_0000001975861586_section1272104915439"></a>

```
mindio_ttp.framework_ttp.tft_set_step_args(args)
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
<tbody><tr id="zh-cn_topic_0000001975861586_row6325848193719"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001975861586_p332516489375"><a name="zh-cn_topic_0000001975861586_p332516489375"></a><a name="zh-cn_topic_0000001975861586_p332516489375"></a>args</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001975861586_p1485713265318"><a name="zh-cn_topic_0000001975861586_p1485713265318"></a><a name="zh-cn_topic_0000001975861586_p1485713265318"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001975861586_p10325348163714"><a name="zh-cn_topic_0000001975861586_p10325348163714"></a><a name="zh-cn_topic_0000001975861586_p10325348163714"></a>训练框架设置需要保存的参数集合。<span id="ph19229134314220"><a name="ph19229134314220"></a><a name="ph19229134314220"></a>MindIO TFT</span>在stop/clean/repair/rollback等阶段调用注册的回调函数时，将参数集合传回，框架根据参数集合完成相应功能。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001975861586_p2032534812372"><a name="zh-cn_topic_0000001975861586_p2032534812372"></a><a name="zh-cn_topic_0000001975861586_p2032534812372"></a>由训练框架决定，<span id="ph75244918428"><a name="ph75244918428"></a><a name="ph75244918428"></a>MindIO TFT</span>不访问也不修改该参数集合，在stop/clean/repair/rollback等阶段时调用注册的业务回调将其传回，业务回调负责对取值范围进行校验。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

