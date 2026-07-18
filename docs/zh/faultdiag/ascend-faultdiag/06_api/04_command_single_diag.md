# single-diag 命令（单机故障诊断）

## 功能说明

用于在单台设备上直接完成日志清洗和故障诊断，无需进行多节点日志转储。

## 命令格式

```shell
ascend-fd single-diag [-h] [-i INPUT_PATH] -o OUTPUT_PATH \
    [--host_log HOST_LOG] [--device_log DEVICE_LOG] \
    [--train_log TRAIN_LOG [TRAIN_LOG ...]] [--process_log PROCESS_LOG] \
    [--env_check ENV_CHECK] [--dl_log DL_LOG] [--mindie_log MINDIE_LOG] \
    [--amct_log AMCT_LOG] [--bus_log BUS_LOG] \
    [--pymotor_vllm_log PYMOTOR_VLLM_LOG]
```

## 参数说明

| 参数               | 类型                | 必选 | 说明                                     |
|--------------------|---------------------|------|------------------------------------------|
| -h, --help         | -                   | 否   | 显示帮助信息                             |
| -i, --input_path   | String              | 否   | 预处理数据输入路径                       |
| -o, --output_path  | String              | 是   | 诊断结果输出路径                         |
| --host_log         | String              | 否   | 主机侧操作系统日志目录                   |
| --device_log       | String              | 否   | Device 侧日志目录                        |
| --train_log        | String/List[String] | 否   | 用户训练及推理日志目录，支持传递多个目录 |
| --process_log      | String              | 否   | CANN 应用类日志目录                      |
| --env_check        | String              | 否   | NPU 网口、状态信息、资源信息目录         |
| --dl_log           | String              | 否   | MindCluster 组件日志目录                 |
| --mindie_log       | String              | 否   | MindIE 组件日志目录                      |
| --amct_log         | String              | 否   | AMCT 组件日志目录                        |
| --bus_log          | String              | 否   | LCNE 组件日志目录                        |
| --pymotor_vllm_log | String              | 否   | PyMotor/vLLM 日志目录                    |

## 使用示例

### 执行诊断命令

```shell
ascend-fd single-diag -i /tmp/log_dir -o /tmp/diag_out
```

### 分类输入日志目录诊断命令

```shell
ascend-fd single-diag --process_log {采集目录}/process_log -o /tmp/diag_out
```

## 注意事项

- 单机诊断默认返回故障事件分析结果。
- 如果诊断出故障，状态码为具体故障码；未诊断出故障时，状态码为 `NORMAL_OR_UNSUPPORTED`。
- 单机诊断会扫描节点中所有有效日志的故障事件。
- ascend-fd 运行错误码请查阅[组件错误码](../07_references/04_appendix.md#组件错误码)。
- 单机诊断结果可参考 [基础诊断](03_command_diag.md#基础诊断)。
