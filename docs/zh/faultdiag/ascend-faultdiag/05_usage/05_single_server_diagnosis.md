# 单机故障诊断

单机故障诊断可以在单台设备上快速完成日志清洗和故障诊断，无需进行多节点日志转储。

适用于：快速排查单台设备的问题，不需要跨节点分析。

## 操作步骤

1. 先根据[日志采集指南](./02_log_collection.md)进行日志采集。

2. 创建单机诊断结果输出目录：

    ```shell
    mkdir <output_dir>
    ```

    > - `output_dir` 单机诊断结果输出目录。

3. 执行诊断命令：

    ```shell
    ascend-fd single-diag -i <input_dir> -o <output_dir>
    ```

    > [!NOTE]
    >
    > - `input_dir` 采集目录。
    > - `output_dir` 单机诊断结果输出目录。
    > - 单机诊断默认返回故障事件分析结果。

## 诊断报告与诊断结果

诊断报告解读请见[诊断报告解读](./04_fault_diagnosis.md#诊断报告解读)。

诊断结果文件请见[诊断结果文件](./04_fault_diagnosis.md#诊断结果文件)。

> [!NOTE]
>
> - 注意：单机诊断会扫描节点中所有有效日志的故障事件。如果诊断执行出错，可以通过 `diag_report.json` 文件查看所有异常信息。
