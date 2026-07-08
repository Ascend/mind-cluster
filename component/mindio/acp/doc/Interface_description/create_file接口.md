# create\_file接口<a name="ZH-CN_TOPIC_0000002502153947"></a>

本接口只支持MindSpore框架。

## 接口功能<a name="zh-cn_topic_0000002112429502_section107101141937"></a>

使用with调用create\_file接口，用于创建文件，并返回对应的\_WriteableFileWrapper实例。该实例提供write\(\)、drop\(\)和close\(\)方法。

-   write：向文件中写入数据。

    ```
    write(self, data: bytes)
    ```

    <a name="table114591927106"></a>
    <table><thead align="left"><tr id="row17459122121018"><th class="cellrowborder" valign="top" width="12.43%" id="mcps1.1.5.1.1"><p id="p745918219107"><a name="p745918219107"></a><a name="p745918219107"></a>参数</p>
    </th>
    <th class="cellrowborder" valign="top" width="22.14%" id="mcps1.1.5.1.2"><p id="p5459729101"><a name="p5459729101"></a><a name="p5459729101"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="32.64%" id="mcps1.1.5.1.3"><p id="p164597215105"><a name="p164597215105"></a><a name="p164597215105"></a>说明</p>
    </th>
    <th class="cellrowborder" valign="top" width="32.79%" id="mcps1.1.5.1.4"><p id="p14598271013"><a name="p14598271013"></a><a name="p14598271013"></a>取值要求</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row645910291013"><td class="cellrowborder" valign="top" width="12.43%" headers="mcps1.1.5.1.1 "><p id="p0459626106"><a name="p0459626106"></a><a name="p0459626106"></a>data</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.14%" headers="mcps1.1.5.1.2 "><p id="p445942101012"><a name="p445942101012"></a><a name="p445942101012"></a>必选</p>
    </td>
    <td class="cellrowborder" valign="top" width="32.64%" headers="mcps1.1.5.1.3 "><p id="p545962131012"><a name="p545962131012"></a><a name="p545962131012"></a>需要写入的对象。</p>
    </td>
    <td class="cellrowborder" valign="top" width="32.79%" headers="mcps1.1.5.1.4 "><p id="p204597291020"><a name="p204597291020"></a><a name="p204597291020"></a>bytes对象。</p>
    </td>
    </tr>
    </tbody>
    </table>

-   drop：删除文件。

    ```
    drop(self)
    ```

-   close：关闭文件。

    该方法在with退出上下文的时候自动调用。

    ```
    close(self)
    ```

## 接口格式<a name="zh-cn_topic_0000002112429502_section13362162011417"></a>

```
mindio_acp.create_file(path: str, mode: int = 0o600)
```

## 接口参数<a name="zh-cn_topic_0000002112429502_section171201830749"></a>

<a name="zh-cn_topic_0000001671257765_table22461616201318"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001671257765_row524621616133"><th class="cellrowborder" valign="top" width="13.13%" id="mcps1.1.5.1.1"><p id="zh-cn_topic_0000001671257765_p19247716161318"><a name="zh-cn_topic_0000001671257765_p19247716161318"></a><a name="zh-cn_topic_0000001671257765_p19247716161318"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="19.81%" id="mcps1.1.5.1.2"><p id="zh-cn_topic_0000001671257765_p10247201671310"><a name="zh-cn_topic_0000001671257765_p10247201671310"></a><a name="zh-cn_topic_0000001671257765_p10247201671310"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="38.46%" id="mcps1.1.5.1.3"><p id="zh-cn_topic_0000001671257765_p9247141618136"><a name="zh-cn_topic_0000001671257765_p9247141618136"></a><a name="zh-cn_topic_0000001671257765_p9247141618136"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="28.599999999999998%" id="mcps1.1.5.1.4"><p id="zh-cn_topic_0000001671257765_p11247121615131"><a name="zh-cn_topic_0000001671257765_p11247121615131"></a><a name="zh-cn_topic_0000001671257765_p11247121615131"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row129665813187"><td class="cellrowborder" valign="top" width="13.13%" headers="mcps1.1.5.1.1 "><p id="p1596610811817"><a name="p1596610811817"></a><a name="p1596610811817"></a>path</p>
</td>
<td class="cellrowborder" valign="top" width="19.81%" headers="mcps1.1.5.1.2 "><p id="p99661989186"><a name="p99661989186"></a><a name="p99661989186"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="38.46%" headers="mcps1.1.5.1.3 "><p id="p5966689182"><a name="p5966689182"></a><a name="p5966689182"></a>数据保存路径。</p>
</td>
<td class="cellrowborder" valign="top" width="28.599999999999998%" headers="mcps1.1.5.1.4 "><p id="p139661484188"><a name="p139661484188"></a><a name="p139661484188"></a>有效文件路径。</p>
</td>
</tr>
<tr id="row185421256195915"><td class="cellrowborder" valign="top" width="13.13%" headers="mcps1.1.5.1.1 "><p id="p354245611599"><a name="p354245611599"></a><a name="p354245611599"></a>mode</p>
</td>
<td class="cellrowborder" valign="top" width="19.81%" headers="mcps1.1.5.1.2 "><p id="p1254218565599"><a name="p1254218565599"></a><a name="p1254218565599"></a>可选</p>
</td>
<td class="cellrowborder" valign="top" width="38.46%" headers="mcps1.1.5.1.3 "><p id="p1542185635912"><a name="p1542185635912"></a><a name="p1542185635912"></a>文件创建权限。</p>
</td>
<td class="cellrowborder" valign="top" width="28.599999999999998%" headers="mcps1.1.5.1.4 "><p id="p15543125616594"><a name="p15543125616594"></a><a name="p15543125616594"></a>[0o000, 0o777]</p>
</td>
</tr>
</tbody>
</table>

## 使用样例<a name="zh-cn_topic_0000002112429502_section81115380412"></a>

```
>>> x = b'\x00\x01\x02\x03\x04'
>>> with mindio_acp.create_file('/mnt/dpc01/checkpoint/rank-0.pt') as f:
...     write_result = f.write(x)
```

## 返回值<a name="zh-cn_topic_0000002112429502_section17538071458"></a>

\_WriteableFileWrapper实例。

>**说明：**
>接口详情请参见[MindSpore文档](https://www.mindspore.cn/docs/zh-CN/master/api_python/mindspore/mindspore.save_checkpoint.html#mindspore.save_checkpoint)。
