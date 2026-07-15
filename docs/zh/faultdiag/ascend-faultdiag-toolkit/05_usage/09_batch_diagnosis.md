# 分批诊断

分批诊断适用于**多网络平面**或**设备数量较大**的场景。通过多次执行 `auto_collect` 分批采集不同批次的设备信息，最后统一执行一次 `auto_diag` 完成全量诊断。

## 非交互式命令执行（展示命令与回显）

非交互式模式将多个命令串联在一行中执行，适用于自动化运维场景。

> 设置配置文件路径命令 `set_config_dir` 为可选命令。

### 1. 清理缓存

执行任务前，建议清理缓存，避免上次诊断结果影响本次诊断：

```bash
ascend-fd-tk clear_cache
清理完成
```

### 2. 配置数据源并分批诊断

- 在线场景

```bash
# 第一批：采集网络平面 A 的设备
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn_plane_a.ini auto_collect
收集完成，若完成全部收集请进行诊断/巡检

# 第二批：采集网络平面 B 的设备
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn_plane_b.ini auto_collect
收集完成，若完成全部收集请进行诊断/巡检

...

# 统一诊断
ascend-fd-tk auto_diag
诊断完成
```

- 离线场景

```bash
# 第一批：采集平面 A 的离线日志
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs_a set_bmc_dump_log /path/to/bmc_logs_a set_switch_dump_log /path/to/switch_logs_a auto_collect
收集完成，若完成全部收集请进行诊断/巡检

# 第二批：采集平面 B 的离线日志
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs_b set_bmc_dump_log /path/to/bmc_logs_b set_switch_dump_log /path/to/switch_logs_b auto_collect
收集完成，若完成全部收集请进行诊断/巡检

...

# 统一诊断
ascend-fd-tk auto_diag
诊断完成
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

### 3. 设置配置文件路径（可选）

```bash
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
```

### 4. 配置数据源并分批诊断

- 在线场景

```bash
>>> set_conn_config /path/to/conn_plane_a.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> auto_collect
收集完成，若完成全部收集请进行诊断/巡检
>>>
>>> # 切换到其他网络平面，重新配置 conn.ini 后再次采集
>>> set_conn_config /path/to/conn_plane_b.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> auto_collect
收集完成，若完成全部收集请进行诊断/巡检

...

>>> auto_diag
诊断完成
```

- 离线场景

```bash
>>> set_host_dump_log /path/to/host_logs_a
设置成功
>>> set_bmc_dump_log /path/to/bmc_logs_a
设置成功
>>> set_switch_dump_log /path/to/switch_logs_a
设置成功
>>> auto_collect
收集完成，若完成全部收集请进行诊断/巡检
>>>
>>> # 追加其他批次的离线日志目录后再次采集
>>> set_host_dump_log /path/to/host_logs_b
设置成功
>>> set_bmc_dump_log /path/to/bmc_logs_b
设置成功
>>> set_switch_dump_log /path/to/switch_logs_b
设置成功
>>> auto_collect
收集完成，若完成全部收集请进行诊断/巡检

...

>>> auto_diag
诊断完成
```

## 查看诊断报告

诊断完成后报告自动生成至工具家目录下的 report 子目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。
