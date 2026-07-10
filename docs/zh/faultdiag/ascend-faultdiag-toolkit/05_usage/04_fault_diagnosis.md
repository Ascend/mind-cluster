# 故障诊断

ascend-fd-tk 工具完成 [日志采集](02_log_collection.md) 与 [日志清洗](03_log_parse.md)后，使用诊断功能对清洗后的结构数据进行故障分析。

## 诊断命令

工具提供两种诊断方式：

| 命令 | 说明                                                     | 执行步骤 |
|------|--------------------------------------------------------|----------|
| `auto_collect_diag` | 一键式诊断。整合采集、清洗、诊断三个步骤，等价于 `auto_collect` + `auto_diag`。适用于一次完成端到端诊断的场景 | 采集与清洗 → 诊断 → 生成报告 |
| `auto_diag` | 仅诊断（需先完成采集）                                            | 加载清洗缓存数据 → 诊断 → 生成报告 |

## 使用 `auto_collect_diag` 命令

### 在线场景

非交互式方式：

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect_diag
诊断完成
```

交互式方式：

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

非交互式方式：

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect_diag
诊断完成
```

交互式方式：

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

非交互式方式：

```bash
# 第一步：采集与清洗
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

# 第二步：执行诊断
ascend-fd-tk auto_diag
诊断完成
```

交互式方式：

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_conn_config /path/to/conn.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
>>> auto_diag
诊断完成
```

### 离线场景

非交互式方式：

```bash
# 第一步：采集与清洗
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

# 第二步：执行诊断
ascend-fd-tk auto_diag
诊断完成
```

交互式方式：

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
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
>>> auto_diag
诊断完成
```

### 分批诊断场景

分批诊断适用于**多网络平面**或**设备数量较大**的场景。通过多次执行 `auto_collect` 分批采集不同批次的设备信息，最后统一执行一次 `auto_diag` 完成全量诊断。

> 每次 `auto_collect` 采集的数据会累积到缓存中，不会覆盖之前批次的数据。

#### 在线分批诊断

每次 `auto_collect` 前设置本次要采集的设备连接信息。

非交互式方式：

```bash
# 清理缓存（开始新任务前）
ascend-fd-tk clear_cache

# 第一批：采集网络平面 A 的设备
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn_plane_a.ini auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

# 第二批：采集网络平面 B 的设备
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn_plane_b.ini auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

...

# 统一诊断
ascend-fd-tk auto_diag
诊断完成
```

交互式方式：

```bash
ascend-fd-tk
>>> clear_cache
清理完成
>>>
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_conn_config /path/to/conn_plane_a.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
>>>
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_conn_config /path/to/conn_plane_b.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

...

>>> auto_diag
诊断完成
```

#### 离线分批诊断

每次 `auto_collect` 前设置本次要分析的日志目录。

非交互式方式：

```bash
# 清理缓存（开始新任务前）
ascend-fd-tk clear_cache

# 第一批：采集服务器日志
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

# 第二批：采集 BMC 日志和交换机日志
ascend-fd-tk set_config_dir /path/to/your_config_path set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

...

# 统一诊断
ascend-fd-tk auto_diag
诊断完成
```

交互式方式：

```bash
ascend-fd-tk
>>> clear_cache
清理完成
>>>
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_host_dump_log /path/to/host_logs_a
设置成功
>>> set_bmc_dump_log /path/to/bmc_logs_a
设置成功
>>> set_switch_dump_log /path/to/switch_logs_a
设置成功
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
>>>
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_host_dump_log /path/to/host_logs_b
设置成功
>>> set_bmc_dump_log /path/to/bmc_logs_b
设置成功
>>> set_switch_dump_log /path/to/switch_logs_b
设置成功
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

...

>>> auto_diag
诊断完成
```

## 查看诊断报告

诊断完成后报告自动生成至工具家目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。
