# 快速入门

本文将指导完成 ascend-fd-tk 首次诊断操作，基于离线交换机日志演示链路故障诊断功能。

## 前置条件

- 在 Linux 操作系统上确定已安装 Python 3.8 及以上版本、对应的 pip3 版本和 unzip 解压工具
- 确认网络连接正常，安装过程需联网下载三方依赖库

## 步骤 1：安装工具

1. 下载链路诊断安装包：

    ```bash
    # 下载 v26.1.0 版本的 故障诊断 ZIP 压缩包（因为 ascend-fd-tk WHL 安装包不区分架构，所以以下示例直接下载 aarch64 架构的）。
    wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.1.0/Ascend-mindxdl-faultdiag_26.1.0_linux-aarch64.zip
    unzip Ascend-mindxdl-faultdiag_26.1.0_linux-aarch64.zip
    ```

2. 安装 WHL 包：

    ```bash
    pip3 install ascend_faultdiag_toolkit-26.1.0-py3-none-any.whl
    ```

    成功回显示例：

    ```txt
    Successfully installed ascend-faultdiag-toolkit-26.1.0
    ```

3. 验证安装是否成功：

    ```bash
    ascend-fd-tk about
    ```

    成功回显示例：

    ```text
    MindCluster ascend-faultdiag-toolkit诊断工具版本：26.1.0
    ```

## 步骤 2：清理缓存

首次使用或重新诊断前，建议清理缓存以避免上次诊断结果影响本次诊断：

```bash
ascend-fd-tk clear_cache
```

成功回显示例：

```text
清理完成
```

## 步骤 3：配置数据源（离线模式）并一键诊断

以交换机离线日志为例，将[示例日志](../../../resource/switch_logs)放到服务器目录（如 temp 目录）。工具支持自动解压压缩包，无需提前解压。

1. 获取交换机示例离线日志：

    ```bash
    mkdir -p /temp/switch_logs && cd /temp/switch_logs
    wget \
    https://raw.gitcode.com/Ascend/mind-cluster/blobs/19ab0e6d1acc5b64e5153d479f632302e9d82827/switch_logs/diagnostic_information_NAME-D01-XX.224_20260327113617.zip \
    https://raw.gitcode.com/Ascend/mind-cluster/blobs/afd88c0586ad298f1eb46f3f515e82f9b5dd47ab/switch_logs/diagnostic_information_NAME-D01-XX.254_20260327113617.zip
    ```

2. 执行一键式诊断命令，工具将自动完成日志清洗与故障诊断：

    ```bash
    cd /temp/
    ascend-fd-tk set_switch_dump_log /temp/switch_logs auto_collect_diag
    ```

    成功回显示例：

    ```text
    设置成功
    ...
    诊断完成
    ```

## 步骤 4：查看报告

诊断完成后，报告自动生成至目录：`~/.ascend-faultdiag-toolkit/report/diag_report_{YYYYMMDD_HHMMSS}.xlsx`。报告字段含义与解读方法详见[诊断 / 巡检报告说明](../05_usage/06_fault_analysis_report.md)。

诊断报告示例展示：

图1 交换机故障分析报告示例

![交换机故障分析](../../figures/ascend-faultdiag-toolkit/交换机故障分析-case.png)

图2 交换机间端口连接光模块信息报告示例

![交换机间端口连接光模块信息](../../figures/ascend-faultdiag-toolkit/交换机间端口连接光模块信息-case.png)

## 下一步

- 详细安装说明可以参考[安装说明](../04_installation_guide/01_installation.md)。
- 参考[特性概览](../05_usage/01_usage_overview.md)了解更多的功能。
- 参考[API 概述](../06_api/01_api_overview.md)了解更多的命令。
