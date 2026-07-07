# Diagnosing Faults on a Single Server<a name="ZH-CN_TOPIC_0000002072345910"></a>

1. Create a directory for storing the single-server diagnosis result.

    ```shell
    mkdir *Directory for storing the single-server diagnosis result*
    ```

2. Start diagnosis.

    By default, the data of the fault event module is returned for single-server diagnosis.

    ```shell
    ascend-fd single-diag -i Collection directory -o Directory for storing the single-server diagnosis result
    ```

    The following information is displayed when a training job exits abnormally during diagnosis:

    ```ColdFusion
    The single-diag job starts. Please wait. Job id: [****], run log file is [****].
    +------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |                                                                                       Ascend Fault-Diag Report                                                                                       |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    | Version information   | Type | Version                                                                                                                                                                      |
    +--------------+------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.0.0                                                                                                                                                                    |
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

    The following table describes key parameters in the command output.

    **Table 1** Key parameters

    |Level-1 Parameter|Level-2 Parameter|Description|
    |--|--|--|
    |Fault event analysis|-|Used to analyze the root cause of the device where the root cause node is located.|
    |-|Status code|<ul><li>If a fault is diagnosed, a specific fault code is displayed. </li><li>If no fault is diagnosed, `NORMAL OR UNSUPPORTED` is displayed.</li></ul>|
    |-|Fault name|Specific fault name.|
    |-|Fault type|Type of a fault and the component and module where the fault occurs.|
    |-|Faulty device|Device where a fault occurs.|
    |-|Fault description|Detailed description of a fault.|
    |-|Suggestion|Handling suggestions for a fault.|
    |-|Key log|Key log of a fault.|
    |-|Key propagation chain|Used to display the longest fault link.|

    >[!NOTE]
    >During single-server diagnosis, fault events in all valid logs on the node are scanned. If results of fault event analysis are displayed in the command output, the current fault may cause the training or inference job to exit abnormally.

    After the diagnosis is complete, you can perform optimization based on the recommended solution in the single-server diagnosis result.

    ```text
    Directory of the single-server diagnosis result
    ├── fault_diag_result
        ├── diag_report.json    # Diagnosis result
    ```

    >[!NOTE]
    >If an error occurs during single-server fault diagnosis, the "fault description" (or "analysis failure analysis") field under "fault event analysis" will display the failure information. To view all exception information, check the `diag_report.json` file.
