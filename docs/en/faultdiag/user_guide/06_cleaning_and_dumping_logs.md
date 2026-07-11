# Cleaning and Dumping Logs<a name="ZH-CN_TOPIC_0000001541788946"></a>

>[!NOTE]
>
>- The drive space of the output directory specified by the cleaning command must be greater than 5 GB. If the drive space is insufficient, some cleaning results may be lost, causing abnormal or inaccurate diagnosis results.
>- When cleaning logs, MindCluster Ascend FaultDiag reads the log files and monitoring metric files collected by users. Ensure that these files do not contain sensitive information to prevent information leakage.
>- During cleaning, ensure that the directory to be cleaned contains only the original logs and monitoring metric files of a single training device. If the directory contains files related to other devices, cleaning may fail.
>- For cleaning data from the device resource and network congestion performance degradation detection modules, specify the `--performance(-p)` parameter in the cleaning command. If this parameter is not specified, the program cleans only the data of the root cause node and fault event module by default.

1. (Optional) Install the component as the `root` user. To use the component as a common user, configure environment variables. If no dependency can be found, check whether the dependency has been installed or whether the permission is correct.
    1. Log in as the `root` user and query the component location.

        ```shell
        which ascend-fd
        ```

        The following information is displayed. The actual location is subject to the query result.

        ```ColdFusion
        /usr/local/python3.7.5/bin/ascend-fd
        ```

    2. Log in as a common user and configure environment variables.

        ```shell
        export PATH=$PATH:/usr/local/python3.7.5/bin
        ```

    3. Run the command to check whether the configuration is complete.

        ```shell
        ascend-fd -h
        ```

        If the following information is displayed, the configuration is complete:

        ```ColdFusion
        usage: ascend-fd [-h] {version,parse,diag,blacklist,config,entity,single-diag} ...
        Ascend Fault Diag
        positional arguments:
          {version,parse,diag,blacklist,config,entity,single-diag}
            version             show ascend-fd version
            parse               parse origin log files
            diag                diag parsed log files
            blacklist           filter invalid CANN logs by blacklist for parsing
            config              custom configuration parsing files
            entity              perform operations on the user-defined faulty entity.
            single-diag         single parse and diag log files
        optional arguments:
          -h, --help            show this help message and exit
        ```

2. Collect logs of the training device by referring to [Collecting Logs](./03_collecting_logs.md).

    Upload the logs to any directory (for example, `/home`) on the server. For example, if the `-i` parameter is used, all logs are collected to the same collection directory for cleaning. The directory structure is as follows:

    - Host

        ```text
        Collection directory
        |-- messages        # Host OS logs
        |-- dmesg                # Host kernel message logs
        |-- crash
            |-- Host + Fault time (eg:127.xx.xx.1-2024-09-23-11:25:29)
                |-- vmcore_dmesg.txt     # Host kernel message log saved when the system breaks down
        |-- sysmonitor.log       # System monitoring log
        |-- rank-0.txt     # Training console logs
        ...
        |-- rank-7.txt     # Training console logs
        |-- process_log          # Original App logs of CANN. The directory name must be process_log.
        |-- device_log           # Device logs, which must be stored in the device_log directory.
        |-- dl_log                # MindCluster component logs. The directory name must be dl_log.
            |-- devicePlugin # Ascend Device Plugin logs
            |-- noded               # NodeD logs
            |-- ascend-docker-runtime              #  Ascend Docker Runtime logs
            |-- volcano-scheduler         # volcano-scheduler logs
            |-- volcano-controller            # volcano-controller logs

            |-- npu-exporter              # NPU Exporter logs
            |-- ttp_log                   # MindIO component logs
        |-- mindie               # MindIE component logs
            |-- log
                |-- debug        # MindIE run logs
                |-- security     # MindIE audit logs
                |-- mindie_cluster_log     # MindIE Pod console logs
        |-- amct_log             # AMCT logs
        |-- bus_log              # LCNE logs (Ascend 950)
        |-- environment_check # Information about the NPU network port, status, and resource
            |-- npu_smi_0_details.csv   # NPU status monitoring metrics
             ...
            |-- npu_smi_7_details.csv   # NPU status monitoring metrics
            |-- npu_0_details.csv         # NPU network port monitoring metrics
             ...
            |-- npu_7_details.csv       # NPU network port monitoring metrics
            |-- npu_info_before/after.txt  # NPU network port information before or after training
            |-- host_metrics_{core_num}.json  # Host resource monitoring metrics
        ```

    - BMC and LCNE:

        Decompress the BMC and LCNE logs exported from Computing ToolKit or CCAE recursively, and then place and clean the logs on a single node.

        ```shell
        ascend-fd parse --lcne_log ${Decompressed_LCNE_log_directory_of_a_single_node} -o ${Cleaning_result_directory}
        ascend-fd parse --bmc_log ${Decompressed_BMC_log_directory_of_a_single_node} -o ${Cleaning_result_directory}
        ```

        >[!NOTE]
        >- For details about how to use CCAE to collect logs, see [LingQu Log Collection](https://support.huawei.com/hedex/hdx.do?docid=EDOC1100499782&id=EN-US_TOPIC_0000002245284913).
        >- For details about how to use Computing ToolKit to collect logs, see "Using Computing ToolKit" > "Log Collect" > "Usage Guide" > "Collecting BMC, IES, and Switch Logs" in the [*Computing ToolKit User Guide*](https://support.huawei.com/carrier/productNewOffering?col=product&path=PBI1-262732867/PBI1-262735884/PBI1-261914673/PBI1-264314551).

3. Create an output directory for storing log cleaning results.

    ```shell
    mkdir ${Cleaning_result_output_directory}
    ```

4. Run the command to start cleaning logs.

    ```shell
    ascend-fd parse -i ${Collection_directory} -o ${Cleaning_result_output_directory} --performance
    ```

    Command output:

    ```ColdFusion
    The parse job starts. Please wait. Job id: [****], run log file is [****].
    These job ['Module 1', 'Module 2',...] succeeded.
    The parse job is complete.
    ```

    Structure of the cleaning result directory:

    ```text
    └── Cleaning result directory
       ├── ascend-kg-parser.json       #Result of fault event analysis cleaning, which is the input file of the inference engine
       ├── ascend-kg-analyzer.json      # Result of fault event analysis cleaning
       ├── ascend-rc-parser.json       # Result of root cause analysis cleaning
       ├── device_ip_info.json          # Device IP address information
       ├── mindie-cluster-info.json      # Result of MindIE Pod console log cleaning
       ├── server-info.json             # Result of MindIE component log cleaning
       ├── nad_clean.csv                # Result of computing throttling cleaning
       ├── nic_clean.csv               # Result of network congestion cleaning
       ├── process_{core_num}.csv       # Result of CPU resource preemption cleaning
       ├── plog-parser-{pid}-{0/1}.log  # Logs generated after root cause analysis cleaning, including key information such as error and trace logs. The logs are saved by PID.
        ...
       └── plog-parser-{pid}-{0/1}.log
    ```

5. Dump logs.

    Dump all files in the cleaning output directory of each server in a centralized manner. The dump directory structure is as follows:

    ```text
    Diagnosis input directory
        |-- Cleaning result output directory 1
           |--plog-parser-{pid}-{0/1}.log     # Logs of the cleaned root cause nodes, including key information such as error and trace logs. The logs are saved by PID.
           |--nic_clean.csv                      # Result of network congestion cleaning
           |--nad_clean.csv                      # Result of computing throttling cleaning
           |--mem_used.csv                     # Result of memory resource preemption cleaning. This file is reserved and is not used currently.
           |--process_{core_num}.csv           # Result of CPU resource preemption cleaning
           |--device_ip_info.json               # Device IP address information
           |--ascend-kg-parser.json # Result of fault event analysis cleaning, which is the input file of the inference engine
           |--ascend-kg-analyzer.json           # Result of fault event analysis cleaning
           |--ascend-rc-parser.json           # Result of root cause node analysis cleaning
           |--mindie-cluster-info.json           # Result of MindIE Pod console log cleaning
           |--server-info.json                   # Result of MindIE component log cleaning

        |-- Cleaning result output directory 2
           |--plog-parser-{pid}-{0/1}.log
           |--nic_clean.csv
           |--nad_clean.csv
           |--mem_used.csv
           |--process_{core_num}.csv
           |--device_ip_info.json
           |--ascend-kg-parser.json
           |--ascend-kg-analyzer.json
           |--ascend-rc-parser.json
           |--server-info.json                   ...
        |-- Cleaning result output directory n
    ```

>[!NOTE]
>
>- You are advised to change the name of the cleaning result output directory to a directory name that can identify device node information, for example, `host1-192.168.x.x`.
>- Store the MindIE Pod console log cleaning result only on one node.
