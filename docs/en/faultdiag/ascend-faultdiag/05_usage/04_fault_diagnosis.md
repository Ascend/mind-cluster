# Fault Diagnosis

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:27:36.691Z pushedAt=2026-08-24T02:51:43.250Z -->

After log cleaning is complete, use the diagnosis function to analyze the root cause of a fault.

## Procedure

1. Collect logs by following [Log Collection](./02_log_collection.md).

2. Clean logs by following [Log Parsing](./03_log_parsing.md).

3. Create an output directory for diagnostic results.

    ```shell
    mkdir <diagnosis_output_directory>
    ```

4. Run the diagnostic command.

    ```shell
    ascend-fd diag -i <diagnostic_input_directory> -o <diagnosis_output_directory>
    ```

    To also diagnose performance degradation issues, add the `-p` parameter:

    ```shell
    ascend-fd diag -i <diagnostic_input_directory> -o <diagnosis_output_directory> -p
    ```

## Diagnosis Report Interpretation

After diagnosis is complete, the diagnosis report is displayed in the terminal.

 <!-- markdownlint-disable-next-line MD033 -->
<pre>
The diag job starts. Please wait. Job id: [****], run log file is [****].
+------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
|                                                                                       Ascend Fault-Diag Report                                                                                       |
+--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
|   Version Information   |   Type     | Version                                                                                                                                                                     |
+--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
|              | Fault-Diag | 26.1.0                                                                                                                                                                   |
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
|              |            | 2. Note: Some faults involve too many devices; only the first 16 are displayed here. All faulty devices can be queried in diag_report.json.                                                                                             |
+--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
|              |   Status Code   | xxx                                                                                                                                                                      |
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
The diag job is complete.
</pre>

The report contains the following information:

### Root Cause Node Analysis

| Field             | Description                                        |
|-------------------|----------------------------------------------------|
| Root Cause Node   | Locates the server and NPU that caused the fault |
| First Error Node  | The device where the error first occurred in the task |
| Symptom | Describes the issue symptom                    |
| Root Cause Node Chain | Shows the fault propagation path              |

### Fault Event Analysis

| Field      | Description                        |
|------------|------------------------------------|
| Status Code | Unique fault identifier |
| Fault Category | Fault category, component, and module |
| Fault Device | Device where the fault occurs |
| Fault Name | Specific name of the fault |
| Fault Desc | Detailed description of the fault |
| Suggestion | Handling suggestions for the fault |
| Key Log | Raw logs corresponding to the fault |
| Key Propagation Chain | Fault propagation relationship chain |

> [!NOTE]
>
> - For more diagnosis output examples, see [diag Usage Examples](../06_api/03_command_diag.md#usage-examples).
> - For more field explanations, see [diag Diagnostic Result Description](../06_api/03_command_diag.md#diagnostic-result-description).

### Diagnostic Result File

The diagnostic result output directory contains:

```text
fault_diag_result/
└── diag_report.json    # Complete diagnosis results (in JSON format)
```

For the structure of the diagnostic result file `diag_report.json`, see [Diagnostic Result Output](../06_api/03_command_diag.md#diagnostic-result-output).

## Diagnosing Performance Degradation Issues

When a task does not exit abnormally but runs slowly, use the `-p` parameter to diagnose performance degradation issues.

- **Device resource analysis**: Checks NPU computation frequency reduction and CPU resource preemption issues.
- **Network congestion analysis**: Checks network congestion between nodes.

For use cases, refer to [Performance Degradation Diagnosis](../06_api/03_command_diag.md#performance-degradation-diagnosis).

## Notes

- When multiple faults are diagnosed, they are sorted by occurrence time. It is recommended to troubleshoot the earlier faults first.
- When there are too many fault devices, only the first 16 are displayed in the terminal. Complete information can be found in `diag_report.json`.
- For ascend-fd runtime error codes, see [Component Error Codes](../07_references/04_appendix.md#component-error-codes).
