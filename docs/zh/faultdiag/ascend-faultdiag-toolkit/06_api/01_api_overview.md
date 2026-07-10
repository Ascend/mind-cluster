# API 概述

本文档汇总 ascend-fd-tk 提供的全部命令及其功能分类，便于快速检索。各命令的详细说明见对应章节。

工具共提供 16 个命令，按用途分为 6 类。

## 命令索引

### 基础命令

| 命令 | 功能简述 | 是否需要参数 | 详细说明 |
|------|----------|--------------|----------|
| `help` | 显示所有可用命令的帮助信息 | 否 | [help](01_basic/01_help.md) |
| `exit` | 退出程序 | 否 | [exit](01_basic/02_exit.md) |
| `clear` | 清屏 | 否 | [clear](01_basic/03_clear.md) |
| `about` | 查看工具版本信息 | 否 | [about](01_basic/04_about.md) |
| `guide` | 获取使用向导信息 | 否 | [guide](01_basic/05_guide.md) |

### 配置命令

| 命令 | 功能简述 | 是否需要参数 | 详细说明 |
|------|----------|--------------|----------|
| `set_config_dir` | 设置配置文件目录（扫描 `LLD.xlsx`） | 是（目录路径） | [set_config_dir](02_config/01_set_config_dir.md) |
| `set_conn_config` | 设置设备连接配置（主机 / BMC / 交换机） | 是（配置文件路径） | [set_conn_config](02_config/02_set_conn_config.md) |
| `set_host_dump_log` | 设置服务器导出日志目录（离线） | 是（目录路径） | [set_host_dump_log](02_config/03_set_host_dump_log.md) |
| `set_bmc_dump_log` | 设置 BMC 日志目录（离线） | 是（目录路径） | [set_bmc_dump_log](02_config/04_set_bmc_dump_log.md) |
| `set_switch_dump_log` | 设置交换机命令回显文本目录（离线） | 是（目录路径） | [set_switch_dump_log](02_config/05_set_switch_dump_log.md) |

### 采集命令

| 命令 | 功能简述 | 是否需要参数 | 详细说明 |
|------|----------|--------------|----------|
| `collect_bmc_dump_info` | 在线收集 BMC dump info 日志 | 否 | [collect_bmc_dump_info](03_collect/collect_bmc_dump_info.md) |

### 清洗&诊断命令

| 命令 | 功能简述 | 是否需要参数 | 详细说明 |
|------|----------|--------------|----------|
| `auto_collect` | 启动自动信息采集，支持离线、在线采集，适用于不同网络平面分批收集 | 否 | [auto_collect](04_parse_diagnosis/01_auto_collect.md) |
| `auto_diag` | 启动自动诊断（配合分批采集使用） | 否 | [auto_diag](04_parse_diagnosis/02_auto_diag.md) |
| `auto_collect_diag` | 一键式自动收集 + 诊断 | 否 | [auto_collect_diag](04_parse_diagnosis/03_auto_collect_diag.md) |

### 巡检命令

| 命令 | 功能简述 | 是否需要参数 | 详细说明 |
|------|----------|--------------|----------|
| `auto_inspection` | 启动巡检诊断，适用于客户定制化巡检场景 | 可选（客户类型） | [auto_inspection](05_inspection/auto_inspection.md) |

### 维护命令

| 命令 | 功能简述 | 是否需要参数 | 详细说明 |
|------|----------|--------------|----------|
| `clear_cache` | 清理工具运行缓存 | 否 | [clear_cache](06_maintenance/clear_cache.md) |

## 使用指引

- 首次使用或查看能力：依次执行 `guide` → `help`。
- 查看单个命令的详细描述与使用方式：在命令后加 `?`，例如 `set_conn_config ?`。
