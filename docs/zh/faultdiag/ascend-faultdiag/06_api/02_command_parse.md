# parse 命令（日志清洗）

## 功能说明

对原始日志进行清洗，提取有效信息。

## 命令格式

```shell
ascend-fd parse [-h] [-i INPUT_PATH] -o OUTPUT_PATH \
    [--host_log HOST_LOG] [--device_log DEVICE_LOG] \
    [--train_log TRAIN_LOG [TRAIN_LOG ...]] [--process_log PROCESS_LOG] \
    [--env_check ENV_CHECK] [--dl_log DL_LOG] [--mindie_log MINDIE_LOG] \
    [--amct_log AMCT_LOG] [--bmc_log BMC_LOG] [--lcne_log LCNE_LOG] \
    [--bus_log BUS_LOG] [--pymotor_vllm_log PYMOTOR_VLLM_LOG] \
    [--custom_log CUSTOM_LOG] [-p]
```

## 参数说明

| 参数               | 类型   | 必选 | 说明                                             |
|--------------------|--------|------|--------------------------------------------------|
| -h, --help         | -      | 否   | 显示帮助信息                                     |
| -i, --input_path   | string | 否   | 预处理数据输入路径                               |
| -o, --output_path  | string | 是   | 清洗结果输出路径                                 |
| --host_log         | string | 否   | 主机侧操作系统日志目录                           |
| --device_log       | string | 否   | Device 侧日志目录                                |
| --train_log        | string | 否   | 用户训练及推理日志目录，最多 20 个               |
| --process_log      | string | 否   | CANN 应用类日志目录                              |
| --env_check        | string | 否   | NPU 网口、状态信息、资源信息目录                 |
| --dl_log           | string | 否   | MindCluster 组件日志目录                         |
| --mindie_log       | string | 否   | MindIE 组件日志目录                              |
| --amct_log         | string | 否   | AMCT 组件日志目录                                |
| --bmc_log          | string | 否   | BMC 组件日志目录                                 |
| --lcne_log         | string | 否   | LCNE 组件日志目录                                |
| --bus_log          | string | 否   | Ascend 950 系列 LCNE 组件日志目录                |
| --pymotor_vllm_log | string | 否   | PyMotor/vLLM 日志目录                            |
| --custom_log       | string | 否   | 自定义解析文件目录                               |
| -p, --performance  | -      | 否   | 清洗设备资源、网络拥塞两个性能劣化检测模块的数据 |

## 使用示例

### 基础清洗

```shell
ascend-fd parse -i /tmp/log_dir -o /tmp/parse_out
```

### 含性能劣化数据的清洗

```shell
ascend-fd parse -i /tmp/log_dir -o /tmp/parse_out -p
```

### 特定组件日志清洗

```shell
ascend-fd parse --process_log /tmp/cann_log --train_log /tmp/train_log -o /tmp/parse_out
```

## 清洗输出结果

清洗输出目录结构：

```text
└── 清洗输出目录
    ├── ascend-kg-parser.json
    ├── ascend-kg-analyzer.json
    ├── ascend-rc-parser.json
    ├── device_ip_info.json
    ├── mindie-cluster-info.json
    ├── server-info.json
    ├── nad_clean.csv
    ├── nic_clean.csv
    ├── process_{core_num}.csv
    ├── plog-parser-{pid}-{0/1}.log
    ...
    └── plog-parser-{pid}-{0/1}.log
```

### 清洗结果说明

| 文件                          | 说明                                                                  |
|-------------------------------|-----------------------------------------------------------------------|
| `ascend-kg-parser.json`       | 故障事件分析清洗结果。（旧版本文件，兼容 ascend-fd 6.0.0 及之前版本） |
| `ascend-kg-analyzer.json`     | 故障事件分析清洗结果。（新版本文件，ascend-fd 6.0.0 之后版本）        |
| `ascend-rc-parser.json`       | 根因节点分析清洗结果                                                  |
| `device_ip_info.json`         | 设备 IP 信息                                                          |
| `plog-parser-{pid}-{0/1}.log` | 根因节点分析清洗后日志，按 PID 分类保存                               |
| `mindie-cluster-info.json`    | MindIE Pod 日志清洗结果                                               |
| `server-info.json`            | MindIE 组件服务器信息                                                 |
| `nad_clean.csv`               | 计算降频清洗结果（需 `-p` 参数）                                      |
| `nic_clean.csv`               | 网络拥塞清洗结果（需 `-p` 参数）                                      |
| `process_{core_num}.csv`      | CPU 资源抢占清洗结果（需 `-p` 参数）                                  |

## 注意事项

- 清洗前，请确保输出目录有 5GB 以上的可用磁盘空间
- ascend-fd 运行错误码请查阅[组件错误码](../07_references/04_appendix.md#组件错误码)
