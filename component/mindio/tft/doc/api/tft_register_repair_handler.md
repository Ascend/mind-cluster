# tft\_register\_repair\_handler<a name="ZH-CN_TOPIC_0000002511721150"></a>

## 接口功能<a name="zh-cn_topic_0000001976021318_section719820274302"></a>

注册repair回调函数。

>**说明：** 
>-   对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。
>-   MindIO TFT已在回调函数中对模型优化器中的变量进行重建与覆写，用户在框架中自定义的其他参与计算的变量，需在repair中自行实现对其的重建与覆写。

## 接口格式<a name="zh-cn_topic_0000001976021318_section14612105214308"></a>

```
mindio_ttp.framework_ttp.tft_register_repair_handler(func: Callable, ctx = None)
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
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p197733810160"><a name="p197733810160"></a><a name="p197733810160"></a>repair回调函数，完成优化器修复等数据修复功能。</p>
<p id="p597511127397"><a name="p597511127397"></a><a name="p597511127397"></a>回调函数执行超时时间默认为180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001976021318_p06901456172711"><a name="zh-cn_topic_0000001976021318_p06901456172711"></a><a name="zh-cn_topic_0000001976021318_p06901456172711"></a>回调函数，不为空，回调函数的入参要求请参见<a href="#table1543916174616">表1</a>，约定该回调函数无返回值，执行失败抛出异常。</p>
</td>
</tr>
<tr id="row595311251434"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p885015213404"><a name="p885015213404"></a><a name="p885015213404"></a>ctx</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="p198501621194011"><a name="p198501621194011"></a><a name="p198501621194011"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p20850421184019"><a name="p20850421184019"></a><a name="p20850421184019"></a>回调上下文。</p>
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
<tbody><tr id="row743917171568"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p17439117667"><a name="p17439117667"></a><a name="p17439117667"></a>step</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p124390171366"><a name="p124390171366"></a><a name="p124390171366"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p843919171362"><a name="p843919171362"></a><a name="p843919171362"></a>修复时对应的step。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p17440917661"><a name="p17440917661"></a><a name="p17440917661"></a>正整数。</p>
</td>
</tr>
<tr id="row129812501141"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p429811507420"><a name="p429811507420"></a><a name="p429811507420"></a>need_rebuild</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p52981550646"><a name="p52981550646"></a><a name="p52981550646"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p1929813501412"><a name="p1929813501412"></a><a name="p1929813501412"></a>修复是否需要重建模型和优化器。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><a name="ul11025321747"></a><a name="ul11025321747"></a><ul id="ul11025321747"><li>False：无需重建。</li><li>True：需要重建。</li></ul>
</td>
</tr>
<tr id="row1412812138515"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p51284130510"><a name="p51284130510"></a><a name="p51284130510"></a>error_ranks</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p4128713357"><a name="p4128713357"></a><a name="p4128713357"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p71280131156"><a name="p71280131156"></a><a name="p71280131156"></a>需要修复的故障卡list。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1112841319515"><a name="p1112841319515"></a><a name="p1112841319515"></a>list。</p>
</td>
</tr>
<tr id="row124401317666"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p6440121719612"><a name="p6440121719612"></a><a name="p6440121719612"></a>repair_info</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p94409174613"><a name="p94409174613"></a><a name="p94409174613"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p7440161710611"><a name="p7440161710611"></a><a name="p7440161710611"></a>修复策略dict，其中优化器类型按照ATTENTION（0）、MOE（1）的关系对应。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><pre class="screen" id="screen177560382815"><a name="screen177560382815"></a><a name="screen177560382815"></a>{
"type": int，优化器类型
"repair_type": Enum，枚举类型取值参考<a href="RepairType.md">5.34-RepairType</a>
"src": list，优化器修复数据的来源卡列表
"dst": list，优化器修复数据的目的卡列表
"rank_list": list，修复通信组建立所需要的卡列表
}</pre>
</td>
</tr>
<tr id="row17371029114418"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p8737182964412"><a name="p8737182964412"></a><a name="p8737182964412"></a>args</p>
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
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p18871133174412"><a name="p18871133174412"></a><a name="p18871133174412"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p18887143394417"><a name="p18887143394417"></a><a name="p18887143394417"></a>回调函数上下文。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1388743344416"><a name="p1388743344416"></a><a name="p1388743344416"></a>由注册方决定。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001976021318_section16811972329"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

