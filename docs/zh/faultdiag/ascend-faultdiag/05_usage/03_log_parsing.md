# 日志清洗

日志清洗是从原始日志中提取关键信息的过程。清洗完成后，原始日志中的有效信息会被提取出来，供后续诊断使用。

## 使用方式

### 前置条件

1. 确保日志已按[日志采集](./02_log_collection.md)的要求收集完成。

2. 创建清洗输出目录：

    ```shell
    mkdir <output_dir>
    ```

    > - `output_dir` 清洗输出目录。

### 全量清洗（推荐）

需要将所有的模块日志都按照[采集日志归档目录结构](./02_log_collection.md#采集日志归档目录结构)的要求存放。

执行清洗命令：

```shell
ascend-fd parse -i <input_dir> -o <output_dir>
```

> - `input_dir` 采集目录。
> - `output_dir` 清洗输出目录。

如果需要同时清洗性能劣化（设备资源和网络拥塞）数据，添加 `-p` 参数：

```shell
ascend-fd parse -i <input_dir> -o <output_dir> -p
```

> - `input_dir` 采集目录。
> - `output_dir` 清洗输出目录。

回显如下表示清洗成功：

```text
The parse job starts. Please wait. Job id: [****], run log file is [****].
These job ['模块1', '模块2'...] succeeded.
The parse job is complete.
```

### 特定模块清洗

按模块日志类型分别指定输入目录：

```shell
ascend-fd parse \
    --host_log <主机侧操作系统日志目录> \
    --device_log <Device 侧日志目录> \
    --train_log <用户训练及推理日志目录> \
    --process_log <CANN 应用类日志目录> \
    --env_check <NPU 网口/状态信息/资源信息目录> \
    --dl_log <MindCluster 组件日志目录> \
    --mindie_log <MindIE 组件日志目录> \
    --amct_log <AMCT 组件日志目录> \
    --bus_log <LCNE 组件日志目录> \
    --pymotor_vllm_log <PyMotor/vLLM 日志目录> \
    --bmc_log <BMC 侧日志目录> \
    --lcne_log <LCNE 侧日志目录> \
    -o <清洗输出目录>
```

> [!NOTE]
>
> - 参数包含 --bus_log 命令时，传递的组件日志需为 LCNE 组件日志目录。
> - 说明：同时使用 `-i` 与详细日志目录参数时，会优先读取详细日志目录参数的值，再根据 `-i` 参数读取剩余日志目录。

### 清洗参数详细说明

清洗的详细参数说明，请阅读 [parse 详细参数说明](../06_api/02_command_parse.md#参数说明)

### 清洗输出结果

清洗输出结果请阅读 [parse 清洗输出结果说明](../06_api/02_command_parse.md#清洗输出结果)

## 多节点故障诊断

涉及多节点故障诊断，需要将各节点的清洗结果汇总到同一目录下。

每台服务器清洗完成后，需要将所有服务器的清洗结果汇总到同一台设备上，目录结构如下：

```text
诊断输入目录
    |-- 清洗输出目录 1（建议命名为节点标识，如 host1-192.168.1.1）
    |-- 清洗输出目录 2
    └── 清洗输出目录 N
```

## 注意事项

- 清洗输出目录的磁盘空间需大于 5GB，空间不足可能导致部分清洗结果丢失
- 在进行清洗时，请确保待清洗目录仅包含单台设备的日志
