# （可选）自定义配置文件<a name="ZH-CN_TOPIC_0000002447217833"></a>

支持用户自定义配置文件，可以配置是否支持清洗ModelArts关键日志、配置读取控制台日志大小、配置解析自定义的文件。用户自定义的配置信息保存在“$\{HOME\}/.ascend\_faultdiag/custom-fd-config.json”文件中。在执行故障诊断功能时，MindCluster Ascend FaultDiag会自动在相应路径下加载用户自定义配置信息，根据配置信息进行清洗和诊断。

>[!NOTE] 
>若用户需要自定义故障文件的保存路径，可以参考[自定义MindCluster Ascend FaultDiag家目录](../common_operations.md#自定义mindcluster-ascend-faultdiag家目录)章节进行操作。

**操作步骤<a name="section5727144718537"></a>**

1. <a name="li180165713467"></a>通过JSON文件，新增或修改自定义配置信息。

    ```shell
    ascend-fd config --update custom-config.json
    ```

    回显示例如下，表示操作成功。

    ```ColdFusion
    The custom config file was updated successfully.
    ```

    JSON文件示例如下，该示例不可直接使用，用户需根据实际情况修改自定义配置信息。文件中的参数说明请参见[表1](#table122501013255309)。

    ```json
    {
        "enable_model_asrt": false,   # 是否支持清洗ModelArts关键日志。默认关闭
        "train_log_size": 1048576,    # 配置读取控制台日志大小。默认1MB=1024*1024B=1048576B
        "custom_parse_file": [        # 配置解析自定义的文件。可配置为[]，最大支持配置10个
            {
                "file_path_glob": "test_custom/*.log",      # --custom_log xx，指定大目录。对应路径下，按照 Unix 风格的通配符模式匹配文件。
                "log_time_format": "%Y-%m-%d-%H:%M:%S.%f",  # 日志文件的时间格式，日期时间解析或格式化的标准格式字符串。     
                "source_file": ["CustomLog"]                # 日志文件类型，最大支持配置10个
            }
        ],
        "timezone_config" : {
            "lcne" : true    # 是否支持LCNE日志时区转换。默认关闭
        }
    }
    ```

    >[!NOTE] 
    >配置解析自定义的文件越多，整体的清洗性能可能会下降。

    **表 1**  参数说明

    <a name="table122501013255309"></a>

    |参数名称|取值类型|参数说明|是否必选|取值说明|
    |--|--|--|--|--|
    |enable_model_asrt|Bool|是否支持清洗ModelArts关键日志。|可选|默认为false。<ul><li>true</li><li>false</li></ul>|
    |train_log_size|Int|配置读取控制台日志大小。|可选|正整数，默认为1048576（1MB=1024\*1024B=1048576B）。|
    |custom_parse_file|List|配置解析自定义的文件。|可选|列表格式，最大支持配置10个文件。|
    |file_path_glob|String|自定义的解析文件（Unix风格的通配符模式）。|custom_parse_file存在且不为[]时必选|支持英文字母、数字、英文符号、空格与“\*”（例如配置："test_custom/\*.log"）。|
    |log_time_format|String|自定义解析文件中日志打印的时间格式字符串。|可选|取值长度为1~50个字符，支持字符："YmdHMSf%- :,."（例如配置为"%Y-%m-%d %H:%M:%S.%f"）。<ul><li>%Y：4位年份（例如：2023、2024）。</li><li>%m：2位月份（01-12，例如：03 表示3月）。</li><li>%d：2位日期（01-31，例如：05表示5号）。</li><li>%H：24小时制的小时数（00-23）。</li><li>%M：2位分钟数（00-59）。</li><li>%S：2位秒数（00-59）。</li><li>%f：微秒数。</li></ul>|
    |source_file|List|日志文件类型。|custom_parse_file存在且不为[]时必选|列表格式，最大支持配置10个字符串。每个字符串取值长度为1~50个字符，支持英文字母、数字、英文符号与空格。|
    |timezone_config|Dictionary|日志时区转换。|可选|-|
    |lcne|Bool|是否支持LCNE日志时区转换。|可选|默认为false。<ul><li>true</li><li>false</li></ul>|
    |mindie|Bool|是否支持MindIE日志时区转换。该功能暂不支持。|可选|默认为false。<ul><li>true</li><li>false</li></ul>|

2. 查看用户自定义的配置信息。

    ```shell
    ascend-fd config --show
    ```

3. （可选）校验custom-fd-config.json文件。若用户直接修改custom-fd-config.json文件的相关自定义故障实体信息，可以执行以下命令，校验修改后文件的完整性和可用性。

    ```shell
    ascend-fd config --check
    ```

    回显示例如下，表示文件校验通过。

    ```ColdFusion
    The custom config file was updated successfully.
    ```

    >[!NOTE] 
    >不建议用户直接更改custom-fd-config.json文件信息，可能造成MindCluster Ascend FaultDiag组件功能异常。

4. （可选）若配置文件中添加了自定义的解析文件（如[步骤1](#li180165713467)中JSON文件示例），可执行以下命令对自定义的解析文件进行清洗，将会清洗通配符模式（worker-0/test\_custom/\*.log）匹配到的文件。

    ```shell
    ascend-fd parse --custom_log worker-0/ -o 清洗输出目录
    ```

    >[!NOTE] 
    >清洗自定义解析文件时，只支持--custom\_log命令，不支持-i命令。
