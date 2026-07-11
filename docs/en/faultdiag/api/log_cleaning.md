# Log Cleaning Interface<a name="ZH-CN_TOPIC_0000001541788954"></a>

## Prototype<a name="zh-cn_topic_0000001461778658_section876116162918"></a>

- Integrate all logs for cleaning.

    ```shell
    ascend-fd parse -i ${Collection_directory} -o ${Cleaning_result_output_directory}
    ```

- Classify input log directories for cleaning.

    ```shell
    ascend-fd parse --host_log ${host_os_log_directory} --device_log ${device-side_log_directory} --train_log ${user_training_and_inference_log_directory} --process_log ${cann_app_log_directory} --env_check ${npu_port_status_resource_info_directory} --dl_log ${mindcluster_log_directory} --mindie_log ${mindie_log_directory} --amct_log ${amct_log_directory} --bus_log ${lcne_(ascend_950)_log_directory} --custom_log ${custom_parser_directory} -o ${Cleaning_result_output_directory}
    ```

- (Optional) If there are BMC logs, execute the following command.

    ```shell
    ascend-fd parse --bmc_log ${bmc_log_directory} -o ${directory_for_saving_cleaning_results}
    ```

    For example:

    ```shell
    ascend-fd parse --bmc_log  "bmc/worker-00" -o "auto_diag_combine/bmc/worker-00"
    ```

- (Optional) If there are LCNE logs, execute the following command.

    ```shell
    ascend-fd parse --lcne_log ${lcne_log_directory} -o ${directory_for_saving_cleaning_results}
    ```

    For example:

    ```shell
    ascend-fd parse --lcne_log  "lcne/worker-111" -o "auto_diag_combine/lcne/worker-111"
    ```

**NOTE**

- When both `-i` and detailed log collection directory parameters are used together, the input values of the detailed log collection directory parameters will be read first, and then the remaining log collection directories will be read based on the `-i` parameter.
- At least one of the following parameters must be specified: `--input_path`, `--host_log`, `--device_log`, `--train_log`, `--process_log`, `--env_check`, `--dl_log`, `--mindie_log`, `--amct_log`, `--custom_log`, and `--bus_log`. Otherwise, the cleaning command will fail.
- The disk space of the output directory specified by the cleaning command must be greater than 5 GB. Insufficient space may cause partial loss of cleaning results, leading to abnormal or inaccurate diagnostic results.

## Function<a name="zh-cn_topic_0000001461778658_section10145143713297"></a>

Starts a log cleaning task to clean raw logs such as run logs and NPU environment check files after a training or inference failure.

## Parameters<a name="zh-cn_topic_0000001461778658_section1094205815292"></a>

**Table 1** Parameters

|Parameter|Abbreviation|Mandatory|Value Type|Description|
|--|--|--|--|--|
|`--host_log`|None|No|String|Host-side OS log collection directory. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--device_log`|None|No|String|Device-side log collection directory. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--train_log`|None|No|String|User training and inference log collection directory.<ul><li>`--train_log` supports multiple path inputs. A path can be a file name for a single collection log or a collection directory for dump logs. However, a maximum of 20 paths are read, and the excessive paths are discarded.</li><li>When a file name is specified using `--train_log`, user training and inference logs are no longer subject to naming constraints. When a path is specified using `--train_log`, files ending with `.txt` or `.log` in the path are considered training and inference logs.</li></ul>|
|`--process_log`|None|No|String|CANN App log collection directory. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--env_check`|None|No|String|NPU network port, status information, and resource information collection directory. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--dl_log`|None|No|String|Log directory for MindCluster components, including Ascend Device Plugin, NodeD, Ascend Docker Runtime, NPU Exporter, and Volcano. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--mindie_log`|None|No|String|Log directory for MindIE components, including MindIE Server, MindIE LLM, MindIE SD, MindIE RT, MindIE Torch, MindIE MS, MindIE Benchmark, and MindIE Client. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--amct_log`|None|No|String|AMCT log directory. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--bmc_log`|None|No|String|BMC log directory. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--lcne_log`|None|No|String|LCNE log directory. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--bus_log`|None|No|String| LCNE (Ascend 950) log directory. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--custom_log`|None|No|String|Custom parser directory. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--input_path`|`-i`|No|String|Preprocessing data input path. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--output_path`|`-o`|Yes|String|Cleaned data output path. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--performance`|`-p`|No|Bool|When this parameter is specified, all cleaning modules are executed. If not specified, only the root cause node and fault event modules perform log cleaning.|
|`--help`|`-h`|No|-|Queries the meanings of level-2 commands and parameters and usage instructions.|

## Returns<a name="zh-cn_topic_0000001461778658_section2134184616351"></a>

Example: Return log cleaning task execution status.

```ColdFusion
The parse job starts. Please wait. Job id: [****], run log file is [****].
These job ['Module 1', 'Module 2'...] succeeded.
The parse job is complete.
```
