# 在线诊断

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

在线采集模式，需在配置文件 `conn.ini` 中配置 IP、账号、密码 / 密钥等信息，详细配置内容请参考 [set_conn_config](../06_api/02_config/02_set_conn_config.md)。

```bash
>>> set_conn_config /path/to/conn.ini
设置成功，请尽快删除包含明文密码的配置文件
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

### 2. 配置在线数据源并一键诊断

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect_diag
诊断完成
```

### 3. 查看诊断报告

诊断完成后报告自动生成至工具家目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。
