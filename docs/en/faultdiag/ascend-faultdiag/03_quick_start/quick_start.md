# Quick Start

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:26:40.617Z pushedAt=2026-08-24T02:51:43.232Z -->

This chapter helps users install ascend-fd and run a basic diagnostic example, using sample logs to experience how ascend-fd diagnoses the issue of an NPU optical module not in position.

## Prerequisites

- Linux system with the unzip tool installed.
- Python 3.7 or later installed.
- pip3 installed.
- Ensure normal network connectivity, as the installation process requires downloading third-party dependency libraries.

## Step 1: Install ascend-fd

1. Obtain the software package.

    Use the `arch` command to obtain the current environment architecture, and then use the following command to automatically download the software package from the open-source community.

    - aarch64

        ```shell
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.1.0/Ascend-mindxdl-faultdiag_26.1.0_linux-aarch64.zip
        ```

    - x86_64

        ```shell
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.1.0/Ascend-mindxdl-faultdiag_26.1.0_linux-x86_64.zip
        ```

2. Decompress and install the package.

    - aarch64

        ```shell
        unzip Ascend-mindxdl-faultdiag_26.1.0_linux-aarch64.zip
        pip3 install ascend_faultdiag-26.1.0-py3-none-linux_aarch64.whl
        ```

    - x86_64

        ```shell
        unzip Ascend-mindxdl-faultdiag_26.1.0_linux-x86_64.zip
        pip3 install ascend_faultdiag-26.1.0-py3-none-linux_x86_64.whl
        ```

3. Verify whether the installation is successful.

    ```shell
    ascend-fd version
    ```

    If the version number is displayed, the installation is successful.

    ```shell
    ascend-fd v26.1.0
    ```

## Step 2: Prepare Logs

This example only requires preparing the environment check logs.

1. Create a collection directory.

    ```shell
    mkdir -p /tmp/faultdiag_demo/log_dir
    ```

2. Place the log files into the collection directory.

    - Obtain the sample logs using the following command. These logs are the corresponding environment check logs collected before and after training. For details, see [Log Collection](../05_usage/02_log_collection.md).

        ```shell
        wget https://raw.gitcode.com/Ascend/mind-cluster/blobs/d58f9ef2e4c1930ba720353d452ebdcdb3ee2aad/environment_check.zip
        ```

    - Extract logs to the `/tmp/faultdiag_demo/log_dir` directory.

        ```shell
        unzip environment_check.zip -d /tmp/faultdiag_demo/log_dir
        ```

## Step 3: Parse Logs

1. Create the parsing output directory.

    ```shell
    mkdir -p /tmp/faultdiag_demo/parse_out
    ```

2. Run the parsing command.

    ```shell
    ascend-fd parse -i /tmp/faultdiag_demo/log_dir -o /tmp/faultdiag_demo/parse_out
    ```

    If the output is similar to the following, the parsing is successful:

    ```text
    The parse job starts. Please wait. Job id: [20260701031834593100_c414615b-0550-467f-b84b-24791115befa], run log file is [ascend_faultdiag_815671.log].
    These job ['KNOWLEDGE_GRAPH'] succeeded.
    Warn: The job ROOT_CLUSTER failed. The error is: [FileNotExistError(502): No plog file that meets the path specifications is found.].
    The parse job is complete.
    ```

    > [!NOTE]
    >
    > - For details about the command, see [parse Command (log parsing)](../06_api/02_command_parse.md).
    > - `ROOT_CLUSTER failed` occurs because there is no plog (CANN app log) file, so the root cause data cannot be parsed. This warning can be ignored.

## Step 4: Fault Diagnosis

1. Create the diagnosis output directory.

    ```shell
    mkdir -p /tmp/faultdiag_demo/diag_out
    ```

2. Run the diagnosis command.

    ```shell
    ascend-fd diag -i /tmp/faultdiag_demo -o /tmp/faultdiag_demo/diag_out
    ```

    > [!NOTE]
    >
    > For the specific command, see [diag Command (cluster fault diagnosis)](../06_api/03_command_diag.md).

    After the diagnosis is complete, the terminal outputs the diagnosis report:

    <!-- markdownlint-disable-next-line MD033 -->
    <pre>
    The diag job starts. Please wait. Job id: [20260701033838010337_09257e70-b5e9-452d-9bb6-bb32aa32507c], run log file is [ascend_faultdiag_848424.log].
    +---------------------------------------------------------------------------------------------------------------------------------------+
    |                                                      Ascend-fd Fault-Diag Report                                                      |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    |   Version Information  |    Type    | Version                                                                                                      |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.1.0                                                                                                    |
    |              |   Driver   | 23.0.7                                                                                                    |
    |              |  Firmware  | 7.1.0.11.220                                                                                              |
    |              |    NNAE    | 8.0.0                                                                                                     |
    |              |  Toolkit   | 8.0.RC3                                                                                                   |
    |              |  PyTorch   | 1.13                                                                                                      |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    | Root Cause Node Analysis |    Type    | Description                                                                                                      |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    |              |    Description    | No root cause node was diagnosed. The fault event analysis will attempt to detect all devices.                                                            |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    |              |  Root Cause Node  | ['Unknown Device']                                                                                        |
    |              |  Symptom  | No valid Plog file found, root cause node cannot be located. Please confirm whether a Plog file exists.                                        |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    | Fault Event Analysis |    Type    | Description                                                                                                      |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    |              |    Description   | 1. Some analysis sub-items under this module failed to execute, which may affect the accuracy of the diagnostic results. Failure details can be found in diag_report.json. |
    |              |            | 2. Only the longest propagation chain for each faulty device is displayed.                                                             |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    | Suspected Fault |   Status Code   | Comp_Network_Custom_05                                                                                    |
    |              |  Fault Category  | Category: Network Component: Network Module: Network                                                                    |
    |              |  Fault Device  | ['parse_out device-0', 'parse_out device-4']                                                              |
    |              |  Fault Name  | NPU optical module is not present                                                                                           |
    |              |  Fault Description  | NPU optical module not present detected.                                                                                   |
    |              |  Suggestion  | 1. Use the msnpureport tool to collect NPU logs and contact Huawei engineer.                                               |
    |              |  Key Log  | /usr/local/Ascend/driver/tools/hccn_tool -i 0 -optical -g                                                 |
    |              |            | present              : not present                                                                        |
    |              | Key Propagation Chain | ['parse_out device-4']                                                                                    |
    |              |            | Comp_Network_Custom_05 (NPU optical module not present) -> Comp_Network_Custom_09 (Optical module RX/TX no light emission/reception)                 |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    The diag job is complete.
    </pre>

## Result Interpretation

- The versions of related software are read from the logs and displayed in the diagnostic report.
- If no CANN logs are included in the input, the diagnostic result indicates that no plog file is found.
- The root cause node displays `Unknown Device` because log collection is incomplete. This is normal.
- Based on the environment check logs, the NPU optical modules on device-0 and device-4 are detected as not present.
- For related status codes, see [Supported Faults](../07_references/04_appendix.md#known-faults).
- For the detailed report, see `/tmp/faultdiag_demo/diag_out/fault_diag_result/diag_report.json`.

## Next Steps

- Refer to [Feature Guide](../05_usage/menu_usage.md) to learn more about features.
- Refer to [API Reference](../06_api/menu_api.md) to learn more about commands.
