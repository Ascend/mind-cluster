# 日志清洗

根据不同的采集模式使用 `auto_collect` 命令进行数据清洗，将原始日志解析为结构化数据，供后续诊断和巡检使用。

| 采集模式 | 清洗行为                                   |
|----------|----------------------------------------|
| 在线模式 | 自动 SSH 采集日志 + 清洗 → 清洗结果落盘到 `家目录/cache` |
| 离线模式 | 提前收集离线日志文件 → 清洗 → 清洗结果落盘到 `家目录/cache`  |

## 在线采集与清洗

配置文件 `conn.ini` 包含设备 IP、账号、密码 / 密钥等信息，详细配置内容请参考 [set_conn_config](../06_api/02_config/02_set_conn_config.md)。

非交互式方式：

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
```

交互式方式：

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_conn_config /path/to/conn.ini
设置成功，请尽快删除包含明文密码的配置文件
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
```

## 离线日志清洗

集群场景下，只需要将多个设备的日志放到对应目录下即可。目录结构示例如下：

```text
# 直接将收集的日志压缩包放入到目录即可，工具支持自动解压日志压缩包，无需手动解压。至少配置一个日志目录即可进行诊断。
host日志采集目录/
    ├── {host_01_file_name}.tar.gz
    ├── {host_02_file_name}.tar.gz
    └── ...
bmc日志采集目录/
    ├── {bmc_01_file_name}.tar.gz
    ├── {bmc_02_file_name}.tar.gz
    └── ...
switch日志采集目录/
    ├── {switch_011_file_name}.zip
    ├── {switch_02_file_name}.zip
    └── ...
```

非交互式方式：

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log host日志采集目录/ set_bmc_dump_log bmc日志采集目录/ set_switch_dump_log switch日志采集目录/ auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
```

交互式方式：

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
设置成功，配置目录：/path/to/your_config_path
>>> set_host_dump_log host日志采集目录/
设置成功
>>> set_bmc_dump_log bmc日志采集目录/
设置成功
>>> set_switch_dump_log switch日志采集目录/
设置成功
>>> auto_collect
收集完成，若完成全部收集请使用 "auto_diag" 进行诊断
```

## 默认路径读取

当未手动设置连接配置文件 `conn.ini` 或离线日志目录时，工具会自动读取[ascend-fd-tk 家目录](01_usage_overview.md)路径下的以下默认文件或目录，相关文件或目录需用户提前手动创建、配置设备连接信息和放置离线日志到对应目录。

- 连接配置：`家目录/conn.ini`
- BMC 日志目录：`家目录/bmc_dump_log`
- Host 日志目录：`家目录/host_dump_log`
- 交换机日志目录：`家目录/switch_dump_log`
