# 特性概览

ascend-fd-tk 工具提供完整的链路故障诊断能力，覆盖从数据采集到故障定位的全流程。本章介绍工具特性的使用方式与典型场景。

## 工具使用方式

工具支持两种使用方式：**交互式模式**与**非交互式模式**。

| 模式 | 适用场景 | 特点 |
|------|----------|------|
| 交互式 | 临时调试、问题排查、逐步操作 | 进入 `>>>` 提示符，逐条输入命令 |
| 非交互式 | 自动化运维、定时任务、脚本集成 | 一行命令完成全流程 |

## ascend-fd-tk 家目录说明

工具运行时产生的数据存储于家目录，不同平台路径不同：

- Linux 平台路径基于用户主目录（`~`），家目录为：`~/.ascend-faultdiag-toolkit/`。
- Windows 平台基于**当前工作目录**（启动工具时所在的目录），家目录为：`{当前工作目录}/.ascend-faultdiag-toolkit/`。

目录结构：

| 路径 | 用途 |
|------|------|
| `家目录/cache/` | 清洗结果缓存信息（host / bmc / switch 分类，按照设备 IP 或目录名落盘的 JSON 文件） |
| `家目录/logs/ascend-fd-tk.log` | 工具运行日志 |
| `家目录/report/` | 诊断 / 巡检报告输出目录（`diag_report_{YYYYMMDD_HHMMSS}.xlsx` / `inspection_errors.csv`） |
| `家目录/encrypted_conn_config` | 在线连接配置文件 `conn.ini` 加密后的文件 |

> **说明**：工具运行日志单文件上限为 10MB，文件达到阈值后自动触发日志切分归档，归档文件命名规则为 ascend-fd-tk.log.1、ascend-fd-tk.log.2…… 其中编号数值越小代表日志生成时间越新，系统最多留存 5 份归档日志文件。

## 特性概览

| 特性 | 说明 |
|------|------|
| [日志采集](02_log_collection.md)与[日志清洗](03_log_parse.md) | 在线模式自动收集并清洗；离线模式提前收集日志再清洗 |
| [故障诊断](04_fault_diagnosis.md) | 对清洗后的数据进行故障检测和根因分析，生成 Excel 诊断报告 |
| [故障巡检](05_fault_inspection.md) | 按不同客户类型预定义规则批量健康检查，生成 CSV 巡检报告 |

## 场景命令流程

| 流程              | 适用场景 | 核心步骤 |
|-----------------|----------|----------|
| [在线诊断流程](07_online_diagnosis.md) | 设备可访问（IP / 凭据齐备） | `clear_cache` → `set_config_dir`（可选）→ `set_conn_config` → `auto_collect_diag` 或 `auto_collect` + `auto_diag` |
| [离线诊断流程](08_offline_diagnosis.md)      | 仅日志可获取 | `clear_cache` → `set_config_dir`（可选）→ `set_host_dump_log` / `set_bmc_dump_log` / `set_switch_dump_log`（任选）→ `auto_collect_diag` 或 `auto_collect` + `auto_diag` |
| [分批诊断流程](09_batch_diagnosis.md)      | 多网络平面、设备数量大 | `clear_cache` → `set_config_dir`（可选）→ 重复 N 次[`set_conn_config` + `auto_collect` 或 `set_host_dump_log` / `set_bmc_dump_log` / `set_switch_dump_log`（任选）+ `auto_collect`] → `auto_diag` |
| [客户定制化巡检](10_customized_inspection.md)     | 按不同客户类型预定义规则批量健康检查 | `clear_cache` → `set_config_dir`（可选）→ `set_conn_config` 或 `set_host_dump_log` / `set_bmc_dump_log` / `set_switch_dump_log`（任选）→ `auto_collect` + `auto_inspection` |
