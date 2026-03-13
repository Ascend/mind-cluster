# tft\_register\_rebuild\_group\_handler<a name="ZH-CN_TOPIC_0000002511561160"></a>

## 接口功能<a name="zh-cn_topic_0000001976021318_section719820274302"></a>

注册MindIO ARF重新建组的回调函数。

>**说明：** 
>对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

## 接口格式<a name="zh-cn_topic_0000001976021318_section14612105214308"></a>

```
mindio_ttp.framework_ttp.tft_register_rebuild_group_handler(func: Callable, ctx = None)
```

## 接口参数<a name="zh-cn_topic_0000001976021318_section228315213599"></a>

<a name="zh-cn_topic_0000001976021318_table5690105618279"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001976021318_row146900565278"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001976021318_p166901756172719"><a name="zh-cn_topic_0000001976021318_p166901756172719"></a><a name="zh-cn_topic_0000001976021318_p166901756172719"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001976021318_p869035652719"><a name="zh-cn_topic_0000001976021318_p869035652719"></a><a name="zh-cn_topic_0000001976021318_p869035652719"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001976021318_p069095611271"><a name="zh-cn_topic_0000001976021318_p069095611271"></a><a name="zh-cn_topic_0000001976021318_p069095611271"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001976021318_p8690356122714"><a name="zh-cn_topic_0000001976021318_p8690356122714"></a><a name="zh-cn_topic_0000001976021318_p8690356122714"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001976021318_row11690105610273"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001976021318_p46901256142718"><a name="zh-cn_topic_0000001976021318_p46901256142718"></a><a name="zh-cn_topic_0000001976021318_p46901256142718"></a>func</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001976021318_p126901556192712"><a name="zh-cn_topic_0000001976021318_p126901556192712"></a><a name="zh-cn_topic_0000001976021318_p126901556192712"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p11580182493914"><a name="p11580182493914"></a><a name="p11580182493914"></a><span id="ph9475122584810"><a name="ph9475122584810"></a><a name="ph9475122584810"></a>MindIO ARF</span>重新建组的回调函数，完成正常节点与重启节点清理旧通信组并重建新通信组的功能。</p>
<p id="p73335815147"><a name="p73335815147"></a><a name="p73335815147"></a>回调函数执行超时时间默认为180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001976021318_p06901456172711"><a name="zh-cn_topic_0000001976021318_p06901456172711"></a><a name="zh-cn_topic_0000001976021318_p06901456172711"></a>回调函数，不为空，回调函数的入参要求请参见<a href="#table1543916174616">表1</a>，约定该回调函数无返回值，执行失败抛出异常。</p>
</td>
</tr>
<tr id="row10231729163810"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p123216291384"><a name="p123216291384"></a><a name="p123216291384"></a>ctx</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p198501621194011"><a name="p198501621194011"></a><a name="p198501621194011"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p20850421184019"><a name="p20850421184019"></a><a name="p20850421184019"></a>回调函数上下文。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="p1985092116406"><a name="p1985092116406"></a><a name="p1985092116406"></a>默认为空。</p>
</td>
</tr>
</tbody>
</table>

**表 1**  回调函数参数

<a name="table1543916174616"></a>
<table><thead align="left"><tr id="row04391172064"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.1"><p id="p943912177615"><a name="p943912177615"></a><a name="p943912177615"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.2"><p id="p143914170610"><a name="p143914170610"></a><a name="p143914170610"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.3"><p id="p04391317161"><a name="p04391317161"></a><a name="p04391317161"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p3439131716610"><a name="p3439131716610"></a><a name="p3439131716610"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row743917171568"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p17439117667"><a name="p17439117667"></a><a name="p17439117667"></a>fault_ranks</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p124390171366"><a name="p124390171366"></a><a name="p124390171366"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p843919171362"><a name="p843919171362"></a><a name="p843919171362"></a>故障卡集合。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p17440917661"><a name="p17440917661"></a><a name="p17440917661"></a>list。</p>
</td>
</tr>
<tr id="row1625714496358"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p8737182964412"><a name="p8737182964412"></a><a name="p8737182964412"></a>args</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p3737229124420"><a name="p3737229124420"></a><a name="p3737229124420"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p1173711298449"><a name="p1173711298449"></a><a name="p1173711298449"></a>tft_set_step_args设置的参数。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1481616464339"><a name="p1481616464339"></a><a name="p1481616464339"></a>由注册方决定。</p>
</td>
</tr>
<tr id="row288714335442"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p128871833174415"><a name="p128871833174415"></a><a name="p128871833174415"></a>ctx</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1602161014390"><a name="p1602161014390"></a><a name="p1602161014390"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p14602121013912"><a name="p14602121013912"></a><a name="p14602121013912"></a>回调函数上下文。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1388743344416"><a name="p1388743344416"></a><a name="p1388743344416"></a>由注册方决定。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001976021318_section16811972329"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

