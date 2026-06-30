# (Optional) Customizing a Configuration File<a name="ZH-CN_TOPIC_0000002447217833"></a>

You can customize a configuration file to clean key ModelArts logs or not, configure the size of console logs to be read, and configure file parsing. The user-defined configurations are saved in the `${HOME}/.ascend_faultdiag/custom-fd-config.json` file. During fault diagnosis, MindCluster Ascend FaultDiag automatically loads custom configurations in the corresponding path and cleans and diagnoses faults based on the configurations.

>[!NOTE]
>To customize the path for storing customized files, see [Customizing the MindCluster Ascend FaultDiag Home Directory](../common_operations.md#Customizing-the-MindCluster-Ascend-FaultDiag-Home-Directory).

**Procedure<a name="section5727144718537"></a>**

1. <a name="li180165713467"></a>Add or modify custom configurations in a JSON file.

    ```shell
    ascend-fd config --update custom-config.json
    ```

    If the following information is displayed, the operation is successful:

    ```ColdFusion
    The custom config file was updated successfully.
    ```

    The following is an example of the JSON file, which is for reference only. You need to modify the configurations as required. For details about parameters in the file, see [Table 1](#table122501013255309).

    ```json
    {
        "enable_model_asrt": false,   # Whether to enable cleaning of critical ModelArts logs. Disabled by default.
        "train_log_size": 1048576,  # Configures the console log size to be read. Default is 1MB = 1024*1024B = 1048576B.
        "custom_parse_file": [       # Configures custom files to be parsed. Can be set to []. Supports up to 10 files.
            {
                "file_path_glob": "test_custom/*.log",     # --custom_log xx, specifies a top-level directory. Matches files under the corresponding path using Unix-style wildcard patterns.
                "log_time_format": "%Y-%m-%d-%H:%M:%S.%f",  # Time format of the log file, which should be a standard format string for date-time parsing or formatting.
                "source_file": ["CustomLog"]              # Log file type. A maximum of 10 types can be configured.
            }
        ],
        "timezone_config" : {
            "lcne": true   # Whether to enable timezone conversion for LCNE logs. Disabled by default.
        }
    }
    ```

    >[!NOTE]
    >The more custom files configured for parsing, the lower the overall cleaning performance may become.

    **Table 1** Parameter description

    <a name="table122501013255309"></a>

    |Parameter|Value Type|Parameter Description|Required|Value Description|
    |--|--|--|--|--|
    |enable_model_asrt|Bool|Whether to clean key ModelArts logs|Optional|The default value is `false`. <ul><li>`true`</li><li>`false`</li></ul>|
    |train_log_size|Int|Size of the console log to be read|Optional|The value is a positive integer. The default value is `1048576` (1 MB = 1024 × 1024 bytes = 1048576 bytes).|
    |custom_parse_file|List|Custom parser file|Optional|A maximum of 10 files can be configured.|
    |file_path_glob|String|Custom parser file (Unix wildcard pattern)|Required when `custom_parse_file` exists and is not `[]`.|Letters, digits, English punctuation marks, spaces, and backslashes (\) and asterisks (*) are supported,for example, `test_custom/*.log`.|
    |log_time_format|String|Time format string for printing logs in the custom parser file|Optional|The value can contain 1 to 50 characters, including letters (YmdHMSf) and special characters (percent sign, hyphen, space, colon, comma, and period), for example, `%Y-%m-%d %H:%M:%S.%f`. <ul><li>%Y: 4-digit year (for example, 2023 or 2024) </li><li>%m: 2-digit month (01–12, for example, 03 indicates March) </li><li>%d: 2-digit date (01–31, for example, 05 indicates the fifth day of a month) </li><li>%H: number of hours in the 24-hour format (00–23). </li><li>%M: 2-digit minute (00–59). </li><li>%S: 2-digit second (00–59). </li><li>%f: microseconds.</li></ul>|
    |source_file|List|Log file type|Required when `custom_parse_file` exists and is not `[]`.|A maximum of 10 character strings can be configured. Each string contains 1 to 50 characters, including letters, digits, English symbols, and spaces.|
    |timezone_config|Dictionary|Log timezone conversion|Optional|-|
    |lcne|Bool|Specifies whether to support LCNE log timezone conversion.|Optional|The default value is `false`. <ul><li>`true`</li><li>`false`</li></ul>|
    |mindie|Bool|Specifies whether to support MindIE log timezone conversion. The function is not supported currently.|Optional|The default value is `false`. <ul><li>`true`</li><li>`false`</li></ul>|

2. Views the custom configurations.

    ```shell
    ascend-fd config --show
    ```

3. (Optional) Verify the `custom-fd-config.json` file. If you directly modify the custom fault entity information in the `custom-fd-config.json` file, run the following command to verify the integrity and availability of the modified file:

    ```shell
    ascend-fd config --check
    ```

    If the following information is displayed, the file verification is successful:

    ```ColdFusion
    The custom config file was updated successfully.
    ```

    >[!NOTE]
    >You are not advised to directly modify the `custom-fd-config.json` file. Otherwise, MindCluster Ascend FaultDiag may become abnormal.

4. (Optional) If a custom parser file (for example, the JSON file example in [Step 1](#li180165713467)) is added to the configuration file, run the following command to clean the custom parser file. The files matched by the wildcard pattern (`worker-0/test_custom/*.log`) will be cleaned.

    ```shell
    ascend-fd parse --custom_log worker-0/ -o *Cleaning output directory*
    ```

    >[!NOTE]
    >When cleaning a custom parser file, only the `--custom_log` command is supported. The `-i` command is not supported.
