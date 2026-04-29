# （可选）屏蔽故障日志<a name="ZH-CN_TOPIC_0000001846651017"></a>

通过新增含有故障关键词的屏蔽规则，实现在日志清洗时不将含有故障关键词的信息记录到日志清洗后的文件中。系统默认将屏蔽故障关键词文件保存在“$\{HOME\}/.ascend\_faultdiag/custom-blacklist.json”文件中。

>[!NOTE]
>
>- 当前仅支持对CANN应用类日志的ERROR日志进行屏蔽操作。
>- 若用户需要自定义屏蔽故障关键词文件的保存路径，可以参考[自定义MindCluster Ascend FaultDiag家目录](../common_operations.md#自定义mindcluster-ascend-faultdiag家目录)章节进行操作。

**操作步骤<a name="section71028910168"></a>**

1. 新增屏蔽规则。例如屏蔽带有ERROR和FUSION关键词的日志信息。

    ```shell
    ascend-fd blacklist --add ERROR FUSION
    ```

    >[!NOTE] 
    >- 关键词支持的最大长度为200个字符，支持大小写字母、数字和其他字符（如-./等）。
    >- 一条屏蔽规则最多支持10个关键词，支持通过空格分隔多个关键词，如ERROR FUSION、ERROR "FUSION"或"ERROR" "FUSION"。
    >- 如果关键词包含“\\”，为避免“\\”被忽略，整个关键词须打引号，如“ERR\\OR”。
    >- 当前最多保存50条屏蔽规则。若超出50条，系统将丢弃最前面的屏蔽规则，保留新增的屏蔽规则。

2. 通过JSON文件，导入屏蔽规则。

    ```shell
    ascend-fd blacklist --file test.json 
    ```

    >[!NOTE] 
    >通过导入JSON文件的方式新增屏蔽规则，导入的JSON文件会覆盖系统原有的JSON文件中的内容。

    JSON文件格式如下所示：

    ```json
    {
        "blacklist":[["ERROR2","ERROR3","ERROR3"],
            ["ERR1","ERR2","ERR3","ERR4"]
        ]
    }
    ```

3. 查看当前已有的屏蔽规则。

    ```shell
    ascend-fd blacklist --show
    ```

    回显示例如下：

    ```ColdFusion
    [BLACKLIST]
    0. 'ERROR', 'FUSION'
    1. 'ERROR2', 'ERROR2', 'ERROR3'
    2. 'ERR1', 'ERR2', 'ERR3', 'ERR4'
    ```

4. （可选）删除屏蔽规则，如删除第0条规则。支持同时删除多条规则，使用空格分隔。

    ```shell
    ascend-fd blacklist --delete 0
    ```
    