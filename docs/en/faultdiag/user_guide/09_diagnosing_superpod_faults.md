# Diagnosing SuperPoD Faults<a name="ZH-CN_TOPIC_0000002333351060"></a>

The SuperPoD fault diagnosis involves three scenarios: manually configured SuperPoD topology, non-manually configured SuperPoD topology, and missing host logs.

- In [non-manually configured SuperPoD topology](#non-manually-configured-SuperPoD-topology) scenario, any of the BMC, Host, and LCNE logs cannot be missing.
- If LCNE or BMC logs are missing, see [manually configured SuperPoD topology](#manually-configured-SuperPoD-topology).
- If host logs are missing, store the cleaning output results in the same folder for diagnosis by referring to [host log missing](#host-log-missing). 

## Non-manually Configured SuperPoD Topology<a name="section181656598409"></a>

When performing diagnosis with `-s super_pod`, ensure that BMC, host, and LCNE logs are all available.

1. Create a directory for storing the SuperPoD fault diagnosis result.

    ```shell
    mkdir *Directory for storing the SuperPoD fault diagnosis result*
    ```

2. Place the SuperPoD fault diagnosis cleaning result as follows.

    ```text
    Directory for storing the SuperPoD fault diagnosis cleaning result/
    ├── bmc
    │   ├── bmc_xxx.xx.xx.xx4_1
    │   │   ├── ascend-kg-analyzer.json
    │   │   ├── ascend-kg-parser.json
    │   │   └── server-info.json
    │   └── bmc_xxx.xx.xx.xx5_1
    │       ├── ascend-kg-analyzer.json
    │       ├── ascend-kg-parser.json
    │       └── server-info.json
    ├── host
    │   ├── log_collect_node-29-121_20250616
    │   │   ├── ascend-kg-analyzer.json
    │   │   ├── ascend-kg-parser.json
    │   │   ├── ascend-rc-parser.json
    │   │   ├── plog-parser-9891-1.log
    │   │   └── server-info.json
    │   └── log_collect_node-29-124_20250616
    │       ├── ascend-kg-analyzer.json
    │       ├── ascend-kg-parser.json
    │       ├── ascend-rc-parser.json
    │       ├── device_ip_info.json
    │       ├── plog-parser-10802-1.log
    │       ├── plog-parser-1132-0.log

    │       └── server-info.json
    └── lcne
        ├── xxx.xx.xx.xx6
        │   ├── ascend-kg-analyzer.json
        │   ├── ascend-kg-parser.json
        │   └── server-info.json
        └── xxx.xx.xx.xx7
            ├── ascend-kg-analyzer.json
            ├── ascend-kg-parser.json
            └── server-info.json
    ```

3. Start diagnosis.

    By default, the data of the fault event module is returned for SuperPoD fault diagnosis.

    ```shell
    ascend-fd diag -i *Diagnosis_input_directory* -o *Diagnosis_result_output_directory* -s super_pod
    ```

    Command output:

    ```ColdFusion
    The diag job starts. Please wait. Job id: [***], run log file is [***].
    +-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                                   Ascend Fault-Diag Report                                                                  |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Version information | Type | Version number                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.0.0                                                                                                                                         |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Root cause node analysis |    Type    | Description                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |  Root cause node | ['xxxxxxxxxxx']                                                                                             |
    |              |  Symptom  | No timeout information is found in any Plog. A node may have a process exit exception or a suspended state.                                                                                        |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Fault event analysis |    Type    | Description                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Note   | 1. Some sub-items of this analysis module fail to be executed, which may affect the accuracy of the diagnosis result. Check the detailed information in diag_report.json.                                       |
    |              |            | 2. Multiple faults are diagnosed and prioritized. Focus on the top-ranked faults.                                                                                       |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Potential root cause  |   Status code   | ******                                                                                                                                          |
    |              |  Fault type | Type: Network Component: Switch Module: Chip                                                                                                              |
    |              |  Faulty device  | ['LCNE:xxx.xx.xx.xx5']                                                                                                                          |
    |              |  Fault name | Forwarding engine failure                                                                                                                            |
    |              |  Fault description | The LAN switch chip is unstable.                                                                                                                           |
    |              |  Suggestion | 1. Contact Huawei technical support.                                                                                                                                      |
    |              |  Key log  | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |   Status code   | ******                                                                                                                                          |
    |              |  Fault type | Type: Network Component: Switch Module: Chip                                                                                                              |
    |              |  Faulty device  | ['LCNE:xxx.xx.xx.xx5']                                                                                                                          |
    |                | Fault name | Forwarding chip port reduced to 1/2 lane capacity                                                                                                                    |
    |              | Fault Description | The forwarding chip port has been reduced to half its lanes. (L1<-->CPU).                                                                                                        |
    |              |  Suggestion | 1. Contact Huawei technical support.                                                                                                                                      |
    |              |  Key log  | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    The diag job is complete.
    ```

4. Check the diagnosis result.

    ```text
    fault_diag_result/
    ├── diag_report.json # Diagnosis result
    └── topo_info.json # SuperPoD topology information
    ```

>[!NOTE]
>
>- For more details about the example provided in this section, refer to [SuperPoD Log Cleaning and Diagnosis Script](https://gitcode.com/Ascend/mindxdl-deploy/tree/master/super_pod_diag) to decompress, clean, and diagnose superPoD logs in batches.
>- When the log level is low, key logs may be refreshed and cannot be diagnosed. The involved environment variables include `ASCEND_GLOBAL_EVENT_ENABLE`, `HCCL_ENTRY_LOG_ENABLE`, `ASCEND_GLOBAL_LOG_LEVEL`, and `ASCEND_MODULE_LOG_LEVEL`. For more information, see [Environment Variable List](https://www.hiascend.com/document/detail/zh/canncommercial/900/maintenref/envvar/envref_07_0001.html) in the *CANN Environment Variable Reference*.

## Manually Configured SuperPoD Topology<a name="section117571749184019"></a>

During the cleaning, you need to manually associate the BMC, host, and LCNE logs and summarize the cleaning results to the same directory.

Example:

```ColdFusion
ascend-fd parse --host_log parse_input/host/xxx.xx.xx.131/host_log/   --mindie_log parse_input/host/xxx.xx.xx.131/mindie/ --process_log parse_input/host/xxx.xx.xx.131/process_log/  --bmc_log parse_input/bmc/worker-104 --lcne_log parse_input/lcne/worker-204 -o *Cleaning result output directory*/worker-1
ascend-fd parse --host_log parse_input/host/xxx.xx.xx.129/host_log/   --mindie_log parse_input/host/xxx.xx.xx.129/mindie/ --process_log parse_input/host/xxx.xx.xx.129/process_log/  --bmc_log parse_input/bmc/worker-102 --lcne_log parse_input/lcne/worker-202 -o *Cleaning result output directory*/worker-2
ascend-fd parse --host_log parse_input/host/xxx.xx.xx.127/host_log/   --mindie_log parse_input/host/xxx.xx.xx.127/mindie/ --process_log parse_input/host/xxx.xx.xx.127/process_log/  --bmc_log parse_input/bmc/worker-100 --lcne_log parse_input/lcne/worker-200 -o *Cleaning result output directory*/worker-3
ascend-fd parse --host_log parse_input/host/xxx.xx.xx.130/host_log/   --mindie_log parse_input/host/xxx.xx.xx.130/mindie/ --process_log parse_input/host/xxx.xx.xx.130/process_log/  --bmc_log parse_input/bmc/worker-103 --lcne_log parse_input/lcne/worker-203 -o *Cleaning result output directory*/worker-4
ascend-fd parse --host_log parse_input/host/xxx.xx.xx.128/host_log/   --mindie_log parse_input/host/xxx.xx.xx.128/mindie/ --process_log parse_input/host/xxx.xx.xx.128/process_log/  --bmc_log parse_input/bmc/worker-101 --lcne_log parse_input/lcne/worker-201 -o *Cleaning result output directory*/worker-5
```

1. Place the SuperPoD fault diagnosis cleaning result as follows.

    ```text
    Directory for storing the SuperPoD fault diagnosis cleaning result/
    ├── worker-1
    │   ├── ascend-kg-analyzer.json
    │   ├── ascend-kg-parser.json
    │   ├── ascend-rc-parser.json
    │   ├── device_ip_info.json
    │   ├── plog-parser-14121-1.log
    │   └── server-info.json
    ├── worker-2
    │   ├── ascend-kg-analyzer.json
    │   ├── ascend-kg-parser.json
    │   ├── ascend-rc-parser.json
    │   ├── device_ip_info.json
    │   ├── plog-parser-14139-1.log
    │   └── server-info.json
    ├── worker-3
    │   ├── ascend-kg-analyzer.json
    │   ├── ascend-kg-parser.json
    │   ├── ascend-rc-parser.json
    │   ├── device_ip_info.json
    │   ├── mindie-cluster-info.json
    │   ├── plog-parser-14160-1.log
    │   └── server-info.json
    ├── worker-4
    │   ├── ascend-kg-analyzer.json
    │   ├── ascend-kg-parser.json
    │   ├── ascend-rc-parser.json
    │   ├── device_ip_info.json
    │   ├── plog-parser-14175-1.log
    │   └── server-info.json
    └── worker-5
        ├── ascend-kg-analyzer.json
        ├── ascend-kg-parser.json
        ├── ascend-rc-parser.json
        ├── device_ip_info.json
        ├── plog-parser-19333-1.log
        └── server-info.json
    ```

2. Start diagnosis.

    By default, the data of the fault event module is returned for SuperPoD fault diagnosis.

    ```shell
    ascend-fd diag -i *Diagnosis_input_directory* -o *Diagnosis_result_output_directory*
    ```

    Command output:

    ```ColdFusion
    The diag job starts. Please wait. Job id: [***], run log file is [***].
    +-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                        Ascend Fault-Diag Report                                                                             |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Version information | Type | Version number                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.0.0                                                                                                                                         |
    |              |   Driver   | 25.2.0                                                                                                                                          |
    |              |  Firmware  | 7.7.0.3.220                                                                                                                                     |
    |              |  Toolkit   | 8.1.RC1                                                                                                                                         |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Root cause node analysis |    Type    | Description                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Description    | Multiple potential root cause nodes are detected. Preferentially check them.                                                                                |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |  Root cause node | ['xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx']                                                                |
    |              |  Symptom  |  No error information is found in the Plog of all valid nodes. As a result, the root cause node cannot be located. In addition, check whether the job is normal.                                                            |
    |              |            | MindIE fails to establish a link in the inference instance. Check the node with link establishment failure.                                                                                         |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Fault event analysis |    Type    | Description                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Description    | 1. Only the longest link of each faulty device is displayed in the key propagation chain.                                                                                                   |
    |              |            | 2. Multiple faults are diagnosed and prioritized. Focus on the top-ranked faults.                                                                                       |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Potential root cause  |   Status code   | ******                                                                                                                                          |
    |              |  Fault type | Type: Network Component: Network Module: Network                                                                                                          |
    |              |  Faulty device | ['xxxxxx']                                                                                                                           |
    |              |  Fault name | Link Down: NPU intermittent disconnection                                                                                                                        |
    |              |  Fault description | A Link Down intermittent disconnection error occurs on an NPU network port of the server, and the intermittent disconnection duration exceeds 30s. |
    |              | Handling suggestions  | 1. Contact physical network O&M personnel to collect switch logs and check whether hardware problems occur (for example, whether the optical module is properly installed or the switch link is intermittently disconnected).                                      |
    |              |  Key log  | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    |              | Key propagation chain | ['xxxxxx']                                                                                                                                      |
    |              |            | Comp_Network_Custom_01 (Link Down: intermittent disconnection on NPU ports -> 0x81078603 (Network port link status change, Up -> Down)                                                   |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+

    ============================
    Instance name: xxx.xxx.xx8.201-xxx.xxx.xx2.204-xxx.xxx.8.183-xxx.xxx.x7.203
    Node name: ['xxx.xxx.xx8.201', 'xxx.xxx.xx2.204', 'xxx.xxx.8.183', 'xxx.xxx.x7.203']
    +-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                     Ascend Fault-Diag Report                                                                                |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Version information | Type | Version number                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.0.0                                                                                                                                         |
    |              |   Driver   | 25.2.0                                                                                                                                          |
    |              |  Firmware  | 7.7.0.3.220                                                                                                                                     |
    |              |  Toolkit   | 8.1.RC1                                                                                                                                         |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Root cause node analysis |    Type    | Description                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Note    | Some devices are in waiting status, which is displayed in "Inter-device waiting chain".                                                                                            |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |  Root cause node ['xxxxxx']                                                                                                                                      |
    |              |  Symptom| The Plog of all nodes running training/inference jobs reports a timeout error during operator delivery link setup, and the time difference between the first and last reported error nodes is within the 480s threshold. In this case, check the devices involved in the waiting chain or those at the end of the dependency.                                                                                                                                                            |
    |              | Inter-device waiting chain | worker-2 device-0 -> worker-5 device-0                                                                                                          |
    |              |  First error node  | worker-2 device-0: 2025-06-23-11:10:41.730228                                                                                                   |
    |              |  Last error node  | worker-3 device-6: 2025-06-23-11:10:44.883255                                                                                                   |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Fault event analysis |    Type    | Description                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Note      | Multiple faults are diagnosed and prioritized. Focus on the top-ranked faults.                                                                                       |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Potential root cause  |   Status code   | ******                                                                                                                                          |
    |              |  Fault type | Type: Network Component: Switch Module: Chip                                                                                                              |
    |              |  Faulty device | ['worker-2']                                                                                                                                    |
    |              |  Fault name | Forwarding engine failure                                                                                                                            |
    |              |  Fault description | A fatal fault occurs in the forwarding chip.                                                                                                                          |
    |              |  Suggestion | 1. Contact Huawei technical support.                                                                                                                                      |
    |              |  Key log  | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |   Status code   | ******                                                                                                                                          |
    |              |  Fault type | Type: Software Component: MindIE Module: LLM                                                                                                                             |
    |              |  Faulty device | ['xxxxxx']                                                                                                                                 |
    |              |  Fault name | BackendConfig verification failure                                                                                                                                  |
    |              |  Fault description | The configured parameter is invalid.                                                                                                                                               |
    |              |  Suggestion | 1. Contact Huawei technical support.                                                                                                                                      |
    |              |  Key log  | ******                                                                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    The diag job is complete.
    ```

3. Check the diagnosis results.

    ```ColdFusion
    fault_diag_result/
    ├── diag_report_xxx.xxx.xx8.201-xxx.xxx.xx2.204-xxx.xxx.8.183-xxx.xxx.x7.203.json # Diagnosis result 1
    └── diag_report_xxx.xxx.xx7.11.json # Diagnosis result 2
    ```

## Host Log Missing<a name="section2910173013308"></a>

If the host logs are missing, store the BMC and LCNE cleaning results in the same directory. The following example illustrates a diagnostic scenario involving only LCNE logs.

1. Create a directory for storing the SuperPoD fault diagnosis result.

    ```shell
    mkdir *Directory for storing the SuperPoD fault diagnosis result*
    ```

2. Place the SuperPoD fault diagnosis cleaning result as follows.

    ```text
    Directory for storing the SuperPoD cleaning result/lcne/
    ├── worker-200
    │   ├── ascend-kg-analyzer.json
    │   ├── ascend-kg-parser.json
    │   └── server-info.json
    ├── worker-201
    │   ├── ascend-kg-analyzer.json
    │   ├── ascend-kg-parser.json
    │   └── server-info.json
    ├── worker-202
    │   ├── ascend-kg-analyzer.json
    │   ├── ascend-kg-parser.json
    │   └── server-info.json
    ├── worker-203
    │   ├── ascend-kg-analyzer.json
    │   ├── ascend-kg-parser.json
    │   └── server-info.json
    └── worker-204
        ├── ascend-kg-analyzer.json
        ├── ascend-kg-parser.json
        └── server-info.json
    ```

3. Start diagnosis.

    By default, the data of the fault event module is returned for SuperPoD fault diagnosis.

    ```shell
    ascend-fd diag -i *Diagnosis input directory*/lcne -o *Diagnosis result output directory*
    ```

    Command output:

    ```ColdFusion
    The diag job starts. Please wait. Job id: [***], run log file is [***].
    +-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                      Ascend Fault-Diag Report                                                                               |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Version information | Type | Version number                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.0.0                                                                                                                                         |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Root cause node analysis |    Type    | Description                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Description | If no root cause node is diagnosed, fault event analysis attempts to detect all devices.                                                                                                               |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |  Root cause node | ['Unknown Device']                                                                                                                                             |
    |              |  Symptom    |No valid Plog file is found. As a result, the root cause node cannot be located. Check whether the Plog file exists.                                                                               |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Fault event analysis |    Type    | Description                                                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |    Note      | Multiple faults are diagnosed and prioritized. Focus on the top-ranked faults.                                                                                       |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    | Potential root cause  |   Status code   | ******                                                                                                                                          |
    |              |  Fault type | Type: Network Component: Switch Module: Chip                                                                                                              |
    |              |  Faulty device  | ['xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx']                                                                                              |
    |                | Fault name | Forwarding chip port reduced to 1/2 lane capacity                                                                                                                    |
    |              | Fault Description | The forwarding chip port has been reduced to half its lanes. (L1<-->CPU).                                                                                                        |
    |              |  Suggestion | 1. Contact Huawei technical support.                                                                                                                                      |
    |              |  Key log  | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |   Status code   | ******                                                                                                                                          |
    |              |  Fault type | Type: Network Component: Switch Module: Chip                                                                                                              |
    |              |  Faulty device  | ['xxxxxx', 'xxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx']                                                                                              |
    |              |  Fault name | Forwarding chip port down                                                                                                                            |
    |              |  Fault description  | The forwarding chip port is linkdown: L1<-->CPU.                                                                                                                |
    |              |  Suggestion | 1. Contact Huawei technical support.                                                                                                                                      |
    |              |  Key log  | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |   Status code   | ******                                                                                                                                          |
    |              |  Fault type | Type: Network Component: Switch Module: Chip                                                                                                              |
    |              |  Faulty device  | ['xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx']                                                                                              |
    |              |  Fault name | Partial forwarding engine failure                                                                                                                            |
    |              | Fault description  | The forwarding chip is incorrectly configured.                                                                                                                           |
    |              |  Suggestion | 1. Contact Huawei technical support.                                                                                                                                      |
    |              |  Key log  | ******                                                                                                                                          |
    |              |            | ******                                                                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |   Status code   | ******                                                                                                                                          |
    |              |  Fault type | Type: Network Component: Switch Module: Chip                                                                                                              |
    |              |  Faulty device  | ['xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx']                                                                                              |
    |              |  Fault name | Forwarding engine failure                                                                                                                            |
    |              |  Fault description | The LAN switch chip is unstable.                                                                                                                           |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    |              |   Status code   | ******                                                                                                                                          |
    |              |  Fault type | Type: Network Component: Switch Module: Chip                                                                                                              |
    |              |  Faulty device  | ['xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxx', 'xxxxxxx']                                                                                             |
    |              |  Fault name | Forwarding engine failure                                                                                                                            |
    |              |  Fault description | A fatal fault occurs in the forwarding chip.                                                                                                                          |
    +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------+
    The diag job is complete.
    ```

4. Check the diagnosis result.

    ```ColdFusion
    fault_diag_result/
    ├── diag_report.json # Diagnosis result
    ```
