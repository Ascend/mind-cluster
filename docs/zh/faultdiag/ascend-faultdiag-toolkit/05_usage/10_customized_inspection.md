# 客户定制化巡检

客户定制化巡检适用于按需自定义巡检维度，满足客户特定场景的链路巡检需求。巡检特性是 beta 特性，不建议在正式环境中使用。

## 非交互式命令执行（展示命令与回显）

非交互式模式将多个命令串联在一行中执行，适用于自动化运维场景。

> 设置配置文件路径命令 `set_config_dir` 为可选命令。

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
收集完成，若完成全部收集请进行诊断/巡检

# 第二步：执行特定客户类型的巡检
ascend-fd-tk auto_inspection <客户类型>
巡检完成
```

- 离线场景

```bash
# 第一步：采集信息
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect
收集完成，若完成全部收集请进行诊断/巡检

# 第二步：执行特定客户类型的巡检
ascend-fd-tk auto_inspection <客户类型>
巡检完成
```

## 交互式命令执行（展示命令与回显）

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

使用 [auto_inspection](../06_api/05_inspection/auto_inspection.md) 命令执行特定客户类型的巡检。

```bash
>>> auto_inspection <客户类型>
巡检完成
```

## 查看巡检报告

巡检完成后报告自动生成至工具家目录下的 report 子目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。
