# 离线诊断

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

### 4. 配置数据源

离线分析模式，需设置日志目录。详细配置内容请参考 [日志收集与数据源](02_log_collection.md)。

```bash
>>> set_host_dump_log /path/to/host_logs
设置成功
>>> set_bmc_dump_log /path/to/bmc_logs
设置成功
>>> set_switch_dump_log /path/to/switch_logs
设置成功
```

### 5. 一键式诊断

```bash
# 自动完成采集 + 诊断
>>> auto_collect_diag
诊断完成
```

### 6. 查看诊断报告

诊断完成后报告自动生成至工具家目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。

## 非交互式命令执行

非交互式模式将多个命令串联在一行中执行，适用于自动化运维场景。

### 1. 清理缓存

执行任务前，建议清理缓存，避免上次诊断结果影响本次诊断：

```bash
ascend-fd-tk clear_cache
清理完成
```

### 2. 配置离线数据源并一键诊断

```bash
# 设置 Host 服务器日志目录、BMC 日志目录和交换机日志目录 + 诊断
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect_diag
诊断完成
```

### 3. 查看诊断报告

诊断完成后报告自动生成至工具家目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。
