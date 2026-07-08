# open\_file接口<a name="ZH-CN_TOPIC_0000002502153985"></a>

本接口只支持MindSpore框架。

## 接口功能<a name="zh-cn_topic_0000002112429502_section107101141937"></a>

使用with调用open\_file接口，以只读的方式打开文件，并返回对应的\_ReadableFileWrapper实例。该实例提供read\(\)和close\(\)方法。

-   read：读取文件内容。

    ```
    read(self, offset=0, count=-1)
    ```

    <a name="table137338119383"></a>
    <table><thead align="left"><tr id="row147324113813"><th class="cellrowborder" valign="top" width="12.43%" id="mcps1.1.5.1.1"><p id="p473221103815"><a name="p473221103815"></a><a name="p473221103815"></a>参数</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.64%" id="mcps1.1.5.1.2"><p id="p373220163810"><a name="p373220163810"></a><a name="p373220163810"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="41.02%" id="mcps1.1.5.1.3"><p id="p373221173812"><a name="p373221173812"></a><a name="p373221173812"></a>说明</p>
    </th>
    <th class="cellrowborder" valign="top" width="34.910000000000004%" id="mcps1.1.5.1.4"><p id="p1773211123816"><a name="p1773211123816"></a><a name="p1773211123816"></a>取值要求</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row17732615387"><td class="cellrowborder" valign="top" width="12.43%" headers="mcps1.1.5.1.1 "><p id="p7732917386"><a name="p7732917386"></a><a name="p7732917386"></a>offset</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.64%" headers="mcps1.1.5.1.2 "><p id="p273210118385"><a name="p273210118385"></a><a name="p273210118385"></a>可选</p>
    </td>
    <td class="cellrowborder" valign="top" width="41.02%" headers="mcps1.1.5.1.3 "><p id="p157321118384"><a name="p157321118384"></a><a name="p157321118384"></a>读取文件的偏移位置。</p>
    <p id="p16345618135410"><a name="p16345618135410"></a><a name="p16345618135410"></a>需满足count + offset &lt;= file_size</p>
    </td>
    <td class="cellrowborder" valign="top" width="34.910000000000004%" headers="mcps1.1.5.1.4 "><p id="p47323183818"><a name="p47323183818"></a><a name="p47323183818"></a>[0, file_size)</p>
    </td>
    </tr>
    <tr id="row1733913382"><td class="cellrowborder" valign="top" width="12.43%" headers="mcps1.1.5.1.1 "><p id="p1173212183815"><a name="p1173212183815"></a><a name="p1173212183815"></a>count</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.64%" headers="mcps1.1.5.1.2 "><p id="p117323119382"><a name="p117323119382"></a><a name="p117323119382"></a>可选</p>
    </td>
    <td class="cellrowborder" valign="top" width="41.02%" headers="mcps1.1.5.1.3 "><p id="p57324120386"><a name="p57324120386"></a><a name="p57324120386"></a>读取文件的大小。</p>
    <p id="p18586164910547"><a name="p18586164910547"></a><a name="p18586164910547"></a>需满足count + offset &lt;= file_size</p>
    </td>
    <td class="cellrowborder" valign="top" width="34.910000000000004%" headers="mcps1.1.5.1.4 "><a name="ul273316183817"></a><a name="ul273316183817"></a><ul id="ul273316183817"><li>-1：读取整个文件。</li><li>(0, file_size]</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

-   close：关闭文件。

    该方法在with退出上下文的时候自动调用。

    ```
    close(self)
    ```

## 接口格式<a name="zh-cn_topic_0000002112429502_section13362162011417"></a>

```
mindio_acp.open_file(path: str)
```

## 接口参数<a name="zh-cn_topic_0000002112429502_section171201830749"></a>

<a name="table1177375161011"></a>
<table><thead align="left"><tr id="row5773115151012"><th class="cellrowborder" valign="top" width="12.43%" id="mcps1.1.5.1.1"><p id="p177749591014"><a name="p177749591014"></a><a name="p177749591014"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="17.79%" id="mcps1.1.5.1.2"><p id="p87747512109"><a name="p87747512109"></a><a name="p87747512109"></a>是否必选</p>
</th>
<th class="cellrowborder" valign="top" width="42.71%" id="mcps1.1.5.1.3"><p id="p177745561017"><a name="p177745561017"></a><a name="p177745561017"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="27.07%" id="mcps1.1.5.1.4"><p id="p1377455131014"><a name="p1377455131014"></a><a name="p1377455131014"></a>取值要求</p>
</th>
</tr>
</thead>
<tbody><tr id="row1677415571015"><td class="cellrowborder" valign="top" width="12.43%" headers="mcps1.1.5.1.1 "><p id="p197744561019"><a name="p197744561019"></a><a name="p197744561019"></a>path</p>
</td>
<td class="cellrowborder" valign="top" width="17.79%" headers="mcps1.1.5.1.2 "><p id="p17774165101020"><a name="p17774165101020"></a><a name="p17774165101020"></a>必选</p>
</td>
<td class="cellrowborder" valign="top" width="42.71%" headers="mcps1.1.5.1.3 "><p id="p2774154106"><a name="p2774154106"></a><a name="p2774154106"></a>加载路径。</p>
</td>
<td class="cellrowborder" valign="top" width="27.07%" headers="mcps1.1.5.1.4 "><p id="p87741652107"><a name="p87741652107"></a><a name="p87741652107"></a>有效文件路径。</p>
</td>
</tr>
</tbody>
</table>

## 使用样例<a name="zh-cn_topic_0000002112429502_section81115380412"></a>

```
>>> with mindio_acp.open_file('/mnt/dpc01/checkpoint/rank-0.pt') as f:
...     read_data = f.read()
```

## 返回值<a name="zh-cn_topic_0000002112429502_section17538071458"></a>

\_ReadableFileWrapper实例。

>**说明：**
>接口详情请参见[MindSpore文档](https://www.mindspore.cn/docs/zh-CN/master/api_python/mindspore/mindspore.load_checkpoint.html#mindspore.load_checkpoint)。
