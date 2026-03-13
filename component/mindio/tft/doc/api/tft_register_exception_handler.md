# tft\_register\_exception\_handler<a name="ZH-CN_TOPIC_0000002514082514"></a>

## 接口功能<a name="zh-cn_topic_0000001975861586_section1346094194320"></a>

注册异常处理程序。

## 接口格式<a name="zh-cn_topic_0000001975861586_section1272104915439"></a>

```
mindio_ttp.framework_ttp.tft_register_exception_handler(fault_pattern: str, fault_type: str, fault_handle: Callable)
```

## 接口参数<a name="zh-cn_topic_0000001975861586_section1148355114320"></a>

<a name="zh-cn_topic_0000001975861586_table173251148163716"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001975861586_row14325848123711"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001975861586_p6325184810378"><a name="zh-cn_topic_0000001975861586_p6325184810378"></a><a name="zh-cn_topic_0000001975861586_p6325184810378"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="14.610000000000001%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001975861586_p1132534853714"><a name="zh-cn_topic_0000001975861586_p1132534853714"></a><a name="zh-cn_topic_0000001975861586_p1132534853714"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="35.39%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001975861586_p17325144811374"><a name="zh-cn_topic_0000001975861586_p17325144811374"></a><a name="zh-cn_topic_0000001975861586_p17325144811374"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001975861586_p1325048183713"><a name="zh-cn_topic_0000001975861586_p1325048183713"></a><a name="zh-cn_topic_0000001975861586_p1325048183713"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row364941564412"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p1129417476504"><a name="p1129417476504"></a><a name="p1129417476504"></a>fault_pattern</p>
</td>
<td class="cellrowborder" valign="top" width="14.610000000000001%" headers="mcps1.1.5.1.2 "><p id="p15292947105015"><a name="p15292947105015"></a><a name="p15292947105015"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="35.39%" headers="mcps1.1.5.1.3 "><p id="p1529013472508"><a name="p1529013472508"></a><a name="p1529013472508"></a>异常关键字。用于匹配精确匹配异常类型。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p8316102393313"><a name="p8316102393313"></a><a name="p8316102393313"></a>异常信息中的关键字字符串。</p>
</td>
</tr>
<tr id="row592195116184"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p1595454771314"><a name="p1595454771314"></a><a name="p1595454771314"></a>fault_type</p>
</td>
<td class="cellrowborder" valign="top" width="14.610000000000001%" headers="mcps1.1.5.1.2 "><p id="p224914851811"><a name="p224914851811"></a><a name="p224914851811"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="35.39%" headers="mcps1.1.5.1.3 "><p id="p2249048181819"><a name="p2249048181819"></a><a name="p2249048181819"></a>异常类型，用于在捕获对应的异常时，与与fault_handle的返回值一起在MindIO上报异常信息时使用。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p4316183018510"><a name="p4316183018510"></a><a name="p4316183018510"></a>字符串，取值范围如下（详情请参见<a href="ReportState.md">ReportState</a>）：</p>
<a name="ul154519138216"></a><a name="ul154519138216"></a><ul id="ul154519138216"><li>RS_NORMAL</li><li>RS_RETRY</li><li>RS_UCE</li><li>RS_UCE_CORRUPTED</li><li>RS_HCCL_FAILED</li><li>RS_INIT_FINISH</li><li>RS_PREREPAIR_FINISH</li><li>RS_STEP_FINISH</li><li>RS_UNKNOWN</li></ul>
</td>
</tr>
<tr id="row182781554141812"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p1559317516134"><a name="p1559317516134"></a><a name="p1559317516134"></a>fault_handle</p>
</td>
<td class="cellrowborder" valign="top" width="14.610000000000001%" headers="mcps1.1.5.1.2 "><p id="p17219053191811"><a name="p17219053191811"></a><a name="p17219053191811"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="35.39%" headers="mcps1.1.5.1.3 "><p id="p1521955371818"><a name="p1521955371818"></a><a name="p1521955371818"></a>异常处理方法，用于接收异常信息字符串，并返回一个字符串。该返回值与fault_type一起在上报异常信息时使用。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p821985314188"><a name="p821985314188"></a><a name="p821985314188"></a>可执行的方法，该方法需要接收异常字符串，并且返回值为字符串。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861586_section1777516402588"></a>

无返回值。

