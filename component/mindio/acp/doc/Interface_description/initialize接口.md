# initialize接口<a name="ZH-CN_TOPIC_0000002468994016"></a>

## 接口功能<a name="zh-cn_topic_0000001671257765_section190115624413"></a>

初始化MindIO ACP  Client。

## 接口格式<a name="zh-cn_topic_0000001671257765_section4717142204511"></a>

```
mindio_acp.initialize(server_info: Dict[str, str] = None) -> int
```

## 接口参数<a name="zh-cn_topic_0000001671257765_section1232552884519"></a>

<a name="zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="15.310000000000002%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="11.17%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="47.199999999999996%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="26.32%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001671257765_row6247116171310"><td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.1.5.1.1 "><p id="p113710115016"><a name="p113710115016"></a><a name="p113710115016"></a>server_info</p>
</td>
<td class="cellrowborder" valign="top" width="11.17%" headers="mcps1.1.5.1.2 "><p id="zh-cn_topic_0000001671257765_p1265420391361"><a name="zh-cn_topic_0000001671257765_p1265420391361"></a><a name="zh-cn_topic_0000001671257765_p1265420391361"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="47.199999999999996%" headers="mcps1.1.5.1.3 "><p id="zh-cn_topic_0000001671257765_p13247141681315"><a name="zh-cn_topic_0000001671257765_p13247141681315"></a><a name="zh-cn_topic_0000001671257765_p13247141681315"></a>自启动的Server进程需要配置参数信息。若不传入该参数，则全部使用默认值。</p>
</td>
<td class="cellrowborder" valign="top" width="26.32%" headers="mcps1.1.5.1.4 "><p id="p19198120201018"><a name="p19198120201018"></a><a name="p19198120201018"></a>有效参数集合或None。</p>
</td>
</tr>
</tbody>
</table>

**表 1**  server\_info参数说明

<a name="table89621535104115"></a>
<table><thead align="left"><tr id="row89621435164112"><th class="cellrowborder" valign="top" width="18.13%" id="mcps1.2.6.1.1"><p id="p1296218351414"><a name="p1296218351414"></a><a name="p1296218351414"></a>参数key</p>
</th>
<th class="cellrowborder" valign="top" width="12.75%" id="mcps1.2.6.1.2"><p id="p14744193611425"><a name="p14744193611425"></a><a name="p14744193611425"></a>默认参数value</p>
</th>
<th class="cellrowborder" valign="top" width="9.5%" id="mcps1.2.6.1.3"><p id="p173891952436"><a name="p173891952436"></a><a name="p173891952436"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="40.61%" id="mcps1.2.6.1.4"><p id="p696214355414"><a name="p696214355414"></a><a name="p696214355414"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="19.009999999999998%" id="mcps1.2.6.1.5"><p id="p14962123515414"><a name="p14962123515414"></a><a name="p14962123515414"></a>取值范围</p>
</th>
</tr>
</thead>
<tbody><tr id="row16963173516416"><td class="cellrowborder" valign="top" width="18.13%" headers="mcps1.2.6.1.1 "><p id="p09631235194119"><a name="p09631235194119"></a><a name="p09631235194119"></a>'memfs.data_block_pool_capacity_in_gb'</p>
</td>
<td class="cellrowborder" valign="top" width="12.75%" headers="mcps1.2.6.1.2 "><p id="p5963163564115"><a name="p5963163564115"></a><a name="p5963163564115"></a>'128'</p>
</td>
<td class="cellrowborder" valign="top" width="9.5%" headers="mcps1.2.6.1.3 "><p id="p896313513414"><a name="p896313513414"></a><a name="p896313513414"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="40.61%" headers="mcps1.2.6.1.4 "><p id="p092210172815"><a name="p092210172815"></a><a name="p092210172815"></a><span id="ph623218378171"><a name="ph623218378171"></a><a name="ph623218378171"></a>MindIO ACP</span>文件系统内存分配大小，单位：GB，根据服务器内存大小来配置，建议不超过系统总内存的25%。</p>
</td>
<td class="cellrowborder" valign="top" width="19.009999999999998%" headers="mcps1.2.6.1.5 "><p id="p39251252213"><a name="p39251252213"></a><a name="p39251252213"></a>[1, 1024]</p>
</td>
</tr>
<tr id="row99638355413"><td class="cellrowborder" valign="top" width="18.13%" headers="mcps1.2.6.1.1 "><p id="p3567191417480"><a name="p3567191417480"></a><a name="p3567191417480"></a>'memfs.data_block_size_in_mb'</p>
</td>
<td class="cellrowborder" valign="top" width="12.75%" headers="mcps1.2.6.1.2 "><p id="p11963163564111"><a name="p11963163564111"></a><a name="p11963163564111"></a>'128'</p>
</td>
<td class="cellrowborder" valign="top" width="9.5%" headers="mcps1.2.6.1.3 "><p id="p10963935194111"><a name="p10963935194111"></a><a name="p10963935194111"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="40.61%" headers="mcps1.2.6.1.4 "><p id="p29281016281"><a name="p29281016281"></a><a name="p29281016281"></a>文件数据块分配最小粒度，单位：MB，根据使用场景中大多数文件的size决定配置，建议平均每个文件的数据块大小不超过128MB。</p>
</td>
<td class="cellrowborder" valign="top" width="19.009999999999998%" headers="mcps1.2.6.1.5 "><p id="p19260292213"><a name="p19260292213"></a><a name="p19260292213"></a>[1, 1024]</p>
</td>
</tr>
<tr id="row59634357412"><td class="cellrowborder" valign="top" width="18.13%" headers="mcps1.2.6.1.1 "><p id="p96541910161112"><a name="p96541910161112"></a><a name="p96541910161112"></a>'memfs.write.parallel.enabled'</p>
</td>
<td class="cellrowborder" valign="top" width="12.75%" headers="mcps1.2.6.1.2 "><p id="p12222131710118"><a name="p12222131710118"></a><a name="p12222131710118"></a>'true'</p>
</td>
<td class="cellrowborder" valign="top" width="9.5%" headers="mcps1.2.6.1.3 "><p id="p14963135114118"><a name="p14963135114118"></a><a name="p14963135114118"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="40.61%" headers="mcps1.2.6.1.4 "><p id="p398530111310"><a name="p398530111310"></a><a name="p398530111310"></a><span id="ph1659218427175"><a name="ph1659218427175"></a><a name="ph1659218427175"></a>MindIO ACP</span>并发读写性能优化开关配置，用户需结合业务数据模型特征决定是否打开本配置。</p>
</td>
<td class="cellrowborder" valign="top" width="19.009999999999998%" headers="mcps1.2.6.1.5 "><a name="ul198471539563"></a><a name="ul198471539563"></a><ul id="ul198471539563"><li>false：关闭</li><li>true：开启</li></ul>
</td>
</tr>
<tr id="row1987710127503"><td class="cellrowborder" valign="top" width="18.13%" headers="mcps1.2.6.1.1 "><p id="p1638672421119"><a name="p1638672421119"></a><a name="p1638672421119"></a>'memfs.write.parallel.thread_num'</p>
</td>
<td class="cellrowborder" valign="top" width="12.75%" headers="mcps1.2.6.1.2 "><p id="p1898813021110"><a name="p1898813021110"></a><a name="p1898813021110"></a>'16'</p>
</td>
<td class="cellrowborder" valign="top" width="9.5%" headers="mcps1.2.6.1.3 "><p id="p168789127503"><a name="p168789127503"></a><a name="p168789127503"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="40.61%" headers="mcps1.2.6.1.4 "><p id="p2453153931317"><a name="p2453153931317"></a><a name="p2453153931317"></a><span id="ph1827175111712"><a name="ph1827175111712"></a><a name="ph1827175111712"></a>MindIO ACP</span>并发读写性能优化并发数。</p>
</td>
<td class="cellrowborder" valign="top" width="19.009999999999998%" headers="mcps1.2.6.1.5 "><p id="p154531339111311"><a name="p154531339111311"></a><a name="p154531339111311"></a>[2, 96]</p>
</td>
</tr>
<tr id="row17083921117"><td class="cellrowborder" valign="top" width="18.13%" headers="mcps1.2.6.1.1 "><p id="p1770439181120"><a name="p1770439181120"></a><a name="p1770439181120"></a>'memfs.write.parallel.slice_in_mb'</p>
</td>
<td class="cellrowborder" valign="top" width="12.75%" headers="mcps1.2.6.1.2 "><p id="p1870739181111"><a name="p1870739181111"></a><a name="p1870739181111"></a>'16'</p>
</td>
<td class="cellrowborder" valign="top" width="9.5%" headers="mcps1.2.6.1.3 "><p id="p4294178141216"><a name="p4294178141216"></a><a name="p4294178141216"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="40.61%" headers="mcps1.2.6.1.4 "><p id="p5891745131312"><a name="p5891745131312"></a><a name="p5891745131312"></a><span id="ph141051455181710"><a name="ph141051455181710"></a><a name="ph141051455181710"></a>MindIO ACP</span>并发写性能优化数据切分粒度，单位：MB。</p>
</td>
<td class="cellrowborder" valign="top" width="19.009999999999998%" headers="mcps1.2.6.1.5 "><p id="p1989164511130"><a name="p1989164511130"></a><a name="p1989164511130"></a>[1, 1024]</p>
</td>
</tr>
<tr id="row1786918469197"><td class="cellrowborder" valign="top" width="18.13%" headers="mcps1.2.6.1.1 "><p id="p686934612199"><a name="p686934612199"></a><a name="p686934612199"></a>'background.backup.thread_num'</p>
</td>
<td class="cellrowborder" valign="top" width="12.75%" headers="mcps1.2.6.1.2 "><p id="p12869154611193"><a name="p12869154611193"></a><a name="p12869154611193"></a>'32'</p>
</td>
<td class="cellrowborder" valign="top" width="9.5%" headers="mcps1.2.6.1.3 "><p id="p4121440172113"><a name="p4121440172113"></a><a name="p4121440172113"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="40.61%" headers="mcps1.2.6.1.4 "><p id="p139317107286"><a name="p139317107286"></a><a name="p139317107286"></a>备份线程数量。</p>
</td>
<td class="cellrowborder" valign="top" width="19.009999999999998%" headers="mcps1.2.6.1.5 "><p id="p1392620242213"><a name="p1392620242213"></a><a name="p1392620242213"></a>[1, 256]</p>
</td>
</tr>
</tbody>
</table>

>**说明：**
>mindio\_acp.initialize如果不传入server\_info参数，则按照表中默认参数启动Server。

## 使用样例1<a name="zh-cn_topic_0000001671257765_section5115161344717"></a>

```
>>> # Initialize with default param
>>> mindio_acp.initialize()
```

## 使用样例2<a name="section1280112065715"></a>

```
>>> # Initialize with server_info
>>> server_info = {
        'memfs.data_block_pool_capacity_in_gb': '200',
    }
>>> mindio_acp.initialize(server_info=server_info)
```

## 返回值<a name="section88014005720"></a>

-   0：成功
-   -1：失败
