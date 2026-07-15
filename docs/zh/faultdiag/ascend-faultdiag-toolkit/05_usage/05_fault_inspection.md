# 故障巡检

ascend-fd-tk 工具完成 [日志采集](02_log_collection.md) 与 [日志清洗](03_log_parse.md)后，使用 [auto_inspection](../06_api/05_inspection/auto_inspection.md) 巡检命令，按不同客户类型预定义规则进行批量健康检查，定期对集群进行巡检，提前发现潜在链路异常。巡检特性是 beta 特性，不建议在正式环境中使用。

> 诊断与巡检的主要区别：
>
> - **诊断**：基于故障已发生时的日志数据进行根因分析，输出 `.xlsx` 诊断报告。
> - **巡检**：针对未发生明确故障的场景，基于预定义规则批量健康检查，输出 `.csv` 巡检报告。

## 在线场景巡检

非交互式方式（展示命令与回显）：

```bash
# 第一步：采集与清洗
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect
收集完成，若完成全部收集请进行诊断/巡检

# 第二步：执行巡检
ascend-fd-tk auto_inspection
巡检完成
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
>>> auto_inspection
巡检完成
```

## 离线场景巡检

非交互式方式（展示命令与回显）：

```bash
# 第一步：采集与清洗
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect
收集完成，若完成全部收集请进行诊断/巡检

# 第二步：执行巡检
ascend-fd-tk auto_inspection
巡检完成
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
>>> auto_inspection
巡检完成
```

## 查看巡检报告

巡检完成后报告自动生成至工具家目录下的 report 子目录，详见[诊断 / 巡检报告说明](06_fault_analysis_report.md)。
