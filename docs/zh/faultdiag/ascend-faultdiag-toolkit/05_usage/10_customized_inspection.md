# 客户定制化巡检

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

### 3. 配置数据源

请参考[在线诊断](07_online_diagnosis.md)或[离线诊断](08_offline_diagnosis.md)中的数据源配置步骤。

### 4. 巡检

```bash
# 执行特定客户类型的巡检
>>> auto_inspection <客户类型>
诊断完成
```

### 5. 查看巡检报告

巡检完成后报告自动生成至工具家目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。

## 非交互式命令执行

非交互式模式将多个命令串联在一行中执行，适用于自动化运维场景。

### 1. 清理缓存

执行任务前，建议清理缓存，避免上次诊断结果影响本次诊断：

```bash
ascend-fd-tk clear_cache
清理完成
```

### 2. 配置数据源并巡检

- 在线场景

```bash
# 第一步：采集信息
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

# 第二步：执行特定客户类型的巡检
ascend-fd-tk auto_inspection <客户类型>
诊断完成
```

- 离线场景

```bash
# 第一步：采集信息
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断

# 第二步：执行特定客户类型的巡检
ascend-fd-tk auto_inspection <客户类型>
诊断完成
```

### 3. 查看巡检报告

巡检完成后报告自动生成至工具家目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。
