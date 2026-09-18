# SuperPoD Fault Diagnosis

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:29:08.887Z pushedAt=2026-08-24T02:51:43.284Z -->

SuperPoD fault diagnosis is a diagnostic feature specifically designed for the Atlas A3 SuperPoD, which requires simultaneous consideration of three types of logs: Host (server), BMC (management controller), and LCNE (UnifiedBus switch).

Host logs refer to all logs other than BMC and LCNE logs in the [Log Collection Archive Directory Structure](02_log_collection.md#log-collection-archive-directory-structure).

## Use Cases

SuperPoD fault diagnosis supports the following two scenarios:

| Scenario                          | Applicable Condition                              | Description                                        |
|-----------------------------------|---------------------------------------------------|----------------------------------------------------|
| Automatic topology association    | Host, BMC, and LCNE logs are all available.        | The simplest method; topology is associated automatically. |
| Manual topology association       | Any of the host, BMC, or LCNE logs is missing     | The parsing results need to be associated manually. |

>[!NOTE]
>When the log level is configured too low, log flooding may overwrite key logs and make diagnosis impossible. Configure the log level appropriately.

## Scenario 1: Automatic Topology Association (Recommended)

**Prerequisite:** Host, BMC, and LCNE logs must all be available. In this scenario, the diagnostic command requires the `-s super_pod` parameter.

1. Perform the parsing operation.

    Example:

    ```shell
    ascend-fd parse -i <Host_log_directory_on_device_0> -o <SuperPoD_parsing_output_directory/host/worker-0>
    ascend-fd parse --lcne_log <LCNE_log_directory_on_device_0> -o <SuperPoD_parsing_output_directory/lcne/worker-0>
    ascend-fd parse --bmc_log <BMC_log_directory_on_device_0> -o <SuperPoD_parsing_output_directory/bmc/worker-0>
    ```

    > [!NOTE]
    >
    > - The host-side logs must contain the `dmidecode.txt` information.
    > - If there are multiple devices, parse them into the worker-1, worker-2, and other directories accordingly. You can also name them by device IP.

2. The parsed directory structure is as follows:

    ```text
    SuperPoD parsing output directory/
        ├── bmc
        |   └── worker-0
        |       ├── ascend-kg-analyzer.json
        |       ├── ascend-kg-parser.json
        |       └── server-info.json
        ├── host
        |   └── worker-0
        |       ├── ascend-kg-analyzer.json
        |       ├── ascend-kg-parser.json
        |       ├── ascend-rc-parser.json
        |       └── server-info.json
        └── lcne
            └── worker-0
                ├── ascend-kg-analyzer.json
                ├── ascend-kg-parser.json
                └── server-info.json
    ```

3. Perform diagnosis (the `-s super_pod` parameter is required).

    ```shell
    ascend-fd diag -i <diagnostic_input_directory> -o <diagnosis_output_directory> -s super_pod
    ```

    The diagnosis output is as follows:

    <pre>
    The diag job starts. Please wait. Job id: [***], run log file is [***].
    +-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                                   Ascend Fault-Diag Report                                                                  |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |   Version Information   |    Type    | Version                                                                                                                                            |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.1.0                                                                                                                                         |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Root Node Cause Analysis |    Type    | Description                                                                                                                                            |
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
    The diag job is complete.
    </pre>

4. Check the diagnostic result.

    ```text
    fault_diag_result/
    ├── diag_report.json # Diagnostic result
    └── topo_info.json # SuperPoD topology information
    ```

## Scenario 2: Manual Topology Association

When any of the host, BMC, or LCNE logs is missing, manual association is required to parse data from the same node into the same directory.

1. Perform parsing operations.

    ```shell
    # Specify both Host and BMC logs during cleaning.
    ascend-fd parse --process_log <Plog_log_directory_on_device_0> --bmc_log <BMC_log_directory_on_device_0> -o <SuperPoD_parsing_output_directory/worker-0>
    ascend-fd parse --process_log <Plog_log_directory_on_device_1> --bmc_log <BMC_log_directory_on_device_1> -o <SuperPoD_parsing_output_directory/worker-1>
    ascend-fd parse --process_log <Plog_log_directory_on_device_2> --bmc_log <BMC_log_directory_on_device_2> -o <SuperPoD_parsing_output_directory/worker-2>
    ```

    > [!NOTE]
    >
    > - The example assumes that LCNE logs are missing.
    > - The host-side logs use `--process_log` as an example. You can directly use `-i` to include other collected logs (such as `device_log`, `dl_log`, etc.).
    > - Parsing results from different devices should be placed under different worker directories.
    > - Different devices should correspond to different worker directories (for example, five devices correspond to worker-0 through worker-4 respectively).
    > - The example only shows the parsing results of three devices. In actual scenarios, you need to map different worker directories according to the number of devices.

2. The parsed directory structure is as follows:

    ```text
        SuperPod parsing output directory/
        ├── worker-0
        │   ├── ascend-kg-analyzer.json
        │   ├── ascend-kg-parser.json
        │   ├── ascend-rc-parser.json
        │   ├── device_ip_info.json
        │   ├── plog-parser-14121-1.log
        │   └── server-info.json
        ├── worker-1
        │   ├── ascend-kg-analyzer.json
        │   ├── ascend-kg-parser.json
        │   ├── ascend-rc-parser.json
        │   ├── device_ip_info.json
        │   ├── plog-parser-14139-1.log
        │   └── server-info.json
        └── worker-2
            ├── ascend-kg-analyzer.json
            ├── ascend-kg-parser.json
            ├── ascend-rc-parser.json
            ├── device_ip_info.json
            ├── mindie-cluster-info.json
            ├── plog-parser-14160-1.log
            └── server-info.json
    ```

3. Perform diagnosis (the `-s super_pod` parameter is not required).

    ```shell
    ascend-fd diag -i <diagnostic_input_directory> -o <diagnosis_output_directory>
    ```

    Taking the PD-disaggregated SuperPoD scenario as an example, the diagnostic output is as follows:

    <pre>
    The diag job starts. Please wait. Job id: [***], run log file is [***].
    +-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                        Ascend Fault-Diag Report                                                                             |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |   Version Information   |    Type    | Version                                                                                                                                                                        |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.1.0                                                                                                                                         |
    |              |   Driver   | 25.2.0                                                                                                                                          |
    |              |  Firmware  | 7.7.0.3.220                                                                                                                                     |
    |              |  Toolkit   | 8.1.RC1                                                                                                                                         |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Root Node Cause Analysis |    Type    | Description                                                                                                                                            |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Description    | Root cause node analysis has detected multiple suspected root cause nodes. Please prioritize investigating these nodes.                                                                                |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |  Root Node Cause  | ['xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx']                                                                |
    |              |  Symptom   |No error log information found in Plogs from all valid nodes. Root cause node cannot be located. Please also confirm whether this is a normal job.                                                            |
    |              |            | This inference instance experienced MindIE connection establishment failure. Please investigate the nodes where connection establishment failed.                                                                                        |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Fault Event Analysis |    Type    | Description                                                                                                                                            |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Description    | 1. Only the longest propagation chain for each faulty device is displayed.                                                                                                   |
    |              |            | 2. Multiple faults have been diagnosed and are sorted by priority. Please focus on the earlier faults for investigation.                                                                                      |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Suspected Root Cause Fault |   Status Code   | ******                                                                                                                                          |
    |              |  Fault Category  | Category: Network Component: Network Module: Network                                                                                                          |
    |              |  Fault Device  | ['xxxxxx']                                                                                                                           |
    |              |  Fault Name  | Link Down: NPU-side intermittent disconnection                                                                                                                        |
    |              |  Fault Desc  | An NPU network port on this server experienced a Link Down intermittent disconnection error, and the intermittent disconnection duration exceeded 30 seconds.                                                                                   |
    |              |  Suggestion  | 1. Please contact the physical network operations team to collect switch logs and check for hardware issues (whether optical modules are present, whether switch links disconnected intermittently).                                       |
    |              |  Key Log  | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    |              | Key Propagation Chain | ['xxxxxx']                                                                                                                                      |
    |              |            | Comp_Network_Custom_01 (Link Down: NPU-side intermittent disconnection)-> 0x81078603 (Network port link status change, Up->Down)                                                    |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+

    ============================
    Instance: xxx.xxx.xx8.201-xxx.xxx.xx2.204-xxx.xxx.8.183-xxx.xxx.x7.203
    Node: ['xxx.xxx.xx8.201', 'xxx.xxx.xx2.204', 'xxx.xxx.8.183', 'xxx.xxx.x7.203']
    +-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                     Ascend Fault-Diag Report                                                                                |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |   Version Information   |    Type    | Version                                                                                                                                            |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.1.0                                                                                                                                         |
    |              |   Driver   | 25.2.0                                                                                                                                          |
    |              |  Firmware  | 7.7.0.3.220                                                                                                                                     |
    |              |  Toolkit   | 8.1.RC1                                                                                                                                         |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Root Node Cause Analysis |    Type    | Description                                                                                                                                            |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Description    |  A waiting relationship exists between some devices. An example is shown in the "Inter-device Waiting Chain".                                                                                            |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |  Root Node Cause  | ['xxxxxx']                                                                                                                                      |
    |              |  Symptom  | All nodes used by the training/inference job reported operator dispatch connection establishment timeout errors in Plogs. The time interval between the earliest error node and the latest error node did not exceed the configured timeout (480s). Please prioritize investigating devices involved in mutual waiting or at the end of the waiting chain.                                                                                                                                                            |
    |              | Inter-device Waiting Chain | worker-2 device-0 -> worker-5 device-0                                                                                                          |
    |              |  First Error Node  | worker-2 device-0: 2025-06-23-11:10:41.730228                                                                                                   |
    |              |  Last Error Node  | worker-3 device-6: 2025-06-23-11:10:44.883255                                                                                                   |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Fault Event Analysis |    Type    | Description                                                                                                                                            |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Description    | Multiple faults have been diagnosed and are sorted by priority. Please focus on the earlier faults for investigation.                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Suspected Root Cause Fault |   Status Code   | ******                                                                                                                                          |
    |              |  Fault Category  | Category: Network Component: Switch Module: Chip                                                                                                              |
    |              |  Fault Device  | ['worker-2']                                                                                                                                    |
    |              |  Fault Name  | Forwarding engine overall function failure                                                                                                                            |
    |              |  Fault Desc  | Fatal internal fault in forwarding chip.                                                                                                                          |
    |              |  Suggestion  | 1. Please contact Huawei engineers.                                                                                                                       |
    |              |  Key Log  | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |   Status Code   | ******                                                                                                                                          |
    |              |  Fault Category  | Category: Software Component: MindIE Module: LLM                                                                                                              |
    |              |  Fault Device  | ['xxxxxx']                                                                                                                                      |
    |              |  Fault Name  | BackendConfig configuration parameter validation failure                                                                                                                   |
    |              |  Fault Desc  | Invalid configuration parameter.                                                                                                                                |
    |              |  Suggestion  | 1. Please contact Huawei engineers.                                                                                                                       |
    |              |  Key Log  | ******                                                                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    The diag job is complete.
    </pre>

4. Check the diagnostic results.

    ```text
    fault_diag_result/
    ├── diag_report_xxx.xxx.xx8.201-xxx.xxx.xx2.204-xxx.xxx.8.183-xxx.xxx.x7.203.json # Diagnostic result 1
    └── diag_report_xxx.xxx.xx7.11.json #Diagnostic result 2
    ```
