# set_host_dump_log

## 命令功能

设置服务器导出日志目录，用于离线分析场景。

## 命令格式

| 命令格式 | 描述 |
|---------|------|
| `set_host_dump_log <目录>` | 设置服务器导出日志目录 |
| `set_host_dump_log ?` | 查看详情 |

## 参数说明

| 参数 | 类型 | 是否必填 | 说明 |
|------|-----|------|------|
| `<目录>` | string | 是 | 服务器日志目录路径。 |

## 支持的日志类型

日志详情请参考 [host 离线日志采集](../../05_usage/02_log_collection.md#host-offline-log)。

## 输出说明

- 设置成功时返回：`设置成功`。
- 设置失败时返回：`地址为空，请重新设置` 或 `地址{dir_path}不存在，请重新设置` 或 `地址{dir_path}非文件夹，请重新设置`。

## 示例

非交互式方式（展示命令与回显）：

```bash
ascend-fd-tk set_host_dump_log /data/host_logs auto_collect_diag
设置成功
# 其他日志输出...
```

交互式方式（展示命令与回显）：

```bash
ascend-fd-tk
>>> set_host_dump_log /data/host_logs
设置成功
```
