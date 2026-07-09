# 快速入门

本章节旨在帮助用户完成 ascend-fd 的安装并运行一个基础诊断示例，使用示例日志体验 ascend-fd 诊断出 NPU 光模块不在位的问题。

## 前提条件

- Linux 系统，已安装 unzip 工具
- 已安装 Python 3.8 及以上版本
- 已安装 pip3

## 步骤1：安装 ascend-fd

1. 从开源社区获取软件包 `Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip`，[下载链接](https://gitcode.com/Ascend/mind-cluster/releases)。

    > - `{version}` 为软件包版本号。
    > - `{arch}` 为软件包架构，分为 x86_64 和 aarch64，请根据实际需要修改，可通过 `arch` 命令查看。

2. 将软件包上传到 linux 机器。

3. 在 linux 机器上解压并安装：

    ```shell
    unzip Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip
    pip3 install ascend_faultdiag-{version}-py3-none-linux_{arch}.whl
    ```

4. 验证安装是否成功：

    ```shell
    ascend-fd version
    ```

    如果回显版本号，说明安装成功，如:

    ```shell
    ascend-fd v26.1.0
    ```

## 步骤2：准备日志

本示例使用**基础应用场景**，只需要准备环境检查日志。

1. 创建采集目录：

    ```shell
    mkdir -p /tmp/faultdiag_demo/log_dir
    ```

2. 将日志文件放入采集目录：

    - 获取示例日志文件（[点击下载](../../../resource/environment_check.zip)），该日志为训练前后收集的对应环境检查日志，具体请参考[日志采集](../05_usage/02_log_collection.md)。
    - 解压到 `/tmp/faultdiag_demo/log_dir` 目录

    ```shell
    unzip environment_check.zip -d /tmp/faultdiag_demo/log_dir
    ```

## 步骤3：日志清洗

1. 创建清洗输出目录：

    ```shell
    mkdir -p /tmp/faultdiag_demo/parse_out
    ```

2. 执行清洗命令：

    ```shell
    ascend-fd parse -i /tmp/faultdiag_demo/log_dir -o /tmp/faultdiag_demo/parse_out
    ```

    如果回显类似以下内容，说明清洗成功：

    ```text
    The parse job starts. Please wait. Job id: [20260701031834593100_c414615b-0550-467f-b84b-24791115befa], run log file is [ascend_faultdiag_815671.log].
    These job ['KNOWLEDGE_GRAPH'] succeeded.
    Warn: The job ROOT_CLUSTER failed. The error is: [FileNotExistError(502): No plog file that meets the path specifications is found.].
    The parse job is complete.
    ```

    > [!NOTE]
    >
    > - 具体的命令可以查询[API参考 -> parse命令（日志清洗）](../06_api/02_command_parse.md)。
    > - `ROOT_CLUSTER failed` 由于无 Plog（CANN 应用类日志） 文件，根因数据不能清洗，该告警可以忽略。

## 步骤4：故障诊断

1. 创建诊断输出目录：

    ```shell
    mkdir -p /tmp/faultdiag_demo/diag_out
    ```

2. 执行诊断命令：

    ```shell
    ascend-fd diag -i /tmp/faultdiag_demo -o /tmp/faultdiag_demo/diag_out
    ```

    > [!NOTE]
    >
    > - 具体的命令可以查询[API参考 -> diag命令（集群故障诊断）](../06_api/03_command_diag.md)。

    诊断完成后，终端会输出诊断报告：

    <!-- markdownlint-disable-next-line MD033 -->
    <pre>
    The diag job starts. Please wait. Job id: [20260701033838010337_09257e70-b5e9-452d-9bb6-bb32aa32507c], run log file is [ascend_faultdiag_848424.log].
    +---------------------------------------------------------------------------------------------------------------------------------------+
    |                                                      Ascend-fd Fault-Diag Report                                                      |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    |   版本信息   |    类型    | 版本                                                                                                      |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    |              | Fault-Diag | 26.1.0                                                                                                    |
    |              |   Driver   | 23.0.7                                                                                                    |
    |              |  Firmware  | 7.1.0.11.220                                                                                              |
    |              |    NNAE    | 8.0.0                                                                                                     |
    |              |  Toolkit   | 8.0.RC3                                                                                                   |
    |              |  PyTorch   | 1.13                                                                                                      |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    | 根因节点分析 |    类型    | 描述                                                                                                      |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    |              |    说明    | 未诊断出根因节点，故障事件分析将尝试检测全部设备                                                          |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    |              |  根因节点  | ['Unknown Device']                                                                                        |
    |              |  现象描述  | 未查找到有效的Plog文件，无法定位根因节点。请确认是否存在Plog文件？                                        |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    | 故障事件分析 |    类型    | 描述                                                                                                      |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    |              |    说明    | 1. 本分析模块下部分分析子项执行失败，诊断结果可能会受到影响从而不准确。失败信息可在diag_report.json中查询 |
    |              |            | 2. 关键传播链只展示每个故障设备最长的一条链路                                                             |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    | 疑似根因故障 |   状态码   | Comp_Network_Custom_05                                                                                    |
    |              |  故障分类  | 类型:Network 组件:Network 模块:Network                                                                    |
    |              |  故障设备  | ['parse_out device-0', 'parse_out device-4']                                                              |
    |              |  故障名称  | NPU光模块不在位                                                                                           |
    |              |  故障描述  | 检测到NPU光模块不在位。                                                                                   |
    |              |  建议方案  | 1. 建议使用msnpureport工具收集NPU日志，联系华为工程师处理；                                               |
    |              |  关键日志  | /usr/local/Ascend/driver/tools/hccn_tool -i 0 -optical -g                                                 |
    |              |            | present              : not present                                                                        |
    |              | 关键传播链 | ['parse_out device-4']                                                                                    |
    |              |            | Comp_Network_Custom_05（NPU光模块不在位）-> Comp_Network_Custom_09（光模块RX/TX无收发光）                 |
    +--------------+------------+-----------------------------------------------------------------------------------------------------------+
    The diag job is complete.
    </pre>

## 结果解读

- 从日志读取到相关软件的版本，并在诊断报告中展示。
- 日志中没有 CANN 日志，诊断结果中提示没有 Plog 文件。
- 诊断结果显示 `Unknown Device`，是由于日志采集不完整，此处是正常情况。
- 根据环境检查日志，检测出 device-0 和 device-4 上的 NPU 光模块不在位。
- 相关状态码可以参考[已支持故障](../07_references/04_appendix.md#已支持故障)。
- 详细报告可以查看 `/tmp/faultdiag_demo/diag_out/fault_diag_result/diag_report.json`。

## 下一步

- 参考[特性指南](../05_usage/menu_usage.md)了解更多功能。
- 参考[API参考](../06_api/menu_api.md)了解更多命令。
