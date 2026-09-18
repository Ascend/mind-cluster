# diag Command (Fault Diagnosis)

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:30:32.971Z pushedAt=2026-08-24T02:51:43.311Z -->

## Function Description

Used to analyze the root cause of node faults. To analyze the root cause of faults across multiple nodes, you need to aggregate the parsing results of each node into the same directory before diagnosis. For the directory format, see [Log Collection Archive Directory Structure](../05_usage/02_log_collection.md#log-collection-archive-directory-structure).

## Command Format

```shell
ascend-fd diag [-h] -i INPUT_PATH -o OUTPUT_PATH [-p] [-s {host,super_pod}]
```

## Parameter Description

| Name              | Type   | Mandatory (Yes/No) | Description                                                                         |
|-------------------|--------|-----------|-------------------------------------------------------------------------------------|
| `-h`, `--help`        | -      | No        | Displays help information.                                                          |
| `-p`, `--performance` | -      | No        | Executes all diagnostic modules (including device resource analysis and network congestion analysis). |
| `-i`, `--input_path`  | String | Yes       | Input path of the parsed data.                                                     |
| `-o`, `--output_path` | String | Yes       | Output path of the diagnostic results.                                              |
| `-s`, `--scene`       | String | No        | Diagnostic scenario. Options: `host` and `super_pod`. Default: `host`. |

Diagnostic input directory structure:

```text
Diagnostic input directory
├── Node 1 parsing results/
│   ├── ascend-rc-parser.json
│   ├── ascend-kg-analyzer.json
│   ├── ascend-kg-parser.json
│   └── ...
├── Node 2 parsing results/
│   └── ...
└── Node N parsing results/
    └── ...
```

> [!NOTE]
>
> - For multi-node parsed data, manually merge the parsing results from each node into a single directory.
> - For the single-node parsed data, refer to the parsing result of node 1. For field meanings, refer to [Parsing Result Description](02_command_parse.md#parsing-result-description).

## Usage Examples

### Basic Diagnosis

```shell
ascend-fd diag -i /tmp/parse_out -o /tmp/diag_out
```

1. Diagnosing abnormal termination of a training job. Example output:

    <!-- markdownlint-disable-next-line MD033 -->
    <pre>
    +------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                                       Ascend Fault-Diag Report                                                                                       |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+                                                                                                                                                                       |   Version Information   |   Type     | Version                                                                                                                                                                     |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.1.0                                                                                                                                                                    |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    | Root Cause Node Analysis |    Type    | Description                                                                                                                                                                     |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Description    | The peer node in the fault node chain could not confirm the specific Worker ID and Device ID. Please check the corresponding Device via IP.                                                                                       |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |  Root Cause Node  | ['worker-0 device-2']                                                                                                                                                    |
    |              |  First Error Node  | worker-1 device-2: 2023-09-01-06:35:52.960343                                                                                                                            |
    |              |  Symptom  | Some nodes experienced excessive RoCE retransmissions (ERROR CQE). These nodes are suspected root cause nodes. Please investigate.                                                                                                    |
    |              | Root Cause Node Chain | ['worker-0 device-2 -> 192.168.102.220']                                                                                                                                 |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    | Fault Event Analysis |    Type    | Description                                                                                                                                                                     |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Description    | 1. Multiple faults have been diagnosed and are sorted by occurrence time. Please prioritize the earlier faults for investigation.                                                                                                             |
    |              |            | 2. Note: Some faults involve too many devices; only the first 16 are displayed here. All faulty devices can be queried in diag_report.json.                                                                                               |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |   Status Code    | xxx                                                                                                                                                                      |
    |              |  Fault Category  | Category: Network Component: Network Module: Network                                                                                                                                   |
    |              |  Fault Device  | ['worker-0 device-2']                                                                                                                                                    |
    |              |  Fault Name  | Link Down: NPU-side intermittent disconnection                                                                                                                                                 |
    |              |  Fault Desc  | An NPU network port on this server experienced a Link Down intermittent disconnection error, and the intermittent disconnection duration exceeded 30 seconds.                                                                                                            |
    |              |  Suggestion  | 1. Please contact the physical network operations team to collect switch logs and check for hardware issues (whether optical modules are present, whether switch links disconnected intermittently).                                                              |
    |              |  Key Log  | /usr/local/Ascend/driver/tools/hccn_tool -i 2 -link_stat -g                                                                                                              |
    |              |            | [devid 2]current time        : Fri Sep  1 06:37:26 2023                                                                                                                  |
    |              |            | [devid 2]link up count       : 2                                                                                                                                         |
    |              |            | [devid 2]link change records :                                                                                                                                           |
    |              |            | [devid 2]    Fri Sep  1 06:34:43 2023    LINK DOWN                                                                                                                       |
    |              |            | [devid 2]    Thu Aug 31 07:30:46 2023    LINK UP                                                                                                                         |
    |              |            | [devid 2]    Thu Aug 31 07:30:44 2023    LINK DOWN                                                                                                                       |
    |              |            | [devid 2]    Thu Aug 31 07:30:43 2023    LINK UP                                                                                                                         |
    |              | Key Propagation Chain | ['worker-0']                                                                                                                                                             |
    |              |            | Fault code 1 (Link Down: NPU-side intermittent disconnection)-> Fault code 2 (RDMA communication excessive retransmission) -> Fault code 3 (notify wait timeout)                                                                           |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    </pre>

    > [!NOTE]
    >
    > - When the peer device is not found, the root cause node chain is displayed differently for different product forms.
    >   - <term>Ascend 950 products</term>: displays the peer device EID.
    >   - Other products: displays the peer device IP.
    > - For field descriptions of the displayed results, see [Diagnostic Result Description](#diagnostic-result-description).

2. Diagnose multi-instance inference job exceptions. Example output:

    <!-- markdownlint-disable-next-line MD033 -->
    <pre>
    ============================
    Instance name:****
    Node name: [****,****]
    +--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                                  Ascend Fault-Diag Report                                                                                  |
    +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |   Version Information   |    Type    | Version                                                                                                                                                           |
    +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.1.0                                                                                                                                                        |
    +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
    | Root Cause Node Analysis |    Type    | Description                                                                                                                                                           |
    +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Description    | No root cause node was diagnosed. The fault event analysis will attempt to detect all devices.                                                                                                               |
    +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |  Root Cause Node  | ['Unknown Device']                                                                                                                                             |
    |              |  Symptom  | No error log information found in Plogs from all valid nodes. Root cause node cannot be located. Please also confirm whether this is a normal job.                                                                           |
    +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
    | Fault Event Analysis |   Type    | Description                                                                                                                                                           |
    +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
    | Suspected Root Cause Fault |   Status Code   | xxx                                                                                                                                                            |
    |              |  Fault Category  | Category: Software Component: MindIE Module: LLM                                                                                                                             |
    |              |  Fault Device  | ['worker-0 device-2']                                                                                                                                          |
    |              |  Fault Name  | BackendConfig configuration parameter validation failure                                                                                                                                  |
    |              |  Fault Desc  | Invalid configuration parameter.                                                                                                                                               |
    |              |  Suggestion  | 1. Please contact Huawei engineers.                                                                                                                                      |
    |              |  Key Log  | [2025-06-17 16:59:10.282+08:00] [97] [147] [server] [WARN] [llm_infer_engine.cpp:117] : MIE05E040000[llm_backend] get model instance processing request failed |
    +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+

    ============================
    Instance name: ****
    Node name: [****, ****]
    +-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                                   Ascend Fault-Diag Report                                                                                    |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |   Version Information   |    Type    | Version                                                                                                                                                              |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.1.0                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    | Root Cause Analysis |    Type    | Description                                                                                                                                                              |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |  Root Cause Node  | ['worker-1 device-11']                                                                                                                                            |
    |              |  Symptom  |  No timeout error logs were recorded in Plogs from all nodes. Nodes with error logs are suspected root cause nodes. Please investigate.                                                                                |
    |              |  First Error Node  | xx.xx.xx.xx device-11: 2025-06-17-17:03:02.614708                                                                                                                 |
    |              |  Last Error Node  | xx.xx.xx.xx device-11: 2025-06-17-17:03:02.614708                                                                                                                 |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    | Fault Event Analysis |    Description    | Description                                                                                                                                                              |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    | Suspected Root Cause Fault |   Status Code   | xxx                                                                                                                                                            |
    |              |  Fault Category  | Category: Software Component: MindIE Module: LLM                                                                                                                             |
    |              |  Fault Device  | ['worker-1 device-11']                                                                                                                                            |
    |              |  Fault Name  | BackendConfig configuration parameter validation failure                                                                                                                                  |
    |              |  Fault Desc  | Invalid configuration parameter.                                                                                                                                               |
    |              |  Suggestion  | 1. Please contact Huawei engineers.                                                                                                                                      |
    |              |  Key Log  | [2025-06-17 17:01:44.605+08:00] [2015] [5415] [server] [WARN] [llm_infer_engine.cpp:117] : MIE05E040000[llm_backend] get model instance processing request failed |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    </pre>

    For descriptions of the fields in the returned result, see [Diagnostic Result Description](#diagnostic-result-description).

### Performance Degradation Diagnosis

```shell
ascend-fd diag -i /tmp/parse_out -o /tmp/diag_out -p
```

Diagnose performance degradation issues during training. Example output:

<!-- markdownlint-disable-next-line MD033 -->
<pre>
+--------------------------------------------------------------------------------------------------------------------------------------------------+
|                                                             Ascend Fault-Diag Report                                                             |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
|   Version Information   |    Type    | Version                                                                                                                   |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
|              |Fault-Diag| 26.1.0                                                                                                                  |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
| Root Cause Analysis |    Type    | Description                                                                                                                   |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
|              |   Description   | No root cause node was diagnosed. The fault event analysis will attempt to detect all devices.                                                                       |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
|              | Root Cause Node | ['Unknown Device']                                                                                                     |
|              | Symptom | No error logs found in Plogs from all valid nodes, and no heartbeat information found in Plogs from all nodes. | | | | Root cause node cannot be located. Please also confirm whether this is a normal training job. |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
| Fault Event Analysis |   Type   | Description                                                                                                                   |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
|              |  Status Code  | NORMAL_OR_UNSUPPORTED                                                                                                  |
|              | Result Desc | The fault event analysis module returned no results. This may indicate a normal training job with no faults. If the training job terminated abnormally and the issue persists, please contact Huawei engineers. |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
| Device Resource Analysis |   Type   | Description                                                                                                                   |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
|              |   Description   | Some analysis sub-items under this module failed to execute, which may affect the accuracy of the diagnostic results. Please check the module logs for details.                         |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
|              |  Status Code  | xxx                                                                                                                    |
|              | Fault Device | worker-0                                                                                                               |
|              | Fault Process | [2381084, 2381097]                                                                                                     |
|              | Fault Interval | [('2023-08-11 02:18:00', '2023-08-11 02:21:00'), 'fault probability:  0.663']                                                   |
|              | Fault Name | CPU preemption (partial process preemption)                                                                                                |
|              | Fault Desc | Abnormal device resource condition detected: CPU resource preemption occurred for some training processes.                                                                        |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
| Network Congestion Analysis |   Type   | Description                                                                                                                   |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
|              |  Status Code  | xxx                                                                                                                    |
|              | -------- |                                                                                                                        |
|              | Fault Device | worker-0                                                                                                               |
|              | Fault Node | ['device-0', 'device-1', 'device-2', 'device-3', 'device-4', 'device-5', 'device-6', 'device-7']                       |
|              | -------- |                                                                                                                        |
|              | Fault Device | worker-1                                                                                                               |
|              | Fault Node | ['device-0', 'device-1', 'device-2', 'device-3', 'device-4', 'device-5', 'device-6', 'device-7']                       |
|              | -------- |                                                                                                                        |
|              | Fault Name | Link congestion anomaly                                                                                                           |
|              | Fault Desc | Congestion and contention detected on some communication links.                                                                                             |
|              | Suggestion | It is recommended to check the switch routing policy.                                                                                               |
+--------------+----------+------------------------------------------------------------------------------------------------------------------------+
</pre>

### SuperPoD Diagnosis

Take automatic association of topology information as an example:

```shell
ascend-fd diag -i /tmp/super_pod_parse_out -o /tmp/diag_out -s super_pod
```

Output example:

<!-- markdownlint-disable -->
<pre>
+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
|                                                                                   Ascend Fault-Diag Report                                                                  |
+--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
|   Version Information   |    Type    | Version                                                                                                                                            |
+--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
|              | Fault-Diag | 26.1.0                                                                                                                                         |
+--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
| Root Cause Node Analysis |    Type    | Description                                                                                                                                            |
+--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
|              |  Root Node Cause  | ['xxxxxxxxxxx']                                                                                             |
|              |  Symptom  | No timeout information found in all Plogs. Some nodes may have experienced abnormal process exit or hang.                                                                                        |
+--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
| Fault Event Analysis |    Type    | Description                                                                                                                                            |
+--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
|              |    Description    | 1. Some analysis sub-items under this module failed to execute, which may affect the accuracy of the diagnostic results. Failure details can be found in diag_report.json.                                       |
|              |            | 2. Multiple faults have been diagnosed and are sorted by priority. Please focus on the earlier faults for investigation.                                                                                      |
+--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
| Suspected Root Cause Fault |   Status Code   | ******                                                                                                                                          |
|              |  Fault Category  | Category: Network Component: Switch Module: Chip                                                                                                              |
|              |  Fault Device  | ['LCNE:xxx.xx.xx.xx5']                                                                                                                          |
|              |  Fault Name  | Forwarding engine module function failure                                                                                                                           |
|              |  Fault Desc  | LANSWITCH chip is unstable.                                                                                                                           |
|              |  Suggestion  | 1. Please contact Huawei engineers.                                                                                                                       |
|              |  Key Log  | ******                                                                                                                                          |
|              |            | ******                                                                                                                                          |
+--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
|              |   Status Code   | ******                                                                                                                                          |
|              |  Fault Category  | Category: Network Component: Switch Module: Chip                                                                                                              |
|              |  Fault Device  | ['LCNE:xxx.xx.xx.xx5']                                                                                                                          |
|              |  Fault Name  | Forwarding chip port downgraded to 1/2 lane                                                                                                                    |
|              |  Fault Desc  | Forwarding chip port downgraded to 1/2 lane L1&lt;--&gt;CPU.                                                                                                        |
|              |  Suggestion  | 1. Please contact Huawei engineers.                                                                                                                       |
|              |  Key Log  | ******                                                                                                                                          |
|              |            | ******                                                                                                                                          |
+--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
</pre>
<!-- markdownlint-enable -->

## Diagnostic Result Description

<!-- markdownlint-disable -->
| Level-1 Parameter    | Level-2 Parameter  | Description                                                                                                       |
|--------------|------------|------------------------------------------------------------------------------------------------------------|
| Root Cause Node Analysis | -          | Used to analyze the fault root cause node.                                                                                     |
| -            | Root Cause Node   | The device where the root cause resides. If it is `Unknown Device`, it is because log collection is incomplete or there is no valid device information in the logs.                                                                                     |
| -            | Symptom    | Symptom of the root cause node analysis.                                                                                   |
| -            | First Error Node   | The device where the earliest error occurred in the task.                                                                               |
| -            | Last Error Node   | The device where the latest error occurred in the task.                                                                               |
| -            | Plog Log   | When the root cause node is `Unknown Device` and a first error node exists, displays the first 10 lines of Plog logs starting from the first error log line of the first error node.        |
| -            | Log Desc   | When the root cause node is `Unknown Device` and a first error node exists, displays the path description of the original Plog logs of the first error node.                     |
| -            | Root Cause Node Chain | Propagation relationship of the faulty node when retransmission times out.                                                                             |
| -            | Inter-device Waiting Chain | Propagation relationship of the faulty node when a Socket/Notify timeout occurs.                                                                  |
| Fault event analysis | -          | Used to analyze the root cause error of the device where the fault root cause node resides.                                                                   |
| -            | Status Code     | <ul><li>When a fault is diagnosed, displays the specific fault code.</li><li>When no fault is diagnosed, displays `NORMAL_OR_UNSUPPORTED`.</li></ul> |
| -            | Fault Name   | The specific fault name.                                                                                           |
| -            | Fault Category   | The category of the fault and the component and module where it resides.                                                                             |
| -            | Fault Device   | The device where the fault occurred. If it is Unknown Device, the device is not located, for example, there is no valid device information in the logs.                                                                                             |
| -            | Fault Desc   | Detailed description or explanation of the fault.                                                                               |
| -            | Suggestion   | Handling suggestions for the fault.                                                                                     |
| -            | Key Log   | The fault log corresponding to the fault.                                                                                     |
| -            | Key Propagation Chain | Displays the longest link in the causal relationship of the fault.                                                                       |
| Device resource analysis | -          | Used to analyze the resource status of the device.                                                                                   |
| -            | Status Code     | <ul><li>When a fault is diagnosed, displays the specific fault code.</li><li>When no fault is diagnosed, displays `NODE_DIAGNOSIS_NORMAL`.</li></ul> |
| -            | Fault Device   | The name of the node where the fault occurred.                                                                                   |
| -            | Fault Name   | The specific fault name.                                                                                           |
| -            | Suggestion   | Handling suggestions for the fault.                                                                                     |
| Network Congestion Analysis | -          | Used to analyze the network status between nodes.                                                                                 |
| -            | Status Code     | <ul><li>When a fault is diagnosed, displays the specific fault code.</li><li>When no fault is diagnosed, displays `NET_DIAGNOSIS_NORMAL`.</li></ul>  |
| -            | Fault Device   | The name of the node where the fault occurred.                                                                                   |
| -            | Fault Name   | The specific fault name.                                                                                           |
| -            | Suggestion   | Handling suggestions for the fault.                                                                                     |

> [!NOTE]
>
> - The output showing root cause node analysis and fault event analysis indicates that the current fault has caused the training job to exit abnormally.
> - Only when the output does not diagnose a root cause node and the fault event analysis has no result will device resource analysis and network congestion analysis be performed, indicating that the current fault is a performance degradation issue and will not cause the training job to exit abnormally.
> - Some key parameters exist only in specific scenarios.

## Diagnostic Result Output

```text
Diagnosis output directory
├── fault_diag_result
    ├── diag_report.json    # Diagnostic result
    └── diag_report_{instance_name}.json    # Multi-instance inference diagnosis results, for example, diag_report_192.168.0.1-192.168.0.2.json
```

Example of `diag_report.json`:

```json
{
    "Version": "26.1.0",
    "Build_Time": "2026-05-19",
    "Task_id": "20260714035618106072_df7bf6b1-c77a-4991-b92c-a563e952077e",
    "Root_Cluster": {
        "analyze_success": true,
        "fault_description": {
            "code": 113,
            "string": "The error reported by the earliest error node is not a timeout error. All error nodes that did not report a timeout error are suspected root cause nodes. Please troubleshoot them."
        },
        "root_cause_device": [
            "worker-0 device-4",
            "worker-0 device-Unknown",
            "worker-0 device-0",
            "worker-0 device-3",
            "worker-0 device-2",
            "worker-0 device-7",
            "worker-0 device-5"
        ],
        "device_link": [],
        "remote_link": "",
        "first_error_device": "worker-0 device-4: 2026-05-29-15:29:54.823918",
        "last_error_device": "worker-0 device-6: 2026-05-29-15:35:11.210962",
        "note": "The root cause node analysis detected multiple suspected fault root cause nodes. These nodes will be prioritized for troubleshooting.",
        "fault_description_list": [],
        "mindie_error_device": [],
        "show_device_info": {
            "device_type": "first_root_device",
            "device": "worker-0 device-4",
            "plog_file_path": "/tmp/fault_diag_data/worker-0/plog-parser-505389-1.log",
            "error_log": "[ERROR] TBE(505389,python3):2026-05-29-15:29:54.823.918 [ascendc_runtime.cpp:228]  505389 AscendCheckSoCVersion:cur soc version ascend950pr_9599 not found.\n"
        },
        "detect_workers_devices": {
            "worker-0": [
                "4",
                "Unknown",
                "0",
                "3",
                "2",
                "7",
                "5"
            ]
        }
    },
    "Knowledge_Graph": {
        "analyze_success": true,
        "version_info": {},
        "note": "",
        "fault": [
            {
                "code": "AISW_CANN_ACL_02",
                "component": "CANN",
                "module": "ACL",
                "cause_zh": "ACL单算子接口执行失败",
                "description_zh": "a. 单算子编译失败。\nb. 单算子执行超时。\nc. device异常。",
                "suggestion_zh": [
                    "1. 记录错误日志、ERR MSG、API返回错误码；"
                ],
                "class": "Software",
                "event_attr": {
                    "worker-0 device-Unknown": [
                        {
                            "event_code": "AISW_CANN_ACL_02",
                            "key_info": "[ERROR] OP(3072863,python3):2026-05-29-15:35:10.674.067 [engine.cc:755]3074390 Execute op failed. op type = fasfasdfasdfsaf",
                            "is_custom_event": false,
                            "source_device": "Unknown",
                            "source_file": "plog-662661_20260529153510798.log",
                            "occur_time": "2026-05-29 15:35:10.674067",
                            "type": "CANN_Plog",
                            "occurrence": [
                                [
                                    "2026-05-29 15:35:10.674067",
                                    "[ERROR] OP(3072863,python3):2026-05-29-15:35:10.674.067 [engine.cc:755]3074390 Execute op failed. op type = fasfasdfasdfsaf"
                                ]
                            ],
                            "event_id": "key1"
                        }
                    ]
                },
                "fault_source": [
                    "worker-0 device-Unknown"
                ],
                "fault_chains": []
            }
        ]
    }
}
```

<!-- markdownlint-enable MD013 -->

### `diag_report.json` Field Description

- Top-level fields

<!-- markdownlint-disable -->
| Field             | Type   | Description                                       |
|-------------------|--------|---------------------------------------------------|
| `Version`         | String | Fault-Diag version number.                        |
| `Build_Time`      | String | Fault-Diag build time.                            |
| `Task_id`         | String | Diagnosis task ID.                                |
| `Root_Cluster`    | Object | Root cause node analysis result. Not output in single-node diagnosis. |
| `Knowledge_Graph` | Object | Fault event analysis result.                      |
<!-- markdownlint-enable -->

- `Root_Cluster` fields

<!-- markdownlint-disable -->
| Field                      | Type                       | Description                                                           |
|----------------------------|----------------------------|-----------------------------------------------------------------------|
| `analyze_success`          | Boolean                    | Whether root cause node analysis succeeds.<ul><li>`true`: success</li><li>`false`: failure</li></ul> |
| `fault_description`        | Object                     | Fault description.                                                    |
| `fault_description.code`   | Integer                    | Fault code.                                                           |
| `fault_description.string` | String                     | Fault code description.                                               |
| `root_cause_device`        | List[String]               | List of root cause devices.                                           |
| `device_link`              | List                       | Root cause node chain, the propagation relationship of faulty nodes when retransmission timeout is exceeded. |
| `remote_link`              | String                     | Inter-device waiting chain, the propagation relationship of nodes when a timeout fault occurs. |
| `first_error_device`       | String                     | The device where the earliest error occurred in the task and its error time. |
| `last_error_device`        | String                     | The device where the latest error occurred in the task and its error time. |
| `note`                     | String                     | Remarks, supplementary description of the root cause node analysis.  |
| `fault_description_list`   | List                       | List of fault descriptions.                                           |
| `mindie_error_device`      | List[String]               | List of devices with MindIE link establishment faults.                |
| `show_device_info`         | Object                     | Display information of the first root cause node or first error node. See the table below for fields. |
| `detect_workers_devices`   | Dict[String, List[String]] | List of device IDs detected under each worker.                        |
<!-- markdownlint-enable -->

- `show_device_info` field

<!-- markdownlint-disable -->
| Field            | Type   | Description                                                                                |
|------------------|--------|---------------------------------------------------------------------------------------------|
| `device_type`    | String | Device type. The value is `first_root_device` (first root device) or `first_error_device` (first error node). |
| `device`         | String | Device information.                                                                          |
| `plog_file_path` | String | Plog log file path.                                                                         |
| `error_log`      | String | Error log content.                                                                          |
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013-->
>[!NOTE]
>
>`show_device_info` is output only when the root cause node is not an `Unknown Device` or a first error node exists, and it contains fields only when the first root cause/error node has error logs.

- `Knowledge_Graph` field

<!-- markdownlint-disable -->
| Field             | Type         | Description                                          |
|-------------------|--------------|------------------------------------------------------|
| `analyze_success` | Boolean      | Whether fault event analysis succeeds.<ul><li>`true`: success</li><li>`false`: failure</li></ul> |
| `version_info`    | Object       | Version information.                                    |
| `note`            | String       | Remarks.                                        |
| `fault`           | List[Object] | Fault event list. See the following table for fields.                    |
<!-- markdownlint-enable -->

- `fault` field

<!-- markdownlint-disable -->
| Field           | Type         | Description                                                          |
|------------------|--------------|----------------------------------------------------------------------|
| `code`           | String       | Fault code.                                                          |
| `component`      | String       | Fault component.                                                    |
| `module`         | String       | Fault module.                                                       |
| `cause_zh`       | String       | Fault cause (Chinese).                                               |
| `description_zh` | String       | Fault description (Chinese).                                         |
| `suggestion_zh`  | List[String] | Fault suggestions (Chinese).                                         |
| `class`          | String       | Fault category.                                                      |
| `event_attr`     | Object       | Fault event attributes. The key is the device name, and the value is a list of event attributes. See the following table. |
| `fault_source`   | List[String] | List of fault source devices.                                        |
| `fault_chains`   | List         | Fault propagation chain.                                             |
<!-- markdownlint-enable -->

- `event_attr` field

`event_attr` is a list of event attribute objects, each containing the following:

<!-- markdownlint-disable -->
| Field             | Type               | Description                                                          |
|-------------------|--------------------|----------------------------------------------------------------------|
| `event_code`      | String             | Event fault code.                                                    |
| `key_info`        | String             | Key log information.                                                 |
| `is_custom_event` | Boolean            | Whether it is a custom event.                                        |
| `source_device`   | String             | Source device.                                                       |
| `source_file`     | String             | Source log file name.                                                |
| `occur_time`      | String             | Fault occurrence time.                                               |
| `type`            | String             | Log type.                                                            |
| `occurrence`      | List[List[String]] | List of fault occurrence records. Each record is [occurrence time, key log]. |
| `event_id`        | String             | Event ID.                                                            |
<!-- markdownlint-enable -->

## Precautions

- Before diagnosis, ensure that the system time of each node is synchronized.
- For ascend-fd runtime error codes, see [Component Error Codes](../07_references/04_appendix.md#component-error-codes).
