# Single-Server Fault Diagnosis Interface<a name="ZH-CN_TOPIC_0000002107732029"></a>

## Prototype<a name="section416811814162"></a>

- Cleans all logs on a single server, processes the log cleaning results to diagnose fault events, and outputs an analysis report.

    ```shell
    ascend-fd single-diag -i collection_directory -o single_server_fault_diagnosis_output_directory
    ```

- Performs single-server diagnosis by categorizing the input log directory.

    ```shell
    ascend-fd single-diag --host_log host_os_log_directory --device_log device-side_log_directory --train_log user_training_and_inference_log_directory --process_log cann_app_log_directory --env_check npu_port_status_resource_info_directory --dl_log mindcluster_log_directory --mindie_log mindie_log_directory --amct_log amct_log_directory -o cleaning_result_output_directory
    ```

>[!NOTE]
>
- When both `-i` and detailed log collection directory parameters are used together, the input values of the detailed log collection directory parameters will be read first, and then the remaining log collection directories will be read based on the `-i` parameter.
- At least one of the following parameters must be specified: `--input_path`, `--host_log`, `--device_log`, `--train_log`, `--process_log`, `--env_check`, `--dl_log`, `--mindie_log`, `--amct_log`, `--custom_log`, and `--bus_log`. Otherwise, the cleaning command will fail.
- The disk space of the output directory specified by the cleaning command must be greater than 5 GB. Insufficient space may cause partial loss of cleaning results, leading to abnormal or inaccurate diagnostic results.

## Function<a name="section67721623124010"></a>

Starts a single-server diagnosis task to diagnoses raw logs such as single-server run logs and NPU environment check files after a training or inference failure.

## Parameters<a name="section7746133874017"></a>

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
|`--output_path`|`-`o|Yes|String|Cleaned data output path. It only supports digits, letters, and spaces and characters `~`, `-`, `+`, `_`, `.`, `/`.|
|`--help`|`-h`|No|-|Queries the meanings of level-2 commands and parameters and usage instructions.|

## Returns<a name="section115671821144111"></a>

Example: Return single-server fault diagnosis task execution status.

```ColdFusion
The single-diag job starts. Please wait. Job id: [****], run log file is [****].
Diagnostic content
The single-diag job is complete.
```
