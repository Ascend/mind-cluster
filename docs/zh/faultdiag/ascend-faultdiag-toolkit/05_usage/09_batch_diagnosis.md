# 分批诊断

分批诊断适用于**多网络平面**或**设备数量较大**的场景。通过多次执行 `auto_collect` 分批采集不同批次的设备信息，最后统一执行一次 `auto_diag` 完成全量诊断。

## 交互式命令执行

### 1. 启动工具

```bash
# 启动交互式命令行
ascend-fd-tk
```

进入 `>>>` 提示符后，逐条输入命令。工具启动时会自动显示帮助信息。

### 2. 清理缓存

执行任务前，建议清理缓存，避免上次诊断结果影响本次诊断。

```bash
>>> clear_cache
清理完成
```

### 3. 设置配置文件路径（可选）

```bash
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
```

### 4. 配置数据源并分批诊断

- 在线场景

```bash
>>> set_conn_config /path/to/conn_network_plane_1.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
>>> # 切换到其他网络平面，重新配置 conn.ini 后再次采集
>>> set_conn_config /path/to/conn_network_plane_2.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
>>> auto_diag
诊断完成
```

- 离线场景

```bash
>>> set_host_dump_log /path/to/host_logs
设置成功
>>> set_bmc_dump_log /path/to/bmc_logs
设置成功
>>> set_switch_dump_log /path/to/switch_logs
设置成功
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
>>> # 追加其他批次的离线日志目录后再次采集
>>> set_host_dump_log /path/to/host_logs_batch2
设置成功
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
>>> auto_diag
诊断完成
```

### 5. 查看诊断报告

诊断完成后报告自动生成至工具家目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。

## 非交互式命令执行

非交互式模式将多个命令串联在一行中执行，适用于自动化运维场景。

### 1. 清理缓存

执行任务前，建议清理缓存，避免上次诊断结果影响本次诊断：

```bash
ascend-fd-tk clear_cache
清理完成
```

### 2. 配置数据源并分批诊断

- 在线场景

```bash
# 第一步：采集信息（可多次执行，汇总多次采集结果）
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

# 第二步：执行诊断
ascend-fd-tk auto_diag
诊断完成
```

- 离线场景

```bash
# 第一步：采集信息（可多次执行，汇总多次采集结果）
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

# 第二步：执行诊断
ascend-fd-tk auto_diag
诊断完成
```

### 3. 查看诊断报告

诊断完成后报告自动生成至工具家目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。
