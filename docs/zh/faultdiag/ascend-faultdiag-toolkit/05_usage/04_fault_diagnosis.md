# 故障诊断

ascend-fd-tk 工具完成 [日志采集](02_log_collection.md) 与 [日志清洗](03_log_parse.md)后，使用诊断功能对清洗后的结构化数据进行故障分析。

## 诊断命令

工具提供两种诊断方式：

| 命令                                                    | 说明                                                     |
|-------------------------------------------------------|--------------------------------------------------------|
| [auto_collect_diag](../06_api/04_parse_diagnosis/03_auto_collect_diag.md) | 一键式诊断。整合采集、清洗、诊断三个步骤，等价于 `auto_collect` + `auto_diag`。适用于一次完成端到端诊断的场景 |
| [auto_diag](../06_api/04_parse_diagnosis/02_auto_diag.md)            | 仅诊断（需先完成采集）                                            |

## 使用 `auto_collect_diag` 命令

### 在线场景

非交互式方式（展示命令与回显）：

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect_diag
诊断完成
```

交互式方式（展示命令与回显）：

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_conn_config /path/to/conn.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> auto_collect_diag
诊断完成
```

### 离线场景

非交互式方式（展示命令与回显）：

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect_diag
诊断完成
```

交互式方式（展示命令与回显）：

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_host_dump_log /path/to/host_logs
设置成功
>>> set_bmc_dump_log /path/to/bmc_logs
设置成功
>>> set_switch_dump_log /path/to/switch_logs
设置成功
>>> auto_collect_diag
诊断完成
```

## 使用 `auto_diag` 命令

### 在线场景

非交互式方式（展示命令与回显）：

```bash
# 第一步：采集与清洗
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect
收集完成，若完成全部收集请进行诊断/巡检

# 第二步：执行诊断
ascend-fd-tk auto_diag
诊断完成
```

交互式方式（展示命令与回显）：

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_conn_config /path/to/conn.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> auto_collect
收集完成，若完成全部收集请进行诊断/巡检
>>> auto_diag
诊断完成
```

### 离线场景

非交互式方式（展示命令与回显）：

```bash
# 第一步：采集与清洗
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect
收集完成，若完成全部收集请进行诊断/巡检

# 第二步：执行诊断
ascend-fd-tk auto_diag
诊断完成
```

交互式方式（展示命令与回显）：

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_host_dump_log /path/to/host_logs
设置成功
>>> set_bmc_dump_log /path/to/bmc_logs
设置成功
>>> set_switch_dump_log /path/to/switch_logs
设置成功
>>> auto_collect
收集完成，若完成全部收集请进行诊断/巡检
>>> auto_diag
诊断完成
```

## 查看诊断报告

诊断完成后报告自动生成至工具家目录下的 report 子目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。
