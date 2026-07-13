# set_bmc_dump_log

## 命令功能

设置 BMC 日志目录。

## 命令格式

| 命令格式 | 描述 |
|---------|------|
| `set_bmc_dump_log <目录>` | 设置 BMC 日志目录。 |
| `set_bmc_dump_log ?` | 查看详细说明。 |

## 参数说明

| 参数 | 说明 |
|------|------|
| `<目录>` | BMC 日志目录路径。 |

## 支持的日志类型

- 手动通过 BMC 网页"一键收集"按钮下载的日志。
- 使用 `ipmcget -d diaginfo` 命令采集的日志。
- 通过 ascend-fd-tk 工具内置命令 `collect_bmc_dump_info` 收集的日志。

> **注**：日志详情请参考[日志收集与数据源](../../05_usage/02_log_collection.md)。

## 输出说明

- 设置成功时返回：`设置成功`。
- 设置失败时返回：`地址为空，请重新设置` 或 `地址{dir_path}不存在，请重新设置` 或 `地址{dir_path}非文件夹，请重新设置`。

## 示例

非交互式方式：

```bash
ascend-fd-tk set_bmc_dump_log /data/bmc_logs auto_collect_diag
设置成功
# 其他日志输出...
```

交互式方式：

```bash
ascend-fd-tk
>>> set_bmc_dump_log /data/bmc_logs
设置成功
```
