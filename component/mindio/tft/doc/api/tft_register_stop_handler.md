# tft\_register\_stop\_handler<a name="ZH-CN_TOPIC_0000002511561166"></a>

## 接口功能<a name="zh-cn_topic_0000001976021318_section719820274302"></a>

在恢复过程中注册停止训练的回调函数。

>**说明：** 
>对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

## 接口格式<a name="zh-cn_topic_0000001976021318_section14612105214308"></a>

```
mindio_ttp.framework_ttp.tft_register_stop_handler(func: Callable, ctx = None)
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
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001976021318_p1690185632717"><a name="zh-cn_topic_0000001976021318_p1690185632717"></a><a name="zh-cn_topic_0000001976021318_p1690185632717"></a>停止训练的回调函数，实现停止训练的功能，并抛出FORCE STOP异常将训练主线程控制权交由装饰器接管。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001976021318_p06901456172711"><a name="zh-cn_topic_0000001976021318_p06901456172711"></a><a name="zh-cn_topic_0000001976021318_p06901456172711"></a>回调函数，不为空，回调函数的入参要求请参见<a href="#table1543916174616">表1</a>，约定该回调函数无返回值，执行失败抛出异常。</p>
</td>
</tr>
<tr id="row595311251434"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p885015213404"><a name="p885015213404"></a><a name="p885015213404"></a>ctx</p>
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
<tbody><tr id="row288714335442"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p8737182964412"><a name="p8737182964412"></a><a name="p8737182964412"></a>args</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p3737229124420"><a name="p3737229124420"></a><a name="p3737229124420"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p1173711298449"><a name="p1173711298449"></a><a name="p1173711298449"></a>tft_set_step_args设置的参数。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p17737142910443"><a name="p17737142910443"></a><a name="p17737142910443"></a>由注册方决定。</p>
</td>
</tr>
<tr id="row1547432082819"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p156371038102810"><a name="p156371038102810"></a><a name="p156371038102810"></a>ctx</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p1963773817282"><a name="p1963773817282"></a><a name="p1963773817282"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p2637133882812"><a name="p2637133882812"></a><a name="p2637133882812"></a>回调函数上下文。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p7637338192813"><a name="p7637338192813"></a><a name="p7637338192813"></a>由注册方决定。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001976021318_section16811972329"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

