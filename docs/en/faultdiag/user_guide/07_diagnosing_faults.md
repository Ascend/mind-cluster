# Diagnosing Faults<a name="ZH-CN_TOPIC_0000001541948906"></a>

1. Create a diagnosis result output directory.

    ```shell
    mkdir *Diagnosis result output directory*
    ```

2. Start diagnosis.

    - If the `--performance(-p)` parameter is not specified, the program cleans only the data of the root cause node and fault event modules by default.

        ```shell
        ascend-fd diag -i *Diagnosis input directory* -o *Diagnosis result output directory*
        ```

        The following information is displayed when a training job exits abnormally during diagnosis:

        ```ColdFusion
        The diag job starts. Please wait. Job id: [****], run log file is [****].
        +------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |                                                                                       Ascend Fault-Diag Report                                                                                       |
        +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+                                                                                                                                                                       | Version information | Type | Version |
        +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |              | Fault-Diag | 26.0.0                                                                                                                                                                    |
        +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        | Root cause node analysis |    Type    | Description                                                                                                                                                                     |
        +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |              |    Description    | The worker ID and device ID of the peer node in the fault node chain cannot be identified. Confirm the device through IP.                                                                                       |
        +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |              |  Root cause node  | ['worker-0 device-2']                                                                                                                                                    |
        |              |  First error node  | worker-1 device-2: 2023-09-01-06:35:52.960343                                                                                                                            |
        |              |  Symptom  | Excessive RoCE retransmissions (ERROR CQE) occur on some nodes. These nodes may be the root cause nodes. Please check.                                                                                                    |
        |              | Root cause node chain | ['worker-0 device-2 -> 192.168.102.220']                                                                                                                                 |
        +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        | Fault event analysis |    Type    | Description                                                                                                                                                          |
        +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |              |    Note | 1. Multiple faults are diagnosed and sorted by occurrence time. Check the faults that occurred earlier.                                                                                                              |
        |              |            | 2. Only 16 faulty devices are displayed. All faulty devices can be queried in the diag_report.json file.                                                                                             |
        +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |              |   Status code   |  xxx                                                                                                                                                                      |
        |              |  Fault type | Type: Network Component: Network Module: Network                                                                                                                                   |
        |              | Faulty device | ['worker-0 device-2']                                                                                                                                                    |
        |              |  Fault name | Link Down: NPU intermittent disconnection                                                                                                                                                 |
        |             |  Fault description  | The link of an NPU network port on the server is down for more than 30s.                                                                                                            |
        |              |  Solution  | 1. Contact physical network O&M personnel to collect switch logs and check whether hardware faults occur (for example, whether optical modules work properly and whether switch links are intermittently disconnected).                                                              |
        |              |  Key log  | /usr/local/Ascend/driver/tools/hccn_tool -i 2 -link_stat -g                                                                                                               |
        |              |            | [devid 2]current time        : Fri Sep  1 06:37:26 2023                                                                                                                  |
        |              |            | [devid 2]link up count       : 2                                                                                                                                         |
        |              |            | [devid 2]link change records :                                                                                                                                           |
        |              |            | [devid 2]    Fri Sep  1 06:34:43 2023    LINK DOWN                                                                                                                       |
        |              |            | [devid 2]    Thu Aug 31 07:30:46 2023    LINK UP                                                                                                                         |
        |              |            | [devid 2]    Thu Aug 31 07:30:44 2023    LINK DOWN                                                                                                                       |
        |              |            | [devid 2]    Thu Aug 31 07:30:43 2023    LINK UP                                                                                                                         |
        |              | Key propagation chain | ['worker-0']                                                                                                                                                             |
        |              |            | Fault code 1 (Link Down: NPU intermittent disconnection error) -> Fault code 2 (excessive RDMA retransmissions) -> Fault code 3 (notify wait timeout)                                                                           |
        +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        The diag job is complete.
        ```

        The following information is displayed when exceptions of an inference job with multiple instances are diagnosed:

        ```ColdFusion
        The diag job starts. Please wait. Job id: [****], run log file is [****].

        ============================
        Instance name: ****
        Node name: [****, ****]
        +--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |                                                                                  Ascend Fault-Diag Report                                                                                  |
        +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
        | Version information | Type | Version                                                                                                                                                           |
        +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |              | Fault-Diag | 26.0.0                                                                                                                                                        |
        +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
        | Root cause node analysis |    Type    | Description                                                                                                                                                           |
        +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |              |    Description | If no root cause node is diagnosed, fault event analysis attempts to detect all devices.                                                                                                               |
        +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |              |  Root cause node | ['Unknown Device']                                                                                                                                             |
        |              |  Symptom  |  No error information is found in the Plog of all valid nodes. As a result, the root cause node cannot be located. In addition, check whether the job is normal.                                                                           |
        +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
        | Fault event analysis |    Type    | Description                                                                                                                                                           |
        +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+
        | Suspected root cause fault |   Status code   | xxx | xxx                                                                                                                                                            |
        |              |  Fault type | Type: Software Component: MindIE Module: LLM                                                                                                                             |
        |              | Faulty device | ['worker-0 device-2']                                                                                                                                          |
        |              |  Fault | BackendConfig verification failure                                                                                                                                  |
        |              |  Fault description | The configured parameter is invalid.                                                                                                                                               |
        |              |  Handling Suggestions | 1. Contact Huawei technical support.                                                                                                                                      |
        |              |  Key log  | [2025-06-17 16:59:10.282+08:00] [97] [147] [server] [WARN] [llm_infer_engine.cpp:117] : MIE05E040000[llm_backend] get model instance processing request failed |
        +--------------+------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------+

        ============================
        Instance name: ****
        Node name: [****, ****]
        +-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |                                                                                   Ascend Fault-Diag Report                                                                                    |
        +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        | Version information | Type | Version                                                                                                                                                           |
        +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |              | Fault-Diag | 26.0.0                                                                                                                                                           |
        +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        | Root cause node analysis |    Type    | Description                                                                                                                                                           |
        +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        |              |  Root cause node | ['worker-1 device-11']                                                                                                                                            |
        |              |  Symptom  | No timeout error is recorded in the Plog of all nodes. The node with error logs is the suspected root cause node. Check the node.                                                                                 |
        |              |  First error node  | xx.xx.xx.xx device-11: 2025-06-17-17:03:02.614708                                                                                                                 |
        |              |  Last error node  | xx.xx.xx.xx device-11: 2025-06-17-17:03:02.614708                                                                                                                 |
        +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        | Fault event analysis |    Type    | Description                                                                                                                                                           |
        +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        | Suspected root cause fault |   Status code   | xxx                                                                                                                                                            |
        |              |  Fault type | Type: Software Component: MindIE Module: LLM                                                                                                                                |
        |              |  Faulty device | ['worker-1 device-11']                                                                                                                                            |
        |              |  Fault | BackendConfig verification failure                                                                                                                             |
        |              |  Fault description | The configured parameter is invalid.                                                                                                                                              |
        |              |  Handling Suggestions | 1. Contact Huawei technical support.                                                                                                                                      |
        |              |  Key log  | [2025-06-17 17:01:44.605+08:00] [2015] [5415] [server] [WARN] [llm_infer_engine.cpp:117] : MIE05E040000[llm_backend] get model instance processing request failed |
        +--------------+------------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------+
        ...
        The diag job is complete.
        ```

    - If the `--performance (-p)` parameter is specified, all modules are diagnosed.

        ```shell
        ascend-fd diag -i *Diagnosis input directory* -o *Diagnosis result output directory* --performance
        ```

        The following is a command output example of diagnosing performance issues during training:

        ```ColdFusion
        +--------------------------------------------------------------------------------------------------------------------------------------------------+
        |                                                             Ascend Fault-Diag Report                                                             |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        |   Version information   |   Type   | Version                                                                                                                   |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        |              |Fault-Diag| 26.0.0                                                                                                                  |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        | Root cause node analysis |    Type    | Description                                                                                           |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        |              |  Note | If the root cause node is not diagnosed, the fault event analysis module will attempt to detect all devices.                                                                       |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        |              |  Root cause node | ['Unknown Device']                                                                                              |
        |              |  Symptom | The Plog of all valid nodes does not contain error log information and heartbeat information. The root cause node cannot be located. Check whether the training job is normal.|
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        | Fault event analysis |   Type   | Description                                                                                                                  |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        |              |  Status code | NORMAL_OR_UNSUPPORTED                                                                                                 |
        |              | Result description | No result is displayed in the fault event analysis module. The possible cause is that the training job is normal. If the training job is interrupted unexpectedly and the fault cannot be rectified, contact Huawei technical support. |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        | Device resource analysis |    Type    | Description                                                                                                            |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        |              |    Note    | Some analysis sub-items of this analysis module fail to be executed, and the diagnosis result may be affected and inaccurate. Check the detailed information of the module logs.                         |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        |              |  Status code | xxx                                                                                                                    |
        |              |  Faulty device | worker-0                                                                                                       |
        |              | Fault process | [2381084, 2381097]                                                                                                     |
        |              | Fault existing time | [('2023-08-11 02:18:00', '2023-08-11 02:21:00'),' Fault probability: 0.663'] |
        |              | Fault name | CPU preemption (preempted by some processes)                                                                                                |
        |              | Fault description | Device resources are abnormal, but CPU resource preemption occurs on some training processes.                                                                        |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        | Network congestion analysis | Type | Description                                                                                                            |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        |              |  Status code | xxx                                                                                                                    |
        |              | -------- |                                                                                                                        |
        |              |  Faulty device | worker-0                                                                                                       |
        |               | Faulty node | | ['device-0', 'device-1', 'device-2', 'device-3', 'device-4', 'device-5', 'device-6', 'device-7']                       |
        |              | -------- |                                                                                                                        |
        |              | Faulty device | worker-1                                                                                                               |
        |               | Faulty node | | ['device-0', 'device-1', 'device-2', 'device-3', 'device-4', 'device-5', 'device-6', 'device-7']                       |
        |              | -------- |                                                                                                                        |
        |              |  Fault name | Link congestion                                                                                                    |
        |              |  Fault description  | Some communication links are congested due to conflicts.                                                                                      |
        |              | Suggestion | Check the switch routing policy.                                                                                               |
        +--------------+----------+------------------------------------------------------------------------------------------------------------------------+
        ```

    The following table describes key parameters in the command output.

    **Table 1** Key parameters

    |Level-1 Parameter|Level-2 Parameter|Description|
    |--|--|--|
    |Root cause node analysis|-|Used to analyze the root cause of a fault.|
    |-|Root cause node|Device where the root cause device is located.|
    |-|Symptom|Symptom of the root cause node.|
    |-|First error node|Device where the first fault occurs|
    |-|Last error node|Device where the latest fault occurs|
    |-|Plog logs |If the root cause node is **Unknown Device** and the first error node exists, the first 10 lines of Plog logs starting of the first error log of the first error node are displayed.|
    |-|Log description|Used to describe the original plog path of the first error node when the root cause node is **Unknown Device** and the first error node exists.|
    |-|Root cause node chain|Propagation relationship of faulty nodes when the retransmission times exceed the threshold.|
    |-|Inter-device waiting chain|Propagation relationship of nodes where the Socket/Notify timeout fault occurs.|
    |Fault event analysis|-|Used to analyze the root cause of the device where the root cause node is located.|
    |-|Status code|<ul><li>If a fault is diagnosed, a specific fault code is displayed. </li><li>If no fault is diagnosed, `NORMAL OR UNSUPPORTED` is displayed.</li></ul>|
    |-|Fault name|Specific fault name.|
    |-|Fault type|Type of a fault and the component and module where the fault occurs.|
    |-|Faulty device|Device where a fault occurs.|
    |-|Fault description|Detailed description of a fault.|
    |-|Suggestion|Handling suggestions for a fault.|
    |-|Key log|Key log of a fault.|
    |-|Key propagation chain|Used to display the longest fault link.|
    |Device resource analysis|-|Used to analyze the resource status of devices.|
    |-|Status code|<ul><li>If a fault is diagnosed, a specific fault code is displayed. </li><li>When no fault is diagnosed, `NODE_DIAGNOSIS_NORMAL` is displayed.</li></ul>|
    |-|Faulty device|Name of the node where a fault occurs.|
    |-|Fault name|Specific fault name.|
    |-|Suggestion|Handling suggestions for a fault.|
    |Network congestion analysis|-|Used to analyze the network status between nodes.|
    |-|Status code|<ul><li>If a fault is diagnosed, a specific fault code is displayed. </li><li>If no fault is diagnosed, `NET_DIAGNOSIS_NORMAL` is displayed.</li></ul>|
    |-|Faulty device|Name of the node where a fault occurs.|
    |-|Fault name|Specific fault name.|
    |-|Suggestion|Handling suggestions for a fault.|

    >[!NOTE]
    >- If the root cause node analysis and fault event analysis results are displayed in the command output, the current fault has caused the training job to exit abnormally.
    >- Device resource analysis and network congestion analysis are performed only when the root cause node is not diagnosed and no fault event analysis result is displayed in the command output. This indicates that the current fault is a performance deterioration issue and does not cause the training job to exit unexpectedly.
    >- Some key parameters are available only in specific scenarios.

    After the diagnosis is complete, you can perform optimization based on the recommended solution in the diagnosis result.

    ```text
    Diagnosis result output directory
    ├── fault_diag_result
        ├── diag_report.json    # Diagnosis result
        ├── diag_report_{Instance name}.json    # Diagnosis result of multi-instance inference
    ```

    >[!NOTE]
    >- If an error occurs during fault diagnosis, the `description` (or analysis failure) field in the fault event analysis command output will display the failure information. To view all exception information, check the `diag_report.json` file.
    >- If the `--performance(-p)` parameter is not specified when the diagnosis command is executed, only the root cause node analysis and fault event analysis modules are executed. The corresponding JSON file contains only the results of the two modules.
    >- Currently, fault diagnosis is not supported for multi-instance inference clusters that have undergone recovery (with/without redundant resources). Due to this recovery feature, MindIE pods are restarted and instance information is updated. The logs of MindIE components generated before and after the restart are stored in the same directory, preventing components from correctly separating instances and logs.
    >- When the log level is low, key logs may be refreshed and cannot be diagnosed. The involved environment variables include `ASCEND_GLOBAL_EVENT_ENABLE`, `HCCL_ENTRY_LOG_ENABLE`, `ASCEND_GLOBAL_LOG_LEVEL`, and `ASCEND_MODULE_LOG_LEVEL`. For more information, see [Environment Variable List](<https://www.hiascend.com/document/detail/en/canncommercial/900/maintenref/envvar/envref_07_0001.html>) in the *CANN Environment Variable Reference*.
