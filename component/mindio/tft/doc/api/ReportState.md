# ReportState<a name="ZH-CN_TOPIC_0000002511561194"></a>

## 接口功能<a name="zh-cn_topic_0000001975861594_section106131225152112"></a>

装饰器上报训练状态枚举。

## 接口格式<a name="zh-cn_topic_0000001975861594_section5166205542118"></a>

```
mindio_ttp.framework_ttp.ReportState
```

## 接口参数<a name="zh-cn_topic_0000001975861594_section28331950124012"></a>

<a name="zh-cn_topic_0000001975861594_table191585974013"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001975861594_row171517595400"><th class="cellrowborder" valign="top" width="17.23%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001975861594_p915759154011"><a name="zh-cn_topic_0000001975861594_p915759154011"></a><a name="zh-cn_topic_0000001975861594_p915759154011"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="14.62%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001975861594_p21525904019"><a name="zh-cn_topic_0000001975861594_p21525904019"></a><a name="zh-cn_topic_0000001975861594_p21525904019"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="37.1%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001975861594_p1915175994014"><a name="zh-cn_topic_0000001975861594_p1915175994014"></a><a name="zh-cn_topic_0000001975861594_p1915175994014"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="31.05%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001975861594_p1815159104017"><a name="zh-cn_topic_0000001975861594_p1815159104017"></a><a name="zh-cn_topic_0000001975861594_p1815159104017"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001975861594_row151545954011"><td class="cellrowborder" valign="top" width="17.23%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001975861594_p1415175914017"><a name="zh-cn_topic_0000001975861594_p1415175914017"></a><a name="zh-cn_topic_0000001975861594_p1415175914017"></a>ReportState</p>
</td>
<td class="cellrowborder" valign="top" width="14.62%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001975861594_p815125916409"><a name="zh-cn_topic_0000001975861594_p815125916409"></a><a name="zh-cn_topic_0000001975861594_p815125916409"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="37.1%" headers="mcps1.1.5.1.3 "><p id="p1625316435556"><a name="p1625316435556"></a><a name="p1625316435556"></a>区分上报的训练状态类型：</p>
<a name="ul174494217555"></a><a name="ul174494217555"></a><ul id="ul174494217555"><li>RS_NORMAL：正常状态。</li><li>RS_RETRY：精度异常。</li><li>RS_UCE：UCE错误。</li><li>RS_UCE_CORRUPTED：片上内存MULTI BIT ECC故障。</li><li>RS_HCCL_FAILED：HCCL重计算失败。</li><li>RS_UNKNOWN：其他错误。</li><li>RS_INIT_FINISH：在MindSpore框架中，ARF新启动的节点在训练进程完成初始化后抛出的异常。</li><li>RS_PREREPAIR_FINISH：ARF新启动的节点抛出的异常。</li><li>RS_STEP_FINISH：亚健康热切中step级暂停已经完成抛出的异常。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="31.05%" headers="mcps1.1.5.1.4 "><a name="ul20904204220574"></a><a name="ul20904204220574"></a><ul id="ul20904204220574"><li>RS_NORMAL.value：ttp_c2python_api.ReportState_RS_NORMAL。</li><li>RS_RETRY.value：ttp_c2python_api.ReportState_RS_RETRY。</li><li>RS_UCE.value：ttp_c2python_api.ReportState_RS_UCE。</li><li>RS_UCE_CORRUPTED：<p id="p107428543917"><a name="p107428543917"></a><a name="p107428543917"></a>ttp_c2python_api.ReportState_RS_UCE_CORRUPTED。</p>
</li><li>RS_HCCL_FAILED.value: ttp_c2python_api.ReportState_RS_HCCL_FAILED。</li><li>RS_UNKNOWN.value：ttp_c2python_api.ReportState_RS_UNKNOWN。</li><li>RS_INIT_FINISH：<p id="p939015192101"><a name="p939015192101"></a><a name="p939015192101"></a>ttp_c2python_api.ReportState_RS_INIT_FINISH。</p>
</li><li>RS_PREREPAIR_FINISH.value：ttp_c2python_api.ReportState_RS_PREREPAIR_FINISH。</li><li>RS_STEP_FINISH：<p id="p6134848141016"><a name="p6134848141016"></a><a name="p6134848141016"></a>ttp_c2python_api.ReportState_RS_STEP_FINISH。</p>
</li></ul>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

无

