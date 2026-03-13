# tft\_register\_save\_ckpt\_handler<a name="ZH-CN_TOPIC_0000002511561184"></a>

## 接口功能<a name="zh-cn_topic_0000001975861574_section1186682416267"></a>

注册框架侧dump回调函数。

>**说明：** 
>对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

## 接口格式<a name="zh-cn_topic_0000001975861574_section425913537262"></a>

```
mindio_ttp.framework_ttp.tft_register_save_ckpt_handler(func: Callable, ctx = None)
```

## 接口参数<a name="zh-cn_topic_0000001975861574_section1878744619272"></a>

<a name="zh-cn_topic_0000001975861574_table5690105618279"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001975861574_row146900565278"><th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001975861574_p166901756172719"><a name="zh-cn_topic_0000001975861574_p166901756172719"></a><a name="zh-cn_topic_0000001975861574_p166901756172719"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="20%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001975861574_p869035652719"><a name="zh-cn_topic_0000001975861574_p869035652719"></a><a name="zh-cn_topic_0000001975861574_p869035652719"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001975861574_p069095611271"><a name="zh-cn_topic_0000001975861574_p069095611271"></a><a name="zh-cn_topic_0000001975861574_p069095611271"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001975861574_p8690356122714"><a name="zh-cn_topic_0000001975861574_p8690356122714"></a><a name="zh-cn_topic_0000001975861574_p8690356122714"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001975861574_row11690105610273"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="zh-cn_topic_0000001975861574_p46901256142718"><a name="zh-cn_topic_0000001975861574_p46901256142718"></a><a name="zh-cn_topic_0000001975861574_p46901256142718"></a>func</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001975861574_p126901556192712"><a name="zh-cn_topic_0000001975861574_p126901556192712"></a><a name="zh-cn_topic_0000001975861574_p126901556192712"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.3 "><p id="p15381656134213"><a name="p15381656134213"></a><a name="p15381656134213"></a>临终Checkpoint保存函数，完成保存临终Checkpoint的功能。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.1.5.1.4 "><p id="zh-cn_topic_0000001975861574_p06901456172711"><a name="zh-cn_topic_0000001975861574_p06901456172711"></a><a name="zh-cn_topic_0000001975861574_p06901456172711"></a>回调函数，不为空，回调函数的入参要求请参见<a href="#table1543916174616">表1</a>，约定该回调函数无返回值，执行失败抛出异常。</p>
</td>
</tr>
<tr id="row1850421104012"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.1.5.1.1 "><p id="p885015213404"><a name="p885015213404"></a><a name="p885015213404"></a>ctx</p>
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
<tbody><tr id="row743917171568"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p17439117667"><a name="p17439117667"></a><a name="p17439117667"></a>step</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p124390171366"><a name="p124390171366"></a><a name="p124390171366"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p843919171362"><a name="p843919171362"></a><a name="p843919171362"></a>dump优化器数据时对应的step。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p17440917661"><a name="p17440917661"></a><a name="p17440917661"></a>正整数。</p>
</td>
</tr>
<tr id="row124401317666"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p6440121719612"><a name="p6440121719612"></a><a name="p6440121719612"></a>save_info</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p94409174613"><a name="p94409174613"></a><a name="p94409174613"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p1163285072011"><a name="p1163285072011"></a><a name="p1163285072011"></a>不同优化器参与保存临终遗言时的rank list，其中每个元素是一个字典，字典按照ATTENTION（0）、MOE（1）的索引顺序排列。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><pre class="screen" id="screen119319521419"><a name="screen119319521419"></a><a name="screen119319521419"></a>[
{
"type": int，优化器类型
"ranks": list，参与对应优化器保存临终遗言时的rank列表
},
]</pre>
</td>
</tr>
<tr id="row17371029114418"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p8737182964412"><a name="p8737182964412"></a><a name="p8737182964412"></a>args</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p3737229124420"><a name="p3737229124420"></a><a name="p3737229124420"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p1173711298449"><a name="p1173711298449"></a><a name="p1173711298449"></a>tft_set_step_args设置的参数。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p17737142910443"><a name="p17737142910443"></a><a name="p17737142910443"></a>由注册方决定。</p>
</td>
</tr>
<tr id="row288714335442"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.1 "><p id="p128871833174415"><a name="p128871833174415"></a><a name="p128871833174415"></a>ctx</p>
</td>
<td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.2 "><p id="p18871133174412"><a name="p18871133174412"></a><a name="p18871133174412"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p18887143394417"><a name="p18887143394417"></a><a name="p18887143394417"></a>回调函数上下文。</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p76391118110"><a name="p76391118110"></a><a name="p76391118110"></a>由注册方决定。</p>
</td>
</tr>
</tbody>
</table>

## 返回值<a name="zh-cn_topic_0000001975861574_section24422154297"></a>

无返回值。出错时会打印ERROR日志并抛出异常。

