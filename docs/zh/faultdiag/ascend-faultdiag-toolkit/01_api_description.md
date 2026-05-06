# 接口描述

## 版本信息

工具版本信息可通过 `about` 命令查看。

## 默认配置

当未手动设置以下路径时，工具会自动读取执行路径下的默认文件或目录：

- 连接配置：conn.ini
- BMC日志目录：bmc_dump_log
- Host日志目录：host_dump_log
- 交换机日志目录：switch_dump_log

## 接口调用流程

**在线诊断流程**

1. 使用 `set_conn_config` 设置设备连接配置。
2. 使用 `auto_collect_diag` 启动一键式诊断。
3. 诊断完成后使用 `clear_cache` 清理缓存。

**离线诊断流程**

1. 使用 `set_host_dump_log` 设置服务器日志目录。
2. 使用 `set_bmc_dump_log` 设置BMC日志目录。
3. 使用 `set_switch_dump_log` 设置交换机日志目录。
4. 使用 `auto_collect_diag` 启动一键式诊断。
5. 诊断完成后使用 `clear_cache` 清理缓存。

**分批诊断流程**

1. 使用配置命令设置部分设备信息。
2. 使用 `auto_collect` 收集设备信息。
3. 重复执行步骤1和步骤2，设置和收集其他设备信息。
4. 使用 `auto_diag` 启动统一诊断。
5. 诊断完成后使用 `clear_cache` 清理缓存。

## 基础命令

### help

**命令功能**

显示帮助信息。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| help | 显示所有可用命令的帮助信息。 |
| help ? | 查看详情。  |

### exit

**命令功能**

退出链路诊断工具。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| exit | 退出链路诊断工具。 |
| exit ? | 查看详情。 |

### clear

**命令功能**

清空终端屏幕。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| clear | 清空终端屏幕。 |
| clear ? | 查看详情。  |

### about

**命令功能**

查看链路诊断工具信息。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| about | 显示链路诊断工具的版本和联系信息。 |
| about ? | 查看详情。       |

### guide

**命令功能**

获取链路诊断工具的使用向导信息。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| guide | 显示链路诊断工具的使用向导。 |
| guide ? | 查看详情。        |

## 配置命令

### set_conn_config

**命令功能**

设置设备连接配置信息。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| set_conn_config <i><文件地址></i> | 设置设备连接配置文件。 |
| set_conn_config ? | 查看详细配置说明。 |

**参数说明**

|参数|说明|
|---|---|
|<i><文件地址></i>|连接配置文件的路径。|

**配置文件结构**

```ini
[host]
# port指定端口,不写默认为22, username指定用户名, password指定密码, private_key指定私钥文件
1.1.1.1 port="22" username="root" private_key="~/.ssh/your_private_key"
1.1.2.1 port="22" username="root" password="321" 

[bmc]
1.1.1.2 username="Administrator" password="123"

[switch]
# 支持ip1-ip2 ip段方式填写(需保证账号密码相同), 通过step设置步长
1.1.1.3-1.1.1.10 step=1 username="root" password="123"

[config]
# 支持设置全局的私钥文件
private_key="~/.ssh/your_private_key"
```

### set_host_dump_log

**命令功能**

设置服务器导出日志目录。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| set_host_dump_log <i><目录></i> | 设置服务器导出日志目录。 |
| set_host_dump_log ? | 查看详细说明。 |

**参数说明**

|参数|说明|
|---|---|
|<i><目录></i>|服务器日志目录路径。|

**支持的日志类型**

- `A3device日志一键采集脚本<version>.sh`收集的日志
- `link_down_collect_<version>.sh`收集的日志
- `tool_log_collection_out_version_all_<version>.sh`收集的日志

### set_bmc_dump_log

**命令功能**

设置BMC日志目录。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| set_bmc_dump_log <i><目录></i> | 设置BMC日志目录。 |
| set_bmc_dump_log ? | 查看详细说明。 |

**参数说明**

|参数|说明|
|---|---|
|<i><目录></i>|BMC日志目录路径。|

**支持的日志类型**

- 手动通过BMC网页 "一键收集" 按钮下载的日志。
- 使用 `ipmcget -d diaginfo` 命令采集的日志。

### set_switch_dump_log

**命令功能**

设置交换机命令回显文本目录。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| set_switch_dump_log <i><目录></i> | 设置交换机命令回显目录。 |
| set_switch_dump_log ? | 查看详细说明。 |

**参数说明**

|参数|说明|
|---|---|
|<i><目录></i>|交换机日志目录路径。|

**支持的日志类型**

- 使用交换机 `display diagnostic-information <filename>` 命令导出的结果或者查询关键命令后复制的shell回显文本。
- 使用交换机 `collect diagnostic-information` 命令导出的日志zip包。

## 采集命令

### collect_bmc_dump_info

**命令功能**

在线收集BMC dump info日志。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| collect_bmc_dump_info | 在线收集BMC dump info日志。 |
| collect_bmc_dump_info ? | 查看详情。        |

**输出说明**

收集完成后，日志位于 `CommonPath.TOOL_HOME_BMC_DUMP_CACHE_DIR` 目录。

### auto_collect

**命令功能**

启动自动信息采集。

- 支持离线和在线采集。
- 适用于不同网络平面分批收集。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| auto_collect | 启动自动信息采集。 |
| auto_collect ? | 查看详细说明。 |

## 诊断命令

### auto_inspection

**命令功能**

启动巡检结果诊断。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| auto_inspection | 使用默认客户类型启动诊断。 |
| auto_inspection <i><客户类型></i> | 使用指定客户类型启动诊断。 |
| auto_inspection ? | 查看支持的客户类型。 |

**参数说明**：

|参数|说明|
|---|---|
|<i><客户类型></i>|支持的客户类型枚举值。目前支持default。|

### auto_diag

**命令功能**

启动自动诊断，适用于分批收集后统一诊断。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| auto_diag | 启动自动诊断。 |
| auto_diag ? | 查看详细说明。 |

### auto_collect_diag

**命令功能**

启动一键式自动收集诊断。自动执行收集（在线设备采集或离线日志收集）和诊断流程。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| auto_collect_diag | 启动一键式自动收集和诊断。 |
| auto_collect_diag ? | 查看详情。        |

## 维护命令

### clear_cache

**命令功能**

清理缓存。

- 清理工具运行过程中产生的缓存文件。
- 建议在执行新诊断任务前执行。
- 若清理未生效，请使用管理员模式打开工具。

**命令格式**

| 命令格式 | 描述 |
|---------|------|
| clear_cache | 清理链路诊断工具缓存。 |
| clear_cache ? | 查看详情。        |
