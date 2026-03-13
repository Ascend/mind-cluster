# RepairType<a name="ZH-CN_TOPIC_0000002511561192"></a>

## 接口功能<a name="zh-cn_topic_0000001975861594_section106131225152112"></a>

定义修复类型枚举。

## 接口格式<a name="zh-cn_topic_0000001975861594_section5166205542118"></a>

```
mindio_ttp.framework_ttp.RepairType
```

## 接口参数<a name="zh-cn_topic_0000001975861594_section28331950124012"></a>

<a name="zh-cn_topic_0000001975861594_table191585974013"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001975861594_row171517595400"><th class="cellrowborder" valign="top" width="19.96%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001975861594_p915759154011"><a name="zh-cn_topic_0000001975861594_p915759154011"></a><a name="zh-cn_topic_0000001975861594_p915759154011"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20.04%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001975861594_p21525904019"><a name="zh-cn_topic_0000001975861594_p21525904019"></a><a name="zh-cn_topic_0000001975861594_p21525904019"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001975861594_p1915175994014"><a name="zh-cn_topic_0000001975861594_p1915175994014"></a><a name="zh-cn_topic_0000001975861594_p1915175994014"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001975861594_p1815159104017"><a name="zh-cn_topic_0000001975861594_p1815159104017"></a><a name="zh-cn_topic_0000001975861594_p1815159104017"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001975861594_row151545954011"><td class="cellrowborder" valign="top" width="19.96%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001975861594_p1415175914017"><a name="zh-cn_topic_0000001975861594_p1415175914017"></a><a name="zh-cn_topic_0000001975861594_p1415175914017"></a>RepairType</p>
</td>
<td class="cellrowborder" valign="top" width="20.04%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001975861594_p815125916409"><a name="zh-cn_topic_0000001975861594_p815125916409"></a><a name="zh-cn_topic_0000001975861594_p815125916409"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001975861594_p41505912408"><a name="zh-cn_topic_0000001975861594_p41505912408"></a><a name="zh-cn_topic_0000001975861594_p41505912408"></a>区分修复类型：</p>
<a name="ul7946141919598"></a><a name="ul7946141919598"></a><ul id="ul7946141919598"><li>RT_SEND：备份卡发送数据。</li><li>RT_UCE_HIGHLEVEL：故障卡需要优化器和模型重建。</li><li>RT_UCE_LOWLEVEL：故障卡不需要优化器和模型重建。</li><li>RT_ROLLBACK：回滚数据集。</li><li>RT_RECV_REPAIR：ARF新拉起卡接收数据。</li><li>RT_LOAD_CKPT：周期Checkpoint数据修复。</li><li>RT_LOAD_REBUILD：重建模型优化器周期Checkpoint数据修复。</li></ul>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><a name="ul386410221592"></a><a name="ul386410221592"></a><ul id="ul386410221592"><li>RT_SEND.value：ttp_c2python_api.RepairType_RT_SEND。</li><li>RT_UCE_HIGHLEVEL.value：ttp_c2python_api.RepairType_RT_UCE_HIGHLEVEL。</li><li>RT_UCE_LOWLEVEL.value：ttp_c2python_api.RepairType_RT_UCE_LOWLEVEL。</li><li>RT_ROLLBACK.value：ttp_c2python_api.RepairType_RT_ROLLBACK。</li><li>RT_RECV_REPAIR.value：ttp_c2python_api.RepairType_RT_RECV_REPAIR。</li><li>RT_LOAD_CKPT.value：ttp_c2python_api.RepairType_RT_LOAD_CKPT。</li><li>RT_LOAD_REBUILD.value：ttp_c2python_api.RepairType_RT_LOAD_REBUILD。</li></ul>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

无

